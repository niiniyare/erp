package domain

import "github.com/google/uuid"

// PlanType represents subscription plans.
type PlanType string

const (
	PlanBasic        PlanType = "BASIC"
	PlanProfessional PlanType = "PROFESSIONAL"
	PlanEnterprise   PlanType = "ENTERPRISE"
)

// CompanySize represents the size category of a tenant organization.
type CompanySize string

const (
	CompanySizeStartup    CompanySize = "STARTUP"
	CompanySizeSmall      CompanySize = "SMALL"
	CompanySizeMedium     CompanySize = "MEDIUM"
	CompanySizeLarge      CompanySize = "LARGE"
	CompanySizeEnterprise CompanySize = "ENTERPRISE"
)

// ValidCompanySize returns true if the given size is a valid CompanySize value.
func ValidCompanySize(s string) bool {
	switch CompanySize(s) {
	case CompanySizeStartup, CompanySizeSmall, CompanySizeMedium, CompanySizeLarge, CompanySizeEnterprise:
		return true
	}
	return false
}

// AccountingMethod represents the accounting method for a tenant.
type AccountingMethod string

const (
	AccountingMethodAccrual AccountingMethod = "ACCRUAL"
	AccountingMethodCash    AccountingMethod = "CASH"
)

// ValidAccountingMethod returns true if the given method is valid.
func ValidAccountingMethod(s string) bool {
	switch AccountingMethod(s) {
	case AccountingMethodAccrual, AccountingMethodCash:
		return true
	}
	return false
}

// TenantLimits represents tenant usage limits.
type TenantLimits struct {
	MaxUsers           uint   `json:"max_users"`
	MaxStorageMB       uint64 `json:"max_storage_mb"`
	MaxAPICallsPerHour uint   `json:"max_api_calls_per_hour"`
}

// CreateTenantRequest represents tenant creation input.
type CreateTenantRequest struct {
	Name               string         `json:"name" validate:"required,min=2,max=255"`
	Slug               string         `json:"slug,omitempty"`
	Email              string         `json:"email" validate:"required,email"`
	Subdomain          *string        `json:"subdomain,omitempty"`
	Status             TenantStatus   `json:"status,omitempty"`
	Industry           *string        `json:"industry,omitempty"`
	CompanySize        *string        `json:"company_size,omitempty"`
	TaxID              *string        `json:"tax_id,omitempty"`
	RegistrationNumber *string        `json:"registration_number,omitempty"`
	LegalEntityType    *string        `json:"legal_entity_type,omitempty"`
	Settings           map[string]any `json:"settings,omitempty"`
	CountryCode        string         `json:"country_code" validate:"iso3166_1_alpha2"`
	CurrencyCode       string         `json:"currency_code" validate:"iso4217"`
}

// UpdateTenantRequest represents tenant update input.
type UpdateTenantRequest struct {
	Name               *string        `json:"name,omitempty"`
	Email              *string        `json:"email,omitempty"`
	Subdomain          *string        `json:"subdomain,omitempty"`
	Status             *TenantStatus  `json:"status,omitempty"`
	Industry           *string        `json:"industry,omitempty"`
	CompanySize        *string        `json:"company_size,omitempty"`
	TaxID              *string        `json:"tax_id,omitempty"`
	RegistrationNumber *string        `json:"registration_number,omitempty"`
	LegalEntityType    *string        `json:"legal_entity_type,omitempty"`
	Settings           map[string]any `json:"settings,omitempty"`
}

// TenantFilter is used for listing tenants with advanced filtering.
type TenantFilter struct {
	NameFilter     *string `json:"name_filter,omitempty"`
	StatusFilter   *string `json:"status_filter,omitempty"`
	IndustryFilter *string `json:"industry_filter,omitempty"`
	SortBy         string  `json:"sort_by,omitempty"`
	Offset         int32   `json:"offset"`
	Limit          int32   `json:"limit"`
}

// ProvisioningInput is what the provisioning workflow receives.
type ProvisioningInput struct {
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	Subdomain    *string `json:"subdomain,omitempty"`
	Industry     *string `json:"industry,omitempty"`
	CompanySize  *string `json:"company_size,omitempty"`
	CountryCode  string  `json:"country_code"`
	CurrencyCode string  `json:"currency_code"`
}

// ProvisioningResult is the output of provisioning.
type ProvisioningResult struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	Slug      string    `json:"slug"`
	Subdomain *string   `json:"subdomain,omitempty"`
	Status    string    `json:"status"`
}

// GrowthStat represents tenant growth data for a period.
type GrowthStat struct {
	Month            string `json:"month"`
	NewTenants       int64  `json:"new_tenants"`
	ActiveNewTenants int64  `json:"active_new_tenants"`
}

// StatusCount represents the distribution of tenant statuses.
type StatusCount struct {
	TotalTenants     int64   `json:"total_tenants"`
	ActiveTenants    int64   `json:"active_tenants"`
	SuspendedTenants int64   `json:"suspended_tenants"`
	PendingTenants   int64   `json:"pending_tenants"`
	ActivePct        float64 `json:"active_pct"`
	SuspendedPct     float64 `json:"suspended_pct"`
	PendingPct       float64 `json:"pending_pct"`
}

// ReservedSubdomains is the list of system-reserved subdomain names.
var ReservedSubdomains = map[string]bool{
	"admin": true, "api": true, "www": true, "app": true,
	"cdn": true, "static": true, "assets": true, "mail": true,
	"ftp": true, "smtp": true, "dev": true, "staging": true,
	"prod": true, "production": true, "test": true, "localhost": true,
	"dashboard": true, "portal": true, "auth": true, "login": true,
	"signup": true, "register": true, "billing": true, "payment": true,
	"invoice": true, "support": true, "help": true, "docs": true,
	"status": true, "blog": true, "news": true, "about": true,
	"contact": true, "legal": true, "privacy": true, "terms": true,
}
