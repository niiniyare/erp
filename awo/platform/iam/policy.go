package iam

import (
	"context"

	"awo.so/awo/def"
)

// UserPolicy applies no additional row restriction for iam_user beyond RLS.
// Tenant isolation is enforced by the database. Casbin gates Read/Write access
// by role so platform-admin sees all users, tenant.admin sees tenant users only.
var UserPolicy def.PolicyFunc = func(_ context.Context) def.Filter {
	return nil
}

// SessionPolicy restricts sessions to those owned by the current actor.
// Platform admins bypass this policy (Casbin handles the bypass).
var SessionPolicy def.PolicyFunc = func(_ context.Context) def.Filter {
	// TODO: inject actor from context and return filter.Eq("user_id", actor.UserID)
	// when filter package is available.
	return nil
}
