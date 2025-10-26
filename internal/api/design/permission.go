package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// PERMISSION & ACCESS CONTROL TYPES
// ============================================================================

// Permission represents a specific access right that can be granted to roles.
// Permissions define atomic operations on resources and support ABAC conditions
// for fine-grained access control.
var Permission = ResultType("application/vnd.erp.permission", func() {
	Description("Atomic permission defining specific access rights to resources with ABAC condition support")
	
	Attributes(func() {
		Field(1, "id", String, "Unique permission identifier", func() {
			Format(FormatUUID)
			Example("perm-123e4567-e89b-12d3-a456-426614174000")
			Description("Primary key for permission references")
		})
		
		Field(2, "tenant_id", String, "Associated tenant identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
			Description("Tenant isolation boundary for permission scope")
		})
		
		Field(3, "resource_id", String, "Target resource identifier", func() {
			Format(FormatUUID)
			Example("resource-456e7890-e89b-12d3-a456-426614174000")
			Description("Resource this permission applies to")
		})
		
		Field(4, "action_id", String, "Action identifier", func() {
			Format(FormatUUID)
			Example("action-789abc12-def3-4567-890a-bcdef1234567")
			Description("Specific action that can be performed")
		})
		
		Field(5, "name", String, "Unique permission name", func() {
			Pattern("^[a-z][a-z0-9_.]{2,99}$")
			MinLength(3)
			MaxLength(100)
			Example("finance.accounts.read")
			Description("Hierarchical dot-notation permission identifier")
		})
		
		Field(6, "display_name", String, "Human-readable permission title", func() {
			MinLength(3)
			MaxLength(100)
			Example("Read Financial Accounts")
			Description("User-friendly name for UI display")
		})
		
		Field(7, "description", String, "Detailed permission description", func() {
			MaxLength(500)
			Example("Allows viewing financial account details, balances, and transaction history")
			Description("Comprehensive explanation of what this permission grants")
		})
		
		Field(8, "resource_type", String, "Type of resource being protected", func() {
			Pattern("^[a-z][a-z0-9_]{2,49}$")
			Example("financial_account")
			Description("Category of resource for grouping permissions")
		})
		
		Field(9, "action_type", String, "Type of action being permitted", func() {
			Enum("CREATE", "READ", "UPDATE", "DELETE", "EXECUTE", "APPROVE", "EXPORT", 
				"IMPORT", "ARCHIVE", "RESTORE", "ASSIGN", "UNASSIGN")
			Example("READ")
			Description("Standard CRUD+ action classification")
		})
		
		Field(10, "effect", String, "Permission effect on access decision", func() {
			Enum("ALLOW", "DENY")
			Default("ALLOW")
			Example("ALLOW")
			Description("Whether this permission grants or denies access")
		})
		
		Field(11, "scope", String, "Permission application scope", func() {
			Enum("GLOBAL", "TENANT", "ENTITY", "DEPARTMENT", "PROJECT", "PERSONAL")
			Default("ENTITY")
			Example("ENTITY")
			Description("Boundary within which permission is effective")
		})
		
		Field(12, "priority", UInt, "Permission priority for conflict resolution", func() {
			Minimum(1)
			Maximum(1000)
			Default(100)
			Example(200)
			Description("Higher numbers take precedence in conflicts")
		})
		
		Field(13, "conditions", MapOf(String, Any), "ABAC conditions for dynamic evaluation", func() {
			Example(map[string]any{
				"attribute_constraints": map[string]any{
					"user.department":    []string{"finance", "accounting"},
					"resource.status":    "active",
					"time.business_hours": true,
				},
				"context_requirements": map[string]any{
					"ip_whitelist":       []string{"192.168.1.0/24"},
					"require_mfa":        true,
					"max_amount_limit":   10000.00,
				},
			})
			Description("Dynamic conditions evaluated at runtime")
		})
		
		Field(14, "obligations", MapOf(String, Any), "Required actions when permission is used", func() {
			Example(map[string]any{
				"audit_log":    true,
				"notify_admin": true,
				"require_reason": true,
				"approval_workflow": "high_value_transactions",
			})
			Description("Mandatory actions triggered by permission usage")
		})
		
		Field(15, "is_system_permission", Boolean, "System-defined permission flag", func() {
			Default(false)
			Example(false)
			Description("Cannot be modified or deleted if true")
		})
		
		Field(16, "is_active", Boolean, "Permission availability status", func() {
			Default(true)
			Example(true)
			Description("Controls whether permission can be assigned")
		})
		
		Field(17, "is_delegable", Boolean, "Can be delegated to other users", func() {
			Default(false)
			Example(true)
			Description("Allows users to grant this permission to others")
		})
		
		Field(18, "category", String, "Permission functional category", func() {
			Enum("SYSTEM", "BUSINESS", "SECURITY", "ADMINISTRATIVE", "REPORTING", "INTEGRATION")
			Default("BUSINESS")
			Example("BUSINESS")
			Description("Functional grouping for organization")
		})
		
		Field(19, "tags", ArrayOf(String), "Classification tags", func() {
			Example([]string{"financial", "sensitive", "audit_required"})
			Description("Searchable tags for permission categorization")
		})
		
		Field(20, "dependencies", ArrayOf(String), "Required prerequisite permissions", func() {
			Elem(func() {
				Format(FormatUUID)
			})
			Example([]string{
				"perm-base-read-access",
				"perm-entity-membership",
			})
			Description("Other permissions that must be present")
		})
		
		Field(21, "conflicts_with", ArrayOf(String), "Mutually exclusive permissions", func() {
			Elem(func() {
				Format(FormatUUID)
			})
			Example([]string{
				"perm-accounts-readonly-mode",
			})
			Description("Permissions that cannot coexist with this one")
		})
		
		Field(22, "resource_attributes", MapOf(String, Any), "Resource-specific attributes", func() {
			Example(map[string]any{
				"data_classification": "confidential",
				"compliance_level":    "SOX",
				"encryption_required": true,
			})
			Description("Metadata about the protected resource")
		})
		
		Field(23, "usage_count", UInt, "Number of times permission has been used", func() {
			Example(1247)
			Description("Statistical usage tracking")
		})
		
		Field(24, "last_used_at", String, "Most recent usage timestamp", func() {
			Format(FormatDateTime)
			Example("2023-12-07T14:30:00Z")
			Description("Latest activity using this permission")
		})
		
		// Audit fields from common.go
		AuditFields()
		
		Required("id", "tenant_id", "resource_id", "action_id", "name", "display_name", 
			"resource_type", "action_type", "effect", "scope", "is_active", "created_at")
	})
	
	View("default", func() {
		Description("Standard permission view for listings and general display")
		Attribute("id")
		Attribute("name")
		Attribute("display_name")
		Attribute("resource_type")
		Attribute("action_type")
		Attribute("effect")
		Attribute("scope")
		Attribute("is_active")
		Attribute("created_at")
	})
	
	View("detailed", func() {
		Description("Complete permission view with all configuration and metadata")
		Attribute("id")
		Attribute("tenant_id")
		Attribute("resource_id")
		Attribute("action_id")
		Attribute("name")
		Attribute("display_name")
		Attribute("description")
		Attribute("resource_type")
		Attribute("action_type")
		Attribute("effect")
		Attribute("scope")
		Attribute("priority")
		Attribute("conditions")
		Attribute("obligations")
		Attribute("is_system_permission")
		Attribute("is_active")
		Attribute("is_delegable")
		Attribute("category")
		Attribute("tags")
		Attribute("dependencies")
		Attribute("conflicts_with")
		Attribute("resource_attributes")
		Attribute("usage_count")
		Attribute("last_used_at")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})
	
	View("summary", func() {
		Description("Minimal view for references and permission checks")
		Attribute("id")
		Attribute("name")
		Attribute("resource_type")
		Attribute("action_type")
		Attribute("effect")
		Attribute("is_active")
	})
	
	View("evaluation", func() {
		Description("Optimized view for permission evaluation engines")
		Attribute("id")
		Attribute("name")
		Attribute("effect")
		Attribute("priority")
		Attribute("conditions")
		Attribute("obligations")
		Attribute("scope")
	})
})

