---
title: Integration Test Setup
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Testing Overview](01-testing-overview.md)"
  - "[Database Setup](../21-server-startup/05-database-setup.md)"
  - "[Worked Example: Tests](../23-worked-example/09-tests.md)"
---

# Integration Test Setup

## Build Tag

All integration tests use the `integration` build tag to keep them out of unit test runs:

```go
//go:build integration

package repository_test
```

Run with: `go test -tags integration ./...`

## Test Database Setup

```go
// internal/testutil/db.go

//go:build integration

package testutil

import (
    "context"
    "os"
    "testing"

    "github.com/jackc/pgx/v5/pgxpool"
)

// TestDB creates a pool for integration tests.
// Reads DATABASE_URL from environment.
func TestDB(t *testing.T) *pgxpool.Pool {
    t.Helper()
    url := os.Getenv("DATABASE_URL")
    if url == "" {
        t.Skip("DATABASE_URL not set — skipping integration test")
    }
    pool, err := pgxpool.New(context.Background(), url)
    if err != nil {
        t.Fatalf("connect to test DB: %v", err)
    }
    t.Cleanup(pool.Close)
    return pool
}

// TestTenant creates a tenant for a test and cleans it up after.
func TestTenant(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
    t.Helper()
    tenantID := uuid.New()
    _, err := pool.Exec(context.Background(), `
        INSERT INTO tenants (id, name, subdomain, status, plan, company_size, country)
        VALUES ($1, $2, $3, 'ACTIVE', 'starter', 'SMALL', 'US')
    `, tenantID, "Test Tenant "+tenantID.String()[:8], "test-"+tenantID.String()[:8])
    if err != nil {
        t.Fatalf("create test tenant: %v", err)
    }
    t.Cleanup(func() {
        pool.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
    })
    return tenantID
}
```

## Isolation: Transaction Rollback

For fast, isolated tests, wrap each test in a transaction and roll back:

```go
func TestContractRepo_Create_WithRollback(t *testing.T) {
    pool := testutil.TestDB(t)

    // Begin transaction for this test
    tx, err := pool.Begin(context.Background())
    require.NoError(t, err)
    t.Cleanup(func() { tx.Rollback(context.Background()) })

    tenantID := uuid.New()
    _, err = tx.Exec(context.Background(), `
        INSERT INTO tenants (id, name, subdomain, status, plan, company_size, country)
        VALUES ($1, 'Test', 'test', 'ACTIVE', 'starter', 'SMALL', 'US')`, tenantID)
    require.NoError(t, err)

    // Set tenant GUC
    _, err = tx.Exec(context.Background(), "SET LOCAL app.tenant_id = $1", tenantID.String())
    require.NoError(t, err)

    // Run test using tx directly
    q := sqlc.New(tx)
    contract, err := q.CreateContract(context.Background(), sqlc.CreateContractParams{
        ContractNumber: "TEST-001",
        Title:          "Test",
        ContractType:   "service",
        TotalValue:     decimal.NewFromFloat(1000),
        Currency:       "USD",
        // ...
    })
    require.NoError(t, err)
    assert.Equal(t, "TEST-001", contract.ContractNumber)

    // Rollback at end — no cleanup needed
}
```

## Isolation: Separate Tenant Per Test

Alternatively, create a unique tenant per test and clean up:

```go
func TestContractRepo_RLS_Isolation(t *testing.T) {
    pool := testutil.TestDB(t)
    tenant1 := testutil.TestTenant(t, pool)
    tenant2 := testutil.TestTenant(t, pool)
    store := db.NewStore(pool)

    // Create contract in tenant1
    var contract *domain.Contract
    err := store.WithTenant(context.Background(), tenant1, func(tx *pgx.Tx) error {
        q := sqlc.New(tx)
        row, err := q.CreateContract(context.Background(), sqlc.CreateContractParams{
            ContractNumber: "TEST-001",
            Title: "Tenant 1 Contract",
            // ...
        })
        if err != nil {
            return err
        }
        contract = toDomain(row)
        return nil
    })
    require.NoError(t, err)

    // Attempt to access from tenant2 — must fail
    err = store.WithTenant(context.Background(), tenant2, func(tx *pgx.Tx) error {
        q := sqlc.New(tx)
        _, err := q.GetContractByID(context.Background(), contract.ID)
        return err
    })
    require.ErrorIs(t, err, pgx.ErrNoRows, "RLS must block cross-tenant access")
}
```

## Handler Integration Tests

```go
func setupTestApp(t *testing.T) *fiber.App {
    t.Helper()
    pool := testutil.TestDB(t)
    store := db.NewStore(pool)

    // Real dependencies
    repo := repository.NewContractRepo(store)
    authzSvc := authz.NewCasbinService(...)  // or mock
    svc := service.NewContractService(repo, authzSvc, ...)

    app := fiber.New()
    handler := handler.NewContractHandler(svc)

    // Inject test session via middleware
    app.Use(func(c *fiber.Ctx) error {
        c.Locals("resolved_session", testSession(testTenantID))
        return c.Next()
    })

    registrar := handler.NewContractRouteRegistrar(handler, authzSvc)
    registrar.Register(app.Group("/api/v1"))
    return app
}

func TestContractHandler_Create_Integration(t *testing.T) {
    app := setupTestApp(t)

    body := `{
        "contract_number": "IT-001",
        "title": "Integration Test",
        "contract_type": "service",
        "total_value": "5000",
        "currency": "USD",
        "start_date": "2025-01-01",
        "end_date": "2025-12-31"
    }`
    req := httptest.NewRequest("POST", "/api/v1/contracts", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    resp, err := app.Test(req, -1)
    require.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
    assert.NotEmpty(t, resp.Header.Get("Location"))
}
```

## Parallel Tests

Mark independent integration tests parallel to speed up the suite:

```go
func TestContractRepo_Create(t *testing.T) {
    t.Parallel()   // run concurrently with other parallel tests
    pool := testutil.TestDB(t)
    // each parallel test uses its own tenant — no contention
    tenant := testutil.TestTenant(t, pool)
    // ...
}
```

Parallel tests must use separate tenants to avoid RLS conflicts.
