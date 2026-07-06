package def

// PermissionSet declares the RBAC gates for each operation on an entity.
// Each field is a list of Casbin role or user subjects that are allowed to
// perform the operation. An empty slice means "deny all" (no one can perform
// the operation).
//
// The platform-admin role ("role:platform-admin") bypasses all Casbin checks
// and is never listed here — it is granted unconditionally by the IAM
// middleware.
//
// Subject formats:
//   - "role:{name}"  — e.g. "role:tenant.admin", "role:finance.accounts_payable"
//   - "user:{uuid}"  — per-user grant (uncommon; prefer role-based grants)
//
// Role inheritance accumulates permissions upward via Casbin g assertions.
// Granting a parent role implicitly grants all child role permissions.
type PermissionSet struct {
	// Create lists subjects allowed to create new records.
	Create []string

	// Read lists subjects allowed to read records.
	Read []string

	// Write lists subjects allowed to update existing records.
	Write []string

	// Delete lists subjects allowed to delete records.
	Delete []string

	// Actions maps action names to their allowed subjects. The framework
	// merges these with the route-level permission declared on [ActionDef].
	// If both are set, both are checked (AND semantics).
	Actions map[string][]string

	// Policy is the row-level filter applied to all read operations for this
	// entity. Nil means no additional row restriction beyond RLS.
	Policy PolicyFunc
}
