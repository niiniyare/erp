package tenant

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
)

// Tenant represents a tenant in the system (domain model)
type Tenant struct {
	ID                 uuid.UUID      `json:"id"`
	Slug               string         `json:"slug"`
	Name               string         `json:"name"`
	Email              string         `json:"email"`
	Subdomain          *string        `json:"subdomain"`
	Status             Status         `json:"status"`
	Timezone           string         `json:"timezone"`
	CurrencyCode       string         `json:"currency_code"`
	Metadata           map[string]any `json:"metadata"`
	Industry           *string        `json:"industry"`
	CompanySize        *string        `json:"company_size"`
	TaxID              *string        `json:"tax_id"`
	RegistrationNumber *string        `json:"registration_number"`
	LegalEntityType    *string        `json:"legal_entity_type"`
	Settings           map[string]any `json:"settings"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          *time.Time     `json:"deleted_at,omitempty"`
}

// FromSQLCTenant converts SQLC Tenant to domain Tenant
func FromSQLCTenant(sqlcTenant *db.Tenant) (*Tenant, error) {
	var metadata map[string]any
	if len(sqlcTenant.Metadata) > 0 {
		if err := json.Unmarshal(sqlcTenant.Metadata, &metadata); err != nil {
			return nil, err
		}
	}

	var settings map[string]any
	if len(sqlcTenant.Settings) > 0 {
		if err := json.Unmarshal(sqlcTenant.Settings, &settings); err != nil {
			return nil, err
		}
	}

	var deletedAt *time.Time
	if sqlcTenant.DeletedAt.Valid {
		deletedAt = &sqlcTenant.DeletedAt.Time
	}

	return &Tenant{
		ID:                 sqlcTenant.ID,
		Slug:               sqlcTenant.Slug,
		Name:               sqlcTenant.Name,
		Email:              sqlcTenant.Email,
		Subdomain:          sqlcTenant.Subdomain,
		Status:             Status(sqlcTenant.Status),
		Timezone:           sqlcTenant.Timezone,
		CurrencyCode:       sqlcTenant.CurrencyCode,
		Metadata:           metadata,
		Industry:           sqlcTenant.Industry,
		CompanySize:        sqlcTenant.CompanySize,
		TaxID:              sqlcTenant.TaxID,
		RegistrationNumber: sqlcTenant.RegistrationNumber,
		LegalEntityType:    sqlcTenant.LegalEntityType,
		Settings:           settings,
		CreatedAt:          sqlcTenant.CreatedAt,
		UpdatedAt:          sqlcTenant.UpdatedAt,
		DeletedAt:          deletedAt,
	}, nil
}

// ToSQLCCreateParams converts domain CreateTenantRequest to SQLC params
func (req *CreateTenantRequest) ToSQLCCreateParams() (db.CreateTenantParams, error) {
	return db.CreateTenantParams{
		Name:      req.Name,
		Slug:      req.Slug,
		Email:     req.Email,
		Subdomain: req.Subdomain,
		Status:    string(req.Status),
		Industry:  req.Industry,
	}, nil
}

// PlanType represents subscription plans
type PlanType string

const (
	PlanTypeBasic        PlanType = "basic"
	PlanTypeProfessional PlanType = "professional"
	PlanTypeEnterprise   PlanType = "enterprise"
)

// Status represents tenant status
type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusTrial     Status = "trial"
)

// CreateTenantRequest represents tenant creation request
type CreateTenantRequest struct {
	Name               string         `json:"name" validate:"required,min=2,max=100"`
	Slug               string         `json:"slug,omitempty"`
	Email              string         `json:"email" validate:"required,email"`
	Subdomain          *string        `json:"subdomain,omitempty"`
	Status             Status         `json:"status,omitempty"`
	Industry           *string        `json:"industry,omitempty"`
	CompanySize        *string        `json:"company_size,omitempty"`
	TaxID              *string        `json:"tax_id,omitempty"`
	RegistrationNumber *string        `json:"registration_number,omitempty"`
	LegalEntityType    *string        `json:"legal_entity_type,omitempty"`
	Settings           map[string]any `json:"settings,omitempty"`
	CountryCode        string         `json:"country_code" validate:"iso3166_1_alpha2"`
	CurrencyCode       string         `json:"currency_code" validate:"iso4217"`
}

// UpdateTenantRequest represents tenant update request
type UpdateTenantRequest struct {
	Name               *string        `json:"name,omitempty"`
	Email              *string        `json:"email,omitempty"`
	Subdomain          *string        `json:"subdomain,omitempty"`
	Status             *Status        `json:"status,omitempty"`
	Industry           *string        `json:"industry,omitempty"`
	CompanySize        *string        `json:"company_size,omitempty"`
	TaxID              *string        `json:"tax_id,omitempty"`
	RegistrationNumber *string        `json:"registration_number,omitempty"`
	LegalEntityType    *string        `json:"legal_entity_type,omitempty"`
	Settings           map[string]any `json:"settings,omitempty"`
}

// TenantLimits represents tenant usage limits
type TenantLimits struct {
	MaxUsers           uint   `json:"max_users"`
	MaxStorageMB       uint64 `json:"max_storage_mb"`
	MaxAPICallsPerHour uint   `json:"max_api_calls_per_hour"`
}
