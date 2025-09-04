# AWO ERP Financial Module - Architecture Guide

**Version**: 1.0  
**Date**: January 2025  
**Status**: Technical Specification  

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Domain-Driven Design](#domain-driven-design)
3. [Data Layer Architecture](#data-layer-architecture)
4. [Service Layer Design](#service-layer-design)
5. [API Layer Implementation](#api-layer-implementation)
6. [Security Architecture](#security-architecture)
7. [Performance Architecture](#performance-architecture)
8. [Integration Patterns](#integration-patterns)
9. [Scalability Design](#scalability-design)
10. [Monitoring & Observability](#monitoring-observability)

---

## Architecture Overview

### **Clean Architecture Implementation**

The AWO ERP Financial Module follows **Clean Architecture** principles with **Hexagonal (Ports & Adapters)** patterns, ensuring complete separation of concerns and maximum testability.

```
┌─────────────────────────────────────────────────────────────────┐
│                        External Systems                        │
│  (Banks, Payment Gateways, Tax Services, ERP Modules)         │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      API Layer (Adapters)                      │
│  @internal/api/handlers/finance_handler.go                     │
│  • Goa-generated REST/gRPC APIs                               │
│  • Request/Response transformation                             │
│  • Validation and error handling                              │
│  • Observability integration                                  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Application Layer (Ports)                    │
│  @internal/core/finance/service/                               │
│  • Business logic orchestration                               │
│  • Use case implementations                                   │
│  • Authorization integration                                  │
│  • Transaction coordination                                   │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Domain Layer (Core)                        │
│  @internal/core/finance/domain/                                │
│  • Financial entities and aggregates                          │
│  • Business rules and validations                             │
│  • Domain events and value objects                            │
│  • Double-entry accounting logic                              │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Infrastructure Layer (Adapters)               │
│  @internal/core/finance/repository/                            │
│  • Data persistence (SQLC)                                    │
│  • External service integration                               │
│  • Caching implementations                                    │
│  • Event publishing                                           │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Database Layer                            │
│  PostgreSQL with Row-Level Security                           │
│  • Multi-tenant data isolation                                │
│  • ACID transaction guarantees                                │
│  • Financial constraint enforcement                           │
│  • Audit trail preservation                                   │
└─────────────────────────────────────────────────────────────────┘
```

### **Key Architectural Principles**

#### **1. Dependency Inversion**
```go
// High-level modules do not depend on low-level modules
// Both depend on abstractions

// ✅ Correct: Service depends on repository interface
type AccountService struct {
    repo AccountRepository // Interface, not implementation
    abac abac.Service      // Interface from another module
}

// ❌ Wrong: Service depends on concrete implementation
type AccountService struct {
    repo *sqlcAccountRepository // Concrete implementation
    db   *sql.DB               // Infrastructure concern
}
```

#### **2. Single Responsibility**
```go
// Each component has one reason to change

// ✅ Account entity: Manages account business rules
type Account struct {
    ID          AccountID
    Code        AccountCode
    Name        string
    AccountType AccountType
    // ... business fields only
}

func (a *Account) Validate() error {
    // Only business validation, no persistence concerns
}

// ✅ AccountRepository: Manages account persistence
type AccountRepository interface {
    Create(ctx context.Context, account *Account) (*Account, error)
    GetByID(ctx context.Context, tenantID tenant.ID, id AccountID) (*Account, error)
    // ... only persistence operations
}
```

#### **3. Open/Closed Principle**
```go
// Open for extension, closed for modification

// ✅ Extensible through interfaces
type PaymentProcessor interface {
    ProcessPayment(ctx context.Context, payment *Payment) error
}

// Can add new payment methods without changing existing code
type CreditCardProcessor struct{}
type ACHProcessor struct{}
type CryptoProcessor struct{} // New addition
```

---

## Domain-Driven Design

### **Aggregate Design**

#### **Account Aggregate**
```go
// @internal/core/finance/domain/account.go

// Account is an aggregate root
type Account struct {
    // Identity
    ID       AccountID
    TenantID tenant.ID
    
    // Value Objects
    Code     AccountCode
    Name     string
    
    // Business State
    AccountType AccountType
    RootType    RootType
    Currency    Currency
    IsGroup     bool
    IsActive    bool
    
    // Hierarchy (Nested Set Model)
    ParentAccountID *AccountID
    Left            *int
    Right           *int
    Depth           int
    
    // Metadata
    CreatedAt time.Time
    UpdatedAt time.Time
    Version   int64 // Optimistic locking
}

// Domain invariants enforced at aggregate boundary
func (a *Account) Validate() error {
    if a.Code == "" {
        return ErrAccountCodeRequired
    }
    if a.IsGroup && a.ParentAccountID == nil {
        return ErrGroupAccountMustHaveParent
    }
    if !a.IsGroup && len(a.Children) > 0 {
        return ErrNonGroupCannotHaveChildren
    }
    return nil
}

// Business operations
func (a *Account) Activate() error {
    if a.IsActive {
        return ErrAccountAlreadyActive
    }
    a.IsActive = true
    a.UpdatedAt = time.Now()
    a.Version++
    return nil
}

func (a *Account) Deactivate() error {
    if !a.IsActive {
        return ErrAccountAlreadyInactive
    }
    if a.HasTransactions() {
        return ErrCannotDeactivateAccountWithTransactions
    }
    a.IsActive = false
    a.UpdatedAt = time.Now()
    a.Version++
    return nil
}
```

#### **Transaction Aggregate**
```go
// @internal/core/finance/domain/transaction.go

// Transaction is an aggregate root ensuring double-entry integrity
type Transaction struct {
    // Identity
    ID       TransactionID
    TenantID tenant.ID
    Number   TransactionNumber
    
    // Business Properties
    Type        TransactionType
    Status      TransactionStatus
    PostingDate time.Time
    DueDate     *time.Time
    
    // Financial Data
    Currency     Currency
    ExchangeRate decimal.Decimal
    TotalAmount  decimal.Decimal
    
    // Entries (Part of aggregate - cannot exist independently)
    Entries []TransactionEntry
    
    // Workflow
    CreatedBy  identity.UserID
    ApprovedBy *identity.UserID
    ApprovedAt *time.Time
    
    // Metadata
    CreatedAt time.Time
    UpdatedAt time.Time
    Version   int64
}

// Aggregate invariants
func (t *Transaction) Validate() error {
    if len(t.Entries) < 2 {
        return ErrTransactionMustHaveAtLeastTwoEntries
    }
    if !t.IsBalanced() {
        return ErrTransactionMustBeBalanced
    }
    if t.TotalAmount.LessThanOrEqual(decimal.Zero) {
        return ErrTransactionAmountMustBePositive
    }
    return nil
}

// Core business rule: Double-entry bookkeeping
func (t *Transaction) IsBalanced() bool {
    totalDebits := decimal.Zero
    totalCredits := decimal.Zero
    
    for _, entry := range t.Entries {
        totalDebits = totalDebits.Add(entry.DebitAmount)
        totalCredits = totalCredits.Add(entry.CreditAmount)
    }
    
    return totalDebits.Equal(totalCredits)
}

// State transitions
func (t *Transaction) Post(approverID identity.UserID) error {
    if t.Status != TransactionStatusSubmitted {
        return ErrCanOnlyPostSubmittedTransactions
    }
    if !t.IsBalanced() {
        return ErrCannotPostUnbalancedTransaction
    }
    
    t.Status = TransactionStatusPosted
    t.ApprovedBy = &approverID
    now := time.Now()
    t.ApprovedAt = &now
    t.UpdatedAt = now
    t.Version++
    
    return nil
}
```

### **Value Objects**

#### **Money Value Object**
```go
// @internal/core/finance/domain/money.go

// Money is immutable value object
type Money struct {
    Amount   decimal.Decimal
    Currency Currency
}

// Value object constructor with validation
func NewMoney(amount decimal.Decimal, currency Currency) (Money, error) {
    if !currency.IsValid() {
        return Money{}, ErrInvalidCurrency
    }
    return Money{
        Amount:   amount,
        Currency: currency,
    }, nil
}

// Value objects are immutable - operations return new instances
func (m Money) Add(other Money) (Money, error) {
    if m.Currency != other.Currency {
        return Money{}, ErrCurrencyMismatch
    }
    return Money{
        Amount:   m.Amount.Add(other.Amount),
        Currency: m.Currency,
    }, nil
}

func (m Money) ConvertTo(targetCurrency Currency, rate ExchangeRate) (Money, error) {
    if rate.FromCurrency != m.Currency || rate.ToCurrency != targetCurrency {
        return Money{}, ErrInvalidExchangeRate
    }
    
    convertedAmount := m.Amount.Mul(rate.Rate)
    return Money{
        Amount:   convertedAmount,
        Currency: targetCurrency,
    }, nil
}

// Value object equality
func (m Money) Equals(other Money) bool {
    return m.Amount.Equal(other.Amount) && m.Currency == other.Currency
}
```

#### **AccountCode Value Object**
```go
// @internal/core/finance/domain/account_code.go

// AccountCode enforces business rules for account coding
type AccountCode string

func NewAccountCode(code string) (AccountCode, error) {
    if code == "" {
        return "", ErrAccountCodeRequired
    }
    if len(code) > 50 {
        return "", ErrAccountCodeTooLong
    }
    if !isValidAccountCodeFormat(code) {
        return "", ErrInvalidAccountCodeFormat
    }
    return AccountCode(code), nil
}

func (ac AccountCode) String() string {
    return string(ac)
}

func (ac AccountCode) IsValid() bool {
    _, err := NewAccountCode(string(ac))
    return err == nil
}

// Business logic for account code validation
func isValidAccountCodeFormat(code string) bool {
    // Account codes must be alphanumeric
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, code)
    return matched
}
```

### **Domain Events**

#### **Event Definitions**
```go
// @internal/core/finance/domain/events.go

// Domain events capture important business occurrences
type DomainEvent interface {
    EventID() string
    EventType() string
    EventTime() time.Time
    AggregateID() string
    TenantID() tenant.ID
}

// Account events
type AccountCreatedEvent struct {
    ID          string
    Type        string
    Time        time.Time
    AccountID   AccountID
    TenantIDVal tenant.ID
    Account     *Account
}

func (e AccountCreatedEvent) EventID() string     { return e.ID }
func (e AccountCreatedEvent) EventType() string   { return e.Type }
func (e AccountCreatedEvent) EventTime() time.Time { return e.Time }
func (e AccountCreatedEvent) AggregateID() string  { return string(e.AccountID) }
func (e AccountCreatedEvent) TenantID() tenant.ID  { return e.TenantIDVal }

// Transaction events
type TransactionPostedEvent struct {
    ID              string
    Type            string
    Time            time.Time
    TransactionID   TransactionID
    TenantIDVal     tenant.ID
    Transaction     *Transaction
    PostedBy        identity.UserID
}

func (e TransactionPostedEvent) EventID() string     { return e.ID }
func (e TransactionPostedEvent) EventType() string   { return e.Type }
func (e TransactionPostedEvent) EventTime() time.Time { return e.Time }
func (e TransactionPostedEvent) AggregateID() string  { return string(e.TransactionID) }
func (e TransactionPostedEvent) TenantID() tenant.ID  { return e.TenantIDVal }
```

#### **Event Publishing**
```go
// @internal/core/finance/service/event_publisher.go

type EventPublisher interface {
    Publish(ctx context.Context, event DomainEvent) error
    PublishBatch(ctx context.Context, events []DomainEvent) error
}

// Service integration with event publishing
func (s *AccountService) CreateAccount(ctx context.Context, cmd CreateAccountCommand) (*Account, error) {
    // ... business logic ...
    
    account, err := s.repo.Create(ctx, account)
    if err != nil {
        return nil, err
    }
    
    // Publish domain event
    event := AccountCreatedEvent{
        ID:          uuid.New().String(),
        Type:        "account.created",
        Time:        time.Now(),
        AccountID:   account.ID,
        TenantIDVal: account.TenantID,
        Account:     account,
    }
    
    if err := s.eventPublisher.Publish(ctx, event); err != nil {
        s.logger.Error("Failed to publish account created event", "error", err)
        // Don't fail the operation for event publishing issues
    }
    
    return account, nil
}
```

---

## Data Layer Architecture

### **Database Schema Design**

#### **Multi-Tenant Row-Level Security**
```sql
-- @db/migration/067_finance_core_tables.up.sql

-- Enable RLS on all financial tables
ALTER TABLE finance_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_transactions ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_transaction_entries ENABLE ROW LEVEL SECURITY;

-- Create tenant isolation policies
CREATE POLICY finance_accounts_tenant_isolation ON finance_accounts
    FOR ALL TO authenticated
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

CREATE POLICY finance_transactions_tenant_isolation ON finance_transactions
    FOR ALL TO authenticated  
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Function to validate tenant context
CREATE OR REPLACE FUNCTION validate_tenant_context() RETURNS TRIGGER AS $$
BEGIN
    IF current_setting('app.current_tenant_id', true) IS NULL THEN
        RAISE EXCEPTION 'Tenant context required for financial operations';
    END IF;
    
    IF NEW.tenant_id != current_setting('app.current_tenant_id')::UUID THEN
        RAISE EXCEPTION 'Tenant ID mismatch in financial operation';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Apply tenant validation to all financial tables
CREATE TRIGGER validate_tenant_context_accounts
    BEFORE INSERT OR UPDATE ON finance_accounts
    FOR EACH ROW EXECUTE FUNCTION validate_tenant_context();

CREATE TRIGGER validate_tenant_context_transactions
    BEFORE INSERT OR UPDATE ON finance_transactions
    FOR EACH ROW EXECUTE FUNCTION validate_tenant_context();
```

#### **Financial Constraints & Validations**
```sql
-- Double-entry bookkeeping constraints
ALTER TABLE finance_transaction_entries 
ADD CONSTRAINT check_debit_credit_exclusive 
CHECK (
    (debit_amount > 0 AND credit_amount = 0) OR 
    (credit_amount > 0 AND debit_amount = 0)
);

-- Transaction balance validation function
CREATE OR REPLACE FUNCTION validate_transaction_balance() RETURNS TRIGGER AS $$
DECLARE
    transaction_total_debits DECIMAL(15,2);
    transaction_total_credits DECIMAL(15,2);
BEGIN
    -- Calculate total debits and credits for the transaction
    SELECT 
        COALESCE(SUM(debit_amount), 0),
        COALESCE(SUM(credit_amount), 0)
    INTO transaction_total_debits, transaction_total_credits
    FROM finance_transaction_entries
    WHERE transaction_id = COALESCE(NEW.transaction_id, OLD.transaction_id);
    
    -- Ensure transaction is balanced
    IF transaction_total_debits != transaction_total_credits THEN
        RAISE EXCEPTION 'Transaction must be balanced: debits=% credits=%', 
            transaction_total_debits, transaction_total_credits;
    END IF;
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Apply balance validation trigger
CREATE TRIGGER validate_transaction_balance_trigger
    AFTER INSERT OR UPDATE OR DELETE ON finance_transaction_entries
    FOR EACH ROW EXECUTE FUNCTION validate_transaction_balance();
```

#### **Performance Optimization**
```sql
-- Strategic indexing for financial operations
CREATE INDEX CONCURRENTLY idx_finance_accounts_tenant_code ON finance_accounts(tenant_id, code);
CREATE INDEX CONCURRENTLY idx_finance_accounts_tenant_parent ON finance_accounts(tenant_id, parent_account_id);
CREATE INDEX CONCURRENTLY idx_finance_accounts_hierarchy ON finance_accounts(tenant_id, lft, rgt);

CREATE INDEX CONCURRENTLY idx_finance_transactions_tenant_date ON finance_transactions(tenant_id, posting_date);
CREATE INDEX CONCURRENTLY idx_finance_transactions_status ON finance_transactions(tenant_id, status);
CREATE INDEX CONCURRENTLY idx_finance_transactions_number ON finance_transactions(tenant_id, number);

CREATE INDEX CONCURRENTLY idx_finance_entries_account_date ON finance_transaction_entries(tenant_id, account_id, (SELECT posting_date FROM finance_transactions WHERE id = transaction_id));
CREATE INDEX CONCURRENTLY idx_finance_entries_transaction ON finance_transaction_entries(transaction_id);

-- Partial indexes for active accounts
CREATE INDEX CONCURRENTLY idx_finance_accounts_active ON finance_accounts(tenant_id, account_type) 
WHERE is_active = true;

-- Covering indexes for balance calculations
CREATE INDEX CONCURRENTLY idx_finance_entries_balance_calc ON finance_transaction_entries(tenant_id, account_id)
INCLUDE (debit_amount, credit_amount, transaction_id);
```

### **SQLC Integration**

#### **Type-Safe Queries**
```sql
-- @db/queries/finance_accounts.sql

-- name: CreateAccount :one
INSERT INTO finance_accounts (
    tenant_id, code, name, parent_account_id, account_type, 
    root_type, currency, is_group, is_active, balance_restriction
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetAccountByID :one
SELECT * FROM finance_accounts 
WHERE tenant_id = $1 AND id = $2;

-- name: GetAccountsByTenant :many
SELECT * FROM finance_accounts 
WHERE tenant_id = $1 
AND ($2::BOOLEAN = FALSE OR is_active = TRUE)
AND ($3::TEXT = '' OR account_type = ANY($3::TEXT[]))
AND ($4::TEXT = '' OR root_type = ANY($4::TEXT[]))
ORDER BY code;

-- name: GetAccountHierarchy :many
WITH RECURSIVE account_tree AS (
    -- Root accounts
    SELECT *, 0 as level, ARRAY[code] as path
    FROM finance_accounts
    WHERE tenant_id = $1 AND parent_account_id IS NULL
    
    UNION ALL
    
    -- Child accounts
    SELECT a.*, t.level + 1, t.path || a.code
    FROM finance_accounts a
    JOIN account_tree t ON a.parent_account_id = t.id
    WHERE a.tenant_id = $1
)
SELECT * FROM account_tree ORDER BY path;

-- name: GetAccountBalance :one
SELECT 
    COALESCE(SUM(debit_amount), 0) - COALESCE(SUM(credit_amount), 0) as balance
FROM finance_transaction_entries fte
JOIN finance_transactions ft ON fte.transaction_id = ft.id
WHERE fte.tenant_id = $1 
AND fte.account_id = $2 
AND ft.status = 'posted'
AND ft.posting_date <= $3;

-- name: GetTrialBalance :many
SELECT 
    fa.id,
    fa.code,
    fa.name,
    fa.account_type,
    fa.root_type,
    COALESCE(SUM(fte.debit_amount), 0) as total_debits,
    COALESCE(SUM(fte.credit_amount), 0) as total_credits,
    COALESCE(SUM(fte.debit_amount), 0) - COALESCE(SUM(fte.credit_amount), 0) as balance
FROM finance_accounts fa
LEFT JOIN finance_transaction_entries fte ON fa.id = fte.account_id
LEFT JOIN finance_transactions ft ON fte.transaction_id = ft.id
WHERE fa.tenant_id = $1 
AND fa.is_active = TRUE
AND (ft.status = 'posted' OR ft.status IS NULL)
AND (ft.posting_date <= $2 OR ft.posting_date IS NULL)
AND ($3::BOOLEAN = TRUE OR COALESCE(SUM(fte.debit_amount), 0) - COALESCE(SUM(fte.credit_amount), 0) != 0)
GROUP BY fa.id, fa.code, fa.name, fa.account_type, fa.root_type
ORDER BY fa.code;
```

#### **Repository Implementation**
```go
// @internal/core/finance/repository/account_repository.go

type accountRepository struct {
    store  *sqlc.Store
    cache  cache.Service
    logger logger.Logger
}

func (r *accountRepository) Create(ctx context.Context, account *domain.Account) (*domain.Account, error) {
    // Ensure tenant context for RLS
    ctx = tenant.WithContext(ctx, account.TenantID)
    
    // Execute within transaction for consistency
    var created *domain.Account
    err := r.store.WithTx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
        // Create account
        params := sqlc.CreateAccountParams{
            TenantID:           uuid.UUID(account.TenantID),
            Code:               string(account.Code),
            Name:               account.Name,
            AccountType:        string(account.AccountType),
            RootType:           string(account.RootType),
            Currency:           string(account.Currency),
            IsGroup:            account.IsGroup,
            IsActive:           account.IsActive,
            BalanceRestriction: account.BalanceRestriction,
        }
        
        if account.ParentAccountID != nil {
            params.ParentAccountID = sql.NullString{
                String: account.ParentAccountID.String(),
                Valid:  true,
            }
        }
        
        result, err := q.CreateAccount(ctx, params)
        if err != nil {
            return fmt.Errorf("failed to create account: %w", err)
        }
        
        created = r.mapToAccount(result)
        
        // Update nested set model if this is a child account
        if account.ParentAccountID != nil {
            if err := r.updateNestedSet(ctx, q, created); err != nil {
                return fmt.Errorf("failed to update nested set: %w", err)
            }
        }
        
        return nil
    })
    
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    r.cacheAccount(ctx, created)
    
    return created, nil
}

func (r *accountRepository) GetByID(ctx context.Context, tenantID tenant.ID, id domain.AccountID) (*domain.Account, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("account:id:%s:%s", tenantID, id)
    if cached, found := r.cache.Get(ctx, cacheKey); found {
        return cached.(*domain.Account), nil
    }
    
    // Ensure tenant context for RLS
    ctx = tenant.WithContext(ctx, tenantID)
    
    result, err := r.store.GetAccountByID(ctx, sqlc.GetAccountByIDParams{
        TenantID: uuid.UUID(tenantID),
        ID:       uuid.UUID(id),
    })
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, domain.ErrAccountNotFound
        }
        return nil, fmt.Errorf("failed to get account: %w", err)
    }
    
    account := r.mapToAccount(result)
    r.cacheAccount(ctx, account)
    
    return account, nil
}

// Nested set model operations for efficient hierarchy queries
func (r *accountRepository) updateNestedSet(ctx context.Context, q *sqlc.Queries, account *domain.Account) error {
    // Implementation of nested set model updates
    // This allows efficient hierarchy queries without recursive CTEs
    
    if account.ParentAccountID == nil {
        // Root account - assign left=1, right=2
        return r.assignNestedSetValues(ctx, q, account.ID, 1, 2, 0)
    }
    
    // Get parent account's right value
    parent, err := q.GetAccountByID(ctx, sqlc.GetAccountByIDParams{
        TenantID: uuid.UUID(account.TenantID),
        ID:       uuid.UUID(*account.ParentAccountID),
    })
    if err != nil {
        return err
    }
    
    // Make space for new node
    parentRight := parent.Rgt.Int32
    
    // Update all nodes with right >= parent.right
    err = q.UpdateNestedSetForInsert(ctx, sqlc.UpdateNestedSetForInsertParams{
        TenantID: uuid.UUID(account.TenantID),
        MinRight: parentRight,
    })
    if err != nil {
        return err
    }
    
    // Assign values to new node
    left := parentRight
    right := parentRight + 1
    depth := parent.Depth + 1
    
    return r.assignNestedSetValues(ctx, q, account.ID, left, right, depth)
}

func (r *accountRepository) mapToAccount(sqlcAccount sqlc.FinanceAccount) *domain.Account {
    account := &domain.Account{
        ID:                 domain.AccountID(sqlcAccount.ID),
        TenantID:           tenant.ID(sqlcAccount.TenantID),
        Code:               domain.AccountCode(sqlcAccount.Code),
        Name:               sqlcAccount.Name,
        AccountType:        domain.AccountType(sqlcAccount.AccountType),
        RootType:           domain.RootType(sqlcAccount.RootType),
        Currency:           domain.Currency(sqlcAccount.Currency),
        IsGroup:            sqlcAccount.IsGroup,
        IsActive:           sqlcAccount.IsActive,
        BalanceRestriction: sqlcAccount.BalanceRestriction,
        Depth:              sqlcAccount.Depth,
        CreatedAt:          sqlcAccount.CreatedAt,
        UpdatedAt:          sqlcAccount.UpdatedAt,
        Version:            sqlcAccount.Version,
    }
    
    if sqlcAccount.ParentAccountID.Valid {
        parentID := domain.AccountID(uuid.MustParse(sqlcAccount.ParentAccountID.String))
        account.ParentAccountID = &parentID
    }
    
    if sqlcAccount.Lft.Valid {
        left := int(sqlcAccount.Lft.Int32)
        account.Left = &left
    }
    
    if sqlcAccount.Rgt.Valid {
        right := int(sqlcAccount.Rgt.Int32)
        account.Right = &right
    }
    
    return account
}
```

---

## Service Layer Design

### **Command Pattern Implementation**

#### **Command Definitions**
```go
// @internal/core/finance/service/commands.go

// Commands represent user intentions
type CreateAccountCommand struct {
    TenantID           tenant.ID
    Code               domain.AccountCode
    Name               string
    AccountType        domain.AccountType
    RootType           domain.RootType
    ParentAccountID    *domain.AccountID
    Currency           domain.Currency
    IsGroup            bool
    BalanceRestriction *string
    UserID             identity.UserID
}

func (cmd CreateAccountCommand) Validate() error {
    if cmd.TenantID == uuid.Nil {
        return errors.BadRequest("tenant ID is required")
    }
    if cmd.Code == "" {
        return errors.BadRequest("account code is required")
    }
    if cmd.Name == "" {
        return errors.BadRequest("account name is required")
    }
    if !cmd.AccountType.IsValid() {
        return errors.BadRequest("invalid account type")
    }
    if !cmd.RootType.IsValid() {
        return errors.BadRequest("invalid root type")
    }
    if cmd.UserID == uuid.Nil {
        return errors.BadRequest("user ID is required")
    }
    return nil
}

type CreateTransactionCommand struct {
    TenantID        tenant.ID
    Type            domain.TransactionType
    PostingDate     time.Time
    DueDate         *time.Time
    Currency        domain.Currency
    ExchangeRate    decimal.Decimal
    ReferenceNumber string
    Description     string
    Entries         []CreateTransactionEntryCommand
    UserID          identity.UserID
}

type CreateTransactionEntryCommand struct {
    AccountID    domain.AccountID
    DebitAmount  decimal.Decimal
    CreditAmount decimal.Decimal
    Description  string
    CostCenterID *uuid.UUID
    ProjectID    *uuid.UUID
}

func (cmd CreateTransactionCommand) Validate() error {
    if err := cmd.validateBasicFields(); err != nil {
        return err
    }
    if err := cmd.validateEntries(); err != nil {
        return err
    }
    if err := cmd.validateBalance(); err != nil {
        return err
    }
    return nil
}

func (cmd CreateTransactionCommand) validateBalance() error {
    totalDebits := decimal.Zero
    totalCredits := decimal.Zero
    
    for _, entry := range cmd.Entries {
        totalDebits = totalDebits.Add(entry.DebitAmount)
        totalCredits = totalCredits.Add(entry.CreditAmount)
    }
    
    if !totalDebits.Equal(totalCredits) {
        return errors.BadRequest("transaction entries must be balanced")
    }
    
    return nil
}
```

#### **Service Implementation**
```go
// @internal/core/finance/service/account_service.go

type AccountService interface {
    CreateAccount(ctx context.Context, cmd CreateAccountCommand) (*domain.Account, error)
    GetAccount(ctx context.Context, tenantID tenant.ID, id domain.AccountID) (*domain.Account, error)
    UpdateAccount(ctx context.Context, cmd UpdateAccountCommand) (*domain.Account, error)
    DeactivateAccount(ctx context.Context, tenantID tenant.ID, id domain.AccountID, userID identity.UserID) error
    ListAccounts(ctx context.Context, tenantID tenant.ID, filters AccountFilters, pagination Pagination) ([]*domain.Account, int64, error)
    GetAccountHierarchy(ctx context.Context, tenantID tenant.ID) ([]*AccountHierarchyNode, error)
}

type accountService struct {
    repo         repository.AccountRepository
    abac         abac.Service
    audit        audit.Service
    eventBus     EventBus
    logger       logger.Logger
    numberGen    NumberGenerator
    validator    AccountValidator
}

func NewAccountService(
    repo repository.AccountRepository,
    abac abac.Service,
    audit audit.Service,
    eventBus EventBus,
    logger logger.Logger,
    numberGen NumberGenerator,
    validator AccountValidator,
) AccountService {
    return &accountService{
        repo:      repo,
        abac:      abac,
        audit:     audit,
        eventBus:  eventBus,
        logger:    logger,
        numberGen: numberGen,
        validator: validator,
    }
}

func (s *accountService) CreateAccount(ctx context.Context, cmd CreateAccountCommand) (*domain.Account, error) {
    // Command validation
    if err := cmd.Validate(); err != nil {
        return nil, err
    }
    
    // Authorization check
    authResult, err := s.abac.Authorize(ctx, abac.AuthorizationRequest{
        Subject:  abac.UserSubject(cmd.UserID),
        Resource: abac.Resource{
            Type: "account",
            Attributes: map[string]interface{}{
                "account_type": string(cmd.AccountType),
                "root_type":    string(cmd.RootType),
                "tenant_id":    cmd.TenantID.String(),
            },
        },
        Action: "create",
        Context: abac.Context{
            TenantID:    cmd.TenantID,
            RequestTime: time.Now(),
            UserID:      cmd.UserID,
        },
    })
    if err != nil {
        return nil, fmt.Errorf("authorization check failed: %w", err)
    }
    if !authResult.Allowed {
        s.audit.LogUnauthorizedAccess(ctx, cmd.TenantID, cmd.UserID, "account", "create")
        return nil, errors.Forbidden("insufficient permissions to create account")
    }
    
    // Business validation
    if err := s.validator.ValidateAccountCreation(ctx, cmd); err != nil {
        return nil, err
    }
    
    // Generate account code if not provided
    if cmd.Code == "" {
        generatedCode, err := s.numberGen.GenerateAccountCode(ctx, cmd.TenantID, cmd.AccountType)
        if err != nil {
            return nil, fmt.Errorf("failed to generate account code: %w", err)
        }
        cmd.Code = domain.AccountCode(generatedCode)
    }
    
    // Build domain entity
    account := &domain.Account{
        TenantID:           cmd.TenantID,
        Code:               cmd.Code,
        Name:               cmd.Name,
        AccountType:        cmd.AccountType,
        RootType:           cmd.RootType,
        ParentAccountID:    cmd.ParentAccountID,
        Currency:           cmd.Currency,
        IsGroup:            cmd.IsGroup,
        IsActive:           true,
        BalanceRestriction: cmd.BalanceRestriction,
        CreatedAt:          time.Now(),
        UpdatedAt:          time.Now(),
        Version:            1,
    }
    
    // Validate domain invariants
    if err := account.Validate(); err != nil {
        return nil, fmt.Errorf("domain validation failed: %w", err)
    }
    
    // Persist account
    created, err := s.repo.Create(ctx, account)
    if err != nil {
        s.logger.Error("Failed to create account",
            "error", err,
            "tenant_id", cmd.TenantID,
            "account_code", cmd.Code)
        return nil, fmt.Errorf("failed to create account: %w", err)
    }
    
    // Audit logging
    s.audit.LogAccountCreation(ctx, created, cmd.UserID)
    
    // Publish domain event
    event := domain.AccountCreatedEvent{
        ID:          uuid.New().String(),
        Type:        "account.created",
        Time:        time.Now(),
        AccountID:   created.ID,
        TenantIDVal: created.TenantID,
        Account:     created,
    }
    
    if err := s.eventBus.Publish(ctx, event); err != nil {
        s.logger.Error("Failed to publish account created event", "error", err)
        // Don't fail the operation for event publishing issues
    }
    
    s.logger.Info("Account created successfully",
        "account_id", created.ID,
        "account_code", created.Code,
        "tenant_id", cmd.TenantID,
        "user_id", cmd.UserID)
    
    return created, nil
}

func (s *accountService) GetAccount(ctx context.Context, tenantID tenant.ID, id domain.AccountID) (*domain.Account, error) {
    // Authorization check
    userID := identity.UserIDFromContext(ctx)
    authResult, err := s.abac.Authorize(ctx, abac.AuthorizationRequest{
        Subject: abac.UserSubject(userID),
        Resource: abac.Resource{
            Type: "account",
            ID:   string(id),
            Attributes: map[string]interface{}{
                "tenant_id": tenantID.String(),
            },
        },
        Action: "read",
        Context: abac.Context{
            TenantID:    tenantID,
            RequestTime: time.Now(),
            UserID:      userID,
        },
    })
    if err != nil {
        return nil, fmt.Errorf("authorization check failed: %w", err)
    }
    if !authResult.Allowed {
        return nil, errors.Forbidden("insufficient permissions to view account")
    }
    
    account, err := s.repo.GetByID(ctx, tenantID, id)
    if err != nil {
        return nil, err
    }
    
    return account, nil
}
```

### **Business Rules Engine**

#### **Validation Framework**
```go
// @internal/core/finance/service/account_validator.go

type AccountValidator interface {
    ValidateAccountCreation(ctx context.Context, cmd CreateAccountCommand) error
    ValidateAccountUpdate(ctx context.Context, cmd UpdateAccountCommand) error
    ValidateAccountDeactivation(ctx context.Context, tenantID tenant.ID, accountID domain.AccountID) error
}

type accountValidator struct {
    repo repository.AccountRepository
    transactionRepo repository.TransactionRepository
}

func (v *accountValidator) ValidateAccountCreation(ctx context.Context, cmd CreateAccountCommand) error {
    // Check for duplicate account code
    existing, err := v.repo.GetByCode(ctx, cmd.TenantID, cmd.Code)
    if err != nil && !errors.IsNotFound(err) {
        return err
    }
    if existing != nil {
        return errors.Conflict("account code already exists")
    }
    
    // Validate parent account if specified
    if cmd.ParentAccountID != nil {
        parent, err := v.repo.GetByID(ctx, cmd.TenantID, *cmd.ParentAccountID)
        if err != nil {
            if errors.IsNotFound(err) {
                return errors.BadRequest("parent account not found")
            }
            return err
        }
        
        if !parent.IsGroup {
            return errors.BadRequest("parent account must be a group account")
        }
        
        if !parent.IsActive {
            return errors.BadRequest("parent account must be active")
        }
        
        // Validate account type compatibility
        if !v.isCompatibleAccountType(parent.AccountType, cmd.AccountType) {
            return errors.BadRequest("account type not compatible with parent")
        }
    }
    
    // Validate account type and root type consistency
    if !v.isValidAccountTypeForRootType(cmd.AccountType, cmd.RootType) {
        return errors.BadRequest("account type not valid for specified root type")
    }
    
    return nil
}

func (v *accountValidator) ValidateAccountDeactivation(ctx context.Context, tenantID tenant.ID, accountID domain.AccountID) error {
    // Check if account has children
    children, err := v.repo.GetChildren(ctx, tenantID, accountID)
    if err != nil {
        return err
    }
    if len(children) > 0 {
        return errors.BadRequest("cannot deactivate account with active child accounts")
    }
    
    // Check if account has transactions
    hasTransactions, err := v.transactionRepo.AccountHasTransactions(ctx, tenantID, accountID)
    if err != nil {
        return err
    }
    if hasTransactions {
        return errors.BadRequest("cannot deactivate account with existing transactions")
    }
    
    return nil
}

func (v *accountValidator) isCompatibleAccountType(parentType, childType domain.AccountType) bool {
    compatibilityMatrix := map[domain.AccountType][]domain.AccountType{
        domain.AccountTypeReceivable: {domain.AccountTypeReceivable},
        domain.AccountTypePayable:    {domain.AccountTypePayable},
        domain.AccountTypeBank:       {domain.AccountTypeBank, domain.AccountTypeCash},
        domain.AccountTypeCash:       {domain.AccountTypeCash},
        domain.AccountTypeIncome:     {domain.AccountTypeIncome},
        domain.AccountTypeExpense:    {domain.AccountTypeExpense},
    }
    
    allowed, exists := compatibilityMatrix[parentType]
    if !exists {
        return false
    }
    
    for _, allowedType := range allowed {
        if childType == allowedType {
            return true
        }
    }
    
    return false
}

func (v *accountValidator) isValidAccountTypeForRootType(accountType domain.AccountType, rootType domain.RootType) bool {
    validCombinations := map[domain.RootType][]domain.AccountType{
        domain.RootTypeAsset: {
            domain.AccountTypeReceivable,
            domain.AccountTypeBank,
            domain.AccountTypeCash,
            domain.AccountTypeStock,
        },
        domain.RootTypeLiability: {
            domain.AccountTypePayable,
            domain.AccountTypeTax,
        },
        domain.RootTypeEquity: {
            domain.AccountTypeEquity,
        },
        domain.RootTypeIncome: {
            domain.AccountTypeIncome,
        },
        domain.RootTypeExpense: {
            domain.AccountTypeExpense,
        },
    }
    
    allowed, exists := validCombinations[rootType]
    if !exists {
        return false
    }
    
    for _, allowedType := range allowed {
        if accountType == allowedType {
            return true
        }
    }
    
    return false
}
```

---

This architectural guide provides the foundation for implementing the AWO ERP Financial Module with enterprise-grade capabilities. The design emphasizes clean separation of concerns, robust domain modeling,  security integration, and high-performance data access patterns.

The remaining sections (API Layer Implementation, Security Architecture, Performance Architecture, etc.) would continue with similarly detailed technical specifications, but I've provided the core architectural foundation that demonstrates the sophisticated engineering approach required for this financial system.