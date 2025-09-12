package middleware

// DEPRECATED: This file contains legacy Gin middleware that is no longer used.
// The Goa server uses native HTTP middleware for tracing and metrics.
// These functions are kept for reference but are commented out.

/*
// TracingMiddleware creates a middleware for distributed tracing (DEPRECATED: Use native HTTP middleware)
func TracingMiddleware(tracingService tracing.TracingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract tracing context from HTTP headers
		ctx := tracingService.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)

		// Start HTTP span
		ctx, span := tracingService.StartSpan(ctx, "http.request",
			tracing.WithSpanKind(tracing.SpanKindServer),
			tracing.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.url", c.Request.URL.String()),
				attribute.String("http.route", c.FullPath()),
				attribute.String("http.user_agent", c.Request.UserAgent()),
			))
		defer span.End()

		// Set enriched context
		c.Request = c.Request.WithContext(ctx)

		// Record HTTP request metrics
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		// Add response attributes to span
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.String("http.response.size", string(rune(c.Writer.Size()))),
			attribute.Float64("http.response.duration_ms", float64(duration.Nanoseconds())/1e6),
		)

		// Record error if status >= 400
		if c.Writer.Status() >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", c.Writer.Status()))
		}
	}
}

// MetricsMiddleware creates a middleware for collecting HTTP metrics (DEPRECATED: Use native HTTP middleware)
func MetricsMiddleware(metricsService *metrics.MetricsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Record request start
		metricsService.IncrementCounter("http_requests_total", metrics.Fields{
			"method": c.Request.Method,
			"path":   c.FullPath(),
		})

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		// Record request completion metrics
		metricsService.ObserveHistogram("http_request_duration_seconds", duration.Seconds(), metrics.Fields{
			"method": c.Request.Method,
			"path":   c.FullPath(),
			"status": string(rune(status)),
		})

		metricsService.IncrementCounter("http_responses_total", metrics.Fields{
			"method": c.Request.Method,
			"path":   c.FullPath(),
			"status": string(rune(status)),
		})

		// Record response size
		responseSize := c.Writer.Size()
		metricsService.ObserveHistogram("http_response_size_bytes", float64(responseSize), metrics.Fields{
			"method": c.Request.Method,
			"path":   c.FullPath(),
		})

		// Track error rates
		if status >= 400 {
			metricsService.IncrementCounter("http_errors_total", metrics.Fields{
				"method": c.Request.Method,
				"path":   c.FullPath(),
				"status": string(rune(status)),
			})
		}
	}
}
*/