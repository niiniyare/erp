package user

import (
	. "goa.design/goa/v3/dsl"
)

// ABAC-related type definitions for user service

// User Attributes Types
var UserAttributesResult = Type("UserAttributesResult", func() {
	Description("User attributes for ABAC evaluation")

	Attribute("user_id", String, "User identifier", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("user_type", String, "User type", func() {
		Enum("INTERNAL", "CUSTOMER", "VENDOR", "PARTNER", "API", "SERVICE", "ADMIN")
		Example("INTERNAL")
	})
	Attribute("retrieved_at", String, "Attribute retrieval timestamp", func() {
		Format(FormatDateTime)
		Example("2025-01-15T14:30:00Z")
	})

	// Core user attributes
	Attribute("attributes", MapOf(String, Any), "User attributes for ABAC", func() {
		Example(map[string]any{
			"user.department":         "finance",
			"user.security_level":     7,
			"user.employment_status":  "active",
			"user.roles":              []string{"financial_analyst", "report_viewer"},
			"user.manager_id":         "mgr_456",
			"user.hire_date":          "2020-03-15",
			"user.last_training_date": "2024-12-01",
			"user.clearance":          "SECRET",
			"user.location":           "headquarters",
			"user.cost_center":        "FIN001",
		})
	})

	// Derived/computed attributes
	Attribute("derived_attributes", MapOf(String, Any), "Computed user attributes", func() {
		Example(map[string]any{
			"user.experience_level":           "senior",
			"user.risk_profile":               "low",
			"user.clearance_expired":          false,
			"user.training_current":           false,
			"user.access_level":               "standard",
			"user.years_of_service":           4.8,
			"user.is_manager":                 false,
			"user.emergency_contact_verified": true,
		})
	})

	// Attribute metadata
	Attribute("attribute_metadata", MapOf(String, AttributeMetadata), "Attribute metadata")

	// Freshness information
	Attribute("freshness_status", FreshnessStatus, "Attribute freshness status")

	// Cache information
	Attribute("cache_info", CacheInfo, "Cache information")

	Required("user_id", "user_type", "retrieved_at", "attributes")
})

var AttributeMetadata = Type("AttributeMetadata", func() {
	Description("Metadata for an attribute")

	Attribute("source", String, "Attribute source system", func() {
		Example("hr_system")
	})
	Attribute("last_updated", String, "Last update timestamp", func() {
		Format(FormatDateTime)
		Example("2025-01-15T09:00:00Z")
	})
	Attribute("confidence", Float64, "Confidence score (0-1)", func() {
		Minimum(0.0)
		Maximum(1.0)
		Example(1.0)
	})
	Attribute("expires_at", String, "Expiration timestamp", func() {
		Format(FormatDateTime)
		Example("2025-01-16T09:00:00Z")
	})
	Attribute("verification_method", String, "Verification method", func() {
		Enum("system_sync", "manual_assignment", "computed", "self_reported")
		Example("system_sync")
	})
	Attribute("data_classification", String, "Data classification level", func() {
		Enum("public", "internal", "confidential", "restricted")
		Example("internal")
	})
})

var FreshnessStatus = Type("FreshnessStatus", func() {
	Description("Attribute freshness status")

	Attribute("fresh_attributes", UInt, "Number of fresh attributes", func() {
		Example(10)
	})
	Attribute("stale_attributes", UInt, "Number of stale attributes", func() {
		Example(1)
	})
	Attribute("expired_attributes", UInt, "Number of expired attributes", func() {
		Example(0)
	})
	Attribute("unknown_freshness", UInt, "Number of attributes with unknown freshness", func() {
		Example(0)
	})
})

var CacheInfo = Type("CacheInfo", func() {
	Description("Cache information")

	Attribute("cache_hit", Boolean, "Whether this was a cache hit", func() {
		Example(true)
	})
	Attribute("cache_age_seconds", UInt, "Age of cached data in seconds", func() {
		Example(145)
	})
	Attribute("cache_ttl_seconds", UInt, "TTL of cached data in seconds", func() {
		Example(1655)
	})
})

// Set Attributes Types
var SetUserAttributesPayload = Type("SetUserAttributesPayload", func() {
	Description("Payload for setting user attributes")

	Attribute("id", String, "User ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("attributes", MapOf(String, Any), "User attributes to set", func() {
		Example(map[string]any{
			"user.department":        "finance",
			"user.security_level":    7,
			"user.employment_status": "active",
		})
	})
	Attribute("metadata", SetAttributeMetadata, "Attribute metadata")
	Attribute("options", SetAttributeOptions, "Set options")

	Required("id", "attributes")
})

