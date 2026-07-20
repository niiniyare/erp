> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Repository Interface
portal: 4 — Backend Engineering
section: 00-module-development-guide/05-repository-layer
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-sqlc-adapter.md
    title: SQLC Adapter
  - path: ./03-with-tenant-pattern.md
    title: WithTenant Pattern
---

# Repository Interface

`repository/<noun>.go` declares the interface (port) that the service depends on. It defines what operations are possible — not how they are implemented. The interface lives in the `repository` sub-package but is dependency-injected into the service, so the service never imports `repository` directly — it receives the interface from Wire.

## Contracts Repository Interface

```go
// internal/core/contracts/repository/contract.go
package repository

import (
	"context"

	"github.com/google/uuid"

	"awo.so/internal/core/contracts/domain"
)

// ContractRepository defines the persistence contract for contracts.
// The service depends on this interface — the SQLC adapter implements it.
// No business logic, no permission checks — pure persistence operations.
type ContractRepository interface {
	// Create persists a new contract and returns the persisted entity.
	// Returns ErrContractAlreadyExists if the contract_number is already in use.
	Create(ctx context.Context, params CreateContractParams) (*domain.Contract, error)

	// GetByID fetches a contract by its primary key within the tenant.
	// Returns ErrContractNotFound if the record does not exist or is soft-deleted.
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Contract, error)

	// List returns a paginated, filtered list of contracts for the tenant.
	// Never returns ErrContractNotFound — an empty result is an empty slice.
	List(ctx context.Context, params ListContractsParams) ([]*domain.Contract, error)

	// Count returns the total number of contracts matching the filter params.
	// Used to compute pagination metadata alongside List.
	Count(ctx context.Context, params ListContractsParams) (int64, error)

	// Update modifies the contract's fields and increments the version.
	// Returns ErrContractNotFound if the ID does not exist.
	// Returns ErrContractConflict if the version does not match (concurrent write).
	Update(ctx context.Context, params UpdateContractParams) (*domain.Contract, error)

	// UpdateStatus transitions the contract to a new status.
	// Returns ErrContractNotFound if the ID does not exist.
	// Returns ErrContractConflict if the version does not match.
	// Does NOT check CanTransitionTo — the service must call that before this.
	UpdateStatus(ctx context.Context, params UpdateContractStatusParams) (*domain.Contract, error)

	// UpdateTotalValue recalculates the contract total from its lines and persists it.
	// Called after any line create/update/delete.
	UpdateTotalValue(ctx context.Context, id, tenantID, updatedBy uuid.UUID) (*domain.Contract, error)

	// Delete soft-deletes the contract and returns the deleted entity.
	// Returns ErrContractNotFound if the ID does not exist or is already deleted.
	// Returns ErrContractConflict if the version does not match.
	Delete(ctx context.Context, id, tenantID, deletedBy uuid.UUID, version int) (*domain.Contract, error)

	// ExistsWithNumber checks whether a contract with the given number exists in the tenant.
	ExistsWithNumber(ctx context.Context, number string, tenantID uuid.UUID) (bool, error)
}

// ContractLineRepository defines persistence for contract lines.
type ContractLineRepository interface {
	Create(ctx context.Context, params CreateContractLineParams) (*domain.ContractLine, error)
	ListByContract(ctx context.Context, contractID, tenantID uuid.UUID) ([]*domain.ContractLine, error)
	Update(ctx context.Context, params UpdateContractLineParams) (*domain.ContractLine, error)
	Delete(ctx context.Context, id, contractID, tenantID, deletedBy uuid.UUID, version int) (*domain.ContractLine, error)
	DeleteByContractID(ctx context.Context, contractID, tenantID, deletedBy uuid.UUID) error
	NextLineNumber(ctx context.Context, contractID, tenantID uuid.UUID) (int, error)
}
```

## Params Structs

Params structs are defined in the `repository` package. They carry the input data for each operation. Use value types (not pointers) for required fields, pointer types for optional fields.

```go
// internal/core/contracts/repository/params.go
package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/contracts/domain"
)

type CreateContractParams struct {
	TenantID       uuid.UUID
	EntityID       uuid.UUID
	ContractNumber string
	Title          string
	Description    string
	VendorID       uuid.UUID
	ContractType   domain.ContractType
	StartDate      time.Time
	EndDate        time.Time
	TotalValue     decimal.Decimal
	Currency       string
	CreatedBy      uuid.UUID
}

type UpdateContractParams struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Title        string
	Description  string
	VendorID     uuid.UUID
	ContractType domain.ContractType
	StartDate    time.Time
	EndDate      time.Time
	TotalValue   decimal.Decimal
	Currency     string
	Version      int
	UpdatedBy    uuid.UUID
}

type UpdateContractStatusParams struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Status    domain.ContractStatus
	Version   int
	UpdatedBy uuid.UUID
}

type ListContractsParams struct {
	TenantID     uuid.UUID
	EntityID     *uuid.UUID
	VendorID     *uuid.UUID
	Status       *domain.ContractStatus
	ContractType *domain.ContractType
	Search       *string
	SortBy       string  // "created_at" | "total_value"
	SortDir      string  // "asc" | "desc"
	PageSize     int
	PageOffset   int
}

type CreateContractLineParams struct {
	TenantID    uuid.UUID
	ContractID  uuid.UUID
	Description string
	Quantity    int
	UnitPrice   decimal.Decimal
	CreatedBy   uuid.UUID
}

type UpdateContractLineParams struct {
	ID          uuid.UUID
	ContractID  uuid.UUID
	TenantID    uuid.UUID
	Description string
	Quantity    int
	UnitPrice   decimal.Decimal
	Version     int
	UpdatedBy   uuid.UUID
}
```

## Interface Design Rules

**No business logic.** The repository does not validate whether a status transition is valid, whether the user has permission, or whether a contract value exceeds a threshold. Those checks belong in the service.

**No SQLC types in the interface.** The interface uses domain types (`domain.Contract`, `domain.ContractStatus`) and standard Go types (`uuid.UUID`, `decimal.Decimal`). The SQLC-generated types are an implementation detail of the adapter.

**Separate methods for separate operations.** `Update` (field changes) and `UpdateStatus` (status transition) are separate methods. This prevents the service from accidentally updating fields during a status change.

**Document the error contract.** Every method's doc comment states what errors it returns. The service relies on this contract to translate errors correctly.
