package mock

import (
	"fmt"
	"strings"
	"time"

	"koperasi-frontend/core/model"
)

// NextMemberID returns the next ID for a new member.
func NextMemberID() int {
	max := 0
	for _, m := range Members {
		if m.ID > max {
			max = m.ID
		}
	}
	return max + 1
}

// GenerateNomorAnggota produces a fresh "KOP-XXXX" ID slot for an approved member.
func GenerateNomorAnggota() string {
	highest := 0
	for _, m := range Members {
		var n int
		_, err := fmt.Sscanf(m.NomorAnggota, "KOP-%04d", &n)
		if err == nil && n > highest {
			highest = n
		}
	}
	return fmt.Sprintf("KOP-%04d", highest+1)
}

// SimpananByMember returns transactions for a member, latest first.
func SimpananByMember(nama string) []model.SimpananTransaction {
	out := []model.SimpananTransaction{}
	for i := len(SimpananTransactions) - 1; i >= 0; i-- {
		if SimpananTransactions[i].MemberNama == nama {
			out = append(out, SimpananTransactions[i])
		}
	}
	return out
}

// LoansByMember returns loans for a member.
func LoansByMember(nama string) []model.Loan {
	out := []model.Loan{}
	for _, l := range Loans {
		if l.MemberNama == nama {
			out = append(out, l)
		}
	}
	return out
}

// NextProductID returns the next id slot for products.
func NextProductID() int {
	max := 0
	for _, p := range Products {
		if p.ID > max {
			max = p.ID
		}
	}
	return max + 1
}

// StockChangesByProduct returns history (latest first) for a given product.
func StockChangesByProduct(productID int) []model.StockChange {
	out := []model.StockChange{}
	for i := len(StockChanges) - 1; i >= 0; i-- {
		if StockChanges[i].ProductID == productID {
			out = append(out, StockChanges[i])
		}
	}
	return out
}

// AppendStockChange records a stock movement and updates the product's current stock.
func AppendStockChange(productID int, tipe string, jumlah int, keterangan, createdAt string) error {
	p := FindProductByID(productID)
	if p == nil {
		return fmt.Errorf("produk tidak ditemukan")
	}
	p.Stok += jumlah
	if p.Stok < 0 {
		p.Stok -= jumlah
		return fmt.Errorf("stok tidak boleh negatif")
	}
	nextID := 0
	for _, s := range StockChanges {
		if s.ID > nextID {
			nextID = s.ID
		}
	}
	StockChanges = append(StockChanges, model.StockChange{
		ID:          nextID + 1,
		ProductID:   productID,
		ProductNama: p.Nama,
		Tipe:        tipe,
		Jumlah:      jumlah,
		StokSetelah: p.Stok,
		Keterangan:  keterangan,
		CreatedAt:   createdAt,
	})
	return nil
}

// NextOrderID returns the next order id slot.
func NextOrderID() int {
	max := 0
	for _, o := range Orders {
		if o.ID > max {
			max = o.ID
		}
	}
	return max + 1
}

// GenerateNomorOrder produces "ORD-YYYYMMDD-NNN" using the order count today + 1.
func GenerateNomorOrder(today string) string {
	prefix := "ORD-" + strings.ReplaceAll(today, "-", "") + "-"
	count := 0
	for _, o := range Orders {
		if strings.HasPrefix(o.NomorOrder, prefix) {
			count++
		}
	}
	return fmt.Sprintf("%s%03d", prefix, count+1)
}

// AppendOrder records a new order, decrements stock, and writes a journal entry.
func AppendOrder(o model.Order) *model.Order {
	o.ID = NextOrderID()
	if o.NomorOrder == "" {
		o.NomorOrder = GenerateNomorOrder(o.CreatedAt[:10])
	}
	Orders = append(Orders, o)
	added := &Orders[len(Orders)-1]

	// decrement stock + record stock changes
	for _, it := range added.Items {
		_ = AppendStockChange(it.ProductID, "KELUAR_PENJUALAN", -it.Jumlah,
			"Penjualan "+added.NomorOrder, added.CreatedAt[:10])
	}

	// auto journal
	nextJ := 0
	for _, j := range JournalEntries {
		if j.ID > nextJ {
			nextJ = j.ID
		}
	}
	JournalEntries = append(JournalEntries, model.JournalEntry{
		ID:            nextJ + 1,
		Tanggal:       added.CreatedAt[:10],
		Keterangan:    "Penjualan POS " + added.NomorOrder,
		AkunDebit:     "Kas",
		AkunKredit:    "Pendapatan Penjualan",
		Nominal:       added.TotalHarga,
		TipeTransaksi: "POS",
	})
	return added
}

