package design

import (
	. "goa.design/goa/v3/dsl"
)

// API defines the User Management & Permissions service
var _ = API("user-management", func() {
	Title("User Management & Permissions Service")
	Description("Enterprise-grade identity and access management system with RBAC and ABAC capabilities")
	Version("1.0")
	Server("user-management", func() {
		Host("localhost", func() {
			URI("http://localhost:8080")
			URI("grpc://localhost:8090")
		})
	})
})

// =============================================================================
// SECURITY DEFINITIONS
// =============================================================================

var JWT = JWTSecurity("jwt", func() {
	Description("JWT token authentication")
	Scope("api:access", "Access to API")
})

// var APIKey = APIKeySecurity("api_key", func() {
// 	Description("API key authentication")
// 	Header("X-API-Key")
// })
//
// =============================================================================
// COMMON HEADERS
// =============================================================================

var CommonHeaders = func() {
	Header("X-Tenant-ID", String, "Tenant identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Header("X-Entity-ID", String, "Entity/Company identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("987fcdeb-51d2-43b8-a456-426614174000")
	})
	Header("X-User-ID", String, "Current user identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("456e7890-e89b-12d3-a456-426614174000")
	})
}

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

var ActionCategory = Type("ActionCategory", String, func() {
	Description("Category of action based on sensitivity")
	Enum("STANDARD", "ADMINISTRATIVE", "SENSITIVE", "BULK", "SYSTEM")
	Example("STANDARD")
})

var RiskLevel = Type("RiskLevel", String, func() {
	Description("Risk level associated with an action")
	Enum("LOW", "MEDIUM", "HIGH", "CRITICAL")
	Example("LOW")
})

var PermissionEffect = Type("PermissionEffect", String, func() {
	Description("Effect of a permission grant")
	Enum("ALLOW", "DENY")
	Example("ALLOW")
})

var PolicyType = Type("PolicyType", String, func() {
	Description("Type of access control policy")
	Enum("ABAC", "RBAC", "HYBRID", "TIME_BASED", "LOCATION_BASED")
	Example("ABAC")
})

var PolicyCategory = Type("PolicyCategory", String, func() {
	Description("Category of policy based on purpose")
	Enum("ACCESS", "DATA_FILTER", "FIELD_MASK", "AUDIT", "COMPLIANCE")
	Example("ACCESS")
})

var AssignmentType = Type("AssignmentType", String, func() {
	Description("Type of role assignment")
	Enum("DIRECT", "INHERITED", "DELEGATED", "TEMPORARY")
	Example("DIRECT")
})

var RequestType = Type("RequestType", String, func() {
	Description("Type of access request")
	Enum("ROLE_ASSIGNMENT", "PERMISSION_GRANT", "RESOURCE_ACCESS", "ELEVATION")
	Example("ROLE_ASSIGNMENT")
})

var ApprovalStatus = Type("ApprovalStatus", String, func() {
	Description("Status of approval request")
	Enum("PENDING", "APPROVED", "REJECTED", "EXPIRED", "REVOKED")
	Example("PENDING")
})

var EventCategory = Type("EventCategory", String, func() {
	Description("Category of audit event")
	Enum("ACCESS", "ADMIN", "DATA", "AUTH", "SYSTEM", "COMPLIANCE")
	Example("ACCESS")
})

var EventSeverity = Type("EventSeverity", String, func() {
	Description("Severity level of audit event")
	Enum("LOW", "INFO", "WARN", "HIGH", "CRITICAL")
	Example("INFO")
})

var DataType = Type("DataType", String, func() {
	Description("Data type for attribute definitions")
	Enum("STRING", "NUMBER", "BOOLEAN", "DATE", "TIME", "JSON", "ARRAY", "ENUM")
	Example("STRING")
})

var AttributeCategory = Type("AttributeCategory", String, func() {
	Description("Category of ABAC attribute")
	Enum("USER", "RESOURCE", "ENVIRONMENT", "ACTION", "ENTITY", "SESSION")
	Example("USER")
})

// =============================================================================
// CORE TYPE DEFINITIONS
// =============================================================================

var Address = Type("Address", func() {
	Description("Physical address information")
	Attribute("street", String, "Street address", func() {
		MaxLength(255)
		Example("123 Main Street")
	})
	Attribute("city", String, "City name", func() {
		MaxLength(100)
		Example("New York")
	})
	Attribute("state", String, "State or province", func() {
		MaxLength(100)
		Example("NY")
	})
	Attribute("postal_code", String, "Postal or ZIP code", func() {
		MaxLength(20)
		Example("10001")
	})
	Attribute("country", String, "Country name", func() {
		MaxLength(100)
		Example("United States")
	})
	Required("street", "city", "country")
})

var Metadata = Type("Metadata", func() {
	Description("Flexible metadata storage")
	AdditionalProperties(Any)
	Example(map[string]interface{}{
		"department":  "Engineering",
		"cost_center": "ENG-001",
		"location":    "New York Office",
	})
})

var SecurityAttributes = Type("SecurityAttributes", func() {
	Description("Security attributes for ABAC evaluation")
	AdditionalProperties(Any)
	Example(map[string]interface{}{
		"clearance_level": "SECRET",
		"need_to_know":    []string{"PROJECT_ALPHA", "FINANCE_DATA"},
		"citizenship":     "US",
	})
})

var WorkSchedule = Type("WorkSchedule", func() {
	Description("Employee work schedule information")
	Attribute("days_of_week", ArrayOf(Int), "Working days (1=Monday, 7=Sunday)", func() {
		MinLength(1)
		MaxLength(7)
		Example([]int{1, 2, 3, 4, 5})
	})
	Attribute("start_time", String, "Start time in HH:MM format", func() {
		Pattern("^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$")
		Example("09:00")
	})
	Attribute("end_time", String, "End time in HH:MM format", func() {
		Pattern("^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$")
		Example("17:00")
	})
	Attribute("timezone", String, "Timezone identifier", func() {
		Example("America/New_York")
	})
	Attribute("is_flexible", Boolean, "Whether schedule is flexible", func() {
		Default(false)
	})
	Required("days_of_week", "start_time", "end_time", "timezone")
})

var SalaryInfo = Type("SalaryInfo", func() {
	Description("Employee salary information (encrypted)")
	Attribute("annual_salary", Number, "Annual salary amount", func() {
		Minimum(0)
		Example(75000)
	})
	Attribute("currency", String, "Currency code", func() {
		Pattern("^[A-Z]{3}$")
		Example("USD")
	})
	Attribute("effective_date", String, "Effective date of salary", func() {
		Format(FormatDate)
		Example("2024-01-01")
	})
	Attribute("pay_frequency", String, "Pay frequency", func() {
		Enum("WEEKLY", "BI_WEEKLY", "MONTHLY", "ANNUALLY")
		Example("BI_WEEKLY")
	})
	Required("annual_salary", "currency", "effective_date")
})

var DeviceInfo = Type("DeviceInfo", func() {
	Description("Device information for session tracking")
	Attribute("device_id", String, "Unique device identifier", func() {
		MaxLength(255)
		Example("device_123456")
	})
	Attribute("device_type", String, "Type of device", func() {
		Enum("DESKTOP", "LAPTOP", "TABLET", "MOBILE", "SERVER")
		Example("LAPTOP")
	})
	Attribute("os", String, "Operating system", func() {
		MaxLength(100)
		Example("Windows 11")
	})
	Attribute("browser", String, "Browser information", func() {
		MaxLength(200)
		Example("Chrome 121.0.0.0")
	})
	Attribute("is_trusted", Boolean, "Whether device is trusted", func() {
		Default(false)
	})
})

var LocationInfo = Type("LocationInfo", func() {
	Description("Geographic location information")
	Attribute("latitude", Number, "Latitude coordinate", func() {
		Minimum(-90)
		Maximum(90)
		Example(40.7128)
	})
	Attribute("longitude", Number, "Longitude coordinate", func() {
		Minimum(-180)
		Maximum(180)
		Example(-74.0060)
	})
	Attribute("city", String, "City name", func() {
		MaxLength(100)
		Example("New York")
	})
	Attribute("country", String, "Country code", func() {
		Pattern("^[A-Z]{2}$")
		Example("US")
	})
	Attribute("timezone", String, "Timezone identifier", func() {
		Example("America/New_York")
	})
	Attribute("accuracy_meters", Int, "Location accuracy in meters", func() {
		Minimum(0)
		Example(10)
	})
})

var PolicyTarget = Type("PolicyTarget", func() {
	Description("Target specification for ABAC policy")
	Attribute("users", Metadata, "User attribute matching criteria")
	Attribute("resources", Metadata, "Resource attribute matching criteria")
	Attribute("actions", ArrayOf(String), "List of applicable actions", func() {
		Example([]string{"READ", "UPDATE"})
	})
	Attribute("entities", ArrayOf(String), "List of applicable entity IDs", func() {
		Example([]string{"entity-123", "entity-456"})
	})
})

var PolicyRule = Type("PolicyRule", func() {
	Description("Rule logic for ABAC policy evaluation")
	Attribute("and", ArrayOf(Metadata), "Conditions that must all be true")
	Attribute("or", ArrayOf(Metadata), "Conditions where at least one must be true")
	Attribute("not", Metadata, "Condition that must be false")
})

var PolicyObligations = Type("PolicyObligations", func() {
	Description("Required actions when policy applies")
	Attribute("log_access", Boolean, "Require access logging", func() {
		Default(false)
	})
	Attribute("notify_manager", Boolean, "Notify user's manager", func() {
		Default(false)
	})
	Attribute("require_justification", Boolean, "Require access justification", func() {
		Default(false)
	})
	Attribute("time_limit_hours", Int, "Time limit for access in hours", func() {
		Minimum(1)
		Maximum(8760) // 1 year
		Example(8)
	})
})

