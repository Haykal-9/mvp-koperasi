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

// OrderHandler handles checkout, order listing, and tracking.
type OrderHandler struct {
	Render ECRenderer
	Svc    *service.ECShopService
}

func NewOrderHandler(render ECRenderer, svc *service.ECShopService) *OrderHandler {
	return &OrderHandler{Render: render, Svc: svc}
}

func (h *OrderHandler) ready(c *gin.Context) bool {
	if h.Svc == nil {
		c.String(http.StatusServiceUnavailable, "E-Commerce sementara tidak tersedia (database tidak terhubung).")
		return false
	}
	return true
}

// ShowCheckout renders the checkout page with product selection, address, shipping, voucher, and payment.
// GET /ecommerce/checkout?product_id=X&qty=N
func (h *OrderHandler) ShowCheckout(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)

	productID, _ := strconv.Atoi(c.Query("product_id"))
	qty, _ := strconv.Atoi(c.DefaultQuery("qty", "1"))

	v, err := h.Svc.Checkout(c.Request.Context(), ecUserID, productID, qty)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if v == nil {
		handler.SetFlash(c, "ec_error", "Produk tidak ditemukan.")
		c.Redirect(http.StatusFound, "/ecommerce/products")
		return
	}

	flashErr, _ := handler.PopFlash(c, "ec_error")

	h.Render(c, "ec_base", "ecommerce/order/checkout", gin.H{
		"Title":         "Checkout",
		"Active":        "orders",
		"Items":         v.Items,
		"Subtotal":      v.Subtotal,
		"TotalWeight":   v.TotalWeight,
		"Addresses":     v.Addresses,
		"ShippingOpts":  v.ShippingOpts,
		"PointsBalance": v.PointsBalance,
		"Product":       v.Product,
		"Qty":           v.Qty,
		"IsLinked":      v.IsLinked,
		"FlashError":    flashErr,
	})
}

// DoCheckout processes the checkout form and creates an order.
// POST /ecommerce/checkout
func (h *OrderHandler) DoCheckout(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)

	productID, _ := strconv.Atoi(c.PostForm("product_id"))
	qty, _ := strconv.Atoi(c.PostForm("qty"))
	alamat := strings.TrimSpace(c.PostForm("alamat"))
	shippingOptID, _ := strconv.Atoi(c.PostForm("shipping_option"))
	voucherCode := strings.TrimSpace(c.PostForm("voucher_code"))
	metodeBayar := c.PostForm("metode_bayar")
	pointsUsed, _ := strconv.ParseFloat(c.DefaultPostForm("points_used", "0"), 64)

	if qty < 1 {
		qty = 1
	}

	if alamat == "" {
		handler.SetFlash(c, "ec_error", "Alamat pengiriman wajib diisi.")
		c.Redirect(http.StatusFound, fmt.Sprintf("/ecommerce/checkout?product_id=%d&qty=%d", productID, qty))
		return
	}

	order, err := h.Svc.PlaceOrder(c.Request.Context(), ecUserID, productID, qty, alamat, shippingOptID, voucherCode, metodeBayar, pointsUsed)
	if err != nil {
		handler.SetFlash(c, "ec_error", err.Error())
		c.Redirect(http.StatusFound, fmt.Sprintf("/ecommerce/checkout?product_id=%d&qty=%d", productID, qty))
		return
	}

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
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)
	statusFilter := c.Query("status")

	orders, err := h.Svc.Orders(c.Request.Context(), ecUserID, statusFilter)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
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
	if !h.ready(c) {
		return
	}
	orderID, _ := strconv.Atoi(c.Param("id"))

	order, err := h.Svc.OrderByID(c.Request.Context(), orderID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if order == nil {
		handler.SetFlash(c, "ec_error", "Pesanan tidak ditemukan.")
		c.Redirect(http.StatusFound, "/ecommerce/orders")
		return
	}

	events, err := h.Svc.ShipmentEvents(c.Request.Context(), orderID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/order/tracking", gin.H{
		"Title":  "Tracking: " + order.NomorOrder,
		"Active": "orders",
		"Order":  order,
		"Events": events,
	})
}
