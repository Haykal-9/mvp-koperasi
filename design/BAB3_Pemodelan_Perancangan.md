# BAB 3 — PEMODELAN DAN PERANCANGAN

> Teks siap salin ke Word. Sisipkan gambar dari folder `design/*.png` pada setiap penanda **[SISIPKAN Gambar 3.x]**.
> Urutan gambar: 3.1 Arsitektur · 3.2 Use Case · 3.3 ERD · 3.4 Class · 3.5–3.8 Sequence · 3.9–3.18 Rancangan Antarmuka.

---

## 3.1 Arsitektur Sistem

Untuk memberikan ilustrasi interaksi antar komponen pada sistem yang dikembangkan, perlu dijelaskan arsitektur sistem yang menjadi kerangka kerja utama platform SmartMart. Sistem SmartMart dirancang menggunakan pendekatan arsitektur terdistribusi berbasis *Client-Server* yang memisahkan beban kerja ke dalam tiga lapisan (*layer*) utama, yaitu Lapisan Klien (*Frontend*), Lapisan Cloud Server (*Backend*), dan Lapisan Data. Pemisahan ini bertujuan untuk menjamin keamanan data, meringankan beban komputasi pada perangkat pengguna, serta memungkinkan satu sumber data terpusat (*single source of truth*) diakses oleh banyak jenis aktor melalui antarmuka yang berbeda.

**[SISIPKAN Gambar 3.1 — Arsitektur Sistem SmartMart] (design/arsitektur_smartmart.png)**

Berdasarkan Gambar 3.1, alur komunikasi sistem dijabarkan sebagai berikut:

1. **Lapisan Klien (*Frontend Layer*):** Merupakan antarmuka visual yang digunakan oleh manusia. Lapisan ini terdiri atas aplikasi *mobile* lintas platform berbasis **Flutter** — meliputi *User App* (Anggota), *Seller App* (Penjual), dan *Delivery App* (Kurir) — serta *Web Admin Dashboard* untuk pengurus koperasi. Lapisan klien tidak memproses data secara mandiri, melainkan murni mengonsumsi layanan API dari server melalui protokol HTTPS dengan format pertukaran data JSON.

2. **Lapisan Cloud Server (*Backend Layer*):** Bertindak sebagai pusat kendali utama yang dibangun menggunakan *framework* **Laravel** dan memaparkan layanan **RESTful API**. Seluruh logika aplikasi diproses di sini, mulai dari penyaringan permintaan oleh *Middleware* (autentikasi token dan validasi *Role-Based Access Control*), eksekusi logika bisnis pada *Controller* dan *Service* (seperti proses *checkout*, *split payment*, dan akumulasi poin), hingga pemetaan objek ke basis data melalui *Eloquent ORM*.

3. **Lapisan Data:** Seluruh data transaksional dan master disimpan secara permanen pada basis data relasional **MySQL/MariaDB**, mencakup data pengguna, produk, pesanan, pembayaran, poin, serta simpanan anggota.

Selain ketiga lapisan tersebut, sistem terhubung dengan **layanan eksternal** berupa *Payment Gateway* (QRIS/Transfer) untuk memvalidasi pembayaran digital secara *real-time* tanpa membebani logika inti aplikasi.

---

## 3.2 Pemodelan Sistem dan Data

Pada tahap pemodelan sistem dan data, arsitektur perangkat lunak SmartMart diterjemahkan ke dalam bentuk cetak biru (*blueprint*) visual. Perancangan platform ini dibangun menggunakan pendekatan berorientasi objek (*Object-Oriented*) dan mengacu pada standar *Unified Modeling Language* (UML) untuk memodelkan fungsionalitas serta perilaku sistem. Untuk merancang struktur basis data relasional, pemodelan dilakukan menggunakan notasi *Entity Relationship Diagram* (ERD) dengan notasi *Crow's Foot*. Rangkaian cetak biru yang disusun mencakup **Use Case Diagram, Entity Relationship Diagram, Class Diagram, dan Sequence Diagram**.

### 3.2.1 Use Case Diagram

Pemodelan fungsionalitas dirancang menggunakan *Use Case Diagram* untuk memvisualisasikan batas sistem (*system boundary*) serta interaksi aktor dengan sistem, pada lingkup **Modul Pengguna**. Modul ini berfokus pada satu aktor utama, yaitu **Anggota (Pengguna)**.

**[SISIPKAN Gambar 3.2 — Use Case Diagram Anggota SmartMart] (design/use_case_smartmart.png)**

Berdasarkan Gambar 3.2, Anggota dapat mengakses fungsionalitas berikut:

