> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Consuming Events
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Publishing Events](03-publishing-events.md)"
  - "[Event Bus](02-event-bus.md)"
  - "[Event Patterns Reference](06-event-patterns-reference.md)"
---

# Consuming Events

## Subscription Registration Pattern

Register subscriptions at startup via a dedicated subscriber struct per consumed module:

```go
// internal/core/finance/subscriptions/subscriptions.go
package subscriptions

// FinanceSubscriptions holds all subscriptions for the finance module.
type FinanceSubscriptions struct {
    contractSub *ContractEventSubscriber
}

func NewFinanceSubscriptions(contractSub *ContractEventSubscriber) *FinanceSubscriptions {
    return &FinanceSubscriptions{contractSub: contractSub}
}

// Register registers all finance module subscriptions on the bus.
func (s *FinanceSubscriptions) Register(bus eventbus.Subscriber) error {
    return s.contractSub.Register(bus)
}
```

```go
// In server startup — after Wire
for _, subscriber := range app.Subscribers {
    if err := subscriber.Register(app.EventBus); err != nil {
        log.Fatal().Err(err).Msg("failed to register event subscriptions")
    }
}
```

## Handler Responsibilities

Each event handler must:

1. **Unmarshal** the payload into the expected event struct
2. **Validate** required fields (don't assume payload is well-formed)
3. **Apply idempotency** — use event ID as idempotency key
4. **Return error** only for retryable failures (transient DB errors)
5. **Return nil** for non-retryable cases (already processed, resource not found)

```go
func (s *ContractEventSubscriber) handleContractActivated(ctx context.Context, event eventbus.Event) error {
    // 1. Unmarshal
    var payload contractEvents.ContractActivatedEvent
    if err := json.Unmarshal(event.Payload, &payload); err != nil {
        // Non-retryable: bad payload format
        s.log.Error().Err(err).Str("event_id", event.ID).Msg("failed to unmarshal ContractActivated")
        return nil // return nil so bus doesn't retry
    }

    // 2. Validate
    if payload.ContractID == uuid.Nil {
        s.log.Warn().Str("event_id", event.ID).Msg("ContractActivated missing ContractID — skipping")
        return nil
    }

    // 3. Apply (idempotent)
    err := s.financeSvc.CreateContractLiability(ctx, service.CreateLiabilityRequest{
        IdempotencyKey: event.ID, // deduplicate on this
        TenantID:       payload.TenantID,
        ContractID:     payload.ContractID,
        TotalValue:     payload.TotalValue,
        Currency:       payload.Currency,
    })
    if errors.Is(err, finance.ErrLiabilityAlreadyExists) {
        // Already processed — idempotent success
        return nil
    }
    return err // retryable transient error
}
```

## Dead Letter Queue

Events that fail after N retries move to the dead letter queue. Operations teams monitor the DLQ and replay events after fixing the underlying issue:

```bash
# Replay dead letter events (ops command)
awoctl events replay --dlq contracts.activated --from 2025-01-01
```

## Event Replay Safety

Handlers must tolerate replayed events. Checklist:

- [ ] INSERT uses `ON CONFLICT (idempotency_key) DO NOTHING`
- [ ] UPDATE uses `WHERE NOT processed_event_ids @> ARRAY[idempotency_key]`
- [ ] External API calls deduplicated using event ID as request ID
- [ ] Notifications not re-sent for replayed events (check `notifications` table for prior record)
