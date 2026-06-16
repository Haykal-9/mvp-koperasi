package service

import (
	"context"
	"testing"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
)

type fakeECAdminRepository struct {
	repository.ECommerceRepository

	product        *model.ECProduct
	statusCalled   bool
	gotStatus      string
	gotExpected    string
	gotAuditAction repository.AuditAction
}

func (f *fakeECAdminRepository) ProductByID(context.Context, int) (*model.ECProduct, error) {
	return f.product, nil
}

func (f *fakeECAdminRepository) SetECProductStatusWithAudit(_ context.Context, _ int, status, expectedStatus string, audit repository.AuditAction) error {
	f.statusCalled = true
	f.gotStatus = status
	f.gotExpected = expectedStatus
	f.gotAuditAction = audit
	return nil
}

func TestApproveProductUsesAtomicAuditMutation(t *testing.T) {
	repo := &fakeECAdminRepository{
		product: &model.ECProduct{
			ID:         12,
			Nama:       "Kopi",
			SellerName: "Toko Rapi",
			Status:     "PENDING_APPROVAL",
		},
	}
	svc := NewECAdminService(repo)

	_, err := svc.ApproveProduct(context.Background(), 1, "admin", 12)
	if err != nil {
		t.Fatal(err)
	}
	if !repo.statusCalled {
		t.Fatal("service harus memakai mutasi status + audit atomik")
	}
	if repo.gotStatus != "APPROVED" || repo.gotExpected != "PENDING_APPROVAL" {
		t.Fatalf("status=%q expected=%q", repo.gotStatus, repo.gotExpected)
	}
	if repo.gotAuditAction.Action != "APPROVE_PRODUCT" || repo.gotAuditAction.Resource != "product:12" {
		t.Fatalf("audit action salah: %#v", repo.gotAuditAction)
	}
}
