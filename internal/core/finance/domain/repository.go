package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Repository interfaces define the contracts for data persistence
// These interfaces are part of the domain layer and define what the domain needs
// Implementations are in the infrastructure layer, following the Dependency Inversion Principle

// AccountsRepository defines the contract for chart of accounts persistence
type AccountsRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, account *Accounts) error
	GetByID(ctx context.Context, id uuid.UUID) (*Accounts, error)
	GetByCode(ctx context.Context, entityID *uuid.UUID, accountCode string) (*Accounts, error)
	Update(ctx context.Context, account *Accounts) error
	Delete(ctx context.Context, id uuid.UUID) error

	// List and filtering operations
	List(ctx context.Context, filter *AccountFilter) ([]*Accounts, error)
	Count(ctx context.Context, filter *AccountFilter) (int64, error)
	ListByParent(ctx context.Context, parentID uuid.UUID) ([]*Accounts, error)
	ListByRootType(ctx context.Context, rootType RootType) ([]*Accounts, error)

	// Hierarchy operations
	GetAccountHierarchy(ctx context.Context, rootID uuid.UUID) ([]*Accounts, error)
	GetAccountPath(ctx context.Context, accountID uuid.UUID) ([]Accounts, error)
	ValidateHierarchy(ctx context.Context, accountID, parentID uuid.UUID) error

	// Balance operations
	GetAccountBalance(ctx context.Context, accountID uuid.UUID, asOfDate *time.Time) (*AccountBalance, error)
	GetAccountBalances(ctx context.Context, accountIDs []uuid.UUID, asOfDate *time.Time) ([]*AccountBalance, error)
	GetTrialBalance(ctx context.Context, entityID *uuid.UUID, asOfDate *time.Time) ([]*TrialBalanceEntry, error)

	// Business logic queries
	GetActiveAccounts(ctx context.Context, entityID *uuid.UUID) ([]*Accounts, error)
	GetControlAccounts(ctx context.Context, entityID *uuid.UUID) ([]*Accounts, error)
	GetAccountsByType(ctx context.Context, accountType string, rootType *RootType) ([]*Accounts, error)
	Search(ctx context.Context, query string, limit int) ([]*Accounts, error)

	// Validation helpers
	ValidateAccountCode(ctx context.Context, code string, excludeID *uuid.UUID) error
	IsAccountCodeUnique(ctx context.Context, entityID *uuid.UUID, accountCode string, excludeID *uuid.UUID) (bool, error)
	HasChildren(ctx context.Context, accountID uuid.UUID) (bool, error)
	GetChildren(ctx context.Context, accountID uuid.UUID) ([]*Accounts, error)
	HasTransactions(ctx context.Context, accountID uuid.UUID) (bool, error)
	UpdateBalance(ctx context.Context, accountID uuid.UUID, balance AccountBalance) error

	// FIXME:
	// Enhanced view-based operations
	// GetAccountWithGroups(ctx context.Context, id uuid.UUID) (*AccountWithGroups, error)
	// GetAccountWithGroupsByCode(ctx context.Context, code string) (*AccountWithGroups, error)
	// ListAccountsWithGroups(ctx context.Context, filter *AccountFilter) ([]*AccountWithGroups, error)
	// SearchAccountsWithGroups(ctx context.Context, query string, limit int) ([]*AccountWithGroups, error)
	// GetLeafAccountsOnly(ctx context.Context, rootType *string) ([]*AccountWithGroups, error)

	// Complete chart of accounts operations
	// GetCompleteChartOfAccounts(ctx context.Context, filter *ChartOfAccountsFilter) ([]*ChartOfAccountsComplete, error)
	// // GetAccountForReporting(ctx context.Context, accountID uuid.UUID) (*ChartOfAccountsComplete, error)
	// GetAccountsByStatementSection(ctx context.Context, section string, entityID *uuid.UUID) ([]*ChartOfAccountsComplete, error)
	// GetAccountsByGroup(ctx context.Context, groupCode string, entityID *uuid.UUID) ([]*ChartOfAccountsComplete, error)
	// GetAccountsByHeader(ctx context.Context, headerCode string, entityID *uuid.UUID) ([]*ChartOfAccountsComplete, error)
	//
	// // Financial reporting operations
	// GetTrialBalanceAccounts(ctx context.Context, entityID *uuid.UUID, nonZeroOnly bool) ([]*TrialBalanceSummary, error)
	// GetAccountsWithBalances(ctx context.Context, filter *BalanceFilter) ([]*AccountGroupRepository, error)
	// GetCashFlowAccounts(ctx context.Context, entityID *uuid.UUID) ([]*CashFlowAccount, error)
	// GetAccountSummaryByGroup(ctx context.Context, entityID *uuid.UUID) ([]*AccountGroup, error)

	// Account Group operations (unified in AccountsRepository)
	// Basic CRUD operations for account groups
	CreateAccountGroup(ctx context.Context, group *AccountGroup) error
	GetAccountGroupByID(ctx context.Context, id uuid.UUID) (*AccountGroup, error)
	GetAccountGroupByCode(ctx context.Context, code string, entityID *uuid.UUID) (*AccountGroup, error)
	UpdateAccountGroup(ctx context.Context, id uuid.UUID, group *AccountGroup) error
	DeleteAccountGroup(ctx context.Context, id uuid.UUID, entityID *uuid.UUID) error

	// List and filtering operations for account groups
	ListAccountGroups(ctx context.Context, filter *AccountGroupFilter) ([]*AccountGroup, error)
	CountAccountGroups(ctx context.Context, filter *AccountGroupFilter) (int64, error)

	// Hierarchy operations for account groups
	GetAccountGroupHierarchy(ctx context.Context, rootGroupID *uuid.UUID, entityID *uuid.UUID) ([]*AccountGroup, error)

	// Financial statement operations for account groups
	GetGroupsByFinancialStatement(ctx context.Context, statementType string, entityID *uuid.UUID) ([]*AccountGroup, error)
	GetGroupsByCashFlowCategory(ctx context.Context, category string, entityID *uuid.UUID) ([]*AccountGroup, error)

	// Validation helpers for account groups
	ValidateAccountGroupCode(ctx context.Context, code string, excludeID *uuid.UUID, entityID *uuid.UUID) error

	// Enhanced analytics and hierarchy operations
	GetAccountChildrenHierarchy(ctx context.Context, parentAccountID uuid.UUID) ([]*AccountHierarchy, error)
	GetAccountSubtree(ctx context.Context, accountID uuid.UUID) ([]*AccountHierarchy, error)
	GetAccountsWithRecentActivity(ctx context.Context, filter *AccountActivityFilter) ([]*AccountActivity, error)
	GetStaleAccountBalances(ctx context.Context, filter *AccountActivityFilter) ([]*AccountActivity, error)
	GetAccountActivitySummary(ctx context.Context, filter *AccountActivityFilter) ([]*AccountActivitySummary, error)
}

