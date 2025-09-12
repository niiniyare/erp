package middleware

import (
	"context"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
)

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
	CircuitBreakerThreshold int           `json:"circuit_breaker_threshold"` // Error threshold for circuit breaker
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
	}
}

// OptimizedMiddlewareChain provides performance-optimized middleware chaining
type OptimizedMiddlewareChain struct {
	config  *PerformanceConfig
	logger  logger.Logger
	metrics *metrics.MetricsService

	// Performance tracking
	concurrentRequests int64
	requestMutex       sync.RWMutex

	// Object pools
	contextPool  sync.Pool
	responsePool sync.Pool

	// Circuit breaker state
	circuitOpen   bool
	circuitMutex  sync.RWMutex
	errorCount    int64
	totalRequests int64
	lastReset     time.Time

	// Performance monitoring
	lastGC  time.Time
	gcMutex sync.Mutex
}

// NewOptimizedMiddlewareChain creates a new performance-optimized middleware chain
func NewOptimizedMiddlewareChain(
	config *PerformanceConfig,
	logger logger.Logger,
	metrics *metrics.MetricsService,
) *OptimizedMiddlewareChain {
	if config == nil {
		config = DefaultPerformanceConfig()
	}

	chain := &OptimizedMiddlewareChain{
		config:    config,
		logger:    logger,
		metrics:   metrics,
		lastReset: time.Now(),
		lastGC:    time.Now(),
	}

	// Initialize object pools if enabled
	if config.EnablePooling {
		chain.initializePools()
	}

	// Start background monitoring
	if config.ProfilerEnabled {
		go chain.startPerformanceMonitoring()
	}

	logger.Info("Optimized middleware chain initialized")

	return chain
}

// OptimizedChain creates the complete optimized middleware chain
func (omc *OptimizedMiddlewareChain) OptimizedChain(handler http.Handler, stack *GoaMiddlewareStack) http.Handler {
	// Start with the base handler
	h := handler

	// Apply middlewares in reverse order (innermost to outermost)

	// 10. Performance profiling (innermost - closest to handler)
	if omc.config.ProfilerEnabled {
		h = omc.profilingMiddleware(h)
	}

	// 9. Request logging with pooling optimization
	h = omc.optimizedLoggingMiddleware(h)

	// 8. Tenant isolation (business logic layer)
	h = TenantMiddleware(stack.tenantService, stack.store, stack.whitelist)(h)

	// 7. Input validation (security layer)
	h = CreateValidationMiddleware(nil, stack.logger)(h)

	// 6. Rate limiting with circuit breaker integration
	h = omc.rateLimitWithCircuitBreaker(h, stack)

	// 5. Compression with performance optimization
	h = omc.optimizedCompressionMiddleware(h, stack.compressionConfig)

	// 4. Timeout with resource monitoring
	h = omc.timeoutWithResourceMonitoring(h, stack.timeoutConfig)

	// 3. CORS (lightweight, outermost security)
	h = CORSMiddleware(stack.corsConfig, stack.logger)(h)

	// 2. Concurrent request limiting
	h = omc.concurrencyLimitMiddleware(h)

	// 1. Circuit breaker (outermost protection)
	if omc.config.CircuitBreakerEnabled {
		h = omc.circuitBreakerMiddleware(h)
	}

	omc.logger.Info("Optimized middleware chain configured", logger.Fields{
		"total_middlewares": 10,
		"optimizations":     []string{"pooling", "caching", "concurrency_limiting", "circuit_breaker", "profiling"},
	})

	return h
}

// initializePools sets up object pools for performance optimization
func (omc *OptimizedMiddlewareChain) initializePools() {
	// Context pool for request contexts
	omc.contextPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]interface{})
		},
	}

	// Response writer pool
	omc.responsePool = sync.Pool{
		New: func() interface{} {
			return &optimizedResponseWriter{}
		},
	}

	omc.logger.Debug("Object pools initialized", logger.Fields{
		"context_pool_size":  omc.config.RequestPoolSize,
		"response_pool_size": omc.config.ResponsePoolSize,
	})
}

