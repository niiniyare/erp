package cache

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"awo.so/internal/platform/config"
	sharedErrors "awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Context keys for tenant information.
type contextKey string

const (
	// TenantIDKey is the context key for tenant UUID.
	TenantIDKey contextKey = "tenant_id"
	// TenantSlugKey is the context key for tenant slug.
	TenantSlugKey contextKey = "tenant_slug"
	// TenantSubdomainKey is the context key for tenant subdomain.
	TenantSubdomainKey contextKey = "tenant_subdomain"
	// NamespaceKey is the context key for cache namespace.
	NamespaceKey contextKey = "cache_namespace"
	// CompressionKey is the context key for compression override.
	CompressionKey contextKey = "cache_compression"

	// Cache configuration constants.
	defaultKeyPrefix     = "erp"
	tenantKeyPrefix      = "tenant"
	globalKeyPrefix      = "global"
	compressionThreshold = 1024 // Compress values larger than 1KB
	maxKeyLength         = 250  // Redis key length limit
	compressionMarker    = 0x1F // Unit separator - unlikely in normal data
	batchDeleteSize      = 1000 // Keys to delete per batch
)

// Standard errors.
var (
	// ErrCacheMiss indicates the requested key was not found.
	ErrCacheMiss = errors.New("cache: key not found")
	// ErrCircuitOpen indicates the circuit breaker is open.
	ErrCircuitOpen = errors.New("cache: circuit breaker open")
	// ErrInvalidData indicates corrupted or invalid cached data.
	ErrInvalidData = errors.New("cache: invalid data format")
	// ErrNilValue indicates a nil value was provided.
	ErrNilValue = errors.New("cache: nil value provided")
	// ErrNoTenantContext indicates tenant information is missing from context.
	ErrNoTenantContext = errors.New("cache: tenant context required - missing tenant_id, tenant_slug, or tenant_subdomain")
)

// Service defines the interface for a cache service.
// All methods are safe for concurrent use and require tenant context.
type Service interface {
	// Core operations
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Flush(ctx context.Context) error

	// Bulk operations
	MGet(ctx context.Context, keys []string) ([]Result, error)
	MSet(ctx context.Context, pairs map[string]any, expiration time.Duration) error
	MDelete(ctx context.Context, keys []string) error

	// Pattern operations
	DeletePattern(ctx context.Context, pattern string) error
	Keys(ctx context.Context, pattern string) ([]string, error)
	Exists(ctx context.Context, key string) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error

	// In-memory cache operations (tenant-aware)
	GetMemory(ctx context.Context, key string, dest any) error
	SetMemory(ctx context.Context, key string, value any, expiration time.Duration) error
	DeleteMemory(ctx context.Context, key string) error

	// Global in-memory cache operations (not tenant-aware, for shared data like formulas)
	GetGlobalMemory(key string, dest any) error
	SetGlobalMemory(key string, value any, expiration time.Duration) error
	DeleteGlobalMemory(key string) error

	// Health and monitoring
	Ping(ctx context.Context) error
	Stats() CacheStats
	Reset() // Reset statistics
	Close() error
}

// Result represents a single cache get result.
type Result struct {
	Key   string
	Value any
	Err   error
}

// CacheStats provides cache metrics.
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
	MemoryCacheSize   int           `json:"memory_cache_size"`
	GlobalCacheSize   int           `json:"global_cache_size"`
	LastError         string        `json:"last_error,omitempty"`
	LastErrorTime     time.Time     `json:"last_error_time,omitempty"`
}

