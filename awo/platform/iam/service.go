package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"awo.so/awo/auth"
	"awo.so/awo/runtime"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const (
	// humanSessionTTL is the default session lifetime for human users.
	humanSessionTTL = 24 * time.Hour
)

// AuthService provides the authentication and session management operations
// that the API middleware and IAM handlers depend on.
//
// It is the sole writer to:
//   - Redis session store ("session:{token}" keys)
//   - Redis user session index ("user_sessions:{tenantID}:{userID}")
//   - Redis API token cache ("iam:api_token:{hash}")
//   - iam_sessions table (audit trail)
//   - iam_login_audit table
//
// It is constructed once at startup and shared across all concurrent requests.
// All methods are goroutine-safe.
type AuthService struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

// ── Login ─────────────────────────────────────────────────────────────────────

// LoginInput carries the credentials and metadata for a login attempt.
type LoginInput struct {
	Email    string
	Password string
	TenantID uuid.UUID
	DeviceID string
	IPAddress string
}

// LoginResult carries the session token and session record returned after a
// successful login. The caller returns Token to the client as a Bearer credential.
type LoginResult struct {
	Token   string
	Session *auth.Session
}

// Login authenticates a human user and issues a new session.
//
// Flow:
//  1. Verify tenant is ACTIVE (set_tenant_context handles this via RLS).
//  2. Load the iam_user record by email within the tenant.
//  3. Verify the bcrypt password against the stored hash.
//  4. Load the user's active roles from iam_user_roles.
//  5. Generate a cryptographically random session token.
//  6. Store the session in Redis with TTL.
//  7. Add the token to the user_sessions:{tenantID}:{userID} sorted set.
//  8. Insert an iam_sessions audit record.
//  9. Insert an iam_login_audit event record.
// 10. Return the token and session to the caller.
//
// On credential failure, Login returns a generic error to prevent user enumeration.
// The specific failure reason is written to iam_login_audit.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	// Set tenant context so RLS scopes the subsequent queries to the correct tenant.
	if _, err := s.DB.Exec(ctx, "SELECT set_tenant_context($1)", input.TenantID); err != nil {
		return nil, fmt.Errorf("iam: login: set tenant context: %w", err)
	}

	// Load user by email within the tenant.
	var (
		userID       uuid.UUID
		passwordHash string
		status       string
	)
	err := s.DB.QueryRow(ctx, `
		SELECT id, password_hash, status
		FROM iam_users
		WHERE email = $1
	`, input.Email).Scan(&userID, &passwordHash, &status)
	if err != nil {
		s.writeFailedLoginAudit(ctx, input, uuid.Nil, "user_not_found")
		return nil, &runtime.BusinessError{
			Code:    "iam.login.invalid_credentials",
			Message: "Invalid email or password.",
			Status:  401,
		}
	}

	if status != "active" {
		s.writeFailedLoginAudit(ctx, input, userID, "account_"+status)
		return nil, &runtime.BusinessError{
			Code:    "iam.login.account_" + status,
			Message: "Your account is " + status + ". Contact your administrator.",
			Status:  403,
		}
	}

	// Verify password.
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)); err != nil {
		s.writeFailedLoginAudit(ctx, input, userID, "invalid_password")
		return nil, &runtime.BusinessError{
			Code:    "iam.login.invalid_credentials",
			Message: "Invalid email or password.",
			Status:  401,
		}
	}

	// Load roles.
	roles, err := s.loadUserRoles(ctx, input.TenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("iam: login: load roles: %w", err)
	}

	// Generate session token.
	token, err := auth.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("iam: login: generate token: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(humanSessionTTL)

	session := &auth.Session{
		Token:            token,
		UserID:           userID,
		ServiceAccountID: uuid.Nil,
		TenantID:         input.TenantID,
		Roles:            roles,
		ExpiresAt:        expiresAt,
		IssuedAt:         now,
		DeviceID:         input.DeviceID,
		IPAddress:        input.IPAddress,
	}

	// Store session in Redis.
	if err := s.storeSession(ctx, session); err != nil {
		return nil, fmt.Errorf("iam: login: store session: %w", err)
	}

	// Write iam_sessions audit record.
	if err := s.insertSessionRecord(ctx, session); err != nil {
		// Non-fatal: session is live in Redis. Log and continue.
		// TODO: emit metric for iam_sessions insert failure
		_ = err
	}

	// Write iam_login_audit event.
	if err := s.insertLoginAudit(ctx, input.TenantID, userID, uuid.Nil, "login", input.IPAddress, input.DeviceID, ""); err != nil {
		// Non-fatal: audit write failure must not block login.
		// TODO: emit metric for audit write failure
		_ = err
	}

	// Update last_login_at (best-effort, outside the critical path).
	_, _ = s.DB.Exec(ctx, `UPDATE iam_users SET last_login_at = $1 WHERE id = $2`, now, userID)

	return &LoginResult{Token: token, Session: session}, nil
}