// TransactionRepository defines the contract for transaction persistence
type TransactionRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, transaction *Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	GetByNumber(ctx context.Context, entityID *uuid.UUID, transactionNumber string) (*Transaction, error)
	Update(ctx context.Context, transaction *Transaction) error
	Delete(ctx context.Context, id uuid.UUID) error

	// List and filtering operations
	List(ctx context.Context, filter *TransactionFilter) ([]*Transaction, error)
	Count(ctx context.Context, filter *TransactionFilter) (int64, error)
	ListByAccount(ctx context.Context, accountID uuid.UUID, filter *TransactionFilter) ([]*Transaction, error)
	ListByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*Transaction, error)

	// Status and workflow operations
	GetByStatus(ctx context.Context, status TransactionStatus, limit int) ([]*Transaction, error)
	GetPendingApproval(ctx context.Context, userID *uuid.UUID) ([]*Transaction, error)
	GetRecurringTransactions(ctx context.Context, dueDate time.Time) ([]*Transaction, error)
	UpdateNextRecurringDate(ctx context.Context, transactionID uuid.UUID, nextDate time.Time) error

	// Transaction entries
	CreateEntry(ctx context.Context, entry *TransactionEntry) error
	CreateEntries(ctx context.Context, entries []*TransactionEntry) error
	GetEntryByID(ctx context.Context, id uuid.UUID) (*TransactionEntry, error)
	GetEntriesByTransaction(ctx context.Context, transactionID uuid.UUID) ([]TransactionEntry, error)
	GetEntriesByAccount(ctx context.Context, accountID uuid.UUID, filter *EntryFilter) ([]TransactionEntry, error)
	UpdateEntry(ctx context.Context, entry *TransactionEntry) error
	DeleteEntry(ctx context.Context, id uuid.UUID) error
	SearchEntries(ctx context.Context, query string, filters *EntryFilter, limit int, offset int) ([]*TransactionEntry, error)
	UpdateReconciliationStatus(ctx context.Context, entryID uuid.UUID, reconciled bool, reconciledDate *time.Time, reconciliationRef *string) error
	GetUnreconciledEntries(ctx context.Context, accountID uuid.UUID, cutoffDate *time.Time) ([]*TransactionEntry, error)
	GetEntrySummary(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) (*TransactionSummary, error)

	// Balance operations
	CalculateAccountBalance(ctx context.Context, accountID uuid.UUID, asOfDate *time.Time) (decimal.Decimal, error)
	GetAccountTransactionSummary(ctx context.Context, accountID uuid.UUID, dateRange *DateRange) (*TransactionSummary, error)

	// Posting and approval operations
	Post(ctx context.Context, transactionID uuid.UUID, postedBy uuid.UUID, postedAt time.Time) error
	Approve(ctx context.Context, transactionID uuid.UUID, approvedBy uuid.UUID, approvedAt time.Time, notes *string) error
	Reject(ctx context.Context, transactionID uuid.UUID, rejectedBy uuid.UUID, rejectedAt time.Time, reason RejectionReason, notes *string) error
	Reverse(ctx context.Context, originalID, reversalID uuid.UUID, reason string) error

	// Validation helpers
	IsTransactionNumberUnique(ctx context.Context, entityID *uuid.UUID, transactionNumber string, excludeID *uuid.UUID) (bool, error)
	ValidateAccountsExist(ctx context.Context, accountIDs []uuid.UUID) error
	GetNextTransactionNumber(ctx context.Context, entityID *uuid.UUID, transactionType TransactionType) (string, error)
}

