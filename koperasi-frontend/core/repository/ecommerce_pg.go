package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/security"
)

// pgECommerceRepository: implementasi ECommerceRepository di atas PostgreSQL (M5).
// Dibangun bertahap per sub-fase 4d.
type pgECommerceRepository struct {
	db *gorm.DB
}

func isPGUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func NewECommerceRepository(db *gorm.DB) ECommerceRepository {
	return &pgECommerceRepository{db: db}
}

// Kolom eksplisit: tanggal disimpan TIMESTAMPTZ tetapi model memakai string,
// jadi diformat lewat to_char agar bisa di-scan langsung (ECProduct dsb. tanpa slice).
const ecProductCols = `id, seller_id, seller_name, nama, deskripsi, kategori,
	harga, stok, berat, foto_url, rating, total_review, total_sold, status,
	to_char(created_at,'YYYY-MM-DD') AS created_at`

const sellerProfileCols = `seller_id, store_name, description, rating,
	response_time, total_sold, to_char(joined_at,'YYYY-MM-DD') AS joined_at`

const reviewCols = `id, product_id, user_id, username, rating, komentar,
	to_char(created_at,'YYYY-MM-DD') AS created_at`

// ---- Katalog / storefront (4d-1) ----

func (r *pgECommerceRepository) ProductsApproved(ctx context.Context) ([]model.ECProduct, error) {
	out := []model.ECProduct{}
	return out, r.db.WithContext(ctx).Raw(
		"SELECT " + ecProductCols + " FROM ec_products WHERE status = 'APPROVED' ORDER BY id DESC").
		Scan(&out).Error
}

func (r *pgECommerceRepository) ProductsByCategory(ctx context.Context, kategori string) ([]model.ECProduct, error) {
	out := []model.ECProduct{}
	return out, r.db.WithContext(ctx).Raw(
		"SELECT "+ecProductCols+" FROM ec_products WHERE status = 'APPROVED' AND lower(kategori) = lower(?) ORDER BY id DESC",
		kategori).Scan(&out).Error
}

func (r *pgECommerceRepository) ProductsBySeller(ctx context.Context, sellerID int) ([]model.ECProduct, error) {
	out := []model.ECProduct{}
	return out, r.db.WithContext(ctx).Raw(
		"SELECT "+ecProductCols+" FROM ec_products WHERE seller_id = ? ORDER BY id DESC",
		sellerID).Scan(&out).Error
}

func (r *pgECommerceRepository) SearchProducts(ctx context.Context, q string) ([]model.ECProduct, error) {
	out := []model.ECProduct{}
	like := "%" + q + "%"
	return out, r.db.WithContext(ctx).Raw(
		`SELECT `+ecProductCols+` FROM ec_products
		 WHERE status = 'APPROVED'
		   AND (nama ILIKE ? OR deskripsi ILIKE ? OR kategori ILIKE ?)
		 ORDER BY id DESC`, like, like, like).Scan(&out).Error
}

func (r *pgECommerceRepository) ProductByID(ctx context.Context, id int) (*model.ECProduct, error) {
	out := []model.ECProduct{}
	if err := r.db.WithContext(ctx).Raw(
		"SELECT "+ecProductCols+" FROM ec_products WHERE id = ? LIMIT 1", id).
		Scan(&out).Error; err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, nil
	}
	return &out[0], nil
}

func (r *pgECommerceRepository) Categories(ctx context.Context) ([]string, error) {
	out := []string{}
	return out, r.db.WithContext(ctx).Raw(
		"SELECT DISTINCT kategori FROM ec_products WHERE status = 'APPROVED' AND kategori <> '' ORDER BY kategori").
		Scan(&out).Error
}

func (r *pgECommerceRepository) SellerProfile(ctx context.Context, sellerID int) (*model.SellerProfile, error) {
	out := []model.SellerProfile{}
	if err := r.db.WithContext(ctx).Raw(
		"SELECT "+sellerProfileCols+" FROM seller_profiles WHERE seller_id = ? LIMIT 1", sellerID).
		Scan(&out).Error; err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, nil
	}
	return &out[0], nil
}

func (r *pgECommerceRepository) SellerProfiles(ctx context.Context) ([]model.SellerProfile, error) {
	out := []model.SellerProfile{}
	return out, r.db.WithContext(ctx).Raw(
		"SELECT " + sellerProfileCols + " FROM seller_profiles ORDER BY seller_id").
		Scan(&out).Error
}

func (r *pgECommerceRepository) ReviewsByProduct(ctx context.Context, productID int) ([]model.ProductReview, error) {
	out := []model.ProductReview{}
	return out, r.db.WithContext(ctx).Raw(
		"SELECT "+reviewCols+" FROM product_reviews WHERE product_id = ? ORDER BY id DESC",
		productID).Scan(&out).Error
}

func (r *pgECommerceRepository) IsWishlisted(ctx context.Context, userID, productID int) (bool, error) {
	var n int64
	if err := r.db.WithContext(ctx).Raw(
		"SELECT count(*) FROM wishlists WHERE user_id = ? AND product_id = ?",
		userID, productID).Scan(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

// ---- Akun & autentikasi (4d-2) ----

// Kolom user tanpa password_hash; linked id nullable -> 0 bila NULL.
const ecUserCols = `id, username, email, role, is_seller_active, seller_rating,
	COALESCE(linked_koperasi_member_id, 0) AS linked_koperasi_member_id,
	to_char(created_at,'YYYY-MM-DD') AS created_at`

func (r *pgECommerceRepository) AuthenticateUser(ctx context.Context, email, password string) (*model.ECommerceUser, error) {
	// Ambil hash terpisah dari kolom model agar tidak bocor ke struct domain.
	type authRow struct {
		model.ECommerceUser
		PasswordHash string
	}
	rows := []authRow{}
	if err := r.db.WithContext(ctx).Raw(
		"SELECT "+ecUserCols+", password_hash FROM ecommerce_users WHERE email = ? LIMIT 1",
		email).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	if !security.VerifyPassword(rows[0].PasswordHash, password) {
		return nil, nil
	}
	u := rows[0].ECommerceUser
	return &u, nil
}

func (r *pgECommerceRepository) UserByEmail(ctx context.Context, email string) (*model.ECommerceUser, error) {
	out := []model.ECommerceUser{}
	if err := r.db.WithContext(ctx).Raw(
		"SELECT "+ecUserCols+" FROM ecommerce_users WHERE email = ? LIMIT 1", email).
		Scan(&out).Error; err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, nil
	}
	return &out[0], nil
}

func (r *pgECommerceRepository) UserByID(ctx context.Context, id int) (*model.ECommerceUser, error) {
	out := []model.ECommerceUser{}
	if err := r.db.WithContext(ctx).Raw(
		"SELECT "+ecUserCols+" FROM ecommerce_users WHERE id = ? LIMIT 1", id).
		Scan(&out).Error; err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, nil
	}
	return &out[0], nil
}

