package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Application-wide constants for the finance module
const (
	// Version information
	ModuleVersion = "1.0.0"
	APIVersion    = "v1"

	// Database-related constants
	DefaultPageSize = 50
	MaxPageSize     = 1000
	MinPageSize     = 1

	// Decimal precision for monetary calculations
	DefaultDecimalPrecision = 4 // Standard precision for currency calculations
	HighDecimalPrecision    = 8 // High precision for exchange rates
	DisplayDecimalPrecision = 2 // Precision for display purposes

	// Default currency settings
	DefaultBaseCurrency = "USD"
	DefaultExchangeRate = "1.0000"

	// Transaction numbering
	DefaultTransactionNumberPrefix  = "TXN"
	DefaultTransactionNumberPadding = 6

	// Account code settings
	DefaultAccountCodeLength = 10
	MaxAccountHierarchyDepth = 10

	// Validation limits
	MaxTransactionEntries = 1000
	MaxBatchSize          = 5000
	// MaxDescriptionLength     = 1000
	// MaxReferenceLength       = 100
	MaxAttributeValueLength = 500

	// Date range limits
	MinAccountingYear = 1900
	MaxAccountingYear = 2100

	// Reconciliation settings
	DefaultReconciliationTolerance = "0.01"
	MaxReconciliationAge           = 90 // days

	// Approval workflow
	DefaultApprovalTimeout = 7 // days
	MaxApprovalLevels      = 5

	// Reporting settings
	DefaultFiscalYearStart = 1  // January 1st
	DefaultReportTimeout   = 30 // seconds

	// Cache settings
	DefaultCacheExpiry        = 5 * time.Minute
	ExchangeRateCacheExpiry   = 1 * time.Hour
	AccountBalanceCacheExpiry = 1 * time.Minute

	// Rate limiting
	DefaultRateLimit       = 1000 // requests per hour
	BulkOperationRateLimit = 100  // bulk operations per hour
)

// Standard account types based on common accounting practices
const (
	// Asset account types
	AccountTypeCurrentAsset    = "CURRENT_ASSET"
	AccountTypeFixedAsset      = "FIXED_ASSET"
	AccountTypeIntangibleAsset = "INTANGIBLE_ASSET"
	AccountTypeInvestment      = "INVESTMENT"
	AccountTypeOtherAsset      = "OTHER_ASSET"

	// Liability account types
	AccountTypeCurrentLiability  = "CURRENT_LIABILITY"
	AccountTypeLongTermLiability = "LONG_TERM_LIABILITY"
	AccountTypeAccruedLiability  = "ACCRUED_LIABILITY"
	AccountTypeOtherLiability    = "OTHER_LIABILITY"

	// Equity account types
	AccountTypeShareCapital     = "SHARE_CAPITAL"
	AccountTypeRetainedEarnings = "RETAINED_EARNINGS"
	AccountTypeOtherEquity      = "OTHER_EQUITY"

	// Revenue account types
	AccountTypeOperatingRevenue    = "OPERATING_REVENUE"
	AccountTypeNonOperatingRevenue = "NON_OPERATING_REVENUE"
	AccountTypeOtherRevenue        = "OTHER_REVENUE"

	// Expense account types
	AccountTypeOperatingExpense    = "OPERATING_EXPENSE"
	AccountTypeNonOperatingExpense = "NON_OPERATING_EXPENSE"
	AccountTypeCostOfGoodsSold     = "COST_OF_GOODS_SOLD"
	AccountTypeOtherExpense        = "OTHER_EXPENSE"
)

