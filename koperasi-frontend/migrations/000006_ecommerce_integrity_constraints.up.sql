-- Menutup invariant e-commerce yang sebelumnya hanya dijaga di handler/service.

-- Jika ada review dobel dari double submit sebelum constraint ini, simpan yang
-- paling awal lalu hitung ulang agregat rating produk.
DELETE FROM product_reviews newer
USING product_reviews older
WHERE newer.user_id = older.user_id
  AND newer.product_id = older.product_id
  AND newer.id > older.id;

UPDATE ec_products p
SET total_review = stats.total_review,
    rating = stats.rating
FROM (
  SELECT product_id, count(*) AS total_review, COALESCE(avg(rating), 0) AS rating
  FROM product_reviews
  GROUP BY product_id
) stats
WHERE p.id = stats.product_id;

UPDATE ec_products p
SET total_review = 0,
    rating = 0
WHERE NOT EXISTS (
  SELECT 1 FROM product_reviews r WHERE r.product_id = p.id
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_product_reviews_user_product
  ON product_reviews (user_id, product_id);

-- Normalisasi nilai produk lama sebelum check constraint dipasang.
UPDATE ec_products SET harga = 1 WHERE harga <= 0;
UPDATE ec_products SET stok = 0 WHERE stok < 0;
UPDATE ec_products SET berat = 1 WHERE berat <= 0;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'ec_products_harga_positive'
      AND conrelid = 'ec_products'::regclass
  ) THEN
    ALTER TABLE ec_products
      ADD CONSTRAINT ec_products_harga_positive CHECK (harga > 0);
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'ec_products_stok_non_negative'
      AND conrelid = 'ec_products'::regclass
  ) THEN
    ALTER TABLE ec_products
      ADD CONSTRAINT ec_products_stok_non_negative CHECK (stok >= 0);
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'ec_products_berat_positive'
      AND conrelid = 'ec_products'::regclass
  ) THEN
    ALTER TABLE ec_products
      ADD CONSTRAINT ec_products_berat_positive CHECK (berat > 0);
  END IF;
END $$;
