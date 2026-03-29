[<-- Back to Index](README.md)

## Casbin Policy Engine

### The Casbin CONF Model

The authorization module uses a single Casbin model defined in `model.go`. Every enforcement decision is evaluated against this model:

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

> **Note:** `keyMatch2` on `obj` supports URL-style path params (`invoice/:id` matches `invoice/abc-123`).
> `keyMatch` on `act` supports glob wildcards (`*` matches any verb like `read` or `delete`).
> Using `keyMatch2` on `act` was a prior bug — it treats `:param` syntax, so `*` would **never** match `read`, silently denying all wildcard-action policies.

### Breaking Down Each Section

#### `[request_definition]` — What an Enforce Call Looks Like

```markdown
r = sub, dom, obj, act

sub → WHO is acting         "tenant:usr_a1b2c3d4"
dom → IN WHICH scope        "a1b2c3d4-tenant-uuid"
obj → ON WHAT resource      "invoice/inv_2026_001"
act → DOING WHAT action     "read"

Go code:
  authz.Request{
      Subject: "tenant:usr_a1b2c3d4",
      Domain:  "a1b2c3d4-tenant-uuid",
      Object:  "invoice/inv_2026_001",
      Action:  "read",
  }
```

#### `[policy_definition]` — What a Policy Row Looks Like

```markdown
p = sub, dom, obj, act, eft

eft = "allow" or "deny"

Examples in casbin_rule table:
ptype | v0                     | v1             | v2          | v3     | v4    | v5
──────┼────────────────────────┼────────────────┼─────────────┼────────┼───────┼───
p     | role:finance-manager   | {tenantID}     | invoice/*   | *      | allow |
p     | role:sales-rep         | {tenantID}     | order/*     | read   | allow |
p     | role:sales-rep         | {tenantID}     | order/*     | create | allow |
p     | tenant:sanctioned-user | {tenantID}     | *           | *      | deny  |
p     | role:auditor           | {tenantID}     | */export    | *      | deny  |
```

#### `[role_definition]` — How Role Hierarchy Works

```markdown
g = _, _, _  (three-parameter role — user, role, domain)

Examples in casbin_rule table:
ptype | v0                     | v1                    | v2         | v3 | v4 | v5
──────┼────────────────────────┼───────────────────────┼────────────┼────┼────┼───
g     | tenant:usr_a1b2c3d4    | role:finance-manager  | {tenantID} |    |    |
g     | tenant:usr_b2c3d4e5    | role:sales-rep        | {tenantID} |    |    |
g     | role:finance-manager   | role:finance-viewer   | {tenantID} |    |    |
                                ↑
                        Role inherits from role:finance-viewer
                        finance-manager gets all viewer permissions
                        PLUS any policies assigned to finance-manager
```

#### `[policy_effect]` — Deny-Override Logic

```markdown
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

Translated:
  ALLOW  if: at least one matching policy says "allow"
         AND no matching policy says "deny"
  DENY   if: no matching allow rule
             OR at least one matching deny rule

This is deny-override. One deny beats 100 allows.

Example:
  User has: role:finance-manager  → policy: invoice/* allow
  User also: on sanctions list    → policy: * deny (explicit)

  Request: Enforce("tenant:usr", domain, "invoice/123", "read")
  Result:  DENIED — the deny rule wins
```

#### `[matchers]` — How a Request Is Evaluated

```markdown
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && keyMatch(r.act, p.act)

Step 1: g(r.sub, p.sub, r.dom)
  Does r.sub have the role p.sub in domain r.dom?
  Example: does "tenant:usr_001" have role "role:finance-manager" in "tenant-abc"?
  → Yes (from the g-rule) → match continues

Step 2: r.dom == p.dom
  Is the request domain the same as the policy domain?
  → Prevents cross-domain policy leakage
  → "tenant-abc" != "tenant-xyz" → no match (isolation enforced)

Step 3: keyMatch2(r.obj, p.obj)
  Does the requested object match the policy object pattern?
  Uses URL-style :param wildcards for path segments.
  keyMatch2("invoice/inv_123", "invoice/:id") → true
  keyMatch2("invoice/inv_123", "order/:id")   → false
  keyMatch2("invoice/inv_123", "*")           → true

Step 4: keyMatch(r.act, p.act)   ← keyMatch, NOT keyMatch2
  Does the requested action match the policy action pattern?
  Uses glob wildcards. keyMatch2 would break "*" action policies.
  keyMatch("read",   "read") → true
  keyMatch("read",   "*")    → true   (glob: * matches any verb)
  keyMatch("delete", "read") → false
```

