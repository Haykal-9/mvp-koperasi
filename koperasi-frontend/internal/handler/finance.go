package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"koperasi-frontend/internal/mock"
	"koperasi-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

type FinanceHandler struct {
	Render Renderer
}

func NewFinanceHandler(render Renderer) *FinanceHandler {
	return &FinanceHandler{Render: render}
}

// ===== GET /finance/journals =====
func (h *FinanceHandler) Journals(c *gin.Context) {
	tipeFilter := strings.ToUpper(c.Query("tipe"))
	out := []model.JournalEntry{}
	for i := len(mock.JournalEntries) - 1; i >= 0; i-- {
		j := mock.JournalEntries[i]
		if tipeFilter != "" && tipeFilter != "ALL" && j.TipeTransaksi != tipeFilter {
			continue
		}
		out = append(out, j)
	}
	cnt := map[string]int{}
	for _, j := range mock.JournalEntries {
		cnt[j.TipeTransaksi]++
	}
	h.Render(c, "base", "finance/journals", gin.H{
		"Title":         "Jurnal Keuangan",
		"Active":        "journals",
		"Journals":      out,
		"TipeFilter":    tipeFilter,
		"CountAll":      len(mock.JournalEntries),
		"CountSimpanan": cnt["SIMPANAN"],
		"CountPOS":      cnt["POS"],
		"CountPinjaman": cnt["PINJAMAN"],
		"CountManual":   cnt["MANUAL"],
	})
}

// ===== GET /finance/journals/create =====
func (h *FinanceHandler) ShowCreateJournal(c *gin.Context) {
	h.Render(c, "base", "finance/create_journal", gin.H{
		"Title":  "Input Jurnal Manual",
		"Active": "journals",
	})
}

// ===== POST /finance/journals/create =====
func (h *FinanceHandler) DoCreateJournal(c *gin.Context) {
	tanggal := strings.TrimSpace(c.PostForm("tanggal"))
	keterangan := strings.TrimSpace(c.PostForm("keterangan"))
	akunDebit := strings.TrimSpace(c.PostForm("akun_debit"))
	akunKredit := strings.TrimSpace(c.PostForm("akun_kredit"))
	nominalStr := strings.ReplaceAll(c.PostForm("nominal"), ".", "")
	nominal := 0.0
	if nominalStr != "" {
		nominal, _ = parseFloat(nominalStr)
	}

	if tanggal == "" || keterangan == "" || akunDebit == "" || akunKredit == "" || nominal <= 0 {
		SetFlash(c, "error", "Semua field wajib diisi dan nominal harus > 0.")
		c.Redirect(http.StatusFound, "/finance/journals/create")
		return
	}

	mock.AppendJournalEntry(keterangan, akunDebit, akunKredit, nominal, "MANUAL", tanggal)
	SetFlash(c, "success", "Jurnal manual berhasil ditambahkan.")
	c.Redirect(http.StatusFound, "/finance/journals")
}

// ===== GET /finance/summary =====
func (h *FinanceHandler) Summary(c *gin.Context) {
	// total kas = sum all debit to Kas - sum all kredit from Kas
	totalKas := 0.0
	totalSimpanan := 0.0
	totalPiutang := 0.0
	totalPendapatan := 0.0

	for _, j := range mock.JournalEntries {
		if j.AkunDebit == "Kas" {
			totalKas += j.Nominal
		}
		if j.AkunKredit == "Kas" {
			totalKas -= j.Nominal
		}
		if strings.HasPrefix(j.AkunKredit, "Simpanan") {
			totalSimpanan += j.Nominal
		}
		if strings.HasPrefix(j.AkunDebit, "Simpanan") {
			totalSimpanan -= j.Nominal
		}
		if j.AkunDebit == "Piutang Anggota" {
			totalPiutang += j.Nominal
		}
		if j.AkunKredit == "Piutang Anggota" {
			totalPiutang -= j.Nominal
		}
		if j.AkunKredit == "Pendapatan Penjualan" {
			totalPendapatan += j.Nominal
		}
	}

	// monthly breakdown (last 6 months)
	type monthRow struct {
		Bulan      string
		Pemasukan  float64
		Pengeluaran float64
		Selisih    float64
	}
	months := []monthRow{}
	now := time.Now()
	for i := 5; i >= 0; i-- {
		t := now.AddDate(0, -i, 0)
		prefix := t.Format("2006-01")
		label := t.Format("Jan 2006")
		pemasukan := 0.0
		pengeluaran := 0.0
		for _, j := range mock.JournalEntries {
			if strings.HasPrefix(j.Tanggal, prefix) {
				if j.AkunDebit == "Kas" {
					pemasukan += j.Nominal
				}
				if j.AkunKredit == "Kas" {
					pengeluaran += j.Nominal
				}
			}
		}
		months = append(months, monthRow{
			Bulan:       label,
			Pemasukan:   pemasukan,
			Pengeluaran: pengeluaran,
			Selisih:     pemasukan - pengeluaran,
		})
	}

	h.Render(c, "base", "finance/summary", gin.H{
		"Title":          "Ringkasan Keuangan",
		"Active":         "finance-summary",
		"TotalKas":       totalKas,
		"TotalSimpanan":  totalSimpanan,
		"TotalPiutang":   totalPiutang,
		"TotalPendapatan": totalPendapatan,
		"Months":         months,
	})
}

func parseFloat(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", ".")
	return strconv.ParseFloat(s, 64)
}
