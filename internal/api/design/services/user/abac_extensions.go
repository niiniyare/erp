package user

import (
	. "goa.design/goa/v3/dsl"
)

// ABAC extensions for the user service
var _ = Service("user", func() {
	// ... existing user service methods ...

	// User Attributes for ABAC
	Method("get_attributes", func() {
		Description("Get user attributes for ABAC evaluation")

		Payload(func() {
			Attribute("id", String, "User ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Attribute("include_metadata", Boolean, "Include attribute metadata", func() {
				Default(false)
			})
			Attribute("include_derived", Boolean, "Include computed attributes", func() {
				Default(false)
			})
			Attribute("fresh_only", Boolean, "Only return non-expired attributes", func() {
				Default(false)
			})
			Attribute("attribute_filter", String, "Comma-separated attribute names", func() {
				Example("user.security_level,user.department")
			})
			Required("id")
		})

		Result(UserAttributesResult)

		Error("not_found")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			GET("/{id}/attributes")
			Param("include_metadata")
			Param("include_derived")
			Param("fresh_only")
			Param("attribute_filter")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("set_attributes", func() {
		Description("Set/update user attributes for ABAC")

		Payload(SetUserAttributesPayload)
		Result(SetUserAttributesResult)

		Error("bad_request")
		Error("not_found")
		Error("unauthorized")
		Error("forbidden")
		Error("unprocessable_entity")

		HTTP(func() {
			PUT("/{id}/attributes")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("unprocessable_entity", StatusUnprocessableEntity)
		})
	})

	Method("bulk_update_attributes", func() {
		Description("Bulk update user attributes for ABAC")

		Payload(BulkUpdateUserAttributesPayload)
		Result(BulkUpdateUserAttributesResult)

		Error("bad_request")
		Error("unauthorized")
		Error("forbidden")
		Error("unprocessable_entity")

		HTTP(func() {
			POST("/bulk-attributes")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("unprocessable_entity", StatusUnprocessableEntity)
		})
	})

	// ABAC Permission Checking
	Method("check_permission", func() {
		Description("Check if user has permission using ABAC context")

		Payload(CheckPermissionPayload)
		Result(PermissionCheckResult)

		Error("bad_request")
		Error("not_found")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			POST("/{user_id}/check-permission")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("authorize_action", func() {
		Description("Authorize user action with ABAC context")

		Payload(AuthorizeActionPayload)
		Result(AuthorizationResult)

		Error("bad_request")
		Error("not_found")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			POST("/{user_id}/authorize")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// Session Context for ABAC
	Method("get_session_attributes", func() {
		Description("Get user session attributes for ABAC")

		Payload(func() {
			Attribute("user_id", String, "User ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Attribute("session_id", String, "Session ID", func() {
				Example("session_abc123")
			})
			Attribute("include_analytics", Boolean, "Include session analytics", func() {
				Default(false)
			})
			Attribute("include_risk_assessment", Boolean, "Include risk assessment", func() {
				Default(false)
			})
			Required("user_id", "session_id")
		})

		Result(SessionAttributesResult)

		Error("not_found")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			GET("/{user_id}/sessions/{session_id}/attributes")
			Param("include_analytics")
			Param("include_risk_assessment")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("set_session_context", func() {
		Description("Set session context for user")

		Payload(SetSessionContextPayload)
		Result(SetSessionContextResult)

		Error("bad_request")
		Error("not_found")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			PUT("/{user_id}/sessions/{session_id}/context")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// User Context Analytics
	Method("get_user_context", func() {
		Description("Get comprehensive user context for ABAC")

		Payload(func() {
			Attribute("id", String, "User ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Attribute("include_derived", Boolean, "Include derived attributes", func() {
				Default(true)
			})
			Attribute("include_session", Boolean, "Include current session context", func() {
				Default(true)
			})
			Attribute("include_access_patterns", Boolean, "Include access patterns", func() {
				Default(false)
			})
			Required("id")
		})

		Result(UserContextResult)

		Error("not_found")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			GET("/{id}/context")
			Param("include_derived")
			Param("include_session")
			Param("include_access_patterns")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// Attribute Health and Management
	Method("validate_attributes", func() {
		Description("Validate user attributes for ABAC compliance")

		Payload(func() {
			Attribute("id", String, "User ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Attribute("attributes", MapOf(String, Any), "Attributes to validate")
			Attribute("validation_rules", ArrayOf(String), "Specific validation rules", func() {
				Example([]string{"security_clearance", "department_consistency", "expiration_check"})
			})
			Required("id")
		})

		Result(AttributeValidationResult)

		Error("bad_request")
		Error("not_found")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			POST("/{id}/validate-attributes")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("refresh_attributes", func() {
		Description("Refresh user attributes from authoritative sources")

		Payload(func() {
			Attribute("id", String, "User ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Attribute("sources", ArrayOf(String), "Specific sources to refresh", func() {
				Example([]string{"hr_system", "security_system", "ldap"})
			})
			Attribute("force_refresh", Boolean, "Force refresh even if not stale", func() {
				Default(false)
			})
			Required("id")
		})

		Result(RefreshAttributesResult)

		Error("bad_request")
		Error("not_found")
		Error("unauthorized")
		Error("forbidden")

		HTTP(func() {
			POST("/{id}/refresh-attributes")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})
})
