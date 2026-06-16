package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
