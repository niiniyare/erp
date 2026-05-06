// Package service provides business logic and service layer implementation
// for the application. It handles core operations, data processing, and
// coordinates between different components of the system.
package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

type TransactionEntryService interface {
	CreateEntry(ctx context.Context, entry *domain.TransactionEntry) error
	CreateEntries(ctx context.Context, entries []*domain.TransactionEntry) error
	GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.TransactionEntry, error)
	GetEntriesByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*domain.TransactionEntry, error)
	UpdateEntry(ctx context.Context, id uuid.UUID, req domain.TransactionEntry) (*domain.TransactionEntry, error)
	DeleteEntry(ctx context.Context, id uuid.UUID) error
	GetEntriesByAccountID(ctx context.Context, accountID uuid.UUID, limit int, offset int) ([]*domain.TransactionEntry, error)
	SearchEntries(ctx context.Context, query string, filters *domain.EntryFilter, limit int, offset int) ([]*domain.TransactionEntry, error)
	ReconcileEntries(ctx context.Context, entryIDs []uuid.UUID, reconciliationRef string) error
	UnreconcileEntries(ctx context.Context, entryIDs []uuid.UUID) error
	GetUnreconciledEntries(ctx context.Context, accountID uuid.UUID, cutoffDate *time.Time) ([]*domain.TransactionEntry, error)
	ValidateEntryConsistency(ctx context.Context, entry *domain.TransactionEntry) ([]domain.ValidationError, error)
	GetEntrySummary(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) (*domain.TransactionSummary, error)
}

type transactionEntryService struct {
	repo        domain.TransactionRepository
	accountRepo domain.AccountsRepository
	tracing     tracing.Service
	metrics     metrics.MetricsProvider
}

func NewTransactionEntryService(
	repo domain.TransactionRepository,
	accountRepo domain.AccountsRepository,
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
) TransactionEntryService {
	return &transactionEntryService{
		repo:        repo,
		accountRepo: accountRepo,
		tracing:     tracing,
		metrics:     metrics,
	}
}

func (s *transactionEntryService) CreateEntry(ctx context.Context, entry *domain.TransactionEntry) error {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_entry_service.create_entry",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("entry.transaction_id", entry.TransactionID.String()),
			attribute.String("entry.account_id", entry.AccountID.String()),
			attribute.Int("entry.entry_number", int(entry.EntryNumber)),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting entry creation",
		logger.Fields{
			"transaction_id": entry.TransactionID.String(),
			"account_id":     entry.AccountID.String(),
			"entry_number":   entry.EntryNumber,
			"description":    entry.Description,
		})

	if validationErrors := entry.Validate(); len(validationErrors) > 0 {
		s.metrics.IncrementCounter("entry_creation_errors", metrics.Fields{
			"error_type": "validation_error",
		})

		logger.WarnContext(ctx, "Entry validation failed",
			logger.Fields{
				"transaction_id": entry.TransactionID.String(),
				"account_id":     entry.AccountID.String(),
				"errors":         len(validationErrors),
			})

		return errors.NewBusinessError("VALIDATION_ERROR", "Entry validation failed")
	}

	account, err := s.accountRepo.GetByID(ctx, entry.AccountID)
	if err != nil {
		s.metrics.IncrementCounter("entry_creation_errors", metrics.Fields{
			"error_type": "invalid_account",
		})

		logger.WarnContext(ctx, "Account not found for entry",
			logger.Fields{"account_id": entry.AccountID.String()})

		return fmt.Errorf("account not found: %w", err)
	}

	if businessRuleErrors := entry.ValidateBusinessRules(account); len(businessRuleErrors) > 0 {
		s.metrics.IncrementCounter("entry_creation_errors", metrics.Fields{
			"error_type": "business_rule_error",
		})

		logger.WarnContext(ctx, "Entry business rule validation failed",
			logger.Fields{
				"transaction_id": entry.TransactionID.String(),
				"account_id":     entry.AccountID.String(),
				"errors":         len(businessRuleErrors),
			})

		return errors.NewBusinessError("BUSINESS_RULE_ERROR", "Entry business rule validation failed")
	}

	if currencyErrors := entry.ValidateAmountConsistency(); len(currencyErrors) > 0 {
		s.metrics.IncrementCounter("entry_creation_errors", metrics.Fields{
			"error_type": "currency_error",
		})

		logger.WarnContext(ctx, "Entry currency validation failed",
			logger.Fields{
				"transaction_id": entry.TransactionID.String(),
				"account_id":     entry.AccountID.String(),
				"errors":         len(currencyErrors),
			})

		return errors.NewBusinessError("CURRENCY_ERROR", "Entry currency validation failed")
	}

	timer := s.metrics.Timer("entry_creation_duration", metrics.Fields{})

	err = s.repo.CreateEntry(ctx, entry)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("entry_creation_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to create entry",
			logger.Fields{"error": err.Error()})

		return fmt.Errorf("failed to create entry: %w", err)
	}

	s.metrics.IncrementCounter("entries_created_total", metrics.Fields{
		"status": "success",
	})

	s.metrics.ObserveHistogram("entry_creation_duration",
		duration.Seconds(), metrics.Fields{})

	logger.InfoContext(ctx, "Entry created successfully",
		logger.Fields{
			"entry_id":       entry.ID.String(),
			"transaction_id": entry.TransactionID.String(),
			"account_id":     entry.AccountID.String(),
			"duration_ms":    duration.Milliseconds(),
		})

	return nil
}