func (r *pgECommerceRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var n int64
	if err := r.db.WithContext(ctx).Raw(
		"SELECT count(*) FROM ecommerce_users WHERE username = ?", username).Scan(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *pgECommerceRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var n int64
	if err := r.db.WithContext(ctx).Raw(
		"SELECT count(*) FROM ecommerce_users WHERE email = ?", email).Scan(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *pgECommerceRepository) CreateUser(ctx context.Context, username, email, password string) (*model.ECommerceUser, error) {
	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, err
	}
	var newID int
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := tx.Raw(
			`INSERT INTO ecommerce_users (username, email, password_hash, role, is_seller_active)
			 VALUES (?, ?, ?, 'BUYER', false) RETURNING id`,
			username, email, hash).Scan(&newID).Error; e != nil {
			return e
		}
		return tx.Exec(
			"INSERT INTO user_points (user_id, balance, total_earned, total_redeemed) VALUES (?, 0, 0, 0)",
			newID).Error
	})
	if err != nil {
		return nil, err
	}
	return r.UserByID(ctx, newID)
}

// ---- Integrasi koperasi (4d-2) ----

func (r *pgECommerceRepository) MemberByID(ctx context.Context, id int) (*model.Member, error) {
	out := []model.Member{}
	if err := r.db.WithContext(ctx).Raw(
		"SELECT "+memberCols+" FROM members WHERE id = ? LIMIT 1", id).
		Scan(&out).Error; err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, nil
	}
	return &out[0], nil
}

func (r *pgECommerceRepository) SearchMembers(ctx context.Context, q string) ([]model.Member, error) {
	out := []model.Member{}
	like := "%" + q + "%"
	return out, r.db.WithContext(ctx).Raw(
		"SELECT "+memberCols+" FROM members WHERE nama ILIKE ? OR nomor_anggota ILIKE ? ORDER BY id LIMIT 5",
		like, like).Scan(&out).Error
}

func (r *pgECommerceRepository) MemberLinkedToOtherUser(ctx context.Context, memberID, exceptUserID int) (bool, error) {
	var n int64
	if err := r.db.WithContext(ctx).Raw(
		"SELECT count(*) FROM ecommerce_users WHERE linked_koperasi_member_id = ? AND id <> ?",
		memberID, exceptUserID).Scan(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *pgECommerceRepository) LinkMember(ctx context.Context, ecUserID, memberID int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		type idRow struct {
			ID int
		}
		var member idRow
		if e := tx.Raw(
			"SELECT id FROM members WHERE id = ? FOR UPDATE",
			memberID).Scan(&member).Error; e != nil {
			return e
		}
		if member.ID == 0 {
			return fmt.Errorf("Member koperasi tidak ditemukan.")
		}

		type userLinkRow struct {
			ID                     int
			LinkedKoperasiMemberID int
		}
		var user userLinkRow
		if e := tx.Raw(
			`SELECT id, COALESCE(linked_koperasi_member_id, 0) AS linked_koperasi_member_id
			 FROM ecommerce_users WHERE id = ? FOR UPDATE`,
			ecUserID).Scan(&user).Error; e != nil {
			return e
		}
		if user.ID == 0 {
			return fmt.Errorf("User tidak ditemukan.")
		}

		var otherID int
		if e := tx.Raw(
			`SELECT id FROM ecommerce_users
			 WHERE linked_koperasi_member_id = ? AND id <> ?
			 LIMIT 1 FOR UPDATE`,
			memberID, ecUserID).Scan(&otherID).Error; e != nil {
			return e
		}
		if otherID != 0 {
			return fmt.Errorf("Member koperasi ini sudah di-link ke akun E-Commerce lain.")
		}

		if e := tx.Exec(
			"UPDATE ecommerce_users SET linked_koperasi_member_id = ? WHERE id = ?",
			memberID, ecUserID).Error; e != nil {
			if isPGUniqueViolation(e) {
				return fmt.Errorf("Member koperasi ini sudah di-link ke akun E-Commerce lain.")
			}
			return e
		}
		return tx.Exec(
			`INSERT INTO audit_logs (action, user_id, username, resource, details)
			 SELECT 'LINK_KOPERASI', u.id, u.username, ?, ?
			 FROM ecommerce_users u WHERE u.id = ?`,
			fmt.Sprintf("member:%d", memberID),
			fmt.Sprintf("Linked to koperasi member %d", memberID), ecUserID).Error
	})
}

