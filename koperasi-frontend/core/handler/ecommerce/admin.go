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

// AdminHandler handles admin dashboard and moderation.
type AdminHandler struct {
	Render ECRenderer
}

func NewAdminHandler(render ECRenderer) *AdminHandler {
	return &AdminHandler{Render: render}
}

// Dashboard shows admin KPI metrics and pending items.
// GET /ecommerce/admin
func (h *AdminHandler) Dashboard(c *gin.Context) {
	pendingProducts := mock.GetPendingECProducts()
	allOrders := mock.GetAllECOrders()
	allUsers := mock.ECommerceUsers
	allVouchers := mock.ECVouchers
	auditLogs := mock.GetECAuditLogs()

	// KPI metrics
	totalRevenue := 0.0
	totalOrders := len(allOrders)
	completedOrders := 0
	for _, o := range allOrders {
		if o.Status == "SELESAI" {
			totalRevenue += o.TotalHarga
			completedOrders++
		}
	}

	activeSellers := 0
	for _, u := range allUsers {
		if u.IsSellerActive {
			activeSellers++
		}
	}

	activeVouchers := 0
	for _, v := range allVouchers {
		if v.Status == "ACTIVE" {
			activeVouchers++
		}
	}

	// Recent audit logs (last 5)
	recentActivity := auditLogs
	if len(recentActivity) > 5 {
		recentActivity = recentActivity[:5]
	}

	h.Render(c, "ec_base", "ecommerce/admin/dashboard", gin.H{
		"Title":           "Admin Dashboard",
		"Active":          "admin",
		"PendingProducts": len(pendingProducts),
		"TotalOrders":     totalOrders,
		"CompletedOrders": completedOrders,
		"TotalRevenue":    totalRevenue,
		"TotalUsers":      len(allUsers),
		"ActiveSellers":   activeSellers,
		"ActiveVouchers":  activeVouchers,
		"RecentActivity":  recentActivity,
	})
}

// Analytics shows revenue charts and seller/product performance.
// GET /ecommerce/admin/analytics
func (h *AdminHandler) Analytics(c *gin.Context) {
	allOrders := mock.GetAllECOrders()
	allProducts := mock.ECProducts
	allSellers := mock.ECSellerProfiles

	// Revenue by seller
	sellerRevenue := map[int]float64{}
	sellerOrderCount := map[int]int{}
	for _, o := range allOrders {
		if o.Status == "SELESAI" {
			sellerRevenue[o.SellerID] += o.Subtotal
			sellerOrderCount[o.SellerID]++
		}
	}

	type SellerStat struct {
		Profile    model.SellerProfile
		Revenue    float64
		OrderCount int
	}
	var sellerStats []SellerStat
	for _, sp := range allSellers {
		sellerStats = append(sellerStats, SellerStat{
			Profile:    sp,
			Revenue:    sellerRevenue[sp.SellerID],
			OrderCount: sellerOrderCount[sp.SellerID],
		})
	}

	// Top products by total sold
	type ProductStat struct {
		Product model.ECProduct
	}
	var topProducts []model.ECProduct
	for _, p := range allProducts {
		if p.Status == "APPROVED" {
			topProducts = append(topProducts, p)
		}
	}
	// Sort by TotalSold descending (simple bubble-ish)
	for i := 0; i < len(topProducts)-1; i++ {
		for j := i + 1; j < len(topProducts); j++ {
			if topProducts[j].TotalSold > topProducts[i].TotalSold {
				topProducts[i], topProducts[j] = topProducts[j], topProducts[i]
			}
		}
	}
	if len(topProducts) > 5 {
		topProducts = topProducts[:5]
	}

	// Order status breakdown
	statusCount := map[string]int{}
	for _, o := range allOrders {
		statusCount[o.Status]++
	}

	totalRevenue := 0.0
	for _, o := range allOrders {
		if o.Status == "SELESAI" {
			totalRevenue += o.TotalHarga
		}
	}

	h.Render(c, "ec_base", "ecommerce/admin/analytics", gin.H{
		"Title":        "Analytics",
		"Active":       "admin_analytics",
		"SellerStats":  sellerStats,
		"TopProducts":  topProducts,
		"StatusCount":  statusCount,
		"TotalRevenue": totalRevenue,
		"TotalOrders":  len(allOrders),
	})
}

