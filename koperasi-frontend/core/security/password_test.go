package security

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("password-rahasia")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "password-rahasia" {
		t.Fatal("password tersimpan sebagai plain-text")
	}
	if !VerifyPassword(hash, "password-rahasia") {
		t.Fatal("password yang benar ditolak")
	}
	if VerifyPassword(hash, "password-salah") {
		t.Fatal("password yang salah diterima")
	}
}
