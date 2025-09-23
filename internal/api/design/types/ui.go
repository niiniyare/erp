package types

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// UI COMMON TYPES
// ============================================================================

// UIContext provides the current user's UI context and permissions
var UIContext = Type("UIContext", func() {
	Description("Current user UI context and permissions")
	Attribute("user_id", String, "Current user ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("tenant_id", String, "Current tenant ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("role", String, "User role", func() {
		Enum("admin", "tenant_admin", "user", "client")
		Example("tenant_admin")
	})
	Attribute("permissions", ArrayOf(String), "User permissions", func() {
		Example([]string{"tenant:read", "user:write", "finance:read"})
	})
	Attribute("ui_mode", String, "UI mode/interface type", func() {
		Enum("console", "workspace", "portal")
		Example("workspace")
	})
	Attribute("features", MapOf(String, Boolean), "Enabled feature flags", func() {
		Example(map[string]any{
			"finance_module":   true,
			"advanced_reports": false,
			"user_management":  true,
		})
	})
	Required("user_id", "role", "ui_mode", "permissions", "features")
})

// UIMenuItem represents a navigation menu item
var UIMenuItem = Type("UIMenuItem", func() {
	Description("Navigation menu item")
	Attribute("id", String, "Unique menu item identifier", func() {
		Example("dashboard")
	})
	Attribute("label", String, "Display label", func() {
		Example("Dashboard")
	})
	Attribute("icon", String, "Icon class or name", func() {
		Example("fas fa-tachometer-alt")
	})
	Attribute("url", String, "URL path", func() {
		Example("/dashboard")
	})
	Attribute("order", UInt, "Display order", func() {
		Example(1)
	})
	Attribute("enabled", Boolean, "Whether the item is enabled", func() {
		Default(true)
		Example(true)
	})
	Attribute("children", ArrayOf("UIMenuItem"), "Sub-menu items", func() {
		Example([]map[string]any{
			{
				"id":      "reports",
				"label":   "Reports",
				"url":     "/reports",
				"enabled": true,
			},
		})
	})
	Attribute("permissions", ArrayOf(String), "Required permissions", func() {
		Example([]string{"dashboard:read"})
	})
	Required("id", "label", "url", "enabled")
})

// UINavigation represents the navigation structure
var UINavigation = Type("UINavigation", func() {
	Description("Complete navigation structure")
	Attribute("main_menu", ArrayOf("UIMenuItem"), "Main navigation menu")
	Attribute("user_menu", ArrayOf("UIMenuItem"), "User-specific menu items")
	Attribute("quick_actions", ArrayOf("UIMenuItem"), "Quick action items")
	Required("main_menu")
})

// UITheme represents theme configuration
var UITheme = Type("UITheme", func() {
	Description("UI theme configuration")
	Attribute("name", String, "Theme name", func() {
		Example("default")
	})
	Attribute("primary_color", String, "Primary color", func() {
		Pattern(HexColorPattern)
		Example("#667eea")
	})
	Attribute("secondary_color", String, "Secondary color", func() {
		Pattern(HexColorPattern)
		Example("#764ba2")
	})
	Attribute("sidebar_color", String, "Sidebar background color", func() {
		Pattern(HexColorPattern)
		Example("#1f2937")
	})
	Attribute("dark_mode", Boolean, "Dark mode enabled", func() {
		Default(false)
		Example(false)
	})
	Required("name", "primary_color", "secondary_color")
})

// UIWidgetPosition represents widget position in a grid
var UIWidgetPosition = Type("UIWidgetPosition", func() {
	Description("Widget position in dashboard grid")
	Attribute("row", UInt, "Row position", func() {
		Example(1)
	})
	Attribute("col", UInt, "Column position", func() {
		Example(1)
	})
	Required("row", "col")
})

// UIWidget represents a dashboard widget
var UIWidget = Type("UIWidget", func() {
	Description("Dashboard widget configuration")
	Attribute("id", String, "Widget ID", func() {
		Example("tenant_stats")
	})
	Attribute("type", String, "Widget type", func() {
		Enum("stat_card", "chart", "table", "list", "custom")
		Example("stat_card")
	})
	Attribute("title", String, "Widget title", func() {
		Example("Total Tenants")
	})
	Attribute("size", String, "Widget size", func() {
		Enum("small", "medium", "large", "full")
		Default("medium")
		Example("medium")
	})
	Attribute("position", "UIWidgetPosition", "Widget position")
	Attribute("data_source", String, "Data source endpoint", func() {
		Example("/api/v1/admin/stats/tenants")
	})
	Attribute("refresh_interval", UInt, "Refresh interval in seconds", func() {
		Default(300)
		Example(300)
	})
	Attribute("permissions", ArrayOf(String), "Required permissions to view", func() {
		Example([]string{"admin:read"})
	})
	Required("id", "type", "title", "position")
})

// UIDashboard represents a dashboard configuration
var UIDashboard = Type("UIDashboard", func() {
	Description("Dashboard configuration")
	Attribute("id", String, "Dashboard ID", func() {
		Example("admin_dashboard")
	})
	Attribute("name", String, "Dashboard name", func() {
		Example("Admin Dashboard")
	})
	Attribute("description", String, "Dashboard description", func() {
		Example("System administration overview")
	})
	Attribute("widgets", ArrayOf("UIWidget"), "Dashboard widgets")
	Attribute("layout", String, "Dashboard layout type", func() {
		Enum("grid", "flex", "masonry")
		Default("grid")
		Example("grid")
	})
	Attribute("permissions", ArrayOf(String), "Required permissions", func() {
		Example([]string{"admin:read"})
	})
	Required("id", "name", "widgets", "layout")
})

// UINotification represents a user notification
var UINotification = Type("UINotification", func() {
	Description("User notification")
	Attribute("id", String, "Notification ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("type", String, "Notification type", func() {
		Enum("info", "success", "warning", "error")
		Example("info")
	})
	Attribute("title", String, "Notification title", func() {
		Example("System Update")
	})
	Attribute("message", String, "Notification message", func() {
		Example("System will be updated at 2 AM EST")
	})
	Attribute("read", Boolean, "Whether notification is read", func() {
		Default(false)
		Example(false)
	})
	Attribute("priority", String, "Notification priority", func() {
		Enum("low", "normal", "high", "critical")
		Default("normal")
		Example("normal")
	})
	Attribute("expires_at", String, "Expiration timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-31T23:59:59Z")
	})
	AuditFields()
	Required("id", "type", "title", "message", "read", "priority")
})

