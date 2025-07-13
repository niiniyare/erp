package design

import . "goa.design/goa/v3/dsl"

var _ = Service("user", func() {
	Description("User management service")
	
	HTTP(func() {
		Path("/users")
	})

	Method("create_user", func() {
		Description("Create a new user")
		
		Payload(func() {
			Attribute("entity_id", String, "Entity ID")
			Attribute("first_name", String, "First name")
			Attribute("last_name", String, "Last name")
			Attribute("middle_name", String, "Middle name")
			Attribute("email", String, "Email address")
			Attribute("phone", String, "Phone number")
			Attribute("username", String, "Username")
			Attribute("password", String, "Password")
			Attribute("user_type", String, "User type (ADMIN, INTERNAL, CUSTOMER, VENDOR)")
			Attribute("role_names", ArrayOf(String), "Role names to assign")
			Attribute("security_level", Int32, "Security level (1-10)")
			Attribute("department", String, "Department")
			Attribute("position", String, "Position title")
			Attribute("is_active", Boolean, "Is user active")
			
			Required("entity_id", "first_name", "last_name", "email", "username", "password", "user_type")
		})
		
		Result(func() {
			Attribute("user_id", String, "User ID")
			Attribute("person_id", String, "Person ID")
			Attribute("employee_id", String, "Employee ID")
			Attribute("username", String, "Username")
			Attribute("email", String, "Email")
			Attribute("role_ids", ArrayOf(String), "Assigned role IDs")
			Attribute("created_at", String, "Creation timestamp")
		})
		
		Error("unauthorized", String, "Unauthorized")
		Error("bad_request", String, "Bad request")
		Error("internal_error", String, "Internal server error")
		
		HTTP(func() {
			POST("/")
			Response(StatusCreated)
			Response("unauthorized", StatusUnauthorized)
			Response("bad_request", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
		
		GRPC(func() {})
	})

	Method("get_user", func() {
		Description("Get user by ID")
		
		Payload(func() {
			Attribute("user_id", String, "User ID")
			Required("user_id")
		})
		
		Result(func() {
			Attribute("id", String, "User ID")
			Attribute("entity_id", String, "Entity ID")
			Attribute("person_id", String, "Person ID")
			Attribute("employee_id", String, "Employee ID")
			Attribute("username", String, "Username")
			Attribute("email", String, "Email")
			Attribute("user_type", String, "User type")
			Attribute("account_status", String, "Account status")
			Attribute("is_active", Boolean, "Is active")
			Attribute("last_login_at", String, "Last login timestamp")
			Attribute("mfa_enabled", Boolean, "MFA enabled")
			Attribute("created_at", String, "Creation timestamp")
			Attribute("updated_at", String, "Update timestamp")
		})
		
		Error("unauthorized", String, "Unauthorized")
		Error("not_found", String, "User not found")
		Error("internal_error", String, "Internal server error")
		
		HTTP(func() {
			GET("/{user_id}")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("not_found", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
		})
		
		GRPC(func() {})
	})

	Method("list_users", func() {
		Description("List users with pagination")
		
		Payload(func() {
			Attribute("user_type", String, "Filter by user type")
			Attribute("account_status", String, "Filter by account status")
			Attribute("limit", Int32, "Page size", func() {
				Default(20)
				Minimum(1)
				Maximum(100)
			})
			Attribute("offset", Int32, "Page offset", func() {
				Default(0)
				Minimum(0)
			})
		})
		
		Result(func() {
			Attribute("users", ArrayOf(UserResult), "List of users")
			Attribute("total", Int32, "Total count")
			Attribute("limit", Int32, "Page size")
			Attribute("offset", Int32, "Page offset")
		})
		
		Error("unauthorized", String, "Unauthorized")
		Error("bad_request", String, "Bad request")
		Error("internal_error", String, "Internal server error")
		
		HTTP(func() {
			GET("/")
			Param("user_type")
			Param("account_status")
			Param("limit")
			Param("offset")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("bad_request", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
		
		GRPC(func() {})
	})

	Method("update_user", func() {
		Description("Update user information")
		
		Payload(func() {
			Attribute("user_id", String, "User ID")
			Attribute("username", String, "Username")
			Attribute("email", String, "Email")
			Attribute("user_type", String, "User type")
			Attribute("account_status", String, "Account status")
			Attribute("mfa_enabled", Boolean, "MFA enabled")
			Attribute("session_timeout_minutes", Int32, "Session timeout in minutes")
			
			Required("user_id")
		})
		
		Result(UserResult)
		
		Error("unauthorized", String, "Unauthorized")
		Error("not_found", String, "User not found")
		Error("bad_request", String, "Bad request")
		Error("internal_error", String, "Internal server error")
		
		HTTP(func() {
			PUT("/{user_id}")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("not_found", StatusNotFound)
			Response("bad_request", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
		
		GRPC(func() {})
	})

	Method("update_user_password", func() {
		Description("Update user password")
		
		Payload(func() {
			Attribute("user_id", String, "User ID")
			Attribute("current_password", String, "Current password")
			Attribute("new_password", String, "New password")
			
			Required("user_id", "current_password", "new_password")
		})
		
		Result(func() {
			Attribute("success", Boolean, "Password updated successfully")
			Attribute("message", String, "Success message")
		})
		
		Error("unauthorized", String, "Unauthorized")
		Error("not_found", String, "User not found")
		Error("bad_request", String, "Bad request")
		Error("internal_error", String, "Internal server error")
		
		HTTP(func() {
			PUT("/{user_id}/password")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("not_found", StatusNotFound)
			Response("bad_request", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
		
		GRPC(func() {})
	})

	Method("delete_user", func() {
		Description("Soft delete user")
		
		Payload(func() {
			Attribute("user_id", String, "User ID")
			Required("user_id")
		})
		
		Result(func() {
			Attribute("success", Boolean, "User deleted successfully")
			Attribute("message", String, "Success message")
		})
		
		Error("unauthorized", String, "Unauthorized")
		Error("not_found", String, "User not found")
		Error("internal_error", String, "Internal server error")
		
		HTTP(func() {
			DELETE("/{user_id}")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("not_found", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
		})
		
		GRPC(func() {})
	})

	Method("authenticate_user", func() {
		Description("Authenticate user with username/email and password")
		
		Payload(func() {
			Attribute("identifier", String, "Username or email")
			Attribute("password", String, "Password")
			
			Required("identifier", "password")
		})
		
		Result(func() {
			Attribute("user", UserResult, "User information")
			Attribute("token", String, "Authentication token")
			Attribute("expires_at", String, "Token expiration")
		})
		
		Error("unauthorized", String, "Invalid credentials")
		Error("account_locked", String, "Account is locked")
		Error("bad_request", String, "Bad request")
		Error("internal_error", String, "Internal server error")
		
		HTTP(func() {
			POST("/auth")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("account_locked", StatusLocked)
			Response("bad_request", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
		
		GRPC(func() {})
	})

	Method("get_user_roles", func() {
		Description("Get user roles")
		
		Payload(func() {
			Attribute("user_id", String, "User ID")
			Required("user_id")
		})
		
		Result(func() {
			Attribute("roles", ArrayOf(UserRoleResult), "User roles")
		})
		
		Error("unauthorized", String, "Unauthorized")
		Error("not_found", String, "User not found")
		Error("internal_error", String, "Internal server error")
		
		HTTP(func() {
			GET("/{user_id}/roles")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("not_found", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
		})
		
		GRPC(func() {})
	})

	Method("get_user_permissions", func() {
		Description("Get user effective permissions")
		
		Payload(func() {
			Attribute("user_id", String, "User ID")
			Required("user_id")
		})
		
		Result(func() {
			Attribute("permissions", ArrayOf(UserPermissionResult), "User permissions")
		})
		
		Error("unauthorized", String, "Unauthorized")
		Error("not_found", String, "User not found")
		Error("internal_error", String, "Internal server error")
		
		HTTP(func() {
			GET("/{user_id}/permissions")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("not_found", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
		})
		
		GRPC(func() {})
	})

	Method("search_users", func() {
		Description("Search users by various criteria")
		
		Payload(func() {
			Attribute("query", String, "Search query")
			Attribute("limit", Int32, "Page size", func() {
				Default(20)
				Minimum(1)
				Maximum(100)
			})
			Attribute("offset", Int32, "Page offset", func() {
				Default(0)
				Minimum(0)
			})
			
			Required("query")
		})
		
		Result(func() {
			Attribute("users", ArrayOf(UserSearchResult), "Search results")
			Attribute("total", Int32, "Total count")
			Attribute("limit", Int32, "Page size")
			Attribute("offset", Int32, "Page offset")
		})
		
		Error("unauthorized", String, "Unauthorized")
		Error("bad_request", String, "Bad request")
		Error("internal_error", String, "Internal server error")
		
		HTTP(func() {
			GET("/search")
			Param("query")
			Param("limit")
			Param("offset")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("bad_request", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
		
		GRPC(func() {})
	})
})

// UserResult represents a user in API responses
var UserResult = Type("UserResult", func() {
	Attribute("id", String, "User ID")
	Attribute("entity_id", String, "Entity ID")
	Attribute("person_id", String, "Person ID")
	Attribute("employee_id", String, "Employee ID")
	Attribute("username", String, "Username")
	Attribute("email", String, "Email")
	Attribute("user_type", String, "User type")
	Attribute("account_status", String, "Account status")
	Attribute("is_active", Boolean, "Is active")
	Attribute("last_login_at", String, "Last login timestamp")
	Attribute("mfa_enabled", Boolean, "MFA enabled")
	Attribute("created_at", String, "Creation timestamp")
	Attribute("updated_at", String, "Update timestamp")
})

// UserRoleResult represents a user role in API responses
var UserRoleResult = Type("UserRoleResult", func() {
	Attribute("id", String, "Role ID")
	Attribute("name", String, "Role name")
	Attribute("display_name", String, "Role display name")
	Attribute("description", String, "Role description")
	Attribute("role_type", String, "Role type")
	Attribute("assignment_type", String, "Assignment type")
	Attribute("assigned_at", String, "Assignment timestamp")
	Attribute("expires_at", String, "Expiration timestamp")
})

// UserPermissionResult represents a user permission in API responses
var UserPermissionResult = Type("UserPermissionResult", func() {
	Attribute("id", String, "Permission ID")
	Attribute("name", String, "Permission name")
	Attribute("display_name", String, "Permission display name")
	Attribute("resource_name", String, "Resource name")
	Attribute("action_name", String, "Action name")
	Attribute("permission_source", String, "Permission source (ROLE, DIRECT)")
	Attribute("source_role", String, "Source role name")
	Attribute("effect", String, "Effect (ALLOW, DENY)")
})

// UserSearchResult represents a user search result
var UserSearchResult = Type("UserSearchResult", func() {
	Attribute("id", String, "User ID")
	Attribute("username", String, "Username")
	Attribute("email", String, "Email")
	Attribute("full_name", String, "Full name")
	Attribute("user_type", String, "User type")
	Attribute("account_status", String, "Account status")
	Attribute("is_active", Boolean, "Is active")
	Attribute("employee_number", String, "Employee number")
	Attribute("position_title", String, "Position title")
	Attribute("roles", String, "Roles (comma-separated)")
})