package middleware

import (
	"net/http"
	"strconv"
	"time"

	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// SimpleObservabilityMiddleware creates a basic observability middleware for Goa services
func SimpleObservabilityMiddleware(
	logger logger.Logger,
	metrics *metrics.MetricsService,
	tracer tracing.Service,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			ctx := r.Context()

			// Start tracing span
			var span tracing.Span
			ctx, span = tracer.StartSpan(ctx, "http.request")
			defer span.End()

			// Set span attributes
			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.host", r.Host),
			)

			// Create response writer wrapper
			wrw := &simpleResponseWriter{ResponseWriter: w, statusCode: 200}

			// Update request context
			r = r.WithContext(ctx)

			// Log request start (simplified)
			logger.InfoContext(ctx, "HTTP request started")

			// Increment request counter (without fields for now)
			emptyFields := make(map[string]any)
			emptyFields["method"] = r.Method
			emptyFields["path"] = r.URL.Path
			metrics.IncrementCounter("http_requests_total", emptyFields)

			// Execute the handler
			next.ServeHTTP(wrw, r)

			// Calculate duration
			duration := time.Since(startTime)

			// Update span with response data
			span.SetAttributes(
				attribute.Int("http.status_code", wrw.statusCode),
				attribute.Float64("http.duration_ms", float64(duration.Nanoseconds())/1e6),
			)

			if wrw.statusCode >= 400 {
				span.SetStatus(codes.Error, http.StatusText(wrw.statusCode))
			} else {
				span.SetStatus(codes.Ok, "Request completed successfully")
			}

			// Record response metrics
			responseFields := make(map[string]any)
			responseFields["method"] = r.Method
			responseFields["path"] = r.URL.Path
			responseFields["status"] = strconv.Itoa(wrw.statusCode)

			metrics.IncrementCounter("http_responses_total", responseFields)
			metrics.ObserveHistogram("http_request_duration_seconds", duration.Seconds(), responseFields)

			// Log request completion (simplified)
			if wrw.statusCode >= 500 {
				logger.ErrorContext(ctx, "HTTP request failed with server error")
			} else if wrw.statusCode >= 400 {
				logger.WarnContext(ctx, "HTTP request failed with client error")
			} else {
				logger.InfoContext(ctx, "HTTP request completed")
			}
		})
	}
}

// simpleResponseWriter wraps http.ResponseWriter to capture response metrics
type simpleResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (srw *simpleResponseWriter) WriteHeader(statusCode int) {
	srw.statusCode = statusCode
	srw.ResponseWriter.WriteHeader(statusCode)
}

// Write ensures status code is set
func (srw *simpleResponseWriter) Write(data []byte) (int, error) {
	if srw.statusCode == 0 {
		srw.statusCode = 200
	}
	return srw.ResponseWriter.Write(data)
}
