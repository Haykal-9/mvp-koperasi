package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
)

// ECAdminService: aturan bisnis sisi admin/pengurus E-Commerce (moderasi,
// analytics, voucher, monitoring). Memakai ECommerceRepository.
type ECAdminService struct {
	repo repository.ECommerceRepository
}

func NewECAdminService(repo repository.ECommerceRepository) *ECAdminService {
	return &ECAdminService{repo: repo}
}

// AdminDashboard menampung KPI dashboard admin.
type AdminDashboard struct {
	PendingProducts int
	TotalOrders     int
	CompletedOrders int
	TotalRevenue    float64
	TotalUsers      int
	ActiveSellers   int
	ActiveVouchers  int
	RecentActivity  []model.AuditLog
}

func (s *ECAdminService) Dashboard(ctx context.Context) (*AdminDashboard, error) {
	products, err := s.repo.AllProducts(ctx)
	if err != nil {
		return nil, err
	}
	orders, err := s.repo.AllOrders(ctx)
	if err != nil {
		return nil, err
	}
	users, err := s.repo.AllUsers(ctx)
	if err != nil {
		return nil, err
	}
	vouchers, err := s.repo.AllVouchers(ctx)
	if err != nil {
		return nil, err
	}
	logs, err := s.repo.AuditLogs(ctx)
	if err != nil {
		return nil, err
	}

	d := &AdminDashboard{TotalOrders: len(orders), TotalUsers: len(users)}
	for _, p := range products {
		if p.Status == "PENDING_APPROVAL" {
			d.PendingProducts++
		}
	}
	for _, o := range orders {
		if o.Status == "SELESAI" {
			d.TotalRevenue += o.TotalHarga
			d.CompletedOrders++
		}
	}
	for _, u := range users {
		if u.IsSellerActive {
			d.ActiveSellers++
		}
	}
	for _, v := range vouchers {
		if v.Status == "ACTIVE" {
			d.ActiveVouchers++
		}
	}
	d.RecentActivity = logs
	if len(d.RecentActivity) > 5 {
		d.RecentActivity = d.RecentActivity[:5]
	}
	return d, nil
}

// SellerStat: performa seller untuk analytics.
type SellerStat struct {
	Profile    model.SellerProfile
	Revenue    float64
	OrderCount int
}

// AdminAnalytics menampung data halaman analytics.
type AdminAnalytics struct {
	SellerStats  []SellerStat
	TopProducts  []model.ECProduct
	StatusCount  map[string]int
	TotalRevenue float64
	TotalOrders  int
}

func (s *ECAdminService) Analytics(ctx context.Context) (*AdminAnalytics, error) {
	orders, err := s.repo.AllOrders(ctx)
	if err != nil {
		return nil, err
	}
	products, err := s.repo.AllProducts(ctx)
	if err != nil {
		return nil, err
	}
	sellers, err := s.repo.SellerProfiles(ctx)
	if err != nil {
		return nil, err
	}

	a := &AdminAnalytics{StatusCount: map[string]int{}, TotalOrders: len(orders)}
	sellerRevenue := map[int]float64{}
	sellerOrderCount := map[int]int{}
	for _, o := range orders {
		a.StatusCount[o.Status]++
		if o.Status == "SELESAI" {
			a.TotalRevenue += o.TotalHarga
			sellerRevenue[o.SellerID] += o.Subtotal
			sellerOrderCount[o.SellerID]++
		}
	}
	for _, sp := range sellers {
		a.SellerStats = append(a.SellerStats, SellerStat{
			Profile:    sp,
			Revenue:    sellerRevenue[sp.SellerID],
			OrderCount: sellerOrderCount[sp.SellerID],
		})
	}
	for _, p := range products {
		if p.Status == "APPROVED" {
			a.TopProducts = append(a.TopProducts, p)
		}
	}
	// Urutkan TopProducts berdasar TotalSold menurun, ambil 5 teratas.
	for i := 0; i < len(a.TopProducts)-1; i++ {
		for j := i + 1; j < len(a.TopProducts); j++ {
			if a.TopProducts[j].TotalSold > a.TopProducts[i].TotalSold {
				a.TopProducts[i], a.TopProducts[j] = a.TopProducts[j], a.TopProducts[i]
			}
		}
	}
	if len(a.TopProducts) > 5 {
		a.TopProducts = a.TopProducts[:5]
	}
	return a, nil
}

