package design

import (
	. "goa.design/goa/v3/dsl"
)

// Security schemes used across the API
var _ = SecurityScheme("jwt", func() {
	JWTSecurity("JWT", func() {
		Description("JWT token authentication")
		Scope("api:read", "Read access to API")
		Scope("api:write", "Write access to API")
		Scope("admin", "Administrative access")
	})
})

var _ = SecurityScheme("api_key", func() {
	APIKeySecurity("api_key", func() {
		Description("API key authentication for service-to-service communication")
	})
})

var _ = SecurityScheme("basic_auth", func() {
	BasicAuthSecurity("basic_auth", func() {
		Description("Basic authentication")
	})
})

// Global middleware
var LoggingMiddleware = func() {
	Middleware(func() {
		Description("Request logging middleware")
	})
}

var AuthMiddleware = func() {
	Middleware(func() {
		Description("Authentication middleware")
	})
}

var TenantMiddleware = func() {
	Middleware(func() {
		Description("Multi-tenant context middleware")
	})
}

// OpenAPI service for documentation
var _ = Service("openapi", func() {
	Description("OpenAPI documentation service")
	Files("/openapi.json", "./gen/http/openapi.json")
	Files("/openapi.yaml", "./gen/http/openapi.yaml")
})
