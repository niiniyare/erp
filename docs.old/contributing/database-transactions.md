# Awo ERP Database Transactions and Tenant Isolation
## *WithTenantFromCtx Pattern and Context-Based Multi-Tenancy*

*Guide for managing database transactions with proper tenant isolation using WithTenantFromCtx pattern, context propagation, and SQLC integration*

> ** Related Documentation:**
> - `docs/contributing/architecture.md` - System architecture and context patterns
> - `docs/contributing/sqlc-integration.md` - SQLC query patterns and code generation
> - `docs/contributing/01-best-practices.md` - Development guidelines and patterns

## Overview

Awo ERP uses the **WithTenantFromCtx pattern** for database transactions, combining PostgreSQL transaction management with tenant context isolation. The `db.Store` interface provides tenant-aware transaction methods that ensure all operations are properly isolated by setting database session variables within transaction scope.

## Key Architecture Components

1. **Context Propagation**: Tenant ID stored in Go context via `shared.GetTenantID(ctx)`
2. **WithTenantFromCtx Pattern**: Database transactions with tenant session variables  
3. **SQLC Integration**: Type-safe queries with automatic tenant isolation
4. **Repository Layer**: Clean abstractions over database operations

## Store Interface Transaction Patterns

The `db.Store` interface in `db/sqlc/store.go` provides these transaction methods:

```go
type Store interface {
    Querier
    // Tenant context methods
    SetTenantContext(ctx context.Context, tenantID uuid.UUID) error
    WithTenantFromCtx(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, Store) error) error
    BeginTxWithTenant(ctx context.Context, tenantID uuid.UUID) (pgx.Tx, Store, error)
    WithTx(ctx context.Context, fn func(context.Context, Store) error) error
    // Connection management
    Close()
    GetPool() *pgxpool.Pool
}
```

### 1. **WithTenant Pattern (Most Common)**

This is the primary pattern used throughout the codebase for tenant-isolated operations:

```go
// internal/core/finance/repository/accounts.go
func (r *chartOfAccountsRepository) Create(ctx context.Context, account *domain.Accounts) error {
    // Get tenant ID from context
    tenantID, ok := shared.GetTenantID(ctx)
    if !ok {
        return fmt.Errorf("tenant ID not found in context")
    }

    // Use tenant-aware transaction for proper isolation
    return r.store.WithTenantFromCtx(ctx, tenantID, func(ctx context.Context, s db.Store) error {
        // Map domain account to SQLC parameters
        params, err := mapDomainAccountToSQLCCreateDirect(account)
        if err != nil {
            return fmt.Errorf("failed to map create account request: %w", err)
        }

        // Execute SQLC query within tenant context
        sqlcAccount, err := s.CreateAccount(ctx, params)
        if err != nil {
            return r.mapDatabaseError(err, "create_account")
        }

        // Update the account with generated fields
        account.ID = sqlcAccount.ID
        account.TenantID = sqlcAccount.TenantID
        account.CreatedAt = sqlcAccount.CreatedAt
        account.UpdatedAt = sqlcAccount.UpdatedAt

        return nil
    })
}
```

### 2. **Multi-Operation WithTenantFromCtx Example**

Complex operations that need multiple queries in the same tenant context:

```go
// internal/core/finance/repository/transaction.go
func (r *transactionRepository) CreateWithEntries(ctx context.Context, transaction *domain.Transaction) error {
    // Get tenant ID from context
    tenantID, ok := shared.GetTenantID(ctx)
    if !ok {
        return fmt.Errorf("tenant ID not found in context")
    }

    return r.store.WithTenantFromCtx(ctx, tenantID, func(ctx context.Context, s db.Store) error {
        // 1. Create the main transaction
        params := db.CreateTransactionParams{
            EntityID:          transaction.EntityID,
            TransactionNumber: transaction.TransactionNumber,
            TransactionType:   mapDomainTransactionTypeToSQLCEnum(transaction.TransactionType),
            TransactionDate:   transaction.TransactionDate,
            Description:       transaction.Description,
            // ... other fields
        }

        sqlcTransaction, err := s.CreateTransaction(ctx, params)
        if err != nil {
            return r.mapDatabaseError(err, "create_transaction")
        }
        
        transaction.ID = sqlcTransaction.ID
        transaction.TenantID = sqlcTransaction.TenantID

        // 2. Create transaction entries in the same transaction
        for _, entry := range transaction.Entries {
            entryParams := db.CreateTransactionEntryParams{
                TransactionID: transaction.ID,
                AccountID:     entry.AccountID,
                DebitAmount:   decimalToPgNumeric(&entry.DebitAmount),
                CreditAmount:  decimalToPgNumeric(&entry.CreditAmount),
                Description:   entry.Description,
            }
            
            _, err := s.CreateTransactionEntry(ctx, entryParams)
            if err != nil {
                return r.mapDatabaseError(err, "create_transaction_entry")
            }
        }
        
        return nil
    })
}
```

