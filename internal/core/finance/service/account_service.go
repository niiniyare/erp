package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"awo.so/internal/core/featureflag"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/iam"
	"awo.so/internal/shared"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

type AccountService interface {
	// Handler convenience methods (for backward compatibility)
	Create(ctx context.Context, req *domain.CreateAccountRequest) (*domain.Accounts, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Accounts, error)
	List(ctx context.Context, filter *domain.AccountFilter) ([]*domain.Accounts, error)
	Update(ctx context.Context, id uuid.UUID, req *domain.UpdateAccountRequest) (*domain.Accounts, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Core account operations
	CreateAccount(ctx context.Context, req domain.CreateAccountRequest) (*domain.Accounts, error)
	GetAccountByID(ctx context.Context, id uuid.UUID) (*domain.Accounts, error)
	GetAccountByCode(ctx context.Context, code string) (*domain.Accounts, error)
	UpdateAccount(ctx context.Context, id uuid.UUID, req domain.UpdateAccountRequest) (*domain.Accounts, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
	ListAccounts(ctx context.Context, filter *domain.AccountFilter) ([]*domain.Accounts, error)
	GetAccountHierarchy(ctx context.Context, rootAccountID uuid.UUID) ([]*domain.Accounts, error)
	ValidateAccountCode(ctx context.Context, code string, excludeID *uuid.UUID) error
	GetAccountsByType(ctx context.Context, accountType string, rootType *domain.RootType) ([]*domain.Accounts, error)
	GetActiveAccounts(ctx context.Context) ([]*domain.Accounts, error)
	SearchAccounts(ctx context.Context, query string, limit int) ([]*domain.Accounts, error)
	UpdateAccountBalance(ctx context.Context, accountID uuid.UUID, balance domain.AccountBalance) error

	// Account Groups Management
	CreateAccountGroup(ctx context.Context, req domain.CreateAccountGroupRequest) (*domain.AccountGroup, error)
	GetAccountGroupByID(ctx context.Context, id uuid.UUID) (*domain.AccountGroup, error)
	GetAccountGroupByCode(ctx context.Context, code string) (*domain.AccountGroup, error)
	UpdateAccountGroup(ctx context.Context, id uuid.UUID, req domain.UpdateAccountGroupRequest) (*domain.AccountGroup, error)
	DeleteAccountGroup(ctx context.Context, id uuid.UUID) error
	ListAccountGroups(ctx context.Context, filter *domain.AccountGroupFilter) ([]*domain.AccountGroup, error)
	GetAccountGroupHierarchy(ctx context.Context) ([]*domain.AccountGroup, error)
	GetGroupsByFinancialStatement(ctx context.Context, statementType string) ([]*domain.AccountGroup, error)
	GetGroupsByCashFlowCategory(ctx context.Context, category string) ([]*domain.AccountGroup, error)

	// Unified Operations
	ListAccountsAndGroups(ctx context.Context, filter *domain.UnifiedFilter) ([]*domain.AccountNode, error)
	GetAccountHierarchyWithGroups(ctx context.Context) ([]*domain.AccountNode, error)
	SearchAccountsAndGroups(ctx context.Context, query string, limit int) ([]*domain.AccountNode, error)

	// Enhanced view-based operations
	GetAccountWithGroups(ctx context.Context, id uuid.UUID) (*domain.AccountWithGroups, error)
	GetAccountWithGroupsByCode(ctx context.Context, code string) (*domain.AccountWithGroups, error)
	ListAccountsWithGroups(ctx context.Context, filter *domain.AccountFilter) ([]*domain.AccountWithGroups, error)
	SearchAccountsWithGroups(ctx context.Context, query string, limit int) ([]*domain.AccountWithGroups, error)
	GetLeafAccountsOnly(ctx context.Context, rootType *string) ([]*domain.AccountWithGroups, error)

	// Complete chart of accounts operations
	GetCompleteChartOfAccounts(ctx context.Context, filter *domain.ChartOfAccountsFilter) ([]*domain.ChartOfAccountsComplete, error)
	GetAccountForReporting(ctx context.Context, accountID uuid.UUID) (*domain.ChartOfAccountsComplete, error)
	GetAccountsByStatementSection(ctx context.Context, section string, entityID *uuid.UUID) ([]*domain.ChartOfAccountsComplete, error)
	GetAccountsByGroup(ctx context.Context, groupCode string, entityID *uuid.UUID) ([]*domain.ChartOfAccountsComplete, error)
	GetAccountsByHeader(ctx context.Context, headerCode string, entityID *uuid.UUID) ([]*domain.ChartOfAccountsComplete, error)

	// Financial reporting operations
	GetTrialBalanceAccounts(ctx context.Context, entityID *uuid.UUID, nonZeroOnly bool) ([]*domain.TrialBalanceSummary, error)
	GetAccountsWithBalances(ctx context.Context, filter *domain.BalanceFilter) ([]*domain.ChartOfAccountsComplete, error)
	GetCashFlowAccounts(ctx context.Context, entityID *uuid.UUID) ([]*domain.CashFlowAccount, error)
	GetAccountSummaryByGroup(ctx context.Context, entityID *uuid.UUID) ([]*domain.AccountGroupSummary, error)

	// Enhanced analytics and hierarchy operations
	GetAccountChildrenHierarchy(ctx context.Context, parentAccountID uuid.UUID) ([]*domain.AccountHierarchy, error)
	GetAccountSubtree(ctx context.Context, accountID uuid.UUID) ([]*domain.AccountHierarchy, error)
	GetAccountsWithRecentActivity(ctx context.Context, filter *domain.AccountActivityFilter) ([]*domain.AccountActivity, error)
	GetStaleAccountBalances(ctx context.Context, filter *domain.AccountActivityFilter) ([]*domain.AccountActivity, error)
	GetAccountActivitySummary(ctx context.Context, filter *domain.AccountActivityFilter) ([]*domain.AccountActivitySummary, error)
}

type accountService struct {
	accountRepo        domain.AccountsRepository
	accountGroupRepo   domain.AccountGroupRepository
	tracing            tracing.Service
	metrics            metrics.MetricsProvider
	settingsHelper     *SettingsHelper
	iamService         iam.Service
	featureFlagService featureflag.Service
}

func NewAccountService(
	accountRepo domain.AccountsRepository,
	accountGroupRepo domain.AccountGroupRepository,
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
	iamService iam.Service,
	featureFlagService featureflag.Service,
) AccountService {
	return &accountService{
		accountRepo:        accountRepo,
		accountGroupRepo:   accountGroupRepo,
		tracing:            tracing,
		metrics:            metrics,
		settingsHelper:     NewSettingsHelper(),
		iamService:         iamService,
		featureFlagService: featureFlagService,
	}
}

// Handler convenience methods (delegate to main methods)
func (s *accountService) Create(ctx context.Context, req *domain.CreateAccountRequest) (*domain.Accounts, error) {
	if req == nil {
		return nil, errors.NewBusinessError("INVALID_REQUEST", "request cannot be nil")
	}
	return s.CreateAccount(ctx, *req)
}

func (s *accountService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Accounts, error) {
	return s.GetAccountByID(ctx, id)
}

func (s *accountService) List(ctx context.Context, filter *domain.AccountFilter) ([]*domain.Accounts, error) {
	return s.ListAccounts(ctx, filter)
}

func (s *accountService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateAccountRequest) (*domain.Accounts, error) {
	if req == nil {
		return nil, errors.NewBusinessError("INVALID_REQUEST", "request cannot be nil")
	}
	return s.UpdateAccount(ctx, id, *req)
}

func (s *accountService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.DeleteAccount(ctx, id)
}

func (s *accountService) CreateAccount(ctx context.Context, req domain.CreateAccountRequest) (*domain.Accounts, error) {
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

	// TODO(authz): enforce finance.accounts.create via iam.Service.Enforce() once
	// the session principal is wired into ctx. Example:
	//   if userID, ok := shared.GetUserID(ctx); ok {
	//       ok, _ := s.iamService.Enforce(ctx, iam.Request{Subject: ..., Object: "finance.accounts", Action: "create"})
	//       if !ok { return nil, errors.NewBusinessError("UNAUTHORIZED", "Cannot create account") }
	//   }

	// Settings-driven defaults - apply if not specified
	if req.CurrencyCode == nil || *req.CurrencyCode == "" {
		currency := domain.DefaultBaseCurrency
		req.CurrencyCode = &currency // Uses existing constant
	}

	// Feature Flag - Use enhanced validation if enabled
	useEnhancedValidation := false
	tenantID, _ := shared.GetTenantID(ctx)
	userID, _ := shared.GetUserID(ctx)

	evalCtx := &featureflag.EvaluationContext{
		TenantID:    tenantID,
		UserID:      &userID,
		Environment: "production", // TODO: Get from config
		Attributes: map[string]string{
			"module":        "finance",
			"resource_type": "account",
		},
	}

	if evalResult, err := s.featureFlagService.IsEnabled(ctx, "enhanced_account_validation", evalCtx); err == nil {
		useEnhancedValidation = evalResult
		logger.DebugContext(ctx, "Feature flag evaluated", logger.Fields{
			"flag":    "enhanced_account_validation",
			"enabled": useEnhancedValidation,
		})
	}

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

	// Enhanced validation if feature flag is enabled
	if useEnhancedValidation {
		// Additional validation rules when enhanced validation is enabled
		if len(req.AccountCode) < 3 {
			return nil, errors.NewBusinessError("VALIDATION_ERROR", "Account code must be at least 3 characters when enhanced validation is enabled")
		}
		logger.DebugContext(ctx, "Using enhanced account validation", logger.Fields{"account_code": req.AccountCode})
	}

	// Settings-driven account code validation
	// TODO: Get account code length setting when Settings service is available
	// accountCodeLength := s.settingsService.GetEffectiveConfiguration(ctx, req.EntityID, "finance", "account_code_length")
	// if len(req.AccountCode) > accountCodeLength { ... }

	if err := s.accountRepo.ValidateAccountCode(ctx, req.AccountCode, nil); err != nil {
		s.metrics.IncrementCounter("account_creation_errors", metrics.Fields{
			"error_type": "duplicate_code",
		})

		logger.WarnContext(ctx, "Account code already exists",
			logger.Fields{"account_code": req.AccountCode})

		return nil, err
	}

	if req.ParentAccountID != nil {
		parentAccount, err := s.accountRepo.GetByID(ctx, *req.ParentAccountID)
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
	account := &domain.Accounts{
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

	err := s.accountRepo.Create(ctx, account)
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
	createdAccount, err := s.accountRepo.GetByCode(ctx, req.EntityID, req.AccountCode)
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

func (s *accountService) GetAccountByID(ctx context.Context, id uuid.UUID) (*domain.Accounts, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_by_id",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting account by ID",
		logger.Fields{"account_id": id.String()})

	account, err := s.accountRepo.GetByID(ctx, id)
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

func (s *accountService) GetAccountByCode(ctx context.Context, code string) (*domain.Accounts, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_by_code",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.code", code),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting account by code",
		logger.Fields{"account_code": code})

	// TODO(authz): enforce finance.accounts.read via iam.Service.Enforce() once
	// the session principal is wired into ctx.

	// TODO: Get entityID from context or parameter
	account, err := s.accountRepo.GetByCode(ctx, nil, code)
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

func (s *accountService) UpdateAccount(ctx context.Context, id uuid.UUID, req domain.UpdateAccountRequest) (*domain.Accounts, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.update_account",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting account update",
		logger.Fields{"account_id": id.String()})

	// TODO(authz): enforce finance.accounts.update via iam.Service.Enforce() once
	// the session principal is wired into ctx.

	existingAccount, err := s.accountRepo.GetByID(ctx, id)
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
		if err := s.accountRepo.ValidateAccountCode(ctx, *req.AccountCode, &id); err != nil {
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

	err = s.accountRepo.Update(ctx, existingAccount)
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

	// TODO(authz): enforce finance.accounts.delete via iam.Service.Enforce() once
	// the session principal is wired into ctx.

	// Feature Flag - Check if enhanced deletion checks are enabled
	tenantID, _ := shared.GetTenantID(ctx)
	userID, _ := shared.GetUserID(ctx)

	evalCtx := &featureflag.EvaluationContext{
		TenantID:    tenantID,
		UserID:      &userID,
		Environment: "production",
		Attributes: map[string]string{
			"module":    "finance",
			"operation": "delete",
		},
	}

	useEnhancedDeletionChecks := false
	if evalResult, err := s.featureFlagService.IsEnabled(ctx, "enhanced_account_deletion", evalCtx); err == nil {
		useEnhancedDeletionChecks = evalResult
	}

	account, err := s.accountRepo.GetByID(ctx, id)
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

	hasTransactions, err := s.accountRepo.HasTransactions(ctx, id)
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

	children, err := s.accountRepo.GetChildren(ctx, id)
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

	// Enhanced deletion checks if feature flag is enabled
	if useEnhancedDeletionChecks {
		// Additional checks when enhanced deletion is enabled
		logger.InfoContext(ctx, "Performing enhanced deletion checks", logger.Fields{"account_id": id.String()})

		// Example: Check for pending reconciliations, scheduled transactions, etc.
		// This would be additional business logic when the feature is enabled
	}

	timer := s.metrics.Timer("account_deletion_duration", metrics.Fields{
		"account_type": string(account.AccountType),
	})

	err = s.accountRepo.Delete(ctx, id)
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

func (s *accountService) ListAccounts(ctx context.Context, filter *domain.AccountFilter) ([]*domain.Accounts, error) {
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

	// TODO(authz): filter accounts based on user permissions via iam.Service.Enforce() once
	// the session principal is wired into ctx.

	// Feature Flag - Use enhanced listing if enabled
	tenantID, _ := shared.GetTenantID(ctx)
	userID, _ := shared.GetUserID(ctx)

	evalCtx := &featureflag.EvaluationContext{
		TenantID:    tenantID,
		UserID:      &userID,
		Environment: "production",
		Attributes: map[string]string{
			"module":    "finance",
			"operation": "list",
		},
	}

	useEnhancedListing := false
	if evalResult, err := s.featureFlagService.IsEnabled(ctx, "enhanced_account_listing", evalCtx); err == nil {
		useEnhancedListing = evalResult
		if useEnhancedListing {
			logger.DebugContext(ctx, "Using enhanced account listing", logger.Fields{"enhanced": true})
		}
	}

	if filter.Limit == nil || *filter.Limit <= 0 {
		limit := 50
		filter.Limit = &limit
	}

	if *filter.Limit > 1000 {
		limit := 1000
		filter.Limit = &limit
	}

	timer := s.metrics.Timer("account_list_duration", metrics.Fields{})

	accounts, err := s.accountRepo.List(ctx, filter)
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

func (s *accountService) GetAccountHierarchy(ctx context.Context, rootAccountID uuid.UUID) ([]*domain.Accounts, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_hierarchy",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.root_id", rootAccountID.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting account hierarchy",
		logger.Fields{"root_account_id": rootAccountID.String()})

	hierarchy, err := s.accountRepo.GetAccountHierarchy(ctx, rootAccountID)
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

	return s.accountRepo.ValidateAccountCode(ctx, code, excludeID)
}

func (s *accountService) GetAccountsByType(ctx context.Context, accountType string, rootType *domain.RootType) ([]*domain.Accounts, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_accounts_by_type",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.type", string(accountType)),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting accounts by type",
		logger.Fields{"account_type": string(accountType)})

	accounts, err := s.accountRepo.GetAccountsByType(ctx, accountType, rootType)
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

func (s *accountService) GetActiveAccounts(ctx context.Context) ([]*domain.Accounts, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_active_accounts",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	logger.DebugContext(ctx, "Getting active accounts")

	accounts, err := s.accountRepo.GetActiveAccounts(ctx, nil)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get active accounts",
			logger.Fields{"error": err.Error()})
		return nil, fmt.Errorf("failed to get active accounts: %w", err)
	}

	logger.DebugContext(ctx, "Active accounts retrieved successfully",
		logger.Fields{"accounts_count": len(accounts)})

	return accounts, nil
}

func (s *accountService) SearchAccounts(ctx context.Context, query string, limit int) ([]*domain.Accounts, error) {
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

	// TODO(authz): enforce finance.accounts.search via iam.Service.Enforce() once
	// the session principal is wired into ctx.

	// Feature Flag - Enhanced search capabilities
	tenantID, _ := shared.GetTenantID(ctx)
	userID, _ := shared.GetUserID(ctx)

	evalCtx := &featureflag.EvaluationContext{
		TenantID:    tenantID,
		UserID:      &userID,
		Environment: "production",
		Attributes: map[string]string{
			"module":    "finance",
			"operation": "search",
		},
	}

	useEnhancedSearch := false
	if evalResult, err := s.featureFlagService.IsEnabled(ctx, "enhanced_account_search", evalCtx); err == nil {
		useEnhancedSearch = evalResult
		if useEnhancedSearch {
			logger.DebugContext(ctx, "Using enhanced account search", logger.Fields{"enhanced": true})
		}
	}

	// Settings-driven limit configuration
	if limit <= 0 {
		// Default from settings helper
		limit = 50
	}

	maxLimit := 200
	if useEnhancedSearch {
		// Enhanced search allows more results
		maxLimit = 500
	}

	if limit > maxLimit {
		limit = maxLimit
	}

	accounts, err := s.accountRepo.Search(ctx, query, limit)
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

	err := s.accountRepo.UpdateBalance(ctx, accountID, balance)
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

// Enhanced view-based operations
func (s *accountService) GetAccountWithGroups(ctx context.Context, id uuid.UUID) (*domain.AccountWithGroups, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_with_groups",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting account with groups by ID",
		logger.Fields{"account_id": id.String()})

	// TODO: Implement repository method GetAccountWithGroups
	// This should use the GetAccountWithGroupsByID query
	account, err := s.accountRepo.GetAccountWithGroups(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get account with groups",
			logger.Fields{
				"account_id": id.String(),
				"error":      err.Error(),
			})
		return nil, fmt.Errorf("failed to get account with groups: %w", err)
	}

	logger.DebugContext(ctx, "Account with groups retrieved successfully",
		logger.Fields{
			"account_id":   account.ID.String(),
			"account_code": account.AccountCode,
			"group_name":   getStringValue(account.GroupName),
		})

	return account, nil
}

func (s *accountService) GetAccountWithGroupsByCode(ctx context.Context, code string) (*domain.AccountWithGroups, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_with_groups_by_code",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.code", code),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting account with groups by code",
		logger.Fields{"account_code": code})

	// TODO: Implement repository method GetAccountWithGroupsByCode
	account, err := s.accountRepo.GetAccountWithGroupsByCode(ctx, code)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get account with groups by code",
			logger.Fields{
				"account_code": code,
				"error":        err.Error(),
			})
		return nil, fmt.Errorf("failed to get account with groups by code: %w", err)
	}

	return account, nil
}

