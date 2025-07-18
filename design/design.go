package design

import (
	. "goa.design/goa/v3/dsl"
)

// API describes the global properties of the API server.
var _ = API("awo", func() {
	Title("Enterprise ERP System API")
	Description("Comprehensive multi-tenant ERP system with RBAC and ABAC capabilities")
	Version("1.0.0")
	
	// Global configuration
	Server("erp", func() {
		Host("localhost", func() {
			URI("http://localhost:8080")
			URI("https://api.awo.com")
		})
	})
	
	// Global CORS policy
	CORS(func() {
		Origin("*", func() {
			Methods("GET", "POST", "PUT", "DELETE", "OPTIONS")
			Headers("*")
			MaxAge(600)
			Credentials()
		})
	})
	
	// Global error responses
	Error("internal_error", ErrorResult)
	Error("bad_request", ErrorResult)
	Error("unauthorized", ErrorResult)
	Error("forbidden", ErrorResult)
	Error("not_found", ErrorResult)
	Error("conflict", ErrorResult)
	Error("unprocessable_entity", ErrorResult)
})

// ErrorResult defines the error response structure
var ErrorResult = ResultType("application/vnd.erp.error", func() {
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