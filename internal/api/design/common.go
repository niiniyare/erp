package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// COMMON REGEX PATTERNS
// ============================================================================

var (
	UUIDPattern         = "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
	PhonePattern        = "^\\+?[1-9]\\d{1,14}$"
	CountryCodePattern  = "^[A-Z]{2}$"
	CurrencyCodePattern = "^[A-Z]{3}$"
	SHA256Pattern       = "^[a-f0-9]{64}$"
	LanguageCodePattern = "^[a-z]{2}(-[A-Z]{2})?$"
	APIVersionPattern   = "^v\\d+(\\.\\d+)?$"
	DecimalPattern      = "^-?\\d+(\\.\\d{1,4})?$"
	MIMETypePattern     = "^[a-z-]+/[a-z0-9][a-z0-9!#\\-\\^_]*$"
	EmailPattern        = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
	URLPattern          = "^(https?|ftp)://[^\\s/$.?#].[^\\s]*$"
	IPAddressPattern    = "^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$"
	IPv6Pattern         = "^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$"
	CIDRPattern         = "^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)/(3[0-2]|[12]?[0-9])$"
	DatePattern         = "^\\d{4}-\\d{2}-\\d{2}$"
	DateTimePattern     = "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}(?:\\.\\d+)?Z?$"
	TimePattern         = "^\\d{2}:\\d{2}:\\d{2}$"
	ZipCodePattern      = "^\\d{5}(-\\d{4})?$"
	CreditCardPattern   = "^\\d{13,19}$"
	HexColorPattern     = "^#([a-f0-9]{6}|[a-f0-9]{3})$"
	Base64Pattern       = "^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$"
	JWTPattern          = "^[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]*$"
	SlugPattern         = "^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$"
	SubdomainPattern    = "^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$"
	TimezonePattern     = "^[A-Za-z_]+/[A-Za-z_]+$"
)

// ============================================================================
// AUDIT & TIMESTAMPS (Reusable Functions for Field Definitions)
// ============================================================================

// AuditFields provides common audit trail fields
var AuditFields = func() {
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
		Meta("struct:tag:json", "created_at")
		Meta("struct:tag:db", "created_at,omitempty")
	})
	Attribute("updated_at", String, "Last update timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T15:45:00Z")
		Meta("struct:tag:json", "updated_at,omitempty")
		Meta("struct:tag:db", "updated_at,omitempty")
	})
	Attribute("created_by", String, "ID of user who created the record", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Meta("struct:tag:json", "created_by,omitempty")
		Meta("struct:tag:db", "created_by,omitempty")
	})
	Attribute("updated_by", String, "ID of user who last updated the record", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Meta("struct:tag:json", "updated_by,omitempty")
		Meta("struct:tag:db", "updated_by,omitempty")
	})
}

// SoftDeleteFields provides soft delete audit fields
var SoftDeleteFields = func() {
	Attribute("deleted_at", String, "Deletion timestamp (null if not deleted)", func() {
		Format(FormatDateTime)
		Example("2023-12-07T16:00:00Z")
		Meta("struct:tag:json", "deleted_at,omitempty")
		Meta("struct:tag:db", "deleted_at,omitempty")
	})
	Attribute("deleted_by", String, "ID of user who deleted the record", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Meta("struct:tag:json", "deleted_by,omitempty")
		Meta("struct:tag:db", "deleted_by,omitempty")
	})
}

// VersioningFields provides record versioning support
var VersioningFields = func() {
	Attribute("version", UInt, "Optimistic locking version number", func() {
		Minimum(1)
		Example(5)
		Meta("struct:tag:json", "version,omitempty")
		Meta("struct:tag:db", "version,omitempty")
		Description("Incremented on each update for optimistic concurrency control")
	})
}

// ============================================================================
// PAGINATION & SEARCH
// ============================================================================

// Pagination provides common pagination parameters
var Pagination = Type("Pagination", func() {
	Description("Standard pagination parameters for list endpoints")
	Attribute("page", UInt, "Page number (1-based)", func() {
		Default(1)
		Minimum(1)
		Example(1)
		Meta("struct:tag:json", "page,omitempty")
		Meta("struct:tag:query", "page")
		Meta("struct:tag:validate", "omitempty,min=1")
	})
	Attribute("page_size", UInt, "Number of items per page", func() {
		Default(20)
		Minimum(1)
		Maximum(100)
		Example(20)
		Meta("struct:tag:json", "page_size,omitempty")
		Meta("struct:tag:query", "page_size")
		Meta("struct:tag:validate", "omitempty,min=1,max=100")
	})
	Attribute("sort_by", String, "Field to sort by", func() {
		Default("created_at")
		Example("name")
		Meta("struct:tag:json", "sort_by,omitempty")
		Meta("struct:tag:query", "sort_by")
	})
	Attribute("sort_order", String, "Sort direction", func() {
		Enum("asc", "desc")
		Default("desc")
		Example("desc")
		Meta("struct:tag:json", "sort_order,omitempty")
		Meta("struct:tag:query", "sort_order")
		Meta("struct:tag:validate", "omitempty,oneof=asc desc")
	})
})

// PaginationMeta provides pagination metadata for responses
var PaginationMeta = Type("PaginationMeta", func() {
	Description("Pagination metadata returned with list responses")
	Attribute("current_page", UInt, "Current page number", func() {
		Example(1)
		Meta("struct:tag:json", "current_page")
	})
	Attribute("page_size", UInt, "Items per page", func() {
		Example(20)
		Meta("struct:tag:json", "page_size")
	})
	Attribute("total_items", UInt, "Total number of items across all pages", func() {
		Example(156)
		Meta("struct:tag:json", "total_items")
	})
	Attribute("total_pages", UInt, "Total number of pages", func() {
		Example(8)
		Meta("struct:tag:json", "total_pages")
	})
	Attribute("has_next", Boolean, "Whether there is a next page", func() {
		Example(true)
		Meta("struct:tag:json", "has_next")
	})
	Attribute("has_prev", Boolean, "Whether there is a previous page", func() {
		Example(false)
		Meta("struct:tag:json", "has_prev")
	})
	Required("current_page", "page_size", "total_items", "total_pages", "has_next", "has_prev")
})

