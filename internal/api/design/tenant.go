package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// TENANT MANAGEMENT TYPES - Optimized for Fiber Integration
// ============================================================================

// Tenant represents a business tenant in the multi-tenant ERP system.
// Each tenant is an isolated business organization with its own data,
// configurations, and user base.
var Tenant = ResultType("application/vnd.tenant", func() {
	Description("Business tenant with complete multi-tenancy isolation and configuration support")
	ContentType("application/json")

	Attributes(func() {
		Field(1, "id", String, "Unique tenant identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
			Meta("struct:tag:json", "id")
			Meta("struct:tag:db", "id,omitempty")
		})

		Field(2, "slug", String, "URL-friendly unique identifier for tenant", func() {
			Pattern("^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$")
			MinLength(3)
			MaxLength(50)
			Example("acme-corp")
			Description("Used in URLs and subdomain routing. Must be globally unique.")
			Meta("struct:tag:json", "slug")
			Meta("struct:tag:db", "slug,omitempty")
		})

		Field(3, "name", String, "Human-readable tenant name", func() {
			MinLength(2)
			MaxLength(100)
			Example("ACME Corporation")
			Description("Display name shown in UI")
			Meta("struct:tag:json", "name")
			Meta("struct:tag:db", "name,omitempty")
		})

		Field(4, "email", String, "Primary contact email for tenant", func() {
			Format(FormatEmail)
			Example("admin@acme.com")
			Description("Used for administrative communications")
			Meta("struct:tag:json", "email")
			Meta("struct:tag:db", "email,omitempty")
		})

		Field(5, "billing_email", String, "Billing contact email", func() {
			Format(FormatEmail)
			Example("billing@acme.com")
			Description("Separate billing contact for financial operations")
			Meta("struct:tag:json", "billing_email,omitempty")
			Meta("struct:tag:db", "billing_email,omitempty")
		})

		Field(6, "subdomain", String, "Custom subdomain for tenant branding", func() {
			Pattern("^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$")
			MinLength(3)
			MaxLength(63)
			Example("acme")
			Description("Custom subdomain (e.g., acme.erp.com). Must be globally unique.")
			Meta("struct:tag:json", "subdomain,omitempty")
			Meta("struct:tag:db", "subdomain,omitempty")
		})

		Field(7, "status", String, "Current operational status of the tenant", func() {
			Enum("ACTIVE", "SUSPENDED", "TRIAL", "ARCHIVED", "PENDING_SETUP")
			Default("TRIAL")
			Example("ACTIVE")
			Description("Controls tenant access and billing")
			Meta("struct:tag:json", "status")
			Meta("struct:tag:db", "status,omitempty")
		})

		Field(8, "suspension_reason", String, "Reason for suspension", func() {
			MaxLength(500)
			Example("Payment overdue")
			Description("Required when status is SUSPENDED")
			Meta("struct:tag:json", "suspension_reason,omitempty")
			Meta("struct:tag:db", "suspension_reason,omitempty")
		})

		Field(9, "timezone", String, "Default timezone for tenant operations", func() {
			Pattern("^[A-Za-z_]+/[A-Za-z_]+$")
			Default("UTC")
			Example("America/New_York")
			Description("IANA timezone identifier")
			Meta("struct:tag:json", "timezone")
			Meta("struct:tag:db", "timezone,omitempty")
		})

		Field(10, "currency_code", String, "Default display currency", func() {
			Pattern("^[A-Z]{3}$")
			Default("USD")
			Example("USD")
			Description("ISO 4217 currency code for UI display")
			Meta("struct:tag:json", "currency_code")
			Meta("struct:tag:db", "currency_code,omitempty")
		})

		Field(11, "industry", String, "Business industry classification", func() {
			MaxLength(100)
			Example("Technology")
			Description("Used for industry-specific features and compliance")
			Meta("struct:tag:json", "industry,omitempty")
			Meta("struct:tag:db", "industry,omitempty")
		})

		Field(12, "company_size", String, "Organization size category", func() {
			Enum("STARTUP", "SMALL", "MEDIUM", "LARGE", "ENTERPRISE")
			Example("MEDIUM")
			Description("Affects feature availability and pricing tiers")
			Meta("struct:tag:json", "company_size,omitempty")
			Meta("struct:tag:db", "company_size,omitempty")
		})

		Field(13, "plan", String, "Current subscription plan", func() {
			Enum("FREE", "STARTER", "PROFESSIONAL", "ENTERPRISE", "CUSTOM")
			Default("FREE")
			Example("PROFESSIONAL")
			Description("Determines feature access and limits")
			Meta("struct:tag:json", "plan")
			Meta("struct:tag:db", "plan,omitempty")
		})

		Field(14, "trial_ends_at", String, "Trial expiration timestamp", func() {
			Format(FormatDateTime)
			Example("2024-01-31T23:59:59Z")
			Description("Required when status is TRIAL")
			Meta("struct:tag:json", "trial_ends_at,omitempty")
			Meta("struct:tag:db", "trial_ends_at,omitempty")
		})

		Field(15, "parent_tenant_id", String, "Parent tenant for enterprise hierarchies", func() {
			Format(FormatUUID)
			Example("650e8400-e29b-41d4-a716-446655440000")
			Description("Enables multi-tenant enterprise structures")
			Meta("struct:tag:json", "parent_tenant_id,omitempty")
			Meta("struct:tag:db", "parent_tenant_id,omitempty")
		})

		Field(16, "branding", Any, "Tenant branding configuration", func() {
			Example(map[string]any{
				"logo_url":        "https://cdn.acme.com/logo.png",
				"favicon_url":     "https://cdn.acme.com/favicon.ico",
				"primary_color":   "#1E3A8A",
				"secondary_color": "#3B82F6",
				"custom_css_url":  "https://cdn.acme.com/custom.css",
			})
			Description("Visual customization settings")
			Meta("struct:tag:json", "branding,omitempty")
			Meta("struct:tag:db", "branding,omitempty")
		})

		Field(17, "contact_info", Any, "Contact information", func() {
			Example(map[string]any{
				"phone":         "+1-555-123-4567",
				"support_email": "support@acme.com",
				"address":       "123 Main St, City, State 12345",
				"website":       "https://acme.com",
			})
			Description("Additional contact details")
			Meta("struct:tag:json", "contact_info,omitempty")
			Meta("struct:tag:db", "contact_info,omitempty")
		})

		Field(18, "custom_fields", Any, "Extensible custom metadata", func() {
			Example(map[string]any{
				"sales_rep":        "John Doe",
				"onboarding_notes": "Requires custom integration",
			})
			Description("Flexible key-value pairs for tenant-specific data")
			Meta("struct:tag:json", "custom_fields,omitempty")
			Meta("struct:tag:db", "custom_fields,omitempty")
		})

		// Audit fields
		AuditFields()

		Required("id", "slug", "name", "email", "status", "timezone", "currency_code", "created_at")
	})

	View("default", func() {
		Description("Standard tenant view for listings and general operations")
		Attribute("id")
		Attribute("slug")
		Attribute("name")
		Attribute("email")
		Attribute("status")
		Attribute("timezone")
		Attribute("currency_code")
		Attribute("plan")
		Attribute("created_at")
		Attribute("updated_at")
	})

	View("detailed", func() {
		Description("Complete tenant view with all configuration details")
		Attribute("id")
		Attribute("slug")
		Attribute("name")
		Attribute("email")
		Attribute("billing_email")
		Attribute("subdomain")
		Attribute("status")
		Attribute("suspension_reason")
		Attribute("timezone")
		Attribute("currency_code")
		Attribute("industry")
		Attribute("company_size")
		Attribute("plan")
		Attribute("trial_ends_at")
		Attribute("parent_tenant_id")
		Attribute("branding")
		Attribute("contact_info")
		Attribute("custom_fields")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})

	View("public", func() {
		Description("Public tenant information for branding and subdomain resolution")
		Attribute("slug")
		Attribute("name")
		Attribute("subdomain")
		Attribute("branding")
	})

	View("summary", func() {
		Description("Minimal tenant info for dropdowns and references")
		Attribute("id")
		Attribute("slug")
		Attribute("name")
		Attribute("status")
	})
})

