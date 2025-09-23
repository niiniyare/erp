package ui

import (
	"github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// PORTAL-SPECIFIC TYPES (Client Interface)
// ============================================================================

// PortalClientInfo represents client account information
var PortalClientInfo = Type("PortalClientInfo", func() {
	Description("Client account information for portal interface")
	Attribute("id", String, "Client ID", func() {
		Format(FormatUUID)
		Example("client-123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("company_name", String, "Client company name", func() {
		Example("ABC Manufacturing")
	})
	Attribute("contact_name", String, "Primary contact name", func() {
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
	Attribute("account_number", String, "Account number", func() {
		Example("ACC-2023-001")
	})
	Attribute("status", String, "Account status", func() {
		Enum("active", "suspended", "pending")
		Example("active")
	})
	Attribute("tier", String, "Client tier/level", func() {
		Enum("basic", "premium", "enterprise")
		Example("premium")
	})
	Attribute("billing_address", types.Address, "Billing address")
	Attribute("shipping_address", types.Address, "Shipping address")
	Attribute("since", String, "Client since date", func() {
		Format(FormatDate)
		Example("2022-03-15")
	})
	Required("id", "company_name", "contact_name", "email", "account_number", "status")
})

// PortalAccountSummary represents client account summary
var PortalAccountSummary = Type("PortalAccountSummary", func() {
	Description("Client account summary for portal dashboard")
	Attribute("balance", types.Money, "Current account balance")
	Attribute("credit_limit", types.Money, "Credit limit")
	Attribute("available_credit", types.Money, "Available credit")
	Attribute("last_payment", MapOf(String, Any), "Last payment information", func() {
		Attribute("amount", types.Money, "Payment amount")
		Attribute("date", String, "Payment date", func() {
			Format(FormatDate)
			Example("2023-11-30")
		})
		Attribute("method", String, "Payment method", func() {
			Example("Bank Transfer")
		})
	})
	Attribute("next_invoice", MapOf(String, Any), "Next invoice information", func() {
		Attribute("estimated_amount", types.Money, "Estimated amount")
		Attribute("due_date", String, "Estimated due date", func() {
			Format(FormatDate)
			Example("2023-12-31")
		})
	})
	Required("balance", "credit_limit", "available_credit")
})

// PortalInvoice represents a client invoice in portal
var PortalInvoice = Type("PortalInvoice", func() {
	Description("Client invoice information for portal")
	Attribute("id", String, "Invoice ID", func() {
		Example("INV-2023-001")
	})
	Attribute("number", String, "Invoice number", func() {
		Example("INV-2023-001")
	})
	Attribute("amount", types.Money, "Invoice total amount")
	Attribute("amount_due", types.Money, "Amount due")
	Attribute("status", String, "Invoice status", func() {
		Enum("sent", "viewed", "paid", "overdue", "partial")
		Example("sent")
	})
	Attribute("issue_date", String, "Issue date", func() {
		Format(FormatDate)
		Example("2023-12-01")
	})
	Attribute("due_date", String, "Due date", func() {
		Format(FormatDate)
		Example("2023-12-31")
	})
	Attribute("payment_terms", String, "Payment terms", func() {
		Example("Net 30")
	})
	Attribute("description", String, "Invoice description", func() {
		Example("Monthly service charges for November 2023")
	})
	Attribute("line_items", ArrayOf(MapOf(String, Any)), "Invoice line items", func() {
		Elem(func() {
			Attribute("description", String, "Item description", func() {
				Example("Consulting Services")
			})
			Attribute("quantity", Float64, "Quantity", func() {
				Example(40.0)
			})
			Attribute("unit", String, "Unit of measurement", func() {
				Example("hours")
			})
			Attribute("rate", types.Money, "Unit rate")
			Attribute("amount", types.Money, "Line total")
			Required("description", "quantity", "rate", "amount")
		})
	})
	Attribute("download_url", String, "PDF download URL", func() {
		Format(FormatURI)
		Example("https://portal.acme.com/invoices/INV-2023-001/download")
	})
	Required("id", "number", "amount", "amount_due", "status", "issue_date", "due_date")
})

// PortalOrder represents a client order
var PortalOrder = Type("PortalOrder", func() {
	Description("Client order information for portal")
	Attribute("id", String, "Order ID", func() {
		Example("ORD-2023-001")
	})
	Attribute("number", String, "Order number", func() {
		Example("ORD-2023-001")
	})
	Attribute("status", String, "Order status", func() {
		Enum("pending", "confirmed", "processing", "shipped", "delivered", "cancelled")
		Example("processing")
	})
	Attribute("order_date", String, "Order date", func() {
		Format(FormatDate)
		Example("2023-11-15")
	})
	Attribute("estimated_delivery", String, "Estimated delivery date", func() {
		Format(FormatDate)
		Example("2023-12-01")
	})
	Attribute("total_amount", types.Money, "Total order amount")
	Attribute("items", ArrayOf(MapOf(String, Any)), "Order items", func() {
		Elem(func() {
			Attribute("sku", String, "Product SKU", func() {
				Example("PROD-001")
			})
			Attribute("name", String, "Product name", func() {
				Example("Professional Service Package")
			})
			Attribute("quantity", UInt, "Quantity ordered", func() {
				Example(2)
			})
			Attribute("unit_price", types.Money, "Unit price")
			Attribute("total_price", types.Money, "Line total")
			Required("sku", "name", "quantity", "unit_price", "total_price")
		})
	})
	Attribute("shipping_address", types.Address, "Shipping address")
	Attribute("tracking_number", String, "Shipment tracking number", func() {
		Example("1Z999AA1234567890")
	})
	Required("id", "number", "status", "order_date", "total_amount")
})

// PortalTicket represents a support ticket
var PortalTicket = Type("PortalTicket", func() {
	Description("Support ticket information for portal")
	Attribute("id", String, "Ticket ID", func() {
		Example("TKT-2023-001")
	})
	Attribute("number", String, "Ticket number", func() {
		Example("TKT-2023-001")
	})
	Attribute("subject", String, "Ticket subject", func() {
		Example("Issue with login access")
	})
	Attribute("description", String, "Ticket description", func() {
		Example("Unable to login to the system since yesterday")
	})
	Attribute("status", String, "Ticket status", func() {
		Enum("open", "in_progress", "waiting", "resolved", "closed")
		Example("in_progress")
	})
	Attribute("priority", String, "Ticket priority", func() {
		Enum("low", "medium", "high", "urgent")
		Example("medium")
	})
	Attribute("category", String, "Ticket category", func() {
		Example("Technical Support")
	})
	Attribute("created_date", String, "Creation date", func() {
		Format(FormatDateTime)
		Example("2023-12-05T09:30:00Z")
	})
	Attribute("last_updated", String, "Last update date", func() {
		Format(FormatDateTime)
		Example("2023-12-06T14:22:00Z")
	})
	Attribute("assigned_to", String, "Assigned support agent", func() {
		Example("Sarah Wilson")
	})
	Attribute("messages", ArrayOf(MapOf(String, Any)), "Ticket messages", func() {
		Elem(func() {
			Attribute("id", String, "Message ID")
			Attribute("sender", String, "Message sender", func() {
				Example("Robert Johnson")
			})
			Attribute("sender_type", String, "Sender type", func() {
				Enum("client", "agent", "system")
				Example("client")
			})
			Attribute("message", String, "Message content")
			Attribute("timestamp", String, "Message timestamp", func() {
				Format(FormatDateTime)
			})
			Attribute("attachments", ArrayOf(types.FileInfo), "Message attachments")
			Required("id", "sender", "sender_type", "message", "timestamp")
		})
	})
	Required("id", "number", "subject", "status", "priority", "created_date")
})

// ============================================================================
// PORTAL REQUEST/RESPONSE TYPES
// ============================================================================

// PortalDashboardResponse represents portal dashboard data
var PortalDashboardResponse = Type("PortalDashboardResponse", func() {
	Description("Portal dashboard response")
	Attribute("client_info", "PortalClientInfo", "Client account information")
	Attribute("account_summary", "PortalAccountSummary", "Account summary")
	Attribute("recent_invoices", ArrayOf("PortalInvoice"), "Recent invoices")
	Attribute("recent_orders", ArrayOf("PortalOrder"), "Recent orders")
	Attribute("open_tickets", ArrayOf("PortalTicket"), "Open support tickets")
	Attribute("announcements", ArrayOf(MapOf(String, Any)), "Company announcements", func() {
		Elem(func() {
			Attribute("id", String, "Announcement ID")
			Attribute("title", String, "Announcement title")
			Attribute("message", String, "Announcement message")
			Attribute("type", String, "Announcement type", func() {
				Enum("info", "warning", "maintenance", "promotion")
			})
			Attribute("published_at", String, "Publication date", func() {
				Format(FormatDateTime)
			})
			Required("id", "title", "message", "type", "published_at")
		})
	})
	Required("client_info", "account_summary")
})

// PortalInvoiceListResponse represents invoice list for portal
var PortalInvoiceListResponse = Type("PortalInvoiceListResponse", func() {
	Description("Portal invoice list response")
	Attribute("invoices", ArrayOf("PortalInvoice"), "List of invoices")
	Attribute("pagination", types.PaginationMeta, "Pagination information")
	Attribute("summary", MapOf(String, Any), "Invoice summary", func() {
		Attribute("total_outstanding", types.Money, "Total outstanding amount")
		Attribute("overdue_count", UInt, "Number of overdue invoices")
		Attribute("overdue_amount", types.Money, "Total overdue amount")
		Required("total_outstanding", "overdue_count", "overdue_amount")
	})
	Required("invoices", "pagination", "summary")
})

// PortalTicketCreateRequest represents support ticket creation
var PortalTicketCreateRequest = Type("PortalTicketCreateRequest", func() {
	Description("Portal support ticket creation request")
	Attribute("subject", String, "Ticket subject", func() {
		MinLength(5)
		MaxLength(200)
		Example("Unable to access my account")
	})
	Attribute("description", String, "Ticket description", func() {
		MinLength(10)
		MaxLength(2000)
		Example("I have been unable to login to my account since yesterday morning. I keep getting an error message.")
	})
	Attribute("category", String, "Ticket category", func() {
		Enum("technical_support", "billing", "general_inquiry", "feature_request", "bug_report")
		Example("technical_support")
	})
	Attribute("priority", String, "Ticket priority", func() {
		Enum("low", "medium", "high", "urgent")
		Default("medium")
		Example("medium")
	})
	Attribute("attachments", ArrayOf(String), "Attachment file IDs", func() {
		Example([]string{"file-123e4567-e89b-12d3-a456-426614174000"})
	})
	Required("subject", "description", "category", "priority")
})

// PortalTicketMessageRequest represents adding a message to ticket
var PortalTicketMessageRequest = Type("PortalTicketMessageRequest", func() {
	Description("Portal ticket message request")
	Attribute("message", String, "Message content", func() {
		MinLength(1)
		MaxLength(2000)
		Example("I tried the suggested solution but it didn't work.")
	})
	Attribute("attachments", ArrayOf(String), "Attachment file IDs", func() {
		Example([]string{"file-123e4567-e89b-12d3-a456-426614174000"})
	})
	Required("message")
})

// ============================================================================
// PORTAL TEMPLATE DATA TYPES
// ============================================================================

// PortalPageData represents data for portal template rendering
var PortalPageData = Type("PortalPageData", func() {
	Description("Portal page template data")
	Attribute("page_metadata", types.UIPageMetadata, "Page metadata")
	Attribute("ui_context", types.UIContext, "Current UI context")
	Attribute("client_info", "PortalClientInfo", "Current client information")
	Attribute("navigation", types.UINavigation, "Navigation structure")
	Attribute("theme", types.UITheme, "UI theme configuration")
	Attribute("notifications", ArrayOf(types.UINotification), "Client notifications")
	Attribute("data", Any, "Page-specific data")
	Required("page_metadata", "ui_context", "client_info", "navigation")
})

// PortalTicketFormData represents support ticket form data
var PortalTicketFormData = Type("PortalTicketFormData", func() {
	Description("Portal support ticket form data")
	Attribute("ticket", "PortalTicket", "Existing ticket data (for replies)")
	Attribute("form_fields", ArrayOf(types.UIFormField), "Form field configurations")
	Attribute("validation_errors", MapOf(String, String), "Field validation errors")
	Attribute("category_options", ArrayOf(MapOf(String, Any)), "Available category options", func() {
		Elem(func() {
			Attribute("value", String, "Category value")
			Attribute("label", String, "Category label")
			Attribute("description", String, "Category description")
			Required("value", "label")
		})
	})
	Attribute("priority_options", ArrayOf(MapOf(String, Any)), "Available priority options", func() {
		Elem(func() {
			Attribute("value", String, "Priority value")
			Attribute("label", String, "Priority label")
			Attribute("description", String, "Priority description")
			Required("value", "label")
		})
	})
	Required("form_fields")
})
