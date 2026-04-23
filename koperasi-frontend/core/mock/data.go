package mock

import "koperasi-frontend/core/model"

var Users = []model.User{
	{ID: 1, Email: "owner@koperasi.id", Password: "owner123", Role: "OWNER", Nama: "Budi Santoso"},
	{ID: 2, Email: "kasir@koperasi.id", Password: "kasir123", Role: "KASIR", Nama: "Siti Aminah"},
	{ID: 3, Email: "anggota@koperasi.id", Password: "anggota123", Role: "ANGGOTA", Nama: "Andi Wijaya"},
	{ID: 4, Email: "rina@koperasi.id", Password: "rina123", Role: "ANGGOTA", Nama: "Rina Pertiwi"},
}

var Members = []model.Member{
	{
		ID: 1, NomorAnggota: "KOP-0001", Nama: "Andi Wijaya", NIK: "3201010101900001",
		Alamat: "Jl. Merdeka No. 12, Bandung", NoHP: "081234567890", Status: "AKTIF",
		SimpananPokok: 500000, SimpananWajib: 1200000, SimpananSukarela: 2500000,
		TanggalMasuk: "2024-01-15",
	},
	{
		ID: 2, NomorAnggota: "KOP-0002", Nama: "Rina Pertiwi", NIK: "3201020202910002",
		Alamat: "Jl. Asia Afrika No. 45, Bandung", NoHP: "081298765432", Status: "AKTIF",
		SimpananPokok: 500000, SimpananWajib: 900000, SimpananSukarela: 1500000,
		TanggalMasuk: "2024-03-20",
	},
	{
		ID: 3, NomorAnggota: "KOP-0003", Nama: "Dewi Lestari", NIK: "3201030303920003",
		Alamat: "Jl. Dago No. 78, Bandung", NoHP: "082134567891", Status: "AKTIF",
		SimpananPokok: 500000, SimpananWajib: 600000, SimpananSukarela: 800000,
		TanggalMasuk: "2024-06-10",
	},
	{
		ID: 4, NomorAnggota: "-", Nama: "Joko Priyono", NIK: "3201040404930004",
		Alamat: "Jl. Braga No. 22, Bandung", NoHP: "085712345678", Status: "PENDING",
		SimpananPokok: 0, SimpananWajib: 0, SimpananSukarela: 0,
		TanggalMasuk: "2026-04-15",
	},
	{
		ID: 5, NomorAnggota: "KOP-0004", Nama: "Maya Sari", NIK: "3201050505940005",
		Alamat: "Jl. Setiabudi No. 100, Bandung", NoHP: "081345678901", Status: "NON_AKTIF",
		SimpananPokok: 500000, SimpananWajib: 300000, SimpananSukarela: 0,
		TanggalMasuk: "2023-08-05",
	},
}

var Products = []model.Product{
	{
		ID: 1, Nama: "Beras Premium 5kg", Kategori: "Sembako", Harga: 75000, Stok: 50,
		BatasStokMinimum: 10, Deskripsi: "Beras premium kualitas terbaik, pulen dan wangi.",
		FotoURL: "https://placehold.co/400x300?text=Beras", PenjualNama: "Koperasi", Status: "APPROVED",
	},
	{
		ID: 2, Nama: "Minyak Goreng 2L", Kategori: "Sembako", Harga: 35000, Stok: 8,
		BatasStokMinimum: 15, Deskripsi: "Minyak goreng kemasan 2 liter.",
		FotoURL: "https://placehold.co/400x300?text=Minyak", PenjualNama: "Koperasi", Status: "APPROVED",
	},
	{
		ID: 3, Nama: "Kopi Bubuk 250gr", Kategori: "Minuman", Harga: 28000, Stok: 30,
		BatasStokMinimum: 5, Deskripsi: "Kopi bubuk arabika asli Jawa Barat.",
		FotoURL: "https://placehold.co/400x300?text=Kopi", PenjualNama: "Andi Wijaya", Status: "APPROVED",
	},
	{
		ID: 4, Nama: "Keripik Singkong", Kategori: "Snack", Harga: 15000, Stok: 20,
		BatasStokMinimum: 5, Deskripsi: "Keripik singkong renyah produksi rumahan.",
		FotoURL: "https://placehold.co/400x300?text=Keripik", PenjualNama: "Rina Pertiwi", Status: "PENDING",
	},
	{
		ID: 5, Nama: "Gula Pasir 1kg", Kategori: "Sembako", Harga: 14000, Stok: 45,
		BatasStokMinimum: 10, Deskripsi: "Gula pasir putih bersih.",
		FotoURL: "https://placehold.co/400x300?text=Gula", PenjualNama: "Koperasi", Status: "APPROVED",
	},
}

