package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/service"
)

type DashboardHandler struct {
	Render Renderer
	Svc    *service.DashboardService
}

func NewDashboardHandler(render Renderer, svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{Render: render, Svc: svc}
}

// Index routes the dashboard view by role:
//
//	OWNER / KASIR  -> full operational dashboard
//	ANGGOTA        -> personal self-service dashboard
func (h *DashboardHandler) Index(c *gin.Context) {
	if CurrentUserRole(c) == "ANGGOTA" {
		h.anggotaView(c)
		return
	}
	h.staffView(c)
}

// ===== Staff (OWNER/KASIR) — operational overview =====
func (h *DashboardHandler) staffView(c *gin.Context) {
	if h.Svc == nil {
		h.Render(c, "base", "dashboard/index", gin.H{"Title": "Dashboard", "Active": "dashboard"})
		return
	}
	d, err := h.Svc.Staff(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "base", "dashboard/index", gin.H{
		"Title":            "Dashboard",
		"Active":           "dashboard",
		"AnggotaAktif":     d.AnggotaAktif,
		"AnggotaPending":   d.AnggotaPending,
		"TotalSimpanan":    d.TotalSimpanan,
		"PinjamanAktif":    d.PinjamanAktif,
		"PinjamanPending":  d.PinjamanPending,
		"ProdukPending":    d.ProdukPending,
		"ProdukStokRendah": d.ProdukStokRendah,
		"OrderHariIni":     d.OrderHariIni,
		"OrderDisputed":    d.OrderDisputed,
		"TotalOmzet":       d.TotalOmzet,
		"Activities":       d.Activities,
	})
}

// ===== ANGGOTA — personal self-service view =====
func (h *DashboardHandler) anggotaView(c *gin.Context) {
	if h.Svc == nil {
		h.Render(c, "base", "dashboard/anggota", gin.H{"Title": "Dashboard", "Active": "dashboard"})
		return
	}
	d, err := h.Svc.Anggota(c.Request.Context(), CurrentUserNama(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "base", "dashboard/anggota", gin.H{
		"Title":           "Dashboard",
		"Active":          "dashboard",
		"Member":          d.Member,
		"TotalSimpanan":   d.TotalSimpanan,
		"PinjamanAktif":   d.PinjamanAktif,
		"PinjamanPending": d.PinjamanPending,
		"SisaPinjaman":    d.SisaPinjaman,
		"OrderAktif":      d.OrderAktif,
		"OrderSelesai":    d.OrderSelesai,
		"RecentOrders":    d.RecentOrders,
	})
}
