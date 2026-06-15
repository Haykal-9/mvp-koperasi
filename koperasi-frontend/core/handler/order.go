package handler

import (
	"net/http"
	"strconv"
	"strings"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	Render     Renderer
	OrderSvc   *service.OrderService
	ProductSvc *service.ProductService
}

func NewOrderHandler(render Renderer, orderSvc *service.OrderService, productSvc *service.ProductService) *OrderHandler {
	return &OrderHandler{Render: render, OrderSvc: orderSvc, ProductSvc: productSvc}
}

func (h *OrderHandler) ready(c *gin.Context) bool {
	if h.OrderSvc == nil || h.ProductSvc == nil {
		SetFlash(c, "error", "Fitur order membutuhkan koneksi database.")
		c.Redirect(http.StatusFound, "/dashboard")
		return false
	}
	return true
}

// resolved cart row (joined with product) for display
type cartRow struct {
	ProductID    int
	ProductNama  string
	FotoURL      string
	HargaSatuan  float64
	StokTersedia int
	Jumlah       int
	Subtotal     float64
}

func (h *OrderHandler) resolveCart(c *gin.Context, items []CartItem) ([]cartRow, float64, error) {
	rows := []cartRow{}
	total := 0.0
	for _, it := range items {
		p, err := h.ProductSvc.FindByID(c.Request.Context(), it.ProductID)
		if err != nil {
			return nil, 0, err
		}
		if p == nil {
			continue
		}
		sub := p.Harga * float64(it.Jumlah)
		rows = append(rows, cartRow{
			ProductID:    p.ID,
			ProductNama:  p.Nama,
			FotoURL:      p.FotoURL,
			HargaSatuan:  p.Harga,
			StokTersedia: p.Stok,
			Jumlah:       it.Jumlah,
			Subtotal:     sub,
		})
		total += sub
	}
	return rows, total, nil
}

