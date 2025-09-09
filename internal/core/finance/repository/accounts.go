package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/shared"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

type chartOfAccountsRepository struct {
	store   db.Store
	tracing tracing.TracingService
}

func NewAccountsRepository(store db.Store, tracing tracing.TracingService) domain.AccountsRepository {
	return &chartOfAccountsRepository{
		store:   store,
		tracing: tracing,
	}
}

// NewAccountRepository is an alias for NewAccountsRepository for test compatibility
func NewAccountRepository(store db.Store, tracing tracing.TracingService) domain.AccountsRepository {
	return NewAccountsRepository(store, tracing)
}

// Basic CRUD Operations

func (r *chartOfAccountsRepository) Create(ctx context.Context, account *domain.Accounts) error {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.Create")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Use tenant-aware transaction for proper isolation
	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Map domain account directly to SQLC parameters
		params, err := mapDomainAccountToSQLCCreateDirect(account)
		if err != nil {
			return fmt.Errorf("failed to map create account request: %w", err)
		}

		// Execute SQLC query within tenant context
		sqlcAccount, err := s.CreateAccount(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "create_account")
		}

		// Update the account with generated fields
		account.ID = sqlcAccount.ID
		account.TenantID = sqlcAccount.TenantID
		account.AccountLevel = sqlcAccount.AccountLevel
		account.AccountPath = sqlcAccount.AccountPath
		account.CreatedAt = sqlcAccount.CreatedAt
		account.UpdatedAt = sqlcAccount.UpdatedAt

		return nil
	})
}

func (r *chartOfAccountsRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Accounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetByID")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var account *domain.Accounts
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		sqlcAccount, err := s.GetAccountByID(ctx, id)
		if err != nil {
			if err == db.ErrNoRows {
				return domain.ErrAccountNotFound
			}
			return r.mapDatabaseError(err, "get_account_by_id")
		}

		account, err = mapSQLCAccountToDomain(sqlcAccount)
		if err != nil {
			return fmt.Errorf("failed to map SQLC account to domain: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (r *chartOfAccountsRepository) GetByCode(ctx context.Context, entityID *uuid.UUID, accountCode string) (*domain.Accounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetByCode")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var account *domain.Accounts
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Note: SQLC GetAccountByCode only takes accountCode string, entityID filtering handled by RLS
		sqlcAccount, err := s.GetAccountByCode(ctx, accountCode)
		if err != nil {
			if err == db.ErrNoRows {
				return domain.ErrAccountNotFound
			}
			return r.mapDatabaseError(err, "get_account_by_code")
		}

		account, err = mapSQLCAccountToDomain(sqlcAccount)
		if err != nil {
			return fmt.Errorf("failed to map SQLC account to domain: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (r *chartOfAccountsRepository) Update(ctx context.Context, account *domain.Accounts) error {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.Update")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Map account to SQLC parameters
		params := db.UpdateAccountParams{
			AccountID:              account.ID,
			AccountName:            &account.AccountName,
			AccountDescription:     account.AccountDescription,
			AccountType:            &account.AccountType,
			AccountSubtype:         account.AccountSubtype,
			IsActive:               &account.IsActive,
			AllowManualEntries:     &account.AllowManualEntries,
			RequireReference:       &account.RequireReference,
			FinancialStatementLine: account.FinancialStatementLine,
			ReportOrder:            &account.ReportOrder,
			IsBudgetable:           &account.IsBudgetable,
			UpdatedBy:              account.UpdatedBy,
		}

		// Handle optional budget variance threshold
		if !account.BudgetVarianceThreshold.IsZero() {
			params.BudgetVarianceThreshold = pgtype.Numeric{
				Int:   account.BudgetVarianceThreshold.BigInt(),
				Valid: true,
			}
		}

		// Handle optional attributes
		if account.AccountAttributes != nil {
			attributes, err := json.Marshal(account.AccountAttributes)
			if err != nil {
				return fmt.Errorf("failed to marshal account attributes: %w", err)
			}
			params.AccountAttributes = attributes
		}

		// Execute SQLC update query within tenant context
		_, err := s.UpdateAccount(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "update_account")
		}

		return nil
	})
}