var SetAttributeMetadata = Type("SetAttributeMetadata", func() {
	Description("Metadata for setting attributes")

	Attribute("source", String, "Attribute source", func() {
		Example("hr_system")
	})
	Attribute("updated_by", String, "Updated by user/service", func() {
		Example("hr_sync_service")
	})
	Attribute("confidence_score", Float64, "Confidence score", func() {
		Minimum(0.0)
		Maximum(1.0)
		Example(1.0)
	})
	Attribute("last_verified", String, "Last verification timestamp", func() {
		Format(FormatDateTime)
		Example("2025-01-15T14:30:00Z")
	})
	Attribute("expires_at", String, "Expiration timestamp", func() {
		Format(FormatDateTime)
		Example("2025-01-16T14:30:00Z")
	})
	Attribute("data_classification", String, "Data classification", func() {
		Enum("public", "internal", "confidential", "restricted")
		Example("internal")
	})
})

var SetAttributeOptions = Type("SetAttributeOptions", func() {
	Description("Options for setting attributes")

	Attribute("merge_strategy", String, "How to merge with existing attributes", func() {
		Enum("overwrite", "merge", "append")
		Default("overwrite")
		Example("overwrite")
	})
	Attribute("validate_attributes", Boolean, "Validate attributes before setting", func() {
		Default(true)
	})
	Attribute("invalidate_cache", Boolean, "Invalidate related caches", func() {
		Default(true)
	})
	Attribute("notify_subscribers", Boolean, "Notify attribute change subscribers", func() {
		Default(true)
	})
	Attribute("create_audit_trail", Boolean, "Create audit trail entry", func() {
		Default(true)
	})
})

var SetUserAttributesResult = Type("SetUserAttributesResult", func() {
	Description("Result of setting user attributes")

	Attribute("user_id", String, "User identifier", func() {
		Format(FormatUUID)
	})
	Attribute("operation", String, "Operation performed", func() {
		Enum("create", "update", "merge")
		Example("update")
	})
	Attribute("completed_at", String, "Operation completion time", func() {
		Format(FormatDateTime)
		Example("2025-01-15T14:30:00Z")
	})
	Attribute("attributes_summary", AttributesSummary, "Summary of attribute changes")
	Attribute("validation_results", UserValidationResults, "Validation results")
	Attribute("cache_impact", CacheImpact, "Cache impact information")
	Attribute("propagation", PropagationInfo, "Propagation information")
	Attribute("performance", PerformanceInfo, "Performance metrics")

	Required("user_id", "operation", "completed_at", "attributes_summary")
})

var AttributesSummary = Type("AttributesSummary", func() {
	Description("Summary of attribute changes")

	Attribute("total_attributes", UInt, "Total attributes processed", func() {
		Example(10)
	})
	Attribute("attributes_updated", UInt, "Attributes updated", func() {
		Example(7)
	})
	Attribute("attributes_added", UInt, "Attributes added", func() {
		Example(3)
	})
	Attribute("attributes_removed", UInt, "Attributes removed", func() {
		Example(0)
	})
	Attribute("unchanged_attributes", UInt, "Unchanged attributes", func() {
		Example(0)
	})
})

var UserValidationResults = Type("UserValidationResults", func() {
	Description("Attribute validation results")

	Attribute("valid", Boolean, "Whether validation passed", func() {
		Example(true)
	})
	Attribute("warnings", ArrayOf(ValidationMessage), "Validation warnings")
	Attribute("errors", ArrayOf(ValidationMessage), "Validation errors")
})

var ValidationMessage = Type("ValidationMessage", func() {
	Description("Validation message")

	Attribute("attribute", String, "Attribute name", func() {
		Example("user.last_training_date")
	})
	Attribute("message", String, "Validation message", func() {
		Example("Training date is over 30 days old")
	})
	Attribute("severity", String, "Message severity", func() {
		Enum("info", "warning", "error")
		Example("warning")
	})
})

var CacheImpact = Type("CacheImpact", func() {
	Description("Cache impact information")

	Attribute("entries_invalidated", UInt, "Cache entries invalidated", func() {
		Example(45)
	})
	Attribute("policies_affected", UInt, "Policies affected", func() {
		Example(12)
	})
	Attribute("dependent_evaluations_cleared", UInt, "Dependent evaluations cleared", func() {
		Example(156)
	})
})

var PropagationInfo = Type("PropagationInfo", func() {
	Description("Change propagation information")

	Attribute("notifications_sent", UInt, "Notifications sent", func() {
		Example(3)
	})
	Attribute("downstream_services", ArrayOf(String), "Downstream services notified", func() {
		Example([]string{"policy_engine", "audit_service"})
	})
	Attribute("propagation_time_ms", UInt, "Propagation time in milliseconds", func() {
		Example(234)
	})
})

var PerformanceInfo = Type("PerformanceInfo", func() {
	Description("Performance information")

	Attribute("validation_time_ms", Float64, "Validation time in milliseconds", func() {
		Example(4.2)
	})
	Attribute("storage_time_ms", Float64, "Storage time in milliseconds", func() {
		Example(8.7)
	})
	Attribute("cache_update_time_ms", Float64, "Cache update time in milliseconds", func() {
		Example(2.1)
	})
})

