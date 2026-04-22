package handler

import (
	"koperasi-frontend/internal/mock"

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

func (h *DashboardHandler) Index(c *gin.Context) {
	// stats
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

	// recent activity (last 4 orders by ID desc)
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
