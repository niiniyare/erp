// Package domain contains pure IAM domain models and business rules.
// No external framework dependencies — only stdlib and uuid.
package domain

import "time"

// =============================================================================
// Isolation Model — read this first
// =============================================================================
//
// Awo ERP enforces tenant isolation at two independent layers.  Every piece of
// IAM code must be clear about which layer it is operating in.
//

// Layer 1 — PostgreSQL Row-Level Security  (TenantID)

// Every table has a tenant_id column and an RLS policy of the form:
// USING (tenant_id = ccurrent_tenant_id()::uuid)

// The service layer sets this GUC at the start of every transaction:
// SET LOCAL awo.tenant_id = '<session.TenantID>'

// This is the HARD boundary.  Even a bug in Go code cannot return
// rows from another tenant; the DB will silently filter them out.

//

// Layer 2 — Application-layer Entity Scope  (EntityID)

// Within a single tenant there are entities (branches, departments,
// subsidiaries …) organised as an ltree hierarchy.  A user's
// EntityScope (see session.go) restricts which entity rows they see:
// EntityScopeEntity  → own entity only
// EntityScopeSubtree → own entity + all descendants
// EntityScopeAll     → all entities in the tenant

// This is an APPLICATION boundary.  Service methods build WHERE
// clauses or ltree path filters based on EntityScope before issuing
// queries.  It is intentionally softer — a manager can be granted
// atemporary cross-entity view without any DB schema change.

// =============================================================================
// Actor Types
// =============================================================================

// ActorType identifies the class of principal making a request.
// Using a typed string prevents raw string literals from sneaking into
// positions that expect an actor class and makes exhaustive switches detectable
// by the compiler.
type ActorType string

const (
	// ActorPlatform is an Awo system administrator with cross-tenant visibility.
	// Platform actors operate in the reserved DomainPlatform Casbin domain and
	// are never bound to a single tenant's RLS context.
	ActorPlatform ActorType = "platform"

	// ActorTenant is a regular internal user (employee, manager, accountant …)
	// belonging to a specific tenant.  The tenant's RLS policy governs every
	// DB query issued within their session.  This is the most common actor type.
	ActorTenant ActorType = "tenant"

	// ActorPortal is an external-facing user (customer, supplier, partner …)
	// who authenticates through a tenant-branded portal.  Portal actors are
	// further scoped to a PrincipalID — the contact or party record they
	// represent — and operate in the "<tenantID>:portal" Casbin domain to
	// prevent portal policies from colliding with internal tenant policies.
	ActorPortal ActorType = "portal"

	// ActorAPI is a machine principal (service account, webhook consumer,
	// third-party integration client) identified by a client_id / secret pair
	// rather than a human credential.  API actors use the "<tenantID>:api"
	// Casbin domain.
	//
	// NOTE: API key sessions are not yet wired into ResolvedSession or the
	// login flow.  The constants, helpers, and Casbin domain are defined here
	// for forward-compatibility.  See TODO in session.go → ToPrincipal().
	ActorAPI ActorType = "api"
)

// ActorTypeFromUserType translates the persisted UserType string stored on the
// users table (an ALL-CAPS enum: "INTERNAL", "SYSADMIN", "CUSTOMER" …) into
// the lowercase ActorType used throughout the authorization layer.
//
// This is the SINGLE canonical mapping point.  All other code that needs to
// reason about actor class must call this function rather than comparing
// UserType strings directly.
//
// Mapping table:
//
//	"SYSADMIN" | "PLATFORM"          → ActorPlatform
//	"PORTAL"   | "CUSTOMER"          → ActorPortal
//	"API"      | "SERVICE"           → ActorAPI
//	everything else (incl. "INTERNAL", "EMPLOYEE") → ActorTenant
func ActorTypeFromUserType(userType string) ActorType {
	switch userType {
	case "SYSADMIN", "PLATFORM":
		return ActorPlatform
	case "PORTAL", "CUSTOMER":
		return ActorPortal
	case "API", "SERVICE":
		return ActorAPI
	default:
		return ActorTenant
	}
}

// =============================================================================
// Subject / Domain Helpers
// =============================================================================
//
// Casbin stores authorization rules as (sub, dom, obj, act, eft) tuples.
// The Subject and Domain strings below are the canonical formats for each
// actor class.  Always construct them through these helpers — never build the
// strings inline — so that a format change only requires editing one place.
//
// Subject format: "<actor-class>:<id>"
// PlatformSubject("uuid") → "platform:uuid"
// TenantSubject("uuid")   → "tenant:uuid"
// PortalSubject("uuid")   → "portal:uuid"
// APISubject("client-id") → "api:client-id"
//
// Domain format:
// PlatformDomain()          → "_platform_"   (constant, no variable part)
// TenantDomain("tenantID")  → "tenantID"      (bare UUID — most common case)
// PortalDomain("tenantID")  → "tenantID:portal"
// APIDomain("tenantID")     → "tenantID:api"
//
// The ":portal" and ":api" suffixes prevent portal/API policies from
// accidentally matching internal tenant rules that use the same role names.

