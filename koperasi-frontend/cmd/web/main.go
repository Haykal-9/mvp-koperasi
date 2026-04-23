package main

import (
	"log"

	koperasifrontend "koperasi-frontend"
	"koperasi-frontend/core/server"
)

func main() {
	server.Assets = koperasifrontend.FS
	r := server.SetupApp()
	addr := ":8080"
	log.Printf("Koperasi Digital running at http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
