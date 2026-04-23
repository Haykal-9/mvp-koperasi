package model

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
	ID               int
	Nama             string
	Kategori         string
	Harga            float64
	Stok             int
	BatasStokMinimum int
	Deskripsi        string
	FotoURL          string
	PenjualNama      string
	Status           string // PENDING, APPROVED, REJECTED
}

type StockChange struct {
	ID          int
	ProductID   int
	ProductNama string
	Tipe        string // RESTOCK, KELUAR_PENJUALAN, KOREKSI
	Jumlah      int    // bisa negatif untuk pengurangan
	StokSetelah int
	Keterangan  string
	CreatedAt   string
}

type OrderItem struct {
	ProductID   int
	ProductNama string
	Jumlah      int
	HargaSatuan float64
	Subtotal    float64
}

type Order struct {
	ID              int
	NomorOrder      string
	PembeliNama     string
	Items           []OrderItem
	TotalHarga      float64
	FeeKoperasi     float64
	Status          string // PENDING, DIBAYAR, DIKIRIM, SELESAI, BATAL, DISPUTED
	MetodeBayar     string
	CreatedAt       string
	KomplainAlasan  string
	KomplainTanggal string
	KomplainBukti   string
}

type Installment struct {
	BulanKe      int
	JatuhTempo   string
	NominalPokok float64
	NominalBunga float64
	TotalBayar   float64
	Status       string // BELUM, DIBAYAR, OVERDUE
	TanggalBayar string
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
	ID            int
	Tanggal       string
	Keterangan    string
	AkunDebit     string
	AkunKredit    string
	Nominal       float64
	TipeTransaksi string // SIMPANAN, POS, PINJAMAN, MANUAL
}
