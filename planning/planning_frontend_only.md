# 📋 MVP Koperasi — Planning Frontend Only (Golang SSR)

## 🎯 Tujuan
Membuat tampilan frontend aplikasi koperasi menggunakan **Golang** dengan pendekatan
**Server-Side Rendering (SSR)** untuk keperluan **laporan progress** pengerjaan proyek.
Data menggunakan **mock/dummy** (belum terkoneksi backend API nyata).

---

## 👥 Pembagian Tim (4 Orang Backend — Fokus Halaman)

| No | Nama | NIM | PIC Halaman |
|----|------|-----|-------------|
| 1 | M Haykal Lazuardy | 607012400068 | Auth + Member (Login, Register, Dashboard Anggota, Simpanan) |
| 2 | Ega Fiandra Pratama | 607012400032 | Product (Katalog, Tambah Produk, Detail, Manajemen Stok) |
| 3 | Raynaldi Paniroean Panjaitan | 607012400099 | Order + Payment (Keranjang, Checkout, Riwayat Order, Komplain) |
| 4 | Diki Alif Taufik | 607012400005 | Loan + Finance (Pengajuan Pinjaman, Angsuran, Jurnal Keuangan) |

> **Saat ini Haykal mengerjakan semua** untuk laporan progress, pembagian di atas untuk dokumentasi.

---

## 🏗️ Tech Stack

| Komponen | Teknologi |
|----------|-----------|
| Bahasa | Golang |
| Web Framework | **Gin** |
| Templating | `html/template` (bawaan Go) |
| Interaktivitas | **HTMX** (opsional, untuk partial reload tanpa JS) |
| CSS Framework | **Bootstrap 5** (via CDN) |
| Data | Mock/dummy data (struct di memory) |
| Icon | Bootstrap Icons (via CDN) |

---

## 📁 Struktur Folder

```
koperasi-frontend/
├── cmd/
│   └── web/
│       └── main.go              # Entry point, setup router
├── internal/
│   ├── handler/
│   │   ├── auth.go              # Login, Register handler
│   │   ├── member.go            # Dashboard anggota, simpanan
│   │   ├── product.go           # Katalog, CRUD produk
│   │   ├── order.go             # Order, checkout, riwayat
│   │   ├── loan.go              # Pinjaman, angsuran
│   │   └── finance.go           # Jurnal keuangan
│   ├── model/
│   │   └── models.go            # Semua struct data
│   └── mock/
│       └── data.go              # Dummy data untuk semua modul
├── templates/
│   ├── layouts/
│   │   ├── base.html            # Layout utama (navbar, sidebar, footer)
│   │   └── auth.html            # Layout halaman login/register
│   ├── auth/
│   │   ├── login.html
│   │   └── register.html
│   ├── member/
│   │   ├── dashboard.html
│   │   ├── list.html
│   │   ├── detail.html
│   │   ├── register_member.html
│   │   └── simpanan.html
│   ├── product/
│   │   ├── catalog.html
│   │   ├── detail.html
│   │   ├── create.html
│   │   └── stock.html
│   ├── order/
│   │   ├── cart.html
│   │   ├── checkout.html
│   │   ├── history.html
│   │   ├── detail.html
│   │   └── complain.html
│   ├── loan/
│   │   ├── apply.html
│   │   ├── list.html
│   │   ├── detail.html
│   │   └── installments.html
│   └── finance/
│       ├── journals.html
│       ├── create_journal.html
│       └── summary.html
├── static/
│   ├── css/
│   │   └── style.css            # Custom CSS tambahan
│   └── js/
│       └── app.js               # JS minimal (jika perlu)
├── go.mod
└── go.sum
```

---

## 🔄 Step-by-Step Pengerjaan

### STEP 1 — Project Setup & Layout (Hari 1-2)

**Task:**
1. `go mod init koperasi-frontend`
2. Install Gin: `go get github.com/gin-gonic/gin`
3. Buat `cmd/web/main.go` — setup Gin router, load templates, serve static
4. Buat `templates/layouts/base.html` — layout utama:
   - Navbar (logo, menu navigasi, user dropdown)
   - Sidebar (menu modul: Dashboard, Anggota, Produk, Order, Pinjaman, Keuangan)
   - Content area
   - Footer
