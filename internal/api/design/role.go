package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// ROLE-BASED ACCESS CONTROL TYPES
// ============================================================================

// Role represents a collection of permissions that can be assigned to users.
// Roles support hierarchical inheritance and can be scoped to specific
// tenants and entities for fine-grained access control.
var Role = ResultType("application/vnd.erp.role", func() {
	Description("Role definition containing permissions and access rights for user authorization within the RBAC system")
	
	Attributes(func() {
		Field(1, "id", String, "Unique role identifier", func() {
			Format(FormatUUID)
			Example("role-123e4567-e89b-12d3-a456-426614174000")
			Description("Primary key for role references")
		})
		
		Field(2, "tenant_id", String, "Associated tenant identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
			Description("Tenant isolation boundary for role scope")
		})
		
		Field(3, "entity_id", String, "Entity scope for role application", func() {
			Format(FormatUUID)
			Example("987fcdeb-51d2-43b8-a456-426614174000")
			Description("Optional entity restriction for role permissions")
		})
		
		Field(4, "name", String, "Unique role name within tenant", func() {
			Pattern("^[a-z][a-z0-9_]{2,49}$")
			MinLength(3)
			MaxLength(50)
			Example("department_manager")
			Description("Machine-readable role identifier")
		})
		
		Field(5, "display_name", String, "Human-readable role title", func() {
			MinLength(3)
			MaxLength(100)
			Example("Department Manager")
			Description("User-friendly name for UI display")
		})
		
		Field(6, "description", String, "Detailed role description", func() {
			MaxLength(500)
			Example("Manages departmental operations, approves budgets, and oversees team members")
			Description("Comprehensive explanation of role responsibilities")
		})
		
		Field(7, "role_type", String, "Classification of role scope and purpose", func() {
			Enum("SYSTEM", "TENANT", "ENTITY", "CUSTOM", "FUNCTIONAL", "ADMINISTRATIVE", "OPERATIONAL")
			Default("CUSTOM")
			Example("FUNCTIONAL")
			Description("Determines role behavior and inheritance patterns")
		})
		
		Field(8, "parent_role_id", String, "Parent role for inheritance", func() {
			Format(FormatUUID)
			Example("parent-role-456e7890-e89b-12d3-a456-426614174000")
			Description("Inherits permissions from parent role")
		})
		
		Field(9, "level", UInt, "Hierarchical level in role tree", func() {
			Minimum(0)
			Maximum(10)
			Default(0)
			Example(2)
			Description("Depth in role hierarchy (0 = root level)")
		})
		
		Field(10, "priority", UInt, "Role priority for conflict resolution", func() {
			Minimum(1)
			Maximum(1000)
			Default(100)
			Example(50)
			Description("Higher numbers take precedence in permission conflicts")
		})
		
		Field(11, "is_system_role", Boolean, "System-defined role flag", func() {
			Default(false)
			Example(false)
			Description("Cannot be modified or deleted if true")
		})
		
		Field(12, "is_active", Boolean, "Role availability status", func() {
			Default(true)
			Example(true)
			Description("Controls whether role can be assigned to users")
		})
		
		Field(13, "is_assignable", Boolean, "Can be assigned to users", func() {
			Default(true)
			Example(true)
			Description("Some roles may exist only for inheritance")
		})
		
		Field(14, "auto_assign_conditions", MapOf(String, Any), "Automatic role assignment rules", func() {
			Example(map[string]any{
				"user_type":         "EMPLOYEE",
				"department":        "ENGINEERING",
				"employment_status": "ACTIVE",
			})
			Description("Conditions for automatic role assignment")
		})
		
		Field(15, "permission_count", UInt, "Number of associated permissions", func() {
			Example(15)
			Description("Total permissions directly assigned to this role")
		})
		
		Field(16, "inherited_permission_count", UInt, "Number of inherited permissions", func() {
			Example(8)
			Description("Permissions inherited from parent roles")
		})
		
		Field(17, "user_count", UInt, "Number of users with this role", func() {
			Example(25)
			Description("Active users currently assigned this role")
		})
		
		Field(18, "expires_at", String, "Role expiration timestamp", func() {
			Format(FormatDateTime)
			Example("2024-12-31T23:59:59Z")
			Description("Optional expiration date for temporary roles")
		})
		
		Field(19, "approval_required", Boolean, "Requires approval for assignment", func() {
			Default(false)
			Example(true)
			Description("High-privilege roles may require manager approval")
		})
		
		Field(20, "approval_workflow_id", String, "Associated approval workflow", func() {
			Format(FormatUUID)
			Example("workflow-789abc12-def3-4567-890a-bcdef1234567")
			Description("Workflow for role assignment approval")
		})
		
		Field(21, "tags", ArrayOf(String), "Classification tags", func() {
			Example([]string{"financial", "management", "approval_authority"})
			Description("Searchable tags for role categorization")
		})
		
		Field(22, "constraints", MapOf(String, Any), "Role assignment constraints", func() {
			Example(map[string]any{
				"max_users":           50,
				"require_mfa":         true,
				"allowed_entities":    []string{"entity-1", "entity-2"},
				"time_restrictions":   map[string]any{
					"business_hours_only": true,
					"timezone":           "America/New_York",
				},
			})
			Description("Limitations and requirements for role usage")
		})
		
		Field(23, "settings", MapOf(String, Any), "Role-specific configuration", func() {
			Example(map[string]any{
				"session_timeout":      480,
				"ip_restrictions":      []string{"192.168.1.0/24"},
				"require_reason":       true,
				"audit_all_actions":   true,
			})
			Description("Custom settings that apply when role is active")
		})
		
		// Audit fields from common.go
		AuditFields()
		
		Required("id", "tenant_id", "name", "display_name", "role_type", "is_active", "created_at")
	})
	
	View("default", func() {
		Description("Standard role view for listings and general display")
		Attribute("id")
		Attribute("name")
		Attribute("display_name")
		Attribute("role_type")
		Attribute("is_active")
		Attribute("permission_count")
		Attribute("user_count")
		Attribute("created_at")
	})
	
	View("detailed", func() {
		Description("Complete role view with all configuration and metadata")
		Attribute("id")
		Attribute("tenant_id")
		Attribute("entity_id")
		Attribute("name")
		Attribute("display_name")
		Attribute("description")
		Attribute("role_type")
		Attribute("parent_role_id")
		Attribute("level")
		Attribute("priority")
		Attribute("is_system_role")
		Attribute("is_active")
		Attribute("is_assignable")
		Attribute("auto_assign_conditions")
		Attribute("permission_count")
		Attribute("inherited_permission_count")
		Attribute("user_count")
		Attribute("expires_at")
		Attribute("approval_required")
		Attribute("approval_workflow_id")
		Attribute("tags")
		Attribute("constraints")
		Attribute("settings")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})
	
	View("hierarchy", func() {
		Description("Hierarchical view for role inheritance trees")
		Attribute("id")
		Attribute("name")
		Attribute("display_name")
		Attribute("parent_role_id")
		Attribute("level")
		Attribute("priority")
		Attribute("is_active")
	})
	
	View("assignment", func() {
		Description("Role view optimized for user assignment interfaces")
		Attribute("id")
		Attribute("name")
		Attribute("display_name")
		Attribute("description")
		Attribute("role_type")
		Attribute("is_assignable")
		Attribute("approval_required")
		Attribute("constraints")
	})
})

