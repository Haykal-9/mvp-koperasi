# 📋 MVP Koperasi — Planning Fullstack (Backend API + Frontend SSR)

## 🎯 Tujuan
Membuat MVP **lengkap** aplikasi koperasi dengan **Backend REST API** dan **Frontend SSR**,
keduanya menggunakan **Golang**. Backend menyediakan API, Frontend mengonsumsi API tersebut.

---

## 👥 Pembagian Tim (4 Orang)

| No | Nama | NIM | PIC |
|----|------|-----|-----|
| 1 | M Haykal Lazuardy | 607012400068 | Auth + Member (M1) |
| 2 | Ega Fiandra Pratama | 607012400032 | Product + Stok (M3.1, M3.3) |
| 3 | Raynaldi Paniroean Panjaitan | 607012400099 | Order + Payment (M3.2, M3.4) |
| 4 | Diki Alif Taufik | 607012400005 | Loan + Finance (M4, M2) |

> Setiap PIC bertanggung jawab atas **backend API + frontend halaman** modulnya.

---

## 🏗️ Tech Stack

| Komponen | Teknologi |
|----------|-----------|
| Bahasa | Golang |
| Backend Framework | Gin |
| Database | PostgreSQL |
| ORM | GORM |
| Auth | JWT (golang-jwt/jwt/v5) |
| Password | bcrypt |
| Frontend Rendering | html/template |
| CSS | Bootstrap 5 (CDN) |
| Config | Viper + .env |

---

## 📁 Struktur Folder (Monorepo)

```
koperasi-app/
├── cmd/
│   ├── api/
│   │   └── main.go              # Entry point Backend API (:8081)
│   └── web/
│       └── main.go              # Entry point Frontend Web (:8080)
├── internal/
│   ├── api/                     # Backend API layer
│   │   ├── auth/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── model.go
│   │   ├── member/
│   │   ├── product/
│   │   ├── order/
│   │   ├── loan/
│   │   └── finance/
│   └── web/                     # Frontend Web layer
│       └── handler/
│           ├── auth.go
│           ├── member.go
│           ├── product.go
│           ├── order.go
│           ├── loan.go
│           └── finance.go
├── pkg/
│   ├── middleware/               # JWT, CORS, logger
│   ├── response/                 # Standard JSON response
│   ├── database/                 # DB connection (GORM)
│   └── httpclient/               # HTTP client untuk frontend call API
├── templates/                    # HTML templates
│   ├── layouts/
│   │   ├── base.html
│   │   └── auth.html
│   ├── auth/
│   ├── member/
│   ├── product/
│   ├── order/
│   ├── loan/
│   └── finance/
├── static/
│   ├── css/style.css
│   └── js/app.js
├── migrations/                   # SQL migration files
├── config/
├── .env
├── go.mod
└── Makefile
```

---

## 🔄 Step-by-Step Pengerjaan

### STEP 1 — Foundation (Semua, 2-3 hari)

**Backend:**
1. `go mod init koperasi-app`
2. Setup Gin router di `cmd/api/main.go` (port 8081)
3. Setup PostgreSQL + GORM di `pkg/database/`
4. Buat `pkg/response/` — standard JSON response
5. Buat `pkg/middleware/` — JWT auth, CORS, logger
6. Buat semua migration SQL (tabel: users, members, products, orders, order_items, loans, loan_installments, simpanan_transactions, journal_entries)
7. Jalankan migration

**Frontend:**
8. Setup Gin router di `cmd/web/main.go` (port 8080)
9. Buat `templates/layouts/base.html` + `auth.html`
10. Buat `static/css/style.css`
11. Buat `pkg/httpclient/` — HTTP client wrapper untuk call API

---

### STEP 2 — Auth & Member (PIC: Haykal, 4-5 hari)

**Backend API:**
```
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
GET  /api/v1/members
GET  /api/v1/members/:id
POST /api/v1/members
PUT  /api/v1/members/:id/approve
PUT  /api/v1/members/:id/reject
POST /api/v1/members/:id/simpanan
GET  /api/v1/members/:id/simpanan
GET  /api/v1/members/:id/saldo
POST /api/v1/members/:id/resign
PUT  /api/v1/members/:id/resign/approve
```

**Frontend Pages:**
- `/login`, `/register`, `/logout`
- `/dashboard` — statistik
- `/members` — daftar anggota + approve/reject
- `/members/:id` — detail
- `/members/register` — form pendaftaran
- `/members/:id/simpanan` — saldo + bayar + riwayat

**Proses Bisnis:** M1.1, M1.2, M1.3

---

### STEP 3 — Product & Stok (PIC: Ega, 4-5 hari)

**Backend API:**
```
GET    /api/v1/products
GET    /api/v1/products/:id
POST   /api/v1/products
PUT    /api/v1/products/:id
DELETE /api/v1/products/:id
PUT    /api/v1/products/:id/approve
PUT    /api/v1/products/:id/reject
PUT    /api/v1/products/:id/stock
GET    /api/v1/products/:id/stock-log
```

