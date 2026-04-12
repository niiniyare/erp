package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/tracing"
)

type transactionRepository struct {
	store   db.Store
	tracing tracing.Service
}

func NewTransactionRepository(store db.Store, tracing tracing.Service) domain.TransactionRepository {
	return &transactionRepository{
		store:   store,
		tracing: tracing,
	}
}

// Basic CRUD Operations

func (r *transactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.Create")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Use tenant-aware transaction for proper isolation
	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Map domain transaction to SQLC parameters
		params := db.CreateTransactionParams{
			EntityID:              transaction.EntityID,
			TransactionNumber:     transaction.TransactionNumber,
			TransactionType:       mapDomainTransactionTypeToSQLCEnum(transaction.TransactionType),
			TransactionStatus:     mapDomainTransactionStatusToSQLCEnum(transaction.TransactionStatus),
			TransactionDate:       transaction.TransactionDate,
			PostingDate:           timePointerToTimeValue(transaction.PostingDate),
			DueDate:               timePointerToTimeValue(transaction.DueDate),
			Description:           transaction.Description,
			ReferenceNumber:       transaction.ReferenceNumber,
			ExternalReference:     transaction.ExternalReference,
			Memo:                  transaction.ReferenceNumber, // Use reference as memo if available
			CurrencyCode:          transaction.CurrencyCode,
			ExchangeRate:          decimalToPgNumeric(&transaction.ExchangeRate),
			TotalDebitAmount:      decimalToPgNumeric(&transaction.TotalDebitAmount),
			TotalCreditAmount:     decimalToPgNumeric(&transaction.TotalCreditAmount),
			SourceModule:          transaction.SourceModule,
			SourceDocumentType:    transaction.SourceDocumentType,
			SourceDocumentID:      transaction.SourceDocumentID,
			BatchID:               transaction.BatchID,
			ApprovalRequired:      boolToPtr(transaction.ApprovalRequired),
			ApprovalStatus:        mapDomainApprovalStatusToNullEnum(&transaction.ApprovalStatus),
			IsRecurring:           boolToPtr(transaction.IsRecurring),
			RecurringFrequency:    mapDomainRecurringFrequencyToNullEnum(transaction.RecurringFrequency),
			NextRecurringDate:     timePointerToTimeValue(transaction.NextRecurringDate),
			TransactionAttributes: mapAttributesToJSON(transaction.TransactionAttributes),
			AttachmentIds:         transaction.AttachmentIds,
			Tags:                  transaction.Tags,
			CreatedBy:             transaction.CreatedBy,
		}

		// Execute SQLC query within tenant context
		sqlcTransaction, err := s.CreateTransaction(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "create_transaction")
		}

		// Update the transaction with generated fields
		transaction.ID = sqlcTransaction.ID
		transaction.TenantID = sqlcTransaction.TenantID
		transaction.CreatedAt = sqlcTransaction.CreatedAt
		transaction.UpdatedAt = sqlcTransaction.UpdatedAt

		return nil
	})
}

func (r *transactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.GetByID")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var transaction *domain.Transaction
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		sqlcTransaction, err := s.GetTransactionByID(ctx, id)
		if err != nil {
			if err == db.ErrNoRows {
				return domain.ErrTransactionNotFound
			}
			return r.mapDatabaseError(err, "get_transaction_by_id")
		}

		transaction, err = r.mapSQLCTransactionToDomain(sqlcTransaction)
		return err
	})
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (r *transactionRepository) GetByNumber(ctx context.Context, entityID *uuid.UUID, transactionNumber string) (*domain.Transaction, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.GetByNumber")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var transaction *domain.Transaction
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		sqlcTransaction, err := s.GetTransactionByNumber(ctx, transactionNumber)
		if err != nil {
			if err == db.ErrNoRows {
				return domain.ErrTransactionNotFound
			}
			return r.mapDatabaseError(err, "get_transaction_by_number")
		}

		transaction, err = r.mapSQLCTransactionToDomain(sqlcTransaction)
		return err
	})
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (r *transactionRepository) Update(ctx context.Context, transaction *domain.Transaction) error {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.Update")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Map transaction to SQLC parameters
		params := db.UpdateTransactionParams{
			TransactionID:         transaction.ID,
			TransactionStatus:     mapDomainTransactionStatusToSQLCEnumPtr(transaction.TransactionStatus),
			PostingDate:           timePointerToTimeValue(transaction.PostingDate),
			DueDate:               timePointerToTimeValue(transaction.DueDate),
			Description:           &transaction.Description,
			ReferenceNumber:       transaction.ReferenceNumber,
			ExternalReference:     transaction.ExternalReference,
			Memo:                  transaction.ReferenceNumber, // Use reference as memo
			TotalDebitAmount:      decimalToPgNumeric(&transaction.TotalDebitAmount),
			TotalCreditAmount:     decimalToPgNumeric(&transaction.TotalCreditAmount),
			ApprovalStatus:        mapDomainApprovalStatusToNullEnum(&transaction.ApprovalStatus),
			ApprovedBy:            transaction.ApprovedBy,
			ApprovedAt:            timePointerToNullTime(transaction.ApprovedAt),
			ApprovalNotes:         transaction.ApprovalNotes,
			TransactionAttributes: mapAttributesToJSON(transaction.TransactionAttributes),
			AttachmentIds:         transaction.AttachmentIds,
			Tags:                  transaction.Tags,
			UpdatedBy:             transaction.UpdatedBy,
		}

		// Execute SQLC update query
		_, err := s.UpdateTransaction(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "update_transaction")
		}

		return nil
	})
}

