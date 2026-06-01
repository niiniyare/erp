---
title: Domain Events
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Domain Layer Overview](01-domain-overview.md)"
  - "[Event Patterns Reference](../13-event-driven/06-event-patterns-reference.md)"
  - "[Event Architecture](../../../03-platform-architecture/04-event-architecture/01-event-architecture.md)"
---

# Domain Events

Domain events represent facts that occurred in the domain. They are the primary mechanism for cross-module communication and for triggering async side effects.

## Event Interface

Every event implements:

```go
// awo.so/internal/platform/events
type Event interface {
    Topic() string           // stream/topic name
    GetTenantID() uuid.UUID  // for tenant-scoped routing
}
```

## Defining Events

One event per significant domain fact. Put them in `domain/events.go`:

```go
// internal/core/contracts/domain/events.go
package domain

import (
    "time"
    "github.com/google/uuid"
)

// ContractCreatedEvent fires when a contract transitions from non-existent to draft.
type ContractCreatedEvent struct {
    ContractID     uuid.UUID
    TenantID       uuid.UUID
    EntityID       uuid.UUID
    ContractNumber string
    CreatedBy      uuid.UUID
    OccurredAt     time.Time
}

func (e ContractCreatedEvent) Topic() string          { return "contracts.contract.created" }
func (e ContractCreatedEvent) GetTenantID() uuid.UUID { return e.TenantID }

// ContractSubmittedEvent fires when a contract is submitted for review.
type ContractSubmittedEvent struct {
    ContractID     uuid.UUID
    TenantID       uuid.UUID
    ContractNumber string
    SubmittedBy    uuid.UUID
    OccurredAt     time.Time
}

func (e ContractSubmittedEvent) Topic() string          { return "contracts.contract.submitted" }
func (e ContractSubmittedEvent) GetTenantID() uuid.UUID { return e.TenantID }

// ContractActivatedEvent fires when a contract becomes active.
// Monetary fields are strings to avoid float64 precision issues in JSON.
type ContractActivatedEvent struct {
    ContractID     uuid.UUID
    TenantID       uuid.UUID
    EntityID       uuid.UUID
    ContractNumber string
    TotalValue     string    // decimal string, not decimal.Decimal
    Currency       string
    StartDate      time.Time
    EndDate        time.Time
    ActivatedBy    uuid.UUID
    OccurredAt     time.Time
}

func (e ContractActivatedEvent) Topic() string          { return "contracts.contract.activated" }
func (e ContractActivatedEvent) GetTenantID() uuid.UUID { return e.TenantID }

// ContractTerminatedEvent fires when a contract is terminated.
type ContractTerminatedEvent struct {
    ContractID     uuid.UUID
    TenantID       uuid.UUID
    ContractNumber string
    TotalValue     string
    Currency       string
    TerminatedBy   uuid.UUID
    OccurredAt     time.Time
}

func (e ContractTerminatedEvent) Topic() string          { return "contracts.contract.terminated" }
func (e ContractTerminatedEvent) GetTenantID() uuid.UUID { return e.TenantID }
```

## Monetary Fields as Strings

Monetary values in events **must** be `string`, not `decimal.Decimal`:

```go
// ✅ Correct
TotalValue string  // "50000.000000"

// ❌ Wrong — decimal.Decimal may not deserialize correctly in all consumers
TotalValue decimal.Decimal
```

In the service, convert before publishing:

```go
event := domain.ContractActivatedEvent{
    TotalValue: contract.TotalValue.String(),  // decimal → string
    // ...
}
```

## What Events to Define

Emit an event for every **externally significant** state change — states that downstream modules or external systems may want to react to:

| State Change | Event? | Reason |
|-------------|--------|--------|
| Contract created (draft) | Yes | Finance may need to track |
| Contract submitted | Yes | Workflow may need to start |
| Contract approved | Yes | Budgets may need updating |
| Contract activated | Yes | Finance module integration point |
| Contract suspended | Yes | Purchase orders may need hold |
| Contract terminated | Yes | Finance close-out may be needed |
| Contract field updated (title) | No | Internal change, no external impact |
| Contract line added | No | Contained within contract lifecycle |

## Event vs. Notification

| Concern | Domain Event | Notification |
|---------|------------|--------------|
| Audience | Services (internal) | Users (human) |
| Format | Go struct → JSON | Human-readable text |
| Storage | Event outbox / Redis stream | `notifications` table |
| Trigger | Always on state change | Conditional on user prefs |
| Failure | Retried via outbox | Logged, not retried |

Both are triggered from the service layer after a successful state change.

## Publishing Pattern

Events always publish in a goroutine — never on the hot path:

```go
// In service — after successful repo write
s.publishAsync(ctx, domain.ContractActivatedEvent{
    ContractID:     updated.ID,
    TenantID:       req.TenantID,
    ContractNumber: updated.ContractNumber,
    TotalValue:     updated.TotalValue.String(),
    Currency:       updated.Currency,
    ActivatedBy:    req.UserID,
    OccurredAt:     time.Now().UTC(),
})
```

```go
func (s *contractService) publishAsync(ctx context.Context, e domain.Event) {
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        defer func() {
            if r := recover(); r != nil {
                s.log.Error().Interface("panic", r).Msg("event publish panic")
            }
        }()
        payload, _ := json.Marshal(e)
        if err := s.eventBus.Publish(ctx, eventbus.PublishRequest{
            ID:          uuid.New().String(),
            Topic:       e.Topic(),
            TenantID:    e.GetTenantID().String(),
            PublishedAt: time.Now().UTC(),
            Payload:     payload,
        }); err != nil {
            s.log.Warn().Err(err).Str("topic", e.Topic()).Msg("event publish failed")
        }
    }()
}
```

## Testing Events

Use `RecordingEventBus` to capture published events in unit tests:

```go
type RecordingEventBus struct {
    mu     sync.Mutex
    events []eventbus.PublishRequest
}

func (r *RecordingEventBus) Publish(_ context.Context, req eventbus.PublishRequest) error {
    r.mu.Lock()
    r.events = append(r.events, req)
    r.mu.Unlock()
    return nil
}

func (r *RecordingEventBus) Published() []eventbus.PublishRequest {
    r.mu.Lock()
    defer r.mu.Unlock()
    return append([]eventbus.PublishRequest(nil), r.events...)
}
```

In test:
```go
bus := &RecordingEventBus{}
svc := NewContractService(repo, authz, audit, notif, bus, log, tracer)

_, err := svc.Activate(ctx, req)
require.NoError(t, err)

// Note: publishAsync is a goroutine — allow time for it
time.Sleep(50 * time.Millisecond)

published := bus.Published()
require.Len(t, published, 1)
assert.Equal(t, "contracts.contract.activated", published[0].Topic)
```
