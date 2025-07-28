package types

import (
	. "goa.design/goa/v3/dsl"
)

var PolicyEvaluationRequest = Type("PolicyEvaluationRequest", func() {
	Attribute("user_id", String, "User ID")
	Attribute("resource_type", String, "Resource type")
	Attribute("resource_id", String, "Resource ID")
	Attribute("action", String, "Action")
	Attribute("context", MapOf(String, Any), "Context")
	Attribute("use_cache", Boolean, "Use cache")
	Attribute("cache_results", Boolean, "Cache results")
	Attribute("include_advice", Boolean, "Include advice")
	Attribute("explain_decision", Boolean, "Explain decision")
	Attribute("request_id", String, "Request ID")
})

var PolicyEvaluationResponse = Type("PolicyEvaluationResponse", func() {
	Attribute("decision", String, "Decision")
	Attribute("allowed", Boolean, "Allowed")
	Attribute("evaluation_time_ms", UInt, "Evaluation time in ms")
	Attribute("evaluated_at", String, "Evaluation timestamp", func() { Format(FormatDateTime) })
	Attribute("request_id", String, "Request ID")
	Attribute("cache_hit", Boolean, "Cache hit")
	Attribute("policy_count", UInt, "Policy count")
	Attribute("obligations", ArrayOf(PolicyObligation), "Obligations")
	Attribute("advice", ArrayOf(PolicyAdvice), "Advice")
	Attribute("explanation", PolicyExplanation, "Explanation")
	Attribute("audit_trail", DecisionAuditTrail, "Audit trail")
})

var BulkPolicyEvaluationRequest = Type("BulkPolicyEvaluationRequest", func() {
	Attribute("user_id", String, "User ID")
	Attribute("requests", ArrayOf(PolicyEvaluationRequest), "Requests")
	Attribute("request_id", String, "Request ID")
	Attribute("fail_fast", Boolean, "Fail fast")
	Attribute("use_cache", Boolean, "Use cache")
	Attribute("cache_results", Boolean, "Cache results")
})

var BulkPolicyEvaluationResponse = Type("BulkPolicyEvaluationResponse", func() {
	Attribute("success_count", UInt, "Success count")
	Attribute("error_count", UInt, "Error count")
	Attribute("total_requests", UInt, "Total requests")
	Attribute("evaluation_time_ms", UInt, "Evaluation time in ms")
	Attribute("request_id", String, "Request ID")
	Attribute("partial_failure", Boolean, "Partial failure")
	Attribute("responses", ArrayOf(PolicyEvaluationResponse), "Responses")
})

var AuthorizationRequest = Type("AuthorizationRequest", func() {
	Attribute("user_id", String, "User ID")
	Attribute("resource_type", String, "Resource type")
	Attribute("resource_id", String, "Resource ID")
	Attribute("action", String, "Action")
	Attribute("context", MapOf(String, Any), "Context")
})

var AuthorizationResponse = Type("AuthorizationResponse", func() {
	Attribute("allowed", Boolean, "Allowed")
	Attribute("decision", String, "Decision")
	Attribute("evaluation_time_ms", UInt, "Evaluation time in ms")
	Attribute("request_id", String, "Request ID")
})

var PolicyExplanationRequest = Type("PolicyExplanationRequest", func() {
	Attribute("user_id", String, "User ID")
	Attribute("resource_type", String, "Resource type")
	Attribute("resource_id", String, "Resource ID")
	Attribute("action", String, "Action")
	Attribute("context", MapOf(String, Any), "Context")
	Attribute("detail_level", String, "Detail level")
})

var PolicyExplanationResponse = Type("PolicyExplanationResponse", func() {
	Attribute("final_decision", String, "Final decision")
	Attribute("reasoning_summary", String, "Reasoning summary")
	Attribute("combining_algorithm", String, "Combining algorithm")
	Attribute("attributes_used", MapOf(String, Any), "Attributes used")
	Attribute("recommendations", ArrayOf(String), "Recommendations")
	Attribute("policy_evaluations", ArrayOf(PolicyEvaluationSummary), "Policy evaluations")
	Attribute("conflict_resolution", ConflictResolutionSummary, "Conflict resolution")
})

var PolicyDiscoveryRequest = Type("PolicyDiscoveryRequest", func() {
	Attribute("user_id", String, "User ID")
	Attribute("resource_type", String, "Resource type")
	Attribute("action", String, "Action")
	Attribute("context", MapOf(String, Any), "Context")
})

var PolicyDiscoveryResponse = Type("PolicyDiscoveryResponse", func() {
	Attribute("policies", ArrayOf(PolicySummary), "Policies")
	Attribute("total_policies", UInt, "Total policies")
	Attribute("applicable_policies", UInt, "Applicable policies")
})