func (s *transactionEntryService) CreateEntries(ctx context.Context, entries []*domain.TransactionEntry) error {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_entry_service.create_entries",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.Int("entries.count", len(entries)),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting bulk entry creation",
		logger.Fields{"entries_count": len(entries)})

	if len(entries) == 0 {
		return errors.NewBusinessError("VALIDATION_ERROR", "No entries provided")
	}

	var allErrors []domain.ValidationError
	accountCache := make(map[uuid.UUID]*domain.Accounts)

	for i, entry := range entries {
		if validationErrors := entry.Validate(); len(validationErrors) > 0 {
			for _, ve := range validationErrors {
				ve.Field = fmt.Sprintf("entries[%d].%s", i, ve.Field)
				allErrors = append(allErrors, ve)
			}
			continue
		}

		var account *domain.Accounts
		var ok bool
		if account, ok = accountCache[entry.AccountID]; !ok {
			var err error
			account, err = s.accountRepo.GetByID(ctx, entry.AccountID)
			if err != nil {
				allErrors = append(allErrors, domain.ValidationError{
					Field:   fmt.Sprintf("entries[%d].account_id", i),
					Message: "Account not found",
					Code:    "INVALID_ACCOUNT",
				})
				continue
			}
			accountCache[entry.AccountID] = account
		}

		if businessRuleErrors := entry.ValidateBusinessRules(account); len(businessRuleErrors) > 0 {
			for _, bre := range businessRuleErrors {
				bre.Field = fmt.Sprintf("entries[%d].%s", i, bre.Field)
				allErrors = append(allErrors, bre)
			}
		}

		if currencyErrors := entry.ValidateAmountConsistency(); len(currencyErrors) > 0 {
			for _, ce := range currencyErrors {
				ce.Field = fmt.Sprintf("entries[%d].%s", i, ce.Field)
				allErrors = append(allErrors, ce)
			}
		}
	}

	if len(allErrors) > 0 {
		s.metrics.IncrementCounter("bulk_entry_creation_errors", metrics.Fields{
			"error_type": "validation_error",
		})

		logger.WarnContext(ctx, "Bulk entry validation failed",
			logger.Fields{
				"entries_count": len(entries),
				"errors":        len(allErrors),
			})

		return errors.NewBusinessError("VALIDATION_ERROR", "Bulk entry validation failed")
	}

	timer := s.metrics.Timer("bulk_entry_creation_duration", metrics.Fields{
		"entries_count": len(entries),
	})

	err := s.repo.CreateEntries(ctx, entries)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("bulk_entry_creation_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to create bulk entries",
			logger.Fields{"error": err.Error()})

		return fmt.Errorf("failed to create bulk entries: %w", err)
	}

	s.metrics.IncrementCounter("bulk_entries_created_total", metrics.Fields{
		"entries_count": len(entries),
		"status":        "success",
	})

	s.metrics.ObserveHistogram("bulk_entry_creation_duration",
		duration.Seconds(), metrics.Fields{
			"entries_count": len(entries),
		})

	logger.InfoContext(ctx, "Bulk entries created successfully",
		logger.Fields{
			"entries_count": len(entries),
			"duration_ms":   duration.Milliseconds(),
		})

	return nil
}

