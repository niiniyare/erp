package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

type reconciliationRepository struct {
	store   db.Store
	tracing tracing.Service
	logger  logger.Logger
	metrics metrics.MetricsProvider
}

// NewReconciliationRepository returns a new domain.ReconciliationRepository.
func NewReconciliationRepository(store db.Store, tracer tracing.Service, log logger.Logger, met metrics.MetricsProvider) domain.ReconciliationRepository {
	initFinanceRepoMetrics(met)
	return &reconciliationRepository{store: store, tracing: tracer, logger: log, metrics: met}
}

// ── Bank Statement ────────────────────────────────────────────────────────────

func (r *reconciliationRepository) CreateStatement(ctx context.Context, s *domain.BankStatement) (err error) {
	ctx, span := r.tracing.StartSpan(ctx, "ReconciliationRepository.CreateStatement")
	defer span.End()
	start := time.Now()
	defer func() { observeOp(ctx, "CreateStatement", "reconciliation", start, err, r.logger, r.metrics) }()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenantFromCtx(ctx, func(ctx context.Context, st db.Store) error {
		tx, err := txFrom(st)
		if err != nil {
			return err
		}
		const q = `
INSERT INTO finance_bank_statements
  (tenant_id, account_id, statement_reference, statement_date, start_date,
   currency_code, opening_balance, closing_balance, status,
   matched_count, unmatched_count, difference_amount, created_by)
VALUES
  (current_tenant_id(), $1, $2, $3, $4,
   $5, $6, $7, $8,
   0, 0, $9, $10)
RETURNING id, created_at, updated_at`

		return tx.QueryRow(ctx, q,
			s.AccountID,
			s.StatementReference,
			s.StatementDate,
			s.StartDate,
			s.CurrencyCode,
			s.OpeningBalance.String(),
			s.ClosingBalance.String(),
			string(s.Status),
			s.DifferenceAmount.String(),
			nullUUID(s.CreatedBy),
		).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
	})
}

func (r *reconciliationRepository) GetStatementByID(ctx context.Context, id uuid.UUID) (*domain.BankStatement, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ReconciliationRepository.GetStatementByID")
	defer span.End()
	start := time.Now()
	var err error
	defer func() { observeOp(ctx, "GetStatementByID", "reconciliation", start, err, r.logger, r.metrics) }()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result *domain.BankStatement
	err = r.store.WithTenantFromCtx(ctx, func(ctx context.Context, st db.Store) error {
		tx, err := txFrom(st)
		if err != nil {
			return err
		}
		const q = `
SELECT id, tenant_id, account_id, statement_reference, statement_date, start_date,
       currency_code, opening_balance, closing_balance, status,
       matched_count, unmatched_count, difference_amount,
       created_at, updated_at, created_by, updated_by
FROM   finance_bank_statements
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1`

		s, scanErr := scanStatement(tx.QueryRow(ctx, q, id))
		if scanErr == pgx.ErrNoRows {
			return domain.ErrStatementNotFound
		}
		if scanErr != nil {
			return fmt.Errorf("get bank statement: %w", scanErr)
		}
		result = s
		return nil
	})
	return result, err
}

