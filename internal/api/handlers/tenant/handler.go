package tenant

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	genTenant "github.com/niiniyare/erp/internal/api/gen/tenant"
	coreTenant "github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared/encryption"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// TenantHandler handles tenant-related HTTP requests with centralized error handling
type TenantHandler struct {
	service    coreTenant.Service
	logger     logger.Logger
	metrics    metrics.MetricsProvider
	tracer     tracing.Service
	validator  *validator.Validate
	encryption encryption.EncryptionService
}

// NewTenantHandler creates a new tenant handler with dependencies
func NewTenantHandler(
	tenantService coreTenant.Service,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
) *TenantHandler {
	validator := validator.New()
	
	// Register custom validators
	registerCustomValidators(validator)
	
	return &TenantHandler{
		service:    tenantService,
		logger:     logger,
		metrics:    metrics,
		tracer:     tracer,
		validator:  validator,
		encryption: nil, // Will be added later
	}
}

// ============================================================================
// CRUD OPERATIONS FOLLOWING 5-STEP PATTERN
// ============================================================================

// Create handles tenant creation (POST /api/v1/tenants)
// @Summary Create tenant
// @Description Creates a new tenant in the system
// @Tags tenants
// @Accept json
// @Produce json
// @Param tenant body coreTenant.CreateTenantRequest true "Tenant data"
// @Success 201 {object} genTenant.Tenant
// @Router /api/v1/tenants [post]
func (h *TenantHandler) Create(c *fiber.Ctx) error {
	// Step 1: Start tracing
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.Create")
	defer span.End()
	c.SetUserContext(ctx)

	// Step 2: Parse and validate request
	var req coreTenant.CreateTenantRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return h.HandleError(c, err)
	}

	// Step 3: Delegate to service
	createdTenant, err := h.service.CreateTenant(c.Context(), req)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}

	// Step 4: Convert to API response
	response := h.tenantToAPIResponse(createdTenant)

	// Step 5: Return successful response
	span.SetAttributes(
		attribute.String("tenant.id", createdTenant.ID.String()),
		attribute.String("tenant.name", createdTenant.Name),
	)

	return h.Created(c, response)
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
		return h.HandleError(c, errors.NewBusinessError("MISSING_TENANT_ID", "Tenant ID is required").
			WithCategory(errors.CategoryValidation).
			WithSeverity(errors.SeverityError))
	}

	// Step 3: Extract Context
	view := c.Query("view", "default")

	// Step 4: Delegate to Service
	// Parse tenant ID to UUID
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_TENANT_ID", "Invalid tenant ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	tenantEntity, err := h.service.GetTenantByID(ctx, tenantUUID)
	if err != nil {
		return h.HandleError(c, err)
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

	// Step 5: Return Response
	h.metrics.IncrementCounter("tenant_retrieved_total", metrics.Fields{
		"status": "success",
		"view":   view,
	})

	return h.Success(c, response)
}

// List handles tenant listing with pagination (GET /api/v1/tenants)
// @Summary List tenants
// @Description List tenants with pagination
// @Tags tenants
// @Accept json
// @Produce json
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/tenants [get]
func (h *TenantHandler) List(c *fiber.Ctx) error {
	// Step 1: Start tracing
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.List")
	defer span.End()
	c.SetUserContext(ctx)

	// Step 2: Extract pagination parameters
	offset, limit := h.ExtractPaginationParams(c)

	// Step 3: Delegate to service
	tenants, err := h.service.ListTenants(c.Context(), offset, limit)
	if err != nil {
		span.RecordError(err)
		return h.HandleError(c, err)
	}

	// Step 4: Convert to API response
	apiTenants := make([]*genTenant.Tenant, len(tenants))
	for i, tenantEntity := range tenants {
		apiTenants[i] = h.tenantToAPIResponse(tenantEntity)
	}

	// Step 5: Return response with metadata
	meta := map[string]interface{}{
		"pagination": map[string]interface{}{
			"offset": offset,
			"limit":  limit,
			"count":  len(apiTenants),
		},
	}

	span.SetAttributes(
		attribute.Int("pagination.offset", offset),
		attribute.Int("pagination.limit", limit),
		attribute.Int("results.count", len(apiTenants)),
	)

	return h.SuccessWithMeta(c, apiTenants, meta)
}

// ============================================================================
// BASE HANDLER METHODS (from BaseHandler pattern)
// ============================================================================

// HandleError provides centralized error handling
func (h *TenantHandler) HandleError(c *fiber.Ctx, err error) error {
	requestID := h.getRequestID(c)
	tenantID := h.getTenantID(c)
	
	// Convert to HTTP error using existing system
	httpErr := errors.ToHTTPError(err)
	httpErr.RequestID = requestID
	
	// Log the error with context
	h.logger.Error("Request error", logger.Fields{
		"error":      err.Error(),
		"request_id": requestID,
		"tenant_id":  tenantID,
		"method":     c.Method(),
		"path":       c.Path(),
		"ip":         c.IP(),
		"status":     httpErr.Status,
		"code":       httpErr.Code,
	})
	
	// Record metrics
	h.recordErrorMetrics(c, httpErr)
	
	return c.Status(httpErr.Status).JSON(httpErr)
}

