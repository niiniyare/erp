package tenant

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"awo.so/internal/core/entity"
	"awo.so/internal/core/iam/contract"
	iamDomain "awo.so/internal/core/iam/domain"
	coreTenant "awo.so/internal/core/tenant"
	"awo.so/internal/core/tenant/domain"
	"awo.so/internal/shared"
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

// syncOnboardRequest is the payload for the synchronous onboarding endpoint.
type syncOnboardRequest struct {
	Name          string  `json:"name"           validate:"required,min=2,max=255"`
	Email         string  `json:"email"          validate:"required,email"`
	CountryCode   string  `json:"country_code"   validate:"required,len=2"`
	CurrencyCode  string  `json:"currency_code"  validate:"required,len=3"`
	AdminPassword string  `json:"admin_password" validate:"required,min=8"`
	AdminUsername string  `json:"admin_username" validate:"omitempty,min=3,max=50"`
	Subdomain     *string `json:"subdomain,omitempty"`
	Industry      *string `json:"industry,omitempty"`
	CompanySize   *string `json:"company_size,omitempty"`
}

// OnboardSync provisions a tenant and its first admin user in a single
// synchronous request. No Temporal required.
//
// POST /api/v1/tenants/onboard/sync
//
// Flow:
//  1. Create tenant (status = PENDING).
//  2. Activate tenant → ACTIVE.
//  3. Seed default IAM roles for the new tenant.
//  4. Create root COMPANY entity for the tenant.
//  5. Create the first admin user scoped to the tenant + root entity.
//  6. Assign role:tenant.admin to the new user.
//
// Returns 201 with tenant_id, user_id, and email on success.
func (h *TenantHandler) OnboardSync(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.handler.OnboardSync")
	defer span.End()

	var req syncOnboardRequest
	if err := h.validateRequest(c, &req); err != nil {
		return h.handleError(c, err)
	}

	// ── Step 1: Create tenant ────────────────────────────────────────────────
	username := strings.ToLower(strings.ReplaceAll(req.AdminUsername, " ", "_"))
	if username == "" {
		// derive from email local part
		parts := strings.SplitN(req.Email, "@", 2)
		username = strings.ToLower(parts[0])
	}

	createReq := coreTenant.CreateTenantRequest{
		Name:         req.Name,
		Email:        req.Email,
		CountryCode:  req.CountryCode,
		CurrencyCode: req.CurrencyCode,
		Subdomain:    req.Subdomain,
		Industry:     req.Industry,
		CompanySize:  req.CompanySize,
	}

	t, err := h.service.CreateTenant(ctx, createReq)
	if err != nil {
		span.RecordError(err)
		return h.handleError(c, err)
	}
	span.SetAttributes(attribute.String("tenant.id", t.ID.String()))

	// ── Step 2: Activate tenant ──────────────────────────────────────────────
	if err := h.service.ActivateTenant(ctx, t.ID); err != nil {
		span.RecordError(err)
		return h.handleError(c, err)
	}

	// Inject tenant context so subsequent service calls use the correct RLS.
	ctx = shared.WithTenantID(ctx, t.ID)
	c.SetUserContext(ctx)

	// ── Step 3: Seed default IAM roles ───────────────────────────────────────
	if h.authzSvc != nil {
		if err := contract.SeedDefaultRoles(ctx, h.authzSvc, t.ID.String()); err != nil {
			span.RecordError(err)
			h.logger.Error("OnboardSync: seed roles failed")
			return h.handleError(c, err)
		}
	}

	// ── Step 4: Create root entity ───────────────────────────────────────────
	if h.userSvc == nil || h.entitySvc == nil {
		return h.created(c, fiber.Map{
			"tenant_id": t.ID,
			"status":    "active",
			"message":   "Tenant created. User/entity service not configured — create admin manually.",
		})
	}

	rootEntity, err := h.entitySvc.CreateEntity(ctx, entity.CreateEntityRequest{
		Name:     req.Name,
		Code:     toEntityCode(req.Name),
		Type:     entity.EntityTypeCompany,
		IsActive: true,
	})
	if err != nil {
		span.RecordError(err)
		return h.handleError(c, err)
	}
	span.SetAttributes(attribute.String("entity.id", rootEntity.ID.String()))

	// ── Step 5: Create admin user ────────────────────────────────────────────
	displayName := req.Name + " Admin"
	userReq := &iamDomain.CreateUserRequest{
		EntityID:      rootEntity.ID,
		Username:      username,
		Email:         req.Email,
		DisplayName:   &displayName,
		Password:      req.AdminPassword,
		UserType:      "INTERNAL",
		AccountStatus: string(iamDomain.AccountStatusActive),
	}

	adminUser, err := h.userSvc.RegisterNewUser(ctx, userReq)
	if err != nil {
		span.RecordError(err)
		return h.handleError(c, err)
	}
	span.SetAttributes(attribute.String("user.id", adminUser.ID.String()))

	// ── Step 6: Assign tenant.admin role ─────────────────────────────────────
	if h.authzSvc != nil {
		if err := contract.AssignAdminRole(ctx, h.authzSvc, t.ID.String(), adminUser.ID.String()); err != nil {
			span.RecordError(err)
			h.logger.Error("OnboardSync: assign admin role failed")
			return h.handleError(c, err)
		}
	}

	return h.created(c, fiber.Map{
		"tenant_id": t.ID,
		"entity_id": rootEntity.ID,
		"user_id":   adminUser.ID,
		"email":     adminUser.Email,
		"username":  adminUser.Username,
		"status":    "active",
	})
}

// toEntityCode converts a company name to a short uppercase entity code.
// "Acme Corp" → "ACMECORP" (letters only, max 10 chars, uppercase).
func toEntityCode(name string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
		if b.Len() >= 10 {
			break
		}
	}
	if b.Len() == 0 {
		return "ROOT"
	}
	return b.String()
}
