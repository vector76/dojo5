package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("mysecretpassword")
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}

	if len(hash) == 0 {
		t.Fatal("expected non-empty hash")
	}

	if hash == "mysecretpassword" {
		t.Fatal("hash must not equal the plaintext password")
	}
}

func TestHashPasswordDifferentHashes(t *testing.T) {
	hash1, err := HashPassword("samepassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hash2, err := HashPassword("samepassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected different hashes for the same password (bcrypt uses random salt)")
	}
}

func TestCheckPasswordCorrect(t *testing.T) {
	password := "correctpassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := CheckPassword(hash, password); err != nil {
		t.Errorf("expected correct password to pass verification, got error: %v", err)
	}
}

func TestCheckPasswordWrong(t *testing.T) {
	hash, err := HashPassword("rightpassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := CheckPassword(hash, "wrongpassword"); err == nil {
		t.Error("expected wrong password to fail verification, but got nil error")
	}
}

func TestHashPasswordEmpty(t *testing.T) {
	hash, err := HashPassword("")
	if err != nil {
		t.Fatalf("unexpected error hashing empty password: %v", err)
	}

	if err := CheckPassword(hash, ""); err != nil {
		t.Errorf("expected empty password to verify against its own hash, got error: %v", err)
	}

	if err := CheckPassword(hash, "notempty"); err == nil {
		t.Error("expected non-empty password to fail against empty password hash")
	}
}
