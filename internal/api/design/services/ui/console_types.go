package ui

import (
	"github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// CONSOLE-SPECIFIC TYPES
// ============================================================================

// ConsoleStats represents system-wide statistics for admin console
var ConsoleStats = Type("ConsoleStats", func() {
	Description("System-wide statistics for admin console")
	Attribute("tenants", MapOf(String, Any), "Tenant statistics", func() {
		Attribute("total", UInt, "Total number of tenants", func() {
			Example(42)
		})
		Attribute("active", UInt, "Active tenants", func() {
			Example(38)
		})
		Attribute("suspended", UInt, "Suspended tenants", func() {
			Example(3)
		})
		Attribute("pending", UInt, "Pending tenants", func() {
			Example(1)
		})
		Required("total", "active", "suspended", "pending")
	})
	Attribute("users", MapOf(String, Any), "User statistics", func() {
		Attribute("total", UInt, "Total users across all tenants", func() {
			Example(1247)
		})
		Attribute("active_sessions", UInt, "Currently active sessions", func() {
			Example(156)
		})
		Attribute("new_today", UInt, "New users registered today", func() {
			Example(8)
		})
		Required("total", "active_sessions", "new_today")
	})
	Attribute("system", MapOf(String, Any), "System statistics", func() {
		Attribute("uptime", String, "System uptime", func() {
			Example("15d 4h 23m")
		})
		Attribute("version", String, "Application version", func() {
			Example("1.2.3")
		})
		Attribute("storage_used", String, "Storage usage", func() {
			Example("2.3GB")
		})
		Attribute("cpu_usage", Float64, "CPU usage percentage", func() {
			Example(15.7)
		})
		Attribute("memory_usage", Float64, "Memory usage percentage", func() {
			Example(68.2)
		})
		Required("uptime", "version", "storage_used", "cpu_usage", "memory_usage")
	})
	Attribute("revenue", MapOf(String, Any), "Revenue statistics", func() {
		Attribute("monthly_recurring", String, "Monthly recurring revenue", func() {
			Example("$12,450.00")
		})
		Attribute("annual_recurring", String, "Annual recurring revenue", func() {
			Example("$149,400.00")
		})
		Attribute("churn_rate", Float64, "Monthly churn rate percentage", func() {
			Example(2.3)
		})
		Required("monthly_recurring", "annual_recurring", "churn_rate")
	})
	Required("tenants", "users", "system")
})

// ConsoleTenantDetail represents detailed tenant information for console
var ConsoleTenantDetail = Type("ConsoleTenantDetail", func() {
	Description("Detailed tenant information for admin console")
	Attribute("id", String, "Tenant ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("name", String, "Tenant name", func() {
		Example("Acme Corporation")
	})
	Attribute("slug", String, "Tenant slug", func() {
		Example("acme")
	})
	Attribute("subdomain", String, "Subdomain", func() {
		Example("acme")
	})
	Attribute("status", String, "Tenant status", func() {
		Enum("active", "inactive", "suspended", "pending", "archived")
		Example("active")
	})
	Attribute("plan_type", String, "Subscription plan", func() {
		Example("enterprise")
	})
	Attribute("industry", String, "Industry type", func() {
		Example("Technology")
	})
	Attribute("company_size", String, "Company size", func() {
		Enum("startup", "small", "medium", "large", "enterprise")
		Example("medium")
	})
	Attribute("contact_info", types.ContactInfo, "Primary contact information")
	Attribute("billing_info", MapOf(String, Any), "Billing information", func() {
		Attribute("billing_email", String, "Billing email", func() {
			Format(FormatEmail)
			Example("billing@acme.com")
		})
		Attribute("payment_method", String, "Payment method", func() {
			Example("Credit Card")
		})
		Attribute("next_billing_date", String, "Next billing date", func() {
			Format(FormatDate)
			Example("2024-01-15")
		})
		Attribute("amount", String, "Monthly amount", func() {
			Example("$299.00")
		})
	})
	Attribute("usage_stats", MapOf(String, Any), "Usage statistics", func() {
		Attribute("users_count", UInt, "Number of users", func() {
			Example(25)
		})
		Attribute("storage_used", String, "Storage used", func() {
			Example("1.2GB")
		})
		Attribute("api_calls_month", UInt, "API calls this month", func() {
			Example(125000)
		})
		Attribute("last_login", String, "Last user login", func() {
			Format(FormatDateTime)
			Example("2023-12-07T09:30:00Z")
		})
	})
	Attribute("feature_flags", MapOf(String, Boolean), "Enabled feature flags", func() {
		Example(map[string]any{
			"advanced_reports": true,
			"api_access":       true,
			"custom_fields":    false,
		})
	})
	types.AuditFields()
	Required("id", "name", "slug", "status", "plan_type")
})

