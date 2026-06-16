package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
	"koperasi-frontend/core/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func TestRequireECommerceAdminRejectsPengurus(t *testing.T) {
	rec := requestRoleProtectedRoute("PENGURUS", RequireECommerceAdmin())

	if rec.Code != http.StatusForbidden {
		t.Fatalf("PENGURUS ke route admin-only = %d, mau %d", rec.Code, http.StatusForbidden)
	}
	if !strings.Contains(rec.Body.String(), "role PENGURUS tidak diizinkan") {
		t.Fatalf("body 403 tidak menjelaskan role: %q", rec.Body.String())
	}
}

func TestRequireECommerceAdminAreaAllowsPengurus(t *testing.T) {
	rec := requestRoleProtectedRoute("PENGURUS", RequireECommerceAdminArea())

	if rec.Code != http.StatusOK {
		t.Fatalf("PENGURUS ke admin area = %d, mau %d", rec.Code, http.StatusOK)
	}
}

type fakeSSORepo struct {
	repository.ECommerceRepository
	called bool
}

func (f *fakeSSORepo) EnsureUserFromKoperasi(_ context.Context, email string) (*model.ECommerceUser, error) {
	f.called = true
	if email != "anggota@koperasi.id" {
		return nil, nil
	}
	return &model.ECommerceUser{
		ID:       9,
		Username: "anggota_demo",
		Email:    email,
		Role:     "BUYER",
	}, nil
}

func TestRequireECommerceAuthProvisionsFromKoperasiSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeSSORepo{}
	r := gin.New()
	store := cookie.NewStore([]byte("0123456789abcdef0123456789abcdef"))
	r.Use(sessions.Sessions("test_session", store))
	r.GET("/ecommerce/profile",
		func(c *gin.Context) {
			sess := sessions.Default(c)
			sess.Set("user_id", 3)
			sess.Set("user_email", "anggota@koperasi.id")
			c.Next()
		},
		RequireECommerceAuth(service.NewECAccountService(repo)),
		func(c *gin.Context) {
			if got := sessions.Default(c).Get("ec_user_id"); got != 9 {
				t.Fatalf("ec_user_id = %v, mau 9", got)
			}
			c.String(http.StatusOK, "ok")
		},
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ecommerce/profile", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("protected EC route = %d, mau %d", rec.Code, http.StatusOK)
	}
	if !repo.called {
		t.Fatal("EnsureUserFromKoperasi harus dipanggil untuk SSO dari session koperasi")
	}
}

func requestRoleProtectedRoute(role string, guards ...gin.HandlerFunc) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	store := cookie.NewStore([]byte("0123456789abcdef0123456789abcdef"))
	r.Use(sessions.Sessions("test_session", store))

	handlers := []gin.HandlerFunc{
		func(c *gin.Context) {
			sessions.Default(c).Set("ec_role", role)
			c.Next()
		},
	}
	handlers = append(handlers, guards...)
	handlers = append(handlers, func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	r.GET("/protected", handlers...)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(rec, req)
	return rec
}
