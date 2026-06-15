-- Rollback data seed. TRUNCATE ... RESTART IDENTITY mengosongkan tabel & mereset sequence.
-- CASCADE menghapus baris anak (order_items, installments, dst).
TRUNCATE
  audit_logs, shipment_events, points_transactions, user_points, wishlists,
  product_reviews, ec_order_items, ec_orders, vouchers, shipping_options,
  ec_products, ec_addresses, seller_profiles, ecommerce_users,
  journal_entries, simpanan_transactions, installments, loans,
  order_items, orders, stock_changes, products, members, users
  RESTART IDENTITY CASCADE;