func (r *transactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.Delete")
	defer span.End()

	// Get tenant and user ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	userID, _ := shared.GetUserID(ctx) // Optional for soft delete

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// First check if transaction exists and is deletable
		_, err := s.GetTransactionByID(ctx, id)
		if err != nil {
			if err == db.ErrNoRows {
				return domain.ErrTransactionNotFound
			}
			return r.mapDatabaseError(err, "get_transaction_for_delete")
		}

		var userIDPtr *uuid.UUID
		if userID != uuid.Nil {
			userIDPtr = &userID
		}

		err = s.SoftDeleteTransaction(ctx, db.SoftDeleteTransactionParams{
			TransactionID: id,
			UpdatedBy:     userIDPtr,
		})
		if err != nil {
			return r.mapDatabaseError(err, "soft_delete_transaction")
		}
		return nil
	})
}

// List and Filter Operations

func (r *transactionRepository) List(ctx context.Context, filter *domain.TransactionFilter) ([]*domain.Transaction, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.List")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var transactions []*domain.Transaction
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Map domain filter to SQLC parameters
		var limitCount, offsetCount int32
		if filter.Limit != nil {
			limitCount = int32(*filter.Limit)
		} else {
			limitCount = 50 // Default limit
		}
		if filter.Offset != nil {
			offsetCount = int32(*filter.Offset)
		}

		params := db.ListTransactionsParams{
			LimitCount:  limitCount,
			OffsetCount: offsetCount,
			// Initialize date fields with zero time to avoid filtering issues
			DateFrom: time.Time{},
			DateTo:   time.Time{},
		}

		// Map optional filters
		if filter.TransactionType != nil {
			typeStr := string(*filter.TransactionType)
			params.TransactionType = &typeStr
		}

		if filter.Status != nil {
			statusStr := string(*filter.Status)
			params.TransactionStatus = &statusStr
		}

		if filter.DateRange != nil {
			if !filter.DateRange.StartDate.IsZero() {
				params.DateFrom = filter.DateRange.StartDate
			} else {
				params.DateFrom = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC) // Very old date to include all transactions
			}
			if filter.DateRange.EndDate != nil {
				params.DateTo = *filter.DateRange.EndDate
			} else {
				params.DateTo = time.Date(2100, 12, 31, 23, 59, 59, 0, time.UTC) // Far future date to include all transactions
			}
		} else {
			// No date filtering - use very wide date range
			params.DateFrom = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
			params.DateTo = time.Date(2100, 12, 31, 23, 59, 59, 0, time.UTC)
		}

		// Execute SQLC query
		sqlcTransactions, err := s.ListTransactions(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "list_transactions")
		}

		// Map results to domain models
		transactions = make([]*domain.Transaction, 0, len(sqlcTransactions))
		for _, sqlcTransaction := range sqlcTransactions {
			transaction, err := r.mapSQLCTransactionToDomain(sqlcTransaction)
			if err != nil {
				return fmt.Errorf("failed to map SQLC transaction to domain: %w", err)
			}
			transactions = append(transactions, transaction)
		}

		return nil
	})

	return transactions, err
}

func (r *transactionRepository) Count(ctx context.Context, filter *domain.TransactionFilter) (int64, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.Count")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return 0, fmt.Errorf("tenant ID not found in context")
	}

	var count int64
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Map basic filter parameters for count
		params := db.CountTransactionsParams{}

		// Map optional filters
		if filter.TransactionType != nil {
			typeStr := string(*filter.TransactionType)
			params.TransactionType = &typeStr
		}

		if filter.Status != nil {
			statusStr := string(*filter.Status)
			params.TransactionStatus = &statusStr
		}

		if filter.DateRange != nil {
			if !filter.DateRange.StartDate.IsZero() {
				params.DateFrom = filter.DateRange.StartDate
			}
			if filter.DateRange.EndDate != nil {
				params.DateTo = *filter.DateRange.EndDate
			}
		}

		result, err := s.CountTransactions(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "count_transactions")
		}
		count = result
		return nil
	})

	return count, err
}

// Status Operations - these are critical business operations requiring proper transaction handling

func (r *transactionRepository) PostTransaction(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.Post")
	defer span.End()

	// Get tenant and user ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	userID, ok := shared.GetUserID(ctx)
	if !ok {
		return fmt.Errorf("user ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		_, err := s.PostTransaction(ctx, db.PostTransactionParams{
			TransactionID: id,
			PostingDate:   time.Now(),
			PostedBy:      &userID,
		})
		if err != nil {
			return r.mapDatabaseError(err, "post_transaction")
		}
		return nil
	})
}

func (r *transactionRepository) ApproveTransaction(ctx context.Context, transactionID uuid.UUID, approvedBy uuid.UUID, approvedAt time.Time, notes *string) error {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.Approve")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		notesStr := ""
		if notes != nil {
			notesStr = *notes
		}
		_, err := s.ApproveTransaction(ctx, db.ApproveTransactionParams{
			TransactionID: transactionID,
			ApprovedBy:    &approvedBy,
			ApprovalNotes: &notesStr,
		})
		if err != nil {
			return r.mapDatabaseError(err, "approve_transaction")
		}
		return nil
	})
}

