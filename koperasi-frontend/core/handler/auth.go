package handler

import (
	"net/http"
	"strings"

	"koperasi-frontend/core/mock"
	"koperasi-frontend/core/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Renderer is the function the auth handler uses to render templates.
// Defined as a function variable so the package doesn't depend on main.
type Renderer func(c *gin.Context, layout, page string, data gin.H)

type AuthHandler struct {
	Render Renderer
}

func NewAuthHandler(render Renderer) *AuthHandler {
	return &AuthHandler{Render: render}
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
	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")

	if email == "" || password == "" {
		SetFlash(c, "error", "Email dan password wajib diisi.")
		SetFlash(c, "form_email", email)
		c.Redirect(http.StatusFound, "/login")
		return
	}

	user := mock.FindUserByEmail(email)
	if user == nil || user.Password != password {
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
	if mock.FindUserByEmail(email) != nil {
		SetFlash(c, "error", "Email sudah terdaftar.")
		preserve()
		c.Redirect(http.StatusFound, "/register")
		return
	}

	// append to in-memory mock data
	newID := nextUserID()
	mock.Users = append(mock.Users, model.User{
		ID:       newID,
		Email:    email,
		Password: password,
		Nama:     nama,
		Role:     "ANGGOTA",
	})

	SetFlash(c, "success", "Registrasi berhasil. Silakan login.")
	SetFlash(c, "form_email", email)
	c.Redirect(http.StatusFound, "/login")
}

// ===== GET /logout =====
func (h *AuthHandler) DoLogout(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Clear()
	_ = sess.Save()
	SetFlash(c, "success", "Anda telah keluar.")
	c.Redirect(http.StatusFound, "/login")
}

func nextUserID() int {
	max := 0
	for _, u := range mock.Users {
		if u.ID > max {
			max = u.ID
		}
	}
	return max + 1
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
