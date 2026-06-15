package ecommerce

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/service"
)

// BuyerHandler handles buyer dashboard pages.
type BuyerHandler struct {
	Render ECRenderer
	Svc    *service.ECAccountService
}

func NewBuyerHandler(render ECRenderer, svc *service.ECAccountService) *BuyerHandler {
	return &BuyerHandler{Render: render, Svc: svc}
}

func (h *BuyerHandler) ready(c *gin.Context) bool {
	if h.Svc == nil {
		c.String(http.StatusServiceUnavailable, "E-Commerce sementara tidak tersedia (database tidak terhubung).")
		return false
	}
	return true
}

// Dashboard shows the buyer's main dashboard.
// GET /ecommerce/buyer
func (h *BuyerHandler) Dashboard(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)

	d, err := h.Svc.BuyerDashboard(c.Request.Context(), ecUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/buyer/dashboard", gin.H{
		"Title":         "Dashboard Buyer",
		"Active":        "buyer",
		"RecentOrders":  d.RecentOrders,
		"PointsBalance": d.PointsBalance,
		"Addresses":     d.Addresses,
		"WishlistCount": d.WishlistCount,
		"Recommended":   d.Recommended,
		"LinkedMember":  d.LinkedMember,
		"OrderCount":    d.OrderCount,
	})
}