// TxRunner wraps a unit of work in a database transaction.
// The supplied fn receives a child context that carries the transaction;
// all repository calls made with that context will participate in the same tx.
// Commit is automatic on nil return; rollback on any error.
//
// Implementations live in the infrastructure layer (e.g. pgx pool).
// This interface is defined here so the domain/service layer can depend on it
// without importing driver packages.
type TxRunner interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// ReversalHistoryRecord is the persisted record of a single reversal event.
type ReversalHistoryRecord struct {
	ID                    uuid.UUID `json:"id"`
	OriginalTransactionID uuid.UUID `json:"original_transaction_id"`
	ReversalTransactionID uuid.UUID `json:"reversal_transaction_id"`
	Reason                string    `json:"reason"`
	ReversedBy            uuid.UUID `json:"reversed_by"`
	CreatedAt             time.Time `json:"created_at"`
}

// ReversalHistoryRepository persists and queries the finance_reversal_history table.
type ReversalHistoryRepository interface {
	// Insert persists a reversal event immediately after the reversal transaction is saved.
	Insert(ctx context.Context, rec *ReversalHistoryRecord) error

	// IsReversal returns true if transactionID appears as a reversal_transaction_id
	// in the history table — i.e., the transaction is itself a reversal of another.
	IsReversal(ctx context.Context, transactionID uuid.UUID) (bool, error)

	// GetByOriginal returns all reversal records for a given original transaction.
	GetByOriginal(ctx context.Context, originalTransactionID uuid.UUID) ([]*ReversalHistoryRecord, error)
}

