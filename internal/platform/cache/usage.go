package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// Example 1: Basic Usage (Backward Compatible)
func ExampleBasicUsage() {
	// Your existing code works exactly as before
	baseConfig := &config.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
	}

	cfg := DefaultRedisConfig(baseConfig)
	client, err := NewRedisClient(cfg)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx := context.Background()
	ctx = WithTenantID(ctx, uuid.New())

	// All existing operations work the same
	_ = client.Set(ctx, "user:123", map[string]string{"name": "John"}, 5*time.Minute)

	var user map[string]string
	_ = client.Get(ctx, "user:123", &user)

	_ = client.Delete(ctx, "user:123")

	fmt.Println("Basic usage completed")
}

// Example 2: Full Observability Stack
func ExampleWithFullObservability() error {
	// 1. Initialize logger
	logConfig := logger.DefaultConfig()
	logConfig.Level = logger.InfoLevel
	logConfig.Format = "json"
	logConfig.ServiceName = "cache-service"
	if err := logger.Initialize(logConfig); err != nil {
		return err
	}
	defer logger.Close()

	// 2. Initialize tracing
	tracingConfig := tracing.DefaultConfig()
	tracingConfig.ServiceName = "cache-service"
	tracingConfig.Environment = "production"
	tracingConfig.SamplingRate = 1.0
	tracingService, err := tracing.NewService(tracingConfig)
	if err != nil {
		return err
	}
	defer tracingService.Shutdown(context.Background())

	// 3. Initialize metrics
	metricsConfig := metrics.MetricsConfig{
		Provider:  "prometheus",
		Namespace: "erp",
		Subsystem: "cache",
		Enabled:   true,
	}
	metricsService, err := metrics.NewMetricsService(metricsConfig)
	if err != nil {
		return err
	}
	defer metricsService.Close()

	// 4. Create cache with full observability
	baseConfig := &config.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
	}

	cfg := DefaultRedisConfig(baseConfig)
	cfg.EnableTracing = true
	cfg.EnableMetrics = true
	cfg.EnableLogging = true

	log := logger.WithFields(logger.Fields{"component": "cache"})

	cacheClient, err := NewRedisClientWithObservability(
		cfg,
		log,
		tracingService,
		metricsService,
	)
	if err != nil {
		return err
	}
	defer cacheClient.Close()

	// 5. Use cache with automatic tracing, logging, and metrics
	ctx := context.Background()
	ctx = WithTenantID(ctx, uuid.New())

	// This operation will:
	// - Create a trace span
	// - Log debug/error messages
	// - Record metrics (hits, misses, latency)
	user := map[string]any{
		"id":    "123",
		"name":  "John Doe",
		"email": "john@example.com",
	}

	err = cacheClient.Set(ctx, "user:123", user, 10*time.Minute)
	if err != nil {
		log.Error("Failed to set cache", logger.Fields{"error": err.Error()})
		return err
	}

	var retrievedUser map[string]any
	err = cacheClient.Get(ctx, "user:123", &retrievedUser)
	if err != nil {
		log.Error("Failed to get cache", logger.Fields{"error": err.Error()})
		return err
	}

	log.Info("Cache operation successful", logger.Fields{
		"user_id": retrievedUser["id"],
		"name":    retrievedUser["name"],
	})

	// Get statistics
	stats := cacheClient.Stats()
	log.Info("Cache statistics", logger.Fields{
		"hits":        stats.Hits,
		"misses":      stats.Misses,
		"hit_ratio":   stats.HitRatio,
		"avg_latency": stats.AverageLatency.String(),
	})

	return nil
}

// Example 3: Multi-Tenant Usage
func ExampleMultiTenantOperations() error {
	baseConfig := &config.RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	cfg := DefaultRedisConfig(baseConfig)
	cfg.RequireTenantContext = true // Enforce tenant context
	cfg.AllowGlobalOperations = false

	log := logger.WithFields(logger.Fields{"component": "cache"})
	cacheClient, err := NewRedisClientWithObservability(cfg, log, nil, nil)
	if err != nil {
		return err
	}
	defer cacheClient.Close()

	// Tenant 1 operations
	tenant1ID := uuid.New()
	ctx1 := WithTenantID(context.Background(), tenant1ID)

	err = cacheClient.Set(ctx1, "config:theme", "dark", 1*time.Hour)
	if err != nil {
		return err
	}

	// Tenant 2 operations (isolated from Tenant 1)
	tenant2ID := uuid.New()
	ctx2 := WithTenantID(context.Background(), tenant2ID)

	err = cacheClient.Set(ctx2, "config:theme", "light", 1*time.Hour)
	if err != nil {
		return err
	}

	// Each tenant gets their own value
	var theme1, theme2 string
	_ = cacheClient.Get(ctx1, "config:theme", &theme1) // "dark"
	_ = cacheClient.Get(ctx2, "config:theme", &theme2) // "light"

	log.Info("Multi-tenant operations completed", logger.Fields{
		"tenant1_theme": theme1,
		"tenant2_theme": theme2,
	})

	// Using different tenant identifiers
	ctx3 := WithTenantSlug(context.Background(), "acme-corp")
	_ = cacheClient.Set(ctx3, "settings", map[string]any{"lang": "en"}, 1*time.Hour)

	ctx4 := WithTenantSubdomain(context.Background(), "startup")
	_ = cacheClient.Set(ctx4, "settings", map[string]any{"lang": "es"}, 1*time.Hour)

	return nil
}

