package featureflag

import (
	"github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// Service describes the feature flag management service
var _ = Service("featureflag", func() {
	Description("Feature flag management service for controlled feature releases")

	// Apply global middleware
	HTTP(func() {
		Path("/api/v1/feature-flags")
	})

	// Create feature flag endpoint
	Method("create", func() {
		Description("Create a new feature flag")

		Payload(CreateFeatureFlagPayload)
		Result(FeatureFlagResult)

		Error("bad_request")
		Error("conflict") // For name conflicts
		Error("unauthorized")
		Error("unprocessable_entity")

		HTTP(func() {
			POST("/")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("conflict", StatusConflict)
			Response("unauthorized", StatusUnauthorized)
			Response("unprocessable_entity", StatusUnprocessableEntity)
		})
	})

	// Get feature flag by name
	Method("get", func() {
		Description("Get feature flag by name")

		Payload(func() {
			Attribute("name", String, "Feature flag name", func() {
				Pattern("^[a-zA-Z0-9_-]+$")
				MinLength(1)
				MaxLength(255)
				Example("enhanced_dashboard")
			})
			Required("name")
		})

		Result(FeatureFlagResult)

		Error("not_found")
		Error("unauthorized")

		HTTP(func() {
			GET("/{name}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	// Get feature flag by ID
	Method("getById", func() {
		Description("Get feature flag by ID")

		Payload(func() {
			Attribute("id", String, "Feature flag ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Required("id")
		})

		Result(FeatureFlagResult)

		Error("not_found")
		Error("unauthorized")

		HTTP(func() {
			GET("/by-id/{id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	// List feature flags with pagination
	Method("list", func() {
		Description("List feature flags with pagination and filtering")

		Payload(func() {
			Extend(types.Pagination)
			Attribute("flag_type", String, "Filter by flag type", func() {
				Enum("boolean", "string", "number", "json")
				Example("boolean")
			})
			Attribute("name_filter", String, "Filter by flag name", func() {
				Example("dashboard")
			})
			Attribute("enabled_only", Boolean, "Show only enabled flags", func() {
				Example(true)
			})
		})

		Result(func() {
			Attribute("data", ArrayOf(FeatureFlagResult), "The feature flags")
			Attribute("pagination", types.PaginationMeta, "Pagination metadata")
			Required("data", "pagination")
		})

		Error("bad_request")
		Error("unauthorized")

		HTTP(func() {
			GET("/")
			Param("page")
			Param("page_size")
			Param("sort_by")
			Param("sort_order")
			Param("flag_type")
			Param("name_filter")
			Param("enabled_only")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	// Update feature flag
	Method("update", func() {
		Description("Update an existing feature flag")

		Payload(UpdateFeatureFlagPayload)
		Result(FeatureFlagResult)

		Error("bad_request")
		Error("not_found")
		Error("conflict")
		Error("unauthorized")
		Error("unprocessable_entity")

		HTTP(func() {
			PUT("/{id}")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("conflict", StatusConflict)
			Response("unauthorized", StatusUnauthorized)
			Response("unprocessable_entity", StatusUnprocessableEntity)
		})
	})

	// Delete feature flag
	Method("delete", func() {
		Description("Delete a feature flag (soft delete)")

		Payload(func() {
			Attribute("id", String, "Feature flag ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Required("id")
		})

		Result(Empty) // No content response

		Error("not_found")
		Error("unauthorized")
		Error("conflict") // If flag is in use

		HTTP(func() {
			DELETE("/{id}")
			Response(StatusNoContent)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("conflict", StatusConflict)
		})
	})

	// Evaluate single feature flag
	Method("evaluate", func() {
		Description("Evaluate a feature flag for a specific context")

		Payload(EvaluateFeatureFlagPayload)
		Result(EvaluationResult)

		Error("bad_request")
		Error("not_found")
		Error("unauthorized")

		HTTP(func() {
			POST("/evaluate/{name}")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	// Bulk evaluate multiple feature flags
	Method("evaluateMultiple", func() {
		Description("Evaluate multiple feature flags for a specific context")

		Payload(BulkEvaluatePayload)
		Result(func() {
			Attribute("results", MapOf(String, EvaluationResult), "Evaluation results by flag name")
			Attribute("context", EvaluationContext, "The evaluation context used")
			Required("results", "context")
		})

		Error("bad_request")
		Error("unauthorized")

		HTTP(func() {
			POST("/evaluate")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	// Get feature flag statistics
	Method("getStats", func() {
		Description("Get feature flag usage statistics")

		Result(FeatureFlagStats)

		Error("unauthorized")

		HTTP(func() {
			GET("/stats")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	// Search feature flags
	Method("search", func() {
		Description("Search feature flags by name or description")

		Payload(func() {
			Attribute("query", String, "Search query", func() {
				MinLength(1)
				MaxLength(100)
				Example("dashboard")
			})
			Attribute("limit", UInt, "Maximum number of results", func() {
				Default(10)
				Maximum(100)
				Example(10)
			})
			Attribute("offset", UInt, "Number of results to skip", func() {
				Default(0)
				Example(0)
			})
			Required("query")
		})

		Result(func() {
			Attribute("data", ArrayOf(FeatureFlagResult), "Search results")
			Attribute("total_count", UInt, "Total number of matching flags")
			Attribute("query", String, "The search query used")
			Required("data", "total_count", "query")
		})

		Error("bad_request")
		Error("unauthorized")

		HTTP(func() {
			GET("/search")
			Param("query")
			Param("limit")
			Param("offset")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	// Get flags by type
	Method("getByType", func() {
		Description("Get all feature flags of a specific type")

		Payload(func() {
			Attribute("flag_type", String, "Flag type", func() {
				Enum("boolean", "string", "number", "json")
				Example("boolean")
			})
			Required("flag_type")
		})

		Result(func() {
			Attribute("data", ArrayOf(FeatureFlagResult), "Feature flags of the specified type")
			Attribute("flag_type", String, "The flag type requested")
			Attribute("count", UInt, "Number of flags returned")
			Required("data", "flag_type", "count")
		})

		Error("bad_request")
		Error("unauthorized")

		HTTP(func() {
			GET("/type/{flag_type}")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	// Health check endpoint
	Method("health", func() {
		Description("Health check for feature flag service")

		Result(func() {
			Attribute("status", String, "Service status", func() {
				Enum("healthy", "degraded", "unhealthy")
				Example("healthy")
			})
			Attribute("timestamp", String, "Check timestamp", func() {
				Format(FormatDateTime)
				Example("2025-01-08T10:30:00Z")
			})
			Attribute("version", String, "Service version", func() {
				Example("1.0.0")
			})
			Attribute("database_status", String, "Database connection status", func() {
				Enum("connected", "disconnected", "degraded")
				Example("connected")
			})
			Attribute("cache_status", String, "Cache status", func() {
				Enum("available", "unavailable", "degraded")
				Example("available")
			})
			Required("status", "timestamp", "version")
		})

		HTTP(func() {
			GET("/health")
			Response(StatusOK)
		})

		// No authentication required for health checks
		NoSecurity()
	})
})
