> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Event Architecture Overview
portal: 3 — Platform Architecture
section: 04-event-architecture
audience: [architect, backend-engineer, tech-lead]
related:
  - "[Event Bus](02-event-bus-internals.md)"
  - "[Outbox Pattern](03-outbox-pattern.md)"
  - "[Event-Driven Guide](../../04-backend-engineering/00-module-development-guide/13-event-driven/01-event-driven-overview.md)"
---

# Event Architecture Overview

## Design Goals

1. **Module decoupling**: publisher does not know who consumes its events
2. **No data loss**: critical events use the transactional outbox pattern
3. **Pluggable transport**: swap from in-process to NATS/Kafka without changing module code
4. **At-least-once delivery**: consumers must be idempotent
5. **Async only**: publishing never blocks the HTTP request

## Event Flow

```
Module publishes event (async goroutine)
    │
    ▼
EventBus.Publish(ctx, Event)
    │
    ├── Development/test: in-process channel delivery
    ├── Production (single host): Redis Streams
    └── Production (multi-host): NATS or Kafka
    │
    ▼
Subscriber.Handler(ctx, Event) called by bus
    │
    ├── Unmarshal payload
    ├── Apply idempotency check (event ID)
    ├── Execute domain logic
    └── Retry on transient error / DLQ on permanent error
```

## Event Envelope

```go
type Event struct {
    ID          string          // UUID — used as idempotency key
    Topic       string          // "contracts.submitted"
    TenantID    string          // tenant UUID string
    PublishedAt time.Time
    Payload     json.RawMessage // serialized domain event struct
}
```

The envelope separates routing concerns (ID, Topic, TenantID) from payload. Consumers unmarshal `Payload` into the specific event struct for that topic.

## Topic Registry

Topics follow `{module}.{event_name}` convention. Each module owns its topics — no two modules publish to the same topic.

```
contracts.created
contracts.submitted
contracts.approved
contracts.activated
contracts.terminated
finance.payment_recorded
finance.ledger_posted
hr.employee_hired
hr.employee_offboarded
inventory.stock_adjusted
```

## Delivery Guarantees

| Guarantee | Provided by | Consumer requirement |
|-----------|------------|---------------------|
| At-least-once | Bus retry on handler error | Idempotent handlers |
| Ordering within topic | Redis Streams / Kafka partition | Consumers process sequentially per partition |
| No data loss on publish failure | Outbox pattern | Relay reads from DB, not from service layer |

## Transport Comparison

| Transport | Throughput | Ordering | Persistence | Use case |
|-----------|-----------|---------|-------------|---------|
| In-process channel | Very high | Yes | No | Dev/test only |
| Redis Streams | High | Per-stream | Yes (configurable) | Single-host production |
| NATS JetStream | Very high | Per-subject | Yes | Multi-host |
| Kafka | Extreme | Per-partition | Yes (configurable) | High-volume analytics |

The `Publisher` and `Subscriber` interfaces are identical across all backends. Swap via Wire binding.
