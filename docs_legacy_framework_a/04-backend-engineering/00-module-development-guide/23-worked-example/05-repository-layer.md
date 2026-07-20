> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Worked Example — Repository Layer
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Repository Layer Overview](../05-repository-layer/01-repository-overview.md)"
  - "[Repository Patterns](../05-repository-layer/02-repository-patterns.md)"
  - "[Worked Example Overview](01-worked-example-overview.md)"
---

# Worked Example — Repository Layer

## interface.go

```go
// internal/core/contracts/repository/interface.go
package repository

import (
    "context"
    "github.com/google/uuid"
    "github.com/shopspring/decimal"
    "awo.so/internal/core/contracts/domain"
)

type ContractRepository interface {
    Create(ctx context.Context, p CreateContractParams) (*domain.Contract, error)
    GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Contract, error)
    Update(ctx context.Context, p UpdateContractParams) (*domain.Contract, error)
    UpdateStatus(ctx context.Context, p UpdateContractStatusParams) (*domain.Contract, error)
    Delete(ctx context.Context, id, tenantID uuid.UUID, version int, deletedBy uuid.UUID) (*domain.Contract, error)
    List(ctx context.Context, p ListContractsParams) ([]*domain.Contract, error)
    Count(ctx context.Context, p CountContractsParams) (int, error)
}

type CreateContractParams struct {
    TenantID       uuid.UUID
    EntityID       uuid.UUID
    ContractNumber string
    Title          string
    Description    string
    ContractType   domain.ContractType
    TotalValue     decimal.Decimal
    Currency       string
    StartDate      time.Time
    EndDate        time.Time
    VendorID       uuid.UUID
    AssignedTo     *uuid.UUID
    CreatedBy      uuid.UUID
}

type UpdateContractParams struct {
    ID          uuid.UUID
    TenantID    uuid.UUID
    Title       string
    Description string
    TotalValue  decimal.Decimal
    Currency    string
    StartDate   time.Time
    EndDate     time.Time
    VendorID    uuid.UUID
    AssignedTo  *uuid.UUID
    Version     int
    UpdatedBy   uuid.UUID
}

type UpdateContractStatusParams struct {
    ID        uuid.UUID
    TenantID  uuid.UUID
    Status    domain.ContractStatus
    UpdatedBy uuid.UUID
}

type ListContractsParams struct {
    TenantID uuid.UUID
    Status   *domain.ContractStatus
    EntityID *uuid.UUID
    Limit    int
    Offset   int
}

type CountContractsParams struct {
    TenantID uuid.UUID
    Status   *domain.ContractStatus
    EntityID *uuid.UUID
}
```

## contract_sqlc.go (key methods)

```go
// internal/core/contracts/repository/contract_sqlc.go
package repository

import (
    "context"
    "errors"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    db "awo.so/db/sqlc"
    "awo.so/internal/core/contracts/domain"
    "awo.so/internal/platform/dbutil"
)

type contractRepository struct{ store db.Store }

func NewContractRepository(store db.Store) ContractRepository {
    return &contractRepository{store: store}
}

func (r *contractRepository) Create(ctx context.Context, p CreateContractParams) (*domain.Contract, error) {
    var contract *domain.Contract
    err := r.store.WithTenant(ctx, p.TenantID, func(q *db.Queries) error {
        row, err := q.CreateContract(ctx, db.CreateContractParams{
            TenantID:       p.TenantID,
            EntityID:       p.EntityID,
            ContractNumber: p.ContractNumber,
            Title:          p.Title,
            Description:    p.Description,
            ContractType:   string(p.ContractType),
            TotalValue:     dbutil.DecimalToPgNumeric(p.TotalValue),
            Currency:       p.Currency,
            StartDate:      pgtype.Date{Time: p.StartDate, Valid: true},
            EndDate:        pgtype.Date{Time: p.EndDate, Valid: true},
            VendorID:       p.VendorID,
            AssignedTo:     dbutil.UUIDPtrToNullUUID(p.AssignedTo),
            CreatedBy:      p.CreatedBy,
        })
        if err != nil {
            return mapContractDBError(err, "Create")
        }
        contract = mapContractRowToDomain(row)
        return nil
    })
    return contract, err
}

func (r *contractRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Contract, error) {
    var contract *domain.Contract
    err := r.store.WithTenant(ctx, tenantID, func(q *db.Queries) error {
        row, err := q.GetContractByID(ctx, db.GetContractByIDParams{
            ID: id, TenantID: tenantID,
        })
        if err != nil {
            if errors.Is(err, pgx.ErrNoRows) {
                return domain.ErrContractNotFound
            }
            return mapContractDBError(err, "GetByID")
        }
        contract = mapContractRowToDomain(row)
        return nil
    })
    return contract, err
}

func (r *contractRepository) Update(ctx context.Context, p UpdateContractParams) (*domain.Contract, error) {
    var contract *domain.Contract
    err := r.store.WithTenant(ctx, p.TenantID, func(q *db.Queries) error {
        row, err := q.UpdateContract(ctx, db.UpdateContractParams{
            ID: p.ID, TenantID: p.TenantID,
            Title: p.Title, Version: int32(p.Version),
            // ... other fields ...
        })
        if err != nil {
            if errors.Is(err, pgx.ErrNoRows) {
                // Version conflict or deleted — disambiguate
                exists, _ := q.ContractExistsWithID(ctx, db.ContractExistsWithIDParams{
                    ID: p.ID, TenantID: p.TenantID,
                })
                if exists {
                    return domain.ErrContractConflict
                }
                return domain.ErrContractNotFound
            }
            return mapContractDBError(err, "Update")
        }
        contract = mapContractRowToDomain(row)
        return nil
    })
    return contract, err
}
```

## error_mapping.go

```go
// internal/core/contracts/repository/error_mapping.go
package repository

import (
    "fmt"
    "github.com/jackc/pgx/v5/pgconn"
    "awo.so/internal/core/contracts/domain"
)

func mapContractDBError(err error, op string) error {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        switch pgErr.Code {
        case "23505": // unique_violation
            if strings.Contains(pgErr.ConstraintName, "contract_number") {
                return domain.ErrContractAlreadyExists
            }
        case "23514": // check_violation
            return fmt.Errorf("contracts/%s: check constraint violated: %w", op, err)
        }
    }
    return fmt.Errorf("contracts/%s: %w", op, err)
}
```

**Gate**: Repository written → run integration tests against test DB before proceeding.
