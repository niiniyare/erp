package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/attribute"
)

type TransactionService interface {
	CreateTransaction(ctx context.Context, req domain.CreateTransactionRequest) (*domain.Transaction, error)
	GetTransactionByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error)
	GetTransactionByNumber(ctx context.Context, number string) (*domain.Transaction, error)
	UpdateTransaction(ctx context.Context, id uuid.UUID, req domain.Transaction) (*domain.Transaction, error)
	DeleteTransaction(ctx context.Context, id uuid.UUID) error
	ListTransactions(ctx context.Context, req *domain.TransactionFilter) ([]*domain.Transaction, error)
	PostTransaction(ctx context.Context, id uuid.UUID, postingDate *time.Time) (*domain.Transaction, error)
	ReverseTransaction(ctx context.Context, id uuid.UUID, reason string) (*domain.Transaction, error)
	ApproveTransaction(ctx context.Context, id uuid.UUID, notes string) (*domain.Transaction, error)
	RejectTransaction(ctx context.Context, id uuid.UUID, notes string) (*domain.Transaction, error)
	// GetTransactionWithEntries - TODO: Implement with proper type
	// GetTransactionWithEntries(ctx context.Context, id uuid.UUID) (*domain.TransactionWithEntries, error)
	ValidateTransaction(ctx context.Context, transaction *domain.Transaction, entries []*domain.TransactionEntry) ([]domain.ValidationError, error)
	SearchTransactions(ctx context.Context, query string, limit int, offset int) ([]*domain.Transaction, error)
	GetTransactionSummary(ctx context.Context, startDate, endDate time.Time) (*domain.TransactionSummary, error)
	GetPendingApprovalTransactions(ctx context.Context, limit int, offset int) ([]*domain.Transaction, error)
	GetRecurringTransactionsDue(ctx context.Context, date time.Time) ([]*domain.Transaction, error)
	CreateRecurringTransaction(ctx context.Context, templateID uuid.UUID, date time.Time) (*domain.Transaction, error)
}

type transactionService struct {
	repo         domain.TransactionRepository
	accountRepo  domain.ChartOfAccountsRepository
	entryService TransactionEntryService
	tracing      tracing.TracingService
	metrics      metrics.MetricsProvider
}

func NewTransactionService(
	repo domain.TransactionRepository,
	accountRepo domain.ChartOfAccountsRepository,
	entryService TransactionEntryService,
	tracing tracing.TracingService,
	metrics metrics.MetricsProvider,
) TransactionService {
	return &transactionService{
		repo:         repo,
		accountRepo:  accountRepo,
		entryService: entryService,
		tracing:      tracing,
		metrics:      metrics,
	}
}

func (s *transactionService) CreateTransaction(ctx context.Context, req domain.CreateTransactionRequest) (*domain.Transaction, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_service.create_transaction",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("transaction.type", string(req.TransactionType)),
			attribute.String("transaction.number", req.TransactionNumber),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting transaction creation",
		logger.Fields{
			"transaction_type":   string(req.TransactionType),
			"transaction_number": req.TransactionNumber,
			"description":        req.Description,
		})

	if err := req.Validate(); err != nil {
		s.metrics.IncrementCounter("transaction_creation_errors", metrics.Fields{
			"error_type": "validation_error",
		})

		logger.WarnContext(ctx, "Transaction validation failed",
			logger.Fields{
				"transaction_number": req.TransactionNumber,
				"errors":             len(err),
			})

		return nil, domain.NewValidationError("transaction_validation", "Transaction validation failed", err)
	}

	if err := s.repo.ValidateTransactionNumber(ctx, req.TransactionNumber, nil); err != nil {
		s.metrics.IncrementCounter("transaction_creation_errors", metrics.Fields{
			"error_type": "duplicate_number",
		})

		logger.WarnContext(ctx, "Transaction number already exists",
			logger.Fields{"transaction_number": req.TransactionNumber})

		return nil, err
	}

	timer := s.metrics.Timer("transaction_creation_duration", metrics.Fields{
		"transaction_type": string(req.TransactionType),
	})

	transaction, err := s.repo.Create(ctx, &req)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("transaction_creation_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to create transaction",
			logger.Fields{"error": err.Error()})

		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	s.metrics.IncrementCounter("transactions_created_total", metrics.Fields{
		"transaction_type": string(req.TransactionType),
		"status":           "success",
	})

	s.metrics.ObserveHistogram("transaction_creation_duration",
		duration.Seconds(), metrics.Fields{
			"transaction_type": string(req.TransactionType),
		})

	logger.InfoContext(ctx, "Transaction created successfully",
		logger.Fields{
			"transaction_id":     transaction.ID.String(),
			"transaction_number": transaction.TransactionNumber,
			"duration_ms":        duration.Milliseconds(),
		})

	return transaction, nil
}

