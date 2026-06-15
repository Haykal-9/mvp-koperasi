package ecommerce

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/service"
)

// IntegrationHandler provides JSON API endpoints for koperasi member lookup,
// linking, and points→simpanan conversion (integrasi M1↔M5).
type IntegrationHandler struct {
	Svc *service.ECAccountService
}

func NewIntegrationHandler(svc *service.ECAccountService) *IntegrationHandler {
	return &IntegrationHandler{Svc: svc}
}

// GetKoperasiMember returns koperasi member data as JSON.
// GET /ecommerce/api/koperasi/member/:id
func (h *IntegrationHandler) GetKoperasiMember(c *gin.Context) {
	if h.Svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	m, err := h.Svc.MemberByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kesalahan database"})
		return
	}
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
	if h.Svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}
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

	m, linkErr := h.Svc.LinkMember(c.Request.Context(), ecUserID, memberID)
	if linkErr != nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": linkErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Berhasil link dengan " + m.Nama,
		"member_id":   memberID,
		"member_nama": m.Nama,
	})
}

// AddToSimpanan converts EC points to koperasi simpanan sukarela.
// POST /ecommerce/api/koperasi/simpanan/add
// Form: points float
func (h *IntegrationHandler) AddToSimpanan(c *gin.Context) {
	if h.Svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak terhubung"})
		return
	}
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

	msg, rupiah, newSukarela, convErr := h.Svc.ConvertToSimpanan(c.Request.Context(), ecUserID, points)
	if convErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": convErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":               true,
		"message":               msg,
		"points_used":           points,
		"rupiah_added":          rupiah,
		"new_simpanan_sukarela": newSukarela,
	})
}