1. **Autentikasi:** Registrasi Akun, Login Sistem, dan akses Dashboard Pengguna.
2. **Belanja & Pesanan:** Lihat Katalog Produk, Cari & Filter Produk, Kelola Wishlist, dan Checkout Pesanan.
3. **Transaksi & Profil:** Lihat & Lacak Pesanan, Tulis Review Produk, serta Kelola Profil & Alamat.
4. **Poin & Simpanan:** Lihat Saldo & Poin, Tukar/Konversi Poin, dan Lihat Saldo Simpanan.

Diagram ini memodelkan relasi `«include»` untuk perilaku wajib, yaitu proses **Login** yang selalu menyertakan **Validasi Role (RBAC)**, serta proses **Checkout** yang selalu menyertakan **Konfirmasi Pembayaran** dan **perolehan Reward Poin** (1% dari total transaksi).

### 3.2.2 Entity Relationship Diagram

Untuk merepresentasikan struktur penyimpanan data di dalam *backend* Laravel, pemodelan data dirancang menggunakan *Entity Relationship Diagram* (ERD) dengan notasi *Crow's Foot*. Skema ini menggambarkan tabel-tabel utama (entitas) yang saling terhubung (*relational database*) pada arsitektur SmartMart secara menyeluruh.

**[SISIPKAN Gambar 3.3 — Entity Relationship Diagram SmartMart] (design/erd_smartmart.png)**

Berdasarkan ERD pada Gambar 3.3, pangkalan data (MySQL/MariaDB) dinormalisasi dengan rancangan sebagai berikut:

1. **Entitas Pusat (Pengguna & Akses):** Tabel `users` menjadi inti sistem dengan kardinalitas *One-to-Many* terhadap berbagai tabel transaksional. Tabel ini diperkuat oleh `member_profiles`, `addresses`, dan `seller_profiles` untuk memisahkan data profil sesuai peran (*RBAC*).
2. **Entitas Produk & Ulasan:** Tabel `categories` mengklasifikasikan `products`, sedangkan `product_reviews` dan `wishlists` mencatat interaksi anggota terhadap produk.
3. **Entitas Order & Logistik:** Tabel `orders` memuat banyak `order_items` (*One-to-Many*), memiliki tepat satu `payments` (*One-to-One*), serta satu `shipments` untuk pelacakan logistik.
4. **Entitas Poin, Voucher & Simpanan:** Tabel `user_points` dan `point_transactions` mengelola sistem loyalitas, `vouchers` menyimpan promo diskon, sedangkan `savings` mengakomodasi simpanan anggota (pokok/wajib/sukarela). Penerapan *Foreign Key* pada seluruh relasi menjamin integritas data tanpa redundansi yang berarti.

### 3.2.3 Class Diagram

Untuk menggambarkan struktur statis perangkat lunak, perancangan *Class Diagram* difokuskan pada arsitektur *backend* berbasis *framework* Laravel. Arsitektur ini menitikberatkan pada interaksi antara lapisan *Controller* pengelola logika (*business logic*) dan lapisan *Model* representasi data (*Eloquent ORM*), khusus pada cakupan Modul Pengguna.

**[SISIPKAN Gambar 3.4 — Class Diagram SmartMart (Backend Laravel)] (design/class_diagram_smartmart.png)**

Berdasarkan Gambar 3.4, arsitektur kelas *backend* dirancang dengan karakteristik berorientasi objek (OOP) sebagai berikut:

1. **Pewarisan (*Inheritance*):** Seluruh entitas data (`User`, `Product`, `Order`, `OrderItem`, `Payment`, `Voucher`, `UserPoint`, `PointTransaction`, `Address`) merupakan turunan dari *abstract class* `Model` bawaan Laravel. Begitu pula kelas penangan rute API (`AuthApiController`, `OrderApiController`, `PaymentApiController`, `PointApiController`, `ProductApiController`, `ProfileApiController`) yang mewarisi *abstract class* `Controller`.
2. **Enkapsulasi Atribut & Metode:** Atribut pada kelas *Model* (seperti `harga`, `stok`, `balance`) dilindungi dengan *modifier* `private` (`-`), sementara metode fungsional (`decrementStock()`, `verify()`, `redeem()`) diberi hak akses `public` (`+`) agar dapat dieksekusi melalui pemanggilan API.
3. **Ketergantungan (*Dependency*):** Terdapat relasi ketergantungan antara *Controller* dan *Model*. Misalnya, saat aplikasi melakukan *POST* ke `OrderApiController`, controller akan menginstansiasi objek `Order` dan `Product` untuk memvalidasi stok dan menyimpan pesanan, menjaga *Separation of Concerns* pada *backend*.

### 3.2.4 Sequence Diagram

