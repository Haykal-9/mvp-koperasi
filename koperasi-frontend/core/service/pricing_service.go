package service

import (
	"fmt"

	"koperasi-frontend/core/model"
)

// Aturan harga e-commerce (voucher, poin, ongkir). Sebelumnya tersebar di mock.
const (
	// PointEarnPerRupiah: 1 poin diperoleh per kelipatan rupiah ini saat belanja.
	PointEarnPerRupiah = 1000.0
	// PointValueRupiah: nilai baku 1 poin untuk redeem dan konversi simpanan.
	PointValueRupiah = 1.0
	// MinPointConversion: minimum poin yang dapat dikonversi ke simpanan.
	MinPointConversion = 100.0
)

// PricingService memusatkan perhitungan diskon voucher, perolehan poin, dan
// ongkos kirim. Murni (tanpa akses data) agar mudah diuji.
type PricingService struct{}

func NewPricingService() *PricingService { return &PricingService{} }

// VoucherDiscount menghitung diskon untuk voucher v atas subtotal tertentu.
// Mengembalikan (diskon, pesan). Diskon 0 bila voucher tidak valid.
func (s *PricingService) VoucherDiscount(v *model.Voucher, subtotal float64) (float64, string) {
	if v == nil {
		return 0, "Voucher tidak ditemukan"
	}
	if v.Status != "ACTIVE" {
		return 0, "Voucher sudah tidak berlaku"
	}
	if v.Kuota <= 0 {
		return 0, "Kuota voucher sudah habis"
	}
	if subtotal < v.MinPembelian {
		return 0, fmt.Sprintf("Minimum pembelian Rp %.0f", v.MinPembelian)
	}
	discount := 0.0
	if v.TipeDiskon == "PERCENT" {
		discount = subtotal * v.NilaiDiskon / 100
		if discount > v.MaksDiskon {
			discount = v.MaksDiskon
		}
	} else {
		discount = v.NilaiDiskon
	}
	return discount, "Voucher valid"
}

// PointsEarned menghitung poin yang diperoleh dari subtotal belanja.
func (s *PricingService) PointsEarned(subtotal float64) float64 {
	return float64(int(subtotal / PointEarnPerRupiah))
}

// ShippingCost menghitung ongkir = harga/kg * berat (dibulatkan naik, minimal 1 kg).
func (s *PricingService) ShippingCost(opt *model.ShippingOption, weightGrams int) float64 {
	if opt == nil {
		return 0
	}
	kg := float64(weightGrams) / 1000.0
	if kg < 1 {
		kg = 1
	}
	return opt.Harga * kg
}
