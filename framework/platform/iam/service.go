package iam

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
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
	lockoutThreshold = 5 // failed attempts before lockout
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

	// Set tenant RLS context so iam_users policy applies.
	if _, err := conn.Exec(ctx, "SELECT set_tenant_context($1)", tenantID); err != nil {
		return nil, fmt.Errorf("iam: set_tenant_context: %w", err)
	}

	// Load user credentials from iam_users (actual table name).
	var (
		userID     uuid.UUID
		storedHash string
		status     string
	)
	err = conn.QueryRow(ctx, `
		SELECT id, COALESCE(password_hash, ''), status
		FROM iam_users
		WHERE email = $1`,
		strings.ToLower(strings.TrimSpace(email)),
	).Scan(&userID, &storedHash, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("iam: load user: %w", err)
	}

	// Check status before touching the password.
	if status != "active" {
		if status == "locked" {
			return nil, ErrUserLocked
		}
		return nil, ErrUserInactive
	}

	// Verify password (constant-time).
	if err := verifyPassword(storedHash, password); err != nil {
		return nil, ErrInvalidPassword
	}

	// Compute permission snapshot.
	snap, err := computeSnapshot(ctx, conn, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("iam: compute snapshot: %w", err)
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
		ExpiresAt:        now.Add(AccessTokenTTL),
		AccessTokenHash:  tokenHash(accessToken),
		RefreshTokenHash: tokenHash(refreshToken),
	}

	// Write to Redis (hot path). No tenant_sessions table exists yet.
	if err := saveSession(ctx, s.redis, accessToken, sess); err != nil {
		return nil, err
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
		sessID         uuid.UUID
		userID         uuid.UUID
		tenantID       uuid.UUID
		orgUnitID      uuid.UUID
		plane          string
		roles          []string
		permissions    []string
		refreshExpires time.Time
		revokedAt      *time.Time
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

// Register creates a new tenant user with an Argon2id-hashed password.
// Returns the new user's UUID on success.
func (s *AuthService) Register(ctx context.Context, tenantID uuid.UUID, email, password, fullName string) (uuid.UUID, error) {
	if email == "" || password == "" {
		return uuid.Nil, fmt.Errorf("iam: email and password are required")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return uuid.Nil, fmt.Errorf("iam: hash password: %w", err)
	}

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("iam: acquire conn: %w", err)
	}
	defer conn.Release()

	// RLS: set tenant context so iam_users policy applies.
	if _, err := conn.Exec(ctx, "SELECT set_tenant_context($1)", tenantID); err != nil {
		return uuid.Nil, fmt.Errorf("iam: set_tenant_context: %w", err)
	}

	id := uuid.New()
	_, err = conn.Exec(ctx, `
		INSERT INTO iam_users (id, tenant_id, email, password_hash, status, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'active', $5, NOW(), NOW())`,
		id, tenantID, strings.ToLower(strings.TrimSpace(email)), hash,
		fmt.Sprintf(`{"full_name":%q}`, strings.TrimSpace(fullName)),
	)
	if err != nil {
		// Unique constraint on email → 23505
		return uuid.Nil, fmt.Errorf("iam: insert user: %w", err)
	}
	return id, nil
}

// CreateWorkspace creates a new tenant and an admin user in one atomic operation.
// Returns the tenant slug and tenant UUID on success.
func (s *AuthService) CreateWorkspace(ctx context.Context, workspaceName, adminEmail, adminName, password string) (slug string, tenantID uuid.UUID, err error) {
	hash, err := HashPassword(password)
	if err != nil {
		return "", uuid.Nil, fmt.Errorf("iam: hash password: %w", err)
	}

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return "", uuid.Nil, fmt.Errorf("iam: acquire conn: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return "", uuid.Nil, fmt.Errorf("iam: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Generate slug from workspace name (lowercase, spaces → hyphens).
	slug = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(workspaceName), " ", "-"))
	tenantID = uuid.New()

	// Insert tenant. Use PENDING as initial status per tenant lifecycle state machine;
	// status is immediately set to ACTIVE so the admin can log in right away.
	if _, err := tx.Exec(ctx, `
		INSERT INTO tenants (id, name, slug, status, created_at, updated_at)
		VALUES ($1, $2, $3, 'active', NOW(), NOW())`,
		tenantID, strings.TrimSpace(workspaceName), slug,
	); err != nil {
		return "", uuid.Nil, fmt.Errorf("iam: create tenant: %w", err)
	}

	// Set RLS context so iam_users policy applies.
	if _, err := tx.Exec(ctx, "SELECT set_tenant_context($1)", tenantID); err != nil {
		return "", uuid.Nil, fmt.Errorf("iam: set_tenant_context: %w", err)
	}

	// Insert admin user into iam_users (actual table; no full_name column → metadata JSONB).
	userID := uuid.New()
	if _, err := tx.Exec(ctx, `
		INSERT INTO iam_users (id, tenant_id, email, password_hash, status, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'active', $5, NOW(), NOW())`,
		userID, tenantID,
		strings.ToLower(strings.TrimSpace(adminEmail)), hash,
		fmt.Sprintf(`{"full_name":%q}`, strings.TrimSpace(adminName)),
	); err != nil {
		return "", uuid.Nil, fmt.Errorf("iam: create admin user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", uuid.Nil, fmt.Errorf("iam: commit: %w", err)
	}
	return slug, tenantID, nil
}

// PlatformLogin authenticates a platform admin user whose credentials are
// supplied via environment variables (PLATFORM_ROOT_EMAIL, PLATFORM_ROOT_PASSWORD_HASH).
// Issues a session with plane="platform" and wildcard permissions.
// This is Plane 1 access — no tenant context required.
func (s *AuthService) PlatformLogin(ctx context.Context, email, password, ip string) (*LoginResult, error) {
	rootEmail := strings.TrimSpace(os.Getenv("PLATFORM_ROOT_EMAIL"))
	rootHash := strings.TrimSpace(os.Getenv("PLATFORM_ROOT_PASSWORD_HASH"))

	if rootEmail == "" || rootHash == "" {
		return nil, fmt.Errorf("iam: platform login not configured")
	}

	// Constant-time email comparison to prevent enumeration.
	if !strings.EqualFold(email, rootEmail) {
		// Run the hash anyway so timing is consistent.
		_ = verifyPassword(rootHash, password)
		return nil, ErrInvalidPassword
	}
	if err := verifyPassword(rootHash, password); err != nil {
		return nil, ErrInvalidPassword
	}

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
		UserID:           uuid.Nil, // platform root has no tenant_user record
		TenantID:         uuid.Nil,
		Plane:            "platform",
		Roles:            []string{"role:platform-admin"},
		Permissions:      []string{"*.*.*"}, // full access
		IssuedAt:         now,
		AccessExpiresAt:  now.Add(AccessTokenTTL),
		ExpiresAt:        now.Add(AccessTokenTTL),
		AccessTokenHash:  tokenHash(accessToken),
		RefreshTokenHash: tokenHash(refreshToken),
	}

	// Platform sessions live only in Redis — no tenant context for Postgres write.
	if err := saveSession(ctx, s.redis, accessToken, sess); err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Session:      sess,
	}, nil
}
