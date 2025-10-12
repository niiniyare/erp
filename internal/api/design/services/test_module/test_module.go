package test_module

import (
	. "goa.design/goa/v3/dsl"
)

// TestModuleService describes the testmodule management service with GOA DSL
var _ = Service("test_module", func() {
	Description("TestModule management service with complete CRUD operations")

	HTTP(func() {
		Path("/api/v1/test-module")
	})

	// Core CRUD Methods
	Method("createTestModule", func() {
		Description("Create a new testmodule")
		Payload(CreateTestModulePayload)
		Result(TestModuleResult)
		Error("bad_request")
		Error("conflict")
		Error("unauthorized")
		Error("unprocessable_entity")
		HTTP(func() {
			POST("/test-modules")
			Response(StatusCreated)
		})
	})

	Method("getTestModule", func() {
		Description("Get testmodule by ID")
		Payload(GetTestModuleByIDPayload)
		Result(TestModuleResult)
		Error("not_found")
		Error("unauthorized")
		HTTP(func() {
			GET("/test-modules/" + "{id}")
			Param("id")
		})
	})

	Method("getTestModuleByCode", func() {
		Description("Get testmodule by code")
		Payload(GetTestModuleByCodePayload)
		Result(TestModuleResult)
		Error("not_found")
		Error("unauthorized")
		HTTP(func() {
			GET("/test-modules/by-code/" + "{code}")
			Param("code")
		})
	})

	Method("listTestModule", func() {
		Description("List testmodules with filtering")
		Payload(ListTestModulePayload)
		Result(TestModuleListResult)
		Error("bad_request")
		Error("unauthorized")
		HTTP(func() {
			GET("/test-modules")
			Param("entity_id")
			Param("test_module_type")
			Param("test_module_category")
			Param("status")
			Param("is_active")
			Param("parent_id")
			Param("search_query")
			Param("show_in_reports")
			Param("limit")
			Param("offset")
			Param("sort_by")
			Param("sort_order")
		})
	})

	Method("updateTestModule", func() {
		Description("Update an existing testmodule")
		Payload(UpdateTestModulePayload)
		Result(TestModuleResult)
		Error("not_found")
		Error("bad_request")
		Error("unauthorized")
		Error("unprocessable_entity")
		HTTP(func() {
			PUT("/test-modules/" + "{id}")
			Param("id")
		})
	})

	Method("deleteTestModule", func() {
		Description("Soft delete a testmodule")
		Payload(DeleteTestModulePayload)
		Error("not_found")
		Error("unauthorized")
		Error("conflict")
		HTTP(func() {
			DELETE("/test-modules/" + "{id}")
			Param("id")
			Response(StatusNoContent)
		})
	})

	// Search and Query Methods
	Method("searchTestModule", func() {
		Description("Search testmodules")
		Payload(SearchTestModulePayload)
		Result(SearchTestModuleResult)
		Error("bad_request")
		Error("unauthorized")
		HTTP(func() {
			GET("/test-modules/search")
			Param("query")
			Param("limit")
			Param("include_inactive")
		})
	})

	Method("getActiveTestModule", func() {
		Description("Get all active testmodules")
		Payload(GetActiveTestModulePayload)
		Result(TestModuleListResult)
		Error("unauthorized")
		HTTP(func() {
			GET("/test-modules/active")
			Param("entity_id")
		})
	})

	// Status Management Methods
	Method("activateTestModule", func() {
		Description("Activate a testmodule")
		Payload(ChangeTestModuleStatusPayload)
		Result(TestModuleResult)
		Error("not_found")
		Error("unauthorized")
		Error("conflict")
		HTTP(func() {
			POST("/test-modules/" + "{id}/activate")
			Param("id")
		})
	})

	Method("deactivateTestModule", func() {
		Description("Deactivate a testmodule")
		Payload(ChangeTestModuleStatusPayload)
		Result(TestModuleResult)
		Error("not_found")
		Error("unauthorized")
		Error("conflict")
		HTTP(func() {
			POST("/test-modules/" + "{id}/deactivate")
			Param("id")
		})
	})

	// Hierarchy Methods
	Method("getTestModuleHierarchy", func() {
		Description("Get testmodule hierarchy")
		Payload(GetTestModuleHierarchyPayload)
		Result(TestModuleHierarchyResult)
		Error("not_found")
		Error("unauthorized")
		HTTP(func() {
			GET("/test-modules/" + "{id}/hierarchy")
			Param("id")
			Param("max_depth")
			Param("include_inactive")
		})
	})

	Method("getTestModuleChildren", func() {
		Description("Get testmodule children")
		Payload(GetTestModuleChildrenPayload)
		Result(TestModuleListResult)
		Error("not_found")
		Error("unauthorized")
		HTTP(func() {
			GET("/test-modules/" + "{id}/children")
			Param("id")
			Param("include_inactive")
		})
	})

	// Validation Methods
	Method("validateTestModuleCode", func() {
		Description("Validate testmodule code uniqueness")
		Payload(ValidateTestModuleCodePayload)
		Result(ValidationResult)
		Error("unauthorized")
		HTTP(func() {
			GET("/test-modules/validate-code")
			Param("code")
			Param("exclude_id")
		})
	})

	// Bulk Operations
	Method("createTestModuleBulk", func() {
		Description("Create multiple testmodules in batch")
		Payload(CreateTestModuleBulkPayload)
		Result(TestModuleBulkResult)
		Error("bad_request")
		Error("unauthorized")
		Error("unprocessable_entity")
		HTTP(func() {
			POST("/test-modules/bulk")
			Response(StatusCreated)
		})
	})

	// Export/Import Methods
	Method("exportTestModule", func() {
		Description("Export testmodules data")
		Payload(ExportTestModulePayload)
		Result(ExportResult)
		Error("bad_request")
		Error("unauthorized")
		HTTP(func() {
			GET("/test-modules/export")
			Param("format")
			Param("entity_id")
			Param("status")
			Param("include_inactive")
		})
	})

	// Analytics Methods
	Method("getTestModuleAnalytics", func() {
		Description("Get testmodule analytics and statistics")
		Payload(GetTestModuleAnalyticsPayload)
		Result(TestModuleAnalyticsResult)
		Error("unauthorized")
		HTTP(func() {
			GET("/test-modules/analytics")
			Param("entity_id")
			Param("period")
			Param("group_by")
		})
	})

	// Audit Methods
	Method("getTestModuleAuditTrail", func() {
		Description("Get testmodule audit trail")
		Payload(GetTestModuleAuditTrailPayload)
		Result(AuditTrailResult)
		Error("not_found")
		Error("unauthorized")
		HTTP(func() {
			GET("/test-modules/" + "{id}/audit")
			Param("id")
			Param("limit")
			Param("offset")
		})
	})
})

// TODO: Add Temporal workflow endpoints when workflow integration is ready
// TODO: Consider adding GraphQL endpoint for complex queries
// TODO: Add rate limiting and caching headers configuration