package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
)

// ErrECUserNotFound menandakan sesi menunjuk ke akun EC yang tidak ada lagi.
// Handler memakai ini untuk membersihkan sesi dan mengarahkan ke login.
var ErrECUserNotFound = errors.New("ec user not found")

// ECAccountService: aturan bisnis akun E-Commerce (autentikasi, pendaftaran,
// integrasi koperasi, dashboard buyer). Memakai ECommerceRepository.
type ECAccountService struct {
	repo repository.ECommerceRepository
}

func NewECAccountService(repo repository.ECommerceRepository) *ECAccountService {
	return &ECAccountService{repo: repo}
}

// Authenticate memvalidasi kredensial; mengembalikan user atau nil bila salah.
func (s *ECAccountService) Authenticate(ctx context.Context, email, password string) (*model.ECommerceUser, error) {
	return s.repo.AuthenticateUser(ctx, email, password)
}

func (s *ECAccountService) UserByEmail(ctx context.Context, email string) (*model.ECommerceUser, error) {
	return s.repo.UserByEmail(ctx, email)
}

func (s *ECAccountService) EnsureUserFromKoperasi(ctx context.Context, email string) (*model.ECommerceUser, error) {
	return s.repo.EnsureUserFromKoperasi(ctx, email)
}

func (s *ECAccountService) UserByID(ctx context.Context, id int) (*model.ECommerceUser, error) {
	return s.repo.UserByID(ctx, id)
}

// Signup memvalidasi & membuat akun BUYER baru. Mengembalikan error berisi
// pesan yang siap ditampilkan bila username/email sudah dipakai.
func (s *ECAccountService) Signup(ctx context.Context, username, email, password string) (*model.ECommerceUser, error) {
	taken, err := s.repo.UsernameExists(ctx, username)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, fmt.Errorf("Username sudah digunakan.")
	}
	emailTaken, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return nil, err
	}
	if emailTaken {
		return nil, fmt.Errorf("Email sudah terdaftar.")
	}
	return s.repo.CreateUser(ctx, username, email, password)
}

// ---- Integrasi koperasi ----

func (s *ECAccountService) MemberByID(ctx context.Context, id int) (*model.Member, error) {
	return s.repo.MemberByID(ctx, id)
}

func (s *ECAccountService) SearchMembers(ctx context.Context, q string) ([]model.Member, error) {
	return s.repo.SearchMembers(ctx, q)
}

// LinkMember menautkan EC user ke member koperasi (cek konflik & keberadaan member).
// Mengembalikan member yang ditautkan untuk keperluan flash message.
func (s *ECAccountService) LinkMember(ctx context.Context, ecUserID, memberID int) (*model.Member, error) {
	m, err := s.repo.MemberByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("Member koperasi tidak ditemukan.")
	}
	conflict, err := s.repo.MemberLinkedToOtherUser(ctx, memberID, ecUserID)
	if err != nil {
		return nil, err
	}
	if conflict {
		return nil, fmt.Errorf("Member koperasi ini sudah di-link ke akun E-Commerce lain.")
	}
	if err := s.repo.LinkMember(ctx, ecUserID, memberID); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *ECAccountService) UnlinkMember(ctx context.Context, ecUserID int) error {
	_, err := s.repo.UnlinkMember(ctx, ecUserID)
	return err
}

// ConvertToSimpanan mengonversi poin user menjadi simpanan sukarela member
// koperasi yang tertaut (1 poin = PointValueRupiah). Mengembalikan pesan sukses,
// jumlah rupiah, dan saldo sukarela terbaru.
func (s *ECAccountService) ConvertToSimpanan(ctx context.Context, ecUserID int, points float64) (msg string, rupiah, newSukarela float64, err error) {
	if points < MinPointConversion {
		return "", 0, 0, fmt.Errorf("Minimum konversi %.0f poin.", MinPointConversion)
	}
	u, err := s.repo.UserByID(ctx, ecUserID)
	if err != nil {
		return "", 0, 0, err
	}
	if u == nil {
		return "", 0, 0, fmt.Errorf("User tidak ditemukan.")
	}
	if u.LinkedKoperasiMemberID == 0 {
		return "", 0, 0, fmt.Errorf("Akun belum di-link dengan member koperasi.")
	}
	up, err := s.repo.UserPoints(ctx, ecUserID)
	if err != nil {
		return "", 0, 0, err
	}
	if up == nil || up.Balance < points {
		return "", 0, 0, fmt.Errorf("Poin tidak cukup.")
	}
	m, err := s.repo.MemberByID(ctx, u.LinkedKoperasiMemberID)
	if err != nil {
		return "", 0, 0, err
	}
	if m == nil {
		return "", 0, 0, fmt.Errorf("Member koperasi tidak ditemukan.")
	}
	if m.Status != "AKTIF" {
		return "", 0, 0, fmt.Errorf("Member koperasi belum aktif. Konversi tersedia setelah disetujui pengurus.")
	}
	rupiah = points * PointValueRupiah
	newSukarela, err = s.repo.ConvertPointsToSimpanan(ctx, ecUserID, m.ID, points, rupiah)
	if err != nil {
		return "", 0, 0, err
	}
	msg = fmt.Sprintf("Berhasil konversi %.0f poin (Rp %.0f) ke Simpanan Sukarela %s", points, rupiah, m.Nama)
	return msg, rupiah, newSukarela, nil
}

