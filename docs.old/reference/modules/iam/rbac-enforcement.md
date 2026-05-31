[<-- Back to Index](README.md)

## RBAC Enforcement — Single-Path Casbin Model

> **[IMPLEMENTED]** — the authoritative description of how authorization decisions are made
> in v1.0. Every permission check in the system flows through this path.

---

### The Single Enforcement Authority

In v1.0, **Casbin is the only authorization authority**. There is no secondary permission map, no `session.Can()` shortcut, and no parallel enforcement path. Every `allow` / `deny` decision passes through `authzService.Enforce()`.

```
Request arrives
    │
    ▼
Authenticate → build ResolvedSession (context only — no permissions)
    │
    ▼
ToPrincipal() → (Subject, Domain)
    │
    ▼
authzService.Enforce(ctx, domain.Request{Subject, Domain, Object, Action})
    │
    ├── revokeExpiredRoles(Subject, Domain)   [lazy cleanup — non-fatal]
    ├── enforcer.Enforce(sub, dom, obj, act)  [in-memory SyncedEnforcer]
    │     ├── g-rules: does Subject have role R in Domain?
    │     ├── p-rules: does role R have Object/Action policy in Domain?
    │     └── effect: some(allow) && !some(deny)
    │
    ├── false → 403 Forbidden
    └── true  → c.Next()
```

---

### What Goes Into an Enforce Call

Every `Enforce` call requires four values: `Subject`, `Domain`, `Object`, `Action`.

**Subject and Domain** always come from `session.ToPrincipal()`. They are derived from the authenticated session's `UserType` and `TenantID` — never from the HTTP request body or headers.

```go
principal := sess.ToPrincipal()
// principal.Subject = "tenant:<userID>"   (for a tenant user)
// principal.Domain  = "<tenantID>"
```

**Object** is the resource being accessed. Convention: `"{resource}/{id}"` for instance-level, `"{resource}"` for collection-level.

```go
// Collection: "invoice"   (list, create)
// Instance:   "invoice/inv_123"
// Sub-action: "invoice/inv_123/approve"
```

**Action** is the operation verb: `"read"`, `"create"`, `"update"`, `"delete"`, `"approve"`, `"export"`, etc.

---

### Full Casbin Model Evaluation Steps

For a request `(sub, dom, obj, act)`:

**Step 1 — Role expansion (g-rules)**

For all g-rules `(sub, role, domain)` where `domain == dom`:
- Collect all roles directly assigned to `sub`.
- Recursively expand role-to-role inheritance.
- Build the effective role set.

**Step 2 — Policy match (p-rules)**

For each effective role in the set, check all p-rules `(role, dom, obj_pattern, act_pattern, eft)`:
- `r.dom == p.dom` — hard domain equality check.
- `keyMatch2(r.obj, p.obj)` — URL-style wildcard on resource.
- `keyMatch(r.act, p.act)` — glob wildcard on action.

**Step 3 — Effect**

```
allow  if: ≥1 matching p-rule has eft="allow"
       AND 0 matching p-rules have eft="deny"

deny   otherwise (default deny, or any deny rule present)
```

---

### Actor Types and Domain Mapping

| UserType | ActorType | Subject prefix | Domain |
|---|---|---|---|
| SYSADMIN, PLATFORM | ActorPlatform | `platform:` | `_platform_` |
| PORTAL, CUSTOMER | ActorPortal | `portal:` | `<tenantID>:portal` |
| API, SERVICE | ActorAPI | `api:` | `<tenantID>:api` |
| INTERNAL, (default) | ActorTenant | `tenant:` | `<tenantID>` |

Each actor type lives in its own domain. A `tenant:` subject has no policies or roles in the `_platform_` domain even if the same UUID exists there.

---

### Bootstrapped Wildcard Policy

When a tenant user is first registered (via `RegisterNewUser` with `authz` wired), `BootstrapTenantAdmin` runs:

```
Adds p-rule:  (tenant_admin, {tenantID}, *, *, allow)
Adds g-rule:  (tenant:<userID>, tenant_admin, {tenantID})
```

This gives the first user in a tenant full access to all objects and actions within that tenant's domain. Subsequent users are not auto-assigned any role — they must be explicitly assigned by a `tenant_admin`.

---

### Policy Examples

**Allow role to read all invoices:**
```
p | finance-viewer | {tenantID} | invoice/* | read | allow
```

**Allow role to do anything with invoices:**
```
p | finance-manager | {tenantID} | invoice/* | * | allow
```

**Deny specific user all access (terminated employee):**
```
p | tenant:<userID> | {tenantID} | * | * | deny
```

**Allow role-to-role inheritance (finance-manager inherits finance-viewer):**
```
g | finance-manager | finance-viewer | {tenantID}
```

**Assign user to role:**
```
g | tenant:<userID> | finance-manager | {tenantID}
```

---

### EnforceBatch

When a handler needs to check multiple permissions at once (e.g. to build a permissions object for the frontend), use `EnforceBatch`:

```go
results, err := authzSvc.EnforceBatch(ctx, []domain.Request{
    {Subject: p.Subject, Domain: p.Domain, Object: "invoice/" + id, Action: "read"},
    {Subject: p.Subject, Domain: p.Domain, Object: "invoice/" + id, Action: "update"},
    {Subject: p.Subject, Domain: p.Domain, Object: "invoice/" + id, Action: "delete"},
    {Subject: p.Subject, Domain: p.Domain, Object: "invoice/" + id, Action: "approve"},
})
// results[0]=canRead, [1]=canUpdate, [2]=canDelete, [3]=canApprove
```

The entire batch is evaluated in the in-memory model in a single call. Latency is approximately the same as a single `Enforce()` call for small batches.

---

### What Authorization Does NOT Cover

Authorization (`Enforce`) answers: "is this actor allowed to perform this action on this resource type?"

It does NOT answer:
- "Does this specific invoice row belong to this tenant?" — enforced by PostgreSQL RLS.
- "Is this user's entity scope allowed to see this entity?" — enforced by `EntityScope` WHERE clauses in the repository layer.
- "Is this feature enabled for this tenant?" — answered by `session.FeatureEnabled()`, not authorization.

These are distinct enforcement layers that work together. Authorization is the outermost gate; RLS and entity scope filtering are inner gates.

---

### Default Deny

Casbin's deny-override effect model means: if no matching allow policy exists, the result is `DENY`. There is no "allow by default" state. A user with no roles assigned to them in a domain cannot perform any action in that domain.

This means:
- Creating a user does not grant them any access unless `BootstrapTenantAdmin` runs (for the first tenant user) or an admin explicitly assigns a role.
- Removing all roles from a user immediately locks them out of all resources (next `Enforce()` call returns `DENY`).

---

Next: [Deferred Features](./deferred-features.md)
