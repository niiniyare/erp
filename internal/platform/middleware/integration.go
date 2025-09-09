package middleware

import (
	"fmt"
	"net/http"

	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/cache"
	loggerPkg "github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// MiddlewareStack provides a complete middleware stack for the ERP system
type MiddlewareStack struct {
	// Individual middleware components
	JWTAuth         *JWTAuthMiddleware
	Authorization   *AuthorizationMiddleware
	RateLimit       *RateLimitMiddleware
	SecurityHeaders *SecurityHeadersMiddleware
	Validation      *ValidationMiddleware
	Tenant          tenant.Service // Using the existing tenant middleware

	// Configuration
	Config MiddlewareConfig

	// Infrastructure
	logger  loggerPkg.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// MiddlewareConfig consolidates all middleware configurations
type MiddlewareConfig struct {
	// Environment settings
	Environment string `json:"environment"` // development, staging, production
	
	// Feature toggles
	EnableRateLimit       bool `json:"enable_rate_limit"`
	EnableSecurityHeaders bool `json:"enable_security_headers"`
	EnableValidation      bool `json:"enable_validation"`
	EnableAuthorization   bool `json:"enable_authorization"`
	EnableTenantIsolation bool `json:"enable_tenant_isolation"`

	// Individual middleware configs
	RateLimit       RateLimitConfig       `json:"rate_limit"`
	SecurityHeaders SecurityHeadersConfig `json:"security_headers"`
	Validation      ValidationConfig      `json:"validation"`
	Authorization   AuthorizationConfig   `json:"authorization"`
}

// NewMiddlewareStack creates a fully configured middleware stack
func NewMiddlewareStack(
	iamService iam.Service,
	tenantService tenant.Service,
	cacheService cache.Service,
	config MiddlewareConfig,
	logger loggerPkg.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) (*MiddlewareStack, error) {
	stack := &MiddlewareStack{
		Config:  config,
		Tenant:  tenantService,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}

	// Initialize JWT authentication middleware
	stack.JWTAuth = NewJWTAuthMiddleware(iamService, logger, metrics, tracer)

	// Initialize authorization middleware
	if config.EnableAuthorization {
		stack.Authorization = NewAuthorizationMiddleware(
			iamService,
			config.Authorization,
			logger,
			metrics,
			tracer,
		)
	}

	// Initialize rate limiting middleware
	if config.EnableRateLimit {
		stack.RateLimit = NewRateLimitMiddleware(
			cacheService,
			config.RateLimit,
			logger,
			metrics,
			tracer,
		)
	}

	// Initialize security headers middleware
	if config.EnableSecurityHeaders {
		stack.SecurityHeaders = NewSecurityHeadersMiddleware(
			config.SecurityHeaders,
			logger,
			metrics,
		)
	}

	// Initialize validation middleware
	if config.EnableValidation {
		stack.Validation = NewValidationMiddleware(
			config.Validation,
			logger,
			metrics,
			tracer,
		)
	}

	logger.Info("Middleware stack initialized", loggerPkg.Fields{
		"environment":         config.Environment,
		"rate_limit_enabled":  config.EnableRateLimit,
		"security_headers":    config.EnableSecurityHeaders,
		"validation_enabled":  config.EnableValidation,
		"authz_enabled":       config.EnableAuthorization,
		"tenant_isolation":    config.EnableTenantIsolation,
	})

	return stack, nil
}

// createTenantHTTPMiddleware creates an HTTP middleware for tenant isolation
func (m *MiddlewareStack) createTenantHTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract tenant info from request (subdomain, header, etc.)
			// This is a simplified version - you might want to implement
			// the full tenant extraction logic from the gin version
			
			// For now, pass through without tenant context
			// TODO: Implement proper tenant extraction for HTTP middleware
			next.ServeHTTP(w, r)
		})
	}
}

// DefaultMiddlewareConfig returns environment-appropriate default configuration
func DefaultMiddlewareConfig(environment string) MiddlewareConfig {
	config := MiddlewareConfig{
		Environment: environment,

		// Enable all features by default
		EnableRateLimit:       true,
		EnableSecurityHeaders: true,
		EnableValidation:      true,
		EnableAuthorization:   true,
		EnableTenantIsolation: true,
	}

	// Environment-specific configurations
	switch environment {
	case "development":
		config.RateLimit = DefaultRateLimitConfig()
		config.SecurityHeaders = DevelopmentSecurityHeadersConfig()
		config.Validation = DefaultValidationConfig()
		config.Authorization = DefaultAuthorizationConfig()

		// More permissive settings for development
		config.RateLimit.GlobalRPS = 10000
		config.Validation.BlockSuspiciousRequests = false

	case "staging":
		config.RateLimit = DefaultRateLimitConfig()
		config.SecurityHeaders = DefaultSecurityHeadersConfig()
		config.Validation = DefaultValidationConfig()
		config.Authorization = DefaultAuthorizationConfig()

		// Balanced settings for staging
		config.RateLimit.GlobalRPS = 5000

	case "production":
		config.RateLimit = DefaultRateLimitConfig()
		config.SecurityHeaders = DefaultSecurityHeadersConfig()
		config.Validation = DefaultValidationConfig()
		config.Authorization = DefaultAuthorizationConfig()

		// Strict settings for production
		config.RateLimit.GlobalRPS = 1000
		config.SecurityHeaders.HSTSMaxAge = 31536000 // 1 year
		config.Validation.BlockSuspiciousRequests = true

	default:
		// Default to production settings for unknown environments
		return DefaultMiddlewareConfig("production")
	}

	return config
}

