package api

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"koperasi-frontend/core/config"
	"koperasi-frontend/core/database"
	"koperasi-frontend/core/server"
)

//go:embed templates static
var embeddedAssets embed.FS

var app http.Handler

func init() {
	server.Assets, _ = fs.Sub(embeddedAssets, ".")
	cfg, err := config.Load()
	if err != nil {
		log.Printf("api: konfigurasi tidak valid: %v", err)
		app = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "Konfigurasi server tidak valid.", http.StatusServiceUnavailable)
		})
		return
	}

	opts := server.AppOptions{
		SessionSecret: cfg.SessionSecret,
		SessionSecure: cfg.SessionSecure,
	}
	if cfg.DatabaseURL == "" {
		log.Println("api: DATABASE_URL kosong, fitur berbasis database tidak tersedia")
		app = server.SetupApp(nil, opts)
		return
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Printf("api: koneksi database gagal: %v", err)
		app = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "Database sementara tidak tersedia.", http.StatusServiceUnavailable)
		})
		return
	}
	app = server.SetupApp(db, opts)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
