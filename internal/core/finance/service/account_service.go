package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

type AccountService interface {
	CreateAccount(ctx context.Context, req domain.CreateAccountRequest) (*domain.ChartOfAccounts, error)
	GetAccountByID(ctx context.Context, id uuid.UUID) (*domain.ChartOfAccounts, error)
	GetAccountByCode(ctx context.Context, code string) (*domain.ChartOfAccounts, error)
	UpdateAccount(ctx context.Context, id uuid.UUID, req domain.UpdateAccountRequest) (*domain.ChartOfAccounts, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
	ListAccounts(ctx context.Context, filter *domain.AccountFilter) ([]*domain.ChartOfAccounts, error)
	GetAccountHierarchy(ctx context.Context, rootAccountID uuid.UUID) ([]*domain.ChartOfAccounts, error)
	ValidateAccountCode(ctx context.Context, code string, excludeID *uuid.UUID) error
	GetAccountsByType(ctx context.Context, accountType string, rootType *domain.RootType) ([]*domain.ChartOfAccounts, error)
	GetActiveAccounts(ctx context.Context) ([]*domain.ChartOfAccounts, error)
	SearchAccounts(ctx context.Context, query string, limit int) ([]*domain.ChartOfAccounts, error)
	UpdateAccountBalance(ctx context.Context, accountID uuid.UUID, balance domain.AccountBalance) error
}

type accountService struct {
	repo    domain.ChartOfAccountsRepository
	tracing tracing.TracingService
	metrics metrics.MetricsProvider
}

func NewAccountService(repo domain.ChartOfAccountsRepository, tracing tracing.TracingService, metrics metrics.MetricsProvider) AccountService {
	return &accountService{
		repo:    repo,
		tracing: tracing,
		metrics: metrics,
	}
}

func (s *accountService) CreateAccount(ctx context.Context, req domain.CreateAccountRequest) (*domain.ChartOfAccounts, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.create_account",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.code", req.AccountCode),
			attribute.String("account.name", req.AccountName),
			attribute.String("account.type", string(req.AccountType)),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting account creation",
		logger.Fields{
			"account_code": req.AccountCode,
			"account_name": req.AccountName,
			"account_type": string(req.AccountType),
		})

	if err := req.Validate(); err != nil {
		s.metrics.IncrementCounter("account_creation_errors", metrics.Fields{
			"error_type": "validation_error",
		})

		logger.WarnContext(ctx, "Account validation failed",
			logger.Fields{
				"account_code": req.AccountCode,
				"errors":       len(err),
			})

		return nil, fmt.Errorf("account validation failed: %v", err)
	}

	if err := s.repo.ValidateAccountCode(ctx, req.AccountCode, nil); err != nil {
		s.metrics.IncrementCounter("account_creation_errors", metrics.Fields{
			"error_type": "duplicate_code",
		})

		logger.WarnContext(ctx, "Account code already exists",
			logger.Fields{"account_code": req.AccountCode})

		return nil, err
	}

	if req.ParentAccountID != nil {
		parentAccount, err := s.repo.GetByID(ctx, *req.ParentAccountID)
		if err != nil {
			s.metrics.IncrementCounter("account_creation_errors", metrics.Fields{
				"error_type": "invalid_parent",
			})

			logger.WarnContext(ctx, "Invalid parent account",
				logger.Fields{"parent_id": req.ParentAccountID.String()})

			return nil, fmt.Errorf("parent account not found: %w", err)
		}

		if !parentAccount.IsActive {
			return nil, fmt.Errorf("cannot create account under inactive parent")
		}
	}

	timer := s.metrics.Timer("account_creation_duration", metrics.Fields{
		"account_type": string(req.AccountType),
	})

	// Create domain account from request
	account := &domain.ChartOfAccounts{
		EntityID:                    req.EntityID,
		AccountCode:                 req.AccountCode,
		AccountName:                 req.AccountName,
		AccountDescription:          req.AccountDescription,
		ParentAccountID:             req.ParentAccountID,
		RootType:                    req.RootType,
		AccountType:                 req.AccountType,
		AccountSubtype:              req.AccountSubtype,
		NormalBalance:               req.NormalBalance,
		IsControlAccount:            req.IsControlAccount,
		ControlAccountID:            req.ControlAccountID,
		CurrencyCode:                req.CurrencyCode,
		IsMultiCurrency:             req.IsMultiCurrency,
		CurrencyRevaluationRequired: req.CurrencyRevaluationRequired,
		IsActive:                    req.IsActive,
		AllowManualEntries:          req.AllowManualEntries,
		RequireReference:            req.RequireReference,
		FinancialStatementLine:      req.FinancialStatementLine,
		ReportOrder:                 req.ReportOrder,
		IsBudgetable:                req.IsBudgetable,
		BudgetVarianceThreshold:     req.BudgetVarianceThreshold,
		AccountAttributes:           req.AccountAttributes,
	}

	err := s.repo.Create(ctx, account)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("account_creation_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to create account",
			logger.Fields{"error": err.Error()})

		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	// Retrieve the created account with generated ID
	createdAccount, err := s.repo.GetByCode(ctx, req.EntityID, req.AccountCode)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created account: %w", err)
	}

	s.metrics.IncrementCounter("accounts_created_total", metrics.Fields{
		"account_type": string(req.AccountType),
		"status":       "success",
	})

	s.metrics.ObserveHistogram("account_creation_duration",
		duration.Seconds(), metrics.Fields{
			"account_type": string(req.AccountType),
		})

	logger.InfoContext(ctx, "Account created successfully",
		logger.Fields{
			"account_id":   createdAccount.ID.String(),
			"account_code": createdAccount.AccountCode,
			"account_name": createdAccount.AccountName,
			"duration_ms":  duration.Milliseconds(),
		})

	return createdAccount, nil
}

