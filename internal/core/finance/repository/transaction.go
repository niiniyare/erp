package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/shared"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

type transactionRepository struct {
	store   db.Store
	tracing tracing.TracingService
}

func NewTransactionRepository(store db.Store, tracing tracing.TracingService) domain.TransactionRepository {
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
			Memo:                  getStringValue(transaction.ReferenceNumber), // Use reference as memo if available
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

	sqlcTransaction, err := r.store.GetTransactionByID(ctx, id)
	if err != nil {
		if err == db.ErrNoRows {
			return nil, domain.ErrTransactionNotFound
		}
		return nil, r.mapDatabaseError(err, "get_transaction_by_id")
	}

	transaction, err := r.mapSQLCTransactionToDomain(sqlcTransaction)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (r *transactionRepository) GetByNumber(ctx context.Context, entityID *uuid.UUID, transactionNumber string) (*domain.Transaction, error) {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.GetByNumber")
	defer span.End()

	sqlcTransaction, err := r.store.GetTransactionByNumber(ctx, transactionNumber)
	if err != nil {
		if err == db.ErrNoRows {
			return nil, domain.ErrTransactionNotFound
		}
		return nil, r.mapDatabaseError(err, "get_transaction_by_number")
	}

	transaction, err := r.mapSQLCTransactionToDomain(sqlcTransaction)
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
			Description:           transaction.Description,
			ReferenceNumber:       transaction.ReferenceNumber,
			ExternalReference:     transaction.ExternalReference,
			Memo:                  getStringValue(transaction.ReferenceNumber), // Use reference as memo
			TotalDebitAmount:      decimalToPgNumeric(&transaction.TotalDebitAmount),
			TotalCreditAmount:     decimalToPgNumeric(&transaction.TotalCreditAmount),
			ApprovalStatus:        mapDomainApprovalStatusToNullEnum(&transaction.ApprovalStatus),
			ApprovedBy:            transaction.ApprovedBy,
			ApprovedAt:            timePointerToNullTime(transaction.ApprovedAt),
			ApprovalNotes:         getStringValue(transaction.ApprovalNotes),
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
		var userIDPtr *uuid.UUID
		if userID != uuid.Nil {
			userIDPtr = &userID
		}

		err := s.SoftDeleteTransaction(ctx, db.SoftDeleteTransactionParams{
			TransactionID: id,
			UpdatedBy:     userIDPtr,
		})
		if err != nil {
			if err == db.ErrNoRows {
				return domain.ErrTransactionNotFound
			}
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
			}
			if filter.DateRange.EndDate != nil {
				params.DateTo = *filter.DateRange.EndDate
			}
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
			ApprovalNotes: notesStr,
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
			ApprovalNotes: notesStr,
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
			ReversalReason:          reason,
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
					Description:   result.EntryDescription,
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
			SearchTerm:  query,
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
	var metadata map[string]interface{}
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
		ApprovalNotes:         stringPtr(sqlcTransaction.ApprovalNotes),
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
	var metadata map[string]interface{}
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
		ApprovalNotes:         stringPtr(row.ApprovalNotes),
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

	var balance decimal.Decimal
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// TODO: Implement proper balance calculation logic
		// This would involve summing debit/credit entries for the account
		// For now, return zero as a stub implementation
		balance = decimal.Zero
		return nil
	})

	return balance, err
}

// Stub implementations for missing interface methods
// These are minimal implementations to satisfy the interface

func (r *transactionRepository) CreateEntries(ctx context.Context, entries []*domain.TransactionEntry) error {
	// TODO: Implement proper entry creation
	return fmt.Errorf("CreateEntries not implemented")
}

func (r *transactionRepository) ListByAccount(ctx context.Context, accountID uuid.UUID, filter *domain.TransactionFilter) ([]*domain.Transaction, error) {
	// TODO: Implement proper account-based listing
	return nil, fmt.Errorf("ListByAccount not implemented")
}

func (r *transactionRepository) ListByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*domain.Transaction, error) {
	// TODO: Implement proper date range listing
	return nil, fmt.Errorf("ListByDateRange not implemented")
}

func (r *transactionRepository) GetByStatus(ctx context.Context, status domain.TransactionStatus, limit int) ([]*domain.Transaction, error) {
	// TODO: Implement proper status-based retrieval
	return nil, fmt.Errorf("GetByStatus not implemented")
}

func (r *transactionRepository) GetPendingApproval(ctx context.Context, userID *uuid.UUID) ([]*domain.Transaction, error) {
	// Use existing implementation but adapt signature
	return r.GetPendingApprovalTransactions(ctx, 50, 0)
}

func (r *transactionRepository) GetRecurringTransactions(ctx context.Context, dueDate time.Time) ([]*domain.Transaction, error) {
	// TODO: Implement proper recurring transaction retrieval
	return nil, fmt.Errorf("GetRecurringTransactions not implemented")
}

func (r *transactionRepository) CreateEntry(ctx context.Context, entry *domain.TransactionEntry) error {
	// TODO: Implement proper single entry creation
	return fmt.Errorf("CreateEntry not implemented")
}

func (r *transactionRepository) GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.TransactionEntry, error) {
	// TODO: Implement proper entry retrieval
	return nil, fmt.Errorf("GetEntryByID not implemented")
}

func (r *transactionRepository) GetEntriesByTransaction(ctx context.Context, transactionID uuid.UUID) ([]domain.TransactionEntry, error) {
	// TODO: Implement proper entries by transaction retrieval
	return nil, fmt.Errorf("GetEntriesByTransaction not implemented")
}

