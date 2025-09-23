package types

import "time"

// LoginPageData represents data for the login page
type LoginPageData struct {
	Title       string
	CSRFToken   string
	Error       string
	RedirectURL string
}

// DashboardData represents data for the dashboard page
type DashboardData struct {
	Title           string
	User            UserInfo
	SystemStats     SystemStats
	RecentActivity  []ActivityItem
	QuickActions    []QuickAction
	Notifications   []Notification
	CSRFToken       string
}

// UserInfo represents current user information
type UserInfo struct {
	ID       string
	Name     string
	Email    string
	Role     string
	TenantID string
}

// SystemStats represents system-wide statistics
type SystemStats struct {
	TotalTenants      int
	ActiveUsers       int
	TotalTransactions int64
	SystemHealth      string
	LastBackup        time.Time
}

// ActivityItem represents a recent activity item
type ActivityItem struct {
	ID          string
	Type        string
	Description string
	UserID      string
	UserName    string
	Timestamp   time.Time
	Severity    string
}

// QuickAction represents a quick action button
type QuickAction struct {
	Title       string
	Description string
	URL         string
	Icon        string
	Permission  string
}

// Notification represents a system notification
type Notification struct {
	ID       string
	Type     string
	Title    string
	Message  string
	Severity string
	Created  time.Time
	Read     bool
}