// ============================================================================
// TENANT CONFIGURATION
// ============================================================================

// TenantConfiguration represents detailed operational configuration for a tenant.
var TenantConfiguration = Type("TenantConfiguration", func() {
	Description("Comprehensive tenant configuration including limits, quotas, and operational parameters")

	// Resource Limits
	Field(1, "max_users", UInt, "Maximum number of users allowed", func() {
		Minimum(1)
		Maximum(10000)
		Default(10)
		Example(100)
		Description("Enforced during user creation")
		Meta("struct:tag:json", "max_users")
	})

	Field(2, "max_entities", UInt, "Maximum number of organizational entities", func() {
		Minimum(1)
		Maximum(1000)
		Default(5)
		Example(25)
		Description("Limit on company/department hierarchies")
		Meta("struct:tag:json", "max_entities")
	})

	Field(3, "max_transactions_per_month", UInt, "Monthly transaction processing limit", func() {
		Minimum(100)
		Maximum(10000000)
		Default(1000)
		Example(50000)
		Description("Financial transaction quota per calendar month")
		Meta("struct:tag:json", "max_transactions_per_month")
	})

	Field(4, "storage_quota_mb", UInt, "Storage quota in megabytes", func() {
		Minimum(100)
		Maximum(10240000) // 10TB
		Default(1024)     // 1GB
		Example(102400)   // 100GB
		Description("File and document storage limit (using MB for precision)")
		Meta("struct:tag:json", "storage_quota_mb")
	})

	Field(5, "api_rate_limit_per_minute", UInt, "API requests per minute per tenant", func() {
		Minimum(10)
		Maximum(10000)
		Default(100)
		Example(500)
		Description("Global rate limit for all tenant API calls")
		Meta("struct:tag:json", "api_rate_limit_per_minute")
	})

	Field(6, "api_rate_limit_per_user", UInt, "API requests per minute per user", func() {
		Minimum(5)
		Maximum(1000)
		Default(60)
		Example(120)
		Description("Per-user rate limiting")
		Meta("struct:tag:json", "api_rate_limit_per_user,omitempty")
	})

	// Financial Configuration
	Field(7, "accounting_method", String, "Preferred accounting methodology", func() {
		Enum("ACCRUAL", "CASH", "HYBRID")
		Default("ACCRUAL")
		Example("ACCRUAL")
		Description("Affects financial reporting and calculations")
		Meta("struct:tag:json", "accounting_method")
	})

	Field(8, "fiscal_year_start_month", UInt, "Fiscal year start month (1-12)", func() {
		Minimum(1)
		Maximum(12)
		Default(1)
		Example(4)
		Description("Used for financial reporting periods")
		Meta("struct:tag:json", "fiscal_year_start_month")
	})

	Field(9, "base_currency", String, "Primary currency for accounting operations", func() {
		Pattern("^[A-Z]{3}$")
		Default("USD")
		Example("USD")
		Description("ISO 4217 currency code - base for all calculations")
		Meta("struct:tag:json", "base_currency")
	})

	Field(10, "supported_currencies", ArrayOf(String), "Additional supported currencies", func() {
		MinLength(1)
		MaxLength(20)
		Elem(func() {
			Pattern("^[A-Z]{3}$")
		})
		Example([]string{"USD", "EUR", "GBP", "CAD"})
		Description("Multi-currency support (must include base_currency)")
		Meta("struct:tag:json", "supported_currencies")
	})

	// Security Policies
	Field(11, "password_policy", Any, "Password security requirements", func() {
		Example(map[string]any{
			"min_length":        12,
			"require_uppercase": true,
			"require_lowercase": true,
			"require_numbers":   true,
			"require_symbols":   true,
			"max_age_days":      90,
			"history_count":     5,
		})
		Description("Enforced password complexity and rotation rules")
		Meta("struct:tag:json", "password_policy")
	})

	Field(12, "session_policy", Any, "User session configuration", func() {
		Example(map[string]any{
			"timeout_minutes":      480,
			"idle_timeout_minutes": 30,
			"concurrent_sessions":  3,
			"require_mfa":          false,
		})
		Description("Session management and security settings")
		Meta("struct:tag:json", "session_policy")
	})

	Field(13, "ip_whitelist", ArrayOf(String), "Allowed IP ranges", func() {
		MaxLength(100)
		Example([]string{"192.168.1.0/24", "10.0.0.0/8"})
		Description("CIDR notation IP ranges. Empty means no IP restriction.")
		Meta("struct:tag:json", "ip_whitelist,omitempty")
	})

	// Data Retention
	Field(14, "backup_retention_days", UInt, "Data backup retention period", func() {
		Minimum(7)
		Maximum(2555) // ~7 years
		Default(90)
		Example(365)
		Description("Number of days to retain backup data")
		Meta("struct:tag:json", "backup_retention_days")
	})

	Field(15, "audit_log_retention_days", UInt, "Audit log retention period", func() {
		Minimum(30)
		Maximum(2555) // ~7 years
		Default(365)
		Example(2555)
		Description("Compliance requirement for audit trail retention")
		Meta("struct:tag:json", "audit_log_retention_days")
	})

	Field(16, "data_residency", String, "Data storage region", func() {
		Enum("US", "EU", "UK", "APAC", "CA", "AU")
		Default("US")
		Example("EU")
		Description("Geographic data residency for compliance")
		Meta("struct:tag:json", "data_residency")
	})

	// Features and Integrations
	Field(17, "features_enabled", ArrayOf(String), "Enabled feature flags", func() {
		MinLength(0)
		MaxLength(100)
		Example([]string{
			"advanced_reporting",
			"multi_entity_consolidation",
			"api_access",
			"mobile_app",
			"integration_webhooks",
		})
		Description("List of enabled features for this tenant")
		Meta("struct:tag:json", "features_enabled,omitempty")
	})

	Field(18, "integrations_allowed", ArrayOf(String), "Permitted third-party integrations", func() {
		MaxLength(50)
		Example([]string{
			"salesforce",
			"quickbooks",
			"stripe",
			"slack",
		})
		Description("Approved external system integrations")
		Meta("struct:tag:json", "integrations_allowed,omitempty")
	})

	Field(19, "webhook_endpoints", ArrayOf(String), "Registered webhook URLs", func() {
		MaxLength(20)
		Elem(func() {
			Format(FormatURI)
		})
		Example([]string{"https://api.acme.com/webhooks/erp"})
		Description("Webhook destinations for event notifications")
		Meta("struct:tag:json", "webhook_endpoints,omitempty")
	})

	// Notification Settings
	Field(20, "notification_settings", Any, "Notification preferences", func() {
		Example(map[string]any{
			"email_notifications":   true,
			"sms_notifications":     false,
			"slack_notifications":   true,
			"webhook_notifications": true,
		})
		Description("Channels for system notifications")
		Meta("struct:tag:json", "notification_settings,omitempty")
	})

	// Audit fields
	AuditFields()

	Required("max_users", "max_entities", "max_transactions_per_month", "storage_quota_mb",
		"accounting_method", "fiscal_year_start_month", "base_currency", "supported_currencies",
		"password_policy", "session_policy", "backup_retention_days", "audit_log_retention_days")
})