var DataFilters = Type("DataFilters", func() {
	Description("Row-level security filters")
	Attribute("sql_conditions", ArrayOf(String), "SQL WHERE conditions", func() {
		Example([]string{"department_id = user.department_id", "created_by = user.id"})
	})
	Attribute("entity_scope", ArrayOf(String), "Limited to specific entities", func() {
		Example([]string{"entity-123"})
	})
	Attribute("date_range", Metadata, "Date range restrictions")
})

var FieldRestrictions = Type("FieldRestrictions", func() {
	Description("Column-level access restrictions")
	Attribute("allowed_fields", ArrayOf(String), "Fields user can access", func() {
		Example([]string{"name", "email", "department"})
	})
	Attribute("denied_fields", ArrayOf(String), "Fields user cannot access", func() {
		Example([]string{"salary", "ssn"})
	})
	Attribute("masked_fields", ArrayOf(String), "Fields that should be masked", func() {
		Example([]string{"phone", "address"})
	})
})

var ValidationRules = Type("ValidationRules", func() {
	Description("Validation rules for attribute values")
	Attribute("min_length", Int, "Minimum string length", func() {
		Minimum(0)
		Example(3)
	})
	Attribute("max_length", Int, "Maximum string length", func() {
		Minimum(1)
		Example(255)
	})
	Attribute("pattern", String, "Regular expression pattern", func() {
		Example("^[A-Z]{2,5}$")
	})
	Attribute("min_value", Number, "Minimum numeric value", func() {
		Example(0)
	})
	Attribute("max_value", Number, "Maximum numeric value", func() {
		Example(100)
	})
})

// =============================================================================
// ENTITY DEFINITIONS
// =============================================================================

var Person = Type("Person", func() {
	Description("Person entity representing individuals across various contexts")
	Attribute("id", String, "Unique person identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("tenant-123")
	})
	Attribute("entity_id", String, "Associated entity identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("entity-456")
	})
	Attribute("person_type", PersonType, "Type of person")
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
	Attribute("middle_name", String, "Middle name", func() {
		MaxLength(100)
		Example("Michael")
	})
	Attribute("email", String, "Email address", func() {
		Format(FormatEmail)
		MaxLength(255)
		Example("john.doe@company.com")
	})
	Attribute("phone", String, "Phone number", func() {
		Pattern("^\\+?[1-9]\\d{1,14}$")
		Example("+1-555-123-4567")
	})
	Attribute("birth_date", String, "Date of birth", func() {
		Format(FormatDate)
		Example("1990-01-01")
	})
	Attribute("national_id", String, "National identification number", func() {
		MaxLength(50)
		Example("123-45-6789")
	})
	Attribute("tax_id", String, "Tax identification number", func() {
		MaxLength(50)
		Example("12-3456789")
	})
	Attribute("address", Address, "Physical address")
	Attribute("security_attributes", SecurityAttributes, "ABAC security attributes")
	Attribute("metadata", Metadata, "Additional metadata")
	Attribute("is_active", Boolean, "Whether person is active", func() {
		Default(true)
	})
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Attribute("updated_at", String, "Last update timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Attribute("deleted_at", String, "Soft deletion timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Required("tenant_id", "entity_id", "person_type", "first_name", "last_name")
})

var Employee = Type("Employee", func() {
	Description("Employee entity extending person with employment-specific data")
	Attribute("id", String, "Unique employee identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("emp-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("tenant-123")
	})
	Attribute("person_id", String, "Associated person identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("person-456")
	})
	Attribute("entity_id", String, "Associated entity identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("entity-789")
	})
	Attribute("employee_number", String, "Unique employee number", func() {
		MinLength(1)
		MaxLength(50)
		Example("EMP-001234")
	})
	Attribute("position_title", String, "Job position title", func() {
		MaxLength(100)
		Example("Senior Software Engineer")
	})
	Attribute("department_id", String, "Department entity identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("dept-123")
	})
	Attribute("manager_id", String, "Manager employee identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("mgr-456")
	})
	Attribute("hire_date", String, "Date of hire", func() {
		Format(FormatDate)
		Example("2024-01-01")
	})
	Attribute("termination_date", String, "Date of termination", func() {
		Format(FormatDate)
		Example("2024-12-31")
	})
	Attribute("salary_info", SalaryInfo, "Salary information")
	Attribute("employment_status", EmploymentStatus, "Current employment status")
	Attribute("work_schedule", WorkSchedule, "Work schedule details")
	Attribute("security_level", Int, "Security clearance level (0=lowest)", func() {
		Minimum(0)
		Maximum(10)
		Default(0)
		Example(3)
	})
	Attribute("access_attributes", SecurityAttributes, "ABAC access attributes")
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Attribute("updated_at", String, "Last update timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Attribute("deleted_at", String, "Soft deletion timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Required("tenant_id", "person_id", "entity_id", "employee_number", "hire_date", "employment_status")
})

var User = Type("User", func() {
	Description("System user account with authentication data and security controls")
	Attribute("id", String, "Unique user identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("user-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("tenant-123")
	})
	Attribute("entity_id", String, "Associated entity identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("entity-456")
	})
	Attribute("person_id", String, "Associated person identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("person-789")
	})
	Attribute("employee_id", String, "Associated employee identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("emp-012")
	})
	Attribute("username", String, "Unique username", func() {
		MinLength(3)
		MaxLength(100)
		Pattern("^[a-zA-Z0-9._-]+$")
		Example("john.doe")
	})
	Attribute("email", String, "Email address", func() {
		Format(FormatEmail)
		MaxLength(255)
		Example("john.doe@company.com")
	})
	Attribute("user_type", UserType, "Type of user account")
	Attribute("account_status", AccountStatus, "Current account status")
	Attribute("is_active", Boolean, "Whether user account is active", func() {
		Default(true)
	})
	Attribute("last_login_at", String, "Last login timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Attribute("password_changed_at", String, "Password last changed timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Attribute("failed_login_attempts", Int, "Number of failed login attempts", func() {
		Minimum(0)
		Default(0)
		Example(0)
	})
	Attribute("lockout_until", String, "Account lockout expiration", func() {
		Format(FormatDateTime)
		Example("2024-01-01T11:00:00Z")
	})
	Attribute("session_timeout_minutes", Int, "Session timeout in minutes", func() {
		Minimum(5)
		Maximum(1440) // 24 hours
		Default(480)  // 8 hours
		Example(480)
	})
	Attribute("mfa_enabled", Boolean, "Whether MFA is enabled", func() {
		Default(false)
	})
	Attribute("user_attributes", SecurityAttributes, "ABAC user attributes")
	Attribute("settings", Metadata, "User preferences and settings")
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Attribute("updated_at", String, "Last update timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Attribute("deleted_at", String, "Soft deletion timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Required("tenant_id", "entity_id", "email", "user_type")
})

var UserSession = Type("UserSession", func() {
	Description("User session information for tracking and security")
	Attribute("id", String, "Unique session identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("session-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("tenant-123")
	})
	Attribute("user_id", String, "User identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("user-456")
	})
	Attribute("session_token", String, "Session token", func() {
		MinLength(32)
		MaxLength(255)
		Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
	})
	Attribute("refresh_token", String, "Refresh token", func() {
		MinLength(32)
		MaxLength(255)
		Example("refresh_eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
	})
	Attribute("ip_address", String, "Client IP address", func() {
		Pattern("^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}$|^(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$")
		Example("192.168.1.100")
	})
	Attribute("user_agent", String, "Client user agent", func() {
		MaxLength(500)
		Example("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	})
	Attribute("device_info", DeviceInfo, "Device information")
	Attribute("location_info", LocationInfo, "Location information")
	Attribute("expires_at", String, "Session expiration timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T18:00:00Z")
	})
	Attribute("created_at", String, "Session creation timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Attribute("last_accessed_at", String, "Last access timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T15:30:00Z")
	})
	Attribute("is_active", Boolean, "Whether session is active", func() {
		Default(true)
	})
	Required("tenant_id", "user_id", "session_token", "expires_at")
})

// =============================================================================
// RBAC DEFINITIONS
// =============================================================================

var Module = Type("Module", func() {
	Description("System module for organizing functionality")
	Attribute("id", String, "Unique module identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("module-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("tenant-123")
	})
	Attribute("name", String, "Module name", func() {
		MinLength(1)
		MaxLength(50)
		Pattern("^[A-Z_]+$")
		Example("HR_MANAGEMENT")
	})
	Attribute("display_name", String, "Human-readable module name", func() {
		MaxLength(100)
		Example("Human Resources Management")
	})
	Attribute("description", String, "Module description", func() {
		MaxLength(500)
		Example("Module for managing employee data and HR processes")
	})
	Attribute("category", String, "Module category", func() {
		Enum("CORE", "HR", "FINANCE", "SALES", "INVENTORY", "PROJECT", "CUSTOM")
		Example("HR")
	})
	Attribute("version", String, "Module version", func() {
		Pattern("^\\d+\\.\\d+\\.\\d+$")
		Example("1.2.3")
	})
	Attribute("is_active", Boolean, "Whether module is active", func() {
		Default(true)
	})
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Required("tenant_id", "name", "display_name", "category")
})

var Resource = Type("Resource", func() {
	Description("System resource that can be protected")
	Attribute("id", String, "Unique resource identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("resource-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("tenant-123")
	})
	Attribute("module_id", String, "Associated module identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("module-456")
	})
	Attribute("entity_id", String, "Associated entity identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("entity-789")
	})
	Attribute("name", String, "Resource name", func() {
		MinLength(1)
		MaxLength(100)
		Pattern("^[a-z0-9._-]+$")
		Example("employee.records")
	})
	Attribute("display_name", String, "Human-readable resource name", func() {
		MaxLength(150)
		Example("Employee Records")
	})
	Attribute("description", String, "Resource description", func() {
		MaxLength(500)
		Example("Employee personal and employment information")
	})
	Attribute("resource_type", ResourceType, "Type of resource")
	Attribute("parent_resource_id", String, "Parent resource identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("parent-resource-123")
	})
	Attribute("path", String, "Resource path (URL, API endpoint, file path)", func() {
		MaxLength(500)
		Example("/api/v1/employees")
	})
	Attribute("resource_attributes", SecurityAttributes, "ABAC resource attributes")
	Attribute("is_active", Boolean, "Whether resource is active", func() {
		Default(true)
	})
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Attribute("deleted_at", String, "Soft deletion timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Required("tenant_id", "module_id", "name", "resource_type")
})

