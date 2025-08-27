package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

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
			EntityID:                transaction.EntityID,
			TransactionNumber:       transaction.TransactionNumber,
			TransactionType:         mapDomainTransactionTypeToSQLCEnum(transaction.TransactionType),
			TransactionStatus:       mapDomainTransactionStatusToSQLCEnum(transaction.Status),
			TransactionDate:         transaction.TransactionDate,
			PostingDate:             transaction.PostingDate,
			DueDate:                 transaction.DueDate,
			Description:             transaction.Description,
			ReferenceNumber:         transaction.ReferenceNumber,
			ExternalReference:       transaction.ExternalReference,
			CurrencyCode:            transaction.CurrencyCode,
			ExchangeRate:            decimalToPgNumeric(transaction.ExchangeRate),
			TotalDebitAmount:        decimalToPgNumeric(&transaction.TotalDebitAmount),
			TotalCreditAmount:       decimalToPgNumeric(&transaction.TotalCreditAmount),
			SourceModule:            transaction.SourceModule,
			SourceDocumentType:      transaction.SourceDocumentType,
			SourceDocumentID:        transaction.SourceDocumentID,
			BatchID:                 transaction.BatchID,
			ApprovalRequired:        boolToPtr(transaction.ApprovalRequired),
			ApprovalStatus:          mapDomainApprovalStatusToString(transaction.ApprovalStatus),
			IsRecurring:             boolToPtr(transaction.IsRecurring),
			RecurringFrequency:      transaction.RecurringFrequency,
			NextRecurringDate:       transaction.NextRecurringDate,
			TransactionAttributes:   mapAttributesToJSON(transaction.Metadata),
			CreatedBy:               transaction.CreatedBy,
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
			if err == sql.ErrNoRows {
				return domain.ErrTransactionNotFound
			}
			return r.mapDatabaseError(err, "get_transaction_by_id")
		}

		transaction, err = r.mapSQLCTransactionToDomain(&sqlcTransaction)
		return err
	})

	return transaction, err
}

func (r *transactionRepository) GetByNumber(ctx context.Context, transactionNumber string) (*domain.Transaction, error) {
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
			if err == sql.ErrNoRows {
				return domain.ErrTransactionNotFound
			}
			return r.mapDatabaseError(err, "get_transaction_by_number")
		}

		transaction, err = r.mapSQLCTransactionToDomain(&sqlcTransaction)
		return err
	})

	return transaction, err
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
			ID:                    transaction.ID,
			TransactionStatus:     mapDomainTransactionStatusToSQLCEnumPtr(transaction.Status),
			PostingDate:           transaction.PostingDate,
			DueDate:               transaction.DueDate,
			Description:           &transaction.Description,
			ReferenceNumber:       transaction.ReferenceNumber,
			ExternalReference:     transaction.ExternalReference,
			TotalDebitAmount:      decimalToPgNumeric(&transaction.TotalDebitAmount),
			TotalCreditAmount:     decimalToPgNumeric(&transaction.TotalCreditAmount),
			ApprovalStatus:        mapDomainApprovalStatusToStringPtr(transaction.ApprovalStatus),
			ApprovedBy:            transaction.ApprovedBy,
			ApprovedAt:            transaction.ApprovedAt,
			ApprovalNotes:         transaction.ApprovalNotes,
			TransactionAttributes: mapAttributesToJSON(transaction.Metadata),
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
			ID:        id,
			UpdatedBy: userIDPtr,
		})
		if err != nil {
			if err == sql.ErrNoRows {
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
		params := db.ListTransactionsParams{
			Limit:  int32(filter.Limit),
			Offset: int32(filter.Offset),
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

		if filter.FromDate != nil {
			params.DateFrom = filter.FromDate
		}

		if filter.ToDate != nil {
			params.DateTo = filter.ToDate
		}

		// Execute SQLC query
		sqlcTransactions, err := s.ListTransactions(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "list_transactions")
		}

		// Map results to domain models
		transactions = make([]*domain.Transaction, 0, len(sqlcTransactions))
		for _, sqlcTransaction := range sqlcTransactions {
			transaction, err := r.mapSQLCTransactionToDomain(&sqlcTransaction)
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

		if filter.FromDate != nil {
			params.DateFrom = filter.FromDate
		}

		if filter.ToDate != nil {
			params.DateTo = filter.ToDate
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

func (r *transactionRepository) Post(ctx context.Context, id uuid.UUID) error {
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
			ID:        id,
			PostedBy:  &userID,
			UpdatedBy: userID,
		})
		if err != nil {
			return r.mapDatabaseError(err, "post_transaction")
		}
		return nil
	})
}

