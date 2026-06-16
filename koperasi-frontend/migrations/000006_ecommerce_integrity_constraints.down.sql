DROP INDEX IF EXISTS ux_product_reviews_user_product;

ALTER TABLE ec_products
  DROP CONSTRAINT IF EXISTS ec_products_harga_positive,
  DROP CONSTRAINT IF EXISTS ec_products_stok_non_negative,
  DROP CONSTRAINT IF EXISTS ec_products_berat_positive;