func (s *transactionEntryService) GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.TransactionEntry, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_entry_service.get_entry_by_id",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("entry.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting entry by ID",
		logger.Fields{"entry_id": id.String()})

	entry, err := s.repo.GetEntryByID(ctx, id)
	if err != nil {
		if err == errors.ErrNotFound {
			logger.WarnContext(ctx, "Entry not found",
				logger.Fields{"entry_id": id.String()})
			return nil, errors.NewBusinessError("NOT_FOUND", "Entry not found")
		}

		logger.ErrorContext(ctx, "Failed to get entry by ID",
			logger.Fields{
				"entry_id": id.String(),
				"error":    err.Error(),
			})
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	logger.DebugContext(ctx, "Entry retrieved successfully",
		logger.Fields{
			"entry_id":       entry.ID.String(),
			"transaction_id": entry.TransactionID.String(),
			"account_id":     entry.AccountID.String(),
		})

	return entry, nil
}

func (s *transactionEntryService) GetEntriesByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*domain.TransactionEntry, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_entry_service.get_entries_by_transaction_id",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("transaction.id", transactionID.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting entries by transaction ID",
		logger.Fields{"transaction_id": transactionID.String()})

	entries, err := s.repo.GetEntriesByTransaction(ctx, transactionID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get entries by transaction ID",
			logger.Fields{
				"transaction_id": transactionID.String(),
				"error":          err.Error(),
			})
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	// convert []domain.TransactionEntry to []*domain.TransactionEntry
	result := make([]*domain.TransactionEntry, len(entries))
	for i := range entries {
		result[i] = &entries[i]
	}

	logger.DebugContext(ctx, "Entries retrieved successfully",
		logger.Fields{
			"transaction_id": transactionID.String(),
			"entries_count":  len(entries),
		})

	return result, nil
}

func (s *transactionEntryService) UpdateEntry(ctx context.Context, id uuid.UUID, req domain.TransactionEntry) (*domain.TransactionEntry, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_entry_service.update_entry",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("entry.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting entry update",
		logger.Fields{"entry_id": id.String()})

	_, err := s.repo.GetEntryByID(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Entry not found for update",
			logger.Fields{
				"entry_id": id.String(),
				"error":    err.Error(),
			})
		return nil, err
	}

	if errs := req.Validate(); len(errs) > 0 {
		s.metrics.IncrementCounter("entry_update_errors", metrics.Fields{
			"error_type": "validation_error",
		})

		logger.WarnContext(ctx, "Entry update validation failed",
			logger.Fields{
				"entry_id": id.String(),
				"errors":   len(errs),
			})

		return nil, errors.NewBusinessError("VALIDATION_ERROR", "Entry update validation failed")
	}

	timer := s.metrics.Timer("entry_update_duration", metrics.Fields{})

	err = s.repo.UpdateEntry(ctx, &req)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("entry_update_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to update entry",
			logger.Fields{
				"entry_id": id.String(),
				"error":    err.Error(),
			})

		return nil, fmt.Errorf("failed to update entry: %w", err)
	}

	// Re-fetch to return the post-update state, not the stale pre-update value.
	updatedEntry, fetchErr := s.repo.GetEntryByID(ctx, id)
	if fetchErr != nil {
		return nil, fmt.Errorf("entry updated but re-fetch failed: %w", fetchErr)
	}

	s.metrics.IncrementCounter("entries_updated_total", metrics.Fields{
		"status": "success",
	})

	s.metrics.ObserveHistogram("entry_update_duration",
		duration.Seconds(), metrics.Fields{})

	logger.InfoContext(ctx, "Entry updated successfully",
		logger.Fields{
			"entry_id":    updatedEntry.ID.String(),
			"duration_ms": duration.Milliseconds(),
		})

	return updatedEntry, nil
}