// HTTPMiddlewareChain returns an ordered chain of HTTP middleware
func (m *MiddlewareStack) HTTPMiddlewareChain() []func(http.Handler) http.Handler {
	var chain []func(http.Handler) http.Handler

	// 1. Security headers (first - affects all responses)
	if m.SecurityHeaders != nil {
		chain = append(chain, m.SecurityHeaders.HTTPMiddleware())
	}

	// 2. Rate limiting (early - protects against abuse)
	if m.RateLimit != nil {
		chain = append(chain, m.RateLimit.HTTPMiddleware())
	}

	// 3. Request validation (before processing)
	if m.Validation != nil {
		chain = append(chain, m.Validation.HTTPMiddleware())
	}

	// 4. JWT token extraction (prepares for authentication)
	chain = append(chain, m.JWTAuth.HTTPJWTMiddleware())

	// 5. Tenant isolation (after auth context is available)
	if m.Config.EnableTenantIsolation {
		chain = append(chain, m.createTenantHTTPMiddleware())
	}

	// 6. Authorization (last - after all context is established)
	if m.Authorization != nil {
		chain = append(chain, m.Authorization.HTTPMiddleware())
	}

	return chain
}

// ApplyToHTTPHandler applies all middleware to an HTTP handler
func (m *MiddlewareStack) ApplyToHTTPHandler(handler http.Handler) http.Handler {
	middlewares := m.HTTPMiddlewareChain()

	// Apply middleware in reverse order (last middleware wraps first)
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}

// GetSecuritySummary returns a summary of applied security measures
func (m *MiddlewareStack) GetSecuritySummary() SecuritySummary {
	summary := SecuritySummary{
		Environment: m.Config.Environment,
		Features:    make(map[string]bool),
		Metrics:     make(map[string]interface{}),
	}

	// Feature flags
	summary.Features["rate_limiting"] = m.Config.EnableRateLimit
	summary.Features["security_headers"] = m.Config.EnableSecurityHeaders
	summary.Features["input_validation"] = m.Config.EnableValidation
	summary.Features["authorization"] = m.Config.EnableAuthorization
	summary.Features["tenant_isolation"] = m.Config.EnableTenantIsolation

	// Configuration metrics
	if m.Config.EnableRateLimit {
		summary.Metrics["rate_limit_global_rps"] = m.Config.RateLimit.GlobalRPS
		summary.Metrics["rate_limit_user_rps"] = m.Config.RateLimit.UserRPS
	}

	if m.Config.EnableSecurityHeaders {
		summary.Metrics["hsts_max_age"] = m.Config.SecurityHeaders.HSTSMaxAge
		summary.Metrics["csp_enabled"] = m.Config.SecurityHeaders.CSPPolicy != ""
	}

	if m.Config.EnableValidation {
		summary.Metrics["max_request_size"] = m.Config.Validation.MaxRequestSize
		summary.Metrics["sql_injection_check"] = m.Config.Validation.EnableSQLInjectionCheck
	}

	return summary
}

// SecuritySummary provides an overview of applied security measures
type SecuritySummary struct {
	Environment string                 `json:"environment"`
	Features    map[string]bool        `json:"features"`
	Metrics     map[string]interface{} `json:"metrics"`
}

// ValidateConfiguration validates the middleware configuration
func ValidateConfiguration(config MiddlewareConfig) error {
	// Validate security headers config
	if config.EnableSecurityHeaders {
		if err := ValidateSecurityConfig(config.SecurityHeaders); err != nil {
			return fmt.Errorf("security headers config invalid: %w", err)
		}
	}

	// Validate rate limiting config
	if config.EnableRateLimit {
		if config.RateLimit.GlobalRPS <= 0 {
			return fmt.Errorf("rate limit GlobalRPS must be positive")
		}
		if config.RateLimit.UserRPS <= 0 {
			return fmt.Errorf("rate limit UserRPS must be positive")
		}
	}

	// Validate validation config
	if config.EnableValidation {
		if config.Validation.MaxRequestSize <= 0 {
			return fmt.Errorf("validation MaxRequestSize must be positive")
		}
		if config.Validation.MaxFieldLength <= 0 {
			return fmt.Errorf("validation MaxFieldLength must be positive")
		}
	}

	return nil
}

// HealthCheck provides a health check for all middleware components
func (m *MiddlewareStack) HealthCheck() map[string]interface{} {
	health := make(map[string]interface{})

	health["middleware_stack"] = "healthy"
	health["environment"] = m.Config.Environment

	// Check individual components
	if m.Config.EnableRateLimit {
		health["rate_limiting"] = "enabled"
	}

	if m.Config.EnableSecurityHeaders {
		health["security_headers"] = "enabled"
	}

	if m.Config.EnableValidation {
		health["validation"] = "enabled"
	}

	if m.Config.EnableAuthorization {
		health["authorization"] = "enabled"
	}

	return health
}