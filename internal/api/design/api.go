package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = API("awo", func() {
	Title("Multi-Tenant AWO ERP System API")
	Description("A scalable multi-tenant ERP system with REST and gRPC interfaces")
	Version("1.0")

	Server("erp-server", func() {
		Description("Main ERP server")
		Host("localhost", func() {
			URI("http://localhost:8080")
		})
	})
})

var _ = Service("openapi", func() {
	Files("/openapi.json", "./gen/http/openapi.json")
})

var BasicAuth = BasicAuthSecurity("basic_auth")
var JWTAuth = JWTSecurity("jwt", func() {
	Description("Secures endpoint by requiring a valid JWT token.")
	Scope("admin", "Admin scope")
})

// Common result types
var TenantResult = ResultType("application/vnd.tenant", func() {
	Description("Tenant result type")

	Attributes(func() {
		Attribute("id", String, "Tenant unique identifier", func() {
			Meta("rpc:tag", "1")
		})
		Attribute("name", String, "Tenant name", func() {
			Meta("rpc:tag", "2")
		})
		Attribute("subdomain", String, "Tenant subdomain", func() {
			Meta("rpc:tag", "3")
		})
		Attribute("plan_type", String, "Subscription plan type", func() {
			Meta("rpc:tag", "4")
		})
		Attribute("status", String, "Tenant status", func() {
			Meta("rpc:tag", "5")
		})
		Attribute("created_at", String, "Creation timestamp", func() {
			Meta("rpc:tag", "6")
		})
		Attribute("updated_at", String, "Update timestamp", func() {
			Meta("rpc:tag", "7")
		})

		Required("id", "name", "subdomain", "plan_type", "status")
	})

	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("subdomain")
		Attribute("plan_type")
		Attribute("status")
		Attribute("created_at")
		Attribute("updated_at")
	})

	View("minimal", func() {
		Attribute("id")
		Attribute("name")
		Attribute("subdomain")
	})
})
