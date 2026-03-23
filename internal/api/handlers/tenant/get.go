package tenant

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// Get handles retrieval of a single organization by its ID.
// @Summary Get Organization
// @Description Retrieves an organization's details by its UUID. Supports different views.
// @Tags Organizations
// @Produce json
// @Param id path string true "Organization ID"
// @Param view query string false "Response view (summary, default, detailed)" Enums(summary, default, detailed) default(default)
// @Success 200 {object} map[string]interface{} "Successful response"
// @Failure 400 {object} errors.HTTPError "Invalid ID format"
// @Failure 404 {object} errors.HTTPError "Organization not found"
// @Failure 500 {object} errors.HTTPError "Internal Server Error"
// @Router /organizations/{id} [get]
func (h *TenantHandler) Get(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.handler.Get")
	defer span.End()

	tenantID := c.Params("id")
	view := c.Query("view", "default")

	h.logger.InfoContext(ctx, "Retrieving organization", logger.Fields{
		"organization_id": tenantID,
		"view":            view,
	})

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return h.handleError(c, errors.NewBusinessError("INVALID_ORGANIZATION_ID", "Invalid organization ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	tenantEntity, err := h.service.GetTenantByID(ctx, tenantUUID)
	if err != nil {
		return h.handleError(c, err)
	}

	response := toResponse(tenantEntity, view)

	h.metrics.IncrementCounter("organization_retrieved_total", metrics.Fields{
		"status": "success",
		"view":   view,
	})

	return h.success(c, response)
}
