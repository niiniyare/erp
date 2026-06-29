// Package ff provides feature flag evaluation for tenants.
//
// Flags can be global (tenant_id IS NULL) or per-tenant. A per-tenant flag
// takes precedence over a global flag with the same key.
package ff

import (
	"context"

	"github.com/google/uuid"
)

// Flag represents a feature flag record.
type Flag struct {
	Key        string     `json:"key"`
	TenantID   *uuid.UUID `json:"tenant_id,omitempty"`
	Enabled    bool       `json:"enabled"`
	RolloutPct int        `json:"rollout_pct"`
}

// Store is the persistence interface for feature flags.
type Store interface {
	// Get returns the most specific flag for (tenantID, key):
	// tenant-scoped first, then global.
	Get(ctx context.Context, tenantID uuid.UUID, key string) (*Flag, error)
	Set(ctx context.Context, f *Flag) error
	SetGlobal(ctx context.Context, f *Flag) error
}

// Service evaluates feature flags.
type Service struct {
	store Store
}

func NewService(store Store) *Service { return &Service{store: store} }

// IsEnabled returns true when the flag is enabled for tenantID.
// Missing flags are treated as disabled.
func (s *Service) IsEnabled(ctx context.Context, tenantID uuid.UUID, key string) bool {
	f, err := s.store.Get(ctx, tenantID, key)
	if err != nil || f == nil {
		return false
	}
	return f.Enabled
}
