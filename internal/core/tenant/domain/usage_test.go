package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"awo.so/internal/core/tenant/domain"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UsageSuite struct {
	suite.Suite
}

func TestUsageSuite(t *testing.T) { suite.Run(t, new(UsageSuite)) }

// ---- IsWithinLimits -------------------------------------------------------

func (s *UsageSuite) TestIsWithinLimits() {
	cfg := domain.NewDefaultConfig(uuid.New()) // 100 users, 1000 entities, 10000 txns, 10GB

	tests := []struct {
		name   string
		usage  domain.TenantUsage
		config *domain.TenantConfiguration
		want   bool
	}{
		{
			name:   "all within limits",
			usage:  domain.TenantUsage{ActiveUsers: 50, TotalEntities: 500, TotalTransactions: 5000, StorageUsed: 5 * 1024 * 1024 * 1024},
			config: cfg,
			want:   true,
		},
		{
			name:   "at exact limits",
			usage:  domain.TenantUsage{ActiveUsers: 100, TotalEntities: 1000, TotalTransactions: 10000, StorageUsed: 10 * 1024 * 1024 * 1024},
			config: cfg,
			want:   true,
		},
		{
			name:   "users exceeded",
			usage:  domain.TenantUsage{ActiveUsers: 101, TotalEntities: 500, TotalTransactions: 5000, StorageUsed: 0},
			config: cfg,
			want:   false,
		},
		{
			name:   "entities exceeded",
			usage:  domain.TenantUsage{ActiveUsers: 50, TotalEntities: 1001, TotalTransactions: 5000, StorageUsed: 0},
			config: cfg,
			want:   false,
		},
		{
			name:   "transactions exceeded",
			usage:  domain.TenantUsage{ActiveUsers: 50, TotalEntities: 500, TotalTransactions: 10001, StorageUsed: 0},
			config: cfg,
			want:   false,
		},
		{
			name:   "storage exceeded",
			usage:  domain.TenantUsage{ActiveUsers: 50, TotalEntities: 500, TotalTransactions: 5000, StorageUsed: 11 * 1024 * 1024 * 1024},
			config: cfg,
			want:   false,
		},
		{
			name:   "nil config returns true",
			usage:  domain.TenantUsage{ActiveUsers: 999},
			config: nil,
			want:   true,
		},
		{
			name:   "zero usage always within",
			usage:  domain.TenantUsage{},
			config: cfg,
			want:   true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			require.Equal(s.T(), tc.want, tc.usage.IsWithinLimits(tc.config))
		})
	}
}

// ---- UsagePercentage ------------------------------------------------------

func (s *UsageSuite) TestUsagePercentage() {
	cfg := &domain.TenantConfiguration{
		MaxUsers:                100,
		MaxEntities:             1000,
		MaxTransactionsPerMonth: 10000,
		StorageQuota:            10 * 1024 * 1024 * 1024,
	}

	tests := []struct {
		name    string
		usage   domain.TenantUsage
		config  *domain.TenantConfiguration
		wantPct float64
		delta   float64
	}{
		{
			name:    "50% users is highest",
			usage:   domain.TenantUsage{ActiveUsers: 50, TotalEntities: 100, TotalTransactions: 1000, StorageUsed: 1024 * 1024 * 1024},
			config:  cfg,
			wantPct: 0.5,
			delta:   0.001,
		},
		{
			name:    "80% entities is highest",
			usage:   domain.TenantUsage{ActiveUsers: 10, TotalEntities: 800, TotalTransactions: 100, StorageUsed: 0},
			config:  cfg,
			wantPct: 0.8,
			delta:   0.001,
		},
		{
			name:    "100% at limit",
			usage:   domain.TenantUsage{ActiveUsers: 100, TotalEntities: 1000, TotalTransactions: 10000, StorageUsed: 10 * 1024 * 1024 * 1024},
			config:  cfg,
			wantPct: 1.0,
			delta:   0.001,
		},
		{
			name:    "zero usage",
			usage:   domain.TenantUsage{},
			config:  cfg,
			wantPct: 0.0,
			delta:   0.001,
		},
		{
			name:    "nil config returns 0",
			usage:   domain.TenantUsage{ActiveUsers: 999},
			config:  nil,
			wantPct: 0.0,
			delta:   0.001,
		},
		{
			name:    "over limit exceeds 1.0",
			usage:   domain.TenantUsage{ActiveUsers: 200},
			config:  cfg,
			wantPct: 2.0,
			delta:   0.001,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got := tc.usage.UsagePercentage(tc.config)
			require.InDelta(s.T(), tc.wantPct, got, tc.delta)
		})
	}
}
