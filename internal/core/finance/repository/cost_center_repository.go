package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/tracing"
)

type costCenterRepository struct {
	store   db.Store
	tracing tracing.Service
}

// NewCostCenterRepository returns a new domain.CostCenterRepository.
func NewCostCenterRepository(store db.Store, tracing tracing.Service) domain.CostCenterRepository {
	return &costCenterRepository{store: store, tracing: tracing}
}

func (r *costCenterRepository) Create(ctx context.Context, cc *domain.CostCenter) error {
	ctx, span := r.tracing.StartSpan(ctx, "CostCenterRepository.Create")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
INSERT INTO finance_cost_centers
  (tenant_id, code, name, description, parent_id, is_group, is_distributed, allocation_method, is_active, created_by)
VALUES
  (current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, created_at, updated_at`

		var desc *string
		if *cc.Description != "" {
			desc = cc.Description
		}
		var method *string
		if cc.AllocationMethod != nil {
			m := string(*cc.AllocationMethod)
			method = &m
		}

		return tx.QueryRow(ctx, q,
			cc.Code,
			cc.Name,
			desc,
			nullUUID2(cc.ParentID),
			cc.IsGroup,
			cc.IsDistributed,
			method,
			cc.IsActive,
			nullUUID(cc.CreatedBy),
		).Scan(&cc.ID, &cc.CreatedAt, &cc.UpdatedAt)
	})
}

func (r *costCenterRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CostCenter, error) {
	ctx, span := r.tracing.StartSpan(ctx, "CostCenterRepository.GetByID")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result *domain.CostCenter
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tenant_id, code, name, description, parent_id,
       is_group, is_distributed, allocation_method, is_active,
       created_at, updated_at, created_by, updated_by
FROM   finance_cost_centers
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1`

		cc := &domain.CostCenter{}
		var desc *string
		var method *string
		err = tx.QueryRow(ctx, q, id).Scan(
			&cc.ID, &cc.TenantID,
			&cc.Code, &cc.Name, &desc, &cc.ParentID,
			&cc.IsGroup, &cc.IsDistributed, &method, &cc.IsActive,
			&cc.CreatedAt, &cc.UpdatedAt,
			&cc.CreatedBy, &cc.UpdatedBy,
		)
		if err == pgx.ErrNoRows {
			return domain.ErrCostCenterNotFound
		}
		if err != nil {
			return fmt.Errorf("get cost centre by id: %w", err)
		}
		if desc != nil {
			cc.Description = *desc
		}
		if method != nil {
			m := domain.AllocationMethod(*method)
			cc.AllocationMethod = &m
		}
		result = cc
		return nil
	})
	return result, err
}

func (r *costCenterRepository) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*domain.CostCenter, error) {
	ctx, span := r.tracing.StartSpan(ctx, "CostCenterRepository.GetByCode")
	defer span.End()

	var result *domain.CostCenter
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tenant_id, code, name, description, parent_id,
       is_group, is_distributed, allocation_method, is_active,
       created_at, updated_at, created_by, updated_by
FROM   finance_cost_centers
WHERE  tenant_id = current_tenant_id()
  AND  code      = $1`

		cc := &domain.CostCenter{}
		var desc *string
		var method *string
		err = tx.QueryRow(ctx, q, code).Scan(
			&cc.ID, &cc.TenantID,
			&cc.Code, &cc.Name, &desc, &cc.ParentID,
			&cc.IsGroup, &cc.IsDistributed, &method, &cc.IsActive,
			&cc.CreatedAt, &cc.UpdatedAt,
			&cc.CreatedBy, &cc.UpdatedBy,
		)
		if err == pgx.ErrNoRows {
			return domain.ErrCostCenterNotFound
		}
		if err != nil {
			return fmt.Errorf("get cost centre by code: %w", err)
		}
		if desc != nil {
			cc.Description = *desc
		}
		if method != nil {
			m := domain.AllocationMethod(*method)
			cc.AllocationMethod = &m
		}
		result = cc
		return nil
	})
	return result, err
}

func (r *costCenterRepository) Update(ctx context.Context, cc *domain.CostCenter) error {
	ctx, span := r.tracing.StartSpan(ctx, "CostCenterRepository.Update")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
UPDATE finance_cost_centers
SET    code              = $2,
       name              = $3,
       description       = $4,
       parent_id         = $5,
       is_group          = $6,
       is_distributed    = $7,
       allocation_method = $8,
       is_active         = $9,
       updated_at        = NOW(),
       updated_by        = $10
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1
RETURNING updated_at`

		var desc *string
		if cc.Description != "" {
			desc = &cc.Description
		}
		var method *string
		if cc.AllocationMethod != nil {
			m := string(*cc.AllocationMethod)
			method = &m
		}

		return tx.QueryRow(ctx, q,
			cc.ID,
			cc.Code,
			cc.Name,
			desc,
			nullUUID2(cc.ParentID),
			cc.IsGroup,
			cc.IsDistributed,
			method,
			cc.IsActive,
			nullUUID2(cc.UpdatedBy),
		).Scan(&cc.UpdatedAt)
	})
}

