package service

import (
	"context"
	"time"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
)

// DashboardService menyusun ringkasan dashboard (staff & anggota) dengan
// mengomposisi repository yang sudah ada — tidak mengakses DB langsung.
type DashboardService struct {
	members  repository.MemberRepository
	loans    repository.LoanRepository
	products repository.ProductRepository
	orders   repository.OrderRepository
}

func NewDashboardService(m repository.MemberRepository, l repository.LoanRepository, p repository.ProductRepository, o repository.OrderRepository) *DashboardService {
	return &DashboardService{members: m, loans: l, products: p, orders: o}
}

// DashboardActivity adalah satu baris aktivitas terbaru di dashboard staff.
type DashboardActivity struct {
	Waktu     string
	Aktivitas string
	Nominal   float64
}

// StaffDashboard menampung KPI operasional (OWNER/KASIR).
type StaffDashboard struct {
	AnggotaAktif     int
	AnggotaPending   int
	TotalSimpanan    float64
	PinjamanAktif    int
	PinjamanPending  int
	ProdukPending    int
	ProdukStokRendah int
	OrderHariIni     int
	OrderDisputed    int
	TotalOmzet       float64
	Activities       []DashboardActivity
}

// AnggotaDashboard menampung ringkasan personal anggota.
type AnggotaDashboard struct {
	Member          *model.Member
	TotalSimpanan   float64
	PinjamanAktif   int
	PinjamanPending int
	SisaPinjaman    float64
	OrderAktif      int
	OrderSelesai    int
	RecentOrders    []model.Order
}

// Staff menghitung ringkasan operasional untuk OWNER/KASIR.
func (s *DashboardService) Staff(ctx context.Context) (*StaffDashboard, error) {
	d := &StaffDashboard{}

	members, err := s.members.List(ctx, "ALL")
	if err != nil {
		return nil, err
	}
	for _, m := range members {
		switch m.Status {
		case "AKTIF":
			d.AnggotaAktif++
		case "PENDING":
			d.AnggotaPending++
		}
		d.TotalSimpanan += m.SimpananPokok + m.SimpananWajib + m.SimpananSukarela
	}

	loans, err := s.loans.List(ctx, "")
	if err != nil {
		return nil, err
	}
	for _, l := range loans {
		switch l.Status {
		case "AKTIF":
			d.PinjamanAktif++
		case "PENDING":
			d.PinjamanPending++
		}
	}

	pending, err := s.products.ListByStatus(ctx, "PENDING")
	if err != nil {
		return nil, err
	}
	d.ProdukPending = len(pending)
	// Stok rendah dihitung pada produk APPROVED (inventaris yang dijual).
	approved, err := s.products.Catalog(ctx, "", "")
	if err != nil {
		return nil, err
	}
	for _, p := range approved {
		if p.Stok < p.BatasStokMinimum {
			d.ProdukStokRendah++
		}
	}

	orders, err := s.orders.List(ctx, "")
	if err != nil {
		return nil, err
	}
	today := time.Now().Format("2006-01-02")
	for _, o := range orders {
		if len(o.CreatedAt) >= 10 && o.CreatedAt[:10] == today {
			d.OrderHariIni++
		}
		switch o.Status {
		case "DISPUTED":
			d.OrderDisputed++
		case "SELESAI":
			d.TotalOmzet += o.TotalHarga
		}
	}
	// orders sudah terbaru-dulu; ambil 4 aktivitas teratas.
	for _, o := range orders {
		if len(d.Activities) >= 4 {
			break
		}
		d.Activities = append(d.Activities, DashboardActivity{
			Waktu:     o.CreatedAt,
			Aktivitas: "Order " + o.NomorOrder + " (" + o.PembeliNama + ") — " + o.Status,
			Nominal:   o.TotalHarga,
		})
	}
	return d, nil
}

// Anggota menghitung ringkasan personal untuk satu anggota berdasarkan nama.
func (s *DashboardService) Anggota(ctx context.Context, nama string) (*AnggotaDashboard, error) {
	d := &AnggotaDashboard{}

	m, err := s.members.FindByNama(ctx, nama)
	if err != nil {
		return nil, err
	}
	d.Member = m
	if m != nil {
		d.TotalSimpanan = m.SimpananPokok + m.SimpananWajib + m.SimpananSukarela
	}

	loans, err := s.loans.List(ctx, nama)
	if err != nil {
		return nil, err
	}
	for _, l := range loans {
		switch l.Status {
		case "AKTIF":
			d.PinjamanAktif++
			d.SisaPinjaman += l.SisaPokok
		case "PENDING":
			d.PinjamanPending++
		}
	}

	orders, err := s.orders.List(ctx, nama)
	if err != nil {
		return nil, err
	}
	for _, o := range orders {
		switch o.Status {
		case "SELESAI":
			d.OrderSelesai++
		case "BATAL":
			// tidak dihitung sebagai aktif maupun selesai
		default:
			d.OrderAktif++
		}
		if len(d.RecentOrders) < 5 {
			d.RecentOrders = append(d.RecentOrders, o)
		}
	}
	return d, nil
}
