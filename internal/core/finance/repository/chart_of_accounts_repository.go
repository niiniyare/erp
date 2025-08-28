package repository

import (
	"context"
	"database/sql"
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
	"github.com/niiniyare/erp/internal/shared/tracing"
)

type chartOfAccountsRepository struct {
	store   db.Store
	tracing tracing.TracingService
}

func NewChartOfAccountsRepository(store db.Store, tracing tracing.TracingService) domain.ChartOfAccountsRepository {
	return &chartOfAccountsRepository{
		store:   store,
		tracing: tracing,
	}
}

// NewAccountRepository is an alias for NewChartOfAccountsRepository for test compatibility
func NewAccountRepository(store db.Store, tracing tracing.TracingService) domain.ChartOfAccountsRepository {
	return NewChartOfAccountsRepository(store, tracing)
}

// Basic CRUD Operations

func (r *chartOfAccountsRepository) Create(ctx context.Context, account *domain.ChartOfAccounts) error {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.Create")
	defer span.End()

	// Create a CreateAccountRequest from the domain account
	req := &domain.CreateAccountRequest{
		EntityID:                    account.EntityID,
		AccountCode:                 account.AccountCode,
		AccountName:                 account.AccountName,
		AccountDescription:          account.AccountDescription,
		ParentAccountID:             account.ParentAccountID,
		RootType:                    account.RootType,
		AccountType:                 account.AccountType,
		AccountSubtype:              account.AccountSubtype,
		NormalBalance:               account.NormalBalance,
		IsControlAccount:            account.IsControlAccount,
		ControlAccountID:            account.ControlAccountID,
		CurrencyCode:                account.CurrencyCode,
		IsMultiCurrency:             account.IsMultiCurrency,
		CurrencyRevaluationRequired: account.CurrencyRevaluationRequired,
		IsActive:                    account.IsActive,
		AllowManualEntries:          account.AllowManualEntries,
		RequireReference:            account.RequireReference,
		FinancialStatementLine:      account.FinancialStatementLine,
		ReportOrder:                 account.ReportOrder,
		IsBudgetable:                account.IsBudgetable,
		BudgetVarianceThreshold:     account.BudgetVarianceThreshold,
		AccountAttributes:           account.AccountAttributes,
	}

	// Map domain request to SQLC parameters
	params, err := mapDomainAccountToSQLCCreate(req)
	if err != nil {
		return fmt.Errorf("failed to map create account request: %w", err)
	}

	// Execute SQLC query
	sqlcAccount, err := r.store.CreateAccount(ctx, params)
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
}

func (r *chartOfAccountsRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ChartOfAccounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.GetByID")
	defer span.End()

	sqlcAccount, err := r.store.GetAccountByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAccountNotFound
		}
		return nil, r.mapDatabaseError(err, "get_account_by_id")
	}

	account, err := mapSQLCAccountToDomain(sqlcAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to map SQLC account to domain: %w", err)
	}

	return account, nil
}

func (r *chartOfAccountsRepository) GetByCode(ctx context.Context, entityID *uuid.UUID, accountCode string) (*domain.ChartOfAccounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.GetByCode")
	defer span.End()

	// Note: SQLC GetAccountByCode only takes accountCode string, entityID filtering handled by RLS
	sqlcAccount, err := r.store.GetAccountByCode(ctx, accountCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAccountNotFound
		}
		return nil, r.mapDatabaseError(err, "get_account_by_code")
	}

	account, err := mapSQLCAccountToDomain(sqlcAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to map SQLC account to domain: %w", err)
	}

	return account, nil
}

func (r *chartOfAccountsRepository) Update(ctx context.Context, account *domain.ChartOfAccounts) error {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.Update")
	defer span.End()

	// Map account to SQLC parameters
	params := db.UpdateAccountParams{
		AccountID:              account.ID,
		AccountName:            &account.AccountName,
		AccountDescription:     getStringValue(account.AccountDescription),
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

	// Execute SQLC update query
	_, err := r.store.UpdateAccount(ctx, params)
	if err != nil {
		return r.mapDatabaseError(err, "update_account")
	}

	return nil
}

func (r *chartOfAccountsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.Delete")
	defer span.End()

	params := db.SoftDeleteAccountParams{
		AccountID: id,
		UpdatedBy: nil, // TODO: Get from context
	}
	err := r.store.SoftDeleteAccount(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrAccountNotFound
		}
		return r.mapDatabaseError(err, "soft_delete_account")
	}

	return nil
}

// List and Filter Operations

