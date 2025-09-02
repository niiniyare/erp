package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/finance/domain"
)

// Account Domain to SQLC mappings

func mapDomainAccountToSQLCCreateDirect(account *domain.Accounts) (db.CreateAccountParams, error) {
	// Map root type string
	var rootType string
	switch account.RootType {
	case domain.RootTypeAsset:
		rootType = "ASSET"
	case domain.RootTypeLiability:
		rootType = "LIABILITY"
	case domain.RootTypeEquity:
		rootType = "EQUITY"
	case domain.RootTypeRevenue:
		rootType = "REVENUE"
	case domain.RootTypeExpense:
		rootType = "EXPENSE"
	default:
		return db.CreateAccountParams{}, fmt.Errorf("invalid root type: %s", account.RootType)
	}

	// Map normal balance string
	var normalBalance string
	switch account.NormalBalance {
	case domain.NormalBalanceDebit:
		normalBalance = "DEBIT"
	case domain.NormalBalanceCredit:
		normalBalance = "CREDIT"
	default:
		return db.CreateAccountParams{}, fmt.Errorf("invalid normal balance: %s", account.NormalBalance)
	}

	// Map account attributes to JSONB
	var attributes []byte
	if account.AccountAttributes != nil {
		var err error
		attributes, err = json.Marshal(account.AccountAttributes)
		if err != nil {
			return db.CreateAccountParams{}, fmt.Errorf("failed to marshal account attributes: %w", err)
		}
	} else {
		attributes = []byte("{}")
	}

	// Map decimal fields to pgtype.Numeric
	var budgetVarianceThreshold pgtype.Numeric
	if !account.BudgetVarianceThreshold.IsZero() {
		budgetVarianceThreshold = pgtype.Numeric{
			Int:   account.BudgetVarianceThreshold.BigInt(),
			Valid: true,
		}
	}

	var currentBalance pgtype.Numeric
	if !account.CurrentBalance.IsZero() {
		currentBalance = pgtype.Numeric{
			Int:   account.CurrentBalance.BigInt(),
			Valid: true,
		}
	}

	var ytdBalance pgtype.Numeric
	if !account.YTDBalance.IsZero() {
		ytdBalance = pgtype.Numeric{
			Int:   account.YTDBalance.BigInt(),
			Valid: true,
		}
	}

	// Map validation status
	var validationStatus string
	switch account.ValidationStatus {
	case domain.ValidationStatusPending:
		validationStatus = "PENDING"
	case domain.ValidationStatusValid:
		validationStatus = "VALID"
	case domain.ValidationStatusError:
		validationStatus = "ERROR"
	case domain.ValidationStatusWarning:
		validationStatus = "WARNING"
	default:
		validationStatus = "PENDING"
	}

	return db.CreateAccountParams{
		EntityID:                    account.EntityID,
		AccountCode:                 account.AccountCode,
		AccountName:                 account.AccountName,
		AccountDescription:          getStringValue(account.AccountDescription),
		ParentAccountID:             account.ParentAccountID,
		AccountLevel:                &account.AccountLevel,
		AccountPath:                 account.AccountPath,
		AccountCategory:             account.AccountCategory,
		SubCategory:                 account.SubCategory,
		DisplayOrder:                &account.DisplayOrder,
		ShowInReports:               &account.ShowInReports,
		ConsolidationAccount:        account.ConsolidationAccount,
		CashFlowType:                account.CashFlowType,
		RootType:                    rootType,
		AccountType:                 account.AccountType,
		AccountSubtype:              account.AccountSubtype,
		NormalBalance:               normalBalance,
		IsControlAccount:            boolPtr(account.IsControlAccount),
		ControlAccountID:            account.ControlAccountID,
		CurrencyCode:                account.CurrencyCode,
		IsMultiCurrency:             &account.IsMultiCurrency,
		CurrencyRevaluationRequired: &account.CurrencyRevaluationRequired,
		IsActive:                    boolPtr(account.IsActive),
		IsSystemAccount:             boolPtr(account.IsSystemAccount),
		AllowManualEntries:          boolPtr(account.AllowManualEntries),
		RequireReference:            boolPtr(account.RequireReference),
		CurrentBalance:              currentBalance,
		YtdBalance:                  ytdBalance,
		LastTransactionDate:         time.Time{}, // Zero time as default
		FinancialStatementLine:      account.FinancialStatementLine,
		ReportOrder:                 &account.ReportOrder,
		IsBudgetable:                &account.IsBudgetable,
		BudgetVarianceThreshold:     budgetVarianceThreshold,
		Version:                     &account.Version,
		LastValidationRun:           sql.NullTime{}, // Not set initially
		ValidationStatus:            &validationStatus,
		AccountAttributes:           attributes,
		HasChildren:                 &account.HasChildren,
		IsLeafAccount:               &account.IsLeafAccount,
		// CreatedBy:                   &account.CreatedBy, // Will be set by triggers or defaults
	}, nil
}

