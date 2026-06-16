package service

import (
	"context"
	"strings"
	"testing"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
)

type fakeSellerRepository struct {
	repository.ECommerceRepository

	profile        *model.SellerProfile
	order          *model.ECOrder
	created        bool
	createdProduct model.ECProduct
	shipped        bool
	gotSellerID    int
	gotOrderID     int
}

func (f *fakeSellerRepository) SellerProfile(context.Context, int) (*model.SellerProfile, error) {
	return f.profile, nil
}

func (f *fakeSellerRepository) CreateECProduct(_ context.Context, p model.ECProduct) (int, error) {
	f.created = true
	f.createdProduct = p
	return 10, nil
}

func (f *fakeSellerRepository) OrderByID(context.Context, int) (*model.ECOrder, error) {
	return f.order, nil
}

func (f *fakeSellerRepository) MarkOrderShipped(_ context.Context, sellerID, orderID int, _, _, _ string) error {
	f.shipped = true
	f.gotSellerID = sellerID
	f.gotOrderID = orderID
	return nil
}

func TestCreateProductRejectsInvalidStockAndWeight(t *testing.T) {
	cases := []struct {
		name  string
		stok  int
		berat int
		want  string
	}{
		{"stok negatif", -1, 100, "stok"},
		{"berat nol", 1, 0, "berat"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeSellerRepository{}
			svc := NewECSellerService(repo)

			err := svc.CreateProduct(context.Background(), 7, "Produk", "Desc", "Kategori", 10_000, tc.stok, tc.berat)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, mau memuat %q", err, tc.want)
			}
			if repo.created {
				t.Fatal("produk invalid tidak boleh dikirim ke repository")
			}
		})
	}
}

func TestCreateProductSendsValidatedProduct(t *testing.T) {
	repo := &fakeSellerRepository{profile: &model.SellerProfile{StoreName: "Toko Rapi"}}
	svc := NewECSellerService(repo)

	err := svc.CreateProduct(context.Background(), 7, " Produk ", "Desc", " Kategori ", 10_000, 3, 250)
	if err != nil {
		t.Fatal(err)
	}
	if !repo.created {
		t.Fatal("repository harus dipanggil untuk produk valid")
	}
	if repo.createdProduct.Nama != "Produk" || repo.createdProduct.Kategori != "Kategori" {
		t.Fatalf("produk belum dinormalisasi: %#v", repo.createdProduct)
	}
	if repo.createdProduct.SellerName != "Toko Rapi" {
		t.Fatalf("seller name = %q, mau Toko Rapi", repo.createdProduct.SellerName)
	}
}

func TestMarkShippedPassesSellerIDToRepository(t *testing.T) {
	repo := &fakeSellerRepository{
		profile: &model.SellerProfile{StoreName: "Toko Rapi"},
		order:   &model.ECOrder{ID: 99, SellerID: 7, Status: "DIBAYAR"},
	}
	svc := NewECSellerService(repo)

	_, err := svc.MarkShipped(context.Background(), 7, 99, "RESI-1")
	if err != nil {
		t.Fatal(err)
	}
	if !repo.shipped || repo.gotSellerID != 7 || repo.gotOrderID != 99 {
		t.Fatalf("shipment repo call = shipped:%v seller:%d order:%d", repo.shipped, repo.gotSellerID, repo.gotOrderID)
	}
}
