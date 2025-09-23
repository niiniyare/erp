package ui

import (
	"github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// WorkspaceService defines the tenant workspace UI service
var _ = Service("workspace", func() {
	Description("Tenant workspace UI service - provides tenant-specific interface for business operations")

	// Security schemes for workspace access
	Security("jwt", func() {
		Scope("tenant:read")
		Scope("tenant:write")
		Scope("workspace:access")
	})

	// Common error responses
	Error("unauthorized", ErrorResult, "Unauthorized access")
	Error("forbidden", ErrorResult, "Insufficient permissions")
	Error("not_found", ErrorResult, "Resource not found")
	Error("bad_request", ErrorResult, "Invalid request")
	Error("internal_error", ErrorResult, "Internal server error")

	// ========================================================================
	// DASHBOARD ENDPOINTS
	// ========================================================================

	Method("dashboard", func() {
		Description("Render workspace dashboard")
		Security("jwt", func() {
			Scope("tenant:read")
			Scope("workspace:access")
		})

		Payload(func() {
			types.UIHeaders()
			// X-Tenant-ID is automatically set by backend middleware after auth
		})

		Result("WorkspacePageData", func() {
			Description("Workspace dashboard page data")
		})

		HTTP(func() {
			GET("/workspace/dashboard")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("dashboard_data", func() {
		Description("Get dashboard data as JSON")
		Security("jwt", func() {
			Scope("tenant:read")
		})

		Payload(func() {
			types.UIHeaders()
		})

		Result("WorkspaceDashboardResponse")

		HTTP(func() {
			GET("/workspace/api/dashboard")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// ========================================================================
	// USER MANAGEMENT ENDPOINTS (within tenant)
	// ========================================================================

	Method("users_list", func() {
		Description("Render user list page for current tenant")
		Security("jwt", func() {
			Scope("tenant:read")
			Scope("user:list")
		})

		Payload(func() {
			types.UIHeaders()
			types.Pagination()
			types.SearchFilter()
			Attribute("role_filter", String, "Filter by user role", func() {
				Enum("admin", "manager", "user", "viewer", "all")
				Default("all")
				Example("manager")
			})
			Attribute("department_filter", String, "Filter by department", func() {
				Example("Engineering")
			})
			Attribute("status_filter", String, "Filter by user status", func() {
				Enum("active", "inactive", "pending", "suspended", "all")
				Default("all")
				Example("active")
			})
		})

		Result("WorkspacePageData")

		HTTP(func() {
			GET("/workspace/users")
			Param("page", "page")
			Param("page_size", "page_size")
			Param("q", "q")
			Param("role", "role_filter")
			Param("department", "department_filter")
			Param("status", "status_filter")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("users_data", func() {
		Description("Get user list data as JSON")
		Security("jwt", func() {
			Scope("tenant:read")
			Scope("user:list")
		})

		Payload(func() {
			types.UIHeaders()
			types.Pagination()
			types.SearchFilter()
			Attribute("role_filter", String, "Filter by user role", func() {
				Enum("admin", "manager", "user", "viewer", "all")
				Default("all")
			})
			Attribute("department_filter", String, "Filter by department")
			Attribute("status_filter", String, "Filter by user status", func() {
				Enum("active", "inactive", "pending", "suspended", "all")
				Default("all")
			})
		})

		Result("WorkspaceUserListResponse")

		HTTP(func() {
			GET("/workspace/api/users")
			Param("page", "page")
			Param("page_size", "page_size")
			Param("q", "q")
			Param("role", "role_filter")
			Param("department", "department_filter")
			Param("status", "status_filter")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("user_create_form", func() {
		Description("Render user creation form")
		Security("jwt", func() {
			Scope("tenant:write")
			Scope("user:create")
		})

		Payload(func() {
			types.UIHeaders()
		})

		Result("WorkspacePageData")

		HTTP(func() {
			GET("/workspace/users/new")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("user_create", func() {
		Description("Create new user in current tenant")
		Security("jwt", func() {
			Scope("tenant:write")
			Scope("user:create")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("user_data", "WorkspaceUserCreateRequest", "User creation data")
			Required("user_data")
		})

		Result("WorkspaceUser")

		HTTP(func() {
			POST("/workspace/users")
			Body("user_data")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("user_detail", func() {
		Description("Render user detail page")
		Security("jwt", func() {
			Scope("tenant:read")
			Scope("user:read")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("user_id", String, "User ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Required("user_id")
		})

		Result("WorkspacePageData")

		HTTP(func() {
			GET("/workspace/users/{user_id}")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("user_edit_form", func() {
		Description("Render user edit form")
		Security("jwt", func() {
			Scope("tenant:write")
			Scope("user:update")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("user_id", String, "User ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Required("user_id")
		})

		Result("WorkspacePageData")

		HTTP(func() {
			GET("/workspace/users/{user_id}/edit")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("user_update", func() {
		Description("Update user in current tenant")
		Security("jwt", func() {
			Scope("tenant:write")
			Scope("user:update")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("user_id", String, "User ID", func() {
				Format(FormatUUID)
				Example("550e8400-e29b-41d4-a716-446655440000")
			})
			Attribute("user_data", "WorkspaceUserUpdateRequest", "User update data")
			Required("user_id", "user_data")
		})

		Result("WorkspaceUser")

		HTTP(func() {
			PUT("/workspace/users/{user_id}")
			Body("user_data")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// ========================================================================
	// PROJECT MANAGEMENT ENDPOINTS
	// ========================================================================

	Method("projects_list", func() {
		Description("Render project list page")
		Security("jwt", func() {
			Scope("tenant:read")
			Scope("project:list")
		})

		Payload(func() {
			types.UIHeaders()
			types.Pagination()
			types.SearchFilter()
			Attribute("status_filter", String, "Filter by project status", func() {
				Enum("planning", "active", "on_hold", "completed", "cancelled", "all")
				Default("all")
				Example("active")
			})
			Attribute("priority_filter", String, "Filter by priority", func() {
				Enum("low", "medium", "high", "critical", "all")
				Default("all")
				Example("high")
			})
		})

		Result("WorkspacePageData")

		HTTP(func() {
			GET("/workspace/projects")
			Param("page", "page")
			Param("page_size", "page_size")
			Param("q", "q")
			Param("status", "status_filter")
			Param("priority", "priority_filter")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("projects_data", func() {
		Description("Get project list data as JSON")
		Security("jwt", func() {
			Scope("tenant:read")
			Scope("project:list")
		})

		Payload(func() {
			types.UIHeaders()
			types.Pagination()
			types.SearchFilter()
			Attribute("status_filter", String, "Filter by project status", func() {
				Enum("planning", "active", "on_hold", "completed", "cancelled", "all")
				Default("all")
			})
			Attribute("priority_filter", String, "Filter by priority", func() {
				Enum("low", "medium", "high", "critical", "all")
				Default("all")
			})
		})

		Result(func() {
			Attribute("projects", ArrayOf("WorkspaceProject"), "List of projects")
			Attribute("pagination", types.PaginationMeta, "Pagination information")
			Required("projects", "pagination")
		})

		HTTP(func() {
			GET("/workspace/api/projects")
			Param("page", "page")
			Param("page_size", "page_size")
			Param("q", "q")
			Param("status", "status_filter")
			Param("priority", "priority_filter")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// ========================================================================
	// FINANCE ENDPOINTS (Limited tenant view)
	// ========================================================================

	Method("finance_dashboard", func() {
		Description("Render finance dashboard for current tenant")
		Security("jwt", func() {
			Scope("tenant:read")
			Scope("finance:read")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("period", String, "Time period for reports", func() {
				Enum("current_month", "last_month", "current_quarter", "last_quarter", "current_year")
				Default("current_month")
				Example("current_month")
			})
		})

		Result("WorkspacePageData")

		HTTP(func() {
			GET("/workspace/finance")
			Param("period", "period")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("invoices_list", func() {
		Description("Render invoices list page")
		Security("jwt", func() {
			Scope("tenant:read")
			Scope("finance:read")
		})

		Payload(func() {
			types.UIHeaders()
			types.Pagination()
			types.SearchFilter()
			Attribute("status_filter", String, "Filter by invoice status", func() {
				Enum("draft", "sent", "viewed", "paid", "overdue", "cancelled", "all")
				Default("all")
				Example("sent")
			})
		})

		Result("WorkspacePageData")

		HTTP(func() {
			GET("/workspace/finance/invoices")
			Param("page", "page")
			Param("page_size", "page_size")
			Param("q", "q")
			Param("status", "status_filter")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("invoices_data", func() {
		Description("Get invoices data as JSON")
		Security("jwt", func() {
			Scope("tenant:read")
			Scope("finance:read")
		})

		Payload(func() {
			types.UIHeaders()
			types.Pagination()
			types.SearchFilter()
			Attribute("status_filter", String, "Filter by invoice status", func() {
				Enum("draft", "sent", "viewed", "paid", "overdue", "cancelled", "all")
				Default("all")
			})
		})

		Result(func() {
			Attribute("invoices", ArrayOf("WorkspaceInvoice"), "List of invoices")
			Attribute("pagination", types.PaginationMeta, "Pagination information")
			Required("invoices", "pagination")
		})

		HTTP(func() {
			GET("/workspace/api/finance/invoices")
			Param("page", "page")
			Param("page_size", "page_size")
			Param("q", "q")
			Param("status", "status_filter")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// ========================================================================
	// SETTINGS ENDPOINTS
	// ========================================================================

	Method("settings", func() {
		Description("Render tenant settings page")
		Security("jwt", func() {
			Scope("tenant:read")
			Scope("tenant:settings")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("section", String, "Settings section to display", func() {
				Enum("general", "users", "billing", "integrations", "security")
				Default("general")
				Example("general")
			})
		})

		Result("WorkspacePageData")

		HTTP(func() {
			GET("/workspace/settings")
			Param("section", "section")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("update_settings", func() {
		Description("Update tenant settings")
		Security("jwt", func() {
			Scope("tenant:write")
			Scope("tenant:settings")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("settings", MapOf(String, Any), "Settings to update", func() {
				Attribute("name", String, "Tenant name")
				Attribute("timezone", String, "Default timezone")
				Attribute("currency", String, "Default currency")
				Attribute("date_format", String, "Date format preference")
				Attribute("language", String, "Default language")
				Attribute("logo", String, "Logo URL")
			})
			Required("settings")
		})

		Result("WorkspaceTenantInfo")

		HTTP(func() {
			PUT("/workspace/settings")
			Body("settings")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// ========================================================================
	// UTILITY ENDPOINTS
	// ========================================================================

	Method("search_suggestions", func() {
		Description("Get search suggestions for workspace search")
		Security("jwt", func() {
			Scope("tenant:read")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("query", String, "Search query", func() {
				MinLength(1)
				MaxLength(100)
				Example("john")
			})
			Attribute("category", String, "Search category", func() {
				Enum("users", "projects", "invoices", "all")
				Default("all")
				Example("users")
			})
			Required("query")
		})

		Result(func() {
			Attribute("suggestions", ArrayOf(MapOf(String, Any)), "Search suggestions", func() {
				Elem(func() {
					Attribute("type", String, "Suggestion type", func() {
						Example("user")
					})
					Attribute("id", String, "Resource ID", func() {
						Example("550e8400-e29b-41d4-a716-446655440000")
					})
					Attribute("title", String, "Display title", func() {
						Example("John Doe")
					})
					Attribute("subtitle", String, "Display subtitle", func() {
						Example("john.doe@acme.com")
					})
					Attribute("url", String, "Link URL", func() {
						Example("/workspace/users/550e8400-e29b-41d4-a716-446655440000")
					})
					Required("type", "id", "title", "url")
				})
			})
			Required("suggestions")
		})

		HTTP(func() {
			GET("/workspace/api/search/suggestions")
			Param("q", "query")
			Param("category", "category")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	Method("notifications", func() {
		Description("Get user notifications")
		Security("jwt", func() {
			Scope("tenant:read")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("unread_only", Boolean, "Show only unread notifications", func() {
				Default(false)
				Example(true)
			})
		})

		Result(func() {
			Attribute("notifications", ArrayOf(types.UINotification), "User notifications")
			Attribute("unread_count", UInt, "Number of unread notifications")
			Required("notifications", "unread_count")
		})

		HTTP(func() {
			GET("/workspace/api/notifications")
			Param("unread_only", "unread_only")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	// ========================================================================
	// FILES - TEMPLATE SERVING
	// ========================================================================

	Files("/workspace/assets/*filepath", "./web/workspace/assets/", func() {
		Description("Serve workspace static assets")
	})
})
