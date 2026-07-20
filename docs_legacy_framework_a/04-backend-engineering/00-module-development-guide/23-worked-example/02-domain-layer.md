> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Worked Example — Domain Layer
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Domain Layer Overview](../02-domain-layer/01-domain-overview.md)"
  - "[State Machine Design](../02-domain-layer/02-state-machine-design.md)"
  - "[Worked Example Overview](01-worked-example-overview.md)"
---

# Worked Example — Domain Layer

All domain code lives in `internal/core/contracts/domain/`. No external dependencies except `uuid` and `decimal`.

## Step 1: Create the module directory

```bash
mkdir -p internal/core/contracts/domain
```

## Step 2: contract.go

```go
// internal/core/contracts/domain/contract.go
package domain

import (
    "time"

    "github.com/google/uuid"
    "github.com/shopspring/decimal"
)

type ContractType string

const (
    ContractTypeService ContractType = "service"
    ContractTypeGoods   ContractType = "goods"
    ContractTypeLease   ContractType = "lease"
    ContractTypeOther   ContractType = "other"
)

type Contract struct {
    ID             uuid.UUID
    TenantID       uuid.UUID
    EntityID       uuid.UUID
    ContractNumber string
    Title          string
    Description    string
    Status         ContractStatus
    ContractType   ContractType
    TotalValue     decimal.Decimal
    Currency       string
    StartDate      time.Time
    EndDate        time.Time
    VendorID       uuid.UUID
    AssignedTo     *uuid.UUID
    Version        int
    CreatedBy      uuid.UUID
    UpdatedBy      uuid.UUID
    CreatedAt      time.Time
    UpdatedAt      time.Time
    DeletedAt      *time.Time
}

func (c *Contract) IsEditable() bool {
    return c.Status == ContractStatusDraft
}

func (c *Contract) IsActive() bool {
    return c.Status == ContractStatusActive
}

func (c *Contract) IsDeleted() bool {
    return c.DeletedAt != nil
}

type ContractLine struct {
    ID           uuid.UUID
    TenantID     uuid.UUID
    ContractID   uuid.UUID
    LineNumber   int
    Description  string
    Quantity     decimal.Decimal
    UnitPrice    decimal.Decimal
    TotalPrice   decimal.Decimal // GENERATED ALWAYS AS (quantity * unit_price) STORED
    Unit         string
    Version      int
    CreatedBy    uuid.UUID
    UpdatedBy    uuid.UUID
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    *time.Time
}
```

## Step 3: status.go

```go
// internal/core/contracts/domain/status.go
package domain

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

var contractTransitions = map[ContractStatus][]ContractStatus{
    ContractStatusDraft:       {ContractStatusSubmitted},
    ContractStatusSubmitted:   {ContractStatusUnderReview, ContractStatusDraft},
    ContractStatusUnderReview: {ContractStatusApproved, ContractStatusDraft},
    ContractStatusApproved:    {ContractStatusActive, ContractStatusDraft},
    ContractStatusActive:      {ContractStatusSuspended, ContractStatusTerminated},
    ContractStatusSuspended:   {ContractStatusActive, ContractStatusTerminated},
    ContractStatusTerminated:  {},
}

func (c *Contract) CanTransitionTo(next ContractStatus) bool {
    allowed, ok := contractTransitions[c.Status]
    if !ok {
        return false
    }
    for _, s := range allowed {
        if s == next {
            return true
        }
    }
    return false
}
```

## Step 4: value_objects.go

```go
// internal/core/contracts/domain/value_objects.go
package domain

import (
    "regexp"
    "github.com/shopspring/decimal"
)

var contractNumberRegex = regexp.MustCompile(`^CONT-\d{4}-\d{4,}$`)

type ContractValue struct{ amount decimal.Decimal }

func NewContractValue(amount decimal.Decimal) (ContractValue, error) {
    if amount.IsNegative() {
        return ContractValue{}, ErrContractValueNegative
    }
    return ContractValue{amount: amount}, nil
}

func (v ContractValue) Amount() decimal.Decimal { return v.amount }
func (v ContractValue) IsZero() bool            { return v.amount.IsZero() }

type ContractNumber struct{ value string }

func NewContractNumber(s string) (ContractNumber, error) {
    if s == "" {
        return ContractNumber{}, ErrContractNumberEmpty
    }
    if !contractNumberRegex.MatchString(s) {
        return ContractNumber{}, ErrContractNumberInvalid
    }
    return ContractNumber{value: s}, nil
}

func (n ContractNumber) Value() string { return n.value }
```

