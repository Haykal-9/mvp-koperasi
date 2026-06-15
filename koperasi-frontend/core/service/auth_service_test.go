package service

import (
	"context"
	"errors"
	"testing"

	"koperasi-frontend/core/model"
)

type fakeUserRepository struct {
	exists  bool
	created *model.User
}

func (f *fakeUserRepository) Authenticate(context.Context, string, string) (*model.User, error) {
	return nil, nil
}

func (f *fakeUserRepository) EmailExists(context.Context, string) (bool, error) {
	return f.exists, nil
}

func (f *fakeUserRepository) Create(_ context.Context, nama, email, _ string) (*model.User, error) {
	f.created = &model.User{ID: 10, Email: email, Role: "ANGGOTA", Nama: nama}
	return f.created, nil
}

func TestAuthServiceRegisterRejectsDuplicateEmail(t *testing.T) {
	svc := NewAuthService(&fakeUserRepository{exists: true})

	if _, err := svc.Register(context.Background(), "Nama", "user@example.com", "secret"); !errors.Is(err, ErrEmailAlreadyRegistered) {
		t.Fatalf("error = %v, ingin ErrEmailAlreadyRegistered", err)
	}
}

func TestAuthServiceRegisterCreatesAnggota(t *testing.T) {
	repo := &fakeUserRepository{}
	svc := NewAuthService(repo)

	user, err := svc.Register(context.Background(), "Nama", "user@example.com", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if user.Role != "ANGGOTA" || repo.created == nil {
		t.Fatalf("user tidak dibuat sebagai ANGGOTA: %#v", user)
	}
}
