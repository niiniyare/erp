package middleware

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
)

// NOTE: This file provides additional integration methods and helper utilities
// for the Fiber performance middleware. It reuses ALL types and functions from
// the main Fiber implementation to avoid duplication.

// IMPORTANT: Reusing these types from the main implementation:
// - PerformanceConfig
// - OptimizedMiddlewareChain
// - ResponseMetrics
// - TenantCache
// - All atomic fields and methods
// PerformanceConfig defines performance optimization settings for middleware stack
type PerformanceConfig struct {
	EnablePooling           bool          `json:"enable_pooling"`            // Enable object pooling
	EnableCaching           bool          `json:"enable_caching"`            // Enable middleware result caching
	MaxConcurrentRequests   int           `json:"max_concurrent_requests"`   // Global concurrent request limit
	RequestPoolSize         int           `json:"request_pool_size"`         // Size of request context pools
	ResponsePoolSize        int           `json:"response_pool_size"`        // Size of response writer pools
	MemoryThreshold         int64         `json:"memory_threshold"`          // Memory usage threshold (bytes)
	GCInterval              time.Duration `json:"gc_interval"`               // Forced GC interval
	ProfilerEnabled         bool          `json:"profiler_enabled"`          // Enable performance profiling
	CircuitBreakerEnabled   bool          `json:"circuit_breaker_enabled"`   // Enable circuit breaker
	CircuitBreakerThreshold int           `json:"circuit_breaker_threshold"` // Error threshold for circuit breaker (percentage)
	CircuitBreakerWindow    time.Duration `json:"circuit_breaker_window"`    // Time window for error rate calculation
	HealthCheckPath         string        `json:"health_check_path"`         // Health check endpoint path
}

// DefaultPerformanceConfig returns production-optimized performance settings
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		EnablePooling:           true,
		EnableCaching:           true,
		MaxConcurrentRequests:   1000,
		RequestPoolSize:         100,
		ResponsePoolSize:        100,
		MemoryThreshold:         1 << 30, // 1GB
		GCInterval:              5 * time.Minute,
		ProfilerEnabled:         os.Getenv("ENVIRONMENT") == "development",
		CircuitBreakerEnabled:   true,
		CircuitBreakerThreshold: 50, // 50% error rate triggers circuit breaker
		CircuitBreakerWindow:    1 * time.Minute,
		HealthCheckPath:         "/health",
	}
}

// OptimizedMiddlewareChain provides performance-optimized middleware chaining for Fiber
type OptimizedMiddlewareChain struct {
	config  *PerformanceConfig
	logger  logger.Logger
	metrics *metrics.MetricsService

	// Performance tracking (using atomic for thread-safety without locks)
	concurrentRequests atomic.Int64

	// Object pools
	contextPool  sync.Pool
	responsePool sync.Pool

	// Circuit breaker state
	circuitOpen   atomic.Bool
	errorCount    atomic.Int64
	totalRequests atomic.Int64
	lastReset     atomic.Int64 // Unix timestamp

	// Performance monitoring
	lastGC    atomic.Int64 // Unix timestamp
	gcTrigger sync.Once    // Ensures only one GC trigger at a time
}

// =============================================================================
// Alternative Middleware Stack Configuration
// =============================================================================

// SimplifiedMiddlewareStack provides a minimal dependency configuration
// Use this when you don't need the full FiberMiddlewareStack
type SimplifiedMiddlewareStack struct {
	Logger  logger.Logger
	Metrics *metrics.MetricsService
	// NOTE: Using interfaces to avoid circular dependencies
	// Production implementation should inject these via dependency injection
	TenantService any // Use tenant.Service interface from internal/core/tenant
	Store         any // Use db.Store interface from db/sqlc
	Cache         any // Use cache.Service interface from internal/platform/cache
}

