package service

import (
	"testing"

	"koperasi-frontend/core/model"
)

func TestVoucherDiscount(t *testing.T) {
	s := NewPricingService()

	percent := &model.Voucher{Status: "ACTIVE", Kuota: 5, MinPembelian: 100_000, TipeDiskon: "PERCENT", NilaiDiskon: 10, MaksDiskon: 25_000}
	fixed := &model.Voucher{Status: "ACTIVE", Kuota: 5, MinPembelian: 50_000, TipeDiskon: "FIXED", NilaiDiskon: 15_000}

	cases := []struct {
		name     string
		v        *model.Voucher
		subtotal float64
		want     float64
		valid    bool
	}{
		{"nil voucher", nil, 100_000, 0, false},
		{"persen normal", percent, 200_000, 20_000, true},
		{"persen kena cap", percent, 500_000, 25_000, true}, // 10% = 50rb -> dibatasi 25rb
		{"di bawah minimum", percent, 50_000, 0, false},     // < MinPembelian 100rb
		{"fixed", fixed, 100_000, 15_000, true},
		{"tidak aktif", &model.Voucher{Status: "EXPIRED", Kuota: 5}, 100_000, 0, false},
		{"kuota habis", &model.Voucher{Status: "ACTIVE", Kuota: 0}, 100_000, 0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, msg := s.VoucherDiscount(c.v, c.subtotal)
			if got != c.want {
				t.Errorf("diskon = %.0f, mau %.0f (%s)", got, c.want, msg)
			}
			if (msg == "Voucher valid") != c.valid {
				t.Errorf("valid msg = %q, mau valid=%v", msg, c.valid)
			}
		})
	}
}

func TestPointsEarned(t *testing.T) {
	s := NewPricingService()
	cases := []struct {
		subtotal float64
		want     float64
	}{
		{999, 0},
		{1000, 1},
		{1500, 1},
		{150_000, 150},
	}
	for _, c := range cases {
		if got := s.PointsEarned(c.subtotal); got != c.want {
			t.Errorf("PointsEarned(%.0f) = %.0f, mau %.0f", c.subtotal, got, c.want)
		}
	}
}

func TestPointValueMatchesRedeemValue(t *testing.T) {
	if PointValueRupiah != 1 {
		t.Fatalf("PointValueRupiah = %.0f, mau 1 agar redeem dan konversi konsisten", PointValueRupiah)
	}
}

func TestShippingCost(t *testing.T) {
	s := NewPricingService()
	opt := &model.ShippingOption{Harga: 10_000}
	cases := []struct {
		name string
		opt  *model.ShippingOption
		gram int
		want float64
	}{
		{"nil option", nil, 2000, 0},
		{"di bawah 1kg dibulatkan", opt, 500, 10_000},
		{"2kg", opt, 2000, 20_000},
		{"2.5kg", opt, 2500, 25_000},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := s.ShippingCost(c.opt, c.gram); got != c.want {
				t.Errorf("ShippingCost = %.0f, mau %.0f", got, c.want)
			}
		})
	}
}