// TimeRange represents a time range for filtering
var TimeRange = Type("TimeRange", func() {
	Description("Time range with start and end dates for filtering")
	Attribute("start_date", String, "Start date (inclusive)", func() {
		Format(FormatDateTime)
		Example("2023-12-01T00:00:00Z")
		Meta("struct:tag:json", "start_date")
		Meta("struct:tag:query", "start_date")
		Meta("struct:tag:validate", "required")
	})
	Attribute("end_date", String, "End date (inclusive)", func() {
		Format(FormatDateTime)
		Example("2023-12-31T23:59:59Z")
		Meta("struct:tag:json", "end_date")
		Meta("struct:tag:query", "end_date")
		Meta("struct:tag:validate", "required")
	})
	Required("start_date", "end_date")
})

// ============================================================================
// AUTHENTICATION & HEADERS
// ============================================================================

// JWTToken provides JWT token attribute for authentication
var JWTToken = func() {
	Token("token", String, "JWT authentication token", func() {
		Description("Bearer token for API authentication")
		Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
		Meta("struct:tag:json", "token")
	})
}

// CommonHeaders for multi-tenant and request context
var CommonHeaders = func() {
	Header("X-Tenant-ID", String, "Tenant identifier for multi-tenancy", func() {
		Pattern(UUIDPattern)
		Example("123e4567-e89b-12d3-a456-426614174000")
		Description("Required for tenant-scoped operations")
	})
	Header("X-Entity-ID", String, "Entity/Company identifier within tenant", func() {
		Pattern(UUIDPattern)
		Example("987fcdeb-51d2-43b8-a456-426614174000")
		Description("Optional sub-tenant organizational unit")
	})
	Header("X-User-ID", String, "Current authenticated user identifier", func() {
		Pattern(UUIDPattern)
		Example("456e7890-e89b-12d3-a456-426614174000")
	})
	Header("X-Request-ID", String, "Unique request correlation ID for tracing", func() {
		Pattern(UUIDPattern)
		Example("req-123e4567-e89b-12d3-a456-426614174000")
	})
	Header("X-API-Version", String, "API version for request", func() {
		Pattern(APIVersionPattern)
		Example("v1.2")
		Default("v1")
	})
}

// RateLimitHeaders provides rate limiting response headers
var RateLimitHeaders = func() {
	Header("X-Rate-Limit-Limit", UInt, "Request limit per window", func() {
		Example(1000)
		Description("Maximum requests allowed in the current window")
	})
	Header("X-Rate-Limit-Remaining", UInt, "Remaining requests in current window", func() {
		Example(999)
		Description("Number of requests remaining")
	})
	Header("X-Rate-Limit-Reset", UInt, "Window reset time (Unix timestamp)", func() {
		Example(1638360000)
		Description("When the rate limit window resets")
	})
	Header("Retry-After", UInt, "Seconds to wait before retrying", func() {
		Example(60)
		Description("Included when rate limit is exceeded")
	})
}

// LocaleHeaders provides localization support
var LocaleHeaders = func() {
	Header("Accept-Language", String, "Preferred language(s) with quality values", func() {
		Example("en-US,en;q=0.9")
		Description("Standard HTTP Accept-Language header")
	})
	Header("X-Timezone", String, "Client timezone (IANA identifier)", func() {
		Pattern(TimezonePattern)
		Example("America/New_York")
		Description("Used for date/time formatting in responses")
	})
	Header("X-Currency", String, "Preferred currency for monetary values", func() {
		Pattern(CurrencyCodePattern)
		Example("USD")
		Description("ISO 4217 currency code")
	})
}

// SecurityHeaders provides security-related headers
var SecurityHeaders = func() {
	Header("X-API-Key", String, "API key for authentication", func() {
		Example("ak_1234567890abcdef")
		Description("Alternative to JWT for service-to-service authentication")
	})
	Header("X-Client-IP", String, "Original client IP address", func() {
		Pattern(IPAddressPattern)
		Example("192.168.1.100")
		Description("For IP-based access control and audit")
	})
	Header("X-Forwarded-For", String, "Forwarded client IP chain", func() {
		Example("203.0.113.195, 70.41.3.18")
	})
}

// ============================================================================
// GRPC HELPER FUNCTIONS
// ============================================================================

// GRPCSecurityHeaders provides gRPC security-related headers
var GRPCSecurityHeaders = func() {
	Header("authorization", String, "Bearer token for gRPC authentication", func() {
		Example("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
	})
	Header("grpc-metadata-api-key", String, "API key in gRPC metadata", func() {
		Example("ak_1234567890abcdef")
	})
}

// GRPCHeaders provides standard gRPC metadata headers
var GRPCHeaders = func() {
	Header("grpc-metadata-content-type", String, "Content type for gRPC requests", func() {
		Example("application/grpc+proto")
	})
	Header("grpc-metadata-user-agent", String, "gRPC user agent", func() {
		Example("grpc-go/1.40.0")
	})
	Header("grpc-metadata-accept-encoding", String, "gRPC accept encoding", func() {
		Example("gzip")
	})
	Header("grpc-metadata-content-encoding", String, "gRPC content encoding", func() {
		Example("gzip")
	})
	Header("grpc-timeout", String, "gRPC request timeout", func() {
		Example("10S") // 10 seconds
	})
}

// GRPCCommonHeaders combines standard headers for gRPC services
var GRPCCommonHeaders = func() {
	CommonHeaders()       // X-Tenant-ID, X-User-ID, etc.
	GRPCHeaders()         // gRPC-specific headers
	GRPCSecurityHeaders() // Auth headers
	LocaleHeaders()       // Localization
}

// ============================================================================
// ERROR HANDLING TYPES
// ============================================================================

// ErrorResponse provides standardized error response format
var ErrorResponse = Type("ErrorResponse", func() {
	Description("Standard error response structure for all API errors")
	Attribute("error", String, "Human-readable error message", func() {
		MinLength(1)
		MaxLength(500)
		Example("Validation failed")
		Meta("struct:tag:json", "error")
	})
	Attribute("error_code", String, "Machine-readable error code for client handling", func() {
		Pattern("^[A-Z_]+$")
		Example("VALIDATION_ERROR")
		Meta("struct:tag:json", "error_code")
	})
	Attribute("details", ArrayOf(String), "Additional error details or field-specific messages", func() {
		Example([]string{"field 'email' is required", "field 'name' must be at least 2 characters"})
		Meta("struct:tag:json", "details,omitempty")
	})
	Attribute("request_id", String, "Request correlation ID for support and debugging", func() {
		Format(FormatUUID)
		Example("req-123e4567-e89b-12d3-a456-426614174000")
		Meta("struct:tag:json", "request_id")
	})
	Attribute("timestamp", String, "Error occurrence timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
		Meta("struct:tag:json", "timestamp")
	})
	Attribute("path", String, "Request path where error occurred", func() {
		Example("/api/v1/tenants")
		Meta("struct:tag:json", "path,omitempty")
	})
	Required("error", "error_code", "request_id", "timestamp")
})

