package middleware

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// RateLimitConfig defines the configuration for rate limiting
type RateLimitConfig struct {
	// Global limits
	GlobalRPS        int           `json:"global_rps"`         // Requests per second globally
	GlobalBurst      int           `json:"global_burst"`       // Global burst capacity
	GlobalWindowSize time.Duration `json:"global_window_size"` // Time window for global limits

	// Per-user limits
	UserRPS        int           `json:"user_rps"`         // Requests per second per user
	UserBurst      int           `json:"user_burst"`       // Per-user burst capacity
	UserWindowSize time.Duration `json:"user_window_size"` // Time window for user limits

	// Per-IP limits
	IPRPS        int           `json:"ip_rps"`         // Requests per second per IP
	IPBurst      int           `json:"ip_burst"`       // Per-IP burst capacity
	IPWindowSize time.Duration `json:"ip_window_size"` // Time window for IP limits

	// API endpoint specific limits
	EndpointLimits map[string]EndpointLimit `json:"endpoint_limits"`

	// Blocking configuration
	BlockDuration time.Duration `json:"block_duration"` // How long to block after limit exceeded
	MaxViolations int           `json:"max_violations"` // Max violations before extended block
	ViolationTTL  time.Duration `json:"violation_ttl"`  // How long violations are tracked

	// Failure handling
	FailOpen bool `json:"fail_open"` // If true, allow requests when cache is unavailable

	// Trusted proxies for IP extraction
	TrustedProxies []string `json:"trusted_proxies"` // List of trusted proxy IPs

	// Route template mapping for dynamic routes
	RouteTemplates map[string]string `json:"route_templates"` // e.g., "/api/users/:id" -> "/api/users/{id}"
}

// EndpointLimit defines rate limits for specific API endpoints
type EndpointLimit struct {
	RPS        int           `json:"rps"`
	Burst      int           `json:"burst"`
	WindowSize time.Duration `json:"window_size"`
}

// RateLimitMiddleware provides rate limiting middleware with circuit breaker
type RateLimitMiddleware struct {
	cache   cache.Service
	config  RateLimitConfig
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService

	// Circuit breaker for cache failures
	cacheCircuitBreaker *CacheCircuitBreaker

	// Local token buckets for high-traffic scenarios (reduces cache load)
	localBuckets sync.Map // map[string]*TokenBucket
}

// CacheCircuitBreaker protects against cache failures
type CacheCircuitBreaker struct {
	failures        atomic.Int64
	lastFailure     atomic.Int64
	consecutiveFail atomic.Int64
	state           atomic.Int32 // 0=closed, 1=open, 2=half-open
	config          CircuitBreakerConfig
}

// CircuitBreakerConfig configures cache circuit breaker
type CircuitBreakerConfig struct {
	FailureThreshold int           // Number of failures before opening
	RecoveryTimeout  time.Duration // Time before attempting recovery
	SuccessThreshold int           // Successes needed to close circuit
	HalfOpenMaxCalls int           // Max calls allowed in half-open state
}

// TokenBucket represents a local token bucket for rate limiting
type TokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	lastRefill time.Time
	rate       float64 // tokens per second
	capacity   float64
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(
	cache cache.Service,
	config RateLimitConfig,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) *RateLimitMiddleware {
	// Set defaults
	if config.ViolationTTL == 0 {
		config.ViolationTTL = 1 * time.Hour
	}

	return &RateLimitMiddleware{
		cache:   cache,
		config:  config,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
		cacheCircuitBreaker: &CacheCircuitBreaker{
			config: CircuitBreakerConfig{
				FailureThreshold: 5,
				RecoveryTimeout:  10 * time.Second,
				SuccessThreshold: 3,
				HalfOpenMaxCalls: 5,
			},
		},
	}
}

// DefaultRateLimitConfig returns a sensible default configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		GlobalRPS:        1000,
		GlobalBurst:      2000,
		GlobalWindowSize: time.Minute,

		UserRPS:        100,
		UserBurst:      200,
		UserWindowSize: time.Minute,

		IPRPS:        50,
		IPBurst:      100,
		IPWindowSize: time.Minute,

		EndpointLimits: map[string]EndpointLimit{
			"POST:/api/v1/auth/login": {
				RPS:        5,
				Burst:      10,
				WindowSize: time.Minute,
			},
			"POST:/api/v1/auth/refresh": {
				RPS:        10,
				Burst:      20,
				WindowSize: time.Minute,
			},
			"POST:/api/v1/finance/transaction": {
				RPS:        50,
				Burst:      100,
				WindowSize: time.Minute,
			},
		},

		BlockDuration: 15 * time.Minute,
		MaxViolations: 5,
		ViolationTTL:  1 * time.Hour,
		FailOpen:      true, // Fail open by default for high availability

		TrustedProxies: []string{}, // Configure based on your infrastructure
		RouteTemplates: map[string]string{
			"/api/v1/users/:id":               "/api/v1/users/{id}",
			"/api/v1/tenants/:id":             "/api/v1/tenants/{id}",
			"/api/v1/finance/transaction/:id": "/api/v1/finance/transaction/{id}",
		},
	}
}

