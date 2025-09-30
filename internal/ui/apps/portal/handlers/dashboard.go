package portal

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

// DashboardHandler handles portal dashboard operations for client users
type DashboardHandler struct {
	iamService    iam.Service
	abacService   abac.Service
	tenantService tenant.Service
	auditService  audit.Service
	logger        logger.Logger
}

// NewDashboardHandler creates a new portal dashboard handler
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

// ShowDashboard displays the client portal dashboard
func (h *DashboardHandler) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get UI context (user should be authenticated by middleware)
	uiCtx, ok := middleware.GetUIContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Ensure user has client access role
	if err := middleware.RequireRole(ctx, middleware.UIRoleClientUser); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Verify user belongs to a client context
	if uiCtx.ClientID == uuid.Nil {
		http.Error(w, "No client context", http.StatusForbidden)
		return
	}

	// Gather dashboard data for this client
	data, err := h.gatherDashboardData(ctx, uiCtx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to gather portal dashboard data", logger.Fields{
			"error":     err.Error(),
			"user_id":   uiCtx.UserID,
			"client_id": uiCtx.ClientID,
			"tenant_id": uiCtx.TenantID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Add CSRF token
	data.CSRFToken = uiCtx.CSRFToken

	// Render dashboard template
	// TODO: Create portal dashboard template
	// if err := portal.DashboardPage(*data).Render(ctx, w); err != nil {
	//     h.logger.ErrorContext(ctx, "Failed to render portal dashboard", logger.Fields{
	//         "error":     err.Error(),
	//         "user_id":   uiCtx.UserID,
	//         "client_id": uiCtx.ClientID,
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
			<title>Client Portal Dashboard</title>
			<style>
				body { font-family: Arial, sans-serif; margin: 40px; background: #f8fafc; }
				.header { background: linear-gradient(135deg, #8b5cf6, #7c3aed); color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
				.welcome { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
				.stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin: 20px 0; }
				.stat-card { background: white; border-left: 4px solid #8b5cf6; padding: 20px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
				.stat-value { font-size: 2em; font-weight: bold; color: #8b5cf6; margin-bottom: 5px; }
				.stat-label { color: #6b7280; font-size: 14px; }
				.quick-actions { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin: 20px 0; }
				.action-card { background: white; padding: 15px; border-radius: 8px; text-align: center; box-shadow: 0 1px 3px rgba(0,0,0,0.1); text-decoration: none; color: inherit; }
				.action-card:hover { transform: translateY(-2px); box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
				.action-icon { font-size: 24px; margin-bottom: 10px; }
				.recent-activity { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
				.activity-item { padding: 12px 0; border-bottom: 1px solid #f3f4f6; display: flex; align-items: center; }
				.activity-icon { width: 40px; height: 40px; background: #f3f4f6; border-radius: 50%%; display: flex; align-items: center; justify-content: center; margin-right: 12px; }
			</style>
		</head>
		<body>
			<div class="header">
				<h1>Client Portal</h1>
				<p>Welcome, %s - %s</p>
			</div>

			<div class="welcome">
				<h2>Account Overview</h2>
				<p>Access your invoices, orders, support tickets, and account information all in one place.</p>
			</div>
			
			<div class="stats">
				<div class="stat-card">
					<div class="stat-value">$%s</div>
					<div class="stat-label">Account Balance</div>
				</div>
				<div class="stat-card">
					<div class="stat-value">%d</div>
					<div class="stat-label">Open Invoices</div>
				</div>
				<div class="stat-card">
					<div class="stat-value">%d</div>
					<div class="stat-label">Active Orders</div>
				</div>
				<div class="stat-card">
					<div class="stat-value">%d</div>
					<div class="stat-label">Support Tickets</div>
				</div>
			</div>

			<h3>Quick Actions</h3>
			<div class="quick-actions">
				<a href="/portal/invoices" class="action-card">
					<div class="action-icon">📄</div>
					<div><strong>View Invoices</strong></div>
					<div style="color: #6b7280; font-size: 12px;">Manage your billing</div>
				</a>
				<a href="/portal/orders" class="action-card">
					<div class="action-icon">📦</div>
					<div><strong>Track Orders</strong></div>
					<div style="color: #6b7280; font-size: 12px;">Check order status</div>
				</a>
				<a href="/portal/support" class="action-card">
					<div class="action-icon">🎧</div>
					<div><strong>Support Center</strong></div>
					<div style="color: #6b7280; font-size: 12px;">Get help when needed</div>
				</a>
				<a href="/portal/account" class="action-card">
					<div class="action-icon">⚙️</div>
					<div><strong>Account Settings</strong></div>
					<div style="color: #6b7280; font-size: 12px;">Update your information</div>
				</a>
			</div>

			<div class="recent-activity">
				<h3>Recent Activity</h3>
				<div class="activity-item">
					<div class="activity-icon">💰</div>
					<div>
						<div><strong>Payment Received</strong></div>
						<div style="color: #6b7280; font-size: 12px;">Invoice #INV-2024-001 • 2 hours ago</div>
					</div>
				</div>
				<div class="activity-item">
					<div class="activity-icon">📦</div>
					<div>
						<div><strong>Order Shipped</strong></div>
						<div style="color: #6b7280; font-size: 12px;">Order #ORD-2024-003 • 1 day ago</div>
					</div>
				</div>
				<div class="activity-item">
					<div class="activity-icon">📋</div>
					<div>
						<div><strong>New Invoice</strong></div>
						<div style="color: #6b7280; font-size: 12px;">Invoice #INV-2024-002 • 3 days ago</div>
					</div>
				</div>
			</div>

			<div style="margin-top: 30px; text-align: center; color: #6b7280; font-size: 14px;">
				<p>This is a placeholder for the Portal dashboard. Templates will be implemented next.</p>
			</div>
		</body>
		</html>
	`, 
		data.ClientInfo.ContactName, 
		data.ClientInfo.CompanyName,
		fmt.Sprintf("%.2f", data.AccountSummary.CurrentBalance),
		data.AccountSummary.OpenInvoices,
		data.AccountSummary.ActiveOrders,
		data.AccountSummary.OpenTickets,
	)

	// Log dashboard access
	h.logger.InfoContext(ctx, "Portal dashboard accessed", logger.Fields{
		"user_id":   uiCtx.UserID,
		"client_id": uiCtx.ClientID,
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

	// Ensure client access
	if err := middleware.RequireRole(ctx, middleware.UIRoleClientUser); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get account summary
	summary, err := h.getAccountSummary(ctx, uiCtx.ClientID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get account summary", logger.Fields{
			"error":     err.Error(),
			"user_id":   uiCtx.UserID,
			"client_id": uiCtx.ClientID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Render summary as JSON
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"currentBalance": %.2f,
		"openInvoices": %d,
		"activeOrders": %d,
		"openTickets": %d,
		"nextDueDate": "%s",
		"nextDueAmount": %.2f,
		"lastUpdated": "%s"
	}`, 
		summary.CurrentBalance,
		summary.OpenInvoices,
		summary.ActiveOrders,
		summary.OpenTickets,
		summary.NextDueDate.Format("2006-01-02"),
		summary.NextDueAmount,
		time.Now().Format(time.RFC3339),
	)
}

// GetRecentActivity returns recent client activity as HTML (for HTMX requests)
func (h *DashboardHandler) GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get UI context
	uiCtx, ok := middleware.GetUIContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get recent activity for this client
	activities, err := h.getRecentActivity(ctx, uiCtx.ClientID, 20)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get recent activity", logger.Fields{
			"error":     err.Error(),
			"user_id":   uiCtx.UserID,
			"client_id": uiCtx.ClientID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Render activity component as HTML
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<div class='activity-list'>"))
	for _, activity := range activities {
		icon := "📝"
		switch activity.Type {
		case "payment":
			icon = "💰"
		case "invoice":
			icon = "📋"
		case "order":
			icon = "📦"
		case "support":
			icon = "🎧"
		}
		
		fmt.Fprintf(w, `
			<div class="activity-item">
				<div class="activity-icon">%s</div>
				<div class="activity-content">
					<div class="activity-title">%s</div>
					<div class="activity-time">%s</div>
				</div>
			</div>
		`, icon, activity.Description, activity.Timestamp.Format("Jan 2, 3:04 PM"))
	}
	w.Write([]byte("</div>"))
}

// Private helper methods

// gatherDashboardData collects all data needed for the portal dashboard
func (h *DashboardHandler) gatherDashboardData(ctx context.Context, uiCtx *middleware.UIContext) (*types.PortalDashboardData, error) {
	// Get client information
	clientInfo, err := h.getClientInfo(ctx, uiCtx.ClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client info: %w", err)
	}

	// Get account summary
	accountSummary, err := h.getAccountSummary(ctx, uiCtx.ClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account summary: %w", err)
	}

	// Get recent activity
	recentActivity, err := h.getRecentActivity(ctx, uiCtx.ClientID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	// Get quick actions for portal
	quickActions := h.getPortalQuickActions(ctx, uiCtx)

	// Get notifications for this client
	notifications, err := h.getNotifications(ctx, uiCtx.ClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	return &types.PortalDashboardData{
		Title:          "Client Portal Dashboard",
		ClientInfo:     *clientInfo,
		AccountSummary: *accountSummary,
		RecentActivity: recentActivity,
		QuickActions:   quickActions,
		Notifications:  notifications,
	}, nil
}

// getClientInfo gets information about the current client
func (h *DashboardHandler) getClientInfo(ctx context.Context, clientID uuid.UUID) (*types.PortalClientInfo, error) {
	// TODO: Implement actual client info retrieval
	// For now, return placeholder data
	return &types.PortalClientInfo{
		ID:          clientID.String(),
		CompanyName: "ACME Manufacturing",
		ContactName: "Robert Johnson",
		Email:       "robert@acmemfg.com",
		Phone:       "+1-555-987-6543",
		Status:      "active",
		AccountType: "premium",
		MemberSince: time.Now().Add(-365 * 24 * time.Hour), // 1 year ago
		BillingAddress: types.Address{
			Street:     "123 Industrial Blvd",
			City:       "Manufacturing City",
			State:      "MC",
			PostalCode: "12345",
			Country:    "USA",
		},
		ShippingAddress: types.Address{
			Street:     "456 Warehouse Dr",
			City:       "Distribution City", 
			State:      "DC",
			PostalCode: "67890",
			Country:    "USA",
		},
	}, nil
}

// getAccountSummary gets the account summary for the client
func (h *DashboardHandler) getAccountSummary(ctx context.Context, clientID uuid.UUID) (*types.PortalAccountSummary, error) {
	// TODO: Implement actual account summary retrieval from finance service
	// For now, return placeholder data
	return &types.PortalAccountSummary{
		CurrentBalance:  2457.89,
		OpenInvoices:    3,
		ActiveOrders:    2,
		OpenTickets:     1,
		NextDueDate:     time.Now().Add(7 * 24 * time.Hour), // Due in 7 days
		NextDueAmount:   1250.00,
		CreditLimit:     10000.00,
		AvailableCredit: 7542.11,
		LastPayment: &types.PaymentInfo{
			Amount: 1850.00,
			Date:   time.Now().Add(-14 * 24 * time.Hour), // 2 weeks ago
			Method: "Bank Transfer",
		},
		YearToDate: &types.YearToDateSummary{
			TotalPurchases: 45678.90,
			TotalPayments:  43221.01,
			OrderCount:     24,
			AverageOrderValue: 1903.29,
		},
	}, nil
}

// getRecentActivity gets recent activity for this client
func (h *DashboardHandler) getRecentActivity(ctx context.Context, clientID uuid.UUID, limit int) ([]types.ActivityItem, error) {
	// For now, create placeholder activities
	// TODO: Implement proper activity log retrieval with client context
	activities := make([]types.ActivityItem, limit)
	activityTypes := []string{"payment", "invoice", "order", "support", "account_update"}
	descriptions := map[string][]string{
		"payment":        {"Payment of $1,850.00 processed", "Payment of $3,250.00 received", "Automatic payment processed"},
		"invoice":        {"New invoice #INV-2024-001 created", "Invoice #INV-2024-002 sent", "Invoice #INV-2023-145 paid"},
		"order":          {"Order #ORD-2024-003 shipped", "New order #ORD-2024-004 placed", "Order #ORD-2024-002 delivered"},
		"support":        {"Support ticket #TKT-001 opened", "Support ticket #TKT-002 resolved", "Support inquiry submitted"},
		"account_update": {"Billing address updated", "Contact information updated", "Payment method added"},
	}
	
	for i := 0; i < limit; i++ {
		activityType := activityTypes[i%len(activityTypes)]
		typeDescriptions := descriptions[activityType]
		description := typeDescriptions[i%len(typeDescriptions)]
		
		activities[i] = types.ActivityItem{
			ID:          uuid.New().String(),
			Type:        activityType,
			Description: description,
			UserID:      clientID.String(),
			UserName:    "System",
			Timestamp:   time.Now().Add(-time.Duration(i*30) * time.Minute),
			Severity:    "info",
			ClientID:    clientID.String(),
		}
	}

	return activities, nil
}

// getPortalQuickActions returns available quick actions for portal users
func (h *DashboardHandler) getPortalQuickActions(ctx context.Context, uiCtx *middleware.UIContext) []types.QuickAction {
	actions := []types.QuickAction{
		{
			Title:       "View Invoices",
			Description: "Access your billing statements and payment history",
			URL:         "/portal/invoices",
			Icon:        "receipt",
			Permission:  "client:read",
			Color:       "blue",
		},
		{
			Title:       "Track Orders",
			Description: "Monitor the status of your current orders",
			URL:         "/portal/orders",
			Icon:        "truck",
			Permission:  "client:read",
			Color:       "green",
		},
		{
			Title:       "Support Center",
			Description: "Get help and submit support tickets",
			URL:         "/portal/support",
			Icon:        "headphones",
			Permission:  "client:read",
			Color:       "yellow",
		},
		{
			Title:       "Account Info",
			Description: "View and update your account information",
			URL:         "/portal/account",
			Icon:        "user",
			Permission:  "client:read",
			Color:       "purple",
		},
		{
			Title:       "Download Reports",
			Description: "Access your account statements and reports",
			URL:         "/portal/reports",
			Icon:        "download",
			Permission:  "client:read",
			Color:       "gray",
		},
	}

	// TODO: Filter actions based on ABAC permissions
	return actions
}

// getNotifications gets notifications for the client
func (h *DashboardHandler) getNotifications(ctx context.Context, clientID uuid.UUID) ([]types.Notification, error) {
	// Placeholder implementation - client-scoped notifications
	notifications := []types.Notification{
		{
			ID:       uuid.New().String(),
			Type:     "payment",
			Title:    "Payment Due Soon",
			Message:  "Invoice #INV-2024-001 is due in 3 days ($1,250.00)",
			Severity: "warning",
			Created:  time.Now().Add(-1 * time.Hour),
			Read:     false,
			ClientID: clientID.String(),
		},
		{
			ID:       uuid.New().String(),
			Type:     "order",
			Title:    "Order Shipped",
			Message:  "Your order #ORD-2024-003 has been shipped and is on its way",
			Severity: "info",
			Created:  time.Now().Add(-4 * time.Hour),
			Read:     false,
			ClientID: clientID.String(),
		},
		{
			ID:       uuid.New().String(),
			Type:     "account",
			Title:    "Account Statement Available",
			Message:  "Your monthly statement for March 2024 is now available",
			Severity: "info",
			Created:  time.Now().Add(-2 * 24 * time.Hour),
			Read:     true,
			ClientID: clientID.String(),
		},
	}

	return notifications, nil
}

// SetupRoutes sets up portal dashboard routes
func (h *DashboardHandler) SetupRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	// Apply authentication middleware to all portal routes
	mux.Handle("/portal/dashboard", authMiddleware(http.HandlerFunc(h.ShowDashboard)))
	mux.Handle("/portal/api/dashboard", authMiddleware(http.HandlerFunc(h.GetDashboardData)))
	mux.Handle("/portal/api/activity", authMiddleware(http.HandlerFunc(h.GetRecentActivity)))
	mux.Handle("/portal/", authMiddleware(http.HandlerFunc(h.ShowDashboard))) // Default route
}