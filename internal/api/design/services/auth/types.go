package auth

import (
	. "goa.design/goa/v3/dsl"
)

// AuthResult describes the authentication response
var AuthResult = ResultType("application/vnd.auth", func() {
	Description("Authentication result")
	Attributes(func() {
		Attribute("access_token", String, "JWT access token", func() {
			Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
		})
		Attribute("refresh_token", String, "Refresh token", func() {
			Example("refresh_token_here")
		})
		Attribute("token_type", String, "Token type", func() {
			Enum("Bearer")
			Default("Bearer")
			Example("Bearer")
		})
		Attribute("expires_in", UInt, "Token expiration time in seconds", func() {
			Example(3600)
		})
		Attribute("user", UserInfo, "User information")
		Attribute("tenant", TenantInfo, "Tenant information")
		Attribute("permissions", ArrayOf(String), "User permissions", func() {
			Example([]string{"users.read", "users.write", "tenants.read"})
		})
	})
	Required("access_token", "token_type", "expires_in", "user")
})

// TokenValidationResult describes token validation response
var TokenValidationResult = ResultType("application/vnd.token.validation", func() {
	Description("Token validation result")
	Attributes(func() {
		Attribute("valid", Boolean, "Whether token is valid", func() {
			Example(true)
		})
		Attribute("user", UserInfo, "User information")
		Attribute("tenant", TenantInfo, "Tenant information")
		Attribute("permissions", ArrayOf(String), "User permissions", func() {
			Example([]string{"users.read", "users.write"})
		})
		Attribute("expires_at", String, "Token expiration timestamp", func() {
			Format(FormatDateTime)
			Example("2023-12-07T15:45:00Z")
		})
	})
	Required("valid")
})

// UserInfo describes basic user information in auth context
var UserInfo = Type("UserInfo", func() {
	Description("User information for authentication")
	Attribute("id", String, "User ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("email", String, "User email", func() {
		Format(FormatEmail)
		Example("user@example.com")
	})
	Attribute("first_name", String, "First name", func() {
		Example("John")
	})
	Attribute("last_name", String, "Last name", func() {
		Example("Doe")
	})
	Attribute("role", String, "User role", func() {
		Example("admin")
	})
	Required("id", "email", "first_name", "last_name")
})

// TenantInfo describes basic tenant information in auth context
var TenantInfo = Type("TenantInfo", func() {
	Description("Tenant information for authentication")
	Attribute("id", String, "Tenant ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("name", String, "Tenant name", func() {
		Example("Acme Corporation")
	})
	Attribute("subdomain", String, "Tenant subdomain", func() {
		Example("acme")
	})
	Attribute("status", String, "Tenant status", func() {
		Enum("active", "inactive", "suspended")
		Example("active")
	})
	Required("id", "name", "subdomain", "status")
})
