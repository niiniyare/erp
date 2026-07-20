> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Service Interface
portal: 4 — Backend Engineering
section: 00-module-development-guide/06-service-layer
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-service-overview.md
    title: Service Layer Overview
  - path: ./03-write-operations.md
    title: Write Operations
---

# Service Interface

The service interface defines one method per use-case. Method signatures use domain types only — no SQLC types, no HTTP types.

## ContractService Interface

```go
// internal/core/contracts/service/contract.go
package service

import (
	"context"

	"github.com/google/uuid"

	"awo.so/internal/core/contracts/domain"
	"awo.so/internal/core/contracts/repository"
	iam "awo.so/internal/core/iam"
)

// ContractService defines the use-cases for the contracts bounded context.
type ContractService interface {
	// Create creates a new draft contract.
	// Permission: contracts.contract.create
	Create(ctx context.Context, req CreateContractRequest) (*domain.Contract, error)

	// GetByID fetches a contract by primary key within the tenant.
	// Permission: contracts.contract.read
	GetByID(ctx context.Context, id, tenantID uuid.UUID, principal iam.Principal) (*domain.Contract, error)

	// List returns a paginated, filtered list of contracts.
	// Permission: contracts.contract.read
	List(ctx context.Context, params repository.ListContractsParams, principal iam.Principal) (*ListContractsResult, error)

	// Update modifies the fields of a draft contract.
	// Permission: contracts.contract.update
	// Business rule: only draft contracts can be edited (IsEditable check).
	Update(ctx context.Context, req UpdateContractRequest) (*domain.Contract, error)

	// Submit transitions a draft contract to submitted status.
	// Permission: contracts.contract.submit
	Submit(ctx context.Context, id, tenantID uuid.UUID, version int, submittedBy uuid.UUID, principal iam.Principal) (*domain.Contract, error)

	// Approve transitions a submitted/under-review contract to approved status.
	// Permission: contracts.contract.approve
	// Business rule: contract value must not exceed approval threshold.
	Approve(ctx context.Context, id, tenantID uuid.UUID, version int, approvedBy uuid.UUID, principal iam.Principal) (*domain.Contract, error)

	// Activate transitions an approved contract to active status.
	// Permission: contracts.contract.activate
	Activate(ctx context.Context, id, tenantID uuid.UUID, version int, activatedBy uuid.UUID, principal iam.Principal) (*domain.Contract, error)

	// Terminate transitions an active or suspended contract to terminated status.
	// Permission: contracts.contract.terminate
	Terminate(ctx context.Context, req TerminateContractRequest) (*domain.Contract, error)

	// Delete soft-deletes a contract.
	// Permission: contracts.contract.delete
	// Business rule: only draft contracts can be deleted.
	Delete(ctx context.Context, id, tenantID uuid.UUID, version int, deletedBy uuid.UUID, principal iam.Principal) error

	// AddLine adds a line item to a draft contract.
	// Permission: contracts.contract_line.create
	AddLine(ctx context.Context, req AddContractLineRequest) (*domain.ContractLine, error)

	// UpdateLine modifies a line item.
	// Permission: contracts.contract_line.update
	UpdateLine(ctx context.Context, req UpdateContractLineRequest) (*domain.ContractLine, error)

	// RemoveLine soft-deletes a line item.
	// Permission: contracts.contract_line.delete
	RemoveLine(ctx context.Context, lineID, contractID, tenantID uuid.UUID, version int, removedBy uuid.UUID, principal iam.Principal) error
}

// ============================================================
// Request types
// ============================================================

type CreateContractRequest struct {
	TenantID       uuid.UUID
	EntityID       uuid.UUID
	ContractNumber string
	Title          string
	Description    string
	VendorID       uuid.UUID
	ContractType   domain.ContractType
	StartDate      string  // ISO 8601 date string; service parses
	EndDate        string
	Currency       string
	CreatedBy      uuid.UUID
	Principal      iam.Principal
}

type UpdateContractRequest struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Title        string
	Description  string
	VendorID     uuid.UUID
	ContractType domain.ContractType
	StartDate    string
	EndDate      string
	Currency     string
	Version      int
	UpdatedBy    uuid.UUID
	Principal    iam.Principal
}

type TerminateContractRequest struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Version      int
	TerminatedBy uuid.UUID
	Reason       string
	Principal    iam.Principal
}

type AddContractLineRequest struct {
	ContractID  uuid.UUID
	TenantID    uuid.UUID
	Description string
	Quantity    int
	UnitPrice   string  // decimal string; service parses
	CreatedBy   uuid.UUID
	Principal   iam.Principal
}

type UpdateContractLineRequest struct {
	LineID      uuid.UUID
	ContractID  uuid.UUID
	TenantID    uuid.UUID
	Description string
	Quantity    int
	UnitPrice   string
	Version     int
	UpdatedBy   uuid.UUID
	Principal   iam.Principal
}

// ============================================================
// Result types
// ============================================================

type ListContractsResult struct {
	Items      []*domain.Contract
	TotalCount int64
	PageSize   int
	PageOffset int
	TotalPages int
}
```

## Interface Design Rules

**One method per use-case.** `Submit`, `Approve`, `Activate` are separate methods, not a generic `UpdateStatus(newStatus)`. This makes the interface self-documenting and allows different permission strings per transition.

**Principal passed explicitly.** Every method that requires authorization receives a `Principal` parameter. The service calls `authzSvc.Enforce` with this principal. Never stored as service state.

**Request structs for methods with many parameters.** Methods with > 4 parameters use a dedicated request struct (`CreateContractRequest`, `UpdateContractRequest`). Simple lookups (`GetByID`) use positional parameters.

**No HTTP types.** `fiber.Ctx`, `http.Request`, raw JSON bytes — none of these appear in the service interface. The service receives parsed, typed data.

**No database types.** `db.Contract`, `pgtype.UUID`, SQLC-generated structs — none appear in the service interface.
