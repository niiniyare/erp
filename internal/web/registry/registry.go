// Package registry maps route paths to page schema builders.
//
// Every page registers exactly one PageRegistration in its init() function.
// Call ValidateRegistry() at startup (after all init() functions have run,
// before the HTTP server starts) to catch misconfigured registrations early.
//
// Usage:
//
//	func init() {
//	    registry.RegisterPage(registry.PageRegistration{
//	        Route:  "/finance/invoices",
//	        Module: "finance",
//	        Title:  "Invoices",
//	        Fn:     Schema,           // legacy PageFn
//	        // ASTFn: ASTSchema,      // preferred once migrated
//	    })
//	}
package registry

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	sharedErrors "awo.so/internal/shared/errors"
	"awo.so/internal/web/ui"
)

// ─── REGISTRY_* ERROR CODES ───────────────────────────────────────────────────

const (
	// CodeRegistryDuplicateRoute is returned when two init() functions register the same route.
	CodeRegistryDuplicateRoute = "REGISTRY_DUPLICATE_ROUTE"

	// CodeRegistryInvalidRoute is returned when a route does not start with "/" or has a trailing slash.
	CodeRegistryInvalidRoute = "REGISTRY_INVALID_ROUTE"

	// CodeRegistryMissingModule is returned when Module is empty.
	CodeRegistryMissingModule = "REGISTRY_MISSING_MODULE"

	// CodeRegistryMissingTitle is returned when Title is empty.
	CodeRegistryMissingTitle = "REGISTRY_MISSING_TITLE"

	// CodeRegistryMissingFn is returned when neither Fn nor ASTFn is set.
	CodeRegistryMissingFn = "REGISTRY_MISSING_FN"

	// CodeRegistryValidationFailed is the aggregate error returned by ValidateRegistry
	// when one or more registrations are invalid.
	CodeRegistryValidationFailed = "REGISTRY_VALIDATION_FAILED"
)

// ─── PageRegistration ─────────────────────────────────────────────────────────

// PageRegistration describes a single page in the UI registry.
//
// Route, Module, and Title are required. At least one of Fn or ASTFn must be set.
// Prefer ASTFn for new pages — the typed AST gives compile-time guarantees that
// the legacy PageFn cannot provide.
//
// ValidateRegistry() checks all registrations at startup. A missing or invalid
// registration causes a panic before the HTTP server accepts any traffic.
type PageRegistration struct {
	// Route is the URL path served by this page, e.g. "/finance/invoices".
	// Must start with "/" and must not have a trailing slash.
	Route string

	// Module groups related pages for cache invalidation and observability.
	// Use the top-level domain name: "finance", "iam", "inventory", "dashboard".
	Module string

	// Title is the human-readable page title used in logs, traces, and the nav tree.
	Title string

	// Description is optional documentation surfaced in the registry debug endpoint.
	Description string

	// Fn is the legacy schema builder. Nil if ASTFn is set.
	// Kept for backward compatibility during migration to the typed AST.
	Fn ui.PageFn

	// ASTFn is the preferred typed AST builder. Returns ast.Node (typed as any
	// to avoid the circular import between ui and ast packages).
	// When both Fn and ASTFn are set, CompileStage uses ASTFn and ignores Fn.
	ASTFn ui.ASTPageFn
}

// validate checks all required fields and returns a list of violation strings.
func (r PageRegistration) validate() []string {
	var v []string
	if r.Route == "" || !strings.HasPrefix(r.Route, "/") {
		v = append(v, fmt.Sprintf("route %q: must start with \"/\"", r.Route))
	} else if r.Route != "/" && strings.HasSuffix(r.Route, "/") {
		v = append(v, fmt.Sprintf("route %q: trailing slash not allowed", r.Route))
	}
	if r.Module == "" {
		v = append(v, fmt.Sprintf("route %q: Module is required", r.Route))
	}
	if r.Title == "" {
		v = append(v, fmt.Sprintf("route %q: Title is required", r.Route))
	}
	if r.Fn == nil && r.ASTFn == nil {
		v = append(v, fmt.Sprintf("route %q: at least one of Fn or ASTFn must be set", r.Route))
	}
	return v
}

// ─── Registry ─────────────────────────────────────────────────────────────────

// paramRoute holds a pre-parsed parameterised route pattern (contains ":").
type paramRoute struct {
	pattern  string   // original pattern, e.g. "/finance/invoices/:id"
	segments []string // split on "/", e.g. ["finance", "invoices", ":id"]
	reg      PageRegistration
}

var (
	mu            sync.RWMutex
	registrations = map[string]PageRegistration{} // exact-match routes only
	paramRoutes   []paramRoute                     // param-pattern routes (contain ":")
	patternSet    = map[string]bool{}              // all registered patterns for de-dup
)

