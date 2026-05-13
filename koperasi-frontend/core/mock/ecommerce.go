package mock

import "koperasi-frontend/core/model"

// ============================================================
// E-Commerce Mock Data — separate from koperasi data
// ============================================================

// ECommerceUsers holds all e-commerce user accounts.
// Credentials (email + password) are identical to koperasi Users — single login for both systems.
var ECommerceUsers = []model.ECommerceUser{
	{
		ID: 1, Username: "andi_wijaya", Email: "anggota@koperasi.id", Password: "anggota123",
		Role: "BUYER", IsSellerActive: true, SellerRating: 4.7,
		LinkedKoperasiMemberID: 1, // Andi Wijaya — KOP-0001
		CreatedAt: "2026-01-10",
	},
	{
		ID: 2, Username: "rina_pertiwi", Email: "rina@koperasi.id", Password: "rina123",
		Role: "BUYER", IsSellerActive: true, SellerRating: 4.5,
		LinkedKoperasiMemberID: 2, // Rina Pertiwi — KOP-0002
		CreatedAt: "2026-01-15",
	},
	{
		ID: 3, Username: "budi_santoso", Email: "owner@koperasi.id", Password: "owner123",
		Role: "ADMIN", IsSellerActive: false, SellerRating: 0,
		LinkedKoperasiMemberID: 0,
		CreatedAt: "2026-01-01",
	},
	{
		ID: 4, Username: "siti_aminah", Email: "kasir@koperasi.id", Password: "kasir123",
		Role: "BUYER", IsSellerActive: false, SellerRating: 0,
		LinkedKoperasiMemberID: 0,
		CreatedAt: "2026-01-05",
	},
}

// ECSellerProfiles holds seller profile details (for users with IsSellerActive=true).
var ECSellerProfiles = []model.SellerProfile{
	{
		SellerID: 1, StoreName: "Toko Andi Jaya",
		Description:  "Menjual berbagai kebutuhan pokok dan sembako berkualitas dengan harga terjangkau.",
		Rating:       4.7, ResponseTime: "< 1 jam", TotalSold: 234,
		JoinedAt: "2026-01-10",
	},
	{
		SellerID: 2, StoreName: "Rina Craft & Food",
		Description:  "Produk makanan rumahan dan kerajinan tangan khas Bandung.",
		Rating:       4.5, ResponseTime: "< 2 jam", TotalSold: 156,
		JoinedAt: "2026-01-15",
	},
}