### Pattern Matching Reference

**`keyMatch2`** is used for `obj` (resource objects) — URL-style `:param` path segments.

| Pattern | Request Object | Match? | Notes |
|---------|---------------|--------|-------|
| `invoice/:id` | `invoice/inv_123` | ✅ | Named path parameter |
| `invoice/*` | `invoice/inv_123` | ✅ | Glob wildcard also works |
| `invoice/*` | `invoice/inv_123/pdf` | ❌ | Does NOT match sub-paths |
| `*/export` | `invoice/export` | ✅ | Any resource, export action |
| `*` | `invoice/inv_123` | ✅ | Matches everything |
| `report/finance/*` | `report/finance/pnl` | ✅ | Scoped to finance reports |

**`keyMatch`** is used for `act` (action verbs) — glob wildcards only.

| Pattern | Request Action | Match? | Notes |
|---------|---------------|--------|-------|
| `read` | `read` | ✅ | Exact match |
| `*` | `read` | ✅ | Wildcard — any action |
| `*` | `delete` | ✅ | Wildcard — any action |
| `read` | `delete` | ❌ | No match |

### Policy Examples by Module

```markdown
FINANCE MODULE POLICIES:

Allow finance manager to do anything with invoices:
  p | role:finance-manager | {tenantID} | invoice/* | * | allow

Allow finance viewer to read invoices:
  p | role:finance-viewer  | {tenantID} | invoice/* | read | allow

Allow CFO to approve invoices over limit:
  p | role:cfo             | {tenantID} | invoice/*/approve | execute | allow

Block everyone from deleting closed-period journals:
  p | *                    | {tenantID} | journal/closed/* | delete | deny


SALES MODULE POLICIES:

Allow sales rep to create/read orders:
  p | role:sales-rep    | {tenantID} | order/*    | create | allow
  p | role:sales-rep    | {tenantID} | order/*    | read   | allow

Block sales rep from applying discounts > 10%:
  p | role:sales-rep    | {tenantID} | discount/high/* | create | deny

Allow sales manager all order operations:
  p | role:sales-manager | {tenantID} | order/*   | * | allow
  p | role:sales-manager | {tenantID} | discount/* | * | allow


PORTAL POLICIES:

Allow customer to view their own invoices:
  p | portal:cust-001  | {tenantID}:portal | invoice/* | read | allow

Block portal users from any internal resource:
  p | role:portal-user | {tenantID}:portal | internal/* | * | deny
```

### Role Inheritance Example

```markdown
ROLE HIERARCHY FOR FINANCE:

Roles and their parent roles (g-rules):
  role:finance-manager  inherits from  role:finance-viewer
  role:finance-viewer   inherits from  role:report-viewer

Policies assigned:
  role:report-viewer   → report/* read allow
  role:finance-viewer  → invoice/* read allow
  role:finance-viewer  → payment/* read allow
  role:finance-manager → invoice/* * allow
  role:finance-manager → payment/* * allow

User assignment:
  tenant:usr_cfo → role:finance-manager (in {tenantID})

Effective permissions for usr_cfo:
  → report/* read         (inherited from report-viewer)
  → invoice/* read        (inherited from finance-viewer)
  → payment/* read        (inherited from finance-viewer)
  → invoice/* *           (from finance-manager)
  → payment/* *           (from finance-manager)

Result: CFO can do everything with invoices and payments,
        and can read all reports.
```

---

Next: [Database Architecture](./06-database-architecture.md)