// ---- Dashboard buyer ----

// BuyerDashboard menampung data ringkas untuk dashboard buyer.
type BuyerDashboard struct {
	RecentOrders  []model.ECOrder
	OrderCount    int
	PointsBalance float64
	Addresses     []model.ECAddress
	WishlistCount int
	Recommended   []model.ECProduct
	LinkedMember  string
}

func (s *ECAccountService) BuyerDashboard(ctx context.Context, ecUserID int) (*BuyerDashboard, error) {
	orders, err := s.repo.OrdersByBuyer(ctx, ecUserID)
	if err != nil {
		return nil, err
	}
	d := &BuyerDashboard{OrderCount: len(orders)}
	d.RecentOrders = orders
	if len(d.RecentOrders) > 5 {
		d.RecentOrders = d.RecentOrders[:5]
	}

	if up, err := s.repo.UserPoints(ctx, ecUserID); err != nil {
		return nil, err
	} else if up != nil {
		d.PointsBalance = up.Balance
	}

	if d.Addresses, err = s.repo.AddressesByUser(ctx, ecUserID); err != nil {
		return nil, err
	}

	wishlist, err := s.repo.WishlistProducts(ctx, ecUserID)
	if err != nil {
		return nil, err
	}
	d.WishlistCount = len(wishlist)

	recommended, err := s.repo.ProductsApproved(ctx)
	if err != nil {
		return nil, err
	}
	if len(recommended) > 4 {
		recommended = recommended[:4]
	}
	d.Recommended = recommended

	if u, err := s.repo.UserByID(ctx, ecUserID); err != nil {
		return nil, err
	} else if u != nil && u.LinkedKoperasiMemberID > 0 {
		if m, err := s.repo.MemberByID(ctx, u.LinkedKoperasiMemberID); err != nil {
			return nil, err
		} else if m != nil {
			d.LinkedMember = strings.TrimSpace(m.Nama + " (" + m.NomorAnggota + ")")
		}
	}

	return d, nil
}

// ---- Profil & alamat (4d-3) ----

// ProfileOverview menampung ringkasan untuk halaman profil.
type ProfileOverview struct {
	User            *model.ECommerceUser
	TotalOrders     int
	CompletedOrders int
	PointsBalance   float64
	AddressCount    int
	LinkedMember    *model.Member
}

// ProfileOverview mengembalikan ringkasan profil; nil bila user tidak ada.
func (s *ECAccountService) ProfileOverview(ctx context.Context, ecUserID int) (*ProfileOverview, error) {
	u, err := s.repo.UserByID(ctx, ecUserID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, nil
	}
	o := &ProfileOverview{User: u}

	orders, err := s.repo.OrdersByBuyer(ctx, ecUserID)
	if err != nil {
		return nil, err
	}
	o.TotalOrders = len(orders)
	for _, ord := range orders {
		if ord.Status == "SELESAI" {
			o.CompletedOrders++
		}
	}

	if up, err := s.repo.UserPoints(ctx, ecUserID); err != nil {
		return nil, err
	} else if up != nil {
		o.PointsBalance = up.Balance
	}

	addrs, err := s.repo.AddressesByUser(ctx, ecUserID)
	if err != nil {
		return nil, err
	}
	o.AddressCount = len(addrs)

	if u.LinkedKoperasiMemberID > 0 {
		if o.LinkedMember, err = s.repo.MemberByID(ctx, u.LinkedKoperasiMemberID); err != nil {
			return nil, err
		}
	}
	return o, nil
}

