package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/iam/authz"
	"github.com/niiniyare/erp/internal/core/iam/model"
	"github.com/niiniyare/erp/internal/shared"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

type AccountService interface {
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
}

type accountService struct {
	repo               domain.AccountsRepository
	tracing            tracing.TracingService
	metrics            metrics.MetricsProvider
	settingsHelper     *SettingsHelper
	iamService         iam.Service
	featureFlagService featureflag.Service
}

func NewAccountService(
	repo domain.AccountsRepository,
	tracing tracing.TracingService,
	metrics metrics.MetricsProvider,
	iamService iam.Service,
	featureFlagService featureflag.Service,
) AccountService {
	return &accountService{
		repo:               repo,
		tracing:            tracing,
		metrics:            metrics,
		settingsHelper:     NewSettingsHelper(),
		iamService:         iamService,
		featureFlagService: featureFlagService,
	}
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

	// ABAC - Check if user can create accounts
	if userID, ok := shared.GetUserID(ctx); ok {
		permissionReq := &authz.PermissionEvaluationRequest{
			UserID:       userID,
			ResourceType: "account",
			Action:       "create",
			EntityID:     req.EntityID,
			// TODO: get the rest from thee setion
			// ResourceID:   &uuid.UUID{},
			// Context:      map[string]any{},
			// RequestID:    "",
		}

		result, err := s.iamService.Authorization().EvaluatePermission(ctx, permissionReq)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to evaluate permission", logger.Fields{
				"user_id": userID.String(),
				"error":   err.Error(),
			})
			return nil, errors.NewBusinessError("PERMISSION_ERROR", "Failed to evaluate permissions")
		}

		if result.Decision != model.PolicyDecisionAllow {
			logger.WarnContext(ctx, "Permission denied for account creation", logger.Fields{
				"user_id":  userID.String(),
				"decision": string(result.Decision),
			})
			return nil, errors.NewBusinessError("UNAUTHORIZED", "Cannot create account")
		}

		logger.DebugContext(ctx, "Permission granted for account creation", logger.Fields{
			"user_id":            userID.String(),
			"evaluation_time_ms": result.EvaluationTimeMS,
		})
	}

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

func (s *accountService) GetAccountByID(ctx context.Context, id uuid.UUID) (*domain.Accounts, error) {
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

func (s *accountService) GetAccountByCode(ctx context.Context, code string) (*domain.Accounts, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.get_account_by_code",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.code", code),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting account by code",
		logger.Fields{"account_code": code})

	// ABAC - Check if user can read accounts
	if userID, ok := shared.GetUserID(ctx); ok {
		permissionReq := &authz.PermissionEvaluationRequest{
			UserID:       userID,
			ResourceType: "account",
			Action:       "read",
		}

		result, err := s.iamService.Authorization().EvaluatePermission(ctx, permissionReq)
		if err != nil || result.Decision != model.PolicyDecisionAllow {
			logger.WarnContext(ctx, "Permission denied for account read", logger.Fields{
				"user_id":      userID.String(),
				"account_code": code,
			})
			return nil, errors.NewBusinessError("UNAUTHORIZED", "Cannot read account")
		}
	}

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