func (s *transactionEntryService) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_entry_service.delete_entry",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("entry.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting entry deletion",
		logger.Fields{"entry_id": id.String()})

	entry, err := s.repo.GetEntryByID(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Entry not found for deletion",
			logger.Fields{
				"entry_id": id.String(),
				"error":    err.Error(),
			})
		return err
	}

	if entry.Reconciled {
		s.metrics.IncrementCounter("entry_deletion_errors", metrics.Fields{
			"error_type": "reconciled_entry",
		})

		logger.WarnContext(ctx, "Cannot delete reconciled entry",
			logger.Fields{"entry_id": id.String()})

		return errors.NewBusinessError("RECONCILED_ENTRY", "Cannot delete reconciled entry").
			WithHTTPStatus(http.StatusUnprocessableEntity).
			WithCategory(errors.CategoryBusiness)
	}

	// Guard: deleting an entry from a POSTED transaction corrupts the general ledger.
	parentTxn, err := s.repo.GetByID(ctx, entry.TransactionID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to fetch parent transaction for entry deletion guard",
			logger.Fields{
				"entry_id":       id.String(),
				"transaction_id": entry.TransactionID.String(),
				"error":          err.Error(),
			})
		return fmt.Errorf("failed to verify parent transaction status: %w", err)
	}
	if parentTxn.TransactionStatus == domain.TransactionStatusPosted ||
		parentTxn.TransactionStatus == domain.TransactionStatusReversed {
		s.metrics.IncrementCounter("entry_deletion_errors", metrics.Fields{
			"error_type": "posted_transaction",
		})

		logger.WarnContext(ctx, "Cannot delete entry from posted/reversed transaction",
			logger.Fields{
				"entry_id":       id.String(),
				"transaction_id": entry.TransactionID.String(),
				"status":         string(parentTxn.TransactionStatus),
			})

		return errors.NewBusinessError("ENTRY_IMMUTABLE",
			fmt.Sprintf("entries on %q transactions cannot be deleted; reverse the transaction instead",
				parentTxn.TransactionStatus)).
			WithHTTPStatus(http.StatusUnprocessableEntity).
			WithCategory(errors.CategoryBusiness)
	}

	timer := s.metrics.Timer("entry_deletion_duration", metrics.Fields{})

	err = s.repo.DeleteEntry(ctx, id)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("entry_deletion_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to delete entry",
			logger.Fields{
				"entry_id": id.String(),
				"error":    err.Error(),
			})

		return fmt.Errorf("failed to delete entry: %w", err)
	}

	s.metrics.IncrementCounter("entries_deleted_total", metrics.Fields{
		"status": "success",
	})

	s.metrics.ObserveHistogram("entry_deletion_duration",
		duration.Seconds(), metrics.Fields{})

	logger.InfoContext(ctx, "Entry deleted successfully",
		logger.Fields{
			"entry_id":    id.String(),
			"duration_ms": duration.Milliseconds(),
		})

	return nil
}

func (s *transactionEntryService) GetEntriesByAccountID(ctx context.Context, accountID uuid.UUID, limit int, offset int) ([]*domain.TransactionEntry, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_entry_service.get_entries_by_account_id",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.id", accountID.String()),
			attribute.Int("limit", limit),
			attribute.Int("offset", offset),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting entries by account ID",
		logger.Fields{
			"account_id": accountID.String(),
			"limit":      limit,
			"offset":     offset,
		})

	if limit <= 0 {
		limit = 100
	}

	if limit > 1000 {
		limit = 1000
	}

	entries, err := s.repo.GetEntriesByAccount(ctx, accountID, &domain.EntryFilter{Limit: &limit, Offset: &offset})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get entries by account ID",
			logger.Fields{
				"account_id": accountID.String(),
				"error":      err.Error(),
			})
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	result := make([]*domain.TransactionEntry, len(entries))
	for i := range entries {
		result[i] = &entries[i]
	}

	logger.DebugContext(ctx, "Entries by account ID retrieved successfully",
		logger.Fields{
			"account_id":    accountID.String(),
			"entries_count": len(entries),
		})

	return result, nil
}