func (r *costCenterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "CostCenterRepository.Delete")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
DELETE FROM finance_cost_centers
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1`

		tag, err := tx.Exec(ctx, q, id)
		if err != nil {
			return fmt.Errorf("delete cost centre: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return domain.ErrCostCenterNotFound
		}
		return nil
	})
}

func (r *costCenterRepository) List(ctx context.Context, tenantID uuid.UUID, activeOnly bool) ([]*domain.CostCenter, error) {
	ctx, span := r.tracing.StartSpan(ctx, "CostCenterRepository.List")
	defer span.End()

	var results []*domain.CostCenter
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tenant_id, code, name, description, parent_id,
       is_group, is_distributed, allocation_method, is_active,
       created_at, updated_at, created_by, updated_by
FROM   finance_cost_centers
WHERE  tenant_id = current_tenant_id()
  AND  ($1 = FALSE OR is_active = TRUE)
ORDER  BY code ASC`

		rows, err := tx.Query(ctx, q, activeOnly)
		if err != nil {
			return fmt.Errorf("list cost centres: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			cc := &domain.CostCenter{}
			var desc *string
			var method *string
			if err := rows.Scan(
				&cc.ID, &cc.TenantID,
				&cc.Code, &cc.Name, &desc, &cc.ParentID,
				&cc.IsGroup, &cc.IsDistributed, &method, &cc.IsActive,
				&cc.CreatedAt, &cc.UpdatedAt,
				&cc.CreatedBy, &cc.UpdatedBy,
			); err != nil {
				return fmt.Errorf("scan cost centre: %w", err)
			}
			if desc != nil {
				cc.Description = *desc
			}
			if method != nil {
				m := domain.AllocationMethod(*method)
				cc.AllocationMethod = &m
			}
			results = append(results, cc)
		}
		return rows.Err()
	})
	return results, err
}

func (r *costCenterRepository) ValidateCode(ctx context.Context, tenantID uuid.UUID, code string, excludeID *uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "CostCenterRepository.ValidateCode")
	defer span.End()

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT COUNT(*) FROM finance_cost_centers
WHERE  tenant_id = current_tenant_id()
  AND  code      = $1
  AND  ($2::UUID IS NULL OR id <> $2)`

		var count int64
		if err := tx.QueryRow(ctx, q, code, excludeID).Scan(&count); err != nil {
			return fmt.Errorf("validate cost centre code: %w", err)
		}
		if count > 0 {
			return domain.ErrCostCenterCodeExists
		}
		return nil
	})
}

// nullUUID2 converts a *uuid.UUID to nil if it is nil or zero.
func nullUUID2(id *uuid.UUID) interface{} {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	return *id
}