// Standard account subtypes for more granular classification
const (
	// Current Asset subtypes
	AccountSubtypeCash                = "CASH"
	AccountSubtypePettyCache          = "PETTY_CASH"
	AccountSubtypeBankAccount         = "BANK_ACCOUNT"
	AccountSubtypeAccountsReceivable  = "ACCOUNTS_RECEIVABLE"
	AccountSubtypeInventory           = "INVENTORY"
	AccountSubtypePrepaidExpenses     = "PREPAID_EXPENSES"
	AccountSubtypeShortTermInvestment = "SHORT_TERM_INVESTMENT"

	// Fixed Asset subtypes
	AccountSubtypeBuildings         = "BUILDINGS"
	AccountSubtypeEquipment         = "EQUIPMENT"
	AccountSubtypeVehicles          = "VEHICLES"
	AccountSubtypeFurnitureFixtures = "FURNITURE_FIXTURES"
	AccountSubtypeAccumulatedDep    = "ACCUMULATED_DEPRECIATION"

	// Current Liability subtypes
	AccountSubtypeAccountsPayable = "ACCOUNTS_PAYABLE"
	AccountSubtypeAccruedExpenses = "ACCRUED_EXPENSES"
	AccountSubtypeShortTermDebt   = "SHORT_TERM_DEBT"
	AccountSubtypeTaxPayable      = "TAX_PAYABLE"
	AccountSubtypeWagesPayable    = "WAGES_PAYABLE"

	// Revenue subtypes
	AccountSubtypeSalesRevenue   = "SALES_REVENUE"
	AccountSubtypeServiceRevenue = "SERVICE_REVENUE"
	AccountSubtypeInterestIncome = "INTEREST_INCOME"
	AccountSubtypeDividendIncome = "DIVIDEND_INCOME"

	// Expense subtypes
	AccountSubtypeOfficeExpenses      = "OFFICE_EXPENSES"
	AccountSubtypeUtilities           = "UTILITIES"
	AccountSubtypeRent                = "RENT"
	AccountSubtypeSalariesWages       = "SALARIES_WAGES"
	AccountSubtypeInterestExpense     = "INTEREST_EXPENSE"
	AccountSubtypeDepreciationExpense = "DEPRECIATION_EXPENSE"
)

// Financial statement line mappings
const (
	// Balance Sheet lines
	FSLineCashEquivalents    = "CASH_EQUIVALENTS"
	FSLineAccountsReceivable = "ACCOUNTS_RECEIVABLE"
	FSLineInventory          = "INVENTORY"
	FSLineTotalCurrentAssets = "TOTAL_CURRENT_ASSETS"
	FSLinePropertyPlantEquip = "PROPERTY_PLANT_EQUIPMENT"
	FSLineTotalAssets        = "TOTAL_ASSETS"
	FSLineAccountsPayable    = "ACCOUNTS_PAYABLE"
	FSLineAccruedLiabilities = "ACCRUED_LIABILITIES"
	FSLineTotalCurrentLiab   = "TOTAL_CURRENT_LIABILITIES"
	FSLineLongTermDebt       = "LONG_TERM_DEBT"
	FSLineTotalLiabilities   = "TOTAL_LIABILITIES"
	FSLineShareholderEquity  = "SHAREHOLDER_EQUITY"
	FSLineTotalLiabEquity    = "TOTAL_LIAB_EQUITY"

	// Income Statement lines
	FSLineRevenue           = "REVENUE"
	FSLineCostOfRevenue     = "COST_OF_REVENUE"
	FSLineGrossProfit       = "GROSS_PROFIT"
	FSLineOperatingExpenses = "OPERATING_EXPENSES"
	FSLineOperatingIncome   = "OPERATING_INCOME"
	FSLineInterestExpense   = "INTEREST_EXPENSE"
	FSLineNetIncome         = "NET_INCOME"

	// Cash Flow Statement lines
	FSLineCashFromOperations = "CASH_FROM_OPERATIONS"
	FSLineCashFromInvesting  = "CASH_FROM_INVESTING"
	FSLineCashFromFinancing  = "CASH_FROM_FINANCING"
	FSLineNetCashFlow        = "NET_CASH_FLOW"
)

