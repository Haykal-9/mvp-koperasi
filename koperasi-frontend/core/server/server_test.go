package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestSessionCookieSecurityOptions(t *testing.T) {
	Assets = os.DirFS("../../api")
	app := SetupApp(nil, AppOptions{
		SessionSecret: "0123456789abcdef0123456789abcdef",
		SessionSecure: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	cookie := rec.Header().Get("Set-Cookie")
	for _, want := range []string{"HttpOnly", "Secure", "SameSite=Lax"} {
		if !strings.Contains(cookie, want) {
			t.Fatalf("cookie tidak memuat %q: %s", want, cookie)
		}
	}
}

func TestCSRFMiddlewareRejectsMissingToken(t *testing.T) {
	Assets = os.DirFS("../../api")
	app := SetupApp(nil, AppOptions{
		SessionSecret: "0123456789abcdef0123456789abcdef",
	})

	form := url.Values{"email": {"owner@koperasi.id"}, "password": {"owner123"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("POST tanpa CSRF token = %d, mau %d", rec.Code, http.StatusForbidden)
	}
}

func TestCSRFMiddlewareAcceptsSessionToken(t *testing.T) {
	Assets = os.DirFS("../../api")
	app := SetupApp(nil, AppOptions{
		SessionSecret: "0123456789abcdef0123456789abcdef",
	})

	getReq := httptest.NewRequest(http.MethodGet, "/login", nil)
	getRec := httptest.NewRecorder()
	app.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("GET /login = %d, mau %d", getRec.Code, http.StatusOK)
	}

	matches := regexp.MustCompile(`const token = "([^"]+)"`).FindStringSubmatch(getRec.Body.String())
	if len(matches) != 2 {
		t.Fatalf("CSRF token tidak ditemukan di halaman login")
	}

	form := url.Values{
		"email":    {"owner@koperasi.id"},
		"password": {"owner123"},
		"_csrf":    {matches[1]},
	}
	postReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range getRec.Result().Cookies() {
		postReq.AddCookie(cookie)
	}
	postRec := httptest.NewRecorder()
	app.ServeHTTP(postRec, postReq)

	if postRec.Code == http.StatusForbidden {
		t.Fatalf("POST dengan CSRF token valid tetap ditolak: %s", postRec.Body.String())
	}
}
