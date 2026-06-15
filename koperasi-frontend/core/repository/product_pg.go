package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"koperasi-frontend/core/model"
)

// pgProductRepository: implementasi ProductRepository di atas PostgreSQL.
type pgProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &pgProductRepository{db: db}
}

const productCols = `id, nama, kategori, harga, stok, batas_stok_minimum,
	deskripsi, foto_url, penjual_nama, status`

func (r *pgProductRepository) Catalog(ctx context.Context, q, kategori string) ([]model.Product, error) {
	out := []model.Product{}
	sql := "SELECT " + productCols + " FROM products WHERE status = 'APPROVED'"
	args := []interface{}{}
	if q != "" {
		sql += " AND LOWER(nama) LIKE ?"
		args = append(args, "%"+q+"%")
	}
	if kategori != "" {
		sql += " AND kategori = ?"
		args = append(args, kategori)
	}
	sql += " ORDER BY id"
	return out, r.db.WithContext(ctx).Raw(sql, args...).Scan(&out).Error
}

func (r *pgProductRepository) Categories(ctx context.Context) ([]string, error) {
	out := []string{}
	err := r.db.WithContext(ctx).
		Raw("SELECT DISTINCT kategori FROM products WHERE kategori <> '' ORDER BY kategori").
		Scan(&out).Error
	return out, err
}

func (r *pgProductRepository) ListByStatus(ctx context.Context, status string) ([]model.Product, error) {
	out := []model.Product{}
	err := r.db.WithContext(ctx).
		Raw("SELECT "+productCols+" FROM products WHERE status = ? ORDER BY id", status).
		Scan(&out).Error
	return out, err
}

func (r *pgProductRepository) FindByID(ctx context.Context, id int) (*model.Product, error) {
	var p model.Product
	err := r.db.WithContext(ctx).
		Raw("SELECT "+productCols+" FROM products WHERE id = ? LIMIT 1", id).Scan(&p).Error
	if err != nil {
		return nil, err
	}
	if p.ID == 0 {
		return nil, nil
	}
	return &p, nil
}

func (r *pgProductRepository) Create(ctx context.Context, p model.Product) (int, error) {
	var id int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(
			`INSERT INTO products (nama, kategori, harga, stok, batas_stok_minimum,
				deskripsi, foto_url, penjual_nama, status)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
			p.Nama, p.Kategori, p.Harga, p.Stok, p.BatasStokMinimum,
			p.Deskripsi, p.FotoURL, p.PenjualNama, p.Status,
		).Scan(&id).Error; err != nil {
			return err
		}
		if p.Stok > 0 {
			return tx.Exec(
				`INSERT INTO stock_changes (product_id, tipe, jumlah, stok_setelah, keterangan, created_at)
				 VALUES (?, 'RESTOCK', ?, ?, ?, now())`,
				id, p.Stok, p.Stok, "Stok awal saat ditambahkan").Error
		}
		return nil
	})
	return id, err
}

func (r *pgProductRepository) SetStatus(ctx context.Context, id int, status string) error {
	return r.db.WithContext(ctx).
		Exec("UPDATE products SET status = ? WHERE id = ?", status, id).Error
}

func (r *pgProductRepository) StockHistory(ctx context.Context, productID int) ([]model.StockChange, error) {
	out := []model.StockChange{}
	err := r.db.WithContext(ctx).Raw(
		`SELECT sc.id, sc.product_id, p.nama AS product_nama, sc.tipe, sc.jumlah,
		        sc.stok_setelah, sc.keterangan,
		        to_char(sc.created_at,'YYYY-MM-DD') AS created_at
		   FROM stock_changes sc
		   JOIN products p ON p.id = sc.product_id
		  WHERE sc.product_id = ?
		  ORDER BY sc.id DESC`, productID).Scan(&out).Error
	return out, err
}

func (r *pgProductRepository) AdjustStock(ctx context.Context, productID int, tipe string, jumlah int, keterangan string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Kunci baris produk untuk mencegah balapan saat update stok.
		var stok int
		if err := tx.Raw("SELECT stok FROM products WHERE id = ? FOR UPDATE", productID).Scan(&stok).Error; err != nil {
			return err
		}
		newStok := stok + jumlah
		if newStok < 0 {
			return fmt.Errorf("stok tidak boleh negatif")
		}
		if err := tx.Exec("UPDATE products SET stok = ? WHERE id = ?", newStok, productID).Error; err != nil {
			return err
		}
		return tx.Exec(
			`INSERT INTO stock_changes (product_id, tipe, jumlah, stok_setelah, keterangan, created_at)
			 VALUES (?, ?, ?, ?, ?, now())`,
			productID, tipe, jumlah, newStok, keterangan).Error
	})
}
