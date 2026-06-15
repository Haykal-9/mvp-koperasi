-- Fase 6: satu member koperasi hanya boleh ditautkan ke satu akun E-Commerce.
-- NULL tetap diizinkan agar banyak akun EC bisa belum tertaut.
CREATE UNIQUE INDEX IF NOT EXISTS ecommerce_users_linked_member_unique
ON ecommerce_users (linked_koperasi_member_id)
WHERE linked_koperasi_member_id IS NOT NULL;
