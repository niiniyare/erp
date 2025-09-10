package organization

import (
	"github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// =============================================================================
// ENUM DEFINITIONS
// =============================================================================

var EntityType = Type("EntityType", String, func() {
	Description("Type of entity in the organizational hierarchy")
	Enum("COMPANY", "SUBSIDIARY", "REGION", "BRANCH", "LOCATION", "DEPARTMENT", "DIVISION", "COST_CENTER", "PROJECT", "BUDGET_UNIT")
	Example("COMPANY")
})

var OrganizationStatus = Type("OrganizationStatus", String, func() {
	Description("Status of organization")
	Enum("ACTIVE", "INACTIVE", "SUSPENDED", "DISSOLVED")
	Example("ACTIVE")
})

var AddressType = Type("AddressType", String, func() {
	Description("Type of address")
	Enum("HEADQUARTERS", "BRANCH", "BILLING", "SHIPPING", "MAILING", "WAREHOUSE")
	Example("HEADQUARTERS")
})

// =============================================================================
// RESULT TYPES
// =============================================================================

// OrganizationResult describes the organization response
var OrganizationResult = ResultType("application/vnd.organization", func() {
	Description("Organization information")
	Attributes(func() {
		Attribute("id", String, "Unique organization identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
		})
		Attribute("name", String, "Organization name", func() {
			MinLength(1)
			MaxLength(200)
			Example("Acme Corporation")
		})
		Attribute("legal_name", String, "Legal business name", func() {
			MaxLength(200)
			Example("Acme Corporation Inc.")
		})
		Attribute("display_name", String, "Display name", func() {
			MaxLength(200)
			Example("ACME Corp")
		})
		Attribute("entity_type", EntityType, "Type of entity in hierarchy")
	Attribute("legal_entity_type", String, "Legal entity type (Corporation, LLC, etc.)", func() {
		Enum("CORPORATION", "LLC", "PARTNERSHIP", "SOLE_PROPRIETORSHIP", "NON_PROFIT", "GOVERNMENT", "OTHER")
		Example("CORPORATION")
	})
		Attribute("status", OrganizationStatus, "Organization status")
		Attribute("tax_id", String, "Tax identification number", func() {
			Pattern("^[0-9-]+$")
			Example("12-3456789")
		})
		Attribute("registration_number", String, "Business registration number", func() {
			MaxLength(50)
			Example("REG123456789")
		})
		Attribute("incorporation_date", String, "Date of incorporation", func() {
			Format(FormatDate)
			Example("2020-01-15")
		})
		Attribute("jurisdiction", String, "Jurisdiction of incorporation", func() {
			MaxLength(100)
			Example("Delaware, USA")
		})
		Attribute("description", String, "Organization description", func() {
			MaxLength(1000)
			Example("Leading provider of innovative business solutions")
		})
		Attribute("website", String, "Organization website", func() {
			Format(FormatURI)
			Example("https://www.acme.com")
		})
		Attribute("industry", String, "Industry classification", func() {
			MaxLength(100)
			Example("Technology")
		})
		Attribute("employee_count", UInt, "Number of employees", func() {
			Example(500)
		})
		Attribute("parent_id", String, "Parent organization ID", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440001")
		})
		Attribute("addresses", ArrayOf(AddressInfo), "Organization addresses")
		Attribute("contacts", ArrayOf(types.ContactInfo), "Organization contacts")
		Attribute("settings", OrganizationSettings, "Organization settings")
		types.AuditFields()
	})
	Required("id", "name", "entity_type", "status", "created_at", "updated_at")

	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("legal_name")
		Attribute("display_name")
		Attribute("entity_type")
		Attribute("legal_entity_type")
		Attribute("status")
		Attribute("tax_id")
		Attribute("registration_number")
		Attribute("incorporation_date")
		Attribute("jurisdiction")
		Attribute("description")
		Attribute("website")
		Attribute("industry")
		Attribute("employee_count")
		Attribute("parent_id")
		Attribute("addresses")
		Attribute("contacts")
		Attribute("settings")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})

	View("minimal", func() {
		Attribute("id")
		Attribute("name")
		Attribute("entity_type")
		Attribute("legal_entity_type")
		Attribute("status")
	})

	View("tree", func() {
		Attribute("id")
		Attribute("name")
		Attribute("display_name")
		Attribute("entity_type")
		Attribute("legal_entity_type")
		Attribute("status")
		Attribute("parent_id")
	})
})

// HierarchyResult describes organization hierarchy response
var HierarchyResult = ResultType("application/vnd.organization.hierarchy", func() {
	Description("Organization hierarchy information")
	Attributes(func() {
		Attribute("root", OrganizationNode, "Root organization")
		Attribute("children", ArrayOf(OrganizationNode), "Child organizations")
		Attribute("ancestors", ArrayOf(OrganizationNode), "Ancestor organizations")
		Attribute("depth", UInt, "Current hierarchy depth", func() {
			Example(3)
		})
	})
	Required("root", "children", "ancestors", "depth")
})

