-- Fase 1 — Data seed (port dari core/mock/data.go & ecommerce.go).
-- Password demo disimpan sebagai hash bcrypt. Email+password sama untuk Koperasi & E-Commerce (Opsi A).

-- ============ DOMAIN KOPERASI ============

INSERT INTO users (id, email, password_hash, role, nama) VALUES
  (1, 'owner@koperasi.id',    '$2a$10$ziC5K65cq/hZEtuxfWdmZOCTRCGd6.xSjFOnj0HdpdjTM8oB.mxtG', 'OWNER',    'Budi Santoso'),
  (2, 'kasir@koperasi.id',    '$2a$10$afbed69AaUh6EcwisdNUUOPNeZynke4Zkv/bRi8M0YRKFWiWD.rT.', 'KASIR',    'Siti Aminah'),
  (3, 'anggota@koperasi.id',  '$2a$10$QKjvwsmNwv71BUgST6zunuXOZCimAtVM3irKMfVaS7tpkKjqzuVEW', 'ANGGOTA',  'Andi Wijaya'),
  (4, 'rina@koperasi.id',     '$2a$10$KNCILIyV3lsCEmuDTuUo/u.nnvhpvHreuTBWHyEXiu78cBpsasvwG', 'ANGGOTA',  'Rina Pertiwi'),
  (5, 'pengurus@koperasi.id', '$2a$10$J8n2AsGm6KXdIs1D6ZUfpuR6NUUVQQ1fsmTZ0lLlQOt17qgsVuvAK', 'PENGURUS', 'Dewi Lestari');

INSERT INTO members (id, nomor_anggota, nama, nik, alamat, no_hp, status, simpanan_pokok, simpanan_wajib, simpanan_sukarela, tanggal_masuk) VALUES
  (1, 'KOP-0001', 'Andi Wijaya',  '3201010101900001', 'Jl. Merdeka No. 12, Bandung',       '081234567890', 'AKTIF',     500000, 1200000, 2500000, '2024-01-15'),
  (2, 'KOP-0002', 'Rina Pertiwi', '3201020202910002', 'Jl. Asia Afrika No. 45, Bandung',   '081298765432', 'AKTIF',     500000,  900000, 1500000, '2024-03-20'),
  (3, 'KOP-0003', 'Dewi Lestari', '3201030303920003', 'Jl. Dago No. 78, Bandung',          '082134567891', 'AKTIF',     500000,  600000,  800000, '2024-06-10'),
  (4, '-',        'Joko Priyono', '3201040404930004', 'Jl. Braga No. 22, Bandung',         '085712345678', 'PENDING',        0,       0,       0, '2026-04-15'),
  (5, 'KOP-0004', 'Maya Sari',    '3201050505940005', 'Jl. Setiabudi No. 100, Bandung',    '081345678901', 'NON_AKTIF', 500000,  300000,       0, '2023-08-05');

INSERT INTO products (id, nama, kategori, harga, stok, batas_stok_minimum, deskripsi, foto_url, penjual_nama, status) VALUES
  (1, 'Beras Premium 5kg', 'Sembako',  75000, 50, 10, 'Beras premium kualitas terbaik, pulen dan wangi.', 'https://placehold.co/400x300?text=Beras',   'Koperasi',     'APPROVED'),
  (2, 'Minyak Goreng 2L',  'Sembako',  35000,  8, 15, 'Minyak goreng kemasan 2 liter.',                   'https://placehold.co/400x300?text=Minyak',  'Koperasi',     'APPROVED'),
  (3, 'Kopi Bubuk 250gr',  'Minuman',  28000, 30,  5, 'Kopi bubuk arabika asli Jawa Barat.',              'https://placehold.co/400x300?text=Kopi',    'Andi Wijaya',  'APPROVED'),
  (4, 'Keripik Singkong',  'Snack',    15000, 20,  5, 'Keripik singkong renyah produksi rumahan.',        'https://placehold.co/400x300?text=Keripik', 'Rina Pertiwi', 'PENDING'),
  (5, 'Gula Pasir 1kg',    'Sembako',  14000, 45, 10, 'Gula pasir putih bersih.',                         'https://placehold.co/400x300?text=Gula',    'Koperasi',     'APPROVED');

