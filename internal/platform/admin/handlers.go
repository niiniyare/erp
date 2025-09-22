package admin

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/admin/templates"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// AdminHandlers contains the dependencies for admin handlers
type AdminHandlers struct {
	tenantService tenant.Service
	startTime     time.Time
}

// NewAdminHandlers creates a new admin handlers instance
func NewAdminHandlers(tenantService tenant.Service) *AdminHandlers {
	return &AdminHandlers{
		tenantService: tenantService,
		startTime:     time.Now(),
	}
}

// DashboardHandler serves the admin dashboard
func (h *AdminHandlers) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get dashboard data
	data, err := h.getDashboardData(ctx)
	if err != nil {
		logger.Error("Failed to get dashboard data", logger.Fields{
			"error": err.Error(),
		})
		http.Error(w, "Failed to load dashboard", http.StatusInternalServerError)
		return
	}

	// Render template
	component := templates.Dashboard(data)
	if err := component.Render(ctx, w); err != nil {
		logger.Error("Failed to render dashboard template", logger.Fields{
			"error": err.Error(),
		})
		http.Error(w, "Failed to render dashboard", http.StatusInternalServerError)
		return
	}
}

// TenantsHandler serves the tenants management page
func (h *AdminHandlers) TenantsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get all tenants (bypass tenant context for admin)
	tenants, err := h.getAllTenants(ctx)
	if err != nil {
		logger.Error("Failed to get tenants", logger.Fields{
			"error": err.Error(),
		})
		http.Error(w, "Failed to load tenants", http.StatusInternalServerError)
		return
	}

	// Render template
	component := templates.TenantsList(tenants)
	if err := component.Render(ctx, w); err != nil {
		logger.Error("Failed to render tenants template", logger.Fields{
			"error": err.Error(),
		})
		http.Error(w, "Failed to render tenants", http.StatusInternalServerError)
		return
	}
}

// UsersHandler serves the users management page
func (h *AdminHandlers) UsersHandler(w http.ResponseWriter, r *http.Request) {
	// For now, just render a placeholder
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Users - AWO ERP Admin</title>
			<script src="https://cdn.tailwindcss.com"></script>
		</head>
		<body class="bg-gray-50">
			<div class="max-w-7xl mx-auto py-12 px-4">
				<h1 class="text-3xl font-bold text-gray-900 mb-8">User Management</h1>
				<div class="bg-white shadow rounded-lg p-6">
					<p class="text-gray-600">User management functionality coming soon...</p>
					<a href="/admin/dashboard" class="inline-block mt-4 text-indigo-600 hover:text-indigo-500">
						← Back to Dashboard
					</a>
				</div>
			</div>
		</body>
		</html>
	`))
}

// getDashboardData aggregates data for the dashboard
func (h *AdminHandlers) getDashboardData(ctx context.Context) (templates.DashboardData, error) {
	data := templates.DashboardData{
		SystemInfo: templates.SystemInfo{
			Version:   "0.1.0 beta",
			Uptime:    time.Since(h.startTime).Round(time.Second).String(),
			GoVersion: runtime.Version(),
		},
	}

	// Get tenant count
	tenants, err := h.getAllTenants(ctx)
	if err != nil {
		logger.Warn("Failed to get tenant count for dashboard", logger.Fields{
			"error": err.Error(),
		})
	} else {
		data.TenantCount = len(tenants)
	}

	// User count is harder to get without tenant context, so we'll set it to 0 for now
	data.UserCount = 0

	return data, nil
}

// getAllTenants gets all tenants in the system (admin operation)
func (h *AdminHandlers) getAllTenants(ctx context.Context) ([]tenant.Tenant, error) {
	// Create a system context that bypasses tenant restrictions
	systemCtx := context.Background()

	// Use maximum allowed limit to get tenants (offset 0, limit 100)
	tenantPointers, err := h.tenantService.ListTenants(systemCtx, 0, 100)
	if err != nil {
		return nil, err
	}

	// Convert pointers to values
	tenants := make([]tenant.Tenant, len(tenantPointers))
	for i, t := range tenantPointers {
		if t != nil {
			tenants[i] = *t
		}
	}

	return tenants, nil
}