func (r *reconciliationRepository) ListStatements(ctx context.Context, tenantID uuid.UUID, accountID *uuid.UUID) ([]*domain.BankStatement, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ReconciliationRepository.ListStatements")
	defer span.End()
	start := time.Now()
	var err error
	defer func() { observeOp(ctx, "ListStatements", "reconciliation", start, err, r.logger, r.metrics) }()

	// Validate that the explicit tenantID matches the context tenant to prevent
	// cross-tenant queries if callers accidentally pass the wrong tenant.
	ctxTenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}
	if ctxTenantID != tenantID {
		return nil, fmt.Errorf("tenant ID mismatch: context tenant %s does not match parameter %s", ctxTenantID, tenantID)
	}

	var results []*domain.BankStatement
	err = r.store.WithTenantFromCtx(ctx, func(ctx context.Context, st db.Store) error {
		tx, err := txFrom(st)
		if err != nil {
			return err
		}

		var sb strings.Builder
		sb.WriteString(`
SELECT id, tenant_id, account_id, statement_reference, statement_date, start_date,
       currency_code, opening_balance, closing_balance, status,
       matched_count, unmatched_count, difference_amount,
       created_at, updated_at, created_by, updated_by
FROM   finance_bank_statements
WHERE  tenant_id = current_tenant_id()`)

		args := []interface{}{}
		if accountID != nil {
			args = append(args, *accountID)
			sb.WriteString(fmt.Sprintf(" AND account_id = $%d", len(args)))
		}
		sb.WriteString(" ORDER BY statement_date DESC")

		rows, err := tx.Query(ctx, sb.String(), args...)
		if err != nil {
			return fmt.Errorf("list bank statements: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			s, scanErr := scanStatement(rows)
			if scanErr != nil {
				return fmt.Errorf("scan bank statement: %w", scanErr)
			}
			results = append(results, s)
		}
		return rows.Err()
	})
	return results, err
}

func (r *reconciliationRepository) UpdateStatement(ctx context.Context, s *domain.BankStatement) (err error) {
	ctx, span := r.tracing.StartSpan(ctx, "ReconciliationRepository.UpdateStatement")
	defer span.End()
	start := time.Now()
	defer func() { observeOp(ctx, "UpdateStatement", "reconciliation", start, err, r.logger, r.metrics) }()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenantFromCtx(ctx, func(ctx context.Context, st db.Store) error {
		tx, err := txFrom(st)
		if err != nil {
			return err
		}
		const q = `
UPDATE finance_bank_statements
SET    status            = $2,
       matched_count     = $3,
       unmatched_count   = $4,
       difference_amount = $5,
       updated_at        = NOW(),
       updated_by        = $6
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1
RETURNING updated_at`

		return tx.QueryRow(ctx, q,
			s.ID,
			string(s.Status),
			s.MatchedCount,
			s.UnmatchedCount,
			s.DifferenceAmount.String(),
			nullUUID2(s.UpdatedBy),
		).Scan(&s.UpdatedAt)
	})
}

// ── Statement Lines ────────────────────────────────────────────────────────────

func (r *reconciliationRepository) CreateLines(ctx context.Context, lines []*domain.BankStatementLine) (err error) {
	ctx, span := r.tracing.StartSpan(ctx, "ReconciliationRepository.CreateLines")
	defer span.End()
	start := time.Now()
	defer func() { observeOp(ctx, "CreateLines", "reconciliation", start, err, r.logger, r.metrics) }()

	if len(lines) == 0 {
		return nil
	}

	if _, ok := shared.GetTenantID(ctx); !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenantFromCtx(ctx, func(ctx context.Context, st db.Store) error {
		tx, err := txFrom(st)
		if err != nil {
			return err
		}
		const q = `
INSERT INTO finance_bank_statement_lines
  (statement_id, tenant_id, transaction_date, value_date,
   description, reference, debit_amount, credit_amount, balance)
VALUES
  ($1, current_tenant_id(), $2, $3, $4, $5, $6, $7, $8)
RETURNING id, created_at`

		for _, li := range lines {
			if err := tx.QueryRow(ctx, q,
				li.StatementID,
				li.TransactionDate,
				li.ValueDate,
				li.Description,
				li.Reference,
				li.DebitAmount.String(),
				li.CreditAmount.String(),
				li.Balance.String(),
			).Scan(&li.ID, &li.CreatedAt); err != nil {
				return fmt.Errorf("insert statement line: %w", err)
			}
		}
		return nil
	})
}

