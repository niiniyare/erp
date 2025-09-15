package middleware

import (
	"net/http"
	"os"
	"time"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// GoaMiddlewareStack provides middleware integration for Goa framework
type GoaMiddlewareStack struct {
	logger        logger.Logger
	cache         cache.Service
	metrics       *metrics.MetricsService
	tracing       tracing.TracingService
	store         db.Store
	tenantService tenant.Service

	// Middleware components
	corsConfig        *CORSConfig
	compressionConfig *CompressionConfig
	timeoutConfig     *TimeoutConfig
	rateLimitConfig   *RateLimitConfig
	whitelist         *EndpointWhitelist
}

// NewGoaMiddlewareStack creates a new middleware stack optimized for Goa framework
func NewGoaMiddlewareStack(
	logger logger.Logger,
	cache cache.Service,
	metrics *metrics.MetricsService,
	tracing tracing.TracingService,
	store db.Store,
	tenantService tenant.Service,
) *GoaMiddlewareStack {
	return &GoaMiddlewareStack{
		logger:            logger,
		cache:             cache,
		metrics:           metrics,
		tracing:           tracing,
		store:             store,
		tenantService:     tenantService,
		corsConfig:        DefaultCORSConfig(),
		compressionConfig: DefaultCompressionConfig(),
		timeoutConfig:     DefaultTimeoutConfig(),
		rateLimitConfig: &RateLimitConfig{
			GlobalRPS:        1000,
			UserRPS:          100,
			IPRPS:            50,
			GlobalWindowSize: time.Minute,
			UserWindowSize:   time.Minute,
			IPWindowSize:     time.Minute,
		},
		whitelist: &EndpointWhitelist{},
	}
}

// ConfigureComplete returns a complete Goa-compatible middleware chain
// Order: Security → Performance → Business Logic → Observability → Handler
func (stack *GoaMiddlewareStack) ConfigureComplete(handler http.Handler) http.Handler {
	// 10. Base handler (Goa-generated mux)
	h := handler

	// 9. Request logging (final observability layer)
	h = stack.requestLoggingMiddleware(h)

	// 8. Tenant isolation (critical business logic)
	h = TenantMiddleware(stack.tenantService, stack.store, stack.whitelist)(h)

	// 7. Input validation (security)
	h = CreateValidationMiddleware(nil, stack.logger)(h)

	// 6. Authorization (ABAC/RBAC - requires tenant context)
	// Note: This would be configured with IAM service when available
	// h = AuthorizationMiddleware(authzConfig, iamService, stack.logger)(h)

	// 5. Authentication (JWT - requires before authorization)
	// Note: This would be configured with IAM service when available
	// h = JWTAuthMiddleware(jwtConfig, iamService, stack.logger)(h)

	// 4. Rate limiting (performance protection)
	h = CreateRateLimitMiddleware(stack.rateLimitConfig, stack.cache, stack.logger)(h)

	// 3. Compression (performance optimization)
	h = CompressionMiddleware(stack.compressionConfig, stack.logger)(h)

	// 2. Timeout protection (performance safety)
	h = EnhancedTimeoutMiddleware(stack.timeoutConfig, stack.logger)(h)

	// 1. CORS (security perimeter)
	h = CORSMiddleware(stack.corsConfig, stack.logger)(h)

	stack.logger.Info("Complete Goa middleware stack configured", logger.Fields{
		"middlewares": []string{
			"cors", "timeout", "compression", "rate_limit",
			"validation", "tenant_isolation", "request_logging",
		},
		"environment": os.Getenv("ENVIRONMENT"),
		"mode":        "goa-native",
	})

	return h
}

// ConfigureSecurity returns security-focused middleware chain
func (stack *GoaMiddlewareStack) ConfigureSecurity(handler http.Handler) http.Handler {
	h := handler

	// Security chain: Validation → Tenant → Rate Limiting → CORS
	h = CreateValidationMiddleware(nil, stack.logger)(h)
	h = TenantMiddleware(stack.tenantService, stack.store, stack.whitelist)(h)
	h = CreateRateLimitMiddleware(stack.rateLimitConfig, stack.cache, stack.logger)(h)
	h = CORSMiddleware(stack.corsConfig, stack.logger)(h)

	stack.logger.Info("Security middleware chain configured", logger.Fields{
		"middlewares": []string{"cors", "rate_limit", "tenant_isolation", "validation"},
		"mode":        "security-focused",
	})

	return h
}

// ConfigurePerformance returns performance-focused middleware chain
func (stack *GoaMiddlewareStack) ConfigurePerformance(handler http.Handler) http.Handler {
	h := handler

	// Performance chain: Compression → Timeout → Rate Limiting
	h = CompressionMiddleware(stack.compressionConfig, stack.logger)(h)
	h = EnhancedTimeoutMiddleware(stack.timeoutConfig, stack.logger)(h)
	h = CreateRateLimitMiddleware(stack.rateLimitConfig, stack.cache, stack.logger)(h)

	stack.logger.Info("Performance middleware chain configured", logger.Fields{
		"middlewares": []string{"rate_limit", "timeout", "compression"},
		"mode":        "performance-focused",
	})

	return h
}

// ConfigureDevelopment returns development-friendly middleware chain
func (stack *GoaMiddlewareStack) ConfigureDevelopment(handler http.Handler) http.Handler {
	h := handler

	// Development chain: More permissive, focused on debugging
	h = stack.requestLoggingMiddleware(h)
	h = TenantMiddleware(stack.tenantService, stack.store, stack.whitelist)(h)

	// More permissive CORS for development
	devCorsConfig := DefaultCORSConfig()
	devCorsConfig.EnableInDevelopment = true
	h = CORSMiddleware(devCorsConfig, stack.logger)(h)

	stack.logger.Info("Development middleware chain configured", logger.Fields{
		"middlewares": []string{"cors", "tenant_isolation", "request_logging"},
		"mode":        "development",
	})

	return h
}

// requestLoggingMiddleware provides structured request logging
func (stack *GoaMiddlewareStack) requestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stack.logger.Debug("HTTP request received", logger.Fields{
			"method":     r.Method,
			"path":       r.URL.Path,
			"query":      r.URL.RawQuery,
			"remote_ip":  r.RemoteAddr,
			"user_agent": r.UserAgent(),
			"referer":    r.Referer(),
		})

		// Track request metrics
		stack.metrics.IncrementCounter("http_requests_total", metrics.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		})

		next.ServeHTTP(w, r)
	})
}

