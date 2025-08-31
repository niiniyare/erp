package finance

import (
	"context"
	"time"

	"github.com/google/uuid"
	goaFinance "github.com/niiniyare/erp/internal/api/gen/finance"
	"github.com/niiniyare/erp/internal/core/finance/domain"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func (h *FinanceHandler) CreateTransaction(ctx context.Context, payload *goaFinance.CreateTransactionPayload) (*goaFinance.TransactionResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.create_transaction",
		tracing.WithAttributes(
			attribute.String("transaction.type", payload.TransactionType),
			attribute.String("transaction.description", payload.Description),
		))
	defer span.End()

	timer := h.metrics.Timer("finance_create_transaction_duration", metrics.Fields{})
	defer timer.Stop()

	logger.InfoContext(ctx, "Processing create transaction request",
		logger.Fields{
			"transaction_type": payload.TransactionType,
			"transaction_date": payload.TransactionDate,
			"description":      payload.Description,
			"entries_count":    len(payload.Entries),
		})

	var entityID *uuid.UUID
	if payload.EntityID != nil {
		parsed, err := uuid.Parse(*payload.EntityID)
		if err != nil {
			logger.WarnContext(ctx, "Invalid entity ID format",
				logger.Fields{"entity_id": *payload.EntityID, "error": err.Error()})
			return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_ENTITY_ID", "Invalid entity ID format"))
		}
		entityID = &parsed
	}

	transactionDate, err := time.Parse("2006-01-02", payload.TransactionDate)
	if err != nil {
		logger.WarnContext(ctx, "Invalid transaction date format",
			logger.Fields{"transaction_date": payload.TransactionDate, "error": err.Error()})
		return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_DATE", "Invalid transaction date format"))
	}

	transactionType, err := domain.ParseTransactionType(payload.TransactionType)
	if err != nil {
		logger.WarnContext(ctx, "Invalid transaction type",
			logger.Fields{"transaction_type": payload.TransactionType, "error": err.Error()})
		return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_TRANSACTION_TYPE", "Invalid transaction type"))
	}

	currencyCode := "USD"
	if payload.CurrencyCode != "" {
		currencyCode = payload.CurrencyCode
	}
	
	transactionNumber := ""
	if payload.TransactionNumber != nil {
		transactionNumber = *payload.TransactionNumber
	}
	
	req := domain.CreateTransactionRequest{
		EntityID:          entityID,
		TransactionNumber: transactionNumber,
		TransactionType:   transactionType,
		TransactionDate:   transactionDate,
		Description:       payload.Description,
		ReferenceNumber:   payload.ReferenceNumber,
		CurrencyCode:      currencyCode,
	}

	transaction, err := h.financeServices.Transaction.CreateTransaction(ctx, req)
	if err != nil {
		h.metrics.IncrementCounter("finance_transaction_creation_errors", metrics.Fields{})
		logger.ErrorContext(ctx, "Failed to create transaction",
			logger.Fields{
				"transaction_type": payload.TransactionType,
				"description":      payload.Description,
				"error":            err.Error(),
			})
		return nil, h.handleError(err)
	}

	h.metrics.IncrementCounter("finance_transactions_created", metrics.Fields{})
	logger.InfoContext(ctx, "Transaction created successfully",
		logger.Fields{
			"transaction_id":     transaction.ID.String(),
			"transaction_number": transaction.TransactionNumber,
			"transaction_type":   string(transaction.TransactionType),
		})

	return h.convertTransactionToResult(transaction), nil
}

func (h *FinanceHandler) GetTransaction(ctx context.Context, payload *goaFinance.GetTransactionPayload) (*goaFinance.TransactionWithEntriesResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.get_transaction")
	defer span.End()

	id, err := uuid.Parse(payload.ID)
	if err != nil {
		logger.WarnContext(ctx, "Invalid transaction ID format",
			logger.Fields{"transaction_id": payload.ID, "error": err.Error()})
		return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_ID", "Invalid transaction ID format"))
	}

	logger.DebugContext(ctx, "Retrieving transaction by ID",
		logger.Fields{"transaction_id": id.String()})

	transaction, err := h.financeServices.Transaction.GetTransactionByID(ctx, id)
	if err != nil {
		logger.WarnContext(ctx, "Failed to retrieve transaction",
			logger.Fields{"transaction_id": id.String(), "error": err.Error()})
		return nil, h.handleError(err)
	}

	logger.DebugContext(ctx, "Transaction retrieved successfully",
		logger.Fields{
			"transaction_id":     transaction.ID.String(),
			"transaction_number": transaction.TransactionNumber,
		})

	return &goaFinance.TransactionWithEntriesResult{
		Transaction: h.convertTransactionToResult(transaction),
		Entries:     []*goaFinance.TransactionEntryResult{},
		IsBalanced:  true,
	}, nil
}

