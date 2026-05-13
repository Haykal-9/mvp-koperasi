package ecommerce

import (
	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/mock"
)

// BuyerHandler handles buyer dashboard pages.
type BuyerHandler struct {
	Render ECRenderer
}

func NewBuyerHandler(render ECRenderer) *BuyerHandler {
	return &BuyerHandler{Render: render}
}

// Dashboard shows the buyer's main dashboard.
// GET /ecommerce/buyer
func (h *BuyerHandler) Dashboard(c *gin.Context) {
	ecUserID := GetECUserID(c)

	recentOrders := mock.GetECUserOrders(ecUserID)
	if len(recentOrders) > 5 {
		recentOrders = recentOrders[:5]
	}

	points := mock.GetECUserPoints(ecUserID)
	pointsBalance := 0.0
	if points != nil {
		pointsBalance = points.Balance
	}

	addresses := mock.GetECUserAddresses(ecUserID)
	wishlist := mock.GetECUserWishlist(ecUserID)

	// Recommended products (mock: just show featured)
	recommended := mock.GetAllApprovedECProducts()
	if len(recommended) > 4 {
		recommended = recommended[:4]
	}

	// Check koperasi link
	user := mock.FindECommerceUserByID(ecUserID)
	linkedMember := ""
	if user != nil && user.LinkedKoperasiMemberID > 0 {
		m := mock.FindMemberByID(user.LinkedKoperasiMemberID)
		if m != nil {
			linkedMember = m.Nama + " (" + m.NomorAnggota + ")"
		}
	}

	h.Render(c, "ec_base", "ecommerce/buyer/dashboard", gin.H{
		"Title":          "Dashboard Buyer",
		"Active":         "buyer",
		"RecentOrders":   recentOrders,
		"PointsBalance":  pointsBalance,
		"Addresses":      addresses,
		"WishlistCount":  len(wishlist),
		"Recommended":    recommended,
		"LinkedMember":   linkedMember,
		"OrderCount":     len(mock.GetECUserOrders(ecUserID)),
	})
}
