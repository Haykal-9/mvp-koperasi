# Draft BAB 3 SmartCoop

Catatan penyusunan: format mengikuti pola `Buku_salman.docx`, yaitu Bab 3 berisi arsitektur sistem, pemodelan sistem dan data, perancangan antarmuka pengguna, kebutuhan perangkat keras/perangkat lunak, dan subbab tambahan bila diperlukan. Isi disesuaikan dengan implementasi aktual pada repo `koperasi-frontend`, bukan template lama yang masih menyebut stack lain.

## Ringkasan Analisis Repositori

MVP SmartCoop yang dianalisis merupakan aplikasi web server-side rendering berbasis Golang. Aplikasi menggunakan framework Gin, template HTML bawaan Go, session cookie dari `gin-contrib/sessions`, Bootstrap melalui CDN, serta data mock in-memory yang didefinisikan sebagai slice struct pada package `core/mock`.

Struktur kode utama berada di folder `koperasi-frontend` dengan komposisi 47 file dan 5.348 baris kode, terdiri dari 2.721 baris Go, 2.186 baris HTML template, 387 baris CSS, dan 54 baris JavaScript. Entry point lokal berada pada `cmd/web/main.go`, sedangkan entry point deployment Vercel berada pada `api/index.go`. Routing, template loader, session middleware, dan protected route didefinisikan pada `core/server/server.go`.

Modul aplikasi dipisahkan berdasarkan handler:

| Modul | File utama | Fungsi utama |
|---|---|---|
| Autentikasi | `core/handler/auth.go` | Login, register akun, logout, dan middleware `AuthRequired` |
| Dashboard | `core/handler/dashboard.go` | Statistik anggota, simpanan, pinjaman, produk, order, dan aktivitas terbaru |
| Keanggotaan | `core/handler/member.go` | Registrasi anggota, approval, simpanan, detail anggota, dan resign |
| Produk | `core/handler/product.go` | Katalog, tambah produk, review produk pending, dan stok |
| Order/POS | `core/handler/order.go` | Cart berbasis session, checkout, order, pengiriman, komplain, dan resolve dispute |
| Pinjaman | `core/handler/loan.go` | Pengajuan, approval, pencairan, generate angsuran, dan pembayaran angsuran |
| Keuangan | `core/handler/finance.go` | Jurnal otomatis/manual dan ringkasan kas, simpanan, piutang, pendapatan |
| Model data | `core/model/models.go` | Struct User, Member, Product, Order, Loan, SimpananTransaction, JournalEntry |
| Data mock | `core/mock/data.go`, `core/mock/helpers.go` | Dummy data, lookup, generator ID, mutasi stok/simpanan/order/pinjaman/jurnal |

## BAB 3 PEMODELAN DAN PERANCANGAN

### 3.1 Arsitektur Sistem

SmartCoop dirancang sebagai aplikasi web client-server berbasis server-side rendering. Pada arsitektur ini, browser pengguna bertindak sebagai client yang mengirimkan request HTTP ke web server, sedangkan server Golang bertanggung jawab mengelola routing, autentikasi, validasi form, pemrosesan logika bisnis, pengambilan data mock, dan rendering halaman HTML sebelum dikirim kembali ke browser.

Secara konseptual, arsitektur sistem dibagi menjadi empat lapisan utama:

1. Lapisan klien
   Lapisan ini merupakan browser yang digunakan oleh Owner, Kasir, dan Anggota untuk mengakses fitur SmartCoop. Seluruh interaksi pengguna seperti login, membuka dashboard, mengisi formulir simpanan, menambah produk ke keranjang, melakukan checkout, atau mengajukan pinjaman dikirim ke server melalui request HTTP GET atau POST.

2. Lapisan web server dan routing
   Lapisan ini dibangun menggunakan Golang dengan framework Gin. File `core/server/server.go` berfungsi sebagai pusat konfigurasi aplikasi, mulai dari pemuatan template, penyediaan static file, konfigurasi session cookie, redirect halaman utama, registrasi route publik, hingga route yang dilindungi middleware `AuthRequired`. Dengan pendekatan ini, seluruh modul berada dalam satu aplikasi monolitik yang ringkas dan mudah diuji pada tahap MVP.

