package types

import (
	"time"
)

// WorkspaceDashboardData represents data for the workspace dashboard
type WorkspaceDashboardData struct {
	Title          string         `json:"title"`
	User           UserInfo       `json:"user"`
	TenantInfo     TenantInfo     `json:"tenant_info"`
	TenantStats    WorkspaceTenantStats    `json:"tenant_stats"`
	RecentActivity []ActivityItem `json:"recent_activity"`
	QuickActions   []QuickAction  `json:"quick_actions"`
	Notifications  []Notification `json:"notifications"`
	CSRFToken      string         `json:"csrf_token"`
}

// TenantInfo represents information about a tenant
type TenantInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Domain    string    `json:"domain"`
	Status    string    `json:"status"`
	Plan      string    `json:"plan"`
	UserCount int       `json:"user_count"`
	CreatedAt time.Time `json:"created_at"`
}

// WorkspaceTenantStats represents tenant-specific statistics for workspace dashboard
type WorkspaceTenantStats struct {
	ActiveUsers    int                    `json:"active_users"`
	ActiveProjects int                    `json:"active_projects"`
	OpenInvoices   int                    `json:"open_invoices"`
	MonthlyRevenue float64                `json:"monthly_revenue"`
	UpcomingTasks  int                    `json:"upcoming_tasks"`
	TeamMembers    int                    `json:"team_members"`
	LastMonth      *TenantStatsComparison `json:"last_month,omitempty"`
}

// TenantStatsComparison represents comparison stats with previous period
type TenantStatsComparison struct {
	Users    int     `json:"users"`
	Projects int     `json:"projects"`
	Revenue  float64 `json:"revenue"`
}

// WorkspaceUser represents a user in the workspace context
type WorkspaceUser struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	Department string    `json:"department"`
	Status     string    `json:"status"`
	Title      string    `json:"title,omitempty"`
	Phone      string    `json:"phone,omitempty"`
	LastLogin  time.Time `json:"last_login"`
	CreatedAt  time.Time `json:"created_at"`
}

// WorkspaceUserFilters represents filtering options for user lists
type WorkspaceUserFilters struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	Search     string `json:"search,omitempty"`
	Role       string `json:"role,omitempty"`
	Department string `json:"department,omitempty"`
	Status     string `json:"status,omitempty"`
}

// WorkspaceUsersPageData represents data for the workspace users page
type WorkspaceUsersPageData struct {
	Title      string                `json:"title"`
	Users      []WorkspaceUser       `json:"users"`
	Pagination *PaginationMeta       `json:"pagination"`
	Filters    *WorkspaceUserFilters `json:"filters"`
	CSRFToken  string                `json:"csrf_token"`
}

// WorkspaceUserCreateRequest represents data for creating a new user
type WorkspaceUserCreateRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	Department string `json:"department,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Title      string `json:"title,omitempty"`
}

// WorkspaceUserUpdateRequest represents data for updating a user
type WorkspaceUserUpdateRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	Department string `json:"department,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Title      string `json:"title,omitempty"`
	Status     string `json:"status,omitempty"`
}

// WorkspaceUserFormData represents data for user forms
type WorkspaceUserFormData struct {
	Title       string         `json:"title"`
	Action      string         `json:"action"`
	Method      string         `json:"method"`
	User        *WorkspaceUser `json:"user,omitempty"`
	Roles       []string       `json:"roles"`
	Departments []string       `json:"departments"`
	CSRFToken   string         `json:"csrf_token"`
}

// WorkspaceUserDetailPageData represents data for user detail page
type WorkspaceUserDetailPageData struct {
	Title     string         `json:"title"`
	User      *WorkspaceUser `json:"user"`
	CSRFToken string         `json:"csrf_token"`
}

// WorkspaceProject represents a project in the workspace
type WorkspaceProject struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Budget      float64    `json:"budget"`
	Spent       float64    `json:"spent"`
	Progress    int        `json:"progress"` // 0-100
	TeamSize    int        `json:"team_size"`
	CreatedAt   time.Time  `json:"created_at"`
}

// WorkspaceInvoice represents an invoice in the workspace context
type WorkspaceInvoice struct {
	ID          string     `json:"id"`
	Number      string     `json:"number"`
	ClientName  string     `json:"client_name"`
	Amount      float64    `json:"amount"`
	Status      string     `json:"status"`
	Date        time.Time  `json:"date"`
	DueDate     time.Time  `json:"due_date"`
	PaidDate    *time.Time `json:"paid_date,omitempty"`
	Description string     `json:"description"`
	Currency    string     `json:"currency"`
}

// WorkspaceTenantInfo represents tenant information for workspace
type WorkspaceTenantInfo struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Settings map[string]string `json:"settings"`
}


