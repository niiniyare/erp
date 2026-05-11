[<-- Back to Index](README.md)

## Role Management

> **[IMPLEMENTED]** — describes the v1.0 role model as built.
> The `roles` table (migration 000405) is reserved for v2.0 ABAC and is NOT active.
> In v1.0, roles are Casbin string names (no `roles` table row required).

---

### What Is a Role?

A role is a named string used in Casbin g-rules. When you assign `"tenant_admin"` to a user in domain `"{tenantID}"`, Casbin creates:

```
g | tenant:<userID> | tenant_admin | {tenantID}
```

When you add a policy for `"tenant_admin"`:

```
p | tenant_admin | {tenantID} | * | * | allow
```

Then `Enforce("tenant:<userID>", "{tenantID}", "invoice/123", "read")` returns `true` because:
1. The user has role `tenant_admin` in `{tenantID}` (g-rule).
2. `tenant_admin` has a wildcard allow policy in `{tenantID}` (p-rule).
3. No deny rule overrides it.

Roles are **domain-scoped**. The same role name in Tenant A is completely independent from Tenant B.

---

### System Roles vs Custom Roles [IMPLEMENTED]

In v1.0 there is no formal `roles` DB table. "System roles" are a convention:

- `tenant_admin` is seeded by `BootstrapTenantAdmin()` with a wildcard allow policy.
- Other role names are arbitrary strings — no validation exists at the service layer that a role name is "registered."
- **[PLANNED - NOT IN v1.0]**: The `roles` table (migration 000405) provides typed role definitions with `is_system_role`, `role_type`, and `parent_role_id`. This enables formal immutability enforcement and UI role management. Not active in v1.0.

---

### AssignRole [IMPLEMENTED]

```go
AuthzService.AssignRole(ctx, tenantID, subject, role, domainName string, opts ...AssignOpt) error
```

**What happens:**

1. `enforcer.AddGroupingPolicy(subject, role, domainName)` — adds g-rule to Casbin in-memory model and writes to `casbin_rule` via AutoSave.
2. `repo.UpsertRoleAssignment(ctx, ...)` — writes metadata to `role_assignments` table.

On conflict (same subject+role+domain already exists): upsert updates the existing row (idempotent).

**Functional options:**

```go
// Permanent (default)
svc.AssignRole(ctx, tenantID.String(), "tenant:usr_001", "finance-manager", domain)

// Time-limited
svc.AssignRole(ctx, tenantID.String(), "tenant:usr_001", "auditor", domain,
    iam.WithExpiry(time.Now().Add(30*24*time.Hour)),
    iam.WithAssignedBy("tenant:manager-5"),
)

// Delegated
svc.AssignRole(ctx, tenantID.String(), "tenant:usr_001", "sales-rep", domain,
    iam.WithAssignedBy("tenant:manager-5"),
    iam.WithDelegatedBy("tenant:ceo-1"),
)
```

**Session cache eviction on assignment**: Not performed automatically. The user's existing sessions remain valid. Since authorization is live via Casbin, the new role takes effect on the next `Enforce()` call — no re-login needed.

---

### RevokeRole [IMPLEMENTED]

```go
AuthzService.RevokeRole(ctx, subject, role, domainName string) error
```

**What happens:**

1. `enforcer.DeleteRoleForUserInDomain(subject, role, domainName)` — removes g-rule from in-memory model immediately.
2. `repo.DeactivateRoleAssignment(ctx, ...)` — sets `is_active=false` in `role_assignments` (row preserved for audit trail).
3. If `SessionInvalidator` is wired: `InvalidateByUser(userID)` — evicts all cached sessions for the user.

The `role_assignments` row is **not deleted** — it is kept inactive for compliance audit.

---

### GetRoles vs GetImplicitRoles [IMPLEMENTED]

```go
GetRoles(ctx, subject, domainName) ([]string, error)
```
Returns directly assigned roles only (no inheritance). Uses Casbin in-memory model — no DB query.

```go
GetImplicitRoles(ctx, subject, domainName) ([]string, error)
```
Returns all effective roles including those inherited through role-to-role g-rules. Use this when you need the complete effective permission set (e.g. for UI display).

---

### HasRole [IMPLEMENTED]

```go
AuthzService.HasRole(ctx, subject, role, domainName) (bool, error)
```
Checks whether subject directly holds a role in a domain. In-memory, no DB query.

---

### GetAssignments [IMPLEMENTED]

```go
AuthzService.GetAssignments(ctx, subject, domainName) ([]RoleAssignment, error)
```
Queries `role_assignments` table (not Casbin) for full metadata: active and inactive rows, assigned_by, delegated_by, expires_at. Use for audit UI.

---

### Role Inheritance [IMPLEMENTED]

Role-to-role inheritance is expressed with a g-rule where both user and role are role names:

```
g | finance-manager | finance-viewer | {domain}
```

This means `finance-manager` inherits all policies of `finance-viewer`. You can chain:

```
g | finance-manager | finance-viewer  | {dom}
g | finance-viewer  | report-viewer   | {dom}

User assigned: tenant:usr_cfo → finance-manager
Effective policies:
  finance-manager (direct) + finance-viewer (inherited) + report-viewer (inherited)
```

**Cycle prevention**: The domain model documents cycle prevention as a concern, but no Go-level cycle detection is implemented in v1.0. Avoid circular role inheritance — Casbin would loop indefinitely.

---

### Temporal Assignments [IMPLEMENTED]

Assignments with `WithExpiry(t)` are lazily expired: on every `Enforce()` call, `revokeExpiredRoles()` checks if any roles for the subject have `expires_at < NOW()` and removes them. The removal is best-effort (logged but non-fatal on failure).

---

### Role Naming in v1.0

No formal role registry. Seeded roles by convention:

```
tenant_admin     — wildcard allow for all objects/actions in tenant domain (seeded by BootstrapTenantAdmin)
```

Custom role names are free-form strings. By convention use lowercase kebab-case:

```
finance-manager
finance-viewer
sales-rep
auditor
portal-customer
api-readonly
```

Platform domain roles:
```
platform-admin       — use for cross-tenant admin subjects
platform-support     — read-only cross-tenant access
```

---

### Roles Table (V2.0 Reserved) [PLANNED - NOT IN v1.0]

Migration `000405_auth_create_roles.up.sql` creates a `roles` table with:
- `is_system_role BOOLEAN` — immutability flag
- `parent_role_id UUID` — DB-level hierarchy
- `role_type` (SYSTEM, TENANT, ENTITY, CUSTOM, FUNCTIONAL)
- `conditions JSONB` — time/location/device conditions (ABAC)
- `permissions JSONB` — cached permissions

This table is **deployed** (migration runs) but **not connected** to any service code in v1.0. The `//go:build ignore` gate on `internal/core/access/` means no code reads or writes this table at runtime.

---

Next: [Policy Management](./08-policy-management.md)
