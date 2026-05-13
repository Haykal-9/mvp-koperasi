package ecommerce

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/handler"
	"koperasi-frontend/core/mock"
	"koperasi-frontend/core/model"
)

// OrderHandler handles checkout, order listing, and tracking.
type OrderHandler struct {
	Render ECRenderer
}

func NewOrderHandler(render ECRenderer) *OrderHandler {
	return &OrderHandler{Render: render}
}

// ShowCheckout renders the checkout page with product selection, address, shipping, voucher, and payment.
// GET /ecommerce/checkout?product_id=X&qty=N
func (h *OrderHandler) ShowCheckout(c *gin.Context) {
	ecUserID := GetECUserID(c)

	productIDStr := c.Query("product_id")
	qtyStr := c.DefaultQuery("qty", "1")
	productID, _ := strconv.Atoi(productIDStr)
	qty, _ := strconv.Atoi(qtyStr)
	if qty < 1 {
		qty = 1
	}

	product := mock.FindECProductByID(productID)
	if product == nil {
		handler.SetFlash(c, "ec_error", "Produk tidak ditemukan.")
		c.Redirect(http.StatusFound, "/ecommerce/products")
		return
	}

	items := []model.ECOrderItem{
		{
			ProductID:   product.ID,
			ProductNama: product.Nama,
			SellerID:    product.SellerID,
			Jumlah:      qty,
			HargaSatuan: product.Harga,
			Subtotal:    product.Harga * float64(qty),
		},
	}
	subtotal := product.Harga * float64(qty)
	totalWeight := product.Berat * qty

	addresses := mock.GetECUserAddresses(ecUserID)
	shippingOpts := mock.GetECShippingOptions()
	points := mock.GetECUserPoints(ecUserID)
	pointsBalance := 0.0
	if points != nil {
		pointsBalance = points.Balance
	}

	// Check koperasi link
	user := mock.FindECommerceUserByID(ecUserID)
	isLinked := user != nil && user.LinkedKoperasiMemberID > 0

	flashErr, _ := handler.PopFlash(c, "ec_error")

	h.Render(c, "ec_base", "ecommerce/order/checkout", gin.H{
		"Title":         "Checkout",
		"Active":        "orders",
		"Items":         items,
		"Subtotal":      subtotal,
		"TotalWeight":   totalWeight,
		"Addresses":     addresses,
		"ShippingOpts":  shippingOpts,
		"PointsBalance": pointsBalance,
		"Product":       product,
		"Qty":           qty,
		"IsLinked":      isLinked,
		"FlashError":    flashErr,
	})
}

// DoCheckout processes the checkout form and creates an order.
// POST /ecommerce/checkout
func (h *OrderHandler) DoCheckout(c *gin.Context) {
	ecUserID := GetECUserID(c)

	productIDStr := c.PostForm("product_id")
	qtyStr := c.PostForm("qty")
	alamat := strings.TrimSpace(c.PostForm("alamat"))
	shippingOptStr := c.PostForm("shipping_option")
	voucherCode := strings.TrimSpace(c.PostForm("voucher_code"))
	metodeBayar := c.PostForm("metode_bayar")
	pointsUsedStr := c.DefaultPostForm("points_used", "0")

	productID, _ := strconv.Atoi(productIDStr)
	qty, _ := strconv.Atoi(qtyStr)
	shippingOptID, _ := strconv.Atoi(shippingOptStr)
	pointsUsed, _ := strconv.ParseFloat(pointsUsedStr, 64)

	if qty < 1 {
		qty = 1
	}

	product := mock.FindECProductByID(productID)
	if product == nil {
		handler.SetFlash(c, "ec_error", "Produk tidak ditemukan.")
		c.Redirect(http.StatusFound, "/ecommerce/products")
		return
	}

	if alamat == "" {
		handler.SetFlash(c, "ec_error", "Alamat pengiriman wajib diisi.")
		c.Redirect(http.StatusFound, fmt.Sprintf("/ecommerce/checkout?product_id=%d&qty=%d", productID, qty))
		return
	}

	if metodeBayar == "" {
		metodeBayar = "Transfer Bank"
	}

	items := []model.ECOrderItem{
		{
			ProductID:   product.ID,
			ProductNama: product.Nama,
			SellerID:    product.SellerID,
			Jumlah:      qty,
			HargaSatuan: product.Harga,
			Subtotal:    product.Harga * float64(qty),
		},
	}

	// Calculate shipping
	totalWeight := product.Berat * qty
	shippingCost := mock.CalculateECShippingCost(shippingOptID, totalWeight)

	// Shipping option name
	shippingName := ""
	for _, s := range mock.ECShippingOptions {
		if s.ID == shippingOptID {
			shippingName = s.Nama
			break
		}
	}

	// Validate points
	if pointsUsed > 0 {
		up := mock.GetECUserPoints(ecUserID)
		if up == nil || up.Balance < pointsUsed {
			pointsUsed = 0
		}
	}

	// Create order
	order := mock.CreateECOrder(ecUserID, items, alamat, shippingName, voucherCode, metodeBayar, shippingCost, pointsUsed)
	if order == nil {
		handler.SetFlash(c, "ec_error", "Gagal membuat pesanan.")
		c.Redirect(http.StatusFound, fmt.Sprintf("/ecommerce/checkout?product_id=%d&qty=%d", productID, qty))
		return
	}

	// Deduct stock
	product.Stok -= qty
	if product.Stok < 0 {
		product.Stok = 0
	}
	product.TotalSold += qty

	// Flash message
	msg := fmt.Sprintf("Pesanan %s berhasil dibuat! Total: Rp %.0f.", order.NomorOrder, order.TotalHarga)
	if order.PointsEarned > 0 {
		msg += fmt.Sprintf(" Anda mendapat %.0f poin dari pembelian ini!", order.PointsEarned)
	}
	handler.SetFlash(c, "ec_success", msg)

	c.Redirect(http.StatusFound, fmt.Sprintf("/ecommerce/orders/%d/track", order.ID))
}

// OrderList shows the buyer's order history.
// GET /ecommerce/orders
func (h *OrderHandler) OrderList(c *gin.Context) {
	ecUserID := GetECUserID(c)
	orders := mock.GetECUserOrders(ecUserID)

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

	h.Render(c, "ec_base", "ecommerce/order/list", gin.H{
		"Title":        "Pesanan Saya",
		"Active":       "orders",
		"Orders":       orders,
		"StatusFilter": statusFilter,
	})
}

// ShowTracking shows order details and tracking timeline.
// GET /ecommerce/orders/:id/track
func (h *OrderHandler) ShowTracking(c *gin.Context) {
	var orderID int
	fmt.Sscanf(c.Param("id"), "%d", &orderID)

	order := mock.FindECOrderByID(orderID)
	if order == nil {
		handler.SetFlash(c, "ec_error", "Pesanan tidak ditemukan.")
		c.Redirect(http.StatusFound, "/ecommerce/orders")
		return
	}

	events := mock.GetECShipmentEvents(orderID)

	h.Render(c, "ec_base", "ecommerce/order/tracking", gin.H{
		"Title":   "Tracking: " + order.NomorOrder,
		"Active":  "orders",
		"Order":   order,
		"Events":  events,
	})
}
