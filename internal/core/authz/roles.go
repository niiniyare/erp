package authz

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// AssignRole grants subject the named role in domain and writes metadata to
// role_assignments. The Casbin g-rule is added after the DB write succeeds.
func (s *service) AssignRole(ctx context.Context, tenantID, subject, role, domain string, opts ...AssignOpt) error {
	o := &assignOpts{}
	for _, opt := range opts {
		opt(o)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("authz AssignRole begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx, `
		INSERT INTO role_assignments
			(id, tenant_id, subject, role_name, domain, assigned_by, delegated_by, expires_at, is_active)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, TRUE)
		ON CONFLICT (subject, role_name, domain)
		DO UPDATE SET
			is_active    = TRUE,
			assigned_by  = EXCLUDED.assigned_by,
			delegated_by = EXCLUDED.delegated_by,
			expires_at   = EXCLUDED.expires_at`,
		uuid.New().String(), tenantID, subject, role, domain,
		nullableString(o.assignedBy), nullableString(o.delegatedBy), o.expiresAt,
	)
	if err != nil {
		return fmt.Errorf("authz AssignRole insert: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("authz AssignRole commit: %w", err)
	}

	// Add the Casbin g-rule after the DB write succeeds.
	if _, err := s.enforcer.AddRoleForUserInDomain(subject, role, domain); err != nil {
		return fmt.Errorf("authz AssignRole casbin: %w", err)
	}
	return nil
}

// RevokeRole removes subject's role in domain from Casbin and marks the
// role_assignments row inactive.
func (s *service) RevokeRole(ctx context.Context, subject, role, domain string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE role_assignments SET is_active = FALSE
		 WHERE subject=$1 AND role_name=$2 AND domain=$3`,
		subject, role, domain,
	)
	if err != nil {
		return fmt.Errorf("authz RevokeRole update: %w", err)
	}

	if _, err := s.enforcer.DeleteRoleForUserInDomain(subject, role, domain); err != nil {
		return fmt.Errorf("authz RevokeRole casbin: %w", err)
	}
	return nil
}

// GetRoles returns all roles subject holds in domain (from Casbin in-memory).
// Returns an empty slice (not an error) when subject has no roles.
func (s *service) GetRoles(ctx context.Context, subject, domain string) ([]string, error) {
	return s.enforcer.GetRolesForUserInDomain(subject, domain), nil
}

// HasRole reports whether subject currently holds role in domain.
func (s *service) HasRole(ctx context.Context, subject, role, domain string) (bool, error) {
	ok, err := s.enforcer.HasRoleForUser(subject, role, domain)
	if err != nil {
		return false, fmt.Errorf("authz HasRole: %w", err)
	}
	return ok, nil
}

// GetAssignments queries the role_assignments metadata table for subject+domain.
func (s *service) GetAssignments(ctx context.Context, subject, domain string) ([]RoleAssignment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, subject, role_name, domain, tenant_id,
		       COALESCE(assigned_by,''), COALESCE(delegated_by,''),
		       expires_at, is_active, created_at
		FROM role_assignments
		WHERE subject=$1 AND domain=$2
		ORDER BY created_at DESC`,
		subject, domain,
	)
	if err != nil {
		return nil, fmt.Errorf("authz GetAssignments query: %w", err)
	}
	defer rows.Close()

	var out []RoleAssignment
	for rows.Next() {
		var ra RoleAssignment
		if err := rows.Scan(
			&ra.ID, &ra.Subject, &ra.Role, &ra.Domain, &ra.TenantID,
			&ra.AssignedBy, &ra.DelegatedBy,
			&ra.ExpiresAt, &ra.IsActive, &ra.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("authz GetAssignments scan: %w", err)
		}
		out = append(out, ra)
	}
	return out, rows.Err()
}

// revokeExpiredRoles is called at the start of Enforce for lazy cleanup of
// any role assignments whose expires_at has passed.
func (s *service) revokeExpiredRoles(ctx context.Context, subject, domain string) error {
	rows, err := s.pool.Query(ctx, `
		SELECT role_name FROM role_assignments
		WHERE subject=$1 AND domain=$2
		  AND is_active = TRUE
		  AND expires_at IS NOT NULL
		  AND expires_at < NOW()`,
		subject, domain,
	)
	if err != nil {
		return fmt.Errorf("revokeExpiredRoles query: %w", err)
	}
	defer rows.Close()

	var expired []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return fmt.Errorf("revokeExpiredRoles scan: %w", err)
		}
		expired = append(expired, role)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, role := range expired {
		if err := s.RevokeRole(ctx, subject, role, domain); err != nil {
			return err
		}
	}
	return nil
}

// nullableString returns nil for empty strings so pgx stores SQL NULL.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