// =============================================================================
// PAYLOAD TYPES
// =============================================================================

// CreateOrganizationPayload describes the payload for creating an organization
var CreateOrganizationPayload = Type("CreateOrganizationPayload", func() {
	Description("Payload for creating a new organization")
	Attribute("name", String, "Organization name", func() {
		MinLength(1)
		MaxLength(200)
		Example("Acme Corporation")
	})
	Attribute("legal_name", String, "Legal business name", func() {
		MaxLength(200)
		Example("Acme Corporation Inc.")
	})
	Attribute("display_name", String, "Display name", func() {
		MaxLength(200)
		Example("ACME Corp")
	})
	Attribute("entity_type", EntityType, "Type of entity in hierarchy")
	Attribute("legal_entity_type", String, "Legal entity type (Corporation, LLC, etc.)", func() {
		Enum("CORPORATION", "LLC", "PARTNERSHIP", "SOLE_PROPRIETORSHIP", "NON_PROFIT", "GOVERNMENT", "OTHER")
		Example("CORPORATION")
	})
	Attribute("tax_id", String, "Tax identification number", func() {
		Pattern("^[0-9-]+$")
		Example("12-3456789")
	})
	Attribute("registration_number", String, "Business registration number", func() {
		MaxLength(50)
		Example("REG123456789")
	})
	Attribute("incorporation_date", String, "Date of incorporation", func() {
		Format(FormatDate)
		Example("2020-01-15")
	})
	Attribute("jurisdiction", String, "Jurisdiction of incorporation", func() {
		MaxLength(100)
		Example("Delaware, USA")
	})
	Attribute("description", String, "Organization description", func() {
		MaxLength(1000)
		Example("Leading provider of innovative business solutions")
	})
	Attribute("website", String, "Organization website", func() {
		Format(FormatURI)
		Example("https://www.acme.com")
	})
	Attribute("industry", String, "Industry classification", func() {
		MaxLength(100)
		Example("Technology")
	})
	Attribute("employee_count", UInt, "Number of employees", func() {
		Example(500)
	})
	Attribute("parent_id", String, "Parent organization ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440001")
	})
	Attribute("addresses", ArrayOf(AddressInfo), "Organization addresses")
	Attribute("contacts", ArrayOf(types.ContactInfo), "Organization contacts")
	Attribute("settings", OrganizationSettings, "Organization settings")
	Required("name", "entity_type")
})

// UpdateOrganizationPayload describes the payload for updating an organization
var UpdateOrganizationPayload = Type("UpdateOrganizationPayload", func() {
	Description("Payload for updating an existing organization")
	Attribute("id", String, "Organization ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("name", String, "Organization name", func() {
		MinLength(1)
		MaxLength(200)
		Example("Acme Corporation")
	})
	Attribute("legal_name", String, "Legal business name", func() {
		MaxLength(200)
		Example("Acme Corporation Inc.")
	})
	Attribute("display_name", String, "Display name", func() {
		MaxLength(200)
		Example("ACME Corp")
	})
	Attribute("entity_type", EntityType, "Type of entity in hierarchy")
	Attribute("legal_entity_type", String, "Legal entity type (Corporation, LLC, etc.)", func() {
		Enum("CORPORATION", "LLC", "PARTNERSHIP", "SOLE_PROPRIETORSHIP", "NON_PROFIT", "GOVERNMENT", "OTHER")
		Example("CORPORATION")
	})
	Attribute("status", OrganizationStatus, "Organization status")
	Attribute("tax_id", String, "Tax identification number", func() {
		Pattern("^[0-9-]+$")
		Example("12-3456789")
	})
	Attribute("registration_number", String, "Business registration number", func() {
		MaxLength(50)
		Example("REG123456789")
	})
	Attribute("incorporation_date", String, "Date of incorporation", func() {
		Format(FormatDate)
		Example("2020-01-15")
	})
	Attribute("jurisdiction", String, "Jurisdiction of incorporation", func() {
		MaxLength(100)
		Example("Delaware, USA")
	})
	Attribute("description", String, "Organization description", func() {
		MaxLength(1000)
		Example("Leading provider of innovative business solutions")
	})
	Attribute("website", String, "Organization website", func() {
		Format(FormatURI)
		Example("https://www.acme.com")
	})
	Attribute("industry", String, "Industry classification", func() {
		MaxLength(100)
		Example("Technology")
	})
	Attribute("employee_count", UInt, "Number of employees", func() {
		Example(500)
	})
	Attribute("parent_id", String, "Parent organization ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440001")
	})
	Attribute("addresses", ArrayOf(AddressInfo), "Organization addresses")
	Attribute("contacts", ArrayOf(types.ContactInfo), "Organization contacts")
	Attribute("settings", OrganizationSettings, "Organization settings")
	Required("id")
})

