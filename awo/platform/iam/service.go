package iam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"awo.so/awo/auth"
	"awo.so/awo/runtime"
	goredis "github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const (
	// humanSessionTTL is the default session lifetime for human users.
	humanSessionTTL = 24 * time.Hour

	// apiTokenCacheTTL is how long a validated API token is cached in Redis.
	// Avoids a database round-trip on every service account request.
	// A short TTL (60s) ensures revocation takes effect promptly.
	apiTokenCacheTTL = 60 * time.Second

	// serviceAccountRoles is the default role set for all service accounts.
	// Service account role management is a future enhancement; for now all
	// service accounts carry role:api-client.
	serviceAccountDefaultRole = "role:api-client"
)

// AuthService provides the authentication and session management operations
// that the API middleware and IAM handlers depend on.
//
// It is the sole writer to:
//   - Redis session store ("session:{token}" keys)
//   - Redis user session index ("user_sessions:{tenantID}:{userID}" sorted sets)
//   - Redis API token cache ("iam:api_token:{sha256_hex}" keys)
//   - iam_sessions table (SQL audit trail for human sessions)
//   - iam_login_audits table (append-only auth event log)
//
// AuthService is constructed once at startup by [New] and shared across all
// concurrent requests. All methods are goroutine-safe.
type AuthService struct {
	DB    *pgxpool.Pool
	Redis *goredis.Client
}

// ── Login ──────────────────────────────────────────────────────────────────────

// LoginInput carries the credentials and metadata for a login attempt.
type LoginInput struct {
	Email     string
	Password  string
	TenantID  uuid.UUID
	DeviceID  string
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
// The entire credential verification sequence runs inside a single database
// transaction so that the tenant RLS context (set by set_tenant_context) is
// guaranteed to be active for all subsequent queries on the same connection.
// Without a transaction, pgxpool may dispatch each DB call to a different
// connection where the tenant context has not been set, making RLS invisible.
//
// Flow:
//  1. Open a read transaction and establish tenant RLS context.
//  2. Load the iam_user record by email (RLS scopes to tenant).
//  3. Verify the bcrypt password hash (constant-time comparison).
//  4. Load the user's active roles.
//  5. Commit the read transaction.
//  6. Generate a cryptographically random 256-bit session token.
//  7. Store the session in Redis with TTL (critical path — failure aborts login).
//  8. Write SQL audit records in a separate best-effort transaction.
//
// On credential failure, Login returns a generic error to prevent user
// enumeration. The specific failure reason is written to iam_login_audits.
//
// Redis failure on step 7 causes Login to fail — sessions cannot be issued when
// the session store is unavailable. This is correct security behaviour.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	// Phase 1: Verify credentials in a transaction.
	// The transaction keeps the tenant RLS context active across all DB calls.
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("iam: login: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) // no-op after Commit; ensures cleanup on early return

	// Establish tenant RLS context for all queries in this transaction.
	// set_tenant_context() calls set_config('app.current_tenant_id', $1, TRUE).
	// The TRUE (is_local) flag makes the setting transaction-local — it reverts
	// automatically on COMMIT or ROLLBACK, so the connection returned to the pool
	// is clean and will not leak the tenant context to the next request.
	if _, err := tx.Exec(ctx, "SELECT set_tenant_context($1)", input.TenantID); err != nil {
		return nil, fmt.Errorf("iam: login: set tenant context: %w", err)
	}

	var (
		userID       uuid.UUID
		passwordHash string
		status       string
	)
	err = tx.QueryRow(ctx, `
		SELECT id, password_hash, status
		FROM iam_users
		WHERE email = $1
	`, input.Email).Scan(&userID, &passwordHash, &status)
	if err != nil {
		// Rollback the read transaction before writing the failure audit.
		tx.Rollback(ctx)
		s.writeFailedLoginAudit(ctx, input, uuid.Nil, "user_not_found")
		return nil, &runtime.BusinessError{
			Code:    "iam.login.invalid_credentials",
			Message: "Invalid email or password.",
			Status:  401,
		}
	}

	if status != "active" {
		tx.Rollback(ctx)
		s.writeFailedLoginAudit(ctx, input, userID, "account_"+status)
		return nil, &runtime.BusinessError{
			Code:    "iam.login.account_" + status,
			Message: "Your account is " + status + ". Contact your administrator.",
			Status:  403,
		}
	}

	// bcrypt.CompareHashAndPassword performs a constant-time comparison, making
	// it resistant to timing attacks that could reveal whether a user exists.
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)); err != nil {
		tx.Rollback(ctx)
		s.writeFailedLoginAudit(ctx, input, userID, "invalid_password")
		return nil, &runtime.BusinessError{
			Code:    "iam.login.invalid_credentials",
			Message: "Invalid email or password.",
			Status:  401,
		}
	}

	// Load roles within the same transaction (RLS context is active).
	roles, err := s.loadUserRoles(ctx, tx, input.TenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("iam: login: load roles: %w", err)
	}

	// Commit the read transaction — credentials are verified and roles are loaded.
	// Subsequent operations (Redis, audit writes) use separate connections.
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("iam: login: commit: %w", err)
	}

	// Phase 2: Generate and store the session.
	// Token is 256 bits of cryptographically random entropy, base64url-encoded.
	token, err := auth.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("iam: login: generate token: %w", err)
	}

	now := time.Now()
	session := &auth.Session{
		Token:            token,
		UserID:           userID,
		ServiceAccountID: uuid.Nil,
		TenantID:         input.TenantID,
		Roles:            roles,
		ExpiresAt:        now.Add(humanSessionTTL),
		IssuedAt:         now,
		DeviceID:         input.DeviceID,
		IPAddress:        input.IPAddress,
	}

	// Store in Redis — this is the critical path. If Redis is unavailable, login
	// fails. A session that exists only in Redis but not in SQL is live and valid;
	// the SQL record is an audit trail, not the authoritative session store.
	if err := s.storeSession(ctx, session); err != nil {
		return nil, fmt.Errorf("iam: login: store session: %w", err)
	}

	// Phase 3: Best-effort audit writes in a separate mini-transaction.
	// These are non-fatal — the session is already live in Redis.
	// A separate transaction is required because the original read tx is committed.
	s.auditLogin(ctx, session, input)

	return &LoginResult{Token: token, Session: session}, nil
}