func (s *accountService) ListAccountsWithGroups(ctx context.Context, filter *domain.AccountFilter) ([]*domain.AccountWithGroups, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.list_accounts_with_groups",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	logger.DebugContext(ctx, "Listing accounts with groups")

	if filter.Limit == nil || *filter.Limit <= 0 {
		limit := 50
		filter.Limit = &limit
	}

	if *filter.Limit > 1000 {
		limit := 1000
		filter.Limit = &limit
	}

	// TODO: Implement repository method ListAccountsWithGroups
	accounts, err := s.accountRepo.ListAccountsWithGroups(ctx, filter)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list accounts with groups",
			logger.Fields{"error": err.Error()})
		return nil, fmt.Errorf("failed to list accounts with groups: %w", err)
	}

	logger.DebugContext(ctx, "Accounts with groups listed successfully",
		logger.Fields{"count": len(accounts)})

	return accounts, nil
}

func (s *accountService) SearchAccountsWithGroups(ctx context.Context, query string, limit int) ([]*domain.AccountWithGroups, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.search_accounts_with_groups",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("search.query", query),
			attribute.Int("search.limit", limit),
		))
	defer span.End()

	logger.DebugContext(ctx, "Searching accounts with groups",
		logger.Fields{"query": query, "limit": limit})

	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	// TODO: Implement repository method SearchAccountsWithGroups
	accounts, err := s.accountRepo.SearchAccountsWithGroups(ctx, query, limit)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to search accounts with groups",
			logger.Fields{"query": query, "error": err.Error()})
		return nil, fmt.Errorf("failed to search accounts with groups: %w", err)
	}

	return accounts, nil
}