func (s *transactionService) GetTransactionByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_service.get_transaction_by_id",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("transaction.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting transaction by ID",
		logger.Fields{"transaction_id": id.String()})

	transaction, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == errors.ErrNotFound {
			logger.WarnContext(ctx, "Transaction not found",
				logger.Fields{"transaction_id": id.String()})
			return nil, domain.NewNotFoundError("transaction", id.String())
		}

		logger.ErrorContext(ctx, "Failed to get transaction by ID",
			logger.Fields{
				"transaction_id": id.String(),
				"error":          err.Error(),
			})
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	logger.DebugContext(ctx, "Transaction retrieved successfully",
		logger.Fields{
			"transaction_id":     transaction.ID.String(),
			"transaction_number": transaction.TransactionNumber,
		})

	return transaction, nil
}

func (s *transactionService) GetTransactionByNumber(ctx context.Context, number string) (*domain.Transaction, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_service.get_transaction_by_number",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("transaction.number", number),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting transaction by number",
		logger.Fields{"transaction_number": number})

	transaction, err := s.repo.GetByNumber(ctx, number)
	if err != nil {
		if err == errors.ErrNotFound {
			logger.WarnContext(ctx, "Transaction not found",
				logger.Fields{"transaction_number": number})
			return nil, domain.NewNotFoundError("transaction", number)
		}

		logger.ErrorContext(ctx, "Failed to get transaction by number",
			logger.Fields{
				"transaction_number": number,
				"error":              err.Error(),
			})
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	logger.DebugContext(ctx, "Transaction retrieved successfully",
		logger.Fields{
			"transaction_id":     transaction.ID.String(),
			"transaction_number": transaction.TransactionNumber,
		})

	return transaction, nil
}

func (s *transactionService) UpdateTransaction(ctx context.Context, id uuid.UUID, req domain.Transaction) (*domain.Transaction, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_service.update_transaction",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("transaction.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting transaction update",
		logger.Fields{"transaction_id": id.String()})

	existingTransaction, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Transaction not found for update",
			logger.Fields{
				"transaction_id": id.String(),
				"error":          err.Error(),
			})
		return nil, err
	}

	if existingTransaction.TransactionStatus == domain.TransactionStatusPosted {
		s.metrics.IncrementCounter("transaction_update_errors", metrics.Fields{
			"error_type": "posted_transaction",
		})

		logger.WarnContext(ctx, "Cannot update posted transaction",
			logger.Fields{"transaction_id": id.String()})

		return nil, domain.NewBusinessRuleError("posted_transaction", "Cannot update posted transaction")
	}

	if err := req.Validate(); err != nil {
		s.metrics.IncrementCounter("transaction_update_errors", metrics.Fields{
			"error_type": "validation_error",
		})

		logger.WarnContext(ctx, "Transaction update validation failed",
			logger.Fields{
				"transaction_id": id.String(),
				"errors":         len(err),
			})

		return nil, domain.NewValidationError("transaction_update_validation", "Transaction update validation failed", err)
	}

	timer := s.metrics.Timer("transaction_update_duration", metrics.Fields{
		"transaction_type": string(existingTransaction.TransactionType),
	})

	updatedTransaction, err := s.repo.Update(ctx, id, &req)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("transaction_update_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to update transaction",
			logger.Fields{
				"transaction_id": id.String(),
				"error":          err.Error(),
			})

		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	s.metrics.IncrementCounter("transactions_updated_total", metrics.Fields{
		"transaction_type": string(existingTransaction.TransactionType),
		"status":           "success",
	})

	s.metrics.ObserveHistogram("transaction_update_duration",
		duration.Seconds(), metrics.Fields{
			"transaction_type": string(existingTransaction.TransactionType),
		})

	logger.InfoContext(ctx, "Transaction updated successfully",
		logger.Fields{
			"transaction_id":     updatedTransaction.ID.String(),
			"transaction_number": updatedTransaction.TransactionNumber,
			"duration_ms":        duration.Milliseconds(),
		})

	return updatedTransaction, nil
}

