// Package migrations menyematkan (embed) berkas SQL migrasi ke dalam binary
// agar tetap satu artefak deploy (mis. Vercel) tanpa berkas eksternal.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
