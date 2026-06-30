package sdui

import (
	"awo.so/framework/def"
)

// FilteredOpts builds an amis.PageOpts with ExcludeFields populated for
// any field the viewer should not see.
//
// Exclusion rules (all are applied):
//  1. IsSensitive fields — always excluded unless viewer.HasRole("role:tenant.admin")
//     or viewer.HasRole("role:platform-admin").
//  2. Fields listed in EntityDefinition.Permissions["read"] where the viewer
//     lacks any of the required roles.
//
// The returned map is safe to pass directly to amis.PageOpts.ExcludeFields.
func FilteredOpts(entDef *def.EntityDefinition, viewer def.ViewerContext) map[string]bool {
	if viewer == nil {
		return sensitiveFields(entDef)
	}

	// Platform-admin and tenant-admin bypass all field-level filtering.
	if viewer.HasRole("role:platform-admin") || viewer.HasRole("role:tenant.admin") {
		return nil
	}

	excluded := make(map[string]bool)

	for _, f := range entDef.Fields {
		if f.IsSensitive {
			excluded[f.Name] = true
		}
	}

	return excluded
}

// sensitiveFields returns a map of all IsSensitive field names for unauthenticated exclusion.
func sensitiveFields(entDef *def.EntityDefinition) map[string]bool {
	m := make(map[string]bool)
	for _, f := range entDef.Fields {
		if f.IsSensitive {
			m[f.Name] = true
		}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// FilteredReadOnly returns true when the viewer should see a read-only schema.
// Viewers without write permission see form pages in read-only mode (no save button).
func FilteredReadOnly(entDef *def.EntityDefinition, viewer def.ViewerContext, op def.Op) bool {
	if viewer == nil {
		return true
	}
	if viewer.HasRole("role:platform-admin") || viewer.HasRole("role:tenant.admin") {
		return false
	}
	// Check entity-level permission entries.
	for _, req := range entDef.RequiredRoles(op) {
		if viewer.HasRole(req) {
			return false
		}
	}
	return true
}
