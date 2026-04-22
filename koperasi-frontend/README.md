# 🏦 Koperasi Digital — MVP Frontend

Aplikasi frontend koperasi menggunakan **Golang SSR** (Server-Side Rendering) dengan framework **Gin** dan templating **html/template**. Data menggunakan mock/dummy di memory.

## 📋 Tentang Proyek

Proyek ini merupakan MVP (Minimum Viable Product) tampilan frontend aplikasi koperasi digital yang mencakup 4 proses bisnis utama:

1. **Keanggotaan (M1)** — Registrasi, review, simpanan, pengunduran diri
2. **Keuangan (M2)** — Jurnal keuangan otomatis & manual, ringkasan keuangan
3. **POS & Produk (M3)** — Katalog, CRUD produk, keranjang, checkout, order, komplain
4. **Pinjaman (M4)** — Pengajuan, persetujuan, pencairan, angsuran

## 👥 Tim Pengembang

| No | Nama | NIM | PIC Halaman |
|----|------|-----|-------------|
| 1 | M Haykal Lazuardy | 607012400068 | Auth + Member |
| 2 | Ega Fiandra Pratama | 607012400032 | Product |
| 3 | Raynaldi Paniroean Panjaitan | 607012400099 | Order + Payment |
| 4 | Diki Alif Taufik | 607012400005 | Loan + Finance |

## 🏗️ Tech Stack

