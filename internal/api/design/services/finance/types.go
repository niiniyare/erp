package finance

import (
	. "goa.design/goa/v3/dsl"
)

// CreateAccountPayload is Account-related types
var CreateAccountPayload = Type("CreateAccountPayload", func() {
	Description("Payload for creating a new account")

	Attribute("entity_id", String, "Entity ID (optional)", func() {
		Format(FormatUUID)
	})
	Attribute("account_code", String, "Unique account code", func() {
		MinLength(1)
		MaxLength(20)
		Example("1100")
	})
	Attribute("account_name", String, "Account name", func() {
		MinLength(1)
		MaxLength(255)
		Example("Cash - Operating Account")
	})
	Attribute("account_description", String, "Account description (optional)", func() {
		MaxLength(1000)
	})

	// Hierarchy and grouping
	Attribute("parent_account_id", String, "Parent account ID for hierarchy (optional)", func() {
		Format(FormatUUID)
	})
	Attribute("account_group_id", String, "Account group ID for organization (optional)", func() {
		Format(FormatUUID)
	})
	Attribute("account_header_id", String, "Account header ID for grouping (optional)", func() {
		Format(FormatUUID)
	})

	// Classification
	Attribute("root_type", String, "Root account type", func() {
		Enum("ASSET", "LIABILITY", "EQUITY", "REVENUE", "EXPENSE")
		Example("ASSET")
	})
	Attribute("account_type", String, "Detailed account type", func() {
		Example("BANK")
	})
	Attribute("account_subtype", String, "Account subtype (optional)")
	Attribute("account_category", String, "Account category for grouping (optional)", func() {
		MaxLength(100)
		Example("Current Assets")
	})
	Attribute("sub_category", String, "Sub-category within main category (optional)", func() {
		MaxLength(100)
		Example("Cash and Equivalents")
	})
	Attribute("normal_balance", String, "Normal balance side", func() {
		Enum("DEBIT", "CREDIT")
		Example("DEBIT")
	})

	// Currency settings
	Attribute("currency_code", String, "Primary currency code (optional)", func() {
		Pattern("^[A-Z]{3}$")
		Example("USD")
	})

	// Operational settings
	Attribute("is_active", Boolean, "Whether account is active", func() {
		Default(true)
	})

	// Reporting and display
	Attribute("display_order", Int32, "Display order in UI/reports", func() {
		Minimum(0)
		Default(0)
	})
	Attribute("show_in_reports", Boolean, "Whether to include in standard reports", func() {
		Default(true)
	})
	Attribute("consolidation_account", String, "Consolidation mapping for multi-entity (optional)", func() {
		MaxLength(100)
	})
	Attribute("cash_flow_type", String, "Cash flow statement classification (optional)", func() {
		Enum("OPERATING", "INVESTING", "FINANCING")
		Example("OPERATING")
	})

	Required("account_code", "account_name", "root_type", "normal_balance")
})

var UpdateAccountPayload = Type("UpdateAccountPayload", func() {
	Description("Payload for updating an account")

	Attribute("id", String, "Account ID", func() {
		Format(FormatUUID)
	})
	Attribute("account_name", String, "Account name", func() {
		MinLength(1)
		MaxLength(255)
	})
	Attribute("account_description", String, "Account description")

	// Grouping and hierarchy (limited updates for data integrity)
	Attribute("account_group_id", String, "Account group ID (optional)", func() {
		Format(FormatUUID)
	})
	Attribute("account_header_id", String, "Account header ID (optional)", func() {
		Format(FormatUUID)
	})

	// Classification updates
	Attribute("account_category", String, "Account category for grouping (optional)", func() {
		MaxLength(100)
	})
	Attribute("sub_category", String, "Sub-category within main category (optional)", func() {
		MaxLength(100)
	})

	// Operational settings
	Attribute("is_active", Boolean, "Whether account is active")
	Attribute("allow_manual_entries", Boolean, "Allow manual journal entries")
	Attribute("require_reference", Boolean, "Require reference for entries")

	// Reporting and display
	Attribute("display_order", Int32, "Display order in UI/reports", func() {
		Minimum(0)
	})
	Attribute("show_in_reports", Boolean, "Whether to include in standard reports")
	Attribute("consolidation_account", String, "Consolidation mapping for multi-entity (optional)", func() {
		MaxLength(100)
	})
	Attribute("cash_flow_type", String, "Cash flow statement classification (optional)", func() {
		Enum("OPERATING", "INVESTING", "FINANCING")
	})

	Required("id")
})