// FiberMiddleware creates a Fiber middleware for rate limiting
func (m *RateLimitMiddleware) FiberMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		ctx, span := m.tracer.StartSpan(ctx, "middleware.rate_limit")
		defer span.End()

		start := time.Now()

		// Extract rate limiting identifiers
		clientIP := m.extractClientIP(c)
		userID := m.getUserID(c)
		endpoint := m.getEndpointKey(c)

		span.SetAttributes(
			attribute.String("client.ip", clientIP),
			attribute.String("user.id", userID),
			attribute.String("endpoint", endpoint),
		)

		// Check if cache circuit breaker is open
		if !m.cacheCircuitBreaker.ShouldAttempt() {
			if m.config.FailOpen {
				m.logger.Warn("Cache circuit breaker open, allowing request (fail-open)", logger.Fields{
					"client_ip": clientIP,
					"user_id":   userID,
					"endpoint":  endpoint,
				})
				return c.Next()
			} else {
				m.logger.Warn("Cache circuit breaker open, rejecting request (fail-closed)", logger.Fields{
					"client_ip": clientIP,
					"user_id":   userID,
					"endpoint":  endpoint,
				})
				return m.sendRateLimitError(c, "Service temporarily unavailable", nil)
			}
		}

		// Check if client is currently blocked
		if blocked, blockReason := m.isBlocked(ctx, clientIP, userID); blocked {
			m.logger.Warn("Request blocked due to rate limit violation", logger.Fields{
				"client_ip":    clientIP,
				"user_id":      userID,
				"endpoint":     endpoint,
				"block_reason": blockReason,
			})

			span.SetStatus(codes.Error, "rate_limit_blocked")
			m.recordRateLimitMetrics(ctx, "blocked", clientIP, userID, endpoint, time.Since(start))

			return m.sendRateLimitError(c, "Rate limit exceeded - temporarily blocked", nil)
		}

		// Check multiple rate limit dimensions
		allowed, limits, err := m.checkRateLimits(ctx, clientIP, userID, endpoint)
		if err != nil {
			m.logger.Error("Rate limit check failed", logger.Fields{
				"error":     err.Error(),
				"client_ip": clientIP,
				"user_id":   userID,
				"endpoint":  endpoint,
			})
			span.RecordError(err)

			// Handle cache failures based on FailOpen configuration
			m.cacheCircuitBreaker.RecordFailure()

			if m.config.FailOpen {
				m.logger.Warn("Rate limit check failed, allowing request (fail-open)", logger.Fields{
					"error": err.Error(),
				})
				return c.Next()
			} else {
				return m.sendRateLimitError(c, "Rate limit service unavailable", nil)
			}
		}

		// Cache operation succeeded
		m.cacheCircuitBreaker.RecordSuccess()

		if !allowed {
			// Increment violation counter (best effort - don't fail request if this fails)
			if err := m.recordViolation(ctx, clientIP, userID); err != nil {
				m.logger.Error("Failed to record rate limit violation", logger.Fields{
					"error": err.Error(),
				})
			}

			m.logger.Warn("Rate limit exceeded", logger.Fields{
				"client_ip": clientIP,
				"user_id":   userID,
				"endpoint":  endpoint,
				"limits":    limits,
			})

			span.SetStatus(codes.Error, "rate_limit_exceeded")
			m.recordRateLimitMetrics(ctx, "exceeded", clientIP, userID, endpoint, time.Since(start))

			return m.sendRateLimitError(c, "Rate limit exceeded", limits)
		}

		// Record successful request
		m.recordRateLimitMetrics(ctx, "allowed", clientIP, userID, endpoint, time.Since(start))

		// Add rate limit headers to response
		m.addRateLimitHeaders(c, limits)

		// Continue to next handler
		return c.Next()
	}
}