5. Buat `templates/layouts/auth.html` — layout tanpa sidebar (untuk login/register)
6. Buat `static/css/style.css` — custom styling
7. Buat `internal/model/models.go` — semua struct
8. Buat `internal/mock/data.go` — dummy data

**Models yang dibuat:**
```go
type User struct {
    ID       int
    Email    string
    Password string
    Role     string // OWNER, KASIR, ANGGOTA
    Nama     string
}

type Member struct {
    ID               int
    NomorAnggota     string
    Nama             string
    NIK              string
    Alamat           string
    NoHP             string
    Status           string // PENDING, AKTIF, NON_AKTIF
    SimpananPokok    float64
    SimpananWajib    float64
    SimpananSukarela float64
    TanggalMasuk     string
}

type Product struct {
    ID                int
    Nama              string
    Kategori          string
    Harga             float64
    Stok              int
    BatasStokMinimum  int
    Deskripsi         string
    FotoURL           string
    PenjualNama       string
    Status            string // PENDING, APPROVED, REJECTED
}

type Order struct {
    ID          int
    NomorOrder  string
    PembeliNama string
    Items       []OrderItem
    TotalHarga  float64
    FeeKoperasi float64
    Status      string // PENDING, DIBAYAR, DIKIRIM, SELESAI, BATAL, DISPUTED
    MetodeBayar string
    CreatedAt   string
}

type OrderItem struct {
    ProductNama string
    Jumlah      int
    HargaSatuan float64
    Subtotal    float64
}

type Loan struct {
    ID           int
    MemberNama   string
    Nominal      float64
    TenorBulan   int
    BungaPersen  float64
    Tujuan       string
    SisaPokok    float64
    Status       string // PENDING, DISETUJUI, AKTIF, LUNAS, DITOLAK
    TanggalCair  string
    Installments []Installment
}

type Installment struct {
    BulanKe       int
    JatuhTempo    string
    NominalPokok  float64
    NominalBunga  float64
    TotalBayar    float64
    Status        string // BELUM, DIBAYAR, OVERDUE
    TanggalBayar  string
}

type SimpananTransaction struct {
    ID         int
    MemberNama string
    Jenis      string // POKOK, WAJIB, SUKARELA
    Tipe       string // MASUK, KELUAR
    Nominal    float64
    Keterangan string
    CreatedAt  string
}

type JournalEntry struct {
    ID             int
    Tanggal        string
    Keterangan     string
    AkunDebit      string
    AkunKredit     string
    Nominal        float64
    TipeTransaksi  string // SIMPANAN, POS, PINJAMAN, MANUAL
}
```

---

### STEP 2 — Auth Pages (Hari 2-3)

**Halaman:**
1. `/login` — Form login (email + password)
2. `/register` — Form register (email, password, nama)
3. `/logout` — Clear session, redirect ke login

**Fitur:**
- Session sederhana pakai Gin session/cookie
- Setelah login, redirect ke dashboard sesuai role
- Mock: cek email+password dari dummy data

---

### STEP 3 — Member / Anggota Pages (Hari 3-5)

**Halaman sesuai proses bisnis M1:**

1. `/dashboard` — Dashboard utama
   - Ringkasan: jumlah anggota, total simpanan, pinjaman aktif, transaksi hari ini
   - Card statistik dengan angka

2. `/members` — Daftar Anggota (OWNER view)
   - Tabel: No Anggota, Nama, Status, Simpanan Total, Aksi
   - Filter by status (PENDING, AKTIF, NON_AKTIF)
   - Tombol Approve/Reject untuk yang PENDING

3. `/members/:id` — Detail Anggota
   - Info profil lengkap
   - Riwayat simpanan
   - Riwayat pinjaman

4. `/members/register` — Form Pendaftaran Anggota Baru **(M1.1)**
   - Input: Nama, NIK, Alamat, No HP, Upload KTP (dummy)
   - Submit → status PENDING → muncul di daftar review pengurus

5. `/members/:id/simpanan` — Halaman Simpanan **(M1.2)**
   - Saldo: Pokok, Wajib, Sukarela
   - Form bayar simpanan (pilih jenis, nominal)
   - Riwayat transaksi simpanan

