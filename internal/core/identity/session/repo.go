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
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// Repository defines the session persistence interface.
type Repository interface {
	CreateSession(ctx context.Context, s Session) error
	GetByTokenHash(ctx context.Context, hash string) (*Session, error)
	Invalidate(ctx context.Context, hash string) error
	UpdateLastSeen(ctx context.Context, hash string) // fire-and-forget
}

// sessionRepo implements Repository against user_sessions via SQLC.
//
// NOTE: Requires SQLC-generated methods from db/queries/sessions.sql.
// Run `make sqlc` before compiling (migration 000305 must be applied first).
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

	// Nullable principal_id — *uuid.UUID matches the generated SQLC type.
	var principalID *uuid.UUID
	if s.PrincipalID != uuid.Nil {
		id := s.PrincipalID
		principalID = &id
	}

	// INET column → *netip.Addr in generated SQLC code.
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

	riskScore := int32(s.RiskScore)

	if err := r.store.CreateSession(ctx, db.CreateSessionParams{
		UserID:       s.UserID,
		SessionToken: s.TokenHash,
		Permissions:  permsJSON,
		PrincipalID:  principalID,
		IpAddress:    ipAddr,
		UserAgent:    userAgent,
		ExpiresAt:    s.ExpiresAt,
		RiskScore:    &riskScore,
	}); err != nil {
		return fmt.Errorf("session repo: create session: %w", err)
	}
	return nil
}

// GetByTokenHash retrieves an active, non-expired session by its token hash.
// Returns nil, nil when not found.
func (r *sessionRepo) GetByTokenHash(ctx context.Context, hash string) (*Session, error) {
	ctx, span := r.tracing.StartSpan(ctx, "session.repo.GetByTokenHash")
	defer span.End()

	row, err := r.store.GetSessionByToken(ctx, hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // session not found or expired
		}
		return nil, fmt.Errorf("session repo: get session: %w", err)
	}
	if row == nil {
		return nil, nil
	}

	perms := make(map[string]bool)
	if len(row.Permissions) > 0 {
		if err := json.Unmarshal(row.Permissions, &perms); err != nil {
			return nil, fmt.Errorf("session repo: unmarshal permissions: %w", err)
		}
	}

	var principalID uuid.UUID
	if row.PrincipalID != nil {
		principalID = *row.PrincipalID
	}

	var ipStr string
	if row.IpAddress != nil {
		ipStr = row.IpAddress.String()
	}

	var userAgent string
	if row.UserAgent != nil {
		userAgent = *row.UserAgent
	}

	var isActive bool
	if row.IsActive != nil {
		isActive = *row.IsActive
	}

	return &Session{
		ID:          row.ID,
		UserID:      row.UserID,
		TenantID:    row.TenantID,
		TokenHash:   row.SessionToken,
		Permissions: perms,
		PrincipalID: principalID,
		IsActive:    isActive,
		ExpiresAt:   row.ExpiresAt,
		LastSeenAt:  row.LastAccessedAt.Time,
		IPAddress:   ipStr,
		UserAgent:   userAgent,
		RiskScore:   int(derefInt32(row.RiskScore)),
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
// Errors are swallowed intentionally; this is best-effort observability only.
func (r *sessionRepo) UpdateLastSeen(ctx context.Context, hash string) {
	go func() {
		// NOTE(tenant-context): ctx carries tenant_id for RLS.
		if err := r.store.UpdateSessionLastSeen(ctx, hash); err != nil {
			_ = err // best-effort: silently drop
		}
	}()
}

// sessionCacheKey returns the Redis key for a session hash.
func sessionCacheKey(hash string) string {
	return "session:" + hash
}

// cacheSession stores a ResolvedSession in the cache under the token hash.
func (r *sessionRepo) cacheSession(ctx context.Context, hash string, resolved *ResolvedSession, ttl time.Duration) {
	_ = r.cache.Set(ctx, sessionCacheKey(hash), resolved, ttl)
}

// getCachedSession attempts a cache lookup; returns nil on miss or error.
func (r *sessionRepo) getCachedSession(ctx context.Context, hash string) *ResolvedSession {
	var resolved ResolvedSession
	if err := r.cache.Get(ctx, sessionCacheKey(hash), &resolved); err != nil {
		return nil
	}
	return &resolved
}

// deleteCachedSession removes a session from the cache.
func (r *sessionRepo) deleteCachedSession(ctx context.Context, hash string) {
	_ = r.cache.Delete(ctx, sessionCacheKey(hash))
}

func derefInt32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}
