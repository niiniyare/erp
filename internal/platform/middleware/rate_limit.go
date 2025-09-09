package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

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
}

// EndpointLimit defines rate limits for specific API endpoints
type EndpointLimit struct {
	RPS        int           `json:"rps"`
	Burst      int           `json:"burst"`
	WindowSize time.Duration `json:"window_size"`
}

// RateLimitMiddleware provides comprehensive rate limiting middleware
type RateLimitMiddleware struct {
	cache   cache.Service
	config  RateLimitConfig
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(
	cache cache.Service,
	config RateLimitConfig,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		cache:   cache,
		config:  config,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
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
			"auth:login": {
				RPS:        5,
				Burst:      10,
				WindowSize: time.Minute,
			},
			"auth:refresh": {
				RPS:        10,
				Burst:      20,
				WindowSize: time.Minute,
			},
			"finance:transaction": {
				RPS:        50,
				Burst:      100,
				WindowSize: time.Minute,
			},
		},

		BlockDuration: 15 * time.Minute,
		MaxViolations: 5,
	}
}

// HTTPMiddleware creates an HTTP middleware for rate limiting
func (m *RateLimitMiddleware) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx, span := m.tracer.StartSpan(ctx, "middleware.rate_limit")
			defer span.End()

			start := time.Now()

			// Extract rate limiting identifiers
			clientIP := getClientIP(r)
			userID := getUserIDFromContext(ctx)
			endpoint := getEndpointKey(r)

			span.SetAttributes(
				attribute.String("client.ip", clientIP),
				attribute.String("user.id", userID),
				attribute.String("endpoint", endpoint),
			)

			// Check if client is currently blocked
			if blocked, blockReason := m.isBlocked(ctx, clientIP, userID); blocked {
				m.logger.WarnContext(ctx, "Request blocked due to rate limit violation", logger.Fields{
					"client_ip":    clientIP,
					"user_id":      userID,
					"endpoint":     endpoint,
					"block_reason": blockReason,
				})

				span.SetStatus(codes.Error, "rate_limit_blocked")
				m.recordRateLimitMetrics(ctx, "blocked", clientIP, userID, endpoint, time.Since(start))

				m.writeRateLimitResponse(w, http.StatusTooManyRequests, "Rate limit exceeded - temporarily blocked", nil)
				return
			}

			// Check multiple rate limit dimensions
			allowed, limits, err := m.checkRateLimits(ctx, clientIP, userID, endpoint)
			if err != nil {
				m.logger.ErrorContext(ctx, "Rate limit check failed", logger.Fields{
					"error":     err.Error(),
					"client_ip": clientIP,
					"user_id":   userID,
					"endpoint":  endpoint,
				})
				span.RecordError(err)
				// Allow request to proceed on rate limit check errors
			}

			if !allowed {
				// Increment violation counter
				if err := m.recordViolation(ctx, clientIP, userID); err != nil {
					m.logger.ErrorContext(ctx, "Failed to record rate limit violation", logger.Fields{
						"error": err.Error(),
					})
				}

				m.logger.WarnContext(ctx, "Rate limit exceeded", logger.Fields{
					"client_ip": clientIP,
					"user_id":   userID,
					"endpoint":  endpoint,
					"limits":    limits,
				})

				span.SetStatus(codes.Error, "rate_limit_exceeded")
				m.recordRateLimitMetrics(ctx, "exceeded", clientIP, userID, endpoint, time.Since(start))

				m.writeRateLimitResponse(w, http.StatusTooManyRequests, "Rate limit exceeded", limits)
				return
			}

			// Record successful request
			m.recordRateLimitMetrics(ctx, "allowed", clientIP, userID, endpoint, time.Since(start))

			// Add rate limit headers to response
			m.addRateLimitHeaders(w, limits)

			// Continue to next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// checkRateLimits checks all applicable rate limits
