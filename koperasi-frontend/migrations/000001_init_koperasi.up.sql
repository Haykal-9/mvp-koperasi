-- Fase 1 — Skema domain Koperasi (PRD §5.2).
-- Konvensi: tabel snake_case jamak; id BIGSERIAL PK; uang NUMERIC(15,2); waktu TIMESTAMPTZ.

CREATE TABLE users (
  id            BIGSERIAL PRIMARY KEY,
  email         VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role          VARCHAR(20)  NOT NULL CHECK (role IN ('OWNER','KASIR','ANGGOTA','PENGURUS')),
  nama          VARCHAR(150) NOT NULL,
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE members (
  id                BIGSERIAL PRIMARY KEY,
  nomor_anggota     VARCHAR(20) UNIQUE NOT NULL,
  nama              VARCHAR(150) NOT NULL,
  nik               VARCHAR(20),
  alamat            TEXT,
  no_hp             VARCHAR(20),
  status            VARCHAR(15) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','AKTIF','NON_AKTIF')),
  simpanan_pokok    NUMERIC(15,2) NOT NULL DEFAULT 0,
  simpanan_wajib    NUMERIC(15,2) NOT NULL DEFAULT 0,
  simpanan_sukarela NUMERIC(15,2) NOT NULL DEFAULT 0,
  tanggal_masuk     DATE,
  user_id           BIGINT REFERENCES users(id)
);

CREATE TABLE products (
  id                  BIGSERIAL PRIMARY KEY,
  nama                VARCHAR(150) NOT NULL,
  kategori            VARCHAR(50),
  harga               NUMERIC(15,2) NOT NULL,
  stok                INT NOT NULL DEFAULT 0,
  batas_stok_minimum  INT NOT NULL DEFAULT 0,
  deskripsi           TEXT,
  foto_url            TEXT,
  penjual_nama        VARCHAR(150),
  status              VARCHAR(15) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','APPROVED','REJECTED'))
);

CREATE TABLE stock_changes (
  id           BIGSERIAL PRIMARY KEY,
  product_id   BIGINT NOT NULL REFERENCES products(id),
  tipe         VARCHAR(20) NOT NULL CHECK (tipe IN ('RESTOCK','KELUAR_PENJUALAN','KOREKSI')),
  jumlah       INT NOT NULL,
  stok_setelah INT NOT NULL,
  keterangan   TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE orders (
  id               BIGSERIAL PRIMARY KEY,
  nomor_order      VARCHAR(30) UNIQUE NOT NULL,
  pembeli_nama     VARCHAR(150),
  total_harga      NUMERIC(15,2) NOT NULL,
  fee_koperasi     NUMERIC(15,2) NOT NULL DEFAULT 0,
  status           VARCHAR(15) NOT NULL DEFAULT 'PENDING'
                   CHECK (status IN ('PENDING','DIBAYAR','DIKIRIM','SELESAI','BATAL','DISPUTED')),
  metode_bayar     VARCHAR(30),
  komplain_alasan  TEXT,
  komplain_tanggal DATE,
  komplain_bukti   TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
  id           BIGSERIAL PRIMARY KEY,
  order_id     BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id   BIGINT REFERENCES products(id),
  product_nama VARCHAR(150),
  jumlah       INT NOT NULL,
  harga_satuan NUMERIC(15,2) NOT NULL,
  subtotal     NUMERIC(15,2) NOT NULL
);

CREATE TABLE loans (
  id           BIGSERIAL PRIMARY KEY,
  member_id    BIGINT REFERENCES members(id),
  nominal      NUMERIC(15,2) NOT NULL CHECK (nominal >= 500000),
  tenor_bulan  INT NOT NULL CHECK (tenor_bulan BETWEEN 3 AND 24),
  bunga_persen NUMERIC(5,2) NOT NULL DEFAULT 1.5,
  tujuan       TEXT,
  sisa_pokok   NUMERIC(15,2) NOT NULL,
  status       VARCHAR(15) NOT NULL DEFAULT 'PENDING'
               CHECK (status IN ('PENDING','DISETUJUI','AKTIF','LUNAS','DITOLAK')),
  tanggal_cair DATE
);

CREATE TABLE installments (
  id            BIGSERIAL PRIMARY KEY,
  loan_id       BIGINT NOT NULL REFERENCES loans(id) ON DELETE CASCADE,
  bulan_ke      INT NOT NULL,
  jatuh_tempo   DATE NOT NULL,
  nominal_pokok NUMERIC(15,2) NOT NULL,
  nominal_bunga NUMERIC(15,2) NOT NULL,
  total_bayar   NUMERIC(15,2) NOT NULL,
  status        VARCHAR(10) NOT NULL DEFAULT 'BELUM' CHECK (status IN ('BELUM','DIBAYAR','OVERDUE')),
  tanggal_bayar DATE
);

CREATE TABLE simpanan_transactions (
  id         BIGSERIAL PRIMARY KEY,
  member_id  BIGINT NOT NULL REFERENCES members(id),
  jenis      VARCHAR(10) NOT NULL CHECK (jenis IN ('POKOK','WAJIB','SUKARELA')),
  tipe       VARCHAR(10) NOT NULL CHECK (tipe IN ('MASUK','KELUAR')),
  nominal    NUMERIC(15,2) NOT NULL,
  keterangan TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE journal_entries (
  id             BIGSERIAL PRIMARY KEY,
  tanggal        DATE NOT NULL,
  keterangan     TEXT,
  akun_debit     VARCHAR(100) NOT NULL,
  akun_kredit    VARCHAR(100) NOT NULL,
  nominal        NUMERIC(15,2) NOT NULL,
  tipe_transaksi VARCHAR(15) NOT NULL CHECK (tipe_transaksi IN ('SIMPANAN','POS','PINJAMAN','MANUAL'))
);

-- Indeks pada FK yang sering difilter (PRD §5.4).
CREATE INDEX idx_stock_changes_product   ON stock_changes(product_id);
CREATE INDEX idx_order_items_order       ON order_items(order_id);
CREATE INDEX idx_loans_member            ON loans(member_id);
CREATE INDEX idx_installments_loan       ON installments(loan_id);
CREATE INDEX idx_simpanan_member         ON simpanan_transactions(member_id);