func (s *transactionService) DeleteTransaction(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_service.delete_transaction",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("transaction.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting transaction deletion",
		logger.Fields{"transaction_id": id.String()})

	transaction, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Transaction not found for deletion",
			logger.Fields{
				"transaction_id": id.String(),
				"error":          err.Error(),
			})
		return err
	}

	if transaction.TransactionStatus == domain.TransactionStatusPosted {
		s.metrics.IncrementCounter("transaction_deletion_errors", metrics.Fields{
			"error_type": "posted_transaction",
		})

		logger.WarnContext(ctx, "Cannot delete posted transaction",
			logger.Fields{"transaction_id": id.String()})

		return domain.NewBusinessRuleError("posted_transaction", "Cannot delete posted transaction")
	}

	timer := s.metrics.Timer("transaction_deletion_duration", metrics.Fields{
		"transaction_type": string(transaction.TransactionType),
	})

	err = s.repo.Delete(ctx, id)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("transaction_deletion_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to delete transaction",
			logger.Fields{
				"transaction_id": id.String(),
				"error":          err.Error(),
			})

		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	s.metrics.IncrementCounter("transactions_deleted_total", metrics.Fields{
		"transaction_type": string(transaction.TransactionType),
		"status":           "success",
	})

	s.metrics.ObserveHistogram("transaction_deletion_duration",
		duration.Seconds(), metrics.Fields{
			"transaction_type": string(transaction.TransactionType),
		})

	logger.InfoContext(ctx, "Transaction deleted successfully",
		logger.Fields{
			"transaction_id":     id.String(),
			"transaction_number": transaction.TransactionNumber,
			"duration_ms":        duration.Milliseconds(),
		})

	return nil
}

func (s *transactionService) ListTransactions(ctx context.Context, req *domain.TransactionFilter) ([]*domain.Transaction, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_service.list_transactions",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.Int("limit", req.Limit),
			attribute.Int("offset", req.Offset),
		))
	defer span.End()

	logger.DebugContext(ctx, "Listing transactions",
		logger.Fields{
			"limit":  req.Limit,
			"offset": req.Offset,
		})

	if req.Limit <= 0 {
		req.Limit = 50
	}

	if req.Limit > 1000 {
		req.Limit = 1000
	}

	timer := s.metrics.Timer("transaction_list_duration", metrics.Fields{})

	transactions, err := s.repo.List(ctx, &req)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("transaction_list_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to list transactions",
			logger.Fields{"error": err.Error()})

		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	s.metrics.IncrementCounter("transaction_list_requests_total", metrics.Fields{
		"status": "success",
	})

	s.metrics.ObserveHistogram("transaction_list_duration",
		duration.Seconds(), metrics.Fields{})

	logger.DebugContext(ctx, "Transactions listed successfully",
		logger.Fields{
			"count":       len(transactions),
			"duration_ms": duration.Milliseconds(),
		})

	return transactions, nil
}

