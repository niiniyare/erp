package repository

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/tenant/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ---------------------------------------------------------------------------
// MapperSuite — tests for toDomain, filterRowToDomain, configToDomain, usageToDomain
// ---------------------------------------------------------------------------

type MapperSuite struct {
	suite.Suite
}

func TestMapperSuite(t *testing.T) { suite.Run(t, new(MapperSuite)) }

// ---- toDomain -------------------------------------------------------------

func (s *MapperSuite) TestToDomain_FullRow() {
	id := uuid.New()
	sub := "acme"
	industry := "tech"
	size := "SMALL"
	taxID := "TAX123"
	regNum := "REG456"
	legalType := "LLC"
	now := time.Now()
	deletedAt := now.Add(-time.Hour)
	lastActivity := now.Add(-30 * time.Minute)

	meta, _ := json.Marshal(map[string]any{"key": "value"})
	settings, _ := json.Marshal(map[string]any{"theme": "dark"})

	row := &db.Tenant{
		ID:                 id,
		Slug:               "acme-corp",
		Name:               "Acme Corp",
		Email:              "admin@acme.com",
		Subdomain:          &sub,
		Status:             "ACTIVE",
		Timezone:           "America/New_York",
		CurrencyCode:       "USD",
		Metadata:           meta,
		Industry:           &industry,
		CompanySize:        &size,
		TaxID:              &taxID,
		RegistrationNumber: &regNum,
		LegalEntityType:    &legalType,
		LastActivityAt:     lastActivity,
		Settings:           settings,
		CreatedAt:          now,
		UpdatedAt:          now,
		DeletedAt:          sql.NullTime{Time: deletedAt, Valid: true},
	}

	tenant, err := toDomain(row)

	require.NoError(s.T(), err)
	require.Equal(s.T(), id, tenant.ID)
	require.Equal(s.T(), "acme-corp", tenant.Slug)
	require.Equal(s.T(), "Acme Corp", tenant.Name)
	require.Equal(s.T(), "admin@acme.com", tenant.Email)
	require.Equal(s.T(), &sub, tenant.Subdomain)
	require.Equal(s.T(), domain.StatusActive, tenant.Status)
	require.Equal(s.T(), "America/New_York", tenant.Timezone)
	require.Equal(s.T(), "USD", tenant.CurrencyCode)
	require.Equal(s.T(), "value", tenant.Metadata["key"])
	require.Equal(s.T(), "dark", tenant.Settings["theme"])
	require.Equal(s.T(), &industry, tenant.Industry)
	require.Equal(s.T(), &size, tenant.CompanySize)
	require.Equal(s.T(), &taxID, tenant.TaxID)
	require.Equal(s.T(), &regNum, tenant.RegistrationNumber)
	require.Equal(s.T(), &legalType, tenant.LegalEntityType)
	require.NotNil(s.T(), tenant.LastActivityAt)
	require.WithinDuration(s.T(), lastActivity, *tenant.LastActivityAt, time.Millisecond)
	require.NotNil(s.T(), tenant.DeletedAt)
	require.WithinDuration(s.T(), deletedAt, *tenant.DeletedAt, time.Millisecond)
	require.Equal(s.T(), now, tenant.CreatedAt)
	require.Equal(s.T(), now, tenant.UpdatedAt)
}

func (s *MapperSuite) TestToDomain_NullableFieldsNil() {
	row := &db.Tenant{
		ID:             uuid.New(),
		Slug:           "minimal",
		Name:           "Minimal",
		Email:          "min@test.com",
		Status:         "PENDING",
		Timezone:       "UTC",
		CurrencyCode:   "USD",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		DeletedAt:      sql.NullTime{Valid: false},
		LastActivityAt: time.Time{},
	}

	tenant, err := toDomain(row)

	require.NoError(s.T(), err)
	require.Nil(s.T(), tenant.Subdomain)
	require.Nil(s.T(), tenant.Industry)
	require.Nil(s.T(), tenant.CompanySize)
	require.Nil(s.T(), tenant.TaxID)
	require.Nil(s.T(), tenant.DeletedAt)
	require.Nil(s.T(), tenant.LastActivityAt)
}

