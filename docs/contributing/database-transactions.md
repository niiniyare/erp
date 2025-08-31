# Database Transactions and Tenant Lifecycle

This guide explains how to properly manage database transactions and tenant context in the ERP system.

## Overview

The ERP system uses PostgreSQL with row-level security (RLS) for multi-tenant isolation. All database operations must be performed within the correct tenant context to ensure data isolation.

## Store Interface

The `Store` interface in `db/sqlc/store.go` provides three main transaction patterns:

### 1. Basic Transactions (`WithTx`)

Use for simple operations that don't require tenant context:

```go
err := store.WithTx(ctx, func(ctx context.Context, s Store) error {
    // Perform multiple database operations
    user, err := s.CreateUser(ctx, params)
    if err != nil {
        return err
    }
    
    _, err = s.UpdateUserProfile(ctx, profileParams)
    return err
})
```

### 2. Tenant-Aware Transactions (`WithTenant`)

Use for operations requiring tenant isolation (most common):

```go
err := store.WithTenant(ctx, tenantID, func(ctx context.Context, s Store) error {
    // All operations automatically isolated to tenantID
    entities, err := s.ListEntities(ctx)
    if err != nil {
        return err
    }
    
    for _, entity := range entities {
        _, err = s.UpdateEntityStatus(ctx, UpdateEntityStatusParams{
            ID:     entity.ID,
            Status: "active",
        })
        if err != nil {
            return err
        }
    }
    
    return nil
})
```

### 3. Manual Transaction Control (`BeginTxWithTenant`)

Use when you need explicit transaction control:

```go
tx, txStore, err := store.BeginTxWithTenant(ctx, tenantID)
if err != nil {
    return err
}
defer tx.Rollback(ctx)

// Perform operations
user, err := txStore.CreateUser(ctx, params)
if err != nil {
    return err
}

// Conditional logic
if user.RequiresApproval {
    _, err = txStore.CreateApprovalRequest(ctx, approvalParams)
    if err != nil {
        return err
    }
}

// Explicit commit
return tx.Commit(ctx)
```

## Tenant Context Management

### Setting Tenant Context

The system uses PostgreSQL session variables to enforce tenant isolation:

- `WithTenant`: Sets `app.current_tenant_id` with transaction scope
- `BeginTxWithTenant`: Uses `set_tenant_context()` database function
- `SetTenantContext`: Sets session-level tenant context

### Tenant ID Sources

Tenant ID should come from:

1. **JWT Claims**: Extract from authenticated user token
2. **Request Headers**: For service-to-service calls
3. **Context Values**: Passed through request chain

Example:

```go
// Extract tenant from JWT claims
tenantID, err := auth.GetTenantIDFromContext(ctx)
if err != nil {
    return fmt.Errorf("failed to get tenant ID: %w", err)
}

// Use in database operations
err = store.WithTenant(ctx, tenantID, func(ctx context.Context, s Store) error {
    return s.CreateEntity(ctx, params)
})
```

## Best Practices

### 1. Always Use Tenant Context

```go
// ❌ Wrong - bypasses tenant isolation
err := store.WithTx(ctx, func(ctx context.Context, s Store) error {
    return s.ListEntities(ctx) // May return data from all tenants
})

// ✅ Correct - enforces tenant isolation
err := store.WithTenant(ctx, tenantID, func(ctx context.Context, s Store) error {
    return s.ListEntities(ctx) // Only returns tenant's data
})
```

### 2. Handle Errors Properly

```go
err := store.WithTenant(ctx, tenantID, func(ctx context.Context, s Store) error {
    entity, err := s.GetEntity(ctx, entityID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return ErrEntityNotFound
        }
        return fmt.Errorf("failed to get entity: %w", err)
    }
    
    // Continue processing...
    return nil
})

if err != nil {
    // Transaction automatically rolled back
    return err
}
```

### 3. Keep Transactions Short

```go
// ❌ Wrong - long-running transaction
err := store.WithTenant(ctx, tenantID, func(ctx context.Context, s Store) error {
    entities, err := s.ListEntities(ctx)
    if err != nil {
        return err
    }
    
    for _, entity := range entities {
        // Long processing that might fail
        processedData := heavyProcessing(entity)
        
        _, err = s.UpdateEntity(ctx, UpdateEntityParams{
            ID:   entity.ID,
            Data: processedData,
        })
        if err != nil {
            return err
        }
    }
    return nil
})

// ✅ Better - separate processing from database operations
entities, err := store.WithTenant(ctx, tenantID, func(ctx context.Context, s Store) error {
    return s.ListEntities(ctx)
})
if err != nil {
    return err
}

for _, entity := range entities {
    processedData := heavyProcessing(entity) // Outside transaction
    
    err := store.WithTenant(ctx, tenantID, func(ctx context.Context, s Store) error {
        _, err := s.UpdateEntity(ctx, UpdateEntityParams{
            ID:   entity.ID,
            Data: processedData,
        })
        return err
    })
    if err != nil {
        return err
    }
}
```

## Common Patterns

### Bulk Operations

```go
func (s *Service) BulkUpdateEntities(ctx context.Context, tenantID uuid.UUID, updates []EntityUpdate) error {
    return s.store.WithTenant(ctx, tenantID, func(ctx context.Context, store Store) error {
        for _, update := range updates {
            _, err := store.UpdateEntity(ctx, UpdateEntityParams{
                ID:     update.ID,
                Status: update.Status,
            })
            if err != nil {
                return fmt.Errorf("failed to update entity %s: %w", update.ID, err)
            }
        }
        return nil
    })
}
```

### Complex Business Logic

```go
func (s *Service) ProcessApproval(ctx context.Context, tenantID uuid.UUID, approvalID uuid.UUID) error {
    return s.store.WithTenant(ctx, tenantID, func(ctx context.Context, store Store) error {
        // Get approval request
        approval, err := store.GetApprovalRequest(ctx, approvalID)
        if err != nil {
            return err
        }
        
        // Update entity status
        _, err = store.UpdateEntityStatus(ctx, UpdateEntityStatusParams{
            ID:     approval.EntityID,
            Status: "approved",
        })
        if err != nil {
            return err
        }
        
        // Mark approval as completed
        _, err = store.CompleteApproval(ctx, approvalID)
        return err
    })
}
```

## Testing Transactions

```go
func TestEntityService_CreateWithApproval(t *testing.T) {
    store := setupTestStore(t)
    service := NewEntityService(store)
    
    tenantID := uuid.New()
    
    // Test successful creation
    entity, err := service.CreateEntityWithApproval(ctx, tenantID, CreateEntityParams{
        Name: "Test Entity",
    })
    
    require.NoError(t, err)
    assert.NotEmpty(t, entity.ID)
    
    // Verify approval was created
    approvals, err := store.WithTenant(ctx, tenantID, func(ctx context.Context, s Store) error {
        return s.ListApprovalRequests(ctx)
    })
    require.NoError(t, err)
    assert.Len(t, approvals, 1)
}
```

## Migration Considerations

When adding new tenant-aware tables:

1. Add RLS policies in migration
2. Update SQLC queries with tenant context
3. Test tenant isolation in integration tests
4. Update this documentation

## Troubleshooting

### Common Issues

1. **Data Leakage**: Always verify tenant context is set
2. **Deadlocks**: Keep transactions short and consistent lock ordering
3. **Connection Pool Exhaustion**: Avoid nested transactions
4. **RLS Violations**: Ensure all queries respect tenant boundaries

### Debug Tips

```go
// Log tenant context in development
_, err = tx.Exec(ctx, "SELECT current_setting('app.current_tenant_id')")
```