// auditLogin writes the session SQL record, login audit event, and updates
// last_login_at in a single best-effort transaction. Failures are silently
// ignored — the session is already live in Redis and the user is authenticated.
//
// A new mini-transaction is opened here because the verification transaction
// has already been committed. Tenant RLS context must be re-established.
func (s *AuthService) auditLogin(ctx context.Context, session *auth.Session, input LoginInput) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		// TODO: emit metric for audit tx open failure
		return
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_tenant_context($1)", session.TenantID); err != nil {
		return
	}

	_ = s.insertSessionRecord(ctx, tx, session)
	_ = s.insertLoginAudit(ctx, tx, session.TenantID, session.UserID, uuid.Nil, "login", input.IPAddress, input.DeviceID, "")
	_, _ = tx.Exec(ctx, `UPDATE iam_users SET last_login_at = $1 WHERE id = $2`, session.IssuedAt, session.UserID)

	_ = tx.Commit(ctx)
}

// ── Logout ────────────────────────────────────────────────────────────────────

// Logout invalidates an active session identified by the session record.
//
// Flow:
//  1. DEL session:{token} from Redis (authoritative revocation).
//  2. ZREM token from the user_sessions sorted set index.
//  3. UPDATE iam_sessions SET revoked_at in a tenant-scoped transaction.
//  4. INSERT iam_login_audits logout event.
func (s *AuthService) Logout(ctx context.Context, session *auth.Session) error {
	indexKey := fmt.Sprintf("user_sessions:%s:%s", session.TenantID, session.UserID)

	// Redis is the authoritative session store. Revocation is a DEL + ZREM,
	// executed atomically via pipeline to minimize race window.
	pipe := s.Redis.Pipeline()
	pipe.Del(ctx, session.RedisKey())
	pipe.ZRem(ctx, indexKey, session.Token)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("iam: logout: redis: %w", err)
	}

	// Best-effort: mark the SQL audit record revoked and write a logout event.
	// Uses a separate transaction to establish tenant RLS context.
	s.auditLogout(ctx, session)
	return nil
}