### 3. **Manual Transaction Control (BeginTxWithTenant)**

For advanced scenarios requiring explicit transaction management:

```go
func (r *repository) CreateComplexFinancialOperation(ctx context.Context, req *domain.ComplexRequest) error {
    // Get tenant ID from context
    tenantID, ok := shared.GetTenantID(ctx)
    if !ok {
        return fmt.Errorf("tenant ID not found in context")
    }

    tx, txStore, err := r.store.BeginTxWithTenant(ctx, tenantID)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    // Step 1: Create primary record
    record, err := txStore.CreatePrimaryRecord(ctx, params)
    if err != nil {
        return err
    }

    // Step 2: Conditional logic based on business rules
    if record.RequiresApproval {
        _, err = txStore.CreateApprovalRequest(ctx, db.CreateApprovalParams{
            RecordID:   record.ID,
            ApproverID: req.ApproverID,
        })
        if err != nil {
            return err
        }
    }

    // Step 3: Create related records
    for _, item := range req.RelatedItems {
        _, err = txStore.CreateRelatedRecord(ctx, db.CreateRelatedParams{
            PrimaryID: record.ID,
            Data:      item.Data,
        })
        if err != nil {
            return err
        }
    }

    // Explicit commit
    return tx.Commit(ctx)
}
```

### 4. **Basic Transactions (WithTx)**

For operations that don't require tenant context (rare):

```go
func (r *repository) PerformSystemOperation(ctx context.Context) error {
    return r.store.WithTx(ctx, func(ctx context.Context, s db.Store) error {
        // System-level operations without tenant isolation
        return s.UpdateSystemSettings(ctx, params)
    })
}
```

## How WithTenant Works Internally

The `WithTenant` method in `db/sqlc/store.go` handles the transaction and tenant context setup:

```go
// WithTenant executes a function with tenant context set
func (s *SQLStore) WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, Store) error) error {
    tx, err := s.connPool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    // Set tenant context with transaction scope (true)
    _, err = tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID.String())
    if err != nil {
        return err
    }

    // Create store instance with transaction
    txStore := &SQLStore{
        connPool: s.connPool,
        Queries:  s.Queries.WithTx(tx),
    }

    // Execute function
    if err := fn(ctx, txStore); err != nil {
        return err
    }

    return tx.Commit(ctx)
}
```

## Context Resolution in Repositories

All repositories follow the same pattern for extracting tenant context:

```go
// internal/shared/context.go - Context helper functions
func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
    tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID)
    return tenantID, ok
}

func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
    return context.WithValue(ctx, TenantIDKey, tenantID)
}

// Repository usage pattern
func (r *repository) SomeOperation(ctx context.Context, req *domain.Request) error {
    // ✅ Standard pattern: Get tenant ID from context
    tenantID, ok := shared.GetTenantID(ctx)
    if !ok {
        return fmt.Errorf("tenant ID not found in context")
    }

    // ✅ Use WithTenant for all database operations
    return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
        // All database operations within this block are tenant-isolated
        return s.SomeQuery(ctx, params)
    })
}
```

## Database Session Variables

The tenant context is maintained using PostgreSQL session variables:

```sql
-- Set tenant context (done automatically by WithTenant)
SELECT set_config('app.current_tenant_id', $1, true);

-- Database functions can access current tenant
CREATE OR REPLACE FUNCTION current_tenant_id() 
RETURNS UUID AS $$
BEGIN
    RETURN current_setting('app.current_tenant_id')::UUID;
END;
$$ LANGUAGE plpgsql;

-- Usage in SQLC queries
-- name: CreateAccount :one
INSERT INTO accounts (name, code, tenant_id, created_at, updated_at)
VALUES ($1, $2, current_tenant_id(), NOW(), NOW())
RETURNING *;
```

## Best Practices

### 1. **Always Use WithTenant for Multi-Tenant Data**

```go
// ✅ Correct - ensures tenant isolation
func (r *repository) GetAccount(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
    tenantID, ok := shared.GetTenantID(ctx)
    if !ok {
        return nil, fmt.Errorf("tenant ID not found in context")
    }

    var account *domain.Account
    err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
        sqlcAccount, err := s.GetAccountByID(ctx, id)
        if err != nil {
            if err == db.ErrNoRows {
                return domain.ErrAccountNotFound
            }
            return err
        }
        
        account, err = mapSQLCAccountToDomain(sqlcAccount)
        return err
    })
    
    return account, err
}

// ❌ Wrong - bypasses tenant isolation
func (r *repository) GetAccountUnsafe(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
    return r.store.WithTx(ctx, func(ctx context.Context, s db.Store) error {
        // No tenant context - could return data from any tenant!
        return s.GetAccountByID(ctx, id)
    })
}
```

