package ecommerce

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/handler"
	"koperasi-frontend/core/service"
)

// PointsHandler handles loyalty points operations.
type PointsHandler struct {
	Render ECRenderer
	Svc    *service.ECAccountService
}

func NewPointsHandler(render ECRenderer, svc *service.ECAccountService) *PointsHandler {
	return &PointsHandler{Render: render, Svc: svc}
}

func (h *PointsHandler) ready(c *gin.Context) bool {
	if h.Svc == nil {
		c.String(http.StatusServiceUnavailable, "E-Commerce sementara tidak tersedia (database tidak terhubung).")
		return false
	}
	return true
}

// Balance shows the user's points balance, linked member, and transaction history.
// GET /ecommerce/points
func (h *PointsHandler) Balance(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)

	o, err := h.Svc.PointsOverview(c.Request.Context(), ecUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/points/balance", gin.H{
		"Title":         "Poin & Reward",
		"Active":        "points",
		"Balance":       o.Balance,
		"TotalEarned":   o.TotalEarned,
		"TotalRedeemed": o.TotalRedeemed,
		"RpEquivalent":  o.RpEquivalent,
		"PointValue":    service.PointValueRupiah,
		"Transactions":  o.Transactions,
		"LinkedMember":  o.LinkedMember,
	})
}

// ConvertForm shows the points-to-simpanan conversion page.
// GET /ecommerce/points/convert
func (h *PointsHandler) ConvertForm(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)

	o, err := h.Svc.PointsOverview(c.Request.Context(), ecUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/points/convert", gin.H{
		"Title":            "Konversi Poin ke Simpanan",
		"Active":           "points",
		"Balance":          o.Balance,
		"PointValue":       service.PointValueRupiah,
		"MinConvertPoints": service.MinPointConversion,
		"MinConvertRupiah": service.MinPointConversion * service.PointValueRupiah,
		"LinkedMember":     o.LinkedMember,
	})
}

// DoConvert processes the points conversion to koperasi simpanan.
// POST /ecommerce/points/convert
func (h *PointsHandler) DoConvert(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)

	amount, err := strconv.ParseFloat(c.PostForm("amount"), 64)
	if err != nil || amount < service.MinPointConversion {
		handler.SetFlash(c, "ec_error", fmt.Sprintf("Minimum konversi %.0f poin.", service.MinPointConversion))
		c.Redirect(http.StatusFound, "/ecommerce/points/convert")
		return
	}

	msg, _, _, convErr := h.Svc.ConvertToSimpanan(c.Request.Context(), ecUserID, amount)
	if convErr != nil {
		handler.SetFlash(c, "ec_error", convErr.Error())
		c.Redirect(http.StatusFound, "/ecommerce/points/convert")
		return
	}

	handler.SetFlash(c, "ec_success", msg)
	c.Redirect(http.StatusFound, "/ecommerce/points")
}

// invalidateECSession clears the e-commerce session keys. Used when the
// session references an EC user that no longer exists.
func invalidateECSession(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Delete("ec_user_id")
	sess.Delete("ec_username")
	sess.Delete("ec_email")
	sess.Delete("ec_role")
	sess.Delete("ec_is_seller")
	_ = sess.Save()
}

// LinkMemberPage menampilkan form pendaftaran sebagai anggota koperasi baru.
// GET /ecommerce/points/link-member
func (h *PointsHandler) LinkMemberPage(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)

	ecUser, err := h.Svc.UserByID(c.Request.Context(), ecUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if ecUser == nil {
		invalidateECSession(c)
		handler.SetFlash(c, "ec_error", "Sesi tidak valid. Silakan login kembali.")
		c.Redirect(http.StatusFound, "/ecommerce/login")
		return
	}

	linkedMember, err := h.Svc.LinkedMember(c.Request.Context(), ecUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	// Preserve form values on validation error
	formNama, _ := handler.PopFlash(c, "form_nama")
	formNIK, _ := handler.PopFlash(c, "form_nik")
	formAlamat, _ := handler.PopFlash(c, "form_alamat")
	formNoHP, _ := handler.PopFlash(c, "form_nohp")
	if formNama == "" {
		formNama = ecUser.Username // default nama dari username EC
	}

	h.Render(c, "ec_base", "ecommerce/points/link_member", gin.H{
		"Title":        "Daftar Jadi Anggota Koperasi",
		"Active":       "points",
		"LinkedMember": linkedMember,
		"FormNama":     formNama,
		"FormNIK":      formNIK,
		"FormAlamat":   formAlamat,
		"FormNoHP":     formNoHP,
	})
}

// DoLinkMember memproses pendaftaran anggota koperasi baru dari EC, lalu
// auto-link akun EC ke Member record yang baru dibuat (status PENDING).
// POST /ecommerce/points/link-member
func (h *PointsHandler) DoLinkMember(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)
	if ecUserID == 0 {
		handler.SetFlash(c, "ec_error", "Silakan login terlebih dahulu.")
		c.Redirect(http.StatusFound, "/ecommerce/login")
		return
	}

	nama := strings.TrimSpace(c.PostForm("nama"))
	nik := strings.TrimSpace(c.PostForm("nik"))
	alamat := strings.TrimSpace(c.PostForm("alamat"))
	noHP := strings.TrimSpace(c.PostForm("no_hp"))

	preserve := func() {
		handler.SetFlash(c, "form_nama", nama)
		handler.SetFlash(c, "form_nik", nik)
		handler.SetFlash(c, "form_alamat", alamat)
		handler.SetFlash(c, "form_nohp", noHP)
	}

	// Validation: semua field wajib
	if nama == "" || nik == "" || alamat == "" || noHP == "" {
		handler.SetFlash(c, "ec_error", "Semua field wajib diisi.")
		preserve()
		c.Redirect(http.StatusFound, "/ecommerce/points/link-member")
		return
	}
	if len(nik) != 16 {
		handler.SetFlash(c, "ec_error", "NIK harus 16 digit.")
		preserve()
		c.Redirect(http.StatusFound, "/ecommerce/points/link-member")
		return
	}
	for _, r := range nik {
		if r < '0' || r > '9' {
			handler.SetFlash(c, "ec_error", "NIK hanya boleh angka.")
			preserve()
			c.Redirect(http.StatusFound, "/ecommerce/points/link-member")
			return
		}
	}

	err := h.Svc.RegisterMemberAsKoperasi(c.Request.Context(), ecUserID, nama, nik, alamat, noHP)
	if errors.Is(err, service.ErrECUserNotFound) {
		invalidateECSession(c)
		handler.SetFlash(c, "ec_error", "Sesi tidak valid. Silakan login kembali.")
		c.Redirect(http.StatusFound, "/ecommerce/login")
		return
	}
	if err != nil {
		handler.SetFlash(c, "ec_error", err.Error())
		preserve()
		c.Redirect(http.StatusFound, "/ecommerce/points/link-member")
		return
	}

	handler.SetFlash(c, "ec_success", fmt.Sprintf(
		"Pendaftaran sebagai anggota koperasi terkirim. Status: PENDING — menunggu review pengurus. Setelah disetujui, %s bisa konversi poin ke simpanan.",
		nama,
	))
	c.Redirect(http.StatusFound, "/ecommerce/points")
}

// CheckConversionStatus returns JSON status of recent conversions.
// GET /ecommerce/api/points/conversion-status
func (h *PointsHandler) CheckConversionStatus(c *gin.Context) {
	if h.Svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}
	ecUserID := GetECUserID(c)
	transactions, err := h.Svc.PointsTransactions(c.Request.Context(), ecUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kesalahan database"})
		return
	}

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
