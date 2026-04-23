package main

import (
	"log"
	"os"

	"koperasi-frontend/core/server"
)

func main() {
	// For local dev, read templates/static from api/ sub-directory on disk.
	// For Vercel, api/index.go uses //go:embed instead.
	server.Assets = os.DirFS("api")
	r := server.SetupApp()
	addr := ":8080"
	log.Printf("Koperasi Digital running at http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
