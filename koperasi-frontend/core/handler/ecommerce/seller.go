package ecommerce

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/handler"
	"koperasi-frontend/core/mock"
	"koperasi-frontend/core/model"
)

// SellerHandler handles seller dashboard and product management.
type SellerHandler struct {
	Render ECRenderer
}

func NewSellerHandler(render ECRenderer) *SellerHandler {
	return &SellerHandler{Render: render}
}

// Dashboard shows seller metrics and recent activity.
// GET /ecommerce/seller
func (h *SellerHandler) Dashboard(c *gin.Context) {
	sellerID := GetECUserID(c)
	profile := mock.GetSellerProfile(sellerID)
	products := mock.GetSellerProducts(sellerID)
	orders := mock.GetECSellerReceivedOrders(sellerID)

	// Metrics
	totalRevenue := 0.0
	pendingOrders := 0
	completedOrders := 0
	for _, o := range orders {
		if o.Status == "SELESAI" {
			totalRevenue += o.Subtotal
			completedOrders++
		}
		if o.Status == "DIBAYAR" || o.Status == "DIPROSES" {
			pendingOrders++
		}
	}

	// Low stock alerts
	var lowStock []model.ECProduct
	for _, p := range products {
		if p.Stok <= 5 && p.Status == "APPROVED" {
			lowStock = append(lowStock, p)
		}
	}

	// Recent orders (max 5)
	recentOrders := orders
	if len(recentOrders) > 5 {
		recentOrders = recentOrders[:5]
	}

	// Pending approval
	pendingProducts := 0
	for _, p := range products {
		if p.Status == "PENDING_APPROVAL" {
			pendingProducts++
		}
	}

	h.Render(c, "ec_base", "ecommerce/seller/dashboard", gin.H{
		"Title":           "Seller Dashboard",
		"Active":          "seller",
		"Profile":         profile,
		"TotalProducts":   len(products),
		"TotalRevenue":    totalRevenue,
		"PendingOrders":   pendingOrders,
		"CompletedOrders": completedOrders,
		"LowStock":        lowStock,
		"RecentOrders":    recentOrders,
		"PendingProducts": pendingProducts,
	})
}

// ProductList shows all products owned by this seller.
// GET /ecommerce/seller/products
func (h *SellerHandler) ProductList(c *gin.Context) {
	sellerID := GetECUserID(c)
	products := mock.GetSellerProducts(sellerID)

	statusFilter := c.Query("status")
	if statusFilter != "" {
		var filtered []model.ECProduct
		for _, p := range products {
			if p.Status == statusFilter {
				filtered = append(filtered, p)
			}
		}
		products = filtered
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
	categories := mock.GetECCategories()

	h.Render(c, "ec_base", "ecommerce/seller/product_form", gin.H{
		"Title":      "Tambah Produk Baru",
		"Active":     "seller_products",
		"Categories": categories,
	})
}

// DoCreateProduct processes the product creation form.
// POST /ecommerce/seller/products/create
func (h *SellerHandler) DoCreateProduct(c *gin.Context) {
	sellerID := GetECUserID(c)
	seller := mock.GetSellerProfile(sellerID)
	sellerName := "Seller"
	if seller != nil {
		sellerName = seller.StoreName
	}

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

	newProduct := model.ECProduct{
		ID:         mock.NextECProductID(),
		SellerID:   sellerID,
		SellerName: sellerName,
		Nama:       nama,
		Deskripsi:  deskripsi,
		Kategori:   kategori,
		Harga:      harga,
		Stok:       stok,
		Berat:      berat,
		FotoURL:    fmt.Sprintf("https://placehold.co/400x400/2d1b69/e2e8f0?text=%s", strings.ReplaceAll(nama, " ", "+")),
		Status:     "PENDING_APPROVAL",
		CreatedAt:  time.Now().Format("2006-01-02"),
	}

	mock.ECProducts = append(mock.ECProducts, newProduct)
	handler.SetFlash(c, "ec_success", fmt.Sprintf("Produk \"%s\" berhasil ditambahkan. Menunggu approval admin.", nama))
	c.Redirect(http.StatusFound, "/ecommerce/seller/products")
}

// OrderList shows orders received by this seller.
// GET /ecommerce/seller/orders
func (h *SellerHandler) OrderList(c *gin.Context) {
	sellerID := GetECUserID(c)
	orders := mock.GetECSellerReceivedOrders(sellerID)

	statusFilter := c.Query("status")
	if statusFilter != "" {
		var filtered []model.ECOrder
		for _, o := range orders {
			if o.Status == statusFilter {
				filtered = append(filtered, o)
			}
		}
		orders = filtered
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
	var orderID int
	fmt.Sscanf(c.Param("id"), "%d", &orderID)

	resi := strings.TrimSpace(c.PostForm("resi"))
	order := mock.FindECOrderByID(orderID)
	if order == nil {
		handler.SetFlash(c, "ec_error", "Pesanan tidak ditemukan.")
		c.Redirect(http.StatusFound, "/ecommerce/seller/orders")
		return
	}

	if order.Status != "DIBAYAR" && order.Status != "DIPROSES" {
		handler.SetFlash(c, "ec_error", "Pesanan tidak bisa dikirim (status: "+order.Status+").")
		c.Redirect(http.StatusFound, "/ecommerce/seller/orders")
		return
	}

	order.Status = "DIKIRIM"
	order.ResiPengiriman = resi
	order.UpdatedAt = time.Now().Format("2006-01-02 15:04")

	sellerName := "Seller"
	sp := mock.GetSellerProfile(GetECUserID(c))
	if sp != nil {
		sellerName = sp.StoreName
	}

	mock.AddECShipmentEvent(orderID, "DIKIRIM", "Gudang "+sellerName, "Paket diserahkan ke kurir. Resi: "+resi)

	handler.SetFlash(c, "ec_success", fmt.Sprintf("Pesanan %s berhasil dikirim. Resi: %s", order.NomorOrder, resi))
	c.Redirect(http.StatusFound, "/ecommerce/seller/orders")
}

// Earnings shows seller revenue and commission breakdown.
// GET /ecommerce/seller/earnings
func (h *SellerHandler) Earnings(c *gin.Context) {
	sellerID := GetECUserID(c)
	orders := mock.GetECSellerReceivedOrders(sellerID)

	totalRevenue := 0.0
	totalCommission := 0.0
	completedCount := 0
	for _, o := range orders {
		if o.Status == "SELESAI" {
			totalRevenue += o.Subtotal
			totalCommission += o.Subtotal * 0.03 // 3% commission
			completedCount++
		}
	}
	netEarnings := totalRevenue - totalCommission

	h.Render(c, "ec_base", "ecommerce/seller/earnings", gin.H{
		"Title":           "Pendapatan",
		"Active":          "seller",
		"TotalRevenue":    totalRevenue,
		"TotalCommission": totalCommission,
		"NetEarnings":     netEarnings,
		"CompletedOrders": completedCount,
		"Orders":          orders,
	})
}
