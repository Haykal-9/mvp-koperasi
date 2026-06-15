package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"koperasi-frontend/core/model"
)

// pgOrderRepository: implementasi OrderRepository di atas PostgreSQL.
type pgOrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &pgOrderRepository{db: db}
}

const orderCols = `id, nomor_order, pembeli_nama, total_harga, fee_koperasi, status,
	metode_bayar, COALESCE(komplain_alasan,'') AS komplain_alasan,
	COALESCE(to_char(komplain_tanggal,'YYYY-MM-DD'),'') AS komplain_tanggal,
	COALESCE(komplain_bukti,'') AS komplain_bukti,
	to_char(created_at,'YYYY-MM-DD HH24:MI') AS created_at`

// orderRow: DTO scan tanpa field relasi (model.Order.Items membuat GORM gagal
// memetakan saat Scan). Dikonversi ke model.Order, item dimuat terpisah.
type orderRow struct {
	ID              int
	NomorOrder      string
	PembeliNama     string
	TotalHarga      float64
	FeeKoperasi     float64
	Status          string
	MetodeBayar     string
	KomplainAlasan  string
	KomplainTanggal string
	KomplainBukti   string
	CreatedAt       string
}

func (x orderRow) toModel() model.Order {
	return model.Order{
		ID: x.ID, NomorOrder: x.NomorOrder, PembeliNama: x.PembeliNama,
		TotalHarga: x.TotalHarga, FeeKoperasi: x.FeeKoperasi, Status: x.Status,
		MetodeBayar: x.MetodeBayar, KomplainAlasan: x.KomplainAlasan,
		KomplainTanggal: x.KomplainTanggal, KomplainBukti: x.KomplainBukti,
		CreatedAt: x.CreatedAt,
	}
}

func (r *pgOrderRepository) scanOrders(ctx context.Context, sql string, args ...interface{}) ([]model.Order, error) {
	rows := []orderRow{}
	if err := r.db.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model.Order, len(rows))
	for i, x := range rows {
		out[i] = x.toModel()
	}
	return out, nil
}

// attachItems memuat order_items untuk sekumpulan order dalam satu query.
func (r *pgOrderRepository) attachItems(ctx context.Context, orders []model.Order) error {
	if len(orders) == 0 {
		return nil
	}
	ids := make([]int, len(orders))
	idx := make(map[int]int, len(orders))
	for i := range orders {
		ids[i] = orders[i].ID
		idx[orders[i].ID] = i
	}
	type itemRow struct {
		OrderID     int
		ProductID   int
		ProductNama string
		Jumlah      int
		HargaSatuan float64
		Subtotal    float64
	}
	rows := []itemRow{}
	if err := r.db.WithContext(ctx).Raw(
		`SELECT order_id, product_id, product_nama, jumlah, harga_satuan, subtotal
		   FROM order_items WHERE order_id IN ? ORDER BY id`, ids).Scan(&rows).Error; err != nil {
		return err
	}
	for _, it := range rows {
		i := idx[it.OrderID]
		orders[i].Items = append(orders[i].Items, model.OrderItem{
			ProductID: it.ProductID, ProductNama: it.ProductNama, Jumlah: it.Jumlah,
			HargaSatuan: it.HargaSatuan, Subtotal: it.Subtotal,
		})
	}
	return nil
}

func (r *pgOrderRepository) List(ctx context.Context, scopeNama string) ([]model.Order, error) {
	sql := "SELECT " + orderCols + " FROM orders"
	var out []model.Order
	var err error
	if scopeNama != "" {
		out, err = r.scanOrders(ctx, sql+" WHERE pembeli_nama = ? ORDER BY id DESC", scopeNama)
	} else {
		out, err = r.scanOrders(ctx, sql+" ORDER BY id DESC")
	}
	if err != nil {
		return nil, err
	}
	return out, r.attachItems(ctx, out)
}

func (r *pgOrderRepository) ListByStatus(ctx context.Context, status string) ([]model.Order, error) {
	out, err := r.scanOrders(ctx, "SELECT "+orderCols+" FROM orders WHERE status = ? ORDER BY id DESC", status)
	if err != nil {
		return nil, err
	}
	return out, r.attachItems(ctx, out)
}

func (r *pgOrderRepository) FindByID(ctx context.Context, id int) (*model.Order, error) {
	out, err := r.scanOrders(ctx, "SELECT "+orderCols+" FROM orders WHERE id = ? LIMIT 1", id)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 || out[0].ID == 0 {
		return nil, nil
	}
	if err := r.attachItems(ctx, out); err != nil {
		return nil, err
	}
	return &out[0], nil
}

