package tests

import (
	"testing"

	"github.com/havilz/caelus-cloud/backend/pkg/hasher"
)

// TestArgon2HashingAndComparison memvalidasi fungsi pembuatan hash Argon2id dan kecocokan verifikasi kata sandi.
func TestArgon2HashingAndComparison(t *testing.T) {
	password := "SecureP@ssw0rd123!"

	hash, err := hasher.Hash(password, nil)
	if err != nil {
		t.Fatalf("gagal melakukan hash password: %v", err)
	}

	if hash == "" {
		t.Fatal("hasil hash tidak boleh kosong")
	}

	match, err := hasher.Compare(password, hash)
	if err != nil {
		t.Fatalf("gagal membandingkan password: %v", err)
	}
	if !match {
		t.Error("password valid harus menghasilkan perbandingan true")
	}

	wrongMatch, err := hasher.Compare("WrongP@ssw0rd!", hash)
	if err != nil {
		t.Fatalf("gagal membandingkan password yang salah: %v", err)
	}
	if wrongMatch {
		t.Error("password salah harus menghasilkan perbandingan false")
	}

	_, err = hasher.Compare(password, "invalid_hash_string")
	if err == nil {
		t.Error("hash tidak valid harus menghasilkan error")
	}
}