func (r *transactionRepository) RejectTransaction(ctx context.Context, id uuid.UUID, notes *string) error {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.Reject")
	defer span.End()

	// Get tenant and user ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	userID, ok := shared.GetUserID(ctx)
	if !ok {
		return fmt.Errorf("user ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		notesStr := ""
		if notes != nil {
			notesStr = *notes
		}
		_, err := s.RejectTransaction(ctx, db.RejectTransactionParams{
			TransactionID: id,
			ApprovedBy:    &userID,
			ApprovalNotes: &notesStr,
		})
		if err != nil {
			return r.mapDatabaseError(err, "reject_transaction")
		}
		return nil
	})
}

func (r *transactionRepository) ReverseTransaction(ctx context.Context, id uuid.UUID, reversalTransactionID uuid.UUID, reason string) error {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.Reverse")
	defer span.End()

	// Get tenant and user ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	userID, ok := shared.GetUserID(ctx)
	if !ok {
		return fmt.Errorf("user ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		_, err := s.ReverseTransaction(ctx, db.ReverseTransactionParams{
			TransactionID:           id,
			ReversedByTransactionID: &reversalTransactionID,
			ReversalReason:          &reason,
			UpdatedBy:               &userID,
		})
		if err != nil {
			return r.mapDatabaseError(err, "reverse_transaction")
		}
		return nil
	})
}

// Validation and Balance Operations

func (r *transactionRepository) ValidateBalance(ctx context.Context, id uuid.UUID) (bool, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.ValidateBalance")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return false, fmt.Errorf("tenant ID not found in context")
	}

	var isBalanced bool
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		result, err := s.ValidateTransactionBalance(ctx, id)
		if err != nil {
			return r.mapDatabaseError(err, "validate_transaction_balance")
		}
		isBalanced = result.IsBalanced
		return nil
	})

	return isBalanced, err
}

func (r *transactionRepository) GetWithEntries(ctx context.Context, id uuid.UUID) (*domain.TransactionWithEntries, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.GetWithEntries")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var transactionWithEntries *domain.TransactionWithEntries
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Execute SQLC query to get transaction with entries
		results, err := s.GetTransactionWithEntries(ctx, id)
		if err != nil {
			return r.mapDatabaseError(err, "get_transaction_with_entries")
		}

		if len(results) == 0 {
			return domain.ErrTransactionNotFound
		}

		// Map first result to transaction
		firstResult := results[0]
		transaction, err := r.mapSQLCTransactionRowToDomain(firstResult)
		if err != nil {
			return err
		}

		// Map entries
		entries := make([]*domain.TransactionEntry, 0)
		for _, result := range results {
			if result.EntryID != nil {
				entry := &domain.TransactionEntry{
					ID:            *result.EntryID,
					TransactionID: transaction.ID,
					EntryNumber:   getInt32Value(result.EntryNumber),
					AccountID:     getUUIDValue(result.AccountID),
					DebitAmount:   pgNumericToDecimal(result.DebitAmount),
					CreditAmount:  pgNumericToDecimal(result.CreditAmount),
					Description:   ptrStringToString(result.EntryDescription),
					Reference:     result.EntryReference,
					CostCenter:    result.CostCenter,
					Department:    result.Department,
					ProjectID:     result.ProjectID,
				}
				entries = append(entries, entry)
			}
		}

		// Convert []*TransactionEntry to []TransactionEntry
		entriesSlice := make([]domain.TransactionEntry, 0, len(entries))
		for _, entry := range entries {
			entriesSlice = append(entriesSlice, *entry)
		}

		transactionWithEntries = &domain.TransactionWithEntries{
			Transaction: *transaction,
			Entries:     entriesSlice,
		}
		return nil
	})

	return transactionWithEntries, err
}

// Specialized Queries

func (r *transactionRepository) GetByBatch(ctx context.Context, batchID uuid.UUID) ([]*domain.Transaction, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.GetByBatch")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var transactions []*domain.Transaction
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		sqlcTransactions, err := s.GetTransactionsByBatch(ctx, &batchID)
		if err != nil {
			return r.mapDatabaseError(err, "get_transactions_by_batch")
		}

		transactions = make([]*domain.Transaction, 0, len(sqlcTransactions))
		for _, sqlcTransaction := range sqlcTransactions {
			transaction, err := r.mapSQLCTransactionToDomain(sqlcTransaction)
			if err != nil {
				return fmt.Errorf("failed to map SQLC transaction to domain: %w", err)
			}
			transactions = append(transactions, transaction)
		}
		return nil
	})

	return transactions, err
}

func (r *transactionRepository) GetPendingApprovalTransactions(ctx context.Context, limit, offset int32) ([]*domain.Transaction, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.GetPendingApproval")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var transactions []*domain.Transaction
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		sqlcTransactions, err := s.GetPendingApprovalTransactions(ctx, db.GetPendingApprovalTransactionsParams{
			LimitCount:  limit,
			OffsetCount: offset,
		})
		if err != nil {
			return r.mapDatabaseError(err, "get_pending_approval_transactions")
		}

		transactions = make([]*domain.Transaction, 0, len(sqlcTransactions))
		for _, sqlcTransaction := range sqlcTransactions {
			transaction, err := r.mapSQLCTransactionToDomain(sqlcTransaction)
			if err != nil {
				return fmt.Errorf("failed to map SQLC transaction to domain: %w", err)
			}
			transactions = append(transactions, transaction)
		}
		return nil
	})

	return transactions, err
}

