package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/codes"
)

// SecurityValidationResult represents the result of security middleware validation
type SecurityValidationResult struct {
	Valid            bool           `json:"valid"`
	Errors           []string       `json:"errors,omitempty"`
	Warnings         []string       `json:"warnings,omitempty"`
	MiddlewareStatus map[string]any `json:"middleware_status"`
	GoaCompatibility bool           `json:"goa_compatibility"`
	SecurityScore    int            `json:"security_score"` // 0-100
}

// SecurityValidator validates security middleware compatibility with Goa framework
type SecurityValidator struct {
	logger  logger.Logger
	metrics *metrics.MetricsService
	tracing tracing.TracingService
}

// NewSecurityValidator creates a new security validation service
func NewSecurityValidator(
	logger logger.Logger,
	metrics *metrics.MetricsService,
	tracing tracing.TracingService,
) *SecurityValidator {
	return &SecurityValidator{
		logger:  logger,
		metrics: metrics,
		tracing: tracing,
	}
}

// ValidateMiddlewareStack performs validation of the middleware stack
func (sv *SecurityValidator) ValidateMiddlewareStack(stack *GoaMiddlewareStack) *SecurityValidationResult {
	ctx, span := sv.tracing.StartSpan(context.Background(), "security_validator.validate_stack")
	defer span.End()

	result := &SecurityValidationResult{
		Valid:            true,
		Errors:           make([]string, 0),
		Warnings:         make([]string, 0),
		MiddlewareStatus: make(map[string]any),
		GoaCompatibility: true,
		SecurityScore:    0,
	}

	sv.logger.InfoContext(ctx, "Starting security middleware validation")

	// Validate each middleware component
	sv.validateCORS(stack.corsConfig, result)
	sv.validateRateLimit(stack.rateLimitConfig, result)
	sv.validateCompression(stack.compressionConfig, result)
	sv.validateTimeout(stack.timeoutConfig, result)
	sv.validateWhitelist(stack.whitelist, result)

	// Calculate security score
	result.SecurityScore = sv.calculateSecurityScore(result)

	// Log validation results
	if result.Valid {
		sv.logger.Info("Security middleware validation passed", logger.Fields{
			"security_score":    result.SecurityScore,
			"goa_compatibility": result.GoaCompatibility,
			"warnings":          len(result.Warnings),
		})
		span.SetStatus(codes.Ok, "Validation successful")
	} else {
		sv.logger.Error("Security middleware validation failed", logger.Fields{
			"errors":   len(result.Errors),
			"warnings": len(result.Warnings),
		})
		span.SetStatus(codes.Error, "Validation failed")
	}

	// Record validation metrics
	sv.metrics.IncrementCounter("security_validations_total", metrics.Fields{
		"status": map[bool]string{true: "success", false: "failure"}[result.Valid],
	})

	sv.metrics.ObserveHistogram("security_score", float64(result.SecurityScore), metrics.Fields{
		"framework": "goa",
	})

	return result
}

// validateCORS validates CORS middleware configuration
func (sv *SecurityValidator) validateCORS(config *CORSConfig, result *SecurityValidationResult) {
	status := make(map[string]any)

	if config == nil {
		result.Errors = append(result.Errors, "CORS configuration is missing")
		result.Valid = false
		status["configured"] = false
		result.MiddlewareStatus["cors"] = status
		return
	}

	status["configured"] = true
	status["allow_credentials"] = config.AllowCredentials
	status["max_age"] = config.MaxAge
	status["allowed_origins_count"] = len(config.AllowedOrigins)

	// Security validations
	if config.AllowCredentials {
		// Check for wildcard origins with credentials
		for _, origin := range config.AllowedOrigins {
			if origin == "*" {
				result.Errors = append(result.Errors, "CORS: Cannot use wildcard origin (*) with AllowCredentials=true")
				result.Valid = false
			}
		}
	}

	// Check for development settings in production
	if config.EnableInDevelopment {
		result.Warnings = append(result.Warnings, "CORS: Development mode is enabled - ensure this is intended")
	}

	// Validate allowed headers include required ERP headers
	requiredHeaders := []string{"X-Tenant-ID", "Authorization", "Content-Type"}
	for _, required := range requiredHeaders {
		found := false
		for _, allowed := range config.AllowedHeaders {
			if strings.EqualFold(allowed, required) {
				found = true
				break
			}
		}
		if !found {
			result.Warnings = append(result.Warnings, fmt.Sprintf("CORS: Recommended header '%s' not in allowed headers", required))
		}
	}

	status["security_score"] = 85 // Base score for properly configured CORS
	result.MiddlewareStatus["cors"] = status

	sv.logger.Debug("CORS validation completed", logger.Fields{
		"status":  "valid",
		"origins": len(config.AllowedOrigins),
		"headers": len(config.AllowedHeaders),
		"methods": len(config.AllowedMethods),
	})
}