func (r *reconciliationRepository) GetLine(ctx context.Context, lineID uuid.UUID) (*domain.BankStatementLine, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ReconciliationRepository.GetLine")
	defer span.End()
	start := time.Now()
	var err error
	defer func() { observeOp(ctx, "GetLine", "reconciliation", start, err, r.logger, r.metrics) }()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result *domain.BankStatementLine
	err = r.store.WithTenantFromCtx(ctx, func(ctx context.Context, st db.Store) error {
		tx, err := txFrom(st)
		if err != nil {
			return err
		}
		const q = `
SELECT id, statement_id, tenant_id, transaction_date, value_date,
       description, reference, debit_amount, credit_amount, balance,
       is_reconciled, reconciled_at, reconciled_by, matched_entry_id, created_at
FROM   finance_bank_statement_lines
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1`

		li, scanErr := scanLine(tx.QueryRow(ctx, q, lineID))
		if scanErr == pgx.ErrNoRows {
			return domain.ErrLineNotFound
		}
		if scanErr != nil {
			return fmt.Errorf("get statement line: %w", scanErr)
		}
		result = li
		return nil
	})
	return result, err
}

func (r *reconciliationRepository) ListLines(ctx context.Context, statementID uuid.UUID, unmatchedOnly bool) ([]*domain.BankStatementLine, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ReconciliationRepository.ListLines")
	defer span.End()
	start := time.Now()
	var err error
	defer func() { observeOp(ctx, "ListLines", "reconciliation", start, err, r.logger, r.metrics) }()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var results []*domain.BankStatementLine
	err = r.store.WithTenantFromCtx(ctx, func(ctx context.Context, st db.Store) error {
		tx, err := txFrom(st)
		if err != nil {
			return err
		}

		var sb strings.Builder
		sb.WriteString(`
SELECT id, statement_id, tenant_id, transaction_date, value_date,
       description, reference, debit_amount, credit_amount, balance,
       is_reconciled, reconciled_at, reconciled_by, matched_entry_id, created_at
FROM   finance_bank_statement_lines
WHERE  tenant_id    = current_tenant_id()
  AND  statement_id = $1`)
		if unmatchedOnly {
			sb.WriteString(` AND is_reconciled = FALSE`)
		}
		sb.WriteString(` ORDER BY transaction_date, id`)

		rows, err := tx.Query(ctx, sb.String(), statementID)
		if err != nil {
			return fmt.Errorf("list statement lines: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			li, scanErr := scanLine(rows)
			if scanErr != nil {
				return fmt.Errorf("scan statement line: %w", scanErr)
			}
			results = append(results, li)
		}
		return rows.Err()
	})
	return results, err
}

func (r *reconciliationRepository) MatchLine(ctx context.Context, lineID, entryID uuid.UUID, byUserID uuid.UUID) (err error) {
	ctx, span := r.tracing.StartSpan(ctx, "ReconciliationRepository.MatchLine")
	defer span.End()
	start := time.Now()
	defer func() { observeOp(ctx, "MatchLine", "reconciliation", start, err, r.logger, r.metrics) }()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	now := time.Now()
	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, st db.Store) error {
		tx, err := txFrom(st)
		if err != nil {
			return err
		}
		const q = `
UPDATE finance_bank_statement_lines
SET    is_reconciled    = TRUE,
       reconciled_at    = $3,
       reconciled_by    = $4,
       matched_entry_id = $2
WHERE  tenant_id     = current_tenant_id()
  AND  id            = $1
  AND  is_reconciled = FALSE`

		tag, err := tx.Exec(ctx, q, lineID, entryID, now, byUserID)
		if err != nil {
			return fmt.Errorf("match statement line: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return domain.ErrLineAlreadyMatched
		}
		return nil
	})
}