func (r *chartOfAccountsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.Delete")
	defer span.End()

	// Get tenant and user ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	userID, _ := shared.GetUserID(ctx) // Optional

	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		params := db.SoftDeleteAccountParams{
			AccountID: id,
			UpdatedBy: nil, // Don't set updated_by if userID is not available
		}

		// Only set UpdatedBy if we have a valid user ID
		if userID != uuid.Nil {
			params.UpdatedBy = &userID
		}
		rowsAffected, err := s.SoftDeleteAccount(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "soft_delete_account")
		}

		// Check if the account was found and deleted
		if rowsAffected == 0 {
			return domain.ErrAccountNotFound
		}

		return nil
	})
}

// List and Filter Operations

func (r *chartOfAccountsRepository) List(ctx context.Context, filter *domain.AccountFilter) ([]*domain.Accounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.List")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var accounts []*domain.Accounts
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Map domain filter to SQLC parameters
		params, err := mapAccountFilterToSQLCParams(filter)
		if err != nil {
			return fmt.Errorf("failed to map account filter: %w", err)
		}

		// Execute SQLC query within tenant context
		sqlcAccounts, err := s.ListAccounts(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "list_accounts")
		}

		// Map results to domain models
		accounts = make([]*domain.Accounts, 0, len(sqlcAccounts))
		for _, sqlcAccount := range sqlcAccounts {
			account, err := mapSQLCAccountToDomain(sqlcAccount)
			if err != nil {
				return fmt.Errorf("failed to map SQLC account to domain: %w", err)
			}
			accounts = append(accounts, account)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *chartOfAccountsRepository) Count(ctx context.Context, filter *domain.AccountFilter) (int64, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.Count")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return 0, fmt.Errorf("tenant ID not found in context")
	}

	var count int64
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Map basic filter parameters for count
		params := db.CountAccountsParams{}

		if filter.RootType != nil {
			switch *filter.RootType {
			case domain.RootTypeAsset:
				rootTypeStr := "ASSET"
				params.RootType = &rootTypeStr
			case domain.RootTypeLiability:
				rootTypeStr := "LIABILITY"
				params.RootType = &rootTypeStr
			case domain.RootTypeEquity:
				rootTypeStr := "EQUITY"
				params.RootType = &rootTypeStr
			case domain.RootTypeRevenue:
				rootTypeStr := "REVENUE"
				params.RootType = &rootTypeStr
			case domain.RootTypeExpense:
				rootTypeStr := "EXPENSE"
				params.RootType = &rootTypeStr
			}
		}

		if filter.IsActive != nil {
			params.IsActive = filter.IsActive
		}

		var err error
		count, err = s.CountAccounts(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "count_accounts")
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *chartOfAccountsRepository) ListByParent(ctx context.Context, parentID uuid.UUID) ([]*domain.Accounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.ListByParent")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var accounts []*domain.Accounts
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		sqlcAccounts, err := s.ListAccountsByParent(ctx, &parentID)
		if err != nil {
			return r.mapDatabaseError(err, "list_accounts_by_parent")
		}

		accounts = make([]*domain.Accounts, 0, len(sqlcAccounts))
		for _, sqlcAccount := range sqlcAccounts {
			account, err := mapSQLCAccountToDomain(sqlcAccount)
			if err != nil {
				return fmt.Errorf("failed to map SQLC account to domain: %w", err)
			}
			accounts = append(accounts, account)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

// ListByRootType retrieves accounts by root type
func (r *chartOfAccountsRepository) ListByRootType(ctx context.Context, rootType domain.RootType) ([]*domain.Accounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.ListByRootType")
	defer span.End()

	// Create filter with root type
	filter := &domain.AccountFilter{
		RootType: &rootType,
	}

	return r.List(ctx, filter)
}

// Hierarchy Operations

func (r *chartOfAccountsRepository) GetAccountHierarchy(ctx context.Context, rootID uuid.UUID) ([]*domain.Accounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetAccountHierarchy")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var accounts []*domain.Accounts
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Convert rootID to string for hierarchy query
		rootIDStr := rootID.String()
		sqlcAccounts, err := s.GetAccountHierarchy(ctx, &rootIDStr)
		if err != nil {
			return r.mapDatabaseError(err, "get_account_hierarchy")
		}

		accounts = make([]*domain.Accounts, 0, len(sqlcAccounts))
		for _, sqlcAccount := range sqlcAccounts {
			account, err := mapSQLCAccountToDomain(sqlcAccount)
			if err != nil {
				return fmt.Errorf("failed to map SQLC account to domain: %w", err)
			}
			accounts = append(accounts, account)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

// GetAccountPath retrieves the full path of accounts from root to specified account
func (r *chartOfAccountsRepository) GetAccountPath(ctx context.Context, accountID uuid.UUID) ([]domain.Accounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetAccountPath")
	defer span.End()

	// Get the account first
	account, err := r.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// If no path is set, return just the account itself
	if account.AccountPath == nil || *account.AccountPath == "" {
		return []domain.Accounts{*account}, nil
	}

	// TODO: Parse account path and retrieve all accounts in the path
	// For now, just return the single account
	return []domain.Accounts{*account}, nil
}

// ValidateHierarchy validates if the parent-child relationship is valid
func (r *chartOfAccountsRepository) ValidateHierarchy(ctx context.Context, accountID, parentID uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.ValidateHierarchy")
	defer span.End()

	// Check if parent exists
	_, err := r.GetByID(ctx, parentID)
	if err != nil {
		return err
	}

	// Check if account exists
	_, err = r.GetByID(ctx, accountID)
	if err != nil {
		return err
	}

	// TODO: Add circular reference detection
	// TODO: Add business rule validations

	return nil
}

// Balance Operations

func (r *chartOfAccountsRepository) GetAccountBalance(ctx context.Context, accountID uuid.UUID, asOfDate *time.Time) (*domain.AccountBalance, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetAccountBalance")
	defer span.End()

	// Get account to retrieve current balance
	account, err := r.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Use current balance from account record
	balanceAmount := account.CurrentBalance

	// Get account details
	accountDetails, err := r.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	asOfTime := time.Now()
	if asOfDate != nil {
		asOfTime = *asOfDate
	}

	return &domain.AccountBalance{
		AccountID:    accountID,
		Account:      *accountDetails,
		TotalDebits:  decimal.Zero, // TODO: Calculate from transaction entries
		TotalCredits: decimal.Zero, // TODO: Calculate from transaction entries
		NetBalance:   balanceAmount,
		AsOfDate:     asOfTime,
	}, nil
}

// GetAccountBalances retrieves balances for multiple accounts
func (r *chartOfAccountsRepository) GetAccountBalances(ctx context.Context, accountIDs []uuid.UUID, asOfDate *time.Time) ([]*domain.AccountBalance, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetAccountBalances")
	defer span.End()

	balances := make([]*domain.AccountBalance, 0, len(accountIDs))
	for _, accountID := range accountIDs {
		balance, err := r.GetAccountBalance(ctx, accountID, asOfDate)
		if err != nil {
			return nil, fmt.Errorf("failed to get balance for account %s: %w", accountID, err)
		}
		balances = append(balances, balance)
	}

	return balances, nil
}

func (r *chartOfAccountsRepository) GetTrialBalance(ctx context.Context, entityID *uuid.UUID, asOfDate *time.Time) ([]*domain.TrialBalanceEntry, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetTrialBalance")
	defer span.End()

	// TODO: Implement proper trial balance calculation
	// For now, return empty result
	entries := make([]*domain.TrialBalanceEntry, 0)

	return entries, nil
}

// Business Logic Queries

func (r *chartOfAccountsRepository) GetActiveAccounts(ctx context.Context, entityID *uuid.UUID) ([]*domain.Accounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetActiveAccounts")
	defer span.End()

	// Create filter for active accounts only
	filter := &domain.AccountFilter{
		EntityID: entityID,
		IsActive: &[]bool{true}[0],
	}

	return r.List(ctx, filter)
}

func (r *chartOfAccountsRepository) GetControlAccounts(ctx context.Context, entityID *uuid.UUID) ([]*domain.Accounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetControlAccounts")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var accounts []*domain.Accounts
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Pass entityID parameter to GetControlAccounts
		sqlcAccounts, err := s.GetControlAccounts(ctx, entityID)
		if err != nil {
			return r.mapDatabaseError(err, "get_control_accounts")
		}

		accounts = make([]*domain.Accounts, 0, len(sqlcAccounts))
		for _, sqlcAccount := range sqlcAccounts {
			account, err := mapSQLCAccountToDomain(sqlcAccount)
			if err != nil {
				return fmt.Errorf("failed to map SQLC account to domain: %w", err)
			}
			accounts = append(accounts, account)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *chartOfAccountsRepository) GetAccountsByType(ctx context.Context, accountType string, rootType *domain.RootType) ([]*domain.Accounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetAccountsByType")
	defer span.End()

	// Create filter - AccountType not available in domain.AccountFilter
	// Use a simple List call instead
	filter := &domain.AccountFilter{
		RootType: rootType,
	}

	return r.List(ctx, filter)
}

// Validation Helpers

func (r *chartOfAccountsRepository) ValidateAccountCode(ctx context.Context, code string, excludeID *uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.ValidateAccountCode")
	defer span.End()

	unique, err := r.IsAccountCodeUnique(ctx, nil, code, excludeID)
	if err != nil {
		return err
	}
	if !unique {
		return domain.ErrAccountCodeExists
	}
	return nil
}

func (r *chartOfAccountsRepository) IsAccountCodeUnique(ctx context.Context, entityID *uuid.UUID, accountCode string, excludeID *uuid.UUID) (bool, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.IsAccountCodeUnique")
	defer span.End()

	account, err := r.GetByCode(ctx, entityID, accountCode)
	if err != nil {
		// If account not found, code is unique
		if err == domain.ErrAccountNotFound {
			return true, nil
		}
		return false, err
	}

	// If we're excluding a specific account ID and that's the one we found, code is unique
	if excludeID != nil && account.ID == *excludeID {
		return true, nil
	}

	// Code is not unique
	return false, nil
}

func (r *chartOfAccountsRepository) HasChildren(ctx context.Context, accountID uuid.UUID) (bool, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.HasChildren")
	defer span.End()

	// Get tenant ID from context
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return false, fmt.Errorf("tenant ID not found in context")
	}

	var hasChildren bool
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Use ListAccountsByParent to check if children exist
		children, err := s.ListAccountsByParent(ctx, &accountID)
		if err != nil {
			return r.mapDatabaseError(err, "list_accounts_by_parent")
		}

		hasChildren = len(children) > 0
		return nil
	})
	if err != nil {
		return false, err
	}

	return hasChildren, nil
}

func (r *chartOfAccountsRepository) HasTransactions(ctx context.Context, accountID uuid.UUID) (bool, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.HasTransactions")
	defer span.End()

	// TODO: Implement by checking transaction entries table
	// For now, return false
	return false, nil
}

func (r *chartOfAccountsRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Accounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.Search")
	defer span.End()

	// For now, return empty result
	return make([]*domain.Accounts, 0), nil
}

func (r *chartOfAccountsRepository) UpdateBalance(ctx context.Context, accountID uuid.UUID, balance domain.AccountBalance) error {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.UpdateBalance")
	defer span.End()

	// For now, do nothing
	return nil
}

func (r *chartOfAccountsRepository) GetChildren(ctx context.Context, accountID uuid.UUID) ([]*domain.Accounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetChildren")
	defer span.End()

	return r.ListByParent(ctx, accountID)
}

// Error mapping helper
func (r *chartOfAccountsRepository) mapDatabaseError(err error, operation string) error {
	// Convert database-specific errors to domain errors
	if err == db.ErrNoRows {
		return domain.ErrAccountNotFound
	}

	// Check for constraint violations
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "23505": // unique violation
			if strings.Contains(pgErr.Message, "account_code") {
				return domain.ErrAccountCodeExists
			}
		case "23503": // foreign key violation
			// Include more details about the constraint violation for debugging
			return fmt.Errorf("foreign key constraint violation: %s (detail: %s, constraint: %s)", pgErr.Message, pgErr.Detail, pgErr.ConstraintName)
		}
	}

	// Default to internal server error
	return fmt.Errorf("database operation failed: %s: %w", operation, err)
}