## Step 5: errors.go

```go
// internal/core/contracts/domain/errors.go
package domain

import "errors"

var (
    ErrContractNotFound          = errors.New("contract not found")
    ErrContractAlreadyExists     = errors.New("contract already exists")
    ErrContractConflict          = errors.New("contract version conflict")
    ErrContractNotEditable       = errors.New("contract is not editable")
    ErrContractInvalidTransition = errors.New("invalid contract status transition")
    ErrContractValueNegative     = errors.New("contract value cannot be negative")
    ErrContractNumberEmpty       = errors.New("contract number cannot be empty")
    ErrContractNumberInvalid     = errors.New("contract number format invalid (expected CONT-YYYY-NNNN)")
)
```

## Step 6: events.go

```go
// internal/core/contracts/domain/events.go
package domain

import (
    "time"
    "github.com/google/uuid"
)

type Event interface {
    Topic() string
    GetTenantID() uuid.UUID
}

type ContractCreatedEvent struct {
    ContractID     uuid.UUID; TenantID uuid.UUID; ContractNumber string
    EntityID       uuid.UUID; CreatedBy uuid.UUID; OccurredAt time.Time
}
func (e ContractCreatedEvent) Topic() string          { return "contracts.created" }
func (e ContractCreatedEvent) GetTenantID() uuid.UUID { return e.TenantID }

type ContractSubmittedEvent struct {
    ContractID uuid.UUID; TenantID uuid.UUID; ContractNumber string
    SubmittedBy uuid.UUID; OccurredAt time.Time
}
func (e ContractSubmittedEvent) Topic() string          { return "contracts.submitted" }
func (e ContractSubmittedEvent) GetTenantID() uuid.UUID { return e.TenantID }

type ContractApprovedEvent struct {
    ContractID uuid.UUID; TenantID uuid.UUID; ApprovedBy uuid.UUID; OccurredAt time.Time
}
func (e ContractApprovedEvent) Topic() string          { return "contracts.approved" }
func (e ContractApprovedEvent) GetTenantID() uuid.UUID { return e.TenantID }

type ContractActivatedEvent struct {
    ContractID  uuid.UUID; TenantID uuid.UUID; EntityID uuid.UUID
    TotalValue  string; Currency string; ActivatedBy uuid.UUID; OccurredAt time.Time
}
func (e ContractActivatedEvent) Topic() string          { return "contracts.activated" }
func (e ContractActivatedEvent) GetTenantID() uuid.UUID { return e.TenantID }

type ContractTerminatedEvent struct {
    ContractID uuid.UUID; TenantID uuid.UUID; TerminatedBy uuid.UUID; OccurredAt time.Time
}
func (e ContractTerminatedEvent) Topic() string          { return "contracts.terminated" }
func (e ContractTerminatedEvent) GetTenantID() uuid.UUID { return e.TenantID }
```

## Step 7: notification_categories.go

```go
// internal/core/contracts/domain/notification_categories.go
package domain

const (
    NotifContractSubmitted    = "contracts.submitted"
    NotifContractUnderReview  = "contracts.under_review"
    NotifContractApproved     = "contracts.approved"
    NotifContractRejected     = "contracts.rejected"
    NotifContractActivated    = "contracts.activated"
    NotifContractSuspended    = "contracts.suspended"
    NotifContractTerminated   = "contracts.terminated"
    NotifContractExpiringSoon = "contracts.expiring_soon"
)
```

**Gate**: Domain files written → run `go vet ./internal/core/contracts/domain/...` before proceeding.
