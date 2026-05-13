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
)

// ECRenderer is the function the e-commerce handlers use to render templates.
type ECRenderer func(c *gin.Context, layout, page string, data gin.H)

// AuthHandler handles e-commerce authentication (login, signup, logout).
type AuthHandler struct {
	Render ECRenderer
}

// NewAuthHandler creates a new e-commerce auth handler.
func NewAuthHandler(render ECRenderer) *AuthHandler {
	return &AuthHandler{Render: render}
}

// Login renders the e-commerce login page.
// GET /ecommerce/login
func (h *AuthHandler) Login(c *gin.Context) {
	sess := sessions.Default(c)

	// Already logged in → redirect to buyer dashboard
	if sess.Get("ec_user_id") != nil {
		c.Redirect(http.StatusFound, "/ecommerce/buyer")
		return
	}

	flashErr, _ := handler.PopFlash(c, "ec_error")
	flashOk, _ := handler.PopFlash(c, "ec_success")
	prefillUsername, _ := handler.PopFlash(c, "ec_form_username")

	h.Render(c, "ec_auth", "ecommerce/auth/login", gin.H{
		"Title":        "E-Commerce Login",
		"FlashError":   flashErr,
		"FlashSuccess": flashOk,
		"FormUsername":  prefillUsername,
	})
}

// DoLogin processes the e-commerce login form.
// POST /ecommerce/login
func (h *AuthHandler) DoLogin(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")

	if username == "" || password == "" {
		handler.SetFlash(c, "ec_error", "Username dan password wajib diisi.")
		handler.SetFlash(c, "ec_form_username", username)
		c.Redirect(http.StatusFound, "/ecommerce/login")
		return
	}

	user, ok := mock.AuthenticateECommerceUser(username, password)
	if !ok {
		handler.SetFlash(c, "ec_error", "Username atau password salah.")
		handler.SetFlash(c, "ec_form_username", username)
		c.Redirect(http.StatusFound, "/ecommerce/login")
		return
	}

	// Create e-commerce session
	if err := SetECSession(c, user.ID, user.Username, user.Email, user.Role, user.IsSellerActive); err != nil {
		handler.SetFlash(c, "ec_error", "Gagal menyimpan sesi. Coba lagi.")
		c.Redirect(http.StatusFound, "/ecommerce/login")
		return
	}

	// Flash message based on koperasi link status
	if user.LinkedKoperasiMemberID > 0 {
		m := mock.FindMemberByID(user.LinkedKoperasiMemberID)
		if m != nil {
			handler.SetFlash(c, "ec_success", fmt.Sprintf("Selamat datang, %s! Akun Koperasi Anda (%s) sudah linked. Poin bisa dikonversi ke Simpanan.", user.Username, m.Nama))
		}
	} else {
		handler.SetFlash(c, "ec_success", fmt.Sprintf("Selamat datang, %s! Tip: Link dengan akun Koperasi di halaman Poin untuk konversi poin ke Simpanan.", user.Username))
	}

	// Redirect based on role
	if user.Role == "ADMIN" {
		c.Redirect(http.StatusFound, "/ecommerce/buyer")
		return
	}
	if user.IsSellerActive {
		c.Redirect(http.StatusFound, "/ecommerce/seller")
		return
	}
	c.Redirect(http.StatusFound, "/ecommerce/buyer")
}

// Signup renders the e-commerce signup page.
// GET /ecommerce/signup
func (h *AuthHandler) Signup(c *gin.Context) {
	sess := sessions.Default(c)
	if sess.Get("ec_user_id") != nil {
		c.Redirect(http.StatusFound, "/ecommerce/buyer")
		return
	}

	flashErr, _ := handler.PopFlash(c, "ec_error")
	prefillUsername, _ := handler.PopFlash(c, "ec_form_username")
	prefillEmail, _ := handler.PopFlash(c, "ec_form_email")

	h.Render(c, "ec_auth", "ecommerce/auth/signup", gin.H{
		"Title":        "Daftar E-Commerce",
		"FlashError":   flashErr,
		"FormUsername":  prefillUsername,
		"FormEmail":    prefillEmail,
	})
}