func (r *transactionRepository) GetEntriesByAccount(ctx context.Context, accountID uuid.UUID, filter *domain.EntryFilter) ([]domain.TransactionEntry, error) {
	// TODO: Implement proper entries by account retrieval
	return nil, fmt.Errorf("GetEntriesByAccount not implemented")
}

func (r *transactionRepository) UpdateEntry(ctx context.Context, entry *domain.TransactionEntry) error {
	// TODO: Implement proper entry update
	return fmt.Errorf("UpdateEntry not implemented")
}

func (r *transactionRepository) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	// TODO: Implement proper entry deletion
	return fmt.Errorf("DeleteEntry not implemented")
}

func (r *transactionRepository) SearchEntries(ctx context.Context, query string, filters *domain.EntryFilter, limit int, offset int) ([]*domain.TransactionEntry, error) {
	// TODO: Implement proper entry search
	return nil, fmt.Errorf("SearchEntries not implemented")
}

func (r *transactionRepository) UpdateReconciliationStatus(ctx context.Context, entryID uuid.UUID, reconciled bool, reconciledDate *time.Time, reconciliationRef *string) error {
	// TODO: Implement proper reconciliation status update
	return fmt.Errorf("UpdateReconciliationStatus not implemented")
}

func (r *transactionRepository) GetUnreconciledEntries(ctx context.Context, accountID uuid.UUID, cutoffDate *time.Time) ([]*domain.TransactionEntry, error) {
	// TODO: Implement proper unreconciled entries retrieval
	return nil, fmt.Errorf("GetUnreconciledEntries not implemented")
}

func (r *transactionRepository) GetEntrySummary(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) (*domain.TransactionSummary, error) {
	// TODO: Implement proper entry summary
	return nil, fmt.Errorf("GetEntrySummary not implemented")
}

func (r *transactionRepository) GetAccountTransactionSummary(ctx context.Context, accountID uuid.UUID, dateRange *domain.DateRange) (*domain.TransactionSummary, error) {
	// TODO: Implement proper account transaction summary
	return nil, fmt.Errorf("GetAccountTransactionSummary not implemented")
}

func (r *transactionRepository) Post(ctx context.Context, transactionID uuid.UUID, postedBy uuid.UUID, postedAt time.Time) error {
	// Use existing implementation but adapt signature
	return r.PostTransaction(ctx, transactionID)
}

func (r *transactionRepository) Approve(ctx context.Context, transactionID uuid.UUID, approvedBy uuid.UUID, approvedAt time.Time, notes *string) error {
	// Use existing implementation but adapt signature
	return r.ApproveTransaction(ctx, transactionID, approvedBy, approvedAt, notes)
}

func (r *transactionRepository) Reject(ctx context.Context, transactionID uuid.UUID, rejectedBy uuid.UUID, rejectedAt time.Time, reason domain.RejectionReason, notes *string) error {
	// Use existing implementation but adapt signature
	return r.RejectTransaction(ctx, transactionID, notes)
}

func (r *transactionRepository) Reverse(ctx context.Context, originalID, reversalID uuid.UUID, reason string) error {
	// Use existing implementation but adapt signature
	return r.ReverseTransaction(ctx, originalID, reversalID, reason)
}

func (r *transactionRepository) IsTransactionNumberUnique(ctx context.Context, entityID *uuid.UUID, transactionNumber string, excludeID *uuid.UUID) (bool, error) {
	// TODO: Implement proper uniqueness check
	return true, nil
}

func (r *transactionRepository) GetTransactionSummary(ctx context.Context, filter *domain.TransactionFilter) ([]*domain.TransactionSummary, error) {
	// TODO: Implement proper transaction summary
	return nil, fmt.Errorf("GetTransactionSummary not implemented")
}

func (r *transactionRepository) ValidateTransaction(ctx context.Context, transaction *domain.Transaction) ([]domain.ValidationError, error) {
	// TODO: Implement proper transaction validation
	return transaction.Validate(), nil
}

func (r *transactionRepository) BulkCreate(ctx context.Context, transactions []*domain.Transaction) error {
	// TODO: Implement proper bulk creation
	return fmt.Errorf("BulkCreate not implemented")
}

func (r *transactionRepository) BulkUpdate(ctx context.Context, transactions []*domain.Transaction) error {
	// TODO: Implement proper bulk update
	return fmt.Errorf("BulkUpdate not implemented")
}

func (r *transactionRepository) Archive(ctx context.Context, transactionID uuid.UUID, archivedBy uuid.UUID, archivedAt time.Time) error {
	// TODO: Implement proper archiving
	return fmt.Errorf("Archive not implemented")
}

func (r *transactionRepository) Restore(ctx context.Context, transactionID uuid.UUID, restoredBy uuid.UUID, restoredAt time.Time) error {
	// TODO: Implement proper restoration
	return fmt.Errorf("Restore not implemented")
}

func (r *transactionRepository) ValidateAccountsExist(ctx context.Context, accountIDs []uuid.UUID) error {
	// TODO: Implement proper account existence validation
	return fmt.Errorf("ValidateAccountsExist not implemented")
}

func (r *transactionRepository) GetNextTransactionNumber(ctx context.Context, entityID *uuid.UUID, transactionType domain.TransactionType) (string, error) {
	// TODO: Implement proper transaction number generation
	return "TXN-001", nil
}

// Helper function
func timePointerToValue(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