6. `/members/:id/resign` — Form Pengunduran Diri **(M1.3)**
   - Tampilkan kewajiban yang belum lunas (jika ada)
   - Form alasan resign

---

### STEP 4 — Product Pages (Hari 5-7)

**Halaman sesuai proses bisnis M3.1 & M3.3:**

1. `/products` — Katalog Produk
   - Grid card produk (foto, nama, harga, stok)
   - Search bar + filter kategori
   - Badge status (Approved, Pending, Rejected)

2. `/products/:id` — Detail Produk
   - Foto besar, deskripsi, harga, stok
   - Info penjual
   - Tombol "Tambah ke Keranjang"

3. `/products/create` — Tambah Produk Baru **(M3.1)**
   - Form: Foto, Nama, Kategori, Harga, Stok, Deskripsi
   - Jika Pengurus → auto approve
   - Jika Anggota → pending review

4. `/products/:id/stock` — Manajemen Stok **(M3.3)**
   - Stok saat ini vs batas minimum
   - Warning jika stok menipis (badge merah)
   - Form restock (tambah stok)
   - Riwayat perubahan stok

5. `/products/review` — Review Produk Pending (OWNER)
   - List produk PENDING dari anggota
   - Tombol Approve / Reject + alasan

---

### STEP 5 — Order & Payment Pages (Hari 7-10)

**Halaman sesuai proses bisnis M3.2 & M3.4:**

1. `/cart` — Keranjang Belanja
   - List item di keranjang (nama, jumlah, harga, subtotal)
   - Tombol +/- jumlah, hapus item
   - Total harga + tombol Checkout

2. `/checkout` — Halaman Checkout **(M3.2)**
   - Ringkasan order
   - Pilih metode bayar: Tunai / QRIS / Saldo Anggota
   - Tombol Bayar

3. `/orders` — Riwayat Order
   - Tabel: No Order, Tanggal, Total, Status, Aksi
   - Filter by status
   - Badge warna per status

4. `/orders/:id` — Detail Order
   - Info order lengkap + list item
   - Status tracking (timeline visual)
   - Tombol aksi sesuai status:
     - DIBAYAR → "Kirim Barang" (penjual)
     - DIKIRIM → "Konfirmasi Terima" (pembeli)
     - SELESAI → tampilkan split payment info

5. `/orders/:id/complain` — Form Komplain **(M3.4)**
   - Pilih alasan (Barang Rusak, Tidak Sesuai, dll)
   - Upload bukti (dummy)
   - Tombol Ajukan Komplain

6. `/orders/complaints` — Daftar Komplain (OWNER)
   - List order DISPUTED
   - Detail bukti + tanggapan penjual
   - Tombol Approve Retur / Reject Komplain

---

### STEP 6 — Loan & Finance Pages (Hari 10-13)

**Halaman sesuai proses bisnis M4.1, M4.2, M2.1, M2.2:**

1. `/loans/apply` — Form Pengajuan Pinjaman **(M4.1)**
   - Input: Nominal, Tenor (bulan), Tujuan
   - Tampilkan estimasi angsuran per bulan
   - Info credit scoring: saldo simpanan, tunggakan aktif
   - Tombol Ajukan

2. `/loans` — Daftar Pinjaman
   - Tabel: Nominal, Tenor, Status, Sisa Pokok
   - Filter by status

3. `/loans/:id` — Detail Pinjaman
   - Info pinjaman lengkap
   - Progress bar sisa pokok
   - Jadwal angsuran (tabel)
   - Tombol Approve/Reject/Cairkan (OWNER)

4. `/loans/:id/pay` — Bayar Angsuran **(M4.2)**
   - Info tagihan bulan ini (pokok + bunga)
   - Denda jika overdue
   - Pilih metode bayar
   - Tombol Bayar

5. `/finance/journals` — Jurnal Keuangan **(M2.1 & M2.2)**
   - Tabel jurnal: Tanggal, Keterangan, Debit, Kredit, Nominal
   - Filter by tipe transaksi
   - Badge tipe: SIMPANAN, POS, PINJAMAN, MANUAL