func (r *pgOrderRepository) Create(ctx context.Context, o model.Order) (*model.Order, error) {
	today := o.CreatedAt
	if len(today) >= 10 {
		today = today[:10]
	}
	prefix := "ORD-" + strings.ReplaceAll(today, "-", "") + "-"

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cnt int
		if err := tx.Raw("SELECT count(*) FROM orders WHERE nomor_order LIKE ?", prefix+"%").Scan(&cnt).Error; err != nil {
			return err
		}
		nomor := fmt.Sprintf("%s%03d", prefix, cnt+1)

		var id int
		if err := tx.Raw(
			`INSERT INTO orders (nomor_order, pembeli_nama, total_harga, fee_koperasi, status, metode_bayar, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, now()) RETURNING id`,
			nomor, o.PembeliNama, o.TotalHarga, o.FeeKoperasi, o.Status, o.MetodeBayar,
		).Scan(&id).Error; err != nil {
			return err
		}
		o.ID = id
		o.NomorOrder = nomor

		for _, it := range o.Items {
			if err := tx.Exec(
				`INSERT INTO order_items (order_id, product_id, product_nama, jumlah, harga_satuan, subtotal)
				 VALUES (?, ?, ?, ?, ?, ?)`,
				id, it.ProductID, it.ProductNama, it.Jumlah, it.HargaSatuan, it.Subtotal).Error; err != nil {
				return err
			}
			// Kurangi stok + catat pergerakan stok.
			var stok int
			if err := tx.Raw("SELECT stok FROM products WHERE id = ? FOR UPDATE", it.ProductID).Scan(&stok).Error; err != nil {
				return err
			}
			newStok := stok - it.Jumlah
			if newStok < 0 {
				return fmt.Errorf("stok produk %s tidak cukup", it.ProductNama)
			}
			if err := tx.Exec("UPDATE products SET stok = ? WHERE id = ?", newStok, it.ProductID).Error; err != nil {
				return err
			}
			if err := tx.Exec(
				`INSERT INTO stock_changes (product_id, tipe, jumlah, stok_setelah, keterangan, created_at)
				 VALUES (?, 'KELUAR_PENJUALAN', ?, ?, ?, now())`,
				it.ProductID, -it.Jumlah, newStok, "Penjualan "+nomor).Error; err != nil {
				return err
			}
		}

		// Jurnal otomatis penjualan POS.
		return tx.Exec(
			`INSERT INTO journal_entries (tanggal, keterangan, akun_debit, akun_kredit, nominal, tipe_transaksi)
			 VALUES (CURRENT_DATE, ?, 'Kas', 'Pendapatan Penjualan', ?, 'POS')`,
			"Penjualan POS "+nomor, o.TotalHarga).Error
	})
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *pgOrderRepository) SetStatus(ctx context.Context, id int, status string) error {
	return r.db.WithContext(ctx).
		Exec("UPDATE orders SET status = ? WHERE id = ?", status, id).Error
}

func (r *pgOrderRepository) SetComplaint(ctx context.Context, id int, alasan, bukti string) error {
	return r.db.WithContext(ctx).Exec(
		`UPDATE orders SET status = 'DISPUTED', komplain_alasan = ?, komplain_bukti = ?,
		        komplain_tanggal = CURRENT_DATE WHERE id = ?`,
		alasan, bukti, id).Error
}

func (r *pgOrderRepository) Resolve(ctx context.Context, id int, approveRetur bool) error {
	if !approveRetur {
		return r.db.WithContext(ctx).Exec("UPDATE orders SET status = 'SELESAI' WHERE id = ?", id).Error
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var nomor string
		if err := tx.Raw("SELECT nomor_order FROM orders WHERE id = ?", id).Scan(&nomor).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE orders SET status = 'BATAL' WHERE id = ?", id).Error; err != nil {
			return err
		}
		// Kembalikan stok untuk setiap item (KOREKSI positif).
		type itemRow struct {
			ProductID int
			Jumlah    int
		}
		items := []itemRow{}
		if err := tx.Raw("SELECT product_id, jumlah FROM order_items WHERE order_id = ?", id).Scan(&items).Error; err != nil {
			return err
		}
		for _, it := range items {
			var stok int
			if err := tx.Raw("SELECT stok FROM products WHERE id = ? FOR UPDATE", it.ProductID).Scan(&stok).Error; err != nil {
				return err
			}
			newStok := stok + it.Jumlah
			if err := tx.Exec("UPDATE products SET stok = ? WHERE id = ?", newStok, it.ProductID).Error; err != nil {
				return err
			}
			if err := tx.Exec(
				`INSERT INTO stock_changes (product_id, tipe, jumlah, stok_setelah, keterangan, created_at)
				 VALUES (?, 'KOREKSI', ?, ?, ?, now())`,
				it.ProductID, it.Jumlah, newStok, "Retur dari komplain "+nomor).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