func (r *transactionRepository) Search(ctx context.Context, query string, limit, offset int32) ([]*domain.Transaction, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.Search")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var transactions []*domain.Transaction
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		sqlcTransactions, err := s.SearchTransactions(ctx, db.SearchTransactionsParams{
			SearchTerm:  &query,
			LimitCount:  limit,
			OffsetCount: offset,
		})
		if err != nil {
			return r.mapDatabaseError(err, "search_transactions")
		}

		transactions = make([]*domain.Transaction, 0, len(sqlcTransactions))
		for _, sqlcTransaction := range sqlcTransactions {
			transaction, err := r.mapSQLCTransactionToDomain(sqlcTransaction)
			if err != nil {
				return fmt.Errorf("failed to map SQLC transaction to domain: %w", err)
			}
			transactions = append(transactions, transaction)
		}
		return nil
	})

	return transactions, err
}

// Error mapping helper
func (r *transactionRepository) mapDatabaseError(err error, operation string) error {
	// Convert database-specific errors to domain errors
	if err == db.ErrNoRows {
		return domain.ErrTransactionNotFound
	}

	// Check for constraint violations
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "23505": // unique violation
			return domain.ErrTransactionNumberExists
		case "23503": // foreign key violation
			// Include more details about the constraint violation for debugging
			return fmt.Errorf("foreign key constraint violation: %s (detail: %s, constraint: %s)", pgErr.Message, pgErr.Detail, pgErr.ConstraintName)
		}
	}

	// Default to internal server error
	return fmt.Errorf("database operation failed: %s: %w", operation, err)
}

// Mapping helper functions
func (r *transactionRepository) mapSQLCTransactionToDomain(sqlcTransaction *db.FinanceTransaction) (*domain.Transaction, error) {
	// Parse metadata from JSONB
	var metadata map[string]any
	if len(sqlcTransaction.TransactionAttributes) > 0 {
		if err := json.Unmarshal(sqlcTransaction.TransactionAttributes, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal transaction metadata: %w", err)
		}
	}

	return &domain.Transaction{
		ID:                    sqlcTransaction.ID,
		TenantID:              sqlcTransaction.TenantID,
		EntityID:              sqlcTransaction.EntityID,
		TransactionNumber:     sqlcTransaction.TransactionNumber,
		TransactionType:       mapSQLCTransactionTypeToDomain(sqlcTransaction.TransactionType),
		TransactionStatus:     mapSQLCTransactionStatusToDomain(sqlcTransaction.TransactionStatus),
		TransactionDate:       sqlcTransaction.TransactionDate,
		PostingDate:           &sqlcTransaction.PostingDate,
		DueDate:               &sqlcTransaction.DueDate,
		Description:           sqlcTransaction.Description,
		ReferenceNumber:       sqlcTransaction.ReferenceNumber,
		ExternalReference:     sqlcTransaction.ExternalReference,
		CurrencyCode:          sqlcTransaction.CurrencyCode,
		ExchangeRate:          getDecimalValue(pgNumericToDecimalPtr(sqlcTransaction.ExchangeRate)),
		TotalDebitAmount:      pgNumericToDecimal(sqlcTransaction.TotalDebitAmount),
		TotalCreditAmount:     pgNumericToDecimal(sqlcTransaction.TotalCreditAmount),
		SourceModule:          sqlcTransaction.SourceModule,
		SourceDocumentType:    sqlcTransaction.SourceDocumentType,
		SourceDocumentID:      sqlcTransaction.SourceDocumentID,
		BatchID:               sqlcTransaction.BatchID,
		ApprovalRequired:      getBoolValue(sqlcTransaction.ApprovalRequired),
		ApprovalStatus:        mapNullApprovalStatusToDomain(sqlcTransaction.ApprovalStatus),
		ApprovedBy:            sqlcTransaction.ApprovedBy,
		ApprovedAt:            nullTimeToPointer(sqlcTransaction.ApprovedAt),
		ApprovalNotes:         sqlcTransaction.ApprovalNotes,
		IsRecurring:           getBoolValue(sqlcTransaction.IsRecurring),
		RecurringFrequency:    mapNullRecurringFrequencyToDomainString(sqlcTransaction.RecurringFrequency),
		NextRecurringDate:     &sqlcTransaction.NextRecurringDate,
		TransactionAttributes: metadata,
		CreatedAt:             sqlcTransaction.CreatedAt,
		UpdatedAt:             sqlcTransaction.UpdatedAt,
		CreatedBy:             sqlcTransaction.CreatedBy,
		UpdatedBy:             sqlcTransaction.UpdatedBy,
	}, nil
}

