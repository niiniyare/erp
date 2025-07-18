package organization

import (
	"github.com/niiniyare/erp/design/types"
	. "goa.design/goa/v3/dsl"
)

// Service describes the organization management service
var _ = Service("organization", func() {
	Description("Organization and entity management service")

	// Apply global middleware
	HTTP(func() {
		Path("/api/v1/organizations")
	})

	// Security requirements
	Security("jwt", func() {
		Scope("api:read", "api:write")
	})

	// Create organization endpoint
	Method("create", func() {
		Description("Create a new organization")

		Payload(func() {
			Attribute("name", String, "Organization name", func() {
				MinLength(1)
				MaxLength(100)
				Example("Acme Corporation")
			})
			Attribute("description", String, "Organization description", func() {
				MaxLength(500)
				Example("Leading provider of roadrunner traps and anvils")
			})
			Attribute("tax_id", String, "Tax identification number", func() {
				Example("12-3456789")
			})
			Attribute("industry", String, "Industry sector", func() {
				Example("Manufacturing")
			})
			Attribute("website", String, "Organization website", func() {
				Format(FormatURI)
				Example("https://acme.com")
			})
			Attribute("contact", types.ContactInfo, "Primary contact information")
			Attribute("address", OrganizationAddress, "Organization address")
			Attribute("tenant_id", String, "Tenant identifier", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Required("name", "tenant_id")
		})

		Result(OrganizationResult)

		Error("bad_request")
		Error("conflict")
		Error("unauthorized")
		Error("forbidden")
		Error("unprocessable_entity")

		HTTP(func() {
			POST("/")
			types.CommonHeaders()
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("conflict", StatusConflict)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("unprocessable_entity", StatusUnprocessableEntity)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("conflict", CodeAlreadyExists)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("unprocessable_entity", CodeInvalidArgument)
		})
	})

	// Get organization by ID
	Method("get", func() {
		Description("Get organization by ID")

		Payload(func() {
			Attribute("id", String, "Organization ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Required("id")
		})

		Result(OrganizationResult)

		Error("not_found")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			GET("/{id}")
			types.CommonHeaders()
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("not_found", CodeNotFound)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
		})
	})

	// List organizations with pagination
	Method("list", func() {
		Description("List organizations with pagination and filtering")

		Payload(func() {
			Extend(types.Pagination)
			Attribute("name_filter", String, "Filter by organization name", func() {
				Example("acme")
			})
			Attribute("industry_filter", String, "Filter by industry", func() {
				Example("Manufacturing")
			})
			Attribute("status_filter", String, "Filter by status", func() {
				Enum("active", "inactive", "suspended")
				Example("active")
			})
		})

		Result(types.PaginatedResponse(OrganizationResult))

		Error("bad_request")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			GET("/")
			types.CommonHeaders()
			Param("page")
			Param("page_size")
			Param("sort_by")
			Param("sort_order")
			Param("name_filter")
			Param("industry_filter")
			Param("status_filter")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
		})
	})

	// Update organization
	Method("update", func() {
		Description("Update an existing organization")

		Payload(func() {
			Attribute("id", String, "Organization ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Attribute("name", String, "Organization name", func() {
				MinLength(1)
				MaxLength(100)
				Example("Acme Corporation Updated")
			})
			Attribute("description", String, "Organization description", func() {
				MaxLength(500)
				Example("Updated description")
			})
			Attribute("tax_id", String, "Tax identification number", func() {
				Example("12-3456789")
			})
			Attribute("industry", String, "Industry sector", func() {
				Example("Manufacturing")
			})
			Attribute("website", String, "Organization website", func() {
				Format(FormatURI)
				Example("https://acme.com")
			})
			Attribute("status", String, "Organization status", func() {
				Enum("active", "inactive", "suspended")
				Example("active")
			})
			Attribute("contact", types.ContactInfo, "Primary contact information")
			Attribute("address", OrganizationAddress, "Organization address")
			Required("id")
		})

		Result(OrganizationResult)

		Error("bad_request")
		Error("not_found")
		Error("conflict")
		Error("unauthorized")
		Error("forbidden")
		Error("unprocessable_entity")

		HTTP(func() {
			PUT("/{id}")
			types.CommonHeaders()
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("conflict", StatusConflict)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("unprocessable_entity", StatusUnprocessableEntity)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("not_found", CodeNotFound)
			Response("conflict", CodeAlreadyExists)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("unprocessable_entity", CodeInvalidArgument)
		})
	})

	// Delete organization
	Method("delete", func() {
		Description("Delete an organization (soft delete)")

		Payload(func() {
			Attribute("id", String, "Organization ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Required("id")
		})

		Result(Empty)

		Error("not_found")
		Error("unauthorized")
		Error("forbidden")
		Error("conflict") // If organization has dependencies

		HTTP(func() {
			DELETE("/{id}")
			types.CommonHeaders()
			Response(StatusNoContent)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("conflict", StatusConflict)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("not_found", CodeNotFound)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("conflict", CodeFailedPrecondition)
		})
	})

	// Health check endpoint
	Method("health", func() {
		Description("Health check for organization service")

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
				Example("1.2.3")
			})
			Required("status", "timestamp", "version")
		})

		HTTP(func() {
			GET("/health")
			Response(StatusOK)
		})

		GRPC(func() {
			Response(CodeOK)
		})

		// No authentication required for health checks
		NoSecurity()
	})
})

