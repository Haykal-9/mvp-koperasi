// Package repository mendefinisikan kontrak akses data (interface) dan
// implementasinya untuk PostgreSQL. Interface membuat service/handler tidak
// bergantung pada detail basis data (dapat ditukar implementasinya).
package repository

import (
	"context"

	"koperasi-frontend/core/model"
)

// UserRepository: akses akun dan autentikasi domain koperasi.
type UserRepository interface {
	// Authenticate memverifikasi password bcrypt dan tidak pernah mengembalikan hash.
	Authenticate(ctx context.Context, email, password string) (*model.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	// Create menyimpan akun ANGGOTA baru dengan password bcrypt.
	Create(ctx context.Context, nama, email, password string) (*model.User, error)
}

// MemberCounts menampung jumlah anggota per status untuk badge tab.
type MemberCounts struct {
	Aktif    int
	Pending  int
	NonAktif int
	All      int
}

// MemberRepository: akses data domain Keanggotaan (M1).
type MemberRepository interface {
	List(ctx context.Context, status string) ([]model.Member, error)
	Counts(ctx context.Context) (MemberCounts, error)
	FindByID(ctx context.Context, id int) (*model.Member, error)
	// FindByNama mengembalikan anggota dengan nama persis, atau nil. Dipakai untuk
	// sidebar (MemberID) dan dashboard anggota.
	FindByNama(ctx context.Context, nama string) (*model.Member, error)
	Create(ctx context.Context, nama, nik, alamat, noHP string) (int, error)
	SetStatus(ctx context.Context, id int, status string) error
	// Approve mengubah status PENDING -> AKTIF, membangkitkan NomorAnggota,
	// dan mengisi tanggal masuk. Mengembalikan nomor anggota baru.
	Approve(ctx context.Context, id int) (string, error)
	SimpananHistory(ctx context.Context, memberID int) ([]model.SimpananTransaction, error)
	LoansByMember(ctx context.Context, memberID int) ([]model.Loan, error)
	// RecordSimpanan mencatat transaksi simpanan secara transaksional:
	// perbarui saldo agregat anggota + sisipkan transaksi + jurnal otomatis.
	RecordSimpanan(ctx context.Context, memberID int, jenis, tipe string, nominal float64, keterangan string) error
}

// ProductRepository: akses data produk & stok (M3 POS).
type ProductRepository interface {
	// Catalog mengembalikan produk APPROVED terfilter (q pada nama, kategori opsional).
	Catalog(ctx context.Context, q, kategori string) ([]model.Product, error)
	// Categories mengembalikan daftar kategori unik dari seluruh produk.
	Categories(ctx context.Context) ([]string, error)
	ListByStatus(ctx context.Context, status string) ([]model.Product, error)
	FindByID(ctx context.Context, id int) (*model.Product, error)
	// Create menyisipkan produk; bila stok awal > 0, catat stock_change RESTOCK
	// dalam satu transaksi. Mengembalikan id baru.
	Create(ctx context.Context, p model.Product) (int, error)
	SetStatus(ctx context.Context, id int, status string) error
	StockHistory(ctx context.Context, productID int) ([]model.StockChange, error)
	// AdjustStock memperbarui stok + mencatat stock_change transaksional.
	// jumlah bisa negatif; menolak bila stok menjadi negatif.
	AdjustStock(ctx context.Context, productID int, tipe string, jumlah int, keterangan string) error
}

// OrderRepository: akses data order POS (M3).
type OrderRepository interface {
	// List mengembalikan order (terbaru dulu); bila scopeNama != "" hanya milik pembeli itu.
	List(ctx context.Context, scopeNama string) ([]model.Order, error)
	ListByStatus(ctx context.Context, status string) ([]model.Order, error)
	FindByID(ctx context.Context, id int) (*model.Order, error)
	// Create membuat order secara transaksional: bangkitkan nomor order, sisipkan
	// order + item, kurangi stok (+stock_changes), dan catat jurnal POS.
	Create(ctx context.Context, o model.Order) (*model.Order, error)
	SetStatus(ctx context.Context, id int, status string) error
	// SetComplaint menandai order DISPUTED beserta alasan & bukti komplain.
	SetComplaint(ctx context.Context, id int, alasan, bukti string) error
	// Resolve menuntaskan komplain: approveRetur=true -> BATAL + kembalikan stok,
	// false -> SELESAI. Transaksional.
	Resolve(ctx context.Context, id int, approveRetur bool) error
}

// LoanRepository: akses data pinjaman & angsuran (M4).
type LoanRepository interface {
	// List mengembalikan pinjaman (terbaru dulu, tanpa angsuran); bila scopeNama != ""
	// hanya milik anggota itu.
	List(ctx context.Context, scopeNama string) ([]model.Loan, error)
	// FindByID mengembalikan satu pinjaman beserta jadwal angsurannya.
	FindByID(ctx context.Context, id int) (*model.Loan, error)
	// Create menyisipkan pengajuan pinjaman PENDING; member_id di-resolve dari
	// memberNama (error bila anggota tak ditemukan). Mengembalikan id baru.
	Create(ctx context.Context, memberNama string, l model.Loan) (int, error)
	SetStatus(ctx context.Context, id int, status string) error
	// Disburse mencairkan pinjaman secara transaksional: set AKTIF + tanggal_cair,
	// sisipkan seluruh angsuran, dan catat jurnal pencairan.
	Disburse(ctx context.Context, id int, tanggalCair string, installments []model.Installment, journal model.JournalEntry) error
	// Pay membayar satu angsuran secara transaksional: tandai angsuran DIBAYAR,
	// perbarui sisa_pokok (+status LUNAS bila lunas), dan catat jurnal angsuran.
	Pay(ctx context.Context, id, bulanKe int, tanggalBayar string, newSisaPokok float64, lunas bool, journal model.JournalEntry) error
	// CreditInfo mengembalikan total simpanan & jumlah pinjaman AKTIF anggota
	// (untuk informasi credit scoring di form pengajuan).
	CreditInfo(ctx context.Context, memberNama string) (totalSimpanan float64, activeLoans int, err error)
}

// JournalCounts menampung jumlah entri jurnal per tipe transaksi (badge tab).
type JournalCounts struct {
	All      int
	Simpanan int
	POS      int
	Pinjaman int
	Manual   int
}

// FinanceSummary menampung saldo agregat buku besar (PRD §4.6 ringkasan keuangan).
type FinanceSummary struct {
	TotalKas        float64
	TotalSimpanan   float64
	TotalPiutang    float64
	TotalPendapatan float64
}

// MonthlyCashflow menampung arus kas per bulan ('YYYY-MM') berbasis akun Kas.
type MonthlyCashflow struct {
	Bulan       string
	Pemasukan   float64
	Pengeluaran float64
}

// AuditAction membawa metadata audit untuk operasi admin yang harus atomik
// dengan mutasi data yang dicatatnya.
type AuditAction struct {
	Action   string
	UserID   int
	Username string
	Resource string
	Details  string
}

// ECommerceRepository: akses data domain E-Commerce SmartMart (M5).
// Dibangun bertahap per sub-fase 4d; bagian katalog (read) terlebih dahulu.
type ECommerceRepository interface {
	// ---- Katalog / storefront (4d-1) ----
	// ProductsApproved mengembalikan seluruh produk APPROVED (terbaru dulu).
	ProductsApproved(ctx context.Context) ([]model.ECProduct, error)
	// ProductsByCategory mengembalikan produk APPROVED pada kategori tertentu.
	ProductsByCategory(ctx context.Context, kategori string) ([]model.ECProduct, error)
	// ProductsBySeller mengembalikan seluruh produk milik seller (semua status).
	ProductsBySeller(ctx context.Context, sellerID int) ([]model.ECProduct, error)
	// SearchProducts mencari produk APPROVED pada nama/deskripsi/kategori.
	SearchProducts(ctx context.Context, q string) ([]model.ECProduct, error)
	// ProductByID mengembalikan satu produk (status apa pun) atau nil.
	ProductByID(ctx context.Context, id int) (*model.ECProduct, error)
	// Categories mengembalikan daftar kategori unik dari produk APPROVED.
	Categories(ctx context.Context) ([]string, error)
	// SellerProfile mengembalikan profil seller atau nil.
	SellerProfile(ctx context.Context, sellerID int) (*model.SellerProfile, error)
	// SellerProfiles mengembalikan seluruh profil seller.
	SellerProfiles(ctx context.Context) ([]model.SellerProfile, error)
	// ReviewsByProduct mengembalikan ulasan satu produk (terbaru dulu).
	ReviewsByProduct(ctx context.Context, productID int) ([]model.ProductReview, error)
	// IsWishlisted memeriksa apakah produk ada di wishlist user.
	IsWishlisted(ctx context.Context, userID, productID int) (bool, error)

	// ---- Akun & autentikasi (4d-2) ----
	// AuthenticateUser memvalidasi email+password (bcrypt) terhadap ecommerce_users.
	// Mengembalikan user (tanpa password) bila valid, nil bila kredensial salah.
	AuthenticateUser(ctx context.Context, email, password string) (*model.ECommerceUser, error)
	// EnsureUserFromKoperasi memastikan akun e-commerce tersedia untuk email user
	// koperasi yang sudah terautentikasi. Dipakai untuk SSO dari session koperasi.
	EnsureUserFromKoperasi(ctx context.Context, email string) (*model.ECommerceUser, error)
	UserByEmail(ctx context.Context, email string) (*model.ECommerceUser, error)
	UserByID(ctx context.Context, id int) (*model.ECommerceUser, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	// CreateUser menyisipkan akun BUYER baru (password di-hash bcrypt) dan
	// menginisialisasi baris user_points dalam satu transaksi.
	CreateUser(ctx context.Context, username, email, password string) (*model.ECommerceUser, error)

	// ---- Integrasi koperasi (4d-2) ----
	MemberByID(ctx context.Context, id int) (*model.Member, error)
	// SearchMembers mencari anggota pada nama/nomor anggota (maks 5).
	SearchMembers(ctx context.Context, q string) ([]model.Member, error)
	// MemberLinkedToOtherUser true bila member sudah di-link ke EC user lain.
	MemberLinkedToOtherUser(ctx context.Context, memberID, exceptUserID int) (bool, error)
	// LinkMember menautkan EC user ke member koperasi + catat audit (transaksional).
	LinkMember(ctx context.Context, ecUserID, memberID int) error
	// UnlinkMember melepas tautan + catat audit; mengembalikan member lama.
	UnlinkMember(ctx context.Context, ecUserID int) (oldMemberID int, err error)
	// ConvertPointsToSimpanan mengurangi poin user dan menambah simpanan sukarela
	// member (1 transaksi: deduct saldo + points_transaction + update member).
	ConvertPointsToSimpanan(ctx context.Context, ecUserID, memberID int, points, rupiah float64) (newSukarela float64, err error)

	// ---- Audit (4d-2) ----
	LogAudit(ctx context.Context, action string, userID int, username, resource, details string) error

	// ---- Dashboard buyer / read (4d-2) ----
	// OrdersByBuyer mengembalikan order milik buyer (terbaru dulu, tanpa item).
	OrdersByBuyer(ctx context.Context, buyerID int) ([]model.ECOrder, error)
	UserPoints(ctx context.Context, userID int) (*model.UserPoints, error)
	AddressesByUser(ctx context.Context, userID int) ([]model.ECAddress, error)
	// WishlistProducts mengembalikan produk yang ada di wishlist user.
	WishlistProducts(ctx context.Context, userID int) ([]model.ECProduct, error)

	// ---- Alur beli / order (4d-3) ----
	ShippingOptions(ctx context.Context) ([]model.ShippingOption, error)
	ShippingOptionByID(ctx context.Context, id int) (*model.ShippingOption, error)
	VoucherByCode(ctx context.Context, code string) (*model.Voucher, error)
	// CreateOrder menyimpan order secara transaksional: bangkitkan nomor order,
	// sisipkan order + item, kurangi stok produk, dan catat poin (earn/redeem).
	CreateOrder(ctx context.Context, o model.ECOrder, pointsUsed, pointsEarned float64) (*model.ECOrder, error)
	// OrderByID mengembalikan satu order beserta itemnya, atau nil.
	OrderByID(ctx context.Context, id int) (*model.ECOrder, error)
	ShipmentEvents(ctx context.Context, orderID int) ([]model.ShipmentEvent, error)
	// ToggleWishlist menambah/menghapus produk dari wishlist; added=true bila ditambah.
	ToggleWishlist(ctx context.Context, userID, productID int) (added bool, err error)

	// ---- Profil & alamat (4d-3) ----
	UserByUsername(ctx context.Context, username string) (*model.ECommerceUser, error)
	CreateAddress(ctx context.Context, userID int, a model.ECAddress) error
	SetDefaultAddress(ctx context.Context, userID, addressID int) error
	DeleteAddress(ctx context.Context, userID, addressID int) error
	UpdateUserContact(ctx context.Context, userID int, username, email string) error

	// ---- Poin & loyalty (4d-4) ----
	PointsTransactions(ctx context.Context, userID int) ([]model.PointsTransaction, error)
	NIKExists(ctx context.Context, nik string) (bool, error)
	// RegisterMemberAndLink membuat anggota koperasi PENDING dari data EC user
	// dan menautkannya ke akun EC dalam satu transaksi (+audit). Mengembalikan id member.
	RegisterMemberAndLink(ctx context.Context, ecUserID int, nama, nik, alamat, noHP string) (memberID int, err error)

	// ---- Seller (4d-5) ----
	// OrdersBySeller mengembalikan order yang diterima seller (terbaru dulu, tanpa item).
	OrdersBySeller(ctx context.Context, sellerID int) ([]model.ECOrder, error)
	// CreateECProduct menyisipkan produk seller (status PENDING_APPROVAL). Mengembalikan id.
	CreateECProduct(ctx context.Context, p model.ECProduct) (int, error)
	// MarkOrderShipped menandai order DIKIRIM + resi dan mencatat shipment event (transaksional).
	MarkOrderShipped(ctx context.Context, sellerID, orderID int, resi, lokasi, keterangan string) error

	// ---- Review (4d-6) ----
	HasReviewed(ctx context.Context, userID, productID int) (bool, error)
	// SubmitReview menyisipkan ulasan + memperbarui rating/total_review produk (transaksional).
	SubmitReview(ctx context.Context, userID, productID, rating int, komentar, username string) error

	// ---- Admin (4d-6) ----
	AllProducts(ctx context.Context) ([]model.ECProduct, error)
	AllOrders(ctx context.Context) ([]model.ECOrder, error)
	AllUsers(ctx context.Context) ([]model.ECommerceUser, error)
	AllVouchers(ctx context.Context) ([]model.Voucher, error)
	AuditLogs(ctx context.Context) ([]model.AuditLog, error)
	AllPointsTransactions(ctx context.Context) ([]model.PointsTransaction, error)
	AllUserPoints(ctx context.Context) ([]model.UserPoints, error)
	// ActivateSeller mengaktifkan akun seller + membuat seller_profile (transaksional).
	ActivateSeller(ctx context.Context, userID int, storeName string) error
	ActivateSellerWithAudit(ctx context.Context, userID int, storeName string, audit AuditAction) error
	SetECProductStatus(ctx context.Context, id int, status string) error
	SetECProductStatusWithAudit(ctx context.Context, id int, status, expectedStatus string, audit AuditAction) error
	VoucherByID(ctx context.Context, id int) (*model.Voucher, error)
	CreateVoucher(ctx context.Context, v model.Voucher) (int, error)
	CreateVoucherWithAudit(ctx context.Context, v model.Voucher, audit AuditAction) (int, error)
	SetVoucherStatus(ctx context.Context, id int, status string) error
	SetVoucherStatusWithAudit(ctx context.Context, id int, status, expectedStatus string, audit AuditAction) error
}

// JournalRepository: akses data buku besar / jurnal keuangan (M2).
type JournalRepository interface {
	// List mengembalikan entri jurnal (terbaru dulu); tipe filter opsional.
	List(ctx context.Context, tipe string) ([]model.JournalEntry, error)
	Counts(ctx context.Context) (JournalCounts, error)
	// Create menyisipkan entri jurnal (mis. jurnal manual MANUAL).
	Create(ctx context.Context, j model.JournalEntry) error
	// Summary menghitung saldo agregat Kas/Simpanan/Piutang/Pendapatan.
	Summary(ctx context.Context) (FinanceSummary, error)
	// MonthlyCashflow mengembalikan arus kas per bulan (semua bulan yang ada data).
	MonthlyCashflow(ctx context.Context) ([]MonthlyCashflow, error)
}
