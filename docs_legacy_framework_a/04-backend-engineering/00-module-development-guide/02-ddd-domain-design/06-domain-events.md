> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Domain Events
portal: 4 — Backend Engineering
section: 00-module-development-guide/02-ddd-domain-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./05-domain-errors.md
    title: Domain Errors
  - path: ../13-event-driven-integration/01-event-overview.md
    title: Event-Driven Integration Overview
---

# Domain Events

`domain/events.go` declares the event types this module emits after successful state changes. Events are immutable records that describe what happened. They are the module's public API for cross-module integration.

## The File

```go
// internal/core/contracts/domain/events.go
package domain

import (
	"time"

	"github.com/google/uuid"
)

// ContractCreated is emitted after a new contract is persisted.
type ContractCreated struct {
	ContractID     uuid.UUID
	TenantID       uuid.UUID
	EntityID       uuid.UUID
	ContractNumber string
	CreatedBy      uuid.UUID
	OccurredAt     time.Time
}

// ContractSubmitted is emitted after a contract transitions to submitted status.
type ContractSubmitted struct {
	ContractID uuid.UUID
	TenantID   uuid.UUID
	EntityID   uuid.UUID
	SubmittedBy uuid.UUID
	OccurredAt time.Time
}

// ContractApproved is emitted after a contract transitions to approved status.
type ContractApproved struct {
	ContractID  uuid.UUID
	TenantID    uuid.UUID
	EntityID    uuid.UUID
	ApprovedBy  uuid.UUID
	TotalValue  string    // serialised decimal string, not float
	Currency    string
	OccurredAt  time.Time
}

// ContractActivated is emitted after a contract transitions to active status.
type ContractActivated struct {
	ContractID  uuid.UUID
	TenantID    uuid.UUID
	EntityID    uuid.UUID
	ActivatedBy uuid.UUID
	StartDate   time.Time
	EndDate     time.Time
	OccurredAt  time.Time
}

// ContractSuspended is emitted after a contract is placed on hold.
type ContractSuspended struct {
	ContractID  uuid.UUID
	TenantID    uuid.UUID
	EntityID    uuid.UUID
	SuspendedBy uuid.UUID
	Reason      string
	OccurredAt  time.Time
}

// ContractTerminated is emitted after a contract is terminated.
type ContractTerminated struct {
	ContractID    uuid.UUID
	TenantID      uuid.UUID
	EntityID      uuid.UUID
	TerminatedBy  uuid.UUID
	Reason        string
	OccurredAt    time.Time
}

// ContractUpdated is emitted after a contract's fields are modified.
// Only emitted for field edits, not status transitions (which have their own events).
type ContractUpdated struct {
	ContractID uuid.UUID
	TenantID   uuid.UUID
	EntityID   uuid.UUID
	UpdatedBy  uuid.UUID
	OccurredAt time.Time
}
```

## Design Rules

### Named in Past Tense

Events describe something that **already happened**. Past tense enforces this semantically.

```go
// CORRECT
type ContractApproved struct { ... }   // happened
type ContractCreated  struct { ... }   // happened

// WRONG
type ApproveContract struct { ... }    // command, not event
type ContractApproval struct { ... }   // noun, ambiguous timing
```

### Always Include TenantID

Every event includes `TenantID`. Events are consumed by other modules and by the audit service. Without `TenantID`, consumers cannot scope their operations correctly.

### One Event Per Significant State Change

Significant = a state change that another module or the UI might care about. Not every field update needs an event. Fine-grained events add noise.

| State change | Emit event? |
|-------------|------------|
| `draft → submitted` | Yes — reviewer needs to know |
| `submitted → under_review` | Yes — submitter needs to know |
| `under_review → approved` | Yes — downstream activation flow |
| `approved → active` | Yes — vendor, finance module care |
| `active → suspended` | Yes — operations team needs notification |
| `active → terminated` | Yes — vendor, finance module care |
| Field edit (title, description) | Emit `ContractUpdated` (single event for all field edits) |

### Monetary Values as Strings

Monetary amounts in events are serialised as `string` (decimal string), not `float64`. Float serialisation is lossy. JSON `"totalValue": "50000.000000"` is exact; `"totalValue": 50000.000001` is not.

### OccurredAt Is Set by the Publisher

The service sets `OccurredAt: time.Now()` when building the event, immediately after the successful write. Do not use `CreatedAt` from the persisted entity — that is the persistence timestamp, not the event publication timestamp. In practice they are nearly identical, but semantically they are different.

## How Events Are Published

Events are published by the service after a successful repository write. The exact transport (in-process bus, NATS, Kafka) is an infrastructure concern injected at the service level. The domain event struct is transport-agnostic.

```go
// internal/core/contracts/service/contract.go
func (s *contractService) Create(ctx context.Context, params repository.CreateContractParams) (*domain.Contract, error) {
	// 1. Persist
	contract, err := s.repo.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	// 2. Emit event (fire-and-forget — never block response on event delivery)
	go func() {
		evt := domain.ContractCreated{
			ContractID:     contract.ID,
			TenantID:       contract.TenantID,
			EntityID:       contract.EntityID,
			ContractNumber: contract.ContractNumber,
			CreatedBy:      contract.CreatedBy,
			OccurredAt:     time.Now(),
		}
		if err := s.eventBus.Publish(ctx, evt); err != nil {
			s.logger.Error().Err(err).Msg("failed to publish ContractCreated event")
		}
	}()

	return contract, nil
}
```

Event publication failure is **logged, not returned**. The create operation succeeded — the contract is persisted. Failing the HTTP response because the event bus hiccupped would be wrong. Consumers use idempotent processing to handle re-delivery.

## Event Consumers

Events emitted by this module may be consumed by:

- **Notification service** — sends approval request to reviewers
- **Finance module** — creates a payment plan when a contract activates
- **Audit service** — archives the event in the audit log
- **Vendor module** — marks the vendor as having an active contract

Consumers subscribe to events by type name. The event struct is the contract. Adding fields is backward-compatible (consumers ignore unknown fields). Removing fields is a breaking change — coordinate with consumers.
