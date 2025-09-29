package console

//FIXME:
// import (
// 	"fmt"
// 	"net/http"
// 	"strconv"
//
// 	"github.com/google/uuid"
// 	"github.com/niiniyare/erp/internal/core/tenant"
// 	"github.com/niiniyare/erp/internal/shared/logger"
// 	"github.com/niiniyare/erp/internal/ui/middleware"
// 	"github.com/niiniyare/erp/internal/ui/services/console/templates"
// 	"github.com/niiniyare/erp/internal/ui/types"
// )
//
// // TenantHandler handles tenant management operations for admin console
// type TenantHandler struct {
// 	tenantService tenant.Service
// 	logger        logger.Logger
// }
//
// // NewTenantHandler creates a new tenant management handler
// func NewTenantHandler(
// 	tenantService tenant.Service,
// 	logger logger.Logger,
// ) *TenantHandler {
// 	return &TenantHandler{
// 		tenantService: tenantService,
// 		logger:        logger,
// 	}
// }
//
// // ListTenants displays the tenant management page with list of all tenants
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
// 	// Get pagination parameters
// 	page := 1
// 	limit := 20
// 	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
// 		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
// 			page = p
// 		}
// 	}
// 	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
// 		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
// 			limit = l
// 		}
// 	}
//
// 	offset := (page - 1) * limit
//
// 	// Get tenants from service
// 	tenants, err := h.tenantService.ListTenants(ctx, offset, limit)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to list tenants", logger.Fields{
// 			"error":   err.Error(),
// 			"user_id": uiCtx.UserID,
// 			"page":    page,
// 			"limit":   limit,
// 		})
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Convert to UI types
// 	uiTenants := make([]types.TenantItem, len(tenants))
// 	for i, t := range tenants {
// 		industry := ""
// 		if t.Industry != nil {
// 			industry = *t.Industry
// 		}
// 		companySize := ""
// 		if t.CompanySize != nil {
// 			companySize = *t.CompanySize
// 		}
//
// 		uiTenants[i] = types.TenantItem{
// 			ID:          t.ID.String(),
// 			Name:        t.Name,
// 			Email:       t.Email,
// 			Subdomain:   t.Subdomain,
// 			Status:      string(t.Status),
// 			Industry:    industry,
// 			CompanySize: companySize,
// 			CreatedAt:   t.CreatedAt,
// 			UpdatedAt:   t.UpdatedAt,
// 		}
// 	}
//
// 	// Prepare page data
// 	data := types.TenantsPageData{
// 		Title:     "Tenant Management",
// 		CSRFToken: uiCtx.CSRFToken,
// 		Tenants:   uiTenants,
// 		Pagination: types.Pagination{
// 			CurrentPage: page,
// 			Limit:       limit,
// 			Total:       len(uiTenants), // TODO: Get actual total count
// 			HasNext:     len(uiTenants) == limit,
// 			HasPrev:     page > 1,
// 		},
// 	}
//
// 	// Render template
// 	if err := console.TenantsPage(data).Render(ctx, w); err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to render tenants page", logger.Fields{
// 			"error":   err.Error(),
// 			"user_id": uiCtx.UserID,
// 		})
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	h.logger.InfoContext(ctx, "Tenants page accessed", logger.Fields{
// 		"user_id": uiCtx.UserID,
// 		"page":    page,
// 		"count":   len(uiTenants),
// 	})
// }
//
// // ShowCreateTenant displays the tenant creation form
// func (h *TenantHandler) ShowCreateTenant(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
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
// 	// Prepare form data
// 	data := types.CreateTenantPageData{
// 		Title:     "Create New Tenant",
// 		CSRFToken: uiCtx.CSRFToken,
// 		Form: types.TenantForm{
// 			Name:        "",
// 			Email:       "",
// 			Subdomain:   nil,
// 			Industry:    "",
// 			CompanySize: "",
// 		},
// 		Industries: []types.SelectOption{
// 			{Value: "technology", Label: "Technology"},
// 			{Value: "finance", Label: "Finance"},
// 			{Value: "healthcare", Label: "Healthcare"},
// 			{Value: "retail", Label: "Retail"},
// 			{Value: "manufacturing", Label: "Manufacturing"},
// 			{Value: "other", Label: "Other"},
// 		},
// 		CompanySizes: []types.SelectOption{
// 			{Value: "1-10", Label: "1-10 employees"},
// 			{Value: "11-50", Label: "11-50 employees"},
// 			{Value: "51-200", Label: "51-200 employees"},
// 			{Value: "201-500", Label: "201-500 employees"},
// 			{Value: "500+", Label: "500+ employees"},
// 		},
// 	}
//
// 	// Render template
// 	if err := console.CreateTenantPage(data).Render(ctx, w); err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to render create tenant page", logger.Fields{
// 			"error":   err.Error(),
// 			"user_id": uiCtx.UserID,
// 		})
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
// }
//
// // CreateTenant handles tenant creation form submission
// func (h *TenantHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Parse form
// 	if err := r.ParseForm(); err != nil {
// 		h.redirectWithError(w, r, "Invalid form data", "/console/tenants/create")
// 		return
// 	}
//
// 	// Extract form data
// 	name := r.FormValue("name")
// 	email := r.FormValue("email")
// 	subdomain := r.FormValue("subdomain")
// 	industry := r.FormValue("industry")
// 	companySize := r.FormValue("company_size")
//
// 	// Basic validation
// 	if name == "" || email == "" {
// 		h.redirectWithError(w, r, "Name and email are required", "/console/tenants/create")
// 		return
// 	}
//
// 	// Prepare tenant creation request
// 	req := tenant.CreateTenantRequest{
// 		Name:   name,
// 		Email:  email,
// 		Status: tenant.StatusActive,
// 	}
//
// 	// Set optional fields
// 	if industry != "" {
// 		req.Industry = &industry
// 	}
// 	if companySize != "" {
// 		req.CompanySize = &companySize
// 	}
//
// 	// Set subdomain if provided
// 	if subdomain != "" {
// 		req.Subdomain = &subdomain
// 	}
//
// 	// Create tenant
// 	newTenant, err := h.tenantService.CreateTenant(ctx, req)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to create tenant", logger.Fields{
// 			"error":   err.Error(),
// 			"user_id": uiCtx.UserID,
// 			"name":    name,
// 			"email":   email,
// 		})
// 		h.redirectWithError(w, r, "Failed to create tenant", "/console/tenants/create")
// 		return
// 	}
//
// 	h.logger.InfoContext(ctx, "Tenant created successfully", logger.Fields{
// 		"tenant_id":   newTenant.ID,
// 		"tenant_name": newTenant.Name,
// 		"user_id":     uiCtx.UserID,
// 	})
//
// 	// Redirect to tenant details or list
// 	redirectURL := fmt.Sprintf("/console/tenants/%s", newTenant.ID.String())
// 	if r.Header.Get("HX-Request") == "true" {
// 		w.Header().Set("HX-Redirect", redirectURL)
// 		w.WriteHeader(http.StatusOK)
// 		return
// 	}
//
// 	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
// }
//
// // ShowTenant displays details of a specific tenant
// func (h *TenantHandler) ShowTenant(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Extract tenant ID from URL path
// 	tenantIDStr := r.URL.Path[len("/console/tenants/"):]
// 	if tenantIDStr == "" {
// 		http.Error(w, "Tenant ID required", http.StatusBadRequest)
// 		return
// 	}
//
// 	tenantID, err := uuid.Parse(tenantIDStr)
// 	if err != nil {
// 		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
// 		return
// 	}
//
// 	// Get tenant
// 	tenantData, err := h.tenantService.GetTenantByID(ctx, tenantID)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to get tenant", logger.Fields{
// 			"error":     err.Error(),
// 			"tenant_id": tenantID,
// 			"user_id":   uiCtx.UserID,
// 		})
// 		http.Error(w, "Tenant not found", http.StatusNotFound)
// 		return
// 	}
//
// 	// Convert to UI type
// 	industry := ""
// 	if tenantData.Industry != nil {
// 		industry = *tenantData.Industry
// 	}
// 	companySize := ""
// 	if tenantData.CompanySize != nil {
// 		companySize = *tenantData.CompanySize
// 	}
//
// 	uiTenant := types.TenantDetail{
// 		ID:          tenantData.ID.String(),
// 		Name:        tenantData.Name,
// 		Email:       tenantData.Email,
// 		Subdomain:   tenantData.Subdomain,
// 		Status:      string(tenantData.Status),
// 		Industry:    industry,
// 		CompanySize: companySize,
// 		Timezone:    tenantData.Timezone,
// 		Currency:    tenantData.CurrencyCode,
// 		CreatedAt:   tenantData.CreatedAt,
// 		UpdatedAt:   tenantData.UpdatedAt,
// 	}
//
// 	// Prepare page data
// 	data := types.TenantDetailPageData{
// 		Title:     fmt.Sprintf("Tenant: %s", tenantData.Name),
// 		CSRFToken: uiCtx.CSRFToken,
// 		Tenant:    uiTenant,
// 	}
//
// 	// Render template
// 	if err := console.TenantDetailPage(data).Render(ctx, w); err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to render tenant detail page", logger.Fields{
// 			"error":     err.Error(),
// 			"tenant_id": tenantID,
// 			"user_id":   uiCtx.UserID,
// 		})
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
// }
//
// // UpdateTenantStatus handles tenant status changes (activate/suspend)
// func (h *TenantHandler) UpdateTenantStatus(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Parse form
// 	if err := r.ParseForm(); err != nil {
// 		http.Error(w, "Invalid form data", http.StatusBadRequest)
// 		return
// 	}
//
// 	tenantIDStr := r.FormValue("tenant_id")
// 	action := r.FormValue("action")
//
// 	tenantID, err := uuid.Parse(tenantIDStr)
// 	if err != nil {
// 		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
// 		return
// 	}
//
// 	// Perform action
// 	switch action {
// 	case "activate":
// 		err = h.tenantService.ActivateTenant(ctx, tenantID)
// 	case "suspend":
// 		err = h.tenantService.DeactivateTenant(ctx, tenantID)
// 	default:
// 		http.Error(w, "Invalid action", http.StatusBadRequest)
// 		return
// 	}
//
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to update tenant status", logger.Fields{
// 			"error":     err.Error(),
// 			"tenant_id": tenantID,
// 			"action":    action,
// 			"user_id":   uiCtx.UserID,
// 		})
// 		http.Error(w, "Failed to update tenant status", http.StatusInternalServerError)
// 		return
// 	}
//
// 	h.logger.InfoContext(ctx, "Tenant status updated", logger.Fields{
// 		"tenant_id": tenantID,
// 		"action":    action,
// 		"user_id":   uiCtx.UserID,
// 	})
//
// 	// Return success response
// 	if r.Header.Get("HX-Request") == "true" {
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte(`{"success": true}`))
// 		return
// 	}
//
// 	// Redirect back to tenant detail
// 	redirectURL := fmt.Sprintf("/console/tenants/%s", tenantID.String())
// 	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
// }
//
// // Helper method for redirecting with error
// func (h *TenantHandler) redirectWithError(w http.ResponseWriter, r *http.Request, errorMsg, redirectURL string) {
// 	finalURL := redirectURL + "?error=" + errorMsg
//
// 	if r.Header.Get("HX-Request") == "true" {
// 		w.Header().Set("HX-Redirect", finalURL)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
//
// 	http.Redirect(w, r, finalURL, http.StatusSeeOther)
// }
//
// // SetupRoutes sets up tenant management routes
// func (h *TenantHandler) SetupRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
// 	// Apply authentication middleware to all tenant routes
// 	mux.Handle("/console/tenants", authMiddleware(http.HandlerFunc(h.ListTenants)))
// 	mux.Handle("/console/tenants/", authMiddleware(http.HandlerFunc(h.RouteTenantRequest)))
// }
//
// // routeTenantRequest routes tenant-specific requests
// func (h *TenantHandler) RouteTenantRequest(w http.ResponseWriter, r *http.Request) {
// 	path := r.URL.Path
//
// 	if path == "/console/tenants/create" {
// 		if r.Method == http.MethodGet {
// 			h.ShowCreateTenant(w, r)
// 		} else if r.Method == http.MethodPost {
// 			h.CreateTenant(w, r)
// 		} else {
// 			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		}
// 		return
// 	}
//
// 	if path == "/console/tenants/status" && r.Method == http.MethodPost {
// 		h.UpdateTenantStatus(w, r)
// 		return
// 	}
//
// 	// Handle individual tenant details (e.g., /console/tenants/uuid)
// 	if len(path) > len("/console/tenants/") {
// 		h.ShowTenant(w, r)
// 		return
// 	}
//
// 	http.Error(w, "Not found", http.StatusNotFound)
// }
