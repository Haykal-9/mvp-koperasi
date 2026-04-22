package handler

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"koperasi-frontend/internal/mock"
	"koperasi-frontend/internal/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoanHandler struct {
	Render Renderer
}

func NewLoanHandler(render Renderer) *LoanHandler {
	return &LoanHandler{Render: render}
}

// ===== GET /loans/apply =====
func (h *LoanHandler) ShowApply(c *gin.Context) {
	sess := sessions.Default(c)
	userNama, _ := sess.Get("user_nama").(string)

	// credit scoring info — find member
	var member *model.Member
	for i := range mock.Members {
		if mock.Members[i].Nama == userNama {
			member = &mock.Members[i]
			break
		}
	}
	totalSimpanan := 0.0
	tunggakanAktif := 0
	if member != nil {
		totalSimpanan = member.SimpananPokok + member.SimpananWajib + member.SimpananSukarela
	}
	for _, l := range mock.Loans {
		if l.MemberNama == userNama && l.Status == "AKTIF" {
			tunggakanAktif++
		}
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
	sess := sessions.Default(c)
	userNama, _ := sess.Get("user_nama").(string)
	if userNama == "" {
		userNama = "Walk-in"
	}

	nominalStr := strings.ReplaceAll(c.PostForm("nominal"), ".", "")
	nominal, _ := strconv.ParseFloat(nominalStr, 64)
	tenor, _ := strconv.Atoi(c.PostForm("tenor"))
	tujuan := strings.TrimSpace(c.PostForm("tujuan"))

	if nominal < 500000 {
		SetFlash(c, "error", "Nominal pinjaman minimal Rp 500.000.")
		c.Redirect(http.StatusFound, "/loans/apply")
		return
	}
	if tenor < 1 || tenor > 24 {
		SetFlash(c, "error", "Tenor harus 1-24 bulan.")
		c.Redirect(http.StatusFound, "/loans/apply")
		return
	}
	if tujuan == "" {
		SetFlash(c, "error", "Tujuan pinjaman wajib diisi.")
		c.Redirect(http.StatusFound, "/loans/apply")
		return
	}

	loan := model.Loan{
		MemberNama: userNama,
		Nominal:    nominal,
		TenorBulan: tenor,
		BungaPersen: 1.5,
		Tujuan:     tujuan,
		SisaPokok:  nominal,
		Status:     "PENDING",
	}
	saved := mock.AppendLoan(loan)
	SetFlash(c, "success", fmt.Sprintf("Pengajuan pinjaman berhasil (ID: %d). Menunggu persetujuan.", saved.ID))
	c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(saved.ID))
}

// ===== GET /loans =====
func (h *LoanHandler) List(c *gin.Context) {
	statusFilter := strings.ToUpper(c.Query("status"))
	out := []model.Loan{}
	for i := len(mock.Loans) - 1; i >= 0; i-- {
		l := mock.Loans[i]
		if statusFilter != "" && statusFilter != "ALL" && l.Status != statusFilter {
			continue
		}
		out = append(out, l)
	}
	cnt := map[string]int{}
	for _, l := range mock.Loans {
		cnt[l.Status]++
	}
	h.Render(c, "base", "loan/list", gin.H{
		"Title":          "Daftar Pinjaman",
		"Active":         "loans",
		"Loans":          out,
		"StatusFilter":   statusFilter,
		"CountAll":       len(mock.Loans),
		"CountPending":   cnt["PENDING"],
		"CountDisetujui": cnt["DISETUJUI"],
		"CountAktif":     cnt["AKTIF"],
		"CountLunas":     cnt["LUNAS"],
		"CountDitolak":   cnt["DITOLAK"],
	})
}

// ===== GET /loans/:id =====
func (h *LoanHandler) Detail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	l := mock.FindLoanByID(id)
	if l == nil {
		c.String(http.StatusNotFound, "Pinjaman tidak ditemukan")
		return
	}
	progressPct := 0.0
	if l.Nominal > 0 {
		progressPct = math.Round((1 - l.SisaPokok/l.Nominal) * 100)
	}
	// find next unpaid installment
	var nextInst *model.Installment
	for i := range l.Installments {
		if l.Installments[i].Status == "BELUM" || l.Installments[i].Status == "OVERDUE" {
			nextInst = &l.Installments[i]
			break
		}
	}

	estAngsuran := 0.0
	if l.TenorBulan > 0 {
		pokok := l.Nominal / float64(l.TenorBulan)
		bunga := l.Nominal * l.BungaPersen / 100
		estAngsuran = pokok + bunga
	}

	h.Render(c, "base", "loan/detail", gin.H{
		"Title":       "Pinjaman #" + strconv.Itoa(l.ID),
		"Active":      "loans",
		"Loan":        l,
		"ProgressPct": progressPct,
		"NextInst":    nextInst,
		"EstAngsuran": estAngsuran,
	})
}

