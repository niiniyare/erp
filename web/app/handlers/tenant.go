package handlers

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"strconv"
//
// 	"github.com/google/uuid"
// 	"github.com/niiniyare/erp/internal/shared/logger"
// 	"github.com/niiniyare/erp/internal/ui/middleware"
// 	temporalUI "github.com/niiniyare/erp/internal/ui/services/temporal"
// 	"github.com/niiniyare/erp/internal/ui/types"
// )
//
// // TenantHandler handles tenant management operations for admin console using Temporal workflows
// type TenantHandler struct {
// 	temporalClient *temporalUI.UITemporalClient
// 	logger         logger.Logger
// }
//
// // NewTenantHandler creates a new tenant management handler with Temporal integration
// func NewTenantHandler(
// 	temporalClient *temporalUI.UITemporalClient,
// 	logger logger.Logger,
// ) *TenantHandler {
// 	return &TenantHandler{
// 		temporalClient: temporalClient,
// 		logger:         logger,
// 	}
// }
//
// // ListTenants displays the tenant management page with DataTable integration
// func (h *TenantHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context (user should be authenticated by middleware)
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Ensure user has console admin role
// 	if err := middleware.RequireRole(ctx, middleware.UIRoleConsoleAdmin); err != nil {
// 		http.Error(w, "Forbidden", http.StatusForbidden)
// 		return
// 	}
//
// 	// Parse query parameters for filtering
// 	filters := h.parseTenantFilters(r)
//
// 	// Get tenants via Temporal workflow
// 	tenantWorkflows := h.temporalClient.GetTenantWorkflows()
// 	result, err := tenantWorkflows.ListTenants(ctx, filters)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to list tenants via Temporal", logger.Fields{
// 			"error":   err.Error(),
// 			"user_id": uiCtx.UserID,
// 			"filters": filters,
// 		})
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Prepare page data for Console DataTable
// 	pageData := &types.ConsoleTenantsPageData{
// 		Title:   "Tenant Management",
// 		Tenants: result.Tenants,
// 		Pagination: &types.PaginationMeta{
// 			Page:       result.Page,
// 			PageSize:   result.PageSize,
// 			Total:      result.TotalCount,
// 			TotalPages: (result.TotalCount + result.PageSize - 1) / result.PageSize,
// 			HasNext:    result.HasNext,
// 			HasPrev:    result.HasPrev,
// 		},
// 		Filters:   filters,
// 		CSRFToken: uiCtx.CSRFToken,
// 	}
//
// 	// Render tenant list template with Console DataTable
// 	h.renderTenantsListTemp(w, pageData)
//
// 	// Log tenant list access
// 	h.logger.InfoContext(ctx, "Console tenants list accessed via Temporal", logger.Fields{
// 		"user_id":     uiCtx.UserID,
// 		"total_count": result.TotalCount,
// 		"page":        result.Page,
// 		"page_size":   result.PageSize,
// 	})
// }
//
// // GetTenantsData returns tenant list data as JSON for DataTable HTMX requests
// func (h *TenantHandler) GetTenantsData(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Ensure admin access
// 	if err := middleware.RequireRole(ctx, middleware.UIRoleConsoleAdmin); err != nil {
// 		http.Error(w, "Forbidden", http.StatusForbidden)
// 		return
// 	}
//
// 	// Parse filters
// 	filters := h.parseTenantFilters(r)
//
// 	// Get tenants data via Temporal
// 	tenantWorkflows := h.temporalClient.GetTenantWorkflows()
// 	result, err := tenantWorkflows.ListTenants(ctx, filters)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to get tenants data via Temporal", logger.Fields{
// 			"error":   err.Error(),
// 			"user_id": uiCtx.UserID,
// 		})
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Return JSON response for DataTable
// 	response := map[string]interface{}{
// 		"tenants":    result.Tenants,
// 		"pagination": result,
// 		"success":    true,
// 	}
//
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(response)
// }
//
// // ShowCreateTenantForm displays the tenant creation form
// func (h *TenantHandler) ShowCreateTenantForm(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Ensure admin access
// 	if err := middleware.RequireRole(ctx, middleware.UIRoleConsoleAdmin); err != nil {
// 		http.Error(w, "Forbidden", http.StatusForbidden)
// 		return
// 	}
//
// 	// Prepare form data
// 	formData := &types.ConsoleTenantFormData{
// 		Title:     "Create New Tenant",
// 		Action:    "/console/tenants",
// 		Method:    "POST",
// 		CSRFToken: uiCtx.CSRFToken,
// 	}
//
// 	// Render tenant creation form
// 	h.renderTenantFormTemp(w, formData)
// }
//
// // CreateTenant creates a new tenant using Temporal onboarding workflow
// func (h *TenantHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Ensure admin access
// 	if err := middleware.RequireRole(ctx, middleware.UIRoleConsoleAdmin); err != nil {
// 		http.Error(w, "Forbidden", http.StatusForbidden)
// 		return
// 	}
//
// 	// Parse form data
// 	if err := r.ParseForm(); err != nil {
// 		http.Error(w, "Invalid form data", http.StatusBadRequest)
// 		return
// 	}
//
// 	// Extract tenant data
// 	tenantData := &types.ConsoleTenantCreateRequest{
// 		Name:         r.FormValue("name"),
// 		Slug:         r.FormValue("slug"),
// 		Industry:     r.FormValue("industry"),
// 		ContactEmail: r.FormValue("contact_email"),
// 		ContactPhone: r.FormValue("contact_phone"),
// 		CompanyName:  r.FormValue("company_name"),
// 	}
//
// 	// Validate tenant data
// 	if err := h.validateTenantData(tenantData); err != nil {
// 		h.renderTenantFormWithError(w, tenantData, err.Error(), uiCtx.CSRFToken)
// 		return
// 	}
//
// 	// Create tenant via Temporal onboarding workflow
// 	tenantWorkflows := h.temporalClient.GetTenantWorkflows()
// 	result, err := tenantWorkflows.CreateTenant(ctx, tenantData)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to create tenant via Temporal workflow", logger.Fields{
// 			"error":       err.Error(),
// 			"user_id":     uiCtx.UserID,
// 			"tenant_name": tenantData.Name,
// 		})
// 		h.renderTenantFormWithError(w, tenantData, "Failed to create tenant", uiCtx.CSRFToken)
// 		return
// 	}
//
// 	// Log tenant creation
// 	h.logger.InfoContext(ctx, "Tenant created via Temporal onboarding workflow", logger.Fields{
// 		"tenant_id":   result.TenantID,
// 		"workflow_id": result.WorkflowID,
// 		"creator_id":  uiCtx.UserID,
// 		"tenant_name": tenantData.Name,
// 	})
//
// 	// Redirect to tenant list with success message
// 	http.Redirect(w, r, "/console/tenants?success=tenant_created", http.StatusSeeOther)
// }
//
// // BulkTenantActions handles bulk operations on tenants via Temporal workflows
// func (h *TenantHandler) BulkTenantActions(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Ensure admin access
// 	if err := middleware.RequireRole(ctx, middleware.UIRoleConsoleAdmin); err != nil {
// 		http.Error(w, "Forbidden", http.StatusForbidden)
// 		return
// 	}
//
// 	// Parse request
// 	if err := r.ParseForm(); err != nil {
// 		http.Error(w, "Invalid form data", http.StatusBadRequest)
// 		return
// 	}
//
// 	action := r.FormValue("action")
// 	tenantIDStrs := r.Form["ids[]"]
//
// 	if len(tenantIDStrs) == 0 {
// 		http.Error(w, "No tenants selected", http.StatusBadRequest)
// 		return
// 	}
//
// 	// Validate action
// 	switch action {
// 	case "activate", "suspend", "delete":
// 		// Valid actions
// 	default:
// 		http.Error(w, "Invalid action", http.StatusBadRequest)
// 		return
// 	}
//
// 	// Convert tenant IDs
// 	tenantIDs := make([]uuid.UUID, 0, len(tenantIDStrs))
// 	for _, idStr := range tenantIDStrs {
// 		id, err := uuid.Parse(idStr)
// 		if err != nil {
// 			http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
// 			return
// 		}
// 		tenantIDs = append(tenantIDs, id)
// 	}
//
// 	// Process bulk action via Temporal workflow
// 	tenantWorkflows := h.temporalClient.GetTenantWorkflows()
// 	results, err := tenantWorkflows.ProcessBulkTenantAction(ctx, action, tenantIDs)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to process bulk tenant action via Temporal", logger.Fields{
// 			"error":        err.Error(),
// 			"user_id":      uiCtx.UserID,
// 			"action":       action,
// 			"tenant_count": len(tenantIDs),
// 		})
// 		http.Error(w, "Failed to process bulk action", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Log bulk action
// 	h.logger.InfoContext(ctx, "Bulk tenant action processed via Temporal", logger.Fields{
// 		"user_id":       uiCtx.UserID,
// 		"action":        action,
// 		"tenant_count":  len(tenantIDs),
// 		"success_count": results.SuccessCount,
// 		"error_count":   results.ErrorCount,
// 	})
//
// 	// Return JSON response
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]interface{}{
// 		"success": true,
// 		"message": fmt.Sprintf("Bulk action completed: %d successful, %d failed", results.SuccessCount, results.ErrorCount),
// 		"results": results,
// 	})
// }
//
// // Private helper methods
//
// // parseTenantFilters extracts tenant filtering parameters from request
// func (h *TenantHandler) parseTenantFilters(r *http.Request) *types.ConsoleTenantFilters {
// 	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
// 	if page < 1 {
// 		page = 1
// 	}
//
// 	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
// 	if pageSize < 1 || pageSize > 100 {
// 		pageSize = 20
// 	}
//
// 	return &types.ConsoleTenantFilters{
// 		Page:     page,
// 		PageSize: pageSize,
// 		Search:   r.URL.Query().Get("q"),
// 		Status:   r.URL.Query().Get("status"),
// 		Industry: r.URL.Query().Get("industry"),
// 	}
// }
//
// // validateTenantData validates tenant creation data
// func (h *TenantHandler) validateTenantData(data *types.ConsoleTenantCreateRequest) error {
// 	if data.Name == "" {
// 		return fmt.Errorf("tenant name is required")
// 	}
// 	if data.Slug == "" {
// 		return fmt.Errorf("tenant slug is required")
// 	}
// 	if data.ContactEmail == "" {
// 		return fmt.Errorf("contact email is required")
// 	}
// 	if data.CompanyName == "" {
// 		return fmt.Errorf("company name is required")
// 	}
// 	return nil
// }
//
// // Temporary rendering methods (until templ templates are created)
//
// func (h *TenantHandler) renderTenantsListTemp(w http.ResponseWriter, data *types.ConsoleTenantsPageData) {
// 	w.Header().Set("Content-Type", "text/html")
// 	fmt.Fprintf(w, `
// 		<!DOCTYPE html>
// 		<html>
// 		<head>
// 			<title>%s</title>
// 			<script src="/static/hooks/datatable.js"></script>
// 			<style>
// 				body { font-family: Arial, sans-serif; margin: 40px; background: #f8fafc; }
// 				.header { background: linear-gradient(135deg, #1e40af, #1d4ed8); color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
// 				.actions { margin: 20px 0; }
// 				.btn { background: #1e40af; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; margin-right: 10px; }
// 				.table { width: 100%%; border-collapse: collapse; background: white; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
// 				.table th, .table td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
// 				.table th { background: #f8fafc; }
// 				.status-active { color: #059669; }
// 				.status-inactive { color: #dc2626; }
// 				.status-suspended { color: #d97706; }
// 				.status-trial { color: #7c3aed; }
// 			</style>
// 		</head>
// 		<body>
// 			<div class="header">
// 				<h1>%s</h1>
// 				<p>Manage tenant organizations across the ERP system</p>
// 				<p><strong>🔗 Powered by Temporal Workflows</strong></p>
// 			</div>
//
// 			<div class="actions">
// 				<button class="btn" onclick="location.href='/console/tenants/new'">Create New Tenant</button>
// 				<button class="btn" id="bulk-activate">Activate Selected</button>
// 				<button class="btn" id="bulk-suspend">Suspend Selected</button>
// 				<button class="btn" id="bulk-delete" style="background: #dc2626;">Delete Selected</button>
// 			</div>
//
// 			<div x-data="createConsoleDataTable({
// 				rows: %s,
// 				searchColumns: ['name', 'slug', 'contactInfo.email', 'status', 'industry'],
// 				serviceContext: 'console'
// 			})">
// 				<div class="filters" style="margin-bottom: 20px; background: white; padding: 15px; border-radius: 8px;">
// 					<input type="text" x-model="searchQuery" @input="handleSearch" placeholder="Search tenants..." style="padding: 8px; margin-right: 10px; width: 300px;">
// 					<select @change="applyFilter" style="padding: 8px; margin-right: 10px;">
// 						<option value="all">All Statuses</option>
// 						<option value="active">Active</option>
// 						<option value="inactive">Inactive</option>
// 						<option value="suspended">Suspended</option>
// 						<option value="trial">Trial</option>
// 					</select>
// 					<select @change="applyFilter" style="padding: 8px;">
// 						<option value="all">All Industries</option>
// 						<option value="Technology">Technology</option>
// 						<option value="Healthcare">Healthcare</option>
// 						<option value="Finance">Finance</option>
// 						<option value="Manufacturing">Manufacturing</option>
// 						<option value="Retail">Retail</option>
// 					</select>
// 				</div>
//
// 				<table class="table">
// 					<thead>
// 						<tr>
// 							<th><input type="checkbox" @change="toggleSelectAll"></th>
// 							<th @click="sort('name')">Tenant Name</th>
// 							<th @click="sort('slug')">Slug</th>
// 							<th @click="sort('contactInfo.email')">Contact Email</th>
// 							<th @click="sort('industry')">Industry</th>
// 							<th @click="sort('status')">Status</th>
// 							<th @click="sort('stats.userCount')">Users</th>
// 							<th @click="sort('createdAt')">Created</th>
// 							<th>Actions</th>
// 						</tr>
// 					</thead>
// 					<tbody>
// 						<template x-for="tenant in paginatedRows" :key="tenant.id">
// 							<tr>
// 								<td><input type="checkbox" :value="tenant.id" @change="toggleRowSelection(tenant.id)"></td>
// 								<td>
// 									<div style="font-weight: bold;" x-text="tenant.name"></div>
// 									<div style="font-size: 12px; color: #6b7280;" x-text="tenant.contactInfo.companyName"></div>
// 								</td>
// 								<td x-text="tenant.slug"></td>
// 								<td x-text="tenant.contactInfo.email"></td>
// 								<td x-text="tenant.industry"></td>
// 								<td>
// 									<span :class="'status-' + tenant.status" x-text="tenant.status"></span>
// 								</td>
// 								<td>
// 									<div x-text="tenant.stats.userCount + ' users'"></div>
// 									<div style="font-size: 12px; color: #6b7280;" x-text="tenant.stats.activeUsers + ' active'"></div>
// 								</td>
// 								<td x-text="new Date(tenant.createdAt).toLocaleDateString()"></td>
// 								<td>
// 									<a :href="'/console/tenants/' + tenant.id" style="margin-right: 10px;">View</a>
// 									<a :href="'/console/tenants/' + tenant.id + '/edit'">Edit</a>
// 								</td>
// 							</tr>
// 						</template>
// 					</tbody>
// 				</table>
//
// 				<div class="pagination" style="margin-top: 20px; text-align: center;">
// 					<button @click="prevPage" :disabled="!hasPrevPage" class="btn">Previous</button>
// 					<span x-text="'Page ' + currentPage + ' of ' + totalPages" style="margin: 0 20px;"></span>
// 					<button @click="nextPage" :disabled="!hasNextPage" class="btn">Next</button>
// 				</div>
// 			</div>
//
// 			<script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
// 		</body>
// 		</html>
// 	`, data.Title, data.Title, h.tenantsToJSON(data.Tenants))
// }
//
// func (h *TenantHandler) renderTenantFormTemp(w http.ResponseWriter, data *types.ConsoleTenantFormData) {
// 	w.Header().Set("Content-Type", "text/html")
// 	fmt.Fprintf(w, `
// 		<!DOCTYPE html>
// 		<html>
// 		<head>
// 			<title>%s</title>
// 			<script src="/static/hooks/forms.js"></script>
// 			<style>
// 				body { font-family: Arial, sans-serif; margin: 40px; background: #f8fafc; }
// 				.form-container { max-width: 600px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
// 				.form-group { margin-bottom: 20px; }
// 				.form-group label { display: block; margin-bottom: 5px; font-weight: bold; color: #374151; }
// 				.form-group input, .form-group select { width: 100%%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; }
// 				.btn-primary { background: #1e40af; color: white; padding: 12px 24px; border: none; border-radius: 4px; cursor: pointer; }
// 				.btn-secondary { background: #6b7280; color: white; padding: 12px 24px; border: none; border-radius: 4px; cursor: pointer; margin-left: 10px; }
// 				.workflow-note { background: #dbeafe; border: 1px solid #93c5fd; color: #1e40af; padding: 12px; border-radius: 4px; margin-bottom: 20px; }
// 			</style>
// 		</head>
// 		<body>
// 			<div class="form-container">
// 				<h1>%s</h1>
//
// 				<div class="workflow-note">
// 					<strong>🔗 Temporal Workflow Integration:</strong> This form will trigger the tenant onboarding workflow which includes database setup, default configuration, and welcome email.
// 				</div>
//
// 				<form x-data="useFormValidation({
// 					initialData: {},
// 					rules: {
// 						name: { required: true, minLength: 2 },
// 						slug: { required: true, minLength: 2, pattern: /^[a-z0-9-]+$/ },
// 						contact_email: { required: true, email: true },
// 						company_name: { required: true, minLength: 2 }
// 					}
// 				})" @submit="submitForm($event, '%s')">
//
// 					<div class="form-group">
// 						<label for="name">Tenant Name</label>
// 						<input type="text" id="name" name="name"
// 							   x-model="formData.name"
// 							   @blur="handleFieldBlur('name')"
// 							   placeholder="Acme Corporation"
// 							   required>
// 						<div x-show="hasFieldError('name')"
// 							 x-text="getFirstFieldError('name')"
// 							 style="color: red; font-size: 12px;"></div>
// 					</div>
//
// 					<div class="form-group">
// 						<label for="slug">Tenant Slug</label>
// 						<input type="text" id="slug" name="slug"
// 							   x-model="formData.slug"
// 							   @blur="handleFieldBlur('slug')"
// 							   placeholder="acme-corp"
// 							   pattern="[a-z0-9-]+"
// 							   required>
// 						<div style="font-size: 12px; color: #6b7280;">Used in URLs (lowercase, numbers, hyphens only)</div>
// 						<div x-show="hasFieldError('slug')"
// 							 x-text="getFirstFieldError('slug')"
// 							 style="color: red; font-size: 12px;"></div>
// 					</div>
//
// 					<div class="form-group">
// 						<label for="company_name">Company Name</label>
// 						<input type="text" id="company_name" name="company_name"
// 							   x-model="formData.company_name"
// 							   @blur="handleFieldBlur('company_name')"
// 							   placeholder="Acme Corporation Inc."
// 							   required>
// 						<div x-show="hasFieldError('company_name')"
// 							 x-text="getFirstFieldError('company_name')"
// 							 style="color: red; font-size: 12px;"></div>
// 					</div>
//
// 					<div class="form-group">
// 						<label for="contact_email">Primary Contact Email</label>
// 						<input type="email" id="contact_email" name="contact_email"
// 							   x-model="formData.contact_email"
// 							   @blur="handleFieldBlur('contact_email')"
// 							   placeholder="admin@acme.com"
// 							   required>
// 						<div x-show="hasFieldError('contact_email')"
// 							 x-text="getFirstFieldError('contact_email')"
// 							 style="color: red; font-size: 12px;"></div>
// 					</div>
//
// 					<div class="form-group">
// 						<label for="contact_phone">Contact Phone</label>
// 						<input type="tel" id="contact_phone" name="contact_phone"
// 							   x-model="formData.contact_phone"
// 							   placeholder="+1-555-0123">
// 					</div>
//
// 					<div class="form-group">
// 						<label for="industry">Industry</label>
// 						<select id="industry" name="industry" x-model="formData.industry">
// 							<option value="">Select an industry</option>
// 							<option value="Technology">Technology</option>
// 							<option value="Healthcare">Healthcare</option>
// 							<option value="Finance">Finance</option>
// 							<option value="Manufacturing">Manufacturing</option>
// 							<option value="Retail">Retail</option>
// 							<option value="Education">Education</option>
// 							<option value="Other">Other</option>
// 						</select>
// 					</div>
//
// 					<input type="hidden" name="_token" value="%s">
//
// 					<div class="form-actions">
// 						<button type="submit" class="btn-primary" :disabled="!canSubmit">
// 							<span x-show="!isSubmitting">Create Tenant</span>
// 							<span x-show="isSubmitting">Creating via Temporal...</span>
// 						</button>
// 						<button type="button" class="btn-secondary" onclick="location.href='/console/tenants'">Cancel</button>
// 					</div>
// 				</form>
// 			</div>
//
// 			<script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
// 		</body>
// 		</html>
// 	`, data.Title, data.Title, data.Action, data.CSRFToken)
// }
//
// func (h *TenantHandler) renderTenantFormWithError(w http.ResponseWriter, tenantData *types.ConsoleTenantCreateRequest, errorMsg, csrfToken string) {
// 	w.Header().Set("Content-Type", "text/html")
// 	fmt.Fprintf(w, `
// 		<div style="background: #fee2e2; border: 1px solid #fecaca; color: #991b1b; padding: 12px; border-radius: 4px; margin-bottom: 20px;">
// 			Error: %s
// 		</div>
// 	`, errorMsg)
// 	// TODO: Re-render form with existing data
// }
//
// // Helper methods for temporary rendering
//
// func (h *TenantHandler) tenantsToJSON(tenants []types.ConsoleTenant) string {
// 	data, _ := json.Marshal(tenants)
// 	return string(data)
// }
//
// // SetupRoutes sets up console tenant management routes
// func (h *TenantHandler) SetupRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
// 	// Apply authentication middleware to all tenant routes
// 	mux.Handle("/console/tenants", authMiddleware(http.HandlerFunc(h.ListTenants)))
// 	mux.Handle("/console/api/tenants", authMiddleware(http.HandlerFunc(h.GetTenantsData)))
// 	mux.Handle("/console/tenants/new", authMiddleware(http.HandlerFunc(h.ShowCreateTenantForm)))
// 	mux.Handle("/console/tenants/bulk", authMiddleware(http.HandlerFunc(h.BulkTenantActions)))
// 	// Note: Tenant detail routes need path parameter handling in actual router
// }
//
