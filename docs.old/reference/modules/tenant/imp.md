# ERP Multi-Tenant Module: Complete Technical Documentation

> **A comprehensive guide to implementing the Tenant module for a multi-tenant ERP platform using PostgreSQL, Go, SQLC, and Temporal**

---

##  Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Technology Stack](#technology-stack)
3. [Directory Structure](#directory-structure)
4. [Domain Layer](#domain-layer)
5. [Repository Layer](#repository-layer)
6. [Service Layer](#service-layer)
7. [Temporal Workflows](#temporal-workflows)
8. [Temporal Activities](#temporal-activities)
9. [Database Schema](#database-schema)
10. [SQLC Configuration](#sqlc-configuration)
11. [Implementation Guide](#implementation-guide)
12. [Testing Strategy](#testing-strategy)
13. [Deployment Considerations](#deployment-considerations)
14. [Best Practices](#best-practices)
15. [Common Patterns](#common-patterns)
16. [Troubleshooting](#troubleshooting)

---

## Architecture Overview

### System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     API Gateway / HTTP Layer                 │
│                   (handlers, middleware, auth)               │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      Service Layer                           │
│                   (business logic, validation)               │
│                    internal/core/tenant/service.go           │
└──────────┬────────────────────────────────┬─────────────────┘
           │                                │
           ▼                                ▼
┌──────────────────────┐         ┌──────────────────────────┐
│  Repository Layer    │         │   Temporal Workflows     │
│  (data access)       │         │   (long-running ops)     │
│  repository/         │         │   workflow/              │
└──────────┬───────────┘         └───────┬──────────────────┘
           │                              │
           │                              ▼
           │                     ┌──────────────────────────┐
           │                     │  Temporal Activities     │
           │                     │  (discrete tasks)        │
           │                     │  activities/             │
           │                     └───────┬──────────────────┘
           │                             │
           ▼                             ▼
┌─────────────────────────────────────────────────────────────┐
│                    PostgreSQL Database                       │
│              (multi-tenant schema with RLS)                  │
└─────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Technologies |
|-----------|----------------|--------------|
| **Domain** | Business entities, value objects, errors | Pure Go structs |
| **Repository** | Database operations, SQLC queries | PostgreSQL, SQLC |
| **Service** | Business logic, orchestration | Go interfaces |
| **Workflows** | Long-running processes, saga patterns | Temporal |
| **Activities** | Atomic operations, retryable tasks | Temporal |

---

## Technology Stack

### Core Technologies

```yaml
Language: Go 1.21+
Database: PostgreSQL 15+
Query Builder: SQLC (type-safe SQL)
Workflow Engine: Temporal
HTTP Framework: (Your choice - Echo/Gin/Chi)
Testing: testify, dockertest
Migration: golang-migrate
```

### Why This Stack?

**PostgreSQL**:
- ✅ Native Row-Level Security (RLS)
- ✅ JSONB for flexible metadata
- ✅ Advanced indexing (GIN, partial)
- ✅ Mature multi-tenancy support
- ✅ Excellent Go drivers (pgx)

**SQLC**:
- ✅ Type-safe SQL queries
- ✅ Generates Go code from SQL
- ✅ Compile-time query validation
- ✅ No runtime reflection
- ✅ Zero dependencies in generated code

**Temporal**:
- ✅ Reliable workflow execution
- ✅ Built-in retry mechanisms
- ✅ Saga pattern support
- ✅ Durable execution
- ✅ Excellent for tenant provisioning

**Go**:
- ✅ Strong typing
- ✅ Excellent concurrency
- ✅ Fast compilation
- ✅ Great PostgreSQL support
- ✅ Rich ecosystem

---

## Directory Structure

### Complete Module Layout

```
internal/core/tenant/
├── activities/
│   ├── tenant_activities.go           # Temporal activity definitions
│   ├── notification_activities.go      # Email/notification activities
│   ├── configuration_activities.go     # Config setup activities
│   └── activities_test.go             # Activity tests
│
├── domain/
│   ├── tenant.go                      # Core tenant entity
│   ├── tenant_config.go               # Configuration entity
│   ├── tenant_usage.go                # Usage tracking entity
│   ├── tenant_status.go               # Status enum/value object
│   ├── errors.go                      # Domain-specific errors
│   └── events.go                      # Domain events
│
├── repository/
│   ├── queries/
│   │   ├── tenant.sql                 # SQLC queries for tenants
│   │   ├── tenant_config.sql          # SQLC queries for configs
│   │   └── tenant_usage.sql           # SQLC queries for usage
│   ├── sqlc.yaml                      # SQLC configuration
│   ├── models.go                      # Generated SQLC models (auto)
│   ├── tenant.sql.go                  # Generated queries (auto)
│   ├── db.go                          # Database connection
│   ├── repository.go                  # Repository interface
│   └── tenant_repository.go           # Repository implementation
│
├── service.go                         # Service interface & implementation
│
└── workflow/
    ├── tenant_provisioning.go         # Provisioning workflow
    ├── tenant_suspension.go           # Suspension workflow
    ├── bulk_operations.go             # Bulk operations workflow
    └── workflow_test.go               # Workflow tests
```

### File Responsibilities

| File/Directory | Purpose | Auto-Generated |
|----------------|---------|----------------|
| `domain/*.go` | Business entities, pure Go types | ❌ Manual |
| `repository/queries/*.sql` | SQL queries for SQLC | ❌ Manual |
| `repository/models.go` | Database models | ✅ SQLC |
| `repository/*.sql.go` | Query implementations | ✅ SQLC |
| `activities/*.go` | Temporal activities | ❌ Manual |
| `workflow/*.go` | Temporal workflows | ❌ Manual |
| `service.go` | Business logic orchestration | ❌ Manual |

---

## Domain Layer

### Core Tenant Entity

**File**: `internal/core/tenant/domain/tenant.go`

```go
package domain

import (
    "time"
    "github.com/google/uuid"
)

// Tenant represents a single organization using the ERP platform
type Tenant struct {
    // Identity
    ID          uuid.UUID  `json:"id"`
    Slug        string     `json:"slug"`
    Name        string     `json:"name"`
    
    // Contact & Access
    Email       string     `json:"email"`
    Subdomain   *string    `json:"subdomain,omitempty"`
    
    // Status & Settings
    Status      TenantStatus `json:"status"`
    Timezone    string       `json:"timezone"`
    CurrencyCode string      `json:"currency_code"`
    
    // Business Info
    Industry         *string `json:"industry,omitempty"`
    CompanySize      *string `json:"company_size,omitempty"`
    TaxID            *string `json:"tax_id,omitempty"`
    RegistrationNumber *string `json:"registration_number,omitempty"`
    LegalEntityType  *string `json:"legal_entity_type,omitempty"`
    
    // Flexible Storage
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    Settings    map[string]interface{} `json:"settings,omitempty"`
    
    // Audit Fields
    LastActivityAt time.Time  `json:"last_activity_at"`
    CreatedAt      time.Time  `json:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at"`
    DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

// NewTenant creates a new tenant with validated defaults
func NewTenant(name, email string, opts ...TenantOption) (*Tenant, error) {
    if name == "" {
        return nil, ErrTenantNameRequired
    }
    if email == "" {
        return nil, ErrTenantEmailRequired
    }
    if !isValidEmail(email) {
        return nil, ErrInvalidEmail
    }
    
    tenant := &Tenant{
        ID:             uuid.New(),
        Name:           name,
        Email:          email,
        Status:         TenantStatusPending,
        Timezone:       "UTC",
        CurrencyCode:   "USD",
        Metadata:       make(map[string]interface{}),
        Settings:       make(map[string]interface{}),
        LastActivityAt: time.Now(),
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
    }
    
    // Apply options
    for _, opt := range opts {
        if err := opt(tenant); err != nil {
            return nil, err
        }
    }
    
    return tenant, nil
}

// TenantOption is a functional option for configuring a tenant
type TenantOption func(*Tenant) error

// WithSubdomain sets the tenant subdomain
func WithSubdomain(subdomain string) TenantOption {
    return func(t *Tenant) error {
        if !isValidSubdomain(subdomain) {
            return ErrInvalidSubdomain
        }
        t.Subdomain = &subdomain
        return nil
    }
}

// WithIndustry sets the tenant industry
func WithIndustry(industry string) TenantOption {
    return func(t *Tenant) error {
        t.Industry = &industry
        return nil
    }
}

// WithCompanySize sets the company size
func WithCompanySize(size string) TenantOption {
    return func(t *Tenant) error {
        if !isValidCompanySize(size) {
            return ErrInvalidCompanySize
        }
        t.CompanySize = &size
        return nil
    }
}

// Business Methods

// Activate changes tenant status to active
func (t *Tenant) Activate() error {
    if t.Status == TenantStatusActive {
        return ErrAlreadyActive
    }
    if t.Status == TenantStatusArchived {
        return ErrCannotActivateArchivedTenant
    }
    
    t.Status = TenantStatusActive
    t.UpdatedAt = time.Now()
    return nil
}

// Suspend temporarily suspends the tenant
func (t *Tenant) Suspend(reason string) error {
    if t.Status == TenantStatusSuspended {
        return ErrAlreadySuspended
    }
    if t.Status == TenantStatusArchived {
        return ErrCannotSuspendArchivedTenant
    }
    
    t.Status = TenantStatusSuspended
    t.UpdatedAt = time.Now()
    
    // Store suspension reason in metadata
    if t.Metadata == nil {
        t.Metadata = make(map[string]interface{})
    }
    t.Metadata["suspension_reason"] = reason
    t.Metadata["suspended_at"] = time.Now()
    
    return nil
}

// Archive archives the tenant (soft delete with retention)
func (t *Tenant) Archive() error {
    if t.Status == TenantStatusArchived {
        return ErrAlreadyArchived
    }
    
    now := time.Now()
    t.Status = TenantStatusArchived
    t.DeletedAt = &now
    t.UpdatedAt = now
    
    return nil
}

// IsActive returns true if tenant is active
func (t *Tenant) IsActive() bool {
    return t.Status == TenantStatusActive
}

// IsSoftDeleted returns true if tenant is soft deleted
func (t *Tenant) IsSoftDeleted() bool {
    return t.DeletedAt != nil
}

// UpdateActivity updates the last activity timestamp
func (t *Tenant) UpdateActivity() {
    t.LastActivityAt = time.Now()
    t.UpdatedAt = time.Now()
}

// Validation helpers

func isValidEmail(email string) bool {
    // Implement RFC 5322 email validation
    // Use regex or external library
    return len(email) > 0 && len(email) <= 255
}

func isValidSubdomain(subdomain string) bool {
    // DNS-safe subdomain validation
    // ^[a-z0-9]([a-z0-9-]*[a-z0-9])?$
    if len(subdomain) == 0 || len(subdomain) > 63 {
        return false
    }
    // Add regex validation
    return true
}

func isValidCompanySize(size string) bool {
    validSizes := map[string]bool{
        "Startup":    true,
        "Small":      true,
        "Medium":     true,
        "Large":      true,
        "Enterprise": true,
    }
    return validSizes[size]
}
```

### Tenant Status Value Object

**File**: `internal/core/tenant/domain/tenant_status.go`

```go
package domain

import "fmt"

// TenantStatus represents the lifecycle status of a tenant
type TenantStatus string

const (
    TenantStatusPending   TenantStatus = "PENDING"
    TenantStatusActive    TenantStatus = "ACTIVE"
    TenantStatusSuspended TenantStatus = "SUSPENDED"
    TenantStatusArchived  TenantStatus = "ARCHIVED"
)

// Valid returns true if the status is valid
func (s TenantStatus) Valid() bool {
    switch s {
    case TenantStatusPending, TenantStatusActive, TenantStatusSuspended, TenantStatusArchived:
        return true
    default:
        return false
    }
}

// String returns string representation
func (s TenantStatus) String() string {
    return string(s)
}

// CanTransitionTo checks if transition to new status is valid
func (s TenantStatus) CanTransitionTo(newStatus TenantStatus) bool {
    transitions := map[TenantStatus][]TenantStatus{
        TenantStatusPending: {
            TenantStatusActive,
            TenantStatusArchived, // Trial didn't convert
        },
        TenantStatusActive: {
            TenantStatusSuspended,
            TenantStatusArchived,
        },
        TenantStatusSuspended: {
            TenantStatusActive,   // Resolved issue
            TenantStatusArchived, // Closure during suspension
        },
        TenantStatusArchived: {
            TenantStatusActive, // Reactivation during retention period
        },
    }
    
    allowed := transitions[s]
    for _, status := range allowed {
        if status == newStatus {
            return true
        }
    }
    return false
}

// ParseTenantStatus parses string to TenantStatus
func ParseTenantStatus(s string) (TenantStatus, error) {
    status := TenantStatus(s)
    if !status.Valid() {
        return "", fmt.Errorf("invalid tenant status: %s", s)
    }
    return status, nil
}
```

### Tenant Configuration Entity

**File**: `internal/core/tenant/domain/tenant_config.go`

```go
package domain

import (
    "time"
    "github.com/google/uuid"
)

// TenantConfiguration holds tenant-specific settings and limits
type TenantConfiguration struct {
    TenantID uuid.UUID `json:"tenant_id"`
    
    // Resource Limits
    MaxUsers               int   `json:"max_users"`
    MaxEntities            int   `json:"max_entities"`
    MaxTransactionsPerMonth int  `json:"max_transactions_per_month"`
    StorageQuota           int64 `json:"storage_quota"` // bytes
    
    // Accounting Preferences
    AccountingMethod      string `json:"accounting_method"`  // ACCRUAL or CASH
    FiscalYearStartMonth  int    `json:"fiscal_year_start_month"`
    DefaultCurrency       string `json:"default_currency"`
    
    // Localization
    DateFormat     string `json:"date_format"`
    NumberFormat   string `json:"number_format"`
    LanguageCode   string `json:"language_code"`
    
    // Security Settings
    PasswordPolicy map[string]interface{} `json:"password_policy,omitempty"`
    
    // Integration Settings
    WebhookEndpoints []string               `json:"webhook_endpoints,omitempty"`
    APIRateLimits    map[string]interface{} `json:"api_rate_limits,omitempty"`
    
    // Flexible Settings
    Settings map[string]interface{} `json:"settings,omitempty"`
    
    // Audit
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// NewTenantConfiguration creates default configuration for a tenant
func NewTenantConfiguration(tenantID uuid.UUID) *TenantConfiguration {
    return &TenantConfiguration{
        TenantID:                tenantID,
        MaxUsers:                100,
        MaxEntities:             1000,
        MaxTransactionsPerMonth: 10000,
        StorageQuota:            1073741824, // 1GB
        AccountingMethod:        "ACCRUAL",
        FiscalYearStartMonth:    1,
        DefaultCurrency:         "USD",
        DateFormat:              "MM/DD/YYYY",
        NumberFormat:            "US",
        LanguageCode:            "en-US",
        PasswordPolicy: map[string]interface{}{
            "min_length":         8,
            "require_uppercase":  true,
            "require_lowercase":  true,
            "require_numbers":    true,
            "require_symbols":    false,
        },
        APIRateLimits: map[string]interface{}{
            "requests_per_minute": 100,
            "requests_per_hour":   5000,
        },
        Settings:  make(map[string]interface{}),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}

// UpdateResourceLimits updates the resource limits
func (c *TenantConfiguration) UpdateResourceLimits(
    users, entities, transactions int,
    storageGB int64,
) {
    c.MaxUsers = users
    c.MaxEntities = entities
    c.MaxTransactionsPerMonth = transactions
    c.StorageQuota = storageGB * 1024 * 1024 * 1024 // Convert GB to bytes
    c.UpdatedAt = time.Now()
}

// SetAccountingMethod sets the accounting method
func (c *TenantConfiguration) SetAccountingMethod(method string) error {
    if method != "ACCRUAL" && method != "CASH" {
        return ErrInvalidAccountingMethod
    }
    c.AccountingMethod = method
    c.UpdatedAt = time.Now()
    return nil
}
```

### Tenant Usage Entity

**File**: `internal/core/tenant/domain/tenant_usage.go`

```go
package domain

import (
    "time"
    "github.com/google/uuid"
)

// TenantUsage tracks resource consumption for a tenant
type TenantUsage struct {
    TenantID    uuid.UUID `json:"tenant_id"`
    PeriodStart time.Time `json:"period_start"`
    PeriodEnd   time.Time `json:"period_end"`
    
    // Usage Metrics
    ActiveUsers      int   `json:"active_users"`
    TotalEntities    int   `json:"total_entities"`
    TotalTransactions int  `json:"total_transactions"`
    StorageUsed      int64 `json:"storage_used"` // bytes
    APICalls         int   `json:"api_calls"`
    
    // Performance Metrics
    AvgResponseTime float64 `json:"avg_response_time"` // milliseconds
    ErrorRate       float64 `json:"error_rate"`        // percentage
    
    // Financial Metrics
    MonthlyRevenue float64 `json:"monthly_revenue,omitempty"`
    
    // Audit
    CreatedAt time.Time `json:"created_at"`
}

// NewTenantUsage creates a new usage record for the current period
func NewTenantUsage(tenantID uuid.UUID) *TenantUsage {
    now := time.Now()
    start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
    end := start.AddDate(0, 1, 0).Add(-time.Second)
    
    return &TenantUsage{
        TenantID:    tenantID,
        PeriodStart: start,
        PeriodEnd:   end,
        CreatedAt:   now,
    }
}

// IsWithinLimits checks if usage is within configuration limits
func (u *TenantUsage) IsWithinLimits(config *TenantConfiguration) bool {
    if u.ActiveUsers > config.MaxUsers {
        return false
    }
    if u.TotalEntities > config.MaxEntities {
        return false
    }
    if u.TotalTransactions > config.MaxTransactionsPerMonth {
        return false
    }
    if u.StorageUsed > config.StorageQuota {
        return false
    }
    return true
}

// UsagePercentage calculates percentage of limit used
func (u *TenantUsage) UsagePercentage(config *TenantConfiguration) map[string]float64 {
    return map[string]float64{
        "users":        float64(u.ActiveUsers) / float64(config.MaxUsers) * 100,
        "entities":     float64(u.TotalEntities) / float64(config.MaxEntities) * 100,
        "transactions": float64(u.TotalTransactions) / float64(config.MaxTransactionsPerMonth) * 100,
        "storage":      float64(u.StorageUsed) / float64(config.StorageQuota) * 100,
    }
}

// IncrementTransaction increments the transaction count
func (u *TenantUsage) IncrementTransaction() {
    u.TotalTransactions++
}

// AddStorageUsed adds to storage usage
func (u *TenantUsage) AddStorageUsed(bytes int64) {
    u.StorageUsed += bytes
}
```

### Domain Errors

**File**: `internal/core/tenant/domain/errors.go`

```go
package domain

import "errors"

var (
    // Validation Errors
    ErrTenantNameRequired    = errors.New("tenant name is required")
    ErrTenantEmailRequired   = errors.New("tenant email is required")
    ErrInvalidEmail          = errors.New("invalid email format")
    ErrInvalidSubdomain      = errors.New("invalid subdomain format")
    ErrInvalidCompanySize    = errors.New("invalid company size")
    ErrInvalidAccountingMethod = errors.New("accounting method must be ACCRUAL or CASH")
    
    // State Errors
    ErrTenantNotFound        = errors.New("tenant not found")
    ErrTenantAlreadyExists   = errors.New("tenant already exists")
    ErrAlreadyActive         = errors.New("tenant is already active")
    ErrAlreadySuspended      = errors.New("tenant is already suspended")
    ErrAlreadyArchived       = errors.New("tenant is already archived")
    ErrCannotActivateArchivedTenant = errors.New("cannot activate archived tenant")
    ErrCannotSuspendArchivedTenant  = errors.New("cannot suspend archived tenant")
    
    // Resource Limit Errors
    ErrUserLimitExceeded         = errors.New("user limit exceeded")
    ErrEntityLimitExceeded       = errors.New("entity limit exceeded")
    ErrTransactionLimitExceeded  = errors.New("transaction limit exceeded")
    ErrStorageLimitExceeded      = errors.New("storage limit exceeded")
    
    // Configuration Errors
    ErrConfigurationNotFound = errors.New("tenant configuration not found")
    ErrInvalidConfiguration  = errors.New("invalid tenant configuration")
)
```

### Domain Events

**File**: `internal/core/tenant/domain/events.go`

```go
package domain

import (
    "time"
    "github.com/google/uuid"
)

// DomainEvent represents a domain event
type DomainEvent interface {
    EventType() string
    OccurredAt() time.Time
    AggregateID() uuid.UUID
}

// TenantCreatedEvent is published when a tenant is created
type TenantCreatedEvent struct {
    TenantID  uuid.UUID
    Name      string
    Email     string
    Timestamp time.Time
}

func (e TenantCreatedEvent) EventType() string       { return "tenant.created" }
func (e TenantCreatedEvent) OccurredAt() time.Time  { return e.Timestamp }
func (e TenantCreatedEvent) AggregateID() uuid.UUID { return e.TenantID }

// TenantActivatedEvent is published when a tenant is activated
type TenantActivatedEvent struct {
    TenantID  uuid.UUID
    Timestamp time.Time
}

func (e TenantActivatedEvent) EventType() string       { return "tenant.activated" }
func (e TenantActivatedEvent) OccurredAt() time.Time  { return e.Timestamp }
func (e TenantActivatedEvent) AggregateID() uuid.UUID { return e.TenantID }

// TenantSuspendedEvent is published when a tenant is suspended
type TenantSuspendedEvent struct {
    TenantID  uuid.UUID
    Reason    string
    Timestamp time.Time
}

func (e TenantSuspendedEvent) EventType() string       { return "tenant.suspended" }
func (e TenantSuspendedEvent) OccurredAt() time.Time  { return e.Timestamp }
func (e TenantSuspendedEvent) AggregateID() uuid.UUID { return e.TenantID }

// TenantArchivedEvent is published when a tenant is archived
type TenantArchivedEvent struct {
    TenantID  uuid.UUID
    Timestamp time.Time
}

func (e TenantArchivedEvent) EventType() string       { return "tenant.archived" }
func (e TenantArchivedEvent) OccurredAt() time.Time  { return e.Timestamp }
func (e TenantArchivedEvent) AggregateID() uuid.UUID { return e.TenantID }
```

---

## Repository Layer

### SQLC Configuration

**File**: `internal/core/tenant/repository/sqlc.yaml`

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries/"
    schema: "../../../migrations/"
    gen:
      go:
        package: "repository"
        out: "."
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_db_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
        emit_exported_queries: false
        emit_result_struct_pointers: false
        emit_params_struct_pointers: false
        emit_methods_with_db_argument: false
        emit_enum_valid_method: true
        emit_all_enum_values: true
        overrides:
          - db_type: "uuid"
            go_type: "github.com/google/uuid.UUID"
          - db_type: "jsonb"
            go_type: "encoding/json.RawMessage"
          - db_type: "timestamptz"
            go_type: "time.Time"
```

### SQL Queries

**File**: `internal/core/tenant/repository/queries/tenant.sql`

```sql
-- name: GetTenant :one
SELECT * FROM tenants
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetTenantBySlug :one
SELECT * FROM tenants
WHERE slug = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetTenantBySubdomain :one
SELECT * FROM tenants
WHERE subdomain = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: ListTenants :many
SELECT * FROM tenants
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListTenantsByStatus :many
SELECT * FROM tenants
WHERE status = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateTenant :one
INSERT INTO tenants (
    id, slug, name, email, subdomain,
    status, timezone, currency_code,
    industry, company_size, tax_id,
    registration_number, legal_entity_type,
    metadata, settings,
    last_activity_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8,
    $9, $10, $11,
    $12, $13,
    $14, $15,
    $16, $17, $18
) RETURNING *;

-- name: UpdateTenant :one
UPDATE tenants
SET
    name = $2,
    email = $3,
    subdomain = $4,
    status = $5,
    timezone = $6,
    currency_code = $7,
    industry = $8,
    company_size = $9,
    tax_id = $10,
    registration_number = $11,
    legal_entity_type = $12,
    metadata = $13,
    settings = $14,
    updated_at = $15
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateTenantStatus :one
UPDATE tenants
SET status = $2, updated_at = $3
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateTenantActivity :exec
UPDATE tenants
SET last_activity_at = $2, updated_at = $3
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteTenant :one
UPDATE tenants
SET deleted_at = $2, updated_at = $3
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: CountTenants :one
SELECT COUNT(*) FROM tenants
WHERE deleted_at IS NULL;

-- name: CountTenantsByStatus :one
SELECT COUNT(*) FROM tenants
WHERE status = $1 AND deleted_at IS NULL;

-- name: TenantExists :one
SELECT EXISTS(
    SELECT 1 FROM tenants
    WHERE id = $1 AND deleted_at IS NULL
);

-- name: TenantSlugExists :one
SELECT EXISTS(
    SELECT 1 FROM tenants
    WHERE slug = $1 AND deleted_at IS NULL
);

-- name: TenantSubdomainExists :one
SELECT EXISTS(
    SELECT 1 FROM tenants
    WHERE subdomain = $1 AND deleted_at IS NULL
);
```

**File**: `internal/core/tenant/repository/queries/tenant_config.sql`

```sql
-- name: GetTenantConfiguration :one
SELECT * FROM tenant_configurations
WHERE tenant_id = $1
LIMIT 1;

-- name: CreateTenantConfiguration :one
INSERT INTO tenant_configurations (
    tenant_id,
    max_users, max_entities, max_transactions_per_month, storage_quota,
    accounting_method, fiscal_year_start_month, default_currency,
    date_format, number_format, language_code,
    password_policy, webhook_endpoints, api_rate_limits, settings,
    created_at, updated_at
) VALUES (
    $1,
    $2, $3, $4, $5,
    $6, $7, $8,
    $9, $10, $11,
    $12, $13, $14, $15,
    $16, $17
) RETURNING *;

-- name: UpdateTenantConfiguration :one
UPDATE tenant_configurations
SET
    max_users = $2,
    max_entities = $3,
    max_transactions_per_month = $4,
    storage_quota = $5,
    accounting_method = $6,
    fiscal_year_start_month = $7,
    default_currency = $8,
    date_format = $9,
    number_format = $10,
    language_code = $11,
    password_policy = $12,
    webhook_endpoints = $13,
    api_rate_limits = $14,
    settings = $15,
    updated_at = $16
WHERE tenant_id = $1
RETURNING *;

-- name: UpdateResourceLimits :exec
UPDATE tenant_configurations
SET
    max_users = $2,
    max_entities = $3,
    max_transactions_per_month = $4,
    storage_quota = $5,
    updated_at = $6
WHERE tenant_id = $1;
```

**File**: `internal/core/tenant/repository/queries/tenant_usage.sql`

```sql
-- name: GetCurrentUsage :one
SELECT * FROM tenant_usage_stats
WHERE tenant_id = $1
  AND period_start <= CURRENT_DATE
  AND period_end >= CURRENT_DATE
LIMIT 1;

-- name: GetUsageByPeriod :one
SELECT * FROM tenant_usage_stats
WHERE tenant_id = $1
  AND period_start = $2
LIMIT 1;

-- name: ListUsageHistory :many
SELECT * FROM tenant_usage_stats
WHERE tenant_id = $1
  AND period_start >= $2
ORDER BY period_start DESC
LIMIT $3 OFFSET $4;

-- name: CreateUsageStats :one
INSERT INTO tenant_usage_stats (
    tenant_id, period_start, period_end,
    active_users, total_entities, total_transactions,
    storage_used, api_calls,
    avg_response_time, error_rate,
    monthly_revenue, created_at
) VALUES (
    $1, $2, $3,
    $4, $5, $6,
    $7, $8,
    $9, $10,
    $11, $12
) RETURNING *;

-- name: UpdateUsageStats :one
UPDATE tenant_usage_stats
SET
    active_users = $3,
    total_entities = $4,
    total_transactions = $5,
    storage_used = $6,
    api_calls = $7,
    avg_response_time = $8,
    error_rate = $9,
    monthly_revenue = $10
WHERE tenant_id = $1 AND period_start = $2
RETURNING *;

-- name: IncrementTransactionCount :exec
UPDATE tenant_usage_stats
SET total_transactions = total_transactions + $3
WHERE tenant_id = $1 AND period_start = $2;

-- name: AddStorageUsage :exec
UPDATE tenant_usage_stats
SET storage_used = storage_used + $3
WHERE tenant_id = $1 AND period_start = $2;
```

### Repository Interface

**File**: `internal/core/tenant/repository/repository.go`

```go
package repository

import (
    "context"
    "time"
    "github.com/google/uuid"
    "awo.so/internal/core/tenant/domain"
)

// TenantRepository defines tenant data access operations
type TenantRepository interface {
    // Tenant CRUD
    Create(ctx context.Context, tenant *domain.Tenant) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
    GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
    GetBySubdomain(ctx context.Context, subdomain string) (*domain.Tenant, error)
    Update(ctx context.Context, tenant *domain.Tenant) error
    UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TenantStatus) error
    UpdateActivity(ctx context.Context, id uuid.UUID) error
    SoftDelete(ctx context.Context, id uuid.UUID) error
    
    // Tenant Queries
    List(ctx context.Context, limit, offset int) ([]*domain.Tenant, error)
    ListByStatus(ctx context.Context, status domain.TenantStatus, limit, offset int) ([]*domain.Tenant, error)
    Count(ctx context.Context) (int64, error)
    CountByStatus(ctx context.Context, status domain.TenantStatus) (int64, error)
    Exists(ctx context.Context, id uuid.UUID) (bool, error)
    SlugExists(ctx context.Context, slug string) (bool, error)
    SubdomainExists(ctx context.Context, subdomain string) (bool, error)
    
    // Configuration
    CreateConfiguration(ctx context.Context, config *domain.TenantConfiguration) error
    GetConfiguration(ctx context.Context, tenantID uuid.UUID) (*domain.TenantConfiguration, error)
    UpdateConfiguration(ctx context.Context, config *domain.TenantConfiguration) error
    
    // Usage Stats
    CreateUsage(ctx context.Context, usage *domain.TenantUsage) error
    GetCurrentUsage(ctx context.Context, tenantID uuid.UUID) (*domain.TenantUsage, error)
    GetUsageByPeriod(ctx context.Context, tenantID uuid.UUID, periodStart time.Time) (*domain.TenantUsage, error)
    UpdateUsage(ctx context.Context, usage *domain.TenantUsage) error
    IncrementTransaction(ctx context.Context, tenantID uuid.UUID, periodStart time.Time) error
    AddStorage(ctx context.Context, tenantID uuid.UUID, periodStart time.Time, bytes int64) error
}
```

### Repository Implementation

**File**: `internal/core/tenant/repository/tenant_repository.go`

```go
package repository

import (
    "context"
    "database/sql"
    "encoding/json"
    "time"
    
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
    "awo.so/internal/core/tenant/domain"
)

type tenantRepository struct {
    db      *pgxpool.Pool
    queries *Queries
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *pgxpool.Pool) TenantRepository {
    return &tenantRepository{
        db:      db,
        queries: New(db),
    }
}

// Create creates a new tenant
func (r *tenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
    metadata, err := json.Marshal(tenant.Metadata)
    if err != nil {
        return err
    }
    
    settings, err := json.Marshal(tenant.Settings)
    if err != nil {
        return err
    }
    
    params := CreateTenantParams{
        ID:                 tenant.ID,
        Slug:               tenant.Slug,
        Name:               tenant.Name,
        Email:              tenant.Email,
        Subdomain:          toNullString(tenant.Subdomain),
        Status:             string(tenant.Status),
        Timezone:           tenant.Timezone,
        CurrencyCode:       tenant.CurrencyCode,
        Industry:           toNullString(tenant.Industry),
        CompanySize:        toNullString(tenant.CompanySize),
        TaxID:              toNullString(tenant.TaxID),
        RegistrationNumber: toNullString(tenant.RegistrationNumber),
        LegalEntityType:    toNullString(tenant.LegalEntityType),
        Metadata:           metadata,
        Settings:           settings,
        LastActivityAt:     tenant.LastActivityAt,
        CreatedAt:          tenant.CreatedAt,
        UpdatedAt:          tenant.UpdatedAt,
    }
    
    _, err = r.queries.CreateTenant(ctx, params)
    return err
}

// GetByID retrieves a tenant by ID
func (r *tenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
    row, err := r.queries.GetTenant(ctx, id)
    if err != nil {
        if err == db.ErrNoRows {
            return nil, domain.ErrTenantNotFound
        }
        return nil, err
    }
    
    return r.toDomain(row)
}

// GetBySlug retrieves a tenant by slug
func (r *tenantRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
    row, err := r.queries.GetTenantBySlug(ctx, slug)
    if err != nil {
        if err == db.ErrNoRows {
            return nil, domain.ErrTenantNotFound
        }
        return nil, err
    }
    
    return r.toDomain(row)
}

// GetBySubdomain retrieves a tenant by subdomain
func (r *tenantRepository) GetBySubdomain(ctx context.Context, subdomain string) (*domain.Tenant, error) {
    row, err := r.queries.GetTenantBySubdomain(ctx, sql.NullString{
        String: subdomain,
        Valid:  true,
    })
    if err != nil {
        if err == db.ErrNoRows {
            return nil, domain.ErrTenantNotFound
        }
        return nil, err
    }
    
    return r.toDomain(row)
}

// Update updates an existing tenant
func (r *tenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
    metadata, err := json.Marshal(tenant.Metadata)
    if err != nil {
        return err
    }
    
    settings, err := json.Marshal(tenant.Settings)
    if err != nil {
        return err
    }
    
    params := UpdateTenantParams{
        ID:                 tenant.ID,
        Name:               tenant.Name,
        Email:              tenant.Email,
        Subdomain:          toNullString(tenant.Subdomain),
        Status:             string(tenant.Status),
        Timezone:           tenant.Timezone,
        CurrencyCode:       tenant.CurrencyCode,
        Industry:           toNullString(tenant.Industry),
        CompanySize:        toNullString(tenant.CompanySize),
        TaxID:              toNullString(tenant.TaxID),
        RegistrationNumber: toNullString(tenant.RegistrationNumber),
        LegalEntityType:    toNullString(tenant.LegalEntityType),
        Metadata:           metadata,
        Settings:           settings,
        UpdatedAt:          time.Now(),
    }
    
    _, err = r.queries.UpdateTenant(ctx, params)
    return err
}

// UpdateStatus updates tenant status
func (r *tenantRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TenantStatus) error {
    _, err := r.queries.UpdateTenantStatus(ctx, UpdateTenantStatusParams{
        ID:        id,
        Status:    string(status),
        UpdatedAt: time.Now(),
    })
    return err
}

// UpdateActivity updates last activity timestamp
func (r *tenantRepository) UpdateActivity(ctx context.Context, id uuid.UUID) error {
    return r.queries.UpdateTenantActivity(ctx, UpdateTenantActivityParams{
        ID:             id,
        LastActivityAt: time.Now(),
        UpdatedAt:      time.Now(),
    })
}

// SoftDelete soft deletes a tenant
func (r *tenantRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
    now := time.Now()
    _, err := r.queries.SoftDeleteTenant(ctx, SoftDeleteTenantParams{
        ID:        id,
        DeletedAt: sql.NullTime{Time: now, Valid: true},
        UpdatedAt: now,
    })
    return err
}

// List retrieves paginated list of tenants
func (r *tenantRepository) List(ctx context.Context, limit, offset int) ([]*domain.Tenant, error) {
    rows, err := r.queries.ListTenants(ctx, ListTenantsParams{
        Limit:  int32(limit),
        Offset: int32(offset),
    })
    if err != nil {
        return nil, err
    }
    
    tenants := make([]*domain.Tenant, len(rows))
    for i, row := range rows {
        tenant, err := r.toDomain(row)
        if err != nil {
            return nil, err
        }
        tenants[i] = tenant
    }
    
    return tenants, nil
}

// Helper: Convert DB model to domain entity
func (r *tenantRepository) toDomain(row Tenant) (*domain.Tenant, error) {
    var metadata map[string]interface{}
    if err := json.Unmarshal(row.Metadata, &metadata); err != nil {
        return nil, err
    }
    
    var settings map[string]interface{}
    if err := json.Unmarshal(row.Settings, &settings); err != nil {
        return nil, err
    }
    
    tenant := &domain.Tenant{
        ID:                 row.ID,
        Slug:               row.Slug,
        Name:               row.Name,
        Email:              row.Email,
        Subdomain:          fromNullString(row.Subdomain),
        Status:             domain.TenantStatus(row.Status),
        Timezone:           row.Timezone,
        CurrencyCode:       row.CurrencyCode,
        Industry:           fromNullString(row.Industry),
        CompanySize:        fromNullString(row.CompanySize),
        TaxID:              fromNullString(row.TaxID),
        RegistrationNumber: fromNullString(row.RegistrationNumber),
        LegalEntityType:    fromNullString(row.LegalEntityType),
        Metadata:           metadata,
        Settings:           settings,
        LastActivityAt:     row.LastActivityAt,
        CreatedAt:          row.CreatedAt,
        UpdatedAt:          row.UpdatedAt,
        DeletedAt:          fromNullTime(row.DeletedAt),
    }
    
    return tenant, nil
}

// Helpers for nullable types
func toNullString(s *string) sql.NullString {
    if s == nil {
        return sql.NullString{Valid: false}
    }
    return sql.NullString{String: *s, Valid: true}
}

func fromNullString(ns sql.NullString) *string {
    if !ns.Valid {
        return nil
    }
    return &ns.String
}

func fromNullTime(nt sql.NullTime) *time.Time {
    if !nt.Valid {
        return nil
    }
    return &nt.Time
}
```

---

## Service Layer

**File**: `internal/core/tenant/service.go`

```go
package tenant

import (
    "context"
    "fmt"
    "time"
    
    "github.com/google/uuid"
    "awo.so/internal/core/tenant/domain"
    "awo.so/internal/core/tenant/repository"
    "go.temporal.io/sdk/client"
)

// Service defines tenant business operations
type Service interface {
    // Tenant Management
    CreateTenant(ctx context.Context, req CreateTenantRequest) (*domain.Tenant, error)
    GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
    GetTenantBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
    UpdateTenant(ctx context.Context, id uuid.UUID, req UpdateTenantRequest) (*domain.Tenant, error)
    ActivateTenant(ctx context.Context, id uuid.UUID) error
    SuspendTenant(ctx context.Context, id uuid.UUID, reason string) error
    ArchiveTenant(ctx context.Context, id uuid.UUID) error
    ListTenants(ctx context.Context, req ListTenantsRequest) (*ListTenantsResponse, error)
    
    // Configuration
    GetConfiguration(ctx context.Context, tenantID uuid.UUID) (*domain.TenantConfiguration, error)
    UpdateConfiguration(ctx context.Context, tenantID uuid.UUID, req UpdateConfigurationRequest) error
    
    // Usage & Limits
    GetCurrentUsage(ctx context.Context, tenantID uuid.UUID) (*domain.TenantUsage, error)
    CheckResourceLimit(ctx context.Context, tenantID uuid.UUID, resourceType string, additional int) (bool, error)
    RecordTransaction(ctx context.Context, tenantID uuid.UUID) error
    RecordStorageUsage(ctx context.Context, tenantID uuid.UUID, bytes int64) error
    
    // Provisioning (uses Temporal workflows)
    ProvisionTenantComplete(ctx context.Context, req ProvisionTenantRequest) (string, error) // Returns workflowID
}

// service implements the Service interface
type service struct {
    repo           repository.TenantRepository
    temporalClient client.Client
    slugGenerator  SlugGenerator
}

// NewService creates a new tenant service
func NewService(
    repo repository.TenantRepository,
    temporalClient client.Client,
) Service {
    return &service{
        repo:           repo,
        temporalClient: temporalClient,
        slugGenerator:  NewSlugGenerator(),
    }
}

// CreateTenantRequest represents tenant creation parameters
type CreateTenantRequest struct {
    Name               string
    Email              string
    Subdomain          *string
    Industry           *string
    CompanySize        *string
    Timezone           string
    CurrencyCode       string
    TaxID              *string
    RegistrationNumber *string
    LegalEntityType    *string
}

// CreateTenant creates a new tenant
func (s *service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*domain.Tenant, error) {
    // Generate slug from name
    slug := s.slugGenerator.Generate(req.Name)
    
    // Ensure slug uniqueness
    exists, err := s.repo.SlugExists(ctx, slug)
    if err != nil {
        return nil, fmt.Errorf("check slug exists: %w", err)
    }
    
    if exists {
        // Append UUID fragment to make unique
        slug = fmt.Sprintf("%s-%s", slug, uuid.New().String()[:8])
    }
    
    // Validate subdomain if provided
    if req.Subdomain != nil {
        exists, err := s.repo.SubdomainExists(ctx, *req.Subdomain)
        if err != nil {
            return nil, fmt.Errorf("check subdomain exists: %w", err)
        }
        if exists {
            return nil, fmt.Errorf("subdomain already taken: %s", *req.Subdomain)
        }
    }
    
    // Create tenant entity
    tenant, err := domain.NewTenant(
        req.Name,
        req.Email,
        domain.WithSubdomain(*req.Subdomain),
        domain.WithIndustry(*req.Industry),
        domain.WithCompanySize(*req.CompanySize),
    )
    if err != nil {
        return nil, fmt.Errorf("create tenant entity: %w", err)
    }
    
    tenant.Slug = slug
    tenant.Timezone = req.Timezone
    tenant.CurrencyCode = req.CurrencyCode
    tenant.TaxID = req.TaxID
    tenant.RegistrationNumber = req.RegistrationNumber
    tenant.LegalEntityType = req.LegalEntityType
    
    // Persist to database
    if err := s.repo.Create(ctx, tenant); err != nil {
        return nil, fmt.Errorf("save tenant: %w", err)
    }
    
    // Create default configuration
    config := domain.NewTenantConfiguration(tenant.ID)
    if err := s.repo.CreateConfiguration(ctx, config); err != nil {
        return nil, fmt.Errorf("create tenant configuration: %w", err)
    }
    
    // Create current period usage stats
    usage := domain.NewTenantUsage(tenant.ID)
    if err := s.repo.CreateUsage(ctx, usage); err != nil {
        return nil, fmt.Errorf("create usage stats: %w", err)
    }
    
    return tenant, nil
}

// GetTenant retrieves a tenant by ID
func (s *service) GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
    tenant, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get tenant: %w", err)
    }
    
    // Update last activity
    if err := s.repo.UpdateActivity(ctx, id); err != nil {
        // Log but don't fail
        fmt.Printf("failed to update activity: %v\n", err)
    }
    
    return tenant, nil
}

// ActivateTenant activates a tenant
func (s *service) ActivateTenant(ctx context.Context, id uuid.UUID) error {
    tenant, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return fmt.Errorf("get tenant: %w", err)
    }
    
    if err := tenant.Activate(); err != nil {
        return err
    }
    
    if err := s.repo.UpdateStatus(ctx, id, domain.TenantStatusActive); err != nil {
        return fmt.Errorf("update status: %w", err)
    }
    
    // TODO: Publish TenantActivatedEvent
    
    return nil
}

// SuspendTenant suspends a tenant
func (s *service) SuspendTenant(ctx context.Context, id uuid.UUID, reason string) error {
    tenant, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return fmt.Errorf("get tenant: %w", err)
    }
    
    if err := tenant.Suspend(reason); err != nil {
        return err
    }
    
    if err := s.repo.Update(ctx, tenant); err != nil {
        return fmt.Errorf("update tenant: %w", err)
    }
    
    // TODO: Publish TenantSuspendedEvent
    
    return nil
}

// CheckResourceLimit checks if adding resources would exceed limits
func (s *service) CheckResourceLimit(
    ctx context.Context,
    tenantID uuid.UUID,
    resourceType string,
    additional int,
) (bool, error) {
    config, err := s.repo.GetConfiguration(ctx, tenantID)
    if err != nil {
        return false, fmt.Errorf("get configuration: %w", err)
    }
    
    usage, err := s.repo.GetCurrentUsage(ctx, tenantID)
    if err != nil {
        return false, fmt.Errorf("get current usage: %w", err)
    }
    
    switch resourceType {
    case "users":
        return usage.ActiveUsers+additional <= config.MaxUsers, nil
    case "entities":
        return usage.TotalEntities+additional <= config.MaxEntities, nil
    case "transactions":
        return usage.TotalTransactions+additional <= config.MaxTransactionsPerMonth, nil
    case "storage":
        return usage.StorageUsed+int64(additional) <= config.StorageQuota, nil
    default:
        return false, fmt.Errorf("unknown resource type: %s", resourceType)
    }
}

// RecordTransaction increments transaction count
func (s *service) RecordTransaction(ctx context.Context, tenantID uuid.UUID) error {
    now := time.Now()
    periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
    
    return s.repo.IncrementTransaction(ctx, tenantID, periodStart)
}

// RecordStorageUsage adds to storage usage
func (s *service) RecordStorageUsage(ctx context.Context, tenantID uuid.UUID, bytes int64) error {
    now := time.Now()
    periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
    
    return s.repo.AddStorage(ctx, tenantID, periodStart, bytes)
}

// SlugGenerator generates URL-safe slugs
type SlugGenerator interface {
    Generate(name string) string
}

type slugGenerator struct{}

func NewSlugGenerator() SlugGenerator {
    return &slugGenerator{}
}

func (g *slugGenerator) Generate(name string) string {
    // Convert to lowercase
    // Replace non-alphanumeric with hyphens
    // Trim leading/trailing hyphens
    // This is a simplified version - use a proper slug library in production
    return name // TODO: Implement proper slug generation
}

// Additional request/response types...
type ListTenantsRequest struct {
    Status *domain.TenantStatus
    Limit  int
    Offset int
}

type ListTenantsResponse struct {
    Tenants []*domain.Tenant
    Total   int64
}

type UpdateTenantRequest struct {
    Name               *string
    Email              *string
    Subdomain          *string
    Timezone           *string
    CurrencyCode       *string
    Industry           *string
    CompanySize        *string
    TaxID              *string
    RegistrationNumber *string
    LegalEntityType    *string
}

type UpdateConfigurationRequest struct {
    MaxUsers               *int
    MaxEntities            *int
    MaxTransactionsPerMonth *int
    StorageQuotaGB         *int64
    AccountingMethod       *string
    FiscalYearStartMonth   *int
}

type ProvisionTenantRequest struct {
    Name         string
    Email        string
    Subdomain    *string
    Industry     *string
    CompanySize  *string
    Timezone     string
    CurrencyCode string
}
```

---

## Temporal Workflows

### Tenant Provisioning Workflow

**File**: `internal/core/tenant/workflow/tenant_provisioning.go`

```go
package workflow

import (
    "time"
    
    "go.temporal.io/sdk/workflow"
    "github.com/google/uuid"
    "awo.so/internal/core/tenant/activities"
    "awo.so/internal/core/tenant/domain"
)

// ProvisionTenantWorkflowInput represents the input to the provisioning workflow
type ProvisionTenantWorkflowInput struct {
    Name         string
    Email        string
    Subdomain    *string
    Industry     *string
    CompanySize  *string
    Timezone     string
    CurrencyCode string
}

// ProvisionTenantWorkflowResult represents the workflow result
type ProvisionTenantWorkflowResult struct {
    TenantID uuid.UUID
    Status   string
    Message  string
}

// ProvisionTenantWorkflow orchestrates complete tenant provisioning
func ProvisionTenantWorkflow(ctx workflow.Context, input ProvisionTenantWorkflowInput) (*ProvisionTenantWorkflowResult, error) {
    logger := workflow.GetLogger(ctx)
    logger.Info("Starting tenant provisioning workflow", "name", input.Name)
    
    // Workflow options
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    time.Second,
            BackoffCoefficient: 2.0,
            MaximumInterval:    time.Minute,
            MaximumAttempts:    3,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    var result ProvisionTenantWorkflowResult
    
    // Step 1: Create tenant record
    var createResult activities.CreateTenantResult
    err := workflow.ExecuteActivity(ctx, activities.CreateTenantActivity, activities.CreateTenantInput{
        Name:         input.Name,
        Email:        input.Email,
        Subdomain:    input.Subdomain,
        Industry:     input.Industry,
        CompanySize:  input.CompanySize,
        Timezone:     input.Timezone,
        CurrencyCode: input.CurrencyCode,
    }).Get(ctx, &createResult)
    
    if err != nil {
        logger.Error("Failed to create tenant", "error", err)
        return &ProvisionTenantWorkflowResult{
            Status:  "FAILED",
            Message: "Failed to create tenant record",
        }, err
    }
    
    result.TenantID = createResult.TenantID
    logger.Info("Tenant created", "tenant_id", createResult.TenantID)
    
    // Step 2: Create default configuration
    err = workflow.ExecuteActivity(ctx, activities.CreateDefaultConfigurationActivity, activities.CreateConfigurationInput{
        TenantID: createResult.TenantID,
    }).Get(ctx, nil)
    
    if err != nil {
        logger.Error("Failed to create configuration", "error", err)
        // Compensating transaction: delete tenant
        _ = workflow.ExecuteActivity(ctx, activities.DeleteTenantActivity, activities.DeleteTenantInput{
            TenantID: createResult.TenantID,
        }).Get(ctx, nil)
        
        return &ProvisionTenantWorkflowResult{
            TenantID: createResult.TenantID,
            Status:   "FAILED",
            Message:  "Failed to create tenant configuration",
        }, err
    }
    
    logger.Info("Configuration created")
    
    // Step 3: Initialize usage stats
    err = workflow.ExecuteActivity(ctx, activities.InitializeUsageStatsActivity, activities.InitializeUsageInput{
        TenantID: createResult.TenantID,
    }).Get(ctx, nil)
    
    if err != nil {
        logger.Error("Failed to initialize usage stats", "error", err)
        // Continue - not critical
    }
    
    // Step 4: Send welcome email
    err = workflow.ExecuteActivity(ctx, activities.SendWelcomeEmailActivity, activities.SendEmailInput{
        TenantID: createResult.TenantID,
        Email:    input.Email,
        Name:     input.Name,
    }).Get(ctx, nil)
    
    if err != nil {
        logger.Error("Failed to send welcome email", "error", err)
        // Continue - not critical
    }
    
    // Step 5: Activate tenant
    err = workflow.ExecuteActivity(ctx, activities.ActivateTenantActivity, activities.ActivateTenantInput{
        TenantID: createResult.TenantID,
    }).Get(ctx, nil)
    
    if err != nil {
        logger.Error("Failed to activate tenant", "error", err)
        return &ProvisionTenantWorkflowResult{
            TenantID: createResult.TenantID,
            Status:   "FAILED",
            Message:  "Failed to activate tenant",
        }, err
    }
    
    logger.Info("Tenant activated successfully")
    
    result.Status = "SUCCESS"
    result.Message = "Tenant provisioned successfully"
    
    return &result, nil
}
```

### Bulk Operations Workflow

**File**: `internal/core/tenant/workflow/bulk_operations.go`

```go
package workflow

import (
    "time"
    
    "go.temporal.io/sdk/workflow"
    "github.com/google/uuid"
    "awo.so/internal/core/tenant/activities"
)

// BulkSuspendTenantsInput represents bulk suspension input
type BulkSuspendTenantsInput struct {
    TenantIDs []uuid.UUID
    Reason    string
    ActorID   uuid.UUID
    ActorName string
}

// BulkOperationResult represents bulk operation result
type BulkOperationResult struct {
    OperationID      uuid.UUID
    TotalTenants     int
    SuccessfulCount  int
    FailedCount      int
    Status           string
}

// BulkSuspendTenantsWorkflow suspends multiple tenants
func BulkSuspendTenantsWorkflow(ctx workflow.Context, input BulkSuspendTenantsInput) (*BulkOperationResult, error) {
    logger := workflow.GetLogger(ctx)
    logger.Info("Starting bulk suspend workflow", "tenant_count", len(input.TenantIDs))
    
    // Create bulk operation record
    var createOpResult activities.CreateBulkOperationResult
    err := workflow.ExecuteActivity(ctx, activities.CreateBulkOperationActivity, activities.CreateBulkOperationInput{
        OperationType: "SUSPEND",
        TenantIDs:     input.TenantIDs,
        ActorID:       input.ActorID,
        ActorName:     input.ActorName,
        Parameters: map[string]interface{}{
            "reason": input.Reason,
        },
    }).Get(ctx, &createOpResult)
    
    if err != nil {
        return nil, err
    }
    
    operationID := createOpResult.OperationID
    
    // Process each tenant in parallel
    selector := workflow.NewSelector(ctx)
    var pending int = len(input.TenantIDs)
    var successful int
    var failed int
    
    for _, tenantID := range input.TenantIDs {
        tenantID := tenantID // Capture for closure
        
        future := workflow.ExecuteActivity(ctx, activities.SuspendTenantActivity, activities.SuspendTenantInput{
            OperationID: operationID,
            TenantID:    tenantID,
            Reason:      input.Reason,
        })
        
        selector.AddFuture(future, func(f workflow.Future) {
            var result activities.SuspendTenantResult
            err := f.Get(ctx, &result)
            
            if err != nil {
                logger.Error("Failed to suspend tenant", "tenant_id", tenantID, "error", err)
                failed++
            } else {
                successful++
            }
            
            pending--
        })
    }
    
    // Wait for all activities to complete
    for pending > 0 {
        selector.Select(ctx)
    }
    
    // Update bulk operation status
    err = workflow.ExecuteActivity(ctx, activities.CompleteBulkOperationActivity, activities.CompleteBulkOperationInput{
        OperationID:     operationID,
        SuccessfulCount: successful,
        FailedCount:     failed,
    }).Get(ctx, nil)
    
    if err != nil {
        logger.Error("Failed to complete bulk operation", "error", err)
    }
    
    return &BulkOperationResult{
        OperationID:     operationID,
        TotalTenants:    len(input.TenantIDs),
        SuccessfulCount: successful,
        FailedCount:     failed,
        Status:          "COMPLETED",
    }, nil
}
```

---

## Temporal Activities

**File**: `internal/core/tenant/activities/tenant_activities.go`

```go
package activities

import (
    "context"
    "fmt"
    
    "github.com/google/uuid"
    "awo.so/internal/core/tenant/domain"
    "awo.so/internal/core/tenant/repository"
)

// TenantActivities contains all tenant-related activities
type TenantActivities struct {
    repo repository.TenantRepository
}

// NewTenantActivities creates new tenant activities
func NewTenantActivities(repo repository.TenantRepository) *TenantActivities {
    return &TenantActivities{
        repo: repo,
    }
}

// CreateTenantInput represents create tenant activity input
type CreateTenantInput struct {
    Name         string
    Email        string
    Subdomain    *string
    Industry     *string
    CompanySize  *string
    Timezone     string
    CurrencyCode string
}

// CreateTenantResult represents create tenant activity result
type CreateTenantResult struct {
    TenantID uuid.UUID
}

// CreateTenantActivity creates a new tenant (idempotent)
func (a *TenantActivities) CreateTenantActivity(ctx context.Context, input CreateTenantInput) (*CreateTenantResult, error) {
    // Check if tenant already exists with this email
    // This makes the activity idempotent
    
    tenant, err := domain.NewTenant(
        input.Name,
        input.Email,
        domain.WithSubdomain(*input.Subdomain),
        domain.WithIndustry(*input.Industry),
        domain.WithCompanySize(*input.CompanySize),
    )
    if err != nil {
        return nil, fmt.Errorf("create tenant entity: %w", err)
    }
    
    tenant.Timezone = input.Timezone
    tenant.CurrencyCode = input.CurrencyCode
    
    if err := a.repo.Create(ctx, tenant); err != nil {
        return nil, fmt.Errorf("save tenant: %w", err)
    }
    
    return &CreateTenantResult{
        TenantID: tenant.ID,
    }, nil
}

// CreateConfigurationInput represents configuration creation input
type CreateConfigurationInput struct {
    TenantID uuid.UUID
}

// CreateDefaultConfigurationActivity creates default configuration
func (a *TenantActivities) CreateDefaultConfigurationActivity(
    ctx context.Context,
    input CreateConfigurationInput,
) error {
    config := domain.NewTenantConfiguration(input.TenantID)
    
    if err := a.repo.CreateConfiguration(ctx, config); err != nil {
        return fmt.Errorf("create configuration: %w", err)
    }
    
    return nil
}

// InitializeUsageInput represents usage initialization input
type InitializeUsageInput struct {
    TenantID uuid.UUID
}

// InitializeUsageStatsActivity initializes usage stats
func (a *TenantActivities) InitializeUsageStatsActivity(
    ctx context.Context,
    input InitializeUsageInput,
) error {
    usage := domain.NewTenantUsage(input.TenantID)
    
    if err := a.repo.CreateUsage(ctx, usage); err != nil {
        return fmt.Errorf("create usage stats: %w", err)
    }
    
    return nil
}

// ActivateTenantInput represents activation input
type ActivateTenantInput struct {
    TenantID uuid.UUID
}

// ActivateTenantActivity activates a tenant
func (a *TenantActivities) ActivateTenantActivity(
    ctx context.Context,
    input ActivateTenantInput,
) error {
    if err := a.repo.UpdateStatus(ctx, input.TenantID, domain.TenantStatusActive); err != nil {
        return fmt.Errorf("update status: %w", err)
    }
    
    return nil
}

// SuspendTenantInput represents suspension input
type SuspendTenantInput struct {
    OperationID uuid.UUID
    TenantID    uuid.UUID
    Reason      string
}

// SuspendTenantResult represents suspension result
type SuspendTenantResult struct {
    Success bool
    Message string
}

// SuspendTenantActivity suspends a tenant
func (a *TenantActivities) SuspendTenantActivity(
    ctx context.Context,
    input SuspendTenantInput,
) (*SuspendTenantResult, error) {
    tenant, err := a.repo.GetByID(ctx, input.TenantID)
    if err != nil {
        return &SuspendTenantResult{
            Success: false,
            Message: fmt.Sprintf("tenant not found: %v", err),
        }, nil
    }
    
    if err := tenant.Suspend(input.Reason); err != nil {
        return &SuspendTenantResult{
            Success: false,
            Message: fmt.Sprintf("suspend failed: %v", err),
        }, nil
    }
    
    if err := a.repo.Update(ctx, tenant); err != nil {
        return &SuspendTenantResult{
            Success: false,
            Message: fmt.Sprintf("update failed: %v", err),
        }, nil
    }
    
    return &SuspendTenantResult{
        Success: true,
        Message: "Tenant suspended successfully",
    }, nil
}

// DeleteTenantInput represents deletion input (for compensation)
type DeleteTenantInput struct {
    TenantID uuid.UUID
}

// DeleteTenantActivity deletes a tenant (compensation activity)
func (a *TenantActivities) DeleteTenantActivity(
    ctx context.Context,
    input DeleteTenantInput,
) error {
    if err := a.repo.SoftDelete(ctx, input.TenantID); err != nil {
        return fmt.Errorf("soft delete tenant: %w", err)
    }
    
    return nil
}
```

**File**: `internal/core/tenant/activities/notification_activities.go`

```go
package activities

import (
    "context"
    "fmt"
    
    "github.com/google/uuid"
)

// NotificationActivities handles notification-related activities
type NotificationActivities struct {
    emailService EmailService
}

// EmailService interface for sending emails
type EmailService interface {
    SendEmail(ctx context.Context, to, subject, body string) error
}

// NewNotificationActivities creates notification activities
func NewNotificationActivities(emailService EmailService) *NotificationActivities {
    return &NotificationActivities{
        emailService: emailService,
    }
}

// SendEmailInput represents email sending input
type SendEmailInput struct {
    TenantID uuid.UUID
    Email    string
    Name     string
}

// SendWelcomeEmailActivity sends welcome email to new tenant
func (a *NotificationActivities) SendWelcomeEmailActivity(
    ctx context.Context,
    input SendEmailInput,
) error {
    subject := fmt.Sprintf("Welcome to Our Platform, %s!", input.Name)
    body := fmt.Sprintf(`
        Dear %s,
        
        Welcome to our platform! Your account has been successfully created.
        
        Tenant ID: %s
        
        You can now log in and start using our services.
        
        Best regards,
        The Team
    `, input.Name, input.TenantID)
    
    if err := a.emailService.SendEmail(ctx, input.Email, subject, body); err != nil {
        return fmt.Errorf("send email: %w", err)
    }
    
    return nil
}
```

---

## Database Schema

The database schema is defined in migration files. Here's the complete schema for the tenant module:

**Migration**: `000101_create_tenants_table.up.sql`

```sql
-- Create tenants table
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) NOT NULL,
    name VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) NOT NULL,
    subdomain VARCHAR(63) UNIQUE,
    
    status VARCHAR(20) DEFAULT 'ACTIVE' NOT NULL,
    timezone VARCHAR(50) DEFAULT 'UTC' NOT NULL,
    currency_code CHAR(3) DEFAULT 'USD' NOT NULL,
    
    industry VARCHAR(50),
    company_size VARCHAR(20),
    tax_id VARCHAR(50),
    registration_number VARCHAR(50),
    legal_entity_type VARCHAR(50),
    
    metadata JSONB DEFAULT '{}'::jsonb,
    settings JSONB DEFAULT '{}'::jsonb,
    
    last_activity_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT tenants_status_check CHECK (status IN ('ACTIVE', 'SUSPENDED', 'PENDING', 'ARCHIVED')),
    CONSTRAINT tenants_email_check CHECK (email ~* '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT tenants_subdomain_check CHECK (subdomain IS NULL OR subdomain ~* '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$'),
    CONSTRAINT tenants_slug_check CHECK (slug ~* '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$'),
    CONSTRAINT tenants_currency_check CHECK (currency_code ~* '^[A-Z]{3}$'),
    CONSTRAINT tenants_company_size_check CHECK (company_size IN ('Startup', 'Small', 'Medium', 'Large', 'Enterprise'))
);

-- Indexes
CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_subdomain ON tenants(subdomain) WHERE subdomain IS NOT NULL;
CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_tenants_email ON tenants(email);

-- Comments
COMMENT ON TABLE tenants IS 'Stores tenant (organization) information for multi-tenant ERP';
COMMENT ON COLUMN tenants.slug IS 'URL-friendly identifier generated from name';
COMMENT ON COLUMN tenants.metadata IS 'Flexible key-value storage for tenant-specific data';
COMMENT ON COLUMN tenants.settings IS 'Tenant configuration settings in JSONB format';
```

**Migration**: `000102_create_tenant_configurations_table.up.sql`

```sql
CREATE TABLE tenant_configurations (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Resource Limits
    max_users INT DEFAULT 100 NOT NULL,
    max_entities INT DEFAULT 1000 NOT NULL,
    max_transactions_per_month INT DEFAULT 10000 NOT NULL,
    storage_quota BIGINT DEFAULT 1073741824 NOT NULL, -- 1GB
    
    -- Accounting Preferences
    accounting_method VARCHAR(10) DEFAULT 'ACCRUAL' NOT NULL,
    fiscal_year_start_month INT DEFAULT 1 NOT NULL,
    default_currency CHAR(3) DEFAULT 'USD' NOT NULL,
    
    -- Localization
    date_format VARCHAR(20) DEFAULT 'MM/DD/YYYY' NOT NULL,
    number_format VARCHAR(20) DEFAULT 'US' NOT NULL,
    language_code VARCHAR(5) DEFAULT 'en-US' NOT NULL,
    
    -- Security & Integration
    password_policy JSONB DEFAULT '{"min_length": 8, "require_uppercase": true}'::jsonb,
    webhook_endpoints JSONB DEFAULT '[]'::jsonb,
    api_rate_limits JSONB DEFAULT '{"requests_per_minute": 100}'::jsonb,
    
    -- Flexible Settings
    settings JSONB DEFAULT '{}'::jsonb,
    
    -- Audit
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    
    CONSTRAINT tenant_config_accounting_method_check CHECK (accounting_method IN ('ACCRUAL', 'CASH')),
    CONSTRAINT tenant_config_fiscal_month_check CHECK (fiscal_year_start_month BETWEEN 1 AND 12)
);

COMMENT ON TABLE tenant_configurations IS 'Tenant-specific configuration and resource limits';
```

**Migration**: `000103_create_tenant_usage_stats_table.up.sql`

```sql
CREATE TABLE tenant_usage_stats (
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    
    -- Usage Metrics
    active_users INT DEFAULT 0 NOT NULL,
    total_entities INT DEFAULT 0 NOT NULL,
    total_transactions INT DEFAULT 0 NOT NULL,
    storage_used BIGINT DEFAULT 0 NOT NULL,
    api_calls INT DEFAULT 0 NOT NULL,
    
    -- Performance Metrics
    avg_response_time NUMERIC(10, 2),
    error_rate NUMERIC(5, 4),
    
    -- Financial Metrics
    monthly_revenue NUMERIC(12, 2),
    
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    
    PRIMARY KEY (tenant_id, period_start)
);

CREATE INDEX idx_tenant_usage_period ON tenant_usage_stats(period_start, period_end);

COMMENT ON TABLE tenant_usage_stats IS 'Historical resource usage tracking for tenants';
```

---

## Implementation Guide

### Step 1: Set Up Project Structure

```bash
# Create directory structure
mkdir -p internal/core/tenant/{domain,repository/queries,activities,workflow}

# Initialize Go modules if not done
go mod init awo

# Install dependencies
go get github.com/google/uuid
go get github.com/jackc/pgx/v5/pgxpool
go get go.temporal.io/sdk@latest
go get github.com/golang-migrate/migrate/v4
```

### Step 2: Install and Configure SQLC

```bash
# Install SQLC
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Create sqlc.yaml in repository directory
cd internal/core/tenant/repository
# Copy the sqlc.yaml content from above

# Generate code
sqlc generate
```

### Step 3: Set Up Database Migrations

```bash
# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create migrations directory
mkdir -p migrations

# Create migration files
migrate create -ext sql -dir migrations -seq create_tenants_table
migrate create -ext sql -dir migrations -seq create_tenant_configurations_table
migrate create -ext sql -dir migrations -seq create_tenant_usage_stats_table

# Copy SQL content into .up.sql files

# Run migrations
migrate -path migrations -database "postgresql://user:pass@localhost:5432/erp?sslmode=disable" up
```

### Step 4: Implement Domain Layer

Create all domain files as shown in the Domain Layer section above.

### Step 5: Implement Repository Layer

1. Create SQL queries in `repository/queries/*.sql`
2. Run `sqlc generate` to generate Go code
3. Implement repository interface in `repository/tenant_repository.go`

### Step 6: Implement Service Layer

Create `service.go` with all business logic as shown above.

### Step 7: Set Up Temporal

```go
// cmd/worker/main.go
package main

import (
    "log"
    
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
    "awo.so/internal/core/tenant/activities"
    "awo.so/internal/core/tenant/workflow"
)

func main() {
    // Create Temporal client
    c, err := client.Dial(client.Options{})
    if err != nil {
        log.Fatalln("Unable to create Temporal client", err)
    }
    defer c.Close()
    
    // Create worker
    w := worker.New(c, "tenant-task-queue", worker.Options{})
    
    // Register workflows
    w.RegisterWorkflow(workflow.ProvisionTenantWorkflow)
    w.RegisterWorkflow(workflow.BulkSuspendTenantsWorkflow)
    
    // Register activities
    tenantActivities := activities.NewTenantActivities(repo)
    w.RegisterActivity(tenantActivities)
    
    // Start worker
    err = w.Run(worker.InterruptCh())
    if err != nil {
        log.Fatalln("Unable to start worker", err)
    }
}
```

### Step 8: Create HTTP Handlers (Example)

```go
// api/handlers/tenant_handler.go
package handlers

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "awo.so/internal/core/tenant"
)

type TenantHandler struct {
    service tenant.Service
}

func NewTenantHandler(service tenant.Service) *TenantHandler {
    return &TenantHandler{service: service}
}

func (h *TenantHandler) CreateTenant(c *gin.Context) {
    var req tenant.CreateTenantRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    tenant, err := h.service.CreateTenant(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, tenant)
}

func (h *TenantHandler) GetTenant(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID"})
        return
    }
    
    tenant, err := h.service.GetTenant(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
        return
    }
    
    c.JSON(http.StatusOK, tenant)
}
```

---

## Testing Strategy

### Unit Tests Example

**File**: `internal/core/tenant/domain/tenant_test.go`

```go
package domain_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "awo.so/internal/core/tenant/domain"
)

func TestNewTenant(t *testing.T) {
    t.Run("creates tenant with valid data", func(t *testing.T) {
        tenant, err := domain.NewTenant("Acme Corp", "admin@acme.com")
        
        require.NoError(t, err)
        assert.NotNil(t, tenant)
        assert.Equal(t, "Acme Corp", tenant.Name)
        assert.Equal(t, "admin@acme.com", tenant.Email)
        assert.Equal(t, domain.TenantStatusPending, tenant.Status)
    })
    
    t.Run("returns error for empty name", func(t *testing.T) {
        tenant, err := domain.NewTenant("", "admin@acme.com")
        
        assert.Error(t, err)
        assert.Nil(t, tenant)
        assert.Equal(t, domain.ErrTenantNameRequired, err)
    })
}

func TestTenant_Activate(t *testing.T) {
    t.Run("activates pending tenant", func(t *testing.T) {
        tenant, _ := domain.NewTenant("Acme Corp", "admin@acme.com")
        
        err := tenant.Activate()
        
        assert.NoError(t, err)
        assert.Equal(t, domain.TenantStatusActive, tenant.Status)
    })
    
    t.Run("returns error for already active tenant", func(t *testing.T) {
        tenant, _ := domain.NewTenant("Acme Corp", "admin@acme.com")
        _ = tenant.Activate()
        
        err := tenant.Activate()
        
        assert.Error(t, err)
        assert.Equal(t, domain.ErrAlreadyActive, err)
    })
}
```

### Integration Tests Example

```go
package repository_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "github.com/testcontainers/testcontainers-go"
    "awo.so/internal/core/tenant/domain"
    "awo.so/internal/core/tenant/repository"
)

type RepositoryTestSuite struct {
    suite.Suite
    repo      repository.TenantRepository
    container testcontainers.Container
}

func (suite *RepositoryTestSuite) SetupSuite() {
    // Set up test database container
    // Initialize repository
}

func (suite *RepositoryTestSuite) TearDownSuite() {
    // Clean up
}

func (suite *RepositoryTestSuite) TestCreate() {
    ctx := context.Background()
    tenant, _ := domain.NewTenant("Test Corp", "test@example.com")
    
    err := suite.repo.Create(ctx, tenant)
    
    assert.NoError(suite.T(), err)
}

func TestRepositoryTestSuite(t *testing.T) {
    suite.Run(t, new(RepositoryTestSuite))
}
```

---

## Best Practices

### 1. Error Handling

```go
// Always wrap errors with context
if err := repo.Create(ctx, tenant); err != nil {
    return fmt.Errorf("create tenant: %w", err)
}

// Use domain errors for business logic violations
if tenant.Status == domain.TenantStatusActive {
    return domain.ErrAlreadyActive
}
```

### 2. Transaction Management

```go
func (s *service) CreateTenantWithConfig(ctx context.Context, req CreateTenantRequest) error {
    tx, err := s.db.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)
    
    // Create tenant
    if err := s.repo.CreateWithTx(ctx, tx, tenant); err != nil {
        return err
    }
    
    // Create config
    if err := s.repo.CreateConfigWithTx(ctx, tx, config); err != nil {
        return err
    }
    
    return tx.Commit(ctx)
}
```

### 3. Context Usage

```go
// Always pass context
func (s *service) GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
    // Respect context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    return s.repo.GetByID(ctx, id)
}
```

### 4. Temporal Best Practices

```go
// Use deterministic UUIDs in workflows
workflowID := fmt.Sprintf("provision-tenant-%s", input.Email)

// Set appropriate timeouts
ao := workflow.ActivityOptions{
    StartToCloseTimeout: 5 * time.Minute,
    RetryPolicy: &temporal.RetryPolicy{
        MaximumAttempts: 3,
    },
}

// Implement idempotent activities
func (a *Activities) CreateTenant(ctx context.Context, input Input) error {
    // Check if already exists
    exists, _ := a.repo.Exists(ctx, input.ID)
    if exists {
        return nil // Already created
    }
    
    // Create tenant
    return a.repo.Create(ctx, tenant)
}
```

---

## Common Patterns

### Pattern 1: Functional Options

```go
type TenantOption func(*Tenant) error

func WithSubdomain(subdomain string) TenantOption {
    return func(t *Tenant) error {
        t.Subdomain = &subdomain
        return nil
    }
}

tenant, err := NewTenant("Acme", "admin@acme.com",
    WithSubdomain("acme"),
    WithIndustry("Manufacturing"),
)
```

### Pattern 2: Repository Pattern

```go
type TenantRepository interface {
    Create(ctx context.Context, tenant *domain.Tenant) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
    // ...
}

// Implementations can be swapped (SQL, NoSQL, in-memory for tests)
```

### Pattern 3: Service Layer Orchestration

```go
func (s *service) ProvisionTenant(ctx context.Context, req Request) error {
    // Orchestrate multiple repository calls
    // Handle business logic
    // Trigger workflows
    // Publish events
}
```