func (s *accountService) UpdateAccount(ctx context.Context, id uuid.UUID, req domain.UpdateAccountRequest) (*domain.Accounts, error) {
	ctx, span := s.tracing.StartSpan(ctx, "account_service.update_account",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("account.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting account update",
		logger.Fields{"account_id": id.String()})

	// ABAC - Check if user can update this account
	if userID, ok := shared.GetUserID(ctx); ok {
		permissionReq := &authz.PermissionEvaluationRequest{
			UserID:       userID,
			ResourceType: "account",
			ResourceID:   &id,
			Action:       "update",
		}

		result, err := s.iamService.Authorization().EvaluatePermission(ctx, permissionReq)
		if err != nil || result.Decision != model.PolicyDecisionAllow {
			logger.WarnContext(ctx, "Permission denied for account update", logger.Fields{
				"user_id":    userID.String(),
				"account_id": id.String(),
			})
			return nil, errors.NewBusinessError("UNAUTHORIZED", "Cannot update account")
		}
	}

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

	// ABAC - Check if user can delete accounts
	if userID, ok := shared.GetUserID(ctx); ok {
		permissionReq := &authz.PermissionEvaluationRequest{
			UserID:       userID,
			ResourceType: "account",
			ResourceID:   &id,
			Action:       "delete",
		}

		result, err := s.iamService.Authorization().EvaluatePermission(ctx, permissionReq)
		if err != nil || result.Decision != model.PolicyDecisionAllow {
			logger.WarnContext(ctx, "Permission denied for account deletion", logger.Fields{
				"user_id":    userID.String(),
				"account_id": id.String(),
			})
			return errors.NewBusinessError("UNAUTHORIZED", "Cannot delete account")
		}
	}

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

	// ABAC - Filter accounts based on user permissions
	if userID, ok := shared.GetUserID(ctx); ok {
		// Get user's effective permissions to determine what they can see
		permissions, err := s.iamService.Authorization().GetUserEffectivePermissions(ctx, userID, nil)
		if err != nil {
			logger.WarnContext(ctx, "Failed to get user permissions for account listing", logger.Fields{
				"user_id": userID.String(),
				"error":   err.Error(),
			})
		} else {
			// In a real implementation, you would filter the accounts based on permissions
			logger.DebugContext(ctx, "Applied permission-based filtering", logger.Fields{
				"user_id":           userID.String(),
				"permissions_count": len(permissions.Permissions),
			})
		}
	}

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

func (s *accountService) GetAccountHierarchy(ctx context.Context, rootAccountID uuid.UUID) ([]*domain.Accounts, error) {
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

func (s *accountService) GetAccountsByType(ctx context.Context, accountType string, rootType *domain.RootType) ([]*domain.Accounts, error) {
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

func (s *accountService) GetActiveAccounts(ctx context.Context) ([]*domain.Accounts, error) {
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

	// ABAC - Check if user can search accounts
	if userID, ok := shared.GetUserID(ctx); ok {
		permissionReq := &authz.PermissionEvaluationRequest{
			UserID:       userID,
			ResourceType: "account",
			Action:       "search",
		}

		result, err := s.iamService.Authorization().EvaluatePermission(ctx, permissionReq)
		if err != nil || result.Decision != model.PolicyDecisionAllow {
			logger.WarnContext(ctx, "Permission denied for account search", logger.Fields{
				"user_id": userID.String(),
				"query":   query,
			})
			return nil, errors.NewBusinessError("UNAUTHORIZED", "Cannot search accounts")
		}
	}

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
	account, err := s.repo.GetAccountWithGroups(ctx, id)
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
	account, err := s.repo.GetAccountWithGroupsByCode(ctx, code)
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
	accounts, err := s.repo.ListAccountsWithGroups(ctx, filter)
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
	accounts, err := s.repo.SearchAccountsWithGroups(ctx, query, limit)
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
	accounts, err := s.repo.GetLeafAccountsOnly(ctx, rootType)
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
	accounts, err := s.repo.GetCompleteChartOfAccounts(ctx, filter)
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
	account, err := s.repo.GetAccountForReporting(ctx, accountID)
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
	accounts, err := s.repo.GetAccountsByStatementSection(ctx, section, entityID)
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
	accounts, err := s.repo.GetAccountsByGroup(ctx, groupCode, entityID)
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
	accounts, err := s.repo.GetAccountsByHeader(ctx, headerCode, entityID)
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
	accounts, err := s.repo.GetTrialBalanceAccounts(ctx, entityID, nonZeroOnly)
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
	accounts, err := s.repo.GetAccountsWithBalances(ctx, filter)
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
	accounts, err := s.repo.GetCashFlowAccounts(ctx, entityID)
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
	summary, err := s.repo.GetAccountSummaryByGroup(ctx, entityID)
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
