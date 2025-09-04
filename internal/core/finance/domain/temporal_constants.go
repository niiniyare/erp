package domain

import (
	"fmt"
	"time"
)

// Temporal-specific constants for finance module workflows and activities
// These constants complement the existing constant.go file with Temporal orchestration configurations

// ========================================
// TEMPORAL WORKFLOW CONFIGURATION
// ========================================

// Temporal task queue names - organize workflows by domain and priority
const (
	// Primary finance task queue for standard operations
	TemporalTaskQueueFinance = "finance"
	
	// High priority queue for time-sensitive operations (period closing, urgent approvals)
	TemporalTaskQueueFinanceHighPriority = "finance-high-priority"
	
	// Bulk operations queue for heavy processing (data imports, bulk posting)
	TemporalTaskQueueFinanceBulk = "finance-bulk"
	
	// Long-running processes queue (month-end closing, year-end processing)
	TemporalTaskQueueFinanceLongRunning = "finance-long-running"
)

// Temporal workflow types - identify different business processes
const (
	// Account management workflows
	WorkflowTypeAccountCreation     = "finance.account.creation"
	WorkflowTypeAccountActivation   = "finance.account.activation"
	WorkflowTypeAccountDeactivation = "finance.account.deactivation"
	WorkflowTypeBulkAccountUpdate   = "finance.account.bulk_update"
	
	// Transaction processing workflows
	WorkflowTypeTransactionProcessing = "finance.transaction.processing"
	WorkflowTypeTransactionApproval   = "finance.transaction.approval"
	WorkflowTypeTransactionPosting    = "finance.transaction.posting"
	WorkflowTypeTransactionReversal   = "finance.transaction.reversal"
	WorkflowTypeBulkTransactionImport = "finance.transaction.bulk_import"
	
	// Financial period workflows
	WorkflowTypePeriodClosing  = "finance.period.closing"
	WorkflowTypePeriodOpening  = "finance.period.opening"
	WorkflowTypeYearEndClosing = "finance.period.year_end_closing"
	
	// Reconciliation workflows
	WorkflowTypeBankReconciliation    = "finance.reconciliation.bank"
	WorkflowTypeAccountReconciliation = "finance.reconciliation.account"
	
	// Reporting workflows
	WorkflowTypeFinancialReporting = "finance.reporting.generation"
	WorkflowTypeBulkReportExport   = "finance.reporting.bulk_export"
)

// Temporal activity types - map to specific business operations
const (
	// Account management activities
	ActivityTypeAccountValidation    = "finance.activity.account.validation"
	ActivityTypeAccountCreation      = "finance.activity.account.creation"
	ActivityTypeAccountUpdate        = "finance.activity.account.update"
	ActivityTypeAccountStatusChange  = "finance.activity.account.status_change"
	ActivityTypeAccountBalanceUpdate = "finance.activity.account.balance_update"
	
	// Transaction processing activities
	ActivityTypeTransactionValidation = "finance.activity.transaction.validation"
	ActivityTypeTransactionCreation   = "finance.activity.transaction.creation"
	ActivityTypeTransactionPosting    = "finance.activity.transaction.posting"
	ActivityTypeTransactionReversal   = "finance.activity.transaction.reversal"
	ActivityTypeEntryGeneration       = "finance.activity.entry.generation"
	
	// External integration activities
	ActivityTypeAuditLogging      = "finance.activity.audit.logging"
	ActivityTypeNotification      = "finance.activity.notification.send"
	ActivityTypeFeatureFlagCheck  = "finance.activity.feature_flag.check"
	ActivityTypeSettingsRetrieval = "finance.activity.settings.retrieval"
	ActivityTypeIAMValidation     = "finance.activity.iam.validation"
	
	// Financial process activities
	ActivityTypePeriodValidation    = "finance.activity.period.validation"
	ActivityTypeBalanceCalculation  = "finance.activity.balance.calculation"
	ActivityTypeReportGeneration    = "finance.activity.report.generation"
	ActivityTypeReconciliationMatch = "finance.activity.reconciliation.match"
)

// ========================================
// TEMPORAL TIMEOUT CONFIGURATION
// ========================================

// Workflow execution timeouts - control maximum workflow execution time
const (
	// Standard operation timeouts
	DefaultWorkflowTimeout        = 30 * time.Minute  // Most operations should complete within 30 minutes
	AccountWorkflowTimeout        = 10 * time.Minute  // Account operations are typically fast
	TransactionWorkflowTimeout    = 15 * time.Minute  // Transaction processing with validation
	
	// Long-running process timeouts  
	BulkOperationWorkflowTimeout  = 2 * time.Hour     // Bulk imports and processing
	PeriodClosingWorkflowTimeout  = 4 * time.Hour     // Month-end closing processes
	YearEndClosingWorkflowTimeout = 8 * time.Hour     // Year-end closing with reports
	ReportingWorkflowTimeout      = 1 * time.Hour     // Financial report generation
	
	// Critical business process timeouts
	ReconciliationWorkflowTimeout = 45 * time.Minute  // Bank reconciliation processes
	ApprovalWorkflowTimeout       = 24 * time.Hour    // Waiting for human approval
)