func (r *chartOfAccountsRepository) List(ctx context.Context, filter *domain.AccountFilter) ([]*domain.ChartOfAccounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.List")
	defer span.End()

	// Map domain filter to SQLC parameters
	params, err := mapAccountFilterToSQLCParams(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to map account filter: %w", err)
	}

	// Execute SQLC query
	sqlcAccounts, err := r.store.ListAccounts(ctx, params)
	if err != nil {
		return nil, r.mapDatabaseError(err, "list_accounts")
	}

	// Map results to domain models
	accounts := make([]*domain.ChartOfAccounts, 0, len(sqlcAccounts))
	for _, sqlcAccount := range sqlcAccounts {
		account, err := mapSQLCAccountToDomain(sqlcAccount)
		if err != nil {
			return nil, fmt.Errorf("failed to map SQLC account to domain: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (r *chartOfAccountsRepository) Count(ctx context.Context, filter *domain.AccountFilter) (int64, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.Count")
	defer span.End()

	// Map basic filter parameters for count
	params := db.CountAccountsParams{}

	if filter.RootType != nil {
		switch *filter.RootType {
		case domain.RootTypeAsset:
			params.RootType = db.NullRootTypeEnum{RootTypeEnum: db.RootTypeEnumASSET, Valid: true}
		case domain.RootTypeLiability:
			params.RootType = db.NullRootTypeEnum{RootTypeEnum: db.RootTypeEnumLIABILITY, Valid: true}
		case domain.RootTypeEquity:
			params.RootType = db.NullRootTypeEnum{RootTypeEnum: db.RootTypeEnumEQUITY, Valid: true}
		case domain.RootTypeRevenue:
			params.RootType = db.NullRootTypeEnum{RootTypeEnum: db.RootTypeEnumREVENUE, Valid: true}
		case domain.RootTypeExpense:
			params.RootType = db.NullRootTypeEnum{RootTypeEnum: db.RootTypeEnumEXPENSE, Valid: true}
		}
	}

	// Note: AccountType filter not available in domain.AccountFilter

	if filter.IsActive != nil {
		params.IsActive = *filter.IsActive
	}

	count, err := r.store.CountAccounts(ctx, params)
	if err != nil {
		return 0, r.mapDatabaseError(err, "count_accounts")
	}

	return count, nil
}

func (r *chartOfAccountsRepository) ListByParent(ctx context.Context, parentID uuid.UUID) ([]*domain.ChartOfAccounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.ListByParent")
	defer span.End()

	sqlcAccounts, err := r.store.ListAccountsByParent(ctx, &parentID)
	if err != nil {
		return nil, r.mapDatabaseError(err, "list_accounts_by_parent")
	}

	accounts := make([]*domain.ChartOfAccounts, 0, len(sqlcAccounts))
	for _, sqlcAccount := range sqlcAccounts {
		account, err := mapSQLCAccountToDomain(sqlcAccount)
		if err != nil {
			return nil, fmt.Errorf("failed to map SQLC account to domain: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// ListByRootType retrieves accounts by root type
func (r *chartOfAccountsRepository) ListByRootType(ctx context.Context, rootType domain.RootType) ([]*domain.ChartOfAccounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.ListByRootType")
	defer span.End()

	// Create filter with root type
	filter := &domain.AccountFilter{
		RootType: &rootType,
	}

	return r.List(ctx, filter)
}

// Hierarchy Operations

func (r *chartOfAccountsRepository) GetAccountHierarchy(ctx context.Context, rootID uuid.UUID) ([]*domain.ChartOfAccounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.GetAccountHierarchy")
	defer span.End()

	// Convert rootID to string for hierarchy query
	rootIDStr := rootID.String()
	sqlcAccounts, err := r.store.GetAccountHierarchy(ctx, rootIDStr)
	if err != nil {
		return nil, r.mapDatabaseError(err, "get_account_hierarchy")
	}

	accounts := make([]*domain.ChartOfAccounts, 0, len(sqlcAccounts))
	for _, sqlcAccount := range sqlcAccounts {
		account, err := mapSQLCAccountToDomain(sqlcAccount)
		if err != nil {
			return nil, fmt.Errorf("failed to map SQLC account to domain: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// GetAccountPath retrieves the full path of accounts from root to specified account
func (r *chartOfAccountsRepository) GetAccountPath(ctx context.Context, accountID uuid.UUID) ([]domain.ChartOfAccounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.GetAccountPath")
	defer span.End()

	// Get the account first
	account, err := r.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// If no path is set, return just the account itself
	if account.AccountPath == nil || *account.AccountPath == "" {
		return []domain.ChartOfAccounts{*account}, nil
	}

	// TODO: Parse account path and retrieve all accounts in the path
	// For now, just return the single account
	return []domain.ChartOfAccounts{*account}, nil
}

// ValidateHierarchy validates if the parent-child relationship is valid
func (r *chartOfAccountsRepository) ValidateHierarchy(ctx context.Context, accountID, parentID uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.ValidateHierarchy")
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
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.GetAccountBalance")
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
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.GetAccountBalances")
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
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.GetTrialBalance")
	defer span.End()

	// TODO: Implement proper trial balance calculation
	// For now, return empty result
	entries := make([]*domain.TrialBalanceEntry, 0)

	return entries, nil
}

// Business Logic Queries

func (r *chartOfAccountsRepository) GetActiveAccounts(ctx context.Context, entityID *uuid.UUID) ([]*domain.ChartOfAccounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.GetActiveAccounts")
	defer span.End()

	// Create filter for active accounts only
	filter := &domain.AccountFilter{
		EntityID: entityID,
		IsActive: &[]bool{true}[0],
	}

	return r.List(ctx, filter)
}

func (r *chartOfAccountsRepository) GetControlAccounts(ctx context.Context, entityID *uuid.UUID) ([]*domain.ChartOfAccounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.GetControlAccounts")
	defer span.End()

	// Note: GetControlAccounts doesn't take entityID parameter, uses tenant context
	sqlcAccounts, err := r.store.GetControlAccounts(ctx)
	if err != nil {
		return nil, r.mapDatabaseError(err, "get_control_accounts")
	}

	accounts := make([]*domain.ChartOfAccounts, 0, len(sqlcAccounts))
	for _, sqlcAccount := range sqlcAccounts {
		account, err := mapSQLCAccountToDomain(sqlcAccount)
		if err != nil {
			return nil, fmt.Errorf("failed to map SQLC account to domain: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (r *chartOfAccountsRepository) GetAccountsByType(ctx context.Context, accountType string, rootType *domain.RootType) ([]*domain.ChartOfAccounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.GetAccountsByType")
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
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.ValidateAccountCode")
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
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.IsAccountCodeUnique")
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
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.HasChildren")
	defer span.End()

	// Use ListAccountsByParent to check if children exist
	children, err := r.store.ListAccountsByParent(ctx, &accountID)
	if err != nil {
		return false, r.mapDatabaseError(err, "list_accounts_by_parent")
	}

	return len(children) > 0, nil
}

func (r *chartOfAccountsRepository) HasTransactions(ctx context.Context, accountID uuid.UUID) (bool, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.HasTransactions")
	defer span.End()

	// TODO: Implement by checking transaction entries table
	// For now, return false
	return false, nil
}

func (r *chartOfAccountsRepository) Search(ctx context.Context, query string, limit int) ([]*domain.ChartOfAccounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.Search")
	defer span.End()

	// For now, return empty result
	return make([]*domain.ChartOfAccounts, 0), nil
}

func (r *chartOfAccountsRepository) UpdateBalance(ctx context.Context, accountID uuid.UUID, balance domain.AccountBalance) error {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.UpdateBalance")
	defer span.End()

	// For now, do nothing
	return nil
}

func (r *chartOfAccountsRepository) GetChildren(ctx context.Context, accountID uuid.UUID) ([]*domain.ChartOfAccounts, error) {
	ctx, span := r.tracing.StartSpan(ctx, "ChartOfAccountsRepository.GetChildren")
	defer span.End()

	return r.ListByParent(ctx, accountID)
}

// Error mapping helper
func (r *chartOfAccountsRepository) mapDatabaseError(err error, operation string) error {
	// Convert database-specific errors to domain errors
	if err == sql.ErrNoRows {
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
			return fmt.Errorf("invalid parent account reference")
		}
	}

	// Default to internal server error
	return fmt.Errorf("database operation failed: %s: %w", operation, err)
}

// Helper function to map SQLC root type enum to domain
func mapSQLCRootTypeToDomain(sqlcRootType db.RootTypeEnum) domain.RootType {
	switch sqlcRootType {
	case db.RootTypeEnumASSET:
		return domain.RootTypeAsset
	case db.RootTypeEnumLIABILITY:
		return domain.RootTypeLiability
	case db.RootTypeEnumEQUITY:
		return domain.RootTypeEquity
	case db.RootTypeEnumREVENUE:
		return domain.RootTypeRevenue
	case db.RootTypeEnumEXPENSE:
		return domain.RootTypeExpense
	default:
		return domain.RootTypeAsset // fallback
	}
}

// Helper functions
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
