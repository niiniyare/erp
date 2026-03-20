package admin

//FIXME:
// import (
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"runtime"
// 	"strings"
// 	"time"
//
// 	"github.com/google/uuid"
// 	"awo/internal/core/tenant"
// 	"awo/internal/platform/admin/templates"
// 	"awo/internal/shared/logger"
// )
//
// // AdminHandlers contains the dependencies for admin handlers
// type AdminHandlers struct {
// 	tenantService tenant.Service
// 	startTime     time.Time
// }
//
// // NewAdminHandlers creates a new admin handlers instance
// func NewAdminHandlers(tenantService tenant.Service) *AdminHandlers {
// 	return &AdminHandlers{
// 		tenantService: tenantService,
// 		startTime:     time.Now(),
// 	}
// }
//
// // DashboardHandler serves the admin dashboard
// func (h *AdminHandlers) DashboardHandler(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get dashboard data
// 	data, err := h.getDashboardData(ctx)
// 	if err != nil {
// 		logger.Error("Failed to get dashboard data", logger.Fields{
// 			"error": err.Error(),
// 		})
// 		http.Error(w, "Failed to load dashboard", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Render template
// 	component := templates.Dashboard(data)
// 	if err := component.Render(ctx, w); err != nil {
// 		logger.Error("Failed to render dashboard template", logger.Fields{
// 			"error": err.Error(),
// 		})
// 		http.Error(w, "Failed to render dashboard", http.StatusInternalServerError)
// 		return
// 	}
// }
//
// // TenantsHandler serves the tenants management page
// func (h *AdminHandlers) TenantsHandler(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get all tenants (bypass tenant context for admin)
// 	tenants, err := h.getAllTenants(ctx)
// 	if err != nil {
// 		logger.Error("Failed to get tenants", logger.Fields{
// 			"error": err.Error(),
// 		})
// 		http.Error(w, "Failed to load tenants", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Render template
// 	component := templates.TenantsList(tenants)
// 	if err := component.Render(ctx, w); err != nil {
// 		logger.Error("Failed to render tenants template", logger.Fields{
// 			"error": err.Error(),
// 		})
// 		http.Error(w, "Failed to render tenants", http.StatusInternalServerError)
// 		return
// 	}
// }
//
// // UsersHandler serves the users management page
// func (h *AdminHandlers) UsersHandler(w http.ResponseWriter, r *http.Request) {
// 	// For now, just render a placeholder
// 	w.Header().Set("Content-Type", "text/html")
// 	w.Write([]byte(`
// 		<!DOCTYPE html>
// 		<html>
// 		<head>
// 			<title>Users - AWO ERP Admin</title>
// 			<script src="https://cdn.tailwindcss.com"></script>
// 		</head>
// 		<body class="bg-gray-50">
// 			<div class="max-w-7xl mx-auto py-12 px-4">
// 				<h1 class="text-3xl font-bold text-gray-900 mb-8">User Management</h1>
// 				<div class="bg-white shadow rounded-lg p-6">
// 					<p class="text-gray-600">User management functionality coming soon...</p>
// 					<a href="/admin/dashboard" class="inline-block mt-4 text-indigo-600 hover:text-indigo-500">
// 						← Back to Dashboard
// 					</a>
// 				</div>
// 			</div>
// 		</body>
// 		</html>
// 	`))
// }
//
// // getDashboardData aggregates data for the dashboard
// func (h *AdminHandlers) getDashboardData(ctx context.Context) (templates.DashboardData, error) {
// 	data := templates.DashboardData{
// 		SystemInfo: templates.SystemInfo{
// 			Version:   "0.1.0 beta",
// 			Uptime:    time.Since(h.startTime).Round(time.Second).String(),
// 			GoVersion: runtime.Version(),
// 		},
// 	}
//
// 	// Get tenant count
// 	tenants, err := h.getAllTenants(ctx)
// 	if err != nil {
// 		logger.Warn("Failed to get tenant count for dashboard", logger.Fields{
// 			"error": err.Error(),
// 		})
// 	} else {
// 		data.TenantCount = len(tenants)
// 	}
//
// 	// User count is harder to get without tenant context, so we'll set it to 0 for now
// 	data.UserCount = 0
//
// 	return data, nil
// }
//
// // getAllTenants gets all tenants in the system (admin operation)
// func (h *AdminHandlers) getAllTenants(ctx context.Context) ([]tenant.Tenant, error) {
// 	// Create a system context that bypasses tenant restrictions
// 	systemCtx := context.Background()
//
// 	// For admin operations, we need to bypass RLS to see all tenants
// 	// Clear any tenant context first
// 	if err := h.clearTenantContext(systemCtx); err != nil {
// 		logger.Warn("Failed to clear tenant context", logger.Fields{"error": err.Error()})
// 	}
//
// 	// Use maximum allowed limit to get tenants (offset 0, limit 100)
// 	tenantPointers, err := h.tenantService.ListTenants(systemCtx, 0, 100)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// Convert pointers to values
// 	tenants := make([]tenant.Tenant, len(tenantPointers))
// 	for i, t := range tenantPointers {
// 		if t != nil {
// 			tenants[i] = *t
// 		}
// 	}
//
// 	return tenants, nil
// }
//
// // clearTenantContext ensures no tenant context is set for admin operations
// func (h *AdminHandlers) clearTenantContext(ctx context.Context) error {
// 	// This is a placeholder - we'll implement the actual clearing if needed
// 	return nil
// }
//
// // NewTenantHandler serves the create tenant form
// func (h *AdminHandlers) NewTenantHandler(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	data := templates.TenantFormData{
// 		Tenant: nil,
// 		IsEdit: false,
// 		Errors: make(map[string]string),
// 	}
//
// 	component := templates.TenantForm(data)
// 	if err := component.Render(ctx, w); err != nil {
// 		logger.Error("Failed to render new tenant form", logger.Fields{
// 			"error": err.Error(),
// 		})
// 		http.Error(w, "Failed to render form", http.StatusInternalServerError)
// 		return
// 	}
// }
//
// // CreateTenantHandler handles POST requests to create a new tenant
// func (h *AdminHandlers) CreateTenantHandler(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	if err := r.ParseForm(); err != nil {
// 		http.Error(w, "Invalid form data", http.StatusBadRequest)
// 		return
// 	}
//
// 	// Extract form values
// 	name := strings.TrimSpace(r.FormValue("name"))
// 	slug := strings.TrimSpace(r.FormValue("slug"))
// 	email := strings.TrimSpace(r.FormValue("email"))
// 	subdomain := strings.TrimSpace(r.FormValue("subdomain"))
// 	statusStr := r.FormValue("status")
// 	timezone := r.FormValue("timezone")
// 	currencyCode := r.FormValue("currency_code")
//
// 	// Validation
// 	errors := make(map[string]string)
// 	if name == "" {
// 		errors["name"] = "Name is required"
// 	}
// 	if slug == "" {
// 		errors["slug"] = "Slug is required"
// 	}
// 	if email == "" {
// 		errors["email"] = "Email is required"
// 	}
//
// 	// If validation failed, show form with errors
// 	if len(errors) > 0 {
// 		data := templates.TenantFormData{
// 			Tenant: &tenant.Tenant{
// 				Name:         name,
// 				Slug:         slug,
// 				Email:        email,
// 				Subdomain:    &subdomain,
// 				Status:       tenant.Status(statusStr),
// 				Timezone:     timezone,
// 				CurrencyCode: currencyCode,
// 			},
// 			IsEdit: false,
// 			Errors: errors,
// 		}
//
// 		component := templates.TenantForm(data)
// 		component.Render(ctx, w)
// 		return
// 	}
//
// 	// Create tenant
// 	systemCtx := context.Background()
//
// 	var subdomainPtr *string
// 	if subdomain != "" {
// 		subdomainPtr = &subdomain
// 	}
//
// 	newTenantReq := tenant.CreateTenantRequest{
// 		Name:      name,
// 		Slug:      slug,
// 		Email:     email,
// 		Subdomain: subdomainPtr,
// 		Status:    tenant.Status(statusStr),
// 	}
//
// 	createdTenant, err := h.tenantService.CreateTenant(systemCtx, newTenantReq)
// 	if err != nil {
// 		logger.Error("Failed to create tenant", logger.Fields{
// 			"error": err.Error(),
// 			"name":  name,
// 			"slug":  slug,
// 		})
//
// 		errors["general"] = fmt.Sprintf("Failed to create tenant: %v", err)
// 		data := templates.TenantFormData{
// 			Tenant: &tenant.Tenant{
// 				Name:      name,
// 				Slug:      slug,
// 				Email:     email,
// 				Subdomain: subdomainPtr,
// 				Status:    tenant.Status(statusStr),
// 			},
// 			IsEdit: false,
// 			Errors: errors,
// 		}
//
// 		component := templates.TenantForm(data)
// 		component.Render(ctx, w)
// 		return
// 	}
//
// 	logger.Info("Tenant created successfully", logger.Fields{
// 		"id":   createdTenant.ID.String(),
// 		"name": createdTenant.Name,
// 		"slug": createdTenant.Slug,
// 	})
//
// 	// Redirect to tenants list
// 	http.Redirect(w, r, "/admin/tenants", http.StatusSeeOther)
// }
//
// // ViewTenantHandler serves the tenant detail view
// func (h *AdminHandlers) ViewTenantHandler(w http.ResponseWriter, r *http.Request) {
// 	// Remove unused ctx variable
// 	_ = r.Context()
//
// 	// Extract tenant ID from URL path
// 	tenantIDStr := strings.TrimPrefix(r.URL.Path, "/admin/tenants/")
// 	tenantID, err := uuid.Parse(tenantIDStr)
// 	if err != nil {
// 		http.NotFound(w, r)
// 		return
// 	}
//
// 	// Get tenant
// 	systemCtx := context.Background()
// 	tenantPtr, err := h.tenantService.GetTenantByID(systemCtx, tenantID)
// 	if err != nil {
// 		logger.Error("Failed to get tenant", logger.Fields{
// 			"error":     err.Error(),
// 			"tenant_id": tenantID.String(),
// 		})
// 		http.NotFound(w, r)
// 		return
// 	}
//
// 	// For now, just show tenant details as JSON (can be improved with a proper template)
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write([]byte(fmt.Sprintf(`{
// 		"id": "%s",
// 		"name": "%s",
// 		"slug": "%s",
// 		"email": "%s",
// 		"status": "%s",
// 		"created_at": "%s"
// 	}`, tenantPtr.ID, tenantPtr.Name, tenantPtr.Slug, tenantPtr.Email, tenantPtr.Status, tenantPtr.CreatedAt)))
// }
//
// // EditTenantHandler serves the edit tenant form
// func (h *AdminHandlers) EditTenantHandler(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Extract tenant ID from URL path
// 	path := strings.TrimPrefix(r.URL.Path, "/admin/tenants/")
// 	path = strings.TrimSuffix(path, "/edit")
// 	tenantID, err := uuid.Parse(path)
// 	if err != nil {
// 		http.NotFound(w, r)
// 		return
// 	}
//
// 	// Get tenant
// 	systemCtx := context.Background()
// 	tenantPtr, err := h.tenantService.GetTenantByID(systemCtx, tenantID)
// 	if err != nil {
// 		logger.Error("Failed to get tenant for edit", logger.Fields{
// 			"error":     err.Error(),
// 			"tenant_id": tenantID.String(),
// 		})
// 		http.NotFound(w, r)
// 		return
// 	}
//
// 	data := templates.TenantFormData{
// 		Tenant: tenantPtr,
// 		IsEdit: true,
// 		Errors: make(map[string]string),
// 	}
//
// 	component := templates.TenantForm(data)
// 	if err := component.Render(ctx, w); err != nil {
// 		logger.Error("Failed to render edit tenant form", logger.Fields{
// 			"error": err.Error(),
// 		})
// 		http.Error(w, "Failed to render form", http.StatusInternalServerError)
// 		return
// 	}
// }
//
// // UpdateTenantHandler handles POST requests to update a tenant
// func (h *AdminHandlers) UpdateTenantHandler(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Extract tenant ID from URL path
// 	tenantIDStr := strings.TrimPrefix(r.URL.Path, "/admin/tenants/")
// 	tenantID, err := uuid.Parse(tenantIDStr)
// 	if err != nil {
// 		http.NotFound(w, r)
// 		return
// 	}
//
// 	if err := r.ParseForm(); err != nil {
// 		http.Error(w, "Invalid form data", http.StatusBadRequest)
// 		return
// 	}
//
// 	// Get existing tenant
// 	systemCtx := context.Background()
// 	existingTenant, err := h.tenantService.GetTenantByID(systemCtx, tenantID)
// 	if err != nil {
// 		http.NotFound(w, r)
// 		return
// 	}
//
// 	// Extract form values
// 	name := strings.TrimSpace(r.FormValue("name"))
// 	slug := strings.TrimSpace(r.FormValue("slug"))
// 	email := strings.TrimSpace(r.FormValue("email"))
// 	subdomain := strings.TrimSpace(r.FormValue("subdomain"))
// 	statusStr := r.FormValue("status")
// 	timezone := r.FormValue("timezone")
// 	currencyCode := r.FormValue("currency_code")
//
// 	// Validation
// 	errors := make(map[string]string)
// 	if name == "" {
// 		errors["name"] = "Name is required"
// 	}
// 	if slug == "" {
// 		errors["slug"] = "Slug is required"
// 	}
// 	if email == "" {
// 		errors["email"] = "Email is required"
// 	}
//
// 	// If validation failed, show form with errors
// 	if len(errors) > 0 {
// 		data := templates.TenantFormData{
// 			Tenant: &tenant.Tenant{
// 				ID:           tenantID,
// 				Name:         name,
// 				Slug:         slug,
// 				Email:        email,
// 				Subdomain:    &subdomain,
// 				Status:       tenant.Status(statusStr),
// 				Timezone:     timezone,
// 				CurrencyCode: currencyCode,
// 			},
// 			IsEdit: true,
// 			Errors: errors,
// 		}
//
// 		component := templates.TenantForm(data)
// 		component.Render(ctx, w)
// 		return
// 	}
//
// 	// Update tenant
// 	var subdomainPtr *string
// 	if subdomain != "" {
// 		subdomainPtr = &subdomain
// 	}
//
// 	status := tenant.Status(statusStr)
// 	updateReq := tenant.UpdateTenantRequest{
// 		Name:      &name,
// 		Email:     &email,
// 		Subdomain: subdomainPtr,
// 		Status:    &status,
// 	}
//
// 	_, err = h.tenantService.UpdateTenant(systemCtx, tenantID, updateReq)
// 	if err != nil {
// 		logger.Error("Failed to update tenant", logger.Fields{
// 			"error":     err.Error(),
// 			"tenant_id": tenantID.String(),
// 		})
//
// 		errors["general"] = fmt.Sprintf("Failed to update tenant: %v", err)
// 		data := templates.TenantFormData{
// 			Tenant: &tenant.Tenant{
// 				ID:        tenantID,
// 				Name:      name,
// 				Slug:      slug,
// 				Email:     email,
// 				Subdomain: subdomainPtr,
// 				Status:    tenant.Status(statusStr),
// 				CreatedAt: existingTenant.CreatedAt,
// 			},
// 			IsEdit: true,
// 			Errors: errors,
// 		}
//
// 		component := templates.TenantForm(data)
// 		component.Render(ctx, w)
// 		return
// 	}
//
// 	logger.Info("Tenant updated successfully", logger.Fields{
// 		"tenant_id": tenantID.String(),
// 		"name":      name,
// 		"slug":      slug,
// 	})
//
// 	// Redirect to tenants list
// 	http.Redirect(w, r, "/admin/tenants", http.StatusSeeOther)
// }
//
// // DeleteTenantHandler handles POST requests to delete a tenant
// func (h *AdminHandlers) DeleteTenantHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}
//
// 	// Extract tenant ID from URL path
// 	path := strings.TrimPrefix(r.URL.Path, "/admin/tenants/")
// 	path = strings.TrimSuffix(path, "/delete")
// 	tenantID, err := uuid.Parse(path)
// 	if err != nil {
// 		http.NotFound(w, r)
// 		return
// 	}
//
// 	// Delete tenant
// 	systemCtx := context.Background()
// 	err = h.tenantService.DeleteTenant(systemCtx, tenantID)
// 	if err != nil {
// 		logger.Error("Failed to delete tenant", logger.Fields{
// 			"error":     err.Error(),
// 			"tenant_id": tenantID.String(),
// 		})
// 		http.Error(w, fmt.Sprintf("Failed to delete tenant: %v", err), http.StatusInternalServerError)
// 		return
// 	}
//
// 	logger.Info("Tenant deleted successfully", logger.Fields{
// 		"tenant_id": tenantID.String(),
// 	})
//
// 	// Redirect to tenants list
// 	http.Redirect(w, r, "/admin/tenants", http.StatusSeeOther)
// }