func (s *transactionService) PostTransaction(ctx context.Context, id uuid.UUID, postingDate *time.Time) (*domain.Transaction, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_service.post_transaction",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("transaction.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting transaction posting",
		logger.Fields{"transaction_id": id.String()})

	transaction, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if transaction.TransactionStatus != domain.TransactionStatusApproved &&
		transaction.TransactionStatus != domain.TransactionStatusDraft {
		s.metrics.IncrementCounter("transaction_posting_errors", metrics.Fields{
			"error_type": "invalid_status",
		})

		logger.WarnContext(ctx, "Transaction cannot be posted",
			logger.Fields{
				"transaction_id":     id.String(),
				"transaction_status": string(transaction.TransactionStatus),
			})

		return nil, domain.NewBusinessRuleError("invalid_status", "Transaction must be approved or draft to be posted")
	}

	entries, err := s.entryService.GetEntriesByTransactionID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction entries: %w", err)
	}

	if len(entries) == 0 {
		s.metrics.IncrementCounter("transaction_posting_errors", metrics.Fields{
			"error_type": "no_entries",
		})

		logger.WarnContext(ctx, "Cannot post transaction without entries",
			logger.Fields{"transaction_id": id.String()})

		return nil, domain.NewBusinessRuleError("no_entries", "Cannot post transaction without entries")
	}

	validationErrors, err := s.ValidateTransaction(ctx, transaction, entries)
	if err != nil {
		return nil, fmt.Errorf("failed to validate transaction: %w", err)
	}

	if len(validationErrors) > 0 {
		s.metrics.IncrementCounter("transaction_posting_errors", metrics.Fields{
			"error_type": "validation_failed",
		})

		logger.WarnContext(ctx, "Transaction validation failed",
			logger.Fields{
				"transaction_id": id.String(),
				"errors":         len(validationErrors),
			})

		return nil, domain.NewValidationError("transaction_posting_validation", "Transaction validation failed", validationErrors)
	}

	if postingDate == nil {
		now := time.Now()
		postingDate = &now
	}

	timer := s.metrics.Timer("transaction_posting_duration", metrics.Fields{
		"transaction_type": string(transaction.TransactionType),
	})

	postedTransaction, err := s.repo.Post(ctx, id, *postingDate)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("transaction_posting_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to post transaction",
			logger.Fields{
				"transaction_id": id.String(),
				"error":          err.Error(),
			})

		return nil, fmt.Errorf("failed to post transaction: %w", err)
	}

	if err := s.updateAccountBalances(ctx, entries); err != nil {
		logger.ErrorContext(ctx, "Failed to update account balances after posting",
			logger.Fields{
				"transaction_id": id.String(),
				"error":          err.Error(),
			})
	}

	s.metrics.IncrementCounter("transactions_posted_total", metrics.Fields{
		"transaction_type": string(transaction.TransactionType),
		"status":           "success",
	})

	s.metrics.ObserveHistogram("transaction_posting_duration",
		duration.Seconds(), metrics.Fields{
			"transaction_type": string(transaction.TransactionType),
		})

	logger.InfoContext(ctx, "Transaction posted successfully",
		logger.Fields{
			"transaction_id":     postedTransaction.ID.String(),
			"transaction_number": postedTransaction.TransactionNumber,
			"posting_date":       postingDate.Format("2006-01-02"),
			"duration_ms":        duration.Milliseconds(),
		})

	return postedTransaction, nil
}

func (s *transactionService) ReverseTransaction(ctx context.Context, id uuid.UUID, reason string) (*domain.Transaction, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_service.reverse_transaction",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("transaction.id", id.String()),
			attribute.String("reversal.reason", reason),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting transaction reversal",
		logger.Fields{
			"transaction_id": id.String(),
			"reason":         reason,
		})

	transaction, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if transaction.TransactionStatus != domain.TransactionStatusPosted {
		s.metrics.IncrementCounter("transaction_reversal_errors", metrics.Fields{
			"error_type": "not_posted",
		})

		logger.WarnContext(ctx, "Only posted transactions can be reversed",
			logger.Fields{
				"transaction_id":     id.String(),
				"transaction_status": string(transaction.TransactionStatus),
			})

		return nil, domain.NewBusinessRuleError("not_posted", "Only posted transactions can be reversed")
	}

	if transaction.IsReversed {
		s.metrics.IncrementCounter("transaction_reversal_errors", metrics.Fields{
			"error_type": "already_reversed",
		})

		logger.WarnContext(ctx, "Transaction is already reversed",
			logger.Fields{"transaction_id": id.String()})

		return nil, domain.NewBusinessRuleError("already_reversed", "Transaction is already reversed")
	}

	entries, err := s.entryService.GetEntriesByTransactionID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction entries: %w", err)
	}

	reversalEntries := make([]*domain.TransactionEntry, len(entries))
	for i, entry := range entries {
		reversalEntry := entry.Clone()
		if entry.IsDebit() {
			reversalEntry.CreditAmount = entry.DebitAmount
			reversalEntry.DebitAmount = decimal.Zero
		} else {
			reversalEntry.DebitAmount = entry.CreditAmount
			reversalEntry.CreditAmount = decimal.Zero
		}
		reversalEntry.Description = fmt.Sprintf("REVERSAL: %s", entry.Description)
		reversalEntries[i] = reversalEntry
	}

	reversalTransaction := &domain.CreateTransactionRequest{
		TransactionNumber: fmt.Sprintf("REV-%s", transaction.TransactionNumber),
		TransactionType:   transaction.TransactionType,
		TransactionDate:   time.Now(),
		Description:       fmt.Sprintf("REVERSAL: %s - %s", transaction.Description, reason),
		ReferenceNumber:   &transaction.TransactionNumber,
		CurrencyCode:      transaction.CurrencyCode,
		ExchangeRate:      transaction.ExchangeRate,
	}

	createdReversal, err := s.repo.Create(ctx, reversalTransaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create reversal transaction: %w", err)
	}

	for i, entry := range reversalEntries {
		entry.TransactionID = createdReversal.ID
		entry.EntryNumber = int32(i + 1)
		if err := s.entryService.CreateEntry(ctx, entry); err != nil {
			return nil, fmt.Errorf("failed to create reversal entry: %w", err)
		}
	}

	if err := s.repo.MarkAsReversed(ctx, id, createdReversal.ID, reason); err != nil {
		return nil, fmt.Errorf("failed to mark transaction as reversed: %w", err)
	}

	_, err = s.PostTransaction(ctx, createdReversal.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to post reversal transaction: %w", err)
	}

	logger.InfoContext(ctx, "Transaction reversed successfully",
		logger.Fields{
			"original_transaction_id": id.String(),
			"reversal_transaction_id": createdReversal.ID.String(),
			"reason":                  reason,
		})

	return createdReversal, nil
}

