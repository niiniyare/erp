package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/iam/domain"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

//  Port (interface)

// SessionRepository defines the persistence port for sessions.
// All cache operations are encapsulated here — callers never touch cache.
type SessionRepository interface {
	Create(ctx context.Context, s domain.Session) error

	// ValidateToken is cache-aside: cache hit → return; miss → DB → populate cache.
	// Returns nil, nil when not found or expired.
	ValidateToken(ctx context.Context, hash string) (*domain.ResolvedSession, error)

	// CacheResolved stores a ResolvedSession built at Login (which has DisplayName).
	// The service calls this after Create so the first ValidateToken hits cache.
	CacheResolved(ctx context.Context, hash string, r *domain.ResolvedSession, ttl time.Duration)

	// Invalidate marks the DB session inactive and evicts the cache entry.
	Invalidate(ctx context.Context, hash string) error

	// InvalidateByUser marks all active sessions for the given user as inactive in the DB.
	// Cache entries for those sessions expire naturally within the session TTL.
	InvalidateByUser(ctx context.Context, userID uuid.UUID) error

	// InvalidateByTenant marks all active sessions for the given tenant as inactive in the DB.
	// Cache entries for those sessions expire naturally within the session TTL.
	InvalidateByTenant(ctx context.Context, tenantID uuid.UUID) error

	// UpdateLastSeen is fire-and-forget.
	UpdateLastSeen(ctx context.Context, hash string)

	// ── Login pre-computation ─────────────────────────────────────────────────

	// LoadLoginConfig fetches the pre-computed login configuration snapshot:
	// feature flags (ResolveAllFlagsForTenant), tenant settings
	// (ResolveAllSettingsForTenant), and user preferences (GetUserPreferences).
	// Called exactly once during Login; the result is embedded in the session row.
	// Non-fatal: partial failures populate what's available and log warnings.
	LoadLoginConfig(ctx context.Context, userID, tenantID uuid.UUID) (domain.Configuration, error)

	// ResolveEntityScope derives the session's EntityScope from the user's entity.
	// Called exactly once during Login; the result is embedded in the session row.
	// Non-fatal: falls back to EntityScopeEntity on DB error.
	ResolveEntityScope(ctx context.Context, entityID uuid.UUID) (domain.EntityScope, error)

	// MFA pending login state — temporary (5 min) pre-session for MFA step 2.
	StorePendingMFA(ctx context.Context, pendingToken string, userID uuid.UUID) error
	GetPendingMFA(ctx context.Context, pendingToken string) (uuid.UUID, error)
	DeletePendingMFA(ctx context.Context, pendingToken string) error
}

//  Adapter (implementation)

type sessionRepo struct {
	store   db.Store
	cache   cache.Service
	tracing tracing.Service
	metrics metrics.MetricsProvider
}

// NewSessionRepository constructs a cache-backed Postgres SessionRepository.
func NewSessionRepository(
	store db.Store,
	cacheSvc cache.Service,
	tracer tracing.Service,
	m metrics.MetricsProvider,
) SessionRepository {
	return &sessionRepo{
		store:   store,
		cache:   cacheSvc,
		tracing: tracer,
		metrics: m,
	}
}

func (r *sessionRepo) Create(ctx context.Context, s domain.Session) error {
	ctx, span := r.tracing.StartSpan(ctx, "session.repo.Create")
	defer span.End()

	permsJSON, err := json.Marshal(s.Permissions)
	if err != nil {
		return fmt.Errorf("session repo: marshal permissions: %w", err)
	}

	entityScopeJSON := domain.MarshalSessionJSON(s.EntityScope)
	configJSON := domain.MarshalSessionJSON(s.Configuration)

	// s.PrincipalID is already *uuid.UUID — nil means non-portal session.
	// No indirection needed; assign directly.
	var principalID *uuid.UUID
	if s.PrincipalID != nil {
		principalID = s.PrincipalID
	}

	var ipAddr *netip.Addr
	if s.IPAddress != "" {
		if parsed, err := netip.ParseAddr(s.IPAddress); err == nil {
			ipAddr = &parsed
		}
	}

	var userAgent *string
	if s.UserAgent != "" {
		userAgent = &s.UserAgent
	}

	var userType *string
	if s.UserType != "" {
		userType = &s.UserType
	}

	riskScore := int32(s.RiskScore)

	if err := r.store.CreateSession(ctx, db.CreateSessionParams{
		UserID:        s.UserID,
		UserType:      userType,
		SessionToken:  s.TokenHash,
		Permissions:   permsJSON,
		PrincipalID:   principalID,
		EntityScope:   entityScopeJSON,
		Configuration: configJSON,
		IpAddress:     ipAddr,
		UserAgent:     userAgent,
		ExpiresAt:     s.ExpiresAt,
		RiskScore:     &riskScore,
	}); err != nil {
		return fmt.Errorf("session repo: create session: %w", err)
	}
	return nil
}