// RolePermissionMapping represents the association between roles and permissions.
// This type supports both direct and inherited permission relationships.
var RolePermissionMapping = Type("RolePermissionMapping", func() {
	Description("Association between roles and permissions supporting inheritance and grant types")
	
	Field(1, "id", String, "Unique mapping identifier", func() {
		Format(FormatUUID)
		Example("mapping-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary key for mapping record")
	})
	
	Field(2, "role_id", String, "Associated role identifier", func() {
		Format(FormatUUID)
		Example("role-123e4567-e89b-12d3-a456-426614174000")
		Description("Role receiving the permission")
	})
	
	Field(3, "permission_id", String, "Associated permission identifier", func() {
		Format(FormatUUID)
		Example("perm-456e7890-e89b-12d3-a456-426614174000")
		Description("Permission being granted")
	})
	
	Field(4, "grant_type", String, "Type of permission grant", func() {
		Enum("DIRECT", "INHERITED", "CONDITIONAL", "TEMPORARY")
		Default("DIRECT")
		Example("DIRECT")
		Description("How the permission was assigned to the role")
	})
	
	Field(5, "inherited_from_role_id", String, "Source role for inherited permissions", func() {
		Format(FormatUUID)
		Example("parent-role-456e7890-e89b-12d3-a456-426614174000")
		Description("Parent role that provided this permission")
	})
	
	Field(6, "conditions", MapOf(String, Any), "Conditional permission constraints", func() {
		Example(map[string]any{
			"time_restriction": map[string]any{
				"start_time": "09:00",
				"end_time":   "17:00",
				"timezone":   "America/New_York",
			},
			"entity_scope": []string{"entity-1", "entity-2"},
		})
		Description("Conditions that must be met for permission to be active")
	})
	
	Field(7, "expires_at", String, "Permission expiration timestamp", func() {
		Format(FormatDateTime)
		Example("2024-06-30T23:59:59Z")
		Description("When this permission grant expires")
	})
	
	Field(8, "is_active", Boolean, "Mapping active status", func() {
		Default(true)
		Example(true)
		Description("Whether this permission is currently effective")
	})
	
	// Audit fields
	AuditFields()
	
	Required("id", "role_id", "permission_id", "grant_type", "is_active", "created_at")
})

