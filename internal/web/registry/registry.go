// Package registry maps route paths to schema builder functions.
// Every page schema is registered here once. The handler dispatches by path.
//
// Usage:
//
//	func init() {
//	    registry.Register("/dashboard", dashboard.Schema)
//	    registry.Register("/finance/invoices", invoices.Schema)
//	}
package registry

import (
	"fmt"
	"sync"

	"awo.so/internal/web/ui"
)

var (
	mu       sync.RWMutex
	handlers = map[string]ui.PageFn{}
)

// Register adds a schema function for a route path.
// path should match the URL segment after /schema/, e.g. "/dashboard".
// Panics on duplicate registration to catch typos at startup.
func Register(path string, fn ui.PageFn) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := handlers[path]; exists {
		panic(fmt.Sprintf("ui/registry: duplicate schema registration for %q", path))
	}
	handlers[path] = fn
}

// Get looks up the schema function for path. Returns nil if not registered.
func Get(path string) ui.PageFn {
	mu.RLock()
	defer mu.RUnlock()
	return handlers[path]
}

// Paths returns all registered paths.
func Paths() []string {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]string, 0, len(handlers))
	for p := range handlers {
		out = append(out, p)
	}
	return out
}
