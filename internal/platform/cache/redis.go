package cache

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/platform/config"
)

// Context keys for tenant information
type contextKey string

const (
	TenantIDKey    contextKey = "tenant_id"
	TenantSlugKey  contextKey = "tenant_slug"
	NamespaceKey   contextKey = "cache_namespace"
	CompressionKey contextKey = "cache_compression"

	// Cache configuration constants
	defaultKeyPrefix     = "erp"
	tenantKeyPrefix      = "tenant"
	globalKeyPrefix      = "global"
	compressionThreshold = 1024 // Compress values larger than 1KB
	maxKeyLength         = 250  // Redis key length limit
)

// // Service interface remains exactly the same for backward compatibility
// type Service interface {
// 	Get(ctx context.Context, key string, dest any) error
// 	Set(ctx context.Context, key string, value any, expiration time.Duration) error
// 	Delete(ctx context.Context, key string) error
// 	Flush(ctx context.Context) error
// }

// CacheStats provides cache metrics
type CacheStats struct {
	Hits              int64         `json:"hits"`
	Misses            int64         `json:"misses"`
	Sets              int64         `json:"sets"`
	Deletes           int64         `json:"deletes"`
	Errors            int64         `json:"errors"`
	HitRatio          float64       `json:"hit_ratio"`
	AverageLatency    time.Duration `json:"average_latency"`
	ConnectionsActive int           `json:"connections_active"`
	ConnectionsIdle   int           `json:"connections_idle"`
	LastError         string        `json:"last_error,omitempty"`
}

// RedisConfig extends the base config with additional options
type RedisConfig struct {
	*config.RedisConfig

	// Connection pool settings
	PoolSize     int           `yaml:"pool_size" default:"10"`
	MinIdleConns int           `yaml:"min_idle_conns" default:"3"`
	MaxConnAge   time.Duration `yaml:"max_conn_age" default:"30m"`
	PoolTimeout  time.Duration `yaml:"pool_timeout" default:"4s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" default:"5m"`

	// Performance settings
	EnableCompression bool          `yaml:"enable_compression" default:"true"`
	CompressionLevel  int           `yaml:"compression_level" default:"6"`
	KeyPrefix         string        `yaml:"key_prefix" default:"erp"`
	MaxRetries        int           `yaml:"max_retries" default:"3"`
	RetryDelay        time.Duration `yaml:"retry_delay" default:"100ms"`

	// Circuit breaker settings
	EnableCircuitBreaker    bool          `yaml:"enable_circuit_breaker" default:"true"`
	CircuitBreakerThreshold int           `yaml:"circuit_breaker_threshold" default:"5"`
	CircuitBreakerTimeout   time.Duration `yaml:"circuit_breaker_timeout" default:"30s"`
}

// DefaultRedisConfig returns production-ready defaults
func DefaultRedisConfig(baseConfig *config.RedisConfig) *RedisConfig {
	return &RedisConfig{
		RedisConfig:             baseConfig,
		PoolSize:                10,
		MinIdleConns:            3,
		MaxConnAge:              30 * time.Minute,
		PoolTimeout:             4 * time.Second,
		IdleTimeout:             5 * time.Minute,
		EnableCompression:       true,
		CompressionLevel:        6,
		KeyPrefix:               "erp",
		MaxRetries:              3,
		RetryDelay:              100 * time.Millisecond,
		EnableCircuitBreaker:    true,
		CircuitBreakerThreshold: 5,
		CircuitBreakerTimeout:   30 * time.Second,
	}
}

// redisClient implements the Service interface with multi-tenant support
type redisClient struct {
	client             *redis.Client
	config             *RedisConfig
	stats              *CacheStats
	statsMutex         sync.RWMutex
	circuitBreaker     *CircuitBreaker
	keyHashingEnabled  bool
	compressionEnabled bool
}

// // NewRedisClient creates a new Redis client with the original signature
//
//	func NewRedisClient(cfg *config.RedisConfig) Service {
//		Config := DefaultRedisConfig(cfg)
//		return NewRedisClient(Config)
//	}
//
// NewRedisClient creates a new Redis client with  configuration
func NewRedisClient(cfg *RedisConfig) Service {
	// Create Redis client with  options
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxConnAge:   cfg.MaxConnAge,
		PoolTimeout:  cfg.PoolTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		MaxRetries:   cfg.MaxRetries,
		// retry_delay:   cfg.RetryDelay,
	})

	client := &redisClient{
		client:             rdb,
		config:             cfg,
		stats:              &CacheStats{},
		keyHashingEnabled:  true,
		compressionEnabled: cfg.EnableCompression,
	}

	// Initialize circuit breaker if enabled
	if cfg.EnableCircuitBreaker {
		client.circuitBreaker = NewCircuitBreaker(
			cfg.CircuitBreakerThreshold,
			cfg.CircuitBreakerTimeout,
		)
	}

	return client
}

