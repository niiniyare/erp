package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niiniyare/erp/cmd/server/services"
	"github.com/niiniyare/erp/internal/platform/middleware"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// TenantFiberHandler handles tenant management endpoints
type TenantFiberHandler struct {
	coreServices   *services.CoreServices
	tracingService tracing.TracingService
	metricsService *metrics.MetricsService
	logger         logger.Logger
}

// NewTenantFiberHandler creates a new tenant handler
func NewTenantFiberHandler(
	coreServices *services.CoreServices,
	tracingService tracing.TracingService,
	metricsService *metrics.MetricsService,
) *TenantFiberHandler {
	return &TenantFiberHandler{
		coreServices:   coreServices,
		tracingService: tracingService,
		metricsService: metricsService,
		logger:         logger.WithFields(logger.Fields{"handler": "tenant"}),
	}
}

// TenantListResponse represents tenant list item
type TenantListResponse struct {
	ID           string    `json:"id"`
	Slug         string    `json:"slug"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Status       string    `json:"status"`
	Timezone     string    `json:"timezone"`
	CurrencyCode string    `json:"currency_code"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TenantDetailResponse represents detailed tenant information
type TenantDetailResponse struct {
	ID           string         `json:"id"`
	Slug         string         `json:"slug"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	Subdomain    string         `json:"subdomain,omitempty"`
	Status       string         `json:"status"`
	Timezone     string         `json:"timezone"`
	CurrencyCode string         `json:"currency_code"`
	Industry     string         `json:"industry,omitempty"`
	CompanySize  string         `json:"company_size,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	Settings     map[string]any `json:"settings,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// CreateTenantRequest represents tenant creation request
type CreateTenantRequest struct {
	Slug         string `json:"slug" validate:"required,min=3,max=63"`
	Name         string `json:"name" validate:"required,min=1,max=255"`
	Email        string `json:"email" validate:"required,email"`
	Timezone     string `json:"timezone" validate:"required"`
	CurrencyCode string `json:"currency_code" validate:"required,len=3"`
	Industry     string `json:"industry,omitempty"`
	CompanySize  string `json:"company_size,omitempty"`
}

// UpdateTenantRequest represents tenant update request
type UpdateTenantRequest struct {
	Name     string         `json:"name,omitempty"`
	Email    string         `json:"email,omitempty"`
	Status   string         `json:"status,omitempty"`
	Settings map[string]any `json:"settings,omitempty"`
}

// TenantConfiguration represents tenant configuration
type TenantConfiguration struct {
	MaxUsers                int            `json:"max_users"`
	MaxEntities             int            `json:"max_entities"`
	MaxTransactionsPerMonth int            `json:"max_transactions_per_month"`
	StorageQuota            int64          `json:"storage_quota"`
	AccountingMethod        string         `json:"accounting_method"`
	FiscalYearStartMonth    int            `json:"fiscal_year_start_month"`
	DefaultCurrency         string         `json:"default_currency"`
	PasswordPolicy          map[string]any `json:"password_policy"`
	APIRateLimits           map[string]any `json:"api_rate_limits"`
}

// ListTenants handles GET /tenants (Admin Console Only)
func (h *TenantFiberHandler) ListTenants(c *fiber.Ctx) error {
	ctx, span := h.tracingService.StartSpan(c.Context(), "tenant.list")
	defer span.End()

	// Parse pagination parameters
	pagination := ParsePagination(c)

	// Parse query parameters
	status := c.Query("status")
	search := c.Query("search")

	h.logger.InfoContext(ctx, "Listing tenants", logger.Fields{
		"page":     pagination.Page,
		"per_page": pagination.PerPage,
		"status":   status,
		"search":   search,
	})

	// TODO: Implement actual tenant listing using tenant service
	// For now, return mock data
	tenants := []TenantListResponse{
		{
			ID:           "550e8400-e29b-41d4-a716-446655440000",
			Slug:         "acme-corp",
			Name:         "ACME Corporation",
			Email:        "admin@acme.com",
			Status:       "ACTIVE",
			Timezone:     "America/New_York",
			CurrencyCode: "USD",
			CreatedAt:    time.Now().AddDate(0, -6, 0),
			UpdatedAt:    time.Now(),
		},
	}

	paginationInfo := CalculatePagination(int64(len(tenants)), pagination.Page, pagination.PerPage)

	return SuccessPaginated(c, tenants, paginationInfo)
}

// GetTenant handles GET /tenants/{id}
func (h *TenantFiberHandler) GetTenant(c *fiber.Ctx) error {
	ctx, span := h.tracingService.StartSpan(c.Context(), "tenant.get")
	defer span.End()

	tenantID := c.Params("id")
	if tenantID == "" {
		return BadRequest(c, "Tenant ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(tenantID); err != nil {
		return BadRequest(c, "Invalid tenant ID format")
	}

	h.logger.InfoContext(ctx, "Getting tenant details", logger.Fields{
		"tenant_id": tenantID,
	})

	// TODO: Implement actual tenant retrieval using tenant service
	// For now, return mock data
	tenant := TenantDetailResponse{
		ID:           tenantID,
		Slug:         "acme-corp",
		Name:         "ACME Corporation",
		Email:        "admin@acme.com",
		Subdomain:    "acme",
		Status:       "ACTIVE",
		Timezone:     "America/New_York",
		CurrencyCode: "USD",
		Industry:     "Technology",
		CompanySize:  "ENTERPRISE",
		Metadata: map[string]any{
			"crm_id":    "12345",
			"sales_rep": "Jane Smith",
		},
		Settings: map[string]any{
			"allow_api_access": true,
			"sso_enabled":      true,
		},
		CreatedAt: time.Now().AddDate(0, -6, 0),
		UpdatedAt: time.Now(),
	}

	return Success(c, tenant)
}

// CreateTenant handles POST /tenants
func (h *TenantFiberHandler) CreateTenant(c *fiber.Ctx) error {
	ctx, span := h.tracingService.StartSpan(c.Context(), "tenant.create")
	defer span.End()

	var req CreateTenantRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	// Basic validation
	if req.Slug == "" || req.Name == "" || req.Email == "" {
		return ValidationError(c, "Slug, name, and email are required")
	}

	h.logger.InfoContext(ctx, "Creating new tenant", logger.Fields{
		"slug":  req.Slug,
		"name":  req.Name,
		"email": req.Email,
	})

	// TODO: Implement actual tenant creation using tenant service
	// For now, return mock response
	tenant := TenantDetailResponse{
		ID:           uuid.New().String(),
		Slug:         req.Slug,
		Name:         req.Name,
		Email:        req.Email,
		Status:       "TRIAL",
		Timezone:     req.Timezone,
		CurrencyCode: req.CurrencyCode,
		Industry:     req.Industry,
		CompanySize:  req.CompanySize,
		Metadata:     make(map[string]any),
		Settings:     make(map[string]any),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	h.logger.InfoContext(ctx, "Tenant created successfully", logger.Fields{
		"tenant_id": tenant.ID,
		"slug":      tenant.Slug,
	})

	return Created(c, tenant)
}

// UpdateTenant handles PATCH /tenants/{id}
func (h *TenantFiberHandler) UpdateTenant(c *fiber.Ctx) error {
	ctx, span := h.tracingService.StartSpan(c.Context(), "tenant.update")
	defer span.End()

	tenantID := c.Params("id")
	if tenantID == "" {
		return BadRequest(c, "Tenant ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(tenantID); err != nil {
		return BadRequest(c, "Invalid tenant ID format")
	}

	var req UpdateTenantRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	h.logger.InfoContext(ctx, "Updating tenant", logger.Fields{
		"tenant_id": tenantID,
	})

	// TODO: Implement actual tenant update using tenant service
	// For now, return mock updated tenant
	tenant := TenantDetailResponse{
		ID:           tenantID,
		Slug:         "acme-corp",
		Name:         req.Name,
		Email:        req.Email,
		Status:       req.Status,
		Timezone:     "America/New_York",
		CurrencyCode: "USD",
		Industry:     "Technology",
		CompanySize:  "ENTERPRISE",
		Settings:     req.Settings,
		CreatedAt:    time.Now().AddDate(0, -6, 0),
		UpdatedAt:    time.Now(),
	}

	return Success(c, tenant)
}

// GetTenantConfiguration handles GET /tenants/{id}/configuration
func (h *TenantFiberHandler) GetTenantConfiguration(c *fiber.Ctx) error {
	ctx, span := h.tracingService.StartSpan(c.Context(), "tenant.get_configuration")
	defer span.End()

	tenantID := c.Params("id")
	if tenantID == "" {
		return BadRequest(c, "Tenant ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(tenantID); err != nil {
		return BadRequest(c, "Invalid tenant ID format")
	}

	h.logger.InfoContext(ctx, "Getting tenant configuration", logger.Fields{
		"tenant_id": tenantID,
	})

	// TODO: Implement actual configuration retrieval
	// For now, return mock configuration
	config := TenantConfiguration{
		MaxUsers:                100,
		MaxEntities:             50,
		MaxTransactionsPerMonth: 10000,
		StorageQuota:            10737418240, // 10GB
		AccountingMethod:        "ACCRUAL",
		FiscalYearStartMonth:    1,
		DefaultCurrency:         "USD",
		PasswordPolicy: map[string]any{
			"min_length":        12,
			"require_uppercase": true,
			"require_numbers":   true,
		},
		APIRateLimits: map[string]any{
			"requests_per_minute": 60,
		},
	}

	return Success(c, config)
}

// UpdateTenantConfiguration handles PUT /tenants/{id}/configuration
func (h *TenantFiberHandler) UpdateTenantConfiguration(c *fiber.Ctx) error {
	ctx, span := h.tracingService.StartSpan(c.Context(), "tenant.update_configuration")
	defer span.End()

	tenantID := c.Params("id")
	if tenantID == "" {
		return BadRequest(c, "Tenant ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(tenantID); err != nil {
		return BadRequest(c, "Invalid tenant ID format")
	}

	var req TenantConfiguration
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	h.logger.InfoContext(ctx, "Updating tenant configuration", logger.Fields{
		"tenant_id": tenantID,
	})

	// TODO: Implement actual configuration update
	// For now, return the updated configuration
	return Success(c, req)
}

// GetCurrentTenant handles GET /tenant (for current tenant context)
func (h *TenantFiberHandler) GetCurrentTenant(c *fiber.Ctx) error {
	ctx, span := h.tracingService.StartSpan(c.Context(), "tenant.get_current")
	defer span.End()

	// Get tenant ID from context (set by tenant middleware)
	tenantID, ok := middleware.GetTenantIDFromFiber(c)
	if !ok {
		return BadRequest(c, "No tenant context found")
	}

	h.logger.InfoContext(ctx, "Getting current tenant", logger.Fields{
		"tenant_id": tenantID.String(),
	})

	// Use the existing GetTenant logic but with current tenant ID
	c.Params("id", tenantID.String())
	return h.GetTenant(c)
}

// SetupTenantRoutes sets up tenant routes for Fiber
func SetupTenantRoutes(app fiber.Router, handler *TenantFiberHandler) {
	// Admin Console routes (for tenant management)
	admin := app.Group("/admin/tenants")
	admin.Get("/", handler.ListTenants)
	admin.Post("/", handler.CreateTenant)
	admin.Get("/:id", handler.GetTenant)
	admin.Patch("/:id", handler.UpdateTenant)
	admin.Get("/:id/configuration", handler.GetTenantConfiguration)
	admin.Put("/:id/configuration", handler.UpdateTenantConfiguration)

	// Public tenant creation route
	app.Post("/tenants", handler.CreateTenant)

	// Current tenant route (tenant-scoped)
	app.Get("/tenant", handler.GetCurrentTenant)
}