// Error codes for standardized error handling
const (
	// General error codes
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeInternalError    = "INTERNAL_ERROR"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodeConflict         = "CONFLICT"
	ErrCodeBadRequest       = "BAD_REQUEST"

	// Business logic error codes
	ErrCodeTransactionUnbalanced = "TRANSACTION_UNBALANCED"
	ErrCodeAccountInactive       = "ACCOUNT_INACTIVE"
	ErrCodeInsufficientFunds     = "INSUFFICIENT_FUNDS"
	ErrCodeClosedPeriod          = "CLOSED_PERIOD"
	ErrCodeDuplicateEntry        = "DUPLICATE_ENTRY"
	ErrCodeInvalidStatus         = "INVALID_STATUS"
	ErrCodeApprovalRequired      = "APPROVAL_REQUIRED"
	ErrCodeAlreadyPosted         = "ALREADY_POSTED"
	ErrCodeCannotReverse         = "CANNOT_REVERSE"

	// Data integrity error codes
	ErrCodeInvalidReference    = "INVALID_REFERENCE"
	ErrCodeCircularReference   = "CIRCULAR_REFERENCE"
	ErrCodeConstraintViolation = "CONSTRAINT_VIOLATION"
	ErrCodeDataCorruption      = "DATA_CORRUPTION"

	// Currency and exchange rate error codes
	ErrCodeUnsupportedCurrency  = "UNSUPPORTED_CURRENCY"
	ErrCodeExchangeRateNotFound = "EXCHANGE_RATE_NOT_FOUND"
	ErrCodeCurrencyMismatch     = "CURRENCY_MISMATCH"

	// Reconciliation error codes
	ErrCodeReconciliationMismatch = "RECONCILIATION_MISMATCH"
	ErrCodeAlreadyReconciled      = "ALREADY_RECONCILED"
	ErrCodeReconciliationLocked   = "RECONCILIATION_LOCKED"
)

// Default values for various entities
var (
	// Default decimal values
	DecimalZero    = decimal.NewFromInt(0)
	DecimalOne     = decimal.NewFromInt(1)
	DecimalHundred = decimal.NewFromInt(100)

	// Default exchange rate (1.0)
	DefaultExchangeRateDecimal = decimal.NewFromFloat(1.0)

	// Default budget variance threshold (5%)
	DefaultBudgetVarianceThreshold = decimal.NewFromFloat(5.0)

	// Minimum and maximum exchange rates
	MinExchangeRateDecimal = decimal.NewFromFloat(0.000001)
	MaxExchangeRateDecimal = decimal.NewFromFloat(1000000.0)

	// Default reconciliation tolerance
	DefaultReconciliationToleranceDecimal = decimal.NewFromFloat(0.01)
)

// Recurring frequency patterns
const (
	RecurringFrequencyDaily     = "DAILY"
	RecurringFrequencyWeekly    = "WEEKLY"
	RecurringFrequencyMonthly   = "MONTHLY"
	RecurringFrequencyQuarterly = "QUARTERLY"
	RecurringFrequencyYearly    = "YEARLY"
)

// System account codes for standard accounts
const (
	SystemAccountCodeCash               = "1000"
	SystemAccountCodeAccountsReceivable = "1200"
	SystemAccountCodeInventory          = "1300"
	SystemAccountCodeAccountsPayable    = "2000"
	SystemAccountCodeRetainedEarnings   = "3200"
	SystemAccountCodeSalesRevenue       = "4000"
	SystemAccountCodeCostOfGoodsSold    = "5000"
)

// Audit event types for tracking changes
const (
	AuditEventAccountCreated         = "ACCOUNT_CREATED"
	AuditEventAccountUpdated         = "ACCOUNT_UPDATED"
	AuditEventAccountDeactivated     = "ACCOUNT_DEACTIVATED"
	AuditEventTransactionCreated     = "TRANSACTION_CREATED"
	AuditEventTransactionUpdated     = "TRANSACTION_UPDATED"
	AuditEventTransactionApproved    = "TRANSACTION_APPROVED"
	AuditEventTransactionRejected    = "TRANSACTION_REJECTED"
	AuditEventTransactionPosted      = "TRANSACTION_POSTED"
	AuditEventTransactionReversed    = "TRANSACTION_REVERSED"
	AuditEventReconciliationStart    = "RECONCILIATION_STARTED"
	AuditEventReconciliationComplete = "RECONCILIATION_COMPLETED"
	AuditEventPeriodClosed           = "PERIOD_CLOSED"
	AuditEventPeriodReopened         = "PERIOD_REOPENED"
)

// HTTP status codes for API responses
const (
	StatusOK                  = 200
	StatusCreated             = 201
	StatusAccepted            = 202
	StatusNoContent           = 204
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusConflict            = 409
	StatusUnprocessableEntity = 422
	StatusInternalServerError = 500
)

