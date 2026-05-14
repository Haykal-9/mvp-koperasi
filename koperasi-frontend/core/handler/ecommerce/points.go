package ecommerce

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/handler"
	"koperasi-frontend/core/mock"
	"koperasi-frontend/core/model"
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

// invalidateECSession clears the e-commerce session keys. Used when the
// session references an EC user that no longer exists (e.g. server restart
// in dev wiped in-memory data, but the cookie still holds the old ID).
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
// EC user yang belum jadi anggota mengisi data tambahan (NIK, alamat, no HP)
// untuk membuat record Member berstatus PENDING, sekaligus auto-link ke akun EC.
// GET /ecommerce/points/link-member
func (h *PointsHandler) LinkMemberPage(c *gin.Context) {
	ecUserID := GetECUserID(c)

	// Resolve current EC user explicitly. If the session points to a user
	// that doesn't exist anymore (server restarted, mock data reset),
	// clear the session and bounce to login — otherwise we'd render a form
	// that's guaranteed to fail on submit.
	ecUser := mock.FindECommerceUserByID(ecUserID)
	if ecUser == nil {
		invalidateECSession(c)
		handler.SetFlash(c, "ec_error", "Sesi tidak valid. Silakan login kembali.")
		c.Redirect(http.StatusFound, "/ecommerce/login")
		return
	}

	var linkedMember *model.Member
	if ecUser.LinkedKoperasiMemberID != 0 {
		linkedMember = mock.FindMemberByID(ecUser.LinkedKoperasiMemberID)
	}

	// Default nama dari username EC, untuk pre-fill form
	defaultNama := ecUser.Username

	// Preserve form values on validation error
	formNama, _ := handler.PopFlash(c, "form_nama")
	formNIK, _ := handler.PopFlash(c, "form_nik")
	formAlamat, _ := handler.PopFlash(c, "form_alamat")
	formNoHP, _ := handler.PopFlash(c, "form_nohp")
	if formNama == "" {
		formNama = defaultNama
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
// auto-link akun EC ke Member record yang baru dibuat (status PENDING,
// menunggu approval pengurus sesuai BPMN M1.1).
// POST /ecommerce/points/link-member
func (h *PointsHandler) DoLinkMember(c *gin.Context) {
	ecUserID := GetECUserID(c)
	if ecUserID == 0 {
		handler.SetFlash(c, "ec_error", "Silakan login terlebih dahulu.")
		c.Redirect(http.StatusFound, "/ecommerce/login")
		return
	}

	// Resolve EC user explicitly. Nil here means the session is pointing at
	// a user that no longer exists (typical after dev restart) — bounce to
	// login instead of creating an orphaned Member record.
	ecUser := mock.FindECommerceUserByID(ecUserID)
	if ecUser == nil {
		invalidateECSession(c)
		handler.SetFlash(c, "ec_error", "Sesi tidak valid. Silakan login kembali.")
		c.Redirect(http.StatusFound, "/ecommerce/login")
		return
	}

	// Sudah linked → tolak. Harus unlink dulu untuk daftar ulang.
	if ecUser.LinkedKoperasiMemberID != 0 {
		handler.SetFlash(c, "ec_error", "Akun Anda sudah terhubung ke member koperasi.")
		c.Redirect(http.StatusFound, "/ecommerce/points")
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
	if mock.IsNIKRegistered(nik) {
		handler.SetFlash(c, "ec_error", "NIK sudah terdaftar pada anggota lain.")
		preserve()
		c.Redirect(http.StatusFound, "/ecommerce/points/link-member")
		return
	}

	// Buat Member baru status PENDING. Setelah berhasil tersimpan, link akun
	// EC ke record tersebut. Kalau link gagal (skenario edge — misalnya user
	// dihapus tepat di antara dua call), rollback Member supaya tidak ada
	// record yatim di /members.
	newID := mock.RegisterPendingMember(nama, nik, alamat, noHP)
	if !mock.LinkToKoperasiMember(ecUserID, newID) {
		mock.RemoveMemberByID(newID)
		invalidateECSession(c)
		handler.SetFlash(c, "ec_error", "Sesi tidak valid. Silakan login kembali.")
		c.Redirect(http.StatusFound, "/ecommerce/login")
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
