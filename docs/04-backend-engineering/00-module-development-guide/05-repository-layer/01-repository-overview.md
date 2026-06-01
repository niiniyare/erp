---
title: Repository Layer Overview
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[SQLC Layer](../04-sqlc-layer/01-sqlc-overview.md)"
  - "[Service Layer](../06-service-layer/01-service-overview.md)"
  - "[RLS and Tenant Isolation](02-rls-and-tenant-isolation.md)"
  - "[Worked Example: Repository Layer](../23-worked-example/05-repository-layer.md)"
---

# Repository Layer Overview

## Purpose

The repository layer:
- Translates domain params → SQLC params
- Translates SQLC rows → domain structs
- Maps DB errors → domain errors
- Wraps all queries in `WithTenant` for RLS

The service layer never sees `pgx`, `sqlc`, or SQL. It only sees domain types.

## Package Layout

```
internal/core/contracts/repository/
  interface.go          ← ContractRepository interface + params structs
  contract_sqlc.go      ← sqlcContractRepo implementation
  error_mapping.go      ← parseContractDBError()
  mappers.go            ← sqlc row → domain struct
```

## Repository Interface

Define in `interface.go`:

```go
// interface.go
package repository

import (
    "context"
    "github.com/google/uuid"
    "awo.so/internal/core/contracts/domain"
)

type ContractRepository interface {
    Create(ctx context.Context, tenantID uuid.UUID, params CreateParams) (*domain.Contract, error)
    GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Contract, error)
    List(ctx context.Context, tenantID uuid.UUID, params ListParams) ([]domain.Contract, int64, error)
    Update(ctx context.Context, id, tenantID uuid.UUID, params UpdateParams) (*domain.Contract, error)
    UpdateStatus(ctx context.Context, id, tenantID uuid.UUID, params UpdateStatusParams) (*domain.Contract, error)
    SoftDelete(ctx context.Context, id, tenantID uuid.UUID, version int) error
}

type CreateParams struct {
    ContractNumber string
    Title          string
    ContractType   domain.ContractType
    TotalValue     decimal.Decimal
    Currency       string
    StartDate      time.Time
    EndDate        time.Time
    VendorID       *uuid.UUID
    Description    string
}

type ListParams struct {
    Status  *domain.ContractStatus
    Limit   int32
    Offset  int32
}

type UpdateStatusParams struct {
    Status  domain.ContractStatus
    Version int
}
```

## Implementation

```go
// contract_sqlc.go
type sqlcContractRepo struct {
    store *db.Store
}

func NewContractRepo(store *db.Store) ContractRepository {
    return &sqlcContractRepo{store: store}
}

func (r *sqlcContractRepo) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Contract, error) {
    var contract *domain.Contract
    err := r.store.WithTenant(ctx, tenantID, func(tx *pgx.Tx) error {
        q := sqlc.New(tx)
        row, err := q.GetContractByID(ctx, id)
        if err != nil {
            return parseContractDBError(err, "GetByID")
        }
        contract = toDomain(row)
        return nil
    })
    return contract, err
}

func (r *sqlcContractRepo) Create(ctx context.Context, tenantID uuid.UUID, params CreateParams) (*domain.Contract, error) {
    var contract *domain.Contract
    err := r.store.WithTenant(ctx, tenantID, func(tx *pgx.Tx) error {
        q := sqlc.New(tx)
        row, err := q.CreateContract(ctx, sqlc.CreateContractParams{
            ContractNumber: params.ContractNumber,
            Title:          params.Title,
            ContractType:   string(params.ContractType),
            TotalValue:     params.TotalValue,
            Currency:       params.Currency,
            StartDate:      pgtype.Date{Time: params.StartDate, Valid: true},
            EndDate:        pgtype.Date{Time: params.EndDate, Valid: true},
            Description:    pgtype.Text{String: params.Description, Valid: params.Description != ""},
        })
        if err != nil {
            return parseContractDBError(err, "Create")
        }
        contract = toDomain(row)
        return nil
    })
    return contract, err
}
```

## WithTenant on Every Call

**Every** repository method must call `store.WithTenant(ctx, tenantID, ...)`. No exceptions.

Without `WithTenant`, `current_setting('app.tenant_id')` is unset, and the RLS policy blocks all rows — or worse, if RLS is disabled in tests, the query runs without tenant filtering.

## Error Mapping

```go
// error_mapping.go
func parseContractDBError(err error, op string) error {
    if errors.Is(err, pgx.ErrNoRows) {
        return domain.ErrContractNotFound
    }
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        switch pgErr.Code {
        case "23505":
            if strings.Contains(pgErr.ConstraintName, "contract_number") {
                return &domain.BusinessError{
                    Code:    "DUPLICATE_CONTRACT_NUMBER",
                    Message: "contract number already exists in this tenant",
                    Status:  409,
                    Err:     domain.ErrDuplicateNumber,
                }
            }
        case "23514":
            return &domain.BusinessError{
                Code:    "VALIDATION_ERROR",
                Message: "data constraint violated: " + pgErr.ConstraintName,
                Status:  422,
            }
        }
    }
    return fmt.Errorf("%s: %w", op, err)
}
```

## Domain Mapper

```go
// mappers.go
func toDomain(row sqlc.Contract) *domain.Contract {
    return &domain.Contract{
        ID:             row.ID,
        TenantID:       row.TenantID,
        ContractNumber: row.ContractNumber,
        Title:          row.Title,
        Status:         domain.ContractStatus(row.Status),
        ContractType:   domain.ContractType(row.ContractType),
        TotalValue:     row.TotalValue,
        Currency:       row.Currency,
        StartDate:      row.StartDate.Time,
        EndDate:        row.EndDate.Time,
        Version:        int(row.Version),
        CreatedAt:      row.CreatedAt.Time,
        UpdatedAt:      row.UpdatedAt.Time,
        DeletedAt:      nullTimePtr(row.DeletedAt),
    }
}
```

Mapper lives in the repository — the domain struct has no knowledge of SQLC types.
