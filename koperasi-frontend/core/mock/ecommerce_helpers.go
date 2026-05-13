package mock

import (
	"fmt"
	"strings"
	"time"

	"koperasi-frontend/core/model"
)

// ============================================================
// Authentication Helpers
// ============================================================

func FindECommerceUserByUsername(username string) *model.ECommerceUser {
	for i := range ECommerceUsers {
		if ECommerceUsers[i].Username == username {
			return &ECommerceUsers[i]
		}
	}
	return nil
}

func FindECommerceUserByEmail(email string) *model.ECommerceUser {
	for i := range ECommerceUsers {
		if ECommerceUsers[i].Email == email {
			return &ECommerceUsers[i]
		}
	}
	return nil
}

func FindECommerceUserByID(id int) *model.ECommerceUser {
	for i := range ECommerceUsers {
		if ECommerceUsers[i].ID == id {
			return &ECommerceUsers[i]
		}
	}
	return nil
}

// AuthenticateECommerceUser validates credentials against koperasi Users (single source of truth).
// Email and password are the same as koperasi management login.
func AuthenticateECommerceUser(email, password string) (*model.ECommerceUser, bool) {
	// Validate against koperasi user credentials
	kUser := FindUserByEmail(email)
	if kUser == nil || kUser.Password != password {
		return nil, false
	}
	// Find the corresponding EC user
	u := FindECommerceUserByEmail(email)
	return u, u != nil
}

func CreateECommerceUser(username, email, password string) *model.ECommerceUser {
	nextID := 0
	for _, u := range ECommerceUsers {
		if u.ID > nextID {
			nextID = u.ID
		}
	}
	newUser := model.ECommerceUser{
		ID: nextID + 1, Username: username, Email: email, Password: password,
		Role: "BUYER", IsSellerActive: false, CreatedAt: time.Now().Format("2006-01-02"),
	}
	ECommerceUsers = append(ECommerceUsers, newUser)
	// Initialize points
	ECUserPoints = append(ECUserPoints, model.UserPoints{UserID: newUser.ID})
	return &ECommerceUsers[len(ECommerceUsers)-1]
}

func LinkToKoperasiMember(ecUserID, memberID int) bool {
	u := FindECommerceUserByID(ecUserID)
	if u == nil {
		return false
	}
	m := FindMemberByID(memberID)
	if m == nil {
		return false
	}
	u.LinkedKoperasiMemberID = memberID
	LogAuditAction("LINK_KOPERASI", ecUserID, u.Username, fmt.Sprintf("member:%d", memberID),
		fmt.Sprintf("Linked to koperasi member %s (%s)", m.Nama, m.NomorAnggota))
	return true
}

func SearchKoperasiMembers(query string) []model.Member {
	q := strings.ToLower(query)
	var results []model.Member
	for _, m := range Members {
		if strings.Contains(strings.ToLower(m.Nama), q) || strings.Contains(m.NomorAnggota, query) {
			results = append(results, m)
			if len(results) >= 5 {
				break
			}
		}
	}
	return results
}

// ============================================================
// Seller Helpers
// ============================================================

func GetSellerProfile(sellerID int) *model.SellerProfile {
	for i := range ECSellerProfiles {
		if ECSellerProfiles[i].SellerID == sellerID {
			return &ECSellerProfiles[i]
		}
	}
	return nil
}

func GetSellerProducts(sellerID int) []model.ECProduct {
	var out []model.ECProduct
	for _, p := range ECProducts {
		if p.SellerID == sellerID {
			out = append(out, p)
		}
	}
	return out
}

func GetSellerOrders(sellerID int) []model.ECOrder {
	var out []model.ECOrder
	for _, o := range ECOrders {
		if o.SellerID == sellerID {
			out = append(out, o)
		}
	}
	return out
}

func ActivateSellerAccount(userID int) bool {
	u := FindECommerceUserByID(userID)
	if u == nil {
		return false
	}
	u.IsSellerActive = true
	ECSellerProfiles = append(ECSellerProfiles, model.SellerProfile{
		SellerID: userID, StoreName: "Toko " + u.Username,
		Description: "Toko baru", Rating: 0, ResponseTime: "-", TotalSold: 0,
		JoinedAt: time.Now().Format("2006-01-02"),
	})
	return true
}

func UpdateSellerProfile(sellerID int, storeName, description string) bool {
	sp := GetSellerProfile(sellerID)
	if sp == nil {
		return false
	}
	if storeName != "" {
		sp.StoreName = storeName
	}
	if description != "" {
		sp.Description = description
	}
	return true
}