// Bulk Update Types
var BulkUpdateUserAttributesPayload = Type("BulkUpdateUserAttributesPayload", func() {
	Description("Payload for bulk updating user attributes")

	Attribute("updates", ArrayOf(UserAttributeUpdate), "User attribute updates")
	Attribute("options", BulkUpdateOptions, "Bulk update options")

	Required("updates")
})

var UserAttributeUpdate = Type("UserAttributeUpdate", func() {
	Description("Single user attribute update")

	Attribute("user_id", String, "User ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("attributes", MapOf(String, Any), "Attributes to update")
	Attribute("metadata", SetAttributeMetadata, "Attribute metadata")

	Required("user_id", "attributes")
})

var BulkUpdateOptions = Type("BulkUpdateOptions", func() {
	Description("Options for bulk updates")

	Attribute("validate_all", Boolean, "Validate all attributes", func() {
		Default(true)
	})
	Attribute("fail_on_error", Boolean, "Fail entire operation on first error", func() {
		Default(false)
	})
	Attribute("batch_size", UInt, "Batch size for processing", func() {
		Default(100)
		Example(100)
	})
	Attribute("parallel_processing", Boolean, "Use parallel processing", func() {
		Default(true)
	})
	Attribute("invalidate_cache", Boolean, "Invalidate caches", func() {
		Default(true)
	})
})

var BulkUpdateUserAttributesResult = Type("BulkUpdateUserAttributesResult", func() {
	Description("Result of bulk updating user attributes")

	Attribute("bulk_update_id", String, "Bulk update identifier", func() {
		Example("bulk_550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("started_at", String, "Start timestamp", func() {
		Format(FormatDateTime)
		Example("2025-01-15T14:30:00Z")
	})
	Attribute("completed_at", String, "Completion timestamp", func() {
		Format(FormatDateTime)
		Example("2025-01-15T14:30:01.234Z")
	})
	Attribute("summary", BulkUpdateSummary, "Update summary")
	Attribute("results", ArrayOf(BulkUpdateResult), "Individual results")
	Attribute("performance_metrics", BulkPerformanceMetrics, "Performance metrics")
	Attribute("downstream_impact", DownstreamImpact, "Downstream impact")

	Required("bulk_update_id", "started_at", "completed_at", "summary", "results")
})

var BulkUpdateSummary = Type("BulkUpdateSummary", func() {
	Description("Summary of bulk update operation")

	Attribute("total_users", UInt, "Total users processed", func() {
		Example(3)
	})
	Attribute("successful_updates", UInt, "Successful updates", func() {
		Example(3)
	})
	Attribute("failed_updates", UInt, "Failed updates", func() {
		Example(0)
	})
	Attribute("warnings", UInt, "Number of warnings", func() {
		Example(1)
	})
	Attribute("total_attributes_updated", UInt, "Total attributes updated", func() {
		Example(8)
	})
})

var BulkUpdateResult = Type("BulkUpdateResult", func() {
	Description("Individual bulk update result")

	Attribute("user_id", String, "User ID", func() {
		Format(FormatUUID)
	})
	Attribute("status", String, "Update status", func() {
		Enum("success", "failed", "partial")
		Example("success")
	})
	Attribute("attributes_updated", UInt, "Attributes updated", func() {
		Example(2)
	})
	Attribute("processing_time_ms", Float64, "Processing time in milliseconds", func() {
		Example(156.7)
	})
	Attribute("warnings", ArrayOf(ValidationMessage), "Warnings")
	Attribute("errors", ArrayOf(ValidationMessage), "Errors")
})

var BulkPerformanceMetrics = Type("BulkPerformanceMetrics", func() {
	Description("Bulk operation performance metrics")

	Attribute("total_processing_time_ms", Float64, "Total processing time", func() {
		Example(488.2)
	})
	Attribute("avg_processing_time_ms", Float64, "Average processing time", func() {
		Example(162.7)
	})
	Attribute("parallel_efficiency", Float64, "Parallel efficiency (0-1)", func() {
		Example(0.89)
	})
	Attribute("cache_invalidations", UInt, "Cache invalidations", func() {
		Example(156)
	})
})

var DownstreamImpact = Type("DownstreamImpact", func() {
	Description("Downstream impact information")

	Attribute("policies_affected", UInt, "Policies affected", func() {
		Example(23)
	})
	Attribute("evaluations_invalidated", UInt, "Evaluations invalidated", func() {
		Example(567)
	})
	Attribute("notifications_triggered", UInt, "Notifications triggered", func() {
		Example(8)
	})
})

// Permission Check Types
var CheckPermissionPayload = Type("CheckPermissionPayload", func() {
	Description("Payload for checking user permission")

	Attribute("user_id", String, "User ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("resource_type", String, "Resource type", func() {
		Example("financial_report")
	})
	Attribute("resource_id", String, "Resource ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440001")
	})
	Attribute("action", String, "Action to check", func() {
		Example("read")
	})
	Attribute("context", MapOf(String, Any), "Additional context", func() {
		Example(map[string]any{
			"urgency":         "normal",
			"business_reason": "quarterly_reporting",
		})
	})
	Attribute("include_explanation", Boolean, "Include decision explanation", func() {
		Default(false)
	})

	Required("user_id", "resource_type", "action")
})

