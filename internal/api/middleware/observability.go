package middleware

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// ObservabilityConfig configures the observability middleware
type ObservabilityConfig struct {
	// ServiceName for tracing and metrics
	ServiceName string

	// SkipPaths are paths to exclude from observability (e.g., health checks)
	SkipPaths []string

	// DetailedLogging enables verbose request/response logging
	DetailedLogging bool

	// SensitiveHeaders to exclude from logs and traces
	SensitiveHeaders []string

	// MaxRequestBodySize for logging (0 = disabled)
	MaxRequestBodySize int64

	// Custom metric labels to add to all requests
	CustomLabels map[string]string
}

// ObservabilityMiddleware provides comprehensive observability for all API requests
type ObservabilityMiddleware struct {
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.Service
	config  ObservabilityConfig
}

// NewObservabilityMiddleware creates a new observability middleware
func NewObservabilityMiddleware(
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
	config ObservabilityConfig,
) *ObservabilityMiddleware {
	// Set defaults
	if config.ServiceName == "" {
		config.ServiceName = "erp-api"
	}
	if config.SensitiveHeaders == nil {
		config.SensitiveHeaders = []string{
			"authorization", "cookie", "x-api-key", "x-auth-token",
			"x-access-token", "x-refresh-token", "x-csrf-token",
		}
	}
	if config.SkipPaths == nil {
		config.SkipPaths = []string{"/health", "/metrics", "/favicon.ico"}
	}

	return &ObservabilityMiddleware{
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
		config:  config,
	}
}

// FiberMiddleware returns the Fiber middleware handler
func (m *ObservabilityMiddleware) FiberMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()

		// Skip observability for certain paths
		if m.shouldSkipPath(path) {
			return c.Next()
		}

		// Create tracing span
		ctx := c.UserContext()
		operationName := m.getOperationName(method, path)
		ctx, span := m.tracer.StartSpan(ctx, operationName)
		defer span.End()

		// Store context back in Fiber
		c.SetUserContext(ctx)

		// Add request attributes to span
		m.addRequestAttributes(span, c)

		// Get request ID for correlation
		requestID := m.getOrCreateRequestID(c)

		// Log request start
		m.logRequestStart(ctx, c, requestID)

		// Execute request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)
		statusCode := c.Response().StatusCode()

		// Handle error if present
		if err != nil {
			m.handleError(span, c, err, statusCode)
		}

		// Add response attributes to span
		m.addResponseAttributes(span, c, duration)

		// Set span status based on response
		m.setSpanStatus(span, statusCode, err)

		// Record metrics
		m.recordMetrics(method, path, statusCode, duration, requestID)

		// Log request completion
		m.logRequestEnd(ctx, c, requestID, statusCode, duration, err)

		return err
	}
}

// shouldSkipPath checks if a path should be excluded from observability
func (m *ObservabilityMiddleware) shouldSkipPath(path string) bool {
	for _, skipPath := range m.config.SkipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// getOperationName creates a standardized operation name for tracing
func (m *ObservabilityMiddleware) getOperationName(method, path string) string {
	// Normalize dynamic route segments for better grouping
	normalizedPath := m.normalizePath(path)
	return m.config.ServiceName + "." + method + " " + normalizedPath
}

// normalizePath replaces dynamic segments with placeholders
func (m *ObservabilityMiddleware) normalizePath(path string) string {
	// TODO: Implement route template detection
	// For now, return the path as-is
	// In production, you'd want to replace UUID patterns, etc.
	return path
}

// addRequestAttributes adds HTTP request attributes to the span
func (m *ObservabilityMiddleware) addRequestAttributes(span tracing.Span, c *fiber.Ctx) {
	span.SetAttributes(
		attribute.String("http.method", c.Method()),
		attribute.String("http.url", string(c.Request().URI().FullURI())),
		attribute.String("http.scheme", c.Protocol()),
		attribute.String("http.host", c.Hostname()),
		attribute.String("http.route", c.Route().Path),
		attribute.String("http.user_agent", c.Get("User-Agent")),
		attribute.String("http.remote_addr", c.IP()),
		attribute.Int("http.request_content_length", len(c.Body())),
	)

	// Add tenant and user context if available
	if tenantID := c.Locals("tenant_id"); tenantID != nil {
		span.SetAttributes(attribute.String("tenant.id", tenantID.(string)))
	}
	if userID := c.Locals("user_id"); userID != nil {
		span.SetAttributes(attribute.String("user.id", userID.(string)))
	}

	// Add custom headers (excluding sensitive ones)
	m.addSafeHeaders(span, c)
}

// addResponseAttributes adds HTTP response attributes to the span
func (m *ObservabilityMiddleware) addResponseAttributes(span tracing.Span, c *fiber.Ctx, duration time.Duration) {
	span.SetAttributes(
		attribute.Int("http.status_code", c.Response().StatusCode()),
		attribute.Int("http.response_content_length", len(c.Response().Body())),
		attribute.Float64("http.duration", duration.Seconds()),
	)
}

// addSafeHeaders adds non-sensitive headers to the span
func (m *ObservabilityMiddleware) addSafeHeaders(span tracing.Span, c *fiber.Ctx) {
	c.Request().Header.VisitAll(func(key, value []byte) {
		headerName := string(key)

		// Skip sensitive headers
		for _, sensitive := range m.config.SensitiveHeaders {
			if headerName == sensitive {
				return
			}
		}

		// Add safe headers with prefix
		span.SetAttributes(attribute.String("http.request.header."+headerName, string(value)))
	})
}

// setSpanStatus sets the appropriate span status based on response
func (m *ObservabilityMiddleware) setSpanStatus(span tracing.Span, statusCode int, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return
	}

	if statusCode >= 400 {
		if statusCode >= 500 {
			span.SetStatus(codes.Error, "Server error")
		} else {
			span.SetStatus(codes.Error, "Client error")
		}
	} else {
		span.SetStatus(codes.Ok, "Success")
	}
}

