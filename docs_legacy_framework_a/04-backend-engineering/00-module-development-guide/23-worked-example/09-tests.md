> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Worked Example — Tests
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Testing Overview](../18-testing/01-testing-overview.md)"
  - "[Integration Test Setup](../18-testing/02-integration-test-setup.md)"
  - "[Service Testing](../06-service-layer/03-service-testing.md)"
  - "[Worked Example Overview](01-worked-example-overview.md)"
---

# Worked Example — Tests

## Domain Test: State Machine Matrix

```go
// internal/core/contracts/domain/contract_test.go
package domain_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "awo.so/internal/core/contracts/domain"
)

func TestCanTransitionTo(t *testing.T) {
    tests := []struct {
        from    domain.ContractStatus
        to      domain.ContractStatus
        allowed bool
    }{
        {domain.ContractStatusDraft,       domain.ContractStatusSubmitted,   true},
        {domain.ContractStatusDraft,       domain.ContractStatusActive,      false},
        {domain.ContractStatusSubmitted,   domain.ContractStatusUnderReview, true},
        {domain.ContractStatusSubmitted,   domain.ContractStatusDraft,       true},
        {domain.ContractStatusSubmitted,   domain.ContractStatusApproved,    false},
        {domain.ContractStatusUnderReview, domain.ContractStatusApproved,    true},
        {domain.ContractStatusUnderReview, domain.ContractStatusDraft,       true},
        {domain.ContractStatusApproved,    domain.ContractStatusActive,      true},
        {domain.ContractStatusApproved,    domain.ContractStatusDraft,       true},
        {domain.ContractStatusActive,      domain.ContractStatusSuspended,   true},
        {domain.ContractStatusActive,      domain.ContractStatusTerminated,  true},
        {domain.ContractStatusActive,      domain.ContractStatusDraft,       false},
        {domain.ContractStatusSuspended,   domain.ContractStatusActive,      true},
        {domain.ContractStatusTerminated,  domain.ContractStatusDraft,       false},
        {domain.ContractStatusTerminated,  domain.ContractStatusActive,      false},
    }

    for _, tt := range tests {
        name := fmt.Sprintf("%s→%s", tt.from, tt.to)
        t.Run(name, func(t *testing.T) {
            c := &domain.Contract{Status: tt.from}
            assert.Equal(t, tt.allowed, c.CanTransitionTo(tt.to))
        })
    }
}
```

## Repository Test: Create and RLS

```go
// internal/core/contracts/repository/contract_integration_test.go
package repository_test

func (s *ContractRepositoryTestSuite) TestCreate_Success() {
    contract, err := s.repo.Create(context.Background(), repository.CreateContractParams{
        TenantID: s.tenantID, EntityID: s.entityID,
        ContractNumber: "CONT-2025-0001", Title: "Test Contract",
        VendorID: uuid.New(), ContractType: domain.ContractTypeService,
        StartDate: time.Now(), EndDate: time.Now().AddDate(0, 6, 0),
        TotalValue: decimal.Zero, Currency: "USD", CreatedBy: uuid.New(),
    })
    require.NoError(s.T(), err)
    assert.NotEqual(s.T(), uuid.Nil, contract.ID)
    assert.Equal(s.T(), domain.ContractStatusDraft, contract.Status)
    assert.Equal(s.T(), 1, contract.Version)
}

func (s *ContractRepositoryTestSuite) TestGetByID_CrossTenantBlocked() {
    contract := s.createTestContract()
    tenantB  := testutil.CreateTestTenant(s.T(), s.store)

    _, err := s.repo.GetByID(context.Background(), contract.ID, tenantB)
    assert.ErrorIs(s.T(), err, domain.ErrContractNotFound,
        "cross-tenant access must be invisible")
}

func (s *ContractRepositoryTestSuite) TestUpdate_VersionConflict() {
    contract := s.createTestContract()

    // Correct version succeeds
    updated, err := s.repo.Update(context.Background(), repository.UpdateContractParams{
        ID: contract.ID, TenantID: s.tenantID,
        Title: "Updated", Version: 1, UpdatedBy: uuid.New(),
    })
    require.NoError(s.T(), err)
    assert.Equal(s.T(), 2, updated.Version)

    // Stale version fails
    _, err = s.repo.Update(context.Background(), repository.UpdateContractParams{
        ID: contract.ID, TenantID: s.tenantID,
        Title: "Stale", Version: 1, UpdatedBy: uuid.New(), // stale
    })
    assert.ErrorIs(s.T(), err, domain.ErrContractConflict)
}
```

## Service Test: Authorization Blocks Create

```go
// internal/core/contracts/service/contract_service_test.go
func TestCreate_AuthDenied_RepoNotCalled(t *testing.T) {
    repo  := &mockContractRepo{}
    authz := &mockAuthzService{}

    authz.On("Enforce", mock.Anything, mock.Anything).Return(false, nil)

    svc := buildService(repo, authz)
    _, err := svc.Create(context.Background(), service.CreateContractRequest{
        TenantID: testutil.TestTenantID, Principal: testutil.TestPrincipal(),
        // ...
    })

    assert.ErrorIs(t, err, iam.ErrForbidden)
    repo.AssertNotCalled(t, "Create")
}
```

## Run Tests

```bash
# Domain tests (no DB needed)
go test ./internal/core/contracts/domain/...

# Repository tests (requires TEST_DATABASE_URL)
TEST_DATABASE_URL=postgres://... go test ./internal/core/contracts/repository/...

# Service unit tests (no DB needed)
go test ./internal/core/contracts/service/...

# Full coverage report
go test -coverprofile=coverage.out ./internal/core/contracts/...
go tool cover -func=coverage.out | grep total
```
