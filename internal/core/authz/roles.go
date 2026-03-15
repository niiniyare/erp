package authz

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/shared"
)

// roleMetadata holds static constraints for a built-in role.
// Roles not listed here are treated as unrestricted (custom/tenant-defined).
type roleMetadata struct {
	// AssignableTo lists the ActorType values that may assign this role.
	// Empty slice means any actor type may assign it.
	AssignableTo []string
}

// builtinRoles is the registry of roles whose assignment is restricted by actor type.
// Keys use the same strings returned by TenantSubject/PlatformSubject prefixes.
//
// Add entries here when introducing new system roles whose assignment should be
// limited to specific actor types (platform, tenant, portal).
var builtinRoles = map[string]roleMetadata{
	"role:platform.admin":   {AssignableTo: []string{string(ActorPlatform)}},
	"role:platform.support": {AssignableTo: []string{string(ActorPlatform)}},
	"role:tenant.admin":     {AssignableTo: []string{string(ActorPlatform), string(ActorTenant)}},
	"role:tenant.manager":   {AssignableTo: []string{string(ActorPlatform), string(ActorTenant)}},
	"role:portal.user":      {AssignableTo: []string{string(ActorTenant)}},
}

// actorTypeFromSubject extracts the ActorType prefix from a Casbin subject string.
// e.g. "platform:abc-123" → "platform", "tenant:xyz" → "tenant".
// Returns "" if the subject has no recognised prefix.
func actorTypeFromSubject(subject string) string {
	if idx := strings.Index(subject, ":"); idx > 0 {
		return subject[:idx]
	}
	return ""
}

// AssignRole grants subject the named role in domain and writes metadata to
// role_assignments. The Casbin g-rule is added after the DB write succeeds.
func (s *service) AssignRole(ctx context.Context, tenantID, subject, role, domain string, opts ...AssignOpt) error {
	o := &assignOpts{}
	for _, opt := range opts {
		opt(o)
	}

	// G5: AssignableTo guard — check that the assigning actor's type is allowed
	// to grant this particular role.  Only enforced for roles in builtinRoles;
	// custom tenant roles are unrestricted.
	if meta, ok := builtinRoles[role]; ok && len(meta.AssignableTo) > 0 && o.assignedBy != "" {
		actorType := actorTypeFromSubject(o.assignedBy)
		if !slices.Contains(meta.AssignableTo, actorType) {
			return fmt.Errorf("authz: role %q is not assignable by actor type %q (allowed: %v)",
				role, actorType, meta.AssignableTo)
		}
	}

	// Inject tenant UUID into context so the repo can use WithTenantFromCtx.
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return fmt.Errorf("authz AssignRole: invalid tenant ID %q: %w", tenantID, err)
	}
	ctx = shared.WithTenantID(ctx, tID)

	if err := s.repo.UpsertRoleAssignment(ctx, tID, subject, role, domain,
		nullableString(o.assignedBy), nullableString(o.delegatedBy), o.expiresAt,
	); err != nil {
		return fmt.Errorf("authz AssignRole: %w", err)
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
	// Inject tenant UUID when the domain is a parseable UUID (tenant-scoped roles).
	// Platform domain (_platform_) has no tenant UUID — repo falls back to WithTx.
	if tID, err := uuid.Parse(domain); err == nil {
		ctx = shared.WithTenantID(ctx, tID)
	}

	if err := s.repo.DeactivateRoleAssignment(ctx, subject, role, domain); err != nil {
		return fmt.Errorf("authz RevokeRole: %w", err)
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
	return s.repo.ListRoleAssignments(ctx, subject, domain)
}

// revokeExpiredRoles is called at the start of Enforce for lazy cleanup of
// any role assignments whose expires_at has passed.
func (s *service) revokeExpiredRoles(ctx context.Context, subject, domain string) error {
	expired, err := s.repo.ListExpiredActiveRoleNames(ctx, subject, domain)
	if err != nil {
		return err
	}
	for _, role := range expired {
		if err := s.RevokeRole(ctx, subject, role, domain); err != nil {
			return err
		}
	}
	return nil
}
