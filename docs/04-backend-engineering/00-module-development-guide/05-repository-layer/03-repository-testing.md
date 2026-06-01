---
title: Repository Testing
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Repository Layer Overview](01-repository-overview.md)"
  - "[RLS and Tenant Isolation](02-rls-and-tenant-isolation.md)"
  - "[Integration Test Setup](../18-testing/02-integration-test-setup.md)"
---

# Repository Testing

Repository tests hit a real database. No mocks — the goal is to verify SQL correctness and RLS isolation.

## Test Database Setup

```go
// internal/testutil/db.go
package testutil

import (
    "context"
    "os"
    "testing"

    "github.com/jackc/pgx/v5/pgxpool"
)

// TestDB returns a pool connected to the test database.
// Uses TEST_DATABASE_URL or falls back to a local default.
func TestDB(t *testing.T) *pgxpool.Pool {
    t.Helper()

    url := os.Getenv("TEST_DATABASE_URL")
    if url == "" {
        url = "postgres://postgres:postgres@localhost:5432/awoerp_test"
    }

    pool, err := pgxpool.New(context.Background(), url)
    if err != nil {
        t.Fatalf("connect test db: %v", err)
    }

    t.Cleanup(pool.Close)
    return pool
}
```

## Test Tenant and Data Helpers

```go
// internal/testutil/fixtures.go
package testutil

import (
    "context"
    "testing"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
)

// CreateTestTenant inserts an isolated tenant for a test.
// Deleted in t.Cleanup so tests don't pollute each other.
func CreateTestTenant(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
    t.Helper()

    tenantID := uuid.New()
    _, err := pool.Exec(context.Background(),
        `INSERT INTO tenants (id, name, slug, status) VALUES ($1, $2, $3, 'active')`,
        tenantID, "Test Tenant "+tenantID.String()[:8], "test-"+tenantID.String()[:8],
    )
    if err != nil {
        t.Fatalf("create test tenant: %v", err)
    }

    t.Cleanup(func() {
        pool.Exec(context.Background(),
            "DELETE FROM tenants WHERE id = $1", tenantID)
    })

    return tenantID
}
```

## Basic CRUD Test

```go
func TestContractRepo_Create(t *testing.T) {
    pool := testutil.TestDB(t)
    tenantID := testutil.CreateTestTenant(t, pool)

    store := db.NewStore(pool)
    repo := repository.NewContractRepository(store)

    contract, err := repo.Create(context.Background(), tenantID, repository.CreateContractParams{
        ContractNumber: "CONT-TEST-001",
        Title:          "Test Contract",
        TotalValue:     decimal.NewFromFloat(10000),
        Currency:       "USD",
        StartDate:      time.Now(),
        EndDate:        time.Now().AddDate(1, 0, 0),
        CreatedBy:      uuid.New(),
    })

    require.NoError(t, err)
    assert.Equal(t, "CONT-TEST-001", contract.ContractNumber)
    assert.Equal(t, tenantID, contract.TenantID)
    assert.Equal(t, domain.ContractStatusDraft, contract.Status)
    assert.Equal(t, 1, contract.Version)
}
```

## RLS Isolation Test

The most important test: verify tenant A cannot read tenant B's data.

```go
func TestContractRepo_RLS_CrossTenantIsolation(t *testing.T) {
    pool := testutil.TestDB(t)
    tenantA := testutil.CreateTestTenant(t, pool)
    tenantB := testutil.CreateTestTenant(t, pool)

    store := db.NewStore(pool)
    repo := repository.NewContractRepository(store)

    // Create contract in tenant A
    contract, err := repo.Create(context.Background(), tenantA, repository.CreateContractParams{
        ContractNumber: "CONT-A-001",
        Title:          "Tenant A Contract",
        TotalValue:     decimal.NewFromFloat(1000),
        Currency:       "USD",
        StartDate:      time.Now(),
        EndDate:        time.Now().AddDate(1, 0, 0),
        CreatedBy:      uuid.New(),
    })
    require.NoError(t, err)

    // Attempt to read from tenant B's context — must return not found
    _, err = repo.GetByID(context.Background(), contract.ID, tenantB)
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrContractNotFound,
        "tenant B should not be able to read tenant A's contract")

    // Confirm tenant A can still read it
    found, err := repo.GetByID(context.Background(), contract.ID, tenantA)
    require.NoError(t, err)
    assert.Equal(t, contract.ID, found.ID)
}
```

## Optimistic Lock Test

