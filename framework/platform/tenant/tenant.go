// Package tenant manages the global tenant lifecycle.
//
// Tenants progress through states: pending → active → suspended → archived.
// Only the transitions listed in validTransitions are allowed.
package tenant

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Status values.
const (
	StatusPending   = "pending"
	StatusActive    = "active"
	StatusSuspended = "suspended"
	StatusArchived  = "archived"
)

var validTransitions = map[string][]string{
	StatusPending:   {StatusActive, StatusArchived},
	StatusActive:    {StatusSuspended, StatusArchived},
	StatusSuspended: {StatusActive, StatusArchived},
	StatusArchived:  {},
}

// ErrInvalidTransition is returned when an illegal status change is attempted.
var ErrInvalidTransition = errors.New("tenant: invalid status transition")

// Tenant is the domain model for a platform tenant.
type Tenant struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Status      string         `json:"status"`
	Plan        string         `json:"plan"`
	CompanySize string         `json:"company_size,omitempty"`
	Country     string         `json:"country,omitempty"`
	Currency    string         `json:"currency"`
	Timezone    string         `json:"timezone"`
	Locale      string         `json:"locale"`
	Metadata    map[string]any `json:"metadata"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// Store is the persistence interface for tenants.
type Store interface {
	Create(ctx context.Context, t *Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*Tenant, error)
	Update(ctx context.Context, t *Tenant) error
	SetStatus(ctx context.Context, id uuid.UUID, status string) error
	List(ctx context.Context, limit, offset int) ([]*Tenant, error)
}

// Service provides tenant lifecycle operations.
type Service struct {
	store Store
}

// NewService creates a tenant Service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// Create provisions a new tenant in pending status.
func (s *Service) Create(ctx context.Context, t *Tenant) error {
	t.Status = StatusPending
	if t.Currency == "" {
		t.Currency = "KES"
	}
	if t.Timezone == "" {
		t.Timezone = "Africa/Nairobi"
	}
	if t.Locale == "" {
		t.Locale = "en"
	}
	return s.store.Create(ctx, t)
}

// Transition moves a tenant to newStatus, enforcing the state machine.
func (s *Service) Transition(ctx context.Context, id uuid.UUID, newStatus string) error {
	t, err := s.store.GetByID(ctx, id)
	if err != nil {
		return err
	}
	allowed := validTransitions[t.Status]
	for _, a := range allowed {
		if a == newStatus {
			return s.store.SetStatus(ctx, id, newStatus)
		}
	}
	return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, t.Status, newStatus)
}