func (s *accountService) GetLeafAccountsOnly(ctx context.Context, rootType *string) ([]*domain.AccountWithGroups, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_leaf_accounts_only",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	logger.DebugContext(ctx, "Getting leaf accounts only",
		logger.Fields{"root_type": getStringValue(rootType)})

	// TODO: Implement repository method GetLeafAccountsOnly
	accounts, err := s.accountRepo.GetLeafAccountsOnly(ctx, rootType)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get leaf accounts",
			logger.Fields{"error": err.Error()})
		return nil, fmt.Errorf("failed to get leaf accounts: %w", err)
	}

	return accounts, nil
}

// Complete chart of accounts operations
func (s *accountService) GetCompleteChartOfAccounts(ctx context.Context, filter *domain.ChartOfAccountsFilter) ([]*domain.ChartOfAccountsComplete, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_complete_chart_of_accounts",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	logger.DebugContext(ctx, "Getting complete chart of accounts")

	if err := filter.Validate(); err != nil && len(err) > 0 {
		logger.WarnContext(ctx, "Invalid chart of accounts filter",
			logger.Fields{"errors": len(err)})
		return nil, fmt.Errorf("invalid filter: %v", err)
	}

	if filter.Limit == nil || *filter.Limit <= 0 {
		limit := 100
		filter.Limit = &limit
	}

	// TODO: Implement repository method GetCompleteChartOfAccounts
	accounts, err := s.accountRepo.GetCompleteChartOfAccounts(ctx, filter)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get complete chart of accounts",
			logger.Fields{"error": err.Error()})
		return nil, fmt.Errorf("failed to get complete chart of accounts: %w", err)
	}

	return accounts, nil
}

