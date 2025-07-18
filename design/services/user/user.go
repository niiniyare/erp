package user

import (
	. "goa.design/goa/v3/dsl"
	"design/types"
)

// Service describes the user management service
var _ = Service("user", func() {
	Description("User management service with RBAC and ABAC capabilities")
	
	// Apply global middleware
	HTTP(func() {
		Path("/api/v1/users")
	})
	
	// Security requirements
	Security("jwt", func() {
		Scope("api:read", "api:write")
	})
	
	// Create user endpoint
	Method("create", func() {
		Description("Create a new user")
		
		Payload(types.CreateUserPayload)
		Result(types.UserResult)
		
		Error("bad_request")
		Error("conflict") // For username/email conflicts
		Error("unauthorized")
		Error("unprocessable_entity")
		
		HTTP(func() {
			POST("/")
			types.CommonHeaders()
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("conflict", StatusConflict)
			Response("unauthorized", StatusUnauthorized)
			Response("unprocessable_entity", StatusUnprocessableEntity)
		})
		
		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("conflict", CodeAlreadyExists)
			Response("unauthorized", CodeUnauthenticated)
			Response("unprocessable_entity", CodeInvalidArgument)
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
		
		Result(types.UserResult)
		
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
	
	// List users with pagination
	Method("list", func() {
		Description("List users with pagination and filtering")
		
		Payload(func() {
			Extend(types.Pagination)
			Attribute("username_filter", String, "Filter by username", func() {
				Example("john")
			})
			Attribute("email_filter", String, "Filter by email", func() {
				Example("john@acme.com")
			})
			Attribute("status_filter", String, "Filter by status", func() {
				Enum("ACTIVE", "INACTIVE", "LOCKED", "SUSPENDED")
				Example("ACTIVE")
			})
			Attribute("user_type_filter", String, "Filter by user type", func() {
				Enum("INTERNAL", "CUSTOMER", "VENDOR", "PARTNER")
				Example("INTERNAL")
			})
		})
		
		Result(types.PaginatedResponse(types.UserResult))
		
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
			Param("username_filter")
			Param("email_filter")
			Param("status_filter")
			Param("user_type_filter")
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
	
	// Update user
	Method("update", func() {
		Description("Update an existing user")
		
		Payload(types.UpdateUserPayload)
		Result(types.UserResult)
		
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
	
	// Delete user
	Method("delete", func() {
		Description("Delete a user (soft delete)")
		
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
		Error("conflict") // If user has dependencies
		
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
	
	// Change password
	Method("change_password", func() {
		Description("Change user password")
		
		Payload(func() {
			Attribute("id", String, "User ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Attribute("current_password", String, "Current password", func() {
				MinLength(8)
				MaxLength(100)
				Example("OldPassword123!")
			})
			Attribute("new_password", String, "New password", func() {
				MinLength(8)
				MaxLength(100)
				Example("NewPassword123!")
			})
			Required("id", "current_password", "new_password")
		})
		
		Result(func() {
			Attribute("success", Boolean, "Whether password change was successful")
			Attribute("message", String, "Success message")
			Required("success", "message")
		})
		
		Error("bad_request")
		Error("unauthorized")
		Error("forbidden")
		Error("not_found")
		Error("unprocessable_entity")
		
		HTTP(func() {
			POST("/{id}/change-password")
			types.CommonHeaders()
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
			Response("unprocessable_entity", StatusUnprocessableEntity)
		})
		
		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
			Response("unprocessable_entity", CodeInvalidArgument)
		})
	})
	
	// Get user permissions
	Method("get_permissions", func() {
		Description("Get user permissions")
		
		Payload(func() {
			Attribute("id", String, "User ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Required("id")
		})
		
		Result(func() {
			Attribute("user_id", String, "User ID")
			Attribute("permissions", ArrayOf(types.PermissionInfo), "User permissions")
			Required("user_id", "permissions")
		})
		
		Error("not_found")
		Error("unauthorized")
		Error("forbidden")
		
		HTTP(func() {
			GET("/{id}/permissions")
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
	
	// Health check endpoint
	Method("health", func() {
		Description("Health check for user service")
		
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