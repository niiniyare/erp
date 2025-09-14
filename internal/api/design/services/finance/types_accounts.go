package finance

import (
	. "github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// =============================================================================
// LEGACY ACCOUNT PAYLOADS (from transactions.go service)
// =============================================================================
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

	// Operational settings
	Attribute("is_active", Boolean, "Whether account is active")
	Attribute("allow_manual_entries", Boolean, "Allow manual journal entries")
	Attribute("require_reference", Boolean, "Require reference for entries")

	// Reporting and display
	Attribute("display_order", Int32, "Display order in UI/reports", func() {
		Minimum(0)
	})
	Attribute("show_in_reports", Boolean, "Whether to include in standard reports")
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

	// Classification
	Attribute("root_type", String, "Root account type")
	Attribute("account_type", String, "Account type")
	Attribute("normal_balance", String, "Normal balance side")

	// Operational settings
	Attribute("is_active", Boolean, "Whether account is active")
	Attribute("allow_manual_entries", Boolean, "Whether manual entries are allowed")
	Attribute("require_reference", Boolean, "Whether reference is required")

	// Balance tracking
	Attribute("current_balance", String, "Current account balance")
	Attribute("ytd_balance", String, "Year-to-date balance")
	Attribute("last_transaction_date", String, "Last transaction date", func() {
		Format(FormatDateTime)
	})

	// Currency and localization
	Attribute("currency_code", String, "Primary currency code")

	// Use common audit fields
	AuditFields()

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

	// Use common pagination
	Attribute("pagination", Pagination, "Pagination parameters")
})