// ValidationError provides detailed validation error information
var ValidationError = Type("ValidationError", func() {
	Description("Detailed validation error for specific field failures")
	Attribute("field", String, "Field name that failed validation", func() {
		Example("email")
		Meta("struct:tag:json", "field")
	})
	Attribute("message", String, "Human-readable validation error message", func() {
		Example("must be a valid email address")
		Meta("struct:tag:json", "message")
	})
	Attribute("code", String, "Validation error code", func() {
		Enum("REQUIRED", "INVALID_FORMAT", "OUT_OF_RANGE", "TOO_SHORT", "TOO_LONG", "INVALID_VALUE", "DUPLICATE", "NOT_FOUND")
		Example("INVALID_FORMAT")
		Meta("struct:tag:json", "code")
	})
	Attribute("value", Any, "The invalid value that was provided", func() {
		Example("invalid-email")
		Meta("struct:tag:json", "value,omitempty")
	})
	Attribute("constraint", String, "Validation constraint that was violated", func() {
		Example("email format")
		Meta("struct:tag:json", "constraint,omitempty")
	})
	Required("field", "message", "code")
})

// ============================================================================
// BUSINESS TYPES
// ============================================================================

// ContactInfo describes contact information
var ContactInfo = Type("ContactInfo", func() {
	Description("Contact person information")
	Attribute("name", String, "Full name of contact person", func() {
		MinLength(1)
		MaxLength(100)
		Example("John Doe")
		Meta("struct:tag:json", "name,omitempty")
		Meta("struct:tag:validate", "omitempty,min=1,max=100")
	})
	Attribute("email", String, "Contact email address", func() {
		Format(FormatEmail)
		Example("john.doe@acme.com")
		Meta("struct:tag:json", "email,omitempty")
		Meta("struct:tag:validate", "omitempty,email")
	})
	Attribute("phone", String, "Contact phone number with country code", func() {
		Pattern(PhonePattern)
		Example("+1-555-123-4567")
		Meta("struct:tag:json", "phone,omitempty")
		Meta("struct:tag:validate", "omitempty")
	})
	Attribute("mobile", String, "Mobile phone number", func() {
		Pattern(PhonePattern)
		Example("+1-555-987-6543")
		Meta("struct:tag:json", "mobile,omitempty")
		Meta("struct:tag:validate", "omitempty")
	})
	Attribute("title", String, "Job title or role", func() {
		MaxLength(100)
		Example("Chief Technology Officer")
		Meta("struct:tag:json", "title,omitempty")
		Meta("struct:tag:validate", "omitempty,max=100")
	})
	Attribute("department", String, "Department or division", func() {
		MaxLength(100)
		Example("Engineering")
		Meta("struct:tag:json", "department,omitempty")
		Meta("struct:tag:validate", "omitempty,max=100")
	})
})

// Address provides standardized physical address structure
var Address = Type("Address", func() {
	Description("Physical mailing or business address")
	Attribute("street_line_1", String, "Primary street address", func() {
		MinLength(1)
		MaxLength(200)
		Example("123 Main Street")
		Meta("struct:tag:json", "street_line_1")
		Meta("struct:tag:validate", "required,max=200")
	})
	Attribute("street_line_2", String, "Secondary street address (apartment, suite, etc.)", func() {
		MaxLength(200)
		Example("Apt 4B")
		Meta("struct:tag:json", "street_line_2,omitempty")
		Meta("struct:tag:validate", "omitempty,max=200")
	})
	Attribute("city", String, "City name", func() {
		MinLength(1)
		MaxLength(100)
		Example("New York")
		Meta("struct:tag:json", "city")
		Meta("struct:tag:validate", "required,max=100")
	})
	Attribute("state_province", String, "State or province", func() {
		MaxLength(100)
		Example("NY")
		Meta("struct:tag:json", "state_province,omitempty")
		Meta("struct:tag:validate", "omitempty,max=100")
	})
	Attribute("postal_code", String, "Postal or ZIP code", func() {
		MinLength(1)
		MaxLength(20)
		Example("10001")
		Meta("struct:tag:json", "postal_code")
		Meta("struct:tag:validate", "required,max=20")
	})
	Attribute("country", String, "Country code (ISO 3166-1 alpha-2)", func() {
		Pattern(CountryCodePattern)
		Example("US")
		Meta("struct:tag:json", "country,omitempty")
		Meta("struct:tag:validate", "omitempty,len=2")
	})
	Attribute("timezone", String, "Timezone (IANA timezone database)", func() {
		Pattern(TimezonePattern)
		Default("UTC")
		Example("America/New_York")
		Meta("struct:tag:json", "timezone")
		Meta("struct:tag:validate", "required")
	})
	Attribute("date_format", String, "Preferred date format", func() {
		Enum("MM/DD/YYYY", "DD/MM/YYYY", "YYYY-MM-DD", "DD.MM.YYYY")
		Default("MM/DD/YYYY")
		Example("MM/DD/YYYY")
		Meta("struct:tag:json", "date_format,omitempty")
	})
	Attribute("time_format", String, "Preferred time format", func() {
		Enum("12h", "24h")
		Default("12h")
		Example("24h")
		Meta("struct:tag:json", "time_format,omitempty")
	})
	Attribute("currency", String, "Preferred display currency (ISO 4217)", func() {
		Pattern(CurrencyCodePattern)
		Default("USD")
		Example("USD")
		Meta("struct:tag:json", "currency")
		Meta("struct:tag:validate", "required,len=3")
	})
	Attribute("number_format", String, "Number formatting style", func() {
		Enum("1,234.56", "1.234,56", "1 234,56")
		Default("1,234.56")
		Example("1,234.56")
		Meta("struct:tag:json", "number_format,omitempty")
	})
	Required("language", "timezone", "currency")
})

