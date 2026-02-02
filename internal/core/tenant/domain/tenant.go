package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

// Tenant is the core domain entity.
type Tenant struct {
	ID                 uuid.UUID      `json:"id"`
	Slug               string         `json:"slug"`
	Name               string         `json:"name"`
	Email              string         `json:"email"`
	Subdomain          *string        `json:"subdomain"`
	Status             TenantStatus   `json:"status"`
	Timezone           string         `json:"timezone"`
	CurrencyCode       string         `json:"currency_code"`
	Metadata           map[string]any `json:"metadata"`
	Industry           *string        `json:"industry"`
	CompanySize        *string        `json:"company_size"`
	TaxID              *string        `json:"tax_id"`
	RegistrationNumber *string        `json:"registration_number"`
	LegalEntityType    *string        `json:"legal_entity_type"`
	Settings           map[string]any `json:"settings"`
	LastActivityAt     *time.Time     `json:"last_activity_at,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          *time.Time     `json:"deleted_at,omitempty"`
}

// Option is a functional option for NewTenant.
type Option func(*Tenant)

// NewTenant creates a Tenant with sensible defaults and applies options.
func NewTenant(name, email string, opts ...Option) *Tenant {
	now := time.Now()
	t := &Tenant{
		ID:           uuid.New(),
		Slug:         slug.Make(name),
		Name:         name,
		Email:        email,
		Status:       StatusPending,
		Timezone:     "UTC",
		CurrencyCode: "USD",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	for _, o := range opts {
		o(t)
	}
	return t
}

func WithSubdomain(s string) Option   { return func(t *Tenant) { t.Subdomain = &s } }
func WithPlan(tz, currency string) Option {
	return func(t *Tenant) {
		if tz != "" {
			t.Timezone = tz
		}
		if currency != "" {
			t.CurrencyCode = currency
		}
	}
}
func WithLimits(industry, companySize *string) Option {
	return func(t *Tenant) {
		t.Industry = industry
		t.CompanySize = companySize
	}
}
func WithTimezone(tz string) Option    { return func(t *Tenant) { t.Timezone = tz } }
func WithCurrency(c string) Option     { return func(t *Tenant) { t.CurrencyCode = c } }
func WithStatus(s TenantStatus) Option { return func(t *Tenant) { t.Status = s } }

// Activate transitions the tenant to Active status.
func (t *Tenant) Activate() error {
	if !t.Status.CanTransitionTo(StatusActive) {
		return ErrInvalidTransition
	}
	t.Status = StatusActive
	t.UpdatedAt = time.Now()
	return nil
}

// Suspend transitions the tenant to Suspended status.
func (t *Tenant) Suspend(reason string) error {
	if !t.Status.CanTransitionTo(StatusSuspended) {
		return ErrInvalidTransition
	}
	t.Status = StatusSuspended
	t.UpdatedAt = time.Now()
	return nil
}

// Archive transitions the tenant to Archived (soft-delete).
func (t *Tenant) Archive() error {
	if !t.Status.CanTransitionTo(StatusArchived) {
		return ErrInvalidTransition
	}
	t.Status = StatusArchived
	now := time.Now()
	t.DeletedAt = &now
	t.UpdatedAt = now
	return nil
}

// Predicates

func (t *Tenant) IsActive() bool      { return t.Status == StatusActive }
func (t *Tenant) IsSuspended() bool    { return t.Status == StatusSuspended }
func (t *Tenant) IsArchived() bool     { return t.Status == StatusArchived }
func (t *Tenant) IsSoftDeleted() bool  { return t.DeletedAt != nil }
