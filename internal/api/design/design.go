package design

import (
	_ "github.com/niiniyare/erp/internal/api/design/services/abac"
	_ "github.com/niiniyare/erp/internal/api/design/services/access_request"
	_ "github.com/niiniyare/erp/internal/api/design/services/auth"
	_ "github.com/niiniyare/erp/internal/api/design/services/featureflag"
	_ "github.com/niiniyare/erp/internal/api/design/services/finance"
	_ "github.com/niiniyare/erp/internal/api/design/services/health"
	_ "github.com/niiniyare/erp/internal/api/design/services/organization"
	_ "github.com/niiniyare/erp/internal/api/design/services/tenant"
	_ "github.com/niiniyare/erp/internal/api/design/services/user"
	. "goa.design/goa/v3/dsl"
)

// API describes the global properties of the API server.
var _ = API("awo", func() {
	Title("Enterprise AWO ERP System API")
	Description(" ERP system with multi-tenant support")
	Version("1.0.0")

	// Global configuration
	Server("erp", func() {
		Host("localhost", func() {
			URI("http://localhost:8080")
		})
	})

	// Global error responses
	Error("internal_error", APIError)
	Error("bad_request", APIError)
	Error("unauthorized", APIError)
	Error("forbidden", APIError)
	Error("not_found", APIError)
	Error("conflict", APIError)
	Error("unprocessable_entity", APIError)
})

// APIError defines the error response structure
var APIError = ResultType("application/vnd.erp.error", func() {
	Description("Error response")
	Attributes(func() {
		Attribute("code", String, "Error code", func() {
			Example("TENANT_NOT_FOUND")
		})
		Attribute("message", String, "Error message", func() {
			Example("Tenant with ID 'abc123' not found")
		})
		Attribute("details", MapOf(String, Any), "Additional error details")
		Attribute("timestamp", String, "Error timestamp", func() {
			Format(FormatDateTime)
			Example("2023-12-07T10:30:00Z")
		})
		Attribute("request_id", String, "Request ID for tracking", func() {
			Example("req_abc123def456")
		})
	})
	Required("code", "message", "timestamp", "request_id")
})

// Security schemes
var JWTAuth = JWTSecurity("jwt", func() {
	Description("JWT token authentication")
	Scope("api:read", "Read access to API")
	Scope("api:write", "Write access to API")
	Scope("admin", "Administrative access")
})

var BasicAuth = BasicAuthSecurity("basic_auth", func() {
	Description("Basic authentication for initial setup")
})

var APIKeyAuth = APIKeySecurity("api_key", func() {
	Description("API key authentication for service-to-service communication")
})