// ============================================================================
// BATCH OPERATIONS
// ============================================================================

// BatchOperation represents batch operation request parameters
var BatchOperation = Type("BatchOperation", func() {
	Description("Batch operation request for bulk actions")
	Attribute("operation", String, "Type of batch operation to perform", func() {
		Enum("create", "update", "delete", "patch", "archive", "restore")
		Example("update")
		Meta("struct:tag:json", "operation")
		Meta("struct:tag:validate", "required,oneof=create update delete patch archive restore")
	})
	Attribute("items", ArrayOf(String), "Item IDs to process in batch", func() {
		MinLength(1)
		MaxLength(1000)
		Example([]string{
			"550e8400-e29b-41d4-a716-446655440000",
			"550e8400-e29b-41d4-a716-446655440001",
		})
		Meta("struct:tag:json", "items")
		Meta("struct:tag:validate", "required,min=1,max=1000")
	})
	Attribute("data", Any, "Operation-specific data payload", func() {
		Example(map[string]any{
			"status": "active",
			"plan":   "professional",
		})
		Meta("struct:tag:json", "data,omitempty")
	})
	Attribute("options", MapOf(String, Any), "Operation-specific options and flags", func() {
		Example(map[string]any{
			"validate":          true,
			"dry_run":           false,
			"continue_on_error": true,
			"async":             false,
		})
		Meta("struct:tag:json", "options,omitempty")
	})
	Required("operation", "items")
})

// BatchResult represents batch operation results and summary
var BatchResult = Type("BatchResult", func() {
	Description("Batch operation results with success/failure breakdown")
	Attribute("operation_id", String, "Unique batch operation identifier", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Meta("struct:tag:json", "operation_id")
	})
	Attribute("status", String, "Overall operation status", func() {
		Enum("pending", "processing", "completed", "failed", "partial", "cancelled")
		Example("completed")
		Meta("struct:tag:json", "status")
	})
	Attribute("total", UInt, "Total items submitted for processing", func() {
		Example(100)
		Meta("struct:tag:json", "total")
	})
	Attribute("successful", UInt, "Number of successfully processed items", func() {
		Example(95)
		Meta("struct:tag:json", "successful")
	})
	Attribute("failed", UInt, "Number of failed items", func() {
		Example(5)
		Meta("struct:tag:json", "failed")
	})
	Attribute("skipped", UInt, "Number of skipped items", func() {
		Example(0)
		Meta("struct:tag:json", "skipped,omitempty")
	})
	Attribute("errors", ArrayOf("ValidationError"), "Errors for failed items with details", func() {
		Meta("struct:tag:json", "errors,omitempty")
	})
	Attribute("warnings", ArrayOf(String), "Non-critical warnings during processing", func() {
		Example([]string{"Item abc123: quota near limit"})
		Meta("struct:tag:json", "warnings,omitempty")
	})
	Attribute("started_at", String, "Operation start timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
		Meta("struct:tag:json", "started_at")
	})
	Attribute("completed_at", String, "Operation completion timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:35:00Z")
		Meta("struct:tag:json", "completed_at,omitempty")
	})
	Attribute("duration_ms", UInt, "Operation duration in milliseconds", func() {
		Example(5432)
		Meta("struct:tag:json", "duration_ms,omitempty")
	})
	Required("operation_id", "status", "total", "successful", "failed", "started_at")
})

// BulkActionRequest provides bulk action request structure
var BulkActionRequest = Type("BulkActionRequest", func() {
	Description("Request for performing bulk actions on multiple items")
	Attribute("action", String, "Action to perform on all items", func() {
		Example("archive")
		Meta("struct:tag:json", "action")
		Meta("struct:tag:validate", "required")
	})
	Attribute("items", ArrayOf(String), "Item identifiers to act upon", func() {
		MinLength(1)
		MaxLength(1000)
		Example([]string{
			"550e8400-e29b-41d4-a716-446655440000",
			"550e8400-e29b-41d4-a716-446655440001",
		})
		Meta("struct:tag:json", "items")
		Meta("struct:tag:validate", "required,min=1,max=1000")
	})
	Attribute("parameters", MapOf(String, Any), "Action-specific parameters", func() {
		Example(map[string]any{
			"reason": "bulk cleanup",
			"notify": false,
			"force":  false,
		})
		Meta("struct:tag:json", "parameters,omitempty")
	})
	Attribute("scheduled_at", String, "Optional scheduled execution time", func() {
		Format(FormatDateTime)
		Example("2023-12-08T00:00:00Z")
		Meta("struct:tag:json", "scheduled_at,omitempty")
	})
	Required("action", "items")
})

// ============================================================================
// UTILITY TYPES
// ============================================================================

// CustomMetadata provides generic key-value metadata storage
var CustomMetadata = MapOf(String, Any, func() {
	Description("Generic metadata key-value pairs for extensibility")
	Example(map[string]any{
		"source":       "api",
		"priority":     "high",
		"tags":         []string{"important", "customer"},
		"custom_field": "custom_value",
	})
	Meta("struct:tag:json", "metadata,omitempty")
	Meta("struct:tag:db", "metadata,omitempty")
})

// GenericResponse provides a flexible response structure
var GenericResponse = Type("GenericResponse", func() {
	Description("Generic API response structure for various endpoints")
	Attribute("success", Boolean, "Whether the operation was successful", func() {
		Example(true)
		Meta("struct:tag:json", "success")
	})
	Attribute("message", String, "Human-readable response message", func() {
		MaxLength(500)
		Example("Operation completed successfully")
		Meta("struct:tag:json", "message,omitempty")
	})
	Attribute("data", Any, "Response data payload", func() {
		Meta("struct:tag:json", "data,omitempty")
	})
	Attribute("metadata", CustomMetadata, "Additional response metadata", func() {
		Meta("struct:tag:json", "metadata,omitempty")
	})
	Required("success")
})