// Helper function to map SQLC root type enum to domain
func mapSQLCRootTypeToDomain(sqlcRootType string) domain.RootType {
	switch sqlcRootType {
	case "ASSET":
		return domain.RootTypeAsset
	case "LIABILITY":
		return domain.RootTypeLiability
	case "EQUITY":
		return domain.RootTypeEquity
	case "REVENUE":
		return domain.RootTypeRevenue
	case "EXPENSE":
		return domain.RootTypeExpense
	default:
		return domain.RootTypeAsset // fallback
	}
}

// Enhanced view-based operations

func (r *chartOfAccountsRepository) GetAccountWithGroups(ctx context.Context, id uuid.UUID) (*domain.AccountWithGroups, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetAccountWithGroups")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result *domain.AccountWithGroups
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		sqlcAccount, err := s.GetAccountWithGroupsByID(ctx, id)
		if err != nil {
			return r.mapDatabaseError(err, "get_account_with_groups")
		}

		result, err = mapSQLCAccountWithGroupsToDomain(*sqlcAccount)
		return err
	})

	return result, err
}

func (r *chartOfAccountsRepository) GetAccountWithGroupsByCode(ctx context.Context, code string) (*domain.AccountWithGroups, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetAccountWithGroupsByCode")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result *domain.AccountWithGroups
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		sqlcAccount, err := s.GetAccountWithGroupsByCode(ctx, code)
		if err != nil {
			return r.mapDatabaseError(err, "get_account_with_groups_by_code")
		}

		result, err = mapSQLCAccountWithGroupsToDomain(*sqlcAccount)
		return err
	})

	return result, err
}