func (s *transactionService) ApproveTransaction(ctx context.Context, id uuid.UUID, notes string) (*domain.Transaction, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_service.approve_transaction",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("transaction.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting transaction approval",
		logger.Fields{
			"transaction_id": id.String(),
			"notes":          notes,
		})

	transaction, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !transaction.ApprovalRequired || transaction.ApprovalStatus != domain.ApprovalStatusPending {
		s.metrics.IncrementCounter("transaction_approval_errors", metrics.Fields{
			"error_type": "invalid_status",
		})

		logger.WarnContext(ctx, "Transaction cannot be approved",
			logger.Fields{
				"transaction_id":    id.String(),
				"approval_required": transaction.ApprovalRequired,
				"approval_status":   string(transaction.ApprovalStatus),
			})

		return nil, domain.NewBusinessRuleError("invalid_status", "Transaction is not pending approval")
	}

	timer := s.metrics.Timer("transaction_approval_duration", metrics.Fields{
		"transaction_type": string(transaction.TransactionType),
	})

	approvedTransaction, err := s.repo.Approve(ctx, id, notes)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("transaction_approval_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to approve transaction",
			logger.Fields{
				"transaction_id": id.String(),
				"error":          err.Error(),
			})

		return nil, fmt.Errorf("failed to approve transaction: %w", err)
	}

	s.metrics.IncrementCounter("transactions_approved_total", metrics.Fields{
		"transaction_type": string(transaction.TransactionType),
		"status":           "success",
	})

	s.metrics.ObserveHistogram("transaction_approval_duration",
		duration.Seconds(), metrics.Fields{
			"transaction_type": string(transaction.TransactionType),
		})

	logger.InfoContext(ctx, "Transaction approved successfully",
		logger.Fields{
			"transaction_id":     approvedTransaction.ID.String(),
			"transaction_number": approvedTransaction.TransactionNumber,
			"duration_ms":        duration.Milliseconds(),
		})

	return approvedTransaction, nil
}