// SellerApprovals memisahkan user non-seller dan seller aktif (ADMIN/KASIR dikecualikan).
func (s *ECAdminService) SellerApprovals(ctx context.Context) (nonSellers, activeSellers []model.ECommerceUser, err error) {
	users, err := s.repo.AllUsers(ctx)
	if err != nil {
		return nil, nil, err
	}
	for _, u := range users {
		if u.Role == "ADMIN" || u.Role == "KASIR" {
			continue
		}
		if u.IsSellerActive {
			activeSellers = append(activeSellers, u)
		} else {
			nonSellers = append(nonSellers, u)
		}
	}
	return nonSellers, activeSellers, nil
}

// ApproveSeller mengaktifkan akun seller; mengembalikan username untuk flash.
func (s *ECAdminService) ApproveSeller(ctx context.Context, adminID int, adminUsername string, sellerID int) (string, error) {
	u, err := s.repo.UserByID(ctx, sellerID)
	if err != nil {
		return "", err
	}
	if u == nil {
		return "", fmt.Errorf("User tidak ditemukan.")
	}
	if u.IsSellerActive {
		return "", fmt.Errorf("User sudah menjadi seller aktif.")
	}
	if err := s.repo.ActivateSeller(ctx, sellerID, "Toko "+u.Username); err != nil {
		return "", err
	}
	if err := s.repo.LogAudit(ctx, "APPROVE_SELLER", adminID, adminUsername,
		fmt.Sprintf("seller:%d", sellerID),
		fmt.Sprintf("Activated seller account for user: %s (%s)", u.Username, u.Email)); err != nil {
		return "", err
	}
	return u.Username, nil
}

// RejectSeller mencatat penolakan pengajuan seller; mengembalikan username.
func (s *ECAdminService) RejectSeller(ctx context.Context, adminID int, adminUsername string, sellerID int, alasan string) (string, error) {
	u, err := s.repo.UserByID(ctx, sellerID)
	if err != nil {
		return "", err
	}
	if u == nil {
		return "", fmt.Errorf("User tidak ditemukan.")
	}
	detail := fmt.Sprintf("Rejected seller application for user: %s", u.Username)
	if alasan != "" {
		detail += ". Alasan: " + alasan
	}
	if err := s.repo.LogAudit(ctx, "REJECT_SELLER", adminID, adminUsername,
		fmt.Sprintf("seller:%d", sellerID), detail); err != nil {
		return "", err
	}
	return u.Username, nil
}

// ProductApprovals mengembalikan produk PENDING dan seluruh produk.
func (s *ECAdminService) ProductApprovals(ctx context.Context) (pending, all []model.ECProduct, err error) {
	all, err = s.repo.AllProducts(ctx)
	if err != nil {
		return nil, nil, err
	}
	for _, p := range all {
		if p.Status == "PENDING_APPROVAL" {
			pending = append(pending, p)
		}
	}
	return pending, all, nil
}

func (s *ECAdminService) ApproveProduct(ctx context.Context, adminID int, adminUsername string, productID int) (*model.ECProduct, error) {
	p, err := s.repo.ProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, fmt.Errorf("Produk tidak ditemukan.")
	}
	if p.Status != "PENDING_APPROVAL" {
		return nil, fmt.Errorf("Produk tidak dalam status PENDING.")
	}
	if err := s.repo.SetECProductStatus(ctx, productID, "APPROVED"); err != nil {
		return nil, err
	}
	if err := s.repo.LogAudit(ctx, "APPROVE_PRODUCT", adminID, adminUsername,
		fmt.Sprintf("product:%d", productID),
		fmt.Sprintf("Approved: %s (seller: %s)", p.Nama, p.SellerName)); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ECAdminService) RejectProduct(ctx context.Context, adminID int, adminUsername string, productID int, alasan string) (*model.ECProduct, error) {
	p, err := s.repo.ProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, fmt.Errorf("Produk tidak ditemukan.")
	}
	if err := s.repo.SetECProductStatus(ctx, productID, "REJECTED"); err != nil {
		return nil, err
	}
	detail := fmt.Sprintf("Rejected: %s (seller: %s)", p.Nama, p.SellerName)
	if alasan != "" {
		detail += ". Alasan: " + alasan
	}
	if err := s.repo.LogAudit(ctx, "REJECT_PRODUCT", adminID, adminUsername,
		fmt.Sprintf("product:%d", productID), detail); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ECAdminService) Vouchers(ctx context.Context) ([]model.Voucher, error) {
	return s.repo.AllVouchers(ctx)
}