var Action = Type("Action", func() {
	Description("Action that can be performed on resources")
	Attribute("id", String, "Unique action identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("action-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("tenant-123")
	})
	Attribute("name", String, "Action name", func() {
		MinLength(1)
		MaxLength(100)
		Pattern("^[a-z0-9._-]+$")
		Example("employee.read")
	})
	Attribute("display_name", String, "Human-readable action name", func() {
		MaxLength(150)
		Example("Read Employee Information")
	})
	Attribute("description", String, "Action description", func() {
		MaxLength(500)
		Example("Permission to view employee information")
	})
	Attribute("action_type", ActionType, "Type of action")
	Attribute("action_category", ActionCategory, "Category of action")
	Attribute("risk_level", RiskLevel, "Risk level of action")
	Attribute("requires_approval", Boolean, "Whether action requires approval", func() {
		Default(false)
	})
	Attribute("is_active", Boolean, "Whether action is active", func() {
		Default(true)
	})
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Required("tenant_id", "name", "action_type", "action_category", "risk_level")
})

var Role = Type("Role", func() {
	Description("Role definition with permissions and hierarchy")
	Attribute("id", String, "Unique role identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("role-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("tenant-123")
	})
	Attribute("entity_id", String, "Associated entity identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("entity-456")
	})
	Attribute("name", String, "Role name", func() {
		MinLength(1)
		MaxLength(50)
		Pattern("^[A-Z_]+$")
		Example("HR_MANAGER")
	})
	Attribute("display_name", String, "Human-readable role name", func() {
		MaxLength(100)
		Example("HR Manager")
	})
	Attribute("description", String, "Role description", func() {
		MaxLength(500)
		Example("Role for managing HR processes and employee data")
	})
	Attribute("module_id", String, "Associated module identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("module-789")
	})
	Attribute("role_type", RoleType, "Type of role")
	Attribute("parent_role_id", String, "Parent role identifier for hierarchy", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("parent-role-123")
	})
	Attribute("level", Int, "Calculated hierarchy level", func() {
		Minimum(0)
		Default(0)
		Example(2)
	})
	Attribute("permissions", Metadata, "Cached permission calculations")
	Attribute("entity_scope", Metadata, "Entity access rules")
	Attribute("conditions", Metadata, "Conditional access rules")
	Attribute("is_system_role", Boolean, "Whether role is system-defined", func() {
		Default(false)
	})
	Attribute("is_active", Boolean, "Whether role is active", func() {
		Default(true)
	})
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Attribute("updated_at", String, "Last update timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Attribute("deleted_at", String, "Soft deletion timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Required("tenant_id", "entity_id", "name", "role_type")
})

var Permission = Type("Permission", func() {
	Description("Granular permission definition")
	Attribute("id", String, "Unique permission identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("permission-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("tenant-123")
	})
	Attribute("resource_id", String, "Associated resource identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("resource-456")
	})
	Attribute("action_id", String, "Associated action identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("action-789")
	})
	Attribute("name", String, "Permission name", func() {
		MinLength(1)
		MaxLength(200)
		Example("employee.records.read")
	})
	Attribute("display_name", String, "Human-readable permission name", func() {
		MaxLength(250)
		Example("Read Employee Records")
	})
	Attribute("description", String, "Permission description", func() {
		MaxLength(500)
		Example("Permission to read employee records and personal information")
	})
	Attribute("effect", PermissionEffect, "Permission effect (ALLOW/DENY)")
	Attribute("conditions", Metadata, "ABAC conditions for permission")
	Attribute("data_filters", DataFilters, "Row-level security filters")
	Attribute("field_restrictions", FieldRestrictions, "Column-level access restrictions")
	Attribute("is_active", Boolean, "Whether permission is active", func() {
		Default(true)
	})
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Required("tenant_id", "resource_id", "action_id", "name", "effect")
})

var UserRole = Type("UserRole", func() {
	Description("User role assignment with conditions")
	Attribute("id", String, "Unique assignment identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("assignment-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("user_id", String, "User identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("user-456")
	})
	Attribute("role_id", String, "Role identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("role-789")
	})
	Attribute("entity_id", String, "Entity scope identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("entity-012")
	})
	Attribute("assignment_type", AssignmentType, "Type of assignment")
	Attribute("delegated_by", String, "User who delegated the role", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("delegator-345")
	})
	Attribute("assigned_at", String, "Assignment timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Attribute("assigned_by", String, "User who made the assignment", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("admin-678")
	})
	Attribute("expires_at", String, "Assignment expiration timestamp", func() {
		Format(FormatDateTime)
		Example("2024-12-31T23:59:59Z")
	})
	Attribute("conditions", Metadata, "Conditional access rules")
	Attribute("is_active", Boolean, "Whether assignment is active", func() {
		Default(true)
	})
	Required("user_id", "role_id", "entity_id", "assignment_type")
})

// =============================================================================
// ABAC DEFINITIONS
// =============================================================================

var AttributeDefinition = Type("AttributeDefinition", func() {
	Description("Definition of ABAC attribute for consistent evaluation")
	Attribute("id", String, "Unique attribute definition identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("attr-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("tenant-123")
	})
	Attribute("name", String, "Attribute name", func() {
		MinLength(1)
		MaxLength(100)
		Pattern("^[a-z_][a-z0-9_]*$")
		Example("security_clearance")
	})
	Attribute("display_name", String, "Human-readable attribute name", func() {
		MaxLength(150)
		Example("Security Clearance Level")
	})
	Attribute("description", String, "Attribute description", func() {
		MaxLength(500)
		Example("Security clearance level for access to classified information")
	})
	Attribute("data_type", DataType, "Data type of attribute")
	Attribute("category", AttributeCategory, "Category of attribute")
	Attribute("is_required", Boolean, "Whether attribute is required", func() {
		Default(false)
	})
	Attribute("is_sensitive", Boolean, "Whether attribute contains PII", func() {
		Default(false)
	})
	Attribute("default_value", String, "Default value for attribute", func() {
		MaxLength(255)
		Example("PUBLIC")
	})
	Attribute("allowed_values", ArrayOf(String), "Valid values for enum types", func() {
		Example([]string{"PUBLIC", "CONFIDENTIAL", "SECRET", "TOP_SECRET"})
	})
	Attribute("validation_rules", ValidationRules, "Validation rules for attribute")
	Attribute("encryption_required", Boolean, "Whether attribute requires encryption", func() {
		Default(false)
	})
	Attribute("is_active", Boolean, "Whether attribute definition is active", func() {
		Default(true)
	})
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Required("tenant_id", "name", "data_type", "category")
})

var Policy = Type("Policy", func() {
	Description("ABAC policy with rule-based evaluation")
	Attribute("id", String, "Unique policy identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("policy-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("tenant-123")
	})
	Attribute("entity_id", String, "Associated entity identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("entity-456")
	})
	Attribute("name", String, "Policy name", func() {
		MinLength(1)
		MaxLength(150)
		Example("business_hours_access_policy")
	})
	Attribute("display_name", String, "Human-readable policy name", func() {
		MaxLength(200)
		Example("Business Hours Access Policy")
	})
	Attribute("description", String, "Policy description", func() {
		MaxLength(1000)
		Example("Policy restricting access to financial data during business hours only")
	})
	Attribute("policy_type", PolicyType, "Type of policy")
	Attribute("effect", PermissionEffect, "Policy effect (ALLOW/DENY)")
	Attribute("priority", Int, "Policy priority for conflict resolution", func() {
		Minimum(1)
		Maximum(1000)
		Default(100)
		Example(100)
	})
	Attribute("category", PolicyCategory, "Policy category")
	Attribute("target", PolicyTarget, "When policy applies")
	Attribute("rule", PolicyRule, "Policy evaluation logic")
	Attribute("obligations", PolicyObligations, "Required actions when policy applies")
	Attribute("advice", Metadata, "Optional actions and recommendations")
	Attribute("is_active", Boolean, "Whether policy is active", func() {
		Default(true)
	})
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Attribute("updated_at", String, "Last update timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Attribute("created_by", String, "User who created the policy", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("creator-789")
	})
	Attribute("deleted_at", String, "Soft deletion timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Required("tenant_id", "name", "policy_type", "effect", "target", "rule")
})

// =============================================================================
// ACCESS REQUEST DEFINITIONS
// =============================================================================

