package handler

import (
	"net/http"
	"strconv"
	"strings"

	"koperasi-frontend/core/service"

	"github.com/gin-gonic/gin"
)

type MemberHandler struct {
	Render Renderer
	Svc    *service.MemberService
}

func NewMemberHandler(render Renderer, svc *service.MemberService) *MemberHandler {
	return &MemberHandler{Render: render, Svc: svc}
}

// svcReady memastikan service (DB) tersedia. Jika tidak, beri pesan & redirect.
func (h *MemberHandler) svcReady(c *gin.Context) bool {
	if h.Svc == nil {
		SetFlash(c, "error", "Fitur keanggotaan memerlukan database (DATABASE_URL belum dikonfigurasi).")
		c.Redirect(http.StatusFound, "/dashboard")
		return false
	}
	return true
}

// ===== GET /members =====
func (h *MemberHandler) List(c *gin.Context) {
	if !h.svcReady(c) {
		return
	}
	statusFilter := strings.ToUpper(c.Query("status"))

	members, counts, err := h.Svc.List(c.Request.Context(), statusFilter)
	if err != nil {
		c.String(http.StatusInternalServerError, "Gagal memuat anggota: %v", err)
		return
	}

	h.Render(c, "base", "member/list", gin.H{
		"Title":         "Daftar Anggota",
		"Active":        "members",
		"Members":       members,
		"StatusFilter":  statusFilter,
		"CountAktif":    counts.Aktif,
		"CountPending":  counts.Pending,
		"CountNonAktif": counts.NonAktif,
		"CountAll":      counts.All,
	})
}

