package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/authz"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// Config holds session-related settings.
// TODO(settings): read MaxFailedAttempts, LockoutDuration from iam module settings (per-tenant)
// TODO(settings): read SessionTTL from settings key "iam.session_ttl_hours" per tenant
// FIXME(feature-flags): MFAEnabled should be read from feature-flag "iam.mfa_enabled" via condition.Evaluator
type Config struct {
	SessionTTL  time.Duration // default: 8h
	CookieName  string        // default: "session"
	// NOTE(settings): per-tenant overrides for identity.Config thresholds come from
	// the Settings module — see docs/reference/modules/settings/15-service-integration.md §Pattern 1
}

// DefaultConfig returns safe defaults for development.
func DefaultConfig() Config {
	return Config{
		SessionTTL: 8 * time.Hour,
		CookieName: "session",
	}
}

// Service handles Login / ValidateSession / Logout.
type Service interface {
	Login(ctx context.Context, email, password string) (*ResolvedSession, string, error)
	ValidateSession(ctx context.Context, token string) (*ResolvedSession, error)
	Logout(ctx context.Context, token string) error
}

// service is the concrete implementation.
type service struct {
	identity identity.Service
	authz    authz.Service
	repo     Repository
	cache    cache.Service
	tracer   tracing.Service
	metrics  metrics.MetricsProvider
	log      logger.Logger
	cfg      Config
}

// New constructs a session Service with default config.
func New(
	identitySvc identity.Service,
	authzSvc authz.Service,
	repo Repository,
	cacheSvc cache.Service,
	tracer tracing.Service,
	metricsProv metrics.MetricsProvider,
	log logger.Logger,
) Service {
	return NewWithConfig(identitySvc, authzSvc, repo, cacheSvc, tracer, metricsProv, log, DefaultConfig())
}

// NewWithConfig constructs a session Service with explicit config.
func NewWithConfig(
	identitySvc identity.Service,
	authzSvc authz.Service,
	repo Repository,
	cacheSvc cache.Service,
	tracer tracing.Service,
	metricsProv metrics.MetricsProvider,
	log logger.Logger,
	cfg Config,
) Service {
	return &service{
		identity: identitySvc,
		authz:    authzSvc,
		repo:     repo,
		cache:    cacheSvc,
		tracer:   tracer,
		metrics:  metricsProv,
		log:      log,
		cfg:      cfg,
	}
}

// Login authenticates the user, computes their permission map, persists the
// session, caches it, and returns the raw token to the caller.
//
// NOTE(tenant-context): ctx must carry tenant_id via cache.TenantIDKey — set
// by the ResolveTenant middleware before this handler is called.
func (s *service) Login(ctx context.Context, email, password string) (*ResolvedSession, string, error) {
	ctx, span := s.tracer.StartSpan(ctx, "session.Login")
	defer span.End()

	// 1. Authenticate — handles lockout and brute-force tracking (identity S2).
	user, err := s.identity.Authenticate(ctx, email, password)
	if err != nil {
		s.metrics.RecordCount("session.login.failure", 1, nil)
		return nil, "", err
	}

	// 2. Build the pre-computed permission map for O(1) per-request authz.
	perms, err := s.buildPermissions(ctx, user)
	if err != nil {
		// Non-fatal: log and continue with empty map — Casbin fallback still works.
		s.log.WarnContext(ctx, "session: buildPermissions failed, continuing with empty map", logger.Fields{
			"user_id": user.ID.String(),
			"error":   err.Error(),
		})
		perms = make(map[string]bool)
	}

	// 3. Generate opaque token and compute its hash.
	rawToken, hash, err := generateToken()
	if err != nil {
		return nil, "", fmt.Errorf("session: generate token: %w", err)
	}

	// 4. Determine tenant from context.
	tenantID := tenantIDFromCtx(ctx)

	// 5. Build the session row and persist it.
	now := time.Now()
	sess := Session{
		UserID:      user.ID,
		TenantID:    tenantID,
		TokenHash:   hash,
		Permissions: perms,
		IsActive:    true,
		ExpiresAt:   now.Add(s.cfg.SessionTTL),
		LastSeenAt:  now,
		RiskScore:   0,
	}
	if err := s.repo.CreateSession(ctx, sess); err != nil {
		return nil, "", fmt.Errorf("session: persist session: %w", err)
	}

	// 6. Build the resolved view for the caller and the cache.
	resolved := &ResolvedSession{
		UserID:      user.ID,
		UserType:    user.UserType,
		TenantID:    tenantID,
		DisplayName: displayName(user),
		Permissions: perms,
	}

	// 7. Cache keyed by hash, TTL matches session expiry.
	if err := s.cache.Set(ctx, sessionCacheKey(hash), resolved, s.cfg.SessionTTL); err != nil {
		// Non-fatal: the session is in the DB; cache miss just costs a DB round-trip.
		s.log.WarnContext(ctx, "session: cache.Set failed", logger.Fields{"error": err.Error()})
	}

	s.metrics.RecordCount("session.login.success", 1, nil)
	return resolved, rawToken, nil
}

