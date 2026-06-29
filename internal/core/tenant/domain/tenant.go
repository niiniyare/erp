package domain

import (
	"regexp"
	"strings"
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
	PlanTier           string         `json:"plan_tier"`
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
type Option func(*Tenant) error

// NewTenant creates a Tenant with validated defaults and applies options.
// Returns an error if required fields are missing or invalid.
func NewTenant(name, email string, opts ...Option) (*Tenant, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrTenantNameRequired
	}
	if strings.TrimSpace(email) == "" {
		return nil, ErrTenantEmailRequired
	}
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	now := time.Now()
	t := &Tenant{
		ID:           uuid.New(),
		Slug:         slug.Make(name),
		Name:         strings.TrimSpace(name),
		Email:        strings.ToLower(strings.TrimSpace(email)),
		Status:       StatusPending,
		Timezone:     "UTC",
		CurrencyCode: "USD",
		Metadata:     make(map[string]any),
		Settings:     make(map[string]any),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	for _, o := range opts {
		if err := o(t); err != nil {
			return nil, err
		}
	}
	return t, nil
}

// Functional options

func WithSubdomain(s string) Option {
	return func(t *Tenant) error {
		if !isValidSubdomain(s) {
			return ErrInvalidSubdomain
		}
		if ReservedSubdomains[strings.ToLower(s)] {
			return ErrInvalidSubdomain
		}
		lower := strings.ToLower(s)
		t.Subdomain = &lower
		return nil
	}
}

func WithPlan(tz, currency string) Option {
	return func(t *Tenant) error {
		if tz != "" {
			t.Timezone = tz
		}
		if currency != "" {
			t.CurrencyCode = currency
		}
		return nil
	}
}

func WithLimits(industry, companySize *string) Option {
	return func(t *Tenant) error {
		t.Industry = industry
		if companySize != nil && *companySize != "" {
			if !ValidCompanySize(*companySize) {
				return ErrInvalidCompanySize
			}
			t.CompanySize = companySize
		}
		return nil
	}
}

func WithTimezone(tz string) Option    { return func(t *Tenant) error { t.Timezone = tz; return nil } }
func WithCurrency(c string) Option     { return func(t *Tenant) error { t.CurrencyCode = c; return nil } }
func WithStatus(s TenantStatus) Option { return func(t *Tenant) error { t.Status = s; return nil } }

// Business methods

// Activate transitions the tenant to Active status.
func (t *Tenant) Activate() error {
	if t.Status == StatusActive {
		return ErrAlreadyActive
	}
	if t.Status == StatusArchived {
		return ErrCannotActivateArchivedTenant
	}
	if !t.Status.CanTransitionTo(StatusActive) {
		return ErrInvalidTransition
	}
	t.Status = StatusActive
	t.UpdatedAt = time.Now()
	return nil
}

// Suspend transitions the tenant to Suspended status.
func (t *Tenant) Suspend(reason string) error {
	if t.Status == StatusSuspended {
		return ErrAlreadySuspended
	}
	if t.Status == StatusArchived {
		return ErrCannotSuspendArchivedTenant
	}
	if !t.Status.CanTransitionTo(StatusSuspended) {
		return ErrInvalidTransition
	}
	t.Status = StatusSuspended
	t.UpdatedAt = time.Now()

	// Store suspension reason in metadata
	if t.Metadata == nil {
		t.Metadata = make(map[string]any)
	}
	t.Metadata["suspension_reason"] = reason
	t.Metadata["suspended_at"] = time.Now()
	return nil
}

// Archive transitions the tenant to Archived (terminal state).
func (t *Tenant) Archive() error {
	if t.Status == StatusArchived {
		return ErrAlreadyArchived
	}
	if !t.Status.CanTransitionTo(StatusArchived) {
		return ErrInvalidTransition
	}
	now := time.Now()
	t.Status = StatusArchived
	t.DeletedAt = &now
	t.UpdatedAt = now
	return nil
}

// UpdateActivity updates the last activity timestamp.
func (t *Tenant) UpdateActivity() {
	now := time.Now()
	t.LastActivityAt = &now
	t.UpdatedAt = now
}

// Predicates

func (t *Tenant) IsActive() bool      { return t.Status == StatusActive }
func (t *Tenant) IsSuspended() bool   { return t.Status == StatusSuspended }
func (t *Tenant) IsArchived() bool    { return t.Status == StatusArchived }
func (t *Tenant) IsSoftDeleted() bool { return t.DeletedAt != nil }

// Validation helpers

var subdomainRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

func isValidEmail(email string) bool {
	if len(email) == 0 || len(email) > 255 {
		return false
	}
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func isValidSubdomain(subdomain string) bool {
	if len(subdomain) == 0 || len(subdomain) > 63 {
		return false
	}
	return subdomainRegex.MatchString(strings.ToLower(subdomain))
}
