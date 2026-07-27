package audit

import (
	"testing"
)

func TestHashSessionToken_Deterministic(t *testing.T) {
	t.Parallel()

	token := "some-session-token"
	secret := "server-secret"

	h1 := HashSessionToken(token, secret)
	h2 := HashSessionToken(token, secret)
	if h1 != h2 {
		t.Error("HashSessionToken must be deterministic")
	}
}

func TestHashSessionToken_Length(t *testing.T) {
	t.Parallel()

	h := HashSessionToken("token", "secret")
	// HMAC-SHA256 = 32 bytes = 64 hex chars.
	if len(h) != 64 {
		t.Errorf("HashSessionToken: expected 64 hex chars, got %d", len(h))
	}
}

func TestHashSessionToken_SecretSensitivity(t *testing.T) {
	t.Parallel()

	token := "token"
	h1 := HashSessionToken(token, "secret-a")
	h2 := HashSessionToken(token, "secret-b")
	if h1 == h2 {
		t.Error("different secrets must produce different hashes")
	}
}

func TestHashSessionToken_TokenSensitivity(t *testing.T) {
	t.Parallel()

	secret := "shared-secret"
	h1 := HashSessionToken("token-a", secret)
	h2 := HashSessionToken("token-b", secret)
	if h1 == h2 {
		t.Error("different tokens must produce different hashes")
	}
}
