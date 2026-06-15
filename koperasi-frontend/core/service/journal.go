package service

import (
	"fmt"

	"koperasi-frontend/core/model"
)

// Pembuat entri jurnal otomatis (M2). Memusatkan aturan akun debit/kredit per
// jenis transaksi agar tidak tersebar di handler. Entri yang dikembalikan belum
// punya ID; ID diisi saat persistensi.

// JournalForLoanDisburse: pencairan pinjaman menambah Piutang Anggota (debit)
// dan mengurangi Kas (kredit).
func JournalForLoanDisburse(memberNama string, nominal float64, tanggal string) model.JournalEntry {
	return model.JournalEntry{
		Tanggal:       tanggal,
		Keterangan:    "Pencairan pinjaman " + memberNama,
		AkunDebit:     "Piutang Anggota",
		AkunKredit:    "Kas",
		Nominal:       nominal,
		TipeTransaksi: "PINJAMAN",
	}
}

// JournalForLoanPayment: pembayaran angsuran menambah Kas (debit) dan mengurangi
// Piutang Anggota (kredit).
func JournalForLoanPayment(memberNama string, bulanKe int, total float64, tanggal string) model.JournalEntry {
	return model.JournalEntry{
		Tanggal:       tanggal,
		Keterangan:    fmt.Sprintf("Angsuran pinjaman %s bulan ke-%d", memberNama, bulanKe),
		AkunDebit:     "Kas",
		AkunKredit:    "Piutang Anggota",
		Nominal:       total,
		TipeTransaksi: "PINJAMAN",
	}
}
