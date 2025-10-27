package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// ORGANIZATIONAL ENTITY TYPES
// ============================================================================

// Entity represents an organizational unit within a tenant's business structure.
// Entities form hierarchical relationships (companies, divisions, departments, etc.)
// and serve as the foundation for access control and financial reporting.
var Entity = ResultType("application/vnd.erp.entity", func() {
	Description("Organizational entity representing companies, departments, cost centers, and other business units with hierarchical relationships")

	Attributes(func() {
		Field(1, "uuid", String, "Unique entity identifier", func() {
			Format(FormatUUID)
			Example("987fcdeb-51d2-43b8-a456-426614174000")
			Description("Primary key for entity references")
		})

		Field(2, "tenant_id", String, "Parent tenant identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
			Description("Tenant isolation boundary")
		})

		Field(3, "parent_id", String, "Parent entity identifier for hierarchy", func() {
			Format(FormatUUID)
			Example("123e4567-e89b-12d3-a456-426614174000")
			Description("Creates organizational hierarchy; null for root entities")
		})

		Field(4, "name", String, "Entity display name", func() {
			MinLength(2)
			MaxLength(100)
			Example("Engineering Department")
			Description("Human-readable name shown in UI")
		})

		Field(5, "code", String, "Internal reference code for entity", func() {
			Pattern("^[A-Z0-9][A-Z0-9_-]{1,18}[A-Z0-9]$")
			MinLength(2)
			MaxLength(20)
			Example("ENG-001")
			Description("Unique alphanumeric code for integration and reporting")
		})

		Field(6, "type", String, "Entity classification type", func() {
			Enum("COMPANY", "SUBSIDIARY", "DIVISION", "DEPARTMENT", "COST_CENTER",
				"PROJECT", "REGION", "BRANCH", "TEAM", "UNIT")
			Example("DEPARTMENT")
			Description("Determines entity behavior and available features")
		})

		Field(7, "is_active", Boolean, "Active operational status", func() {
			Default(true)
			Example(true)
			Description("Controls entity visibility and transaction processing")
		})

		Field(8, "hidden", Boolean, "Visibility flag for UI display", func() {
			Default(false)
			Example(false)
			Description("Hides entity from standard user interfaces")
		})

		Field(9, "description", String, "Detailed entity description", func() {
			MaxLength(500)
			Example("Software engineering and development teams responsible for product development")
			Description("Optional detailed description of entity purpose")
		})

		Field(10, "accrual_method", Boolean, "Accounting method preference", func() {
			Default(true)
			Example(true)
			Description("True for accrual accounting, false for cash basis")
		})

		Field(11, "fy_start_month", UInt, "Fiscal year start month (1-12)", func() {
			Minimum(1)
			Maximum(12)
			Default(1)
			Example(4)
			Description("Entity-specific fiscal year for financial reporting")
		})

		Field(12, "currency_code", String, "Primary operating currency", func() {
			Pattern("^[A-Z]{3}$")
			Default("USD")
			Example("USD")
			Description("ISO 4217 currency code for financial operations")
		})

		Field(13, "address", Address, "Physical business address", func() {
			Description("Primary business location for this entity")
		})

		Field(14, "contact_info", ContactInfo, "Primary contact information", func() {
			Description("Main contact details for entity communications")
		})

		Field(15, "settings", MapOf(String, Any), "Entity-specific configuration", func() {
			Example(map[string]any{
				"auto_approve_limit":    1000.00,
				"require_dual_approval": true,
				"cost_center_tracking":  true,
				"budget_alerts_enabled": true,
			})
			Description("Flexible configuration options for entity behavior")
		})

		Field(16, "metadata", MapOf(String, Any), "Custom entity metadata", func() {
			Example(map[string]any{
				"manager_employee_id": "emp-123456",
				"budget_code":         "BUD-ENG-2024",
				"external_system_id":  "SAP-DEPT-789",
			})
			Description("Additional data for integrations and custom fields")
		})

		Field(17, "validation_status", String, "Entity validation state", func() {
			Enum("VALID", "PENDING", "INVALID", "NEEDS_REVIEW")
			Default("PENDING")
			Example("VALID")
			Description("Validation status for compliance and data integrity")
		})

		Field(18, "level", UInt, "Hierarchical depth level", func() {
			Minimum(0)
			Maximum(10)
			Example(2)
			Description("Depth in organizational hierarchy (0 = root)")
		})

		Field(19, "path", String, "Hierarchical path from root", func() {
			Pattern("^(/[A-Z0-9_-]+)+$")
			Example("/ACME/ENGINEERING/BACKEND")
			Description("Full path showing hierarchy from root to current entity")
		})

		// Audit fields from common.go
		AuditFields()

		Required("uuid", "tenant_id", "name", "type", "is_active", "created_at")
	})

	View("default", func() {
		Description("Standard entity view for listings and general use")
		Attribute("uuid")
		Attribute("tenant_id")
		Attribute("parent_id")
		Attribute("name")
		Attribute("code")
		Attribute("type")
		Attribute("is_active")
		Attribute("level")
		Attribute("created_at")
	})

	View("detailed", func() {
		Description("Complete entity view with all configuration and metadata")
		Attribute("uuid")
		Attribute("tenant_id")
		Attribute("parent_id")
		Attribute("name")
		Attribute("code")
		Attribute("type")
		Attribute("is_active")
		Attribute("hidden")
		Attribute("description")
		Attribute("accrual_method")
		Attribute("fy_start_month")
		Attribute("currency_code")
		Attribute("address")
		Attribute("contact_info")
		Attribute("settings")
		Attribute("metadata")
		Attribute("validation_status")
		Attribute("level")
		Attribute("path")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})

	View("hierarchy", func() {
		Description("Hierarchical view for organizational charts")
		Attribute("uuid")
		Attribute("parent_id")
		Attribute("name")
		Attribute("code")
		Attribute("type")
		Attribute("level")
		Attribute("path")
		Attribute("is_active")
	})

	View("summary", func() {
		Description("Minimal view for references and dropdowns")
		Attribute("uuid")
		Attribute("name")
		Attribute("code")
		Attribute("type")
	})
})