// ============================================================================
// TENANT USAGE STATISTICS
// ============================================================================

// TenantUsageStats tracks tenant resource consumption and limits
var TenantUsageStats = Type("TenantUsageStats", func() {
	Description("Real-time usage statistics and quota consumption for tenant")

	Field(1, "tenant_id", String, "Tenant identifier", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Meta("struct:tag:json", "tenant_id")
	})

	// User Metrics
	Field(2, "current_users", UInt, "Active user count", func() {
		Example(45)
		Description("Currently active (non-deleted) users")
		Meta("struct:tag:json", "current_users")
	})

	Field(3, "current_entities", UInt, "Active entity count", func() {
		Example(8)
		Description("Organizational entities in use")
		Meta("struct:tag:json", "current_entities")
	})

	// Transaction Metrics
	Field(4, "transactions_this_month", UInt, "Current month transaction count", func() {
		Example(1250)
		Description("Financial transactions processed this calendar month")
		Meta("struct:tag:json", "transactions_this_month")
	})

	Field(5, "transactions_today", UInt, "Today's transaction count", func() {
		Example(42)
		Description("Transactions processed today")
		Meta("struct:tag:json", "transactions_today")
	})

	// Storage Metrics
	Field(6, "storage_used_mb", UInt, "Storage consumption in MB", func() {
		Example(15732)
		Description("Current storage usage including files and database")
		Meta("struct:tag:json", "storage_used_mb")
	})

	Field(7, "file_count", UInt, "Total file count", func() {
		Example(1247)
		Description("Number of uploaded files")
		Meta("struct:tag:json", "file_count,omitempty")
	})

	// API Metrics
	Field(8, "api_calls_today", UInt, "API requests made today", func() {
		Example(2847)
		Description("API usage for current day")
		Meta("struct:tag:json", "api_calls_today")
	})

	Field(9, "api_calls_this_month", UInt, "API requests this month", func() {
		Example(89432)
		Description("Monthly API usage")
		Meta("struct:tag:json", "api_calls_this_month,omitempty")
	})

	// Activity Metrics
	Field(10, "last_login", String, "Most recent user login timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
		Description("Latest user activity timestamp")
		Meta("struct:tag:json", "last_login,omitempty")
	})

	Field(11, "active_sessions", UInt, "Current active sessions", func() {
		Example(12)
		Description("Number of logged-in users right now")
		Meta("struct:tag:json", "active_sessions,omitempty")
	})

	// Quota Warnings
	Field(12, "quota_warnings", ArrayOf(String), "Current quota warnings", func() {
		Example([]string{
			"storage_80_percent",
			"users_near_limit",
			"transactions_75_percent",
		})
		Description("Active quota threshold warnings")
		Meta("struct:tag:json", "quota_warnings,omitempty")
	})

	Field(13, "quota_percentages", Any, "Quota usage percentages", func() {
		Example(map[string]any{
			"users":        75.0,
			"storage":      83.5,
			"transactions": 25.0,
		})
		Description("Percentage of quota used for each resource")
		Meta("struct:tag:json", "quota_percentages,omitempty")
	})

	// Metadata
	Field(14, "calculated_at", String, "Statistics calculation timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T15:45:00Z")
		Description("When these statistics were last computed")
		Meta("struct:tag:json", "calculated_at")
	})

	Field(15, "period_start", String, "Current period start date", func() {
		Format(FormatDateTime)
		Example("2023-12-01T00:00:00Z")
		Description("Start of current billing/quota period")
		Meta("struct:tag:json", "period_start,omitempty")
	})

	Field(16, "period_end", String, "Current period end date", func() {
		Format(FormatDateTime)
		Example("2023-12-31T23:59:59Z")
		Description("End of current billing/quota period")
		Meta("struct:tag:json", "period_end,omitempty")
	})

	Required("tenant_id", "current_users", "current_entities", "transactions_this_month",
		"storage_used_mb", "api_calls_today", "calculated_at")
})

