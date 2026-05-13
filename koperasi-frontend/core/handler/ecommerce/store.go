package ecommerce

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/mock"
	"koperasi-frontend/core/model"
)

// StoreHandler handles storefront pages.
type StoreHandler struct {
	Render ECRenderer
}

func NewStoreHandler(render ECRenderer) *StoreHandler {
	return &StoreHandler{Render: render}
}

// Home renders the storefront homepage.
// GET /ecommerce/store
func (h *StoreHandler) Home(c *gin.Context) {
	featured := mock.GetAllApprovedECProducts()
	categories := mock.GetECCategories()

	if len(featured) > 8 {
		featured = featured[:8]
	}

	h.Render(c, "ec_base", "ecommerce/store/home", gin.H{
		"Title":            "Storefront",
		"Active":           "store",
		"FeaturedProducts": featured,
		"Categories":       categories,
		"Sellers":          mock.ECSellerProfiles,
	})
}

// CategoryPage shows products filtered by category.
// GET /ecommerce/category/:slug
func (h *StoreHandler) CategoryPage(c *gin.Context) {
	slug := c.Param("slug")
	products := mock.GetECProductsByCategory(slug)

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
	category := c.Query("category")
	query := c.Query("q")

	var products []model.ECProduct
	if query != "" {
		products = mock.SearchECProducts(query)
	} else if category != "" {
		products = mock.GetECProductsByCategory(category)
	} else {
		products = mock.GetAllApprovedECProducts()
	}

	categories := mock.GetECCategories()

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
	var id int
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		c.Redirect(http.StatusFound, "/ecommerce/products")
		return
	}

	product := mock.FindECProductByID(id)
	if product == nil {
		c.Redirect(http.StatusFound, "/ecommerce/products")
		return
	}

	seller := mock.GetSellerProfile(product.SellerID)
	reviews := mock.GetECProductReviews(id)

	related := mock.GetSellerProducts(product.SellerID)
	var relatedFiltered []model.ECProduct
	for _, p := range related {
		if p.ID != id && p.Status == "APPROVED" {
			relatedFiltered = append(relatedFiltered, p)
		}
	}
	if len(relatedFiltered) > 4 {
		relatedFiltered = relatedFiltered[:4]
	}

	wishlisted := false
	ecUserID := GetECUserID(c)
	if ecUserID > 0 {
		wishlisted = mock.IsECWishlisted(ecUserID, id)
	}

	h.Render(c, "ec_base", "ecommerce/product/detail", gin.H{
		"Title":           product.Nama,
		"Active":          "products",
		"Product":         product,
		"SellerProfile":   seller,
		"Reviews":         reviews,
		"RelatedProducts": relatedFiltered,
		"IsWishlisted":    wishlisted,
	})
}

// SellerProfilePage shows a seller's public profile.
// GET /ecommerce/seller/:id/profile
func (h *StoreHandler) SellerProfilePage(c *gin.Context) {
	var id int
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		c.Redirect(http.StatusFound, "/ecommerce/store")
		return
	}

	seller := mock.GetSellerProfile(id)
	if seller == nil {
		c.Redirect(http.StatusFound, "/ecommerce/store")
		return
	}

	products := mock.GetSellerProducts(id)
	var approved []model.ECProduct
	for _, p := range products {
		if p.Status == "APPROVED" {
			approved = append(approved, p)
		}
	}

	h.Render(c, "ec_base", "ecommerce/seller/profile", gin.H{
		"Title":         seller.StoreName,
		"Active":        "store",
		"SellerProfile": seller,
		"Products":      approved,
	})
}
