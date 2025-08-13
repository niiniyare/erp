package tenant

import "github.com/google/uuid"

// ProvisionTenantRequest represents the data needed to provision a new tenant.
type ProvisionTenantRequest struct {
	Name         string  `json:"name" validate:"required,min=2,max=100"`
	Email        string  `json:"email" validate:"required,email"`
	Subdomain    *string `json:"subdomain,omitempty"`
	Industry     *string `json:"industry,omitempty"`
	CompanySize  *string `json:"company_size,omitempty"`
	CountryCode  string  `json:"country_code" validate:"iso3166_1_alpha2"`
	CurrencyCode string  `json:"currency_code" validate:"iso4217"`
}

// ProvisionedTenantInfo holds the essential information returned after a successful provisioning.
type ProvisionedTenantInfo struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	Slug      string    `json:"slug"`
	Subdomain *string   `json:"subdomain,omitempty"`
	Status    string    `json:"status"`
}