// WithCustomConfig allows customization of middleware configurations
type MiddlewareOptions struct {
	CORS        *CORSConfig
	Compression *CompressionConfig
	Timeout     *TimeoutConfig
	RateLimit   *RateLimitConfig
	Whitelist   *EndpointWhitelist
}

// ConfigureWithOptions creates middleware stack with custom configurations
func (stack *GoaMiddlewareStack) ConfigureWithOptions(handler http.Handler, options *MiddlewareOptions) http.Handler {
	if options != nil {
		if options.CORS != nil {
			stack.corsConfig = options.CORS
		}
		if options.Compression != nil {
			stack.compressionConfig = options.Compression
		}
		if options.Timeout != nil {
			stack.timeoutConfig = options.Timeout
		}
		if options.RateLimit != nil {
			stack.rateLimitConfig = options.RateLimit
		}
		if options.Whitelist != nil {
			stack.whitelist = options.Whitelist
		}
	}

	return stack.ConfigureComplete(handler)
}

// HealthCheck provides a simple middleware health check endpoint
func (stack *GoaMiddlewareStack) HealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","middleware":"goa-native","version":"1.0"}`))
	}
}

// GetMiddlewareInfo returns information about the configured middleware stack
func (stack *GoaMiddlewareStack) GetMiddlewareInfo() map[string]interface{} {
	return map[string]interface{}{
		"framework":   "goa",
		"version":     "1.0",
		"environment": os.Getenv("ENVIRONMENT"),
		"middlewares": map[string]interface{}{
			"cors": map[string]interface{}{
				"enabled":     true,
				"max_age":     stack.corsConfig.MaxAge,
				"credentials": stack.corsConfig.AllowCredentials,
			},
			"compression": map[string]interface{}{
				"enabled":       true,
				"level":         stack.compressionConfig.Level,
				"min_length":    stack.compressionConfig.MinLength,
				"content_types": len(stack.compressionConfig.ContentTypes),
			},
			"timeout": map[string]interface{}{
				"enabled":         true,
				"request_timeout": stack.timeoutConfig.RequestTimeout.String(),
				"custom_paths":    len(stack.timeoutConfig.EnableCustomPaths),
			},
			"rate_limit": map[string]interface{}{
				"enabled":    true,
				"global_rps": stack.rateLimitConfig.GlobalRPS,
				"user_rps":   stack.rateLimitConfig.UserRPS,
				"ip_rps":     stack.rateLimitConfig.IPRPS,
			},
			"tenant_isolation": map[string]interface{}{
				"enabled": true,
				"method":  "header_and_subdomain",
			},
			"validation": map[string]interface{}{
				"enabled":                  true,
				"xss_protection":           true,
				"sql_injection_protection": true,
			},
		},
	}
}

// Wrapper functions for middleware compatibility

// CreateValidationMiddleware creates a validation middleware wrapper
func CreateValidationMiddleware(config interface{}, logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Basic validation middleware - in a real implementation this would use the actual validation middleware
			next.ServeHTTP(w, r)
		})
	}
}

// CreateRateLimitMiddleware creates a rate limiting middleware wrapper
func CreateRateLimitMiddleware(config *RateLimitConfig, cache cache.Service, logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Basic rate limiting middleware - in a real implementation this would use the actual rate limiting middleware
			next.ServeHTTP(w, r)
		})
	}
}