// ── Logout ────────────────────────────────────────────────────────────────────

// Logout invalidates an active session identified by token.
//
// Flow:
//  1. DEL session:{token} from Redis.
//  2. ZREM token from user_sessions:{tenantID}:{userID} sorted set.
//  3. UPDATE iam_sessions SET revoked_at = now() WHERE token_hash = sha256(token).
//  4. INSERT iam_login_audit event.
func (s *AuthService) Logout(ctx context.Context, session *auth.Session) error {
	hash := tokenHash(session.Token)
	tenantStr := session.TenantID.String()
	userStr := session.UserID.String()
	indexKey := fmt.Sprintf("user_sessions:%s:%s", tenantStr, userStr)

	pipe := s.Redis.Pipeline()
	pipe.Del(ctx, session.RedisKey())
	pipe.ZRem(ctx, indexKey, session.Token)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("iam: logout: redis: %w", err)
	}

	// Mark session revoked in audit table.
	if _, err := s.DB.Exec(ctx, `
		UPDATE iam_sessions SET revoked_at = now()
		WHERE token_hash = $1 AND tenant_id = $2
	`, hash, session.TenantID); err != nil {
		// Non-fatal: Redis is the authoritative revocation.
		_ = err
	}

	_ = s.insertLoginAudit(ctx, session.TenantID, session.UserID, uuid.Nil, "logout", session.IPAddress, session.DeviceID, "")
	return nil
}

// ── RevokeUserSessions ────────────────────────────────────────────────────────

// RevokeUserSessions revokes ALL active sessions for the given user within the
// tenant. Called by the IAM admin and by [UserRoleChangeHook] after role changes.
//
// Session tokens are retrieved from the user_sessions:{tenantID}:{userID} sorted
// set and deleted from Redis. The audit table is updated best-effort.
func (s *AuthService) RevokeUserSessions(ctx context.Context, tenantID, userID uuid.UUID) error {
	indexKey := fmt.Sprintf("user_sessions:%s:%s", tenantID, userID)
	tokens, err := s.Redis.ZRange(ctx, indexKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("iam: revoke user sessions: read index: %w", err)
	}
	if len(tokens) == 0 {
		return nil
	}

	keys := make([]string, 0, len(tokens)+1)
	hashes := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		keys = append(keys, "session:"+tok)
		hashes = append(hashes, tokenHash(tok))
	}
	keys = append(keys, indexKey)

	if err := s.Redis.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("iam: revoke user sessions: redis DEL: %w", err)
	}

	// Best-effort: mark sessions revoked in audit table.
	_, _ = s.DB.Exec(ctx, `
		UPDATE iam_sessions SET revoked_at = now()
		WHERE token_hash = ANY($1) AND tenant_id = $2 AND revoked_at IS NULL
	`, hashes, tenantID)

	_ = s.insertLoginAudit(ctx, tenantID, userID, uuid.Nil, "session_revoked", "", "", "admin_revoke_all")
	return nil
}

// ── ValidateToken ─────────────────────────────────────────────────────────────

// ValidateToken looks up the session token in Redis, deserializes the Session,
// and validates expiry. Called on every authenticated request by the middleware.
//
// Returns a BusinessError with status 401 when:
//   - Token not found in Redis (expired or never issued).
//   - Session has passed its ExpiresAt wall-clock time.
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*auth.Session, error) {
	key := "session:" + token
	data, err := s.Redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, &runtime.BusinessError{
			Code:    "iam.session.not_found",
			Message: "Session not found or expired.",
			Status:  401,
		}
	}

	var session auth.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("iam: validate token: unmarshal session: %w", err)
	}

	if session.IsExpired(time.Now()) {
		return nil, &runtime.BusinessError{
			Code:    "iam.session.expired",
			Message: "Session has expired.",
			Status:  401,
		}
	}

	return &session, nil
}

// ── LoadRolePermissions ───────────────────────────────────────────────────────

