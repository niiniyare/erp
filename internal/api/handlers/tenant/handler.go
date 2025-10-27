package tenant

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	genTenant "github.com/niiniyare/erp/internal/api/gen/tenant"
	"github.com/niiniyare/erp/internal/api/handlers/common"
	coreTenant "github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/web/components/tenant"
)

// TenantHandler handles tenant-related HTTP requests following the 5-step handler pattern
type TenantHandler struct {
	helper  *common.HandlerHelper
	service coreTenant.Service
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewTenantHandler creates a new tenant handler with dependencies
func NewTenantHandler(tenantService coreTenant.Service, logger logger.Logger, metrics metrics.MetricsProvider, tracer tracing.TracingService) *TenantHandler {
	return &TenantHandler{
		helper:  common.NewHandlerHelper(logger, metrics, tracer),
		service: tenantService,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// ============================================================================
// CRUD OPERATIONS FOLLOWING 5-STEP PATTERN
// ============================================================================

// Create handles tenant creation (POST /api/v1/tenants)
func (h *TenantHandler) Create(c *fiber.Ctx) error {
	// Step 1: Observability
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.Create")
	defer span.End()

	h.logger.InfoContext(ctx, "Creating new tenant", logger.Fields{
		"method": c.Method(),
		"path":   c.Path(),
	})

	// Step 2: Parse and Validate
	var req coreTenant.CreateTenantRequest
	if err := c.BodyParser(&req); err != nil {
		return h.helper.HandleError(c, errors.NewBusinessError("INVALID_JSON", "Invalid JSON payload").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation).
			WithSeverity(errors.SeverityError).
			WithSuggestion("Ensure request body contains valid JSON"))
	}

	// Step 3: Extract Context (tenant validation, user context, etc.)
	// TODO: Extract tenant context from middleware
	// TODO: Validate user permissions

	// Step 4: Delegate to Service
	createdTenant, err := h.service.CreateTenant(ctx, req)
	if err != nil {
		return h.helper.HandleError(c, err)
	}

	// Convert tenant to API response
	response := h.tenantToAPIResponse(createdTenant)

	// Step 5: Return Response with content negotiation
	// For HTML responses, render tenant card component
	component := tenant.TenantCard(tenant.TenantCardProps{
		Tenant:   response,
		ShowEdit: true,
	})
	
	return h.helper.RespondWithComponent(c, fiber.StatusCreated, response, component)
}

// Get handles tenant retrieval by ID (GET /api/v1/tenants/:id)
func (h *TenantHandler) Get(c *fiber.Ctx) error {
	// Step 1: Observability
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.Get")
	defer span.End()

	tenantID := c.Params("id")
	h.logger.InfoContext(ctx, "Retrieving tenant", logger.Fields{
		"tenant_id": tenantID,
		"method":    c.Method(),
		"path":      c.Path(),
	})

	// Step 2: Parse and Validate
	if tenantID == "" {
		return h.helper.HandleError(c, errors.NewBusinessError("MISSING_TENANT_ID", "Tenant ID is required").
			WithCategory(errors.CategoryValidation).
			WithSeverity(errors.SeverityError))
	}

	// Step 3: Extract Context
	view := c.Query("view", "default")

	// Step 4: Delegate to Service
	// Parse tenant ID to UUID
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return h.helper.HandleError(c, errors.NewBusinessError("INVALID_TENANT_ID", "Invalid tenant ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	tenantEntity, err := h.service.GetTenantByID(ctx, tenantUUID)
	if err != nil {
		return h.helper.HandleError(c, err)
	}

	// Convert tenant to API response
	response := h.tenantToAPIResponse(tenantEntity)

	// Add view-specific fields
	if view == "detailed" {
		// For detailed view, include additional metadata
		if tenantEntity.Metadata != nil {
			response.Branding = tenantEntity.Metadata
		}
		if tenantEntity.Settings != nil {
			response.ContactInfo = tenantEntity.Settings
		}
	}

	// Step 5: Return Response with content negotiation
	h.metrics.IncrementCounter("tenant_retrieved_total", metrics.Fields{
		"status": "success",
		"view":   view,
	})

	// For HTML responses, render tenant detail component
	component := tenant.TenantDetail(response)
	
	return h.helper.RespondWithComponent(c, fiber.StatusOK, response, component)
}

// List handles tenant listing with pagination (GET /api/v1/tenants)
func (h *TenantHandler) List(c *fiber.Ctx) error {
	// Step 1: Observability
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.List")
	defer span.End()

	h.logger.InfoContext(ctx, "Listing tenants", logger.Fields{
		"method": c.Method(),
		"path":   c.Path(),
	})

	// Step 2: Parse and Validate query parameters
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)

	// Validate page size
	if pageSize > 100 {
		return h.helper.HandleError(c, errors.NewBusinessError("INVALID_PAGE_SIZE", "Page size cannot exceed 100").
			WithCategory(errors.CategoryValidation).
			WithSeverity(errors.SeverityError).
			WithDetail("page_size", pageSize).
			WithSuggestion("Use page_size between 1 and 100"))
	}

	// Calculate offset for the existing service
	offset := (page - 1) * pageSize

	// Step 3: Extract Context
	// TODO: Extract tenant context and user permissions

	// Step 4: Delegate to Service
	tenants, err := h.service.ListTenants(ctx, offset, pageSize)
	if err != nil {
		return h.helper.HandleError(c, err)
	}

	// Convert tenants to API response
	apiTenants := make([]*genTenant.Tenant, len(tenants))
	for i, tenantEntity := range tenants {
		apiTenants[i] = h.tenantToAPIResponse(tenantEntity)
	}

	// Calculate total pages (simplified - in real implementation would need total count)
	totalPages := 1
	if len(tenants) == pageSize {
		totalPages = page + 1 // Simple estimation
	}

	// Create pagination response
	paginationResponse := &tenant.PaginationInfo{
		Page:       page,
		PerPage:    pageSize,
		Total:      len(tenants),
		TotalPages: totalPages,
	}

	response := map[string]interface{}{
		"data": apiTenants,
		"pagination": map[string]interface{}{
			"page":        page,
			"page_size":   pageSize,
			"total_items": len(tenants),
			"total_pages": totalPages,
		},
	}

	// Step 5: Return Response with content negotiation
	h.metrics.IncrementCounter("tenant_listed_total", metrics.Fields{
		"status": "success",
		"count":  len(apiTenants),
	})

	// For HTML responses, render tenant list component
	component := tenant.TenantList(tenant.TenantListProps{
		Tenants:    apiTenants,
		Pagination: paginationResponse,
	})
	
	return h.helper.RespondWithComponent(c, fiber.StatusOK, response, component)
}


// Update handles tenant updates (PUT /api/v1/tenants/:id)
func (h *TenantHandler) Update(c *fiber.Ctx) error {
	// Step 1: Observability
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.Update")
	defer span.End()

	tenantID := c.Params("id")
	h.logger.InfoContext(ctx, "Updating tenant", logger.Fields{
		"tenant_id": tenantID,
		"method":    c.Method(),
		"path":      c.Path(),
	})

	// Step 2: Parse and Validate
	var req coreTenant.UpdateTenantRequest
	if err := c.BodyParser(&req); err != nil {
		return h.helper.HandleError(c, errors.NewBusinessError("INVALID_JSON", "Invalid JSON payload").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation).
			WithSeverity(errors.SeverityError))
	}

	// Parse tenant ID to UUID
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return h.helper.HandleError(c, errors.NewBusinessError("INVALID_TENANT_ID", "Invalid tenant ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	// Step 3: Extract Context
	// TODO: Extract tenant context and validate permissions

	// Step 4: Delegate to Service
	updatedTenant, err := h.service.UpdateTenant(ctx, tenantUUID, req)
	if err != nil {
		return h.helper.HandleError(c, err)
	}

	// Convert tenant to API response
	response := h.tenantToAPIResponse(updatedTenant)

	// Step 5: Return Response
	return h.helper.Respond(c, fiber.StatusOK, response)
}

// Delete handles tenant deletion (DELETE /api/v1/tenants/:id)
func (h *TenantHandler) Delete(c *fiber.Ctx) error {
	// Step 1: Observability
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.Delete")
	defer span.End()

	tenantID := c.Params("id")
	h.logger.InfoContext(ctx, "Deleting tenant", logger.Fields{
		"tenant_id": tenantID,
		"method":    c.Method(),
		"path":      c.Path(),
	})

	// Step 2: Parse and Validate
	// Parse tenant ID to UUID
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return h.helper.HandleError(c, errors.NewBusinessError("INVALID_TENANT_ID", "Invalid tenant ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	// Step 3: Extract Context
	// TODO: Extract tenant context and validate permissions

	// Step 4: Delegate to Service
	if err := h.service.DeleteTenant(ctx, tenantUUID); err != nil {
		return h.helper.HandleError(c, err)
	}

	// Step 5: Return Response (204 No Content)
	return c.SendStatus(fiber.StatusNoContent)
}

// ============================================================================
// UI FORM HANDLERS
// ============================================================================

// NewForm handles GET requests for the new tenant form
func (h *TenantHandler) NewForm(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.NewForm")
	defer span.End()

	h.logger.InfoContext(ctx, "Serving new tenant form", logger.Fields{
		"method": c.Method(),
		"path":   c.Path(),
	})

	// Render the tenant form component for creation
	component := tenant.TenantForm(tenant.TenantFormProps{
		Action:    "create",
		ActionURL: "/api/v1/tenants",
	})

	return h.helper.RenderTemplComponent(c, fiber.StatusOK, component)
}

// EditForm handles GET requests for the edit tenant form
func (h *TenantHandler) EditForm(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.EditForm")
	defer span.End()

	tenantID := c.Params("id")
	h.logger.InfoContext(ctx, "Serving edit tenant form", logger.Fields{
		"tenant_id": tenantID,
		"method":    c.Method(),
		"path":      c.Path(),
	})

	// Parse tenant ID
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return h.helper.HandleError(c, errors.NewBusinessError("INVALID_TENANT_ID", "Invalid tenant ID format").
			WithCategory(errors.CategoryValidation).
			WithSeverity(errors.SeverityError))
	}

	// Get the tenant
	tenantEntity, err := h.service.GetTenantByID(ctx, tenantUUID)
	if err != nil {
		return h.helper.HandleError(c, err)
	}

	// Convert to API response
	tenantResponse := h.tenantToAPIResponse(tenantEntity)

	// Render the tenant form component for editing
	component := tenant.TenantForm(tenant.TenantFormProps{
		Tenant:    tenantResponse,
		Action:    "edit",
		ActionURL: "/api/v1/tenants/" + tenantID,
	})

	return h.helper.RenderTemplComponent(c, fiber.StatusOK, component)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// tenantToAPIResponse converts a tenant entity to API tenant response
func (h *TenantHandler) tenantToAPIResponse(tenantEntity *coreTenant.Tenant) *genTenant.Tenant {
	if tenantEntity == nil {
		return nil
	}

	apiTenant := &genTenant.Tenant{
		ID:           tenantEntity.ID.String(),
		Slug:         tenantEntity.Slug,
		Name:         tenantEntity.Name,
		Email:        tenantEntity.Email,
		Status:       string(tenantEntity.Status),
		Timezone:     tenantEntity.Timezone,
		CurrencyCode: tenantEntity.CurrencyCode,
		CreatedAt:    tenantEntity.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Optional fields
	if tenantEntity.Industry != nil {
		apiTenant.Industry = tenantEntity.Industry
	}
	if tenantEntity.CompanySize != nil {
		apiTenant.CompanySize = tenantEntity.CompanySize
	}
	if tenantEntity.Subdomain != nil {
		// For API compatibility, we can map subdomain to a custom field if needed
	}
	updatedAtStr := tenantEntity.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	apiTenant.UpdatedAt = &updatedAtStr
	if tenantEntity.Metadata != nil {
		apiTenant.Branding = tenantEntity.Metadata
	}
	if tenantEntity.Settings != nil {
		apiTenant.ContactInfo = tenantEntity.Settings
	}

	return apiTenant
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}

// setupTenantRoutes sets up tenant routes (used by tests)
func setupTenantRoutes(app *fiber.App, handler *TenantHandler) {
	api := app.Group("/api/v1")
	tenants := api.Group("/tenants")

	// CRUD operations
	tenants.Post("/", handler.Create)
	tenants.Get("/", handler.List)
	tenants.Get("/:id", handler.Get)
	tenants.Put("/:id", handler.Update)
	tenants.Delete("/:id", handler.Delete)
	
	// UI form routes
	tenants.Get("/new", handler.NewForm)      // GET /api/v1/tenants/new - New tenant form
	tenants.Get("/:id/edit", handler.EditForm) // GET /api/v1/tenants/:id/edit - Edit tenant form
}
