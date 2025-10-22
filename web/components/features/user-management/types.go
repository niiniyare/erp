package usermanagement

import (
	"github.com/niiniyare/erp/web/components/atoms"
)

// Re-export Permission from atoms package for backward compatibility
type Permission = atoms.Permission

// User represents a system user
type User struct {
	ID           string   `json:"id"`
	Email        string   `json:"email"`
	Username     string   `json:"username,omitempty"`
	FirstName    string   `json:"firstName"`
	LastName     string   `json:"lastName"`
	DisplayName  string   `json:"displayName,omitempty"`
	Avatar       string   `json:"avatar,omitempty"`
	Phone        string   `json:"phone,omitempty"`
	Department   string   `json:"department,omitempty"`
	JobTitle     string   `json:"jobTitle,omitempty"`
	Status       string   `json:"status"` // active, inactive, pending, suspended
	RoleIDs      []string `json:"roleIds,omitempty"`
	Permissions  []string `json:"permissions,omitempty"`
	LastLoginAt  string   `json:"lastLoginAt,omitempty"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
	EmailVerified bool    `json:"emailVerified"`
	TwoFactorEnabled bool `json:"twoFactorEnabled"`
}

// Role represents a user role with associated permissions
type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	Description string   `json:"description,omitempty"`
	Color       string   `json:"color,omitempty"`
	Permissions []string `json:"permissions"`
	IsSystem    bool     `json:"isSystem"`     // System roles cannot be deleted
	IsDefault   bool     `json:"isDefault"`   // Default role for new users
	UserCount   int      `json:"userCount,omitempty"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

// UserFormData represents data for user creation/editing forms
type UserFormData struct {
	ID              string   `json:"id,omitempty"`
	Email           string   `json:"email"`
	Username        string   `json:"username,omitempty"`
	FirstName       string   `json:"firstName"`
	LastName        string   `json:"lastName"`
	Phone           string   `json:"phone,omitempty"`
	Department      string   `json:"department,omitempty"`
	JobTitle        string   `json:"jobTitle,omitempty"`
	RoleIDs         []string `json:"roleIds,omitempty"`
	Status          string   `json:"status"`
	SendWelcomeEmail bool    `json:"sendWelcomeEmail,omitempty"`
	RequirePasswordChange bool `json:"requirePasswordChange,omitempty"`
	TwoFactorEnabled bool    `json:"twoFactorEnabled,omitempty"`
}

// UserListItem represents a user in list/table views
type UserListItem struct {
	ID           string   `json:"id"`
	Email        string   `json:"email"`
	FirstName    string   `json:"firstName"`
	LastName     string   `json:"lastName"`
	DisplayName  string   `json:"displayName"`
	Avatar       string   `json:"avatar,omitempty"`
	Department   string   `json:"department,omitempty"`
	JobTitle     string   `json:"jobTitle,omitempty"`
	Status       string   `json:"status"`
	Roles        []string `json:"roles,omitempty"`
	LastLoginAt  string   `json:"lastLoginAt,omitempty"`
	CreatedAt    string   `json:"createdAt"`
	IsActive     bool     `json:"isActive"`
}

// UserProfile represents detailed user profile information
type UserProfile struct {
	User
	// Additional profile fields
	Bio          string            `json:"bio,omitempty"`
	Location     string            `json:"location,omitempty"`
	Website      string            `json:"website,omitempty"`
	Timezone     string            `json:"timezone,omitempty"`
	Language     string            `json:"language,omitempty"`
	Theme        string            `json:"theme,omitempty"`
	Preferences  map[string]interface{} `json:"preferences,omitempty"`
	SocialLinks  map[string]string `json:"socialLinks,omitempty"`
	Skills       []string          `json:"skills,omitempty"`
	Certifications []string        `json:"certifications,omitempty"`
}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers    int `json:"totalUsers"`
	ActiveUsers   int `json:"activeUsers"`
	InactiveUsers int `json:"inactiveUsers"`
	PendingUsers  int `json:"pendingUsers"`
	OnlineUsers   int `json:"onlineUsers"`
}

// RoleOption represents a role option for selectors
type RoleOption struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	UserCount   int    `json:"userCount,omitempty"`
	IsSystem    bool   `json:"isSystem"`
	Selected    bool   `json:"selected,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
}

// Helper functions

// GetFullName returns the full name of a user
func (u User) GetFullName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.FirstName + " " + u.LastName
}

// GetInitials returns the initials of a user
func (u User) GetInitials() string {
	initials := ""
	if u.FirstName != "" {
		initials += string(u.FirstName[0])
	}
	if u.LastName != "" {
		initials += string(u.LastName[0])
	}
	if initials == "" && u.Email != "" {
		initials = string(u.Email[0])
	}
	return initials
}

// IsActive returns whether the user is active
func (u User) IsActive() bool {
	return u.Status == "active"
}

// HasRole checks if the user has a specific role
func (u User) HasRole(roleID string) bool {
	for _, id := range u.RoleIDs {
		if id == roleID {
			return true
		}
	}
	return false
}

// HasPermission checks if the user has a specific permission
func (u User) HasPermission(permission string) bool {
	for _, perm := range u.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}