func (s *accountService) GetAccountByID(ctx context.Context, id uuid.UUID) (*domain.ChartOfAccounts, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_by_id",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting account by ID",
		logger.Fields{"account_id": id.String()})

	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == errors.ErrNotFound {
			logger.WarnContext(ctx, "Account not found",
				logger.Fields{"account_id": id.String()})
			return nil, fmt.Errorf("account not found: %s", id.String())
		}

		logger.ErrorContext(ctx, "Failed to get account by ID",
			logger.Fields{
				"account_id": id.String(),
				"error":      err.Error(),
			})
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	logger.DebugContext(ctx, "Account retrieved successfully",
		logger.Fields{
			"account_id":   account.ID.String(),
			"account_code": account.AccountCode,
		})

	return account, nil
}

func (s *accountService) GetAccountByCode(ctx context.Context, code string) (*domain.ChartOfAccounts, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_by_code",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.code", code),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting account by code",
		logger.Fields{"account_code": code})

	// TODO: Get entityID from context or parameter
	account, err := s.repo.GetByCode(ctx, nil, code)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			logger.WarnContext(ctx, "Account not found",
				logger.Fields{"account_code": code})
			return nil, fmt.Errorf("account not found: %s", code)
		}

		logger.ErrorContext(ctx, "Failed to get account by code",
			logger.Fields{
				"account_code": code,
				"error":        err.Error(),
			})
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	logger.DebugContext(ctx, "Account retrieved successfully",
		logger.Fields{
			"account_id":   account.ID.String(),
			"account_code": account.AccountCode,
		})

	return account, nil
}

func (s *accountService) UpdateAccount(ctx context.Context, id uuid.UUID, req domain.UpdateAccountRequest) (*domain.ChartOfAccounts, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.update_account",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting account update",
		logger.Fields{"account_id": id.String()})

	existingAccount, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Account not found for update",
			logger.Fields{
				"account_id": id.String(),
				"error":      err.Error(),
			})
		return nil, err
	}

	if err := req.Validate(); err != nil {
		s.metrics.IncrementCounter("account_update_errors", metrics.Fields{
			"error_type": "validation_error",
		})

		logger.WarnContext(ctx, "Account update validation failed",
			logger.Fields{
				"account_id": id.String(),
				"errors":     len(err),
			})

		return nil, errors.NewBusinessError("VALIDATION_ERROR", "Account update validation failed")
	}

	if req.AccountCode != nil && *req.AccountCode != existingAccount.AccountCode {
		if err := s.repo.ValidateAccountCode(ctx, *req.AccountCode, &id); err != nil {
			s.metrics.IncrementCounter("account_update_errors", metrics.Fields{
				"error_type": "duplicate_code",
			})

			logger.WarnContext(ctx, "Account code already exists",
				logger.Fields{"account_code": *req.AccountCode})

			return nil, err
		}
		existingAccount.AccountCode = *req.AccountCode
	}

	if req.AccountName != nil {
		existingAccount.AccountName = *req.AccountName
	}
	if req.AccountDescription != nil {
		existingAccount.AccountDescription = req.AccountDescription
	}
	if req.AccountType != nil {
		existingAccount.AccountType = *req.AccountType
	}
	if req.AccountSubtype != nil {
		existingAccount.AccountSubtype = req.AccountSubtype
	}
	if req.IsActive != nil {
		existingAccount.IsActive = *req.IsActive
	}
	if req.AllowManualEntries != nil {
		existingAccount.AllowManualEntries = *req.AllowManualEntries
	}
	if req.RequireReference != nil {
		existingAccount.RequireReference = *req.RequireReference
	}
	if req.FinancialStatementLine != nil {
		existingAccount.FinancialStatementLine = req.FinancialStatementLine
	}
	if req.ReportOrder != nil {
		existingAccount.ReportOrder = *req.ReportOrder
	}
	if req.IsBudgetable != nil {
		existingAccount.IsBudgetable = *req.IsBudgetable
	}
	if req.BudgetVarianceThreshold != nil {
		existingAccount.BudgetVarianceThreshold = *req.BudgetVarianceThreshold
	}
	if req.AccountAttributes != nil {
		existingAccount.AccountAttributes = req.AccountAttributes
	}

	timer := s.metrics.Timer("account_update_duration", metrics.Fields{
		"account_type": string(existingAccount.AccountType),
	})

	err = s.repo.Update(ctx, existingAccount)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("account_update_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to update account",
			logger.Fields{
				"account_id": id.String(),
				"error":      err.Error(),
			})

		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	s.metrics.IncrementCounter("accounts_updated_total", metrics.Fields{
		"account_type": string(existingAccount.AccountType),
		"status":       "success",
	})

	s.metrics.ObserveHistogram("account_update_duration",
		duration.Seconds(), metrics.Fields{
			"account_type": string(existingAccount.AccountType),
		})

	logger.InfoContext(ctx, "Account updated successfully",
		logger.Fields{
			"account_id":   existingAccount.ID.String(),
			"account_code": existingAccount.AccountCode,
			"duration_ms":  duration.Milliseconds(),
		})

	return existingAccount, nil
}

