package ui

import (
	"github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// WORKSPACE-SPECIFIC TYPES (Tenant Interface)
// ============================================================================

// WorkspaceTenantInfo represents current tenant information for workspace
var WorkspaceTenantInfo = Type("WorkspaceTenantInfo", func() {
	Description("Current tenant information for workspace interface")
	Attribute("id", String, "Tenant ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("name", String, "Tenant name", func() {
		Example("Acme Corporation")
	})
	Attribute("subdomain", String, "Subdomain", func() {
		Example("acme")
	})
	Attribute("plan_type", String, "Current plan", func() {
		Example("enterprise")
	})
	Attribute("logo_url", String, "Tenant logo URL", func() {
		Format(FormatURI)
		Example("https://cdn.example.com/logos/acme.png")
	})
	Attribute("settings", MapOf(String, Any), "Tenant-specific settings", func() {
		Attribute("timezone", String, "Default timezone", func() {
			Example("America/New_York")
		})
		Attribute("currency", String, "Default currency", func() {
			Pattern(types.CurrencyCodePattern)
			Example("USD")
		})
		Attribute("date_format", String, "Date format preference", func() {
			Example("MM/DD/YYYY")
		})
		Attribute("language", String, "Default language", func() {
			Example("en")
		})
	})
	Attribute("subscription", MapOf(String, Any), "Subscription information", func() {
		Attribute("status", String, "Subscription status", func() {
			Enum("active", "trial", "expired", "suspended")
			Example("active")
		})
		Attribute("expires_at", String, "Subscription expiry", func() {
			Format(FormatDateTime)
			Example("2024-12-31T23:59:59Z")
		})
		Attribute("users_limit", UInt, "Maximum users allowed", func() {
			Example(50)
		})
		Attribute("users_count", UInt, "Current user count", func() {
			Example(25)
		})
		Attribute("storage_limit", String, "Storage limit", func() {
			Example("100GB")
		})
		Attribute("storage_used", String, "Storage used", func() {
			Example("25GB")
		})
	})
	Required("id", "name", "plan_type")
})

// WorkspaceStats represents workspace dashboard statistics
var WorkspaceStats = Type("WorkspaceStats", func() {
	Description("Workspace dashboard statistics for tenant")
	Attribute("users", MapOf(String, Any), "User statistics", func() {
		Attribute("total", UInt, "Total users in tenant", func() {
			Example(25)
		})
		Attribute("active", UInt, "Active users", func() {
			Example(23)
		})
		Attribute("online", UInt, "Currently online users", func() {
			Example(8)
		})
		Attribute("new_this_month", UInt, "New users this month", func() {
			Example(3)
		})
		Required("total", "active", "online")
	})
	Attribute("projects", MapOf(String, Any), "Project statistics", func() {
		Attribute("total", UInt, "Total projects", func() {
			Example(12)
		})
		Attribute("active", UInt, "Active projects", func() {
			Example(8)
		})
		Attribute("completed", UInt, "Completed projects", func() {
			Example(4)
		})
		Required("total", "active", "completed")
	})
	Attribute("finance", MapOf(String, Any), "Financial statistics", func() {
		Attribute("revenue_this_month", String, "Revenue this month", func() {
			Example("$125,430.50")
		})
		Attribute("expenses_this_month", String, "Expenses this month", func() {
			Example("$89,200.25")
		})
		Attribute("profit_margin", Float64, "Profit margin percentage", func() {
			Example(28.7)
		})
		Attribute("pending_invoices", UInt, "Number of pending invoices", func() {
			Example(7)
		})
		Required("revenue_this_month", "expenses_this_month", "pending_invoices")
	})
	Attribute("tasks", MapOf(String, Any), "Task statistics", func() {
		Attribute("total", UInt, "Total tasks", func() {
			Example(156)
		})
		Attribute("completed", UInt, "Completed tasks", func() {
			Example(89)
		})
		Attribute("overdue", UInt, "Overdue tasks", func() {
			Example(12)
		})
		Attribute("due_today", UInt, "Tasks due today", func() {
			Example(5)
		})
		Required("total", "completed", "overdue", "due_today")
	})
	Required("users", "projects", "finance", "tasks")
})