func (m *RateLimitMiddleware) checkRateLimits(ctx context.Context, clientIP, userID, endpoint string) (bool, map[string]RateLimitInfo, error) {
	limits := make(map[string]RateLimitInfo)

	// Check global rate limit
	globalAllowed, globalInfo, err := m.checkLimit(ctx, "global", "", m.config.GlobalRPS, m.config.GlobalBurst, m.config.GlobalWindowSize)
	if err != nil {
		return false, limits, fmt.Errorf("global rate limit check failed: %w", err)
	}
	limits["global"] = globalInfo
	if !globalAllowed {
		return false, limits, nil
	}

	// Check IP-based rate limit
	ipAllowed, ipInfo, err := m.checkLimit(ctx, "ip", clientIP, m.config.IPRPS, m.config.IPBurst, m.config.IPWindowSize)
	if err != nil {
		return false, limits, fmt.Errorf("IP rate limit check failed: %w", err)
	}
	limits["ip"] = ipInfo
	if !ipAllowed {
		return false, limits, nil
	}

	// Check user-based rate limit (if user is authenticated)
	if userID != "" {
		userAllowed, userInfo, err := m.checkLimit(ctx, "user", userID, m.config.UserRPS, m.config.UserBurst, m.config.UserWindowSize)
		if err != nil {
			return false, limits, fmt.Errorf("user rate limit check failed: %w", err)
		}
		limits["user"] = userInfo
		if !userAllowed {
			return false, limits, nil
		}
	}

	// Check endpoint-specific rate limit
	if endpointLimit, exists := m.config.EndpointLimits[endpoint]; exists {
		endpointKey := fmt.Sprintf("endpoint:%s:%s", endpoint, clientIP)
		if userID != "" {
			endpointKey = fmt.Sprintf("endpoint:%s:user:%s", endpoint, userID)
		}

		endpointAllowed, endpointInfo, err := m.checkLimit(ctx, "endpoint", endpointKey, endpointLimit.RPS, endpointLimit.Burst, endpointLimit.WindowSize)
		if err != nil {
			return false, limits, fmt.Errorf("endpoint rate limit check failed: %w", err)
		}
		limits["endpoint"] = endpointInfo
		if !endpointAllowed {
			return false, limits, nil
		}
	}

	return true, limits, nil
}

// checkLimit checks a single rate limit using sliding window algorithm
func (m *RateLimitMiddleware) checkLimit(ctx context.Context, limitType, key string, rps, burst int, windowSize time.Duration) (bool, RateLimitInfo, error) {
	cacheKey := fmt.Sprintf("rate_limit:%s:%s", limitType, key)
	now := time.Now()
	windowStart := now.Add(-windowSize)

	// Get current window data from cache
	windowData, err := m.getWindowData(ctx, cacheKey)
	if err != nil {
		return false, RateLimitInfo{}, err
	}

	// Remove old entries outside the window
	windowData = m.cleanWindow(windowData, windowStart)

	// Check if we can allow this request
	currentCount := len(windowData.Requests)
	
	info := RateLimitInfo{
		Limit:     rps,
		Remaining: max(0, rps-currentCount-1),
		ResetTime: windowStart.Add(windowSize),
		Window:    windowSize,
	}

	if currentCount >= rps {
		return false, info, nil
	}

	// Add current request to window
	windowData.Requests = append(windowData.Requests, now.Unix())

	// Save updated window data
	if err := m.saveWindowData(ctx, cacheKey, windowData, windowSize); err != nil {
		return false, info, err
	}

	return true, info, nil
}

// WindowData represents the sliding window data for rate limiting
type WindowData struct {
	Requests []int64 `json:"requests"`
}

// getWindowData retrieves window data from cache
func (m *RateLimitMiddleware) getWindowData(ctx context.Context, key string) (WindowData, error) {
	var data WindowData
	
	err := m.cache.Get(ctx, key, &data)
	if err != nil {
		return WindowData{Requests: []int64{}}, nil
	}

	return data, nil
}

// saveWindowData saves window data to cache
func (m *RateLimitMiddleware) saveWindowData(ctx context.Context, key string, data WindowData, ttl time.Duration) error {
	return m.cache.Set(ctx, key, data, ttl)
}

// cleanWindow removes requests outside the time window
func (m *RateLimitMiddleware) cleanWindow(data WindowData, windowStart time.Time) WindowData {
	windowStartUnix := windowStart.Unix()
	var validRequests []int64

	for _, timestamp := range data.Requests {
		if timestamp >= windowStartUnix {
			validRequests = append(validRequests, timestamp)
		}
	}

	return WindowData{Requests: validRequests}
}

// isBlocked checks if a client is currently blocked
func (m *RateLimitMiddleware) isBlocked(ctx context.Context, clientIP, userID string) (bool, string) {
	// Check IP block
	ipBlockKey := fmt.Sprintf("rate_limit:block:ip:%s", clientIP)
	blocked, err := m.cache.Exists(ctx, ipBlockKey)
	if err == nil && blocked {
		return true, "ip_blocked"
	}

	// Check user block (if authenticated)
	if userID != "" {
		userBlockKey := fmt.Sprintf("rate_limit:block:user:%s", userID)
		blocked, err := m.cache.Exists(ctx, userBlockKey)
		if err == nil && blocked {
			return true, "user_blocked"
		}
	}

	return false, ""
}

