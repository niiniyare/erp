package types

import (
	. "goa.design/goa/v3/dsl"
)

// PolicyEvaluationRequest describes a policy evaluation request
var PolicyEvaluationRequest = Type("PolicyEvaluationRequest", func() {
	Description("Request for ABAC policy evaluation")

	// Authentication and common headers
	JWTToken()
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("entity_id", String, "Entity identifier", func() {
		Format(FormatUUID)
		Example("987fcdeb-51d2-43b8-a456-426614174000")
	})
	Attribute("current_user_id", String, "Current user identifier", func() {
		Format(FormatUUID)
		Example("456e7890-e89b-12d3-a456-426614174000")
	})

	Attribute("user_id", String, "Subject user ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("resource_type", String, "Type of resource being accessed", func() {
		Example("document")
	})
	Attribute("resource_id", String, "Specific resource ID (optional)", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("action", String, "Action being performed", func() {
		Example("read")
	})
	Attribute("context", MapOf(String, Any), "Additional context attributes", func() {
		Example(map[string]interface{}{
			"ip_address":  "192.168.1.100",
			"user_agent":  "Mozilla/5.0...",
			"time_of_day": "business_hours",
			"department":  "engineering",
		})
	})
	Attribute("session_data", MapOf(String, Any), "Session-specific data", func() {
		Example(map[string]interface{}{
			"session_id":    "sess_123456789",
			"login_time":    "2023-12-07T08:00:00Z",
			"last_activity": "2023-12-07T10:30:00Z",
		})
	})
	Attribute("environment_data", MapOf(String, Any), "Environment data", func() {
		Example(map[string]interface{}{
			"ip_address":  "192.168.1.100",
			"location":    "US-CA",
			"device_type": "desktop",
		})
	})
	Attribute("request_id", String, "Request ID for tracking", func() {
		Example("req_abc123def456")
	})
	Attribute("use_cache", Boolean, "Whether to use cached results", func() {
		Default(true)
	})
	Attribute("cache_results", Boolean, "Whether to cache the results", func() {
		Default(true)
	})
	Attribute("include_advice", Boolean, "Include advice in response", func() {
		Default(false)
	})
	Attribute("explain_decision", Boolean, "Include decision explanation", func() {
		Default(false)
	})
	Required("token", "tenant_id", "entity_id", "current_user_id", "user_id", "resource_type", "action")
})

// PolicyEvaluationResponse describes a policy evaluation response
var PolicyEvaluationResponse = Type("PolicyEvaluationResponse", func() {
	Description("Response from ABAC policy evaluation")
	Attribute("decision", String, "Final access decision", func() {
		Enum("allow", "deny", "not_applicable")
		Example("allow")
	})
	Attribute("allowed", Boolean, "Whether access is allowed")
	Attribute("obligations", ArrayOf(PolicyObligation), "Actions that must be performed if access is granted")
	Attribute("advice", ArrayOf(PolicyAdvice), "Recommendations for the requesting system")
	Attribute("evaluation_time_ms", UInt, "Time taken to evaluate in milliseconds")
	Attribute("evaluated_at", String, "Timestamp of evaluation", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	Attribute("request_id", String, "Request ID for tracking")
	Attribute("cache_hit", Boolean, "Whether result came from cache")
	Attribute("policy_count", UInt, "Number of policies evaluated")
	Attribute("explanation", PolicyExplanation, "Decision explanation (if requested)")
	Attribute("audit_trail", DecisionAuditTrail, "Detailed audit information (if requested)")
	Required("decision", "allowed", "evaluation_time_ms", "evaluated_at", "cache_hit", "policy_count")
})

// BulkPolicyEvaluationRequest describes a bulk evaluation request
var BulkPolicyEvaluationRequest = Type("BulkPolicyEvaluationRequest", func() {
	Description("Request for bulk ABAC policy evaluation")

	// Authentication and common headers
	JWTToken()
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("entity_id", String, "Entity identifier", func() {
		Format(FormatUUID)
		Example("987fcdeb-51d2-43b8-a456-426614174000")
	})
	Attribute("current_user_id", String, "Current user identifier", func() {
		Format(FormatUUID)
		Example("456e7890-e89b-12d3-a456-426614174000")
	})

	Attribute("user_id", String, "Subject user ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("requests", ArrayOf(PolicyEvaluationRequest), "Individual evaluation requests", func() {
		MinLength(1)
		MaxLength(100)
	})
	Attribute("request_id", String, "Bulk request ID for tracking")
	Attribute("fail_fast", Boolean, "Stop on first error", func() {
		Default(false)
	})
	Attribute("use_cache", Boolean, "Whether to use cached results", func() {
		Default(true)
	})
	Attribute("cache_results", Boolean, "Whether to cache results", func() {
		Default(true)
	})
	Required("token", "tenant_id", "entity_id", "current_user_id", "user_id", "requests")
})