// IDRequest represents a simple ID-based request
var IDRequest = Type("IDRequest", func() {
	Description("Request containing a single identifier")
	Attribute("id", String, "Resource identifier", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Meta("struct:tag:json", "id")
		Meta("struct:tag:validate", "required,uuid")
	})
	Required("id")
})

// IDsRequest represents a request with multiple identifiers
var IDsRequest = Type("IDsRequest", func() {
	Description("Request containing multiple identifiers")
	Attribute("ids", ArrayOf(String), "Array of resource identifiers", func() {
		MinLength(1)
		MaxLength(100)
		Elem(func() {
			Format(FormatUUID)
		})
		Example([]string{
			"550e8400-e29b-41d4-a716-446655440000",
			"650e8400-e29b-41d4-a716-446655440001",
		})
		Meta("struct:tag:json", "ids")
		Meta("struct:tag:validate", "required,min=1,max=100")
	})
	Required("ids")
})

// ============================================================================
// WEBHOOKS & EVENTS
// ============================================================================

// WebhookPayload represents a webhook event payload structure
var WebhookPayload = Type("WebhookPayload", func() {
	Description("Webhook event payload sent to registered endpoints")
	Attribute("id", String, "Unique webhook event identifier", func() {
		Format(FormatUUID)
		Example("webhook-123e4567-e89b-12d3-a456-426614174000")
		Meta("struct:tag:json", "id")
	})
	Attribute("event", String, "Event type (resource.action format)", func() {
		Example("tenant.created")
		Meta("struct:tag:json", "event")
		Description("Format: {resource}.{action} e.g., invoice.paid, user.updated")
	})
	Attribute("tenant_id", String, "Tenant identifier for multi-tenant events", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Meta("struct:tag:json", "tenant_id,omitempty")
	})
	Attribute("data", Any, "Event-specific payload data", func() {
		Meta("struct:tag:json", "data")
		Description("Contains the resource or change information")
	})
	Attribute("metadata", CustomMetadata, "Additional event metadata", func() {
		Meta("struct:tag:json", "metadata,omitempty")
	})
	Attribute("timestamp", String, "Event occurrence timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
		Meta("struct:tag:json", "timestamp")
	})
	Attribute("signature", String, "HMAC-SHA256 signature for verification", func() {
		Pattern(SHA256Pattern)
		Example("5d41402abc4b2a76b9719d911017c592")
		Meta("struct:tag:json", "signature,omitempty")
		Description("Used to verify webhook authenticity")
	})
	Attribute("delivery_attempt", UInt, "Delivery attempt number (for retries)", func() {
		Minimum(1)
		Example(1)
		Meta("struct:tag:json", "delivery_attempt,omitempty")
	})
	Required("id", "event", "data", "timestamp")
})

// ============================================================================
// EXPORT/IMPORT TYPES
// ============================================================================

// ExportRequest represents a data export request
var ExportRequest = Type("ExportRequest", func() {
	Description("Request parameters for data export operations")
	Attribute("format", String, "Export file format", func() {
		Enum("csv", "xlsx", "json", "xml", "pdf")
		Default("csv")
		Example("xlsx")
		Meta("struct:tag:json", "format")
		Meta("struct:tag:validate", "required,oneof=csv xlsx json xml pdf")
	})
	Attribute("filters", MapOf(String, Any), "Filters to apply to exported data", func() {
		Example(map[string]any{
			"status":        "active",
			"created_after": "2023-01-01T00:00:00Z",
		})
		Meta("struct:tag:json", "filters,omitempty")
	})
	Attribute("fields", ArrayOf(String), "Specific fields to include in export", func() {
		Example([]string{"id", "name", "email", "created_at"})
		Meta("struct:tag:json", "fields,omitempty")
	})
	Attribute("sort_by", String, "Field to sort exported data by", func() {
		Example("created_at")
		Meta("struct:tag:json", "sort_by,omitempty")
	})
	Attribute("sort_order", String, "Sort order for export", func() {
		Enum("asc", "desc")
		Default("asc")
		Example("asc")
		Meta("struct:tag:json", "sort_order,omitempty")
	})
	Attribute("limit", UInt, "Maximum number of records to export", func() {
		Maximum(100000)
		Example(10000)
		Meta("struct:tag:json", "limit,omitempty")
	})
	Required("format")
})

// ExportStatus represents the status of an export operation
var ExportStatus = Type("ExportStatus", func() {
	Description("Status and progress of data export operation")
	Attribute("id", String, "Export operation identifier", func() {
		Format(FormatUUID)
		Example("export-123e4567-e89b-12d3-a456-426614174000")
		Meta("struct:tag:json", "id")
	})
	Attribute("status", String, "Current export status", func() {
		Enum("pending", "processing", "completed", "failed", "cancelled")
		Example("completed")
		Meta("struct:tag:json", "status")
	})
	Attribute("progress", UInt, "Export progress percentage (0-100)", func() {
		Maximum(100)
		Example(85)
		Meta("struct:tag:json", "progress,omitempty")
	})
	Attribute("total_records", UInt, "Total records to export", func() {
		Example(10000)
		Meta("struct:tag:json", "total_records,omitempty")
	})
	Attribute("processed_records", UInt, "Records processed so far", func() {
		Example(8500)
		Meta("struct:tag:json", "processed_records,omitempty")
	})
	Attribute("file_url", String, "Download URL for completed export", func() {
		Format(FormatURI)
		Example("https://cdn.example.com/exports/export-123.xlsx")
		Meta("struct:tag:json", "file_url,omitempty")
	})
	Attribute("file_size_bytes", UInt64, "Size of export file in bytes", func() {
		Example(2048576)
		Meta("struct:tag:json", "file_size_bytes,omitempty")
	})
	Attribute("expires_at", String, "Download URL expiration time", func() {
		Format(FormatDateTime)
		Example("2023-12-08T10:30:00Z")
		Meta("struct:tag:json", "expires_at,omitempty")
	})
	Attribute("error", String, "Error message if export failed", func() {
		Example("Export failed: insufficient storage")
		Meta("struct:tag:json", "error,omitempty")
	})
	AuditFields()
	Required("id", "status")
})

// ============================================================================
// STANDARD HTTP RESPONSE HELPERS
// ============================================================================

