// Package iam — service.go provides the IAM service layer for authentication
// and session management.
//
// The service never touches the database directly — all persistence goes
// through EntityRepository. Redis is used exclusively for session token
// caching. The source of truth is always PostgreSQL.
package iam

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"awo.so/awo/cache"
	"awo.so/awo/def"
	"awo.so/awo/driver"
	"awo.so/awo/filter"
	"awo.so/awo/runtime"
)

const (
	sessionTTL      = 24 * time.Hour
	sessionCacheKey = "session:"
	bcryptCost      = 12
	tokenBytes      = 32
)

// SessionClaims holds the authenticated principal extracted from a valid session.
type SessionClaims struct {
	SessionID uuid.UUID
	UserID    uuid.UUID
	TenantID  uuid.UUID
	Email     string
	Roles     []string
	ExpiresAt time.Time
}

// Service provides authentication, session, and user management operations.
type Service struct {
	users    driver.EntityRepository[*def.EntityRecord]
	sessions driver.EntityRepository[*def.EntityRecord]
	cache    cache.Cache
}

// NewService creates a Service with the given dependencies.
func NewService(
	users driver.EntityRepository[*def.EntityRecord],
	sessions driver.EntityRepository[*def.EntityRecord],
	c cache.Cache,
) *Service {
	return &Service{users: users, sessions: sessions, cache: c}
}

// Login authenticates a user by email and password.
// Returns a session token on success. The token is stored in Redis for fast
// validation and written to PostgreSQL as the durable record.
func (s *Service) Login(ctx context.Context, email, password string) (string, *SessionClaims, error) {
	// Find user by email.
	results, _, err := s.users.Query(ctx, filter.Eq("email", email), driver.WithSkipCount())
	if err != nil {
		return "", nil, fmt.Errorf("iam.Login: query user: %w", err)
	}
	if len(results) == 0 {
		// Return a generic error — do not reveal whether the email exists.
		return "", nil, &runtime.BusinessError{
			Code:    "iam.invalid_credentials",
			Message: "Invalid email or password",
			Status:  401,
		}
	}
	user := results[0]

	// Verify status.
	if user.GetString("status") != "active" {
		return "", nil, &runtime.BusinessError{
			Code:    "iam.account_not_active",
			Message: "Account is not active",
			Status:  403,
		}
	}

	// Verify password.
	hash := user.GetString("password_hash")
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return "", nil, &runtime.BusinessError{
			Code:    "iam.invalid_credentials",
			Message: "Invalid email or password",
			Status:  401,
		}
	}

	// Generate session token.
	token, tokenHash, err := generateToken()
	if err != nil {
		return "", nil, fmt.Errorf("iam.Login: generate token: %w", err)
	}

	expiresAt := time.Now().UTC().Add(sessionTTL)

	// Persist session.
	session, err := s.sessions.Create(ctx, driver.CreateInput{
		Data: map[string]any{
			"user_id":    user.ID,
			"token_hash": tokenHash,
			"expires_at": expiresAt,
		},
		Actor: nil, // system-created
	})
	if err != nil {
		return "", nil, fmt.Errorf("iam.Login: create session: %w", err)
	}

	claims := &SessionClaims{
		SessionID: session.ID,
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Email:     user.GetString("email"),
		ExpiresAt: expiresAt,
	}

	// Cache claims for fast validation.
	if err := s.cache.Set(ctx, sessionCacheKey+tokenHash, claims, sessionTTL); err != nil {
		// Cache failure is non-fatal here — session is in DB. Subsequent
		// requests will miss cache and re-validate from DB.
		// Log only; don't fail the login.
		_ = err
	}

	// Update last_login_at.
	_, _ = s.users.Update(ctx, user.ID, driver.UpdateInput{
		Data: map[string]any{"last_login_at": time.Now().UTC()},
	})

	return token, claims, nil
}

// ValidateToken validates a session token and returns its claims.
// Fast path: Redis cache lookup. Slow path: PostgreSQL lookup + cache fill.
func (s *Service) ValidateToken(ctx context.Context, token string) (*SessionClaims, error) {
	tokenHash := hashToken(token)

	// Fast path.
	var claims SessionClaims
	if err := s.cache.Get(ctx, sessionCacheKey+tokenHash, &claims); err == nil {
		if time.Now().After(claims.ExpiresAt) {
			_ = s.cache.Delete(ctx, sessionCacheKey+tokenHash)
			return nil, expiredErr()
		}
		return &claims, nil
	}

	// Slow path: DB lookup.
	results, _, err := s.sessions.Query(ctx, filter.And(
		filter.Eq("token_hash", tokenHash),
		filter.Eq("revoked", false),
	), driver.WithSkipCount())
	if err != nil {
		return nil, fmt.Errorf("iam.ValidateToken: query: %w", err)
	}
	if len(results) == 0 {
		return nil, &runtime.BusinessError{
			Code:    "iam.invalid_token",
			Message: "Session token is invalid or expired",
			Status:  401,
		}
	}
	session := results[0]

	expiresAt, _ := session.Get("expires_at").(time.Time)
	if time.Now().After(expiresAt) {
		return nil, expiredErr()
	}

	userID, _ := session.Get("user_id").(uuid.UUID)
	user, err := s.users.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("iam.ValidateToken: get user: %w", err)
	}

	claims = SessionClaims{
		SessionID: session.ID,
		UserID:    userID,
		TenantID:  user.TenantID,
		Email:     user.GetString("email"),
		ExpiresAt: expiresAt,
	}

	// Re-populate cache.
	_ = s.cache.Set(ctx, sessionCacheKey+tokenHash, claims, time.Until(expiresAt))

	return &claims, nil
}

// Logout revokes the session for the given token.
func (s *Service) Logout(ctx context.Context, token string) error {
	tokenHash := hashToken(token)

	_ = s.cache.Delete(ctx, sessionCacheKey+tokenHash)

	results, _, err := s.sessions.Query(ctx, filter.Eq("token_hash", tokenHash), driver.WithSkipCount())
	if err != nil {
		return fmt.Errorf("iam.Logout: query: %w", err)
	}
	if len(results) == 0 {
		return nil // already gone
	}
	_, err = s.sessions.Update(ctx, results[0].ID, driver.UpdateInput{
		Data: map[string]any{
			"revoked":    true,
			"revoked_at": time.Now().UTC(),
		},
	})
	return err
}

// HashPassword hashes a plaintext password using bcrypt.
// Module authors call this before passing password_hash to Create.
func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(h), nil
}

// generateToken creates a cryptographically random token and its SHA-256 hash.
// The token is returned to the client; only the hash is stored in the database.
func generateToken() (token, hash string, err error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token = hex.EncodeToString(b)
	hash = hashToken(token)
	return token, hash, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func expiredErr() error {
	return &runtime.BusinessError{
		Code:    "iam.session_expired",
		Message: "Session has expired — please log in again",
		Status:  401,
	}
}
