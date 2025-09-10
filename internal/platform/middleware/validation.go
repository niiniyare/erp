package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// ValidationConfig defines the configuration for input validation
type ValidationConfig struct {
	// Request size limits
	MaxRequestSize    int64 `json:"max_request_size"`     // Maximum request body size in bytes
	MaxHeaderSize     int64 `json:"max_header_size"`      // Maximum header size
	MaxQueryParamSize int64 `json:"max_query_param_size"` // Maximum query parameter size

	// Content type validation
	AllowedContentTypes []string `json:"allowed_content_types"`
	StrictContentType   bool     `json:"strict_content_type"`

	// Input sanitization
	EnableHTMLSanitization  bool `json:"enable_html_sanitization"`
	EnableSQLInjectionCheck bool `json:"enable_sql_injection_check"`
	EnableXSSCheck          bool `json:"enable_xss_check"`
	EnablePathTraversal     bool `json:"enable_path_traversal_check"`

	// Field validation
	MaxFieldLength    int      `json:"max_field_length"`
	ForbiddenPatterns []string `json:"forbidden_patterns"`
	RequiredHeaders   []string `json:"required_headers"`

	// Specific endpoint validation
	EndpointRules map[string]EndpointValidationRule `json:"endpoint_rules"`

	// Security settings
	BlockSuspiciousRequests bool `json:"block_suspicious_requests"`
	LogSuspiciousRequests   bool `json:"log_suspicious_requests"`
}

// EndpointValidationRule defines validation rules for specific endpoints
type EndpointValidationRule struct {
	MaxRequestSize      int64    `json:"max_request_size"`
	AllowedMethods      []string `json:"allowed_methods"`
	RequiredHeaders     []string `json:"required_headers"`
	AllowedContentTypes []string `json:"allowed_content_types"`
	CustomPatterns      []string `json:"custom_patterns"`
}

// ValidationMiddleware provides comprehensive input validation and sanitization
type ValidationMiddleware struct {
	config  ValidationConfig
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService

	// Compiled regex patterns for performance
	sqlInjectionPatterns []*regexp.Regexp
	xssPatterns          []*regexp.Regexp
	pathTraversalPattern *regexp.Regexp
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware(
	config ValidationConfig,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) *ValidationMiddleware {
	m := &ValidationMiddleware{
		config:  config,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}

	// Compile regex patterns for performance
	m.compileSecurityPatterns()

	return m
}

// DefaultValidationConfig returns a secure default configuration
func DefaultValidationConfig() ValidationConfig {
	return ValidationConfig{
		// Request size limits (10MB max request, 8KB max header)
		MaxRequestSize:    10 * 1024 * 1024, // 10MB
		MaxHeaderSize:     8 * 1024,         // 8KB
		MaxQueryParamSize: 4 * 1024,         // 4KB

		// Allowed content types for ERP system
		AllowedContentTypes: []string{
			"application/json",
			"application/x-www-form-urlencoded",
			"multipart/form-data",
			"text/plain",
		},
		StrictContentType: true,

		// Enable all security checks
		EnableHTMLSanitization:  true,
		EnableSQLInjectionCheck: true,
		EnableXSSCheck:          true,
		EnablePathTraversal:     true,

		// Field validation
		MaxFieldLength: 10000, // 10KB per field

		// Common forbidden patterns
		ForbiddenPatterns: []string{
			`(?i)<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>`, // Script tags
			`(?i)javascript:`,   // JavaScript URLs
			`(?i)vbscript:`,     // VBScript URLs
			`(?i)data:.*base64`, // Base64 data URLs
		},

		// Security headers that should be present
		RequiredHeaders: []string{
			"Content-Type",
		},

		// Specific validation rules for sensitive endpoints
		EndpointRules: map[string]EndpointValidationRule{
			"POST:/api/v1/auth/login": {
				MaxRequestSize:      1024, // 1KB for login
				AllowedMethods:      []string{"POST"},
				AllowedContentTypes: []string{"application/json"},
				RequiredHeaders:     []string{"Content-Type", "X-Tenant-ID"},
			},
			"POST:/api/v1/finance/transactions": {
				MaxRequestSize:      100 * 1024, // 100KB for transaction
				AllowedMethods:      []string{"POST"},
				AllowedContentTypes: []string{"application/json"},
				RequiredHeaders:     []string{"Content-Type", "Authorization"},
			},
		},

		// Security behavior
		BlockSuspiciousRequests: true,
		LogSuspiciousRequests:   true,
	}
}

// HTTPMiddleware creates an HTTP middleware for input validation
func (m *ValidationMiddleware) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx, span := m.tracer.StartSpan(ctx, "middleware.validation")
			defer span.End()

			start := time.Now()

			// Get endpoint key for specific rules
			endpointKey := fmt.Sprintf("%s:%s", r.Method, r.URL.Path)

			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.Path),
				attribute.String("endpoint.key", endpointKey),
			)

			// Validate request size
			if err := m.validateRequestSize(r, endpointKey); err != nil {
				m.handleValidationError(ctx, w, r, "request_size", err, start)
				return
			}

			// Validate headers
			if err := m.validateHeaders(r, endpointKey); err != nil {
				m.handleValidationError(ctx, w, r, "headers", err, start)
				return
			}

			// Validate content type
			if err := m.validateContentType(r, endpointKey); err != nil {
				m.handleValidationError(ctx, w, r, "content_type", err, start)
				return
			}

			// Validate method
			if err := m.validateMethod(r, endpointKey); err != nil {
				m.handleValidationError(ctx, w, r, "method", err, start)
				return
			}

			// Validate query parameters
			if err := m.validateQueryParams(r); err != nil {
				m.handleValidationError(ctx, w, r, "query_params", err, start)
				return
			}

			// Validate and sanitize request body
			if r.Body != nil && r.ContentLength > 0 {
				sanitizedBody, err := m.validateAndSanitizeBody(ctx, r)
				if err != nil {
					m.handleValidationError(ctx, w, r, "body", err, start)
					return
				}

				// Replace request body with sanitized version
				if sanitizedBody != nil {
					r.Body = io.NopCloser(bytes.NewReader(sanitizedBody))
					r.ContentLength = int64(len(sanitizedBody))
				}
			}

			// Record successful validation
			m.recordValidationMetrics(ctx, endpointKey, "success", time.Since(start))

			// Continue to next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// validateRequestSize validates the request size limits
