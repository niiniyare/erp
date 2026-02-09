package tenant

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// Delete handles the soft deletion of an organization.
// @Summary Delete Organization
// @Description Soft-deletes an organization by its UUID. This is a non-recoverable action from the API.
// @Tags Organizations
// @Produce json
// @Param id path string true "Organization ID"
// @Success 204 "No Content"
// @Failure 400 {object} errors.HTTPError "Invalid ID format"
// @Failure 404 {object} errors.HTTPError "Organization not found"
// @Failure 500 {object} errors.HTTPError "Internal Server Error"
// @Router /organizations/{id} [delete]
func (h *TenantHandler) Delete(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.handler.Delete")
	defer span.End()

	tenantID := c.Params("id")
	h.logger.InfoContext(ctx, "Deleting organization", logger.Fields{"organization_id": tenantID})

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return h.handleError(c, errors.NewBusinessError("INVALID_ORGANIZATION_ID", "Invalid organization ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	if err := h.service.DeleteTenant(ctx, tenantUUID); err != nil {
		return h.handleError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