// ============================================================================
// REQUEST/RESPONSE PAYLOAD TYPES FOR FIBER
// ============================================================================

// CreateTenantRequest - Payload for creating a new tenant
var CreateTenantRequest = Type("CreateTenantRequest", func() {
	Description("Request payload for creating a new tenant")

	Field(1, "slug", String, "URL-friendly unique identifier", func() {
		Pattern("^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$")
		MinLength(3)
		MaxLength(50)
		Example("acme-corp")
		Meta("struct:tag:json", "slug")
		Meta("struct:tag:validate", "required,min=3,max=50")
	})

	Field(2, "name", String, "Tenant display name", func() {
		MinLength(2)
		MaxLength(100)
		Example("ACME Corporation")
		Meta("struct:tag:json", "name")
		Meta("struct:tag:validate", "required,min=2,max=100")
	})

	Field(3, "email", String, "Primary contact email", func() {
		Format(FormatEmail)
		Example("admin@acme.com")
		Meta("struct:tag:json", "email")
		Meta("struct:tag:validate", "required,email")
	})

	Field(4, "billing_email", String, "Billing contact email", func() {
		Format(FormatEmail)
		Example("billing@acme.com")
		Meta("struct:tag:json", "billing_email,omitempty")
		Meta("struct:tag:validate", "omitempty,email")
	})

	Field(5, "subdomain", String, "Custom subdomain", func() {
		Pattern("^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$")
		MinLength(3)
		MaxLength(63)
		Example("acme")
		Meta("struct:tag:json", "subdomain,omitempty")
		Meta("struct:tag:validate", "omitempty,min=3,max=63")
	})

	Field(6, "timezone", String, "Default timezone", func() {
		Pattern("^[A-Za-z_]+/[A-Za-z_]+$")
		Default("UTC")
		Example("America/New_York")
		Meta("struct:tag:json", "timezone,omitempty")
		Meta("struct:tag:validate", "omitempty")
	})

	Field(7, "currency_code", String, "Default currency", func() {
		Pattern("^[A-Z]{3}$")
		Default("USD")
		Example("USD")
		Meta("struct:tag:json", "currency_code,omitempty")
		Meta("struct:tag:validate", "omitempty,len=3")
	})

	Field(8, "industry", String, "Business industry", func() {
		MaxLength(100)
		Example("Technology")
		Meta("struct:tag:json", "industry,omitempty")
		Meta("struct:tag:validate", "omitempty,max=100")
	})

	Field(9, "company_size", String, "Company size category", func() {
		Enum("STARTUP", "SMALL", "MEDIUM", "LARGE", "ENTERPRISE")
		Example("MEDIUM")
		Meta("struct:tag:json", "company_size,omitempty")
		Meta("struct:tag:validate", "omitempty,oneof=STARTUP SMALL MEDIUM LARGE ENTERPRISE")
	})

	Field(10, "plan", String, "Initial subscription plan", func() {
		Enum("FREE", "STARTER", "PROFESSIONAL", "ENTERPRISE", "CUSTOM")
		Default("FREE")
		Example("PROFESSIONAL")
		Meta("struct:tag:json", "plan,omitempty")
		Meta("struct:tag:validate", "omitempty,oneof=FREE STARTER PROFESSIONAL ENTERPRISE CUSTOM")
	})

	Required("slug", "name", "email")
})