func (s *accountService) GetAccountForReporting(ctx context.Context, accountID uuid.UUID) (*domain.ChartOfAccountsComplete, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_for_reporting",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.id", accountID.String()),
		))
	defer span.End()

	// TODO: Implement repository method GetAccountForReporting
	account, err := s.accountRepo.GetAccountForReporting(ctx, accountID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get account for reporting",
			logger.Fields{"account_id": accountID.String(), "error": err.Error()})
		return nil, fmt.Errorf("failed to get account for reporting: %w", err)
	}

	return account, nil
}

func (s *accountService) GetAccountsByStatementSection(ctx context.Context, section string, entityID *uuid.UUID) ([]*domain.ChartOfAccountsComplete, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_accounts_by_statement_section",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("statement.section", section),
		))
	defer span.End()

	// TODO: Implement repository method GetAccountsByStatementSection
	accounts, err := s.accountRepo.GetAccountsByStatementSection(ctx, section, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts by statement section: %w", err)
	}

	return accounts, nil
}

func (s *accountService) GetAccountsByGroup(ctx context.Context, groupCode string, entityID *uuid.UUID) ([]*domain.ChartOfAccountsComplete, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_accounts_by_group",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("group.code", groupCode),
		))
	defer span.End()

	// TODO: Implement repository method GetAccountsByGroup
	accounts, err := s.accountRepo.GetAccountsByGroup(ctx, groupCode, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts by group: %w", err)
	}

	return accounts, nil
}

func (s *accountService) GetAccountsByHeader(ctx context.Context, headerCode string, entityID *uuid.UUID) ([]*domain.ChartOfAccountsComplete, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_accounts_by_header",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("header.code", headerCode),
		))
	defer span.End()

	// TODO: Implement repository method GetAccountsByHeader
	accounts, err := s.accountRepo.GetAccountsByHeader(ctx, headerCode, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts by header: %w", err)
	}

	return accounts, nil
}

// Financial reporting operations
func (s *accountService) GetTrialBalanceAccounts(ctx context.Context, entityID *uuid.UUID, nonZeroOnly bool) ([]*domain.TrialBalanceSummary, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_trial_balance_accounts",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	logger.DebugContext(ctx, "Getting trial balance accounts",
		logger.Fields{"non_zero_only": nonZeroOnly})

	// TODO: Implement repository method GetTrialBalanceAccounts
	accounts, err := s.accountRepo.GetTrialBalanceAccounts(ctx, entityID, nonZeroOnly)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get trial balance accounts",
			logger.Fields{"error": err.Error()})
		return nil, fmt.Errorf("failed to get trial balance accounts: %w", err)
	}

	return accounts, nil
}

func (s *accountService) GetAccountsWithBalances(ctx context.Context, filter *domain.BalanceFilter) ([]*domain.ChartOfAccountsComplete, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_accounts_with_balances",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	if err := filter.Validate(); err != nil && len(err) > 0 {
		logger.WarnContext(ctx, "Invalid balance filter",
			logger.Fields{"errors": len(err)})
		return nil, fmt.Errorf("invalid filter: %v", err)
	}

	// TODO: Implement repository method GetAccountsWithBalances
	accounts, err := s.accountRepo.GetAccountsWithBalances(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts with balances: %w", err)
	}

	return accounts, nil
}

func (s *accountService) GetCashFlowAccounts(ctx context.Context, entityID *uuid.UUID) ([]*domain.CashFlowAccount, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_cash_flow_accounts",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	logger.DebugContext(ctx, "Getting cash flow accounts")

	// TODO: Implement repository method GetCashFlowAccounts
	accounts, err := s.accountRepo.GetCashFlowAccounts(ctx, entityID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get cash flow accounts",
			logger.Fields{"error": err.Error()})
		return nil, fmt.Errorf("failed to get cash flow accounts: %w", err)
	}

	return accounts, nil
}

func (s *accountService) GetAccountSummaryByGroup(ctx context.Context, entityID *uuid.UUID) ([]*domain.AccountGroupSummary, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_summary_by_group",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	logger.DebugContext(ctx, "Getting account summary by group")

	// TODO: Implement repository method GetAccountSummaryByGroup
	summary, err := s.accountRepo.GetAccountSummaryByGroup(ctx, entityID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get account summary by group",
			logger.Fields{"error": err.Error()})
		return nil, fmt.Errorf("failed to get account summary by group: %w", err)
	}

	return summary, nil
}

