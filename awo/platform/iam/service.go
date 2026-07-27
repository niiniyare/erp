package iam

import (
	"context"
	"errors"
	"fmt"
	"time"

	"awo.so/awo/audit"
	"awo.so/awo/auth"
	"awo.so/awo/cache"
	"awo.so/awo/def"
	"awo.so/awo/runtime"
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
//   - Session store (human sessions, via [Sessions])
//   - API token cache (service account token cache, via [Cache])
//   - iam_sessions table (SQL audit trail for human sessions)
//   - platform_audit_log (unified audit trail, via [AuditWriter])
//
// AuthService is constructed once at startup by [New] and shared across all
// concurrent requests. All methods are goroutine-safe.
type AuthService struct {
	DB          *pgxpool.Pool
	Sessions    auth.SessionStore
	Cache       cache.Cache
	AuditWriter audit.AuditWriter // nil = audit disabled (dev/test)
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
	if _, err := tx.Exec(ctx, sqlSetTenantContext, input.TenantID); err != nil {
		return nil, fmt.Errorf("iam: login: set tenant context: %w", err)
	}

	var (
		userID       uuid.UUID
		passwordHash string
		status       string
	)
	err = tx.QueryRow(ctx, sqlLookupCredentials, input.Email).Scan(&userID, &passwordHash, &status)
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

	if _, err := tx.Exec(ctx, sqlSetTenantContext, session.TenantID); err != nil {
		return
	}

	_ = s.insertSessionRecord(ctx, tx, session)
	_, _ = tx.Exec(ctx, sqlUpdateLastLoginAt, session.IssuedAt, session.UserID)

	_ = tx.Commit(ctx)

	// Write the unified audit record after the TX commits — platform_audit_log
	// has no RLS so no tenant context is required; the fallback querier is used.
	s.writeAuthAudit(ctx, session.TenantID,
		&def.Actor{UserID: session.UserID, TenantID: session.TenantID},
		audit.OperationLogin,
		input.IPAddress,
		map[string]any{"device_id": input.DeviceID},
	)
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
	// SessionStore.Delete is the authoritative revocation. The primary session
	// key removal is critical (error returned on failure); the user index entry
	// is removed best-effort inside Delete.
	if err := s.Sessions.Delete(ctx, session); err != nil {
		return fmt.Errorf("iam: logout: %w", err)
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

	if _, err := tx.Exec(ctx, sqlSetTenantContext, session.TenantID); err != nil {
		return
	}

	_, _ = tx.Exec(ctx, sqlRevokeSessionByHash, hash, session.TenantID)

	_ = tx.Commit(ctx)

	// Write the unified audit record after TX commits.
	s.writeAuthAudit(ctx, session.TenantID,
		&def.Actor{UserID: session.UserID, TenantID: session.TenantID},
		audit.OperationLogout,
		session.IPAddress,
		map[string]any{"device_id": session.DeviceID},
	)
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
	tokens, err := s.Sessions.ListUserTokens(ctx, tenantID, userID)
	if err != nil {
		return fmt.Errorf("iam: revoke user sessions: list tokens: %w", err)
	}
	if len(tokens) == 0 {
		return nil
	}

	hashes := make([]string, len(tokens))
	for i, tok := range tokens {
		hashes[i] = tokenHash(tok)
	}

	if err := s.Sessions.DeleteAll(ctx, tenantID, userID, tokens); err != nil {
		return fmt.Errorf("iam: revoke user sessions: delete: %w", err)
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

	if _, err := tx.Exec(ctx, sqlSetTenantContext, tenantID); err != nil {
		return
	}

	_, _ = tx.Exec(ctx, sqlRevokeSessionsByHashes, hashes, tenantID)

	_ = tx.Commit(ctx)

	// Write the unified audit record after TX commits.
	s.writeAuthAudit(ctx, tenantID,
		&def.Actor{UserID: userID, TenantID: tenantID},
		audit.OperationSystem,
		"",
		map[string]any{"reason": "admin_revoke_all", "session_count": len(hashes)},
	)
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
	session, err := s.Sessions.Load(ctx, token)
	if err != nil {
		if errors.Is(err, auth.ErrSessionNotFound) {
			// Key absent — never issued, TTL-expired, or explicitly revoked.
			return nil, &runtime.BusinessError{
				Code:    "iam.session.not_found",
				Message: "Session not found or expired.",
				Status:  401,
			}
		}
		// Infrastructure failure — store unavailable. Return 503 so the client
		// knows to retry rather than re-authenticate unnecessarily.
		return nil, &runtime.BusinessError{
			Code:    "iam.service_unavailable",
			Message: "Authentication service temporarily unavailable.",
			Status:  503,
		}
	}

	// Defense-in-depth: check wall-clock expiry even though store TTL should
	// have already evicted the key. Protects against clock skew or TTL bugs.
	if session.IsExpired(time.Now()) {
		return nil, &runtime.BusinessError{
			Code:    "iam.session.expired",
			Message: "Session has expired.",
			Status:  401,
		}
	}

	return session, nil
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

	// Fast path: check the token cache (60-second TTL).
	var cached auth.Session
	err := s.Cache.Get(ctx, cacheKey, &cached)
	if err == nil {
		// Cache hit — validate expiry.
		if cached.IsExpired(time.Now()) {
			// Token expired since it was cached. Evict the stale entry.
			_ = s.Cache.Delete(ctx, cacheKey)
			return nil, &runtime.BusinessError{
				Code:    "iam.api_token.expired",
				Message: "API token has expired.",
				Status:  401,
			}
		}
		return &cached, nil
	}
	if !errors.Is(err, cache.ErrMiss) {
		// Infrastructure failure — cache unavailable.
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

	// Populate cache for subsequent requests. TTL is the minimum of
	// apiTokenCacheTTL and the remaining token lifetime.
	ttl := apiTokenCacheTTL
	if remaining := session.TTL(time.Now()); remaining < ttl {
		ttl = remaining
	}
	if ttl > 0 {
		_ = s.Cache.Set(ctx, cacheKey, session, ttl)
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

	if _, err := tx.Exec(ctx, sqlSetTenantContext, tenantID); err != nil {
		return nil, fmt.Errorf("iam: lookup api token: set tenant context: %w", err)
	}

	var (
		serviceAccountID uuid.UUID
		expiresAt        *time.Time // nullable
		isRevoked        bool
		saStatus         string
	)
	err = tx.QueryRow(ctx, sqlLookupAPIToken, hash).Scan(&serviceAccountID, &expiresAt, &isRevoked, &saStatus)
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
		Token:            "", // not stored in Redis under session:{token}
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
	rows, err := s.DB.Query(ctx, sqlLoadRolePermissions)
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

// ── Audit helpers ──────────────────────────────────────────────────────────────

// writeAuthAudit emits an AUTH-category audit record to the unified pipeline.
// Best-effort — failures are silently ignored; session state is already committed.
//
// op should be OperationLogin, OperationLogout, or OperationSystem.
// actor carries the authenticated principal; context carries event-specific metadata.
func (s *AuthService) writeAuthAudit(ctx context.Context, tenantID uuid.UUID, actor *def.Actor, op audit.OperationType, ipAddress string, extra map[string]any) {
	if s.AuditWriter == nil {
		return
	}
	rec := audit.AuditRecord{
		TenantID:      tenantID,
		EntityName:    "iam_session",
		Operation:     op,
		EventCategory: audit.CategoryAuth,
		Actor:         actor,
		IPAddress:     ipAddress,
		Context:       extra,
	}
	_ = s.AuditWriter.Write(ctx, rec)
}

// ── Internal helpers ───────────────────────────────────────────────────────────

// loadUserRoles queries the iam_user_roles table for all role names assigned to
// the user within the tenant. Must be called within a transaction that has
// already established the tenant RLS context via set_tenant_context.
func (s *AuthService) loadUserRoles(ctx context.Context, tx pgx.Tx, tenantID, userID uuid.UUID) ([]string, error) {
	rows, err := tx.Query(ctx, sqlLoadUserRoles, tenantID, userID)
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

// storeSession persists the session via [SessionStore.Store].
// Called on the critical login path — failure aborts login.
func (s *AuthService) storeSession(ctx context.Context, session *auth.Session) error {
	if err := s.Sessions.Store(ctx, session); err != nil {
		return fmt.Errorf("iam: store session: %w", err)
	}
	return nil
}

// insertSessionRecord writes a session audit record to iam_sessions.
// Must be called within a transaction that has established tenant RLS context.
func (s *AuthService) insertSessionRecord(ctx context.Context, tx pgx.Tx, session *auth.Session) error {
	_, err := tx.Exec(ctx, sqlInsertSessionRecord,
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

// writeFailedLoginAudit writes a failed_login audit record to the unified pipeline.
// Best-effort — failures are silently ignored; the login error is already returned to
// the caller. platform_audit_log has no RLS so no TX or tenant context is required.
func (s *AuthService) writeFailedLoginAudit(ctx context.Context, input LoginInput, userID uuid.UUID, reason string) {
	s.writeAuthAudit(ctx, input.TenantID,
		&def.Actor{UserID: userID, TenantID: input.TenantID},
		audit.OperationLogin,
		input.IPAddress,
		map[string]any{"device_id": input.DeviceID, "failure_reason": reason, "failed": true},
	)
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