// EntityHierarchy represents hierarchical relationships between entities.
// This type supports efficient querying of organizational structures
// and is used for closure table implementations.
var EntityHierarchy = Type("EntityHierarchy", func() {
	Description("Hierarchical relationship mapping for efficient organizational structure queries")

	Field(1, "entity_id", String, "Target entity identifier", func() {
		Format(FormatUUID)
		Example("987fcdeb-51d2-43b8-a456-426614174000")
		Description("The entity for which relationships are defined")
	})

	Field(2, "ancestor_id", String, "Ancestor entity identifier", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
		Description("Parent entity in the hierarchy")
	})

	Field(3, "descendant_id", String, "Descendant entity identifier", func() {
		Format(FormatUUID)
		Example("456e7890-e89b-12d3-a456-426614174000")
		Description("Child entity in the hierarchy")
	})

	Field(4, "depth", UInt, "Hierarchical depth between entities", func() {
		Minimum(0)
		Maximum(10)
		Example(2)
		Description("Number of levels between ancestor and descendant (0 = same entity)")
	})

	Field(5, "path", ArrayOf(String), "Full path of entity IDs from ancestor to descendant", func() {
		Example([]string{
			"123e4567-e89b-12d3-a456-426614174000",
			"987fcdeb-51d2-43b8-a456-426614174000",
			"456e7890-e89b-12d3-a456-426614174000",
		})
		Description("Ordered list of entity IDs forming the hierarchical path")
	})

	Field(6, "is_direct", Boolean, "Direct parent-child relationship flag", func() {
		Example(true)
		Description("True if entities have direct parent-child relationship")
	})

	Required("entity_id", "ancestor_id", "descendant_id", "depth")
})