func (r *pgECommerceRepository) UnlinkMember(ctx context.Context, ecUserID int) (int, error) {
	var oldID int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := tx.Raw(
			"SELECT COALESCE(linked_koperasi_member_id, 0) FROM ecommerce_users WHERE id = ? FOR UPDATE",
			ecUserID).Scan(&oldID).Error; e != nil {
			return e
		}
		if e := tx.Exec(
			"UPDATE ecommerce_users SET linked_koperasi_member_id = NULL WHERE id = ?",
			ecUserID).Error; e != nil {
			return e
		}
		return tx.Exec(
			`INSERT INTO audit_logs (action, user_id, username, resource, details)
			 SELECT 'UNLINK_KOPERASI', u.id, u.username, ?, 'Unlinked from koperasi member'
			 FROM ecommerce_users u WHERE u.id = ?`,
			fmt.Sprintf("member:%d", oldID), ecUserID).Error
	})
	return oldID, err
}

func (r *pgECommerceRepository) ConvertPointsToSimpanan(ctx context.Context, ecUserID, memberID int, points, rupiah float64) (float64, error) {
	var newSukarela float64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		type userLinkRow struct {
			ID                     int
			LinkedKoperasiMemberID int
		}
		var user userLinkRow
		if e := tx.Raw(
			`SELECT id, COALESCE(linked_koperasi_member_id, 0) AS linked_koperasi_member_id
			 FROM ecommerce_users WHERE id = ? FOR UPDATE`,
			ecUserID).Scan(&user).Error; e != nil {
			return e
		}
		if user.ID == 0 {
			return fmt.Errorf("user tidak ditemukan")
		}
		if user.LinkedKoperasiMemberID == 0 {
			return fmt.Errorf("akun belum di-link dengan member koperasi")
		}
		if user.LinkedKoperasiMemberID != memberID {
			return fmt.Errorf("tautan member koperasi berubah, muat ulang halaman")
		}

		if e := tx.Exec(
			`INSERT INTO user_points (user_id, balance, total_earned, total_redeemed)
			 VALUES (?, 0, 0, 0) ON CONFLICT (user_id) DO NOTHING`,
			ecUserID).Error; e != nil {
			return e
		}
		var balance float64
		if e := tx.Raw(
			"SELECT balance FROM user_points WHERE user_id = ? FOR UPDATE",
			ecUserID).Scan(&balance).Error; e != nil {
			return e
		}
		if balance < points {
			return fmt.Errorf("poin tidak cukup")
		}
		type memberRow struct {
			ID   int
			Nama string
		}
		var member memberRow
		if e := tx.Raw(
			"SELECT id, nama FROM members WHERE id = ? FOR UPDATE",
			memberID).Scan(&member).Error; e != nil {
			return e
		}
		if member.ID == 0 {
			return fmt.Errorf("member koperasi tidak ditemukan")
		}
		if e := tx.Exec(
			"UPDATE user_points SET balance = balance - ?, total_redeemed = total_redeemed + ? WHERE user_id = ?",
			points, points, ecUserID).Error; e != nil {
			return e
		}
		if e := tx.Exec(
			`INSERT INTO points_transactions (user_id, tipe, amount, keterangan)
			 VALUES (?, 'CONVERT_SIMPANAN', ?, ?)`,
			ecUserID, -points,
			fmt.Sprintf("Konversi %.0f poin -> Rp %.0f ke Simpanan Sukarela", points, rupiah)).Error; e != nil {
			return e
		}
		if e := tx.Raw(
			"UPDATE members SET simpanan_sukarela = simpanan_sukarela + ? WHERE id = ? RETURNING simpanan_sukarela",
			rupiah, memberID).Scan(&newSukarela).Error; e != nil {
			return e
		}
		if e := tx.Exec(
			`INSERT INTO simpanan_transactions (member_id, jenis, tipe, nominal, keterangan)
			 VALUES (?, 'SUKARELA', 'MASUK', ?, ?)`,
			memberID, rupiah,
			fmt.Sprintf("Konversi %.0f poin e-commerce", points)).Error; e != nil {
			return e
		}
		return tx.Exec(
			`INSERT INTO journal_entries (tanggal, keterangan, akun_debit, akun_kredit, nominal, tipe_transaksi)
			 VALUES (CURRENT_DATE, ?, 'Beban Reward Poin', 'Simpanan Sukarela', ?, 'SIMPANAN')`,
			fmt.Sprintf("Konversi poin e-commerce - %s", member.Nama), rupiah).Error
	})
	return newSukarela, err
}

// ---- Audit (4d-2) ----

func (r *pgECommerceRepository) LogAudit(ctx context.Context, action string, userID int, username, resource, details string) error {
	return r.db.WithContext(ctx).Exec(
		`INSERT INTO audit_logs (action, user_id, username, resource, details)
		 VALUES (?, ?, ?, ?, ?)`,
		action, userID, username, resource, details).Error
}

func insertAuditLogTx(tx *gorm.DB, audit AuditAction) error {
	return tx.Exec(
		`INSERT INTO audit_logs (action, user_id, username, resource, details)
		 VALUES (?, ?, ?, ?, ?)`,
		audit.Action, audit.UserID, audit.Username, audit.Resource, audit.Details).Error
}

// ---- Dashboard buyer / read (4d-2) ----

// ecOrderRow: DTO datar untuk ec_orders (model.ECOrder punya slice Items yang
// tak bisa di-scan GORM). Nama buyer/seller di-JOIN; kolom teks nullable di-COALESCE.
type ecOrderRow struct {
	ID               int
	NomorOrder       string
	BuyerID          int
	BuyerName        string
	SellerID         int
	SellerName       string
	AlamatPengiriman string
	ShippingOption   string
	ShippingCost     float64
	Subtotal         float64
	Discount         float64
	PointsUsed       float64
	TotalHarga       float64
	VoucherCode      string
	MetodeBayar      string
	Status           string
	ResiPengiriman   string
	PointsEarned     float64
	CreatedAt        string
	UpdatedAt        string
}

