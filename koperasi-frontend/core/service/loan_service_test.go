package service

import (
	"testing"

	"koperasi-frontend/core/model"
)

func TestValidateApplication(t *testing.T) {
	s := NewLoanService(nil)
	cases := []struct {
		name    string
		nominal float64
		tenor   int
		tujuan  string
		wantErr bool
	}{
		{"valid", 1_000_000, 12, "Modal usaha", false},
		{"nominal kurang", 400_000, 12, "x", true},
		{"tenor 0", 1_000_000, 0, "x", true},
		{"tenor kurang", 1_000_000, 2, "x", true},
		{"tenor lebih", 1_000_000, 25, "x", true},
		{"tujuan kosong", 1_000_000, 12, "", true},
		{"batas minimal nominal", 500_000, 3, "x", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := s.ValidateApplication(c.nominal, c.tenor, c.tujuan)
			if (err != nil) != c.wantErr {
				t.Fatalf("ValidateApplication() err=%v, wantErr=%v", err, c.wantErr)
			}
		})
	}
}

func TestBuildSchedule(t *testing.T) {
	s := NewLoanService(nil)
	insts := s.BuildSchedule(1_200_000, 3, 1.5, "2025-01-10")
	if len(insts) != 3 {
		t.Fatalf("jumlah angsuran = %d, mau 3", len(insts))
	}
	// pokok rata: 1.200.000 / 3 = 400.000
	for i, in := range insts {
		if in.NominalPokok != 400_000 {
			t.Errorf("angsuran %d pokok = %.0f, mau 400000", i+1, in.NominalPokok)
		}
		if in.Status != "BELUM" {
			t.Errorf("angsuran %d status = %s, mau BELUM", i+1, in.Status)
		}
	}
	// bunga menurun: bulan1 atas 1.2jt, bulan2 atas 800rb, bulan3 atas 400rb (1.5%)
	wantBunga := []float64{18_000, 12_000, 6_000}
	for i, w := range wantBunga {
		if insts[i].NominalBunga != w {
			t.Errorf("angsuran %d bunga = %.0f, mau %.0f", i+1, insts[i].NominalBunga, w)
		}
	}
	// jatuh tempo bulan pertama = sebulan setelah tanggal cair
	if insts[0].JatuhTempo != "2025-02-10" {
		t.Errorf("jatuh tempo bulan1 = %s, mau 2025-02-10", insts[0].JatuhTempo)
	}
}

func TestBuildScheduleTenorNol(t *testing.T) {
	if got := NewLoanService(nil).BuildSchedule(1_000_000, 0, 1.5, "2025-01-01"); got != nil {
		t.Fatalf("tenor 0 harus nil, dapat %v", got)
	}
}

func TestApplyPaymentDanLunas(t *testing.T) {
	s := NewLoanService(nil)
	l := &model.Loan{
		Nominal:      900_000,
		TenorBulan:   3,
		BungaPersen:  1.5,
		SisaPokok:    900_000,
		Status:       "AKTIF",
		Installments: s.BuildSchedule(900_000, 3, 1.5, "2025-01-01"),
	}
	// bayar 1
	inst, lunas, err := s.ApplyPayment(l, "2025-02-01")
	if err != nil || lunas {
		t.Fatalf("bayar1: err=%v lunas=%v (mau err=nil lunas=false)", err, lunas)
	}
	if inst.BulanKe != 1 || inst.Status != "DIBAYAR" {
		t.Fatalf("bayar1: inst=%+v", inst)
	}
	if l.SisaPokok != 600_000 {
		t.Fatalf("sisa pokok setelah bayar1 = %.0f, mau 600000", l.SisaPokok)
	}
	// bayar 2
	if _, lunas, _ := s.ApplyPayment(l, "2025-03-01"); lunas {
		t.Fatal("bayar2 belum boleh lunas")
	}
	// bayar 3 -> lunas
	_, lunas, err = s.ApplyPayment(l, "2025-04-01")
	if err != nil || !lunas {
		t.Fatalf("bayar3: err=%v lunas=%v (mau lunas=true)", err, lunas)
	}
	if l.Status != "LUNAS" {
		t.Fatalf("status = %s, mau LUNAS", l.Status)
	}
	if l.SisaPokok != 0 {
		t.Fatalf("sisa pokok = %.0f, mau 0", l.SisaPokok)
	}
	// bayar lagi -> error
	if _, _, err := s.ApplyPayment(l, "2025-05-01"); err == nil {
		t.Fatal("bayar setelah lunas harus error")
	}
}

func TestProgressDanEstimate(t *testing.T) {
	s := NewLoanService(nil)
	if got := s.Progress(1_000_000, 250_000); got != 75 {
		t.Errorf("Progress = %.0f, mau 75", got)
	}
	if got := s.Progress(0, 0); got != 0 {
		t.Errorf("Progress(0,0) = %.0f, mau 0", got)
	}
	// estimasi: pokok 1.2jt/12 = 100rb + bunga flat 1.2jt*1.5% = 18rb = 118rb
	if got := s.EstimateMonthly(1_200_000, 1.5, 12); got != 118_000 {
		t.Errorf("EstimateMonthly = %.0f, mau 118000", got)
	}
}