// UpdateTenantRequest - Payload for updating tenant information
var UpdateTenantRequest = Type("UpdateTenantRequest", func() {
	Description("Request payload for updating tenant information")

	Field(1, "name", String, "Tenant display name", func() {
		MinLength(2)
		MaxLength(100)
		Example("ACME Corporation Ltd")
		Meta("struct:tag:json", "name,omitempty")
		Meta("struct:tag:validate", "omitempty,min=2,max=100")
	})

	Field(2, "email", String, "Primary contact email", func() {
		Format(FormatEmail)
		Example("admin@acme.com")
		Meta("struct:tag:json", "email,omitempty")
		Meta("struct:tag:validate", "omitempty,email")
	})

	Field(3, "billing_email", String, "Billing contact email", func() {
		Format(FormatEmail)
		Example("billing@acme.com")
		Meta("struct:tag:json", "billing_email,omitempty")
		Meta("struct:tag:validate", "omitempty,email")
	})

	Field(4, "status", String, "Tenant status", func() {
		Enum("ACTIVE", "SUSPENDED", "TRIAL", "ARCHIVED", "PENDING_SETUP")
		Example("ACTIVE")
		Meta("struct:tag:json", "status,omitempty")
		Meta("struct:tag:validate", "omitempty,oneof=ACTIVE SUSPENDED TRIAL ARCHIVED PENDING_SETUP")
	})

	Field(5, "suspension_reason", String, "Suspension reason", func() {
		MaxLength(500)
		Example("Payment overdue")
		Meta("struct:tag:json", "suspension_reason,omitempty")
		Meta("struct:tag:validate", "omitempty,max=500")
	})

	Field(6, "timezone", String, "Default timezone", func() {
		Pattern("^[A-Za-z_]+/[A-Za-z_]+$")
		Example("Europe/London")
		Meta("struct:tag:json", "timezone,omitempty")
		Meta("struct:tag:validate", "omitempty")
	})

	Field(7, "currency_code", String, "Default currency", func() {
		Pattern("^[A-Z]{3}$")
		Example("EUR")
		Meta("struct:tag:json", "currency_code,omitempty")
		Meta("struct:tag:validate", "omitempty,len=3")
	})

	Field(8, "branding", Any, "Branding configuration", func() {
		Meta("struct:tag:json", "branding,omitempty")
		Meta("struct:tag:validate", "omitempty")
	})

	Field(9, "contact_info", Any, "Contact information", func() {
		Meta("struct:tag:json", "contact_info,omitempty")
		Meta("struct:tag:validate", "omitempty")
	})
})

