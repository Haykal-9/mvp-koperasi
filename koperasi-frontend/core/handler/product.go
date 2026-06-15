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

type ProductHandler struct {
	Render Renderer
	Svc    *service.ProductService
}

func NewProductHandler(render Renderer, svc *service.ProductService) *ProductHandler {
	return &ProductHandler{Render: render, Svc: svc}
}

// ready memastikan service (DB) tersedia; jika tidak, alihkan dengan pesan.
func (h *ProductHandler) ready(c *gin.Context) bool {
	if h.Svc == nil {
		SetFlash(c, "error", "Fitur produk membutuhkan koneksi database.")
		c.Redirect(http.StatusFound, "/dashboard")
		return false
	}
	return true
}

// ===== GET /products =====
func (h *ProductHandler) Catalog(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	q := c.Query("q")
	kategori := c.Query("kategori")

	products, categories, err := h.Svc.Catalog(c.Request.Context(), q, kategori)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "base", "product/catalog", gin.H{
		"Title":      "Katalog Produk",
		"Active":     "products",
		"Products":   products,
		"Q":          q,
		"Kategori":   kategori,
		"Categories": categories,
	})
}

// ===== GET /products/:id =====
func (h *ProductHandler) Detail(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	p, err := h.Svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if p == nil {
		c.String(http.StatusNotFound, "Produk tidak ditemukan")
		return
	}
	h.Render(c, "base", "product/detail", gin.H{
		"Title":   p.Nama,
		"Active":  "products",
		"Product": p,
	})
}

// ===== GET /products/create =====
func (h *ProductHandler) ShowCreate(c *gin.Context) {
	formNama, _ := PopFlash(c, "form_nama")
	formKategori, _ := PopFlash(c, "form_kategori")
	formHarga, _ := PopFlash(c, "form_harga")
	formStok, _ := PopFlash(c, "form_stok")
	formDeskripsi, _ := PopFlash(c, "form_deskripsi")
	formMin, _ := PopFlash(c, "form_min")

	h.Render(c, "base", "product/create", gin.H{
		"Title":         "Tambah Produk",
		"Active":        "product-create",
		"FormNama":      formNama,
		"FormKategori":  formKategori,
		"FormHarga":     formHarga,
		"FormStok":      formStok,
		"FormDeskripsi": formDeskripsi,
		"FormMin":       formMin,
	})
}

// ===== POST /products/create =====
func (h *ProductHandler) DoCreate(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	nama := strings.TrimSpace(c.PostForm("nama"))
	kategori := strings.TrimSpace(c.PostForm("kategori"))
	hargaStr := c.PostForm("harga")
	stokStr := c.PostForm("stok")
	minStr := c.PostForm("batas_min")
	deskripsi := strings.TrimSpace(c.PostForm("deskripsi"))

	preserve := func() {
		SetFlash(c, "form_nama", nama)
		SetFlash(c, "form_kategori", kategori)
		SetFlash(c, "form_harga", hargaStr)
		SetFlash(c, "form_stok", stokStr)
		SetFlash(c, "form_min", minStr)
		SetFlash(c, "form_deskripsi", deskripsi)
	}

	harga, _ := strconv.ParseFloat(hargaStr, 64)
	stok, _ := strconv.Atoi(stokStr)
	minStok, _ := strconv.Atoi(minStr)

	if err := h.Svc.ValidateCreate(nama, kategori, harga, stok); err != nil {
		SetFlash(c, "error", err.Error())
		preserve()
		c.Redirect(http.StatusFound, "/products/create")
		return
	}

	// auto-approve jika OWNER/KASIR, pending jika peran lain
	sess := sessions.Default(c)
	role, _ := sess.Get("user_role").(string)
	namaUser, _ := sess.Get("user_nama").(string)
	if namaUser == "" {
		namaUser = "Anggota"
	}
	status := h.Svc.StatusForRole(role)

	_, err := h.Svc.Create(c.Request.Context(), model.Product{
		Nama:             nama,
		Kategori:         kategori,
		Harga:            harga,
		Stok:             stok,
		BatasStokMinimum: minStok,
		Deskripsi:        deskripsi,
		PenjualNama:      namaUser,
		Status:           status,
	})
	if err != nil {
		SetFlash(c, "error", "Gagal menyimpan produk.")
		preserve()
		c.Redirect(http.StatusFound, "/products/create")
		return
	}

	if status == "APPROVED" {
		SetFlash(c, "success", "Produk \""+nama+"\" berhasil ditambahkan dan langsung tayang.")
		c.Redirect(http.StatusFound, "/products")
	} else {
		SetFlash(c, "success", "Produk \""+nama+"\" terkirim. Menunggu review pengurus.")
		c.Redirect(http.StatusFound, "/products/review")
	}
}

