# Proses Bisnis To-Be: Sistem Ekosistem Jual Beli Koperasi Digital

> **Platform:** KopHub — Sistem Manajemen Koperasi & Marketplace Terintegrasi  
> **Versi Dokumen:** 1.0  
> **Tujuan:** Mendeskripsikan transformasi proses bisnis dari sistem konvensional (As-Is) menuju ekosistem digital terintegrasi (To-Be)

---
Gambaran sistem yang berjalan saat ini (As-Is) pada ekosistem jual beli koperasi masih sepenuhnya mengandalkan metode konvensional dan tradisional di mana seluruh alur informasi transaksi belum terfasilitasi oleh sistem digital yang terintegrasi secara real-time . Ketiadaan platform khusus ini mengakibatkan data transaksi operasional masih dicatat menggunakan media buku fisik atau dokumen mandiri yang tersebar di berbagai berkas, sehingga menyulitkan pengelola dalam melakukan pengumpulan data secara terpusat . Dalam praktiknya, proses jual beli antar anggota dilakukan secara langsung melalui pertemuan fisik atau memanfaatkan media pesan singkat pihak ketiga seperti WhatsApp yang hanya berfungsi sebagai sarana komunikasi dasar tanpa adanya dukungan basis data transaksi yang terstruktur.  Seluruh mekanisme pembayaran dalam transaksi tersebut juga masih dilakukan secara manual atau offline melalui uang tunai yang mengharuskan kehadiran fisik anggota di lokasi koperasi. Kondisi ini menciptakan hambatan operasional di mana alur informasi sering kali terputus akibat ketiadaan sistem agregasi data yang otomatis, sehingga anggota koperasi mengalami kesulitan untuk memantau riwayat belanja, histori transaksi, atau sisa saldo simpanan mereka secara mandiri dan transparan. Ketiadaan sistem yang menyatukan data transaksi ini secara real-time pada akhirnya membuat proses pelaporan dan pemantauan kesehatan ekonomi koperasi menjadi tidak optimal dan rentan terhadap kesalahan pencatatan.

diatas adalah penjelasan tentang proses bisnis as-is nya, saya ingin kamu membuatkan proses bisnis to be berdasarkan data diatas yang sesuai dengan alur penggunaan website project ini.
buatkan dalam bentuk file .md
## 1. Ringkasan Eksekutif

Platform KopHub hadir sebagai jawaban atas ketidakefisienan sistem koperasi konvensional yang sepenuhnya bergantung pada pencatatan manual, komunikasi informal via WhatsApp, dan transaksi tunai yang mengharuskan kehadiran fisik. Sistem ini menyatukan dua modul utama dalam satu ekosistem terintegrasi:

| Modul | Fungsi Utama |
|---|---|
| **KopHub Management** | Manajemen anggota, simpanan, pinjaman, produk POS, keuangan & pelaporan |
| **KopHub E-Commerce** | Marketplace digital antar anggota — belanja, berjualan, poin loyalitas |

Kedua modul dihubungkan oleh **Single Sign-On (SSO)**: satu akun, satu login, akses ke seluruh ekosistem.

---

## 2. Perbandingan As-Is vs To-Be

| # | Kondisi As-Is (Konvensional) | Kondisi To-Be (Digital) |
|---|---|---|
| 1 | Pencatatan transaksi di buku fisik & dokumen tersebar | Basis data terpusat, semua transaksi tercatat otomatis real-time |
| 2 | Jual beli melalui pertemuan fisik atau pesan WhatsApp | Marketplace digital 24/7 — anggota bisa belanja & berjualan kapan saja |
| 3 | Pembayaran hanya tunai, mengharuskan kehadiran fisik | Multi-metode pembayaran: Tunai, QRIS, Saldo Anggota, Transfer Bank |
| 4 | Anggota tidak bisa memantau saldo simpanan secara mandiri | Dashboard mandiri — saldo simpanan, riwayat transaksi, angsuran pinjaman terlihat transparan |
| 5 | Pelaporan keuangan manual, rentan kesalahan pencatatan | Jurnal akuntansi & laporan keuangan di-generate otomatis dari setiap transaksi |
| 6 | Alur informasi terputus, data tersebar di berbagai pihak | Integrasi real-time dua modul — data simpanan, pinjaman, & poin sinkron dalam satu sistem |
| 7 | Tidak ada mekanisme loyalitas atau reward bagi anggota aktif | Sistem poin loyalitas: belanja → earn poin → redeem diskon atau konversi ke Simpanan Sukarela |
| 8 | Pengelola sulit memantau kesehatan ekonomi koperasi | Dashboard eksekutif: KPI real-time, alert stok/pinjaman, ringkasan keuangan 6 bulan |