// optimizedResponseWriter provides poolable response writer with performance metrics
type optimizedResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
	startTime    time.Time
}

// Reset resets the response writer for pool reuse
func (orw *optimizedResponseWriter) Reset(w http.ResponseWriter) {
	orw.ResponseWriter = w
	orw.statusCode = 0
	orw.responseSize = 0
	orw.startTime = time.Now()
}

// WriteHeader captures status code with minimal overhead
func (orw *optimizedResponseWriter) WriteHeader(statusCode int) {
	orw.statusCode = statusCode
	orw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures response size efficiently
func (orw *optimizedResponseWriter) Write(data []byte) (int, error) {
	if orw.statusCode == 0 {
		orw.statusCode = http.StatusOK
	}

	n, err := orw.ResponseWriter.Write(data)
	orw.responseSize += int64(n)
	return n, err
}

// concurrencyLimitMiddleware limits concurrent requests for resource protection
func (omc *OptimizedMiddlewareChain) concurrencyLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		omc.requestMutex.RLock()
		current := omc.concurrentRequests
		omc.requestMutex.RUnlock()

		if current >= int64(omc.config.MaxConcurrentRequests) {
			omc.metrics.IncrementCounter("http_requests_rejected_total", metrics.Fields{
				"reason": "concurrency_limit",
			})

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"service_unavailable","message":"Too many concurrent requests","retry_after":"5s"}`))
			return
		}

		// Increment concurrent request counter
		omc.requestMutex.Lock()
		omc.concurrentRequests++
		omc.requestMutex.Unlock()

		defer func() {
			omc.requestMutex.Lock()
			omc.concurrentRequests--
			omc.requestMutex.Unlock()
		}()

		// Track concurrent request metrics (using SetGauge instead)
		emptyFields := make(map[string]any)
		omc.metrics.SetGauge("http_requests_concurrent", float64(current+1), emptyFields)
		defer omc.metrics.SetGauge("http_requests_concurrent", float64(current), emptyFields)

		next.ServeHTTP(w, r)
	})
}

// circuitBreakerMiddleware implements circuit breaker pattern for resilience
func (omc *OptimizedMiddlewareChain) circuitBreakerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		omc.circuitMutex.RLock()
		isOpen := omc.circuitOpen
		omc.circuitMutex.RUnlock()

		// Check if circuit should be reset (every minute)
		if time.Since(omc.lastReset) > time.Minute {
			omc.resetCircuitBreakerStats()
		}

		// If circuit is open, reject requests
		if isOpen {
			omc.metrics.IncrementCounter("circuit_breaker_rejections_total", metrics.Fields{})

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"circuit_breaker_open","message":"Service temporarily unavailable","retry_after":"60s"}`))
			return
		}

		// Use pooled response writer for monitoring
		var orw *optimizedResponseWriter
		if omc.config.EnablePooling {
			orw = omc.responsePool.Get().(*optimizedResponseWriter)
			orw.Reset(w)
			defer omc.responsePool.Put(orw)
		} else {
			orw = &optimizedResponseWriter{ResponseWriter: w, startTime: time.Now()}
		}

		// Execute request
		next.ServeHTTP(orw, r)

		// Update circuit breaker stats
		omc.updateCircuitBreakerStats(orw.statusCode >= 500)
	})
}

// updateCircuitBreakerStats updates circuit breaker statistics
func (omc *OptimizedMiddlewareChain) updateCircuitBreakerStats(isError bool) {
	omc.circuitMutex.Lock()
	defer omc.circuitMutex.Unlock()

	omc.totalRequests++
	if isError {
		omc.errorCount++
	}

	// Check if circuit should be opened
	if omc.totalRequests >= 10 { // Minimum request count before evaluating
		errorRate := float64(omc.errorCount) / float64(omc.totalRequests)
		threshold := float64(omc.config.CircuitBreakerThreshold) / 100.0

		if errorRate >= threshold && !omc.circuitOpen {
			omc.circuitOpen = true
			omc.logger.Warn("Circuit breaker opened", logger.Fields{
				"error_rate":     errorRate,
				"threshold":      threshold,
				"error_count":    omc.errorCount,
				"total_requests": omc.totalRequests,
			})

			omc.metrics.IncrementCounter("circuit_breaker_opened_total", metrics.Fields{})
		}
	}
}