// ============================================================
// Product & Review Helpers
// ============================================================

func FindECProductByID(id int) *model.ECProduct {
	for i := range ECProducts {
		if ECProducts[i].ID == id {
			return &ECProducts[i]
		}
	}
	return nil
}

func GetECProductsByCategory(category string) []model.ECProduct {
	var out []model.ECProduct
	for _, p := range ECProducts {
		if p.Status == "APPROVED" && strings.EqualFold(p.Kategori, category) {
			out = append(out, p)
		}
	}
	return out
}

func GetECProductsByPriceRange(min, max float64) []model.ECProduct {
	var out []model.ECProduct
	for _, p := range ECProducts {
		if p.Status == "APPROVED" && p.Harga >= min && p.Harga <= max {
			out = append(out, p)
		}
	}
	return out
}

func SearchECProducts(query string) []model.ECProduct {
	q := strings.ToLower(query)
	var out []model.ECProduct
	for _, p := range ECProducts {
		if p.Status == "APPROVED" && (strings.Contains(strings.ToLower(p.Nama), q) ||
			strings.Contains(strings.ToLower(p.Deskripsi), q) ||
			strings.Contains(strings.ToLower(p.Kategori), q)) {
			out = append(out, p)
		}
	}
	return out
}

func GetAllApprovedECProducts() []model.ECProduct {
	var out []model.ECProduct
	for _, p := range ECProducts {
		if p.Status == "APPROVED" {
			out = append(out, p)
		}
	}
	return out
}

func GetECProductReviews(productID int) []model.ProductReview {
	var out []model.ProductReview
	for _, r := range ECProductReviews {
		if r.ProductID == productID {
			out = append(out, r)
		}
	}
	return out
}

func GetECAverageRating(productID int) float64 {
	reviews := GetECProductReviews(productID)
	if len(reviews) == 0 {
		return 0
	}
	sum := 0
	for _, r := range reviews {
		sum += r.Rating
	}
	return float64(sum) / float64(len(reviews))
}

func SubmitECReview(userID, productID, rating int, komentar string) bool {
	u := FindECommerceUserByID(userID)
	if u == nil {
		return false
	}
	nextID := 0
	for _, r := range ECProductReviews {
		if r.ID > nextID {
			nextID = r.ID
		}
	}
	ECProductReviews = append(ECProductReviews, model.ProductReview{
		ID: nextID + 1, ProductID: productID, UserID: userID,
		Username: u.Username, Rating: rating, Komentar: komentar,
		CreatedAt: time.Now().Format("2006-01-02"),
	})
	// Update product rating
	p := FindECProductByID(productID)
	if p != nil {
		p.TotalReview++
		p.Rating = GetECAverageRating(productID)
	}
	return true
}

func NextECProductID() int {
	max := 0
	for _, p := range ECProducts {
		if p.ID > max {
			max = p.ID
		}
	}
	return max + 1
}

// ============================================================
// Order & Cart Helpers
// ============================================================

func NextECOrderID() int {
	max := 0
	for _, o := range ECOrders {
		if o.ID > max {
			max = o.ID
		}
	}
	return max + 1
}

func CreateECOrder(buyerID int, items []model.ECOrderItem, alamat, shippingOpt, voucherCode, metodeBayar string, shippingCost, pointsUsed float64) *model.ECOrder {
	buyer := FindECommerceUserByID(buyerID)
	if buyer == nil {
		return nil
	}
	subtotal := 0.0
	sellerID := 0
	sellerName := ""
	for _, it := range items {
		subtotal += it.Subtotal
		sellerID = it.SellerID
	}
	if sp := GetSellerProfile(sellerID); sp != nil {
		sellerName = sp.StoreName
	}
	discount := 0.0
	if voucherCode != "" {
		d, _ := ValidateECVoucher(voucherCode, subtotal)
		discount = d
	}
	total := subtotal + shippingCost - discount - pointsUsed
	if total < 0 {
		total = 0
	}
	today := time.Now().Format("2006-01-02")
	nomorOrder := fmt.Sprintf("EC-%s-%03d", strings.ReplaceAll(today, "-", ""), NextECOrderID())

	order := model.ECOrder{
		ID: NextECOrderID(), NomorOrder: nomorOrder,
		BuyerID: buyerID, BuyerName: buyer.Username,
		SellerID: sellerID, SellerName: sellerName,
		Items: items, AlamatPengiriman: alamat,
		ShippingOption: shippingOpt, ShippingCost: shippingCost,
		Subtotal: subtotal, Discount: discount, PointsUsed: pointsUsed,
		TotalHarga: total, VoucherCode: voucherCode,
		MetodeBayar: metodeBayar, Status: "DIBAYAR",
		CreatedAt: time.Now().Format("2006-01-02 15:04"),
		UpdatedAt: time.Now().Format("2006-01-02 15:04"),
	}
	ECOrders = append(ECOrders, order)
	// Earn points (1 point per Rp 1000 spent)
	earned := float64(int(subtotal / 1000))
	if earned > 0 {
		ECOrders[len(ECOrders)-1].PointsEarned = earned
		AddECPoints(buyerID, earned, "EARN_PURCHASE", order.ID, "Pembelian order "+nomorOrder)
	}
	return &ECOrders[len(ECOrders)-1]
}