3. Lapisan handler dan logika bisnis
   Setiap modul bisnis memiliki handler tersendiri, misalnya `MemberHandler`, `ProductHandler`, `OrderHandler`, `LoanHandler`, dan `FinanceHandler`. Handler bertugas membaca input form, melakukan validasi, memanggil helper data mock, mengubah status data, menyimpan flash message, dan menentukan halaman tujuan berikutnya. Contohnya, `OrderHandler.DoCheckout` memvalidasi cart, membentuk `OrderItem`, menyimpan order, mengurangi stok produk, membuat jurnal POS otomatis, membersihkan cart session, lalu mengarahkan pengguna ke halaman detail order.

4. Lapisan domain dan penyimpanan data sementara
   Model domain didefinisikan pada `core/model/models.go`, sedangkan data sementara disimpan pada package `core/mock`. Data mock ini mencakup user, anggota, produk, order, pinjaman, transaksi simpanan, log stok, dan jurnal. Karena data masih disimpan di memory, perubahan data akan hilang ketika server restart. Pada implementasi produksi, lapisan ini perlu diganti dengan database relasional agar data bersifat persisten dan memiliki foreign key yang konsisten.

Arsitektur MVP ini dipilih karena sesuai untuk validasi awal proses bisnis koperasi dasar. Sistem dapat memperlihatkan alur end-to-end mulai dari autentikasi, pengelolaan anggota, transaksi simpanan, katalog dan POS, pinjaman, sampai pencatatan jurnal keuangan tanpa kompleksitas deployment database pada tahap prototipe.

### 3.2 Pemodelan Sistem dan Data

Pemodelan sistem dan data dilakukan untuk menerjemahkan kebutuhan fungsional SmartCoop ke dalam rancangan visual yang mudah dipahami sebelum implementasi penuh. Mengikuti format pada dokumen acuan, pemodelan terdiri dari Use Case Diagram, Entity Relationship Diagram, Class Diagram, dan Sequence Diagram. File XML diagram telah disiapkan pada folder `refrensi/uml-diagram` dan dapat dibuka melalui diagrams.net/draw.io.

#### 3.2.1 Use Case Diagram

Use Case Diagram menggambarkan interaksi antara aktor dengan fitur utama SmartCoop. Berdasarkan implementasi akun demo dan modul yang tersedia, terdapat tiga aktor utama, yaitu Owner/Pengurus, Kasir, dan Anggota.

Owner/Pengurus merupakan aktor dengan kewenangan manajerial. Owner dapat memantau dashboard, melakukan review pendaftaran anggota, menyetujui atau menolak anggota, meninjau produk pending, mengelola stok, memantau order, menyelesaikan komplain, menyetujui atau menolak pinjaman, mencairkan pinjaman, serta melihat jurnal dan ringkasan keuangan.

Kasir merupakan aktor operasional yang membantu transaksi harian koperasi. Kasir dapat mencatat transaksi simpanan, mengelola produk dan stok, memproses order, serta membantu pencatatan jurnal manual dan pemantauan keuangan.

Anggota merupakan aktor layanan mandiri. Anggota dapat login, mendaftarkan data anggota, melihat katalog, menambahkan produk ke keranjang, melakukan checkout, melihat riwayat order, mengajukan komplain, mengajukan pinjaman, dan membayar angsuran.

Beberapa use case memiliki hubungan include terhadap pencatatan jurnal otomatis. Transaksi simpanan akan membuat jurnal bertipe `SIMPANAN`, checkout POS akan membuat jurnal bertipe `POS`, sedangkan pencairan dan pembayaran pinjaman akan membuat jurnal bertipe `PINJAMAN`. Mekanisme ini terlihat pada helper `AppendSimpanan`, `AppendOrder`, dan `AppendJournalEntry`.

Gambar yang digunakan:
`refrensi/uml-diagram/use-case-diagram.drawio.xml`

#### 3.2.2 Entity Relationship Diagram