// Activity execution timeouts - control individual activity execution time
const (
	// Database operation timeouts
	StandardActivityTimeout   = 30 * time.Second   // Standard CRUD operations
	ValidationActivityTimeout = 15 * time.Second   // Business rule validation
	CalculationActivityTimeout = 45 * time.Second  // Balance calculations
	
	// External service timeouts
	NotificationActivityTimeout   = 10 * time.Second   // Send notifications
	AuditActivityTimeout         = 5 * time.Second    // Audit logging
	FeatureFlagActivityTimeout   = 3 * time.Second    // Feature flag checks
	SettingsActivityTimeout      = 5 * time.Second    // Settings retrieval
	IAMActivityTimeout           = 10 * time.Second   // IAM validation calls
	
	// Heavy processing timeouts
	BulkProcessingActivityTimeout = 10 * time.Minute  // Bulk data processing
	ReportGenerationActivityTimeout = 5 * time.Minute // Individual report generation
)

// Retry policy configurations for different operation types
const (
	// Standard retry attempts for transient failures
	DefaultRetryAttempts    = 3
	ValidationRetryAttempts = 1  // Don't retry validation failures
	NetworkRetryAttempts    = 5  // Retry network calls more aggressively
	
	// Retry backoff intervals
	StandardRetryBackoff = 2 * time.Second
	NetworkRetryBackoff  = 1 * time.Second
	BulkRetryBackoff     = 5 * time.Second
)

// ========================================
// MULTI-TENANT CACHE KEY PATTERNS
// ========================================

// Cache key templates following multi-tenant pattern: "tenant:{tenant_id}:entity:{entity_id}:finance:..."
// These keys ensure proper tenant and entity isolation for company data separation
const (
	// Account-related cache keys with tenant and entity isolation
	CacheKeyAccountByID       = "tenant:%s:entity:%s:finance:account:id:%s"           // tenant_id, entity_id, account_id
	CacheKeyAccountByCode     = "tenant:%s:entity:%s:finance:account:code:%s"         // tenant_id, entity_id, account_code
	CacheKeyAccountBalance    = "tenant:%s:entity:%s:finance:account:balance:%s"      // tenant_id, entity_id, account_id
	CacheKeyAccountList       = "tenant:%s:entity:%s:finance:account:list:%s"         // tenant_id, entity_id, filter_hash
	CacheKeyAccountHierarchy  = "tenant:%s:entity:%s:finance:account:hierarchy"       // tenant_id, entity_id
	
	// Transaction-related cache keys with multi-company isolation
	CacheKeyTransactionByID    = "tenant:%s:entity:%s:finance:transaction:id:%s"      // tenant_id, entity_id, transaction_id  
	CacheKeyTransactionList    = "tenant:%s:entity:%s:finance:transaction:list:%s"    // tenant_id, entity_id, filter_hash
	CacheKeyTransactionEntries = "tenant:%s:entity:%s:finance:transaction:entries:%s" // tenant_id, entity_id, transaction_id
	CacheKeyTransactionTotals  = "tenant:%s:entity:%s:finance:transaction:totals:%s"  // tenant_id, entity_id, date_range_hash
	
	// Exchange rate cache keys (may be shared across entities within tenant)
	CacheKeyExchangeRate     = "tenant:%s:finance:exchange_rate:%s:%s"        // tenant_id, from_currency, to_currency
	CacheKeyExchangeRateList = "tenant:%s:finance:exchange_rate:list:%s"      // tenant_id, date
	
	// Financial reporting cache keys with entity-specific data
	CacheKeyReportTrialBalance  = "tenant:%s:entity:%s:finance:report:trial_balance:%s"   // tenant_id, entity_id, params_hash
	CacheKeyReportIncomeStmt    = "tenant:%s:entity:%s:finance:report:income_stmt:%s"     // tenant_id, entity_id, params_hash
	CacheKeyReportBalanceSheet  = "tenant:%s:entity:%s:finance:report:balance_sheet:%s"   // tenant_id, entity_id, params_hash
	CacheKeyReportCashFlow     = "tenant:%s:entity:%s:finance:report:cash_flow:%s"       // tenant_id, entity_id, params_hash
	
	// Period and closing cache keys for entity-specific financial periods
	CacheKeyFinancialPeriod    = "tenant:%s:entity:%s:finance:period:%s"           // tenant_id, entity_id, period_id
	CacheKeyPeriodStatus       = "tenant:%s:entity:%s:finance:period:status:%s"    // tenant_id, entity_id, year_month
	CacheKeyClosingStatus      = "tenant:%s:entity:%s:finance:closing:status:%s"   // tenant_id, entity_id, period_id
	
	// Settings cache keys for finance configuration per entity
	CacheKeyFinanceSettings    = "tenant:%s:entity:%s:finance:settings"                // tenant_id, entity_id
	CacheKeyChartOfAccounts    = "tenant:%s:entity:%s:finance:chart_of_accounts"       // tenant_id, entity_id
	CacheKeyAccountingPolicies = "tenant:%s:entity:%s:finance:accounting_policies"     // tenant_id, entity_id
	
	// Workflow state cache keys for tracking process status
	CacheKeyWorkflowState      = "tenant:%s:entity:%s:finance:workflow:%s:%s"      // tenant_id, entity_id, workflow_type, workflow_id
	CacheKeyPendingApprovals   = "tenant:%s:entity:%s:finance:approvals:pending"   // tenant_id, entity_id
	CacheKeyActiveWorkflows    = "tenant:%s:entity:%s:finance:workflows:active"    // tenant_id, entity_id
)