func (r *chartOfAccountsRepository) ListAccountsWithGroups(ctx context.Context, filter *domain.AccountFilter) ([]*domain.AccountWithGroups, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.ListAccountsWithGroups")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result []*domain.AccountWithGroups
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		params, err := mapAccountFilterToSQLCWithGroups(filter)
		if err != nil {
			return err
		}

		sqlcAccounts, err := s.ListAccountsWithGroups(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "list_accounts_with_groups")
		}

		result = make([]*domain.AccountWithGroups, len(sqlcAccounts))
		for i, sqlcAccount := range sqlcAccounts {
			result[i], err = mapSQLCAccountWithGroupsToDomain(*sqlcAccount)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

func (r *chartOfAccountsRepository) SearchAccountsWithGroups(ctx context.Context, query string, limit int) ([]*domain.AccountWithGroups, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.SearchAccountsWithGroups")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result []*domain.AccountWithGroups
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		params := db.SearchAccountsWithGroupInfoParams{
			SearchTerm: &query,
			Limit:      int32(limit),
			Offset:     0,
		}

		sqlcAccounts, err := s.SearchAccountsWithGroupInfo(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "search_accounts_with_groups")
		}

		result = make([]*domain.AccountWithGroups, len(sqlcAccounts))
		for i, sqlcAccount := range sqlcAccounts {
			result[i], err = mapSQLCAccountWithGroupsToDomain(*sqlcAccount)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

func (r *chartOfAccountsRepository) GetLeafAccountsOnly(ctx context.Context, rootType *string) ([]*domain.AccountWithGroups, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetLeafAccountsOnly")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result []*domain.AccountWithGroups
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		var rootTypeStr string
		if rootType != nil {
			rootTypeStr = *rootType
		}
		params := db.GetLeafAccountsWithGroupsParams{
			EntityID: nil, // Will be set by tenant context
			RootType: &rootTypeStr,
		}

		sqlcAccounts, err := s.GetLeafAccountsWithGroups(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "get_leaf_accounts_only")
		}

		result = make([]*domain.AccountWithGroups, len(sqlcAccounts))
		for i, sqlcAccount := range sqlcAccounts {
			result[i], err = mapSQLCAccountWithGroupsToDomain(*sqlcAccount)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

// Complete chart of accounts operations

func (r *chartOfAccountsRepository) GetCompleteChartOfAccounts(ctx context.Context, filter *domain.ChartOfAccountsFilter) ([]*domain.ChartOfAccountsComplete, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetCompleteChartOfAccounts")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result []*domain.ChartOfAccountsComplete
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		params, err := mapChartOfAccountsFilterToSQLC(filter)
		if err != nil {
			return err
		}

		sqlcAccounts, err := s.GetChartOfAccountsComplete(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "get_complete_chart_of_accounts")
		}

		result = make([]*domain.ChartOfAccountsComplete, len(sqlcAccounts))
		for i, sqlcAccount := range sqlcAccounts {
			result[i] = mapSQLCChartOfAccountsCompleteToDomain(*sqlcAccount)
		}

		return nil
	})

	return result, err
}

