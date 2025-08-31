# SQLC Integration Guide

This guide explains how we integrate SQLC (SQL Compiler) for type-safe database operations in our ERP system.

## 🎯 What is SQLC?

SQLC generates type-safe Go code from SQL queries. Instead of writing error-prone string queries, we write SQL and SQLC generates:
- Strongly typed Go structs
- Type-safe query methods
- Compile-time query validation

## 🏗️ SQLC Architecture

```
SQL Queries (.sql) → SQLC Generator → Generated Go Code
                                   ↓
Repository Layer ← Store Interface ← Generated Models & Methods
```

## 📁 Project Structure

```
db/
├── migrations/           # Database migrations
│   ├── 001_create_tenants.up.sql
│   └── 001_create_tenants.down.sql
├── queries/             # SQLC query files
│   ├── tenants.sql
│   ├── users.sql
│   └── permissions.sql
├── sqlc.yaml           # SQLC configuration
└── generated/          # Generated Go code (don't edit)
    ├── models.go       # Database models
    ├── querier.go      # Query interface
    ├── tenants.sql.go  # Generated query methods
    └── store.go        # Store implementation
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
        emit_empty_slices: true
        emit_exact_table_names: false
        emit_pointers_for_null_types: true
        query_parameter_limit: 5
```

### Key Configuration Options:
- **emit_interface**: Generates `Querier` interface
- **emit_json_tags**: Adds JSON tags to structs
- **emit_pointers_for_null_types**: Uses pointers for nullable columns
- **query_parameter_limit**: Limits parameters per query

## 🛠️ Store Interface Pattern

### Generated Store Interface
```go
// db/generated/store.go
type Store interface {
    Querier                    // All SQLC generated methods
    
    // Custom methods we add
    SetTenantContext(ctx context.Context, tenantID uuid.UUID) error
    WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, Store) error) error
    WithTx(ctx context.Context, fn func(context.Context, Store) error) error
    Close()
}

type Querier interface {
    // Generated methods from SQL queries
    CreateTenant(ctx context.Context, arg CreateTenantParams) (Tenant, error)
    GetTenantByID(ctx context.Context, id uuid.UUID) (Tenant, error)
    GetTenantBySubdomain(ctx context.Context, subdomain string) (Tenant, error)
    ListTenants(ctx context.Context, arg ListTenantsParams) ([]Tenant, error)
    UpdateTenant(ctx context.Context, arg UpdateTenantParams) error
    DeleteTenant(ctx context.Context, id uuid.UUID) error
}
```

### Store Implementation
```go
// db/store.go
type store struct {
    *Queries
    db *pgx.Conn
}

func NewStore(db *pgx.Conn) Store {
    return &store{
        Queries: New(db),
        db:      db,
    }
}

// Transaction support
func (s *store) WithTx(ctx context.Context, fn func(context.Context, Store) error) error {
    tx, err := s.db.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)
    
    txStore := &store{
        Queries: s.Queries.WithTx(tx),
        db:      s.db,
    }
    
    if err := fn(ctx, txStore); err != nil {
        return err
    }
    
    return tx.Commit(ctx)
}

// Multi-tenant support
func (s *store) WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, Store) error) error {
    // Set tenant context for RLS (Row Level Security)
    if err := s.SetTenantContext(ctx, tenantID); err != nil {
        return err
    }
    
    return fn(ctx, s)
}
```

## 📝 Writing SQLC Queries

### Query File Structure
```sql
-- db/queries/tenants.sql

-- name: CreateTenant :one
INSERT INTO tenants (
    id, name, slug, email, subdomain, status, 
    timezone, currency_code, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetTenantByID :one
SELECT * FROM tenants 
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetTenantBySubdomain :one
SELECT * FROM tenants 
WHERE subdomain = $1 AND deleted_at IS NULL;

-- name: ListTenants :many
SELECT * FROM tenants
WHERE deleted_at IS NULL
  AND ($1::text IS NULL OR name ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountTenants :one
SELECT COUNT(*) FROM tenants
WHERE deleted_at IS NULL
  AND ($1::text IS NULL OR name ILIKE '%' || $1 || '%');

-- name: UpdateTenant :exec
UPDATE tenants 
SET 
    name = COALESCE($2, name),
    email = COALESCE($3, email),
    subdomain = COALESCE($4, subdomain),
    status = COALESCE($5, status),
    updated_at = $6
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteTenant :exec
UPDATE tenants 
SET deleted_at = $2, updated_at = $2
WHERE id = $1 AND deleted_at IS NULL;
```