INSERT INTO orders (id, nomor_order, pembeli_nama, total_harga, fee_koperasi, status, metode_bayar, komplain_alasan, komplain_tanggal, komplain_bukti, created_at) VALUES
  (1, 'ORD-20260420-001', 'Andi Wijaya',  103000, 3090, 'SELESAI',  'Tunai',          NULL, NULL, NULL, '2026-04-20 10:30'),
  (2, 'ORD-20260421-002', 'Rina Pertiwi',  56000, 1680, 'DIKIRIM',  'QRIS',           NULL, NULL, NULL, '2026-04-21 14:15'),
  (3, 'ORD-20260422-003', 'Dewi Lestari',  35000, 1050, 'DIBAYAR',  'Saldo Anggota',  NULL, NULL, NULL, '2026-04-22 08:00'),
  (4, 'ORD-20260422-004', 'Andi Wijaya',   45000, 1350, 'DISPUTED', 'Tunai',          'Barang tidak sesuai — kemasan rusak saat sampai.', '2026-04-22', 'foto-bukti-rusak.jpg', '2026-04-22 11:20');

INSERT INTO order_items (order_id, product_id, product_nama, jumlah, harga_satuan, subtotal) VALUES
  (1, 1, 'Beras Premium 5kg', 1, 75000, 75000),
  (1, 5, 'Gula Pasir 1kg',    2, 14000, 28000),
  (2, 3, 'Kopi Bubuk 250gr',  2, 28000, 56000),
  (3, 2, 'Minyak Goreng 2L',  1, 35000, 35000),
  (4, 4, 'Keripik Singkong',  3, 15000, 45000);

INSERT INTO loans (id, member_id, nominal, tenor_bulan, bunga_persen, tujuan, sisa_pokok, status, tanggal_cair) VALUES
  (1, 1,  5000000, 10, 1.5, 'Modal usaha warung',  3000000, 'AKTIF', '2026-01-15'),
  (2, 2,  3000000,  6, 1.5, 'Biaya sekolah anak',  3000000, 'PENDING', NULL),
  (3, 3, 10000000, 12, 1.5, 'Renovasi rumah',            0, 'LUNAS', '2025-03-10');

INSERT INTO installments (loan_id, bulan_ke, jatuh_tempo, nominal_pokok, nominal_bunga, total_bayar, status, tanggal_bayar) VALUES
  (1, 1, '2026-02-15', 500000, 75000, 575000, 'DIBAYAR', '2026-02-14'),
  (1, 2, '2026-03-15', 500000, 67500, 567500, 'DIBAYAR', '2026-03-13'),
  (1, 3, '2026-04-15', 500000, 60000, 560000, 'DIBAYAR', '2026-04-15'),
  (1, 4, '2026-05-15', 500000, 52500, 552500, 'BELUM',   NULL),
  (1, 5, '2026-06-15', 500000, 45000, 545000, 'BELUM',   NULL);

INSERT INTO simpanan_transactions (id, member_id, jenis, tipe, nominal, keterangan, created_at) VALUES
  (1, 1, 'POKOK',    'MASUK', 500000, 'Simpanan pokok pendaftaran',            '2024-01-15'),
  (2, 1, 'WAJIB',    'MASUK', 100000, 'Simpanan wajib bulan Januari 2026',     '2026-01-05'),
  (3, 1, 'WAJIB',    'MASUK', 100000, 'Simpanan wajib bulan Februari 2026',    '2026-02-05'),
  (4, 1, 'SUKARELA', 'MASUK', 500000, 'Setoran sukarela',                      '2026-03-10'),
  (5, 2, 'POKOK',    'MASUK', 500000, 'Simpanan pokok pendaftaran',            '2024-03-20');

INSERT INTO stock_changes (id, product_id, tipe, jumlah, stok_setelah, keterangan, created_at) VALUES
  (1, 1, 'RESTOCK',          50,  50, 'Stok awal',           '2026-04-01'),
  (2, 2, 'RESTOCK',          30,  30, 'Stok awal',           '2026-04-01'),
  (3, 2, 'KELUAR_PENJUALAN', -22,  8, 'Penjualan akumulasi', '2026-04-15');

