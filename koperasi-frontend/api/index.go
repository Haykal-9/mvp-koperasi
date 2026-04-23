package api

import (
	"net/http"

	"koperasi-frontend/core/server"
)

var app *http.Handler

func init() {
	r := server.SetupApp()
	h := http.Handler(r)
	app = &h
}

func Handler(w http.ResponseWriter, r *http.Request) {
	(*app).ServeHTTP(w, r)
}
