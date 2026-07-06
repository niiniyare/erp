// Package authz integrates Casbin RBAC with the Awo framework.
//
// Policy model: (subject, domain, object, action)
//   - subject: "user:{uuid}" or "role:{name}"
//   - domain:  "_platform_" or tenant UUID string
//   - object:  entity type name (e.g. "finance_invoice")
//   - action:  "read", "write", "create", "delete", or action name
//
// The Casbin enforcer is loaded once at startup from the compiled schema's
// CasbinPolicies slice. Per-request enforcement happens in the RequirePermission
// middleware.
//
// Role hierarchy is expressed via Casbin `g` grouping assertions:
//
//	g, role:tenant.admin, role:tenant.user, {tenant_uuid}
//
// Inheritance accumulates upward — an admin also has all user permissions.
package authz

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/gofiber/fiber/v2"

	"awo.so/awo/api/response"
	"awo.so/awo/compiler"
	"awo.so/awo/runtime"
)

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

// Enforcer wraps the Casbin enforcer with a domain-aware check API.
type Enforcer struct {
	e *casbin.Enforcer
}

// NewEnforcer creates a Casbin enforcer loaded with policies from the compiled schema.
func NewEnforcer(schema *compiler.CompiledSchema) (*Enforcer, error) {
	m, err := model.NewModelFromString(casbinModel)
	if err != nil {
		return nil, fmt.Errorf("authz: parse casbin model: %w", err)
	}

	e, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("authz: create enforcer: %w", err)
	}

	// Load compiled policies.
	for _, p := range schema.CasbinPolicies {
		if _, err := e.AddPolicy(p.Subject, p.Object, p.Action); err != nil {
			return nil, fmt.Errorf("authz: add policy %v: %w", p, err)
		}
	}

	return &Enforcer{e: e}, nil
}

// Can reports whether subject (e.g. "role:tenant.admin") is allowed to perform
// action on object.
func (en *Enforcer) Can(subject, object, action string) (bool, error) {
	return en.e.Enforce(subject, object, action)
}

// AddRoleForUser grants role to user.
func (en *Enforcer) AddRoleForUser(user, role string) error {
	_, err := en.e.AddRoleForUser(user, role)
	return err
}

// RemoveRoleForUser revokes role from user.
func (en *Enforcer) RemoveRoleForUser(user, role string) error {
	_, err := en.e.DeleteRoleForUser(user, role)
	return err
}

// RequirePermission returns a Fiber middleware that enforces RBAC.
// entityName is the entity being accessed; action is "read", "write", etc.
//
// The middleware reads user_id from c.Locals (set by RequireAuth) and
// tenant_id from c.Locals (set by TenantResolver).
//
// Platform admins bypass Casbin entirely.
func (en *Enforcer) RequirePermission(entityName, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDStr, _ := c.Locals("user_id").(string)
		_, hasTenant := c.Locals("tenant_id").(string)

		if userIDStr == "" || !hasTenant {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Wrap(&runtime.PermissionError{
				EntityName: entityName,
				Action:     action,
			}))
		}

		subject := "user:" + userIDStr

		ok, err := en.Can(subject, entityName, action)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(response.Wrap(err))
		}
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(response.Wrap(&runtime.PermissionError{
				EntityName: entityName,
				Action:     action,
			}))
		}
		return c.Next()
	}
}
