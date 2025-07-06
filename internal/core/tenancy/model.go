package tenancy

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Tenant represents a tenant in the multi-tenant ERP system
// This maps to the `tenants` table in the database schema
type Tenant struct {
	ID                 uuid.UUID    `json:"id" db:"id"`
	Slug               string       `json:"slug" db:"slug"`
	Name               string       `json:"name" db:"name"`
	Email              string       `json:"email" db:"email"`
	Subdomain          *string      `json:"subdomain,omitempty" db:"subdomain"`
	Status             string       `json:"status" db:"status"`
	Timezone           string       `json:"timezone" db:"timezone"`
	CurrencyCode       string       `json:"currency_code" db:"currency_code"`
	Metadata           []byte       `json:"metadata" db:"metadata"`
	Industry           *string      `json:"industry,omitempty" db:"industry"`
	CompanySize        *string      `json:"company_size,omitempty" db:"company_size"`
	TaxID              *string      `json:"tax_id,omitempty" db:"tax_id"`
	RegistrationNumber *string      `json:"registration_number,omitempty" db:"registration_number"`
	LegalEntityType    *string      `json:"legal_entity_type,omitempty" db:"legal_entity_type"`
	Settings           []byte       `json:"settings" db:"settings"`
	CreatedAt          time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt          sql.NullTime `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Configuration represents tenant-specific configuration settings
// This maps to the `tenant_configurations` table in the database schema
type Configuration struct {
	TenantID                uuid.UUID `json:"tenant_id" db:"tenant_id"`
	MaxUsers                int32     `json:"max_users" db:"max_users"`
	MaxEntities             int32     `json:"max_entities" db:"max_entities"`
	MaxTransactionsPerMonth int32     `json:"max_transactions_per_month" db:"max_transactions_per_month"`
	StorageQuota            int64     `json:"storage_quota" db:"storage_quota"`
	Features                []byte    `json:"features" db:"features"`
	ModulesEnabled          []byte    `json:"modules_enabled" db:"modules_enabled"`
	AccountingMethod        string    `json:"accounting_method" db:"accounting_method"`
	FiscalYearStartMonth    int32     `json:"fiscal_year_start_month" db:"fiscal_year_start_month"`
	DefaultCurrency         string    `json:"default_currency" db:"default_currency"`
	DateFormat              string    `json:"date_format" db:"date_format"`
	NumberFormat            string    `json:"number_format" db:"number_format"`
	LanguageCode            string    `json:"language_code" db:"language_code"`
	PasswordPolicy          []byte    `json:"password_policy" db:"password_policy"`
	WebhookEndpoints        []byte    `json:"webhook_endpoints" db:"webhook_endpoints"`
	ApiRateLimits           []byte    `json:"api_rate_limits" db:"api_rate_limits"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time `json:"updated_at" db:"updated_at"`
}

// UsageStats represents tenant usage statistics
// This maps to the `tenant_usage_stats` table in the database schema
type UsageStats struct {
	TenantID          uuid.UUID `json:"tenant_id" db:"tenant_id"`
	PeriodStart       time.Time `json:"period_start" db:"period_start"`
	PeriodEnd         time.Time `json:"period_end" db:"period_end"`
	ActiveUsers       int32     `json:"active_users" db:"active_users"`
	TotalEntities     int32     `json:"total_entities" db:"total_entities"`
	TotalTransactions int32     `json:"total_transactions" db:"total_transactions"`
	StorageUsed       int64     `json:"storage_used" db:"storage_used"`
	ApiCalls          int32     `json:"api_calls" db:"api_calls"`
	AvgResponseTime   float64   `json:"avg_response_time" db:"avg_response_time"`
	ErrorRate         float64   `json:"error_rate" db:"error_rate"`
	MonthlyRevenue    float64   `json:"monthly_revenue" db:"monthly_revenue"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// Constants for tenant status
const (
	StatusActive    = "active"
	StatusSuspended = "suspended"
	StatusPending   = "pending"
)

// Constants for accounting method
const (
	AccountingMethodAccrual = "accrual"
	AccountingMethodCash    = "cash"
)

// Constants for company size
const (
	CompanySizeStartup    = "startup"
	CompanySizeSmall      = "small"
	CompanySizeMedium     = "medium"
	CompanySizeLarge      = "large"
	CompanySizeEnterprise = "enterprise"
)

// Request and Response DTOs

// CreateTenantRequest represents tenant creation request
type CreateTenantRequest struct {
	Name               string                 `json:"name" validate:"required,min=1,max=255"`
	Email              string                 `json:"email" validate:"required,email"`
	Subdomain          string                 `json:"subdomain,omitempty" validate:"omitempty,min=1,max=63"`
	Timezone           string                 `json:"timezone" validate:"required"`
	CurrencyCode       string                 `json:"currency_code" validate:"required,len=3"`
	Industry           string                 `json:"industry,omitempty"`
	CompanySize        string                 `json:"company_size,omitempty"`
	TaxID              string                 `json:"tax_id,omitempty"`
	RegistrationNumber string                 `json:"registration_number,omitempty"`
	LegalEntityType    string                 `json:"legal_entity_type,omitempty"`
	Settings           map[string]interface{} `json:"settings,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTenantRequest represents tenant update request
type UpdateTenantRequest struct {
	Name               *string                `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Email              *string                `json:"email,omitempty" validate:"omitempty,email"`
	Subdomain          *string                `json:"subdomain,omitempty" validate:"omitempty,min=1,max=63"`
	Status             *string                `json:"status,omitempty"`
	Timezone           *string                `json:"timezone,omitempty"`
	CurrencyCode       *string                `json:"currency_code,omitempty" validate:"omitempty,len=3"`
	Industry           *string                `json:"industry,omitempty"`
	CompanySize        *string                `json:"company_size,omitempty"`
	TaxID              *string                `json:"tax_id,omitempty"`
	RegistrationNumber *string                `json:"registration_number,omitempty"`
	LegalEntityType    *string                `json:"legal_entity_type,omitempty"`
	Settings           map[string]interface{} `json:"settings,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// TenantFilter represents filtering options for tenant queries
type TenantFilter struct {
	StatusFilter   *string    `json:"status_filter,omitempty"`
	IndustryFilter *string    `json:"industry_filter,omitempty"`
	CreatedAfter   *time.Time `json:"created_after,omitempty"`
	CreatedBefore  *time.Time `json:"created_before,omitempty"`
	SortBy         string     `json:"sort_by,omitempty"`
	SortOrder      string     `json:"sort_order,omitempty"`
	Offset         int        `json:"offset"`
	Limit          int        `json:"limit"`
}

// Helper methods

// IsActive checks if tenant is active
func (t *Tenant) IsActive() bool {
	return t.Status == StatusActive && !t.DeletedAt.Valid
}

// IsSuspended checks if tenant is suspended
func (t *Tenant) IsSuspended() bool {
	return t.Status == StatusSuspended
}

// GetDisplayName returns the display name for the tenant
func (t *Tenant) GetDisplayName() string {
	if t.Name != "" {
		return t.Name
	}
	return t.Slug
}

// Common error types
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

// Error constants
const (
	ErrTenantNotFound         = "TENANT_NOT_FOUND"
	ErrTenantAlreadyExists    = "TENANT_ALREADY_EXISTS"
	ErrSubdomainAlreadyExists = "SUBDOMAIN_ALREADY_EXISTS"
	ErrTenantInactive         = "TENANT_INACTIVE"
	ErrInvalidTenantData      = "INVALID_TENANT_DATA"
	ErrAccessDenied           = "ACCESS_DENIED"
)