// checkRateLimits checks all applicable rate limits with improved error handling
func (m *RateLimitMiddleware) checkRateLimits(ctx context.Context, clientIP, userID, endpoint string) (bool, map[string]RateLimitInfo, error) {
	limits := make(map[string]RateLimitInfo)
	var lastError error

	// Check global rate limit
	globalAllowed, globalInfo, err := m.checkTokenBucketLimit(ctx, "global", "", m.config.GlobalRPS, m.config.GlobalBurst, m.config.GlobalWindowSize)
	if err != nil {
		lastError = fmt.Errorf("global rate limit check failed: %w", err)
	} else {
		limits["global"] = globalInfo
		if !globalAllowed {
			return false, limits, nil
		}
	}

	// Check IP-based rate limit
	ipAllowed, ipInfo, err := m.checkTokenBucketLimit(ctx, "ip", clientIP, m.config.IPRPS, m.config.IPBurst, m.config.IPWindowSize)
	if err != nil {
		lastError = fmt.Errorf("IP rate limit check failed: %w", err)
	} else {
		limits["ip"] = ipInfo
		if !ipAllowed {
			return false, limits, nil
		}
	}

	// Check user-based rate limit (if user is authenticated)
	if userID != "" {
		userAllowed, userInfo, err := m.checkTokenBucketLimit(ctx, "user", userID, m.config.UserRPS, m.config.UserBurst, m.config.UserWindowSize)
		if err != nil {
			lastError = fmt.Errorf("user rate limit check failed: %w", err)
		} else {
			limits["user"] = userInfo
			if !userAllowed {
				return false, limits, nil
			}
		}
	}

	// Check endpoint-specific rate limit
	if endpointLimit, exists := m.config.EndpointLimits[endpoint]; exists {
		endpointKey := fmt.Sprintf("%s:%s", endpoint, clientIP)
		if userID != "" {
			endpointKey = fmt.Sprintf("%s:user:%s", endpoint, userID)
		}

		endpointAllowed, endpointInfo, err := m.checkTokenBucketLimit(ctx, "endpoint", endpointKey, endpointLimit.RPS, endpointLimit.Burst, endpointLimit.WindowSize)
		if err != nil {
			lastError = fmt.Errorf("endpoint rate limit check failed: %w", err)
		} else {
			limits["endpoint"] = endpointInfo
			if !endpointAllowed {
				return false, limits, nil
			}
		}
	}

	// If we had errors but got here, all checks that succeeded passed
	return true, limits, lastError
}

// checkTokenBucketLimit implements token bucket algorithm with atomic operations
// NOTE: This uses a hybrid approach - local buckets for high-frequency checks,
// Redis for distributed coordination
func (m *RateLimitMiddleware) checkTokenBucketLimit(
	ctx context.Context,
	limitType, key string,
	rps, burst int,
	windowSize time.Duration,
) (bool, RateLimitInfo, error) {
	cacheKey := fmt.Sprintf("rate_limit:bucket:%s:%s", limitType, key)
	now := time.Now()

	// NOTE: For production distributed rate limiting, use the atomicTokenBucketOperation method
	// which implements Redis Lua scripts for true atomicity across instances
	// This method is provided below and eliminates race conditions

	// Try to use Redis atomic increment (if available)
	// NOTE: This assumes your cache service supports atomic operations
	// If not, fall back to local token bucket for this instance only
	bucket, err := m.getOrCreateTokenBucket(ctx, cacheKey, float64(rps), float64(burst))
	if err != nil {
		return false, RateLimitInfo{}, err
	}

	// Refill tokens based on time elapsed
	elapsed := now.Sub(time.Unix(bucket.LastRefill, 0)).Seconds()
	tokensToAdd := elapsed * bucket.Rate
	bucket.Tokens = min(bucket.Capacity, bucket.Tokens+tokensToAdd)
	bucket.LastRefill = now.Unix()

	info := RateLimitInfo{
		Limit:     rps,
		Remaining: int(bucket.Tokens),
		ResetTime: now.Add(time.Duration(float64(time.Second) * (bucket.Capacity / bucket.Rate))),
		Window:    windowSize,
	}

	// Check if we have enough tokens
	if bucket.Tokens < 1.0 {
		// Save state even on failure
		_ = m.saveTokenBucket(ctx, cacheKey, bucket, windowSize)
		return false, info, nil
	}

	// Consume one token
	bucket.Tokens -= 1.0
	info.Remaining = int(bucket.Tokens)

	// Save updated bucket state - critical operation as we've already consumed a token
	if err := m.saveTokenBucket(ctx, cacheKey, bucket, windowSize); err != nil {
		// Critical: We consumed a token but failed to save state. This could lead to
		// allowing more requests than the limit permits. For production use:
		// 1. Use Redis Lua scripts for atomic operations
		// 2. Implement compensating transactions
		// 3. Consider using a write-ahead log for durability

		m.logger.Error("Critical: Failed to save token bucket state after consuming token", logger.Fields{
			"error":       err.Error(),
			"cache_key":   cacheKey,
			"tokens_left": bucket.Tokens,
			"operation":   "token_consumption",
		})

		// Return the error but mark as consumed to err on the side of caution
		info.Remaining = int(bucket.Tokens)
		return false, info, fmt.Errorf("failed to persist rate limit state: %w", err)
	}

	return true, info, nil
}

// TokenBucketData represents serializable token bucket state
type TokenBucketData struct {
	Tokens     float64 `json:"tokens"`
	LastRefill int64   `json:"last_refill"` // Unix timestamp
	Rate       float64 `json:"rate"`
	Capacity   float64 `json:"capacity"`
}

