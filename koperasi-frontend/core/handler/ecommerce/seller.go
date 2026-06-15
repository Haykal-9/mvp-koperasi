package ecommerce

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/handler"
	"koperasi-frontend/core/service"
)

// SellerHandler handles seller dashboard and product management.
type SellerHandler struct {
	Render ECRenderer
	Svc    *service.ECSellerService
}

func NewSellerHandler(render ECRenderer, svc *service.ECSellerService) *SellerHandler {
	return &SellerHandler{Render: render, Svc: svc}
}

func (h *SellerHandler) ready(c *gin.Context) bool {
	if h.Svc == nil {
		c.String(http.StatusServiceUnavailable, "E-Commerce sementara tidak tersedia (database tidak terhubung).")
		return false
	}
	return true
}

// Dashboard shows seller metrics and recent activity.
// GET /ecommerce/seller
func (h *SellerHandler) Dashboard(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	sellerID := GetECUserID(c)
	d, err := h.Svc.Dashboard(c.Request.Context(), sellerID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/seller/dashboard", gin.H{
		"Title":           "Seller Dashboard",
		"Active":          "seller",
		"Profile":         d.Profile,
		"TotalProducts":   d.TotalProducts,
		"TotalRevenue":    d.TotalRevenue,
		"PendingOrders":   d.PendingOrders,
		"CompletedOrders": d.CompletedOrders,
		"LowStock":        d.LowStock,
		"RecentOrders":    d.RecentOrders,
		"PendingProducts": d.PendingProducts,
	})
}

// ProductList shows all products owned by this seller.
// GET /ecommerce/seller/products
func (h *SellerHandler) ProductList(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	sellerID := GetECUserID(c)
	statusFilter := c.Query("status")

	products, err := h.Svc.Products(c.Request.Context(), sellerID, statusFilter)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/seller/products", gin.H{
		"Title":        "Produk Saya",
		"Active":       "seller_products",
		"Products":     products,
		"StatusFilter": statusFilter,
	})
}

// CreateProduct shows the product creation form.
// GET /ecommerce/seller/products/create
func (h *SellerHandler) CreateProduct(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	categories, err := h.Svc.Categories(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/seller/product_form", gin.H{
		"Title":      "Tambah Produk Baru",
		"Active":     "seller_products",
		"Categories": categories,
	})
}

// DoCreateProduct processes the product creation form.
// POST /ecommerce/seller/products/create
func (h *SellerHandler) DoCreateProduct(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	sellerID := GetECUserID(c)

	nama := strings.TrimSpace(c.PostForm("nama"))
	deskripsi := strings.TrimSpace(c.PostForm("deskripsi"))
	kategori := strings.TrimSpace(c.PostForm("kategori"))
	hargaStr := c.PostForm("harga")
	stokStr := c.PostForm("stok")
	beratStr := c.PostForm("berat")

	if nama == "" || kategori == "" || hargaStr == "" {
		handler.SetFlash(c, "ec_error", "Nama, kategori, dan harga wajib diisi.")
		c.Redirect(http.StatusFound, "/ecommerce/seller/products/create")
		return
	}

	harga, _ := strconv.ParseFloat(hargaStr, 64)
	stok, _ := strconv.Atoi(stokStr)
	berat, _ := strconv.Atoi(beratStr)

	if harga <= 0 {
		handler.SetFlash(c, "ec_error", "Harga harus lebih dari 0.")
		c.Redirect(http.StatusFound, "/ecommerce/seller/products/create")
		return
	}

	if err := h.Svc.CreateProduct(c.Request.Context(), sellerID, nama, deskripsi, kategori, harga, stok, berat); err != nil {
		handler.SetFlash(c, "ec_error", "Gagal menambahkan produk.")
		c.Redirect(http.StatusFound, "/ecommerce/seller/products/create")
		return
	}

	handler.SetFlash(c, "ec_success", fmt.Sprintf("Produk \"%s\" berhasil ditambahkan. Menunggu approval admin.", nama))
	c.Redirect(http.StatusFound, "/ecommerce/seller/products")
}

// OrderList shows orders received by this seller.
// GET /ecommerce/seller/orders
func (h *SellerHandler) OrderList(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	sellerID := GetECUserID(c)
	statusFilter := c.Query("status")

	orders, err := h.Svc.Orders(c.Request.Context(), sellerID, statusFilter)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/seller/orders", gin.H{
		"Title":        "Pesanan Masuk",
		"Active":       "seller_orders",
		"Orders":       orders,
		"StatusFilter": statusFilter,
	})
}

// MarkShipped updates an order to DIKIRIM status.
// POST /ecommerce/seller/orders/:id/shipped
func (h *SellerHandler) MarkShipped(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	sellerID := GetECUserID(c)
	orderID, _ := strconv.Atoi(c.Param("id"))
	resi := strings.TrimSpace(c.PostForm("resi"))

	order, err := h.Svc.MarkShipped(c.Request.Context(), sellerID, orderID, resi)
	if err != nil {
		handler.SetFlash(c, "ec_error", err.Error())
		c.Redirect(http.StatusFound, "/ecommerce/seller/orders")
		return
	}

	handler.SetFlash(c, "ec_success", fmt.Sprintf("Pesanan %s berhasil dikirim. Resi: %s", order.NomorOrder, resi))
	c.Redirect(http.StatusFound, "/ecommerce/seller/orders")
}

// Earnings shows seller revenue and commission breakdown.
// GET /ecommerce/seller/earnings
func (h *SellerHandler) Earnings(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	sellerID := GetECUserID(c)
	e, err := h.Svc.Earnings(c.Request.Context(), sellerID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/seller/earnings", gin.H{
		"Title":           "Pendapatan",
		"Active":          "seller",
		"TotalRevenue":    e.TotalRevenue,
		"TotalCommission": e.TotalCommission,
		"NetEarnings":     e.NetEarnings,
		"CompletedOrders": e.CompletedOrders,
		"Orders":          e.Orders,
	})
}