var AttributeCollectionRequest = Type("AttributeCollectionRequest", func() {
	Attribute("user_id", String, "User ID")
	Attribute("resource_type", String, "Resource type")
	Attribute("resource_id", String, "Resource ID")
	Attribute("action", String, "Action")
	Attribute("entity_id", String, "Entity ID")
	Attribute("include_expired", Boolean, "Include expired")
})

var AttributeCollectionResponse = Type("AttributeCollectionResponse", func() {
	Attribute("collection_time_ms", UInt, "Collection time in ms")
	Attribute("total_attributes", UInt, "Total attributes")
	Attribute("user_attributes", MapOf(String, AttributeValue), "User attributes")
	Attribute("resource_attributes", MapOf(String, AttributeValue), "Resource attributes")
	Attribute("environment_attributes", MapOf(String, AttributeValue), "Environment attributes")
	Attribute("action_attributes", MapOf(String, AttributeValue), "Action attributes")
	Attribute("entity_attributes", MapOf(String, AttributeValue), "Entity attributes")
	Attribute("session_attributes", MapOf(String, AttributeValue), "Session attributes")
})

var AuditDecisionsPayload = Type("AuditDecisionsPayload", func() {
	Attribute("user_id", String, "User ID")
	Attribute("limit", UInt, "Limit")
})

var DecisionAuditResponse = Type("DecisionAuditResponse", func() {
	Attribute("decisions", ArrayOf(DecisionAuditEntry), "Decisions")
	Attribute("total_count", UInt, "Total count")
})

var InvalidateCachePayload = Type("InvalidateCachePayload", func() {
	Attribute("user_id", String, "User ID")
	Attribute("resource_type", String, "Resource type")
	Attribute("pattern", String, "Pattern")
})

var InvalidateCacheResult = Type("InvalidateCacheResult", func() {
	Attribute("invalidated_count", UInt, "Invalidated count")
	Attribute("success", Boolean, "Success")
})

var HealthResult = Type("HealthResult", func() {
	Attribute("status", String, "Status")
	Attribute("timestamp", String, "Timestamp", func() { Format(FormatDateTime) })
	Attribute("version", String, "Version")
	Attribute("components", MapOf(String, ComponentHealth), "Components")
})

var MetricsResult = Type("MetricsResult", func() {
	Attribute("evaluation_metrics", EvaluationMetrics, "Evaluation metrics")
	Attribute("cache_metrics", CacheMetrics, "Cache metrics")
	Attribute("attribute_metrics", AttributeMetrics, "Attribute metrics")
})

var PolicyObligation = Type("PolicyObligation", func() {
	Attribute("id", String, "ID")
	Attribute("type", String, "Type")
	Attribute("description", String, "Description")
	Attribute("parameters", MapOf(String, Any), "Parameters")
})

var PolicyAdvice = Type("PolicyAdvice", func() {
	Attribute("id", String, "ID")
	Attribute("type", String, "Type")
	Attribute("description", String, "Description")
	Attribute("severity", String, "Severity")
	Attribute("parameters", MapOf(String, Any), "Parameters")
})

var PolicyExplanation = Type("PolicyExplanation", func() {
	Attribute("final_decision", String, "Final decision")
	Attribute("reasoning_summary", String, "Reasoning summary")
	Attribute("combining_algorithm", String, "Combining algorithm")
	Attribute("attributes_used", MapOf(String, Any), "Attributes used")
	Attribute("recommendations", ArrayOf(String), "Recommendations")
	Attribute("policy_evaluations", ArrayOf(PolicyEvaluationSummary), "Policy evaluations")
	Attribute("conflict_resolution", ConflictResolutionSummary, "Conflict resolution")
})

var DecisionAuditTrail = Type("DecisionAuditTrail", func() {
	Attribute("policies_applied", ArrayOf(String), "Policies applied")
	Attribute("attributes_seen", MapOf(String, Any), "Attributes seen")
	Attribute("evaluation_steps", ArrayOf(EvaluationStep), "Evaluation steps")
	Attribute("cache_events", ArrayOf(CacheEvent), "Cache events")
	Attribute("timing", EvaluationTiming, "Timing")
})

var PolicyEvaluationSummary = Type("PolicyEvaluationSummary", func() {
	Attribute("policy_name", String, "Policy name")
	Attribute("decision", String, "Decision")
	Attribute("applicable", Boolean, "Applicable")
	Attribute("matched_rules", ArrayOf(String), "Matched rules")
	Attribute("failed_rules", ArrayOf(String), "Failed rules")
	Attribute("reason", String, "Reason")
})