---

## 3. Aktor & Hak Akses Sistem

```
┌──────────────────────────────────────────────────────────────────┐
│                         EKOSISTEM KOPHUB                         │
│                                                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   OWNER     │  │   KASIR     │  │       ANGGOTA           │  │
│  │             │  │             │  │                         │  │
│  │ Akses penuh │  │ Operasional │  │ Layanan mandiri         │  │
│  │ semua modul │  │ harian POS, │  │ + Buyer & Seller        │  │
│  │ + Admin EC  │  │ simpanan,   │  │ di marketplace          │  │
│  │             │  │ pinjaman    │  │                         │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

| Aktor | Manajemen Koperasi | E-Commerce |
|---|---|---|
| **OWNER** | Akses penuh: approve anggota, produk, pinjaman; laporan keuangan | Admin: approve seller & produk, kelola voucher, analytics |
| **KASIR** | POS checkout, input simpanan, proses pinjaman | Buyer: belanja di marketplace |
| **ANGGOTA** | Ajukan pinjaman, pantau simpanan, daftarkan produk | Buyer + Seller (setelah aktivasi): belanja & berjualan |

---

## 4. Alur Proses To-Be — Modul Manajemen Koperasi

### 4.1 Pendaftaran & Verifikasi Anggota

**Masalah As-Is:** Pendaftaran manual di kantor, formulir kertas, proses tidak transparan.

**Solusi To-Be:**

```
[Calon Anggota]
     │
     ▼
Isi Form Digital ──► Nama, NIK (16 digit), Alamat, No. HP
     │
     ▼
Status: PENDING ──► Masuk antrian review pengurus
     │
     ▼
[Pengurus / OWNER / KASIR]
     │
     ├──► APPROVE ──► Nomor Anggota auto-generate (KOP-XXXX)
     │                Status: AKTIF
     │                Akun tersedia untuk seluruh layanan
     │
     └──► REJECT  ──► Status: NON_AKTIF, alasan tercatat
```

**Perubahan kunci:**
- Tidak perlu hadir fisik ke kantor untuk mendaftar
- Status pendaftaran dapat dipantau secara transparan
- Nomor anggota di-generate otomatis dengan format standar

---

### 4.2 Manajemen Simpanan (3 Jenis)

**Masalah As-Is:** Setoran dicatat di buku tabungan fisik, saldo tidak bisa dicek mandiri.

**Solusi To-Be:**

```
Jenis Simpanan yang Dikelola:
  ├── Simpanan Pokok    → Dibayar sekali saat mendaftar
  ├── Simpanan Wajib    → Dibayar rutin (bulanan)
  └── Simpanan Sukarela → Setoran bebas (termasuk dari konversi poin EC)

Alur Transaksi:
[Kasir / Anggota]
     │
     ▼
Input Transaksi ──► Pilih jenis (Pokok/Wajib/Sukarela)
     │               Pilih tipe (MASUK/KELUAR)
     │               Input nominal & keterangan
     ▼
Validasi Sistem ──► Jika KELUAR: cek saldo mencukupi
     │
     ▼
Tercatat di Basis Data
     │
     ▼
Jurnal Akuntansi auto-generate ──► Kas (D) vs Simpanan (K)
     │
     ▼
Anggota dapat pantau saldo real-time via dashboard mandiri
```

---

### 4.3 Pengajuan & Pengelolaan Pinjaman

**Masalah As-Is:** Pengajuan tatap muka, penghitungan angsuran manual, tidak ada tracking progress pelunasan.

**Solusi To-Be:**

```
[Anggota]
     │
     ▼