// DomainPlatform is the reserved Casbin domain for cross-tenant platform
// policies.  It is a constant because there is exactly one platform domain.
const DomainPlatform = "_platform_"

// PlatformDomain returns DomainPlatform as a function so callers that build
// domain strings through a table of functions can include the platform case
// without special-casing the constant.
func PlatformDomain() string { return DomainPlatform }

// Subject helpers — build the Casbin "sub" field.

func PlatformSubject(userID string) string { return "platform:" + userID }
func TenantSubject(userID string) string   { return "tenant:" + userID }
func PortalSubject(userID string) string   { return "portal:" + userID }
func APISubject(clientID string) string    { return "api:" + clientID }

// Domain helpers — build the Casbin "dom" field.

func TenantDomain(tenantID string) string { return tenantID }
func PortalDomain(tenantID string) string { return tenantID + ":portal" }
func APIDomain(tenantID string) string    { return tenantID + ":api" }

// =============================================================================
// Value Objects — Authorization Primitives
// =============================================================================

// Principal is the subject-domain pair set in Fiber context by the authn
// middleware and consumed by the Casbin enforcement middleware.
//
// Principal is a value object: immutable once constructed, with no identity of
// its own, and fully described by its two fields.  Construct it via
// ResolvedSession.ToPrincipal() rather than directly.
type Principal struct {
	// Subject identifies who is making the request.
	// Format: "<actor-class>:<id>" — e.g. "tenant:550e8400-…"
	Subject string

	// Domain scopes the Casbin policy lookup.
	// Format: tenant UUID, "_platform_", "<tenantID>:portal", or "<tenantID>:api".
	Domain string
}

// LocalsKeyPrincipal is the Fiber Locals key for the authenticated Principal.
//
// Usage in authorization middleware:
//
//	p := c.Locals(domain.LocalsKeyPrincipal).(domain.Principal)
const LocalsKeyPrincipal = "authz_principal"

// Request is a single authorization check passed to the Casbin enforcer.
// It mirrors the (sub, dom, obj, act) tuple of the Casbin request definition.
//
// Object follows a slash-separated resource path convention:
//
//	"invoice/123"  — a specific resource instance
//	"invoice/*"    — all instances of a resource type
//	"report/gl/*"  — all GL reports (hierarchical wildcard)
//
// Action is a lowercase verb matched with keyMatch (glob), so "*" matches any
// action:
//
//	"read" | "create" | "update" | "delete" | "approve" | "*"
type Request struct {
	Subject string // who   — e.g. "tenant:user_uuid"
	Domain  string // scope — e.g. tenantID or "_platform_"
	Object  string // what  — e.g. "invoice/123" or "invoice/*"
	Action  string // how   — e.g. "read", "create", "delete", "*"
}

// Policy is a Casbin p-rule (permission row) used when seeding or inspecting
// the policy store programmatically.
//
// Effect must be "allow" or "deny".  Due to the deny-override model in
// CasbinModel, a single matching deny policy cancels all allows for the same
// (sub, dom, obj, act) combination — use deny rules sparingly and only when
// you need to carve exceptions out of a broad allow.
type Policy struct {
	Subject string
	Domain  string
	Object  string
	Action  string
	Effect  string // "allow" | "deny"
}

// =============================================================================
// RoleAssignment — Domain Entity
// =============================================================================

// RoleAssignment is a metadata record that annotates a Casbin g-rule
// (user-to-role grouping) with audit and lifecycle fields.
//
// The Casbin g-rule is the source of truth for enforcement; this entity exists
// for audit trails, expiry management, and delegation tracking.
// It maps to the role_assignments table.
//
// Relationship to TenantID / EntityID
//
//	TenantID is denormalised here for fast lookup queries but adds no extra
//	enforcement boundary — the Casbin Domain already encodes the tenant.
//
//	Entity-level role scoping (e.g. "accountant for branch A only") is an
//	application-layer concern handled by service methods that inspect the
//	caller's EntityScope before calling AssignRole.  It is NOT enforced by
//	Casbin policies directly.
type RoleAssignment struct {
	ID          string
	Subject     string     // Casbin sub — e.g. "tenant:uuid"
	Role        string     // Casbin role name — e.g. "accountant"
	Domain      string     // Casbin dom — e.g. tenant UUID or "_platform_"
	TenantID    string     // denormalised for query convenience; not an extra enforcement layer
	AssignedBy  string     // subject who performed the assignment; empty → system assignment
	DelegatedBy string     // subject who delegated authority, if any; empty → direct assignment
	ExpiresAt   *time.Time // nil → no expiry
	IsActive    bool
	CreatedAt   time.Time
}

