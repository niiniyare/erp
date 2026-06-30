package iam

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// Token TTLs per v2.1 spec.
const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour

	sessionKeyPrefix = "session:"
	tokenPlane       = "tnt"
)

// ErrSessionNotFound is returned when a token does not map to an active session.
var ErrSessionNotFound = errors.New("iam: session not found or expired")

// Session is the identity snapshot stored in Redis (hot path) and PostgreSQL
// (durable copy) at login time.
type Session struct {
	// ID is a stable internal identifier for audit and revocation.
	// It is NOT the token; clients never see this.
	ID uuid.UUID `json:"id"`

	UserID    uuid.UUID `json:"user_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	OrgUnitID uuid.UUID `json:"org_unit_id"` // uuid.Nil = tenant-wide viewer
	Plane     string    `json:"plane"`        // always "tenant" for this package

	// Roles holds the role slugs active at login.
	Roles []string `json:"roles"`

	// Permissions holds the baked permission snapshot computed at login.
	// Format: "module.resource.action" or wildcard "*.*.*".
	Permissions []string `json:"permissions"`

	IssuedAt        time.Time `json:"issued_at"`
	AccessExpiresAt time.Time `json:"access_expires_at"`
	// ExpiresAt aliases AccessExpiresAt for ViewerContext compatibility.
	ExpiresAt time.Time `json:"expires_at"`

	// AccessTokenHash and RefreshTokenHash are sha256(rawToken), stored for
	// revocation lookups. Never returned to clients.
	AccessTokenHash  string `json:"access_token_hash"`
	RefreshTokenHash string `json:"refresh_token_hash"`
}

// generateToken returns an opaque bearer token with the v2.1 format:
//
//	awosess_tnt_<hex(32-byte CSPRNG)>
//
// The hex encoding is used instead of base62 to avoid adding an external
// dependency — the token is equally random and opaque either way.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("iam: generate token: %w", err)
	}
	return "awosess_" + tokenPlane + "_" + hex.EncodeToString(b), nil
}

// tokenHash returns the hex-encoded SHA-256 hash of rawToken.
// This is used as the Redis key and Postgres lookup value so that neither
// store ever holds a usable raw credential.
func tokenHash(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}

// accessSessionKey returns the Redis key for an access token.
func accessSessionKey(rawToken string) string {
	return sessionKeyPrefix + tokenHash(rawToken)
}

// saveSession persists sess in Redis under rawAccessToken's hash.
// The caller is responsible for writing the Postgres record (see service.go).
func saveSession(ctx context.Context, r redis.Cmdable, rawAccessToken string, sess *Session) error {
	data, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("iam: marshal session: %w", err)
	}
	ttl := time.Until(sess.AccessExpiresAt)
	if ttl <= 0 {
		ttl = AccessTokenTTL
	}
	if err := r.Set(ctx, accessSessionKey(rawAccessToken), data, ttl).Err(); err != nil {
		return fmt.Errorf("iam: save session redis: %w", err)
	}
	return nil
}

// loadSession retrieves the session for rawAccessToken from Redis.
// Returns ErrSessionNotFound when absent or expired.
func loadSession(ctx context.Context, r redis.Cmdable, rawAccessToken string) (*Session, error) {
	data, err := r.Get(ctx, accessSessionKey(rawAccessToken)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("iam: load session: %w", err)
	}
	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("iam: unmarshal session: %w", err)
	}
	return &sess, nil
}

// deleteSession removes the access-token entry from Redis (logout / revocation).
func deleteSession(ctx context.Context, r redis.Cmdable, rawAccessToken string) error {
	return r.Del(ctx, accessSessionKey(rawAccessToken)).Err()
}

// pgSessionWriter is the subset of pgx.Tx or pgxpool.Conn needed to write sessions.
type pgSessionWriter interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// persistSession writes the session record to PostgreSQL.
// Call this after committing the login transaction (outside the auth TX so a
// session-table write failure does not roll back the login-time stat updates).
func persistSession(ctx context.Context, db pgSessionWriter, sess *Session) error {
	var orgUnitID interface{} = nil
	if sess.OrgUnitID != uuid.Nil {
		orgUnitID = sess.OrgUnitID
	}
	_, err := db.Exec(ctx, `
		INSERT INTO tenant_sessions (
			id, user_id, tenant_id, plane,
			access_token_hash, refresh_token_hash,
			permissions, roles, org_unit_id,
			issued_at, access_expires_at, refresh_expires_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		sess.ID, sess.UserID, sess.TenantID, sess.Plane,
		sess.AccessTokenHash, sess.RefreshTokenHash,
		sess.Permissions, sess.Roles, orgUnitID,
		sess.IssuedAt, sess.AccessExpiresAt,
		sess.IssuedAt.Add(RefreshTokenTTL),
	)
	if err != nil {
		return fmt.Errorf("iam: persist session: %w", err)
	}
	return nil
}

// revokeSessionInDB marks a session as revoked by access_token_hash.
func revokeSessionInDB(ctx context.Context, db pgSessionWriter, accessHash, reason string) error {
	_, err := db.Exec(ctx, `
		UPDATE tenant_sessions
		SET revoked_at = NOW(), revoke_reason = $2
		WHERE access_token_hash = $1 AND revoked_at IS NULL`,
		accessHash, reason,
	)
	return err
}