// WorkspaceUser represents a user in the workspace
var WorkspaceUser = Type("WorkspaceUser", func() {
	Description("User information for workspace interface")
	Attribute("id", String, "User ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("email", String, "User email", func() {
		Format(FormatEmail)
		Example("john.doe@acme.com")
	})
	Attribute("first_name", String, "First name", func() {
		Example("John")
	})
	Attribute("last_name", String, "Last name", func() {
		Example("Doe")
	})
	Attribute("role", String, "User role within tenant", func() {
		Enum("admin", "manager", "user", "viewer")
		Example("manager")
	})
	Attribute("department", String, "Department", func() {
		Example("Engineering")
	})
	Attribute("status", String, "User status", func() {
		Enum("active", "inactive", "pending", "suspended")
		Example("active")
	})
	Attribute("last_login", String, "Last login timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T09:30:00Z")
	})
	Attribute("avatar_url", String, "User avatar URL", func() {
		Format(FormatURI)
		Example("https://cdn.example.com/avatars/john.jpg")
	})
	Attribute("permissions", ArrayOf(String), "User permissions", func() {
		Example([]string{"user:read", "project:write", "finance:read"})
	})
	types.AuditFields()
	Required("id", "email", "first_name", "last_name", "role", "status")
})

// WorkspaceProject represents a project in the workspace
var WorkspaceProject = Type("WorkspaceProject", func() {
	Description("Project information for workspace")
	Attribute("id", String, "Project ID", func() {
		Format(FormatUUID)
		Example("proj-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("name", String, "Project name", func() {
		Example("Website Redesign")
	})
	Attribute("description", String, "Project description", func() {
		Example("Complete redesign of company website")
	})
	Attribute("status", String, "Project status", func() {
		Enum("planning", "active", "on_hold", "completed", "cancelled")
		Example("active")
	})
	Attribute("priority", String, "Project priority", func() {
		Enum("low", "medium", "high", "critical")
		Example("high")
	})
	Attribute("start_date", String, "Project start date", func() {
		Format(FormatDate)
		Example("2023-11-01")
	})
	Attribute("end_date", String, "Project end date", func() {
		Format(FormatDate)
		Example("2024-02-15")
	})
	Attribute("progress", UInt, "Completion percentage", func() {
		Maximum(100)
		Example(65)
	})
	Attribute("team_members", ArrayOf(String), "Team member IDs", func() {
		Example([]string{
			"550e8400-e29b-41d4-a716-446655440000",
			"550e8400-e29b-41d4-a716-446655440001",
		})
	})
	Attribute("budget", types.Money, "Project budget")
	Attribute("spent", types.Money, "Amount spent")
	types.AuditFields()
	Required("id", "name", "status", "priority", "progress")
})

// WorkspaceInvoice represents an invoice in the workspace
var WorkspaceInvoice = Type("WorkspaceInvoice", func() {
	Description("Invoice information for workspace")
	Attribute("id", String, "Invoice ID", func() {
		Example("INV-2023-001")
	})
	Attribute("number", String, "Invoice number", func() {
		Example("INV-2023-001")
	})
	Attribute("customer_name", String, "Customer name", func() {
		Example("ABC Company")
	})
	Attribute("amount", types.Money, "Invoice amount")
	Attribute("status", String, "Invoice status", func() {
		Enum("draft", "sent", "viewed", "paid", "overdue", "cancelled")
		Example("sent")
	})
	Attribute("due_date", String, "Due date", func() {
		Format(FormatDate)
		Example("2023-12-31")
	})
	Attribute("issue_date", String, "Issue date", func() {
		Format(FormatDate)
		Example("2023-12-01")
	})
	Attribute("payment_terms", String, "Payment terms", func() {
		Example("Net 30")
	})
	types.AuditFields()
	Required("id", "number", "customer_name", "amount", "status", "due_date")
})

// ============================================================================
// WORKSPACE REQUEST/RESPONSE TYPES
// ============================================================================

// WorkspaceDashboardResponse represents workspace dashboard data
var WorkspaceDashboardResponse = Type("WorkspaceDashboardResponse", func() {
	Description("Workspace dashboard response")
	Attribute("tenant_info", "WorkspaceTenantInfo", "Current tenant information")
	Attribute("stats", "WorkspaceStats", "Dashboard statistics")
	Attribute("recent_projects", ArrayOf("WorkspaceProject"), "Recent projects")
	Attribute("pending_invoices", ArrayOf("WorkspaceInvoice"), "Pending invoices")
	Attribute("team_activity", ArrayOf(MapOf(String, Any)), "Recent team activity", func() {
		Elem(func() {
			Attribute("user_name", String, "User who performed action")
			Attribute("action", String, "Action description")
			Attribute("timestamp", String, "Action timestamp", func() {
				Format(FormatDateTime)
			})
			Attribute("type", String, "Activity type", func() {
				Enum("project", "task", "invoice", "user")
			})
			Required("user_name", "action", "timestamp", "type")
		})
	})
	Required("tenant_info", "stats")
})

// WorkspaceUserListResponse represents user list for workspace
var WorkspaceUserListResponse = Type("WorkspaceUserListResponse", func() {
	Description("Workspace user list response")
	Attribute("users", ArrayOf("WorkspaceUser"), "List of users")
	Attribute("pagination", types.PaginationMeta, "Pagination information")
	Attribute("filters", MapOf(String, Any), "Applied filters", func() {
		Attribute("role", String, "Role filter")
		Attribute("department", String, "Department filter")
		Attribute("status", String, "Status filter")
		Attribute("search", String, "Search query")
	})
	Required("users", "pagination")
})

// WorkspaceUserCreateRequest represents user creation request
var WorkspaceUserCreateRequest = Type("WorkspaceUserCreateRequest", func() {
	Description("Workspace user creation request")
	Attribute("email", String, "User email", func() {
		Format(FormatEmail)
		Example("jane.smith@acme.com")
	})
	Attribute("first_name", String, "First name", func() {
		MinLength(1)
		MaxLength(50)
		Example("Jane")
	})
	Attribute("last_name", String, "Last name", func() {
		MinLength(1)
		MaxLength(50)
		Example("Smith")
	})
	Attribute("role", String, "User role", func() {
		Enum("admin", "manager", "user", "viewer")
		Default("user")
		Example("user")
	})
	Attribute("department", String, "Department", func() {
		MaxLength(100)
		Example("Marketing")
	})
	Attribute("send_invitation", Boolean, "Send invitation email", func() {
		Default(true)
		Example(true)
	})
	Attribute("permissions", ArrayOf(String), "Custom permissions", func() {
		Example([]string{"project:read", "finance:read"})
	})
	Required("email", "first_name", "last_name", "role")
})

// WorkspaceUserUpdateRequest represents user update request
var WorkspaceUserUpdateRequest = Type("WorkspaceUserUpdateRequest", func() {
	Description("Workspace user update request")
	Attribute("first_name", String, "First name", func() {
		MinLength(1)
		MaxLength(50)
		Example("Jane")
	})
	Attribute("last_name", String, "Last name", func() {
		MinLength(1)
		MaxLength(50)
		Example("Smith")
	})
	Attribute("role", String, "User role", func() {
		Enum("admin", "manager", "user", "viewer")
		Example("manager")
	})
	Attribute("department", String, "Department", func() {
		MaxLength(100)
		Example("Engineering")
	})
	Attribute("status", String, "User status", func() {
		Enum("active", "inactive", "suspended")
		Example("active")
	})
	Attribute("permissions", ArrayOf(String), "Custom permissions", func() {
		Example([]string{"project:write", "finance:read", "user:read"})
	})
})

// ============================================================================
// WORKSPACE TEMPLATE DATA TYPES
// ============================================================================

// WorkspacePageData represents data for workspace template rendering
var WorkspacePageData = Type("WorkspacePageData", func() {
	Description("Workspace page template data")
	Attribute("page_metadata", types.UIPageMetadata, "Page metadata")
	Attribute("ui_context", types.UIContext, "Current UI context")
	Attribute("tenant_info", "WorkspaceTenantInfo", "Current tenant information")
	Attribute("navigation", types.UINavigation, "Navigation structure")
	Attribute("theme", types.UITheme, "UI theme configuration")
	Attribute("notifications", ArrayOf(types.UINotification), "User notifications")
	Attribute("data", Any, "Page-specific data")
	Required("page_metadata", "ui_context", "tenant_info", "navigation")
})

// WorkspaceUserFormData represents user form data for workspace
var WorkspaceUserFormData = Type("WorkspaceUserFormData", func() {
	Description("Workspace user form data")
	Attribute("user", "WorkspaceUser", "Existing user data (for edit)")
	Attribute("form_fields", ArrayOf(types.UIFormField), "Form field configurations")
	Attribute("validation_errors", MapOf(String, String), "Field validation errors")
	Attribute("is_edit", Boolean, "Whether this is an edit form", func() {
		Example(false)
	})
	Attribute("role_options", ArrayOf(MapOf(String, Any)), "Available role options", func() {
		Elem(func() {
			Attribute("value", String, "Role value")
			Attribute("label", String, "Role label")
			Attribute("description", String, "Role description")
			Attribute("permissions", ArrayOf(String), "Default permissions")
			Required("value", "label")
		})
	})
	Attribute("department_options", ArrayOf(String), "Available departments", func() {
		Example([]string{"Engineering", "Marketing", "Sales", "HR"})
	})
	Required("form_fields", "is_edit")
})