func mapDomainAccountToSQLCCreate(req *domain.CreateAccountRequest) (db.CreateAccountParams, error) {
	// Map root type string
	var rootType string
	switch req.RootType {
	case domain.RootTypeAsset:
		rootType = "ASSET"
	case domain.RootTypeLiability:
		rootType = "LIABILITY"
	case domain.RootTypeEquity:
		rootType = "EQUITY"
	case domain.RootTypeRevenue:
		rootType = "REVENUE"
	case domain.RootTypeExpense:
		rootType = "EXPENSE"
	default:
		return db.CreateAccountParams{}, fmt.Errorf("invalid root type: %s", req.RootType)
	}

	// Map normal balance string
	var normalBalance string
	switch req.NormalBalance {
	case domain.NormalBalanceDebit:
		normalBalance = "DEBIT"
	case domain.NormalBalanceCredit:
		normalBalance = "CREDIT"
	default:
		return db.CreateAccountParams{}, fmt.Errorf("invalid normal balance: %s", req.NormalBalance)
	}

	// Map account attributes to JSONB
	var attributes []byte
	if req.AccountAttributes != nil {
		var err error
		attributes, err = json.Marshal(req.AccountAttributes)
		if err != nil {
			return db.CreateAccountParams{}, fmt.Errorf("failed to marshal account attributes: %w", err)
		}
	} else {
		attributes = []byte("{}")
	}

	// Map budget variance threshold to pgtype.Numeric
	var budgetVarianceThreshold pgtype.Numeric
	if !req.BudgetVarianceThreshold.IsZero() {
		budgetVarianceThreshold = pgtype.Numeric{
			Int:   req.BudgetVarianceThreshold.BigInt(),
			Valid: true,
		}
	}

	return db.CreateAccountParams{
		EntityID:                    req.EntityID,
		AccountCode:                 req.AccountCode,
		AccountName:                 req.AccountName,
		AccountDescription:          getStringValue(req.AccountDescription),
		ParentAccountID:             req.ParentAccountID,
		RootType:                    rootType,
		AccountType:                 req.AccountType,
		AccountSubtype:              req.AccountSubtype,
		NormalBalance:               normalBalance,
		IsControlAccount:            boolPtr(req.IsControlAccount),
		ControlAccountID:            req.ControlAccountID,
		CurrencyCode:                req.CurrencyCode,
		IsMultiCurrency:             &req.IsMultiCurrency,
		CurrencyRevaluationRequired: &req.CurrencyRevaluationRequired,
		IsActive:                    boolPtr(req.IsActive),
		// IsSystemAccount not in CreateAccountRequest, default to false
		IsSystemAccount:         boolPtr(false),
		AllowManualEntries:      boolPtr(req.AllowManualEntries),
		RequireReference:        boolPtr(req.RequireReference),
		FinancialStatementLine:  req.FinancialStatementLine,
		ReportOrder:             &req.ReportOrder,
		IsBudgetable:            &req.IsBudgetable,
		BudgetVarianceThreshold: budgetVarianceThreshold,
		AccountAttributes:       attributes,
		// CreatedBy not in CreateAccountRequest, will be nil
		// Version not in CreateAccountRequest, will be nil
	}, nil
}

