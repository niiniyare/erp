> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Repository Integration Tests
portal: 4 — Backend Engineering
section: 00-module-development-guide/22-testing-guide
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-testing-overview.md
    title: Testing Overview
---

# Repository Integration Tests

Repository tests run against a real PostgreSQL test database. They test the complete persistence layer: SQL correctness, RLS enforcement, optimistic locking, and error mapping.

## Test Setup

```go
// internal/core/contracts/repository/contract_integration_test.go
package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"awo.so/internal/core/contracts/domain"
	"awo.so/internal/core/contracts/repository"
	"awo.so/internal/testutil"
)

type ContractRepositoryTestSuite struct {
	suite.Suite
	store    db.Store
	repo     repository.ContractRepository
	tenantID uuid.UUID
	entityID uuid.UUID
}

func (s *ContractRepositoryTestSuite) SetupSuite() {
	s.store = testutil.NewTestStore(s.T())
}

func (s *ContractRepositoryTestSuite) SetupTest() {
	// Create fresh tenant and entity for each test — isolation
	s.tenantID = testutil.CreateTestTenant(s.T(), s.store)
	s.entityID = testutil.CreateTestEntity(s.T(), s.store, s.tenantID)
	s.repo = repository.NewContractRepository(s.store)
}

func (s *ContractRepositoryTestSuite) TearDownTest() {
	testutil.CleanupTenant(s.T(), s.store, s.tenantID)
}

func TestContractRepository(t *testing.T) {
	suite.Run(t, new(ContractRepositoryTestSuite))
}
```

## Test: Create and Retrieve

```go
func (s *ContractRepositoryTestSuite) TestCreate_Success() {
	params := repository.CreateContractParams{
		TenantID:       s.tenantID,
		EntityID:       s.entityID,
		ContractNumber: "CONT-2025-0001",
		Title:          "Test Contract",
		VendorID:       uuid.New(),
		ContractType:   domain.ContractTypeService,
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 6, 0),
		TotalValue:     decimal.Zero,
		Currency:       "USD",
		CreatedBy:      uuid.New(),
	}

	contract, err := s.repo.Create(context.Background(), params)
	require.NoError(s.T(), err)

	assert.NotEqual(s.T(), uuid.Nil, contract.ID)
	assert.Equal(s.T(), domain.ContractStatusDraft, contract.Status)
	assert.Equal(s.T(), 1, contract.Version)
	assert.Equal(s.T(), "CONT-2025-0001", contract.ContractNumber)

	// Verify we can fetch it back
	fetched, err := s.repo.GetByID(context.Background(), contract.ID, s.tenantID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), contract.ID, fetched.ID)
}
```

## Test: Not Found

```go
func (s *ContractRepositoryTestSuite) TestGetByID_NotFound() {
	_, err := s.repo.GetByID(context.Background(), uuid.New(), s.tenantID)
	assert.ErrorIs(s.T(), err, domain.ErrContractNotFound)
}
```

## Test: Optimistic Lock Conflict

```go
func (s *ContractRepositoryTestSuite) TestUpdate_VersionConflict() {
	contract := s.createTestContract()

	// Update with correct version → succeeds
	updated, err := s.repo.Update(context.Background(), repository.UpdateContractParams{
		ID:       contract.ID,
		TenantID: s.tenantID,
		Title:    "Updated Title",
		Version:  1,
		UpdatedBy: uuid.New(),
	})
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, updated.Version)

	// Update with stale version (1) → conflict
	_, err = s.repo.Update(context.Background(), repository.UpdateContractParams{
		ID:       contract.ID,
		TenantID: s.tenantID,
		Title:    "Another Update",
		Version:  1,   // stale
		UpdatedBy: uuid.New(),
	})
	assert.ErrorIs(s.T(), err, domain.ErrContractConflict)
}
```

## Test: RLS Tenant Isolation

```go
func (s *ContractRepositoryTestSuite) TestGetByID_CrossTenantBlocked() {
	// Create contract in Tenant A
	contract := s.createTestContract()

	// Create Tenant B
	tenantB := testutil.CreateTestTenant(s.T(), s.store)

	// Attempt to fetch Tenant A's contract using Tenant B's credentials
	_, err := s.repo.GetByID(context.Background(), contract.ID, tenantB)
	assert.ErrorIs(s.T(), err, domain.ErrContractNotFound,
		"contract from another tenant must be invisible")
}
```

## Test: Duplicate Contract Number

```go
func (s *ContractRepositoryTestSuite) TestCreate_DuplicateNumber() {
	s.createTestContractWithNumber("CONT-2025-0001")

	_, err := s.createTestContractWithNumber("CONT-2025-0001")
	assert.ErrorIs(s.T(), err, domain.ErrContractAlreadyExists)
}
```

## Test Helper

```go
func (s *ContractRepositoryTestSuite) createTestContract() *domain.Contract {
	return s.createTestContractWithNumber("CONT-TEST-" + uuid.New().String()[:8])
}

func (s *ContractRepositoryTestSuite) createTestContractWithNumber(number string) *domain.Contract {
	contract, err := s.repo.Create(context.Background(), repository.CreateContractParams{
		TenantID:       s.tenantID,
		EntityID:       s.entityID,
		ContractNumber: number,
		Title:          "Test Contract",
		VendorID:       uuid.New(),
		ContractType:   domain.ContractTypeService,
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(1, 0, 0),
		TotalValue:     decimal.Zero,
		Currency:       "USD",
		CreatedBy:      uuid.New(),
	})
	require.NoError(s.T(), err)
	return contract
}
```