Form Pengajuan Pinjaman Online
  ├── Nominal (min. Rp 500.000)
  ├── Tenor (1–24 bulan)
  └── Tujuan pinjaman

Sistem menampilkan real-time:
  ├── Credit scoring (total simpanan & pinjaman aktif)
  └── Estimasi angsuran (pokok/bln + bunga 1,5%/bln)
     │
     ▼
Status: PENDING ──► Antrian review pengurus
     │
     ▼
[OWNER / KASIR]
  ├── APPROVE ──► Status: DISETUJUI
  └── REJECT  ──► Status: DITOLAK + alasan

     │ (jika DISETUJUI)
     ▼
Anggota cairkan pinjaman
  └── Jadwal angsuran auto-generate (per bulan)
      Jurnal pencairan: Piutang Anggota (D) vs Kas (K)
     │
     ▼
Pembayaran Angsuran Rutin
  └── Sistem mark DIBAYAR, update sisa pokok
      Progress pelunasan visual (%)
      Jurnal angsuran auto-create
     │
     ▼
Semua angsuran LUNAS ──► Status pinjaman: LUNAS (otomatis)
```

**Perubahan kunci:**
- Transparansi penuh: anggota tahu estimasi angsuran sebelum mengajukan
- Tidak perlu hadir fisik untuk mengajukan atau membayar angsuran
- Riwayat pinjaman tersimpan permanen untuk referensi kredit

---

### 4.4 Point of Sale (POS) — Jual Beli Internal Koperasi

**Masalah As-Is:** Pencatatan penjualan manual di buku kas, stok dihitung manual.

**Solusi To-Be:**

```
[Kasir / Anggota]
     │
     ▼
Browse Katalog Digital ──► Filter by nama/kategori
     │                     Lihat stok real-time & harga
     ▼
Tambah ke Keranjang
  └── Validasi stok otomatis
     │
     ▼
Checkout
  ├── Pilih metode bayar: Tunai / QRIS / Saldo Anggota
  ├── Fee koperasi 3% dihitung otomatis
  └── Validasi ulang stok sebelum konfirmasi
     │
     ▼
Order: DIBAYAR ──► DIKIRIM ──► SELESAI
  └── Jurnal penjualan auto-create di setiap tahap
      Stok berkurang otomatis
     │
     ▼ (jika ada masalah)
Ajukan Komplain
  └── [Admin review]
        ├── Approve Retur ──► Status: BATAL, stok dikembalikan
        └── Tolak Komplain ──► Status: SELESAI dilanjutkan
```

---

### 4.5 Manajemen Produk & Stok

**Masalah As-Is:** Daftar produk di kertas, stok dihitung manual, tidak ada notifikasi stok habis.

**Solusi To-Be:**

```
[Tambah Produk Baru]
  Input: Nama, Kategori, Harga, Stok Awal, Batas Stok Minimum, Deskripsi
     │
     ▼
Auto-assign status berdasarkan role:
  ├── OWNER / KASIR ──► Langsung APPROVED (tayang di katalog)
  └── ANGGOTA ──────► PENDING (antri review pengurus)
     │
     ▼
[Manajemen Stok Berkelanjutan]
  ├── RESTOCK: tambah stok masuk
  └── KOREKSI: koreksi stok (dapat negatif)
  Semua perubahan ter-log dengan timestamp & keterangan

[Alert Otomatis]
  └── Stok < batas minimum ──► Muncul di dashboard pengurus
```

---

### 4.6 Pelaporan & Akuntansi Otomatis

**Masalah As-Is:** Laporan keuangan dibuat manual di akhir periode, rentan salah hitung.

**Solusi To-Be:**

```
Semua transaksi auto-generate jurnal:
  ├── Setoran simpanan  ──► Kas (D) vs Simpanan (K)
  ├── Pencairan pinjaman ──► Piutang Anggota (D) vs Kas (K)
  ├── Angsuran pinjaman  ──► Kas (D) vs Piutang Anggota (K)
  └── Penjualan POS      ──► Kas (D) vs Pendapatan (K)

Jurnal Manual:
  └── Tersedia untuk transaksi operasional lainnya

