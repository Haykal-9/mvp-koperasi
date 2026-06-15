-- Perbaikan: anggota PENDING memakai nomor_anggota '-' sebagai placeholder,
-- sehingga UNIQUE penuh menolak >1 anggota PENDING. Ganti ke UNIQUE parsial
-- (abaikan '-'), tetap menjamin keunikan nomor KOP-#### yang asli.
ALTER TABLE members DROP CONSTRAINT members_nomor_anggota_key;
CREATE UNIQUE INDEX members_nomor_anggota_key ON members (nomor_anggota) WHERE nomor_anggota <> '-';