func (r *reconciliationRepository) UnmatchLine(ctx context.Context, lineID uuid.UUID) (err error) {
	ctx, span := r.tracing.StartSpan(ctx, "ReconciliationRepository.UnmatchLine")
	defer span.End()
	start := time.Now()
	defer func() { observeOp(ctx, "UnmatchLine", "reconciliation", start, err, r.logger, r.metrics) }()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenantFromCtx(ctx, func(ctx context.Context, st db.Store) error {
		tx, err := txFrom(st)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
UPDATE finance_bank_statement_lines
SET    is_reconciled    = FALSE,
       reconciled_at    = NULL,
       reconciled_by    = NULL,
       matched_entry_id = NULL
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1`, lineID)
		return err
	})
}

// CompleteReconciliation marks the statement COMPLETED and stamps all matched
// finance_transaction_entries rows as reconciled.
func (r *reconciliationRepository) CompleteReconciliation(ctx context.Context, statementID uuid.UUID, byUserID uuid.UUID) (err error) {
	ctx, span := r.tracing.StartSpan(ctx, "ReconciliationRepository.CompleteReconciliation")
	defer span.End()
	start := time.Now()
	defer func() { observeOp(ctx, "CompleteReconciliation", "reconciliation", start, err, r.logger, r.metrics) }()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	now := time.Now()
	reconcilRef := "RECON-" + statementID.String()[:8]

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, st db.Store) error {
		tx, err := txFrom(st)
		if err != nil {
			return err
		}

		// 1. Mark statement COMPLETED
		const updateStmt = `
UPDATE finance_bank_statements
SET    status     = 'COMPLETED',
       updated_at = $2,
       updated_by = $3
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1
  AND  status   != 'COMPLETED'`

		tag, err := tx.Exec(ctx, updateStmt, statementID, now, byUserID)
		if err != nil {
			return fmt.Errorf("complete statement: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return domain.ErrStatementAlreadyClosed
		}

		// 2. Mark matched journal entries as reconciled using the entry IDs collected
		//    from the matched lines of this statement.
		const updateEntries = `
UPDATE finance_transaction_entries te
SET    reconciled               = TRUE,
       reconciled_date          = $1,
       reconciliation_reference = $2
FROM   finance_bank_statement_lines bsl
WHERE  bsl.statement_id     = $3
  AND  bsl.matched_entry_id = te.id
  AND  bsl.is_reconciled    = TRUE`

		_, err = tx.Exec(ctx, updateEntries, now, reconcilRef, statementID)
		return err
	})
}

// ── Scan helpers ──────────────────────────────────────────────────────────────

func scanStatement(row scannable) (*domain.BankStatement, error) {
	s := &domain.BankStatement{}
	var status string
	var openStr, closeStr, diffStr string
	var createdBy *uuid.UUID
	err := row.Scan(
		&s.ID, &s.TenantID, &s.AccountID,
		&s.StatementReference, &s.StatementDate, &s.StartDate,
		&s.CurrencyCode, &openStr, &closeStr, &status,
		&s.MatchedCount, &s.UnmatchedCount, &diffStr,
		&s.CreatedAt, &s.UpdatedAt, &createdBy, &s.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	if createdBy != nil {
		s.CreatedBy = *createdBy
	}
	s.Status = domain.ReconciliationStatus(status)
	if s.OpeningBalance, err = decimal.NewFromString(openStr); err != nil {
		return nil, fmt.Errorf("parse opening_balance: %w", err)
	}
	if s.ClosingBalance, err = decimal.NewFromString(closeStr); err != nil {
		return nil, fmt.Errorf("parse closing_balance: %w", err)
	}
	if s.DifferenceAmount, err = decimal.NewFromString(diffStr); err != nil {
		return nil, fmt.Errorf("parse difference_amount: %w", err)
	}
	return s, nil
}

func scanLine(row scannable) (*domain.BankStatementLine, error) {
	li := &domain.BankStatementLine{}
	var debitStr, creditStr, balStr string
	err := row.Scan(
		&li.ID, &li.StatementID, &li.TenantID,
		&li.TransactionDate, &li.ValueDate,
		&li.Description, &li.Reference,
		&debitStr, &creditStr, &balStr,
		&li.IsReconciled, &li.ReconciledAt, &li.ReconciledBy,
		&li.MatchedEntryID, &li.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if li.DebitAmount, err = decimal.NewFromString(debitStr); err != nil {
		return nil, fmt.Errorf("parse debit_amount: %w", err)
	}
	if li.CreditAmount, err = decimal.NewFromString(creditStr); err != nil {
		return nil, fmt.Errorf("parse credit_amount: %w", err)
	}
	if li.Balance, err = decimal.NewFromString(balStr); err != nil {
		return nil, fmt.Errorf("parse balance: %w", err)
	}
	return li, nil
}
