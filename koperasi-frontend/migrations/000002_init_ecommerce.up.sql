-- Fase 1 — Skema domain E-Commerce SmartMart (PRD §5.2).
-- Catatan: vouchers didefinisikan sebelum ec_orders (ec_orders.voucher_code -> vouchers.code).

CREATE TABLE ecommerce_users (
  id                        BIGSERIAL PRIMARY KEY,
  username                  VARCHAR(50) UNIQUE NOT NULL,
  email                     VARCHAR(255) UNIQUE NOT NULL,
  password_hash             VARCHAR(255) NOT NULL,
  role                      VARCHAR(20) NOT NULL CHECK (role IN ('BUYER','SELLER','ADMIN','PENGURUS','KASIR')),
  is_seller_active          BOOLEAN NOT NULL DEFAULT false,
  seller_rating             NUMERIC(3,2) NOT NULL DEFAULT 0,
  linked_koperasi_member_id BIGINT REFERENCES members(id),
  created_at                TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE seller_profiles (
  seller_id     BIGINT PRIMARY KEY REFERENCES ecommerce_users(id),
  store_name    VARCHAR(150) NOT NULL,
  description   TEXT,
  rating        NUMERIC(3,2) NOT NULL DEFAULT 0,
  response_time VARCHAR(30),
  total_sold    INT NOT NULL DEFAULT 0,
  joined_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ec_addresses (
  id        BIGSERIAL PRIMARY KEY,
  user_id   BIGINT NOT NULL REFERENCES ecommerce_users(id),
  label     VARCHAR(30),
  penerima  VARCHAR(150),
  no_hp     VARCHAR(20),
  alamat    TEXT,
  kota      VARCHAR(80),
  provinsi  VARCHAR(80),
  kode_pos  VARCHAR(10),
  is_default BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE ec_products (
  id           BIGSERIAL PRIMARY KEY,
  seller_id    BIGINT NOT NULL REFERENCES ecommerce_users(id),
  seller_name  VARCHAR(150),
  nama         VARCHAR(150) NOT NULL,
  deskripsi    TEXT,
  kategori     VARCHAR(50),
  harga        NUMERIC(15,2) NOT NULL,
  stok         INT NOT NULL DEFAULT 0,
  berat        INT NOT NULL DEFAULT 0,
  foto_url     TEXT,
  rating       NUMERIC(3,2) NOT NULL DEFAULT 0,
  total_review INT NOT NULL DEFAULT 0,
  total_sold   INT NOT NULL DEFAULT 0,
  status       VARCHAR(20) NOT NULL DEFAULT 'PENDING_APPROVAL'
               CHECK (status IN ('PENDING_APPROVAL','APPROVED','REJECTED','ARCHIVED')),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE shipping_options (
  id       BIGSERIAL PRIMARY KEY,
  nama     VARCHAR(50) NOT NULL,
  provider VARCHAR(30),
  estimasi VARCHAR(30),
  harga    NUMERIC(15,2) NOT NULL
);

CREATE TABLE vouchers (
  id             BIGSERIAL PRIMARY KEY,
  code           VARCHAR(30) UNIQUE NOT NULL,
  deskripsi      TEXT,
  tipe_diskon    VARCHAR(10) NOT NULL CHECK (tipe_diskon IN ('PERCENT','FIXED')),
  nilai_diskon   NUMERIC(15,2) NOT NULL,
  min_pembelian  NUMERIC(15,2) NOT NULL DEFAULT 0,
  maks_diskon    NUMERIC(15,2) NOT NULL DEFAULT 0,
  kuota          INT NOT NULL DEFAULT 0,
  status         VARCHAR(10) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE','EXPIRED','USED_UP')),
  berlaku_sampai DATE
);

CREATE TABLE ec_orders (
  id                BIGSERIAL PRIMARY KEY,
  nomor_order       VARCHAR(30) UNIQUE NOT NULL,
  buyer_id          BIGINT NOT NULL REFERENCES ecommerce_users(id),
  seller_id         BIGINT NOT NULL REFERENCES ecommerce_users(id),
  alamat_pengiriman TEXT,
  shipping_option   VARCHAR(50),
  shipping_cost     NUMERIC(15,2) NOT NULL DEFAULT 0,
  subtotal          NUMERIC(15,2) NOT NULL,
  discount          NUMERIC(15,2) NOT NULL DEFAULT 0,
  points_used       NUMERIC(15,2) NOT NULL DEFAULT 0,
  total_harga       NUMERIC(15,2) NOT NULL,
  voucher_code      VARCHAR(30) REFERENCES vouchers(code),
  metode_bayar      VARCHAR(30),
  status            VARCHAR(15) NOT NULL DEFAULT 'PENDING'
                    CHECK (status IN ('PENDING','DIBAYAR','DIPROSES','DIKIRIM','SELESAI','BATAL')),
  resi_pengiriman   VARCHAR(50),
  points_earned     NUMERIC(15,2) NOT NULL DEFAULT 0,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ec_order_items (
  id           BIGSERIAL PRIMARY KEY,
  order_id     BIGINT NOT NULL REFERENCES ec_orders(id) ON DELETE CASCADE,
  product_id   BIGINT REFERENCES ec_products(id),
  product_nama VARCHAR(150),
  seller_id    BIGINT,
  jumlah       INT NOT NULL,
  harga_satuan NUMERIC(15,2) NOT NULL,
  subtotal     NUMERIC(15,2) NOT NULL
);

CREATE TABLE product_reviews (
  id         BIGSERIAL PRIMARY KEY,
  product_id BIGINT NOT NULL REFERENCES ec_products(id),
  user_id    BIGINT NOT NULL REFERENCES ecommerce_users(id),
  username   VARCHAR(50),
  rating     INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
  komentar   TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE wishlists (
  id         BIGSERIAL PRIMARY KEY,
  user_id    BIGINT NOT NULL REFERENCES ecommerce_users(id),
  product_id BIGINT NOT NULL REFERENCES ec_products(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, product_id)
);

CREATE TABLE user_points (
  user_id        BIGINT PRIMARY KEY REFERENCES ecommerce_users(id),
  balance        NUMERIC(15,2) NOT NULL DEFAULT 0,
  total_earned   NUMERIC(15,2) NOT NULL DEFAULT 0,
  total_redeemed NUMERIC(15,2) NOT NULL DEFAULT 0
);

CREATE TABLE points_transactions (
  id         BIGSERIAL PRIMARY KEY,
  user_id    BIGINT NOT NULL REFERENCES ecommerce_users(id),
  tipe       VARCHAR(20) NOT NULL CHECK (tipe IN ('EARN_PURCHASE','REDEEM_VOUCHER','REDEEM_DISCOUNT','CONVERT_SIMPANAN')),
  amount     NUMERIC(15,2) NOT NULL,
  order_id   BIGINT REFERENCES ec_orders(id),
  keterangan TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE shipment_events (
  id         BIGSERIAL PRIMARY KEY,
  order_id   BIGINT NOT NULL REFERENCES ec_orders(id),
  status     VARCHAR(15) NOT NULL CHECK (status IN ('DIKEMAS','DIKIRIM','TRANSIT','SAMPAI')),
  lokasi     VARCHAR(120),
  keterangan TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE audit_logs (
  id         BIGSERIAL PRIMARY KEY,
  action     VARCHAR(40) NOT NULL,
  user_id    BIGINT REFERENCES ecommerce_users(id),
  username   VARCHAR(50),
  resource   VARCHAR(50),
  details    TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indeks pada FK yang sering difilter (PRD §5.4).
CREATE INDEX idx_ec_products_seller_status ON ec_products(seller_id, status);
CREATE INDEX idx_ec_orders_buyer           ON ec_orders(buyer_id);
CREATE INDEX idx_ec_orders_seller          ON ec_orders(seller_id);
CREATE INDEX idx_ec_order_items_order      ON ec_order_items(order_id);
CREATE INDEX idx_points_tx_user            ON points_transactions(user_id);
CREATE INDEX idx_reviews_product           ON product_reviews(product_id);
CREATE INDEX idx_shipment_events_order     ON shipment_events(order_id);
CREATE INDEX idx_ec_addresses_user         ON ec_addresses(user_id);
