package featureflag

import (
	"github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// FeatureFlagResult describes the feature flag response
var FeatureFlagResult = ResultType("application/vnd.featureflag", func() {
	Description("Feature flag information")
	Attributes(func() {
		Attribute("id", String, "Unique feature flag identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
		})
		Attribute("tenant_id", String, "Tenant identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440001")
		})
		Attribute("name", String, "Feature flag name", func() {
			Pattern("^[a-zA-Z0-9_-]+$")
			MinLength(1)
			MaxLength(255)
			Example("enhanced_dashboard")
		})
		Attribute("description", String, "Feature flag description", func() {
			MaxLength(1000)
			Example("Enable the new enhanced dashboard interface with improved analytics")
		})
		Attribute("flag_type", String, "Type of the flag", func() {
			Enum("boolean", "string", "number", "json")
			Example("boolean")
		})
		Attribute("default_value", Boolean, "Default value for the flag", func() {
			Example(false)
		})
		Attribute("rollout_percentage", UInt, "Percentage rollout (0-100)", func() {
			Maximum(100)
			Example(25)
		})
		Attribute("target_audience", MapOf(String, Any), "Target audience configuration", func() {
			Example(map[string]interface{}{
				"user_roles":           []string{"admin", "manager"},
				"min_account_age_days": 30,
				"beta_users":           true,
			})
		})
		Attribute("metadata", MapOf(String, Any), "Additional metadata", func() {
			Example(map[string]interface{}{
				"category":    "ui",
				"owner":       "frontend-team",
				"jira_ticket": "DASH-123",
				"launch_date": "2025-02-01",
			})
		})
		types.AuditFields()
	})
	Required("id", "tenant_id", "name", "description", "flag_type", "default_value", "created_at", "updated_at")

	View("default", func() {
		Attribute("id")
		Attribute("tenant_id")
		Attribute("name")
		Attribute("description")
		Attribute("flag_type")
		Attribute("default_value")
		Attribute("rollout_percentage")
		Attribute("target_audience")
		Attribute("metadata")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})

	View("minimal", func() {
		Attribute("id")
		Attribute("name")
		Attribute("flag_type")
		Attribute("default_value")
		Attribute("rollout_percentage")
	})

	View("evaluation", func() {
		Attribute("name")
		Attribute("flag_type")
		Attribute("default_value")
		Attribute("rollout_percentage")
		Attribute("target_audience")
	})
})

// CreateFeatureFlagPayload describes the payload for creating a feature flag
var CreateFeatureFlagPayload = Type("CreateFeatureFlagPayload", func() {
	Description("Payload for creating a new feature flag")
	Attribute("name", String, "Feature flag name", func() {
		Pattern("^[a-zA-Z0-9_-]+$")
		MinLength(1)
		MaxLength(255)
		Example("enhanced_dashboard")
	})
	Attribute("description", String, "Feature flag description", func() {
		MaxLength(1000)
		Example("Enable the new enhanced dashboard interface with improved analytics")
	})
	Attribute("flag_type", String, "Type of the flag", func() {
		Enum("boolean", "string", "number", "json")
		Default("boolean")
		Example("boolean")
	})
	Attribute("default_value", Boolean, "Default value for the flag", func() {
		Default(false)
		Example(false)
	})
	Attribute("rollout_percentage", UInt, "Percentage rollout (0-100)", func() {
		Maximum(100)
		Example(25)
	})
	Attribute("target_audience", MapOf(String, Any), "Target audience configuration", func() {
		Example(map[string]interface{}{
			"user_roles":           []string{"admin", "manager"},
			"min_account_age_days": 30,
			"beta_users":           true,
		})
	})
	Attribute("metadata", MapOf(String, Any), "Additional metadata", func() {
		Example(map[string]interface{}{
			"category":    "ui",
			"owner":       "frontend-team",
			"jira_ticket": "DASH-123",
			"launch_date": "2025-02-01",
		})
	})
	Required("name", "description")
})

// UpdateFeatureFlagPayload describes the payload for updating a feature flag
var UpdateFeatureFlagPayload = Type("UpdateFeatureFlagPayload", func() {
	Description("Payload for updating an existing feature flag")
	Attribute("id", String, "Feature flag ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("description", String, "Feature flag description", func() {
		MaxLength(1000)
		Example("Updated description for the enhanced dashboard")
	})
	Attribute("default_value", Boolean, "Default value for the flag", func() {
		Example(true)
	})
	Attribute("rollout_percentage", UInt, "Percentage rollout (0-100)", func() {
		Maximum(100)
		Example(50)
	})
	Attribute("target_audience", MapOf(String, Any), "Target audience configuration", func() {
		Example(map[string]interface{}{
			"user_roles":           []string{"admin", "manager", "user"},
			"min_account_age_days": 7,
			"beta_users":           true,
		})
	})
	Attribute("metadata", MapOf(String, Any), "Additional metadata", func() {
		Example(map[string]interface{}{
			"category":        "ui",
			"owner":           "frontend-team",
			"jira_ticket":     "DASH-123",
			"rollout_date":    "2025-02-15",
			"success_metrics": []string{"engagement", "conversion"},
		})
	})
	Required("id")
})

// EvaluationContext describes the context for flag evaluation
var EvaluationContext = Type("EvaluationContext", func() {
	Description("Context information for feature flag evaluation")
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440001")
	})
	Attribute("user_id", String, "User identifier", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440002")
	})
	Attribute("environment", String, "Environment name", func() {
		Enum("development", "staging", "production")
		Example("production")
	})
	Attribute("attributes", MapOf(String, String), "Additional context attributes", func() {
		Example(map[string]string{
			"user_role":         "admin",
			"subscription_plan": "enterprise",
			"user_agent":        "Mozilla/5.0",
			"country":           "US",
		})
	})
	Attribute("client_info", ClientInfo, "Client application information")
	Required("tenant_id", "environment")
})