// auditLogout marks the session revoked in iam_sessions and writes the logout
// event to iam_login_audits. Best-effort — failure does not affect the Redis
// revocation that has already occurred.
func (s *AuthService) auditLogout(ctx context.Context, session *auth.Session) {
	hash := tokenHash(session.Token)

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_tenant_context($1)", session.TenantID); err != nil {
		return
	}

	_, _ = tx.Exec(ctx, `
		UPDATE iam_sessions SET revoked_at = now()
		WHERE token_hash = $1 AND tenant_id = $2
	`, hash, session.TenantID)

	_ = s.insertLoginAudit(ctx, tx, session.TenantID, session.UserID, uuid.Nil, "logout", session.IPAddress, session.DeviceID, "")

	_ = tx.Commit(ctx)
}

// ── RevokeUserSessions ────────────────────────────────────────────────────────

// RevokeUserSessions revokes ALL active sessions for the given user within the
// tenant. Called by the UserRoleChangeHook after role assignments change, and by
// IAM administrators via the management API.
//
// Session tokens are retrieved from the user_sessions:{tenantID}:{userID} sorted
// set. All session keys are deleted from Redis atomically. The SQL audit table is
// updated and an audit event written in a best-effort transaction.
//
// Note: if storeSession failed to add a token to the sorted set (it is
// best-effort), RevokeUserSessions will not find that token and cannot revoke it.
// This edge case cannot be avoided without making session creation synchronous
// on the sorted set write.
func (s *AuthService) RevokeUserSessions(ctx context.Context, tenantID, userID uuid.UUID) error {
	indexKey := fmt.Sprintf("user_sessions:%s:%s", tenantID, userID)
	tokens, err := s.Redis.ZRange(ctx, indexKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("iam: revoke user sessions: read index: %w", err)
	}
	if len(tokens) == 0 {
		return nil
	}

	// Build the list of session keys to delete.  Use the same key format as
	// Session.RedisKey() ("session:{token}") to avoid format drift.
	keys := make([]string, 0, len(tokens)+1)
	hashes := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		keys = append(keys, "session:"+tok) // matches auth.Session.RedisKey()
		hashes = append(hashes, tokenHash(tok))
	}
	keys = append(keys, indexKey) // delete the index set itself

	if err := s.Redis.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("iam: revoke user sessions: redis DEL: %w", err)
	}

	// Best-effort: mark sessions revoked in SQL and write audit event.
	s.auditRevoke(ctx, tenantID, userID, hashes)
	return nil
}

// auditRevoke marks sessions revoked in iam_sessions and writes a
// session_revoked audit event. Best-effort — Redis revocation has already
// occurred.
func (s *AuthService) auditRevoke(ctx context.Context, tenantID, userID uuid.UUID, hashes []string) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_tenant_context($1)", tenantID); err != nil {
		return
	}

	_, _ = tx.Exec(ctx, `
		UPDATE iam_sessions SET revoked_at = now()
		WHERE token_hash = ANY($1) AND tenant_id = $2 AND revoked_at IS NULL
	`, hashes, tenantID)

	_ = s.insertLoginAudit(ctx, tx, tenantID, userID, uuid.Nil, "session_revoked", "", "", "admin_revoke_all")

	_ = tx.Commit(ctx)
}

// ── ValidateToken ─────────────────────────────────────────────────────────────

