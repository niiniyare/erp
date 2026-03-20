package tenant

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	coreTenant "awo/internal/core/tenant"
	"awo/internal/shared/errors"
	"awo/internal/shared/logger"
)

// Update handles partial or full updates of an organization's details.
// It supports both PUT (full replacement) and PATCH (partial update).
// @Summary Update Organization
// @Description Updates an organization's details by its UUID.
// @Tags Organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param organization body coreTenant.UpdateTenantRequest true "Organization Update Request"
// @Success 200 {object} map[string]interface{} "Organization updated successfully"
// @Failure 400 {object} errors.HTTPError "Invalid request body or ID"
// @Failure 404 {object} errors.HTTPError "Organization not found"
// @Failure 409 {object} errors.HTTPError "Conflict (e.g., subdomain taken)"
// @Failure 500 {object} errors.HTTPError "Internal Server Error"
// @Router /organizations/{id} [put]
// @Router /organizations/{id} [patch]
func (h *TenantHandler) Update(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.handler.Update")
	defer span.End()

	tenantID := c.Params("id")
	h.logger.InfoContext(ctx, "Updating organization", logger.Fields{"organization_id": tenantID})

	var req coreTenant.UpdateTenantRequest
	// Note: For a true PATCH with field masks, more complex logic would be needed here.
	// For now, we treat PUT and PATCH similarly, relying on the request struct's `omitempty`.
	if err := h.validateRequest(c, &req); err != nil {
		return h.handleError(c, err)
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return h.handleError(c, errors.NewBusinessError("INVALID_ORGANIZATION_ID", "Invalid organization ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	updatedTenant, err := h.service.UpdateTenant(ctx, tenantUUID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return h.success(c, toResponse(updatedTenant, "detailed"))
}