// ApplyBasicMiddlewareChain applies essential middlewares without full stack dependencies
// This is useful for microservices or APIs that don't need all features
func (omc *OptimizedMiddlewareChain) ApplyBasicMiddlewareChain(app *fiber.App, stack *SimplifiedMiddlewareStack) {
	adapter := NewHTTPMiddlewareAdapter(omc.logger, omc.metrics)

	// 1. Recovery (outermost - catches all panics)
	app.Use(HTTPToFiberAdapter(adapter.RecoveryMiddleware()))

	// 2. Circuit breaker protection
	if omc.config.CircuitBreakerEnabled {
		app.Use(HTTPToFiberAdapter(adapter.CircuitBreakerMiddleware()))
	}

	// 3. Concurrency limiting
	app.Use(HTTPToFiberAdapter(adapter.ConcurrencyLimitMiddleware(omc.config.MaxConcurrentRequests)))

	// 4. Security headers
	app.Use(HTTPToFiberAdapter(adapter.SecurityHeadersMiddleware()))

	// 5. Request logging
	app.Use(HTTPToFiberAdapter(adapter.OptimizedLoggingMiddleware()))

	// 6. Performance profiling (if enabled)
	if omc.config.ProfilerEnabled {
		app.Use(HTTPToFiberAdapter(adapter.ProfilingMiddleware()))
	}

	omc.logger.Info("Basic middleware chain applied", logger.Fields{
		"middlewares": []string{"recovery", "circuit_breaker", "concurrency_limit", "security", "logging", "profiling"},
	})
}

// =============================================================================
// Extended Middleware Utilities
// =============================================================================

// RequestSizeLimitMiddleware limits request body size to prevent memory exhaustion
func (omc *OptimizedMiddlewareChain) RequestSizeLimitMiddleware(maxSize int64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check Content-Length header
		if contentLength := c.Request().Header.ContentLength(); contentLength > int(maxSize) {
			omc.metrics.IncrementCounter("request_size_exceeded_total", metrics.Fields{
				"path": c.Path(),
			})

			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error":         "request_too_large",
				"message":       "Request body exceeds maximum allowed size",
				"max_size":      maxSize,
				"received_size": contentLength,
			})
		}

		return c.Next()
	}
}

// CacheControlMiddleware adds cache control headers based on route patterns
type CacheControlConfig struct {
	// Map of path prefixes to cache durations
	Rules map[string]time.Duration
	// Default cache duration
	DefaultDuration time.Duration
	// Skip paths (no cache headers added)
	SkipPaths []string
}

func (omc *OptimizedMiddlewareChain) CacheControlMiddleware(config *CacheControlConfig) fiber.Handler {
	if config == nil {
		config = &CacheControlConfig{
			Rules:           make(map[string]time.Duration),
			DefaultDuration: 0, // No cache by default
		}
	}

	return func(c *fiber.Ctx) error {
		path := c.Path()

		// Check skip paths
		for _, skipPath := range config.SkipPaths {
			if strings.HasPrefix(path, skipPath) {
				return c.Next()
			}
		}

		// Find matching rule
		var cacheDuration time.Duration
		found := false
		for prefix, duration := range config.Rules {
			if strings.HasPrefix(path, prefix) {
				cacheDuration = duration
				found = true
				break
			}
		}

		if !found {
			cacheDuration = config.DefaultDuration
		}

		// Set cache headers
		if cacheDuration > 0 {
			c.Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(cacheDuration.Seconds())))
		} else {
			c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Set("Pragma", "no-cache")
			c.Set("Expires", "0")
		}

		return c.Next()
	}
}

