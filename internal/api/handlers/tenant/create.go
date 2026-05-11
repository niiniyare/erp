package tenant

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	coreTenant "awo.so/internal/core/tenant"
	"awo.so/internal/core/tenant/domain"
	"go.opentelemetry.io/otel/attribute"
)

// Create handles the creation of a new organization (tenant).
// @Summary Create Organization
// @Description Creates a new organization with the provided details.
// @Tags Organizations
// @Accept json
// @Produce json
// @Param organization body coreTenant.CreateTenantRequest true "Organization Creation Request"
// @Success 201 {object} map[string]any "Organization created successfully"
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

	c.Set("Location", fmt.Sprintf("/api/v1/tenants/%s", createdTenant.ID))
	return h.created(c, toResponse(createdTenant, "detailed"))
}

// onboardRequest is the payload for the onboarding endpoint.
type onboardRequest struct {
	Name         string  `json:"name"          validate:"required,min=2,max=255"`
	Email        string  `json:"email"         validate:"required,email"`
	CountryCode  string  `json:"country_code"  validate:"required,len=2"`
	CurrencyCode string  `json:"currency_code" validate:"required,len=3"`
	Subdomain    *string `json:"subdomain,omitempty"`
	Industry     *string `json:"industry,omitempty"`
	CompanySize  *string `json:"company_size,omitempty"`
}

// Onboard enqueues a TenantProvisioningWorkflow and returns 202 Accepted with
// the workflow and run IDs. The caller can poll GET /api/v1/tenants/:id for
// status once provisioning completes and the tenant becomes ACTIVE.
func (h *TenantHandler) Onboard(c *fiber.Ctx) error {
	if h.onboardStarter == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "onboarding workflow not available — Temporal is not configured",
		})
	}

	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.handler.Onboard")
	defer span.End()

	var req onboardRequest
	if err := h.validateRequest(c, &req); err != nil {
		return h.handleError(c, err)
	}

	input := domain.ProvisioningInput{
		Name:         req.Name,
		Email:        req.Email,
		CountryCode:  req.CountryCode,
		CurrencyCode: req.CurrencyCode,
		Subdomain:    req.Subdomain,
		Industry:     req.Industry,
		CompanySize:  req.CompanySize,
	}

	result, err := h.onboardStarter.StartOnboarding(ctx, input)
	if err != nil {
		span.RecordError(err)
		return h.handleError(c, err)
	}

	return h.accepted(c, fiber.Map{
		"workflow_id": result.WorkflowID,
		"run_id":      result.RunID,
		"status":      "provisioning",
		"message":     "Tenant provisioning started. Poll GET /api/v1/tenants once the workflow completes.",
	})
}