func (row ecOrderRow) toModel() model.ECOrder {
	return model.ECOrder{
		ID: row.ID, NomorOrder: row.NomorOrder,
		BuyerID: row.BuyerID, BuyerName: row.BuyerName,
		SellerID: row.SellerID, SellerName: row.SellerName,
		AlamatPengiriman: row.AlamatPengiriman, ShippingOption: row.ShippingOption,
		ShippingCost: row.ShippingCost, Subtotal: row.Subtotal,
		Discount: row.Discount, PointsUsed: row.PointsUsed,
		TotalHarga: row.TotalHarga, VoucherCode: row.VoucherCode,
		MetodeBayar: row.MetodeBayar, Status: row.Status,
		ResiPengiriman: row.ResiPengiriman, PointsEarned: row.PointsEarned,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

const ecOrderSelect = `SELECT o.id, o.nomor_order, o.buyer_id,
	COALESCE(bu.username,'') AS buyer_name, o.seller_id,
	COALESCE(sp.store_name, su.username, '') AS seller_name,
	COALESCE(o.alamat_pengiriman,'') AS alamat_pengiriman,
	COALESCE(o.shipping_option,'') AS shipping_option, o.shipping_cost,
	o.subtotal, o.discount, o.points_used, o.total_harga,
	COALESCE(o.voucher_code,'') AS voucher_code,
	COALESCE(o.metode_bayar,'') AS metode_bayar, o.status,
	COALESCE(o.resi_pengiriman,'') AS resi_pengiriman, o.points_earned,
	to_char(o.created_at,'YYYY-MM-DD HH24:MI') AS created_at,
	to_char(o.updated_at,'YYYY-MM-DD HH24:MI') AS updated_at
	FROM ec_orders o
	LEFT JOIN ecommerce_users bu ON bu.id = o.buyer_id
	LEFT JOIN ecommerce_users su ON su.id = o.seller_id
	LEFT JOIN seller_profiles sp ON sp.seller_id = o.seller_id`

func (r *pgECommerceRepository) OrdersByBuyer(ctx context.Context, buyerID int) ([]model.ECOrder, error) {
	rows := []ecOrderRow{}
	if err := r.db.WithContext(ctx).Raw(
		ecOrderSelect+" WHERE o.buyer_id = ? ORDER BY o.id DESC", buyerID).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model.ECOrder, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toModel())
	}
	return out, nil
}

func (r *pgECommerceRepository) UserPoints(ctx context.Context, userID int) (*model.UserPoints, error) {
	out := []model.UserPoints{}
	if err := r.db.WithContext(ctx).Raw(
		"SELECT user_id, balance, total_earned, total_redeemed FROM user_points WHERE user_id = ? LIMIT 1",
		userID).Scan(&out).Error; err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, nil
	}
	return &out[0], nil
}

func (r *pgECommerceRepository) AddressesByUser(ctx context.Context, userID int) ([]model.ECAddress, error) {
	out := []model.ECAddress{}
	return out, r.db.WithContext(ctx).Raw(
		`SELECT id, user_id, COALESCE(label,'') AS label, COALESCE(penerima,'') AS penerima,
			COALESCE(no_hp,'') AS no_hp, COALESCE(alamat,'') AS alamat, COALESCE(kota,'') AS kota,
			COALESCE(provinsi,'') AS provinsi, COALESCE(kode_pos,'') AS kode_pos, is_default
		 FROM ec_addresses WHERE user_id = ? ORDER BY is_default DESC, id`, userID).
		Scan(&out).Error
}

func (r *pgECommerceRepository) WishlistProducts(ctx context.Context, userID int) ([]model.ECProduct, error) {
	out := []model.ECProduct{}
	cols := `p.id, p.seller_id, p.seller_name, p.nama, p.deskripsi, p.kategori,
		p.harga, p.stok, p.berat, p.foto_url, p.rating, p.total_review, p.total_sold, p.status,
		to_char(p.created_at,'YYYY-MM-DD') AS created_at`
	return out, r.db.WithContext(ctx).Raw(
		"SELECT "+cols+" FROM wishlists w JOIN ec_products p ON p.id = w.product_id WHERE w.user_id = ? ORDER BY w.id DESC",
		userID).Scan(&out).Error
}

// ---- Alur beli / order (4d-3) ----

const shippingCols = `id, nama, COALESCE(provider,'') AS provider,
	COALESCE(estimasi,'') AS estimasi, harga`

func (r *pgECommerceRepository) ShippingOptions(ctx context.Context) ([]model.ShippingOption, error) {
	out := []model.ShippingOption{}
	return out, r.db.WithContext(ctx).Raw(
		"SELECT " + shippingCols + " FROM shipping_options ORDER BY id").Scan(&out).Error
}

func (r *pgECommerceRepository) ShippingOptionByID(ctx context.Context, id int) (*model.ShippingOption, error) {
	out := []model.ShippingOption{}
	if err := r.db.WithContext(ctx).Raw(
		"SELECT "+shippingCols+" FROM shipping_options WHERE id = ? LIMIT 1", id).
		Scan(&out).Error; err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, nil
	}
	return &out[0], nil
}

func (r *pgECommerceRepository) VoucherByCode(ctx context.Context, code string) (*model.Voucher, error) {
	out := []model.Voucher{}
	cols := `id, code, COALESCE(deskripsi,'') AS deskripsi, tipe_diskon, nilai_diskon,
		min_pembelian, maks_diskon, kuota, status,
		to_char(berlaku_sampai,'YYYY-MM-DD') AS berlaku_sampai`
	if err := r.db.WithContext(ctx).Raw(
		"SELECT "+cols+" FROM vouchers WHERE upper(code) = upper(?) LIMIT 1", code).
		Scan(&out).Error; err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, nil
	}
	return &out[0], nil
}

