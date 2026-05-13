package ecommerce

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"koperasi-frontend/core/handler"
	"koperasi-frontend/core/mock"
	"koperasi-frontend/core/model"
)

// ProfileHandler handles e-commerce user profile and address management.
type ProfileHandler struct {
	Render ECRenderer
}

func NewProfileHandler(render ECRenderer) *ProfileHandler {
	return &ProfileHandler{Render: render}
}

// Profile shows the user's e-commerce profile overview.
// GET /ecommerce/profile
func (h *ProfileHandler) Profile(c *gin.Context) {
	ecUserID := GetECUserID(c)
	user := mock.FindECommerceUserByID(ecUserID)
	if user == nil {
		c.Redirect(http.StatusFound, "/ecommerce/logout")
		return
	}

	orders := mock.GetECUserOrders(ecUserID)
	points := mock.GetECUserPoints(ecUserID)
	addresses := mock.GetECUserAddresses(ecUserID)
	linkedMember := mock.GetLinkedKoperasiMember(ecUserID)

	completedOrders := 0
	for _, o := range orders {
		if o.Status == "SELESAI" {
			completedOrders++
		}
	}

	balance := 0.0
	if points != nil {
		balance = points.Balance
	}

	h.Render(c, "ec_base", "ecommerce/profile/profile", gin.H{
		"Title":           "Profil Saya",
		"Active":          "profile",
		"User":            user,
		"TotalOrders":     len(orders),
		"CompletedOrders": completedOrders,
		"PointsBalance":   balance,
		"AddressCount":    len(addresses),
		"LinkedMember":    linkedMember,
	})
}

// Addresses shows the user's saved shipping addresses.
// GET /ecommerce/profile/addresses
func (h *ProfileHandler) Addresses(c *gin.Context) {
	ecUserID := GetECUserID(c)
	addresses := mock.GetECUserAddresses(ecUserID)

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
	mock.CreateECAddress(ecUserID, addr)

	handler.SetFlash(c, "ec_success", fmt.Sprintf("Alamat \"%s\" berhasil ditambahkan.", label))
	c.Redirect(http.StatusFound, "/ecommerce/profile/addresses")
}

// SetDefaultAddress marks an address as the default.
// POST /ecommerce/profile/addresses/:id/default
func (h *ProfileHandler) SetDefaultAddress(c *gin.Context) {
	ecUserID := GetECUserID(c)
	var addrID int
	fmt.Sscanf(c.Param("id"), "%d", &addrID)

	for i := range mock.ECAddresses {
		if mock.ECAddresses[i].UserID == ecUserID {
			mock.ECAddresses[i].IsDefault = (mock.ECAddresses[i].ID == addrID)
		}
	}

	handler.SetFlash(c, "ec_success", "Alamat utama berhasil diubah.")
	c.Redirect(http.StatusFound, "/ecommerce/profile/addresses")
}

// DeleteAddress removes an address.
// POST /ecommerce/profile/addresses/:id/delete
func (h *ProfileHandler) DeleteAddress(c *gin.Context) {
	ecUserID := GetECUserID(c)
	var addrID int
	fmt.Sscanf(c.Param("id"), "%d", &addrID)

	// Verify ownership
	owned := false
	for _, a := range mock.ECAddresses {
		if a.ID == addrID && a.UserID == ecUserID {
			owned = true
			break
		}
	}
	if !owned {
		handler.SetFlash(c, "ec_error", "Alamat tidak ditemukan.")
		c.Redirect(http.StatusFound, "/ecommerce/profile/addresses")
		return
	}

	mock.DeleteECAddress(addrID)
	handler.SetFlash(c, "ec_success", "Alamat berhasil dihapus.")
	c.Redirect(http.StatusFound, "/ecommerce/profile/addresses")
}

// Settings shows account settings.
// GET /ecommerce/profile/settings
func (h *ProfileHandler) Settings(c *gin.Context) {
	ecUserID := GetECUserID(c)
	user := mock.FindECommerceUserByID(ecUserID)

	h.Render(c, "ec_base", "ecommerce/profile/settings", gin.H{
		"Title":  "Pengaturan Akun",
		"Active": "profile",
		"User":   user,
	})
}

// UpdateSettings processes account settings update (username/email mock).
// POST /ecommerce/profile/settings
func (h *ProfileHandler) UpdateSettings(c *gin.Context) {
	ecUserID := GetECUserID(c)
	user := mock.FindECommerceUserByID(ecUserID)
	if user == nil {
		c.Redirect(http.StatusFound, "/ecommerce/logout")
		return
	}

	newUsername := strings.TrimSpace(c.PostForm("username"))
	newEmail := strings.TrimSpace(c.PostForm("email"))

	if newUsername != "" && newUsername != user.Username {
		// Check uniqueness
		existing := mock.FindECommerceUserByUsername(newUsername)
		if existing != nil && existing.ID != ecUserID {
			handler.SetFlash(c, "ec_error", "Username sudah digunakan.")
			c.Redirect(http.StatusFound, "/ecommerce/profile/settings")
			return
		}
		user.Username = newUsername
	}
	if newEmail != "" {
		user.Email = newEmail
	}

	handler.SetFlash(c, "ec_success", "Pengaturan akun berhasil diperbarui.")
	c.Redirect(http.StatusFound, "/ecommerce/profile/settings")
}
