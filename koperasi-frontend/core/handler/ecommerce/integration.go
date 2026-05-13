package ecommerce

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/mock"
)

// IntegrationHandler provides mock JSON API endpoints that simulate
// a separate koperasi backend service for member lookup, linking, and simpanan.
type IntegrationHandler struct{}

func NewIntegrationHandler() *IntegrationHandler {
	return &IntegrationHandler{}
}

// GetKoperasiMember returns koperasi member data as JSON.
// GET /ecommerce/api/koperasi/member/:id
func (h *IntegrationHandler) GetKoperasiMember(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	m := mock.FindMemberByID(id)
	if m == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Member tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                m.ID,
		"nomor_anggota":     m.NomorAnggota,
		"nama":              m.Nama,
		"status":            m.Status,
		"simpanan_pokok":    m.SimpananPokok,
		"simpanan_wajib":    m.SimpananWajib,
		"simpanan_sukarela": m.SimpananSukarela,
	})
}

// LinkToKoperasi links the current EC user to a koperasi member.
// POST /ecommerce/api/koperasi/link
// Form: member_id int
func (h *IntegrationHandler) LinkToKoperasi(c *gin.Context) {
	ecUserID := GetECUserID(c)
	if ecUserID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	memberID, err := strconv.Atoi(c.PostForm("member_id"))
	if err != nil || memberID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "member_id tidak valid"})
		return
	}

	for _, u := range mock.ECommerceUsers {
		if u.LinkedKoperasiMemberID == memberID && u.ID != ecUserID {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Member ini sudah di-link ke akun lain",
			})
			return
		}
	}

	if !mock.LinkToKoperasiMember(ecUserID, memberID) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Gagal link member"})
		return
	}

	m := mock.FindMemberByID(memberID)
	nama := "member"
	if m != nil {
		nama = m.Nama
	}
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Berhasil link dengan " + nama,
		"member_id":   memberID,
		"member_nama": nama,
	})
}

// AddToSimpanan converts EC points to koperasi simpanan sukarela.
// POST /ecommerce/api/koperasi/simpanan/add
// Form: points float
func (h *IntegrationHandler) AddToSimpanan(c *gin.Context) {
	ecUserID := GetECUserID(c)
	if ecUserID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	points, err := strconv.ParseFloat(c.PostForm("points"), 64)
	if err != nil || points <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Jumlah poin tidak valid"})
		return
	}

	ok, msg := mock.ConvertPointsToKoperasiSimpanan(ecUserID, points)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": msg})
		return
	}

	resp := gin.H{
		"success":      true,
		"message":      msg,
		"points_used":  points,
		"rupiah_added": points * 100,
	}
	u := mock.FindECommerceUserByID(ecUserID)
	if u != nil && u.LinkedKoperasiMemberID != 0 {
		if m := mock.FindMemberByID(u.LinkedKoperasiMemberID); m != nil {
			resp["new_simpanan_sukarela"] = m.SimpananSukarela
		}
	}
	c.JSON(http.StatusOK, resp)
}
