package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"koperasi-frontend/core/mock"
	"koperasi-frontend/core/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	Render Renderer
}

func NewOrderHandler(render Renderer) *OrderHandler {
	return &OrderHandler{Render: render}
}

// resolved cart row (joined with product) for display
type cartRow struct {
	ProductID   int
	ProductNama string
	FotoURL     string
	HargaSatuan float64
	StokTersedia int
	Jumlah      int
	Subtotal    float64
}

func resolveCart(items []CartItem) ([]cartRow, float64) {
	rows := []cartRow{}
	total := 0.0
	for _, it := range items {
		p := mock.FindProductByID(it.ProductID)
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
	return rows, total
}

// ===== GET /cart =====
func (h *OrderHandler) ShowCart(c *gin.Context) {
	rows, total := resolveCart(GetCart(c))
	h.Render(c, "base", "order/cart", gin.H{
		"Title":  "Keranjang",
		"Active": "cart",
		"Items":  rows,
		"Total":  total,
	})
}

// ===== POST /cart/add =====
func (h *OrderHandler) AddToCart(c *gin.Context) {
	pid, _ := strconv.Atoi(c.PostForm("product_id"))
	jumlah, _ := strconv.Atoi(c.PostForm("jumlah"))
	if jumlah < 1 {
		jumlah = 1
	}
	p := mock.FindProductByID(pid)
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
	pid, _ := strconv.Atoi(c.PostForm("product_id"))
	jumlah, _ := strconv.Atoi(c.PostForm("jumlah"))
	p := mock.FindProductByID(pid)
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
	rows, total := resolveCart(GetCart(c))
	if len(rows) == 0 {
		SetFlash(c, "error", "Keranjang kosong.")
		c.Redirect(http.StatusFound, "/cart")
		return
	}
	fee := total * 0.03
	h.Render(c, "base", "order/checkout", gin.H{
		"Title":       "Checkout",
		"Active":      "cart",
		"Items":       rows,
		"Total":       total,
		"FeeKoperasi": fee,
	})
}

// ===== POST /checkout =====
func (h *OrderHandler) DoCheckout(c *gin.Context) {
	rows, total := resolveCart(GetCart(c))
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

	order := model.Order{
		PembeliNama: pembeliNama,
		Items:       items,
		TotalHarga:  total,
		FeeKoperasi: total * 0.03,
		Status:      "DIBAYAR",
		MetodeBayar: metode,
		CreatedAt:   time.Now().Format("2006-01-02 15:04"),
	}
	saved := mock.AppendOrder(order)
	ClearCart(c)
	SetFlash(c, "success", "Order "+saved.NomorOrder+" berhasil dibuat.")
	c.Redirect(http.StatusFound, "/orders/"+strconv.Itoa(saved.ID))
}

// ===== GET /orders =====
// ANGGOTA hanya melihat order miliknya; OWNER/KASIR melihat semua order.
func (h *OrderHandler) List(c *gin.Context) {
	statusFilter := strings.ToUpper(c.Query("status"))
	role := CurrentUserRole(c)
	userNama := CurrentUserNama(c)
	scopeOwn := role == "ANGGOTA"

	out := []model.Order{}
	cnt := map[string]int{}
	totalAll := 0
	for i := len(mock.Orders) - 1; i >= 0; i-- {
		o := mock.Orders[i]
		if scopeOwn && o.PembeliNama != userNama {
			continue
		}
		totalAll++
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
		"CountAll":      totalAll,
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
	id, _ := strconv.Atoi(c.Param("id"))
	o := mock.FindOrderByID(id)
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
	id, _ := strconv.Atoi(c.Param("id"))
	o := mock.FindOrderByID(id)
	if o == nil {
		c.String(http.StatusNotFound, "Order tidak ditemukan")
		return
	}
	if o.Status != "DIBAYAR" {
		SetFlash(c, "error", "Hanya order DIBAYAR yang bisa dikirim.")
		c.Redirect(http.StatusFound, "/orders/"+strconv.Itoa(id))
		return
	}
	o.Status = "DIKIRIM"
	SetFlash(c, "success", "Order "+o.NomorOrder+" ditandai DIKIRIM.")
	c.Redirect(http.StatusFound, "/orders/"+strconv.Itoa(id))
}

// ===== POST /orders/:id/complete =====
// Konfirmasi terima — boleh pembeli sendiri, OWNER, atau KASIR.
func (h *OrderHandler) Complete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	o := mock.FindOrderByID(id)
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
	o.Status = "SELESAI"
	SetFlash(c, "success", "Order "+o.NomorOrder+" SELESAI. Split payment dijalankan.")
	c.Redirect(http.StatusFound, "/orders/"+strconv.Itoa(id))
}

// ===== GET /orders/:id/complain =====
// Hanya pemilik order yang boleh membuka form komplain (kecuali staff).
func (h *OrderHandler) ShowComplain(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	o := mock.FindOrderByID(id)
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
	id, _ := strconv.Atoi(c.Param("id"))
	o := mock.FindOrderByID(id)
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
	o.Status = "DISPUTED"
	o.KomplainAlasan = alasan
	o.KomplainBukti = bukti
	o.KomplainTanggal = time.Now().Format("2006-01-02 15:04")
	SetFlash(c, "success", "Komplain dikirim. Menunggu review pengurus.")
	c.Redirect(http.StatusFound, "/orders/"+strconv.Itoa(id))
}

// ===== GET /orders/complaints =====
func (h *OrderHandler) Complaints(c *gin.Context) {
	out := []model.Order{}
	for _, o := range mock.Orders {
		if o.Status == "DISPUTED" {
			out = append(out, o)
		}
	}
	h.Render(c, "base", "order/complaints", gin.H{
		"Title":  "Daftar Komplain",
		"Active": "complaints",
		"Orders": out,
	})
}

// ===== POST /orders/:id/resolve =====
func (h *OrderHandler) Resolve(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	o := mock.FindOrderByID(id)
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
		o.Status = "BATAL"
		// kembalikan stok
		for _, it := range o.Items {
			_ = mock.AppendStockChange(it.ProductID, "KOREKSI", it.Jumlah,
				"Retur dari komplain "+o.NomorOrder, time.Now().Format("2006-01-02"))
		}
		SetFlash(c, "success", "Komplain disetujui. Order "+o.NomorOrder+" dibatalkan & stok dikembalikan.")
	case "reject":
		o.Status = "SELESAI"
		SetFlash(c, "success", "Komplain ditolak. Order "+o.NomorOrder+" diselesaikan.")
	default:
		SetFlash(c, "error", "Keputusan tidak valid.")
		c.Redirect(http.StatusFound, "/orders/complaints")
		return
	}
	c.Redirect(http.StatusFound, "/orders/complaints")
}