var PermissionCheckResult = Type("PermissionCheckResult", func() {
	Description("Result of permission check")

	Attribute("allowed", Boolean, "Whether permission is granted", func() {
		Example(true)
	})
	Attribute("decision", String, "ABAC decision", func() {
		Enum("ALLOW", "DENY", "PENDING_APPROVAL", "ERROR")
		Example("ALLOW")
	})
	Attribute("evaluation_time_ms", UInt, "Evaluation time in milliseconds", func() {
		Example(15)
	})
	Attribute("request_id", String, "Request identifier", func() {
		Example("req_550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("policy_count", UInt, "Number of policies evaluated", func() {
		Example(3)
	})
	Attribute("cache_hit", Boolean, "Whether this was a cache hit", func() {
		Example(false)
	})
	Attribute("obligations", ArrayOf(UserPolicyObligation), "Policy obligations")
	Attribute("advice", ArrayOf(UserPolicyAdvice), "Policy advice")
	Attribute("explanation", UserPolicyExplanation, "Decision explanation")

	Required("allowed", "decision", "evaluation_time_ms")
})

// Authorization Types
var AuthorizeActionPayload = Type("AuthorizeActionPayload", func() {
	Description("Payload for authorizing user action")

	Attribute("user_id", String, "User ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("resource_type", String, "Resource type", func() {
		Example("financial_report")
	})
	Attribute("resource_id", String, "Resource ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440001")
	})
	Attribute("action", String, "Action to authorize", func() {
		Example("read")
	})
	Attribute("environment", EnvironmentContext, "Environment context")
	Attribute("session_context", SessionContext, "Session context")
	Attribute("additional_context", MapOf(String, Any), "Additional context")

	Required("user_id", "resource_type", "action")
})

var EnvironmentContext = Type("EnvironmentContext", func() {
	Description("Environment context for authorization")

	Attribute("time", TimeContext, "Time context")
	Attribute("location", LocationContext, "Location context")
	Attribute("system", SystemContext, "System context")
	Attribute("security", SecurityContext, "Security context")
})

var TimeContext = Type("TimeContext", func() {
	Description("Time-related context")

	Attribute("current_time", String, "Current timestamp", func() {
		Format(FormatDateTime)
		Example("2025-01-15T14:30:00Z")
	})
	Attribute("timezone", String, "Timezone", func() {
		Example("UTC")
	})
	Attribute("business_hours", Boolean, "Whether in business hours", func() {
		Example(true)
	})
	Attribute("current_hour", UInt, "Current hour (0-23)", func() {
		Example(14)
	})
	Attribute("day_of_week", UInt, "Day of week (1-7)", func() {
		Example(3)
	})
	Attribute("is_weekend", Boolean, "Whether it's weekend", func() {
		Example(false)
	})
	Attribute("is_holiday", Boolean, "Whether it's a holiday", func() {
		Example(false)
	})
})

var LocationContext = Type("LocationContext", func() {
	Description("Location context")

	Attribute("network", String, "Network type", func() {
		Enum("corporate", "vpn", "public", "unknown")
		Example("corporate")
	})
	Attribute("country", String, "Country code", func() {
		Example("US")
	})
	Attribute("region", String, "Region", func() {
		Example("California")
	})
	Attribute("city", String, "City", func() {
		Example("San Francisco")
	})
	Attribute("building", String, "Building", func() {
		Example("headquarters")
	})
	Attribute("floor", String, "Floor", func() {
		Example("5")
	})
})

var SystemContext = Type("SystemContext", func() {
	Description("System context")

	Attribute("load_average", Float64, "System load average", func() {
		Example(0.65)
	})
	Attribute("active_users", UInt, "Active users", func() {
		Example(1247)
	})
	Attribute("maintenance_mode", Boolean, "Whether in maintenance mode", func() {
		Example(false)
	})
	Attribute("version", String, "System version", func() {
		Example("2.1.5")
	})
})

var SecurityContext = Type("SecurityContext", func() {
	Description("Security context")

	Attribute("threat_level", String, "Current threat level", func() {
		Enum("low", "medium", "high", "critical")
		Example("low")
	})
	Attribute("active_incidents", UInt, "Active security incidents", func() {
		Example(0)
	})
	Attribute("security_alert_level", String, "Security alert level", func() {
		Enum("green", "yellow", "orange", "red")
		Example("green")
	})
})

