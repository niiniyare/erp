package sdui

import (
	"context"
	"encoding/json"
	"fmt"

	"awo.so/framework/internal/core"
	"awo.so/framework/internal/redis"
)

// CachedBuilder wraps BuildPage with a Redis read-through cache.
type CachedBuilder struct {
	redis *redis.Client
}

// NewCachedBuilder returns a CachedBuilder backed by the given Redis client.
func NewCachedBuilder(r *redis.Client) *CachedBuilder {
	return &CachedBuilder{redis: r}
}

// GetOrBuild returns the cached page JSON for (tenant, entity, role),
// calling BuildPage on a cache miss and storing the result.
func (b *CachedBuilder) GetOrBuild(ctx context.Context, tenantID string, def *core.EntityDefinition, role string) (json.RawMessage, error) {
	cached, err := b.redis.GetPage(ctx, tenantID, def.Name, role)
	if err == nil {
		return json.RawMessage(cached), nil
	}
	if err != redis.ErrKeyNotFound {
		// Redis error — fall through to builder, log in production.
		_ = err
	}

	page, err := BuildPage(def)
	if err != nil {
		return nil, fmt.Errorf("sdui: build page: %w", err)
	}

	// Best-effort cache write — do not fail the request if Redis is down.
	_ = b.redis.SetPage(ctx, tenantID, def.Name, role, string(page))

	return page, nil
}

// Invalidate removes cached pages for the given entity across all roles.
func (b *CachedBuilder) Invalidate(ctx context.Context, tenantID, entityName string) error {
	return b.redis.InvalidateEntityPages(ctx, tenantID, entityName)
}
