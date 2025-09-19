package tenant

import (
	. "goa.design/goa/v3/dsl"
)

// TenantManagement service provides comprehensive tenant lifecycle management
var _ = Service("tenant_management", func() {
	Description("Comprehensive tenant management service for provisioning, configuration, and lifecycle operations")

	// =====================================================
	// TENANT PROVISIONING
	// =====================================================

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
			POST("/tenant-management/provision")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// =====================================================
	// TENANT LIFECYCLE OPERATIONS
	// =====================================================

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
			POST("/tenant-management/{id}/suspend")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("bad_request", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
	})

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
			POST("/tenant-management/{id}/reactivate")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("bad_request", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// =====================================================
	// CONFIGURATION MANAGEMENT
	// =====================================================

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
			PUT("/tenant-management/{id}/configuration")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("bad_request", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// =====================================================
	// ANALYTICS
	// =====================================================

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
			GET("/tenant-management/{id}/analytics")
			Param("period", String, "Time period", func() {
				Default("current_month")
			})
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// Security will be added in next iteration
})