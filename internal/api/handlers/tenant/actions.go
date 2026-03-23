package tenant

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
)

// Activate handles the activation of a tenant.
// @Summary Activate Organization
// @Description Activates a PENDING or SUSPENDED organization.
// @Tags Organizations
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} map[string]interface{} "Activation successful"
// @Failure 400 {object} errors.HTTPError "Invalid ID or invalid state transition"
// @Failure 404 {object} errors.HTTPError "Organization not found"
// @Router /organizations/{id}/activate [post]
func (h *TenantHandler) Activate(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.handler.Activate")
	defer span.End()

	tenantID := c.Params("id")
	h.logger.InfoContext(ctx, "Activating organization", logger.Fields{"organization_id": tenantID})

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return h.handleError(c, errors.NewBusinessError("INVALID_ORGANIZATION_ID", "Invalid organization ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	if err := h.service.ActivateTenant(ctx, tenantUUID); err != nil {
		return h.handleError(c, err)
	}

	updated, err := h.service.GetTenantByID(ctx, tenantUUID)
	if err != nil {
		return h.handleError(c, err)
	}

	return h.success(c, toActionResponse(updated, "activated"))
}

// Suspend handles the suspension of a tenant.
// @Summary Suspend Organization
// @Description Suspends an ACTIVE organization for a specified reason.
// @Tags Organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param reason body object true "Suspension Reason"
// @Success 200 {object} map[string]interface{} "Suspension successful"
// @Failure 400 {object} errors.HTTPError "Invalid ID or invalid state transition"
// @Failure 404 {object} errors.HTTPError "Organization not found"
// @Router /organizations/{id}/suspend [post]
func (h *TenantHandler) Suspend(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.handler.Suspend")
	defer span.End()

	tenantID := c.Params("id")
	h.logger.InfoContext(ctx, "Suspending organization", logger.Fields{"organization_id": tenantID})

	var body struct {
		Reason string `json:"reason" validate:"required,min=10"`
	}
	if err := h.validateRequest(c, &body); err != nil {
		return h.handleError(c, err)
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return h.handleError(c, errors.NewBusinessError("INVALID_ORGANIZATION_ID", "Invalid organization ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	if err := h.service.SuspendTenant(ctx, tenantUUID, body.Reason); err != nil {
		return h.handleError(c, err)
	}

	updated, err := h.service.GetTenantByID(ctx, tenantUUID)
	if err != nil {
		return h.handleError(c, err)
	}

	resp := toActionResponse(updated, "suspended")
	resp["reason"] = body.Reason
	return h.success(c, resp)
}

// Archive handles the archival of a tenant.
// @Summary Archive Organization
// @Description Archives an organization. This is a terminal state.
// @Tags Organizations
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} map[string]interface{} "Archival successful"
// @Failure 400 {object} errors.HTTPError "Invalid ID or invalid state transition"
// @Failure 404 {object} errors.HTTPError "Organization not found"
// @Router /organizations/{id}/archive [post]
func (h *TenantHandler) Archive(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.handler.Archive")
	defer span.End()

	tenantID := c.Params("id")
	h.logger.InfoContext(ctx, "Archiving organization", logger.Fields{"organization_id": tenantID})

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return h.handleError(c, errors.NewBusinessError("INVALID_ORGANIZATION_ID", "Invalid organization ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	if err := h.service.ArchiveTenant(ctx, tenantUUID); err != nil {
		return h.handleError(c, err)
	}

	return h.success(c, fiber.Map{
		"id":     tenantUUID.String(),
		"object": "tenant",
		"status": "archived",
		"links": fiber.Map{
			"self": fmt.Sprintf("/api/v1/tenants/%s", tenantUUID),
		},
	})
}