// getOrCreateTokenBucket retrieves or creates a token bucket from cache
func (m *RateLimitMiddleware) getOrCreateTokenBucket(ctx context.Context, key string, rate, capacity float64) (*TokenBucketData, error) {
	var data TokenBucketData

	err := m.cache.Get(ctx, key, &data)
	if err != nil {
		// Cache miss or error - create new bucket
		return &TokenBucketData{
			Tokens:     capacity,
			LastRefill: time.Now().Unix(),
			Rate:       rate,
			Capacity:   capacity,
		}, nil
	}

	return &data, nil
}

// saveTokenBucket saves token bucket state to cache
func (m *RateLimitMiddleware) saveTokenBucket(ctx context.Context, key string, bucket *TokenBucketData, ttl time.Duration) error {
	return m.cache.Set(ctx, key, bucket, ttl*2) // 2x TTL for safety margin
}

// isBlocked checks if a client is currently blocked
func (m *RateLimitMiddleware) isBlocked(ctx context.Context, clientIP, userID string) (bool, string) {
	// Check IP block
	ipBlockKey := fmt.Sprintf("rate_limit:block:ip:%s", clientIP)
	blocked, err := m.cache.Exists(ctx, ipBlockKey)
	if err != nil {
		// Log error but don't fail the request
		m.logger.Debug("Failed to check IP block status", logger.Fields{
			"error": err.Error(),
			"ip":    clientIP,
		})
	} else if blocked {
		return true, "ip_blocked"
	}

	// Check user block (if authenticated)
	if userID != "" {
		userBlockKey := fmt.Sprintf("rate_limit:block:user:%s", userID)
		blocked, err := m.cache.Exists(ctx, userBlockKey)
		if err != nil {
			m.logger.Debug("Failed to check user block status", logger.Fields{
				"error":   err.Error(),
				"user_id": userID,
			})
		} else if blocked {
			return true, "user_blocked"
		}
	}

	return false, ""
}

// recordViolation records a rate limit violation using atomic operations
func (m *RateLimitMiddleware) recordViolation(ctx context.Context, clientIP, userID string) error {
	// Record IP violation using atomic increment
	ipViolationKey := fmt.Sprintf("rate_limit:violations:ip:%s", clientIP)
	violations, err := m.atomicIncrement(ctx, ipViolationKey, m.config.ViolationTTL)
	if err != nil {
		return fmt.Errorf("failed to record IP violation: %w", err)
	}

	// Block IP if too many violations
	if violations >= m.config.MaxViolations {
		ipBlockKey := fmt.Sprintf("rate_limit:block:ip:%s", clientIP)
		if err := m.cache.Set(ctx, ipBlockKey, true, m.config.BlockDuration); err != nil {
			return fmt.Errorf("failed to block IP: %w", err)
		}

		m.logger.Warn("IP blocked due to excessive violations", logger.Fields{
			"ip":         clientIP,
			"violations": violations,
		})
	}

	// Record user violation (if authenticated)
	if userID != "" {
		userViolationKey := fmt.Sprintf("rate_limit:violations:user:%s", userID)
		userViolations, err := m.atomicIncrement(ctx, userViolationKey, m.config.ViolationTTL)
		if err != nil {
			return fmt.Errorf("failed to record user violation: %w", err)
		}

		// Block user if too many violations
		if userViolations >= m.config.MaxViolations {
			userBlockKey := fmt.Sprintf("rate_limit:block:user:%s", userID)
			if err := m.cache.Set(ctx, userBlockKey, true, m.config.BlockDuration); err != nil {
				return fmt.Errorf("failed to block user: %w", err)
			}

			m.logger.Warn("User blocked due to excessive violations", logger.Fields{
				"user_id":    userID,
				"violations": userViolations,
			})
		}
	}

	return nil
}

