package main

import (
	"log"
	"os"

	"koperasi-frontend/core/config"
	"koperasi-frontend/core/database"
	"koperasi-frontend/core/server"

	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Koneksi DB saat startup. Kegagalan koneksi tidak menjalankan fallback mock;
	// handler berbasis DB akan mengembalikan status tidak tersedia.
	var db *gorm.DB
	if cfg.DatabaseURL != "" {
		// Fase 1: jalankan migrasi (skema + seed) otomatis saat startup.
		if err := database.Migrate(cfg.MigrationDatabaseURL); err != nil {
			log.Printf("peringatan: migrasi DB gagal: %v", err)
		}
		var err error
		if db, err = database.Connect(cfg.DatabaseURL); err != nil {
			log.Printf("peringatan: koneksi DB gagal: %v (fitur berbasis database tidak tersedia)", err)
		}
	} else {
		log.Println("DATABASE_URL kosong: fitur berbasis database tidak tersedia")
	}

	// For local dev, read templates/static from api/ sub-directory on disk.
	// For Vercel, api/index.go uses //go:embed instead.
	server.Assets = os.DirFS("api")
	r := server.SetupApp(db, server.AppOptions{
		SessionSecret: cfg.SessionSecret,
		SessionSecure: cfg.SessionSecure,
	})
	addr := ":" + cfg.Port
	log.Printf("Koperasi Digital running at http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