// AuditRepository defines the contract for audit trail persistence
type AuditRepository interface {
	CreateAuditEntry(ctx context.Context, entry *AuditEntry) error
	GetAuditTrail(ctx context.Context, entityType, entityID string, filter *AuditFilter) ([]*AuditEntry, error)
	GetUserActivity(ctx context.Context, userID uuid.UUID, filter *AuditFilter) ([]*AuditEntry, error)
	PurgeOldEntries(ctx context.Context, olderThan time.Time) error
}

// Filter and parameter structures

// AccountFilter defines filtering options for chart of accounts queries.
// All pointer fields are optional; nil means "no filter on this field".
//
//	type AccountFilter struct {
//		TenantID    uuid.UUID  `json:"tenant_id"`
//		EntityID    *uuid.UUID `json:"entity_id,omitempty"`
//		RootType    *RootType  `json:"root_type,omitempty"`
//		AccountType *string    `json:"account_type,omitempty"` // e.g. "BANK", "CASH", "PAYABLE"
//		ParentID    *uuid.UUID `json:"parent_id,omitempty"`
//		ParentCode  *string    `json:"parent_code,omitempty"` // alternative to ParentID
//		IsActive    *bool      `json:"is_active,omitempty"`
//		IsGroup     *bool      `json:"is_group,omitempty"` // true = groups only, false = leaf only
//		// Query is a full-text search term matched against account code and name.
//		// SearchTerm is kept for backward compatibility and is equivalent to Query.
//		Query      string  `json:"query,omitempty"`
//		SearchTerm *string `json:"search_term,omitempty"`
//
//		// Pagination — Page/PerPage are preferred; Limit/Offset are kept for backward compat.
//		Page    int  `json:"page,omitempty"`
//		PerPage int  `json:"per_page,omitempty"`
//		Limit   *int `json:"limit,omitempty"`
//		Offset  *int `json:"offset,omitempty"`
//
//		// Sorting
//		SortBy    *string `json:"sort_by,omitempty"`
//		SortOrder *string `json:"sort_order,omitempty"` // ASC or DESC
//	}
//
// TransactionFilter defines filtering options for transaction queries
type TransactionFilter struct {
	EntityID        *uuid.UUID         `json:"entity_id,omitempty"`
	TransactionType *TransactionType   `json:"transaction_type,omitempty"`
	Status          *TransactionStatus `json:"status,omitempty"`
	ApprovalStatus  *ApprovalStatus    `json:"approval_status,omitempty"`
	AccountID       *uuid.UUID         `json:"account_id,omitempty"`
	DateRange       *DateRange         `json:"date_range,omitempty"`
	AmountRange     *AmountRange       `json:"amount_range,omitempty"`
	SearchTerm      *string            `json:"search_term,omitempty"`

	// Pagination
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`

	// Sorting
	SortBy    *string `json:"sort_by,omitempty"`
	SortOrder *string `json:"sort_order,omitempty"`
}

// EntryFilter defines filtering options for transaction entry queries
type EntryFilter struct {
	TransactionID *uuid.UUID   `json:"transaction_id,omitempty"`
	AccountID     *uuid.UUID   `json:"account_id,omitempty"`
	DateRange     *DateRange   `json:"date_range,omitempty"`
	AmountRange   *AmountRange `json:"amount_range,omitempty"`
	Reconciled    *bool        `json:"reconciled,omitempty"`
	CostCenterID  *uuid.UUID   `json:"cost_center_id,omitempty"` // Filter by cost centre FK
	CostCenter    *string      `json:"cost_center,omitempty"`    // Deprecated; prefer CostCenterID
	Department    *string      `json:"department,omitempty"`
	ProjectID     *uuid.UUID   `json:"project_id,omitempty"`

	// Pagination
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`
}

