# BAB 3
# PEMODELAN DAN PERANCANGAN
## (Modul Admin & Dashboard Administrator)

---

### 3.1 Arsitektur Sistem

Untuk memberikan ilustrasi interaksi antar komponen pada sistem yang dikembangkan, perlu dijelaskan arsitektur sistem yang menjadi fondasi **Modul Admin & Dashboard SmartMart**. Arsitektur ini dirancang untuk memungkinkan integrasi dan *interoperabilitas* di antara berbagai lapisan teknologi yang berbeda, mulai dari peramban web Administrator, *server* aplikasi berbasis Go/Gin, *engine* templat berbasis `html/template`, hingga lapisan data *in-memory* yang menyimpan seluruh transaksi dan *audit log* koperasi.

Aplikasi dibangun mengikuti pendekatan **monolith server-rendered**: seluruh logika bisnis, *routing*, dan komposisi tampilan dijalankan di satu *binary* Go yang sama. Pendekatan ini dipilih karena ringan, mudah di-*deploy* ke *platform serverless* (Vercel), dan tetap kompatibel dengan eksekusi *standalone* pada lingkungan koperasi yang memiliki keterbatasan infrastruktur. Asset (templat HTML, CSS, JavaScript, dan ikon) ditanamkan langsung ke dalam *binary* melalui mekanisme `//go:embed` Go, sehingga satu *binary* tunggal cukup untuk menjalankan seluruh modul Admin tanpa perlu *deploy* berkas terpisah.

Secara teknis, arsitektur ini terdiri atas tiga lapisan utama:

1. **Client Layer (Administrator)** — Peramban modern (Chrome, Firefox, Edge) yang diakses oleh pengurus koperasi melalui koneksi HTTPS dengan *cookie session* sebagai pembawa identitas.
2. **Server Layer (Vercel Serverless / Standalone `:8080`)** — *Router* Gin yang memuat *middleware* berlapis (`RequireECommerceAuth()` lalu `RequireECommerceAdmin()`), *handler* admin di [admin.go](koperasi-frontend/core/handler/ecommerce/admin.go), *engine* templat `html/template`, dan *session store* berbasis Gorilla sessions (`gin-contrib/sessions`).
3. **Embedded Asset Layer (`//go:embed`)** — Direktori [api/templates/](koperasi-frontend/api/templates/) dan [api/static/](koperasi-frontend/api/static/) yang di-*embed* secara *build-time* ke dalam *binary* sehingga *deployment* tetap satu *file*.