6. `/finance/journals/create` — Input Jurnal Manual **(M2.2)**
   - Form: Tanggal, Keterangan, Akun Debit, Akun Kredit, Nominal

7. `/finance/summary` — Ringkasan Keuangan
   - Card: Total Kas, Total Simpanan, Total Piutang, Total Pendapatan
   - Tabel ringkasan per kategori

---

### STEP 7 — Polish & Dokumentasi (Hari 13-14)

1. Responsive design (mobile-friendly)
2. Konsistensi tampilan semua halaman
3. Screenshot setiap halaman untuk laporan
4. README.md cara menjalankan project

---

## 🚀 Cara Menjalankan

```bash
cd koperasi-frontend
go mod tidy
go run cmd/web/main.go
# Buka http://localhost:8080
```

---

## 🤖 Prompt untuk AI Lain

### Master Prompt (Context Transfer)

```
Saya sedang membuat MVP Frontend Koperasi menggunakan Golang SSR.

TECH STACK:
- Golang + Gin web framework
- html/template untuk templating
- Bootstrap 5 via CDN untuk styling
- Data dummy/mock (belum ada backend API)

STRUKTUR:
- cmd/web/main.go — entry point
- internal/handler/ — handler per modul (auth, member, product, order, loan, finance)
- internal/model/models.go — struct data
- internal/mock/data.go — dummy data
- templates/layouts/ — base layout (navbar + sidebar)
- templates/<modul>/ — halaman per modul
- static/css/style.css — custom CSS

PROSES BISNIS:
1. M1 Keanggotaan: registrasi (form→validasi→review→approve→aktif), simpanan, resign
2. M3 POS: tambah produk (owner=auto, anggota=pending), transaksi (cart→checkout→bayar→kirim→selesai), stok, komplain/retur
3. M4 Pinjaman: ajukan (credit scoring→review→approve→cairkan→angsuran), bayar angsuran
4. M2 Finance: jurnal otomatis dari transaksi, jurnal manual, ringkasan

STATUS: [STEP TERAKHIR YANG SELESAI]
FILE YANG ADA: [LIST FILE]
LANJUTKAN: [STEP BERIKUTNYA]
```

### Prompt Step 1 — Setup & Layout

```
Buatkan project Golang frontend koperasi dengan Gin + html/template + Bootstrap 5.

1. go mod init koperasi-frontend
2. Buat cmd/web/main.go:
   - Setup Gin router
   - Load semua template dari folder templates/
   - Serve static files dari folder static/
   - Register semua route
3. Buat templates/layouts/base.html:
   - Bootstrap 5 via CDN
   - Navbar: logo "Koperasi Digital", menu (Dashboard, Anggota, Produk, Order, Pinjaman, Keuangan), dropdown user
   - Sidebar: menu navigasi per modul
   - Content block {{template "content" .}}
   - Footer
   - Responsive (mobile friendly)
4. Buat templates/layouts/auth.html — layout tanpa sidebar
5. Buat static/css/style.css — warna tema hijau koperasi
6. Buat internal/model/models.go — struct: User, Member, Product, Order, OrderItem, Loan, Installment, SimpananTransaction, JournalEntry
7. Buat internal/mock/data.go — dummy data minimal 3-5 item per model
```

### Prompt Step 2 — Auth

```
Lanjutkan project koperasi-frontend Golang.
Buatkan halaman auth di internal/handler/auth.go dan templates/auth/:

1. GET /login — tampilkan form login (email + password)
2. POST /login — cek dari mock data, set session cookie, redirect ke /dashboard
3. GET /register — form register (nama, email, password)
4. POST /register — simpan ke mock data, redirect ke /login
5. GET /logout — hapus session, redirect ke /login

Gunakan Gin session (cookie-based).
Layout menggunakan templates/layouts/auth.html.
```

### Prompt Step 3 — Member Pages