// ClientInfo describes client application details
var ClientInfo = Type("ClientInfo", func() {
	Description("Client application information")
	Attribute("version", String, "Client version", func() {
		Example("1.2.3")
	})
	Attribute("platform", String, "Client platform", func() {
		Enum("web", "mobile", "desktop", "api")
		Example("web")
	})
	Attribute("ip_address", String, "Client IP address", func() {
		Format(FormatIP)
		Example("192.168.1.100")
	})
	Attribute("user_agent", String, "User agent string", func() {
		Example("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	})
})

// EvaluateFeatureFlagPayload describes the payload for evaluating a single flag
var EvaluateFeatureFlagPayload = Type("EvaluateFeatureFlagPayload", func() {
	Description("Payload for evaluating a feature flag")
	Attribute("name", String, "Feature flag name", func() {
		Pattern("^[a-zA-Z0-9_-]+$")
		MinLength(1)
		MaxLength(255)
		Example("enhanced_dashboard")
	})
	Attribute("context", EvaluationContext, "Evaluation context")
	Required("name", "context")
})

// BulkEvaluatePayload describes the payload for bulk evaluation
var BulkEvaluatePayload = Type("BulkEvaluatePayload", func() {
	Description("Payload for evaluating multiple feature flags")
	Attribute("flag_names", ArrayOf(String), "List of feature flag names to evaluate", func() {
		Example([]string{"enhanced_dashboard", "new_checkout", "advanced_analytics"})
		MinLength(1)
		MaxLength(50)
	})
	Attribute("context", EvaluationContext, "Evaluation context")
	Required("flag_names", "context")
})

// EvaluationResult describes the result of flag evaluation
var EvaluationResult = Type("EvaluationResult", func() {
	Description("Result of feature flag evaluation")
	Attribute("flag_name", String, "Feature flag name", func() {
		Example("enhanced_dashboard")
	})
	Attribute("value", Boolean, "Evaluated flag value", func() {
		Example(true)
	})
	Attribute("enabled", Boolean, "Whether the flag is enabled for this context", func() {
		Example(true)
	})
	Attribute("reason", String, "Reason for the evaluation result", func() {
		Enum("default_value", "percentage_rollout", "target_audience", "override", "disabled")
		Example("percentage_rollout")
	})
	Attribute("rule_matched", String, "Name of the rule that matched (if any)", func() {
		Example("beta_users_rule")
	})
	Attribute("metadata", ResultMetadata, "Evaluation metadata")
	Required("flag_name", "value", "enabled", "reason")
})

// ResultMetadata provides additional information about the evaluation
var ResultMetadata = Type("ResultMetadata", func() {
	Description("Metadata about the flag evaluation")
	Attribute("evaluated_at", String, "Timestamp when evaluation occurred", func() {
		Format(FormatDateTime)
		Example("2025-01-08T10:30:00Z")
	})
	Attribute("cache_hit", Boolean, "Whether the result came from cache", func() {
		Example(false)
	})
	Attribute("evaluation_time_ms", Float64, "Time taken for evaluation in milliseconds", func() {
		Example(0.5)
	})
	Attribute("config_version", String, "Version of the flag configuration used", func() {
		Example("v1.2.3")
	})
	Attribute("tenant_id", String, "Tenant ID for this evaluation", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440001")
	})
	Required("evaluated_at", "cache_hit", "evaluation_time_ms")
})

// FeatureFlagStats describes feature flag usage statistics
var FeatureFlagStats = Type("FeatureFlagStats", func() {
	Description("Feature flag usage statistics")
	Attribute("total_flags", UInt64, "Total number of feature flags", func() {
		Example(25)
	})
	Attribute("enabled_flags", UInt64, "Number of enabled flags", func() {
		Example(18)
	})
	Attribute("rollout_flags", UInt64, "Number of flags with percentage rollout", func() {
		Example(12)
	})
	Attribute("avg_rollout_percentage", Float64, "Average rollout percentage", func() {
		Example(42.5)
	})
	Attribute("flags_by_type", MapOf(String, UInt64), "Flag count by type", func() {
		Example(map[string]interface{}{
			"boolean": 20,
			"string":  3,
			"number":  1,
			"json":    1,
		})
	})
	Attribute("recent_activity", FeatureFlagActivitySummary, "Recent activity summary")
	Required("total_flags", "enabled_flags", "rollout_flags", "avg_rollout_percentage")
})

// FeatureFlagActivitySummary describes recent feature flag activity
var FeatureFlagActivitySummary = Type("FeatureFlagActivitySummary", func() {
	Description("Summary of recent feature flag activity")
	Attribute("flags_created_today", UInt, "Flags created today", func() {
		Example(2)
	})
	Attribute("flags_updated_today", UInt, "Flags updated today", func() {
		Example(5)
	})
	Attribute("flags_evaluated_today", UInt64, "Total evaluations today", func() {
		Example(15420)
	})
	Attribute("most_evaluated_flag", String, "Most frequently evaluated flag", func() {
		Example("enhanced_dashboard")
	})
	Attribute("last_activity_at", String, "Timestamp of last activity", func() {
		Format(FormatDateTime)
		Example("2025-01-08T10:25:00Z")
	})
})
