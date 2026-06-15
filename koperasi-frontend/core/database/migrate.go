package database

import (
	"errors"
	"log"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // driver "postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"koperasi-frontend/migrations"
)

// Migrate menjalankan seluruh migrasi `up` yang belum diterapkan (idempoten:
// golang-migrate mencatat versi di tabel schema_migrations).
//
// Migrasi dijalankan lewat SESSION pooler (port 5432), bukan transaction pooler
// (6543) yang dipakai aplikasi. Sebabnya: golang-migrate memakai advisory lock
// tingkat sesi yang tidak bertahan di transaction pooling (PgBouncer).
func Migrate(databaseURL string) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, migrationURL(databaseURL))
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	log.Println("migrations applied")
	return nil
}

// migrationURL menyesuaikan DSN aplikasi untuk koneksi migrasi:
//   - skema "postgresql://" -> "postgres://" (driver golang-migrate terdaftar "postgres")
//   - port transaction pooler 6543 -> session pooler 5432 (afinitas sesi untuk advisory lock)
func migrationURL(appURL string) string {
	u := strings.Replace(appURL, "postgresql://", "postgres://", 1)
	u = strings.Replace(u, ":6543/", ":5432/", 1)
	return u
}