INSERT INTO journal_entries (id, tanggal, keterangan, akun_debit, akun_kredit, nominal, tipe_transaksi) VALUES
  (1, '2026-04-20', 'Penjualan POS ORD-20260420-001',           'Kas',               'Pendapatan Penjualan', 103000, 'POS'),
  (2, '2026-04-21', 'Penjualan POS ORD-20260421-002',           'Kas',               'Pendapatan Penjualan',  56000, 'POS'),
  (3, '2026-04-05', 'Simpanan wajib Andi Wijaya Apr 2026',      'Kas',               'Simpanan Wajib',       100000, 'SIMPANAN'),
  (4, '2026-04-15', 'Angsuran pinjaman Andi Wijaya bulan ke-3', 'Kas',               'Piutang Anggota',      560000, 'PINJAMAN'),
  (5, '2026-04-10', 'Pembelian ATK kantor',                     'Beban Operasional', 'Kas',                  250000, 'MANUAL');

-- ============ DOMAIN E-COMMERCE ============

INSERT INTO ecommerce_users (id, username, email, password_hash, role, is_seller_active, seller_rating, linked_koperasi_member_id, created_at) VALUES
  (1, 'andi_wijaya',  'anggota@koperasi.id',  '$2a$10$QKjvwsmNwv71BUgST6zunuXOZCimAtVM3irKMfVaS7tpkKjqzuVEW', 'BUYER',    true,  4.7, 1,    '2026-01-10'),
  (2, 'rina_pertiwi', 'rina@koperasi.id',     '$2a$10$KNCILIyV3lsCEmuDTuUo/u.nnvhpvHreuTBWHyEXiu78cBpsasvwG', 'BUYER',    true,  4.5, 2,    '2026-01-15'),
  (3, 'budi_santoso', 'owner@koperasi.id',    '$2a$10$ziC5K65cq/hZEtuxfWdmZOCTRCGd6.xSjFOnj0HdpdjTM8oB.mxtG', 'ADMIN',    false, 0,   NULL, '2026-01-01'),
  (4, 'siti_aminah',  'kasir@koperasi.id',    '$2a$10$afbed69AaUh6EcwisdNUUOPNeZynke4Zkv/bRi8M0YRKFWiWD.rT.', 'BUYER',    false, 0,   NULL, '2026-01-05'),
  (5, 'dewi_lestari', 'pengurus@koperasi.id', '$2a$10$J8n2AsGm6KXdIs1D6ZUfpuR6NUUVQQ1fsmTZ0lLlQOt17qgsVuvAK', 'PENGURUS', false, 0,   NULL, '2026-02-01');

INSERT INTO seller_profiles (seller_id, store_name, description, rating, response_time, total_sold, joined_at) VALUES
  (1, 'Toko Andi Jaya',     'Menjual berbagai kebutuhan pokok dan sembako berkualitas dengan harga terjangkau.', 4.7, '< 1 jam', 234, '2026-01-10'),
  (2, 'Rina Craft & Food',  'Produk makanan rumahan dan kerajinan tangan khas Bandung.',                         4.5, '< 2 jam', 156, '2026-01-15');

INSERT INTO ec_addresses (id, user_id, label, penerima, no_hp, alamat, kota, provinsi, kode_pos, is_default) VALUES
  (1, 3, 'Rumah',  'Budi Santoso', '081345678901', 'Jl. Cihampelas No. 55, Rt 03/Rw 05',  'Bandung', 'Jawa Barat', '40131', true),
  (2, 3, 'Kantor', 'Budi Santoso', '081345678901', 'Jl. Sudirman No. 100, Gedung A Lt. 3', 'Bandung', 'Jawa Barat', '40261', false),
  (3, 1, 'Rumah',  'Andi Wijaya',  '081234567890', 'Jl. Merdeka No. 12',                  'Bandung', 'Jawa Barat', '40117', true);