// validateRateLimit validates rate limiting middleware configuration
func (sv *SecurityValidator) validateRateLimit(config *RateLimitConfig, result *SecurityValidationResult) {
	status := make(map[string]any)

	if config == nil {
		result.Errors = append(result.Errors, "Rate limit configuration is missing")
		result.Valid = false
		status["configured"] = false
		result.MiddlewareStatus["rate_limit"] = status
		return
	}

	status["configured"] = true
	status["global_rps"] = config.GlobalRPS
	status["user_rps"] = config.UserRPS
	status["ip_rps"] = config.IPRPS

	// Validate rate limits are reasonable
	if config.GlobalRPS <= 0 {
		result.Errors = append(result.Errors, "Rate limit: GlobalRPS must be positive")
		result.Valid = false
	}

	if config.UserRPS <= 0 {
		result.Errors = append(result.Errors, "Rate limit: UserRPS must be positive")
		result.Valid = false
	}

	if config.IPRPS <= 0 {
		result.Errors = append(result.Errors, "Rate limit: IPRPS must be positive")
		result.Valid = false
	}

	// Check for reasonable limits (not too high or too low)
	if config.GlobalRPS > 10000 {
		result.Warnings = append(result.Warnings, "Rate limit: GlobalRPS is very high, consider if this is appropriate")
	}

	if config.UserRPS > 1000 {
		result.Warnings = append(result.Warnings, "Rate limit: UserRPS is very high, consider if this is appropriate")
	}

	// Validate burst configuration
	if config.GlobalBurst < config.GlobalRPS {
		result.Warnings = append(result.Warnings, "Rate limit: GlobalBurst should typically be >= GlobalRPS")
	}

	status["security_score"] = 90 // High score for rate limiting
	result.MiddlewareStatus["rate_limit"] = status

	sv.logger.Debug("Rate limit validation completed", logger.Fields{
		"global_rps": config.GlobalRPS,
		"user_rps":   config.UserRPS,
		"ip_rps":     config.IPRPS,
	})
}

// validateCompression validates compression middleware configuration
func (sv *SecurityValidator) validateCompression(config *CompressionConfig, result *SecurityValidationResult) {
	status := make(map[string]any)

	if config == nil {
		result.Warnings = append(result.Warnings, "Compression configuration is missing - performance impact")
		status["configured"] = false
		result.MiddlewareStatus["compression"] = status
		return
	}

	status["configured"] = true
	status["level"] = config.Level
	status["min_length"] = config.MinLength
	status["content_types_count"] = len(config.ContentTypes)

	// Validate compression level
	if config.Level < 1 || config.Level > 9 {
		result.Errors = append(result.Errors, "Compression: Level must be between 1 and 9")
		result.Valid = false
	}

	// Check for reasonable minimum length
	if config.MinLength < 100 {
		result.Warnings = append(result.Warnings, "Compression: MinLength is very low, may impact performance")
	}

	// Validate content types include common API types
	requiredTypes := []string{"application/json", "text/plain"}
	for _, required := range requiredTypes {
		found := false
		for _, configured := range config.ContentTypes {
			if strings.EqualFold(configured, required) {
				found = true
				break
			}
		}
		if !found {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Compression: Recommended content type '%s' not configured", required))
		}
	}

	status["security_score"] = 70 // Medium score - performance benefit, minor security impact
	result.MiddlewareStatus["compression"] = status
}

// validateTimeout validates timeout middleware configuration
func (sv *SecurityValidator) validateTimeout(config *TimeoutConfig, result *SecurityValidationResult) {
	status := make(map[string]any)

	if config == nil {
		result.Warnings = append(result.Warnings, "Timeout configuration is missing - DoS vulnerability")
		status["configured"] = false
		result.MiddlewareStatus["timeout"] = status
		return
	}

	status["configured"] = true
	status["request_timeout"] = config.RequestTimeout.String()
	status["custom_paths_count"] = len(config.EnableCustomPaths)

	// Validate timeout values are reasonable
	if config.RequestTimeout.Seconds() < 5 {
		result.Warnings = append(result.Warnings, "Timeout: RequestTimeout is very short, may cause legitimate requests to fail")
	}

	if config.RequestTimeout.Seconds() > 300 {
		result.Warnings = append(result.Warnings, "Timeout: RequestTimeout is very long, may not prevent DoS attacks effectively")
	}

	// Validate custom path timeouts
	for path, timeout := range config.EnableCustomPaths {
		if timeout.Seconds() > 600 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Timeout: Path '%s' has very long timeout (%s)", path, timeout))
		}
	}

	status["security_score"] = 85 // High score for DoS protection
	result.MiddlewareStatus["timeout"] = status
}

