package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoanHandler struct {
	Render Renderer
	Svc    *service.LoanService
}

func NewLoanHandler(render Renderer, svc *service.LoanService) *LoanHandler {
	return &LoanHandler{Render: render, Svc: svc}
}

// ready memastikan service (DB) tersedia; jika tidak, alihkan dengan pesan.
func (h *LoanHandler) ready(c *gin.Context) bool {
	if h.Svc == nil {
		SetFlash(c, "error", "Fitur pinjaman membutuhkan koneksi database.")
		c.Redirect(http.StatusFound, "/dashboard")
		return false
	}
	return true
}

// ===== GET /loans/apply =====
func (h *LoanHandler) ShowApply(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	sess := sessions.Default(c)
	userNama, _ := sess.Get("user_nama").(string)

	// Informasi credit scoring (total simpanan & pinjaman aktif anggota).
	totalSimpanan, tunggakanAktif, err := h.Svc.CreditInfo(c.Request.Context(), userNama)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "base", "loan/apply", gin.H{
		"Title":          "Ajukan Pinjaman",
		"Active":         "loan-apply",
		"TotalSimpanan":  totalSimpanan,
		"TunggakanAktif": tunggakanAktif,
	})
}

// ===== POST /loans/apply =====
func (h *LoanHandler) DoApply(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	sess := sessions.Default(c)
	userNama, _ := sess.Get("user_nama").(string)
	if userNama == "" {
		userNama = "Walk-in"
	}

	nominalStr := strings.ReplaceAll(c.PostForm("nominal"), ".", "")
	nominal, _ := strconv.ParseFloat(nominalStr, 64)
	tenor, _ := strconv.Atoi(c.PostForm("tenor"))
	tujuan := strings.TrimSpace(c.PostForm("tujuan"))

	id, err := h.Svc.Apply(c.Request.Context(), userNama, nominal, tenor, tujuan)
	if err != nil {
		SetFlash(c, "error", err.Error())
		c.Redirect(http.StatusFound, "/loans/apply")
		return
	}
	SetFlash(c, "success", fmt.Sprintf("Pengajuan pinjaman berhasil (ID: %d). Menunggu persetujuan.", id))
	c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
}

// ===== GET /loans =====
// ANGGOTA hanya melihat pinjaman miliknya; OWNER/KASIR melihat semua.
func (h *LoanHandler) List(c *gin.Context) {
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
	loans, err := h.Svc.List(c.Request.Context(), scope)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	out := make([]model.Loan, 0, len(loans))
	cnt := map[string]int{}
	for _, l := range loans {
		cnt[l.Status]++
		if statusFilter != "" && statusFilter != "ALL" && l.Status != statusFilter {
			continue
		}
		out = append(out, l)
	}
	h.Render(c, "base", "loan/list", gin.H{
		"Title":          "Daftar Pinjaman",
		"Active":         "loans",
		"Loans":          out,
		"StatusFilter":   statusFilter,
		"CountAll":       len(loans),
		"CountPending":   cnt["PENDING"],
		"CountDisetujui": cnt["DISETUJUI"],
		"CountAktif":     cnt["AKTIF"],
		"CountLunas":     cnt["LUNAS"],
		"CountDitolak":   cnt["DITOLAK"],
		"ScopeOwn":       scopeOwn,
	})
}