// ===== GET /products/review =====
func (h *ProductHandler) Review(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	pending, err := h.Svc.PendingReview(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	h.Render(c, "base", "product/review", gin.H{
		"Title":    "Review Produk Pending",
		"Active":   "product-review",
		"Products": pending,
	})
}

// ===== POST /products/:id/approve =====
func (h *ProductHandler) Approve(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	p, err := h.Svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if p == nil {
		c.String(http.StatusNotFound, "Produk tidak ditemukan")
		return
	}
	if p.Status != "PENDING" {
		SetFlash(c, "error", "Produk tidak dalam status PENDING.")
		c.Redirect(http.StatusFound, "/products/review")
		return
	}
	if err := h.Svc.Approve(c.Request.Context(), id); err != nil {
		SetFlash(c, "error", "Gagal menyetujui produk.")
		c.Redirect(http.StatusFound, "/products/review")
		return
	}
	SetFlash(c, "success", "Produk \""+p.Nama+"\" disetujui dan tayang di katalog.")
	c.Redirect(http.StatusFound, "/products/review")
}

// ===== POST /products/:id/reject =====
func (h *ProductHandler) Reject(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	p, err := h.Svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if p == nil {
		c.String(http.StatusNotFound, "Produk tidak ditemukan")
		return
	}
	if p.Status != "PENDING" {
		SetFlash(c, "error", "Produk tidak dalam status PENDING.")
		c.Redirect(http.StatusFound, "/products/review")
		return
	}
	if err := h.Svc.Reject(c.Request.Context(), id); err != nil {
		SetFlash(c, "error", "Gagal menolak produk.")
		c.Redirect(http.StatusFound, "/products/review")
		return
	}
	SetFlash(c, "success", "Produk \""+p.Nama+"\" ditolak.")
	c.Redirect(http.StatusFound, "/products/review")
}

// ===== GET /products/:id/stock =====
func (h *ProductHandler) ShowStock(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	p, err := h.Svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if p == nil {
		c.String(http.StatusNotFound, "Produk tidak ditemukan")
		return
	}
	history, err := h.Svc.StockHistory(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	h.Render(c, "base", "product/stock", gin.H{
		"Title":      "Stok — " + p.Nama,
		"Active":     "products",
		"Product":    p,
		"StokRendah": p.Stok < p.BatasStokMinimum,
		"History":    history,
	})
}

// ===== POST /products/:id/stock =====
func (h *ProductHandler) DoStock(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	tipe := strings.ToUpper(c.PostForm("tipe"))
	jumlah, _ := strconv.Atoi(c.PostForm("jumlah"))
	keterangan := strings.TrimSpace(c.PostForm("keterangan"))

	if err := h.Svc.ValidateStock(tipe, jumlah); err != nil {
		SetFlash(c, "error", err.Error())
		c.Redirect(http.StatusFound, "/products/"+strconv.Itoa(id)+"/stock")
		return
	}
	if err := h.Svc.AdjustStock(c.Request.Context(), id, tipe, jumlah, keterangan); err != nil {
		SetFlash(c, "error", err.Error())
		c.Redirect(http.StatusFound, "/products/"+strconv.Itoa(id)+"/stock")
		return
	}
	SetFlash(c, "success", "Stok diperbarui.")
	c.Redirect(http.StatusFound, "/products/"+strconv.Itoa(id)+"/stock")
}
