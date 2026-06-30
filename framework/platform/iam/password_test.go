package iam

import (
	"strings"
	"testing"
)

func TestHashAndVerify_RoundTrip(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("expected $argon2id$ prefix, got %q", hash[:10])
	}
	if err := verifyPassword(hash, "correct-horse-battery-staple"); err != nil {
		t.Errorf("verify correct password: want nil, got %v", err)
	}
}

func TestVerify_WrongPassword(t *testing.T) {
	hash, _ := HashPassword("secret")
	if err := verifyPassword(hash, "wrong"); err == nil {
		t.Error("want error for wrong password, got nil")
	}
}

func TestVerify_UniqueHashes(t *testing.T) {
	h1, _ := HashPassword("same-password")
	h2, _ := HashPassword("same-password")
	if h1 == h2 {
		t.Error("two hashes of the same password must differ (random salt)")
	}
	// Both must still verify correctly.
	if err := verifyPassword(h1, "same-password"); err != nil {
		t.Errorf("h1 verify: %v", err)
	}
	if err := verifyPassword(h2, "same-password"); err != nil {
		t.Errorf("h2 verify: %v", err)
	}
}

func TestVerify_MalformedHash(t *testing.T) {
	if err := verifyPassword("not-a-valid-hash", "password"); err == nil {
		t.Error("want error for malformed hash, got nil")
	}
}
