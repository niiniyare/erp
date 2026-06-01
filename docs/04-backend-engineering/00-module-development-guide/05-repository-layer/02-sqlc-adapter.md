---
title: SQLC Adapter
portal: 4 — Backend Engineering
section: 00-module-development-guide/05-repository-layer
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-repository-interface.md
    title: Repository Interface
  - path: ./03-with-tenant-pattern.md
    title: WithTenant Pattern
  - path: ./04-error-mapping.md
    title: Error Mapping
---

# SQLC Adapter

`repository/<noun>_sqlc.go` is the concrete implementation of the repository interface. It translates between domain types and SQLC-generated types, and wraps every query in `store.WithTenant`.

## Complete Adapter: contract_sqlc.go

```go
// internal/core/contracts/repository/contract_sqlc.go
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/contracts/domain"
)

// contractSQLCRepository is the SQLC-backed implementation of ContractRepository.
type contractSQLCRepository struct {
	store db.Store
}

// NewContractRepository constructs a ContractRepository backed by the given store.
func NewContractRepository(store db.Store) ContractRepository {
	return &contractSQLCRepository{store: store}
}

// ============================================================
// ContractRepository implementation
// ============================================================

func (r *contractSQLCRepository) Create(ctx context.Context, params CreateContractParams) (*domain.Contract, error) {
	var row db.Contract
	err := r.store.WithTenant(ctx, params.TenantID, func(q *db.Queries) error {
		var e error
		row, e = q.CreateContract(ctx, db.CreateContractParams{
			TenantID:       params.TenantID,
			EntityID:       params.EntityID,
			ContractNumber: params.ContractNumber,
			Title:          params.Title,
			Description:    params.Description,
			VendorID:       params.VendorID,
			ContractType:   string(params.ContractType),
			StartDate:      params.StartDate,
			EndDate:        params.EndDate,
			TotalValue:     params.TotalValue,
			Currency:       params.Currency,
			CreatedBy:      params.CreatedBy,
		})
		return e
	})
	if err != nil {
		return nil, mapContractDBError(err, "Create")
	}
	return mapContractRowToDomain(row), nil
}

func (r *contractSQLCRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Contract, error) {
	var row db.Contract
	err := r.store.WithTenant(ctx, tenantID, func(q *db.Queries) error {
		var e error
		row, e = q.GetContractByID(ctx, db.GetContractByIDParams{
			ID:       id,
			TenantID: tenantID,
		})
		return e
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrContractNotFound
		}
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return mapContractRowToDomain(row), nil
}

func (r *contractSQLCRepository) List(ctx context.Context, params ListContractsParams) ([]*domain.Contract, error) {
	var rows []db.Contract
	err := r.store.WithTenant(ctx, params.TenantID, func(q *db.Queries) error {
		var e error
		rows, e = q.ListContracts(ctx, db.ListContractsParams{
			TenantID:     params.TenantID,
			EntityID:     uuidPtrToNullUUID(params.EntityID),
			VendorID:     uuidPtrToNullUUID(params.VendorID),
			Status:       stringPtrToNullString((*string)(params.Status)),
			ContractType: stringPtrToNullString((*string)(params.ContractType)),
			Search:       params.Search,
			SortBy:       params.SortBy,
			SortDir:      params.SortDir,
			PageSize:     int32(params.PageSize),
			PageOffset:   int32(params.PageOffset),
		})
		return e
	})
	if err != nil {
		return nil, fmt.Errorf("List: %w", err)
	}
	result := make([]*domain.Contract, len(rows))
	for i, row := range rows {
		result[i] = mapContractRowToDomain(row)
	}
	return result, nil
}

func (r *contractSQLCRepository) Count(ctx context.Context, params ListContractsParams) (int64, error) {
	var count int64
	err := r.store.WithTenant(ctx, params.TenantID, func(q *db.Queries) error {
		var e error
		count, e = q.CountContracts(ctx, db.CountContractsParams{
			TenantID:     params.TenantID,
			EntityID:     uuidPtrToNullUUID(params.EntityID),
			VendorID:     uuidPtrToNullUUID(params.VendorID),
			Status:       stringPtrToNullString((*string)(params.Status)),
			ContractType: stringPtrToNullString((*string)(params.ContractType)),
			Search:       params.Search,
		})
		return e
	})
	if err != nil {
		return 0, fmt.Errorf("Count: %w", err)
	}
	return count, nil
}

func (r *contractSQLCRepository) Update(ctx context.Context, params UpdateContractParams) (*domain.Contract, error) {
	var row db.Contract
	err := r.store.WithTenant(ctx, params.TenantID, func(q *db.Queries) error {
		var e error
		row, e = q.UpdateContract(ctx, db.UpdateContractParams{
			ID:           params.ID,
			TenantID:     params.TenantID,
			Title:        params.Title,
			Description:  params.Description,
			VendorID:     params.VendorID,
			ContractType: string(params.ContractType),
			StartDate:    params.StartDate,
			EndDate:      params.EndDate,
			TotalValue:   params.TotalValue,
			Currency:     params.Currency,
			Version:      int32(params.Version),
			UpdatedBy:    params.UpdatedBy,
		})
		return e
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if exists, _ := r.exists(ctx, params.ID, params.TenantID); exists {
				return nil, domain.ErrContractConflict
			}
			return nil, domain.ErrContractNotFound
		}
		return nil, mapContractDBError(err, "Update")
	}
	return mapContractRowToDomain(row), nil
}

func (r *contractSQLCRepository) UpdateStatus(ctx context.Context, params UpdateContractStatusParams) (*domain.Contract, error) {
	var row db.Contract
	err := r.store.WithTenant(ctx, params.TenantID, func(q *db.Queries) error {
		var e error
		row, e = q.UpdateContractStatus(ctx, db.UpdateContractStatusParams{
			ID:        params.ID,
			TenantID:  params.TenantID,
			Status:    string(params.Status),
			Version:   int32(params.Version),
			UpdatedBy: params.UpdatedBy,
		})
		return e
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if exists, _ := r.exists(ctx, params.ID, params.TenantID); exists {
				return nil, domain.ErrContractConflict
			}
			return nil, domain.ErrContractNotFound
		}
		return nil, fmt.Errorf("UpdateStatus: %w", err)
	}
	return mapContractRowToDomain(row), nil
}

func (r *contractSQLCRepository) UpdateTotalValue(ctx context.Context, id, tenantID, updatedBy uuid.UUID) (*domain.Contract, error) {
	var row db.Contract
	err := r.store.WithTenant(ctx, tenantID, func(q *db.Queries) error {
		var e error
		row, e = q.UpdateContractTotalValue(ctx, db.UpdateContractTotalValueParams{
			ID:        id,
			TenantID:  tenantID,
			UpdatedBy: updatedBy,
		})
		return e
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrContractNotFound
		}
		return nil, fmt.Errorf("UpdateTotalValue: %w", err)
	}
	return mapContractRowToDomain(row), nil
}

func (r *contractSQLCRepository) Delete(ctx context.Context, id, tenantID, deletedBy uuid.UUID, version int) (*domain.Contract, error) {
	var row db.Contract
	err := r.store.WithTenant(ctx, tenantID, func(q *db.Queries) error {
		var e error
		row, e = q.SoftDeleteContract(ctx, db.SoftDeleteContractParams{
			ID:        id,
			TenantID:  tenantID,
			Version:   int32(version),
			UpdatedBy: deletedBy,
		})
		return e
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrContractNotFound
		}
		return nil, fmt.Errorf("Delete: %w", err)
	}
	return mapContractRowToDomain(row), nil
}

func (r *contractSQLCRepository) ExistsWithNumber(ctx context.Context, number string, tenantID uuid.UUID) (bool, error) {
	var exists bool
	err := r.store.WithTenant(ctx, tenantID, func(q *db.Queries) error {
		var e error
		exists, e = q.ContractExistsWithNumber(ctx, db.ContractExistsWithNumberParams{
			ContractNumber: number,
			TenantID:       tenantID,
		})
		return e
	})
	if err != nil {
		return false, fmt.Errorf("ExistsWithNumber: %w", err)
	}
	return exists, nil
}

// ============================================================
// Internal helpers
// ============================================================

// exists checks whether the contract exists (ignoring soft-delete).
// Used to disambiguate ErrNoRows into not-found vs version conflict.
func (r *contractSQLCRepository) exists(ctx context.Context, id, tenantID uuid.UUID) (bool, error) {
	var count int64
	err := r.store.WithTenant(ctx, tenantID, func(q *db.Queries) error {
		var e error
		count, e = q.CountContractByID(ctx, db.CountContractByIDParams{
			ID:       id,
			TenantID: tenantID,
		})
		return e
	})
	return count > 0, err
}

// ============================================================
// Row → Domain mapping
// ============================================================

func mapContractRowToDomain(row db.Contract) *domain.Contract {
	totalValue, _ := domain.NewContractValue(row.TotalValue)
	return &domain.Contract{
		ID:             row.ID,
		TenantID:       row.TenantID,
		EntityID:       row.EntityID,
		ContractNumber: row.ContractNumber,
		Title:          row.Title,
		Description:    row.Description,
		VendorID:       row.VendorID,
		ContractType:   domain.ContractType(row.ContractType),
		StartDate:      row.StartDate,
		EndDate:        row.EndDate,
		TotalValue:     totalValue,
		Currency:       row.Currency,
		Status:         domain.ContractStatus(row.Status),
		Version:        int(row.Version),
		CreatedBy:      row.CreatedBy,
		UpdatedBy:      row.UpdatedBy,
		DeletedAt:      nullTimeToPtr(row.DeletedAt),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}
```

## Adapter Rules

**Only the adapter imports `db/sqlc` types.** The service never sees SQLC types. This means changing the query parameter structure requires only updating the adapter's mapping functions, not the service.

**Map all errors at the boundary.** The adapter is the last place where `pgx.ErrNoRows` and `*pgconn.PgError` appear. Above the adapter, callers only see domain errors.

**Disambiguate ErrNoRows.** `pgx.ErrNoRows` on an UPDATE/DELETE with RETURNING can mean either "record not found" or "version conflict". The adapter calls `exists()` to disambiguate.

**Map domain types, not raw strings.** SQLC generates `string` for VARCHAR columns. The adapter casts them to domain types: `domain.ContractStatus(row.Status)`, `domain.ContractType(row.ContractType)`. Invalid values from the DB produce a zero-value domain type — add validation if legacy data is untrustworthy.