## 🔒 Multi-Tenant Query Requirements

### **CRITICAL**: Every Query Must Include Tenant Context

All queries that operate on tenant-scoped data **MUST** include `tenant_id = current_tenant_id()` in the WHERE clause to ensure proper Row-Level Security (RLS) enforcement.

#### ✅ Correct Multi-Tenant Query Pattern:
```sql
-- name: GetUserByID :one
SELECT * FROM users 
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: ListUsersByDepartment :many  
SELECT * FROM users
WHERE department_id = $1
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateUserStatus :exec
UPDATE users
SET status = $2, updated_at = $3
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;
```

#### ❌ Wrong - Missing Tenant Context:
```sql
-- NEVER DO THIS - Security vulnerability!
-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;  -- Missing tenant_id check!
```

### **CRITICAL**: Always Use Soft Delete Pattern

**NEVER perform hard deletes** in the application. Always use soft deletes by setting `deleted_at` timestamp.

#### ✅ Correct Soft Delete Pattern:
```sql  
-- name: SoftDeleteUser :exec
UPDATE users 
SET deleted_at = $2, updated_at = $2
WHERE id = $1 
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: RestoreUser :exec  
UPDATE users
SET deleted_at = NULL, updated_at = $2
WHERE id = $1
  AND tenant_id = current_tenant_id();
```

#### ❌ Wrong - Hard Delete (Forbidden):
```sql
-- NEVER DO THIS - Data loss risk!
-- name: DeleteUser :exec  
DELETE FROM users WHERE id = $1;  -- FORBIDDEN!
```

### Multi-Tenant Query Examples:

#### User Management:
```sql
-- name: CreateUser :one
INSERT INTO users (
    id, tenant_id, email, name, department_id, 
    status, created_at, updated_at
) VALUES (
    $1, current_tenant_id(), $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetActiveUsers :many
SELECT * FROM users
WHERE tenant_id = current_tenant_id()
  AND status = 'active'
  AND deleted_at IS NULL
ORDER BY name ASC;

-- name: CountUsersByStatus :one
SELECT COUNT(*) FROM users  
WHERE tenant_id = current_tenant_id()
  AND status = $1
  AND deleted_at IS NULL;
```

#### Entity Relationships:
```sql
-- name: GetUserWithDepartment :one
SELECT 
    u.*, 
    d.name as department_name
FROM users u
JOIN departments d ON u.department_id = d.id
WHERE u.id = $1
  AND u.tenant_id = current_tenant_id()
  AND d.tenant_id = current_tenant_id()  -- Join tables also need tenant check
  AND u.deleted_at IS NULL
  AND d.deleted_at IS NULL;
```

### Query Annotations Explained:
- `:one` - Returns single row
- `:many` - Returns multiple rows  
- `:exec` - Returns only execution result
- `:execrows` - Returns number of affected rows

## 🔄 Model Conversion Patterns

### Generated SQLC Model
```go
// db/generated/models.go
type Tenant struct {
    ID           uuid.UUID      `json:"id"`
    Name         string         `json:"name"`
    Slug         string         `json:"slug"`
    Email        string         `json:"email"`
    Subdomain    *string        `json:"subdomain"`
    Status       string         `json:"status"`
    Timezone     string         `json:"timezone"`
    CurrencyCode string         `json:"currency_code"`
    Metadata     []byte         `json:"metadata"`     // JSONB
    Settings     []byte         `json:"settings"`     // JSONB
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
    DeletedAt    *time.Time     `json:"deleted_at"`
}
```

### Domain Model
```go
// internal/core/tenant/model.go
type Tenant struct {
    ID           uuid.UUID              `json:"id"`
    Name         string                 `json:"name"`
    Slug         string                 `json:"slug"`
    Email        string                 `json:"email"`
    Subdomain    *string                `json:"subdomain,omitempty"`
    Status       Status                 `json:"status"`
    Timezone     string                 `json:"timezone"`
    CurrencyCode string                 `json:"currency_code"`
    Metadata     map[string]interface{} `json:"metadata,omitempty"`
    Settings     map[string]interface{} `json:"settings,omitempty"`
    CreatedAt    time.Time              `json:"created_at"`
    UpdatedAt    time.Time              `json:"updated_at"`
    DeletedAt    *time.Time             `json:"deleted_at,omitempty"`
}

type Status string

const (
    StatusActive   Status = "active"
    StatusInactive Status = "inactive"
    StatusSuspended Status = "suspended"
)
```