// recordViolation records a rate limit violation and potentially blocks the client
func (m *RateLimitMiddleware) recordViolation(ctx context.Context, clientIP, userID string) error {
	// Record IP violation
	ipViolationKey := fmt.Sprintf("rate_limit:violations:ip:%s", clientIP)
	violations, err := m.incrementViolations(ctx, ipViolationKey)
	if err != nil {
		return err
	}

	// Block IP if too many violations
	if violations >= m.config.MaxViolations {
		ipBlockKey := fmt.Sprintf("rate_limit:block:ip:%s", clientIP)
		if err := m.cache.Set(ctx, ipBlockKey, true, m.config.BlockDuration); err != nil {
			return err
		}
	}

	// Record user violation (if authenticated)
	if userID != "" {
		userViolationKey := fmt.Sprintf("rate_limit:violations:user:%s", userID)
		userViolations, err := m.incrementViolations(ctx, userViolationKey)
		if err != nil {
			return err
		}

		// Block user if too many violations
		if userViolations >= m.config.MaxViolations {
			userBlockKey := fmt.Sprintf("rate_limit:block:user:%s", userID)
			if err := m.cache.Set(ctx, userBlockKey, true, m.config.BlockDuration); err != nil {
				return err
			}
		}
	}

	return nil
}

// incrementViolations increments and returns the current violation count
func (m *RateLimitMiddleware) incrementViolations(ctx context.Context, key string) (int, error) {
	var count int
	err := m.cache.Get(ctx, key, &count)
	if err != nil {
		// If key doesn't exist, start with 1
		if err := m.cache.Set(ctx, key, 1, time.Hour); err != nil {
			return 0, err
		}
		return 1, nil
	}

	count++
	if err := m.cache.Set(ctx, key, count, time.Hour); err != nil {
		return 0, err
	}

	return count, nil
}

// Helper functions

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

func getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

func getEndpointKey(r *http.Request) string {
	// Create endpoint key based on method and path
	return fmt.Sprintf("%s:%s", r.Method, r.URL.Path)
}

// RateLimitInfo contains rate limit information for headers
type RateLimitInfo struct {
	Limit     int           `json:"limit"`
	Remaining int           `json:"remaining"`
	ResetTime time.Time     `json:"reset_time"`
	Window    time.Duration `json:"window"`
}

// addRateLimitHeaders adds standard rate limit headers to the response
func (m *RateLimitMiddleware) addRateLimitHeaders(w http.ResponseWriter, limits map[string]RateLimitInfo) {
	// Use the most restrictive limit for headers
	var mostRestrictive RateLimitInfo
	mostRestrictive.Remaining = int(^uint(0) >> 1) // Max int

	for _, limit := range limits {
		if limit.Remaining < mostRestrictive.Remaining {
			mostRestrictive = limit
		}
	}

	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(mostRestrictive.Limit))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(mostRestrictive.Remaining))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(mostRestrictive.ResetTime.Unix(), 10))
}

// writeRateLimitResponse writes a rate limit exceeded response
func (m *RateLimitMiddleware) writeRateLimitResponse(w http.ResponseWriter, status int, message string, limits map[string]RateLimitInfo) {
	if limits != nil {
		m.addRateLimitHeaders(w, limits)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", strconv.Itoa(int(m.config.GlobalWindowSize.Seconds())))
	w.WriteHeader(status)

	response := map[string]interface{}{
		"error":   "rate_limit_exceeded",
		"message": message,
	}

	// Don't return error if JSON marshaling fails, just log it
	if jsonResp, err := json.Marshal(response); err == nil {
		w.Write(jsonResp)
	}
}

// recordRateLimitMetrics records rate limiting metrics
func (m *RateLimitMiddleware) recordRateLimitMetrics(ctx context.Context, result, clientIP, userID, endpoint string, duration time.Duration) {
	labels := metrics.Fields{
		"result":   result,
		"endpoint": endpoint,
	}

	if userID != "" {
		labels["has_user"] = "true"
	} else {
		labels["has_user"] = "false"
	}

	m.metrics.IncrementCounter("rate_limit_requests_total", labels)
	m.metrics.ObserveHistogram("rate_limit_check_duration", duration.Seconds(), labels)

	if result == "exceeded" || result == "blocked" {
		m.metrics.IncrementCounter("rate_limit_violations_total", labels)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}