var AccessRequest = Type("AccessRequest", func() {
	Description("Request for access to resources or roles")
	Attribute("id", String, "Unique request identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("request-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("tenant-123")
	})
	Attribute("requester_id", String, "User making the request", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("requester-456")
	})
	Attribute("target_user_id", String, "User for whom access is requested", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("target-789")
	})
	Attribute("entity_id", String, "Entity scope identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("entity-012")
	})
	Attribute("request_type", RequestType, "Type of access request")
	Attribute("role_id", String, "Role identifier for role assignments", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("role-345")
	})
	Attribute("permission_id", String, "Permission identifier for permission grants", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("permission-678")
	})
	Attribute("resource_id", String, "Resource identifier for resource access", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("resource-901")
	})
	Attribute("justification", String, "Justification for the request", func() {
		MinLength(10)
		MaxLength(1000)
		Example("Need access to employee records to complete quarterly review process")
	})
	Attribute("business_reason", String, "Business reason for access", func() {
		MaxLength(500)
		Example("Quarterly performance review cycle")
	})
	Attribute("duration_hours", Int, "Duration of temporary access in hours", func() {
		Minimum(1)
		Maximum(8760) // 1 year
		Example(72)   // 3 days
	})
	Attribute("approval_status", ApprovalStatus, "Status of the request")
	Attribute("approved_by", String, "User who approved the request", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("approver-234")
	})
	Attribute("approved_at", String, "Approval timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T15:00:00Z")
	})
	Attribute("approval_comments", String, "Comments from approver", func() {
		MaxLength(500)
		Example("Approved for quarterly review period only")
	})
	Attribute("expires_at", String, "Request expiration timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-04T15:00:00Z")
	})
	Attribute("auto_revoke", Boolean, "Whether access should be auto-revoked", func() {
		Default(true)
	})
	Attribute("created_at", String, "Request creation timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Attribute("updated_at", String, "Last update timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Required("tenant_id", "requester_id", "entity_id", "request_type", "justification")
})

// =============================================================================
// AUDIT DEFINITIONS
// =============================================================================

var AuditLog = Type("AuditLog", func() {
	Description("Comprehensive audit log entry")
	Attribute("id", String, "Unique audit log identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("audit-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("tenant-123")
	})
	Attribute("event_type", String, "Type of event", func() {
		MinLength(1)
		MaxLength(50)
		Example("user_login")
	})
	Attribute("event_category", EventCategory, "Category of event")
	Attribute("severity", EventSeverity, "Event severity level")
	Attribute("user_id", String, "User who performed the action", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("user-456")
	})
	Attribute("target_user_id", String, "Target user of the action", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("target-789")
	})
	Attribute("entity_id", String, "Entity scope identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("entity-012")
	})
	Attribute("resource_id", String, "Resource identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("resource-345")
	})
	Attribute("action_id", String, "Action identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("action-678")
	})
	Attribute("role_id", String, "Role identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("role-901")
	})
	Attribute("permission_id", String, "Permission identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("permission-234")
	})
	Attribute("decision", String, "Access decision (ALLOW/DENY)", func() {
		Enum("ALLOW", "DENY")
		Example("ALLOW")
	})
	Attribute("reason", String, "Reason for the decision", func() {
		MaxLength(500)
		Example("User has required permissions for this resource")
	})
	Attribute("risk_score", Int, "Risk score (0-100)", func() {
		Minimum(0)
		Maximum(100)
		Default(0)
		Example(25)
	})
	Attribute("context", Metadata, "Additional context information")
	Attribute("ip_address", String, "Client IP address", func() {
		Pattern("^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}$|^(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$")
		Example("192.168.1.100")
	})
	Attribute("user_agent", String, "Client user agent", func() {
		MaxLength(500)
		Example("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	})
	Attribute("session_id", String, "Session identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("session-567")
	})
	Attribute("compliance_flags", Metadata, "Compliance-related flags")
	Attribute("created_at", String, "Event timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-01T10:00:00Z")
	})
	Required("tenant_id", "event_type", "event_category", "severity")
})

// =============================================================================
// REQUEST/RESPONSE PAYLOADS
// =============================================================================

var PersonRequest = Type("PersonRequest", func() {
	Description("Request payload for creating or updating a person")
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
	Attribute("middle_name", String, "Middle name", func() {
		MaxLength(100)
		Example("Michael")
	})
	Attribute("email", String, "Email address", func() {
		Format(FormatEmail)
		MaxLength(255)
		Example("john.doe@company.com")
	})
	Attribute("phone", String, "Phone number", func() {
		Pattern("^\\+?[1-9]\\d{1,14}$")
		Example("+1-555-123-4567")
	})
	Attribute("person_type", PersonType, "Type of person")
	Attribute("birth_date", String, "Date of birth", func() {
		Format(FormatDate)
		Example("1990-01-01")
	})
	Attribute("national_id", String, "National identification number", func() {
		MaxLength(50)
		Example("123-45-6789")
	})
	Attribute("tax_id", String, "Tax identification number", func() {
		MaxLength(50)
		Example("12-3456789")
	})
	Attribute("address", Address, "Physical address")
	Attribute("security_attributes", SecurityAttributes, "ABAC security attributes")
	Attribute("metadata", Metadata, "Additional metadata")
	Required("first_name", "last_name", "person_type")
})

var EmployeeRequest = Type("EmployeeRequest", func() {
	Description("Request payload for creating or updating an employee")
	Attribute("person_id", String, "Associated person identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("person-456")
	})
	Attribute("employee_number", String, "Unique employee number", func() {
		MinLength(1)
		MaxLength(50)
		Example("EMP-001234")
	})
	Attribute("position_title", String, "Job position title", func() {
		MaxLength(100)
		Example("Senior Software Engineer")
	})
	Attribute("department_id", String, "Department entity identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("dept-123")
	})
	Attribute("manager_id", String, "Manager employee identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("mgr-456")
	})
	Attribute("hire_date", String, "Date of hire", func() {
		Format(FormatDate)
		Example("2024-01-01")
	})
	Attribute("employment_status", EmploymentStatus, "Current employment status")
	Attribute("salary_info", SalaryInfo, "Salary information")
	Attribute("work_schedule", WorkSchedule, "Work schedule details")
	Attribute("security_level", Int, "Security clearance level (0=lowest)", func() {
		Minimum(0)
		Maximum(10)
		Default(0)
		Example(3)
	})
	Attribute("access_attributes", SecurityAttributes, "ABAC access attributes")
	Required("person_id", "employee_number", "hire_date", "employment_status")
})

var UserRequest = Type("UserRequest", func() {
	Description("Request payload for creating or updating a user")
	Attribute("username", String, "Unique username", func() {
		MinLength(3)
		MaxLength(100)
		Pattern("^[a-zA-Z0-9._-]+$")
		Example("john.doe")
	})
	Attribute("email", String, "Email address", func() {
		Format(FormatEmail)
		MaxLength(255)
		Example("john.doe@company.com")
	})
	Attribute("user_type", UserType, "Type of user account")
	Attribute("person_id", String, "Associated person identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("person-789")
	})
	Attribute("employee_id", String, "Associated employee identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("emp-012")
	})
	Attribute("session_timeout_minutes", Int, "Session timeout in minutes", func() {
		Minimum(5)
		Maximum(1440) // 24 hours
		Default(480)  // 8 hours
		Example(480)
	})
	Attribute("user_attributes", SecurityAttributes, "ABAC user attributes")
	Attribute("settings", Metadata, "User preferences and settings")
	Required("email", "user_type")
})

var RoleRequest = Type("RoleRequest", func() {
	Description("Request payload for creating or updating a role")
	Attribute("name", String, "Role name", func() {
		MinLength(1)
		MaxLength(50)
		Pattern("^[A-Z_]+$")
		Example("HR_MANAGER")
	})
	Attribute("display_name", String, "Human-readable role name", func() {
		MaxLength(100)
		Example("HR Manager")
	})
	Attribute("description", String, "Role description", func() {
		MaxLength(500)
		Example("Role for managing HR processes and employee data")
	})
	Attribute("module_id", String, "Associated module identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("module-789")
	})
	Attribute("role_type", RoleType, "Type of role")
	Attribute("parent_role_id", String, "Parent role identifier for hierarchy", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("parent-role-123")
	})
	Attribute("entity_scope", Metadata, "Entity access rules")
	Attribute("conditions", Metadata, "Conditional access rules")
	Required("name", "role_type")
})

var PermissionEvaluationRequest = Type("PermissionEvaluationRequest", func() {
	Description("Request for permission evaluation")
	Attribute("user_id", String, "User identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("user-456")
	})
	Attribute("resource_name", String, "Resource name", func() {
		MinLength(1)
		MaxLength(100)
		Example("employee.records")
	})
	Attribute("action_name", String, "Action name", func() {
		MinLength(1)
		MaxLength(100)
		Example("read")
	})
	Attribute("entity_id", String, "Entity scope identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("entity-789")
	})
	Attribute("context", Metadata, "Additional evaluation context")
	Required("user_id", "resource_name", "action_name")
})

var AccessRequestRequest = Type("AccessRequestRequest", func() {
	Description("Request for creating an access request")
	Attribute("target_user_id", String, "User for whom access is requested", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("target-789")
	})
	Attribute("request_type", RequestType, "Type of access request")
	Attribute("role_id", String, "Role identifier for role assignments", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("role-345")
	})
	Attribute("permission_id", String, "Permission identifier for permission grants", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("permission-678")
	})
	Attribute("resource_id", String, "Resource identifier for resource access", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("resource-901")
	})
	Attribute("justification", String, "Justification for the request", func() {
		MinLength(10)
		MaxLength(1000)
		Example("Need access to employee records to complete quarterly review process")
	})
	Attribute("business_reason", String, "Business reason for access", func() {
		MaxLength(500)
		Example("Quarterly performance review cycle")
	})
	Attribute("duration_hours", Int, "Duration of temporary access in hours", func() {
		Minimum(1)
		Maximum(8760) // 1 year
		Example(72)   // 3 days
	})
	Required("request_type", "justification")
})

var LoginRequest = Type("LoginRequest", func() {
	Description("User login request")
	Attribute("username", String, "Username or email", func() {
		MinLength(1)
		MaxLength(255)
		Example("john.doe@company.com")
	})
	Attribute("password", String, "User password", func() {
		MinLength(8)
		MaxLength(255)
		Example("SecurePassword123!")
	})
	Attribute("device_info", DeviceInfo, "Device information")
	Attribute("location_info", LocationInfo, "Location information")
	Attribute("remember_me", Boolean, "Whether to remember the session", func() {
		Default(false)
	})
	Required("username", "password")
})

// =============================================================================
// RESPONSE PAYLOADS
// =============================================================================