// Context keys for request context
const (
	ContextKeyUserID    = "user_id"
	ContextKeyTenantID  = "tenant_id"
	ContextKeyEntityID  = "entity_id"
	ContextKeyRequestID = "request_id"
	ContextKeyTraceID   = "trace_id"
)

// Feature flags for optional functionality
const (
	FeatureFlagMultiCurrency        = "MULTI_CURRENCY"
	FeatureFlagAdvancedReporting    = "ADVANCED_REPORTING"
	FeatureFlagBudgetManagement     = "BUDGET_MANAGEMENT"
	FeatureFlagRecurringTrans       = "RECURRING_TRANSACTIONS"
	FeatureFlagApprovalWorkflow     = "APPROVAL_WORKFLOW"
	FeatureFlagBankReconciliation   = "BANK_RECONCILIATION"
	FeatureFlagProjectAccounting    = "PROJECT_ACCOUNTING"
	FeatureFlagCostCenterAccounting = "COST_CENTER_ACCOUNTING"
	FeatureFlagMultiEntity          = "MULTI_ENTITY"
	FeatureFlagAuditTrail           = "AUDIT_TRAIL"
)

// Notification types for system events
const (
	NotificationTypeTransactionApprovalRequired = "TRANSACTION_APPROVAL_REQUIRED"
	NotificationTypeTransactionApproved         = "TRANSACTION_APPROVED"
	NotificationTypeTransactionRejected         = "TRANSACTION_REJECTED"
	NotificationTypePeriodClosing               = "PERIOD_CLOSING"
	NotificationTypeBudgetVarianceAlert         = "BUDGET_VARIANCE_ALERT"
	NotificationTypeReconciliationRequired      = "RECONCILIATION_REQUIRED"
	NotificationTypeSystemMaintenanceScheduled  = "SYSTEM_MAINTENANCE_SCHEDULED"
)

// Report types available in the system
const (
	ReportTypeTrialBalance       = "TRIAL_BALANCE"
	ReportTypeIncomeStatement    = "INCOME_STATEMENT"
	ReportTypeBalanceSheet       = "BALANCE_SHEET"
	ReportTypeCashFlowStatement  = "CASH_FLOW_STATEMENT"
	ReportTypeGeneralLedger      = "GENERAL_LEDGER"
	ReportTypeAccountStatement   = "ACCOUNT_STATEMENT"
	ReportTypeBudgetVariance     = "BUDGET_VARIANCE"
	ReportTypeAccountsReceivable = "ACCOUNTS_RECEIVABLE"
	ReportTypeAccountsPayable    = "ACCOUNTS_PAYABLE"
	ReportTypeInventoryValuation = "INVENTORY_VALUATION"
	ReportTypeFixedAssetSchedule = "FIXED_ASSET_SCHEDULE"
)

// Export formats supported by the system
const (
	ExportFormatPDF   = "PDF"
	ExportFormatExcel = "EXCEL"
	ExportFormatCSV   = "CSV"
	ExportFormatJSON  = "JSON"
	ExportFormatXML   = "XML"
)

// Currency settings and defaults
var (
	// Most common currencies
	CommonCurrencies = []string{
		"USD", "EUR", "GBP", "JPY", "CHF",
		"CAD", "AUD", "CNY", "INR", "BRL",
		"KES", "UGX", "TZS", "RWF", "ZAR", // East African currencies
	}

	// Default currency display formats
	CurrencyDisplayFormats = map[string]string{
		"USD": "$%.4f",
		"EUR": "€%.4f",
		"GBP": "£%.4f",
		"JPY": "¥%.0f",
		"KES": "KSh %.4f",
		"UGX": "UGX %.2f",
	}

	// Currencies that don't use decimal places
	ZeroDecimalCurrencies = map[string]bool{
		"JPY": true, // Japanese Yen
		"KRW": true, // Korean Won
		"VND": true, // Vietnamese Dong
		"CLP": true, // Chilean Peso
		"ISK": true, // Icelandic Króna
	}
)