// RoleHierarchy represents hierarchical relationships between roles for inheritance.
var RoleHierarchy = Type("RoleHierarchy", func() {
	Description("Hierarchical relationship mapping for role inheritance and permission propagation")
	
	Field(1, "parent_role_id", String, "Parent role identifier", func() {
		Format(FormatUUID)
		Example("parent-role-123e4567-e89b-12d3-a456-426614174000")
		Description("Role providing inherited permissions")
	})
	
	Field(2, "child_role_id", String, "Child role identifier", func() {
		Format(FormatUUID)
		Example("child-role-456e7890-e89b-12d3-a456-426614174000")
		Description("Role receiving inherited permissions")
	})
	
	Field(3, "depth", UInt, "Hierarchical depth", func() {
		Minimum(1)
		Maximum(10)
		Example(1)
		Description("Number of levels between parent and child")
	})
	
	Field(4, "inheritance_type", String, "Type of inheritance relationship", func() {
		Enum("FULL", "PARTIAL", "CONDITIONAL", "ADDITIVE")
		Default("FULL")
		Example("FULL")
		Description("How permissions are inherited from parent")
	})
	
	Field(5, "inheritance_rules", MapOf(String, Any), "Rules governing inheritance", func() {
		Example(map[string]any{
			"exclude_permissions": []string{"admin.users.delete"},
			"include_only":       []string{"finance.*", "reports.view"},
			"modify_scope":       true,
		})
		Description("Customization rules for permission inheritance")
	})
	
	Field(6, "is_active", Boolean, "Hierarchy relationship status", func() {
		Default(true)
		Example(true)
		Description("Whether inheritance is currently active")
	})
	
	// Audit fields
	AuditFields()
	
	Required("parent_role_id", "child_role_id", "depth", "inheritance_type", "is_active", "created_at")
})

// UserRoleAssignment represents the assignment of roles to users with context.
var UserRoleAssignment = Type("UserRoleAssignment", func() {
	Description("Assignment of roles to users with scope, conditions, and audit trail")
	
	Field(1, "id", String, "Unique assignment identifier", func() {
		Format(FormatUUID)
		Example("assignment-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary key for assignment record")
	})
	
	Field(2, "user_id", String, "Assigned user identifier", func() {
		Format(FormatUUID)
		Example("user-456e7890-e89b-12d3-a456-426614174000")
		Description("User receiving the role")
	})
	
	Field(3, "role_id", String, "Assigned role identifier", func() {
		Format(FormatUUID)
		Example("role-123e4567-e89b-12d3-a456-426614174000")
		Description("Role being assigned")
	})
	
	Field(4, "entity_scope", ArrayOf(String), "Entity scope for role assignment", func() {
		Elem(func() {
			Format(FormatUUID)
		})
		Example([]string{
			"entity-123e4567-e89b-12d3-a456-426614174000",
			"entity-456e7890-e89b-12d3-a456-426614174000",
		})
		Description("Entities where this role assignment is effective")
	})
	
	Field(5, "assignment_type", String, "Type of role assignment", func() {
		Enum("PERMANENT", "TEMPORARY", "CONDITIONAL", "EMERGENCY", "INHERITED")
		Default("PERMANENT")
		Example("TEMPORARY")
		Description("Nature and duration of the assignment")
	})
	
	Field(6, "assigned_by", String, "User who made the assignment", func() {
		Format(FormatUUID)
		Example("admin-789abc12-def3-4567-890a-bcdef1234567")
		Description("Administrator or manager who granted the role")
	})
	
	Field(7, "assignment_reason", String, "Justification for role assignment", func() {
		MaxLength(500)
		Example("Temporary assignment during manager's vacation leave")
		Description("Business justification for granting role")
	})
	
	Field(8, "approval_status", String, "Assignment approval status", func() {
		Enum("PENDING", "APPROVED", "REJECTED", "AUTO_APPROVED", "EXPIRED")
		Default("PENDING")
		Example("APPROVED")
		Description("Workflow approval state")
	})
	
	Field(9, "approved_by", String, "Approver user identifier", func() {
		Format(FormatUUID)
		Example("approver-abc123de-f456-7890-abc1-23def4567890")
		Description("User who approved the assignment")
	})
	
	Field(10, "approved_at", String, "Approval timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
		Description("When the assignment was approved")
	})
	
	Field(11, "effective_from", String, "Assignment effective start time", func() {
		Format(FormatDateTime)
		Example("2023-12-07T00:00:00Z")
		Description("When the role assignment becomes active")
	})
	
	Field(12, "expires_at", String, "Assignment expiration time", func() {
		Format(FormatDateTime)
		Example("2024-01-07T23:59:59Z")
		Description("When the role assignment automatically expires")
	})
	
	Field(13, "is_active", Boolean, "Assignment active status", func() {
		Default(true)
		Example(true)
		Description("Whether the assignment is currently effective")
	})
	
	Field(14, "conditions", MapOf(String, Any), "Assignment conditions and constraints", func() {
		Example(map[string]any{
			"ip_restrictions":    []string{"192.168.1.0/24"},
			"time_restrictions": map[string]any{
				"business_hours_only": true,
				"timezone":           "America/New_York",
			},
			"require_approval_for": []string{"budget_changes", "user_management"},
		})
		Description("Conditions that must be met for role to be active")
	})
	
	Field(15, "last_used_at", String, "Last time role was used", func() {
		Format(FormatDateTime)
		Example("2023-12-07T14:30:00Z")
		Description("Most recent activity using this role")
	})
	
	// Audit fields
	AuditFields()
	
	Required("id", "user_id", "role_id", "assignment_type", "assigned_by", "approval_status", 
		"effective_from", "is_active", "created_at")
})