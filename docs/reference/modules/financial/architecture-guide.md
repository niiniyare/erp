# Financial Module Architecture Guide

## Overview

The Financial Module implements Clean Architecture principles with Domain-Driven Design (DDD) patterns, providing a robust, scalable, and maintainable financial accounting system. This guide details the architectural layers, design patterns, SQL schema implementation, and SQLC integration strategies used throughout the module.

## SQL Schema Architecture

### Performance Optimization Strategy

**1. Comprehensive Indexing**
```sql
-- Multi-column indexes for common access patterns
CREATE INDEX idx_finance_transactions_tenant_date 
  ON finance_transactions (tenant_id, transaction_date);

-- Partial indexes to exclude unnecessary data
CREATE INDEX idx_finance_transactions_approval 
  ON finance_transactions (tenant_id, approval_status)
  WHERE approval_required = TRUE;

-- Conditional indexes for specific use cases
CREATE INDEX idx_finance_transactions_recurring 
  ON finance_transactions (tenant_id, next_recurring_date)
  WHERE is_recurring = TRUE;
```

**2. N+1 Query Prevention**
```sql
-- Single query to get transactions with entries and account details
SELECT t.*, te.*, a.account_code, a.account_name, a.root_type
FROM finance_transactions t
LEFT JOIN finance_transaction_entries te ON t.id = te.transaction_id
LEFT JOIN finance_accounts a ON te.account_id = a.id
WHERE t.id = $1 AND t.tenant_id = current_tenant_id()
ORDER BY te.entry_number ASC;
```

**3. View-Based Query Optimization**
```sql
-- Pre-computed views for complex reporting queries
-- v_chart_of_accounts_complete includes group hierarchies
-- v_finance_account_activity includes 30-day metrics
-- v_financial_statement_builder for report generation

-- Queries leverage these views instead of complex joins
SELECT * FROM v_chart_of_accounts_complete 
WHERE statement_section = 'Balance Sheet'
ORDER BY display_order;
```

### Audit Trail & Versioning Implementation

**1. Comprehensive Audit Fields**
```sql
-- Standard audit pattern across all tables
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
deleted_at TIMESTAMPTZ,           -- Soft delete pattern
created_by UUID REFERENCES users(id),
updated_by UUID REFERENCES users(id),
posted_by UUID REFERENCES users(id),  -- Financial-specific
posted_at TIMESTAMPTZ
```

**2. Optimistic Locking Pattern**
```sql
-- Version-based concurrency control
version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),

-- Auto-increment version trigger
CREATE TRIGGER trigger_finance_transaction_version BEFORE UPDATE
  ON finance_transactions FOR EACH ROW 
  EXECUTE FUNCTION increment_finance_transaction_version();
```

**3. Transaction Reversal Audit Trail**
```sql
-- Complete reversal tracking
is_reversed BOOLEAN DEFAULT false,
reversed_by_transaction_id UUID REFERENCES finance_transactions(id),
reversal_reason TEXT,

-- Constraint ensures reversal integrity
CONSTRAINT reversal_logic CHECK (
  CASE WHEN is_reversed = TRUE 
  THEN reversed_by_transaction_id IS NOT NULL
  ELSE reversed_by_transaction_id IS NULL END
)
```

### SQLC Integration Patterns

**1. Type-Safe Parameter Mapping**
```go
// SQLC generates strongly-typed parameter structs
type ApproveTransactionParams struct {
    ApprovedBy    *uuid.UUID `json:"approved_by"`
    ApprovalNotes *string    `json:"approval_notes"`
    TransactionID uuid.UUID  `json:"transaction_id"`
}

// Nullable parameters use sqlc.narg(), required use sqlc.arg()
// SQL: WHERE id = sqlc.arg('transaction_id') AND approval_status = 'PENDING'
// Go:  row := q.db.QueryRow(ctx, query, arg.TransactionID)
```

