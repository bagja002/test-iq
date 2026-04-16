package auth

import "testing"

func TestHashPasswordAndVerify(t *testing.T) {
	hash, err := HashPassword("super-secret")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if !VerifyPassword(hash, "super-secret") {
		t.Fatal("expected password verification to succeed")
	}

	if VerifyPassword(hash, "wrong-password") {
		t.Fatal("expected password verification to fail")
	}
}
