package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/tenant/domain"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ConfigSuite struct {
	suite.Suite
}

func TestConfigSuite(t *testing.T) { suite.Run(t, new(ConfigSuite)) }

func (s *ConfigSuite) TestNewDefaultConfig() {
	id := uuid.New()
	cfg := domain.NewDefaultConfig(id)

	require.NotNil(s.T(), cfg)
	require.Equal(s.T(), id, cfg.TenantID)

	// Resource limits
	require.Equal(s.T(), int32(100), cfg.MaxUsers)
	require.Equal(s.T(), int32(1000), cfg.MaxEntities)
	require.Equal(s.T(), int32(10000), cfg.MaxTransactionsPerMonth)
	require.Equal(s.T(), int64(10*1024*1024*1024), cfg.StorageQuota) // 10GB

	// Accounting — must be uppercase ACCRUAL
	require.Equal(s.T(), "ACCRUAL", cfg.AccountingMethod)

	// Localisation
	require.Equal(s.T(), int32(1), cfg.FiscalYearStartMonth)
	require.Equal(s.T(), "USD", cfg.DefaultCurrency)
	require.Equal(s.T(), "YYYY-MM-DD", cfg.DateFormat)
	require.Equal(s.T(), "1,234.56", cfg.NumberFormat)
	require.Equal(s.T(), "en", cfg.LanguageCode)
}

func (s *ConfigSuite) TestNewDefaultConfig_DifferentIDs() {
	id1 := uuid.New()
	id2 := uuid.New()

	cfg1 := domain.NewDefaultConfig(id1)
	cfg2 := domain.NewDefaultConfig(id2)

	require.Equal(s.T(), id1, cfg1.TenantID)
	require.Equal(s.T(), id2, cfg2.TenantID)
	require.NotEqual(s.T(), cfg1.TenantID, cfg2.TenantID)
}
