-- Rollback skema domain E-Commerce. Urutan terbalik agar FK aman.
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS shipment_events;
DROP TABLE IF EXISTS points_transactions;
DROP TABLE IF EXISTS user_points;
DROP TABLE IF EXISTS wishlists;
DROP TABLE IF EXISTS product_reviews;
DROP TABLE IF EXISTS ec_order_items;
DROP TABLE IF EXISTS ec_orders;
DROP TABLE IF EXISTS vouchers;
DROP TABLE IF EXISTS shipping_options;
DROP TABLE IF EXISTS ec_products;
DROP TABLE IF EXISTS ec_addresses;
DROP TABLE IF EXISTS seller_profiles;
DROP TABLE IF EXISTS ecommerce_users;