func (m *ValidationMiddleware) validateRequestSize(r *http.Request, endpointKey string) error {
	maxSize := m.config.MaxRequestSize

	// Check for endpoint-specific limit
	if rule, exists := m.config.EndpointRules[endpointKey]; exists && rule.MaxRequestSize > 0 {
		maxSize = rule.MaxRequestSize
	}

	if r.ContentLength > maxSize {
		return fmt.Errorf("request body too large: %d bytes (max: %d)", r.ContentLength, maxSize)
	}

	return nil
}

// validateHeaders validates request headers
func (m *ValidationMiddleware) validateHeaders(r *http.Request, endpointKey string) error {
	// Check total header size
	headerSize := int64(0)
	for name, values := range r.Header {
		for _, value := range values {
			headerSize += int64(len(name) + len(value) + 4) // +4 for ": " and "\r\n"
		}
	}

	if headerSize > m.config.MaxHeaderSize {
		return fmt.Errorf("headers too large: %d bytes (max: %d)", headerSize, m.config.MaxHeaderSize)
	}

	// Check required headers
	requiredHeaders := m.config.RequiredHeaders
	if rule, exists := m.config.EndpointRules[endpointKey]; exists {
		requiredHeaders = rule.RequiredHeaders
	}

	for _, header := range requiredHeaders {
		if r.Header.Get(header) == "" {
			return fmt.Errorf("missing required header: %s", header)
		}
	}

	// Check for suspicious header values
	for name, values := range r.Header {
		for _, value := range values {
			if m.containsSuspiciousContent(value) {
				return fmt.Errorf("suspicious content in header %s", name)
			}
		}
	}

	return nil
}

// validateContentType validates the request content type
func (m *ValidationMiddleware) validateContentType(r *http.Request, endpointKey string) error {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" && r.ContentLength > 0 {
		if m.config.StrictContentType {
			return fmt.Errorf("content-type header required for non-empty request body")
		}
		return nil
	}

	// Get allowed content types
	allowedTypes := m.config.AllowedContentTypes
	if rule, exists := m.config.EndpointRules[endpointKey]; exists && len(rule.AllowedContentTypes) > 0 {
		allowedTypes = rule.AllowedContentTypes
	}

	// Extract main content type (ignore parameters like charset)
	mainType := strings.Split(contentType, ";")[0]
	mainType = strings.TrimSpace(strings.ToLower(mainType))

	for _, allowed := range allowedTypes {
		if strings.ToLower(allowed) == mainType {
			return nil
		}
	}

	return fmt.Errorf("content type not allowed: %s", contentType)
}

