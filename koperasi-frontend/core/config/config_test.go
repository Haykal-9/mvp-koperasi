package config

import "testing"

func TestLoadRequiresStrongSessionSecret(t *testing.T) {
	t.Setenv("SESSION_SECRET", "")

	if _, err := Load(); err == nil {
		t.Fatal("SESSION_SECRET kosong harus ditolak")
	}
}

func TestLoadSessionSecure(t *testing.T) {
	t.Setenv("SESSION_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("SESSION_SECURE", "true")
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://app-db")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.SessionSecure {
		t.Fatal("SESSION_SECURE=true tidak diterapkan")
	}
	if cfg.Port != "9090" {
		t.Fatalf("port = %q, ingin 9090", cfg.Port)
	}
	if cfg.MigrationDatabaseURL != "postgres://app-db" {
		t.Fatalf("migration url default = %q, ingin DATABASE_URL", cfg.MigrationDatabaseURL)
	}
}

func TestLoadMigrationDatabaseURLOverride(t *testing.T) {
	t.Setenv("SESSION_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("DATABASE_URL", "postgres://transaction-pooler")
	t.Setenv("MIGRATION_DATABASE_URL", "postgres://session-pooler")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MigrationDatabaseURL != "postgres://session-pooler" {
		t.Fatalf("migration url = %q, ingin override MIGRATION_DATABASE_URL", cfg.MigrationDatabaseURL)
	}
}