### Conversion Functions
```go
// internal/core/tenant/model.go

// Convert SQLC model to domain model
func FromSQLCTenant(sqlcTenant db.Tenant) (*Tenant, error) {
    // Handle JSONB fields
    var metadata map[string]interface{}
    if len(sqlcTenant.Metadata) > 0 {
        if err := json.Unmarshal(sqlcTenant.Metadata, &metadata); err != nil {
            return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
        }
    }
    
    var settings map[string]interface{}
    if len(sqlcTenant.Settings) > 0 {
        if err := json.Unmarshal(sqlcTenant.Settings, &settings); err != nil {
            return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
        }
    }
    
    return &Tenant{
        ID:           sqlcTenant.ID,
        Name:         sqlcTenant.Name,
        Slug:         sqlcTenant.Slug,
        Email:        sqlcTenant.Email,
        Subdomain:    sqlcTenant.Subdomain,
        Status:       Status(sqlcTenant.Status),
        Timezone:     sqlcTenant.Timezone,
        CurrencyCode: sqlcTenant.CurrencyCode,
        Metadata:     metadata,
        Settings:     settings,
        CreatedAt:    sqlcTenant.CreatedAt,
        UpdatedAt:    sqlcTenant.UpdatedAt,
        DeletedAt:    sqlcTenant.DeletedAt,
    }, nil
}

// Convert domain model to SQLC parameters
func ToSQLCCreateParams(tenant *Tenant) (db.CreateTenantParams, error) {
    // Marshal JSONB fields
    metadataBytes, err := json.Marshal(tenant.Metadata)
    if err != nil {
        return db.CreateTenantParams{}, fmt.Errorf("failed to marshal metadata: %w", err)
    }
    
    settingsBytes, err := json.Marshal(tenant.Settings)
    if err != nil {
        return db.CreateTenantParams{}, fmt.Errorf("failed to marshal settings: %w", err)
    }
    
    return db.CreateTenantParams{
        ID:           tenant.ID,
        Name:         tenant.Name,
        Slug:         tenant.Slug,
        Email:        tenant.Email,
        Subdomain:    tenant.Subdomain,
        Status:       string(tenant.Status),
        Timezone:     tenant.Timezone,
        CurrencyCode: tenant.CurrencyCode,
        Metadata:     metadataBytes,
        Settings:     settingsBytes,
        CreatedAt:    tenant.CreatedAt,
        UpdatedAt:    tenant.UpdatedAt,
    }, nil
}
```

## 🏪 Repository Implementation

### Repository Interface
```go
// internal/core/tenant/repository.go
type Repository interface {
    Create(ctx context.Context, tenant *Tenant) error
    GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
    GetBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
    List(ctx context.Context, filter ListFilter) ([]*Tenant, int, error)
    Update(ctx context.Context, id uuid.UUID, updates UpdateTenantRequest) error
    Delete(ctx context.Context, id uuid.UUID) error
    Exists(ctx context.Context, subdomain string) (bool, error)
}
```

### Repository Implementation
```go
// internal/core/tenant/repository.go
type repository struct {
    store db.Store
}

func NewRepository(store db.Store) Repository {
    return &repository{store: store}
}

func (r *repository) Create(ctx context.Context, tenant *Tenant) error {
    params, err := ToSQLCCreateParams(tenant)
    if err != nil {
        return fmt.Errorf("failed to convert tenant: %w", err)
    }
    
    _, err = r.store.CreateTenant(ctx, params)
    if err != nil {
        return fmt.Errorf("failed to create tenant: %w", err)
    }
    
    return nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
    sqlcTenant, err := r.store.GetTenantByID(ctx, id)
    if err != nil {
        if err.Error() == "no rows in result set" {
            return nil, ErrTenantNotFound
        }
        return nil, fmt.Errorf("failed to get tenant: %w", err)
    }
    
    return FromSQLCTenant(sqlcTenant)
}

func (r *repository) List(ctx context.Context, filter ListFilter) ([]*Tenant, int, error) {
    // Get total count
    var searchTerm *string
    if filter.Search != "" {
        searchTerm = &filter.Search
    }
    
    total, err := r.store.CountTenants(ctx, searchTerm)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
    }
    
    // Get paginated results
    offset := (filter.Page - 1) * filter.Limit
    sqlcTenants, err := r.store.ListTenants(ctx, db.ListTenantsParams{
        Search: searchTerm,
        Limit:  int32(filter.Limit),
        Offset: int32(offset),
    })
    if err != nil {
        return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
    }
    
    // Convert to domain models
    tenants := make([]*Tenant, len(sqlcTenants))
    for i, sqlcTenant := range sqlcTenants {
        tenant, err := FromSQLCTenant(sqlcTenant)
        if err != nil {
            return nil, 0, fmt.Errorf("failed to convert tenant %d: %w", i, err)
        }
        tenants[i] = tenant
    }
    
    return tenants, int(total), nil
}
```

