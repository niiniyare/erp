package tenant

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	coreTenant "awo/internal/core/tenant"
	"awo/internal/core/tenant/domain"
	"awo/internal/shared/errors"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	stderrors "errors"
)

// mapTenantError converts domain-layer errors into structured *BusinessError values
// so that ToHTTPError can assign the correct HTTP status instead of defaulting to 500.
func mapTenantError(err error) error {
	switch {
	case stderrors.Is(err, domain.ErrTenantNotFound), stderrors.Is(err, domain.ErrConfigurationNotFound):
		return errors.NewBusinessError("TENANT_NOT_FOUND", "Tenant not found").
			WithHTTPStatus(http.StatusNotFound).
			WithCategory(errors.CategoryTenant).
			WithSuggestion("Verify the tenant ID is correct")

	case stderrors.Is(err, domain.ErrSubdomainTaken):
		return errors.NewBusinessError("SUBDOMAIN_TAKEN", "Subdomain is already in use").
			WithHTTPStatus(http.StatusConflict).
			WithCategory(errors.CategoryTenant).
			WithSuggestion("Choose a different subdomain")

	case stderrors.Is(err, domain.ErrTenantAlreadyExists):
		return errors.NewBusinessError("TENANT_EXISTS", "A tenant with that name or slug already exists").
			WithHTTPStatus(http.StatusConflict).
			WithCategory(errors.CategoryTenant).
			WithSuggestion("Use a different name or slug")

	case stderrors.Is(err, domain.ErrTenantSuspended):
		return errors.NewBusinessError("TENANT_SUSPENDED", "Tenant account is suspended").
			WithHTTPStatus(http.StatusForbidden).
			WithCategory(errors.CategoryTenant).
			WithSuggestion("Contact support to reactivate the account")

	case stderrors.Is(err, domain.ErrAlreadyActive):
		return errors.NewBusinessError("ALREADY_ACTIVE", "Tenant is already active").
			WithHTTPStatus(http.StatusConflict).
			WithCategory(errors.CategoryBusiness)

	case stderrors.Is(err, domain.ErrAlreadySuspended):
		return errors.NewBusinessError("ALREADY_SUSPENDED", "Tenant is already suspended").
			WithHTTPStatus(http.StatusConflict).
			WithCategory(errors.CategoryBusiness)

	case stderrors.Is(err, domain.ErrAlreadyArchived):
		return errors.NewBusinessError("ALREADY_ARCHIVED", "Tenant is already archived").
			WithHTTPStatus(http.StatusConflict).
			WithCategory(errors.CategoryBusiness)

	case stderrors.Is(err, domain.ErrCannotActivateArchivedTenant):
		return errors.NewBusinessError("CANNOT_ACTIVATE_ARCHIVED", "Archived tenants cannot be activated").
			WithHTTPStatus(http.StatusUnprocessableEntity).
			WithCategory(errors.CategoryBusiness).
			WithSuggestion("Create a new tenant instead")

	case stderrors.Is(err, domain.ErrCannotSuspendArchivedTenant):
		return errors.NewBusinessError("CANNOT_SUSPEND_ARCHIVED", "Archived tenants cannot be suspended").
			WithHTTPStatus(http.StatusUnprocessableEntity).
			WithCategory(errors.CategoryBusiness)

	case stderrors.Is(err, domain.ErrInvalidTransition):
		return errors.NewBusinessError("INVALID_STATUS_TRANSITION", "This status transition is not allowed").
			WithHTTPStatus(http.StatusUnprocessableEntity).
			WithCategory(errors.CategoryBusiness)

	case stderrors.Is(err, domain.ErrInvalidCompanySize):
		return errors.NewBusinessError("INVALID_COMPANY_SIZE", "Invalid company size").
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(errors.CategoryValidation).
			WithSuggestion("Valid values: STARTUP, SMALL, MEDIUM, LARGE, ENTERPRISE")

	case stderrors.Is(err, domain.ErrInvalidRequest),
		stderrors.Is(err, domain.ErrInvalidEmail),
		stderrors.Is(err, domain.ErrInvalidSubdomain),
		stderrors.Is(err, domain.ErrTenantNameRequired),
		stderrors.Is(err, domain.ErrTenantEmailRequired):
		return errors.NewBusinessError("VALIDATION_ERROR", err.Error()).
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(errors.CategoryValidation)
	}
	return err
}

// handleError provides centralized error handling for the tenant module.
// It logs the error, records metrics, and formats a structured error response.
func (h *TenantHandler) handleError(c *fiber.Ctx, err error) error {
	err = mapTenantError(err)

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
