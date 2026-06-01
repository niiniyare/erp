---
title: Event Patterns Quick Reference
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Event Driven Overview](01-event-driven-overview.md)"
  - "[Event Bus](02-event-bus.md)"
  - "[Publishing Events](03-publishing-events.md)"
  - "[Consuming Events](04-consuming-events.md)"
---

# Event Patterns Quick Reference

## Define an Event

```go
// domain/events.go
type ContractApproved struct {
    ContractID     uuid.UUID
    TenantID       uuid.UUID
    ContractNumber string
    ApprovedByID   uuid.UUID
    TotalValue     string   // decimal string
    OccurredAt     time.Time
}

func (e ContractApproved) Topic() string          { return "contracts.approved" }
func (e ContractApproved) GetTenantID() uuid.UUID { return e.TenantID }
```

Rules:
- Monetary values: `string` (decimal representation)
- UUIDs: `uuid.UUID`
- Timestamps: `time.Time`
- Immutable — no pointer receivers

## Publish Async (After HTTP Response)

```go
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    defer func() { recover() }()
    if err := s.events.Publish(ctx, domain.ContractApproved{
        ContractID:     contract.ID,
        TenantID:       sess.TenantID,
        ContractNumber: contract.ContractNumber,
        ApprovedByID:   sess.UserID,
        TotalValue:     contract.TotalValue.String(),
        OccurredAt:     time.Now(),
    }); err != nil {
        s.logger.Error("publish failed", "topic", "contracts.approved", "error", err)
    }
}()
```

## Publish via Outbox (Critical Events)

```go
// Within store.WithTenant transaction
err := s.store.WithTenant(ctx, tenantID, func(tx *pgx.Tx) error {
    // Business write
    contract, err := q.UpdateContractStatus(ctx, ...)
    if err != nil {
        return err
    }
    // Outbox write — same transaction
    _, err = q.InsertOutboxEvent(ctx, sqlc.InsertOutboxEventParams{
        TenantID: tenantID,
        Topic:    "contracts.approved",
        Payload:  toJSON(domain.ContractApproved{...}),
    })
    return err
})
```

## Subscribe to an Event

```go
// In module wire.go init or worker registration
func RegisterContractSubscriptions(bus events.Bus, handler *ContractEventHandler) {
    bus.Subscribe("contracts.submitted", handler.OnContractSubmitted)
    bus.Subscribe("contracts.approved", handler.OnContractApproved)
}
```

## Handler Implementation

```go
type ContractEventHandler struct {
    notif notifications.Service
    logger *slog.Logger
}

func (h *ContractEventHandler) OnContractApproved(ctx context.Context, raw []byte) error {
    var event domain.ContractApproved
    if err := json.Unmarshal(raw, &event); err != nil {
        h.logger.Error("unmarshal failed", "topic", "contracts.approved", "error", err)
        return nil   // return nil — bad payload, don't retry
    }

    // Handler must be idempotent
    return h.notif.Send(ctx, notifications.Notification{
        TenantID:   event.TenantID,
        Category:   notifications.CategoryContractApproved,
        Title:      "Contract Approved",
        Body:       fmt.Sprintf("Contract %s has been approved.", event.ContractNumber),
        ResourceID: event.ContractID,
    })
}
```

## Topic Naming

```
{module}.{entity}_{past_tense_verb}

contracts.submitted
contracts.approved
contracts.activated
contracts.terminated
finance.transaction_posted
iam.user_created
iam.role_assigned
tenant.activated
tenant.suspended
```

## When to Use Outbox vs Async

| Scenario | Use |
|----------|-----|
| Event must not be lost (payment, compliance) | Outbox pattern |
| Event is informational (notification, cache update) | Async goroutine |
| Consumer must process in order | Outbox (ordered by `created_at`) |
| Consumer failure is acceptable | Async goroutine |
| External system integration | Outbox |

## Event Bus Interface

```go
type Publisher interface {
    Publish(ctx context.Context, event Event) error
}

type Subscriber interface {
    Subscribe(topic string, handler HandlerFunc) error
}

type Bus interface {
    Publisher
    Subscriber
}

type Event interface {
    Topic() string
    GetTenantID() uuid.UUID
}

type HandlerFunc func(ctx context.Context, payload []byte) error
```

## Test Double

```go
type RecordingEventBus struct {
    mu     sync.Mutex
    events []events.Event
}

func (r *RecordingEventBus) Publish(_ context.Context, e events.Event) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.events = append(r.events, e)
    return nil
}

func (r *RecordingEventBus) Published(topic string) []events.Event {
    r.mu.Lock()
    defer r.mu.Unlock()
    var result []events.Event
    for _, e := range r.events {
        if e.Topic() == topic {
            result = append(result, e)
        }
    }
    return result
}
```