*Sequence Diagram* memodelkan alur interaksi antar-objek berdasarkan urutan waktu (*time-based*), menjabarkan pertukaran pesan antara *Frontend*, *Controller* API, dan *Database*. Berikut empat fungsionalitas paling krusial pada Modul Pengguna:

**A. Skenario Autentikasi & Login** — Anggota memasukkan kredensial; `AuthApiController` memvalidasi melalui `User` Model ke *Database*. Jika *hash password* benar, dihasilkan token sesi (HTTP 200); jika gagal, ditolak (HTTP 401).

**[SISIPKAN Gambar 3.5 — Sequence Diagram Autentikasi & Login] (design/seq_login.png)**

**B. Skenario Checkout & Pemesanan** — `OrderApiController` memvalidasi stok via `Product` Model, membuat `Order` + `order_items`, mengurangi stok, membuat catatan `Payment` *pending*, dan menambahkan reward poin, lalu mengembalikan invoice (HTTP 201). Instruksi bayar diteruskan ke *Payment Gateway*.

**[SISIPKAN Gambar 3.6 — Sequence Diagram Checkout & Pemesanan] (design/seq_checkout.png)**

**C. Skenario Penukaran / Konversi Poin** — `PointApiController` mengecek saldo via `UserPoint` Model; bila mencukupi, saldo dikurangi dan dicatat pada `point_transactions` (HTTP 200); bila tidak, HTTP 422.

**[SISIPKAN Gambar 3.7 — Sequence Diagram Penukaran Poin] (design/seq_poin.png)**

**D. Skenario Pelacakan Pengiriman** — Anggota membuka detail pesanan dan memilih "Lacak"; `OrderApiController` mengambil data resi & status via `Shipment` Model dari *Database*, lalu menampilkan posisi dan estimasi tiba (HTTP 200). Status pengiriman (dikemas → dikirim → transit → sampai) diperbarui oleh Ekspedisi/Kurir melalui *endpoint* terpisah.

**[SISIPKAN Gambar 3.8 — Sequence Diagram Pelacakan Pengiriman] (design/seq_tracking.png)**

---

## 3.3 Perancangan Antarmuka Pengguna

Perancangan antarmuka pengguna SmartMart disusun berdasarkan prinsip *user-centered design* dengan mengacu pada heuristik *usability* Jakob Nielsen, sehingga visualisasi informasi disajikan secara konsisten, intuitif, dan minim friksi pada setiap titik interaksi. Mengingat platform melibatkan empat peran pengguna dengan kebutuhan kerja yang berbeda, rancangan antarmuka dikelompokkan ke dalam dua kanal utama, yaitu aplikasi *mobile* berbasis Flutter untuk Pembeli, Penjual, dan Kurir, serta *web dashboard* berbasis Laravel Blade untuk Admin Koperasi. Untuk menjaga keterbacaan dokumen, hanya rancangan antarmuka pada fungsionalitas-fungsionalitas prioritas tinggi yang ditampilkan pada subbab ini, sementara antarmuka pendukung lain disertakan pada lampiran.

### 3.3.1 Kanal Aplikasi Mobile

Pada kanal aplikasi *mobile*, rancangan antarmuka mengikuti pola navigasi *bottom navigation bar* dengan lima tab utama (Beranda, Katalog, Keranjang, Aktivitas, Profil) untuk peran Pembeli, dan pola *sidebar navigation* untuk peran Penjual dan Kurir karena karakter kerja mereka yang lebih operasional. Halaman Beranda Pembeli menampilkan ringkasan saldo poin loyalitas, *banner* promosi koperasi, serta rekomendasi produk berbasis riwayat transaksi. Halaman Katalog menyajikan *grid* produk dengan filter kategori dan rentang harga. Halaman Checkout menerapkan *stepper progress* untuk memandu pengguna melewati pemilihan alamat, pemilihan kurir, hingga konfirmasi pembayaran QRIS. Halaman Tracking Kurir menampilkan peta interaktif (Google Maps SDK) yang menggambarkan posisi kurir secara *real-time* disertai estimasi waktu kedatangan.

**[SISIPKAN Gambar 3.9 — Rancangan Antarmuka Beranda Pembeli] (design/mockup_beranda.png)**

**[SISIPKAN Gambar 3.10 — Rancangan Antarmuka Katalog Produk] (design/mockup_katalog.png)**

**[SISIPKAN Gambar 3.11 — Rancangan Antarmuka Checkout (Stepper)] (design/mockup_checkout.png)**

**[SISIPKAN Gambar 3.12 — Rancangan Antarmuka Tracking Kurir] (design/mockup_tracking.png)**