var SessionContext = Type("SessionContext", func() {
	Description("Session context")

	Attribute("id", String, "Session ID", func() {
		Example("session_abc123")
	})
	Attribute("mfa_verified", Boolean, "MFA verification status", func() {
		Example(true)
	})
	Attribute("mfa_verified_at", String, "MFA verification time", func() {
		Format(FormatDateTime)
		Example("2025-01-15T09:00:00Z")
	})
	Attribute("ip_address", String, "IP address", func() {
		Example("192.168.1.100")
	})
	Attribute("user_agent", String, "User agent", func() {
		Example("Mozilla/5.0...")
	})
	Attribute("device_type", String, "Device type", func() {
		Enum("desktop", "mobile", "tablet", "server", "unknown")
		Example("desktop")
	})
	Attribute("security_score", Float64, "Session security score", func() {
		Example(0.95)
	})
})

var AuthorizationResult = Type("AuthorizationResult", func() {
	Description("Authorization result")

	Attribute("allowed", Boolean, "Whether action is authorized", func() {
		Example(true)
	})
	Attribute("decision", String, "Authorization decision", func() {
		Enum("ALLOW", "DENY", "PENDING_APPROVAL", "ERROR")
		Example("ALLOW")
	})
	Attribute("evaluation_time_ms", UInt, "Evaluation time", func() {
		Example(15)
	})
	Attribute("request_id", String, "Request identifier")
	Attribute("obligations", ArrayOf(UserPolicyObligation), "Required obligations")
	Attribute("advice", ArrayOf(UserPolicyAdvice), "Optional advice")
	Attribute("risk_assessment", RiskAssessment, "Risk assessment")
	Attribute("explanation", UserPolicyExplanation, "Decision explanation")

	Required("allowed", "decision", "evaluation_time_ms")
})

// Supporting Types - User service specific versions
var UserPolicyObligation = Type("UserPolicyObligation", func() {
	Description("Policy obligation for user service")

	Attribute("id", String, "Obligation ID", func() {
		Example("obligation_001")
	})
	Attribute("type", String, "Obligation type", func() {
		Example("LOG_ACCESS")
	})
	Attribute("description", String, "Obligation description", func() {
		Example("Log high-value resource access")
	})
	Attribute("parameters", MapOf(String, Any), "Obligation parameters")
})

var UserPolicyAdvice = Type("UserPolicyAdvice", func() {
	Description("Policy advice for user service")

	Attribute("id", String, "Advice ID", func() {
		Example("advice_001")
	})
	Attribute("type", String, "Advice type", func() {
		Example("NOTIFY_MANAGER")
	})
	Attribute("description", String, "Advice description", func() {
		Example("Notify manager of sensitive access")
	})
	Attribute("severity", String, "Advice severity", func() {
		Enum("low", "medium", "high")
		Example("low")
	})
	Attribute("parameters", MapOf(String, Any), "Advice parameters")
})

var UserPolicyExplanation = Type("UserPolicyExplanation", func() {
	Description("Policy decision explanation for user service")

	Attribute("final_decision", String, "Final decision", func() {
		Enum("ALLOW", "DENY", "PENDING_APPROVAL")
		Example("ALLOW")
	})
	Attribute("reasoning_summary", String, "Reasoning summary", func() {
		Example("Access allowed based on user security level and business hours")
	})
	Attribute("combining_algorithm", String, "Policy combining algorithm", func() {
		Example("deny-overrides")
	})
	Attribute("attributes_used", ArrayOf(String), "Attributes used in evaluation", func() {
		Example([]string{"user.security_level", "time.business_hours", "session.mfa_verified"})
	})
	Attribute("recommendations", ArrayOf(String), "Recommendations", func() {
		Example([]string{"Consider additional monitoring for high-value access"})
	})
})

var RiskAssessment = Type("RiskAssessment", func() {
	Description("Risk assessment")

	Attribute("overall_risk_score", UInt, "Overall risk score (0-100)", func() {
		Example(25)
	})
	Attribute("risk_level", String, "Risk level", func() {
		Enum("low", "medium", "high", "critical")
		Example("medium")
	})
	Attribute("risk_factors", ArrayOf(UserRiskFactor), "Risk factors")
	Attribute("additional_monitoring", Boolean, "Whether additional monitoring is recommended", func() {
		Example(false)
	})
})

var UserRiskFactor = Type("UserRiskFactor", func() {
	Description("Individual risk factor")

	Attribute("factor", String, "Risk factor name", func() {
		Example("high_value_resource")
	})
	Attribute("score", UInt, "Risk score contribution", func() {
		Example(15)
	})
	Attribute("weight", Float64, "Risk factor weight", func() {
		Example(0.3)
	})
	Attribute("description", String, "Risk factor description", func() {
		Example("Access to high-value financial resource")
	})
})