func (r *pgECommerceRepository) CreateOrder(ctx context.Context, o model.ECOrder, pointsUsed, pointsEarned float64) (*model.ECOrder, error) {
	now := time.Now()
	var newID int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := tx.Raw("SELECT nextval(pg_get_serial_sequence('ec_orders','id'))").Scan(&newID).Error; e != nil {
			return e
		}
		o.NomorOrder = fmt.Sprintf("EC-%s-%03d", now.Format("20060102"), newID)

		var voucher interface{}
		if o.VoucherCode != "" {
			type voucherRow struct {
				Code   string
				Status string
				Kuota  int
			}
			var v voucherRow
			if e := tx.Raw(
				`SELECT code, status, kuota FROM vouchers
				 WHERE upper(code) = upper(?) FOR UPDATE`,
				o.VoucherCode).Scan(&v).Error; e != nil {
				return e
			}
			if v.Code == "" || v.Status != "ACTIVE" || v.Kuota <= 0 {
				return fmt.Errorf("voucher tidak tersedia")
			}
			o.VoucherCode = strings.ToUpper(v.Code)
			voucher = o.VoucherCode
		}
		if e := tx.Raw(
			`INSERT INTO ec_orders (id, nomor_order, buyer_id, seller_id, alamat_pengiriman,
				shipping_option, shipping_cost, subtotal, discount, points_used, total_harga,
				voucher_code, metode_bayar, status, resi_pengiriman, points_earned)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
			newID, o.NomorOrder, o.BuyerID, o.SellerID, o.AlamatPengiriman,
			o.ShippingOption, o.ShippingCost, o.Subtotal, o.Discount, pointsUsed, o.TotalHarga,
			voucher, o.MetodeBayar, o.Status, o.ResiPengiriman, pointsEarned).
			Scan(&newID).Error; e != nil {
			return e
		}

		for _, it := range o.Items {
			type productRow struct {
				ID     int
				Nama   string
				Stok   int
				Status string
			}
			var p productRow
			if e := tx.Raw(
				"SELECT id, nama, stok, status FROM ec_products WHERE id = ? FOR UPDATE",
				it.ProductID).Scan(&p).Error; e != nil {
				return e
			}
			if p.ID == 0 {
				return fmt.Errorf("produk tidak ditemukan")
			}
			if p.Status != "APPROVED" {
				return fmt.Errorf("produk %s belum tersedia", p.Nama)
			}
			if p.Stok < it.Jumlah {
				return fmt.Errorf("stok produk %s tidak cukup", p.Nama)
			}
			if e := tx.Exec(
				`INSERT INTO ec_order_items (order_id, product_id, product_nama, seller_id, jumlah, harga_satuan, subtotal)
				 VALUES (?, ?, ?, ?, ?, ?, ?)`,
				newID, it.ProductID, it.ProductNama, it.SellerID, it.Jumlah, it.HargaSatuan, it.Subtotal).Error; e != nil {
				return e
			}
			if e := tx.Exec(
				"UPDATE ec_products SET stok = stok - ?, total_sold = total_sold + ? WHERE id = ?",
				it.Jumlah, it.Jumlah, it.ProductID).Error; e != nil {
				return e
			}
		}

		if o.VoucherCode != "" {
			if e := tx.Exec(
				`UPDATE vouchers
				    SET kuota = kuota - 1,
				        status = CASE WHEN kuota - 1 <= 0 THEN 'USED_UP' ELSE status END
				  WHERE code = ?`,
				o.VoucherCode).Error; e != nil {
				return e
			}
		}

		if pointsUsed > 0 {
			if e := tx.Exec(
				`INSERT INTO user_points (user_id, balance, total_earned, total_redeemed)
				 VALUES (?, 0, 0, 0) ON CONFLICT (user_id) DO NOTHING`,
				o.BuyerID).Error; e != nil {
				return e
			}
			var balance float64
			if e := tx.Raw(
				"SELECT balance FROM user_points WHERE user_id = ? FOR UPDATE",
				o.BuyerID).Scan(&balance).Error; e != nil {
				return e
			}
			if balance < pointsUsed {
				return fmt.Errorf("poin tidak cukup")
			}
			if e := tx.Exec(
				`UPDATE user_points
				    SET balance = balance - ?, total_redeemed = total_redeemed + ?
				  WHERE user_id = ?`,
				pointsUsed, pointsUsed, o.BuyerID).Error; e != nil {
				return e
			}
			if e := tx.Exec(
				`INSERT INTO points_transactions (user_id, tipe, amount, order_id, keterangan)
				 VALUES (?, 'REDEEM_DISCOUNT', ?, ?, ?)`,
				o.BuyerID, -pointsUsed, newID,
				fmt.Sprintf("Redeem %.0f poin untuk order %s", pointsUsed, o.NomorOrder)).Error; e != nil {
				return e
			}
		}
		if pointsEarned > 0 {
			if e := tx.Exec(
				`INSERT INTO user_points (user_id, balance, total_earned, total_redeemed)
				 VALUES (?, ?, ?, 0) ON CONFLICT (user_id)
				 DO UPDATE SET balance = user_points.balance + ?, total_earned = user_points.total_earned + ?`,
				o.BuyerID, pointsEarned, pointsEarned, pointsEarned, pointsEarned).Error; e != nil {
				return e
			}
			if e := tx.Exec(
				`INSERT INTO points_transactions (user_id, tipe, amount, order_id, keterangan)
				 VALUES (?, 'EARN_PURCHASE', ?, ?, ?)`,
				o.BuyerID, pointsEarned, newID, "Pembelian order "+o.NomorOrder).Error; e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	o.ID = newID
	o.PointsUsed = pointsUsed
	o.PointsEarned = pointsEarned
	o.CreatedAt = now.Format("2006-01-02 15:04")
	o.UpdatedAt = o.CreatedAt
	return &o, nil
}

func (r *pgECommerceRepository) OrderByID(ctx context.Context, id int) (*model.ECOrder, error) {
	rows := []ecOrderRow{}
	if err := r.db.WithContext(ctx).Raw(
		ecOrderSelect+" WHERE o.id = ? LIMIT 1", id).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	order := rows[0].toModel()
	items := []model.ECOrderItem{}
	if err := r.db.WithContext(ctx).Raw(
		`SELECT product_id, COALESCE(product_nama,'') AS product_nama, seller_id,
			jumlah, harga_satuan, subtotal
		 FROM ec_order_items WHERE order_id = ? ORDER BY id`, id).
		Scan(&items).Error; err != nil {
		return nil, err
	}
	order.Items = items
	return &order, nil
}

func (r *pgECommerceRepository) ShipmentEvents(ctx context.Context, orderID int) ([]model.ShipmentEvent, error) {
	out := []model.ShipmentEvent{}
	return out, r.db.WithContext(ctx).Raw(
		`SELECT id, order_id, status, COALESCE(lokasi,'') AS lokasi,
			COALESCE(keterangan,'') AS keterangan,
			to_char(created_at,'YYYY-MM-DD HH24:MI') AS created_at
		 FROM shipment_events WHERE order_id = ? ORDER BY id`, orderID).Scan(&out).Error
}

func (r *pgECommerceRepository) ToggleWishlist(ctx context.Context, userID, productID int) (bool, error) {
	var added bool
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var n int64
		if e := tx.Raw(
			"SELECT count(*) FROM wishlists WHERE user_id = ? AND product_id = ?",
			userID, productID).Scan(&n).Error; e != nil {
			return e
		}
		if n > 0 {
			added = false
			return tx.Exec(
				"DELETE FROM wishlists WHERE user_id = ? AND product_id = ?",
				userID, productID).Error
		}
		added = true
		return tx.Exec(
			"INSERT INTO wishlists (user_id, product_id) VALUES (?, ?)",
			userID, productID).Error
	})
	return added, err
}

// ---- Profil & alamat (4d-3) ----

func (r *pgECommerceRepository) UserByUsername(ctx context.Context, username string) (*model.ECommerceUser, error) {
	out := []model.ECommerceUser{}
	if err := r.db.WithContext(ctx).Raw(
		"SELECT "+ecUserCols+" FROM ecommerce_users WHERE username = ? LIMIT 1", username).
		Scan(&out).Error; err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, nil
	}
	return &out[0], nil
}

func (r *pgECommerceRepository) CreateAddress(ctx context.Context, userID int, a model.ECAddress) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var n int64
		if e := tx.Raw("SELECT count(*) FROM ec_addresses WHERE user_id = ?", userID).
			Scan(&n).Error; e != nil {
			return e
		}
		isDefault := n == 0 // alamat pertama jadi default
		return tx.Exec(
			`INSERT INTO ec_addresses (user_id, label, penerima, no_hp, alamat, kota, provinsi, kode_pos, is_default)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			userID, a.Label, a.Penerima, a.NoHP, a.Alamat, a.Kota, a.Provinsi, a.KodePos, isDefault).Error
	})
}

