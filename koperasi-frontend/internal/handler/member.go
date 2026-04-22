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

type MemberHandler struct {
	Render Renderer
}

func NewMemberHandler(render Renderer) *MemberHandler {
	return &MemberHandler{Render: render}
}

// ===== GET /members =====
func (h *MemberHandler) List(c *gin.Context) {
	statusFilter := strings.ToUpper(c.Query("status"))

	members := []model.Member{}
	for _, m := range mock.Members {
		if statusFilter != "" && statusFilter != "ALL" && m.Status != statusFilter {
			continue
		}
		members = append(members, m)
	}

	// counts for tab badges
	var aktif, pending, nonAktif int
	for _, m := range mock.Members {
		switch m.Status {
		case "AKTIF":
			aktif++
		case "PENDING":
			pending++
		case "NON_AKTIF":
			nonAktif++
		}
	}

	h.Render(c, "base", "member/list", gin.H{
		"Title":         "Daftar Anggota",
		"Active":        "members",
		"Members":       members,
		"StatusFilter":  statusFilter,
		"CountAktif":    aktif,
		"CountPending":  pending,
		"CountNonAktif": nonAktif,
		"CountAll":      len(mock.Members),
	})
}

// ===== GET /members/:id =====
func (h *MemberHandler) Detail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	m := mock.FindMemberByID(id)
	if m == nil {
		c.String(http.StatusNotFound, "Anggota tidak ditemukan")
		return
	}
	h.Render(c, "base", "member/detail", gin.H{
		"Title":            "Detail Anggota",
		"Active":           "members",
		"Member":           m,
		"TotalSimpanan":    m.SimpananPokok + m.SimpananWajib + m.SimpananSukarela,
		"SimpananHistory":  mock.SimpananByMember(m.Nama),
		"PinjamanHistory":  mock.LoansByMember(m.Nama),
	})
}

// ===== GET /members/register =====
func (h *MemberHandler) ShowRegister(c *gin.Context) {
	prefillNama, _ := PopFlash(c, "form_nama")
	prefillNIK, _ := PopFlash(c, "form_nik")
	prefillAlamat, _ := PopFlash(c, "form_alamat")
	prefillNoHP, _ := PopFlash(c, "form_nohp")

	h.Render(c, "base", "member/register_member", gin.H{
		"Title":      "Pendaftaran Anggota",
		"Active":     "member-register",
		"FormNama":   prefillNama,
		"FormNIK":    prefillNIK,
		"FormAlamat": prefillAlamat,
		"FormNoHP":   prefillNoHP,
	})
}

// ===== POST /members/register =====
func (h *MemberHandler) DoRegister(c *gin.Context) {
	nama := strings.TrimSpace(c.PostForm("nama"))
	nik := strings.TrimSpace(c.PostForm("nik"))
	alamat := strings.TrimSpace(c.PostForm("alamat"))
	noHP := strings.TrimSpace(c.PostForm("no_hp"))

	preserve := func() {
		SetFlash(c, "form_nama", nama)
		SetFlash(c, "form_nik", nik)
		SetFlash(c, "form_alamat", alamat)
		SetFlash(c, "form_nohp", noHP)
	}

	if nama == "" || nik == "" || alamat == "" || noHP == "" {
		SetFlash(c, "error", "Semua field wajib diisi.")
		preserve()
		c.Redirect(http.StatusFound, "/members/register")
		return
	}
	if len(nik) != 16 {
		SetFlash(c, "error", "NIK harus 16 digit.")
		preserve()
		c.Redirect(http.StatusFound, "/members/register")
		return
	}

	mock.Members = append(mock.Members, model.Member{
		ID:           mock.NextMemberID(),
		NomorAnggota: "-",
		Nama:         nama,
		NIK:          nik,
		Alamat:       alamat,
		NoHP:         noHP,
		Status:       "PENDING",
		TanggalMasuk: time.Now().Format("2006-01-02"),
	})

	SetFlash(c, "success", "Pendaftaran terkirim. Menunggu review pengurus.")
	c.Redirect(http.StatusFound, "/members?status=PENDING")
}

// ===== POST /members/:id/approve =====
func (h *MemberHandler) Approve(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	m := mock.FindMemberByID(id)
	if m == nil {
		c.String(http.StatusNotFound, "Anggota tidak ditemukan")
		return
	}
	if m.Status != "PENDING" {
		SetFlash(c, "error", "Anggota tidak dalam status PENDING.")
		c.Redirect(http.StatusFound, "/members")
		return
	}
	m.Status = "AKTIF"
	m.NomorAnggota = mock.GenerateNomorAnggota()
	m.TanggalMasuk = time.Now().Format("2006-01-02")

	SetFlash(c, "success", "Anggota "+m.Nama+" disetujui (Nomor "+m.NomorAnggota+").")
	c.Redirect(http.StatusFound, "/members")
}