func (s *transactionEntryService) SearchEntries(ctx context.Context, query string, filters *domain.EntryFilter, limit int, offset int) ([]*domain.TransactionEntry, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_entry_service.search_entries",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("search.query", query),
			attribute.Int("search.limit", limit),
			attribute.Int("search.offset", offset),
		))
	defer span.End()

	logger.DebugContext(ctx, "Searching entries",
		logger.Fields{
			"query":  query,
			"limit":  limit,
			"offset": offset,
		})

	if limit <= 0 {
		limit = 100
	}

	if limit > 500 {
		limit = 500
	}

	entries, err := s.repo.SearchEntries(ctx, query, filters, limit, offset)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to search entries",
			logger.Fields{
				"query": query,
				"error": err.Error(),
			})
		return nil, fmt.Errorf("failed to search entries: %w", err)
	}

	logger.DebugContext(ctx, "Entry search completed successfully",
		logger.Fields{
			"query":         query,
			"results_count": len(entries),
		})

	return entries, nil
}

func (s *transactionEntryService) ReconcileEntries(ctx context.Context, entryIDs []uuid.UUID, reconciliationRef string) error {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_entry_service.reconcile_entries",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.Int("entries.count", len(entryIDs)),
			attribute.String("reconciliation.reference", reconciliationRef),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting entry reconciliation",
		logger.Fields{
			"entries_count":      len(entryIDs),
			"reconciliation_ref": reconciliationRef,
		})

	if len(entryIDs) == 0 {
		return errors.NewBusinessError("VALIDATION_ERROR", "No entries provided for reconciliation")
	}

	for _, entryID := range entryIDs {
		entry, err := s.repo.GetEntryByID(ctx, entryID)
		if err != nil {
			// Not-found during reconciliation is an error, not a skip.
			// Silently continuing would report success when entries are missing.
			s.metrics.IncrementCounter("reconciliation_errors", metrics.Fields{
				"error_type": "entry_not_found",
			})
			logger.ErrorContext(ctx, "Entry not found during reconciliation — aborting batch",
				logger.Fields{
					"entry_id":           entryID.String(),
					"reconciliation_ref": reconciliationRef,
					"error":              err.Error(),
				})
			return fmt.Errorf("reconciliation aborted: entry %s not found: %w", entryID.String(), err)
		}

		if entry.Reconciled {
			// Already reconciled is idempotent — skip without error.
			logger.WarnContext(ctx, "Entry already reconciled — skipping",
				logger.Fields{
					"entry_id":                entryID.String(),
					"existing_reconcile_ref":  fmt.Sprintf("%v", entry.ReconciliationReference),
				})
			continue
		}

		entry.MarkReconciled(reconciliationRef)

		if err := s.repo.UpdateReconciliationStatus(ctx, entryID, true, entry.ReconciledDate, &reconciliationRef); err != nil {
			s.metrics.IncrementCounter("reconciliation_errors", metrics.Fields{
				"error_type": "update_failed",
			})
			logger.ErrorContext(ctx, "Failed to mark entry as reconciled",
				logger.Fields{
					"entry_id": entryID.String(),
					"error":    err.Error(),
				})
			return fmt.Errorf("failed to reconcile entry %s: %w", entryID.String(), err)
		}
	}

	s.metrics.IncrementCounter("entries_reconciled_total", metrics.Fields{
		"entries_count": len(entryIDs),
		"status":        "success",
	})

	logger.InfoContext(ctx, "Entries reconciled successfully",
		logger.Fields{
			"entries_count":      len(entryIDs),
			"reconciliation_ref": reconciliationRef,
		})

	return nil
}