// Session Attributes Types
var SessionAttributesResult = Type("SessionAttributesResult", func() {
	Description("Session attributes result")

	Attribute("session_id", String, "Session ID", func() {
		Example("session_abc123")
	})
	Attribute("user_id", String, "User ID", func() {
		Format(FormatUUID)
	})
	Attribute("status", String, "Session status", func() {
		Enum("active", "inactive", "expired", "terminated")
		Example("active")
	})
	Attribute("created_at", String, "Session creation time", func() {
		Format(FormatDateTime)
	})
	Attribute("last_activity", String, "Last activity time", func() {
		Format(FormatDateTime)
	})
	Attribute("duration_minutes", UInt, "Session duration in minutes", func() {
		Example(325)
	})
	Attribute("attributes", MapOf(String, Any), "Session attributes")
	Attribute("activity_summary", ActivitySummary, "Activity summary")
	Attribute("risk_assessment", SessionRiskAssessment, "Session risk assessment")
	Attribute("performance_metrics", SessionPerformanceMetrics, "Performance metrics")

	Required("session_id", "user_id", "status", "created_at")
})

var ActivitySummary = Type("ActivitySummary", func() {
	Description("Session activity summary")

	Attribute("actions_performed", UInt, "Actions performed", func() {
		Example(47)
	})
	Attribute("resources_accessed", UInt, "Resources accessed", func() {
		Example(12)
	})
	Attribute("policy_evaluations", UInt, "Policy evaluations", func() {
		Example(89)
	})
	Attribute("last_action", String, "Last action performed", func() {
		Example("document_read")
	})
	Attribute("last_action_time", String, "Last action time", func() {
		Format(FormatDateTime)
	})
	Attribute("unusual_activity", Boolean, "Whether unusual activity detected", func() {
		Example(false)
	})
	Attribute("activity_pattern", String, "Activity pattern", func() {
		Enum("normal", "unusual", "suspicious", "unknown")
		Example("normal")
	})
})

var SessionRiskAssessment = Type("SessionRiskAssessment", func() {
	Description("Session-specific risk assessment")

	Attribute("overall_risk_score", UInt, "Overall risk score", func() {
		Example(15)
	})
	Attribute("risk_factors", ArrayOf(UserRiskFactor), "Risk factors")
	Attribute("behavioral_analysis", BehavioralAnalysis, "Behavioral analysis")
	Attribute("anomalies", ArrayOf(String), "Detected anomalies")
	Attribute("recommendations", ArrayOf(String), "Security recommendations")
})

var BehavioralAnalysis = Type("BehavioralAnalysis", func() {
	Description("Behavioral pattern analysis")

	Attribute("typing_pattern_match", Float64, "Typing pattern match score", func() {
		Example(0.94)
	})
	Attribute("click_pattern_match", Float64, "Click pattern match score", func() {
		Example(0.87)
	})
	Attribute("navigation_pattern_match", Float64, "Navigation pattern match score", func() {
		Example(0.91)
	})
	Attribute("time_pattern_match", Float64, "Time pattern match score", func() {
		Example(0.89)
	})
})

var SessionPerformanceMetrics = Type("SessionPerformanceMetrics", func() {
	Description("Session performance metrics")

	Attribute("avg_response_time_ms", Float64, "Average response time", func() {
		Example(234.5)
	})
	Attribute("total_data_transferred_mb", Float64, "Total data transferred", func() {
		Example(45.7)
	})
	Attribute("error_count", UInt, "Error count", func() {
		Example(0)
	})
	Attribute("timeout_count", UInt, "Timeout count", func() {
		Example(0)
	})
})

// Set Session Context Types
var SetSessionContextPayload = Type("SetSessionContextPayload", func() {
	Description("Payload for setting session context")

	Attribute("user_id", String, "User ID", func() {
		Format(FormatUUID)
	})
	Attribute("session_id", String, "Session ID", func() {
		Example("session_abc123")
	})
	Attribute("session", SessionContext, "Session information")
	Attribute("computed_attributes", MapOf(String, Any), "Computed session attributes")

	Required("user_id", "session_id", "session")
})

var SetSessionContextResult = Type("SetSessionContextResult", func() {
	Description("Result of setting session context")

	Attribute("session_id", String, "Session ID")
	Attribute("operation", String, "Operation performed", func() {
		Enum("create", "update")
		Example("update")
	})
	Attribute("completed_at", String, "Completion time", func() {
		Format(FormatDateTime)
	})
	Attribute("context_updated", Boolean, "Whether context was updated", func() {
		Example(true)
	})
	Attribute("security_assessment", SecurityAssessment, "Security assessment")
	Attribute("recommendations", ArrayOf(String), "Recommendations")
	Attribute("monitoring", UserMonitoringInfo, "Monitoring information")

	Required("session_id", "operation", "completed_at", "context_updated")
})