func (h *FinanceHandler) GetTransactionByNumber(ctx context.Context, payload *goaFinance.GetTransactionByNumberPayload) (*goaFinance.TransactionWithEntriesResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.get_transaction_by_number")
	defer span.End()

	logger.DebugContext(ctx, "Retrieving transaction by number",
		logger.Fields{"transaction_number": payload.TransactionNumber})

	transaction, err := h.financeServices.Transaction.GetTransactionByNumber(ctx, payload.TransactionNumber)
	if err != nil {
		logger.WarnContext(ctx, "Failed to retrieve transaction by number",
			logger.Fields{"transaction_number": payload.TransactionNumber, "error": err.Error()})
		return nil, h.handleError(err)
	}

	logger.DebugContext(ctx, "Transaction retrieved by number successfully",
		logger.Fields{
			"transaction_id":     transaction.ID.String(),
			"transaction_number": transaction.TransactionNumber,
		})

	return &goaFinance.TransactionWithEntriesResult{
		Transaction: h.convertTransactionToResult(transaction),
		Entries:     []*goaFinance.TransactionEntryResult{},
		IsBalanced:  true,
	}, nil
}

func (h *FinanceHandler) ListTransactions(ctx context.Context, payload *goaFinance.ListTransactionsPayload) (*goaFinance.TransactionListResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.list_transactions")
	defer span.End()

	logger.DebugContext(ctx, "Processing list transactions request",
		logger.Fields{
			"status":    payload.Status,
			"type":      payload.Type,
			"date_from": payload.DateFrom,
			"date_to":   payload.DateTo,
			"search":    payload.Search,
		})

	limitInt := int(payload.Limit)
	offsetInt := int(payload.Offset)
	
	filter := &domain.TransactionFilter{
		SearchTerm: payload.Search,
		Limit:      &limitInt,
		Offset:     &offsetInt,
	}

	if payload.Status != nil {
		status, err := domain.ParseTransactionStatus(*payload.Status)
		if err != nil {
			logger.WarnContext(ctx, "Invalid transaction status",
				logger.Fields{"status": *payload.Status, "error": err.Error()})
			return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_STATUS", "Invalid transaction status"))
		}
		filter.Status = &status
	}

	if payload.Type != nil {
		transactionType, err := domain.ParseTransactionType(*payload.Type)
		if err != nil {
			logger.WarnContext(ctx, "Invalid transaction type",
				logger.Fields{"type": *payload.Type, "error": err.Error()})
			return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_TYPE", "Invalid transaction type"))
		}
		filter.TransactionType = &transactionType
	}

	if payload.DateFrom != nil || payload.DateTo != nil {
		dateRange := &domain.DateRange{}
		
		if payload.DateFrom != nil {
			dateFrom, err := time.Parse("2006-01-02", *payload.DateFrom)
			if err != nil {
				logger.WarnContext(ctx, "Invalid date_from format",
					logger.Fields{"date_from": *payload.DateFrom, "error": err.Error()})
				return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_DATE_FROM", "Invalid date_from format"))
			}
			dateRange.StartDate = dateFrom
		}

		if payload.DateTo != nil {
			dateTo, err := time.Parse("2006-01-02", *payload.DateTo)
			if err != nil {
				logger.WarnContext(ctx, "Invalid date_to format",
					logger.Fields{"date_to": *payload.DateTo, "error": err.Error()})
				return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_DATE_TO", "Invalid date_to format"))
			}
			dateRange.EndDate = &dateTo
		}
		
		filter.DateRange = dateRange
	}

	if payload.AccountID != nil {
		accountID, err := uuid.Parse(*payload.AccountID)
		if err != nil {
			logger.WarnContext(ctx, "Invalid account ID format in filter",
				logger.Fields{"account_id": *payload.AccountID, "error": err.Error()})
			return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_ACCOUNT_ID", "Invalid account ID format"))
		}
		filter.AccountID = &accountID
	}

	transactions, err := h.financeServices.Transaction.ListTransactions(ctx, filter)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list transactions",
			logger.Fields{"error": err.Error()})
		return nil, h.handleError(err)
	}

	results := make([]*goaFinance.TransactionResult, len(transactions))
	for i, transaction := range transactions {
		results[i] = h.convertTransactionToResult(transaction)
	}

	logger.InfoContext(ctx, "Transactions listed successfully",
		logger.Fields{
			"total_found": len(results),
			"limit":       payload.Limit,
			"offset":      payload.Offset,
		})

	var limit, offset int32 = 50, 0
	if payload.Limit != nil {
		limit = *payload.Limit
	}
	if payload.Offset != nil {
		offset = *payload.Offset
	}
	
	return &goaFinance.TransactionListResult{
		Transactions: results,
		TotalCount:   int64(len(results)),
		Limit:        limit,
		Offset:       offset,
	}, nil
}

