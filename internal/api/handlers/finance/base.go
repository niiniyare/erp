package finance

// base.go — shared HTTP response helpers used by every finance handler.
//
// All handler methods call into these helpers rather than writing raw
// c.Status(...).JSON(...) calls. This guarantees:
//   - Consistent response envelope across all finance endpoints
//   - Single place to change the envelope shape
//   - Error logging and metrics recorded on every path
//
// Response envelope shape:
//
//	{
//	  "success":    true | false,
//	  "data":       <payload>,
//	  "meta":       <optional metadata>,
//	  "request_id": "<uuid>",
//	  "timestamp":  "<rfc3339>"
//	}

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	sharedErrors "awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// ============================================================================
// Response helpers
// ============================================================================

// ok200 writes a 200 JSON response with the standard success envelope.
// Use this for GET and PUT endpoints that return the affected resource.
func (h *FinanceHandler) ok200(c *fiber.Ctx, data any) error {
	h.recordSuccess(c)
	return c.JSON(envelope(c, data, nil))
}

// ok200Meta writes a 200 response with an additional metadata object.
// Use this for list endpoints that include pagination or filter context.
func (h *FinanceHandler) ok200Meta(c *fiber.Ctx, data any, meta map[string]any) error {
	h.recordSuccess(c)
	return c.JSON(envelope(c, data, meta))
}

// ok201 writes a 201 Created response.
// Use this for POST endpoints that create a new resource.
func (h *FinanceHandler) ok201(c *fiber.Ctx, data any) error {
	h.recordSuccess(c)
	return c.Status(fiber.StatusCreated).JSON(envelope(c, data, nil))
}

// ok204 writes a 204 No Content response.
// Use this for DELETE endpoints. No body is written.
func (h *FinanceHandler) ok204(c *fiber.Ctx) error {
	h.recordSuccess(c)
	return c.SendStatus(fiber.StatusNoContent)
}

// notImplemented writes a 501 Not Implemented response.
// Use this as a placeholder for endpoints whose service methods are not yet built.
// The handler is registered in routing so clients receive a clear signal rather
// than a 404.
func (h *FinanceHandler) notImplemented(c *fiber.Ctx, feature string) error {
	httpErr := &sharedErrors.HTTPError{
		Status:    fiber.StatusNotImplemented,
		Code:      "NOT_IMPLEMENTED",
		Message:   feature + " is not yet implemented",
		RequestID: requestID(c),
		Timestamp: time.Now(),
	}
	return c.Status(fiber.StatusNotImplemented).JSON(httpErr)
}

// envelope builds the standard response body.
// meta is omitted from the JSON output when nil.
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

// fail converts any error into an HTTP response and logs it with request
// context. It is the single exit point for all error paths in finance handlers.
//
// Error → HTTP mapping is delegated to sharedErrors.ToHTTPError which walks
// the error chain via errors.As, so BusinessErrors, RepositoryErrors, and
// plain sentinel errors are all handled correctly.
func (h *FinanceHandler) fail(c *fiber.Ctx, err error) error {
	httpErr := sharedErrors.ToHTTPError(err)
	httpErr.RequestID = requestID(c)

	h.logger.Error("finance handler error", logger.Fields{
		"error":      err.Error(),
		"code":       httpErr.Code,
		"status":     httpErr.Status,
		"request_id": httpErr.RequestID,
		"tenant_id":  tenantID(c),
		"method":     c.Method(),
		"path":       c.Path(),
		"ip":         c.IP(),
	})

	h.recordError(c, httpErr)

	return c.Status(httpErr.Status).JSON(httpErr)
}

// ============================================================================
// Request parsing
// ============================================================================

// bind parses the request body into dst and runs struct validation.
//
// Returns a ready-to-pass-to-fail error on any parse or validation failure;
// the caller only needs to check err != nil and return h.fail(c, err).
//
// Example:
//
//	var req financeDomain.CreateAccountRequest
//	if err := h.bind(c, &req); err != nil {
//	    return h.fail(c, err)
//	}
func (h *FinanceHandler) bind(c *fiber.Ctx, dst any) error {
	if err := c.BodyParser(dst); err != nil {
		return sharedErrors.NewBusinessError("INVALID_JSON", "request body contains invalid JSON").
			WithHTTPStatus(fiber.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation).
			WithSuggestion("Ensure Content-Type is application/json and the body is well-formed")
	}
	if err := h.validator.Struct(dst); err != nil {
		return err // go-playground/validator errors are handled by ToHTTPError
	}
	return nil
}

// ============================================================================
// Pagination
// ============================================================================

// pagination extracts offset/limit query parameters with sane defaults and
// caps. All list endpoints use this; never read offset/limit inline.
//
// Defaults: offset=0, limit=20
// Caps:     limit max=100
func (h *FinanceHandler) pagination(c *fiber.Ctx) (offset, limit int) {
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
// Context value extraction
// ============================================================================

// requestID returns the request ID injected by the observability middleware.
// Falls back to "unknown" if not present so log lines are never empty.
func requestID(c *fiber.Ctx) string {
	if v := c.Locals("requestid"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return "unknown"
}

// tenantID returns the tenant ID injected by TenantMiddleware.
// Used in error logs only — never use this to make business decisions;
// read tenant from context via shared.GetTenantID(ctx) in the service layer.
func tenantID(c *fiber.Ctx) string {
	if v := c.Locals("tenant_id"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// boolQuery parses a "true"/"false" query parameter.
// Returns nil when the key is absent, so callers can distinguish
// "not provided" from an explicit false.
func boolQuery(c *fiber.Ctx, key string) *bool {
	switch c.Query(key) {
	case "true":
		v := true
		return &v
	case "false":
		v := false
		return &v
	default:
		return nil
	}
}

// ============================================================================
// Metrics
// ============================================================================

// recordSuccess increments the success counter for the current route.
func (h *FinanceHandler) recordSuccess(c *fiber.Ctx) {
	h.metrics.IncrementCounter("finance_api_requests_total", metrics.Fields{
		"method":   c.Method(),
		"endpoint": c.Route().Path,
		"status":   "success",
	})
}

// recordError increments the error counter for the current route.
//
// Note: httpErr.Status is an int; strconv.Itoa is required — string(rune(int))
// would produce the Unicode character at that code point, not the digit string.
func (h *FinanceHandler) recordError(c *fiber.Ctx, httpErr *sharedErrors.HTTPError) {
	h.metrics.IncrementCounter("finance_api_errors_total", metrics.Fields{
		"method":   c.Method(),
		"endpoint": c.Route().Path,
		"code":     httpErr.Code,
		"status":   strconv.Itoa(httpErr.Status),
	})
}
