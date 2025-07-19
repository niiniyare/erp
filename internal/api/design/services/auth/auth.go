package auth

import (
	. "goa.design/goa/v3/dsl"
)

// AuthService defines the authentication service
var _ = Service("auth", func() {
	Description("Authentication and authorization service")

	HTTP(func() {
		Path("/api/v1/auth")
	})

	// Login endpoint
	Method("login", func() {
		Description("Authenticate user and return JWT token")

		Payload(func() {
			Attribute("email", String, "User email", func() {
				Format(FormatEmail)
				Example("user@example.com")
			})
			Attribute("password", String, "User password", func() {
				MinLength(8)
				Example("password123")
			})
			Attribute("tenant_id", String, "Tenant ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Required("email", "password")
		})

		Result(AuthResult)

		Error("bad_request")
		Error("unauthorized")
		Error("not_found")
		Error("internal_error")

		HTTP(func() {
			POST("/login")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("not_found", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
		})

		// No security required for login
		NoSecurity()
	})

	// Refresh token endpoint
	Method("refresh", func() {
		Description("Refresh JWT token")

		Payload(func() {
			Attribute("refresh_token", String, "Refresh token", func() {
				Example("refresh_token_here")
			})
			Required("refresh_token")
		})

		Result(AuthResult)

		Error("bad_request")
		Error("unauthorized")
		Error("internal_error")

		HTTP(func() {
			POST("/refresh")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("internal_error", StatusInternalServerError)
		})

		NoSecurity()
	})

	// Logout endpoint
	Method("logout", func() {
		Description("Logout user and invalidate token")

		Payload(func() {
			Token("token", String, "JWT token")
		})

		Result(Empty)

		Error("unauthorized")
		Error("internal_error")

		Security("jwt")

		HTTP(func() {
			POST("/logout")
			Response(StatusNoContent)
			Response("unauthorized", StatusUnauthorized)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// Validate token endpoint
	Method("validate", func() {
		Description("Validate JWT token")

		Payload(func() {
			Token("token", String, "JWT token")
		})

		Result(TokenValidationResult)

		Error("unauthorized")
		Error("internal_error")

		Security("jwt")

		HTTP(func() {
			GET("/validate")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("internal_error", StatusInternalServerError)
		})
	})
})
