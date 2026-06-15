package service

import (
	"context"
	"fmt"
	"time"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
)

// FinanceService: aturan bisnis buku besar / keuangan (M2). Memusatkan validasi
// jurnal manual dan penyusunan laporan ringkasan; agregasi data didelegasikan ke
// JournalRepository.
type FinanceService struct {
	repo repository.JournalRepository
}

func NewFinanceService(repo repository.JournalRepository) *FinanceService {
	return &FinanceService{repo: repo}
}

// CashflowMonth: satu baris arus kas bulanan siap tampil (label sudah lokal).
type CashflowMonth struct {
	Bulan       string
	Pemasukan   float64
	Pengeluaran float64
	Selisih     float64
}

func (s *FinanceService) Journals(ctx context.Context, tipe string) ([]model.JournalEntry, error) {
	return s.repo.List(ctx, tipe)
}

func (s *FinanceService) Counts(ctx context.Context) (repository.JournalCounts, error) {
	return s.repo.Counts(ctx)
}

func (s *FinanceService) Summary(ctx context.Context) (repository.FinanceSummary, error) {
	return s.repo.Summary(ctx)
}

// CreateManual memvalidasi lalu menyimpan jurnal manual (tipe MANUAL).
func (s *FinanceService) CreateManual(ctx context.Context, tanggal, keterangan, akunDebit, akunKredit string, nominal float64) error {
	if tanggal == "" || keterangan == "" || akunDebit == "" || akunKredit == "" || nominal <= 0 {
		return fmt.Errorf("Semua field wajib diisi dan nominal harus > 0.")
	}
	return s.repo.Create(ctx, model.JournalEntry{
		Tanggal:       tanggal,
		Keterangan:    keterangan,
		AkunDebit:     akunDebit,
		AkunKredit:    akunKredit,
		Nominal:       nominal,
		TipeTransaksi: "MANUAL",
	})
}

// MonthlyRows menyusun arus kas n bulan terakhir (termasuk bulan kosong) dengan
// label "Jan 2006", berbasis data agregat per bulan dari repository.
func (s *FinanceService) MonthlyRows(ctx context.Context, n int) ([]CashflowMonth, error) {
	rows, err := s.repo.MonthlyCashflow(ctx)
	if err != nil {
		return nil, err
	}
	byMonth := make(map[string]repository.MonthlyCashflow, len(rows))
	for _, r := range rows {
		byMonth[r.Bulan] = r
	}
	out := make([]CashflowMonth, 0, n)
	now := time.Now()
	for i := n - 1; i >= 0; i-- {
		t := now.AddDate(0, -i, 0)
		key := t.Format("2006-01")
		m := byMonth[key] // zero value bila tak ada data bulan itu
		out = append(out, CashflowMonth{
			Bulan:       t.Format("Jan 2006"),
			Pemasukan:   m.Pemasukan,
			Pengeluaran: m.Pengeluaran,
			Selisih:     m.Pemasukan - m.Pengeluaran,
		})
	}
	return out, nil
}