// Helper functions
func getIntValue(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Settings integration helper functions
func (s *accountService) generateAccountCode(ctx context.Context, entityID *uuid.UUID, accountType string) (string, error) {
	// Settings-driven account code generation
	// TODO: When Settings service is available, get numbering scheme and format
	// scheme := s.settingsService.GetEffectiveConfiguration(ctx, entityID, "finance", "account_numbering_scheme")
	// format := s.settingsService.GetEffectiveConfiguration(ctx, entityID, "finance", "account_code_format")

	// For now, use constants from settings helper
	if accountType == domain.AccountTypeCurrentAsset {
		// Asset accounts start with 1
		return "1000", nil
	} else if accountType == domain.AccountTypeCurrentLiability {
		// Liability accounts start with 2
		return "2000", nil
	}

	// Default fallback
	return "9000", nil
}

// Business logic integration examples
func (s *accountService) shouldRequireApproval(ctx context.Context, entityID *uuid.UUID, operation string) bool {
	// Settings-driven approval requirements
	// TODO: When Settings service is available
	// approvalRequired := s.settingsService.GetEffectiveConfiguration(ctx, entityID, "finance", "require_account_approval")

	// For now, use default behavior
	return operation == "delete" // Only require approval for deletions
}

func (s *accountService) isEnhancedFeatureEnabled(ctx context.Context, featureName string, entityID *uuid.UUID) bool {
	// Feature flag evaluation with proper context
	tenantID, _ := shared.GetTenantID(ctx)
	userID, _ := shared.GetUserID(ctx)

	evalCtx := &featureflag.EvaluationContext{
		TenantID:    tenantID,
		UserID:      &userID,
		Environment: "production",
		Attributes: map[string]string{
			"module":    "finance",
			"entity_id": entityID.String(),
		},
	}

	if evalResult, err := s.featureFlagService.IsEnabled(ctx, featureName, evalCtx); err == nil {
		return evalResult
	}
	return false
}

// ============================================================================
// Account Groups Management Methods
// ============================================================================

func (s *accountService) CreateAccountGroup(ctx context.Context, req domain.CreateAccountGroupRequest) (*domain.AccountGroup, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.create_account_group",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("group.code", req.GroupCode),
			attribute.String("group.name", req.GroupName),
			attribute.String("group.type", req.GroupType),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting account group creation",
		logger.Fields{
			"group_code": req.GroupCode,
			"group_name": req.GroupName,
			"group_type": req.GroupType,
		})

	// TODO(authz): enforce finance.account_groups.create via iam.Service.Enforce() once
	// the session principal is wired into ctx.

	// Validate account group code uniqueness
	if err := s.accountGroupRepo.ValidateGroupCode(ctx, req.GroupCode, nil); err != nil {
		return nil, fmt.Errorf("account group validation failed: %w", err)
	}

	// Validate parent group if specified
	if req.ParentGroupID != nil {
		parentGroup, err := s.accountGroupRepo.GetByID(ctx, *req.ParentGroupID)
		if err != nil {
			return nil, fmt.Errorf("parent group validation failed: %w", err)
		}
		if !parentGroup.IsActive {
			return nil, errors.NewBusinessError("VALIDATION_ERROR", "Parent group must be active")
		}
	}

	// Get current user for audit
	userID, _ := shared.GetUserID(ctx)

	// Create account group entity
	group := &domain.AccountGroup{
		ID:                        uuid.New(),
		EntityID:                  req.EntityID,
		GroupCode:                 req.GroupCode,
		GroupName:                 req.GroupName,
		Description:               req.Description,
		GroupType:                 req.GroupType,
		ParentGroupID:             req.ParentGroupID,
		FinancialStatementSection: req.FinancialStatementSection,
		ConsolidationMethod:       req.ConsolidationMethod,
		CashFlowCategory:          req.CashFlowCategory,
		DisplayOrder:              req.DisplayOrder,
		IsSystemDefined:           false, // User-created groups are not system defined
		IsActive:                  true,
		IsHeader:                  req.IsHeader,
		ShowTotals:                req.ShowTotals,
		IndentLevel:               req.IndentLevel,
		BoldDisplay:               req.BoldDisplay,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
		CreatedBy:                 userID,
		UpdatedBy:                 userID,
		Version:                   1,
	}

	// Validate business rules
	if err := group.Validate(); err != nil {
		return nil, fmt.Errorf("account group validation failed: %w", err)
	}

	// Create in repository
	err := s.accountGroupRepo.Create(ctx, group)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create account group",
			logger.Fields{
				"group_code": req.GroupCode,
				"error":      err.Error(),
			})
		return nil, fmt.Errorf("failed to create account group: %w", err)
	}

	// Retrieve the created group with computed fields
	createdGroup, err := s.accountGroupRepo.GetByCode(ctx, req.EntityID, req.GroupCode)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to retrieve created account group",
			logger.Fields{
				"group_code": req.GroupCode,
				"error":      err.Error(),
			})
		return nil, fmt.Errorf("failed to retrieve created account group: %w", err)
	}

	logger.InfoContext(ctx, "Account group created successfully",
		logger.Fields{
			"group_id":   createdGroup.ID.String(),
			"group_code": createdGroup.GroupCode,
			"group_name": createdGroup.GroupName,
		})

	return createdGroup, nil
}

func (s *accountService) GetAccountGroupByID(ctx context.Context, id uuid.UUID) (*domain.AccountGroup, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_group_by_id",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("group.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting account group by ID",
		logger.Fields{"group_id": id.String()})

	group, err := s.accountGroupRepo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get account group",
			logger.Fields{
				"group_id": id.String(),
				"error":    err.Error(),
			})
		return nil, fmt.Errorf("failed to get account group: %w", err)
	}

	logger.DebugContext(ctx, "Account group retrieved successfully",
		logger.Fields{
			"group_id":   group.ID.String(),
			"group_code": group.GroupCode,
			"group_name": group.GroupName,
		})

	return group, nil
}

func (s *accountService) GetAccountGroupByCode(ctx context.Context, code string) (*domain.AccountGroup, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_group_by_code",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("group.code", code),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting account group by code",
		logger.Fields{"group_code": code})

	group, err := s.accountGroupRepo.GetByCode(ctx, nil, code)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get account group by code",
			logger.Fields{
				"group_code": code,
				"error":      err.Error(),
			})
		return nil, fmt.Errorf("failed to get account group by code: %w", err)
	}

	logger.DebugContext(ctx, "Account group retrieved successfully",
		logger.Fields{
			"group_id":   group.ID.String(),
			"group_code": group.GroupCode,
			"group_name": group.GroupName,
		})

	return group, nil
}

