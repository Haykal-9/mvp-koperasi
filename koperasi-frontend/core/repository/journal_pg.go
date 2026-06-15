package repository

import (
	"context"

	"gorm.io/gorm"

	"koperasi-frontend/core/model"
)

// pgJournalRepository: implementasi JournalRepository di atas PostgreSQL (M2).
type pgJournalRepository struct {
	db *gorm.DB
}

func NewJournalRepository(db *gorm.DB) JournalRepository {
	return &pgJournalRepository{db: db}
}

const journalCols = `id, to_char(tanggal,'YYYY-MM-DD') AS tanggal, keterangan,
	akun_debit, akun_kredit, nominal, tipe_transaksi`

func (r *pgJournalRepository) List(ctx context.Context, tipe string) ([]model.JournalEntry, error) {
	out := []model.JournalEntry{}
	q := r.db.WithContext(ctx)
	sql := "SELECT " + journalCols + " FROM journal_entries"
	if tipe != "" && tipe != "ALL" {
		return out, q.Raw(sql+" WHERE tipe_transaksi = ? ORDER BY id DESC", tipe).Scan(&out).Error
	}
	return out, q.Raw(sql + " ORDER BY id DESC").Scan(&out).Error
}

func (r *pgJournalRepository) Counts(ctx context.Context) (JournalCounts, error) {
	type row struct {
		TipeTransaksi string
		N             int
	}
	rows := []row{}
	if err := r.db.WithContext(ctx).
		Raw("SELECT tipe_transaksi, count(*) AS n FROM journal_entries GROUP BY tipe_transaksi").
		Scan(&rows).Error; err != nil {
		return JournalCounts{}, err
	}
	var c JournalCounts
	for _, x := range rows {
		switch x.TipeTransaksi {
		case "SIMPANAN":
			c.Simpanan = x.N
		case "POS":
			c.POS = x.N
		case "PINJAMAN":
			c.Pinjaman = x.N
		case "MANUAL":
			c.Manual = x.N
		}
		c.All += x.N
	}
	return c, nil
}

func (r *pgJournalRepository) Create(ctx context.Context, j model.JournalEntry) error {
	return r.db.WithContext(ctx).Exec(
		`INSERT INTO journal_entries (tanggal, keterangan, akun_debit, akun_kredit, nominal, tipe_transaksi)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		j.Tanggal, j.Keterangan, j.AkunDebit, j.AkunKredit, j.Nominal, j.TipeTransaksi).Error
}

func (r *pgJournalRepository) Summary(ctx context.Context) (FinanceSummary, error) {
	var s FinanceSummary
	err := r.db.WithContext(ctx).Raw(
		`SELECT
		   COALESCE(SUM(CASE WHEN akun_debit='Kas' THEN nominal ELSE 0 END),0)
		     - COALESCE(SUM(CASE WHEN akun_kredit='Kas' THEN nominal ELSE 0 END),0) AS total_kas,
		   COALESCE(SUM(CASE WHEN akun_kredit LIKE 'Simpanan%' THEN nominal ELSE 0 END),0)
		     - COALESCE(SUM(CASE WHEN akun_debit LIKE 'Simpanan%' THEN nominal ELSE 0 END),0) AS total_simpanan,
		   COALESCE(SUM(CASE WHEN akun_debit='Piutang Anggota' THEN nominal ELSE 0 END),0)
		     - COALESCE(SUM(CASE WHEN akun_kredit='Piutang Anggota' THEN nominal ELSE 0 END),0) AS total_piutang,
		   COALESCE(SUM(CASE WHEN akun_kredit='Pendapatan Penjualan' THEN nominal ELSE 0 END),0) AS total_pendapatan
		 FROM journal_entries`).Scan(&s).Error
	return s, err
}

func (r *pgJournalRepository) MonthlyCashflow(ctx context.Context) ([]MonthlyCashflow, error) {
	out := []MonthlyCashflow{}
	err := r.db.WithContext(ctx).Raw(
		`SELECT to_char(tanggal,'YYYY-MM') AS bulan,
		        COALESCE(SUM(CASE WHEN akun_debit='Kas' THEN nominal ELSE 0 END),0) AS pemasukan,
		        COALESCE(SUM(CASE WHEN akun_kredit='Kas' THEN nominal ELSE 0 END),0) AS pengeluaran
		   FROM journal_entries
		  GROUP BY to_char(tanggal,'YYYY-MM')`).Scan(&out).Error
	return out, err
}