Entity Relationship Diagram digunakan untuk menggambarkan rancangan penyimpanan data berdasarkan struct domain yang saat ini digunakan aplikasi. Entitas utama yang teridentifikasi adalah `User`, `Member`, `Product`, `StockChange`, `Order`, `OrderItem`, `Loan`, `Installment`, `SimpananTransaction`, dan `JournalEntry`.

Relasi utama dalam rancangan data SmartCoop adalah sebagai berikut:

| Relasi | Keterangan |
|---|---|
| Member ke SimpananTransaction | Satu anggota dapat memiliki banyak transaksi simpanan |
| Member ke Loan | Satu anggota dapat memiliki banyak pinjaman |
| Loan ke Installment | Satu pinjaman memiliki banyak jadwal angsuran |
| Product ke StockChange | Satu produk memiliki banyak riwayat perubahan stok |
| Order ke OrderItem | Satu order terdiri dari satu atau banyak item |
| Product ke OrderItem | Satu produk dapat muncul pada banyak item order |
| SimpananTransaction ke JournalEntry | Transaksi simpanan menghasilkan jurnal otomatis |
| Order ke JournalEntry | Checkout order menghasilkan jurnal POS otomatis |
| Loan ke JournalEntry | Pencairan dan angsuran pinjaman menghasilkan jurnal pinjaman |

Pada kode MVP, beberapa relasi masih disimpan menggunakan nama, seperti `member_nama`, `pembeli_nama`, dan `penjual_nama`. Untuk implementasi database produksi, atribut tersebut sebaiknya tetap dipertahankan sebagai snapshot tampilan, tetapi relasi utama perlu menggunakan foreign key seperti `member_id`, `buyer_id`, `seller_id`, dan `order_id` agar integritas data lebih terjamin.

Gambar yang digunakan:
`refrensi/uml-diagram/entity-relationship-diagram.drawio.xml`

#### 3.2.3 Class Diagram

Class Diagram menggambarkan struktur statis kode SmartCoop. Aplikasi memiliki tiga kelompok kelas atau komponen utama, yaitu package server, package handler, dan package model/mock.

Package `server` berperan sebagai pengatur aplikasi. Fungsi `SetupApp` membuat instance Gin, mengaktifkan session cookie, memuat template, menyajikan static file, mendaftarkan route publik, dan membuat group route terlindungi. Fungsi `renderPage` menjadi mekanisme umum untuk merender halaman HTML sekaligus menyisipkan data session, flash message, dan jumlah item cart.

Package `handler` berisi class handler per modul. Setiap handler menyimpan dependensi `Render Renderer` sehingga handler tidak perlu mengetahui detail pemuatan template. Pola ini membuat handler fokus pada logika modul, sedangkan rendering halaman dikelola oleh package server.

Package `model` berisi struct domain, sementara package `mock` berisi data in-memory dan helper operasi data. Helper seperti `AppendSimpanan`, `AppendOrder`, `AppendLoan`, `GenerateInstallments`, dan `AppendJournalEntry` menjadi pusat mutasi data. Dengan pola ini, logika perubahan data tidak tersebar sepenuhnya di template, tetapi dikendalikan oleh handler dan helper.

Gambar yang digunakan:
`refrensi/uml-diagram/class-diagram.drawio.xml`

#### 3.2.4 Sequence Diagram

Sequence Diagram digunakan untuk menjelaskan urutan interaksi antar komponen pada skenario utama aplikasi. Pada SmartCoop, sequence diagram difokuskan pada empat proses yang paling mewakili modul inti:

A. Skenario autentikasi login

Pengguna membuka halaman login, browser mengirim request `GET /login`, lalu `AuthHandler.ShowLogin` merender template login. Setelah pengguna mengisi email dan password, browser mengirim `POST /login`. `AuthHandler.DoLogin` membaca form, mencari user melalui `mock.FindUserByEmail`, memvalidasi password, menyimpan `user_id`, `user_email`, `user_nama`, dan `user_role` ke session cookie, kemudian mengarahkan pengguna ke dashboard. Route dashboard hanya dapat diakses setelah middleware `AuthRequired` menemukan session yang valid.

