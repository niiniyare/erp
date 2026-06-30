package tenant

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

const (
	// statusTTL is the Redis cache TTL for tenant status values.
	statusTTL = 5 * time.Minute

	// cacheKeyPrefix is prepended to all tenant-status cache keys.
	cacheKeyPrefix = "tenant:status:"
)

// RedisStatusService is a Redis-backed implementation of StatusService.
// Cache misses fall through to the provided loader function, which should
// query PostgreSQL. The loaded value is written back to Redis before returning.
//
// Cache invalidation: call SetStatus after any tenant status change —
// it overwrites the cached value immediately and resets the TTL.
type RedisStatusService struct {
	rdb    redis.UniversalClient
	loader func(ctx context.Context, tenantID uuid.UUID) (TenantStatus, error)
}

// NewRedisStatusService creates a RedisStatusService.
// loader is called on cache miss; it must return ErrTenantNotFound when
// the tenant does not exist in the database.
func NewRedisStatusService(
	rdb redis.UniversalClient,
	loader func(ctx context.Context, tenantID uuid.UUID) (TenantStatus, error),
) *RedisStatusService {
	return &RedisStatusService{rdb: rdb, loader: loader}
}

// GetStatus returns the tenant status, reading from Redis first.
func (s *RedisStatusService) GetStatus(ctx context.Context, tenantID uuid.UUID) (TenantStatus, error) {
	key := cacheKeyPrefix + tenantID.String()

	val, err := s.rdb.Get(ctx, key).Result()
	if err == nil {
		return TenantStatus(val), nil
	}
	if !errors.Is(err, redis.Nil) {
		// Redis error — fall through to loader rather than returning an error,
		// so a Redis hiccup doesn't block all authenticated requests.
		// The loader result is not cached to avoid a thundering herd on Redis restart.
		return s.loader(ctx, tenantID)
	}

	// Cache miss — load from DB.
	status, err := s.loader(ctx, tenantID)
	if err != nil {
		return "", err
	}

	// Best-effort cache write; ignore errors.
	_ = s.rdb.Set(ctx, key, string(status), statusTTL).Err()

	return status, nil
}

// SetStatus persists the new status in Redis immediately (write-through).
// It does NOT update PostgreSQL — the caller must persist to the database first.
// Typical call site: tenant admin handler → update DB → call SetStatus.
func (s *RedisStatusService) SetStatus(ctx context.Context, tenantID uuid.UUID, status TenantStatus) error {
	key := cacheKeyPrefix + tenantID.String()
	if err := s.rdb.Set(ctx, key, string(status), statusTTL).Err(); err != nil {
		return fmt.Errorf("tenant cache: SetStatus %s: %w", tenantID, err)
	}
	return nil
}

// Invalidate removes the cached status for tenantID, forcing the next
// GetStatus call to re-load from the database.
// Useful after manual database edits or during tests.
func (s *RedisStatusService) Invalidate(ctx context.Context, tenantID uuid.UUID) error {
	key := cacheKeyPrefix + tenantID.String()
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("tenant cache: Invalidate %s: %w", tenantID, err)
	}
	return nil
}

// compile-time check
var _ StatusService = (*RedisStatusService)(nil)
