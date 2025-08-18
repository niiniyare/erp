package cache

//go:generate go run go.uber.org/mock/mockgen -source=interface.go -destination=mock.go -package=cache

import (
	"context"
	"errors"
	"time"
)

// Cache errors
var (
	ErrCacheMiss = errors.New("cache miss")
)

// // Service defines the interface for a cache service
// type Service interface {
// 	Get(ctx context.Context, key string, dest any) error
// 	Set(ctx context.Context, key string, value any, expiration time.Duration) error
// 	Delete(ctx context.Context, key string) error
// 	Flush(ctx context.Context) error
// }

// service interface for additional features (optional)
type Service interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Flush(ctx context.Context) error

	// Bulk operations
	MGet(ctx context.Context, keys []string, dest any) error
	MSet(ctx context.Context, pairs map[string]any, expiration time.Duration) error
	MDelete(ctx context.Context, keys []string) error

	// Pattern operations
	DeletePattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Health and monitoring
	Ping(ctx context.Context) error
	Stats() *CacheStats
	Close() error
}
