package session

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"
	db "awo/db/sqlc"
	"awo/internal/platform/cache"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
)

// Repository defines the session persistence interface.
type Repository interface {
	CreateSession(ctx context.Context, s Session) error
	GetByTokenHash(ctx context.Context, hash string) (*Session, error)
	Invalidate(ctx context.Context, hash string) error
	UpdateLastSeen(ctx context.Context, hash string) // fire-and-forget
}

// sessionRepo implements Repository against user_sessions via SQLC.
type sessionRepo struct {
	store   db.Store
	cache   cache.Service
	tracing tracing.Service
	metrics metrics.MetricsProvider
}

// NewRepository constructs a session Repository.
func NewRepository(store db.Store, cache cache.Service, tracing tracing.Service, metrics metrics.MetricsProvider) Repository {
	return &sessionRepo{
		store:   store,
		cache:   cache,
		tracing: tracing,
		metrics: metrics,
	}
}

// CreateSession persists a new session row.
// The token hash (sha256hex) must be pre-computed by the caller — never pass the raw token.
func (r *sessionRepo) CreateSession(ctx context.Context, s Session) error {
	ctx, span := r.tracing.StartSpan(ctx, "session.repo.CreateSession")
	defer span.End()

	permsJSON, err := json.Marshal(s.Permissions)
	if err != nil {
		return fmt.Errorf("session repo: marshal permissions: %w", err)
	}

	entityScopeJSON := marshalJSON(s.EntityScope)
	configJSON := marshalJSON(s.Configuration)

	var principalID *uuid.UUID
	if s.PrincipalID != uuid.Nil {
		id := s.PrincipalID
		principalID = &id
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

// GetByTokenHash atomically touches last_accessed_at and retrieves the session.
// Returns nil, nil when not found or expired.
func (r *sessionRepo) GetByTokenHash(ctx context.Context, hash string) (*Session, error) {
	ctx, span := r.tracing.StartSpan(ctx, "session.repo.GetByTokenHash")
	defer span.End()

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

	return rowToSession(row.ID, row.UserID, row.TenantID, row.UserType, row.SessionToken,
		row.Permissions, row.EntityScope, row.Configuration,
		row.IpAddress, row.UserAgent, row.ExpiresAt, row.RiskScore, row.IsActive, row.LastAccessedAt)
}

// rowToSession converts raw sqlc scan results into a domain Session.
func rowToSession(
	id, userID, tenantID uuid.UUID,
	userType *string, sessionToken string,
	permsRaw, entityScopeRaw, configRaw []byte,
	ipAddr *netip.Addr,
	userAgent *string,
	expiresAt time.Time,
	riskScore *int32,
	isActivePt *bool,
	lastAccessed sql.NullTime,
) (*Session, error) {
	perms := make(map[string]bool)
	if len(permsRaw) > 0 {
		if err := json.Unmarshal(permsRaw, &perms); err != nil {
			return nil, fmt.Errorf("session repo: unmarshal permissions: %w", err)
		}
	}

	var scope EntityScope
	if len(entityScopeRaw) > 0 {
		_ = json.Unmarshal(entityScopeRaw, &scope)
	}

	var cfg Configuration
	if len(configRaw) > 0 {
		_ = json.Unmarshal(configRaw, &cfg)
	}
	if cfg.Flags == nil {
		cfg.Flags = map[string]bool{}
	}
	if cfg.Settings == nil {
		cfg.Settings = map[string]string{}
	}
	if cfg.Prefs == nil {
		cfg.Prefs = map[string]string{}
	}

	var utype string
	if userType != nil {
		utype = *userType
	}

	var ipStr string
	if ipAddr != nil {
		ipStr = ipAddr.String()
	}

	var agent string
	if userAgent != nil {
		agent = *userAgent
	}

	var isActive bool
	if isActivePt != nil {
		isActive = *isActivePt
	}

	return &Session{
		ID:            id,
		UserID:        userID,
		TenantID:      tenantID,
		UserType:      utype,
		TokenHash:     sessionToken,
		Permissions:   perms,
		PrincipalID:   uuid.Nil,
		EntityScope:   scope,
		Configuration: cfg,
		IsActive:      isActive,
		ExpiresAt:     expiresAt,
		LastSeenAt:    lastAccessed.Time,
		IPAddress:     ipStr,
		UserAgent:     agent,
		RiskScore:     int(derefInt32(riskScore)),
	}, nil
}

// Invalidate marks a session as inactive (logical delete).
func (r *sessionRepo) Invalidate(ctx context.Context, hash string) error {
	ctx, span := r.tracing.StartSpan(ctx, "session.repo.Invalidate")
	defer span.End()

	if err := r.store.InvalidateSession(ctx, hash); err != nil {
		return fmt.Errorf("session repo: invalidate session: %w", err)
	}
	return nil
}

// UpdateLastSeen fires an async last_accessed_at update — does not block the caller.
func (r *sessionRepo) UpdateLastSeen(ctx context.Context, hash string) {
	go func() {
		if err := r.store.UpdateSessionLastSeen(ctx, hash); err != nil {
			_ = err
		}
	}()
}

// sessionCacheKey returns the Redis key for a session hash.
func sessionCacheKey(hash string) string {
	return "session:" + hash
}

func (r *sessionRepo) cacheSession(ctx context.Context, hash string, resolved *ResolvedSession, ttl time.Duration) {
	_ = r.cache.Set(ctx, sessionCacheKey(hash), resolved, ttl)
}

func (r *sessionRepo) getCachedSession(ctx context.Context, hash string) *ResolvedSession {
	var resolved ResolvedSession
	if err := r.cache.Get(ctx, sessionCacheKey(hash), &resolved); err != nil {
		return nil
	}
	return &resolved
}

func (r *sessionRepo) deleteCachedSession(ctx context.Context, hash string) {
	_ = r.cache.Delete(ctx, sessionCacheKey(hash))
}

func derefInt32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}