// ===== GET /loans/:id =====
// ANGGOTA hanya boleh akses pinjaman miliknya.
func (h *LoanHandler) Detail(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	l, err := h.Svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if l == nil {
		c.String(http.StatusNotFound, "Pinjaman tidak ditemukan")
		return
	}
	if CurrentUserRole(c) == "ANGGOTA" && l.MemberNama != CurrentUserNama(c) {
		SetFlash(c, "error", "Anda hanya bisa melihat pinjaman milik sendiri.")
		c.Redirect(http.StatusFound, "/loans")
		return
	}
	progressPct := h.Svc.Progress(l.Nominal, l.SisaPokok)
	nextInst := h.Svc.NextUnpaid(l.Installments)
	estAngsuran := h.Svc.EstimateMonthly(l.Nominal, l.BungaPersen, l.TenorBulan)

	h.Render(c, "base", "loan/detail", gin.H{
		"Title":       "Pinjaman #" + strconv.Itoa(l.ID),
		"Active":      "loans",
		"Loan":        l,
		"ProgressPct": progressPct,
		"NextInst":    nextInst,
		"EstAngsuran": estAngsuran,
		"CanReview":   CurrentUserRole(c) == "OWNER",
	})
}

// ===== POST /loans/:id/approve =====
func (h *LoanHandler) Approve(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	l, err := h.Svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if l == nil {
		c.String(http.StatusNotFound, "Pinjaman tidak ditemukan")
		return
	}
	if l.Status != "PENDING" {
		SetFlash(c, "error", "Hanya pinjaman PENDING yang bisa disetujui.")
		c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
		return
	}
	if err := h.Svc.Approve(c.Request.Context(), id); err != nil {
		SetFlash(c, "error", "Gagal menyetujui pinjaman.")
		c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
		return
	}
	SetFlash(c, "success", "Pinjaman disetujui. Silakan cairkan.")
	c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
}

// ===== POST /loans/:id/reject =====
func (h *LoanHandler) Reject(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	l, err := h.Svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if l == nil {
		c.String(http.StatusNotFound, "Pinjaman tidak ditemukan")
		return
	}
	if l.Status != "PENDING" {
		SetFlash(c, "error", "Hanya pinjaman PENDING yang bisa ditolak.")
		c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
		return
	}
	if err := h.Svc.Reject(c.Request.Context(), id); err != nil {
		SetFlash(c, "error", "Gagal menolak pinjaman.")
		c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
		return
	}
	SetFlash(c, "success", "Pinjaman ditolak.")
	c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
}

// ===== POST /loans/:id/disburse =====
func (h *LoanHandler) Disburse(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	l, err := h.Svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if l == nil {
		c.String(http.StatusNotFound, "Pinjaman tidak ditemukan")
		return
	}
	if l.Status != "DISETUJUI" {
		SetFlash(c, "error", "Hanya pinjaman DISETUJUI yang bisa dicairkan.")
		c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
		return
	}
	if err := h.Svc.Disburse(c.Request.Context(), l); err != nil {
		SetFlash(c, "error", "Gagal mencairkan pinjaman.")
		c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
		return
	}
	SetFlash(c, "success", "Pinjaman dicairkan. Jadwal angsuran dibuat.")
	c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
}

// ===== POST /loans/:id/pay =====
// Pemilik pinjaman bisa bayar sendiri; KASIR/OWNER bisa bayar atas nama anggota.
func (h *LoanHandler) Pay(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	l, err := h.Svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if l == nil {
		c.String(http.StatusNotFound, "Pinjaman tidak ditemukan")
		return
	}
	if !IsOwnerOrKasir(c) && l.MemberNama != CurrentUserNama(c) {
		SetFlash(c, "error", "Anda hanya bisa membayar angsuran pinjaman sendiri.")
		c.Redirect(http.StatusFound, "/loans")
		return
	}
	if l.Status != "AKTIF" {
		SetFlash(c, "error", "Hanya pinjaman AKTIF yang bisa dibayar.")
		c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
		return
	}

	_, lunas, err := h.Svc.Pay(c.Request.Context(), l, time.Now().Format("2006-01-02"))
	if err != nil {
		SetFlash(c, "error", "Tidak ada angsuran yang perlu dibayar.")
		c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
		return
	}

	if lunas {
		SetFlash(c, "success", "Angsuran dibayar. Pinjaman LUNAS! 🎉")
	} else {
		SetFlash(c, "success", "Angsuran berhasil dibayar.")
	}
	c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
}