// Resource represents a protected resource in the permission system.
var Resource = Type("Resource", func() {
	Description("Protected resource definition for permission-based access control")
	
	Field(1, "id", String, "Unique resource identifier", func() {
		Format(FormatUUID)
		Example("resource-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary key for resource references")
	})
	
	Field(2, "tenant_id", String, "Associated tenant identifier", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Description("Tenant isolation boundary")
	})
	
	Field(3, "name", String, "Unique resource name", func() {
		Pattern("^[a-z][a-z0-9_]{2,49}$")
		MinLength(3)
		MaxLength(50)
		Example("financial_accounts")
		Description("Machine-readable resource identifier")
	})
	
	Field(4, "display_name", String, "Human-readable resource title", func() {
		MinLength(3)
		MaxLength(100)
		Example("Financial Accounts")
		Description("User-friendly name for UI display")
	})
	
	Field(5, "description", String, "Resource description", func() {
		MaxLength(500)
		Example("Chart of accounts and financial account management")
		Description("Detailed description of protected resource")
	})
	
	Field(6, "resource_type", String, "Resource category", func() {
		Enum("DATA", "FUNCTION", "SERVICE", "UI_COMPONENT", "API_ENDPOINT", "REPORT")
		Example("DATA")
		Description("Type of resource being protected")
	})
	
	Field(7, "attributes", MapOf(String, Any), "Resource metadata attributes", func() {
		Example(map[string]any{
			"data_classification": "confidential",
			"compliance_tags":     []string{"SOX", "PCI"},
			"owner_department":    "finance",
		})
		Description("Metadata for ABAC evaluation")
	})
	
	Field(8, "is_active", Boolean, "Resource active status", func() {
		Default(true)
		Example(true)
		Description("Whether resource is available for protection")
	})
	
	// Audit fields
	AuditFields()
	
	Required("id", "tenant_id", "name", "display_name", "resource_type", "is_active", "created_at")
})

