package auth

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"awo.so/awo/compiler"
	casbinv2 "github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

// casbinModel is the Casbin RBAC model text used by the default evaluator.
// It maps permission identifiers (p.sub) to entity+action pairs, and roles
// to permission identifiers via g (role hierarchy / inheritance) assertions.
//
// Authorization flow in Casbin:
//   - Phase 1 CapabilityGrants: p, {permission}, {entity}, {action}
//     e.g. p, finance.invoice.create, finance_invoice, create
//   - Phase 2 role-to-permission: g, {role}, {permission}
//     e.g. g, role:finance.accounts_payable, finance.invoice.create
//   - Enforcer check: does g(role, permission) && p(permission, entity, action)?
//
// The matcher g(r.sub, p.sub) resolves role inheritance chains so that a role
// granted a parent permission automatically satisfies child permission checks.
const casbinModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

// RolePermission maps a role name to a permission identifier it grants.
// The IAM module loads these from the iam_role_permissions table at startup
// and passes them to [NewCasbinEvaluator].
type RolePermission struct {
	// Role is the role name. Format: "role:{domain}.{name}".
	// e.g. "role:finance.accounts_payable", "role:tenant.admin".
	Role string

	// Permission is the permission identifier granted to this role.
	// Format: "{module}.{entity}.{operation}".
	// e.g. "finance.invoice.create", "finance.invoice.submit".
	Permission string
}

// CasbinEvaluator is the default [PolicyEvaluator] implementation. It uses
// Casbin v2 with a two-phase policy loading strategy:
//
//   - Phase 1: [compiler.CapabilityGrant] values are loaded as Casbin p
//     assertions, binding permission identifiers to entity+action pairs.
//   - Phase 2: [RolePermission] values from the IAM module are loaded as
//     Casbin g assertions, binding roles to permission identifiers.
//
// The enforcer is held in memory and safe for concurrent reads. Policy reload
// (e.g. after a role-to-permission change) uses a write-locked swap via
// [CasbinEvaluator.Reload].
type CasbinEvaluator struct {
	mu       sync.RWMutex
	enforcer *casbinv2.Enforcer
}

// NewCasbinEvaluator constructs a CasbinEvaluator and loads the initial
// policy from grants (Phase 1) and rolePerms (Phase 2).
//
// grants comes from [compiler.CompiledSchema.CapabilityGrants].
// rolePerms comes from the IAM module's iam_role_permissions table.
//
// Returns an error only if the Casbin model fails to parse (should never
// happen with the embedded model constant) or policy loading fails.
func NewCasbinEvaluator(grants []compiler.CapabilityGrant, rolePerms []RolePermission) (*CasbinEvaluator, error) {
	enforcer, err := buildEnforcer(grants, rolePerms)
	if err != nil {
		return nil, fmt.Errorf("auth: build casbin enforcer: %w", err)
	}
	return &CasbinEvaluator{enforcer: enforcer}, nil
}

// Ensure CasbinEvaluator implements PolicyEvaluator at compile time.
var _ PolicyEvaluator = (*CasbinEvaluator)(nil)

// CanPerform returns true if any role held by viewer grants permission to
// perform action on object.
//
// The check iterates viewer.Roles() and queries the Casbin enforcer for each
// role. The first matching role short-circuits and returns true. This is O(n)
// in the number of roles, which in practice is very small (< 10).
//
// Platform admins are NOT handled here — the middleware short-circuits before
// consulting CanPerform when viewer.IsPlatformAdmin() is true.
func (e *CasbinEvaluator) CanPerform(ctx context.Context, viewer ViewerContext, object string, action string) (bool, error) {
	e.mu.RLock()
	enf := e.enforcer
	e.mu.RUnlock()

	for _, role := range viewer.Roles() {
		ok, err := enf.Enforce(role, object, action)
		if err != nil {
			return false, fmt.Errorf("auth: casbin enforce %q on %s.%s: %w", role, object, action, err)
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// Reload atomically replaces the enforcer with a new one loaded from the
// provided grants and rolePerms. Use after a role-to-permission change in the
// IAM module to apply new permissions without restarting the process.
//
// The swap is write-locked; in-flight CanPerform calls using the old enforcer
// complete before the swap takes effect.
func (e *CasbinEvaluator) Reload(grants []compiler.CapabilityGrant, rolePerms []RolePermission) error {
	enforcer, err := buildEnforcer(grants, rolePerms)
	if err != nil {
		return fmt.Errorf("auth: reload casbin enforcer: %w", err)
	}
	e.mu.Lock()
	e.enforcer = enforcer
	e.mu.Unlock()
	return nil
}

// buildEnforcer constructs a new Casbin Enforcer from the embedded model
// and loads Phase 1 (CapabilityGrants) and Phase 2 (RolePermission) policy data.
func buildEnforcer(grants []compiler.CapabilityGrant, rolePerms []RolePermission) (*casbinv2.Enforcer, error) {
	m, err := model.NewModelFromString(strings.TrimSpace(casbinModel))
	if err != nil {
		return nil, fmt.Errorf("parse casbin model: %w", err)
	}

	// Use an empty adapter; we load all policies programmatically.
	adapter := &memoryAdapter{}
	enforcer, err := casbinv2.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("new casbin enforcer: %w", err)
	}

	// Phase 1: load CapabilityGrants as p assertions.
	// p, {permission}, {entity}, {action}
	for _, g := range grants {
		if _, err := enforcer.AddPolicy(g.Permission, g.Entity, g.Action); err != nil {
			return nil, fmt.Errorf("add capability grant p(%s, %s, %s): %w", g.Permission, g.Entity, g.Action, err)
		}
	}

	// Phase 2: load role-to-permission bindings as g assertions.
	// g, {role}, {permission}
	for _, rp := range rolePerms {
		if _, err := enforcer.AddRoleForUser(rp.Role, rp.Permission); err != nil {
			return nil, fmt.Errorf("add role permission g(%s, %s): %w", rp.Role, rp.Permission, err)
		}
	}

	return enforcer, nil
}

// memoryAdapter is a no-op Casbin persist.Adapter. We load all policies
// programmatically via enforcer.AddPolicy / AddRoleForUser; the adapter is
// never used for load/save cycles. This avoids the need for a file or DB
// adapter while still satisfying the Casbin Enforcer constructor.
type memoryAdapter struct{}

var _ persist.Adapter = (*memoryAdapter)(nil)

func (*memoryAdapter) LoadPolicy(_ model.Model) error  { return nil }
func (*memoryAdapter) SavePolicy(_ model.Model) error  { return nil }
func (*memoryAdapter) AddPolicy(_ string, _ string, _ []string) error { return nil }
func (*memoryAdapter) RemovePolicy(_ string, _ string, _ []string) error { return nil }
func (*memoryAdapter) RemoveFilteredPolicy(_ string, _ string, _ int, _ ...string) error {
	return nil
}