func (s *accountService) UpdateAccountGroup(ctx context.Context, id uuid.UUID, req domain.UpdateAccountGroupRequest) (*domain.AccountGroup, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.update_account_group",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("group.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting account group update",
		logger.Fields{"group_id": id.String()})

	// Get existing group
	existingGroup, err := s.accountGroupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing account group: %w", err)
	}

	// TODO(authz): enforce finance.account_groups.update via iam.Service.Enforce() once
	// the session principal is wired into ctx.

	// Apply updates
	if req.GroupName != nil {
		existingGroup.GroupName = *req.GroupName
	}
	if req.Description != nil {
		existingGroup.Description = req.Description
	}
	if req.GroupType != nil {
		existingGroup.GroupType = *req.GroupType
	}
	if req.ParentGroupID != nil {
		existingGroup.ParentGroupID = req.ParentGroupID
	}
	if req.FinancialStatementSection != nil {
		existingGroup.FinancialStatementSection = req.FinancialStatementSection
	}
	if req.ConsolidationMethod != nil {
		existingGroup.ConsolidationMethod = req.ConsolidationMethod
	}
	if req.CashFlowCategory != nil {
		existingGroup.CashFlowCategory = req.CashFlowCategory
	}
	if req.DisplayOrder != nil {
		existingGroup.DisplayOrder = *req.DisplayOrder
	}
	if req.IsActive != nil {
		existingGroup.IsActive = *req.IsActive
	}
	if req.IsHeader != nil {
		existingGroup.IsHeader = *req.IsHeader
	}
	if req.ShowTotals != nil {
		existingGroup.ShowTotals = *req.ShowTotals
	}
	if req.IndentLevel != nil {
		existingGroup.IndentLevel = *req.IndentLevel
	}
	if req.BoldDisplay != nil {
		existingGroup.BoldDisplay = *req.BoldDisplay
	}

	// Update audit fields
	userID, _ := shared.GetUserID(ctx)
	existingGroup.UpdatedAt = time.Now()
	existingGroup.UpdatedBy = userID
	existingGroup.Version++

	// Validate business rules
	if err := existingGroup.Validate(); err != nil {
		return nil, fmt.Errorf("account group validation failed: %w", err)
	}

	// Update in repository
	err = s.accountGroupRepo.Update(ctx, existingGroup)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to update account group",
			logger.Fields{
				"group_id": id.String(),
				"error":    err.Error(),
			})
		return nil, fmt.Errorf("failed to update account group: %w", err)
	}

	logger.InfoContext(ctx, "Account group updated successfully",
		logger.Fields{
			"group_id":   existingGroup.ID.String(),
			"group_code": existingGroup.GroupCode,
			"group_name": existingGroup.GroupName,
		})

	return existingGroup, nil
}

func (s *accountService) DeleteAccountGroup(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.delete_account_group",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("group.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting account group deletion",
		logger.Fields{"group_id": id.String()})

	// Get existing group
	group, err := s.accountGroupRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get account group: %w", err)
	}

	// TODO(authz): enforce finance.account_groups.delete via iam.Service.Enforce() once
	// the session principal is wired into ctx.

	// Check if group can be deleted
	canDelete, err := s.accountGroupRepo.CanDeleteGroup(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to validate group deletion: %w", err)
	}
	if !canDelete {
		return errors.NewBusinessError("VALIDATION_ERROR", "Account group cannot be deleted - it may have child groups or associated accounts")
	}

	// Check if group has children
	hasChildren, err := s.accountGroupRepo.HasChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check for child groups: %w", err)
	}
	if hasChildren {
		return errors.NewBusinessError("VALIDATION_ERROR", "Cannot delete account group with child groups")
	}

	// Delete the group
	err = s.accountGroupRepo.Delete(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to delete account group",
			logger.Fields{
				"group_id": id.String(),
				"error":    err.Error(),
			})
		return fmt.Errorf("failed to delete account group: %w", err)
	}

	logger.InfoContext(ctx, "Account group deleted successfully",
		logger.Fields{
			"group_id":   id.String(),
			"group_code": group.GroupCode,
			"group_name": group.GroupName,
		})

	return nil
}

func (s *accountService) ListAccountGroups(ctx context.Context, filter *domain.AccountGroupFilter) ([]*domain.AccountGroup, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.list_account_groups",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	logger.DebugContext(ctx, "Listing account groups with filter")

	groups, err := s.accountGroupRepo.List(ctx, filter)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list account groups",
			logger.Fields{"error": err.Error()})
		return nil, fmt.Errorf("failed to list account groups: %w", err)
	}

	logger.DebugContext(ctx, "Account groups listed successfully",
		logger.Fields{"count": len(groups)})

	return groups, nil
}

func (s *accountService) GetAccountGroupHierarchy(ctx context.Context) ([]*domain.AccountGroup, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_group_hierarchy",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	logger.DebugContext(ctx, "Getting account group hierarchy")

	hierarchy, err := s.accountGroupRepo.GetGroupHierarchy(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get account group hierarchy",
			logger.Fields{"error": err.Error()})
		return nil, fmt.Errorf("failed to get account group hierarchy: %w", err)
	}

	logger.DebugContext(ctx, "Account group hierarchy retrieved successfully",
		logger.Fields{"count": len(hierarchy)})

	return hierarchy, nil
}

func (s *accountService) GetGroupsByFinancialStatement(ctx context.Context, statementType string) ([]*domain.AccountGroup, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_groups_by_financial_statement",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("statement.type", statementType),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting groups by financial statement",
		logger.Fields{"statement_type": statementType})

	groups, err := s.accountGroupRepo.GetGroupsByFinancialStatement(ctx, statementType)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get groups by financial statement",
			logger.Fields{
				"statement_type": statementType,
				"error":          err.Error(),
			})
		return nil, fmt.Errorf("failed to get groups by financial statement: %w", err)
	}

	logger.DebugContext(ctx, "Groups retrieved successfully",
		logger.Fields{
			"statement_type": statementType,
			"count":          len(groups),
		})

	return groups, nil
}

func (s *accountService) GetGroupsByCashFlowCategory(ctx context.Context, category string) ([]*domain.AccountGroup, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_groups_by_cash_flow_category",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("cash_flow.category", category),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting groups by cash flow category",
		logger.Fields{"category": category})

	// Convert string to CashFlowCategory enum
	cashFlowCategory := domain.CashFlowCategory(category)
	groups, err := s.accountGroupRepo.GetGroupsByCashFlowCategory(ctx, cashFlowCategory)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get groups by cash flow category",
			logger.Fields{
				"category": category,
				"error":    err.Error(),
			})
		return nil, fmt.Errorf("failed to get groups by cash flow category: %w", err)
	}

	logger.DebugContext(ctx, "Groups retrieved successfully",
		logger.Fields{
			"category": category,
			"count":    len(groups),
		})

	return groups, nil
}