func (s *transactionService) RejectTransaction(ctx context.Context, id uuid.UUID, notes string) (*domain.Transaction, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_service.reject_transaction",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("transaction.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting transaction rejection",
		logger.Fields{
			"transaction_id": id.String(),
			"notes":          notes,
		})

	transaction, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !transaction.ApprovalRequired || transaction.ApprovalStatus != domain.ApprovalStatusPending {
		s.metrics.IncrementCounter("transaction_rejection_errors", metrics.Fields{
			"error_type": "invalid_status",
		})

		logger.WarnContext(ctx, "Transaction cannot be rejected",
			logger.Fields{
				"transaction_id":    id.String(),
				"approval_required": transaction.ApprovalRequired,
				"approval_status":   string(transaction.ApprovalStatus),
			})

		return nil, domain.NewBusinessRuleError("invalid_status", "Transaction is not pending approval")
	}

	timer := s.metrics.Timer("transaction_rejection_duration", metrics.Fields{
		"transaction_type": string(transaction.TransactionType),
	})

	rejectedTransaction, err := s.repo.Reject(ctx, id, notes)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("transaction_rejection_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to reject transaction",
			logger.Fields{
				"transaction_id": id.String(),
				"error":          err.Error(),
			})

		return nil, fmt.Errorf("failed to reject transaction: %w", err)
	}

	s.metrics.IncrementCounter("transactions_rejected_total", metrics.Fields{
		"transaction_type": string(transaction.TransactionType),
		"status":           "success",
	})

	s.metrics.ObserveHistogram("transaction_rejection_duration",
		duration.Seconds(), metrics.Fields{
			"transaction_type": string(transaction.TransactionType),
		})

	logger.InfoContext(ctx, "Transaction rejected successfully",
		logger.Fields{
			"transaction_id":     rejectedTransaction.ID.String(),
			"transaction_number": rejectedTransaction.TransactionNumber,
			"duration_ms":        duration.Milliseconds(),
		})

	return rejectedTransaction, nil
}

// TODO: Implement TransactionWithEntries type
/*
func (s *transactionService) GetTransactionWithEntries(ctx context.Context, id uuid.UUID) (*domain.TransactionWithEntries, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_service.get_transaction_with_entries",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("transaction.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting transaction with entries",
		logger.Fields{"transaction_id": id.String()})

	transactionWithEntries, err := s.repo.GetWithEntries(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get transaction with entries",
			logger.Fields{
				"transaction_id": id.String(),
				"error":          err.Error(),
			})
		return nil, fmt.Errorf("failed to get transaction with entries: %w", err)
	}

	logger.DebugContext(ctx, "Transaction with entries retrieved successfully",
		logger.Fields{
			"transaction_id": id.String(),
			"entries_count":  len(transactionWithEntries.Entries),
		})

	return transactionWithEntries, nil
}
*/

func (s *transactionService) ValidateTransaction(ctx context.Context, transaction *domain.Transaction, entries []*domain.TransactionEntry) ([]domain.ValidationError, error) {
	var errors []domain.ValidationError

	if len(entries) == 0 {
		errors = append(errors, domain.ValidationError{
			Field:   "entries",
			Message: "Transaction must have at least one entry",
			Code:    "MISSING_ENTRIES",
		})
		return errors, nil
	}

	var totalDebits, totalCredits decimal.Decimal
	accountMap := make(map[uuid.UUID]bool)

	for _, entry := range entries {
		if validationErrs := entry.Validate(); len(validationErrs) > 0 {
			errors = append(errors, validationErrs...)
		}

		totalDebits = totalDebits.Add(entry.DebitAmount)
		totalCredits = totalCredits.Add(entry.CreditAmount)
		accountMap[entry.AccountID] = true

		account, err := s.accountRepo.GetByID(ctx, entry.AccountID)
		if err != nil {
			errors = append(errors, domain.ValidationError{
				Field:   fmt.Sprintf("entry.%d.account_id", entry.EntryNumber),
				Message: "Account not found",
				Code:    "INVALID_ACCOUNT",
			})
			continue
		}

		if businessRuleErrs := entry.ValidateBusinessRules(account); len(businessRuleErrs) > 0 {
			errors = append(errors, businessRuleErrs...)
		}

		if currencyErrs := entry.ValidateAmountConsistency(); len(currencyErrs) > 0 {
			errors = append(errors, currencyErrs...)
		}
	}

	if !totalDebits.Equal(totalCredits) {
		errors = append(errors, domain.ValidationError{
			Field:   "transaction",
			Message: fmt.Sprintf("Transaction is not balanced - debits: %s, credits: %s", totalDebits.String(), totalCredits.String()),
			Code:    "UNBALANCED_TRANSACTION",
		})
	}

	if totalDebits.Equal(decimal.Zero) {
		errors = append(errors, domain.ValidationError{
			Field:   "transaction",
			Message: "Transaction cannot have zero amounts",
			Code:    "ZERO_AMOUNT_TRANSACTION",
		})
	}

	if len(accountMap) < 2 {
		errors = append(errors, domain.ValidationError{
			Field:   "transaction",
			Message: "Transaction must affect at least two accounts",
			Code:    "SINGLE_ACCOUNT_TRANSACTION",
		})
	}

	return errors, nil
}

