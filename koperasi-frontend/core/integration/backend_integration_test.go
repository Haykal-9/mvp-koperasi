package integration

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"koperasi-frontend/core/database"
	"koperasi-frontend/core/repository"
	"koperasi-frontend/core/service"
)

func TestPostgresCheckoutPersistsCriticalSideEffects(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := repository.NewECommerceRepository(db)
	shop := service.NewECShopService(repo)

	beforeStock := queryInt(t, db, "SELECT stok FROM ec_products WHERE id = ?", 6)
	beforeSold := queryInt(t, db, "SELECT total_sold FROM ec_products WHERE id = ?", 6)
	beforeVoucher := queryInt(t, db, "SELECT kuota FROM vouchers WHERE code = ?", "HEMAT10")
	beforePoints := queryFloat(t, db, "SELECT balance FROM user_points WHERE user_id = ?", 3)

	order, err := shop.PlaceOrder(ctx, 3, 6, 3, "Jl. Test Integrasi", 3, "hemat10", "QRIS", 50)
	if err != nil {
		t.Fatal(err)
	}
	if order.ID == 0 || order.Status != "DIBAYAR" {
		t.Fatalf("order tidak tersimpan dengan status benar: %+v", order)
	}
	assertFloat(t, order.Subtotal, 54_000, "subtotal")
	assertFloat(t, order.Discount, 5_400, "discount")
	assertFloat(t, order.ShippingCost, 10_000, "shipping")
	assertFloat(t, order.TotalHarga, 58_550, "total")
	assertFloat(t, order.PointsUsed, 50, "points used")
	assertFloat(t, order.PointsEarned, 54, "points earned")
	if order.VoucherCode != "HEMAT10" {
		t.Fatalf("voucher code = %q, mau HEMAT10", order.VoucherCode)
	}

	afterStock := queryInt(t, db, "SELECT stok FROM ec_products WHERE id = ?", 6)
	afterSold := queryInt(t, db, "SELECT total_sold FROM ec_products WHERE id = ?", 6)
	afterVoucher := queryInt(t, db, "SELECT kuota FROM vouchers WHERE code = ?", "HEMAT10")
	afterPoints := queryFloat(t, db, "SELECT balance FROM user_points WHERE user_id = ?", 3)

	if afterStock != beforeStock-3 {
		t.Fatalf("stok produk = %d, mau %d", afterStock, beforeStock-3)
	}
	if afterSold != beforeSold+3 {
		t.Fatalf("total sold = %d, mau %d", afterSold, beforeSold+3)
	}
	if afterVoucher != beforeVoucher-1 {
		t.Fatalf("kuota voucher = %d, mau %d", afterVoucher, beforeVoucher-1)
	}
	assertFloat(t, afterPoints, beforePoints-50+54, "saldo poin")

	earnTx := queryInt(t, db, "SELECT count(*) FROM points_transactions WHERE order_id = ? AND tipe = ?", order.ID, "EARN_PURCHASE")
	redeemTx := queryInt(t, db, "SELECT count(*) FROM points_transactions WHERE order_id = ? AND tipe = ?", order.ID, "REDEEM_DISCOUNT")
	if earnTx != 1 || redeemTx != 1 {
		t.Fatalf("points tx earn=%d redeem=%d, mau 1 dan 1", earnTx, redeemTx)
	}
}

func TestPostgresMemberRepositoryRejectsSimpananOverdraw(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := repository.NewMemberRepository(db)

	before := queryFloat(t, db, "SELECT simpanan_sukarela FROM members WHERE id = ?", 5)
	beforeTx := queryInt(t, db, "SELECT count(*) FROM simpanan_transactions WHERE member_id = ? AND jenis = ?", 5, "SUKARELA")
	beforeJournal := queryInt(t, db, "SELECT count(*) FROM journal_entries WHERE tipe_transaksi = ?", "SIMPANAN")

	err := repo.RecordSimpanan(ctx, 5, "SUKARELA", "KELUAR", 1, "uji saldo negatif")
	if err == nil || !strings.Contains(err.Error(), "tidak mencukupi") {
		t.Fatalf("err = %v, mau saldo tidak mencukupi", err)
	}

	after := queryFloat(t, db, "SELECT simpanan_sukarela FROM members WHERE id = ?", 5)
	afterTx := queryInt(t, db, "SELECT count(*) FROM simpanan_transactions WHERE member_id = ? AND jenis = ?", 5, "SUKARELA")
	afterJournal := queryInt(t, db, "SELECT count(*) FROM journal_entries WHERE tipe_transaksi = ?", "SIMPANAN")
	assertFloat(t, after, before, "saldo sukarela setelah gagal")
	if afterTx != beforeTx {
		t.Fatalf("transaksi simpanan berubah setelah gagal: %d -> %d", beforeTx, afterTx)
	}
	if afterJournal != beforeJournal {
		t.Fatalf("jurnal SIMPANAN berubah setelah gagal: %d -> %d", beforeJournal, afterJournal)
	}
}

