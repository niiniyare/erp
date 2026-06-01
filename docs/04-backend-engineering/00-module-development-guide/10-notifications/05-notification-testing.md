---
title: Notification Testing
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Sending Notifications](03-sending-notifications.md)"
  - "[Service Testing](../06-service-layer/03-service-testing.md)"
  - "[Testing Overview](../18-testing/01-testing-overview.md)"
---

# Notification Testing

## Mock Implementation

```go
// internal/testutil/notifications.go
package testutil

import (
    "context"
    "sync"

    "github.com/google/uuid"

    "awo.so/internal/core/notifications/port"
)

// RecordingNotifSvc captures all Notify calls for assertion.
// Thread-safe — safe for use in goroutine-firing service tests.
type RecordingNotifSvc struct {
    mu    sync.Mutex
    calls []port.Notification
    done  chan struct{}
    count int
}

func NewRecordingNotifSvc(expectedCalls int) *RecordingNotifSvc {
    return &RecordingNotifSvc{
        done:  make(chan struct{}),
        count: expectedCalls,
    }
}

func (r *RecordingNotifSvc) Notify(_ context.Context, n port.Notification) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.calls = append(r.calls, n)
    if len(r.calls) >= r.count {
        select {
        case <-r.done: // already closed
        default:
            close(r.done)
        }
    }
    return nil
}

func (r *RecordingNotifSvc) NotifyTenant(_ context.Context, _ uuid.UUID, _ string, n port.Notification) error {
    return r.Notify(context.Background(), n)
}

// Wait blocks until expectedCalls notifications have been received or timeout.
func (r *RecordingNotifSvc) Wait(timeout time.Duration) bool {
    select {
    case <-r.done:
        return true
    case <-time.After(timeout):
        return false
    }
}

func (r *RecordingNotifSvc) Calls() []port.Notification {
    r.mu.Lock()
    defer r.mu.Unlock()
    out := make([]port.Notification, len(r.calls))
    copy(out, r.calls)
    return out
}
```

## Service Test: Notification Sent on Submit

```go
func TestContractService_Submit_SendsNotification(t *testing.T) {
    repo   := &mockContractRepo{}
    notif  := testutil.NewRecordingNotifSvc(1)
    authz  := &mockAuthzService{}

    authz.On("Enforce", mock.Anything, mock.Anything).Return(true, nil)
    repo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).
        Return(&domain.Contract{Status: domain.ContractStatusDraft, Version: 1}, nil)
    repo.On("UpdateStatus", mock.Anything, mock.Anything).
        Return(&domain.Contract{
            Status:         domain.ContractStatusSubmitted,
            ContractNumber: "CONT-2025-0001",
        }, nil)

    svc := service.NewContractService(repo, nil, authz, nil, notif, nil, ...)

    _, err := svc.Submit(context.Background(), service.SubmitContractRequest{
        ContractID:     testContractID,
        TenantID:       testTenantID,
        Version:        1,
        UserID:         testUserID,
        Principal:      testPrincipal,
        NotifyOnSubmit: true,
    })
    require.NoError(t, err)

    // Wait up to 100ms for the async goroutine
    ok := notif.Wait(100 * time.Millisecond)
    require.True(t, ok, "notification not delivered within timeout")

    calls := notif.Calls()
    require.Len(t, calls, 1)
    assert.Equal(t, domain.NotifContractSubmitted, calls[0].Category)
    assert.Equal(t, port.AudienceRole, calls[0].Audience)
    assert.Equal(t, "contracts.reviewer", calls[0].Role)
    assert.Equal(t, port.PriorityNormal, calls[0].Priority)
}
```

## Service Test: No Notification When Flag Off

```go
func TestContractService_Submit_NoNotifWhenFlagOff(t *testing.T) {
    notif := testutil.NewRecordingNotifSvc(0)
    // ... setup authz, repo mocks as above ...

    _, err := svc.Submit(context.Background(), service.SubmitContractRequest{
        NotifyOnSubmit: false, // flag off
        // ...
    })
    require.NoError(t, err)

    // Brief wait — if goroutine fired wrongly, we'd catch it
    time.Sleep(20 * time.Millisecond)
    assert.Empty(t, notif.Calls(), "should not notify when flag off")
}
```

## Service Test: Notification Failure Does Not Fail Request

```go
func TestContractService_Submit_NotifErrorNonFatal(t *testing.T) {
    failingNotif := &alwaysFailNotifSvc{}
    // ... setup authz, repo mocks ...

    _, err := svc.Submit(context.Background(), service.SubmitContractRequest{
        NotifyOnSubmit: true,
        // ...
    })

    // Request must succeed even if notification errors
    require.NoError(t, err)
}

type alwaysFailNotifSvc struct{}
func (a *alwaysFailNotifSvc) Notify(_ context.Context, _ port.Notification) error {
    return errors.New("notification service unavailable")
}
func (a *alwaysFailNotifSvc) NotifyTenant(_ context.Context, _ uuid.UUID, _ string, _ port.Notification) error {
    return errors.New("notification service unavailable")
}
```