INSERT INTO ec_products (id, seller_id, seller_name, nama, deskripsi, kategori, harga, stok, berat, foto_url, rating, total_review, total_sold, status, created_at) VALUES
  (1,  1, 'Toko Andi Jaya',    'Beras Organik Premium 5kg',     'Beras organik dari petani lokal, pulen dan wangi. Ditanam tanpa pestisida.', 'Sembako', 85000, 50, 5000, 'https://placehold.co/400x400/2d1b69/e2e8f0?text=Beras+Organik',  4.8, 12, 87, 'APPROVED',         '2026-01-20'),
  (2,  1, 'Toko Andi Jaya',    'Minyak Goreng Kelapa 2L',       'Minyak goreng dari kelapa murni, lebih sehat dan tahan panas.',              'Sembako', 42000, 35, 2100, 'https://placehold.co/400x400/2d1b69/e2e8f0?text=Minyak+Goreng', 4.5,  8, 65, 'APPROVED',         '2026-01-22'),
  (3,  1, 'Toko Andi Jaya',    'Gula Aren Bubuk 500g',          'Gula aren asli dari Banten, cocok untuk kopi dan masakan.',                  'Sembako', 35000, 25,  500, 'https://placehold.co/400x400/2d1b69/e2e8f0?text=Gula+Aren',     4.9, 15, 42, 'APPROVED',         '2026-02-01'),
  (4,  1, 'Toko Andi Jaya',    'Kopi Arabika Jawa Barat 250g',  'Biji kopi arabika pilihan dari perkebunan di Jawa Barat.',                   'Minuman', 55000, 20,  250, 'https://placehold.co/400x400/2d1b69/e2e8f0?text=Kopi+Arabika',  4.7, 10, 30, 'APPROVED',         '2026-02-05'),
  (5,  1, 'Toko Andi Jaya',    'Teh Hijau Premium 100g',        'Teh hijau organik, kaya antioksidan dan menyegarkan.',                       'Minuman', 28000, 40,  100, 'https://placehold.co/400x400/2d1b69/e2e8f0?text=Teh+Hijau',     4.3,  5, 18, 'PENDING_APPROVAL', '2026-04-10'),
  (6,  2, 'Rina Craft & Food', 'Keripik Singkong Pedas 200g',   'Keripik singkong renyah level pedas, produksi rumahan berkualitas.',         'Snack',   18000, 60,  200, 'https://placehold.co/400x400/1b4d2e/e2e8f0?text=Keripik+Pedas', 4.6, 20, 98, 'APPROVED',         '2026-01-25'),
  (7,  2, 'Rina Craft & Food', 'Dodol Garut 300g',              'Dodol khas Garut, kenyal dan manis legit. Cocok untuk oleh-oleh.',           'Snack',   25000, 30,  300, 'https://placehold.co/400x400/1b4d2e/e2e8f0?text=Dodol+Garut',   4.4,  7, 45, 'APPROVED',         '2026-02-10'),
  (8,  2, 'Rina Craft & Food', 'Sambal Matah Bali 250ml',       'Sambal matah segar khas Bali, pedas dan harum.',                            'Bumbu',   22000, 45,  300, 'https://placehold.co/400x400/1b4d2e/e2e8f0?text=Sambal+Matah',  4.8, 11, 67, 'APPROVED',         '2026-02-15'),
  (9,  2, 'Rina Craft & Food', 'Tas Rajut Handmade',            'Tas rajut buatan tangan dari bahan katun premium.',                         'Fashion', 120000, 10, 200, 'https://placehold.co/400x400/1b4d2e/e2e8f0?text=Tas+Rajut',     4.9,  3, 12, 'APPROVED',         '2026-03-01'),
  (10, 2, 'Rina Craft & Food', 'Gelang Manik-manik Set',        'Set gelang manik-manik warna-warni, handmade.',                             'Fashion', 35000, 15,   50, 'https://placehold.co/400x400/1b4d2e/e2e8f0?text=Gelang+Manik',  4.2,  2,  8, 'PENDING_APPROVAL', '2026-04-15');

INSERT INTO shipping_options (id, nama, provider, estimasi, harga) VALUES
  (1, 'JNE Reguler', 'JNE',     '2-3 hari', 12000),
  (2, 'JNE YES',     'JNE',     '1 hari',   25000),
  (3, 'J&T Express', 'J&T',     '2-3 hari', 10000),
  (4, 'SiCepat BEST','SiCepat', '1-2 hari', 15000);