func (r *sessionRepo) ValidateToken(ctx context.Context, hash string) (*domain.ResolvedSession, error) {
	ctx, span := r.tracing.StartSpan(ctx, "session.repo.ValidateToken")
	defer span.End()

	// 1. Cache hit
	if cached := r.getCachedResolved(ctx, hash); cached != nil {
		r.UpdateLastSeen(ctx, hash)
		return cached, nil
	}

	// 2. DB fallback
	row, err := r.store.TouchAndGetSession(ctx, hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("session repo: get session: %w", err)
	}
	if row == nil {
		return nil, nil
	}

	// 3. Convert session row → ResolvedSession

	perms := make(map[string]bool)
	if len(row.Permissions) > 0 {
		if err := json.Unmarshal(row.Permissions, &perms); err != nil {
			return nil, fmt.Errorf("session repo: unmarshal permissions: %w", err)
		}
	}

	var entityScope domain.EntityScope
	if len(row.EntityScope) > 0 {
		_ = json.Unmarshal(row.EntityScope, &entityScope)
	}

	var config domain.Configuration = domain.DefaultConfiguration()
	if len(row.Configuration) > 0 {
		_ = json.Unmarshal(row.Configuration, &config)
	}

	var utype string
	if row.UserType != nil {
		utype = *row.UserType
	}

	// PrincipalID is *uuid.UUID — nil for all non-portal sessions.
	// row.PrincipalID should be *uuid.UUID from the SQLC-generated type;
	// assign directly without wrapping.
	resolved := &domain.ResolvedSession{
		UserID:        row.UserID,
		UserType:      utype,
		TenantID:      row.TenantID,
		PrincipalID:   row.PrincipalID, // *uuid.UUID; nil for non-portal sessions
		Permissions:   perms,
		EntityScope:   entityScope,
		Configuration: config,
		// DisplayName is not stored on the session row — it is populated only
		// on cache-hits that originate from the Login path.
	}

	// 4. Re-populate cache for future requests
	if ttl := time.Until(row.ExpiresAt); ttl > 0 {
		r.cacheResolved(ctx, hash, resolved, ttl)
	}

	return resolved, nil
}

func (r *sessionRepo) CacheResolved(ctx context.Context, hash string, resolved *domain.ResolvedSession, ttl time.Duration) {
	r.cacheResolved(ctx, hash, resolved, ttl)
}

func (r *sessionRepo) Invalidate(ctx context.Context, hash string) error {
	ctx, span := r.tracing.StartSpan(ctx, "session.repo.Invalidate")
	defer span.End()

	if err := r.store.InvalidateSession(ctx, hash); err != nil {
		return fmt.Errorf("session repo: invalidate session: %w", err)
	}
	_ = r.cache.Delete(ctx, sessionCacheKey(hash))
	return nil
}

func (r *sessionRepo) InvalidateByUser(ctx context.Context, userID uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "session.repo.InvalidateByUser")
	defer span.End()
	if err := r.store.InvalidateSessionsByUser(ctx, userID); err != nil {
		return fmt.Errorf("session repo: invalidate by user: %w", err)
	}
	return nil
}

func (r *sessionRepo) InvalidateByTenant(ctx context.Context, tenantID uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "session.repo.InvalidateByTenant")
	defer span.End()
	if err := r.store.InvalidateSessionsByTenant(ctx, tenantID); err != nil {
		return fmt.Errorf("session repo: invalidate by tenant: %w", err)
	}
	return nil
}

func (r *sessionRepo) UpdateLastSeen(ctx context.Context, hash string) {
	go func() {
		_ = r.store.UpdateSessionLastSeen(ctx, hash)
	}()
}

// ── Login pre-computation ─────────────────────────────────────────────────────

