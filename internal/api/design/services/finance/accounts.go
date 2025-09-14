package finance

import (
	// _ "github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// AccountsService describes the unified accounts service with proper GOA DSL
var _ = Service("finance", func() {
	Description("Accounts management service with unified account/group operations")

	HTTP(func() {
		Path("/api/v1/finance")
	})

	// Unified Account and Group Management Methods
	Method("createAccountNode", func() {
		Description("Create a new account or account group using unified endpoint")
		Payload(CreateAccountNodePayload)
		Result(AccountNodeResult)
		Error("bad_request")
		Error("conflict")
		Error("unauthorized")
		Error("unprocessable_entity")
		HTTP(func() {
			POST("/accounts")
			Response(StatusCreated)
		})
	})

	Method("getAccountNode", func() {
		Description("Get account or account group by ID")
		Payload(GetAccountNodeByIdPayload)
		Result(AccountNodeResult)
		Error("not_found")
		Error("unauthorized")
		HTTP(func() {
			GET("/{id}")
			Param("id")
		})
	})

	Method("getAccountNodeByCode", func() {
		Description("Get account or account group by code")
		Payload(GetAccountNodeByCodePayload)
		Result(AccountNodeResult)
		Error("not_found")
		Error("unauthorized")
		HTTP(func() {
			GET("/accounts/by-code/{code}")
			Param("code")
		})
	})

	Method("listAccountNodes", func() {
		Description("List unified accounts and groups with filtering")
		Payload(ListAccountNodesPayload)
		Result(AccountNodeListResult)
		Error("bad_request")
		Error("unauthorized")
		HTTP(func() {
			GET("/accounts")
			Param("node_types")
			Param("parent_id")
			Param("max_level")
			Param("include_children")
			Param("root_type")
			Param("account_type")
			Param("financial_statement_section")
			Param("cash_flow_category")
			Param("is_active")
			Param("search_query")
			Param("include_balances")
			Param("sort_by")
			Param("sort_order")
		})
	})

	Method("updateAccountNode", func() {
		Description("Update an existing account or account group")
		Payload(UpdateAccountNodePayload)
		Result(AccountNodeResult)
		Error("not_found")
		Error("bad_request")
		Error("unauthorized")
		Error("unprocessable_entity")
		HTTP(func() {
			PUT("/{id}")
			Param("id")
		})
	})

	Method("deleteAccountNode", func() {
		Description("Soft delete an account or account group")
		Payload(DeleteAccountNodePayload)
		Error("not_found")
		Error("unauthorized")
		Error("conflict")
		HTTP(func() {
			DELETE("/{id}")
			Param("id")
			Response(StatusNoContent)
		})
	})

	Method("searchAccountNodes", func() {
		Description("Search accounts and groups")
		Payload(SearchAccountNodesPayload)
		Result(SearchAccountNodesResult)
		Error("bad_request")
		Error("unauthorized")
		HTTP(func() {
			GET("/accounts/search")
			Param("query")
			Param("node_types")
			Param("limit")
			Param("include_inactive")
		})
	})

	// Account Balance Methods
	Method("getAccountBalance", func() {
		Description("Get current balance for an account")
		Payload(GetAccountBalancePayload)
		Result(AccountBalanceResult)
		Error("not_found")
		Error("unauthorized")
		HTTP(func() {
			GET("/accounts/{account_id}/balance")
			Param("account_id")
			Param("as_of_date")
		})
	})

	// Hierarchy Analysis
	Method("getHierarchyAnalysis", func() {
		Description("Get hierarchy analysis for a parent node")
		Payload(GetHierarchyAnalysisPayload)
		Result(HierarchyAnalysisResult)
		Error("not_found")
		Error("unauthorized")
		HTTP(func() {
			GET("/analytics/hierarchy/{parent_id}")
			Param("parent_id")
			Param("include_balance_data")
			Param("max_depth")
			Param("as_of_date")
		})
	})
})