func (r *transactionRepository) mapSQLCTransactionRowToDomain(row *db.GetTransactionWithEntriesRow) (*domain.Transaction, error) {
	// Parse metadata from JSONB
	var metadata map[string]any
	if len(row.TransactionAttributes) > 0 {
		if err := json.Unmarshal(row.TransactionAttributes, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal transaction metadata: %w", err)
		}
	}

	return &domain.Transaction{
		ID:                    row.ID,
		TenantID:              row.TenantID,
		EntityID:              row.EntityID,
		TransactionNumber:     row.TransactionNumber,
		TransactionType:       mapSQLCTransactionTypeToDomain(row.TransactionType),
		TransactionStatus:     mapSQLCTransactionStatusToDomain(row.TransactionStatus),
		TransactionDate:       row.TransactionDate,
		PostingDate:           &row.PostingDate,
		DueDate:               &row.DueDate,
		Description:           row.Description,
		ReferenceNumber:       row.ReferenceNumber,
		ExternalReference:     row.ExternalReference,
		CurrencyCode:          row.CurrencyCode,
		ExchangeRate:          getDecimalValue(pgNumericToDecimalPtr(row.ExchangeRate)),
		TotalDebitAmount:      pgNumericToDecimal(row.TotalDebitAmount),
		TotalCreditAmount:     pgNumericToDecimal(row.TotalCreditAmount),
		SourceModule:          row.SourceModule,
		SourceDocumentType:    row.SourceDocumentType,
		SourceDocumentID:      row.SourceDocumentID,
		BatchID:               row.BatchID,
		ApprovalRequired:      getBoolValue(row.ApprovalRequired),
		ApprovalStatus:        mapNullApprovalStatusToDomain(row.ApprovalStatus),
		ApprovedBy:            row.ApprovedBy,
		ApprovedAt:            nullTimeToPointer(row.ApprovedAt),
		ApprovalNotes:         row.ApprovalNotes,
		IsRecurring:           getBoolValue(row.IsRecurring),
		RecurringFrequency:    mapNullRecurringFrequencyToDomainString(row.RecurringFrequency),
		NextRecurringDate:     &row.NextRecurringDate,
		TransactionAttributes: metadata,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
		CreatedBy:             row.CreatedBy,
		UpdatedBy:             row.UpdatedBy,
	}, nil
}

// CalculateAccountBalance calculates the balance for an account as of a specific date
func (r *transactionRepository) CalculateAccountBalance(ctx context.Context, accountID uuid.UUID, asOfDate *time.Time) (decimal.Decimal, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.CalculateAccountBalance")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return decimal.Zero, fmt.Errorf("tenant ID not found in context")
	}

	date := time.Now()
	if asOfDate != nil {
		date = *asOfDate
	}

	var balance decimal.Decimal
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		row, err := s.GetAccountTransactionBalance(ctx, db.GetAccountTransactionBalanceParams{
			AccountID: accountID,
			AsOfDate:  date,
		})
		if err != nil {
			return fmt.Errorf("calculate balance: %w", err)
		}
		// GetAccountTransactionBalanceRow returns int64 SUM values (scaled integers from NUMERIC)
		balance = decimal.NewFromInt(row.NetBalance)
		return nil
	})

	return balance, err
}

// ── Entry mapper ──────────────────────────────────────────────────────────────

func mapSQLCEntryToDomain(e *db.FinanceTransactionEntry) domain.TransactionEntry {
	var reconciledDate *time.Time
	if !e.ReconciledDate.IsZero() {
		reconciledDate = &e.ReconciledDate
	}
	var deletedAt *time.Time
	if e.DeletedAt.Valid {
		deletedAt = &e.DeletedAt.Time
	}
	return domain.TransactionEntry{
		ID:                      e.ID,
		TenantID:                e.TenantID,
		TransactionID:           e.TransactionID,
		EntryNumber:             e.EntryNumber,
		AccountID:               e.AccountID,
		DebitAmount:             pgNumericToDecimal(e.DebitAmount),
		CreditAmount:            pgNumericToDecimal(e.CreditAmount),
		Description:             e.Description,
		Reference:               e.Reference,
		CostCenter:              e.CostCenter,
		Department:              e.Department,
		ProjectID:               e.ProjectID,
		OriginalCurrency:        e.OriginalCurrency,
		OriginalAmount:          pgNumericToDecimal(e.OriginalAmount),
		ExchangeRate:            pgNumericToDecimal(e.ExchangeRate),
		TaxCode:                 e.TaxCode,
		TaxRate:                 pgNumericToDecimal(e.TaxRate),
		TaxAmount:               pgNumericToDecimal(e.TaxAmount),
		Reconciled:              getBoolValue(e.Reconciled),
		ReconciledDate:          reconciledDate,
		ReconciliationReference: e.ReconciliationReference,
		CreatedAt:               e.CreatedAt,
		UpdatedAt:               e.UpdatedAt,
		DeletedAt:               deletedAt,
	}
}

// ── Entry CRUD ────────────────────────────────────────────────────────────────

func (r *transactionRepository) CreateEntry(ctx context.Context, entry *domain.TransactionEntry) error {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.CreateEntry")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		_, err := s.CreateTransactionEntry(ctx, db.CreateTransactionEntryParams{
			TransactionID:    entry.TransactionID,
			EntryNumber:      entry.EntryNumber,
			AccountID:        entry.AccountID,
			DebitAmount:      decimalToPgNumeric(&entry.DebitAmount),
			CreditAmount:     decimalToPgNumeric(&entry.CreditAmount),
			Description:      entry.Description,
			Reference:        entry.Reference,
			CostCenter:       entry.CostCenter,
			Department:       entry.Department,
			ProjectID:        entry.ProjectID,
			OriginalCurrency: entry.OriginalCurrency,
			OriginalAmount:   decimalToPgNumeric(&entry.OriginalAmount),
			ExchangeRate:     decimalToPgNumeric(&entry.ExchangeRate),
			TaxCode:          entry.TaxCode,
			TaxRate:          decimalToPgNumeric(&entry.TaxRate),
			TaxAmount:        decimalToPgNumeric(&entry.TaxAmount),
		})
		return r.mapDatabaseError(err, "create_entry")
	})
}

