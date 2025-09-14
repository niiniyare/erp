package types

import (
	. "goa.design/goa/v3/dsl"
)

// Common regex patterns as constants for reuse
var (
	UUIDPattern         = "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
	PhonePattern        = "^\\+?[1-9]\\d{1,14}$"
	CountryCodePattern  = "^[A-Z]{2}$"
	CurrencyCodePattern = "^[A-Z]{3}$"
	SHA256Pattern       = "^[a-f0-9]{64}$"
	LanguageCodePattern = "^[a-z]{2}(-[A-Z]{2})?$"
	APIVersionPattern   = "^v\\d+(\\.\\d+)?$"
	DecimalPattern      = "^\\d+(\\.\\d{1,4})?$"
	MIMETypePattern     = "^[a-z-]+/[a-z0-9][a-z0-9!#$&\\-\\^_]*$"
	EmailPattern        = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
	URLPattern          = "^(https?|ftp)://[^\\s/$.?#].[^\\s]*$"
	IPAddressPattern    = "^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$"
	IPv6Pattern         = "^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$"
	DatePattern         = "^\\d{4}-\\d{2}-\\d{2}$"
	DateTimePattern     = "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}(?:\\.\\d+)?Z?$"
	TimePattern         = "^\\d{2}:\\d{2}:\\d{2}$"
	ZipCodePattern      = "^\\d{5}(-\\d{4})?$"
	CreditCardPattern   = "^\\d{13,19}$"
	HexColorPattern     = "^#([a-f0-9]{6}|[a-f0-9]{3})$"
	Base64Pattern       = "^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$"
	JWTPattern          = "^[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]*$"
)

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
// PAGINATION & SEARCH
// ============================================================================

// Pagination provides common pagination parameters
var Pagination = Type("Pagination", func() {
	Description("Pagination parameters")
	Attribute("page", UInt, "Page number (1-based)", func() {
		Default(1)
		Minimum(1)
		Example(1)
	})
	Attribute("page_size", UInt, "Number of items per page", func() {
		Default(20)
		Minimum(1)
		Maximum(100)
		Example(20)
	})
	Attribute("sort_by", String, "Field to sort by", func() {
		Example("created_at")
	})
	Attribute("sort_order", String, "Sort order", func() {
		Enum("asc", "desc")
		Default("desc")
		Example("desc")
	})
})

// PaginationMeta provides pagination metadata
var PaginationMeta = Type("PaginationMeta", func() {
	Description("Pagination metadata")
	Attribute("current_page", UInt, "Current page number")
	Attribute("page_size", UInt, "Items per page")
	Attribute("total_items", UInt, "Total number of items")
	Attribute("total_pages", UInt, "Total number of pages")
	Attribute("has_next", Boolean, "Whether there is a next page")
	Attribute("has_prev", Boolean, "Whether there is a previous page")
	Required("current_page", "page_size", "total_items", "total_pages", "has_next", "has_prev")
})

// SearchFilter provides common search/filter parameters
var SearchFilter = Type("SearchFilter", func() {
	Description("Common search and filter parameters")
	Attribute("q", String, "Search query", func() {
		MaxLength(500)
		Example("john doe")
	})
	Attribute("filters", MapOf(String, Any), "Field-specific filters", func() {
		Example(map[string]any{
			"status":        "active",
			"created_after": "2023-01-01T00:00:00Z",
		})
	})
	Attribute("facets", ArrayOf(String), "Fields to include facet counts for", func() {
		Example([]string{"status", "category"})
	})
})

// ============================================================================
// AUDIT & TIMESTAMPS
// ============================================================================

// AuditFields provides common audit trail fields
var AuditFields = func() {
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	Attribute("updated_at", String, "Last update timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T15:45:00Z")
	})
	Attribute("created_by", String, "ID of user who created the record", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("updated_by", String, "ID of user who last updated the record", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
}

// SoftDeleteFields provides soft delete audit fields
var SoftDeleteFields = func() {
	Attribute("deleted_at", String, "Deletion timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T16:00:00Z")
	})
	Attribute("deleted_by", String, "ID of user who deleted the record", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
}

