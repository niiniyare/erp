package domain

import (
	"time"

	"github.com/google/uuid"
)

// TenantUsage tracks resource consumption for a billing period.
type TenantUsage struct {
	TenantID          uuid.UUID `json:"tenant_id"`
	PeriodStart       time.Time `json:"period_start"`
	PeriodEnd         time.Time `json:"period_end"`
	ActiveUsers       int32     `json:"active_users"`
	TotalEntities     int32     `json:"total_entities"`
	TotalTransactions int32     `json:"total_transactions"`
	StorageUsed       int64     `json:"storage_used"`
	APICalls          int32     `json:"api_calls"`
}

// IsWithinLimits checks whether usage is within the given configuration limits.
func (u *TenantUsage) IsWithinLimits(cfg *TenantConfiguration) bool {
	if cfg == nil {
		return true
	}
	return u.ActiveUsers <= cfg.MaxUsers &&
		u.TotalEntities <= cfg.MaxEntities &&
		u.TotalTransactions <= cfg.MaxTransactionsPerMonth &&
		u.StorageUsed <= cfg.StorageQuota
}

// UsagePercentage returns the highest resource utilisation ratio (0-1).
func (u *TenantUsage) UsagePercentage(cfg *TenantConfiguration) float64 {
	if cfg == nil {
		return 0
	}
	max := func(a, b float64) float64 {
		if a > b {
			return a
		}
		return b
	}
	pct := 0.0
	if cfg.MaxUsers > 0 {
		pct = max(pct, float64(u.ActiveUsers)/float64(cfg.MaxUsers))
	}
	if cfg.MaxEntities > 0 {
		pct = max(pct, float64(u.TotalEntities)/float64(cfg.MaxEntities))
	}
	if cfg.MaxTransactionsPerMonth > 0 {
		pct = max(pct, float64(u.TotalTransactions)/float64(cfg.MaxTransactionsPerMonth))
	}
	if cfg.StorageQuota > 0 {
		pct = max(pct, float64(u.StorageUsed)/float64(cfg.StorageQuota))
	}
	return pct
}