// RegisterPage adds a PageRegistration to the registry.
// Routes containing ":" are treated as param patterns (e.g. "/finance/invoices/:id")
// and matched via Match() at request time. Exact routes use O(1) map lookup.
// Panics on duplicate route/pattern to catch init() ordering bugs at startup.
// Call ValidateRegistry() after all init() functions have run to validate fields.
func RegisterPage(reg PageRegistration) {
	mu.Lock()
	defer mu.Unlock()
	if patternSet[reg.Route] {
		panic(fmt.Sprintf("ui/registry: duplicate registration for route %q", reg.Route))
	}
	patternSet[reg.Route] = true
	if strings.Contains(reg.Route, ":") {
		segs := strings.Split(strings.TrimPrefix(reg.Route, "/"), "/")
		paramRoutes = append(paramRoutes, paramRoute{
			pattern:  reg.Route,
			segments: segs,
			reg:      reg,
		})
	} else {
		registrations[reg.Route] = reg
	}
}

// Register is the legacy API kept for backward compatibility.
// New pages must use RegisterPage with full metadata.
//
// Deprecated: use RegisterPage.
func Register(path string, fn ui.PageFn) {
	RegisterPage(PageRegistration{
		Route:  path,
		Module: "unknown",
		Title:  path,
		Fn:     fn,
	})
}

// Match returns the PageRegistration and extracted URL params for path.
// Tries exact lookup first (O(1)), then param pattern matching (O(n) over param routes).
// Params map is always non-nil on a match; empty for exact routes with no segments.
// Returns nil, nil when no registration matches.
func Match(path string) (*PageRegistration, map[string]string) {
	mu.RLock()
	defer mu.RUnlock()

	// Exact match — O(1).
	if reg, ok := registrations[path]; ok {
		return &reg, map[string]string{}
	}

	// Param pattern matching — O(n) over param routes.
	pathSegs := strings.Split(strings.TrimPrefix(path, "/"), "/")
	for i := range paramRoutes {
		pr := &paramRoutes[i]
		if len(pr.segments) != len(pathSegs) {
			continue
		}
		params := make(map[string]string, len(pr.segments))
		matched := true
		for j, seg := range pr.segments {
			if strings.HasPrefix(seg, ":") {
				params[seg[1:]] = pathSegs[j]
			} else if seg != pathSegs[j] {
				matched = false
				break
			}
		}
		if matched {
			return &pr.reg, params
		}
	}
	return nil, nil
}

// GetRegistration returns the full PageRegistration for path, or nil if not registered.
// Supports both exact routes and param patterns (e.g. "/finance/invoices/:id").
// Use Match() directly when you also need the extracted URL params.
func GetRegistration(path string) *PageRegistration {
	reg, _ := Match(path)
	return reg
}

// Get returns the legacy PageFn for path. Returns nil if not registered or if the
// page uses ASTFn only. Kept for DevSchemaHandler backward compatibility.
func Get(path string) ui.PageFn {
	mu.RLock()
	defer mu.RUnlock()
	reg, ok := registrations[path]
	if !ok {
		return nil
	}
	return reg.Fn
}

// Paths returns all registered routes and patterns (exact + param).
func Paths() []string {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]string, 0, len(registrations)+len(paramRoutes))
	for p := range registrations {
		out = append(out, p)
	}
	for _, pr := range paramRoutes {
		out = append(out, pr.pattern)
	}
	return out
}

// Registrations returns a snapshot of all PageRegistrations, keyed by route/pattern.
func Registrations() map[string]PageRegistration {
	mu.RLock()
	defer mu.RUnlock()
	out := make(map[string]PageRegistration, len(registrations)+len(paramRoutes))
	for k, v := range registrations {
		out[k] = v
	}
	for _, pr := range paramRoutes {
		out[pr.pattern] = pr.reg
	}
	return out
}

// ValidateRegistry checks every registered PageRegistration for required fields.
// Returns *sharedErrors.BusinessError with code REGISTRY_VALIDATION_FAILED if any
// registration is invalid. Returns nil when all registrations are valid.
//
// Call this at application startup after all init() functions have run:
//
//	if err := registry.ValidateRegistry(); err != nil {
//	    panic(err)
//	}
func ValidateRegistry() error {
	mu.RLock()
	defer mu.RUnlock()

	var violations []string
	for _, reg := range registrations {
		violations = append(violations, reg.validate()...)
	}
	for _, pr := range paramRoutes {
		violations = append(violations, pr.reg.validate()...)
	}

	if len(violations) == 0 {
		return nil
	}

	details := make(map[string]any, len(violations))
	for i, v := range violations {
		details[fmt.Sprintf("violation_%d", i+1)] = v
	}

	err := sharedErrors.NewBusinessError(
		CodeRegistryValidationFailed,
		fmt.Sprintf("UI page registry has %d invalid registration(s)", len(violations)),
	).
		WithHTTPStatus(http.StatusInternalServerError).
		WithCategory(sharedErrors.CategorySystem).
		WithDetail("violation_count", len(violations))

	for k, v := range details {
		err = err.WithDetail(k, v)
	}

	return err
}