// MetricsMiddleware provides detailed request metrics collection
func (omc *OptimizedMiddlewareChain) MetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		startTime := time.Now()
		path := c.Path()
		method := c.Method()

		// Track request start
		omc.metrics.IncrementCounter("http_requests_total", metrics.Fields{
			"method": method,
			"path":   path,
		})

		// Process request
		err := c.Next()

		duration := time.Since(startTime)
		statusCode := c.Response().StatusCode()
		responseSize := len(c.Response().Body())

		// Record detailed metrics
		omc.metrics.ObserveHistogram("http_request_duration_seconds", duration.Seconds(), metrics.Fields{
			"method": method,
			"path":   path,
			"status": statusCode,
		})

		omc.metrics.ObserveHistogram("http_response_size_bytes", float64(responseSize), metrics.Fields{
			"method": method,
			"path":   path,
		})

		// Track status code distribution
		statusClass := statusCode / 100
		omc.metrics.IncrementCounter("http_responses_total", metrics.Fields{
			"method":       method,
			"path":         path,
			"status":       statusCode,
			"status_class": fmt.Sprintf("%dxx", statusClass),
		})

		return err
	}
}

// CircuitBreakerMiddleware provides circuit breaker functionality for Fiber
func (omc *OptimizedMiddlewareChain) CircuitBreakerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if circuit is open
		if omc.circuitOpen.Load() {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":   "service_unavailable",
				"message": "Circuit breaker is open",
			})
		}

		// Increment concurrent requests
		current := omc.concurrentRequests.Add(1)
		defer omc.concurrentRequests.Add(-1)

		// Check concurrency limit
		if omc.config.MaxConcurrentRequests > 0 && current > int64(omc.config.MaxConcurrentRequests) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "too_many_requests",
				"message": "Concurrency limit exceeded",
			})
		}

		// Process request
		err := c.Next()

		// Update request counters
		omc.totalRequests.Add(1)
		if err != nil || c.Response().StatusCode() >= 500 {
			omc.errorCount.Add(1)
		}

		// Check if circuit breaker should open
		if omc.config.CircuitBreakerEnabled {
			omc.checkCircuitBreaker()
		}

		return err
	}
}

// ConcurrencyLimitMiddleware limits concurrent requests
func (omc *OptimizedMiddlewareChain) ConcurrencyLimitMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		current := omc.concurrentRequests.Add(1)
		defer omc.concurrentRequests.Add(-1)

		if omc.config.MaxConcurrentRequests > 0 && current > int64(omc.config.MaxConcurrentRequests) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "too_many_requests",
				"message": "Concurrency limit exceeded",
			})
		}

		return c.Next()
	}
}

// OptimizedLoggingMiddleware provides optimized request logging
func (omc *OptimizedMiddlewareChain) OptimizedLoggingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()

		// Process request
		err := c.Next()

		// Log request
		omc.logger.Info("HTTP request processed", logger.Fields{
			"method":      method,
			"path":        path,
			"status":      c.Response().StatusCode(),
			"duration_ms": time.Since(start).Milliseconds(),
			"ip":          c.IP(),
			"user_agent":  c.Get("User-Agent"),
		})

		return err
	}
}

// ProfilingMiddleware provides performance profiling
func (omc *OptimizedMiddlewareChain) ProfilingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !omc.config.ProfilerEnabled {
			return c.Next()
		}

		start := time.Now()

		// Add profiling headers
		c.Set("X-Profile-Start", start.Format(time.RFC3339Nano))

		err := c.Next()

		duration := time.Since(start)
		c.Set("X-Profile-Duration", duration.String())

		// Log slow requests
		if duration > 100*time.Millisecond {
			omc.logger.Warn("Slow request detected", logger.Fields{
				"path":        c.Path(),
				"method":      c.Method(),
				"duration_ms": duration.Milliseconds(),
			})
		}

		return err
	}
}

