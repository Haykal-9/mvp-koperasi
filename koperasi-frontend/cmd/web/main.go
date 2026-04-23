package main

import (
	"log"
	"koperasi-frontend/core/server"
)

func main() {
	r := server.SetupApp()
	addr := ":8080"
	log.Printf("Koperasi Digital running at http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
