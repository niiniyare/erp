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
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// snapshot holds the resolved identity attributes baked into a Session at login.
type snapshot struct {
	OrgUnitID   uuid.UUID
	Roles       []string
	Permissions []string
}

// computeSnapshot queries the DB for the user's primary OU, active role slugs,
// and net-granted permission strings.
//
// Rules:
//   - Roles come from active, non-expired user_role_assignments.
//   - Permissions are the UNION of all granted permissions across those roles.
//   - An explicit deny (granted=FALSE) on any role removes the permission for
//     that user regardless of grants from other roles (fail-closed).
//   - Role inheritance (parent_role_id) is NOT walked in this MVP; parent perms
//     must be duplicated on the child or fetched separately.
//
// Caller must have called set_tenant_context before this function.
func computeSnapshot(ctx context.Context, q pgQuerier, userID, tenantID uuid.UUID) (*snapshot, error) {
	snap := &snapshot{}

	// Primary org unit (uuid.Nil if not set).
	var orgUnitID *uuid.UUID
	if err := q.QueryRow(ctx,
		`SELECT org_unit_id FROM tenant_users WHERE id = $1`, userID,
	).Scan(&orgUnitID); err != nil {
		return nil, fmt.Errorf("iam: load user org unit: %w", err)
	}
	if orgUnitID != nil {
		snap.OrgUnitID = *orgUnitID
	}

	// Active role slugs.
	roleRows, err := q.Query(ctx, `
		SELECT DISTINCT r.slug
		FROM user_role_assignments ura
		JOIN roles r ON r.id = ura.role_id
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

	// Net-granted permissions: include perm only if ALL grants for it are TRUE.
	// Any explicit deny (granted=FALSE) removes the permission entirely.
	permRows, err := q.Query(ctx, `
		SELECT p.permission
		FROM user_role_assignments ura
		JOIN permissions p ON p.role_id = ura.role_id
		WHERE ura.user_id   = $1
		  AND ura.tenant_id = $2
		  AND ura.is_active = TRUE
		  AND (ura.expires_at IS NULL OR ura.expires_at > NOW())
		GROUP BY p.permission
		HAVING bool_and(p.granted) = TRUE
		ORDER BY p.permission`,
		userID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("iam: query permissions: %w", err)
	}
	defer permRows.Close()
	for permRows.Next() {
		var perm string
		if err := permRows.Scan(&perm); err != nil {
			return nil, fmt.Errorf("iam: scan permission: %w", err)
		}
		snap.Permissions = append(snap.Permissions, perm)
	}
	if err := permRows.Err(); err != nil {
		return nil, fmt.Errorf("iam: permissions rows: %w", err)
	}

	return snap, nil
}
