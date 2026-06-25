// Package framework provides the bridge between the app's auth/session layer
// and the awo.so/framework definition.ViewerContext interface.
package framework

import (
	"github.com/gofiber/fiber/v2"

	"awo.so/framework/api"
	"awo.so/framework/definition"
)

// sessionViewer implements definition.ViewerContext from a populated Fiber session.
type sessionViewer struct {
	actorID  string
	tenantID string
	roles    map[string]bool
	isSystem bool
}

func (v *sessionViewer) ActorID() string       { return v.actorID }
func (v *sessionViewer) TenantID() string      { return v.tenantID }
func (v *sessionViewer) IsSystem() bool        { return v.isSystem }
func (v *sessionViewer) HasRole(r string) bool { return v.roles[r] }

var _ definition.ViewerContext = (*sessionViewer)(nil)

// ViewerFromFiber builds the ViewerFromCtx function used by generic handlers.
// It reads session locals populated by the authentication middleware.
//
// Expected Fiber locals (set by auth middleware before this runs):
//   - "actor_id"   string  — authenticated user UUID
//   - "tenant_id"  string  — tenant UUID (from validated session)
//   - "roles"      []string — role names held by the actor
//   - "is_system"  bool    — true for machine/service tokens
//
// Falls back to dev stub when locals are absent (development only).
func ViewerFromFiber() api.ViewerFromCtx {
	return func(c *fiber.Ctx) (definition.ViewerContext, error) {
		actorID, _ := c.Locals("actor_id").(string)
		tenantID, _ := c.Locals("tenant_id").(string)

		// If auth middleware hasn't run (e.g. dev server without auth), fall back
		// to the tenant slug from X-Awo-Tenant header.
		if tenantID == "" {
			tenantID = c.Get("X-Awo-Tenant")
		}
		if actorID == "" {
			actorID = "anonymous"
		}

		roleList, _ := c.Locals("roles").([]string)
		roles := make(map[string]bool, len(roleList))
		for _, r := range roleList {
			roles[r] = true
		}

		isSystem, _ := c.Locals("is_system").(bool)

		return &sessionViewer{
			actorID:  actorID,
			tenantID: tenantID,
			roles:    roles,
			isSystem: isSystem,
		}, nil
	}
}
