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

// TenantUnifiedHandler handles both HTML and JSON tenant requests
type TenantUnifiedHandler struct {
	coreServices   *services.CoreServices
	tracingService tracing.TracingService
	metricsService *metrics.MetricsService
	logger         logger.Logger
}

// NewTenantUnifiedHandler creates a new unified tenant handler
func NewTenantUnifiedHandler(
	coreServices *services.CoreServices,
	tracingService tracing.TracingService,
	metricsService *metrics.MetricsService,
) *TenantUnifiedHandler {
	return &TenantUnifiedHandler{
		coreServices:   coreServices,
		tracingService: tracingService,
		metricsService: metricsService,
		logger:         logger.WithFields(logger.Fields{"handler": "tenant_unified"}),
	}
}

// ListTenants handles GET /tenants - serves tenant list page or returns JSON data
func (h *TenantUnifiedHandler) ListTenants(c *fiber.Ctx) error {
	ctx, span := h.tracingService.StartSpan(c.Context(), "tenant.list")
	defer span.End()

	// Parse pagination parameters
	pagination := ParsePagination(c)
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
		{
			ID:           "550e8400-e29b-41d4-a716-446655440001",
			Slug:         "tech-startup",
			Name:         "Tech Startup Inc",
			Email:        "admin@techstartup.com",
			Status:       "TRIAL",
			Timezone:     "America/Los_Angeles",
			CurrencyCode: "USD",
			CreatedAt:    time.Now().AddDate(0, -2, 0),
			UpdatedAt:    time.Now(),
		},
	}

	paginationInfo := CalculatePagination(int64(len(tenants)), pagination.Page, pagination.PerPage)
	responseMode := middleware.GetResponseMode(c)

	switch responseMode {
	case middleware.ResponseModeJSON:
		// Return JSON for API consumers
		return SuccessPaginated(c, tenants, paginationInfo)

	case middleware.ResponseModeFragment:
		// Return HTML fragment for HTMX requests
		html := h.renderTenantsTable(tenants, paginationInfo)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)

	default: // ResponseModePage
		// Return full HTML page for browser navigation
		html := h.renderTenantsPage(tenants, paginationInfo)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

// GetTenant handles GET /tenants/{id} - serves tenant detail page or returns JSON
func (h *TenantUnifiedHandler) GetTenant(c *fiber.Ctx) error {
	ctx, span := h.tracingService.StartSpan(c.Context(), "tenant.get")
	defer span.End()

	tenantID := c.Params("id")
	if tenantID == "" {
		return h.handleTenantError(c, "Tenant ID is required", 400)
	}

	// Validate UUID format
	if _, err := uuid.Parse(tenantID); err != nil {
		return h.handleTenantError(c, "Invalid tenant ID format", 400)
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
		Metadata: map[string]interface{}{
			"crm_id":    "12345",
			"sales_rep": "Jane Smith",
		},
		Settings: map[string]interface{}{
			"allow_api_access": true,
			"sso_enabled":      true,
		},
		CreatedAt: time.Now().AddDate(0, -6, 0),
		UpdatedAt: time.Now(),
	}

	responseMode := middleware.GetResponseMode(c)

	switch responseMode {
	case middleware.ResponseModeJSON:
		return Success(c, tenant)

	case middleware.ResponseModeFragment:
		html := h.renderTenantDetail(tenant)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)

	default: // ResponseModePage
		html := h.renderTenantDetailPage(tenant)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

// CreateTenant handles POST /tenants - processes tenant creation
func (h *TenantUnifiedHandler) CreateTenant(c *fiber.Ctx) error {
	ctx, span := h.tracingService.StartSpan(c.Context(), "tenant.create")
	defer span.End()

	var req CreateTenantRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleTenantError(c, "Invalid request body", 400)
	}

	// Basic validation
	if req.Slug == "" || req.Name == "" || req.Email == "" {
		return h.handleTenantError(c, "Slug, name, and email are required", 400)
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
		Metadata:     make(map[string]interface{}),
		Settings:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	responseMode := middleware.GetResponseMode(c)

	switch responseMode {
	case middleware.ResponseModeJSON:
		return Created(c, tenant)

	case middleware.ResponseModeFragment:
		// For HTMX requests, refresh the tenant list
		middleware.SetHTMXHeaders(c, middleware.HTMXResponseHeaders{
			Trigger: "tenantCreated",
		})
		html := h.renderTenantDetail(tenant)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)

	default: // ResponseModePage
		// Redirect to tenant detail page
		return c.Redirect("/tenants/"+tenant.ID, 302)
	}
}