// ConsoleSystemHealth represents system health information
var ConsoleSystemHealth = Type("ConsoleSystemHealth", func() {
	Description("System health information for admin console")
	Attribute("overall_status", String, "Overall system status", func() {
		Enum("healthy", "degraded", "unhealthy")
		Example("healthy")
	})
	Attribute("services", ArrayOf(MapOf(String, Any)), "Individual service statuses", func() {
		Elem(func() {
			Attribute("name", String, "Service name", func() {
				Example("database")
			})
			Attribute("status", String, "Service status", func() {
				Enum("healthy", "degraded", "unhealthy")
				Example("healthy")
			})
			Attribute("response_time", String, "Average response time", func() {
				Example("2ms")
			})
			Attribute("last_check", String, "Last health check", func() {
				Format(FormatDateTime)
				Example("2023-12-07T10:30:00Z")
			})
			Required("name", "status")
		})
	})
	Attribute("alerts", ArrayOf(MapOf(String, Any)), "Active system alerts", func() {
		Elem(func() {
			Attribute("id", String, "Alert ID", func() {
				Example("alert-001")
			})
			Attribute("severity", String, "Alert severity", func() {
				Enum("info", "warning", "error", "critical")
				Example("warning")
			})
			Attribute("message", String, "Alert message", func() {
				Example("High memory usage detected")
			})
			Attribute("created_at", String, "Alert creation time", func() {
				Format(FormatDateTime)
				Example("2023-12-07T10:25:00Z")
			})
			Required("id", "severity", "message", "created_at")
		})
	})
	Required("overall_status", "services")
})

