package service

import (
	"context"
	"time"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
)

// FeeKoperasiRate: potongan koperasi atas total order POS (3%).
const FeeKoperasiRate = 0.03

// OrderService: aturan bisnis order POS (M3).
type OrderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// Fee menghitung fee koperasi atas total.
func (s *OrderService) Fee(total float64) float64 { return total * FeeKoperasiRate }

func (s *OrderService) List(ctx context.Context, scopeNama string) ([]model.Order, error) {
	return s.repo.List(ctx, scopeNama)
}

func (s *OrderService) Complaints(ctx context.Context) ([]model.Order, error) {
	return s.repo.ListByStatus(ctx, "DISPUTED")
}

func (s *OrderService) FindByID(ctx context.Context, id int) (*model.Order, error) {
	return s.repo.FindByID(ctx, id)
}

// Checkout membuat order POS berstatus DIBAYAR (fee + waktu diisi di sini),
// lalu mendelegasikan persistensi transaksional ke repository.
func (s *OrderService) Checkout(ctx context.Context, pembeliNama string, items []model.OrderItem, total float64, metode string) (*model.Order, error) {
	o := model.Order{
		PembeliNama: pembeliNama,
		Items:       items,
		TotalHarga:  total,
		FeeKoperasi: s.Fee(total),
		Status:      "DIBAYAR",
		MetodeBayar: metode,
		CreatedAt:   time.Now().Format("2006-01-02 15:04"),
	}
	return s.repo.Create(ctx, o)
}

func (s *OrderService) Ship(ctx context.Context, id int) error {
	return s.repo.SetStatus(ctx, id, "DIKIRIM")
}

func (s *OrderService) Complete(ctx context.Context, id int) error {
	return s.repo.SetStatus(ctx, id, "SELESAI")
}

func (s *OrderService) Complain(ctx context.Context, id int, alasan, bukti string) error {
	return s.repo.SetComplaint(ctx, id, alasan, bukti)
}

func (s *OrderService) Resolve(ctx context.Context, id int, approveRetur bool) error {
	return s.repo.Resolve(ctx, id, approveRetur)
}