// ValidateSession checks the cache first, then falls back to the DB.
// It also fires an async last-seen update and re-populates the cache on miss.
func (s *service) ValidateSession(ctx context.Context, token string) (*ResolvedSession, error) {
	ctx, span := s.tracer.StartSpan(ctx, "session.ValidateSession")
	defer span.End()

	hash := sha256hex(token)

	// 1. Cache hit — fast path.
	var resolved ResolvedSession
	if err := s.cache.Get(ctx, sessionCacheKey(hash), &resolved); err == nil {
		// Async update — do not block the request.
		s.repo.UpdateLastSeen(ctx, hash)
		return &resolved, nil
	}

	// 2. Cache miss — hit the DB.
	sess, err := s.repo.GetByTokenHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("session: get by hash: %w", err)
	}
	if sess == nil {
		return nil, errors.ErrAuthenticationFailed
	}

	// 3. Re-hydrate ResolvedSession from DB row.
	// NOTE(tenant-context): user display name would require a user lookup; omit for now.
	// TODO(perf): consider caching user display name separately or including it in the session row.
	r := &ResolvedSession{
		UserID:      sess.UserID,
		TenantID:    sess.TenantID,
		PrincipalID: sess.PrincipalID,
		Permissions: sess.Permissions,
		// NOTE: UserType and DisplayName are not stored in user_sessions today.
		// They will be populated once the session row includes a user_type column,
		// or after a follow-up user lookup (TODO: add user_type to CreateSession).
	}

	// 4. Async last-seen update.
	s.repo.UpdateLastSeen(ctx, hash)

	// 5. Re-populate cache.
	ttl := time.Until(sess.ExpiresAt)
	if ttl > 0 {
		if err := s.cache.Set(ctx, sessionCacheKey(hash), r, ttl); err != nil {
			s.log.WarnContext(ctx, "session: re-populate cache failed", logger.Fields{"error": err.Error()})
		}
	}

	return r, nil
}

// Logout invalidates the DB session and removes the cache entry.
func (s *service) Logout(ctx context.Context, token string) error {
	ctx, span := s.tracer.StartSpan(ctx, "session.Logout")
	defer span.End()

	hash := sha256hex(token)

	if err := s.repo.Invalidate(ctx, hash); err != nil {
		return fmt.Errorf("session: logout invalidate: %w", err)
	}
	if err := s.cache.Delete(ctx, sessionCacheKey(hash)); err != nil {
		// Non-fatal: session is already invalidated in DB.
		s.log.WarnContext(ctx, "session: cache.Delete failed on logout", logger.Fields{"error": err.Error()})
	}
	return nil
}

// buildPermissions uses authz.GetRoles + authz.GetPolicies to compute a flat
// permission map at login time. This map is stored in user_sessions.permissions
// and drives O(1) authz checks on every subsequent request.
//
// See docs/reference/modules/authz/14-how-other-packages-use-authz.md §0 for the full design.
func (s *service) buildPermissions(ctx context.Context, user *identity.User) (map[string]bool, error) {
	subject := subjectForUser(user)
	domain := domainForUser(user)

	roles, err := s.authz.GetRoles(ctx, subject, domain)
	if err != nil {
		return nil, fmt.Errorf("buildPermissions: GetRoles: %w", err)
	}

	perms := make(map[string]bool)
	for _, role := range roles {
		policies, err := s.authz.GetPolicies(ctx, domain)
		if err != nil {
			return nil, fmt.Errorf("buildPermissions: GetPolicies for role %s: %w", role, err)
		}
		for _, p := range policies {
			if p.Subject == role && p.Effect == "allow" {
				// Compose as "object.action" — mirrors the permission string used in Can().
				key := p.Object + "." + p.Action
				perms[key] = true
			}
		}
	}
	return perms, nil
}

// --- helpers ---

// generateToken creates a cryptographically random 32-byte token, returning
// both the raw hex string (sent to the client) and its sha256 hash (stored in DB).
func generateToken() (rawToken, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	rawToken = hex.EncodeToString(b)
	hash = sha256hex(rawToken)
	return rawToken, hash, nil
}

// sha256hex returns the lowercase hex-encoded sha256 digest of s.
func sha256hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// sessionCacheKey returns the Redis key for a resolved session.
func sessionCacheKey(hash string) string {
	return "session:" + hash
}

// tenantIDFromCtx reads the tenant UUID from the context.
// Returns uuid.Nil if absent (will be caught by RLS on DB write).
func tenantIDFromCtx(ctx context.Context) uuid.UUID {
	if v, ok := ctx.Value(cache.TenantIDKey).(string); ok && v != "" {
		if id, err := uuid.Parse(v); err == nil {
			return id
		}
	}
	return uuid.Nil
}

// subjectForUser builds the Casbin subject string for a user.
func subjectForUser(user *identity.User) string {
	id := user.ID.String()
	switch user.UserType {
	case string(authz.ActorPlatform):
		return authz.PlatformSubject(id)
	case string(authz.ActorPortal):
		return authz.PortalSubject(id)
	default:
		return authz.TenantSubject(id)
	}
}

// domainForUser builds the Casbin domain for a user.
func domainForUser(user *identity.User) string {
	switch user.UserType {
	case string(authz.ActorPlatform):
		return authz.DomainPlatform
	default:
		return authz.TenantDomain(user.TenantID.String())
	}
}

// displayName returns a human-readable name for the user.
func displayName(user *identity.User) string {
	if user.Username != "" {
		return user.Username
	}
	return user.Email
}
