package cache

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

// import (
// 	"bytes"
// 	"compress/gzip"
// 	"context"
// 	"crypto/sha256"
// 	"encoding/hex"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"strings"
// 	"sync"
// 	"sync/atomic"
// 	"time"
//
// 	"github.com/go-redis/redis/v8"
// 	"github.com/google/uuid"
// 	"github.com/niiniyare/erp/internal/platform/config"
// )
//
// // Context keys for tenant information.
// type contextKey string
//
// const (
// 	// TenantIDKey is the context key for tenant UUID.
// 	TenantIDKey contextKey = "tenant_id"
// 	// TenantSlugKey is the context key for tenant slug.
// 	TenantSlugKey contextKey = "tenant_slug"
// 	// TenantSubdomainKey is the context key for tenant subdomain.
// 	TenantSubdomainKey contextKey = "tenant_subdomain"
// 	// NamespaceKey is the context key for cache namespace.
// 	NamespaceKey contextKey = "cache_namespace"
// 	// CompressionKey is the context key for compression override.
// 	CompressionKey contextKey = "cache_compression"
//
// 	// Cache configuration constants.
// 	defaultKeyPrefix     = "erp"
// 	tenantKeyPrefix      = "tenant"
// 	globalKeyPrefix      = "global"
// 	compressionThreshold = 1024 // Compress values larger than 1KB
// 	maxKeyLength         = 250  // Redis key length limit
// 	compressionMarker    = 0x1F // Unit separator - unlikely in normal data
// 	batchDeleteSize      = 1000 // Keys to delete per batch
// )
//
// // Standard errors.
// var (
// 	// ErrCacheMiss indicates the requested key was not found.
// 	ErrCacheMiss = errors.New("cache: key not found")
// 	// ErrCircuitOpen indicates the circuit breaker is open.
// 	ErrCircuitOpen = errors.New("cache: circuit breaker open")
// 	// ErrInvalidData indicates corrupted or invalid cached data.
// 	ErrInvalidData = errors.New("cache: invalid data format")
// 	// ErrNilValue indicates a nil value was provided.
// 	ErrNilValue = errors.New("cache: nil value provided")
// 	// ErrNoTenantContext indicates tenant information is missing from context.
// 	ErrNoTenantContext = errors.New("cache: tenant context required - missing tenant_id, tenant_slug, or tenant_subdomain")
// )
//
// // Service defines the interface for a cache service.
// // All methods are safe for concurrent use and require tenant context.
// type Service interface {
// 	// Core operations
// 	Get(ctx context.Context, key string, dest any) error
// 	Set(ctx context.Context, key string, value any, expiration time.Duration) error
// 	Delete(ctx context.Context, key string) error
// 	Flush(ctx context.Context) error
//
// 	// Bulk operations
// 	MGet(ctx context.Context, keys []string) ([]Result, error)
// 	MSet(ctx context.Context, pairs map[string]any, expiration time.Duration) error
// 	MDelete(ctx context.Context, keys []string) error
//
// 	// Pattern operations
// 	DeletePattern(ctx context.Context, pattern string) error
// 	Keys(ctx context.Context, pattern string) ([]string, error)
// 	Exists(ctx context.Context, key string) (bool, error)
// 	TTL(ctx context.Context, key string) (time.Duration, error)
// 	Expire(ctx context.Context, key string, expiration time.Duration) error
//
// 	// In-memory cache operations (tenant-aware)
// 	GetMemory(ctx context.Context, key string, dest any) error
// 	SetMemory(ctx context.Context, key string, value any, expiration time.Duration) error
// 	DeleteMemory(ctx context.Context, key string) error
//
// 	// Global in-memory cache operations (not tenant-aware, for shared data like formulas)
// 	GetGlobalMemory(key string, dest any) error
// 	SetGlobalMemory(key string, value any, expiration time.Duration) error
// 	DeleteGlobalMemory(key string) error
//
// 	// Health and monitoring
// 	Ping(ctx context.Context) error
// 	Stats() CacheStats
// 	Reset() // Reset statistics
// 	Close() error
// }
//
// // Result represents a single cache get result.
// type Result struct {
// 	Key   string
// 	Value any
// 	Err   error
// }
//
// // CacheStats provides cache metrics.
// type CacheStats struct {
// 	Hits              int64         `json:"hits"`
// 	Misses            int64         `json:"misses"`
// 	Sets              int64         `json:"sets"`
// 	Deletes           int64         `json:"deletes"`
// 	Errors            int64         `json:"errors"`
// 	HitRatio          float64       `json:"hit_ratio"`
// 	AverageLatency    time.Duration `json:"average_latency"`
// 	ConnectionsActive int           `json:"connections_active"`
// 	ConnectionsIdle   int           `json:"connections_idle"`
// 	MemoryCacheSize   int           `json:"memory_cache_size"`
// 	GlobalCacheSize   int           `json:"global_cache_size"`
// 	LastError         string        `json:"last_error,omitempty"`
// 	LastErrorTime     time.Time     `json:"last_error_time,omitempty"`
// }
//
// // RedisConfig extends the base config with advanced options.
// type RedisConfig struct {
// 	*config.RedisConfig
//
// 	// Connection pool settings
// 	PoolSize     int           `yaml:"pool_size"`
// 	MinIdleConns int           `yaml:"min_idle_conns"`
// 	MaxConnAge   time.Duration `yaml:"max_conn_age"`
// 	PoolTimeout  time.Duration `yaml:"pool_timeout"`
// 	IdleTimeout  time.Duration `yaml:"idle_timeout"`
//
// 	// Performance settings
// 	EnableCompression bool          `yaml:"enable_compression"`
// 	CompressionLevel  int           `yaml:"compression_level"` // 1-9, default 6
// 	KeyPrefix         string        `yaml:"key_prefix"`
// 	MaxRetries        int           `yaml:"max_retries"`
// 	RetryDelay        time.Duration `yaml:"retry_delay"`
// 	BatchDeleteSize   int           `yaml:"batch_delete_size"`
//
// 	// Circuit breaker settings
// 	EnableCircuitBreaker    bool          `yaml:"enable_circuit_breaker"`
// 	CircuitBreakerThreshold int           `yaml:"circuit_breaker_threshold"`
// 	CircuitBreakerTimeout   time.Duration `yaml:"circuit_breaker_timeout"`
//
// 	// Tenant settings
// 	RequireTenantContext  bool `yaml:"require_tenant_context"`  // Enforce tenant context on all operations
// 	AllowGlobalOperations bool `yaml:"allow_global_operations"` // Allow operations without tenant context
//
// 	// In-memory cache settings
// 	EnableMemoryCache          bool          `yaml:"enable_memory_cache"`
// 	MemoryCacheMaxSize         int           `yaml:"memory_cache_max_size"` // Max items in memory cache
// 	MemoryCacheDefaultTTL      time.Duration `yaml:"memory_cache_default_ttl"`
// 	MemoryCacheCleanupInterval time.Duration `yaml:"memory_cache_cleanup_interval"`
// }
//
// // DefaultRedisConfig returns production-ready defaults.
// func DefaultRedisConfig(baseConfig *config.RedisConfig) *RedisConfig {
// 	if baseConfig == nil {
// 		panic("cache: base config cannot be nil")
// 	}
//
// 	return &RedisConfig{
// 		RedisConfig:                baseConfig,
// 		PoolSize:                   10,
// 		MinIdleConns:               3,
// 		MaxConnAge:                 30 * time.Minute,
// 		PoolTimeout:                4 * time.Second,
// 		IdleTimeout:                5 * time.Minute,
// 		EnableCompression:          true,
// 		CompressionLevel:           6,
// 		KeyPrefix:                  defaultKeyPrefix,
// 		MaxRetries:                 3,
// 		RetryDelay:                 100 * time.Millisecond,
// 		BatchDeleteSize:            batchDeleteSize,
// 		EnableCircuitBreaker:       true,
// 		CircuitBreakerThreshold:    5,
// 		CircuitBreakerTimeout:      30 * time.Second,
// 		RequireTenantContext:       true,  // Enforce tenant context by default
// 		AllowGlobalOperations:      false, // Disallow global operations by default
// 		EnableMemoryCache:          true,
// 		MemoryCacheMaxSize:         1000,
// 		MemoryCacheDefaultTTL:      5 * time.Minute,
// 		MemoryCacheCleanupInterval: 1 * time.Minute,
// 	}
// }
//
// // Validate checks if the configuration is valid.
// func (c *RedisConfig) Validate() error {
// 	if c.RedisConfig == nil {
// 		return errors.New("cache: base redis config is required")
// 	}
// 	if c.PoolSize < 1 {
// 		return errors.New("cache: pool size must be at least 1")
// 	}
// 	if c.CompressionLevel < 1 || c.CompressionLevel > 9 {
// 		return errors.New("cache: compression level must be between 1 and 9")
// 	}
// 	if c.KeyPrefix == "" {
// 		return errors.New("cache: key prefix cannot be empty")
// 	}
// 	if c.BatchDeleteSize < 1 {
// 		return errors.New("cache: batch delete size must be at least 1")
// 	}
// 	if c.EnableMemoryCache && c.MemoryCacheMaxSize < 1 {
// 		return errors.New("cache: memory cache max size must be at least 1")
// 	}
// 	return nil
// }
//
// // memoryEntry represents a cached item in memory.
// type memoryEntry struct {
// 	value     any
// 	expiresAt time.Time
// }
//
// // redisClient implements the Service interface with multi-tenant support.
// type redisClient struct {
// 	client             *redis.Client
// 	config             *RedisConfig
// 	circuitBreaker     *circuitBreaker
// 	keyHashingEnabled  bool
// 	compressionEnabled bool
//
// 	// In-memory caches
// 	memoryCache       map[string]*memoryEntry // Tenant-aware memory cache
// 	globalMemoryCache map[string]*memoryEntry // Global memory cache (formulas, etc.)
// 	memoryCacheMu     sync.RWMutex
// 	globalCacheMu     sync.RWMutex
//
// 	// Statistics (use atomic operations for thread-safety)
// 	hits      atomic.Int64
// 	misses    atomic.Int64
// 	sets      atomic.Int64
// 	deletes   atomic.Int64
// 	errors    atomic.Int64
// 	latencyNs atomic.Int64
// 	opsCount  atomic.Int64
//
// 	lastError     string
// 	lastErrorTime time.Time
// 	statsMutex    sync.RWMutex
//
// 	stopCleanup chan struct{}
// 	cleanupWg   sync.WaitGroup
// }
//
// // NewRedisClient creates a new Redis client with enhanced configuration.
// func NewRedisClient(cfg *RedisConfig) (Service, error) {
// 	if err := cfg.Validate(); err != nil {
// 		return nil, fmt.Errorf("cache: invalid config: %w", err)
// 	}
//
// 	rdb := redis.NewClient(&redis.Options{
// 		Addr:            fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
// 		Password:        cfg.Password,
// 		DB:              cfg.DB,
// 		PoolSize:        cfg.PoolSize,
// 		MinIdleConns:    cfg.MinIdleConns,
// 		MaxConnAge:      cfg.MaxConnAge,
// 		PoolTimeout:     cfg.PoolTimeout,
// 		IdleTimeout:     cfg.IdleTimeout,
// 		MaxRetries:      cfg.MaxRetries,
// 		MinRetryBackoff: cfg.RetryDelay,
// 		MaxRetryBackoff: cfg.RetryDelay * 5,
// 	})
//
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
//
// 	if err := rdb.Ping(ctx).Err(); err != nil {
// 		rdb.Close()
// 		return nil, fmt.Errorf("cache: failed to connect to redis: %w", err)
// 	}
//
// 	client := &redisClient{
// 		client:             rdb,
// 		config:             cfg,
// 		keyHashingEnabled:  true,
// 		compressionEnabled: cfg.EnableCompression,
// 		stopCleanup:        make(chan struct{}),
// 	}
//
// 	if cfg.EnableMemoryCache {
// 		client.memoryCache = make(map[string]*memoryEntry)
// 		client.globalMemoryCache = make(map[string]*memoryEntry)
// 		client.startCleanupRoutine()
// 	}
//
// 	if cfg.EnableCircuitBreaker {
// 		client.circuitBreaker = newCircuitBreaker(
// 			cfg.CircuitBreakerThreshold,
// 			cfg.CircuitBreakerTimeout,
// 		)
// 	}
//
// 	return client, nil
// }
//
// // NewRedisClientMust is like NewRedisClient but panics on error.
// func NewRedisClientMust(cfg *RedisConfig) Service {
// 	client, err := NewRedisClient(cfg)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return client
// }
//
// // startCleanupRoutine starts a background goroutine to clean up expired entries.
// func (r *redisClient) startCleanupRoutine() {
// 	r.cleanupWg.Add(1)
// 	go func() {
// 		defer r.cleanupWg.Done()
// 		ticker := time.NewTicker(r.config.MemoryCacheCleanupInterval)
// 		defer ticker.Stop()
//
// 		for {
// 			select {
// 			case <-ticker.C:
// 				r.cleanupExpiredEntries()
// 			case <-r.stopCleanup:
// 				return
// 			}
// 		}
// 	}()
// }
//
// // cleanupExpiredEntries removes expired entries from memory caches.
// func (r *redisClient) cleanupExpiredEntries() {
// 	now := time.Now()
//
// 	// Clean tenant-aware cache
// 	r.memoryCacheMu.Lock()
// 	for key, entry := range r.memoryCache {
// 		if !entry.expiresAt.IsZero() && entry.expiresAt.Before(now) {
// 			delete(r.memoryCache, key)
// 		}
// 	}
// 	r.memoryCacheMu.Unlock()
//
// 	// Clean global cache
// 	r.globalCacheMu.Lock()
// 	for key, entry := range r.globalMemoryCache {
// 		if !entry.expiresAt.IsZero() && entry.expiresAt.Before(now) {
// 			delete(r.globalMemoryCache, key)
// 		}
// 	}
// 	r.globalCacheMu.Unlock()
// }
//
// // getTenantInfo extracts tenant information from context.
// // Returns the first available identifier in priority order: ID, slug, subdomain.
// func (r *redisClient) getTenantInfo(ctx context.Context) (identifier string, identifierType string, err error) {
// 	// Check for tenant ID first
// 	if id, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok && id != uuid.Nil {
// 		return id.String(), "id", nil
// 	}
//
// 	// Check for tenant slug
// 	if slug, ok := ctx.Value(TenantSlugKey).(string); ok && slug != "" {
// 		return slug, "slug", nil
// 	}
//
// 	// Check for tenant subdomain
// 	if subdomain, ok := ctx.Value(TenantSubdomainKey).(string); ok && subdomain != "" {
// 		return subdomain, "subdomain", nil
// 	}
//
// 	// No tenant information found
// 	if r.config.RequireTenantContext && !r.config.AllowGlobalOperations {
// 		return "", "", ErrNoTenantContext
// 	}
//
// 	return "", "", nil
// }
//
// // buildKey constructs a tenant-aware cache key.
// func (r *redisClient) buildKey(ctx context.Context, key string) (string, error) {
// 	if key == "" {
// 		return "", errors.New("cache: key cannot be empty")
// 	}
//
// 	parts := []string{r.config.KeyPrefix}
//
// 	// Get tenant information
// 	tenantID, idType, err := r.getTenantInfo(ctx)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	if tenantID != "" {
// 		parts = append(parts, tenantKeyPrefix, fmt.Sprintf("%s-%s", idType, tenantID))
// 	} else if r.config.AllowGlobalOperations {
// 		parts = append(parts, globalKeyPrefix)
// 	} else {
// 		return "", ErrNoTenantContext
// 	}
//
// 	// Add custom namespace if provided
// 	if namespace, ok := ctx.Value(NamespaceKey).(string); ok && namespace != "" {
// 		parts = append(parts, namespace)
// 	}
//
// 	parts = append(parts, key)
// 	finalKey := strings.Join(parts, ":")
//
// 	// Hash key if it exceeds Redis limits
// 	if r.keyHashingEnabled && len(finalKey) > maxKeyLength {
// 		hash := sha256.Sum256([]byte(finalKey))
// 		hashedKey := hex.EncodeToString(hash[:])
// 		prefix := strings.Join(parts[:len(parts)-1], ":")
// 		maxPrefixLen := maxKeyLength - 65
// 		if len(prefix) > maxPrefixLen {
// 			prefix = prefix[:maxPrefixLen]
// 		}
// 		finalKey = prefix + ":" + hashedKey
// 	}
//
// 	return finalKey, nil
// }
//
// // buildMemoryKey constructs a tenant-aware memory cache key.
// func (r *redisClient) buildMemoryKey(ctx context.Context, key string) (string, error) {
// 	tenantID, idType, err := r.getTenantInfo(ctx)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	if tenantID == "" {
// 		return "", ErrNoTenantContext
// 	}
//
// 	return fmt.Sprintf("%s-%s:%s", idType, tenantID, key), nil
// }
//
// // shouldCompress determines if a value should be compressed.
// func (r *redisClient) shouldCompress(ctx context.Context, data []byte) bool {
// 	if !r.compressionEnabled {
// 		return false
// 	}
//
// 	if enabled, ok := ctx.Value(CompressionKey).(bool); ok {
// 		return enabled && len(data) >= compressionThreshold
// 	}
//
// 	return len(data) >= compressionThreshold
// }
//
// // compressData compresses data using gzip.
// func (r *redisClient) compressData(data []byte) ([]byte, error) {
// 	var buf bytes.Buffer
// 	buf.WriteByte(compressionMarker)
//
// 	writer, err := gzip.NewWriterLevel(&buf, r.config.CompressionLevel)
// 	if err != nil {
// 		return nil, fmt.Errorf("cache: failed to create gzip writer: %w", err)
// 	}
//
// 	if _, err := writer.Write(data); err != nil {
// 		writer.Close()
// 		return nil, fmt.Errorf("cache: failed to compress data: %w", err)
// 	}
//
// 	if err := writer.Close(); err != nil {
// 		return nil, fmt.Errorf("cache: failed to finalize compression: %w", err)
// 	}
//
// 	return buf.Bytes(), nil
// }
//
// // decompressData decompresses gzip data.
// func (r *redisClient) decompressData(data []byte) ([]byte, error) {
// 	if len(data) == 0 {
// 		return data, nil
// 	}
//
// 	if data[0] != compressionMarker {
// 		return data, nil
// 	}
//
// 	compressedData := data[1:]
// 	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
// 	if err != nil {
// 		return nil, fmt.Errorf("cache: %w: %v", ErrInvalidData, err)
// 	}
// 	defer reader.Close()
//
// 	var buf bytes.Buffer
// 	if _, err := io.Copy(&buf, reader); err != nil {
// 		return nil, fmt.Errorf("cache: failed to decompress: %w", err)
// 	}
//
// 	return buf.Bytes(), nil
// }
//
// // serializeValue serializes and optionally compresses a value.
// func (r *redisClient) serializeValue(ctx context.Context, value any) ([]byte, error) {
// 	if value == nil {
// 		return nil, ErrNilValue
// 	}
//
// 	data, err := json.Marshal(value)
// 	if err != nil {
// 		return nil, fmt.Errorf("cache: failed to marshal value: %w", err)
// 	}
//
// 	if r.shouldCompress(ctx, data) {
// 		return r.compressData(data)
// 	}
//
// 	return data, nil
// }
//
// // deserializeValue deserializes and optionally decompresses a value.
// func (r *redisClient) deserializeValue(data []byte, dest any) error {
// 	if dest == nil {
// 		return ErrNilValue
// 	}
//
// 	decompressed, err := r.decompressData(data)
// 	if err != nil {
// 		return err
// 	}
//
// 	if err := json.Unmarshal(decompressed, dest); err != nil {
// 		return fmt.Errorf("cache: failed to unmarshal value: %w", err)
// 	}
//
// 	return nil
// }
//
// // executeWithCircuitBreaker executes a Redis operation with circuit breaker protection.
// func (r *redisClient) executeWithCircuitBreaker(ctx context.Context, operation func() error) error {
// 	if r.circuitBreaker == nil {
// 		return operation()
// 	}
// 	return r.circuitBreaker.execute(operation)
// }
//
// // updateStats updates cache statistics atomically.
// func (r *redisClient) updateStats(operation string, err error, startTime time.Time) {
// 	latency := time.Since(startTime)
// 	r.latencyNs.Add(latency.Nanoseconds())
// 	r.opsCount.Add(1)
//
// 	switch operation {
// 	case "get":
// 		if err == ErrCacheMiss || errors.Is(err, ErrCacheMiss) {
// 			r.misses.Add(1)
// 		} else if err == nil {
// 			r.hits.Add(1)
// 		}
// 	case "set":
// 		if err == nil {
// 			r.sets.Add(1)
// 		}
// 	case "delete":
// 		if err == nil {
// 			r.deletes.Add(1)
// 		}
// 	}
//
// 	if err != nil && !errors.Is(err, ErrCacheMiss) {
// 		r.errors.Add(1)
// 		r.statsMutex.Lock()
// 		r.lastError = err.Error()
// 		r.lastErrorTime = time.Now()
// 		r.statsMutex.Unlock()
// 	}
// }
//
// // Get retrieves a value from Redis cache (tenant-aware).
// func (r *redisClient) Get(ctx context.Context, key string, dest any) error {
// 	startTime := time.Now()
//
// 	if dest == nil {
// 		return ErrNilValue
// 	}
//
// 	var err error
// 	err = r.executeWithCircuitBreaker(ctx, func() error {
// 		finalKey, keyErr := r.buildKey(ctx, key)
// 		if keyErr != nil {
// 			return keyErr
// 		}
//
// 		val, getErr := r.client.Get(ctx, finalKey).Bytes()
// 		if getErr != nil {
// 			if getErr == redis.Nil {
// 				return ErrCacheMiss
// 			}
// 			return fmt.Errorf("cache: get failed: %w", getErr)
// 		}
//
// 		return r.deserializeValue(val, dest)
// 	})
//
// 	r.updateStats("get", err, startTime)
// 	return err
// }
//
// // Set stores a value in Redis cache (tenant-aware).
// func (r *redisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
// 	startTime := time.Now()
//
// 	var err error
// 	err = r.executeWithCircuitBreaker(ctx, func() error {
// 		data, serErr := r.serializeValue(ctx, value)
// 		if serErr != nil {
// 			return serErr
// 		}
//
// 		finalKey, keyErr := r.buildKey(ctx, key)
// 		if keyErr != nil {
// 			return keyErr
// 		}
//
// 		if setErr := r.client.Set(ctx, finalKey, data, expiration).Err(); setErr != nil {
// 			return fmt.Errorf("cache: set failed: %w", setErr)
// 		}
// 		return nil
// 	})
//
// 	r.updateStats("set", err, startTime)
// 	return err
// }
//
// // Delete removes a value from Redis cache (tenant-aware).
// func (r *redisClient) Delete(ctx context.Context, key string) error {
// 	startTime := time.Now()
//
// 	var err error
// 	err = r.executeWithCircuitBreaker(ctx, func() error {
// 		finalKey, keyErr := r.buildKey(ctx, key)
// 		if keyErr != nil {
// 			return keyErr
// 		}
//
// 		if delErr := r.client.Del(ctx, finalKey).Err(); delErr != nil {
// 			return fmt.Errorf("cache: delete failed: %w", delErr)
// 		}
// 		return nil
// 	})
//
// 	r.updateStats("delete", err, startTime)
// 	return err
// }
//
// // Flush clears all cache entries for the current tenant.
// func (r *redisClient) Flush(ctx context.Context) error {
// 	tenantID, idType, err := r.getTenantInfo(ctx)
// 	if err != nil {
// 		return err
// 	}
//
// 	if tenantID == "" {
// 		return ErrNoTenantContext
// 	}
//
// 	pattern := fmt.Sprintf("%s:%s:%s-%s:*", r.config.KeyPrefix, tenantKeyPrefix, idType, tenantID)
// 	return r.deleteByPattern(ctx, pattern)
// }
//
// // MGet retrieves multiple values efficiently (tenant-aware).
// func (r *redisClient) MGet(ctx context.Context, keys []string) ([]Result, error) {
// 	if len(keys) == 0 {
// 		return []Result{}, nil
// 	}
//
// 	finalKeys := make([]string, len(keys))
// 	for i, key := range keys {
// 		fk, err := r.buildKey(ctx, key)
// 		if err != nil {
// 			return nil, err
// 		}
// 		finalKeys[i] = fk
// 	}
//
// 	results := make([]Result, len(keys))
// 	err := r.executeWithCircuitBreaker(ctx, func() error {
// 		vals, err := r.client.MGet(ctx, finalKeys...).Result()
// 		if err != nil {
// 			return fmt.Errorf("cache: mget failed: %w", err)
// 		}
//
// 		for i, val := range vals {
// 			results[i].Key = keys[i]
//
// 			if val == nil {
// 				results[i].Err = ErrCacheMiss
// 				continue
// 			}
//
// 			strVal, ok := val.(string)
// 			if !ok {
// 				results[i].Err = fmt.Errorf("%w: unexpected type %T", ErrInvalidData, val)
// 				continue
// 			}
//
// 			var decoded any
// 			if err := r.deserializeValue([]byte(strVal), &decoded); err != nil {
// 				results[i].Err = err
// 				continue
// 			}
//
// 			results[i].Value = decoded
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return results, nil
// }
//
// // MSet sets multiple values with the same expiration (tenant-aware).
// func (r *redisClient) MSet(ctx context.Context, pairs map[string]any, expiration time.Duration) error {
// 	if len(pairs) == 0 {
// 		return nil
// 	}
//
// 	return r.executeWithCircuitBreaker(ctx, func() error {
// 		pipe := r.client.Pipeline()
//
// 		for key, value := range pairs {
// 			data, err := r.serializeValue(ctx, value)
// 			if err != nil {
// 				return fmt.Errorf("cache: mset serialize failed for key %s: %w", key, err)
// 			}
//
// 			finalKey, keyErr := r.buildKey(ctx, key)
// 			if keyErr != nil {
// 				return keyErr
// 			}
//
// 			pipe.Set(ctx, finalKey, data, expiration)
// 		}
//
// 		if _, err := pipe.Exec(ctx); err != nil {
// 			return fmt.Errorf("cache: mset failed: %w", err)
// 		}
// 		return nil
// 	})
// }
//
// // MDelete deletes multiple keys efficiently (tenant-aware).
// func (r *redisClient) MDelete(ctx context.Context, keys []string) error {
// 	if len(keys) == 0 {
// 		return nil
// 	}
//
// 	finalKeys := make([]string, len(keys))
// 	for i, key := range keys {
// 		fk, err := r.buildKey(ctx, key)
// 		if err != nil {
// 			return err
// 		}
// 		finalKeys[i] = fk
// 	}
//
// 	return r.executeWithCircuitBreaker(ctx, func() error {
// 		if err := r.client.Del(ctx, finalKeys...).Err(); err != nil {
// 			return fmt.Errorf("cache: mdelete failed: %w", err)
// 		}
// 		return nil
// 	})
// }
//
// // DeletePattern deletes keys matching a pattern (tenant-aware).
// func (r *redisClient) DeletePattern(ctx context.Context, pattern string) error {
// 	finalPattern, err := r.buildKey(ctx, pattern)
// 	if err != nil {
// 		return err
// 	}
// 	return r.deleteByPattern(ctx, finalPattern)
// }
//
// // Keys returns keys matching a pattern (tenant-aware).
// func (r *redisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
// 	finalPattern, err := r.buildKey(ctx, pattern)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var keys []string
// 	err = r.executeWithCircuitBreaker(ctx, func() error {
// 		iter := r.client.Scan(ctx, 0, finalPattern, 0).Iterator()
//
// 		for iter.Next(ctx) {
// 			keys = append(keys, iter.Val())
// 		}
//
// 		return iter.Err()
// 	})
//
// 	return keys, err
// }
//
// // deleteByPattern internal method to delete by pattern in batches.
// func (r *redisClient) deleteByPattern(ctx context.Context, pattern string) error {
// 	return r.executeWithCircuitBreaker(ctx, func() error {
// 		iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
// 		keys := make([]string, 0, r.config.BatchDeleteSize)
//
// 		for iter.Next(ctx) {
// 			keys = append(keys, iter.Val())
//
// 			if len(keys) >= r.config.BatchDeleteSize {
// 				if err := r.client.Del(ctx, keys...).Err(); err != nil {
// 					return fmt.Errorf("cache: batch delete failed: %w", err)
// 				}
// 				keys = keys[:0]
// 			}
// 		}
//
// 		if len(keys) > 0 {
// 			if err := r.client.Del(ctx, keys...).Err(); err != nil {
// 				return fmt.Errorf("cache: final batch delete failed: %w", err)
// 			}
// 		}
//
// 		if err := iter.Err(); err != nil {
// 			return fmt.Errorf("cache: scan failed: %w", err)
// 		}
//
// 		return nil
// 	})
// }
//
// // Exists checks if a key exists (tenant-aware).
// func (r *redisClient) Exists(ctx context.Context, key string) (bool, error) {
// 	var exists bool
// 	err := r.executeWithCircuitBreaker(ctx, func() error {
// 		finalKey, keyErr := r.buildKey(ctx, key)
// 		if keyErr != nil {
// 			return keyErr
// 		}
//
// 		count, err := r.client.Exists(ctx, finalKey).Result()
// 		if err != nil {
// 			return fmt.Errorf("cache: exists check failed: %w", err)
// 		}
// 		exists = count > 0
// 		return nil
// 	})
// 	return exists, err
// }
//
// // TTL returns the time to live for a key (tenant-aware).
// // Returns -1 if key exists but has no expiration.
// // Returns -2 if key does not exist.
// func (r *redisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
// 	var ttl time.Duration
// 	err := r.executeWithCircuitBreaker(ctx, func() error {
// 		finalKey, keyErr := r.buildKey(ctx, key)
// 		if keyErr != nil {
// 			return keyErr
// 		}
//
// 		result, err := r.client.TTL(ctx, finalKey).Result()
// 		if err != nil {
// 			return fmt.Errorf("cache: ttl check failed: %w", err)
// 		}
// 		ttl = result
// 		return nil
// 	})
// 	return ttl, err
// }
//
// // Expire sets a timeout on a key (tenant-aware).
// func (r *redisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
// 	return r.executeWithCircuitBreaker(ctx, func() error {
// 		finalKey, keyErr := r.buildKey(ctx, key)
// 		if keyErr != nil {
// 			return keyErr
// 		}
//
// 		if err := r.client.Expire(ctx, finalKey, expiration).Err(); err != nil {
// 			return fmt.Errorf("cache: expire failed: %w", err)
// 		}
// 		return nil
// 	})
// }
//
// // GetMemory retrieves a value from in-memory cache (tenant-aware).
// func (r *redisClient) GetMemory(ctx context.Context, key string, dest any) error {
// 	if !r.config.EnableMemoryCache {
// 		return errors.New("cache: memory cache is disabled")
// 	}
//
// 	if dest == nil {
// 		return ErrNilValue
// 	}
//
// 	memKey, err := r.buildMemoryKey(ctx, key)
// 	if err != nil {
// 		return err
// 	}
//
// 	r.memoryCacheMu.RLock()
// 	entry, exists := r.memoryCache[memKey]
// 	r.memoryCacheMu.RUnlock()
//
// 	if !exists {
// 		return ErrCacheMiss
// 	}
//
// 	// Check expiration
// 	if !entry.expiresAt.IsZero() && entry.expiresAt.Before(time.Now()) {
// 		r.memoryCacheMu.Lock()
// 		delete(r.memoryCache, memKey)
// 		r.memoryCacheMu.Unlock()
// 		return ErrCacheMiss
// 	}
//
// 	// Copy value to dest
// 	data, err := json.Marshal(entry.value)
// 	if err != nil {
// 		return fmt.Errorf("cache: failed to marshal memory value: %w", err)
// 	}
//
// 	if err := json.Unmarshal(data, dest); err != nil {
// 		return fmt.Errorf("cache: failed to unmarshal memory value: %w", err)
// 	}
//
// 	return nil
// }
//
// // SetMemory stores a value in in-memory cache (tenant-aware).
// func (r *redisClient) SetMemory(ctx context.Context, key string, value any, expiration time.Duration) error {
// 	if !r.config.EnableMemoryCache {
// 		return errors.New("cache: memory cache is disabled")
// 	}
//
// 	if value == nil {
// 		return ErrNilValue
// 	}
//
// 	memKey, err := r.buildMemoryKey(ctx, key)
// 	if err != nil {
// 		return err
// 	}
//
// 	var expiresAt time.Time
// 	if expiration > 0 {
// 		expiresAt = time.Now().Add(expiration)
// 	}
//
// 	entry := &memoryEntry{
// 		value:     value,
// 		expiresAt: expiresAt,
// 	}
//
// 	r.memoryCacheMu.Lock()
// 	defer r.memoryCacheMu.Unlock()
//
// 	// Check size limit
// 	if len(r.memoryCache) >= r.config.MemoryCacheMaxSize {
// 		// Simple eviction: remove first entry (could be improved with LRU)
// 		for k := range r.memoryCache {
// 			delete(r.memoryCache, k)
// 			break
// 		}
// 	}
//
// 	r.memoryCache[memKey] = entry
// 	return nil
// }
//
// // DeleteMemory removes a value from in-memory cache (tenant-aware).
// func (r *redisClient) DeleteMemory(ctx context.Context, key string) error {
// 	if !r.config.EnableMemoryCache {
// 		return errors.New("cache: memory cache is disabled")
// 	}
//
// 	memKey, err := r.buildMemoryKey(ctx, key)
// 	if err != nil {
// 		return err
// 	}
//
// 	r.memoryCacheMu.Lock()
// 	delete(r.memoryCache, memKey)
// 	r.memoryCacheMu.Unlock()
//
// 	return nil
// }
//
// // GetGlobalMemory retrieves a value from global in-memory cache (not tenant-aware).
// // Use this for shared data like formulas, configurations, etc.
// func (r *redisClient) GetGlobalMemory(key string, dest any) error {
// 	if !r.config.EnableMemoryCache {
// 		return errors.New("cache: memory cache is disabled")
// 	}
//
// 	if dest == nil {
// 		return ErrNilValue
// 	}
//
// 	r.globalCacheMu.RLock()
// 	entry, exists := r.globalMemoryCache[key]
// 	r.globalCacheMu.RUnlock()
//
// 	if !exists {
// 		return ErrCacheMiss
// 	}
//
// 	// Check expiration
// 	if !entry.expiresAt.IsZero() && entry.expiresAt.Before(time.Now()) {
// 		r.globalCacheMu.Lock()
// 		delete(r.globalMemoryCache, key)
// 		r.globalCacheMu.Unlock()
// 		return ErrCacheMiss
// 	}
//
// 	// Copy value to dest
// 	data, err := json.Marshal(entry.value)
// 	if err != nil {
// 		return fmt.Errorf("cache: failed to marshal global memory value: %w", err)
// 	}
//
// 	if err := json.Unmarshal(data, dest); err != nil {
// 		return fmt.Errorf("cache: failed to unmarshal global memory value: %w", err)
// 	}
//
// 	return nil
// }
//
// // SetGlobalMemory stores a value in global in-memory cache (not tenant-aware).
// // Use this for shared data like formulas, configurations, etc.
// func (r *redisClient) SetGlobalMemory(key string, value any, expiration time.Duration) error {
// 	if !r.config.EnableMemoryCache {
// 		return errors.New("cache: memory cache is disabled")
// 	}
//
// 	if value == nil {
// 		return ErrNilValue
// 	}
//
// 	var expiresAt time.Time
// 	if expiration > 0 {
// 		expiresAt = time.Now().Add(expiration)
// 	}
//
// 	entry := &memoryEntry{
// 		value:     value,
// 		expiresAt: expiresAt,
// 	}
//
// 	r.globalCacheMu.Lock()
// 	defer r.globalCacheMu.Unlock()
//
// 	// Check size limit
// 	if len(r.globalMemoryCache) >= r.config.MemoryCacheMaxSize {
// 		// Simple eviction: remove first entry
// 		for k := range r.globalMemoryCache {
// 			delete(r.globalMemoryCache, k)
// 			break
// 		}
// 	}
//
// 	r.globalMemoryCache[key] = entry
// 	return nil
// }
//
// // DeleteGlobalMemory removes a value from global in-memory cache (not tenant-aware).
// func (r *redisClient) DeleteGlobalMemory(key string) error {
// 	if !r.config.EnableMemoryCache {
// 		return errors.New("cache: memory cache is disabled")
// 	}
//
// 	r.globalCacheMu.Lock()
// 	delete(r.globalMemoryCache, key)
// 	r.globalCacheMu.Unlock()
//
// 	return nil
// }
//
// // Ping tests the Redis connection.
// func (r *redisClient) Ping(ctx context.Context) error {
// 	return r.executeWithCircuitBreaker(ctx, func() error {
// 		if err := r.client.Ping(ctx).Err(); err != nil {
// 			return fmt.Errorf("cache: ping failed: %w", err)
// 		}
// 		return nil
// 	})
// }
//
// // Stats returns a snapshot of cache statistics.
// func (r *redisClient) Stats() CacheStats {
// 	hits := r.hits.Load()
// 	misses := r.misses.Load()
// 	sets := r.sets.Load()
// 	deletes := r.deletes.Load()
// 	errors := r.errors.Load()
// 	opsCount := r.opsCount.Load()
// 	latencyNs := r.latencyNs.Load()
//
// 	var hitRatio float64
// 	if hits+misses > 0 {
// 		hitRatio = float64(hits) / float64(hits+misses)
// 	}
//
// 	var avgLatency time.Duration
// 	if opsCount > 0 {
// 		avgLatency = time.Duration(latencyNs / opsCount)
// 	}
//
// 	poolStats := r.client.PoolStats()
//
// 	r.statsMutex.RLock()
// 	lastError := r.lastError
// 	lastErrorTime := r.lastErrorTime
// 	r.statsMutex.RUnlock()
//
// 	// Get memory cache sizes
// 	r.memoryCacheMu.RLock()
// 	memoryCacheSize := len(r.memoryCache)
// 	r.memoryCacheMu.RUnlock()
//
// 	r.globalCacheMu.RLock()
// 	globalCacheSize := len(r.globalMemoryCache)
// 	r.globalCacheMu.RUnlock()
//
// 	return CacheStats{
// 		Hits:              hits,
// 		Misses:            misses,
// 		Sets:              sets,
// 		Deletes:           deletes,
// 		Errors:            errors,
// 		HitRatio:          hitRatio,
// 		AverageLatency:    avgLatency,
// 		ConnectionsActive: int(poolStats.TotalConns - poolStats.IdleConns),
// 		ConnectionsIdle:   int(poolStats.IdleConns),
// 		MemoryCacheSize:   memoryCacheSize,
// 		GlobalCacheSize:   globalCacheSize,
// 		LastError:         lastError,
// 		LastErrorTime:     lastErrorTime,
// 	}
// }
//
// // Reset resets all statistics counters.
// func (r *redisClient) Reset() {
// 	r.hits.Store(0)
// 	r.misses.Store(0)
// 	r.sets.Store(0)
// 	r.deletes.Store(0)
// 	r.errors.Store(0)
// 	r.opsCount.Store(0)
// 	r.latencyNs.Store(0)
//
// 	r.statsMutex.Lock()
// 	r.lastError = ""
// 	r.lastErrorTime = time.Time{}
// 	r.statsMutex.Unlock()
// }
//
// // Close closes the Redis connection gracefully.
// func (r *redisClient) Close() error {
// 	// Stop cleanup routine
// 	close(r.stopCleanup)
// 	r.cleanupWg.Wait()
//
// 	if err := r.client.Close(); err != nil {
// 		return fmt.Errorf("cache: failed to close connection: %w", err)
// 	}
// 	return nil
// }
//
// // circuitBreaker implements the circuit breaker pattern.
// type circuitBreaker struct {
// 	threshold    int
// 	timeout      time.Duration
// 	failureCount atomic.Int64
// 	lastFailure  atomic.Int64
// 	state        atomic.Int32
// 	mu           sync.Mutex
// }
//
// const (
// 	stateClosed   int32 = 0
// 	stateOpen     int32 = 1
// 	stateHalfOpen int32 = 2
// )
//
// func newCircuitBreaker(threshold int, timeout time.Duration) *circuitBreaker {
// 	return &circuitBreaker{
// 		threshold: threshold,
// 		timeout:   timeout,
// 	}
// }
//
// func (cb *circuitBreaker) execute(operation func() error) error {
// 	state := cb.state.Load()
//
// 	if state == stateOpen {
// 		lastFailureNs := cb.lastFailure.Load()
// 		lastFailure := time.Unix(0, lastFailureNs)
//
// 		if time.Since(lastFailure) > cb.timeout {
// 			cb.mu.Lock()
// 			if cb.state.Load() == stateOpen {
// 				cb.state.Store(stateHalfOpen)
// 			}
// 			cb.mu.Unlock()
// 		} else {
// 			return ErrCircuitOpen
// 		}
// 	}
//
// 	err := operation()
//
// 	cb.mu.Lock()
// 	defer cb.mu.Unlock()
//
// 	if err != nil {
// 		failures := cb.failureCount.Add(1)
// 		cb.lastFailure.Store(time.Now().UnixNano())
//
// 		if failures >= int64(cb.threshold) {
// 			cb.state.Store(stateOpen)
// 		}
// 	} else {
// 		cb.failureCount.Store(0)
// 		cb.state.Store(stateClosed)
// 	}
//
// 	return err
// }
//
// func (cb *circuitBreaker) State() string {
// 	switch cb.state.Load() {
// 	case stateClosed:
// 		return "closed"
// 	case stateOpen:
// 		return "open"
// 	case stateHalfOpen:
// 		return "half-open"
// 	default:
// 		return "unknown"
// 	}
// }
//
// func (cb *circuitBreaker) Failures() int64 {
// 	return cb.failureCount.Load()
// }
//
// // Context helper functions for tenant information.
//
// // MustGetTenantID extracts tenant ID from context or panics.
// // Use this in middleware where tenant context is guaranteed.
// func MustGetTenantID(ctx context.Context) uuid.UUID {
// 	if id, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok && id != uuid.Nil {
// 		return id
// 	}
// 	panic("cache: tenant_id not found in context")
// }
//
// // MustGetTenantSlug extracts tenant slug from context or panics.
// func MustGetTenantSlug(ctx context.Context) string {
// 	if slug, ok := ctx.Value(TenantSlugKey).(string); ok && slug != "" {
// 		return slug
// 	}
// 	panic("cache: tenant_slug not found in context")
// }
//
// // MustGetTenantSubdomain extracts tenant subdomain from context or panics.
// func MustGetTenantSubdomain(ctx context.Context) string {
// 	if subdomain, ok := ctx.Value(TenantSubdomainKey).(string); ok && subdomain != "" {
// 		return subdomain
// 	}
// 	panic("cache: tenant_subdomain not found in context")
// }
//
// // WithTenantID creates a context with tenant ID.
// func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
// 	return context.WithValue(ctx, TenantIDKey, tenantID)
// }
//
// // WithTenantSlug creates a context with tenant slug.
// func WithTenantSlug(ctx context.Context, tenantSlug string) context.Context {
// 	return context.WithValue(ctx, TenantSlugKey, tenantSlug)
// }
//
// // WithTenantSubdomain creates a context with tenant subdomain.
// func WithTenantSubdomain(ctx context.Context, subdomain string) context.Context {
// 	return context.WithValue(ctx, TenantSubdomainKey, subdomain)
// }
//
// // WithNamespace creates a context with a custom namespace.
// func WithNamespace(ctx context.Context, namespace string) context.Context {
// 	return context.WithValue(ctx, NamespaceKey, namespace)
// }
//
// // WithCompression creates a context with compression settings.
// func WithCompression(ctx context.Context, enabled bool) context.Context {
// 	return context.WithValue(ctx, CompressionKey, enabled)
// }
