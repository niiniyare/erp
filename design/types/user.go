package types

import (
	. "goa.design/goa/v3/dsl"
)

// Enum definitions for user management
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

var ActionCategory = Type("ActionCategory", String, func() {
	Description("Category of action based on sensitivity")
	Enum("LOW", "MEDIUM", "HIGH", "CRITICAL")
	Example("MEDIUM")
})

// UserResult describes the user response
var UserResult = ResultType("application/vnd.user", func() {
	Description("User information")
	Attributes(func() {
		Attribute("id", String, "Unique user identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
		})
		Attribute("username", String, "Username", func() {
			MinLength(3)
			MaxLength(50)
			Example("john.doe")
		})
		Attribute("email", String, "Email address", func() {
			Format(FormatEmail)
			Example("john.doe@acme.com")
		})
		Attribute("first_name", String, "First name", func() {
			MinLength(1)
			MaxLength(50)
			Example("John")
		})
		Attribute("last_name", String, "Last name", func() {
			MinLength(1)
			MaxLength(50)
			Example("Doe")
		})
		Attribute("user_type", UserType, "Type of user account")
		Attribute("account_status", AccountStatus, "Status of user account")
		Attribute("tenant_id", String, "Tenant identifier", func() {
			Format(FormatUUID)
			Example("123e4567-e89b-12d3-a456-426614174000")
		})
		Attribute("entity_id", String, "Entity identifier", func() {
			Format(FormatUUID)
			Example("987fcdeb-51d2-43b8-a456-426614174000")
		})
		Attribute("person_id", String, "Person identifier", func() {
			Format(FormatUUID)
			Example("456e7890-e89b-12d3-a456-426614174000")
		})
		Attribute("roles", ArrayOf(RoleInfo), "User roles")
		Attribute("permissions", ArrayOf(PermissionInfo), "User permissions")
		Attribute("last_login", String, "Last login timestamp", func() {
			Format(FormatDateTime)
			Example("2023-12-07T10:30:00Z")
		})
		Attribute("password_expires_at", String, "Password expiration timestamp", func() {
			Format(FormatDateTime)
			Example("2024-06-07T10:30:00Z")
		})
		AuditFields()
	})
	Required("id", "username", "email", "first_name", "last_name", "user_type", "account_status", "tenant_id")
	
	View("default", func() {
		Attribute("id")
		Attribute("username")
		Attribute("email")
		Attribute("first_name")
		Attribute("last_name")
		Attribute("user_type")
		Attribute("account_status")
		Attribute("tenant_id")
		Attribute("entity_id")
		Attribute("person_id")
		Attribute("roles")
		Attribute("permissions")
		Attribute("last_login")
		Attribute("password_expires_at")
		Attribute("created_at")
		Attribute("updated_at")
	})
	
	View("minimal", func() {
		Attribute("id")
		Attribute("username")
		Attribute("email")
		Attribute("first_name")
		Attribute("last_name")
		Attribute("account_status")
	})
})

// CreateUserPayload describes the payload for creating a user
var CreateUserPayload = Type("CreateUserPayload", func() {
	Description("Payload for creating a new user")
	Attribute("username", String, "Username", func() {
		MinLength(3)
		MaxLength(50)
		Example("john.doe")
	})
	Attribute("email", String, "Email address", func() {
		Format(FormatEmail)
		Example("john.doe@acme.com")
	})
	Attribute("password", String, "Password", func() {
		MinLength(8)
		MaxLength(100)
		Example("SecurePassword123!")
	})
	Attribute("first_name", String, "First name", func() {
		MinLength(1)
		MaxLength(50)
		Example("John")
	})
	Attribute("last_name", String, "Last name", func() {
		MinLength(1)
		MaxLength(50)
		Example("Doe")
	})
	Attribute("user_type", UserType, "Type of user account")
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("entity_id", String, "Entity identifier", func() {
		Format(FormatUUID)
		Example("987fcdeb-51d2-43b8-a456-426614174000")
	})
	Attribute("person_id", String, "Person identifier", func() {
		Format(FormatUUID)
		Example("456e7890-e89b-12d3-a456-426614174000")
	})
	Attribute("roles", ArrayOf(String), "Role IDs to assign", func() {
		Example([]string{"role1", "role2"})
	})
	Required("username", "email", "password", "first_name", "last_name", "user_type", "tenant_id")
})

// UpdateUserPayload describes the payload for updating a user
var UpdateUserPayload = Type("UpdateUserPayload", func() {
	Description("Payload for updating an existing user")
	Attribute("id", String, "User ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("username", String, "Username", func() {
		MinLength(3)
		MaxLength(50)
		Example("john.doe")
	})
	Attribute("email", String, "Email address", func() {
		Format(FormatEmail)
		Example("john.doe@acme.com")
	})
	Attribute("first_name", String, "First name", func() {
		MinLength(1)
		MaxLength(50)
		Example("John")
	})
	Attribute("last_name", String, "Last name", func() {
		MinLength(1)
		MaxLength(50)
		Example("Doe")
	})
	Attribute("user_type", UserType, "Type of user account")
	Attribute("account_status", AccountStatus, "Status of user account")
	Attribute("roles", ArrayOf(String), "Role IDs to assign", func() {
		Example([]string{"role1", "role2"})
	})
	Required("id")
})

// RoleInfo describes role information
var RoleInfo = Type("RoleInfo", func() {
	Description("Role information")
	Attribute("id", String, "Role ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("name", String, "Role name", func() {
		Example("Manager")
	})
	Attribute("description", String, "Role description", func() {
		Example("Department manager role")
	})
	Attribute("role_type", RoleType, "Type of role")
	Attribute("permissions", ArrayOf(PermissionInfo), "Role permissions")
})

// PermissionInfo describes permission information
var PermissionInfo = Type("PermissionInfo", func() {
	Description("Permission information")
	Attribute("id", String, "Permission ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("resource_type", ResourceType, "Type of resource")
	Attribute("resource_id", String, "Resource identifier", func() {
		Example("invoice_module")
	})
	Attribute("action", ActionType, "Action type")
	Attribute("conditions", MapOf(String, Any), "Permission conditions")
	Attribute("granted", Boolean, "Whether permission is granted")
})

// LoginRequest describes the login request
var LoginRequest = Type("LoginRequest", func() {
	Description("Login request payload")
	Attribute("username", String, "Username or email", func() {
		MinLength(3)
		MaxLength(100)
		Example("john.doe")
	})
	Attribute("password", String, "Password", func() {
		MinLength(8)
		MaxLength(100)
		Example("SecurePassword123!")
	})
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Required("username", "password")
})

// LoginResponse describes the login response
var LoginResponse = Type("LoginResponse", func() {
	Description("Login response")
	Attribute("access_token", String, "JWT access token", func() {
		Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
	})
	Attribute("refresh_token", String, "JWT refresh token", func() {
		Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
	})
	Attribute("expires_in", UInt, "Token expiration time in seconds", func() {
		Example(3600)
	})
	Attribute("token_type", String, "Token type", func() {
		Example("Bearer")
	})
	Attribute("user", UserResult, "User information")
	Required("access_token", "refresh_token", "expires_in", "token_type", "user")
})