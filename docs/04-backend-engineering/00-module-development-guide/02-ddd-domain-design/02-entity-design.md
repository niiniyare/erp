---
title: Entity Design
portal: 4 — Backend Engineering
section: 00-module-development-guide/02-ddd-domain-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-bounded-context.md
    title: Bounded Context
  - path: ./03-state-machine.md
    title: State Machine
  - path: ./04-value-objects.md
    title: Value Objects
---

# Entity Design

`domain/<noun>.go` is the heart of the module. It contains the primary entity struct, status type, state machine method, and value objects. Nothing else imports this package's internals — the domain is the source of truth.

## Required Fields

Every AwoERP entity struct must have these fields, in this order, with these exact types:

```go
ID        uuid.UUID
TenantID  uuid.UUID   // Layer 1: RLS anchor — every DB call scoped to this
EntityID  uuid.UUID   // Layer 2: org hierarchy scope
// ... module-specific fields ...
Status    <Noun>Status
Version   int         // Optimistic locking
CreatedBy uuid.UUID
UpdatedBy uuid.UUID
DeletedAt *time.Time  // nil = not deleted; non-nil = soft-deleted
CreatedAt time.Time
UpdatedAt time.Time
```

`DeletedAt` is `*time.Time` (pointer), not `time.Time`. A nil pointer means the record is live. A non-nil pointer holds the deletion timestamp.

## Contracts Entity — Complete Example

```go
// internal/core/contracts/domain/contract.go
package domain

import (
	"time"

	"github.com/google/uuid"
)

// ContractStatus represents the lifecycle state of a contract.
type ContractStatus string

const (
	ContractStatusDraft       ContractStatus = "draft"
	ContractStatusSubmitted   ContractStatus = "submitted"
	ContractStatusUnderReview ContractStatus = "under_review"
	ContractStatusApproved    ContractStatus = "approved"
	ContractStatusActive      ContractStatus = "active"
	ContractStatusSuspended   ContractStatus = "suspended"
	ContractStatusTerminated  ContractStatus = "terminated"
)

// Contract is the primary domain entity for the contracts bounded context.
type Contract struct {
	ID             uuid.UUID
	TenantID       uuid.UUID      // RLS anchor
	EntityID       uuid.UUID      // owning organisational unit
	ContractNumber string         // human-readable reference e.g. CONT-2025-0042
	Title          string
	Description    string
	VendorID       uuid.UUID      // FK to vendor module — ID only
	ContractType   ContractType
	StartDate      time.Time
	EndDate        time.Time
	TotalValue     ContractValue  // value object — see value_objects.go
	Currency       string         // ISO 4217 e.g. "USD", "KES"
	Status         ContractStatus
	Version        int
	CreatedBy      uuid.UUID
	UpdatedBy      uuid.UUID
	DeletedAt      *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ContractType classifies the contract category.
type ContractType string

const (
	ContractTypeService   ContractType = "service"
	ContractTypeSupply    ContractType = "supply"
	ContractTypeFramework ContractType = "framework"
	ContractTypeLease     ContractType = "lease"
)

// ContractLine is a child record representing one scope item or deliverable.
type ContractLine struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	ContractID  uuid.UUID
	LineNumber  int
	Description string
	Quantity    int
	UnitPrice   ContractValue  // value object
	TotalPrice  ContractValue  // derived: quantity × unit_price
	Status      ContractLineStatus
	Version     int
	CreatedBy   uuid.UUID
	UpdatedBy   uuid.UUID
	DeletedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ContractLineStatus represents the state of a contract line.
type ContractLineStatus string

const (
	ContractLineStatusActive    ContractLineStatus = "active"
	ContractLineStatusCancelled ContractLineStatus = "cancelled"
)

// CanTransitionTo reports whether transitioning from the current status to next
// is a valid state machine move. Called by the service before every status update.
func (c *Contract) CanTransitionTo(next ContractStatus) bool {
	allowed := map[ContractStatus][]ContractStatus{
		ContractStatusDraft:       {ContractStatusSubmitted},
		ContractStatusSubmitted:   {ContractStatusUnderReview, ContractStatusDraft},
		ContractStatusUnderReview: {ContractStatusApproved, ContractStatusDraft},
		ContractStatusApproved:    {ContractStatusActive, ContractStatusDraft},
		ContractStatusActive:      {ContractStatusSuspended, ContractStatusTerminated},
		ContractStatusSuspended:   {ContractStatusActive, ContractStatusTerminated},
		ContractStatusTerminated:  {},  // terminal state — no exits
	}
	for _, s := range allowed[c.Status] {
		if s == next {
			return true
		}
	}
	return false
}

// IsDeleted reports whether the contract has been soft-deleted.
func (c *Contract) IsDeleted() bool {
	return c.DeletedAt != nil
}

// IsActive reports whether the contract is in the active state.
func (c *Contract) IsActive() bool {
	return c.Status == ContractStatusActive
}

// IsEditable reports whether fields can be modified. Only draft contracts are editable.
func (c *Contract) IsEditable() bool {
	return c.Status == ContractStatusDraft
}
```

## Status Type Convention

Status types use `string` as the underlying type, not `int` or `iota`. This makes them:
- Readable in the database (`WHERE status = 'draft'` not `WHERE status = 2`)
- Safe to log without a string conversion
- Stable across code changes (adding a status doesn't shift values)

```go
// CORRECT
type ContractStatus string
const ContractStatusDraft ContractStatus = "draft"

// WRONG
type ContractStatus int
const (
    ContractStatusDraft ContractStatus = iota  // fragile — order-dependent
)
```

## Naming Conventions for Status Constants

Pattern: `<Entity>Status<PascalCaseState>`

```go
ContractStatusDraft        // not: StatusDraft, ContractDraft, Draft
ContractStatusUnderReview  // multi-word: PascalCase, no underscores
ContractStatusActive
```

## Helper Methods on Entity

Add boolean helper methods for frequently-checked states. Keep them side-effect free.

```go
func (c *Contract) IsEditable() bool { return c.Status == ContractStatusDraft }
func (c *Contract) IsDeleted() bool  { return c.DeletedAt != nil }
func (c *Contract) IsActive() bool   { return c.Status == ContractStatusActive }
```

Do **not** add methods that call external dependencies (databases, services, loggers). The domain entity is a pure value — it must be testable with no infrastructure.

## What Does NOT Go in domain/<noun>.go

| Thing | Where it goes instead |
|-------|----------------------|
| Database queries | `db/queries/<module>.sql` |
| HTTP request parsing | `handlers/request.go` |
| Permission checks | `service/<noun>.go` |
| Logging | `service/<noun>.go` |
| Error HTTP status mapping | `handlers/errors.go` |
| Wire providers | `<module>.go` |
| Notification sending | `service/<noun>.go` |

The domain package has zero external imports except `github.com/google/uuid`, `time`, and `errors`. If you find yourself importing a platform service from the domain package, the logic belongs in the service layer.