// AppendSimpanan saves a new simpanan transaction and updates the member balance.
func AppendSimpanan(memberID int, jenis, tipe string, nominal float64, keterangan, createdAt string) error {
	m := FindMemberByID(memberID)
	if m == nil {
		return fmt.Errorf("anggota tidak ditemukan")
	}
	delta := nominal
	if tipe == "KELUAR" {
		delta = -nominal
	}
	switch jenis {
	case "POKOK":
		m.SimpananPokok += delta
	case "WAJIB":
		m.SimpananWajib += delta
	case "SUKARELA":
		m.SimpananSukarela += delta
	default:
		return fmt.Errorf("jenis simpanan tidak valid")
	}

	nextID := 0
	for _, t := range SimpananTransactions {
		if t.ID > nextID {
			nextID = t.ID
		}
	}
	SimpananTransactions = append(SimpananTransactions, model.SimpananTransaction{
		ID:         nextID + 1,
		MemberNama: m.Nama,
		Jenis:      jenis,
		Tipe:       tipe,
		Nominal:    nominal,
		Keterangan: keterangan,
		CreatedAt:  createdAt,
	})

	// auto journal entry
	nextJ := 0
	for _, j := range JournalEntries {
		if j.ID > nextJ {
			nextJ = j.ID
		}
	}
	debit, kredit := "Kas", "Simpanan "+jenis
	if tipe == "KELUAR" {
		debit, kredit = "Simpanan "+jenis, "Kas"
	}
	JournalEntries = append(JournalEntries, model.JournalEntry{
		ID:            nextJ + 1,
		Tanggal:       createdAt,
		Keterangan:    fmt.Sprintf("Simpanan %s %s — %s", jenis, tipe, m.Nama),
		AkunDebit:     debit,
		AkunKredit:    kredit,
		Nominal:       nominal,
		TipeTransaksi: "SIMPANAN",
	})
	return nil
}

// NextLoanID returns the next loan id slot.
func NextLoanID() int {
	max := 0
	for _, l := range Loans {
		if l.ID > max {
			max = l.ID
		}
	}
	return max + 1
}

// AppendLoan saves a new loan to mock data.
func AppendLoan(l model.Loan) *model.Loan {
	l.ID = NextLoanID()
	Loans = append(Loans, l)
	return &Loans[len(Loans)-1]
}

// GenerateInstallments creates the monthly installment schedule for an active loan.
func GenerateInstallments(l *model.Loan) {
	if l.TenorBulan <= 0 {
		return
	}
	l.Installments = nil
	pokokPerBulan := l.Nominal / float64(l.TenorBulan)
	sisaPokok := l.Nominal

	base, err := parseDate(l.TanggalCair)
	if err != nil {
		base = time.Now()
	}

	for i := 1; i <= l.TenorBulan; i++ {
		bunga := sisaPokok * l.BungaPersen / 100
		jatuhTempo := base.AddDate(0, i, 0).Format("2006-01-02")
		l.Installments = append(l.Installments, model.Installment{
			BulanKe:      i,
			JatuhTempo:   jatuhTempo,
			NominalPokok: pokokPerBulan,
			NominalBunga: bunga,
			TotalBayar:   pokokPerBulan + bunga,
			Status:       "BELUM",
		})
		sisaPokok -= pokokPerBulan
	}
}

// AppendJournalEntry is a convenience function for auto-journaling.
func AppendJournalEntry(keterangan, akunDebit, akunKredit string, nominal float64, tipe, tanggal string) {
	nextJ := 0
	for _, j := range JournalEntries {
		if j.ID > nextJ {
			nextJ = j.ID
		}
	}
	JournalEntries = append(JournalEntries, model.JournalEntry{
		ID:            nextJ + 1,
		Tanggal:       tanggal,
		Keterangan:    keterangan,
		AkunDebit:     akunDebit,
		AkunKredit:    akunKredit,
		Nominal:       nominal,
		TipeTransaksi: tipe,
	})
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

