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

type approvalWorkflowRepository struct {
	store   db.Store
	tracing tracing.Service
}

// NewApprovalWorkflowRepository returns a new ApprovalWorkflowRepository.
func NewApprovalWorkflowRepository(store db.Store, tracing tracing.Service) domain.ApprovalWorkflowRepository {
	return &approvalWorkflowRepository{store: store, tracing: tracing}
}

func (r *approvalWorkflowRepository) CreateWorkflow(ctx context.Context, rec *domain.WorkflowRecord) error {
	ctx, span := r.tracing.StartSpan(ctx, "ApprovalWorkflowRepository.CreateWorkflow")
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
INSERT INTO finance_workflow_records
  (tenant_id, transaction_id, status, current_tier, initiated_by, due_at)
VALUES
  (current_tenant_id(), $1, $2, $3, $4, $5)
RETURNING id, created_at, updated_at`

		return tx.QueryRow(ctx, q,
			rec.TransactionID,
			string(rec.Status),
			rec.CurrentTier,
			rec.InitiatedBy,
			rec.DueAt,
		).Scan(&rec.ID, &rec.CreatedAt, &rec.UpdatedAt)
	})
}

func (r *approvalWorkflowRepository) GetWorkflowByTransaction(ctx context.Context, transactionID uuid.UUID) (*domain.WorkflowRecord, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ApprovalWorkflowRepository.GetWorkflowByTransaction")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result *domain.WorkflowRecord
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, transaction_id, status, current_tier, initiated_by, due_at, created_at, updated_at
FROM   finance_workflow_records
WHERE  tenant_id      = current_tenant_id()
  AND  transaction_id = $1
LIMIT  1`

		rec := &domain.WorkflowRecord{}
		var initiatedBy *uuid.UUID
		var status string
		err = tx.QueryRow(ctx, q, transactionID).Scan(
			&rec.ID,
			&rec.TransactionID,
			&status,
			&rec.CurrentTier,
			&initiatedBy,
			&rec.DueAt,
			&rec.CreatedAt,
			&rec.UpdatedAt,
		)
		if err == pgx.ErrNoRows {
			return domain.ErrTransactionNotFound
		}
		if err != nil {
			return fmt.Errorf("get workflow: %w", err)
		}
		rec.Status = domain.ApprovalWorkflowStatus(status)
		if initiatedBy != nil {
			rec.InitiatedBy = *initiatedBy
		}
		result = rec
		return nil
	})
	return result, err
}

func (r *approvalWorkflowRepository) UpdateWorkflowStatus(ctx context.Context, workflowID uuid.UUID, status domain.ApprovalWorkflowStatus, currentTier int32) error {
	ctx, span := r.tracing.StartSpan(ctx, "ApprovalWorkflowRepository.UpdateWorkflowStatus")
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
UPDATE finance_workflow_records
SET    status       = $2,
       current_tier = $3,
       updated_at   = NOW()
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1`

		_, err = tx.Exec(ctx, q, workflowID, string(status), currentTier)
		return err
	})
}

func (r *approvalWorkflowRepository) InsertApprovalHistory(ctx context.Context, entry *domain.ApprovalHistoryEntry) error {
	ctx, span := r.tracing.StartSpan(ctx, "ApprovalWorkflowRepository.InsertApprovalHistory")
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
INSERT INTO finance_approval_history
  (tenant_id, workflow_id, transaction_id, tier, action, performed_by, notes)
VALUES
  (current_tenant_id(), $1, $2, $3, $4, $5, $6)
RETURNING id, created_at`

		return tx.QueryRow(ctx, q,
			entry.WorkflowID,
			entry.TransactionID,
			entry.Tier,
			string(entry.Action),
			entry.PerformedBy,
			entry.Notes,
		).Scan(&entry.ID, &entry.CreatedAt)
	})
}

func (r *approvalWorkflowRepository) GetApprovalHistory(ctx context.Context, transactionID uuid.UUID) ([]*domain.ApprovalHistoryEntry, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ApprovalWorkflowRepository.GetApprovalHistory")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var results []*domain.ApprovalHistoryEntry
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, workflow_id, transaction_id, tier, action, performed_by, notes, created_at
FROM   finance_approval_history
WHERE  tenant_id      = current_tenant_id()
  AND  transaction_id = $1
ORDER  BY created_at ASC`

		rows, err := tx.Query(ctx, q, transactionID)
		if err != nil {
			return fmt.Errorf("get approval history: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			e := &domain.ApprovalHistoryEntry{}
			var action string
			var performedBy *uuid.UUID
			if err := rows.Scan(
				&e.ID, &e.WorkflowID, &e.TransactionID,
				&e.Tier, &action, &performedBy, &e.Notes, &e.CreatedAt,
			); err != nil {
				return fmt.Errorf("scan approval history: %w", err)
			}
			e.Action = domain.ApprovalAction(action)
			if performedBy != nil {
				e.PerformedBy = *performedBy
			}
			results = append(results, e)
		}
		return rows.Err()
	})
	return results, err
}

func (r *approvalWorkflowRepository) GetPendingByUser(ctx context.Context, userID uuid.UUID, role string) ([]*domain.WorkflowRecord, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ApprovalWorkflowRepository.GetPendingByUser")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var results []*domain.WorkflowRecord
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		// Return all pending/in-progress workflows for the tenant.
		// The service layer filters by assignee/role since approval assignment
		// is managed outside this table (no FK to users).
		const q = `
SELECT id, transaction_id, status, current_tier, initiated_by, due_at, created_at, updated_at
FROM   finance_workflow_records
WHERE  tenant_id = current_tenant_id()
  AND  status    IN ('pending', 'in_progress')
ORDER  BY created_at ASC`

		rows, err := tx.Query(ctx, q)
		if err != nil {
			return fmt.Errorf("get pending workflows: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			rec := &domain.WorkflowRecord{}
			var status string
			var initiatedBy *uuid.UUID
			if err := rows.Scan(
				&rec.ID, &rec.TransactionID, &status, &rec.CurrentTier,
				&initiatedBy, &rec.DueAt, &rec.CreatedAt, &rec.UpdatedAt,
			); err != nil {
				return fmt.Errorf("scan pending workflow: %w", err)
			}
			rec.Status = domain.ApprovalWorkflowStatus(status)
			if initiatedBy != nil {
				rec.InitiatedBy = *initiatedBy
			}
			results = append(results, rec)
		}
		return rows.Err()
	})
	return results, err
}
