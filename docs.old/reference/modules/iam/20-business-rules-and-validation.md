[<-- Back to Index](README.md)

## Business Rules & Validation

### Rule 1: Subject Prefix Is Mandatory and Must Match Actor Type

```markdown
RULE: Every subject must have one of the four recognized prefixes.

Valid subjects:
  "platform:{any-non-empty-string}"
  "tenant:{any-non-empty-string}"
  "portal:{any-non-empty-string}"
  "api:{any-non-empty-string}"

Invalid subjects:
  "usr_001"              ← no prefix → rejected by ErrInvalidRequest
  ""                     ← empty → rejected
  "admin"                ← no prefix → rejected
  "platform:"            ← empty ID after prefix → rejected

WHY:
  The prefix is the machine-readable actor type.
  It drives domain defaulting in authn middleware.
  Without it, domain isolation cannot be guaranteed.

ENFORCEMENT:
  authz.Request validation: Subject == "" → ErrInvalidRequest
  Subject prefix validation is done by the authn middleware before authz.
  The authn middleware MUST enforce subject format.
```

---

### Rule 2: Domain Must Be Non-Empty and Must Match Request's Actor Type

```markdown
RULE: Every Enforce request must have a non-empty domain.

Valid domains:
  "_platform_"                                 (platform actors)
  "a1b2c3d4-5678-90ab-cdef-0123456789ab"      (tenant actors, UUID format)
  "a1b2c3d4-5678-90ab-cdef-0123456789ab:portal"  (portal actors)
  "a1b2c3d4-5678-90ab-cdef-0123456789ab:api"     (api actors)

Invalid domains:
  ""                                           ← empty → ErrInvalidRequest
  "savannah-electronics"                       ← slug, not UUID → will not match
  "a1b2c3d4-uuid:tenant"                       ← ":tenant" suffix not defined

CONSEQUENCES OF WRONG DOMAIN:
  Using tenant UUID slug instead of UUID → no policies match → always DENY
  Using platform domain for tenant request → policies don't match → DENY
  These fail safely (deny is safe), but cause confusing "access denied" in logs.
```

---

### Rule 3: Policy Effect Must Be Exactly "allow" or "deny"

```markdown
RULE: Policy.Effect must be the string "allow" or "deny" (lowercase).

Valid:
  Effect: "allow"
  Effect: "deny"

Invalid:
  Effect: "ALLOW"     ← uppercase → rejected in AddPolicy validation
  Effect: "permit"    ← unknown → rejected
  Effect: "block"     ← unknown → rejected
  Effect: ""          ← empty → rejected

WHY LOWERCASE:
  Casbin's effect evaluation compares: p.eft == allow
  Casbin stores and compares as-is. Uppercase "ALLOW" would not match.
  AddPolicy validation catches this before it reaches Casbin.

ENFORCEMENT:
  In AddPolicy(): if Effect != "allow" && Effect != "deny" → return error
```

---

### Rule 4: Role Names Should Follow the role: Prefix Convention

```markdown
RULE (convention, not enforced by code): All role names should start with "role:".

Recommended:
  "role:finance-manager"
  "role:sales-rep"
  "role:tenant-admin"
  "role:portal-customer"
  "role:api-readonly"

Technically valid but not recommended:
  "finance-manager"   ← no prefix, works but harder to distinguish from user subjects
  "FINANCE_MANAGER"   ← uppercase, inconsistent
  "Finance Manager"   ← spaces, will cause problems in some query contexts

WHY:
  The "role:" prefix makes it visually clear in casbin_rule rows
  whether v0 is a user subject or a role.
  It also prevents accidental policy assignment to a role that looks like a user.
```

---

### Rule 5: Deny Rules Are Permanent Overrides — Use with Care

```markdown
RULE: A deny policy with eft="deny" cannot be overcome by any number of allow rules.

Example:
  p | role:finance-manager | {dom} | invoice/* | * | allow   ← 100 allows
  p | tenant:usr_001       | {dom} | *          | * | deny    ← 1 deny

  Enforce({Subject:"tenant:usr_001", ...}) → DENY
  (usr_001 has role:finance-manager which has allow, but the deny wins)

WHEN TO USE DENY RULES:
  ✓ Terminated employees: add deny on * * (blanket)
  ✓ Sanctioned entities: add deny to block specific actions
  ✓ Closed periods: deny journal/closed/* create/update
  ✓ Feature not available on plan: deny feature/* *

WHEN NOT TO USE DENY RULES:
  ✗ Simply "not granting" a permission — just don't add the allow rule
  ✗ Scoping a user to a subset of resources — use specific object paths instead
  ✗ As a way to "undo" a role grant — revoke the role instead

CLEANUP:
  When a terminated employee is re-hired:
    RemovePolicy(Policy{Subject:"tenant:usr_001", Object:"*", Action:"*", Effect:"deny"})
    Then re-assign their new role.
    The blanket deny must be explicitly removed — it doesn't auto-expire.

  EXCEPTION: If the deny policy was set with a specific subject and
  the employee gets a NEW user ID on re-hire, the deny rule for the old ID
  is harmless (old ID has no JWT anyway).
```