// UIBreadcrumb represents breadcrumb navigation
var UIBreadcrumb = Type("UIBreadcrumb", func() {
	Description("Breadcrumb navigation item")
	Attribute("label", String, "Breadcrumb label", func() {
		Example("Dashboard")
	})
	Attribute("url", String, "Breadcrumb URL", func() {
		Example("/dashboard")
	})
	Attribute("active", Boolean, "Whether this is the current page", func() {
		Default(false)
		Example(false)
	})
	Required("label")
})

// UIPageMetadata represents page metadata
var UIPageMetadata = Type("UIPageMetadata", func() {
	Description("Page metadata and configuration")
	Attribute("title", String, "Page title", func() {
		Example("Dashboard - Awo ERP")
	})
	Attribute("description", String, "Page description", func() {
		Example("System administration dashboard")
	})
	Attribute("breadcrumbs", ArrayOf("UIBreadcrumb"), "Breadcrumb navigation")
	Attribute("actions", ArrayOf("UIMenuItem"), "Page-specific actions")
	Attribute("permissions", ArrayOf(String), "Required permissions", func() {
		Example([]string{"dashboard:read"})
	})
	Required("title")
})

// UITableColumn represents a data table column
var UITableColumn = Type("UITableColumn", func() {
	Description("Data table column configuration")
	Attribute("key", String, "Column data key", func() {
		Example("name")
	})
	Attribute("label", String, "Column header label", func() {
		Example("Name")
	})
	Attribute("type", String, "Column data type", func() {
		Enum("text", "number", "date", "boolean", "status", "actions", "link")
		Default("text")
		Example("text")
	})
	Attribute("sortable", Boolean, "Whether column is sortable", func() {
		Default(true)
		Example(true)
	})
	Attribute("filterable", Boolean, "Whether column is filterable", func() {
		Default(false)
		Example(false)
	})
	Attribute("width", String, "Column width", func() {
		Example("200px")
	})
	Attribute("align", String, "Text alignment", func() {
		Enum("left", "center", "right")
		Default("left")
		Example("left")
	})
	Required("key", "label", "type")
})

