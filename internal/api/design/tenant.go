package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// TENANT MANAGEMENT TYPES
// ============================================================================

// Tenant represents a business tenant in the multi-tenant ERP system.
// Each tenant is an isolated business organization with its own data,
// configurations, and user base.
var Tenant = ResultType("application/vnd.erp.tenant", func() {
	Description("Business tenant with complete multi-tenancy isolation and configuration support")
	
	Attributes(func() {
		Field(1, "id", String, "Unique tenant identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
		})
		
		Field(2, "slug", String, "URL-friendly unique identifier for tenant", func() {
			Pattern("^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$")
			MinLength(3)
			MaxLength(50)
			Example("acme-corp")
			Description("Used in URLs and subdomain routing")
		})
		
		Field(3, "name", String, "Human-readable tenant name", func() {
			MinLength(2)
			MaxLength(100)
			Example("ACME Corporation")
			Description("Display name shown in UI")
		})
		
		Field(4, "email", String, "Primary contact email for tenant", func() {
			Format(FormatEmail)
			Example("admin@acme.com")
			Description("Used for administrative communications")
		})
		
		Field(5, "subdomain", String, "Custom subdomain for tenant branding", func() {
			Pattern("^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$")
			MinLength(3)
			MaxLength(63)
			Example("acme")
			Description("Optional custom subdomain (e.g., acme.erp.com)")
		})
		
		Field(6, "status", String, "Current operational status of the tenant", func() {
			Enum("ACTIVE", "SUSPENDED", "TRIAL", "ARCHIVED", "PENDING_SETUP")
			Default("TRIAL")
			Example("ACTIVE")
			Description("Controls tenant access and billing")
		})
		
		Field(7, "timezone", String, "Default timezone for tenant operations", func() {
			Pattern("^[A-Za-z]+/[A-Za-z_]+$")
			Default("UTC")
			Example("America/New_York")
			Description("IANA timezone identifier")
		})
		
		Field(8, "currency_code", String, "Default currency for financial operations", func() {
			Pattern("^[A-Z]{3}$")
			Default("USD")
			Example("USD")
			Description("ISO 4217 currency code")
		})
		
		Field(9, "industry", String, "Business industry classification", func() {
			MaxLength(100)
			Example("Technology")
			Description("Used for industry-specific features and compliance")
		})
		
		Field(10, "company_size", String, "Organization size category", func() {
			Enum("STARTUP", "SMALL", "MEDIUM", "LARGE", "ENTERPRISE")
			Example("MEDIUM")
			Description("Affects feature availability and pricing tiers")
		})
		
		Field(11, "plan", String, "Current subscription plan", func() {
			Enum("FREE", "STARTER", "PROFESSIONAL", "ENTERPRISE", "CUSTOM")
			Default("FREE")
			Example("PROFESSIONAL")
			Description("Determines feature access and limits")
		})
		
		Field(12, "metadata", MapOf(String, Any), "Flexible metadata for tenant customization", func() {
			Example(map[string]any{
				"logo_url":        "https://cdn.acme.com/logo.png",
				"primary_color":   "#1E3A8A",
				"support_contact": "+1-555-123-4567",
			})
			Description("Custom key-value pairs for tenant configuration")
		})
		
		Field(13, "settings", MapOf(String, Any), "Tenant-specific system configuration", func() {
			Example(map[string]any{
				"two_factor_required":    true,
				"session_timeout_hours":  8,
				"allowed_ip_ranges":      []string{"192.168.1.0/24"},
				"data_retention_months":  36,
			})
			Description("Security and operational settings")
		})
		
		// Audit fields from common.go
		AuditFields()
		
		Required("id", "slug", "name", "email", "status", "timezone", "currency_code", "created_at")
	})
	
	View("default", func() {
		Description("Standard tenant view for listings")
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
		Attribute("subdomain")
		Attribute("status")
		Attribute("timezone")
		Attribute("currency_code")
		Attribute("industry")
		Attribute("company_size")
		Attribute("plan")
		Attribute("metadata")
		Attribute("settings")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})
	
	View("public", func() {
		Description("Public tenant information for branding")
		Attribute("slug")
		Attribute("name")
		Attribute("subdomain")
	})
})

