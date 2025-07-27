package abac

import (
	. "github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// Service describes the ABAC (Attribute-Based Access Control) service
var _ = Service("abac", func() {
	Description("ABAC service for fine-grained access control and policy evaluation")

	HTTP(func() {
		Path("/api/v1/abac")
	})

	// Security requirements
	Security("jwt", func() {
		Scope("api:read")
		Scope("api:write")
	})

	// Policy Decision Point (PDP) endpoint
	Method("evaluate", func() {
		Description("Evaluate access decision using ABAC policies")

		Payload(PolicyEvaluationRequest)
		Result(PolicyEvaluationResponse)

		Error("bad_request")
		Error("unauthorized")
		Error("forbidden")
		Error("internal_error")

		HTTP(func() {
			POST("/evaluate")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// Bulk policy evaluation
	Method("evaluate_bulk", func() {
		Description("Evaluate multiple access decisions in a single request")

		Payload(BulkPolicyEvaluationRequest)
		Result(BulkPolicyEvaluationResponse)

		Error("bad_request")
		Error("unauthorized")
		Error("internal_error")

		HTTP(func() {
			POST("/evaluate/bulk")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// Simple authorization check
	Method("authorize", func() {
		Description("Simple authorization check - returns boolean result")

		Payload(AuthorizationRequest)
		Result(AuthorizationResponse)

		Error("bad_request")
		Error("unauthorized")
		Error("internal_error")

		HTTP(func() {
			POST("/authorize")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// Decision explanation endpoint
	Method("explain", func() {
		Description("Get detailed explanation of how a decision was made")

		Payload(PolicyExplanationRequest)
		Result(PolicyExplanationResponse)

		Error("bad_request")
		Error("unauthorized")
		Error("internal_error")

		HTTP(func() {
			POST("/explain")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// Policy discovery
	Method("discover_policies", func() {
		Description("Discover applicable policies for a given request")

		Payload(PolicyDiscoveryRequest)
		Result(PolicyDiscoveryResponse)

		Error("bad_request")
		Error("unauthorized")
		Error("internal_error")

		HTTP(func() {
			POST("/policies/discover")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// Attribute collection endpoint
	Method("collect_attributes", func() {
		Description("Collect attributes for a given context")

		Payload(AttributeCollectionRequest)
		Result(AttributeCollectionResponse)

		Error("bad_request")
		Error("unauthorized")
		Error("internal_error")

		HTTP(func() {
			POST("/attributes/collect")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// Decision audit log
	Method("audit_decisions", func() {
		Description("Get decision audit history for a user")

		Payload(func() {
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

			Attribute("user_id", String, "User ID to get audit history for", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Attribute("limit", UInt, "Maximum number of records to return", func() {
				Default(100)
				Maximum(1000)
				Example(50)
			})
			Attribute("start_date", String, "Start date for audit filter", func() {
				Format(FormatDateTime)
				Example("2023-12-01T00:00:00Z")
			})
			Attribute("end_date", String, "End date for audit filter", func() {
				Format(FormatDateTime)
				Example("2023-12-07T23:59:59Z")
			})
			Required("token", "tenant_id", "entity_id", "current_user_id", "user_id")
		})

		Result(DecisionAuditResponse)

		Error("bad_request")
		Error("unauthorized")
		Error("not_found")
		Error("internal_error")

		HTTP(func() {
			GET("/audit/decisions/{user_id}")
			Param("limit")
			Param("start_date")
			Param("end_date")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("not_found", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// Cache management
	Method("invalidate_cache", func() {
		Description("Invalidate cached policy evaluations")

		Payload(func() {
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

			Attribute("user_id", String, "User ID to invalidate cache for", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Attribute("resource_type", String, "Resource type to invalidate", func() {
				Example("document")
			})
			Attribute("pattern", String, "Cache key pattern to invalidate", func() {
				Example("policy_eval:*")
			})
			Required("token", "tenant_id", "entity_id", "current_user_id")
		})

		Result(func() {
			Attribute("invalidated_count", UInt, "Number of cache entries invalidated")
			Attribute("success", Boolean, "Whether the operation was successful")
			Required("invalidated_count", "success")
		})

		Error("bad_request")
		Error("unauthorized")
		Error("internal_error")

		HTTP(func() {
			POST("/cache/invalidate")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("internal_error", StatusInternalServerError)
		})

		// Only admin users can invalidate cache
		Security("jwt", func() {
			Scope("admin")
		})
	})

	// Health check
	Method("health", func() {
		Description("Health check for ABAC service")

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
				Example("1.0.0")
			})
			Attribute("components", MapOf(String, ComponentHealth), "Component health status")
			Required("status", "timestamp", "version")
		})

		HTTP(func() {
			GET("/health")
			Response(StatusOK)
		})

		// No authentication required for health checks
		NoSecurity()
	})

	// Metrics endpoint
	Method("metrics", func() {
		Description("Get ABAC service metrics")

		Payload(func() {
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
			Required("token", "tenant_id", "entity_id", "current_user_id")
		})

		Result(func() {
			Attribute("evaluation_metrics", EvaluationMetrics, "Policy evaluation metrics")
			Attribute("cache_metrics", CacheMetrics, "Cache performance metrics")
			Attribute("attribute_metrics", AttributeMetrics, "Attribute collection metrics")
			Required("evaluation_metrics", "cache_metrics", "attribute_metrics")
		})

		HTTP(func() {
			GET("/metrics")
			Response(StatusOK)
		})

		// Only admin users can view metrics
		Security("jwt", func() {
			Scope("admin")
		})
	})
})
