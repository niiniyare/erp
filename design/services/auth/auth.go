package auth

import (
	. "goa.design/goa/v3/dsl"
	"design/types"
)

// Service describes the authentication service
var _ = Service("auth", func() {
	Description("Authentication and authorization service")
	
	// Apply global middleware
	HTTP(func() {
		Path("/api/v1/auth")
	})
	
	// Login endpoint
	Method("login", func() {
		Description("Authenticate user and return JWT tokens")
		
		Payload(types.LoginRequest)
		Result(types.LoginResponse)
		
		Error("bad_request")
		Error("unauthorized")
		Error("forbidden")
		Error("not_found")
		
		HTTP(func() {
			POST("/login")
			types.CommonHeaders()
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
		})
		
		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("not_found", CodeNotFound)
		})
		
		// No authentication required for login
		NoSecurity()
	})
	
	// Refresh token endpoint
	Method("refresh", func() {
		Description("Refresh JWT access token using refresh token")
		
		Payload(func() {
			Attribute("refresh_token", String, "Refresh token", func() {
				Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
			})
			Required("refresh_token")
		})
		
		Result(func() {
			Attribute("access_token", String, "New JWT access token", func() {
				Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
			})
			Attribute("expires_in", UInt, "Token expiration time in seconds", func() {
				Example(3600)
			})
			Attribute("token_type", String, "Token type", func() {
				Example("Bearer")
			})
			Required("access_token", "expires_in", "token_type")
		})
		
		Error("bad_request")
		Error("unauthorized")
		
		HTTP(func() {
			POST("/refresh")
			types.CommonHeaders()
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
		})
		
		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
		})
		
		// No authentication required for refresh
		NoSecurity()
	})
	
	// Logout endpoint
	Method("logout", func() {
		Description("Logout user and invalidate tokens")
		
		Payload(func() {
			Attribute("refresh_token", String, "Refresh token to invalidate", func() {
				Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
			})
			Required("refresh_token")
		})
		
		Result(func() {
			Attribute("success", Boolean, "Whether logout was successful")
			Attribute("message", String, "Success message")
			Required("success", "message")
		})
		
		Error("bad_request")
		Error("unauthorized")
		
		HTTP(func() {
			POST("/logout")
			types.CommonHeaders()
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
		})
		
		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
		})
		
		// JWT required for logout
		Security("jwt")
	})
	
	// Verify token endpoint
	Method("verify", func() {
		Description("Verify JWT token and return user information")
		
		Payload(func() {
			Attribute("token", String, "JWT token to verify", func() {
				Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
			})
			Required("token")
		})
		
		Result(func() {
			Attribute("valid", Boolean, "Whether token is valid")
			Attribute("user", types.UserResult, "User information", func() {
				View("minimal")
			})
			Attribute("permissions", ArrayOf(types.PermissionInfo), "User permissions")
			Attribute("expires_at", String, "Token expiration timestamp", func() {
				Format(FormatDateTime)
				Example("2023-12-07T10:30:00Z")
			})
			Required("valid")
		})
		
		Error("bad_request")
		Error("unauthorized")
		
		HTTP(func() {
			POST("/verify")
			types.CommonHeaders()
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
		})
		
		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
		})
		
		// No authentication required for verification
		NoSecurity()
	})
	
	// Change password endpoint
	Method("change_password", func() {
		Description("Change user password")
		
		Payload(func() {
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
			Required("current_password", "new_password")
		})
		
		Result(func() {
			Attribute("success", Boolean, "Whether password change was successful")
			Attribute("message", String, "Success message")
			Required("success", "message")
		})
		
		Error("bad_request")
		Error("unauthorized")
		Error("forbidden")
		Error("unprocessable_entity")
		
		HTTP(func() {
			POST("/change-password")
			types.CommonHeaders()
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("unprocessable_entity", StatusUnprocessableEntity)
		})
		
		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("unprocessable_entity", CodeInvalidArgument)
		})
		
		// JWT required for password change
		Security("jwt")
	})
	
	// Forgot password endpoint
	Method("forgot_password", func() {
		Description("Initiate password reset process")
		
		Payload(func() {
			Attribute("email", String, "User email address", func() {
				Format(FormatEmail)
				Example("john.doe@acme.com")
			})
			Attribute("tenant_id", String, "Tenant identifier", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Required("email")
		})
		
		Result(func() {
			Attribute("success", Boolean, "Whether reset email was sent")
			Attribute("message", String, "Success message")
			Required("success", "message")
		})
		
		Error("bad_request")
		Error("not_found")
		
		HTTP(func() {
			POST("/forgot-password")
			types.CommonHeaders()
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
		})
		
		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("not_found", CodeNotFound)
		})
		
		// No authentication required for forgot password
		NoSecurity()
	})
	
	// Reset password endpoint
	Method("reset_password", func() {
		Description("Reset password using reset token")
		
		Payload(func() {
			Attribute("token", String, "Password reset token", func() {
				Example("reset_token_abc123")
			})
			Attribute("new_password", String, "New password", func() {
				MinLength(8)
				MaxLength(100)
				Example("NewPassword123!")
			})
			Required("token", "new_password")
		})
		
		Result(func() {
			Attribute("success", Boolean, "Whether password reset was successful")
			Attribute("message", String, "Success message")
			Required("success", "message")
		})
		
		Error("bad_request")
		Error("unauthorized")
		Error("not_found")
		Error("unprocessable_entity")
		
		HTTP(func() {
			POST("/reset-password")
			types.CommonHeaders()
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("not_found", StatusNotFound)
			Response("unprocessable_entity", StatusUnprocessableEntity)
		})
		
		GRPC(func() {
			Response(CodeOK)
			Response("bad_request", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("not_found", CodeNotFound)
			Response("unprocessable_entity", CodeInvalidArgument)
		})
		
		// No authentication required for reset password
		NoSecurity()
	})
	
	// Health check endpoint
	Method("health", func() {
		Description("Health check for auth service")
		
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