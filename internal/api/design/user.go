package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// USER & AUTHENTICATION TYPES
// ============================================================================

// User represents a system user with authentication and authorization capabilities.
// Users are associated with tenants and entities, and have roles that determine
// their access permissions throughout the ERP system.
var User = ResultType("application/vnd.erp.user", func() {
	Description("System user with authentication, authorization, and profile information for secure ERP access")

	Attributes(func() {
		Field(1, "id", String, "Unique user identifier", func() {
			Format(FormatUUID)
			Example("456e7890-e89b-12d3-a456-426614174000")
			Description("Primary key for user references")
		})

		Field(2, "tenant_id", String, "Associated tenant identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
			Description("Tenant isolation boundary for multi-tenancy")
		})

		Field(3, "entity_id", String, "Primary entity association", func() {
			Format(FormatUUID)
			Example("987fcdeb-51d2-43b8-a456-426614174000")
			Description("Default organizational entity for user access")
		})

		Field(4, "person_id", String, "Associated person record identifier", func() {
			Format(FormatUUID)
			Example("person-123e4567-e89b-12d3-a456-426614174000")
			Description("Links to detailed personal information")
		})

		Field(5, "employee_id", String, "Associated employee record identifier", func() {
			Format(FormatUUID)
			Example("emp-456e7890-e89b-12d3-a456-426614174000")
			Description("Links to employment details if user is an employee")
		})

		Field(6, "email", String, "Primary email address for authentication", func() {
			Format(FormatEmail)
			Example("john.doe@acme.com")
			Description("Unique email used for login and communications")
		})

		Field(7, "username", String, "Unique username for authentication", func() {
			Pattern("^[a-zA-Z0-9][a-zA-Z0-9._-]{1,28}[a-zA-Z0-9]$")
			MinLength(3)
			MaxLength(30)
			Example("john.doe")
			Description("Alternative login identifier")
		})

		Field(8, "user_type", String, "Classification of user account", func() {
			Enum("INTERNAL", "CUSTOMER", "VENDOR", "PARTNER", "API", "SERVICE", "ADMIN", "SYSTEM")
			Default("INTERNAL")
			Example("INTERNAL")
			Description("Determines available features and access patterns")
		})

		Field(9, "account_status", String, "Current account status", func() {
			Enum("ACTIVE", "INACTIVE", "SUSPENDED", "LOCKED", "PENDING_VERIFICATION", "ARCHIVED")
			Default("PENDING_VERIFICATION")
			Example("ACTIVE")
			Description("Controls user access and authentication")
		})

		Field(10, "is_active", Boolean, "Quick active status flag", func() {
			Default(true)
			Example(true)
			Description("Fast check for user availability")
		})

		Field(11, "email_verified", Boolean, "Email verification status", func() {
			Default(false)
			Example(true)
			Description("Required for full account activation")
		})

		Field(12, "phone", String, "Contact phone number", func() {
			Pattern("^\\+?[1-9]\\d{1,14}$")
			Example("+1-555-123-4567")
			Description("Optional phone number for 2FA and contact")
		})

		Field(13, "phone_verified", Boolean, "Phone verification status", func() {
			Default(false)
			Example(true)
			Description("Required for SMS-based 2FA")
		})

		Field(14, "last_login_at", String, "Most recent successful login", func() {
			Format(FormatDateTime)
			Example("2023-12-07T10:30:00Z")
			Description("Tracks user activity for security monitoring")
		})

		Field(15, "last_login_ip", String, "IP address of last login", func() {
			Pattern("^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$")
			Example("192.168.1.100")
			Description("Security tracking for suspicious activity")
		})

		Field(16, "failed_login_attempts", UInt, "Count of consecutive failed logins", func() {
			Maximum(10)
			Default(0)
			Example(0)
			Description("Used for account locking and security alerts")
		})

		Field(17, "mfa_enabled", Boolean, "Multi-factor authentication status", func() {
			Default(false)
			Example(true)
			Description("Enhanced security requirement")
		})

		Field(18, "mfa_method", String, "Preferred MFA method", func() {
			Enum("TOTP", "SMS", "EMAIL", "HARDWARE_TOKEN", "BIOMETRIC")
			Example("TOTP")
			Description("Primary method for second factor authentication")
		})

		Field(19, "session_timeout_minutes", UInt, "Custom session timeout", func() {
			Minimum(5)
			Maximum(1440) // 24 hours
			Default(480)  // 8 hours
			Example(240)
			Description("User-specific session duration override")
		})

		Field(20, "timezone", String, "User's preferred timezone", func() {
			Pattern("^[A-Za-z]+/[A-Za-z_]+$")
			Default("UTC")
			Example("America/New_York")
			Description("IANA timezone for UI display")
		})

		Field(21, "language", String, "Preferred interface language", func() {
			Pattern("^[a-z]{2}(-[A-Z]{2})?$")
			Default("en")
			Example("en-US")
			Description("Locale for UI translation")
		})

		Field(22, "roles", ArrayOf(String), "Assigned role names", func() {
			Example([]string{"department_manager", "budget_approver", "user"})
			Description("List of roles determining user permissions")
		})

		Field(23, "permissions", ArrayOf(String), "Direct permission grants", func() {
			Example([]string{"finance.accounts.read", "hr.employees.write"})
			Description("Additional permissions beyond roles")
		})

		Field(24, "preferences", MapOf(String, Any), "User interface preferences", func() {
			Example(map[string]any{
				"theme":               "dark",
				"date_format":         "MM/DD/YYYY",
				"currency_symbol":     "$",
				"notifications_email": true,
				"dashboard_layout":    []string{"financial_summary", "recent_activity"},
			})
			Description("Personalization settings for user experience")
		})

		Field(25, "metadata", MapOf(String, Any), "Additional user metadata", func() {
			Example(map[string]any{
				"department":       "Engineering",
				"employee_number":  "EMP-12345",
				"hire_date":        "2023-01-15",
				"external_user_id": "ad-user-789",
			})
			Description("Flexible storage for integration and custom data")
		})

		Field(26, "password_changed_at", String, "Last password change timestamp", func() {
			Format(FormatDateTime)
			Example("2023-11-15T14:30:00Z")
			Description("Tracks password age for policy enforcement")
		})

		Field(27, "terms_accepted_at", String, "Terms of service acceptance timestamp", func() {
			Format(FormatDateTime)
			Example("2023-12-01T09:00:00Z")
			Description("Legal compliance tracking")
		})

		Field(28, "privacy_policy_accepted_at", String, "Privacy policy acceptance timestamp", func() {
			Format(FormatDateTime)
			Example("2023-12-01T09:00:00Z")
			Description("Privacy compliance tracking")
		})

		// Audit fields from common.go
		AuditFields()

		Required("id", "tenant_id", "email", "username", "user_type", "account_status", "is_active", "created_at")
	})

	View("default", func() {
		Description("Standard user view for general listings and references")
		Attribute("id")
		Attribute("email")
		Attribute("username")
		Attribute("user_type")
		Attribute("account_status")
		Attribute("is_active")
		Attribute("roles")
		Attribute("last_login_at")
		Attribute("created_at")
	})

	View("detailed", func() {
		Description("Complete user profile with security and preference details")
		Attribute("id")
		Attribute("tenant_id")
		Attribute("entity_id")
		Attribute("person_id")
		Attribute("employee_id")
		Attribute("email")
		Attribute("username")
		Attribute("user_type")
		Attribute("account_status")
		Attribute("is_active")
		Attribute("email_verified")
		Attribute("phone")
		Attribute("phone_verified")
		Attribute("last_login_at")
		Attribute("last_login_ip")
		Attribute("mfa_enabled")
		Attribute("mfa_method")
		Attribute("session_timeout_minutes")
		Attribute("timezone")
		Attribute("language")
		Attribute("roles")
		Attribute("permissions")
		Attribute("preferences")
		Attribute("metadata")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})

	View("security", func() {
		Description("Security-focused view for admin and audit purposes")
		Attribute("id")
		Attribute("email")
		Attribute("username")
		Attribute("account_status")
		Attribute("last_login_at")
		Attribute("last_login_ip")
		Attribute("failed_login_attempts")
		Attribute("mfa_enabled")
		Attribute("password_changed_at")
		Attribute("created_at")
		Attribute("updated_at")
	})

	View("public", func() {
		Description("Minimal public view for safe external references")
		Attribute("id")
		Attribute("username")
		Attribute("user_type")
	})
})

