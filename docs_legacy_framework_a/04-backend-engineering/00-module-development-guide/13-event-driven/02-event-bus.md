> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Event Bus
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Event-Driven Overview](01-event-driven-overview.md)"
  - "[Publishing Events](03-publishing-events.md)"
  - "[Event Bus Internals](../../../03-platform-architecture/04-event-architecture/02-event-bus-internals.md)"
---

# Event Bus

## Publishing an Event

Events are published asynchronously after a successful write. The service holds a `eventbus.Publisher` dependency:

```go
// internal/core/contracts/service/contract_service.go

func (s *contractService) publishAsync(ctx context.Context, e domain.Event) {
    go func() {
        pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        defer func() {
            if r := recover(); r != nil {
                s.log.Error().Interface("panic", r).Msg("event publish panic recovered")
            }
        }()

        payload, err := json.Marshal(e)
        if err != nil {
            s.log.Error().Err(err).Msg("failed to marshal event payload")
            return
        }

        if err := s.eventBus.Publish(pubCtx, eventbus.Event{
            ID:          uuid.New().String(),
            Topic:       e.Topic(),
            TenantID:    e.GetTenantID().String(),
            PublishedAt: time.Now().UTC(),
            Payload:     payload,
        }); err != nil {
            s.log.Warn().Err(err).Str("topic", e.Topic()).Msg("event publish failed — non-fatal")
        }
    }()
}
```

## Subscribing to Events

Subscribers register handlers at startup via Wire:

```go
// internal/core/finance/subscriptions/contract_subscriptions.go
package subscriptions

import (
    "context"
    "encoding/json"

    "awo.so/internal/platform/eventbus"
    contractEvents "awo.so/internal/core/contracts/domain/events"
    "awo.so/internal/core/finance/service"
)

type ContractEventSubscriber struct {
    financeSvc service.FinanceService
}

func NewContractEventSubscriber(financeSvc service.FinanceService) *ContractEventSubscriber {
    return &ContractEventSubscriber{financeSvc: financeSvc}
}

func (s *ContractEventSubscriber) Register(bus eventbus.Subscriber) error {
    return bus.Subscribe("contracts.activated", s.handleContractActivated)
}

func (s *ContractEventSubscriber) handleContractActivated(ctx context.Context, event eventbus.Event) error {
    var payload contractEvents.ContractActivated
    if err := json.Unmarshal(event.Payload, &payload); err != nil {
        return fmt.Errorf("unmarshal ContractActivated: %w", err)
    }

    return s.financeSvc.CreateContractLiability(ctx, service.CreateLiabilityRequest{
        IdempotencyKey: event.ID,
        TenantID:       uuid.MustParse(payload.TenantID),
        ContractID:     payload.ContractID,
        TotalValue:     payload.TotalValue,
        Currency:       payload.Currency,
    })
}
```

## Outbox Pattern (Transactional Events)

For events that must not be lost (finance, compliance), use the transactional outbox pattern instead of direct async publish:

1. In the same DB transaction that persists the contract, insert a row into `event_outbox`
2. A background relay reads from `event_outbox` and publishes to the bus
3. On successful publish, the outbox row is marked delivered

```go
// In repository — within the same transaction
func (r *contractRepo) CreateWithOutbox(ctx context.Context, params repository.CreateContractParams) (*domain.Contract, error) {
    var contract *domain.Contract
    err := r.store.WithTenant(ctx, params.TenantID, func(q *sqlc.Queries) error {
        // Create contract
        row, err := q.CreateContract(ctx, /* ... */)
        if err != nil {
            return err
        }
        contract = mapContractRowToDomain(row)

        // Write event to outbox in same transaction
        eventPayload, _ := json.Marshal(domain.ContractCreatedEvent{
            ContractID: contract.ID,
            TenantID:   params.TenantID,
        })
        return q.InsertEventOutbox(ctx, sqlc.InsertEventOutboxParams{
            ID:      uuid.New(),
            Topic:   "contracts.created",
            Payload: eventPayload,
        })
    })
    return contract, err
}
```

## Event Bus Implementations

| Backend | Use Case |
|---------|---------|
| In-process (chan) | Development, testing |
| Redis Streams | Single-host, low-latency |
| NATS | Multi-host, moderate scale |
| Kafka | High throughput, audit log |

The interface is identical across backends — swap via Wire binding.