// ValidateToken looks up the session token in Redis, deserializes the session,
// and validates the expiry timestamp. Called on every authenticated request by
// the RequireAuth middleware.
//
// Redis failure semantics:
//   - goredis.Nil (key not found or expired) → 401 Unauthorized. The session
//     does not exist; the client must re-authenticate.
//   - Any other Redis error → 503 Service Unavailable. This is an infrastructure
//     failure. Masking it as 401 would hide outages behind spurious auth errors.
//
// The Redis TTL provides primary expiry enforcement. The IsExpired wall-clock
// check is a defense-in-depth measure in case of clock skew or TTL misconfiguration.
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*auth.Session, error) {
	data, err := s.Redis.Get(ctx, "session:"+token).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			// Session key not found — it was never issued, has expired (TTL), or
			// was explicitly revoked via DEL by Logout/RevokeUserSessions.
			return nil, &runtime.BusinessError{
				Code:    "iam.session.not_found",
				Message: "Session not found or expired.",
				Status:  401,
			}
		}
		// Infrastructure failure — Redis is unavailable.  Return 503 so the
		// client knows to retry rather than re-authenticate unnecessarily.
		return nil, &runtime.BusinessError{
			Code:    "iam.service_unavailable",
			Message: "Authentication service temporarily unavailable.",
			Status:  503,
		}
	}

	var session auth.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("iam: validate token: unmarshal session: %w", err)
	}

	// Defense-in-depth: check wall-clock expiry even though Redis TTL should
	// have already removed the key.  Protects against clock skew or TTL bugs.
	if session.IsExpired(time.Now()) {
		return nil, &runtime.BusinessError{
			Code:    "iam.session.expired",
			Message: "Session has expired.",
			Status:  401,
		}
	}

	return &session, nil
}

// ── ValidateAPIToken ──────────────────────────────────────────────────────────

// ValidateAPIToken authenticates a service account API token and returns a
// synthetic session for the service account principal.
//
// API tokens are long-lived opaque strings. They are validated by SHA-256
// hashing the raw token and looking up the hash in iam_api_tokens. To avoid a
// database round-trip on every request, the validation result is cached in Redis
// for apiTokenCacheTTL (60 seconds). Cache invalidation on token revocation is
// the caller's responsibility (revoke the iam_api_token record and DEL the cache key).
//
// The returned Session carries:
//   - ServiceAccountID (not UserID) as the authenticated principal.
//   - Roles: ["role:api-client"] (the default service account role).
//   - ExpiresAt: from the iam_api_tokens.expires_at column.
//
// Redis failure semantics follow the same rules as ValidateToken:
//   - goredis.Nil on cache miss → fall through to DB lookup (normal path).
//   - Any other Redis error → 503.
//
// tenantID must be pre-resolved by the TenantResolver middleware before this
// method is called. It is used to establish RLS context for the DB lookup.
func (s *AuthService) ValidateAPIToken(ctx context.Context, rawToken string, tenantID uuid.UUID) (*auth.Session, error) {
	hash := tokenHash(rawToken)
	cacheKey := "iam:api_token:" + hash

	// Fast path: check the Redis cache first (60-second TTL).
	data, err := s.Redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		// Cache hit — deserialize and validate.
		var session auth.Session
		if jsonErr := json.Unmarshal(data, &session); jsonErr != nil {
			return nil, fmt.Errorf("iam: validate api token: unmarshal cache: %w", jsonErr)
		}
		if session.IsExpired(time.Now()) {
			// Token expired since it was cached.  Evict the stale cache entry.
			_ = s.Redis.Del(ctx, cacheKey).Err()
			return nil, &runtime.BusinessError{
				Code:    "iam.api_token.expired",
				Message: "API token has expired.",
				Status:  401,
			}
		}
		return &session, nil
	}
	if !errors.Is(err, goredis.Nil) {
		// Infrastructure failure — Redis is unavailable.
		return nil, &runtime.BusinessError{
			Code:    "iam.service_unavailable",
			Message: "Authentication service temporarily unavailable.",
			Status:  503,
		}
	}

	// Cache miss — look up the token in the database.
	// A transaction is required to establish tenant RLS context.
	session, err := s.lookupAPIToken(ctx, hash, tenantID)
	if err != nil {
		return nil, err
	}

	// Populate the cache for subsequent requests.  TTL is the minimum of
	// apiTokenCacheTTL and the remaining token lifetime.
	ttl := apiTokenCacheTTL
	if remaining := session.TTL(time.Now()); remaining < ttl {
		ttl = remaining
	}
	if ttl > 0 {
		if cacheData, marshalErr := json.Marshal(session); marshalErr == nil {
			_ = s.Redis.Set(ctx, cacheKey, cacheData, ttl).Err()
		}
	}

	return session, nil
}

