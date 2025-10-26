package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// FINANCIAL ACCOUNT TYPES
// ============================================================================

// FinanceAccount represents a financial account in the chart of accounts.
// Supports hierarchical account structures, multi-currency operations,
// and comprehensive financial reporting capabilities.
var FinanceAccount = ResultType("application/vnd.erp.finance.account", func() {
	Description("Financial account in the chart of accounts with comprehensive accounting features and multi-currency support")
	
	Attributes(func() {
		Field(1, "id", String, "Unique account identifier", func() {
			Format(FormatUUID)
			Example("account-123e4567-e89b-12d3-a456-426614174000")
			Description("Primary key for account references")
		})
		
		Field(2, "tenant_id", String, "Associated tenant identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
			Description("Tenant isolation boundary")
		})
		
		Field(3, "entity_id", String, "Associated entity identifier", func() {
			Format(FormatUUID)
			Example("987fcdeb-51d2-43b8-a456-426614174000")
			Description("Entity that owns this account")
		})
		
		Field(4, "account_code", String, "Unique account code within entity", func() {
			Pattern("^[0-9]{1,4}(-[0-9]{1,4})*$")
			MinLength(1)
			MaxLength(20)
			Example("1100-001")
			Description("Hierarchical numeric code following accounting standards")
		})
		
		Field(5, "account_name", String, "Account display name", func() {
			MinLength(2)
			MaxLength(100)
			Example("Cash - Operating Account")
			Description("Human-readable account name")
		})
		
		Field(6, "account_description", String, "Detailed account description", func() {
			MaxLength(500)
			Example("Primary operating cash account for daily business transactions")
			Description("Comprehensive account purpose and usage explanation")
		})
		
		Field(7, "root_type", String, "Fundamental accounting classification", func() {
			Enum("ASSET", "LIABILITY", "EQUITY", "INCOME", "EXPENSE")
			Example("ASSET")
			Description("Primary accounting equation classification")
		})
		
		Field(8, "account_type", String, "Specific account type classification", func() {
			Enum("CURRENT_ASSET", "FIXED_ASSET", "CURRENT_LIABILITY", "LONG_TERM_LIABILITY",
				"OWNERS_EQUITY", "RETAINED_EARNINGS", "OPERATING_INCOME", "NON_OPERATING_INCOME",
				"OPERATING_EXPENSE", "NON_OPERATING_EXPENSE", "COST_OF_GOODS_SOLD")
			Example("CURRENT_ASSET")
			Description("Detailed sub-classification for financial reporting")
		})
		
		Field(9, "account_category", String, "Business category for grouping", func() {
			Enum("CASH_EQUIVALENTS", "ACCOUNTS_RECEIVABLE", "INVENTORY", "PREPAID_EXPENSES",
				"PROPERTY_PLANT_EQUIPMENT", "ACCOUNTS_PAYABLE", "ACCRUED_LIABILITIES",
				"SALES_REVENUE", "SERVICE_REVENUE", "ADMINISTRATIVE_EXPENSE", "SELLING_EXPENSE")
			Example("CASH_EQUIVALENTS")
			Description("Business-oriented categorization for reporting")
		})
		
		Field(10, "normal_balance", String, "Natural balance side for account", func() {
			Enum("DEBIT", "CREDIT")
			Example("DEBIT")
			Description("Accounting side where increases are recorded")
		})
		
		Field(11, "current_balance", String, "Current account balance", func() {
			Pattern("^-?\\d+(\\.\\d{1,4})?$")
			Example("15750.25")
			Description("Current balance in account's base currency")
		})
		
		Field(12, "base_currency", String, "Account's base currency", func() {
			Pattern("^[A-Z]{3}$")
			Default("USD")
			Example("USD")
			Description("ISO 4217 currency code for account operations")
		})
		
		Field(13, "is_active", Boolean, "Account operational status", func() {
			Default(true)
			Example(true)
			Description("Whether account accepts new transactions")
		})
		
		Field(14, "is_control_account", Boolean, "Control account designation", func() {
			Default(false)
			Example(false)
			Description("Summary account that controls subsidiary ledgers")
		})
		
		Field(15, "is_system_account", Boolean, "System-managed account flag", func() {
			Default(false)
			Example(false)
			Description("Automatically managed by system processes")
		})
		
		Field(16, "parent_account_id", String, "Parent account for hierarchy", func() {
			Format(FormatUUID)
			Example("parent-account-456e7890-e89b-12d3-a456-426614174000")
			Description("Creates chart of accounts hierarchy")
		})
		
		Field(17, "account_level", UInt, "Hierarchical depth level", func() {
			Minimum(0)
			Maximum(10)
			Example(2)
			Description("Depth in chart of accounts hierarchy")
		})
		
		Field(18, "has_children", Boolean, "Has sub-account indicator", func() {
			Default(false)
			Example(true)
			Description("Whether account has subordinate accounts")
		})
		
		Field(19, "allows_transactions", Boolean, "Direct transaction posting allowed", func() {
			Default(true)
			Example(true)
			Description("Whether transactions can be posted directly")
		})
		
		Field(20, "requires_cost_center", Boolean, "Cost center requirement", func() {
			Default(false)
			Example(true)
			Description("Transactions must specify cost center")
		})
		
		Field(21, "requires_project", Boolean, "Project tracking requirement", func() {
			Default(false)
			Example(false)
			Description("Transactions must specify project")
		})
		
		Field(22, "tax_account", Boolean, "Tax-related account flag", func() {
			Default(false)
			Example(false)
			Description("Used for tax calculations and reporting")
		})
		
		Field(23, "bank_account_info", MapOf(String, Any), "Banking details for cash accounts", func() {
			Example(map[string]any{
				"bank_name":      "First National Bank",
				"account_number": "****1234",
				"routing_number": "021000021",
				"account_type":   "checking",
			})
			Description("Bank account details for reconciliation")
		})
		
		Field(24, "reconciliation_settings", MapOf(String, Any), "Reconciliation configuration", func() {
			Example(map[string]any{
				"auto_reconcile":        true,
				"tolerance_amount":      10.00,
				"reconcile_frequency":   "daily",
				"notification_enabled":  true,
			})
			Description("Automated reconciliation parameters")
		})
		
		Field(25, "reporting_settings", MapOf(String, Any), "Financial reporting configuration", func() {
			Example(map[string]any{
				"include_in_balance_sheet": true,
				"include_in_income_stmt":   false,
				"include_in_cash_flow":     true,
				"consolidation_method":     "line_by_line",
			})
			Description("Settings for financial statement inclusion")
		})
		
		Field(26, "budget_settings", MapOf(String, Any), "Budget tracking configuration", func() {
			Example(map[string]any{
				"budget_enabled":        true,
				"budget_period":         "annual",
				"variance_threshold":    0.10,
				"alert_on_overspend":   true,
			})
			Description("Budget monitoring and alerts")
		})
		
		Field(27, "opening_balance", String, "Opening balance for period", func() {
			Pattern("^-?\\d+(\\.\\d{1,4})?$")
			Example("12500.00")
			Description("Balance at start of accounting period")
		})
		
		Field(28, "closing_balance", String, "Closing balance for period", func() {
			Pattern("^-?\\d+(\\.\\d{1,4})?$")
			Example("15750.25")
			Description("Balance at end of accounting period")
		})
		
		Field(29, "ytd_debits", String, "Year-to-date debit total", func() {
			Pattern("^\\d+(\\.\\d{1,4})?$")
			Example("45250.75")
			Description("Cumulative debits for current fiscal year")
		})
		
		Field(30, "ytd_credits", String, "Year-to-date credit total", func() {
			Pattern("^\\d+(\\.\\d{1,4})?$")
			Example("42000.50")
			Description("Cumulative credits for current fiscal year")
		})
		
		Field(31, "last_transaction_date", String, "Most recent transaction date", func() {
			Format(FormatDateTime)
			Example("2023-12-07T14:30:00Z")
			Description("Timestamp of latest account activity")
		})
		
		Field(32, "tags", ArrayOf(String), "Account classification tags", func() {
			Example([]string{"cash", "operating", "primary"})
			Description("Searchable tags for account categorization")
		})
		
		Field(33, "external_references", MapOf(String, String), "External system mappings", func() {
			Example(map[string]any{
				"quickbooks_id": "QB-1100-001",
				"sap_account":   "SAP-GL-1100001",
				"legacy_code":   "LEGACY-CASH-01",
			})
			Description("Integration references to external systems")
		})
		
		// Audit fields from common.go
		AuditFields()
		
		Required("id", "tenant_id", "entity_id", "account_code", "account_name", 
			"root_type", "account_type", "normal_balance", "base_currency", 
			"is_active", "created_at")
	})
	
	View("default", func() {
		Description("Standard account view for listings and general display")
		Attribute("id")
		Attribute("account_code")
		Attribute("account_name")
		Attribute("root_type")
		Attribute("account_type")
		Attribute("current_balance")
		Attribute("base_currency")
		Attribute("is_active")
		Attribute("last_transaction_date")
	})
	
	View("detailed", func() {
		Description("Complete account view with all configuration and balances")
		Attribute("id")
		Attribute("tenant_id")
		Attribute("entity_id")
		Attribute("account_code")
		Attribute("account_name")
		Attribute("account_description")
		Attribute("root_type")
		Attribute("account_type")
		Attribute("account_category")
		Attribute("normal_balance")
		Attribute("current_balance")
		Attribute("base_currency")
		Attribute("is_active")
		Attribute("is_control_account")
		Attribute("parent_account_id")
		Attribute("account_level")
		Attribute("has_children")
		Attribute("allows_transactions")
		Attribute("bank_account_info")
		Attribute("reconciliation_settings")
		Attribute("reporting_settings")
		Attribute("budget_settings")
		Attribute("opening_balance")
		Attribute("closing_balance")
		Attribute("ytd_debits")
		Attribute("ytd_credits")
		Attribute("last_transaction_date")
		Attribute("tags")
		Attribute("external_references")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})
	
	View("hierarchy", func() {
		Description("Hierarchical view for chart of accounts display")
		Attribute("id")
		Attribute("parent_account_id")
		Attribute("account_code")
		Attribute("account_name")
		Attribute("root_type")
		Attribute("account_level")
		Attribute("has_children")
		Attribute("is_active")
	})
	
	View("balance", func() {
		Description("Balance-focused view for financial reporting")
		Attribute("id")
		Attribute("account_code")
		Attribute("account_name")
		Attribute("root_type")
		Attribute("account_type")
		Attribute("current_balance")
		Attribute("opening_balance")
		Attribute("closing_balance")
		Attribute("base_currency")
	})
	
	View("summary", func() {
		Description("Minimal view for references and dropdowns")
		Attribute("id")
		Attribute("account_code")
		Attribute("account_name")
		Attribute("root_type")
		Attribute("is_active")
	})
})

