package featureflag

import (
	. "goa.design/goa/v3/dsl"
)

// AdminFeatureFlagService provides administrative operations for feature flags
var _ = Service("admin-featureflag", func() {
	Description("Administrative interface for feature flag management")

	HTTP(func() {
		Path("/api/v1/admin/feature-flags")
	})

	Security("jwt")

	// Bulk enable flags
	Method("bulk_enable", func() {
		Description("Enable multiple feature flags in bulk")
		Payload(func() {
			Attribute("flag_names", ArrayOf(String), "List of flag names to enable")
			Attribute("reason", String, "Reason for bulk operation")
			Attribute("tenant_id", String, "Tenant ID")
			Token("token", String, "JWT token for authentication")
			Required("flag_names", "reason", "tenant_id", "token")
		})
		Result(func() {
			Attribute("total_requested", Int, "Total number of flags requested")
			Attribute("successful", Int, "Number of successfully processed flags")
			Attribute("failed", Int, "Number of failed operations")
			Attribute("executed_at", String, "Execution timestamp")
			Required("total_requested", "successful", "failed", "executed_at")
		})
		Error("bad_request")
		Error("unauthorized")
		Error("internal_error")

		HTTP(func() {
			POST("/bulk/enable")
			Header("tenant_id:X-Tenant-ID")
			Response(StatusOK)
		})
	})

	// Bulk disable flags
	Method("bulk_disable", func() {
		Description("Disable multiple feature flags in bulk")
		Payload(func() {
			Attribute("flag_names", ArrayOf(String), "List of flag names to disable")
			Attribute("reason", String, "Reason for bulk operation")
			Attribute("tenant_id", String, "Tenant ID")
			Token("token", String, "JWT token for authentication")
			Required("flag_names", "reason", "tenant_id", "token")
		})
		Result(func() {
			Attribute("total_requested", Int, "Total number of flags requested")
			Attribute("successful", Int, "Number of successfully processed flags")
			Attribute("failed", Int, "Number of failed operations")
			Attribute("executed_at", String, "Execution timestamp")
			Required("total_requested", "successful", "failed", "executed_at")
		})
		Error("bad_request")
		Error("unauthorized")
		Error("internal_error")

		HTTP(func() {
			POST("/bulk/disable")
			Header("tenant_id:X-Tenant-ID")
			Response(StatusOK)
		})
	})

	// System health check
	Method("system_health", func() {
		Description("Get system health status")
		Payload(func() {
			Attribute("tenant_id", String, "Tenant ID")
			Token("token", String, "JWT token for authentication")
			Required("tenant_id", "token")
		})
		Result(func() {
			Attribute("status", String, "Overall system status")
			Attribute("timestamp", String, "Health check timestamp")
			Attribute("database_status", String, "Database status")
			Attribute("cache_status", String, "Cache status")
			Attribute("overall_score", Int, "Overall health score (0-100)")
			Required("status", "timestamp", "database_status", "cache_status", "overall_score")
		})
		Error("unauthorized")
		Error("internal_error")

		HTTP(func() {
			GET("/health")
			Header("tenant_id:X-Tenant-ID")
			Response(StatusOK)
		})
	})
})
