package service

// TOTP and AES-256-GCM helpers for MFA.
//
// TOTP implementation follows RFC 6238 (TOTP) and RFC 4226 (HOTP).
// All cryptographic operations use stdlib only — no external TOTP library.
//
// Secret lifecycle:
//   1. InitiateMFA generates a random 20-byte secret → base32-encode → encrypt → cache.
//   2. ConfirmMFA decrypts from cache → verifies first TOTP code → saves encrypted to DB.
//   3. ValidateMFACode decrypts from DB → verifies code → replay-prevents via cache.
//   4. DisableMFA clears encrypted secret from DB after password re-verification.

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1" //nolint:gosec // TOTP RFC 4226 specifies HMAC-SHA1
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"time"
)

const (
	totpDigits = 6
	totpPeriod = 30 // seconds
)

// generateTOTPSecret creates a cryptographically random 20-byte TOTP secret
// and returns it as a base32-encoded string (no padding) suitable for use
// with RFC 4226 authenticator apps (Google Authenticator, Authy, etc.).
func generateTOTPSecret() (string, error) {
	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("mfa: generate secret: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw), nil
}

// buildTOTPURI returns the otpauth:// URI used to provision authenticator apps.
//
//	otpauth://totp/{issuer}:{account}?secret={secret}&issuer={issuer}&algorithm=SHA1&digits=6&period=30
func buildTOTPURI(issuer, account, base32Secret string) string {
	return fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d",
		issuer, account, base32Secret, issuer, totpDigits, totpPeriod,
	)
}

// verifyTOTP checks the given 6-digit code against the base32-encoded secret
// using a ±window tolerance (window=1 means T-1, T, T+1 are accepted).
// Returns the matched window index (0 = current, ±1 = adjacent) and true on
// success.  The window index is used by the replay-prevention layer.
func verifyTOTP(base32Secret, code string, window int) (int64, bool) {
	raw, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(base32Secret)
	if err != nil {
		// Try with padding in case secret was stored with it.
		raw, err = base32.StdEncoding.DecodeString(base32Secret)
		if err != nil {
			return 0, false
		}
	}

	now := time.Now().Unix()
	T := now / totpPeriod

	for i := -int64(window); i <= int64(window); i++ {
		if hotp(raw, T+i) == code {
			return T + i, true
		}
	}
	return 0, false
}

// hotp computes a 6-digit HOTP code (RFC 4226) for the given key and counter.
func hotp(key []byte, counter int64) string {
	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, uint64(counter)) //nolint:gosec // counter is non-negative by construction

	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(msg)
	h := mac.Sum(nil) // 20 bytes

	offset := h[len(h)-1] & 0x0F
	code := (uint32(h[offset])&0x7F)<<24 |
		uint32(h[offset+1])<<16 |
		uint32(h[offset+2])<<8 |
		uint32(h[offset+3])
	code %= 1_000_000

	return fmt.Sprintf("%06d", code)
}

// encryptSecret encrypts a plaintext string with AES-256-GCM using the
// provided 32-byte key and returns a base64url-encoded "nonce+ciphertext".
func encryptSecret(key []byte, plaintext string) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("mfa: encrypt: key must be 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("mfa: encrypt: new cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("mfa: encrypt: new gcm: %w", err)
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", fmt.Errorf("mfa: encrypt: nonce: %w", err)
	}
	// Seal appends the ciphertext + tag after the nonce.
	ciphertext := aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// decryptSecret reverses encryptSecret.
func decryptSecret(key []byte, encoded string) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("mfa: decrypt: key must be 32 bytes, got %d", len(key))
	}
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("mfa: decrypt: base64: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("mfa: decrypt: new cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("mfa: decrypt: new gcm: %w", err)
	}
	ns := aead.NonceSize()
	if len(data) < ns {
		return "", fmt.Errorf("mfa: decrypt: ciphertext too short")
	}
	nonce, ciphertext := data[:ns], data[ns:]
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("mfa: decrypt: open: %w", err)
	}
	return string(plaintext), nil
}