func (r *transactionRepository) CreateEntries(ctx context.Context, entries []*domain.TransactionEntry) error {
	for _, e := range entries {
		if err := r.CreateEntry(ctx, e); err != nil {
			return fmt.Errorf("CreateEntries: entry %d: %w", e.EntryNumber, err)
		}
	}
	return nil
}

func (r *transactionRepository) GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.TransactionEntry, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.GetEntryByID")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var entry *domain.TransactionEntry
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		row, err := s.GetTransactionEntryByID(ctx, id)
		if err != nil {
			return r.mapDatabaseError(err, "get_entry_by_id")
		}
		e := mapSQLCEntryToDomain(row)
		entry = &e
		return nil
	})
	return entry, err
}

func (r *transactionRepository) GetEntriesByTransaction(ctx context.Context, transactionID uuid.UUID) ([]domain.TransactionEntry, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.GetEntriesByTransaction")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var entries []domain.TransactionEntry
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		rows, err := s.ListTransactionEntries(ctx, transactionID)
		if err != nil {
			return r.mapDatabaseError(err, "get_entries_by_transaction")
		}
		entries = make([]domain.TransactionEntry, 0, len(rows))
		for _, row := range rows {
			entries = append(entries, mapSQLCEntryToDomain(row))
		}
		return nil
	})
	return entries, err
}

func (r *transactionRepository) GetEntriesByAccount(ctx context.Context, accountID uuid.UUID, filter *domain.EntryFilter) ([]domain.TransactionEntry, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.GetEntriesByAccount")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	limit := int32(100)
	offset := int32(0)
	from := time.Time{}
	to := time.Now()
	if filter != nil {
		if filter.Limit != nil {
			limit = int32(*filter.Limit)
		}
		if filter.Offset != nil {
			offset = int32(*filter.Offset)
		}
		if filter.DateRange != nil {
			from = filter.DateRange.StartDate
			if filter.DateRange.EndDate != nil {
				to = *filter.DateRange.EndDate
			}
		}
	}

	var entries []domain.TransactionEntry
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		rows, err := s.GetAccountEntries(ctx, db.GetAccountEntriesParams{
			AccountID:         accountID,
			DateFrom:          from,
			DateTo:            to,
			TransactionStatus: string(domain.TransactionStatusPosted),
			Offset:            offset,
			Limit:             limit,
		})
		if err != nil {
			return r.mapDatabaseError(err, "get_entries_by_account")
		}
		entries = make([]domain.TransactionEntry, 0, len(rows))
		for _, row := range rows {
			entries = append(entries, domain.TransactionEntry{
				ID:            row.ID,
				TransactionID: row.TransactionID,
				EntryNumber:   row.EntryNumber,
				AccountID:     row.AccountID,
				DebitAmount:   pgNumericToDecimal(row.DebitAmount),
				CreditAmount:  pgNumericToDecimal(row.CreditAmount),
				Description:   row.Description,
				Reference:     row.Reference,
				CostCenter:    row.CostCenter,
				Department:    row.Department,
				ProjectID:     row.ProjectID,
				CreatedAt:     row.CreatedAt,
			})
		}
		return nil
	})
	return entries, err
}

func (r *transactionRepository) UpdateEntry(ctx context.Context, entry *domain.TransactionEntry) error {
	// Entry lines are immutable after posting per double-entry accounting rules.
	// Corrections are made by reversing and reposting the parent transaction.
	return fmt.Errorf("transaction entries are immutable — reverse and repost to correct")
}

func (r *transactionRepository) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.DeleteEntry")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		return r.mapDatabaseError(s.DeleteTransactionEntry(ctx, id), "delete_entry")
	})
}

func (r *transactionRepository) SearchEntries(ctx context.Context, query string, filters *domain.EntryFilter, limit int, offset int) ([]*domain.TransactionEntry, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.SearchEntries")
	defer span.End()

	// No dedicated full-text SQLC query for entries exists yet.
	// Fall back: if a specific account is requested, delegate to GetEntriesByAccount
	// and post-filter by description/reference containing the query string.
	if filters != nil && filters.AccountID != nil {
		lim := limit
		off := offset
		f := &domain.EntryFilter{
			AccountID: filters.AccountID,
			DateRange: filters.DateRange,
			Limit:     &lim,
			Offset:    &off,
		}
		all, err := r.GetEntriesByAccount(ctx, *filters.AccountID, f)
		if err != nil {
			return nil, err
		}
		q := strings.ToLower(query)
		out := make([]*domain.TransactionEntry, 0)
		for i := range all {
			e := all[i]
			if strings.Contains(strings.ToLower(e.Description), q) ||
				(e.Reference != nil && strings.Contains(strings.ToLower(*e.Reference), q)) {
				out = append(out, &e)
			}
		}
		return out, nil
	}
	// Without an account filter a global search is not supported without a dedicated index.
	return nil, fmt.Errorf("SearchEntries: provide an account_id filter for entry search (global full-text search requires a dedicated SQLC query)")
}