// // Tax-related constants
// const (
// 	// Standard tax types
// 	TaxTypeVAT         = "VAT"
// 	TaxTypeSalesTax    = "SALES_TAX"
// 	TaxTypeWithholding = "WITHHOLDING"
// 	TaxTypeExcise      = "EXCISE"
// 	TaxTypeCustoms     = "CUSTOMS"
//
// 	// Tax calculation methods
// 	TaxCalculationInclusive = "INCLUSIVE"
// 	TaxCalculationExclusive = "EXCLUSIVE"
// 	TaxCalculationCompound  = "COMPOUND"
//
// 	// Default tax rates (can be overridden per jurisdiction)
// 	DefaultVATRate         = 16.0 // Kenya's standard VAT rate
// 	DefaultWithholdingRate = 5.0  // Common withholding tax rate
// )
//
// // Account numbering schemes
// const (
// 	NumberingSchemeManual    = "MANUAL"
// 	NumberingSchemeAutomatic = "AUTOMATIC"
// 	NumberingSchemeTemplate  = "TEMPLATE"
// )

// Account code patterns for different numbering schemes
var (
	// Standard chart of accounts structure
	AccountCodeRanges = map[RootType]struct {
		Start int
		End   int
	}{
		RootTypeAsset:     {Start: 1000, End: 1999},
		RootTypeLiability: {Start: 2000, End: 2999},
		RootTypeEquity:    {Start: 3000, End: 3999},
		RootTypeRevenue:   {Start: 4000, End: 4999},
		RootTypeExpense:   {Start: 5000, End: 9999},
	}
)

// Workflow states for different processes
const (
	// Account workflow states
	WorkflowStateAccountDraft    = "DRAFT"
	WorkflowStateAccountActive   = "ACTIVE"
	WorkflowStateAccountInactive = "INACTIVE"
	WorkflowStateAccountArchived = "ARCHIVED"

	// Transaction workflow states
	WorkflowStateTransactionDraft     = "DRAFT"
	WorkflowStateTransactionSubmitted = "SUBMITTED"
	WorkflowStateTransactionApproved  = "APPROVED"
	WorkflowStateTransactionPosted    = "POSTED"
	WorkflowStateTransactionCancelled = "CANCELLED"
	WorkflowStateTransactionReversed  = "REVERSED"

	// Reconciliation workflow states
	WorkflowStateReconciliationNew        = "NEW"
	WorkflowStateReconciliationInProgress = "IN_PROGRESS"
	WorkflowStateReconciliationCompleted  = "COMPLETED"
	WorkflowStateReconciliationLocked     = "LOCKED"
)

// Integration-related constants
const (
	// External system types
	IntegrationTypeBankFeed  = "BANK_FEED"
	IntegrationTypeERP       = "ERP"
	IntegrationTypeCRM       = "CRM"
	IntegrationTypePayroll   = "PAYROLL"
	IntegrationTypeInventory = "INVENTORY"
	IntegrationTypeTax       = "TAX"

	// Integration status
	IntegrationStatusActive   = "ACTIVE"
	IntegrationStatusInactive = "INACTIVE"
	IntegrationStatusError    = "ERROR"
	IntegrationStatusSyncing  = "SYNCING"

	// Sync frequencies
	SyncFrequencyRealTime = "REAL_TIME"
	SyncFrequencyHourly   = "HOURLY"
	SyncFrequencyDaily    = "DAILY"
	SyncFrequencyWeekly   = "WEEKLY"
	SyncFrequencyManual   = "MANUAL"
)

// Performance and optimization settings
const (
	// Batch processing sizes
	SmallBatchSize  = 100
	MediumBatchSize = 500
	LargeBatchSize  = 1000

	// Query limits
	DefaultQueryLimit = 1000
	MaxQueryLimit     = 10000

	// Cache TTL settings
	ShortCacheTTL  = 5 * time.Minute
	MediumCacheTTL = 30 * time.Minute
	LongCacheTTL   = 2 * time.Hour

	// Connection pool settings
	DefaultMaxOpenConns    = 25
	DefaultMaxIdleConns    = 5
	DefaultConnMaxLifetime = 5 * time.Minute
)