// Cache TTL (Time To Live) constants for different data types
// Balances business need for performance vs data freshness
const (
	// Account data caching - relatively stable, can cache longer
	AccountDataCacheTTL        = 30 * time.Minute   // Account details change infrequently
	AccountBalanceCacheTTL     = 5 * time.Minute    // Balances change more frequently  
	AccountHierarchyCacheTTL   = 1 * time.Hour      // Chart of accounts changes rarely
	
	// Transaction data caching - more dynamic, shorter TTL
	TransactionDataCacheTTL    = 15 * time.Minute   // Transaction details
	TransactionListCacheTTL    = 5 * time.Minute    // Transaction lists change often
	TransactionTotalsCacheTTL  = 10 * time.Minute   // Calculated totals
	
	// Exchange rate caching - external data, moderate TTL
	ExchangeRateCacheTTL       = 1 * time.Hour      // Exchange rates from external sources
	
	// Report caching - expensive to generate, longer TTL
	FinancialReportCacheTTL    = 2 * time.Hour      // Generated financial reports
	ReportDataCacheTTL         = 30 * time.Minute   // Report data queries
	
	// Configuration caching - rarely changes, long TTL
	FinanceSettingsCacheTTL    = 4 * time.Hour      // Finance module settings
	AccountingPoliciesCacheTTL = 6 * time.Hour      // Accounting policies and rules
	
	// Workflow state caching - dynamic during processing
	WorkflowStateCacheTTL      = 2 * time.Minute    // Active workflow states
	PendingApprovalsCacheTTL   = 5 * time.Minute    // Pending approval lists
	
	// Period and closing caching - stable during period, changes at period boundaries  
	FinancialPeriodCacheTTL    = 1 * time.Hour      // Financial period definitions
	PeriodStatusCacheTTL       = 15 * time.Minute   // Period open/closed status
)

// ========================================
// TEMPORAL SIGNAL AND QUERY TYPES
// ========================================

// Temporal signal names for workflow communication
const (
	// Approval process signals
	SignalApprovalGranted  = "approval_granted"
	SignalApprovalRejected = "approval_rejected" 
	SignalApprovalTimeout  = "approval_timeout"
	
	// Process control signals
	SignalProcessCancel   = "process_cancel"
	SignalProcessPause    = "process_pause"
	SignalProcessResume   = "process_resume"
	SignalProcessPriority = "process_priority_change"
	
	// Data update signals
	SignalAccountUpdated     = "account_updated"
	SignalTransactionUpdated = "transaction_updated"
	SignalSettingsUpdated    = "settings_updated"
)

// Temporal query types for workflow status inspection
const (
	QueryWorkflowStatus        = "workflow_status"
	QueryProcessingProgress    = "processing_progress"  
	QueryApprovalStatus        = "approval_status"
	QueryValidationResults     = "validation_results"
	QueryCurrentStep           = "current_step"
	QueryRemainingWork         = "remaining_work"
	QueryErrorDetails          = "error_details"
	QueryPerformanceMetrics    = "performance_metrics"
)

// ========================================
// WORKFLOW STATE AND STATUS CONSTANTS  
// ========================================

