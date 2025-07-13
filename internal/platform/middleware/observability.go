package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// TracingMiddleware creates a middleware for distributed tracing
func TracingMiddleware(tracingService *tracing.TracingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract tracing context from HTTP headers
		ctx := tracingService.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)
		
		// Start HTTP span
		ctx, span := tracingService.StartSpan(ctx, "http.request",
			tracing.WithSpanKind(tracing.SpanKindServer),
			tracing.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.url", c.Request.URL.String()),
				attribute.String("http.scheme", c.Request.URL.Scheme),
				attribute.String("http.host", c.Request.Host),
				attribute.String("http.target", c.Request.URL.Path),
				attribute.String("http.user_agent", c.Request.UserAgent()),
				attribute.String("http.remote_addr", c.ClientIP()),
			))
		defer span.End()
		
		// Update request context with tracing context
		c.Request = c.Request.WithContext(ctx)
		
		// Process request
		c.Next()
		
		// Set response attributes
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int("http.response_size", c.Writer.Size()),
		)
		
		// Record errors if any
		if len(c.Errors) > 0 {
			tracingService.RecordError(ctx, c.Errors.Last().Err, tracing.WithErrorStatus())
		}
		
		// Inject tracing headers into response
		tracingService.InjectHTTPHeaders(ctx, c.Writer.Header())
	}
}

// MetricsMiddleware creates a middleware for collecting HTTP metrics
func MetricsMiddleware(metricsService *metrics.MetricsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		
		// Process request
		c.Next()
		
		// Collect metrics
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		
		// Increment request counter
		metricsService.IncrementCounter("http_requests_total", metrics.Fields{
			"method": method,
			"path":   path,
			"status": statusCodeClass(statusCode),
		})
		
		// Record request duration
		metricsService.ObserveHistogram("http_request_duration_seconds", 
			duration.Seconds(), metrics.Fields{
				"method": method,
				"path":   path,
			})
		
		// Record response size
		if c.Writer.Size() > 0 {
			metricsService.ObserveHistogram("http_response_size_bytes", 
				float64(c.Writer.Size()), metrics.Fields{
					"method": method,
					"path":   path,
				})
		}
		
		// Count errors
		if statusCode >= 400 {
			metricsService.IncrementCounter("http_errors_total", metrics.Fields{
				"method": method,
				"path":   path,
				"status": statusCodeClass(statusCode),
			})
		}
	}
}

// statusCodeClass returns the status code class (1xx, 2xx, 3xx, 4xx, 5xx)
func statusCodeClass(statusCode int) string {
	switch {
	case statusCode >= 100 && statusCode < 200:
		return "1xx"
	case statusCode >= 200 && statusCode < 300:
		return "2xx"
	case statusCode >= 300 && statusCode < 400:
		return "3xx"
	case statusCode >= 400 && statusCode < 500:
		return "4xx"
	case statusCode >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}