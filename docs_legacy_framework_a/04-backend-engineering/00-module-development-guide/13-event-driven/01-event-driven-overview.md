> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Event-Driven Overview
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Event Patterns Reference](06-event-patterns-reference.md)"
  - "[Domain Layer](../02-domain-layer/01-domain-overview.md)"
  - "[Redis Architecture](../../../03-platform-architecture/08-redis-architecture/01-redis-overview.md)"
---

# Event-Driven Overview

The event-driven layer enables modules to communicate **without direct imports** of each other's service interfaces. A module publishes events; interested modules subscribe and react.

## Why Events

| Concern | Direct Call | Event |
|---------|------------|-------|
| Coupling | Tight — importer depends on exportee | Loose — publisher doesn't know subscribers |
| Failure isolation | Subscriber failure blocks publisher | Subscriber failure doesn't affect publisher |
| Fan-out | N service calls needed | One publish, N subscribers |
| Latency | Synchronous, adds to request time | Asynchronous, request returns immediately |
| Ordering | Guaranteed | At-least-once, unordered by default |

## Event Bus Interface

```go
// internal/platform/eventbus/eventbus.go
package eventbus

import "context"

// Publisher publishes events to the bus.
type Publisher interface {
    Publish(ctx context.Context, event Event) error
}

// Subscriber receives events from the bus.
type Subscriber interface {
    Subscribe(topic string, handler EventHandler) error
}

// EventHandler processes a received event.
type EventHandler func(ctx context.Context, event Event) error

// Event is the envelope wrapping a domain event payload.
type Event struct {
    ID          string          // UUID
    Topic       string          // e.g. "contracts.submitted"
    TenantID    string
    PublishedAt time.Time
    Payload     json.RawMessage // serialized domain event
}
```

## Topic Naming

Topics follow the same naming as notification categories:

```
{module}.{event_name}
```

Examples:
```
contracts.submitted
contracts.approved
contracts.terminated
finance.payment_recorded
hr.employee_offboarded
```

## Contracts Events Published

| Topic | Published When |
|-------|---------------|
| `contracts.created` | Contract created (draft) |
| `contracts.submitted` | Contract submitted for review |
| `contracts.approved` | Contract approved |
| `contracts.activated` | Contract activated |
| `contracts.terminated` | Contract terminated |
| `contracts.value_changed` | Total value updated |

## Consumers of Contract Events

| Consumer Module | Topic | Action Taken |
|----------------|-------|-------------|
| Finance | `contracts.activated` | Create liability entry |
| Finance | `contracts.terminated` | Close liability entry |
| HR | `contracts.activated` | Link to employee record if service contract |
| Audit | All | Archive event snapshot |
| Reporting | All | Update analytics cache |

## Delivery Guarantees

The event bus uses **at-least-once delivery**. Subscribers must be idempotent — they may receive the same event more than once.

Use the event `ID` (UUID) as an idempotency key when the handler writes to a database:

```go
func handleContractActivated(ctx context.Context, event eventbus.Event) error {
    var payload events.ContractActivated
    if err := json.Unmarshal(event.Payload, &payload); err != nil {
        return err
    }

    // Idempotent insert: skip if already processed
    return financeRepo.CreateLiabilityIfNotExists(ctx, finance.CreateLiabilityParams{
        IdempotencyKey: event.ID,
        ContractID:     payload.ContractID,
        // ...
    })
}
```