// StandardErrorResponses adds common HTTP error responses to service definitions
var StandardErrorResponses = func() {
	HTTP(func() {
		Response(StatusBadRequest, func() {
			Description("Bad request - validation or syntax error")
			ContentType("application/json")
		})
		Response(StatusUnauthorized, func() {
			Description("Unauthorized - authentication required or failed")
			ContentType("application/json")
		})
		Response(StatusForbidden, func() {
			Description("Forbidden - insufficient permissions")
			ContentType("application/json")
		})
		Response(StatusNotFound, func() {
			Description("Resource not found")
			ContentType("application/json")
		})
		Response(StatusConflict, func() {
			Description("Conflict - resource already exists or state conflict")
			ContentType("application/json")
		})
		Response(StatusUnprocessableEntity, func() {
			Description("Unprocessable entity - semantic validation failed")
			ContentType("application/json")
		})
		Response(StatusTooManyRequests, func() {
			Description("Rate limit exceeded")
			ContentType("application/json")
			Header("Retry-After")
		})
		Response(StatusInternalServerError, func() {
			Description("Internal server error")
			ContentType("application/json")
		})
		Response(StatusServiceUnavailable, func() {
			Description("Service temporarily unavailable")
			ContentType("application/json")
		})
	})
}

// ============================================================================
// METADATA TAGS FOR ENDPOINTS
// ============================================================================

// PublicEndpoint marks an endpoint as publicly accessible
var PublicEndpoint = func() {
	Meta("swagger:tag:Public")
	Meta("goa:public", "true")
	Description("This endpoint is publicly accessible without authentication")
}

// BetaEndpoint marks an endpoint as beta/experimental
var BetaEndpoint = func() {
	Meta("swagger:tag:Beta")
	Meta("goa:beta", "true")
	Description("This endpoint is in beta and may change without notice")
}

// DeprecatedEndpoint marks an endpoint as deprecated
var DeprecatedEndpoint = func() {
	Meta("swagger:deprecated", "true")
	Meta("goa:deprecated", "true")
	Description("This endpoint is deprecated and will be removed in a future version")
}

// ============================================================================
// CACHE CONTROL HELPERS
// ============================================================================

// CacheableResponse adds cache control headers for GET requests
var CacheableResponse = func(maxAge int) func() {
	return func() {
		Header("Cache-Control", String, "Cache control directives", func() {
			Example("public, max-age=3600")
		})
		Header("ETag", String, "Entity tag for cache validation", func() {
			Example("\"33a64df551425fcc55e4d42a148795d9f25f89d4\"")
		})
		Header("Last-Modified", String, "Last modification timestamp", func() {
			Format(FormatDateTime)
			Example("Wed, 07 Dec 2023 10:30:00 GMT")
		})
	}
}

// NoCacheResponse disables caching for sensitive endpoints
var NoCacheResponse = func() {
	Header("Cache-Control", String, "Disable caching", func() {
		Example("no-store, no-cache, must-revalidate, private")
	})
	Header("Pragma", String, "Legacy cache control", func() {
		Example("no-cache")
	})
}

// ============================================================================
// COMMON FIELD HELPERS
// ============================================================================

// UUIDField creates a standard UUID field definition
var UUIDField = func(number int, name, description string) {
	Field(number, name, String, description, func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Meta("struct:tag:json", name)
		Meta("struct:tag:db", name+",omitempty")
		Meta("struct:tag:validate", "required,uuid")
	})
}

// EmailField creates a standard email field definition
var EmailField = func(number int, name, description string, required bool) {
	Field(number, name, String, description, func() {
		Format(FormatEmail)
		Example("user@example.com")
		Meta("struct:tag:json", name)
		if required {
			Meta("struct:tag:validate", "required,email")
		} else {
			Meta("struct:tag:json", name+",omitempty")
			Meta("struct:tag:validate", "omitempty,email")
		}
	})
}

// SlugField creates a URL-friendly slug field
var SlugField = func(number int, name, description string) {
	Field(number, name, String, description, func() {
		Pattern(SlugPattern)
		MinLength(3)
		MaxLength(50)
		Example("example-slug")
		Meta("struct:tag:json", name)
		Meta("struct:tag:validate", "required,min=3,max=50")
	})
}

// Coordinates provides geographic coordinates (standalone)
var Coordinates = Type("Coordinates", func() {
	Description("Geographic coordinates in decimal degrees")
	Attribute("latitude", Float64, "Latitude in decimal degrees", func() {
		Minimum(-90.0)
		Maximum(90.0)
		Example(40.7128)
		Meta("struct:tag:json", "latitude")
		Meta("struct:tag:validate", "required,min=-90,max=90")
	})
	Attribute("longitude", Float64, "Longitude in decimal degrees", func() {
		Minimum(-180.0)
		Maximum(180.0)
		Example(-74.0060)
		Meta("struct:tag:json", "longitude")
		Meta("struct:tag:validate", "required,min=-180,max=180")
	})
	Required("latitude", "longitude")
})

// ============================================================================
// FILE MANAGEMENT
// ============================================================================