// =============================================================================
// NESTED TYPES
// =============================================================================

// OrganizationNode represents a node in the organization hierarchy
var OrganizationNode = Type("OrganizationNode", func() {
	Description("Organization node in hierarchy")
	Attribute("id", String, "Organization ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("name", String, "Organization name", func() {
		MinLength(1)
		MaxLength(200)
		Example("Acme Corporation")
	})
	Attribute("entity_type", String, "Type of entity", func() {
		Enum("COMPANY", "SUBSIDIARY", "REGION", "BRANCH", "LOCATION", "DEPARTMENT", "DIVISION", "COST_CENTER", "PROJECT", "BUDGET_UNIT")
		Example("COMPANY")
	})
	Attribute("status", String, "Organization status", func() {
		Enum("ACTIVE", "INACTIVE", "SUSPENDED", "DISSOLVED")
		Example("ACTIVE")
	})
	Attribute("children", ArrayOf(String), "Child organization IDs")
	Attribute("level", UInt, "Level in hierarchy", func() {
		Example(1)
	})
	Required("id", "name", "entity_type", "status", "children", "level")
})

// AddressInfo describes address information
var AddressInfo = Type("AddressInfo", func() {
	Description("Address information")
	Attribute("type", AddressType, "Address type")
	Attribute("street_address_1", String, "Street address line 1", func() {
		MaxLength(200)
		Example("123 Main Street")
	})
	Attribute("street_address_2", String, "Street address line 2", func() {
		MaxLength(200)
		Example("Suite 456")
	})
	Attribute("city", String, "City", func() {
		MaxLength(100)
		Example("New York")
	})
	Attribute("state_province", String, "State or province", func() {
		MaxLength(100)
		Example("NY")
	})
	Attribute("postal_code", String, "Postal code", func() {
		MaxLength(20)
		Example("10001")
	})
	Attribute("country", String, "Country", func() {
		MaxLength(100)
		Example("United States")
	})
	Attribute("country_code", String, "ISO country code", func() {
		Pattern("^[A-Z]{2}$")
		Example("US")
	})
	Attribute("is_primary", Boolean, "Whether this is the primary address", func() {
		Default(false)
		Example(true)
	})
	Required("type", "street_address_1", "city", "country")
})

// OrganizationSettings describes organization-specific settings
var OrganizationSettings = Type("OrganizationSettings", func() {
	Description("Organization-specific settings and configuration")
	Attribute("timezone", String, "Default timezone", func() {
		Example("America/New_York")
		Default("UTC")
	})
	Attribute("currency", String, "Default currency code", func() {
		Pattern("^[A-Z]{3}$")
		Example("USD")
		Default("USD")
	})
	Attribute("fiscal_year_start", String, "Fiscal year start date", func() {
		Pattern("^(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$")
		Example("01-01")
		Default("01-01")
	})
	Attribute("business_hours", BusinessHours, "Business operating hours")
	Attribute("integrations", MapOf(String, Any), "Third-party integrations config")
	Attribute("preferences", MapOf(String, Any), "Organization preferences")
})

// BusinessHours describes business operating hours
var BusinessHours = Type("BusinessHours", func() {
	Description("Business operating hours")
	Attribute("monday", DayHours, "Monday hours")
	Attribute("tuesday", DayHours, "Tuesday hours")
	Attribute("wednesday", DayHours, "Wednesday hours")
	Attribute("thursday", DayHours, "Thursday hours")
	Attribute("friday", DayHours, "Friday hours")
	Attribute("saturday", DayHours, "Saturday hours")
	Attribute("sunday", DayHours, "Sunday hours")
})

// DayHours describes operating hours for a specific day
var DayHours = Type("DayHours", func() {
	Description("Operating hours for a specific day")
	Attribute("is_open", Boolean, "Whether open on this day", func() {
		Default(true)
		Example(true)
	})
	Attribute("open_time", String, "Opening time", func() {
		Pattern("^([01]?[0-9]|2[0-3]):[0-5][0-9]$")
		Example("09:00")
	})
	Attribute("close_time", String, "Closing time", func() {
		Pattern("^([01]?[0-9]|2[0-3]):[0-5][0-9]$")
		Example("17:00")
	})
	Attribute("break_start", String, "Break start time", func() {
		Pattern("^([01]?[0-9]|2[0-3]):[0-5][0-9]$")
		Example("12:00")
	})
	Attribute("break_end", String, "Break end time", func() {
		Pattern("^([01]?[0-9]|2[0-3]):[0-5][0-9]$")
		Example("13:00")
	})
})
