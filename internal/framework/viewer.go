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
	actorID    string
	tenantID   string
	companyID  string
	divisionID string
	roles      map[string]bool
	isSystem   bool
}

func (v *sessionViewer) ActorID() string   { return v.actorID }
func (v *sessionViewer) TenantID() string  { return v.tenantID }
func (v *sessionViewer) CompanyID() string { return v.companyID }
func (v *sessionViewer) DivisionID() string { return v.divisionID }
func (v *sessionViewer) IsSystem() bool        { return v.isSystem }
func (v *sessionViewer) HasRole(r string) bool { return v.roles[r] }

// OrgScope builds an org.Scope from the viewer's parsed IDs.
func (v *sessionViewer) OrgScope() org.Scope {
	tenantID, err := uuid.Parse(v.tenantID)
	if err != nil {
		return org.Scope{}
	}
	s := org.TenantOnly(tenantID)
	if v.companyID != "" {
		if cid, err := uuid.Parse(v.companyID); err == nil {
			s = org.WithCompany(tenantID, cid)
		}
	}
	if v.divisionID != "" && s.CompanyID != nil {
		if did, err := uuid.Parse(v.divisionID); err == nil {
			s = org.WithDivision(tenantID, *s.CompanyID, did)
		}
	}
	return s
}

var _ definition.ViewerContext = (*sessionViewer)(nil)

// ViewerFromFiber builds the ViewerFromCtx function used by generic handlers.
// It reads session locals populated by the authentication middleware.
//
// Expected Fiber locals (set by auth middleware before this runs):
//   - "actor_id"    string   — authenticated user UUID
//   - "tenant_id"   string   — tenant UUID (from validated session)
//   - "company_id"  string   — active company UUID (optional)
//   - "division_id" string   — active division UUID (optional)
//   - "roles"       []string — role names held by the actor
//   - "is_system"   bool     — true for machine/service tokens
//
// Falls back to dev stub when locals are absent (development only).
func ViewerFromFiber() api.ViewerFromCtx {
	return func(c *fiber.Ctx) (definition.ViewerContext, error) {
		actorID, _ := c.Locals("actor_id").(string)
		tenantID, _ := c.Locals("tenant_id").(string)
		companyID, _ := c.Locals("company_id").(string)
		divisionID, _ := c.Locals("division_id").(string)

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
			actorID:    actorID,
			tenantID:   tenantID,
			companyID:  companyID,
			divisionID: divisionID,
			roles:      roles,
			isSystem:   isSystem,
		}, nil
	}
}