**2. Financial Data Type Safety**
```go
// PostgreSQL DECIMAL maps to pgtype.Numeric for precision
type FinanceTransaction struct {
    ExchangeRate      pgtype.Numeric `json:"exchange_rate"`
    TotalDebitAmount  pgtype.Numeric `json:"total_debit_amount"`
    TotalCreditAmount pgtype.Numeric `json:"total_credit_amount"`
}

// Arrays mapped to Go slices
AttachmentIds []string `json:"attachment_ids"`
Tags          []string `json:"tags"`

// JSONB fields mapped to []byte for flexibility
ValidationErrors      []byte `json:"validation_errors"`
TransactionAttributes []byte `json:"transaction_attributes"`
```

**3. Query Method Generation Patterns**
```go
// :one queries return single struct pointer
func (q *Queries) GetTransactionByID(ctx context.Context, 
    transactionID uuid.UUID) (*FinanceTransaction, error)

// :many queries return slices
func (q *Queries) ListTransactions(ctx context.Context, 
    arg ListTransactionsParams) ([]*FinanceTransaction, error)

// :exec queries return error only
func (q *Queries) SoftDeleteTransaction(ctx context.Context, 
    arg SoftDeleteTransactionParams) error
```

### Bulk Operations & Performance

**1. Array-Based Bulk Updates**
```sql
-- Bulk transaction tag updates
UPDATE finance_transactions 
SET tags = $1::VARCHAR[], updated_by = $2
WHERE id = ANY($3::UUID[]) AND tenant_id = current_tenant_id()

-- Bulk entry reconciliation
UPDATE finance_transaction_entries
SET reconciled = TRUE, reconciled_date = $1
WHERE id = ANY($2::uuid[]) AND tenant_id = current_tenant_id()
```

**2. Efficient Aggregation**
```sql
-- Account balance calculations with proper grouping
SELECT a.account_code, a.account_name,
       COALESCE(SUM(CASE WHEN te.debit_amount > 0 THEN te.debit_amount ELSE 0 END), 0) AS total_debits,
       COALESCE(SUM(CASE WHEN te.credit_amount > 0 THEN te.credit_amount ELSE 0 END), 0) AS total_credits
FROM finance_accounts a
LEFT JOIN finance_transaction_entries te ON a.id = te.account_id
GROUP BY a.id, a.account_code, a.account_name
```

**3. Pagination Patterns**
```sql
-- Consistent pagination across all list queries
ORDER BY transaction_date DESC, created_at DESC
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count')

-- Count queries for pagination metadata
SELECT COUNT(*) FROM finance_transactions WHERE [conditions]
```

## Architectural Layers

### Transaction Management Patterns

**1. State Machine Enforcement**
```sql
-- Transaction status progression constraints
transaction_status VARCHAR(20) NOT NULL DEFAULT 'DRAFT' CHECK (
  transaction_status IN ('DRAFT', 'PENDING_APPROVAL', 'APPROVED', 'POSTED', 'CANCELLED', 'REVERSED')
)

-- Business logic constraints for state transitions
CONSTRAINT posting_date_logic CHECK (
  CASE WHEN transaction_status = 'POSTED' 
  THEN posting_date IS NOT NULL AND posted_by IS NOT NULL
  ELSE TRUE END
)
```

**2. Atomic State Changes with Validation**
```sql
-- PostTransaction ensures valid status transition
UPDATE finance_transactions 
SET transaction_status = 'POSTED',
    posted_by = $1, posted_at = NOW()
WHERE transaction_status IN ('APPROVED', 'DRAFT')
  AND id = $2 AND tenant_id = current_tenant_id()
```

**3. Double-Entry Validation**
```sql
-- Database-level balance validation
CONSTRAINT balanced_transaction CHECK (
  CASE WHEN transaction_status IN ('POSTED', 'APPROVED') 
  THEN total_debit_amount = total_credit_amount
  ELSE TRUE END
)

-- Entry-level validation ensures single-sided entries
CHECK (NOT (debit_amount > 0 AND credit_amount > 0))
CHECK (debit_amount > 0 OR credit_amount > 0)
```

### Complex Query Patterns & View Optimization

