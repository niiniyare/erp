# SQLC Integration Guide - Multi-Tenant ERP

This guide explains how we integrate SQLC (SQL Compiler) for type-safe database operations in our multi-tenant ERP system with entity-level data isolation.

## 🎯 What is SQLC?

SQLC generates type-safe Go code from SQL queries, providing:
- Strongly typed Go structs
- Type-safe query methods  
- Compile-time query validation

## 🏗️ Architecture Overview

```
SQL Queries (.sql) → SQLC Generator → Generated Go Code
                                   ↓
Repository Layer ← Store Interface ← Generated Models & Methods
```

## 📁 Project Structure

```
db/
├── migrations/           # Database migrations
├── queries/             # SQLC query files (organized by module)
│   ├── tenant_core.sql     # Core tenant operations
│   ├── user_auth.sql       # User authentication queries  
│   ├── fin_account.sql     # Financial accounts
│   ├── fin_tx.sql          # Financial transactions
│   ├── fin_tx_entry.sql    # Transaction entries
│   ├── inv_item.sql        # Inventory items
│   └── inv_movement.sql    # Inventory movements
├── sqlc.yaml           # SQLC configuration
└── generated/          # Generated Go code (don't edit)
```

## ⚙️ SQLC Configuration

### sqlc.yaml
```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "./db/queries"
    schema: "./db/migrations"
    gen:
      go:
        package: "db"
        out: "./db/generated"
        sql_package: "pgx/v5"
        emit_interface: true
        emit_json_tags: true
        emit_exact_table_names: false
        emit_pointers_for_null_types: true
```

## 🔐 Multi-Tenant & Entity Security Model

Our ERP system uses a **three-level security model**:

### 1. **Tenant Level**: Complete data isolation between tenants
### 2. **Entity Level**: Sub-companies within a tenant (configurable isolation)  
### 3. **User Level**: Role-based access within entities

> **CRITICAL SECURITY REQUIREMENT**: Every query must include both `tenant_id` AND `entity_id` checks where applicable.

## 📝 Query Organization & Naming

### **File Naming Convention**
- Format: `[module]_[feature].sql`
- Examples:
  - `fin_tx.sql` - Financial transactions
  - `fin_tx_entry.sql` - Transaction entries
  - `inv_item.sql` - Inventory items
  - `user_auth.sql` - User authentication
  - `tenant_core.sql` - Core tenant operations

### **Query Naming Convention**
- Use descriptive names: `GetActiveUsersByEntity`, `CreateFinancialTransaction`
- Include operation type: `Create`, `Get`, `List`, `Update`, `SoftDelete`

## 🛡️ Security Patterns (CRITICAL)

### **Required Security Checks in Every Query**

#### ✅ Tenant + Entity Scoped Queries:
```sql
-- name: GetUserByID :one
SELECT * FROM users 
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND entity_id = $2
  AND deleted_at IS NULL;

-- name: ListTransactionsByAccount :many
-- Get all transactions for a specific account within entity
SELECT * FROM financial_transactions
WHERE account_id = $1
  AND tenant_id = current_tenant_id() 
  AND entity_id = $2
  AND deleted_at IS NULL
ORDER BY created_at DESC;
```

#### ✅ Tenant-Only Scoped Queries (Cross-Entity Access):
```sql
-- name: GetTenantSettings :one
-- Tenant-wide settings (no entity restriction)
SELECT * FROM tenant_settings
WHERE tenant_id = current_tenant_id();

-- name: ListEntitiesInTenant :many  
-- List all entities user has access to
SELECT * FROM entities
WHERE tenant_id = current_tenant_id()
  AND id = ANY(@accessible_entity_ids::uuid[])
  AND deleted_at IS NULL;
```

#### ❌ NEVER Do This:
```sql
-- SECURITY VULNERABILITY - Missing tenant/entity checks
SELECT * FROM users WHERE id = $1;
```

### **Soft Delete Pattern (Required)**

**NEVER** perform hard deletes. Always use soft deletes:

```sql
-- name: SoftDeleteUser :exec
UPDATE users 
SET deleted_at = $2, updated_at = $2
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND entity_id = $2
  AND deleted_at IS NULL;
```

## 🏷️ Handling Complex Queries with sqlc.arg

For complex queries with multiple parameters, use `sqlc.arg` and `sqlc.narg` to avoid generic naming like `column1`, `param2`:

### **Using sqlc.arg for Named Parameters**
```sql
-- name: SearchFinancialTransactions :many
-- Search transactions with multiple filters
SELECT 
    ft.*,
    fa.name as account_name,
    fa.code as account_code
FROM financial_transactions ft
JOIN financial_accounts fa ON ft.account_id = fa.id
WHERE ft.tenant_id = current_tenant_id()
  AND ft.entity_id = sqlc.arg('entity_id')
  AND ft.deleted_at IS NULL
  AND fa.deleted_at IS NULL
  AND (sqlc.narg('account_type')::text IS NULL OR fa.account_type = sqlc.arg('account_type'))
  AND (sqlc.narg('date_from')::date IS NULL OR ft.transaction_date >= sqlc.arg('date_from'))
  AND (sqlc.narg('date_to')::date IS NULL OR ft.transaction_date <= sqlc.arg('date_to'))
  AND (sqlc.narg('amount_min')::decimal IS NULL OR ft.amount >= sqlc.arg('amount_min'))
  AND (sqlc.narg('amount_max')::decimal IS NULL OR ft.amount <= sqlc.arg('amount_max'))
  AND (sqlc.narg('search_term')::text IS NULL OR 
       ft.description ILIKE '%' || sqlc.arg('search_term') || '%')
ORDER BY ft.transaction_date DESC, ft.created_at DESC
LIMIT sqlc.arg('limit_count') 
OFFSET sqlc.arg('offset_count');
```

### **Benefits of sqlc.arg**:
- **Clear parameter names** instead of `$1, $2, $3...`
- **Self-documenting** query parameters
- **Easier maintenance** when parameters change
- **Better IDE support** with generated Go code

### **Generated Go Code Example**:
```go
type SearchFinancialTransactionsParams struct {
    EntityID     uuid.UUID  `json:"entity_id"`
    AccountType  *string    `json:"account_type"`
    DateFrom     *time.Time `json:"date_from"`
    DateTo       *time.Time `json:"date_to"`
    AmountMin    *decimal.Decimal `json:"amount_min"`
    AmountMax    *decimal.Decimal `json:"amount_max"`
    SearchTerm   *string    `json:"search_term"`
    LimitCount   int32      `json:"limit_count"`
    OffsetCount  int32      `json:"offset_count"`
}
```

## 📋 Query Examples by Module

### **Financial Module (fin_tx.sql)**
```sql
-- name: CreateFinancialTransaction :one
-- Create a new financial transaction with proper entity isolation
INSERT INTO financial_transactions (
    id, tenant_id, entity_id, account_id, 
    amount, description, transaction_date, 
    reference_number, created_by, created_at, updated_at
) VALUES (
    sqlc.arg('id'), current_tenant_id(), sqlc.arg('entity_id'),
    sqlc.arg('account_id'), sqlc.arg('amount'), sqlc.arg('description'),
    sqlc.arg('transaction_date'), sqlc.arg('reference_number'),
    sqlc.arg('created_by'), sqlc.arg('created_at'), sqlc.arg('updated_at')
) RETURNING *;

-- name: GetTransactionWithEntries :one
-- Get transaction with all its journal entries (complex join)
SELECT 
    ft.id, ft.amount, ft.description, ft.transaction_date,
    json_agg(
        json_build_object(
            'id', fte.id,
            'account_id', fte.account_id, 
            'debit_amount', fte.debit_amount,
            'credit_amount', fte.credit_amount,
            'account_name', fa.name
        )
    ) as entries
FROM financial_transactions ft
JOIN financial_transaction_entries fte ON ft.id = fte.transaction_id
JOIN financial_accounts fa ON fte.account_id = fa.id
WHERE ft.id = sqlc.arg('transaction_id')
  AND ft.tenant_id = current_tenant_id()
  AND ft.entity_id = sqlc.arg('entity_id')
  AND ft.deleted_at IS NULL
GROUP BY ft.id, ft.amount, ft.description, ft.transaction_date;
```

### **User Management (user_auth.sql)**
```sql
-- name: GetUserWithPermissions :one
-- Get user with entity-specific permissions (complex permissions logic)
SELECT 
    u.id, u.email, u.name, u.status,
    json_agg(
        DISTINCT json_build_object(
            'entity_id', ep.entity_id,
            'role', ep.role,
            'permissions', ep.permissions
        )
    ) as entity_permissions
FROM users u
JOIN entity_permissions ep ON u.id = ep.user_id
WHERE u.id = sqlc.arg('user_id')
  AND u.tenant_id = current_tenant_id()
  AND ep.entity_id = sqlc.arg('entity_id')
  AND u.deleted_at IS NULL
  AND ep.deleted_at IS NULL
GROUP BY u.id, u.email, u.name, u.status;
```