func TestPostgresLoanPaymentPersistsInstallmentAndJournal(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := repository.NewLoanRepository(db)
	loans := service.NewLoanService(repo)

	loan, err := loans.FindByID(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if loan == nil {
		t.Fatal("seed loan id=1 tidak ditemukan")
	}
	beforeJournal := queryInt(t, db, "SELECT count(*) FROM journal_entries WHERE tipe_transaksi = ?", "PINJAMAN")

	inst, lunas, err := loans.Pay(ctx, loan, "2026-06-16")
	if err != nil {
		t.Fatal(err)
	}
	if lunas {
		t.Fatal("loan seed id=1 belum boleh lunas setelah satu pembayaran")
	}
	if inst.BulanKe != 4 || inst.Status != "DIBAYAR" {
		t.Fatalf("angsuran terbayar salah: %+v", inst)
	}

	after, err := loans.FindByID(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	assertFloat(t, after.SisaPokok, 2_500_000, "sisa pokok")
	if after.Status != "AKTIF" {
		t.Fatalf("status loan = %q, mau AKTIF", after.Status)
	}
	status := queryString(t, db, "SELECT status FROM installments WHERE loan_id = ? AND bulan_ke = ?", 1, 4)
	paidAt := queryString(t, db, "SELECT to_char(tanggal_bayar,'YYYY-MM-DD') FROM installments WHERE loan_id = ? AND bulan_ke = ?", 1, 4)
	if status != "DIBAYAR" || paidAt != "2026-06-16" {
		t.Fatalf("angsuran bulan 4 status=%q paidAt=%q", status, paidAt)
	}
	afterJournal := queryInt(t, db, "SELECT count(*) FROM journal_entries WHERE tipe_transaksi = ?", "PINJAMAN")
	if afterJournal != beforeJournal+1 {
		t.Fatalf("jumlah jurnal PINJAMAN = %d, mau %d", afterJournal, beforeJournal+1)
	}
}

func TestPostgresConvertPointsToSimpananPersistsAllLedgers(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := repository.NewECommerceRepository(db)
	accounts := service.NewECAccountService(repo)

	beforePoints := queryFloat(t, db, "SELECT balance FROM user_points WHERE user_id = ?", 1)
	beforeRedeemed := queryFloat(t, db, "SELECT total_redeemed FROM user_points WHERE user_id = ?", 1)
	beforeSukarela := queryFloat(t, db, "SELECT simpanan_sukarela FROM members WHERE id = ?", 1)
	beforeSimpananTx := queryInt(t, db, "SELECT count(*) FROM simpanan_transactions WHERE member_id = ? AND jenis = ?", 1, "SUKARELA")
	beforeJournal := queryInt(t, db, "SELECT count(*) FROM journal_entries WHERE tipe_transaksi = ?", "SIMPANAN")

	msg, rupiah, newSukarela, err := accounts.ConvertToSimpanan(ctx, 1, 250)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msg, "Rp 250") {
		t.Fatalf("pesan sukses tidak memuat nilai konversi baku: %q", msg)
	}
	assertFloat(t, rupiah, 250, "rupiah")
	assertFloat(t, newSukarela, beforeSukarela+250, "return simpanan sukarela")

	afterPoints := queryFloat(t, db, "SELECT balance FROM user_points WHERE user_id = ?", 1)
	afterRedeemed := queryFloat(t, db, "SELECT total_redeemed FROM user_points WHERE user_id = ?", 1)
	afterSukarela := queryFloat(t, db, "SELECT simpanan_sukarela FROM members WHERE id = ?", 1)
	afterSimpananTx := queryInt(t, db, "SELECT count(*) FROM simpanan_transactions WHERE member_id = ? AND jenis = ?", 1, "SUKARELA")
	afterJournal := queryInt(t, db, "SELECT count(*) FROM journal_entries WHERE tipe_transaksi = ?", "SIMPANAN")
	convertTx := queryInt(t, db, "SELECT count(*) FROM points_transactions WHERE user_id = ? AND tipe = ? AND amount = ?", 1, "CONVERT_SIMPANAN", -250)

	assertFloat(t, afterPoints, beforePoints-250, "saldo poin")
	assertFloat(t, afterRedeemed, beforeRedeemed+250, "total redeemed")
	assertFloat(t, afterSukarela, beforeSukarela+250, "simpanan sukarela")
	if afterSimpananTx != beforeSimpananTx+1 {
		t.Fatalf("jumlah transaksi simpanan = %d, mau %d", afterSimpananTx, beforeSimpananTx+1)
	}
	if afterJournal != beforeJournal+1 {
		t.Fatalf("jumlah jurnal SIMPANAN = %d, mau %d", afterJournal, beforeJournal+1)
	}
	if convertTx != 1 {
		t.Fatalf("points transaction CONVERT_SIMPANAN = %d, mau 1", convertTx)
	}
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	if testing.Short() {
		t.Skip("skip integration test pada mode -short")
	}
	rawURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if rawURL == "" {
		t.Skip("set TEST_DATABASE_URL untuk menjalankan integration test PostgreSQL")
	}

	adminDB, err := database.Connect(rawURL)
	if err != nil {
		t.Fatalf("connect TEST_DATABASE_URL: %v", err)
	}
	adminSQL, err := adminDB.DB()
	if err != nil {
		t.Fatalf("admin sql db: %v", err)
	}

	schema := fmt.Sprintf("test_koperasi_%d", time.Now().UnixNano())
	if err := adminDB.Exec("CREATE SCHEMA " + quoteIdent(schema)).Error; err != nil {
		_ = adminSQL.Close()
		t.Fatalf("create schema %s: %v", schema, err)
	}
	t.Cleanup(func() {
		if err := adminDB.Exec("DROP SCHEMA IF EXISTS " + quoteIdent(schema) + " CASCADE").Error; err != nil {
			t.Logf("cleanup drop schema %s: %v", schema, err)
		}
		_ = adminSQL.Close()
	})

	testURL := databaseURLWithSearchPath(t, rawURL, schema)
	if err := database.Migrate(testURL); err != nil {
		t.Fatalf("migrate test schema %s: %v", schema, err)
	}
	db, err := database.Connect(testURL)
	if err != nil {
		t.Fatalf("connect migrated test schema %s: %v", schema, err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

func databaseURLWithSearchPath(t *testing.T, rawURL, schema string) string {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		t.Fatalf("TEST_DATABASE_URL harus berbentuk URL PostgreSQL valid: %v", err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	return u.String()
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func queryInt(t *testing.T, db *gorm.DB, sql string, args ...any) int {
	t.Helper()
	var out int
	if err := db.Raw(sql, args...).Scan(&out).Error; err != nil {
		t.Fatalf("query int %q: %v", sql, err)
	}
	return out
}

func queryFloat(t *testing.T, db *gorm.DB, sql string, args ...any) float64 {
	t.Helper()
	var out float64
	if err := db.Raw(sql, args...).Scan(&out).Error; err != nil {
		t.Fatalf("query float %q: %v", sql, err)
	}
	return out
}

func queryString(t *testing.T, db *gorm.DB, sql string, args ...any) string {
	t.Helper()
	var out string
	if err := db.Raw(sql, args...).Scan(&out).Error; err != nil {
		t.Fatalf("query string %q: %v", sql, err)
	}
	return out
}

func assertFloat(t *testing.T, got, want float64, label string) {
	t.Helper()
	if math.Abs(got-want) > 0.001 {
		t.Fatalf("%s = %.2f, mau %.2f", label, got, want)
	}
}
