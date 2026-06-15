package service

import (
	"context"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
)

// ECStoreService: aturan bisnis storefront E-Commerce (M5, sisi pembaca).
// Menyusun data katalog/etalase; agregasi data didelegasikan ke ECommerceRepository.
type ECStoreService struct {
	repo repository.ECommerceRepository
}

func NewECStoreService(repo repository.ECommerceRepository) *ECStoreService {
	return &ECStoreService{repo: repo}
}

// Featured mengembalikan produk APPROVED untuk etalase, dibatasi limit (0 = semua).
func (s *ECStoreService) Featured(ctx context.Context, limit int) ([]model.ECProduct, error) {
	products, err := s.repo.ProductsApproved(ctx)
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(products) > limit {
		products = products[:limit]
	}
	return products, nil
}

func (s *ECStoreService) Categories(ctx context.Context) ([]string, error) {
	return s.repo.Categories(ctx)
}

func (s *ECStoreService) SellerProfiles(ctx context.Context) ([]model.SellerProfile, error) {
	return s.repo.SellerProfiles(ctx)
}

// Catalog memilih sumber produk: pencarian > kategori > seluruh APPROVED.
func (s *ECStoreService) Catalog(ctx context.Context, query, category string) ([]model.ECProduct, error) {
	switch {
	case query != "":
		return s.repo.SearchProducts(ctx, query)
	case category != "":
		return s.repo.ProductsByCategory(ctx, category)
	default:
		return s.repo.ProductsApproved(ctx)
	}
}

func (s *ECStoreService) ProductByID(ctx context.Context, id int) (*model.ECProduct, error) {
	return s.repo.ProductByID(ctx, id)
}

func (s *ECStoreService) SellerProfile(ctx context.Context, sellerID int) (*model.SellerProfile, error) {
	return s.repo.SellerProfile(ctx, sellerID)
}

func (s *ECStoreService) Reviews(ctx context.Context, productID int) ([]model.ProductReview, error) {
	return s.repo.ReviewsByProduct(ctx, productID)
}

// Related mengembalikan produk APPROVED lain dari seller yang sama (kecuali excludeID),
// dibatasi limit. Untuk rekomendasi di halaman detail produk.
func (s *ECStoreService) Related(ctx context.Context, sellerID, excludeID, limit int) ([]model.ECProduct, error) {
	all, err := s.repo.ProductsBySeller(ctx, sellerID)
	if err != nil {
		return nil, err
	}
	out := make([]model.ECProduct, 0, len(all))
	for _, p := range all {
		if p.ID != excludeID && p.Status == "APPROVED" {
			out = append(out, p)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// SellerApprovedProducts mengembalikan produk APPROVED milik seller (etalase toko).
func (s *ECStoreService) SellerApprovedProducts(ctx context.Context, sellerID int) ([]model.ECProduct, error) {
	all, err := s.repo.ProductsBySeller(ctx, sellerID)
	if err != nil {
		return nil, err
	}
	out := make([]model.ECProduct, 0, len(all))
	for _, p := range all {
		if p.Status == "APPROVED" {
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *ECStoreService) IsWishlisted(ctx context.Context, userID, productID int) (bool, error) {
	return s.repo.IsWishlisted(ctx, userID, productID)
}