func (s *transactionEntryService) UnreconcileEntries(ctx context.Context, entryIDs []uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_entry_service.unreconcile_entries",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.Int("entries.count", len(entryIDs)),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting entry unreconciliation",
		logger.Fields{"entries_count": len(entryIDs)})

	if len(entryIDs) == 0 {
		return errors.NewBusinessError("VALIDATION_ERROR", "No entries provided for unreconciliation")
	}

	for _, entryID := range entryIDs {
		entry, err := s.repo.GetEntryByID(ctx, entryID)
		if err != nil {
			// Not-found is an error — caller may have wrong IDs.
			s.metrics.IncrementCounter("reconciliation_errors", metrics.Fields{
				"error_type": "entry_not_found_on_unreconcile",
			})
			logger.ErrorContext(ctx, "Entry not found during unreconciliation — aborting batch",
				logger.Fields{
					"entry_id": entryID.String(),
					"error":    err.Error(),
				})
			return fmt.Errorf("unreconciliation aborted: entry %s not found: %w", entryID.String(), err)
		}

		if !entry.Reconciled {
			// Not reconciled — idempotent skip.
			logger.WarnContext(ctx, "Entry not reconciled — skipping unreconcile",
				logger.Fields{"entry_id": entryID.String()})
			continue
		}

		entry.UnmarkReconciled()

		if err := s.repo.UpdateReconciliationStatus(ctx, entryID, false, nil, nil); err != nil {
			logger.ErrorContext(ctx, "Failed to mark entry as unreconciled",
				logger.Fields{
					"entry_id": entryID.String(),
					"error":    err.Error(),
				})
			return fmt.Errorf("failed to unreconcile entry %s: %w", entryID.String(), err)
		}
	}

	s.metrics.IncrementCounter("entries_unreconciled_total", metrics.Fields{
		"entries_count": len(entryIDs),
		"status":        "success",
	})

	logger.InfoContext(ctx, "Entries unreconciled successfully",
		logger.Fields{"entries_count": len(entryIDs)})

	return nil
}

func (s *transactionEntryService) GetUnreconciledEntries(ctx context.Context, accountID uuid.UUID, cutoffDate *time.Time) ([]*domain.TransactionEntry, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_entry_service.get_unreconciled_entries",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.id", accountID.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting unreconciled entries",
		logger.Fields{"account_id": accountID.String()})

	entries, err := s.repo.GetUnreconciledEntries(ctx, accountID, cutoffDate)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get unreconciled entries",
			logger.Fields{
				"account_id": accountID.String(),
				"error":      err.Error(),
			})
		return nil, fmt.Errorf("failed to get unreconciled entries: %w", err)
	}

	logger.DebugContext(ctx, "Unreconciled entries retrieved successfully",
		logger.Fields{
			"account_id":    accountID.String(),
			"entries_count": len(entries),
		})

	return entries, nil
}

func (s *transactionEntryService) ValidateEntryConsistency(ctx context.Context, entry *domain.TransactionEntry) ([]domain.ValidationError, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_entry_service.validate_entry_consistency",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("entry.id", entry.ID.String()),
		))
	defer span.End()

	var validationErrors []domain.ValidationError

	if basicErrors := entry.Validate(); len(basicErrors) > 0 {
		validationErrors = append(validationErrors, basicErrors...)
	}

	account, err := s.accountRepo.GetByID(ctx, entry.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account for validation: %w", err)
	}

	if businessRuleErrors := entry.ValidateBusinessRules(account); len(businessRuleErrors) > 0 {
		validationErrors = append(validationErrors, businessRuleErrors...)
	}

	if currencyErrors := entry.ValidateAmountConsistency(); len(currencyErrors) > 0 {
		validationErrors = append(validationErrors, currencyErrors...)
	}

	return validationErrors, nil
}

func (s *transactionEntryService) GetEntrySummary(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) (*domain.TransactionSummary, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_entry_service.get_entry_summary",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.id", accountID.String()),
			attribute.String("date_range.start", startDate.Format("2006-01-02")),
			attribute.String("date_range.end", endDate.Format("2006-01-02")),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting entry summary",
		logger.Fields{
			"account_id": accountID.String(),
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		})

	summary, err := s.repo.GetEntrySummary(ctx, accountID, startDate, endDate)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get entry summary",
			logger.Fields{
				"account_id": accountID.String(),
				"start_date": startDate.Format("2006-01-02"),
				"end_date":   endDate.Format("2006-01-02"),
				"error":      err.Error(),
			})
		return nil, fmt.Errorf("failed to get entry summary: %w", err)
	}

	logger.DebugContext(ctx, "Entry summary retrieved successfully",
		logger.Fields{
			"account_id":   accountID.String(),
			"start_date":   startDate.Format("2006-01-02"),
			"end_date":     endDate.Format("2006-01-02"),
			"entry_count":  summary.TransactionCount,
			"total_debits": summary.TotalDebit.String(),
		})

	return summary, nil
}
