package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"

	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// HandlerHelper provides common functionality for all API handlers
type HandlerHelper struct {
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewHandlerHelper creates a new handler helper with dependencies
func NewHandlerHelper(logger logger.Logger, metrics metrics.MetricsProvider, tracer tracing.TracingService) *HandlerHelper {
	return &HandlerHelper{
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// Respond handles content negotiation and sends appropriate response
func (h *HandlerHelper) Respond(c *fiber.Ctx, statusCode int, data interface{}) error {
	acceptHeader := c.Get("Accept")

	// Determine response format based on Accept header
	if strings.Contains(acceptHeader, "text/html") {
		// For HTML responses, render component if data is string, otherwise convert to simple HTML
		if componentName, ok := data.(string); ok {
			return h.RenderComponent(c, componentName, nil)
		}
		// Simple HTML response for non-string data
		c.Set("Content-Type", "text/html; charset=utf-8")
		c.Status(statusCode)
		return c.SendString(fmt.Sprintf("<html><body><pre>%+v</pre></body></html>", data))
	}

	// Default to JSON response
	c.Set("Content-Type", "application/json; charset=utf-8")
	c.Status(statusCode)
	return c.JSON(data)
}

// RenderComponent renders a TemplUI component with data
func (h *HandlerHelper) RenderComponent(c *fiber.Ctx, component string, data interface{}) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Status(200)
	
	// Simple component rendering - in real implementation this would use TemplUI
	if data != nil {
		return c.SendString(fmt.Sprintf("<div class='%s'>%+v</div>", component, data))
	}
	return c.SendString(fmt.Sprintf("<div class='%s'></div>", component))
}

// HandleError processes errors and returns appropriate responses
func (h *HandlerHelper) HandleError(c *fiber.Ctx, err error) error {
	ctx := c.Context()
	
	// Log the error
	h.logger.ErrorContext(ctx, "Handler error occurred", logger.Fields{
		"error": err.Error(),
		"path":  c.Path(),
		"method": c.Method(),
	})

	// Determine status code based on error type
	var statusCode int
	var message string

	switch e := err.(type) {
	case ValidationError:
		statusCode = 400
		message = e.Error()
	case NotFoundError:
		statusCode = 404
		message = e.Error()
	default:
		statusCode = 500
		message = "Internal server error"
		// Don't expose internal errors to clients
		if strings.Contains(err.Error(), "database") || strings.Contains(err.Error(), "connection") {
			message = "Service temporarily unavailable"
		}
	}

	// Record error metrics
	h.metrics.IncrementCounter("handler_errors_total", metrics.Fields{
		"status_code": fmt.Sprintf("%d", statusCode),
		"error_type":  fmt.Sprintf("%T", err),
	})

	// Return error response
	return h.Respond(c, statusCode, map[string]interface{}{
		"error":   message,
		"status":  statusCode,
		"path":    c.Path(),
		"method":  c.Method(),
	})
}

// BadRequest returns a 400 Bad Request response
func (h *HandlerHelper) BadRequest(c *fiber.Ctx, message string) error {
	return h.HandleError(c, ValidationError{Field: "request", Message: message})
}

// Observability helper methods

// StartSpan starts a new tracing span
func (h *HandlerHelper) StartSpan(ctx context.Context, name string, opts ...tracing.SpanOption) (context.Context, tracing.Span) {
	return h.tracer.StartSpan(ctx, name, opts...)
}

// LogInfo logs an info message with context
func (h *HandlerHelper) LogInfo(ctx context.Context, msg string, fields logger.Fields) {
	h.logger.InfoContext(ctx, msg, fields)
}

// LogError logs an error message with context
func (h *HandlerHelper) LogError(ctx context.Context, msg string, err error, fields logger.Fields) {
	if fields == nil {
		fields = make(logger.Fields)
	}
	fields["error"] = err.Error()
	h.logger.ErrorContext(ctx, msg, fields)
}

// LogWarn logs a warning message with context
func (h *HandlerHelper) LogWarn(ctx context.Context, msg string, fields logger.Fields) {
	h.logger.WarnContext(ctx, msg, fields)
}

// LogDebug logs a debug message with context
func (h *HandlerHelper) LogDebug(ctx context.Context, msg string, fields logger.Fields) {
	h.logger.DebugContext(ctx, msg, fields)
}

// IncrementCounter increments a metrics counter
func (h *HandlerHelper) IncrementCounter(name string, labels metrics.Fields) {
	h.metrics.IncrementCounter(name, labels)
}

// ObserveHistogram records a histogram observation
func (h *HandlerHelper) ObserveHistogram(name string, value float64, labels metrics.Fields) {
	h.metrics.ObserveHistogram(name, value, labels)
}

// RecordTiming records a timing metric as histogram
func (h *HandlerHelper) RecordTiming(name string, duration time.Duration, labels metrics.Fields) {
	h.metrics.ObserveHistogram(name+"_duration_seconds", duration.Seconds(), labels)
}

// RecordError records an error in the current span
func (h *HandlerHelper) RecordError(ctx context.Context, err error, opts ...tracing.ErrorOption) {
	h.tracer.RecordError(ctx, err, opts...)
}

// AddEvent adds an event to the current span
func (h *HandlerHelper) AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	h.tracer.AddEvent(ctx, name, attrs...)
}

// SetAttributes sets attributes on the current span
func (h *HandlerHelper) SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	h.tracer.SetAttributes(ctx, attrs...)
}