// ===== GET /cart =====
func (h *OrderHandler) ShowCart(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	rows, total, err := h.resolveCart(c, GetCart(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	h.Render(c, "base", "order/cart", gin.H{
		"Title":  "Keranjang",
		"Active": "cart",
		"Items":  rows,
		"Total":  total,
	})
}

// ===== POST /cart/add =====
func (h *OrderHandler) AddToCart(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	pid, _ := strconv.Atoi(c.PostForm("product_id"))
	jumlah, _ := strconv.Atoi(c.PostForm("jumlah"))
	if jumlah < 1 {
		jumlah = 1
	}
	p, err := h.ProductSvc.FindByID(c.Request.Context(), pid)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if p == nil {
		SetFlash(c, "error", "Produk tidak ditemukan.")
		c.Redirect(http.StatusFound, "/products")
		return
	}
	if p.Stok < jumlah {
		SetFlash(c, "error", "Stok tidak mencukupi.")
		c.Redirect(http.StatusFound, "/products/"+strconv.Itoa(pid))
		return
	}
	AddOrIncrement(c, pid, jumlah)
	SetFlash(c, "success", "\""+p.Nama+"\" ditambahkan ke keranjang.")
	c.Redirect(http.StatusFound, "/cart")
}

// ===== POST /cart/update =====
func (h *OrderHandler) UpdateCart(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	pid, _ := strconv.Atoi(c.PostForm("product_id"))
	jumlah, _ := strconv.Atoi(c.PostForm("jumlah"))
	p, err := h.ProductSvc.FindByID(c.Request.Context(), pid)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if p != nil && jumlah > p.Stok {
		SetFlash(c, "error", "Jumlah melebihi stok tersedia.")
		c.Redirect(http.StatusFound, "/cart")
		return
	}
	SetQty(c, pid, jumlah)
	c.Redirect(http.StatusFound, "/cart")
}

// ===== POST /cart/remove =====
func (h *OrderHandler) RemoveFromCart(c *gin.Context) {
	pid, _ := strconv.Atoi(c.PostForm("product_id"))
	RemoveItem(c, pid)
	SetFlash(c, "success", "Item dihapus dari keranjang.")
	c.Redirect(http.StatusFound, "/cart")
}

// ===== GET /checkout =====
func (h *OrderHandler) ShowCheckout(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	rows, total, err := h.resolveCart(c, GetCart(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if len(rows) == 0 {
		SetFlash(c, "error", "Keranjang kosong.")
		c.Redirect(http.StatusFound, "/cart")
		return
	}
	h.Render(c, "base", "order/checkout", gin.H{
		"Title":       "Checkout",
		"Active":      "cart",
		"Items":       rows,
		"Total":       total,
		"FeeKoperasi": h.OrderSvc.Fee(total),
	})
}

// ===== POST /checkout =====
func (h *OrderHandler) DoCheckout(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	rows, total, err := h.resolveCart(c, GetCart(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if len(rows) == 0 {
		SetFlash(c, "error", "Keranjang kosong.")
		c.Redirect(http.StatusFound, "/cart")
		return
	}
	metode := strings.TrimSpace(c.PostForm("metode_bayar"))
	if metode == "" {
		SetFlash(c, "error", "Pilih metode pembayaran.")
		c.Redirect(http.StatusFound, "/checkout")
		return
	}
	// re-validate stock
	for _, r := range rows {
		if r.Jumlah > r.StokTersedia {
			SetFlash(c, "error", "Stok \""+r.ProductNama+"\" tidak mencukupi.")
			c.Redirect(http.StatusFound, "/cart")
			return
		}
	}

	sess := sessions.Default(c)
	pembeliNama, _ := sess.Get("user_nama").(string)
	if pembeliNama == "" {
		pembeliNama = "Walk-in"
	}

	items := make([]model.OrderItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, model.OrderItem{
			ProductID:   r.ProductID,
			ProductNama: r.ProductNama,
			Jumlah:      r.Jumlah,
			HargaSatuan: r.HargaSatuan,
			Subtotal:    r.Subtotal,
		})
	}

	saved, err := h.OrderSvc.Checkout(c.Request.Context(), pembeliNama, items, total, metode)
	if err != nil {
		SetFlash(c, "error", "Gagal membuat order.")
		c.Redirect(http.StatusFound, "/checkout")
		return
	}
	ClearCart(c)
	SetFlash(c, "success", "Order "+saved.NomorOrder+" berhasil dibuat.")
	c.Redirect(http.StatusFound, "/orders/"+strconv.Itoa(saved.ID))
}

// ===== GET /orders =====
// ANGGOTA hanya melihat order miliknya; OWNER/KASIR melihat semua order.
func (h *OrderHandler) List(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	statusFilter := strings.ToUpper(c.Query("status"))
	role := CurrentUserRole(c)
	userNama := CurrentUserNama(c)
	scopeOwn := role == "ANGGOTA"
	scope := ""
	if scopeOwn {
		scope = userNama
	}

	all, err := h.OrderSvc.List(c.Request.Context(), scope)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	out := []model.Order{}
	cnt := map[string]int{}
	for _, o := range all {
		cnt[o.Status]++
		if statusFilter != "" && statusFilter != "ALL" && o.Status != statusFilter {
			continue
		}
		out = append(out, o)
	}
	h.Render(c, "base", "order/history", gin.H{
		"Title":         "Riwayat Order",
		"Active":        "orders",
		"Orders":        out,
		"StatusFilter":  statusFilter,
		"CountAll":      len(all),
		"CountDibayar":  cnt["DIBAYAR"],
		"CountDikirim":  cnt["DIKIRIM"],
		"CountSelesai":  cnt["SELESAI"],
		"CountDisputed": cnt["DISPUTED"],
		"CountBatal":    cnt["BATAL"],
		"ScopeOwn":      scopeOwn,
	})
}

// ===== GET /orders/:id =====
// ANGGOTA hanya boleh akses order miliknya sendiri.
func (h *OrderHandler) Detail(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	o, err := h.OrderSvc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if o == nil {
		c.String(http.StatusNotFound, "Order tidak ditemukan")
		return
	}
	if CurrentUserRole(c) == "ANGGOTA" && o.PembeliNama != CurrentUserNama(c) {
		SetFlash(c, "error", "Anda hanya bisa melihat pesanan milik sendiri.")
		c.Redirect(http.StatusFound, "/orders")
		return
	}
	timeline := buildTimeline(o)
	netPenjual := o.TotalHarga - o.FeeKoperasi
	h.Render(c, "base", "order/detail", gin.H{
		"Title":      "Order " + o.NomorOrder,
		"Active":     "orders",
		"Order":      o,
		"Timeline":   timeline,
		"NetPenjual": netPenjual,
		"CanShip":    IsOwnerOrKasir(c),
		"CanResolve": CurrentUserRole(c) == "OWNER",
	})
}

type timelineStep struct {
	Label string
	Done  bool
	Time  string
	Icon  string
}

func buildTimeline(o *model.Order) []timelineStep {
	steps := []timelineStep{
		{Label: "Dibuat", Icon: "bi-cart-check"},
		{Label: "Dibayar", Icon: "bi-credit-card"},
		{Label: "Dikirim", Icon: "bi-truck"},
		{Label: "Selesai", Icon: "bi-check-circle"},
	}
	stage := 0
	switch o.Status {
	case "PENDING":
		stage = 1
	case "DIBAYAR":
		stage = 2
	case "DIKIRIM":
		stage = 3
	case "SELESAI":
		stage = 4
	case "DISPUTED":
		stage = 3
	case "BATAL":
		stage = 1
	}
	for i := range steps {
		if i < stage {
			steps[i].Done = true
		}
	}
	if len(steps) > 0 {
		steps[0].Time = o.CreatedAt
	}
	return steps
}

// ===== POST /orders/:id/ship =====
func (h *OrderHandler) Ship(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	o, err := h.OrderSvc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if o == nil {
		c.String(http.StatusNotFound, "Order tidak ditemukan")
		return
	}
	if o.Status != "DIBAYAR" {
		SetFlash(c, "error", "Hanya order DIBAYAR yang bisa dikirim.")
		c.Redirect(http.StatusFound, "/orders/"+strconv.Itoa(id))
		return
	}
	if err := h.OrderSvc.Ship(c.Request.Context(), id); err != nil {
		SetFlash(c, "error", "Gagal memperbarui order.")
		c.Redirect(http.StatusFound, "/orders/"+strconv.Itoa(id))
		return
	}
	SetFlash(c, "success", "Order "+o.NomorOrder+" ditandai DIKIRIM.")
	c.Redirect(http.StatusFound, "/orders/"+strconv.Itoa(id))
}

// ===== POST /orders/:id/complete =====
// Konfirmasi terima — boleh pembeli sendiri, OWNER, atau KASIR.
func (h *OrderHandler) Complete(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	o, err := h.OrderSvc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if o == nil {
		c.String(http.StatusNotFound, "Order tidak ditemukan")
		return
	}
	if !IsOwnerOrKasir(c) && o.PembeliNama != CurrentUserNama(c) {
		SetFlash(c, "error", "Anda hanya bisa mengonfirmasi pesanan milik sendiri.")
		c.Redirect(http.StatusFound, "/orders")
		return
	}
	if o.Status != "DIKIRIM" {
		SetFlash(c, "error", "Hanya order DIKIRIM yang bisa diselesaikan.")
		c.Redirect(http.StatusFound, "/orders/"+strconv.Itoa(id))
		return
	}
	if err := h.OrderSvc.Complete(c.Request.Context(), id); err != nil {
		SetFlash(c, "error", "Gagal memperbarui order.")
		c.Redirect(http.StatusFound, "/orders/"+strconv.Itoa(id))
		return
	}
	SetFlash(c, "success", "Order "+o.NomorOrder+" SELESAI. Split payment dijalankan.")
	c.Redirect(http.StatusFound, "/orders/"+strconv.Itoa(id))
}

// ===== GET /orders/:id/complain =====
// Hanya pemilik order yang boleh membuka form komplain (kecuali staff).
func (h *OrderHandler) ShowComplain(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	o, err := h.OrderSvc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if o == nil {
		c.String(http.StatusNotFound, "Order tidak ditemukan")
		return
	}
	if !IsOwnerOrKasir(c) && o.PembeliNama != CurrentUserNama(c) {
		SetFlash(c, "error", "Anda hanya bisa mengajukan komplain untuk pesanan milik sendiri.")
		c.Redirect(http.StatusFound, "/orders")
		return
	}
	h.Render(c, "base", "order/complain", gin.H{
		"Title":  "Komplain Order " + o.NomorOrder,
		"Active": "orders",
		"Order":  o,
	})
}

// ===== POST /orders/:id/complain =====
func (h *OrderHandler) DoComplain(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	o, err := h.OrderSvc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if o == nil {
		c.String(http.StatusNotFound, "Order tidak ditemukan")
		return
	}
	if !IsOwnerOrKasir(c) && o.PembeliNama != CurrentUserNama(c) {
		SetFlash(c, "error", "Anda hanya bisa mengajukan komplain untuk pesanan milik sendiri.")
		c.Redirect(http.StatusFound, "/orders")
		return
	}
	if o.Status != "DIKIRIM" && o.Status != "SELESAI" {
		SetFlash(c, "error", "Komplain hanya untuk order DIKIRIM atau SELESAI.")
		c.Redirect(http.StatusFound, "/orders/"+strconv.Itoa(id))
		return
	}
	alasan := strings.TrimSpace(c.PostForm("alasan"))
	bukti := strings.TrimSpace(c.PostForm("bukti"))
	if alasan == "" {
		SetFlash(c, "error", "Alasan komplain wajib diisi.")
		c.Redirect(http.StatusFound, "/orders/"+strconv.Itoa(id)+"/complain")
		return
	}
	if err := h.OrderSvc.Complain(c.Request.Context(), id, alasan, bukti); err != nil {
		SetFlash(c, "error", "Gagal mengirim komplain.")
		c.Redirect(http.StatusFound, "/orders/"+strconv.Itoa(id))
		return
	}
	SetFlash(c, "success", "Komplain dikirim. Menunggu review pengurus.")
	c.Redirect(http.StatusFound, "/orders/"+strconv.Itoa(id))
}

// ===== GET /orders/complaints =====
func (h *OrderHandler) Complaints(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	out, err := h.OrderSvc.Complaints(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	h.Render(c, "base", "order/complaints", gin.H{
		"Title":  "Daftar Komplain",
		"Active": "complaints",
		"Orders": out,
	})
}

// ===== POST /orders/:id/resolve =====
func (h *OrderHandler) Resolve(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	o, err := h.OrderSvc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if o == nil {
		c.String(http.StatusNotFound, "Order tidak ditemukan")
		return
	}
	if o.Status != "DISPUTED" {
		SetFlash(c, "error", "Order tidak dalam status DISPUTED.")
		c.Redirect(http.StatusFound, "/orders/complaints")
		return
	}
	keputusan := c.PostForm("keputusan") // approve_retur | reject
	switch keputusan {
	case "approve_retur":
		if err := h.OrderSvc.Resolve(c.Request.Context(), id, true); err != nil {
			SetFlash(c, "error", "Gagal memproses komplain.")
			c.Redirect(http.StatusFound, "/orders/complaints")
			return
		}
		SetFlash(c, "success", "Komplain disetujui. Order "+o.NomorOrder+" dibatalkan & stok dikembalikan.")
	case "reject":
		if err := h.OrderSvc.Resolve(c.Request.Context(), id, false); err != nil {
			SetFlash(c, "error", "Gagal memproses komplain.")
			c.Redirect(http.StatusFound, "/orders/complaints")
			return
		}
		SetFlash(c, "success", "Komplain ditolak. Order "+o.NomorOrder+" diselesaikan.")
	default:
		SetFlash(c, "error", "Keputusan tidak valid.")
		c.Redirect(http.StatusFound, "/orders/complaints")
		return
	}
	c.Redirect(http.StatusFound, "/orders/complaints")
}
