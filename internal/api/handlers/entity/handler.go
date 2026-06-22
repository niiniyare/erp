// Package entity provides HTTP handlers for entity management endpoints.
//
// File layout:
//
//	handler.go — EntityHandler struct, constructor
//	base.go    — Shared response helpers, error handling, pagination
//	crud.go    — CRUD and hierarchy handler methods
package entity

import (
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	entityDomain "awo.so/internal/core/entity"
	sharedErrors "awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// EntityHandler handles all entity HTTP endpoints.
type EntityHandler struct {
	service   entityDomain.Service
	logger    logger.Logger
	metrics   metrics.MetricsProvider
	tracer    tracing.Service
	validator *validator.Validate
}

// NewEntityHandler constructs an EntityHandler.
func NewEntityHandler(
	svc entityDomain.Service,
	log logger.Logger,
	m metrics.MetricsProvider,
	tracer tracing.Service,
) *EntityHandler {
	return &EntityHandler{
		service:   svc,
		logger:    log,
		metrics:   m,
		tracer:    tracer,
		validator: validator.New(),
	}
}

// ============================================================================
// Response helpers
// ============================================================================

func (h *EntityHandler) ok200(c *fiber.Ctx, data any) error {
	h.recordSuccess(c)
	return c.JSON(envelope(c, data, nil))
}

func (h *EntityHandler) ok200Meta(c *fiber.Ctx, data any, meta map[string]any) error {
	h.recordSuccess(c)
	return c.JSON(envelope(c, data, meta))
}

func (h *EntityHandler) ok201(c *fiber.Ctx, data any) error {
	h.recordSuccess(c)
	return c.Status(fiber.StatusCreated).JSON(envelope(c, data, nil))
}

func (h *EntityHandler) ok204(c *fiber.Ctx) error {
	h.recordSuccess(c)
	return c.SendStatus(fiber.StatusNoContent)
}

func envelope(c *fiber.Ctx, data any, meta map[string]any) fiber.Map {
	resp := fiber.Map{
		"success":    true,
		"data":       data,
		"request_id": requestID(c),
		"timestamp":  time.Now(),
	}
	if meta != nil {
		resp["meta"] = meta
	}
	return resp
}

// ============================================================================
// Error handling
// ============================================================================

func (h *EntityHandler) fail(c *fiber.Ctx, err error) error {
	httpErr := sharedErrors.ToHTTPError(err)
	httpErr.RequestID = requestID(c)

	h.logger.Error("entity handler error", logger.Fields{
		"error":      err.Error(),
		"code":       httpErr.Code,
		"status":     httpErr.Status,
		"request_id": httpErr.RequestID,
		"method":     c.Method(),
		"path":       c.Path(),
	})
	h.recordError(c, httpErr)
	return c.Status(httpErr.Status).JSON(httpErr)
}

// bind parses the request body into dst and validates it.
func (h *EntityHandler) bind(c *fiber.Ctx, dst any) error {
	if err := c.BodyParser(dst); err != nil {
		return sharedErrors.NewBusinessError("INVALID_JSON", "request body contains invalid JSON").
			WithHTTPStatus(fiber.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation)
	}
	if err := h.validator.Struct(dst); err != nil {
		return err
	}
	return nil
}

// ============================================================================
// Context helpers
// ============================================================================

func requestID(c *fiber.Ctx) string {
	if v := c.Locals("requestid"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return "unknown"
}

// ============================================================================
// Pagination
// ============================================================================

func (h *EntityHandler) pagination(c *fiber.Ctx) (offset, limit int) {
	offset = c.QueryInt("offset", 0)
	limit = c.QueryInt("limit", 20)
	if offset < 0 {
		offset = 0
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return offset, limit
}

// ============================================================================
// Metrics
// ============================================================================

func (h *EntityHandler) recordSuccess(c *fiber.Ctx) {
	h.metrics.IncrementCounter("entity_api_requests_total", metrics.Fields{
		"method":   c.Method(),
		"endpoint": c.Route().Path,
		"status":   "success",
	})
}

func (h *EntityHandler) recordError(c *fiber.Ctx, httpErr *sharedErrors.HTTPError) {
	h.metrics.IncrementCounter("entity_api_errors_total", metrics.Fields{
		"method":   c.Method(),
		"endpoint": c.Route().Path,
		"code":     httpErr.Code,
		"status":   strconv.Itoa(httpErr.Status),
	})
}