// Example 4: Advanced Operations
func ExampleAdvancedOperations() error {
	baseConfig := &config.RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	cfg := DefaultRedisConfig(baseConfig)
	log := logger.WithFields(logger.Fields{"component": "cache"})

	cacheClient, err := NewRedisClientWithObservability(cfg, log, nil, nil)
	if err != nil {
		return err
	}
	defer cacheClient.Close()

	ctx := WithTenantID(context.Background(), uuid.New())

	// Batch operations
	users := map[string]any{
		"user:1": map[string]string{"name": "Alice"},
		"user:2": map[string]string{"name": "Bob"},
		"user:3": map[string]string{"name": "Charlie"},
	}

	err = cacheClient.MSet(ctx, users, 10*time.Minute)
	if err != nil {
		return err
	}

	// Retrieve multiple keys
	results, err := cacheClient.MGet(ctx, []string{"user:1", "user:2", "user:3"})
	if err != nil {
		return err
	}

	for _, result := range results {
		if result.Err == nil {
			log.Info("Retrieved user", logger.Fields{
				"key":   result.Key,
				"value": result.Value,
			})
		}
	}

	// Pattern operations
	keys, err := cacheClient.Keys(ctx, "user:*")
	if err != nil {
		return err
	}

	log.Info("Found keys", logger.Fields{"count": len(keys)})

	// Delete by pattern
	err = cacheClient.DeletePattern(ctx, "user:*")
	if err != nil {
		return err
	}

	// Check existence
	exists, err := cacheClient.Exists(ctx, "user:1")
	log.Info("Key exists", logger.Fields{"exists": exists})

	// Get TTL
	ttl, err := cacheClient.TTL(ctx, "user:1")
	log.Info("Key TTL", logger.Fields{"ttl": ttl.String()})

	return nil
}

// Example 5: Memory Cache Usage
func ExampleMemoryCacheOperations() error {
	baseConfig := &config.RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	cfg := DefaultRedisConfig(baseConfig)
	cfg.EnableMemoryCache = true
	cfg.MemoryCacheMaxSize = 1000

	log := logger.WithFields(logger.Fields{"component": "cache"})
	cacheClient, err := NewRedisClientWithObservability(cfg, log, nil, nil)
	if err != nil {
		return err
	}
	defer cacheClient.Close()

	ctx := WithTenantID(context.Background(), uuid.New())

	// Tenant-aware memory cache (hot data)
	hotConfig := map[string]any{
		"max_users":      100,
		"feature_flags":  []string{"beta", "premium"},
		"api_rate_limit": 1000,
	}

	err = cacheClient.SetMemory(ctx, "hot:config", hotConfig, 5*time.Minute)
	if err != nil {
		return err
	}

	var retrievedConfig map[string]any
	err = cacheClient.GetMemory(ctx, "hot:config", &retrievedConfig)
	if err != nil {
		return err
	}

	log.Info("Retrieved from memory cache", logger.Fields{
		"config": retrievedConfig,
	})

	// Global memory cache (shared across tenants)
	taxFormulas := map[string]string{
		"vat_standard": "amount * 0.20",
		"vat_reduced":  "amount * 0.05",
	}

	err = cacheClient.SetGlobalMemory("formulas:tax", taxFormulas, 1*time.Hour)
	if err != nil {
		return err
	}

	var retrievedFormulas map[string]string
	err = cacheClient.GetGlobalMemory("formulas:tax", &retrievedFormulas)
	if err != nil {
		return err
	}

	log.Info("Retrieved from global memory cache", logger.Fields{
		"formulas": retrievedFormulas,
	})

	return nil
}

// Example 6: Error Handling with Observability
func ExampleErrorHandling() error {
	baseConfig := &config.RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	cfg := DefaultRedisConfig(baseConfig)
	cfg.RequireTenantContext = true
	cfg.AllowGlobalOperations = false

	log := logger.WithFields(logger.Fields{"component": "cache"})
	tracingService := tracing.NewNoOpService()

	cacheClient, err := NewRedisClientWithObservability(cfg, log, tracingService, nil)
	if err != nil {
		return err
	}
	defer cacheClient.Close()

	// Missing tenant context - will log and fail
	ctx := context.Background()

	err = cacheClient.Set(ctx, "test", "value", 1*time.Minute)
	if err != nil {
		// Error is logged automatically
		// Trace span is marked as error
		// Metrics are recorded

		if errors.Is(err, ErrNoTenantContext) {
			log.Warn("Tenant context required", logger.Fields{
				"operation": "set",
			})
		}

		// Handle error appropriately
		return err
	}

	return nil
}