// ============================================================================
// Unified Operations (Accounts and Groups Together)
// ============================================================================

// TODO: These methods will be implemented once UnifiedAccountRepository is created
// They represent the unified view of accounts and account groups in a single hierarchy

func (s *accountService) ListAccountsAndGroups(ctx context.Context, filter *domain.UnifiedFilter) ([]*domain.AccountNode, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.list_accounts_and_groups",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	logger.DebugContext(ctx, "Listing accounts and groups unified")

	// TODO: Implement unified repository call
	// For now, return placeholder to maintain interface compliance
	return nil, fmt.Errorf("unified account listing not yet implemented - requires UnifiedAccountRepository")
}

func (s *accountService) GetAccountHierarchyWithGroups(ctx context.Context) ([]*domain.AccountNode, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_hierarchy_with_groups",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	logger.DebugContext(ctx, "Getting account hierarchy with groups")

	// TODO: Implement unified hierarchy query
	// For now, return placeholder to maintain interface compliance
	return nil, fmt.Errorf("unified hierarchy not yet implemented - requires UnifiedAccountRepository")
}

func (s *accountService) SearchAccountsAndGroups(ctx context.Context, query string, limit int) ([]*domain.AccountNode, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.search_accounts_and_groups",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("search.query", query),
			attribute.Int("search.limit", limit),
		))
	defer span.End()

	logger.DebugContext(ctx, "Searching accounts and groups unified",
		logger.Fields{
			"query": query,
			"limit": limit,
		})

	// TODO: Implement unified search
	// For now, return placeholder to maintain interface compliance
	return nil, fmt.Errorf("unified search not yet implemented - requires UnifiedAccountRepository")
}

// =====================================================================
// ENHANCED ANALYTICS AND HIERARCHY OPERATIONS
// =====================================================================

// GetAccountChildrenHierarchy retrieves child accounts using hierarchy view with business validation
func (s *accountService) GetAccountChildrenHierarchy(ctx context.Context, parentAccountID uuid.UUID) ([]*domain.AccountHierarchy, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_children_hierarchy",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("parent_account_id", parentAccountID.String()),
		))
	defer span.End()

	// Permission check
	if err := s.checkAccountReadPermission(ctx); err != nil {
		return nil, err
	}

	// Validate that parent account exists and is accessible
	parentAccount, err := s.accountRepo.GetByID(ctx, parentAccountID)
	if err != nil {
		s.metrics.IncrementCounter("account.hierarchy.parent_not_found", metrics.Fields{
			"parent_account_id": parentAccountID.String(),
		})
		return nil, fmt.Errorf("parent account not found: %w", err)
	}

	// Business validation: ensure parent account can have children
	if !parentAccount.IsActive {
		return nil, errors.NewBusinessError("INACTIVE_PARENT_ACCOUNT", "parent account must be active to retrieve children")
	}

	logger.InfoContext(ctx, "Retrieving account children hierarchy",
		logger.Fields{
			"parent_account_id":   parentAccountID,
			"parent_account_code": parentAccount.AccountCode,
		})

	result, err := s.accountRepo.GetAccountChildrenHierarchy(ctx, parentAccountID)
	if err != nil {
		s.metrics.IncrementCounter("account.hierarchy.children_fetch_error", metrics.Fields{
			"parent_account_id": parentAccountID.String(),
		})
		return nil, fmt.Errorf("failed to get account children hierarchy: %w", err)
	}

	s.metrics.IncrementCounter("account.hierarchy.children_fetched", metrics.Fields{
		"parent_account_id": parentAccountID.String(),
	})
	s.metrics.SetGauge("account.hierarchy.children_count", float64(len(result)), metrics.Fields{
		"parent_account_id": parentAccountID.String(),
	})

	logger.InfoContext(ctx, "Successfully retrieved account children hierarchy",
		logger.Fields{
			"parent_account_id": parentAccountID,
			"children_count":    len(result),
		})

	return result, nil
}

// GetAccountSubtree retrieves an account and all its descendants with business validation
func (s *accountService) GetAccountSubtree(ctx context.Context, accountID uuid.UUID) ([]*domain.AccountHierarchy, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_subtree",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account_id", accountID.String()),
		))
	defer span.End()

	// Permission check
	if err := s.checkAccountReadPermission(ctx); err != nil {
		return nil, err
	}

	// Validate that account exists and is accessible
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		s.metrics.IncrementCounter("account.subtree.account_not_found", metrics.Fields{
			"account_id": accountID.String(),
		})
		return nil, fmt.Errorf("account not found: %w", err)
	}

	logger.InfoContext(ctx, "Retrieving account subtree",
		logger.Fields{
			"account_id":   accountID,
			"account_code": account.AccountCode,
		})

	result, err := s.accountRepo.GetAccountSubtree(ctx, accountID)
	if err != nil {
		s.metrics.IncrementCounter("account.subtree.fetch_error", metrics.Fields{
			"account_id": accountID.String(),
		})
		return nil, fmt.Errorf("failed to get account subtree: %w", err)
	}

	s.metrics.IncrementCounter("account.subtree.fetched", metrics.Fields{
		"account_id": accountID.String(),
	})
	s.metrics.SetGauge("account.subtree.node_count", float64(len(result)), metrics.Fields{
		"account_id": accountID.String(),
	})

	logger.InfoContext(ctx, "Successfully retrieved account subtree",
		logger.Fields{
			"account_id": accountID,
			"node_count": len(result),
		})

	return result, nil
}

// GetAccountsWithRecentActivity retrieves accounts with recent transaction activity
func (s *accountService) GetAccountsWithRecentActivity(ctx context.Context, filter *domain.AccountActivityFilter) ([]*domain.AccountActivity, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_accounts_with_recent_activity",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	// Permission check
	if err := s.checkAccountReadPermission(ctx); err != nil {
		return nil, err
	}

	// Validate filter parameters
	if err := s.validateActivityFilter(filter); err != nil {
		return nil, fmt.Errorf("invalid activity filter: %w", err)
	}

	logger.InfoContext(ctx, "Retrieving accounts with recent activity",
		logger.Fields{
			"filter": filter,
		})

	result, err := s.accountRepo.GetAccountsWithRecentActivity(ctx, filter)
	if err != nil {
		s.metrics.IncrementCounter("account.activity.recent_fetch_error", metrics.Fields{
			"error_type": "fetch_error",
		})
		return nil, fmt.Errorf("failed to get accounts with recent activity: %w", err)
	}

	s.metrics.IncrementCounter("account.activity.recent_fetched", metrics.Fields{
		"count": len(result),
	})
	s.metrics.SetGauge("account.activity.recent_count", float64(len(result)), metrics.Fields{})

	logger.InfoContext(ctx, "Successfully retrieved accounts with recent activity",
		logger.Fields{
			"account_count": len(result),
		})

	return result, nil
}