// Action represents an operation that can be performed on resources.
var Action = Type("Action", func() {
	Description("Operation definition for resource-based permission control")
	
	Field(1, "id", String, "Unique action identifier", func() {
		Format(FormatUUID)
		Example("action-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary key for action references")
	})
	
	Field(2, "tenant_id", String, "Associated tenant identifier", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Description("Tenant isolation boundary")
	})
	
	Field(3, "name", String, "Unique action name", func() {
		Pattern("^[a-z][a-z0-9_]{2,49}$")
		MinLength(3)
		MaxLength(50)
		Example("read")
		Description("Machine-readable action identifier")
	})
	
	Field(4, "display_name", String, "Human-readable action title", func() {
		MinLength(3)
		MaxLength(100)
		Example("Read Access")
		Description("User-friendly name for UI display")
	})
	
	Field(5, "description", String, "Action description", func() {
		MaxLength(500)
		Example("View and read resource data without modification")
		Description("Detailed description of the operation")
	})
	
	Field(6, "action_type", String, "Action category", func() {
		Enum("CREATE", "READ", "UPDATE", "DELETE", "EXECUTE", "APPROVE", "EXPORT", 
			"IMPORT", "ARCHIVE", "RESTORE", "ASSIGN", "UNASSIGN")
		Example("READ")
		Description("Standard operation classification")
	})
	
	Field(7, "risk_level", String, "Security risk assessment", func() {
		Enum("LOW", "MEDIUM", "HIGH", "CRITICAL")
		Default("MEDIUM")
		Example("LOW")
		Description("Risk level for audit and approval requirements")
	})
	
	Field(8, "requires_approval", Boolean, "Action requires approval workflow", func() {
		Default(false)
		Example(true)
		Description("High-risk actions may require pre-approval")
	})
	
	Field(9, "is_auditable", Boolean, "Action should be audited", func() {
		Default(true)
		Example(true)
		Description("Whether to log usage in audit trail")
	})
	
	Field(10, "attributes", MapOf(String, Any), "Action metadata attributes", func() {
		Example(map[string]any{
			"reversible":        false,
			"data_modification": true,
			"external_effect":   false,
		})
		Description("Metadata for policy evaluation")
	})
	
	Field(11, "is_active", Boolean, "Action active status", func() {
		Default(true)
		Example(true)
		Description("Whether action is available for use")
	})
	
	// Audit fields
	AuditFields()
	
	Required("id", "tenant_id", "name", "display_name", "action_type", "risk_level", "is_active", "created_at")
})