// TenantQueryParams - Query parameters for listing tenants
var TenantQueryParams = Type("TenantQueryParams", func() {
	Description("Query parameters for filtering and paginating tenant listings")

	Field(1, "page", UInt, "Page number", func() {
		Minimum(1)
		Default(1)
		Example(1)
		Meta("struct:tag:json", "page,omitempty")
		Meta("struct:tag:query", "page")
	})

	Field(2, "page_size", UInt, "Items per page", func() {
		Minimum(1)
		Maximum(100)
		Default(20)
		Example(20)
		Meta("struct:tag:json", "page_size,omitempty")
		Meta("struct:tag:query", "page_size")
	})

	Field(3, "status", String, "Filter by status", func() {
		Enum("ACTIVE", "SUSPENDED", "TRIAL", "ARCHIVED", "PENDING_SETUP")
		Example("ACTIVE")
		Meta("struct:tag:json", "status,omitempty")
		Meta("struct:tag:query", "status")
	})

	Field(4, "plan", String, "Filter by plan", func() {
		Enum("FREE", "STARTER", "PROFESSIONAL", "ENTERPRISE", "CUSTOM")
		Example("PROFESSIONAL")
		Meta("struct:tag:json", "plan,omitempty")
		Meta("struct:tag:query", "plan")
	})

	Field(5, "company_size", String, "Filter by company size", func() {
		Enum("STARTUP", "SMALL", "MEDIUM", "LARGE", "ENTERPRISE")
		Example("MEDIUM")
		Meta("struct:tag:json", "company_size,omitempty")
		Meta("struct:tag:query", "company_size")
	})

	Field(6, "search", String, "Search in name, slug, email", func() {
		MinLength(2)
		MaxLength(100)
		Example("acme")
		Meta("struct:tag:json", "search,omitempty")
		Meta("struct:tag:query", "search")
	})

	Field(7, "sort_by", String, "Sort field", func() {
		Enum("name", "created_at", "updated_at", "slug", "status")
		Default("created_at")
		Example("name")
		Meta("struct:tag:json", "sort_by,omitempty")
		Meta("struct:tag:query", "sort_by")
	})

	Field(8, "sort_order", String, "Sort direction", func() {
		Enum("asc", "desc")
		Default("desc")
		Example("asc")
		Meta("struct:tag:json", "sort_order,omitempty")
		Meta("struct:tag:query", "sort_order")
	})

	Field(9, "created_after", String, "Filter tenants created after date", func() {
		Format(FormatDateTime)
		Example("2023-01-01T00:00:00Z")
		Meta("struct:tag:json", "created_after,omitempty")
		Meta("struct:tag:query", "created_after")
	})

	Field(10, "created_before", String, "Filter tenants created before date", func() {
		Format(FormatDateTime)
		Example("2023-12-31T23:59:59Z")
		Meta("struct:tag:json", "created_before,omitempty")
		Meta("struct:tag:query", "created_before")
	})
})

// PaginatedTenantsResponse - Paginated list of tenants
var PaginatedTenantsResponse = Type("PaginatedTenantsResponse", func() {
	Description("Paginated response for tenant listings")

	Field(1, "data", ArrayOf(Tenant), "Array of tenant records", func() {
		Meta("struct:tag:json", "data")
	})

	Field(2, "pagination", Any, "Pagination metadata", func() {
		Example(map[string]any{
			"page":        1,
			"page_size":   20,
			"total_items": 156,
			"total_pages": 8,
		})
		Meta("struct:tag:json", "pagination")
	})

	Required("data", "pagination")
})

