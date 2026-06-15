package ecommerce

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/handler"
	"koperasi-frontend/core/service"
)

// WishlistHandler handles wishlist operations.
type WishlistHandler struct {
	Render ECRenderer
	Svc    *service.ECShopService
}

func NewWishlistHandler(render ECRenderer, svc *service.ECShopService) *WishlistHandler {
	return &WishlistHandler{Render: render, Svc: svc}
}

func (h *WishlistHandler) ready(c *gin.Context) bool {
	if h.Svc == nil {
		c.String(http.StatusServiceUnavailable, "E-Commerce sementara tidak tersedia (database tidak terhubung).")
		return false
	}
	return true
}

// List shows the user's wishlist.
// GET /ecommerce/wishlist
func (h *WishlistHandler) List(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)
	products, err := h.Svc.Wishlist(c.Request.Context(), ecUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/wishlist/list", gin.H{
		"Title":    "Wishlist Saya",
		"Active":   "wishlist",
		"Products": products,
	})
}

// Toggle adds or removes a product from the wishlist.
// POST /ecommerce/wishlist/toggle
func (h *WishlistHandler) Toggle(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)
	productID, err := strconv.Atoi(c.PostForm("product_id"))
	if err != nil || productID <= 0 {
		handler.SetFlash(c, "ec_error", "Produk tidak valid.")
		c.Redirect(http.StatusFound, "/ecommerce/products")
		return
	}

	added, name, err := h.Svc.ToggleWishlist(c.Request.Context(), ecUserID, productID)
	if err != nil {
		handler.SetFlash(c, "ec_error", "Gagal memperbarui wishlist.")
		c.Redirect(http.StatusFound, "/ecommerce/wishlist")
		return
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