// DoSignup processes the e-commerce signup form.
// POST /ecommerce/signup
func (h *AuthHandler) DoSignup(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")
	koperasiMemberIDStr := c.PostForm("koperasi_member_id")

	preserve := func() {
		handler.SetFlash(c, "ec_form_username", username)
		handler.SetFlash(c, "ec_form_email", email)
	}

	// Validation
	if username == "" || email == "" || password == "" {
		handler.SetFlash(c, "ec_error", "Username, email, dan password wajib diisi.")
		preserve()
		c.Redirect(http.StatusFound, "/ecommerce/signup")
		return
	}
	if len(password) < 6 {
		handler.SetFlash(c, "ec_error", "Password minimal 6 karakter.")
		preserve()
		c.Redirect(http.StatusFound, "/ecommerce/signup")
		return
	}
	if password != confirmPassword {
		handler.SetFlash(c, "ec_error", "Konfirmasi password tidak cocok.")
		preserve()
		c.Redirect(http.StatusFound, "/ecommerce/signup")
		return
	}
	if mock.FindECommerceUserByUsername(username) != nil {
		handler.SetFlash(c, "ec_error", "Username sudah digunakan.")
		preserve()
		c.Redirect(http.StatusFound, "/ecommerce/signup")
		return
	}

	// Create user
	newUser := mock.CreateECommerceUser(username, email, password)

	// Optional: Link to koperasi member
	if koperasiMemberIDStr != "" {
		memberID, err := strconv.Atoi(koperasiMemberIDStr)
		if err == nil && memberID > 0 {
			if mock.LinkToKoperasiMember(newUser.ID, memberID) {
				m := mock.FindMemberByID(memberID)
				if m != nil {
					handler.SetFlash(c, "ec_success", fmt.Sprintf("Akun berhasil dibuat & linked dengan member %s (%s). Poin siap dikonversi ke Simpanan!", m.Nama, m.NomorAnggota))
				}
			}
		}
	}

	// Create session
	if err := SetECSession(c, newUser.ID, newUser.Username, newUser.Email, newUser.Role, newUser.IsSellerActive); err != nil {
		handler.SetFlash(c, "ec_error", "Gagal menyimpan sesi. Coba login.")
		c.Redirect(http.StatusFound, "/ecommerce/login")
		return
	}

	if handler.PopFlashCheck(c, "ec_success") == "" {
		handler.SetFlash(c, "ec_success", fmt.Sprintf("Selamat datang, %s! Akun E-Commerce berhasil dibuat.", username))
	}

	c.Redirect(http.StatusFound, "/ecommerce/buyer")
}

// SearchKoperasiMember returns JSON search results for koperasi members.
// GET /ecommerce/api/members/search?q=...
func (h *AuthHandler) SearchKoperasiMember(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusOK, []gin.H{})
		return
	}

	members := mock.SearchKoperasiMembers(q)
	results := make([]gin.H, 0, len(members))
	for _, m := range members {
		results = append(results, gin.H{
			"id":            m.ID,
			"nama":          m.Nama,
			"nomor_anggota": m.NomorAnggota,
			"status":        m.Status,
		})
	}
	c.JSON(http.StatusOK, gin.H{"results": results, "query": q})
}

