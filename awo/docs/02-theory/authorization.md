# Authorization Model

## Overview

Authorization in Awo is a two-stage pipeline evaluated on every request. The stages are independent and must not be conflated.

```
HTTP Request
  ↓
Stage 1: Tenant Isolation   (framework — PostgreSQL RLS)
  ↓
Stage 2: Organization Scope (application — OrganizationService)
  ↓
  Operation Permissions     (RBAC — Casbin, per entity per action)
  ↓
  Business Policies         (EntityDefinition.Policy — row-level filter)
  ↓
Repository.Query()
```

---

## Stage 1 — Tenant Isolation (Framework)

**Who enforces it:** PostgreSQL Row-Level Security via `current_tenant_id()`.

**What it guarantees:** Data for tenant A never appears in tenant B's queries, regardless of application code.

**How it works:**

1. Auth middleware resolves the tenant from `X-Tenant-ID` header (or subdomain).
2. `set_tenant_context($tenantID)` stored procedure runs at the start of each connection checkout.
3. PostgreSQL RLS policy fires on every table access: `USING (tenant_id = current_tenant_id())`.
4. No application query can bypass this — it is enforced at the DB kernel level.

**What it does NOT do:** It does not filter by organization, user, role, or any business concept. Tenant isolation is the floor — not the ceiling.

---

## Stage 2 — Organization Scope (Application)

**Who enforces it:** Application services via `OrganizationService.ResolveScope()`.

**What it governs:** Within a tenant, which organization nodes is the current user allowed to see?

**How it works:**

1. Auth middleware loads the user's org assignments and populates `ViewerContext`.
2. Auth middleware sets `ViewerContext.VisibilityMode` based on user's IAM roles.
3. Application service calls `OrganizationService.ResolveScope(ctx, viewer)`.
4. Result is a `[]uuid.UUID` of visible org IDs (or `nil` for entire-tenant access).
5. Application service injects the IDs as an explicit `IN` predicate in the filter.

**What it does NOT do:** Organization scope does not use RLS. Organization tables (`platform_organization`, `platform_org_type`, `platform_org_assignment`) have no RLS policies. The application is solely responsible for applying org scope.

### Sequence Diagram

```
Client          Middleware          OrgService       Repository
  │                 │                   │                │
  │── GET /invoices ▶│                   │                │
  │                 │── validate session │                │
  │                 │── load org assigns │                │
  │                 │── set ViewerCtx    │                │
  │                 │── set_tenant_ctx() │                │
  │                 │                   │                │
  │           InvoiceService.List(ctx)  │                │
  │                 │── ResolveScope ──▶│                │
  │                 │◀── []uuid.UUID ───│                │
  │                 │── build filter    │                │
  │                 │── repo.Query ────────────────────▶│
  │                 │◀─ records ────────────────────────│
  │◀── 200 JSON ────│                   │                │
```

---

## Operation Permissions (RBAC)

After tenant isolation and org scope are applied, Casbin evaluates whether the authenticated user may perform the requested operation on the target entity type.

**Policy model:** `(subject, domain, object, action)`

- Subject: `user:{uuid}` or `role:{name}`
- Domain: tenant UUID or `_platform_`
- Object: entity type name (e.g. `invoice`, `platform_organization`)
- Action: `read`, `write`, `create`, `delete`, `submit`, `cancel`, …

**Evaluation:**

```go
enforcer.Enforce(userID, tenantID, entityType, action) → bool
```

Casbin policies are compiled from `EntityDefinition.Permissions` at startup. The Casbin enforcer is wired in `api/authz` and applied per-route.

---

## Business Policies (Row-Level Filter)

The final layer is the `EntityDefinition.Policy` function — an optional row-level predicate injected into every query for that entity type.

```go
Policy: def.PolicyFunc(func(ctx context.Context) def.Filter {
    viewer, _ := organization.ViewerFromContext(ctx)
    return filter.Eq("assigned_to", viewer.UserID.String())  // OwnerOnly
})
```

Policies are declared once on the entity definition and enforced everywhere. They run after tenant isolation and org scope, so they operate on the already-filtered dataset.

---

## VisibilityMode → SQL Predicate Mapping

| VisibilityMode | Resulting SQL (conceptual) |
|---|---|
| `VisibilityCurrent` | `org_id = $active_org_id` |
| `VisibilityDescendants` | `org_id IN (SELECT id FROM platform_organization WHERE path LIKE $prefix)` |
| `VisibilityAncestors` | `org_id IN ($ancestor_ids)` |
| `VisibilityEntireTenant` | *(no org filter — already tenant-scoped by RLS)* |
| `VisibilityExplicit` | `org_id IN ($explicit_ids)` |
| `VisibilityCustom` | `org_id IN ($resolver_result)` |

All predicates are constructed in Go and passed to the repository as typed `filter.Filter` values — no string interpolation, no SQL injection risk.

---

## Headquarters Users

HQ users require `VisibilityEntireTenant`. This is granted by:

1. IAM role `role:tenant.admin` — automatically resolves to `VisibilityEntireTenant`.
2. Explicit grant: middleware sets `VisibilityEntireTenant` based on a custom business rule.

No special framework path. HQ access is just `VisibilityEntireTenant` — the application omits the org filter, and RLS ensures the data stays within the tenant.

```
Holding Company Finance Manager
  IAM role: tenant.admin (or explicit HQ grant)
  VisibilityMode: VisibilityEntireTenant
  ResolveScope() → nil
  Query: SELECT * FROM invoice WHERE tenant_id = $1
         (no org filter — sees all branches, all regions)
```

---

## Multi-Organization Users

A user belonging to multiple orgs can switch their **active organization** during a session. The switch is validated against `ViewerContext.OrganizationAssignments` — a user cannot switch to an org they are not assigned to.

```
User: Jane
  Assignments: Nairobi Branch (manager), Mombasa Branch (viewer)
  Active org: Nairobi Branch

  → Scope: VisibilityDescendants of Nairobi
  → Sees: Nairobi + Retail Team + Credit Team

  Switch active org to Mombasa Branch:
  → Scope: VisibilityDescendants of Mombasa
  → Sees: Mombasa only (no sub-teams)
```

---

## Platform Admins

`role:platform-admin` bypasses both Casbin and org scope:

- No Casbin check — platform admins can perform any operation on any tenant.
- `VisibilityMode` set to `VisibilityEntireTenant` across all tenants.
- `ViewerContext.IsPlatformAdmin = true` — used for audit log annotation.

Platform admins never impersonate tenant users — all actions are logged under the platform admin's own identity.

---

## Summary

| Question | Answered By |
|---|---|
| Is this request for the right tenant? | RLS (`current_tenant_id()`) — framework |
| Which orgs can this user see? | `OrganizationService.ResolveScope()` — application |
| Can this user perform this action on this entity type? | Casbin RBAC — framework + application |
| Can this user see this specific row? | `EntityDefinition.Policy` — application |