// UpdateTenantConfigRequest - Payload for updating tenant configuration
var UpdateTenantConfigRequest = Type("UpdateTenantConfigRequest", func() {
	Description("Request payload for updating tenant configuration")

	// Resource Limits
	Field(1, "max_users", UInt, "Maximum users", func() {
		Minimum(1)
		Maximum(10000)
		Example(100)
		Meta("struct:tag:json", "max_users,omitempty")
		Meta("struct:tag:validate", "omitempty,min=1,max=10000")
	})

	Field(2, "max_entities", UInt, "Maximum entities", func() {
		Minimum(1)
		Maximum(1000)
		Example(25)
		Meta("struct:tag:json", "max_entities,omitempty")
		Meta("struct:tag:validate", "omitempty,min=1,max=1000")
	})

	Field(3, "max_transactions_per_month", UInt, "Monthly transaction limit", func() {
		Minimum(100)
		Maximum(10000000)
		Example(50000)
		Meta("struct:tag:json", "max_transactions_per_month,omitempty")
		Meta("struct:tag:validate", "omitempty,min=100")
	})

	Field(4, "storage_quota_mb", UInt, "Storage quota in MB", func() {
		Minimum(100)
		Maximum(10240000)
		Example(102400)
		Meta("struct:tag:json", "storage_quota_mb,omitempty")
		Meta("struct:tag:validate", "omitempty,min=100")
	})

	Field(5, "api_rate_limit_per_minute", UInt, "API rate limit per minute", func() {
		Minimum(10)
		Maximum(10000)
		Example(500)
		Meta("struct:tag:json", "api_rate_limit_per_minute,omitempty")
		Meta("struct:tag:validate", "omitempty,min=10,max=10000")
	})

	// Financial Configuration
	Field(6, "accounting_method", String, "Accounting method", func() {
		Enum("ACCRUAL", "CASH", "HYBRID")
		Example("ACCRUAL")
		Meta("struct:tag:json", "accounting_method,omitempty")
		Meta("struct:tag:validate", "omitempty,oneof=ACCRUAL CASH HYBRID")
	})

	Field(7, "fiscal_year_start_month", UInt, "Fiscal year start month", func() {
		Minimum(1)
		Maximum(12)
		Example(4)
		Meta("struct:tag:json", "fiscal_year_start_month,omitempty")
		Meta("struct:tag:validate", "omitempty,min=1,max=12")
	})

	Field(8, "base_currency", String, "Base currency", func() {
		Pattern("^[A-Z]{3}$")
		Example("USD")
		Meta("struct:tag:json", "base_currency,omitempty")
		Meta("struct:tag:validate", "omitempty,len=3")
	})

	Field(9, "supported_currencies", ArrayOf(String), "Supported currencies", func() {
		MinLength(1)
		MaxLength(20)
		Elem(func() {
			Pattern("^[A-Z]{3}$")
		})
		Example([]string{"USD", "EUR", "GBP"})
		Meta("struct:tag:json", "supported_currencies,omitempty")
		Meta("struct:tag:validate", "omitempty,min=1,max=20")
	})

	// Security Policies
	Field(10, "password_policy", Any, "Password policy", func() {
		Meta("struct:tag:json", "password_policy,omitempty")
		Meta("struct:tag:validate", "omitempty")
	})

	Field(11, "session_policy", Any, "Session policy", func() {
		Meta("struct:tag:json", "session_policy,omitempty")
		Meta("struct:tag:validate", "omitempty")
	})

	Field(12, "ip_whitelist", ArrayOf(String), "IP whitelist", func() {
		MaxLength(100)
		Example([]string{"192.168.1.0/24"})
		Meta("struct:tag:json", "ip_whitelist,omitempty")
		Meta("struct:tag:validate", "omitempty,max=100")
	})

	// Features
	Field(13, "features_enabled", ArrayOf(String), "Enabled features", func() {
		MaxLength(100)
		Example([]string{"advanced_reporting", "api_access"})
		Meta("struct:tag:json", "features_enabled,omitempty")
		Meta("struct:tag:validate", "omitempty,max=100")
	})

	Field(14, "integrations_allowed", ArrayOf(String), "Allowed integrations", func() {
		MaxLength(50)
		Example([]string{"salesforce", "quickbooks"})
		Meta("struct:tag:json", "integrations_allowed,omitempty")
		Meta("struct:tag:validate", "omitempty,max=50")
	})
})

// ============================================================================
// ERROR TYPES
// ============================================================================

// TenantError - Standard error response for tenant operations
var TenantError = Type("TenantError", func() {
	Description("Error response for tenant operations")

	Field(1, "code", String, "Error code", func() {
		Example("TENANT_NOT_FOUND")
		Meta("struct:tag:json", "code")
	})

	Field(2, "message", String, "Human-readable error message", func() {
		Example("Tenant with slug 'acme-corp' not found")
		Meta("struct:tag:json", "message")
	})

	Field(3, "details", Any, "Additional error details", func() {
		Example(map[string]any{
			"field":      "slug",
			"constraint": "unique",
			"value":      "acme-corp",
		})
		Meta("struct:tag:json", "details,omitempty")
	})

	Field(4, "timestamp", String, "Error timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T15:45:00Z")
		Meta("struct:tag:json", "timestamp")
	})

	Field(5, "request_id", String, "Request tracking ID", func() {
		Example("req_8f7d9e5c4b3a2f1e")
		Meta("struct:tag:json", "request_id,omitempty")
	})

	Required("code", "message", "timestamp")
})

// QuotaExceededError - Quota/limit exceeded error
var QuotaExceededError = Type("QuotaExceededError", func() {
	Description("Error when tenant quota or limit is exceeded")

	Field(1, "code", String, "Error code", func() {
		Default("QUOTA_EXCEEDED")
		Example("QUOTA_EXCEEDED")
		Meta("struct:tag:json", "code")
	})

	Field(2, "message", String, "Error message", func() {
		Example("Monthly transaction limit exceeded")
		Meta("struct:tag:json", "message")
	})

	Field(3, "resource", String, "Resource type", func() {
		Enum("users", "entities", "transactions", "storage", "api_calls")
		Example("transactions")
		Meta("struct:tag:json", "resource")
	})

	Field(4, "limit", UInt, "Quota limit", func() {
		Example(50000)
		Meta("struct:tag:json", "limit")
	})

	Field(5, "current", UInt, "Current usage", func() {
		Example(50127)
		Meta("struct:tag:json", "current")
	})

	Field(6, "timestamp", String, "Error timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T15:45:00Z")
		Meta("struct:tag:json", "timestamp")
	})

	Required("code", "message", "resource", "limit", "current", "timestamp")
})

// ============================================================================
// BATCH OPERATIONS
// ============================================================================

