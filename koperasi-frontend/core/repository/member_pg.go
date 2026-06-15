package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"koperasi-frontend/core/model"
)

// pgMemberRepository: implementasi MemberRepository di atas PostgreSQL (GORM).
// Memakai SQL eksplisit + to_char untuk DATE agar memetakan langsung ke
// model.* yang memakai tanggal bertipe string (kompatibel dengan templat SSR).
type pgMemberRepository struct {
	db *gorm.DB
}

// NewMemberRepository membuat MemberRepository berbasis PostgreSQL.
func NewMemberRepository(db *gorm.DB) MemberRepository {
	return &pgMemberRepository{db: db}
}

const memberCols = `id, nomor_anggota, nama, nik, alamat, no_hp, status,
	simpanan_pokok, simpanan_wajib, simpanan_sukarela,
	to_char(tanggal_masuk,'YYYY-MM-DD') AS tanggal_masuk`

func (r *pgMemberRepository) List(ctx context.Context, status string) ([]model.Member, error) {
	out := []model.Member{}
	q := r.db.WithContext(ctx)
	sql := "SELECT " + memberCols + " FROM members"
	if status != "" && status != "ALL" {
		sql += " WHERE status = ?"
		q = q.Raw(sql+" ORDER BY id", status)
	} else {
		q = q.Raw(sql + " ORDER BY id")
	}
	return out, q.Scan(&out).Error
}

func (r *pgMemberRepository) Counts(ctx context.Context) (MemberCounts, error) {
	type row struct {
		Status string
		N      int
	}
	rows := []row{}
	if err := r.db.WithContext(ctx).
		Raw("SELECT status, count(*) AS n FROM members GROUP BY status").
		Scan(&rows).Error; err != nil {
		return MemberCounts{}, err
	}
	var c MemberCounts
	for _, x := range rows {
		switch x.Status {
		case "AKTIF":
			c.Aktif = x.N
		case "PENDING":
			c.Pending = x.N
		case "NON_AKTIF":
			c.NonAktif = x.N
		}
		c.All += x.N
	}
	return c, nil
}

func (r *pgMemberRepository) FindByID(ctx context.Context, id int) (*model.Member, error) {
	var m model.Member
	err := r.db.WithContext(ctx).
		Raw("SELECT "+memberCols+" FROM members WHERE id = ? LIMIT 1", id).
		Scan(&m).Error
	if err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil // tidak ditemukan
	}
	return &m, nil
}

func (r *pgMemberRepository) FindByNama(ctx context.Context, nama string) (*model.Member, error) {
	var m model.Member
	err := r.db.WithContext(ctx).
		Raw("SELECT "+memberCols+" FROM members WHERE nama = ? LIMIT 1", nama).Scan(&m).Error
	if err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil // tidak ditemukan
	}
	return &m, nil
}

func (r *pgMemberRepository) Create(ctx context.Context, nama, nik, alamat, noHP string) (int, error) {
	var id int
	err := r.db.WithContext(ctx).Raw(
		`INSERT INTO members (nomor_anggota, nama, nik, alamat, no_hp, status, tanggal_masuk)
		 VALUES ('-', ?, ?, ?, ?, 'PENDING', CURRENT_DATE) RETURNING id`,
		nama, nik, alamat, noHP,
	).Scan(&id).Error
	return id, err
}

func (r *pgMemberRepository) SetStatus(ctx context.Context, id int, status string) error {
	return r.db.WithContext(ctx).
		Exec("UPDATE members SET status = ? WHERE id = ?", status, id).Error
}

func (r *pgMemberRepository) Approve(ctx context.Context, id int) (string, error) {
	var nomor string
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Bangkitkan nomor anggota berikutnya: KOP-#### berdasarkan max yang ada.
		var next int
		if err := tx.Raw(
			`SELECT COALESCE(MAX(NULLIF(substring(nomor_anggota from 5), '')::int), 0) + 1
			   FROM members WHERE nomor_anggota LIKE 'KOP-%'`).Scan(&next).Error; err != nil {
			return err
		}
		nomor = fmt.Sprintf("KOP-%04d", next)

		res := tx.Exec(
			`UPDATE members SET status='AKTIF', nomor_anggota=?, tanggal_masuk=CURRENT_DATE
			  WHERE id=? AND status='PENDING'`, nomor, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("anggota tidak dalam status PENDING")
		}
		return nil
	})
	return nomor, err
}