### 2. **Handle Context Errors Properly**

```go
// ✅ Proper error handling for missing tenant context
func (r *repository) CreateAccount(ctx context.Context, account *domain.Account) error {
    tenantID, ok := shared.GetTenantID(ctx)
    if !ok {
        // This should never happen in normal flow - indicates middleware/context issue
        return fmt.Errorf("tenant ID not found in context - ensure request has proper authentication")
    }

    return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
        // Implementation...
    })
}
```

### 3. **Keep Transactions Focused and Short**

```go
// ✅ Good - focused transaction
func (r *repository) UpdateAccountBalance(ctx context.Context, accountID uuid.UUID, newBalance decimal.Decimal) error {
    tenantID, ok := shared.GetTenantID(ctx)
    if !ok {
        return fmt.Errorf("tenant ID not found in context")
    }

    return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
        return s.UpdateAccountBalance(ctx, db.UpdateAccountBalanceParams{
            ID:      accountID,
            Balance: decimalToPgNumeric(&newBalance),
        })
    })
}

// ❌ Avoid - long-running operations in transaction
func (r *repository) ProcessMonthEndBad(ctx context.Context) error {
    tenantID, ok := shared.GetTenantID(ctx)
    if !ok {
        return fmt.Errorf("tenant ID not found in context")
    }

    return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
        accounts, err := s.ListAccounts(ctx)
        if err != nil {
            return err
        }
        
        // ❌ Heavy processing inside transaction
        for _, account := range accounts {
            report := generateComplexReport(account) // Could take minutes!
            s.CreateReport(ctx, report)
        }
        return nil
    })
}
```

### 4. **Use Temporal for Complex Multi-Step Operations**

```go
// ✅ Better approach for complex operations
func (s *service) ProcessMonthEnd(ctx context.Context) error {
    // Use Temporal workflow for complex, long-running processes
    return s.temporal.ExecuteWorkflow(ctx, workflows.MonthEndProcess, &workflows.MonthEndInput{
        TenantID: shared.GetTenantID(ctx),
    })
}

// The workflow coordinates individual database operations
func MonthEndWorkflow(ctx workflow.Context, input *MonthEndInput) error {
    // Each activity uses WithTenant for focused database operations
    err := workflow.ExecuteActivity(ctx, activities.GenerateReports, input).Get(ctx, nil)
    if err != nil {
        return err
    }
    
    err = workflow.ExecuteActivity(ctx, activities.UpdateBalances, input).Get(ctx, nil)
    return err
}
```

## Testing Database Transactions

### Repository Testing with Test Database

```go
func TestAccountRepository_Create(t *testing.T) {
    // Setup test database
    testDB := setupTestDB(t)
    defer testDB.Close()
    
    store := db.NewStore(testDB)
    repo := NewAccountsRepository(store, tracing.NewNoopTracer())
    
    // Create test context with tenant ID
    ctx := context.Background()
    tenantID := uuid.New()
    ctx = shared.WithTenantID(ctx, tenantID)
    
    // Test data
    account := &domain.Accounts{
        Name:        "Test Account",
        Code:        "1000",
        AccountType: domain.AssetAccount,
    }
    
    // Execute
    err := repo.Create(ctx, account)
    require.NoError(t, err)
    
    // Verify the account was created with correct tenant ID
    assert.Equal(t, tenantID, account.TenantID)
    assert.NotEqual(t, uuid.Nil, account.ID)
    
    // Verify persistence
    retrieved, err := repo.GetByID(ctx, account.ID)
    require.NoError(t, err)
    assert.Equal(t, account.Name, retrieved.Name)
    assert.Equal(t, account.Code, retrieved.Code)
}
```

## Common Patterns Summary

| Pattern | Use Case | Example |
|---------|----------|---------|
| `WithTenant` | Standard multi-tenant operations | Creating accounts, transactions, most business operations |
| `BeginTxWithTenant` | Complex operations requiring transaction control | Multi-step processes with conditional logic |
| `WithTx` | System-level operations | Migrations, system settings (no tenant context needed) |
| `SetTenantContext` | Session-level tenant setting | Middleware setup, connection initialization |

---

 **Next Steps**:
- [SQLC Integration](./sqlc-integration.md) - Type-safe query generation patterns
- [Architecture Overview](./architecture.md) - Overall system design and context flow
- [Best Practices](./01-best-practices.md) - Development guidelines and standards
