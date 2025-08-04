package health

import (
	. "goa.design/goa/v3/dsl"
)

// HealthService defines the health check service
var _ = Service("health", func() {
	Description("Health check service for monitoring and readiness")

	// Health check endpoint
	Method("health", func() {
		Description("Check if the service is running")

		HTTP(func() {
			GET("/health")
			Response(StatusOK, func() {
				ContentType("application/json")
			})
		})

		Result(HealthStatus)
	})

	// Readiness check endpoint
	Method("ready", func() {
		Description("Check if the service is ready to handle requests")

		HTTP(func() {
			GET("/ready")
			Response(StatusOK, func() {
				ContentType("application/json")
			})
		})

		Result(ReadinessStatus)
	})
})

// HealthStatus defines the health check response
var HealthStatus = Type("HealthStatus", func() {
	Description("Health check response")

	Attribute("status", String, "Health status", func() {
		Example("ok")
	})
	Attribute("service", String, "Service name", func() {
		Example("awo")
	})
	Attribute("version", String, "Service version", func() {
		Example("1.0.0")
	})

	Required("status", "service", "version")
})

// ReadinessStatus defines the readiness check response
var ReadinessStatus = Type("ReadinessStatus", func() {
	Description("Readiness check response")

	Attribute("status", String, "Readiness status", func() {
		Example("ready")
	})
	Attribute("checks", HealthChecks, "Individual component health checks")

	Required("status", "checks")
})

// HealthChecks defines individual component health status
var HealthChecks = Type("HealthChecks", func() {
	Description("Individual component health checks")

	Attribute("database", String, "Database connection status", func() {
		Example("ok")
	})
	Attribute("cache", String, "Cache connection status", func() {
		Example("ok")
	})

	Required("database", "cache")
})