**1. Hierarchical Query Decomposition**
```sql
-- Recursive CTE broken into logical steps for account hierarchy
WITH RECURSIVE account_hierarchy AS (
  -- Step 1: Root accounts
  SELECT *, 1 AS hierarchy_level, account_code::text AS full_path
  FROM v_chart_of_accounts_complete 
  WHERE parent_account_id IS NULL AND tenant_id = current_tenant_id()
  UNION ALL  
  -- Step 2: Child accounts with depth limiting
  SELECT coa.*, ah.hierarchy_level + 1, 
         (ah.full_path || '.' || coa.account_code)::text
  FROM v_chart_of_accounts_complete coa
  JOIN account_hierarchy ah ON coa.parent_account_id = ah.account_id
  WHERE ah.hierarchy_level < 10 -- Prevent infinite recursion
)
SELECT * FROM account_hierarchy
WHERE hierarchy_level <= $1 -- Max depth parameter
ORDER BY full_path;
```

**2. View-Based Query Decomposition**
```sql
-- Complex hierarchy queries broken into manageable views
-- Base view: v_finance_accounts_with_groups
-- Enhanced view: v_chart_of_accounts_complete  
-- Reporting view: v_financial_statement_builder

-- Instead of complex joins in every query:
SELECT account_code, account_name, group_name, statement_section
FROM v_chart_of_accounts_complete
WHERE tenant_id = current_tenant_id()
  AND statement_section = 'Balance Sheet'
ORDER BY display_order;
```

**3. Conditional Query Optimization**
```sql
-- Optional filters using COALESCE pattern for single query handling multiple scenarios
WHERE (sqlc.narg('account_type')::VARCHAR IS NULL OR account_type = sqlc.narg('account_type'))
  AND (sqlc.narg('is_active')::bool IS NULL OR is_active = sqlc.narg('is_active'))
  AND (sqlc.narg('date_from')::date IS NULL OR transaction_date >= sqlc.narg('date_from'))
  AND (sqlc.narg('date_to')::date IS NULL OR transaction_date <= sqlc.narg('date_to'))

-- This allows single query to handle multiple filter combinations efficiently
```

**4. Partitioning Strategy (Future-Ready)**
```sql
-- Logical partitioning via tenant isolation
-- Row Level Security: Automatic tenant-based partitioning through RLS policies
ALTER TABLE finance_transactions ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON finance_transactions 
FOR ALL TO application_role USING (
  current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id()
);

-- Date-based query patterns ready for time-based partitioning
-- Candidate tables: finance_transactions (by transaction_date), 
-- audit_log (by created_at)
```

### 1. Handler Layer (Interface Adapters)
**Location**: `/internal/api/handlers/finance/`

**Responsibilities**:
- HTTP request/response handling
- Input validation and sanitization
- Authentication and authorization verification
- Error handling and response formatting
- Observability (tracing, metrics, logging)

**Key Components**:
```go
// FinanceHandler handles finance-related HTTP requests
type FinanceHandler struct {
    services  *financeService.Services
    logger    logger.Logger
    metrics   metrics.MetricsProvider
    tracer    tracing.Service
    validator *validator.Validate
}

// Example handler method with SQLC integration
func (h *FinanceHandler) CreateAccount(c *fiber.Ctx) error {
    // 1. Start tracing
    ctx, span := h.tracer.StartSpan(c.Context(), "finance.CreateAccount")
    defer span.End()
    
    // 2. Parse and validate request
    var req financeDomain.CreateAccountRequest
    if err := h.ValidateRequest(c, &req); err != nil {
        return h.HandleError(c, err)
    }
    
    // 3. Delegate to service layer (uses SQLC-generated queries)
    account, err := h.services.Account.Create(c.Context(), &req)
    if err != nil {
        return h.HandleError(c, err)
    }
    
    // 4. Return response
    return h.Created(c, account)
}
```

### 2. Service Layer (Application/Use Cases)
**Location**: `/internal/core/finance/service/`

**Responsibilities**:
- Business logic orchestration
- Transaction coordination
- Cross-cutting concerns (caching, notifications)
- Integration with external services
- Workflow management

**Key Services**:

