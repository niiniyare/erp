# Financial Module Architecture Guide

## Overview

The Financial Module implements Clean Architecture principles with Domain-Driven Design (DDD) patterns, providing a robust, scalable, and maintainable financial accounting system. This guide details the architectural layers, design patterns, and implementation strategies used throughout the module.

## Architectural Layers

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

// Example handler method
func (h *FinanceHandler) CreateAccount(c *fiber.Ctx) error {
    // 1. Start tracing
    ctx, span := h.tracer.StartSpan(c.Context(), "finance.CreateAccount")
    defer span.End()
    
    // 2. Parse and validate request
    var req financeDomain.CreateAccountRequest
    if err := h.ValidateRequest(c, &req); err != nil {
        return h.HandleError(c, err)
    }
    
    // 3. Delegate to service layer
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