func GetECUserOrders(userID int) []model.ECOrder {
	var out []model.ECOrder
	for i := len(ECOrders) - 1; i >= 0; i-- {
		if ECOrders[i].BuyerID == userID {
			out = append(out, ECOrders[i])
		}
	}
	return out
}

func GetECSellerReceivedOrders(sellerID int) []model.ECOrder {
	var out []model.ECOrder
	for i := len(ECOrders) - 1; i >= 0; i-- {
		if ECOrders[i].SellerID == sellerID {
			out = append(out, ECOrders[i])
		}
	}
	return out
}

func FindECOrderByID(id int) *model.ECOrder {
	for i := range ECOrders {
		if ECOrders[i].ID == id {
			return &ECOrders[i]
		}
	}
	return nil
}

func UpdateECOrderStatus(orderID int, newStatus string) bool {
	o := FindECOrderByID(orderID)
	if o == nil {
		return false
	}
	o.Status = newStatus
	o.UpdatedAt = time.Now().Format("2006-01-02 15:04")
	return true
}

func AddECShipmentEvent(orderID int, status, lokasi, keterangan string) bool {
	nextID := 0
	for _, e := range ECShipmentEvents {
		if e.ID > nextID {
			nextID = e.ID
		}
	}
	ECShipmentEvents = append(ECShipmentEvents, model.ShipmentEvent{
		ID: nextID + 1, OrderID: orderID, Status: status,
		Lokasi: lokasi, Keterangan: keterangan,
		CreatedAt: time.Now().Format("2006-01-02 15:04"),
	})
	return true
}

func GetECShipmentEvents(orderID int) []model.ShipmentEvent {
	var out []model.ShipmentEvent
	for _, e := range ECShipmentEvents {
		if e.OrderID == orderID {
			out = append(out, e)
		}
	}
	return out
}

// ============================================================
// Points Helpers
// ============================================================

func GetECUserPoints(userID int) *model.UserPoints {
	for i := range ECUserPoints {
		if ECUserPoints[i].UserID == userID {
			return &ECUserPoints[i]
		}
	}
	return nil
}

func AddECPoints(userID int, amount float64, txType string, orderID int, keterangan string) bool {
	up := GetECUserPoints(userID)
	if up == nil {
		ECUserPoints = append(ECUserPoints, model.UserPoints{UserID: userID})
		up = &ECUserPoints[len(ECUserPoints)-1]
	}
	up.Balance += amount
	if amount > 0 {
		up.TotalEarned += amount
	} else {
		up.TotalRedeemed += -amount
	}
	nextID := 0
	for _, t := range ECPointsTransactions {
		if t.ID > nextID {
			nextID = t.ID
		}
	}
	ECPointsTransactions = append(ECPointsTransactions, model.PointsTransaction{
		ID: nextID + 1, UserID: userID, Tipe: txType, Amount: amount,
		OrderID: orderID, Keterangan: keterangan,
		CreatedAt: time.Now().Format("2006-01-02"),
	})
	return true
}

func RedeemECPoints(userID int, amount float64) (bool, string) {
	up := GetECUserPoints(userID)
	if up == nil {
		return false, "User tidak ditemukan"
	}
	if up.Balance < amount {
		return false, fmt.Sprintf("Poin tidak cukup (saldo: %.0f, diminta: %.0f)", up.Balance, amount)
	}
	AddECPoints(userID, -amount, "REDEEM_DISCOUNT", 0, fmt.Sprintf("Redeem %.0f poin", amount))
	return true, "Berhasil redeem poin"
}