// Context helper functions for multi-tenant support
func SetTenantInContext(ctx context.Context, tenantID uuid.UUID, tenantSlug string) context.Context {
	ctx = context.WithValue(ctx, TenantIDKey, tenantID)
	ctx = context.WithValue(ctx, TenantSlugKey, tenantSlug)
	return ctx
}

func SetNamespaceInContext(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, NamespaceKey, namespace)
}

func SetCompressionInContext(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, CompressionKey, enabled)
}

// getTenantFromContext extracts tenant information from context
func (r *redisClient) getTenantFromContext(ctx context.Context) (tenantID uuid.UUID, tenantSlug string) {
	if id, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok {
		tenantID = id
	}
	if slug, ok := ctx.Value(TenantSlugKey).(string); ok {
		tenantSlug = slug
	}
	return
}

// buildKey constructs a tenant-aware cache key
func (r *redisClient) buildKey(ctx context.Context, key string) string {
	parts := []string{r.config.KeyPrefix}

	// Add tenant information if available
	tenantID, tenantSlug := r.getTenantFromContext(ctx)
	if tenantID != uuid.Nil {
		if tenantSlug != "" {
			parts = append(parts, tenantKeyPrefix, tenantSlug)
		} else {
			parts = append(parts, tenantKeyPrefix, tenantID.String())
		}
	} else {
		parts = append(parts, globalKeyPrefix)
	}

	// Add custom namespace if provided
	if namespace, ok := ctx.Value(NamespaceKey).(string); ok && namespace != "" {
		parts = append(parts, namespace)
	}

	// Add the actual key
	parts = append(parts, key)

	finalKey := strings.Join(parts, ":")

	// Hash key if it's too long
	if r.keyHashingEnabled && len(finalKey) > maxKeyLength {
		hash := sha256.Sum256([]byte(finalKey))
		hashedKey := hex.EncodeToString(hash[:])
		// Keep a readable prefix
		prefix := strings.Join(parts[:len(parts)-1], ":")
		if len(prefix) > maxKeyLength-65 { // 64 chars for hash + 1 for ":"
			prefix = prefix[:maxKeyLength-65]
		}
		finalKey = prefix + ":" + hashedKey
	}

	return finalKey
}

// shouldCompress determines if a value should be compressed
func (r *redisClient) shouldCompress(ctx context.Context, data []byte) bool {
	if !r.compressionEnabled {
		return false
	}

	// Check context override
	if enabled, ok := ctx.Value(CompressionKey).(bool); ok {
		return enabled && len(data) > compressionThreshold
	}

	return len(data) > compressionThreshold
}

// compressData compresses data using gzip
func (r *redisClient) compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, r.config.CompressionLevel)
	if err != nil {
		return nil, err
	}

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	// Add compression marker
	compressed := append([]byte("GZIP:"), buf.Bytes()...)
	return compressed, nil
}

