package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/tracing"
)

type reversalHistoryRepository struct {
	store   db.Store
	tracing tracing.Service
}

// NewReversalHistoryRepository returns a new ReversalHistoryRepository.
func NewReversalHistoryRepository(store db.Store, tracing tracing.Service) domain.ReversalHistoryRepository {
	return &reversalHistoryRepository{store: store, tracing: tracing}
}

// Insert persists one reversal event inside the caller's tenant context.
func (r *reversalHistoryRepository) Insert(ctx context.Context, rec *domain.ReversalHistoryRecord) error {
	ctx, span := r.tracing.StartSpan(ctx, "ReversalHistoryRepository.Insert")
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
INSERT INTO finance_reversal_history
  (tenant_id, original_transaction_id, reversal_transaction_id, reason, reversed_by)
VALUES
  (current_tenant_id(), $1, $2, $3, $4)`

		_, err = tx.Exec(ctx, q,
			rec.OriginalTransactionID,
			rec.ReversalTransactionID,
			rec.Reason,
			rec.ReversedBy,
		)
		return err
	})
}

// IsReversal returns true when transactionID is itself a reversal of another transaction.
func (r *reversalHistoryRepository) IsReversal(ctx context.Context, transactionID uuid.UUID) (bool, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ReversalHistoryRepository.IsReversal")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return false, fmt.Errorf("tenant ID not found in context")
	}

	var isReversal bool
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT EXISTS (
  SELECT 1
  FROM   finance_reversal_history
  WHERE  tenant_id               = current_tenant_id()
    AND  reversal_transaction_id = $1
)`
		return tx.QueryRow(ctx, q, transactionID).Scan(&isReversal)
	})
	return isReversal, err
}

// GetByOriginal returns all reversal records for the given original transaction.
func (r *reversalHistoryRepository) GetByOriginal(ctx context.Context, originalTransactionID uuid.UUID) ([]*domain.ReversalHistoryRecord, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ReversalHistoryRepository.GetByOriginal")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var results []*domain.ReversalHistoryRecord
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		tx, err := txFrom(s)
		if err != nil {
			return err
		}
		const q = `
SELECT id, original_transaction_id, reversal_transaction_id, reason, reversed_by, created_at
FROM   finance_reversal_history
WHERE  tenant_id               = current_tenant_id()
  AND  original_transaction_id = $1
ORDER  BY created_at ASC`

		rows, err := tx.Query(ctx, q, originalTransactionID)
		if err != nil {
			return fmt.Errorf("get reversal history: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			rec := &domain.ReversalHistoryRecord{}
			var reversedBy *uuid.UUID
			if err := rows.Scan(
				&rec.ID,
				&rec.OriginalTransactionID,
				&rec.ReversalTransactionID,
				&rec.Reason,
				&reversedBy,
				&rec.CreatedAt,
			); err != nil {
				return fmt.Errorf("scan reversal history: %w", err)
			}
			if reversedBy != nil {
				rec.ReversedBy = *reversedBy
			}
			results = append(results, rec)
		}
		return rows.Err()
	})
	return results, err
}