---

### Rule 6: Role Assignments Are Idempotent

```markdown
RULE: Calling AssignRole with the same (subject, role, domain) multiple times
is safe and idempotent.

Behavior:
  First call: INSERT into role_assignments, add g-rule to Casbin
  Subsequent calls (same params):
    → ON CONFLICT (subject, role_name, domain) DO UPDATE
    → Updates assigned_by, expires_at, is_active=TRUE
    → Re-adds g-rule to Casbin (AddRoleForUserInDomain is idempotent)

WHY IMPORTANT:
  Provisioning workflows may call AssignRole multiple times if retried.
  HR systems may sync role assignments periodically.
  Both are safe — no duplicate g-rules are created.
```

---

### Rule 7: Platform Domain Policies Never Apply to Tenant Requests

```markdown
RULE: A policy in domain "_platform_" is evaluated ONLY when the request
has domain = "_platform_". It never matches tenant-domain requests.

This is guaranteed by the Casbin matcher:
  m = ... && r.dom == p.dom && ...

  r.dom = "a1b2c3d4-tenant-uuid"
  p.dom = "_platform_"
  "a1b2c3d4-tenant-uuid" == "_platform_" → FALSE → rule does not apply

IMPLICATION:
  A policy like: p | role:platform-admin | _platform_ | * | * | allow
  Does NOT give platform-admin rights within any tenant domain.
  Platform admins need separate role assignments in each tenant domain
  if they need to act as tenant users.
```

---

### Rule 8: The role_assignments Table Is the Authoritative Audit Source

```markdown
RULE: For UI display, expiry tracking, and audit reporting,
use role_assignments. For enforcement, casbin_rule is authoritative.

The two tables should be consistent:
  Every active row in role_assignments → matching g-rule in casbin_rule
  Every g-rule in casbin_rule (for tenant actors) → matching row in role_assignments

Inconsistency can happen if:
  → DB transaction committed role_assignments but Casbin call failed
  → Direct SQL deletion of casbin_rule without updating role_assignments

RECONCILIATION:
  SELECT ra.subject, ra.role_name, ra.domain
  FROM role_assignments ra
  LEFT JOIN casbin_rule cr
    ON cr.ptype='g' AND cr.v0=ra.subject AND cr.v1=ra.role_name AND cr.v2=ra.domain
  WHERE ra.is_active = TRUE
    AND cr.id IS NULL;
  → Rows found: role_assignments says active but casbin_rule is missing
  → Fix: re-add g-rule or mark role_assignments as inactive

  SELECT cr.v0, cr.v1, cr.v2
  FROM casbin_rule cr
  WHERE cr.ptype = 'g'
    AND NOT EXISTS (
        SELECT 1 FROM role_assignments ra
        WHERE ra.subject=cr.v0 AND ra.role_name=cr.v1 AND ra.domain=cr.v2
          AND ra.is_active=TRUE
    )
    AND cr.v2 != '_platform_';
  → Orphan g-rules in casbin_rule → should be deleted
```

---

### Validation Reference Table

| Parameter | Valid Values | Notes |
|-----------|-------------|-------|
| Subject prefix | `platform:`, `tenant:`, `portal:`, `api:` | Required, enforced by authn |
| Domain format | UUID, `UUID:portal`, `UUID:api`, `_platform_` | UUID is tenant's primary key |
| Object | Any non-empty string | `keyMatch2` patterns recommended |
| Action | Any non-empty string | Conventions: `read`, `create`, `update`, `delete`, `approve`, `export`, `execute`, `*` |
| Effect | `"allow"` or `"deny"` | Validated in AddPolicy |
| Role name | Any string, `role:` prefix recommended | Convention only |
| Expiry | Any future `time.Time`, or nil | nil = permanent |

---

Next: [API Reference](./21-api-reference.md)
