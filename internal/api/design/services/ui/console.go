package ui

import (
	"github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// ConsoleService defines the admin console UI service
var _ = Service("console", func() {
	Description("Admin console UI service - provides comprehensive system administration interface")

	// Security schemes for console access
	Security("jwt", func() {
		Scope("admin:read")
		Scope("admin:write")
		Scope("console:access")
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
		Description("Render admin console dashboard")
		Security("jwt", func() {
			Scope("admin:read")
			Scope("console:access")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("tenant_id", String, "Optional tenant to impersonate", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
		})

		Result("ConsolePageData", func() {
			Description("Console dashboard page data")
		})

		HTTP(func() {
			GET("/console/dashboard")
			Header("X-Tenant-ID:tenant_id")
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
			Scope("admin:read")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("tenant_id", String, "Optional tenant to impersonate", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
		})

		Result("ConsoleDashboardResponse")

		HTTP(func() {
			GET("/console/api/dashboard")
			Header("X-Tenant-ID:tenant_id")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// ========================================================================
	// TENANT MANAGEMENT ENDPOINTS
	// ========================================================================

	Method("tenants_list", func() {
		Description("Render tenant list page")
		Security("jwt", func() {
			Scope("admin:read")
			Scope("tenant:list")
		})

		Payload(func() {
			types.UIHeaders()
			types.Pagination()
			types.SearchFilter()
			Attribute("status_filter", String, "Filter by tenant status", func() {
				Enum("active", "inactive", "suspended", "pending", "archived", "all")
				Default("all")
				Example("active")
			})
			Attribute("plan_filter", String, "Filter by plan type", func() {
				Enum("startup", "small", "medium", "large", "enterprise", "all")
				Default("all")
				Example("enterprise")
			})
		})

		Result("ConsolePageData")

		HTTP(func() {
			GET("/console/tenants")
			Param("page", "page")
			Param("page_size", "page_size")
			Param("q", "q")
			Param("status", "status_filter")
			Param("plan", "plan_filter")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("tenants_data", func() {
		Description("Get tenant list data as JSON")
		Security("jwt", func() {
			Scope("admin:read")
			Scope("tenant:list")
		})

		Payload(func() {
			types.UIHeaders()
			types.Pagination()
			types.SearchFilter()
			Attribute("status_filter", String, "Filter by tenant status", func() {
				Enum("active", "inactive", "suspended", "pending", "archived", "all")
				Default("all")
			})
			Attribute("plan_filter", String, "Filter by plan type", func() {
				Enum("startup", "small", "medium", "large", "enterprise", "all")
				Default("all")
			})
		})

		Result("ConsoleTenantListResponse")

		HTTP(func() {
			GET("/console/api/tenants")
			Param("page", "page")
			Param("page_size", "page_size")
			Param("q", "q")
			Param("status", "status_filter")
			Param("plan", "plan_filter")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("tenant_create_form", func() {
		Description("Render tenant creation form")
		Security("jwt", func() {
			Scope("admin:write")
			Scope("tenant:create")
		})

		Payload(func() {
			types.UIHeaders()
		})

		Result("ConsolePageData")

		HTTP(func() {
			GET("/console/tenants/new")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("tenant_create", func() {
		Description("Create new tenant")
		Security("jwt", func() {
			Scope("admin:write")
			Scope("tenant:create")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("tenant_data", "ConsoleTenantCreateRequest", "Tenant creation data")
			Required("tenant_data")
		})

		Result("ConsoleTenantDetail")

		HTTP(func() {
			POST("/console/tenants")
			Body("tenant_data")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("tenant_detail", func() {
		Description("Render tenant detail page")
		Security("jwt", func() {
			Scope("admin:read")
			Scope("tenant:read")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("tenant_id", String, "Tenant ID", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Required("tenant_id")
		})

		Result("ConsolePageData")

		HTTP(func() {
			GET("/console/tenants/{tenant_id}")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("tenant_edit_form", func() {
		Description("Render tenant edit form")
		Security("jwt", func() {
			Scope("admin:write")
			Scope("tenant:update")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("tenant_id", String, "Tenant ID", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Required("tenant_id")
		})

		Result("ConsolePageData")

		HTTP(func() {
			GET("/console/tenants/{tenant_id}/edit")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("tenant_update", func() {
		Description("Update tenant")
		Security("jwt", func() {
			Scope("admin:write")
			Scope("tenant:update")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("tenant_id", String, "Tenant ID", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("tenant_data", "ConsoleTenantUpdateRequest", "Tenant update data")
			Required("tenant_id", "tenant_data")
		})

		Result("ConsoleTenantDetail")

		HTTP(func() {
			PUT("/console/tenants/{tenant_id}")
			Body("tenant_data")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("tenant_delete", func() {
		Description("Delete tenant (soft delete)")
		Security("jwt", func() {
			Scope("admin:write")
			Scope("tenant:delete")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("tenant_id", String, "Tenant ID", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("reason", String, "Deletion reason", func() {
				MinLength(5)
				MaxLength(500)
				Example("Account closure requested by customer")
			})
			Required("tenant_id", "reason")
		})

		Result(String, func() {
			Example("Tenant deleted successfully")
		})

		HTTP(func() {
			DELETE("/console/tenants/{tenant_id}")
			Body(func() {
				Attribute("reason")
				Required("reason")
			})
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("tenant_bulk_action", func() {
		Description("Perform bulk actions on tenants")
		Security("jwt", func() {
			Scope("admin:write")
			Scope("tenant:bulk")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("bulk_data", "ConsoleBulkActionRequest", "Bulk action data")
			Required("bulk_data")
		})

		Result(types.BatchResult)

		HTTP(func() {
			POST("/console/tenants/bulk")
			Body("bulk_data")
			Response(StatusAccepted)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// ========================================================================
	// SYSTEM ADMINISTRATION ENDPOINTS
	// ========================================================================

	Method("system_health", func() {
		Description("Get system health status")
		Security("jwt", func() {
			Scope("admin:read")
			Scope("system:health")
		})

		Payload(func() {
			types.UIHeaders()
		})

		Result("ConsoleSystemHealth")

		HTTP(func() {
			GET("/console/api/system/health")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("activity_logs", func() {
		Description("Get system activity logs")
		Security("jwt", func() {
			Scope("admin:read")
			Scope("system:audit")
		})

		Payload(func() {
			types.UIHeaders()
			types.Pagination()
			Attribute("log_type", String, "Filter by log type", func() {
				Enum("user_action", "system_event", "security_event", "data_change", "all")
				Default("all")
			})
			Attribute("severity", String, "Filter by severity", func() {
				Enum("info", "warning", "error", "critical", "all")
				Default("all")
			})
			Attribute("time_range", types.TimeRange, "Time range filter")
		})

		Result(func() {
			Attribute("logs", ArrayOf("ConsoleActivityLog"), "Activity logs")
			Attribute("pagination", types.PaginationMeta, "Pagination metadata")
			Required("logs", "pagination")
		})

		HTTP(func() {
			GET("/console/api/system/logs")
			Param("page", "page")
			Param("page_size", "page_size")
			Param("type", "log_type")
			Param("severity", "severity")
			Param("start_date", "time_range.start_date")
			Param("end_date", "time_range.end_date")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// ========================================================================
	// TENANT CONTEXT SWITCHING (Admin Impersonation)
	// ========================================================================

	Method("switch_tenant_context", func() {
		Description("Switch to a specific tenant context for admin impersonation")
		Security("jwt", func() {
			Scope("admin:read")
			Scope("admin:impersonate")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("tenant_id", String, "Tenant ID to switch to", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Attribute("return_url", String, "URL to redirect after context switch", func() {
				Default("/console/dashboard")
				Example("/console/tenants")
			})
			Required("tenant_id")
		})

		Result(func() {
			Attribute("success", Boolean, "Whether context switch was successful")
			Attribute("redirect_url", String, "URL to redirect to")
			Attribute("tenant_info", types.UITenantInfo, "Switched tenant information")
			Required("success", "redirect_url")
		})

		HTTP(func() {
			POST("/console/switch-context/{tenant_id}")
			Body(func() {
				Attribute("return_url")
			})
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("clear_tenant_context", func() {
		Description("Clear tenant context and return to global admin view")
		Security("jwt", func() {
			Scope("admin:read")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("return_url", String, "URL to redirect after clearing context", func() {
				Default("/console/dashboard")
				Example("/console/dashboard")
			})
		})

		Result(func() {
			Attribute("success", Boolean, "Whether context was cleared")
			Attribute("redirect_url", String, "URL to redirect to")
			Required("success", "redirect_url")
		})

		HTTP(func() {
			POST("/console/clear-context")
			Body(func() {
				Attribute("return_url")
			})
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	// ========================================================================
	// UTILITY ENDPOINTS
	// ========================================================================

	Method("search_suggestions", func() {
		Description("Get search suggestions for console search")
		Security("jwt", func() {
			Scope("admin:read")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("query", String, "Search query", func() {
				MinLength(1)
				MaxLength(100)
				Example("acme")
			})
			Attribute("category", String, "Search category", func() {
				Enum("tenants", "users", "all")
				Default("all")
				Example("tenants")
			})
			Required("query")
		})

		Result(func() {
			Attribute("suggestions", ArrayOf(Object), "Search suggestions", func() {
				Elem(func() {
					Attribute("type", String, "Suggestion type", func() {
						Example("tenant")
					})
					Attribute("id", String, "Resource ID", func() {
						Example("123e4567-e89b-12d3-a456-426614174000")
					})
					Attribute("title", String, "Display title", func() {
						Example("Acme Corporation")
					})
					Attribute("subtitle", String, "Display subtitle", func() {
						Example("acme.example.com")
					})
					Attribute("url", String, "Link URL", func() {
						Example("/console/tenants/123e4567-e89b-12d3-a456-426614174000")
					})
					Required("type", "id", "title", "url")
				})
			})
			Required("suggestions")
		})

		HTTP(func() {
			GET("/console/api/search/suggestions")
			Param("q", "query")
			Param("category", "category")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	// ========================================================================
	// FILES - TEMPLATE SERVING
	// ========================================================================

	Files("/console/assets/*filepath", "./web/admin/assets/", func() {
		Description("Serve console static assets")
	})
})
