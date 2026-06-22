package definition

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

// Lookup returns the EntityDefinition for name from the global registry.
// Returns nil if not found.
func Lookup(name string) *EntityDefinition {
	return global.Lookup(name)
}

// All returns all registered definitions from the global registry.
func All() []*EntityDefinition {
	return global.All()
}

// Register validates and adds def to the registry.
func (r *Registry) Register(def *EntityDefinition) {
	if err := def.Validate(); err != nil {
		panic(fmt.Sprintf("definition.Register: invalid entity %q: %v", def.Name, err))
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.entries[def.Name]; exists {
		panic(fmt.Sprintf("definition.Register: entity %q already registered", def.Name))
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
