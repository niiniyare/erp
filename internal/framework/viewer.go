// Package framework provides the bridge between the app's auth/session layer
// and the awo.so/framework definition.ViewerContext interface.
package framework

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/framework/api"
	"awo.so/framework/definition"
	"awo.so/framework/org"
)

// sessionViewer implements definition.ViewerContext from a populated Fiber session.
type sessionViewer struct {
	actorID   string
	tenantID  string
	orgUnitID uuid.UUID // parsed from "org_unit_id" local; uuid.Nil = tenant-wide
	roles     map[string]bool
	isSystem  bool
}

func (v *sessionViewer) ActorID() string       { return v.actorID }
func (v *sessionViewer) TenantID() string      { return v.tenantID }
func (v *sessionViewer) OrgUnitID() uuid.UUID  { return v.orgUnitID }
func (v *sessionViewer) IsSystem() bool        { return v.isSystem }
func (v *sessionViewer) HasRole(r string) bool { return v.roles[r] }

// OrgScope builds an org.Scope from the viewer's parsed IDs.
func (v *sessionViewer) OrgScope() org.Scope {
	tenantID, err := uuid.Parse(v.tenantID)
	if err != nil {
		return org.Scope{}
	}
	if v.orgUnitID == uuid.Nil {
		return org.TenantOnly(tenantID)
	}
	return org.WithUnit(tenantID, v.orgUnitID)
}

var _ definition.ViewerContext = (*sessionViewer)(nil)

// ViewerFromFiber builds the ViewerFromCtx function used by generic handlers.
// It reads session locals populated by the authentication middleware.
//
// Expected Fiber locals (set by auth middleware before this runs):
//   - "actor_id"    string — authenticated user UUID
//   - "tenant_id"   string — tenant UUID (from validated session)
//   - "org_unit_id" string — active org unit UUID (optional; uuid.Nil = tenant-wide)
//   - "roles"       []string — role names held by the actor
//   - "is_system"   bool — true for machine/service tokens
//
// Falls back gracefully when locals are absent (e.g. dev server without auth middleware).
func ViewerFromFiber() api.ViewerFromCtx {
	return func(c *fiber.Ctx) (definition.ViewerContext, error) {
		actorID, _ := c.Locals("actor_id").(string)
		tenantID, _ := c.Locals("tenant_id").(string)

		// Fallback: read tenant from header when middleware hasn't run.
		if tenantID == "" {
			tenantID = c.Get("X-Awo-Tenant")
		}
		if actorID == "" {
			actorID = "anonymous"
		}

		// Parse org_unit_id; uuid.Nil signals tenant-wide scope.
		var orgUnitID uuid.UUID
		if raw, _ := c.Locals("org_unit_id").(string); raw != "" {
			orgUnitID, _ = uuid.Parse(raw)
		}

		roleList, _ := c.Locals("roles").([]string)
		roles := make(map[string]bool, len(roleList))
		for _, r := range roleList {
			roles[r] = true
		}

		isSystem, _ := c.Locals("is_system").(bool)

		return &sessionViewer{
			actorID:   actorID,
			tenantID:  tenantID,
			orgUnitID: orgUnitID,
			roles:     roles,
			isSystem:  isSystem,
		}, nil
	}
}
