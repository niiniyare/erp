package domain

import (
	"time"

	"github.com/google/uuid"
)

// TenantConfiguration holds tenant-specific settings and operational limits.
type TenantConfiguration struct {
	TenantID                uuid.UUID      `json:"tenant_id"`
	MaxUsers                int32          `json:"max_users"`
	MaxEntities             int32          `json:"max_entities"`
	MaxTransactionsPerMonth int32          `json:"max_transactions_per_month"`
	StorageQuota            int64          `json:"storage_quota"`
	AccountingMethod        string         `json:"accounting_method"`
	FiscalYearStartMonth    int32          `json:"fiscal_year_start_month"`
	DefaultCurrency         string         `json:"default_currency"`
	DateFormat              string         `json:"date_format"`
	NumberFormat            string         `json:"number_format"`
	LanguageCode            string         `json:"language_code"`
	PasswordPolicy          map[string]any `json:"password_policy,omitempty"`
	Settings                map[string]any `json:"settings,omitempty"`
	WebhookEndpoints        []any          `json:"webhook_endpoints,omitempty"`
	ApiRateLimits           map[string]any `json:"api_rate_limits,omitempty"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
}

// NewDefaultConfig creates a configuration with sensible defaults.
func NewDefaultConfig(tenantID uuid.UUID) *TenantConfiguration {
	return &TenantConfiguration{
		TenantID:                tenantID,
		MaxUsers:                100,
		MaxEntities:             1000,
		MaxTransactionsPerMonth: 10000,
		StorageQuota:            10 * 1024 * 1024 * 1024, // 10GB
		AccountingMethod:        string(AccountingMethodAccrual),
		FiscalYearStartMonth:    1,
		DefaultCurrency:         "USD",
		DateFormat:              "YYYY-MM-DD",
		NumberFormat:            "1,234.56",
		LanguageCode:            "en",
	}
}