**Frontend Pages:**
- `/products` — katalog grid
- `/products/:id` — detail + tambah ke keranjang
- `/products/create` — form tambah
- `/products/review` — review pending
- `/products/:id/stock` — manajemen stok

**Proses Bisnis:** M3.1, M3.3

---

### STEP 4 — Order & Payment (PIC: Raynaldi, 5-6 hari)

**Backend API:**
```
POST   /api/v1/orders
GET    /api/v1/orders
GET    /api/v1/orders/:id
PUT    /api/v1/orders/:id/pay
PUT    /api/v1/orders/:id/ship
PUT    /api/v1/orders/:id/complete
PUT    /api/v1/orders/:id/cancel
POST   /api/v1/orders/:id/complain
PUT    /api/v1/orders/:id/complain/resolve
PUT    /api/v1/orders/:id/refund
```

**Frontend Pages:**
- `/cart`, `/checkout`
- `/orders` — riwayat
- `/orders/:id` — detail + timeline status
- `/orders/:id/complain` — form komplain
- `/orders/complaints` — daftar komplain (OWNER)

**Proses Bisnis:** M3.2, M3.4

---

### STEP 5 — Loan & Finance (PIC: Diki, 5-6 hari)

**Backend API:**
```
POST   /api/v1/loans
GET    /api/v1/loans
GET    /api/v1/loans/:id
PUT    /api/v1/loans/:id/approve
PUT    /api/v1/loans/:id/reject
PUT    /api/v1/loans/:id/disburse
GET    /api/v1/loans/:id/installments
POST   /api/v1/loans/:id/pay
GET    /api/v1/finance/journals
POST   /api/v1/finance/journals
GET    /api/v1/finance/summary
```

**Frontend Pages:**
- `/loans/apply`, `/loans`, `/loans/:id`
- `/loans/:id/pay` — bayar angsuran
- `/finance/journals`, `/finance/journals/create`
- `/finance/summary`

**Proses Bisnis:** M4.1, M4.2, M2.1, M2.2

---

### STEP 6 — Integrasi & Testing (Semua, 3-4 hari)

1. Integrasi lintas modul (order → stok, payment → jurnal)
2. Test semua endpoint API via Postman
3. Test semua halaman frontend
4. Fix bugs
5. Dokumentasi API + README

---

## ⏱️ Timeline

| Step | Durasi | Deadline |
|------|--------|----------|
| Step 1 — Foundation | 2-3 hari | Hari ke-3 |
| Step 2 — Auth & Member | 4-5 hari | Hari ke-8 |
| Step 3 — Product & Stok | 4-5 hari | Hari ke-8 |
| Step 4 — Order & Payment | 5-6 hari | Hari ke-14 |
| Step 5 — Loan & Finance | 5-6 hari | Hari ke-14 |
| Step 6 — Integrasi | 3-4 hari | Hari ke-18 |

> Step 2 & 3 paralel. Step 4 & 5 paralel (tapi Step 4 butuh Step 3).

**Total estimasi: ~18 hari kerja**

---

## 🤖 Prompt untuk AI Lain

### Master Prompt

```
Saya membuat MVP Fullstack Koperasi dengan Golang (Backend API + Frontend SSR).

TECH STACK:
- Golang + Gin
- PostgreSQL + GORM
- JWT auth
- html/template + Bootstrap 5

ARSITEKTUR:
- cmd/api/main.go — Backend API (port 8081)
- cmd/web/main.go — Frontend Web (port 8080)
- internal/api/<modul>/ — handler, service, repository, model per modul
- internal/web/handler/ — frontend handler per modul
- templates/<modul>/ — HTML templates
- pkg/ — middleware, response, database, httpclient

DATABASE: PostgreSQL dengan tabel:
users, members, products, orders, order_items, loans,
loan_installments, simpanan_transactions, journal_entries

PROSES BISNIS:
1. M1 Keanggotaan: registrasi→validasi→review→approve→simpanan→aktif
2. M3 POS: tambah produk, transaksi penjualan + split payment, stok, komplain
3. M4 Pinjaman: ajukan→credit scoring→approve→cairkan→angsuran
4. M2 Finance: jurnal otomatis + manual, ringkasan keuangan

STATUS: [STEP TERAKHIR]
FILE YANG ADA: [LIST FILE]
LANJUTKAN: [STEP BERIKUTNYA]
```

---

## 📝 Perbedaan vs Versi Frontend-Only

| Aspek | Frontend Only | Fullstack |
|-------|--------------|-----------|
| Data | Mock/dummy (in-memory) | PostgreSQL real |
| API | Tidak ada | REST API lengkap |
| Persistensi | Hilang saat restart | Tersimpan di DB |
| Kompleksitas | Rendah | Tinggi |
| Waktu | ~15 hari | ~18 hari |
| Cocok untuk | Laporan progress / demo UI | Produksi / MVP real |
