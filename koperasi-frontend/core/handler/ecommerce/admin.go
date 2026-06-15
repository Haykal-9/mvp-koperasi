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

// AdminHandler handles admin dashboard and moderation.
type AdminHandler struct {
	Render ECRenderer
	Svc    *service.ECAdminService
}

func NewAdminHandler(render ECRenderer, svc *service.ECAdminService) *AdminHandler {
	return &AdminHandler{Render: render, Svc: svc}
}

func (h *AdminHandler) ready(c *gin.Context) bool {
	if h.Svc == nil {
		c.String(http.StatusServiceUnavailable, "E-Commerce sementara tidak tersedia (database tidak terhubung).")
		return false
	}
	return true
}

// Dashboard shows admin KPI metrics and pending items.
// GET /ecommerce/admin
func (h *AdminHandler) Dashboard(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	d, err := h.Svc.Dashboard(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/admin/dashboard", gin.H{
		"Title":           "Admin Dashboard",
		"Active":          "admin",
		"PendingProducts": d.PendingProducts,
		"TotalOrders":     d.TotalOrders,
		"CompletedOrders": d.CompletedOrders,
		"TotalRevenue":    d.TotalRevenue,
		"TotalUsers":      d.TotalUsers,
		"ActiveSellers":   d.ActiveSellers,
		"ActiveVouchers":  d.ActiveVouchers,
		"RecentActivity":  d.RecentActivity,
	})
}

// Analytics shows revenue charts and seller/product performance.
// GET /ecommerce/admin/analytics
func (h *AdminHandler) Analytics(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	a, err := h.Svc.Analytics(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/admin/analytics", gin.H{
		"Title":        "Analytics",
		"Active":       "admin_analytics",
		"SellerStats":  a.SellerStats,
		"TopProducts":  a.TopProducts,
		"StatusCount":  a.StatusCount,
		"TotalRevenue": a.TotalRevenue,
		"TotalOrders":  a.TotalOrders,
	})
}

// SellerApprovals lists non-seller users who could be activated as sellers.
// GET /ecommerce/admin/sellers
func (h *AdminHandler) SellerApprovals(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	nonSellers, activeSellers, err := h.Svc.SellerApprovals(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/admin/sellers", gin.H{
		"Title":         "Manajemen Seller",
		"Active":        "admin_sellers",
		"NonSellers":    nonSellers,
		"ActiveSellers": activeSellers,
	})
}

// ApproveSeller activates a user's seller account.
// POST /ecommerce/admin/sellers/:id/approve
func (h *AdminHandler) ApproveSeller(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	sellerID, _ := strconv.Atoi(c.Param("id"))

	username, err := h.Svc.ApproveSeller(c.Request.Context(), GetECUserID(c), GetECUsername(c), sellerID)
	if err != nil {
		handler.SetFlash(c, "ec_error", err.Error())
		c.Redirect(http.StatusFound, "/ecommerce/admin/sellers")
		return
	}

	handler.SetFlash(c, "ec_success", fmt.Sprintf("Seller %s berhasil diaktifkan.", username))
	c.Redirect(http.StatusFound, "/ecommerce/admin/sellers")
}

// RejectSeller logs a rejection action for a seller application.
// POST /ecommerce/admin/sellers/:id/reject
func (h *AdminHandler) RejectSeller(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	sellerID, _ := strconv.Atoi(c.Param("id"))
	alasan := strings.TrimSpace(c.PostForm("alasan"))

	username, err := h.Svc.RejectSeller(c.Request.Context(), GetECUserID(c), GetECUsername(c), sellerID, alasan)
	if err != nil {
		handler.SetFlash(c, "ec_error", err.Error())
		c.Redirect(http.StatusFound, "/ecommerce/admin/sellers")
		return
	}

	handler.SetFlash(c, "ec_success", fmt.Sprintf("Pengajuan seller %s telah ditolak.", username))
	c.Redirect(http.StatusFound, "/ecommerce/admin/sellers")
}

// ProductApprovals lists products pending admin approval.
// GET /ecommerce/admin/products
func (h *AdminHandler) ProductApprovals(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	pending, all, err := h.Svc.ProductApprovals(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/admin/products", gin.H{
		"Title":       "Review Produk",
		"Active":      "admin_products",
		"Pending":     pending,
		"AllProducts": all,
	})
}

// ApproveProduct approves a pending product.
// POST /ecommerce/admin/products/:id/approve
func (h *AdminHandler) ApproveProduct(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	productID, _ := strconv.Atoi(c.Param("id"))

	p, err := h.Svc.ApproveProduct(c.Request.Context(), GetECUserID(c), GetECUsername(c), productID)
	if err != nil {
		handler.SetFlash(c, "ec_error", err.Error())
		c.Redirect(http.StatusFound, "/ecommerce/admin/products")
		return
	}

	handler.SetFlash(c, "ec_success", fmt.Sprintf("Produk \"%s\" berhasil diapprove.", p.Nama))
	c.Redirect(http.StatusFound, "/ecommerce/admin/products")
}