var AccountResult = Type("AccountResult", func() {
	Description("Account information")

	Attribute("id", String, "Account ID", func() {
		Format(FormatUUID)
	})
	Attribute("tenant_id", String, "Tenant ID", func() {
		Format(FormatUUID)
	})
	Attribute("entity_id", String, "Entity ID", func() {
		Format(FormatUUID)
	})
	Attribute("account_code", String, "Account code")
	Attribute("account_name", String, "Account name")
	Attribute("account_description", String, "Account description")

	// Hierarchy and grouping
	Attribute("parent_account_id", String, "Parent account ID", func() {
		Format(FormatUUID)
	})
	Attribute("account_level", Int32, "Hierarchy level")
	Attribute("account_path", String, "Hierarchy path")
	Attribute("has_children", Boolean, "Whether account has child accounts")
	Attribute("is_leaf_account", Boolean, "Whether account is a leaf node")
	Attribute("account_group_id", String, "Account group ID", func() {
		Format(FormatUUID)
	})
	Attribute("account_header_id", String, "Account header ID", func() {
		Format(FormatUUID)
	})

	// Classification
	Attribute("root_type", String, "Root account type")
	Attribute("account_type", String, "Account type")
	Attribute("account_subtype", String, "Account subtype")
	Attribute("account_category", String, "Account category for grouping")
	Attribute("sub_category", String, "Sub-category within main category")
	Attribute("normal_balance", String, "Normal balance side")

	// Operational settings
	Attribute("is_active", Boolean, "Whether account is active")
	Attribute("is_system_account", Boolean, "Whether this is a system account")
	Attribute("allow_manual_entries", Boolean, "Whether manual entries are allowed")
	Attribute("require_reference", Boolean, "Whether reference is required")

	// Balance tracking
	Attribute("current_balance", String, "Current account balance")
	Attribute("ytd_balance", String, "Year-to-date balance")
	Attribute("last_transaction_date", String, "Last transaction date", func() {
		Format(FormatDateTime)
	})

	// Reporting and display
	Attribute("financial_statement_line", String, "Financial statement line grouping")
	Attribute("report_order", Int32, "Sort order in reports")
	Attribute("display_order", Int32, "Display order in UI/reports")
	Attribute("show_in_reports", Boolean, "Whether to include in standard reports")
	Attribute("consolidation_account", String, "Consolidation mapping for multi-entity")
	Attribute("cash_flow_type", String, "Cash flow statement classification")

	// Currency and localization
	Attribute("currency_code", String, "Primary currency code")
	Attribute("is_multi_currency", Boolean, "Whether account accepts multiple currencies")

	// Budgeting
	Attribute("is_budgetable", Boolean, "Whether account can have budgets")
	Attribute("budget_variance_threshold", String, "Budget variance alert threshold")

	// Audit fields
	Attribute("version", Int32, "Version for optimistic locking")
	Attribute("validation_status", String, "Current validation status")
	Attribute("last_validation_run", String, "Last validation timestamp", func() {
		Format(FormatDateTime)
	})
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
	})
	Attribute("updated_at", String, "Last update timestamp", func() {
		Format(FormatDateTime)
	})

	Required("id", "account_code", "account_name", "root_type", "account_type", "normal_balance", "is_active")
})

var ListAccountsPayload = Type("ListAccountsPayload", func() {
	Description("Payload for listing accounts with filters")

	Attribute("root_type", String, "Filter by root type", func() {
		Enum("ASSET", "LIABILITY", "EQUITY", "REVENUE", "EXPENSE")
	})
	Attribute("account_type", String, "Filter by account type")
	Attribute("is_active", Boolean, "Filter by active status")
	Attribute("parent_id", String, "Filter by parent account", func() {
		Format(FormatUUID)
	})
	Attribute("search", String, "Search in account code or name")
	Attribute("limit", Int32, "Maximum number of results", func() {
		Default(50)
		Minimum(1)
		Maximum(1000)
	})
	Attribute("offset", Int32, "Number of results to skip", func() {
		Default(0)
		Minimum(0)
	})
})

var AccountListResult = Type("AccountListResult", func() {
	Description("List of accounts with pagination info")

	Attribute("accounts", ArrayOf(AccountResult), "List of accounts")
	Attribute("total_count", Int64, "Total number of accounts")
	Attribute("limit", Int32, "Results limit used")
	Attribute("offset", Int32, "Results offset used")

	Required("accounts", "total_count", "limit", "offset")
})

var AccountHierarchyResult = Type("AccountHierarchyResult", func() {
	Description("Hierarchical account structure")

	Attribute("accounts", ArrayOf(AccountResult), "Accounts in hierarchical order")
	Attribute("total_count", Int32, "Total number of accounts in hierarchy")

	Required("accounts", "total_count")
})