### **Inventory Management (inv_item.sql)**
```sql
-- name: GetItemStockByLocation :many
-- Get current stock levels across all locations for an entity
SELECT 
    ii.id, ii.sku, ii.name,
    il.id as location_id, il.name as location_name,
    COALESCE(stock.current_quantity, 0) as current_stock,
    COALESCE(stock.reserved_quantity, 0) as reserved_stock,
    COALESCE(stock.available_quantity, 0) as available_stock
FROM inventory_items ii
CROSS JOIN inventory_locations il
LEFT JOIN (
    -- Subquery to calculate current stock levels
    SELECT 
        item_id, location_id,
        SUM(CASE WHEN movement_type = 'in' THEN quantity ELSE -quantity END) as current_quantity,
        SUM(CASE WHEN is_reserved = true THEN quantity ELSE 0 END) as reserved_quantity,
        SUM(CASE WHEN movement_type = 'in' THEN quantity ELSE -quantity END) - 
        SUM(CASE WHEN is_reserved = true THEN quantity ELSE 0 END) as available_quantity
    FROM inventory_movements
    WHERE tenant_id = current_tenant_id()
      AND entity_id = sqlc.arg('entity_id')
      AND deleted_at IS NULL
    GROUP BY item_id, location_id
) stock ON ii.id = stock.item_id AND il.id = stock.location_id
WHERE ii.tenant_id = current_tenant_id()
  AND ii.entity_id = sqlc.arg('entity_id')
  AND il.entity_id = sqlc.arg('entity_id')
  AND ii.deleted_at IS NULL
  AND il.deleted_at IS NULL
  AND (sqlc.narg('item_category')::text IS NULL OR ii.category = sqlc.arg('item_category'))
ORDER BY ii.name, il.name;
```

## 🏪 Store Interface Pattern

### **Generated Store Interface**
```go
type Store interface {
    Querier
    
    // Custom methods
    SetTenantContext(ctx context.Context, tenantID uuid.UUID) error
    SetEntityContext(ctx context.Context, entityID uuid.UUID) error
    WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, Store) error) error
    WithEntity(ctx context.Context, tenantID, entityID uuid.UUID, fn func(context.Context, Store) error) error
    WithTx(ctx context.Context, fn func(context.Context, Store) error) error
}
```

### **Multi-Tenant & Entity Context**
```go
func (s *store) WithEntity(ctx context.Context, tenantID, entityID uuid.UUID, 
    fn func(context.Context, Store) error) error {
    
    // Set both tenant and entity context
    if err := s.SetTenantContext(ctx, tenantID); err != nil {
        return fmt.Errorf("failed to set tenant context: %w", err)
    }
    
    if err := s.SetEntityContext(ctx, entityID); err != nil {
        return fmt.Errorf("failed to set entity context: %w", err)
    }
    
    return fn(ctx, s)
}
```

## ✅ Best Practices

### **🔒 Security (CRITICAL)**
- **ALWAYS** include `tenant_id = current_tenant_id()` in queries
- **ALWAYS** include `entity_id = $x` for entity-scoped operations  
- **NEVER** use hard deletes - always soft delete with `deleted_at`
- **VALIDATE** tenant and entity context before query execution

### **📝 Query Organization**
- **Group** related queries by module in separate files
- **Use** descriptive naming: `[module]_[feature].sql`
- **Comment** complex queries explaining business logic
- **Use** sqlc.arg/sqlc.narg for complex parameter naming

### **🏷️ Parameter Naming**
```sql
-- ✅ Good: Clear parameter names
WHERE entity_id = sqlc.arg('entity_id')
  AND status = sqlc.arg('user_status')
  AND created_at >= sqlc.arg('date_from')

-- ❌ Avoid: Generic parameter names  
WHERE entity_id = $1 AND status = $2 AND created_at >= $3
```

### **⚡ Performance**
- **Index** on `(tenant_id, entity_id, other_columns)` for multi-tenant queries
- **Use** appropriate query annotations (`:one`, `:many`, `:exec`)
- **Batch** operations when possible
- **Paginate** large result sets

### **🧪 Entity Isolation Testing**
- **Test** cross-entity data access is prevented
- **Verify** entity context is properly enforced
- **Validate** user can only access authorized entities
- **Check** tenant isolation is maintained

---

## 🚀 Quick Start Checklist

- [ ] Configure sqlc.yaml with proper settings
- [ ] Organize queries by module using naming convention
- [ ] Include tenant_id + entity_id in all scoped queries  
- [ ] Use sqlc.arg for complex queries with multiple parameters
- [ ] Comment complex business logic queries
- [ ] Implement soft delete pattern everywhere
- [ ] Test entity isolation thoroughly
- [ ] Add proper indexes for multi-tenant performance

---

📚 **Next Steps**: Migration Patterns | [Testing Guide](./general-testing.md) | Performance Optimization
