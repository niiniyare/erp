> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Domain Layer Overview
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[What This Guide Covers](../01-overview/01-what-this-guide-covers.md)"
  - "[State Machine Design](02-state-machine-design.md)"
  - "[Domain Events](03-domain-events.md)"
  - "[Domain Errors](04-domain-errors.md)"
  - "[Repository Layer](../05-repository-layer/01-repository-overview.md)"
  - "[Worked Example: Domain Layer](../23-worked-example/02-domain-layer.md)"
---

# Domain Layer Overview

## Purpose

The domain layer defines the core business concepts — structs, status types, errors, events, and value objects. It has **no dependencies** on external packages (no DB, no HTTP, no wire).

## Package Layout

```
internal/core/contracts/domain/
  contract.go               ← Contract struct, ContractLine struct
  status.go                 ← ContractStatus type and constants
  value_objects.go          ← ContractType, Currency, etc.
  errors.go                 ← sentinel errors + BusinessError
  events.go                 ← ContractSubmitted, ContractApproved, etc.
  notification_categories.go ← notification category constants
```

## Domain Struct

```go
// domain/contract.go
package domain

import (
    "time"
    "github.com/google/uuid"
    "github.com/shopspring/decimal"
)

type Contract struct {
    ID             uuid.UUID
    TenantID       uuid.UUID
    ContractNumber string
    Title          string
    Description    string
    Status         ContractStatus
    ContractType   ContractType
    TotalValue     decimal.Decimal   // never float64
    Currency       string
    StartDate      time.Time
    EndDate        time.Time
    VendorID       uuid.UUID
    Version        int               // optimistic lock
    CreatedAt      time.Time
    UpdatedAt      time.Time
    DeletedAt      *time.Time        // nil = not deleted
}
```

**Rules:**
- Monetary values: `decimal.Decimal` — never `float64`
- Status: named string type — never `int`/`iota`
- PKs: `uuid.UUID` — never `int64`
- Soft delete: `DeletedAt *time.Time`
- Optimistic lock: `Version int`

## Status Type

```go
// domain/status.go
type ContractStatus string

const (
    StatusDraft       ContractStatus = "draft"
    StatusUnderReview ContractStatus = "under_review"
    StatusApproved    ContractStatus = "approved"
    StatusActive      ContractStatus = "active"
    StatusSuspended   ContractStatus = "suspended"
    StatusTerminated  ContractStatus = "terminated"
)

// CanTransitionTo validates the state machine
func (s ContractStatus) CanTransitionTo(next ContractStatus) bool {
    allowed := map[ContractStatus][]ContractStatus{
        StatusDraft:       {StatusUnderReview},
        StatusUnderReview: {StatusApproved, StatusDraft},
        StatusApproved:    {StatusActive, StatusDraft},
        StatusActive:      {StatusSuspended, StatusTerminated},
        StatusSuspended:   {StatusActive, StatusTerminated},
        StatusTerminated:  {},  // terminal
    }
    for _, s2 := range allowed[s] {
        if s2 == next {
            return true
        }
    }
    return false
}
```

## Value Objects

Small types that provide type safety:

```go
// domain/value_objects.go
type ContractType string

const (
    ContractTypeService   ContractType = "service"
    ContractTypeSupply    ContractType = "supply"
    ContractTypeLicense   ContractType = "license"
    ContractTypeFramework ContractType = "framework"
)

func ParseContractType(s string) (ContractType, error) {
    switch ContractType(s) {
    case ContractTypeService, ContractTypeSupply, ContractTypeLicense, ContractTypeFramework:
        return ContractType(s), nil
    default:
        return "", fmt.Errorf("unknown contract type: %q", s)
    }
}
```

## Errors

```go
// domain/errors.go
var (
    ErrContractNotFound    = errors.New("contract not found")
    ErrContractNotEditable = errors.New("contract is not in editable state")
    ErrVersionConflict     = errors.New("version conflict")
    ErrForbidden           = errors.New("permission denied")
    ErrDuplicateNumber     = errors.New("contract number already exists")
)

type BusinessError struct {
    Code    string
    Message string
    Status  int
    Err     error
}

func (e *BusinessError) Error() string { return e.Message }
func (e *BusinessError) Unwrap() error { return e.Err }
```

## Domain Events

```go
// domain/events.go
type ContractSubmitted struct {
    ContractID     uuid.UUID
    TenantID       uuid.UUID
    ContractNumber string
    TotalValue     string   // decimal string, not decimal.Decimal
    SubmittedByID  uuid.UUID
    OccurredAt     time.Time
}

func (e ContractSubmitted) Topic() string        { return "contracts.submitted" }
func (e ContractSubmitted) GetTenantID() uuid.UUID { return e.TenantID }
```

Monetary values in events are `string` (decimal representation) to ensure JSON-safe transport without precision loss.

## What Does NOT Belong in Domain

| Does NOT belong | Put it in |
|----------------|-----------|
| DB queries | Repository layer |
| HTTP types (request/response DTOs) | Handler layer |
| `pgx` types | Repository layer |
| `fiber.Ctx` | Handler layer |
| Wire providers | `wire.go` |
| Config reading | Config struct |

The domain package imports only standard library and `uuid`/`decimal`. If you find yourself importing `pgx`, `fiber`, or `slog` in domain — stop and restructure.