// GetStaleAccountBalances identifies accounts with non-zero balances but no recent activity
func (s *accountService) GetStaleAccountBalances(ctx context.Context, filter *domain.AccountActivityFilter) ([]*domain.AccountActivity, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_stale_account_balances",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	// Permission check
	if err := s.checkAccountReadPermission(ctx); err != nil {
		return nil, err
	}

	// Validate filter parameters
	if err := s.validateActivityFilter(filter); err != nil {
		return nil, fmt.Errorf("invalid activity filter: %w", err)
	}

	logger.InfoContext(ctx, "Retrieving stale account balances",
		logger.Fields{
			"filter": filter,
		})

	result, err := s.accountRepo.GetStaleAccountBalances(ctx, filter)
	if err != nil {
		s.metrics.IncrementCounter("account.activity.stale_fetch_error", metrics.Fields{
			"error_type": "fetch_error",
		})
		return nil, fmt.Errorf("failed to get stale account balances: %w", err)
	}

	// Log warning if we found stale accounts (potential business issue)
	if len(result) > 0 {
		logger.WarnContext(ctx, "Found accounts with stale balances - may require attention",
			logger.Fields{
				"stale_account_count": len(result),
			})
	}

	s.metrics.IncrementCounter("account.activity.stale_fetched", metrics.Fields{
		"count": len(result),
	})
	s.metrics.SetGauge("account.activity.stale_count", float64(len(result)), metrics.Fields{})

	logger.InfoContext(ctx, "Successfully retrieved stale account balances",
		logger.Fields{
			"stale_account_count": len(result),
		})

	return result, nil
}

// GetAccountActivitySummary provides detailed activity analysis with categorization
func (s *accountService) GetAccountActivitySummary(ctx context.Context, filter *domain.AccountActivityFilter) ([]*domain.AccountActivitySummary, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_activity_summary",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	// Permission check
	if err := s.checkAccountReadPermission(ctx); err != nil {
		return nil, err
	}

	// Validate filter parameters
	if err := s.validateActivityFilter(filter); err != nil {
		return nil, fmt.Errorf("invalid activity filter: %w", err)
	}

	logger.InfoContext(ctx, "Retrieving account activity summary",
		logger.Fields{
			"filter": filter,
		})

	result, err := s.accountRepo.GetAccountActivitySummary(ctx, filter)
	if err != nil {
		s.metrics.IncrementCounter("account.activity.summary_fetch_error", metrics.Fields{
			"error_type": "fetch_error",
		})
		return nil, fmt.Errorf("failed to get account activity summary: %w", err)
	}

	// Analyze activity patterns for business insights
	s.analyzeActivityPatterns(ctx, result)

	s.metrics.IncrementCounter("account.activity.summary_fetched", metrics.Fields{
		"count": len(result),
	})
	s.metrics.SetGauge("account.activity.summary_count", float64(len(result)), metrics.Fields{})

	logger.InfoContext(ctx, "Successfully retrieved account activity summary",
		logger.Fields{
			"account_count": len(result),
		})

	return result, nil
}

// =====================================================================
// PRIVATE HELPER METHODS
// =====================================================================

// validateActivityFilter validates the activity filter parameters
func (s *accountService) validateActivityFilter(filter *domain.AccountActivityFilter) error {
	if filter == nil {
		return errors.ValidationErrors{{
			Field:   "filter",
			Message: "filter cannot be nil",
		}}
	}

	var validationErrors errors.ValidationErrors

	// Validate minimum entries if specified
	if filter.MinEntries != nil && *filter.MinEntries < 0 {
		validationErrors.Add("min_entries", "min_entries must be non-negative")
	}

	// Validate minimum days inactive if specified
	if filter.MinDaysInactive != nil && *filter.MinDaysInactive < 0 {
		validationErrors.Add("min_days_inactive", "min_days_inactive must be non-negative")
	}

	// Validate minimum balance if specified
	if filter.MinBalance != nil && filter.MinBalance.IsNegative() {
		validationErrors.Add("min_balance", "min_balance must be non-negative")
	}

	// Validate activity level if specified
	if filter.ActivityLevel != nil {
		validLevels := []string{"Inactive", "Low Activity", "Medium Activity", "High Activity"}
		valid := false
		for _, level := range validLevels {
			if level == *filter.ActivityLevel {
				valid = true
				break
			}
		}
		if !valid {
			validationErrors.Add("activity_level", "invalid activity_level: must be one of Inactive, Low Activity, Medium Activity, High Activity")
		}
	}

	if validationErrors.HasErrors() {
		return validationErrors
	}
	return nil
}

// analyzeActivityPatterns analyzes activity patterns for business insights
func (s *accountService) analyzeActivityPatterns(ctx context.Context, activities []*domain.AccountActivitySummary) {
	if len(activities) == 0 {
		return
	}

	// Count activity levels for business intelligence
	activityLevelCounts := make(map[string]int)
	totalActivity := int64(0)

	for _, activity := range activities {
		activityLevelCounts[activity.ActivityLevel]++
		totalActivity += activity.EntriesLast30Days
	}

	// Log insights for monitoring and alerting
	logger.InfoContext(ctx, "Account activity pattern analysis",
		logger.Fields{
			"total_accounts":     len(activities),
			"inactive_accounts":  activityLevelCounts["Inactive"],
			"low_activity":       activityLevelCounts["Low Activity"],
			"medium_activity":    activityLevelCounts["Medium Activity"],
			"high_activity":      activityLevelCounts["High Activity"],
			"total_activity_30d": totalActivity,
		})

	// Set metrics for monitoring dashboards
	for level, count := range activityLevelCounts {
		s.metrics.SetGauge(fmt.Sprintf("account.activity.level.%s", level), float64(count), metrics.Fields{
			"level": level,
		})
	}
	s.metrics.SetGauge("account.activity.total_entries_30d", float64(totalActivity), metrics.Fields{})
}

// checkAccountReadPermission validates if the user has permission to read accounts.
// TODO(authz): enforce finance.accounts.read via iam.Service.Enforce() once
// the session principal is wired into ctx.
func (s *accountService) checkAccountReadPermission(ctx context.Context) error {
	return nil
}