// CreateTransactionPayload is Transaction-related types
var CreateTransactionPayload = Type("CreateTransactionPayload", func() {
	Description("Payload for creating a new transaction")

	Attribute("entity_id", String, "Entity ID (optional)", func() {
		Format(FormatUUID)
	})
	Attribute("transaction_number", String, "Transaction number (auto-generated if not provided)")
	Attribute("transaction_type", String, "Transaction type", func() {
		Enum("MANUAL", "SALES_INVOICE", "PURCHASE_INVOICE", "PAYMENT", "RECEIPT", "JOURNAL_ENTRY", "BANK_TRANSFER", "ADJUSTMENT", "OPENING_BALANCE", "CLOSING_ENTRY")
		Example("JOURNAL_ENTRY")
	})
	Attribute("transaction_date", String, "Transaction date", func() {
		Format(FormatDate)
		Example("2025-08-31")
	})
	Attribute("description", String, "Transaction description", func() {
		MinLength(1)
		MaxLength(1000)
		Example("Monthly rent payment")
	})
	Attribute("reference_number", String, "External reference number")
	Attribute("currency_code", String, "Transaction currency", func() {
		Pattern("^[A-Z]{3}$")
		Default("USD")
	})
	Attribute("entries", ArrayOf(TransactionEntryPayload), "Transaction entries", func() {
		MinLength(2) // At least 2 entries for double-entry
	})

	Required("transaction_type", "transaction_date", "description", "entries")
})

var TransactionEntryPayload = Type("TransactionEntryPayload", func() {
	Description("Transaction entry for double-entry bookkeeping")

	Attribute("account_id", String, "Account ID", func() {
		Format(FormatUUID)
	})
	Attribute("debit_amount", String, "Debit amount (decimal)", func() {
		Example("1500.00")
	})
	Attribute("credit_amount", String, "Credit amount (decimal)", func() {
		Example("1500.00")
	})
	Attribute("description", String, "Entry description", func() {
		MinLength(1)
		MaxLength(500)
	})
	Attribute("reference", String, "Entry reference (optional)")
	Attribute("cost_center", String, "Cost center (optional)")
	Attribute("department", String, "Department (optional)")
	Attribute("project_id", String, "Project ID (optional)", func() {
		Format(FormatUUID)
	})
	Attribute("tax_code", String, "Tax code (optional)")
	Attribute("tax_rate", String, "Tax rate (decimal, optional)")

	Required("account_id", "description")
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

var TransactionResult = Type("TransactionResult", func() {
	Description("Transaction information")

	Attribute("id", String, "Transaction ID", func() {
		Format(FormatUUID)
	})
	Attribute("tenant_id", String, "Tenant ID", func() {
		Format(FormatUUID)
	})
	Attribute("entity_id", String, "Entity ID", func() {
		Format(FormatUUID)
	})
	Attribute("transaction_number", String, "Transaction number")
	Attribute("transaction_type", String, "Transaction type")
	Attribute("transaction_status", String, "Transaction status")
	Attribute("transaction_date", String, "Transaction date", func() {
		Format(FormatDate)
	})
	Attribute("posting_date", String, "Posting date", func() {
		Format(FormatDate)
	})
	Attribute("description", String, "Transaction description")
	Attribute("reference_number", String, "Reference number")
	Attribute("currency_code", String, "Currency code")
	Attribute("exchange_rate", String, "Exchange rate")
	Attribute("total_debit_amount", String, "Total debit amount")
	Attribute("total_credit_amount", String, "Total credit amount")
	Attribute("approval_status", String, "Approval status")
	Attribute("approval_required", Boolean, "Whether approval is required")
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
	})
	Attribute("updated_at", String, "Last update timestamp", func() {
		Format(FormatDateTime)
	})

	Required("id", "transaction_number", "transaction_type", "transaction_status", "transaction_date", "description")
})

