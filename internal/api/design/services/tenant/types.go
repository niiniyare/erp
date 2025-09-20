package tenant

import (
	. "github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// TenantResult describes the tenant response
var TenantResult = ResultType("application/vnd.tenant", func() {
	Description("Tenant information")
	Attributes(func() {
		Attribute("id", String, "Unique tenant identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
		})
		Attribute("name", String, "Tenant display name", func() {
			MinLength(1)
			MaxLength(100)
			Example("Acme Corporation")
		})
		Attribute("slug", String, "Tenant URL slug", func() {
			Pattern("^[a-z0-9-]+$")
			MinLength(3)
			MaxLength(50)
			Example("acme-corp")
		})
		Attribute("subdomain", String, "Subdomain for tenant", func() {
			Pattern("^[a-z0-9-]+$")
			MinLength(3)
			MaxLength(50)
			Example("acme")
		})
		Attribute("status", String, "Tenant status", func() {
			Enum("active", "inactive", "suspended", "pending")
			Example("active")
		})
		Attribute("plan_type", String, "Subscription plan", func() {
			Enum("starter", "professional", "enterprise")
			Example("professional")
		})
		Attribute("description", String, "Tenant description", func() {
			MaxLength(500)
			Example("Leading provider of roadrunner traps and anvils")
		})
		Attribute("settings", TenantSettings, "Tenant-specific settings")
		Attribute("subscription", SubscriptionInfo, "Subscription information")
		Attribute("contact", ContactInfo, "Primary contact information")
		AuditFields()
	})
	Required("id", "name", "slug", "status", "plan_type", "created_at", "updated_at")

	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("slug")
		Attribute("subdomain")
		Attribute("status")
		Attribute("plan_type")
		Attribute("description")
		Attribute("settings")
		Attribute("subscription")
		Attribute("contact")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})

	View("minimal", func() {
		Attribute("id")
		Attribute("name")
		Attribute("subdomain")
		Attribute("status")
	})
})

// CreateTenantPayload describes the payload for creating a tenant
var CreateTenantPayload = Type("CreateTenantPayload", func() {
	Description("Payload for creating a new tenant")

	Attribute("name", String, "Tenant display name", func() {
		MinLength(2)
		MaxLength(100)
		Example("Acme Corporation")
	})
	Attribute("slug", String, "Tenant unique slug", func() {
		Pattern("^[a-z0-9-]+$")
		Example("acme-corp")
	})
	Attribute("email", String, "Primary email of the tenant", func() {
		Format(FormatEmail)
		Example("admin@acme.com")
	})
	Attribute("subdomain", String, "Desired subdomain (optional)", func() {
		Pattern("^[a-z0-9-]+$")
		MinLength(3)
		MaxLength(50)
		Example("acme")
	})
	Attribute("status", String, "Tenant status", func() {
		Enum("active", "inactive", "suspended")
		Default("active")
		Example("active")
	})
	Attribute("plan_type", String, "Subscription plan type", func() {
		Enum("starter", "professional", "enterprise")
		Default("starter")
		Example("professional")
	})
	Attribute("industry", String, "Industry of the tenant", func() {
		Example("Technology")
	})
	Attribute("company_size", String, "Size of the company", func() {
		Enum("1-10", "11-50", "51-200", "201-500", "500+")
		Example("51-200")
	})
	Attribute("tax_id", String, "Tax Identification Number", func() {
		Example("123-45-6789")
	})
	Attribute("registration_number", String, "Company registration number", func() {
		Example("REG-987654321")
	})
	Attribute("legal_entity_type", String, "Legal entity type", func() {
		Example("LLC")
	})
	Attribute("contact", ContactInfo, "Primary contact information")
	Attribute("settings", TenantSettings, "Initial tenant settings")
	Attribute("country_code", String, "ISO 3166-1 alpha-2 country code", func() {
		Pattern("^[A-Z]{2}$")
		Example("US")
	})
	Attribute("currency_code", String, "ISO 4217 currency code", func() {
		Pattern("^[A-Z]{3}$")
		Example("USD")
	})

	// Required fields
	Required("name", "email", "country_code", "currency_code")
})

// UpdateTenantPayload describes the payload for updating a tenant
var UpdateTenantPayload = Type("UpdateTenantPayload", func() {
	Description("Payload for updating an existing tenant")
	Attribute("id", String, "Tenant ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("name", String, "Tenant display name", func() {
		MinLength(1)
		MaxLength(100)
		Example("Acme Corporation Updated")
	})
	Attribute("description", String, "Tenant description", func() {
		MaxLength(500)
		Example("Updated description")
	})
	Attribute("status", String, "Tenant status", func() {
		Enum("active", "inactive", "suspended")
		Example("active")
	})
	Attribute("plan_type", String, "Subscription plan type", func() {
		Enum("starter", "professional", "enterprise")
		Example("professional")
	})
	Attribute("contact", ContactInfo, "Primary contact information")
	Attribute("settings", TenantSettings, "Tenant settings")
	Required("id")
})

// TenantSettings describes tenant-specific configuration
var TenantSettings = Type("TenantSettings", func() {
	Description("Tenant-specific settings and configuration")
	Attribute("timezone", String, "Default timezone", func() {
		Example("America/New_York")
		Default("UTC")
	})
	Attribute("currency", String, "Default currency code", func() {
		Pattern("^[A-Z]{3}$")
		Example("USD")
		Default("USD")
	})
	Attribute("date_format", String, "Preferred date format", func() {
		Enum("MM/DD/YYYY", "DD/MM/YYYY", "YYYY-MM-DD")
		Example("MM/DD/YYYY")
		Default("MM/DD/YYYY")
	})
	Attribute("language", String, "Default language", func() {
		Pattern("^[a-z]{2}$")
		Example("en")
		Default("en")
	})
	Attribute("features", ArrayOf(String), "Enabled features", func() {
		Example([]string{"inventory", "financial", "hr"})
	})
	Attribute("limits", TenantLimits, "Usage limits")
})

// TenantLimits describes usage limits for the tenant
var TenantLimits = Type("TenantLimits", func() {
	Description("Usage limits for tenant")
	Attribute("max_users", UInt, "Maximum number of users", func() {
		Minimum(1)
		Example(100)
	})
	Attribute("max_storage_mb", UInt, "Maximum storage in MB", func() {
		Minimum(100)
		Example(10240) // 10GB
	})
	Attribute("max_api_calls_per_hour", UInt, "API rate limit per hour", func() {
		Minimum(100)
		Example(10000)
	})
})

// SubscriptionInfo describes subscription details
var SubscriptionInfo = Type("SubscriptionInfo", func() {
	Description("Tenant subscription information")
	Attribute("plan", String, "Subscription plan", func() {
		Enum("starter", "professional", "enterprise")
		Example("professional")
	})
	Attribute("status", String, "Subscription status", func() {
		Enum("active", "past_due", "canceled", "trialing")
		Example("active")
	})
	Attribute("billing_cycle", String, "Billing cycle", func() {
		Enum("monthly", "yearly")
		Example("monthly")
	})
	Attribute("next_billing_date", String, "Next billing date", func() {
		Format(FormatDate)
		Example("2023-12-07")
	})
	Attribute("trial_ends_at", String, "Trial end date", func() {
		Format(FormatDate)
		Example("2023-12-07")
	})
})