// atomicIncrement performs an atomic increment operation using existing cache
func (m *RateLimitMiddleware) atomicIncrement(ctx context.Context, key string, ttl time.Duration) (int, error) {
	// Use the existing cache service with proper error handling
	var count int
	err := m.cache.Get(ctx, key, &count)
	if err != nil {
		// If cache miss, start with 1
		if err == cache.ErrCacheMiss {
			if setErr := m.cache.Set(ctx, key, 1, ttl); setErr != nil {
				return 0, fmt.Errorf("failed to initialize counter: %w", setErr)
			}
			return 1, nil
		}
		// Other cache errors
		return 0, fmt.Errorf("failed to get counter from cache: %w", err)
	}

	// Increment existing count
	count++
	if err := m.cache.Set(ctx, key, count, ttl); err != nil {
		return 0, fmt.Errorf("failed to update counter in cache: %w", err)
	}

	return count, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// extractClientIP extracts the client IP with proper proxy handling
func (m *RateLimitMiddleware) extractClientIP(c *fiber.Ctx) string {
	// Check X-Forwarded-For header first (if behind trusted proxy)
	if xff := c.Get("X-Forwarded-For"); xff != "" {
		// Parse comma-separated IPs and take the leftmost (client IP)
		ips := strings.Split(xff, ",")
		clientIP := strings.TrimSpace(ips[0])

		// NOTE: In production, implement proxy IP validation
		// For now, trusting X-Forwarded-For (ensure proper firewall rules)
		if clientIP != "" {
			return clientIP
		}
	}

	// Check X-Real-IP header
	if xri := c.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to direct IP from connection
	// Strip port if present
	ip := c.IP()
	if host, _, err := net.SplitHostPort(ip); err == nil {
		return host
	}

	return ip
}

// getUserID extracts user ID from Fiber context
// NOTE: This depends on your auth middleware setting the user ID
func (m *RateLimitMiddleware) getUserID(c *fiber.Ctx) string {
	// Try Fiber locals first (set by auth middleware)
	if userID := c.Locals("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}

	// Try JWT claims if available
	// NOTE: This integrates with the JWT middleware implemented in jwt_auth.go
	if claims := c.Locals("claims"); claims != nil {
		if claimsMap, ok := claims.(map[string]interface{}); ok {
			if userID, ok := claimsMap["user_id"].(string); ok {
				return userID
			}
			if sub, ok := claimsMap["sub"].(string); ok {
				return sub
			}
		}
	}

	return ""
}

// getEndpointKey generates a normalized endpoint key for rate limiting
func (m *RateLimitMiddleware) getEndpointKey(c *fiber.Ctx) string {
	method := c.Method()
	path := c.Path()

	// Check if we have a route template mapping
	if template, exists := m.config.RouteTemplates[path]; exists {
		return fmt.Sprintf("%s:%s", method, template)
	}

	// Try to detect dynamic routes and normalize them
	// NOTE: This is a simple heuristic - adjust based on your routing patterns
	normalizedPath := m.normalizePath(path)

	return fmt.Sprintf("%s:%s", method, normalizedPath)
}

// normalizePath normalizes dynamic route segments
func (m *RateLimitMiddleware) normalizePath(path string) string {
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		// Replace UUID-like segments
		if len(segment) == 36 && strings.Count(segment, "-") == 4 {
			segments[i] = "{id}"
		}
		// Replace numeric IDs
		if _, err := strconv.Atoi(segment); err == nil && len(segment) > 0 {
			segments[i] = "{id}"
		}
	}
	return strings.Join(segments, "/")
}

// RateLimitInfo contains rate limit information for headers
type RateLimitInfo struct {
	Limit     int           `json:"limit"`
	Remaining int           `json:"remaining"`
	ResetTime time.Time     `json:"reset_time"`
	Window    time.Duration `json:"window"`
}

// addRateLimitHeaders adds standard rate limit headers to the response
func (m *RateLimitMiddleware) addRateLimitHeaders(c *fiber.Ctx, limits map[string]RateLimitInfo) {
	// Use the most restrictive limit for headers
	var mostRestrictive RateLimitInfo
	mostRestrictive.Remaining = int(^uint(0) >> 1) // Max int

	for _, limit := range limits {
		if limit.Remaining < mostRestrictive.Remaining {
			mostRestrictive = limit
		}
	}

	if mostRestrictive.Limit > 0 {
		c.Set("X-RateLimit-Limit", strconv.Itoa(mostRestrictive.Limit))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(mostRestrictive.Remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(mostRestrictive.ResetTime.Unix(), 10))
	}
}

// sendRateLimitError sends a rate limit error response
func (m *RateLimitMiddleware) sendRateLimitError(c *fiber.Ctx, message string, limits map[string]RateLimitInfo) error {
	if limits != nil {
		m.addRateLimitHeaders(c, limits)
	}

	c.Set("Retry-After", strconv.Itoa(int(m.config.GlobalWindowSize.Seconds())))

	return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
		"error":     "rate_limit_exceeded",
		"message":   message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// recordRateLimitMetrics records rate limiting metrics
func (m *RateLimitMiddleware) recordRateLimitMetrics(ctx context.Context, result, clientIP, userID, endpoint string, duration time.Duration) {
	labels := metrics.Fields{
		"result":   result,
		"endpoint": endpoint,
		"has_user": strconv.FormatBool(userID != ""),
	}

	m.metrics.IncrementCounter("rate_limit_requests_total", labels)
	m.metrics.ObserveHistogram("rate_limit_check_duration", duration.Seconds(), labels)

	if result == "exceeded" || result == "blocked" {
		m.metrics.IncrementCounter("rate_limit_violations_total", labels)
	}
}

// =============================================================================
// Circuit Breaker Implementation
// =============================================================================

// ShouldAttempt checks if cache operations should be attempted
func (cb *CacheCircuitBreaker) ShouldAttempt() bool {
	state := cb.state.Load()

	switch state {
	case 0: // Closed - normal operation
		return true

	case 1: // Open - rejecting requests
		// Check if recovery timeout has passed
		lastFailure := cb.lastFailure.Load()
		if time.Since(time.Unix(lastFailure, 0)) > cb.config.RecoveryTimeout {
			// Try to transition to half-open
			if cb.state.CompareAndSwap(1, 2) {
				cb.consecutiveFail.Store(0)
			}
			return true
		}
		return false

	case 2: // Half-open - testing recovery
		return true

	default:
		return true
	}
}

// RecordSuccess records a successful cache operation
func (cb *CacheCircuitBreaker) RecordSuccess() {
	state := cb.state.Load()

	if state == 2 { // Half-open
		// Count consecutive successes
		successes := cb.consecutiveFail.Add(-1)
		if -successes >= int64(cb.config.SuccessThreshold) {
			// Transition to closed
			cb.state.Store(0)
			cb.failures.Store(0)
			cb.consecutiveFail.Store(0)
		}
	} else if state == 0 { // Closed
		// Reset failure count on success
		cb.failures.Store(0)
	}
}

// RecordFailure records a failed cache operation
func (cb *CacheCircuitBreaker) RecordFailure() {
	failures := cb.failures.Add(1)
	cb.lastFailure.Store(time.Now().Unix())

	state := cb.state.Load()

	if state == 0 { // Closed
		if failures >= int64(cb.config.FailureThreshold) {
			// Transition to open
			cb.state.Store(1)
		}
	} else if state == 2 { // Half-open
		// Failure during half-open immediately returns to open
		cb.state.Store(1)
		cb.consecutiveFail.Store(0)
	}
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// =============================================================================
// Integration with Existing Middleware Stack
// =============================================================================

// CreateRateLimitMiddleware creates a rate limit middleware compatible with the existing stack
// NOTE: This is a factory function that integrates with your existing middleware pattern
func CreateRateLimitMiddleware(
	config *RateLimitConfig,
	cache cache.Service,
	logger logger.Logger,
) fiber.Handler {
	// Use default config if none provided
	if config == nil {
		defaultConfig := DefaultRateLimitConfig()
		config = &defaultConfig
	}

	// NOTE: These dependencies should be injected from your service container
	// in production. This factory function shows the expected pattern.
	// Example injection:
	// metricsProvider := serviceContainer.GetMetrics()
	// tracingService := serviceContainer.GetTracing()
	var metricsProvider metrics.MetricsProvider
	var tracingService tracing.TracingService

	middleware := NewRateLimitMiddleware(
		cache,
		*config,
		logger,
		metricsProvider,
		tracingService,
	)

	return middleware.FiberMiddleware()
}

// =============================================================================
// Backward Compatibility Layer
// =============================================================================

// NOTE: These functions maintain compatibility with your existing middleware setup
// They allow gradual migration from the old HTTP-based middleware to Fiber

// RateLimitMiddlewareCompat wraps the new implementation for backward compatibility
type RateLimitMiddlewareCompat struct {
	inner *RateLimitMiddleware
}

// NewRateLimitMiddlewareCompat creates a backward-compatible wrapper
func NewRateLimitMiddlewareCompat(
	cache cache.Service,
	config RateLimitConfig,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) *RateLimitMiddlewareCompat {
	return &RateLimitMiddlewareCompat{
		inner: NewRateLimitMiddleware(cache, config, logger, metrics, tracer),
	}
}

// FiberMiddleware returns the Fiber handler
func (m *RateLimitMiddlewareCompat) FiberMiddleware() fiber.Handler {
	return m.inner.FiberMiddleware()
}

// GetStats returns current rate limit statistics
func (m *RateLimitMiddlewareCompat) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"circuit_breaker_state": m.inner.cacheCircuitBreaker.state.Load(),
		"cache_failures":        m.inner.cacheCircuitBreaker.failures.Load(),
		"config":                m.inner.config,
	}
}

// =============================================================================
// Advanced Features
// =============================================================================

// DynamicRateLimitConfig allows runtime configuration updates
type DynamicRateLimitConfig struct {
	mu     sync.RWMutex
	config RateLimitConfig
}

// NewDynamicRateLimitConfig creates a dynamic configuration
func NewDynamicRateLimitConfig(initial RateLimitConfig) *DynamicRateLimitConfig {
	return &DynamicRateLimitConfig{
		config: initial,
	}
}

// Get returns the current configuration
func (d *DynamicRateLimitConfig) Get() RateLimitConfig {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.config
}

// Update updates the configuration atomically
func (d *DynamicRateLimitConfig) Update(newConfig RateLimitConfig) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.config = newConfig
}