// AuditFilter defines filtering options for audit trail queries
type AuditFilter struct {
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	EntityType *string    `json:"entity_type,omitempty"`
	EntityID   *string    `json:"entity_id,omitempty"`
	EventType  *string    `json:"event_type,omitempty"`
	DateRange  *DateRange `json:"date_range,omitempty"`

	// Pagination
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`
}

// DateRange represents a date range filter
type DateRange struct {
	StartDate time.Time  `json:"start_date"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

// AmountRange represents an amount range filter
type AmountRange struct {
	MinAmount *decimal.Decimal `json:"min_amount,omitempty"`
	MaxAmount *decimal.Decimal `json:"max_amount,omitempty"`
}

// AuditEntry represents an audit trail entry
type AuditEntry struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	EntityType string     `json:"entity_type"`
	EntityID   *uuid.UUID `json:"entity_id"`
	EventType  string     `json:"event_type"`
	OldValues  *string    `json:"old_values,omitempty"`
	NewValues  *string    `json:"new_values,omitempty"`
	IPAddress  *string    `json:"ip_address,omitempty"`
	UserAgent  *string    `json:"user_agent,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ────────────────────────────────────────────────────────────────────────────
// Approval workflow persistence (TASK-033)
// ────────────────────────────────────────────────────────────────────────────

// ApprovalWorkflowStatus represents the lifecycle state of a finance approval workflow.
type ApprovalWorkflowStatus string

const (
	ApprovalWorkflowPending    ApprovalWorkflowStatus = "pending"
	ApprovalWorkflowInProgress ApprovalWorkflowStatus = "in_progress"
	ApprovalWorkflowCompleted  ApprovalWorkflowStatus = "completed"
	ApprovalWorkflowRejected   ApprovalWorkflowStatus = "rejected"
	ApprovalWorkflowCancelled  ApprovalWorkflowStatus = "cancelled"
)

// WorkflowRecord is a persisted row in finance_workflow_records.
type WorkflowRecord struct {
	ID            uuid.UUID              `json:"id"`
	TenantID      uuid.UUID              `json:"tenant_id"`
	TransactionID uuid.UUID              `json:"transaction_id"`
	Status        ApprovalWorkflowStatus `json:"status"`
	CurrentTier   int32                  `json:"current_tier"`
	InitiatedBy   uuid.UUID              `json:"initiated_by"`
	DueAt         *time.Time             `json:"due_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// ApprovalAction represents an action taken on an approval workflow.
type ApprovalAction string

const (
	ApprovalActionSubmitted ApprovalAction = "submitted"
	ApprovalActionApproved  ApprovalAction = "approved"
	ApprovalActionRejected  ApprovalAction = "rejected"
	ApprovalActionEscalated ApprovalAction = "escalated"
	ApprovalActionDelegated ApprovalAction = "delegated"
)

// ApprovalHistoryEntry is a persisted row in finance_approval_history.
type ApprovalHistoryEntry struct {
	ID            uuid.UUID      `json:"id"`
	TenantID      uuid.UUID      `json:"tenant_id"`
	WorkflowID    uuid.UUID      `json:"workflow_id"`
	TransactionID uuid.UUID      `json:"transaction_id"`
	Tier          int32          `json:"tier"`
	Action        ApprovalAction `json:"action"`
	PerformedBy   uuid.UUID      `json:"performed_by"`
	Notes         string         `json:"notes"`
	CreatedAt     time.Time      `json:"created_at"`
}

// ApprovalWorkflowRepository persists and queries approval workflow state.
type ApprovalWorkflowRepository interface {
	// CreateWorkflow inserts a new workflow record when a transaction is submitted.
	CreateWorkflow(ctx context.Context, rec *WorkflowRecord) error

	// GetWorkflowByTransaction returns the active workflow for a transaction.
	GetWorkflowByTransaction(ctx context.Context, transactionID uuid.UUID) (*WorkflowRecord, error)

	// UpdateWorkflowStatus updates the status and current approval tier.
	UpdateWorkflowStatus(ctx context.Context, workflowID uuid.UUID, status ApprovalWorkflowStatus, currentTier int32) error

	// InsertApprovalHistory records a single approval decision.
	InsertApprovalHistory(ctx context.Context, entry *ApprovalHistoryEntry) error

	// GetApprovalHistory returns all approval decisions for a transaction, oldest first.
	GetApprovalHistory(ctx context.Context, transactionID uuid.UUID) ([]*ApprovalHistoryEntry, error)

	// GetPendingByUser returns all pending workflow records for the given user or role.
	// role may be empty to match only by userID.
	GetPendingByUser(ctx context.Context, userID uuid.UUID, role string) ([]*WorkflowRecord, error)
}

// Repository aggregation interfaces for dependency injection

// RepositoryManager aggregates all repositories for easier dependency injection
type RepositoryManager interface {
	Accounts() AccountsRepository
	Transaction() TransactionRepository
	Audit() AuditRepository
}

// UnitOfWork defines transaction boundaries for multi-repository operations
type UnitOfWork interface {
	// Transaction management
	Begin(ctx context.Context) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error

	// Repository access within transaction
	Accounts() AccountsRepository
	Transaction() TransactionRepository
	Audit() AuditRepository
}

// AccountGroupRepository defines the contract for account group persistence
type AccountGroupRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, group *AccountGroup) error
	GetByID(ctx context.Context, id uuid.UUID) (*AccountGroup, error)
	GetByCode(ctx context.Context, entityID *uuid.UUID, groupCode string) (*AccountGroup, error)
	Update(ctx context.Context, group *AccountGroup) error
	Delete(ctx context.Context, id uuid.UUID) error

	// List and filtering operations
	List(ctx context.Context, filter *AccountGroupFilter) ([]*AccountGroup, error)
	Count(ctx context.Context, filter *AccountGroupFilter) (int64, error)
	ListByParent(ctx context.Context, parentID uuid.UUID) ([]*AccountGroup, error)
	GetRootGroups(ctx context.Context, entityID *uuid.UUID) ([]*AccountGroup, error)

	// Hierarchy operations
	GetGroupHierarchy(ctx context.Context) ([]*AccountGroup, error)
	GetGroupPath(ctx context.Context, groupID uuid.UUID) ([]*AccountGroup, error)
	ValidateHierarchy(ctx context.Context, groupID, parentID uuid.UUID) error
	GetChildren(ctx context.Context, groupID uuid.UUID) ([]*AccountGroup, error)
	HasChildren(ctx context.Context, groupID uuid.UUID) (bool, error)

	// Financial statement operations
	GetGroupsByFinancialStatement(ctx context.Context, statementType string) ([]*AccountGroup, error)
	GetGroupsByCashFlowCategory(ctx context.Context, category CashFlowCategory) ([]*AccountGroup, error)
	GetGroupsByConsolidationMethod(ctx context.Context, method ConsolidationMethod) ([]*AccountGroup, error)

	// Business logic queries
	GetSystemGroups(ctx context.Context) ([]*AccountGroup, error)
	GetCustomGroups(ctx context.Context, entityID *uuid.UUID) ([]*AccountGroup, error)
	GetActiveGroups(ctx context.Context, entityID *uuid.UUID) ([]*AccountGroup, error)
	Search(ctx context.Context, query string, limit int) ([]*AccountGroup, error)

	// Validation helpers
	ValidateGroupCode(ctx context.Context, code string, excludeID *uuid.UUID) error
	IsGroupCodeUnique(ctx context.Context, entityID *uuid.UUID, groupCode string, excludeID *uuid.UUID) (bool, error)
	CanDeleteGroup(ctx context.Context, groupID uuid.UUID) (bool, error)
}

