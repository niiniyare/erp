// Package customfield provides a runtime registry for per-entity, per-tenant
// custom field definitions.
//
// Custom field definitions are stored in the database as JSONB in the
// awo_custom_fields table. The registry caches them in-process and must be
// refreshed on change.
//
// Required migration: see db/migrations/000002_framework.up.sql
package customfield

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"awo.so/framework/def"
)

// LoadFunc loads the raw JSONB custom-field definitions for (tenantID, entity)
// from the database.
type LoadFunc func(ctx context.Context, tenantID uuid.UUID, entity string) ([]byte, error)

// SaveFunc persists the serialised custom-field definitions.
type SaveFunc func(ctx context.Context, tenantID uuid.UUID, entity string, data []byte) error

// Registry caches per-(tenant, entity) custom field definitions in memory.
// Use New to construct; safe for concurrent use after construction.
type Registry struct {
	mu    sync.RWMutex
	cache map[cacheKey][]*def.FieldDef
	load  LoadFunc
	save  SaveFunc
}

type cacheKey struct {
	tenantID uuid.UUID
	entity   string
}

// New creates a Registry backed by load and save functions.
func New(load LoadFunc, save SaveFunc) *Registry {
	return &Registry{
		cache: make(map[cacheKey][]*def.FieldDef),
		load:  load,
		save:  save,
	}
}

// Get returns the custom field definitions for (tenantID, entity).
// Results are cached in memory; call Invalidate to force a reload.
func (r *Registry) Get(ctx context.Context, tenantID uuid.UUID, entity string) ([]*def.FieldDef, error) {
	key := cacheKey{tenantID, entity}

	r.mu.RLock()
	fields, ok := r.cache[key]
	r.mu.RUnlock()
	if ok {
		return fields, nil
	}

	raw, err := r.load(ctx, tenantID, entity)
	if err != nil {
		return nil, fmt.Errorf("customfield: load %s/%s: %w", entity, tenantID, err)
	}

	var defs []*def.FieldDef
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &defs); err != nil {
			return nil, fmt.Errorf("customfield: unmarshal %s/%s: %w", entity, tenantID, err)
		}
	}

	r.mu.Lock()
	r.cache[key] = defs
	r.mu.Unlock()

	return defs, nil
}

// Set persists and caches new custom field definitions for (tenantID, entity).
func (r *Registry) Set(ctx context.Context, tenantID uuid.UUID, entity string, fields []*def.FieldDef) error {
	raw, err := json.Marshal(fields)
	if err != nil {
		return fmt.Errorf("customfield: marshal %s/%s: %w", entity, tenantID, err)
	}

	if err := r.save(ctx, tenantID, entity, raw); err != nil {
		return fmt.Errorf("customfield: save %s/%s: %w", entity, tenantID, err)
	}

	key := cacheKey{tenantID, entity}
	r.mu.Lock()
	r.cache[key] = fields
	r.mu.Unlock()

	return nil
}

// Invalidate clears the cached definitions for (tenantID, entity).
func (r *Registry) Invalidate(tenantID uuid.UUID, entity string) {
	key := cacheKey{tenantID, entity}
	r.mu.Lock()
	delete(r.cache, key)
	r.mu.Unlock()
}

// InvalidateAll clears all cached definitions.
func (r *Registry) InvalidateAll() {
	r.mu.Lock()
	r.cache = make(map[cacheKey][]*def.FieldDef)
	r.mu.Unlock()
}

// MergedFields returns base entity fields merged with tenant-specific custom
// fields. Custom fields with names that clash with base fields are skipped.
func MergedFields(base []*def.FieldDef, custom []*def.FieldDef) []*def.FieldDef {
	if len(custom) == 0 {
		return base
	}
	seen := make(map[string]struct{}, len(base))
	for _, f := range base {
		seen[f.Name] = struct{}{}
	}
	merged := make([]*def.FieldDef, len(base), len(base)+len(custom))
	copy(merged, base)
	for _, f := range custom {
		if _, clash := seen[f.Name]; !clash {
			merged = append(merged, f)
		}
	}
	return merged
}