// Security and compliance constants
const (
	// Password requirements
	MinPasswordLength = 8
	MaxPasswordLength = 128

	// Session settings
	SessionTimeout         = 30 * time.Minute
	MaxSessionsPerUser     = 5
	SessionCleanupInterval = 1 * time.Hour

	// Audit retention periods
	AuditRetentionPeriod  = 7 * 365 * 24 * time.Hour // 7 years
	BackupRetentionPeriod = 90 * 24 * time.Hour      // 90 days
	LogRetentionPeriod    = 30 * 24 * time.Hour      // 30 days

	// Encryption settings
	DefaultEncryptionAlgorithm = "AES-256-GCM"
	DefaultHashAlgorithm       = "SHA-256"

	// Rate limiting windows
	RateLimitWindowMinute = 1 * time.Minute
	RateLimitWindowHour   = 1 * time.Hour
	RateLimitWindowDay    = 24 * time.Hour
)

// Localization and internationalization
const (
	// Default locale settings
	DefaultLocale     = "en-US"
	DefaultTimezone   = "UTC"
	DefaultDateFormat = "2006-01-02"
	DefaultTimeFormat = "15:04:05"

	// Number formatting
	DefaultNumberGroupSeparator = ","
	DefaultDecimalSeparator     = "."
	DefaultNegativeNumberFormat = "-%s"
)

// Message queue and event settings
const (
	// Queue names
	QueueNameTransactionProcessing = "transaction_processing"
	QueueNameReconciliation        = "reconciliation"
	QueueNameReporting             = "reporting"
	QueueNameNotifications         = "notifications"
	QueueNameAudit                 = "audit"

	// Event types
	EventTypeEntityCreated = "ENTITY_CREATED"
	EventTypeEntityUpdated = "ENTITY_UPDATED"
	EventTypeEntityDeleted = "ENTITY_DELETED"

	// Message priorities
	MessagePriorityLow    = 1
	MessagePriorityNormal = 5
	MessagePriorityHigh   = 10

	// Retry settings
	MaxRetryAttempts = 3
	RetryBackoffBase = 2 * time.Second
)

// Health check and monitoring
const (
	// Health check endpoints
	HealthCheckPath    = "/health"
	ReadinessCheckPath = "/ready"
	LivenessCheckPath  = "/live"

	// Metric names
	MetricTransactionsProcessed = "transactions_processed_total"
	MetricAccountsCreated       = "accounts_created_total"
	MetricValidationErrors      = "validation_errors_total"
	MetricAPIRequests           = "api_requests_total"
	MetricDatabaseConnections   = "database_connections"

	// Alert thresholds
	HighErrorRateThreshold    = 0.05 // 5%
	HighLatencyThreshold      = 2000 // 2 seconds in milliseconds
	DatabaseConnectionWarning = 20   // 80% of max connections
)

// Development and testing constants
const (
	// Test data identifiers
	TestTenantIDString      = "00000000-0000-0000-0000-000000000001"
	TestUserIDString        = "00000000-0000-0000-0000-000000000002"
	TestEntityIDString      = "00000000-0000-0000-0000-000000000003"
	TestAccountIDString     = "00000000-0000-0000-0000-000000000004"
	TestTransactionIDString = "00000000-0000-0000-0000-000000000005"

	// Environment names
	EnvironmentDevelopment = "development"
	EnvironmentTesting     = "testing"
	EnvironmentStaging     = "staging"
	EnvironmentProduction  = "production"

	// Feature flags for testing
	TestFeatureFlagPrefix = "TEST_"
)

// GetDefaultAccountType returns the default account type for a given root type
func GetDefaultAccountType(rootType RootType) string {
	switch rootType {
	case RootTypeAsset:
		return AccountTypeCurrentAsset
	case RootTypeLiability:
		return AccountTypeCurrentLiability
	case RootTypeEquity:
		return AccountTypeOtherEquity
	case RootTypeRevenue:
		return AccountTypeOperatingRevenue
	case RootTypeExpense:
		return AccountTypeOperatingExpense
	default:
		return ""
	}
}

// GetAccountCodeRange returns the account code range for a root type
func GetAccountCodeRange(rootType RootType) (int, int, bool) {
	if rng, exists := AccountCodeRanges[rootType]; exists {
		return rng.Start, rng.End, true
	}
	return 0, 0, false
}