func (r *pgMemberRepository) SimpananHistory(ctx context.Context, memberID int) ([]model.SimpananTransaction, error) {
	out := []model.SimpananTransaction{}
	err := r.db.WithContext(ctx).Raw(
		`SELECT s.id, m.nama AS member_nama, s.jenis, s.tipe, s.nominal, s.keterangan,
		        to_char(s.created_at,'YYYY-MM-DD') AS created_at
		   FROM simpanan_transactions s
		   JOIN members m ON m.id = s.member_id
		  WHERE s.member_id = ?
		  ORDER BY s.id DESC`, memberID).Scan(&out).Error
	return out, err
}

func (r *pgMemberRepository) LoansByMember(ctx context.Context, memberID int) ([]model.Loan, error) {
	type loanRow struct {
		ID          int
		MemberNama  string
		Nominal     float64
		TenorBulan  int
		BungaPersen float64
		Tujuan      string
		SisaPokok   float64
		Status      string
		TanggalCair string
	}
	rows := []loanRow{}
	if err := r.db.WithContext(ctx).Raw(
		`SELECT l.id, m.nama AS member_nama, l.nominal, l.tenor_bulan, l.bunga_persen,
		        l.tujuan, l.sisa_pokok, l.status,
		        COALESCE(to_char(l.tanggal_cair,'YYYY-MM-DD'),'') AS tanggal_cair
		   FROM loans l
		   JOIN members m ON m.id = l.member_id
		  WHERE l.member_id = ?
		  ORDER BY l.id`, memberID).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model.Loan, 0, len(rows))
	for _, x := range rows {
		out = append(out, model.Loan{
			ID: x.ID, MemberNama: x.MemberNama, Nominal: x.Nominal, TenorBulan: x.TenorBulan,
			BungaPersen: x.BungaPersen, Tujuan: x.Tujuan, SisaPokok: x.SisaPokok,
			Status: x.Status, TanggalCair: x.TanggalCair,
		})
	}
	return out, nil
}

func (r *pgMemberRepository) RecordSimpanan(ctx context.Context, memberID int, jenis, tipe string, nominal float64, keterangan string) error {
	// Kolom saldo per jenis (whitelist agar aman dari injeksi).
	var saldoCol string
	switch jenis {
	case "POKOK":
		saldoCol = "simpanan_pokok"
	case "WAJIB":
		saldoCol = "simpanan_wajib"
	case "SUKARELA":
		saldoCol = "simpanan_sukarela"
	default:
		return fmt.Errorf("jenis simpanan tidak valid")
	}

	delta := nominal
	if tipe == "KELUAR" {
		delta = -nominal
	}
	debit, kredit := "Kas", "Simpanan "+jenis
	if tipe == "KELUAR" {
		debit, kredit = "Simpanan "+jenis, "Kas"
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var nama string
		if err := tx.Raw("SELECT nama FROM members WHERE id = ?", memberID).Scan(&nama).Error; err != nil {
			return err
		}
		if nama == "" {
			return fmt.Errorf("anggota tidak ditemukan")
		}
		// 1) Perbarui saldo agregat.
		if err := tx.Exec(
			fmt.Sprintf("UPDATE members SET %s = %s + ? WHERE id = ?", saldoCol, saldoCol),
			delta, memberID).Error; err != nil {
			return err
		}
		// 2) Catat transaksi simpanan.
		if err := tx.Exec(
			`INSERT INTO simpanan_transactions (member_id, jenis, tipe, nominal, keterangan, created_at)
			 VALUES (?, ?, ?, ?, ?, now())`,
			memberID, jenis, tipe, nominal, keterangan).Error; err != nil {
			return err
		}
		// 3) Jurnal otomatis (FR-M2-01).
		return tx.Exec(
			`INSERT INTO journal_entries (tanggal, keterangan, akun_debit, akun_kredit, nominal, tipe_transaksi)
			 VALUES (CURRENT_DATE, ?, ?, ?, ?, 'SIMPANAN')`,
			fmt.Sprintf("Simpanan %s %s — %s", jenis, tipe, nama), debit, kredit, nominal).Error
	})
}