// BulkPolicyEvaluationResponse describes a bulk evaluation response
var BulkPolicyEvaluationResponse = Type("BulkPolicyEvaluationResponse", func() {
	Description("Response from bulk ABAC policy evaluation")
	Attribute("responses", ArrayOf(PolicyEvaluationResponse), "Individual evaluation responses")
	Attribute("success_count", UInt, "Number of successful evaluations")
	Attribute("error_count", UInt, "Number of failed evaluations")
	Attribute("total_requests", UInt, "Total number of requests processed")
	Attribute("evaluation_time_ms", UInt, "Total time taken for all evaluations")
	Attribute("request_id", String, "Bulk request ID")
	Attribute("partial_failure", Boolean, "Whether some requests failed")
	Required("responses", "success_count", "error_count", "total_requests", "evaluation_time_ms")
})

// AuthorizationRequest describes a simple authorization request
var AuthorizationRequest = Type("AuthorizationRequest", func() {
	Description("Simple authorization check request")

	// Authentication and common headers
	JWTToken()
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("entity_id", String, "Entity identifier", func() {
		Format(FormatUUID)
		Example("987fcdeb-51d2-43b8-a456-426614174000")
	})
	Attribute("current_user_id", String, "Current user identifier", func() {
		Format(FormatUUID)
		Example("456e7890-e89b-12d3-a456-426614174000")
	})

	Attribute("user_id", String, "Subject user ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("resource_type", String, "Type of resource being accessed", func() {
		Example("document")
	})
	Attribute("resource_id", String, "Specific resource ID (optional)", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("action", String, "Action being performed", func() {
		Example("read")
	})
	Attribute("context", MapOf(String, Any), "Additional context attributes")
	Required("token", "tenant_id", "entity_id", "current_user_id", "user_id", "resource_type", "action")
})

// AuthorizationResponse describes a simple authorization response
var AuthorizationResponse = Type("AuthorizationResponse", func() {
	Description("Simple authorization check response")
	Attribute("allowed", Boolean, "Whether access is allowed")
	Attribute("decision", String, "Access decision", func() {
		Enum("allow", "deny", "not_applicable")
	})
	Attribute("evaluation_time_ms", UInt, "Time taken to evaluate in milliseconds")
	Attribute("request_id", String, "Request ID for tracking")
	Required("allowed", "decision", "evaluation_time_ms")
})

// PolicyExplanationRequest describes a request for decision explanation
var PolicyExplanationRequest = Type("PolicyExplanationRequest", func() {
	Description("Request for policy decision explanation")

	// Authentication and common headers
	JWTToken()
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("entity_id", String, "Entity identifier", func() {
		Format(FormatUUID)
		Example("987fcdeb-51d2-43b8-a456-426614174000")
	})
	Attribute("current_user_id", String, "Current user identifier", func() {
		Format(FormatUUID)
		Example("456e7890-e89b-12d3-a456-426614174000")
	})

	// Copy PolicyEvaluationRequest fields
	Attribute("user_id", String, "Subject user ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("resource_type", String, "Type of resource being accessed", func() {
		Example("document")
	})
	Attribute("resource_id", String, "Specific resource ID (optional)", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("action", String, "Action being performed", func() {
		Example("read")
	})
	Attribute("context", MapOf(String, Any), "Additional context attributes", func() {
		Example(map[string]interface{}{
			"ip_address":  "192.168.1.100",
			"user_agent":  "Mozilla/5.0...",
			"time_of_day": "business_hours",
			"department":  "engineering",
		})
	})

	Attribute("detail_level", String, "Level of detail in explanation", func() {
		Enum("basic", "detailed", "verbose")
		Default("detailed")
		Example("detailed")
	})

	Required("token", "tenant_id", "entity_id", "current_user_id", "user_id", "resource_type", "action")
})

