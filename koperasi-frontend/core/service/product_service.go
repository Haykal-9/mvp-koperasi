package service

import (
	"context"
	"fmt"
	"strings"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
)

// ProductService: aturan bisnis produk & stok (M3 POS).
type ProductService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// Catalog mengembalikan produk APPROVED terfilter + daftar kategori.
func (s *ProductService) Catalog(ctx context.Context, q, kategori string) ([]model.Product, []string, error) {
	products, err := s.repo.Catalog(ctx, strings.ToLower(strings.TrimSpace(q)), kategori)
	if err != nil {
		return nil, nil, err
	}
	cats, err := s.repo.Categories(ctx)
	if err != nil {
		return nil, nil, err
	}
	return products, cats, nil
}

func (s *ProductService) FindByID(ctx context.Context, id int) (*model.Product, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ProductService) PendingReview(ctx context.Context) ([]model.Product, error) {
	return s.repo.ListByStatus(ctx, "PENDING")
}

// StatusForRole menentukan status awal produk: OWNER/KASIR langsung APPROVED,
// peran lain PENDING (menunggu review pengurus).
func (s *ProductService) StatusForRole(role string) string {
	if role == "OWNER" || role == "KASIR" {
		return "APPROVED"
	}
	return "PENDING"
}

// ValidateCreate memvalidasi input pembuatan produk.
func (s *ProductService) ValidateCreate(nama, kategori string, harga float64, stok int) error {
	if nama == "" || kategori == "" || harga <= 0 || stok < 0 {
		return fmt.Errorf("Nama, kategori, harga (>0), dan stok (≥0) wajib diisi.")
	}
	return nil
}

func (s *ProductService) Create(ctx context.Context, p model.Product) (int, error) {
	if p.FotoURL == "" {
		p.FotoURL = "https://placehold.co/400x300?text=" + strings.ReplaceAll(p.Nama, " ", "+")
	}
	if p.BatasStokMinimum < 0 {
		p.BatasStokMinimum = 0
	}
	return s.repo.Create(ctx, p)
}

func (s *ProductService) Approve(ctx context.Context, id int) error {
	return s.repo.SetStatus(ctx, id, "APPROVED")
}

func (s *ProductService) Reject(ctx context.Context, id int) error {
	return s.repo.SetStatus(ctx, id, "REJECTED")
}

func (s *ProductService) StockHistory(ctx context.Context, productID int) ([]model.StockChange, error) {
	return s.repo.StockHistory(ctx, productID)
}

// ValidateStock memvalidasi penyesuaian stok manual.
func (s *ProductService) ValidateStock(tipe string, jumlah int) error {
	if tipe != "RESTOCK" && tipe != "KOREKSI" {
		return fmt.Errorf("Tipe perubahan stok tidak valid.")
	}
	if jumlah == 0 {
		return fmt.Errorf("Jumlah tidak boleh 0.")
	}
	if tipe == "RESTOCK" && jumlah < 0 {
		return fmt.Errorf("Restock harus berupa angka positif.")
	}
	return nil
}

func (s *ProductService) AdjustStock(ctx context.Context, productID int, tipe string, jumlah int, keterangan string) error {
	return s.repo.AdjustStock(ctx, productID, tipe, jumlah, keterangan)
}
