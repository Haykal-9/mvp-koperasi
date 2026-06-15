package ecommerce

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/handler"
	"koperasi-frontend/core/service"
)

// ReviewHandler handles product reviews after order completion.
type ReviewHandler struct {
	Render ECRenderer
	Svc    *service.ECShopService
}

func NewReviewHandler(render ECRenderer, svc *service.ECShopService) *ReviewHandler {
	return &ReviewHandler{Render: render, Svc: svc}
}

func (h *ReviewHandler) ready(c *gin.Context) bool {
	if h.Svc == nil {
		c.String(http.StatusServiceUnavailable, "E-Commerce sementara tidak tersedia (database tidak terhubung).")
		return false
	}
	return true
}

// ShowReview renders the review form for a completed order.
// GET /ecommerce/order/:id/review
func (h *ReviewHandler) ShowReview(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	orderID, _ := strconv.Atoi(c.Param("id"))
	ecUserID := GetECUserID(c)

	order, err := h.Svc.OrderByID(c.Request.Context(), orderID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if order == nil || order.BuyerID != ecUserID {
		handler.SetFlash(c, "ec_error", "Pesanan tidak ditemukan.")
		c.Redirect(http.StatusFound, "/ecommerce/orders")
		return
	}
	if order.Status != "SELESAI" {
		handler.SetFlash(c, "ec_error", "Hanya pesanan SELESAI yang bisa diulas.")
		c.Redirect(http.StatusFound, fmt.Sprintf("/ecommerce/orders/%d/track", orderID))
		return
	}

	// Check which products in this order already have a review from this user
	type ItemReviewStatus struct {
		ProductID   int
		ProductNama string
		HasReview   bool
	}
	var items []ItemReviewStatus
	for _, item := range order.Items {
		hasReview, _ := h.Svc.HasReviewed(c.Request.Context(), ecUserID, item.ProductID)
		items = append(items, ItemReviewStatus{
			ProductID:   item.ProductID,
			ProductNama: item.ProductNama,
			HasReview:   hasReview,
		})
	}

	h.Render(c, "ec_base", "ecommerce/order/review", gin.H{
		"Title":  "Beri Ulasan",
		"Active": "orders",
		"Order":  order,
		"Items":  items,
	})
}

// DoReview processes the submitted review form.
// POST /ecommerce/order/:id/review
func (h *ReviewHandler) DoReview(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	orderID, _ := strconv.Atoi(c.Param("id"))
	ecUserID := GetECUserID(c)
	username := GetECUsername(c)

	order, err := h.Svc.OrderByID(c.Request.Context(), orderID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if order == nil || order.BuyerID != ecUserID || order.Status != "SELESAI" {
		handler.SetFlash(c, "ec_error", "Tidak dapat memberikan ulasan.")
		c.Redirect(http.StatusFound, "/ecommerce/orders")
		return
	}

	submitted := 0
	for _, item := range order.Items {
		ratingStr := c.PostForm(fmt.Sprintf("rating_%d", item.ProductID))
		komentar := c.PostForm(fmt.Sprintf("komentar_%d", item.ProductID))

		if ratingStr == "" {
			continue // user skipped this product
		}
		rating, err := strconv.Atoi(ratingStr)
		if err != nil || rating < 1 || rating > 5 {
			continue
		}
		if has, _ := h.Svc.HasReviewed(c.Request.Context(), ecUserID, item.ProductID); has {
			continue // already reviewed
		}
		if err := h.Svc.SubmitReview(c.Request.Context(), ecUserID, item.ProductID, rating, komentar, username); err != nil {
			continue
		}
		submitted++
	}

	if submitted == 0 {
		handler.SetFlash(c, "ec_error", "Tidak ada ulasan yang dikirim atau semua produk sudah diulas.")
		c.Redirect(http.StatusFound, fmt.Sprintf("/ecommerce/order/%d/review", orderID))
		return
	}

	handler.SetFlash(c, "ec_success", fmt.Sprintf("%d ulasan berhasil dikirim. Terima kasih!", submitted))
	c.Redirect(http.StatusFound, "/ecommerce/orders")
}