// lookupAPIToken queries iam_api_tokens and iam_service_accounts for the given
// token hash within the tenant. Returns a synthetic auth.Session on success.
// Runs inside a transaction to ensure the tenant RLS context is active.
func (s *AuthService) lookupAPIToken(ctx context.Context, hash string, tenantID uuid.UUID) (*auth.Session, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("iam: lookup api token: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_tenant_context($1)", tenantID); err != nil {
		return nil, fmt.Errorf("iam: lookup api token: set tenant context: %w", err)
	}

	var (
		serviceAccountID uuid.UUID
		expiresAt        *time.Time // nullable
		isRevoked        bool
		saStatus         string
	)
	err = tx.QueryRow(ctx, `
		SELECT
			at.service_account_id,
			at.expires_at,
			at.is_revoked,
			sa.status
		FROM iam_api_tokens at
		JOIN iam_service_accounts sa ON sa.id = at.service_account_id
		WHERE at.token_hash = $1
	`, hash).Scan(&serviceAccountID, &expiresAt, &isRevoked, &saStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &runtime.BusinessError{
				Code:    "iam.api_token.not_found",
				Message: "Invalid API token.",
				Status:  401,
			}
		}
		return nil, fmt.Errorf("iam: lookup api token: query: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("iam: lookup api token: commit: %w", err)
	}

	if isRevoked {
		return nil, &runtime.BusinessError{
			Code:    "iam.api_token.revoked",
			Message: "API token has been revoked.",
			Status:  401,
		}
	}

	now := time.Now()
	if expiresAt != nil && expiresAt.Before(now) {
		return nil, &runtime.BusinessError{
			Code:    "iam.api_token.expired",
			Message: "API token has expired.",
			Status:  401,
		}
	}

	if saStatus != "active" {
		return nil, &runtime.BusinessError{
			Code:    "iam.service_account." + saStatus,
			Message: "Service account is " + saStatus + ".",
			Status:  403,
		}
	}

	// Build a synthetic session for the service account.
	// Service accounts use ServiceAccountID; UserID is uuid.Nil.
	// All service accounts carry role:api-client as the default role.
	// Role management for service accounts is a future enhancement.
	var tokenExpiry time.Time
	if expiresAt != nil {
		tokenExpiry = *expiresAt
	} else {
		// Non-expiring token — set a far-future expiry for the session struct.
		tokenExpiry = now.Add(365 * 24 * time.Hour)
	}

	session := &auth.Session{
		Token:            "",             // not stored in Redis under session:{token}
		UserID:           uuid.Nil,
		ServiceAccountID: serviceAccountID,
		TenantID:         tenantID,
		Roles:            []string{serviceAccountDefaultRole},
		ExpiresAt:        tokenExpiry,
		IssuedAt:         now,
	}

	return session, nil
}

// ── LoadRolePermissions ───────────────────────────────────────────────────────

// LoadRolePermissions returns all role-to-permission bindings from the global
// iam_role_permissions table. Called at startup to initialize
// [auth.CasbinEvaluator] Phase 2, and after any role-permission change to
// trigger a policy reload.
//
// iam_role_permissions is a global table (no tenant_id, no RLS). This method
// queries it without setting tenant context.
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

// ── Internal helpers ───────────────────────────────────────────────────────────

// loadUserRoles queries the iam_user_roles table for all role names assigned to
// the user within the tenant. Must be called within a transaction that has
// already established the tenant RLS context via set_tenant_context.
func (s *AuthService) loadUserRoles(ctx context.Context, tx pgx.Tx, tenantID, userID uuid.UUID) ([]string, error) {
	rows, err := tx.Query(ctx, `
		SELECT role_name FROM iam_user_roles
		WHERE tenant_id = $1 AND user_id = $2
	`, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("iam: load user roles: query: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return nil, fmt.Errorf("iam: load user roles: scan: %w", err)
		}
		roles = append(roles, r)
	}
	return roles, rows.Err()
}

