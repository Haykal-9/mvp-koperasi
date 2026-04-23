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

type ProductHandler struct {
	Render Renderer
}

func NewProductHandler(render Renderer) *ProductHandler {
	return &ProductHandler{Render: render}
}

// ===== GET /products =====
func (h *ProductHandler) Catalog(c *gin.Context) {
	q := strings.ToLower(strings.TrimSpace(c.Query("q")))
	kategori := c.Query("kategori")

	products := []model.Product{}
	categorySet := map[string]bool{}
	for _, p := range mock.Products {
		categorySet[p.Kategori] = true
		// hide PENDING/REJECTED dari katalog publik
		if p.Status != "APPROVED" {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(p.Nama), q) {
			continue
		}
		if kategori != "" && p.Kategori != kategori {
			continue
		}
		products = append(products, p)
	}

	categories := []string{}
	for k := range categorySet {
		categories = append(categories, k)
	}

	h.Render(c, "base", "product/catalog", gin.H{
		"Title":      "Katalog Produk",
		"Active":     "products",
		"Products":   products,
		"Q":          c.Query("q"),
		"Kategori":   kategori,
		"Categories": categories,
	})
}

// ===== GET /products/:id =====
func (h *ProductHandler) Detail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	p := mock.FindProductByID(id)
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

	if nama == "" || kategori == "" || harga <= 0 || stok < 0 {
		SetFlash(c, "error", "Nama, kategori, harga (>0), dan stok (≥0) wajib diisi.")
		preserve()
		c.Redirect(http.StatusFound, "/products/create")
		return
	}
	if minStok < 0 {
		minStok = 0
	}

	// auto-approve jika OWNER/KASIR, pending jika ANGGOTA
	sess := sessions.Default(c)
	role, _ := sess.Get("user_role").(string)
	nama_user, _ := sess.Get("user_nama").(string)
	status := "PENDING"
	if role == "OWNER" || role == "KASIR" {
		status = "APPROVED"
	}
	if nama_user == "" {
		nama_user = "Anggota"
	}

	id := mock.NextProductID()
	mock.Products = append(mock.Products, model.Product{
		ID:               id,
		Nama:             nama,
		Kategori:         kategori,
		Harga:            harga,
		Stok:             stok,
		BatasStokMinimum: minStok,
		Deskripsi:        deskripsi,
		FotoURL:          "https://placehold.co/400x300?text=" + strings.ReplaceAll(nama, " ", "+"),
		PenjualNama:      nama_user,
		Status:           status,
	})
	if stok > 0 {
		_ = mock.AppendStockChange(id, "RESTOCK", stok, "Stok awal saat ditambahkan", time.Now().Format("2006-01-02"))
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
	pending := []model.Product{}
	for _, p := range mock.Products {
		if p.Status == "PENDING" {
			pending = append(pending, p)
		}
	}
	h.Render(c, "base", "product/review", gin.H{
		"Title":    "Review Produk Pending",
		"Active":   "product-review",
		"Products": pending,
	})
}

// ===== POST /products/:id/approve =====
func (h *ProductHandler) Approve(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	p := mock.FindProductByID(id)
	if p == nil {
		c.String(http.StatusNotFound, "Produk tidak ditemukan")
		return
	}
	if p.Status != "PENDING" {
		SetFlash(c, "error", "Produk tidak dalam status PENDING.")
		c.Redirect(http.StatusFound, "/products/review")
		return
	}
	p.Status = "APPROVED"
	SetFlash(c, "success", "Produk \""+p.Nama+"\" disetujui dan tayang di katalog.")
	c.Redirect(http.StatusFound, "/products/review")
}

// ===== POST /products/:id/reject =====
func (h *ProductHandler) Reject(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	p := mock.FindProductByID(id)
	if p == nil {
		c.String(http.StatusNotFound, "Produk tidak ditemukan")
		return
	}
	if p.Status != "PENDING" {
		SetFlash(c, "error", "Produk tidak dalam status PENDING.")
		c.Redirect(http.StatusFound, "/products/review")
		return
	}
	p.Status = "REJECTED"
	SetFlash(c, "success", "Produk \""+p.Nama+"\" ditolak.")
	c.Redirect(http.StatusFound, "/products/review")
}

// ===== GET /products/:id/stock =====
func (h *ProductHandler) ShowStock(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	p := mock.FindProductByID(id)
	if p == nil {
		c.String(http.StatusNotFound, "Produk tidak ditemukan")
		return
	}
	h.Render(c, "base", "product/stock", gin.H{
		"Title":     "Stok — " + p.Nama,
		"Active":    "products",
		"Product":   p,
		"StokRendah": p.Stok < p.BatasStokMinimum,
		"History":   mock.StockChangesByProduct(id),
	})
}

// ===== POST /products/:id/stock =====
func (h *ProductHandler) DoStock(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	p := mock.FindProductByID(id)
	if p == nil {
		c.String(http.StatusNotFound, "Produk tidak ditemukan")
		return
	}
	tipe := strings.ToUpper(c.PostForm("tipe"))
	jumlah, _ := strconv.Atoi(c.PostForm("jumlah"))
	keterangan := strings.TrimSpace(c.PostForm("keterangan"))

	if tipe != "RESTOCK" && tipe != "KOREKSI" {
		SetFlash(c, "error", "Tipe perubahan stok tidak valid.")
		c.Redirect(http.StatusFound, "/products/"+strconv.Itoa(id)+"/stock")
		return
	}
	if jumlah == 0 {
		SetFlash(c, "error", "Jumlah tidak boleh 0.")
		c.Redirect(http.StatusFound, "/products/"+strconv.Itoa(id)+"/stock")
		return
	}
	if tipe == "RESTOCK" && jumlah < 0 {
		SetFlash(c, "error", "Restock harus berupa angka positif.")
		c.Redirect(http.StatusFound, "/products/"+strconv.Itoa(id)+"/stock")
		return
	}

	if err := mock.AppendStockChange(id, tipe, jumlah, keterangan, time.Now().Format("2006-01-02")); err != nil {
		SetFlash(c, "error", err.Error())
		c.Redirect(http.StatusFound, "/products/"+strconv.Itoa(id)+"/stock")
		return
	}
	SetFlash(c, "success", "Stok diperbarui.")
	c.Redirect(http.StatusFound, "/products/"+strconv.Itoa(id)+"/stock")
}
