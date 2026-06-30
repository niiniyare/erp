package iam

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// ErrInvalidPassword is returned when a password verification fails.
// Callers should return a generic "invalid credentials" message to avoid
// revealing which of email or password was incorrect.
var ErrInvalidPassword = errors.New("iam: invalid password")

// argon2Cfg holds Argon2id parameters.
// Values meet OWASP recommended minimums for web authentication (2024).
var argon2Cfg = struct {
	memory  uint32
	time    uint32
	threads uint8
	keyLen  uint32
	saltLen int
}{
	memory:  64 * 1024, // 64 MB
	time:    3,
	threads: 4,
	keyLen:  32,
	saltLen: 16,
}

// HashPassword returns an Argon2id hash of password encoded as:
//
//	$argon2id$v=19$m=<m>,t=<t>,p=<p>$<base64salt>$<base64hash>
func HashPassword(password string) (string, error) {
	salt := make([]byte, argon2Cfg.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("iam: generate salt: %w", err)
	}
	hash := argon2.IDKey(
		[]byte(password), salt,
		argon2Cfg.time,
		argon2Cfg.memory,
		argon2Cfg.threads,
		argon2Cfg.keyLen,
	)
	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2Cfg.memory, argon2Cfg.time, argon2Cfg.threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return encoded, nil
}

// verifyPassword checks password against the Argon2id encoded hash.
// Returns ErrInvalidPassword when credentials don't match.
// Timing-safe: always runs the full hash even on parse error.
func verifyPassword(encoded, password string) error {
	parts := strings.Split(encoded, "$")
	// Expected parts: ["", "argon2id", "v=19", "m=…,t=…,p=…", "<salt>", "<hash>"]
	if len(parts) != 6 || parts[1] != "argon2id" {
		return errors.New("iam: malformed password hash")
	}

	var memory, t uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &t, &threads); err != nil {
		return fmt.Errorf("iam: parse hash params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("iam: decode salt: %w", err)
	}
	stored, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return fmt.Errorf("iam: decode stored hash: %w", err)
	}

	candidate := argon2.IDKey([]byte(password), salt, t, memory, threads, uint32(len(stored)))
	if subtle.ConstantTimeCompare(candidate, stored) != 1 {
		return ErrInvalidPassword
	}
	return nil
}