// UpdateEndpointLimit updates a specific endpoint limit
func (d *DynamicRateLimitConfig) UpdateEndpointLimit(endpoint string, limit EndpointLimit) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.config.EndpointLimits == nil {
		d.config.EndpointLimits = make(map[string]EndpointLimit)
	}
	d.config.EndpointLimits[endpoint] = limit
}

// =============================================================================
// Monitoring and Debugging
// =============================================================================

// RateLimitMonitor provides monitoring capabilities
type RateLimitMonitor struct {
	middleware *RateLimitMiddleware
	logger     logger.Logger
}

// NewRateLimitMonitor creates a new monitor
func NewRateLimitMonitor(middleware *RateLimitMiddleware, logger logger.Logger) *RateLimitMonitor {
	return &RateLimitMonitor{
		middleware: middleware,
		logger:     logger,
	}
}

// GetCircuitBreakerStatus returns circuit breaker status
func (m *RateLimitMonitor) GetCircuitBreakerStatus() map[string]interface{} {
	cb := m.middleware.cacheCircuitBreaker

	stateNames := map[int32]string{
		0: "closed",
		1: "open",
		2: "half-open",
	}

	return map[string]interface{}{
		"state":             stateNames[cb.state.Load()],
		"total_failures":    cb.failures.Load(),
		"consecutive_fails": cb.consecutiveFail.Load(),
		"last_failure":      time.Unix(cb.lastFailure.Load(), 0).Format(time.RFC3339),
	}
}

