package model

type User struct {
	ID       int
	Email    string
	Password string
	Role     string // OWNER, KASIR, ANGGOTA
	Nama     string
}

type Member struct {
	ID               int
	NomorAnggota     string
	Nama             string
	NIK              string
	Alamat           string
	NoHP             string
	Status           string // PENDING, AKTIF, NON_AKTIF
	SimpananPokok    float64
	SimpananWajib    float64
	SimpananSukarela float64
	TanggalMasuk     string
}

type Product struct {
	ID               int
	Nama             string
	Kategori         string
	Harga            float64
	Stok             int
	BatasStokMinimum int
	Deskripsi        string
	FotoURL          string
	PenjualNama      string
	Status           string // PENDING, APPROVED, REJECTED
}

type StockChange struct {
	ID          int
	ProductID   int
	ProductNama string
	Tipe        string // RESTOCK, KELUAR_PENJUALAN, KOREKSI
	Jumlah      int    // bisa negatif untuk pengurangan
	StokSetelah int
	Keterangan  string
	CreatedAt   string
}

type OrderItem struct {
	ProductID   int
	ProductNama string
	Jumlah      int
	HargaSatuan float64
	Subtotal    float64
}

type Order struct {
	ID              int
	NomorOrder      string
	PembeliNama     string
	Items           []OrderItem
	TotalHarga      float64
	FeeKoperasi     float64
	Status          string // PENDING, DIBAYAR, DIKIRIM, SELESAI, BATAL, DISPUTED
	MetodeBayar     string
	CreatedAt       string
	KomplainAlasan  string
	KomplainTanggal string
	KomplainBukti   string
}

type Installment struct {
	BulanKe      int
	JatuhTempo   string
	NominalPokok float64
	NominalBunga float64
	TotalBayar   float64
	Status       string // BELUM, DIBAYAR, OVERDUE
	TanggalBayar string
}

type Loan struct {
	ID           int
	MemberNama   string
	Nominal      float64
	TenorBulan   int
	BungaPersen  float64
	Tujuan       string
	SisaPokok    float64
	Status       string // PENDING, DISETUJUI, AKTIF, LUNAS, DITOLAK
	TanggalCair  string
	Installments []Installment
}

type SimpananTransaction struct {
	ID         int
	MemberNama string
	Jenis      string // POKOK, WAJIB, SUKARELA
	Tipe       string // MASUK, KELUAR
	Nominal    float64
	Keterangan string
	CreatedAt  string
}

type JournalEntry struct {
	ID            int
	Tanggal       string
	Keterangan    string
	AkunDebit     string
	AkunKredit    string
	Nominal       float64
	TipeTransaksi string // SIMPANAN, POS, PINJAMAN, MANUAL
}

// ============================================================
// E-Commerce Models (separate from Koperasi)
// ============================================================

// ECommerceUser represents an e-commerce user account (separate from koperasi User).
type ECommerceUser struct {
	ID                     int
	Username               string
	Email                  string
	Password               string
	Role                   string  // BUYER, SELLER, ADMIN, KASIR
	IsSellerActive         bool    // Has activated seller account?
	SellerRating           float64 // Average seller rating (0-5)
	LinkedKoperasiMemberID int     // Optional link to koperasi Member.ID (0 = not linked)
	CreatedAt              string
}

// SellerProfile holds extra seller information for an ECommerceUser.
type SellerProfile struct {
	SellerID     int     // Same as ECommerceUser.ID
	StoreName    string
	Description  string
	Rating       float64
	ResponseTime string  // e.g. "< 1 jam"
	TotalSold    int
	JoinedAt     string
}

// ECAddress represents a user shipping address.
type ECAddress struct {
	ID        int
	UserID    int    // ECommerceUser.ID
	Label     string // e.g. "Rumah", "Kantor"
	Penerima  string // Recipient name
	NoHP      string
	Alamat    string // Full address
	Kota      string
	Provinsi  string
	KodePos   string
	IsDefault bool
}

