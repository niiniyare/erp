package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/tracing"
)

type budgetRepository struct {
	store   db.Store
	tracing tracing.Service
}

// NewBudgetRepository returns a new domain.BudgetRepository.
func NewBudgetRepository(store db.Store, tracing tracing.Service) domain.BudgetRepository {
	return &budgetRepository{store: store, tracing: tracing}
}

// ── Budget header ─────────────────────────────────────────────────────────────

func (r *budgetRepository) CreateBudget(ctx context.Context, b *domain.Budget) error {
	ctx, span := r.tracing.StartSpan(ctx, "BudgetRepository.CreateBudget")
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
INSERT INTO finance_budgets
  (tenant_id, fiscal_year_id, name, description, budget_type, status, currency_code,
   cost_center_id, version, original_budget_id, created_by)
VALUES
  (current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, created_at, updated_at`

		var desc *string
		if b.Description != nil && *b.Description != "" {
			desc = b.Description
		}

		return tx.QueryRow(ctx, q,
			b.FiscalYearID,
			b.Name,
			desc,
			string(b.BudgetType),
			string(b.Status),
			b.CurrencyCode,
			nullUUID2(b.CostCenterID),
			b.Version,
			nullUUID2(b.OriginalBudgetID),
			nullUUID(b.CreatedBy),
		).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
	})
}

func (r *budgetRepository) GetBudgetByID(ctx context.Context, id uuid.UUID) (*domain.Budget, error) {
	ctx, span := r.tracing.StartSpan(ctx, "BudgetRepository.GetBudgetByID")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result *domain.Budget
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tenant_id, fiscal_year_id, name, description, budget_type, status, currency_code,
       cost_center_id, submitted_at, submitted_by, approved_at, approved_by,
       rejected_at, rejected_by, reject_note, version, original_budget_id,
       created_at, updated_at, created_by, updated_by
FROM   finance_budgets
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1`

		b := &domain.Budget{}
		var budgetType, status string
		var createdBy *uuid.UUID
		err = tx.QueryRow(ctx, q, id).Scan(
			&b.ID, &b.TenantID, &b.FiscalYearID,
			&b.Name, &b.Description, &budgetType, &status, &b.CurrencyCode,
			&b.CostCenterID,
			&b.SubmittedAt, &b.SubmittedBy,
			&b.ApprovedAt, &b.ApprovedBy,
			&b.RejectedAt, &b.RejectedBy, &b.RejectNote,
			&b.Version, &b.OriginalBudgetID,
			&b.CreatedAt, &b.UpdatedAt, &createdBy, &b.UpdatedBy,
		)
		if err == pgx.ErrNoRows {
			return domain.ErrBudgetNotFound
		}
		if err != nil {
			return fmt.Errorf("get budget by id: %w", err)
		}
		if createdBy != nil {
			b.CreatedBy = *createdBy
		}
		b.BudgetType = domain.BudgetType(budgetType)
		b.Status = domain.BudgetStatus(status)
		result = b
		return nil
	})
	return result, err
}

func (r *budgetRepository) ListBudgets(ctx context.Context, tenantID uuid.UUID, fiscalYearID *uuid.UUID) ([]*domain.Budget, error) {
	ctx, span := r.tracing.StartSpan(ctx, "BudgetRepository.ListBudgets")
	defer span.End()

	var results []*domain.Budget
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}

		var sb strings.Builder
		sb.WriteString(`
SELECT id, tenant_id, fiscal_year_id, name, description, budget_type, status, currency_code,
       cost_center_id, submitted_at, submitted_by, approved_at, approved_by,
       rejected_at, rejected_by, reject_note, version, original_budget_id,
       created_at, updated_at, created_by, updated_by
FROM   finance_budgets
WHERE  tenant_id = current_tenant_id()`)

		args := []interface{}{}
		if fiscalYearID != nil {
			args = append(args, *fiscalYearID)
			sb.WriteString(fmt.Sprintf(" AND fiscal_year_id = $%d", len(args)))
		}
		sb.WriteString(" ORDER BY created_at DESC")

		rows, err := tx.Query(ctx, sb.String(), args...)
		if err != nil {
			return fmt.Errorf("list budgets: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			b := &domain.Budget{}
			var budgetType, status string
			var createdBy *uuid.UUID
			if err := rows.Scan(
				&b.ID, &b.TenantID, &b.FiscalYearID,
				&b.Name, &b.Description, &budgetType, &status, &b.CurrencyCode,
				&b.CostCenterID,
				&b.SubmittedAt, &b.SubmittedBy,
				&b.ApprovedAt, &b.ApprovedBy,
				&b.RejectedAt, &b.RejectedBy, &b.RejectNote,
				&b.Version, &b.OriginalBudgetID,
				&b.CreatedAt, &b.UpdatedAt, &createdBy, &b.UpdatedBy,
			); err != nil {
				return fmt.Errorf("scan budget: %w", err)
			}
			if createdBy != nil {
				b.CreatedBy = *createdBy
			}
			b.BudgetType = domain.BudgetType(budgetType)
			b.Status = domain.BudgetStatus(status)
			results = append(results, b)
		}
		return rows.Err()
	})
	return results, err
}