func (s *transactionService) SearchTransactions(ctx context.Context, query string, limit int, offset int) ([]*domain.Transaction, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_service.search_transactions",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("search.query", query),
			attribute.Int("search.limit", limit),
			attribute.Int("search.offset", offset),
		))
	defer span.End()

	logger.DebugContext(ctx, "Searching transactions",
		logger.Fields{
			"query":  query,
			"limit":  limit,
			"offset": offset,
		})

	if limit <= 0 {
		limit = 50
	}

	if limit > 200 {
		limit = 200
	}

	transactions, err := s.repo.Search(ctx, query, limit, offset)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to search transactions",
			logger.Fields{
				"query": query,
				"error": err.Error(),
			})
		return nil, fmt.Errorf("failed to search transactions: %w", err)
	}

	logger.DebugContext(ctx, "Transaction search completed successfully",
		logger.Fields{
			"query":         query,
			"results_count": len(transactions),
		})

	return transactions, nil
}

func (s *transactionService) GetTransactionSummary(ctx context.Context, startDate, endDate time.Time) (*domain.TransactionSummary, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_service.get_transaction_summary",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("date_range.start", startDate.Format("2006-01-02")),
			attribute.String("date_range.end", endDate.Format("2006-01-02")),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting transaction summary",
		logger.Fields{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		})

	summary, err := s.repo.GetSummary(ctx, startDate, endDate)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get transaction summary",
			logger.Fields{
				"start_date": startDate.Format("2006-01-02"),
				"end_date":   endDate.Format("2006-01-02"),
				"error":      err.Error(),
			})
		return nil, fmt.Errorf("failed to get transaction summary: %w", err)
	}

	logger.DebugContext(ctx, "Transaction summary retrieved successfully",
		logger.Fields{
			"start_date":         startDate.Format("2006-01-02"),
			"end_date":           endDate.Format("2006-01-02"),
			"transaction_count":  summary.TotalTransactions,
			"total_debit_amount": summary.TotalDebitAmount.String(),
		})

	return summary, nil
}

func (s *transactionService) GetPendingApprovalTransactions(ctx context.Context, limit int, offset int) ([]*domain.Transaction, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_service.get_pending_approval_transactions",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.Int("limit", limit),
			attribute.Int("offset", offset),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting pending approval transactions",
		logger.Fields{
			"limit":  limit,
			"offset": offset,
		})

	transactions, err := s.repo.GetPendingApproval(ctx, limit, offset)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get pending approval transactions",
			logger.Fields{"error": err.Error()})
		return nil, fmt.Errorf("failed to get pending approval transactions: %w", err)
	}

	logger.DebugContext(ctx, "Pending approval transactions retrieved successfully",
		logger.Fields{"count": len(transactions)})

	return transactions, nil
}

func (s *transactionService) GetRecurringTransactionsDue(ctx context.Context, date time.Time) ([]*domain.Transaction, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_service.get_recurring_transactions_due",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("due_date", date.Format("2006-01-02")),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting recurring transactions due",
		logger.Fields{"due_date": date.Format("2006-01-02")})

	transactions, err := s.repo.GetRecurringDue(ctx, date)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get recurring transactions due",
			logger.Fields{
				"due_date": date.Format("2006-01-02"),
				"error":    err.Error(),
			})
		return nil, fmt.Errorf("failed to get recurring transactions due: %w", err)
	}

	logger.DebugContext(ctx, "Recurring transactions due retrieved successfully",
		logger.Fields{
			"due_date": date.Format("2006-01-02"),
			"count":    len(transactions),
		})

	return transactions, nil
}

