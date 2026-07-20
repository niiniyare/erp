> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Testing Overview
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Service Testing](../06-service-layer/03-service-testing.md)"
  - "[Handler Testing](../07-handler-layer/02-handler-testing.md)"
  - "[Integration Test Setup](02-integration-test-setup.md)"
---

# Testing Overview

## Test Pyramid

```
         ▲
        /E\         End-to-end (few — happy path only)
       /   \
      /integ\       Integration (test DB, real repo + service)
     /       \
    / unit    \     Unit (mocks, fast, many)
   /___________\
```

Ratio target: ~70% unit, ~25% integration, ~5% E2E.

## Test File Layout

```
internal/core/contracts/
  service/
    contract_service.go
    contract_service_test.go        ← unit tests for service
  repository/
    contract_sqlc.go
    contract_sqlc_integration_test.go  ← integration tests with real DB
  handler/
    contract_handler.go
    contract_handler_test.go        ← handler tests with app.Test()
```

Integration test files use `_integration_` in the name and the `//go:build integration` tag:

```go
//go:build integration
```

Run separately: `go test -tags integration ./...`

## Unit Tests: Service Layer

Use mock implementations of dependencies:

```go
// internal/core/contracts/service/contract_service_test.go

func TestContractService_Create_Success(t *testing.T) {
    repo := &mockContractRepo{}
    authz := &mockAuthzService{allow: true}
    events := &mockEventBus{}
    notif := &mockNotifService{}
    audit := &mockAuditService{}

    svc := NewContractService(repo, authz, events, notif, audit, slog.Default())

    sess := testSession()
    req := CreateParams{
        ContractNumber: "CONT-2025-0001",
        Title:          "Test Contract",
        TotalValue:     decimal.NewFromFloat(10000),
        Currency:       "USD",
        StartDate:      civil.DateOf(time.Now()),
    }

    contract, err := svc.Create(context.Background(), sess, req)
    require.NoError(t, err)
    assert.Equal(t, "CONT-2025-0001", contract.ContractNumber)
    assert.Equal(t, domain.StatusDraft, contract.Status)
}

func TestContractService_Create_PermissionDenied(t *testing.T) {
    repo := &mockContractRepo{}
    authz := &mockAuthzService{allow: false}  // deny

    svc := NewContractService(repo, authz, nil, nil, nil, slog.Default())

    _, err := svc.Create(context.Background(), testSession(), CreateParams{})
    require.ErrorIs(t, err, domain.ErrForbidden)
}
```

## Integration Tests: Repository Layer

Use a real test database — no mocks:

```go
//go:build integration

package repository_test

func TestContractRepo_Create_Integration(t *testing.T) {
    db := testDB(t)          // sets up real pool + RLS
    repo := NewContractRepo(db)

    tenantID := testTenant(t, db)
    sess := testSession(tenantID)

    contract, err := repo.Create(context.Background(), tenantID, CreateParams{
        ContractNumber: "CONT-TEST-001",
        Title:          "Integration Test Contract",
        TotalValue:     decimal.NewFromFloat(5000),
        Currency:       "USD",
    })
    require.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, contract.ID)
    assert.Equal(t, domain.StatusDraft, contract.Status)

    // Verify RLS: different tenant cannot see it
    otherTenantID := testTenant(t, db)
    _, err = repo.GetByID(context.Background(), contract.ID, otherTenantID)
    require.ErrorIs(t, err, domain.ErrContractNotFound)
}
```

## Handler Tests

Use `app.Test()` from Fiber to test the full HTTP stack without starting a real server:

```go
func TestContractHandler_Create(t *testing.T) {
    app := setupTestApp(t)

    body := `{"contract_number":"CONT-001","title":"Test","total_value":"1000","currency":"USD"}`
    req := httptest.NewRequest("POST", "/api/v1/contracts", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+testToken)

    resp, err := app.Test(req, -1)
    require.NoError(t, err)
    assert.Equal(t, 201, resp.StatusCode)
}
```

## Test Helpers

```go
// internal/testutil/db.go
func testDB(t *testing.T) *pgxpool.Pool {
    t.Helper()
    pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
    require.NoError(t, err)
    t.Cleanup(pool.Close)
    return pool
}

// internal/testutil/session.go
func testSession(tenantID uuid.UUID) iam.ResolvedSession {
    return iam.ResolvedSession{
        UserID:   uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001"),
        TenantID: tenantID,
        Roles:    []string{"contracts.editor"},
        EntityScope: iam.EntityScope{Type: iam.ScopeAll},
    }
}
```

## Mock Conventions

Keep mocks simple — implement the interface, track calls if needed:

```go
type mockContractRepo struct {
    created []*domain.Contract
    err     error
}

func (m *mockContractRepo) Create(_ context.Context, _ uuid.UUID, req CreateParams) (*domain.Contract, error) {
    if m.err != nil {
        return nil, m.err
    }
    c := &domain.Contract{
        ID:             uuid.New(),
        ContractNumber: req.ContractNumber,
        Title:          req.Title,
        Status:         domain.StatusDraft,
    }
    m.created = append(m.created, c)
    return c, nil
}
```

Don't use `testify/mock` for simple cases — plain structs are easier to read.

## State Machine Tests

Test every valid and invalid transition:

```go
func TestContractStatusTransitions(t *testing.T) {
    cases := []struct {
        from    domain.ContractStatus
        action  string
        allowed bool
    }{
        {domain.StatusDraft,       "submit",    true},
        {domain.StatusUnderReview, "approve",   true},
        {domain.StatusApproved,    "activate",  true},
        {domain.StatusActive,      "terminate", true},
        {domain.StatusDraft,       "approve",   false},  // must submit first
        {domain.StatusTerminated,  "activate",  false},  // terminal state
    }

    for _, tc := range cases {
        t.Run(fmt.Sprintf("%s→%s", tc.from, tc.action), func(t *testing.T) {
            // ...
        })
    }
}
```

## Test Data Cleanup

Integration tests must clean up after themselves:

```go
func TestContractRepo_Create_Integration(t *testing.T) {
    db := testDB(t)
    tenantID := uuid.New()

    // Cleanup runs even if test fails
    t.Cleanup(func() {
        db.Exec(context.Background(),
            "DELETE FROM contracts WHERE tenant_id = $1", tenantID)
        db.Exec(context.Background(),
            "DELETE FROM tenants WHERE id = $1", tenantID)
    })

    // ... test body
}
```

Alternatively, use transaction rollback per test — begin a transaction at test start, rollback at the end. Ensures full cleanup with no risk of leaving orphaned data.
