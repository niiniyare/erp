package contracts

import (
	"errors"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"awo.so/internal/core/contracts/domain"
	sharedErrors "awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// mapContractError converts domain-layer contract errors to structured BusinessErrors
// with the correct HTTP status codes.
func mapContractError(err error) error {
	switch {
	case errors.Is(err, domain.ErrContractNotFound):
		return sharedErrors.NewBusinessError("CONTRACT_NOT_FOUND", "Contract not found").
			WithHTTPStatus(http.StatusNotFound).
			WithCategory(sharedErrors.CategoryBusiness).
			WithSuggestion("Verify the contract ID is correct")

	case errors.Is(err, domain.ErrContractAlreadyExists):
		return sharedErrors.NewBusinessError("CONTRACT_ALREADY_EXISTS", "A contract with that number already exists").
			WithHTTPStatus(http.StatusConflict).
			WithCategory(sharedErrors.CategoryBusiness)

	case errors.Is(err, domain.ErrVersionConflict):
		return sharedErrors.NewBusinessError("VERSION_CONFLICT", "Contract was modified by another request").
			WithHTTPStatus(http.StatusConflict).
			WithCategory(sharedErrors.CategoryBusiness).
			WithSuggestion("Reload the contract and retry with the latest version number")

	case errors.Is(err, domain.ErrInvalidTransition):
		return sharedErrors.NewBusinessError("INVALID_STATUS_TRANSITION", "This status transition is not allowed").
			WithHTTPStatus(http.StatusUnprocessableEntity).
			WithCategory(sharedErrors.CategoryBusiness).
			WithSuggestion("Check the contract's current status and allowed transitions")

	case errors.Is(err, domain.ErrCannotModifyTerminated):
		return sharedErrors.NewBusinessError("CONTRACT_TERMINATED", "Terminated contracts cannot be modified").
			WithHTTPStatus(http.StatusUnprocessableEntity).
			WithCategory(sharedErrors.CategoryBusiness)

	case errors.Is(err, domain.ErrCannotModifyExpired):
		return sharedErrors.NewBusinessError("CONTRACT_EXPIRED", "Expired contracts cannot be modified").
			WithHTTPStatus(http.StatusUnprocessableEntity).
			WithCategory(sharedErrors.CategoryBusiness)

	case errors.Is(err, domain.ErrAlreadyTerminated):
		return sharedErrors.NewBusinessError("ALREADY_TERMINATED", "Contract is already terminated").
			WithHTTPStatus(http.StatusConflict).
			WithCategory(sharedErrors.CategoryBusiness)

	case errors.Is(err, domain.ErrAlreadyExpired):
		return sharedErrors.NewBusinessError("ALREADY_EXPIRED", "Contract is already expired").
			WithHTTPStatus(http.StatusConflict).
			WithCategory(sharedErrors.CategoryBusiness)

	case errors.Is(err, domain.ErrEndDateBeforeStart):
		return sharedErrors.NewBusinessError("INVALID_DATE_RANGE", "End date must be after start date").
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation)

	case errors.Is(err, domain.ErrNegativeValue):
		return sharedErrors.NewBusinessError("INVALID_VALUE", "Contract value cannot be negative").
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation)

	case errors.Is(err, domain.ErrInvalidCurrency):
		return sharedErrors.NewBusinessError("INVALID_CURRENCY", "Currency code must be exactly 3 characters").
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation)

	case errors.Is(err, domain.ErrInvalidContractType):
		return sharedErrors.NewBusinessError("INVALID_CONTRACT_TYPE", "Invalid contract type").
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation).
			WithSuggestion("Valid types: VENDOR, CUSTOMER, EMPLOYEE, SERVICE, LEASE, OTHER")
	}

	return err
}

// handleError maps, logs, and returns a structured HTTP error response.
func (h *Handler) handleError(c *fiber.Ctx, err error) error {
	err = mapContractError(err)
	httpErr := sharedErrors.ToHTTPError(err)
	httpErr.RequestID = getRequestID(c)

	h.log.ErrorContext(c.UserContext(), "contract handler error", logger.Fields{
		"error":      err.Error(),
		"request_id": httpErr.RequestID,
		"method":     c.Method(),
		"path":       c.Path(),
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

// validateRequest parses and validates a JSON request body.
func (h *Handler) validateRequest(c *fiber.Ctx, req any) error {
	if err := c.BodyParser(req); err != nil {
		return sharedErrors.NewBusinessError("INVALID_JSON", "Invalid JSON format").
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation).
			WithSuggestion("Ensure the request body contains valid JSON")
	}
	if err := h.validator.Struct(req); err != nil {
		return sharedErrors.NewBusinessError("VALIDATION_ERROR", err.Error()).
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation)
	}
	return nil
}

// success returns a 200 OK JSON response.
func (h *Handler) success(c *fiber.Ctx, data any) error {
	h.recordSuccess(c)
	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
		"meta": fiber.Map{
			"request_id": getRequestID(c),
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// successWithMeta returns a 200 OK JSON response with extra metadata (e.g. pagination).
func (h *Handler) successWithMeta(c *fiber.Ctx, data any, meta fiber.Map) error {
	h.recordSuccess(c)
	meta["request_id"] = getRequestID(c)
	meta["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
		"meta":    meta,
	})
}

// created returns a 201 Created JSON response.
func (h *Handler) created(c *fiber.Ctx, data any) error {
	h.recordSuccess(c)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    data,
		"meta": fiber.Map{
			"request_id": getRequestID(c),
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// noContent returns 204 No Content for delete/action endpoints.
func (h *Handler) noContent(c *fiber.Ctx) error {
	h.recordSuccess(c)
	return c.SendStatus(fiber.StatusNoContent)
}

// ── private ───────────────────────────────────────────────────────────────────

func getRequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals("requestid").(string); ok {
		return id
	}
	return ""
}

func (h *Handler) recordSuccess(c *fiber.Ctx) {
	h.metrics.IncrementCounter("api_requests_total", metrics.Fields{
		"method":   c.Method(),
		"endpoint": c.Route().Path,
		"status":   "success",
	})
}
