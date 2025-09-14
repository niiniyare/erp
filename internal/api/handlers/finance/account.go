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

func (h *FinanceHandler) CreateAccount(ctx context.Context, payload *goaFinance.CreateAccountPayload) (*goaFinance.AccountResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.create_account",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("account.code", payload.AccountCode),
			attribute.String("account.name", payload.AccountName),
		))
	defer span.End()

	timer := h.metrics.Timer("finance_create_account_duration", metrics.Fields{})
	defer timer.Stop()

	logger.InfoContext(ctx, "Processing create account request",
		logger.Fields{
			"account_code": payload.AccountCode,
			"account_name": payload.AccountName,
			"root_type":    payload.RootType,
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

	var parentAccountID *uuid.UUID
	if payload.ParentAccountID != nil {
		parsed, err := uuid.Parse(*payload.ParentAccountID)
		if err != nil {
			logger.WarnContext(ctx, "Invalid parent account ID format",
				logger.Fields{"parent_account_id": *payload.ParentAccountID, "error": err.Error()})
			return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_PARENT_ID", "Invalid parent account ID format"))
		}
		parentAccountID = &parsed
	}

	rootType, err := domain.ParseRootType(payload.RootType)
	if err != nil {
		logger.WarnContext(ctx, "Invalid root type",
			logger.Fields{"root_type": payload.RootType, "error": err.Error()})
		return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_ROOT_TYPE", "Invalid root type"))
	}

	normalBalance, err := domain.ParseNormalBalance(payload.NormalBalance)
	if err != nil {
		logger.WarnContext(ctx, "Invalid normal balance",
			logger.Fields{"normal_balance": payload.NormalBalance, "error": err.Error()})
		return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_NORMAL_BALANCE", "Invalid normal balance"))
	}

	accountType := payload.AccountType

	req := domain.CreateAccountRequest{
		EntityID:           entityID,
		AccountCode:        payload.AccountCode,
		AccountName:        payload.AccountName,
		AccountDescription: payload.AccountDescription,
		ParentAccountID:    parentAccountID,
		RootType:           rootType,
		AccountType:        accountType,
		AccountSubtype:     payload.AccountSubtype,
		NormalBalance:      normalBalance,
		CurrencyCode:       payload.CurrencyCode,
		IsActive:           payload.IsActive,
	}

	account, err := h.financeServices.Account.CreateAccount(ctx, req)
	if err != nil {
		h.metrics.IncrementCounter("finance_account_creation_errors", metrics.Fields{})
		logger.ErrorContext(ctx, "Failed to create account",
			logger.Fields{
				"account_code": payload.AccountCode,
				"error":        err.Error(),
			})
		return nil, h.handleError(err)
	}

	h.metrics.IncrementCounter("finance_accounts_created", metrics.Fields{})
	logger.InfoContext(ctx, "Account created successfully",
		logger.Fields{
			"account_id":   account.ID.String(),
			"account_code": account.AccountCode,
			"account_name": account.AccountName,
		})

	return h.convertAccountToResult(account), nil
}

func (h *FinanceHandler) GetAccount(ctx context.Context, payload *goaFinance.GetAccountByIDPayload) (*goaFinance.AccountResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.get_account")
	defer span.End()

	id, err := uuid.Parse(payload.ID)
	if err != nil {
		logger.WarnContext(ctx, "Invalid account ID format",
			logger.Fields{"account_id": payload.ID, "error": err.Error()})
		return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_ID", "Invalid account ID format"))
	}

	logger.DebugContext(ctx, "Retrieving account by ID",
		logger.Fields{"account_id": id.String()})

	account, err := h.financeServices.Account.GetAccountByID(ctx, id)
	if err != nil {
		logger.WarnContext(ctx, "Failed to retrieve account",
			logger.Fields{"account_id": id.String(), "error": err.Error()})
		return nil, h.handleError(err)
	}

	logger.DebugContext(ctx, "Account retrieved successfully",
		logger.Fields{
			"account_id":   account.ID.String(),
			"account_code": account.AccountCode,
		})

	return h.convertAccountToResult(account), nil
}

func (h *FinanceHandler) GetAccountByCode(ctx context.Context, payload *goaFinance.GetAccountByCodePayload) (*goaFinance.AccountResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.get_account_by_code")
	defer span.End()

	logger.DebugContext(ctx, "Retrieving account by code",
		logger.Fields{"account_code": payload.AccountCode})

	account, err := h.financeServices.Account.GetAccountByCode(ctx, payload.AccountCode)
	if err != nil {
		logger.WarnContext(ctx, "Failed to retrieve account by code",
			logger.Fields{"account_code": payload.AccountCode, "error": err.Error()})
		return nil, h.handleError(err)
	}

	logger.DebugContext(ctx, "Account retrieved by code successfully",
		logger.Fields{
			"account_id":   account.ID.String(),
			"account_code": account.AccountCode,
		})

	return h.convertAccountToResult(account), nil
}

func (h *FinanceHandler) GetAccountByName(ctx context.Context, payload *goaFinance.GetAccountByNamePayload) (*goaFinance.AccountResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.get_account_by_name")
	defer span.End()

	logger.DebugContext(ctx, "Searching account by name",
		logger.Fields{"account_name": payload.AccountName})

	accounts, err := h.financeServices.Account.SearchAccounts(ctx, payload.AccountName, 1)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to search accounts by name",
			logger.Fields{"account_name": payload.AccountName, "error": err.Error()})
		return nil, h.handleError(err)
	}

	if len(accounts) == 0 {
		logger.WarnContext(ctx, "Account not found by name",
			logger.Fields{"account_name": payload.AccountName})
		return nil, goaFinance.MakeNotFound(sharedErrors.NewBusinessError("ACCOUNT_NOT_FOUND", "Account not found"))
	}

	for _, account := range accounts {
		if account.AccountName == payload.AccountName {
			logger.DebugContext(ctx, "Account found by name",
				logger.Fields{
					"account_id":   account.ID.String(),
					"account_name": account.AccountName,
				})
			return h.convertAccountToResult(account), nil
		}
	}

	logger.WarnContext(ctx, "No exact match found for account name",
		logger.Fields{"account_name": payload.AccountName})
	return nil, goaFinance.MakeNotFound(sharedErrors.NewBusinessError("ACCOUNT_NOT_FOUND", "Account not found"))
}