func (r *transactionRepository) Approve(ctx context.Context, id uuid.UUID, notes *string) error {
	ctx, span := r.tracing.StartSpan(ctx, "TransactionRepository.Approve")
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
		_, err := s.ApproveTransaction(ctx, db.ApproveTransactionParams{
			ID:            id,
			ApprovedBy:    userID,
			ApprovalNotes: notes,
		})
		if err != nil {
			return r.mapDatabaseError(err, "approve_transaction")
		}
		return nil
	})
}

func (r *transactionRepository) Reject(ctx context.Context, id uuid.UUID, notes *string) error {
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
		_, err := s.RejectTransaction(ctx, db.RejectTransactionParams{
			ID:            id,
			ApprovedBy:    userID,
			ApprovalNotes: notes,
		})
		if err != nil {
			return r.mapDatabaseError(err, "reject_transaction")
		}
		return nil
	})
}

func (r *transactionRepository) Reverse(ctx context.Context, id uuid.UUID, reversalTransactionID uuid.UUID, reason string) error {
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
			ID:                       id,
			ReversedByTransactionID: &reversalTransactionID,
			ReversalReason:          &reason,
			UpdatedBy:               userID,
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
		transaction, err := r.mapSQLCTransactionRowToDomain(&firstResult)
		if err != nil {
			return err
		}

		// Map entries
		entries := make([]*domain.TransactionEntry, 0)
		for _, result := range results {
			if result.EntryID != nil {
				entry := &domain.TransactionEntry{
					ID:              *result.EntryID,
					TransactionID:   transaction.ID,
					EntryNumber:     result.EntryNumber,
					AccountID:       result.AccountID,
					DebitAmount:     pgNumericToDecimal(result.DebitAmount),
					CreditAmount:    pgNumericToDecimal(result.CreditAmount),
					Description:     result.EntryDescription,
					Reference:       result.EntryReference,
					CostCenter:      result.CostCenter,
					Department:      result.Department,
					ProjectID:       result.ProjectID,
				}
				entries = append(entries, entry)
			}
		}

		transactionWithEntries = &domain.TransactionWithEntries{
			Transaction: *transaction,
			Entries:     entries,
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
			transaction, err := r.mapSQLCTransactionToDomain(&sqlcTransaction)
			if err != nil {
				return fmt.Errorf("failed to map SQLC transaction to domain: %w", err)
			}
			transactions = append(transactions, transaction)
		}
		return nil
	})

	return transactions, err
}