var TransactionWithEntriesResult = Type("TransactionWithEntriesResult", func() {
	Description("Transaction with its entries")

	Attribute("transaction", TransactionResult, "Transaction details")
	Attribute("entries", ArrayOf(TransactionEntryResult), "Transaction entries")
	Attribute("is_balanced", Boolean, "Whether transaction is balanced")
	Attribute("validation_errors", ArrayOf(ValidationErrorResult), "Validation errors if any")

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

var ListTransactionsPayload = Type("ListTransactionsPayload", func() {
	Description("Payload for listing transactions")

	Attribute("status", String, "Filter by transaction status", func() {
		Enum("DRAFT", "PENDING_APPROVAL", "APPROVED", "POSTED", "CANCELLED", "REVERSED")
	})
	Attribute("type", String, "Filter by transaction type", func() {
		Enum("MANUAL", "SALES_INVOICE", "PURCHASE_INVOICE", "PAYMENT", "RECEIPT", "JOURNAL_ENTRY", "BANK_TRANSFER", "ADJUSTMENT", "OPENING_BALANCE", "CLOSING_ENTRY")
	})
	Attribute("date_from", String, "Start date filter", func() {
		Format(FormatDate)
	})
	Attribute("date_to", String, "End date filter", func() {
		Format(FormatDate)
	})
	Attribute("account_id", String, "Filter by account", func() {
		Format(FormatUUID)
	})
	Attribute("search", String, "Search in transaction number or description")
	Attribute("limit", Int32, "Maximum number of results", func() {
		Default(50)
		Minimum(1)
		Maximum(1000)
	})
	Attribute("offset", Int32, "Number of results to skip", func() {
		Default(0)
		Minimum(0)
	})
})

var TransactionListResult = Type("TransactionListResult", func() {
	Description("List of transactions with pagination")

	Attribute("transactions", ArrayOf(TransactionResult), "List of transactions")
	Attribute("total_count", Int64, "Total number of transactions")
	Attribute("limit", Int32, "Results limit used")
	Attribute("offset", Int32, "Results offset used")

	Required("transactions", "total_count", "limit", "offset")
})

// TrialBalancePayload Financial Reporting types
var TrialBalancePayload = Type("TrialBalancePayload", func() {
	Description("Payload for trial balance report")

	Attribute("as_of_date", String, "Balance as of date (optional, defaults to today)", func() {
		Format(FormatDate)
		Example("2025-08-31")
	})
	Attribute("include_zero_balances", Boolean, "Include accounts with zero balances", func() {
		Default(false)
	})
})

var TrialBalanceResult = Type("TrialBalanceResult", func() {
	Description("Trial balance report")

	Attribute("as_of_date", String, "Report date", func() {
		Format(FormatDate)
	})
	Attribute("accounts", ArrayOf(TrialBalanceEntry), "Account balances")
	Attribute("total_debits", String, "Total debit amounts")
	Attribute("total_credits", String, "Total credit amounts")
	Attribute("is_balanced", Boolean, "Whether debits equal credits")
	Attribute("generated_at", String, "Report generation timestamp", func() {
		Format(FormatDateTime)
	})

	Required("as_of_date", "accounts", "total_debits", "total_credits", "is_balanced", "generated_at")
})

var TrialBalanceEntry = Type("TrialBalanceEntry", func() {
	Description("Trial balance account entry")

	Attribute("account_id", String, "Account ID", func() {
		Format(FormatUUID)
	})
	Attribute("account_code", String, "Account code")
	Attribute("account_name", String, "Account name")
	Attribute("root_type", String, "Root account type")
	Attribute("account_type", String, "Account type")
	Attribute("normal_balance", String, "Normal balance side")
	Attribute("total_debits", String, "Total debit amount")
	Attribute("total_credits", String, "Total credit amount")
	Attribute("net_balance", String, "Net balance amount")

	Required("account_id", "account_code", "account_name", "root_type", "total_debits", "total_credits", "net_balance")
})

var AccountBalanceResult = Type("AccountBalanceResult", func() {
	Description("Account balance information")

	Attribute("account_id", String, "Account ID", func() {
		Format(FormatUUID)
	})
	Attribute("account_code", String, "Account code")
	Attribute("account_name", String, "Account name")
	Attribute("current_balance", String, "Current balance")
	Attribute("total_debits", String, "Total debit transactions")
	Attribute("total_credits", String, "Total credit transactions")
	Attribute("as_of_date", String, "Balance calculation date", func() {
		Format(FormatDate)
	})
	Attribute("last_transaction_date", String, "Date of last transaction", func() {
		Format(FormatDate)
	})

	Required("account_id", "account_code", "account_name", "current_balance", "total_debits", "total_credits", "as_of_date")
})

// Validation types
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
	Attribute("errors", ArrayOf(ValidationErrorResult), "Validation errors")
	Attribute("warnings", ArrayOf(ValidationWarningResult), "Validation warnings")
	Attribute("validation_level", String, "Validation level used")

	Required("is_valid", "is_balanced", "total_debits", "total_credits", "validation_level")
})

var ValidationErrorResult = Type("ValidationErrorResult", func() {
	Description("Validation error details")

	Attribute("field", String, "Field that failed validation")
	Attribute("message", String, "Error message")
	Attribute("code", String, "Error code")
	Attribute("severity", String, "Error severity", func() {
		Enum("ERROR", "WARNING", "INFO")
	})

	Required("field", "message", "code", "severity")
})

var ValidationWarningResult = Type("ValidationWarningResult", func() {
	Description("Validation warning details")

	Attribute("field", String, "Field with warning")
	Attribute("message", String, "Warning message")
	Attribute("code", String, "Warning code")

	Required("field", "message", "code")
})