// TenantConfiguration represents detailed operational configuration for a tenant.
// This includes limits, quotas, and business rules that govern tenant operations.
var TenantConfiguration = Type("TenantConfiguration", func() {
	Description("Comprehensive tenant configuration including limits, quotas, and operational parameters")
	
	Field(1, "max_users", UInt, "Maximum number of users allowed for this tenant", func() {
		Minimum(1)
		Maximum(10000)
		Default(10)
		Example(100)
		Description("Enforced during user creation")
	})
	
	Field(2, "max_entities", UInt, "Maximum number of organizational entities", func() {
		Minimum(1)
		Maximum(1000)
		Default(5)
		Example(25)
		Description("Limit on company/department hierarchies")
	})
	
	Field(3, "max_transactions_per_month", UInt, "Monthly transaction processing limit", func() {
		Minimum(100)
		Maximum(1000000)
		Default(1000)
		Example(50000)
		Description("Financial transaction quota per calendar month")
	})
	
	Field(4, "storage_quota_gb", UInt, "Storage quota in gigabytes", func() {
		Minimum(1)
		Maximum(10000)
		Default(10)
		Example(100)
		Description("File and document storage limit")
	})
	
	Field(5, "api_rate_limit_per_minute", UInt, "API requests per minute limit", func() {
		Minimum(10)
		Maximum(10000)
		Default(100)
		Example(500)
		Description("Rate limiting for API access")
	})
	
	Field(6, "accounting_method", String, "Preferred accounting methodology", func() {
		Enum("ACCRUAL", "CASH", "HYBRID")
		Default("ACCRUAL")
		Example("ACCRUAL")
		Description("Affects financial reporting and calculations")
	})
	
	Field(7, "fiscal_year_start_month", UInt, "Fiscal year start month (1-12)", func() {
		Minimum(1)
		Maximum(12)
		Default(1)
		Example(4)
		Description("Used for financial reporting periods")
	})
	
	Field(8, "default_currency", String, "Primary currency for operations", func() {
		Pattern("^[A-Z]{3}$")
		Default("USD")
		Example("USD")
		Description("ISO 4217 currency code")
	})
	
	Field(9, "supported_currencies", ArrayOf(String), "Additional supported currencies", func() {
		Elem(func() {
			Pattern("^[A-Z]{3}$")
		})
		MaxLength(10)
		Example([]string{"EUR", "GBP", "CAD"})
		Description("Multi-currency support for international operations")
	})
	
	Field(10, "password_policy", MapOf(String, Any), "Password security requirements", func() {
		Example(map[string]any{
			"min_length":         12,
			"require_uppercase":  true,
			"require_lowercase":  true,
			"require_numbers":    true,
			"require_symbols":    true,
			"max_age_days":       90,
			"history_count":      5,
		})
		Description("Enforced password complexity and rotation rules")
	})
	
	Field(11, "session_policy", MapOf(String, Any), "User session configuration", func() {
		Example(map[string]any{
			"timeout_minutes":        480,
			"idle_timeout_minutes":   30,
			"concurrent_sessions":    3,
			"require_mfa":           true,
		})
		Description("Session management and security settings")
	})
	
	Field(12, "backup_retention_days", UInt, "Data backup retention period", func() {
		Minimum(7)
		Maximum(2555) // ~7 years
		Default(90)
		Example(365)
		Description("Number of days to retain backup data")
	})
	
	Field(13, "audit_log_retention_days", UInt, "Audit log retention period", func() {
		Minimum(30)
		Maximum(2555) // ~7 years
		Default(365)
		Example(2555)
		Description("Compliance requirement for audit trail retention")
	})
	
	Field(14, "features_enabled", ArrayOf(String), "Enabled feature flags", func() {
		Example([]string{
			"advanced_reporting",
			"multi_entity_consolidation",
			"api_access",
			"mobile_app",
			"integration_webhooks",
		})
		Description("List of enabled features for this tenant")
	})
	
	Field(15, "integrations_allowed", ArrayOf(String), "Permitted third-party integrations", func() {
		Example([]string{
			"salesforce",
			"quickbooks",
			"stripe",
			"slack",
		})
		Description("Approved external system integrations")
	})
	
	// Audit fields
	AuditFields()
	
	Required("max_users", "max_entities", "max_transactions_per_month", "storage_quota_gb", 
		"accounting_method", "fiscal_year_start_month", "default_currency")
})

// TenantUsageStats tracks tenant resource consumption and limits
var TenantUsageStats = Type("TenantUsageStats", func() {
	Description("Real-time usage statistics and quota consumption for tenant")
	
	Field(1, "tenant_id", String, "Tenant identifier", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	
	Field(2, "current_users", UInt, "Active user count", func() {
		Example(45)
		Description("Currently active (non-deleted) users")
	})
	
	Field(3, "current_entities", UInt, "Active entity count", func() {
		Example(8)
		Description("Organizational entities in use")
	})
	
	Field(4, "transactions_this_month", UInt, "Current month transaction count", func() {
		Example(1250)
		Description("Financial transactions processed this calendar month")
	})
	
	Field(5, "storage_used_gb", Float64, "Storage consumption in GB", func() {
		Minimum(0)
		Example(15.7)
		Description("Current storage usage including files and database")
	})
	
	Field(6, "api_calls_today", UInt, "API requests made today", func() {
		Example(2847)
		Description("API usage for current day")
	})
	
	Field(7, "last_login", String, "Most recent user login timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
		Description("Latest user activity timestamp")
	})
	
	Field(8, "calculated_at", String, "Statistics calculation timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T15:45:00Z")
		Description("When these statistics were last computed")
	})
	
	Required("tenant_id", "current_users", "current_entities", "transactions_this_month", 
		"storage_used_gb", "calculated_at")
})