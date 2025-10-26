package design

import (
	. "goa.design/goa/v3/dsl"
)

// AccountRootType represents the fundamental account classification
var AccountRootType = func() {
	Enum("ASSET", "LIABILITY", "EQUITY", "REVENUE", "EXPENSE")
}

// AccountType represents detailed account categories
var AccountType = func() {
	Enum("BANK", "CASH", "RECEIVABLE", "PAYABLE", "EXPENSE", "REVENUE", "EQUITY",
		"INVENTORY", "FIXED_ASSET", "OTHER_CURRENT_ASSET", "OTHER_CURRENT_LIABILITY",
		"LONG_TERM_LIABILITY", "ACCUMULATED_DEPRECIATION")
}

// NormalBalance represents the normal balance side
var NormalBalance = func() {
	Enum("DEBIT", "CREDIT")
}

// TransactionStatus represents transaction lifecycle states
var TransactionStatus = func() {
	Enum("DRAFT", "PENDING_APPROVAL", "APPROVED", "POSTED", "REVERSED", "CANCELLED", "REJECTED")
}

// TransactionType represents different transaction categories
var TransactionType = func() {
	Enum("GENERAL_JOURNAL", "ACCOUNTS_PAYABLE", "ACCOUNTS_RECEIVABLE", "PAYROLL",
		"INVENTORY", "FIXED_ASSET", "BANK_TRANSFER", "ADJUSTMENT", "OPENING_BALANCE",
		"CLOSING", "REVERSAL", "RECURRING")
}

// ApprovalDecision represents approval workflow decisions
var ApprovalDecision = func() {
	Enum("APPROVE", "REJECT", "REQUEST_CHANGES")
}

// FinancialStatementSection represents sections in financial statements
var FinancialStatementSection = func() {
	Enum("BALANCE_SHEET_ASSETS", "BALANCE_SHEET_LIABILITIES", "BALANCE_SHEET_EQUITY",
		"INCOME_STATEMENT_REVENUE", "INCOME_STATEMENT_EXPENSES", "CASH_FLOW")
}

// CashFlowCategory represents cash flow statement categories
var CashFlowCategory = func() {
	Enum("OPERATING", "INVESTING", "FINANCING")
}

// ConsolidationMethod represents how group balances are calculated
var ConsolidationMethod = func() {
	Enum("SUM", "AVERAGE", "MAX", "MIN", "CUSTOM")
}

// =============================================================================
// CORE FINANCIAL DOMAIN MODELS - Reusable base types
// =============================================================================

// FinancialAmount represents a monetary amount with currency
var FinancialAmount = Type("FinancialAmount", func() {
	Description("Financial amount with currency information")
	Attribute("amount", String, "Amount as decimal string", func() {
		Pattern(DecimalPattern)
		Example("1234.56")
	})
	Attribute("currency", String, "Currency code (ISO 4217)", func() {
		Pattern(CurrencyCodePattern)
		Example("USD")
	})
	Required("amount", "currency")
})

// AccountReference represents a lightweight account reference
var AccountReference = Type("AccountReference", func() {
	Description("Lightweight reference to an account")
	Attribute("id", String, "Account ID", func() {
		Format(FormatUUID)
	})
	Attribute("code", String, "Account code", func() {
		Example("1100")
	})
	Attribute("name", String, "Account name", func() {
		Example("Cash - Operating Account")
	})
	Attribute("root_type", String, "Root account type", AccountRootType)
	Attribute("normal_balance", String, "Normal balance side", NormalBalance)
	Required("id", "code", "name", "root_type", "normal_balance")
})

// JournalEntry represents a single journal entry within a transaction
var JournalEntry = Type("JournalEntry", func() {
	Description("Individual journal entry (debit or credit)")
	Attribute("id", String, "Entry ID", func() {
		Format(FormatUUID)
	})
	Attribute("account", AccountReference, "Account being debited/credited")
	Attribute("debit_amount", String, "Debit amount (if any)", func() {
		Pattern(DecimalPattern)
		Example("0.00")
	})
	Attribute("credit_amount", String, "Credit amount (if any)", func() {
		Pattern(DecimalPattern)
		Example("1234.56")
	})
	Attribute("description", String, "Entry description", func() {
		MaxLength(1000)
		Example("Payment for services rendered")
	})
	Attribute("reference", String, "Entry reference", func() {
		MaxLength(255)
		Example("INV-2025-001")
	})
	Required("id", "account", "debit_amount", "credit_amount")
})

// FinancialTransaction represents a complete financial transaction
var FinancialTransaction = Type("FinancialTransaction", func() {
	Description("Complete financial transaction with entries")
	Attribute("id", String, "Transaction ID", func() {
		Format(FormatUUID)
	})
	Attribute("transaction_number", String, "Unique transaction number", func() {
		Example("TXN-2025-001234")
	})
	Attribute("type", String, "Transaction type", TransactionType)
	Attribute("status", String, "Current status", TransactionStatus)
	Attribute("date", String, "Transaction date", func() {
		Format(FormatDate)
		Example("2025-01-15")
	})
	Attribute("description", String, "Transaction description", func() {
		MaxLength(1000)
		Example("Monthly service payment")
	})
	Attribute("reference", String, "External reference", func() {
		MaxLength(255)
		Example("PO-12345")
	})
	Attribute("total_amount", FinancialAmount, "Total transaction amount")
	Attribute("entries", ArrayOf(JournalEntry), "Journal entries")
	Attribute("attachments", ArrayOf(String), "Attachment URLs")
	AuditFields()
	Required("id", "transaction_number", "type", "status", "date", "total_amount", "entries")
})