// Example 7: Using with HTTP Handlers
func ExampleHTTPHandler() {
	baseConfig := &config.RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	cfg := DefaultRedisConfig(baseConfig)
	log := logger.WithFields(logger.Fields{"component": "cache"})

	tracingConfig := tracing.DefaultConfig()
	tracingService, _ := tracing.NewService(tracingConfig)

	cacheClient, _ := NewRedisClientWithObservability(
		cfg,
		log,
		tracingService,
		nil,
	)
	defer cacheClient.Close()

	http.HandleFunc("/api/users/:id", func(w http.ResponseWriter, r *http.Request) {
		// Extract trace context from headers
		ctx := tracingService.ExtractHTTPHeaders(r.Context(), r.Header)

		// Extract tenant from request (middleware should set this)
		tenantID := uuid.MustParse(r.Header.Get("X-Tenant-ID"))
		ctx = WithTenantID(ctx, tenantID)

		// Start span for this operation
		ctx, span := tracingService.StartSpan(ctx, "GetUser")
		defer span.End()

		userID := r.PathValue("id")
		cacheKey := "user:" + userID

		// Try cache first
		var user map[string]any
		err := cacheClient.Get(ctx, cacheKey, &user)
		if err == nil {
			// Cache hit - traced and logged automatically
			span.SetAttributes(attribute.Bool("cache_hit", true))
			json.NewEncoder(w).Encode(user)
			return
		}

		// Cache miss - load from database
		span.SetAttributes(attribute.Bool("cache_hit", false))
		user, err = loadUserFromDB(ctx, userID)
		if err != nil {
			log.Error("Failed to load user", logger.Fields{
				"user_id": userID,
				"error":   err.Error(),
			})
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Cache the result
		_ = cacheClient.Set(ctx, cacheKey, user, 10*time.Minute)

		json.NewEncoder(w).Encode(user)
	})
}

// Example 8: Health Monitoring
func ExampleHealthMonitoring() error {
	baseConfig := &config.RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	cfg := DefaultRedisConfig(baseConfig)
	log := logger.WithFields(logger.Fields{"component": "cache"})

	metricsConfig := metrics.MetricsConfig{
		Provider:  "prometheus",
		Namespace: "erp",
		Subsystem: "cache",
		Enabled:   true,
	}
	metricsService, _ := metrics.NewMetricsService(metricsConfig)

	cacheClient, err := NewRedisClientWithObservability(
		cfg,
		log,
		nil,
		metricsService,
	)
	if err != nil {
		return err
	}
	defer cacheClient.Close()

	// Health check endpoint
	http.HandleFunc("/health/cache", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Ping Redis
		if err := cacheClient.Ping(ctx); err != nil {
			log.Error("Cache health check failed", logger.Fields{"error": err.Error()})
			http.Error(w, "Cache unavailable", http.StatusServiceUnavailable)
			return
		}

		// Get statistics
		stats := cacheClient.Stats()

		health := map[string]any{
			"status":            "healthy",
			"hit_ratio":         stats.HitRatio,
			"connections":       stats.ConnectionsActive,
			"memory_cache_size": stats.MemoryCacheSize,
			"errors":            stats.Errors,
		}

		// Check if hit ratio is too low
		if stats.HitRatio < 0.5 && stats.Hits+stats.Misses > 100 {
			health["status"] = "degraded"
			health["warning"] = "Low cache hit ratio"
		}

		json.NewEncoder(w).Encode(health)
	})

	// Metrics endpoint (Prometheus)
	http.Handle("/metrics", metricsService.Handler())

	return nil
}

// Helper functions for examples
func loadUserFromDB(ctx context.Context, userID string) (map[string]any, error) {
	// Simulate database load
	return map[string]any{
		"id":    userID,
		"name":  "John Doe",
		"email": "john@example.com",
	}, nil
}

func Usage() {
	// Run examples
	fmt.Println("=== Basic Usage ===")
	ExampleBasicUsage()

	fmt.Println("\n=== With Full Observability ===")
	if err := ExampleWithFullObservability(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Println("\n=== Multi-Tenant Operations ===")
	if err := ExampleMultiTenantOperations(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Println("\n=== Advanced Operations ===")
	if err := ExampleAdvancedOperations(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Println("\n=== Memory Cache Operations ===")
	if err := ExampleMemoryCacheOperations(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Println("\nAll examples completed!")
}
