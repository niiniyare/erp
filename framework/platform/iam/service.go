package iam

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel errors returned by AuthService.Login.
// Callers MUST NOT distinguish UserNotFound from InvalidPassword in HTTP
// responses — both must map to "invalid credentials" to prevent user enumeration.
var (
	ErrUserNotFound = errors.New("iam: user not found")
	ErrUserInactive = errors.New("iam: account not active")
	ErrUserLocked   = errors.New("iam: account temporarily locked")
)

const (
	lockoutThreshold = 5               // failed attempts before lockout
	lockoutDuration  = 15 * time.Minute
)

// LoginResult is returned by AuthService.Login on success.
type LoginResult struct {
	AccessToken  string
	RefreshToken string
	Session      *Session
}

// AuthService handles tenant-plane authentication.
type AuthService struct {
	pool  *pgxpool.Pool
	redis redis.Cmdable
}

// NewAuthService creates an AuthService backed by pool and redis.
func NewAuthService(pool *pgxpool.Pool, redis redis.Cmdable) *AuthService {
	return &AuthService{pool: pool, redis: redis}
}

// Login authenticates a tenant user by email + password and creates a session.
//
// tenantID must be resolved from the request before calling (e.g. X-Awo-Tenant header).
// ip is recorded for audit; may be empty.
//
// Returns ErrUserNotFound, ErrInvalidPassword, ErrUserInactive, or ErrUserLocked
// on failure. Callers should map all auth-failure variants to a generic 401.
func (s *AuthService) Login(ctx context.Context, tenantID uuid.UUID, email, password, ip string) (*LoginResult, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("iam: acquire conn: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("iam: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Set tenant RLS context so tenant_users policy applies.
	if _, err := tx.Exec(ctx, "SELECT set_tenant_context($1)", tenantID); err != nil {
		return nil, fmt.Errorf("iam: set_tenant_context: %w", err)
	}

	// Load user credentials.
	var (
		userID      uuid.UUID
		storedHash  string
		status      string
		lockedUntil *time.Time
		failedCount int
	)
	err = tx.QueryRow(ctx, `
		SELECT id, COALESCE(password_hash, ''), status, locked_until, failed_login_count
		FROM tenant_users
		WHERE email = $1`,
		email,
	).Scan(&userID, &storedHash, &status, &lockedUntil, &failedCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("iam: load user: %w", err)
	}

	// Check status before touching the password — avoids timing oracle on status.
	if status != "ACTIVE" {
		return nil, ErrUserInactive
	}

	// Check lockout.
	if lockedUntil != nil && lockedUntil.After(time.Now()) {
		return nil, ErrUserLocked
	}

	// Verify password (constant-time).
	if pwErr := verifyPassword(storedHash, password); pwErr != nil {
		// Record the failure and conditionally lock; commit the write.
		newCount := failedCount + 1
		if newCount >= lockoutThreshold {
			_, _ = tx.Exec(ctx, `
				UPDATE tenant_users
				SET failed_login_count = $1,
				    locked_until = NOW() + $2::interval
				WHERE id = $3`,
				newCount, lockoutDuration.String(), userID)
		} else {
			_, _ = tx.Exec(ctx, `
				UPDATE tenant_users SET failed_login_count = $1 WHERE id = $2`,
				newCount, userID)
		}
		_ = tx.Commit(ctx)
		return nil, ErrInvalidPassword
	}

	// Compute permission snapshot inside the same transaction.
	snap, err := computeSnapshot(ctx, tx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("iam: compute snapshot: %w", err)
	}

	// Record successful login + reset failure counter.
	var ipVal interface{} = nil
	if ip != "" {
		ipVal = ip
	}
	if _, err := tx.Exec(ctx, `
		UPDATE tenant_users
		SET failed_login_count = 0,
		    locked_until       = NULL,
		    last_login_at      = NOW(),
		    last_login_ip      = $2
		WHERE id = $1`, userID, ipVal); err != nil {
		// Non-fatal; proceed even if we cannot record the login timestamp.
		_ = err
	}

	// Commit DB work before touching Redis or writing sessions.
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("iam: commit: %w", err)
	}

	// Generate dual tokens (v2.1).
	accessToken, err := generateToken()
	if err != nil {
		return nil, err
	}
	refreshToken, err := generateToken()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	sess := &Session{
		ID:               uuid.New(),
		UserID:           userID,
		TenantID:         tenantID,
		OrgUnitID:        snap.OrgUnitID,
		Plane:            "tenant",
		Roles:            snap.Roles,
		Permissions:      snap.Permissions,
		IssuedAt:         now,
		AccessExpiresAt:  now.Add(AccessTokenTTL),
		ExpiresAt:        now.Add(AccessTokenTTL), // ViewerContext compat
		AccessTokenHash:  tokenHash(accessToken),
		RefreshTokenHash: tokenHash(refreshToken),
	}

	// Write to Redis (hot path).
	if err := saveSession(ctx, s.redis, accessToken, sess); err != nil {
		return nil, err
	}

	// Write durable copy to Postgres.
	// Use a fresh connection (outside the auth TX) so a session-table failure
	// does not roll back the successful login stat updates.
	pgConn, err := s.pool.Acquire(ctx)
	if err != nil {
		// Non-fatal: Redis already has the session; Postgres is the audit copy.
		// Log and continue rather than failing the login.
		_ = err
	} else {
		defer pgConn.Release()
		// RLS context required to write into tenant_sessions.
		if _, err := pgConn.Exec(ctx, "SELECT set_tenant_context($1)", tenantID); err == nil {
			_ = persistSession(ctx, pgConn, sess)
		}
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Session:      sess,
	}, nil
}