func ConvertPointsToKoperasiSimpanan(ecUserID int, points float64) (bool, string) {
	u := FindECommerceUserByID(ecUserID)
	if u == nil {
		return false, "User tidak ditemukan"
	}
	if u.LinkedKoperasiMemberID == 0 {
		return false, "Akun belum di-link dengan member koperasi"
	}
	up := GetECUserPoints(ecUserID)
	if up == nil || up.Balance < points {
		return false, "Poin tidak cukup"
	}
	// Convert: 1 poin = Rp 100
	rupiah := points * 100
	m := FindMemberByID(u.LinkedKoperasiMemberID)
	if m == nil {
		return false, "Member koperasi tidak ditemukan"
	}
	// Mock: add to simpanan sukarela
	m.SimpananSukarela += rupiah
	AddECPoints(ecUserID, -points, "CONVERT_SIMPANAN", 0,
		fmt.Sprintf("Konversi %.0f poin → Rp %.0f ke Simpanan Sukarela %s", points, rupiah, m.Nama))
	return true, fmt.Sprintf("Berhasil konversi %.0f poin (Rp %.0f) ke Simpanan Sukarela %s", points, rupiah, m.Nama)
}

func GetECPointsTransactions(userID int) []model.PointsTransaction {
	var out []model.PointsTransaction
	for i := len(ECPointsTransactions) - 1; i >= 0; i-- {
		if ECPointsTransactions[i].UserID == userID {
			out = append(out, ECPointsTransactions[i])
		}
	}
	return out
}

// ============================================================
// Wishlist & Address Helpers
// ============================================================

func GetECUserAddresses(userID int) []model.ECAddress {
	var out []model.ECAddress
	for _, a := range ECAddresses {
		if a.UserID == userID {
			out = append(out, a)
		}
	}
	return out
}

func GetECDefaultAddress(userID int) *model.ECAddress {
	for i := range ECAddresses {
		if ECAddresses[i].UserID == userID && ECAddresses[i].IsDefault {
			return &ECAddresses[i]
		}
	}
	addrs := GetECUserAddresses(userID)
	if len(addrs) > 0 {
		for i := range ECAddresses {
			if ECAddresses[i].ID == addrs[0].ID {
				return &ECAddresses[i]
			}
		}
	}
	return nil
}

func CreateECAddress(userID int, addr model.ECAddress) *model.ECAddress {
	nextID := 0
	for _, a := range ECAddresses {
		if a.ID > nextID {
			nextID = a.ID
		}
	}
	addr.ID = nextID + 1
	addr.UserID = userID
	if len(GetECUserAddresses(userID)) == 0 {
		addr.IsDefault = true
	}
	ECAddresses = append(ECAddresses, addr)
	return &ECAddresses[len(ECAddresses)-1]
}

func UpdateECAddress(addressID int, update model.ECAddress) bool {
	for i := range ECAddresses {
		if ECAddresses[i].ID == addressID {
			if update.Label != "" { ECAddresses[i].Label = update.Label }
			if update.Penerima != "" { ECAddresses[i].Penerima = update.Penerima }
			if update.NoHP != "" { ECAddresses[i].NoHP = update.NoHP }
			if update.Alamat != "" { ECAddresses[i].Alamat = update.Alamat }
			if update.Kota != "" { ECAddresses[i].Kota = update.Kota }
			if update.Provinsi != "" { ECAddresses[i].Provinsi = update.Provinsi }
			if update.KodePos != "" { ECAddresses[i].KodePos = update.KodePos }
			return true
		}
	}
	return false
}

func DeleteECAddress(addressID int) bool {
	for i := range ECAddresses {
		if ECAddresses[i].ID == addressID {
			ECAddresses = append(ECAddresses[:i], ECAddresses[i+1:]...)
			return true
		}
	}
	return false
}

func IsECWishlisted(userID, productID int) bool {
	for _, w := range ECWishlists {
		if w.UserID == userID && w.ProductID == productID {
			return true
		}
	}
	return false
}

func ToggleECWishlist(userID, productID int) bool {
	for i := range ECWishlists {
		if ECWishlists[i].UserID == userID && ECWishlists[i].ProductID == productID {
			ECWishlists = append(ECWishlists[:i], ECWishlists[i+1:]...)
			return false // removed
		}
	}
	nextID := 0
	for _, w := range ECWishlists {
		if w.ID > nextID {
			nextID = w.ID
		}
	}
	ECWishlists = append(ECWishlists, model.Wishlist{
		ID: nextID + 1, UserID: userID, ProductID: productID,
		CreatedAt: time.Now().Format("2006-01-02"),
	})
	return true // added
}

func GetECUserWishlist(userID int) []model.ECProduct {
	var out []model.ECProduct
	for _, w := range ECWishlists {
		if w.UserID == userID {
			p := FindECProductByID(w.ProductID)
			if p != nil {
				out = append(out, *p)
			}
		}
	}
	return out
}