// AuthResponse represents successful authentication response with tokens and user info.
// Contains all necessary information for client-side session management.
var AuthResponse = Type("AuthResponse", func() {
	Description("Authentication response containing tokens and user session information")

	Field(1, "access_token", String, "JWT access token for API authentication", func() {
		Pattern("^[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]*$")
		Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c")
		Description("Bearer token for API requests")
	})

	Field(2, "refresh_token", String, "Refresh token for obtaining new access tokens", func() {
		Pattern("^[A-Za-z0-9-_]{40,}$")
		Example("refresh_8f4a0e7b3c9d2e1f6a5b4c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f")
		Description("Long-lived token for token renewal")
	})

	Field(3, "token_type", String, "Token type specification", func() {
		Enum("Bearer", "MAC")
		Default("Bearer")
		Example("Bearer")
		Description("Authorization header format")
	})

	Field(4, "expires_in", UInt, "Access token expiry time in seconds", func() {
		Minimum(300)   // 5 minutes minimum
		Maximum(86400) // 24 hours maximum
		Default(3600)  // 1 hour default
		Example(3600)
		Description("Time until access token expires")
	})

	Field(5, "refresh_expires_in", UInt, "Refresh token expiry time in seconds", func() {
		Minimum(86400)   // 1 day minimum
		Maximum(7776000) // 90 days maximum
		Default(604800)  // 7 days default
		Example(604800)
		Description("Time until refresh token expires")
	})

	Field(6, "scope", String, "Granted token scope", func() {
		Example("api:read api:write admin")
		Description("Space-separated list of granted permissions")
	})

	Field(7, "user", User, "Authenticated user information", func() {
		Description("User profile data for session establishment")
	})

	Field(8, "session_id", String, "Unique session identifier", func() {
		Format(FormatUUID)
		Example("session-123e4567-e89b-12d3-a456-426614174000")
		Description("Server-side session tracking")
	})

	Field(9, "issued_at", String, "Token issuance timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
		Description("When the tokens were created")
	})

	Field(10, "mfa_required", Boolean, "Multi-factor authentication requirement", func() {
		Default(false)
		Example(false)
		Description("Indicates if additional authentication steps are needed")
	})

	Field(11, "mfa_token", String, "Temporary MFA challenge token", func() {
		Pattern("^[A-Za-z0-9-_]{32,}$")
		Example("mfa_temp_abc123def456ghi789jkl012mno345pqr678stu")
		Description("Used for completing MFA flow")
	})

	Required("access_token", "token_type", "expires_in", "user", "session_id", "issued_at")
})