func mapSQLCAccountToDomain(sqlcAccount *db.FinanceAccount) (*domain.Accounts, error) {
	// Map root type string
	var rootType domain.RootType
	switch sqlcAccount.RootType {
	case "ASSET":
		rootType = domain.RootTypeAsset
	case "LIABILITY":
		rootType = domain.RootTypeLiability
	case "EQUITY":
		rootType = domain.RootTypeEquity
	case "REVENUE":
		rootType = domain.RootTypeRevenue
	case "EXPENSE":
		rootType = domain.RootTypeExpense
	default:
		return nil, fmt.Errorf("unknown root type: %s", sqlcAccount.RootType)
	}

	// Map normal balance string
	var normalBalance domain.NormalBalance
	switch sqlcAccount.NormalBalance {
	case "DEBIT":
		normalBalance = domain.NormalBalanceDebit
	case "CREDIT":
		normalBalance = domain.NormalBalanceCredit
	default:
		return nil, fmt.Errorf("unknown normal balance: %s", sqlcAccount.NormalBalance)
	}

	// Parse account attributes from JSONB
	var attributes map[string]interface{}
	if len(sqlcAccount.AccountAttributes) > 0 {
		if err := json.Unmarshal(sqlcAccount.AccountAttributes, &attributes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal account attributes: %w", err)
		}
	}

	// Convert balances from pgtype.Numeric to decimal.Decimal
	var currentBalance, ytdBalance decimal.Decimal
	if sqlcAccount.CurrentBalance.Valid {
		currentBalance = decimal.NewFromBigInt(sqlcAccount.CurrentBalance.Int, 0)
	}
	if sqlcAccount.YtdBalance.Valid {
		ytdBalance = decimal.NewFromBigInt(sqlcAccount.YtdBalance.Int, 0)
	}

	// Convert budget variance threshold
	var budgetVarianceThreshold *decimal.Decimal
	if sqlcAccount.BudgetVarianceThreshold.Valid {
		threshold := decimal.NewFromBigInt(sqlcAccount.BudgetVarianceThreshold.Int, 0)
		budgetVarianceThreshold = &threshold
	}

	return &domain.Accounts{
		ID:                          sqlcAccount.ID,
		TenantID:                    sqlcAccount.TenantID,
		EntityID:                    sqlcAccount.EntityID,
		AccountCode:                 sqlcAccount.AccountCode,
		AccountName:                 sqlcAccount.AccountName,
		AccountDescription:          &sqlcAccount.AccountDescription,
		ParentAccountID:             sqlcAccount.ParentAccountID,
		AccountLevel:                sqlcAccount.AccountLevel,
		AccountPath:                 sqlcAccount.AccountPath,
		RootType:                    rootType,
		AccountType:                 sqlcAccount.AccountType,
		AccountSubtype:              sqlcAccount.AccountSubtype,
		NormalBalance:               normalBalance,
		IsControlAccount:            sqlcAccount.IsControlAccount,
		ControlAccountID:            sqlcAccount.ControlAccountID,
		CurrencyCode:                sqlcAccount.CurrencyCode,
		IsMultiCurrency:             getBoolValue(sqlcAccount.IsMultiCurrency),
		CurrencyRevaluationRequired: getBoolValue(sqlcAccount.CurrencyRevaluationRequired),
		IsActive:                    sqlcAccount.IsActive,
		IsSystemAccount:             sqlcAccount.IsSystemAccount,
		AllowManualEntries:          sqlcAccount.AllowManualEntries,
		RequireReference:            sqlcAccount.RequireReference,
		CurrentBalance:              currentBalance,
		YTDBalance:                  ytdBalance,
		LastTransactionDate:         &sqlcAccount.LastTransactionDate,
		FinancialStatementLine:      sqlcAccount.FinancialStatementLine,
		ReportOrder:                 getInt32Value(sqlcAccount.ReportOrder),
		IsBudgetable:                getBoolValue(sqlcAccount.IsBudgetable),
		BudgetVarianceThreshold:     getDecimalValue(budgetVarianceThreshold),
		AccountAttributes:           attributes,
		CreatedAt:                   sqlcAccount.CreatedAt,
		UpdatedAt:                   sqlcAccount.UpdatedAt,
		DeletedAt:                   nullTimeToPointer(sqlcAccount.DeletedAt),
		CreatedBy:                   getUUIDValue(sqlcAccount.CreatedBy),
		UpdatedBy:                   sqlcAccount.UpdatedBy,
	}, nil
}

// Filter mappings

func mapAccountFilterToSQLCParams(filter *domain.AccountFilter) (db.ListAccountsParams, error) {
	params := db.ListAccountsParams{}

	// Handle pagination - convert from *int to int32
	if filter.Limit != nil {
		params.Limit = int32(*filter.Limit)
	} else {
		params.Limit = 50 // Default limit
	}

	if filter.Offset != nil {
		params.Offset = int32(*filter.Offset)
	}

	// Map root type filter to RootType
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

	// Map active status to IsActive
	if filter.IsActive != nil {
		params.IsActive = *filter.IsActive
	} else {
		params.IsActive = true // Default to active accounts
	}

	return params, nil
}