// checkCircuitBreaker checks if circuit breaker should open
func (omc *OptimizedMiddlewareChain) checkCircuitBreaker() {
	if !omc.config.CircuitBreakerEnabled {
		return
	}

	now := time.Now().Unix()
	lastReset := omc.lastReset.Load()

	// Check if window has expired
	if now-lastReset > int64(omc.config.CircuitBreakerWindow.Seconds()) {
		// Reset counters
		omc.errorCount.Store(0)
		omc.totalRequests.Store(0)
		omc.lastReset.Store(now)
		omc.circuitOpen.Store(false)
		return
	}

	total := omc.totalRequests.Load()
	errors := omc.errorCount.Load()

	if total > 10 { // Minimum requests before evaluating
		errorRate := float64(errors) / float64(total) * 100
		if errorRate > float64(omc.config.CircuitBreakerThreshold) {
			omc.circuitOpen.Store(true)
			omc.logger.Warn("Circuit breaker opened", logger.Fields{
				"error_rate":  errorRate,
				"threshold":   omc.config.CircuitBreakerThreshold,
				"total_reqs":  total,
				"error_count": errors,
			})
		}
	}
}

// =============================================================================
// Advanced Performance Monitoring
// =============================================================================

// PerformanceMonitor provides real-time performance monitoring
type PerformanceMonitor struct {
	chain   *OptimizedMiddlewareChain
	logger  logger.Logger
	metrics *metrics.MetricsService

	// Sliding window for request tracking
	requestWindow []RequestSnapshot
	windowSize    int
}

// RequestSnapshot captures a point-in-time request state
type RequestSnapshot struct {
	Timestamp      time.Time
	ConcurrentReqs int64
	MemoryUsage    int64
	GoroutineCount int
	CircuitOpen    bool
	ErrorRate      float64
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(
	chain *OptimizedMiddlewareChain,
	logger logger.Logger,
	metrics *metrics.MetricsService,
	windowSize int,
) *PerformanceMonitor {
	if windowSize <= 0 {
		windowSize = 100
	}

	pm := &PerformanceMonitor{
		chain:         chain,
		logger:        logger,
		metrics:       metrics,
		requestWindow: make([]RequestSnapshot, 0, windowSize),
		windowSize:    windowSize,
	}

	// Start monitoring goroutine
	go pm.monitorLoop()

	return pm
}

// monitorLoop continuously monitors system performance
func (pm *PerformanceMonitor) monitorLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		snapshot := pm.captureSnapshot()
		pm.addSnapshot(snapshot)
		pm.analyzePerformance(snapshot)
	}
}