func (h *FinanceHandler) ListAccounts(ctx context.Context, payload *goaFinance.ListAccountsPayload) (*goaFinance.AccountListResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.list_accounts")
	defer span.End()

	logger.DebugContext(ctx, "Processing list accounts request",
		logger.Fields{
			"root_type":    payload.RootType,
			"account_type": payload.AccountType,
			"search":       payload.Search,
		})

	limitInt := int(payload.Limit)
	offsetInt := int(payload.Offset)
	limit := &limitInt
	offset := &offsetInt

	filter := &domain.AccountFilter{
		IsActive:   payload.IsActive,
		SearchTerm: payload.Search,
		Limit:      limit,
		Offset:     offset,
	}

	if payload.RootType != nil {
		rootType, err := domain.ParseRootType(*payload.RootType)
		if err != nil {
			logger.WarnContext(ctx, "Invalid root type in filter",
				logger.Fields{"root_type": *payload.RootType, "error": err.Error()})
			return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_ROOT_TYPE", "Invalid root type"))
		}
		filter.RootType = &rootType
	}

	// Note: AccountType filtering not implemented in domain.AccountFilter yet

	if payload.ParentID != nil {
		parentID, err := uuid.Parse(*payload.ParentID)
		if err != nil {
			logger.WarnContext(ctx, "Invalid parent ID format",
				logger.Fields{"parent_id": *payload.ParentID, "error": err.Error()})
			return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_PARENT_ID", "Invalid parent ID format"))
		}
		filter.ParentID = &parentID
	}

	accounts, err := h.financeServices.Account.ListAccounts(ctx, filter)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list accounts",
			logger.Fields{"error": err.Error()})
		return nil, h.handleError(err)
	}

	results := make([]*goaFinance.AccountResult, len(accounts))
	for i, account := range accounts {
		results[i] = h.convertAccountToResult(account)
	}

	logger.InfoContext(ctx, "Accounts listed successfully",
		logger.Fields{
			"total_found": len(results),
			"limit":       payload.Limit,
			"offset":      payload.Offset,
		})

	var resultLimit, resultOffset int32 = 50, 0
	resultLimit = payload.Limit
	resultOffset = payload.Offset

	return &goaFinance.AccountListResult{
		Accounts:   results,
		TotalCount: int64(len(results)),
		Limit:      resultLimit,
		Offset:     resultOffset,
	}, nil
}

