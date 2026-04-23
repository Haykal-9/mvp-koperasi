package api

import (
	"net/http"

	koperasifrontend "koperasi-frontend"
	"koperasi-frontend/core/server"
)

var app http.Handler

func init() {
	server.Assets = koperasifrontend.FS
	app = server.SetupApp()
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