var Orders = []model.Order{
	{
		ID: 1, NomorOrder: "ORD-20260420-001", PembeliNama: "Andi Wijaya",
		Items: []model.OrderItem{
			{ProductID: 1, ProductNama: "Beras Premium 5kg", Jumlah: 1, HargaSatuan: 75000, Subtotal: 75000},
			{ProductID: 5, ProductNama: "Gula Pasir 1kg", Jumlah: 2, HargaSatuan: 14000, Subtotal: 28000},
		},
		TotalHarga: 103000, FeeKoperasi: 3090, Status: "SELESAI",
		MetodeBayar: "Tunai", CreatedAt: "2026-04-20 10:30",
	},
	{
		ID: 2, NomorOrder: "ORD-20260421-002", PembeliNama: "Rina Pertiwi",
		Items: []model.OrderItem{
			{ProductID: 3, ProductNama: "Kopi Bubuk 250gr", Jumlah: 2, HargaSatuan: 28000, Subtotal: 56000},
		},
		TotalHarga: 56000, FeeKoperasi: 1680, Status: "DIKIRIM",
		MetodeBayar: "QRIS", CreatedAt: "2026-04-21 14:15",
	},
	{
		ID: 3, NomorOrder: "ORD-20260422-003", PembeliNama: "Dewi Lestari",
		Items: []model.OrderItem{
			{ProductID: 2, ProductNama: "Minyak Goreng 2L", Jumlah: 1, HargaSatuan: 35000, Subtotal: 35000},
		},
		TotalHarga: 35000, FeeKoperasi: 1050, Status: "DIBAYAR",
		MetodeBayar: "Saldo Anggota", CreatedAt: "2026-04-22 08:00",
	},
	{
		ID: 4, NomorOrder: "ORD-20260422-004", PembeliNama: "Andi Wijaya",
		Items: []model.OrderItem{
			{ProductID: 4, ProductNama: "Keripik Singkong", Jumlah: 3, HargaSatuan: 15000, Subtotal: 45000},
		},
		TotalHarga: 45000, FeeKoperasi: 1350, Status: "DISPUTED",
		MetodeBayar: "Tunai", CreatedAt: "2026-04-22 11:20",
		KomplainAlasan:  "Barang tidak sesuai — kemasan rusak saat sampai.",
		KomplainTanggal: "2026-04-22 13:00",
		KomplainBukti:   "foto-bukti-rusak.jpg",
	},
}

var Loans = []model.Loan{
	{
		ID: 1, MemberNama: "Andi Wijaya", Nominal: 5000000, TenorBulan: 10,
		BungaPersen: 1.5, Tujuan: "Modal usaha warung", SisaPokok: 3000000,
		Status: "AKTIF", TanggalCair: "2026-01-15",
		Installments: []model.Installment{
			{BulanKe: 1, JatuhTempo: "2026-02-15", NominalPokok: 500000, NominalBunga: 75000, TotalBayar: 575000, Status: "DIBAYAR", TanggalBayar: "2026-02-14"},
			{BulanKe: 2, JatuhTempo: "2026-03-15", NominalPokok: 500000, NominalBunga: 67500, TotalBayar: 567500, Status: "DIBAYAR", TanggalBayar: "2026-03-13"},
			{BulanKe: 3, JatuhTempo: "2026-04-15", NominalPokok: 500000, NominalBunga: 60000, TotalBayar: 560000, Status: "DIBAYAR", TanggalBayar: "2026-04-15"},
			{BulanKe: 4, JatuhTempo: "2026-05-15", NominalPokok: 500000, NominalBunga: 52500, TotalBayar: 552500, Status: "BELUM", TanggalBayar: ""},
			{BulanKe: 5, JatuhTempo: "2026-06-15", NominalPokok: 500000, NominalBunga: 45000, TotalBayar: 545000, Status: "BELUM", TanggalBayar: ""},
		},
	},
	{
		ID: 2, MemberNama: "Rina Pertiwi", Nominal: 3000000, TenorBulan: 6,
		BungaPersen: 1.5, Tujuan: "Biaya sekolah anak", SisaPokok: 3000000,
		Status: "PENDING", TanggalCair: "",
	},
	{
		ID: 3, MemberNama: "Dewi Lestari", Nominal: 10000000, TenorBulan: 12,
		BungaPersen: 1.5, Tujuan: "Renovasi rumah", SisaPokok: 0,
		Status: "LUNAS", TanggalCair: "2025-03-10",
	},
}