// FileInfo represents file metadata for uploads/downloads
var FileInfo = Type("FileInfo", func() {
	Description("File metadata and access information")
	Attribute("id", String, "Unique file identifier", func() {
		Format(FormatUUID)
		Example("file-123e4567-e89b-12d3-a456-426614174000")
		Meta("struct:tag:json", "id")
	})
	Attribute("name", String, "Original filename with extension", func() {
		MinLength(1)
		MaxLength(255)
		Example("invoice_2023_12.pdf")
		Meta("struct:tag:json", "name")
		Meta("struct:tag:validate", "required,max=255")
	})
	Attribute("size", UInt64, "File size in bytes", func() {
		Example(1048576)
		Meta("struct:tag:json", "size")
	})
	Attribute("mime_type", String, "MIME content type", func() {
		Pattern(MIMETypePattern)
		Example("application/pdf")
		Meta("struct:tag:json", "mime_type")
	})
	Attribute("url", String, "Download URL (may be pre-signed and temporary)", func() {
		Format(FormatURI)
		Example("https://cdn.example.com/files/document.pdf")
		Meta("struct:tag:json", "url,omitempty")
	})
	Attribute("thumbnail_url", String, "Thumbnail URL for images/previews", func() {
		Format(FormatURI)
		Example("https://cdn.example.com/thumbs/document_thumb.jpg")
		Meta("struct:tag:json", "thumbnail_url,omitempty")
	})
	Attribute("checksum", String, "SHA-256 checksum for integrity verification", func() {
		Pattern(SHA256Pattern)
		Example("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
		Meta("struct:tag:json", "checksum,omitempty")
	})
	Attribute("storage_path", String, "Internal storage path (not exposed to clients)", func() {
		Example("/uploads/2023/12/file-123e4567.pdf")
		Meta("struct:tag:json", "storage_path,omitempty")
	})
	Attribute("content_disposition", String, "Content-Disposition header value", func() {
		Example("attachment; filename=\"invoice.pdf\"")
		Meta("struct:tag:json", "content_disposition,omitempty")
	})
	AuditFields()
	Required("id", "name", "size", "mime_type")
})

// ============================================================================
// STATUS & HEALTH
// ============================================================================

// CommonStatus provides standard status enumeration function
var CommonStatus = func() {
	Attribute("status", String, "Current operational status", func() {
		Enum("active", "inactive", "pending", "suspended", "archived", "deleted")
		Default("active")
		Example("active")
		Meta("struct:tag:json", "status")
		Meta("struct:tag:validate", "required,oneof=active inactive pending suspended archived deleted")
	})
}

// CommonHealthStatus for health check endpoints
var CommonHealthStatus = Type("CommonHealthStatus", func() {
	Description("Service health status and diagnostics")
	Attribute("status", String, "Overall aggregated health status", func() {
		Enum("healthy", "degraded", "unhealthy")
		Example("healthy")
		Meta("struct:tag:json", "status")
	})
	Attribute("version", String, "Service version identifier", func() {
		Example("1.2.3")
		Meta("struct:tag:json", "version")
	})
	Attribute("timestamp", String, "Health check timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
		Meta("struct:tag:json", "timestamp")
	})
	Attribute("uptime_seconds", UInt, "Service uptime in seconds", func() {
		Example(3600)
		Meta("struct:tag:json", "uptime_seconds,omitempty")
	})
	Attribute("checks", MapOf(String, Any), "Individual component health checks", func() {
		Example(map[string]any{
			"database": map[string]any{
				"status":       "healthy",
				"latency_ms":   2,
				"message":      "Connected",
				"last_checked": "2023-12-07T10:30:00Z",
			},
			"redis": map[string]any{
				"status":     "healthy",
				"latency_ms": 1,
				"message":    "Connected",
			},
			"storage": map[string]any{
				"status":        "healthy",
				"usage_percent": 45.2,
				"available_gb":  500,
			},
		})
		Meta("struct:tag:json", "checks,omitempty")
	})
	Required("status", "version", "timestamp")
})

// ============================================================================
// LOCALIZATION & PREFERENCES
// ============================================================================

// Localization provides locale and timezone information
var Localization = Type("Localization", func() {
	Description("User or tenant localization preferences")

	Attribute("language", String, "Language code (ISO 639-1 with optional region)", func() {
		Pattern(LanguageCodePattern)
		Default("en")
		Example("en-US")
		Meta("struct:tag:json", "language")
		Meta("struct:tag:validate", "required")
	})

	Attribute("country", String, "Country code (ISO 3166-1 alpha-2)", func() {
		Pattern(CountryCodePattern)
		Example("US")
		Meta("struct:tag:json", "country")
		Meta("struct:tag:validate", "required,len=2")
	})
})

// PaginationMetadata contains pagination metadata for responses
var PaginationMetadata = Type("PaginationMetadata", func() {
	Description("Pagination metadata for list responses")

	Field(1, "page", UInt, "Current page number", func() {
		Example(1)
		Meta("struct:tag:json", "page")
	})

	Field(2, "page_size", UInt, "Items per page", func() {
		Example(20)
		Meta("struct:tag:json", "page_size")
	})

	Field(3, "total_items", UInt, "Total number of items", func() {
		Example(156)
		Meta("struct:tag:json", "total_items")
	})

	Field(4, "total_pages", UInt, "Total number of pages", func() {
		Example(8)
		Meta("struct:tag:json", "total_pages")
	})

	Field(5, "has_next", Boolean, "Whether there is a next page", func() {
		Example(true)
		Meta("struct:tag:json", "has_next")
	})

	Field(6, "has_prev", Boolean, "Whether there is a previous page", func() {
		Example(false)
		Meta("struct:tag:json", "has_prev")
	})

	Required("page", "page_size", "total_items", "total_pages", "has_next", "has_prev")
})

// ============================================================================
// MONEY/CURRENCY TYPES
// ============================================================================

// Money represents a monetary amount with currency
var Money = Type("Money", func() {
	Description("Monetary amount with currency information")

	Field(1, "amount", String, "Decimal amount as string for precision", func() {
		Pattern("^-?[0-9]+(\\.[0-9]+)?$")
		Example("1234.56")
		Meta("struct:tag:json", "amount")
		Meta("struct:tag:validate", "required")
		Description("Use string to avoid floating point precision issues")
	})

	Field(2, "currency", String, "Currency code", func() {
		Pattern("^[A-Z]{3}$")
		Example("USD")
		Meta("struct:tag:json", "currency")
		Meta("struct:tag:validate", "required,len=3")
		Description("ISO 4217 currency code")
	})

	Field(3, "formatted", String, "Human-readable formatted amount", func() {
		Example("$1,234.56")
		Meta("struct:tag:json", "formatted,omitempty")
		Description("Locale-specific formatted string")
	})

	Required("amount", "currency")
})

// ============================================================================
// FILE/ATTACHMENT TYPES
// ============================================================================

