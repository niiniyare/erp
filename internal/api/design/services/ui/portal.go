package ui

import (
	"github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// PortalService defines the client portal UI service
var _ = Service("portal", func() {
	Description("Client portal UI service - provides read-only and limited interactive interface for end customers")

	// Security schemes for portal access
	Security("jwt", func() {
		Scope("client:read")
		Scope("portal:access")
	})

	// Alternative API key authentication for portal widgets/embeds
	Security("api_key", func() {
		Scope("portal:widget")
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
		Description("Render client portal dashboard")
		Security("jwt", func() {
			Scope("client:read")
			Scope("portal:access")
		})

		Payload(func() {
			types.UIHeaders()
			// X-Tenant-ID and client context set by backend middleware after auth
		})

		Result("PortalPageData", func() {
			Description("Portal dashboard page data")
		})

		HTTP(func() {
			GET("/portal/dashboard")
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
			Scope("client:read")
		})

		Payload(func() {
			types.UIHeaders()
		})

		Result("PortalDashboardResponse")

		HTTP(func() {
			GET("/portal/api/dashboard")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// ========================================================================
	// ACCOUNT INFORMATION ENDPOINTS
	// ========================================================================

	Method("account_info", func() {
		Description("Render account information page")
		Security("jwt", func() {
			Scope("client:read")
		})

		Payload(func() {
			types.UIHeaders()
		})

		Result("PortalPageData")

		HTTP(func() {
			GET("/portal/account")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("account_data", func() {
		Description("Get account information as JSON")
		Security("jwt", func() {
			Scope("client:read")
		})

		Payload(func() {
			types.UIHeaders()
		})

		Result(func() {
			Attribute("client_info", "PortalClientInfo", "Client information")
			Attribute("account_summary", "PortalAccountSummary", "Account summary")
			Required("client_info", "account_summary")
		})

		HTTP(func() {
			GET("/portal/api/account")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// ========================================================================
	// INVOICE ENDPOINTS
	// ========================================================================

	Method("invoices_list", func() {
		Description("Render invoices list page")
		Security("jwt", func() {
			Scope("client:read")
		})

		Payload(func() {
			types.UIHeaders()
			Reference(types.Pagination)
			Attribute("status_filter", String, "Filter by invoice status", func() {
				Enum("sent", "viewed", "paid", "overdue", "partial", "all")
				Default("all")
				Example("sent")
			})
			Attribute("year_filter", UInt, "Filter by year", func() {
				Minimum(2020)
				Maximum(2030)
				Example(2023)
			})
		})

		Result("PortalPageData")

		HTTP(func() {
			GET("/portal/invoices")
			Param("page", "page")
			Param("page_size", "page_size")
			Param("status", "status_filter")
			Param("year", "year_filter")
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
			Scope("client:read")
		})

		Payload(func() {
			types.UIHeaders()
			Reference(types.Pagination)
			Attribute("status_filter", String, "Filter by invoice status", func() {
				Enum("sent", "viewed", "paid", "overdue", "partial", "all")
				Default("all")
			})
			Attribute("year_filter", UInt, "Filter by year")
		})

		Result("PortalInvoiceListResponse")

		HTTP(func() {
			GET("/portal/api/invoices")
			Param("page", "page")
			Param("page_size", "page_size")
			Param("status", "status_filter")
			Param("year", "year_filter")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("invoice_detail", func() {
		Description("Render invoice detail page")
		Security("jwt", func() {
			Scope("client:read")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("invoice_id", String, "Invoice ID", func() {
				Example("INV-2023-001")
			})
			Required("invoice_id")
		})

		Result("PortalPageData")

		HTTP(func() {
			GET("/portal/invoices/{invoice_id}")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("invoice_download", func() {
		Description("Download invoice PDF")
		Security("jwt", func() {
			Scope("client:read")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("invoice_id", String, "Invoice ID", func() {
				Example("INV-2023-001")
			})
			Required("invoice_id")
		})

		Result(func() {
			Description("Invoice PDF file")
		})

		HTTP(func() {
			GET("/portal/invoices/{invoice_id}/download")
			Response(StatusOK, func() {
				ContentType("application/pdf")
				Header("Content-Disposition", String, "PDF download header", func() {
					Example("attachment; filename=\"INV-2023-001.pdf\"")
				})
			})
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// ========================================================================
	// ORDER TRACKING ENDPOINTS
	// ========================================================================

	Method("orders_list", func() {
		Description("Render orders list page")
		Security("jwt", func() {
			Scope("client:read")
		})

		Payload(func() {
			types.UIHeaders()
			Reference(types.Pagination)
			Attribute("status_filter", String, "Filter by order status", func() {
				Enum("pending", "confirmed", "processing", "shipped", "delivered", "cancelled", "all")
				Default("all")
				Example("shipped")
			})
			Attribute("year_filter", UInt, "Filter by year")
		})

		Result("PortalPageData")

		HTTP(func() {
			GET("/portal/orders")
			Param("page", "page")
			Param("page_size", "page_size")
			Param("status", "status_filter")
			Param("year", "year_filter")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("orders_data", func() {
		Description("Get orders data as JSON")
		Security("jwt", func() {
			Scope("client:read")
		})

		Payload(func() {
			types.UIHeaders()
			Reference(types.Pagination)
			Attribute("status_filter", String, "Filter by order status", func() {
				Enum("pending", "confirmed", "processing", "shipped", "delivered", "cancelled", "all")
				Default("all")
			})
			Attribute("year_filter", UInt, "Filter by year")
		})

		Result(func() {
			Attribute("orders", ArrayOf("PortalOrder"), "List of orders")
			Attribute("pagination", types.PaginationMeta, "Pagination information")
			Required("orders", "pagination")
		})

		HTTP(func() {
			GET("/portal/api/orders")
			Param("page", "page")
			Param("page_size", "page_size")
			Param("status", "status_filter")
			Param("year", "year_filter")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("order_detail", func() {
		Description("Render order detail page")
		Security("jwt", func() {
			Scope("client:read")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("order_id", String, "Order ID", func() {
				Example("ORD-2023-001")
			})
			Required("order_id")
		})

		Result("PortalPageData")

		HTTP(func() {
			GET("/portal/orders/{order_id}")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	// ========================================================================
	// SUPPORT TICKET ENDPOINTS
	// ========================================================================

	Method("tickets_list", func() {
		Description("Render support tickets list page")
		Security("jwt", func() {
			Scope("client:read")
		})

		Payload(func() {
			types.UIHeaders()
			Reference(types.Pagination)
			Attribute("status_filter", String, "Filter by ticket status", func() {
				Enum("open", "in_progress", "waiting", "resolved", "closed", "all")
				Default("all")
				Example("open")
			})
		})

		Result("PortalPageData")

		HTTP(func() {
			GET("/portal/support")
			Param("page", "page")
			Param("page_size", "page_size")
			Param("status", "status_filter")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("tickets_data", func() {
		Description("Get support tickets data as JSON")
		Security("jwt", func() {
			Scope("client:read")
		})

		Payload(func() {
			types.UIHeaders()
			Reference(types.Pagination)
			Attribute("status_filter", String, "Filter by ticket status", func() {
				Enum("open", "in_progress", "waiting", "resolved", "closed", "all")
				Default("all")
			})
		})

		Result(func() {
			Attribute("tickets", ArrayOf("PortalTicket"), "List of support tickets")
			Attribute("pagination", types.PaginationMeta, "Pagination information")
			Required("tickets", "pagination")
		})

		HTTP(func() {
			GET("/portal/api/support")
			Param("page", "page")
			Param("page_size", "page_size")
			Param("status", "status_filter")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("ticket_create_form", func() {
		Description("Render support ticket creation form")
		Security("jwt", func() {
			Scope("client:read")
		})

		Payload(func() {
			types.UIHeaders()
		})

		Result("PortalPageData")

		HTTP(func() {
			GET("/portal/support/new")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("ticket_create", func() {
		Description("Create new support ticket")
		Security("jwt", func() {
			Scope("client:read") // Clients can create tickets
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("ticket_data", "PortalTicketCreateRequest", "Ticket creation data")
			Required("ticket_data")
		})

		Result("PortalTicket")

		HTTP(func() {
			POST("/portal/support")
			Body("ticket_data")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("ticket_detail", func() {
		Description("Render support ticket detail page")
		Security("jwt", func() {
			Scope("client:read")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("ticket_id", String, "Ticket ID", func() {
				Example("TKT-2023-001")
			})
			Required("ticket_id")
		})

		Result("PortalPageData")

		HTTP(func() {
			GET("/portal/support/{ticket_id}")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("ticket_add_message", func() {
		Description("Add message to support ticket")
		Security("jwt", func() {
			Scope("client:read") // Clients can reply to their tickets
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("ticket_id", String, "Ticket ID", func() {
				Example("TKT-2023-001")
			})
			Attribute("message_data", "PortalTicketMessageRequest", "Message data")
			Required("ticket_id", "message_data")
		})

		Result("PortalTicket")

		HTTP(func() {
			POST("/portal/support/{ticket_id}/messages")
			Body("message_data")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// ========================================================================
	// UTILITY ENDPOINTS
	// ========================================================================

	Method("notifications", func() {
		Description("Get client notifications")
		Security("jwt", func() {
			Scope("client:read")
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("unread_only", Boolean, "Show only unread notifications", func() {
				Default(false)
				Example(true)
			})
		})

		Result(func() {
			Attribute("notifications", ArrayOf(types.UINotification), "Client notifications")
			Attribute("unread_count", UInt, "Number of unread notifications")
			Required("notifications", "unread_count")
		})

		HTTP(func() {
			GET("/portal/api/notifications")
			Param("unread_only", "unread_only")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	Method("profile_update_form", func() {
		Description("Render profile update form")
		Security("jwt", func() {
			Scope("client:read")
		})

		Payload(func() {
			types.UIHeaders()
		})

		Result("PortalPageData")

		HTTP(func() {
			GET("/portal/profile")
			Response(StatusOK, func() {
				ContentType("text/html")
			})
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
		})
	})

	Method("profile_update", func() {
		Description("Update client profile information")
		Security("jwt", func() {
			Scope("client:read") // Clients can update their own profile
		})

		Payload(func() {
			types.UIHeaders()
			Attribute("profile_data", MapOf(String, Any), "Profile update data", func() {
				Attribute("contact_name", String, "Contact person name", func() {
					MinLength(2)
					MaxLength(100)
					Example("Robert Johnson")
				})
				Attribute("email", String, "Contact email", func() {
					Format(FormatEmail)
					Example("robert@abcmfg.com")
				})
				Attribute("phone", String, "Contact phone", func() {
					Pattern(types.PhonePattern)
					Example("+1-555-987-6543")
				})
				Attribute("billing_address", types.Address, "Billing address")
				Attribute("shipping_address", types.Address, "Shipping address")
			})
			Required("profile_data")
		})

		Result("PortalClientInfo")

		HTTP(func() {
			PUT("/portal/profile")
			Body("profile_data")
			Response(StatusOK)
			Response("bad_request", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	// ========================================================================
	// WIDGET ENDPOINTS (For embedding in external sites)
	// ========================================================================

	Method("account_widget", func() {
		Description("Get account summary widget data")
		Security("api_key", func() {
			Scope("portal:widget")
		})

		Payload(func() {
			Attribute("client_id", String, "Client ID", func() {
				Format(FormatUUID)
				Example("client-123e4567-e89b-12d3-a456-426614174000")
			})
			Required("client_id")
		})

		Result(func() {
			Attribute("balance", types.Money, "Current balance")
			Attribute("next_due_amount", types.Money, "Next amount due")
			Attribute("next_due_date", String, "Next due date", func() {
				Format(FormatDate)
			})
			Attribute("status", String, "Account status")
			Required("balance", "status")
		})

		HTTP(func() {
			GET("/portal/widget/account/{client_id}")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("not_found", StatusNotFound)
		})
	})

	// ========================================================================
	// FILES - TEMPLATE SERVING
	// ========================================================================

	Files("/portal/assets/*filepath", "./web/portal/assets/", func() {
		Description("Serve portal static assets")
	})
})