// decompressData decompresses gzip data
func (r *redisClient) decompressData(data []byte) ([]byte, error) {
	// Check for compression marker
	if !bytes.HasPrefix(data, []byte("GZIP:")) {
		return data, nil // Not compressed
	}

	// Remove compression marker
	compressedData := data[5:]

	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// serializeValue serializes and optionally compresses a value
func (r *redisClient) serializeValue(ctx context.Context, value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	if r.shouldCompress(ctx, data) {
		return r.compressData(data)
	}

	return data, nil
}

// deserializeValue deserializes and optionally decompresses a value
func (r *redisClient) deserializeValue(data []byte, dest any) error {
	decompressed, err := r.decompressData(data)
	if err != nil {
		return err
	}

	return json.Unmarshal(decompressed, dest)
}

// executeWithCircuitBreaker executes a Redis operation with circuit breaker protection
func (r *redisClient) executeWithCircuitBreaker(ctx context.Context, operation func() error) error {
	if r.circuitBreaker == nil {
		return operation()
	}

	return r.circuitBreaker.Execute(operation)
}

// updateStats updates cache statistics
func (r *redisClient) updateStats(operation string, err error, startTime time.Time) {
	r.statsMutex.Lock()
	defer r.statsMutex.Unlock()

	latency := time.Since(startTime)

	switch operation {
	case "get":
		if err == ErrCacheMiss {
			r.stats.Misses++
		} else if err == nil {
			r.stats.Hits++
		}
	case "set":
		if err == nil {
			r.stats.Sets++
		}
	case "delete":
		if err == nil {
			r.stats.Deletes++
		}
	}

	if err != nil && err != ErrCacheMiss {
		r.stats.Errors++
		r.stats.LastError = err.Error()
	}

	// Update average latency (simple moving average)
	totalOps := r.stats.Hits + r.stats.Misses + r.stats.Sets + r.stats.Deletes
	if totalOps > 0 {
		r.stats.AverageLatency = time.Duration(
			(int64(r.stats.AverageLatency)*int64(totalOps-1) + int64(latency)) / int64(totalOps),
		)
	}

	// Update hit ratio
	if r.stats.Hits+r.stats.Misses > 0 {
		r.stats.HitRatio = float64(r.stats.Hits) / float64(r.stats.Hits+r.stats.Misses)
	}
}

// Get retrieves a value from cache (maintains original signature)
func (r *redisClient) Get(ctx context.Context, key string, dest any) error {
	startTime := time.Now()

	var err error
	err = r.executeWithCircuitBreaker(ctx, func() error {
		finalKey := r.buildKey(ctx, key)
		val, getErr := r.client.Get(ctx, finalKey).Bytes()
		if getErr != nil {
			if getErr == redis.Nil {
				return ErrCacheMiss
			}
			return getErr
		}

		return r.deserializeValue(val, dest)
	})

	r.updateStats("get", err, startTime)
	return err
}

// Set stores a value in cache (maintains original signature)
func (r *redisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	startTime := time.Now()

	var err error
	err = r.executeWithCircuitBreaker(ctx, func() error {
		data, serErr := r.serializeValue(ctx, value)
		if serErr != nil {
			return serErr
		}

		finalKey := r.buildKey(ctx, key)
		return r.client.Set(ctx, finalKey, data, expiration).Err()
	})

	r.updateStats("set", err, startTime)
	return err
}

// Delete removes a value from cache (maintains original signature)
func (r *redisClient) Delete(ctx context.Context, key string) error {
	startTime := time.Now()

	var err error
	err = r.executeWithCircuitBreaker(ctx, func() error {
		finalKey := r.buildKey(ctx, key)
		return r.client.Del(ctx, finalKey).Err()
	})

	r.updateStats("delete", err, startTime)
	return err
}

// Flush clears cache (maintains original signature)
// Note: This respects tenant isolation - only flushes current tenant's data
func (r *redisClient) Flush(ctx context.Context) error {
	tenantID, tenantSlug := r.getTenantFromContext(ctx)

	// If no tenant context, flush the entire database (global operation)
	if tenantID == uuid.Nil {
		return r.client.FlushDB(ctx).Err()
	}

	// For tenant-specific flush, delete by pattern
	var pattern string
	if tenantSlug != "" {
		pattern = fmt.Sprintf("%s:%s:%s:*", r.config.KeyPrefix, tenantKeyPrefix, tenantSlug)
	} else {
		pattern = fmt.Sprintf("%s:%s:%s:*", r.config.KeyPrefix, tenantKeyPrefix, tenantID.String())
	}

	return r.deleteByPattern(ctx, pattern)
}

//  methods (additional functionality)

// MGet retrieves multiple values
func (r *redisClient) MGet(ctx context.Context, keys []string, dest any) error {
	if len(keys) == 0 {
		return nil
	}

	finalKeys := make([]string, len(keys))
	for i, key := range keys {
		finalKeys[i] = r.buildKey(ctx, key)
	}

	return r.executeWithCircuitBreaker(ctx, func() error {
		vals, err := r.client.MGet(ctx, finalKeys...).Result()
		if err != nil {
			return err
		}

		results := make([]any, len(vals))
		for i, val := range vals {
			if val != nil {
				var decoded any
				if strVal, ok := val.(string); ok {
					if err := r.deserializeValue([]byte(strVal), &decoded); err != nil {
						return err
					}
					results[i] = decoded
				}
			}
		}

		return json.Unmarshal([]byte(fmt.Sprintf("%v", results)), dest)
	})
}

// MSet sets multiple values
func (r *redisClient) MSet(ctx context.Context, pairs map[string]any, expiration time.Duration) error {
	if len(pairs) == 0 {
		return nil
	}

	return r.executeWithCircuitBreaker(ctx, func() error {
		pipe := r.client.Pipeline()

		for key, value := range pairs {
			data, err := r.serializeValue(ctx, value)
			if err != nil {
				return err
			}

			finalKey := r.buildKey(ctx, key)
			pipe.Set(ctx, finalKey, data, expiration)
		}

		_, err := pipe.Exec(ctx)
		return err
	})
}

// MDelete deletes multiple keys
func (r *redisClient) MDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	finalKeys := make([]string, len(keys))
	for i, key := range keys {
		finalKeys[i] = r.buildKey(ctx, key)
	}

	return r.executeWithCircuitBreaker(ctx, func() error {
		return r.client.Del(ctx, finalKeys...).Err()
	})
}

