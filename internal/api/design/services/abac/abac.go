package abac

import (
	"github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

var _ = Service("abac", func() {
	Description("Attribute-Based Access Control (ABAC) service for managing policies and evaluating permissions.")

	Security("jwt")

	Error("unauthorized", String, "Unauthorized access")
	Error("forbidden", String, "Forbidden access")
	Error("not_found", String, "Resource not found")
	Error("bad_request", String, "Bad request")

	Method("evaluate", func() {
		Description("Evaluate a policy decision request.")
		Payload(func() {
			Token("token", String, "JWT token")
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
		Result(types.PolicyEvaluationResponse)
		HTTP(func() {
			POST("/api/v1/abac/evaluate")
			Response(StatusOK)
		})
	})

	Method("evaluate_bulk", func() {
		Description("Evaluate a bulk policy decision request.")
		Payload(func() {
			Token("token", String, "JWT token")
			Attribute("user_id", String, "User ID")
			Attribute("requests", ArrayOf(types.PolicyEvaluationRequest), "Requests")
			Attribute("request_id", String, "Request ID")
			Attribute("fail_fast", Boolean, "Fail fast")
			Attribute("use_cache", Boolean, "Use cache")
			Attribute("cache_results", Boolean, "Cache results")
		})
		Result(types.BulkPolicyEvaluationResponse)
		HTTP(func() {
			POST("/api/v1/abac/evaluate-bulk")
			Response(StatusOK)
		})
	})

	Method("authorize", func() {
		Description("Simple authorization check.")
		Payload(func() {
			Token("token", String, "JWT token")
			Attribute("user_id", String, "User ID")
			Attribute("resource_type", String, "Resource type")
			Attribute("resource_id", String, "Resource ID")
			Attribute("action", String, "Action")
			Attribute("context", MapOf(String, Any), "Context")
		})
		Result(types.AuthorizationResponse)
		HTTP(func() {
			POST("/api/v1/abac/authorize")
			Response(StatusOK)
		})
	})

	Method("explain", func() {
		Description("Explain a policy decision.")
		Payload(func() {
			Token("token", String, "JWT token")
			Attribute("user_id", String, "User ID")
			Attribute("resource_type", String, "Resource type")
			Attribute("resource_id", String, "Resource ID")
			Attribute("action", String, "Action")
			Attribute("context", MapOf(String, Any), "Context")
			Attribute("detail_level", String, "Detail level")
		})
		Result(types.PolicyExplanationResponse)
		HTTP(func() {
			POST("/api/v1/abac/explain")
			Response(StatusOK)
		})
	})

	Method("discover_policies", func() {
		Description("Discover applicable policies.")
		Payload(func() {
			Token("token", String, "JWT token")
			Attribute("user_id", String, "User ID")
			Attribute("resource_type", String, "Resource type")
			Attribute("action", String, "Action")
			Attribute("context", MapOf(String, Any), "Context")
		})
		Result(types.PolicyDiscoveryResponse)
		HTTP(func() {
			POST("/api/v1/abac/discover-policies")
			Response(StatusOK)
		})
	})

	Method("collect_attributes", func() {
		Description("Collect attributes for a given context.")
		Payload(func() {
			Token("token", String, "JWT token")
			Attribute("user_id", String, "User ID")
			Attribute("resource_type", String, "Resource type")
			Attribute("resource_id", String, "Resource ID")
			Attribute("action", String, "Action")
			Attribute("entity_id", String, "Entity ID")
			Attribute("include_expired", Boolean, "Include expired")
		})
		Result(types.AttributeCollectionResponse)
		HTTP(func() {
			POST("/api/v1/abac/collect-attributes")
			Response(StatusOK)
		})
	})

	Method("audit_decisions", func() {
		Description("Get a history of policy decisions.")
		Payload(func() {
			Token("token", String, "JWT token")
			Attribute("user_id", String, "User ID")
			Attribute("limit", UInt, "Limit")
		})
		Result(types.DecisionAuditResponse)
		HTTP(func() {
			GET("/api/v1/abac/audit")
			Response(StatusOK)
		})
	})

	Method("invalidate_cache", func() {
		Description("Invalidate the ABAC cache.")
		Payload(func() {
			Token("token", String, "JWT token")
			Attribute("user_id", String, "User ID")
			Attribute("resource_type", String, "Resource type")
			Attribute("pattern", String, "Pattern")
		})
		Result(types.InvalidateCacheResult)
		HTTP(func() {
			POST("/api/v1/abac/invalidate-cache")
			Response(StatusOK)
		})
	})

	Method("health", func() {
		Description("Health check for the ABAC service.")
		Payload(func() {
			Token("token", String, "JWT token")
		})
		Result(types.HealthResult)
		HTTP(func() {
			GET("/api/v1/abac/health")
			Response(StatusOK)
		})
	})

	Method("metrics", func() {
		Description("Get performance metrics for the ABAC service.")
		Payload(func() {
			Token("token", String, "JWT token")
		})
		Result(types.MetricsResult)
		HTTP(func() {
			GET("/api/v1/abac/metrics")
			Response(StatusOK)
		})
	})
})