// ValidationIssue represents a validation problem
var ValidationIssue = Type("ValidationIssue", func() {
	Description("Financial validation issue")
	Attribute("severity", String, "Issue severity", func() {
		Enum("ERROR", "WARNING", "INFO")
		Example("ERROR")
	})
	Attribute("code", String, "Validation rule code", func() {
		Example("UNBALANCED_ENTRIES")
	})
	Attribute("message", String, "Human-readable message", func() {
		Example("Transaction entries do not balance: debits $1000.00, credits $950.00")
	})
	Attribute("field", String, "Related field (if applicable)", func() {
		Example("entries[1].credit_amount")
	})
	Attribute("context", MapOf(String, Any), "Additional context data")
	Required("severity", "code", "message")
})

// =============================================================================
// WORKFLOW AND APPROVAL TYPES
// =============================================================================

// ApprovalStage represents a stage in the approval workflow
var ApprovalStage = Type("ApprovalStage", func() {
	Description("Approval workflow stage")
	Attribute("id", String, "Stage ID", func() {
		Format(FormatUUID)
	})
	Attribute("name", String, "Stage name", func() {
		Example("Finance Manager Review")
	})
	Attribute("sequence", Int32, "Stage sequence number")
	Attribute("required_approvers", Int32, "Number of required approvers", func() {
		Minimum(1)
		Default(1)
	})
	Attribute("current_approvers", Int32, "Current number of approvers")
	Attribute("status", String, "Stage status", func() {
		Enum("PENDING", "IN_PROGRESS", "APPROVED", "REJECTED", "SKIPPED")
	})
	Attribute("due_date", String, "Stage due date", func() {
		Format(FormatDateTime)
	})
	Required("id", "name", "sequence", "required_approvers", "current_approvers", "status")
})

// WorkflowAction represents available actions in workflow
var WorkflowAction = Type("WorkflowAction", func() {
	Description("Available workflow action")
	Attribute("action", String, "Action name", func() {
		Example("approve")
	})
	Attribute("label", String, "Display label", func() {
		Example("Approve Transaction")
	})
	Attribute("description", String, "Action description", func() {
		Example("Approve this transaction for posting")
	})
	Attribute("requires_comment", Boolean, "Whether comment is required", func() {
		Default(false)
	})
	Required("action", "label")
})

// =============================================================================
// REPORTING TYPES
// =============================================================================

// ReportPeriod represents a reporting period
var ReportPeriod = Type("ReportPeriod", func() {
	Description("Financial reporting period")
	Attribute("start_date", String, "Period start date", func() {
		Format(FormatDate)
		Example("2025-01-01")
	})
	Attribute("end_date", String, "Period end date", func() {
		Format(FormatDate)
		Example("2025-01-31")
	})
	Attribute("period_type", String, "Type of period", func() {
		Enum("DAILY", "WEEKLY", "MONTHLY", "QUARTERLY", "YEARLY", "CUSTOM")
		Example("MONTHLY")
	})
	Attribute("period_name", String, "Human-readable period name", func() {
		Example("January 2025")
	})
	Required("start_date", "end_date", "period_type")
})

// BalanceSummary represents account balance information
var BalanceSummary = Type("BalanceSummary", func() {
	Description("Account balance summary")
	Attribute("opening_balance", String, "Opening balance", func() {
		Pattern(DecimalPattern)
		Example("1000.00")
	})
	Attribute("total_debits", String, "Total debit transactions", func() {
		Pattern(DecimalPattern)
		Example("5000.00")
	})
	Attribute("total_credits", String, "Total credit transactions", func() {
		Pattern(DecimalPattern)
		Example("3000.00")
	})
	Attribute("closing_balance", String, "Closing balance", func() {
		Pattern(DecimalPattern)
		Example("3000.00")
	})
	Attribute("net_change", String, "Net change during period", func() {
		Pattern(DecimalPattern)
		Example("2000.00")
	})
	Required("opening_balance", "total_debits", "total_credits", "closing_balance", "net_change")
})

// =============================================================================
// UTILITY TYPES
// =============================================================================

// FinancialValidationResult represents validation outcome
var FinancialValidationResult = Type("FinancialValidationResult", func() {
	Description("Financial data validation result")
	Attribute("is_valid", Boolean, "Whether data passed validation")
	Attribute("issues", ArrayOf(ValidationIssue), "Validation issues found")
	Attribute("warnings_count", Int32, "Number of warnings")
	Attribute("errors_count", Int32, "Number of errors")
	Attribute("validated_at", String, "Validation timestamp", func() {
		Format(FormatDateTime)
	})
	Required("is_valid", "issues", "warnings_count", "errors_count", "validated_at")
})

// ExchangeRate represents currency exchange rate
var ExchangeRate = Type("ExchangeRate", func() {
	Description("Currency exchange rate information")
	Attribute("from_currency", String, "Source currency", func() {
		Pattern(CurrencyCodePattern)
		Example("USD")
	})
	Attribute("to_currency", String, "Target currency", func() {
		Pattern(CurrencyCodePattern)
		Example("EUR")
	})
	Attribute("rate", String, "Exchange rate", func() {
		Pattern(DecimalPattern)
		Example("0.8456")
	})
	Attribute("effective_date", String, "Rate effective date", func() {
		Format(FormatDate)
		Example("2025-01-15")
	})
	Attribute("source", String, "Rate source", func() {
		Example("ECB")
	})
	Required("from_currency", "to_currency", "rate", "effective_date")
})