func (r *pgECommerceRepository) SetDefaultAddress(ctx context.Context, userID, addressID int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := tx.Exec(
			"UPDATE ec_addresses SET is_default = false WHERE user_id = ?", userID).Error; e != nil {
			return e
		}
		return tx.Exec(
			"UPDATE ec_addresses SET is_default = true WHERE id = ? AND user_id = ?",
			addressID, userID).Error
	})
}

func (r *pgECommerceRepository) DeleteAddress(ctx context.Context, userID, addressID int) error {
	return r.db.WithContext(ctx).Exec(
		"DELETE FROM ec_addresses WHERE id = ? AND user_id = ?", addressID, userID).Error
}

func (r *pgECommerceRepository) UpdateUserContact(ctx context.Context, userID int, username, email string) error {
	return r.db.WithContext(ctx).Exec(
		"UPDATE ecommerce_users SET username = ?, email = ? WHERE id = ?",
		username, email, userID).Error
}

// ---- Review (4d-6) ----

func (r *pgECommerceRepository) HasReviewed(ctx context.Context, userID, productID int) (bool, error) {
	var n int64
	if err := r.db.WithContext(ctx).Raw(
		"SELECT count(*) FROM product_reviews WHERE user_id = ? AND product_id = ?",
		userID, productID).Scan(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *pgECommerceRepository) SubmitReview(ctx context.Context, userID, productID, rating int, komentar, username string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var reviewID int
		if e := tx.Raw(
			`INSERT INTO product_reviews (product_id, user_id, username, rating, komentar)
			 VALUES (?, ?, ?, ?, ?)
			 ON CONFLICT (user_id, product_id) DO NOTHING
			 RETURNING id`,
			productID, userID, username, rating, komentar).Scan(&reviewID).Error; e != nil {
			return e
		}
		if reviewID == 0 {
			return nil
		}
		// Perbarui agregat rating produk dari seluruh ulasan.
		return tx.Exec(
			`UPDATE ec_products SET
				total_review = (SELECT count(*) FROM product_reviews WHERE product_id = ?),
				rating = COALESCE((SELECT avg(rating) FROM product_reviews WHERE product_id = ?), 0)
			 WHERE id = ?`,
			productID, productID, productID).Error
	})
}

// ---- Admin (4d-6) ----

func (r *pgECommerceRepository) AllProducts(ctx context.Context) ([]model.ECProduct, error) {
	out := []model.ECProduct{}
	return out, r.db.WithContext(ctx).Raw(
		"SELECT " + ecProductCols + " FROM ec_products ORDER BY id DESC").Scan(&out).Error
}

func (r *pgECommerceRepository) AllOrders(ctx context.Context) ([]model.ECOrder, error) {
	rows := []ecOrderRow{}
	if err := r.db.WithContext(ctx).Raw(
		ecOrderSelect + " ORDER BY o.id DESC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model.ECOrder, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toModel())
	}
	return out, nil
}

func (r *pgECommerceRepository) AllUsers(ctx context.Context) ([]model.ECommerceUser, error) {
	out := []model.ECommerceUser{}
	return out, r.db.WithContext(ctx).Raw(
		"SELECT " + ecUserCols + " FROM ecommerce_users ORDER BY id").Scan(&out).Error
}