// ECProducts holds all e-commerce marketplace products.
var ECProducts = []model.ECProduct{
	// ---- Andi's products (SellerID=1) ----
	{
		ID: 1, SellerID: 1, SellerName: "Toko Andi Jaya",
		Nama: "Beras Organik Premium 5kg", Deskripsi: "Beras organik dari petani lokal, pulen dan wangi. Ditanam tanpa pestisida.",
		Kategori: "Sembako", Harga: 85000, Stok: 50, Berat: 5000,
		FotoURL: "https://placehold.co/400x400/2d1b69/e2e8f0?text=Beras+Organik", Rating: 4.8, TotalReview: 12, TotalSold: 87,
		Status: "APPROVED", CreatedAt: "2026-01-20",
	},
	{
		ID: 2, SellerID: 1, SellerName: "Toko Andi Jaya",
		Nama: "Minyak Goreng Kelapa 2L", Deskripsi: "Minyak goreng dari kelapa murni, lebih sehat dan tahan panas.",
		Kategori: "Sembako", Harga: 42000, Stok: 35, Berat: 2100,
		FotoURL: "https://placehold.co/400x400/2d1b69/e2e8f0?text=Minyak+Goreng", Rating: 4.5, TotalReview: 8, TotalSold: 65,
		Status: "APPROVED", CreatedAt: "2026-01-22",
	},
	{
		ID: 3, SellerID: 1, SellerName: "Toko Andi Jaya",
		Nama: "Gula Aren Bubuk 500g", Deskripsi: "Gula aren asli dari Banten, cocok untuk kopi dan masakan.",
		Kategori: "Sembako", Harga: 35000, Stok: 25, Berat: 500,
		FotoURL: "https://placehold.co/400x400/2d1b69/e2e8f0?text=Gula+Aren", Rating: 4.9, TotalReview: 15, TotalSold: 42,
		Status: "APPROVED", CreatedAt: "2026-02-01",
	},
	{
		ID: 4, SellerID: 1, SellerName: "Toko Andi Jaya",
		Nama: "Kopi Arabika Jawa Barat 250g", Deskripsi: "Biji kopi arabika pilihan dari perkebunan di Jawa Barat.",
		Kategori: "Minuman", Harga: 55000, Stok: 20, Berat: 250,
		FotoURL: "https://placehold.co/400x400/2d1b69/e2e8f0?text=Kopi+Arabika", Rating: 4.7, TotalReview: 10, TotalSold: 30,
		Status: "APPROVED", CreatedAt: "2026-02-05",
	},
	{
		ID: 5, SellerID: 1, SellerName: "Toko Andi Jaya",
		Nama: "Teh Hijau Premium 100g", Deskripsi: "Teh hijau organik, kaya antioksidan dan menyegarkan.",
		Kategori: "Minuman", Harga: 28000, Stok: 40, Berat: 100,
		FotoURL: "https://placehold.co/400x400/2d1b69/e2e8f0?text=Teh+Hijau", Rating: 4.3, TotalReview: 5, TotalSold: 18,
		Status: "PENDING_APPROVAL", CreatedAt: "2026-04-10",
	},

	// ---- Rina's products (SellerID=2) ----
	{
		ID: 6, SellerID: 2, SellerName: "Rina Craft & Food",
		Nama: "Keripik Singkong Pedas 200g", Deskripsi: "Keripik singkong renyah level pedas, produksi rumahan berkualitas.",
		Kategori: "Snack", Harga: 18000, Stok: 60, Berat: 200,
		FotoURL: "https://placehold.co/400x400/1b4d2e/e2e8f0?text=Keripik+Pedas", Rating: 4.6, TotalReview: 20, TotalSold: 98,
		Status: "APPROVED", CreatedAt: "2026-01-25",
	},
	{
		ID: 7, SellerID: 2, SellerName: "Rina Craft & Food",
		Nama: "Dodol Garut 300g", Deskripsi: "Dodol khas Garut, kenyal dan manis legit. Cocok untuk oleh-oleh.",
		Kategori: "Snack", Harga: 25000, Stok: 30, Berat: 300,
		FotoURL: "https://placehold.co/400x400/1b4d2e/e2e8f0?text=Dodol+Garut", Rating: 4.4, TotalReview: 7, TotalSold: 45,
		Status: "APPROVED", CreatedAt: "2026-02-10",
	},
	{
		ID: 8, SellerID: 2, SellerName: "Rina Craft & Food",
		Nama: "Sambal Matah Bali 250ml", Deskripsi: "Sambal matah segar khas Bali, pedas dan harum.",
		Kategori: "Bumbu", Harga: 22000, Stok: 45, Berat: 300,
		FotoURL: "https://placehold.co/400x400/1b4d2e/e2e8f0?text=Sambal+Matah", Rating: 4.8, TotalReview: 11, TotalSold: 67,
		Status: "APPROVED", CreatedAt: "2026-02-15",
	},
	{
		ID: 9, SellerID: 2, SellerName: "Rina Craft & Food",
		Nama: "Tas Rajut Handmade", Deskripsi: "Tas rajut buatan tangan dari bahan katun premium.",
		Kategori: "Fashion", Harga: 120000, Stok: 10, Berat: 200,
		FotoURL: "https://placehold.co/400x400/1b4d2e/e2e8f0?text=Tas+Rajut", Rating: 4.9, TotalReview: 3, TotalSold: 12,
		Status: "APPROVED", CreatedAt: "2026-03-01",
	},
	{
		ID: 10, SellerID: 2, SellerName: "Rina Craft & Food",
		Nama: "Gelang Manik-manik Set", Deskripsi: "Set gelang manik-manik warna-warni, handmade.",
		Kategori: "Fashion", Harga: 35000, Stok: 15, Berat: 50,
		FotoURL: "https://placehold.co/400x400/1b4d2e/e2e8f0?text=Gelang+Manik", Rating: 4.2, TotalReview: 2, TotalSold: 8,
		Status: "PENDING_APPROVAL", CreatedAt: "2026-04-15",
	},
}

