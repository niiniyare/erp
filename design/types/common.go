package types

import (
	. "goa.design/goa/v3/dsl"
)

// Pagination provides common pagination parameters
var Pagination = Type("Pagination", func() {
	Description("Pagination parameters")
	Attribute("page", UInt, "Page number (1-based)", func() {
		Default(1)
		Minimum(1)
		Example(1)
	})
	Attribute("page_size", UInt, "Number of items per page", func() {
		Default(20)
		Minimum(1)
		Maximum(100)
		Example(20)
	})
	Attribute("sort_by", String, "Field to sort by", func() {
		Example("created_at")
	})
	Attribute("sort_order", String, "Sort order", func() {
		Enum("asc", "desc")
		Default("desc")
		Example("desc")
	})
})

// PaginatedResponse provides common paginated response structure
var PaginatedResponse = func(itemType DataType) *ResultTypeExpr {
	return ResultType("application/vnd.erp.paginated", func() {
		Description("Paginated response")
		Attributes(func() {
			Attribute("data", ArrayOf(itemType), "The data items")
			Attribute("pagination", PaginationMeta, "Pagination metadata")
		})
		Required("data", "pagination")
	})
}

// PaginationMeta provides pagination metadata
var PaginationMeta = Type("PaginationMeta", func() {
	Description("Pagination metadata")
	Attribute("current_page", UInt, "Current page number")
	Attribute("page_size", UInt, "Items per page")
	Attribute("total_items", UInt, "Total number of items")
	Attribute("total_pages", UInt, "Total number of pages")
	Attribute("has_next", Boolean, "Whether there is a next page")
	Attribute("has_prev", Boolean, "Whether there is a previous page")
	Required("current_page", "page_size", "total_items", "total_pages", "has_next", "has_prev")
})

// AuditFields provides common audit trail fields
var AuditFields = func() {
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	Attribute("updated_at", String, "Last update timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T15:45:00Z")
	})
	Attribute("created_by", String, "ID of user who created the record", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("updated_by", String, "ID of user who last updated the record", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
}

// CommonHeaders provides standard header attributes
var CommonHeaders = func() {
	Header("X-Tenant-ID", String, "Tenant identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Header("X-Entity-ID", String, "Entity/Company identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("987fcdeb-51d2-43b8-a456-426614174000")
	})
	Header("X-User-ID", String, "Current user identifier", func() {
		Pattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
		Example("456e7890-e89b-12d3-a456-426614174000")
	})
}

// ContactInfo describes contact information
var ContactInfo = Type("ContactInfo", func() {
	Description("Contact information")
	Attribute("name", String, "Contact person name", func() {
		MinLength(1)
		MaxLength(100)
		Example("John Doe")
	})
	Attribute("email", String, "Contact email", func() {
		Format(FormatEmail)
		Example("john.doe@acme.com")
	})
	Attribute("phone", String, "Contact phone number", func() {
		Pattern("^\\+?[1-9]\\d{1,14}$")
		Example("+1-555-123-4567")
	})
	Attribute("title", String, "Job title", func() {
		MaxLength(100)
		Example("Chief Technology Officer")
	})
})
