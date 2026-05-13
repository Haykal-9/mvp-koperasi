package ecommerce

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/handler"
	"koperasi-frontend/core/mock"
)

// PointsHandler handles loyalty points operations.
type PointsHandler struct {
	Render ECRenderer
}

func NewPointsHandler(render ECRenderer) *PointsHandler {
	return &PointsHandler{Render: render}
}

// Balance shows the user's points balance, linked member, and transaction history.
// GET /ecommerce/points
func (h *PointsHandler) Balance(c *gin.Context) {
	ecUserID := GetECUserID(c)

	points := mock.GetECUserPoints(ecUserID)
	balance := 0.0
	totalEarned := 0.0
	totalRedeemed := 0.0
	if points != nil {
		balance = points.Balance
		totalEarned = points.TotalEarned
		totalRedeemed = points.TotalRedeemed
	}

	transactions := mock.GetECPointsTransactions(ecUserID)
	linkedMember := mock.GetLinkedKoperasiMember(ecUserID)

	// Rp equivalent (1 poin = Rp 100)
	rpEquivalent := balance * 100

	h.Render(c, "ec_base", "ecommerce/points/balance", gin.H{
		"Title":         "Poin & Reward",
		"Active":        "points",
		"Balance":       balance,
		"TotalEarned":   totalEarned,
		"TotalRedeemed": totalRedeemed,
		"RpEquivalent":  rpEquivalent,
		"Transactions":  transactions,
		"LinkedMember":  linkedMember,
	})
}

// ConvertForm shows the points-to-simpanan conversion page.
// GET /ecommerce/points/convert
func (h *PointsHandler) ConvertForm(c *gin.Context) {
	ecUserID := GetECUserID(c)

	points := mock.GetECUserPoints(ecUserID)
	balance := 0.0
	if points != nil {
		balance = points.Balance
	}

	linkedMember := mock.GetLinkedKoperasiMember(ecUserID)

	h.Render(c, "ec_base", "ecommerce/points/convert", gin.H{
		"Title":        "Konversi Poin ke Simpanan",
		"Active":       "points",
		"Balance":      balance,
		"LinkedMember": linkedMember,
	})
}

// DoConvert processes the points conversion to koperasi simpanan.
// POST /ecommerce/points/convert
func (h *PointsHandler) DoConvert(c *gin.Context) {
	ecUserID := GetECUserID(c)

	amountStr := c.PostForm("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount < 100 {
		handler.SetFlash(c, "ec_error", "Minimum konversi 100 poin.")
		c.Redirect(http.StatusFound, "/ecommerce/points/convert")
		return
	}

	ok, msg := mock.ConvertPointsToKoperasiSimpanan(ecUserID, amount)
	if !ok {
		handler.SetFlash(c, "ec_error", msg)
		c.Redirect(http.StatusFound, "/ecommerce/points/convert")
		return
	}

	handler.SetFlash(c, "ec_success", msg)
	c.Redirect(http.StatusFound, "/ecommerce/points")
}

// LinkMemberPage shows the koperasi member linking page (from points).
// GET /ecommerce/points/link-member
func (h *PointsHandler) LinkMemberPage(c *gin.Context) {
	ecUserID := GetECUserID(c)
	linkedMember := mock.GetLinkedKoperasiMember(ecUserID)

	h.Render(c, "ec_base", "ecommerce/points/link_member", gin.H{
		"Title":        "Link Member Koperasi",
		"Active":       "points",
		"LinkedMember": linkedMember,
	})
}

// DoLinkMember processes koperasi member linking from points page.
// POST /ecommerce/points/link-member
func (h *PointsHandler) DoLinkMember(c *gin.Context) {
	ecUserID := GetECUserID(c)

	memberIDStr := c.PostForm("koperasi_member_id")
	memberID, err := strconv.Atoi(memberIDStr)
	if err != nil || memberID <= 0 {
		handler.SetFlash(c, "ec_error", "Member ID tidak valid.")
		c.Redirect(http.StatusFound, "/ecommerce/points/link-member")
		return
	}

	// Check not already linked by another user
	for _, u := range mock.ECommerceUsers {
		if u.LinkedKoperasiMemberID == memberID && u.ID != ecUserID {
			handler.SetFlash(c, "ec_error", "Member ini sudah di-link ke akun lain.")
			c.Redirect(http.StatusFound, "/ecommerce/points/link-member")
			return
		}
	}

	if !mock.LinkToKoperasiMember(ecUserID, memberID) {
		handler.SetFlash(c, "ec_error", "Gagal link member.")
		c.Redirect(http.StatusFound, "/ecommerce/points/link-member")
		return
	}

	m := mock.FindMemberByID(memberID)
	name := "member"
	if m != nil {
		name = m.Nama
	}
	handler.SetFlash(c, "ec_success", fmt.Sprintf("Berhasil link dengan %s! Poin siap dikonversi.", name))
	c.Redirect(http.StatusFound, "/ecommerce/points")
}

// CheckConversionStatus returns JSON status of recent conversions.
// GET /ecommerce/api/points/conversion-status
func (h *PointsHandler) CheckConversionStatus(c *gin.Context) {
	ecUserID := GetECUserID(c)
	transactions := mock.GetECPointsTransactions(ecUserID)

	var conversions []gin.H
	for _, t := range transactions {
		if t.Tipe == "CONVERT_SIMPANAN" {
			conversions = append(conversions, gin.H{
				"amount":     t.Amount,
				"date":       t.CreatedAt,
				"keterangan": t.Keterangan,
				"status":     "CONFIRMED",
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":     ecUserID,
		"conversions": conversions,
	})
}