var PermissionEvaluationResponse = Type("PermissionEvaluationResponse", func() {
	Description("Response from permission evaluation")
	Attribute("allowed", Boolean, "Whether access is allowed", func() {
		Example(true)
	})
	Attribute("policy_decisions", ArrayOf(String), "Policies that contributed to the decision", func() {
		Example([]string{"business_hours_policy", "department_access_policy"})
	})
	Attribute("effective_roles", ArrayOf(String), "Roles that were evaluated", func() {
		Example([]string{"HR_MANAGER", "EMPLOYEE"})
	})
	Attribute("evaluation_time_ms", Int, "Evaluation time in milliseconds", func() {
		Minimum(0)
		Example(15)
	})
	Attribute("cache_hit", Boolean, "Whether result came from cache", func() {
		Example(false)
	})
	Attribute("data_filters", DataFilters, "Applied data filters")
	Attribute("field_restrictions", FieldRestrictions, "Applied field restrictions")
	Required("allowed", "evaluation_time_ms", "cache_hit")
})

var LoginResponse = Type("LoginResponse", func() {
	Description("Response from successful login")
	Attribute("access_token", String, "JWT access token", func() {
		MinLength(50)
		Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
	})
	Attribute("refresh_token", String, "Refresh token", func() {
		MinLength(50)
		Example("refresh_eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
	})
	Attribute("token_type", String, "Token type", func() {
		Enum("Bearer")
		Example("Bearer")
	})
	Attribute("expires_in", Int, "Token expiration in seconds", func() {
		Minimum(300)  // 5 minutes
		Maximum(3600) // 1 hour
		Example(3600)
	})
	Attribute("user", User, "User information")
	Attribute("session", UserSession, "Session information")
	Attribute("requires_mfa", Boolean, "Whether MFA is required", func() {
		Default(false)
	})
	Required("access_token", "token_type", "expires_in", "user")
})

var UserBehaviorAnalytics = Type("UserBehaviorAnalytics", func() {
	Description("User behavior analytics and patterns")
	Attribute("user_id", String, "User identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("user-456")
	})
	Attribute("analysis_period", Metadata, "Analysis time period")
	Attribute("activity_patterns", Metadata, "User activity patterns")
	Attribute("productivity_metrics", Metadata, "Productivity measurements")
	Attribute("security_profile", Metadata, "Security-related metrics")
	Attribute("risk_assessment", Metadata, "Risk analysis results")
	Required("user_id", "analysis_period")
})

// =============================================================================
// COMMON RESPONSE TYPES
// =============================================================================

var ErrorResponse = Type("ErrorResponse", func() {
	Description("Error response")
	Attribute("code", String, "Error code", func() {
		Example("VALIDATION_ERROR")
	})
	Attribute("message", String, "Error message", func() {
		Example("Invalid request parameters")
	})
	Attribute("details", ArrayOf(String), "Detailed error information", func() {
		Example([]string{"field 'email' is required", "field 'name' must be at least 1 character"})
	})
	Attribute("request_id", String, "Request identifier for tracking", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("req-123e4567-e89b-12d3-a456-426614174000")
	})
	Required("code", "message")
})

var SuccessResponse = Type("SuccessResponse", func() {
	Description("Generic success response")
	Attribute("message", String, "Success message", func() {
		Example("Operation completed successfully")
	})
	Attribute("request_id", String, "Request identifier for tracking", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("req-123e4567-e89b-12d3-a456-426614174000")
	})
	Required("message")
})

var PaginationResponse = Type("PaginationResponse", func() {
	Description("Pagination information")
	Attribute("page", Int, "Current page number", func() {
		Minimum(1)
		Example(1)
	})
	Attribute("per_page", Int, "Items per page", func() {
		Minimum(1)
		Maximum(100)
		Example(20)
	})
	Attribute("total_pages", Int, "Total number of pages", func() {
		Minimum(0)
		Example(5)
	})
	Attribute("total_items", Int, "Total number of items", func() {
		Minimum(0)
		Example(98)
	})
	Attribute("has_next", Boolean, "Whether there are more pages", func() {
		Example(true)
	})
	Attribute("has_prev", Boolean, "Whether there are previous pages", func() {
		Example(false)
	})
	Required("page", "per_page", "total_pages", "total_items", "has_next", "has_prev")
})

// =============================================================================
// LIST RESPONSE TYPES
// =============================================================================

var PersonListResponse = Type("PersonListResponse", func() {
	Description("List of persons with pagination")
	Attribute("data", ArrayOf(Person), "List of persons")
	Attribute("pagination", PaginationResponse, "Pagination information")
	Required("data", "pagination")
})

var EmployeeListResponse = Type("EmployeeListResponse", func() {
	Description("List of employees with pagination")
	Attribute("data", ArrayOf(Employee), "List of employees")
	Attribute("pagination", PaginationResponse, "Pagination information")
	Required("data", "pagination")
})

var UserListResponse = Type("UserListResponse", func() {
	Description("List of users with pagination")
	Attribute("data", ArrayOf(User), "List of users")
	Attribute("pagination", PaginationResponse, "Pagination information")
	Required("data", "pagination")
})

var RoleListResponse = Type("RoleListResponse", func() {
	Description("List of roles with pagination")
	Attribute("data", ArrayOf(Role), "List of roles")
	Attribute("pagination", PaginationResponse, "Pagination information")
	Required("data", "pagination")
})

var AccessRequestListResponse = Type("AccessRequestListResponse", func() {
	Description("List of access requests with pagination")
	Attribute("data", ArrayOf(AccessRequest), "List of access requests")
	Attribute("pagination", PaginationResponse, "Pagination information")
	Required("data", "pagination")
})

var AuditLogListResponse = Type("AuditLogListResponse", func() {
	Description("List of audit logs with pagination")
	Attribute("data", ArrayOf(AuditLog), "List of audit logs")
	Attribute("pagination", PaginationResponse, "Pagination information")
	Required("data", "pagination")
})

// =============================================================================
// SERVICE DEFINITION
// =============================================================================