var SecurityAssessment = Type("SecurityAssessment", func() {
	Description("Security assessment result")

	Attribute("risk_level", String, "Risk level", func() {
		Enum("low", "medium", "high", "critical")
		Example("low")
	})
	Attribute("trust_score", Float64, "Trust score", func() {
		Example(0.95)
	})
	Attribute("anomalies_detected", UInt, "Anomalies detected", func() {
		Example(0)
	})
	Attribute("location_verified", Boolean, "Location verified", func() {
		Example(true)
	})
	Attribute("device_verified", Boolean, "Device verified", func() {
		Example(true)
	})
	Attribute("behavior_score", Float64, "Behavior score", func() {
		Example(0.92)
	})
})

var UserMonitoringInfo = Type("UserMonitoringInfo", func() {
	Description("Monitoring configuration")

	Attribute("enhanced_monitoring", Boolean, "Enhanced monitoring enabled", func() {
		Example(false)
	})
	Attribute("alert_thresholds", String, "Alert threshold level", func() {
		Enum("minimal", "standard", "enhanced", "maximum")
		Example("standard")
	})
	Attribute("session_timeout_minutes", UInt, "Session timeout in minutes", func() {
		Example(480)
	})
})

// User Context Types
var UserContextResult = Type("UserContextResult", func() {
	Description("Comprehensive user context")

	Attribute("user_id", String, "User ID", func() {
		Format(FormatUUID)
	})
	Attribute("retrieved_at", String, "Context retrieval time", func() {
		Format(FormatDateTime)
	})
	Attribute("user_attributes", MapOf(String, Any), "User attributes")
	Attribute("derived_attributes", MapOf(String, Any), "Derived attributes")
	Attribute("current_session", SessionContext, "Current session context")
	Attribute("access_patterns", UserAccessPatterns, "Access patterns")
	Attribute("risk_profile", UserRiskProfile, "User risk profile")
	Attribute("compliance_status", ComplianceStatus, "Compliance status")

	Required("user_id", "retrieved_at", "user_attributes")
})

var UserAccessPatterns = Type("UserAccessPatterns", func() {
	Description("User access patterns")

	Attribute("frequent_resources", ArrayOf(String), "Frequently accessed resources")
	Attribute("typical_hours", String, "Typical access hours", func() {
		Example("09:00-17:00")
	})
	Attribute("common_locations", ArrayOf(String), "Common access locations")
	Attribute("average_session_duration_minutes", UInt, "Average session duration")
	Attribute("last_30_days_activity", ActivityStats, "Recent activity statistics")
})

var ActivityStats = Type("ActivityStats", func() {
	Description("Activity statistics")

	Attribute("total_sessions", UInt, "Total sessions", func() {
		Example(45)
	})
	Attribute("total_actions", UInt, "Total actions", func() {
		Example(1247)
	})
	Attribute("unique_resources", UInt, "Unique resources accessed", func() {
		Example(89)
	})
	Attribute("policy_violations", UInt, "Policy violations", func() {
		Example(0)
	})
})

var UserRiskProfile = Type("UserRiskProfile", func() {
	Description("User risk profile")

	Attribute("overall_risk_score", UInt, "Overall risk score", func() {
		Example(25)
	})
	Attribute("risk_category", String, "Risk category", func() {
		Enum("low", "medium", "high", "critical")
		Example("low")
	})
	Attribute("risk_factors", ArrayOf(String), "Risk factors")
	Attribute("historical_incidents", UInt, "Historical incidents", func() {
		Example(0)
	})
	Attribute("last_risk_assessment", String, "Last risk assessment", func() {
		Format(FormatDateTime)
	})
})

var ComplianceStatus = Type("ComplianceStatus", func() {
	Description("User compliance status")

	Attribute("overall_status", String, "Overall compliance status", func() {
		Enum("compliant", "non_compliant", "requires_review", "pending")
		Example("compliant")
	})
	Attribute("training_status", String, "Training status", func() {
		Enum("current", "expired", "not_required", "pending")
		Example("current")
	})
	Attribute("certification_status", String, "Certification status", func() {
		Enum("valid", "expired", "not_required", "pending")
		Example("valid")
	})
	Attribute("last_compliance_check", String, "Last compliance check", func() {
		Format(FormatDateTime)
	})
	Attribute("next_review_date", String, "Next review date", func() {
		Format(FormatDateTime)
	})
})

