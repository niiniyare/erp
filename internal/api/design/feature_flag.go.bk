package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// FEATURE FLAG & CONFIGURATION MANAGEMENT TYPES
// ============================================================================

// FeatureFlag represents a feature toggle for controlling application functionality.
// Supports complex targeting rules, gradual rollouts, and A/B testing capabilities.
var FeatureFlag = ResultType("application/vnd.erp.feature_flag", func() {
	Description("Feature flag for controlling application functionality with advanced targeting, gradual rollouts, and A/B testing")
	
	Attributes(func() {
		Field(1, "id", String, "Unique feature flag identifier", func() {
			Format(FormatUUID)
			Example("flag-123e4567-e89b-12d3-a456-426614174000")
			Description("Primary key for feature flag")
		})
		
		Field(2, "tenant_id", String, "Associated tenant identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
			Description("Tenant isolation boundary")
		})
		
		Field(3, "key", String, "Unique feature flag key", func() {
			Pattern("^[a-z][a-z0-9_]{2,99}$")
			MinLength(3)
			MaxLength(100)
			Example("enhanced_financial_dashboard")
			Description("Machine-readable flag identifier")
		})
		
		Field(4, "name", String, "Human-readable flag name", func() {
			MinLength(3)
			MaxLength(150)
			Example("Enhanced Financial Dashboard")
			Description("User-friendly name for UI display")
		})
		
		Field(5, "description", String, "Feature flag description", func() {
			MaxLength(500)
			Example("Enables the new enhanced financial dashboard with real-time analytics and improved visualization")
			Description("Detailed explanation of what this flag controls")
		})
		
		Field(6, "flag_type", String, "Type of feature flag", func() {
			Enum("BOOLEAN", "STRING", "NUMBER", "JSON", "PERCENTAGE")
			Default("BOOLEAN")
			Example("BOOLEAN")
			Description("Data type returned by the flag")
		})
		
		Field(7, "category", String, "Feature category", func() {
			Enum("EXPERIMENT", "ROLLOUT", "OPERATIONAL", "PERMISSION", "KILL_SWITCH", "CONFIG")
			Default("ROLLOUT")
			Example("EXPERIMENT")
			Description("Classification of flag purpose")
		})
		
		Field(8, "environment", String, "Target environment", func() {
			Enum("DEVELOPMENT", "STAGING", "PRODUCTION", "ALL")
			Default("ALL")
			Example("PRODUCTION")
			Description("Environment where flag is active")
		})
		
		Field(9, "is_enabled", Boolean, "Global flag status", func() {
			Default(false)
			Example(true)
			Description("Master on/off switch for the feature")
		})
		
		Field(10, "default_value", Any, "Default flag value", func() {
			Example(false)
			Description("Value returned when no targeting rules match")
		})
		
		Field(11, "targeting_rules", ArrayOf(Type("TargetingRule", func() {
			Field(1, "rule_id", String, "Rule identifier", func() {
				Format(FormatUUID)
				Example("rule-456e7890-e89b-12d3-a456-426614174000")
			})
			Field(2, "name", String, "Rule name", func() {
				MaxLength(100)
				Example("Beta Users Rule")
			})
			Field(3, "description", String, "Rule description", func() {
				MaxLength(300)
				Example("Enable feature for users in beta testing group")
			})
			Field(4, "priority", UInt, "Rule evaluation priority", func() {
				Minimum(1)
				Maximum(1000)
				Example(100)
				Description("Higher numbers evaluated first")
			})
			Field(5, "conditions", ArrayOf(Type("RuleCondition", func() {
				Field(1, "attribute", String, "Attribute to evaluate", func() {
					Example("user_role")
				})
				Field(2, "operator", String, "Comparison operator", func() {
					Enum("EQUALS", "NOT_EQUALS", "IN", "NOT_IN", "CONTAINS", "STARTS_WITH", 
						"ENDS_WITH", "GREATER_THAN", "LESS_THAN", "REGEX_MATCH")
					Example("IN")
				})
				Field(3, "values", ArrayOf(Any), "Values to compare against", func() {
					Example([]any{"beta_tester", "admin", "power_user"})
				})
				Field(4, "attribute_type", String, "Data type of attribute", func() {
					Enum("STRING", "NUMBER", "BOOLEAN", "DATE", "VERSION")
					Example("STRING")
				})
				Required("attribute", "operator", "values", "attribute_type")
			})), "Conditions that must be met", func() {
				Description("All conditions must be true for rule to match")
			})
			Field(6, "rollout_percentage", Float64, "Percentage rollout", func() {
				Minimum(0.0)
				Maximum(100.0)
				Example(25.0)
				Description("Percentage of matching users to receive feature")
			})
			Field(7, "value", Any, "Value to return when rule matches", func() {
				Example(true)
			})
			Field(8, "is_active", Boolean, "Rule active status", func() {
				Default(true)
				Example(true)
			})
			Required("rule_id", "name", "priority", "conditions", "value", "is_active")
		})), "Targeting rules for conditional flag evaluation", func() {
			Description("Rules evaluated in priority order to determine flag value")
		})
		
		Field(12, "audience_filters", MapOf(String, Any), "Audience segmentation filters", func() {
			Example(map[string]any{
				"user_segments": []string{"premium", "enterprise"},
				"geo_regions":  []string{"US", "CA", "EU"},
				"device_types": []string{"desktop", "mobile"},
				"user_attributes": map[string]any{
					"account_age_days": map[string]any{
						"min": 30,
						"max": 365,
					},
					"subscription_tier": []string{"pro", "enterprise"},
				},
			})
			Description("Advanced audience targeting and segmentation")
		})
		
		Field(13, "percentage_rollout", Type("PercentageRollout", func() {
			Field(1, "enabled", Boolean, "Percentage rollout enabled", func() {
				Default(false)
			})
			Field(2, "percentage", Float64, "Current rollout percentage", func() {
				Minimum(0.0)
				Maximum(100.0)
				Default(0.0)
				Example(15.0)
			})
			Field(3, "sticky", Boolean, "Sticky rollout", func() {
				Default(true)
				Description("Whether users stay in same bucket across sessions")
			})
			Field(4, "bucketing_key", String, "Attribute for consistent bucketing", func() {
				Default("user_id")
				Example("user_id")
			})
			Field(5, "rollout_schedule", ArrayOf(Type("RolloutStep", func() {
				Field(1, "target_percentage", Float64, "Target percentage", func() {
					Minimum(0.0)
					Maximum(100.0)
				})
				Field(2, "scheduled_date", String, "Scheduled date", func() {
					Format(FormatDateTime)
					Example("2023-12-15T00:00:00Z")
				})
				Field(3, "status", String, "Step status", func() {
					Enum("PENDING", "ACTIVE", "COMPLETED", "CANCELLED")
					Default("PENDING")
				})
				Required("target_percentage", "scheduled_date", "status")
			})), "Automated rollout schedule", func() {
				Description("Planned percentage increases over time")
			})
			Required("enabled", "percentage", "sticky", "bucketing_key")
		}), "Percentage-based feature rollout configuration", func() {
			Description("Gradual rollout to percentage of users")
		})
		
		Field(14, "variants", ArrayOf(Type("FeatureVariant", func() {
			Field(1, "variant_id", String, "Variant identifier", func() {
				Format(FormatUUID)
			})
			Field(2, "name", String, "Variant name", func() {
				MaxLength(50)
				Example("Control")
			})
			Field(3, "description", String, "Variant description", func() {
				MaxLength(200)
				Example("Original dashboard layout")
			})
			Field(4, "value", Any, "Variant value", func() {
				Example(false)
			})
			Field(5, "weight", Float64, "Traffic allocation weight", func() {
				Minimum(0.0)
				Maximum(100.0)
				Example(50.0)
			})
			Field(6, "is_control", Boolean, "Control variant flag", func() {
				Default(false)
			})
			Required("variant_id", "name", "value", "weight")
		})), "A/B testing variants", func() {
			Description("Different values for A/B testing experiments")
		})
		
		Field(15, "metrics", ArrayOf(Type("FlagMetric", func() {
			Field(1, "metric_name", String, "Metric name", func() {
				Example("conversion_rate")
			})
			Field(2, "metric_type", String, "Type of metric", func() {
				Enum("COUNTER", "GAUGE", "TIMER", "HISTOGRAM")
				Example("GAUGE")
			})
			Field(3, "description", String, "Metric description", func() {
				Example("Percentage of users who complete purchase flow")
			})
			Field(4, "target_value", Float64, "Target metric value", func() {
				Example(0.25)
			})
			Field(5, "current_value", Float64, "Current metric value", func() {
				Example(0.18)
			})
			Required("metric_name", "metric_type", "description")
		})), "Success metrics to track", func() {
			Description("Key performance indicators for feature success")
		})
		
		Field(16, "prerequisites", ArrayOf(String), "Required prerequisite flags", func() {
			Example([]string{"user_authentication_v2", "database_migration_complete"})
			Description("Other flags that must be enabled")
		})
		
		Field(17, "dependencies", ArrayOf(String), "Dependent feature flags", func() {
			Example([]string{"advanced_reporting", "real_time_notifications"})
			Description("Flags that depend on this flag being enabled")
		})
		
		Field(18, "kill_switch", Type("KillSwitch", func() {
			Field(1, "enabled", Boolean, "Kill switch enabled", func() {
				Default(false)
			})
			Field(2, "trigger_conditions", ArrayOf(String), "Conditions that trigger kill switch", func() {
				Example([]string{"error_rate > 5%", "cpu_usage > 90%", "memory_usage > 85%"})
			})
			Field(3, "auto_disable", Boolean, "Automatically disable on trigger", func() {
				Default(true)
			})
			Field(4, "notification_channels", ArrayOf(String), "Alert notification channels", func() {
				Example([]string{"slack", "pagerduty", "email"})
			})
			Required("enabled", "auto_disable")
		}), "Emergency kill switch configuration", func() {
			Description("Automatic feature disabling on error conditions")
		})
		
		Field(19, "evaluation_context", MapOf(String, Any), "Default evaluation context", func() {
			Example(map[string]any{
				"application_version": "2.1.0",
				"deployment_environment": "production",
				"region": "us-east-1",
			})
			Description("Default attributes available for rule evaluation")
		})
		
		Field(20, "usage_stats", Type("FlagUsageStats", func() {
			Field(1, "total_evaluations", UInt, "Total flag evaluations", func() {
				Example(150247)
			})
			Field(2, "unique_users", UInt, "Unique users evaluated", func() {
				Example(12456)
			})
			Field(3, "true_evaluations", UInt, "Evaluations returning true", func() {
				Example(37562)
			})
			Field(4, "false_evaluations", UInt, "Evaluations returning false", func() {
				Example(112685)
			})
			Field(5, "last_evaluation_at", String, "Most recent evaluation", func() {
				Format(FormatDateTime)
				Example("2023-12-07T14:30:00Z")
			})
			Field(6, "evaluation_latency_ms", Float64, "Average evaluation time", func() {
				Example(2.5)
			})
		}), "Flag usage statistics", func() {
			Description("Performance and usage metrics")
		})
		
		Field(21, "schedule", Type("FlagSchedule", func() {
			Field(1, "enabled", Boolean, "Scheduled activation enabled", func() {
				Default(false)
			})
			Field(2, "start_date", String, "Scheduled activation date", func() {
				Format(FormatDateTime)
				Example("2023-12-15T00:00:00Z")
			})
			Field(3, "end_date", String, "Scheduled deactivation date", func() {
				Format(FormatDateTime)
				Example("2024-01-31T23:59:59Z")
			})
			Field(4, "timezone", String, "Schedule timezone", func() {
				Pattern("^[A-Za-z]+/[A-Za-z_]+$")
				Default("UTC")
				Example("America/New_York")
			})
			Field(5, "recurring", Boolean, "Recurring schedule", func() {
				Default(false)
			})
			Field(6, "recurrence_pattern", String, "Recurrence pattern", func() {
				Example("0 9 * * MON-FRI")
				Description("Cron expression for recurring activation")
			})
			Required("enabled")
		}), "Scheduled activation and deactivation", func() {
			Description("Time-based flag activation schedule")
		})
		
		Field(22, "tags", ArrayOf(String), "Flag classification tags", func() {
			Example([]string{"ui", "dashboard", "analytics", "beta"})
			Description("Searchable tags for flag organization")
		})
		
		Field(23, "owner_team", String, "Team responsible for flag", func() {
			Example("product-team")
			Description("Team that owns and manages this flag")
		})
		
		Field(24, "contact_email", String, "Contact email for flag issues", func() {
			Format(FormatEmail)
			Example("product-team@company.com")
			Description("Email for flag-related questions or issues")
		})
		
		Field(25, "documentation_url", String, "Link to feature documentation", func() {
			Format(FormatURI)
			Example("https://docs.company.com/features/enhanced-dashboard")
			Description("External documentation or specification")
		})
		
		Field(26, "is_permanent", Boolean, "Permanent flag indicator", func() {
			Default(false)
			Example(false)
			Description("Whether flag is intended to be permanent")
		})
		
		Field(27, "removal_date", String, "Planned removal date", func() {
			Format(FormatDate)
			Example("2024-06-30")
			Description("When flag should be removed from codebase")
		})
		
		Field(28, "archived", Boolean, "Flag archived status", func() {
			Default(false)
			Example(false)
			Description("Whether flag has been archived")
		})
		
		Field(29, "archived_at", String, "Archive timestamp", func() {
			Format(FormatDateTime)
			Example("2023-12-31T23:59:59Z")
			Description("When flag was archived")
		})
		
		// Audit fields from common.go
		AuditFields()
		
		Required("id", "tenant_id", "key", "name", "flag_type", "category", 
			"environment", "is_enabled", "default_value", "created_at")
	})
	
	View("default", func() {
		Description("Standard feature flag view for general management")
		Attribute("id")
		Attribute("key")
		Attribute("name")
		Attribute("flag_type")
		Attribute("category")
		Attribute("environment")
		Attribute("is_enabled")
		Attribute("default_value")
		Attribute("owner_team")
		Attribute("created_at")
	})
	
	View("detailed", func() {
		Description("Complete feature flag view with all configuration")
		Attribute("id")
		Attribute("tenant_id")
		Attribute("key")
		Attribute("name")
		Attribute("description")
		Attribute("flag_type")
		Attribute("category")
		Attribute("environment")
		Attribute("is_enabled")
		Attribute("default_value")
		Attribute("targeting_rules")
		Attribute("audience_filters")
		Attribute("percentage_rollout")
		Attribute("variants")
		Attribute("metrics")
		Attribute("prerequisites")
		Attribute("dependencies")
		Attribute("kill_switch")
		Attribute("evaluation_context")
		Attribute("usage_stats")
		Attribute("schedule")
		Attribute("tags")
		Attribute("owner_team")
		Attribute("contact_email")
		Attribute("documentation_url")
		Attribute("is_permanent")
		Attribute("removal_date")
		Attribute("archived")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})
	
	View("evaluation", func() {
		Description("Optimized view for flag evaluation engines")
		Attribute("id")
		Attribute("key")
		Attribute("flag_type")
		Attribute("is_enabled")
		Attribute("default_value")
		Attribute("targeting_rules")
		Attribute("percentage_rollout")
		Attribute("variants")
		Attribute("prerequisites")
		Attribute("kill_switch")
		Attribute("evaluation_context")
	})
	
	View("analytics", func() {
		Description("Analytics view for performance monitoring")
		Attribute("id")
		Attribute("key")
		Attribute("name")
		Attribute("category")
		Attribute("metrics")
		Attribute("usage_stats")
		Attribute("variants")
		Attribute("percentage_rollout")
		Attribute("created_at")
	})
	
	View("summary", func() {
		Description("Minimal view for quick references and listings")
		Attribute("id")
		Attribute("key")
		Attribute("name")
		Attribute("flag_type")
		Attribute("is_enabled")
		Attribute("category")
	})
})