func (s *accountService) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.delete_account",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting account deletion",
		logger.Fields{"account_id": id.String()})

	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Account not found for deletion",
			logger.Fields{
				"account_id": id.String(),
				"error":      err.Error(),
			})
		return err
	}

	if account.IsSystemAccount {
		s.metrics.IncrementCounter("account_deletion_errors", metrics.Fields{
			"error_type": "system_account",
		})

		logger.WarnContext(ctx, "Cannot delete system account",
			logger.Fields{"account_id": id.String()})

		return errors.NewBusinessError("SYSTEM_ACCOUNT", "Cannot delete system account")
	}

	hasTransactions, err := s.repo.HasTransactions(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check account transactions: %w", err)
	}

	if hasTransactions {
		s.metrics.IncrementCounter("account_deletion_errors", metrics.Fields{
			"error_type": "has_transactions",
		})

		logger.WarnContext(ctx, "Cannot delete account with transactions",
			logger.Fields{"account_id": id.String()})

		return errors.NewBusinessError("HAS_TRANSACTIONS", "Cannot delete account with existing transactions")
	}

	children, err := s.repo.GetChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check for child accounts: %w", err)
	}

	if len(children) > 0 {
		s.metrics.IncrementCounter("account_deletion_errors", metrics.Fields{
			"error_type": "has_children",
		})

		logger.WarnContext(ctx, "Cannot delete account with children",
			logger.Fields{
				"account_id":     id.String(),
				"children_count": len(children),
			})

		return errors.NewBusinessError("HAS_CHILDREN", "Cannot delete account with child accounts")
	}

	timer := s.metrics.Timer("account_deletion_duration", metrics.Fields{
		"account_type": string(account.AccountType),
	})

	err = s.repo.Delete(ctx, id)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("account_deletion_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to delete account",
			logger.Fields{
				"account_id": id.String(),
				"error":      err.Error(),
			})

		return fmt.Errorf("failed to delete account: %w", err)
	}

	s.metrics.IncrementCounter("accounts_deleted_total", metrics.Fields{
		"account_type": string(account.AccountType),
		"status":       "success",
	})

	s.metrics.ObserveHistogram("account_deletion_duration",
		duration.Seconds(), metrics.Fields{
			"account_type": string(account.AccountType),
		})

	logger.InfoContext(ctx, "Account deleted successfully",
		logger.Fields{
			"account_id":   id.String(),
			"account_code": account.AccountCode,
			"duration_ms":  duration.Milliseconds(),
		})

	return nil
}

