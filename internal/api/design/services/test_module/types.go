package test_module

import (
	. "goa.design/goa/v3/dsl"
)

// TestModuleService defines the test_module API service
var _ = Service("test_module", func() {
	Description("TestModule management module")

	// Security
	Security(JWTAuth)

	// Error definitions
	Error("unauthorized", String, "Unauthorized access")
	Error("forbidden", String, "Forbidden access")
	Error("not_found", String, "TestModule not found")
	Error("validation_error", String, "Validation failed")
	Error("internal_error", String, "Internal server error")

	// CRUD Operations
	Method("createTestModule", func() {
		Description("Create a new test_module")
		Payload(CreateTestModulePayload)
		Result(TestModuleResult)
		HTTP(func() {
			POST("/api/v1/test_modules")
			Response(StatusCreated)
			Response("validation_error", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("getTestModule", func() {
		Description("Get test_module by ID")
		Payload(GetTestModuleByIDPayload)
		Result(TestModuleResult)
		HTTP(func() {
			GET("/api/v1/test_modules/{id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("listTestModule", func() {
		Description("List test_modules with filtering and pagination")
		Payload(ListTestModulePayload)
		Result(TestModuleListResult)
		HTTP(func() {
			GET("/api/v1/test_modules")
			Response(StatusOK)
			Response("validation_error", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("updateTestModule", func() {
		Description("Update an existing test_module")
		Payload(UpdateTestModulePayload)
		Result(TestModuleResult)
		HTTP(func() {
			PUT("/api/v1/test_modules/{id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("validation_error", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("deleteTestModule", func() {
		Description("Delete a test_module (soft delete)")
		Payload(DeleteTestModulePayload)
		HTTP(func() {
			DELETE("/api/v1/test_modules/{id}")
			Response(StatusNoContent)
			Response("not_found", StatusNotFound)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("searchTestModule", func() {
		Description("Search test_modules by query")
		Payload(SearchTestModulePayload)
		Result(SearchTestModuleResult)
		HTTP(func() {
			GET("/api/v1/test_modules/search")
			Response(StatusOK)
			Response("validation_error", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})
})

// =============================================================================
// PAYLOAD DEFINITIONS
// =============================================================================

// CreateTestModulePayload defines the payload for creating a testmodule
var CreateTestModulePayload = Type("CreateTestModulePayload", func() {
	Description("Payload for creating a new test_module")
	
	Attribute("name", String, "TestModule name", func() {
		MinLength(1)
		MaxLength(255)
		Example("Sample TestModule")
	})
	Attribute("description", String, "TestModule description", func() {
		MaxLength(1000)
		Example("Sample test_module description")
	})
	Attribute("is_active", Boolean, "Whether the test_module is active", func() {
		Default(true)
	})

	Required("name")
})

// GetTestModuleByIDPayload defines the payload for getting a testmodule by ID
var GetTestModuleByIDPayload = Type("GetTestModuleByIDPayload", func() {
	Description("Payload for getting test_module by ID")
	
	Attribute("id", String, "TestModule ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})

	Required("id")
})

// ListTestModulePayload defines the payload for listing testmodules
var ListTestModulePayload = Type("ListTestModulePayload", func() {
	Description("Payload for listing test_modules with filtering and pagination")
	
	Attribute("page", Int32, "Page number", func() {
		Minimum(1)
		Default(1)
		Example(1)
	})
	Attribute("limit", Int32, "Number of items per page", func() {
		Minimum(1)
		Maximum(100)
		Default(20)
		Example(20)
	})
	Attribute("is_active", Boolean, "Filter by active status", func() {
		Example(true)
	})
	Attribute("name", String, "Filter by name", func() {
		Example("Sample")
	})
})

// UpdateTestModulePayload defines the payload for updating a testmodule
var UpdateTestModulePayload = Type("UpdateTestModulePayload", func() {
	Description("Payload for updating an existing test_module")
	
	Attribute("id", String, "TestModule ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("name", String, "TestModule name", func() {
		MinLength(1)
		MaxLength(255)
		Example("Updated TestModule")
	})
	Attribute("description", String, "TestModule description", func() {
		MaxLength(1000)
		Example("Updated test_module description")
	})
	Attribute("is_active", Boolean, "Whether the test_module is active", func() {
		Example(true)
	})

	Required("id")
})

// DeleteTestModulePayload defines the payload for deleting a testmodule
var DeleteTestModulePayload = Type("DeleteTestModulePayload", func() {
	Description("Payload for deleting a test_module")
	
	Attribute("id", String, "TestModule ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})

	Required("id")
})

// SearchTestModulePayload defines the payload for searching testmodules
var SearchTestModulePayload = Type("SearchTestModulePayload", func() {
	Description("Payload for searching test_modules")
	
	Attribute("query", String, "Search query", func() {
		MinLength(1)
		MaxLength(255)
		Example("sample")
	})
	Attribute("limit", Int32, "Maximum number of results", func() {
		Minimum(1)
		Maximum(50)
		Default(10)
		Example(10)
	})

	Required("query")
})

// =============================================================================
// RESULT DEFINITIONS
// =============================================================================

// TestModuleResult defines the result structure for testmodule operations
var TestModuleResult = Type("TestModuleResult", func() {
	Description("TestModule result")
	
	Attribute("id", String, "TestModule ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("name", String, "TestModule name", func() {
		Example("Sample TestModule")
	})
	Attribute("description", String, "TestModule description", func() {
		Example("Sample test_module description")
	})
	Attribute("is_active", Boolean, "Whether the test_module is active", func() {
		Example(true)
	})
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-15T10:30:00Z")
	})
	Attribute("updated_at", String, "Last update timestamp", func() {
		Format(FormatDateTime)
		Example("2024-01-15T15:45:00Z")
	})

	Required("id", "name", "is_active", "created_at")
})

// TestModuleListResult defines the result structure for listing testmodules
var TestModuleListResult = Type("TestModuleListResult", func() {
	Description("TestModule list result with pagination")
	
	Attribute("test_modules", ArrayOf(TestModuleResult), "List of test_modules")
	Attribute("pagination", PaginationMeta, "Pagination metadata")

	Required("test_modules", "pagination")
})

// SearchTestModuleResult defines the result structure for searching testmodules
var SearchTestModuleResult = Type("SearchTestModuleResult", func() {
	Description("TestModule search result")
	
	Attribute("results", ArrayOf(SearchResultItem), "Search results")
	Attribute("total_results", Int32, "Total number of results", func() {
		Example(1)
	})
	Attribute("search_duration_ms", Int32, "Search duration in milliseconds", func() {
		Example(50)
	})

	Required("results", "total_results", "search_duration_ms")
})

// =============================================================================
// SHARED TYPES
// =============================================================================

// PaginationMeta defines pagination metadata
var PaginationMeta = Type("PaginationMeta", func() {
	Description("Pagination metadata")
	
	Attribute("current_page", Int32, "Current page number", func() {
		Example(1)
	})
	Attribute("page_size", Int32, "Number of items per page", func() {
		Example(20)
	})
	Attribute("total_items", Int32, "Total number of items", func() {
		Example(100)
	})
	Attribute("total_pages", Int32, "Total number of pages", func() {
		Example(5)
	})
	Attribute("has_next", Boolean, "Whether there is a next page", func() {
		Example(true)
	})
	Attribute("has_prev", Boolean, "Whether there is a previous page", func() {
		Example(false)
	})

	Required("current_page", "page_size", "total_items", "total_pages", "has_next", "has_prev")
})

// SearchResultItem defines a search result item
var SearchResultItem = Type("SearchResultItem", func() {
	Description("Search result item")
	
	Attribute("id", String, "Item ID", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
	})
	Attribute("name", String, "Item name", func() {
		Example("Sample TestModule")
	})
	Attribute("match_type", String, "Type of match", func() {
		Enum("NAME", "CODE", "DESCRIPTION")
		Example("NAME")
	})
	Attribute("relevance_score", Float64, "Relevance score (0-1)", func() {
		Minimum(0.0)
		Maximum(1.0)
		Example(0.95)
	})

	Required("id", "name", "match_type", "relevance_score")
})

// TODO: Add validation rules for all payload types
// TODO: Add proper error response types
// TODO: Add support for additional filtering options
// TODO: Add bulk operation endpoints