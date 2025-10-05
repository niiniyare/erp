package types

import "time"

// QuickAction represents a quick action item
type QuickAction struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Icon        string `json:"icon"`
	Permission  string `json:"permission"`
	Color       string `json:"color,omitempty"`
}

// ActivityItem represents an activity log item
type ActivityItem struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	UserID      string    `json:"user_id"`
	UserName    string    `json:"user_name"`
	Timestamp   time.Time `json:"timestamp"`
	Severity    string    `json:"severity"`
	TenantID    string    `json:"tenant_id,omitempty"`
	ClientID    string    `json:"client_id,omitempty"`
}

// Notification represents a user notification
type Notification struct {
	ID       string    `json:"id"`
	Type     string    `json:"type"`
	Title    string    `json:"title"`
	Message  string    `json:"message"`
	Severity string    `json:"severity"`
	Created  time.Time `json:"created"`
	Read     bool      `json:"read"`
	TenantID string    `json:"tenant_id,omitempty"`
	ClientID string    `json:"client_id,omitempty"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// Pagination represents pagination information
type Pagination struct {
	CurrentPage int
	Limit       int
	Total       int
	HasNext     bool
	HasPrev     bool
}

// Address represents a physical address
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// BulkActionResults represents results from bulk operations
type BulkActionResults struct {
	SuccessCount int      `json:"success_count"`
	ErrorCount   int      `json:"error_count"`
	Errors       []string `json:"errors,omitempty"`
}

// UserInfo represents user information
type UserInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
}

// ContactInfo represents contact information
type ContactInfo struct {
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	CompanyName string `json:"companyName"`
}

// TenantStats represents tenant usage statistics (shared across services)
type TenantStats struct {
	UserCount    int       `json:"userCount"`
	ActiveUsers  int       `json:"activeUsers"`
	StorageUsed  float64   `json:"storageUsed"` // bytes
	LastActivity time.Time `json:"lastActivity"`
}
