package iam

import (
	"context"
	"errors"
	"fmt"
)

// SeedDefaultRoles provisions the built-in role policies for a new tenant.
// It is idempotent — duplicate policies are silently skipped.
//
// Call this from the tenant provisioning flow after the tenant row is created.
//
// Objects and actions match the authorizeMiddleware calls in the route layer
// exactly — splitPermission("finance.accounts.read") → ("finance.accounts","read").
//
// Permissions granted to role:tenant.admin:
//   - finance.accounts          read / write
//   - finance.transactions      read / write
//   - finance.periods           read / write
//   - finance.currencies        read / write
//   - finance.budgets           read / write
//   - finance.cost_centers      read / write
//   - finance.tax               read / write
//   - finance.reconciliation    read / write
//   - iam.policies              read / write
//   - iam.roles                 read / write
//   - contracts.*               read / write
func SeedDefaultRoles(ctx context.Context, svc Service, tenantID string) error {
	domain := TenantDomain(tenantID)
	role := "role:tenant.admin"

	// Each entry mirrors one authorizeMiddleware("<object>.<action>") call.
	// Keeping read and write as separate rows makes it easy to grant read-only
	// roles later without changing the policy model.
	objects := []string{
		// Finance
		"finance.accounts",
		"finance.transactions",
		"finance.periods",
		"finance.currencies",
		"finance.budgets",
		"finance.cost_centers",
		"finance.tax",
		"finance.reconciliation",
		// IAM management (within the tenant's own domain)
		"iam.policies",
		"iam.roles",
		// Contracts
		"contracts",
	}

	var policies []Policy
	for _, obj := range objects {
		policies = append(policies,
			Policy{Subject: role, Domain: domain, Object: obj, Action: "read", Effect: "allow"},
			Policy{Subject: role, Domain: domain, Object: obj, Action: "write", Effect: "allow"},
		)
	}

	for _, p := range policies {
		if err := svc.AddPolicy(ctx, p); err != nil {
			// ErrPolicyConflict means the policy already exists — idempotent, skip.
			if errors.Is(err, ErrPolicyConflict) {
				continue
			}
			return fmt.Errorf("authz SeedDefaultRoles: add policy %s %s/%s: %w",
				p.Object, p.Action, p.Effect, err)
		}
	}
	return nil
}

// AssignAdminRole grants the role:tenant.admin role to a user within a tenant.
// It is idempotent — re-assigning an already-held role updates the metadata.
//
// Call this after SeedDefaultRoles when provisioning the first admin user.
func AssignAdminRole(ctx context.Context, svc Service, tenantID, userID string) error {
	subject := TenantSubject(userID)
	domain := TenantDomain(tenantID)
	role := "role:tenant.admin"

	if err := svc.AssignRole(ctx, tenantID, subject, role, domain,
		WithAssignedBy(PlatformSubject("system")),
	); err != nil {
		return fmt.Errorf("authz AssignAdminRole: %w", err)
	}
	return nil
}