Dashboard Keuangan Real-time:
  ├── Total Kas (saldo terkini)
  ├── Total Simpanan Anggota (utang koperasi)
  ├── Total Piutang (saldo pinjaman aktif)
  ├── Total Pendapatan Penjualan
  └── Breakdown Bulanan (6 bulan): Pemasukan, Pengeluaran, Selisih
```

---

## 5. Alur Proses To-Be — Modul E-Commerce (KopHub Marketplace)

### 5.1 Autentikasi & Single Sign-On (SSO)

**Masalah As-Is:** Tidak ada platform digital; akses terpisah-pisah.

**Solusi To-Be:**

```
[Satu Akun, Satu Login]
  Email + Password (sama untuk kedua modul)
     │
     ▼
Login di Manajemen Koperasi
  └──► Sesi e-commerce aktif otomatis (SSO)
       Tidak perlu login ulang di marketplace

Role mapping:
  ├── OWNER   ──► Admin E-Commerce (approve semua)
  ├── KASIR   ──► Buyer di marketplace
  └── ANGGOTA ──► Buyer (+ Seller setelah aktivasi)
```

---

### 5.2 Belanja Online — Buyer Journey

**Masalah As-Is:** Cari produk via WhatsApp, tidak ada katalog resmi, bayar tunai.

**Solusi To-Be:**

```
[Fase 1: Penemuan Produk] — Tidak perlu login
  └── Browse marketplace publik
      ├── Cari produk (nama/keyword)
      ├── Filter by kategori (Sembako, Snack, Fashion, dll)
      ├── Lihat detail produk: foto, deskripsi, harga, stok, rating
      └── Baca review dari pembeli lain

[Fase 2: Keputusan Beli]
  Login ──► Tambah ke Wishlist (simpan untuk nanti)
            atau langsung Checkout
     │
     ▼
[Fase 3: Checkout Multi-Step]
  1. Pilih alamat pengiriman (tersimpan / buat baru)
  2. Pilih ekspedisi (JNE, J&T, SiCepat — ongkir otomatis dihitung)
  3. Input kode voucher (diskon otomatis diterapkan)
  4. Gunakan poin sebagai potongan harga
  5. Pilih metode pembayaran
  6. Konfirmasi & order dibuat
     │
     ▼
[Fase 4: Pengiriman & Tracking]
  Status: DIBAYAR ──► DIPROSES ──► DIKIRIM ──► SELESAI
  └── Tracking real-time: DIKEMAS → DIKIRIM → TRANSIT → SAMPAI
      Resi pengiriman tersedia dari seller
     │
     ▼
[Fase 5: Pasca Pembelian]
  Order SELESAI ──► Tulis review (rating 1-5 bintang + komentar)
  └── Poin loyalitas earned otomatis (1% dari total belanja)
      Contoh: Belanja Rp 100.000 → Dapat 100 poin (= Rp 10.000)
```

---

### 5.3 Berjualan Online — Seller Journey

**Masalah As-Is:** Tidak ada platform untuk anggota berjualan selain tatap muka.

**Solusi To-Be:**

```
[Aktivasi Akun Seller]
  Anggota aktif ──► Request aktivasi seller
  [Admin / OWNER approve]
  └── Profil toko dibuat (nama toko, deskripsi)

[Kelola Produk]
  Buat listing produk:
  ├── Nama, deskripsi, kategori, harga, stok, berat
  └── Status: PENDING_APPROVAL ──► Admin review
              ──► APPROVED: tayang di marketplace
              ──► REJECTED: ditolak dengan alasan

[Terima & Proses Pesanan]
  Dashboard seller menampilkan:
  ├── Pesanan baru (DIBAYAR)
  ├── Alert stok menipis
  └── KPI: total produk, revenue, pesanan pending & selesai

  Alur fulfillment:
  Order masuk ──► Proses ──► Input resi pengiriman
  └── Mark DIKIRIM ──► Sistem update tracking buyer

[Laporan Pendapatan]
  Revenue = Total penjualan (order SELESAI)
  Net = Revenue - Komisi koperasi 3% (dipotong otomatis)
  └── Laporan detail per order tersedia