// UnifiedAccountRepository defines operations for unified account/group queries
type UnifiedAccountRepository interface {
	// Unified hierarchy operations
	ListAccountsAndGroups(ctx context.Context, filter *AccountGroupFilter) ([]*Accounts, error)
	GetHierarchyWithGroups(ctx context.Context, entityID *uuid.UUID) ([]*Accounts, error)
	SearchAccountsAndGroups(ctx context.Context, query string, limit int) ([]*Accounts, error)
	GetNodePath(ctx context.Context, nodeID uuid.UUID, isGroup bool) ([]*Accounts, error)

	// Tree operations
	GetSubtree(ctx context.Context, rootID uuid.UUID, isGroup bool, maxDepth *int) ([]*Accounts, error)
	GetSiblings(ctx context.Context, nodeID uuid.UUID, isGroup bool) ([]*Accounts, error)
	GetNodeChildren(ctx context.Context, nodeID uuid.UUID, isGroup bool) ([]*Accounts, error)
}

// Repository factory for creating repository instances
type RepositoryFactory interface {
	CreateAccountsRepository() AccountsRepository
	CreateAccountGroupRepository() AccountGroupRepository
	CreateUnifiedAccountRepository() UnifiedAccountRepository
	CreateTransactionRepository() TransactionRepository
	CreateAuditRepository() AuditRepository
	CreateUnitOfWork() UnitOfWork
}