| Komponen | Teknologi |
|----------|-----------|
| Bahasa | Golang 1.21+ |
| Web Framework | [Gin](https://github.com/gin-gonic/gin) |
| Templating | `html/template` (bawaan Go) |
| CSS Framework | [Bootstrap 5.3](https://getbootstrap.com/) via CDN |
| Icon | [Bootstrap Icons 1.11](https://icons.getbootstrap.com/) via CDN |
| Session | [gin-contrib/sessions](https://github.com/gin-contrib/sessions) (cookie-based) |
| Data | Mock/dummy data (struct di memory) |

## 📁 Struktur Folder

```
koperasi-frontend/
├── cmd/
│   └── web/
│       └── main.go              # Entry point, router, template loader
├── internal/
│   ├── handler/
│   │   ├── auth.go              # Login, Register, Logout, AuthRequired middleware
│   │   ├── cart.go              # Session-based cart helpers
│   │   ├── dashboard.go         # Dashboard utama
│   │   ├── flash.go             # Flash message helpers
│   │   ├── member.go            # Anggota CRUD + simpanan
│   │   ├── product.go           # Produk CRUD + stok
│   │   ├── order.go             # Order, checkout, komplain
│   │   ├── loan.go              # Pinjaman, angsuran
│   │   └── finance.go           # Jurnal keuangan, ringkasan
│   ├── model/
│   │   └── models.go            # Semua struct data
│   └── mock/
│       ├── data.go              # Dummy data untuk semua modul
│       └── helpers.go           # Helper functions (CRUD, ID generator, dll)
├── templates/
│   ├── layouts/
│   │   ├── base.html            # Layout utama (navbar, sidebar, footer)
│   │   └── auth.html            # Layout halaman login/register
│   ├── auth/                    # login.html, register.html
│   ├── dashboard/               # index.html
│   ├── member/                  # list, detail, register, simpanan, resign
│   ├── product/                 # catalog, detail, create, stock, review
│   ├── order/                   # cart, checkout, history, detail, complain, complaints
│   ├── loan/                    # apply, list, detail
│   └── finance/                 # journals, create_journal, summary
├── static/
│   ├── css/
│   │   └── style.css            # Custom CSS
│   └── js/
│       └── app.js               # JavaScript minimal
├── go.mod
├── go.sum
└── README.md
```

## 🚀 Cara Menjalankan

### Prasyarat
- [Go](https://golang.org/dl/) versi 1.21 atau lebih baru

### Langkah

```bash
# 1. Clone / masuk ke direktori project
cd koperasi-frontend

# 2. Install dependencies
go mod tidy

# 3. Jalankan server
go run cmd/web/main.go

# 4. Buka browser
# http://localhost:8080
```

### Akun Demo

| Email | Password | Role |
|-------|----------|------|
| `owner@koperasi.id` | `owner123` | OWNER (full access) |
| `kasir@koperasi.id` | `kasir123` | KASIR |
| `anggota@koperasi.id` | `anggota123` | ANGGOTA |
| `rina@koperasi.id` | `rina123` | ANGGOTA |

## 📄 Daftar Halaman

### Auth
| Route | Metode | Deskripsi |
|-------|--------|-----------|
| `/login` | GET/POST | Halaman login |
| `/register` | GET/POST | Halaman registrasi |
| `/logout` | GET | Logout & redirect |

### Dashboard
| Route | Metode | Deskripsi |
|-------|--------|-----------|
| `/dashboard` | GET | Dashboard utama dengan statistik |

### Anggota (Member)
| Route | Metode | Deskripsi |
|-------|--------|-----------|
| `/members` | GET | Daftar anggota + filter status |
| `/members/register` | GET/POST | Form pendaftaran anggota |
| `/members/:id` | GET | Detail profil anggota |
| `/members/:id/approve` | POST | Approve anggota PENDING |
| `/members/:id/reject` | POST | Reject anggota PENDING |
| `/members/:id/simpanan` | GET/POST | Halaman simpanan + form bayar |
| `/members/:id/resign` | GET/POST | Form pengunduran diri |

### Produk (Product)
| Route | Metode | Deskripsi |
|-------|--------|-----------|
| `/products` | GET | Katalog produk + search/filter |
| `/products/create` | GET/POST | Tambah produk baru |
| `/products/review` | GET | Review produk PENDING |
| `/products/:id` | GET | Detail produk |
| `/products/:id/approve` | POST | Approve produk |
| `/products/:id/reject` | POST | Reject produk |
| `/products/:id/stock` | GET/POST | Manajemen stok |

### Order & Cart
| Route | Metode | Deskripsi |
|-------|--------|-----------|
| `/cart` | GET | Keranjang belanja |
| `/cart/add` | POST | Tambah item ke keranjang |
| `/cart/update` | POST | Update jumlah item |
| `/cart/remove` | POST | Hapus item dari keranjang |
| `/checkout` | GET/POST | Checkout & pembayaran |
| `/orders` | GET | Riwayat order + filter status |
| `/orders/:id` | GET | Detail order + timeline |
| `/orders/:id/ship` | POST | Tandai order DIKIRIM |
| `/orders/:id/complete` | POST | Konfirmasi terima (SELESAI) |
| `/orders/:id/complain` | GET/POST | Form komplain |
| `/orders/:id/resolve` | POST | Resolve komplain (OWNER) |
| `/orders/complaints` | GET | Daftar komplain DISPUTED |

### Pinjaman (Loan)
| Route | Metode | Deskripsi |
|-------|--------|-----------|
| `/loans` | GET | Daftar pinjaman + filter status |
| `/loans/apply` | GET/POST | Ajukan pinjaman baru |
| `/loans/:id` | GET | Detail pinjaman + jadwal angsuran |
| `/loans/:id/approve` | POST | Setujui pinjaman |
| `/loans/:id/reject` | POST | Tolak pinjaman |
| `/loans/:id/disburse` | POST | Cairkan dana pinjaman |
| `/loans/:id/pay` | POST | Bayar angsuran |

### Keuangan (Finance)
| Route | Metode | Deskripsi |
|-------|--------|-----------|
| `/finance/journals` | GET | Jurnal keuangan + filter tipe |
| `/finance/journals/create` | GET/POST | Input jurnal manual |
| `/finance/summary` | GET | Ringkasan keuangan + arus kas |

## 🔄 Alur Proses Bisnis

### M1 — Keanggotaan
```
Registrasi → PENDING → Review Pengurus → AKTIF
                                       → DITOLAK
Aktif → Simpanan (Pokok/Wajib/Sukarela)
Aktif → Resign (cek kewajiban) → NON_AKTIF
```

### M3 — POS (Penjualan)
```
Tambah Produk → (Anggota: PENDING → Approve)
             → (Owner: auto APPROVED)
Browse Katalog → Tambah ke Keranjang → Checkout
→ Pilih Metode Bayar → DIBAYAR → DIKIRIM → SELESAI
                                         → KOMPLAIN → DISPUTED → Resolve
```

### M4 — Pinjaman
```
Ajukan → PENDING → Approve → DISETUJUI → Cairkan → AKTIF
                 → Reject → DITOLAK
AKTIF → Bayar Angsuran → ... → LUNAS
```

### M2 — Keuangan
```
Jurnal otomatis dari: Simpanan, POS, Pinjaman
Jurnal manual: Input oleh pengurus
Ringkasan: Total Kas, Simpanan, Piutang, Pendapatan
```

## ⚙️ Catatan Teknis

- **Data bersifat sementara** — semua data tersimpan di memory dan hilang saat server restart
- **Session** menggunakan cookie-based session dari `gin-contrib/sessions`
- **Keranjang belanja** disimpan di session per user
- **Jurnal keuangan** dicatat otomatis saat terjadi transaksi (simpanan, penjualan, pencairan pinjaman, angsuran)
- **Template** menggunakan pattern `layout + content` — setiap halaman define `{{template "content" .}}`

## 📝 Lisensi

Proyek ini dibuat untuk keperluan akademik — Mata Kuliah Proyek Sistem Informasi.
