package repository

import (
	"encoding/json"
	"time"

	db "awo/db/sqlc"
	"awo/internal/core/tenant/domain"
)

// toDomain converts a SQLC Tenant row to the domain entity.
func toDomain(row *db.Tenant) (*domain.Tenant, error) {
	var metadata map[string]any
	if len(row.Metadata) > 0 {
		if err := json.Unmarshal(row.Metadata, &metadata); err != nil {
			return nil, err
		}
	}

	var settings map[string]any
	if len(row.Settings) > 0 {
		if err := json.Unmarshal(row.Settings, &settings); err != nil {
			return nil, err
		}
	}

	var deletedAt *time.Time
	if row.DeletedAt.Valid {
		deletedAt = &row.DeletedAt.Time
	}

	var lastActivity *time.Time
	if !row.LastActivityAt.IsZero() {
		lastActivity = &row.LastActivityAt
	}

	return &domain.Tenant{
		ID:                 row.ID,
		Slug:               row.Slug,
		Name:               row.Name,
		Email:              row.Email,
		Subdomain:          row.Subdomain,
		Status:             domain.TenantStatus(row.Status),
		PlanTier:           row.PlanTier,
		Timezone:           row.Timezone,
		CurrencyCode:       row.CurrencyCode,
		Metadata:           metadata,
		Industry:           row.Industry,
		CompanySize:        row.CompanySize,
		TaxID:              row.TaxID,
		RegistrationNumber: row.RegistrationNumber,
		LegalEntityType:    row.LegalEntityType,
		Settings:           settings,
		LastActivityAt:     lastActivity,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
		DeletedAt:          deletedAt,
	}, nil
}

// filterRowToDomain converts a FilterTenantsRow to domain entity (partial fields).
func filterRowToDomain(row *db.FilterTenantsRow) *domain.Tenant {
	var deletedAt *time.Time
	if row.DeletedAt.Valid {
		deletedAt = &row.DeletedAt.Time
	}

	return &domain.Tenant{
		ID:           row.ID,
		Slug:         row.Slug,
		Name:         row.Name,
		Email:        row.Email,
		Subdomain:    row.Subdomain,
		Status:       domain.TenantStatus(row.Status),
		Timezone:     row.Timezone,
		CurrencyCode: row.CurrencyCode,
		Industry:     row.Industry,
		CompanySize:  row.CompanySize,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		DeletedAt:    deletedAt,
	}
}

// configToDomain converts a SQLC TenantConfiguration to domain.
func configToDomain(row *db.TenantConfiguration) *domain.TenantConfiguration {
	var passwordPolicy map[string]any
	if len(row.PasswordPolicy) > 0 {
		_ = json.Unmarshal(row.PasswordPolicy, &passwordPolicy)
	}

	var settings map[string]any
	if len(row.Settings) > 0 {
		_ = json.Unmarshal(row.Settings, &settings)
	}

	var webhookEndpoints []any
	if len(row.WebhookEndpoints) > 0 {
		_ = json.Unmarshal(row.WebhookEndpoints, &webhookEndpoints)
	}

	var apiRateLimits map[string]any
	if len(row.ApiRateLimits) > 0 {
		_ = json.Unmarshal(row.ApiRateLimits, &apiRateLimits)
	}

	return &domain.TenantConfiguration{
		TenantID:                row.TenantID,
		MaxUsers:                row.MaxUsers,
		MaxEntities:             row.MaxEntities,
		MaxTransactionsPerMonth: row.MaxTransactionsPerMonth,
		StorageQuota:            row.StorageQuota,
		AccountingMethod:        row.AccountingMethod,
		FiscalYearStartMonth:    row.FiscalYearStartMonth,
		DefaultCurrency:         row.DefaultCurrency,
		DateFormat:              row.DateFormat,
		NumberFormat:            row.NumberFormat,
		LanguageCode:            row.LanguageCode,
		PasswordPolicy:          passwordPolicy,
		Settings:                settings,
		WebhookEndpoints:        webhookEndpoints,
		ApiRateLimits:           apiRateLimits,
		CreatedAt:               row.CreatedAt,
		UpdatedAt:               row.UpdatedAt,
	}
}

// usageToDomain converts a SQLC TenantUsageStat to domain.
func usageToDomain(row *db.TenantUsageStat) *domain.TenantUsage {
	return &domain.TenantUsage{
		TenantID:          row.TenantID,
		PeriodStart:       row.PeriodStart,
		PeriodEnd:         row.PeriodEnd,
		ActiveUsers:       row.ActiveUsers,
		TotalEntities:     row.TotalEntities,
		TotalTransactions: row.TotalTransactions,
		StorageUsed:       row.StorageUsed,
		APICalls:          row.ApiCalls,
	}
}