// PeriodRepository defines persistence operations for fiscal years and accounting periods.
type PeriodRepository interface {
	// FiscalYear operations
	CreateFiscalYear(ctx context.Context, fy *FiscalYear) error
	GetFiscalYearByID(ctx context.Context, id uuid.UUID) (*FiscalYear, error)
	GetFiscalYearByYear(ctx context.Context, tenantID uuid.UUID, year int) (*FiscalYear, error)
	ListFiscalYears(ctx context.Context, tenantID uuid.UUID) ([]*FiscalYear, error)
	UpdateFiscalYear(ctx context.Context, fy *FiscalYear) error

	// AccountingPeriod operations
	CreatePeriod(ctx context.Context, period *AccountingPeriod) error
	GetPeriodByID(ctx context.Context, id uuid.UUID) (*AccountingPeriod, error)
	GetPeriodForDate(ctx context.Context, tenantID uuid.UUID, date time.Time) (*AccountingPeriod, error)
	GetCurrentPeriod(ctx context.Context, tenantID uuid.UUID) (*AccountingPeriod, error)
	ListPeriods(ctx context.Context, tenantID, fiscalYearID uuid.UUID) ([]*AccountingPeriod, error)
	UpdatePeriod(ctx context.Context, period *AccountingPeriod) error
}

// CostCenterRepository defines persistence operations for cost centres.
type CostCenterRepository interface {
	Create(ctx context.Context, cc *CostCenter) error
	GetByID(ctx context.Context, id uuid.UUID) (*CostCenter, error)
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*CostCenter, error)
	Update(ctx context.Context, cc *CostCenter) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, activeOnly bool) ([]*CostCenter, error)
	ValidateCode(ctx context.Context, tenantID uuid.UUID, code string, excludeID *uuid.UUID) error
}

// - NotificationRepository for system notifications
// - CurrencyRepository for exchange rate management

// ExchangeRateRepository defines persistence for exchange rates.
// All reads use the on-date-or-before look-up semantics described in the
// currency-management guide (most recent rate on/before the requested date).
type ExchangeRateRepository interface {
	// Upsert inserts or updates a rate for (fromCurrency, toCurrency, effectiveDate, rateType).
	Upsert(ctx context.Context, rate *ExchangeRate) error
	// GetRate returns the most recent rate for the pair on or before asOfDate.
	GetRate(ctx context.Context, fromCurrency, toCurrency string, rateType RateType, asOfDate time.Time) (*ExchangeRate, error)
	// ListRates lists rates for a currency pair within an optional date range.
	ListRates(ctx context.Context, fromCurrency, toCurrency string, from, to *time.Time, limit int) ([]*ExchangeRate, error)
	// DeleteExpired removes rates whose expiry_date is before cutoff.
	DeleteExpired(ctx context.Context, cutoff time.Time) error
}

