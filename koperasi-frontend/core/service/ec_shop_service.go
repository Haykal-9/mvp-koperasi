package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
)

// ECShopService: aturan bisnis alur beli E-Commerce (checkout, order, wishlist).
// Perhitungan harga (ongkir, diskon voucher, poin) didelegasikan ke PricingService.
type ECShopService struct {
	repo    repository.ECommerceRepository
	pricing *PricingService
}

func NewECShopService(repo repository.ECommerceRepository) *ECShopService {
	return &ECShopService{repo: repo, pricing: NewPricingService()}
}

// CheckoutView menampung data untuk halaman checkout satu produk.
type CheckoutView struct {
	Product       *model.ECProduct
	Items         []model.ECOrderItem
	Subtotal      float64
	TotalWeight   int
	Addresses     []model.ECAddress
	ShippingOpts  []model.ShippingOption
	PointsBalance float64
	IsLinked      bool
	Qty           int
}

// Checkout menyiapkan data checkout untuk satu produk. Mengembalikan nil bila
// produk tidak ditemukan.
func (s *ECShopService) Checkout(ctx context.Context, userID, productID, qty int) (*CheckoutView, error) {
	if qty < 1 {
		qty = 1
	}
	product, err := s.repo.ProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, nil
	}
	subtotal := product.Harga * float64(qty)
	v := &CheckoutView{
		Product:     product,
		Subtotal:    subtotal,
		TotalWeight: product.Berat * qty,
		Qty:         qty,
		Items: []model.ECOrderItem{{
			ProductID: product.ID, ProductNama: product.Nama, SellerID: product.SellerID,
			Jumlah: qty, HargaSatuan: product.Harga, Subtotal: subtotal,
		}},
	}
	if v.Addresses, err = s.repo.AddressesByUser(ctx, userID); err != nil {
		return nil, err
	}
	if v.ShippingOpts, err = s.repo.ShippingOptions(ctx); err != nil {
		return nil, err
	}
	if up, err := s.repo.UserPoints(ctx, userID); err != nil {
		return nil, err
	} else if up != nil {
		v.PointsBalance = up.Balance
	}
	if u, err := s.repo.UserByID(ctx, userID); err != nil {
		return nil, err
	} else if u != nil {
		v.IsLinked = u.LinkedKoperasiMemberID > 0
	}
	return v, nil
}

// PlaceOrder memvalidasi & membuat order untuk satu produk (transaksional di repo).
func (s *ECShopService) PlaceOrder(ctx context.Context, userID, productID, qty int, alamat string, shippingOptID int, voucherCode, metodeBayar string, pointsUsed float64) (*model.ECOrder, error) {
	if qty < 1 {
		qty = 1
	}
	product, err := s.repo.ProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, fmt.Errorf("Produk tidak ditemukan.")
	}
	if product.Status != "APPROVED" {
		return nil, fmt.Errorf("Produk belum tersedia untuk dibeli.")
	}
	if product.Stok < qty {
		return nil, fmt.Errorf("Stok produk %s tidak cukup.", product.Nama)
	}
	if metodeBayar == "" {
		metodeBayar = "Transfer Bank"
	}

	subtotal := product.Harga * float64(qty)
	items := []model.ECOrderItem{{
		ProductID: product.ID, ProductNama: product.Nama, SellerID: product.SellerID,
		Jumlah: qty, HargaSatuan: product.Harga, Subtotal: subtotal,
	}}

	// Ongkir
	shipOpt, err := s.repo.ShippingOptionByID(ctx, shippingOptID)
	if err != nil {
		return nil, err
	}
	shippingCost := s.pricing.ShippingCost(shipOpt, product.Berat*qty)
	shippingName := ""
	if shipOpt != nil {
		shippingName = shipOpt.Nama
	}

	// Diskon voucher
	discount := 0.0
	if voucherCode != "" {
		voucherCode = strings.ToUpper(voucherCode)
		v, err := s.repo.VoucherByCode(ctx, voucherCode)
		if err != nil {
			return nil, err
		}
		var msg string
		discount, msg = s.pricing.VoucherDiscount(v, subtotal)
		if msg != "Voucher valid" {
			return nil, errors.New(msg)
		}
	}

	// Validasi poin
	if pointsUsed < 0 {
		pointsUsed = 0
	}
	if pointsUsed > 0 {
		up, err := s.repo.UserPoints(ctx, userID)
		if err != nil {
			return nil, err
		}
		if up == nil || up.Balance < pointsUsed {
			return nil, fmt.Errorf("Poin tidak cukup.")
		}
	}

	payable := subtotal + shippingCost - discount
	if pointsUsed > payable {
		pointsUsed = payable
	}
	total := payable - pointsUsed
	if total < 0 {
		total = 0
	}
	pointsEarned := s.pricing.PointsEarned(subtotal)

	order := model.ECOrder{
		BuyerID: userID, SellerID: product.SellerID,
		AlamatPengiriman: alamat, ShippingOption: shippingName, ShippingCost: shippingCost,
		Subtotal: subtotal, Discount: discount, TotalHarga: total,
		VoucherCode: voucherCode, MetodeBayar: metodeBayar, Status: "DIBAYAR",
		Items: items,
	}
	return s.repo.CreateOrder(ctx, order, pointsUsed, pointsEarned)
}

// Orders mengembalikan order buyer; statusFilter opsional.
func (s *ECShopService) Orders(ctx context.Context, userID int, statusFilter string) ([]model.ECOrder, error) {
	orders, err := s.repo.OrdersByBuyer(ctx, userID)
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

func (s *ECShopService) OrderByID(ctx context.Context, id int) (*model.ECOrder, error) {
	return s.repo.OrderByID(ctx, id)
}

func (s *ECShopService) ShipmentEvents(ctx context.Context, orderID int) ([]model.ShipmentEvent, error) {
	return s.repo.ShipmentEvents(ctx, orderID)
}

func (s *ECShopService) Wishlist(ctx context.Context, userID int) ([]model.ECProduct, error) {
	return s.repo.WishlistProducts(ctx, userID)
}

// ToggleWishlist menambah/menghapus produk dari wishlist; mengembalikan status
// (added) dan nama produk untuk pesan flash.
func (s *ECShopService) ToggleWishlist(ctx context.Context, userID, productID int) (added bool, name string, err error) {
	added, err = s.repo.ToggleWishlist(ctx, userID, productID)
	if err != nil {
		return false, "", err
	}
	name = "produk"
	if p, perr := s.repo.ProductByID(ctx, productID); perr == nil && p != nil {
		name = p.Nama
	}
	return added, name, nil
}

// ---- Review (4d-6) ----

func (s *ECShopService) HasReviewed(ctx context.Context, userID, productID int) (bool, error) {
	return s.repo.HasReviewed(ctx, userID, productID)
}

// SubmitReview menyisipkan ulasan produk (rating 1-5) bila belum pernah diulas.
func (s *ECShopService) SubmitReview(ctx context.Context, userID, productID, rating int, komentar, username string) error {
	return s.repo.SubmitReview(ctx, userID, productID, rating, komentar, username)
}
