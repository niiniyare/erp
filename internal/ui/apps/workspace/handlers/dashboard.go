package workspace

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/abac"
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/ui/middleware"
	"github.com/niiniyare/erp/internal/ui/types"
)

// DashboardHandler handles workspace dashboard operations
type DashboardHandler struct {
	iamService    iam.Service
	abacService   abac.Service
	tenantService tenant.Service
	auditService  audit.Service
	logger        logger.Logger
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(
	iamService iam.Service,
	abacService abac.Service,
	tenantService tenant.Service,
	auditService audit.Service,
	logger logger.Logger,
) *DashboardHandler {
	return &DashboardHandler{
		iamService:    iamService,
		abacService:   abacService,
		tenantService: tenantService,
		auditService:  auditService,
		logger:        logger,
	}
}

// ShowDashboard displays the workspace dashboard for current tenant
func (h *DashboardHandler) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get UI context (user should be authenticated by middleware)
	uiCtx, ok := middleware.GetUIContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Ensure user has tenant access role
	if err := middleware.RequireRole(ctx, middleware.UIRoleTenantUser); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Verify user belongs to a tenant
	if uiCtx.TenantID == uuid.Nil {
		http.Error(w, "No tenant context", http.StatusForbidden)
		return
	}

	// Gather dashboard data for this tenant
	data, err := h.gatherDashboardData(ctx, uiCtx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to gather workspace dashboard data", logger.Fields{
			"error":     err.Error(),
			"user_id":   uiCtx.UserID,
			"tenant_id": uiCtx.TenantID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Add CSRF token
	data.CSRFToken = uiCtx.CSRFToken

	// Render dashboard template
	// TODO: Create workspace dashboard template
	// if err := workspace.DashboardPage(*data).Render(ctx, w); err != nil {
	//     h.logger.ErrorContext(ctx, "Failed to render workspace dashboard", logger.Fields{
	//         "error":     err.Error(),
	//         "user_id":   uiCtx.UserID,
	//         "tenant_id": uiCtx.TenantID,
	//     })
	//     http.Error(w, "Internal server error", http.StatusInternalServerError)
	//     return
	// }

	// Temporary response until templates are created
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Workspace Dashboard</title>
			<style>
				body { font-family: Arial, sans-serif; margin: 40px; }
				.header { background: linear-gradient(135deg, #10b981, #059669); color: white; padding: 20px; border-radius: 8px; }
				.stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin: 20px 0; }
				.stat-card { background: #f8fafc; border: 1px solid #e2e8f0; padding: 20px; border-radius: 8px; }
				.stat-value { font-size: 2em; font-weight: bold; color: #10b981; }
			</style>
		</head>
		<body>
			<div class="header">
				<h1>Workspace Dashboard</h1>
				<p>Welcome, %s - Tenant: %s</p>
			</div>
			<div class="stats">
				<div class="stat-card">
					<div class="stat-value">%d</div>
					<div>Active Users</div>
				</div>
				<div class="stat-card">
					<div class="stat-value">%d</div>
					<div>Active Projects</div>
				</div>
				<div class="stat-card">
					<div class="stat-value">%d</div>
					<div>Open Invoices</div>
				</div>
				<div class="stat-card">
					<div class="stat-value">%.2f</div>
					<div>Monthly Revenue</div>
				</div>
			</div>
			<p>This is a placeholder for the Workspace dashboard. Templates will be implemented next.</p>
		</body>
		</html>
	`,
		data.User.Name,
		data.TenantInfo.Name,
		data.TenantStats.ActiveUsers,
		data.TenantStats.ActiveProjects,
		data.TenantStats.OpenInvoices,
		data.TenantStats.MonthlyRevenue,
	)

	// Log dashboard access
	h.logger.InfoContext(ctx, "Workspace dashboard accessed", logger.Fields{
		"user_id":   uiCtx.UserID,
		"tenant_id": uiCtx.TenantID,
	})
}

// GetDashboardData returns dashboard data as JSON (for HTMX requests)
func (h *DashboardHandler) GetDashboardData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get UI context
	uiCtx, ok := middleware.GetUIContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Ensure tenant access
	if err := middleware.RequireRole(ctx, middleware.UIRoleTenantUser); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get tenant stats
	stats, err := h.getTenantStats(ctx, uiCtx.TenantID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get tenant stats", logger.Fields{
			"error":     err.Error(),
			"user_id":   uiCtx.UserID,
			"tenant_id": uiCtx.TenantID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Render stats as JSON
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"activeUsers": %d,
		"activeProjects": %d,
		"openInvoices": %d,
		"monthlyRevenue": %.2f,
		"lastUpdated": "%s"
	}`,
		stats.ActiveUsers,
		stats.ActiveProjects,
		stats.OpenInvoices,
		stats.MonthlyRevenue,
		time.Now().Format(time.RFC3339),
	)
}

// GetRecentActivity returns recent tenant activity as HTML (for HTMX requests)
func (h *DashboardHandler) GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get UI context
	uiCtx, ok := middleware.GetUIContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get recent activity for this tenant
	activities, err := h.getRecentActivity(ctx, uiCtx.TenantID, 20)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get recent activity", logger.Fields{
			"error":     err.Error(),
			"user_id":   uiCtx.UserID,
			"tenant_id": uiCtx.TenantID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Render activity component as HTML
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<div class='activity-list'>"))
	for _, activity := range activities {
		fmt.Fprintf(w, `
			<div class="activity-item">
				<div class="activity-icon">📝</div>
				<div class="activity-content">
					<div class="activity-title">%s</div>
					<div class="activity-user">by %s</div>
					<div class="activity-time">%s</div>
				</div>
			</div>
		`, activity.Description, activity.UserName, activity.Timestamp.Format("Jan 2, 3:04 PM"))
	}
	w.Write([]byte("</div>"))
}

// Private helper methods

// gatherDashboardData collects all data needed for the workspace dashboard
func (h *DashboardHandler) gatherDashboardData(ctx context.Context, uiCtx *middleware.UIContext) (*types.WorkspaceDashboardData, error) {
	// Get user information
	userInfo, err := h.getUserInfo(ctx, uiCtx.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Get tenant information
	tenantInfo, err := h.getTenantInfo(ctx, uiCtx.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant info: %w", err)
	}

	// Get tenant statistics
	tenantStats, err := h.getTenantStats(ctx, uiCtx.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant stats: %w", err)
	}

	// Get recent activity
	recentActivity, err := h.getRecentActivity(ctx, uiCtx.TenantID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	// Get quick actions for workspace
	quickActions := h.getWorkspaceQuickActions(ctx, uiCtx)

	// Get notifications for this user/tenant
	notifications, err := h.getNotifications(ctx, uiCtx.UserID, uiCtx.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	return &types.WorkspaceDashboardData{
		Title:          "Workspace Dashboard",
		User:           *userInfo,
		TenantInfo:     *tenantInfo,
		TenantStats:    *tenantStats,
		RecentActivity: recentActivity,
		QuickActions:   quickActions,
		Notifications:  notifications,
	}, nil
}

// getUserInfo gets information about the current user
func (h *DashboardHandler) getUserInfo(ctx context.Context, userID uuid.UUID) (*types.UserInfo, error) {
	// Get user from IAM service
	user, err := h.iamService.Authentication().GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &types.UserInfo{
		ID:       user.ID.String(),
		Name:     user.FirstName + " " + user.LastName,
		Email:    user.Email,
		Role:     "Tenant User",
		TenantID: user.TenantID.String(),
	}, nil
}

// getTenantInfo gets information about the current tenant
func (h *DashboardHandler) getTenantInfo(ctx context.Context, tenantID uuid.UUID) (*types.TenantInfo, error) {
	// Get tenant from tenant service
	tenant, err := h.tenantService.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return &types.TenantInfo{
		ID:        tenant.ID.String(),
		Name:      tenant.Name,
		Domain:    *tenant.Subdomain,
		Status:    string(tenant.Status),
		Plan:      "",
		UserCount: 0, // TODO: Get actual user count
		CreatedAt: tenant.CreatedAt,
	}, nil
}

// getTenantStats collects tenant-specific statistics
func (h *DashboardHandler) getTenantStats(ctx context.Context, tenantID uuid.UUID) (*types.TenantStats, error) {
	// For now, return placeholder stats
	// TODO: Implement actual queries to get tenant-specific statistics
	return &types.TenantStats{
		ActiveUsers:    25,       // TODO: Query user service
		ActiveProjects: 12,       // TODO: Query project service
		OpenInvoices:   8,        // TODO: Query finance service
		MonthlyRevenue: 45650.75, // TODO: Query finance service
		UpcomingTasks:  34,       // TODO: Query task/project service
		TeamMembers:    25,       // TODO: Query user service
		LastMonth: &types.TenantStatsComparison{
			Users:    2,
			Projects: 1,
			Revenue:  5430.25,
		},
	}, nil
}

// getRecentActivity gets recent activity for this tenant
func (h *DashboardHandler) getRecentActivity(ctx context.Context, tenantID uuid.UUID, limit int) ([]types.ActivityItem, error) {
	// For now, create placeholder activities
	// TODO: Implement proper audit log retrieval with tenant context
	activities := make([]types.ActivityItem, limit)
	activityTypes := []string{"user_login", "project_created", "invoice_sent", "task_completed", "document_uploaded"}
	userNames := []string{"John Doe", "Jane Smith", "Bob Johnson", "Alice Williams", "Charlie Brown"}

	for i := 0; i < limit; i++ {
		activityType := activityTypes[i%len(activityTypes)]
		userName := userNames[i%len(userNames)]

		description := fmt.Sprintf("%s performed %s", userName, activityType)
		switch activityType {
		case "user_login":
			description = fmt.Sprintf("%s logged into the workspace", userName)
		case "project_created":
			description = fmt.Sprintf("%s created a new project", userName)
		case "invoice_sent":
			description = fmt.Sprintf("%s sent invoice #INV-%d", userName, 1000+i)
		case "task_completed":
			description = fmt.Sprintf("%s completed a task", userName)
		case "document_uploaded":
			description = fmt.Sprintf("%s uploaded a document", userName)
		}

		activities[i] = types.ActivityItem{
			ID:          uuid.New().String(),
			Type:        activityType,
			Description: description,
			UserID:      uuid.New().String(),
			UserName:    userName,
			Timestamp:   time.Now().Add(-time.Duration(i*15) * time.Minute),
			Severity:    "info",
			TenantID:    tenantID.String(),
		}
	}

	return activities, nil
}

// getWorkspaceQuickActions returns available quick actions for workspace users
func (h *DashboardHandler) getWorkspaceQuickActions(ctx context.Context, uiCtx *middleware.UIContext) []types.QuickAction {
	actions := []types.QuickAction{
		{
			Title:       "Team Members",
			Description: "Manage team members and user roles",
			URL:         "/workspace/users",
			Icon:        "users",
			Permission:  "user:list",
			Color:       "blue",
		},
		{
			Title:       "Projects",
			Description: "View and manage active projects",
			URL:         "/workspace/projects",
			Icon:        "folder",
			Permission:  "project:list",
			Color:       "green",
		},
		{
			Title:       "Invoices",
			Description: "Create and manage invoices",
			URL:         "/workspace/finance/invoices",
			Icon:        "receipt",
			Permission:  "finance:read",
			Color:       "yellow",
		},
		{
			Title:       "Reports",
			Description: "View business reports and analytics",
			URL:         "/workspace/reports",
			Icon:        "chart-bar",
			Permission:  "report:view",
			Color:       "purple",
		},
		{
			Title:       "Settings",
			Description: "Configure workspace settings",
			URL:         "/workspace/settings",
			Icon:        "cog",
			Permission:  "tenant:settings",
			Color:       "gray",
		},
	}

	// TODO: Filter actions based on ABAC permissions
	return actions
}

// getNotifications gets notifications for the user in this tenant
func (h *DashboardHandler) getNotifications(ctx context.Context, userID, tenantID uuid.UUID) ([]types.Notification, error) {
	// Placeholder implementation - tenant-scoped notifications
	notifications := []types.Notification{
		{
			ID:       uuid.New().String(),
			Type:     "project",
			Title:    "Project Update",
			Message:  "Project Alpha milestone completed successfully",
			Severity: "info",
			Created:  time.Now().Add(-2 * time.Hour),
			Read:     false,
			TenantID: tenantID.String(),
		},
		{
			ID:       uuid.New().String(),
			Type:     "finance",
			Title:    "Invoice Overdue",
			Message:  "Invoice #INV-2023-015 is now 5 days overdue",
			Severity: "warning",
			Created:  time.Now().Add(-1 * time.Hour),
			Read:     false,
			TenantID: tenantID.String(),
		},
		{
			ID:       uuid.New().String(),
			Type:     "user",
			Title:    "New Team Member",
			Message:  "Sarah Connor has joined your workspace",
			Severity: "info",
			Created:  time.Now().Add(-30 * time.Minute),
			Read:     false,
			TenantID: tenantID.String(),
		},
	}

	return notifications, nil
}

// SetupRoutes sets up workspace dashboard routes
func (h *DashboardHandler) SetupRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	// Apply authentication middleware to all workspace routes
	mux.Handle("/workspace/dashboard", authMiddleware(http.HandlerFunc(h.ShowDashboard)))
	mux.Handle("/workspace/api/dashboard", authMiddleware(http.HandlerFunc(h.GetDashboardData)))
	mux.Handle("/workspace/api/activity", authMiddleware(http.HandlerFunc(h.GetRecentActivity)))
	mux.Handle("/workspace/", authMiddleware(http.HandlerFunc(h.ShowDashboard))) // Default route
}

