[<-- Back to Index](README.md)

## Casbin Policy Engine

> **[IMPLEMENTED]** — Casbin is the sole authorization authority in v1.0.
> All permission decisions go through `authzService.Enforce()`.

---

### The Casbin CONF Model

Defined in `internal/core/iam/domain/authz.go` as the `CasbinModel` constant.
This is the canonical definition — the enforcer is built from this string at startup.

```ini
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act, eft

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && keyMatch(r.act, p.act)
```

**Key design decisions:**

1. **Deny-override effect**: `some(allow) && !some(deny)` — one explicit deny beats all allows.
2. **Domain-scoped RBAC**: `g = _, _, _` (three-parameter role). Roles are tenant-local.
3. **`keyMatch2` on `obj`**: URL-style `:param` wildcards (`invoice/:id` matches `invoice/abc-123`). Also matches glob `*`.
4. **`keyMatch` on `act`**: Glob wildcards only. `*` matches any verb. (`keyMatch2` would break `*` action policies — it uses `:param` syntax, not glob.)

---

### Request Tuple

```go
// domain.Request — the Casbin (sub, dom, obj, act) tuple
authz.Request{
    Subject: "tenant:550e8400-e29b-41d4-a716-446655440000",  // who
    Domain:  "a1b2c3d4-tenant-uuid",                          // which tenant
    Object:  "invoice/inv_2026_001",                          // what resource
    Action:  "read",                                          // what action
}
```

Subject format: `"<actor-class>:<id>"` — built by helpers in `domain/authz.go`:
- `PlatformSubject(userID)` → `"platform:<uuid>"`
- `TenantSubject(userID)` → `"tenant:<uuid>"`
- `PortalSubject(userID)` → `"portal:<uuid>"`
- `APISubject(clientID)` → `"api:<client-id>"`

Domain format:
- `TenantDomain(tenantID)` → `"<tenantUUID>"` (bare UUID)
- `DomainPlatform` constant → `"_platform_"`
- `PortalDomain(tenantID)` → `"<tenantUUID>:portal"`
- `APIDomain(tenantID)` → `"<tenantUUID>:api"`

---

### Policy Rules in casbin_rule

```
ptype | v0 (sub)               | v1 (dom)        | v2 (obj)    | v3 (act)  | v4 (eft)
──────┼────────────────────────┼─────────────────┼─────────────┼───────────┼─────────
p     | tenant_admin           | {tenantID}      | *           | *         | allow
p     | finance-manager        | {tenantID}      | invoice/*   | *         | allow
p     | finance-viewer         | {tenantID}      | invoice/*   | read      | allow
p     | tenant:sanctioned-user | {tenantID}      | *           | *         | deny
```

### Role Assignment (g-rules) in casbin_rule

```
ptype | v0 (user)              | v1 (role)        | v2 (domain)
──────┼────────────────────────┼──────────────────┼────────────
g     | tenant:usr_a1b2c3d4    | tenant_admin     | {tenantID}
g     | tenant:usr_b2c3d4e5    | finance-manager  | {tenantID}
g     | finance-manager        | finance-viewer   | {tenantID}   ← role inherits from role
```

---

### Effect Model: Deny-Override

```
ALLOW  if: at least one matching policy says "allow"
       AND no matching policy says "deny"

DENY   if: no matching allow rule
       OR  at least one matching deny rule (one deny beats all allows)
```

---

### Pattern Matching Reference

**`keyMatch2`** on `obj`:

| Pattern    | Request Object        | Match? |
|------------|-----------------------|--------|
| `invoice/:id` | `invoice/inv_123`  | Yes    |
| `invoice/*`   | `invoice/inv_123`  | Yes    |
| `*`           | `invoice/inv_123`  | Yes    |
| `invoice/*`   | `invoice/inv_123/pdf` | No (single segment) |

**`keyMatch`** on `act`:

| Pattern | Request Action | Match? |
|---------|----------------|--------|
| `read`  | `read`         | Yes    |
| `*`     | `read`         | Yes    |
| `*`     | `delete`       | Yes    |
| `read`  | `delete`       | No     |

---

### Role Inheritance

Role-to-role g-rules create inheritance. `finance-manager` inheriting from `finance-viewer` means `finance-manager` gets all policies assigned to `finance-viewer` as well as its own.

```
g | finance-manager | finance-viewer | {dom}

Effective: a user with finance-manager gets all policies of:
  finance-manager (direct)
  finance-viewer  (inherited)
```

`GetImplicitRoles(subject, domain)` returns all roles including inherited ones. `GetRoles(subject, domain)` returns only directly assigned roles.

---

### Bootstrapped Roles

On tenant user creation (when `authz` is wired into `UserService`), `BootstrapTenantAdmin` is called:

1. Seeds policy: `(tenant_admin, {tenantID}, *, *, allow)` — wildcard allow for tenant_admin.
2. Assigns user to `tenant_admin` role in that domain.
3. Idempotent: duplicate policies return `ErrPolicyConflict` (silently ignored).

**Important**: Bootstrap only runs when `authz` is non-nil in `NewUserServiceWithConfig`. The first user created for a tenant gets `tenant_admin`. Subsequent users do not get auto-assigned any role.

---

### Multi-Instance Propagation [IMPLEMENTED]

- `SyncedEnforcer` with `StartAutoLoadPolicy(30 * time.Second)`: all instances reload from DB every 30s.
- Policy writes (AddPolicy, AssignRole) use `EnableAutoSave(true)` — writes go to DB and update in-memory immediately on the calling instance.
- Other instances pick up changes at the next 30s reload cycle.
- Maximum inconsistency window: 30 seconds.
- `InvalidateCache()` forces immediate reload on the calling instance only.

---

### Lazy Expiry Cleanup

On every `Enforce()` call, `revokeExpiredRoles(subject, domain)` runs:
- Queries `role_assignments` for active rows where `expires_at < NOW()` for the given subject+domain.
- Uses partial index `idx_role_assignments_expires` (only rows with non-null `expires_at`).
- For most subjects: 0 rows, near-zero I/O.
- Removes expired roles from the in-memory enforcer and marks DB rows inactive.
- Non-fatal: enforce proceeds even if cleanup fails.

---

Next: [Database Architecture](./06-database-architecture.md)