// RedisConfig extends the base config with advanced options.
type RedisConfig struct {
	*config.RedisConfig

	// Connection pool settings
	PoolSize     int           `yaml:"pool_size"`
	MinIdleConns int           `yaml:"min_idle_conns"`
	MaxConnAge   time.Duration `yaml:"max_conn_age"`
	PoolTimeout  time.Duration `yaml:"pool_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`

	// Performance settings
	EnableCompression bool          `yaml:"enable_compression"`
	CompressionLevel  int           `yaml:"compression_level"` // 1-9, default 6
	KeyPrefix         string        `yaml:"key_prefix"`
	MaxRetries        int           `yaml:"max_retries"`
	RetryDelay        time.Duration `yaml:"retry_delay"`
	BatchDeleteSize   int           `yaml:"batch_delete_size"`

	// Circuit breaker settings
	EnableCircuitBreaker    bool          `yaml:"enable_circuit_breaker"`
	CircuitBreakerThreshold int           `yaml:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   time.Duration `yaml:"circuit_breaker_timeout"`

	// Tenant settings
	RequireTenantContext  bool `yaml:"require_tenant_context"`  // Enforce tenant context on all operations
	AllowGlobalOperations bool `yaml:"allow_global_operations"` // Allow operations without tenant context

	// In-memory cache settings
	EnableMemoryCache          bool          `yaml:"enable_memory_cache"`
	MemoryCacheMaxSize         int           `yaml:"memory_cache_max_size"` // Max items in memory cache
	MemoryCacheDefaultTTL      time.Duration `yaml:"memory_cache_default_ttl"`
	MemoryCacheCleanupInterval time.Duration `yaml:"memory_cache_cleanup_interval"`

	// Observability settings
	EnableTracing bool `yaml:"enable_tracing"`
	EnableMetrics bool `yaml:"enable_metrics"`
	EnableLogging bool `yaml:"enable_logging"`
}

// DefaultRedisConfig returns production-ready defaults.
func DefaultRedisConfig(baseConfig *config.RedisConfig) *RedisConfig {
	if baseConfig == nil {
		panic("cache: base config cannot be nil")
	}

	return &RedisConfig{
		RedisConfig:                baseConfig,
		PoolSize:                   10,
		MinIdleConns:               3,
		MaxConnAge:                 30 * time.Minute,
		PoolTimeout:                4 * time.Second,
		IdleTimeout:                5 * time.Minute,
		EnableCompression:          true,
		CompressionLevel:           6,
		KeyPrefix:                  defaultKeyPrefix,
		MaxRetries:                 3,
		RetryDelay:                 100 * time.Millisecond,
		BatchDeleteSize:            batchDeleteSize,
		EnableCircuitBreaker:       true,
		CircuitBreakerThreshold:    5,
		CircuitBreakerTimeout:      30 * time.Second,
		RequireTenantContext:       true,  // Enforce tenant context by default
		AllowGlobalOperations:      false, // Disallow global operations by default
		EnableMemoryCache:          true,
		MemoryCacheMaxSize:         1000,
		MemoryCacheDefaultTTL:      5 * time.Minute,
		MemoryCacheCleanupInterval: 1 * time.Minute,
		EnableTracing:              true,
		EnableMetrics:              true,
		EnableLogging:              true,
	}
}

// Validate checks if the configuration is valid.
func (c *RedisConfig) Validate() error {
	if c.RedisConfig == nil {
		return sharedErrors.NewBusinessError("INVALID_CONFIG", "Base redis config is required").
			WithHTTPStatus(400).
			WithCategory(sharedErrors.CategoryValidation)
	}
	if c.PoolSize < 1 {
		return sharedErrors.NewBusinessError("INVALID_CONFIG", "Pool size must be at least 1").
			WithHTTPStatus(400).
			WithCategory(sharedErrors.CategoryValidation).
			WithDetail("pool_size", c.PoolSize)
	}
	if c.CompressionLevel < 1 || c.CompressionLevel > 9 {
		return sharedErrors.NewBusinessError("INVALID_CONFIG", "Compression level must be between 1 and 9").
			WithHTTPStatus(400).
			WithCategory(sharedErrors.CategoryValidation).
			WithDetail("compression_level", c.CompressionLevel)
	}
	if c.KeyPrefix == "" {
		return sharedErrors.NewBusinessError("INVALID_CONFIG", "Key prefix cannot be empty").
			WithHTTPStatus(400).
			WithCategory(sharedErrors.CategoryValidation)
	}
	if c.BatchDeleteSize < 1 {
		return sharedErrors.NewBusinessError("INVALID_CONFIG", "Batch delete size must be at least 1").
			WithHTTPStatus(400).
			WithCategory(sharedErrors.CategoryValidation).
			WithDetail("batch_delete_size", c.BatchDeleteSize)
	}
	if c.EnableMemoryCache && c.MemoryCacheMaxSize < 1 {
		return sharedErrors.NewBusinessError("INVALID_CONFIG", "Memory cache max size must be at least 1").
			WithHTTPStatus(400).
			WithCategory(sharedErrors.CategoryValidation).
			WithDetail("memory_cache_max_size", c.MemoryCacheMaxSize)
	}
	return nil
}

// memoryEntry represents a cached item in memory.
type memoryEntry struct {
	value     any
	expiresAt time.Time
}

// redisClient implements the Service interface with multi-tenant support.
type redisClient struct {
	client             *redis.Client
	config             *RedisConfig
	circuitBreaker     *circuitBreaker
	keyHashingEnabled  bool
	compressionEnabled bool

	// Observability
	logger         logger.Logger
	tracer         tracing.Service
	metricsService *metrics.MetricsService

	// In-memory caches
	memoryCache       map[string]*memoryEntry // Tenant-aware memory cache
	globalMemoryCache map[string]*memoryEntry // Global memory cache (formulas, etc.)
	memoryCacheMu     sync.RWMutex
	globalCacheMu     sync.RWMutex

	// Statistics (use atomic operations for thread-safety)
	hits      atomic.Int64
	misses    atomic.Int64
	sets      atomic.Int64
	deletes   atomic.Int64
	errors    atomic.Int64
	latencyNs atomic.Int64
	opsCount  atomic.Int64

	lastError     string
	lastErrorTime time.Time
	statsMutex    sync.RWMutex

	stopCleanup chan struct{}
	cleanupWg   sync.WaitGroup
}

// NewRedisClient creates a new Redis client with enhanced configuration.
func NewRedisClient(cfg *RedisConfig) (Service, error) {
	return NewRedisClientWithObservability(cfg, nil, nil, nil)
}

// NewRedisClientWithObservability creates a Redis client with observability components.
func NewRedisClientWithObservability(
	cfg *RedisConfig,
	log logger.Logger,
	tracer tracing.Service,
	metricsService *metrics.MetricsService,
) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cache: invalid config: %w", err)
	}

	// Initialize logger if not provided
	if log == nil && cfg.EnableLogging {
		log = logger.WithFields(logger.Fields{"component": "cache"})
	}

	// Initialize tracer if not provided
	if tracer == nil && cfg.EnableTracing {
		tracer = tracing.NewNoOpService()
	}

	// Initialize metrics if not provided
	if metricsService == nil && cfg.EnableMetrics {
		metricsConfig := metrics.MetricsConfig{
			Provider:  "prometheus",
			Namespace: "erp",
			Subsystem: "cache",
			Enabled:   true,
		}
		var err error
		metricsService, err = metrics.NewMetricsService(metricsConfig)
		if err != nil && log != nil {
			log.Warn("Failed to initialize metrics service", logger.Fields{"error": err.Error()})
		}
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		MaxConnAge:      cfg.MaxConnAge,
		PoolTimeout:     cfg.PoolTimeout,
		IdleTimeout:     cfg.IdleTimeout,
		MaxRetries:      cfg.MaxRetries,
		MinRetryBackoff: cfg.RetryDelay,
		MaxRetryBackoff: cfg.RetryDelay * 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		if log != nil {
			log.Error("Failed to connect to Redis", logger.Fields{"error": err.Error()})
		}
		return nil, sharedErrors.NewBusinessError("REDIS_CONNECTION_FAILED", "Failed to connect to Redis").
			WithHTTPStatus(500).
			WithCategory(sharedErrors.CategorySystem).
			WithDetail("error", err.Error()).
			WithSuggestion("Check Redis connection settings").
			WithSuggestion("Ensure Redis server is running")
	}

	if log != nil {
		log.Info("Successfully connected to Redis", logger.Fields{
			"host": cfg.Host,
			"port": cfg.Port,
			"db":   cfg.DB,
		})
	}

	client := &redisClient{
		client:             rdb,
		config:             cfg,
		keyHashingEnabled:  true,
		compressionEnabled: cfg.EnableCompression,
		logger:             log,
		tracer:             tracer,
		metricsService:     metricsService,
		stopCleanup:        make(chan struct{}),
	}

	if cfg.EnableMemoryCache {
		client.memoryCache = make(map[string]*memoryEntry)
		client.globalMemoryCache = make(map[string]*memoryEntry)
		client.startCleanupRoutine()
	}

	if cfg.EnableCircuitBreaker {
		client.circuitBreaker = newCircuitBreaker(
			cfg.CircuitBreakerThreshold,
			cfg.CircuitBreakerTimeout,
		)
	}

	// Initialize metrics
	if cfg.EnableMetrics && metricsService != nil {
		client.initializeMetrics()
	}

	return client, nil
}

// initializeMetrics sets up cache metrics.
func (r *redisClient) initializeMetrics() {
	if r.metricsService == nil {
		return
	}

	// Create counters
	r.metricsService.Counter("operations_total", "Total cache operations", "operation", "result", "tenant_id")
	r.metricsService.Counter("hits_total", "Total cache hits", "tenant_id")
	r.metricsService.Counter("misses_total", "Total cache misses", "tenant_id")
	r.metricsService.Counter("errors_total", "Total cache errors", "operation", "error_type", "tenant_id")

	// Create gauges
	r.metricsService.Gauge("memory_cache_size", "In-memory cache size", "type")
	r.metricsService.Gauge("connections_active", "Active Redis connections")
	r.metricsService.Gauge("connections_idle", "Idle Redis connections")

	// Create histograms
	buckets := []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0}
	r.metricsService.Histogram("operation_duration_seconds", "Cache operation duration", buckets, "operation", "tenant_id")
}

// NewRedisClientMust is like NewRedisClient but panics on error.
func NewRedisClientMust(cfg *RedisConfig) Service {
	client, err := NewRedisClient(cfg)
	if err != nil {
		panic(err)
	}
	return client
}

// startCleanupRoutine starts a background goroutine to clean up expired entries.
func (r *redisClient) startCleanupRoutine() {
	r.cleanupWg.Add(1)
	go func() {
		defer r.cleanupWg.Done()
		ticker := time.NewTicker(r.config.MemoryCacheCleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.cleanupExpiredEntries()
			case <-r.stopCleanup:
				return
			}
		}
	}()
}

// cleanupExpiredEntries removes expired entries from memory caches.
func (r *redisClient) cleanupExpiredEntries() {
	now := time.Now()

	// Clean tenant-aware cache
	r.memoryCacheMu.Lock()
	cleanedCount := 0
	for key, entry := range r.memoryCache {
		if !entry.expiresAt.IsZero() && entry.expiresAt.Before(now) {
			delete(r.memoryCache, key)
			cleanedCount++
		}
	}
	r.memoryCacheMu.Unlock()

	// Clean global cache
	r.globalCacheMu.Lock()
	for key, entry := range r.globalMemoryCache {
		if !entry.expiresAt.IsZero() && entry.expiresAt.Before(now) {
			delete(r.globalMemoryCache, key)
			cleanedCount++
		}
	}
	r.globalCacheMu.Unlock()

	if r.logger != nil && cleanedCount > 0 {
		r.logger.Debug("Cleaned expired cache entries", logger.Fields{
			"cleaned_count": cleanedCount,
		})
	}

	// Update metrics
	if r.metricsService != nil {
		r.updateMemoryCacheMetrics()
	}
}

// getTenantInfo extracts tenant information from context.
// Returns the first available identifier in priority order: ID, slug, subdomain.
func (r *redisClient) getTenantInfo(ctx context.Context) (identifier string, identifierType string, err error) {
	// Check for tenant ID first
	if id, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok && id != uuid.Nil {
		return id.String(), "id", nil
	}

	// Check for tenant slug
	if slug, ok := ctx.Value(TenantSlugKey).(string); ok && slug != "" {
		return slug, "slug", nil
	}

	// Check for tenant subdomain
	if subdomain, ok := ctx.Value(TenantSubdomainKey).(string); ok && subdomain != "" {
		return subdomain, "subdomain", nil
	}

	// No tenant information found
	if r.config.RequireTenantContext && !r.config.AllowGlobalOperations {
		if r.logger != nil {
			r.logger.Warn("Missing tenant context in cache operation", logger.Fields{
				"require_tenant": r.config.RequireTenantContext,
				"allow_global":   r.config.AllowGlobalOperations,
			})
		}
		return "", "", ErrNoTenantContext
	}

	if r.logger != nil {
		r.logger.Debug("No tenant context found, proceeding with global operation")
	}

	return "", "", nil
}

// buildKey constructs a tenant-aware cache key.
func (r *redisClient) buildKey(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", sharedErrors.NewBusinessError("EMPTY_KEY", "Cache key cannot be empty").
			WithHTTPStatus(400).
			WithCategory(sharedErrors.CategoryValidation)
	}

	parts := []string{r.config.KeyPrefix}

	// Get tenant information
	tenantID, idType, err := r.getTenantInfo(ctx)
	if err != nil {
		return "", err
	}

	if tenantID != "" {
		parts = append(parts, tenantKeyPrefix, fmt.Sprintf("%s-%s", idType, tenantID))
	} else if r.config.AllowGlobalOperations {
		parts = append(parts, globalKeyPrefix)
	} else {
		return "", ErrNoTenantContext
	}

	// Add custom namespace if provided
	if namespace, ok := ctx.Value(NamespaceKey).(string); ok && namespace != "" {
		parts = append(parts, namespace)
	}

	parts = append(parts, key)
	finalKey := strings.Join(parts, ":")

	// Hash key if it exceeds Redis limits
	if r.keyHashingEnabled && len(finalKey) > maxKeyLength {
		hash := sha256.Sum256([]byte(finalKey))
		hashedKey := hex.EncodeToString(hash[:])
		prefix := strings.Join(parts[:len(parts)-1], ":")
		maxPrefixLen := maxKeyLength - 65
		if len(prefix) > maxPrefixLen {
			prefix = prefix[:maxPrefixLen]
		}
		finalKey = prefix + ":" + hashedKey
	}

	return finalKey, nil
}

// Get retrieves a value from Redis cache (tenant-aware).
func (r *redisClient) Get(ctx context.Context, key string, dest any) error {
	// Start tracing span
	var span tracing.Span
	if r.tracer != nil {
		ctx, span = r.tracer.StartSpan(ctx, "cache.Get",
			tracing.WithSpanKind(tracing.SpanKindClient),
			tracing.WithAttributes(
				attribute.String("cache.key", key),
				attribute.String("cache.operation", "get"),
			),
		)
		defer span.End()
	}

	startTime := time.Now()
	tenantID, _, _ := r.getTenantInfo(ctx)

	if dest == nil {
		err := ErrNilValue
		r.recordError(ctx, span, "get", err, tenantID, startTime)
		return err
	}

	var err error
	err = r.executeWithCircuitBreaker(ctx, func() error {
		finalKey, keyErr := r.buildKey(ctx, key)
		if keyErr != nil {
			return keyErr
		}

		if span != nil {
			span.SetAttributes(attribute.String("cache.final_key", finalKey))
		}

		val, getErr := r.client.Get(ctx, finalKey).Bytes()
		if getErr != nil {
			if getErr == redis.Nil {
				return ErrCacheMiss
			}
			return sharedErrors.NewBusinessError("CACHE_GET_FAILED", "Failed to get value from cache").
				WithHTTPStatus(500).
				WithCategory(sharedErrors.CategorySystem).
				WithDetail("error", getErr.Error()).
				WithDetail("key", key)
		}

		return r.deserializeValue(val, dest)
	})

	r.updateStats("get", err, startTime)
	r.recordMetrics(ctx, "get", err, tenantID, startTime)

	if err != nil {
		r.recordError(ctx, span, "get", err, tenantID, startTime)
		return err
	}

	if span != nil {
		span.SetStatus(codes.Ok, "Cache hit")
	}

	if r.logger != nil {
		r.logger.Debug("Cache get", logger.Fields{
			"key":       key,
			"tenant_id": tenantID,
			"hit":       err == nil,
		})
	}

	return nil
}

// Set stores a value in Redis cache (tenant-aware).
func (r *redisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	// Start tracing span
	var span tracing.Span
	if r.tracer != nil {
		ctx, span = r.tracer.StartSpan(ctx, "cache.Set",
			tracing.WithSpanKind(tracing.SpanKindClient),
			tracing.WithAttributes(
				attribute.String("cache.key", key),
				attribute.String("cache.operation", "set"),
				attribute.Int64("cache.ttl_seconds", int64(expiration.Seconds())),
			),
		)
		defer span.End()
	}

	startTime := time.Now()
	tenantID, _, _ := r.getTenantInfo(ctx)

	var err error
	err = r.executeWithCircuitBreaker(ctx, func() error {
		data, serErr := r.serializeValue(ctx, value)
		if serErr != nil {
			return serErr
		}

		finalKey, keyErr := r.buildKey(ctx, key)
		if keyErr != nil {
			return keyErr
		}

		if span != nil {
			span.SetAttributes(
				attribute.String("cache.final_key", finalKey),
				attribute.Int("cache.value_size_bytes", len(data)),
			)
		}

		if setErr := r.client.Set(ctx, finalKey, data, expiration).Err(); setErr != nil {
			return sharedErrors.NewBusinessError("CACHE_SET_FAILED", "Failed to set value in cache").
				WithHTTPStatus(500).
				WithCategory(sharedErrors.CategorySystem).
				WithDetail("error", setErr.Error()).
				WithDetail("key", key)
		}
		return nil
	})

	r.updateStats("set", err, startTime)
	r.recordMetrics(ctx, "set", err, tenantID, startTime)

	if err != nil {
		r.recordError(ctx, span, "set", err, tenantID, startTime)
		return err
	}

	if span != nil {
		span.SetStatus(codes.Ok, "Value cached")
	}

	if r.logger != nil {
		r.logger.Debug("Cache set", logger.Fields{
			"key":        key,
			"tenant_id":  tenantID,
			"expiration": expiration.String(),
		})
	}

	return nil
}

// Delete removes a value from Redis cache (tenant-aware).
func (r *redisClient) Delete(ctx context.Context, key string) error {
	// Start tracing span
	var span tracing.Span
	if r.tracer != nil {
		ctx, span = r.tracer.StartSpan(ctx, "cache.Delete",
			tracing.WithSpanKind(tracing.SpanKindClient),
			tracing.WithAttributes(
				attribute.String("cache.key", key),
				attribute.String("cache.operation", "delete"),
			),
		)
		defer span.End()
	}

	startTime := time.Now()
	tenantID, _, _ := r.getTenantInfo(ctx)

	var err error
	err = r.executeWithCircuitBreaker(ctx, func() error {
		finalKey, keyErr := r.buildKey(ctx, key)
		if keyErr != nil {
			return keyErr
		}

		if delErr := r.client.Del(ctx, finalKey).Err(); delErr != nil {
			return sharedErrors.NewBusinessError("CACHE_DELETE_FAILED", "Failed to delete value from cache").
				WithHTTPStatus(500).
				WithCategory(sharedErrors.CategorySystem).
				WithDetail("error", delErr.Error()).
				WithDetail("key", key)
		}
		return nil
	})

	r.updateStats("delete", err, startTime)
	r.recordMetrics(ctx, "delete", err, tenantID, startTime)

	if err != nil {
		r.recordError(ctx, span, "delete", err, tenantID, startTime)
		return err
	}

	if span != nil {
		span.SetStatus(codes.Ok, "Value deleted")
	}

	if r.logger != nil {
		r.logger.Debug("Cache delete", logger.Fields{
			"key":       key,
			"tenant_id": tenantID,
		})
	}

	return nil
}

// Flush clears all cache entries for the current tenant.
func (r *redisClient) Flush(ctx context.Context) error {
	// Start tracing span
	var span tracing.Span
	if r.tracer != nil {
		ctx, span = r.tracer.StartSpan(ctx, "cache.Flush",
			tracing.WithSpanKind(tracing.SpanKindClient),
			tracing.WithAttributes(
				attribute.String("cache.operation", "flush"),
			),
		)
		defer span.End()
	}

	tenantID, idType, err := r.getTenantInfo(ctx)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to get tenant info for flush", logger.Fields{"error": err.Error()})
		}
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	if tenantID == "" {
		err := ErrNoTenantContext
		if r.logger != nil {
			r.logger.Error("No tenant context for flush operation")
		}
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	pattern := fmt.Sprintf("%s:%s:%s-%s:*", r.config.KeyPrefix, tenantKeyPrefix, idType, tenantID)

	if span != nil {
		span.SetAttributes(
			attribute.String("cache.pattern", pattern),
			attribute.String("cache.tenant_id", tenantID),
		)
	}

	if r.logger != nil {
		r.logger.Info("Flushing tenant cache", logger.Fields{
			"tenant_id": tenantID,
			"pattern":   pattern,
		})
	}

	err = r.deleteByPattern(ctx, pattern)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to flush cache", logger.Fields{
				"tenant_id": tenantID,
				"error":     err.Error(),
			})
		}
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	if span != nil {
		span.SetStatus(codes.Ok, "Cache flushed")
	}

	return nil
}

// MGet retrieves multiple values efficiently (tenant-aware).
func (r *redisClient) MGet(ctx context.Context, keys []string) ([]Result, error) {
	// Start tracing span
	var span tracing.Span
	if r.tracer != nil {
		ctx, span = r.tracer.StartSpan(ctx, "cache.MGet",
			tracing.WithSpanKind(tracing.SpanKindClient),
			tracing.WithAttributes(
				attribute.String("cache.operation", "mget"),
				attribute.Int("cache.key_count", len(keys)),
			),
		)
		defer span.End()
	}

	if len(keys) == 0 {
		return []Result{}, nil
	}

	tenantID, _, _ := r.getTenantInfo(ctx)

	finalKeys := make([]string, len(keys))
	for i, key := range keys {
		fk, err := r.buildKey(ctx, key)
		if err != nil {
			if r.logger != nil {
				r.logger.Error("Failed to build key in MGet", logger.Fields{
					"key":   key,
					"error": err.Error(),
				})
			}
			return nil, err
		}
		finalKeys[i] = fk
	}

	results := make([]Result, len(keys))
	err := r.executeWithCircuitBreaker(ctx, func() error {
		vals, err := r.client.MGet(ctx, finalKeys...).Result()
		if err != nil {
			return sharedErrors.NewBusinessError("CACHE_MGET_FAILED", "Failed to get multiple values from cache").
				WithHTTPStatus(500).
				WithCategory(sharedErrors.CategorySystem).
				WithDetail("error", err.Error())
		}

		for i, val := range vals {
			results[i].Key = keys[i]

			if val == nil {
				results[i].Err = ErrCacheMiss
				continue
			}

			strVal, ok := val.(string)
			if !ok {
				results[i].Err = fmt.Errorf("%w: unexpected type %T", ErrInvalidData, val)
				continue
			}

			var decoded any
			if err := r.deserializeValue([]byte(strVal), &decoded); err != nil {
				results[i].Err = err
				continue
			}

			results[i].Value = decoded
		}
		return nil
	})
	if err != nil {
		if r.logger != nil {
			r.logger.Error("MGet operation failed", logger.Fields{
				"tenant_id": tenantID,
				"key_count": len(keys),
				"error":     err.Error(),
			})
		}
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return nil, err
	}

	// Count hits and misses
	hits := 0
	for _, r := range results {
		if r.Err == nil {
			hits++
		}
	}

	if span != nil {
		span.SetAttributes(
			attribute.Int("cache.hits", hits),
			attribute.Int("cache.misses", len(keys)-hits),
		)
		span.SetStatus(codes.Ok, "Multi-get completed")
	}

	if r.logger != nil {
		r.logger.Debug("MGet operation completed", logger.Fields{
			"tenant_id": tenantID,
			"key_count": len(keys),
			"hits":      hits,
			"misses":    len(keys) - hits,
		})
	}

	return results, nil
}

// MSet sets multiple values with the same expiration (tenant-aware).
func (r *redisClient) MSet(ctx context.Context, pairs map[string]any, expiration time.Duration) error {
	// Start tracing span
	var span tracing.Span
	if r.tracer != nil {
		ctx, span = r.tracer.StartSpan(ctx, "cache.MSet",
			tracing.WithSpanKind(tracing.SpanKindClient),
			tracing.WithAttributes(
				attribute.String("cache.operation", "mset"),
				attribute.Int("cache.pair_count", len(pairs)),
				attribute.Int64("cache.ttl_seconds", int64(expiration.Seconds())),
			),
		)
		defer span.End()
	}

	if len(pairs) == 0 {
		return nil
	}

	tenantID, _, _ := r.getTenantInfo(ctx)

	err := r.executeWithCircuitBreaker(ctx, func() error {
		pipe := r.client.Pipeline()

		for key, value := range pairs {
			data, err := r.serializeValue(ctx, value)
			if err != nil {
				return sharedErrors.NewBusinessError("CACHE_MSET_SERIALIZE_FAILED", "Failed to serialize value in MSet").
					WithHTTPStatus(500).
					WithCategory(sharedErrors.CategorySystem).
					WithDetail("key", key).
					WithDetail("error", err.Error())
			}

			finalKey, keyErr := r.buildKey(ctx, key)
			if keyErr != nil {
				return keyErr
			}

			pipe.Set(ctx, finalKey, data, expiration)
		}

		if _, err := pipe.Exec(ctx); err != nil {
			return sharedErrors.NewBusinessError("CACHE_MSET_FAILED", "Failed to set multiple values in cache").
				WithHTTPStatus(500).
				WithCategory(sharedErrors.CategorySystem).
				WithDetail("error", err.Error())
		}
		return nil
	})
	if err != nil {
		if r.logger != nil {
			r.logger.Error("MSet operation failed", logger.Fields{
				"tenant_id":  tenantID,
				"pair_count": len(pairs),
				"error":      err.Error(),
			})
		}
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	if span != nil {
		span.SetStatus(codes.Ok, "Multi-set completed")
	}

	if r.logger != nil {
		r.logger.Debug("MSet operation completed", logger.Fields{
			"tenant_id":  tenantID,
			"pair_count": len(pairs),
		})
	}

	return nil
}

// MDelete deletes multiple keys efficiently (tenant-aware).
func (r *redisClient) MDelete(ctx context.Context, keys []string) error {
	// Start tracing span
	var span tracing.Span
	if r.tracer != nil {
		ctx, span = r.tracer.StartSpan(ctx, "cache.MDelete",
			tracing.WithSpanKind(tracing.SpanKindClient),
			tracing.WithAttributes(
				attribute.String("cache.operation", "mdelete"),
				attribute.Int("cache.key_count", len(keys)),
			),
		)
		defer span.End()
	}

	if len(keys) == 0 {
		return nil
	}

	tenantID, _, _ := r.getTenantInfo(ctx)

	finalKeys := make([]string, len(keys))
	for i, key := range keys {
		fk, err := r.buildKey(ctx, key)
		if err != nil {
			return err
		}
		finalKeys[i] = fk
	}

	err := r.executeWithCircuitBreaker(ctx, func() error {
		if err := r.client.Del(ctx, finalKeys...).Err(); err != nil {
			return sharedErrors.NewBusinessError("CACHE_MDELETE_FAILED", "Failed to delete multiple values from cache").
				WithHTTPStatus(500).
				WithCategory(sharedErrors.CategorySystem).
				WithDetail("error", err.Error())
		}
		return nil
	})
	if err != nil {
		if r.logger != nil {
			r.logger.Error("MDelete operation failed", logger.Fields{
				"tenant_id": tenantID,
				"key_count": len(keys),
				"error":     err.Error(),
			})
		}
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	if span != nil {
		span.SetStatus(codes.Ok, "Multi-delete completed")
	}

	if r.logger != nil {
		r.logger.Debug("MDelete operation completed", logger.Fields{
			"tenant_id": tenantID,
			"key_count": len(keys),
		})
	}

	return nil
}

// recordError logs and records error information.
func (r *redisClient) recordError(ctx context.Context, span tracing.Span, operation string, err error, tenantID string, startTime time.Time) {
	if err == nil {
		return
	}

	// Log error
	if r.logger != nil && err != ErrCacheMiss {
		r.logger.Error("Cache operation failed", logger.Fields{
			"operation": operation,
			"tenant_id": tenantID,
			"error":     err.Error(),
			"duration":  time.Since(startTime).String(),
		})
	}

	// Record in trace
	if span != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	// Record metrics
	if r.metricsService != nil && err != ErrCacheMiss {
		errorType := "unknown"
		if errors.Is(err, ErrCircuitOpen) {
			errorType = "circuit_open"
		} else if errors.Is(err, ErrInvalidData) {
			errorType = "invalid_data"
		}

		r.metricsService.IncrementCounter("errors_total", metrics.Fields{
			"operation":  operation,
			"error_type": errorType,
			"tenant_id":  tenantID,
		})
	}
}

// recordMetrics records operation metrics.
func (r *redisClient) recordMetrics(
	ctx context.Context,
	operation string,
	err error,
	tenantID string,
	startTime time.Time,
) {
	if r.metricsService == nil {
		return
	}
	_ = ctx
	// Calculate duration
	duration := time.Since(startTime).Seconds()

	// Determine result
	result := "success"
	if err != nil {
		if errors.Is(err, ErrCacheMiss) {
			result = "miss"
		} else {
			result = "error"
		}
	}

	// Record total operations
	r.metricsService.IncrementCounter(
		"operations_total",
		metrics.Fields{
			"operation": operation,
			"result":    result,
			"tenant_id": tenantID,
		},
	)

	// Record hits / misses
	if operation == "get" {
		switch {
		case err == nil:
			r.metricsService.IncrementCounter(
				"hits_total",
				metrics.Fields{"tenant_id": tenantID},
			)
		case errors.Is(err, ErrCacheMiss):
			r.metricsService.IncrementCounter(
				"misses_total",
				metrics.Fields{"tenant_id": tenantID},
			)
		}
	}

	// Record duration
	r.metricsService.ObserveHistogram(
		"operation_duration_seconds",
		duration,
		metrics.Fields{
			"operation": operation,
			"tenant_id": tenantID,
		},
	)
}

// updateMemoryCacheMetrics updates memory cache size metrics.
func (r *redisClient) updateMemoryCacheMetrics() {
	if r.metricsService == nil {
		return
	}

	r.memoryCacheMu.RLock()
	memoryCacheSize := len(r.memoryCache)
	r.memoryCacheMu.RUnlock()

	r.globalCacheMu.RLock()
	globalCacheSize := len(r.globalMemoryCache)
	r.globalCacheMu.RUnlock()

	r.metricsService.SetGauge("memory_cache_size", float64(memoryCacheSize), metrics.Fields{"type": "tenant"})
	r.metricsService.SetGauge("memory_cache_size", float64(globalCacheSize), metrics.Fields{"type": "global"})
}

// DeletePattern deletes keys matching a pattern (tenant-aware).
func (r *redisClient) DeletePattern(ctx context.Context, pattern string) error {
	// Start tracing span
	var span tracing.Span
	if r.tracer != nil {
		ctx, span = r.tracer.StartSpan(ctx, "cache.DeletePattern",
			tracing.WithSpanKind(tracing.SpanKindClient),
			tracing.WithAttributes(
				attribute.String("cache.pattern", pattern),
				attribute.String("cache.operation", "delete_pattern"),
			),
		)
		defer span.End()
	}

	tenantID, _, _ := r.getTenantInfo(ctx)

	finalPattern, err := r.buildKey(ctx, pattern)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to build pattern key", logger.Fields{
				"pattern": pattern,
				"error":   err.Error(),
			})
		}
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	if span != nil {
		span.SetAttributes(attribute.String("cache.final_pattern", finalPattern))
	}

	if r.logger != nil {
		r.logger.Info("Deleting keys by pattern", logger.Fields{
			"pattern":       pattern,
			"final_pattern": finalPattern,
			"tenant_id":     tenantID,
		})
	}

	err = r.deleteByPattern(ctx, finalPattern)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to delete pattern", logger.Fields{
				"pattern": pattern,
				"error":   err.Error(),
			})
		}
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	if span != nil {
		span.SetStatus(codes.Ok, "Pattern deleted")
	}

	return nil
}

// Keys returns keys matching a pattern (tenant-aware).
func (r *redisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	// Start tracing span
	var span tracing.Span
	if r.tracer != nil {
		ctx, span = r.tracer.StartSpan(ctx, "cache.Keys",
			tracing.WithSpanKind(tracing.SpanKindClient),
			tracing.WithAttributes(
				attribute.String("cache.pattern", pattern),
				attribute.String("cache.operation", "keys"),
			),
		)
		defer span.End()
	}

	tenantID, _, _ := r.getTenantInfo(ctx)

	finalPattern, err := r.buildKey(ctx, pattern)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to build pattern key", logger.Fields{
				"pattern": pattern,
				"error":   err.Error(),
			})
		}
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return nil, err
	}

	var keys []string
	err = r.executeWithCircuitBreaker(ctx, func() error {
		iter := r.client.Scan(ctx, 0, finalPattern, 0).Iterator()

		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}

		return iter.Err()
	})
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to scan keys", logger.Fields{
				"pattern": pattern,
				"error":   err.Error(),
			})
		}
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return nil, err
	}

	if span != nil {
		span.SetAttributes(attribute.Int("cache.key_count", len(keys)))
		span.SetStatus(codes.Ok, "Keys retrieved")
	}

	if r.logger != nil {
		r.logger.Debug("Keys scan completed", logger.Fields{
			"pattern":   pattern,
			"key_count": len(keys),
			"tenant_id": tenantID,
		})
	}

	return keys, nil
}

// deleteByPattern internal method to delete by pattern in batches.
func (r *redisClient) deleteByPattern(ctx context.Context, pattern string) error {
	return r.executeWithCircuitBreaker(ctx, func() error {
		iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
		keys := make([]string, 0, r.config.BatchDeleteSize)
		deletedCount := 0

		for iter.Next(ctx) {
			keys = append(keys, iter.Val())

			if len(keys) >= r.config.BatchDeleteSize {
				if err := r.client.Del(ctx, keys...).Err(); err != nil {
					return sharedErrors.NewBusinessError("BATCH_DELETE_FAILED", "Failed to delete batch of keys").
						WithHTTPStatus(500).
						WithCategory(sharedErrors.CategorySystem).
						WithDetail("error", err.Error())
				}
				deletedCount += len(keys)
				keys = keys[:0]
			}
		}

		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return sharedErrors.NewBusinessError("FINAL_BATCH_DELETE_FAILED", "Failed to delete final batch").
					WithHTTPStatus(500).
					WithCategory(sharedErrors.CategorySystem).
					WithDetail("error", err.Error())
			}
			deletedCount += len(keys)
		}

		if err := iter.Err(); err != nil {
			return sharedErrors.NewBusinessError("SCAN_FAILED", "Failed to scan keys").
				WithHTTPStatus(500).
				WithCategory(sharedErrors.CategorySystem).
				WithDetail("error", err.Error())
		}

		if r.logger != nil {
			r.logger.Info("Pattern deletion completed", logger.Fields{
				"pattern":       pattern,
				"deleted_count": deletedCount,
			})
		}

		return nil
	})
}

// Exists checks if a key exists (tenant-aware).
func (r *redisClient) Exists(ctx context.Context, key string) (bool, error) {
	// Start tracing span
	var span tracing.Span
	if r.tracer != nil {
		ctx, span = r.tracer.StartSpan(ctx, "cache.Exists",
			tracing.WithSpanKind(tracing.SpanKindClient),
			tracing.WithAttributes(
				attribute.String("cache.key", key),
				attribute.String("cache.operation", "exists"),
			),
		)
		defer span.End()
	}

	tenantID, _, _ := r.getTenantInfo(ctx)

	var exists bool
	err := r.executeWithCircuitBreaker(ctx, func() error {
		finalKey, keyErr := r.buildKey(ctx, key)
		if keyErr != nil {
			return keyErr
		}

		count, err := r.client.Exists(ctx, finalKey).Result()
		if err != nil {
			return sharedErrors.NewBusinessError("EXISTS_CHECK_FAILED", "Failed to check key existence").
				WithHTTPStatus(500).
				WithCategory(sharedErrors.CategorySystem).
				WithDetail("error", err.Error()).
				WithDetail("key", key)
		}
		exists = count > 0
		return nil
	})
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Exists check failed", logger.Fields{
				"key":   key,
				"error": err.Error(),
			})
		}
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return false, err
	}

	if span != nil {
		span.SetAttributes(attribute.Bool("cache.exists", exists))
		span.SetStatus(codes.Ok, "Existence checked")
	}

	if r.logger != nil {
		r.logger.Debug("Exists check", logger.Fields{
			"key":       key,
			"exists":    exists,
			"tenant_id": tenantID,
		})
	}

	return exists, nil
}

// TTL returns the time to live for a key (tenant-aware).
// Returns -1 if key exists but has no expiration.
// Returns -2 if key does not exist.
func (r *redisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	// Start tracing span
	var span tracing.Span
	if r.tracer != nil {
		ctx, span = r.tracer.StartSpan(ctx, "cache.TTL",
			tracing.WithSpanKind(tracing.SpanKindClient),
			tracing.WithAttributes(
				attribute.String("cache.key", key),
				attribute.String("cache.operation", "ttl"),
			),
		)
		defer span.End()
	}

	tenantID, _, _ := r.getTenantInfo(ctx)

	var ttl time.Duration
	err := r.executeWithCircuitBreaker(ctx, func() error {
		finalKey, keyErr := r.buildKey(ctx, key)
		if keyErr != nil {
			return keyErr
		}

		result, err := r.client.TTL(ctx, finalKey).Result()
		if err != nil {
			return sharedErrors.NewBusinessError("TTL_CHECK_FAILED", "Failed to check key TTL").
				WithHTTPStatus(500).
				WithCategory(sharedErrors.CategorySystem).
				WithDetail("error", err.Error()).
				WithDetail("key", key)
		}
		ttl = result
		return nil
	})
	if err != nil {
		if r.logger != nil {
			r.logger.Error("TTL check failed", logger.Fields{
				"key":   key,
				"error": err.Error(),
			})
		}
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return 0, err
	}

	if span != nil {
		span.SetAttributes(attribute.Int64("cache.ttl_seconds", int64(ttl.Seconds())))
		span.SetStatus(codes.Ok, "TTL retrieved")
	}

	if r.logger != nil {
		r.logger.Debug("TTL check", logger.Fields{
			"key":       key,
			"ttl":       ttl.String(),
			"tenant_id": tenantID,
		})
	}

	return ttl, nil
}

// Expire sets a timeout on a key (tenant-aware).
func (r *redisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	// Start tracing span
	var span tracing.Span
	if r.tracer != nil {
		ctx, span = r.tracer.StartSpan(ctx, "cache.Expire",
			tracing.WithSpanKind(tracing.SpanKindClient),
			tracing.WithAttributes(
				attribute.String("cache.key", key),
				attribute.String("cache.operation", "expire"),
				attribute.Int64("cache.ttl_seconds", int64(expiration.Seconds())),
			),
		)
		defer span.End()
	}

	tenantID, _, _ := r.getTenantInfo(ctx)

	err := r.executeWithCircuitBreaker(ctx, func() error {
		finalKey, keyErr := r.buildKey(ctx, key)
		if keyErr != nil {
			return keyErr
		}

		if err := r.client.Expire(ctx, finalKey, expiration).Err(); err != nil {
			return sharedErrors.NewBusinessError("EXPIRE_FAILED", "Failed to set key expiration").
				WithHTTPStatus(500).
				WithCategory(sharedErrors.CategorySystem).
				WithDetail("error", err.Error()).
				WithDetail("key", key)
		}
		return nil
	})
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Expire operation failed", logger.Fields{
				"key":        key,
				"expiration": expiration.String(),
				"error":      err.Error(),
			})
		}
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	if span != nil {
		span.SetStatus(codes.Ok, "Expiration set")
	}

	if r.logger != nil {
		r.logger.Debug("Expire operation", logger.Fields{
			"key":        key,
			"expiration": expiration.String(),
			"tenant_id":  tenantID,
		})
	}

	return nil
}

// GetMemory retrieves a value from in-memory cache (tenant-aware).
func (r *redisClient) GetMemory(ctx context.Context, key string, dest any) error {
	// Start tracing span
	var span tracing.Span
	if r.tracer != nil {
		ctx, span = r.tracer.StartSpan(ctx, "cache.GetMemory",
			tracing.WithSpanKind(tracing.SpanKindInternal),
			tracing.WithAttributes(
				attribute.String("cache.key", key),
				attribute.String("cache.operation", "get_memory"),
			),
		)
		defer span.End()
	}

	if !r.config.EnableMemoryCache {
		err := sharedErrors.NewBusinessError("MEMORY_CACHE_DISABLED", "Memory cache is disabled").
			WithHTTPStatus(400).
			WithCategory(sharedErrors.CategoryBusiness)
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	if dest == nil {
		return ErrNilValue
	}

	memKey, err := r.buildMemoryKey(ctx, key)
	if err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	r.memoryCacheMu.RLock()
	entry, exists := r.memoryCache[memKey]
	r.memoryCacheMu.RUnlock()

	if !exists {
		if span != nil {
			span.SetStatus(codes.Ok, "Memory cache miss")
		}
		return ErrCacheMiss
	}

	// Check expiration
	if !entry.expiresAt.IsZero() && entry.expiresAt.Before(time.Now()) {
		r.memoryCacheMu.Lock()
		delete(r.memoryCache, memKey)
		r.memoryCacheMu.Unlock()

		if span != nil {
			span.SetStatus(codes.Ok, "Memory cache expired")
		}
		return ErrCacheMiss
	}

	// Copy value to dest
	data, err := json.Marshal(entry.value)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to marshal memory value", logger.Fields{"error": err.Error()})
		}
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return sharedErrors.NewBusinessError("MARSHAL_FAILED", "Failed to marshal memory cache value").
			WithHTTPStatus(500).
			WithCategory(sharedErrors.CategorySystem).
			WithDetail("error", err.Error())
	}

	if err := json.Unmarshal(data, dest); err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to unmarshal memory value", logger.Fields{"error": err.Error()})
		}
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return sharedErrors.NewBusinessError("UNMARSHAL_FAILED", "Failed to unmarshal memory cache value").
			WithHTTPStatus(500).
			WithCategory(sharedErrors.CategorySystem).
			WithDetail("error", err.Error())
	}

	if span != nil {
		span.SetStatus(codes.Ok, "Memory cache hit")
	}

	if r.logger != nil {
		tenantID, _, _ := r.getTenantInfo(ctx)
		r.logger.Debug("Memory cache hit", logger.Fields{
			"key":       key,
			"tenant_id": tenantID,
		})
	}

	return nil
}

// SetMemory stores a value in in-memory cache (tenant-aware).
func (r *redisClient) SetMemory(ctx context.Context, key string, value any, expiration time.Duration) error {
	// Start tracing span
	var span tracing.Span
	if r.tracer != nil {
		ctx, span = r.tracer.StartSpan(ctx, "cache.SetMemory",
			tracing.WithSpanKind(tracing.SpanKindInternal),
			tracing.WithAttributes(
				attribute.String("cache.key", key),
				attribute.String("cache.operation", "set_memory"),
				attribute.Int64("cache.ttl_seconds", int64(expiration.Seconds())),
			),
		)
		defer span.End()
	}

	if !r.config.EnableMemoryCache {
		err := sharedErrors.NewBusinessError("MEMORY_CACHE_DISABLED", "Memory cache is disabled").
			WithHTTPStatus(400).
			WithCategory(sharedErrors.CategoryBusiness)
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	if value == nil {
		return ErrNilValue
	}

	memKey, err := r.buildMemoryKey(ctx, key)
	if err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	var expiresAt time.Time
	if expiration > 0 {
		expiresAt = time.Now().Add(expiration)
	}

	entry := &memoryEntry{
		value:     value,
		expiresAt: expiresAt,
	}

	r.memoryCacheMu.Lock()
	defer r.memoryCacheMu.Unlock()

	// Check size limit
	if len(r.memoryCache) >= r.config.MemoryCacheMaxSize {
		// Simple eviction: remove first entry (could be improved with LRU)
		for k := range r.memoryCache {
			delete(r.memoryCache, k)
			break
		}
	}

	r.memoryCache[memKey] = entry

	if span != nil {
		span.SetStatus(codes.Ok, "Value stored in memory")
	}

	if r.logger != nil {
		tenantID, _, _ := r.getTenantInfo(ctx)
		r.logger.Debug("Memory cache set", logger.Fields{
			"key":        key,
			"tenant_id":  tenantID,
			"expiration": expiration.String(),
		})
	}

	r.updateMemoryCacheMetrics()

	return nil
}

// DeleteMemory removes a value from in-memory cache (tenant-aware).
func (r *redisClient) DeleteMemory(ctx context.Context, key string) error {
	// Start tracing span
	var span tracing.Span
	if r.tracer != nil {
		ctx, span = r.tracer.StartSpan(ctx, "cache.DeleteMemory",
			tracing.WithSpanKind(tracing.SpanKindInternal),
			tracing.WithAttributes(
				attribute.String("cache.key", key),
				attribute.String("cache.operation", "delete_memory"),
			),
		)
		defer span.End()
	}

	if !r.config.EnableMemoryCache {
		err := sharedErrors.NewBusinessError("MEMORY_CACHE_DISABLED", "Memory cache is disabled").
			WithHTTPStatus(400).
			WithCategory(sharedErrors.CategoryBusiness)
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	memKey, err := r.buildMemoryKey(ctx, key)
	if err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	r.memoryCacheMu.Lock()
	delete(r.memoryCache, memKey)
	r.memoryCacheMu.Unlock()

	if span != nil {
		span.SetStatus(codes.Ok, "Value deleted from memory")
	}

	if r.logger != nil {
		tenantID, _, _ := r.getTenantInfo(ctx)
		r.logger.Debug("Memory cache delete", logger.Fields{
			"key":       key,
			"tenant_id": tenantID,
		})
	}

	r.updateMemoryCacheMetrics()

	return nil
}

// buildMemoryKey constructs a tenant-aware memory cache key.
func (r *redisClient) buildMemoryKey(ctx context.Context, key string) (string, error) {
	tenantID, idType, err := r.getTenantInfo(ctx)
	if err != nil {
		return "", err
	}

	if tenantID == "" {
		return "", ErrNoTenantContext
	}

	return fmt.Sprintf("%s-%s:%s", idType, tenantID, key), nil
}

// This contains the remaining implementation details including:
// - Global memory cache operations
// - Health checks
// - Serialization/deserialization
// - Circuit breaker
// - Helper functions
// - Context helpers

// GetGlobalMemory retrieves a value from global in-memory cache (not tenant-aware).
// Use this for shared data like formulas, configurations, etc.
func (r *redisClient) GetGlobalMemory(key string, dest any) error {
	if !r.config.EnableMemoryCache {
		return sharedErrors.NewBusinessError("MEMORY_CACHE_DISABLED", "Memory cache is disabled").
			WithHTTPStatus(400).
			WithCategory(sharedErrors.CategoryBusiness)
	}

	if dest == nil {
		return ErrNilValue
	}

	r.globalCacheMu.RLock()
	entry, exists := r.globalMemoryCache[key]
	r.globalCacheMu.RUnlock()

	if !exists {
		return ErrCacheMiss
	}

	// Check expiration
	if !entry.expiresAt.IsZero() && entry.expiresAt.Before(time.Now()) {
		r.globalCacheMu.Lock()
		delete(r.globalMemoryCache, key)
		r.globalCacheMu.Unlock()
		return ErrCacheMiss
	}

	// Copy value to dest
	data, err := json.Marshal(entry.value)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to marshal global memory value", logger.Fields{"error": err.Error()})
		}
		return sharedErrors.NewBusinessError("MARSHAL_FAILED", "Failed to marshal global memory value").
			WithHTTPStatus(500).
			WithCategory(sharedErrors.CategorySystem).
			WithDetail("error", err.Error())
	}

	if err := json.Unmarshal(data, dest); err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to unmarshal global memory value", logger.Fields{"error": err.Error()})
		}
		return sharedErrors.NewBusinessError("UNMARSHAL_FAILED", "Failed to unmarshal global memory value").
			WithHTTPStatus(500).
			WithCategory(sharedErrors.CategorySystem).
			WithDetail("error", err.Error())
	}

	if r.logger != nil {
		r.logger.Debug("Global memory cache hit", logger.Fields{"key": key})
	}

	return nil
}

// SetGlobalMemory stores a value in global in-memory cache (not tenant-aware).
// Use this for shared data like formulas, configurations, etc.
func (r *redisClient) SetGlobalMemory(key string, value any, expiration time.Duration) error {
	if !r.config.EnableMemoryCache {
		return sharedErrors.NewBusinessError("MEMORY_CACHE_DISABLED", "Memory cache is disabled").
			WithHTTPStatus(400).
			WithCategory(sharedErrors.CategoryBusiness)
	}

	if value == nil {
		return ErrNilValue
	}

	var expiresAt time.Time
	if expiration > 0 {
		expiresAt = time.Now().Add(expiration)
	}

	entry := &memoryEntry{
		value:     value,
		expiresAt: expiresAt,
	}

	r.globalCacheMu.Lock()
	defer r.globalCacheMu.Unlock()

	// Check size limit
	if len(r.globalMemoryCache) >= r.config.MemoryCacheMaxSize {
		// Simple eviction: remove first entry
		for k := range r.globalMemoryCache {
			delete(r.globalMemoryCache, k)
			break
		}
	}

	r.globalMemoryCache[key] = entry

	if r.logger != nil {
		r.logger.Debug("Global memory cache set", logger.Fields{
			"key":        key,
			"expiration": expiration.String(),
		})
	}

	r.updateMemoryCacheMetrics()

	return nil
}

// DeleteGlobalMemory removes a value from global in-memory cache (not tenant-aware).
func (r *redisClient) DeleteGlobalMemory(key string) error {
	if !r.config.EnableMemoryCache {
		return sharedErrors.NewBusinessError("MEMORY_CACHE_DISABLED", "Memory cache is disabled").
			WithHTTPStatus(400).
			WithCategory(sharedErrors.CategoryBusiness)
	}

	r.globalCacheMu.Lock()
	delete(r.globalMemoryCache, key)
	r.globalCacheMu.Unlock()

	if r.logger != nil {
		r.logger.Debug("Global memory cache delete", logger.Fields{"key": key})
	}

	r.updateMemoryCacheMetrics()

	return nil
}

// Ping tests the Redis connection.
func (r *redisClient) Ping(ctx context.Context) error {
	// Start tracing span
	var span tracing.Span
	if r.tracer != nil {
		ctx, span = r.tracer.StartSpan(ctx, "cache.Ping",
			tracing.WithSpanKind(tracing.SpanKindClient),
			tracing.WithAttributes(
				attribute.String("cache.operation", "ping"),
			),
		)
		defer span.End()
	}

	err := r.executeWithCircuitBreaker(ctx, func() error {
		if err := r.client.Ping(ctx).Err(); err != nil {
			return sharedErrors.NewBusinessError("PING_FAILED", "Failed to ping Redis").
				WithHTTPStatus(500).
				WithCategory(sharedErrors.CategorySystem).
				WithDetail("error", err.Error())
		}
		return nil
	})
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Ping failed", logger.Fields{"error": err.Error()})
		}
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	if span != nil {
		span.SetStatus(codes.Ok, "Ping successful")
	}

	if r.logger != nil {
		r.logger.Debug("Ping successful")
	}

	return nil
}

// Stats returns a snapshot of cache statistics.
func (r *redisClient) Stats() CacheStats {
	hits := r.hits.Load()
	misses := r.misses.Load()
	sets := r.sets.Load()
	deletes := r.deletes.Load()
	errors := r.errors.Load()
	opsCount := r.opsCount.Load()
	latencyNs := r.latencyNs.Load()

	var hitRatio float64
	if hits+misses > 0 {
		hitRatio = float64(hits) / float64(hits+misses)
	}

	var avgLatency time.Duration
	if opsCount > 0 {
		avgLatency = time.Duration(latencyNs / opsCount)
	}

	poolStats := r.client.PoolStats()

	r.statsMutex.RLock()
	lastError := r.lastError
	lastErrorTime := r.lastErrorTime
	r.statsMutex.RUnlock()

	// Get memory cache sizes
	r.memoryCacheMu.RLock()
	memoryCacheSize := len(r.memoryCache)
	r.memoryCacheMu.RUnlock()

	r.globalCacheMu.RLock()
	globalCacheSize := len(r.globalMemoryCache)
	r.globalCacheMu.RUnlock()

	stats := CacheStats{
		Hits:              hits,
		Misses:            misses,
		Sets:              sets,
		Deletes:           deletes,
		Errors:            errors,
		HitRatio:          hitRatio,
		AverageLatency:    avgLatency,
		ConnectionsActive: int(poolStats.TotalConns - poolStats.IdleConns),
		ConnectionsIdle:   int(poolStats.IdleConns),
		MemoryCacheSize:   memoryCacheSize,
		GlobalCacheSize:   globalCacheSize,
		LastError:         lastError,
		LastErrorTime:     lastErrorTime,
	}

	// Update connection metrics
	if r.metricsService != nil {
		r.metricsService.SetGauge("connections_active", float64(stats.ConnectionsActive), nil)
		r.metricsService.SetGauge("connections_idle", float64(stats.ConnectionsIdle), nil)
	}

	return stats
}

// Reset resets all statistics counters.
func (r *redisClient) Reset() {
	r.hits.Store(0)
	r.misses.Store(0)
	r.sets.Store(0)
	r.deletes.Store(0)
	r.errors.Store(0)
	r.opsCount.Store(0)
	r.latencyNs.Store(0)

	r.statsMutex.Lock()
	r.lastError = ""
	r.lastErrorTime = time.Time{}
	r.statsMutex.Unlock()

	if r.logger != nil {
		r.logger.Info("Cache statistics reset")
	}
}

// Close closes the Redis connection gracefully.
func (r *redisClient) Close() error {
	if r.logger != nil {
		r.logger.Info("Closing cache service")
	}

	// Stop cleanup routine
	close(r.stopCleanup)
	r.cleanupWg.Wait()

	if err := r.client.Close(); err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to close Redis connection", logger.Fields{"error": err.Error()})
		}
		return sharedErrors.NewBusinessError("CLOSE_FAILED", "Failed to close Redis connection").
			WithHTTPStatus(500).
			WithCategory(sharedErrors.CategorySystem).
			WithDetail("error", err.Error())
	}

	if r.logger != nil {
		r.logger.Info("Cache service closed successfully")
	}

	return nil
}

// shouldCompress determines if a value should be compressed.
func (r *redisClient) shouldCompress(ctx context.Context, data []byte) bool {
	if !r.compressionEnabled {
		return false
	}

	if enabled, ok := ctx.Value(CompressionKey).(bool); ok {
		return enabled && len(data) >= compressionThreshold
	}

	return len(data) >= compressionThreshold
}

// compressData compresses data using gzip.
func (r *redisClient) compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(compressionMarker)

	writer, err := gzip.NewWriterLevel(&buf, r.config.CompressionLevel)
	if err != nil {
		return nil, sharedErrors.NewBusinessError("COMPRESSION_INIT_FAILED", "Failed to create gzip writer").
			WithHTTPStatus(500).
			WithCategory(sharedErrors.CategorySystem).
			WithDetail("error", err.Error())
	}

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, sharedErrors.NewBusinessError("COMPRESSION_FAILED", "Failed to compress data").
			WithHTTPStatus(500).
			WithCategory(sharedErrors.CategorySystem).
			WithDetail("error", err.Error())
	}

	if err := writer.Close(); err != nil {
		return nil, sharedErrors.NewBusinessError("COMPRESSION_FINALIZE_FAILED", "Failed to finalize compression").
			WithHTTPStatus(500).
			WithCategory(sharedErrors.CategorySystem).
			WithDetail("error", err.Error())
	}

	return buf.Bytes(), nil
}

// decompressData decompresses gzip data.
func (r *redisClient) decompressData(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	if data[0] != compressionMarker {
		return data, nil
	}

	compressedData := data[1:]
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidData, err)
	}
	defer reader.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return nil, sharedErrors.NewBusinessError("DECOMPRESSION_FAILED", "Failed to decompress data").
			WithHTTPStatus(500).
			WithCategory(sharedErrors.CategorySystem).
			WithDetail("error", err.Error())
	}

	return buf.Bytes(), nil
}

// serializeValue serializes and optionally compresses a value.
func (r *redisClient) serializeValue(ctx context.Context, value any) ([]byte, error) {
	if value == nil {
		return nil, ErrNilValue
	}

	data, err := json.Marshal(value)
	if err != nil {
		return nil, sharedErrors.NewBusinessError("SERIALIZATION_FAILED", "Failed to marshal value").
			WithHTTPStatus(500).
			WithCategory(sharedErrors.CategorySystem).
			WithDetail("error", err.Error())
	}

	if r.shouldCompress(ctx, data) {
		return r.compressData(data)
	}

	return data, nil
}

// deserializeValue deserializes and optionally decompresses a value.
func (r *redisClient) deserializeValue(data []byte, dest any) error {
	if dest == nil {
		return ErrNilValue
	}

	decompressed, err := r.decompressData(data)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(decompressed, dest); err != nil {
		return sharedErrors.NewBusinessError("DESERIALIZATION_FAILED", "Failed to unmarshal value").
			WithHTTPStatus(500).
			WithCategory(sharedErrors.CategorySystem).
			WithDetail("error", err.Error())
	}

	return nil
}

// executeWithCircuitBreaker executes a Redis operation with circuit breaker protection.
func (r *redisClient) executeWithCircuitBreaker(ctx context.Context, operation func() error) error {
	if r.circuitBreaker == nil {
		return operation()
	}
	return r.circuitBreaker.execute(operation)
}

// updateStats updates cache statistics atomically.
func (r *redisClient) updateStats(operation string, err error, startTime time.Time) {
	latency := time.Since(startTime)
	r.latencyNs.Add(latency.Nanoseconds())
	r.opsCount.Add(1)

	switch operation {
	case "get":
		if err == ErrCacheMiss || errors.Is(err, ErrCacheMiss) {
			r.misses.Add(1)
		} else if err == nil {
			r.hits.Add(1)
		}
	case "set":
		if err == nil {
			r.sets.Add(1)
		}
	case "delete":
		if err == nil {
			r.deletes.Add(1)
		}
	}

	if err != nil && !errors.Is(err, ErrCacheMiss) {
		r.errors.Add(1)
		r.statsMutex.Lock()
		r.lastError = err.Error()
		r.lastErrorTime = time.Now()
		r.statsMutex.Unlock()
	}
}

// circuitBreaker implements the circuit breaker pattern.
type circuitBreaker struct {
	threshold    int
	timeout      time.Duration
	failureCount atomic.Int64
	lastFailure  atomic.Int64
	state        atomic.Int32
	mu           sync.Mutex
}

const (
	stateClosed   int32 = 0
	stateOpen     int32 = 1
	stateHalfOpen int32 = 2
)

func newCircuitBreaker(threshold int, timeout time.Duration) *circuitBreaker {
	return &circuitBreaker{
		threshold: threshold,
		timeout:   timeout,
	}
}

func (cb *circuitBreaker) execute(operation func() error) error {
	state := cb.state.Load()

	if state == stateOpen {
		lastFailureNs := cb.lastFailure.Load()
		lastFailure := time.Unix(0, lastFailureNs)

		if time.Since(lastFailure) > cb.timeout {
			cb.mu.Lock()
			if cb.state.Load() == stateOpen {
				cb.state.Store(stateHalfOpen)
			}
			cb.mu.Unlock()
		} else {
			return ErrCircuitOpen
		}
	}

	err := operation()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		failures := cb.failureCount.Add(1)
		cb.lastFailure.Store(time.Now().UnixNano())

		if failures >= int64(cb.threshold) {
			cb.state.Store(stateOpen)
		}
	} else {
		cb.failureCount.Store(0)
		cb.state.Store(stateClosed)
	}

	return err
}

// Context helper functions for tenant information.

// MustGetTenantID extracts tenant ID from context or panics.
func MustGetTenantID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok && id != uuid.Nil {
		return id
	}
	panic("cache: tenant_id not found in context")
}

// MustGetTenantSlug extracts tenant slug from context or panics.
func MustGetTenantSlug(ctx context.Context) string {
	if slug, ok := ctx.Value(TenantSlugKey).(string); ok && slug != "" {
		return slug
	}
	panic("cache: tenant_slug not found in context")
}

// MustGetTenantSubdomain extracts tenant subdomain from context or panics.
func MustGetTenantSubdomain(ctx context.Context) string {
	if subdomain, ok := ctx.Value(TenantSubdomainKey).(string); ok && subdomain != "" {
		return subdomain
	}
	panic("cache: tenant_subdomain not found in context")
}

// WithTenantID creates a context with tenant ID.
func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// WithTenantSlug creates a context with tenant slug.
func WithTenantSlug(ctx context.Context, tenantSlug string) context.Context {
	return context.WithValue(ctx, TenantSlugKey, tenantSlug)
}

// WithTenantSubdomain creates a context with tenant subdomain.
func WithTenantSubdomain(ctx context.Context, subdomain string) context.Context {
	return context.WithValue(ctx, TenantSubdomainKey, subdomain)
}

// WithNamespace creates a context with a custom namespace.
func WithNamespace(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, NamespaceKey, namespace)
}

// WithCompression creates a context with compression settings.
func WithCompression(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, CompressionKey, enabled)
}
