package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"

	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// Custom error types for HTTP handling
type ValidationError struct {
	Field   string
	Message string
}

func NewValidationError(f, m string) *ValidationError {
	return &ValidationError{
		Field:   f,
		Message: m,
	}
}

func (e ValidationError) Error() string {
	return e.Message
}

type NotFoundError struct {
	Resource string
	ID       string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.Resource)
}

// HandlerHelper provides common functionality for all API handlers
type HandlerHelper struct {
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.Service
}

// NewHandlerHelper creates a new handler helper with dependencies
func NewHandlerHelper(logger logger.Logger, metrics metrics.MetricsProvider, tracer tracing.Service) *HandlerHelper {
	return &HandlerHelper{
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

func NewNotFoundError(resource, id string) NotFoundError {
	return NotFoundError{Resource: resource, ID: id}
}

// Respond handles content negotiation and sends appropriate response
func (h *HandlerHelper) Respond(c *fiber.Ctx, statusCode int, data any) error {
	return h.RespondWithComponent(c, statusCode, data, nil)
}

// RespondWithComponent handles content negotiation with optional component override
func (h *HandlerHelper) RespondWithComponent(c *fiber.Ctx, statusCode int, data any, component templ.Component) error {
	// Check for content type preferences
	acceptHeader := c.Get("Accept")
	contentType := c.Get("Content-Type")
	isHTMX := c.Get("HX-Request") == "true"

	// Record response metrics
	h.metrics.IncrementCounter("handler_responses_total", metrics.Fields{
		"status_code":  fmt.Sprintf("%d", statusCode),
		"content_type": h.getResponseContentType(acceptHeader, contentType, isHTMX),
	})

	// Determine if client wants JSON
	wantsJSON := strings.Contains(acceptHeader, "application/json") ||
		strings.Contains(contentType, "application/json") ||
		(!strings.Contains(acceptHeader, "text/html") && !isHTMX && acceptHeader != "")

	if wantsJSON {
		// Return JSON response
		c.Set("Content-Type", "application/json; charset=utf-8")
		c.Status(statusCode)
		return c.JSON(data)
	}

	// Return HTML response using TemplUI component
	if component != nil {
		return h.RenderTemplComponent(c, statusCode, component)
	}

	// Fallback to simple HTML if no component provided
	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Status(statusCode)
	return c.SendString(fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head><title>Response</title></head>
		<body>
			<div class="container mx-auto p-4">
				<pre class="bg-gray-100 p-4 rounded">%+v</pre>
			</div>
		</body>
		</html>
	`, data))
}

// getResponseContentType determines the response content type for metrics
func (h *HandlerHelper) getResponseContentType(accept, contentType string, isHTMX bool) string {
	if strings.Contains(accept, "application/json") || strings.Contains(contentType, "application/json") {
		return "json"
	}
	if isHTMX {
		return "htmx"
	}
	if strings.Contains(accept, "text/html") {
		return "html"
	}
	return "unknown"
}

// RenderTemplComponent renders a templ.Component
func (h *HandlerHelper) RenderTemplComponent(c *fiber.Ctx, statusCode int, component templ.Component) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Status(statusCode)

	// Render the templ component to the response writer
	return component.Render(c.Context(), c.Response().BodyWriter())
}

// RenderComponent renders a TemplUI component with data (legacy method)
func (h *HandlerHelper) RenderComponent(c *fiber.Ctx, component string, data any) error {
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
		"error":  err.Error(),
		"path":   c.Path(),
		"method": c.Method(),
	})

	// Convert to HTTPError for consistent response format
	httpErr := errors.ToHTTPError(err)

	// Handle legacy custom error types
	switch e := err.(type) {
	case ValidationError:
		httpErr.Status = 400
		httpErr.Code = "VALIDATION_ERROR"
		httpErr.Message = e.Error()
		httpErr.Details["field"] = e.Field
	case NotFoundError:
		httpErr.Status = 404
		httpErr.Code = "NOT_FOUND"
		httpErr.Message = e.Error()
		httpErr.Details["resource"] = e.Resource
		if e.ID != "" {
			httpErr.Details["id"] = e.ID
		}
	}

	// Record error metrics
	h.metrics.IncrementCounter("handler_errors_total", metrics.Fields{
		"status_code": fmt.Sprintf("%d", httpErr.Status),
		"error_type":  fmt.Sprintf("%T", err),
		"error_code":  httpErr.Code,
	})

	// Return error response using HTTPError format
	return h.Respond(c, httpErr.Status, httpErr)
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