// UserSession represents an active user session for tracking and management.
var UserSession = Type("UserSession", func() {
	Description("Active user session information for security monitoring and management")

	Field(1, "session_id", String, "Unique session identifier", func() {
		Format(FormatUUID)
		Example("session-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary session reference")
	})

	Field(2, "user_id", String, "Associated user identifier", func() {
		Format(FormatUUID)
		Example("456e7890-e89b-12d3-a456-426614174000")
		Description("User owning this session")
	})

	Field(3, "tenant_id", String, "Tenant context", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Description("Tenant isolation for session")
	})

	Field(4, "ip_address", String, "Client IP address", func() {
		Pattern("^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$")
		Example("192.168.1.100")
		Description("Source IP for security tracking")
	})

	Field(5, "user_agent", String, "Client user agent string", func() {
		MaxLength(500)
		Example("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		Description("Browser/client identification")
	})

	Field(6, "device_fingerprint", String, "Device identification hash", func() {
		Pattern("^[a-f0-9]{64}$")
		Example("a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2")
		Description("Unique device identifier for security")
	})

	Field(7, "started_at", String, "Session start timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
		Description("When session was established")
	})

	Field(8, "last_activity_at", String, "Last activity timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T14:45:00Z")
		Description("Most recent session activity")
	})

	Field(9, "expires_at", String, "Session expiration timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T18:30:00Z")
		Description("When session will automatically expire")
	})

	Field(10, "is_active", Boolean, "Session active status", func() {
		Default(true)
		Example(true)
		Description("Whether session is currently valid")
	})

	Field(11, "logout_reason", String, "Reason for session termination", func() {
		Enum("USER_LOGOUT", "TIMEOUT", "ADMIN_TERMINATED", "SECURITY_VIOLATION", "TOKEN_EXPIRED")
		Example("USER_LOGOUT")
		Description("How/why the session ended")
	})

	Required("session_id", "user_id", "tenant_id", "ip_address", "started_at", "expires_at", "is_active")
})

