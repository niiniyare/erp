> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Test Helpers
portal: 4 — Backend Engineering
section: 00-module-development-guide/22-testing-guide
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-testing-overview.md
    title: Testing Overview
  - path: ./03-repository-tests.md
    title: Repository Tests
---

# Test Helpers

Shared test utilities in `internal/testutil/`. Import them in any `_test.go` file.

## testutil/db.go

```go
package testutil

import (
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	db "awo.so/db/sqlc"
)

// NewTestStore returns a db.Store connected to the test database.
// Fails the test if TEST_DATABASE_URL is not set.
func NewTestStore(t *testing.T) db.Store {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping integration test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect to test DB: %v", err)
	}
	t.Cleanup(pool.Close)
	return db.NewStore(pool)
}
```

## testutil/fixtures.go

```go
package testutil

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"awo.so/internal/core/contracts/domain"
	"awo.so/internal/core/contracts/repository"
)

// BuildDraftContract returns a pre-built domain.Contract in draft status for use in tests.
func BuildDraftContract(tenantID, entityID uuid.UUID) *domain.Contract {
	v, _ := domain.NewContractValue(decimal.NewFromFloat(50000))
	return &domain.Contract{
		ID:             uuid.New(),
		TenantID:       tenantID,
		EntityID:       entityID,
		ContractNumber: "CONT-TEST-" + uuid.New().String()[:8],
		Title:          "Test Contract",
		Status:         domain.ContractStatusDraft,
		Version:        1,
		TotalValue:     v,
		Currency:       "USD",
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(1, 0, 0),
		VendorID:       uuid.New(),
		ContractType:   domain.ContractTypeService,
		CreatedBy:      uuid.New(),
		UpdatedBy:      uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// CreateTestTenant creates a test tenant in the database and returns its ID.
func CreateTestTenant(t *testing.T, store db.Store) uuid.UUID {
	t.Helper()
	// ... insert tenant row ...
	return tenantID
}

// CreateTestEntity creates a test entity in the database and returns its ID.
func CreateTestEntity(t *testing.T, store db.Store, tenantID uuid.UUID) uuid.UUID {
	t.Helper()
	// ... insert entity row ...
	return entityID
}

// CleanupTenant removes all test data for a tenant.
func CleanupTenant(t *testing.T, store db.Store, tenantID uuid.UUID) {
	t.Helper()
	// Cascading delete via tenant_id — respects FK ordering
}
```

## testutil/session.go

```go
package testutil

import (
	"github.com/google/uuid"

	iam "awo.so/internal/core/iam"
)

var (
	TestTenantID = uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	TestUserID   = uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	TestEntityID = uuid.MustParse("cccccccc-0000-0000-0000-000000000001")
)

// TestSession returns a ResolvedSession for use in tests.
func TestSession() *iam.ResolvedSession {
	return &iam.ResolvedSession{
		TenantID: TestTenantID,
		UserID:   TestUserID,
		EntityScope: iam.EntityScope{
			Type:     iam.EntityScopeAll,
			EntityID: TestEntityID.String(),
		},
	}
}

// TestPrincipal returns a Principal for use in service tests.
func TestPrincipal() iam.Principal {
	return iam.Principal{
		Subject: iam.TenantSubject(TestUserID),
		Domain:  iam.TenantDomain(TestTenantID),
	}
}
```

## testutil/assert.go

```go
package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"awo.so/internal/core/contracts/domain"
)

// AssertContractStatus asserts that a contract has the expected status.
func AssertContractStatus(t *testing.T, c *domain.Contract, expected domain.ContractStatus) {
	t.Helper()
	assert.Equal(t, expected, c.Status,
		"expected contract status %q, got %q", expected, c.Status)
}

// AssertVersionIncremented asserts that the version increased by exactly 1.
func AssertVersionIncremented(t *testing.T, before, after int) {
	t.Helper()
	assert.Equal(t, before+1, after,
		"expected version to increment from %d to %d, got %d", before, before+1, after)
}
```
