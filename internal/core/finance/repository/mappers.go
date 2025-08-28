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

func mapDomainAccountToSQLCCreate(req *domain.CreateAccountRequest) (db.CreateAccountParams, error) {
	// Map root type enum
	var rootType db.RootTypeEnum
	switch req.RootType {
	case domain.RootTypeAsset:
		rootType = db.RootTypeEnumASSET
	case domain.RootTypeLiability:
		rootType = db.RootTypeEnumLIABILITY
	case domain.RootTypeEquity:
		rootType = db.RootTypeEnumEQUITY
	case domain.RootTypeRevenue:
		rootType = db.RootTypeEnumREVENUE
	case domain.RootTypeExpense:
		rootType = db.RootTypeEnumEXPENSE
	default:
		return db.CreateAccountParams{}, fmt.Errorf("invalid root type: %s", req.RootType)
	}

	// Map normal balance enum
	var normalBalance db.NormalBalanceEnum
	switch req.NormalBalance {
	case domain.NormalBalanceDebit:
		normalBalance = db.NormalBalanceEnumDEBIT
	case domain.NormalBalanceCredit:
		normalBalance = db.NormalBalanceEnumCREDIT
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
		IsControlAccount:            req.IsControlAccount,
		ControlAccountID:            req.ControlAccountID,
		CurrencyCode:                req.CurrencyCode,
		IsMultiCurrency:             &req.IsMultiCurrency,
		CurrencyRevaluationRequired: &req.CurrencyRevaluationRequired,
		IsActive:                    req.IsActive,
		// IsSystemAccount not in CreateAccountRequest, default to false
		IsSystemAccount:         false,
		AllowManualEntries:      req.AllowManualEntries,
		RequireReference:        req.RequireReference,
		FinancialStatementLine:  req.FinancialStatementLine,
		ReportOrder:             &req.ReportOrder,
		IsBudgetable:            &req.IsBudgetable,
		BudgetVarianceThreshold: budgetVarianceThreshold,
		AccountAttributes:       attributes,
		// CreatedBy not in CreateAccountRequest, will be nil
	}, nil
}

