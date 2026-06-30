package privacy

import (
	"context"

	"github.com/google/uuid"

	"awo.so/framework/def"
	"awo.so/framework/platform/org"
)

// AllowWithinOrgScope returns a PolicyFunc that grants access when the record's
// org unit is a descendant-or-equal of the viewer's org unit.
//
// Grants immediately (ErrAllow) when:
//   - viewer is a system caller
//   - viewer is tenant-wide (OrgUnitID == Nil) — sees all units in the tenant
//
// Abstains (ErrSkip) when:
//   - record is nil (list-level check — the handler applies a subtree filter separately)
//   - record does not implement def.OrgScoped, or its org unit is Nil
//
// Denies (ErrDeny) when:
//   - the viewer's org unit is not an ancestor-or-equal of the record's org unit
//
// Usage:
//
//	Policies: []def.PolicyDef{
//	    def.Policy(def.OpAll, def.AllowSystem),
//	    def.Policy(def.OpAll, privacy.AllowWithinOrgScope(tree)),
//	}
func AllowWithinOrgScope(tree org.Tree) def.PolicyFunc {
	return func(ctx context.Context, viewer def.ViewerContext, _ def.Op, record def.Record) error {
		if record == nil {
			// List-level check: the handler injects an org_unit_id IN (subtree)
			// predicate into the query; no per-record evaluation needed.
			return def.ErrSkip
		}
		if viewer.IsSystem() {
			return def.ErrAllow
		}
		viewerUnitID := viewer.OrgUnitID()
		if viewerUnitID == uuid.Nil {
			// Tenant-wide viewer: unrestricted access within the tenant.
			return def.ErrAllow
		}

		var recordUnitID uuid.UUID
		if scoped, ok := record.(def.OrgScoped); ok {
			recordUnitID = scoped.RecordOrgUnitID()
		} else {
			// Fallback: not all record implementations expose OrgScoped.
			// def.OrgScoped is the preferred path; mapRecord implements it.
			return def.ErrSkip
		}
		if recordUnitID == uuid.Nil {
			return def.ErrSkip // record has no org unit; not our concern
		}

		tenantID, _ := uuid.Parse(viewer.TenantID())
		ok, err := tree.IsAncestorOrEqual(ctx, tenantID, viewerUnitID, recordUnitID)
		if err != nil {
			// Error treated as ErrDeny by the enforcer; logged upstream.
			return err
		}
		if !ok {
			return def.ErrDeny
		}
		return def.ErrAllow
	}
}
