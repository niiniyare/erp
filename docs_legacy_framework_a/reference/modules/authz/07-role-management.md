> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

[<-- Back to Index](README.md)

## Role Management

### What Is a Role?

A role is a named collection of permissions. Instead of assigning 50 individual policies to each user, you assign a role and the role carries all its policies. Roles are domain-scoped — the same role name in Tenant A is completely independent from Tenant B.

```markdown
ROLE ASSIGNMENT FLOW:

1. Platform admin defines role and its policies:
   AddPolicy(ctx, Policy{
       Subject: "role:finance-manager",
       Domain:  tenantID,
       Object:  "invoice/*",
       Action:  "*",
       Effect:  "allow",
   })

2. Tenant admin assigns user to role:
   AssignRole(ctx, tenantID, "tenant:usr_001", "role:finance-manager", domain)

3. User makes request:
   Enforce(ctx, Request{
       Subject: "tenant:usr_001",
       Domain:  tenantID,
       Object:  "invoice/inv_123",
       Action:  "read",
   })
   → Casbin checks: does usr_001 have role:finance-manager?
   → Yes → does role:finance-manager allow invoice/* read?
   → Yes → ALLOW
```

### AssignRole

**Signature:**
```go
AssignRole(ctx context.Context, tenantID, subject, role, domain string, opts ...AssignOpt) error
```

**What it does (transactionally):**
```markdown
1. Upsert into role_assignments:
   → Sets is_active=TRUE
   → Records assigned_by and delegated_by if provided
   → Sets expires_at if time-limited
   → ON CONFLICT → updates the existing row (idempotent)

2. Adds Casbin g-rule:
   → AddRoleForUserInDomain(subject, role, domain)
   → Updates in-memory enforcer immediately

Both steps in sequence (metadata first, then Casbin).
If Casbin step fails, metadata is already committed
but enforcer reload will restore consistency on next startup.
```

**Functional options:**
```go
// Permanent assignment (default)
svc.AssignRole(ctx, tenantID, sub, "role:sales-rep", dom)

// Time-limited: expires at contract end date
expiry := time.Date(2026, 6, 30, 23, 59, 59, 0, time.UTC)
svc.AssignRole(ctx, tenantID, sub, "role:auditor", dom,
    authz.WithExpiry(expiry),
    authz.WithAssignedBy("platform:admin-1"),
)

// Delegated: tenant manager assigned on behalf of another
svc.AssignRole(ctx, tenantID, sub, "role:sales-rep", dom,
    authz.WithAssignedBy("tenant:manager-5"),
    authz.WithDelegatedBy("tenant:ceo-1"),
)
```

### RevokeRole

**Signature:**
```go
RevokeRole(ctx context.Context, subject, role, domain string) error
```

**What it does:**
```markdown
1. UPDATE role_assignments SET is_active = FALSE
   → Keeps the row for audit trail
   → Just marks it inactive

2. DeleteRoleForUserInDomain(subject, role, domain)
   → Removes g-rule from Casbin in-memory model
   → Adapter removes from casbin_rule on next SavePolicy or
     directly via auto-save
```

**Important:** Revoking a role does NOT delete the `role_assignments` row. The row is preserved with `is_active = FALSE` for the audit trail. This means a compliance auditor can query:
```sql
SELECT * FROM role_assignments
WHERE subject = 'tenant:usr_001'
ORDER BY created_at DESC;
-- Shows full history: all roles ever held, when, by whom
```

### GetRoles

```go
GetRoles(ctx context.Context, subject, domain string) ([]string, error)
```

Returns the roles held by subject in domain **from the Casbin in-memory model**. This is fast (no DB query) and reflects the current effective state.

```go
roles, _ := svc.GetRoles(ctx, "tenant:usr_001", tenantDomain)
// ["role:finance-manager", "role:report-viewer"]
// (includes inherited roles if g-rules chain upward)
```

### HasRole

```go
HasRole(ctx context.Context, subject, role, domain string) (bool, error)
```

Checks whether a subject currently holds a specific role in a domain. Uses Casbin's in-memory check — no DB query.

```go
ok, _ := svc.HasRole(ctx, "tenant:usr_001", "role:finance-manager", dom)
// true → user has the role
// false → user does not have the role (not assigned, or expired)
```

### GetAssignments

```go
GetAssignments(ctx context.Context, subject, domain string) ([]RoleAssignment, error)
```

Queries `role_assignments` (not Casbin) for full metadata. Returns all rows — active and inactive — for a full audit view.

```go
assignments, _ := svc.GetAssignments(ctx, "tenant:usr_001", dom)
for _, a := range assignments {
    fmt.Printf("Role: %s, Active: %v, Expires: %v, AssignedBy: %s\n",
        a.Role, a.IsActive, a.ExpiresAt, a.AssignedBy)
}
```

### Role Naming Conventions

```markdown
RECOMMENDED NAMING:

System roles (defined once, apply per-tenant):
  role:tenant-admin          Full access within tenant
  role:finance-manager       Full finance access
  role:finance-viewer        Read-only finance
  role:sales-manager         Full sales + discount approval
  role:sales-rep             Create orders, view customers
  role:hr-admin              Payroll, employee records, hiring
  role:hr-viewer             Read-only HR
  role:inventory-manager     Stock adjust, PO create, cost view
  role:auditor               Read-only, all modules (time-limited typically)
  role:report-viewer         Download/view reports, no data mutation

Portal roles:
  role:portal-customer       View own invoices, pay, download statements
  role:portal-supplier       View POs, submit invoices, track payments

API roles:
  role:api-readonly          GET all resources
  role:api-full-access       All actions (trusted service accounts)
  role:api-invoice-submit    POST invoice/*, read customer/*

Platform roles:
  role:platform-admin        Everything in _platform_
  role:platform-support      Read tenant data, cannot modify
  role:platform-billing      Manage subscriptions and plans only
```

### Role Hierarchy Example

```markdown
FINANCE ROLE HIERARCHY:

         role:report-viewer
               │
     ┌─────────┘
     │
role:finance-viewer
  (inherits from report-viewer)
  + invoice/* read allow
  + payment/* read allow
     │
     └───────────────────┐
                         │
               role:finance-manager
                 (inherits from finance-viewer)
                 + invoice/* * allow
                 + payment/* * allow
                 + journal/* * allow
                         │
                         └──────────────────┐
                                            │
                                   role:cfo
                                     (inherits from finance-manager)
                                     + invoice/*/approve execute allow
                                     + budget/* * allow
                                     + period/*/close execute allow

g-rules to define this hierarchy:
  g | role:finance-viewer  | role:report-viewer   | {dom}
  g | role:finance-manager | role:finance-viewer  | {dom}
  g | role:cfo             | role:finance-manager | {dom}
```

---

Next: [Policy Management](./08-policy-management.md)