B. Skenario pendaftaran dan approval anggota

Pengguna membuka form pendaftaran anggota melalui `GET /members/register`. Setelah form disubmit, `MemberHandler.DoRegister` memvalidasi field wajib dan panjang NIK. Jika valid, data anggota baru ditambahkan ke `mock.Members` dengan status `PENDING`. Owner kemudian membuka daftar anggota pending dan menjalankan aksi approve atau reject. Pada proses approve, handler memastikan status masih `PENDING`, mengubah status menjadi `AKTIF`, membuat nomor anggota baru melalui `mock.GenerateNomorAnggota`, lalu mengarahkan kembali ke daftar anggota.

C. Skenario checkout POS

Anggota memilih produk dan jumlah melalui katalog. `OrderHandler.AddToCart` memvalidasi keberadaan produk dan ketersediaan stok, lalu menyimpan item ke session cart. Pada checkout, `OrderHandler.DoCheckout` mengambil cart, melakukan validasi ulang stok, membentuk list `OrderItem`, menghitung total dan fee koperasi, kemudian memanggil `mock.AppendOrder`. Helper ini membuat nomor order, menyimpan order, mengurangi stok melalui `AppendStockChange`, dan menambahkan `JournalEntry` bertipe `POS`. Setelah berhasil, cart dibersihkan dan pengguna diarahkan ke detail order.

D. Skenario pinjaman dan angsuran

Anggota mengisi nominal, tenor, dan tujuan pinjaman. `LoanHandler.DoApply` memvalidasi nominal minimal, batas tenor, dan tujuan, lalu menyimpan pinjaman berstatus `PENDING` melalui `mock.AppendLoan`. Owner dapat menyetujui atau menolak pinjaman. Jika disetujui, proses pencairan mengubah status menjadi `AKTIF`, mengisi tanggal cair, membuat jadwal angsuran melalui `mock.GenerateInstallments`, dan menulis jurnal pencairan. Saat angsuran dibayar melalui `LoanHandler.Pay`, sistem mencari installment berikutnya yang belum dibayar, menandainya sebagai `DIBAYAR`, mengurangi `SisaPokok`, mencatat jurnal angsuran, dan mengubah status pinjaman menjadi `LUNAS` apabila semua installment sudah dibayar.

Gambar yang digunakan:
`refrensi/uml-diagram/sequence-diagram.drawio.xml`

### 3.3 Perancangan Antarmuka Pengguna

Perancangan antarmuka pengguna pada SmartCoop mengikuti kebutuhan tiga aktor utama, yaitu Owner/Pengurus, Kasir, dan Anggota. Karena aplikasi dibangun sebagai web SSR, setiap halaman dirender dari template HTML yang berada pada folder `api/templates`. Layout umum menggunakan `layouts/base.html` untuk halaman setelah login dan `layouts/auth.html` untuk halaman autentikasi.

Rancangan antarmuka utama yang tersedia pada MVP adalah sebagai berikut:

| Kelompok UI | Template | Deskripsi |
|---|---|---|
| Autentikasi | `auth/login.html`, `auth/register.html` | Form login dan registrasi akun |
| Dashboard | `dashboard/index.html` | Ringkasan anggota aktif, simpanan, pinjaman, produk pending, order, dan aktivitas terbaru |
| Anggota | `member/list.html`, `member/detail.html`, `member/register_member.html`, `member/simpanan.html`, `member/resign.html` | Daftar anggota, detail profil, pendaftaran, simpanan, dan pengunduran diri |
| Produk | `product/catalog.html`, `product/detail.html`, `product/create.html`, `product/review.html`, `product/stock.html` | Katalog, detail, tambah produk, review produk pending, dan stok |
| Order | `order/cart.html`, `order/checkout.html`, `order/history.html`, `order/detail.html`, `order/complain.html`, `order/complaints.html` | Keranjang, checkout, riwayat order, detail order, komplain, dan daftar dispute |
| Pinjaman | `loan/apply.html`, `loan/list.html`, `loan/detail.html` | Pengajuan pinjaman, daftar pinjaman, detail pinjaman, dan jadwal angsuran |
| Keuangan | `finance/journals.html`, `finance/create_journal.html`, `finance/summary.html` | Jurnal keuangan, input jurnal manual, dan ringkasan keuangan |

