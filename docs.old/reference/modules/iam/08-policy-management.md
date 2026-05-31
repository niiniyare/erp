[<-- Back to Index](README.md)

## Policy Management

### What Is a Policy?

A policy is a single row in `casbin_rule` with `ptype = "p"`. It describes: **who** can do **what** on **which resource** in **which domain**, and whether that is an **allow** or **deny**.

```go
type Policy struct {
    Subject string   // "role:finance-manager" or "tenant:usr_001" (direct grant)
    Domain  string   // "{tenantID}" or "_platform_"
    Object  string   // "invoice/*" or "invoice/inv_123" or "*"
    Action  string   // "read", "create", "delete", "execute", "approve", "*"
    Effect  string   // "allow" | "deny"
}
```

### AddPolicy

```go
AddPolicy(ctx context.Context, p Policy) error
```

Validates and inserts a policy rule. Returns `ErrPolicyConflict` (HTTP 409) if the exact rule already exists — idempotent re-add is handled by the caller catching this error.

```go
// Allow finance manager to do anything with invoices
err := svc.AddPolicy(ctx, authz.Policy{
    Subject: "role:finance-manager",
    Domain:  tenantDomain,
    Object:  "invoice/*",
    Action:  "*",
    Effect:  "allow",
})

// Block a specific user from ALL actions (sanctions/termination)
err = svc.AddPolicy(ctx, authz.Policy{
    Subject: "tenant:usr_terminated",
    Domain:  tenantDomain,
    Object:  "*",
    Action:  "*",
    Effect:  "deny",
})

// Deny high-risk export for a specific role
err = svc.AddPolicy(ctx, authz.Policy{
    Subject: "role:sales-rep",
    Domain:  tenantDomain,
    Object:  "report/*/export",
    Action:  "*",
    Effect:  "deny",
})
```

### RemovePolicy

```go
RemovePolicy(ctx context.Context, p Policy) error
```

Removes a policy rule by exact match. All five fields (Subject, Domain, Object, Action, Effect) must match the stored row.

```go
// Lift the block on a user
err := svc.RemovePolicy(ctx, authz.Policy{
    Subject: "tenant:usr_terminated",
    Domain:  tenantDomain,
    Object:  "*",
    Action:  "*",
    Effect:  "deny",
})
```

### GetPolicies

```go
GetPolicies(ctx context.Context, domain string) ([]Policy, error)
```

Returns all p-rules for a given domain. Uses Casbin's `GetFilteredPolicy(1, domain)` — the index `1` targets the domain field (v1). This is the method a UI policy management screen calls to list all rules.

```go
policies, _ := svc.GetPolicies(ctx, tenantDomain)
for _, p := range policies {
    fmt.Printf("%s can %s on %s [%s]\n", p.Subject, p.Action, p.Object, p.Effect)
}
// Output:
// role:finance-manager can * on invoice/* [allow]
// role:sales-rep can read on order/* [allow]
// tenant:usr_xxx can * on * [deny]
```

### InvalidateCache

```go
InvalidateCache(ctx context.Context) error
```

Forces a full policy reload from PostgreSQL into the Casbin in-memory model. Call this after:
- Bulk policy import via direct SQL
- External system modifying `casbin_rule` directly
- After applying a policy template

```go
// After bulk import:
err := svc.InvalidateCache(ctx)
```

In steady state, `EnableAutoSave(true)` means every individual `AddPolicy` / `RemovePolicy` / `AssignRole` / `RevokeRole` call syncs to the DB and updates the in-memory model automatically. `InvalidateCache` is for exceptional cases.

### Policy Design Patterns

#### Pattern 1: Role-Based (Preferred)

Assign policies to **roles**, not users. Then assign users to roles. This is the standard RBAC approach.

```markdown
✅ PREFERRED:

Policies:
  role:finance-manager | {dom} | invoice/* | * | allow

Role assignments:
  tenant:usr_001 → role:finance-manager
  tenant:usr_002 → role:finance-manager
  tenant:usr_003 → role:finance-manager

Effect: all 3 users get finance-manager permissions.
To revoke: revoke the role assignment, not the policy.
To expand permissions: add to role policy, all users get it instantly.
```

#### Pattern 2: Direct User Grant (For Exceptions)

Assign a policy directly to a user subject (not a role). Use sparingly — prefer roles.

```markdown
⚠️ USE SPARINGLY:

Scenario: CFO needs one-time access to a specific audit document.

Direct grant:
  tenant:usr_cfo | {dom} | document/audit-2025/final.pdf | read | allow

Better alternative:
  Create role:audit-2025-reader
  Assign role to usr_cfo with WithExpiry(auditDate)
  → Automatic cleanup when done
```

#### Pattern 3: Blanket Deny (Compliance / Termination)

```markdown
SCENARIO: Employee terminated — immediate access revocation.

Step 1: Revoke all roles (cleanup)
  RevokeRole(ctx, sub, "role:finance-manager", dom)
  RevokeRole(ctx, sub, "role:report-viewer", dom)

Step 2: Add explicit deny (defense in depth)
  AddPolicy(ctx, Policy{
      Subject: "tenant:usr_terminated",
      Domain:  dom, Object: "*", Action: "*", Effect: "deny",
  })

Even if a role was missed in Step 1, the deny rule ensures zero access.
This is the deny-override guarantee.
```

#### Pattern 4: Scoped API Client Policies

```markdown
SCENARIO: Third-party accounting software integration.
Should read invoices and create payments only.

Domain: {tenantID}:api
Subject: api:cli_accounting_sync

Policies:
  api:cli_accounting_sync | {tenantID}:api | invoice/* | read    | allow
  api:cli_accounting_sync | {tenantID}:api | payment/* | create  | allow
  api:cli_accounting_sync | {tenantID}:api | payment/* | read    | allow

Explicitly block everything else:
  api:cli_accounting_sync | {tenantID}:api | *         | delete  | deny
  api:cli_accounting_sync | {tenantID}:api | *         | execute | deny

Effect: Integration can only read invoices and create payments.
        Cannot delete, approve, or execute any action.
```

#### Pattern 5: Platform Cross-Tenant Policies

```markdown
SCENARIO: Platform admin needs to read all tenants' invoices for billing audit.

Platform domain policies:
  platform:admin | _platform_ | tenant/*/invoice | read | allow
                                ↑
                        "tenant/*/" prefix is the AWO convention
                        for cross-tenant resource paths

The platform admin's Subject is "platform:admin"
The domain is "_platform_"
The request passes r.dom == p.dom check in the matcher
```

### Policy Lifecycle

```markdown
POLICY LIFECYCLE:

1. CREATION
   Platform admin or tenant admin calls AddPolicy()
   → Inserted into casbin_rule (DB)
   → Added to Casbin in-memory model
   → Effective immediately

2. ACTIVE
   All Enforce() calls matching this policy are affected
   → In-memory model consulted (no DB read on hot path)

3. MODIFICATION
   No direct edit — must Remove then Add
   → Ensures clean audit trail
   → Prevents race conditions

4. REMOVAL
   Admin calls RemovePolicy()
   → Deleted from casbin_rule (DB)
   → Removed from in-memory model
   → Effective immediately

5. AUDIT
   All current policies visible via GetPolicies(ctx, domain)
   Historical changes tracked if audit module is hooked in
   Direct SQL on casbin_rule shows current state
```

---

Next: [Temporal Roles & Expiry](./09-temporal-roles-and-expiry.md)
