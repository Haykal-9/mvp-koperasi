package service

import (
	"context"
	"errors"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/repository"
)

var ErrEmailAlreadyRegistered = errors.New("email already registered")

// AuthService owns koperasi account authentication and registration rules.
type AuthService struct {
	repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Authenticate(ctx context.Context, email, password string) (*model.User, error) {
	return s.repo.Authenticate(ctx, email, password)
}

func (s *AuthService) Register(ctx context.Context, nama, email, password string) (*model.User, error) {
	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailAlreadyRegistered
	}
	return s.repo.Create(ctx, nama, email, password)
}