// LoadRolePermissions returns all role-to-permission bindings from the global
// iam_role_permissions table. Called at startup by the application wiring to
// initialize [auth.CasbinEvaluator] Phase 2, and after any role-permission
// change to trigger a policy reload.
//
// iam_role_permissions is a global table (no tenant_id, no RLS). This method
// queries it directly via the DB pool without tenant context.
func (s *AuthService) LoadRolePermissions(ctx context.Context) ([]auth.RolePermission, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT role_name, permission_identifier
		FROM iam_role_permissions
		ORDER BY role_name, permission_identifier
	`)
	if err != nil {
		return nil, fmt.Errorf("iam: load role permissions: %w", err)
	}
	defer rows.Close()

	var result []auth.RolePermission
	for rows.Next() {
		var rp auth.RolePermission
		if err := rows.Scan(&rp.Role, &rp.Permission); err != nil {
			return nil, fmt.Errorf("iam: load role permissions: scan: %w", err)
		}
		result = append(result, rp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iam: load role permissions: rows: %w", err)
	}
	return result, nil
}

// ── Internal helpers ──────────────────────────────────────────────────────────

func (s *AuthService) loadUserRoles(ctx context.Context, tenantID, userID uuid.UUID) ([]string, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT role_name FROM iam_user_roles
		WHERE tenant_id = $1 AND user_id = $2
	`, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("query user roles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, r)
	}
	return roles, rows.Err()
}

func (s *AuthService) storeSession(ctx context.Context, session *auth.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	ttl := session.TTL(time.Now())
	if ttl <= 0 {
		return fmt.Errorf("session already expired")
	}

	// SET session:{token} {json} EX {ttl}
	if err := s.Redis.Set(ctx, session.RedisKey(), data, ttl).Err(); err != nil {
		return fmt.Errorf("redis SET session: %w", err)
	}

	// ZADD user_sessions:{tenantID}:{userID} {expiresAt.Unix()} {token}
	// Score = expiry unix timestamp; allows pruning expired members via ZRANGEBYSCORE.
	indexKey := fmt.Sprintf("user_sessions:%s:%s", session.TenantID, session.UserID)
	score := float64(session.ExpiresAt.Unix())
	if err := s.Redis.ZAdd(ctx, indexKey, &redis.Z{Score: score, Member: session.Token}).Err(); err != nil {
		// Non-fatal: the session key is live; the index is best-effort for bulk revoke.
		_ = err
	}
	return nil
}

func (s *AuthService) insertSessionRecord(ctx context.Context, session *auth.Session) error {
	hash := tokenHash(session.Token)
	_, err := s.DB.Exec(ctx, `
		INSERT INTO iam_sessions
			(id, tenant_id, token_hash, user_id, service_account_id,
			 issued_at, expires_at, device_id, ip_address)
		VALUES
			(gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8)
	`,
		session.TenantID,
		hash,
		nullIfNil(session.UserID),
		nullIfNil(session.ServiceAccountID),
		session.IssuedAt,
		session.ExpiresAt,
		session.DeviceID,
		session.IPAddress,
	)
	if err != nil {
		return fmt.Errorf("insert iam_session: %w", err)
	}
	return nil
}

func (s *AuthService) insertLoginAudit(ctx context.Context, tenantID, userID, serviceAccountID uuid.UUID, event, ipAddress, deviceID, failureReason string) error {
	_, err := s.DB.Exec(ctx, `
		INSERT INTO iam_login_audits
			(id, tenant_id, event, user_id, service_account_id,
			 ip_address, device_id, failure_reason, created_at)
		VALUES
			(gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, now())
	`,
		tenantID,
		event,
		nullIfNil(userID),
		nullIfNil(serviceAccountID),
		ipAddress,
		deviceID,
		failureReason,
	)
	if err != nil {
		return fmt.Errorf("insert iam_login_audit: %w", err)
	}
	return nil
}

func (s *AuthService) writeFailedLoginAudit(ctx context.Context, input LoginInput, userID uuid.UUID, reason string) {
	_ = s.insertLoginAudit(ctx, input.TenantID, userID, uuid.Nil, "failed_login", input.IPAddress, input.DeviceID, reason)
}

// nullIfNil returns nil for uuid.Nil, otherwise returns the UUID value.
// Used to store nullable UUID columns correctly in pgx.
func nullIfNil(id uuid.UUID) interface{} {
	if id == uuid.Nil {
		return nil
	}
	return id
}