// UserPreferences represents user customization and interface preferences
var UserPreferences = Type("UserPreferences", func() {
	Description("User interface customization and personal preference settings")

	Field(1, "user_id", String, "Associated user identifier", func() {
		Format(FormatUUID)
		Example("456e7890-e89b-12d3-a456-426614174000")
	})

	Field(2, "theme", String, "UI theme preference", func() {
		Enum("light", "dark", "auto", "high_contrast")
		Default("light")
		Example("dark")
		Description("Visual theme for user interface")
	})

	Field(3, "language", String, "Interface language", func() {
		Pattern("^[a-z]{2}(-[A-Z]{2})?$")
		Default("en")
		Example("en-US")
		Description("Locale for UI translation")
	})

	Field(4, "timezone", String, "Display timezone", func() {
		Pattern("^[A-Za-z]+/[A-Za-z_]+$")
		Default("UTC")
		Example("America/New_York")
		Description("IANA timezone for date/time display")
	})

	Field(5, "date_format", String, "Preferred date format", func() {
		Enum("MM/DD/YYYY", "DD/MM/YYYY", "YYYY-MM-DD", "DD-MMM-YYYY")
		Default("MM/DD/YYYY")
		Example("MM/DD/YYYY")
		Description("Date display format preference")
	})

	Field(6, "time_format", String, "Preferred time format", func() {
		Enum("12h", "24h")
		Default("12h")
		Example("24h")
		Description("Time display format preference")
	})

	Field(7, "currency_display", String, "Currency formatting preference", func() {
		Enum("symbol", "code", "symbol_code")
		Default("symbol")
		Example("symbol")
		Description("How to display currency values")
	})

	Field(8, "number_format", String, "Number formatting locale", func() {
		Enum("US", "EU", "IN", "UK")
		Default("US")
		Example("US")
		Description("Number separators and formatting")
	})

	Field(9, "dashboard_layout", ArrayOf(String), "Dashboard widget arrangement", func() {
		Example([]string{
			"financial_summary",
			"recent_transactions",
			"pending_approvals",
			"team_activity",
		})
		Description("Ordered list of dashboard components")
	})

	Field(10, "items_per_page", UInt, "Default pagination size", func() {
		Minimum(10)
		Maximum(100)
		Default(20)
		Example(25)
		Description("Number of items to show in lists")
	})

	Field(11, "notifications", MapOf(String, Boolean), "Notification preferences", func() {
		Example(map[string]any{
			"email_digest":       true,
			"browser_push":       false,
			"sms_alerts":         true,
			"approval_requests":  true,
			"system_maintenance": false,
		})
		Description("Enable/disable various notification types")
	})

	Field(12, "shortcuts", MapOf(String, String), "Custom keyboard shortcuts", func() {
		Example(map[string]any{
			"new_transaction": "Ctrl+N",
			"search":          "Ctrl+K",
			"dashboard":       "Ctrl+H",
		})
		Description("User-defined keyboard shortcuts")
	})

	// Audit fields
	AuditFields()

	Required("user_id", "theme", "language", "timezone", "date_format")
})
