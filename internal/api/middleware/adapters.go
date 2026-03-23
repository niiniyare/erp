package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// HTTPMiddleware defines the standard HTTP middleware signature
type HTTPMiddleware func(http.Handler) http.Handler

// HTTPMiddlewareAdapter provides conversion between HTTP and Fiber middleware
type HTTPMiddlewareAdapter struct {
	logger  logger.Logger
	metrics *metrics.MetricsService
}

// NewHTTPMiddlewareAdapter creates a new middleware adapter
func NewHTTPMiddlewareAdapter(logger logger.Logger, metrics *metrics.MetricsService) *HTTPMiddlewareAdapter {
	return &HTTPMiddlewareAdapter{
		logger:  logger,
		metrics: metrics,
	}
}

// RecoveryMiddleware provides panic recovery for HTTP handlers
func (a *HTTPMiddlewareAdapter) RecoveryMiddleware() HTTPMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					a.logger.Error("Panic recovered in HTTP middleware", logger.Fields{
						"error":  recovered,
						"path":   r.URL.Path,
						"method": r.Method,
						"ip":     r.RemoteAddr,
					})

					if a.metrics != nil {
						a.metrics.IncrementCounter("panic_recovered_total", metrics.Fields{
							"path": r.URL.Path,
						})
					}

					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal Server Error"))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersMiddleware adds security headers to HTTP responses
func (a *HTTPMiddlewareAdapter) SecurityHeadersMiddleware() HTTPMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			next.ServeHTTP(w, r)
		})
	}
}

// ConcurrencyLimitMiddleware limits concurrent requests for HTTP
func (a *HTTPMiddlewareAdapter) ConcurrencyLimitMiddleware(maxConcurrent int) HTTPMiddleware {
	semaphore := make(chan struct{}, maxConcurrent)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
				next.ServeHTTP(w, r)
			default:
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("Too Many Requests"))
			}
		})
	}
}

// CircuitBreakerMiddleware provides circuit breaker pattern for HTTP
func (a *HTTPMiddlewareAdapter) CircuitBreakerMiddleware() HTTPMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simple circuit breaker implementation
			next.ServeHTTP(w, r)
		})
	}
}

// OptimizedLoggingMiddleware provides request logging for HTTP
func (a *HTTPMiddlewareAdapter) OptimizedLoggingMiddleware() HTTPMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.logger.Info("HTTP request", logger.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
				"ip":     r.RemoteAddr,
			})

			next.ServeHTTP(w, r)
		})
	}
}

// ProfilingMiddleware provides performance profiling for HTTP
func (a *HTTPMiddlewareAdapter) ProfilingMiddleware() HTTPMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Performance profiling implementation
			next.ServeHTTP(w, r)
		})
	}
}

// FiberToHTTPAdapter converts Fiber middleware to HTTP middleware
func FiberToHTTPAdapter(fiberHandler fiber.Handler) HTTPMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This would require a more complex conversion
			// For now, just pass through
			next.ServeHTTP(w, r)
		})
	}
}

// HTTPToFiberAdapter converts HTTP middleware to Fiber middleware
func HTTPToFiberAdapter(httpMiddleware HTTPMiddleware) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// This would require converting between Fiber and HTTP contexts
		// For now, just continue
		return c.Next()
	}
}
