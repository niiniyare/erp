---
title: Service Testing
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Service Layer Overview](01-service-overview.md)"
  - "[Testing Overview](../18-testing/01-testing-overview.md)"
  - "[Worked Example: Tests](../23-worked-example/09-tests.md)"
---

# Service Testing

## Unit Test Setup

Service unit tests use mock implementations. Keep mocks minimal:

```go
// service/mocks_test.go
package service_test

type mockContractRepo struct {
    contracts map[uuid.UUID]*domain.Contract
    createErr error
    getErr    error
}

func newMockRepo() *mockContractRepo {
    return &mockContractRepo{
        contracts: make(map[uuid.UUID]*domain.Contract),
    }
}

func (m *mockContractRepo) Create(_ context.Context, tenantID uuid.UUID, p repository.CreateParams) (*domain.Contract, error) {
    if m.createErr != nil {
        return nil, m.createErr
    }
    c := &domain.Contract{
        ID:             uuid.New(),
        TenantID:       tenantID,
        ContractNumber: p.ContractNumber,
        Title:          p.Title,
        Status:         domain.StatusDraft,
        TotalValue:     p.TotalValue,
        Version:        1,
    }
    m.contracts[c.ID] = c
    return c, nil
}

func (m *mockContractRepo) GetByID(_ context.Context, id, _ uuid.UUID) (*domain.Contract, error) {
    if m.getErr != nil {
        return nil, m.getErr
    }
    c, ok := m.contracts[id]
    if !ok {
        return nil, domain.ErrContractNotFound
    }
    return c, nil
}

// Implement remaining interface methods...

type mockAuthzService struct{ allow bool }

func (m *mockAuthzService) Can(_ context.Context, _ authz.Principal, _ string) error {
    if m.allow {
        return nil
    }
    return errors.New("denied")
}

type recordingEventBus struct {
    published []events.Event
}

func (r *recordingEventBus) Publish(_ context.Context, e events.Event) error {
    r.published = append(r.published, e)
    return nil
}
```

## Create: Happy Path

```go
func TestContractService_Create_Success(t *testing.T) {
    repo := newMockRepo()
    events := &recordingEventBus{}
    svc := NewContractService(repo, &mockAuthzService{allow: true}, events, nil, nil, slog.Default())
    sess := testSession()

    contract, err := svc.Create(context.Background(), sess, CreateParams{
        ContractNumber: "CONT-001",
        Title:          "Test Contract",
        TotalValue:     decimal.NewFromFloat(10000),
        Currency:       "USD",
        StartDate:      time.Now(),
        EndDate:        time.Now().AddDate(1, 0, 0),
    })

    require.NoError(t, err)
    assert.Equal(t, "CONT-001", contract.ContractNumber)
    assert.Equal(t, domain.StatusDraft, contract.Status)
    assert.Equal(t, sess.TenantID, contract.TenantID)
}
```

## Create: Permission Denied

```go
func TestContractService_Create_PermissionDenied(t *testing.T) {
    svc := NewContractService(
        newMockRepo(),
        &mockAuthzService{allow: false},
        nil, nil, nil, slog.Default(),
    )

    _, err := svc.Create(context.Background(), testSession(), CreateParams{})
    require.ErrorIs(t, err, domain.ErrForbidden)
}
```

## Submit: State Transition

```go
func TestContractService_Submit_FromDraft(t *testing.T) {
    repo := newMockRepo()
    svc := NewContractService(repo, &mockAuthzService{allow: true}, &recordingEventBus{}, nil, nil, slog.Default())
    sess := testSession()

    // Setup: create a draft contract
    contract, _ := svc.Create(context.Background(), sess, CreateParams{
        ContractNumber: "CONT-001",
        Title:          "Test",
        TotalValue:     decimal.NewFromFloat(1000),
        Currency:       "USD",
    })

    err := svc.Submit(context.Background(), sess, contract.ID, contract.Version)
    require.NoError(t, err)
}

func TestContractService_Submit_NotDraft(t *testing.T) {
    repo := newMockRepo()
    svc := NewContractService(repo, &mockAuthzService{allow: true}, nil, nil, nil, slog.Default())
    sess := testSession()

    // Insert an already-active contract
    activeContract := &domain.Contract{
        ID:       uuid.New(),
        TenantID: sess.TenantID,
        Status:   domain.StatusActive,
        Version:  1,
    }
    repo.contracts[activeContract.ID] = activeContract

    err := svc.Submit(context.Background(), sess, activeContract.ID, 1)
    require.ErrorIs(t, err, domain.ErrContractNotEditable)
}
```

## State Machine Matrix Test

```go
func TestContractStatusTransitions(t *testing.T) {
    cases := []struct {
        from    domain.ContractStatus
        action  func(svc *ContractService, id uuid.UUID, version int) error
        allowed bool
    }{
        {domain.StatusDraft, func(s *ContractService, id uuid.UUID, v int) error {
            return s.Submit(context.Background(), testSession(), id, v)
        }, true},
        {domain.StatusActive, func(s *ContractService, id uuid.UUID, v int) error {
            return s.Submit(context.Background(), testSession(), id, v)
        }, false},
        // ... add all transitions
    }

    for _, tc := range cases {
        t.Run(fmt.Sprintf("%s", tc.from), func(t *testing.T) {
            repo := newMockRepo()
            c := &domain.Contract{ID: uuid.New(), TenantID: testTenantID, Status: tc.from, Version: 1}
            repo.contracts[c.ID] = c

            svc := NewContractService(repo, &mockAuthzService{allow: true}, nil, nil, nil, slog.Default())
            err := tc.action(svc, c.ID, 1)

            if tc.allowed {
                assert.NoError(t, err)
            } else {
                assert.ErrorIs(t, err, domain.ErrContractNotEditable)
            }
        })
    }
}
```

## Feature Flag Guard Test

```go
func TestContractService_Import_FeatureDisabled(t *testing.T) {
    svc := NewContractService(newMockRepo(), &mockAuthzService{allow: true}, nil, nil, nil, slog.Default())

    sess := testSession()
    sess.FeatureFlags = map[string]bool{
        "contracts.bulk_import": false,
    }

    _, err := svc.Import(context.Background(), sess, nil)
    var bizErr *domain.BusinessError
    require.ErrorAs(t, err, &bizErr)
    assert.Equal(t, "FEATURE_NOT_ENABLED", bizErr.Code)
}
```

## Helpers

```go
var testTenantID = uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
var testUserID   = uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")

func testSession() iam.ResolvedSession {
    return iam.ResolvedSession{
        UserID:       testUserID,
        TenantID:     testTenantID,
        Roles:        []string{"contracts.editor"},
        EntityScope:  iam.EntityScope{Type: iam.ScopeAll},
        FeatureFlags: map[string]bool{"contracts.bulk_import": true},
        Settings:     map[string]string{"finance.default_currency": "USD"},
    }
}
```
