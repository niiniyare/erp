package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// AccountHierarchy represents the hierarchical view of accounts
type AccountHierarchy struct {
	ID               uuid.UUID       `json:"id"`
	TenantID         uuid.UUID       `json:"tenant_id"`
	AccountCode      string          `json:"account_code"`
	AccountName      string          `json:"account_name"`
	ParentAccountID  *uuid.UUID      `json:"parent_account_id,omitempty"`
	RootType         string          `json:"root_type"`
	AccountType      string          `json:"account_type"`
	NormalBalance    string          `json:"normal_balance"`
	CurrentBalance   decimal.Decimal `json:"current_balance"`
	Level            int32           `json:"level"`
	FullPath         string          `json:"full_path"`
	FullName         string          `json:"full_name"`
	ChildCount       int64           `json:"child_count"`
}

// AccountActivity represents account activity metrics from the database view
type AccountActivity struct {
	TenantID             uuid.UUID       `json:"tenant_id"`
	AccountID            uuid.UUID       `json:"account_id"`
	AccountCode          string          `json:"account_code"`
	AccountName          string          `json:"account_name"`
	CurrentBalance       decimal.Decimal `json:"current_balance"`
	LastTransactionDate  *time.Time      `json:"last_transaction_date,omitempty"`
	TotalEntries         int64           `json:"total_entries"`
	EntriesLast30Days    int64           `json:"entries_last_30_days"`
	DebitsLast30Days     decimal.Decimal `json:"debits_last_30_days"`
	CreditsLast30Days    decimal.Decimal `json:"credits_last_30_days"`
}

// AccountActivitySummary represents enhanced activity summary with categorization
type AccountActivitySummary struct {
	AccountID           uuid.UUID       `json:"account_id"`
	AccountCode         string          `json:"account_code"`
	AccountName         string          `json:"account_name"`
	CurrentBalance      decimal.Decimal `json:"current_balance"`
	EntriesLast30Days   int64           `json:"entries_last_30_days"`
	DebitsLast30Days    decimal.Decimal `json:"debits_last_30_days"`
	CreditsLast30Days   decimal.Decimal `json:"credits_last_30_days"`
	TotalActivity30Days decimal.Decimal `json:"total_activity_30_days"`
	ActivityLevel       string          `json:"activity_level"`
}

// AccountActivityFilter for filtering account activity queries
type AccountActivityFilter struct {
	EntityID         *uuid.UUID `json:"entity_id,omitempty"`
	MinEntries       *int32     `json:"min_entries,omitempty"`
	MinDaysInactive  *int32     `json:"min_days_inactive,omitempty"`
	MinBalance       *decimal.Decimal `json:"min_balance,omitempty"`
	ActivityLevel    *string    `json:"activity_level,omitempty"`
}