package ecommerce

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/handler"
	"koperasi-frontend/core/model"
	"koperasi-frontend/core/service"
)

// ProfileHandler handles e-commerce user profile and address management.
type ProfileHandler struct {
	Render ECRenderer
	Svc    *service.ECAccountService
}

func NewProfileHandler(render ECRenderer, svc *service.ECAccountService) *ProfileHandler {
	return &ProfileHandler{Render: render, Svc: svc}
}

func (h *ProfileHandler) ready(c *gin.Context) bool {
	if h.Svc == nil {
		c.String(http.StatusServiceUnavailable, "E-Commerce sementara tidak tersedia (database tidak terhubung).")
		return false
	}
	return true
}

// Profile shows the user's e-commerce profile overview.
// GET /ecommerce/profile
func (h *ProfileHandler) Profile(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)
	o, err := h.Svc.ProfileOverview(c.Request.Context(), ecUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}
	if o == nil {
		c.Redirect(http.StatusFound, "/ecommerce/logout")
		return
	}

	h.Render(c, "ec_base", "ecommerce/profile/profile", gin.H{
		"Title":           "Profil Saya",
		"Active":          "profile",
		"User":            o.User,
		"TotalOrders":     o.TotalOrders,
		"CompletedOrders": o.CompletedOrders,
		"PointsBalance":   o.PointsBalance,
		"AddressCount":    o.AddressCount,
		"LinkedMember":    o.LinkedMember,
	})
}

// Addresses shows the user's saved shipping addresses.
// GET /ecommerce/profile/addresses
func (h *ProfileHandler) Addresses(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)
	addresses, err := h.Svc.Addresses(c.Request.Context(), ecUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/profile/addresses", gin.H{
		"Title":     "Alamat Pengiriman",
		"Active":    "addresses",
		"Addresses": addresses,
	})
}

// ShowCreateAddress shows the form to add a new address.
// GET /ecommerce/profile/addresses/new
func (h *ProfileHandler) ShowCreateAddress(c *gin.Context) {
	h.Render(c, "ec_base", "ecommerce/profile/address_form", gin.H{
		"Title":  "Tambah Alamat",
		"Active": "addresses",
		"IsEdit": false,
	})
}

// CreateAddress saves a new shipping address.
// POST /ecommerce/profile/addresses
func (h *ProfileHandler) CreateAddress(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)

	label := strings.TrimSpace(c.PostForm("label"))
	penerima := strings.TrimSpace(c.PostForm("penerima"))
	noHP := strings.TrimSpace(c.PostForm("no_hp"))
	alamat := strings.TrimSpace(c.PostForm("alamat"))
	kota := strings.TrimSpace(c.PostForm("kota"))
	provinsi := strings.TrimSpace(c.PostForm("provinsi"))
	kodePos := strings.TrimSpace(c.PostForm("kode_pos"))

	if label == "" || penerima == "" || alamat == "" || kota == "" {
		handler.SetFlash(c, "ec_error", "Label, penerima, alamat, dan kota wajib diisi.")
		c.Redirect(http.StatusFound, "/ecommerce/profile/addresses/new")
		return
	}

	addr := model.ECAddress{
		Label:    label,
		Penerima: penerima,
		NoHP:     noHP,
		Alamat:   alamat,
		Kota:     kota,
		Provinsi: provinsi,
		KodePos:  kodePos,
	}
	if err := h.Svc.CreateAddress(c.Request.Context(), ecUserID, addr); err != nil {
		handler.SetFlash(c, "ec_error", "Gagal menambahkan alamat.")
		c.Redirect(http.StatusFound, "/ecommerce/profile/addresses/new")
		return
	}

	handler.SetFlash(c, "ec_success", fmt.Sprintf("Alamat \"%s\" berhasil ditambahkan.", label))
	c.Redirect(http.StatusFound, "/ecommerce/profile/addresses")
}

// SetDefaultAddress marks an address as the default.
// POST /ecommerce/profile/addresses/:id/default
func (h *ProfileHandler) SetDefaultAddress(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)
	addrID, _ := strconv.Atoi(c.Param("id"))

	if err := h.Svc.SetDefaultAddress(c.Request.Context(), ecUserID, addrID); err != nil {
		handler.SetFlash(c, "ec_error", "Gagal mengubah alamat utama.")
		c.Redirect(http.StatusFound, "/ecommerce/profile/addresses")
		return
	}

	handler.SetFlash(c, "ec_success", "Alamat utama berhasil diubah.")
	c.Redirect(http.StatusFound, "/ecommerce/profile/addresses")
}

// DeleteAddress removes an address.
// POST /ecommerce/profile/addresses/:id/delete
func (h *ProfileHandler) DeleteAddress(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)
	addrID, _ := strconv.Atoi(c.Param("id"))

	if err := h.Svc.DeleteAddress(c.Request.Context(), ecUserID, addrID); err != nil {
		handler.SetFlash(c, "ec_error", "Gagal menghapus alamat.")
		c.Redirect(http.StatusFound, "/ecommerce/profile/addresses")
		return
	}
	handler.SetFlash(c, "ec_success", "Alamat berhasil dihapus.")
	c.Redirect(http.StatusFound, "/ecommerce/profile/addresses")
}

// Settings shows account settings.
// GET /ecommerce/profile/settings
func (h *ProfileHandler) Settings(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)
	user, err := h.Svc.UserByID(c.Request.Context(), ecUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Kesalahan database")
		return
	}

	h.Render(c, "ec_base", "ecommerce/profile/settings", gin.H{
		"Title":  "Pengaturan Akun",
		"Active": "profile",
		"User":   user,
	})
}

// UpdateSettings processes account settings update (username/email).
// POST /ecommerce/profile/settings
func (h *ProfileHandler) UpdateSettings(c *gin.Context) {
	if !h.ready(c) {
		return
	}
	ecUserID := GetECUserID(c)

	newUsername := strings.TrimSpace(c.PostForm("username"))
	newEmail := strings.TrimSpace(c.PostForm("email"))

	if err := h.Svc.UpdateSettings(c.Request.Context(), ecUserID, newUsername, newEmail); err != nil {
		handler.SetFlash(c, "ec_error", err.Error())
		c.Redirect(http.StatusFound, "/ecommerce/profile/settings")
		return
	}

	handler.SetFlash(c, "ec_success", "Pengaturan akun berhasil diperbarui.")
	c.Redirect(http.StatusFound, "/ecommerce/profile/settings")
}
