package handler

import (
	"koperasi-frontend/core/mock"
	"koperasi-frontend/core/model"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	Render Renderer
}

func NewDashboardHandler(render Renderer) *DashboardHandler {
	return &DashboardHandler{Render: render}
}

type dashboardActivity struct {
	Waktu     string
	Aktivitas string
	Nominal   float64
}

// Index routes the dashboard view by role:
//   OWNER / KASIR  -> full operational dashboard
//   ANGGOTA        -> personal self-service dashboard
func (h *DashboardHandler) Index(c *gin.Context) {
	if CurrentUserRole(c) == "ANGGOTA" {
		h.anggotaView(c)
		return
	}
	h.staffView(c)
}

// ===== Staff (OWNER/KASIR) — operational overview =====
func (h *DashboardHandler) staffView(c *gin.Context) {
	anggotaAktif := 0
	anggotaPending := 0
	totalSimpanan := 0.0
	for _, m := range mock.Members {
		switch m.Status {
		case "AKTIF":
			anggotaAktif++
		case "PENDING":
			anggotaPending++
		}
		totalSimpanan += m.SimpananPokok + m.SimpananWajib + m.SimpananSukarela
	}

	pinjamanAktif := 0
	pinjamanPending := 0
	for _, l := range mock.Loans {
		switch l.Status {
		case "AKTIF":
			pinjamanAktif++
		case "PENDING":
			pinjamanPending++
		}
	}

	produkPending := 0
	produkStokRendah := 0
	for _, p := range mock.Products {
		if p.Status == "PENDING" {
			produkPending++
		}
		if p.Stok < p.BatasStokMinimum {
			produkStokRendah++
		}
	}

	orderHariIni := 0
	orderDisputed := 0
	totalOmzet := 0.0
	for _, o := range mock.Orders {
		if len(o.CreatedAt) >= 10 && o.CreatedAt[:10] == "2026-04-22" {
			orderHariIni++
		}
		if o.Status == "DISPUTED" {
			orderDisputed++
		}
		if o.Status == "SELESAI" {
			totalOmzet += o.TotalHarga
		}
	}

	activities := []dashboardActivity{}
	for i := len(mock.Orders) - 1; i >= 0 && len(activities) < 4; i-- {
		o := mock.Orders[i]
		activities = append(activities, dashboardActivity{
			Waktu:     o.CreatedAt,
			Aktivitas: "Order " + o.NomorOrder + " (" + o.PembeliNama + ") — " + o.Status,
			Nominal:   o.TotalHarga,
		})
	}

	h.Render(c, "base", "dashboard/index", gin.H{
		"Title":            "Dashboard",
		"Active":           "dashboard",
		"AnggotaAktif":     anggotaAktif,
		"AnggotaPending":   anggotaPending,
		"TotalSimpanan":    totalSimpanan,
		"PinjamanAktif":    pinjamanAktif,
		"PinjamanPending":  pinjamanPending,
		"ProdukPending":    produkPending,
		"ProdukStokRendah": produkStokRendah,
		"OrderHariIni":     orderHariIni,
		"OrderDisputed":    orderDisputed,
		"TotalOmzet":       totalOmzet,
		"Activities":       activities,
	})
}

// ===== ANGGOTA — personal self-service view =====
func (h *DashboardHandler) anggotaView(c *gin.Context) {
	userNama := CurrentUserNama(c)

	var member *model.Member
	for i := range mock.Members {
		if mock.Members[i].Nama == userNama {
			member = &mock.Members[i]
			break
		}
	}

	totalSimpanan := 0.0
	if member != nil {
		totalSimpanan = member.SimpananPokok + member.SimpananWajib + member.SimpananSukarela
	}

	pinjamanAktif := 0
	pinjamanPending := 0
	sisaPinjaman := 0.0
	for _, l := range mock.Loans {
		if l.MemberNama != userNama {
			continue
		}
		switch l.Status {
		case "AKTIF":
			pinjamanAktif++
			sisaPinjaman += l.SisaPokok
		case "PENDING":
			pinjamanPending++
		}
	}

	orderAktif := 0
	orderSelesai := 0
	recentOrders := []model.Order{}
	for i := len(mock.Orders) - 1; i >= 0; i-- {
		o := mock.Orders[i]
		if o.PembeliNama != userNama {
			continue
		}
		switch o.Status {
		case "SELESAI":
			orderSelesai++
		case "BATAL":
		default:
			orderAktif++
		}
		if len(recentOrders) < 5 {
			recentOrders = append(recentOrders, o)
		}
	}

	h.Render(c, "base", "dashboard/anggota", gin.H{
		"Title":           "Dashboard",
		"Active":          "dashboard",
		"Member":          member,
		"TotalSimpanan":   totalSimpanan,
		"PinjamanAktif":   pinjamanAktif,
		"PinjamanPending": pinjamanPending,
		"SisaPinjaman":    sisaPinjaman,
		"OrderAktif":      orderAktif,
		"OrderSelesai":    orderSelesai,
		"RecentOrders":    recentOrders,
	})
}