// ============================================================
// Voucher & Shipping Helpers
// ============================================================

func FindECVoucherByCode(code string) *model.Voucher {
	code = strings.ToUpper(code)
	for i := range ECVouchers {
		if ECVouchers[i].Code == code {
			return &ECVouchers[i]
		}
	}
	return nil
}

func ValidateECVoucher(code string, totalPrice float64) (float64, string) {
	v := FindECVoucherByCode(code)
	if v == nil {
		return 0, "Voucher tidak ditemukan"
	}
	if v.Status != "ACTIVE" {
		return 0, "Voucher sudah tidak berlaku"
	}
	if v.Kuota <= 0 {
		return 0, "Kuota voucher sudah habis"
	}
	if totalPrice < v.MinPembelian {
		return 0, fmt.Sprintf("Minimum pembelian Rp %.0f", v.MinPembelian)
	}
	discount := 0.0
	if v.TipeDiskon == "PERCENT" {
		discount = totalPrice * v.NilaiDiskon / 100
		if discount > v.MaksDiskon {
			discount = v.MaksDiskon
		}
	} else {
		discount = v.NilaiDiskon
	}
	return discount, "Voucher valid"
}

func GetECShippingOptions() []model.ShippingOption {
	return ECShippingOptions
}

func CalculateECShippingCost(shipOptionID int, weightGrams int) float64 {
	for _, s := range ECShippingOptions {
		if s.ID == shipOptionID {
			kg := float64(weightGrams) / 1000.0
			if kg < 1 {
				kg = 1
			}
			return s.Harga * kg
		}
	}
	return 0
}

// ============================================================
// Audit Helpers
// ============================================================

func LogAuditAction(action string, userID int, username, resource, details string) {
	nextID := 0
	for _, a := range ECAuditLogs {
		if a.ID > nextID {
			nextID = a.ID
		}
	}
	ECAuditLogs = append(ECAuditLogs, model.AuditLog{
		ID: nextID + 1, Action: action, UserID: userID, Username: username,
		Resource: resource, Details: details,
		CreatedAt: time.Now().Format("2006-01-02 15:04"),
	})
}

func GetECAuditLogs() []model.AuditLog {
	out := make([]model.AuditLog, len(ECAuditLogs))
	for i, j := 0, len(ECAuditLogs)-1; j >= 0; i, j = i+1, j-1 {
		out[i] = ECAuditLogs[j]
	}
	return out
}

// GetECCategories returns a list of unique product categories.
func GetECCategories() []string {
	seen := map[string]bool{}
	var cats []string
	for _, p := range ECProducts {
		if !seen[p.Kategori] {
			seen[p.Kategori] = true
			cats = append(cats, p.Kategori)
		}
	}
	return cats
}

// ============================================================
// Admin Helpers
// ============================================================

// HasECReviewed checks whether a user has already reviewed a specific product.
func HasECReviewed(userID, productID int) bool {
	for _, r := range ECProductReviews {
		if r.UserID == userID && r.ProductID == productID {
			return true
		}
	}
	return false
}

// GetPendingECProducts returns all products with PENDING_APPROVAL status.
func GetPendingECProducts() []model.ECProduct {
	var out []model.ECProduct
	for _, p := range ECProducts {
		if p.Status == "PENDING_APPROVAL" {
			out = append(out, p)
		}
	}
	return out
}

// GetAllECOrders returns all orders, newest first.
func GetAllECOrders() []model.ECOrder {
	out := make([]model.ECOrder, len(ECOrders))
	for i, j := 0, len(ECOrders)-1; j >= 0; i, j = i+1, j-1 {
		out[i] = ECOrders[j]
	}
	return out
}

// GetLinkedKoperasiMember returns the koperasi Member linked to this EC user, or nil.
func GetLinkedKoperasiMember(ecUserID int) *model.Member {
	u := FindECommerceUserByID(ecUserID)
	if u == nil || u.LinkedKoperasiMemberID == 0 {
		return nil
	}
	return FindMemberByID(u.LinkedKoperasiMemberID)
}

// FindECommerceUserByMemberName finds the EC user whose linked koperasi member has the given name.
// Used for SSO: auto-login to e-commerce when already logged into koperasi management.
func FindECommerceUserByMemberName(nama string) *model.ECommerceUser {
	for i := range Members {
		if Members[i].Nama == nama {
			memberID := Members[i].ID
			for j := range ECommerceUsers {
				if ECommerceUsers[j].LinkedKoperasiMemberID == memberID {
					return &ECommerceUsers[j]
				}
			}
		}
	}
	return nil
}