var SimpananTransactions = []model.SimpananTransaction{
	{ID: 1, MemberNama: "Andi Wijaya", Jenis: "POKOK", Tipe: "MASUK", Nominal: 500000, Keterangan: "Simpanan pokok pendaftaran", CreatedAt: "2024-01-15"},
	{ID: 2, MemberNama: "Andi Wijaya", Jenis: "WAJIB", Tipe: "MASUK", Nominal: 100000, Keterangan: "Simpanan wajib bulan Januari 2026", CreatedAt: "2026-01-05"},
	{ID: 3, MemberNama: "Andi Wijaya", Jenis: "WAJIB", Tipe: "MASUK", Nominal: 100000, Keterangan: "Simpanan wajib bulan Februari 2026", CreatedAt: "2026-02-05"},
	{ID: 4, MemberNama: "Andi Wijaya", Jenis: "SUKARELA", Tipe: "MASUK", Nominal: 500000, Keterangan: "Setoran sukarela", CreatedAt: "2026-03-10"},
	{ID: 5, MemberNama: "Rina Pertiwi", Jenis: "POKOK", Tipe: "MASUK", Nominal: 500000, Keterangan: "Simpanan pokok pendaftaran", CreatedAt: "2024-03-20"},
}

var StockChanges = []model.StockChange{
	{ID: 1, ProductID: 1, ProductNama: "Beras Premium 5kg", Tipe: "RESTOCK", Jumlah: 50, StokSetelah: 50, Keterangan: "Stok awal", CreatedAt: "2026-04-01"},
	{ID: 2, ProductID: 2, ProductNama: "Minyak Goreng 2L", Tipe: "RESTOCK", Jumlah: 30, StokSetelah: 30, Keterangan: "Stok awal", CreatedAt: "2026-04-01"},
	{ID: 3, ProductID: 2, ProductNama: "Minyak Goreng 2L", Tipe: "KELUAR_PENJUALAN", Jumlah: -22, StokSetelah: 8, Keterangan: "Penjualan akumulasi", CreatedAt: "2026-04-15"},
}

var JournalEntries = []model.JournalEntry{
	{ID: 1, Tanggal: "2026-04-20", Keterangan: "Penjualan POS ORD-20260420-001", AkunDebit: "Kas", AkunKredit: "Pendapatan Penjualan", Nominal: 103000, TipeTransaksi: "POS"},
	{ID: 2, Tanggal: "2026-04-21", Keterangan: "Penjualan POS ORD-20260421-002", AkunDebit: "Kas", AkunKredit: "Pendapatan Penjualan", Nominal: 56000, TipeTransaksi: "POS"},
	{ID: 3, Tanggal: "2026-04-05", Keterangan: "Simpanan wajib Andi Wijaya Apr 2026", AkunDebit: "Kas", AkunKredit: "Simpanan Wajib", Nominal: 100000, TipeTransaksi: "SIMPANAN"},
	{ID: 4, Tanggal: "2026-04-15", Keterangan: "Angsuran pinjaman Andi Wijaya bulan ke-3", AkunDebit: "Kas", AkunKredit: "Piutang Anggota", Nominal: 560000, TipeTransaksi: "PINJAMAN"},
	{ID: 5, Tanggal: "2026-04-10", Keterangan: "Pembelian ATK kantor", AkunDebit: "Beban Operasional", AkunKredit: "Kas", Nominal: 250000, TipeTransaksi: "MANUAL"},
}

func FindUserByEmail(email string) *model.User {
	for i := range Users {
		if Users[i].Email == email {
			return &Users[i]
		}
	}
	return nil
}

func FindMemberByID(id int) *model.Member {
	for i := range Members {
		if Members[i].ID == id {
			return &Members[i]
		}
	}
	return nil
}

func FindProductByID(id int) *model.Product {
	for i := range Products {
		if Products[i].ID == id {
			return &Products[i]
		}
	}
	return nil
}

func FindOrderByID(id int) *model.Order {
	for i := range Orders {
		if Orders[i].ID == id {
			return &Orders[i]
		}
	}
	return nil
}

func FindLoanByID(id int) *model.Loan {
	for i := range Loans {
		if Loans[i].ID == id {
			return &Loans[i]
		}
	}
	return nil
}
