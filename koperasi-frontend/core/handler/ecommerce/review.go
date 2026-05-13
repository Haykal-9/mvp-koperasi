package ecommerce

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/handler"
	"koperasi-frontend/core/mock"
)

// ReviewHandler handles product reviews after order completion.
type ReviewHandler struct {
	Render ECRenderer
}

func NewReviewHandler(render ECRenderer) *ReviewHandler {
	return &ReviewHandler{Render: render}
}

// ShowReview renders the review form for a completed order.
// GET /ecommerce/order/:id/review
func (h *ReviewHandler) ShowReview(c *gin.Context) {
	var orderID int
	fmt.Sscanf(c.Param("id"), "%d", &orderID)
	ecUserID := GetECUserID(c)

	order := mock.FindECOrderByID(orderID)
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
		hasReview := mock.HasECReviewed(ecUserID, item.ProductID)
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
	var orderID int
	fmt.Sscanf(c.Param("id"), "%d", &orderID)
	ecUserID := GetECUserID(c)

	order := mock.FindECOrderByID(orderID)
	if order == nil || order.BuyerID != ecUserID || order.Status != "SELESAI" {
		handler.SetFlash(c, "ec_error", "Tidak dapat memberikan ulasan.")
		c.Redirect(http.StatusFound, "/ecommerce/orders")
		return
	}

	submitted := 0
	for _, item := range order.Items {
		ratingKey := fmt.Sprintf("rating_%d", item.ProductID)
		komentarKey := fmt.Sprintf("komentar_%d", item.ProductID)

		ratingStr := c.PostForm(ratingKey)
		komentar := c.PostForm(komentarKey)

		if ratingStr == "" {
			continue // user skipped this product
		}

		rating, err := strconv.Atoi(ratingStr)
		if err != nil || rating < 1 || rating > 5 {
			continue
		}

		if mock.HasECReviewed(ecUserID, item.ProductID) {
			continue // already reviewed
		}

		mock.SubmitECReview(ecUserID, item.ProductID, rating, komentar)
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
