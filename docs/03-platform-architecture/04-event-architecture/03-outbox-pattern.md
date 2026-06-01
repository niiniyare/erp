---
title: Outbox Pattern
portal: 3 — Platform Architecture
section: 04-event-architecture
audience: [architect, backend-engineer, tech-lead]
related:
  - "[Event Architecture Overview](01-event-architecture.md)"
  - "[Event Bus Internals](02-event-bus-internals.md)"
---

# Outbox Pattern

## Problem

If a module publishes an event after committing a database write, two things can go wrong:

1. **Write succeeds, publish fails** — event lost, subscribers never notified
2. **Write fails after publish** — event published for a transaction that was rolled back

The async goroutine approach (used for non-critical events like notifications) accepts this risk. For business-critical events (finance postings, audit records, compliance events), the risk is unacceptable.

## Solution: Transactional Outbox

Write the event to an `event_outbox` table **in the same transaction** as the business data. A background relay reads from the outbox and publishes to the bus. Once published, the outbox row is marked delivered.

```
BEGIN TRANSACTION
    INSERT INTO contracts (...)        ← business write
    INSERT INTO event_outbox (...)     ← event write (same tx)
COMMIT

Background Relay:
    SELECT * FROM event_outbox WHERE delivered_at IS NULL
    → EventBus.Publish(event)
    → UPDATE event_outbox SET delivered_at = now()
```

## Outbox Table

```sql
CREATE TABLE event_outbox (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid        NOT NULL,
    topic        text        NOT NULL,
    payload      jsonb       NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    delivered_at timestamptz,
    attempts     integer     NOT NULL DEFAULT 0,
    last_error   text
);

CREATE INDEX idx_event_outbox_pending
    ON event_outbox (created_at)
    WHERE delivered_at IS NULL;
```

## Repository Usage

```go
// In contract repository — same WithTenant callback as the business write
func (r *contractRepo) CreateWithOutbox(ctx context.Context, p CreateContractParams) (*domain.Contract, error) {
    var contract *domain.Contract
    err := r.store.WithTenant(ctx, p.TenantID, func(q *sqlc.Queries) error {
        // 1. Create contract
        row, err := q.CreateContract(ctx, toSQLCParams(p))
        if err != nil {
            return mapContractDBError(err, "Create")
        }
        contract = mapContractRowToDomain(row)

        // 2. Write event to outbox in same transaction
        payload, _ := json.Marshal(domain.ContractCreatedEvent{
            ContractID:     contract.ID,
            TenantID:       p.TenantID,
            ContractNumber: contract.ContractNumber,
            OccurredAt:     time.Now().UTC(),
        })
        return q.InsertEventOutbox(ctx, sqlc.InsertEventOutboxParams{
            TenantID: p.TenantID,
            Topic:    "contracts.created",
            Payload:  payload,
        })
    })
    return contract, err
}
```

## Relay Process

The outbox relay runs as a background goroutine, polling for undelivered events:

```go
func (r *outboxRelay) Run(ctx context.Context) {
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := r.processBatch(ctx); err != nil {
                r.log.Warn().Err(err).Msg("outbox relay batch failed")
            }
        }
    }
}

func (r *outboxRelay) processBatch(ctx context.Context) error {
    rows, err := r.repo.GetPendingOutboxEvents(ctx, 100)
    if err != nil {
        return err
    }
    for _, row := range rows {
        if err := r.bus.Publish(ctx, toEvent(row)); err != nil {
            _ = r.repo.IncrementOutboxAttempts(ctx, row.ID, err.Error())
            continue
        }
        _ = r.repo.MarkOutboxDelivered(ctx, row.ID)
    }
    return nil
}
```

## When to Use Outbox vs. Direct Async

| Event criticality | Pattern |
|------------------|---------|
| Notification (can be lost) | Direct async goroutine |
| Audit event | Outbox |
| Finance integration event | Outbox |
| Compliance event | Outbox |
| Analytics / reporting | Direct async (acceptable loss) |
