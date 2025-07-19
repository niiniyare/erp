package user

import (
	"github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// =============================================================================
// ENUM DEFINITIONS
// =============================================================================

var PersonType = Type("PersonType", String, func() {
	Description("Type of person entity")
	Enum("INDIVIDUAL", "EMPLOYEE", "CONTACT", "CUSTOMER", "VENDOR", "CONTRACTOR")
	Example("EMPLOYEE")
})

var EmploymentStatus = Type("EmploymentStatus", String, func() {
	Description("Employment status of an employee")
	Enum("ACTIVE", "INACTIVE", "TERMINATED", "ON_LEAVE", "SUSPENDED")
	Example("ACTIVE")
})

var UserType = Type("UserType", String, func() {
	Description("Type of system user account")
	Enum("INTERNAL", "CUSTOMER", "VENDOR", "PARTNER", "API", "SERVICE", "ADMIN")
	Example("INTERNAL")
})

var AccountStatus = Type("AccountStatus", String, func() {
	Description("Status of user account")
	Enum("ACTIVE", "INACTIVE", "LOCKED", "SUSPENDED", "PENDING_VERIFICATION", "EXPIRED")
	Example("ACTIVE")
})

var RoleType = Type("RoleType", String, func() {
	Description("Type of role in the system")
	Enum("SYSTEM", "TENANT", "ENTITY", "CUSTOM", "FUNCTIONAL")
	Example("CUSTOM")
})

var ResourceType = Type("ResourceType", String, func() {
	Description("Type of system resource")
	Enum("API", "UI", "DATA", "FILE", "REPORT", "WORKFLOW", "FUNCTION")
	Example("API")
})

var ActionType = Type("ActionType", String, func() {
	Description("Type of action that can be performed")
	Enum("CREATE", "READ", "UPDATE", "DELETE", "EXECUTE", "APPROVE", "REJECT", "EXPORT", "IMPORT")
	Example("READ")
})

var AssignmentType = Type("AssignmentType", String, func() {
	Description("Type of role assignment")
	Enum("DIRECT", "INHERITED", "DELEGATED", "TEMPORARY")
	Example("DIRECT")
})

// =============================================================================
// RESULT TYPES
// =============================================================================

// UserResult describes the user response
var UserResult = ResultType("application/vnd.user", func() {
	Description("User information")
	Attributes(func() {
		Attribute("id", String, "Unique user identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
		})
		Attribute("email", String, "User email address", func() {
			Format(FormatEmail)
			Example("john.doe@example.com")
		})
		Attribute("username", String, "Username for login", func() {
			MinLength(3)
			MaxLength(50)
			Pattern("^[a-zA-Z0-9_.-]+$")
			Example("john.doe")
		})
		Attribute("first_name", String, "First name", func() {
			MinLength(1)
			MaxLength(100)
			Example("John")
		})
		Attribute("last_name", String, "Last name", func() {
			MinLength(1)
			MaxLength(100)
			Example("Doe")
		})
		Attribute("display_name", String, "Display name", func() {
			MaxLength(200)
			Example("John Doe")
		})
		Attribute("status", AccountStatus, "Account status")
		Attribute("user_type", UserType, "Type of user account")
		Attribute("person_type", PersonType, "Type of person")
		Attribute("employment_status", EmploymentStatus, "Employment status")
		Attribute("phone", String, "Phone number", func() {
			Pattern("^\\+?[1-9]\\d{1,14}$")
			Example("+1-555-123-4567")
		})
		Attribute("profile", UserProfile, "User profile information")
		Attribute("preferences", UserPreferences, "User preferences")
		Attribute("last_login_at", String, "Last login timestamp", func() {
			Format(FormatDateTime)
			Example("2023-12-07T10:30:00Z")
		})
		Attribute("password_changed_at", String, "Password last changed timestamp", func() {
			Format(FormatDateTime)
			Example("2023-12-01T10:30:00Z")
		})
		types.AuditFields()
	})
	Required("id", "email", "first_name", "last_name", "status", "user_type", "created_at", "updated_at")

	View("default", func() {
		Attribute("id")
		Attribute("email")
		Attribute("username")
		Attribute("first_name")
		Attribute("last_name")
		Attribute("display_name")
		Attribute("status")
		Attribute("user_type")
		Attribute("person_type")
		Attribute("employment_status")
		Attribute("phone")
		Attribute("profile")
		Attribute("preferences")
		Attribute("last_login_at")
		Attribute("created_at")
		Attribute("updated_at")
	})

	View("minimal", func() {
		Attribute("id")
		Attribute("email")
		Attribute("first_name")
		Attribute("last_name")
		Attribute("status")
	})

	View("public", func() {
		Attribute("id")
		Attribute("first_name")
		Attribute("last_name")
		Attribute("display_name")
	})
})