// IsExpired reports whether the assignment has passed its expiry time.
// Assignments with a nil ExpiresAt never expire.
func (r *RoleAssignment) IsExpired() bool {
	return r.ExpiresAt != nil && time.Now().After(*r.ExpiresAt)
}

// IsEffective reports whether the assignment is both active and not yet
// expired.  Use this rather than checking IsActive and IsExpired separately.
func (r *RoleAssignment) IsEffective() bool {
	return r.IsActive && !r.IsExpired()
}

// =============================================================================
// Functional Options — AssignRole
// =============================================================================

// AssignOpt is a functional option applied when constructing an AssignOpts.
// Callers build a variadic list; the service unpacks them with ApplyAssignOpts.
type AssignOpt func(*AssignOpts)

// AssignOpts holds all optional parameters for a role assignment operation.
// Zero values are safe defaults: no expiry, no audit context.
type AssignOpts struct {
	ExpiresAt   *time.Time // nil → assignment never expires
	AssignedBy  string     // empty → system / programmatic assignment
	DelegatedBy string     // empty → direct assignment, no delegation chain
}

// WithExpiry sets a point-in-time expiry on the role assignment.  The Casbin
// g-rule will remain in the policy store but IsEffective() returns false once
// this instant passes.  A background job should periodically prune expired rows.
func WithExpiry(t time.Time) AssignOpt {
	return func(o *AssignOpts) { o.ExpiresAt = &t }
}

// WithAssignedBy records the subject (e.g. "tenant:uuid") who performed this
// assignment.  Used for audit trails and self-assignment prevention checks.
func WithAssignedBy(sub string) AssignOpt {
	return func(o *AssignOpts) { o.AssignedBy = sub }
}

// WithDelegatedBy records the subject who delegated authority to the assigner.
// Set this when a manager grants a role on behalf of a higher-ranking principal
// so that the delegation chain is traceable in audit logs.
func WithDelegatedBy(sub string) AssignOpt {
	return func(o *AssignOpts) { o.DelegatedBy = sub }
}

// ApplyAssignOpts applies all provided options and returns a populated
// AssignOpts struct.  Returning the struct (rather than multiple *string return
// values) lets callers access only the fields they need without nil-checking
// every value, and keeps the signature stable as new options are added.
func ApplyAssignOpts(opts []AssignOpt) AssignOpts {
	ao := AssignOpts{}
	for _, o := range opts {
		o(&ao)
	}
	return ao
}

// =============================================================================
// Casbin Model
// =============================================================================

// CasbinModel is the CONF-format Casbin model used by every enforcer in this
// bounded context.  It is embedded here so the model definition travels with
// the domain package and is never silently out of sync with the policy store.
//
// #Design decisions
//
// 1. Deny-override effect
// "some(allow) && !some(deny)" — one explicit deny beats all allows.
// Appropriate for an ERP where sensitive resources (payroll, bank accounts)
// must be lockable per-role without revoking every individual allow rule.
//
// 2. Domain-scoped RBAC  (g = _, _, _)
// Roles are tenant-local.  An "accountant" role in tenant A cannot match
// apolicy in tenant B even if both tenants use the same role name.
// This is the second line of defence after PostgreSQL RLS.
//
// 3. keyMatch2 on obj
// Enables path-parameter wildcards on the resource object field so that
// apolicy for "invoice/:id" matches a request for "invoice/abc-123".
// Resource hierarchies are modelled as slash-separated paths.
//
// 4. keyMatch on act  — NOT keyMatch2
// Actions are simple verb tokens, not path segments.  keyMatch (glob)
// correctly matches the wildcard action "*" against any concrete verb
// such as "read" or "delete".
//
// IMPORTANT: keyMatch2 was previously used here by mistake.  keyMatch2
// uses ":param" path-segment matching and does NOT treat "*" as a glob
// wildcard, so a policy action of "*" would never match "read".  This
// would silently deny all actions covered by wildcard policies.
const CasbinModel = `
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act, eft

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && keyMatch(r.act, p.act)
`

// Odl models TODO: Review
// const CasbinModel = `
// [request_definition]
// r= sub, dom, obj, act
//
// [policy_definition]
// p= sub, dom, obj, act, eft
//
// [role_definition]
// g= _, _, _
//
// [policy_effect]
// e= some(where (p.eft == allow)) && !some(where (p.eft == deny))
//
// [matchers]
// m= g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && keyMatch2(r.act, p.act)
// `
//