INSERT INTO vouchers (id, code, deskripsi, tipe_diskon, nilai_diskon, min_pembelian, maks_diskon, kuota, status, berlaku_sampai) VALUES
  (1, 'HEMAT10',   'Diskon 10% untuk semua produk',                       'PERCENT', 10,    50000, 20000, 50, 'ACTIVE', '2026-12-31'),
  (2, 'GRATIS15K', 'Potongan Rp 15.000 untuk pembelian min Rp 75.000',    'FIXED',   15000, 75000, 15000, 30, 'ACTIVE', '2026-06-30');

INSERT INTO ec_orders (id, nomor_order, buyer_id, seller_id, alamat_pengiriman, shipping_option, shipping_cost, subtotal, discount, points_used, total_harga, voucher_code, metode_bayar, status, resi_pengiriman, points_earned, created_at, updated_at) VALUES
  (1, 'EC-20260401-001', 3, 1, 'Jl. Cihampelas No. 55, Bandung', 'JNE Reguler', 12000, 155000,     0, 0, 167000, NULL,        'Transfer Bank', 'SELESAI', 'JNE1234567890', 155, '2026-04-01 10:30', '2026-04-04 14:00'),
  (2, 'EC-20260410-001', 3, 2, 'Jl. Cihampelas No. 55, Bandung', 'J&T Express', 10000,  76000,     0, 0,  86000, NULL,        'QRIS',          'DIKIRIM', 'JT2345678901',   76, '2026-04-10 14:15', '2026-04-12 09:00'),
  (3, 'EC-20260420-001', 1, 2, 'Jl. Merdeka No. 12, Bandung',    'SiCepat BEST',15000, 120000, 15000, 0, 120000, 'GRATIS15K', 'Transfer Bank', 'DIPROSES', '',               0, '2026-04-20 08:00', '2026-04-20 08:00');

INSERT INTO ec_order_items (order_id, product_id, product_nama, seller_id, jumlah, harga_satuan, subtotal) VALUES
  (1, 1, 'Beras Organik Premium 5kg',   1, 1,  85000,  85000),
  (1, 3, 'Gula Aren Bubuk 500g',        1, 2,  35000,  70000),
  (2, 6, 'Keripik Singkong Pedas 200g', 2, 3,  18000,  54000),
  (2, 8, 'Sambal Matah Bali 250ml',     2, 1,  22000,  22000),
  (3, 9, 'Tas Rajut Handmade',          2, 1, 120000, 120000);

INSERT INTO product_reviews (id, product_id, user_id, username, rating, komentar, created_at) VALUES
  (1, 1, 3, 'budi',         5, 'Berasnya bagus, pulen dan wangi. Keluarga suka sekali!',          '2026-03-15'),
  (2, 6, 3, 'budi_santoso', 4, 'Keripiknya enak dan renyah. Pedasnya pas!',                       '2026-03-20'),
  (3, 8, 1, 'andi',         5, 'Sambal matahnya segar banget, bumbunya terasa. Recommended!',     '2026-04-01');

INSERT INTO wishlists (id, user_id, product_id, created_at) VALUES
  (1, 3, 4, '2026-03-10'),
  (2, 3, 9, '2026-03-15'),
  (3, 1, 7, '2026-04-01');

INSERT INTO user_points (user_id, balance, total_earned, total_redeemed) VALUES
  (1, 5000, 5500, 500),
  (2, 3000, 3000, 0),
  (3, 1231, 1231, 0),
  (4, 0,    0,    0);

INSERT INTO points_transactions (id, user_id, tipe, amount, order_id, keterangan, created_at) VALUES
  (1, 3, 'EARN_PURCHASE',   155,  1,    'Pembelian order EC-20260401-001',     '2026-04-04'),
  (2, 3, 'EARN_PURCHASE',    76,  2,    'Pembelian order EC-20260410-001',     '2026-04-12'),
  (3, 1, 'EARN_PURCHASE',  5500,  NULL, 'Akumulasi pembelian sebelumnya',      '2026-03-01'),
  (4, 1, 'REDEEM_DISCOUNT', -500, NULL, 'Redeem poin untuk diskon belanja',    '2026-03-15'),
  (5, 2, 'EARN_PURCHASE',  3000,  NULL, 'Akumulasi pembelian sebelumnya',      '2026-03-10');