// handleTenantError handles tenant-related errors with appropriate response format
func (h *TenantUnifiedHandler) handleTenantError(c *fiber.Ctx, message string, code int) error {
	responseMode := middleware.GetResponseMode(c)

	switch responseMode {
	case middleware.ResponseModeJSON:
		return c.Status(code).JSON(fiber.Map{
			"error": message,
			"code":  code,
		})

	case middleware.ResponseModeFragment:
		// Return error fragment for HTMX requests
		html := `<div class="bg-red-50 border border-red-200 rounded-md p-4">
			<div class="text-sm text-red-700">` + message + `</div>
		</div>`
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Status(code).SendString(html)

	default: // ResponseModePage
		// Return error page for browser requests
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Error - ERP System</title>
    <meta charset="UTF-8">
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-50">
    <div class="min-h-screen flex items-center justify-center">
        <div class="max-w-md w-full">
            <div class="bg-red-50 border border-red-200 rounded-lg p-6">
                <h1 class="text-lg font-semibold text-red-800 mb-2">Error</h1>
                <p class="text-red-700">` + message + `</p>
                <div class="mt-4">
                    <a href="/tenants" class="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-white bg-red-600 hover:bg-red-700">
                        Back to Tenants
                    </a>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Status(code).SendString(html)
	}
}

// HTML rendering methods (simplified for demo)
func (h *TenantUnifiedHandler) renderTenantsPage(tenants []TenantListResponse, pagination *Pagination) string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>Tenants - ERP System</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://unpkg.com/htmx.org@1.9.8"></script>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-50">
    <div class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div class="px-4 py-6 sm:px-0">
            <div class="border-4 border-dashed border-gray-200 rounded-lg p-6">
                <h1 class="text-2xl font-bold text-gray-900 mb-6">Tenants</h1>
                ` + h.renderTenantsTable(tenants, pagination) + `
            </div>
        </div>
    </div>
</body>
</html>`
}

func (h *TenantUnifiedHandler) renderTenantsTable(tenants []TenantListResponse, pagination *Pagination) string {
	html := `<div id="tenants-table" hx-get="/tenants" hx-trigger="tenantCreated from:body">
        <div class="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
            <table class="min-w-full divide-y divide-gray-300">
                <thead class="bg-gray-50">
                    <tr>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                    </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">`

	for _, tenant := range tenants {
		html += `<tr>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <div class="text-sm font-medium text-gray-900">` + tenant.Name + `</div>
                        <div class="text-sm text-gray-500">` + tenant.Email + `</div>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">` + tenant.Status + `</span>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">` + tenant.CreatedAt.Format("Jan 2, 2006") + `</td>
                </tr>`
	}

	html += `</tbody>
            </table>
        </div>
    </div>`

	return html
}

func (h *TenantUnifiedHandler) renderTenantDetailPage(tenant TenantDetailResponse) string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>` + tenant.Name + ` - ERP System</title>
    <meta charset="UTF-8">
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-50">
    <div class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        ` + h.renderTenantDetail(tenant) + `
    </div>
</body>
</html>`
}

func (h *TenantUnifiedHandler) renderTenantDetail(tenant TenantDetailResponse) string {
	return `<div class="bg-white overflow-hidden shadow rounded-lg">
        <div class="px-4 py-5 sm:p-6">
            <h1 class="text-2xl font-bold text-gray-900 mb-6">` + tenant.Name + `</h1>
            <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
                <div>
                    <dt class="text-sm font-medium text-gray-500">Email</dt>
                    <dd class="mt-1 text-sm text-gray-900">` + tenant.Email + `</dd>
                </div>
                <div>
                    <dt class="text-sm font-medium text-gray-500">Status</dt>
                    <dd class="mt-1 text-sm text-gray-900">` + tenant.Status + `</dd>
                </div>
                <div>
                    <dt class="text-sm font-medium text-gray-500">Timezone</dt>
                    <dd class="mt-1 text-sm text-gray-900">` + tenant.Timezone + `</dd>
                </div>
                <div>
                    <dt class="text-sm font-medium text-gray-500">Currency</dt>
                    <dd class="mt-1 text-sm text-gray-900">` + tenant.CurrencyCode + `</dd>
                </div>
            </dl>
        </div>
    </div>`
}

// SetupTenantUnifiedRoutes sets up unified tenant routes
func SetupTenantUnifiedRoutes(app fiber.Router, handler *TenantUnifiedHandler) {
	// Tenant routes (both HTML and JSON)
	app.Get("/tenants", handler.ListTenants)
	app.Post("/tenants", handler.CreateTenant)
	app.Get("/tenants/:id", handler.GetTenant)

	// API routes with explicit /api prefix
	api := app.Group("/api")
	api.Get("/tenants", handler.ListTenants)
	api.Post("/tenants", handler.CreateTenant)
	api.Get("/tenants/:id", handler.GetTenant)
}