func (h *FinanceHandler) UpdateAccount(ctx context.Context, payload *goaFinance.UpdateAccountPayload) (*goaFinance.AccountResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.update_account")
	defer span.End()

	id, err := uuid.Parse(payload.ID)
	if err != nil {
		logger.WarnContext(ctx, "Invalid account ID format for update",
			logger.Fields{"account_id": payload.ID, "error": err.Error()})
		return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_ID", "Invalid account ID format"))
	}

	logger.InfoContext(ctx, "Processing account update request",
		logger.Fields{
			"account_id":   id.String(),
			"account_name": payload.AccountName,
		})

	req := domain.UpdateAccountRequest{
		AccountName:        payload.AccountName,
		AccountDescription: payload.AccountDescription,
		IsActive:           payload.IsActive,
	}

	account, err := h.financeServices.Account.UpdateAccount(ctx, id, req)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update account",
			logger.Fields{
				"account_id": id.String(),
				"error":      err.Error(),
			})
		return nil, h.handleError(err)
	}

	logger.InfoContext(ctx, "Account updated successfully",
		logger.Fields{
			"account_id":   account.ID.String(),
			"account_code": account.AccountCode,
			"account_name": account.AccountName,
		})

	return h.convertAccountToResult(account), nil
}

func (h *FinanceHandler) DeleteAccount(ctx context.Context, payload *goaFinance.DeleteAccountPayload) error {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.delete_account")
	defer span.End()

	id, err := uuid.Parse(payload.ID)
	if err != nil {
		logger.WarnContext(ctx, "Invalid account ID format for deletion",
			logger.Fields{"account_id": payload.ID, "error": err.Error()})
		return goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_ID", "Invalid account ID format"))
	}

	logger.InfoContext(ctx, "Processing account deletion request",
		logger.Fields{"account_id": id.String()})

	err = h.financeServices.Account.DeleteAccount(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to delete account",
			logger.Fields{
				"account_id": id.String(),
				"error":      err.Error(),
			})
		return h.handleError(err)
	}

	logger.InfoContext(ctx, "Account deleted successfully",
		logger.Fields{"account_id": id.String()})

	return nil
}

func (h *FinanceHandler) GetAccountHierarchy(ctx context.Context, payload *goaFinance.GetAccountHierarchyPayload) (*goaFinance.AccountListResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.get_account_hierarchy")
	defer span.End()

	var rootID uuid.UUID
	if payload.RootID != nil {
		parsed, err := uuid.Parse(*payload.RootID)
		if err != nil {
			logger.WarnContext(ctx, "Invalid root ID format",
				logger.Fields{"root_id": *payload.RootID, "error": err.Error()})
			return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_ROOT_ID", "Invalid root ID format"))
		}
		rootID = parsed
	}

	logger.DebugContext(ctx, "Retrieving account hierarchy",
		logger.Fields{"root_id": rootID.String()})

	accounts, err := h.financeServices.Account.GetAccountHierarchy(ctx, rootID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to retrieve account hierarchy",
			logger.Fields{"root_id": rootID.String(), "error": err.Error()})
		return nil, h.handleError(err)
	}

	results := make([]*goaFinance.AccountResult, len(accounts))
	for i, account := range accounts {
		results[i] = h.convertAccountToResult(account)
	}

	logger.InfoContext(ctx, "Account hierarchy retrieved successfully",
		logger.Fields{
			"root_id":     rootID.String(),
			"total_count": len(results),
		})

	return &goaFinance.AccountHierarchyResult{
		Accounts:   results,
		TotalCount: int32(len(results)),
	}, nil
}

func (h *FinanceHandler) GetAccountBalance(ctx context.Context, payload *goaFinance.GetAccountBalancePayload) (*goaFinance.AccountBalanceResult, error) {
	ctx, span := h.tracing.StartSpan(ctx, "finance_handler.get_account_balance")
	defer span.End()

	accountID, err := uuid.Parse(payload.AccountID)
	if err != nil {
		logger.WarnContext(ctx, "Invalid account ID format for balance query",
			logger.Fields{"account_id": payload.AccountID, "error": err.Error()})
		return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_ACCOUNT_ID", "Invalid account ID format"))
	}

	var asOfDate time.Time
	if payload.AsOfDate != nil {
		parsed, err := time.Parse("2006-01-02", *payload.AsOfDate)
		if err != nil {
			logger.WarnContext(ctx, "Invalid as_of_date format",
				logger.Fields{"as_of_date": *payload.AsOfDate, "error": err.Error()})
			return nil, goaFinance.MakeBadRequest(sharedErrors.NewBusinessError("INVALID_DATE", "Invalid as_of_date format"))
		}
		asOfDate = parsed
	} else {
		asOfDate = time.Now()
	}

	logger.DebugContext(ctx, "Retrieving account balance",
		logger.Fields{
			"account_id": accountID.String(),
			"as_of_date": asOfDate.Format("2006-01-02"),
		})

	account, err := h.financeServices.Account.GetAccountByID(ctx, accountID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to retrieve account for balance",
			logger.Fields{
				"account_id": accountID.String(),
				"error":      err.Error(),
			})
		return nil, h.handleError(err)
	}

	logger.DebugContext(ctx, "Account balance retrieved successfully",
		logger.Fields{
			"account_id":      account.ID.String(),
			"current_balance": account.CurrentBalance.String(),
		})

	return &goaFinance.AccountBalanceResult{
		AccountID:           account.ID.String(),
		AccountCode:         account.AccountCode,
		AccountName:         account.AccountName,
		CurrentBalance:      account.CurrentBalance.String(),
		TotalDebits:         "0.00",
		TotalCredits:        "0.00",
		AsOfDate:            asOfDate.Format("2006-01-02"),
		LastTransactionDate: nil,
	}, nil
}

