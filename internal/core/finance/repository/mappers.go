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

func mapSQLCAccountToDomain(sqlcAccount *db.FinanceChartOfAccount) (*domain.ChartOfAccounts, error) {
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

	return &domain.ChartOfAccounts{
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

	// Convert amounts to pgtype.Numeric
	totalDebitAmount := pgtype.Numeric{
		Int:   req.TotalDebitAmount.BigInt(),
		Valid: true,
	}
	totalCreditAmount := pgtype.Numeric{
		Int:   req.TotalCreditAmount.BigInt(),
		Valid: true,
	}

	// Convert exchange rate to pgtype.Numeric
	var exchangeRate pgtype.Numeric
	if req.ExchangeRate != nil {
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

func mapSQLCTransactionToDomain(sqlcTransaction *db.FinanceTransaction) (*domain.Transaction, error) {
	// Map transaction type enum
	var transactionType domain.TransactionTypeEnum
	switch sqlcTransaction.TransactionType {
	case db.TransactionTypeEnumMANUAL:
		transactionType = domain.MANUAL
	case db.TransactionTypeEnumSALESINVOICE:
		transactionType = domain.SALES_INVOICE
	case db.TransactionTypeEnumPURCHASEINVOICE:
		transactionType = domain.PURCHASE_INVOICE
	case db.TransactionTypeEnumPAYMENT:
		transactionType = domain.PAYMENT
	case db.TransactionTypeEnumRECEIPT:
		transactionType = domain.RECEIPT
	case db.TransactionTypeEnumJOURNALENTRY:
		transactionType = domain.JOURNAL_ENTRY
	case db.TransactionTypeEnumBANKTRANSFER:
		transactionType = domain.BANK_TRANSFER
	case db.TransactionTypeEnumADJUSTMENT:
		transactionType = domain.ADJUSTMENT
	case db.TransactionTypeEnumOPENINGBALANCE:
		transactionType = domain.OPENING_BALANCE
	case db.TransactionTypeEnumCLOSINGENTRY:
		transactionType = domain.CLOSING_ENTRY
	default:
		return nil, fmt.Errorf("unknown transaction type: %s", sqlcTransaction.TransactionType)
	}

	// Map transaction status enum
	var status domain.TransactionStatusEnum
	switch sqlcTransaction.TransactionStatus {
	case db.TransactionStatusEnumDRAFT:
		status = domain.DRAFT
	case db.TransactionStatusEnumPENDINGAPPROVAL:
		status = domain.PENDING_APPROVAL
	case db.TransactionStatusEnumAPPROVED:
		status = domain.APPROVED
	case db.TransactionStatusEnumPOSTED:
		status = domain.POSTED
	case db.TransactionStatusEnumCANCELLED:
		status = domain.CANCELLED
	case db.TransactionStatusEnumREVERSED:
		status = domain.REVERSED
	default:
		return nil, fmt.Errorf("unknown transaction status: %s", sqlcTransaction.TransactionStatus)
	}

	// Convert amounts from pgtype.Numeric to decimal.Decimal
	var totalDebitAmount, totalCreditAmount decimal.Decimal
	if sqlcTransaction.TotalDebitAmount.Valid {
		totalDebitAmount = decimal.NewFromBigInt(sqlcTransaction.TotalDebitAmount.Int, 0)
	}
	if sqlcTransaction.TotalCreditAmount.Valid {
		totalCreditAmount = decimal.NewFromBigInt(sqlcTransaction.TotalCreditAmount.Int, 0)
	}

	// Convert exchange rate
	var exchangeRate *decimal.Decimal
	if sqlcTransaction.ExchangeRate.Valid {
		rate := decimal.NewFromBigInt(sqlcTransaction.ExchangeRate.Int, 0)
		exchangeRate = &rate
	}

	// Parse metadata from JSONB
	var metadata map[string]interface{}
	if len(sqlcTransaction.Metadata) > 0 {
		if err := json.Unmarshal(sqlcTransaction.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal transaction metadata: %w", err)
		}
	}

	return &domain.Transaction{
		ID:                sqlcTransaction.ID,
		TenantID:          sqlcTransaction.TenantID,
		EntityID:          sqlcTransaction.EntityID,
		TransactionNumber: sqlcTransaction.TransactionNumber,
		TransactionType:   transactionType,
		Status:            status,
		TransactionDate:   sqlcTransaction.TransactionDate.Time,
		PostingDate:       timeToPointer(sqlcTransaction.PostingDate),
		Description:       sqlcTransaction.Description,
		ReferenceNumber:   sqlcTransaction.ReferenceNumber,
		CurrencyCode:      sqlcTransaction.CurrencyCode,
		ExchangeRate:      exchangeRate,
		TotalDebitAmount:  totalDebitAmount,
		TotalCreditAmount: totalCreditAmount,
		Memo:              sqlcTransaction.Memo,
		AttachmentIDs:     sqlcTransaction.AttachmentIds,
		Tags:              sqlcTransaction.Tags,
		Metadata:          metadata,
		CreatedAt:         sqlcTransaction.CreatedAt.Time,
		UpdatedAt:         sqlcTransaction.UpdatedAt.Time,
		DeletedAt:         timeToPointer(sqlcTransaction.DeletedAt),
		CreatedBy:         sqlcTransaction.CreatedBy,
		UpdatedBy:         sqlcTransaction.UpdatedBy,
		PostedBy:          sqlcTransaction.PostedBy,
		PostedAt:          timeToPointer(sqlcTransaction.PostedAt),
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

func mapTransactionFilterToSQLCParams(filter *domain.TransactionFilter) (db.ListTransactionsParams, error) {
	params := db.ListTransactionsParams{
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}

	// Map date range filters
	if filter.FromDate != nil {
		params.FromDate = &pgtype.Date{Time: *filter.FromDate, Valid: true}
	}

	if filter.ToDate != nil {
		params.ToDate = &pgtype.Date{Time: *filter.ToDate, Valid: true}
	}

	// Map status filter
	if filter.Status != nil {
		switch *filter.Status {
		case domain.DRAFT:
			params.Status = stringPtr("DRAFT")
		case domain.PENDING_APPROVAL:
			params.Status = stringPtr("PENDING_APPROVAL")
		case domain.APPROVED:
			params.Status = stringPtr("APPROVED")
		case domain.POSTED:
			params.Status = stringPtr("POSTED")
		case domain.CANCELLED:
			params.Status = stringPtr("CANCELLED")
		case domain.REVERSED:
			params.Status = stringPtr("REVERSED")
		}
	}

	// Map transaction type filter
	if filter.TransactionType != nil {
		switch *filter.TransactionType {
		case domain.MANUAL:
			params.TransactionType = stringPtr("MANUAL")
		case domain.SALES_INVOICE:
			params.TransactionType = stringPtr("SALES_INVOICE")
		case domain.PURCHASE_INVOICE:
			params.TransactionType = stringPtr("PURCHASE_INVOICE")
		case domain.PAYMENT:
			params.TransactionType = stringPtr("PAYMENT")
		case domain.RECEIPT:
			params.TransactionType = stringPtr("RECEIPT")
		case domain.JOURNAL_ENTRY:
			params.TransactionType = stringPtr("JOURNAL_ENTRY")
		case domain.BANK_TRANSFER:
			params.TransactionType = stringPtr("BANK_TRANSFER")
		case domain.ADJUSTMENT:
			params.TransactionType = stringPtr("ADJUSTMENT")
		case domain.OPENING_BALANCE:
			params.TransactionType = stringPtr("OPENING_BALANCE")
		case domain.CLOSING_ENTRY:
			params.TransactionType = stringPtr("CLOSING_ENTRY")
		}
	}

	return params, nil
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

func mapDomainTransactionTypeToSQLCEnum(transactionType domain.TransactionTypeEnum) db.TransactionTypeEnum {
	switch transactionType {
	case domain.MANUAL:
		return db.TransactionTypeEnumMANUAL
	case domain.SALES_INVOICE:
		return db.TransactionTypeEnumSALESINVOICE
	case domain.PURCHASE_INVOICE:
		return db.TransactionTypeEnumPURCHASEINVOICE
	case domain.PAYMENT:
		return db.TransactionTypeEnumPAYMENT
	case domain.RECEIPT:
		return db.TransactionTypeEnumRECEIPT
	case domain.JOURNAL_ENTRY:
		return db.TransactionTypeEnumJOURNALENTRY
	case domain.BANK_TRANSFER:
		return db.TransactionTypeEnumBANKTRANSFER
	case domain.ADJUSTMENT:
		return db.TransactionTypeEnumADJUSTMENT
	case domain.OPENING_BALANCE:
		return db.TransactionTypeEnumOPENINGBALANCE
	case domain.CLOSING_ENTRY:
		return db.TransactionTypeEnumCLOSINGENTRY
	default:
		return db.TransactionTypeEnumMANUAL
	}
}

func mapDomainTransactionStatusToSQLCEnum(status domain.TransactionStatusEnum) db.TransactionStatusEnum {
	switch status {
	case domain.DRAFT:
		return db.TransactionStatusEnumDRAFT
	case domain.PENDING_APPROVAL:
		return db.TransactionStatusEnumPENDINGAPPROVAL
	case domain.APPROVED:
		return db.TransactionStatusEnumAPPROVED
	case domain.POSTED:
		return db.TransactionStatusEnumPOSTED
	case domain.CANCELLED:
		return db.TransactionStatusEnumCANCELLED
	case domain.REVERSED:
		return db.TransactionStatusEnumREVERSED
	default:
		return db.TransactionStatusEnumDRAFT
	}
}

func mapDomainTransactionStatusToSQLCEnumPtr(status domain.TransactionStatusEnum) *db.TransactionStatusEnum {
	sqlcEnum := mapDomainTransactionStatusToSQLCEnum(status)
	return &sqlcEnum
}

func mapSQLCTransactionTypeToDomain(sqlcType db.TransactionTypeEnum) domain.TransactionTypeEnum {
	switch sqlcType {
	case db.TransactionTypeEnumMANUAL:
		return domain.MANUAL
	case db.TransactionTypeEnumSALESINVOICE:
		return domain.SALES_INVOICE
	case db.TransactionTypeEnumPURCHASEINVOICE:
		return domain.PURCHASE_INVOICE
	case db.TransactionTypeEnumPAYMENT:
		return domain.PAYMENT
	case db.TransactionTypeEnumRECEIPT:
		return domain.RECEIPT
	case db.TransactionTypeEnumJOURNALENTRY:
		return domain.JOURNAL_ENTRY
	case db.TransactionTypeEnumBANKTRANSFER:
		return domain.BANK_TRANSFER
	case db.TransactionTypeEnumADJUSTMENT:
		return domain.ADJUSTMENT
	case db.TransactionTypeEnumOPENINGBALANCE:
		return domain.OPENING_BALANCE
	case db.TransactionTypeEnumCLOSINGENTRY:
		return domain.CLOSING_ENTRY
	default:
		return domain.MANUAL
	}
}

func mapSQLCTransactionStatusToDomain(sqlcStatus db.TransactionStatusEnum) domain.TransactionStatusEnum {
	switch sqlcStatus {
	case db.TransactionStatusEnumDRAFT:
		return domain.DRAFT
	case db.TransactionStatusEnumPENDINGAPPROVAL:
		return domain.PENDING_APPROVAL
	case db.TransactionStatusEnumAPPROVED:
		return domain.APPROVED
	case db.TransactionStatusEnumPOSTED:
		return domain.POSTED
	case db.TransactionStatusEnumCANCELLED:
		return domain.CANCELLED
	case db.TransactionStatusEnumREVERSED:
		return domain.REVERSED
	default:
		return domain.DRAFT
	}
}

func mapDomainApprovalStatusToString(status *domain.ApprovalStatusEnum) *string {
	if status == nil {
		return nil
	}
	statusStr := string(*status)
	return &statusStr
}

func mapDomainApprovalStatusToStringPtr(status *domain.ApprovalStatusEnum) *string {
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

func mapStringToApprovalStatus(s *string) *domain.ApprovalStatusEnum {
	if s == nil {
		return nil
	}
	status := domain.ApprovalStatusEnum(*s)
	return &status
}