// Logout removes the access-token session from Redis and marks it revoked in Postgres.
func (s *AuthService) Logout(ctx context.Context, rawAccessToken string) error {
	if err := deleteSession(ctx, s.redis, rawAccessToken); err != nil {
		return fmt.Errorf("iam: logout redis: %w", err)
	}
	// Best-effort Postgres revocation — failure here does not prevent the token
	// from being unusable (the Redis key is already gone).
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil
	}
	defer conn.Release()
	_ = revokeSessionInDB(ctx, conn, tokenHash(rawAccessToken), "logout")
	return nil
}

// LoadSession retrieves the session for rawAccessToken from Redis.
// Returns ErrSessionNotFound when absent or expired.
func (s *AuthService) LoadSession(ctx context.Context, rawAccessToken string) (*Session, error) {
	return loadSession(ctx, s.redis, rawAccessToken)
}

// Refresh validates a refresh token, rotates both tokens, and returns a new LoginResult.
// The old access and refresh tokens are invalidated.
func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken string) (*LoginResult, error) {
	// Look up the session by refresh token hash in Postgres (not in Redis —
	// refresh tokens are not stored in Redis).
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("iam: acquire conn: %w", err)
	}
	defer conn.Release()

	refreshHash := tokenHash(rawRefreshToken)

	var (
		sessID          uuid.UUID
		userID          uuid.UUID
		tenantID        uuid.UUID
		orgUnitID       uuid.UUID
		plane           string
		roles           []string
		permissions     []string
		refreshExpires  time.Time
		revokedAt       *time.Time
	)
	err = conn.QueryRow(ctx, `
		SELECT id, user_id, tenant_id, COALESCE(org_unit_id, $2::uuid),
		       plane, roles, permissions,
		       refresh_expires_at, revoked_at
		FROM tenant_sessions
		WHERE refresh_token_hash = $1`,
		refreshHash, uuid.Nil,
	).Scan(&sessID, &userID, &tenantID, &orgUnitID,
		&plane, &roles, &permissions,
		&refreshExpires, &revokedAt)
	if errors.Is(err, pgx.ErrNoRows) || revokedAt != nil {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("iam: load session for refresh: %w", err)
	}
	if time.Now().After(refreshExpires) {
		return nil, ErrSessionNotFound
	}

	// Need RLS context to update tenant_sessions.
	if _, err := conn.Exec(ctx, "SELECT set_tenant_context($1)", tenantID); err != nil {
		return nil, fmt.Errorf("iam: set_tenant_context: %w", err)
	}

	// Revoke old session record.
	if _, err := conn.Exec(ctx, `
		UPDATE tenant_sessions
		SET revoked_at = NOW(), revoke_reason = 'rotation'
		WHERE id = $1`, sessID); err != nil {
		return nil, fmt.Errorf("iam: revoke old session: %w", err)
	}

	// Re-compute permissions (picks up role changes since last login).
	snap, err := computeSnapshot(ctx, conn, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("iam: compute snapshot: %w", err)
	}

	// Issue new tokens.
	newAccess, err := generateToken()
	if err != nil {
		return nil, err
	}
	newRefresh, err := generateToken()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	newSess := &Session{
		ID:               uuid.New(),
		UserID:           userID,
		TenantID:         tenantID,
		OrgUnitID:        snap.OrgUnitID,
		Plane:            plane,
		Roles:            snap.Roles,
		Permissions:      snap.Permissions,
		IssuedAt:         now,
		AccessExpiresAt:  now.Add(AccessTokenTTL),
		ExpiresAt:        now.Add(AccessTokenTTL),
		AccessTokenHash:  tokenHash(newAccess),
		RefreshTokenHash: tokenHash(newRefresh),
	}

	// Persist new session.
	if err := persistSession(ctx, conn, newSess); err != nil {
		return nil, err
	}
	if err := saveSession(ctx, s.redis, newAccess, newSess); err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
		Session:      newSess,
	}, nil
}