func (h *FinanceHandler) convertAccountToResult(account *domain.Accounts) *goaFinance.AccountResult {
	result := &goaFinance.AccountResult{
		ID:             account.ID.String(),
		AccountCode:    account.AccountCode,
		AccountName:    account.AccountName,
		RootType:       string(account.RootType),
		AccountType:    account.AccountType,
		NormalBalance:  string(account.NormalBalance),
		IsActive:       account.IsActive,
		CreatedAt:      stringPtr(account.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:      stringPtr(account.UpdatedAt.Format(time.RFC3339)),
		CurrentBalance: stringPtr(account.CurrentBalance.String()),
	}

	result.TenantID = stringPtr(account.TenantID.String())

	if account.EntityID != nil {
		result.EntityID = stringPtr(account.EntityID.String())
	}

	if account.AccountDescription != nil {
		result.AccountDescription = stringPtr(*account.AccountDescription)
	}

	// Hierarchy and grouping
	if account.ParentAccountID != nil {
		result.ParentAccountID = stringPtr(account.ParentAccountID.String())
	}

	result.AccountLevel = &account.AccountLevel

	if account.AccountPath != nil {
		result.AccountPath = stringPtr(*account.AccountPath)
	}

	// New hierarchy fields
	if account.AccountGroupID != nil {
		result.AccountGroupID = stringPtr(account.AccountGroupID.String())
	}

	if account.AccountHeaderID != nil {
		result.AccountHeaderID = stringPtr(account.AccountHeaderID.String())
	}

	result.HasChildren = &account.HasChildren
	result.IsLeafAccount = &account.IsLeafAccount

	// Classification fields
	if account.AccountSubtype != nil {
		result.AccountSubtype = stringPtr(*account.AccountSubtype)
	}

	if account.AccountCategory != nil {
		result.AccountCategory = stringPtr(*account.AccountCategory)
	}

	if account.SubCategory != nil {
		result.SubCategory = stringPtr(*account.SubCategory)
	}

	// Operational settings
	result.IsSystemAccount = &account.IsSystemAccount
	result.AllowManualEntries = &account.AllowManualEntries
	result.RequireReference = &account.RequireReference

	// Reporting fields
	if account.FinancialStatementLine != nil {
		result.FinancialStatementLine = stringPtr(*account.FinancialStatementLine)
	}

	result.ReportOrder = &account.ReportOrder
	result.DisplayOrder = &account.DisplayOrder
	result.ShowInReports = &account.ShowInReports

	if account.ConsolidationAccount != nil {
		result.ConsolidationAccount = stringPtr(*account.ConsolidationAccount)
	}

	if account.CashFlowType != nil {
		result.CashFlowType = stringPtr(*account.CashFlowType)
	}

	// Balance tracking
	result.YtdBalance = stringPtr(account.YTDBalance.String())

	if account.LastTransactionDate != nil {
		result.LastTransactionDate = stringPtr(account.LastTransactionDate.Format(time.RFC3339))
	}

	// Currency settings
	if account.CurrencyCode != nil {
		result.CurrencyCode = stringPtr(*account.CurrencyCode)
	}
	result.IsMultiCurrency = &account.IsMultiCurrency

	// Budgeting
	result.IsBudgetable = &account.IsBudgetable
	result.BudgetVarianceThreshold = stringPtr(account.BudgetVarianceThreshold.String())

	// Audit fields
	result.Version = &account.Version
	result.ValidationStatus = stringPtr(string(account.ValidationStatus))

	if account.LastValidationRun != nil {
		result.LastValidationRun = stringPtr(account.LastValidationRun.Format(time.RFC3339))
	}

	return result
}

// Helper function to convert string to *string
func stringPtr(s string) *string {
	return &s
}