// Workflow execution states for status tracking
const (
	WorkflowStatusInitializing = "INITIALIZING"  // Workflow starting up
	WorkflowStatusValidating   = "VALIDATING"    // Input validation phase
	WorkflowStatusProcessing   = "PROCESSING"    // Active processing
	WorkflowStatusWaitingApproval = "WAITING_APPROVAL" // Awaiting human approval
	WorkflowStatusApproved     = "APPROVED"      // Approval granted
	WorkflowStatusRejected     = "REJECTED"      // Approval rejected  
	WorkflowStatusCompleted    = "COMPLETED"     // Successfully completed
	WorkflowStatusFailed       = "FAILED"        // Failed with errors
	WorkflowStatusCancelled    = "CANCELLED"     // Cancelled by user/system
	WorkflowStatusTimedOut     = "TIMED_OUT"     // Exceeded timeout limits
)

// Activity execution states for granular tracking
const (
	ActivityStatusPending    = "PENDING"      // Waiting to start
	ActivityStatusStarted    = "STARTED"      // Currently executing
	ActivityStatusCompleted  = "COMPLETED"    // Successfully completed
	ActivityStatusFailed     = "FAILED"       // Failed with error
	ActivityStatusRetrying   = "RETRYING"     // Retrying after failure
	ActivityStatusSkipped    = "SKIPPED"      // Skipped due to conditions
	ActivityStatusTimedOut   = "TIMED_OUT"    // Exceeded execution timeout
)

// ========================================
// HELPER FUNCTIONS FOR CACHE KEY GENERATION
// ========================================

// Helper functions to generate cache keys with proper tenant/entity isolation
// These functions ensure consistent key formatting across the finance module

// GetAccountCacheKey generates cache key for account data with tenant and entity isolation
func GetAccountCacheKey(tenantSlug, entityID, accountID string) string {
	return fmt.Sprintf(CacheKeyAccountByID, tenantSlug, entityID, accountID)
}

// GetAccountBalanceCacheKey generates cache key for account balance with multi-tenant isolation
func GetAccountBalanceCacheKey(tenantSlug, entityID, accountID string) string {
	return fmt.Sprintf(CacheKeyAccountBalance, tenantSlug, entityID, accountID)
}

// GetTransactionCacheKey generates cache key for transaction data with proper isolation
func GetTransactionCacheKey(tenantSlug, entityID, transactionID string) string {
	return fmt.Sprintf(CacheKeyTransactionByID, tenantSlug, entityID, transactionID)
}

// GetReportCacheKey generates cache key for financial reports with entity-specific data
func GetReportCacheKey(tenantSlug, entityID, reportType, paramsHash string) string {
	switch reportType {
	case ReportTypeTrialBalance:
		return fmt.Sprintf(CacheKeyReportTrialBalance, tenantSlug, entityID, paramsHash)
	case ReportTypeIncomeStatement:
		return fmt.Sprintf(CacheKeyReportIncomeStmt, tenantSlug, entityID, paramsHash)
	case ReportTypeBalanceSheet:
		return fmt.Sprintf(CacheKeyReportBalanceSheet, tenantSlug, entityID, paramsHash)
	case ReportTypeCashFlowStatement:
		return fmt.Sprintf(CacheKeyReportCashFlow, tenantSlug, entityID, paramsHash)
	default:
		return fmt.Sprintf("tenant:%s:entity:%s:finance:report:%s:%s", tenantSlug, entityID, reportType, paramsHash)
	}
}

// GetWorkflowStateCacheKey generates cache key for workflow state tracking
func GetWorkflowStateCacheKey(tenantSlug, entityID, workflowType, workflowID string) string {
	return fmt.Sprintf(CacheKeyWorkflowState, tenantSlug, entityID, workflowType, workflowID)
}

// GetFinanceSettingsCacheKey generates cache key for finance module settings
func GetFinanceSettingsCacheKey(tenantSlug, entityID string) string {
	return fmt.Sprintf(CacheKeyFinanceSettings, tenantSlug, entityID)
}

// ========================================
// INTEGRATION WITH EXISTING CONSTANTS
// ========================================

// These constants complement the existing constant.go file and should be used together:
// - Use constants from constant.go for business rules, account types, error codes, etc.
// - Use constants from this file for Temporal orchestration and caching
// - Both files work together to eliminate hard-coded values throughout the finance module

// Example usage combining both constant files:
//   - Account creation workflow: WorkflowTypeAccountCreation + AccountTypeCurrentAsset + DefaultWorkflowTimeout
//   - Cache account data: GetAccountCacheKey() + AccountDataCacheTTL
//   - Transaction processing: WorkflowTypeTransactionProcessing + WorkflowStateTransactionDraft + TransactionWorkflowTimeout
//   - Error handling: Use error codes from constant.go + Temporal retry policies from this file

// NOTE: This file focuses specifically on Temporal workflows and multi-tenant caching patterns
// NOTE: All cache keys follow the pattern documented in /internal/platform/cache/Readme.md
// NOTE: Workflow and activity types use hierarchical naming for easy filtering and monitoring
// NOTE: Timeout values are conservative to prevent premature workflow termination in production