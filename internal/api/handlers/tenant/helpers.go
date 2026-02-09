package tenant

import (
	"time"

	"github.com/gofiber/fiber/v2"
	coreTenant "github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
)

// handleError provides centralized error handling for the tenant module.
// It logs the error, records metrics, and formats a structured error response.
func (h *TenantHandler) handleError(c *fiber.Ctx, err error) error {
	requestID := getRequestID(c)
	// Convert any error into a structured HTTPError
	httpErr := errors.ToHTTPError(err)
	httpErr.RequestID = requestID

	h.logger.Error("Request error", logger.Fields{
		"error":      err.Error(),
		"request_id": requestID,
		"tenant_id":  getTenantID(c),
		"method":     c.Method(),
		"path":       c.Path(),
		"ip":         c.IP(),
		"status":     httpErr.Status,
		"code":       httpErr.Code,
	})

	h.metrics.IncrementCounter("api_errors_total", metrics.Fields{
		"method":   c.Method(),
		"endpoint": c.Route().Path,
		"code":     httpErr.Code,
		"status":   httpErr.Status,
	})

	return c.Status(httpErr.Status).JSON(httpErr)
}

// validateRequest parses and validates the request body against the provided struct.
func (h *TenantHandler) validateRequest(c *fiber.Ctx, req interface{}) error {
	if err := c.BodyParser(req); err != nil {
		return errors.NewBusinessError("INVALID_JSON", "Invalid JSON format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation).
			WithSuggestion("Ensure request body contains valid JSON")
	}
	if err := h.validator.Struct(req); err != nil {
		return errors.NewBusinessError("VALIDATION_ERROR", err.Error()).
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation)
	}
	return nil
}

// success returns a standardized 200 OK success response.
func (h *TenantHandler) success(c *fiber.Ctx, data interface{}) error {
	response := fiber.Map{
		"success":    true,
		"data":       data,
		"meta": fiber.Map{
			"request_id": getRequestID(c),
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		},
	}
	h.recordSuccessMetrics(c)
	return c.JSON(response)
}

// successWithMeta returns a 200 OK response with additional top-level metadata.
func (h *TenantHandler) successWithMeta(c *fiber.Ctx, data interface{}, meta fiber.Map) error {
	response := fiber.Map{
		"success":    true,
		"data":       data,
		"meta":       meta,
	}
	// Add standard meta fields
	meta["request_id"] = getRequestID(c)
	meta["timestamp"] = time.Now().UTC().Format(time.RFC3339)

	h.recordSuccessMetrics(c)
	return c.JSON(response)
}

// created returns a 201 Created response.
func (h *TenantHandler) created(c *fiber.Ctx, data interface{}) error {
	response := fiber.Map{
		"success":    true,
		"data":       data,
		"meta": fiber.Map{
			"request_id": getRequestID(c),
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		},
	}
	h.recordSuccessMetrics(c)
	return c.Status(fiber.StatusCreated).JSON(response)
}

// accepted returns a 202 Accepted response for async operations.
func (h *TenantHandler) accepted(c *fiber.Ctx, data interface{}) error {
	response := fiber.Map{
		"success":    true,
		"data":       data,
		"meta": fiber.Map{
			"request_id": getRequestID(c),
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		},
	}
	h.recordSuccessMetrics(c)
	return c.Status(fiber.StatusAccepted).JSON(response)
}

// extractListParams extracts and validates all list-related query parameters.
func (h *TenantHandler) extractListParams(c *fiber.Ctx) coreTenant.TenantFilter {
	offset := c.QueryInt("offset", 0)
	limit := c.QueryInt("limit", 20)

	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 1
	}
	if offset < 0 {
		offset = 0
	}

	search := c.Query("search")
	status := c.Query("status")
	sortBy := c.Query("sort_by", "created_at")

	filter := coreTenant.TenantFilter{
		Offset: int32(offset),
		Limit:  int32(limit),
		SortBy: sortBy,
	}
	if search != "" {
		filter.NameFilter = &search
	}
	if status != "" {
		filter.StatusFilter = &status
	}
	return filter
}

// --- private helpers ---

func getRequestID(c *fiber.Ctx) string {
	if requestID, ok := c.Locals("requestid").(string); ok {
		return requestID
	}
	return "unknown"
}

func getTenantID(c *fiber.Ctx) string {
	if tenantID, ok := c.Locals("tenant_id").(string); ok {
		return tenantID
	}
	return ""
}

func (h *TenantHandler) recordSuccessMetrics(c *fiber.Ctx) {
	h.metrics.IncrementCounter("api_requests_total", metrics.Fields{
		"method":   c.Method(),
		"endpoint": c.Route().Path,
		"status":   "success",
	})
}
