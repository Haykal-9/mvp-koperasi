package service

import (
	"context"
	"fmt"
	"strings"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
)

// SellerCommissionRate: komisi koperasi atas penjualan seller (3%).
const SellerCommissionRate = 0.03

// ECSellerService: aturan bisnis sisi seller (dashboard, produk, order, earnings).
type ECSellerService struct {
	repo repository.ECommerceRepository
}

func NewECSellerService(repo repository.ECommerceRepository) *ECSellerService {
	return &ECSellerService{repo: repo}
}

// SellerDashboard menampung ringkasan dashboard seller.
type SellerDashboard struct {
	Profile         *model.SellerProfile
	TotalProducts   int
	TotalRevenue    float64
	PendingOrders   int
	CompletedOrders int
	LowStock        []model.ECProduct
	RecentOrders    []model.ECOrder
	PendingProducts int
}

func (s *ECSellerService) Dashboard(ctx context.Context, sellerID int) (*SellerDashboard, error) {
	profile, err := s.repo.SellerProfile(ctx, sellerID)
	if err != nil {
		return nil, err
	}
	products, err := s.repo.ProductsBySeller(ctx, sellerID)
	if err != nil {
		return nil, err
	}
	orders, err := s.repo.OrdersBySeller(ctx, sellerID)
	if err != nil {
		return nil, err
	}

	d := &SellerDashboard{Profile: profile, TotalProducts: len(products)}
	for _, o := range orders {
		if o.Status == "SELESAI" {
			d.TotalRevenue += o.Subtotal
			d.CompletedOrders++
		}
		if o.Status == "DIBAYAR" || o.Status == "DIPROSES" {
			d.PendingOrders++
		}
	}
	for _, p := range products {
		if p.Stok <= 5 && p.Status == "APPROVED" {
			d.LowStock = append(d.LowStock, p)
		}
		if p.Status == "PENDING_APPROVAL" {
			d.PendingProducts++
		}
	}
	d.RecentOrders = orders
	if len(d.RecentOrders) > 5 {
		d.RecentOrders = d.RecentOrders[:5]
	}
	return d, nil
}

func (s *ECSellerService) Products(ctx context.Context, sellerID int, statusFilter string) ([]model.ECProduct, error) {
	products, err := s.repo.ProductsBySeller(ctx, sellerID)
	if err != nil {
		return nil, err
	}
	if statusFilter == "" {
		return products, nil
	}
	out := make([]model.ECProduct, 0, len(products))
	for _, p := range products {
		if p.Status == statusFilter {
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *ECSellerService) Categories(ctx context.Context) ([]string, error) {
	return s.repo.Categories(ctx)
}

// CreateProduct membuat produk baru milik seller (status PENDING_APPROVAL).
func (s *ECSellerService) CreateProduct(ctx context.Context, sellerID int, nama, deskripsi, kategori string, harga float64, stok, berat int) error {
	nama = strings.TrimSpace(nama)
	kategori = strings.TrimSpace(kategori)
	if nama == "" {
		return fmt.Errorf("nama produk wajib diisi")
	}
	if kategori == "" {
		return fmt.Errorf("kategori produk wajib diisi")
	}
	if harga <= 0 {
		return fmt.Errorf("harga produk harus lebih dari 0")
	}
	if stok < 0 {
		return fmt.Errorf("stok produk tidak boleh negatif")
	}
	if berat <= 0 {
		return fmt.Errorf("berat produk harus lebih dari 0")
	}

	sellerName := "Seller"
	if sp, err := s.repo.SellerProfile(ctx, sellerID); err != nil {
		return err
	} else if sp != nil {
		sellerName = sp.StoreName
	}
	p := model.ECProduct{
		SellerID:   sellerID,
		SellerName: sellerName,
		Nama:       nama,
		Deskripsi:  deskripsi,
		Kategori:   kategori,
		Harga:      harga,
		Stok:       stok,
		Berat:      berat,
		FotoURL:    fmt.Sprintf("https://placehold.co/400x400/2d1b69/e2e8f0?text=%s", strings.ReplaceAll(nama, " ", "+")),
	}
	_, err := s.repo.CreateECProduct(ctx, p)
	return err
}

func (s *ECSellerService) Orders(ctx context.Context, sellerID int, statusFilter string) ([]model.ECOrder, error) {
	orders, err := s.repo.OrdersBySeller(ctx, sellerID)
	if err != nil {
		return nil, err
	}
	if statusFilter == "" {
		return orders, nil
	}
	out := make([]model.ECOrder, 0, len(orders))
	for _, o := range orders {
		if o.Status == statusFilter {
			out = append(out, o)
		}
	}
	return out, nil
}

// MarkShipped memvalidasi kepemilikan & status lalu menandai order DIKIRIM.
func (s *ECSellerService) MarkShipped(ctx context.Context, sellerID, orderID int, resi string) (*model.ECOrder, error) {
	order, err := s.repo.OrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil || order.SellerID != sellerID {
		return nil, fmt.Errorf("Pesanan tidak ditemukan.")
	}
	if order.Status != "DIBAYAR" && order.Status != "DIPROSES" {
		return nil, fmt.Errorf("Pesanan tidak bisa dikirim (status: %s).", order.Status)
	}

	sellerName := "Seller"
	if sp, err := s.repo.SellerProfile(ctx, sellerID); err == nil && sp != nil {
		sellerName = sp.StoreName
	}
	if err := s.repo.MarkOrderShipped(ctx, sellerID, orderID, resi,
		"Gudang "+sellerName, "Paket diserahkan ke kurir. Resi: "+resi); err != nil {
		return nil, err
	}
	order.Status = "DIKIRIM"
	order.ResiPengiriman = resi
	return order, nil
}

// SellerEarnings menampung rincian pendapatan seller.
type SellerEarnings struct {
	TotalRevenue    float64
	TotalCommission float64
	NetEarnings     float64
	CompletedOrders int
	Orders          []model.ECOrder
}

func (s *ECSellerService) Earnings(ctx context.Context, sellerID int) (*SellerEarnings, error) {
	orders, err := s.repo.OrdersBySeller(ctx, sellerID)
	if err != nil {
		return nil, err
	}
	e := &SellerEarnings{Orders: orders}
	for _, o := range orders {
		if o.Status == "SELESAI" {
			e.TotalRevenue += o.Subtotal
			e.TotalCommission += o.Subtotal * SellerCommissionRate
			e.CompletedOrders++
		}
	}
	e.NetEarnings = e.TotalRevenue - e.TotalCommission
	return e, nil
}
