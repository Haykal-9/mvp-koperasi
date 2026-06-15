package service

import (
	"context"
	"strings"
	"testing"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
)

type fakeECConvertRepository struct {
	repository.ECommerceRepository

	user        *model.ECommerceUser
	points      *model.UserPoints
	member      *model.Member
	converted   bool
	gotPoints   float64
	gotRupiah   float64
	newSukarela float64
}

func (f *fakeECConvertRepository) UserByID(context.Context, int) (*model.ECommerceUser, error) {
	return f.user, nil
}

func (f *fakeECConvertRepository) UserPoints(context.Context, int) (*model.UserPoints, error) {
	return f.points, nil
}

func (f *fakeECConvertRepository) MemberByID(context.Context, int) (*model.Member, error) {
	return f.member, nil
}

func (f *fakeECConvertRepository) ConvertPointsToSimpanan(_ context.Context, _, _ int, points, rupiah float64) (float64, error) {
	f.converted = true
	f.gotPoints = points
	f.gotRupiah = rupiah
	return f.newSukarela, nil
}

func TestConvertToSimpananRejectsBelowMinimum(t *testing.T) {
	repo := &fakeECConvertRepository{}
	svc := NewECAccountService(repo)

	_, _, _, err := svc.ConvertToSimpanan(context.Background(), 1, MinPointConversion-1)
	if err == nil || !strings.Contains(err.Error(), "Minimum konversi") {
		t.Fatalf("err = %v, mau minimum konversi", err)
	}
	if repo.converted {
		t.Fatal("repository conversion tidak boleh dipanggil untuk poin di bawah minimum")
	}
}

func TestConvertToSimpananRequiresActiveMember(t *testing.T) {
	repo := &fakeECConvertRepository{
		user:   &model.ECommerceUser{ID: 1, LinkedKoperasiMemberID: 10},
		points: &model.UserPoints{UserID: 1, Balance: 500},
		member: &model.Member{ID: 10, Nama: "Rina", Status: "PENDING"},
	}
	svc := NewECAccountService(repo)

	_, _, _, err := svc.ConvertToSimpanan(context.Background(), 1, MinPointConversion)
	if err == nil || !strings.Contains(err.Error(), "belum aktif") {
		t.Fatalf("err = %v, mau member belum aktif", err)
	}
	if repo.converted {
		t.Fatal("repository conversion tidak boleh dipanggil untuk member belum aktif")
	}
}

func TestConvertToSimpananUsesStandardPointValue(t *testing.T) {
	repo := &fakeECConvertRepository{
		user:        &model.ECommerceUser{ID: 1, LinkedKoperasiMemberID: 10},
		points:      &model.UserPoints{UserID: 1, Balance: 500},
		member:      &model.Member{ID: 10, Nama: "Rina", Status: "AKTIF"},
		newSukarela: 1625,
	}
	svc := NewECAccountService(repo)

	msg, rupiah, newSukarela, err := svc.ConvertToSimpanan(context.Background(), 1, 125)
	if err != nil {
		t.Fatal(err)
	}
	if !repo.converted {
		t.Fatal("repository conversion harus dipanggil")
	}
	if repo.gotPoints != 125 || repo.gotRupiah != 125 {
		t.Fatalf("repo menerima points=%.0f rupiah=%.0f, mau 125 dan 125", repo.gotPoints, repo.gotRupiah)
	}
	if rupiah != 125 || newSukarela != 1625 {
		t.Fatalf("rupiah=%.0f newSukarela=%.0f, mau 125 dan 1625", rupiah, newSukarela)
	}
	if !strings.Contains(msg, "Rp 125") {
		t.Fatalf("pesan sukses tidak memuat nilai rupiah baku: %q", msg)
	}
}