// ECAddresses holds shipping addresses.
var ECAddresses = []model.ECAddress{
	{
		ID: 1, UserID: 3, Label: "Rumah", Penerima: "Budi Santoso",
		NoHP: "081345678901", Alamat: "Jl. Cihampelas No. 55, Rt 03/Rw 05",
		Kota: "Bandung", Provinsi: "Jawa Barat", KodePos: "40131", IsDefault: true,
	},
	{
		ID: 2, UserID: 3, Label: "Kantor", Penerima: "Budi Santoso",
		NoHP: "081345678901", Alamat: "Jl. Sudirman No. 100, Gedung A Lt. 3",
		Kota: "Bandung", Provinsi: "Jawa Barat", KodePos: "40261", IsDefault: false,
	},
	{
		ID: 3, UserID: 1, Label: "Rumah", Penerima: "Andi Wijaya",
		NoHP: "081234567890", Alamat: "Jl. Merdeka No. 12",
		Kota: "Bandung", Provinsi: "Jawa Barat", KodePos: "40117", IsDefault: true,
	},
}

// ECProductReviews holds product reviews.
var ECProductReviews = []model.ProductReview{
	{
		ID: 1, ProductID: 1, UserID: 3, Username: "budi",
		Rating: 5, Komentar: "Berasnya bagus, pulen dan wangi. Keluarga suka sekali!",
		CreatedAt: "2026-03-15",
	},
	{
		ID: 2, ProductID: 6, UserID: 3, Username: "budi_santoso",
		Rating: 4, Komentar: "Keripiknya enak dan renyah. Pedasnya pas!",
		CreatedAt: "2026-03-20",
	},
	{
		ID: 3, ProductID: 8, UserID: 1, Username: "andi",
		Rating: 5, Komentar: "Sambal matahnya segar banget, bumbunya terasa. Recommended!",
		CreatedAt: "2026-04-01",
	},
}

// ECWishlists holds user wishlists.
var ECWishlists = []model.Wishlist{
	{ID: 1, UserID: 3, ProductID: 4, CreatedAt: "2026-03-10"},
	{ID: 2, UserID: 3, ProductID: 9, CreatedAt: "2026-03-15"},
	{ID: 3, UserID: 1, ProductID: 7, CreatedAt: "2026-04-01"},
}

// ECVouchers holds promotional vouchers.
var ECVouchers = []model.Voucher{
	{
		ID: 1, Code: "HEMAT10", Deskripsi: "Diskon 10% untuk semua produk",
		TipeDiskon: "PERCENT", NilaiDiskon: 10, MinPembelian: 50000,
		MaksDiskon: 20000, Kuota: 50, Status: "ACTIVE",
		BerlakuSampai: "2026-12-31",
	},
	{
		ID: 2, Code: "GRATIS15K", Deskripsi: "Potongan Rp 15.000 untuk pembelian min Rp 75.000",
		TipeDiskon: "FIXED", NilaiDiskon: 15000, MinPembelian: 75000,
		MaksDiskon: 15000, Kuota: 30, Status: "ACTIVE",
		BerlakuSampai: "2026-06-30",
	},
}

