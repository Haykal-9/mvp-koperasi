# BAB 3
# PEMODELAN DAN PERANCANGAN
## (Modul Aplikasi Mobile — Pembeli, Penjual & Kurir)

---

### 3.1 Arsitektur Sistem

Platform **SmartMart** mengimplementasikan arsitektur tersentralisasi berbasis *Cloud* (*Dedicated Single-Tenant System*) guna menjamin isolasi data yang aman, performa transaksi yang stabil, serta kemudahan distribusi layanan ke tangan pengguna. Arsitektur sistem merupakan kerangka kerja struktural yang dirancang untuk menggambarkan interaksi, aliran data, dan batas tanggung jawab antara komponen perangkat lunak, antarmuka pemrograman, dan pangkalan data terpusat.

Infrastruktur ini memisahkan logika aplikasi menjadi tiga lapisan utama (*Three-Tier Architecture*):

1. **Presentation Layer (Aplikasi Mobile Flutter)** — Berjalan secara lokal pada perangkat *smartphone* milik Anggota (Pembeli/Penjual) dan Kurir. Lapisan ini dikembangkan menggunakan *framework cross-platform* Flutter yang berfokus penuh pada *rendering* antarmuka pengguna (*User Interface*) yang responsif, penanganan interaksi masukan data harian, serta visualisasi grafik *dashboard* personal.

2. **Application Layer (Laravel Web Service / RESTful API)** — Bertindak sebagai pusat pengendali logika bisnis (*backend core*). Lapisan ini memproses seluruh permintaan operasional niaga, manajemen pertukaran data JSON via rute *endpoint* API yang terproteksi, otomasi kalkulasi pembagian keuntungan transaksi, serta eksekusi perhitungan insentif loyalitas digital.

3. **Data Layer (MySQL Database)** — Berfungsi sebagai pangkalan data transaksional terpusat untuk menyimpan seluruh informasi relasional secara terstruktur, mulai dari *master data* komoditas produk UMKM, rekapitulasi riwayat pengumpulan poin loyalitas, hingga log pelacakan kurir pengiriman barang.

[Gambar 3-1: Diagram Arsitektur Sistem SmartMart — Platform Mobile (Three-Tier Architecture)]
Sumber diagram: [xml/arsitektur_sistem_mobile.drawio.xml](xml/arsitektur_sistem_mobile.drawio.xml)

---

### 3.2 Pemodelan Sistem dan Data

Pemodelan sistem dan data bertujuan untuk mendeskripsikan secara visual struktur objek, aliran lojikal instruksi program, serta hubungan interrelasi entitas di dalam pangkalan data SmartMart sebelum memasuki tahapan konstruksi kode program.

#### 3.2.1 Pemodelan Berorientasi Objek (UML)

Pemodelan berorientasi objek diimplementasikan menggunakan notasi **Unified Modeling Language (UML)** untuk merepresentasikan fungsionalitas dan perilaku interaktif sistem dari sudut pandang pengguna maupun modul program. Karena platform SmartMart mencakup proses niaga yang masif dan melibatkan interaksi multi-aktor yang kompleks, pemodelan *Use Case* diorganisasikan ke dalam beberapa klaster fungsional menggunakan konsep *folder package* guna menjaga kerapian struktur arsitektur lunak.

##### 1. Use Case Diagram

[Gambar 3-2: Use Case Diagram Platform SmartMart]
Sumber diagram: [xml/use_case_mobile.drawio.xml](xml/use_case_mobile.drawio.xml)

**Aktor Sistem:**

- **Anggota (Pembeli)** — Mengakses pencarian produk, mengelola keranjang, dan melakukan *checkout* belanja.
- **Anggota (Penjual)** — Mengelola katalog dagangan pribadi dan memproses konfirmasi pesanan masuk.
- **Kurir Koperasi** — Menerima manifes distribusi dan memperbarui status pengiriman logistik harian.

**Pengorganisasian Folder Package Use Case:**

| Package | Daftar Use Case |
|---|---|
| **Package Autentikasi** | Login Akun, Registrasi Akun |
| **Package E-Commerce Niaga** | Melihat Katalog Produk, Mengelola Keranjang Belanja, Melakukan Checkout Pesanan, Konfirmasi Pembayaran Digital |
| **Package Manajemen UMKM Penjual** | Kelola Produk Dagangan (CRUD), Konfirmasi Resi Pengiriman |
| **Package Logistik & Loyalitas** | Pembaruan Status Pengiriman (Khusus Kurir), Melihat Histori Saldo & Poin, Konsultasi Asisten AI |