// storeSession serializes the session to JSON and stores it in Redis under
// session:{token} with a TTL matching the session's ExpiresAt.
//
// Additionally, the token is added to the user_sessions:{tenantID}:{userID}
// sorted set (scored by expiry unix timestamp) to enable bulk revocation via
// [RevokeUserSessions]. The sorted set write is best-effort: if it fails, the
// session is still valid but RevokeUserSessions may not find it.
func (s *AuthService) storeSession(ctx context.Context, session *auth.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("iam: store session: marshal: %w", err)
	}

	ttl := session.TTL(time.Now())
	if ttl <= 0 {
		return fmt.Errorf("iam: store session: session already expired")
	}

	if err := s.Redis.Set(ctx, session.RedisKey(), data, ttl).Err(); err != nil {
		return fmt.Errorf("iam: store session: redis SET: %w", err)
	}

	// Add to the user session index for bulk revocation on role changes.
	// Score = expiry unix timestamp; expired members can be pruned by ZRANGEBYSCORE.
	indexKey := fmt.Sprintf("user_sessions:%s:%s", session.TenantID, session.UserID)
	score := float64(session.ExpiresAt.Unix())
	if err := s.Redis.ZAdd(ctx, indexKey, &goredis.Z{Score: score, Member: session.Token}).Err(); err != nil {
		// Non-fatal: the session is live; only bulk revocation is impaired.
		// TODO: emit metric for index write failure
		_ = err
	}
	return nil
}

// insertSessionRecord writes a session audit record to iam_sessions.
// Must be called within a transaction that has established tenant RLS context.
func (s *AuthService) insertSessionRecord(ctx context.Context, tx pgx.Tx, session *auth.Session) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO iam_sessions
			(id, tenant_id, token_hash, user_id, service_account_id,
			 issued_at, expires_at, device_id, ip_address)
		VALUES
			(gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8)
	`,
		session.TenantID,
		tokenHash(session.Token),
		nullUUID(session.UserID),
		nullUUID(session.ServiceAccountID),
		session.IssuedAt,
		session.ExpiresAt,
		session.DeviceID,
		session.IPAddress,
	)
	if err != nil {
		return fmt.Errorf("iam: insert session record: %w", err)
	}
	return nil
}

// insertLoginAudit writes an authentication event to iam_login_audits.
// Must be called within a transaction that has established tenant RLS context.
func (s *AuthService) insertLoginAudit(ctx context.Context, tx pgx.Tx, tenantID, userID, serviceAccountID uuid.UUID, event, ipAddress, deviceID, failureReason string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO iam_login_audits
			(id, tenant_id, event, user_id, service_account_id,
			 ip_address, device_id, failure_reason, created_at)
		VALUES
			(gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, now())
	`,
		tenantID,
		event,
		nullUUID(userID),
		nullUUID(serviceAccountID),
		ipAddress,
		deviceID,
		failureReason,
	)
	if err != nil {
		return fmt.Errorf("iam: insert login audit: %w", err)
	}
	return nil
}

// writeFailedLoginAudit writes a failed_login event in its own mini-transaction.
// The caller's transaction (if any) is already rolled back or not yet committed,
// so a separate connection is needed. Best-effort — failure is silently ignored.
func (s *AuthService) writeFailedLoginAudit(ctx context.Context, input LoginInput, userID uuid.UUID, reason string) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT set_tenant_context($1)", input.TenantID); err != nil {
		return
	}
	_ = s.insertLoginAudit(ctx, tx, input.TenantID, userID, uuid.Nil, "failed_login", input.IPAddress, input.DeviceID, reason)
	_ = tx.Commit(ctx)
}

// ── Helpers ────────────────────────────────────────────────────────────────────

// tokenHash is declared in hooks.go (shared across the package).

// nullUUID returns nil for uuid.Nil, otherwise returns the UUID value.
// Used to store nullable UUID columns correctly in pgx (uuid.Nil → NULL).
func nullUUID(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
}