// ECShippingOptions holds available shipping services.
var ECShippingOptions = []model.ShippingOption{
	{ID: 1, Nama: "JNE Reguler", Provider: "JNE", Estimasi: "2-3 hari", Harga: 12000},
	{ID: 2, Nama: "JNE YES", Provider: "JNE", Estimasi: "1 hari", Harga: 25000},
	{ID: 3, Nama: "J&T Express", Provider: "J&T", Estimasi: "2-3 hari", Harga: 10000},
	{ID: 4, Nama: "SiCepat BEST", Provider: "SiCepat", Estimasi: "1-2 hari", Harga: 15000},
}

// ECOrders holds e-commerce orders.
var ECOrders = []model.ECOrder{
	{
		ID: 1, NomorOrder: "EC-20260401-001", BuyerID: 3, BuyerName: "budi_santoso",
		SellerID: 1, SellerName: "Toko Andi Jaya",
		Items: []model.ECOrderItem{
			{ProductID: 1, ProductNama: "Beras Organik Premium 5kg", SellerID: 1, Jumlah: 1, HargaSatuan: 85000, Subtotal: 85000},
			{ProductID: 3, ProductNama: "Gula Aren Bubuk 500g", SellerID: 1, Jumlah: 2, HargaSatuan: 35000, Subtotal: 70000},
		},
		AlamatPengiriman: "Jl. Cihampelas No. 55, Bandung",
		ShippingOption: "JNE Reguler", ShippingCost: 12000,
		Subtotal: 155000, Discount: 0, PointsUsed: 0, TotalHarga: 167000,
		VoucherCode: "", MetodeBayar: "Transfer Bank", Status: "SELESAI",
		ResiPengiriman: "JNE1234567890", PointsEarned: 155,
		CreatedAt: "2026-04-01 10:30", UpdatedAt: "2026-04-04 14:00",
	},
	{
		ID: 2, NomorOrder: "EC-20260410-001", BuyerID: 3, BuyerName: "budi_santoso",
		SellerID: 2, SellerName: "Rina Craft & Food",
		Items: []model.ECOrderItem{
			{ProductID: 6, ProductNama: "Keripik Singkong Pedas 200g", SellerID: 2, Jumlah: 3, HargaSatuan: 18000, Subtotal: 54000},
			{ProductID: 8, ProductNama: "Sambal Matah Bali 250ml", SellerID: 2, Jumlah: 1, HargaSatuan: 22000, Subtotal: 22000},
		},
		AlamatPengiriman: "Jl. Cihampelas No. 55, Bandung",
		ShippingOption: "J&T Express", ShippingCost: 10000,
		Subtotal: 76000, Discount: 0, PointsUsed: 0, TotalHarga: 86000,
		VoucherCode: "", MetodeBayar: "QRIS", Status: "DIKIRIM",
		ResiPengiriman: "JT2345678901", PointsEarned: 76,
		CreatedAt: "2026-04-10 14:15", UpdatedAt: "2026-04-12 09:00",
	},
	{
		ID: 3, NomorOrder: "EC-20260420-001", BuyerID: 1, BuyerName: "andi",
		SellerID: 2, SellerName: "Rina Craft & Food",
		Items: []model.ECOrderItem{
			{ProductID: 9, ProductNama: "Tas Rajut Handmade", SellerID: 2, Jumlah: 1, HargaSatuan: 120000, Subtotal: 120000},
		},
		AlamatPengiriman: "Jl. Merdeka No. 12, Bandung",
		ShippingOption: "SiCepat BEST", ShippingCost: 15000,
		Subtotal: 120000, Discount: 15000, PointsUsed: 0, TotalHarga: 120000,
		VoucherCode: "GRATIS15K", MetodeBayar: "Transfer Bank", Status: "DIPROSES",
		ResiPengiriman: "", PointsEarned: 0,
		CreatedAt: "2026-04-20 08:00", UpdatedAt: "2026-04-20 08:00",
	},
}