// handleError records error information in the span
func (m *ObservabilityMiddleware) handleError(span tracing.Span, c *fiber.Ctx, err error, statusCode int) {
	span.RecordError(err)
	span.SetAttributes(
		attribute.String("error.type", "request_error"),
		attribute.String("error.message", err.Error()),
		attribute.Int("error.status_code", statusCode),
	)
}

// getOrCreateRequestID gets or creates a request ID for correlation
func (m *ObservabilityMiddleware) getOrCreateRequestID(c *fiber.Ctx) string {
	// Try to get from header first
	if requestID := c.Get("X-Request-ID"); requestID != "" {
		c.Locals("request_id", requestID)
		return requestID
	}

	// Try to get from locals (set by request ID middleware)
	if requestID := c.Locals("requestid"); requestID != nil {
		return requestID.(string)
	}

	// Generate new request ID
	requestID := generateRequestID()
	c.Locals("request_id", requestID)
	c.Set("X-Request-ID", requestID)
	return requestID
}

// recordMetrics records request metrics
func (m *ObservabilityMiddleware) recordMetrics(method, path string, statusCode int, duration time.Duration, requestID string) {
	// Create base labels
	labels := metrics.Fields{
		"method": method,
		"route":  m.normalizePath(path),
		"status": strconv.Itoa(statusCode),
	}

	// Add custom labels
	for k, v := range m.config.CustomLabels {
		labels[k] = v
	}

	// Record request count
	m.metrics.IncrementCounter("http_requests_total", labels)

	// Record request duration
	m.metrics.ObserveHistogram("http_request_duration_seconds", duration.Seconds(), labels)

	// Record request size metrics if available
	// Note: These would be more accurate with actual request/response sizes
	m.metrics.ObserveHistogram("http_request_size_bytes", 0, labels)
	m.metrics.ObserveHistogram("http_response_size_bytes", 0, labels)

	// Record status-specific metrics
	if statusCode >= 400 {
		errorLabels := make(metrics.Fields)
		for k, v := range labels {
			errorLabels[k] = v
		}
		errorLabels["error_type"] = m.getErrorType(statusCode)
		m.metrics.IncrementCounter("http_errors_total", errorLabels)
	}
}

// getErrorType categorizes HTTP status codes
func (m *ObservabilityMiddleware) getErrorType(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "server_error"
	case statusCode >= 400:
		return "client_error"
	default:
		return "unknown"
	}
}

// logRequestStart logs the start of a request
func (m *ObservabilityMiddleware) logRequestStart(ctx context.Context, c *fiber.Ctx, requestID string) {
	if !m.config.DetailedLogging {
		return
	}

	fields := logger.Fields{
		"request_id":     requestID,
		"method":         c.Method(),
		"path":           c.Path(),
		"remote_addr":    c.IP(),
		"user_agent":     c.Get("User-Agent"),
		"content_type":   c.Get("Content-Type"),
		"content_length": len(c.Body()),
	}

	// Add tenant/user context if available
	if tenantID := c.Locals("tenant_id"); tenantID != nil {
		fields["tenant_id"] = tenantID
	}
	if userID := c.Locals("user_id"); userID != nil {
		fields["user_id"] = userID
	}

	m.logger.InfoContext(ctx, "Request started", fields)
}