func (r *transactionRepository) UpdateReconciliationStatus(ctx context.Context, entryID uuid.UUID, reconciled bool, reconciledDate *time.Time, reconciliationRef *string) error {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.UpdateReconciliationStatus")
	defer span.End()

	if !reconciled {
		// Unreconcile path: no direct SQLC query — tracked in TASK-029.
		return fmt.Errorf("unreconcile not yet implemented (TASK-029)")
	}

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	date := time.Now()
	if reconciledDate != nil {
		date = *reconciledDate
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		return r.mapDatabaseError(s.MarkEntriesReconciled(ctx, db.MarkEntriesReconciledParams{
			ReconciledDate:          date,
			ReconciliationReference: reconciliationRef,
			EntryIds:                []uuid.UUID{entryID},
		}), "mark_reconciled")
	})
}

func (r *transactionRepository) GetUnreconciledEntries(ctx context.Context, accountID uuid.UUID, cutoffDate *time.Time) ([]*domain.TransactionEntry, error) {
	filter := &domain.EntryFilter{}
	if cutoffDate != nil {
		filter.DateRange = &domain.DateRange{EndDate: cutoffDate}
	}
	entries, err := r.GetEntriesByAccount(ctx, accountID, filter)
	if err != nil {
		return nil, err
	}
	var unreconciled []*domain.TransactionEntry
	for i := range entries {
		if !entries[i].Reconciled {
			e := entries[i]
			unreconciled = append(unreconciled, &e)
		}
	}
	return unreconciled, nil
}

func (r *transactionRepository) GetEntrySummary(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) (*domain.TransactionSummary, error) {
	entries, err := r.GetEntriesByAccount(ctx, accountID, &domain.EntryFilter{
		DateRange: &domain.DateRange{StartDate: startDate, EndDate: &endDate},
	})
	if err != nil {
		return nil, err
	}
	var totalDebit, totalCredit decimal.Decimal
	for _, e := range entries {
		totalDebit = totalDebit.Add(e.DebitAmount)
		totalCredit = totalCredit.Add(e.CreditAmount)
	}
	return &domain.TransactionSummary{
		TotalDebit:       totalDebit,
		TotalCredit:      totalCredit,
		TransactionCount: int64(len(entries)),
	}, nil
}

func (r *transactionRepository) GetAccountTransactionSummary(ctx context.Context, accountID uuid.UUID, dateRange *domain.DateRange) (*domain.TransactionSummary, error) {
	var start, end time.Time
	end = time.Now()
	if dateRange != nil {
		start = dateRange.StartDate
		if dateRange.EndDate != nil {
			end = *dateRange.EndDate
		}
	}
	return r.GetEntrySummary(ctx, accountID, start, end)
}

// ── Status / listing helpers ──────────────────────────────────────────────────

func (r *transactionRepository) GetPendingApproval(ctx context.Context, userID *uuid.UUID) ([]*domain.Transaction, error) {
	return r.GetPendingApprovalTransactions(ctx, 50, 0)
}

func (r *transactionRepository) GetByStatus(ctx context.Context, status domain.TransactionStatus, limit int) ([]*domain.Transaction, error) {
	filter := &domain.TransactionFilter{
		Status: &status,
		Limit:  &limit,
	}
	offset := 0
	filter.Offset = &offset
	return r.List(ctx, filter)
}

func (r *transactionRepository) ListByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*domain.Transaction, error) {
	limit := 1000
	offset := 0
	filter := &domain.TransactionFilter{
		DateRange: &domain.DateRange{StartDate: startDate, EndDate: &endDate},
		Limit:     &limit,
		Offset:    &offset,
	}
	return r.List(ctx, filter)
}

func (r *transactionRepository) ListByAccount(ctx context.Context, accountID uuid.UUID, filter *domain.TransactionFilter) ([]*domain.Transaction, error) {
	if filter == nil {
		filter = &domain.TransactionFilter{}
	}
	filter.AccountID = &accountID
	limit := 100
	offset := 0
	if filter.Limit == nil {
		filter.Limit = &limit
	}
	if filter.Offset == nil {
		filter.Offset = &offset
	}
	return r.List(ctx, filter)
}

func (r *transactionRepository) GetRecurringTransactions(ctx context.Context, dueDate time.Time) ([]*domain.Transaction, error) {
	limit := 500
	offset := 0
	filter := &domain.TransactionFilter{
		DateRange: &domain.DateRange{EndDate: &dueDate},
		Limit:     &limit,
		Offset:    &offset,
	}
	all, err := r.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	// Filter to recurring transactions only (List has no IsRecurring param)
	recurring := make([]*domain.Transaction, 0)
	for _, t := range all {
		if t.IsRecurring {
			recurring = append(recurring, t)
		}
	}
	return recurring, nil
}

// ── Post / Approve / Reject / Reverse delegators ─────────────────────────────

func (r *transactionRepository) Post(ctx context.Context, transactionID uuid.UUID, postedBy uuid.UUID, postedAt time.Time) error {
	return r.PostTransaction(ctx, transactionID)
}

func (r *transactionRepository) Approve(ctx context.Context, transactionID uuid.UUID, approvedBy uuid.UUID, approvedAt time.Time, notes *string) error {
	return r.ApproveTransaction(ctx, transactionID, approvedBy, approvedAt, notes)
}

func (r *transactionRepository) Reject(ctx context.Context, transactionID uuid.UUID, rejectedBy uuid.UUID, rejectedAt time.Time, reason domain.RejectionReason, notes *string) error {
	return r.RejectTransaction(ctx, transactionID, notes)
}

func (r *transactionRepository) Reverse(ctx context.Context, originalID, reversalID uuid.UUID, reason string) error {
	return r.ReverseTransaction(ctx, originalID, reversalID, reason)
}

