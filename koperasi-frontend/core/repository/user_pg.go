package repository

import (
	"context"

	"gorm.io/gorm"

	"koperasi-frontend/core/model"
	"koperasi-frontend/core/security"
)

type pgUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &pgUserRepository{db: db}
}

const userCols = `id, email, role, nama`

func (r *pgUserRepository) Authenticate(ctx context.Context, email, password string) (*model.User, error) {
	type authRow struct {
		model.User
		PasswordHash string
	}
	rows := []authRow{}
	if err := r.db.WithContext(ctx).Raw(
		"SELECT "+userCols+", password_hash FROM users WHERE lower(email) = lower(?) LIMIT 1",
		email,
	).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 || !security.VerifyPassword(rows[0].PasswordHash, password) {
		return nil, nil
	}
	user := rows[0].User
	return &user, nil
}

func (r *pgUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Raw(
		"SELECT count(*) FROM users WHERE lower(email) = lower(?)", email,
	).Scan(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *pgUserRepository) Create(ctx context.Context, nama, email, password string) (*model.User, error) {
	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{Email: email, Role: "ANGGOTA", Nama: nama}
	if err := r.db.WithContext(ctx).Raw(
		`INSERT INTO users (email, password_hash, role, nama)
		 VALUES (?, ?, 'ANGGOTA', ?) RETURNING id`,
		email, hash, nama,
	).Scan(&user.ID).Error; err != nil {
		return nil, err
	}
	return user, nil
}