func (r *chartOfAccountsRepository) GetAccountForReporting(ctx context.Context, accountID uuid.UUID) (*domain.ChartOfAccountsComplete, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetAccountForReporting")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result *domain.ChartOfAccountsComplete
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		sqlcAccount, err := s.GetAccountReportingInfo(ctx, accountID)
		if err != nil {
			return r.mapDatabaseError(err, "get_account_for_reporting")
		}

		result = mapSQLCChartOfAccountsCompleteToDomain(*sqlcAccount)
		return nil
	})

	return result, err
}

func (r *chartOfAccountsRepository) GetAccountsByStatementSection(ctx context.Context, section string, entityID *uuid.UUID) ([]*domain.ChartOfAccountsComplete, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetAccountsByStatementSection")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result []*domain.ChartOfAccountsComplete
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		params := db.GetAccountsByStatementParams{
			StatementSection: &section,
			EntityID:         entityID,
		}

		sqlcAccounts, err := s.GetAccountsByStatement(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "get_accounts_by_statement_section")
		}

		result = make([]*domain.ChartOfAccountsComplete, len(sqlcAccounts))
		for i, sqlcAccount := range sqlcAccounts {
			result[i] = mapSQLCChartOfAccountsCompleteToDomain(*sqlcAccount)
		}

		return nil
	})

	return result, err
}