// PolicyExplanationResponse describes a decision explanation response
var PolicyExplanationResponse = Type("PolicyExplanationResponse", func() {
	Description("Detailed explanation of how a policy decision was made")
	Attribute("final_decision", String, "Final access decision", func() {
		Enum("allow", "deny", "not_applicable")
	})
	Attribute("reasoning_summary", String, "High-level explanation of the decision")
	Attribute("policy_evaluations", ArrayOf(PolicyEvaluationSummary), "Details of each policy evaluation")
	Attribute("attributes_used", MapOf(String, Any), "Attributes that influenced the decision")
	Attribute("conflict_resolution", ConflictResolutionSummary, "How conflicts were resolved (if any)")
	Attribute("combining_algorithm", String, "Algorithm used to combine policy decisions")
	Attribute("recommendations", ArrayOf(String), "Recommendations for policy improvements")
	Required("final_decision", "reasoning_summary", "policy_evaluations", "combining_algorithm")
})

// PolicyDiscoveryRequest describes a request to discover applicable policies
var PolicyDiscoveryRequest = Type("PolicyDiscoveryRequest", func() {
	Description("Request to discover applicable policies")

	// Authentication and common headers
	JWTToken()
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("entity_id", String, "Entity identifier", func() {
		Format(FormatUUID)
		Example("987fcdeb-51d2-43b8-a456-426614174000")
	})
	Attribute("current_user_id", String, "Current user identifier", func() {
		Format(FormatUUID)
		Example("456e7890-e89b-12d3-a456-426614174000")
	})

	Attribute("user_id", String, "Subject user ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("resource_type", String, "Type of resource", func() {
		Example("document")
	})
	Attribute("action", String, "Action being performed", func() {
		Example("read")
	})
	Attribute("context", MapOf(String, Any), "Additional context")
	Required("token", "tenant_id", "entity_id", "current_user_id", "user_id", "resource_type", "action")
})

// PolicyDiscoveryResponse describes policies discovery response
var PolicyDiscoveryResponse = Type("PolicyDiscoveryResponse", func() {
	Description("Response with applicable policies")
	Attribute("policies", ArrayOf(PolicySummary), "Applicable policies")
	Attribute("total_policies", UInt, "Total number of policies in system")
	Attribute("applicable_policies", UInt, "Number of applicable policies")
	Required("policies", "total_policies", "applicable_policies")
})

