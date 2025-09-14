package finance

import (
	. "github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// =============================================================================
// TRANSACTION MANAGEMENT TYPES
// =============================================================================

var CreateTransactionPayload = Type("CreateTransactionPayload", func() {
	Description("Payload for creating a new transaction with workflow support")

	Attribute("entity_id", String, "Entity ID (optional)", func() {
		Format(FormatUUID)
	})
	Attribute("transaction_number", String, "Transaction number (auto-generated if not provided)")
	Attribute("transaction_type", String, "Transaction type", func() {
		Enum("MANUAL", "SALES_INVOICE", "PURCHASE_INVOICE", "EXPENSE_PAYMENT", "PAYMENT", "RECEIPT", "JOURNAL_ENTRY", "BANK_TRANSFER", "ADJUSTMENT", "OPENING_BALANCE", "CLOSING_ENTRY")
		Example("EXPENSE_PAYMENT")
	})
	Attribute("transaction_date", String, "Transaction date", func() {
		Format(FormatDate)
		Example("2025-09-13")
	})
	Attribute("description", String, "Transaction description", func() {
		MinLength(1)
		MaxLength(1000)
		Example("Office supplies purchase")
	})
	Attribute("reference_number", String, "External reference number", func() {
		Example("PO-2025-089")
	})
	Attribute("currency", String, "Transaction currency", func() {
		Pattern(CurrencyCodePattern)
		Default("USD")
	})
	Attribute("cost_center", String, "Cost center code", func() {
		Example("CC001")
	})
	Attribute("department", String, "Department", func() {
		Example("administration")
	})
	Attribute("entries", ArrayOf(TransactionEntryPayload), "Transaction entries", func() {
		MinLength(2) // At least 2 entries for double-entry
	})
	Attribute("attachments", ArrayOf(String), "Document attachment UUIDs", func() {
		Elem(func() {
			Format(FormatUUID)
		})
		Example([]string{"receipt-uuid", "approval-form-uuid"})
	})
	Attribute("auto_approve", Boolean, "Skip approval if user has sufficient privileges", func() {
		Default(false)
	})
	Attribute("priority", String, "Processing priority", func() {
		Enum("low", "normal", "high", "urgent")
		Default("normal")
	})

	Required("transaction_type", "transaction_date", "description", "entries")
})

var TransactionEntryPayload = Type("TransactionEntryPayload", func() {
	Description("Transaction entry for double-entry bookkeeping")

	// Support both account ID and account code for flexibility
	Attribute("account_id", String, "Account ID (alternative to account_code)", func() {
		Format(FormatUUID)
	})
	Attribute("account_code", String, "Account code (alternative to account_id)", func() {
		Example("1100")
	})
	Attribute("debit_amount", String, "Debit amount (decimal)", func() {
		Pattern(DecimalPattern)
		Example("1500.00")
	})
	Attribute("credit_amount", String, "Credit amount (decimal)", func() {
		Pattern(DecimalPattern)
		Example("0.00")
	})
	Attribute("description", String, "Entry description", func() {
		MinLength(1)
		MaxLength(500)
		Example("Cash payment for supplies")
	})
	Attribute("reference", String, "Entry reference (optional)")
	Attribute("cost_center", String, "Cost center (optional)")
	Attribute("department", String, "Department (optional)")
	Attribute("project_id", String, "Project ID (optional)", func() {
		Format(FormatUUID)
	})
	Attribute("tax_code", String, "Tax code (optional)")
	Attribute("tax_rate", String, "Tax rate (decimal, optional)")

	Required("description")
})

var TransactionResult = Type("TransactionResult", func() {
	Description("Transaction information with workflow status")

	Attribute("id", String, "Transaction ID", func() {
		Format(FormatUUID)
		Example("txn-uuid")
	})
	Attribute("tenant_id", String, "Tenant ID", func() {
		Format(FormatUUID)
	})
	Attribute("entity_id", String, "Entity ID", func() {
		Format(FormatUUID)
	})
	Attribute("transaction_number", String, "Transaction number", func() {
		Example("TXN-2025-001")
	})
	Attribute("transaction_type", String, "Transaction type")
	Attribute("status", String, "Current transaction status", func() {
		Enum("submitted", "awaiting_approval", "processing", "posted", "rejected", "cancelled")
		Example("submitted")
	})
	Attribute("current_stage", String, "Current workflow stage", func() {
		Enum("validation", "manager_review", "cfo_approval", "posting", "balance_update", "completed")
		Example("validation")
	})
	Attribute("transaction_date", String, "Transaction date", func() {
		Format(FormatDate)
	})
	Attribute("posting_date", String, "Posting date", func() {
		Format(FormatDate)
	})
	Attribute("description", String, "Transaction description")
	Attribute("reference_number", String, "Reference number")
	Attribute("amount", String, "Total transaction amount", func() {
		Example("1500.00")
	})
	Attribute("currency", String, "Currency code", func() {
		Example("USD")
	})
	Attribute("exchange_rate", String, "Exchange rate")
	Attribute("total_debit_amount", String, "Total debit amount")
	Attribute("total_credit_amount", String, "Total credit amount")
	Attribute("cost_center", String, "Cost center")
	Attribute("department", String, "Department")
	Attribute("estimated_completion", String, "Estimated completion time", func() {
		Format(FormatDateTime)
		Example("2025-09-13T16:00:00Z")
	})
	Attribute("progress_percentage", Int32, "Workflow progress (0-100)", func() {
		Example(20)
	})
	Attribute("approval_status", String, "Approval status")
	Attribute("approval_required", Boolean, "Whether approval is required")

	// Use common audit fields
	AuditFields()

	Required("id", "transaction_number", "transaction_type", "status", "current_stage", "transaction_date", "description")
})

var TransactionWithEntriesResult = Type("TransactionWithEntriesResult", func() {
	Description("Transaction with its entries")

	Attribute("transaction", TransactionResult, "Transaction details")
	Attribute("entries", ArrayOf(TransactionEntryResult), "Transaction entries")
	Attribute("is_balanced", Boolean, "Whether transaction is balanced")
	Attribute("validation_errors", ArrayOf(ValidationError), "Validation errors if any")

	Required("transaction", "entries", "is_balanced")
})

var TransactionEntryResult = Type("TransactionEntryResult", func() {
	Description("Transaction entry information")

	Attribute("id", String, "Entry ID", func() {
		Format(FormatUUID)
	})
	Attribute("entry_number", Int32, "Entry sequence number")
	Attribute("account_id", String, "Account ID", func() {
		Format(FormatUUID)
	})
	Attribute("account_code", String, "Account code")
	Attribute("account_name", String, "Account name")
	Attribute("debit_amount", String, "Debit amount")
	Attribute("credit_amount", String, "Credit amount")
	Attribute("description", String, "Entry description")
	Attribute("reference", String, "Entry reference")
	Attribute("cost_center", String, "Cost center")
	Attribute("department", String, "Department")
	Attribute("project_id", String, "Project ID", func() {
		Format(FormatUUID)
	})
	Attribute("tax_code", String, "Tax code")
	Attribute("tax_rate", String, "Tax rate")
	Attribute("tax_amount", String, "Tax amount")
	Attribute("reconciled", Boolean, "Whether entry is reconciled")

	Required("id", "entry_number", "account_id", "account_code", "account_name", "description")
})

var PostTransactionPayload = Type("PostTransactionPayload", func() {
	Description("Payload for posting a transaction")

	Attribute("id", String, "Transaction ID", func() {
		Format(FormatUUID)
	})
	Attribute("posting_date", String, "Posting date (optional, defaults to today)", func() {
		Format(FormatDate)
	})
	Attribute("validate_before_posting", Boolean, "Validate before posting", func() {
		Default(true)
	})
	Attribute("force_post", Boolean, "Force post despite warnings", func() {
		Default(false)
	})

	Required("id")
})

var ReverseTransactionPayload = Type("ReverseTransactionPayload", func() {
	Description("Payload for reversing a transaction")

	Attribute("id", String, "Transaction ID", func() {
		Format(FormatUUID)
	})
	Attribute("reason", String, "Reversal reason", func() {
		MinLength(1)
		MaxLength(500)
		Example("Incorrect entry - duplicate payment")
	})
	Attribute("reversal_date", String, "Reversal date (optional, defaults to today)", func() {
		Format(FormatDate)
	})

	Required("id", "reason")
})

var ApproveTransactionPayload = Type("ApproveTransactionPayload", func() {
	Description("Payload for approving a transaction")

	Attribute("id", String, "Transaction ID", func() {
		Format(FormatUUID)
	})
	Attribute("notes", String, "Approval notes (optional)", func() {
		MaxLength(1000)
	})

	Required("id")
})

var ValidateTransactionPayload = Type("ValidateTransactionPayload", func() {
	Description("Payload for transaction validation")

	Attribute("transaction", CreateTransactionPayload, "Transaction to validate")
	Attribute("validation_level", String, "Validation level", func() {
		Enum("BASIC", "STRICT", "COMPLETE")
		Default("STRICT")
	})

	Required("transaction")
})

var ValidationResult = Type("ValidationResult", func() {
	Description("Transaction validation result")

	Attribute("is_valid", Boolean, "Whether transaction is valid")
	Attribute("is_balanced", Boolean, "Whether debits equal credits")
	Attribute("total_debits", String, "Total debit amount")
	Attribute("total_credits", String, "Total credit amount")
	Attribute("balance_difference", String, "Difference between debits and credits")
	Attribute("errors", ArrayOf(ValidationError), "Validation errors")
	Attribute("warnings", ArrayOf(ValidationWarningResult), "Validation warnings")
	Attribute("validation_level", String, "Validation level used")

	Required("is_valid", "is_balanced", "total_debits", "total_credits", "validation_level")
})

var ValidationWarningResult = Type("ValidationWarningResult", func() {
	Description("Validation warning details")

	Attribute("field", String, "Field with warning")
	Attribute("message", String, "Warning message")
	Attribute("code", String, "Warning code")

	Required("field", "message", "code")
})

var ListTransactionsPayload = Type("ListTransactionsPayload", func() {
	Description("Payload for listing transactions")

	Attribute("status", String, "Filter by transaction status", func() {
		Enum("DRAFT", "PENDING_APPROVAL", "APPROVED", "POSTED", "CANCELLED", "REVERSED")
	})
	Attribute("type", String, "Filter by transaction type", func() {
		Enum("MANUAL", "SALES_INVOICE", "PURCHASE_INVOICE", "PAYMENT", "RECEIPT", "JOURNAL_ENTRY", "BANK_TRANSFER", "ADJUSTMENT", "OPENING_BALANCE", "CLOSING_ENTRY")
	})
	Attribute("date_range", TimeRange, "Date range filter")
	Attribute("account_id", String, "Filter by account", func() {
		Format(FormatUUID)
	})
	Attribute("search", String, "Search in transaction number or description")

	// Use common pagination instead of limit/offset
	Attribute("pagination", Pagination, "Pagination parameters")
})

var TransactionListResult = Type("TransactionListResult", func() {
	Description("List of transactions with pagination")

	Attribute("transactions", ArrayOf(TransactionResult), "List of transactions")
	Attribute("pagination", PaginationMeta, "Pagination metadata")

	Required("transactions", "pagination")
})

// =============================================================================
// INLINE PAYLOADS FROM TRANSACTIONS.GO SERVICE
// =============================================================================

var GetTransactionByIdPayload = Type("GetTransactionByIdPayload", func() {
	Description("Payload for getting transaction by ID")
	Attribute("id", String, "Transaction ID", func() {
		Format(FormatUUID)
	})
	Required("id")
})

var GetTransactionByNumberPayload = Type("GetTransactionByNumberPayload", func() {
	Description("Payload for getting transaction by number")
	Attribute("transaction_number", String, "Transaction number", func() {
		Example("TXN-2025-001")
	})
	Required("transaction_number")
})

var GetTransactionStatusPayload = Type("GetTransactionStatusPayload", func() {
	Description("Payload for getting transaction status")
	Attribute("id", String, "Transaction ID", func() {
		Format(FormatUUID)
	})
	Required("id")
})

var GetTransactionWorkflowPayload = Type("GetTransactionWorkflowPayload", func() {
	Description("Payload for getting transaction workflow")
	Attribute("id", String, "Transaction ID", func() {
		Format(FormatUUID)
	})
	Required("id")
})