// CreateVoucher membuat voucher baru (validasi keunikan kode + default kuota/tanggal).
func (s *ECAdminService) CreateVoucher(ctx context.Context, adminID int, adminUsername, code, deskripsi, tipeDiskon string, nilai, min, maks float64, kuota int, berlakuSampai string) error {
	code = strings.ToUpper(strings.TrimSpace(code))
	existing, err := s.repo.VoucherByCode(ctx, code)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("Kode voucher %s sudah digunakan.", code)
	}
	if kuota <= 0 {
		kuota = 100
	}
	if berlakuSampai == "" {
		berlakuSampai = time.Now().AddDate(0, 3, 0).Format("2006-01-02")
	}
	v := model.Voucher{
		Code: code, Deskripsi: deskripsi, TipeDiskon: tipeDiskon, NilaiDiskon: nilai,
		MinPembelian: min, MaksDiskon: maks, Kuota: kuota, BerlakuSampai: berlakuSampai,
	}
	id, err := s.repo.CreateVoucher(ctx, v)
	if err != nil {
		return err
	}
	if err := s.repo.LogAudit(ctx, "CREATE_VOUCHER", adminID, adminUsername,
		fmt.Sprintf("voucher:%d", id), fmt.Sprintf("Created voucher %s: %s", code, deskripsi)); err != nil {
		return err
	}
	return nil
}

// ToggleVoucher membalik status voucher ACTIVE<->EXPIRED; mengembalikan kode & status baru.
func (s *ECAdminService) ToggleVoucher(ctx context.Context, adminID int, adminUsername string, voucherID int) (code, newStatus string, err error) {
	v, err := s.repo.VoucherByID(ctx, voucherID)
	if err != nil {
		return "", "", err
	}
	if v == nil {
		return "", "", fmt.Errorf("Voucher tidak ditemukan.")
	}
	if v.Status == "ACTIVE" {
		newStatus = "EXPIRED"
	} else {
		newStatus = "ACTIVE"
	}
	if err := s.repo.SetVoucherStatus(ctx, voucherID, newStatus); err != nil {
		return "", "", err
	}
	action := "ACTIVATE_VOUCHER"
	if newStatus == "EXPIRED" {
		action = "DEACTIVATE_VOUCHER"
	}
	if err := s.repo.LogAudit(ctx, action, adminID, adminUsername,
		fmt.Sprintf("voucher:%d", voucherID), fmt.Sprintf("%s voucher %s", action, v.Code)); err != nil {
		return "", "", err
	}
	return v.Code, newStatus, nil
}

// Orders mengembalikan order (terfilter status opsional) + jumlah per status.
func (s *ECAdminService) Orders(ctx context.Context, statusFilter string) (orders []model.ECOrder, counts map[string]int, err error) {
	all, err := s.repo.AllOrders(ctx)
	if err != nil {
		return nil, nil, err
	}
	counts = map[string]int{}
	for _, o := range all {
		counts[o.Status]++
	}
	if statusFilter == "" {
		return all, counts, nil
	}
	for _, o := range all {
		if o.Status == statusFilter {
			orders = append(orders, o)
		}
	}
	return orders, counts, nil
}

// Members mengembalikan user terfilter role + total user.
func (s *ECAdminService) Members(ctx context.Context, roleFilter string) (users []model.ECommerceUser, total int, err error) {
	all, err := s.repo.AllUsers(ctx)
	if err != nil {
		return nil, 0, err
	}
	for _, u := range all {
		if roleFilter == "" || u.Role == roleFilter || (roleFilter == "SELLER" && u.IsSellerActive) {
			users = append(users, u)
		}
	}
	return users, len(all), nil
}

// PointsMonitoring mengembalikan seluruh transaksi poin, saldo user, & total poin beredar.
func (s *ECAdminService) PointsMonitoring(ctx context.Context) (tx []model.PointsTransaction, userPoints []model.UserPoints, totalPoints float64, err error) {
	tx, err = s.repo.AllPointsTransactions(ctx)
	if err != nil {
		return nil, nil, 0, err
	}
	userPoints, err = s.repo.AllUserPoints(ctx)
	if err != nil {
		return nil, nil, 0, err
	}
	for _, up := range userPoints {
		totalPoints += up.Balance
	}
	return tx, userPoints, totalPoints, nil
}

// AuditLogs mengembalikan log audit (terfilter action opsional, substring).
func (s *ECAdminService) AuditLogs(ctx context.Context, actionFilter string) ([]model.AuditLog, error) {
	logs, err := s.repo.AuditLogs(ctx)
	if err != nil {
		return nil, err
	}
	if actionFilter == "" {
		return logs, nil
	}
	out := make([]model.AuditLog, 0, len(logs))
	for _, l := range logs {
		if strings.Contains(l.Action, actionFilter) {
			out = append(out, l)
		}
	}
	return out, nil
}
