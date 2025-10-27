package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// API DEFINITION - Awo Enterprise Resource Planning System
// ============================================================================

// API describes the global properties of the ERP API server
var _ = API("erp", func() {
	Title("Awo Enterprise Resource Planning Multi-Tenant API")
	Description("Comprehensive ERP system with RBAC/ABAC, finance, HR, and organizational management")
	Version("1.0.0")

	// Server configuration
	Server("erp", func() {
		Host("production", func() {
			URI("https://api.erp.example.com")
		})
		Host("development", func() {
			URI("http://localhost:8080")
		})
	})

	// Global security schemes
	// JWTSecurity("jwt")

	// Global error responses
	Error("internal_error", ErrorResponse, "Internal server error")
	Error("bad_request", ErrorResponse, "Invalid request")
	Error("unauthorized", ErrorResponse, "Authentication required")
	Error("forbidden", ErrorResponse, "Access denied")
	Error("not_found", ErrorResponse, "Resource not found")
	Error("conflict", ErrorResponse, "Resource conflict")
	Error("unprocessable_entity", ValidationError, "Validation failed")
	Error("rate_limit_exceeded", ErrorResponse, "Rate limit exceeded")
})

// ============================================================================
// COMMON VALIDATION PATTERNS
// ============================================================================

// UUID validation pattern for consistent UUID handling
var UUID = func() {
	Format(FormatUUID)
	Example("550e8400-e29b-41d4-a716-446655440000")
}

// Timestamp validation for consistent datetime handling
var Timestamp = func() {
	Format(FormatDateTime)
	Example("2023-12-07T10:30:00Z")
}