// validateMethod validates the HTTP method
func (m *ValidationMiddleware) validateMethod(r *http.Request, endpointKey string) error {
	if rule, exists := m.config.EndpointRules[endpointKey]; exists && len(rule.AllowedMethods) > 0 {
		for _, method := range rule.AllowedMethods {
			if strings.ToUpper(method) == r.Method {
				return nil
			}
		}
		return fmt.Errorf("method not allowed: %s", r.Method)
	}

	return nil
}

// validateQueryParams validates query parameters
func (m *ValidationMiddleware) validateQueryParams(r *http.Request) error {
	query := r.URL.RawQuery
	if int64(len(query)) > m.config.MaxQueryParamSize {
		return fmt.Errorf("query parameters too large: %d bytes (max: %d)", len(query), m.config.MaxQueryParamSize)
	}

	// Check for suspicious content in query parameters
	for param, values := range r.URL.Query() {
		if m.containsSuspiciousContent(param) {
			return fmt.Errorf("suspicious parameter name: %s", param)
		}

		for _, value := range values {
			if m.containsSuspiciousContent(value) {
				return fmt.Errorf("suspicious parameter value for %s", param)
			}
		}
	}

	return nil
}

// validateAndSanitizeBody validates and sanitizes the request body
func (m *ValidationMiddleware) validateAndSanitizeBody(ctx context.Context, r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	// Check if body is valid JSON for JSON content types
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(strings.ToLower(contentType), "application/json") {
		var jsonData interface{}
		if err := json.Unmarshal(body, &jsonData); err != nil {
			return nil, fmt.Errorf("invalid JSON body: %w", err)
		}

		// Validate JSON structure recursively
		if err := m.validateJSONValue(jsonData, 0); err != nil {
			return nil, fmt.Errorf("invalid JSON content: %w", err)
		}
	}

	bodyStr := string(body)

	// Check for suspicious content
	if m.containsSuspiciousContent(bodyStr) {
		if m.config.BlockSuspiciousRequests {
			return nil, fmt.Errorf("suspicious content detected in request body")
		}

		if m.config.LogSuspiciousRequests {
			m.logger.WarnContext(ctx, "Suspicious content detected in request body", logger.Fields{
				"content_preview": bodyStr[:min(500, len(bodyStr))],
				"url":             r.URL.Path,
				"method":          r.Method,
			})
		}
	}

	// Apply sanitization if enabled
	if m.config.EnableHTMLSanitization {
		bodyStr = m.sanitizeHTML(bodyStr)
	}

	return []byte(bodyStr), nil
}

// validateJSONValue recursively validates JSON values
func (m *ValidationMiddleware) validateJSONValue(value interface{}, depth int) error {
	const maxDepth = 10
	if depth > maxDepth {
		return fmt.Errorf("JSON nesting too deep (max: %d)", maxDepth)
	}

	switch v := value.(type) {
	case string:
		if len(v) > m.config.MaxFieldLength {
			return fmt.Errorf("string field too long: %d chars (max: %d)", len(v), m.config.MaxFieldLength)
		}
		if m.containsSuspiciousContent(v) {
			return fmt.Errorf("suspicious content in string field")
		}

	case map[string]interface{}:
		const maxFields = 100
		if len(v) > maxFields {
			return fmt.Errorf("too many object fields: %d (max: %d)", len(v), maxFields)
		}

		for key, val := range v {
			if len(key) > m.config.MaxFieldLength {
				return fmt.Errorf("object key too long: %s", key)
			}
			if m.containsSuspiciousContent(key) {
				return fmt.Errorf("suspicious content in object key: %s", key)
			}
			if err := m.validateJSONValue(val, depth+1); err != nil {
				return err
			}
		}

	case []interface{}:
		const maxArrayLength = 1000
		if len(v) > maxArrayLength {
			return fmt.Errorf("array too long: %d items (max: %d)", len(v), maxArrayLength)
		}

		for _, item := range v {
			if err := m.validateJSONValue(item, depth+1); err != nil {
				return err
			}
		}
	}

	return nil
}