// Transaction Domain to SQLC mappings
func mapDomainTransactionToSQLCCreate(req *domain.CreateTransactionRequest) (db.CreateTransactionParams, error) {
	// Map transaction type string
	var transactionType string
	switch req.TransactionType {
	case domain.TransactionTypeManual:
		transactionType = "MANUAL"
	case domain.TransactionTypeSystem:
		transactionType = "SYSTEM"
	case domain.TransactionTypeImported:
		transactionType = "IMPORTED"
	case domain.TransactionTypeRecurring:
		transactionType = "RECURRING"
	case domain.TransactionTypeAdjustment:
		transactionType = "ADJUSTMENT"
	case domain.TransactionTypeClosing:
		transactionType = "CLOSING"
	case domain.TransactionTypeJournal:
		transactionType = "MANUAL" // Based on schema, JOURNAL maps to valid type
	case domain.TransactionTypeInvoice:
		transactionType = "MANUAL" // Based on schema, INVOICE maps to valid type
	case domain.TransactionTypePayment:
		transactionType = "MANUAL" // Based on schema, PAYMENT maps to valid type
	case domain.TransactionTypePurchase:
		transactionType = "MANUAL" // Based on schema, PURCHASE maps to valid type
	default:
		return db.CreateTransactionParams{}, fmt.Errorf("invalid transaction type: %s", req.TransactionType)
	}

	// Map transaction status string
	var status string
	switch req.TransactionStatus {
	case domain.TransactionStatusDraft:
		status = "DRAFT"
	case domain.TransactionStatusPendingApproval:
		status = "PENDING_APPROVAL"
	case domain.TransactionStatusApproved:
		status = "APPROVED"
	case domain.TransactionStatusPosted:
		status = "POSTED"
	case domain.TransactionStatusCancelled:
		status = "CANCELLED"
	case domain.TransactionStatusReversed:
		status = "REVERSED"
	default:
		return db.CreateTransactionParams{}, fmt.Errorf("invalid transaction status: %s", req.TransactionStatus)
	}

	// Calculate total amounts from entries (these would be calculated from the entries)
	// For now, set to zero - this should be calculated from req.Entries
	totalDebitAmount := pgtype.Numeric{
		Int:   decimal.Zero.BigInt(),
		Valid: true,
	}
	totalCreditAmount := pgtype.Numeric{
		Int:   decimal.Zero.BigInt(),
		Valid: true,
	}

	// Convert exchange rate to pgtype.Numeric
	var exchangeRate pgtype.Numeric
	if !req.ExchangeRate.IsZero() {
		exchangeRate = pgtype.Numeric{
			Int:   req.ExchangeRate.BigInt(),
			Valid: true,
		}
	}

	// Marshal metadata to JSONB
	var metadata []byte
	if req.TransactionAttributes != nil {
		var err error
		metadata, err = json.Marshal(req.TransactionAttributes)
		if err != nil {
			return db.CreateTransactionParams{}, fmt.Errorf("failed to marshal transaction metadata: %w", err)
		}
	} else {
		metadata = []byte("{}")
	}

	return db.CreateTransactionParams{
		EntityID:              req.EntityID,
		TransactionNumber:     req.TransactionNumber,
		TransactionType:       transactionType,
		TransactionStatus:     status,
		TransactionDate:       req.TransactionDate,
		PostingDate:           *req.PostingDate,
		Description:           req.Description,
		ReferenceNumber:       req.ReferenceNumber,
		CurrencyCode:          req.CurrencyCode,
		ExchangeRate:          exchangeRate,
		TotalDebitAmount:      totalDebitAmount,
		TotalCreditAmount:     totalCreditAmount,
		Memo:                  req.Memo,
		AttachmentIds:         req.AttachmentIds,
		Tags:                  req.Tags,
		TransactionAttributes: metadata,
		CreatedBy:             req.CreatedBy,
	}, nil
}

// Helper functions

func timeToPointer(t pgtype.Timestamptz) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func dateToPtr(t *time.Time) *pgtype.Date {

	if t != nil {
		return &pgtype.Date{Time: *t, Valid: true}
	}
	return nil
}

func stringPtr(s string) *string {
	return &s
}

func getBoolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func getInt32Value(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func getDecimalValue(d *decimal.Decimal) decimal.Decimal {
	if d == nil {
		return decimal.Zero
	}
	return *d
}

func nullTimeToPointer(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func getUUIDValue(u *uuid.UUID) uuid.UUID {
	if u == nil {
		return uuid.Nil
	}
	return *u
}

func pgTypeNumericToDecimal(n pgtype.Numeric) decimal.Decimal {
	if n.Valid {
		return decimal.NewFromBigInt(n.Int, 0)
	}
	return decimal.Zero
}

// Additional helper functions for transaction repository

func mapDomainTransactionTypeToSQLCEnum(transactionType domain.TransactionType) string {
	switch transactionType {
	case domain.TransactionTypeManual:
		return "MANUAL"
	case domain.TransactionTypeSystem:
		return "SYSTEM"
	case domain.TransactionTypeImported:
		return "IMPORTED"
	case domain.TransactionTypeRecurring:
		return "RECURRING"
	case domain.TransactionTypeAdjustment:
		return "ADJUSTMENT"
	case domain.TransactionTypeClosing:
		return "CLOSING"
	case domain.TransactionTypeJournal:
		return "MANUAL" // Based on schema, JOURNAL maps to valid type
	case domain.TransactionTypeInvoice:
		return "MANUAL" // Based on schema, INVOICE maps to valid type
	case domain.TransactionTypePayment:
		return "MANUAL" // Based on schema, PAYMENT maps to valid type
	case domain.TransactionTypePurchase:
		return "MANUAL" // Based on schema, PURCHASE maps to valid type
	default:
		return "MANUAL"
	}
}

func mapDomainTransactionStatusToSQLCEnum(status domain.TransactionStatus) string {
	switch status {
	case domain.TransactionStatusDraft:
		return "DRAFT"
	case domain.TransactionStatusPendingApproval:
		return "PENDING_APPROVAL"
	case domain.TransactionStatusApproved:
		return "APPROVED"
	case domain.TransactionStatusPosted:
		return "POSTED"
	case domain.TransactionStatusCancelled:
		return "CANCELLED"
	case domain.TransactionStatusReversed:
		return "REVERSED"
	default:
		return "DRAFT"
	}
}

func mapDomainTransactionStatusToSQLCEnumPtr(status domain.TransactionStatus) *string {
	sqlcEnum := mapDomainTransactionStatusToSQLCEnum(status)
	return &sqlcEnum
}

func mapSQLCTransactionTypeToDomain(sqlcType string) domain.TransactionType {
	switch sqlcType {
	case "MANUAL":
		return domain.TransactionTypeManual
	case "SYSTEM":
		return domain.TransactionTypeSystem
	case "IMPORTED":
		return domain.TransactionTypeImported
	case "RECURRING":
		return domain.TransactionTypeRecurring
	case "ADJUSTMENT":
		return domain.TransactionTypeAdjustment
	case "CLOSING":
		return domain.TransactionTypeClosing
	// Note: JOURNAL, INVOICE, PAYMENT, PURCHASE are not in the database schema
	// They would be stored as MANUAL type in the database
	default:
		return domain.TransactionTypeManual
	}
}

func mapSQLCTransactionStatusToDomain(sqlcStatus string) domain.TransactionStatus {
	switch sqlcStatus {
	case "DRAFT":
		return domain.TransactionStatusDraft
	case "PENDING_APPROVAL":
		return domain.TransactionStatusPendingApproval
	case "APPROVED":
		return domain.TransactionStatusApproved
	case "POSTED":
		return domain.TransactionStatusPosted
	case "CANCELLED":
		return domain.TransactionStatusCancelled
	case "REVERSED":
		return domain.TransactionStatusReversed
	default:
		return domain.TransactionStatusDraft
	}
}

func mapDomainApprovalStatusToString(status *domain.ApprovalStatus) *string {
	if status == nil {
		return nil
	}
	statusStr := string(*status)
	return &statusStr
}

func mapDomainApprovalStatusToStringPtr(status *domain.ApprovalStatus) *string {
	return mapDomainApprovalStatusToString(status)
}

func decimalToPgNumeric(d *decimal.Decimal) pgtype.Numeric {
	if d == nil {
		return pgtype.Numeric{}
	}
	return pgtype.Numeric{
		Int:   d.BigInt(),
		Valid: true,
	}
}

func pgNumericToDecimal(n pgtype.Numeric) decimal.Decimal {
	if n.Valid {
		return decimal.NewFromBigInt(n.Int, 0)
	}
	return decimal.Zero
}

func pgNumericToDecimalPtr(n pgtype.Numeric) *decimal.Decimal {
	if n.Valid {
		d := decimal.NewFromBigInt(n.Int, 0)
		return &d
	}
	return nil
}

func mapAttributesToJSON(attributes map[string]interface{}) []byte {
	if attributes == nil {
		return []byte("{}")
	}
	data, err := json.Marshal(attributes)
	if err != nil {
		return []byte("{}")
	}
	return data
}

func boolToPtr(b bool) *bool {
	return &b
}

func mapStringToApprovalStatus(s *string) *domain.ApprovalStatus {
	if s == nil {
		return nil
	}
	status := domain.ApprovalStatus(*s)
	return &status
}

// Additional helper functions for tests and general usage
func intPtr(i int) *int {
	return &i
}

func int32Ptr(i int32) *int32 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func getInt32Ptr(i int32) *int32 {
	return &i
}

func getBoolPtr(b bool) *bool {
	return &b
}

// Additional helper functions for transaction repository enum mappings

func mapDomainApprovalStatusToNullEnum(status *domain.ApprovalStatus) *string {
	if status == nil {
		return nil
	}

	var approvalStr string
	switch *status {
	case domain.ApprovalStatusNotRequired:
		approvalStr = "NOT_REQUIRED"
	case domain.ApprovalStatusPending:
		approvalStr = "PENDING"
	case domain.ApprovalStatusApproved:
		approvalStr = "APPROVED"
	case domain.ApprovalStatusRejected:
		approvalStr = "REJECTED"
	default:
		approvalStr = "NOT_REQUIRED"
	}

	return &approvalStr
}

func mapDomainRecurringFrequencyToNullEnum(frequency *string) *string {
	if frequency == nil || *frequency == "" {
		return nil
	}

	// Validate frequency values against database constraints
	switch *frequency {
	case "DAILY", "WEEKLY", "MONTHLY", "QUARTERLY", "YEARLY":
		return frequency
	default:
		return nil
	}
}

func timePointerToTimeValue(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func timePointerToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func mapNullApprovalStatusToDomain(approvalStatus *string) domain.ApprovalStatus {
	if approvalStatus == nil {
		return domain.ApprovalStatusNotRequired
	}

	switch *approvalStatus {
	case "NOT_REQUIRED":
		return domain.ApprovalStatusNotRequired
	case "PENDING":
		return domain.ApprovalStatusPending
	case "APPROVED":
		return domain.ApprovalStatusApproved
	case "REJECTED":
		return domain.ApprovalStatusRejected
	default:
		return domain.ApprovalStatusNotRequired
	}
}

func mapNullRecurringFrequencyToDomainString(recurringFreq *string) *string {
	if recurringFreq == nil {
		return nil
	}

	// Validate frequency values against database constraints
	switch *recurringFreq {
	case "DAILY", "WEEKLY", "MONTHLY", "QUARTERLY", "YEARLY":
		return recurringFreq
	default:
		return nil
	}
}

// Enhanced view-based mapper functions

// TODO: These mapper functions will be implemented once SQLC generates the types
// mapSQLCAccountWithGroupsToDomain maps SQLC AccountWithGroups to domain type
func mapSQLCAccountWithGroupsToDomain(sqlcAccount db.VFinanceAccountsWithGroup) (*domain.AccountWithGroups, error) {
	// Convert pgtype.Numeric to decimal.Decimal
	currentBalance := pgNumericToDecimal(sqlcAccount.CurrentBalance)

	// Map the base account fields from the view
	baseAccount := domain.Accounts{
		ID:                 sqlcAccount.ID,
		TenantID:           sqlcAccount.TenantID,
		AccountCode:        sqlcAccount.AccountCode,
		AccountName:        sqlcAccount.AccountName,
		AccountDescription: &sqlcAccount.AccountDescription,
		RootType:           mapSQLCRootTypeToDomain(sqlcAccount.RootType),
		AccountType:        sqlcAccount.AccountType,
		AccountCategory:    sqlcAccount.AccountCategory,
		NormalBalance:      mapSQLCNormalBalanceToDomain(sqlcAccount.NormalBalance),
		CurrentBalance:     currentBalance,
		IsActive:           sqlcAccount.IsActive,
		// Note: Many fields from the full account table are not available in the view
		// These would need to be set to defaults or fetched separately if needed
	}

	// Create the AccountWithGroups object
	account := &domain.AccountWithGroups{
		Accounts: baseAccount,
	}

	// Add group information
	account.GroupCode = sqlcAccount.GroupCode
	account.GroupName = sqlcAccount.GroupName
	account.GroupCategory = sqlcAccount.GroupCategory
	account.HeaderCode = sqlcAccount.HeaderCode
	// HeaderName is not in this view structure, would need to be fetched separately
	// For now, set a default value
	account.EffectiveDisplayOrder = 0 // Not available in this view

	return account, nil
}

// mapSQLCChartOfAccountsCompleteToDomain maps SQLC complete chart of accounts to domain
func mapSQLCChartOfAccountsCompleteToDomain(sqlcAccount db.VChartOfAccountsComplete) *domain.ChartOfAccountsComplete {
	currentBalance := pgNumericToDecimal(sqlcAccount.CurrentBalance)

	return &domain.ChartOfAccountsComplete{
		AccountID:              sqlcAccount.AccountID,
		AccountCode:            sqlcAccount.AccountCode,
		AccountName:            sqlcAccount.AccountName,
		RootType:               sqlcAccount.RootType,
		AccountType:            sqlcAccount.AccountType,
		NormalBalance:          sqlcAccount.NormalBalance,
		CurrentBalance:         currentBalance,
		GroupCode:              sqlcAccount.GroupCode,
		GroupName:              sqlcAccount.GroupName,
		GroupCategory:          sqlcAccount.GroupCategory,
		HeaderCode:             sqlcAccount.HeaderCode,
		HeaderName:             sqlcAccount.HeaderName,
		StatementSection:       sqlcAccount.StatementSection,
		CashFlowClassification: sqlcAccount.CashFlowClassification,
		DisplayOrder:           sqlcAccount.DisplayOrder,
		IncludeInReports:       sqlcAccount.IncludeInReports,
		IsActive:               sqlcAccount.IsActive,
		IsLeafAccount:          sqlcAccount.IsLeafAccount,
		TenantID:               sqlcAccount.TenantID,
		// CreatedAt and UpdatedAt are not available in this view
		CreatedAt: time.Now(), // Default value
		UpdatedAt: time.Now(), // Default value
	}
}

// mapSQLCTrialBalanceSummaryToDomain maps SQLC trial balance data to domain
func mapSQLCTrialBalanceSummaryToDomain(sqlcAccount db.GetTrialBalanceDataRow) *domain.TrialBalanceSummary {
	currentBalance := pgNumericToDecimal(sqlcAccount.CurrentBalance)

	return &domain.TrialBalanceSummary{
		AccountID:        sqlcAccount.AccountID,
		AccountCode:      sqlcAccount.AccountCode,
		AccountName:      sqlcAccount.AccountName,
		RootType:         sqlcAccount.RootType,
		NormalBalance:    sqlcAccount.NormalBalance,
		CurrentBalance:   currentBalance,
		GroupName:        sqlcAccount.GroupName,
		StatementSection: sqlcAccount.StatementSection,
	}
}

// mapSQLCCashFlowAccountToDomain maps SQLC cash flow account to domain
func mapSQLCCashFlowAccountToDomain(sqlcAccount db.GetCashFlowAccountsListRow) *domain.CashFlowAccount {
	currentBalance := pgNumericToDecimal(sqlcAccount.CurrentBalance)

	return &domain.CashFlowAccount{
		AccountID:              sqlcAccount.AccountID,
		AccountCode:            sqlcAccount.AccountCode,
		AccountName:            sqlcAccount.AccountName,
		CurrentBalance:         currentBalance,
		CashFlowClassification: sqlcAccount.CashFlowClassification,
		GroupName:              sqlcAccount.GroupName,
	}
}

// mapSQLCAccountGroupSummaryToDomain maps SQLC account group summary to domain
func mapSQLCAccountGroupSummaryToDomain(sqlcSummary db.GetAccountGroupSummaryRow) *domain.AccountGroupSummary {
	return &domain.AccountGroupSummary{
		GroupCode:          *sqlcSummary.GroupCode, // Convert from *string to string
		GroupName:          *sqlcSummary.GroupName, // Convert from *string to string
		GroupCategory:      sqlcSummary.GroupCategory,
		StatementSection:   sqlcSummary.StatementSection,
		AccountCount:       sqlcSummary.AccountCount,
		ActiveAccountCount: sqlcSummary.ActiveAccountCount,
		TotalBalance:       decimal.NewFromInt(sqlcSummary.TotalBalance),
		ActiveBalance:      decimal.NewFromInt(sqlcSummary.ActiveBalance),
	}
}

// Filter mapping functions

// mapAccountFilterToSQLCWithGroups maps domain AccountFilter to SQLC parameters for view queries
func mapAccountFilterToSQLCWithGroups(filter *domain.AccountFilter) (db.ListAccountsWithGroupsParams, error) {
	var rootType pgtype.Text
	if filter.RootType != nil {
		rootType = pgtype.Text{String: string(*filter.RootType), Valid: true}
	}

	var accountType pgtype.Text
	if filter.SearchTerm != nil {
		// For search functionality, we'll need to adapt this
		accountType = pgtype.Text{String: *filter.SearchTerm, Valid: true}
	}

	var isActive pgtype.Bool
	if filter.IsActive != nil {
		isActive = pgtype.Bool{Bool: *filter.IsActive, Valid: true}
	}

	limit := int32(50) // default
	if filter.Limit != nil {
		limit = int32(*filter.Limit)
	}

	offset := int32(0) // default
	if filter.Offset != nil {
		offset = int32(*filter.Offset)
	}

	return db.ListAccountsWithGroupsParams{
		EntityID:    filter.EntityID,
		RootType:    rootType.String,
		AccountType: accountType.String,
		IsActive:    isActive.Bool,
		IsLeafOnly:  false, // Not specified in base filter
		Limit:       limit,
		Offset:      offset,
	}, nil
}

// mapChartOfAccountsFilterToSQLC maps domain ChartOfAccountsFilter to SQLC parameters
func mapChartOfAccountsFilterToSQLC(filter *domain.ChartOfAccountsFilter) (db.GetChartOfAccountsCompleteParams, error) {
	var statementSection string
	if filter.StatementSection != nil {
		statementSection = string(*filter.StatementSection)
	}

	var includeInactive bool
	if filter.IncludeInactive != nil {
		includeInactive = *filter.IncludeInactive
	}

	var includeInReports bool
	if filter.IncludeInReports != nil {
		includeInReports = *filter.IncludeInReports
	}

	return db.GetChartOfAccountsCompleteParams{
		EntityID:         filter.EntityID,
		StatementSection: statementSection,
		IncludeInactive:  includeInactive,
		IncludeInReports: includeInReports,
	}, nil
}

// mapBalanceFilterToSQLC maps domain BalanceFilter to SQLC parameters
func mapBalanceFilterToSQLC(filter *domain.BalanceFilter) (db.GetAccountBalancesListParams, error) {
	var rootType string
	if filter.RootType != nil {
		rootType = string(*filter.RootType)
	}

	var nonZeroOnly bool
	if filter.NonZeroOnly != nil {
		nonZeroOnly = *filter.NonZeroOnly
	}

	return db.GetAccountBalancesListParams{
		EntityID:    filter.EntityID,
		NonZeroOnly: nonZeroOnly,
		RootType:    rootType,
	}, nil
}

// mapSQLCAccountBalanceRowToChartOfAccountsComplete maps SQLC balance row to complete chart
func mapSQLCAccountBalanceRowToChartOfAccountsComplete(sqlcRow db.GetAccountBalancesListRow) *domain.ChartOfAccountsComplete {
	currentBalance := pgNumericToDecimal(sqlcRow.CurrentBalance)

	return &domain.ChartOfAccountsComplete{
		AccountID:              sqlcRow.AccountID,
		AccountCode:            sqlcRow.AccountCode,
		AccountName:            sqlcRow.AccountName,
		CurrentBalance:         currentBalance,
		NormalBalance:          sqlcRow.NormalBalance,
		GroupName:              sqlcRow.GroupName,
		StatementSection:       sqlcRow.StatementSection,
		CashFlowClassification: sqlcRow.CashFlowClassification,
		// Other fields would be zero/default values since they're not in this query
	}
}

// Helper function to map SQLC normal balance to domain
func mapSQLCNormalBalanceToDomain(normalBalance string) domain.NormalBalance {
	switch normalBalance {
	case "DEBIT":
		return domain.NormalBalanceDebit
	case "CREDIT":
		return domain.NormalBalanceCredit
	default:
		return domain.NormalBalanceDebit // fallback
	}
}

// Helper function for null string to pointer conversion
func nullStringToPointer(ns pgtype.Text) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}