// AttributeCollectionRequest describes a request to collect attributes
var AttributeCollectionRequest = Type("AttributeCollectionRequest", func() {
	Description("Request to collect attributes for evaluation context")

	// Authentication and common headers
	JWTToken()
	Attribute("tenant_id", String, "Tenant identifier", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("entity_id", String, "Entity identifier", func() {
		Format(FormatUUID)
		Example("987fcdeb-51d2-43b8-a456-426614174000")
	})
	Attribute("current_user_id", String, "Current user identifier", func() {
		Format(FormatUUID)
		Example("456e7890-e89b-12d3-a456-426614174000")
	})

	Attribute("user_id", String, "Subject user ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("resource_type", String, "Type of resource", func() {
		Example("document")
	})
	Attribute("resource_id", String, "Specific resource ID (optional)", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("action", String, "Action being performed", func() {
		Example("read")
	})
	Attribute("include_expired", Boolean, "Include expired attributes", func() {
		Default(false)
	})
	Required("token", "tenant_id", "entity_id", "current_user_id", "user_id", "resource_type", "action")
})

// AttributeCollectionResponse describes attributes collection response
var AttributeCollectionResponse = Type("AttributeCollectionResponse", func() {
	Description("Response with collected attributes")
	Attribute("user_attributes", MapOf(String, AttributeValue), "User-related attributes")
	Attribute("resource_attributes", MapOf(String, AttributeValue), "Resource-related attributes")
	Attribute("environment_attributes", MapOf(String, AttributeValue), "Environment attributes")
	Attribute("action_attributes", MapOf(String, AttributeValue), "Action-related attributes")
	Attribute("entity_attributes", MapOf(String, AttributeValue), "Entity/Company attributes")
	Attribute("session_attributes", MapOf(String, AttributeValue), "Session attributes")
	Attribute("collection_time_ms", UInt, "Time taken to collect attributes")
	Attribute("total_attributes", UInt, "Total number of attributes collected")
	Required("collection_time_ms", "total_attributes")
})

// DecisionAuditResponse describes decision audit history response
var DecisionAuditResponse = Type("DecisionAuditResponse", func() {
	Description("Decision audit history response")
	Attribute("decisions", ArrayOf(DecisionAuditEntry), "Audit entries")
	Attribute("total_count", UInt, "Total number of audit entries")
	Attribute("pagination", PaginationMeta, "Pagination information")
	Required("decisions", "total_count")
})

// Supporting types

// PolicyObligation represents an obligation that must be fulfilled
var PolicyObligation = Type("PolicyObligation", func() {
	Description("Action that must be performed if access is granted")
	Attribute("id", String, "Obligation identifier", func() {
		Example("log_access")
	})
	Attribute("type", String, "Type of obligation", func() {
		Enum("log", "notify", "encrypt", "audit", "custom")
		Example("log")
	})
	Attribute("description", String, "Human-readable description", func() {
		Example("Log all access to this resource")
	})
	Attribute("parameters", MapOf(String, Any), "Obligation parameters", func() {
		Example(map[string]interface{}{
			"log_level":            "info",
			"include_user_details": true,
		})
	})
	Required("id", "type", "description")
})

// PolicyAdvice represents advice for the requesting system
var PolicyAdvice = Type("PolicyAdvice", func() {
	Description("Recommendation for the requesting system")
	Attribute("id", String, "Advice identifier", func() {
		Example("require_mfa")
	})
	Attribute("type", String, "Type of advice", func() {
		Enum("security", "performance", "compliance", "user_experience", "custom")
		Example("security")
	})
	Attribute("description", String, "Human-readable description", func() {
		Example("Consider requiring multi-factor authentication for this operation")
	})
	Attribute("severity", String, "Advice severity", func() {
		Enum("low", "medium", "high", "critical")
		Example("medium")
	})
	Attribute("parameters", MapOf(String, Any), "Advice parameters")
	Required("id", "type", "description", "severity")
})

// PolicyExplanation provides detailed decision explanation
var PolicyExplanation = Type("PolicyExplanation", func() {
	Description("Detailed explanation of policy decision")
	Attribute("final_decision", String, "Final decision", func() {
		Enum("allow", "deny", "not_applicable")
	})
	Attribute("reasoning_summary", String, "Summary of reasoning")
	Attribute("policy_evaluations", ArrayOf(PolicyEvaluationSummary), "Policy evaluation details")
	Attribute("attributes_used", MapOf(String, Any), "Attributes that influenced decision")
	Attribute("conflict_resolution", ConflictResolutionSummary, "Conflict resolution details")
	Attribute("combining_algorithm", String, "Combining algorithm used")
	Attribute("recommendations", ArrayOf(String), "Recommendations")
	Required("final_decision", "reasoning_summary", "combining_algorithm")
})

// PolicyEvaluationSummary summarizes a single policy evaluation
var PolicyEvaluationSummary = Type("PolicyEvaluationSummary", func() {
	Description("Summary of a single policy evaluation")
	Attribute("policy_name", String, "Name of the policy")
	Attribute("policy_id", String, "Policy identifier", func() {
		Format(FormatUUID)
	})
	Attribute("decision", String, "Policy decision", func() {
		Enum("allow", "deny", "not_applicable")
	})
	Attribute("applicable", Boolean, "Whether policy was applicable")
	Attribute("matched_rules", ArrayOf(String), "Rules that matched")
	Attribute("failed_rules", ArrayOf(String), "Rules that failed to match")
	Attribute("reason", String, "Reason for the decision")
	Required("policy_name", "decision", "applicable")
})

// ConflictResolutionSummary describes how conflicts were resolved
var ConflictResolutionSummary = Type("ConflictResolutionSummary", func() {
	Description("Details of conflict resolution")
	Attribute("conflict_detected", Boolean, "Whether conflicts were detected")
	Attribute("conflicting_policies", ArrayOf(String), "Names of conflicting policies")
	Attribute("resolution_method", String, "Method used to resolve conflicts")
	Attribute("winning_policy", String, "Policy that won the conflict resolution")
	Attribute("explanation", String, "Explanation of resolution")
	Required("conflict_detected", "resolution_method")
})

// PolicySummary provides a summary of a policy
var PolicySummary = Type("PolicySummary", func() {
	Description("Summary of a policy")
	Attribute("id", String, "Policy identifier", func() {
		Format(FormatUUID)
	})
	Attribute("name", String, "Policy name")
	Attribute("description", String, "Policy description")
	Attribute("effect", String, "Policy effect", func() {
		Enum("allow", "deny")
	})
	Attribute("priority", UInt, "Policy priority")
	Attribute("applicable", Boolean, "Whether policy is applicable")
	Required("id", "name", "effect", "applicable")
})

// AttributeValue represents an attribute value with metadata
var AttributeValue = Type("AttributeValue", func() {
	Description("Attribute value with metadata")
	Attribute("name", String, "Attribute name")
	Attribute("value", Any, "Attribute value")
	Attribute("data_type", String, "Data type", func() {
		Enum("string", "number", "boolean", "date", "json", "array", "enum")
	})
	Attribute("category", String, "Attribute category", func() {
		Enum("user", "resource", "environment", "action", "entity", "session")
	})
	Attribute("source", String, "Source of the attribute")
	Attribute("collected_at", String, "When attribute was collected", func() {
		Format(FormatDateTime)
	})
	Attribute("expires_at", String, "When attribute expires (optional)", func() {
		Format(FormatDateTime)
	})
	Required("name", "value", "data_type", "category", "collected_at")
})

// DecisionAuditTrail provides detailed audit information
var DecisionAuditTrail = Type("DecisionAuditTrail", func() {
	Description("Detailed audit trail for decision")
	Attribute("evaluation_steps", ArrayOf(EvaluationStep), "Steps in evaluation process")
	Attribute("policies_applied", ArrayOf(String), "Policies that were applied")
	Attribute("attributes_seen", MapOf(String, String), "Attributes examined during evaluation")
	Attribute("cache_events", ArrayOf(CacheEvent), "Cache-related events")
	Attribute("timing", EvaluationTiming, "Timing breakdown")
	Required("evaluation_steps", "policies_applied", "timing")
})

// EvaluationStep represents a step in the evaluation process
var EvaluationStep = Type("EvaluationStep", func() {
	Description("A step in the policy evaluation process")
	Attribute("step_type", String, "Type of step")
	Attribute("description", String, "Description of what happened")
	Attribute("result", Any, "Result of the step")
	Attribute("duration_ms", UInt, "Duration in milliseconds")
	Attribute("metadata", MapOf(String, Any), "Additional metadata")
	Required("step_type", "description", "duration_ms")
})

// CacheEvent represents a cache-related event
var CacheEvent = Type("CacheEvent", func() {
	Description("Cache-related event during evaluation")
	Attribute("event_type", String, "Type of cache event", func() {
		Enum("hit", "miss", "store", "invalidate")
	})
	Attribute("cache_key", String, "Cache key involved")
	Attribute("hit", Boolean, "Whether it was a cache hit")
	Attribute("timestamp", String, "Event timestamp", func() {
		Format(FormatDateTime)
	})
	Required("event_type", "timestamp")
})

// EvaluationTiming provides timing breakdown
var EvaluationTiming = Type("EvaluationTiming", func() {
	Description("Timing breakdown for evaluation")
	Attribute("attribute_collection_ms", UInt, "Time for attribute collection")
	Attribute("attribute_validation_ms", UInt, "Time for attribute validation")
	Attribute("policy_retrieval_ms", UInt, "Time for policy retrieval")
	Attribute("policy_evaluation_ms", UInt, "Time for policy evaluation")
	Attribute("cache_operation_ms", UInt, "Time for cache operations")
	Attribute("total_ms", UInt, "Total evaluation time")
	Required("total_ms")
})

// DecisionAuditEntry represents a single audit entry
var DecisionAuditEntry = Type("DecisionAuditEntry", func() {
	Description("Single decision audit entry")
	Attribute("id", String, "Audit entry ID", func() {
		Format(FormatUUID)
	})
	Attribute("user_id", String, "User who made the request", func() {
		Format(FormatUUID)
	})
	Attribute("resource_type", String, "Type of resource accessed")
	Attribute("resource_id", String, "Specific resource ID (optional)", func() {
		Format(FormatUUID)
	})
	Attribute("action", String, "Action that was attempted")
	Attribute("decision", String, "Access decision", func() {
		Enum("allow", "deny", "not_applicable")
	})
	Attribute("allowed", Boolean, "Whether access was allowed")
	Attribute("evaluation_time_ms", UInt, "Time taken for evaluation")
	Attribute("evaluated_at", String, "When the decision was made", func() {
		Format(FormatDateTime)
	})
	Attribute("policy_count", UInt, "Number of policies evaluated")
	Attribute("cache_hit", Boolean, "Whether result came from cache")
	Attribute("request_id", String, "Request ID for tracking")
	AuditFields()
	Required("id", "user_id", "resource_type", "action", "decision", "allowed", "evaluated_at")
})

// Health and metrics types

// ComponentHealth describes health of a service component
var ComponentHealth = Type("ComponentHealth", func() {
	Description("Health status of a service component")
	Attribute("status", String, "Component status", func() {
		Enum("healthy", "degraded", "unhealthy")
	})
	Attribute("message", String, "Status message")
	Attribute("last_check", String, "Last health check time", func() {
		Format(FormatDateTime)
	})
	Required("status")
})

// EvaluationMetrics describes policy evaluation metrics
var EvaluationMetrics = Type("EvaluationMetrics", func() {
	Description("Policy evaluation performance metrics")
	Attribute("total_evaluations", UInt, "Total number of evaluations")
	Attribute("successful_evaluations", UInt, "Number of successful evaluations")
	Attribute("failed_evaluations", UInt, "Number of failed evaluations")
	Attribute("average_evaluation_time_ms", UInt, "Average evaluation time in milliseconds")
	Attribute("evaluations_per_second", UInt, "Current evaluations per second")
	Required("total_evaluations", "successful_evaluations", "failed_evaluations")
})

// CacheMetrics describes cache performance metrics
var CacheMetrics = Type("CacheMetrics", func() {
	Description("Cache performance metrics")
	Attribute("total_requests", UInt, "Total cache requests")
	Attribute("cache_hits", UInt, "Number of cache hits")
	Attribute("cache_misses", UInt, "Number of cache misses")
	Attribute("hit_rate", Float64, "Cache hit rate (0.0 to 1.0)")
	Attribute("average_retrieval_time_ms", UInt, "Average cache retrieval time")
	Attribute("total_entries", UInt, "Total number of cached entries")
	Required("total_requests", "cache_hits", "cache_misses", "hit_rate")
})

// AttributeMetrics describes attribute collection metrics
var AttributeMetrics = Type("AttributeMetrics", func() {
	Description("Attribute collection performance metrics")
	Attribute("total_collections", UInt, "Total attribute collections")
	Attribute("successful_collections", UInt, "Successful collections")
	Attribute("failed_collections", UInt, "Failed collections")
	Attribute("average_collection_time_ms", UInt, "Average collection time")
	Attribute("attributes_per_collection", Float64, "Average attributes per collection")
	Required("total_collections", "successful_collections", "failed_collections")
})
