> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Publishing Events
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Event Bus](02-event-bus.md)"
  - "[Domain Events](../02-domain-layer/03-domain-events.md)"
  - "[Outbox Pattern](../../../03-platform-architecture/04-event-architecture/03-outbox-pattern.md)"
---

# Publishing Events

## Event Interface

Every domain event implements the `Event` interface so the service can publish it generically:

```go
// internal/core/contracts/domain/events.go
package domain

import "github.com/google/uuid"

// Event is the interface all domain events implement.
type Event interface {
    Topic() string
    GetTenantID() uuid.UUID
}
```

## Publishing After Create

```go
func (s *contractService) Create(ctx context.Context, req CreateContractRequest) (*domain.Contract, error) {
    // ... authorize, validate, persist ...

    contract, err := s.repo.Create(ctx, /* params */)
    if err != nil {
        return nil, err
    }

    s.publishAsync(ctx, domain.ContractCreatedEvent{
        ContractID:     contract.ID,
        TenantID:       req.TenantID,
        ContractNumber: contract.ContractNumber,
        EntityID:       req.EntityID,
        CreatedBy:      req.UserID,
        OccurredAt:     time.Now().UTC(),
    })

    return contract, nil
}
```

## Publishing After Status Transition

```go
func (s *contractService) Submit(ctx context.Context, req SubmitContractRequest) (*domain.Contract, error) {
    // ... authorize, fetch, state machine, persist ...

    s.publishAsync(ctx, domain.ContractSubmittedEvent{
        ContractID:     updated.ID,
        TenantID:       req.TenantID,
        ContractNumber: updated.ContractNumber,
        SubmittedBy:    req.UserID,
        OccurredAt:     time.Now().UTC(),
    })

    return updated, nil
}
```

## Complete Events File

```go
// internal/core/contracts/domain/events.go
package domain

import (
    "time"
    "github.com/google/uuid"
)

type ContractCreatedEvent struct {
    ContractID     uuid.UUID
    TenantID       uuid.UUID
    ContractNumber string
    EntityID       uuid.UUID
    CreatedBy      uuid.UUID
    OccurredAt     time.Time
}
func (e ContractCreatedEvent) Topic() string          { return "contracts.created" }
func (e ContractCreatedEvent) GetTenantID() uuid.UUID { return e.TenantID }

type ContractSubmittedEvent struct {
    ContractID     uuid.UUID
    TenantID       uuid.UUID
    ContractNumber string
    SubmittedBy    uuid.UUID
    OccurredAt     time.Time
}
func (e ContractSubmittedEvent) Topic() string          { return "contracts.submitted" }
func (e ContractSubmittedEvent) GetTenantID() uuid.UUID { return e.TenantID }

type ContractApprovedEvent struct {
    ContractID  uuid.UUID
    TenantID    uuid.UUID
    ApprovedBy  uuid.UUID
    OccurredAt  time.Time
}
func (e ContractApprovedEvent) Topic() string          { return "contracts.approved" }
func (e ContractApprovedEvent) GetTenantID() uuid.UUID { return e.TenantID }

type ContractActivatedEvent struct {
    ContractID  uuid.UUID
    TenantID    uuid.UUID
    EntityID    uuid.UUID
    TotalValue  string  // decimal string — never float
    Currency    string
    ActivatedBy uuid.UUID
    OccurredAt  time.Time
}
func (e ContractActivatedEvent) Topic() string          { return "contracts.activated" }
func (e ContractActivatedEvent) GetTenantID() uuid.UUID { return e.TenantID }

type ContractTerminatedEvent struct {
    ContractID    uuid.UUID
    TenantID      uuid.UUID
    TerminatedBy  uuid.UUID
    Reason        string
    OccurredAt    time.Time
}
func (e ContractTerminatedEvent) Topic() string          { return "contracts.terminated" }
func (e ContractTerminatedEvent) GetTenantID() uuid.UUID { return e.TenantID }
```

## Publishing Rules

1. Always publish **after** the database write succeeds — never before
2. Always fire in a goroutine — never block the request
3. Always include `TenantID` and `OccurredAt` in every event
4. Monetary values as **decimal strings** — never float64
5. IDs as `uuid.UUID` — unmarshal on the subscriber side
6. Event failure is **non-fatal** — log warn, do not return error to caller
