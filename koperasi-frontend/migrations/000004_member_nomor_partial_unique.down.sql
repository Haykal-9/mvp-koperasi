-- Kembalikan ke UNIQUE penuh.
DROP INDEX IF EXISTS members_nomor_anggota_key;
ALTER TABLE members ADD CONSTRAINT members_nomor_anggota_key UNIQUE (nomor_anggota);
