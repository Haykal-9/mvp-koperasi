package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
)

// Aturan bisnis pinjaman (M4). Konstanta ini sebelumnya tersebar di handler.
const (
	MinPinjaman        = 500000.0 // nominal minimal pengajuan
	MinTenorBulan      = 3        // selaras dengan CHECK (tenor_bulan BETWEEN 3 AND 24)
	MaxTenorBulan      = 24
	DefaultBungaPersen = 1.5 // bunga per bulan atas sisa pokok (flat menurun)
)

// LoanService memusatkan aturan bisnis pinjaman: validasi pengajuan, perhitungan
// bunga & jadwal angsuran, estimasi cicilan, progres, dan penerapan pembayaran.
// Perhitungan murni tetap dapat diuji tanpa DB (repo boleh nil); orkestrasi data
// didelegasikan ke LoanRepository.
type LoanService struct {
	repo repository.LoanRepository
}

func NewLoanService(repo repository.LoanRepository) *LoanService {
	return &LoanService{repo: repo}
}

// ValidateApplication memvalidasi input pengajuan pinjaman.
func (s *LoanService) ValidateApplication(nominal float64, tenor int, tujuan string) error {
	if nominal < MinPinjaman {
		return fmt.Errorf("Nominal pinjaman minimal Rp %.0f.", MinPinjaman)
	}
	if tenor < MinTenorBulan || tenor > MaxTenorBulan {
		return fmt.Errorf("Tenor harus %d-%d bulan.", MinTenorBulan, MaxTenorBulan)
	}
	if tujuan == "" {
		return fmt.Errorf("Tujuan pinjaman wajib diisi.")
	}
	return nil
}

// BuildSchedule membuat jadwal angsuran bulanan dengan bunga menurun atas sisa
// pokok: pokok dibagi rata per bulan, bunga = sisa_pokok * bunga% / 100.
func (s *LoanService) BuildSchedule(nominal float64, tenor int, bungaPersen float64, tanggalCair string) []model.Installment {
	if tenor <= 0 {
		return nil
	}
	base, err := time.Parse("2006-01-02", tanggalCair)
	if err != nil {
		base = time.Now()
	}
	pokokPerBulan := nominal / float64(tenor)
	sisaPokok := nominal

	out := make([]model.Installment, 0, tenor)
	for i := 1; i <= tenor; i++ {
		bunga := sisaPokok * bungaPersen / 100
		out = append(out, model.Installment{
			BulanKe:      i,
			JatuhTempo:   base.AddDate(0, i, 0).Format("2006-01-02"),
			NominalPokok: pokokPerBulan,
			NominalBunga: bunga,
			TotalBayar:   pokokPerBulan + bunga,
			Status:       "BELUM",
		})
		sisaPokok -= pokokPerBulan
	}
	return out
}

// EstimateMonthly memberi estimasi cicilan bulanan (pokok rata + bunga atas
// nominal penuh) untuk ditampilkan sebelum pencairan.
func (s *LoanService) EstimateMonthly(nominal, bungaPersen float64, tenor int) float64 {
	if tenor <= 0 {
		return 0
	}
	pokok := nominal / float64(tenor)
	bunga := nominal * bungaPersen / 100
	return pokok + bunga
}

// Progress mengembalikan persentase pelunasan pokok (0-100, dibulatkan).
func (s *LoanService) Progress(nominal, sisaPokok float64) float64 {
	if nominal <= 0 {
		return 0
	}
	return math.Round((1 - sisaPokok/nominal) * 100)
}

// NextUnpaid mengembalikan angsuran berikutnya yang belum lunas, atau nil.
func (s *LoanService) NextUnpaid(insts []model.Installment) *model.Installment {
	for i := range insts {
		if insts[i].Status == "BELUM" || insts[i].Status == "OVERDUE" {
			return &insts[i]
		}
	}
	return nil
}

