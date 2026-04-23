package api

import (
	"embed"
	"io/fs"
	"net/http"

	"koperasi-frontend/core/server"
)

//go:embed templates static
var embeddedAssets embed.FS

var app http.Handler

func init() {
	server.Assets, _ = fs.Sub(embeddedAssets, ".")
	app = server.SetupApp()
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