// FlagEvaluation represents the result of a feature flag evaluation.
var FlagEvaluation = Type("FlagEvaluation", func() {
	Description("Result of feature flag evaluation with context and reasoning")
	
	Field(1, "flag_key", String, "Evaluated flag key", func() {
		Pattern("^[a-z][a-z0-9_]{2,99}$")
		Example("enhanced_financial_dashboard")
		Description("Feature flag that was evaluated")
	})
	
	Field(2, "user_id", String, "User context for evaluation", func() {
		Format(FormatUUID)
		Example("user-456e7890-e89b-12d3-a456-426614174000")
		Description("User for whom flag was evaluated")
	})
	
	Field(3, "value", Any, "Evaluated flag value", func() {
		Example(true)
		Description("Final value returned by flag evaluation")
	})
	
	Field(4, "variant_id", String, "Selected variant (for A/B tests)", func() {
		Format(FormatUUID)
		Example("variant-123e4567-e89b-12d3-a456-426614174000")
		Description("Specific variant selected during evaluation")
	})
	
	Field(5, "reason", String, "Evaluation reasoning", func() {
		Enum("DEFAULT_VALUE", "TARGETING_RULE", "PERCENTAGE_ROLLOUT", 
			"VARIANT_ASSIGNMENT", "KILL_SWITCH", "PREREQUISITE_FAILED")
		Example("TARGETING_RULE")
		Description("Why this value was returned")
	})
	
	Field(6, "rule_id", String, "Matching targeting rule", func() {
		Format(FormatUUID)
		Example("rule-789abc12-def3-4567-890a-bcdef1234567")
		Description("Rule that determined the flag value")
	})
	
	Field(7, "evaluation_context", MapOf(String, Any), "Context used for evaluation", func() {
		Example(map[string]any{
			"user_role":     "beta_tester",
			"account_tier":  "premium",
			"geo_location":  "US",
			"device_type":   "desktop",
		})
		Description("User and environment attributes")
	})
	
	Field(8, "timestamp", String, "Evaluation timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T14:30:00Z")
		Description("When evaluation occurred")
	})
	
	Field(9, "evaluation_time_ms", Float64, "Evaluation processing time", func() {
		Example(1.5)
		Description("Time taken to evaluate flag in milliseconds")
	})
	
	Field(10, "cache_hit", Boolean, "Result served from cache", func() {
		Default(false)
		Example(true)
		Description("Whether result came from cache")
	})
	
	Required("flag_key", "value", "reason", "evaluation_context", "timestamp")
})