func (s *transactionService) CreateRecurringTransaction(ctx context.Context, templateID uuid.UUID, date time.Time) (*domain.Transaction, error) {
	ctx, span := s.tracing.StartSpan(ctx, "transaction_service.create_recurring_transaction",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("template.id", templateID.String()),
			attribute.String("transaction.date", date.Format("2006-01-02")),
		))
	defer span.End()

	logger.InfoContext(ctx, "Creating recurring transaction",
		logger.Fields{
			"template_id":      templateID.String(),
			"transaction_date": date.Format("2006-01-02"),
		})

	template, err := s.repo.GetByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recurring transaction template: %w", err)
	}

	if !template.IsRecurring {
		return nil, domain.NewBusinessRuleError("not_recurring", "Template is not a recurring transaction")
	}

	entries, err := s.entryService.GetEntriesByTransactionID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template entries: %w", err)
	}

	newTransactionReq := &domain.CreateTransactionRequest{
		TransactionNumber:  fmt.Sprintf("%s-%s", template.TransactionNumber, date.Format("20060102")),
		TransactionType:    template.TransactionType,
		TransactionDate:    date,
		Description:        template.Description,
		ReferenceNumber:    template.ReferenceNumber,
		CurrencyCode:       template.CurrencyCode,
		ExchangeRate:       template.ExchangeRate,
		ApprovalRequired:   template.ApprovalRequired,
		SourceModule:       template.SourceModule,
		SourceDocumentType: template.SourceDocumentType,
	}

	newTransaction, err := s.repo.Create(ctx, newTransactionReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create recurring transaction: %w", err)
	}

	for i, entry := range entries {
		newEntry := entry.Clone()
		newEntry.TransactionID = newTransaction.ID
		newEntry.EntryNumber = int32(i + 1)

		if err := s.entryService.CreateEntry(ctx, newEntry); err != nil {
			return nil, fmt.Errorf("failed to create recurring entry: %w", err)
		}
	}

	if err := s.repo.UpdateNextRecurringDate(ctx, templateID, s.calculateNextRecurringDate(template, date)); err != nil {
		logger.WarnContext(ctx, "Failed to update next recurring date",
			logger.Fields{
				"template_id": templateID.String(),
				"error":       err.Error(),
			})
	}

	logger.InfoContext(ctx, "Recurring transaction created successfully",
		logger.Fields{
			"template_id":      templateID.String(),
			"transaction_id":   newTransaction.ID.String(),
			"transaction_date": date.Format("2006-01-02"),
		})

	return newTransaction, nil
}

func (s *transactionService) updateAccountBalances(ctx context.Context, entries []*domain.TransactionEntry) error {
	accountBalances := make(map[uuid.UUID]decimal.Decimal)

	for _, entry := range entries {
		currentBalance, exists := accountBalances[entry.AccountID]
		if !exists {
			currentBalance = decimal.Zero
		}

		if entry.IsDebit() {
			currentBalance = currentBalance.Add(entry.DebitAmount)
		} else {
			currentBalance = currentBalance.Sub(entry.CreditAmount)
		}

		accountBalances[entry.AccountID] = currentBalance
	}

	for accountID, balanceChange := range accountBalances {
		account, err := s.accountRepo.GetByID(ctx, accountID)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to get account for balance update",
				logger.Fields{
					"account_id": accountID.String(),
					"error":      err.Error(),
				})
			continue
		}

		newBalance := account.CurrentBalance.Add(balanceChange)
		newYTDBalance := account.YTDBalance.Add(balanceChange)

		balance := domain.AccountBalance{
			CurrentBalance:      newBalance,
			YTDBalance:          newYTDBalance,
			LastTransactionDate: &time.Time{},
		}

		if err := s.accountRepo.UpdateBalance(ctx, accountID, balance); err != nil {
			logger.ErrorContext(ctx, "Failed to update account balance",
				logger.Fields{
					"account_id": accountID.String(),
					"error":      err.Error(),
				})
		}
	}

	return nil
}

func (s *transactionService) calculateNextRecurringDate(template *domain.Transaction, currentDate time.Time) time.Time {
	if template.RecurringFrequency == nil {
		return currentDate.AddDate(0, 1, 0) // Default to monthly
	}

	switch *template.RecurringFrequency {
	case domain.RecurringFrequencyDaily:
		return currentDate.AddDate(0, 0, 1)
	case domain.RecurringFrequencyWeekly:
		return currentDate.AddDate(0, 0, 7)
	case domain.RecurringFrequencyBiweekly:
		return currentDate.AddDate(0, 0, 14)
	case domain.RecurringFrequencyMonthly:
		return currentDate.AddDate(0, 1, 0)
	case domain.RecurringFrequencyQuarterly:
		return currentDate.AddDate(0, 3, 0)
	case domain.RecurringFrequencyAnnually:
		return currentDate.AddDate(1, 0, 0)
	default:
		return currentDate.AddDate(0, 1, 0)
	}
}
