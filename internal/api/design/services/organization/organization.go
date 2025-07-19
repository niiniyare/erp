package organization

import (
	"github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// OrganizationService defines the organization management service
var _ = Service("organization", func() {
	Description("Organization and entity management service")

	HTTP(func() {
		Path("/api/v1/organizations")
	})

	// Create organization endpoint
	Method("create", func() {
		Description("Create a new organization")

		Payload(CreateOrganizationPayload)
		Result(OrganizationResult)

		Error("bad_request")
		Error("conflict") // For name conflicts
		Error("unauthorized")
		Error("forbidden")
		Error("unprocessable_entity")

		HTTP(func() {
			POST("/")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("conflict", StatusConflict)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("unprocessable_entity", StatusUnprocessableEntity)
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
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
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
			Attribute("type_filter", String, "Filter by organization type", func() {
				Enum("CORPORATION", "LLC", "PARTNERSHIP", "SOLE_PROPRIETORSHIP", "NON_PROFIT", "GOVERNMENT")
				Example("CORPORATION")
			})
			Attribute("status_filter", String, "Filter by status", func() {
				Enum("ACTIVE", "INACTIVE", "SUSPENDED", "DISSOLVED")
				Example("ACTIVE")
			})
		})

		Result(func() {
			Attribute("data", ArrayOf(OrganizationResult), "The data items")
			Attribute("pagination", types.PaginationMeta, "Pagination metadata")
			Required("data", "pagination")
		})

		Error("bad_request")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			GET("/")
			Param("page")
			Param("page_size")
			Param("sort_by")
			Param("sort_order")
			Param("name_filter")
			Param("type_filter")
			Param("status_filter")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// Update organization
	Method("update", func() {
		Description("Update an existing organization")

		Payload(UpdateOrganizationPayload)
		Result(OrganizationResult)

		Error("bad_request")
		Error("not_found")
		Error("conflict")
		Error("unauthorized")
		Error("forbidden")
		Error("unprocessable_entity")

		HTTP(func() {
			PUT("/{id}")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("conflict", StatusConflict)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("unprocessable_entity", StatusUnprocessableEntity)
		})
	})

	// Get organization hierarchy
	Method("hierarchy", func() {
		Description("Get organization hierarchy (parent/child relationships)")

		Payload(func() {
			Attribute("id", String, "Organization ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Attribute("depth", UInt, "Hierarchy depth to retrieve", func() {
				Default(3)
				Maximum(10)
				Example(3)
			})
			Required("id")
		})

		Result(HierarchyResult)

		Error("not_found")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			GET("/{id}/hierarchy")
			Param("depth")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// Archive organization
	Method("archive", func() {
		Description("Archive an organization (soft delete)")

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
			PATCH("/{id}/archive")
			Response(StatusNoContent)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("conflict", StatusConflict)
		})
	})
})