```go
func TestContractRepo_OptimisticLock(t *testing.T) {
    pool := testutil.TestDB(t)
    tenantID := testutil.CreateTestTenant(t, pool)

    store := db.NewStore(pool)
    repo := repository.NewContractRepository(store)

    contract, _ := repo.Create(context.Background(), tenantID, repository.CreateContractParams{
        ContractNumber: "CONT-LOCK-001",
        Title:          "Lock Test",
        TotalValue:     decimal.NewFromFloat(1000),
        Currency:       "USD",
        StartDate:      time.Now(),
        EndDate:        time.Now().AddDate(1, 0, 0),
        CreatedBy:      uuid.New(),
    })

    // First update with version=1 — should succeed
    updated, err := repo.Update(context.Background(), repository.UpdateContractParams{
        ID:       contract.ID,
        TenantID: tenantID,
        Title:    "Updated Title",
        Version:  1,
    })
    require.NoError(t, err)
    assert.Equal(t, 2, updated.Version)

    // Second update with stale version=1 — should fail
    _, err = repo.Update(context.Background(), repository.UpdateContractParams{
        ID:       contract.ID,
        TenantID: tenantID,
        Title:    "Stale Update",
        Version:  1, // stale
    })
    require.ErrorIs(t, err, domain.ErrVersionConflict)
}
```

## List with Filters Test

```go
func TestContractRepo_List_FilterByStatus(t *testing.T) {
    pool := testutil.TestDB(t)
    tenantID := testutil.CreateTestTenant(t, pool)

    store := db.NewStore(pool)
    repo := repository.NewContractRepository(store)

    createdBy := uuid.New()

    // Create 2 drafts + 1 active
    for i := 0; i < 2; i++ {
        repo.Create(context.Background(), tenantID, repository.CreateContractParams{
            ContractNumber: fmt.Sprintf("DRAFT-%d", i),
            Title:          fmt.Sprintf("Draft %d", i),
            TotalValue:     decimal.NewFromFloat(1000),
            Currency:       "USD",
            StartDate:      time.Now(),
            EndDate:        time.Now().AddDate(1, 0, 0),
            CreatedBy:      createdBy,
        })
    }

    status := domain.ContractStatusDraft
    results, total, err := repo.List(context.Background(), repository.ListContractsParams{
        TenantID: tenantID,
        Status:   &status,
        Limit:    10,
        Offset:   0,
    })

    require.NoError(t, err)
    assert.Equal(t, int64(2), total)
    assert.Len(t, results, 2)
    for _, r := range results {
        assert.Equal(t, domain.ContractStatusDraft, r.Status)
    }
}
```

## Unique Constraint Test

```go
func TestContractRepo_Create_DuplicateNumber(t *testing.T) {
    pool := testutil.TestDB(t)
    tenantID := testutil.CreateTestTenant(t, pool)

    store := db.NewStore(pool)
    repo := repository.NewContractRepository(store)

    params := repository.CreateContractParams{
        ContractNumber: "UNIQUE-001",
        Title:          "Contract",
        TotalValue:     decimal.NewFromFloat(1000),
        Currency:       "USD",
        StartDate:      time.Now(),
        EndDate:        time.Now().AddDate(1, 0, 0),
        CreatedBy:      uuid.New(),
    }

    _, err := repo.Create(context.Background(), tenantID, params)
    require.NoError(t, err)

    // Same number in same tenant — must fail
    _, err = repo.Create(context.Background(), tenantID, params)
    require.ErrorIs(t, err, domain.ErrDuplicateContractNumber)

    // Same number in different tenant — must succeed
    otherTenant := testutil.CreateTestTenant(t, pool)
    _, err = repo.Create(context.Background(), otherTenant, params)
    require.NoError(t, err, "same contract number is allowed across tenants")
}
```

## Soft Delete Test

```go
func TestContractRepo_SoftDelete(t *testing.T) {
    pool := testutil.TestDB(t)
    tenantID := testutil.CreateTestTenant(t, pool)

    store := db.NewStore(pool)
    repo := repository.NewContractRepository(store)

    contract, _ := repo.Create(context.Background(), tenantID, repository.CreateContractParams{
        ContractNumber: "DELETE-001",
        Title:          "To Delete",
        TotalValue:     decimal.NewFromFloat(1),
        Currency:       "USD",
        StartDate:      time.Now(),
        EndDate:        time.Now().AddDate(1, 0, 0),
        CreatedBy:      uuid.New(),
    })

    err := repo.Delete(context.Background(), contract.ID, tenantID)
    require.NoError(t, err)

    // Normal get must return not found
    _, err = repo.GetByID(context.Background(), contract.ID, tenantID)
    require.ErrorIs(t, err, domain.ErrContractNotFound)

    // Row still exists in DB (soft delete — deleted_at is set)
    var count int
    pool.QueryRow(context.Background(),
        "SELECT COUNT(*) FROM contracts WHERE id = $1", contract.ID,
    ).Scan(&count)
    assert.Equal(t, 1, count, "soft delete should not remove the row")
}
```

## Test File Convention

Repository test files live alongside the implementation:

```
internal/core/contracts/
  repository/
    contract_repo.go
    contract_repo_test.go    ← integration tests (require DB)
```

Build tag to skip in unit-only runs:

```go
//go:build integration

package repository_test
```

Run with: `go test -tags=integration ./...`
