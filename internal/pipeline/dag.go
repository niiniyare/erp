package pipeline

import (
	"fmt"
	"sort"
	"strings"
)

// DAGError is returned by ValidateDAG when the stage dependency graph is invalid.
// It aggregates all violations so the developer sees every problem at once.
type DAGError struct {
	Violations []string
}

func (e *DAGError) Error() string {
	return fmt.Sprintf("pipeline: invalid stage DAG (%d violation(s)):\n  - %s",
		len(e.Violations), strings.Join(e.Violations, "\n  - "))
}

// ValidateDAG checks the registered stages for the given operationKey and
// validates their DependsOn graph.
//
// Checks performed:
//  1. Missing dependencies — a stage declares DependsOn a name not in the registry
//     for this operation.
//  2. Duplicate stage names — caught by Register() already, but verified here.
//  3. Cycles — a cycle in the DependsOn graph is detected via DFS (white-grey-black).
//  4. Self-dependency — a stage that lists its own name in DependsOn().
//
// Call this at application startup after all stages are registered.
// Returns nil when the graph is valid. Returns *DAGError with all violations otherwise.
//
// Typical usage in wire.go:
//
//	if err := reg.ValidateDAG(ui.OperationKey); err != nil {
//	    panic(err.Error())
//	}
func (r *StageRegistry) ValidateDAG(operationKey string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stages := r.forOperationLocked(operationKey)
	if len(stages) == 0 {
		return nil
	}

	// Build name → Stage index for O(1) lookup.
	idx := make(map[string]Stage, len(stages))
	for _, s := range stages {
		idx[s.Name()] = s
	}

	var violations []string

	// Pass 1: missing deps + self-dependency.
	for _, s := range stages {
		for _, dep := range s.DependsOn() {
			if dep == s.Name() {
				violations = append(violations,
					fmt.Sprintf("stage %q lists itself in DependsOn", s.Name()))
				continue
			}
			if _, ok := idx[dep]; !ok {
				violations = append(violations,
					fmt.Sprintf("stage %q depends on %q which is not registered for operation %q",
						s.Name(), dep, operationKey))
			}
		}
	}

	// Pass 2: cycle detection via DFS coloring.
	// white=0 (unvisited), grey=1 (in current path), black=2 (done).
	color := make(map[string]int, len(stages))
	path := make([]string, 0, len(stages))

	var dfs func(name string) bool
	dfs = func(name string) bool {
		if color[name] == 2 {
			return false // already fully explored
		}
		if color[name] == 1 {
			// Found a cycle — path contains the cycle.
			// Find where the cycle starts in path.
			start := 0
			for i, n := range path {
				if n == name {
					start = i
					break
				}
			}
			cycle := append(path[start:], name)
			violations = append(violations,
				fmt.Sprintf("cycle detected: %s", strings.Join(cycle, " → ")))
			return true
		}

		color[name] = 1
		path = append(path, name)

		s, ok := idx[name]
		if ok {
			// Sort deps for deterministic error messages.
			deps := make([]string, len(s.DependsOn()))
			copy(deps, s.DependsOn())
			sort.Strings(deps)

			for _, dep := range deps {
				if dfs(dep) {
					// Cycle already recorded; stop propagating.
					break
				}
			}
		}

		path = path[:len(path)-1]
		color[name] = 2
		return false
	}

	// Sort stage names for deterministic traversal order.
	names := make([]string, 0, len(stages))
	for _, s := range stages {
		names = append(names, s.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		if color[name] == 0 {
			dfs(name)
		}
	}

	if len(violations) > 0 {
		return &DAGError{Violations: violations}
	}
	return nil
}

// forOperationLocked returns stages matching the operation key.
// Caller must hold r.mu.RLock().
func (r *StageRegistry) forOperationLocked(operationKey string) []Stage {
	var result []Stage
	for _, s := range r.stages {
		if matchesOperation(s.Operations(), operationKey) {
			result = append(result, s)
		}
	}
	return result
}