## 🔧 Advanced SQLC Patterns

### Complex Queries with Joins
```sql
-- name: GetTenantWithStats :one
SELECT 
    t.*,
    COUNT(u.id) as user_count,
    COUNT(CASE WHEN u.status = 'active' THEN 1 END) as active_user_count
FROM tenants t
LEFT JOIN users u ON t.id = u.tenant_id AND u.deleted_at IS NULL
WHERE t.id = $1 AND t.deleted_at IS NULL
GROUP BY t.id;
```

### Dynamic Queries with Optional Parameters
```sql
-- name: SearchTenants :many
SELECT * FROM tenants
WHERE deleted_at IS NULL
  AND ($1::text IS NULL OR name ILIKE '%' || $1 || '%')
  AND ($2::text IS NULL OR email ILIKE '%' || $2 || '%')
  AND ($3::text IS NULL OR status = $3)
  AND ($4::timestamptz IS NULL OR created_at >= $4)
ORDER BY 
  CASE WHEN $5 = 'name' THEN name END ASC,
  CASE WHEN $5 = 'created_at' THEN created_at END DESC,
  created_at DESC
LIMIT $6 OFFSET $7;
```

### Batch Operations
```sql
-- name: CreateTenantsInBatch :copyfrom
INSERT INTO tenants (
    id, name, slug, email, subdomain, status,
    timezone, currency_code, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
);

-- name: UpdateTenantsStatus :exec
UPDATE tenants 
SET status = $2, updated_at = $3
WHERE id = ANY($1::uuid[]) AND deleted_at IS NULL;
```

### JSONB Operations
```sql
-- name: UpdateTenantMetadata :exec
UPDATE tenants 
SET 
    metadata = metadata || $2::jsonb,
    updated_at = $3
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetTenantsByMetadataKey :many
SELECT * FROM tenants
WHERE deleted_at IS NULL
  AND metadata ? $1
ORDER BY created_at DESC;
```

## 🏗️ Migration Patterns

### Migration File Structure
```sql
-- db/migrations/001_create_tenants.up.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL,
    subdomain VARCHAR(63) UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    currency_code VARCHAR(3) NOT NULL DEFAULT 'USD',
    metadata JSONB DEFAULT '{}',
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

-- Indexes
CREATE INDEX idx_tenants_slug ON tenants(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_subdomain ON tenants(subdomain) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_status ON tenants(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_created_at ON tenants(created_at);

-- Row Level Security for multi-tenancy
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
```

## 🧪 Testing SQLC Code

### Repository Tests
```go
// internal/core/tenant/repository_test.go
func TestRepository_Create(t *testing.T) {
    // Setup test database
    testDB := setupTestDB(t)
    defer testDB.Close()
    
    store := db.NewStore(testDB)
    repo := NewRepository(store)
    
    // Test data
    tenant := &Tenant{
        ID:           uuid.New(),
        Name:         "Test Tenant",
        Slug:         "test-tenant",
        Email:        "test@example.com",
        Status:       StatusActive,
        Timezone:     "UTC",
        CurrencyCode: "USD",
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }
    
    // Execute
    err := repo.Create(context.Background(), tenant)
    
    // Assert
    require.NoError(t, err)
    
    // Verify creation
    found, err := repo.GetByID(context.Background(), tenant.ID)
    require.NoError(t, err)
    assert.Equal(t, tenant.Name, found.Name)
    assert.Equal(t, tenant.Slug, found.Slug)
}
```

