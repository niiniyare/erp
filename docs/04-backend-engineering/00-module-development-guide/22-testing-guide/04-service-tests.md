---
title: Service Unit Tests
portal: 4 — Backend Engineering
section: 00-module-development-guide/22-testing-guide
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-testing-overview.md
    title: Testing Overview
---

# Service Unit Tests

Service tests use mocked repositories and platform services. They test business rules, authorization enforcement, state machine checks, and error propagation — without a database.

## Mock Setup with testify/mock

```go
// internal/core/contracts/service/contract_test.go
package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"awo.so/internal/core/contracts/domain"
	"awo.so/internal/core/contracts/repository"
	"awo.so/internal/core/contracts/service"
	iam "awo.so/internal/core/iam"
)

type mockContractRepo struct{ mock.Mock }

func (m *mockContractRepo) Create(ctx context.Context, p repository.CreateContractParams) (*domain.Contract, error) {
	args := m.Called(ctx, p)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*domain.Contract), args.Error(1)
}

func (m *mockContractRepo) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Contract, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*domain.Contract), args.Error(1)
}

// ... implement remaining interface methods ...

type mockAuthzService struct{ mock.Mock }

func (m *mockAuthzService) Enforce(ctx context.Context, req iam.Request) (bool, error) {
	args := m.Called(ctx, req)
	return args.Bool(0), args.Error(1)
}

// ... implement remaining interface methods ...

func buildService(repo *mockContractRepo, authz *mockAuthzService) service.ContractService {
	return service.NewContractService(
		repo,
		&mockLineRepo{},
		authz,
		&mockAuditSvc{},
		&mockNotifSvc{},
		&mockEventBus{},
		testLogger(),
		noopTracer(),
		noopMetrics(),
	)
}
```

## Test: Authorization Enforced on Create

```go
func TestContractService_Create_AuthorizationEnforced(t *testing.T) {
	repo := &mockContractRepo{}
	authz := &mockAuthzService{}

	// Authz denies the request
	authz.On("Enforce", mock.Anything, mock.MatchedBy(func(req iam.Request) bool {
		return req.Action == "create"
	})).Return(false, nil)

	svc := buildService(repo, authz)

	_, err := svc.Create(context.Background(), service.CreateContractRequest{
		TenantID:  testTenantID,
		Principal: testPrincipal,
		// ...
	})

	assert.ErrorIs(t, err, iam.ErrForbidden)
	repo.AssertNotCalled(t, "Create")  // repo must not be called after auth denial
}
```

## Test: State Machine Checked Before Repo

```go
func TestContractService_Submit_InvalidTransition(t *testing.T) {
	repo := &mockContractRepo{}
	authz := &mockAuthzService{}

	// Auth allows
	authz.On("Enforce", mock.Anything, mock.Anything).Return(true, nil)

	// Contract is already submitted (cannot submit again)
	alreadySubmitted := &domain.Contract{
		ID:       testContractID,
		TenantID: testTenantID,
		Status:   domain.ContractStatusSubmitted,
		Version:  1,
	}
	repo.On("GetByID", mock.Anything, testContractID, testTenantID).
		Return(alreadySubmitted, nil)

	svc := buildService(repo, authz)

	_, err := svc.Submit(context.Background(), testContractID, testTenantID, 1, testUserID, testPrincipal)

	assert.ErrorIs(t, err, domain.ErrContractInvalidTransition)
	repo.AssertNotCalled(t, "UpdateStatus")  // UpdateStatus must not be called after invalid transition
}
```

## Test: Optimistic Conflict Propagates

```go
func TestContractService_Update_ConflictPropagates(t *testing.T) {
	repo := &mockContractRepo{}
	authz := &mockAuthzService{}

	authz.On("Enforce", mock.Anything, mock.Anything).Return(true, nil)
	repo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).
		Return(&domain.Contract{Status: domain.ContractStatusDraft, Version: 2}, nil)
	repo.On("Update", mock.Anything, mock.Anything).
		Return(nil, domain.ErrContractConflict)

	svc := buildService(repo, authz)

	_, err := svc.Update(context.Background(), service.UpdateContractRequest{
		ID:       testContractID,
		TenantID: testTenantID,
		Version:  1,  // stale
		Principal: testPrincipal,
	})

	assert.ErrorIs(t, err, domain.ErrContractConflict)
}
```

## Test: Business Rule — Only Draft is Editable

```go
func TestContractService_Update_NotEditableWhenSubmitted(t *testing.T) {
	repo := &mockContractRepo{}
	authz := &mockAuthzService{}

	authz.On("Enforce", mock.Anything, mock.Anything).Return(true, nil)
	repo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).
		Return(&domain.Contract{Status: domain.ContractStatusSubmitted}, nil)

	svc := buildService(repo, authz)

	_, err := svc.Update(context.Background(), service.UpdateContractRequest{
		ID:        testContractID,
		TenantID:  testTenantID,
		Principal: testPrincipal,
		Version:   1,
	})

	assert.ErrorIs(t, err, domain.ErrContractNotEditable)
	repo.AssertNotCalled(t, "Update")
}
```
