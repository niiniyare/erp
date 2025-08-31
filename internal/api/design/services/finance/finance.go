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

	// Account Management Methods
	Method("createAccount", func() {
		Description("Create a new chart of accounts entry")
		Payload(CreateAccountPayload)
		Result(AccountResult)
		Error("bad_request")
		Error("conflict")
		Error("unauthorized")
		Error("unprocessable_entity")
		HTTP(func() {
			POST("/accounts")
			Response(StatusCreated)
		})
	})

	Method("getAccount", func() {
		Description("Get account by ID")
		Payload(func() {
			Attribute("id", String, "Account ID", func() {
				Format(FormatUUID)
			})
			Required("id")
		})
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
		Payload(func() {
			Attribute("account_code", String, "Account code", func() {
				Example("1100")
			})
			Required("account_code")
		})
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
		Payload(func() {
			Attribute("account_name", String, "Account name", func() {
				Example("Cash - Operating Account")
			})
			Required("account_name")
		})
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
			Param("limit")
			Param("offset")
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
		Payload(func() {
			Attribute("id", String, "Account ID", func() {
				Format(FormatUUID)
			})
			Required("id")
		})
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
		Payload(func() {
			Attribute("root_id", String, "Root account ID (optional)", func() {
				Format(FormatUUID)
			})
		})
		Result(AccountHierarchyResult)
		Error("unauthorized")
		HTTP(func() {
			GET("/accounts/hierarchy")
			Param("root_id")
		})
	})

	Method("getAccountBalance", func() {
		Description("Get current balance for an account")
		Payload(func() {
			Attribute("account_id", String, "Account ID", func() {
				Format(FormatUUID)
			})
			Attribute("as_of_date", String, "Balance as of date (optional)", func() {
				Format(FormatDate)
			})
			Required("account_id")
		})
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
		Payload(func() {
			Attribute("id", String, "Transaction ID", func() {
				Format(FormatUUID)
			})
			Required("id")
		})
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
		Payload(func() {
			Attribute("transaction_number", String, "Transaction number", func() {
				Example("TXN-2025-001")
			})
			Required("transaction_number")
		})
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
			Param("date_from")
			Param("date_to")
			Param("account_id")
			Param("search")
			Param("limit")
			Param("offset")
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

	// Financial Reporting Methods
	Method("getTrialBalance", func() {
		Description("Generate trial balance report")
		Payload(TrialBalancePayload)
		Result(TrialBalanceResult)
		Error("bad_request")
		Error("unauthorized")
		HTTP(func() {
			GET("/reports/trial-balance")
			Param("as_of_date")
			Param("include_zero_balances")
		})
	})
})
