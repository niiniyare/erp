package cache

import (
	"context"
	"time"
)

// Service defines the interface for a cache service
type Service interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Flush(ctx context.Context) error
}
