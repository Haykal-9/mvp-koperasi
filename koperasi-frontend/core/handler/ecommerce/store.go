package ecommerce

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/service"
)

// StoreHandler handles storefront pages.
type StoreHandler struct {
	Render ECRenderer
	Svc    *service.ECStoreService
}

func NewStoreHandler(render ECRenderer, svc *service.ECStoreService) *StoreHandler {
	return &StoreHandler{Render: render, Svc: svc}
}

// ready memastikan service (DB) tersedia untuk halaman storefront.
func (h *StoreHandler) ready(c *gin.Context) bool {
	if h.Svc == nil {
		c.String(http.StatusServiceUnavailable, "E-Commerce sementara tidak tersedia (database tidak terhubung).")
		return false
	}
	return true
}

// Home renders the storefront homepage.
// GET /ecommerce/store
func (h *StoreHandler) Home(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ctx := c.Request.Context()
	featured, err := h.Svc.Featured(ctx, 8)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	categories, err := h.Svc.Categories(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	sellers, err := h.Svc.SellerProfiles(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/store/home", gin.H{
		"Title":            "Storefront",
		"Active":           "store",
		"FeaturedProducts": featured,
		"Categories":       categories,
		"Sellers":          sellers,
	})
}

// CategoryPage shows products filtered by category.
// GET /ecommerce/category/:slug
func (h *StoreHandler) CategoryPage(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	slug := c.Param("slug")
	products, err := h.Svc.Catalog(c.Request.Context(), "", slug)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/store/category", gin.H{
		"Title":    "Kategori: " + slug,
		"Active":   "store",
		"Category": slug,
		"Products": products,
	})
}

// ProductCatalog shows all products with filters.
// GET /ecommerce/products
func (h *StoreHandler) ProductCatalog(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ctx := c.Request.Context()
	category := c.Query("category")
	query := c.Query("q")

	products, err := h.Svc.Catalog(ctx, query, category)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	categories, err := h.Svc.Categories(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/product/catalog", gin.H{
		"Title":      "Katalog Produk",
		"Active":     "products",
		"Products":   products,
		"Categories": categories,
		"Query":      query,
		"Category":   category,
	})
}

// ProductDetail shows a single product with reviews and seller info.
// GET /ecommerce/products/:id
func (h *StoreHandler) ProductDetail(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ctx := c.Request.Context()
	var id int
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		c.Redirect(http.StatusFound, "/ecommerce/products")
		return
	}

	product, err := h.Svc.ProductByID(ctx, id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if product == nil {
		c.Redirect(http.StatusFound, "/ecommerce/products")
		return
	}

	seller, err := h.Svc.SellerProfile(ctx, product.SellerID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	reviews, err := h.Svc.Reviews(ctx, id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	related, err := h.Svc.Related(ctx, product.SellerID, id, 4)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	wishlisted := false
	if ecUserID := GetECUserID(c); ecUserID > 0 {
		if w, werr := h.Svc.IsWishlisted(ctx, ecUserID, id); werr == nil {
			wishlisted = w
		}
	}

	h.Render(c, "ec_base", "ecommerce/product/detail", gin.H{
		"Title":           product.Nama,
		"Active":          "products",
		"Product":         product,
		"SellerProfile":   seller,
		"Reviews":         reviews,
		"RelatedProducts": related,
		"IsWishlisted":    wishlisted,
	})
}

// SellerProfilePage shows a seller's public profile.
// GET /ecommerce/seller/:id/profile
func (h *StoreHandler) SellerProfilePage(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ctx := c.Request.Context()
	var id int
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		c.Redirect(http.StatusFound, "/ecommerce/store")
		return
	}

	seller, err := h.Svc.SellerProfile(ctx, id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if seller == nil {
		c.Redirect(http.StatusFound, "/ecommerce/store")
		return
	}

	approved, err := h.Svc.SellerApprovedProducts(ctx, id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/seller/profile", gin.H{
		"Title":         seller.StoreName,
		"Active":        "store",
		"SellerProfile": seller,
		"Products":      approved,
	})
}
