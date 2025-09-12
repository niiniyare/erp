package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/niiniyare/erp/internal/shared/logger"
)

// TimeoutConfig defines timeout middleware configuration
type TimeoutConfig struct {
	RequestTimeout    time.Duration            `json:"request_timeout"` // Default request timeout
	HandlerTimeout    time.Duration            `json:"handler_timeout"` // Handler-specific timeout
	ReadTimeout       time.Duration            `json:"read_timeout"`    // Read timeout for request body
	WriteTimeout      time.Duration            `json:"write_timeout"`   // Write timeout for response
	EnableCustomPaths map[string]time.Duration `json:"custom_paths"`    // Path-specific timeouts
}

// DefaultTimeoutConfig returns production-ready timeout settings for ERP API
func DefaultTimeoutConfig() *TimeoutConfig {
	return &TimeoutConfig{
		RequestTimeout: 30 * time.Second, // Standard API request timeout
		HandlerTimeout: 25 * time.Second, // Slightly less than request timeout
		ReadTimeout:    10 * time.Second, // Request body read timeout
		WriteTimeout:   10 * time.Second, // Response write timeout
		EnableCustomPaths: map[string]time.Duration{
			"/api/v1/finance/reports": 120 * time.Second, // Financial reports need more time
			"/api/v1/analytics":       90 * time.Second,  // Analytics queries
			"/api/v1/imports":         300 * time.Second, // Data imports
			"/api/v1/exports":         180 * time.Second, // Data exports
			"/api/v1/tenant/migrate":  600 * time.Second, // Tenant migrations
		},
	}
}

// TimeoutMiddleware creates a request timeout middleware for Goa HTTP handlers
func TimeoutMiddleware(config *TimeoutConfig, logger logger.Logger) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultTimeoutConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Determine timeout for this specific path
			timeout := config.RequestTimeout

			// Check for path-specific timeout
			for path, pathTimeout := range config.EnableCustomPaths {
				if r.URL.Path == path || (r.URL.Path+"/" == path) {
					timeout = pathTimeout
					logger.Debug("Using custom timeout for path")
					break
				}
			}

			// Create timeout context
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a channel to track handler completion
			done := make(chan struct{})
			var panicVal interface{}

			// Create new request with timeout context
			r = r.WithContext(ctx)

			// Run the handler in a goroutine to detect timeout vs completion
			go func() {
				defer func() {
					if p := recover(); p != nil {
						panicVal = p
					}
					close(done)
				}()

				next.ServeHTTP(w, r)
			}()

			// Wait for either completion or timeout
			select {
			case <-done:
				// Handler completed normally
				if panicVal != nil {
					panic(panicVal) // Re-panic if handler panicked
				}

				logger.Debug("Request completed within timeout")

			case <-ctx.Done():
				// Request timed out
				if ctx.Err() == context.DeadlineExceeded {
					logger.Warn("Request timed out")

					// Send timeout response if headers haven't been written
					if !isHeadersWritten(w) {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusRequestTimeout)
						w.Write([]byte(`{"error":"request_timeout","message":"Request processing exceeded timeout limit","timeout":"` + timeout.String() + `"}`))
					}
				}
			}
		})
	}
}

// isHeadersWritten checks if response headers have already been written
// This is a best-effort check and may not be 100% reliable across all ResponseWriter implementations
func isHeadersWritten(w http.ResponseWriter) bool {
	// Try to cast to common ResponseWriter types that track header status
	type headerChecker interface {
		Written() bool
	}

	if hc, ok := w.(headerChecker); ok {
		return hc.Written()
	}

	// Fallback: assume headers are not written
	// In practice, this is safer than trying to write after headers are sent
	return false
}

// TimeoutResponseWriter wraps http.ResponseWriter to track if headers have been written
type TimeoutResponseWriter struct {
	http.ResponseWriter
	written bool
}

// WriteHeader tracks that headers have been written
func (trw *TimeoutResponseWriter) WriteHeader(statusCode int) {
	if !trw.written {
		trw.written = true
		trw.ResponseWriter.WriteHeader(statusCode)
	}
}

// Write tracks that headers have been written
func (trw *TimeoutResponseWriter) Write(data []byte) (int, error) {
	if !trw.written {
		trw.WriteHeader(http.StatusOK)
	}
	return trw.ResponseWriter.Write(data)
}

// Written returns whether headers have been written
func (trw *TimeoutResponseWriter) Written() bool {
	return trw.written
}

// Flush implements http.Flusher if supported by underlying ResponseWriter
func (trw *TimeoutResponseWriter) Flush() {
	if flusher, ok := trw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Enhanced timeout middleware that wraps the ResponseWriter for better timeout detection
func EnhancedTimeoutMiddleware(config *TimeoutConfig, logger logger.Logger) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultTimeoutConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Wrap the ResponseWriter to track header status
			trw := &TimeoutResponseWriter{ResponseWriter: w}

			// Determine timeout for this specific path
			timeout := config.RequestTimeout

			// Check for path-specific timeout
			for path, pathTimeout := range config.EnableCustomPaths {
				if r.URL.Path == path {
					timeout = pathTimeout
					break
				}
			}

			// Create timeout context
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a channel to track handler completion
			done := make(chan struct{})
			var panicVal interface{}

			// Create new request with timeout context
			r = r.WithContext(ctx)

			// Run the handler in a goroutine
			go func() {
				defer func() {
					if p := recover(); p != nil {
						panicVal = p
					}
					close(done)
				}()

				next.ServeHTTP(trw, r)
			}()

			// Wait for either completion or timeout
			select {
			case <-done:
				// Handler completed normally
				if panicVal != nil {
					panic(panicVal) // Re-panic if handler panicked
				}

			case <-ctx.Done():
				// Request timed out
				if ctx.Err() == context.DeadlineExceeded && !trw.Written() {
					logger.Warn("Request timed out")

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusRequestTimeout)
					w.Write([]byte(`{"error":"request_timeout","message":"Request processing exceeded timeout limit","timeout":"` + timeout.String() + `"}`))
				}
			}
		})
	}
}