var _ = Service("user-management", func() {
	Description("User Management & Permissions Service - Enterprise-grade identity and access management")

	Security(JWT, APIKey)

	// =============================================================================
	// PERSON MANAGEMENT ENDPOINTS
	// =============================================================================

	Method("list_persons", func() {
		Description("List all persons with filtering and pagination")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("page", Int, "Page number", func() {
				Default(1)
				Minimum(1)
				Example(1)
			})
			Attribute("per_page", Int, "Items per page", func() {
				Default(20)
				Minimum(1)
				Maximum(100)
				Example(20)
			})
			Attribute("person_type", PersonType, "Filter by person type")
			Attribute("is_active", Boolean, "Filter by active status")
			Attribute("search", String, "Search in name and email", func() {
				MaxLength(255)
				Example("john")
			})
		})
		Result(PersonListResponse)
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/persons")
			Param("page")
			Param("per_page")
			Param("person_type")
			Param("is_active")
			Param("search")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("internal", CodeInternal)
		})
	})

	Method("create_person", func() {
		Description("Create a new person entity")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("person", PersonRequest, "Person data", func() {
				Required("person")
			})
		})
		Result(Person)
		Error("bad_request", ErrorResponse, "Invalid request data")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("conflict", ErrorResponse, "Person already exists")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			POST("/api/v1/persons")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("conflict", StatusConflict)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("internal", CodeInternal)
		})
	})

	Method("get_person", func() {
		Description("Get a person by ID")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("person_id", String, "Person identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Required("person_id")
		})
		Result(Person)
		Error("bad_request", ErrorResponse, "Invalid person ID")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "Person not found")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/persons/{person_id}")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("internal", CodeInternal)
		})
	})

	Method("update_person", func() {
		Description("Update an existing person")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("person_id", String, "Person identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("person", PersonRequest, "Updated person data")
			Required("person_id", "person")
		})
		Result(Person)
		Error("bad_request", ErrorResponse, "Invalid request data")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "Person not found")
		Error("conflict", ErrorResponse, "Update conflict")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			PUT("/api/v1/persons/{person_id}")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("conflict", StatusConflict)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("conflict", CodeAborted)
			Response("internal", CodeInternal)
		})
	})

	Method("delete_person", func() {
		Description("Soft delete a person (mark as deleted)")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("person_id", String, "Person identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Required("person_id")
		})
		Result(SuccessResponse)
		Error("bad_request", ErrorResponse, "Invalid person ID")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "Person not found")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			DELETE("/api/v1/persons/{person_id}")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("internal", CodeInternal)
		})
	})

	// =============================================================================
	// EMPLOYEE MANAGEMENT ENDPOINTS
	// =============================================================================

	Method("list_employees", func() {
		Description("List all employees with filtering and pagination")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("page", Int, "Page number", func() {
				Default(1)
				Minimum(1)
				Example(1)
			})
			Attribute("per_page", Int, "Items per page", func() {
				Default(20)
				Minimum(1)
				Maximum(100)
				Example(20)
			})
			Attribute("employment_status", EmploymentStatus, "Filter by employment status")
			Attribute("department_id", String, "Filter by department", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("dept-123")
			})
			Attribute("manager_id", String, "Filter by manager", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("mgr-456")
			})
			Attribute("security_level_min", Int, "Minimum security level", func() {
				Minimum(0)
				Maximum(10)
				Example(2)
			})
			Attribute("search", String, "Search in employee data", func() {
				MaxLength(255)
				Example("john")
			})
		})
		Result(EmployeeListResponse)
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/employees")
			Param("page")
			Param("per_page")
			Param("employment_status")
			Param("department_id")
			Param("manager_id")
			Param("security_level_min")
			Param("search")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("internal", CodeInternal)
		})
	})

	Method("create_employee", func() {
		Description("Create a new employee record")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("employee", EmployeeRequest, "Employee data")
			Required("employee")
		})
		Result(Employee)
		Error("bad_request", ErrorResponse, "Invalid request data")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("conflict", ErrorResponse, "Employee already exists")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			POST("/api/v1/employees")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("conflict", StatusConflict)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("conflict", CodeAlreadyExists)
			Response("internal", CodeInternal)
		})
	})

	Method("get_employee", func() {
		Description("Get an employee by ID")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("employee_id", String, "Employee identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("emp-123e4567-e89b-12d3-a456-426614174000")
			})
			Required("employee_id")
		})
		Result(Employee)
		Error("bad_request", ErrorResponse, "Invalid employee ID")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "Employee not found")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/employees/{employee_id}")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("internal", CodeInternal)
		})
	})

	Method("update_employee", func() {
		Description("Update an existing employee")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("employee_id", String, "Employee identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("emp-123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("employee", EmployeeRequest, "Updated employee data")
			Required("employee_id", "employee")
		})
		Result(Employee)
		Error("bad_request", ErrorResponse, "Invalid request data")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "Employee not found")
		Error("conflict", ErrorResponse, "Update conflict")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			PUT("/api/v1/employees/{employee_id}")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("conflict", StatusConflict)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("conflict", CodeAborted)
			Response("internal", CodeInternal)
		})
	})

	Method("get_employee_subordinates", func() {
		Description("Get all subordinates of an employee (organizational hierarchy)")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("employee_id", String, "Employee identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("emp-123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("levels", Int, "Number of hierarchy levels to include", func() {
				Minimum(1)
				Maximum(10)
				Default(1)
				Example(2)
			})
			Required("employee_id")
		})
		Result(EmployeeListResponse)
		Error("bad_request", ErrorResponse, "Invalid employee ID")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "Employee not found")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/employees/{employee_id}/subordinates")
			Param("levels")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("internal", CodeInternal)
		})
	})

	// =============================================================================
	// USER MANAGEMENT ENDPOINTS
	// =============================================================================

	Method("list_users", func() {
		Description("List all users with filtering and pagination")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("page", Int, "Page number", func() {
				Default(1)
				Minimum(1)
				Example(1)
			})
			Attribute("per_page", Int, "Items per page", func() {
				Default(20)
				Minimum(1)
				Maximum(100)
				Example(20)
			})
			Attribute("user_type", UserType, "Filter by user type")
			Attribute("account_status", AccountStatus, "Filter by account status")
			Attribute("is_active", Boolean, "Filter by active status")
			Attribute("mfa_enabled", Boolean, "Filter by MFA status")
			Attribute("search", String, "Search in username and email", func() {
				MaxLength(255)
				Example("john")
			})
		})
		Result(UserListResponse)
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/users")
			Param("page")
			Param("per_page")
			Param("user_type")
			Param("account_status")
			Param("is_active")
			Param("mfa_enabled")
			Param("search")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("internal", CodeInternal)
		})
	})

	Method("create_user", func() {
		Description("Create a new user account")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("user", UserRequest, "User data")
			Required("user")
		})
		Result(User)
		Error("bad_request", ErrorResponse, "Invalid request data")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("conflict", ErrorResponse, "User already exists")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			POST("/api/v1/users")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("conflict", StatusConflict)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("conflict", CodeAlreadyExists)
			Response("internal", CodeInternal)
		})
	})

	Method("get_user", func() {
		Description("Get a user by ID")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("user_id", String, "User identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("user-123e4567-e89b-12d3-a456-426614174000")
			})
			Required("user_id")
		})
		Result(User)
		Error("bad_request", ErrorResponse, "Invalid user ID")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "User not found")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/users/{user_id}")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("internal", CodeInternal)
		})
	})

	Method("update_user", func() {
		Description("Update an existing user")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("user_id", String, "User identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("user-123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("user", UserRequest, "Updated user data")
			Required("user_id", "user")
		})
		Result(User)
		Error("bad_request", ErrorResponse, "Invalid request data")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "User not found")
		Error("conflict", ErrorResponse, "Update conflict")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			PUT("/api/v1/users/{user_id}")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("conflict", StatusConflict)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("conflict", CodeAborted)
			Response("internal", CodeInternal)
		})
	})

	Method("deactivate_user", func() {
		Description("Deactivate a user account (soft delete)")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("user_id", String, "User identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("user-123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("reason", String, "Reason for deactivation", func() {
				MinLength(1)
				MaxLength(500)
				Example("Employee terminated")
			})
			Required("user_id")
		})
		Result(SuccessResponse)
		Error("bad_request", ErrorResponse, "Invalid user ID")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "User not found")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			DELETE("/api/v1/users/{user_id}")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("internal", CodeInternal)
		})
	})

	// =============================================================================
	// ROLE MANAGEMENT ENDPOINTS
	// =============================================================================

	Method("list_roles", func() {
		Description("List all roles with filtering and pagination")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("page", Int, "Page number", func() {
				Default(1)
				Minimum(1)
				Example(1)
			})
			Attribute("per_page", Int, "Items per page", func() {
				Default(20)
				Minimum(1)
				Maximum(100)
				Example(20)
			})
			Attribute("role_type", RoleType, "Filter by role type")
			Attribute("module_id", String, "Filter by module", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("module-123")
			})
			Attribute("is_active", Boolean, "Filter by active status")
			Attribute("is_system_role", Boolean, "Filter by system role status")
			Attribute("search", String, "Search in role name and description", func() {
				MaxLength(255)
				Example("manager")
			})
		})
		Result(RoleListResponse)
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/roles")
			Param("page")
			Param("per_page")
			Param("role_type")
			Param("module_id")
			Param("is_active")
			Param("is_system_role")
			Param("search")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("internal", CodeInternal)
		})
	})

	Method("create_role", func() {
		Description("Create a new role")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("role", RoleRequest, "Role data")
			Required("role")
		})
		Result(Role)
		Error("bad_request", ErrorResponse, "Invalid request data")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("conflict", ErrorResponse, "Role already exists")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			POST("/api/v1/roles")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("conflict", StatusConflict)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("conflict", CodeAlreadyExists)
			Response("internal", CodeInternal)
		})
	})

	Method("get_role", func() {
		Description("Get a role by ID")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("role_id", String, "Role identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("role-123e4567-e89b-12d3-a456-426614174000")
			})
			Required("role_id")
		})
		Result(Role)
		Error("bad_request", ErrorResponse, "Invalid role ID")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "Role not found")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/roles/{role_id}")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("internal", CodeInternal)
		})
	})

	Method("update_role", func() {
		Description("Update an existing role")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("role_id", String, "Role identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("role-123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("role", RoleRequest, "Updated role data")
			Required("role_id", "role")
		})
		Result(Role)
		Error("bad_request", ErrorResponse, "Invalid request data")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "Role not found")
		Error("conflict", ErrorResponse, "Update conflict")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			PUT("/api/v1/roles/{role_id}")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("conflict", StatusConflict)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("conflict", CodeAborted)
			Response("internal", CodeInternal)
		})
	})

	Method("get_role_inheritance", func() {
		Description("Get role inheritance hierarchy")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("role_id", String, "Role identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("role-123e4567-e89b-12d3-a456-426614174000")
			})
			Required("role_id")
		})
		Result(RoleListResponse)
		Error("bad_request", ErrorResponse, "Invalid role ID")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "Role not found")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/roles/{role_id}/inheritance")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("internal", CodeInternal)
		})
	})

	// =============================================================================
	// PERMISSION EVALUATION ENDPOINTS
	// =============================================================================

	Method("evaluate_permission", func() {
		Description("Evaluate permission for user, resource, and action")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("evaluation", PermissionEvaluationRequest, "Permission evaluation request")
			Required("evaluation")
		})
		Result(PermissionEvaluationResponse)
		Error("bad_request", ErrorResponse, "Invalid evaluation request")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			POST("/api/v1/permissions/evaluate")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("internal", CodeInternal)
		})
	})

	Method("get_user_effective_permissions", func() {
		Description("Get all effective permissions for a user")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("user_id", String, "User identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("user-123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("entity_id", String, "Entity scope identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("entity-456")
			})
			Required("user_id")
		})
		Result(Metadata)
		Error("bad_request", ErrorResponse, "Invalid user ID")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "User not found")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/users/{user_id}/effective-permissions")
			Param("entity_id")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("internal", CodeInternal)
		})
	})

	Method("bulk_evaluate_permissions", func() {
		Description("Evaluate multiple permissions in a single request")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("evaluations", ArrayOf(PermissionEvaluationRequest), "List of permission evaluations", func() {
				MinLength(1)
				MaxLength(100)
			})
			Required("evaluations")
		})
		Result(ArrayOf(PermissionEvaluationResponse))
		Error("bad_request", ErrorResponse, "Invalid evaluation requests")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			POST("/api/v1/permissions/bulk-evaluate")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("internal", CodeInternal)
		})
	})

	// =============================================================================
	// ABAC POLICY MANAGEMENT ENDPOINTS
	// =============================================================================

	Method("list_policies", func() {
		Description("List all ABAC policies with filtering and pagination")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("page", Int, "Page number", func() {
				Default(1)
				Minimum(1)
				Example(1)
			})
			Attribute("per_page", Int, "Items per page", func() {
				Default(20)
				Minimum(1)
				Maximum(100)
				Example(20)
			})
			Attribute("policy_type", PolicyType, "Filter by policy type")
			Attribute("category", PolicyCategory, "Filter by category")
			Attribute("effect", PermissionEffect, "Filter by effect")
			Attribute("is_active", Boolean, "Filter by active status")
			Attribute("search", String, "Search in policy name and description", func() {
				MaxLength(255)
				Example("business hours")
			})
		})
		Result(func() {
			Attribute("data", ArrayOf(Policy), "List of policies")
			Attribute("pagination", PaginationResponse, "Pagination information")
			Required("data", "pagination")
		})
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/policies")
			Param("page")
			Param("per_page")
			Param("policy_type")
			Param("category")
			Param("effect")
			Param("is_active")
			Param("search")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("internal", CodeInternal)
		})
	})

	Method("create_policy", func() {
		Description("Create a new ABAC policy")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("policy", func() {
				Attribute("name", String, "Policy name", func() {
					MinLength(1)
					MaxLength(150)
					Example("business_hours_access_policy")
				})
				Attribute("display_name", String, "Human-readable policy name", func() {
					MaxLength(200)
					Example("Business Hours Access Policy")
				})
				Attribute("description", String, "Policy description", func() {
					MaxLength(1000)
					Example("Policy restricting access to financial data during business hours only")
				})
				Attribute("policy_type", PolicyType, "Type of policy")
				Attribute("effect", PermissionEffect, "Policy effect (ALLOW/DENY)")
				Attribute("priority", Int, "Policy priority for conflict resolution", func() {
					Minimum(1)
					Maximum(1000)
					Default(100)
					Example(100)
				})
				Attribute("category", PolicyCategory, "Policy category")
				Attribute("target", PolicyTarget, "When policy applies")
				Attribute("rule", PolicyRule, "Policy evaluation logic")
				Attribute("obligations", PolicyObligations, "Required actions when policy applies")
				Attribute("advice", Metadata, "Optional actions and recommendations")
				Required("name", "policy_type", "effect", "target", "rule")
			})
			Required("policy")
		})
		Result(Policy)
		Error("bad_request", ErrorResponse, "Invalid request data")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("conflict", ErrorResponse, "Policy already exists")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			POST("/api/v1/policies")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("conflict", StatusConflict)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("conflict", CodeAlreadyExists)
			Response("internal", CodeInternal)
		})
	})

	Method("test_policy", func() {
		Description("Test a policy against sample scenarios")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("policy_id", String, "Policy identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("policy-123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("test_scenarios", ArrayOf(PermissionEvaluationRequest), "Test scenarios", func() {
				MinLength(1)
				MaxLength(10)
			})
			Required("policy_id", "test_scenarios")
		})
		Result(func() {
			Attribute("policy_id", String, "Policy identifier")
			Attribute("test_results", ArrayOf(func() {
				Attribute("scenario", PermissionEvaluationRequest, "Test scenario")
				Attribute("result", PermissionEvaluationResponse, "Evaluation result")
				Attribute("policy_matched", Boolean, "Whether policy matched")
				Required("scenario", "result", "policy_matched")
			}), "Test results")
			Required("policy_id", "test_results")
		})
		Error("bad_request", ErrorResponse, "Invalid test request")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "Policy not found")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			POST("/api/v1/policies/test")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("internal", CodeInternal)
		})
	})

	Method("get_policy_impact", func() {
		Description("Analyze the impact of a policy on users and resources")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("policy_id", String, "Policy identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("policy-123e4567-e89b-12d3-a456-426614174000")
			})
			Required("policy_id")
		})
		Result(func() {
			Attribute("policy_id", String, "Policy identifier")
			Attribute("affected_users_count", Int, "Number of users affected")
			Attribute("affected_resources_count", Int, "Number of resources affected")
			Attribute("access_grants", Int, "Number of access grants")
			Attribute("access_denials", Int, "Number of access denials")
			Attribute("conflict_policies", ArrayOf(String), "Conflicting policy IDs")
			Required("policy_id", "affected_users_count", "affected_resources_count")
		})
		Error("bad_request", ErrorResponse, "Invalid policy ID")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "Policy not found")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/policies/{policy_id}/impact")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("internal", CodeInternal)
		})
	})

	// =============================================================================
	// ACCESS REQUEST WORKFLOW ENDPOINTS
	// =============================================================================

	Method("submit_access_request", func() {
		Description("Submit a new access request")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("request", AccessRequestRequest, "Access request data")
			Required("request")
		})
		Result(AccessRequest)
		Error("bad_request", ErrorResponse, "Invalid request data")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			POST("/api/v1/access-requests")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("internal", CodeInternal)
		})
	})

	Method("list_access_requests", func() {
		Description("List access requests with filtering and pagination")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("page", Int, "Page number", func() {
				Default(1)
				Minimum(1)
				Example(1)
			})
			Attribute("per_page", Int, "Items per page", func() {
				Default(20)
				Minimum(1)
				Maximum(100)
				Example(20)
			})
			Attribute("request_type", RequestType, "Filter by request type")
			Attribute("approval_status", ApprovalStatus, "Filter by approval status")
			Attribute("requester_id", String, "Filter by requester", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("requester-123")
			})
			Attribute("target_user_id", String, "Filter by target user", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("target-456")
			})
		})
		Result(AccessRequestListResponse)
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/access-requests")
			Param("page")
			Param("per_page")
			Param("request_type")
			Param("approval_status")
			Param("requester_id")
			Param("target_user_id")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("internal", CodeInternal)
		})
	})

	Method("approve_access_request", func() {
		Description("Approve an access request")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("request_id", String, "Access request identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("request-123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("comments", String, "Approval comments", func() {
				MaxLength(500)
				Example("Approved for quarterly review period")
			})
			Attribute("expires_at", String, "Override expiration time", func() {
				Format(FormatDateTime)
				Example("2024-01-04T15:00:00Z")
			})
			Required("request_id")
		})
		Result(AccessRequest)
		Error("bad_request", ErrorResponse, "Invalid request ID")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "Request not found")
		Error("conflict", ErrorResponse, "Request already processed")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			POST("/api/v1/access-requests/{request_id}/approve")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("conflict", StatusConflict)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("conflict", CodeAborted)
			Response("internal", CodeInternal)
		})
	})

	Method("reject_access_request", func() {
		Description("Reject an access request")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("request_id", String, "Access request identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("request-123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("comments", String, "Rejection reason", func() {
				MinLength(1)
				MaxLength(500)
				Example("Request does not meet security requirements")
			})
			Required("request_id", "comments")
		})
		Result(AccessRequest)
		Error("bad_request", ErrorResponse, "Invalid request data")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "Request not found")
		Error("conflict", ErrorResponse, "Request already processed")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			POST("/api/v1/access-requests/{request_id}/reject")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("conflict", StatusConflict)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("conflict", CodeAborted)
			Response("internal", CodeInternal)
		})
	})

	// =============================================================================
	// AUTHENTICATION ENDPOINTS
	// =============================================================================

	Method("login", func() {
		Description("Authenticate user and create session")
		Payload(func() {
			CommonHeaders()
			Attribute("credentials", LoginRequest, "Login credentials")
			Required("credentials")
		})
		Result(LoginResponse)
		Error("bad_request", ErrorResponse, "Invalid credentials format")
		Error("unauthorized", ErrorResponse, "Invalid credentials")
		Error("forbidden", ErrorResponse, "Account locked or disabled")
		Error("mfa_required", func() {
			Attribute("mfa_token", String, "MFA challenge token")
			Attribute("mfa_methods", ArrayOf(String), "Available MFA methods")
			Required("mfa_token", "mfa_methods")
		}, "Multi-factor authentication required")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			POST("/api/v1/auth/login")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("mfa_required", StatusPreconditionRequired)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("mfa_required", CodeFailedPrecondition)
			Response("internal", CodeInternal)
		})
	})

	Method("step_up_authentication", func() {
		Description("Perform step-up authentication for sensitive operations")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("password", String, "Current password", func() {
				MinLength(8)
				MaxLength(255)
				Example("CurrentPassword123!")
			})
			Attribute("mfa_code", String, "MFA verification code", func() {
				Pattern("^[0-9]{6}$")
				Example("123456")
			})
			Attribute("operation", String, "Operation requiring step-up", func() {
				MinLength(1)
				MaxLength(100)
				Example("delete_user")
			})
			Required("operation")
		})
		Result(func() {
			Attribute("step_up_token", String, "Step-up authentication token")
			Attribute("expires_at", String, "Token expiration time")
			Required("step_up_token", "expires_at")
		})
		Error("bad_request", ErrorResponse, "Invalid step-up request")
		Error("unauthorized", ErrorResponse, "Invalid credentials")
		Error("forbidden", ErrorResponse, "Step-up not allowed")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			POST("/api/v1/auth/step-up")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("internal", CodeInternal)
		})
	})

	Method("assess_login_risk", func() {
		Description("Assess login risk based on context and behavior")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("user_id", String, "User identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("user-123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("device_info", DeviceInfo, "Device information")
			Attribute("location_info", LocationInfo, "Location information")
			Attribute("behavioral_context", Metadata, "Additional behavioral data")
			Required("user_id")
		})
		Result(func() {
			Attribute("risk_score", Int, "Risk score (0-100)", func() {
				Minimum(0)
				Maximum(100)
				Example(25)
			})
			Attribute("risk_level", RiskLevel, "Risk level category")
			Attribute("risk_factors", ArrayOf(String), "Identified risk factors", func() {
				Example([]string{"new_device", "unusual_location"})
			})
			Attribute("recommended_actions", ArrayOf(String), "Recommended security actions", func() {
				Example([]string{"require_mfa", "verify_identity"})
			})
			Required("risk_score", "risk_level", "risk_factors", "recommended_actions")
		})
		Error("bad_request", ErrorResponse, "Invalid assessment request")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "User not found")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			POST("/api/v1/auth/risk-assessment")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("internal", CodeInternal)
		})
	})

	// =============================================================================
	// AUDIT & ANALYTICS ENDPOINTS
	// =============================================================================

	Method("query_activity_logs", func() {
		Description("Query user activity logs with advanced filtering")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("page", Int, "Page number", func() {
				Default(1)
				Minimum(1)
				Example(1)
			})
			Attribute("per_page", Int, "Items per page", func() {
				Default(20)
				Minimum(1)
				Maximum(100)
				Example(20)
			})
			Attribute("start_date", String, "Start date for query range", func() {
				Format(FormatDateTime)
				Example("2024-01-01T00:00:00Z")
			})
			Attribute("end_date", String, "End date for query range", func() {
				Format(FormatDateTime)
				Example("2024-01-31T23:59:59Z")
			})
			Attribute("user_id", String, "Filter by user", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("user-123")
			})
			Attribute("event_category", EventCategory, "Filter by event category")
			Attribute("severity", EventSeverity, "Filter by severity")
			Attribute("resource_id", String, "Filter by resource", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("resource-456")
			})
			Attribute("action_id", String, "Filter by action", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("action-789")
			})
			Attribute("decision", String, "Filter by access decision", func() {
				Enum("ALLOW", "DENY")
				Example("ALLOW")
			})
			Attribute("risk_score_min", Int, "Minimum risk score", func() {
				Minimum(0)
				Maximum(100)
				Example(50)
			})
		})
		Result(AuditLogListResponse)
		Error("bad_request", ErrorResponse, "Invalid query parameters")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/audit/activities")
			Param("page")
			Param("per_page")
			Param("start_date")
			Param("end_date")
			Param("user_id")
			Param("event_category")
			Param("severity")
			Param("resource_id")
			Param("action_id")
			Param("decision")
			Param("risk_score_min")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("internal", CodeInternal)
		})
	})

	Method("get_security_events", func() {
		Description("Get security-related events and anomalies")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("page", Int, "Page number", func() {
				Default(1)
				Minimum(1)
				Example(1)
			})
			Attribute("per_page", Int, "Items per page", func() {
				Default(20)
				Minimum(1)
				Maximum(100)
				Example(20)
			})
			Attribute("severity", EventSeverity, "Filter by severity")
			Attribute("start_date", String, "Start date for query range", func() {
				Format(FormatDateTime)
				Example("2024-01-01T00:00:00Z")
			})
			Attribute("end_date", String, "End date for query range", func() {
				Format(FormatDateTime)
				Example("2024-01-31T23:59:59Z")
			})
			Attribute("unresolved_only", Boolean, "Show only unresolved events", func() {
				Default(false)
			})
		})
		Result(AuditLogListResponse)
		Error("bad_request", ErrorResponse, "Invalid query parameters")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/audit/security-events")
			Param("page")
			Param("per_page")
			Param("severity")
			Param("start_date")
			Param("end_date")
			Param("unresolved_only")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("internal", CodeInternal)
		})
	})

	Method("get_user_behavior_analytics", func() {
		Description("Get comprehensive user behavior analytics")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("user_id", String, "User identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("user-123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("start_date", String, "Analysis start date", func() {
				Format(FormatDateTime)
				Example("2024-01-01T00:00:00Z")
			})
			Attribute("end_date", String, "Analysis end date", func() {
				Format(FormatDateTime)
				Example("2024-01-31T23:59:59Z")
			})
			Required("user_id")
		})
		Result(UserBehaviorAnalytics)
		Error("bad_request", ErrorResponse, "Invalid analytics request")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "User not found")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/analytics/user-behavior")
			Param("user_id")
			Param("start_date")
			Param("end_date")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("internal", CodeInternal)
		})
	})

	Method("analyze_access_patterns", func() {
		Description("Analyze access patterns across users and resources")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("analysis_type", String, "Type of analysis", func() {
				Enum("USER_PATTERNS", "RESOURCE_USAGE", "ROLE_EFFECTIVENESS", "ANOMALY_DETECTION")
				Example("USER_PATTERNS")
			})
			Attribute("start_date", String, "Analysis start date", func() {
				Format(FormatDateTime)
				Example("2024-01-01T00:00:00Z")
			})
			Attribute("end_date", String, "Analysis end date", func() {
				Format(FormatDateTime)
				Example("2024-01-31T23:59:59Z")
			})
			Attribute("entity_ids", ArrayOf(String), "Limit analysis to specific entities", func() {
				Example([]string{"entity-123", "entity-456"})
			})
			Required("analysis_type")
		})
		Result(Metadata)
		Error("bad_request", ErrorResponse, "Invalid analysis request")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/analytics/access-patterns")
			Param("analysis_type")
			Param("start_date")
			Param("end_date")
			Param("entity_ids")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("internal", CodeInternal)
		})
	})

	Method("generate_risk_assessment", func() {
		Description("Generate comprehensive risk assessment for users or system")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("assessment_type", String, "Type of risk assessment", func() {
				Enum("USER_RISK", "SYSTEM_RISK", "ENTITY_RISK", "COMPLIANCE_RISK")
				Example("USER_RISK")
			})
			Attribute("target_id", String, "Target identifier (user, entity, etc.)", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("user-123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("include_recommendations", Boolean, "Include mitigation recommendations", func() {
				Default(true)
			})
			Required("assessment_type")
		})
		Result(Metadata)
		Error("bad_request", ErrorResponse, "Invalid assessment request")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "Target not found")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			POST("/api/v1/analytics/risk-assessment")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("internal", CodeInternal)
		})
	})

	// =============================================================================
	// COMPLIANCE & REPORTING ENDPOINTS
	// =============================================================================

	Method("get_gdpr_data", func() {
		Description("Get GDPR-compliant user data export")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("user_id", String, "User identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("user-123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("include_audit_logs", Boolean, "Include audit log data", func() {
				Default(false)
			})
			Required("user_id")
		})
		Result(Metadata)
		Error("bad_request", ErrorResponse, "Invalid export request")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "User not found")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/compliance/gdpr-data")
			Param("user_id")
			Param("include_audit_logs")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("internal", CodeInternal)
		})
	})

	Method("export_user_data", func() {
		Description("Export user data in various formats for compliance")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("user_ids", ArrayOf(String), "User identifiers", func() {
				MinLength(1)
				MaxLength(100)
				Example([]string{"user-123", "user-456"})
			})
			Attribute("export_format", String, "Export format", func() {
				Enum("JSON", "CSV", "XML", "PDF")
				Default("JSON")
				Example("JSON")
			})
			Attribute("include_permissions", Boolean, "Include permission data", func() {
				Default(true)
			})
			Attribute("include_audit_trail", Boolean, "Include audit trail", func() {
				Default(false)
			})
			Required("user_ids")
		})
		Result(func() {
			Attribute("export_id", String, "Export job identifier")
			Attribute("download_url", String, "Download URL for the export")
			Attribute("expires_at", String, "Export expiration time")
			Required("export_id", "download_url", "expires_at")
		})
		Error("bad_request", ErrorResponse, "Invalid export request")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			POST("/api/v1/compliance/data-export")
			Response(StatusAccepted)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("internal", CodeInternal)
		})
	})

	Method("process_right_to_be_forgotten", func() {
		Description("Process GDPR right to be forgotten request")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("user_id", String, "User identifier", func() {
				Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
				Example("user-123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("retain_audit_logs", Boolean, "Retain audit logs for compliance", func() {
				Default(true)
			})
			Attribute("anonymize_instead", Boolean, "Anonymize data instead of deletion", func() {
				Default(false)
			})
			Attribute("justification", String, "Justification for the request", func() {
				MinLength(10)
				MaxLength(1000)
				Example("User requested complete data deletion per GDPR Article 17")
			})
			Required("user_id", "justification")
		})
		Result(func() {
			Attribute("deletion_id", String, "Deletion process identifier")
			Attribute("status", String, "Current status of deletion process")
			Attribute("estimated_completion", String, "Estimated completion time")
			Required("deletion_id", "status")
		})
		Error("bad_request", ErrorResponse, "Invalid deletion request")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("not_found", ErrorResponse, "User not found")
		Error("conflict", ErrorResponse, "Deletion not allowed")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			POST("/api/v1/compliance/right-to-be-forgotten")
			Response(StatusAccepted)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("conflict", StatusConflict)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("conflict", CodeAborted)
			Response("internal", CodeInternal)
		})
	})

	Method("generate_access_review", func() {
		Description("Generate access review report for compliance audits")
		Security(JWT)
		Payload(func() {
			CommonHeaders()
			Attribute("review_type", String, "Type of access review", func() {
				Enum("USER_ACCESS", "ROLE_REVIEW", "PERMISSION_AUDIT", "COMPLIANCE_REPORT")
				Example("USER_ACCESS")
			})
			Attribute("scope", String, "Review scope", func() {
				Enum("ALL_USERS", "DEPARTMENT", "ROLE_BASED", "HIGH_PRIVILEGE")
				Example("HIGH_PRIVILEGE")
			})
			Attribute("entity_ids", ArrayOf(String), "Limit review to specific entities", func() {
				Example([]string{"entity-123", "entity-456"})
			})
			Attribute("include_inactive", Boolean, "Include inactive users/roles", func() {
				Default(false)
			})
			Attribute("report_format", String, "Report format", func() {
				Enum("PDF", "EXCEL", "CSV", "JSON")
				Default("PDF")
				Example("PDF")
			})
			Required("review_type", "scope")
		})
		Result(func() {
			Attribute("report_id", String, "Report generation identifier")
			Attribute("download_url", String, "Download URL for the report")
			Attribute("expires_at", String, "Report expiration time")
			Attribute("total_users", Int, "Total users included in review")
			Attribute("total_roles", Int, "Total roles included in review")
			Required("report_id", "download_url", "expires_at")
		})
		Error("bad_request", ErrorResponse, "Invalid review request")
		Error("unauthorized", ErrorResponse, "Authentication required")
		Error("forbidden", ErrorResponse, "Insufficient permissions")
		Error("internal", ErrorResponse, "Internal server error")

		HTTP(func() {
			GET("/api/v1/reports/access-review")
			Param("review_type")
			Param("scope")
			Param("entity_ids")
			Param("include_inactive")
			Param("report_format")
			Response(StatusAccepted)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal", StatusInternalServerError)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("internal", CodeInternal)
		})
	})
	Response("forbidden", CodePermissionDenied)
	Response("conflict", CodeAlreadyExists)
	Response("internal", CodeInternal)
})