### Mock Store for Unit Tests
```go
// mocks/store.go
type MockStore struct {
    mock.Mock
}

func (m *MockStore) CreateTenant(ctx context.Context, arg db.CreateTenantParams) (db.Tenant, error) {
    args := m.Called(ctx, arg)
    return args.Get(0).(db.Tenant), args.Error(1)
}

func (m *MockStore) GetTenantByID(ctx context.Context, id uuid.UUID) (db.Tenant, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(db.Tenant), args.Error(1)
}

// Usage in tests
func TestService_CreateTenant(t *testing.T) {
    mockStore := new(MockStore)
    repo := NewRepository(mockStore)
    service := NewService(repo, nil)
    
    // Setup expectations
    mockStore.On("CreateTenant", mock.Anything, mock.AnythingOfType("db.CreateTenantParams")).
        Return(db.Tenant{ID: uuid.New()}, nil)
    
    // Test execution
    _, err := service.CreateTenant(context.Background(), CreateTenantRequest{
        Name: "Test",
        Slug: "test",
    })
    
    assert.NoError(t, err)
    mockStore.AssertExpectations(t)
}
```

## ⚡ Performance Tips

### 1. **Efficient Pagination**
```sql
-- Use cursor-based pagination for large datasets
-- name: GetTenantsAfterCursor :many
SELECT * FROM tenants
WHERE deleted_at IS NULL
  AND created_at < $1
ORDER BY created_at DESC
LIMIT $2;
```

### 2. **Selective Fields**
```sql
-- Only select needed fields
-- name: GetTenantBasicInfo :many
SELECT id, name, slug, status FROM tenants
WHERE deleted_at IS NULL;
```

### 3. **Proper Indexing**
```sql
-- Index for common query patterns
CREATE INDEX CONCURRENTLY idx_tenants_composite 
ON tenants(status, created_at DESC) 
WHERE deleted_at IS NULL;
```

### 4. **Batch Operations**
```go
// Use batch inserts for multiple records
func (r *repository) CreateBatch(ctx context.Context, tenants []*Tenant) error {
    rows := make([][]interface{}, len(tenants))
    for i, tenant := range tenants {
        rows[i] = []interface{}{
            tenant.ID, tenant.Name, tenant.Slug,
            // ... other fields
        }
    }
    
    _, err := r.store.CopyFrom(ctx, pgx.Identifier{"tenants"}, 
        []string{"id", "name", "slug"}, // columns
        pgx.CopyFromRows(rows))
    
    return err
}
```

## ✅ Best Practices

### 1. **🔒 Security Requirements (CRITICAL)**
- **ALWAYS** include `tenant_id = current_tenant_id()` in WHERE clauses for tenant-scoped data
- **NEVER** use hard deletes - always soft delete with `deleted_at` timestamp
- **ALWAYS** include `deleted_at IS NULL` in SELECT queries to exclude soft-deleted records
- **VALIDATE** tenant context is properly set before executing queries

### 2. **Query Organization**
- Group related queries in same file
- Use descriptive query names
- Include comments for complex queries
- Document multi-tenant requirements clearly

### 3. **Parameter Handling**
- Use proper PostgreSQL parameter syntax (`$1`, `$2`)
- Handle nullable parameters with `::type IS NULL OR`
- Validate parameters at service layer
- Always validate tenant_id parameter when required

### 4. **Error Handling**
- Convert "no rows" errors to domain errors
- Wrap database errors with context
- Use proper error types
- Handle tenant context errors appropriately

### 5. **Model Conversion**
- Always convert at repository boundaries
- Handle JSONB fields properly
- Validate data during conversion
- Ensure tenant_id is preserved in conversions

### 6. **Performance**
- Use appropriate query annotations (`:one`, `:many`, `:exec`)
- Create indexes for common query patterns
- Use batch operations for multiple records
- Index on `(tenant_id, other_columns)` for multi-tenant queries

### 7. **Multi-Tenant Data Integrity**
- Use foreign key constraints that include tenant_id
- Implement check constraints to prevent cross-tenant data access
- Test tenant isolation thoroughly
- Use database functions like `current_tenant_id()` consistently

---

📚 **Next Steps**:
- [Observability](./observability.md) - Add tracing to database operations
- [Code Examples](./code-examples.md) - See complete implementations
- [Error Handling](./error-handling.md) - Database error patterns