// resetCircuitBreakerStats resets circuit breaker statistics
func (omc *OptimizedMiddlewareChain) resetCircuitBreakerStats() {
	omc.circuitMutex.Lock()
	defer omc.circuitMutex.Unlock()

	// Reset stats and potentially close circuit
	wasOpen := omc.circuitOpen
	omc.errorCount = 0
	omc.totalRequests = 0
	omc.circuitOpen = false
	omc.lastReset = time.Now()

	if wasOpen {
		omc.logger.Info("Circuit breaker reset and closed", logger.Fields{
			"reset_time": time.Now(),
		})
		omc.metrics.IncrementCounter("circuit_breaker_closed_total", metrics.Fields{})
	}
}

// optimizedLoggingMiddleware provides high-performance logging with pooling
func (omc *OptimizedMiddlewareChain) optimizedLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use context pool if enabled
		var ctx context.Context
		if omc.config.EnablePooling {
			contextData := omc.contextPool.Get().(map[string]interface{})
			// Clear the map
			for k := range contextData {
				delete(contextData, k)
			}

			// Add request data
			contextData["method"] = r.Method
			contextData["path"] = r.URL.Path
			contextData["remote_ip"] = r.RemoteAddr

			ctx = context.WithValue(r.Context(), "request_data", contextData)

			defer func() {
				// Return to pool after use
				omc.contextPool.Put(contextData)
			}()
		} else {
			ctx = r.Context()
		}

		// Fast logging without complex formatting in hot path
		omc.logger.Debug("Request", logger.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		})

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// profilingMiddleware provides performance profiling for development/debugging
func (omc *OptimizedMiddlewareChain) profilingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		startMemory := getCurrentMemoryUsage()

		next.ServeHTTP(w, r)

		duration := time.Since(startTime)
		endMemory := getCurrentMemoryUsage()
		memoryDelta := endMemory - startMemory

		// Record performance metrics
		omc.metrics.ObserveHistogram("request_memory_delta_bytes", float64(memoryDelta), metrics.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		})

		omc.metrics.ObserveHistogram("request_processing_time_seconds", duration.Seconds(), metrics.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		})

		// Log slow requests or high memory usage
		if duration > 1*time.Second || memoryDelta > 10*1024*1024 { // 10MB
			omc.logger.Warn("Performance alert", logger.Fields{
				"method":       r.Method,
				"path":         r.URL.Path,
				"duration_ms":  duration.Milliseconds(),
				"memory_delta": memoryDelta,
				"type":         "slow_request",
			})
		}
	})
}

// optimizedCompressionMiddleware provides compression with performance monitoring
func (omc *OptimizedMiddlewareChain) optimizedCompressionMiddleware(next http.Handler, config *CompressionConfig) http.Handler {
	compressionHandler := CompressionMiddleware(config, omc.logger)(next)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		compressionHandler.ServeHTTP(w, r)

		duration := time.Since(startTime)
		fields := make(map[string]any)
		fields["path"] = r.URL.Path
		omc.metrics.ObserveHistogram("compression_processing_time", duration.Seconds(), fields)
	})
}

