package domain

import "time"

// ─── Actor Types ──────────────────────────────────────────────────────────────

// ActorType identifies the class of principal making a request.
type ActorType string

const (
	ActorPlatform ActorType = "platform"
	ActorTenant   ActorType = "tenant"
	ActorPortal   ActorType = "portal"
	ActorAPI      ActorType = "api"
)

// DomainPlatform is the reserved Casbin domain for platform-level policies.
const DomainPlatform = "_platform_"

// ─── Subject / Domain Helpers ────────────────────────────────────────────────

// Subject helpers build the "sub" field stored in Casbin rules.
func PlatformSubject(userID string) string { return "platform:" + userID }
func TenantSubject(userID string) string   { return "tenant:" + userID }
func PortalSubject(userID string) string   { return "portal:" + userID }
func APISubject(clientID string) string    { return "api:" + clientID }

// Domain helpers build the "dom" field stored in Casbin rules.
func TenantDomain(id string) string { return id }
func PortalDomain(id string) string { return id + ":portal" }
func APIDomain(id string) string    { return id + ":api" }

// ─── Value Objects ────────────────────────────────────────────────────────────

// Principal is stored in Fiber context by the authn middleware and consumed
// by the Casbin enforcement middleware.
type Principal struct {
	Subject string
	Domain  string
}

// LocalsKeyPrincipal is the Fiber Locals key for the authenticated Principal.
const LocalsKeyPrincipal = "authz_principal"

// Request is a single authorization check.
type Request struct {
	Subject string // who   — e.g. "tenant:user_uuid"
	Domain  string // scope — e.g. "{tenantID}" or "_platform_"
	Object  string // what  — e.g. "invoice/123" or "invoice/*"
	Action  string // how   — e.g. "read", "create", "delete", "*"
}

// Policy is a Casbin p-rule (permission row).
type Policy struct {
	Subject string
	Domain  string
	Object  string
	Action  string
	Effect  string // "allow" | "deny"
}

// RoleAssignment is a row from the role_assignments metadata table.
type RoleAssignment struct {
	ID          string
	Subject     string
	Role        string
	Domain      string
	TenantID    string
	AssignedBy  string
	DelegatedBy string
	ExpiresAt   *time.Time
	IsActive    bool
	CreatedAt   time.Time
}

// ─── Functional Options for AssignRole ───────────────────────────────────────

// AssignOpt is a functional option for AssignRole.
type AssignOpt func(*AssignOpts)

// AssignOpts holds the optional parameters for a role assignment.
type AssignOpts struct {
	ExpiresAt   *time.Time
	AssignedBy  string
	DelegatedBy string
}

// WithExpiry sets a temporal expiry on the role assignment.
func WithExpiry(t time.Time) AssignOpt {
	return func(o *AssignOpts) { o.ExpiresAt = &t }
}

// WithAssignedBy records who performed the assignment.
func WithAssignedBy(sub string) AssignOpt {
	return func(o *AssignOpts) { o.AssignedBy = sub }
}

// WithDelegatedBy records who delegated authority to the assigner.
func WithDelegatedBy(sub string) AssignOpt {
	return func(o *AssignOpts) { o.DelegatedBy = sub }
}

// ApplyAssignOpts applies functional options and returns the unpacked values
// for use by the service layer.
func ApplyAssignOpts(opts []AssignOpt) (expiresAt *time.Time, assignedBy, delegatedBy *string) {
	ao := &AssignOpts{}
	for _, o := range opts {
		o(ao)
	}
	var ab, db *string
	if ao.AssignedBy != "" {
		s := ao.AssignedBy
		ab = &s
	}
	if ao.DelegatedBy != "" {
		s := ao.DelegatedBy
		db = &s
	}
	return ao.ExpiresAt, ab, db
}

// ─── Casbin Model ────────────────────────────────────────────────────────────

// CasbinModel is the Casbin CONF model used by every enforcer in this bounded context.
//
// Key properties:
//   - deny-override: one explicit deny beats all allows
//   - role hierarchy: g(user, role, domain) — domain-scoped RBAC
//   - keyMatch2 on obj and act allows wildcard patterns like "invoice/*"
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
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && keyMatch2(r.act, p.act)
`