func (s *ECAccountService) Addresses(ctx context.Context, ecUserID int) ([]model.ECAddress, error) {
	return s.repo.AddressesByUser(ctx, ecUserID)
}

func (s *ECAccountService) CreateAddress(ctx context.Context, ecUserID int, a model.ECAddress) error {
	return s.repo.CreateAddress(ctx, ecUserID, a)
}

func (s *ECAccountService) SetDefaultAddress(ctx context.Context, ecUserID, addressID int) error {
	return s.repo.SetDefaultAddress(ctx, ecUserID, addressID)
}

func (s *ECAccountService) DeleteAddress(ctx context.Context, ecUserID, addressID int) error {
	return s.repo.DeleteAddress(ctx, ecUserID, addressID)
}

// ---- Poin & loyalty (4d-4) ----

// LinkedMember mengembalikan member koperasi yang tertaut ke EC user, atau nil.
func (s *ECAccountService) LinkedMember(ctx context.Context, ecUserID int) (*model.Member, error) {
	u, err := s.repo.UserByID(ctx, ecUserID)
	if err != nil {
		return nil, err
	}
	if u == nil || u.LinkedKoperasiMemberID == 0 {
		return nil, nil
	}
	return s.repo.MemberByID(ctx, u.LinkedKoperasiMemberID)
}

func (s *ECAccountService) PointsTransactions(ctx context.Context, ecUserID int) ([]model.PointsTransaction, error) {
	return s.repo.PointsTransactions(ctx, ecUserID)
}

// PointsOverview menampung data halaman poin.
type PointsOverview struct {
	Balance       float64
	TotalEarned   float64
	TotalRedeemed float64
	RpEquivalent  float64
	Transactions  []model.PointsTransaction
	LinkedMember  *model.Member
}

func (s *ECAccountService) PointsOverview(ctx context.Context, ecUserID int) (*PointsOverview, error) {
	o := &PointsOverview{}
	if up, err := s.repo.UserPoints(ctx, ecUserID); err != nil {
		return nil, err
	} else if up != nil {
		o.Balance = up.Balance
		o.TotalEarned = up.TotalEarned
		o.TotalRedeemed = up.TotalRedeemed
	}
	o.RpEquivalent = o.Balance * PointValueRupiah

	var err error
	if o.Transactions, err = s.repo.PointsTransactions(ctx, ecUserID); err != nil {
		return nil, err
	}
	if o.LinkedMember, err = s.LinkedMember(ctx, ecUserID); err != nil {
		return nil, err
	}
	return o, nil
}

// RegisterMemberAsKoperasi mendaftarkan EC user sebagai anggota koperasi baru
// (status PENDING) dan menautkannya. Mengembalikan ErrECUserNotFound bila sesi
// tidak valid, atau error berpesan untuk validasi (sudah linked / NIK terpakai).
func (s *ECAccountService) RegisterMemberAsKoperasi(ctx context.Context, ecUserID int, nama, nik, alamat, noHP string) error {
	u, err := s.repo.UserByID(ctx, ecUserID)
	if err != nil {
		return err
	}
	if u == nil {
		return ErrECUserNotFound
	}
	if u.LinkedKoperasiMemberID != 0 {
		return fmt.Errorf("Akun Anda sudah terhubung ke member koperasi.")
	}
	exists, err := s.repo.NIKExists(ctx, nik)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("NIK sudah terdaftar pada anggota lain.")
	}
	_, err = s.repo.RegisterMemberAndLink(ctx, ecUserID, nama, nik, alamat, noHP)
	return err
}

// UpdateSettings memperbarui username/email (memvalidasi keunikan username).
// Field kosong dipertahankan dari nilai lama.
func (s *ECAccountService) UpdateSettings(ctx context.Context, ecUserID int, username, email string) error {
	u, err := s.repo.UserByID(ctx, ecUserID)
	if err != nil {
		return err
	}
	if u == nil {
		return fmt.Errorf("User tidak ditemukan.")
	}
	newUsername := u.Username
	if username != "" && username != u.Username {
		other, err := s.repo.UserByUsername(ctx, username)
		if err != nil {
			return err
		}
		if other != nil && other.ID != ecUserID {
			return fmt.Errorf("Username sudah digunakan.")
		}
		newUsername = username
	}
	newEmail := u.Email
	if email != "" {
		newEmail = email
	}
	return s.repo.UpdateUserContact(ctx, ecUserID, newUsername, newEmail)
}