// LinkKoperasiMember links the current EC user to a koperasi member.
// POST /ecommerce/link-member
func (h *AuthHandler) LinkKoperasiMember(c *gin.Context) {
	ecUserID := GetECUserID(c)
	if ecUserID == 0 {
		handler.SetFlash(c, "ec_error", "Silakan login terlebih dahulu.")
		c.Redirect(http.StatusFound, "/ecommerce/login")
		return
	}

	memberIDStr := c.PostForm("koperasi_member_id")
	memberID, err := strconv.Atoi(memberIDStr)
	if err != nil || memberID <= 0 {
		handler.SetFlash(c, "ec_error", "Member ID tidak valid.")
		c.Redirect(http.StatusFound, "/ecommerce/points")
		return
	}

	// Check if member is already linked to another EC user
	for _, u := range mock.ECommerceUsers {
		if u.LinkedKoperasiMemberID == memberID && u.ID != ecUserID {
			handler.SetFlash(c, "ec_error", "Member koperasi ini sudah di-link ke akun E-Commerce lain.")
			c.Redirect(http.StatusFound, "/ecommerce/points")
			return
		}
	}

	if !mock.LinkToKoperasiMember(ecUserID, memberID) {
		handler.SetFlash(c, "ec_error", "Gagal link ke member koperasi. Pastikan member valid.")
		c.Redirect(http.StatusFound, "/ecommerce/points")
		return
	}

	m := mock.FindMemberByID(memberID)
	name := "member"
	if m != nil {
		name = m.Nama
	}
	handler.SetFlash(c, "ec_success", fmt.Sprintf("Berhasil link dengan member %s. Simpanan siap menerima poin konversi!", name))
	c.Redirect(http.StatusFound, "/ecommerce/points")
}

// UnlinkKoperasiMember removes the koperasi member link from the current EC user.
// POST /ecommerce/unlink-member
func (h *AuthHandler) UnlinkKoperasiMember(c *gin.Context) {
	ecUserID := GetECUserID(c)
	if ecUserID == 0 {
		c.Redirect(http.StatusFound, "/ecommerce/login")
		return
	}

	u := mock.FindECommerceUserByID(ecUserID)
	if u == nil {
		c.Redirect(http.StatusFound, "/ecommerce/login")
		return
	}

	oldMemberID := u.LinkedKoperasiMemberID
	u.LinkedKoperasiMemberID = 0

	mock.LogAuditAction("UNLINK_KOPERASI", ecUserID, u.Username,
		fmt.Sprintf("member:%d", oldMemberID), "Unlinked from koperasi member")

	handler.SetFlash(c, "ec_success", "Berhasil unlink dari member koperasi. Konversi poin dinonaktifkan.")
	c.Redirect(http.StatusFound, "/ecommerce/points")
}

// Logout clears the e-commerce session.
// GET /ecommerce/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Delete("ec_user_id")
	sess.Delete("ec_username")
	sess.Delete("ec_email")
	sess.Delete("ec_role")
	sess.Delete("ec_is_seller")
	_ = sess.Save()

	handler.SetFlash(c, "ec_success", "Anda telah keluar dari E-Commerce.")
	c.Redirect(http.StatusFound, "/ecommerce/login")
}

// ---------- E-Commerce Session Helpers ----------

// SetECSession stores e-commerce session data.
func SetECSession(c *gin.Context, userID int, username, email, role string, isSeller bool) error {
	sess := sessions.Default(c)
	sess.Set("ec_user_id", userID)
	sess.Set("ec_username", username)
	sess.Set("ec_email", email)
	sess.Set("ec_role", role)
	sess.Set("ec_is_seller", isSeller)
	return sess.Save()
}

// GetECUserID returns the e-commerce user ID from session, or 0 if not logged in.
func GetECUserID(c *gin.Context) int {
	sess := sessions.Default(c)
	v := sess.Get("ec_user_id")
	if v == nil {
		return 0
	}
	id, ok := v.(int)
	if !ok {
		return 0
	}
	return id
}

// GetECUsername returns the e-commerce username from session.
func GetECUsername(c *gin.Context) string {
	sess := sessions.Default(c)
	v := sess.Get("ec_username")
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

// GetECRole returns the e-commerce role from session.
func GetECRole(c *gin.Context) string {
	sess := sessions.Default(c)
	v := sess.Get("ec_role")
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

// IsECSeller returns whether the current e-commerce user is a seller.
func IsECSeller(c *gin.Context) bool {
	sess := sessions.Default(c)
	v := sess.Get("ec_is_seller")
	if v == nil {
		return false
	}
	b, _ := v.(bool)
	return b
}
