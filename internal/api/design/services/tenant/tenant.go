package tenant

import (
	. "github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// Service describes the tenant management service
var _ = Service("tenant", func() {
	Description("Tenant management service for multi-tenant ERP system")

	// Apply global middleware
	HTTP(func() {
		Path("/api/v1/tenants")
	})

	// Create tenant endpoint
	Method("create", func() {
		Description("Create a new tenant")

		Payload(CreateTenantPayload)
		Result(CreateTenantResult)

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

		Result(TenantResult)

		Error("not_found")
		Error("unauthorized")

		HTTP(func() {
			GET("/{id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	// List tenants with pagination
	Method("list", func() {
		Description("List tenants with pagination and filtering")

		Payload(func() {
			Extend(Pagination)
			Attribute("name_filter", String, "Filter by tenant name", func() {
				Example("acme")
			})
			Attribute("status_filter", String, "Filter by status", func() {
				Enum("active", "inactive", "suspended")
				Example("active")
			})
		})

		Result(func() {
			Attribute("data", ArrayOf(TenantResult), "The data items")
			Attribute("pagination", PaginationMeta, "Pagination metadata")
			Required("data", "pagination")
		})

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
	})

	// Update tenant
	Method("update", func() {
		Description("Update an existing tenant")

		Payload(UpdateTenantPayload)
		Result(TenantResult)

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

		// No authentication required for health checks
		NoSecurity()
	})

	// =====================================================
	// TENANT PROVISIONING & LIFECYCLE MANAGEMENT
	// =====================================================

	// Provision tenant endpoint
	Method("provision", func() {
		Description("Provision a new tenant with complete setup")
		Payload(func() {
			Attribute("name", String, "Tenant name", func() {
				MinLength(2)
				MaxLength(255)
				Example("Acme Corporation")
			})
			Attribute("subdomain", String, "Subdomain for tenant access", func() {
				Pattern("^[a-z0-9]([a-z0-9-]*[a-z0-9])?$")
				MinLength(2)
				MaxLength(63)
				Example("acme-corp")
			})
			Attribute("contact_email", String, "Primary contact email", func() {
				Format("email")
				Example("admin@acme.com")
			})
			Attribute("admin_email", String, "Initial admin user email", func() {
				Format("email")
				Example("john.smith@acme.com")
			})
			Attribute("admin_first_name", String, "Admin first name", func() {
				Example("John")
			})
			Attribute("admin_last_name", String, "Admin last name", func() {
				Example("Smith")
			})
			Required("name", "subdomain", "contact_email", "admin_email", "admin_first_name", "admin_last_name")
		})
		Result(func() {
			Attribute("tenant_id", String, "Created tenant ID", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("status", String, "Provisioning status", func() {
				Enum("SUCCESS", "PARTIAL", "FAILED")
				Example("SUCCESS")
			})
			Attribute("message", String, "Status message", func() {
				Example("Tenant provisioned successfully")
			})
			Required("tenant_id", "status", "message")
		})
		Error("bad_request", String, "Invalid request")
		Error("internal_error", String, "Internal server error")
		HTTP(func() {
			POST("/provision")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// Suspend tenant endpoint
	Method("suspend", func() {
		Description("Suspend a tenant")
		Payload(func() {
			Attribute("id", String, "Tenant ID", func() {
				Format(FormatUUID)
			})
			Attribute("reason", String, "Reason for suspension", func() {
				MinLength(10)
				MaxLength(500)
				Example("Non-payment of subscription fees")
			})
			Required("id", "reason")
		})
		Result(func() {
			Attribute("tenant_id", String, "Tenant ID", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("action", String, "Action performed", func() {
				Example("suspend")
			})
			Attribute("status", String, "Action status", func() {
				Enum("COMPLETED", "FAILED")
				Example("COMPLETED")
			})
			Attribute("message", String, "Result message", func() {
				Example("Tenant suspended successfully")
			})
			Required("tenant_id", "action", "status", "message")
		})
		Error("not_found", String, "Tenant not found")
		Error("bad_request", String, "Invalid request")
		Error("internal_error", String, "Internal server error")
		HTTP(func() {
			POST("/{id}/suspend")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("bad_request", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// Reactivate tenant endpoint
	Method("reactivate", func() {
		Description("Reactivate a suspended tenant")
		Payload(func() {
			Attribute("id", String, "Tenant ID", func() {
				Format(FormatUUID)
			})
			Attribute("reason", String, "Reason for reactivation", func() {
				MinLength(5)
				MaxLength(500)
				Example("Payment received")
			})
			Required("id", "reason")
		})
		Result(func() {
			Attribute("tenant_id", String, "Tenant ID", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("action", String, "Action performed", func() {
				Example("reactivate")
			})
			Attribute("status", String, "Action status", func() {
				Enum("COMPLETED", "FAILED")
				Example("COMPLETED")
			})
			Attribute("message", String, "Result message", func() {
				Example("Tenant reactivated successfully")
			})
			Required("tenant_id", "action", "status", "message")
		})
		Error("not_found", String, "Tenant not found")
		Error("bad_request", String, "Invalid request")
		Error("internal_error", String, "Internal server error")
		HTTP(func() {
			POST("/{id}/reactivate")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("bad_request", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// Update configuration endpoint
	Method("update_configuration", func() {
		Description("Update tenant configuration and limits")
		Payload(func() {
			Attribute("id", String, "Tenant ID", func() {
				Format(FormatUUID)
			})
			Attribute("max_users", UInt, "Maximum number of users", func() {
				Example(100)
			})
			Attribute("max_storage_mb", UInt64, "Maximum storage in MB", func() {
				Example(10240)
			})
			Attribute("max_api_calls_per_hour", UInt, "Maximum API calls per hour", func() {
				Example(5000)
			})
			Attribute("reason", String, "Reason for configuration change", func() {
				MinLength(5)
				MaxLength(500)
				Example("Upgrading to professional plan")
			})
			Required("id", "reason")
		})
		Result(func() {
			Attribute("tenant_id", String, "Tenant ID", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("status", String, "Update status", func() {
				Enum("SUCCESS", "FAILED")
				Example("SUCCESS")
			})
			Attribute("message", String, "Result message", func() {
				Example("Configuration updated successfully")
			})
			Required("tenant_id", "status", "message")
		})
		Error("not_found", String, "Tenant not found")
		Error("bad_request", String, "Invalid request")
		Error("internal_error", String, "Internal server error")
		HTTP(func() {
			PUT("/{id}/configuration")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("bad_request", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// Usage analytics endpoint
	Method("get_usage_analytics", func() {
		Description("Get tenant usage analytics")
		Payload(func() {
			Attribute("id", String, "Tenant ID", func() {
				Format(FormatUUID)
			})
			Attribute("period", String, "Time period for analytics", func() {
				Enum("current_month", "last_month", "last_3_months", "last_year")
				Default("current_month")
				Example("current_month")
			})
			Required("id")
		})
		Result(func() {
			Attribute("tenant_id", String, "Tenant ID", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("period", String, "Analysis period", func() {
				Example("current_month")
			})
			Attribute("user_count", UInt, "Number of active users", func() {
				Example(25)
			})
			Attribute("storage_used_mb", UInt64, "Storage used in MB", func() {
				Example(2048)
			})
			Attribute("api_calls", UInt64, "Total API calls", func() {
				Example(45000)
			})
			Required("tenant_id", "period")
		})
		Error("not_found", String, "Tenant not found")
		Error("internal_error", String, "Internal server error")
		HTTP(func() {
			GET("/{id}/analytics")
			Param("period", String, "Time period", func() {
				Default("current_month")
			})
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
		})
	})
})