func (r *sessionRepo) LoadLoginConfig(ctx context.Context, userID, tenantID uuid.UUID) (domain.Configuration, error) {
	ctx, span := r.tracing.StartSpan(ctx, "session.repo.LoadLoginConfig")
	defer span.End()

	cfg := domain.DefaultConfiguration()

	// Feature flags — resolves tenant overrides over system defaults.
	flagRows, err := r.store.ResolveAllFlagsForTenant(ctx, tenantID)
	if err == nil {
		for _, row := range flagRows {
			cfg.Flags[row.FlagKey] = row.EffectiveValue
		}
	}

	// Tenant settings — resolves tenant overrides over system defaults.
	settingRows, err := r.store.ResolveAllSettingsForTenant(ctx, tenantID)
	if err == nil {
		for _, row := range settingRows {
			cfg.Settings[row.SettingKey] = row.EffectiveValue
		}
	}

	// User preferences — personal UI/UX settings.
	prefRows, err := r.store.GetUserPreferences(ctx, userID)
	if err == nil {
		for _, row := range prefRows {
			cfg.Prefs[row.PrefKey] = row.Value
		}
	}

	r.metrics.IncrementCounter("iam.session.login_config.loaded", nil)
	return cfg, nil
}

func (r *sessionRepo) ResolveEntityScope(ctx context.Context, entityID uuid.UUID) (domain.EntityScope, error) {
	// Platform/system users have no entity assignment — grant full tenant visibility.
	if entityID == uuid.Nil {
		return domain.EntityScope{Type: domain.EntityScopeAll}, nil
	}

	ctx, span := r.tracing.StartSpan(ctx, "session.repo.ResolveEntityScope")
	defer span.End()

	row, err := r.store.ResolveEntityScope(ctx, entityID)
	if err != nil {
		// Non-fatal: fall back to the narrowest safe scope.
		return domain.EntityScope{Type: domain.EntityScopeEntity, EntityID: entityID.String()}, err
	}

	scope := domain.EntityScope{EntityID: entityID.String()}
	switch {
	case row.EntityLevel == 1:
		// Root entity → full tenant visibility.
		scope.Type = domain.EntityScopeAll
		scope.EntityID = "" // EntityID is only meaningful for entity/subtree scopes.
	case row.HasChildren:
		// Branch entity → subtree visibility.
		scope.Type = domain.EntityScopeSubtree
		if row.EntityPath != nil {
			scope.PathPrefix = *row.EntityPath
		}
	default:
		// Leaf entity → own entity only.
		scope.Type = domain.EntityScopeEntity
	}
	return scope, nil
}

// ── Cache helpers (internal) ──────────────────────────────────────────────────

func (r *sessionRepo) cacheResolved(ctx context.Context, hash string, resolved *domain.ResolvedSession, ttl time.Duration) {
	_ = r.cache.Set(ctx, sessionCacheKey(hash), resolved, ttl)
}

func (r *sessionRepo) getCachedResolved(ctx context.Context, hash string) *domain.ResolvedSession {
	var resolved domain.ResolvedSession
	if err := r.cache.Get(ctx, sessionCacheKey(hash), &resolved); err != nil {
		return nil
	}
	return &resolved
}

func sessionCacheKey(hash string) string { return "session:" + hash }

func derefInt32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}

// keep compiler from complaining about unused derefInt32 if not otherwise used
var _ = derefInt32

// ── MFA pending login state ───────────────────────────────────────────────────

const mfaLoginPendingTTL = 5 * time.Minute

func (r *sessionRepo) StorePendingMFA(ctx context.Context, pendingToken string, userID uuid.UUID) error {
	key := "mfa:login:pending:" + pendingToken
	return r.cache.Set(ctx, key, userID.String(), mfaLoginPendingTTL)
}

func (r *sessionRepo) GetPendingMFA(ctx context.Context, pendingToken string) (uuid.UUID, error) {
	key := "mfa:login:pending:" + pendingToken
	var idStr string
	if err := r.cache.Get(ctx, key, &idStr); err != nil {
		return uuid.Nil, fmt.Errorf("session repo: mfa pending not found or expired")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("session repo: mfa pending: invalid user id")
	}
	return id, nil
}

func (r *sessionRepo) DeletePendingMFA(ctx context.Context, pendingToken string) error {
	return r.cache.Delete(ctx, "mfa:login:pending:"+pendingToken)
}
