package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"koperasi-frontend/core/model"
)

// pgLoanRepository: implementasi LoanRepository di atas PostgreSQL.
// loans memakai member_id FK (tanpa kolom nama) -> JOIN members untuk nama.
type pgLoanRepository struct {
	db *gorm.DB
}

func NewLoanRepository(db *gorm.DB) LoanRepository {
	return &pgLoanRepository{db: db}
}

// loanRow: DTO scan tanpa field relasi (model.Loan.Installments membuat GORM gagal
// memetakan saat Scan). Dikonversi ke model.Loan, angsuran dimuat terpisah.
type loanRow struct {
	ID          int
	MemberNama  string
	Nominal     float64
	TenorBulan  int
	BungaPersen float64
	Tujuan      string
	SisaPokok   float64
	Status      string
	TanggalCair string
}

func (x loanRow) toModel() model.Loan {
	return model.Loan{
		ID: x.ID, MemberNama: x.MemberNama, Nominal: x.Nominal, TenorBulan: x.TenorBulan,
		BungaPersen: x.BungaPersen, Tujuan: x.Tujuan, SisaPokok: x.SisaPokok,
		Status: x.Status, TanggalCair: x.TanggalCair,
	}
}

const loanSelect = `SELECT l.id, m.nama AS member_nama, l.nominal, l.tenor_bulan, l.bunga_persen,
	        l.tujuan, l.sisa_pokok, l.status,
	        COALESCE(to_char(l.tanggal_cair,'YYYY-MM-DD'),'') AS tanggal_cair
	   FROM loans l
	   JOIN members m ON m.id = l.member_id`

func (r *pgLoanRepository) scanLoans(ctx context.Context, sql string, args ...interface{}) ([]model.Loan, error) {
	rows := []loanRow{}
	if err := r.db.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model.Loan, len(rows))
	for i, x := range rows {
		out[i] = x.toModel()
	}
	return out, nil
}

func (r *pgLoanRepository) List(ctx context.Context, scopeNama string) ([]model.Loan, error) {
	if scopeNama != "" {
		return r.scanLoans(ctx, loanSelect+" WHERE m.nama = ? ORDER BY l.id DESC", scopeNama)
	}
	return r.scanLoans(ctx, loanSelect+" ORDER BY l.id DESC")
}

func (r *pgLoanRepository) FindByID(ctx context.Context, id int) (*model.Loan, error) {
	out, err := r.scanLoans(ctx, loanSelect+" WHERE l.id = ? LIMIT 1", id)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 || out[0].ID == 0 {
		return nil, nil
	}
	loan := out[0]
	insts, err := r.installments(ctx, loan.ID)
	if err != nil {
		return nil, err
	}
	loan.Installments = insts
	return &loan, nil
}

func (r *pgLoanRepository) installments(ctx context.Context, loanID int) ([]model.Installment, error) {
	out := []model.Installment{}
	err := r.db.WithContext(ctx).Raw(
		`SELECT bulan_ke, to_char(jatuh_tempo,'YYYY-MM-DD') AS jatuh_tempo,
		        nominal_pokok, nominal_bunga, total_bayar, status,
		        COALESCE(to_char(tanggal_bayar,'YYYY-MM-DD'),'') AS tanggal_bayar
		   FROM installments WHERE loan_id = ? ORDER BY bulan_ke`, loanID).Scan(&out).Error
	return out, err
}

func (r *pgLoanRepository) Create(ctx context.Context, memberNama string, l model.Loan) (int, error) {
	var id int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var memberID int
		if err := tx.Raw("SELECT id FROM members WHERE nama = ? LIMIT 1", memberNama).Scan(&memberID).Error; err != nil {
			return err
		}
		if memberID == 0 {
			return fmt.Errorf("anggota %q tidak ditemukan", memberNama)
		}
		return tx.Raw(
			`INSERT INTO loans (member_id, nominal, tenor_bulan, bunga_persen, tujuan, sisa_pokok, status)
			 VALUES (?, ?, ?, ?, ?, ?, 'PENDING') RETURNING id`,
			memberID, l.Nominal, l.TenorBulan, l.BungaPersen, l.Tujuan, l.SisaPokok,
		).Scan(&id).Error
	})
	return id, err
}

func (r *pgLoanRepository) SetStatus(ctx context.Context, id int, status string) error {
	return r.db.WithContext(ctx).
		Exec("UPDATE loans SET status = ? WHERE id = ?", status, id).Error
}

func (r *pgLoanRepository) Disburse(ctx context.Context, id int, tanggalCair string, installments []model.Installment, journal model.JournalEntry) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Exec(
			`UPDATE loans SET status='AKTIF', tanggal_cair=? WHERE id=? AND status='DISETUJUI'`,
			tanggalCair, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("pinjaman tidak dalam status DISETUJUI")
		}
		for _, in := range installments {
			if err := tx.Exec(
				`INSERT INTO installments (loan_id, bulan_ke, jatuh_tempo, nominal_pokok, nominal_bunga, total_bayar, status)
				 VALUES (?, ?, ?, ?, ?, ?, 'BELUM')`,
				id, in.BulanKe, in.JatuhTempo, in.NominalPokok, in.NominalBunga, in.TotalBayar).Error; err != nil {
				return err
			}
		}
		return tx.Exec(
			`INSERT INTO journal_entries (tanggal, keterangan, akun_debit, akun_kredit, nominal, tipe_transaksi)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			journal.Tanggal, journal.Keterangan, journal.AkunDebit, journal.AkunKredit,
			journal.Nominal, journal.TipeTransaksi).Error
	})
}

func (r *pgLoanRepository) Pay(ctx context.Context, id, bulanKe int, tanggalBayar string, newSisaPokok float64, lunas bool, journal model.JournalEntry) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Exec(
			`UPDATE installments SET status='DIBAYAR', tanggal_bayar=?
			  WHERE loan_id=? AND bulan_ke=? AND status IN ('BELUM','OVERDUE')`,
			tanggalBayar, id, bulanKe)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("angsuran bulan ke-%d tidak dapat dibayar", bulanKe)
		}
		status := "AKTIF"
		if lunas {
			status = "LUNAS"
		}
		if err := tx.Exec(
			"UPDATE loans SET sisa_pokok=?, status=? WHERE id=?",
			newSisaPokok, status, id).Error; err != nil {
			return err
		}
		return tx.Exec(
			`INSERT INTO journal_entries (tanggal, keterangan, akun_debit, akun_kredit, nominal, tipe_transaksi)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			journal.Tanggal, journal.Keterangan, journal.AkunDebit, journal.AkunKredit,
			journal.Nominal, journal.TipeTransaksi).Error
	})
}

func (r *pgLoanRepository) CreditInfo(ctx context.Context, memberNama string) (float64, int, error) {
	type infoRow struct {
		TotalSimpanan float64
		ActiveLoans   int
	}
	var info infoRow
	err := r.db.WithContext(ctx).Raw(
		`SELECT (m.simpanan_pokok + m.simpanan_wajib + m.simpanan_sukarela) AS total_simpanan,
		        (SELECT count(*) FROM loans l WHERE l.member_id = m.id AND l.status = 'AKTIF') AS active_loans
		   FROM members m WHERE m.nama = ? LIMIT 1`, memberNama).Scan(&info).Error
	return info.TotalSimpanan, info.ActiveLoans, err
}
