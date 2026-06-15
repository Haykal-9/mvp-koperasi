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
}
