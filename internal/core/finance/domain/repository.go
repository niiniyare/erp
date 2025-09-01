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

	// Enhanced view-based operations
	GetAccountWithGroups(ctx context.Context, id uuid.UUID) (*AccountWithGroups, error)
	GetAccountWithGroupsByCode(ctx context.Context, code string) (*AccountWithGroups, error)
	ListAccountsWithGroups(ctx context.Context, filter *AccountFilter) ([]*AccountWithGroups, error)
	SearchAccountsWithGroups(ctx context.Context, query string, limit int) ([]*AccountWithGroups, error)
	GetLeafAccountsOnly(ctx context.Context, rootType *string) ([]*AccountWithGroups, error)

	// Complete chart of accounts operations
	GetCompleteChartOfAccounts(ctx context.Context, filter *ChartOfAccountsFilter) ([]*ChartOfAccountsComplete, error)
	GetAccountForReporting(ctx context.Context, accountID uuid.UUID) (*ChartOfAccountsComplete, error)
	GetAccountsByStatementSection(ctx context.Context, section string, entityID *uuid.UUID) ([]*ChartOfAccountsComplete, error)
	GetAccountsByGroup(ctx context.Context, groupCode string, entityID *uuid.UUID) ([]*ChartOfAccountsComplete, error)
	GetAccountsByHeader(ctx context.Context, headerCode string, entityID *uuid.UUID) ([]*ChartOfAccountsComplete, error)

	// Financial reporting operations
	GetTrialBalanceAccounts(ctx context.Context, entityID *uuid.UUID, nonZeroOnly bool) ([]*TrialBalanceSummary, error)
	GetAccountsWithBalances(ctx context.Context, filter *BalanceFilter) ([]*ChartOfAccountsComplete, error)
	GetCashFlowAccounts(ctx context.Context, entityID *uuid.UUID) ([]*CashFlowAccount, error)
	GetAccountSummaryByGroup(ctx context.Context, entityID *uuid.UUID) ([]*AccountGroupSummary, error)
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

// AuditRepository defines the contract for audit trail persistence
type AuditRepository interface {
	CreateAuditEntry(ctx context.Context, entry *AuditEntry) error
	GetAuditTrail(ctx context.Context, entityType, entityID string, filter *AuditFilter) ([]*AuditEntry, error)
	GetUserActivity(ctx context.Context, userID uuid.UUID, filter *AuditFilter) ([]*AuditEntry, error)
	PurgeOldEntries(ctx context.Context, olderThan time.Time) error
}

// Filter and parameter structures

// AccountFilter defines filtering options for chart of accounts queries
type AccountFilter struct {
	EntityID   *uuid.UUID `json:"entity_id,omitempty"`
	RootType   *RootType  `json:"root_type,omitempty"`
	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
	IsActive   *bool      `json:"is_active,omitempty"`
	SearchTerm *string    `json:"search_term,omitempty"`

	// Pagination
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`

	// Sorting
	SortBy    *string `json:"sort_by,omitempty"`
	SortOrder *string `json:"sort_order,omitempty"` // ASC or DESC
}

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
	CostCenter    *string      `json:"cost_center,omitempty"`
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

// Repository factory for creating repository instances
type RepositoryFactory interface {
	CreateAccountsRepository() AccountsRepository
	CreateTransactionRepository() TransactionRepository
	CreateAuditRepository() AuditRepository
	CreateUnitOfWork() UnitOfWork
}

// TODO: Consider adding the following repository interfaces for extended functionality:
// - ReconciliationRepository for bank reconciliation
// - ReportingRepository for optimized reporting queries
// - ConfigurationRepository for tenant/entity-specific settings
// - AttachmentRepository for document management
// - NotificationRepository for system notifications
// - CurrencyRepository for exchange rate management

// NOTE: All repository implementations should:
// 1. Enforce tenant isolation using current_tenant_id()
// 2. Handle optimistic locking using version fields
// 3. Implement proper error handling and logging
// 4. Support both single and batch operations where applicable
// 5. Use prepared statements for better performance
// 6. Implement connection pooling and timeout handling