// AccountBalance represents account balance information for a specific period.
var AccountBalance = Type("AccountBalance", func() {
	Description("Account balance details for specific time periods with drill-down capabilities")
	
	Field(1, "account_id", String, "Associated account identifier", func() {
		Format(FormatUUID)
		Example("account-123e4567-e89b-12d3-a456-426614174000")
		Description("Account for which balance is calculated")
	})
	
	Field(2, "period_start", String, "Balance period start date", func() {
		Format(FormatDateTime)
		Example("2023-12-01T00:00:00Z")
		Description("Beginning of balance calculation period")
	})
	
	Field(3, "period_end", String, "Balance period end date", func() {
		Format(FormatDateTime)
		Example("2023-12-31T23:59:59Z")
		Description("End of balance calculation period")
	})
	
	Field(4, "opening_balance", String, "Balance at period start", func() {
		Pattern("^-?\\d+(\\.\\d{1,4})?$")
		Example("12500.00")
		Description("Account balance at beginning of period")
	})
	
	Field(5, "closing_balance", String, "Balance at period end", func() {
		Pattern("^-?\\d+(\\.\\d{1,4})?$")
		Example("15750.25")
		Description("Account balance at end of period")
	})
	
	Field(6, "period_debits", String, "Total debits during period", func() {
		Pattern("^\\d+(\\.\\d{1,4})?$")
		Example("8250.75")
		Description("Sum of all debit transactions in period")
	})
	
	Field(7, "period_credits", String, "Total credits during period", func() {
		Pattern("^\\d+(\\.\\d{1,4})?$")
		Example("5000.50")
		Description("Sum of all credit transactions in period")
	})
	
	Field(8, "transaction_count", UInt, "Number of transactions in period", func() {
		Example(47)
		Description("Total transaction count for the period")
	})
	
	Field(9, "currency", String, "Balance currency", func() {
		Pattern("^[A-Z]{3}$")
		Example("USD")
		Description("ISO 4217 currency code for balances")
	})
	
	Field(10, "calculated_at", String, "Balance calculation timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T15:00:00Z")
		Description("When balance was computed")
	})
	
	Required("account_id", "period_start", "period_end", "opening_balance", 
		"closing_balance", "period_debits", "period_credits", "currency", "calculated_at")
})

