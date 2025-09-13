package repository

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/finance/domain"
)

// mapSQLCAccountHierarchyToDomain converts SQLC account hierarchy to domain type
func mapSQLCAccountHierarchyToDomain(sqlcAccount db.VFinanceAccountsHierarchy) *domain.AccountHierarchy {
	return &domain.AccountHierarchy{
		ID:               sqlcAccount.ID,
		TenantID:         sqlcAccount.TenantID,
		AccountCode:      sqlcAccount.AccountCode,
		AccountName:      sqlcAccount.AccountName,
		ParentAccountID:  sqlcAccount.ParentAccountID,
		RootType:         sqlcAccount.RootType,
		AccountType:      sqlcAccount.AccountType,
		NormalBalance:    sqlcAccount.NormalBalance,
		CurrentBalance:   convertNullableDecimal(sqlcAccount.CurrentBalance),
		Level:            sqlcAccount.Level,
		FullPath:         sqlcAccount.FullPath,
		FullName:         sqlcAccount.FullName,
		ChildCount:       sqlcAccount.ChildCount,
	}
}

// mapSQLCAccountActivityToDomain converts SQLC account activity to domain type
func mapSQLCAccountActivityToDomain(sqlcActivity db.VFinanceAccountActivity) *domain.AccountActivity {
	// The LastTransactionDate is a regular time.Time, so we need to check if it's zero
	var lastTransactionDate *time.Time
	if !sqlcActivity.LastTransactionDate.IsZero() {
		lastTransactionDate = &sqlcActivity.LastTransactionDate
	}

	return &domain.AccountActivity{
		TenantID:            sqlcActivity.TenantID,
		AccountID:           sqlcActivity.AccountID,
		AccountCode:         sqlcActivity.AccountCode,
		AccountName:         sqlcActivity.AccountName,
		CurrentBalance:      convertNullableDecimal(sqlcActivity.CurrentBalance),
		LastTransactionDate: lastTransactionDate,
		TotalEntries:        sqlcActivity.TotalEntries,
		EntriesLast30Days:   sqlcActivity.EntriesLast30Days,
		DebitsLast30Days:    decimal.NewFromInt(sqlcActivity.DebitsLast30Days),
		CreditsLast30Days:   decimal.NewFromInt(sqlcActivity.CreditsLast30Days),
	}
}

// mapSQLCAccountActivitySummaryToDomain converts SQLC account activity summary to domain type
func mapSQLCAccountActivitySummaryToDomain(sqlcSummary db.GetAccountActivitySummaryRow) *domain.AccountActivitySummary {
	return &domain.AccountActivitySummary{
		AccountID:           sqlcSummary.AccountID,
		AccountCode:         sqlcSummary.AccountCode,
		AccountName:         sqlcSummary.AccountName,
		CurrentBalance:      convertNullableDecimal(sqlcSummary.CurrentBalance),
		EntriesLast30Days:   sqlcSummary.EntriesLast30Days,
		DebitsLast30Days:    decimal.NewFromInt(sqlcSummary.DebitsLast30Days),
		CreditsLast30Days:   decimal.NewFromInt(sqlcSummary.CreditsLast30Days),
		TotalActivity30Days: decimal.NewFromInt32(sqlcSummary.TotalActivity30Days),
		ActivityLevel:       sqlcSummary.ActivityLevel,
	}
}

// mapSQLCStaleAccountToDomain converts SQLC stale account row to domain activity type
func mapSQLCStaleAccountToDomain(sqlcStale db.GetStaleAccountBalancesRow) *domain.AccountActivity {
	// Handle time fields
	var lastTransactionDate *time.Time
	if !sqlcStale.LastTransactionDate.IsZero() {
		lastTransactionDate = &sqlcStale.LastTransactionDate
	}

	return &domain.AccountActivity{
		AccountID:           sqlcStale.AccountID,
		AccountCode:         sqlcStale.AccountCode,
		AccountName:         sqlcStale.AccountName,
		CurrentBalance:      convertNullableDecimal(sqlcStale.CurrentBalance),
		LastTransactionDate: lastTransactionDate,
		TotalEntries:        sqlcStale.TotalEntries,
		// Set default values for fields not in the stale accounts query
		EntriesLast30Days: 0,
		DebitsLast30Days:  decimal.Zero,
		CreditsLast30Days: decimal.Zero,
	}
}

// Helper function to convert nullable decimal from pgtype.Numeric to decimal.Decimal
func convertNullableDecimal(pgNum pgtype.Numeric) decimal.Decimal {
	if !pgNum.Valid {
		return decimal.Zero
	}
	
	// Convert pgtype.Numeric to decimal.Decimal
	// Use the Float64Value method to get the value
	val, err := pgNum.Float64Value()
	if err != nil {
		return decimal.Zero
	}
	
	return decimal.NewFromFloat(val.Float64)
}