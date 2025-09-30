package handlers

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"strconv"
// 	"time"
//
// 	"github.com/google/uuid"
// 	"github.com/niiniyare/erp/internal/core/abac"
// 	"github.com/niiniyare/erp/internal/core/audit"
// 	"github.com/niiniyare/erp/internal/core/iam"
// 	"github.com/niiniyare/erp/internal/core/tenant"
// 	"github.com/niiniyare/erp/internal/shared/logger"
// 	"github.com/niiniyare/erp/internal/ui/middleware"
// 	"github.com/niiniyare/erp/internal/ui/types"
// )
//
// // UsersHandler handles workspace user management operations
// type UsersHandler struct {
// 	iamService    iam.Service
// 	abacService   abac.Service
// 	tenantService tenant.Service
// 	auditService  audit.Service
// 	logger        logger.Logger
// }
//
// // NewUsersHandler creates a new users handler
// func NewUsersHandler(
// 	iamService iam.Service,
// 	abacService abac.Service,
// 	tenantService tenant.Service,
// 	auditService audit.Service,
// 	logger logger.Logger,
// ) *UsersHandler {
// 	return &UsersHandler{
// 		iamService:    iamService,
// 		abacService:   abacService,
// 		tenantService: tenantService,
// 		auditService:  auditService,
// 		logger:        logger,
// 	}
// }
//
// // ListUsers displays the user list page for the current tenant
// func (h *UsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Ensure user has tenant access and user:list permission
// 	if err := middleware.RequireRole(ctx, middleware.UIRoleTenantUser); err != nil {
// 		http.Error(w, "Forbidden", http.StatusForbidden)
// 		return
// 	}
//
// 	// Check ABAC permission for user listing
// 	hasPermission, err := h.checkPermission(ctx, uiCtx, "user", "list")
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to check user list permission", logger.Fields{
// 			"error":     err.Error(),
// 			"user_id":   uiCtx.UserID,
// 			"tenant_id": uiCtx.TenantID,
// 		})
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	if !hasPermission {
// 		http.Error(w, "Insufficient permissions to list users", http.StatusForbidden)
// 		return
// 	}
//
// 	// Parse query parameters
// 	filters := h.parseUserFilters(r)
//
// 	// Get users for this tenant
// 	users, pagination, err := h.getTenantUsers(ctx, uiCtx.TenantID, filters)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to get tenant users", logger.Fields{
// 			"error":     err.Error(),
// 			"user_id":   uiCtx.UserID,
// 			"tenant_id": uiCtx.TenantID,
// 		})
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Prepare page data
// 	pageData := &types.WorkspaceUsersPageData{
// 		Title:      "Team Members",
// 		Users:      users,
// 		Pagination: pagination,
// 		Filters:    filters,
// 		CSRFToken:  uiCtx.CSRFToken,
// 	}
//
// 	// Render user list template
// 	// TODO: Create workspace users template
// 	h.renderUsersListTemp(w, pageData)
//
// 	// Log user list access
// 	h.logger.InfoContext(ctx, "Workspace users list accessed", logger.Fields{
// 		"user_id":   uiCtx.UserID,
// 		"tenant_id": uiCtx.TenantID,
// 		"filters":   filters,
// 	})
// }
//
// // GetUsersData returns user list data as JSON (for HTMX requests and DataTable)
// func (h *UsersHandler) GetUsersData(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Check permissions
// 	hasPermission, err := h.checkPermission(ctx, uiCtx, "user", "list")
// 	if err != nil || !hasPermission {
// 		http.Error(w, "Forbidden", http.StatusForbidden)
// 		return
// 	}
//
// 	// Parse filters
// 	filters := h.parseUserFilters(r)
//
// 	// Get users data
// 	users, pagination, err := h.getTenantUsers(ctx, uiCtx.TenantID, filters)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to get users data", logger.Fields{
// 			"error":     err.Error(),
// 			"user_id":   uiCtx.UserID,
// 			"tenant_id": uiCtx.TenantID,
// 		})
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Return JSON response
// 	response := map[string]interface{}{
// 		"users":      users,
// 		"pagination": pagination,
// 		"success":    true,
// 	}
//
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(response)
// }
//
// // ShowCreateUserForm displays the user creation form
// func (h *UsersHandler) ShowCreateUserForm(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Check permissions
// 	hasPermission, err := h.checkPermission(ctx, uiCtx, "user", "create")
// 	if err != nil || !hasPermission {
// 		http.Error(w, "Forbidden", http.StatusForbidden)
// 		return
// 	}
//
// 	// Get available roles for this tenant
// 	roles := h.getAvailableRoles(ctx, uiCtx.TenantID)
//
// 	// Get departments for this tenant
// 	departments := h.getAvailableDepartments(ctx, uiCtx.TenantID)
//
// 	// Prepare form data
// 	formData := &types.WorkspaceUserFormData{
// 		Title:       "Create New User",
// 		Action:      "/workspace/users",
// 		Method:      "POST",
// 		Roles:       roles,
// 		Departments: departments,
// 		CSRFToken:   uiCtx.CSRFToken,
// 	}
//
// 	// Render user creation form
// 	// TODO: Create workspace user form template
// 	h.renderUserFormTemp(w, formData)
// }
//
// // CreateUser creates a new user in the current tenant
// func (h *UsersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Check permissions
// 	hasPermission, err := h.checkPermission(ctx, uiCtx, "user", "create")
// 	if err != nil || !hasPermission {
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
// 	// Extract user data
// 	userData := &types.WorkspaceUserCreateRequest{
// 		FirstName:  r.FormValue("first_name"),
// 		LastName:   r.FormValue("last_name"),
// 		Email:      r.FormValue("email"),
// 		Role:       r.FormValue("role"),
// 		Department: r.FormValue("department"),
// 		Phone:      r.FormValue("phone"),
// 		Title:      r.FormValue("job_title"),
// 	}
//
// 	// Validate user data
// 	if err := h.validateUserData(userData); err != nil {
// 		h.renderUserFormWithError(w, userData, err.Error(), uiCtx.CSRFToken)
// 		return
// 	}
//
// 	// Create user in IAM service
// 	userID, err := h.createTenantUser(ctx, uiCtx.TenantID, userData)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to create user", logger.Fields{
// 			"error":     err.Error(),
// 			"user_id":   uiCtx.UserID,
// 			"tenant_id": uiCtx.TenantID,
// 			"email":     userData.Email,
// 		})
// 		h.renderUserFormWithError(w, userData, "Failed to create user", uiCtx.CSRFToken)
// 		return
// 	}
//
// 	// Log user creation
// 	h.logger.InfoContext(ctx, "User created in workspace", logger.Fields{
// 		"created_user_id": userID,
// 		"creator_id":      uiCtx.UserID,
// 		"tenant_id":       uiCtx.TenantID,
// 		"email":           userData.Email,
// 	})
//
// 	// Redirect to user list with success message
// 	http.Redirect(w, r, "/workspace/users?success=user_created", http.StatusSeeOther)
// }
//
// // ShowUserDetail displays detailed information about a user
// func (h *UsersHandler) ShowUserDetail(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Check permissions
// 	hasPermission, err := h.checkPermission(ctx, uiCtx, "user", "read")
// 	if err != nil || !hasPermission {
// 		http.Error(w, "Forbidden", http.StatusForbidden)
// 		return
// 	}
//
// 	// Extract user ID from path
// 	userIDStr := r.URL.Path[len("/workspace/users/"):]
// 	if userIDStr == "" {
// 		http.Error(w, "User ID required", http.StatusBadRequest)
// 		return
// 	}
//
// 	userID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		http.Error(w, "Invalid user ID", http.StatusBadRequest)
// 		return
// 	}
//
// 	// Get user details
// 	user, err := h.getTenantUser(ctx, uiCtx.TenantID, userID)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to get user details", logger.Fields{
// 			"error":       err.Error(),
// 			"user_id":     uiCtx.UserID,
// 			"tenant_id":   uiCtx.TenantID,
// 			"target_user": userID,
// 		})
// 		http.Error(w, "User not found", http.StatusNotFound)
// 		return
// 	}
//
// 	// Prepare page data
// 	pageData := &types.WorkspaceUserDetailPageData{
// 		Title:     fmt.Sprintf("User Details - %s", user.Name),
// 		User:      user,
// 		CSRFToken: uiCtx.CSRFToken,
// 	}
//
// 	// Render user detail template
// 	// TODO: Create workspace user detail template
// 	h.renderUserDetailTemp(w, pageData)
// }
//
// // BulkUserActions handles bulk operations on users (activate, deactivate, etc.)
// func (h *UsersHandler) BulkUserActions(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
// 	userIDs := r.Form["ids[]"]
//
// 	if len(userIDs) == 0 {
// 		http.Error(w, "No users selected", http.StatusBadRequest)
// 		return
// 	}
//
// 	// Check permissions based on action
// 	var permission string
// 	switch action {
// 	case "activate", "deactivate":
// 		permission = "user:update"
// 	case "delete":
// 		permission = "user:delete"
// 	case "invite":
// 		permission = "user:invite"
// 	default:
// 		http.Error(w, "Invalid action", http.StatusBadRequest)
// 		return
// 	}
//
// 	hasPermission, err := h.checkPermission(ctx, uiCtx, "user", permission)
// 	if err != nil || !hasPermission {
// 		http.Error(w, "Forbidden", http.StatusForbidden)
// 		return
// 	}
//
// 	// Process bulk action
// 	results, err := h.processBulkUserAction(ctx, uiCtx.TenantID, action, userIDs)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to process bulk user action", logger.Fields{
// 			"error":     err.Error(),
// 			"user_id":   uiCtx.UserID,
// 			"tenant_id": uiCtx.TenantID,
// 			"action":    action,
// 			"count":     len(userIDs),
// 		})
// 		http.Error(w, "Failed to process bulk action", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Log bulk action
// 	h.logger.InfoContext(ctx, "Bulk user action processed", logger.Fields{
// 		"user_id":       uiCtx.UserID,
// 		"tenant_id":     uiCtx.TenantID,
// 		"action":        action,
// 		"count":         len(userIDs),
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
// // parseUserFilters extracts user filtering parameters from request
// func (h *UsersHandler) parseUserFilters(r *http.Request) *types.WorkspaceUserFilters {
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
// 	return &types.WorkspaceUserFilters{
// 		Page:       page,
// 		PageSize:   pageSize,
// 		Search:     r.URL.Query().Get("q"),
// 		Role:       r.URL.Query().Get("role"),
// 		Department: r.URL.Query().Get("department"),
// 		Status:     r.URL.Query().Get("status"),
// 	}
// }
//
// // getTenantUsers retrieves users for the current tenant with filtering
// func (h *UsersHandler) getTenantUsers(ctx context.Context, tenantID uuid.UUID, filters *types.WorkspaceUserFilters) ([]types.WorkspaceUser, *types.PaginationMeta, error) {
// 	// TODO: Implement actual user querying with tenant filtering
// 	// For now, return placeholder data
//
// 	users := []types.WorkspaceUser{}
// 	for i := 1; i <= 15; i++ {
// 		user := types.WorkspaceUser{
// 			ID:         uuid.New().String(),
// 			Name:       fmt.Sprintf("User %d", i),
// 			Email:      fmt.Sprintf("user%d@workspace.com", i),
// 			Role:       []string{"user", "manager", "admin"}[i%3],
// 			Department: []string{"Engineering", "Sales", "Marketing", "HR"}[i%4],
// 			Status:     []string{"active", "inactive", "pending"}[i%3],
// 			LastLogin:  time.Now().Add(-time.Duration(i) * time.Hour),
// 			CreatedAt:  time.Now().Add(-time.Duration(i*24) * time.Hour),
// 		}
// 		users = append(users, user)
// 	}
//
// 	// Apply filters
// 	filteredUsers := h.applyUserFilters(users, filters)
//
// 	// Calculate pagination
// 	total := len(filteredUsers)
// 	start := (filters.Page - 1) * filters.PageSize
// 	end := start + filters.PageSize
//
// 	if start > total {
// 		start = total
// 	}
// 	if end > total {
// 		end = total
// 	}
//
// 	paginatedUsers := filteredUsers[start:end]
//
// 	pagination := &types.PaginationMeta{
// 		Page:       filters.Page,
// 		PageSize:   filters.PageSize,
// 		Total:      total,
// 		TotalPages: (total + filters.PageSize - 1) / filters.PageSize,
// 		HasNext:    filters.Page < (total+filters.PageSize-1)/filters.PageSize,
// 		HasPrev:    filters.Page > 1,
// 	}
//
// 	return paginatedUsers, pagination, nil
// }
//
// // applyUserFilters applies filtering logic to users
// func (h *UsersHandler) applyUserFilters(users []types.WorkspaceUser, filters *types.WorkspaceUserFilters) []types.WorkspaceUser {
// 	filtered := make([]types.WorkspaceUser, 0)
//
// 	for _, user := range users {
// 		// Search filter
// 		if filters.Search != "" {
// 			if !h.userMatchesSearch(user, filters.Search) {
// 				continue
// 			}
// 		}
//
// 		// Role filter
// 		if filters.Role != "" && filters.Role != "all" && user.Role != filters.Role {
// 			continue
// 		}
//
// 		// Department filter
// 		if filters.Department != "" && filters.Department != "all" && user.Department != filters.Department {
// 			continue
// 		}
//
// 		// Status filter
// 		if filters.Status != "" && filters.Status != "all" && user.Status != filters.Status {
// 			continue
// 		}
//
// 		filtered = append(filtered, user)
// 	}
//
// 	return filtered
// }
//
// // userMatchesSearch checks if user matches search query
// func (h *UsersHandler) userMatchesSearch(user types.WorkspaceUser, search string) bool {
// 	search = fmt.Sprintf("%s %s %s", user.Name, user.Email, user.Department)
// 	return len(search) > 0 // Simplified search logic
// }
//
// // checkPermission checks ABAC permissions for the user
// func (h *UsersHandler) checkPermission(ctx context.Context, uiCtx *middleware.UIContext, resource, action string) (bool, error) {
// 	// TODO: Implement proper ABAC permission checking
// 	// For now, return true for demo purposes
// 	return true, nil
// }
//
// // validateUserData validates user creation/update data
// func (h *UsersHandler) validateUserData(userData *types.WorkspaceUserCreateRequest) error {
// 	if userData.FirstName == "" {
// 		return fmt.Errorf("first name is required")
// 	}
// 	if userData.LastName == "" {
// 		return fmt.Errorf("last name is required")
// 	}
// 	if userData.Email == "" {
// 		return fmt.Errorf("email is required")
// 	}
// 	if userData.Role == "" {
// 		return fmt.Errorf("role is required")
// 	}
// 	// TODO: Add more comprehensive validation
// 	return nil
// }
//
// // Temporary rendering methods (until templates are created)
//
// func (h *UsersHandler) renderUsersListTemp(w http.ResponseWriter, data *types.WorkspaceUsersPageData) {
// 	w.Header().Set("Content-Type", "text/html")
// 	fmt.Fprintf(w, `
// 		<!DOCTYPE html>
// 		<html>
// 		<head>
// 			<title>%s</title>
// 			<script src="/static/hooks/datatable.js"></script>
// 			<style>
// 				body { font-family: Arial, sans-serif; margin: 40px; }
// 				.header { background: linear-gradient(135deg, #10b981, #059669); color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
// 				.actions { margin: 20px 0; }
// 				.btn { background: #10b981; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; margin-right: 10px; }
// 				.table { width: 100%%; border-collapse: collapse; }
// 				.table th, .table td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
// 				.table th { background: #f8fafc; }
// 				.status-active { color: #059669; }
// 				.status-inactive { color: #dc2626; }
// 				.status-pending { color: #d97706; }
// 			</style>
// 		</head>
// 		<body>
// 			<div class="header">
// 				<h1>%s</h1>
// 				<p>Manage your team members and their roles</p>
// 			</div>
//
// 			<div class="actions">
// 				<button class="btn" onclick="location.href='/workspace/users/new'">Add New User</button>
// 				<button class="btn" id="bulk-invite">Send Invites</button>
// 				<button class="btn" id="bulk-deactivate">Deactivate Selected</button>
// 			</div>
//
// 			<div x-data="useDataTable({
// 				rows: %s,
// 				searchColumns: ['name', 'email', 'department', 'role'],
// 				serviceContext: 'workspace'
// 			})">
// 				<div class="filters" style="margin-bottom: 20px;">
// 					<input type="text" x-model="searchQuery" @input="handleSearch" placeholder="Search users..." style="padding: 8px; margin-right: 10px;">
// 					<select @change="applyFilter" style="padding: 8px;">
// 						<option value="all">All Roles</option>
// 						<option value="admin">Admin</option>
// 						<option value="manager">Manager</option>
// 						<option value="user">User</option>
// 					</select>
// 				</div>
//
// 				<table class="table">
// 					<thead>
// 						<tr>
// 							<th><input type="checkbox" @change="toggleSelectAll"></th>
// 							<th @click="sort('name')">Name</th>
// 							<th @click="sort('email')">Email</th>
// 							<th @click="sort('role')">Role</th>
// 							<th @click="sort('department')">Department</th>
// 							<th @click="sort('status')">Status</th>
// 							<th>Last Login</th>
// 							<th>Actions</th>
// 						</tr>
// 					</thead>
// 					<tbody>
// 						<template x-for="user in paginatedRows" :key="user.id">
// 							<tr>
// 								<td><input type="checkbox" :value="user.id" @change="toggleRowSelection(user.id)"></td>
// 								<td x-text="user.name"></td>
// 								<td x-text="user.email"></td>
// 								<td x-text="user.role"></td>
// 								<td x-text="user.department"></td>
// 								<td>
// 									<span :class="'status-' + user.status" x-text="user.status"></span>
// 								</td>
// 								<td x-text="new Date(user.last_login).toLocaleDateString()"></td>
// 								<td>
// 									<a :href="'/workspace/users/' + user.id" style="margin-right: 10px;">View</a>
// 									<a :href="'/workspace/users/' + user.id + '/edit'">Edit</a>
// 								</td>
// 							</tr>
// 						</template>
// 					</tbody>
// 				</table>
//
// 				<div class="pagination" style="margin-top: 20px;">
// 					<button @click="prevPage" :disabled="!hasPrevPage">Previous</button>
// 					<span x-text="'Page ' + currentPage + ' of ' + totalPages"></span>
// 					<button @click="nextPage" :disabled="!hasNextPage">Next</button>
// 				</div>
// 			</div>
//
// 			<script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
// 		</body>
// 		</html>
// 	`, data.Title, data.Title, h.usersToJSON(data.Users))
// }
//
// func (h *UsersHandler) renderUserFormTemp(w http.ResponseWriter, data *types.WorkspaceUserFormData) {
// 	w.Header().Set("Content-Type", "text/html")
// 	fmt.Fprintf(w, `
// 		<!DOCTYPE html>
// 		<html>
// 		<head>
// 			<title>%s</title>
// 			<script src="/static/hooks/forms.js"></script>
// 			<style>
// 				body { font-family: Arial, sans-serif; margin: 40px; }
// 				.form-container { max-width: 600px; margin: 0 auto; }
// 				.form-group { margin-bottom: 20px; }
// 				.form-group label { display: block; margin-bottom: 5px; font-weight: bold; }
// 				.form-group input, .form-group select { width: 100%%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; }
// 				.btn-primary { background: #10b981; color: white; padding: 12px 24px; border: none; border-radius: 4px; cursor: pointer; }
// 				.btn-secondary { background: #6b7280; color: white; padding: 12px 24px; border: none; border-radius: 4px; cursor: pointer; margin-left: 10px; }
// 			</style>
// 		</head>
// 		<body>
// 			<div class="form-container">
// 				<h1>%s</h1>
//
// 				<form x-data="useFormValidation({
// 					initialData: {},
// 					rules: {
// 						first_name: { required: true, minLength: 2 },
// 						last_name: { required: true, minLength: 2 },
// 						email: { required: true, email: true },
// 						role: { required: true }
// 					}
// 				})" @submit="submitForm($event, '%s')">
//
// 					<div class="form-group">
// 						<label for="first_name">First Name</label>
// 						<input type="text" id="first_name" name="first_name"
// 							   x-model="formData.first_name"
// 							   @blur="handleFieldBlur('first_name')"
// 							   required>
// 						<div x-show="hasFieldError('first_name')"
// 							 x-text="getFirstFieldError('first_name')"
// 							 style="color: red; font-size: 12px;"></div>
// 					</div>
//
// 					<div class="form-group">
// 						<label for="last_name">Last Name</label>
// 						<input type="text" id="last_name" name="last_name"
// 							   x-model="formData.last_name"
// 							   @blur="handleFieldBlur('last_name')"
// 							   required>
// 						<div x-show="hasFieldError('last_name')"
// 							 x-text="getFirstFieldError('last_name')"
// 							 style="color: red; font-size: 12px;"></div>
// 					</div>
//
// 					<div class="form-group">
// 						<label for="email">Email</label>
// 						<input type="email" id="email" name="email"
// 							   x-model="formData.email"
// 							   @blur="handleFieldBlur('email')"
// 							   required>
// 						<div x-show="hasFieldError('email')"
// 							 x-text="getFirstFieldError('email')"
// 							 style="color: red; font-size: 12px;"></div>
// 					</div>
//
// 					<div class="form-group">
// 						<label for="role">Role</label>
// 						<select id="role" name="role"
// 								x-model="formData.role"
// 								@change="handleFieldBlur('role')"
// 								required>
// 							<option value="">Select a role</option>
// 							<option value="user">User</option>
// 							<option value="manager">Manager</option>
// 							<option value="admin">Admin</option>
// 						</select>
// 						<div x-show="hasFieldError('role')"
// 							 x-text="getFirstFieldError('role')"
// 							 style="color: red; font-size: 12px;"></div>
// 					</div>
//
// 					<div class="form-group">
// 						<label for="department">Department</label>
// 						<select id="department" name="department" x-model="formData.department">
// 							<option value="">Select a department</option>
// 							<option value="Engineering">Engineering</option>
// 							<option value="Sales">Sales</option>
// 							<option value="Marketing">Marketing</option>
// 							<option value="HR">Human Resources</option>
// 							<option value="Finance">Finance</option>
// 						</select>
// 					</div>
//
// 					<div class="form-group">
// 						<label for="job_title">Job Title</label>
// 						<input type="text" id="job_title" name="job_title" x-model="formData.job_title">
// 					</div>
//
// 					<div class="form-group">
// 						<label for="phone">Phone</label>
// 						<input type="tel" id="phone" name="phone" x-model="formData.phone">
// 					</div>
//
// 					<input type="hidden" name="_token" value="%s">
//
// 					<div class="form-actions">
// 						<button type="submit" class="btn-primary" :disabled="!canSubmit">
// 							<span x-show="!isSubmitting">Create User</span>
// 							<span x-show="isSubmitting">Creating...</span>
// 						</button>
// 						<button type="button" class="btn-secondary" onclick="location.href='/workspace/users'">Cancel</button>
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
// func (h *UsersHandler) renderUserDetailTemp(w http.ResponseWriter, data *types.WorkspaceUserDetailPageData) {
// 	w.Header().Set("Content-Type", "text/html")
// 	fmt.Fprintf(w, `
// 		<!DOCTYPE html>
// 		<html>
// 		<head>
// 			<title>%s</title>
// 			<style>
// 				body { font-family: Arial, sans-serif; margin: 40px; }
// 				.user-card { background: white; border: 1px solid #e2e8f0; border-radius: 8px; padding: 24px; margin-bottom: 20px; }
// 				.user-avatar { width: 80px; height: 80px; border-radius: 50%%; background: #10b981; color: white; display: flex; align-items: center; justify-content: center; font-size: 24px; font-weight: bold; }
// 				.user-info { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-top: 20px; }
// 				.info-item { }
// 				.info-label { font-weight: bold; color: #374151; margin-bottom: 4px; }
// 				.info-value { color: #6b7280; }
// 				.actions { margin-top: 20px; }
// 				.btn { background: #10b981; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; margin-right: 10px; text-decoration: none; display: inline-block; }
// 			</style>
// 		</head>
// 		<body>
// 			<div class="user-card">
// 				<div style="display: flex; align-items: center; margin-bottom: 20px;">
// 					<div class="user-avatar">%s</div>
// 					<div style="margin-left: 20px;">
// 						<h1>%s</h1>
// 						<p style="color: #6b7280; margin: 0;">%s • %s</p>
// 					</div>
// 				</div>
//
// 				<div class="user-info">
// 					<div class="info-item">
// 						<div class="info-label">Email</div>
// 						<div class="info-value">%s</div>
// 					</div>
// 					<div class="info-item">
// 						<div class="info-label">Role</div>
// 						<div class="info-value">%s</div>
// 					</div>
// 					<div class="info-item">
// 						<div class="info-label">Department</div>
// 						<div class="info-value">%s</div>
// 					</div>
// 					<div class="info-item">
// 						<div class="info-label">Status</div>
// 						<div class="info-value">%s</div>
// 					</div>
// 					<div class="info-item">
// 						<div class="info-label">Last Login</div>
// 						<div class="info-value">%s</div>
// 					</div>
// 					<div class="info-item">
// 						<div class="info-label">Member Since</div>
// 						<div class="info-value">%s</div>
// 					</div>
// 				</div>
//
// 				<div class="actions">
// 					<a href="/workspace/users/%s/edit" class="btn">Edit User</a>
// 					<a href="/workspace/users" class="btn" style="background: #6b7280;">Back to Users</a>
// 				</div>
// 			</div>
// 		</body>
// 		</html>
// 	`,
// 		data.Title,
// 		string(data.User.Name[0]), // First letter for avatar
// 		data.User.Name,
// 		data.User.Role,
// 		data.User.Department,
// 		data.User.Email,
// 		data.User.Role,
// 		data.User.Department,
// 		data.User.Status,
// 		data.User.LastLogin.Format("Jan 2, 2006 3:04 PM"),
// 		data.User.CreatedAt.Format("Jan 2, 2006"),
// 		data.User.ID,
// 	)
// }
//
// func (h *UsersHandler) renderUserFormWithError(w http.ResponseWriter, userData *types.WorkspaceUserCreateRequest, errorMsg, csrfToken string) {
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
// func (h *UsersHandler) usersToJSON(users []types.WorkspaceUser) string {
// 	data, _ := json.Marshal(users)
// 	return string(data)
// }
//
// func (h *UsersHandler) getAvailableRoles(ctx context.Context, tenantID uuid.UUID) []string {
// 	return []string{"user", "manager", "admin"}
// }
//
// func (h *UsersHandler) getAvailableDepartments(ctx context.Context, tenantID uuid.UUID) []string {
// 	return []string{"Engineering", "Sales", "Marketing", "HR", "Finance"}
// }
//
// func (h *UsersHandler) createTenantUser(ctx context.Context, tenantID uuid.UUID, userData *types.WorkspaceUserCreateRequest) (uuid.UUID, error) {
// 	// TODO: Implement actual user creation via IAM service
// 	return uuid.New(), nil
// }
//
// func (h *UsersHandler) getTenantUser(ctx context.Context, tenantID, userID uuid.UUID) (*types.WorkspaceUser, error) {
// 	// TODO: Implement actual user retrieval
// 	return &types.WorkspaceUser{
// 		ID:         userID.String(),
// 		Name:       "John Doe",
// 		Email:      "john.doe@workspace.com",
// 		Role:       "manager",
// 		Department: "Engineering",
// 		Status:     "active",
// 		LastLogin:  time.Now().Add(-2 * time.Hour),
// 		CreatedAt:  time.Now().Add(-30 * 24 * time.Hour),
// 	}, nil
// }
//
// func (h *UsersHandler) processBulkUserAction(ctx context.Context, tenantID uuid.UUID, action string, userIDs []string) (*types.BulkActionResults, error) {
// 	// TODO: Implement actual bulk operations
// 	return &types.BulkActionResults{
// 		SuccessCount: len(userIDs),
// 		ErrorCount:   0,
// 		Errors:       []string{},
// 	}, nil
// }
//
// // SetupRoutes sets up workspace user management routes
// func (h *UsersHandler) SetupRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
// 	// Apply authentication middleware to all user routes
// 	mux.Handle("/workspace/users", authMiddleware(http.HandlerFunc(h.ListUsers)))
// 	mux.Handle("/workspace/api/users", authMiddleware(http.HandlerFunc(h.GetUsersData)))
// 	mux.Handle("/workspace/users/new", authMiddleware(http.HandlerFunc(h.ShowCreateUserForm)))
// 	mux.Handle("/workspace/users/bulk", authMiddleware(http.HandlerFunc(h.BulkUserActions)))
// 	// Note: User detail routes need path parameter handling in actual router
// }
//