```
Lanjutkan project koperasi-frontend Golang.
Buatkan halaman member di internal/handler/member.go dan templates/member/:

Sesuai proses bisnis M1:
1. GET /dashboard — card statistik (jumlah anggota, total simpanan, pinjaman aktif)
2. GET /members — tabel daftar anggota + filter status + tombol approve/reject
3. GET /members/:id — detail profil + riwayat simpanan + pinjaman
4. GET /members/register — form pendaftaran anggota (nama, NIK, alamat, no HP)
5. POST /members/register — simpan dengan status PENDING
6. POST /members/:id/approve — ubah status jadi AKTIF, generate nomor anggota
7. GET /members/:id/simpanan — saldo + form bayar simpanan + riwayat
8. POST /members/:id/simpanan — catat transaksi simpanan, update saldo

Data dari internal/mock/data.go. Layout dari base.html.
Gunakan Bootstrap card, table, badge, form.
```

### Prompt Step 4 — Product Pages

```
Lanjutkan project koperasi-frontend Golang.
Buatkan halaman product di internal/handler/product.go dan templates/product/:

Sesuai M3.1 & M3.3:
1. GET /products — grid card produk (foto placeholder, nama, harga, stok, status badge) + search + filter kategori
2. GET /products/:id — detail produk + tombol tambah ke keranjang
3. GET /products/create — form tambah produk (nama, kategori, harga, stok, deskripsi)
4. POST /products/create — simpan produk (owner=APPROVED, anggota=PENDING)
5. GET /products/review — list produk PENDING + tombol approve/reject
6. POST /products/:id/approve dan /reject
7. GET /products/:id/stock — info stok + warning jika rendah + form restock

Data dari mock. Bootstrap cards grid responsive.
```

### Prompt Step 5 — Order & Payment Pages

```
Lanjutkan project koperasi-frontend Golang.
Buatkan halaman order di internal/handler/order.go dan templates/order/:

Sesuai M3.2 & M3.4:
1. GET /cart — list item keranjang + total + tombol checkout
2. POST /cart/add — tambah item ke keranjang (simpan di session)
3. GET /checkout — ringkasan + pilih metode bayar (Tunai/QRIS/Saldo)
4. POST /checkout — buat order DIBAYAR, kurangi stok mock
5. GET /orders — tabel riwayat order + filter status + badge warna
6. GET /orders/:id — detail + timeline status + tombol aksi (Kirim/Terima/Komplain)
7. POST /orders/:id/ship — status DIKIRIM
8. POST /orders/:id/complete — status SELESAI
9. GET /orders/:id/complain — form komplain (alasan)
10. POST /orders/:id/complain — status DISPUTED
11. GET /orders/complaints — list DISPUTED + tombol resolve

Data dari mock. Keranjang disimpan di session.
```

### Prompt Step 6 — Loan & Finance Pages

```
Lanjutkan project koperasi-frontend Golang.
Buatkan halaman loan & finance di internal/handler/loan.go, finance.go dan templates/:

Sesuai M4 & M2:
1. GET /loans/apply — form pengajuan (nominal, tenor, tujuan) + estimasi angsuran + info credit scoring
2. POST /loans/apply — simpan loan PENDING
3. GET /loans — tabel pinjaman + filter status
4. GET /loans/:id — detail + progress bar sisa pokok + jadwal angsuran + tombol approve/cairkan
5. POST /loans/:id/approve — status DISETUJUI
6. POST /loans/:id/disburse — status AKTIF + generate jadwal angsuran
7. POST /loans/:id/pay — bayar angsuran, update sisa pokok
8. GET /finance/journals — tabel jurnal keuangan + filter tipe
9. GET /finance/journals/create — form jurnal manual
10. POST /finance/journals/create — simpan jurnal
11. GET /finance/summary — card ringkasan (kas, simpanan, piutang, pendapatan)

Data dari mock. Bootstrap tables + cards + progress bar.
```

---

## ⏱️ Timeline

| Step | Durasi | Target |
|------|--------|--------|
| Step 1 — Setup & Layout | 1-2 hari | Hari ke-2 |
| Step 2 — Auth | 1 hari | Hari ke-3 |
| Step 3 — Member | 2-3 hari | Hari ke-6 |
| Step 4 — Product | 2-3 hari | Hari ke-9 |
| Step 5 — Order | 3 hari | Hari ke-12 |
| Step 6 — Loan & Finance | 2-3 hari | Hari ke-14 |
| Step 7 — Polish | 1 hari | Hari ke-15 |

**Total estimasi: ~15 hari kerja**