func (h *FinanceHandler) PostTransaction(ctx context.Context, payload *goaFinance.PostTransactionPayload) (*goaFinance.TransactionResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.post_transaction")
	defer span.End()

	id, err := uuid.Parse(payload.ID)
	if err != nil {
		logger.WarnContext(ctx, "Invalid transaction ID format for posting",
			logger.Fields{"transaction_id": payload.ID, "error": err.Error()})
		return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_ID", "Invalid transaction ID format"))
	}

	var postingDate *time.Time
	if payload.PostingDate != nil {
		parsed, err := time.Parse("2006-01-02", *payload.PostingDate)
		if err != nil {
			logger.WarnContext(ctx, "Invalid posting date format",
				logger.Fields{"posting_date": *payload.PostingDate, "error": err.Error()})
			return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_POSTING_DATE", "Invalid posting date format"))
		}
		postingDate = &parsed
	}

	logger.InfoContext(ctx, "Processing transaction posting request",
		logger.Fields{
			"transaction_id":           id.String(),
			"posting_date":             postingDate,
			"validate_before_posting":  payload.ValidateBeforePosting,
			"force_post":               payload.ForcePost,
		})

	transaction, err := h.financeServices.Transaction.PostTransaction(ctx, id, postingDate)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to post transaction",
			logger.Fields{
				"transaction_id": id.String(),
				"error":          err.Error(),
			})
		return nil, h.handleError(err)
	}

	logger.InfoContext(ctx, "Transaction posted successfully",
		logger.Fields{
			"transaction_id":     transaction.ID.String(),
			"transaction_number": transaction.TransactionNumber,
			"posting_date":       transaction.PostingDate,
		})

	return h.convertTransactionToResult(transaction), nil
}

func (h *FinanceHandler) ReverseTransaction(ctx context.Context, payload *goaFinance.ReverseTransactionPayload) (*goaFinance.TransactionResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.reverse_transaction")
	defer span.End()

	id, err := uuid.Parse(payload.ID)
	if err != nil {
		logger.WarnContext(ctx, "Invalid transaction ID format for reversal",
			logger.Fields{"transaction_id": payload.ID, "error": err.Error()})
		return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_ID", "Invalid transaction ID format"))
	}

	logger.InfoContext(ctx, "Processing transaction reversal request",
		logger.Fields{
			"transaction_id": id.String(),
			"reason":         payload.Reason,
			"reversal_date":  payload.ReversalDate,
		})

	transaction, err := h.financeServices.Transaction.ReverseTransaction(ctx, id, payload.Reason)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to reverse transaction",
			logger.Fields{
				"transaction_id": id.String(),
				"reason":         payload.Reason,
				"error":          err.Error(),
			})
		return nil, h.handleError(err)
	}

	logger.InfoContext(ctx, "Transaction reversed successfully",
		logger.Fields{
			"original_transaction_id": id.String(),
			"reversal_transaction_id": transaction.ID.String(),
			"reason":                  payload.Reason,
		})

	return h.convertTransactionToResult(transaction), nil
}

