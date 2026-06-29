package def

import (
	"fmt"
	"sync"
)

// Registry holds all registered EntityDefinitions for a running application.
// The default global registry is used by Register/Lookup; custom registries
// can be created for testing.
type Registry struct {
	mu      sync.RWMutex
	entries map[string]*EntityDefinition
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{entries: make(map[string]*EntityDefinition)}
}

// global is the default process-wide registry.
var global = NewRegistry()

// Register validates and adds def to the registry.
// Panics on duplicate name or validation failure — registration is expected to
// happen at startup (init / wire), so a panic is appropriate.
func Register(def *EntityDefinition) {
	global.Register(def)
}

// RegisterSystem is a convenience alias for Register that explicitly marks the
// entity as EntityTypeSystem (dedicated SQL table). Equivalent to setting
// def.Type = EntityTypeSystem before calling Register.
func RegisterSystem(def *EntityDefinition) {
	def.Type = EntityTypeSystem
	global.Register(def)
}

// RegisterCustom is a convenience alias for Register that explicitly marks the
// entity as EntityTypeCustom (JSONB storage in custom_entity_records).
// Custom entities do NOT require SQL migrations — rows are stored generically.
func RegisterCustom(def *EntityDefinition) {
	def.Type = EntityTypeCustom
	global.Register(def)
}

// Lookup returns the EntityDefinition for name from the global registry.
// Returns nil if not found.
func Lookup(name string) *EntityDefinition {
	return global.Lookup(name)
}

// All returns all registered definitions from the global registry.
func All() []*EntityDefinition {
	return global.All()
}

// AllSystem returns only EntityTypeSystem definitions from the global registry.
func AllSystem() []*EntityDefinition {
	return global.AllByType(EntityTypeSystem)
}

// AllCustom returns only EntityTypeCustom definitions from the global registry.
func AllCustom() []*EntityDefinition {
	return global.AllByType(EntityTypeCustom)
}

// Register validates and adds def to the registry.
func (r *Registry) Register(def *EntityDefinition) {
	if err := def.Validate(); err != nil {
		panic(fmt.Sprintf("def.Register: invalid entity %q: %v", def.Name, err))
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.entries[def.Name]; exists {
		panic(fmt.Sprintf("def.Register: entity %q already registered", def.Name))
	}

	r.entries[def.Name] = def
}

// Lookup returns the EntityDefinition for name, or nil if not found.
func (r *Registry) Lookup(name string) *EntityDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.entries[name]
}

// All returns a snapshot of all registered definitions in arbitrary order.
func (r *Registry) All() []*EntityDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*EntityDefinition, 0, len(r.entries))
	for _, d := range r.entries {
		out = append(out, d)
	}
	return out
}

// AllByType returns a snapshot of registered definitions whose Type matches t.
// EntityTypeSystem matches definitions with Type == "" (zero value) as well,
// since the zero value defaults to system.
func (r *Registry) AllByType(t EntityType) []*EntityDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*EntityDefinition, 0, len(r.entries))
	for _, d := range r.entries {
		switch t {
		case EntityTypeSystem:
			if d.Type.IsSystem() {
				out = append(out, d)
			}
		case EntityTypeCustom:
			if d.Type.IsCustom() {
				out = append(out, d)
			}
		}
	}
	return out
}