// logRequestEnd logs the completion of a request
func (m *ObservabilityMiddleware) logRequestEnd(ctx context.Context, c *fiber.Ctx, requestID string, statusCode int, duration time.Duration, err error) {
	fields := logger.Fields{
		"request_id":      requestID,
		"method":          c.Method(),
		"path":            c.Path(),
		"status_code":     statusCode,
		"duration_ms":     duration.Milliseconds(),
		"response_length": len(c.Response().Body()),
	}

	// Add tenant/user context if available
	if tenantID := c.Locals("tenant_id"); tenantID != nil {
		fields["tenant_id"] = tenantID
	}
	if userID := c.Locals("user_id"); userID != nil {
		fields["user_id"] = userID
	}

	var message string

	if err != nil {
		fields["error"] = err.Error()
		message = "Request failed"
		m.logger.ErrorContext(ctx, message, fields)
	} else if statusCode >= 500 {
		message = "Request completed with server error"
		m.logger.ErrorContext(ctx, message, fields)
	} else if statusCode >= 400 {
		message = "Request completed with client error"
		m.logger.WarnContext(ctx, message, fields)
	} else {
		message = "Request completed successfully"
		if m.config.DetailedLogging {
			m.logger.InfoContext(ctx, message, fields)
		} else {
			m.logger.DebugContext(ctx, message, fields)
		}
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	// Simple implementation - in production you might use a more sophisticated approach
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// DefaultObservabilityConfig returns sensible defaults
func DefaultObservabilityConfig() ObservabilityConfig {
	return ObservabilityConfig{
		ServiceName:        "erp-api",
		DetailedLogging:    false,
		MaxRequestBodySize: 1024 * 1024, // 1MB
		CustomLabels: map[string]string{
			"service": "erp",
		},
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
			"/ping",
		},
		SensitiveHeaders: []string{
			"authorization",
			"cookie",
			"x-api-key",
			"x-auth-token",
			"x-access-token",
			"x-refresh-token",
			"x-csrf-token",
			"set-cookie",
		},
	}
}

// CreateObservabilityMiddleware creates observability middleware with defaults
func CreateObservabilityMiddleware(
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
	config *ObservabilityConfig,
) fiber.Handler {
	if config == nil {
		defaultConfig := DefaultObservabilityConfig()
		config = &defaultConfig
	}

	middleware := NewObservabilityMiddleware(logger, metrics, tracer, *config)
	return middleware.FiberMiddleware()
}

// ============================================================================
// INTEGRATION WITH EXISTING MIDDLEWARE CHAIN
// ============================================================================

// ObservabilityMiddlewareBuilder helps integrate with existing middleware patterns
type ObservabilityMiddlewareBuilder struct {
	config  ObservabilityConfig
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.Service
}

// NewObservabilityMiddlewareBuilder creates a new builder
func NewObservabilityMiddlewareBuilder() *ObservabilityMiddlewareBuilder {
	return &ObservabilityMiddlewareBuilder{
		config: DefaultObservabilityConfig(),
	}
}

// WithLogger sets the logger
func (b *ObservabilityMiddlewareBuilder) WithLogger(logger logger.Logger) *ObservabilityMiddlewareBuilder {
	b.logger = logger
	return b
}

// WithMetrics sets the metrics provider
func (b *ObservabilityMiddlewareBuilder) WithMetrics(metrics metrics.MetricsProvider) *ObservabilityMiddlewareBuilder {
	b.metrics = metrics
	return b
}

// WithTracer sets the tracer
func (b *ObservabilityMiddlewareBuilder) WithTracer(tracer tracing.Service) *ObservabilityMiddlewareBuilder {
	b.tracer = tracer
	return b
}

// WithConfig sets the configuration
func (b *ObservabilityMiddlewareBuilder) WithConfig(config ObservabilityConfig) *ObservabilityMiddlewareBuilder {
	b.config = config
	return b
}

// WithServiceName sets the service name
func (b *ObservabilityMiddlewareBuilder) WithServiceName(name string) *ObservabilityMiddlewareBuilder {
	b.config.ServiceName = name
	return b
}

// WithDetailedLogging enables/disables detailed logging
func (b *ObservabilityMiddlewareBuilder) WithDetailedLogging(enabled bool) *ObservabilityMiddlewareBuilder {
	b.config.DetailedLogging = enabled
	return b
}

// WithSkipPaths sets paths to skip
func (b *ObservabilityMiddlewareBuilder) WithSkipPaths(paths []string) *ObservabilityMiddlewareBuilder {
	b.config.SkipPaths = paths
	return b
}

// WithCustomLabels sets custom metric labels
func (b *ObservabilityMiddlewareBuilder) WithCustomLabels(labels map[string]string) *ObservabilityMiddlewareBuilder {
	b.config.CustomLabels = labels
	return b
}

// Build creates the middleware
func (b *ObservabilityMiddlewareBuilder) Build() fiber.Handler {
	return CreateObservabilityMiddleware(b.logger, b.metrics, b.tracer, &b.config)
}