func (s *accountService) ListAccounts(ctx context.Context, filter *domain.AccountFilter) ([]*domain.ChartOfAccounts, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.list_accounts",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.Int("limit", getIntValue(filter.Limit)),
			attribute.Int("offset", getIntValue(filter.Offset)),
		))
	defer span.End()

	logger.DebugContext(ctx, "Listing accounts",
		logger.Fields{
			"limit":  getIntValue(filter.Limit),
			"offset": getIntValue(filter.Offset),
		})

	if filter.Limit == nil || *filter.Limit <= 0 {
		limit := 50
		filter.Limit = &limit
	}

	if *filter.Limit > 1000 {
		limit := 1000
		filter.Limit = &limit
	}

	timer := s.metrics.Timer("account_list_duration", metrics.Fields{})

	accounts, err := s.repo.List(ctx, filter)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("account_list_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to list accounts",
			logger.Fields{"error": err.Error()})

		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	s.metrics.IncrementCounter("account_list_requests_total", metrics.Fields{
		"status": "success",
	})

	s.metrics.ObserveHistogram("account_list_duration",
		duration.Seconds(), metrics.Fields{})

	logger.DebugContext(ctx, "Accounts listed successfully",
		logger.Fields{
			"count":       len(accounts),
			"duration_ms": duration.Milliseconds(),
		})

	return accounts, nil
}

func (s *accountService) GetAccountHierarchy(ctx context.Context, rootAccountID uuid.UUID) ([]*domain.ChartOfAccounts, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_hierarchy",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.root_id", rootAccountID.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting account hierarchy",
		logger.Fields{"root_account_id": rootAccountID.String()})

	hierarchy, err := s.repo.GetAccountHierarchy(ctx, rootAccountID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get account hierarchy",
			logger.Fields{
				"root_account_id": rootAccountID.String(),
				"error":           err.Error(),
			})
		return nil, fmt.Errorf("failed to get account hierarchy: %w", err)
	}

	logger.DebugContext(ctx, "Account hierarchy retrieved successfully",
		logger.Fields{
			"root_account_id": rootAccountID.String(),
			"accounts_count":  len(hierarchy),
		})

	return hierarchy, nil
}

func (s *accountService) ValidateAccountCode(ctx context.Context, code string, excludeID *uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.validate_account_code",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.code", code),
		))
	defer span.End()

	return s.repo.ValidateAccountCode(ctx, code, excludeID)
}

func (s *accountService) GetAccountsByType(ctx context.Context, accountType string, rootType *domain.RootType) ([]*domain.ChartOfAccounts, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_accounts_by_type",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.type", string(accountType)),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting accounts by type",
		logger.Fields{"account_type": string(accountType)})

	accounts, err := s.repo.GetAccountsByType(ctx, accountType, rootType)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get accounts by type",
			logger.Fields{
				"account_type": string(accountType),
				"error":        err.Error(),
			})
		return nil, fmt.Errorf("failed to get accounts by type: %w", err)
	}

	logger.DebugContext(ctx, "Accounts by type retrieved successfully",
		logger.Fields{
			"account_type":   string(accountType),
			"accounts_count": len(accounts),
		})

	return accounts, nil
}

func (s *accountService) GetActiveAccounts(ctx context.Context) ([]*domain.ChartOfAccounts, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_active_accounts",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	logger.DebugContext(ctx, "Getting active accounts")

	accounts, err := s.repo.GetActiveAccounts(ctx, nil)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get active accounts",
			logger.Fields{"error": err.Error()})
		return nil, fmt.Errorf("failed to get active accounts: %w", err)
	}

	logger.DebugContext(ctx, "Active accounts retrieved successfully",
		logger.Fields{"accounts_count": len(accounts)})

	return accounts, nil
}

func (s *accountService) SearchAccounts(ctx context.Context, query string, limit int) ([]*domain.ChartOfAccounts, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.search_accounts",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("search.query", query),
			attribute.Int("search.limit", limit),
		))
	defer span.End()

	logger.DebugContext(ctx, "Searching accounts",
		logger.Fields{
			"query": query,
			"limit": limit,
		})

	if limit <= 0 {
		limit = 50
	}

	if limit > 200 {
		limit = 200
	}

	accounts, err := s.repo.Search(ctx, query, limit)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to search accounts",
			logger.Fields{
				"query": query,
				"error": err.Error(),
			})
		return nil, fmt.Errorf("failed to search accounts: %w", err)
	}

	logger.DebugContext(ctx, "Accounts search completed successfully",
		logger.Fields{
			"query":         query,
			"results_count": len(accounts),
		})

	return accounts, nil
}

func (s *accountService) UpdateAccountBalance(ctx context.Context, accountID uuid.UUID, balance domain.AccountBalance) error {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.update_account_balance",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.id", accountID.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Updating account balance",
		logger.Fields{
			"account_id":      accountID.String(),
			"current_balance": balance.NetBalance.String(),
			"ytd_balance":     balance.NetBalance.String(),
		})

	err := s.repo.UpdateBalance(ctx, accountID, balance)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update account balance",
			logger.Fields{
				"account_id": accountID.String(),
				"error":      err.Error(),
			})
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	logger.DebugContext(ctx, "Account balance updated successfully",
		logger.Fields{"account_id": accountID.String()})

	return nil
}

// Helper function
func getIntValue(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}
