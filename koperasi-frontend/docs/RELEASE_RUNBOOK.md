# Runbook Pengujian, Rilis, Backup, dan Restore

Dokumen ini menutup Fase 7 backend: test inti, konfigurasi deployment, dan prosedur backup/restore PostgreSQL.

## 1. Variabel Environment

Gunakan `.env.example` sebagai acuan.

- `DATABASE_URL`: URL aplikasi. Untuk Supabase/Neon serverless, gunakan transaction pooler bila tersedia.
- `MIGRATION_DATABASE_URL`: URL migrasi. Gunakan direct/session pooler karena `golang-migrate` memakai advisory lock level sesi.
- `SESSION_SECRET`: string acak minimal 32 karakter.
- `SESSION_SECURE`: `true` untuk HTTPS/production.
- `PORT`: port lokal, default `8080`.
- `TEST_DATABASE_URL`: database uji terpisah untuk integration test. Jangan memakai database produksi.

## 2. Verifikasi Sebelum Rilis

Jalankan dari folder `koperasi-frontend`.

```powershell
go test ./...
go vet ./...
go build ./cmd/web ./api
```

Integration test PostgreSQL akan otomatis `skip` bila `TEST_DATABASE_URL` kosong. Untuk menjalankannya:

```powershell
$env:TEST_DATABASE_URL="postgresql://USER:PASSWORD@HOST:5432/koperasi_test?sslmode=require"
go test ./core/integration -count=1 -v
```

Integration test membuat schema sementara, menjalankan semua migrasi, menguji checkout, bayar angsuran, dan konversi poin ke simpanan, lalu menghapus schema tersebut.

## 3. Migrasi Database

Mode lokal atau server biasa:

```powershell
$env:DATABASE_URL="postgresql://USER:PASSWORD@HOST:6543/postgres?sslmode=require"
$env:MIGRATION_DATABASE_URL="postgresql://USER:PASSWORD@HOST:5432/postgres?sslmode=require"
go run ./cmd/web
```

Catatan:

- Binary web menjalankan migrasi saat startup bila `DATABASE_URL` terisi.
- Pada serverless, jangan mengandalkan fungsi request untuk menjalankan migrasi. Jalankan migrasi dari job/terminal deploy menggunakan `cmd/web` atau utilitas migrasi terpisah sebelum traffic diarahkan.
- File `000006_ecommerce_integrity_constraints` merapikan data review duplikat dan nilai produk invalid sebelum menambah constraint, jadi backup dulu sebelum diterapkan ke database yang berisi data nyata.

## 4. Deployment

Checklist minimum production:

- Set `DATABASE_URL` ke pooler aplikasi.
- Set `MIGRATION_DATABASE_URL` ke direct/session URL.
- Set `SESSION_SECRET` kuat dan unik.
- Set `SESSION_SECURE=true`.
- Pastikan `vercel.json` mengarahkan semua route ke `/api` untuk deployment Vercel.
- Jalankan `go test ./...`, `go vet ./...`, dan `go build ./cmd/web ./api` di CI/deploy step.

## 5. Backup

Backup logical dengan `pg_dump`:

```powershell
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$env:PGPASSWORD="PASSWORD"
pg_dump --format=custom --no-owner --no-privileges --file="backup-$timestamp.dump" "postgresql://USER:PASSWORD@HOST:5432/postgres?sslmode=require"
```

Simpan hasil dump di lokasi aman yang tidak ikut repository, misalnya storage terenkripsi atau fitur backup provider database.

## 6. Restore

Restore ke database kosong atau database staging terlebih dahulu:

```powershell
$env:PGPASSWORD="PASSWORD"
pg_restore --clean --if-exists --no-owner --no-privileges --dbname="postgresql://USER:PASSWORD@HOST:5432/postgres?sslmode=require" "backup-YYYYMMDD-HHMMSS.dump"
```

Setelah restore:

```powershell
go test ./...
go run ./cmd/web
```

Lakukan smoke test HTTP pada login, dashboard, checkout, pembayaran angsuran, dan konversi poin.