INSERT INTO shipment_events (id, order_id, status, lokasi, keterangan, created_at) VALUES
  (1, 1, 'DIKEMAS', 'Gudang Toko Andi Jaya, Bandung',    'Pesanan sedang dikemas',           '2026-04-01 15:00'),
  (2, 1, 'DIKIRIM', 'Sortir JNE Bandung',                'Paket diserahkan ke kurir JNE',    '2026-04-02 08:00'),
  (3, 1, 'SAMPAI',  'Bandung — Alamat Penerima',         'Paket diterima oleh Budi Prasetyo','2026-04-04 14:00'),
  (4, 2, 'DIKEMAS', 'Gudang Rina Craft & Food, Bandung', 'Pesanan sedang dikemas',           '2026-04-11 09:00'),
  (5, 2, 'DIKIRIM', 'Sortir J&T Bandung',                'Paket diserahkan ke kurir J&T',    '2026-04-12 09:00');

INSERT INTO audit_logs (id, action, user_id, username, resource, details, created_at) VALUES
  (1, 'APPROVE_PRODUCT', 3, 'budi_santoso', 'product:1', 'Approved: Beras Organik Premium 5kg',                                  '2026-01-21'),
  (2, 'APPROVE_PRODUCT', 3, 'budi_santoso', 'product:6', 'Approved: Keripik Singkong Pedas 200g',                                '2026-01-26'),
  (3, 'LINK_KOPERASI',   1, 'andi_wijaya',  'member:1',  'Linked e-commerce account to koperasi member Andi Wijaya (KOP-0001)',  '2026-01-10');

-- Setel ulang sequence agar insert berikutnya tidak bentrok dengan id eksplisit di atas.
SELECT setval(pg_get_serial_sequence('users','id'),                 (SELECT MAX(id) FROM users));
SELECT setval(pg_get_serial_sequence('members','id'),               (SELECT MAX(id) FROM members));
SELECT setval(pg_get_serial_sequence('products','id'),              (SELECT MAX(id) FROM products));
SELECT setval(pg_get_serial_sequence('orders','id'),                (SELECT MAX(id) FROM orders));
SELECT setval(pg_get_serial_sequence('order_items','id'),           (SELECT MAX(id) FROM order_items));
SELECT setval(pg_get_serial_sequence('loans','id'),                 (SELECT MAX(id) FROM loans));
SELECT setval(pg_get_serial_sequence('installments','id'),          (SELECT MAX(id) FROM installments));
SELECT setval(pg_get_serial_sequence('simpanan_transactions','id'), (SELECT MAX(id) FROM simpanan_transactions));
SELECT setval(pg_get_serial_sequence('stock_changes','id'),         (SELECT MAX(id) FROM stock_changes));
SELECT setval(pg_get_serial_sequence('journal_entries','id'),       (SELECT MAX(id) FROM journal_entries));
SELECT setval(pg_get_serial_sequence('ecommerce_users','id'),       (SELECT MAX(id) FROM ecommerce_users));
SELECT setval(pg_get_serial_sequence('ec_addresses','id'),          (SELECT MAX(id) FROM ec_addresses));
SELECT setval(pg_get_serial_sequence('ec_products','id'),           (SELECT MAX(id) FROM ec_products));
SELECT setval(pg_get_serial_sequence('shipping_options','id'),      (SELECT MAX(id) FROM shipping_options));
SELECT setval(pg_get_serial_sequence('vouchers','id'),              (SELECT MAX(id) FROM vouchers));
SELECT setval(pg_get_serial_sequence('ec_orders','id'),             (SELECT MAX(id) FROM ec_orders));
SELECT setval(pg_get_serial_sequence('ec_order_items','id'),        (SELECT MAX(id) FROM ec_order_items));
SELECT setval(pg_get_serial_sequence('product_reviews','id'),       (SELECT MAX(id) FROM product_reviews));
SELECT setval(pg_get_serial_sequence('wishlists','id'),             (SELECT MAX(id) FROM wishlists));
SELECT setval(pg_get_serial_sequence('points_transactions','id'),   (SELECT MAX(id) FROM points_transactions));
SELECT setval(pg_get_serial_sequence('shipment_events','id'),       (SELECT MAX(id) FROM shipment_events));
SELECT setval(pg_get_serial_sequence('audit_logs','id'),            (SELECT MAX(id) FROM audit_logs));