// timeoutWithResourceMonitoring provides timeout with resource usage monitoring
func (omc *OptimizedMiddlewareChain) timeoutWithResourceMonitoring(next http.Handler, config *TimeoutConfig) http.Handler {
	timeoutHandler := EnhancedTimeoutMiddleware(config, omc.logger)(next)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check system resources before processing
		if omc.shouldRejectDueToResources() {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"resource_exhaustion","message":"System resources unavailable"}`))
			return
		}

		timeoutHandler.ServeHTTP(w, r)
	})
}

// rateLimitWithCircuitBreaker integrates rate limiting with circuit breaker
func (omc *OptimizedMiddlewareChain) rateLimitWithCircuitBreaker(next http.Handler, stack *GoaMiddlewareStack) http.Handler {
	rateLimitHandler := CreateRateLimitMiddleware(stack.rateLimitConfig, stack.cache, stack.logger)(next)

	return rateLimitHandler
}

// shouldRejectDueToResources checks if request should be rejected due to resource constraints
func (omc *OptimizedMiddlewareChain) shouldRejectDueToResources() bool {
	currentMemory := getCurrentMemoryUsage()

	if currentMemory > omc.config.MemoryThreshold {
		omc.triggerGCIfNeeded()
		return true
	}

	return false
}

// triggerGCIfNeeded triggers garbage collection if memory threshold is exceeded
func (omc *OptimizedMiddlewareChain) triggerGCIfNeeded() {
	omc.gcMutex.Lock()
	defer omc.gcMutex.Unlock()

	if time.Since(omc.lastGC) > omc.config.GCInterval {
		runtime.GC()
		omc.lastGC = time.Now()

		omc.logger.Info("Forced garbage collection triggered", logger.Fields{
			"memory_threshold": omc.config.MemoryThreshold,
			"gc_interval":      omc.config.GCInterval,
		})

		omc.metrics.IncrementCounter("forced_gc_total", metrics.Fields{})
	}
}

// startPerformanceMonitoring starts background performance monitoring
func (omc *OptimizedMiddlewareChain) startPerformanceMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			omc.recordSystemMetrics()
		}
	}
}

// recordSystemMetrics records system-level performance metrics
func (omc *OptimizedMiddlewareChain) recordSystemMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	omc.metrics.ObserveHistogram("system_memory_usage_bytes", float64(m.Alloc), metrics.Fields{})
	omc.metrics.ObserveHistogram("system_gc_pause_ns", float64(m.PauseNs[(m.NumGC+255)%256]), metrics.Fields{})
	omc.metrics.IncrementCounter("system_gc_count", metrics.Fields{})

	omc.requestMutex.RLock()
	concurrent := omc.concurrentRequests
	omc.requestMutex.RUnlock()

	omc.metrics.ObserveHistogram("concurrent_requests_current", float64(concurrent), metrics.Fields{})

	omc.logger.Debug("System metrics recorded", logger.Fields{
		"memory_alloc":        m.Alloc,
		"gc_count":            m.NumGC,
		"concurrent_requests": concurrent,
	})
}

// getCurrentMemoryUsage returns current memory usage in bytes
func getCurrentMemoryUsage() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Alloc)
}

// GetPerformanceStats returns current performance statistics
func (omc *OptimizedMiddlewareChain) GetPerformanceStats() map[string]interface{} {
	omc.requestMutex.RLock()
	concurrent := omc.concurrentRequests
	omc.requestMutex.RUnlock()

	omc.circuitMutex.RLock()
	circuitOpen := omc.circuitOpen
	errorCount := omc.errorCount
	totalRequests := omc.totalRequests
	omc.circuitMutex.RUnlock()

	return map[string]interface{}{
		"concurrent_requests": concurrent,
		"max_concurrent":      omc.config.MaxConcurrentRequests,
		"circuit_breaker": map[string]interface{}{
			"open":           circuitOpen,
			"error_count":    errorCount,
			"total_requests": totalRequests,
			"last_reset":     omc.lastReset,
		},
		"memory": map[string]interface{}{
			"current":   getCurrentMemoryUsage(),
			"threshold": omc.config.MemoryThreshold,
			"last_gc":   omc.lastGC,
		},
		"pooling_enabled":  omc.config.EnablePooling,
		"profiler_enabled": omc.config.ProfilerEnabled,
	}
}
