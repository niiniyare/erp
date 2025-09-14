package finance

import (
	. "goa.design/goa/v3/dsl"
)

// Service describes the finance management service
var _ = Service("finance", func() {
	Description("Financial management service for double-entry bookkeeping and accounting")

	// Apply global middleware
	HTTP(func() {
		Path("/api/v1/finance")
	})

	// Legacy Account Management Methods (for backward compatibility)
	Method("createAccount", func() {
		Description("Create a new chart of accounts entry")
		Payload(CreateAccountPayload)
		Result(AccountResult)
		Error("bad_request")
		Error("conflict")
		Error("unauthorized")
		Error("unprocessable_entity")
		HTTP(func() {
			POST("/legacy/accounts")
			Response(StatusCreated)
		})
	})

	Method("getAccount", func() {
		Description("Get account by ID")
		Payload(GetAccountByIdPayload)
		Result(AccountResult)
		Error("not_found")
		Error("unauthorized")
		HTTP(func() {
			GET("/{id}")
			Param("id")
		})
	})

	Method("getAccountByCode", func() {
		Description("Get account by account code")
		Payload(GetAccountByCodePayload)
		Result(AccountResult)
		Error("not_found")
		Error("unauthorized")
		HTTP(func() {
			GET("/accounts/by-code/{account_code}")
			Param("account_code")
		})
	})

	Method("getAccountByName", func() {
		Description("Get account by exact account name")
		Payload(GetAccountByNamePayload)
		Result(AccountResult)
		Error("not_found")
		Error("unauthorized")
		HTTP(func() {
			GET("/accounts/by-name")
			Param("account_name")
		})
	})

	Method("listAccounts", func() {
		Description("List accounts with filtering and pagination")
		Payload(ListAccountsPayload)
		Result(AccountListResult)
		Error("bad_request")
		Error("unauthorized")
		HTTP(func() {
			GET("/accounts")
			Param("root_type")
			Param("account_type")
			Param("is_active")
			Param("parent_id")
			Param("search")
		})
	})

	Method("updateAccount", func() {
		Description("Update an existing account")
		Payload(UpdateAccountPayload)
		Result(AccountResult)
		Error("not_found")
		Error("bad_request")
		Error("unauthorized")
		Error("unprocessable_entity")
		HTTP(func() {
			PUT("/{id}")
			Param("id")
		})
	})

	Method("deleteAccount", func() {
		Description("Soft delete an account")
		Payload(DeleteAccountPayload)
		Error("not_found")
		Error("unauthorized")
		Error("conflict")
		HTTP(func() {
			DELETE("/{id}")
			Param("id")
			Response(StatusNoContent)
		})
	})

	Method("getAccountHierarchy", func() {
		Description("Get account hierarchy tree")
		Payload(GetAccountHierarchyPayload)
		Result(AccountListResult)
		Error("unauthorized")
		HTTP(func() {
			GET("/accounts/hierarchy")
			Param("root_id")
		})
	})

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

	// Transaction Management Methods
	Method("createTransaction", func() {
		Description("Create a new financial transaction")
		Payload(CreateTransactionPayload)
		Result(TransactionResult)
		Error("bad_request")
		Error("unauthorized")
		Error("unprocessable_entity")
		HTTP(func() {
			POST("/transactions")
			Response(StatusCreated)
		})
	})

	Method("getTransaction", func() {
		Description("Get transaction by ID with entries")
		Payload(GetTransactionByIdPayload)
		Result(TransactionWithEntriesResult)
		Error("not_found")
		Error("unauthorized")
		HTTP(func() {
			GET("/transactions/{id}")
			Param("id")
		})
	})

	Method("getTransactionByNumber", func() {
		Description("Get transaction by transaction number")
		Payload(GetTransactionByNumberPayload)
		Result(TransactionWithEntriesResult)
		Error("not_found")
		Error("unauthorized")
		HTTP(func() {
			GET("/transactions/by-number/{transaction_number}")
			Param("transaction_number")
		})
	})

	Method("listTransactions", func() {
		Description("List transactions with filtering and pagination")
		Payload(ListTransactionsPayload)
		Result(TransactionListResult)
		Error("bad_request")
		Error("unauthorized")
		HTTP(func() {
			GET("/transactions")
			Param("status")
			Param("type")
			Param("account_id")
			Param("search")
		})
	})

	Method("postTransaction", func() {
		Description("Post a transaction (make it permanent)")
		Payload(PostTransactionPayload)
		Result(TransactionResult)
		Error("not_found")
		Error("bad_request")
		Error("unauthorized")
		Error("unprocessable_entity")
		HTTP(func() {
			POST("/transactions/{id}/post")
			Param("id")
		})
	})

	Method("reverseTransaction", func() {
		Description("Reverse a posted transaction")
		Payload(ReverseTransactionPayload)
		Result(TransactionResult)
		Error("not_found")
		Error("bad_request")
		Error("unauthorized")
		Error("unprocessable_entity")
		HTTP(func() {
			POST("/transactions/{id}/reverse")
			Param("id")
		})
	})

	Method("approveTransaction", func() {
		Description("Approve a transaction for posting")
		Payload(ApproveTransactionPayload)
		Result(TransactionResult)
		Error("not_found")
		Error("bad_request")
		Error("unauthorized")
		HTTP(func() {
			POST("/transactions/{id}/approve")
			Param("id")
		})
	})

	Method("validateTransaction", func() {
		Description("Validate transaction before posting")
		Payload(ValidateTransactionPayload)
		Result(ValidationResult)
		Error("bad_request")
		Error("unauthorized")
		HTTP(func() {
			POST("/transactions/validate")
		})
	})

	// Transaction Workflow Management Methods
	Method("getTransactionStatus", func() {
		Description("Get detailed transaction status and workflow progress")
		Payload(GetTransactionStatusPayload)
		Result(TransactionStatusResult)
		Error("not_found")
		Error("unauthorized")
		HTTP(func() {
			GET("/transactions/{id}/status")
			Param("id")
		})
	})

	Method("submitApprovalDecision", func() {
		Description("Submit approval decision (approve/reject/request changes)")
		Payload(ApprovalDecisionPayload)
		Result(ApprovalDecisionResult)
		Error("not_found")
		Error("bad_request")
		Error("unauthorized")
		Error("conflict")
		HTTP(func() {
			POST("/transactions/{id}/approvals")
			Param("id")
		})
	})

	Method("requestTransactionChanges", func() {
		Description("Request modifications to a submitted transaction")
		Payload(ChangeRequestPayload)
		Result(ChangeRequestResult)
		Error("not_found")
		Error("bad_request")
		Error("unauthorized")
		HTTP(func() {
			POST("/transactions/{id}/change-requests")
			Param("id")
		})
	})

	Method("getTransactionWorkflow", func() {
		Description("Get complete workflow history and available actions for a transaction")
		Payload(GetTransactionWorkflowPayload)
		Result(TransactionWorkflowResult)
		Error("not_found")
		Error("unauthorized")
		HTTP(func() {
			GET("/transactions/{id}/workflow")
			Param("id")
		})
	})

	// Advanced Financial Reporting Methods
	Method("getTrialBalance", func() {
		Description("Generate trial balance report")
		Payload(GetTrialBalancePayload)
		Result(TrialBalanceResult)
		Error("bad_request")
		Error("unauthorized")
		HTTP(func() {
			GET("/reports/trial-balance")
			Param("as_of_date")
			Param("include_zero_balances")
		})
	})

	// Note: Balance Sheet, P&L, and Cash Flow reports will be added in future iterations
})