**AccountService** - Account management and hierarchy operations
```go
type AccountService interface {
    // CRUD operations
    Create(ctx context.Context, req *domain.CreateAccountRequest) (*domain.Accounts, error)
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Accounts, error)
    Update(ctx context.Context, id uuid.UUID, req *domain.UpdateAccountRequest) (*domain.Accounts, error)
    Delete(ctx context.Context, id uuid.UUID) error
    
    // Business operations
    GetAccountHierarchy(ctx context.Context, rootAccountID uuid.UUID) ([]*domain.Accounts, error)
    ValidateAccountCode(ctx context.Context, code string, excludeID *uuid.UUID) error
    GetTrialBalanceAccounts(ctx context.Context, entityID *uuid.UUID, nonZeroOnly bool) ([]*domain.TrialBalanceSummary, error)
}
```

**TransactionService** - Transaction processing and workflow management
```go
type TransactionService interface {
    // Core operations
    Create(ctx context.Context, req *domain.CreateTransactionRequest) (*domain.Transaction, error)
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error)
    
    // Workflow operations
    Submit(ctx context.Context, id uuid.UUID) error
    Approve(ctx context.Context, id uuid.UUID) error
    Post(ctx context.Context, id uuid.UUID) error
    Reverse(ctx context.Context, id uuid.UUID, reason string) error
    
    // Validation
    ValidateTransaction(ctx context.Context, transaction *domain.Transaction) error
}
```

### 3. Domain Layer (Entities)
**Location**: `/internal/core/finance/domain/`

**Responsibilities**:
- Core business entities and value objects
- Business rules and validation logic
- Domain events and aggregates
- Repository contracts (interfaces)

**Core Entities**:

**Account Entity** - Chart of accounts with business logic
```go
type Accounts struct {
    // Identity
    ID       uuid.UUID  `json:"id"`
    TenantID uuid.UUID  `json:"tenant_id"`
    EntityID *uuid.UUID `json:"entity_id,omitempty"`
    
    // Account identification
    AccountCode string  `json:"account_code"`
    AccountName string  `json:"account_name"`
    
    // Hierarchy
    ParentAccountID *uuid.UUID `json:"parent_account_id,omitempty"`
    AccountLevel    int32      `json:"account_level"`
    AccountPath     *string    `json:"account_path,omitempty"`
    
    // Classification
    RootType        RootType      `json:"root_type"`
    AccountType     string        `json:"account_type"`
    NormalBalance   NormalBalance `json:"normal_balance"`
    
    // Financial tracking
    CurrentBalance decimal.Decimal `json:"current_balance"`
    YTDBalance     decimal.Decimal `json:"ytd_balance"`
    
    // Operational settings
    IsActive           bool `json:"is_active"`
    AllowManualEntries bool `json:"allow_manual_entries"`
    RequireReference   bool `json:"require_reference"`
    
    // Audit fields
    Version   int32     `json:"version"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    CreatedBy uuid.UUID `json:"created_by"`
}

// Business logic methods
func (a *Accounts) CanReceiveEntries() bool {
    return a.IsActive && a.AllowManualEntries && !a.HasChildren
}