```

---

### 5.4 Sistem Poin & Loyalitas

**Masalah As-Is:** Tidak ada insentif atau reward untuk anggota aktif.

**Solusi To-Be:**

```
[Earn Poin]
  Setiap pembelian di marketplace ──► +1% poin dari total belanja
  Contoh: Belanja Rp 200.000 → Dapat 200 poin

[Redeem Poin — 2 Opsi]
  ├── Opsi 1: Potongan Harga Belanja
  │   └── Gunakan poin saat checkout sebagai diskon
  │       (1 poin = Rp 100 diskon)
  │
  └── Opsi 2: Konversi ke Simpanan Koperasi ★
              Minimum 100 poin
              └── Poin dikonversi → masuk Simpanan Sukarela anggota
                  (terintegrasi langsung ke modul manajemen koperasi)

[Pantau Poin]
  Dashboard poin menampilkan:
  ├── Saldo poin saat ini
  ├── Total earned & total redeemed
  ├── Ekuivalen Rupiah
  └── Histori transaksi poin lengkap
```

---

### 5.5 Integrasi Koperasi ↔ E-Commerce

**Masalah As-Is:** Data simpanan, pinjaman, dan transaksi jual beli tidak terhubung sama sekali.

**Solusi To-Be:**

```
[Link Akun E-Commerce ke Keanggotaan Koperasi]
  Anggota link nomor anggota (KOP-XXXX) ke akun marketplace
     │
     ▼
Manfaat yang didapat:
  ├── Status keanggotaan terlihat di profil marketplace
  ├── Poin e-commerce dapat langsung dikonversi ke Simpanan Sukarela
  └── Satu ekosistem: belanja online, earn poin, poin masuk simpanan

[Alur Konversi Poin → Simpanan]
  Poin marketplace (earned dari belanja)
     │
     ▼
Konversi ke Simpanan Sukarela (min. 100 poin)
  └── Poin berkurang di saldo EC user
      Simpanan Sukarela anggota bertambah (di modul koperasi)
      Transaksi tercatat di kedua sistem
```

---

### 5.6 Pengelolaan Admin Marketplace (OWNER / KASIR)

**Solusi To-Be:**

```
[Dashboard Admin E-Commerce]
  KPI: Pending approvals, total order, revenue, user aktif

[Manajemen Konten]
  ├── Approve / reject produk baru dari seller
  ├── Approve / reject aktivasi akun seller
  └── Kelola voucher promo (buat, aktif/nonaktif, pantau kuota)

[Monitoring & Analytics]
  ├── Revenue by seller
  ├── Top produk terlaris
  ├── Breakdown status order
  ├── Monitoring poin seluruh user
  └── Audit log semua aksi sistem (APPROVE_PRODUCT, REJECT_SELLER, dll)
```

---

## 6. Diagram Alur Terintegrasi

### 6.1 Siklus Hidup Anggota (Member Lifecycle)

```
Registrasi Digital
        │
        ▼
   Status: PENDING
        │
   Pengurus review
        │
   ┌────┴────┐
   ▼         ▼
AKTIF      DITOLAK
   │
   ├─── Akses layanan koperasi (simpanan, pinjaman, POS)
   ├─── Akses marketplace (buyer & seller)
   ├─── Earn poin dari belanja → konversi ke simpanan
   │
   ▼
Pengunduran Diri
   └── Syarat: tidak ada pinjaman aktif
       Status: NON_AKTIF
```

### 6.2 Alur Transaksi Terintegrasi (Simpan → Pinjam → Belanja → Poin)

```
ANGGOTA
   │
   ├──[1] Setor Simpanan ──────────────────────┐
   │        └── Saldo simpanan bertambah        │
   │             Jurnal otomatis                │
   │                                            │
   ├──[2] Ajukan Pinjaman                       │ DATA
   │        └── Credit scoring dari saldo       │ TERPUSAT
   │             simpanan                       │ REAL-TIME
   │             Angsuran otomatis              │
   │                                            │
   ├──[3] Belanja di Marketplace                │
   │        └── Poin earned otomatis            │
   │             Riwayat belanja tercatat        │
   │                                            │
   └──[4] Konversi Poin → Simpanan Sukarela ───┘
            └── Poin EC berkurang
                 Simpanan koperasi bertambah
                 Jurnal otomatis
