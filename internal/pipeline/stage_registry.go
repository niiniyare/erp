package pipeline

import (
	"fmt"
	"sync"
)

// StageRegistry is the global catalogue of all pipeline stages across all
// modules. Modules register their stages at application startup (wire.go).
// The PipelineBuilder queries the registry to build per-tenant pipelines.
//
// Thread-safe: concurrent reads are safe after the startup registration phase.
type StageRegistry struct {
	mu     sync.RWMutex
	stages map[string]Stage // keyed by stage Name()
}

// NewStageRegistry returns an empty StageRegistry.
func NewStageRegistry() *StageRegistry {
	return &StageRegistry{
		stages: make(map[string]Stage),
	}
}

// Register adds one or more stages to the registry.
// Panics immediately on a duplicate name — this is a programmer error that
// must be caught at startup, not at runtime.
func (r *StageRegistry) Register(stages ...Stage) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, s := range stages {
		name := s.Name()
		if _, exists := r.stages[name]; exists {
			panic(fmt.Sprintf("pipeline: duplicate stage name %q — each stage must have a unique name", name))
		}
		r.stages[name] = s
	}
}

// ForOperation returns all stages that apply to the given operation key.
// A stage applies when its Operations() slice contains the exact operationKey
// or the wildcard "*".
// The returned slice is a new copy; callers may sort or filter it freely.
func (r *StageRegistry) ForOperation(operationKey string) []Stage {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Stage
	for _, s := range r.stages {
		if matchesOperation(s.Operations(), operationKey) {
			result = append(result, s)
		}
	}
	return result
}

// All returns every registered stage in an unspecified order.
func (r *StageRegistry) All() []Stage {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Stage, 0, len(r.stages))
	for _, s := range r.stages {
		result = append(result, s)
	}
	return result
}

// matchesOperation returns true when ops contains key or the wildcard "*".
func matchesOperation(ops []string, key string) bool {
	for _, op := range ops {
		if op == "*" || op == key {
			return true
		}
	}
	return false
}
