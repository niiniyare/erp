package user

import (
	"github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// UserService defines the user management service
var _ = Service("user", func() {
	Description("User management service with RBAC and ABAC capabilities")

	HTTP(func() {
		Path("/api/v1/users")
	})

	// Create user endpoint
	Method("create", func() {
		Description("Create a new user")

		Payload(CreateUserPayload)
		Result(UserResult)

		Error("bad_request")
		Error("conflict") // For email conflicts
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

	// Get user by ID
	Method("get", func() {
		Description("Get user by ID")

		Payload(func() {
			Attribute("id", String, "User ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Required("id")
		})

		Result(UserResult)

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

	// List users with pagination
	Method("list", func() {
		Description("List users with pagination and filtering")

		Payload(func() {
			Extend(types.Pagination)
			Attribute("email_filter", String, "Filter by email", func() {
				Example("john@example.com")
			})
			Attribute("status_filter", String, "Filter by status", func() {
				Enum("ACTIVE", "INACTIVE", "LOCKED", "SUSPENDED", "PENDING_VERIFICATION", "EXPIRED")
				Example("ACTIVE")
			})
			Attribute("user_type_filter", String, "Filter by user type", func() {
				Enum("INTERNAL", "CUSTOMER", "VENDOR", "PARTNER", "API", "SERVICE", "ADMIN")
				Example("INTERNAL")
			})
		})

		Result(func() {
			Attribute("data", ArrayOf(UserResult), "The data items")
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
			Param("email_filter")
			Param("status_filter")
			Param("user_type_filter")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// Update user
	Method("update", func() {
		Description("Update an existing user")

		Payload(UpdateUserPayload)
		Result(UserResult)

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

	// Deactivate user
	Method("deactivate", func() {
		Description("Deactivate a user account")

		Payload(func() {
			Attribute("id", String, "User ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Required("id")
		})

		Result(Empty)

		Error("not_found")
		Error("unauthorized")
		Error("forbidden")
		Error("conflict") // If user cannot be deactivated

		HTTP(func() {
			PATCH("/{id}/deactivate")
			Response(StatusNoContent)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("conflict", StatusConflict)
		})
	})

	// Get user permissions
	Method("permissions", func() {
		Description("Get user permissions and roles")

		Payload(func() {
			Attribute("id", String, "User ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Required("id")
		})

		Result(UserPermissionsResult)

		Error("not_found")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			GET("/{id}/permissions")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// Assign role to user
	Method("assign_role", func() {
		Description("Assign a role to a user")

		Payload(AssignRolePayload)
		Result(Empty)

		Error("bad_request")
		Error("not_found")
		Error("unauthorized")
		Error("forbidden")
		Error("conflict")

		HTTP(func() {
			POST("/{user_id}/roles")
			Response(StatusNoContent)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("conflict", StatusConflict)
		})
	})

	// Remove role from user
	Method("remove_role", func() {
		Description("Remove a role from a user")

		Payload(func() {
			Attribute("user_id", String, "User ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Attribute("role_id", String, "Role ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440001")
			})
			Required("user_id", "role_id")
		})

		Result(Empty)

		Error("bad_request")
		Error("not_found")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			DELETE("/{user_id}/roles/{role_id}")
			Response(StatusNoContent)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})
})