// ===== POST /members/:id/reject =====
func (h *MemberHandler) Reject(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	m := mock.FindMemberByID(id)
	if m == nil {
		c.String(http.StatusNotFound, "Anggota tidak ditemukan")
		return
	}
	if m.Status != "PENDING" {
		SetFlash(c, "error", "Anggota tidak dalam status PENDING.")
		c.Redirect(http.StatusFound, "/members")
		return
	}
	m.Status = "NON_AKTIF"

	SetFlash(c, "success", "Pendaftaran "+m.Nama+" ditolak.")
	c.Redirect(http.StatusFound, "/members")
}

// ===== GET /members/:id/simpanan =====
func (h *MemberHandler) ShowSimpanan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	m := mock.FindMemberByID(id)
	if m == nil {
		c.String(http.StatusNotFound, "Anggota tidak ditemukan")
		return
	}
	h.Render(c, "base", "member/simpanan", gin.H{
		"Title":           "Simpanan — " + m.Nama,
		"Active":          "members",
		"Member":          m,
		"TotalSimpanan":   m.SimpananPokok + m.SimpananWajib + m.SimpananSukarela,
		"SimpananHistory": mock.SimpananByMember(m.Nama),
	})
}

// ===== POST /members/:id/simpanan =====
func (h *MemberHandler) DoSimpanan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	m := mock.FindMemberByID(id)
	if m == nil {
		c.String(http.StatusNotFound, "Anggota tidak ditemukan")
		return
	}
	jenis := strings.ToUpper(c.PostForm("jenis"))
	tipe := strings.ToUpper(c.PostForm("tipe"))
	nominal, _ := strconv.ParseFloat(c.PostForm("nominal"), 64)
	keterangan := strings.TrimSpace(c.PostForm("keterangan"))

	if jenis == "" || tipe == "" || nominal <= 0 {
		SetFlash(c, "error", "Jenis, tipe, dan nominal wajib diisi (nominal > 0).")
		c.Redirect(http.StatusFound, "/members/"+strconv.Itoa(id)+"/simpanan")
		return
	}
	if tipe == "KELUAR" {
		var saldo float64
		switch jenis {
		case "POKOK":
			saldo = m.SimpananPokok
		case "WAJIB":
			saldo = m.SimpananWajib
		case "SUKARELA":
			saldo = m.SimpananSukarela
		}
		if nominal > saldo {
			SetFlash(c, "error", "Saldo simpanan "+jenis+" tidak mencukupi.")
			c.Redirect(http.StatusFound, "/members/"+strconv.Itoa(id)+"/simpanan")
			return
		}
	}

	if err := mock.AppendSimpanan(id, jenis, tipe, nominal, keterangan, time.Now().Format("2006-01-02")); err != nil {
		SetFlash(c, "error", err.Error())
		c.Redirect(http.StatusFound, "/members/"+strconv.Itoa(id)+"/simpanan")
		return
	}

	SetFlash(c, "success", "Transaksi simpanan dicatat.")
	c.Redirect(http.StatusFound, "/members/"+strconv.Itoa(id)+"/simpanan")
}

// ===== GET /members/:id/resign =====
func (h *MemberHandler) ShowResign(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	m := mock.FindMemberByID(id)
	if m == nil {
		c.String(http.StatusNotFound, "Anggota tidak ditemukan")
		return
	}
	// outstanding obligations: aktif loans
	var sisaPinjaman float64
	for _, l := range mock.LoansByMember(m.Nama) {
		if l.Status == "AKTIF" {
			sisaPinjaman += l.SisaPokok
		}
	}
	h.Render(c, "base", "member/resign", gin.H{
		"Title":         "Pengunduran Diri — " + m.Nama,
		"Active":        "members",
		"Member":        m,
		"TotalSimpanan": m.SimpananPokok + m.SimpananWajib + m.SimpananSukarela,
		"SisaPinjaman":  sisaPinjaman,
		"BisaResign":    sisaPinjaman == 0,
	})
}

// ===== POST /members/:id/resign =====
func (h *MemberHandler) DoResign(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	m := mock.FindMemberByID(id)
	if m == nil {
		c.String(http.StatusNotFound, "Anggota tidak ditemukan")
		return
	}
	var sisaPinjaman float64
	for _, l := range mock.LoansByMember(m.Nama) {
		if l.Status == "AKTIF" {
			sisaPinjaman += l.SisaPokok
		}
	}
	if sisaPinjaman > 0 {
		SetFlash(c, "error", "Tidak bisa resign — masih ada pinjaman aktif.")
		c.Redirect(http.StatusFound, "/members/"+strconv.Itoa(id)+"/resign")
		return
	}
	m.Status = "NON_AKTIF"
	SetFlash(c, "success", "Pengunduran diri "+m.Nama+" diproses.")
	c.Redirect(http.StatusFound, "/members")
}
