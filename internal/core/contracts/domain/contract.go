package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Contract is the core domain entity for the contracts module.
type Contract struct {
	ID                uuid.UUID
	TenantID          uuid.UUID
	EntityID          uuid.UUID
	Number            string
	Title             string
	Status            ContractStatus
	ContractType      ContractType
	CounterpartyName  string
	CounterpartyEmail *string
	StartDate         time.Time
	EndDate           *time.Time
	Value             decimal.Decimal
	CurrencyCode      string
	Description       *string
	Terms             *string
	SignedBy          *uuid.UUID
	SignedAt          *time.Time
	Metadata          map[string]any
	Version           int32
	CreatedBy         *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

// Option is a functional option for NewContract.
type Option func(*Contract) error

// NewContract creates a validated Contract with sensible defaults.
func NewContract(title, counterparty, currencyCode, number string, entityID uuid.UUID, startDate time.Time, opts ...Option) (*Contract, error) {
	if strings.TrimSpace(title) == "" {
		return nil, ErrTitleRequired
	}
	if strings.TrimSpace(counterparty) == "" {
		return nil, ErrCounterpartyRequired
	}
	if strings.TrimSpace(number) == "" {
		return nil, ErrNumberRequired
	}
	if entityID == uuid.Nil {
		return nil, ErrInvalidRequest
	}

	now := time.Now()
	c := &Contract{
		ID:               uuid.New(),
		EntityID:         entityID,
		Number:           strings.TrimSpace(number),
		Title:            strings.TrimSpace(title),
		Status:           StatusDraft,
		ContractType:     TypeVendor,
		CounterpartyName: strings.TrimSpace(counterparty),
		StartDate:        startDate,
		Value:            decimal.Zero,
		CurrencyCode:     currencyCode,
		Metadata:         make(map[string]any),
		Version:          1,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	for _, o := range opts {
		if err := o(c); err != nil {
			return nil, err
		}
	}

	if c.EndDate != nil && c.EndDate.Before(c.StartDate) {
		return nil, ErrEndDateBeforeStart
	}

	return c, nil
}

// Functional options

func WithContractType(t ContractType) Option {
	return func(c *Contract) error {
		if !t.Valid() {
			return ErrInvalidContractType
		}
		c.ContractType = t
		return nil
	}
}

func WithContractValue(v decimal.Decimal, currency string) Option {
	return func(c *Contract) error {
		if v.IsNegative() {
			return ErrNegativeValue
		}
		if len(currency) != 3 {
			return ErrInvalidCurrency
		}
		c.Value = v
		c.CurrencyCode = strings.ToUpper(currency)
		return nil
	}
}

func WithEndDate(d time.Time) Option {
	return func(c *Contract) error {
		c.EndDate = &d
		return nil
	}
}

func WithCounterpartyEmail(email string) Option {
	return func(c *Contract) error {
		e := strings.ToLower(strings.TrimSpace(email))
		c.CounterpartyEmail = &e
		return nil
	}
}

func WithDescription(desc string) Option {
	return func(c *Contract) error {
		c.Description = &desc
		return nil
	}
}

func WithTerms(terms string) Option {
	return func(c *Contract) error {
		c.Terms = &terms
		return nil
	}
}

func WithCreatedBy(id uuid.UUID) Option {
	return func(c *Contract) error {
		c.CreatedBy = &id
		return nil
	}
}

// Business methods

// Submit moves DRAFT → PENDING_APPROVAL.
func (c *Contract) Submit() error {
	if !c.Status.CanTransitionTo(StatusPendingApproval) {
		return ErrInvalidTransition
	}
	c.Status = StatusPendingApproval
	c.UpdatedAt = time.Now()
	return nil
}

// Approve moves PENDING_APPROVAL → ACTIVE and records the signer.
func (c *Contract) Approve(signerID uuid.UUID) error {
	if !c.Status.CanTransitionTo(StatusActive) {
		return ErrInvalidTransition
	}
	now := time.Now()
	c.Status = StatusActive
	c.SignedBy = &signerID
	c.SignedAt = &now
	c.UpdatedAt = now
	return nil
}

// Reject moves PENDING_APPROVAL back to DRAFT.
func (c *Contract) Reject() error {
	if !c.Status.CanTransitionTo(StatusDraft) {
		return ErrInvalidTransition
	}
	c.Status = StatusDraft
	c.UpdatedAt = time.Now()
	return nil
}

// Expire transitions ACTIVE → EXPIRED (typically called by a scheduled job).
func (c *Contract) Expire() error {
	if c.Status == StatusExpired {
		return ErrAlreadyExpired
	}
	if !c.Status.CanTransitionTo(StatusExpired) {
		return ErrInvalidTransition
	}
	c.Status = StatusExpired
	c.UpdatedAt = time.Now()
	return nil
}

// Terminate moves ACTIVE → TERMINATED.
func (c *Contract) Terminate(reason string) error {
	if c.Status == StatusTerminated {
		return ErrAlreadyTerminated
	}
	if !c.Status.CanTransitionTo(StatusTerminated) {
		return ErrInvalidTransition
	}
	c.Status = StatusTerminated
	if c.Metadata == nil {
		c.Metadata = make(map[string]any)
	}
	c.Metadata["termination_reason"] = reason
	c.Metadata["terminated_at"] = time.Now()
	c.UpdatedAt = time.Now()
	return nil
}

// Predicates

func (c *Contract) IsDraft() bool           { return c.Status == StatusDraft }
func (c *Contract) IsPendingApproval() bool { return c.Status == StatusPendingApproval }
func (c *Contract) IsActive() bool          { return c.Status == StatusActive }
func (c *Contract) IsExpired() bool         { return c.Status == StatusExpired }
func (c *Contract) IsTerminated() bool      { return c.Status == StatusTerminated }
func (c *Contract) IsEditable() bool {
	return c.Status == StatusDraft || c.Status == StatusPendingApproval
}
func (c *Contract) IsSoftDeleted() bool { return c.DeletedAt != nil }