// BudgetRepository defines persistence for budget headers and line items.
type BudgetRepository interface {
	// Budget header operations
	CreateBudget(ctx context.Context, b *Budget) error
	GetBudgetByID(ctx context.Context, id uuid.UUID) (*Budget, error)
	ListBudgets(ctx context.Context, tenantID uuid.UUID, fiscalYearID *uuid.UUID) ([]*Budget, error)
	UpdateBudget(ctx context.Context, b *Budget) error

	// Budget line item operations
	CreateLineItems(ctx context.Context, lines []*BudgetLineItem) error
	GetLineItems(ctx context.Context, budgetID uuid.UUID) ([]*BudgetLineItem, error)
	DeleteLineItems(ctx context.Context, budgetID uuid.UUID) error
}

// ReconciliationRepository defines persistence for bank statements and statement lines.
type ReconciliationRepository interface {
	// Statement header operations
	CreateStatement(ctx context.Context, s *BankStatement) error
	GetStatementByID(ctx context.Context, id uuid.UUID) (*BankStatement, error)
	ListStatements(ctx context.Context, tenantID uuid.UUID, accountID *uuid.UUID) ([]*BankStatement, error)
	UpdateStatement(ctx context.Context, s *BankStatement) error

	// Statement line operations
	CreateLines(ctx context.Context, lines []*BankStatementLine) error
	GetLine(ctx context.Context, lineID uuid.UUID) (*BankStatementLine, error)
	ListLines(ctx context.Context, statementID uuid.UUID, unmatchedOnly bool) ([]*BankStatementLine, error)
	MatchLine(ctx context.Context, lineID, entryID uuid.UUID, byUserID uuid.UUID) error
	UnmatchLine(ctx context.Context, lineID uuid.UUID) error

	// Reconciliation completion
	CompleteReconciliation(ctx context.Context, statementID uuid.UUID, byUserID uuid.UUID) error
}

// TaxRepository defines persistence for tax authorities, codes, and brackets.
type TaxRepository interface {
	// Tax authority operations
	CreateAuthority(ctx context.Context, a *TaxAuthority) error
	GetAuthorityByID(ctx context.Context, id uuid.UUID) (*TaxAuthority, error)
	GetAuthorityByCode(ctx context.Context, tenantID uuid.UUID, code string) (*TaxAuthority, error)
	ListAuthorities(ctx context.Context, tenantID uuid.UUID, activeOnly bool) ([]*TaxAuthority, error)
	UpdateAuthority(ctx context.Context, a *TaxAuthority) error

	// Tax code operations
	CreateTaxCode(ctx context.Context, tc *TaxCode) error
	GetTaxCodeByID(ctx context.Context, id uuid.UUID) (*TaxCode, error)
	GetTaxCodeByCode(ctx context.Context, tenantID uuid.UUID, code string) (*TaxCode, error)
	ListTaxCodes(ctx context.Context, tenantID uuid.UUID, taxType *TaxType, activeOnly bool) ([]*TaxCode, error)
	UpdateTaxCode(ctx context.Context, tc *TaxCode) error

	// Tax bracket operations
	CreateBrackets(ctx context.Context, taxCodeID uuid.UUID, brackets []*TaxBracket) error
	GetBrackets(ctx context.Context, taxCodeID uuid.UUID) ([]*TaxBracket, error)
	DeleteBrackets(ctx context.Context, taxCodeID uuid.UUID) error
}

// NOTE: All repository implementations should:
// 1. Enforce tenant isolation using current_tenant_id()
// 2. Handle optimistic locking using version fields
// 3. Implement proper error handling and logging
// 4. Support both single and batch operations where applicable
// 5. Use prepared statements for better performance
// 6. Implement connection pooling and timeout handling