- **Package Autentikasi:** Membawahi *use case* Login dan Registrasi Akun yang diikat oleh aturan proteksi sistem.
- **Package E-Commerce Niaga:** Membawahi *use case* Melihat Katalog Produk, Mengelola Keranjang Belanja, Melakukan *Checkout* Pesanan, dan Konfirmasi Pembayaran Digital.
- **Package Manajemen UMKM Penjual:** Membawahi *use case* Kelola Produk Dagangan (CRUD) dan Konfirmasi Resi Pengiriman.
- **Package Logistik & Loyalitas:** Membawahi *use case* Pembaruan Status Pengiriman (Khusus Kurir), Melihat Histori Saldo & Poin, serta Konsultasi Asisten AI.

##### 2. Class Diagram (Sisi Backend Laravel)

Sesuai dengan ketentuan pembatasan ruang lingkup teknis, *Class Diagram* dirancang secara eksklusif hanya untuk memetakan struktur statis kelas-kelas objek pada komponen *Backend* (Laravel), tanpa melibatkan visualisasi komponen *Frontend* (Flutter). Model ini menggambarkan arsitektur berbasis *Model-Controller* yang menangani operasi data di *server*:

[Gambar 3-3: Backend Class Diagram SmartMart]
Sumber diagram: [xml/class_diagram_mobile.drawio.xml](xml/class_diagram_mobile.drawio.xml)

**Struktur Kelas Pengendali (Controllers & Models):**

| Kelas | Tanggung Jawab |
|---|---|
| `AuthController` | Menangani lojikal token keamanan JWT dan otentikasi akun pengguna |
| `ProductController` | Mengatur fungsi logis manipulasi data inventaris produk milik penjual di *database* |
| `OrderController` | Memproses algoritma *checkout* pesanan, pembobotan tarif ongkir kurir, dan kalkulasi *split payment* otomatis (pemotongan komisi koperasi 3%) |
| `RewardController` | Mengeksekusi algoritma pembagian insentif *Point Rewards* sebesar 1% dari total nilai belanja terverifikasi |
| `AIController` | Menghubungkan *endpoint* aplikasi ke layanan LLM untuk memproses instruksi bahasa natural pengguna |

- **`AuthController`**: Menangani lojikal token keamanan JWT dan otentikasi akun pengguna.
- **`ProductController`**: Mengatur fungsi logis manipulasi data inventaris produk milik penjual di *database*.
- **`OrderController`**: Memproses algoritma *checkout* pesanan, pembobotan tarif ongkir kurir, dan kalkulasi *split payment* otomatis (pemotongan komisi koperasi 3%).
- **`RewardController`**: Mengeksekusi algoritma pembagian insentif *Point Rewards* sebesar 1% dari total nilai belanja terverifikasi.
- **`AIController`**: Menghubungkan *endpoint* aplikasi ke layanan LLM untuk memproses instruksi bahasa natural pengguna.

##### 3. Sequence Diagram

*Sequence diagram* di bawah ini memvisualisasikan urutan interaksi kronologis secara vertikal yang melibatkan tiga komponen utama arsitektur, yaitu: Aplikasi Mobile (Frontend Flutter), Web Service (RESTful API Laravel), dan Pangkalan Data (Backend Database) untuk mengeksekusi fungsionalitas utama transaksi *e-commerce*:

**a. Sequence Diagram: Proses Belanja dan Checkout Pesanan**

[Gambar 3-4: Sequence Diagram — Proses Belanja dan Checkout Pesanan]
Sumber diagram: [xml/sequence_checkout_mobile.drawio.xml](xml/sequence_checkout_mobile.drawio.xml)

Alur Interaksi:

1. Anggota (Pembeli) menekan tombol **"Checkout"** pada layar *Frontend* Flutter.
2. *Frontend* Flutter mengirimkan permintaan HTTP `POST` membawa *payload* data transaksi JSON ke *endpoint* `/api/v1/checkout` pada komponen *RESTful API* Laravel.
3. *RESTful API* Laravel memeriksa ketersediaan stok produk ke *Backend Database* via kueri SQL (`SELECT stock FROM products`).
4. *Backend Database* mengembalikan nilai jumlah stok ke *server backend*.
5. Jika stok mencukupi, *RESTful API* Laravel memerintahkan basis data untuk mengunci inventaris dan membuat baris pesanan baru via kueri (`INSERT INTO orders`).
6. *Backend Database* mengonfirmasi penyimpanan data berhasil.
7. *RESTful API* Laravel meneruskan respons sukses dan menerbitkan *invoice* digital dalam bentuk format JSON ke tampilan layar *Frontend* Flutter.

**b. Sequence Diagram: Konfirmasi Penerimaan Barang dan Otomasi Insentif Poin**

[Gambar 3-5: Sequence Diagram — Konfirmasi Penerimaan Barang & Otomasi Insentif Poin]
Sumber diagram: [xml/sequence_confirm_order_mobile.drawio.xml](xml/sequence_confirm_order_mobile.drawio.xml)

