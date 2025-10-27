package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// PAGINATION & SEARCH TYPES
// ============================================================================

// PaginationParams represents standardized pagination parameters for API requests.
// Provides consistent pagination interface across all ERP services.
var PaginationParams = Type("PaginationParams", func() {
	Description("Standardized pagination parameters with sorting and filtering capabilities")
	
	Field(1, "page", UInt, "Page number (1-based indexing)", func() {
		Minimum(1)
		Maximum(10000)
		Default(1)
		Example(1)
		Description("Current page number starting from 1")
	})
	
	Field(2, "per_page", UInt, "Number of items per page", func() {
		Minimum(1)
		Maximum(100)
		Default(20)
		Example(20)
		Description("Maximum items to return per page")
	})
	
	Field(3, "sort_by", String, "Field name to sort by", func() {
		Pattern("^[a-z][a-z0-9_]{0,49}$")
		Example("created_at")
		Description("Database field name for sorting")
	})
	
	Field(4, "sort_order", String, "Sort direction", func() {
		Enum("asc", "desc")
		Default("desc")
		Example("desc")
		Description("Ascending or descending sort order")
	})
	
	Field(5, "search", String, "Global search query", func() {
		MaxLength(500)
		Example("acme corporation")
		Description("Search term applied across searchable fields")
	})
	
	Field(6, "filters", MapOf(String, Any), "Field-specific filters", func() {
		Example(map[string]any{
			"status":        "ACTIVE",
			"created_after": "2023-01-01T00:00:00Z",
			"amount_min":    100.00,
			"amount_max":    10000.00,
		})
		Description("Key-value pairs for filtering results")
	})
	
	Field(7, "include_inactive", Boolean, "Include inactive/soft-deleted records", func() {
		Default(false)
		Example(false)
		Description("Whether to include logically deleted records")
	})
	
	Field(8, "include_totals", Boolean, "Include aggregation totals", func() {
		Default(false)
		Example(true)
		Description("Whether to calculate and return total counts")
	})
	
	Required("page", "per_page")
})

// PaginationResponse represents standardized pagination metadata returned with results.
var PaginationResponse = Type("PaginationResponse", func() {
	Description("Pagination metadata providing navigation and summary information")
	
	Field(1, "current_page", UInt, "Current page number", func() {
		Minimum(1)
		Example(2)
		Description("The page number that was requested")
	})
	
	Field(2, "per_page", UInt, "Items per page", func() {
		Minimum(1)
		Maximum(100)
		Example(20)
		Description("Number of items returned per page")
	})
	
	Field(3, "total_items", UInt, "Total number of available items", func() {
		Example(1247)
		Description("Total count across all pages")
	})
	
	Field(4, "total_pages", UInt, "Total number of pages", func() {
		Minimum(1)
		Example(63)
		Description("Total pages available based on per_page size")
	})
	
	Field(5, "has_next", Boolean, "Next page availability", func() {
		Example(true)
		Description("Whether there are more pages after current")
	})
	
	Field(6, "has_prev", Boolean, "Previous page availability", func() {
		Example(true)
		Description("Whether there are pages before current")
	})
	
	Field(7, "next_page", UInt, "Next page number", func() {
		Minimum(1)
		Example(3)
		Description("Page number for next page (if available)")
	})
	
	Field(8, "prev_page", UInt, "Previous page number", func() {
		Minimum(1)
		Example(1)
		Description("Page number for previous page (if available)")
	})
	
	Field(9, "first_item", UInt, "First item index on current page", func() {
		Minimum(1)
		Example(21)
		Description("1-based index of first item on current page")
	})
	
	Field(10, "last_item", UInt, "Last item index on current page", func() {
		Minimum(1)
		Example(40)
		Description("1-based index of last item on current page")
	})
	
	Field(11, "items_on_page", UInt, "Actual items returned on current page", func() {
		Example(20)
		Description("May be less than per_page on last page")
	})
	
	Required("current_page", "per_page", "total_items", "total_pages", 
		"has_next", "has_prev", "items_on_page")
})

