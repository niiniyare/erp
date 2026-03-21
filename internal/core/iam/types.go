package iam

import "time"

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

// Subject helpers build the "sub" field stored in Casbin rules.
func PlatformSubject(userID string) string { return "platform:" + userID }
func TenantSubject(userID string) string   { return "tenant:" + userID }
func PortalSubject(userID string) string   { return "portal:" + userID }
func APISubject(clientID string) string    { return "api:" + clientID }

// Domain helpers build the "dom" field stored in Casbin rules.
func TenantDomain(id string) string { return id }
func PortalDomain(id string) string { return id + ":portal" }
func APIDomain(id string) string    { return id + ":api" }

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

// AssignOpt is a functional option for AssignRole.
type AssignOpt func(*assignOpts)

type assignOpts struct {
	expiresAt   *time.Time
	assignedBy  string
	delegatedBy string
}

// WithExpiry sets a temporal expiry on the role assignment.
func WithExpiry(t time.Time) AssignOpt {
	return func(o *assignOpts) { o.expiresAt = &t }
}

// WithAssignedBy records who performed the assignment.
func WithAssignedBy(sub string) AssignOpt {
	return func(o *assignOpts) { o.assignedBy = sub }
}

// WithDelegatedBy records who delegated authority to the assigner.
func WithDelegatedBy(sub string) AssignOpt {
	return func(o *assignOpts) { o.delegatedBy = sub }
}

// Principal is stored in Fiber context by the authn middleware.
type Principal struct {
	Subject string
	Domain  string
}

// LocalsKeyPrincipal is the Fiber Locals key for the authenticated Principal.
//
//	principal := c.Locals(authz.LocalsKeyPrincipal).(authz.Principal)
const LocalsKeyPrincipal = "authz_principal"