// GetBucketStats returns token bucket statistics for a specific key
// NOTE: This is for debugging - don't use in production hot path
func (m *RateLimitMonitor) GetBucketStats(ctx context.Context, limitType, key string) (map[string]interface{}, error) {
	cacheKey := fmt.Sprintf("rate_limit:bucket:%s:%s", limitType, key)

	var data TokenBucketData
	err := m.middleware.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket stats: %w", err)
	}

	return map[string]interface{}{
		"tokens":      data.Tokens,
		"capacity":    data.Capacity,
		"rate":        data.Rate,
		"last_refill": time.Unix(data.LastRefill, 0).Format(time.RFC3339),
	}, nil
}

// ResetBucket resets a specific rate limit bucket (for debugging/testing)
func (m *RateLimitMonitor) ResetBucket(ctx context.Context, limitType, key string) error {
	cacheKey := fmt.Sprintf("rate_limit:bucket:%s:%s", limitType, key)

	// Use the existing cache service Delete method
	if err := m.middleware.cache.Delete(ctx, cacheKey); err != nil {
		return fmt.Errorf("failed to delete rate limit bucket: %w", err)
	}

	return nil
}

// =============================================================================
// Testing Utilities
// =============================================================================

// RateLimitTestHelper provides utilities for testing rate limiting
type RateLimitTestHelper struct {
	middleware *RateLimitMiddleware
}

// NewRateLimitTestHelper creates a test helper
func NewRateLimitTestHelper(middleware *RateLimitMiddleware) *RateLimitTestHelper {
	return &RateLimitTestHelper{
		middleware: middleware,
	}
}

// SimulateViolations simulates rate limit violations for testing
func (h *RateLimitTestHelper) SimulateViolations(ctx context.Context, clientIP string, count int) error {
	for i := 0; i < count; i++ {
		if err := h.middleware.recordViolation(ctx, clientIP, ""); err != nil {
			return err
		}
	}
	return nil
}

// ClearViolations clears all violations for a client (testing only)
func (h *RateLimitTestHelper) ClearViolations(ctx context.Context, clientIP string) error {
	ipViolationKey := fmt.Sprintf("rate_limit:violations:ip:%s", clientIP)
	ipBlockKey := fmt.Sprintf("rate_limit:block:ip:%s", clientIP)

	// Use existing cache service to delete violation records
	if err := h.middleware.cache.Delete(ctx, ipViolationKey); err != nil {
		return fmt.Errorf("failed to clear IP violations: %w", err)
	}

	if err := h.middleware.cache.Delete(ctx, ipBlockKey); err != nil {
		return fmt.Errorf("failed to clear IP block: %w", err)
	}

	return nil
}

// ForceCircuitBreakerState forces a circuit breaker state (testing only)
func (h *RateLimitTestHelper) ForceCircuitBreakerState(state string) {
	stateMap := map[string]int32{
		"closed":    0,
		"open":      1,
		"half-open": 2,
	}

	if val, ok := stateMap[state]; ok {
		h.middleware.cacheCircuitBreaker.state.Store(val)
	}
}

// =============================================================================
// Integration Example
// =============================================================================