// ECUserPoints holds per-user points balances.
var ECUserPoints = []model.UserPoints{
	{UserID: 1, Balance: 5000, TotalEarned: 5500, TotalRedeemed: 500},
	{UserID: 2, Balance: 3000, TotalEarned: 3000, TotalRedeemed: 0},
	{UserID: 3, Balance: 1231, TotalEarned: 1231, TotalRedeemed: 0},
	{UserID: 4, Balance: 0, TotalEarned: 0, TotalRedeemed: 0},
}

// ECPointsTransactions records points activity.
var ECPointsTransactions = []model.PointsTransaction{
	{ID: 1, UserID: 3, Tipe: "EARN_PURCHASE", Amount: 155, OrderID: 1, Keterangan: "Pembelian order EC-20260401-001", CreatedAt: "2026-04-04"},
	{ID: 2, UserID: 3, Tipe: "EARN_PURCHASE", Amount: 76, OrderID: 2, Keterangan: "Pembelian order EC-20260410-001", CreatedAt: "2026-04-12"},
	{ID: 3, UserID: 1, Tipe: "EARN_PURCHASE", Amount: 5500, OrderID: 0, Keterangan: "Akumulasi pembelian sebelumnya", CreatedAt: "2026-03-01"},
	{ID: 4, UserID: 1, Tipe: "REDEEM_DISCOUNT", Amount: -500, OrderID: 0, Keterangan: "Redeem poin untuk diskon belanja", CreatedAt: "2026-03-15"},
	{ID: 5, UserID: 2, Tipe: "EARN_PURCHASE", Amount: 3000, OrderID: 0, Keterangan: "Akumulasi pembelian sebelumnya", CreatedAt: "2026-03-10"},
}

// ECShipmentEvents tracks order shipping events.
var ECShipmentEvents = []model.ShipmentEvent{
	{ID: 1, OrderID: 1, Status: "DIKEMAS", Lokasi: "Gudang Toko Andi Jaya, Bandung", Keterangan: "Pesanan sedang dikemas", CreatedAt: "2026-04-01 15:00"},
	{ID: 2, OrderID: 1, Status: "DIKIRIM", Lokasi: "Sortir JNE Bandung", Keterangan: "Paket diserahkan ke kurir JNE", CreatedAt: "2026-04-02 08:00"},
	{ID: 3, OrderID: 1, Status: "SAMPAI", Lokasi: "Bandung — Alamat Penerima", Keterangan: "Paket diterima oleh Budi Prasetyo", CreatedAt: "2026-04-04 14:00"},
	{ID: 4, OrderID: 2, Status: "DIKEMAS", Lokasi: "Gudang Rina Craft & Food, Bandung", Keterangan: "Pesanan sedang dikemas", CreatedAt: "2026-04-11 09:00"},
	{ID: 5, OrderID: 2, Status: "DIKIRIM", Lokasi: "Sortir J&T Bandung", Keterangan: "Paket diserahkan ke kurir J&T", CreatedAt: "2026-04-12 09:00"},
}

// ECAuditLogs records admin/system actions.
var ECAuditLogs = []model.AuditLog{
	{ID: 1, Action: "APPROVE_PRODUCT", UserID: 3, Username: "budi_santoso", Resource: "product:1", Details: "Approved: Beras Organik Premium 5kg", CreatedAt: "2026-01-21"},
	{ID: 2, Action: "APPROVE_PRODUCT", UserID: 3, Username: "budi_santoso", Resource: "product:6", Details: "Approved: Keripik Singkong Pedas 200g", CreatedAt: "2026-01-26"},
	{ID: 3, Action: "LINK_KOPERASI", UserID: 1, Username: "andi_wijaya", Resource: "member:1", Details: "Linked e-commerce account to koperasi member Andi Wijaya (KOP-0001)", CreatedAt: "2026-01-10"},
}