func (a *Accounts) UpdateBalance(amount decimal.Decimal, isDebit bool) {
    if (isDebit && a.NormalBalance == NormalBalanceDebit) ||
       (!isDebit && a.NormalBalance == NormalBalanceCredit) {
        a.CurrentBalance = a.CurrentBalance.Add(amount)
    } else {
        a.CurrentBalance = a.CurrentBalance.Sub(amount)
    }
    a.UpdatedAt = time.Now()
}
```

**Transaction Entity** - Financial transaction with state machine
```go
type Transaction struct {
    ID                uuid.UUID         `json:"id"`
    TenantID          uuid.UUID         `json:"tenant_id"`
    TransactionNumber string            `json:"transaction_number"`
    TransactionType   TransactionType   `json:"transaction_type"`
    TransactionStatus TransactionStatus `json:"transaction_status"`
    
    // Financial information
    TotalDebitAmount  decimal.Decimal `json:"total_debit_amount"`
    TotalCreditAmount decimal.Decimal `json:"total_credit_amount"`
    CurrencyCode      string          `json:"currency_code"`
    ExchangeRate      decimal.Decimal `json:"exchange_rate"`
    
    // Workflow
    ApprovalRequired bool           `json:"approval_required"`
    ApprovalStatus   ApprovalStatus `json:"approval_status"`
    ApprovedBy       *uuid.UUID     `json:"approved_by,omitempty"`
    ApprovedAt       *time.Time     `json:"approved_at,omitempty"`
    
    // Entries
    Entries []TransactionEntry `json:"entries,omitempty"`
    
    // Audit
    Version   int32     `json:"version"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Business logic methods
func (t *Transaction) IsBalanced() bool {
    return t.TotalDebitAmount.Equal(t.TotalCreditAmount)
}

func (t *Transaction) CalculateTotals() {
    t.TotalDebitAmount = decimal.Zero
    t.TotalCreditAmount = decimal.Zero
    
    for _, entry := range t.Entries {
        t.TotalDebitAmount = t.TotalDebitAmount.Add(entry.DebitAmount)
        t.TotalCreditAmount = t.TotalCreditAmount.Add(entry.CreditAmount)
    }
}

func (t *Transaction) CanBeModified() bool {
    return t.TransactionStatus == TransactionStatusDraft ||
           t.TransactionStatus == TransactionStatusPendingApproval
}
```

### 4. Repository Layer (Infrastructure)
**Location**: `/internal/core/finance/repository/`

**Responsibilities**:
- Data persistence implementation
- Database query optimization
- Tenant isolation enforcement
- Error handling and mapping

**Implementation Patterns**:
```go
type accountsRepository struct {
    store   db.Store
    cache   cache.Service
    tracing tracing.Service
}

func (r *accountsRepository) Create(ctx context.Context, account *domain.Accounts) error {
    ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.Create")
    defer span.End()
    
    // Use tenant-aware transaction
    return r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
        // Map domain to SQLC parameters
        params, err := mapDomainAccountToSQLCCreateDirect(account)
        if err != nil {
            return fmt.Errorf("failed to map create account request: %w", err)
        }
        
        // Execute query
        sqlcAccount, err := s.CreateAccount(ctx, params)
        if err != nil {
            return r.mapDatabaseError(err, "create_account")
        }
        
        // Update domain object with generated fields
        account.ID = sqlcAccount.ID
        account.CreatedAt = sqlcAccount.CreatedAt
        account.UpdatedAt = sqlcAccount.UpdatedAt
        
        return nil
    })
}
```

## Design Patterns Used

### 1. Repository Pattern
**Purpose**: Abstracts data access logic from business logic
**Implementation**: Domain interfaces with infrastructure implementations

```go
// Domain layer interface
type AccountsRepository interface {
    Create(ctx context.Context, account *Accounts) error
    GetByID(ctx context.Context, id uuid.UUID) (*Accounts, error)
    List(ctx context.Context, filter *AccountFilter) ([]*Accounts, error)
    // ... other methods
}

// Infrastructure layer implementation
type accountsRepository struct {
    store   db.Store
    cache   cache.Service
    tracing tracing.Service
}
```

### 2. Unit of Work Pattern
**Purpose**: Manages transactions across multiple repositories
**Implementation**: Database transaction boundaries

```go
type UnitOfWork interface {
    Begin(ctx context.Context) error
    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error
    
    Accounts() AccountsRepository
    Transaction() TransactionRepository
    Audit() AuditRepository
}
```

### 3. State Machine Pattern
**Purpose**: Manages complex transaction status transitions
**Implementation**: Transaction status validation and transitions

```go
type TransactionStateMachine struct {
    transaction *Transaction
}

func (tsm *TransactionStateMachine) CanTransitionTo(toStatus TransactionStatus) (bool, string) {
    currentStatus := tsm.transaction.TransactionStatus
    
    validTransitions, exists := TransactionStatusTransitions[currentStatus]
    if !exists {
        return false, fmt.Sprintf("no transitions defined for status %s", currentStatus)
    }
    
    for _, validStatus := range validTransitions {
        if validStatus == toStatus {
            return tsm.validateTransitionRules(currentStatus, toStatus)
        }
    }
    
    return false, fmt.Sprintf("invalid transition from %s to %s", currentStatus, toStatus)
}
```

### 4. Strategy Pattern
**Purpose**: Different validation strategies and business rules
**Implementation**: Multiple validator implementations

```go
type DoubleEntryValidator interface {
    ValidateTransaction(ctx context.Context, transaction *Transaction) *ValidationResult
    ValidateBalance(ctx context.Context, transaction *Transaction) *ValidationResult
    ValidateAccountCompatibility(ctx context.Context, entries []TransactionEntry) *ValidationResult
}

type StandardDoubleEntryValidator struct {
    accountRepo domain.AccountsRepository
    logger      logger.Logger
}

type EnhancedDoubleEntryValidator struct {
    accountRepo     domain.AccountsRepository
    complianceRules ComplianceRulesEngine
    logger          logger.Logger
}
```

### 5. Factory Pattern
**Purpose**: Service and repository creation with dependency injection
**Implementation**: Constructor functions and wire-generated factories

```go
func NewAccountService(
    accountRepo domain.AccountsRepository,
    accountGroupRepo domain.AccountGroupRepository,
    tracing tracing.Service,
    metrics metrics.MetricsProvider,
    settingsHelper *SettingsHelper,
    iamService iam.Service,
    featureFlagService featureflag.Service,
) AccountService {
    return &accountService{
        accountRepo:        accountRepo,
        accountGroupRepo:   accountGroupRepo,
        tracing:            tracing,
        metrics:            metrics,
        settingsHelper:     settingsHelper,
        iamService:         iamService,
        featureFlagService: featureFlagService,
    }
}
```

### 6. Decorator Pattern
**Purpose**: Adding cross-cutting concerns like caching and observability
**Implementation**: Service wrappers

```go
type CachedAccountService struct {
    AccountService
    cache  cache.Service
    tracer tracing.Service
}

func (cas *CachedAccountService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Accounts, error) {
    ctx, span := cas.tracer.StartSpan(ctx, "CachedAccountService.GetByID")
    defer span.End()
    
    // Try cache first
    cacheKey := fmt.Sprintf("account:%s", id.String())
    var account domain.Accounts
    if err := cas.cache.Get(ctx, cacheKey, &account); err == nil {
        return &account, nil
    }
    
    // Fallback to service
    account, err := cas.AccountService.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    cas.cache.Set(ctx, cacheKey, account, 15*time.Minute)
    return account, nil
}
```

### 7. Observer Pattern
**Purpose**: Event-driven architecture for audit and notifications
**Implementation**: Domain events and event handlers

```go
type DomainEvent interface {
    EventID() string
    EventType() string
    OccurredAt() time.Time
    AggregateID() string
}

type AccountCreatedEvent struct {
    ID         string    `json:"id"`
    AccountID  uuid.UUID `json:"account_id"`
    TenantID   uuid.UUID `json:"tenant_id"`
    CreatedBy  uuid.UUID `json:"created_by"`
    OccurredOn time.Time `json:"occurred_on"`
}

type EventHandler interface {
    Handle(ctx context.Context, event DomainEvent) error
}
```

## Data Flow Architecture

### Request Flow
```
1. HTTP Request → Fiber Router
   ↓
2. Middleware Stack (Auth, Tenant, Logging, Metrics)
   ↓
3. Handler Layer (Validation, Request Mapping)
   ↓
4. Service Layer (Business Logic, Orchestration)
   ↓
5. Repository Layer (Data Access, Caching)
   ↓
6. Database (PostgreSQL with RLS)
```

### Error Handling Flow
```go
// Error mapping and handling
func (r *accountsRepository) mapDatabaseError(err error, operation string) error {
    if err == db.ErrNoRows {
        return domain.ErrAccountNotFound
    }
    
    if pgErr, ok := err.(*pgconn.PgError); ok {
        switch pgErr.Code {
        case "23505": // unique violation
            if strings.Contains(pgErr.Message, "account_code") {
                return domain.ErrAccountCodeExists
            }
        case "23503": // foreign key violation
            return fmt.Errorf("foreign key constraint violation: %s", pgErr.Message)
        }
    }
    
    return fmt.Errorf("database operation failed: %s: %w", operation, err)
}
```

### Transaction Management
```go
// Atomic transaction operations
func (s *transactionService) PostTransaction(ctx context.Context, transactionID uuid.UUID) error {
    return s.unitOfWork.WithTransaction(ctx, func(uow UnitOfWork) error {
        // 1. Get transaction
        transaction, err := uow.Transaction().GetByID(ctx, transactionID)
        if err != nil {
            return err
        }
        
        // 2. Validate
        if err := s.validator.ValidateTransaction(ctx, transaction); err.HasErrors() {
            return err
        }
        
        // 3. Update account balances
        for _, entry := range transaction.Entries {
            if err := uow.Accounts().UpdateBalance(ctx, entry.AccountID, entry); err != nil {
                return err
            }
        }
        
        // 4. Update transaction status
        transaction.TransactionStatus = TransactionStatusPosted
        transaction.PostedAt = time.Now()
        if err := uow.Transaction().Update(ctx, transaction); err != nil {
            return err
        }
        
        // 5. Create audit entry
        auditEntry := &domain.AuditEntry{
            EntityType: "transaction",
            EntityID:   transactionID,
            EventType:  "posted",
            UserID:     s.getCurrentUserID(ctx),
        }
        return uow.Audit().CreateAuditEntry(ctx, auditEntry)
    })
}
```

## Performance Optimizations

### 1. Caching Strategy
- **Account Hierarchy**: 15-minute TTL for hierarchy queries
- **Exchange Rates**: Real-time caching with 1-minute TTL
- **Account Balances**: Cached balance calculations for reporting

### 2. Database Optimizations
- **Indexes**: Optimized for common query patterns
- **Materialized Views**: Pre-computed financial reports
- **Connection Pooling**: Efficient database connection management

### 3. Tenant Isolation
- **Row-Level Security (RLS)**: Database-level tenant isolation
- **Query Optimization**: Tenant-aware query planning
- **Cache Isolation**: Tenant-specific cache keys

## Security Architecture

### 1. Multi-Tenant Security
- **Data Isolation**: Complete separation via RLS policies
- **API Security**: Tenant context validation
- **Cross-Tenant Prevention**: Request filtering and validation

### 2. Access Control Integration
```go
// IAM integration for authorization
func (s *accountService) Create(ctx context.Context, req *domain.CreateAccountRequest) (*domain.Accounts, error) {
    // Check permissions
    if !s.iamService.HasPermission(ctx, "finance:accounts:create") {
        return nil, errors.ErrUnauthorizedAccess
    }
    
    // Check entity access
    if req.EntityID != nil {
        if !s.iamService.CanAccessEntity(ctx, *req.EntityID) {
            return nil, errors.ErrInsufficientPermissions
        }
    }
    
    // Proceed with creation
    return s.accountRepo.Create(ctx, account)
}
```

### 3. Audit Integration
- **Change Tracking**: All modifications logged
- **User Context**: Complete user attribution
- **Immutable Trail**: Audit records cannot be modified

## Scalability Considerations

### 1. Horizontal Scaling
- **Stateless Services**: No server-side state storage
- **Load Balancing**: Request distribution across instances
- **Database Sharding**: Tenant-aware data distribution

### 2. Asynchronous Processing
- **Temporal Workflows**: Long-running business processes
- **Event-Driven Architecture**: Loose coupling between services
- **Background Jobs**: Heavy operations moved to background

### 3. Monitoring and Observability
- **Distributed Tracing**: Request flow tracking
- **Metrics Collection**: Performance and business metrics
- **Structured Logging**: Searchable and analyzable logs

## Testing Architecture

### 1. Unit Testing
- **Domain Logic**: Pure business rule testing
- **Service Layer**: Mocked dependencies
- **Repository Layer**: In-memory test implementations

### 2. Integration Testing
- **Database Integration**: Real database operations
- **Service Integration**: Multi-service workflows
- **API Testing**: End-to-end request/response testing

### 3. Test Utilities
```go
// Test factory for domain objects
func NewTestAccount(options ...func(*domain.Accounts)) *domain.Accounts {
    account := &domain.Accounts{
        ID:          uuid.New(),
        TenantID:    uuid.New(),
        AccountCode: "TEST-001",
        AccountName: "Test Account",
        RootType:    domain.RootTypeAsset,
        IsActive:    true,
    }
    
    for _, option := range options {
        option(account)
    }
    
    return account
}
```

This architecture provides a solid foundation for a scalable, maintainable, and compliant financial accounting system while maintaining clear separation of concerns and following established design patterns.