var AccountListResult = Type("AccountListResult", func() {
	Description("List of accounts with pagination info")

	Attribute("accounts", ArrayOf(AccountResult), "List of accounts")
	Attribute("pagination", PaginationMeta, "Pagination metadata")

	Required("accounts", "pagination")
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

// CreateAccountPayload for legacy account creation endpoint
var CreateAccountPayload = Type("CreateAccountPayload", func() {
	Description("Payload for creating a new chart of accounts entry (legacy)")

	Attribute("entity_id", String, "Entity ID (optional)", func() {
		Format(FormatUUID)
	})
	Attribute("account_code", String, "Unique account code", func() {
		MinLength(1)
		MaxLength(50)
		Example("1100")
	})
	Attribute("account_name", String, "Account name", func() {
		MinLength(1)
		MaxLength(255)
		Example("Cash - Operating Account")
	})
	Attribute("account_description", String, "Account description", func() {
		MaxLength(1000)
	})
	Attribute("parent_account_id", String, "Parent account ID (optional)", func() {
		Format(FormatUUID)
	})
	Attribute("root_type", String, "Root account type", func() {
		Enum("ASSET", "LIABILITY", "EQUITY", "REVENUE", "EXPENSE")
		Example("ASSET")
	})
	Attribute("account_type", String, "Detailed account type", func() {
		Enum("BANK", "CASH", "RECEIVABLE", "PAYABLE", "EXPENSE", "REVENUE", "EQUITY", "INVENTORY", "FIXED_ASSET", "OTHER_CURRENT_ASSET", "OTHER_CURRENT_LIABILITY")
		Example("BANK")
	})
	Attribute("normal_balance", String, "Normal balance side", func() {
		Enum("DEBIT", "CREDIT")
		Example("DEBIT")
	})
	Attribute("currency_code", String, "Primary currency code", func() {
		Pattern(CurrencyCodePattern)
		Example("USD")
		Default("USD")
	})
	Attribute("is_active", Boolean, "Whether account is active", func() {
		Default(true)
	})
	Attribute("allow_manual_entries", Boolean, "Allow manual journal entries", func() {
		Default(true)
	})
	Attribute("require_reference", Boolean, "Require reference for entries", func() {
		Default(false)
	})
	Attribute("cash_flow_type", String, "Cash flow statement classification (optional)", func() {
		Enum("OPERATING", "INVESTING", "FINANCING")
	})

	Required("account_code", "account_name", "root_type", "account_type", "normal_balance")
})

// Inline payloads from accounts.go service
var GetAccountNodeByIdPayload = Type("GetAccountNodeByIdPayload", func() {
	Description("Payload for getting account node by ID")
	Attribute("id", String, "Node ID", func() {
		Format(FormatUUID)
	})
	Required("id")
})

var GetAccountNodeByCodePayload = Type("GetAccountNodeByCodePayload", func() {
	Description("Payload for getting account node by code")
	Attribute("code", String, "Node code", func() {
		Example("1100")
	})
	Required("code")
})

var DeleteAccountNodePayload = Type("DeleteAccountNodePayload", func() {
	Description("Payload for deleting account node")
	Attribute("id", String, "Node ID", func() {
		Format(FormatUUID)
	})
	Required("id")
})

var GetAccountBalancePayload = Type("GetAccountBalancePayload", func() {
	Description("Payload for getting account balance")
	Attribute("account_id", String, "Account ID", func() {
		Format(FormatUUID)
	})
	Attribute("as_of_date", String, "Balance as of date (optional)", func() {
		Format(FormatDate)
	})
	Required("account_id")
})

var GetHierarchyAnalysisPayload = Type("GetHierarchyAnalysisPayload", func() {
	Description("Payload for hierarchy analysis")
	Attribute("parent_id", String, "Parent node ID", func() {
		Format(FormatUUID)
	})
	Attribute("include_balance_data", Boolean, "Include balance information", func() {
		Default(true)
	})
	Attribute("max_depth", Int32, "Maximum depth to analyze")
	Attribute("as_of_date", String, "Analysis date", func() {
		Format(FormatDate)
	})
	Required("parent_id")
})

var HierarchyAnalysisResult = Type("HierarchyAnalysisResult", func() {
	Description("Hierarchy analysis result")
	Attribute("parent_id", String, func() { Format(FormatUUID) })
	Attribute("hierarchy_depth", Int32)
	Attribute("total_accounts", Int32)
	Attribute("total_groups", Int32)
	Attribute("total_balance", String)
	Attribute("analysis_date", String, func() { Format(FormatDate) })
	Required("parent_id", "hierarchy_depth", "total_accounts", "total_groups", "total_balance")
})

// Legacy account inline payloads from transactions.go
var GetAccountByIdPayload = Type("GetAccountByIdPayload", func() {
	Description("Payload for getting account by ID (legacy)")
	Attribute("id", String, "Account ID", func() {
		Format(FormatUUID)
	})
	Required("id")
})

var GetAccountByCodePayload = Type("GetAccountByCodePayload", func() {
	Description("Payload for getting account by code (legacy)")
	Attribute("account_code", String, "Account code", func() {
		Example("1100")
	})
	Required("account_code")
})

var GetAccountByNamePayload = Type("GetAccountByNamePayload", func() {
	Description("Payload for getting account by name (legacy)")
	Attribute("account_name", String, "Account name", func() {
		Example("Cash - Operating Account")
	})
	Required("account_name")
})

var DeleteAccountPayload = Type("DeleteAccountPayload", func() {
	Description("Payload for deleting account (legacy)")
	Attribute("id", String, "Account ID", func() {
		Format(FormatUUID)
	})
	Required("id")
})

var GetAccountHierarchyPayload = Type("GetAccountHierarchyPayload", func() {
	Description("Payload for getting account hierarchy (legacy)")
	Attribute("root_id", String, "Root account ID (optional)", func() {
		Format(FormatUUID)
	})
})

// =============================================================================
// UNIFIED ACCOUNT NODE TYPES
// =============================================================================

var CreateAccountNodePayload = Type("CreateAccountNodePayload", func() {
	Description("Unified payload for creating accounts or account groups")

	// Discriminator field - determines if creating account or group
	Attribute("node_type", String, "Node type discriminator (account or group)", func() {
		Enum("account", "group")
		Example("account")
	})

	// Common fields for both accounts and groups
	Attribute("entity_id", String, "Entity ID (optional)", func() {
		Format(FormatUUID)
	})
	Attribute("code", String, "Unique code (account_code for accounts, group_code for groups)", func() {
		MinLength(1)
		MaxLength(50)
		Example("1100")
	})
	Attribute("name", String, "Name (account_name for accounts, group_name for groups)", func() {
		MinLength(1)
		MaxLength(255)
		Example("Cash - Operating Account")
	})
	Attribute("description", String, "Description (optional)", func() {
		MaxLength(1000)
	})
	Attribute("parent_id", String, "Parent node ID (can be account or group)", func() {
		Format(FormatUUID)
	})

	// Account-specific fields (only when node_type = "account")
	Attribute("account_type", String, "Detailed account type (accounts only)", func() {
		Enum("BANK", "CASH", "RECEIVABLE", "PAYABLE", "EXPENSE", "REVENUE", "EQUITY", "INVENTORY", "FIXED_ASSET", "OTHER_CURRENT_ASSET", "OTHER_CURRENT_LIABILITY")
		Example("BANK")
	})
	Attribute("root_type", String, "Root type (accounts only)", func() {
		Enum("ASSET", "LIABILITY", "EQUITY", "REVENUE", "EXPENSE")
		Example("ASSET")
	})
	Attribute("normal_balance", String, "Normal balance side (accounts only)", func() {
		Enum("DEBIT", "CREDIT")
		Example("DEBIT")
	})
	Attribute("currency_code", String, "Primary currency code (accounts only)", func() {
		Pattern(CurrencyCodePattern)
		Example("USD")
	})
	Attribute("allows_manual_entries", Boolean, "Allow manual journal entries (accounts only)", func() {
		Default(true)
	})
	Attribute("requires_reconciliation", Boolean, "Require reconciliation (accounts only)", func() {
		Default(false)
	})

	// Group-specific fields (only when node_type = "group")
	Attribute("financial_statement_section", String, "Financial statement section (groups only)", func() {
		Enum("BALANCE_SHEET_ASSETS", "BALANCE_SHEET_LIABILITIES", "BALANCE_SHEET_EQUITY", "INCOME_STATEMENT_REVENUE", "INCOME_STATEMENT_EXPENSES", "CASH_FLOW")
		Example("BALANCE_SHEET_ASSETS")
	})
	Attribute("consolidation_method", String, "Balance consolidation method (groups only)", func() {
		Enum("SUM", "AVERAGE", "MAX", "MIN", "CUSTOM")
		Default("SUM")
	})
	Attribute("cash_flow_category", String, "Cash flow statement category (groups only)", func() {
		Enum("OPERATING", "INVESTING", "FINANCING")
		Example("OPERATING")
	})
	Attribute("display_order", Int32, "Display order in reports (groups only)", func() {
		Minimum(0)
		Default(0)
	})
	Attribute("is_header", Boolean, "Is header group (groups only)", func() {
		Default(false)
	})
	Attribute("show_totals", Boolean, "Show group totals (groups only)", func() {
		Default(true)
	})
	Attribute("indent_level", Int32, "Indentation level for display (groups only)", func() {
		Minimum(0)
		Maximum(10)
		Default(0)
	})

	// Common operational settings
	Attribute("is_active", Boolean, "Whether node is active", func() {
		Default(true)
	})

	Required("node_type", "code", "name")
})

var UpdateAccountNodePayload = Type("UpdateAccountNodePayload", func() {
	Description("Unified payload for updating accounts or account groups")

	Attribute("id", String, "Node ID", func() {
		Format(FormatUUID)
	})
	Attribute("name", String, "Node name", func() {
		MinLength(1)
		MaxLength(255)
	})
	Attribute("description", String, "Node description")

	// Account-specific updates
	Attribute("allows_manual_entries", Boolean, "Allow manual journal entries (accounts only)")
	Attribute("requires_reconciliation", Boolean, "Require reconciliation (accounts only)")

	// Group-specific updates
	Attribute("display_order", Int32, "Display order in reports (groups only)", func() {
		Minimum(0)
	})
	Attribute("show_totals", Boolean, "Show group totals (groups only)")
	Attribute("indent_level", Int32, "Indentation level (groups only)", func() {
		Minimum(0)
		Maximum(10)
	})

	// Common operational settings
	Attribute("is_active", Boolean, "Whether node is active")

	Required("id")
})

// Simplified result type
var AccountNodeResult = Type("AccountNodeResult", func() {
	Description("Unified account or group information")

	// Common identification fields
	Attribute("id", String, "Node ID", func() {
		Format(FormatUUID)
	})
	Attribute("node_type", String, "Node type discriminator", func() {
		Enum("account", "group")
	})
	Attribute("code", String, "Node code")
	Attribute("name", String, "Node name")
	Attribute("description", String, "Node description")
	Attribute("parent_id", String, "Parent node ID", func() {
		Format(FormatUUID)
	})

	// Hierarchy metadata
	Attribute("level", Int32, "Hierarchy level (1 = root)")
	Attribute("path", String, "Materialized path")
	Attribute("has_children", Boolean, "Whether node has child nodes")
	Attribute("child_count", Int32, "Number of direct children")
	Attribute("is_active", Boolean, "Whether node is active")

	// Account-specific data (populated only when node_type = "account")
	Attribute("account_type", String, "Account type (if account)")
	Attribute("root_type", String, "Root type (if account)")
	Attribute("normal_balance", String, "Normal balance side (if account)")
	Attribute("currency_code", String, "Currency code (if account)")
	Attribute("current_balance", String, "Current balance (if account)")
	Attribute("allows_manual_entries", Boolean, "Allows manual entries (if account)")
	Attribute("requires_reconciliation", Boolean, "Requires reconciliation (if account)")

	// Group-specific data (populated only when node_type = "group")
	Attribute("financial_statement_section", String, "Financial statement section (if group)")
	Attribute("consolidation_method", String, "Consolidation method (if group)")
	Attribute("cash_flow_category", String, "Cash flow category (if group)")
	Attribute("display_order", Int32, "Display order (if group)")
	Attribute("is_header", Boolean, "Is header group (if group)")
	Attribute("show_totals", Boolean, "Show totals (if group)")
	Attribute("indent_level", Int32, "Indent level (if group)")

	// Balance information (calculated field)
	Attribute("total_balance", String, "Total balance")

	// Use common audit fields
	AuditFields()

	Required("id", "node_type", "code", "name", "level", "path", "has_children", "child_count", "is_active")
})

// Simplified filtering payload
var ListAccountNodesPayload = Type("ListAccountNodesPayload", func() {
	Description("Unified payload for listing accounts and/or groups with filters")

	// Node type filtering
	Attribute("node_types", ArrayOf(String), "Filter by node types", func() {
		Elem(func() {
			Enum("account", "group")
		})
		Example([]string{"account", "group"})
	})

	// Hierarchy filtering
	Attribute("parent_id", String, "Filter by parent node ID", func() {
		Format(FormatUUID)
	})
	Attribute("max_level", Int32, "Maximum hierarchy depth to include")
	Attribute("include_children", Boolean, "Include child count and hierarchy info", func() {
		Default(false)
	})

	// Account-specific filters
	Attribute("root_type", String, "Filter by root type (accounts only)", func() {
		Enum("ASSET", "LIABILITY", "EQUITY", "REVENUE", "EXPENSE")
	})
	Attribute("account_type", String, "Filter by account type (accounts only)")

	// Group-specific filters
	Attribute("financial_statement_section", String, "Filter by financial statement section (groups only)")
	Attribute("cash_flow_category", String, "Filter by cash flow category (groups only)", func() {
		Enum("OPERATING", "INVESTING", "FINANCING")
	})

	// Common filters
	Attribute("is_active", Boolean, "Filter by active status")
	Attribute("search_query", String, "Search in codes, names, or descriptions")
	Attribute("include_balances", Boolean, "Include current balances", func() {
		Default(true)
	})

	// Use common pagination
	Attribute("pagination", Pagination, "Pagination parameters")

	// Sorting
	Attribute("sort_by", String, "Sort field", func() {
		Enum("code", "name", "level", "created_at", "updated_at")
		Default("code")
	})
	Attribute("sort_order", String, "Sort order", func() {
		Enum("asc", "desc")
		Default("asc")
	})
})

var AccountNodeListResult = Type("AccountNodeListResult", func() {
	Description("Unified list of account nodes with pagination")

	Attribute("nodes", ArrayOf(AccountNodeResult), "List of account nodes")
	Attribute("pagination", PaginationMeta, "Pagination metadata")

	Required("nodes", "pagination")
})

// Search payload
var SearchAccountNodesPayload = Type("SearchAccountNodesPayload", func() {
	Description("Payload for searching accounts and groups")

	Attribute("query", String, "Search term", func() {
		MinLength(1)
		MaxLength(100)
		Example("cash")
	})
	Attribute("node_types", ArrayOf(String), "Filter by node types", func() {
		Elem(func() {
			Enum("account", "group")
		})
	})
	Attribute("limit", Int32, "Maximum results", func() {
		Default(20)
		Minimum(1)
		Maximum(100)
	})
	Attribute("include_inactive", Boolean, "Include inactive nodes", func() {
		Default(false)
	})

	Required("query")
})

var SearchAccountNodesResult = Type("SearchAccountNodesResult", func() {
	Description("Search results with relevance scoring")

	Attribute("results", ArrayOf(SearchResultItem), "Search results")
	Attribute("total_results", Int32, "Total number of results")
	Attribute("search_duration_ms", Int32, "Search execution time")

	Required("results", "total_results", "search_duration_ms")
})

var SearchResultItem = Type("SearchResultItem", func() {
	Description("Individual search result")

	Attribute("id", String, "Node ID", func() {
		Format(FormatUUID)
	})
	Attribute("node_type", String, "Node type", func() {
		Enum("account", "group")
	})
	Attribute("code", String, "Node code")
	Attribute("name", String, "Node name")
	Attribute("path", String, "Hierarchy path")
	Attribute("match_type", String, "Type of match", func() {
		Enum("NAME", "CODE", "DESCRIPTION")
	})
	Attribute("relevance_score", Float64, "Relevance score (0-1)")
	Attribute("current_balance", String, "Current balance (if account)")
	Attribute("currency_code", String, "Currency code (if account)")
	Attribute("child_count", Int32, "Number of children (if group)")
	Attribute("total_balance", String, "Total balance (if group)")

	Required("id", "node_type", "code", "name", "path", "match_type", "relevance_score")
})

