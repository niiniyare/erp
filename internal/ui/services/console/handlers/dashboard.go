package console

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
	"github.com/niiniyare/erp/internal/ui/services/console/templates"
	"github.com/niiniyare/erp/internal/ui/types"
	"github.com/niiniyare/erp/internal/ui/widgets"
)

// DashboardHandler handles admin console dashboard operations
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

// ShowDashboard displays the admin console dashboard
func (h *DashboardHandler) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get UI context (user should be authenticated by middleware)
	uiCtx, ok := middleware.GetUIContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Ensure user has console admin role
	if err := middleware.RequireRole(ctx, middleware.UIRoleConsoleAdmin); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Gather dashboard data
	data, err := h.gatherDashboardData(ctx, uiCtx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to gather dashboard data", logger.Fields{
			"error":   err.Error(),
			"user_id": uiCtx.UserID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Add CSRF token
	data.CSRFToken = uiCtx.CSRFToken

	// Render dashboard template
	if err := console.DashboardPage(*data).Render(ctx, w); err != nil {
		h.logger.ErrorContext(ctx, "Failed to render dashboard", logger.Fields{
			"error":   err.Error(),
			"user_id": uiCtx.UserID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log dashboard access
	h.logger.InfoContext(ctx, "Console dashboard accessed", logger.Fields{
		"user_id": uiCtx.UserID,
	})
}

// GetSystemStats returns system statistics as JSON (for HTMX requests)
func (h *DashboardHandler) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get UI context
	uiCtx, ok := middleware.GetUIContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get system stats
	stats, err := h.getSystemStats(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get system stats", logger.Fields{
			"error":   err.Error(),
			"user_id": uiCtx.UserID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Render stats component
	if err := console.SystemStatsComponent(*stats).Render(ctx, w); err != nil {
		h.logger.ErrorContext(ctx, "Failed to render system stats", logger.Fields{
			"error":   err.Error(),
			"user_id": uiCtx.UserID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetRecentActivity returns recent activity as HTML (for HTMX requests)
func (h *DashboardHandler) GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get UI context
	uiCtx, ok := middleware.GetUIContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get recent activity
	activities, err := h.getRecentActivity(ctx, 20) // Get last 20 activities
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get recent activity", logger.Fields{
			"error":   err.Error(),
			"user_id": uiCtx.UserID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Render activity component
	if err := console.RecentActivityComponent(activities).Render(ctx, w); err != nil {
		h.logger.ErrorContext(ctx, "Failed to render recent activity", logger.Fields{
			"error":   err.Error(),
			"user_id": uiCtx.UserID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// ShowBasicElementsDemo displays a showcase of all basic UI elements
func (h *DashboardHandler) ShowBasicElementsDemo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get UI context (user should be authenticated by middleware)
	uiCtx, ok := middleware.GetUIContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Ensure user has console admin role
	if err := middleware.RequireRole(ctx, middleware.UIRoleConsoleAdmin); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get user information for the layout
	userInfo, err := h.getUserInfo(ctx, uiCtx.UserID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get user info for elements demo", logger.Fields{
			"error":   err.Error(),
			"user_id": uiCtx.UserID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Prepare data for the demo page
	demoPageData := widgets.ElementsDemoPageData{
		Title:     "Basic UI Elements Demo",
		User:      *userInfo,
		CSRFToken: uiCtx.CSRFToken,
	}

	if err := widgets.ElementsDemoPage(demoPageData).Render(ctx, w); err != nil {
		h.logger.ErrorContext(ctx, "Failed to render basic elements demo", logger.Fields{
			"error":   err.Error(),
			"user_id": uiCtx.UserID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log demo page access
	h.logger.InfoContext(ctx, "Basic elements demo accessed", logger.Fields{
		"user_id": uiCtx.UserID,
	})
}

// Private helper methods

// gatherDashboardData collects all data needed for the dashboard
func (h *DashboardHandler) gatherDashboardData(ctx context.Context, uiCtx *middleware.UIContext) (*types.DashboardData, error) {
	// Get user information
	userInfo, err := h.getUserInfo(ctx, uiCtx.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Get system statistics
	systemStats, err := h.getSystemStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system stats: %w", err)
	}

	// Get recent activity
	recentActivity, err := h.getRecentActivity(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	// Get quick actions
	quickActions := h.getQuickActions(ctx, uiCtx)

	// Get notifications
	notifications, err := h.getNotifications(ctx, uiCtx.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	return &types.DashboardData{
		Title:          "Admin Console Dashboard",
		User:           *userInfo,
		SystemStats:    *systemStats,
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
		Role:     "Console Admin",
		TenantID: user.TenantID.String(),
	}, nil
}

// getSystemStats collects system-wide statistics
func (h *DashboardHandler) getSystemStats(ctx context.Context) (*types.SystemStats, error) {
	// Get tenant count
	tenants, err := h.tenantService.ListTenants(ctx, 0, 1000) // offset, limit
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant count: %w", err)
	}

	// Get active user count (this would require a more sophisticated query in practice)
	activeUsers := 0 // Placeholder

	// Get transaction count (placeholder - would query finance system)
	totalTransactions := int64(0) // Placeholder

	// System health check (placeholder)
	systemHealth := "Healthy"

	return &types.SystemStats{
		TotalTenants:      len(tenants),
		ActiveUsers:       activeUsers,
		TotalTransactions: totalTransactions,
		SystemHealth:      systemHealth,
		LastBackup:        time.Now().Add(-4 * time.Hour), // Placeholder
	}, nil
}

// getRecentActivity gets recent system activity
func (h *DashboardHandler) getRecentActivity(ctx context.Context, limit int) ([]types.ActivityItem, error) {
	// For now, create placeholder activities since audit service needs proper tenant context
	// TODO: Implement proper audit log retrieval with tenant context
	activities := make([]types.ActivityItem, limit)
	for i := 0; i < limit; i++ {
		activities[i] = types.ActivityItem{
			ID:          uuid.New().String(),
			Type:        "system_activity",
			Description: "System activity placeholder",
			UserID:      uuid.New().String(),
			UserName:    "System User",
			Timestamp:   time.Now().Add(-time.Duration(i) * time.Hour),
			Severity:    "info",
		}
	}

	return activities, nil
}

// getQuickActions returns available quick actions for the user
func (h *DashboardHandler) getQuickActions(ctx context.Context, uiCtx *middleware.UIContext) []types.QuickAction {
	actions := []types.QuickAction{
		{
			Title:       "Manage Tenants",
			Description: "Create, update, and manage tenant organizations",
			URL:         "/console/tenants",
			Icon:        "building",
			Permission:  "tenant:manage",
		},
		{
			Title:       "User Management",
			Description: "Manage users, roles, and permissions",
			URL:         "/console/users",
			Icon:        "users",
			Permission:  "user:manage",
		},
		{
			Title:       "System Configuration",
			Description: "Configure system settings and parameters",
			URL:         "/console/system",
			Icon:        "settings",
			Permission:  "system:configure",
		},
		{
			Title:       "Audit Logs",
			Description: "View system audit logs and security events",
			URL:         "/console/audit",
			Icon:        "shield",
			Permission:  "audit:view",
		},
		{
			Title:       "Financial Overview",
			Description: "View system-wide financial metrics",
			URL:         "/console/finance",
			Icon:        "chart",
			Permission:  "finance:view_all",
		},
	}

	// Filter actions based on permissions (simplified - in practice you'd check ABAC)
	return actions
}

// getNotifications gets notifications for the user
func (h *DashboardHandler) getNotifications(ctx context.Context, userID uuid.UUID) ([]types.Notification, error) {
	// Placeholder implementation - in practice this would query a notifications system
	notifications := []types.Notification{
		{
			ID:       uuid.New().String(),
			Type:     "system",
			Title:    "System Maintenance",
			Message:  "Scheduled maintenance window tomorrow at 2 AM UTC",
			Severity: "info",
			Created:  time.Now().Add(-2 * time.Hour),
			Read:     false,
		},
		{
			ID:       uuid.New().String(),
			Type:     "security",
			Title:    "Security Alert",
			Message:  "Unusual login activity detected from IP 192.168.1.100",
			Severity: "warning",
			Created:  time.Now().Add(-1 * time.Hour),
			Read:     false,
		},
	}

	return notifications, nil
}

// determineSeverity determines the severity level of an audit event
func determineSeverity(eventType string) string {
	switch eventType {
	case "login_failed", "permission_denied", "security_violation":
		return "error"
	case "login_success", "logout", "permission_granted":
		return "info"
	case "user_created", "user_deleted", "role_changed":
		return "warning"
	default:
		return "info"
	}
}

// SetupRoutes sets up dashboard routes
func (h *DashboardHandler) SetupRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	// Apply authentication middleware to all dashboard routes
	mux.Handle("/console/dashboard", authMiddleware(http.HandlerFunc(h.ShowDashboard)))
	mux.Handle("/console/api/stats", authMiddleware(http.HandlerFunc(h.GetSystemStats)))
	mux.Handle("/console/api/activity", authMiddleware(http.HandlerFunc(h.GetRecentActivity)))
	mux.Handle("/console/demo/elements", authMiddleware(http.HandlerFunc(h.ShowBasicElementsDemo)))
	mux.Handle("/console/", authMiddleware(http.HandlerFunc(h.ShowDashboard))) // Default route
}
