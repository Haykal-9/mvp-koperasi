package handler

import (
	"errors"
	"net/http"
	"strings"

	"koperasi-frontend/core/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Renderer is the function the auth handler uses to render templates.
// Defined as a function variable so the package doesn't depend on main.
type Renderer func(c *gin.Context, layout, page string, data gin.H)

type AuthHandler struct {
	Render Renderer
	Svc    *service.AuthService
}

func NewAuthHandler(render Renderer, svc *service.AuthService) *AuthHandler {
	return &AuthHandler{Render: render, Svc: svc}
}

func (h *AuthHandler) ready(c *gin.Context) bool {
	if h.Svc == nil {
		c.String(http.StatusServiceUnavailable, "Autentikasi sementara tidak tersedia (database tidak terhubung).")
		return false
	}
	return true
}

// ===== GET /login =====
func (h *AuthHandler) ShowLogin(c *gin.Context) {
	sess := sessions.Default(c)

	// already logged in -> dashboard
	if sess.Get("user_id") != nil {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	flashErr, _ := PopFlash(c, "error")
	flashOk, _ := PopFlash(c, "success")
	prefillEmail, _ := PopFlash(c, "form_email")

	h.Render(c, "auth", "auth/login", gin.H{
		"Title":        "Login",
		"FlashError":   flashErr,
		"FlashSuccess": flashOk,
		"FormEmail":    prefillEmail,
	})
}

// ===== POST /login =====
func (h *AuthHandler) DoLogin(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")

	if email == "" || password == "" {
		SetFlash(c, "error", "Email dan password wajib diisi.")
		SetFlash(c, "form_email", email)
		c.Redirect(http.StatusFound, "/login")
		return
	}

	user, err := h.Svc.Authenticate(c.Request.Context(), email, password)
	if err != nil {
		SetFlash(c, "error", "Kesalahan server. Coba lagi.")
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if user == nil {
		SetFlash(c, "error", "Email atau password salah.")
		SetFlash(c, "form_email", email)
		c.Redirect(http.StatusFound, "/login")
		return
	}

	sess := sessions.Default(c)
	sess.Set("user_id", user.ID)
	sess.Set("user_email", user.Email)
	sess.Set("user_nama", user.Nama)
	sess.Set("user_role", user.Role)

	if err := sess.Save(); err != nil {
		SetFlash(c, "error", "Gagal menyimpan sesi. Coba lagi.")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.Redirect(http.StatusFound, "/dashboard")
}

// ===== GET /register =====
func (h *AuthHandler) ShowRegister(c *gin.Context) {
	sess := sessions.Default(c)
	if sess.Get("user_id") != nil {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	flashErr, _ := PopFlash(c, "error")
	prefillNama, _ := PopFlash(c, "form_nama")
	prefillEmail, _ := PopFlash(c, "form_email")

	h.Render(c, "auth", "auth/register", gin.H{
		"Title":      "Register",
		"FlashError": flashErr,
		"FormNama":   prefillNama,
		"FormEmail":  prefillEmail,
	})
}

// ===== POST /register =====
func (h *AuthHandler) DoRegister(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	nama := strings.TrimSpace(c.PostForm("nama"))
	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")

	preserve := func() {
		SetFlash(c, "form_nama", nama)
		SetFlash(c, "form_email", email)
	}

	if nama == "" || email == "" || password == "" {
		SetFlash(c, "error", "Nama, email, dan password wajib diisi.")
		preserve()
		c.Redirect(http.StatusFound, "/register")
		return
	}
	if len(password) < 6 {
		SetFlash(c, "error", "Password minimal 6 karakter.")
		preserve()
		c.Redirect(http.StatusFound, "/register")
		return
	}
	if _, err := h.Svc.Register(c.Request.Context(), nama, email, password); err != nil {
		if errors.Is(err, service.ErrEmailAlreadyRegistered) {
			SetFlash(c, "error", "Email sudah terdaftar.")
		} else {
			SetFlash(c, "error", "Registrasi gagal. Coba lagi.")
		}
		preserve()
		c.Redirect(http.StatusFound, "/register")
		return
	}

	SetFlash(c, "success", "Registrasi berhasil. Silakan login.")
	SetFlash(c, "form_email", email)
	c.Redirect(http.StatusFound, "/login")
}

// ===== GET /logout =====
func (h *AuthHandler) DoLogout(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Clear() // clears both koperasi and EC sessions
	_ = sess.Save()
	SetFlash(c, "success", "Anda telah keluar.")
	c.Redirect(http.StatusFound, "/login")
}

// ===== AuthRequired middleware =====
// Aborts with redirect to /login if no user in session.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		sess := sessions.Default(c)
		if sess.Get("user_id") == nil {
			SetFlash(c, "error", "Silakan login terlebih dahulu.")
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

// ===== RequireRole middleware =====
// Aborts with 403 redirect if current user's role is not in allowed list.
// MUST be chained after AuthRequired().
func RequireRole(allowed ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := CurrentUserRole(c)
		for _, r := range allowed {
			if r == role {
				c.Next()
				return
			}
		}
		SetFlash(c, "error", "Anda tidak memiliki akses ke halaman tersebut.")
		c.Redirect(http.StatusFound, "/dashboard")
		c.Abort()
	}
}

// CurrentUserRole reads user_role from session. Returns "" if not logged in.
func CurrentUserRole(c *gin.Context) string {
	sess := sessions.Default(c)
	if v := sess.Get("user_role"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// CurrentUserNama reads user_nama from session. Used for filtering
// records that store only the user's name (Order.PembeliNama, Loan.MemberNama).
func CurrentUserNama(c *gin.Context) string {
	sess := sessions.Default(c)
	if v := sess.Get("user_nama"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// CurrentUserID reads user_id from session.
func CurrentUserID(c *gin.Context) int {
	sess := sessions.Default(c)
	if v := sess.Get("user_id"); v != nil {
		if id, ok := v.(int); ok {
			return id
		}
	}
	return 0
}

// IsOwnerOrKasir returns true if the current user can perform staff actions.
func IsOwnerOrKasir(c *gin.Context) bool {
	r := CurrentUserRole(c)
	return r == "OWNER" || r == "KASIR"
}
