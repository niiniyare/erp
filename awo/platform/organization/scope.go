package organization

// VisibilityMode controls which organization nodes are included when an
// application service resolves the organizations visible to a user.
//
// # Application layer only
//
// VisibilityMode is NOT a database isolation mechanism. Organization scope
// is evaluated entirely in Go by OrganizationService.ResolveScope(). The
// resulting []uuid.UUID is passed as an explicit IN predicate by the calling
// application service. RLS is never involved.
//
// Tenant isolation (RLS) and organization scope are orthogonal:
//   - RLS guarantees: data for tenant A never appears in tenant B's queries.
//   - VisibilityMode governs: within tenant A, which org nodes can user U see?
type VisibilityMode int

const (
	// VisibilityCurrent: only the user's active organization.
	// Typical use: branch staff entering data for their own branch.
	VisibilityCurrent VisibilityMode = iota

	// VisibilityDescendants: active org + all nodes below it in the tree.
	// Uses materialized path prefix match — O(1) regardless of tree depth.
	// Typical use: regional manager viewing all branches under their region.
	VisibilityDescendants

	// VisibilityAncestors: active org + all nodes above it up to the root.
	// Typical use: breadcrumb navigation; escalation chains.
	VisibilityAncestors

	// VisibilityEntireTenant: all org nodes in the tenant, no org filter.
	// Typical use: tenant.admin, platform.admin, HQ Finance Manager.
	VisibilityEntireTenant

	// VisibilityExplicit: a caller-supplied fixed set of org IDs.
	// Typical use: internal auditor assigned to a non-contiguous set of orgs.
	// Requires ViewerContext.ExplicitOrganizationIDs to be populated.
	VisibilityExplicit

	// VisibilityCustom: delegates to a ScopeResolver implementation.
	// Typical use: matrix organizations; project-team cross-org access.
	// Requires ViewerContext.ScopeResolver to be non-nil.
	VisibilityCustom
)

// String returns a stable, human-readable name for the mode constant.
func (m VisibilityMode) String() string {
	switch m {
	case VisibilityCurrent:
		return "current"
	case VisibilityDescendants:
		return "descendants"
	case VisibilityAncestors:
		return "ancestors"
	case VisibilityEntireTenant:
		return "entire_tenant"
	case VisibilityExplicit:
		return "explicit"
	case VisibilityCustom:
		return "custom"
	default:
		return "unknown"
	}
}