func (r *transactionRepository) GetPendingApproval(ctx context.Context, limit, offset int32) ([]*domain.Transaction, error) {
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
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return r.mapDatabaseError(err, "get_pending_approval_transactions")
		}

		transactions = make([]*domain.Transaction, 0, len(sqlcTransactions))
		for _, sqlcTransaction := range sqlcTransactions {
			transaction, err := r.mapSQLCTransactionToDomain(&sqlcTransaction)
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
			Search: query,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return r.mapDatabaseError(err, "search_transactions")
		}

		transactions = make([]*domain.Transaction, 0, len(sqlcTransactions))
		for _, sqlcTransaction := range sqlcTransactions {
			transaction, err := r.mapSQLCTransactionToDomain(&sqlcTransaction)
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
	if err == sql.ErrNoRows {
		return domain.ErrTransactionNotFound
	}

	// Check for constraint violations
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "23505": // unique violation
			return domain.ErrTransactionNumberExists
		case "23503": // foreign key violation
			return fmt.Errorf("invalid reference in transaction")
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
		ID:                sqlcTransaction.ID,
		TenantID:          sqlcTransaction.TenantID,
		EntityID:          sqlcTransaction.EntityID,
		TransactionNumber: sqlcTransaction.TransactionNumber,
		TransactionType:   mapSQLCTransactionTypeToDomain(sqlcTransaction.TransactionType),
		Status:            mapSQLCTransactionStatusToDomain(sqlcTransaction.TransactionStatus),
		TransactionDate:   sqlcTransaction.TransactionDate,
		PostingDate:       sqlcTransaction.PostingDate,
		DueDate:           sqlcTransaction.DueDate,
		Description:       sqlcTransaction.Description,
		ReferenceNumber:   sqlcTransaction.ReferenceNumber,
		ExternalReference: sqlcTransaction.ExternalReference,
		CurrencyCode:      sqlcTransaction.CurrencyCode,
		ExchangeRate:      pgNumericToDecimalPtr(sqlcTransaction.ExchangeRate),
		TotalDebitAmount:  pgNumericToDecimal(sqlcTransaction.TotalDebitAmount),
		TotalCreditAmount: pgNumericToDecimal(sqlcTransaction.TotalCreditAmount),
		SourceModule:      sqlcTransaction.SourceModule,
		SourceDocumentType: sqlcTransaction.SourceDocumentType,
		SourceDocumentID:  sqlcTransaction.SourceDocumentID,
		BatchID:           sqlcTransaction.BatchID,
		ApprovalRequired:  getBoolValue(sqlcTransaction.ApprovalRequired),
		ApprovalStatus:    mapStringToApprovalStatus(sqlcTransaction.ApprovalStatus),
		ApprovedBy:        sqlcTransaction.ApprovedBy,
		ApprovedAt:        sqlcTransaction.ApprovedAt,
		ApprovalNotes:     sqlcTransaction.ApprovalNotes,
		IsRecurring:       getBoolValue(sqlcTransaction.IsRecurring),
		RecurringFrequency: sqlcTransaction.RecurringFrequency,
		NextRecurringDate: sqlcTransaction.NextRecurringDate,
		Metadata:          metadata,
		CreatedAt:         sqlcTransaction.CreatedAt,
		UpdatedAt:         sqlcTransaction.UpdatedAt,
		CreatedBy:         sqlcTransaction.CreatedBy,
		UpdatedBy:         sqlcTransaction.UpdatedBy,
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
		ID:                row.ID,
		TenantID:          row.TenantID,
		EntityID:          row.EntityID,
		TransactionNumber: row.TransactionNumber,
		TransactionType:   mapSQLCTransactionTypeToDomain(row.TransactionType),
		Status:            mapSQLCTransactionStatusToDomain(row.TransactionStatus),
		TransactionDate:   row.TransactionDate,
		PostingDate:       row.PostingDate,
		DueDate:           row.DueDate,
		Description:       row.Description,
		ReferenceNumber:   row.ReferenceNumber,
		ExternalReference: row.ExternalReference,
		CurrencyCode:      row.CurrencyCode,
		ExchangeRate:      pgNumericToDecimalPtr(row.ExchangeRate),
		TotalDebitAmount:  pgNumericToDecimal(row.TotalDebitAmount),
		TotalCreditAmount: pgNumericToDecimal(row.TotalCreditAmount),
		SourceModule:      row.SourceModule,
		SourceDocumentType: row.SourceDocumentType,
		SourceDocumentID:  row.SourceDocumentID,
		BatchID:           row.BatchID,
		ApprovalRequired:  getBoolValue(row.ApprovalRequired),
		ApprovalStatus:    mapStringToApprovalStatus(row.ApprovalStatus),
		ApprovedBy:        row.ApprovedBy,
		ApprovedAt:        row.ApprovedAt,
		ApprovalNotes:     row.ApprovalNotes,
		IsRecurring:       getBoolValue(row.IsRecurring),
		RecurringFrequency: row.RecurringFrequency,
		NextRecurringDate: row.NextRecurringDate,
		Metadata:          metadata,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
		CreatedBy:         row.CreatedBy,
		UpdatedBy:         row.UpdatedBy,
	}, nil
}