// DeletePattern deletes keys matching a pattern
func (r *redisClient) DeletePattern(ctx context.Context, pattern string) error {
	finalPattern := r.buildKey(ctx, pattern)
	return r.deleteByPattern(ctx, finalPattern)
}

// deleteByPattern internal method to delete by pattern
func (r *redisClient) deleteByPattern(ctx context.Context, pattern string) error {
	return r.executeWithCircuitBreaker(ctx, func() error {
		iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
		var keys []string

		for iter.Next(ctx) {
			keys = append(keys, iter.Val())

			// Delete in batches to avoid memory issues
			if len(keys) >= 1000 {
				if err := r.client.Del(ctx, keys...).Err(); err != nil {
					return err
				}
				keys = keys[:0]
			}
		}

		// Delete remaining keys
		if len(keys) > 0 {
			return r.client.Del(ctx, keys...).Err()
		}

		return iter.Err()
	})
}

// Exists checks if a key exists
func (r *redisClient) Exists(ctx context.Context, key string) (bool, error) {
	var exists bool
	err := r.executeWithCircuitBreaker(ctx, func() error {
		finalKey := r.buildKey(ctx, key)
		count, err := r.client.Exists(ctx, finalKey).Result()
		exists = count > 0
		return err
	})
	return exists, err
}

// TTL returns the time to live for a key
func (r *redisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	var ttl time.Duration
	err := r.executeWithCircuitBreaker(ctx, func() error {
		finalKey := r.buildKey(ctx, key)
		result, err := r.client.TTL(ctx, finalKey).Result()
		ttl = result
		return err
	})
	return ttl, err
}

// Ping tests the connection
func (r *redisClient) Ping(ctx context.Context) error {
	return r.executeWithCircuitBreaker(ctx, func() error {
		return r.client.Ping(ctx).Err()
	})
}

// Stats returns cache statistics
func (r *redisClient) Stats() *CacheStats {
	r.statsMutex.RLock()
	defer r.statsMutex.RUnlock()

	// Get connection pool stats
	poolStats := r.client.PoolStats()

	stats := *r.stats // Copy
	stats.ConnectionsActive = int(poolStats.TotalConns - poolStats.IdleConns)
	stats.ConnectionsIdle = int(poolStats.IdleConns)

	return &stats
}

// Close closes the Redis connection
func (r *redisClient) Close() error {
	return r.client.Close()
}

// Circuit breaker implementation
type CircuitBreaker struct {
	threshold    int
	timeout      time.Duration
	failureCount int
	lastFailure  time.Time
	state        string // "closed", "open", "half-open"
	mutex        sync.RWMutex
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
		timeout:   timeout,
		state:     "closed",
	}
}

func (cb *CircuitBreaker) Execute(operation func() error) error {
	cb.mutex.RLock()
	state := cb.state
	cb.mutex.RUnlock()

	if state == "open" {
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.mutex.Lock()
			cb.state = "half-open"
			cb.mutex.Unlock()
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	}

	err := operation()

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if err != nil {
		cb.failureCount++
		cb.lastFailure = time.Now()

		if cb.failureCount >= cb.threshold {
			cb.state = "open"
		}
	} else {
		cb.failureCount = 0
		cb.state = "closed"
	}

	return err
}

// Helper functions for easy tenant context setup in your application

// WithTenant creates a context with tenant information
func WithTenant(ctx context.Context, tenantID uuid.UUID, tenantSlug string) context.Context {
	return SetTenantInContext(ctx, tenantID, tenantSlug)
}

// WithNamespace creates a context with a custom namespace
func WithNamespace(ctx context.Context, namespace string) context.Context {
	return SetNamespaceInContext(ctx, namespace)
}

// WithCompression creates a context with compression settings
func WithCompression(ctx context.Context, enabled bool) context.Context {
	return SetCompressionInContext(ctx, enabled)
}

// WithTenantAndNamespace creates a context with both tenant and namespace
func WithTenantAndNamespace(ctx context.Context, tenantID uuid.UUID, tenantSlug, namespace string) context.Context {
	ctx = SetTenantInContext(ctx, tenantID, tenantSlug)
	ctx = SetNamespaceInContext(ctx, namespace)
	return ctx
}
