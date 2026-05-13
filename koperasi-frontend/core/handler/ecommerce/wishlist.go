package ecommerce

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/handler"
	"koperasi-frontend/core/mock"
)

// WishlistHandler handles wishlist operations.
type WishlistHandler struct {
	Render ECRenderer
}

func NewWishlistHandler(render ECRenderer) *WishlistHandler {
	return &WishlistHandler{Render: render}
}

// List shows the user's wishlist.
// GET /ecommerce/wishlist
func (h *WishlistHandler) List(c *gin.Context) {
	ecUserID := GetECUserID(c)
	products := mock.GetECUserWishlist(ecUserID)

	h.Render(c, "ec_base", "ecommerce/wishlist/list", gin.H{
		"Title":    "Wishlist Saya",
		"Active":   "wishlist",
		"Products": products,
	})
}

// Toggle adds or removes a product from the wishlist.
// POST /ecommerce/wishlist/toggle
func (h *WishlistHandler) Toggle(c *gin.Context) {
	ecUserID := GetECUserID(c)
	productIDStr := c.PostForm("product_id")
	var productID int
	if _, err := fmt.Sscanf(productIDStr, "%d", &productID); err != nil || productID <= 0 {
		handler.SetFlash(c, "ec_error", "Produk tidak valid.")
		c.Redirect(http.StatusFound, "/ecommerce/products")
		return
	}

	added := mock.ToggleECWishlist(ecUserID, productID)
	p := mock.FindECProductByID(productID)
	name := "produk"
	if p != nil {
		name = p.Nama
	}

	if added {
		handler.SetFlash(c, "ec_success", name+" ditambahkan ke wishlist.")
	} else {
		handler.SetFlash(c, "ec_success", name+" dihapus dari wishlist.")
	}

	referer := c.Request.Referer()
	if referer == "" {
		referer = "/ecommerce/wishlist"
	}
	c.Redirect(http.StatusFound, referer)
}
