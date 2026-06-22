// Package iam provides HTTP handlers for IAM management endpoints.
//
// File layout:
//
//	handler.go — IAMHandler struct, constructor, shared helpers
//	crud.go    — Policy and role assignment handler methods
package iam

import (
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"awo.so/internal/core/iam/contract"
	sharedErrors "awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// IAMHandler handles IAM management HTTP endpoints.
type IAMHandler struct {
	service   contract.AuthzService
	logger    logger.Logger
	metrics   metrics.MetricsProvider
	tracer    tracing.Service
	validator *validator.Validate
}

// NewIAMHandler constructs an IAMHandler.
func NewIAMHandler(
	svc contract.AuthzService,
	log logger.Logger,
	m metrics.MetricsProvider,
	tracer tracing.Service,
) *IAMHandler {
	return &IAMHandler{
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

func (h *IAMHandler) ok200(c *fiber.Ctx, data any) error {
	h.recordSuccess(c)
	return c.JSON(envelope(c, data, nil))
}

func (h *IAMHandler) ok204(c *fiber.Ctx) error {
	h.recordSuccess(c)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *IAMHandler) ok201(c *fiber.Ctx, data any) error {
	h.recordSuccess(c)
	return c.Status(fiber.StatusCreated).JSON(envelope(c, data, nil))
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

func (h *IAMHandler) fail(c *fiber.Ctx, err error) error {
	httpErr := sharedErrors.ToHTTPError(err)
	httpErr.RequestID = requestID(c)

	h.logger.Error("iam handler error", logger.Fields{
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

func (h *IAMHandler) bind(c *fiber.Ctx, dst any) error {
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

// tenantDomain returns the Casbin domain for the current request.
// Falls back to the ?domain query param for platform-actor cross-tenant ops.
func tenantDomain(c *fiber.Ctx) string {
	if v := c.Locals("tenant_id"); v != nil {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return c.Query("domain")
}

// ============================================================================
// Metrics
// ============================================================================

func (h *IAMHandler) recordSuccess(c *fiber.Ctx) {
	h.metrics.IncrementCounter("iam_api_requests_total", metrics.Fields{
		"method":   c.Method(),
		"endpoint": c.Route().Path,
		"status":   "success",
	})
}

func (h *IAMHandler) recordError(c *fiber.Ctx, httpErr *sharedErrors.HTTPError) {
	h.metrics.IncrementCounter("iam_api_errors_total", metrics.Fields{
		"method":   c.Method(),
		"endpoint": c.Route().Path,
		"code":     httpErr.Code,
		"status":   strconv.Itoa(httpErr.Status),
	})
}