// TimeRange represents a time range
var TimeRange = Type("TimeRange", func() {
	Description("Time range with start and end dates")
	Attribute("start_date", String, "Start date", func() {
		Format(FormatDateTime)
		Example("2023-12-01T00:00:00Z")
	})
	Attribute("end_date", String, "End date", func() {
		Format(FormatDateTime)
		Example("2023-12-31T23:59:59Z")
	})
	Required("start_date", "end_date")
})

// ============================================================================
// AUTHENTICATION & HEADERS
// ============================================================================

// JWTToken provides JWT token attribute for authentication
var JWTToken = func() {
	Token("token", String, "JWT token", func() {
		Description("JWT authentication token")
		Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
	})
}

// CommonHeaders for multi-tenant context
var CommonHeaders = func() {
	Header("X-Tenant-ID", String, "Tenant identifier", func() {
		Pattern(UUIDPattern)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Header("X-Entity-ID", String, "Entity/Company identifier", func() {
		Pattern(UUIDPattern)
		Example("987fcdeb-51d2-43b8-a456-426614174000")
	})
	Header("X-User-ID", String, "Current user identifier", func() {
		Pattern(UUIDPattern)
		Example("456e7890-e89b-12d3-a456-426614174000")
	})
	Header("X-Request-ID", String, "Request correlation ID", func() {
		Pattern(UUIDPattern)
		Example("req-123e4567-e89b-12d3-a456-426614174000")
	})
	Header("X-API-Version", String, "API version", func() {
		Pattern(APIVersionPattern)
		Example("v1.2")
		Default("v1")
	})
}

// RateLimitHeaders provides rate limiting response headers
var RateLimitHeaders = func() {
	Header("X-Rate-Limit-Limit", UInt, "Request limit per window", func() {
		Example(1000)
	})
	Header("X-Rate-Limit-Remaining", UInt, "Remaining requests in window", func() {
		Example(999)
	})
	Header("X-Rate-Limit-Reset", UInt, "Window reset time (Unix timestamp)", func() {
		Example(1638360000)
	})
}

// LocaleHeaders provides localization support
var LocaleHeaders = func() {
	Header("Accept-Language", String, "Preferred language", func() {
		Example("en-US,en;q=0.9")
	})
	Header("X-Timezone", String, "Client timezone", func() {
		Example("America/New_York")
	})
}

// ============================================================================
// ERROR HANDLING
// ============================================================================

// ErrorResponse provides standardized error response format
var ErrorResponse = Type("ErrorResponse", func() {
	Description("Standard error response")
	Attribute("error", String, "Error message", func() {
		Example("Validation failed")
	})
	Attribute("error_code", String, "Machine-readable error code", func() {
		Example("VALIDATION_ERROR")
	})
	Attribute("details", ArrayOf(String), "Additional error details", func() {
		Example([]string{"field 'email' is required", "field 'name' must be at least 2 characters"})
	})
	Attribute("request_id", String, "Request correlation ID", func() {
		Format(FormatUUID)
		Example("req-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("timestamp", String, "Error timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	Required("error", "error_code", "request_id", "timestamp")
})

// ValidationError provides detailed validation error information
var ValidationError = Type("ValidationError", func() {
	Description("Validation error details")
	Attribute("field", String, "Field name that failed validation", func() {
		Example("email")
	})
	Attribute("message", String, "Validation error message", func() {
		Example("must be a valid email address")
	})
	Attribute("code", String, "Validation error code", func() {
		Example("INVALID_FORMAT")
	})
	Attribute("value", Any, "The invalid value", func() {
		Example("invalid-email")
	})
	Required("field", "message", "code")
})

// ============================================================================
// BUSINESS TYPES
// ============================================================================

// ContactInfo describes contact information
var ContactInfo = Type("ContactInfo", func() {
	Description("Contact information")
	Attribute("name", String, "Contact person name", func() {
		MinLength(1)
		MaxLength(100)
		Example("John Doe")
	})
	Attribute("email", String, "Contact email", func() {
		Format(FormatEmail)
		Example("john.doe@acme.com")
	})
	Attribute("phone", String, "Contact phone number", func() {
		Pattern(PhonePattern)
		Example("+1-555-123-4567")
	})
	Attribute("title", String, "Job title", func() {
		MaxLength(100)
		Example("Chief Technology Officer")
	})
})

// Address provides standardized address structure
var Address = Type("Address", func() {
	Description("Physical address")
	Attribute("street_line_1", String, "Street address line 1", func() {
		MaxLength(100)
		Example("123 Main Street")
	})
	Attribute("street_line_2", String, "Street address line 2 (optional)", func() {
		MaxLength(100)
		Example("Apt 4B")
	})
	Attribute("city", String, "City", func() {
		MaxLength(100)
		Example("New York")
	})
	Attribute("state_province", String, "State or province", func() {
		MaxLength(100)
		Example("NY")
	})
	Attribute("postal_code", String, "Postal or ZIP code", func() {
		MaxLength(20)
		Example("10001")
	})
	Attribute("country", String, "Country code (ISO 3166-1)", func() {
		Pattern(CountryCodePattern)
		Example("US")
	})
	Attribute("latitude", Float64, "Latitude coordinate", func() {
		Minimum(-90)
		Maximum(90)
		Example(40.7128)
	})
	Attribute("longitude", Float64, "Longitude coordinate", func() {
		Minimum(-180)
		Maximum(180)
		Example(-74.0060)
	})
	Required("street_line_1", "city", "country")
})

// Coordinates provides geographic coordinates (standalone)
var Coordinates = Type("Coordinates", func() {
	Description("Geographic coordinates")
	Attribute("latitude", Float64, "Latitude in decimal degrees", func() {
		Minimum(-90.0)
		Maximum(90.0)
		Example(40.7128)
	})
	Attribute("longitude", Float64, "Longitude in decimal degrees", func() {
		Minimum(-180.0)
		Maximum(180.0)
		Example(-74.0060)
	})
	Required("latitude", "longitude")
})

// Money represents monetary values with currency
var Money = Type("Money", func() {
	Description("Monetary amount with currency")
	Attribute("amount", String, "Amount as decimal string to avoid floating point issues", func() {
		Pattern(DecimalPattern)
		Example("99.99")
	})
	Attribute("currency", String, "ISO 4217 currency code", func() {
		Pattern(CurrencyCodePattern)
		Example("USD")
	})
	Required("amount", "currency")
})

// ============================================================================
// FILE MANAGEMENT
// ============================================================================

// FileInfo represents file metadata
var FileInfo = Type("FileInfo", func() {
	Description("File information")
	Attribute("id", String, "File identifier", func() {
		Format(FormatUUID)
		Example("file-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("name", String, "Original filename", func() {
		MaxLength(255)
		Example("document.pdf")
	})
	Attribute("size", UInt64, "File size in bytes", func() {
		Example(1048576)
	})
	Attribute("mime_type", String, "MIME type", func() {
		Pattern(MIMETypePattern)
		Example("application/pdf")
	})
	Attribute("url", String, "File URL", func() {
		Format(FormatURI)
		Example("https://cdn.example.com/files/document.pdf")
	})
	Attribute("checksum", String, "File checksum (SHA-256)", func() {
		Pattern(SHA256Pattern)
		Example("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
	})
	AuditFields()
	Required("id", "name", "size", "mime_type")
})

// ============================================================================
// STATUS & HEALTH
// ============================================================================

// CommonStatus provides standard status enumeration
var CommonStatus = func() {
	Attribute("status", String, "Current status", func() {
		Enum("active", "inactive", "pending", "suspended", "archived")
		Default("active")
		Example("active")
	})
}

// CommonHealthStatus for common health check endpoints
var CommonHealthStatus = Type("CommonHealthStatus", func() {
	Description("Service health status")
	Attribute("status", String, "Overall health status", func() {
		Enum("healthy", "degraded", "unhealthy")
		Example("healthy")
	})
	Attribute("version", String, "Service version", func() {
		Example("1.2.3")
	})
	Attribute("timestamp", String, "Health check timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	Attribute("checks", MapOf(String, Any), "Individual health checks", func() {
		Example(map[string]any{
			"database": map[string]any{
				"status":  "healthy",
				"latency": "2ms",
			},
			"redis": map[string]any{
				"status":  "healthy",
				"latency": "1ms",
			},
		})
	})
	Required("status", "version", "timestamp")
})

// ============================================================================
// LOCALIZATION & PREFERENCES
// ============================================================================

// Localization provides locale and timezone information
var Localization = Type("Localization", func() {
	Description("Localization preferences")
	Attribute("language", String, "Language code (ISO 639-1)", func() {
		Pattern(LanguageCodePattern)
		Default("en")
		Example("en")
	})
	Attribute("country", String, "Country code (ISO 3166-1)", func() {
		Pattern(CountryCodePattern)
		Example("US")
	})
	Attribute("timezone", String, "Timezone (IANA)", func() {
		Example("America/New_York")
	})
	Attribute("date_format", String, "Preferred date format", func() {
		Enum("MM/DD/YYYY", "DD/MM/YYYY", "YYYY-MM-DD")
		Default("MM/DD/YYYY")
		Example("MM/DD/YYYY")
	})
	Attribute("currency", String, "Preferred currency (ISO 4217)", func() {
		Pattern(CurrencyCodePattern)
		Default("USD")
		Example("USD")
	})
})

// ============================================================================
// BATCH OPERATIONS
// ============================================================================

// BatchOperation represents batch operation parameters
var BatchOperation = Type("BatchOperation", func() {
	Description("Batch operation parameters")
	Attribute("operation", String, "Operation type", func() {
		Enum("create", "update", "delete", "patch")
		Example("update")
	})
	Attribute("items", ArrayOf(String), "Item IDs to process", func() {
		MinLength(1)
		MaxLength(1000)
		Example([]string{
			"550e8400-e29b-41d4-a716-446655440000",
			"550e8400-e29b-41d4-a716-446655440001",
		})
	})
	Attribute("options", MapOf(String, Any), "Operation-specific options", func() {
		Example(map[string]any{
			"validate": true,
			"dry_run":  false,
		})
	})
	Required("operation", "items")
})

// BatchResult represents batch operation results
var BatchResult = Type("BatchResult", func() {
	Description("Batch operation results")
	Attribute("operation_id", String, "Unique operation identifier", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("status", String, "Operation status", func() {
		Enum("pending", "processing", "completed", "failed", "partial")
		Example("completed")
	})
	Attribute("total", UInt, "Total items processed", func() {
		Example(100)
	})
	Attribute("successful", UInt, "Successfully processed items", func() {
		Example(95)
	})
	Attribute("failed", UInt, "Failed items", func() {
		Example(5)
	})
	Attribute("errors", ArrayOf("ValidationError"), "Errors for failed items")
	Attribute("started_at", String, "Operation start time", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	Attribute("completed_at", String, "Operation completion time", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:35:00Z")
	})
	Required("operation_id", "status", "total", "successful", "failed")
})

// ============================================================================
// UTILITY TYPES
// ============================================================================

// CustomMetadata provides generic key-value metadata (renamed from MetadataInfo)
var CustomMetadata = MapOf(String, Any, func() {
	Description("Generic metadata key-value pairs")
	Example(map[string]any{
		"source":   "api",
		"priority": "high",
		"tags":     []string{"important", "customer"},
	})
})

// GenericResponse provides a flexible response structure
var GenericResponse = Type("GenericResponse", func() {
	Description("Generic API response structure")
	Attribute("success", Boolean, "Whether the operation was successful", func() {
		Example(true)
	})
	Attribute("message", String, "Response message", func() {
		Example("Operation completed successfully")
	})
	Attribute("data", Any, "Response data")
	Attribute("metadata", CustomMetadata, "Additional metadata") // Changed to CustomMetadata
	Required("success")
})

// BulkActionRequest provides bulk action request structure
var BulkActionRequest = Type("BulkActionRequest", func() {
	Description("Bulk action request")
	Attribute("action", String, "Action to perform", func() {
		Example("archive")
	})
	Attribute("items", ArrayOf(String), "Item identifiers", func() {
		Example([]string{
			"550e8400-e29b-41d4-a716-446655440000",
			"550e8400-e29b-41d4-a716-446655440001",
		})
	})
	Attribute("parameters", MapOf(String, Any), "Action parameters", func() {
		Example(map[string]any{
			"reason": "bulk cleanup",
		})
	})
	Required("action", "items")
})
