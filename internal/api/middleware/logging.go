package middleware

import (
	"net/http"
	"time"

	"awo/internal/shared/logger"
)

// responseWriterWrapper wraps http.ResponseWriter to capture the status code.
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// RequestLogger creates a middleware for logging HTTP requests compatible with Goa.
func RequestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := r.URL.Path
			method := r.Method

			// Wrap the response writer to capture the status code
			wrappedWriter := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default status if not set
			}

			// Process request
			next.ServeHTTP(wrappedWriter, r)

			// Log request details
			duration := time.Since(start)
			statusCode := wrappedWriter.statusCode
			clientIP := r.RemoteAddr
			userAgent := r.UserAgent()

			logLevel := logger.InfoLevel
			if statusCode >= 400 && statusCode < 500 {
				logLevel = logger.WarnLevel
			} else if statusCode >= 500 {
				logLevel = logger.ErrorLevel
			}

			fields := logger.Fields{
				"method":      method,
				"path":        path,
				"status_code": statusCode,
				"duration_ms": float64(duration.Nanoseconds()) / 1e6,
				"client_ip":   clientIP,
				"user_agent":  userAgent,
			}

			// Note: Goa doesn't have an equivalent to c.Errors, so we skip that part

			message := "HTTP request completed"
			switch logLevel {
			case logger.WarnLevel:
				logger.WarnContext(r.Context(), message, fields)
			case logger.ErrorLevel:
				logger.ErrorContext(r.Context(), message, fields)
			default:
				logger.InfoContext(r.Context(), message, fields)
			}
		})
	}
}