func (r *chartOfAccountsRepository) GetAccountsByGroup(ctx context.Context, groupCode string, entityID *uuid.UUID) ([]*domain.ChartOfAccountsComplete, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetAccountsByGroup")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result []*domain.ChartOfAccountsComplete
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		params := db.GetAccountsByGroupCodeParams{
			GroupCode: &groupCode,
			EntityID:  entityID,
		}

		sqlcAccounts, err := s.GetAccountsByGroupCode(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "get_accounts_by_group")
		}

		result = make([]*domain.ChartOfAccountsComplete, len(sqlcAccounts))
		for i, sqlcAccount := range sqlcAccounts {
			result[i] = mapSQLCChartOfAccountsCompleteToDomain(*sqlcAccount)
		}

		return nil
	})

	return result, err
}

func (r *chartOfAccountsRepository) GetAccountsByHeader(ctx context.Context, headerCode string, entityID *uuid.UUID) ([]*domain.ChartOfAccountsComplete, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetAccountsByHeader")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result []*domain.ChartOfAccountsComplete
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		params := db.GetAccountsByHeaderCodeParams{
			HeaderCode: &headerCode,
			EntityID:   entityID,
		}

		sqlcAccounts, err := s.GetAccountsByHeaderCode(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "get_accounts_by_header")
		}

		result = make([]*domain.ChartOfAccountsComplete, len(sqlcAccounts))
		for i, sqlcAccount := range sqlcAccounts {
			result[i] = mapSQLCChartOfAccountsCompleteToDomain(*sqlcAccount)
		}

		return nil
	})

	return result, err
}

