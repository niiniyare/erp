---
title: Testing the Audit Trail
portal: 4 — Backend Engineering
section: 00-module-development-guide/09-audit-trail
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-audit-overview.md
    title: Audit Trail Overview
  - path: ../22-testing-guide/01-testing-overview.md
    title: Testing Overview
---

# Testing the Audit Trail

Audit recording happens asynchronously, which makes it tricky to test. This page shows the strategies used in AwoERP integration tests.

## Strategy: Synchronous Audit in Tests

In integration tests, use a synchronous audit service that records immediately (no goroutine):

```go
// test helper
type syncAuditService struct {
	events []audit.Event
	mu     sync.Mutex
}

func (s *syncAuditService) Record(ctx context.Context, evt audit.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, evt)
	return nil
}

func (s *syncAuditService) LastEvent() *audit.Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.events) == 0 {
		return nil
	}
	return &s.events[len(s.events)-1]
}
```

Inject the synchronous audit service into the service under test:

```go
func TestContractService_Create_AuditRecorded(t *testing.T) {
	auditSvc := &syncAuditService{}
	svc := service.NewContractService(
		mockRepo,
		nil,
		auditSvc,
		// ...
	)

	contract, err := svc.Create(ctx, createParams)
	require.NoError(t, err)

	// In-process goroutine may not have run yet — wait briefly
	time.Sleep(10 * time.Millisecond)

	evt := auditSvc.LastEvent()
	require.NotNil(t, evt, "expected audit event to be recorded")
	assert.Equal(t, "contract.create", evt.Action)
	assert.Equal(t, contract.ID.String(), evt.ResourceID)
	assert.Equal(t, contract.TenantID, evt.TenantID)
	assert.Nil(t, evt.Before, "Before should be nil on create")
	assert.NotNil(t, evt.After, "After should be set on create")
}
```

The `time.Sleep(10 * time.Millisecond)` is a pragmatic wait for the goroutine. In CI environments, this is sufficient. For tighter control, make the goroutine dispatch configurable (synchronous in tests).

## Configurable Async Dispatch

A cleaner pattern: make the goroutine dispatch injectable:

```go
type contractService struct {
	// ...
	dispatch func(fn func())  // injectable for tests
}

func NewContractService(...) ContractService {
	return &contractService{
		// ...
		dispatch: func(fn func()) { go fn() },  // async by default
	}
}

// In tests
svc.dispatch = func(fn func()) { fn() }  // synchronous in tests
```

With synchronous dispatch, no `time.Sleep` needed — audit events are recorded before the service method returns.

## What to Assert

For each write operation, assert:

1. Audit event is recorded (not nil).
2. `Action` matches the operation (`"contract.create"`, `"contract.submit"`, etc.).
3. `ResourceID` matches the created/updated resource.
4. `TenantID` matches the request tenant.
5. `ActorID` matches the actor.
6. `Before` is nil on create; non-nil on update/delete.
7. `After` is non-nil on create/update; nil on delete.
