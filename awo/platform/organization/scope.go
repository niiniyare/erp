package organization

import (
	"context"

	"github.com/google/uuid"
)

// OrganizationScope controls which organization nodes are visible to a query.
// Scope is orthogonal to TenantContext — TenantContext isolates across tenants;
// OrganizationScope filters within a single tenant's org tree.
type OrganizationScope int

const (
	// ScopeCurrent restricts results to the caller's own organization node only.
	ScopeCurrent OrganizationScope = iota

	// ScopeDescendants includes the caller's node and all nodes below it in the
	// tree. Uses the materialized path for O(1) path-prefix filtering.
	ScopeDescendants

	// ScopeAncestors includes the caller's node and all nodes above it up to the
	// root. Useful for breadcrumb / hierarchy display.
	ScopeAncestors

	// ScopeExplicit restricts to a caller-supplied set of organization IDs.
	// Requires OrganizationIDs to be populated on the scope context value.
	ScopeExplicit

	// ScopeEntireTenant bypasses organization filtering — all nodes in the tenant
	// are visible. Requires role:tenant.admin or role:platform-admin.
	ScopeEntireTenant

	// ScopeCustom delegates to an OrganizationResolver implementation. Used by
	// platform modules that need complex visibility logic (e.g. matrix orgs).
	ScopeCustom
)

// String returns a human-readable name for the scope constant.
func (s OrganizationScope) String() string {
	switch s {
	case ScopeCurrent:
		return "current"
	case ScopeDescendants:
		return "descendants"
	case ScopeAncestors:
		return "ancestors"
	case ScopeExplicit:
		return "explicit"
	case ScopeEntireTenant:
		return "entire_tenant"
	case ScopeCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// ScopeContext carries an OrganizationScope and optional explicit IDs through
// a request context. Use WithScope / ScopeFromContext to propagate.
type ScopeContext struct {
	Scope           OrganizationScope
	OrganizationID  uuid.UUID   // the caller's current org node
	OrganizationIDs []uuid.UUID // populated when Scope == ScopeExplicit
	Resolver        OrganizationResolver // populated when Scope == ScopeCustom
}

type scopeKey struct{}

// WithScope attaches an OrganizationScope to ctx.
func WithScope(ctx context.Context, sc ScopeContext) context.Context {
	return context.WithValue(ctx, scopeKey{}, sc)
}

// ScopeFromContext retrieves the ScopeContext from ctx.
// Returns ScopeEntireTenant with zero OrganizationID when absent — callers with
// no scope context see all org nodes (subject to RBAC, not org filtering).
func ScopeFromContext(ctx context.Context) ScopeContext {
	if v, ok := ctx.Value(scopeKey{}).(ScopeContext); ok {
		return v
	}
	return ScopeContext{Scope: ScopeEntireTenant}
}

// OrganizationResolver is implemented by callers that need custom visibility
// logic. It is invoked when ScopeContext.Scope == ScopeCustom.
type OrganizationResolver interface {
	// Resolve returns the set of organization IDs visible for the given tenant
	// and caller organization. Returning nil means no org filter applied.
	Resolve(ctx context.Context, tenantID, callerOrgID uuid.UUID) ([]uuid.UUID, error)
}