// ── Uniqueness / validation ───────────────────────────────────────────────────

func (r *transactionRepository) IsTransactionNumberUnique(ctx context.Context, entityID *uuid.UUID, transactionNumber string, excludeID *uuid.UUID) (bool, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.IsTransactionNumberUnique")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return false, fmt.Errorf("tenant ID not found in context")
	}

	var unique bool
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		existing, err := s.GetTransactionByNumber(ctx, transactionNumber)
		if err != nil {
			// Not found → unique
			unique = true
			return nil
		}
		// Found → unique only if it's the excluded record (update case)
		unique = excludeID != nil && existing.ID == *excludeID
		return nil
	})
	return unique, err
}

func (r *transactionRepository) ValidateTransaction(ctx context.Context, transaction *domain.Transaction) ([]domain.ValidationError, error) {
	return transaction.Validate(), nil
}

func (r *transactionRepository) GetTransactionSummary(ctx context.Context, filter *domain.TransactionFilter) ([]*domain.TransactionSummary, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.GetTransactionSummary")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	dateFrom := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Now()
	if filter != nil && filter.DateRange != nil {
		if !filter.DateRange.StartDate.IsZero() {
			dateFrom = filter.DateRange.StartDate
		}
		if filter.DateRange.EndDate != nil {
			dateTo = *filter.DateRange.EndDate
		}
	}

	var summaries []*domain.TransactionSummary
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		rows, err := s.GetTransactionSummaryByPeriod(ctx, db.GetTransactionSummaryByPeriodParams{
			DateFrom: dateFrom,
			DateTo:   dateTo,
		})
		if err != nil {
			return r.mapDatabaseError(err, "get_transaction_summary")
		}
		summaries = make([]*domain.TransactionSummary, 0, len(rows))
		for _, row := range rows {
			summaries = append(summaries, &domain.TransactionSummary{
				TransactionType:   domain.TransactionType(row.TransactionType),
				TransactionStatus: domain.TransactionStatus(row.TransactionStatus),
				TransactionCount:  row.TransactionCount,
				TotalDebit:        decimal.NewFromInt(row.TotalDebit),
				TotalCredit:       decimal.NewFromInt(row.TotalCredit),
			})
		}
		return nil
	})
	return summaries, err
}

func (r *transactionRepository) UpdateNextRecurringDate(ctx context.Context, transactionID uuid.UUID, nextDate time.Time) error {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.UpdateNextRecurringDate")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		return s.UpdateRecurringTransactionNextDate(ctx, db.UpdateRecurringTransactionNextDateParams{
			TransactionID:     transactionID,
			NextRecurringDate: nextDate,
		})
	})
}

// ── Bulk / archive / restore ──────────────────────────────────────────────────

func (r *transactionRepository) BulkCreate(ctx context.Context, transactions []*domain.Transaction) error {
	for _, t := range transactions {
		if err := r.Create(ctx, t); err != nil {
			return fmt.Errorf("BulkCreate: transaction %s: %w", t.TransactionNumber, err)
		}
	}
	return nil
}

func (r *transactionRepository) BulkUpdate(ctx context.Context, transactions []*domain.Transaction) error {
	for _, t := range transactions {
		if err := r.Update(ctx, t); err != nil {
			return fmt.Errorf("BulkUpdate: transaction %s: %w", t.TransactionNumber, err)
		}
	}
	return nil
}

func (r *transactionRepository) Archive(ctx context.Context, transactionID uuid.UUID, archivedBy uuid.UUID, archivedAt time.Time) error {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.Archive")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		return r.mapDatabaseError(s.SoftDeleteTransaction(ctx, db.SoftDeleteTransactionParams{
			UpdatedBy:     &archivedBy,
			TransactionID: transactionID,
		}), "archive_transaction")
	})
}

func (r *transactionRepository) Restore(ctx context.Context, transactionID uuid.UUID, restoredBy uuid.UUID, restoredAt time.Time) error {
	// Restore from archive requires an UN-delete query not yet in SQLC — tracked in TASK-029.
	return fmt.Errorf("Restore not yet implemented (TASK-029)")
}

func (r *transactionRepository) ValidateAccountsExist(ctx context.Context, accountIDs []uuid.UUID) error {
	for _, id := range accountIDs {
		tenantID, ok := shared.GetTenantID(ctx)
		if !ok {
			return fmt.Errorf("tenant ID not found in context")
		}
		if err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
			_, err := s.GetAccountByID(ctx, id)
			return err
		}); err != nil {
			return fmt.Errorf("account %s not found: %w", id, err)
		}
	}
	return nil
}

func (r *transactionRepository) GetNextTransactionNumber(ctx context.Context, entityID *uuid.UUID, transactionType domain.TransactionType) (string, error) {
	// Sequence-based number generation requires a DB sequence or counter table.
	// Tracked in TASK-029. Using timestamp-based fallback for now.
	prefix := "TXN"
	switch transactionType {
	case domain.TransactionTypeJournalEntry:
		prefix = "JE"
	case domain.TransactionTypeManual:
		prefix = "MAN"
	case domain.TransactionTypeAdjustment:
		prefix = "ADJ"
	case domain.TransactionTypeRecurring:
		prefix = "REC"
	case domain.TransactionTypeClosing:
		prefix = "CLO"
	}
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixMilli()), nil
}

// Helper functions
func ptrStringToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func timePointerToValue(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