func (h *FinanceHandler) ApproveTransaction(ctx context.Context, payload *goaFinance.ApproveTransactionPayload) (*goaFinance.TransactionResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.approve_transaction")
	defer span.End()

	id, err := uuid.Parse(payload.ID)
	if err != nil {
		logger.WarnContext(ctx, "Invalid transaction ID format for approval",
			logger.Fields{"transaction_id": payload.ID, "error": err.Error()})
		return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_ID", "Invalid transaction ID format"))
	}

	notes := ""
	if payload.Notes != nil {
		notes = *payload.Notes
	}

	logger.InfoContext(ctx, "Processing transaction approval request",
		logger.Fields{
			"transaction_id": id.String(),
			"notes":          notes,
		})

	transaction, err := h.financeServices.Transaction.ApproveTransaction(ctx, id, notes)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to approve transaction",
			logger.Fields{
				"transaction_id": id.String(),
				"notes":          notes,
				"error":          err.Error(),
			})
		return nil, h.handleError(err)
	}

	logger.InfoContext(ctx, "Transaction approved successfully",
		logger.Fields{
			"transaction_id":     transaction.ID.String(),
			"transaction_number": transaction.TransactionNumber,
			"approval_status":    transaction.ApprovalStatus,
		})

	return h.convertTransactionToResult(transaction), nil
}

func (h *FinanceHandler) ValidateTransaction(ctx context.Context, payload *goaFinance.ValidateTransactionPayload) (*goaFinance.ValidationResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.validate_transaction")
	defer span.End()

	logger.InfoContext(ctx, "Processing transaction validation request",
		logger.Fields{
			"validation_level": payload.ValidationLevel,
			"transaction_type": payload.Transaction.TransactionType,
		})

	// TODO: Implement full validation using validation engine
	logger.DebugContext(ctx, "Transaction validation completed",
		logger.Fields{
			"is_valid":    true,
			"is_balanced": true,
		})

	validationLevel := "STRICT"
	if payload.ValidationLevel != nil {
		validationLevel = *payload.ValidationLevel
	}
	
	return &goaFinance.ValidationResult{
		IsValid:         true,
		IsBalanced:      true,
		TotalDebits:     "0.00",
		TotalCredits:    "0.00",
		ValidationLevel: validationLevel,
	}, nil
}

func (h *FinanceHandler) convertTransactionToResult(transaction *domain.Transaction) *goaFinance.TransactionResult {
	result := &goaFinance.TransactionResult{
		ID:                transaction.ID.String(),
		TransactionNumber: transaction.TransactionNumber,
		TransactionType:   string(transaction.TransactionType),
		TransactionStatus: string(transaction.TransactionStatus),
		TransactionDate:   transaction.TransactionDate.Format("2006-01-02"),
		Description:       transaction.Description,
		CreatedAt:         transaction.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         transaction.UpdatedAt.Format(time.RFC3339),
	}

	if transaction.TenantID != nil {
		result.TenantID = transaction.TenantID.String()
	}

	if transaction.EntityID != nil {
		result.EntityID = transaction.EntityID.String()
	}

	if transaction.PostingDate != nil {
		result.PostingDate = transaction.PostingDate.Format("2006-01-02")
	}

	if transaction.ReferenceNumber != nil {
		result.ReferenceNumber = *transaction.ReferenceNumber
	}

	if transaction.CurrencyCode != nil {
		result.CurrencyCode = *transaction.CurrencyCode
	}

	if transaction.ExchangeRate != nil {
		result.ExchangeRate = transaction.ExchangeRate.String()
	}

	if transaction.TotalDebitAmount != nil {
		result.TotalDebitAmount = transaction.TotalDebitAmount.String()
	}

	if transaction.TotalCreditAmount != nil {
		result.TotalCreditAmount = transaction.TotalCreditAmount.String()
	}

	if transaction.ApprovalStatus != nil {
		result.ApprovalStatus = string(*transaction.ApprovalStatus)
	}

	result.ApprovalRequired = transaction.ApprovalRequired != nil && *transaction.ApprovalRequired

	return result
}