// FileAttachment represents an uploaded file or document
var FileAttachment = Type("FileAttachment", func() {
	Description("File attachment metadata")

	Field(1, "id", String, "File identifier", func() {
		Format(FormatUUID)
		Example("850e8400-e29b-41d4-a716-446655440000")
		Meta("struct:tag:json", "id")
	})

	Field(2, "filename", String, "Original filename", func() {
		MaxLength(255)
		Example("invoice_2023_12.pdf")
		Meta("struct:tag:json", "filename")
	})

	Field(3, "mime_type", String, "MIME type", func() {
		Example("application/pdf")
		Meta("struct:tag:json", "mime_type")
	})

	Field(4, "size_bytes", UInt, "File size in bytes", func() {
		Example(524288)
		Meta("struct:tag:json", "size_bytes")
	})

	Field(5, "url", String, "Download URL", func() {
		Format(FormatURI)
		Example("https://cdn.example.com/files/850e8400.pdf")
		Meta("struct:tag:json", "url,omitempty")
	})

	Field(6, "thumbnail_url", String, "Thumbnail URL for images", func() {
		Format(FormatURI)
		Example("https://cdn.example.com/thumbs/850e8400.jpg")
		Meta("struct:tag:json", "thumbnail_url,omitempty")
	})

	Field(7, "checksum", String, "File checksum (SHA-256)", func() {
		Pattern("^[a-f0-9]{64}$")
		Example("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
		Meta("struct:tag:json", "checksum,omitempty")
	})

	AuditFields()

	Required("id", "filename", "mime_type", "size_bytes")
})

// ============================================================================
// WEBHOOK TYPES
// ============================================================================

// WebhookEvent represents an event sent to webhooks
var WebhookEvent = Type("WebhookEvent", func() {
	Description("Webhook event payload")

	Field(1, "id", String, "Event identifier", func() {
		Format(FormatUUID)
		Example("a50e8400-e29b-41d4-a716-446655440000")
		Meta("struct:tag:json", "id")
	})

	Field(2, "event_type", String, "Type of event", func() {
		Example("tenant.created")
		Meta("struct:tag:json", "event_type")
		Description("Format: resource.action (e.g., tenant.created, invoice.paid)")
	})

	Field(3, "tenant_id", String, "Tenant identifier", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Meta("struct:tag:json", "tenant_id")
	})

	Field(4, "payload", Any, "Event data", func() {
		Meta("struct:tag:json", "payload")
		Description("Event-specific payload data")
	})

	Field(5, "created_at", String, "Event timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
		Meta("struct:tag:json", "created_at")
	})

	Field(6, "signature", String, "HMAC signature for verification", func() {
		Example("sha256=5d41402abc4b2a76b9719d911017c592")
		Meta("struct:tag:json", "signature,omitempty")
	})

	Required("id", "event_type", "tenant_id", "payload", "created_at")
})

// ============================================================================
// HEALTH CHECK TYPES
// ============================================================================

// HealthCheck represents system health status
var HealthCheck = Type("HealthCheck", func() {
	Description("System health check response")

	Field(1, "status", String, "Overall health status", func() {
		Enum("healthy", "degraded", "unhealthy")
		Example("healthy")
		Meta("struct:tag:json", "status")
	})

	Field(2, "version", String, "Application version", func() {
		Example("1.2.3")
		Meta("struct:tag:json", "version")
	})

	Field(3, "timestamp", String, "Health check timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
		Meta("struct:tag:json", "timestamp")
	})

	Field(4, "checks", Any, "Individual service checks", func() {
		Example(map[string]any{
			"database": map[string]any{
				"status":     "healthy",
				"latency_ms": 12,
				"message":    "Connected",
			},
			"redis": map[string]any{
				"status":     "healthy",
				"latency_ms": 3,
				"message":    "Connected",
			},
			"storage": map[string]any{
				"status":        "healthy",
				"usage_percent": 45.2,
			},
		})
		Meta("struct:tag:json", "checks,omitempty")
	})

	Required("status", "version", "timestamp")
})

// ============================================================================
// FILTER TYPES
// ============================================================================

// DateRangeFilter for filtering by date ranges
var DateRangeFilter = Type("DateRangeFilter", func() {
	Description("Date range filter parameters")

	Field(1, "start_date", String, "Start date (inclusive)", func() {
		Format(FormatDateTime)
		Example("2023-01-01T00:00:00Z")
		Meta("struct:tag:json", "start_date,omitempty")
		Meta("struct:tag:query", "start_date")
	})

	Field(2, "end_date", String, "End date (inclusive)", func() {
		Format(FormatDateTime)
		Example("2023-12-31T23:59:59Z")
		Meta("struct:tag:json", "end_date,omitempty")
		Meta("struct:tag:query", "end_date")
	})
})

// ============================================================================
// RESPONSE WRAPPERS
// ============================================================================

// APIResponse - Generic API response wrapper
var APIResponse = Type("APIResponse", func() {
	Description("Generic API response wrapper")

	Field(1, "success", Boolean, "Operation success status", func() {
		Example(true)
		Meta("struct:tag:json", "success")
	})

	Field(2, "data", Any, "Response data", func() {
		Meta("struct:tag:json", "data,omitempty")
	})

	Field(3, "error", Any, "Error information if success is false", func() {
		Meta("struct:tag:json", "error,omitempty")
	})

	Field(4, "meta", Any, "Additional metadata", func() {
		Example(map[string]any{
			"request_id": "req_8f7d9e5c4b3a2f1e",
			"timestamp":  "2023-12-07T10:30:00Z",
		})
		Meta("struct:tag:json", "meta,omitempty")
	})

	Required("success")
})

// ============================================================================
// ENUM DEFINITIONS (for reference)
// ============================================================================

// Common status values used across the system
const (
	StatusActive       = "ACTIVE"
	StatusInactive     = "INACTIVE"
	StatusSuspended    = "SUSPENDED"
	StatusArchived     = "ARCHIVED"
	StatusPendingSetup = "PENDING_SETUP"
	StatusTrial        = "TRIAL"
)

// Common notification types
const (
	NotificationInfo    = "INFO"
	NotificationSuccess = "SUCCESS"
	NotificationWarning = "WARNING"
	NotificationError   = "ERROR"
)

// Common sort orders
const (
	SortOrderAsc  = "asc"
	SortOrderDesc = "desc"
)

// TenantScopedEndpoint marks an endpoint as requiring tenant context
var TenantScopedEndpoint = func() {
	Meta("swagger:tag:Tenant-Scoped")
	Description("This endpoint requires tenant context resolution")
}

// AdminOnlyEndpoint marks an endpoint as admin-only
var AdminOnlyEndpoint = func() {
	Meta("swagger:tag:Admin")
	Description("This endpoint requires administrator privileges")
}