// UserPermissionsResult describes user permissions and roles
var UserPermissionsResult = ResultType("application/vnd.user.permissions", func() {
	Description("User permissions and role information")
	Attributes(func() {
		Attribute("user_id", String, "User ID", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
		})
		Attribute("roles", ArrayOf(RoleInfo), "Assigned roles")
		Attribute("permissions", ArrayOf(PermissionInfo), "Effective permissions")
		Attribute("computed_at", String, "When permissions were computed", func() {
			Format(FormatDateTime)
			Example("2023-12-07T10:30:00Z")
		})
	})
	Required("user_id", "roles", "permissions", "computed_at")
})

// =============================================================================
// PAYLOAD TYPES
// =============================================================================

// CreateUserPayload describes the payload for creating a user
var CreateUserPayload = Type("CreateUserPayload", func() {
	Description("Payload for creating a new user")
	Attribute("email", String, "User email address", func() {
		Format(FormatEmail)
		Example("john.doe@example.com")
	})
	Attribute("username", String, "Username for login", func() {
		MinLength(3)
		MaxLength(50)
		Pattern("^[a-zA-Z0-9_.-]+$")
		Example("john.doe")
	})
	Attribute("first_name", String, "First name", func() {
		MinLength(1)
		MaxLength(100)
		Example("John")
	})
	Attribute("last_name", String, "Last name", func() {
		MinLength(1)
		MaxLength(100)
		Example("Doe")
	})
	Attribute("password", String, "Initial password", func() {
		MinLength(8)
		Example("SecurePassword123!")
	})
	Attribute("user_type", UserType, "Type of user account")
	Attribute("person_type", PersonType, "Type of person")
	Attribute("employment_status", EmploymentStatus, "Employment status")
	Attribute("phone", String, "Phone number", func() {
		Pattern("^\\+?[1-9]\\d{1,14}$")
		Example("+1-555-123-4567")
	})
	Attribute("profile", UserProfile, "User profile information")
	Attribute("preferences", UserPreferences, "User preferences")
	Attribute("send_invitation", Boolean, "Whether to send invitation email", func() {
		Default(true)
		Example(true)
	})
	Required("email", "first_name", "last_name", "user_type")
})

// UpdateUserPayload describes the payload for updating a user
var UpdateUserPayload = Type("UpdateUserPayload", func() {
	Description("Payload for updating an existing user")
	Attribute("id", String, "User ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("email", String, "User email address", func() {
		Format(FormatEmail)
		Example("john.doe@example.com")
	})
	Attribute("username", String, "Username for login", func() {
		MinLength(3)
		MaxLength(50)
		Pattern("^[a-zA-Z0-9_.-]+$")
		Example("john.doe")
	})
	Attribute("first_name", String, "First name", func() {
		MinLength(1)
		MaxLength(100)
		Example("John")
	})
	Attribute("last_name", String, "Last name", func() {
		MinLength(1)
		MaxLength(100)
		Example("Doe")
	})
	Attribute("status", AccountStatus, "Account status")
	Attribute("user_type", UserType, "Type of user account")
	Attribute("person_type", PersonType, "Type of person")
	Attribute("employment_status", EmploymentStatus, "Employment status")
	Attribute("phone", String, "Phone number", func() {
		Pattern("^\\+?[1-9]\\d{1,14}$")
		Example("+1-555-123-4567")
	})
	Attribute("profile", UserProfile, "User profile information")
	Attribute("preferences", UserPreferences, "User preferences")
	Required("id")
})

