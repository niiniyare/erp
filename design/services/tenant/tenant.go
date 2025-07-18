package tenant

import (
	"github.com/niiniyare/erp/design/types"
	. "goa.design/goa/v3/dsl"
)

// Service describes the tenant management service
var _ = Service("tenant", func() {
	Description("Tenant management service for multi-tenant ERP system")

	// Apply global middleware
	HTTP(func() {
		Path("/api/v1/tenants")
	})

	// Security requirements
	Security("jwt", func() {
		Scope("api:read", "api:write")
	})

	// Create tenant endpoint
	Method("create", func() {
		Description("Create a new tenant")

		Payload(types.CreateTenantPayload)
		Result(types.TenantResult)

		Error("bad_request")
		Error("conflict") // For subdomain conflicts
		Error("unauthorized")
		Error("unprocessable_entity")

		HTTP(func() {
			POST("/")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("conflict", StatusConflict)
			Response("unauthorized", StatusUnauthorized)
			Response("unprocessable_entity", StatusUnprocessableEntity)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("conflict", CodeAlreadyExists)
			Response("unauthorized", CodeUnauthenticated)
			Response("unprocessable_entity", CodeInvalidArgument)
		})
	})

	// Get tenant by ID
	Method("get", func() {
		Description("Get tenant by ID")

		Payload(func() {
			Attribute("id", String, "Tenant ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Required("id")
		})

		Result(types.TenantResult)

		Error("not_found")
		Error("unauthorized")

		HTTP(func() {
			GET("/{id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("not_found", CodeNotFound)
			Response("unauthorized", CodeUnauthenticated)
		})
	})

	// List tenants with pagination
	Method("list", func() {
		Description("List tenants with pagination and filtering")

		Payload(func() {
			Extend(types.Pagination)
			Attribute("name_filter", String, "Filter by tenant name", func() {
				Example("acme")
			})
			Attribute("status_filter", String, "Filter by status", func() {
				Enum("active", "inactive", "suspended")
				Example("active")
			})
		})

		Result(types.PaginatedResponse(types.TenantResult))

		Error("bad_request")
		Error("unauthorized")

		HTTP(func() {
			GET("/")
			Param("page")
			Param("page_size")
			Param("sort_by")
			Param("sort_order")
			Param("name_filter")
			Param("status_filter")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
		})
	})

	// Update tenant
	Method("update", func() {
		Description("Update an existing tenant")

		Payload(types.UpdateTenantPayload)
		Result(types.TenantResult)

		Error("bad_request")
		Error("not_found")
		Error("conflict")
		Error("unauthorized")
		Error("unprocessable_entity")

		HTTP(func() {
			PUT("/{id}")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("conflict", StatusConflict)
			Response("unauthorized", StatusUnauthorized)
			Response("unprocessable_entity", StatusUnprocessableEntity)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("not_found", CodeNotFound)
			Response("conflict", CodeAlreadyExists)
			Response("unauthorized", CodeUnauthenticated)
			Response("unprocessable_entity", CodeInvalidArgument)
		})
	})

	// Delete tenant
	Method("delete", func() {
		Description("Delete a tenant (soft delete)")

		Payload(func() {
			Attribute("id", String, "Tenant ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Required("id")
		})

		Result(Empty) // No content response

		Error("not_found")
		Error("unauthorized")
		Error("conflict") // If tenant has dependencies

		HTTP(func() {
			DELETE("/{id}")
			Response(StatusNoContent)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("conflict", StatusConflict)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("not_found", CodeNotFound)
			Response("unauthorized", CodeUnauthenticated)
			Response("conflict", CodeFailedPrecondition)
		})
	})

	// Health check endpoint
	Method("health", func() {
		Description("Health check for tenant service")

		Result(func() {
			Attribute("status", String, "Service status", func() {
				Enum("healthy", "degraded", "unhealthy")
				Example("healthy")
			})
			Attribute("timestamp", String, "Check timestamp", func() {
				Format(FormatDateTime)
				Example("2023-12-07T10:30:00Z")
			})
			Attribute("version", String, "Service version", func() {
				Example("1.2.3")
			})
			Required("status", "timestamp", "version")
		})

		HTTP(func() {
			GET("/health")
			Response(StatusOK)
		})

		GRPC(func() {
			Response(CodeOK)
		})

		// No authentication required for health checks
		NoSecurity()
	})
})