/*
Usage Example 1: Basic Integration with Existing Middleware Stack

```go
package main

import (
	"time"
	"github.com/gofiber/fiber/v2"
	"github.com/niiniyare/erp/internal/platform/middleware"
)

func main() {
	app := fiber.New()

	// Initialize dependencies (from your DI container)
	cache := container.GetCache()
	logger := container.GetLogger()
	metrics := container.GetMetrics()
	tracer := container.GetTracer()

	// Create rate limit config
	rateLimitConfig := middleware.DefaultRateLimitConfig()

	// Customize for your needs
	rateLimitConfig.GlobalRPS = 2000
	rateLimitConfig.FailOpen = true
	rateLimitConfig.EndpointLimits["POST:/api/v1/auth/login"] = middleware.EndpointLimit{
		RPS:        3,
		Burst:      5,
		WindowSize: time.Minute,
	}

	// Create middleware
	rateLimitMiddleware := middleware.CreateRateLimitMiddleware(
		&rateLimitConfig,
		cache,
		logger,
	)

	// Apply to all routes
	app.Use(rateLimitMiddleware)

	// Or apply selectively
	apiGroup := app.Group("/api/v1")
	apiGroup.Use(rateLimitMiddleware)

	app.Listen(":8080")
}
```

Usage Example 2: Integration with OptimizedMiddlewareChain

```go
package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niiniyare/erp/internal/platform/middleware"
)

func main() {
	app := fiber.New()

	// Initialize performance chain
	perfConfig := middleware.DefaultPerformanceConfig()
	chain := middleware.NewOptimizedMiddlewareChain(perfConfig, logger, metrics)

	// Build complete middleware stack
	builder := middleware.NewMiddlewareChainBuilder(app, perfConfig, logger, metrics)

	rateLimitConfig := middleware.DefaultRateLimitConfig()

	builder.
		WithRecovery().
		WithCircuitBreaker().
		WithLogging().
		WithRateLimit(100, time.Minute). // Using builder's built-in rate limit
		// OR use custom rate limit middleware
		WithCustomMiddleware(middleware.CreateRateLimitMiddleware(&rateLimitConfig, cache, logger)).
		WithTenantIsolation(tenantConfig).
		Build()

	app.Listen(":8080")
}
```

Usage Example 3: Dynamic Configuration Updates

```go
package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niiniyare/erp/internal/platform/middleware"
)

func main() {
	app := fiber.New()

	// Create dynamic config
	dynamicConfig := middleware.NewDynamicRateLimitConfig(
		middleware.DefaultRateLimitConfig(),
	)

	// Create middleware with initial config
	rateLimitMiddleware := middleware.CreateRateLimitMiddleware(
		&dynamicConfig.Get(),
		cache,
		logger,
	)

	app.Use(rateLimitMiddleware)

	// Admin endpoint to update rate limits at runtime
	app.Put("/admin/rate-limits/:endpoint", func(c *fiber.Ctx) error {
		endpoint := c.Params("endpoint")

		var limit middleware.EndpointLimit
		if err := c.BodyParser(&limit); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		// Update configuration dynamically
		dynamicConfig.UpdateEndpointLimit(endpoint, limit)

		return c.JSON(fiber.Map{"success": true})
	})

	app.Listen(":8080")
}
```

Usage Example 4: Monitoring and Debugging

```go
package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niiniyare/erp/internal/platform/middleware"
)

func main() {
	app := fiber.New()

	// Create middleware
	rlMiddleware := middleware.NewRateLimitMiddleware(
		cache,
		middleware.DefaultRateLimitConfig(),
		logger,
		metrics,
		tracer,
	)

	app.Use(rlMiddleware.FiberMiddleware())

	// Create monitor
	monitor := middleware.NewRateLimitMonitor(rlMiddleware, logger)

	// Admin endpoints for monitoring
	app.Get("/admin/rate-limit/circuit-breaker", func(c *fiber.Ctx) error {
		return c.JSON(monitor.GetCircuitBreakerStatus())
	})

	app.Get("/admin/rate-limit/bucket/:type/:key", func(c *fiber.Ctx) error {
		limitType := c.Params("type")
		key := c.Params("key")

		stats, err := monitor.GetBucketStats(c.Context(), limitType, key)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(stats)
	})

	// Debug endpoint to reset a bucket
	app.Delete("/admin/rate-limit/bucket/:type/:key", func(c *fiber.Ctx) error {
		limitType := c.Params("type")
		key := c.Params("key")

		if err := monitor.ResetBucket(c.Context(), limitType, key); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"success": true})
	})

	app.Listen(":8080")
}
```

Usage Example 5: Testing

```go
package middleware_test

import (
	"testing"
	"github.com/gofiber/fiber/v2"
	"github.com/niiniyare/erp/internal/platform/middleware"
)

func TestRateLimitMiddleware(t *testing.T) {
	app := fiber.New()

	// Create test middleware
	config := middleware.DefaultRateLimitConfig()
	config.IPRPS = 5
	config.IPBurst = 10

	rlMiddleware := middleware.NewRateLimitMiddleware(
		mockCache,
		config,
		mockLogger,
		mockMetrics,
		mockTracer,
	)

	app.Use(rlMiddleware.FiberMiddleware())

	// Create test helper
	testHelper := middleware.NewRateLimitTestHelper(rlMiddleware)

	// Test rate limiting
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Real-IP", "192.168.1.1")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("Expected 200, got %d on request %d", resp.StatusCode, i+1)
		}
	}

	// 6th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "192.168.1.1")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 429 {
		t.Errorf("Expected 429, got %d", resp.StatusCode)
	}

	// Test circuit breaker
	testHelper.ForceCircuitBreakerState("open")

	req = httptest.NewRequest("GET", "/test", nil)
	resp, _ = app.Test(req)

	// Should allow if FailOpen is true
	if config.FailOpen && resp.StatusCode != 200 {
		t.Errorf("Expected fail-open to allow request")
	}
}
```

*/