// OrganizationResult describes the organization response
var OrganizationResult = ResultType("application/vnd.organization", func() {
	Description("Organization information")
	Attributes(func() {
		Attribute("id", String, "Unique organization identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
		})
		Attribute("name", String, "Organization name", func() {
			MinLength(1)
			MaxLength(100)
			Example("Acme Corporation")
		})
		Attribute("description", String, "Organization description", func() {
			MaxLength(500)
			Example("Leading provider of roadrunner traps and anvils")
		})
		Attribute("tax_id", String, "Tax identification number", func() {
			Example("12-3456789")
		})
		Attribute("industry", String, "Industry sector", func() {
			Example("Manufacturing")
		})
		Attribute("website", String, "Organization website", func() {
			Format(FormatURI)
			Example("https://acme.com")
		})
		Attribute("status", String, "Organization status", func() {
			Enum("active", "inactive", "suspended")
			Example("active")
		})
		Attribute("tenant_id", String, "Tenant identifier", func() {
			Format(FormatUUID)
			Example("123e4567-e89b-12d3-a456-426614174000")
		})
		Attribute("contact", types.ContactInfo, "Primary contact information")
		Attribute("address", OrganizationAddress, "Organization address")
		types.AuditFields()
	})
	Required("id", "name", "status", "tenant_id", "created_at", "updated_at")

	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("description")
		Attribute("tax_id")
		Attribute("industry")
		Attribute("website")
		Attribute("status")
		Attribute("tenant_id")
		Attribute("contact")
		Attribute("address")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})

	View("minimal", func() {
		Attribute("id")
		Attribute("name")
		Attribute("status")
		Attribute("tenant_id")
	})
})

// OrganizationAddress describes organization address
var OrganizationAddress = Type("OrganizationAddress", func() {
	Description("Organization address information")
	Attribute("street", String, "Street address", func() {
		MaxLength(200)
		Example("123 Main Street")
	})
	Attribute("city", String, "City", func() {
		MaxLength(100)
		Example("New York")
	})
	Attribute("state", String, "State/Province", func() {
		MaxLength(100)
		Example("NY")
	})
	Attribute("postal_code", String, "Postal/ZIP code", func() {
		MaxLength(20)
		Example("10001")
	})
	Attribute("country", String, "Country", func() {
		MaxLength(100)
		Example("United States")
	})
	Attribute("latitude", Float64, "Latitude coordinate", func() {
		Example(40.7128)
	})
	Attribute("longitude", Float64, "Longitude coordinate", func() {
		Example(-74.0060)
	})
})