Alur Interaksi:

1. Anggota (Pembeli) menekan tombol **"Konfirmasi Pesanan Diterima"** pada *Frontend* Flutter.
2. *Frontend* Flutter mengirimkan paket parameter transaksi ke *endpoint* API `/api/v1/orders/complete`.
3. *RESTful API* Laravel mengubah status pesanan di *database* menjadi `'Selesai'` (`UPDATE orders SET status = 'completed'`).
4. Secara paralel, *RESTful API* Laravel menghitung pembagian dana bersih penjual setelah dikurangi komisi 3% dan memperbarui saldo toko (`UPDATE sellers SET balance = balance + net_income`).
5. Sistem mengeksekusi lojikal perhitungan bonus *rewards* (1% dari total belanja) dan memasukkannya ke rekap riwayat poin (`INSERT INTO point_histories`).
6. *Backend Database* melakukan pembaruan seluruh nilai baris data finansial tersebut secara instan.
7. *RESTful API* Laravel mengirimkan sinyal status transaksi berhasil ke *Frontend* Flutter untuk memicu visualisasi teks saldo dan poin terbaru di layar ponsel pengguna.

#### 3.2.2 Pemodelan Data (Entity Relationship Diagram)

Perancangan struktur logis pangkalan data SmartMart dimodelkan menggunakan *Entity Relationship Diagram* (ERD) dengan menerapkan skema notasi kaki gagak (*Crow's Foot / NFS Position*) untuk menjamin standardisasi kardinalitas, integritas data (*foreign key constraints*), serta mencegah terjadinya redundansi data saat transaksi massal berlangsung.

[Gambar 3-6: Entity Relationship Diagram SmartMart]
Sumber diagram: [xml/erd_mobile.drawio.xml](xml/erd_mobile.drawio.xml)

**Struktur Relasi Entitas Utama:**

| Entitas | Relasi | Entitas Terhubung | Kardinalitas | Catatan |
|---|---|---|---|---|
| `Users` | menjual | `Products` | 1..N | Satu user bisa menjual banyak produk |
| `Users` | memesan | `Orders` | 1..N | Satu user bisa melakukan banyak pemesanan |
| `Products` | memuat | `Order_Items` | 1..N | Detail pencatatan barang yang dibeli |
| `Orders` | memiliki | `Payments` | 1..1 | Satu pesanan hanya memiliki satu bukti pembayaran |
| `Point_Histories` | dimiliki | `Users` | N..1 | Banyak histori poin kembali ke satu user pemilik |

- **`Users`**: Menyimpan kredensial dasar akun. Memiliki hubungan *One-to-Many* (1 to N) ke entitas `Products` (satu *user* bisa menjual banyak produk) dan entitas `Orders` (satu *user* bisa melakukan banyak pemesanan pesanan).
- **`Products`**: Menyimpan detail komoditas dagangan UMKM anggota. Memiliki relasi *One-to-Many* ke entitas `Order_Items` sebagai detail pencatatan barang yang dibeli.
- **`Orders`**: Menyimpan data induk transaksi niaga (total harga, alamat, status). Memiliki hubungan *One-to-One* (1 to 1) dengan entitas `Payments` (satu pesanan hanya memiliki satu bukti pembayaran digital yang valid).
- **`Point_Histories`**: Menyimpan log perolehan insentif loyalitas digital harian. Terikat melalui relasi *Many-to-One* (N to 1) kembali ke entitas `Users` untuk melacak kepemilikan akumulasi poin dari kontribusi Jasa Usaha anggota secara akurat.

---

### 3.3 Perancangan Antarmuka Pengguna (User Interface Design)

Rancangan antarmuka pengguna pada modul ini disusun murni menggunakan *Orientasi Vertikal Ponsel Genggam* (*Portrait Mobile Layout*) yang dikelompokkan berdasarkan hak akses peran fungsional (*Role-Based UI*):

#### 3.3.1 Antarmuka Sisi Pembeli (Buyer App Interface)

**Halaman Marketplace Katalog**

- **Tujuan:** Menampilkan seluruh produk UMKM yang tersedia bagi anggota koperasi yang ingin berbelanja.
- **Komponen UI utama:**
  - Baris pencarian dinamis untuk menemukan produk secara cepat berdasarkan kata kunci nama komoditas.
  - Filter kategori komoditas (Sembako, Kuliner, Kerajinan) untuk mempersempit hasil tampilan.
  - Susunan kartu produk (*Product Card*) yang memuat foto produk, harga rupiah, nama toko penjual, dan tombol cepat **"+ Keranjang"**.

**Halaman Dashboard Saldo & Poin**

- **Tujuan:** Menyajikan ringkasan finansial digital anggota koperasi secara seketika dari rumah.
- **Komponen UI utama:**
  - Komponen visual kartu dompet digital yang menampilkan indikator saldo simpanan anggota.
  - Visualisasi bintang emas untuk total akumulasi *Point Rewards* yang dapat dipantau seketika.

**Halaman Pelacakan Logistik**

- **Tujuan:** Memberikan transparansi posisi pergerakan paket kurir kepada pembeli secara *real-time*.
- **Komponen UI utama:**
  - Grafik lini masa (*stepper timeline tracking*) yang memperbarui pergerakan posisi paket kurir secara berkala berdasarkan nomor resi pengiriman.

#### 3.3.2 Antarmuka Sisi Penjual (Seller App Interface)

**Halaman Formulir Produk**

- **Tujuan:** Menyediakan formulir pendaftaran produk bagi pelaku UMKM anggota koperasi.
- **Komponen UI utama:**
  - Formulir masukan (*input form*) meliputi kolom nama produk, deskripsi komponen, kuantitas stok variasi, unggah foto galeri produk, serta opsi penentuan harga jual.

**Halaman Grafik Pendapatan**

- **Tujuan:** Memudahkan anggota memantau keuntungan tanpa rekap nota kertas harian.
- **Komponen UI utama:**
  - Diagram visual performa penjualan toko bulanan yang datanya ditarik langsung dari *database server*, menggantikan rekap nota kertas harian.

---

### 3.4 Kebutuhan Perangkat Keras dan Perangkat Lunak

#### 3.4.1 Kebutuhan Pengembangan Sistem

Spesifikasi infrastruktur perangkat keras dan perangkat lunak yang digunakan oleh penulis selama proses analisis, pemodelan diagram, hingga konstruksi kode pemrograman adalah sebagai berikut:

**Kebutuhan Perangkat Keras (Hardware) — Pengembangan:**

| Komponen | Spesifikasi |
|---|---|
| Perangkat | Laptop Lenovo IdeaPad |
| Prosesor (CPU) | Intel Core minimal 4 Core (Clock Speed 1.20 GHz atau lebih tinggi) |
| Memori (RAM) | 16 GB DDR4 |
| Penyimpanan (Storage) | NVMe SSD 512 GB |

**Kebutuhan Perangkat Lunak (Software) — Pengembangan:**

| Kategori | Tools yang Digunakan |
|---|---|
| Sistem Operasi | Windows 11 Home |
| IDE / Code Editor | Visual Studio Code sebagai editor kode utama |
| Emulasi Perangkat | Android Studio (Android SDK) untuk keperluan emulasi gawai |
| API Testing | Postman untuk kebutuhan pengujian integrasi fungsionalitas rute RESTful API |
| Diagram Modeling | Draw.io untuk visualisasi diagram UML/ERD |

#### 3.4.2 Kebutuhan Implementasi Sistem (Spesifikasi Minimal)

Spesifikasi operasional minimum yang direkomendasikan agar sistem SmartMart dapat berjalan lancar di lingkungan operasional nyata tanpa penurunan performa adalah sebagai berikut:

**Sisi Server (Backend & Database Dedicated):**

| Komponen | Spesifikasi Minimum |
|---|---|
| Jenis Server | Virtual Private Server (VPS) Dedicated |
| Prosesor (vCPU) | Minimal 2 vCPU |
| Memori (RAM) | Alokasi 4 GB |
| Penyimpanan (Storage) | 40 GB SSD |
| Sistem Operasi | Ubuntu Server |
| PHP Engine | Versi 8.2 atau lebih baru |
| Database Engine | MySQL Server 8.0 atau lebih baru |

**Sisi Perangkat Pengguna (Mobile Client):**

| Komponen | Spesifikasi Minimum |
|---|---|
| Jenis Perangkat | Smartphone berbasis Android |
| Sistem Operasi | Android minimal versi 8.0 (Oreo) |
| Memori (RAM) | Minimal 2 GB RAM internal |
| Ruang Penyimpanan | Minimal 100 MB ruang kosong untuk instalasi paket APK Flutter |

---

### 3.5 Subbab Tambahan

Tidak diperlukan subbab tambahan untuk Modul Aplikasi Mobile SmartMart (Pembeli, Penjual & Kurir) pada Bab 3. Seluruh aspek pemodelan sistem (arsitektur *three-tier*, *use case*, *class diagram*, *sequence diagram*, ERD), perancangan antarmuka berdasarkan peran (*Role-Based UI*), serta kebutuhan perangkat keras dan perangkat lunak pengembangan maupun implementasi telah dijabarkan secara komprehensif pada subbab-subbab di atas.