```

### 6.3 Alur Jual Beli E-Commerce End-to-End

```
SELLER                    SISTEM                    BUYER
   │                         │                        │
   │ Listing produk           │                        │
   │─────────────────────────>│                        │
   │                         │ Admin approve           │
   │                         │────────────────────────>│ Produk tayang
   │                         │                        │
   │                         │         Buyer checkout │
   │                         │<────────────────────────│
   │                         │ Validasi stok/voucher/poin
   │                         │ Hitung ongkir otomatis  │
   │                         │ Order created           │
   │                         │─────────────────────────>│ Konfirmasi order
   │                         │                        │
   │ Notifikasi pesanan       │                        │
   │<─────────────────────────│                        │
   │                         │                        │
   │ Proses & kirim           │                        │
   │─────────────────────────>│                        │
   │                         │ Update tracking         │
   │                         │─────────────────────────>│ Lacak real-time
   │                         │                        │
   │ Order selesai            │                        │
   │                         │ Earn poin buyer         │
   │                         │─────────────────────────>│ +poin otomatis
   │                         │ Pendapatan seller        │
   │<─────────────────────────│ (net -3% komisi)        │
```

---

## 7. Manfaat Digitalisasi — Ringkasan

### Bagi Pengurus Koperasi (OWNER / KASIR)
| Sebelum (As-Is) | Sesudah (To-Be) |
|---|---|
| Rekapitulasi keuangan manual di akhir periode | Dashboard real-time — kas, simpanan, piutang selalu terkini |
| Alert stok habis tidak ada | Notifikasi otomatis saat stok < batas minimum |
| Approval dokumen kertas | Workflow digital: approve/reject satu klik |
| Laporan rawan salah hitung | Jurnal otomatis dari setiap transaksi, error minimal |

### Bagi Anggota Koperasi
| Sebelum (As-Is) | Sesudah (To-Be) |
|---|---|
| Tidak tahu saldo simpanan tanpa ke kantor | Pantau saldo simpanan & angsuran kapan saja |
| Belanja produk koperasi harus hadir fisik | Belanja online 24/7 via marketplace |
| Tidak ada reward untuk anggota aktif | Sistem poin: belanja → earn poin → redeem diskon atau tambah simpanan |
| Pengajuan pinjaman tatap muka | Ajukan pinjaman online, tracking status secara transparan |

### Bagi Ekosistem Koperasi
- **Efisiensi operasional**: proses yang sebelumnya membutuhkan kehadiran fisik kini selesai dalam hitungan menit secara digital
- **Transparansi & akuntabilitas**: seluruh transaksi ter-log dengan timestamp, tidak ada data yang hilang
- **Ekosistem tertutup (closed ecosystem)**: uang berputar di dalam koperasi — belanja → fee koperasi, poin → simpanan sukarela
- **Skalabilitas**: platform siap mendukung pertumbuhan anggota tanpa penambahan beban administrasi proporsional
- **Loyalitas anggota**: sistem poin mendorong anggota untuk lebih aktif bertransaksi di ekosistem koperasi

---

## 8. Glossary Istilah

| Istilah | Definisi |
|---|---|
| **SSO** | Single Sign-On — satu login berlaku di kedua modul (Manajemen & E-Commerce) |
| **POS** | Point of Sale — sistem kasir/penjualan internal koperasi |
| **Simpanan Pokok** | Simpanan yang dibayar sekali saat pertama mendaftar sebagai anggota |
| **Simpanan Wajib** | Simpanan yang wajib dibayar rutin setiap bulan |
| **Simpanan Sukarela** | Simpanan bebas, dapat ditambah kapan saja (termasuk dari konversi poin) |
| **Credit Scoring** | Penilaian kelayakan pinjaman berdasarkan total simpanan & pinjaman aktif |
| **Poin Loyalitas** | Reward digital (1% dari nilai belanja) yang dapat ditukar diskon atau simpanan |
| **Komisi Koperasi** | Potongan 3% dari nilai penjualan di marketplace sebagai pendapatan koperasi |
| **Audit Log** | Catatan semua aksi sistem yang penting untuk transparansi & compliance |
