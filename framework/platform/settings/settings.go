// Package settings provides per-tenant key/value configuration storage.
package settings

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// Store is the persistence interface for tenant settings.
type Store interface {
	Get(ctx context.Context, tenantID uuid.UUID, key string) (json.RawMessage, error)
	Set(ctx context.Context, tenantID uuid.UUID, key string, value json.RawMessage) error
	Delete(ctx context.Context, tenantID uuid.UUID, key string) error
}

// Service wraps Store with typed helpers.
type Service struct {
	store Store
}

func NewService(store Store) *Service { return &Service{store: store} }

// GetString returns a string setting, or def when the key is absent.
func (s *Service) GetString(ctx context.Context, tenantID uuid.UUID, key, def string) string {
	raw, err := s.store.Get(ctx, tenantID, key)
	if err != nil || len(raw) == 0 {
		return def
	}
	var v string
	if err := json.Unmarshal(raw, &v); err != nil {
		return def
	}
	return v
}

// GetBool returns a boolean setting, or def when the key is absent.
func (s *Service) GetBool(ctx context.Context, tenantID uuid.UUID, key string, def bool) bool {
	raw, err := s.store.Get(ctx, tenantID, key)
	if err != nil || len(raw) == 0 {
		return def
	}
	var v bool
	if err := json.Unmarshal(raw, &v); err != nil {
		return def
	}
	return v
}

// SetString saves a string setting.
func (s *Service) SetString(ctx context.Context, tenantID uuid.UUID, key, value string) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("settings: marshal: %w", err)
	}
	return s.store.Set(ctx, tenantID, key, raw)
}