// ===== GET /members/:id =====
// ANGGOTA hanya boleh melihat profilnya sendiri.
func (h *MemberHandler) Detail(c *gin.Context) {
	if !h.svcReady(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	m, err := h.Svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Gagal memuat anggota: %v", err)
		return
	}
	if m == nil {
		c.String(http.StatusNotFound, "Anggota tidak ditemukan")
		return
	}
	if !IsOwnerOrKasir(c) && m.Nama != CurrentUserNama(c) {
		SetFlash(c, "error", "Anda hanya bisa melihat profil sendiri.")
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}
	simpanan, _ := h.Svc.SimpananHistory(c.Request.Context(), m.ID)
	loans, _ := h.Svc.LoansByMember(c.Request.Context(), m.ID)
	h.Render(c, "base", "member/detail", gin.H{
		"Title":           "Detail Anggota",
		"Active":          "members",
		"Member":          m,
		"TotalSimpanan":   m.SimpananPokok + m.SimpananWajib + m.SimpananSukarela,
		"SimpananHistory": simpanan,
		"PinjamanHistory": loans,
		"CanReview":       CurrentUserRole(c) == "OWNER",
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
	if !h.svcReady(c) {
		return
	}
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

	if _, err := h.Svc.Register(c.Request.Context(), nama, nik, alamat, noHP); err != nil {
		SetFlash(c, "error", "Gagal menyimpan pendaftaran: "+err.Error())
		preserve()
		c.Redirect(http.StatusFound, "/members/register")
		return
	}

	SetFlash(c, "success", "Pendaftaran terkirim. Menunggu review pengurus.")
	c.Redirect(http.StatusFound, "/members?status=PENDING")
}

// ===== POST /members/:id/approve =====
func (h *MemberHandler) Approve(c *gin.Context) {
	if !h.svcReady(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	m, err := h.Svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Gagal memuat anggota: %v", err)
		return
	}
	if m == nil {
		c.String(http.StatusNotFound, "Anggota tidak ditemukan")
		return
	}
	if m.Status != "PENDING" {
		SetFlash(c, "error", "Anggota tidak dalam status PENDING.")
		c.Redirect(http.StatusFound, "/members")
		return
	}
	nomor, err := h.Svc.Approve(c.Request.Context(), id)
	if err != nil {
		SetFlash(c, "error", "Gagal menyetujui anggota: "+err.Error())
		c.Redirect(http.StatusFound, "/members")
		return
	}

	SetFlash(c, "success", "Anggota "+m.Nama+" disetujui (Nomor "+nomor+").")
	c.Redirect(http.StatusFound, "/members")
}

// ===== POST /members/:id/reject =====
func (h *MemberHandler) Reject(c *gin.Context) {
	if !h.svcReady(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	m, err := h.Svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Gagal memuat anggota: %v", err)
		return
	}
	if m == nil {
		c.String(http.StatusNotFound, "Anggota tidak ditemukan")
		return
	}
	if m.Status != "PENDING" {
		SetFlash(c, "error", "Anggota tidak dalam status PENDING.")
		c.Redirect(http.StatusFound, "/members")
		return
	}
	if err := h.Svc.Reject(c.Request.Context(), id); err != nil {
		SetFlash(c, "error", "Gagal menolak pendaftaran: "+err.Error())
		c.Redirect(http.StatusFound, "/members")
		return
	}

	SetFlash(c, "success", "Pendaftaran "+m.Nama+" ditolak.")
	c.Redirect(http.StatusFound, "/members")
}

// ===== GET /members/:id/simpanan =====
// ANGGOTA hanya boleh melihat simpanan sendiri.
func (h *MemberHandler) ShowSimpanan(c *gin.Context) {
	if !h.svcReady(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	m, err := h.Svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Gagal memuat anggota: %v", err)
		return
	}
	if m == nil {
		c.String(http.StatusNotFound, "Anggota tidak ditemukan")
		return
	}
	if !IsOwnerOrKasir(c) && m.Nama != CurrentUserNama(c) {
		SetFlash(c, "error", "Anda hanya bisa melihat simpanan sendiri.")
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}
	simpanan, _ := h.Svc.SimpananHistory(c.Request.Context(), m.ID)
	h.Render(c, "base", "member/simpanan", gin.H{
		"Title":           "Simpanan — " + m.Nama,
		"Active":          "members",
		"Member":          m,
		"TotalSimpanan":   m.SimpananPokok + m.SimpananWajib + m.SimpananSukarela,
		"SimpananHistory": simpanan,
		"CanRecord":       IsOwnerOrKasir(c),
	})
}

// ===== POST /members/:id/simpanan =====
// Pencatatan transaksi simpanan hanya boleh OWNER/KASIR.
func (h *MemberHandler) DoSimpanan(c *gin.Context) {
	if !h.svcReady(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	m, err := h.Svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Gagal memuat anggota: %v", err)
		return
	}
	if m == nil {
		c.String(http.StatusNotFound, "Anggota tidak ditemukan")
		return
	}
	if !IsOwnerOrKasir(c) {
		SetFlash(c, "error", "Hanya pengurus/kasir yang bisa mencatat transaksi simpanan.")
		c.Redirect(http.StatusFound, "/members/"+strconv.Itoa(id)+"/simpanan")
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

	if err := h.Svc.RecordSimpanan(c.Request.Context(), id, jenis, tipe, nominal, keterangan); err != nil {
		SetFlash(c, "error", err.Error())
		c.Redirect(http.StatusFound, "/members/"+strconv.Itoa(id)+"/simpanan")
		return
	}

	SetFlash(c, "success", "Transaksi simpanan dicatat.")
	c.Redirect(http.StatusFound, "/members/"+strconv.Itoa(id)+"/simpanan")
}

// ===== GET /members/:id/resign =====
// Pengajuan resign — pemilik akun atau OWNER; KASIR tidak boleh approve.
func (h *MemberHandler) ShowResign(c *gin.Context) {
	if !h.svcReady(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	m, err := h.Svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Gagal memuat anggota: %v", err)
		return
	}
	if m == nil {
		c.String(http.StatusNotFound, "Anggota tidak ditemukan")
		return
	}
	role := CurrentUserRole(c)
	if role != "OWNER" && m.Nama != CurrentUserNama(c) {
		SetFlash(c, "error", "Anda hanya bisa mengajukan pengunduran diri sendiri.")
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}
	// outstanding obligations: aktif loans
	var sisaPinjaman float64
	loans, _ := h.Svc.LoansByMember(c.Request.Context(), m.ID)
	for _, l := range loans {
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
// Hanya pemilik akun yang mengajukan, atau OWNER yang memproses.
func (h *MemberHandler) DoResign(c *gin.Context) {
	if !h.svcReady(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	m, err := h.Svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Gagal memuat anggota: %v", err)
		return
	}
	if m == nil {
		c.String(http.StatusNotFound, "Anggota tidak ditemukan")
		return
	}
	role := CurrentUserRole(c)
	if role != "OWNER" && m.Nama != CurrentUserNama(c) {
		SetFlash(c, "error", "Anda hanya bisa mengajukan pengunduran diri sendiri.")
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}
	var sisaPinjaman float64
	loans, _ := h.Svc.LoansByMember(c.Request.Context(), m.ID)
	for _, l := range loans {
		if l.Status == "AKTIF" {
			sisaPinjaman += l.SisaPokok
		}
	}
	if sisaPinjaman > 0 {
		SetFlash(c, "error", "Tidak bisa resign — masih ada pinjaman aktif.")
		c.Redirect(http.StatusFound, "/members/"+strconv.Itoa(id)+"/resign")
		return
	}
	if err := h.Svc.Resign(c.Request.Context(), id); err != nil {
		SetFlash(c, "error", "Gagal memproses pengunduran diri: "+err.Error())
		c.Redirect(http.StatusFound, "/members/"+strconv.Itoa(id)+"/resign")
		return
	}
	SetFlash(c, "success", "Pengunduran diri "+m.Nama+" diproses.")
	c.Redirect(http.StatusFound, "/members")
}