// captureSnapshot captures current system state
func (pm *PerformanceMonitor) captureSnapshot() RequestSnapshot {
	concurrent := pm.chain.concurrentRequests.Load()
	errorCount := pm.chain.errorCount.Load()
	totalRequests := pm.chain.totalRequests.Load()

	var errorRate float64
	if totalRequests > 0 {
		errorRate = float64(errorCount) / float64(totalRequests)
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return RequestSnapshot{
		Timestamp:      time.Now(),
		ConcurrentReqs: concurrent,
		MemoryUsage:    int64(memStats.Alloc),
		GoroutineCount: runtime.NumGoroutine(),
		CircuitOpen:    pm.chain.circuitOpen.Load(),
		ErrorRate:      errorRate,
	}
}

// addSnapshot adds a snapshot to the sliding window
func (pm *PerformanceMonitor) addSnapshot(snapshot RequestSnapshot) {
	pm.requestWindow = append(pm.requestWindow, snapshot)

	// Keep window size limited
	if len(pm.requestWindow) > pm.windowSize {
		pm.requestWindow = pm.requestWindow[1:]
	}
}

// analyzePerformance analyzes performance trends and logs warnings
func (pm *PerformanceMonitor) analyzePerformance(snapshot RequestSnapshot) {
	if len(pm.requestWindow) < 2 {
		return
	}

	// Analyze trends
	memoryTrend := pm.calculateTrend(func(s RequestSnapshot) float64 {
		return float64(s.MemoryUsage)
	})

	goroutineTrend := pm.calculateTrend(func(s RequestSnapshot) float64 {
		return float64(s.GoroutineCount)
	})

	// Log warnings for concerning trends
	if memoryTrend > 0.1 { // Memory growing by >10%
		pm.logger.Warn("Memory usage trending upward", logger.Fields{
			"current_memory_mb": snapshot.MemoryUsage / 1024 / 1024,
			"trend":             memoryTrend,
		})
	}

	if goroutineTrend > 0.15 { // Goroutines growing by >15%
		pm.logger.Warn("Goroutine count trending upward", logger.Fields{
			"current_goroutines": snapshot.GoroutineCount,
			"trend":              goroutineTrend,
		})
	}

	if snapshot.ErrorRate > 0.05 { // >5% error rate
		pm.logger.Warn("High error rate detected", logger.Fields{
			"error_rate":   snapshot.ErrorRate,
			"circuit_open": snapshot.CircuitOpen,
		})
	}
}

// calculateTrend calculates the trend (rate of change) for a metric
func (pm *PerformanceMonitor) calculateTrend(extractor func(RequestSnapshot) float64) float64 {
	if len(pm.requestWindow) < 2 {
		return 0
	}

	first := extractor(pm.requestWindow[0])
	last := extractor(pm.requestWindow[len(pm.requestWindow)-1])

	if first == 0 {
		return 0
	}

	return (last - first) / first
}

// GetAnalytics returns performance analytics
func (pm *PerformanceMonitor) GetAnalytics() map[string]any {
	if len(pm.requestWindow) == 0 {
		return map[string]any{
			"status": "no_data",
		}
	}

	latest := pm.requestWindow[len(pm.requestWindow)-1]

	return map[string]any{
		"current_state": map[string]any{
			"concurrent_requests": latest.ConcurrentReqs,
			"memory_usage_mb":     latest.MemoryUsage / 1024 / 1024,
			"goroutine_count":     latest.GoroutineCount,
			"circuit_open":        latest.CircuitOpen,
			"error_rate":          latest.ErrorRate,
		},
		"trends": map[string]any{
			"memory_trend":    pm.calculateTrend(func(s RequestSnapshot) float64 { return float64(s.MemoryUsage) }),
			"goroutine_trend": pm.calculateTrend(func(s RequestSnapshot) float64 { return float64(s.GoroutineCount) }),
		},
		"window_size":     len(pm.requestWindow),
		"max_window_size": pm.windowSize,
	}
}

// =============================================================================
// Middleware Chain Builder (Fluent API)
// =============================================================================

// MiddlewareChainBuilder provides a fluent API for building middleware chains
type MiddlewareChainBuilder struct {
	chain  *OptimizedMiddlewareChain
	app    *fiber.App
	config []func(*fiber.App)
}

// NewMiddlewareChainBuilder creates a new builder
func NewMiddlewareChainBuilder(
	app *fiber.App,
	perfConfig *PerformanceConfig,
	logger logger.Logger,
	metrics *metrics.MetricsService,
) *MiddlewareChainBuilder {
	chain := &OptimizedMiddlewareChain{
		config:  perfConfig,
		logger:  logger,
		metrics: metrics,
	}

	return &MiddlewareChainBuilder{
		chain:  chain,
		app:    app,
		config: make([]func(*fiber.App), 0),
	}
}

// WithRecovery adds panic recovery middleware
func (b *MiddlewareChainBuilder) WithRecovery() *MiddlewareChainBuilder {
	b.config = append(b.config, func(app *fiber.App) {
		adapter := NewHTTPMiddlewareAdapter(b.chain.logger, b.chain.metrics)
		app.Use(HTTPToFiberAdapter(adapter.RecoveryMiddleware()))
	})
	return b
}

// WithSecurityHeaders adds security headers middleware
func (b *MiddlewareChainBuilder) WithSecurityHeaders() *MiddlewareChainBuilder {
	b.config = append(b.config, func(app *fiber.App) {
		adapter := NewHTTPMiddlewareAdapter(b.chain.logger, b.chain.metrics)
		app.Use(HTTPToFiberAdapter(adapter.SecurityHeadersMiddleware()))
	})
	return b
}

// WithRequestSizeLimit adds request size limiting
func (b *MiddlewareChainBuilder) WithRequestSizeLimit(maxSize int64) *MiddlewareChainBuilder {
	b.config = append(b.config, func(app *fiber.App) {
		app.Use(b.chain.RequestSizeLimitMiddleware(maxSize))
	})
	return b
}

// WithMetrics adds metrics collection middleware
func (b *MiddlewareChainBuilder) WithMetrics() *MiddlewareChainBuilder {
	b.config = append(b.config, func(app *fiber.App) {
		app.Use(b.chain.MetricsMiddleware())
	})
	return b
}

// WithCacheControl adds cache control middleware
func (b *MiddlewareChainBuilder) WithCacheControl(config *CacheControlConfig) *MiddlewareChainBuilder {
	b.config = append(b.config, func(app *fiber.App) {
		app.Use(b.chain.CacheControlMiddleware(config))
	})
	return b
}

// WithCircuitBreaker adds circuit breaker middleware
func (b *MiddlewareChainBuilder) WithCircuitBreaker() *MiddlewareChainBuilder {
	if b.chain.config.CircuitBreakerEnabled {
		b.config = append(b.config, func(app *fiber.App) {
			app.Use(b.chain.CircuitBreakerMiddleware())
		})
	}
	return b
}

// WithConcurrencyLimit adds concurrency limiting
func (b *MiddlewareChainBuilder) WithConcurrencyLimit() *MiddlewareChainBuilder {
	b.config = append(b.config, func(app *fiber.App) {
		app.Use(b.chain.ConcurrencyLimitMiddleware())
	})
	return b
}

// WithLogging adds optimized logging middleware
func (b *MiddlewareChainBuilder) WithLogging() *MiddlewareChainBuilder {
	b.config = append(b.config, func(app *fiber.App) {
		app.Use(b.chain.OptimizedLoggingMiddleware())
	})
	return b
}

// WithProfiling adds performance profiling
func (b *MiddlewareChainBuilder) WithProfiling() *MiddlewareChainBuilder {
	if b.chain.config.ProfilerEnabled {
		b.config = append(b.config, func(app *fiber.App) {
			app.Use(b.chain.ProfilingMiddleware())
		})
	}
	return b
}

// WithCORS adds CORS middleware with custom config
func (b *MiddlewareChainBuilder) WithCORS(config cors.Config) *MiddlewareChainBuilder {
	b.config = append(b.config, func(app *fiber.App) {
		app.Use(cors.New(config))
	})
	return b
}

// WithCompression adds compression middleware
func (b *MiddlewareChainBuilder) WithCompression(level int) *MiddlewareChainBuilder {
	b.config = append(b.config, func(app *fiber.App) {
		app.Use(compress.New(compress.Config{
			Level: compress.Level(level),
		}))
	})
	return b
}

// WithRateLimit adds rate limiting middleware
func (b *MiddlewareChainBuilder) WithRateLimit(max int, expiration time.Duration) *MiddlewareChainBuilder {
	b.config = append(b.config, func(app *fiber.App) {
		app.Use(limiter.New(limiter.Config{
			Max:        max,
			Expiration: expiration,
			KeyGenerator: func(c *fiber.Ctx) string {
				// Use tenant ID + IP for rate limiting
				tenantID := c.Locals(TenantIDKey)
				if tenantID != nil {
					return fmt.Sprintf("%v:%s", tenantID, c.IP())
				}
				return c.IP()
			},
		}))
	})
	return b
}

// WithTenantIsolation adds tenant isolation middleware
func (b *MiddlewareChainBuilder) WithTenantIsolation(config TenantMiddlewareConfig) *MiddlewareChainBuilder {
	b.config = append(b.config, func(app *fiber.App) {
		app.Use(TenantMiddleware(config))
	})
	return b
}

// WithCustomMiddleware adds a custom middleware
func (b *MiddlewareChainBuilder) WithCustomMiddleware(middleware fiber.Handler) *MiddlewareChainBuilder {
	b.config = append(b.config, func(app *fiber.App) {
		app.Use(middleware)
	})
	return b
}

// Build applies all configured middlewares to the app
func (b *MiddlewareChainBuilder) Build() *OptimizedMiddlewareChain {
	for _, configFunc := range b.config {
		configFunc(b.app)
	}
	return b.chain
}

// GetChain returns the underlying chain for direct access
func (b *MiddlewareChainBuilder) GetChain() *OptimizedMiddlewareChain {
	return b.chain
}

// =============================================================================
// Complete Example Usage
// =============================================================================

/*
Example 1: Using the Fluent Builder API

```go
package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"awo/internal/middleware"
)

func main() {
	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		BodyLimit:    4 * 1024 * 1024, // 4MB
	})

	// Create performance config
	perfConfig := middleware.DefaultPerformanceConfig()
	perfConfig.MaxConcurrentRequests = 500
	perfConfig.CircuitBreakerThreshold = 30 // 30% error rate

	// Build middleware chain
	builder := middleware.NewMiddlewareChainBuilder(app, perfConfig, logger, metrics)
	chain := builder.
		WithRecovery().
		WithCircuitBreaker().
		WithConcurrencyLimit().
		WithSecurityHeaders().
		WithRequestSizeLimit(10 * 1024 * 1024). // 10MB
		WithMetrics().
		WithLogging().
		WithProfiling().
		WithCORS(cors.Config{
			AllowOrigins: "https://example.com",
			AllowMethods: "GET,POST,PUT,DELETE",
		}).
		WithCompression(compress.LevelBestSpeed).
		WithRateLimit(100, 1*time.Minute).
		WithCacheControl(&middleware.CacheControlConfig{
			Rules: map[string]time.Duration{
				"/static/": 24 * time.Hour,
				"/api/":    0, // No cache
			},
		}).
		WithTenantIsolation(middleware.TenantMiddlewareConfig{
			TenantService: tenantService,
			Store:         store,
			Whitelist:     whitelist,
			EnableCache:   true,
			CacheTTL:      5 * time.Minute,
		}).
		Build()

	// Start performance monitoring
	monitor := middleware.NewPerformanceMonitor(chain, logger, metrics, 100)

	// Add monitoring endpoints
	app.Get("/metrics/performance", func(c *fiber.Ctx) error {
		return c.JSON(monitor.GetAnalytics())
	})

	app.Get("/metrics/chain", func(c *fiber.Ctx) error {
		return c.JSON(chain.GetPerformanceStats())
	})

	// Your application routes
	app.Get("/api/v1/users", handleGetUsers)
	app.Post("/api/v1/users", handleCreateUser)

	log.Fatal(app.Listen(":8080"))
}
```

Example 2: Using ApplyBasicMiddlewareChain (Minimal Setup)

```go
func main() {
	app := fiber.New()

	perfConfig := middleware.DefaultPerformanceConfig()
	chain := middleware.NewOptimizedMiddlewareChain(perfConfig, logger, metrics)

	stack := &middleware.SimplifiedMiddlewareStack{
		Logger:        logger,
		Metrics:       metrics,
		TenantService: tenantService,
		Store:         store,
		Cache:         cache,
	}

	// Apply basic middleware chain
	chain.ApplyBasicMiddlewareChain(app, stack)

	// Your routes
	app.Get("/api/health", handleHealth)
	app.Get("/api/users", handleUsers)

	log.Fatal(app.Listen(":8080"))
}
```

Example 3: Custom Middleware Integration

```go
func main() {
	app := fiber.New()

	builder := middleware.NewMiddlewareChainBuilder(app, nil, logger, metrics)

	// Add custom authentication middleware
	authMiddleware := func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing_token",
			})
		}
		// Validate token...
		return c.Next()
	}

	chain := builder.
		WithRecovery().
		WithLogging().
		WithCustomMiddleware(authMiddleware).
		WithMetrics().
		Build()

	app.Listen(":8080")
}
```
*/