Untuk peran Penjual, antarmuka berfokus pada pengelolaan toko dengan halaman utama berupa Dashboard Penjual yang menyajikan grafik penjualan harian, jumlah pesanan baru, dan saldo dana siap cair. Halaman Manajemen Produk memungkinkan penambahan, penyuntingan, dan penonaktifan produk dengan formulir input gambar bawaan kamera perangkat. Untuk peran Kurir, antarmuka dirancang seminimum mungkin agar tidak mengganggu fokus berkendara: hanya menampilkan daftar tugas aktif, tombol pemindai QR pengambilan paket, dan tombol *update* status (Diterima, Dalam Perjalanan, Tiba di Tujuan).

**[SISIPKAN Gambar 3.13 — Rancangan Antarmuka Dashboard Penjual] (design/mockup_dashboard_penjual.png)**

**[SISIPKAN Gambar 3.14 — Rancangan Antarmuka Manajemen Produk] (design/mockup_manajemen_produk.png)**

**[SISIPKAN Gambar 3.15 — Rancangan Antarmuka Tugas Kurir] (design/mockup_kurir.png)**

### 3.3.2 Kanal Web Dashboard (Admin)

Pada kanal *web dashboard*, rancangan antarmuka Admin Koperasi mengadopsi pola *three-column layout* yang terdiri atas *sidebar* navigasi modul di sisi kiri, area konten utama di tengah, dan panel aktivitas terkini di sisi kanan. Halaman Dashboard Utama menyajikan empat kartu metrik kunci (jumlah anggota aktif, total transaksi harian, nilai SHU tahun berjalan, dan jumlah pengiriman dalam proses) yang disertai grafik tren tujuh hari terakhir. Halaman Manajemen Anggota menampilkan tabel anggota koperasi dengan kemampuan filter berdasarkan peran dan status verifikasi, serta tombol aksi cepat untuk persetujuan registrasi baru. Halaman Distribusi SHU menyajikan formulir kalkulasi otomatis dengan input persentase pembagian dan *preview* alokasi per anggota sebelum dieksekusi.

**[SISIPKAN Gambar 3.16 — Rancangan Antarmuka Dashboard Admin] (design/mockup_dashboard_admin.png)**

**[SISIPKAN Gambar 3.17 — Rancangan Antarmuka Manajemen Anggota] (design/mockup_manajemen_anggota.png)**

**[SISIPKAN Gambar 3.18 — Rancangan Antarmuka Distribusi SHU] (design/mockup_distribusi_shu.png)**

---

## 3.4 Kebutuhan Perangkat Keras dan Perangkat Lunak

### 3.4.1 Pengembangan Sistem

| Kategori | Spesifikasi Minimum / Komponen |
|---|---|
| Perangkat Komputasi | Laptop/PC prosesor setara Intel Core i5 / AMD Ryzen 5 |
| Memori (RAM) | Minimal 8 GB |
| Penyimpanan | Minimal 256 GB SSD |
| Perangkat Uji *Mobile* | *Smartphone* Android / emulator untuk pengujian aplikasi Flutter |

| Kategori | Nama Perangkat Lunak / Teknologi |
|---|---|
| Sistem Operasi | Windows 10/11, macOS, atau Linux |
| Text Editor / IDE | Visual Studio Code / Android Studio |
| Web Server Lokal | Laragon / XAMPP (Apache, MySQL/MariaDB, PHP 8.x) |
| *Framework Backend* | Laravel (PHP) untuk RESTful API |
| *Framework Frontend* | Flutter (Dart) untuk aplikasi *mobile* lintas platform |
| Database | MySQL / MariaDB |
| API Tester | Postman |
| Version Control | Git & GitHub |

### 3.4.2 Implementasi Sistem

| Kategori | Spesifikasi Infrastruktur Production |
|---|---|
| Cloud Hosting (VPS) | *Virtual Private Server* Linux (Ubuntu Server), minimal RAM 2 GB |
| Web Server | Nginx atau Apache HTTP Server |
| Backend & API Layer | Lingkungan *runtime* PHP 8.x dengan *Framework* Laravel |
| Database Server | MySQL / MariaDB Server |
| Layanan Eksternal | *Payment Gateway* (QRIS/Transfer) untuk transaksi digital |
| Aplikasi Klien | Aplikasi *mobile* (Android/iOS) dan *Web Browser* (Chrome/Safari) |

---

## 3.5 Subbab Tambahan (apabila diperlukan)

Apabila dibutuhkan, dapat ditambahkan subbab seperti kajian pustaka spesifik, regulasi koperasi (AD/ART, UU No. 25/1992), atau pembahasan algoritma pendukung.
