package handler

import (
	"net/http"
	"strconv"
	"strings"

	"koperasi-frontend/core/service"

	"github.com/gin-gonic/gin"
)

type FinanceHandler struct {
	Render Renderer
	Svc    *service.FinanceService
}

func NewFinanceHandler(render Renderer, svc *service.FinanceService) *FinanceHandler {
	return &FinanceHandler{Render: render, Svc: svc}
}

// ready memastikan service (DB) tersedia; jika tidak, alihkan dengan pesan.
func (h *FinanceHandler) ready(c *gin.Context) bool {
	if h.Svc == nil {
		SetFlash(c, "error", "Fitur keuangan membutuhkan koneksi database.")
		c.Redirect(http.StatusFound, "/dashboard")
		return false
	}
	return true
}

// ===== GET /finance/journals =====
func (h *FinanceHandler) Journals(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	tipeFilter := strings.ToUpper(c.Query("tipe"))
	journals, err := h.Svc.Journals(c.Request.Context(), tipeFilter)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	cnt, err := h.Svc.Counts(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	h.Render(c, "base", "finance/journals", gin.H{
		"Title":         "Jurnal Keuangan",
		"Active":        "journals",
		"Journals":      journals,
		"TipeFilter":    tipeFilter,
		"CountAll":      cnt.All,
		"CountSimpanan": cnt.Simpanan,
		"CountPOS":      cnt.POS,
		"CountPinjaman": cnt.Pinjaman,
		"CountManual":   cnt.Manual,
	})
}

// ===== GET /finance/journals/create =====
func (h *FinanceHandler) ShowCreateJournal(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	h.Render(c, "base", "finance/create_journal", gin.H{
		"Title":  "Input Jurnal Manual",
		"Active": "journals",
	})
}

// ===== POST /finance/journals/create =====
func (h *FinanceHandler) DoCreateJournal(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	tanggal := strings.TrimSpace(c.PostForm("tanggal"))
	keterangan := strings.TrimSpace(c.PostForm("keterangan"))
	akunDebit := strings.TrimSpace(c.PostForm("akun_debit"))
	akunKredit := strings.TrimSpace(c.PostForm("akun_kredit"))
	nominalStr := strings.ReplaceAll(c.PostForm("nominal"), ".", "")
	nominal := 0.0
	if nominalStr != "" {
		nominal, _ = parseFloat(nominalStr)
	}

	if err := h.Svc.CreateManual(c.Request.Context(), tanggal, keterangan, akunDebit, akunKredit, nominal); err != nil {
		SetFlash(c, "error", err.Error())
		c.Redirect(http.StatusFound, "/finance/journals/create")
		return
	}
	SetFlash(c, "success", "Jurnal manual berhasil ditambahkan.")
	c.Redirect(http.StatusFound, "/finance/journals")
}

// ===== GET /finance/summary =====
func (h *FinanceHandler) Summary(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	sum, err := h.Svc.Summary(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	months, err := h.Svc.MonthlyRows(c.Request.Context(), 6)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "base", "finance/summary", gin.H{
		"Title":           "Ringkasan Keuangan",
		"Active":          "finance-summary",
		"TotalKas":        sum.TotalKas,
		"TotalSimpanan":   sum.TotalSimpanan,
		"TotalPiutang":    sum.TotalPiutang,
		"TotalPendapatan": sum.TotalPendapatan,
		"Months":          months,
	})
}

func parseFloat(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", ".")
	return strconv.ParseFloat(s, 64)
}