// AssignRolePayload describes the payload for assigning a role to a user
var AssignRolePayload = Type("AssignRolePayload", func() {
	Description("Payload for assigning a role to a user")
	Attribute("user_id", String, "User ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("role_id", String, "Role ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440001")
	})
	Attribute("assignment_type", AssignmentType, "Type of assignment")
	Attribute("expires_at", String, "Assignment expiration time", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	Attribute("reason", String, "Reason for assignment", func() {
		MaxLength(500)
		Example("Temporary assignment for project Alpha")
	})
	Required("user_id", "role_id", "assignment_type")
})

// =============================================================================
// NESTED TYPES
// =============================================================================

// UserProfile describes detailed user profile information
var UserProfile = Type("UserProfile", func() {
	Description("User profile information")
	Attribute("title", String, "Job title", func() {
		MaxLength(100)
		Example("Senior Software Engineer")
	})
	Attribute("department", String, "Department", func() {
		MaxLength(100)
		Example("Engineering")
	})
	Attribute("manager_id", String, "Manager user ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440002")
	})
	Attribute("hire_date", String, "Hire date", func() {
		Format(FormatDate)
		Example("2023-01-15")
	})
	Attribute("timezone", String, "User timezone", func() {
		Example("America/New_York")
		Default("UTC")
	})
	Attribute("avatar_url", String, "Avatar image URL", func() {
		Format(FormatURI)
		Example("https://example.com/avatars/user123.jpg")
	})
	Attribute("bio", String, "User bio/description", func() {
		MaxLength(500)
		Example("Experienced software engineer with expertise in Go and microservices")
	})
})

// UserPreferences describes user preferences and settings
var UserPreferences = Type("UserPreferences", func() {
	Description("User preferences and settings")
	Attribute("language", String, "Preferred language", func() {
		Pattern("^[a-z]{2}$")
		Example("en")
		Default("en")
	})
	Attribute("theme", String, "UI theme preference", func() {
		Enum("light", "dark", "auto")
		Example("dark")
		Default("light")
	})
	Attribute("date_format", String, "Preferred date format", func() {
		Enum("MM/DD/YYYY", "DD/MM/YYYY", "YYYY-MM-DD")
		Example("MM/DD/YYYY")
		Default("MM/DD/YYYY")
	})
	Attribute("time_format", String, "Preferred time format", func() {
		Enum("12h", "24h")
		Example("24h")
		Default("12h")
	})
	Attribute("notifications", NotificationPreferences, "Notification preferences")
})

// NotificationPreferences describes notification settings
var NotificationPreferences = Type("NotificationPreferences", func() {
	Description("User notification preferences")
	Attribute("email_notifications", Boolean, "Enable email notifications", func() {
		Default(true)
		Example(true)
	})
	Attribute("push_notifications", Boolean, "Enable push notifications", func() {
		Default(true)
		Example(true)
	})
	Attribute("sms_notifications", Boolean, "Enable SMS notifications", func() {
		Default(false)
		Example(false)
	})
	Attribute("marketing_emails", Boolean, "Enable marketing emails", func() {
		Default(false)
		Example(false)
	})
})

// RoleInfo describes role information for permissions result
var RoleInfo = Type("RoleInfo", func() {
	Description("Role information")
	Attribute("id", String, "Role ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("name", String, "Role name", func() {
		Example("Administrator")
	})
	Attribute("description", String, "Role description", func() {
		Example("Full system administrator access")
	})
	Attribute("role_type", RoleType, "Type of role")
	Attribute("assignment_type", AssignmentType, "How role was assigned")
	Attribute("assigned_at", String, "When role was assigned", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	Attribute("expires_at", String, "When assignment expires", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	Required("id", "name", "role_type", "assignment_type", "assigned_at")
})

// PermissionInfo describes permission information
var PermissionInfo = Type("PermissionInfo", func() {
	Description("Permission information")
	Attribute("resource", String, "Resource name", func() {
		Example("users")
	})
	Attribute("action", ActionType, "Action type")
	Attribute("effect", String, "Permission effect", func() {
		Enum("ALLOW", "DENY")
		Example("ALLOW")
	})
	Attribute("conditions", MapOf(String, Any), "Permission conditions")
	Attribute("source_role", String, "Role that granted this permission", func() {
		Example("Administrator")
	})
	Required("resource", "action", "effect")
})