// Financial reporting operations

func (r *chartOfAccountsRepository) GetTrialBalanceAccounts(ctx context.Context, entityID *uuid.UUID, nonZeroOnly bool) ([]*domain.TrialBalanceSummary, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetTrialBalanceAccounts")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result []*domain.TrialBalanceSummary
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		params := db.GetTrialBalanceDataParams{
			EntityID:    entityID,
			NonZeroOnly: &nonZeroOnly,
		}

		sqlcAccounts, err := s.GetTrialBalanceData(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "get_trial_balance_accounts")
		}

		result = make([]*domain.TrialBalanceSummary, len(sqlcAccounts))
		for i, sqlcAccount := range sqlcAccounts {
			result[i] = mapSQLCTrialBalanceSummaryToDomain(*sqlcAccount)
		}

		return nil
	})

	return result, err
}

func (r *chartOfAccountsRepository) GetAccountsWithBalances(ctx context.Context, filter *domain.BalanceFilter) ([]*domain.ChartOfAccountsComplete, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetAccountsWithBalances")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result []*domain.ChartOfAccountsComplete
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		params, err := mapBalanceFilterToSQLC(filter)
		if err != nil {
			return err
		}

		sqlcAccounts, err := s.GetAccountBalancesList(ctx, params)
		if err != nil {
			return r.mapDatabaseError(err, "get_accounts_with_balances")
		}

		result = make([]*domain.ChartOfAccountsComplete, len(sqlcAccounts))
		for i, sqlcAccount := range sqlcAccounts {
			result[i] = mapSQLCAccountBalanceRowToChartOfAccountsComplete(*sqlcAccount)
		}

		return nil
	})

	return result, err
}

func (r *chartOfAccountsRepository) GetCashFlowAccounts(ctx context.Context, entityID *uuid.UUID) ([]*domain.CashFlowAccount, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetCashFlowAccounts")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result []*domain.CashFlowAccount
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		sqlcAccounts, err := s.GetCashFlowAccountsList(ctx, entityID)
		if err != nil {
			return r.mapDatabaseError(err, "get_cash_flow_accounts")
		}

		result = make([]*domain.CashFlowAccount, len(sqlcAccounts))
		for i, sqlcAccount := range sqlcAccounts {
			result[i] = mapSQLCCashFlowAccountToDomain(*sqlcAccount)
		}

		return nil
	})

	return result, err
}

func (r *chartOfAccountsRepository) GetAccountSummaryByGroup(ctx context.Context, entityID *uuid.UUID) ([]*domain.AccountGroupSummary, error) {
	ctx, span := r.tracing.StartSpan(ctx, "AccountsRepository.GetAccountSummaryByGroup")
	defer span.End()

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result []*domain.AccountGroupSummary
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		sqlcAccounts, err := s.GetAccountGroupSummary(ctx, entityID)
		if err != nil {
			return r.mapDatabaseError(err, "get_account_summary_by_group")
		}

		result = make([]*domain.AccountGroupSummary, len(sqlcAccounts))
		for i, sqlcAccount := range sqlcAccounts {
			result[i] = mapSQLCAccountGroupSummaryToDomain(*sqlcAccount)
		}

		return nil
	})

	return result, err
}

// Helper functions
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