func (r *budgetRepository) UpdateBudget(ctx context.Context, b *domain.Budget) error {
	ctx, span := r.tracing.StartSpan(ctx, "BudgetRepository.UpdateBudget")
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
UPDATE finance_budgets
SET    name             = $2,
       description      = $3,
       status           = $4,
       submitted_at     = $5,
       submitted_by     = $6,
       approved_at      = $7,
       approved_by      = $8,
       rejected_at      = $9,
       rejected_by      = $10,
       reject_note      = $11,
       updated_at       = NOW(),
       updated_by       = $12
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1
RETURNING updated_at`

		return tx.QueryRow(ctx, q,
			b.ID,
			b.Name,
			b.Description,
			string(b.Status),
			b.SubmittedAt,
			nullUUID2(b.SubmittedBy),
			b.ApprovedAt,
			nullUUID2(b.ApprovedBy),
			b.RejectedAt,
			nullUUID2(b.RejectedBy),
			b.RejectNote,
			nullUUID2(b.UpdatedBy),
		).Scan(&b.UpdatedAt)
	})
}

// ── Budget line items ─────────────────────────────────────────────────────────

func (r *budgetRepository) CreateLineItems(ctx context.Context, lines []*domain.BudgetLineItem) error {
	ctx, span := r.tracing.StartSpan(ctx, "BudgetRepository.CreateLineItems")
	defer span.End()

	if len(lines) == 0 {
		return nil
	}

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
INSERT INTO finance_budget_line_items
  (budget_id, tenant_id, account_id, cost_center_id, period_id, budgeted_amount, notes, created_by)
VALUES
  ($1, current_tenant_id(), $2, $3, $4, $5, $6, $7)
RETURNING id, created_at, updated_at`

		for _, li := range lines {
			var notes *string
			if li.Notes != nil && *li.Notes != "" {
				notes = li.Notes
			}
			if err := tx.QueryRow(ctx, q,
				li.BudgetID,
				li.AccountID,
				nullUUID2(li.CostCenterID),
				nullUUID2(li.PeriodID),
				li.BudgetedAmount,
				notes,
				nullUUID(li.CreatedBy),
			).Scan(&li.ID, &li.CreatedAt, &li.UpdatedAt); err != nil {
				return fmt.Errorf("insert budget line item: %w", err)
			}
		}
		return nil
	})
}

func (r *budgetRepository) GetLineItems(ctx context.Context, budgetID uuid.UUID) ([]*domain.BudgetLineItem, error) {
	ctx, span := r.tracing.StartSpan(ctx, "BudgetRepository.GetLineItems")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var results []*domain.BudgetLineItem
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, budget_id, tenant_id, account_id, cost_center_id, period_id,
       budgeted_amount, actual_amount, notes,
       created_at, updated_at, created_by, updated_by
FROM   finance_budget_line_items
WHERE  tenant_id = current_tenant_id()
  AND  budget_id = $1
ORDER  BY account_id, period_id`

		rows, err := tx.Query(ctx, q, budgetID)
		if err != nil {
			return fmt.Errorf("get budget line items: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			li := &domain.BudgetLineItem{}
			var budgetedStr, actualStr string
			var createdBy *uuid.UUID
			if err := rows.Scan(
				&li.ID, &li.BudgetID, &li.TenantID,
				&li.AccountID, &li.CostCenterID, &li.PeriodID,
				&budgetedStr, &actualStr, &li.Notes,
				&li.CreatedAt, &li.UpdatedAt, &createdBy, &li.UpdatedBy,
			); err != nil {
				return fmt.Errorf("scan budget line item: %w", err)
			}
			if createdBy != nil {
				li.CreatedBy = *createdBy
			}
			if li.BudgetedAmount, err = decimal.NewFromString(budgetedStr); err != nil {
				return fmt.Errorf("parse budgeted_amount: %w", err)
			}
			if li.ActualAmount, err = decimal.NewFromString(actualStr); err != nil {
				return fmt.Errorf("parse actual_amount: %w", err)
			}
			li.VarianceAmount = li.BudgetedAmount.Sub(li.ActualAmount)
			if !li.BudgetedAmount.IsZero() {
				li.VariancePct = li.VarianceAmount.Div(li.BudgetedAmount).Mul(decimal.NewFromInt(100))
			}
			results = append(results, li)
		}
		return rows.Err()
	})
	return results, err
}

func (r *budgetRepository) DeleteLineItems(ctx context.Context, budgetID uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "BudgetRepository.DeleteLineItems")
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
		_, err = tx.Exec(ctx,
			`DELETE FROM finance_budget_line_items WHERE tenant_id = current_tenant_id() AND budget_id = $1`,
			budgetID,
		)
		return err
	})
}
