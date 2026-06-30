package iam

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// pgQuerier is the minimal pgx interface needed for permission queries.
type pgQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// snapshot holds the resolved identity attributes baked into a Session at login.
type snapshot struct {
	OrgUnitID   uuid.UUID
	Roles       []string
	Permissions []string
}

// computeSnapshot queries the DB for the user's active role slugs.
//
// The current DB schema has iam_users, iam_roles, iam_user_roles.
// A separate permissions table does not yet exist; permissions are derived
// from role slugs (e.g. "tenant-admin" → "*.*.*") via rolePermissions().
//
// Caller must have called set_tenant_context before this function.
func computeSnapshot(ctx context.Context, q pgQuerier, userID, tenantID uuid.UUID) (*snapshot, error) {
	snap := &snapshot{OrgUnitID: uuid.Nil}

	// Active role slugs from iam_user_roles → iam_roles.
	roleRows, err := q.Query(ctx, `
		SELECT DISTINCT r.slug
		FROM iam_user_roles ura
		JOIN iam_roles r ON r.id = ura.role_id
		WHERE ura.user_id   = $1
		  AND ura.tenant_id = $2
		  AND ura.is_active = TRUE
		  AND (ura.expires_at IS NULL OR ura.expires_at > NOW())
		  AND r.is_active = TRUE
		ORDER BY r.slug`,
		userID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("iam: query roles: %w", err)
	}
	defer roleRows.Close()
	for roleRows.Next() {
		var slug string
		if err := roleRows.Scan(&slug); err != nil {
			return nil, fmt.Errorf("iam: scan role: %w", err)
		}
		snap.Roles = append(snap.Roles, slug)
	}
	if err := roleRows.Err(); err != nil {
		return nil, fmt.Errorf("iam: roles rows: %w", err)
	}

	// Derive permissions from role slugs.
	// TODO: replace with iam_permissions table query once schema is defined.
	snap.Permissions = rolePermissions(snap.Roles)
	return snap, nil
}

// rolePermissions maps role slugs to permission strings.
// Admins get wildcard; others get narrow defaults.
func rolePermissions(roles []string) []string {
	permSet := map[string]struct{}{}
	for _, r := range roles {
		switch r {
		case "tenant-admin", "role:tenant.admin":
			permSet["*.*.*"] = struct{}{}
		case "tenant-manager", "role:tenant.manager":
			permSet["*.read.*"] = struct{}{}
			permSet["*.write.*"] = struct{}{}
		default:
			permSet["*.read.*"] = struct{}{}
		}
	}
	out := make([]string, 0, len(permSet))
	for p := range permSet {
		out = append(out, p)
	}
	return out
}
