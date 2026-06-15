-- Rollback skema domain Koperasi. Urutan terbalik agar FK aman.
DROP TABLE IF EXISTS journal_entries;
DROP TABLE IF EXISTS simpanan_transactions;
DROP TABLE IF EXISTS installments;
DROP TABLE IF EXISTS loans;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS stock_changes;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS members;
DROP TABLE IF EXISTS users;