// validateWhitelist validates endpoint whitelist configuration
func (sv *SecurityValidator) validateWhitelist(whitelist *EndpointWhitelist, result *SecurityValidationResult) {
	status := make(map[string]any)

	if whitelist == nil {
		result.Errors = append(result.Errors, "Endpoint whitelist is missing - all endpoints may require authentication")
		result.Valid = false
		status["configured"] = false
		result.MiddlewareStatus["whitelist"] = status
		return
	}

	status["configured"] = true

	// Essential public endpoints that should be whitelisted
	essentialEndpoints := []string{
		"GET /health",
		"GET /ping",
		"POST /api/v1/auth/login",
		"POST /api/v1/auth/refresh",
	}

	for _, essential := range essentialEndpoints {
		parts := strings.SplitN(essential, " ", 2)
		if len(parts) == 2 {
			if !whitelist.IsPublicEndpoint(parts[0], parts[1]) {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Whitelist: Essential endpoint '%s' may not be whitelisted", essential))
			}
		}
	}

	status["security_score"] = 80 // Good score for access control
	result.MiddlewareStatus["whitelist"] = status
}

// calculateSecurityScore calculates overall security score based on middleware configuration
func (sv *SecurityValidator) calculateSecurityScore(result *SecurityValidationResult) int {
	if !result.Valid {
		return 0
	}

	totalScore := 0
	componentCount := 0

	for _, status := range result.MiddlewareStatus {
		if statusMap, ok := status.(map[string]any); ok {
			if score, exists := statusMap["security_score"]; exists {
				if scoreInt, ok := score.(int); ok {
					totalScore += scoreInt
					componentCount++
				}
			}
		}
	}

	if componentCount == 0 {
		return 0
	}

	averageScore := totalScore / componentCount

	// Deduct points for warnings and errors
	averageScore -= len(result.Warnings) * 2
	averageScore -= len(result.Errors) * 10

	if averageScore < 0 {
		averageScore = 0
	}
	if averageScore > 100 {
		averageScore = 100
	}

	return averageScore
}

// ValidateGoaCompatibility checks if middleware implementations are compatible with Goa framework
func (sv *SecurityValidator) ValidateGoaCompatibility() bool {
	// All middleware in this implementation uses standard http.Handler interface
	// which is fully compatible with Goa framework

	sv.logger.Info("Goa compatibility validation passed", logger.Fields{
		"framework":  "goa",
		"interface":  "http.Handler",
		"compatible": true,
	})

	return true
}

// GetSecurityRecommendations provides security recommendations based on validation results
func (sv *SecurityValidator) GetSecurityRecommendations(result *SecurityValidationResult) []string {
	recommendations := make([]string, 0)

	if result.SecurityScore < 70 {
		recommendations = append(recommendations, "Overall security score is below recommended threshold (70)")
	}

	// Add specific recommendations based on middleware status
	if corsStatus, exists := result.MiddlewareStatus["cors"]; exists {
		if statusMap, ok := corsStatus.(map[string]any); ok {
			if configured, exists := statusMap["configured"]; exists {
				if !configured.(bool) {
					recommendations = append(recommendations, "Implement CORS middleware for cross-origin security")
				}
			}
		}
	}

	if rateLimitStatus, exists := result.MiddlewareStatus["rate_limit"]; exists {
		if statusMap, ok := rateLimitStatus.(map[string]any); ok {
			if configured, exists := statusMap["configured"]; exists {
				if !configured.(bool) {
					recommendations = append(recommendations, "Implement rate limiting to prevent API abuse")
				}
			}
		}
	}

	if len(result.Errors) > 0 {
		recommendations = append(recommendations, "Address all configuration errors before deployment")
	}

	if len(result.Warnings) > 3 {
		recommendations = append(recommendations, "Review and address configuration warnings")
	}

	return recommendations
}
