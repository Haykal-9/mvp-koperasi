// Package database membuka koneksi PostgreSQL via GORM, mengatur connection pool,
// dan memverifikasi koneksi (ping). Dipakai sebagai fondasi lapisan repository.
package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Connect membuka koneksi ke PostgreSQL menggunakan DSN (DATABASE_URL),
// mengatur parameter pool, lalu melakukan ping. Mengembalikan *gorm.DB siap pakai.
func Connect(databaseURL string) (*gorm.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database: DATABASE_URL kosong")
	}

	// PreferSimpleProtocol menonaktifkan prepared statement (extended protocol).
	// Wajib untuk Supabase transaction pooler (PgBouncer, port 6543) yang tidak
	// mendukung prepared statement lintas-koneksi.
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  databaseURL,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("database: gagal membuka koneksi: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database: gagal mengambil *sql.DB: %w", err)
	}
	// Connection pooling (NFR-04). Batasi koneksi agar aman untuk pooler serverless.
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database: ping gagal: %w", err)
	}

	log.Println("DB connected")
	return db, nil
}