// ValidateRequest validates request data
func (h *TenantHandler) ValidateRequest(c *fiber.Ctx, req interface{}) error {
	if err := c.BodyParser(req); err != nil {
		return errors.NewBusinessError("INVALID_JSON", "Invalid JSON format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation).
			WithSuggestion("Ensure request body contains valid JSON")
	}
	
	if err := h.validator.Struct(req); err != nil {
		return err
	}
	
	return nil
}

// Success returns a standardized success response
func (h *TenantHandler) Success(c *fiber.Ctx, data interface{}) error {
	response := fiber.Map{
		"success":    true,
		"data":       data,
		"request_id": h.getRequestID(c),
		"timestamp":  time.Now(),
	}
	
	h.recordSuccessMetrics(c)
	return c.JSON(response)
}

// SuccessWithMeta returns success response with metadata
func (h *TenantHandler) SuccessWithMeta(c *fiber.Ctx, data interface{}, meta map[string]interface{}) error {
	response := fiber.Map{
		"success":    true,
		"data":       data,
		"meta":       meta,
		"request_id": h.getRequestID(c),
		"timestamp":  time.Now(),
	}
	
	h.recordSuccessMetrics(c)
	return c.JSON(response)
}

// Created returns a 201 response
func (h *TenantHandler) Created(c *fiber.Ctx, data interface{}) error {
	response := fiber.Map{
		"success":    true,
		"data":       data,
		"request_id": h.getRequestID(c),
		"timestamp":  time.Now(),
	}
	
	h.recordSuccessMetrics(c)
	return c.Status(fiber.StatusCreated).JSON(response)
}

// StartTracing begins a new trace span
func (h *TenantHandler) StartTracing(c *fiber.Ctx, operationName string) (tracing.Span, func()) {
	ctx := c.Context()
	newCtx, span := h.tracer.StartSpan(ctx, operationName)
	
	// Update the context in fiber
	c.SetUserContext(newCtx)
	
	span.SetAttributes(
		attribute.String("http.method", c.Method()),
		attribute.String("http.path", c.Path()),
		attribute.String("http.user_agent", c.Get("User-Agent")),
		attribute.String("tenant.id", h.getTenantID(c)),
		attribute.String("request.id", h.getRequestID(c)),
	)
	
	cleanup := func() {
		span.End()
	}
	
	return span, cleanup
}

// ExtractPaginationParams extracts pagination parameters
func (h *TenantHandler) ExtractPaginationParams(c *fiber.Ctx) (offset, limit int) {
	offset = c.QueryInt("offset", 0)
	limit = c.QueryInt("limit", 20)
	
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	
	return offset, limit
}

// Helper methods
func (h *TenantHandler) getRequestID(c *fiber.Ctx) string {
	if requestID := c.Locals("requestid"); requestID != nil {
		return requestID.(string)
	}
	return "unknown"
}

func (h *TenantHandler) getTenantID(c *fiber.Ctx) string {
	if tenantID := c.Locals("tenant_id"); tenantID != nil {
		return tenantID.(string)
	}
	return ""
}

func (h *TenantHandler) recordErrorMetrics(c *fiber.Ctx, httpErr *errors.HTTPError) {
	h.metrics.IncrementCounter("api_errors_total", metrics.Fields{
		"method":   c.Method(),
		"endpoint": c.Route().Path,
		"code":     httpErr.Code,
		"status":   string(rune(httpErr.Status)),
	})
}

func (h *TenantHandler) recordSuccessMetrics(c *fiber.Ctx) {
	h.metrics.IncrementCounter("api_requests_total", metrics.Fields{
		"method":   c.Method(),
		"endpoint": c.Route().Path,
		"status":   "success",
	})
}

// Custom validator registration
func registerCustomValidators(v *validator.Validate) {
	v.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if value == "" {
			return true
		}
		return len(value) == 36 && value[8] == '-' && value[13] == '-' && value[18] == '-' && value[23] == '-'
	})
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
		return h.HandleError(c, errors.NewBusinessError("INVALID_JSON", "Invalid JSON payload").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation).
			WithSeverity(errors.SeverityError))
	}

	// Parse tenant ID to UUID
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return h.HandleError(c, errors.NewBusinessError("INVALID_TENANT_ID", "Invalid tenant ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	// Step 3: Extract Context
	// TODO: Extract tenant context and validate permissions

	// Step 4: Delegate to Service
	updatedTenant, err := h.service.UpdateTenant(ctx, tenantUUID, req)
	if err != nil {
		return h.HandleError(c, err)
	}

	// Convert tenant to API response
	response := h.tenantToAPIResponse(updatedTenant)

	// Step 5: Return Response
	return h.Success(c, response)
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
		return h.HandleError(c, errors.NewBusinessError("INVALID_TENANT_ID", "Invalid tenant ID format").
			WithHTTPStatus(400).
			WithCategory(errors.CategoryValidation))
	}

	// Step 3: Extract Context
	// TODO: Extract tenant context and validate permissions

	// Step 4: Delegate to Service
	if err := h.service.DeleteTenant(ctx, tenantUUID); err != nil {
		return h.HandleError(c, err)
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

	// Return form data for creation
	formData := map[string]interface{}{
		"action":     "create",
		"action_url": "/api/v1/tenants",
	}

	return h.Success(c, formData)
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
		return h.HandleError(c, errors.NewBusinessError("INVALID_TENANT_ID", "Invalid tenant ID format").
			WithCategory(errors.CategoryValidation).
			WithSeverity(errors.SeverityError))
	}

	// Get the tenant
	tenantEntity, err := h.service.GetTenantByID(ctx, tenantUUID)
	if err != nil {
		return h.HandleError(c, err)
	}

	// Convert to API response
	tenantResponse := h.tenantToAPIResponse(tenantEntity)

	// Return form data for editing
	formData := map[string]interface{}{
		"tenant":     tenantResponse,
		"action":     "edit",
		"action_url": "/api/v1/tenants/" + tenantID,
	}

	return h.Success(c, formData)
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
