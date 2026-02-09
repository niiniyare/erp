package tenant

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	coreTenant "github.com/niiniyare/erp/internal/core/tenant"
	"go.opentelemetry.io/otel/attribute"
)

// Create handles the creation of a new organization (tenant).
// @Summary Create Organization
// @Description Creates a new organization with the provided details.
// @Tags Organizations
// @Accept json
// @Produce json
// @Param organization body coreTenant.CreateTenantRequest true "Organization Creation Request"
// @Success 201 {object} map[string]interface{} "Organization created successfully"
// @Failure 400 {object} errors.HTTPError "Validation Error"
// @Failure 409 {object} errors.HTTPError "Conflict (e.g., subdomain taken)"
// @Failure 500 {object} errors.HTTPError "Internal Server Error"
// @Router /organizations [post]
func (h *TenantHandler) Create(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.handler.Create")
	defer span.End()

	var req coreTenant.CreateTenantRequest
	if err := h.validateRequest(c, &req); err != nil {
		return h.handleError(c, err)
	}
	span.SetAttributes(attribute.String("tenant.name.req", req.Name))

	createdTenant, err := h.service.CreateTenant(ctx, req)
	if err != nil {
		return h.handleError(c, err)
	}

	span.SetAttributes(
		attribute.String("tenant.id", createdTenant.ID.String()),
		attribute.String("tenant.name", createdTenant.Name),
	)

	c.Set("Location", fmt.Sprintf("/api/v1/organizations/%s", createdTenant.ID))
	return h.created(c, toResponse(createdTenant, "detailed"))
}

// Onboard handles a complex, nested creation request for a new organization.
// This is an example of the "Complex Nested Operations" pattern from the API design.
// It would typically trigger an asynchronous workflow.
func (h *TenantHandler) Onboard(c *fiber.Ctx) error {
	// For now, this is a placeholder. A real implementation would:
	// 1. Define a complex request struct in `requests.go`.
	// 2. Validate the complex request.
	// 3. Start a Temporal workflow (e.g., OnboardingWorkflow).
	// 4. Return a 202 Accepted response with the job ID, as shown in the API design doc.
	return h.accepted(c, fiber.Map{
		"status":  "onboarding_endpoint_not_implemented",
		"message": "This endpoint is a placeholder for a future complex onboarding workflow.",
	})
}