// SearchFilter represents advanced search and filtering capabilities.
var SearchFilter = Type("SearchFilter", func() {
	Description("Advanced search and filtering parameters for complex queries")
	
	Field(1, "query", String, "Full-text search query", func() {
		MaxLength(1000)
		Example("financial transactions december 2023")
		Description("Search term with support for operators and phrases")
	})
	
	Field(2, "fields", ArrayOf(String), "Specific fields to search", func() {
		Example([]string{"name", "description", "reference_number"})
		Description("Limit search to specific fields")
	})
	
	Field(3, "date_range", TimeRange, "Date range filter", func() {
		Description("Filter results within date range")
	})
	
	Field(4, "amount_range", Type("AmountRange", func() {
		Field(1, "min", String, "Minimum amount", func() {
			Pattern("^\\d+(\\.\\d{1,4})?$")
			Example("100.00")
		})
		Field(2, "max", String, "Maximum amount", func() {
			Pattern("^\\d+(\\.\\d{1,4})?$")
			Example("10000.00")
		})
		Field(3, "currency", String, "Currency for amount range", func() {
			Pattern("^[A-Z]{3}$")
			Default("USD")
			Example("USD")
		})
	}), "Amount range filter", func() {
		Description("Filter by monetary amount range")
	})
	
	Field(5, "status_filters", ArrayOf(String), "Status-based filters", func() {
		Example([]string{"ACTIVE", "PENDING", "APPROVED"})
		Description("Filter by one or more status values")
	})
	
	Field(6, "entity_scope", ArrayOf(String), "Entity scope filter", func() {
		Elem(func() {
			Format(FormatUUID)
		})
		Example([]string{
			"entity-123e4567-e89b-12d3-a456-426614174000",
			"entity-456e7890-e89b-12d3-a456-426614174000",
		})
		Description("Limit results to specific entities")
	})
	
	Field(7, "tags", ArrayOf(String), "Tag-based filters", func() {
		Example([]string{"urgent", "recurring", "automated"})
		Description("Filter by associated tags")
	})
	
	Field(8, "custom_filters", MapOf(String, Any), "Custom field filters", func() {
		Example(map[string]any{
			"department":      "engineering",
			"approval_level":  3,
			"has_attachments": true,
		})
		Description("Flexible filters for custom fields")
	})
	
	Field(9, "facets", ArrayOf(String), "Fields to generate facet counts", func() {
		Example([]string{"status", "entity_id", "transaction_type"})
		Description("Fields for which to return aggregation counts")
	})
	
	Field(10, "highlight", Boolean, "Enable search result highlighting", func() {
		Default(false)
		Example(true)
		Description("Highlight matching terms in results")
	})
})

// SearchResult represents enhanced search results with highlighting and facets.
var SearchResult = Type("SearchResult", func() {
	Description("Search result container with metadata and facet information")
	
	Field(1, "query", String, "Original search query", func() {
		Example("financial transactions december 2023")
		Description("Query that was executed")
	})
	
	Field(2, "total_results", UInt, "Total matching results", func() {
		Example(1247)
		Description("Total items matching search criteria")
	})
	
	Field(3, "execution_time_ms", UInt, "Search execution time", func() {
		Example(45)
		Description("Time taken to execute search in milliseconds")
	})
	
	Field(4, "facets", MapOf(String, MapOf(String, UInt)), "Facet counts by field", func() {
		Example(map[string]any{
			"status": map[string]any{
				"ACTIVE":   850,
				"PENDING":  247,
				"ARCHIVED": 150,
			},
			"transaction_type": map[string]any{
				"PAYMENT":      425,
				"RECEIPT":      380,
				"ADJUSTMENT":   442,
			},
		})
		Description("Aggregated counts for faceted fields")
	})
	
	Field(5, "suggestions", ArrayOf(String), "Search query suggestions", func() {
		Example([]string{
			"financial transactions november 2023",
			"financial reports december 2023",
		})
		Description("Alternative queries based on common patterns")
	})
	
	Field(6, "filters_applied", MapOf(String, Any), "Active filters summary", func() {
		Example(map[string]any{
			"date_range": map[string]any{
				"start": "2023-12-01T00:00:00Z",
				"end":   "2023-12-31T23:59:59Z",
			},
			"status": []string{"ACTIVE", "PENDING"},
		})
		Description("Summary of filters that were applied")
	})
	
	Required("query", "total_results", "execution_time_ms")
})