var ConflictResolutionSummary = Type("ConflictResolutionSummary", func() {
	Attribute("conflict_detected", Boolean, "Conflict detected")
	Attribute("conflicting_policies", ArrayOf(String), "Conflicting policies")
	Attribute("resolution_method", String, "Resolution method")
	Attribute("winning_policy", String, "Winning policy")
	Attribute("explanation", String, "Explanation")
})

var PolicySummary = Type("PolicySummary", func() {
	Attribute("id", String, "ID")
	Attribute("name", String, "Name")
	Attribute("description", String, "Description")
	Attribute("effect", String, "Effect")
	Attribute("priority", UInt, "Priority")
	Attribute("applicable", Boolean, "Applicable")
})

var AttributeValue = Type("AttributeValue", func() {
	Attribute("name", String, "Name")
	Attribute("value", Any, "Value")
	Attribute("data_type", String, "Data type")
	Attribute("category", String, "Category")
	Attribute("source", String, "Source")
	Attribute("collected_at", String, "Collected at", func() { Format(FormatDateTime) })
	Attribute("expires_at", String, "Expires at", func() { Format(FormatDateTime) })
})

var DecisionAuditEntry = Type("DecisionAuditEntry", func() {
	Attribute("id", String, "ID")
	Attribute("user_id", String, "User ID")
	Attribute("resource_type", String, "Resource type")
	Attribute("resource_id", String, "Resource ID")
	Attribute("action", String, "Action")
	Attribute("decision", String, "Decision")
	Attribute("allowed", Boolean, "Allowed")
	Attribute("evaluation_time_ms", UInt, "Evaluation time in ms")
	Attribute("evaluated_at", String, "Evaluated at", func() { Format(FormatDateTime) })
	Attribute("policy_count", UInt, "Policy count")
	Attribute("cache_hit", Boolean, "Cache hit")
	Attribute("request_id", String, "Request ID")
	Attribute("created_at", String, "Created at", func() { Format(FormatDateTime) })
	Attribute("updated_at", String, "Updated at", func() { Format(FormatDateTime) })
	Attribute("created_by", String, "Created by")
	Attribute("updated_by", String, "Updated by")
})

var ComponentHealth = Type("ComponentHealth", func() {
	Attribute("status", String, "Status")
})

var EvaluationMetrics = Type("EvaluationMetrics", func() {
	Attribute("total_evaluations", UInt64, "Total evaluations")
	Attribute("successful_evaluations", UInt64, "Successful evaluations")
	Attribute("failed_evaluations", UInt64, "Failed evaluations")
	Attribute("average_evaluation_time_ms", Float64, "Average evaluation time in ms")
	Attribute("evaluations_per_second", Float64, "Evaluations per second")
})

var CacheMetrics = Type("CacheMetrics", func() {
	Attribute("total_requests", UInt64, "Total requests")
	Attribute("cache_hits", UInt64, "Cache hits")
	Attribute("cache_misses", UInt64, "Cache misses")
	Attribute("hit_rate", Float64, "Hit rate")
	Attribute("average_retrieval_time_ms", Float64, "Average retrieval time in ms")
	Attribute("total_entries", UInt64, "Total entries")
})

var AttributeMetrics = Type("AttributeMetrics", func() {
	Attribute("total_collections", UInt64, "Total collections")
	Attribute("successful_collections", UInt64, "Successful collections")
	Attribute("failed_collections", UInt64, "Failed collections")
	Attribute("average_collection_time_ms", Float64, "Average collection time in ms")
	Attribute("attributes_per_collection", Float64, "Attributes per collection")
})

var EvaluationStep = Type("EvaluationStep", func() {
	Attribute("step_type", String, "Step type")
	Attribute("description", String, "Description")
	Attribute("result", String, "Result")
	Attribute("duration_ms", UInt, "Duration in ms")
	Attribute("metadata", MapOf(String, Any), "Metadata")
})

var CacheEvent = Type("CacheEvent", func() {
	Attribute("event_type", String, "Event type")
	Attribute("cache_key", String, "Cache key")
	Attribute("hit", Boolean, "Hit")
	Attribute("timestamp", String, "Timestamp", func() { Format(FormatDateTime) })
})

var EvaluationTiming = Type("EvaluationTiming", func() {
	Attribute("attribute_collection_ms", UInt, "Attribute collection in ms")
	Attribute("attribute_validation_ms", UInt, "Attribute validation in ms")
	Attribute("policy_retrieval_ms", UInt, "Policy retrieval in ms")
	Attribute("policy_evaluation_ms", UInt, "Policy evaluation in ms")
	Attribute("cache_operation_ms", UInt, "Cache operation in ms")
	Attribute("total_ms", UInt, "Total in ms")
})
