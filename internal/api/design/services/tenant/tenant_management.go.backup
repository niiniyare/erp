package tenant

import (
	"github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// TenantManagementService defines tenant lifecycle management
var _ = Service("tenant_management", func() {
	Description("tenant lifecycle management service for multi-tenant ERP system")

	// Apply security for admin operations
	Security("jwt", func() {
		Scope("admin:tenants")
	})

	HTTP(func() {
		Path("/api/v1/admin/tenants")
		types.CommonHeaders()
	})

	// Provision new tenant with complete setup
	Method("provision", func() {
		Description("Provision a new tenant with complete setup including admin user and default configuration")

		Payload(ProvisionTenantPayload)
		Result(ProvisionTenantResult)

		Error("bad_request")
		Error("conflict") // For subdomain/name conflicts
		Error("unauthorized")
		Error("forbidden")
		Error("unprocessable_entity")
		Error("internal_error")

		HTTP(func() {
			POST("/provision")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("conflict", StatusConflict)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("unprocessable_entity", StatusUnprocessableEntity)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// Get tenant with full details including configuration
	Method("get_detailed", func() {
		Description("Get detailed tenant information including configuration and usage stats")

		Payload(func() {
			Attribute("id", String, "Tenant ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Attribute("include", ArrayOf(String), "Additional data to include", func() {
				Enum("configuration", "usage_stats", "subscription", "audit_log")
				Example([]string{"configuration", "usage_stats"})
			})
			Required("id")
		})

		Result(DetailedTenantResult)

		Error("not_found")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			GET("/{id}/detailed")
			Param("include")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// Update tenant configuration
	Method("update_configuration", func() {
		Description("Update tenant configuration including limits, features, and settings")

		Payload(UpdateTenantConfigurationPayload)
		Result(TenantConfigurationResult)

		Error("bad_request")
		Error("not_found")
		Error("unauthorized")
		Error("forbidden")
		Error("unprocessable_entity")

		HTTP(func() {
			PATCH("/{id}/configuration")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("unprocessable_entity", StatusUnprocessableEntity)
		})
	})

	// Suspend tenant (deactivate with data preservation)
	Method("suspend", func() {
		Description("Suspend tenant access while preserving all data")

		Payload(func() {
			Attribute("id", String, "Tenant ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Attribute("reason", String, "Suspension reason", func() {
				MinLength(10)
				MaxLength(500)
				Example("Non-payment of subscription fees")
			})
			Attribute("notify_users", Boolean, "Whether to notify tenant users", func() {
				Default(true)
			})
			Required("id", "reason")
		})

		Result(TenantActionResult)

		Error("bad_request")
		Error("not_found")
		Error("unauthorized")
		Error("forbidden")
		Error("conflict") // If already suspended

		HTTP(func() {
			POST("/{id}/suspend")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("conflict", StatusConflict)
		})
	})

	// Reactivate suspended tenant
	Method("reactivate", func() {
		Description("Reactivate a suspended tenant")

		Payload(func() {
			Attribute("id", String, "Tenant ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Attribute("reason", String, "Reactivation reason", func() {
				MinLength(10)
				MaxLength(500)
				Example("Payment received, subscription renewed")
			})
			Required("id", "reason")
		})

		Result(TenantActionResult)

		Error("bad_request")
		Error("not_found")
		Error("unauthorized")
		Error("forbidden")
		Error("conflict") // If not suspended

		HTTP(func() {
			POST("/{id}/reactivate")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("conflict", StatusConflict)
		})
	})

	// Archive tenant (soft delete with data retention policy)
	Method("archive", func() {
		Description("Archive tenant with configurable data retention policy")

		Payload(func() {
			Attribute("id", String, "Tenant ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Attribute("reason", String, "Archive reason", func() {
				MinLength(10)
				MaxLength(500)
				Example("Tenant requested account deletion")
			})
			Attribute("data_retention_days", UInt, "Days to retain data before permanent deletion", func() {
				Minimum(0)
				Maximum(365)
				Default(90)
				Example(90)
			})
			Attribute("immediate_deletion", Boolean, "Whether to delete immediately", func() {
				Default(false)
			})
			Required("id", "reason")
		})

		Result(TenantActionResult)

		Error("bad_request")
		Error("not_found")
		Error("unauthorized")
		Error("forbidden")
		Error("conflict") // If already archived

		HTTP(func() {
			POST("/{id}/archive")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("conflict", StatusConflict)
		})
	})

	// Bulk operations for multiple tenants
	Method("bulk_operation", func() {
		Description("Perform bulk operations on multiple tenants")

		Payload(func() {
			Attribute("operation", String, "Operation to perform", func() {
				Enum("suspend", "reactivate", "archive", "update_limits")
				Example("suspend")
			})
			Attribute("tenant_ids", ArrayOf(String), "Tenant IDs to operate on", func() {
				MinLength(1)
				MaxLength(100)
				Example([]string{
					"550e8400-e29b-41d4-a716-446655440000",
					"550e8400-e29b-41d4-a716-446655440001",
				})
			})
			Attribute("parameters", MapOf(String, Any), "Operation-specific parameters", func() {
				Example(map[string]any{
					"reason": "Bulk suspension for non-payment",
				})
			})
			Required("operation", "tenant_ids")
		})

		Result(types.BatchResult)

		Error("bad_request")
		Error("unauthorized")
		Error("forbidden")
		Error("unprocessable_entity")

		HTTP(func() {
			POST("/bulk")
			Response(StatusAccepted) // Async operation
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("unprocessable_entity", StatusUnprocessableEntity)
		})
	})

	// Get tenant usage analytics
	Method("get_usage_analytics", func() {
		Description("Get usage analytics for a tenant")

		Payload(func() {
			Attribute("id", String, "Tenant ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Attribute("period", String, "Analytics period", func() {
				Enum("current_month", "last_month", "last_3_months", "last_year")
				Default("current_month")
				Example("current_month")
			})
			Attribute("metrics", ArrayOf(String), "Specific metrics to include", func() {
				Enum("users", "storage", "api_calls", "transactions", "revenue")
				Example([]string{"users", "storage", "api_calls"})
			})
			Required("id")
		})

		Result(TenantUsageAnalyticsResult)

		Error("not_found")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			GET("/{id}/analytics")
			Param("period")
			Param("metrics")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// List tenants with advanced filtering and admin controls
	Method("list_admin", func() {
		Description("List tenants with advanced filtering for administrative purposes")

		Payload(func() {
			Extend(types.Pagination)
			Attribute("status_filter", ArrayOf(String), "Filter by status", func() {
				Enum("active", "suspended", "pending", "archived")
				Example([]string{"active", "suspended"})
			})
			Attribute("plan_filter", ArrayOf(String), "Filter by subscription plan", func() {
				Enum("starter", "professional", "enterprise")
				Example([]string{"professional", "enterprise"})
			})
			Attribute("created_after", String, "Filter by creation date", func() {
				Format(FormatDate)
				Example("2023-01-01")
			})
			Attribute("created_before", String, "Filter by creation date", func() {
				Format(FormatDate)
				Example("2023-12-31")
			})
			Attribute("search", String, "Search in name, subdomain, or email", func() {
				MaxLength(100)
				Example("acme")
			})
			Attribute("include_archived", Boolean, "Include archived tenants", func() {
				Default(false)
			})
		})

		Result(func() {
			Attribute("data", ArrayOf(DetailedTenantResult), "The tenant data")
			Attribute("pagination", types.PaginationMeta, "Pagination metadata")
			Attribute("summary", TenantSummaryStats, "Summary statistics")
			Required("data", "pagination", "summary")
		})

		Error("bad_request")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			GET("/")
			Param("page")
			Param("page_size")
			Param("sort_by")
			Param("sort_order")
			Param("status_filter")
			Param("plan_filter")
			Param("created_after")
			Param("created_before")
			Param("search")
			Param("include_archived")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// Get tenant audit log
	Method("get_audit_log", func() {
		Description("Get audit log for tenant configuration changes")

		Payload(func() {
			Attribute("id", String, "Tenant ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Extend(types.Pagination)
			Attribute("action_filter", ArrayOf(String), "Filter by action type", func() {
				Enum("created", "updated", "suspended", "reactivated", "archived", "config_updated")
				Example([]string{"suspended", "reactivated"})
			})
			Attribute("date_from", String, "Start date for audit log", func() {
				Format(FormatDate)
				Example("2023-01-01")
			})
			Attribute("date_to", String, "End date for audit log", func() {
				Format(FormatDate)
				Example("2023-12-31")
			})
			Required("id")
		})

		Result(func() {
			Attribute("data", ArrayOf(TenantAuditLogEntry), "Audit log entries")
			Attribute("pagination", types.PaginationMeta, "Pagination metadata")
			Required("data", "pagination")
		})

		Error("not_found")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			GET("/{id}/audit-log")
			Param("page")
			Param("page_size")
			Param("action_filter")
			Param("date_from")
			Param("date_to")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})
})