// SellerApprovals lists non-seller users who could be activated as sellers.
// GET /ecommerce/admin/sellers
func (h *AdminHandler) SellerApprovals(c *gin.Context) {
	var nonSellers []model.ECommerceUser
	var activeSellers []model.ECommerceUser
	for _, u := range mock.ECommerceUsers {
		if u.Role == "ADMIN" || u.Role == "KASIR" {
			continue
		}
		if u.IsSellerActive {
			activeSellers = append(activeSellers, u)
		} else {
			nonSellers = append(nonSellers, u)
		}
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
	var sellerID int
	fmt.Sscanf(c.Param("id"), "%d", &sellerID)

	adminID := GetECUserID(c)
	adminUsername := GetECUsername(c)

	u := mock.FindECommerceUserByID(sellerID)
	if u == nil {
		handler.SetFlash(c, "ec_error", "User tidak ditemukan.")
		c.Redirect(http.StatusFound, "/ecommerce/admin/sellers")
		return
	}
	if u.IsSellerActive {
		handler.SetFlash(c, "ec_error", "User sudah menjadi seller aktif.")
		c.Redirect(http.StatusFound, "/ecommerce/admin/sellers")
		return
	}

	mock.ActivateSellerAccount(sellerID)
	mock.LogAuditAction("APPROVE_SELLER", adminID, adminUsername,
		fmt.Sprintf("seller:%d", sellerID),
		fmt.Sprintf("Activated seller account for user: %s (%s)", u.Username, u.Email))

	handler.SetFlash(c, "ec_success", fmt.Sprintf("Seller %s berhasil diaktifkan.", u.Username))
	c.Redirect(http.StatusFound, "/ecommerce/admin/sellers")
}

// RejectSeller logs a rejection action for a seller application.
// POST /ecommerce/admin/sellers/:id/reject
func (h *AdminHandler) RejectSeller(c *gin.Context) {
	var sellerID int
	fmt.Sscanf(c.Param("id"), "%d", &sellerID)

	adminID := GetECUserID(c)
	adminUsername := GetECUsername(c)
	alasan := strings.TrimSpace(c.PostForm("alasan"))

	u := mock.FindECommerceUserByID(sellerID)
	if u == nil {
		handler.SetFlash(c, "ec_error", "User tidak ditemukan.")
		c.Redirect(http.StatusFound, "/ecommerce/admin/sellers")
		return
	}

	detail := fmt.Sprintf("Rejected seller application for user: %s", u.Username)
	if alasan != "" {
		detail += ". Alasan: " + alasan
	}
	mock.LogAuditAction("REJECT_SELLER", adminID, adminUsername,
		fmt.Sprintf("seller:%d", sellerID), detail)

	handler.SetFlash(c, "ec_success", fmt.Sprintf("Pengajuan seller %s telah ditolak.", u.Username))
	c.Redirect(http.StatusFound, "/ecommerce/admin/sellers")
}

// ProductApprovals lists products pending admin approval.
// GET /ecommerce/admin/products
func (h *AdminHandler) ProductApprovals(c *gin.Context) {
	pending := mock.GetPendingECProducts()
	allProducts := mock.ECProducts

	h.Render(c, "ec_base", "ecommerce/admin/products", gin.H{
		"Title":       "Review Produk",
		"Active":      "admin_products",
		"Pending":     pending,
		"AllProducts": allProducts,
	})
}

// ApproveProduct approves a pending product.
// POST /ecommerce/admin/products/:id/approve
func (h *AdminHandler) ApproveProduct(c *gin.Context) {
	var productID int
	fmt.Sscanf(c.Param("id"), "%d", &productID)

	adminID := GetECUserID(c)
	adminUsername := GetECUsername(c)

	p := mock.FindECProductByID(productID)
	if p == nil {
		handler.SetFlash(c, "ec_error", "Produk tidak ditemukan.")
		c.Redirect(http.StatusFound, "/ecommerce/admin/products")
		return
	}
	if p.Status != "PENDING_APPROVAL" {
		handler.SetFlash(c, "ec_error", "Produk tidak dalam status PENDING.")
		c.Redirect(http.StatusFound, "/ecommerce/admin/products")
		return
	}

	p.Status = "APPROVED"
	mock.LogAuditAction("APPROVE_PRODUCT", adminID, adminUsername,
		fmt.Sprintf("product:%d", productID),
		fmt.Sprintf("Approved: %s (seller: %s)", p.Nama, p.SellerName))

	handler.SetFlash(c, "ec_success", fmt.Sprintf("Produk \"%s\" berhasil diapprove.", p.Nama))
	c.Redirect(http.StatusFound, "/ecommerce/admin/products")
}

// RejectProduct rejects a pending product.
// POST /ecommerce/admin/products/:id/reject
func (h *AdminHandler) RejectProduct(c *gin.Context) {
	var productID int
	fmt.Sscanf(c.Param("id"), "%d", &productID)

	adminID := GetECUserID(c)
	adminUsername := GetECUsername(c)
	alasan := strings.TrimSpace(c.PostForm("alasan"))

	p := mock.FindECProductByID(productID)
	if p == nil {
		handler.SetFlash(c, "ec_error", "Produk tidak ditemukan.")
		c.Redirect(http.StatusFound, "/ecommerce/admin/products")
		return
	}

	p.Status = "REJECTED"
	detail := fmt.Sprintf("Rejected: %s (seller: %s)", p.Nama, p.SellerName)
	if alasan != "" {
		detail += ". Alasan: " + alasan
	}
	mock.LogAuditAction("REJECT_PRODUCT", adminID, adminUsername,
		fmt.Sprintf("product:%d", productID), detail)

	handler.SetFlash(c, "ec_success", fmt.Sprintf("Produk \"%s\" telah ditolak.", p.Nama))
	c.Redirect(http.StatusFound, "/ecommerce/admin/products")
}

// VoucherManagement shows all vouchers with CRUD options.
// GET /ecommerce/admin/vouchers
func (h *AdminHandler) VoucherManagement(c *gin.Context) {
	h.Render(c, "ec_base", "ecommerce/admin/vouchers", gin.H{
		"Title":    "Manajemen Voucher",
		"Active":   "admin_vouchers",
		"Vouchers": mock.ECVouchers,
	})
}

// CreateVoucher processes new voucher creation.
// POST /ecommerce/admin/vouchers/create
func (h *AdminHandler) CreateVoucher(c *gin.Context) {
	adminID := GetECUserID(c)
	adminUsername := GetECUsername(c)

	code := strings.ToUpper(strings.TrimSpace(c.PostForm("code")))
	deskripsi := strings.TrimSpace(c.PostForm("deskripsi"))
	tipeDiskon := c.PostForm("tipe_diskon")
	nilaiStr := c.PostForm("nilai_diskon")
	minStr := c.PostForm("min_pembelian")
	maksStr := c.PostForm("maks_diskon")
	kuotaStr := c.PostForm("kuota")
	berlakuSampai := c.PostForm("berlaku_sampai")

	if code == "" || deskripsi == "" || nilaiStr == "" {
		handler.SetFlash(c, "ec_error", "Kode, deskripsi, dan nilai diskon wajib diisi.")
		c.Redirect(http.StatusFound, "/ecommerce/admin/vouchers")
		return
	}

	if mock.FindECVoucherByCode(code) != nil {
		handler.SetFlash(c, "ec_error", fmt.Sprintf("Kode voucher %s sudah digunakan.", code))
		c.Redirect(http.StatusFound, "/ecommerce/admin/vouchers")
		return
	}

	nilaiDiskon, _ := strconv.ParseFloat(nilaiStr, 64)
	minPembelian, _ := strconv.ParseFloat(minStr, 64)
	maksDiskon, _ := strconv.ParseFloat(maksStr, 64)
	kuota, _ := strconv.Atoi(kuotaStr)
	if kuota <= 0 {
		kuota = 100
	}
	if berlakuSampai == "" {
		berlakuSampai = time.Now().AddDate(0, 3, 0).Format("2006-01-02")
	}

	nextID := 0
	for _, v := range mock.ECVouchers {
		if v.ID > nextID {
			nextID = v.ID
		}
	}
	newVoucher := model.Voucher{
		ID:            nextID + 1,
		Code:          code,
		Deskripsi:     deskripsi,
		TipeDiskon:    tipeDiskon,
		NilaiDiskon:   nilaiDiskon,
		MinPembelian:  minPembelian,
		MaksDiskon:    maksDiskon,
		Kuota:         kuota,
		Status:        "ACTIVE",
		BerlakuSampai: berlakuSampai,
	}
	mock.ECVouchers = append(mock.ECVouchers, newVoucher)
	mock.LogAuditAction("CREATE_VOUCHER", adminID, adminUsername,
		fmt.Sprintf("voucher:%d", newVoucher.ID),
		fmt.Sprintf("Created voucher %s: %s", code, deskripsi))

	handler.SetFlash(c, "ec_success", fmt.Sprintf("Voucher %s berhasil dibuat.", code))
	c.Redirect(http.StatusFound, "/ecommerce/admin/vouchers")
}

// ToggleVoucher toggles a voucher's status between ACTIVE and EXPIRED.
// POST /ecommerce/admin/vouchers/:id/toggle
func (h *AdminHandler) ToggleVoucher(c *gin.Context) {
	var voucherID int
	fmt.Sscanf(c.Param("id"), "%d", &voucherID)

	adminID := GetECUserID(c)
	adminUsername := GetECUsername(c)

	for i := range mock.ECVouchers {
		if mock.ECVouchers[i].ID == voucherID {
			v := &mock.ECVouchers[i]
			if v.Status == "ACTIVE" {
				v.Status = "EXPIRED"
				mock.LogAuditAction("DEACTIVATE_VOUCHER", adminID, adminUsername,
					fmt.Sprintf("voucher:%d", voucherID),
					fmt.Sprintf("Deactivated voucher %s", v.Code))
				handler.SetFlash(c, "ec_success", fmt.Sprintf("Voucher %s dinonaktifkan.", v.Code))
			} else {
				v.Status = "ACTIVE"
				mock.LogAuditAction("ACTIVATE_VOUCHER", adminID, adminUsername,
					fmt.Sprintf("voucher:%d", voucherID),
					fmt.Sprintf("Reactivated voucher %s", v.Code))
				handler.SetFlash(c, "ec_success", fmt.Sprintf("Voucher %s diaktifkan kembali.", v.Code))
			}
			break
		}
	}
	c.Redirect(http.StatusFound, "/ecommerce/admin/vouchers")
}

// OrderManagement shows all orders in the system.
// GET /ecommerce/admin/orders
func (h *AdminHandler) OrderManagement(c *gin.Context) {
	allOrders := mock.GetAllECOrders()

	statusFilter := c.Query("status")
	if statusFilter != "" {
		var filtered []model.ECOrder
		for _, o := range allOrders {
			if o.Status == statusFilter {
				filtered = append(filtered, o)
			}
		}
		allOrders = filtered
	}

	statusCounts := map[string]int{}
	for _, o := range mock.GetAllECOrders() {
		statusCounts[o.Status]++
	}

	h.Render(c, "ec_base", "ecommerce/admin/orders", gin.H{
		"Title":        "Manajemen Order",
		"Active":       "admin_orders",
		"Orders":       allOrders,
		"StatusFilter": statusFilter,
		"StatusCounts": statusCounts,
	})
}

// MemberManagement shows all e-commerce users.
// GET /ecommerce/admin/members
func (h *AdminHandler) MemberManagement(c *gin.Context) {
	allUsers := mock.ECommerceUsers

	roleFilter := c.Query("role")
	var filtered []model.ECommerceUser
	for _, u := range allUsers {
		if roleFilter == "" || u.Role == roleFilter || (roleFilter == "SELLER" && u.IsSellerActive) {
			filtered = append(filtered, u)
		}
	}

	h.Render(c, "ec_base", "ecommerce/admin/members", gin.H{
		"Title":      "Manajemen Member",
		"Active":     "admin_members",
		"Users":      filtered,
		"RoleFilter": roleFilter,
		"TotalUsers": len(allUsers),
	})
}

// PointsMonitoring shows all points transactions.
// GET /ecommerce/admin/points
func (h *AdminHandler) PointsMonitoring(c *gin.Context) {
	allTx := mock.ECPointsTransactions
	allPoints := mock.ECUserPoints

	// Total stats
	totalPoints := 0.0
	for _, up := range allPoints {
		totalPoints += up.Balance
	}

	// Reverse for newest-first
	reversed := make([]model.PointsTransaction, len(allTx))
	for i, j := 0, len(allTx)-1; j >= 0; i, j = i+1, j-1 {
		reversed[i] = allTx[j]
	}

	h.Render(c, "ec_base", "ecommerce/admin/points", gin.H{
		"Title":        "Monitoring Poin",
		"Active":       "admin_points",
		"Transactions": reversed,
		"UserPoints":   allPoints,
		"TotalPoints":  totalPoints,
	})
}

// AuditLog shows all system audit actions.
// GET /ecommerce/admin/audit
func (h *AdminHandler) AuditLog(c *gin.Context) {
	logs := mock.GetECAuditLogs()

	actionFilter := c.Query("action")
	if actionFilter != "" {
		var filtered []model.AuditLog
		for _, l := range logs {
			if strings.Contains(l.Action, actionFilter) {
				filtered = append(filtered, l)
			}
		}
		logs = filtered
	}

	h.Render(c, "ec_base", "ecommerce/admin/audit", gin.H{
		"Title":        "Audit Log",
		"Active":       "admin_audit",
		"Logs":         logs,
		"ActionFilter": actionFilter,
	})
}