// IsSystemAccount determines if an account code represents a system account
func IsSystemAccount(accountCode string) bool {
	systemCodes := []string{
		SystemAccountCodeCash,
		SystemAccountCodeAccountsReceivable,
		SystemAccountCodeInventory,
		SystemAccountCodeAccountsPayable,
		SystemAccountCodeRetainedEarnings,
		SystemAccountCodeSalesRevenue,
		SystemAccountCodeCostOfGoodsSold,
	}

	for _, code := range systemCodes {
		if accountCode == code {
			return true
		}
	}
	return false
}

// GetCurrencyDisplayFormat returns the display format for a currency
func GetCurrencyDisplayFormat(currencyCode string) string {
	if format, exists := CurrencyDisplayFormats[currencyCode]; exists {
		return format
	}
	return "%.2f %s" // Default format
}

// IsZeroDecimalCurrency checks if a currency doesn't use decimal places
func IsZeroDecimalCurrency(currencyCode string) bool {
	return ZeroDecimalCurrencies[currencyCode]
}

// GetDecimalPlacesForCurrency returns the appropriate decimal places for a currency
func GetDecimalPlacesForCurrency(currencyCode string) int32 {
	if IsZeroDecimalCurrency(currencyCode) {
		return 0
	}
	return int32(DisplayDecimalPrecision)
}

// Helper functions for working with constants

// IsValidRootType checks if the provided string is a valid root type
func IsValidRootType(rootType string) bool {
	validTypes := []RootType{
		RootTypeAsset, RootTypeLiability, RootTypeEquity, RootTypeRevenue, RootTypeExpense,
	}

	for _, validType := range validTypes {
		if string(validType) == rootType {
			return true
		}
	}
	return false
}

// IsValidTransactionType checks if the provided string is a valid transaction type
func IsValidTransactionType(transactionType string) bool {
	validTypes := []TransactionType{
		TransactionTypeManual, TransactionTypeSystem, TransactionTypeImported,
		TransactionTypeRecurring, TransactionTypeAdjustment, TransactionTypeClosing,
	}

	for _, validType := range validTypes {
		if string(validType) == transactionType {
			return true
		}
	}
	return false
}

// IsValidCurrency checks if the provided currency code is supported
func IsValidCurrency(currencyCode string) bool {
	for _, currency := range CommonCurrencies {
		if currency == currencyCode {
			return true
		}
	}
	return false
}

// GetEnvironmentDefaults returns environment-specific default values
func GetEnvironmentDefaults(environment string) map[string]any {
	defaults := make(map[string]any)

	switch environment {
	case EnvironmentDevelopment:
		defaults["log_level"] = "debug"
		defaults["rate_limit"] = 10000
		defaults["cache_ttl"] = 1 * time.Minute
		defaults["enable_debug"] = true
	case EnvironmentTesting:
		defaults["log_level"] = "info"
		defaults["rate_limit"] = 5000
		defaults["cache_ttl"] = 30 * time.Second
		defaults["enable_debug"] = true
	case EnvironmentStaging:
		defaults["log_level"] = "warn"
		defaults["rate_limit"] = 2000
		defaults["cache_ttl"] = 5 * time.Minute
		defaults["enable_debug"] = false
	case EnvironmentProduction:
		defaults["log_level"] = "error"
		defaults["rate_limit"] = DefaultRateLimit
		defaults["cache_ttl"] = DefaultCacheExpiry
		defaults["enable_debug"] = false
	default:
		// Return production defaults for unknown environments
		return GetEnvironmentDefaults(EnvironmentProduction)
	}

	return defaults
}

// TODO: Add support for custom account types and subtypes per tenant
// TODO: Implement dynamic currency support with external rate providers
// TODO: Add support for multiple fiscal year calendars
// TODO: Implement configurable validation rules and limits
// TODO: Add support for custom report types and formats
// NOTE: Consider implementing feature flag management system
// NOTE: Future enhancement: Add support for multi-language constants
// NOTE: Consider implementing tenant-specific constant overrides
// NOTE: Add support for regulatory compliance constants per jurisdiction