// BatchUpdateTenantsRequest - Batch update multiple tenants
var BatchUpdateTenantsRequest = Type("BatchUpdateTenantsRequest", func() {
	Description("Request payload for batch updating tenants")

	Field(1, "tenant_ids", ArrayOf(String), "Tenant IDs to update", func() {
		MinLength(1)
		MaxLength(100)
		Elem(func() {
			Format(FormatUUID)
		})
		Example([]string{
			"550e8400-e29b-41d4-a716-446655440000",
			"650e8400-e29b-41d4-a716-446655440001",
		})
		Meta("struct:tag:json", "tenant_ids")
		Meta("struct:tag:validate", "required,min=1,max=100")
	})

	Field(2, "updates", Any, "Fields to update", func() {
		Example(map[string]any{
			"status": "ACTIVE",
			"plan":   "PROFESSIONAL",
		})
		Meta("struct:tag:json", "updates")
		Meta("struct:tag:validate", "required")
	})

	Required("tenant_ids", "updates")
})

// BatchUpdateTenantsResponse - Result of batch update operation
var BatchUpdateTenantsResponse = Type("BatchUpdateTenantsResponse", func() {
	Description("Response for batch update operation")

	Field(1, "success_count", UInt, "Number of successful updates", func() {
		Example(45)
		Meta("struct:tag:json", "success_count")
	})

	Field(2, "failure_count", UInt, "Number of failed updates", func() {
		Example(3)
		Meta("struct:tag:json", "failure_count")
	})

	Field(3, "failures", ArrayOf(Any), "Details of failed updates", func() {
		Example([]any{
			map[string]any{
				"tenant_id": "650e8400-e29b-41d4-a716-446655440001",
				"error":     "Tenant not found",
			},
		})
		Meta("struct:tag:json", "failures,omitempty")
	})

	Field(4, "updated_at", String, "Batch operation timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T15:45:00Z")
		Meta("struct:tag:json", "updated_at")
	})

	Required("success_count", "failure_count", "updated_at")
})

// ============================================================================
// HELPER TYPES FOR FIBER INTEGRATION
// ============================================================================

// SuccessResponse - Generic success response wrapper
var SuccessResponse = Type("SuccessResponse", func() {
	Description("Generic success response for Fiber handlers")

	Field(1, "success", Boolean, "Operation success flag", func() {
		Example(true)
		Meta("struct:tag:json", "success")
	})

	Field(2, "message", String, "Success message", func() {
		Example("Tenant created successfully")
		Meta("struct:tag:json", "message,omitempty")
	})

	Field(3, "data", Any, "Response data", func() {
		Meta("struct:tag:json", "data,omitempty")
	})

	Required("success")
})

// TenantContextKey - Context key for tenant resolution in Fiber middleware
var TenantContextKey = Type("TenantContextKey", func() {
	Description("Context information for tenant resolution")

	Field(1, "tenant_id", String, "Resolved tenant ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Meta("struct:tag:json", "tenant_id")
	})

	Field(2, "tenant_slug", String, "Resolved tenant slug", func() {
		Example("acme-corp")
		Meta("struct:tag:json", "tenant_slug")
	})

	Field(3, "resolved_by", String, "Resolution method", func() {
		Enum("subdomain", "header", "path", "query", "jwt")
		Example("subdomain")
		Meta("struct:tag:json", "resolved_by")
	})

	Required("tenant_id", "tenant_slug", "resolved_by")
})

// ============================================================================
// TENANT SERVICE
// ============================================================================

var _ = Service("tenant", func() {
	Description("Tenant management service")

	Method("list", func() {
		Description("List all tenants")
		Result(ArrayOf(Tenant))
		HTTP(func() {
			GET("/tenants")
			Response(StatusOK)
		})
	})

	Method("get", func() {
		Description("Get tenant by ID")
		Payload(func() {
			Field(1, "id", String, "Tenant ID", func() {
				Format(FormatUUID)
			})
			Required("id")
		})
		Result(Tenant)
		HTTP(func() {
			GET("/tenants/{id}")
			Response(StatusOK)
		})
	})

	Method("create", func() {
		Description("Create a new tenant")
		Payload(func() {
			Field(1, "name", String, "Tenant name")
			Field(2, "slug", String, "Tenant slug")
			Field(3, "email", String, "Contact email")
			Required("name", "slug", "email")
		})
		Result(Tenant)
		HTTP(func() {
			POST("/tenants")
			Response(StatusCreated)
		})
	})

	Method("update", func() {
		Description("Update tenant")
		Payload(func() {
			Field(1, "id", String, "Tenant ID", func() {
				Format(FormatUUID)
			})
			Field(2, "name", String, "Tenant name")
			Field(3, "email", String, "Contact email")
			Required("id")
		})
		Result(Tenant)
		HTTP(func() {
			PUT("/tenants/{id}")
			Response(StatusOK)
		})
	})

	Method("delete", func() {
		Description("Delete tenant")
		Payload(func() {
			Field(1, "id", String, "Tenant ID", func() {
				Format(FormatUUID)
			})
			Required("id")
		})
		HTTP(func() {
			DELETE("/tenants/{id}")
			Response(StatusNoContent)
		})
	})
})
