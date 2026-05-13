# Analisis Modul E-Commerce — KopHub

**Proyek:** koperasi-frontend  
**Modul:** KopHub E-Commerce (Marketplace)  
**Tanggal Analisis:** 2026-05-13

---

## 📋 Daftar Isi

1. [Ringkasan Eksekutif](#ringkasan-eksekutif)
2. [Deskripsi Pengguna Aplikasi](#deskripsi-pengguna-aplikasi)
3. [Hak Akses per Role](#hak-akses-per-role)
4. [Alur Teknis](#alur-teknis)
5. [File-File Kritis](#file-file-kritis)

---

## 🎯 Ringkasan Eksekutif

KopHub E-Commerce adalah modul marketplace terintegrasi yang berjalan di atas aplikasi frontend koperasi. Modul ini memungkinkan:

- **Anggota koperasi** untuk membuka toko dan menjual produk
- **Pembeli** untuk berbelanja dengan sistem poin reward yang terintegrasi ke simpanan koperasi
- **Admin** untuk mengawasi ekosistem marketplace (approval seller, approval produk, manajemen voucher, audit)

Sistem menggunakan **session-based authentication** (cookie) tanpa JWT. Kontrol akses menggunakan kombinasi middleware berbasis role dan flag `is_seller`.

---

## 👥 Deskripsi Pengguna Aplikasi

### 1. Guest (Pengunjung Belum Login)

**Deskripsi:** Pengguna yang tidak terautentikasi atau baru pertama kali mengunjungi toko.

**Akses:**
- Browsing storefront dan katalog produk
- Melihat detail produk
- Filter produk per kategori
- Melihat profil publik seller
- Halaman login & signup

---

### 2. Buyer (Pembeli)

**Deskripsi:** Pengguna yang telah mendaftar dan login ke e-commerce. Ini adalah role **default** saat pertama kali mendaftar.

**Karakteristik:**
- `Role = "BUYER"`
- `IsSellerActive = false` (awalnya)
- Bisa upgrade ke Seller dengan mengaktifkan akun penjual
- Dapat memiliki `LinkedKoperasiMemberID` untuk integrasi dengan anggota koperasi

**Akses:**
- Semua akses Guest
- Dashboard pribadi
- Wishlist (lihat & kelola)
- Belanja & checkout
- Kelola pesanan (lihat, tracking, ulasan)
- Manajemen profil & alamat pengiriman
- Kelola poin loyalty & riwayat
- Konversi poin ke simpanan koperasi (jika sudah link akun koperasi)

---

### 3. Seller (Penjual)

**Deskripsi:** Seorang Buyer yang telah mengaktifkan akun penjualnya. **Bukan role terpisah**, melainkan flag `IsSellerActive = true` di atas akun Buyer.

**Karakteristik:**
- `Role = "BUYER"` (tetap sama)
- `IsSellerActive = true` (diaktifkan)
- Memiliki profil toko sendiri (nama toko, rating, response time, total terjual)
- Produk yang dibuat berstatus `PENDING_APPROVAL` sampai di-approve Admin
- Dapat merangkap berbelanja sebagai Buyer

**Akses:**
- Semua akses Buyer
- Dashboard Seller dengan KPI
- Kelola produk (buat, lihat, edit)
- Kelola pesanan masuk (mark as shipped, tracking)
- Laporan earnings/pendapatan
- **Tidak bisa** akses panel Admin (meski owner adalah Admin, flag `is_seller` harus false)

---

### 4. Admin (Administrator)

**Deskripsi:** Administrator marketplace dengan hak penuh untuk mengelola ekosistem penjualan.

**Karakteristik:**
- `Role = "ADMIN"` (role khusus)
- Tidak memiliki flag `is_seller` (tidak bisa akses seller dashboard)
- Berhak approve/reject seller dan produk
- Dapat membuat voucher dan monitor poin

**Akses:**
- Semua akses Buyer (tapi tidak ada panel Seller)
- Dashboard Admin dengan KPI (pending produk, total order, revenue, seller aktif, voucher, audit)
- Analytics marketplace
- Persetujuan seller baru (approve/reject)
- Persetujuan produk dari seller (approve/reject)
- Manajemen voucher (buat, toggle status)
- Manajemen semua order
- Manajemen member e-commerce
- Pemantauan poin seluruh user & riwayat
- Audit log semua aksi penting

---

### 5. Kasir (Role Belum Diimplementasikan)

**Deskripsi:** Role `KASIR` sudah didefinisikan di model data dan ada dalam seed data, namun belum ada middleware atau route khusus untuk mengatur akses.

**Status:** Placeholder untuk pengembangan fase berikutnya.

---

## 🔐 Hak Akses per Role

### Guest (Tidak Login)

| # | Fitur | Akses |
|---|---|---|
| 1 | Lihat storefront `/ecommerce/store` | ✅ |
| 2 | Lihat katalog produk `/ecommerce/products` | ✅ |
| 3 | Lihat detail produk `/ecommerce/products/:id` | ✅ |
| 4 | Filter produk per kategori | ✅ |
| 5 | Lihat profil publik seller | ✅ |
| 6 | Halaman login & signup | ✅ |
| 7 | API pencarian anggota koperasi | ✅ |
| 8 | Semua fitur lainnya | ❌ |

**Middleware:** Tidak ada (public)

---

### Buyer (Login, `ec_is_seller = false`)

Semua akses Guest **plus**:

| # | Fitur | Route | Akses |
|---|---|---|---|
| 1 | Dashboard Buyer | `/ecommerce/buyer` | ✅ |
| 2 | Wishlist (lihat & toggle) | `/ecommerce/wishlist`, `/ecommerce/wishlist/toggle` | ✅ |
| 3 | Checkout & buat pesanan | `/ecommerce/checkout` | ✅ |
| 4 | Lihat daftar pesanan | `/ecommerce/orders` | ✅ |
| 5 | Tracking pesanan | `/ecommerce/orders/:id/track` | ✅ |
| 6 | Submit ulasan produk | `/ecommerce/order/:id/review` | ✅ |
| 7 | Profil & pengaturan akun | `/ecommerce/profile`, `/ecommerce/profile/settings` | ✅ |
| 8 | Kelola alamat pengiriman | `/ecommerce/profile/addresses`, `/ecommerce/profile/addresses/*` | ✅ |
| 9 | Lihat saldo & riwayat poin | `/ecommerce/points` | ✅ |
| 10 | Konversi poin ke simpanan koperasi | `/ecommerce/points/convert` | ✅ (harus link akun koperasi) |
| 11 | Hubungkan akun ke anggota koperasi | `/ecommerce/points/link-member` | ✅ |
| 12 | API Koperasi (link, simpanan) | `/ecommerce/api/koperasi/*` | ✅ |
| 13 | Panel Seller | `/ecommerce/seller/*` | ❌ |
| 14 | Panel Admin | `/ecommerce/admin/*` | ❌ |

**Middleware:** `RequireECommerceAuth()` (check session `ec_user_id`)

---

### Seller (Login, `ec_is_seller = true`)

Semua akses Buyer **plus**:

| # | Fitur | Route | Akses |
|---|---|---|---|
| 1 | Dashboard Seller | `/ecommerce/seller` | ✅ |
| 2 | Lihat daftar produk milik sendiri | `/ecommerce/seller/products` | ✅ |
| 3 | Buat produk baru | `/ecommerce/seller/products/create` | ✅ |
| 4 | Lihat daftar pesanan masuk | `/ecommerce/seller/orders` | ✅ |
| 5 | Tandai pesanan dikirim | `/ecommerce/seller/orders/:id/shipped` | ✅ |
| 6 | Lihat laporan earnings | `/ecommerce/seller/earnings` | ✅ |
| 7 | Panel Admin | `/ecommerce/admin/*` | ❌ |

**Middleware:** `RequireECommerceAuth()` + `RequireECommerceSeller()` (check `ec_is_seller == true`)

**Catatan Penting:**
- Produk baru yang dibuat akan berstatus `PENDING_APPROVAL`
- Produk tidak tampil di katalog sampai di-approve Admin
- Seller hanya melihat pesanan yang dikirim ke tokonya sendiri
- Seller tetap bisa berbelanja sebagai Buyer (flag `is_seller` terpisah dari role)

---

### Admin (`ec_role = "ADMIN"`)

Semua akses Buyer **plus** (namun **tidak ada akses Seller**):

| # | Fitur | Route | Akses |
|---|---|---|---|
| 1 | Dashboard Admin | `/ecommerce/admin` | ✅ |
| 2 | Analytics marketplace | `/ecommerce/admin/analytics` | ✅ |
| 3 | Approve/reject seller baru | `/ecommerce/admin/sellers` | ✅ |
| 4 | Approve/reject produk dari seller | `/ecommerce/admin/products` | ✅ |
| 5 | Manajemen voucher | `/ecommerce/admin/vouchers` | ✅ |
| 6 | Manajemen semua order | `/ecommerce/admin/orders` | ✅ |
| 7 | Manajemen member e-commerce | `/ecommerce/admin/members` | ✅ |
| 8 | Pemantauan poin seluruh user | `/ecommerce/admin/points` | ✅ |
| 9 | Audit log | `/ecommerce/admin/audit` | ✅ |
| 10 | Panel Seller | `/ecommerce/seller/*` | ❌ |

**Middleware:** `RequireECommerceAuth()` + `RequireECommerceAdmin()` (check `ec_role == "ADMIN"`)

**Catatan Penting:**
- Admin yang login akan di-redirect ke `/ecommerce/buyer` (bukan langsung ke admin panel)
- Admin tidak bisa akses seller dashboard (role ADMIN tidak punya flag `is_seller`)
- Akses ke route `/ecommerce/admin/*` akan return **HTTP 403** jika bukan ADMIN
- Setiap aksi admin penting di-log ke Audit Log

---

## 🔄 Alur Teknis

### Authentication Flow

```
1. Guest akses /ecommerce/login
   ↓
2. Input username + password
   ↓
3. Validasi di AuthenticateECommerceUser()
   ↓
4. Set session keys: ec_user_id, ec_username, ec_email, ec_role, ec_is_seller
   ↓
5. Redirect based on role:
   - BUYER & seller active (is_seller=true) → /ecommerce/seller
   - BUYER & seller inactive (is_seller=false) → /ecommerce/buyer
   - ADMIN → /ecommerce/buyer (redirect di middleware)
```

### Middleware Stack

Setiap request ke protected route melalui stack middleware berikut:

```
Request
  ↓
1. RequireECommerceAuth()
   - Check session key: ec_user_id
   - Jika tidak ada → redirect /ecommerce/login (302)
   ↓
2. Optional: RequireECommerceSeller()  [hanya untuk seller routes]
   - Check session key: ec_is_seller == true
   - Jika false → redirect /ecommerce/buyer (302)
   ↓
3. Optional: RequireECommerceAdmin()  [hanya untuk admin routes]
   - Check session key: ec_role == "ADMIN"
   - Jika bukan ADMIN → HTTP 403 (inline, no redirect)
   ↓
Handler
```

### Product Approval Workflow

```
Seller buat produk
  ↓
Product.Status = "PENDING_APPROVAL"
  ↓
Tidak tampil di katalog publik
  (queryAllApprovedECProducts filter Status == "APPROVED")
  ↓
Admin akses /ecommerce/admin/products
  ↓
Admin approve: Product.Status = "APPROVED"
  ↓
Produk muncul di katalog & bisa dibeli
```

### Points Flow

```
Buyer membuat pesanan
  ↓
CreateECOrder() calculate:
  - Subtotal dari items
  - Shipping cost
  - Discount (voucher)
  - Points used (jika ditukar)
  ↓
Total = subtotal + shipping - discount - pointsUsed
  ↓
Auto-earn: 1 poin per Rp 1,000 subtotal
  ↓
Points recorded in UserPoints.Balance
  ↓
Buyer bisa:
  a) Pakai untuk diskon di order berikutnya
  b) Konversi ke simpanan koperasi (1 poin = Rp 100)
```

---

## 📁 File-File Kritis

### Backend (Core Logic)

| Komponen | Path | Deskripsi |
|---|---|---|
| **Middleware** | `koperasi-frontend/core/middleware/ecommerce.go` | `RequireECommerceAuth()`, `RequireECommerceSeller()`, `RequireECommerceAdmin()` |
| **Routes** | `koperasi-frontend/core/server/server.go` | Pendaftaran semua route e-commerce |
| **Data Model** | `koperasi-frontend/core/model/models.go` | Struct: `ECommerceUser`, `ECProduct`, `ECOrder`, `Voucher`, `UserPoints`, `AuditLog`, dll. |
| **Mock Data** | `koperasi-frontend/core/mock/ecommerce.go` | Seed 5 user, 10 produk, 3 order, 2 voucher, shipping options |
| **Mock Helpers** | `koperasi-frontend/core/mock/ecommerce_helpers.go` | Helper: auth, create user, link koperasi, approve product, order, points, audit log |

### Handlers (HTTP Endpoints)

| Handler | Path | Routes | Fungsi Utama |
|---|---|---|---|
| **Auth** | `core/handler/ecommerce/auth.go` | Login, Signup, Logout, Link Koperasi | User registration & authentication |
| **Buyer** | `core/handler/ecommerce/buyer.go` | Dashboard | Buyer dashboard view |
| **Seller** | `core/handler/ecommerce/seller.go` | Dashboard, Products, Orders, Earnings | Seller dashboard & management |
| **Store** | `core/handler/ecommerce/store.go` | Home, Category, Products, Detail | Public catalog browsing |
| **Order** | `core/handler/ecommerce/order.go` | Checkout, List, Tracking | Order management |
| **Profile** | `core/handler/ecommerce/profile.go` | Profile, Addresses, Settings | User profile & address |
| **Wishlist** | `core/handler/ecommerce/wishlist.go` | List, Toggle | Wishlist management |
| **Review** | `core/handler/ecommerce/review.go` | Show, Do review | Product reviews post-order |
| **Points** | `core/handler/ecommerce/points.go` | Balance, Convert, Link Member | Points & loyalty system |
| **Integration** | `core/handler/ecommerce/integration.go` | Koperasi member API | Cross-system integration (JSON API) |
| **Admin** | `core/handler/ecommerce/admin.go` | All admin operations | Admin dashboard & management |

### Frontend (Templates)

| Template Group | Path | Jumlah File | Fungsi |
|---|---|---|---|
| **Layouts** | `api/templates/ecommerce/layouts/` | 2 | `ec_base.html` (authenticated), `ec_auth.html` (login/signup) |
| **Auth** | `api/templates/ecommerce/auth/` | 2 | Login & signup pages |
| **Store** | `api/templates/ecommerce/store/` | 2 | Storefront & category listing |
| **Product** | `api/templates/ecommerce/product/` | 2 | Catalog & detail page |
| **Seller** | `api/templates/ecommerce/seller/` | 6 | Dashboard, products, orders, earnings, profile, form |
| **Buyer** | `api/templates/ecommerce/buyer/` | 1 | Buyer dashboard |
| **Order** | `api/templates/ecommerce/order/` | 4 | Checkout, list, tracking, review |
| **Wishlist** | `api/templates/ecommerce/wishlist/` | 1 | Wishlist list |
| **Profile** | `api/templates/ecommerce/profile/` | 4 | Profile, addresses, address form, settings |
| **Points** | `api/templates/ecommerce/points/` | 3 | Balance, convert form, link member form |
| **Admin** | `api/templates/ecommerce/admin/` | 8 | Dashboard, analytics, sellers, products, vouchers, orders, members, points, audit |

---

## 📊 Data Model Snapshot

### Entitas Utama

**ECommerceUser**
```
- ID (int)
- Username, Email, Password (string)
- Role: BUYER | SELLER | ADMIN | KASIR (string)
- IsSellerActive (bool)
- SellerRating, LinkedKoperasiMemberID (int)
- CreatedAt (time)
```

**ECProduct**
```
- ID, SellerID (int)
- Nama, Kategori, Harga, Stok, Berat (gram)
- Rating, TotalReview, TotalSold (int)
- Status: PENDING_APPROVAL | APPROVED | REJECTED | ARCHIVED
- CreatedAt (time)
```

**ECOrder**
```
- ID, BuyerID, SellerID (int)
- Items (slice of ECOrderItem)
- AlamatPengiriman, ShippingOption, ShippingCost (string/int)
- Subtotal, Discount, PointsUsed, TotalHarga (int)
- VoucherCode, MetodeBayar (string)
- Status: PENDING | DIBAYAR | DIPROSES | DIKIRIM | SELESAI | BATAL
- ResiPengiriman, PointsEarned (string/int)
- CreatedAt, UpdatedAt (time)
```

**Voucher**
```
- ID (int)
- TipeDiskon: PERCENT | FIXED (string)
- NilaiDiskon, MinPembelian, MaksDiskon, Kuota (int)
- Status: ACTIVE | EXPIRED | USED_UP (string)
- BerlakuSampai (time)
```

**UserPoints**
```
- UserID (int)
- Balance, TotalEarned, TotalRedeemed (int)
```

**AuditLog**
```
- ID (int)
- Action, UserID, Username (string)
- Resource, Details (string)
- CreatedAt (time)
```

---

## 🧪 Test Users (Seed Data)

| Username | Role | IsSellerActive | Toko | Linked Koperasi | Status |
|---|---|---|---|---|---|
| `andi` | BUYER | ✅ | Toko Andi Jaya (234 sold) | Yes (ID 1) | ✅ Approved |
| `rina` | BUYER | ✅ | Rina Craft & Food (156 sold) | Yes (ID 2) | ✅ Approved |
| `budi` | BUYER | ❌ | — | No | ✅ Buyer only |
| `admin` | ADMIN | ❌ | — | No | ✅ Admin |
| `kasir` | KASIR | ❌ | — | No | ⚠️ No routes yet |

---

## 🔗 Integrasi Koperasi

Modul e-commerce terintegrasi dengan sistem koperasi melalui:

1. **Member Linking** — Buyer bisa link akun e-commerce ke anggota koperasi
   - Route: POST `/ecommerce/link-member`, `/ecommerce/unlink-member`
   - Field: `LinkedKoperasiMemberID` di `ECommerceUser`

2. **Points Conversion** — Poin loyalty e-commerce bisa dikonversi ke `SimpananSukarela`
   - Rate: 1 poin = Rp 100
   - Persyaratan: Akun harus sudah di-link ke member koperasi
   - Route: POST `/ecommerce/points/convert`

3. **Member Search** — Seller/Admin bisa cari anggota koperasi saat link
   - Route: GET `/ecommerce/api/members/search?q=...`

---

## 🎯 Kesimpulan

KopHub E-Commerce adalah modul marketplace yang matang dengan **4 role utama** dan kontrol akses berbasis session middleware. Sistem dirancang untuk:

- ✅ Mendukung multi-seller marketplace dengan approval workflow
- ✅ Integrasi points loyalty ke simpanan koperasi
- ✅ Admin oversight penuh dengan audit logging
- ✅ Separation of concern (Buyer, Seller, Admin paths terpisah jelas)
- ✅ Scalable structure dengan handler per domain & mock backend API

Implementasi menggunakan **Gin framework** dengan **session-based auth** dan data in-memory (mock), siap untuk integrasi database real.

---

**Dokumen ini dibuat berdasarkan analisis kode sumber tanggal 2026-05-13.**
