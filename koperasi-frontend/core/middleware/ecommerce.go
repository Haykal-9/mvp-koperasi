package middleware

import (
	"net/http"

	"koperasi-frontend/core/handler"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// RequireECommerceAuth checks that the user has a valid e-commerce session.
// If not authenticated, redirects to /ecommerce/login.
func RequireECommerceAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		sess := sessions.Default(c)
		if sess.Get("ec_user_id") == nil {
			handler.SetFlash(c, "ec_error", "Silakan login E-Commerce terlebih dahulu.")
			c.Redirect(http.StatusFound, "/ecommerce/login")
			c.Abort()
			return
		}
		c.Next()
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
func RequireECommerceAdmin() gin.HandlerFunc {
	return RequireECommerceRole("ADMIN")
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