func (r *pgECommerceRepository) AllVouchers(ctx context.Context) ([]model.Voucher, error) {
	out := []model.Voucher{}
	cols := `id, code, COALESCE(deskripsi,'') AS deskripsi, tipe_diskon, nilai_diskon,
		min_pembelian, maks_diskon, kuota, status,
		to_char(berlaku_sampai,'YYYY-MM-DD') AS berlaku_sampai`
	return out, r.db.WithContext(ctx).Raw(
		"SELECT " + cols + " FROM vouchers ORDER BY id").Scan(&out).Error
}

func (r *pgECommerceRepository) AuditLogs(ctx context.Context) ([]model.AuditLog, error) {
	out := []model.AuditLog{}
	return out, r.db.WithContext(ctx).Raw(
		`SELECT id, action, COALESCE(user_id,0) AS user_id, COALESCE(username,'') AS username,
			COALESCE(resource,'') AS resource, COALESCE(details,'') AS details,
			to_char(created_at,'YYYY-MM-DD HH24:MI') AS created_at
		 FROM audit_logs ORDER BY id DESC`).Scan(&out).Error
}

func (r *pgECommerceRepository) AllPointsTransactions(ctx context.Context) ([]model.PointsTransaction, error) {
	out := []model.PointsTransaction{}
	return out, r.db.WithContext(ctx).Raw(
		`SELECT id, user_id, tipe, amount, COALESCE(order_id,0) AS order_id,
			COALESCE(keterangan,'') AS keterangan,
			to_char(created_at,'YYYY-MM-DD') AS created_at
		 FROM points_transactions ORDER BY id DESC`).Scan(&out).Error
}

func (r *pgECommerceRepository) AllUserPoints(ctx context.Context) ([]model.UserPoints, error) {
	out := []model.UserPoints{}
	return out, r.db.WithContext(ctx).Raw(
		"SELECT user_id, balance, total_earned, total_redeemed FROM user_points ORDER BY user_id").
		Scan(&out).Error
}

func (r *pgECommerceRepository) ActivateSeller(ctx context.Context, userID int, storeName string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return activateSellerTx(tx, userID, storeName)
	})
}

func (r *pgECommerceRepository) ActivateSellerWithAudit(ctx context.Context, userID int, storeName string, audit AuditAction) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := activateSellerTx(tx, userID, storeName); err != nil {
			return err
		}
		return insertAuditLogTx(tx, audit)
	})
}

func activateSellerTx(tx *gorm.DB, userID int, storeName string) error {
	res := tx.Exec(
		"UPDATE ecommerce_users SET is_seller_active = true WHERE id = ? AND is_seller_active = false",
		userID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("user tidak ditemukan atau sudah menjadi seller aktif")
	}
	return tx.Exec(
		`INSERT INTO seller_profiles (seller_id, store_name, description, rating, response_time, total_sold)
		 VALUES (?, ?, 'Toko baru', 0, '-', 0)
		 ON CONFLICT (seller_id) DO NOTHING`,
		userID, storeName).Error
}

func (r *pgECommerceRepository) SetECProductStatus(ctx context.Context, id int, status string) error {
	return r.db.WithContext(ctx).Exec(
		"UPDATE ec_products SET status = ? WHERE id = ?", status, id).Error
}

func (r *pgECommerceRepository) SetECProductStatusWithAudit(ctx context.Context, id int, status, expectedStatus string, audit AuditAction) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var res *gorm.DB
		if expectedStatus != "" {
			res = tx.Exec(
				"UPDATE ec_products SET status = ? WHERE id = ? AND status = ?",
				status, id, expectedStatus)
		} else {
			res = tx.Exec("UPDATE ec_products SET status = ? WHERE id = ?", status, id)
		}
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("produk tidak bisa diubah ke status %s dari status saat ini", status)
		}
		return insertAuditLogTx(tx, audit)
	})
}

func (r *pgECommerceRepository) VoucherByID(ctx context.Context, id int) (*model.Voucher, error) {
	out := []model.Voucher{}
	cols := `id, code, COALESCE(deskripsi,'') AS deskripsi, tipe_diskon, nilai_diskon,
		min_pembelian, maks_diskon, kuota, status,
		to_char(berlaku_sampai,'YYYY-MM-DD') AS berlaku_sampai`
	if err := r.db.WithContext(ctx).Raw(
		"SELECT "+cols+" FROM vouchers WHERE id = ? LIMIT 1", id).Scan(&out).Error; err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, nil
	}
	return &out[0], nil
}

func (r *pgECommerceRepository) CreateVoucher(ctx context.Context, v model.Voucher) (int, error) {
	var id int
	err := r.db.WithContext(ctx).Raw(
		`INSERT INTO vouchers (code, deskripsi, tipe_diskon, nilai_diskon, min_pembelian,
			maks_diskon, kuota, status, berlaku_sampai)
		 VALUES (?, ?, ?, ?, ?, ?, ?, 'ACTIVE', ?) RETURNING id`,
		v.Code, v.Deskripsi, v.TipeDiskon, v.NilaiDiskon, v.MinPembelian,
		v.MaksDiskon, v.Kuota, v.BerlakuSampai).Scan(&id).Error
	return id, err
}

func (r *pgECommerceRepository) CreateVoucherWithAudit(ctx context.Context, v model.Voucher, audit AuditAction) (int, error) {
	var id int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := tx.Raw(
			`INSERT INTO vouchers (code, deskripsi, tipe_diskon, nilai_diskon, min_pembelian,
				maks_diskon, kuota, status, berlaku_sampai)
			 VALUES (?, ?, ?, ?, ?, ?, ?, 'ACTIVE', ?) RETURNING id`,
			v.Code, v.Deskripsi, v.TipeDiskon, v.NilaiDiskon, v.MinPembelian,
			v.MaksDiskon, v.Kuota, v.BerlakuSampai).Scan(&id).Error; e != nil {
			return e
		}
		if audit.Resource == "" {
			audit.Resource = fmt.Sprintf("voucher:%d", id)
		}
		return insertAuditLogTx(tx, audit)
	})
	return id, err
}