// AccountReconciliation represents bank reconciliation information.
var AccountReconciliation = Type("AccountReconciliation", func() {
	Description("Bank reconciliation record for cash accounts with variance tracking")
	
	Field(1, "id", String, "Unique reconciliation identifier", func() {
		Format(FormatUUID)
		Example("recon-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary key for reconciliation record")
	})
	
	Field(2, "account_id", String, "Associated account identifier", func() {
		Format(FormatUUID)
		Example("account-123e4567-e89b-12d3-a456-426614174000")
		Description("Cash account being reconciled")
	})
	
	Field(3, "reconciliation_date", String, "Reconciliation as-of date", func() {
		Format(FormatDateTime)
		Example("2023-12-31T23:59:59Z")
		Description("Cut-off date for reconciliation")
	})
	
	Field(4, "book_balance", String, "General ledger balance", func() {
		Pattern("^-?\\d+(\\.\\d{1,4})?$")
		Example("15750.25")
		Description("Balance per accounting records")
	})
	
	Field(5, "bank_balance", String, "Bank statement balance", func() {
		Pattern("^-?\\d+(\\.\\d{1,4})?$")
		Example("16125.75")
		Description("Balance per bank statement")
	})
	
	Field(6, "reconciled_balance", String, "Final reconciled balance", func() {
		Pattern("^-?\\d+(\\.\\d{1,4})?$")
		Example("15750.25")
		Description("Balance after reconciling items")
	})
	
	Field(7, "outstanding_deposits", String, "Deposits in transit", func() {
		Pattern("^\\d+(\\.\\d{1,4})?$")
		Example("1250.00")
		Description("Deposits recorded but not yet cleared")
	})
	
	Field(8, "outstanding_checks", String, "Outstanding checks total", func() {
		Pattern("^\\d+(\\.\\d{1,4})?$")
		Example("1625.50")
		Description("Checks issued but not yet cleared")
	})
	
	Field(9, "bank_fees", String, "Bank fees and charges", func() {
		Pattern("^\\d+(\\.\\d{1,4})?$")
		Example("25.00")
		Description("Bank charges not yet recorded")
	})
	
	Field(10, "interest_earned", String, "Interest earned", func() {
		Pattern("^\\d+(\\.\\d{1,4})?$")
		Example("15.75")
		Description("Interest income not yet recorded")
	})
	
	Field(11, "reconciliation_status", String, "Reconciliation completion status", func() {
		Enum("IN_PROGRESS", "COMPLETED", "FAILED", "REQUIRES_REVIEW")
		Default("IN_PROGRESS")
		Example("COMPLETED")
		Description("Current state of reconciliation process")
	})
	
	Field(12, "variance_amount", String, "Unresolved variance", func() {
		Pattern("^-?\\d+(\\.\\d{1,4})?$")
		Example("0.00")
		Description("Difference that couldn't be reconciled")
	})
	
	Field(13, "reconciled_items", ArrayOf(String), "Reconciled transaction IDs", func() {
		Elem(func() {
			Format(FormatUUID)
		})
		Example([]string{
			"txn-123e4567-e89b-12d3-a456-426614174000",
			"txn-456e7890-e89b-12d3-a456-426614174000",
		})
		Description("Transactions matched during reconciliation")
	})
	
	Field(14, "exceptions", ArrayOf(String), "Unreconciled item descriptions", func() {
		Example([]string{
			"Check #1234 for $500.00 not yet cleared",
			"Unknown deposit $75.00 on bank statement",
		})
		Description("Items requiring manual review")
	})
	
	Field(15, "notes", String, "Reconciliation notes", func() {
		MaxLength(1000)
		Example("Month-end reconciliation completed. Minor timing differences resolved.")
		Description("Additional comments and explanations")
	})
	
	// Audit fields
	AuditFields()
	
	Required("id", "account_id", "reconciliation_date", "book_balance", "bank_balance", 
		"reconciled_balance", "reconciliation_status", "created_at")
})