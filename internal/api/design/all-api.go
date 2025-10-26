package design

// import (
// 	. "goa.design/goa/v3/dsl"
// )
//
// var _ = API("erp", func() {
// 	Title("Enterprise Resource Planning Multi-Tenant API")
// 	Description("Comprehensive ERP system with RBAC/ABAC, finance, HR, and organizational management")
// 	Version("1.0")
//
// 	Server("erp", func() {
// 		Host("production", func() {
// 			URI("https://api.erp.example.com")
// 		})
// 		Host("development", func() {
// 			URI("http://localhost:8080")
// 		})
// 	})
// })
//
// // ============================================================================
// // SECURITY DEFINITIONS
// // ============================================================================
//
// var JWT = JWTSecurity("jwt", func() {
// 	Description("JWT-based authentication")
// 	Scope("api:read", "Read access to API resources")
// 	Scope("api:write", "Write access to API resources")
// 	Scope("admin", "Administrative access")
// })
//
// // ============================================================================
// // COMMON TYPES
// // ============================================================================
//
// var UUID = func() {
// 	Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
// 	Example("550e8400-e29b-41d4-a716-446655440000")
// }
//
// var Timestamp = func() {
// 	Format(FormatDateTime)
// 	Example("2025-01-15T10:30:00Z")
// }
//
// var PaginationParams = Type("PaginationParams", func() {
// 	Field(1, "page", Int, "Page number", func() {
// 		Default(1)
// 		Minimum(1)
// 	})
// 	Field(2, "per_page", Int, "Items per page", func() {
// 		Default(20)
// 		Minimum(1)
// 		Maximum(100)
// 	})
// 	Required("page", "per_page")
// })
//
// var PaginationResponse = Type("PaginationResponse", func() {
// 	Field(1, "total", Int, "Total number of items")
// 	Field(2, "page", Int, "Current page number")
// 	Field(3, "per_page", Int, "Items per page")
// 	Field(4, "total_pages", Int, "Total number of pages")
// 	Required("total", "page", "per_page", "total_pages")
// })
//
// // ============================================================================
// // TENANT MANAGEMENT
// // ============================================================================
//
// var Tenant = ResultType("application/vnd.erp.tenant", func() {
// 	Description("Business tenant with multi-tenancy support")
//
// 	Attributes(func() {
// 		Field(1, "id", String, "Tenant UUID", UUID)
// 		Field(2, "slug", String, "URL-friendly identifier", func() {
// 			Pattern("^[a-z0-9-]+$")
// 			MinLength(3)
// 			MaxLength(50)
// 		})
// 		Field(3, "name", String, "Tenant name")
// 		Field(4, "email", String, "Contact email", func() {
// 			Format(FormatEmail)
// 		})
// 		Field(5, "subdomain", String, "Custom subdomain")
// 		Field(6, "status", String, "Tenant status", func() {
// 			Enum("ACTIVE", "SUSPENDED", "TRIAL", "ARCHIVED")
// 		})
// 		Field(7, "timezone", String, "Default timezone", func() {
// 			Default("UTC")
// 		})
// 		Field(8, "currency_code", String, "Default currency", func() {
// 			Pattern("^[A-Z]{3}$")
// 			Default("USD")
// 		})
// 		Field(9, "industry", String, "Industry sector")
// 		Field(10, "company_size", String, "Company size category", func() {
// 			Enum("SMALL", "MEDIUM", "LARGE", "ENTERPRISE")
// 		})
// 		Field(11, "metadata", MapOf(String, Any), "Additional metadata")
// 		Field(12, "settings", MapOf(String, Any), "Tenant configuration")
// 		Field(13, "created_at", String, "Creation timestamp", Timestamp)
// 		Field(14, "updated_at", String, "Last update timestamp", Timestamp)
//
// 		Required("id", "slug", "name", "email", "status", "timezone", "currency_code")
// 	})
//
// 	View("default", func() {
// 		Attribute("id")
// 		Attribute("slug")
// 		Attribute("name")
// 		Attribute("email")
// 		Attribute("status")
// 		Attribute("timezone")
// 		Attribute("currency_code")
// 		Attribute("created_at")
// 		Attribute("updated_at")
// 	})
//
// 	View("detailed", func() {
// 		Attribute("id")
// 		Attribute("slug")
// 		Attribute("name")
// 		Attribute("email")
// 		Attribute("subdomain")
// 		Attribute("status")
// 		Attribute("timezone")
// 		Attribute("currency_code")
// 		Attribute("industry")
// 		Attribute("company_size")
// 		Attribute("metadata")
// 		Attribute("settings")
// 		Attribute("created_at")
// 		Attribute("updated_at")
// 	})
// })
//
// var TenantConfiguration = Type("TenantConfiguration", func() {
// 	Field(1, "max_users", Int, "Maximum users allowed")
// 	Field(2, "max_entities", Int, "Maximum entities allowed")
// 	Field(3, "max_transactions_per_month", Int, "Transaction limit per month")
// 	Field(4, "storage_quota", Int64, "Storage quota in bytes")
// 	Field(5, "accounting_method", String, "Accounting method", func() {
// 		Enum("ACCRUAL", "CASH")
// 	})
// 	Field(6, "fiscal_year_start_month", Int, "Fiscal year start month", func() {
// 		Minimum(1)
// 		Maximum(12)
// 	})
// 	Field(7, "default_currency", String, "Default currency code")
// 	Field(8, "password_policy", MapOf(String, Any), "Password requirements")
// 	Field(9, "api_rate_limits", MapOf(String, Any), "API rate limiting config")
// })
//
// var _ = Service("tenants", func() {
// 	Description("Tenant management and configuration")
// 	Security(JWT)
//
// 	Error("unauthorized", String, "Unauthorized access")
// 	Error("not_found", String, "Tenant not found")
// 	Error("invalid_input", String, "Invalid input data")
//
// 	Method("list", func() {
// 		Description("List all tenants with pagination")
// 		Payload(func() {
// 			Extend(PaginationParams)
// 			Field(1, "status", String, "Filter by status")
// 			Field(2, "search", String, "Search in name/email")
// 		})
// 		Result(func() {
// 			Field(1, "data", CollectionOf(Tenant))
// 			Field(2, "pagination", PaginationResponse)
// 			Required("data", "pagination")
// 		})
// 		HTTP(func() {
// 			GET("/tenants")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get", func() {
// 		Description("Get tenant by ID")
// 		Payload(func() {
// 			Field(1, "id", String, "Tenant ID", UUID)
// 			Required("id")
// 		})
// 		Result(Tenant, "detailed")
// 		HTTP(func() {
// 			GET("/tenants/{id}")
// 			Response(StatusOK)
// 			Response("not_found", StatusNotFound)
// 		})
// 	})
//
// 	Method("create", func() {
// 		Description("Create new tenant")
// 		Payload(func() {
// 			Field(1, "slug", String, "URL-friendly identifier")
// 			Field(2, "name", String, "Tenant name")
// 			Field(3, "email", String, "Contact email")
// 			Field(4, "timezone", String, "Default timezone")
// 			Field(5, "currency_code", String, "Default currency")
// 			Field(6, "industry", String, "Industry sector")
// 			Field(7, "company_size", String, "Company size")
// 			Required("slug", "name", "email")
// 		})
// 		Result(Tenant, "detailed")
// 		HTTP(func() {
// 			POST("/tenants")
// 			Response(StatusCreated)
// 			Response("invalid_input", StatusBadRequest)
// 		})
// 	})
//
// 	Method("update", func() {
// 		Description("Update tenant information")
// 		Payload(func() {
// 			Field(1, "id", String, "Tenant ID", UUID)
// 			Field(2, "name", String, "Tenant name")
// 			Field(3, "email", String, "Contact email")
// 			Field(4, "status", String, "Tenant status")
// 			Field(5, "settings", MapOf(String, Any), "Configuration")
// 			Required("id")
// 		})
// 		Result(Tenant, "detailed")
// 		HTTP(func() {
// 			PATCH("/tenants/{id}")
// 			Response(StatusOK)
// 			Response("not_found", StatusNotFound)
// 		})
// 	})
//
// 	Method("get_configuration", func() {
// 		Description("Get tenant configuration")
// 		Payload(func() {
// 			Field(1, "id", String, "Tenant ID", UUID)
// 			Required("id")
// 		})
// 		Result(TenantConfiguration)
// 		HTTP(func() {
// 			GET("/tenants/{id}/configuration")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("update_configuration", func() {
// 		Description("Update tenant configuration")
// 		Payload(func() {
// 			Field(1, "id", String, "Tenant ID", UUID)
// 			Extend(TenantConfiguration)
// 			Required("id")
// 		})
// 		Result(TenantConfiguration)
// 		HTTP(func() {
// 			PUT("/tenants/{id}/configuration")
// 			Response(StatusOK)
// 		})
// 	})
// })
//
// // ============================================================================
// // ENTITY MANAGEMENT
// // ============================================================================
//
// var Entity = ResultType("application/vnd.erp.entity", func() {
// 	Description("Organizational entity (company, department, etc.)")
//
// 	Attributes(func() {
// 		Field(1, "uuid", String, "Entity UUID", UUID)
// 		Field(2, "tenant_id", String, "Tenant ID", UUID)
// 		Field(3, "parent_id", String, "Parent entity ID", UUID)
// 		Field(4, "name", String, "Entity name")
// 		Field(5, "code", String, "Internal reference code")
// 		Field(6, "type", String, "Entity type", func() {
// 			Enum("COMPANY", "REGIONAL", "DEPARTMENT", "COST_CENTER", "PROJECT", "DIVISION")
// 		})
// 		Field(7, "is_active", Boolean, "Active status")
// 		Field(8, "hidden", Boolean, "Visibility flag")
// 		Field(9, "accrual_method", Boolean, "Accounting method")
// 		Field(10, "fy_start_month", Int, "Fiscal year start month")
// 		Field(11, "address", MapOf(String, Any), "Physical address")
// 		Field(12, "settings", MapOf(String, Any), "Entity settings")
// 		Field(13, "metadata", MapOf(String, Any), "Additional data")
// 		Field(14, "validation_status", String, "Validation status")
// 		Field(15, "created_at", String, "Creation timestamp", Timestamp)
// 		Field(16, "updated_at", String, "Update timestamp", Timestamp)
//
// 		Required("uuid", "tenant_id", "name", "type", "is_active")
// 	})
//
// 	View("default", func() {
// 		Attribute("uuid")
// 		Attribute("tenant_id")
// 		Attribute("parent_id")
// 		Attribute("name")
// 		Attribute("code")
// 		Attribute("type")
// 		Attribute("is_active")
// 	})
//
// 	View("detailed", func() {
// 		Attribute("uuid")
// 		Attribute("tenant_id")
// 		Attribute("parent_id")
// 		Attribute("name")
// 		Attribute("code")
// 		Attribute("type")
// 		Attribute("is_active")
// 		Attribute("hidden")
// 		Attribute("accrual_method")
// 		Attribute("fy_start_month")
// 		Attribute("address")
// 		Attribute("settings")
// 		Attribute("metadata")
// 		Attribute("validation_status")
// 		Attribute("created_at")
// 		Attribute("updated_at")
// 	})
// })
//
// var EntityHierarchy = Type("EntityHierarchy", func() {
// 	Field(1, "entity_id", String, "Entity ID", UUID)
// 	Field(2, "ancestor_id", String, "Ancestor ID", UUID)
// 	Field(3, "descendant_id", String, "Descendant ID", UUID)
// 	Field(4, "depth", Int, "Hierarchical depth")
// })
//
// var _ = Service("entities", func() {
// 	Description("Organizational entity management")
// 	Security(JWT)
//
// 	Error("unauthorized", String)
// 	Error("not_found", String)
// 	Error("invalid_input", String)
//
// 	Method("list", func() {
// 		Payload(func() {
// 			Extend(PaginationParams)
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "parent_id", String, "Parent entity ID", UUID)
// 			Field(3, "type", String, "Entity type")
// 			Field(4, "is_active", Boolean, "Filter active only")
// 			Required("tenant_id")
// 		})
// 		Result(func() {
// 			Field(1, "data", CollectionOf(Entity))
// 			Field(2, "pagination", PaginationResponse)
// 			Required("data", "pagination")
// 		})
// 		HTTP(func() {
// 			GET("/entities")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Entity UUID", UUID)
// 			Required("id")
// 		})
// 		Result(Entity, "detailed")
// 		HTTP(func() {
// 			GET("/entities/{id}")
// 			Response(StatusOK)
// 			Response("not_found", StatusNotFound)
// 		})
// 	})
//
// 	Method("create", func() {
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "parent_id", String, "Parent entity ID", UUID)
// 			Field(3, "name", String, "Entity name")
// 			Field(4, "code", String, "Reference code")
// 			Field(5, "type", String, "Entity type")
// 			Field(6, "fy_start_month", Int, "Fiscal year start")
// 			Field(7, "address", MapOf(String, Any), "Address")
// 			Required("tenant_id", "name", "type")
// 		})
// 		Result(Entity, "detailed")
// 		HTTP(func() {
// 			POST("/entities")
// 			Response(StatusCreated)
// 		})
// 	})
//
// 	Method("update", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Entity UUID", UUID)
// 			Field(2, "name", String, "Entity name")
// 			Field(3, "code", String, "Reference code")
// 			Field(4, "is_active", Boolean, "Active status")
// 			Field(5, "settings", MapOf(String, Any), "Settings")
// 			Required("id")
// 		})
// 		Result(Entity, "detailed")
// 		HTTP(func() {
// 			PATCH("/entities/{id}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get_hierarchy", func() {
// 		Description("Get entity hierarchy")
// 		Payload(func() {
// 			Field(1, "id", String, "Entity UUID", UUID)
// 			Field(2, "direction", String, "Hierarchy direction", func() {
// 				Enum("ancestors", "descendants", "both")
// 				Default("descendants")
// 			})
// 			Required("id")
// 		})
// 		Result(CollectionOf(EntityHierarchy))
// 		HTTP(func() {
// 			GET("/entities/{id}/hierarchy")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("delete", func() {
// 		Description("Soft delete entity")
// 		Payload(func() {
// 			Field(1, "id", String, "Entity UUID", UUID)
// 			Required("id")
// 		})
// 		Result(Empty)
// 		HTTP(func() {
// 			DELETE("/entities/{id}")
// 			Response(StatusNoContent)
// 		})
// 	})
// })
//
// // ============================================================================
// // USER & AUTHENTICATION
// // ============================================================================
//
// var User = ResultType("application/vnd.erp.user", func() {
// 	Attributes(func() {
// 		Field(1, "id", String, "User UUID", UUID)
// 		Field(2, "tenant_id", String, "Tenant ID", UUID)
// 		Field(3, "entity_id", String, "Entity ID", UUID)
// 		Field(4, "person_id", String, "Person ID", UUID)
// 		Field(5, "employee_id", String, "Employee ID", UUID)
// 		Field(6, "email", String, "Email address", func() {
// 			Format(FormatEmail)
// 		})
// 		Field(7, "username", String, "Username")
// 		Field(8, "user_type", String, "User type", func() {
// 			Enum("INTERNAL", "CUSTOMER", "VENDOR", "PARTNER", "API", "SERVICE", "ADMIN")
// 		})
// 		Field(9, "account_status", String, "Account status")
// 		Field(10, "is_active", Boolean, "Active flag")
// 		Field(11, "last_login_at", String, "Last login", Timestamp)
// 		Field(12, "mfa_enabled", Boolean, "MFA enabled")
// 		Field(13, "session_timeout_minutes", Int, "Session timeout")
// 		Field(14, "roles", ArrayOf(String), "Assigned role names")
// 		Field(15, "created_at", String, "Created at", Timestamp)
// 		Field(16, "updated_at", String, "Updated at", Timestamp)
//
// 		Required("id", "tenant_id", "email", "username", "user_type", "is_active")
// 	})
//
// 	View("default", func() {
// 		Attribute("id")
// 		Attribute("email")
// 		Attribute("username")
// 		Attribute("user_type")
// 		Attribute("is_active")
// 		Attribute("roles")
// 	})
//
// 	View("detailed", func() {
// 		Attribute("id")
// 		Attribute("tenant_id")
// 		Attribute("entity_id")
// 		Attribute("person_id")
// 		Attribute("employee_id")
// 		Attribute("email")
// 		Attribute("username")
// 		Attribute("user_type")
// 		Attribute("account_status")
// 		Attribute("is_active")
// 		Attribute("last_login_at")
// 		Attribute("mfa_enabled")
// 		Attribute("session_timeout_minutes")
// 		Attribute("roles")
// 		Attribute("created_at")
// 		Attribute("updated_at")
// 	})
// })
//
// var AuthResponse = Type("AuthResponse", func() {
// 	Field(1, "access_token", String, "JWT access token")
// 	Field(2, "refresh_token", String, "Refresh token")
// 	Field(3, "token_type", String, "Token type", func() {
// 		Default("Bearer")
// 	})
// 	Field(4, "expires_in", Int, "Token expiry in seconds")
// 	Field(5, "user", User)
// 	Required("access_token", "token_type", "expires_in", "user")
// })
//
// var _ = Service("auth", func() {
// 	Description("Authentication and authorization")
//
// 	Error("unauthorized", String)
// 	Error("invalid_credentials", String)
// 	Error("account_locked", String)
//
// 	Method("login", func() {
// 		Description("Authenticate user")
// 		Payload(func() {
// 			Field(1, "email", String, "Email or username")
// 			Field(2, "password", String, "Password")
// 			Field(3, "mfa_code", String, "MFA code if enabled")
// 			Required("email", "password")
// 		})
// 		Result(AuthResponse)
// 		HTTP(func() {
// 			POST("/auth/login")
// 			Response(StatusOK)
// 			Response("invalid_credentials", StatusUnauthorized)
// 			Response("account_locked", StatusForbidden)
// 		})
// 	})
//
// 	Method("refresh", func() {
// 		Description("Refresh access token")
// 		Payload(func() {
// 			Field(1, "refresh_token", String, "Refresh token")
// 			Required("refresh_token")
// 		})
// 		Result(AuthResponse)
// 		HTTP(func() {
// 			POST("/auth/refresh")
// 			Response(StatusOK)
// 			Response("unauthorized", StatusUnauthorized)
// 		})
// 	})
//
// 	Method("logout", func() {
// 		Security(JWT)
// 		Description("Logout and invalidate token")
// 		Result(Empty)
// 		HTTP(func() {
// 			POST("/auth/logout")
// 			Response(StatusNoContent)
// 		})
// 	})
//
// 	Method("me", func() {
// 		Security(JWT)
// 		Description("Get current user info")
// 		Result(User, "detailed")
// 		HTTP(func() {
// 			GET("/auth/me")
// 			Response(StatusOK)
// 		})
// 	})
// })
//
// var _ = Service("users", func() {
// 	Security(JWT)
// 	Description("User management")
//
// 	Error("unauthorized", String)
// 	Error("not_found", String)
// 	Error("invalid_input", String)
//
// 	Method("list", func() {
// 		Payload(func() {
// 			Extend(PaginationParams)
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "user_type", String, "User type filter")
// 			Field(4, "is_active", Boolean, "Active status filter")
// 			Field(5, "search", String, "Search term")
// 		})
// 		Result(func() {
// 			Field(1, "data", CollectionOf(User))
// 			Field(2, "pagination", PaginationResponse)
// 			Required("data", "pagination")
// 		})
// 		HTTP(func() {
// 			GET("/users")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "User ID", UUID)
// 			Required("id")
// 		})
// 		Result(User, "detailed")
// 		HTTP(func() {
// 			GET("/users/{id}")
// 			Response(StatusOK)
// 			Response("not_found", StatusNotFound)
// 		})
// 	})
//
// 	Method("create", func() {
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "email", String, "Email address")
// 			Field(4, "username", String, "Username")
// 			Field(5, "password", String, "Password")
// 			Field(6, "user_type", String, "User type")
// 			Field(7, "role_ids", ArrayOf(String), "Role IDs", func() {
// 				Elem(func() {
// 					Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
// 				})
// 			})
// 			Required("tenant_id", "entity_id", "email", "username", "password", "user_type")
// 		})
// 		Result(User, "detailed")
// 		HTTP(func() {
// 			POST("/users")
// 			Response(StatusCreated)
// 		})
// 	})
//
// 	Method("update", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "User ID", UUID)
// 			Field(2, "email", String, "Email address")
// 			Field(3, "account_status", String, "Account status")
// 			Field(4, "is_active", Boolean, "Active flag")
// 			Field(5, "mfa_enabled", Boolean, "Enable MFA")
// 			Required("id")
// 		})
// 		Result(User, "detailed")
// 		HTTP(func() {
// 			PATCH("/users/{id}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("delete", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "User ID", UUID)
// 			Required("id")
// 		})
// 		Result(Empty)
// 		HTTP(func() {
// 			DELETE("/users/{id}")
// 			Response(StatusNoContent)
// 		})
// 	})
// })
//
// // ============================================================================
// // RBAC/ABAC - ROLES & PERMISSIONS
// // ============================================================================
//
// var Role = ResultType("application/vnd.erp.role", func() {
// 	Attributes(func() {
// 		Field(1, "id", String, "Role UUID", UUID)
// 		Field(2, "tenant_id", String, "Tenant ID", UUID)
// 		Field(3, "entity_id", String, "Entity ID", UUID)
// 		Field(4, "name", String, "Role name")
// 		Field(5, "display_name", String, "Display name")
// 		Field(6, "description", String, "Description")
// 		Field(7, "role_type", String, "Role type", func() {
// 			Enum("SYSTEM", "TENANT", "ENTITY", "CUSTOM", "FUNCTIONAL")
// 		})
// 		Field(8, "parent_role_id", String, "Parent role", UUID)
// 		Field(9, "level", Int, "Hierarchy level")
// 		Field(10, "is_system_role", Boolean, "System role flag")
// 		Field(11, "is_active", Boolean, "Active flag")
// 		Field(12, "permission_count", Int, "Number of permissions")
// 		Field(13, "created_at", String, "Created at", Timestamp)
//
// 		Required("id", "tenant_id", "entity_id", "name")
// 	})
//
// 	View("default", func() {
// 		Attribute("id")
// 		Attribute("name")
// 		Attribute("display_name")
// 		Attribute("role_type")
// 		Attribute("is_active")
// 	})
// })
//
// var Permission = ResultType("application/vnd.erp.permission", func() {
// 	Attributes(func() {
// 		Field(1, "id", String, "Permission UUID", UUID)
// 		Field(2, "tenant_id", String, "Tenant ID", UUID)
// 		Field(3, "resource_id", String, "Resource ID", UUID)
// 		Field(4, "action_id", String, "Action ID", UUID)
// 		Field(5, "name", String, "Permission name")
// 		Field(6, "display_name", String, "Display name")
// 		Field(7, "description", String, "Description")
// 		Field(8, "effect", String, "Permission effect", func() {
// 			Enum("ALLOW", "DENY")
// 		})
// 		Field(9, "conditions", MapOf(String, Any), "ABAC conditions")
// 		Field(10, "is_active", Boolean, "Active flag")
//
// 		Required("id", "tenant_id", "resource_id", "action_id", "name")
// 	})
// })
//
// var _ = Service("roles", func() {
// 	Security(JWT)
// 	Description("Role management")
//
// 	Error("unauthorized", String)
// 	Error("not_found", String)
// 	Error("invalid_input", String)
//
// 	Method("list", func() {
// 		Payload(func() {
// 			Extend(PaginationParams)
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "role_type", String, "Role type filter")
// 			Field(4, "is_active", Boolean, "Active filter")
// 		})
// 		Result(func() {
// 			Field(1, "data", CollectionOf(Role))
// 			Field(2, "pagination", PaginationResponse)
// 			Required("data", "pagination")
// 		})
// 		HTTP(func() {
// 			GET("/roles")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Role ID", UUID)
// 			Required("id")
// 		})
// 		Result(Role)
// 		HTTP(func() {
// 			GET("/roles/{id}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("create", func() {
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "name", String, "Role name")
// 			Field(4, "display_name", String, "Display name")
// 			Field(5, "description", String, "Description")
// 			Field(6, "role_type", String, "Role type")
// 			Field(7, "parent_role_id", String, "Parent role", UUID)
// 			Required("tenant_id", "entity_id", "name")
// 		})
// 		Result(Role)
// 		HTTP(func() {
// 			POST("/roles")
// 			Response(StatusCreated)
// 		})
// 	})
//
// 	Method("assign_permissions", func() {
// 		Description("Assign permissions to role")
// 		Payload(func() {
// 			Field(1, "role_id", String, "Role ID", UUID)
// 			Field(2, "permission_ids", ArrayOf(String), "Permission IDs", func() {
// 				Elem(func() {
// 					Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
// 				})
// 				MinLength(1)
// 			})
// 			Required("role_id", "permission_ids")
// 		})
// 		Result(Empty)
// 		HTTP(func() {
// 			POST("/roles/{role_id}/permissions")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get_permissions", func() {
// 		Description("Get role permissions")
// 		Payload(func() {
// 			Field(1, "role_id", String, "Role ID", UUID)
// 			Required("role_id")
// 		})
// 		Result(CollectionOf(Permission))
// 		HTTP(func() {
// 			GET("/roles/{role_id}/permissions")
// 			Response(StatusOK)
// 		})
// 	})
// })
//
// // ============================================================================
// // FINANCE - ACCOUNTS & TRANSACTIONS
// // ============================================================================
//
// var FinanceAccount = ResultType("application/vnd.erp.finance.account", func() {
// 	Attributes(func() {
// 		Field(1, "id", String, "Account UUID", UUID)
// 		Field(2, "tenant_id", String, "Tenant ID", UUID)
// 		Field(3, "entity_id", String, "Entity ID", UUID)
// 		Field(4, "account_code", String, "Account code")
// 		Field(5, "account_name", String, "Account name")
// 		Field(6, "account_description", String, "Description")
// 		Field(7, "root_type", String, "Root type", func() {
// 			Enum("ASSET", "LIABILITY", "EQUITY", "INCOME", "EXPENSE")
// 		})
// 		Field(8, "account_type", String, "Account type")
// 		Field(9, "account_category", String, "Category")
// 		Field(10, "normal_balance", String, "Normal balance", func() {
// 			Enum("DEBIT", "CREDIT")
// 		})
// 		Field(11, "current_balance", String, "Current balance")
// 		Field(12, "is_active", Boolean, "Active status")
// 		Field(13, "is_control_account", Boolean, "Control account flag")
// 		Field(14, "parent_account_id", String, "Parent account", UUID)
// 		Field(15, "account_level", Int, "Hierarchy level")
// 		Field(16, "has_children", Boolean, "Has sub-accounts")
// 		Field(17, "created_at", String, "Created at", Timestamp)
// 		Field(18, "updated_at", String, "Updated at", Timestamp)
//
// 		Required("id", "tenant_id", "account_code", "account_name", "root_type", "account_type", "normal_balance")
// 	})
//
// 	View("default", func() {
// 		Attribute("id")
// 		Attribute("account_code")
// 		Attribute("account_name")
// 		Attribute("root_type")
// 		Attribute("account_type")
// 		Attribute("current_balance")
// 		Attribute("is_active")
// 	})
// })
//
// var FinanceTransaction = ResultType("application/vnd.erp.finance.transaction", func() {
// 	Attributes(func() {
// 		Field(1, "id", String, "Transaction UUID", UUID)
// 		Field(2, "tenant_id", String, "Tenant ID", UUID)
// 		Field(3, "entity_id", String, "Entity ID", UUID)
// 		Field(4, "transaction_number", String, "Transaction number")
// 		Field(5, "transaction_type", String, "Transaction type", func() {
// 			Enum("JOURNAL_ENTRY", "PAYMENT", "RECEIPT", "TRANSFER", "ADJUSTMENT", "ACCRUAL", "REVERSAL")
// 		})
// 		Field(6, "transaction_status", String, "Status", func() {
// 			Enum("DRAFT", "PENDING", "APPROVED", "POSTED", "VOID", "REVERSED")
// 		})
// 		Field(7, "transaction_date", String, "Transaction date", Timestamp)
// 		Field(8, "posting_date", String, "Posting date", Timestamp)
// 		Field(9, "description", String, "Description")
// 		Field(10, "reference_number", String, "Reference number")
// 		Field(11, "currency_code", String, "Currency code")
// 		Field(12, "exchange_rate", String, "Exchange rate")
// 		Field(13, "total_debit_amount", String, "Total debits")
// 		Field(14, "total_credit_amount", String, "Total credits")
// 		Field(15, "approval_required", Boolean, "Requires approval")
// 		Field(16, "approval_status", String, "Approval status")
// 		Field(17, "is_reversed", Boolean, "Reversed flag")
// 		Field(18, "validation_status", String, "Validation status")
// 		Field(19, "created_by", String, "Creator ID", UUID)
// 		Field(20, "created_at", String, "Created at", Timestamp)
// 		Field(21, "posted_at", String, "Posted at", Timestamp)
//
// 		Required("id", "tenant_id", "transaction_number", "transaction_type", "transaction_status", "transaction_date", "description", "currency_code")
// 	})
//
// 	View("default", func() {
// 		Attribute("id")
// 		Attribute("transaction_number")
// 		Attribute("transaction_type")
// 		Attribute("transaction_status")
// 		Attribute("transaction_date")
// 		Attribute("description")
// 		Attribute("total_debit_amount")
// 		Attribute("total_credit_amount")
// 	})
// })
//
// var TransactionEntry = Type("TransactionEntry", func() {
// 	Field(1, "id", String, "Entry UUID", UUID)
// 	Field(2, "entry_number", Int, "Entry sequence number")
// 	Field(3, "account_id", String, "Account ID", UUID)
// 	Field(4, "account_code", String, "Account code")
// 	Field(5, "account_name", String, "Account name")
// 	Field(6, "debit_amount", String, "Debit amount")
// 	Field(7, "credit_amount", String, "Credit amount")
// 	Field(8, "description", String, "Entry description")
// 	Field(9, "reference", String, "Reference")
// 	Field(10, "cost_center", String, "Cost center")
// 	Field(11, "department", String, "Department")
// 	Field(12, "project_id", String, "Project ID", UUID)
//
// 	Required("account_id", "description")
// })
//
// var _ = Service("finance_accounts", func() {
// 	Security(JWT)
// 	Description("Chart of accounts management")
//
// 	Error("unauthorized", String)
// 	Error("not_found", String)
// 	Error("invalid_input", String)
//
// 	Method("list", func() {
// 		Payload(func() {
// 			Extend(PaginationParams)
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "root_type", String, "Root type filter")
// 			Field(4, "account_type", String, "Account type filter")
// 			Field(5, "is_active", Boolean, "Active filter")
// 			Field(6, "search", String, "Search term")
// 		})
// 		Result(func() {
// 			Field(1, "data", CollectionOf(FinanceAccount))
// 			Field(2, "pagination", PaginationResponse)
// 			Required("data", "pagination")
// 		})
// 		HTTP(func() {
// 			GET("/finance/accounts")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Account ID", UUID)
// 			Required("id")
// 		})
// 		Result(FinanceAccount)
// 		HTTP(func() {
// 			GET("/finance/accounts/{id}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("create", func() {
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "account_code", String, "Account code")
// 			Field(4, "account_name", String, "Account name")
// 			Field(5, "account_description", String, "Description")
// 			Field(6, "root_type", String, "Root type")
// 			Field(7, "account_type", String, "Account type")
// 			Field(8, "normal_balance", String, "Normal balance")
// 			Field(9, "parent_account_id", String, "Parent account", UUID)
// 			Required("tenant_id", "account_code", "account_name", "root_type", "account_type", "normal_balance")
// 		})
// 		Result(FinanceAccount)
// 		HTTP(func() {
// 			POST("/finance/accounts")
// 			Response(StatusCreated)
// 		})
// 	})
//
// 	Method("update", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Account ID", UUID)
// 			Field(2, "account_name", String, "Account name")
// 			Field(3, "account_description", String, "Description")
// 			Field(4, "is_active", Boolean, "Active status")
// 			Required("id")
// 		})
// 		Result(FinanceAccount)
// 		HTTP(func() {
// 			PATCH("/finance/accounts/{id}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get_balance", func() {
// 		Description("Get account balance for period")
// 		Payload(func() {
// 			Field(1, "id", String, "Account ID", UUID)
// 			Field(2, "start_date", String, "Start date", Timestamp)
// 			Field(3, "end_date", String, "End date", Timestamp)
// 			Required("id")
// 		})
// 		Result(func() {
// 			Field(1, "account_id", String, "Account ID", UUID)
// 			Field(2, "opening_balance", String, "Opening balance")
// 			Field(3, "closing_balance", String, "Closing balance")
// 			Field(4, "period_debits", String, "Period debits")
// 			Field(5, "period_credits", String, "Period credits")
// 			Required("account_id", "opening_balance", "closing_balance")
// 		})
// 		HTTP(func() {
// 			GET("/finance/accounts/{id}/balance")
// 			Response(StatusOK)
// 		})
// 	})
// })
//
// var _ = Service("finance_transactions", func() {
// 	Security(JWT)
// 	Description("Financial transaction management")
//
// 	Error("unauthorized", String)
// 	Error("not_found", String)
// 	Error("invalid_input", String)
// 	Error("validation_error", String)
//
// 	Method("list", func() {
// 		Payload(func() {
// 			Extend(PaginationParams)
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "transaction_type", String, "Type filter")
// 			Field(4, "transaction_status", String, "Status filter")
// 			Field(5, "start_date", String, "Start date", Timestamp)
// 			Field(6, "end_date", String, "End date", Timestamp)
// 			Field(7, "search", String, "Search term")
// 		})
// 		Result(func() {
// 			Field(1, "data", CollectionOf(FinanceTransaction))
// 			Field(2, "pagination", PaginationResponse)
// 			Required("data", "pagination")
// 		})
// 		HTTP(func() {
// 			GET("/finance/transactions")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Transaction ID", UUID)
// 			Field(2, "include_entries", Boolean, "Include entries", func() {
// 				Default(true)
// 			})
// 			Required("id")
// 		})
// 		Result(func() {
// 			Field(1, "transaction", FinanceTransaction)
// 			Field(2, "entries", CollectionOf(TransactionEntry))
// 			Required("transaction")
// 		})
// 		HTTP(func() {
// 			GET("/finance/transactions/{id}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("create", func() {
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "transaction_type", String, "Transaction type")
// 			Field(4, "transaction_date", String, "Transaction date", Timestamp)
// 			Field(5, "description", String, "Description")
// 			Field(6, "reference_number", String, "Reference number")
// 			Field(7, "currency_code", String, "Currency code", func() {
// 				Default("USD")
// 			})
// 			Field(8, "entries", ArrayOf(TransactionEntry), "Transaction entries", func() {
// 				MinLength(2)
// 			})
// 			Required("tenant_id", "transaction_type", "transaction_date", "description", "entries")
// 		})
// 		Result(func() {
// 			Field(1, "transaction", FinanceTransaction)
// 			Field(2, "entries", CollectionOf(TransactionEntry))
// 			Required("transaction", "entries")
// 		})
// 		HTTP(func() {
// 			POST("/finance/transactions")
// 			Response(StatusCreated)
// 			Response("validation_error", StatusBadRequest)
// 		})
// 	})
//
// 	Method("update", func() {
// 		Description("Update draft transaction")
// 		Payload(func() {
// 			Field(1, "id", String, "Transaction ID", UUID)
// 			Field(2, "description", String, "Description")
// 			Field(3, "reference_number", String, "Reference")
// 			Field(4, "entries", ArrayOf(TransactionEntry), "Updated entries")
// 			Required("id")
// 		})
// 		Result(func() {
// 			Field(1, "transaction", FinanceTransaction)
// 			Field(2, "entries", CollectionOf(TransactionEntry))
// 			Required("transaction")
// 		})
// 		HTTP(func() {
// 			PATCH("/finance/transactions/{id}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("post", func() {
// 		Description("Post transaction to general ledger")
// 		Payload(func() {
// 			Field(1, "id", String, "Transaction ID", UUID)
// 			Field(2, "posting_date", String, "Posting date", Timestamp)
// 			Required("id")
// 		})
// 		Result(FinanceTransaction)
// 		HTTP(func() {
// 			POST("/finance/transactions/{id}/post")
// 			Response(StatusOK)
// 			Response("validation_error", StatusBadRequest)
// 		})
// 	})
//
// 	Method("reverse", func() {
// 		Description("Reverse posted transaction")
// 		Payload(func() {
// 			Field(1, "id", String, "Transaction ID", UUID)
// 			Field(2, "reversal_date", String, "Reversal date", Timestamp)
// 			Field(3, "reversal_reason", String, "Reversal reason")
// 			Required("id", "reversal_reason")
// 		})
// 		Result(func() {
// 			Field(1, "original", FinanceTransaction)
// 			Field(2, "reversal", FinanceTransaction)
// 			Required("original", "reversal")
// 		})
// 		HTTP(func() {
// 			POST("/finance/transactions/{id}/reverse")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("delete", func() {
// 		Description("Delete draft transaction")
// 		Payload(func() {
// 			Field(1, "id", String, "Transaction ID", UUID)
// 			Required("id")
// 		})
// 		Result(Empty)
// 		HTTP(func() {
// 			DELETE("/finance/transactions/{id}")
// 			Response(StatusNoContent)
// 		})
// 	})
// })
//
// // ============================================================================
// // PERSON & EMPLOYEE MANAGEMENT
// // ============================================================================
//
// var Person = ResultType("application/vnd.erp.person", func() {
// 	Attributes(func() {
// 		Field(1, "id", String, "Person UUID", UUID)
// 		Field(2, "tenant_id", String, "Tenant ID", UUID)
// 		Field(3, "entity_id", String, "Entity ID", UUID)
// 		Field(4, "person_type", String, "Person type", func() {
// 			Enum("INDIVIDUAL", "EMPLOYEE", "CONTACT", "CUSTOMER", "VENDOR", "CONTRACTOR")
// 		})
// 		Field(5, "first_name", String, "First name")
// 		Field(6, "last_name", String, "Last name")
// 		Field(7, "middle_name", String, "Middle name")
// 		Field(8, "email", String, "Email address", func() {
// 			Format(FormatEmail)
// 		})
// 		Field(9, "phone", String, "Phone number")
// 		Field(10, "birth_date", String, "Date of birth", Timestamp)
// 		Field(11, "national_id", String, "National ID")
// 		Field(12, "tax_id", String, "Tax ID")
// 		Field(13, "address", MapOf(String, Any), "Address")
// 		Field(14, "is_active", Boolean, "Active status")
// 		Field(15, "created_at", String, "Created at", Timestamp)
// 		Field(16, "updated_at", String, "Updated at", Timestamp)
//
// 		Required("id", "tenant_id", "entity_id", "person_type", "first_name", "last_name", "is_active")
// 	})
//
// 	View("default", func() {
// 		Attribute("id")
// 		Attribute("first_name")
// 		Attribute("last_name")
// 		Attribute("email")
// 		Attribute("person_type")
// 		Attribute("is_active")
// 	})
// })
//
// var Employee = ResultType("application/vnd.erp.employee", func() {
// 	Attributes(func() {
// 		Field(1, "id", String, "Employee UUID", UUID)
// 		Field(2, "tenant_id", String, "Tenant ID", UUID)
// 		Field(3, "person_id", String, "Person ID", UUID)
// 		Field(4, "employee_number", String, "Employee number")
// 		Field(5, "entity_id", String, "Entity ID", UUID)
// 		Field(6, "position_title", String, "Job title")
// 		Field(7, "department_id", String, "Department ID", UUID)
// 		Field(8, "manager_id", String, "Manager ID", UUID)
// 		Field(9, "hire_date", String, "Hire date", Timestamp)
// 		Field(10, "termination_date", String, "Termination date", Timestamp)
// 		Field(11, "employment_status", String, "Employment status", func() {
// 			Enum("ACTIVE", "INACTIVE", "TERMINATED", "ON_LEAVE", "SUSPENDED")
// 		})
// 		Field(12, "security_level", Int, "Security clearance level")
// 		Field(13, "created_at", String, "Created at", Timestamp)
// 		Field(14, "updated_at", String, "Updated at", Timestamp)
//
// 		// Person details (joined)
// 		Field(15, "first_name", String, "First name")
// 		Field(16, "last_name", String, "Last name")
// 		Field(17, "email", String, "Email")
//
// 		Required("id", "tenant_id", "person_id", "employee_number", "entity_id", "hire_date")
// 	})
//
// 	View("default", func() {
// 		Attribute("id")
// 		Attribute("employee_number")
// 		Attribute("first_name")
// 		Attribute("last_name")
// 		Attribute("position_title")
// 		Attribute("employment_status")
// 	})
// })
//
// var _ = Service("persons", func() {
// 	Security(JWT)
// 	Description("Person management")
//
// 	Error("unauthorized", String)
// 	Error("not_found", String)
// 	Error("invalid_input", String)
//
// 	Method("list", func() {
// 		Payload(func() {
// 			Extend(PaginationParams)
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "person_type", String, "Person type filter")
// 			Field(4, "is_active", Boolean, "Active filter")
// 			Field(5, "search", String, "Search term")
// 		})
// 		Result(func() {
// 			Field(1, "data", CollectionOf(Person))
// 			Field(2, "pagination", PaginationResponse)
// 			Required("data", "pagination")
// 		})
// 		HTTP(func() {
// 			GET("/persons")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Person ID", UUID)
// 			Required("id")
// 		})
// 		Result(Person)
// 		HTTP(func() {
// 			GET("/persons/{id}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("create", func() {
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "person_type", String, "Person type")
// 			Field(4, "first_name", String, "First name")
// 			Field(5, "last_name", String, "Last name")
// 			Field(6, "middle_name", String, "Middle name")
// 			Field(7, "email", String, "Email")
// 			Field(8, "phone", String, "Phone")
// 			Field(9, "birth_date", String, "Birth date", Timestamp)
// 			Field(10, "address", MapOf(String, Any), "Address")
// 			Required("tenant_id", "entity_id", "person_type", "first_name", "last_name")
// 		})
// 		Result(Person)
// 		HTTP(func() {
// 			POST("/persons")
// 			Response(StatusCreated)
// 		})
// 	})
//
// 	Method("update", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Person ID", UUID)
// 			Field(2, "first_name", String, "First name")
// 			Field(3, "last_name", String, "Last name")
// 			Field(4, "email", String, "Email")
// 			Field(5, "phone", String, "Phone")
// 			Field(6, "address", MapOf(String, Any), "Address")
// 			Required("id")
// 		})
// 		Result(Person)
// 		HTTP(func() {
// 			PATCH("/persons/{id}")
// 			Response(StatusOK)
// 		})
// 	})
// })
//
// var _ = Service("employees", func() {
// 	Security(JWT)
// 	Description("Employee management")
//
// 	Error("unauthorized", String)
// 	Error("not_found", String)
// 	Error("invalid_input", String)
//
// 	Method("list", func() {
// 		Payload(func() {
// 			Extend(PaginationParams)
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "department_id", String, "Department filter", UUID)
// 			Field(4, "employment_status", String, "Status filter")
// 			Field(5, "search", String, "Search term")
// 		})
// 		Result(func() {
// 			Field(1, "data", CollectionOf(Employee))
// 			Field(2, "pagination", PaginationResponse)
// 			Required("data", "pagination")
// 		})
// 		HTTP(func() {
// 			GET("/employees")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Employee ID", UUID)
// 			Required("id")
// 		})
// 		Result(Employee)
// 		HTTP(func() {
// 			GET("/employees/{id}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("create", func() {
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "person_id", String, "Person ID", UUID)
// 			Field(3, "employee_number", String, "Employee number")
// 			Field(4, "entity_id", String, "Entity ID", UUID)
// 			Field(5, "position_title", String, "Job title")
// 			Field(6, "department_id", String, "Department ID", UUID)
// 			Field(7, "manager_id", String, "Manager ID", UUID)
// 			Field(8, "hire_date", String, "Hire date", Timestamp)
// 			Field(9, "security_level", Int, "Security level")
// 			Required("tenant_id", "person_id", "employee_number", "entity_id", "hire_date")
// 		})
// 		Result(Employee)
// 		HTTP(func() {
// 			POST("/employees")
// 			Response(StatusCreated)
// 		})
// 	})
//
// 	Method("update", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Employee ID", UUID)
// 			Field(2, "position_title", String, "Job title")
// 			Field(3, "department_id", String, "Department ID", UUID)
// 			Field(4, "manager_id", String, "Manager ID", UUID)
// 			Field(5, "employment_status", String, "Employment status")
// 			Field(6, "security_level", Int, "Security level")
// 			Required("id")
// 		})
// 		Result(Employee)
// 		HTTP(func() {
// 			PATCH("/employees/{id}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("terminate", func() {
// 		Description("Terminate employment")
// 		Payload(func() {
// 			Field(1, "id", String, "Employee ID", UUID)
// 			Field(2, "termination_date", String, "Termination date", Timestamp)
// 			Field(3, "reason", String, "Termination reason")
// 			Required("id", "termination_date")
// 		})
// 		Result(Employee)
// 		HTTP(func() {
// 			POST("/employees/{id}/terminate")
// 			Response(StatusOK)
// 		})
// 	})
// })
//
// // ============================================================================
// // AUDIT & SECURITY
// // ============================================================================
//
// var AuditLog = ResultType("application/vnd.erp.audit", func() {
// 	Attributes(func() {
// 		Field(1, "id", String, "Audit log UUID", UUID)
// 		Field(2, "tenant_id", String, "Tenant ID", UUID)
// 		Field(3, "event_type", String, "Event type")
// 		Field(4, "event_category", String, "Event category", func() {
// 			Enum("ACCESS", "ADMIN", "DATA", "AUTH", "SYSTEM", "COMPLIANCE")
// 		})
// 		Field(5, "severity", String, "Severity level", func() {
// 			Enum("LOW", "INFO", "WARN", "HIGH", "CRITICAL")
// 		})
// 		Field(6, "user_id", String, "User ID", UUID)
// 		Field(7, "entity_id", String, "Entity ID", UUID)
// 		Field(8, "resource_id", String, "Resource ID", UUID)
// 		Field(9, "action_id", String, "Action ID", UUID)
// 		Field(10, "decision", String, "Access decision", func() {
// 			Enum("ALLOW", "DENY", "ERROR")
// 		})
// 		Field(11, "reason", String, "Decision reason")
// 		Field(12, "risk_score", Int, "Risk score (0-100)")
// 		Field(13, "ip_address", String, "IP address")
// 		Field(14, "user_agent", String, "User agent")
// 		Field(15, "context", MapOf(String, Any), "Additional context")
// 		Field(16, "created_at", String, "Timestamp", Timestamp)
//
// 		Required("id", "tenant_id", "event_type", "created_at")
// 	})
// })
//
// var _ = Service("audit", func() {
// 	Security(JWT)
// 	Description("Audit log and security monitoring")
//
// 	Error("unauthorized", String)
// 	Error("invalid_input", String)
//
// 	Method("list", func() {
// 		Description("Query audit logs")
// 		Payload(func() {
// 			Extend(PaginationParams)
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "user_id", String, "User filter", UUID)
// 			Field(3, "event_category", String, "Category filter")
// 			Field(4, "severity", String, "Severity filter")
// 			Field(5, "decision", String, "Decision filter")
// 			Field(6, "start_date", String, "Start date", Timestamp)
// 			Field(7, "end_date", String, "End date", Timestamp)
// 			Field(8, "min_risk_score", Int, "Minimum risk score")
// 		})
// 		Result(func() {
// 			Field(1, "data", CollectionOf(AuditLog))
// 			Field(2, "pagination", PaginationResponse)
// 			Required("data", "pagination")
// 		})
// 		HTTP(func() {
// 			GET("/audit/logs")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Audit log ID", UUID)
// 			Required("id")
// 		})
// 		Result(AuditLog)
// 		HTTP(func() {
// 			GET("/audit/logs/{id}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get_summary", func() {
// 		Description("Get audit summary statistics")
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "start_date", String, "Start date", Timestamp)
// 			Field(3, "end_date", String, "End date", Timestamp)
// 			Required("tenant_id")
// 		})
// 		Result(func() {
// 			Field(1, "total_events", Int, "Total events")
// 			Field(2, "by_category", MapOf(String, Int), "Events by category")
// 			Field(3, "by_severity", MapOf(String, Int), "Events by severity")
// 			Field(4, "unique_users", Int, "Unique users")
// 			Field(5, "denied_attempts", Int, "Denied access attempts")
// 			Field(6, "high_risk_events", Int, "High risk events")
// 			Required("total_events", "unique_users")
// 		})
// 		HTTP(func() {
// 			GET("/audit/summary")
// 			Response(StatusOK)
// 		})
// 	})
// })
//
// // ============================================================================
// // ACCESS REQUESTS
// // ============================================================================
//
// var AccessRequest = ResultType("application/vnd.erp.access_request", func() {
// 	Attributes(func() {
// 		Field(1, "id", String, "Request UUID", UUID)
// 		Field(2, "tenant_id", String, "Tenant ID", UUID)
// 		Field(3, "requester_id", String, "Requester ID", UUID)
// 		Field(4, "target_user_id", String, "Target user ID", UUID)
// 		Field(5, "entity_id", String, "Entity ID", UUID)
// 		Field(6, "request_type", String, "Request type", func() {
// 			Enum("ROLE_ASSIGNMENT", "PERMISSION_GRANT", "RESOURCE_ACCESS", "ELEVATION")
// 		})
// 		Field(7, "role_id", String, "Role ID", UUID)
// 		Field(8, "permission_id", String, "Permission ID", UUID)
// 		Field(9, "resource_id", String, "Resource ID", UUID)
// 		Field(10, "justification", String, "Justification")
// 		Field(11, "business_reason", String, "Business reason")
// 		Field(12, "duration_hours", Int, "Duration in hours")
// 		Field(13, "approval_status", String, "Approval status", func() {
// 			Enum("PENDING", "APPROVED", "REJECTED", "EXPIRED", "REVOKED")
// 		})
// 		Field(14, "approved_by", String, "Approver ID", UUID)
// 		Field(15, "approved_at", String, "Approval timestamp", Timestamp)
// 		Field(16, "approval_comments", String, "Approval comments")
// 		Field(17, "expires_at", String, "Expiration", Timestamp)
// 		Field(18, "auto_revoke", Boolean, "Auto-revoke on expiry")
// 		Field(19, "created_at", String, "Created at", Timestamp)
//
// 		Required("id", "tenant_id", "requester_id", "entity_id", "request_type", "justification")
// 	})
// })
//
// var _ = Service("access_requests", func() {
// 	Security(JWT)
// 	Description("Access request approval workflow")
//
// 	Error("unauthorized", String)
// 	Error("not_found", String)
// 	Error("invalid_input", String)
//
// 	Method("list", func() {
// 		Payload(func() {
// 			Extend(PaginationParams)
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "requester_id", String, "Requester filter", UUID)
// 			Field(3, "approval_status", String, "Status filter")
// 			Field(4, "request_type", String, "Type filter")
// 		})
// 		Result(func() {
// 			Field(1, "data", CollectionOf(AccessRequest))
// 			Field(2, "pagination", PaginationResponse)
// 			Required("data", "pagination")
// 		})
// 		HTTP(func() {
// 			GET("/access-requests")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Request ID", UUID)
// 			Required("id")
// 		})
// 		Result(AccessRequest)
// 		HTTP(func() {
// 			GET("/access-requests/{id}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("create", func() {
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "target_user_id", String, "Target user ID", UUID)
// 			Field(3, "entity_id", String, "Entity ID", UUID)
// 			Field(4, "request_type", String, "Request type")
// 			Field(5, "role_id", String, "Role ID", UUID)
// 			Field(6, "permission_id", String, "Permission ID", UUID)
// 			Field(7, "resource_id", String, "Resource ID", UUID)
// 			Field(8, "justification", String, "Justification")
// 			Field(9, "business_reason", String, "Business reason")
// 			Field(10, "duration_hours", Int, "Duration in hours")
// 			Field(11, "auto_revoke", Boolean, "Auto-revoke", func() {
// 				Default(true)
// 			})
// 			Required("tenant_id", "entity_id", "request_type", "justification")
// 		})
// 		Result(AccessRequest)
// 		HTTP(func() {
// 			POST("/access-requests")
// 			Response(StatusCreated)
// 		})
// 	})
//
// 	Method("approve", func() {
// 		Description("Approve access request")
// 		Payload(func() {
// 			Field(1, "id", String, "Request ID", UUID)
// 			Field(2, "comments", String, "Approval comments")
// 			Required("id")
// 		})
// 		Result(AccessRequest)
// 		HTTP(func() {
// 			POST("/access-requests/{id}/approve")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("reject", func() {
// 		Description("Reject access request")
// 		Payload(func() {
// 			Field(1, "id", String, "Request ID", UUID)
// 			Field(2, "comments", String, "Rejection reason")
// 			Required("id", "comments")
// 		})
// 		Result(AccessRequest)
// 		HTTP(func() {
// 			POST("/access-requests/{id}/reject")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("revoke", func() {
// 		Description("Revoke approved access")
// 		Payload(func() {
// 			Field(1, "id", String, "Request ID", UUID)
// 			Field(2, "reason", String, "Revocation reason")
// 			Required("id", "reason")
// 		})
// 		Result(AccessRequest)
// 		HTTP(func() {
// 			POST("/access-requests/{id}/revoke")
// 			Response(StatusOK)
// 		})
// 	})
// })
//
// // ============================================================================
// // FEATURE FLAGS & CONFIGURATION
// // ============================================================================
//
// var FeatureFlag = ResultType("application/vnd.erp.feature_flag", func() {
// 	Attributes(func() {
// 		Field(1, "id", String, "Feature flag UUID", UUID)
// 		Field(2, "tenant_id", String, "Tenant ID", UUID)
// 		Field(3, "entity_id", String, "Entity ID", UUID)
// 		Field(4, "name", String, "Feature flag name")
// 		Field(5, "description", String, "Description")
// 		Field(6, "flag_type", String, "Flag type", func() {
// 			Enum("BOOLEAN", "STRING", "NUMBER", "JSON")
// 		})
// 		Field(7, "default_value", Boolean, "Default value")
// 		Field(8, "enabled", Boolean, "Current enabled state")
// 		Field(9, "rollout_percentage", Int, "Rollout percentage", func() {
// 			Minimum(0)
// 			Maximum(100)
// 		})
// 		Field(10, "target_audience", MapOf(String, Any), "Targeting rules")
// 		Field(11, "metadata", MapOf(String, Any), "Additional metadata")
// 		Field(12, "created_at", String, "Created at", Timestamp)
// 		Field(13, "updated_at", String, "Updated at", Timestamp)
//
// 		Required("id", "tenant_id", "name", "flag_type", "default_value")
// 	})
// })
//
// var ConfigDefinition = ResultType("application/vnd.erp.config_definition", func() {
// 	Attributes(func() {
// 		Field(1, "id", String, "Config UUID", UUID)
// 		Field(2, "tenant_id", String, "Tenant ID", UUID)
// 		Field(3, "entity_id", String, "Entity ID", UUID)
// 		Field(4, "module_name", String, "Module name")
// 		Field(5, "config_key", String, "Configuration key")
// 		Field(6, "data_type", String, "Data type", func() {
// 			Enum("string", "integer", "boolean", "decimal", "json")
// 		})
// 		Field(7, "default_value", Any, "Default value")
// 		Field(8, "current_value", Any, "Current value")
// 		Field(9, "validation_rules", MapOf(String, Any), "Validation rules")
// 		Field(10, "description", String, "Description")
// 		Field(11, "is_overridable", Boolean, "Can be overridden")
// 		Field(12, "required_permission", String, "Required permission")
// 		Field(13, "created_at", String, "Created at", Timestamp)
//
// 		Required("id", "tenant_id", "module_name", "config_key", "data_type")
// 	})
// })
//
// var _ = Service("feature_flags", func() {
// 	Security(JWT)
// 	Description("Feature flag management")
//
// 	Error("unauthorized", String)
// 	Error("not_found", String)
// 	Error("invalid_input", String)
//
// 	Method("list", func() {
// 		Payload(func() {
// 			Extend(PaginationParams)
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "flag_type", String, "Type filter")
// 			Field(4, "search", String, "Search term")
// 		})
// 		Result(func() {
// 			Field(1, "data", CollectionOf(FeatureFlag))
// 			Field(2, "pagination", PaginationResponse)
// 			Required("data", "pagination")
// 		})
// 		HTTP(func() {
// 			GET("/feature-flags")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Flag ID", UUID)
// 			Required("id")
// 		})
// 		Result(FeatureFlag)
// 		HTTP(func() {
// 			GET("/feature-flags/{id}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("evaluate", func() {
// 		Description("Evaluate feature flag for current context")
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "flag_name", String, "Flag name")
// 			Field(3, "context", MapOf(String, Any), "Evaluation context")
// 			Required("tenant_id", "flag_name")
// 		})
// 		Result(func() {
// 			Field(1, "flag_name", String, "Flag name")
// 			Field(2, "enabled", Boolean, "Enabled state")
// 			Field(3, "value", Any, "Flag value")
// 			Field(4, "source", String, "Evaluation source", func() {
// 				Enum("DEFAULT", "OVERRIDE", "ROLLOUT", "TARGETING")
// 			})
// 			Required("flag_name", "enabled")
// 		})
// 		HTTP(func() {
// 			POST("/feature-flags/evaluate")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("create", func() {
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "name", String, "Flag name")
// 			Field(4, "description", String, "Description")
// 			Field(5, "flag_type", String, "Flag type")
// 			Field(6, "default_value", Boolean, "Default value")
// 			Field(7, "rollout_percentage", Int, "Rollout percentage")
// 			Required("tenant_id", "name", "flag_type", "default_value")
// 		})
// 		Result(FeatureFlag)
// 		HTTP(func() {
// 			POST("/feature-flags")
// 			Response(StatusCreated)
// 		})
// 	})
//
// 	Method("update", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Flag ID", UUID)
// 			Field(2, "description", String, "Description")
// 			Field(3, "default_value", Boolean, "Default value")
// 			Field(4, "rollout_percentage", Int, "Rollout percentage")
// 			Field(5, "target_audience", MapOf(String, Any), "Targeting rules")
// 			Required("id")
// 		})
// 		Result(FeatureFlag)
// 		HTTP(func() {
// 			PATCH("/feature-flags/{id}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("toggle", func() {
// 		Description("Toggle feature flag on/off")
// 		Payload(func() {
// 			Field(1, "id", String, "Flag ID", UUID)
// 			Field(2, "enabled", Boolean, "Enabled state")
// 			Required("id", "enabled")
// 		})
// 		Result(FeatureFlag)
// 		HTTP(func() {
// 			POST("/feature-flags/{id}/toggle")
// 			Response(StatusOK)
// 		})
// 	})
// })
//
// var _ = Service("configurations", func() {
// 	Security(JWT)
// 	Description("System configuration management")
//
// 	Error("unauthorized", String)
// 	Error("not_found", String)
// 	Error("invalid_input", String)
//
// 	Method("list", func() {
// 		Payload(func() {
// 			Extend(PaginationParams)
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "module_name", String, "Module filter")
// 			Field(4, "search", String, "Search term")
// 		})
// 		Result(func() {
// 			Field(1, "data", CollectionOf(ConfigDefinition))
// 			Field(2, "pagination", PaginationResponse)
// 			Required("data", "pagination")
// 		})
// 		HTTP(func() {
// 			GET("/configurations")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get", func() {
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "config_key", String, "Configuration key")
// 			Required("tenant_id", "config_key")
// 		})
// 		Result(ConfigDefinition)
// 		HTTP(func() {
// 			GET("/configurations/{config_key}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("update", func() {
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "config_key", String, "Configuration key")
// 			Field(4, "value", Any, "New value")
// 			Required("tenant_id", "config_key", "value")
// 		})
// 		Result(ConfigDefinition)
// 		HTTP(func() {
// 			PUT("/configurations/{config_key}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get_audit", func() {
// 		Description("Get configuration change history")
// 		Payload(func() {
// 			Extend(PaginationParams)
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "config_key", String, "Configuration key")
// 			Field(3, "start_date", String, "Start date", Timestamp)
// 			Field(4, "end_date", String, "End date", Timestamp)
// 			Required("tenant_id")
// 		})
// 		Result(func() {
// 			Field(1, "data", CollectionOf(func() {
// 				Field(1, "id", String, "Audit ID", UUID)
// 				Field(2, "config_key", String, "Configuration key")
// 				Field(3, "old_value", Any, "Old value")
// 				Field(4, "new_value", Any, "New value")
// 				Field(5, "operation", String, "Operation type")
// 				Field(6, "user_id", String, "User ID", UUID)
// 				Field(7, "applied_at", String, "Applied at", Timestamp)
// 			}))
// 			Field(2, "pagination", PaginationResponse)
// 			Required("data", "pagination")
// 		})
// 		HTTP(func() {
// 			GET("/configurations/audit")
// 			Response(StatusOK)
// 		})
// 	})
// })
//
// // ============================================================================
// // POLICY ENGINE & ABAC
// // ============================================================================
//
// var Policy = ResultType("application/vnd.erp.policy", func() {
// 	Attributes(func() {
// 		Field(1, "id", String, "Policy UUID", UUID)
// 		Field(2, "tenant_id", String, "Tenant ID", UUID)
// 		Field(3, "entity_id", String, "Entity ID", UUID)
// 		Field(4, "name", String, "Policy name")
// 		Field(5, "display_name", String, "Display name")
// 		Field(6, "description", String, "Description")
// 		Field(7, "policy_type", String, "Policy type", func() {
// 			Enum("ABAC", "RBAC", "HYBRID", "TIME_BASED", "LOCATION_BASED")
// 		})
// 		Field(8, "effect", String, "Policy effect", func() {
// 			Enum("ALLOW", "DENY")
// 		})
// 		Field(9, "priority", Int, "Priority (higher = first)")
// 		Field(10, "category", String, "Policy category", func() {
// 			Enum("ACCESS", "DATA_FILTER", "FIELD_MASK", "AUDIT", "COMPLIANCE")
// 		})
// 		Field(11, "target", MapOf(String, Any), "Target definition")
// 		Field(12, "rule", MapOf(String, Any), "Evaluation rule")
// 		Field(13, "obligations", MapOf(String, Any), "Required actions")
// 		Field(14, "is_active", Boolean, "Active status")
// 		Field(15, "created_at", String, "Created at", Timestamp)
// 		Field(16, "updated_at", String, "Updated at", Timestamp)
//
// 		Required("id", "tenant_id", "name", "policy_type", "effect", "target", "rule")
// 	})
// })
//
// var PolicyEvaluation = Type("PolicyEvaluation", func() {
// 	Field(1, "id", String, "Evaluation UUID", UUID)
// 	Field(2, "tenant_id", String, "Tenant ID", UUID)
// 	Field(3, "user_id", String, "User ID", UUID)
// 	Field(4, "resource_type", String, "Resource type")
// 	Field(5, "resource_id", String, "Resource ID", UUID)
// 	Field(6, "action", String, "Action")
// 	Field(7, "decision", String, "Final decision", func() {
// 		Enum("ALLOW", "DENY", "NOT_APPLICABLE")
// 	})
// 	Field(8, "applicable_policies", ArrayOf(String), "Applied policy IDs")
// 	Field(9, "evaluation_time_ms", Int, "Evaluation time")
// 	Field(10, "evaluated_at", String, "Evaluation timestamp", Timestamp)
//
// 	Required("tenant_id", "user_id", "resource_type", "action", "decision")
// })
//
// var _ = Service("policies", func() {
// 	Security(JWT)
// 	Description("ABAC policy management")
//
// 	Error("unauthorized", String)
// 	Error("not_found", String)
// 	Error("invalid_input", String)
//
// 	Method("list", func() {
// 		Payload(func() {
// 			Extend(PaginationParams)
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "policy_type", String, "Type filter")
// 			Field(4, "category", String, "Category filter")
// 			Field(5, "is_active", Boolean, "Active filter")
// 		})
// 		Result(func() {
// 			Field(1, "data", CollectionOf(Policy))
// 			Field(2, "pagination", PaginationResponse)
// 			Required("data", "pagination")
// 		})
// 		HTTP(func() {
// 			GET("/policies")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("get", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Policy ID", UUID)
// 			Required("id")
// 		})
// 		Result(Policy)
// 		HTTP(func() {
// 			GET("/policies/{id}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("create", func() {
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "name", String, "Policy name")
// 			Field(4, "description", String, "Description")
// 			Field(5, "policy_type", String, "Policy type")
// 			Field(6, "effect", String, "Effect")
// 			Field(7, "priority", Int, "Priority")
// 			Field(8, "category", String, "Category")
// 			Field(9, "target", MapOf(String, Any), "Target")
// 			Field(10, "rule", MapOf(String, Any), "Rule")
// 			Field(11, "obligations", MapOf(String, Any), "Obligations")
// 			Required("tenant_id", "name", "policy_type", "effect", "target", "rule")
// 		})
// 		Result(Policy)
// 		HTTP(func() {
// 			POST("/policies")
// 			Response(StatusCreated)
// 		})
// 	})
//
// 	Method("update", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Policy ID", UUID)
// 			Field(2, "name", String, "Policy name")
// 			Field(3, "description", String, "Description")
// 			Field(4, "priority", Int, "Priority")
// 			Field(5, "target", MapOf(String, Any), "Target")
// 			Field(6, "rule", MapOf(String, Any), "Rule")
// 			Field(7, "is_active", Boolean, "Active status")
// 			Required("id")
// 		})
// 		Result(Policy)
// 		HTTP(func() {
// 			PATCH("/policies/{id}")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("evaluate", func() {
// 		Description("Evaluate policies for access decision")
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "user_id", String, "User ID", UUID)
// 			Field(3, "resource_type", String, "Resource type")
// 			Field(4, "resource_id", String, "Resource ID", UUID)
// 			Field(5, "action", String, "Action to perform")
// 			Field(6, "context", MapOf(String, Any), "Evaluation context")
// 			Required("tenant_id", "user_id", "resource_type", "action")
// 		})
// 		Result(PolicyEvaluation)
// 		HTTP(func() {
// 			POST("/policies/evaluate")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("delete", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Policy ID", UUID)
// 			Required("id")
// 		})
// 		Result(Empty)
// 		HTTP(func() {
// 			DELETE("/policies/{id}")
// 			Response(StatusNoContent)
// 		})
// 	})
// })
//
// // ============================================================================
// // REPORTING & ANALYTICS
// // ============================================================================
//
// var _ = Service("reports", func() {
// 	Security(JWT)
// 	Description("Reporting and analytics")
//
// 	Error("unauthorized", String)
// 	Error("not_found", String)
// 	Error("invalid_input", String)
//
// 	Method("financial_statement", func() {
// 		Description("Generate financial statement")
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "entity_id", String, "Entity ID", UUID)
// 			Field(3, "statement_type", String, "Statement type", func() {
// 				Enum("BALANCE_SHEET", "INCOME_STATEMENT", "CASH_FLOW", "TRIAL_BALANCE")
// 			})
// 			Field(4, "start_date", String, "Start date", Timestamp)
// 			Field(5, "end_date", String, "End date", Timestamp)
// 			Field(6, "include_inactive", Boolean, "Include inactive accounts")
// 			Field(7, "format", String, "Output format", func() {
// 				Enum("JSON", "PDF", "XLSX", "CSV")
// 				Default("JSON")
// 			})
// 			Required("tenant_id", "statement_type", "start_date", "end_date")
// 		})
// 		Result(func() {
// 			Field(1, "statement_type", String, "Statement type")
// 			Field(2, "period", func() {
// 				Field(1, "start_date", String, "Start date", Timestamp)
// 				Field(2, "end_date", String, "End date", Timestamp)
// 			})
// 			Field(3, "entity", func() {
// 				Field(1, "id", String, "Entity ID", UUID)
// 				Field(2, "name", String, "Entity name")
// 			})
// 			Field(4, "sections", ArrayOf(func() {
// 				Field(1, "name", String, "Section name")
// 				Field(2, "total", String, "Section total")
// 				Field(3, "accounts", ArrayOf(func() {
// 					Field(1, "code", String, "Account code")
// 					Field(2, "name", String, "Account name")
// 					Field(3, "balance", String, "Balance")
// 				}))
// 			}))
// 			Field(5, "totals", MapOf(String, String), "Statement totals")
// 			Field(6, "generated_at", String, "Generation timestamp", Timestamp)
// 			Required("statement_type", "period", "sections", "generated_at")
// 		})
// 		HTTP(func() {
// 			POST("/reports/financial-statement")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("user_activity", func() {
// 		Description("User activity report")
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "start_date", String, "Start date", Timestamp)
// 			Field(3, "end_date", String, "End date", Timestamp)
// 			Field(4, "user_id", String, "User filter", UUID)
// 			Field(5, "group_by", String, "Grouping", func() {
// 				Enum("USER", "HOUR", "DAY", "WEEK", "MONTH")
// 				Default("DAY")
// 			})
// 			Required("tenant_id", "start_date", "end_date")
// 		})
// 		Result(func() {
// 			Field(1, "period", func() {
// 				Field(1, "start_date", String, "Start date", Timestamp)
// 				Field(2, "end_date", String, "End date", Timestamp)
// 			})
// 			Field(2, "total_activities", Int, "Total activities")
// 			Field(3, "unique_users", Int, "Unique users")
// 			Field(4, "activities_by_type", MapOf(String, Int))
// 			Field(5, "timeline", ArrayOf(func() {
// 				Field(1, "period", String, "Time period")
// 				Field(2, "count", Int, "Activity count")
// 				Field(3, "unique_users", Int, "Unique users")
// 			}))
// 			Required("period", "total_activities", "unique_users")
// 		})
// 		HTTP(func() {
// 			POST("/reports/user-activity")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("security_dashboard", func() {
// 		Description("Security monitoring dashboard")
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "time_range", String, "Time range", func() {
// 				Enum("1h", "24h", "7d", "30d")
// 				Default("24h")
// 			})
// 			Required("tenant_id")
// 		})
// 		Result(func() {
// 			Field(1, "high_risk_sessions", Int, "High risk session count")
// 			Field(2, "critical_events", Int, "Critical event count")
// 			Field(3, "denied_attempts", Int, "Denied access attempts")
// 			Field(4, "anomaly_alerts", Int, "Anomaly alert count")
// 			Field(5, "top_threats", ArrayOf(func() {
// 				Field(1, "user_id", String, "User ID", UUID)
// 				Field(2, "username", String, "Username")
// 				Field(3, "risk_score", Int, "Risk score")
// 				Field(4, "event_count", Int, "Event count")
// 			}))
// 			Field(6, "events_by_severity", MapOf(String, Int))
// 			Field(7, "events_by_category", MapOf(String, Int))
// 			Required("high_risk_sessions", "critical_events", "denied_attempts")
// 		})
// 		HTTP(func() {
// 			GET("/reports/security-dashboard")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("entity_hierarchy", func() {
// 		Description("Entity hierarchy report")
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "root_entity_id", String, "Root entity", UUID)
// 			Field(3, "include_inactive", Boolean, "Include inactive")
// 			Required("tenant_id")
// 		})
// 		Result(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "hierarchy", ArrayOf(func() {
// 				Field(1, "entity_id", String, "Entity ID", UUID)
// 				Field(2, "name", String, "Entity name")
// 				Field(3, "type", String, "Entity type")
// 				Field(4, "level", Int, "Hierarchy level")
// 				Field(5, "parent_id", String, "Parent ID", UUID)
// 				Field(6, "children", ArrayOf(Any), "Child entities")
// 			}))
// 			Required("tenant_id", "hierarchy")
// 		})
// 		HTTP(func() {
// 			GET("/reports/entity-hierarchy")
// 			Response(StatusOK)
// 		})
// 	})
// })
//
// // ============================================================================
// // NOTIFICATIONS
// // ============================================================================
//
// // var Notification = ResultType("application/vnd.erp.notification", func() {
// // 	Attributes(func() {
// // 		Field(1, "id", String, "Notification UUID", UUID)
// // 		Field(2, "tenant_id", String, "Tenant ID", UUID)
// // 		Field(3, "user_id", String, "User ID", UUID)
// // 		Field(4, "type", String, "Notification type")
// // 		Field(5, "title", String, "Title")
// // 		Field(6, "message", String, "Message")
// // 		Field(7, "metadata", MapOf(String, Any), "Additional data")
// // 		Field(8, "acknowledged", Boolean, "Acknowledged flag")
// // 		Field(9, "created_at", String, "Created at", Timestamp)
// // 		Field(10, "expires_at", String, "Expires at", Timestamp)
// //
// // 		Required("id", "tenant_id", "type", "title", "message", "created_at")
// // 	})
// // })
//
// var _ = Service("notifications", func() {
// 	Security(JWT)
// 	Description("User notifications")
//
// 	Error("unauthorized", String)
// 	Error("not_found", String)
//
// 	Method("list", func() {
// 		Payload(func() {
// 			Extend(PaginationParams)
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "user_id", String, "User ID", UUID)
// 			Field(3, "type", String, "Type filter")
// 			Field(4, "acknowledged", Boolean, "Acknowledged filter")
// 			Required("tenant_id")
// 		})
// 		Result(func() {
// 			Field(1, "data", CollectionOf(Notification))
// 			Field(2, "pagination", PaginationResponse)
// 			Field(3, "unread_count", Int, "Unread notification count")
// 			Required("data", "pagination", "unread_count")
// 		})
// 		HTTP(func() {
// 			GET("/notifications")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("acknowledge", func() {
// 		Payload(func() {
// 			Field(1, "id", String, "Notification ID", UUID)
// 			Required("id")
// 		})
// 		Result(Notification)
// 		HTTP(func() {
// 			POST("/notifications/{id}/acknowledge")
// 			Response(StatusOK)
// 		})
// 	})
//
// 	Method("acknowledge_all", func() {
// 		Payload(func() {
// 			Field(1, "tenant_id", String, "Tenant ID", UUID)
// 			Field(2, "user_id", String, "User ID", UUID)
// 			Required("tenant_id", "user_id")
// 		})
// 		Result(func() {
// 			Field(1, "acknowledged_count", Int, "Count of acknowledged")
// 			Required("acknowledged_count")
// 		})
// 		HTTP(func() {
// 			POST("/notifications/acknowledge-all")
// 			Response(StatusOK)
// 		})
// 	})
// })