// ECProduct represents a product in the e-commerce marketplace.
type ECProduct struct {
	ID          int
	SellerID    int     // ECommerceUser.ID of the seller
	SellerName  string  // Denormalized seller/store name
	Nama        string
	Deskripsi   string
	Kategori    string  // e.g. "Sembako", "Elektronik", "Fashion"
	Harga       float64
	Stok        int
	Berat       int     // grams
	FotoURL     string
	Rating      float64 // Average rating
	TotalReview int
	TotalSold   int
	Status      string  // PENDING_APPROVAL, APPROVED, REJECTED, ARCHIVED
	CreatedAt   string
}

// ShippingOption represents a delivery/shipping service.
type ShippingOption struct {
	ID       int
	Nama     string  // e.g. "JNE Reguler", "JNE YES"
	Provider string  // e.g. "JNE", "J&T", "SiCepat"
	Estimasi string  // e.g. "2-3 hari"
	Harga    float64 // base cost
}

// ECOrder represents a buyer's order in the e-commerce system.
type ECOrder struct {
	ID             int
	NomorOrder     string
	BuyerID        int
	BuyerName      string
	SellerID       int
	SellerName     string
	Items          []ECOrderItem
	AlamatPengiriman string
	ShippingOption string
	ShippingCost   float64
	Subtotal       float64  // sum of items
	Discount       float64  // voucher discount
	PointsUsed     float64  // points redeemed
	TotalHarga     float64  // final total
	VoucherCode    string
	MetodeBayar    string
	Status         string   // PENDING, DIBAYAR, DIPROSES, DIKIRIM, SELESAI, BATAL
	ResiPengiriman string
	PointsEarned   float64
	CreatedAt      string
	UpdatedAt      string
}

// ECOrderItem represents a single item within an ECOrder.
type ECOrderItem struct {
	ProductID   int
	ProductNama string
	SellerID    int
	Jumlah      int
	HargaSatuan float64
	Subtotal    float64
}

// ProductReview is a rating & review left by a buyer.
type ProductReview struct {
	ID         int
	ProductID  int
	UserID     int
	Username   string
	Rating     int    // 1-5
	Komentar   string
	CreatedAt  string
}

// Wishlist tracks products a user has favorited.
type Wishlist struct {
	ID        int
	UserID    int
	ProductID int
	CreatedAt string
}

// Voucher represents a promotional discount code.
type Voucher struct {
	ID             int
	Code           string
	Deskripsi      string
	TipeDiskon     string  // PERCENT, FIXED
	NilaiDiskon    float64 // percentage (e.g. 10) or fixed amount (e.g. 15000)
	MinPembelian   float64 // minimum order total
	MaksDiskon     float64 // max discount cap (for PERCENT type)
	Kuota          int     // remaining uses
	Status         string  // ACTIVE, EXPIRED, USED_UP
	BerlakuSampai  string
}

// UserPoints holds the current points balance for an e-commerce user.
type UserPoints struct {
	UserID       int
	Balance      float64
	TotalEarned  float64
	TotalRedeemed float64
}

// PointsTransaction records a single points earn or redemption event.
type PointsTransaction struct {
	ID        int
	UserID    int
	Tipe      string  // EARN_PURCHASE, REDEEM_VOUCHER, REDEEM_DISCOUNT, CONVERT_SIMPANAN
	Amount    float64 // positive = earn, negative = redeem
	OrderID   int     // related order (0 if N/A)
	Keterangan string
	CreatedAt string
}

// ShipmentEvent records a tracking event for an order's shipment.
type ShipmentEvent struct {
	ID        int
	OrderID   int
	Status    string // DIKEMAS, DIKIRIM, TRANSIT, SAMPAI
	Lokasi    string
	Keterangan string
	CreatedAt string
}

// AuditLog records admin/system actions for auditing.
type AuditLog struct {
	ID        int
	Action    string // e.g. "APPROVE_PRODUCT", "REJECT_SELLER"
	UserID    int    // Who performed the action
	Username  string
	Resource  string // e.g. "product:5", "seller:2"
	Details   string
	CreatedAt string
}
