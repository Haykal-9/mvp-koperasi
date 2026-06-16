// Package config memuat konfigurasi aplikasi dari variabel lingkungan (ENV).
// Saat pengembangan, berkas .env (bila ada) dimuat otomatis via godotenv.
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config menampung seluruh konfigurasi runtime yang dibaca dari ENV.
type Config struct {
	// DatabaseURL adalah DSN PostgreSQL (mis. dari Supabase/Neon).
	// Kosong = fitur yang memerlukan basis data tidak tersedia.
	DatabaseURL string
	// MigrationDatabaseURL adalah DSN PostgreSQL untuk menjalankan migrasi.
	// Biasanya direct/session pooler; default ke DatabaseURL bila kosong.
	MigrationDatabaseURL string
	// SessionSecret menggantikan secret hardcode lama (FR-SEC-02).
	SessionSecret string
	// SessionSecure menambahkan atribut Secure pada cookie sesi saat HTTPS.
	SessionSecure bool
	// Port HTTP yang didengarkan server.
	Port string
}

// Load membaca konfigurasi dari ENV. Berkas .env dimuat lebih dulu bila tersedia
// (khusus dev); di produksi ENV proses yang dipakai.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("config: .env tidak ditemukan, membaca dari environment proses")
	}
	secret := strings.TrimSpace(os.Getenv("SESSION_SECRET"))
	if len(secret) < 32 {
		return nil, fmt.Errorf("config: SESSION_SECRET wajib diisi minimal 32 karakter")
	}
	secure, err := envBool("SESSION_SECURE", os.Getenv("VERCEL") != "")
	if err != nil {
		return nil, err
	}
	databaseURL := os.Getenv("DATABASE_URL")
	return &Config{
		DatabaseURL:          databaseURL,
		MigrationDatabaseURL: getenv("MIGRATION_DATABASE_URL", databaseURL),
		SessionSecret:        secret,
		SessionSecure:        secure,
		Port:                 getenv("PORT", "8080"),
	}, nil
}

// getenv mengembalikan nilai ENV bila terisi, selain itu nilai default.
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("config: %s harus berupa true atau false", key)
	}
	return value, nil
}