// EntityStats provides statistical information about entity usage and performance
var EntityStats = Type("EntityStats", func() {
	Description("Statistical information and metrics for entity performance monitoring")

	Field(1, "entity_id", String, "Entity identifier", func() {
		Format(FormatUUID)
		Example("987fcdeb-51d2-43b8-a456-426614174000")
	})

	Field(2, "user_count", UInt, "Number of users assigned to entity", func() {
		Example(45)
		Description("Active users with access to this entity")
	})

	Field(3, "child_entity_count", UInt, "Number of direct child entities", func() {
		Example(5)
		Description("Immediate subordinate entities")
	})

	Field(4, "total_descendant_count", UInt, "Total descendant entities", func() {
		Example(23)
		Description("All entities in the subtree")
	})

	Field(5, "transaction_count_ytd", UInt, "Year-to-date transaction count", func() {
		Example(1847)
		Description("Financial transactions processed this fiscal year")
	})

	Field(6, "last_activity", String, "Most recent activity timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T14:30:00Z")
		Description("Latest user or system activity in this entity")
	})

	Field(7, "budget_utilization_percent", Float64, "Budget utilization percentage", func() {
		Minimum(0)
		Maximum(1000) // Allow over-budget scenarios
		Example(78.5)
		Description("Percentage of annual budget utilized")
	})

	Field(8, "calculated_at", String, "Statistics calculation timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T15:00:00Z")
		Description("When these statistics were last computed")
	})

	Required("entity_id", "user_count", "child_entity_count", "total_descendant_count", "calculated_at")
})

// EntityValidationRule defines validation rules for entity data integrity
var EntityValidationRule = Type("EntityValidationRule", func() {
	Description("Validation rules for ensuring entity data integrity and compliance")

	Field(1, "rule_id", String, "Unique rule identifier", func() {
		Format(FormatUUID)
		Example("rule-123e4567-e89b-12d3-a456-426614174000")
	})

	Field(2, "rule_name", String, "Human-readable rule name", func() {
		MinLength(3)
		MaxLength(100)
		Example("Department Code Format")
		Description("Descriptive name for the validation rule")
	})

	Field(3, "entity_type", String, "Entity type this rule applies to", func() {
		Enum("COMPANY", "SUBSIDIARY", "DIVISION", "DEPARTMENT", "COST_CENTER",
			"PROJECT", "REGION", "BRANCH", "TEAM", "UNIT", "ALL")
		Example("DEPARTMENT")
		Description("Scope of entities affected by this rule")
	})

	Field(4, "field_name", String, "Entity field to validate", func() {
		Enum("name", "code", "type", "parent_id", "currency_code", "settings", "metadata")
		Example("code")
		Description("Specific entity field this rule validates")
	})

	Field(5, "validation_type", String, "Type of validation to perform", func() {
		Enum("PATTERN", "LENGTH", "UNIQUENESS", "REQUIRED", "CUSTOM", "REFERENCE")
		Example("PATTERN")
		Description("Validation method to apply")
	})

	Field(6, "rule_definition", MapOf(String, Any), "Detailed rule configuration", func() {
		Example(map[string]any{
			"pattern":       "^DEPT-[0-9]{3}$",
			"error_message": "Department codes must follow format: DEPT-XXX",
			"severity":      "ERROR",
		})
		Description("Flexible rule definition supporting various validation types")
	})

	Field(7, "is_active", Boolean, "Rule activation status", func() {
		Default(true)
		Example(true)
		Description("Controls whether rule is enforced")
	})

	Field(8, "priority", UInt, "Rule execution priority", func() {
		Minimum(1)
		Maximum(100)
		Default(50)
		Example(10)
		Description("Lower numbers execute first (1 = highest priority)")
	})

	// Audit fields
	AuditFields()

	Required("rule_id", "rule_name", "entity_type", "field_name", "validation_type", "rule_definition")
})