// ApplyPayment menandai angsuran terdekat sebagai DIBAYAR, mengurangi sisa pokok,
// dan menyetel status LUNAS bila seluruh angsuran selesai. Mengubah l di tempat.
// Mengembalikan angsuran yang dibayar, flag lunas, dan error bila tak ada yang perlu dibayar.
func (s *LoanService) ApplyPayment(l *model.Loan, tanggalBayar string) (*model.Installment, bool, error) {
	var paid *model.Installment
	for i := range l.Installments {
		inst := &l.Installments[i]
		if inst.Status == "BELUM" || inst.Status == "OVERDUE" {
			inst.Status = "DIBAYAR"
			inst.TanggalBayar = tanggalBayar
			l.SisaPokok -= inst.NominalPokok
			if l.SisaPokok < 0 {
				l.SisaPokok = 0
			}
			paid = inst
			break
		}
	}
	if paid == nil {
		return nil, false, fmt.Errorf("tidak ada angsuran yang perlu dibayar")
	}

	lunas := true
	for _, inst := range l.Installments {
		if inst.Status != "DIBAYAR" {
			lunas = false
			break
		}
	}
	if lunas {
		l.Status = "LUNAS"
	}
	return paid, lunas, nil
}

// ===== Orkestrasi data (memerlukan repo) =====

// CreditInfo mengembalikan total simpanan & jumlah pinjaman AKTIF anggota.
func (s *LoanService) CreditInfo(ctx context.Context, memberNama string) (float64, int, error) {
	return s.repo.CreditInfo(ctx, memberNama)
}

// List mengembalikan pinjaman; scopeNama != "" membatasi ke milik anggota itu.
func (s *LoanService) List(ctx context.Context, scopeNama string) ([]model.Loan, error) {
	return s.repo.List(ctx, scopeNama)
}

// FindByID mengembalikan satu pinjaman beserta angsurannya (nil bila tak ada).
func (s *LoanService) FindByID(ctx context.Context, id int) (*model.Loan, error) {
	return s.repo.FindByID(ctx, id)
}

// Apply memvalidasi lalu menyimpan pengajuan pinjaman PENDING. Mengembalikan id baru.
func (s *LoanService) Apply(ctx context.Context, memberNama string, nominal float64, tenor int, tujuan string) (int, error) {
	if err := s.ValidateApplication(nominal, tenor, tujuan); err != nil {
		return 0, err
	}
	l := model.Loan{
		MemberNama:  memberNama,
		Nominal:     nominal,
		TenorBulan:  tenor,
		BungaPersen: DefaultBungaPersen,
		Tujuan:      tujuan,
		SisaPokok:   nominal,
		Status:      "PENDING",
	}
	return s.repo.Create(ctx, memberNama, l)
}

// Approve mengubah pinjaman PENDING menjadi DISETUJUI.
func (s *LoanService) Approve(ctx context.Context, id int) error {
	return s.repo.SetStatus(ctx, id, "DISETUJUI")
}

// Reject mengubah pinjaman PENDING menjadi DITOLAK.
func (s *LoanService) Reject(ctx context.Context, id int) error {
	return s.repo.SetStatus(ctx, id, "DITOLAK")
}

// Disburse mencairkan pinjaman DISETUJUI: bangun jadwal angsuran + jurnal,
// lalu persist transaksional. Mengubah l di tempat (status/tanggal/angsuran).
func (s *LoanService) Disburse(ctx context.Context, l *model.Loan) error {
	l.TanggalCair = time.Now().Format("2006-01-02")
	l.Installments = s.BuildSchedule(l.Nominal, l.TenorBulan, l.BungaPersen, l.TanggalCair)
	l.Status = "AKTIF"
	je := JournalForLoanDisburse(l.MemberNama, l.Nominal, l.TanggalCair)
	return s.repo.Disburse(ctx, l.ID, l.TanggalCair, l.Installments, je)
}

// Pay menerapkan pembayaran angsuran terdekat (perhitungan murni) lalu persist
// transaksional. Mengembalikan angsuran terbayar & flag lunas.
func (s *LoanService) Pay(ctx context.Context, l *model.Loan, tanggalBayar string) (*model.Installment, bool, error) {
	inst, lunas, err := s.ApplyPayment(l, tanggalBayar)
	if err != nil {
		return nil, false, err
	}
	je := JournalForLoanPayment(l.MemberNama, inst.BulanKe, inst.TotalBayar, inst.TanggalBayar)
	if err := s.repo.Pay(ctx, l.ID, inst.BulanKe, inst.TanggalBayar, l.SisaPokok, lunas, je); err != nil {
		return nil, false, err
	}
	return inst, lunas, nil
}