Dokumentasi screenshot aplikasi sudah tersedia pada folder `koperasi-frontend/docs/screenshots`. Screenshot tersebut dapat digunakan sebagai gambar rancangan atau bukti tampilan pada Bab 3 dan Bab 4, misalnya halaman login, dashboard, daftar anggota, simpanan, katalog produk, checkout, daftar pinjaman, detail pinjaman, jurnal, dan ringkasan keuangan.

### 3.4 Kebutuhan Perangkat Keras dan Perangkat Lunak

#### 3.4.1 Pengembangan Sistem

Kebutuhan pengembangan sistem mengacu pada perangkat yang digunakan untuk menulis kode, menjalankan server lokal, dan melakukan pengujian fungsional MVP.

| Kategori | Spesifikasi atau komponen |
|---|---|
| Perangkat komputasi | Laptop/PC dengan prosesor minimal setara Intel Core i5 atau AMD Ryzen 5 |
| Memori | Minimal 8 GB RAM |
| Penyimpanan | Minimal 256 GB SSD |
| Sistem operasi | Windows, Linux, atau macOS |
| Bahasa pemrograman | Golang sesuai versi pada `go.mod` |
| Framework backend/web | Gin `github.com/gin-gonic/gin` |
| Session | `github.com/gin-contrib/sessions` dengan cookie store |
| Template | `html/template` bawaan Go |
| Frontend | HTML, CSS, JavaScript, Bootstrap 5.3, Bootstrap Icons |
| Tools pengembangan | Git, terminal PowerShell atau shell lain, browser modern |

#### 3.4.2 Implementasi Sistem

Kebutuhan implementasi sistem untuk menjalankan MVP masih relatif ringan karena data disimpan secara in-memory dan halaman dirender langsung oleh server.

| Kategori | Spesifikasi atau komponen |
|---|---|
| Server aplikasi | Mesin lokal atau hosting yang dapat menjalankan binary Go |
| Memori server | Minimal 1 GB RAM untuk demo/MVP |
| Browser pengguna | Chrome, Edge, Firefox, atau browser modern lain |
| Koneksi jaringan | HTTP lokal atau internet jika dideploy |
| Penyimpanan data | In-memory mock data pada MVP |
| Deployment alternatif | Vercel function melalui `api/index.go` dan embedded assets |

Untuk implementasi produksi, sistem perlu ditambah database relasional seperti PostgreSQL atau MySQL, hashing password, authorization per role pada setiap route sensitif, logging, backup data, dan konfigurasi secret yang aman.

### 3.5 Batasan MVP dan Rekomendasi Pengembangan

MVP SmartCoop sudah mampu memperagakan proses bisnis dasar koperasi dari hulu ke hilir, tetapi masih memiliki beberapa batasan yang perlu dicatat dalam laporan agar ruang lingkup sistem jelas:

1. Data masih disimpan di memory sehingga akan hilang saat server restart.
2. Password pada dummy user masih berupa teks biasa untuk kebutuhan demo, sehingga implementasi produksi wajib memakai hashing.
3. Middleware saat ini baru memastikan pengguna sudah login, sementara pembatasan role detail pada route sensitif belum diterapkan penuh.
4. Relasi data masih menggunakan nama pada beberapa struct, sehingga rancangan database produksi perlu menggunakan foreign key berbasis ID.
5. Belum tersedia modul SHU pada kode MVP saat ini, walaupun referensi proses bisnis sudah menempatkan SHU sebagai kebutuhan koperasi yang penting.

Rekomendasi pengembangan berikutnya adalah memigrasikan data mock ke database relasional, menambahkan role-based authorization pada setiap route, menerapkan hashing password, membuat audit trail transaksi, serta melanjutkan implementasi modul SHU agar sistem koperasi dasar menjadi lebih lengkap dan sesuai kebutuhan laporan akhir.