// containsSuspiciousContent checks if content contains suspicious patterns
func (m *ValidationMiddleware) containsSuspiciousContent(content string) bool {
	// Check SQL injection patterns
	if m.config.EnableSQLInjectionCheck {
		for _, pattern := range m.sqlInjectionPatterns {
			if pattern.MatchString(content) {
				return true
			}
		}
	}

	// Check XSS patterns
	if m.config.EnableXSSCheck {
		for _, pattern := range m.xssPatterns {
			if pattern.MatchString(content) {
				return true
			}
		}
	}

	// Check path traversal
	if m.config.EnablePathTraversal && m.pathTraversalPattern != nil {
		if m.pathTraversalPattern.MatchString(content) {
			return true
		}
	}

	// Check forbidden patterns
	for _, patternStr := range m.config.ForbiddenPatterns {
		if matched, _ := regexp.MatchString(patternStr, content); matched {
			return true
		}
	}

	return false
}

// sanitizeHTML removes dangerous HTML content
func (m *ValidationMiddleware) sanitizeHTML(content string) string {
	// Remove script tags
	scriptRegex := regexp.MustCompile(`(?i)<script\b[^>]*>.*?</script>`)
	content = scriptRegex.ReplaceAllString(content, "")

	// Remove dangerous attributes
	eventAttributes := regexp.MustCompile(`(?i)\s+on\w+\s*=\s*["'][^"']*["']`)
	content = eventAttributes.ReplaceAllString(content, "")

	// Remove javascript: URLs
	jsUrls := regexp.MustCompile(`(?i)javascript:\s*[^"'\s]+`)
	content = jsUrls.ReplaceAllString(content, "")

	return content
}

// compileSecurityPatterns compiles regex patterns for performance
func (m *ValidationMiddleware) compileSecurityPatterns() {
	// SQL injection patterns
	sqlPatterns := []string{
		`(?i)(union.*select)`,
		`(?i)(select.*from)`,
		`(?i)(insert.*into)`,
		`(?i)(delete.*from)`,
		`(?i)(update.*set)`,
		`(?i)(drop.*table)`,
		`(?i)(create.*table)`,
		`(?i)(alter.*table)`,
		`(?i)(exec.*\()`,
		`(?i)(execute.*\()`,
		`(?i)(\bor\b.*=.*=)`,
		`(?i)(\band\b.*=.*=)`,
		`(?i)(--.*$)`,
		`(?i)(/\*.*\*/)`,
		`(?i)(;.*--)`,
	}

	for _, pattern := range sqlPatterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			m.sqlInjectionPatterns = append(m.sqlInjectionPatterns, compiled)
		}
	}

	// XSS patterns
	xssPatterns := []string{
		`(?i)<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>`,
		`(?i)<iframe\b[^>]*>.*?<\/iframe>`,
		`(?i)<object\b[^>]*>.*?<\/object>`,
		`(?i)<embed\b[^>]*>`,
		`(?i)<applet\b[^>]*>.*?<\/applet>`,
		`(?i)javascript:`,
		`(?i)vbscript:`,
		`(?i)onload\s*=`,
		`(?i)onerror\s*=`,
		`(?i)onclick\s*=`,
		`(?i)onmouseover\s*=`,
	}

	for _, pattern := range xssPatterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			m.xssPatterns = append(m.xssPatterns, compiled)
		}
	}

	// Path traversal pattern
	m.pathTraversalPattern, _ = regexp.Compile(`(?i)(\.\./)|(\.\.\\)`)
}

// handleValidationError handles validation errors consistently
func (m *ValidationMiddleware) handleValidationError(ctx context.Context, w http.ResponseWriter, r *http.Request, errorType string, err error, startTime time.Time) {
	m.logger.WarnContext(ctx, "Request validation failed", logger.Fields{
		"error_type": errorType,
		"error":      err.Error(),
		"method":     r.Method,
		"url":        r.URL.Path,
		"client_ip":  getClientIP(r),
	})

	m.recordValidationMetrics(ctx, fmt.Sprintf("%s:%s", r.Method, r.URL.Path), "failed_"+errorType, time.Since(startTime))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := map[string]interface{}{
		"error":   "validation_failed",
		"message": err.Error(),
		"type":    errorType,
	}

	if jsonResp, jsonErr := json.Marshal(response); jsonErr == nil {
		w.Write(jsonResp)
	}
}

// recordValidationMetrics records validation metrics
func (m *ValidationMiddleware) recordValidationMetrics(ctx context.Context, endpoint, result string, duration time.Duration) {
	labels := metrics.Fields{
		"endpoint": endpoint,
		"result":   result,
	}

	m.metrics.IncrementCounter("validation_requests_total", labels)
	m.metrics.ObserveHistogram("validation_duration", duration.Seconds(), labels)

	if strings.HasPrefix(result, "failed_") {
		m.metrics.IncrementCounter("validation_failures_total", labels)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

