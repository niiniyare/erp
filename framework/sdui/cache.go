package sdui

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// CacheTTL is the Redis TTL for all cached SDUI page schemas.
	// Invalidate explicitly via InvalidateEntitySchemas on permission/flag change.
	CacheTTL = 5 * time.Minute

	// cacheKeyPrefix distinguishes SDUI keys from session and feature-flag keys.
	cacheKeyPrefix = "page:"
)

// Cache is a minimal Redis interface for schema storage.
// Satisfied by *redis.Client, *redis.ClusterClient, and miniredis in tests.
type Cache interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

// CacheGet retrieves a cached schema byte slice.
// Returns (nil, nil) on a cache miss so callers can distinguish miss from error.
func CacheGet(ctx context.Context, c Cache, key string) ([]byte, error) {
	b, err := c.Get(ctx, key).Bytes()
	if err == nil {
		return b, nil
	}
	if errors.Is(err, redis.Nil) {
		return nil, nil // miss
	}
	return nil, fmt.Errorf("sdui cache get %q: %w", key, err)
}

// CacheSet stores a schema byte slice with the standard TTL.
// Best-effort: callers should log but not fail on error.
func CacheSet(ctx context.Context, c Cache, key string, b []byte) error {
	if err := c.Set(ctx, key, b, CacheTTL).Err(); err != nil {
		return fmt.Errorf("sdui cache set %q: %w", key, err)
	}
	return nil
}

// InvalidateEntitySchemas removes all cached schemas for entity across the
// four view types (crud, form_create, form_edit, detail).
// Pass uuid.Nil as tenantID to invalidate tenant-agnostic (no custom fields) schemas.
// Call after: permission changes, feature flag changes, entity definition reload.
func InvalidateEntitySchemas(ctx context.Context, c Cache, entityHash, tenantID string) error {
	keys := []string{
		buildKey(entityHash, "crud", tenantID),
		buildKey(entityHash, "form_create", tenantID),
		buildKey(entityHash, "form_edit", tenantID),
		buildKey(entityHash, "detail", tenantID),
	}
	if err := c.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("sdui cache invalidate entity %q: %w", entityHash, err)
	}
	return nil
}

// buildKey constructs the Redis key for a schema.
// Format: page:{entityHash}:{viewType}:{tenantID}
// entityHash is the first 8 bytes of sha256(entityName) encoded as hex.
func buildKey(entityHash, viewType, tenantID string) string {
	return fmt.Sprintf("%s%s:%s:%s", cacheKeyPrefix, entityHash, viewType, tenantID)
}