func (r *pgECommerceRepository) SetVoucherStatus(ctx context.Context, id int, status string) error {
	return r.db.WithContext(ctx).Exec(
		"UPDATE vouchers SET status = ? WHERE id = ?", status, id).Error
}

func (r *pgECommerceRepository) SetVoucherStatusWithAudit(ctx context.Context, id int, status, expectedStatus string, audit AuditAction) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var res *gorm.DB
		if expectedStatus != "" {
			res = tx.Exec(
				"UPDATE vouchers SET status = ? WHERE id = ? AND status = ?",
				status, id, expectedStatus)
		} else {
			res = tx.Exec("UPDATE vouchers SET status = ? WHERE id = ?", status, id)
		}
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("voucher tidak bisa diubah ke status %s dari status saat ini", status)
		}
		return insertAuditLogTx(tx, audit)
	})
}

// ---- Poin & loyalty (4d-4) ----

func (r *pgECommerceRepository) PointsTransactions(ctx context.Context, userID int) ([]model.PointsTransaction, error) {
	out := []model.PointsTransaction{}
	return out, r.db.WithContext(ctx).Raw(
		`SELECT id, user_id, tipe, amount, COALESCE(order_id, 0) AS order_id,
			COALESCE(keterangan,'') AS keterangan,
			to_char(created_at,'YYYY-MM-DD') AS created_at
		 FROM points_transactions WHERE user_id = ? ORDER BY id DESC`, userID).Scan(&out).Error
}

func (r *pgECommerceRepository) NIKExists(ctx context.Context, nik string) (bool, error) {
	var n int64
	if err := r.db.WithContext(ctx).Raw(
		"SELECT count(*) FROM members WHERE nik = ?", nik).Scan(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

// ---- Seller (4d-5) ----

func (r *pgECommerceRepository) OrdersBySeller(ctx context.Context, sellerID int) ([]model.ECOrder, error) {
	rows := []ecOrderRow{}
	if err := r.db.WithContext(ctx).Raw(
		ecOrderSelect+" WHERE o.seller_id = ? ORDER BY o.id DESC", sellerID).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model.ECOrder, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toModel())
	}
	return out, nil
}

func (r *pgECommerceRepository) CreateECProduct(ctx context.Context, p model.ECProduct) (int, error) {
	if p.Harga <= 0 {
		return 0, fmt.Errorf("harga produk harus lebih dari 0")
	}
	if p.Stok < 0 {
		return 0, fmt.Errorf("stok produk tidak boleh negatif")
	}
	if p.Berat <= 0 {
		return 0, fmt.Errorf("berat produk harus lebih dari 0")
	}
	var id int
	err := r.db.WithContext(ctx).Raw(
		`INSERT INTO ec_products (seller_id, seller_name, nama, deskripsi, kategori,
			harga, stok, berat, foto_url, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'PENDING_APPROVAL') RETURNING id`,
		p.SellerID, p.SellerName, p.Nama, p.Deskripsi, p.Kategori,
		p.Harga, p.Stok, p.Berat, p.FotoURL).Scan(&id).Error
	return id, err
}

func (r *pgECommerceRepository) MarkOrderShipped(ctx context.Context, sellerID, orderID int, resi, lokasi, keterangan string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Exec(
			`UPDATE ec_orders
			    SET status = 'DIKIRIM', resi_pengiriman = ?, updated_at = now()
			  WHERE id = ? AND seller_id = ? AND status IN ('DIBAYAR', 'DIPROSES')`,
			resi, orderID, sellerID)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("pesanan tidak bisa dikirim dari status saat ini")
		}
		return tx.Exec(
			`INSERT INTO shipment_events (order_id, status, lokasi, keterangan)
			 VALUES (?, 'DIKIRIM', ?, ?)`,
			orderID, lokasi, keterangan).Error
	})
}

func (r *pgECommerceRepository) RegisterMemberAndLink(ctx context.Context, ecUserID int, nama, nik, alamat, noHP string) (int, error) {
	var memberID int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		type userLinkRow struct {
			ID                     int
			LinkedKoperasiMemberID int
		}
		var user userLinkRow
		if e := tx.Raw(
			`SELECT id, COALESCE(linked_koperasi_member_id, 0) AS linked_koperasi_member_id
			 FROM ecommerce_users WHERE id = ? FOR UPDATE`,
			ecUserID).Scan(&user).Error; e != nil {
			return e
		}
		if user.ID == 0 {
			return fmt.Errorf("User tidak ditemukan.")
		}
		if user.LinkedKoperasiMemberID != 0 {
			return fmt.Errorf("Akun Anda sudah terhubung ke member koperasi.")
		}

		if e := tx.Raw(
			`INSERT INTO members (nomor_anggota, nama, nik, alamat, no_hp, status, tanggal_masuk)
			 VALUES ('-', ?, ?, ?, ?, 'PENDING', CURRENT_DATE) RETURNING id`,
			nama, nik, alamat, noHP).Scan(&memberID).Error; e != nil {
			return e
		}
		if e := tx.Exec(
			"UPDATE ecommerce_users SET linked_koperasi_member_id = ? WHERE id = ?",
			memberID, ecUserID).Error; e != nil {
			return e
		}
		return tx.Exec(
			`INSERT INTO audit_logs (action, user_id, username, resource, details)
			 SELECT 'REGISTER_KOPERASI', u.id, u.username, ?, ?
			 FROM ecommerce_users u WHERE u.id = ?`,
			fmt.Sprintf("member:%d", memberID),
			fmt.Sprintf("Registered & linked new pending member %s", nama), ecUserID).Error
	})
	return memberID, err
}
