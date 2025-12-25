package auth

import (
	"testing"
)

func TestPasswordHashing(t *testing.T) {
	password := "mySecretPassword123"

	// 1. Test Hash Generation
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Fatal("HashPassword returned empty string")
	}

	if hash == password {
		t.Fatal("HashPassword returned the original password (not hashed)")
	}

	// 2. Test Verification (Correct Password)
	if !CheckPasswordHash(password, hash) {
		t.Errorf("CheckPasswordHash failed for correct password")
	}

	// 3. Test Verification (Incorrect Password)
	wrongPassword := "wrongPassword456"
	if CheckPasswordHash(wrongPassword, hash) {
		t.Errorf("CheckPasswordHash succeeded for wrong password")
	}
}
