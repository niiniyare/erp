package authz

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
// Permissions granted to role:tenant.admin:
//   - finance.*.*   read/create/update/delete/approve/export
//   - people.*.*    read/create/update/delete
//   - settings.*.*  read/update
func SeedDefaultRoles(ctx context.Context, svc Service, tenantID string) error {
	domain := TenantDomain(tenantID)
	role := "role:tenant.admin"

	policies := []Policy{
		// Finance module
		{Subject: role, Domain: domain, Object: "finance.receivables.invoices", Action: "read", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "finance.receivables.invoices", Action: "create", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "finance.receivables.invoices", Action: "update", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "finance.receivables.invoices", Action: "delete", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "finance.receivables.invoices", Action: "approve", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "finance.receivables.invoices", Action: "export", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "finance.payables.bills", Action: "read", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "finance.payables.bills", Action: "create", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "finance.payables.bills", Action: "update", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "finance.payables.bills", Action: "approve", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "finance.accounts", Action: "read", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "finance.accounts", Action: "create", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "finance.accounts", Action: "update", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "finance.transactions", Action: "read", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "finance.transactions", Action: "create", Effect: "allow"},

		// People module
		{Subject: role, Domain: domain, Object: "people.employees", Action: "read", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "people.employees", Action: "create", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "people.employees", Action: "update", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "people.employees", Action: "delete", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "people.persons", Action: "read", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "people.persons", Action: "create", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "people.persons", Action: "update", Effect: "allow"},

		// Settings module
		{Subject: role, Domain: domain, Object: "settings.iam", Action: "read", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "settings.iam", Action: "update", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "settings.general", Action: "read", Effect: "allow"},
		{Subject: role, Domain: domain, Object: "settings.general", Action: "update", Effect: "allow"},
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