func (s *MapperSuite) TestToDomain_EmptyJSON() {
	row := &db.Tenant{
		ID:           uuid.New(),
		Slug:         "empty-json",
		Name:         "Empty JSON",
		Email:        "ej@test.com",
		Status:       "ACTIVE",
		Timezone:     "UTC",
		CurrencyCode: "USD",
		Metadata:     nil,
		Settings:     nil,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	tenant, err := toDomain(row)

	require.NoError(s.T(), err)
	require.Nil(s.T(), tenant.Metadata)
	require.Nil(s.T(), tenant.Settings)
}

func (s *MapperSuite) TestToDomain_InvalidJSON() {
	row := &db.Tenant{
		ID:           uuid.New(),
		Slug:         "bad-json",
		Name:         "Bad JSON",
		Email:        "bj@test.com",
		Status:       "ACTIVE",
		Timezone:     "UTC",
		CurrencyCode: "USD",
		Metadata:     []byte("{invalid json"),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	tenant, err := toDomain(row)

	require.Error(s.T(), err)
	require.Nil(s.T(), tenant)
}

func (s *MapperSuite) TestToDomain_StatusMapping() {
	tests := []struct {
		dbStatus string
		want     domain.TenantStatus
	}{
		{"ACTIVE", domain.StatusActive},
		{"PENDING", domain.StatusPending},
		{"SUSPENDED", domain.StatusSuspended},
		{"ARCHIVED", domain.StatusArchived},
	}

	for _, tc := range tests {
		s.Run(tc.dbStatus, func() {
			row := &db.Tenant{
				ID:           uuid.New(),
				Slug:         "status-test",
				Name:         "Status Test",
				Email:        "st@test.com",
				Status:       tc.dbStatus,
				Timezone:     "UTC",
				CurrencyCode: "USD",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			tenant, err := toDomain(row)
			require.NoError(s.T(), err)
			require.Equal(s.T(), tc.want, tenant.Status)
		})
	}
}

// ---- filterRowToDomain ----------------------------------------------------

func (s *MapperSuite) TestFilterRowToDomain() {
	id := uuid.New()
	sub := "filtered"
	industry := "finance"
	now := time.Now()
	deleted := now.Add(-time.Hour)

	row := &db.FilterTenantsRow{
		ID:        id,
		Name:      "Filtered Tenant",
		Subdomain: &sub,
		Status:    "ACTIVE",
		Industry:  &industry,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: sql.NullTime{Time: deleted, Valid: true},
	}

	tenant := filterRowToDomain(row)

	require.Equal(s.T(), id, tenant.ID)
	require.Equal(s.T(), "Filtered Tenant", tenant.Name)
	require.Equal(s.T(), &sub, tenant.Subdomain)
	require.Equal(s.T(), domain.StatusActive, tenant.Status)
	require.Equal(s.T(), &industry, tenant.Industry)
	require.NotNil(s.T(), tenant.DeletedAt)
}

func (s *MapperSuite) TestFilterRowToDomain_NullDeletedAt() {
	row := &db.FilterTenantsRow{
		ID:        uuid.New(),
		Name:      "Not Deleted",
		Status:    "PENDING",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: sql.NullTime{Valid: false},
	}

	tenant := filterRowToDomain(row)
	require.Nil(s.T(), tenant.DeletedAt)
}

// ---- configToDomain -------------------------------------------------------

func (s *MapperSuite) TestConfigToDomain() {
	tenantID := uuid.New()
	row := &db.TenantConfiguration{
		TenantID:                tenantID,
		MaxUsers:                100,
		MaxEntities:             1000,
		MaxTransactionsPerMonth: 10000,
		StorageQuota:            10 * 1024 * 1024 * 1024,
		AccountingMethod:        "ACCRUAL",
		FiscalYearStartMonth:    1,
		DefaultCurrency:         "USD",
		DateFormat:              "YYYY-MM-DD",
		NumberFormat:            "1,234.56",
		LanguageCode:            "en",
	}

	cfg := configToDomain(row)

	require.Equal(s.T(), tenantID, cfg.TenantID)
	require.Equal(s.T(), int32(100), cfg.MaxUsers)
	require.Equal(s.T(), int32(1000), cfg.MaxEntities)
	require.Equal(s.T(), int32(10000), cfg.MaxTransactionsPerMonth)
	require.Equal(s.T(), int64(10*1024*1024*1024), cfg.StorageQuota)
	require.Equal(s.T(), "ACCRUAL", cfg.AccountingMethod)
	require.Equal(s.T(), int32(1), cfg.FiscalYearStartMonth)
	require.Equal(s.T(), "USD", cfg.DefaultCurrency)
	require.Equal(s.T(), "YYYY-MM-DD", cfg.DateFormat)
	require.Equal(s.T(), "1,234.56", cfg.NumberFormat)
	require.Equal(s.T(), "en", cfg.LanguageCode)
}

// ---- usageToDomain --------------------------------------------------------

func (s *MapperSuite) TestUsageToDomain() {
	tenantID := uuid.New()
	start := time.Now().Add(-30 * 24 * time.Hour)
	end := time.Now()

	row := &db.TenantUsageStat{
		TenantID:          tenantID,
		PeriodStart:       start,
		PeriodEnd:         end,
		ActiveUsers:       50,
		TotalEntities:     500,
		TotalTransactions: 5000,
		StorageUsed:       5 * 1024 * 1024 * 1024,
		ApiCalls:          10000,
	}

	usage := usageToDomain(row)

	require.Equal(s.T(), tenantID, usage.TenantID)
	require.Equal(s.T(), start, usage.PeriodStart)
	require.Equal(s.T(), end, usage.PeriodEnd)
	require.Equal(s.T(), int32(50), usage.ActiveUsers)
	require.Equal(s.T(), int32(500), usage.TotalEntities)
	require.Equal(s.T(), int32(5000), usage.TotalTransactions)
	require.Equal(s.T(), int64(5*1024*1024*1024), usage.StorageUsed)
	require.Equal(s.T(), int32(10000), usage.APICalls)
}
