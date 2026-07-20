> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Event Testing
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Publishing Events](03-publishing-events.md)"
  - "[Consuming Events](04-consuming-events.md)"
  - "[Event Patterns Reference](06-event-patterns-reference.md)"
---

# Event Testing

## In-Process Test Bus

Use the in-process event bus for service and handler tests:

```go
// internal/testutil/eventbus.go
package testutil

import (
    "context"
    "sync"

    "awo.so/internal/platform/eventbus"
)

// RecordingEventBus captures published events for assertion.
type RecordingEventBus struct {
    mu     sync.Mutex
    events []eventbus.Event
}

func (b *RecordingEventBus) Publish(_ context.Context, event eventbus.Event) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.events = append(b.events, event)
    return nil
}

func (b *RecordingEventBus) Events() []eventbus.Event {
    b.mu.Lock()
    defer b.mu.Unlock()
    out := make([]eventbus.Event, len(b.events))
    copy(out, b.events)
    return out
}

func (b *RecordingEventBus) EventsForTopic(topic string) []eventbus.Event {
    var out []eventbus.Event
    for _, e := range b.Events() {
        if e.Topic == topic {
            out = append(out, e)
        }
    }
    return out
}
```

## Service Test: Event Published After Create

```go
func TestContractService_Create_PublishesEvent(t *testing.T) {
    bus := &testutil.RecordingEventBus{}
    // ... setup mocks ...

    svc := service.NewContractService(repo, nil, authz, nil, nil, bus, ...)
    _, err := svc.Create(context.Background(), service.CreateContractRequest{
        // ...
    })
    require.NoError(t, err)

    // Wait briefly for async goroutine
    time.Sleep(20 * time.Millisecond)

    events := bus.EventsForTopic("contracts.created")
    require.Len(t, events, 1)
    assert.Equal(t, testTenantID.String(), events[0].TenantID)

    var payload domain.ContractCreatedEvent
    require.NoError(t, json.Unmarshal(events[0].Payload, &payload))
    assert.NotEqual(t, uuid.Nil, payload.ContractID)
}
```

## Consumer Handler Test: Idempotency

```go
func TestHandleContractActivated_Idempotent(t *testing.T) {
    store    := testutil.NewTestStore(t)
    financeSvc := finance.NewService(/* ... */)
    sub := subscriptions.NewContractEventSubscriber(financeSvc, log.Logger)

    eventID := uuid.New().String()
    event := eventbus.Event{
        ID:    eventID,
        Topic: "contracts.activated",
        Payload: mustMarshal(domain.ContractActivatedEvent{
            ContractID: uuid.New(),
            TenantID:   testutil.TestTenantID,
            TotalValue: "50000.000000",
            Currency:   "USD",
        }),
    }

    // First call — should succeed
    err := sub.HandleContractActivated(context.Background(), event)
    require.NoError(t, err)

    // Second call (replay) — must not error or double-create
    err = sub.HandleContractActivated(context.Background(), event)
    require.NoError(t, err)

    // Verify only one liability created
    liabilities, err := financeRepo.ListLiabilitiesByIdempotencyKey(context.Background(), eventID)
    require.NoError(t, err)
    assert.Len(t, liabilities, 1)
}
```

## Consumer Handler Test: Bad Payload Doesn't Retry

```go
func TestHandleContractActivated_BadPayload_ReturnsNil(t *testing.T) {
    sub := subscriptions.NewContractEventSubscriber(/* ... */)

    event := eventbus.Event{
        ID:      uuid.New().String(),
        Topic:   "contracts.activated",
        Payload: []byte(`{invalid json`),
    }

    err := sub.HandleContractActivated(context.Background(), event)
    assert.NoError(t, err, "bad payload should return nil (not retry)")
}
```