Sebagaimana dijelaskan pada Bab 2, seluruh rute admin (`/ecommerce/admin/...`) hanya bisa diakses oleh pengguna dengan *role* `ADMIN`. Hal ini ditegakkan melalui *middleware* `RequireECommerceAdmin()` di [server.go:448](koperasi-frontend/core/server/server.go#L448) yang memeriksa nilai `ec_role` pada *session* dan menolak permintaan apabila *role* bukan `ADMIN`. Untuk memberikan pengalaman *Single Sign-On*, *middleware* `RequireECommerceAuth()` juga melakukan fallback otomatis dari *session* Koperasi induk ke *session* e-commerce apabila pengguna sudah masuk di sistem koperasi.

[Gambar 3-1: Diagram Arsitektur Sistem SmartMart — Modul Admin & Dashboard]
Sumber diagram: [xml/arsitektur_sistem.drawio.xml](xml/arsitektur_sistem.drawio.xml)

---

### 3.2 Pemodelan Sistem dan Data

Pemodelan sistem dalam Proyek Sistem Informasi ini bertujuan untuk mendeskripsikan struktur, komponen, dan interaksi berbagai elemen di dalam Modul Admin SmartMart agar seluruh bagian berfungsi secara harmonis. Notasi yang digunakan mengikuti standar **Unified Modeling Language (UML)** untuk pemodelan berorientasi objek, mencakup *use case diagram*, *class diagram*, dan *sequence diagram*. Untuk pemodelan alur proses bisnis, digunakan notasi **Business Process Model and Notation (BPMN 2.0)** yang dapat dieksekusi dan divisualisasikan langsung di *tool* Camunda Modeler. Untuk pemodelan struktur data, digunakan **Entity Relationship Diagram (ERD)** yang menggambarkan entitas, atribut, dan relasi antar entitas pada lapisan basis data.

#### 3.2.1 Use Case Diagram (Administrator)

*Use case diagram* berikut menggambarkan seluruh fungsi yang dapat diakses oleh aktor **Administrator (Pengurus Koperasi)** pada Modul Admin SmartMart. Setiap *use case* dipetakan langsung dengan rute HTTP yang ada pada [server.go:447-464](koperasi-frontend/core/server/server.go#L447).

**Aktor:**
- **Administrator (Pengurus Koperasi)** — Pengguna dengan *role* `ADMIN` yang memiliki akses eksklusif ke seluruh fungsi manajerial dashboard.

**Daftar Use Case:**

1. **Login Sistem** — Administrator melakukan autentikasi melalui halaman *login* e-commerce; sistem memeriksa kredensial dan menetapkan *session* dengan `ec_role = ADMIN`. *Use case* ini meng-`<<include>>` *Validasi Role ADMIN (RBAC)* untuk memastikan otorisasi.
2. **Lihat Dashboard KPI** — Administrator mengakses halaman utama `/ecommerce/admin` untuk melihat ringkasan KPI (jumlah produk *pending*, total order, total *user*, total *revenue*, jumlah *seller* aktif, *recent activity*).
3. **Lihat Analytics & Laporan Kinerja** — Administrator mengakses halaman `/ecommerce/admin/analytics` untuk melihat metrik *real-time*: *total revenue*, *order status breakdown*, *seller performance*, dan *top products*.
4. **Approve / Reject Seller** — Administrator menyetujui atau menolak aktivasi akun *seller* melalui `POST /ecommerce/admin/sellers/:id/approve` atau `/reject`. Use case ini meng-`<<include>>` *Catat Audit Log*.
5. **Approve / Reject Produk** — Administrator menyetujui atau menolak produk yang berstatus `PENDING_APPROVAL` melalui `POST /ecommerce/admin/products/:id/approve` atau `/reject`. Use case ini meng-`<<include>>` *Catat Audit Log*.
6. **Buat Voucher Diskon** — Administrator membuat *voucher* baru melalui `POST /ecommerce/admin/vouchers/create` dengan parameter kode, tipe diskon (PERCENT/FIXED), nilai diskon, minimum pembelian, kuota, dan tanggal berlaku. Use case ini meng-`<<include>>` *Catat Audit Log*.
7. **Toggle Status Voucher** — Administrator mengubah status *voucher* antara `ACTIVE` dan `EXPIRED` melalui `POST /ecommerce/admin/vouchers/:id/toggle`. Use case ini meng-`<<include>>` *Catat Audit Log*.
8. **Pantau Daftar Order** — Administrator melihat seluruh transaksi di `/ecommerce/admin/orders` dengan filter status (`DIBAYAR`, `DIPROSES`, `DIKIRIM`, `SELESAI`, `BATAL`).
9. **Pantau Daftar Member** — Administrator melihat seluruh pengguna e-commerce di `/ecommerce/admin/members` dengan filter *role* (`ADMIN`, `BUYER`, `SELLER`).
10. **Pantau Poin Loyalitas** — Administrator memantau total poin beredar, transaksi poin, dan saldo poin per pengguna di `/ecommerce/admin/points`.
11. **Lihat Audit Log** — Administrator menelusuri riwayat seluruh tindakan administratif di `/ecommerce/admin/audit` dengan filter berdasarkan jenis aksi.

[Gambar 3-2: Use Case Diagram Administrator SmartMart]
Sumber diagram: [xml/use_case_admin.drawio.xml](xml/use_case_admin.drawio.xml)

#### 3.2.2 Class Diagram

*Class diagram* berikut memvisualisasikan struktur data utama yang menjadi fondasi Modul Admin. Seluruh kelas yang ditampilkan diambil langsung dari definisi *struct* Go pada berkas [core/model/models.go](koperasi-frontend/core/model/models.go) baris 120–291.

**Kelas yang divisualisasikan:**

- **`ECommerceUser`** — Representasi akun pengguna e-commerce dengan atribut `ID`, `Username`, `Email`, `Password`, `Role`, `IsSellerActive`, `SellerRating`, `LinkedKoperasiMemberID`, dan `CreatedAt`. *Role* ini menentukan otorisasi RBAC (`BUYER`, `SELLER`, `ADMIN`, `KASIR`).
- **`SellerProfile`** — Profil tambahan untuk pengguna dengan `IsSellerActive = true`. Memiliki relasi *one-to-one* dengan `ECommerceUser`.
- **`ECProduct`** — Representasi produk pada *marketplace* dengan atribut `Status` yang menentukan apakah produk masih `PENDING_APPROVAL`, sudah `APPROVED`, `REJECTED`, atau `ARCHIVED`. Atribut ini menjadi dasar moderasi pada halaman *Review Produk*.
- **`ECOrder`** — Representasi transaksi pembelian dengan rangkaian status (`PENDING`, `DIBAYAR`, `DIPROSES`, `DIKIRIM`, `SELESAI`, `BATAL`), `Subtotal`, `Discount`, `TotalHarga`, `VoucherCode`, dan `PointsEarned`. Mengagregasi `ECOrderItem` melalui *composition*.
- **`ECOrderItem`** — Item individual dalam *order* dengan referensi ke `ProductID`. Termasuk *part* dari `ECOrder`.
- **`Voucher`** — *Voucher* diskon dengan atribut `Code`, `TipeDiskon` (PERCENT/FIXED), `NilaiDiskon`, `MinPembelian`, `MaksDiskon`, `Kuota`, `Status` (ACTIVE/EXPIRED/USED_UP), dan `BerlakuSampai`.
- **`UserPoints`** — Saldo poin loyalitas per pengguna dengan `Balance`, `TotalEarned`, dan `TotalRedeemed`. Memiliki relasi *one-to-one* dengan `ECommerceUser`.
- **`PointsTransaction`** — Riwayat akumulasi atau penukaran poin dengan `Tipe` (`EARN_PURCHASE`, `REDEEM_VOUCHER`, `REDEEM_DISCOUNT`, `CONVERT_SIMPANAN`) dan `Amount` (positif untuk *earn*, negatif untuk *redeem*).
- **`AuditLog`** — Catatan aktivitas administratif dengan `Action`, `UserID`, `Username`, `Resource` (mis. `"product:5"`), `Details` (narasi), dan `CreatedAt` (*timestamp*).

**Relasi antar kelas:**

| Dari | Ke | Multiplisitas | Jenis |
|---|---|---|---|
| ECommerceUser | SellerProfile | 1..1 | Composition |
| ECommerceUser | ECProduct | 1..N | Association (sebagai seller) |
| ECommerceUser | ECOrder | 1..N | Association (sebagai buyer/seller) |
| ECOrder | ECOrderItem | 1..N | Composition |
| ECProduct | ECOrderItem | 1..N | Association |
| ECommerceUser | UserPoints | 1..1 | Association |
| ECommerceUser | PointsTransaction | 1..N | Association |
| ECommerceUser | AuditLog | 1..N | Association |

[Gambar 3-3: Class Diagram Modul Admin SmartMart]
Sumber diagram: [xml/class_diagram_admin.drawio.xml](xml/class_diagram_admin.drawio.xml)

#### 3.2.3 Sequence Diagram

*Sequence diagram* berikut menggambarkan dua alur paling representatif pada Modul Admin: persetujuan produk (sebagai contoh tindakan moderasi dengan *audit trail*) dan pembuatan *voucher* (sebagai contoh tindakan *create* dengan validasi).

**A. Sequence Diagram — Approve Product**

Diagram ini memvisualisasikan urutan pemanggilan antar objek ketika Administrator menekan tombol *"Approve"* pada produk dengan status `PENDING_APPROVAL`. Aliran melewati 6 objek utama: *Web Browser*, *Gin Router*, *Middleware* `RequireECommerceAdmin`, *AdminHandler.ApproveProduct*, *mock.SetProductStatus*, dan *mock.LogAuditAction*. Setiap pemanggilan disertai pesan eksplisit, dimulai dari klik tombol oleh Administrator hingga *redirect* HTTP 302 kembali ke halaman daftar produk dengan status yang sudah diperbarui.

Pemanggilan kritis berada pada langkah ke-7 (`SetProductStatus(id, "APPROVED")`) yang memperbarui status produk pada lapisan data, dan langkah ke-9 (`LogAuditAction(APPROVE_PRODUCT, ...)`) yang menjamin setiap keputusan moderasi tercatat secara permanen di *Audit Log*.

[Gambar 3-4: Sequence Diagram Approve Product]
Sumber diagram: [xml/sequence_approve_product.drawio.xml](xml/sequence_approve_product.drawio.xml)

**B. Sequence Diagram — Create Voucher**

Diagram ini memvisualisasikan urutan pemanggilan ketika Administrator mengirim *form* pembuatan *voucher* baru. Aliran melewati 5 objek: *Web Browser*, *Gin Router*, *AdminHandler.CreateVoucher*, *mock.AddVoucher*, dan *mock.LogAuditAction*. Validasi *form* (langkah 4–6) dilakukan di sisi *handler* mencakup pemeriksaan keunikan `Code`, validitas `TipeDiskon`, nilai diskon harus positif, dan penetapan tanggal berlaku default (+3 bulan dari saat pembuatan).

[Gambar 3-5: Sequence Diagram Create Voucher]
Sumber diagram: [xml/sequence_create_voucher.drawio.xml](xml/sequence_create_voucher.drawio.xml)

#### 3.2.4 Activity Diagram (BPMN 2.0)

Untuk memodelkan alur proses bisnis pada Modul Admin, digunakan notasi **BPMN 2.0** yang merupakan standar industri dan dapat dieksekusi langsung pada *engine* alur kerja seperti Camunda. Dua proses kunci dimodelkan dalam bentuk BPMN:

**A. BPMN — Alur Persetujuan Produk**

Diagram BPMN ini menggambarkan kolaborasi antara tiga *lane* (Penjual, Sistem SmartMart, Administrator) dalam siklus hidup persetujuan produk:

1. **Penjual** menekan tombol *upload* produk pada `/ecommerce/seller/products/create`.
2. **Sistem** otomatis menetapkan `Status = PENDING_APPROVAL` dan menyimpan produk ke lapisan data.
3. **Administrator** membuka halaman `/ecommerce/admin/products` untuk melihat daftar produk yang menunggu persetujuan.
4. **Administrator** meninjau detail produk (nama, deskripsi, harga, foto, stok) dan mengambil keputusan via *Exclusive Gateway*: **Layak disetujui?**
5. Jika **Ya**, sistem menjalankan `POST /products/:id/approve` → `Status = APPROVED` → `LogAuditAction(APPROVE_PRODUCT)` → produk tampil di katalog `/ecommerce/products`.
6. Jika **Tidak**, sistem menjalankan `POST /products/:id/reject` → `Status = REJECTED` (beserta alasan penolakan) → `LogAuditAction(REJECT_PRODUCT)` → notifikasi ke penjual.

[Gambar 3-6: Activity Diagram BPMN Persetujuan Produk]
Sumber diagram: [xml/bpmn_persetujuan_produk.bpmn](xml/bpmn_persetujuan_produk.bpmn)

**B. BPMN — Alur Pembuatan Voucher**

Diagram BPMN ini menggambarkan kolaborasi antara dua *lane* (Administrator dan Sistem SmartMart) dalam alur pembuatan *voucher* baru:

1. **Administrator** membuka halaman `/ecommerce/admin/vouchers` dan mengisi *form* dengan parameter `Code`, `TipeDiskon`, `NilaiDiskon`, `Kuota`, dan `MinPembelian`.
2. **Sistem** menjalankan validasi: keunikan `Code`, nilai diskon harus > 0, dan `TipeDiskon` harus `PERCENT` atau `FIXED`.
3. *Exclusive Gateway* **Valid?**: bila *valid*, sistem memanggil `AddVoucher(...)`, menetapkan `Status = ACTIVE` dan `BerlakuSampai = now + 3 bulan`, lalu mencatat `LogAuditAction(CREATE_VOUCHER, voucher:<code>)`. Bila tidak *valid*, sistem me-*render* ulang *form* dengan *error message*.

[Gambar 3-7: Activity Diagram BPMN Pembuatan Voucher]
Sumber diagram: [xml/bpmn_pembuatan_voucher.bpmn](xml/bpmn_pembuatan_voucher.bpmn)

#### 3.2.5 Entity Relationship Diagram (ERD)

Pemodelan data pada Modul Admin SmartMart difokuskan pada struktur dan hubungan antar entitas yang menjadi sumber data untuk seluruh halaman administratif. Meskipun implementasi MVP saat ini menggunakan *mock data* berbasis *in-memory array* (`core/mock/`), struktur entitas telah dirancang sesuai dengan kaidah normalisasi sehingga dapat dipetakan langsung ke skema basis data relasional (PostgreSQL/MySQL) pada fase produksi.

**Daftar Entitas:**

| Entitas | Deskripsi |
|---|---|
| `ecommerce_user` | Akun pengguna e-commerce dengan *role* RBAC |
| `seller_profile` | Profil tambahan untuk pengguna *seller* |
| `ec_product` | Produk pada *marketplace* dengan status moderasi |
| `ec_order` | Transaksi pembelian dengan siklus status order |
| `ec_order_item` | Item individual dalam satu *order* |
| `voucher` | Kode *voucher* diskon yang dibuat administrator |
| `user_points` | Saldo poin loyalitas per pengguna |
| `points_transaction` | Riwayat akumulasi & penukaran poin |
| `audit_log` | Catatan tindakan administratif untuk akuntabilitas |

**Relasi antar Entitas:**

| Entitas Pertama | Entitas Kedua | Kardinalitas | Catatan |
|---|---|---|---|
| `ecommerce_user` | `seller_profile` | 1..1 | Hanya pengguna dengan `IsSellerActive = true` |
| `ecommerce_user` | `ec_product` | 1..N | `seller_id` FK ke `ecommerce_user.id` |
| `ecommerce_user` | `ec_order` | 1..N | Sebagai *buyer* dan *seller* (dua relasi) |
| `ec_order` | `ec_order_item` | 1..N | *Composition* — item dihapus bila *order* dihapus |
| `ec_product` | `ec_order_item` | 1..N | `product_id` FK ke `ec_product.id` |
| `ecommerce_user` | `user_points` | 1..1 | Tiap user punya satu saldo poin |
| `ecommerce_user` | `points_transaction` | 1..N | Riwayat poin per user |
| `ec_order` | `points_transaction` | 0..N | *Order* boleh menghasilkan banyak transaksi poin (earn + redeem) |
| `ecommerce_user` | `audit_log` | 1..N | Mencatat semua aksi administrator |
| `voucher` | `ec_order` | 0..N | *Voucher* opsional pada *order* |

[Gambar 3-8: Entity Relationship Diagram Modul Admin SmartMart]
Sumber diagram: [xml/erd_admin.drawio.xml](xml/erd_admin.drawio.xml)

---

### 3.3 Perancangan Antarmuka Pengguna

Perancangan antarmuka pengguna pada Modul Admin SmartMart difokuskan pada satu *role* utama, yaitu **Administrator (Pengurus Koperasi)**. Seluruh antarmuka dirancang dengan prinsip kesederhanaan visual, *single-screen overview*, dan aksesibilitas tinggi mengingat target pengguna memiliki literasi teknologi tingkat menengah (lihat *User Persona* Pak Hendra pada Bab 2). Total terdapat **sembilan halaman utama** yang dapat diakses oleh Administrator setelah berhasil *login* dan lolos pemeriksaan *middleware* `RequireECommerceAdmin()`.

#### 3.3.1 Halaman Dashboard

- **Rute:** `GET /ecommerce/admin`
- **Handler:** `AdminHandler.Dashboard`
- **Template:** `ecommerce/admin/dashboard.html`
- **Tujuan:** Memberikan ringkasan satu layar (single-screen overview) atas kondisi koperasi digital secara *real-time*.
- **Komponen UI utama:**
  - **Kartu KPI:** Jumlah Produk *Pending*, Total Order, Total Pengguna, Total *Revenue*, Jumlah *Seller* Aktif, Jumlah Order Selesai, Jumlah Voucher Aktif.
  - **Aksi Cepat (Quick Actions):** Tautan langsung ke halaman moderasi yang paling sering diakses (Approve Produk, Approve Seller, Buat Voucher).
  - **Aktivitas Terbaru (Recent Activity):** Lima entri terakhir dari *Audit Log* dengan *timestamp* dan ringkasan aksi.

[Gambar 3-9: Tampilan Halaman Dashboard Administrator]

#### 3.3.2 Halaman Analytics

- **Rute:** `GET /ecommerce/admin/analytics`
- **Handler:** `AdminHandler.Analytics`
- **Template:** `ecommerce/admin/analytics.html`
- **Tujuan:** Menampilkan laporan kinerja koperasi yang dihitung otomatis dari seluruh data transaksi.
- **Komponen UI utama:**
  - **Ringkasan Metrik:** Kartu untuk *Total Revenue*, Total Order, Order Selesai, dan Order dalam Proses.
  - **Breakdown Status Order:** Tabel distribusi *order* berdasarkan status (`DIBAYAR`, `DIPROSES`, `DIKIRIM`, `SELESAI`, `BATAL`).
  - **Tabel Performa Seller:** Daftar seluruh *seller* aktif dengan *revenue*, jumlah *order*, dan *rating*.
  - **Top 5 Produk Terlaris:** Daftar produk dengan `TotalSold` tertinggi.

[Gambar 3-10: Tampilan Halaman Analytics Administrator]

#### 3.3.3 Halaman Manajemen Seller

- **Rute:** `GET /ecommerce/admin/sellers`
- **Handler:** `AdminHandler.SellerApprovals`
- **Template:** `ecommerce/admin/sellers.html`
- **Tujuan:** Mengelola aktivasi akun *seller* bagi anggota koperasi yang ingin mendaftarkan UMKM-nya.
- **Komponen UI utama:**
  - **Tab Non-Seller:** Daftar pengguna yang belum mengaktifkan akun *seller*, lengkap dengan tombol **Approve** untuk aktivasi.
  - **Tab Seller Aktif:** Daftar pengguna yang sudah berstatus *seller* aktif, lengkap dengan informasi `StoreName`, `Rating`, dan `TotalSold`.
  - **Tombol Aksi:** *Approve* (`POST .../approve`) dan *Reject* (`POST .../reject`); keduanya otomatis memanggil `LogAuditAction()`.

[Gambar 3-11: Tampilan Halaman Manajemen Seller]

#### 3.3.4 Halaman Review Produk

- **Rute:** `GET /ecommerce/admin/products`
- **Handler:** `AdminHandler.ProductApprovals`
- **Template:** `ecommerce/admin/products.html`
- **Tujuan:** Melakukan moderasi produk multi-seller sebelum tampil di katalog *marketplace*.
- **Komponen UI utama:**
  - **Tab Produk Pending:** Daftar produk dengan status `PENDING_APPROVAL`, lengkap dengan tombol **Approve** dan **Reject**.
  - **Tab Semua Produk:** Daftar seluruh produk dengan filter status (APPROVED, REJECTED, ARCHIVED).
  - **Modal Detail Produk:** Menampilkan foto, deskripsi, kategori, harga, dan stok untuk membantu pengambilan keputusan.
  - **Form Alasan Penolakan:** Wajib diisi saat menekan tombol *Reject* — tersimpan ke `AuditLog.Details`.

[Gambar 3-12: Tampilan Halaman Review Produk]

#### 3.3.5 Halaman Manajemen Voucher

- **Rute:** `GET /ecommerce/admin/vouchers`
- **Handler:** `AdminHandler.VoucherManagement`
- **Template:** `ecommerce/admin/vouchers.html`
- **Tujuan:** Membuat, mengelola, dan menonaktifkan kode *voucher* diskon untuk insentif anggota.
- **Komponen UI utama:**
  - **Tabel Daftar Voucher:** Menampilkan kode, tipe diskon, nilai, kuota tersisa, status, dan tanggal berlaku.
  - **Form Create Voucher:** Input untuk `Code` (unik, otomatis di-*uppercase*), `TipeDiskon` (PERCENT/FIXED), `NilaiDiskon`, `MinPembelian`, `MaksDiskon`, `Kuota`, dan `BerlakuSampai` (default +3 bulan dari saat pembuatan).
  - **Tombol Toggle Status:** Mengubah status voucher antara `ACTIVE` dan `EXPIRED` (`POST .../toggle`).

[Gambar 3-13: Tampilan Halaman Manajemen Voucher]

#### 3.3.6 Halaman Manajemen Order

- **Rute:** `GET /ecommerce/admin/orders`
- **Handler:** `AdminHandler.OrderManagement`
- **Template:** `ecommerce/admin/orders.html`
- **Tujuan:** Memantau seluruh transaksi *order* yang terjadi di platform untuk eskalasi dan investigasi komplain.
- **Komponen UI utama:**
  - **Filter Status:** *Dropdown* untuk memilih status (`DIBAYAR`, `DIPROSES`, `DIKIRIM`, `SELESAI`, `BATAL`) — terkirim sebagai *query parameter* `?status=`.
  - **Ringkasan Jumlah Order per Status:** Kartu jumlah untuk tiap status.
  - **Tabel Daftar Order:** Menampilkan `NomorOrder`, `BuyerName`, `SellerName`, `TotalHarga`, `Status`, `VoucherCode`, dan `CreatedAt`.

[Gambar 3-14: Tampilan Halaman Manajemen Order]

#### 3.3.7 Halaman Manajemen Member

- **Rute:** `GET /ecommerce/admin/members`
- **Handler:** `AdminHandler.MemberManagement`
- **Template:** `ecommerce/admin/members.html`
- **Tujuan:** Memantau seluruh anggota e-commerce koperasi dengan kemampuan filter berdasarkan *role*.
- **Komponen UI utama:**
  - **Filter Role:** *Dropdown* untuk memilih *role* (`ADMIN`, `BUYER`, `SELLER`) — terkirim sebagai `?role=`.
  - **Total Jumlah Pengguna:** Kartu jumlah keseluruhan pengguna terdaftar.
  - **Tabel Daftar Member:** Menampilkan `Username`, `Email`, `Role`, `IsSellerActive`, `SellerRating`, `LinkedKoperasiMemberID`, dan `CreatedAt`.

[Gambar 3-15: Tampilan Halaman Manajemen Member]

#### 3.3.8 Halaman Monitoring Poin Loyalitas

- **Rute:** `GET /ecommerce/admin/points`
- **Handler:** `AdminHandler.PointsMonitoring`
- **Template:** `ecommerce/admin/points.html`
- **Tujuan:** Mengawasi seluruh akumulasi dan penukaran *Point Rewards* untuk menjamin transparansi sistem loyalitas digital.
- **Komponen UI utama:**
  - **Kartu Total Poin Beredar:** Total `Balance` dari seluruh entri `UserPoints`.
  - **Tabel Transaksi Poin:** Riwayat lengkap `PointsTransaction` dengan filter berdasarkan `Tipe` (`EARN_PURCHASE`, `REDEEM_VOUCHER`, `REDEEM_DISCOUNT`, `CONVERT_SIMPANAN`).
  - **Tabel Saldo per Pengguna:** Daftar `UserPoints` dengan `Balance`, `TotalEarned`, dan `TotalRedeemed`.

[Gambar 3-16: Tampilan Halaman Monitoring Poin Loyalitas]

#### 3.3.9 Halaman Audit Log

- **Rute:** `GET /ecommerce/admin/audit`
- **Handler:** `AdminHandler.AuditLog`
- **Template:** `ecommerce/admin/audit.html`
- **Tujuan:** Menyediakan jejak audit (*audit trail*) lengkap atas seluruh tindakan administratif untuk kepatuhan, transparansi, dan investigasi.
- **Komponen UI utama:**
  - **Filter Aksi:** *Dropdown* untuk memilih jenis aksi (`APPROVE_SELLER`, `REJECT_SELLER`, `APPROVE_PRODUCT`, `REJECT_PRODUCT`, `CREATE_VOUCHER`, `ACTIVATE_VOUCHER`, `DEACTIVATE_VOUCHER`, `LINK_MEMBER`, dll.) — terkirim sebagai `?action=`.
  - **Tabel Audit Log:** Setiap baris menampilkan `Action`, `Resource`, `Details`, `Username` & `UserID`, dan `CreatedAt`. Urutan kronologis terbaru di atas.

[Gambar 3-17: Tampilan Halaman Audit Log]

---

### 3.4 Kebutuhan Perangkat Keras dan Perangkat Lunak

#### 3.4.1 Pengembangan Sistem

Spesifikasi perangkat keras dan perangkat lunak yang digunakan oleh tim pengembang untuk membangun Modul Admin SmartMart adalah sebagai berikut:

**Kebutuhan Perangkat Keras (Hardware) — Pengembangan:**

| Komponen | Spesifikasi Minimum |
|---|---|
| Prosesor (CPU) | Intel Core i5 generasi ke-8 atau setara |
| Memori (RAM) | 8 GB |
| Penyimpanan (Storage) | SSD 256 GB |
| Resolusi Monitor | 1366 × 768 (Full HD direkomendasikan) |
| Koneksi Internet | Minimal 10 Mbps (untuk *dependency download* dan kolaborasi) |

**Kebutuhan Perangkat Lunak (Software) — Pengembangan:**

| Kategori | Tools yang Digunakan |
|---|---|
| Sistem Operasi | Windows 11 (proyek aktual), kompatibel dengan macOS dan Linux |
| Bahasa Pemrograman | Go SDK versi 1.25.0+ — wajib sesuai dengan deklarasi pada `go.mod` |
| Web Framework | Gin v1.12.0 (`github.com/gin-gonic/gin`) |
| Manajemen Session | gin-contrib/sessions v1.1.0 + Gorilla sessions v1.4.0 |
| IDE / Code Editor | Visual Studio Code (dengan ekstensi Go), JetBrains GoLand |
| Version Control | Git versi 2.40+ |
| Web Browser | Google Chrome (terbaru), Mozilla Firefox, Microsoft Edge |
| Diagram Modeling | Camunda Modeler (untuk berkas `.bpmn`), drawio / app.diagrams.net (untuk berkas `.drawio.xml`) |
| Deployment CLI | Vercel CLI (opsional, untuk *preview deployment*) |
| Build Tool | Toolchain bawaan Go (`go build`, `go run`, `go mod tidy`) |

#### 3.4.2 Implementasi Sistem

Spesifikasi perangkat keras dan perangkat lunak minimum yang dibutuhkan untuk menjalankan Modul Admin SmartMart pada lingkungan produksi adalah sebagai berikut:

**Kebutuhan Perangkat Keras (Hardware) — Sisi Server:**

| Komponen | Spesifikasi Minimum |
|---|---|
| Prosesor (CPU) | 1 vCPU (Vercel Serverless) atau 2 vCPU untuk *standalone* |
| Memori (RAM) | 256 MB per *instance* (sesuai batasan kinerja pada Bab 2.5) |
| Penyimpanan (Storage) | 100 MB (binary Go + asset embedded) |
| Bandwidth | Minimal 50 GB/bulan untuk operasional normal koperasi |

**Kebutuhan Perangkat Lunak (Software) — Sisi Server:**

| Kategori | Spesifikasi |
|---|---|
| Sistem Operasi | Linux (Vercel default) atau Windows Server 2019+ |
| Runtime | Tidak diperlukan — *binary* Go bersifat *statically linked* |
| Web Server / Reverse Proxy | Vercel Edge Network atau Nginx (jika *self-host*) |
| Protokol | HTTPS (TLS 1.2 minimum) |
| Database | Tidak diperlukan untuk MVP saat ini (data *in-memory*). Pada fase produksi disarankan PostgreSQL 14+ atau MySQL 8+. |

**Kebutuhan Perangkat Keras (Hardware) — Sisi Client (Administrator):**

| Komponen | Spesifikasi Minimum |
|---|---|
| Perangkat | Komputer/laptop personal |
| Resolusi Monitor | Minimal 1366 × 768 |
| Koneksi Internet | 4G LTE / Wi-Fi minimal 5 Mbps |

**Kebutuhan Perangkat Lunak (Software) — Sisi Client (Administrator):**

| Kategori | Spesifikasi |
|---|---|
| Web Browser | Google Chrome 110+, Mozilla Firefox 110+, atau Microsoft Edge 110+ |
| Cookie & JavaScript | Wajib diaktifkan (untuk *session* dan interaktivitas dashboard) |

---

### 3.5 Subbab Tambahan

Tidak diperlukan subbab tambahan untuk Modul Admin & Dashboard SmartMart pada Bab 3. Seluruh aspek pemodelan sistem (arsitektur, *use case*, *class diagram*, *sequence diagram*, BPMN, ERD), perancangan antarmuka, serta kebutuhan perangkat telah dijabarkan secara komprehensif pada subbab-subbab di atas.