// PermissionGrant represents an explicit permission grant to a user or role.
var PermissionGrant = Type("PermissionGrant", func() {
	Description("Explicit permission grant with scope and conditions")
	
	Field(1, "id", String, "Unique grant identifier", func() {
		Format(FormatUUID)
		Example("grant-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary key for grant record")
	})
	
	Field(2, "permission_id", String, "Associated permission identifier", func() {
		Format(FormatUUID)
		Example("perm-456e7890-e89b-12d3-a456-426614174000")
		Description("Permission being granted")
	})
	
	Field(3, "grantee_type", String, "Type of entity receiving permission", func() {
		Enum("USER", "ROLE", "GROUP", "API_KEY")
		Example("USER")
		Description("Category of permission recipient")
	})
	
	Field(4, "grantee_id", String, "Identifier of permission recipient", func() {
		Format(FormatUUID)
		Example("user-789abc12-def3-4567-890a-bcdef1234567")
		Description("User, role, or group receiving permission")
	})
	
	Field(5, "grantor_id", String, "User who granted the permission", func() {
		Format(FormatUUID)
		Example("admin-123e4567-e89b-12d3-a456-426614174000")
		Description("Administrator who authorized the grant")
	})
	
	Field(6, "scope_entities", ArrayOf(String), "Entity scope for permission", func() {
		Elem(func() {
			Format(FormatUUID)
		})
		Example([]string{
			"entity-123e4567-e89b-12d3-a456-426614174000",
			"entity-456e7890-e89b-12d3-a456-426614174000",
		})
		Description("Entities where permission is effective")
	})
	
	Field(7, "conditions", MapOf(String, Any), "Grant-specific conditions", func() {
		Example(map[string]any{
			"time_restrictions": map[string]any{
				"business_hours_only": true,
				"timezone":           "America/New_York",
			},
			"ip_restrictions": []string{"192.168.1.0/24"},
		})
		Description("Additional constraints on permission usage")
	})
	
	Field(8, "effective_from", String, "Grant effective start time", func() {
		Format(FormatDateTime)
		Example("2023-12-07T00:00:00Z")
		Description("When permission grant becomes active")
	})
	
	Field(9, "expires_at", String, "Grant expiration time", func() {
		Format(FormatDateTime)
		Example("2024-06-30T23:59:59Z")
		Description("When permission grant automatically expires")
	})
	
	Field(10, "is_active", Boolean, "Grant active status", func() {
		Default(true)
		Example(true)
		Description("Whether grant is currently effective")
	})
	
	Field(11, "reason", String, "Justification for permission grant", func() {
		MaxLength(500)
		Example("Temporary elevated access for end-of-month financial closing")
		Description("Business justification for granting permission")
	})
	
	Field(12, "usage_count", UInt, "Number of times permission was used", func() {
		Example(47)
		Description("Statistical usage tracking")
	})
	
	Field(13, "last_used_at", String, "Most recent usage timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T14:30:00Z")
		Description("Latest activity using this grant")
	})
	
	// Audit fields
	AuditFields()
	
	Required("id", "permission_id", "grantee_type", "grantee_id", "grantor_id", 
		"effective_from", "is_active", "created_at")
})