// SortOption represents available sorting options for collections.
var SortOption = Type("SortOption", func() {
	Description("Available sorting configuration for collection endpoints")
	
	Field(1, "field", String, "Sortable field name", func() {
		Pattern("^[a-z][a-z0-9_]{0,49}$")
		Example("created_at")
		Description("Database field available for sorting")
	})
	
	Field(2, "label", String, "Human-readable field label", func() {
		MaxLength(100)
		Example("Date Created")
		Description("Display name for UI sorting options")
	})
	
	Field(3, "data_type", String, "Field data type", func() {
		Enum("string", "number", "date", "boolean", "currency")
		Example("date")
		Description("Data type for proper sorting behavior")
	})
	
	Field(4, "default_order", String, "Default sort direction", func() {
		Enum("asc", "desc")
		Example("desc")
		Description("Recommended sort order for this field")
	})
	
	Field(5, "is_default", Boolean, "Default sort field", func() {
		Default(false)
		Example(true)
		Description("Whether this is the default sort field")
	})
	
	Required("field", "label", "data_type", "default_order")
})

// FilterOption represents available filter options for collections.
var FilterOption = Type("FilterOption", func() {
	Description("Available filter configuration for collection endpoints")
	
	Field(1, "field", String, "Filterable field name", func() {
		Pattern("^[a-z][a-z0-9_]{0,49}$")
		Example("status")
		Description("Database field available for filtering")
	})
	
	Field(2, "label", String, "Human-readable field label", func() {
		MaxLength(100)
		Example("Status")
		Description("Display name for UI filter options")
	})
	
	Field(3, "filter_type", String, "Type of filter operation", func() {
		Enum("exact", "contains", "range", "date_range", "multi_select", "boolean")
		Example("multi_select")
		Description("Filter operation supported by this field")
	})
	
	Field(4, "options", ArrayOf(String), "Available filter values", func() {
		Example([]string{"ACTIVE", "PENDING", "ARCHIVED", "SUSPENDED"})
		Description("Predefined values for selection filters")
	})
	
	Field(5, "data_type", String, "Field data type", func() {
		Enum("string", "number", "date", "boolean", "currency", "uuid")
		Example("string")
		Description("Data type for proper filter behavior")
	})
	
	Field(6, "is_searchable", Boolean, "Field supports text search", func() {
		Default(false)
		Example(true)
		Description("Whether field is included in full-text search")
	})
	
	Required("field", "label", "filter_type", "data_type")
})

// CollectionMetadata provides metadata about collection endpoints.
var CollectionMetadata = Type("CollectionMetadata", func() {
	Description("Metadata about collection capabilities and configuration")
	
	Field(1, "total_count", UInt, "Total items in collection", func() {
		Example(15247)
		Description("Total count without filters applied")
	})
	
	Field(2, "sort_options", ArrayOf(SortOption), "Available sorting fields", func() {
		Description("Fields that can be used for sorting")
	})
	
	Field(3, "filter_options", ArrayOf(FilterOption), "Available filter fields", func() {
		Description("Fields that can be used for filtering")
	})
	
	Field(4, "searchable_fields", ArrayOf(String), "Full-text searchable fields", func() {
		Example([]string{"name", "description", "reference_number"})
		Description("Fields included in text search")
	})
	
	Field(5, "default_sort", SortOption, "Default sorting configuration", func() {
		Description("Default sort field and order")
	})
	
	Field(6, "max_per_page", UInt, "Maximum items per page", func() {
		Minimum(1)
		Maximum(1000)
		Default(100)
		Example(100)
		Description("Hard limit on page size")
	})
	
	Field(7, "supports_export", Boolean, "Collection supports data export", func() {
		Default(false)
		Example(true)
		Description("Whether collection can be exported to files")
	})
	
	Field(8, "export_formats", ArrayOf(String), "Available export formats", func() {
		Example([]string{"CSV", "XLSX", "PDF", "JSON"})
		Description("Supported formats for data export")
	})
	
	Required("total_count", "sort_options", "filter_options", "default_sort", "max_per_page")
})