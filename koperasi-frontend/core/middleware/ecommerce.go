package middleware

import (
	"net/http"

	"koperasi-frontend/core/handler"
	echandler "koperasi-frontend/core/handler/ecommerce"
	"koperasi-frontend/core/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// RequireECommerceAuth checks that the user has a valid e-commerce session.
// If the user is logged into koperasi management and has a linked e-commerce account,
// they are auto-logged in (SSO via acct). Otherwise redirects to /ecommerce/login.
// acct boleh nil (DB tidak terhubung) — SSO fallback dilewati.
func RequireECommerceAuth(acct *service.ECAccountService) gin.HandlerFunc {
	return func(c *gin.Context) {
		sess := sessions.Default(c)

		// Already logged into e-commerce → continue
		if sess.Get("ec_user_id") != nil {
			c.Next()
			return
		}

		// Auto-login fallback: restore EC session from koperasi session (all roles)
		if acct != nil && sess.Get("user_id") != nil {
			userEmail, _ := sess.Get("user_email").(string)
			if userEmail != "" {
				ecUser, _ := acct.UserByEmail(c.Request.Context(), userEmail)
				if ecUser != nil {
					if err := echandler.SetECSession(c, ecUser.ID, ecUser.Username, ecUser.Email, ecUser.Role, ecUser.IsSellerActive); err == nil {
						c.Next()
						return
					}
				}
			}
		}

		handler.SetFlash(c, "ec_error", "Silakan login E-Commerce terlebih dahulu.")
		c.Redirect(http.StatusFound, "/ecommerce/login")
		c.Abort()
	}
}

// RequireECommerceRole checks that the authenticated e-commerce user has
// one of the allowed roles (e.g., "BUYER", "SELLER", "ADMIN").
func RequireECommerceRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sess := sessions.Default(c)
		userRole, _ := sess.Get("ec_role").(string)
		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}
		c.String(http.StatusForbidden, "Akses ditolak: role %s tidak diizinkan.", userRole)
		c.Abort()
	}
}

// RequireECommerceAdmin is a shorthand for RequireECommerceRole("ADMIN").
// Dipakai untuk rute admin-only yang sensitif (voucher, member, audit log).
func RequireECommerceAdmin() gin.HandlerFunc {
	return RequireECommerceRole("ADMIN")
}

// RequireECommerceAdminArea mengizinkan ADMIN dan PENGURUS masuk ke area admin.
// PENGURUS adalah moderator operasional (hierarki di bawah ADMIN): boleh akses
// dashboard, analytics, review produk, kelola seller, order, dan monitoring poin.
func RequireECommerceAdminArea() gin.HandlerFunc {
	return RequireECommerceRole("ADMIN", "PENGURUS")
}

// RequireECommerceSeller checks that the e-commerce user is an active seller.
func RequireECommerceSeller() gin.HandlerFunc {
	return func(c *gin.Context) {
		sess := sessions.Default(c)
		isSeller, _ := sess.Get("ec_is_seller").(bool)
		if !isSeller {
			handler.SetFlash(c, "ec_error", "Anda belum mengaktifkan akun seller.")
			c.Redirect(http.StatusFound, "/ecommerce/buyer")
			c.Abort()
			return
		}
		c.Next()
	}
}