func mapSQLCAccountToDomain(sqlcAccount *db.FinanceChartOfAccount) (*domain.Accounts, error) {
	// Map root type enum
	var rootType domain.RootType
	switch sqlcAccount.RootType {
	case db.RootTypeEnumASSET:
		rootType = domain.RootTypeAsset
	case db.RootTypeEnumLIABILITY:
		rootType = domain.RootTypeLiability
	case db.RootTypeEnumEQUITY:
		rootType = domain.RootTypeEquity
	case db.RootTypeEnumREVENUE:
		rootType = domain.RootTypeRevenue
	case db.RootTypeEnumEXPENSE:
		rootType = domain.RootTypeExpense
	default:
		return nil, fmt.Errorf("unknown root type: %s", sqlcAccount.RootType)
	}

	// Map normal balance enum
	var normalBalance domain.NormalBalance
	switch sqlcAccount.NormalBalance {
	case db.NormalBalanceEnumDEBIT:
		normalBalance = domain.NormalBalanceDebit
	case db.NormalBalanceEnumCREDIT:
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
	// Map transaction type enum
	var transactionType db.TransactionTypeEnum
	switch req.TransactionType {
	case domain.TransactionTypeManual:
		transactionType = db.TransactionTypeEnumMANUAL
	case domain.TransactionTypeSystem:
		transactionType = db.TransactionTypeEnumSYSTEM
	case domain.TransactionTypeImported:
		transactionType = db.TransactionTypeEnumIMPORTED
	case domain.TransactionTypeRecurring:
		transactionType = db.TransactionTypeEnumRECURRING
	case domain.TransactionTypeAdjustment:
		transactionType = db.TransactionTypeEnumADJUSTMENT
	case domain.TransactionTypeClosing:
		transactionType = db.TransactionTypeEnumCLOSING
	case domain.TransactionTypeJournal:
		transactionType = db.TransactionTypeEnumJOURNAL
	case domain.TransactionTypeInvoice:
		transactionType = db.TransactionTypeEnumINVOICE
	case domain.TransactionTypePayment:
		transactionType = db.TransactionTypeEnumPAYMENT
	case domain.TransactionTypePurchase:
		transactionType = db.TransactionTypeEnumPURCHASE
	default:
		return db.CreateTransactionParams{}, fmt.Errorf("invalid transaction type: %s", req.TransactionType)
	}

	// Map transaction status enum
	var status db.TransactionStatusEnum
	switch req.TransactionStatus {
	case domain.TransactionStatusDraft:
		status = db.TransactionStatusEnumDRAFT
	case domain.TransactionStatusPendingApproval:
		status = db.TransactionStatusEnumPENDINGAPPROVAL
	case domain.TransactionStatusApproved:
		status = db.TransactionStatusEnumAPPROVED
	case domain.TransactionStatusPosted:
		status = db.TransactionStatusEnumPOSTED
	case domain.TransactionStatusCancelled:
		status = db.TransactionStatusEnumCANCELLED
	case domain.TransactionStatusReversed:
		status = db.TransactionStatusEnumREVERSED
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

func mapDomainTransactionTypeToSQLCEnum(transactionType domain.TransactionType) db.TransactionTypeEnum {
	switch transactionType {
	case domain.TransactionTypeManual:
		return db.TransactionTypeEnumMANUAL
	case domain.TransactionTypeSystem:
		return db.TransactionTypeEnumSYSTEM
	case domain.TransactionTypeImported:
		return db.TransactionTypeEnumIMPORTED
	case domain.TransactionTypeRecurring:
		return db.TransactionTypeEnumRECURRING
	case domain.TransactionTypeAdjustment:
		return db.TransactionTypeEnumADJUSTMENT
	case domain.TransactionTypeClosing:
		return db.TransactionTypeEnumCLOSING
	case domain.TransactionTypeJournal:
		return db.TransactionTypeEnumJOURNAL
	case domain.TransactionTypeInvoice:
		return db.TransactionTypeEnumINVOICE
	case domain.TransactionTypePayment:
		return db.TransactionTypeEnumPAYMENT
	case domain.TransactionTypePurchase:
		return db.TransactionTypeEnumPURCHASE
	default:
		return db.TransactionTypeEnumMANUAL
	}
}

func mapDomainTransactionStatusToSQLCEnum(status domain.TransactionStatus) db.TransactionStatusEnum {
	switch status {
	case domain.TransactionStatusDraft:
		return db.TransactionStatusEnumDRAFT
	case domain.TransactionStatusPendingApproval:
		return db.TransactionStatusEnumPENDINGAPPROVAL
	case domain.TransactionStatusApproved:
		return db.TransactionStatusEnumAPPROVED
	case domain.TransactionStatusPosted:
		return db.TransactionStatusEnumPOSTED
	case domain.TransactionStatusCancelled:
		return db.TransactionStatusEnumCANCELLED
	case domain.TransactionStatusReversed:
		return db.TransactionStatusEnumREVERSED
	default:
		return db.TransactionStatusEnumDRAFT
	}
}

func mapDomainTransactionStatusToSQLCEnumPtr(status domain.TransactionStatus) *db.TransactionStatusEnum {
	sqlcEnum := mapDomainTransactionStatusToSQLCEnum(status)
	return &sqlcEnum
}

func mapSQLCTransactionTypeToDomain(sqlcType db.TransactionTypeEnum) domain.TransactionType {
	switch sqlcType {
	case db.TransactionTypeEnumMANUAL:
		return domain.TransactionTypeManual
	case db.TransactionTypeEnumSYSTEM:
		return domain.TransactionTypeSystem
	case db.TransactionTypeEnumIMPORTED:
		return domain.TransactionTypeImported
	case db.TransactionTypeEnumRECURRING:
		return domain.TransactionTypeRecurring
	case db.TransactionTypeEnumADJUSTMENT:
		return domain.TransactionTypeAdjustment
	case db.TransactionTypeEnumCLOSING:
		return domain.TransactionTypeClosing
	case db.TransactionTypeEnumJOURNAL:
		return domain.TransactionTypeJournal
	case db.TransactionTypeEnumINVOICE:
		return domain.TransactionTypeInvoice
	case db.TransactionTypeEnumPAYMENT:
		return domain.TransactionTypePayment
	case db.TransactionTypeEnumPURCHASE:
		return domain.TransactionTypePurchase
	default:
		return domain.TransactionTypeManual
	}
}

func mapSQLCTransactionStatusToDomain(sqlcStatus db.TransactionStatusEnum) domain.TransactionStatus {
	switch sqlcStatus {
	case db.TransactionStatusEnumDRAFT:
		return domain.TransactionStatusDraft
	case db.TransactionStatusEnumPENDINGAPPROVAL:
		return domain.TransactionStatusPendingApproval
	case db.TransactionStatusEnumAPPROVED:
		return domain.TransactionStatusApproved
	case db.TransactionStatusEnumPOSTED:
		return domain.TransactionStatusPosted
	case db.TransactionStatusEnumCANCELLED:
		return domain.TransactionStatusCancelled
	case db.TransactionStatusEnumREVERSED:
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
//
//	func stringPtr(s string) *string {
//		return &s
//	}
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

func mapDomainApprovalStatusToNullEnum(status *domain.ApprovalStatus) db.NullApprovalStatusEnum {
	if status == nil {
		return db.NullApprovalStatusEnum{Valid: false}
	}

	var approvalEnum db.ApprovalStatusEnum
	switch *status {
	case domain.ApprovalStatusNotRequired:
		approvalEnum = db.ApprovalStatusEnumNOTREQUIRED
	case domain.ApprovalStatusPending:
		approvalEnum = db.ApprovalStatusEnumPENDING
	case domain.ApprovalStatusApproved:
		approvalEnum = db.ApprovalStatusEnumAPPROVED
	case domain.ApprovalStatusRejected:
		approvalEnum = db.ApprovalStatusEnumREJECTED
	default:
		approvalEnum = db.ApprovalStatusEnumNOTREQUIRED
	}

	return db.NullApprovalStatusEnum{
		ApprovalStatusEnum: approvalEnum,
		Valid:              true,
	}
}

func mapDomainRecurringFrequencyToNullEnum(frequency *string) db.NullRecurringFrequencyEnum {
	if frequency == nil || *frequency == "" {
		return db.NullRecurringFrequencyEnum{Valid: false}
	}

	var frequencyEnum db.RecurringFrequencyEnum
	switch *frequency {
	case "DAILY":
		frequencyEnum = db.RecurringFrequencyEnumDAILY
	case "WEEKLY":
		frequencyEnum = db.RecurringFrequencyEnumWEEKLY
	case "MONTHLY":
		frequencyEnum = db.RecurringFrequencyEnumMONTHLY
	case "QUARTERLY":
		frequencyEnum = db.RecurringFrequencyEnumQUARTERLY
	case "YEARLY":
		frequencyEnum = db.RecurringFrequencyEnumYEARLY
	default:
		return db.NullRecurringFrequencyEnum{Valid: false}
	}

	return db.NullRecurringFrequencyEnum{
		RecurringFrequencyEnum: frequencyEnum,
		Valid:                  true,
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

func mapNullApprovalStatusToDomain(nullStatus db.NullApprovalStatusEnum) domain.ApprovalStatus {
	if !nullStatus.Valid {
		return domain.ApprovalStatusNotRequired
	}

	switch nullStatus.ApprovalStatusEnum {
	case db.ApprovalStatusEnumNOTREQUIRED:
		return domain.ApprovalStatusNotRequired
	case db.ApprovalStatusEnumPENDING:
		return domain.ApprovalStatusPending
	case db.ApprovalStatusEnumAPPROVED:
		return domain.ApprovalStatusApproved
	case db.ApprovalStatusEnumREJECTED:
		return domain.ApprovalStatusRejected
	default:
		return domain.ApprovalStatusNotRequired
	}
}

func mapNullRecurringFrequencyToDomainString(nullFreq db.NullRecurringFrequencyEnum) *string {
	if !nullFreq.Valid {
		return nil
	}

	var freqStr string
	switch nullFreq.RecurringFrequencyEnum {
	case db.RecurringFrequencyEnumDAILY:
		freqStr = "DAILY"
	case db.RecurringFrequencyEnumWEEKLY:
		freqStr = "WEEKLY"
	case db.RecurringFrequencyEnumMONTHLY:
		freqStr = "MONTHLY"
	case db.RecurringFrequencyEnumQUARTERLY:
		freqStr = "QUARTERLY"
	case db.RecurringFrequencyEnumYEARLY:
		freqStr = "YEARLY"
	default:
		return nil
	}

	return &freqStr
}