// Validation Types
var AttributeValidationResult = Type("AttributeValidationResult", func() {
	Description("Attribute validation result")

	Attribute("user_id", String, "User ID", func() {
		Format(FormatUUID)
	})
	Attribute("validation_id", String, "Validation ID", func() {
		Example("validation_550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("validated_at", String, "Validation timestamp", func() {
		Format(FormatDateTime)
	})
	Attribute("overall_status", String, "Overall validation status", func() {
		Enum("valid", "invalid", "warnings", "incomplete")
		Example("valid")
	})
	Attribute("validation_results", ArrayOf(AttributeValidationDetail), "Detailed validation results")
	Attribute("compliance_check", ComplianceValidation, "Compliance validation")
	Attribute("recommendations", ArrayOf(String), "Validation recommendations")

	Required("user_id", "validation_id", "validated_at", "overall_status")
})

var AttributeValidationDetail = Type("AttributeValidationDetail", func() {
	Description("Detailed attribute validation")

	Attribute("attribute_name", String, "Attribute name", func() {
		Example("user.security_level")
	})
	Attribute("status", String, "Validation status", func() {
		Enum("valid", "invalid", "warning", "missing")
		Example("valid")
	})
	Attribute("message", String, "Validation message", func() {
		Example("Security level is within acceptable range")
	})
	Attribute("expected_value", String, "Expected value or format", func() {
		Example("1-10")
	})
	Attribute("actual_value", Any, "Actual value", func() {
		Example(7)
	})
	Attribute("severity", String, "Issue severity", func() {
		Enum("info", "warning", "error", "critical")
		Example("info")
	})
})

var ComplianceValidation = Type("ComplianceValidation", func() {
	Description("Compliance validation")

	Attribute("compliant", Boolean, "Whether user is compliant", func() {
		Example(true)
	})
	Attribute("compliance_frameworks", ArrayOf(String), "Applicable compliance frameworks", func() {
		Example([]string{"SOX", "PCI", "GDPR"})
	})
	Attribute("violations", ArrayOf(ComplianceViolation), "Compliance violations")
	Attribute("exemptions", ArrayOf(ComplianceExemption), "Compliance exemptions")
})

var ComplianceViolation = Type("ComplianceViolation", func() {
	Description("Compliance violation")

	Attribute("framework", String, "Compliance framework", func() {
		Example("SOX")
	})
	Attribute("rule", String, "Violated rule", func() {
		Example("SEGREGATION_OF_DUTIES")
	})
	Attribute("severity", String, "Violation severity", func() {
		Enum("low", "medium", "high", "critical")
		Example("medium")
	})
	Attribute("description", String, "Violation description", func() {
		Example("User has conflicting roles that violate segregation of duties")
	})
})

var ComplianceExemption = Type("ComplianceExemption", func() {
	Description("Compliance exemption")

	Attribute("framework", String, "Compliance framework", func() {
		Example("PCI")
	})
	Attribute("rule", String, "Exempted rule", func() {
		Example("ACCESS_LOGGING")
	})
	Attribute("reason", String, "Exemption reason", func() {
		Example("Emergency access approval")
	})
	Attribute("expires_at", String, "Exemption expiration", func() {
		Format(FormatDateTime)
	})
	Attribute("approved_by", String, "Approved by", func() {
		Example("compliance_officer_123")
	})
})

// Refresh Attributes Types
var RefreshAttributesResult = Type("RefreshAttributesResult", func() {
	Description("Attribute refresh result")

	Attribute("user_id", String, "User ID", func() {
		Format(FormatUUID)
	})
	Attribute("refresh_id", String, "Refresh operation ID", func() {
		Example("refresh_550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("started_at", String, "Refresh start time", func() {
		Format(FormatDateTime)
	})
	Attribute("completed_at", String, "Refresh completion time", func() {
		Format(FormatDateTime)
	})
	Attribute("sources_refreshed", ArrayOf(SourceRefreshResult), "Source refresh results")
	Attribute("attributes_updated", UInt, "Attributes updated", func() {
		Example(7)
	})
	Attribute("errors", ArrayOf(RefreshError), "Refresh errors")
	Attribute("cache_impact", CacheImpact, "Cache impact")

	Required("user_id", "refresh_id", "started_at", "completed_at", "sources_refreshed")
})

var SourceRefreshResult = Type("SourceRefreshResult", func() {
	Description("Individual source refresh result")

	Attribute("source", String, "Attribute source", func() {
		Example("hr_system")
	})
	Attribute("status", String, "Refresh status", func() {
		Enum("success", "failed", "partial", "skipped")
		Example("success")
	})
	Attribute("attributes_retrieved", UInt, "Attributes retrieved", func() {
		Example(5)
	})
	Attribute("refresh_time_ms", Float64, "Refresh time in milliseconds", func() {
		Example(234.5)
	})
	Attribute("last_successful_refresh", String, "Last successful refresh", func() {
		Format(FormatDateTime)
	})
})

var RefreshError = Type("RefreshError", func() {
	Description("Attribute refresh error")

	Attribute("source", String, "Error source", func() {
		Example("ldap_service")
	})
	Attribute("error_code", String, "Error code", func() {
		Example("CONNECTION_TIMEOUT")
	})
	Attribute("message", String, "Error message", func() {
		Example("Unable to connect to LDAP server")
	})
	Attribute("retry_after_seconds", UInt, "Retry after seconds", func() {
		Example(300)
	})
})