// UIFieldValidation represents field validation rules
var UIFieldValidation = Type("UIFieldValidation", func() {
	Description("Field validation rules")
	Attribute("min_length", UInt, "Minimum length")
	Attribute("max_length", UInt, "Maximum length")
	Attribute("pattern", String, "Regex pattern")
	Attribute("custom_message", String, "Custom error message")
})

// UIFieldOption represents a field option for select fields
var UIFieldOption = Type("UIFieldOption", func() {
	Description("Field option for select fields")
	Attribute("value", String, "Option value")
	Attribute("label", String, "Option label")
	Required("value", "label")
})

// UIFormField represents a form field configuration
var UIFormField = Type("UIFormField", func() {
	Description("Form field configuration")
	Attribute("name", String, "Field name", func() {
		Example("email")
	})
	Attribute("label", String, "Field label", func() {
		Example("Email Address")
	})
	Attribute("type", String, "Field input type", func() {
		Enum("text", "email", "password", "number", "date", "select", "checkbox", "textarea", "file")
		Example("email")
	})
	Attribute("required", Boolean, "Whether field is required", func() {
		Default(false)
		Example(true)
	})
	Attribute("placeholder", String, "Field placeholder", func() {
		Example("Enter your email")
	})
	Attribute("help_text", String, "Help text for field", func() {
		Example("We'll never share your email")
	})
	Attribute("validation", "UIFieldValidation", "Field validation rules")
	Attribute("options", ArrayOf("UIFieldOption"), "Options for select fields")
	Required("name", "label", "type")
})

// ============================================================================
// UI FEATURE FLAGS
// ============================================================================

// UIFeatureFlag represents a UI-specific feature flag
var UIFeatureFlag = Type("UIFeatureFlag", func() {
	Description("UI feature flag configuration")
	Attribute("key", String, "Feature flag key", func() {
		Example("advanced_reports")
	})
	Attribute("enabled", Boolean, "Whether feature is enabled", func() {
		Example(true)
	})
	Attribute("ui_component", String, "UI component this flag affects", func() {
		Example("reports_menu")
	})
	Attribute("description", String, "Feature description", func() {
		Example("Enable advanced reporting features")
	})
	Required("key", "enabled")
})

// ============================================================================
// UI TENANT SWITCHING (for console mode)
// ============================================================================

// UITenantInfo represents tenant information for UI switching
var UITenantInfo = Type("UITenantInfo", func() {
	Description("Tenant information for UI context switching")
	Attribute("id", String, "Tenant ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("name", String, "Tenant name", func() {
		Example("Acme Corporation")
	})
	Attribute("subdomain", String, "Tenant subdomain", func() {
		Example("acme")
	})
	Attribute("status", String, "Tenant status", func() {
		Enum("active", "inactive", "suspended", "pending")
		Example("active")
	})
	Attribute("logo_url", String, "Tenant logo URL", func() {
		Format(FormatURI)
		Example("https://cdn.example.com/logos/acme.png")
	})
	Required("id", "name", "status")
})

// ============================================================================
// UI RESPONSE WRAPPERS
// ============================================================================

// UIResponseMetadata represents response metadata
var UIResponseMetadata = Type("UIResponseMetadata", func() {
	Description("Response metadata information")
	Attribute("timestamp", String, "Response timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	Attribute("request_id", String, "Request correlation ID", func() {
		Format(FormatUUID)
		Example("req-123e4567-e89b-12d3-a456-426614174000")
	})
})

// UIResponse wraps UI-specific responses
var UIResponse = Type("UIResponse", func() {
	Description("Standard UI response wrapper")
	Attribute("success", Boolean, "Operation success status", func() {
		Example(true)
	})
	Attribute("data", Any, "Response data")
	Attribute("ui_context", "UIContext", "Current UI context")
	Attribute("notifications", ArrayOf("UINotification"), "New notifications")
	Attribute("metadata", "UIResponseMetadata", "Response metadata")
	Required("success")
})

// ============================================================================
// UI HEADERS
// ============================================================================

// UIHeaders provides UI-specific headers
var UIHeaders = func() {
	CommonHeaders() // X-Tenant-ID, X-User-ID, etc.
	Header("X-UI-Mode", String, "UI interface mode", func() {
		Enum("console", "workspace", "portal")
		Example("workspace")
	})
	Header("X-Impersonate-Tenant", String, "Tenant ID for admin impersonation", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Header("X-Theme", String, "UI theme preference", func() {
		Example("default")
	})
}