// ConsoleActivityLog represents system activity for admin console
var ConsoleActivityLog = Type("ConsoleActivityLog", func() {
	Description("System activity log entry")
	Attribute("id", String, "Activity ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("timestamp", String, "Activity timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	Attribute("type", String, "Activity type", func() {
		Enum("user_action", "system_event", "security_event", "data_change")
		Example("user_action")
	})
	Attribute("severity", String, "Activity severity", func() {
		Enum("info", "warning", "error", "critical")
		Example("info")
	})
	Attribute("description", String, "Activity description", func() {
		Example("User created new tenant")
	})
	Attribute("user_id", String, "User who performed the action", func() {
		Format(FormatUUID)
		Example("user-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("tenant_id", String, "Affected tenant ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("ip_address", String, "Source IP address", func() {
		Pattern(types.IPAddressPattern)
		Example("192.168.1.100")
	})
	Attribute("user_agent", String, "User agent string", func() {
		Example("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	})
	Attribute("details", types.CustomMetadata, "Additional activity details")
	Required("id", "timestamp", "type", "severity", "description")
})

// ============================================================================
// CONSOLE REQUEST/RESPONSE TYPES
// ============================================================================

// ConsoleDashboardResponse represents the console dashboard data
var ConsoleDashboardResponse = Type("ConsoleDashboardResponse", func() {
	Description("Console dashboard response")
	Attribute("stats", "ConsoleStats", "System statistics")
	Attribute("recent_activity", ArrayOf("ConsoleActivityLog"), "Recent system activity")
	Attribute("system_health", "ConsoleSystemHealth", "System health status")
	Attribute("notifications", ArrayOf(types.UINotification), "Admin notifications")
	Required("stats", "recent_activity", "system_health")
})

// ConsoleTenantListResponse represents tenant list for console
var ConsoleTenantListResponse = Type("ConsoleTenantListResponse", func() {
	Description("Console tenant list response")
	Attribute("tenants", ArrayOf("ConsoleTenantDetail"), "List of tenants")
	Attribute("pagination", types.PaginationMeta, "Pagination information")
	Attribute("filters", MapOf(String, Any), "Applied filters", func() {
		Attribute("status", String, "Status filter")
		Attribute("plan_type", String, "Plan type filter")
		Attribute("search", String, "Search query")
	})
	Required("tenants", "pagination")
})

// ConsoleTenantCreateRequest represents tenant creation request
var ConsoleTenantCreateRequest = Type("ConsoleTenantCreateRequest", func() {
	Description("Console tenant creation request")
	Attribute("name", String, "Tenant name", func() {
		MinLength(2)
		MaxLength(100)
		Example("Acme Corporation")
	})
	Attribute("slug", String, "Tenant slug", func() {
		Pattern("^[a-z0-9-]+$")
		MinLength(2)
		MaxLength(50)
		Example("acme")
	})
	Attribute("subdomain", String, "Subdomain", func() {
		Pattern("^[a-z0-9-]+$")
		MinLength(2)
		MaxLength(50)
		Example("acme")
	})
	Attribute("admin_email", String, "Admin email", func() {
		Format(FormatEmail)
		Example("admin@acme.com")
	})
	Attribute("plan_type", String, "Initial plan type", func() {
		Enum("startup", "small", "medium", "large", "enterprise")
		Default("small")
		Example("medium")
	})
	Attribute("industry", String, "Industry", func() {
		Example("Technology")
	})
	Attribute("company_size", String, "Company size", func() {
		Enum("startup", "small", "medium", "large", "enterprise")
		Example("medium")
	})
	Attribute("contact_info", types.ContactInfo, "Contact information")
	Attribute("feature_flags", MapOf(String, Boolean), "Initial feature flags")
	Required("name", "slug", "admin_email", "plan_type")
})

// ConsoleTenantUpdateRequest represents tenant update request
var ConsoleTenantUpdateRequest = Type("ConsoleTenantUpdateRequest", func() {
	Description("Console tenant update request")
	Attribute("name", String, "Tenant name", func() {
		MinLength(2)
		MaxLength(100)
		Example("Acme Corporation")
	})
	Attribute("status", String, "Tenant status", func() {
		Enum("active", "inactive", "suspended", "pending", "archived")
		Example("active")
	})
	Attribute("plan_type", String, "Plan type", func() {
		Enum("startup", "small", "medium", "large", "enterprise")
		Example("enterprise")
	})
	Attribute("industry", String, "Industry", func() {
		Example("Technology")
	})
	Attribute("company_size", String, "Company size", func() {
		Enum("startup", "small", "medium", "large", "enterprise")
		Example("large")
	})
	Attribute("contact_info", types.ContactInfo, "Contact information")
	Attribute("feature_flags", MapOf(String, Boolean), "Feature flags")
	Attribute("notes", String, "Admin notes", func() {
		MaxLength(1000)
		Example("Upgraded to enterprise plan")
	})
})

// ConsoleBulkActionRequest represents bulk operations on tenants
var ConsoleBulkActionRequest = Type("ConsoleBulkActionRequest", func() {
	Description("Console bulk action request")
	Attribute("action", String, "Action to perform", func() {
		Enum("activate", "suspend", "archive", "update_plan", "update_features")
		Example("suspend")
	})
	Attribute("tenant_ids", ArrayOf(String), "Tenant IDs to process", func() {
		MinLength(1)
		MaxLength(100)
		Example([]string{
			"123e4567-e89b-12d3-a456-426614174000",
			"987fcdeb-51d2-43b8-a456-426614174001",
		})
	})
	Attribute("parameters", MapOf(String, Any), "Action parameters", func() {
		Attribute("reason", String, "Reason for action", func() {
			Example("Payment overdue")
		})
		Attribute("plan_type", String, "New plan type for update_plan action", func() {
			Example("small")
		})
		Attribute("feature_flags", MapOf(String, Boolean), "Feature flags for update_features action")
		Attribute("notify_users", Boolean, "Whether to notify affected users", func() {
			Default(true)
			Example(true)
		})
	})
	Required("action", "tenant_ids")
})

// ============================================================================
// CONSOLE TEMPLATE DATA TYPES
// ============================================================================

// ConsolePageData represents data for console template rendering
var ConsolePageData = Type("ConsolePageData", func() {
	Description("Console page template data")
	Attribute("page_metadata", types.UIPageMetadata, "Page metadata")
	Attribute("ui_context", types.UIContext, "Current UI context")
	Attribute("navigation", types.UINavigation, "Navigation structure")
	Attribute("theme", types.UITheme, "UI theme configuration")
	Attribute("notifications", ArrayOf(types.UINotification), "User notifications")
	Attribute("tenant_selector", ArrayOf(types.UITenantInfo), "Available tenants for switching")
	Attribute("data", Any, "Page-specific data")
	Required("page_metadata", "ui_con ext", "navigation")
})

// ConsoleTenantFormData represents tenant form data
var ConsoleTenantFormData = Type("ConsoleTenantFormData", func() {
	Description("Console tenant form data")
	Attribute("tenant", "ConsoleTenantDetail", "Existing tenant data (for edit)")
	Attribute("form_fields", ArrayOf(types.UIFormField), "Form field configurations")
	Attribute("validation_errors", MapOf(String, String), "Field validation errors")
	Attribute("is_edit", Boolean, "Whether this is an edit form", func() {
		Example(false)
	})
	Attribute("plan_options", ArrayOf(MapOf(String, Any)), "Available plan options", func() {
		Elem(func() {
			Attribute("value", String, "Plan value")
			Attribute("label", String, "Plan label")
			Attribute("description", String, "Plan description")
			Attribute("price", String, "Plan price")
			Required("value", "label")
		})
	})
	Required("form_fields", "is_edit")
})