// ===== POST /loans/:id/approve =====
func (h *LoanHandler) Approve(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	l := mock.FindLoanByID(id)
	if l == nil {
		c.String(http.StatusNotFound, "Pinjaman tidak ditemukan")
		return
	}
	if l.Status != "PENDING" {
		SetFlash(c, "error", "Hanya pinjaman PENDING yang bisa disetujui.")
		c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
		return
	}
	l.Status = "DISETUJUI"
	SetFlash(c, "success", "Pinjaman disetujui. Silakan cairkan.")
	c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
}

// ===== POST /loans/:id/reject =====
func (h *LoanHandler) Reject(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	l := mock.FindLoanByID(id)
	if l == nil {
		c.String(http.StatusNotFound, "Pinjaman tidak ditemukan")
		return
	}
	if l.Status != "PENDING" {
		SetFlash(c, "error", "Hanya pinjaman PENDING yang bisa ditolak.")
		c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
		return
	}
	l.Status = "DITOLAK"
	SetFlash(c, "success", "Pinjaman ditolak.")
	c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
}

// ===== POST /loans/:id/disburse =====
func (h *LoanHandler) Disburse(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	l := mock.FindLoanByID(id)
	if l == nil {
		c.String(http.StatusNotFound, "Pinjaman tidak ditemukan")
		return
	}
	if l.Status != "DISETUJUI" {
		SetFlash(c, "error", "Hanya pinjaman DISETUJUI yang bisa dicairkan.")
		c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
		return
	}
	l.Status = "AKTIF"
	l.TanggalCair = time.Now().Format("2006-01-02")
	mock.GenerateInstallments(l)

	// auto journal for disbursement
	mock.AppendJournalEntry("Pencairan pinjaman "+l.MemberNama, "Piutang Anggota", "Kas", l.Nominal, "PINJAMAN", l.TanggalCair)

	SetFlash(c, "success", "Pinjaman dicairkan. Jadwal angsuran dibuat.")
	c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
}

// ===== POST /loans/:id/pay =====
func (h *LoanHandler) Pay(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	l := mock.FindLoanByID(id)
	if l == nil {
		c.String(http.StatusNotFound, "Pinjaman tidak ditemukan")
		return
	}
	if l.Status != "AKTIF" {
		SetFlash(c, "error", "Hanya pinjaman AKTIF yang bisa dibayar.")
		c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
		return
	}

	// find next unpaid
	paid := false
	for i := range l.Installments {
		inst := &l.Installments[i]
		if inst.Status == "BELUM" || inst.Status == "OVERDUE" {
			inst.Status = "DIBAYAR"
			inst.TanggalBayar = time.Now().Format("2006-01-02")
			l.SisaPokok -= inst.NominalPokok
			if l.SisaPokok < 0 {
				l.SisaPokok = 0
			}

			mock.AppendJournalEntry(
				fmt.Sprintf("Angsuran pinjaman %s bulan ke-%d", l.MemberNama, inst.BulanKe),
				"Kas", "Piutang Anggota", inst.TotalBayar, "PINJAMAN", inst.TanggalBayar,
			)

			paid = true
			break
		}
	}
	if !paid {
		SetFlash(c, "error", "Tidak ada angsuran yang perlu dibayar.")
		c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
		return
	}

	// check if all installments are paid
	allPaid := true
	for _, inst := range l.Installments {
		if inst.Status != "DIBAYAR" {
			allPaid = false
			break
		}
	}
	if allPaid {
		l.Status = "LUNAS"
		SetFlash(c, "success", "Angsuran dibayar. Pinjaman LUNAS! 🎉")
	} else {
		SetFlash(c, "success", "Angsuran berhasil dibayar.")
	}
	c.Redirect(http.StatusFound, "/loans/"+strconv.Itoa(id))
}