// RejectProduct rejects a pending product.
// POST /ecommerce/admin/products/:id/reject
func (h *AdminHandler) RejectProduct(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	productID, _ := strconv.Atoi(c.Param("id"))
	alasan := strings.TrimSpace(c.PostForm("alasan"))

	p, err := h.Svc.RejectProduct(c.Request.Context(), GetECUserID(c), GetECUsername(c), productID, alasan)
	if err != nil {
		handler.SetFlash(c, "ec_error", err.Error())
		c.Redirect(http.StatusFound, "/ecommerce/admin/products")
		return
	}

	handler.SetFlash(c, "ec_success", fmt.Sprintf("Produk \"%s\" telah ditolak.", p.Nama))
	c.Redirect(http.StatusFound, "/ecommerce/admin/products")
}

// VoucherManagement shows all vouchers with CRUD options.
// GET /ecommerce/admin/vouchers
func (h *AdminHandler) VoucherManagement(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	vouchers, err := h.Svc.Vouchers(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/admin/vouchers", gin.H{
		"Title":    "Manajemen Voucher",
		"Active":   "admin_vouchers",
		"Vouchers": vouchers,
	})
}

// CreateVoucher processes new voucher creation.
// POST /ecommerce/admin/vouchers/create
func (h *AdminHandler) CreateVoucher(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	code := strings.ToUpper(strings.TrimSpace(c.PostForm("code")))
	deskripsi := strings.TrimSpace(c.PostForm("deskripsi"))
	tipeDiskon := c.PostForm("tipe_diskon")
	nilaiStr := c.PostForm("nilai_diskon")
	berlakuSampai := c.PostForm("berlaku_sampai")

	if code == "" || deskripsi == "" || nilaiStr == "" {
		handler.SetFlash(c, "ec_error", "Kode, deskripsi, dan nilai diskon wajib diisi.")
		c.Redirect(http.StatusFound, "/ecommerce/admin/vouchers")
		return
	}

	nilai, _ := strconv.ParseFloat(nilaiStr, 64)
	min, _ := strconv.ParseFloat(c.PostForm("min_pembelian"), 64)
	maks, _ := strconv.ParseFloat(c.PostForm("maks_diskon"), 64)
	kuota, _ := strconv.Atoi(c.PostForm("kuota"))

	if err := h.Svc.CreateVoucher(c.Request.Context(), GetECUserID(c), GetECUsername(c),
		code, deskripsi, tipeDiskon, nilai, min, maks, kuota, berlakuSampai); err != nil {
		handler.SetFlash(c, "ec_error", err.Error())
		c.Redirect(http.StatusFound, "/ecommerce/admin/vouchers")
		return
	}

	handler.SetFlash(c, "ec_success", fmt.Sprintf("Voucher %s berhasil dibuat.", code))
	c.Redirect(http.StatusFound, "/ecommerce/admin/vouchers")
}

// ToggleVoucher toggles a voucher's status between ACTIVE and EXPIRED.
// POST /ecommerce/admin/vouchers/:id/toggle
func (h *AdminHandler) ToggleVoucher(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	voucherID, _ := strconv.Atoi(c.Param("id"))

	code, newStatus, err := h.Svc.ToggleVoucher(c.Request.Context(), GetECUserID(c), GetECUsername(c), voucherID)
	if err != nil {
		handler.SetFlash(c, "ec_error", err.Error())
		c.Redirect(http.StatusFound, "/ecommerce/admin/vouchers")
		return
	}

	if newStatus == "EXPIRED" {
		handler.SetFlash(c, "ec_success", fmt.Sprintf("Voucher %s dinonaktifkan.", code))
	} else {
		handler.SetFlash(c, "ec_success", fmt.Sprintf("Voucher %s diaktifkan kembali.", code))
	}
	c.Redirect(http.StatusFound, "/ecommerce/admin/vouchers")
}

// OrderManagement shows all orders in the system.
// GET /ecommerce/admin/orders
func (h *AdminHandler) OrderManagement(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	statusFilter := c.Query("status")

	orders, counts, err := h.Svc.Orders(c.Request.Context(), statusFilter)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/admin/orders", gin.H{
		"Title":        "Manajemen Order",
		"Active":       "admin_orders",
		"Orders":       orders,
		"StatusFilter": statusFilter,
		"StatusCounts": counts,
	})
}

// MemberManagement shows all e-commerce users.
// GET /ecommerce/admin/members
func (h *AdminHandler) MemberManagement(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	roleFilter := c.Query("role")

	users, total, err := h.Svc.Members(c.Request.Context(), roleFilter)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/admin/members", gin.H{
		"Title":      "Manajemen Member",
		"Active":     "admin_members",
		"Users":      users,
		"RoleFilter": roleFilter,
		"TotalUsers": total,
	})
}

// PointsMonitoring shows all points transactions.
// GET /ecommerce/admin/points
func (h *AdminHandler) PointsMonitoring(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	tx, userPoints, totalPoints, err := h.Svc.PointsMonitoring(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/admin/points", gin.H{
		"Title":        "Monitoring Poin",
		"Active":       "admin_points",
		"Transactions": tx,
		"UserPoints":   userPoints,
		"TotalPoints":  totalPoints,
	})
}

// AuditLog shows all system audit actions.
// GET /ecommerce/admin/audit
func (h *AdminHandler) AuditLog(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	actionFilter := c.Query("action")

	logs, err := h.Svc.AuditLogs(c.Request.Context(), actionFilter)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/admin/audit", gin.H{
		"Title":        "Audit Log",
		"Active":       "admin_audit",
		"Logs":         logs,
		"ActionFilter": actionFilter,
	})
}
