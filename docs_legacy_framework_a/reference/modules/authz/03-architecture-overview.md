> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

[<-- Back to Index](README.md)

## Architecture Overview

### Clean Architecture Placement

The authz module sits in `internal/core/authz` — the domain layer. It has **no HTTP handlers of its own** (it provides Fiber middleware), **no repositories** (it owns its DB tables directly via pgx), and **no imports from other business modules**.

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP / Fiber Layer                        │
│  routes register svc.Middleware("invoice","read") per route  │
└────────────────────────────┬────────────────────────────────┘
                             │  fiber.Handler
┌────────────────────────────▼────────────────────────────────┐
│              authz.Service (interface)                       │
│  Enforce / EnforceBatch / AssignRole / AddPolicy / ...       │
│  Package: internal/core/authz                                │
└──────────┬──────────────────────────────────┬───────────────┘
           │ casbin.Enforcer                  │ pgxpool.Pool
┌──────────▼───────────┐       ┌─────────────▼───────────────┐
│   Casbin Engine       │       │   PostgreSQL                 │
│  (in-memory model)   │◄─────►│   casbin_rule               │
│  keyMatch2 + RBAC    │       │   role_assignments           │
└──────────────────────┘       └─────────────────────────────┘
```

### Module File Structure

```markdown
internal/core/authz/
├── authz.go        Service interface — the only file callers need to know
├── types.go        ActorType, Request, Policy, RoleAssignment, Principal
├── errors.go       Self-contained *Error type (AUTHZ_FORBIDDEN, etc.)
├── model.go        Casbin CONF string (casbinModel constant)
├── adapter.go      pgxAdapter — implements persist.BatchAdapter
├── service.go      New() constructor, Enforce, EnforceBatch, InvalidateCache
├── roles.go        AssignRole, RevokeRole, GetRoles, HasRole, revokeExpiredRoles
├── policies.go     AddPolicy, RemovePolicy, GetPolicies
└── middleware.go   Fiber handler factory (svc.Middleware)

db/migration/
├── 000063_authz_casbin_rule.up.sql    — casbin_rule table + RLS
├── 000063_authz_casbin_rule.down.sql
├── 000064_authz_role_assignments.up.sql  — role metadata + RLS
└── 000064_authz_role_assignments.down.sql
```

### Dependency Graph

```markdown
authz imports:
  ├── github.com/casbin/casbin/v2          (policy engine)
  ├── github.com/jackc/pgx/v5/pgxpool      (DB pool)
  ├── github.com/gofiber/fiber/v2           (HTTP middleware)
  ├── github.com/google/uuid                (ID generation)
  ├── internal/shared/logger                (structured logging)
  ├── internal/shared/metrics               (optional — latency histograms)
  └── internal/shared/tracing               (optional — distributed tracing)

authz does NOT import:
  ✗ internal/core/abac       (old custom ABAC engine)
  ✗ internal/core/access     (old access control)
  ✗ internal/core/iam        (old IAM tables)
  ✗ any other business module

Modules that import authz:
  ├── internal/core/finance   (invoice approve, GL access)
  ├── internal/core/sell      (order create, discount apply)
  ├── internal/core/hr        (payroll, employee records)
  ├── internal/core/inventory (stock adjust, PO create)
  └── cmd/api                 (gateway-level enforcement)
```

### Request Flow

Every protected HTTP request goes through this flow:

```markdown
HTTP Request arrives
       │
       ▼
Fiber Router matches route
       │
       ▼
authn middleware (JWT validation)
  → extracts sub, domain
  → stores Principal in c.Locals("authz_principal")
       │
       ▼
authz.Middleware("invoice","read") — per-route protection
  1. Read Principal from c.Locals
  2. Expand object: "invoice" + id param → "invoice/inv_123"
  3. Call service.Enforce(ctx, Request{sub, dom, obj, act})
       │
       ├── revokeExpiredRoles(sub, dom)   ← lazy expiry cleanup
       ├── enforcer.Enforce(sub,dom,obj,act)
       │     ├── check g-rules (role hierarchy)
       │     ├── match p-rules (keyMatch2 on obj, act)
       │     └── apply effect: some(allow) && !some(deny)
       └── return bool, error
  4. false → 403 Forbidden
  5. true  → c.Next()
       │
       ▼
Route handler executes
```

### Data Model Overview

```markdown
TWO TABLES:

casbin_rule — the Casbin policy store
  id    : UUID (surrogate key)
  ptype : "p" (policy) or "g" (role assignment)
  v0-v5 : the rule fields
           p: sub, dom, obj, act, eft
           g: user, role, domain

  Examples:
  p | tenant:user-1 | tenant-abc | invoice/* | read  | allow
  p | role:finance  | tenant-abc | invoice/* | *     | allow
  g | tenant:user-1 | role:finance | tenant-abc | | |

role_assignments — metadata for UI/audit/expiry
  id          : UUID
  tenant_id   : FK → tenants(id) CASCADE
  subject     : "tenant:user-uuid"
  role_name   : "role:finance-manager"
  domain      : "tenant-abc-uuid"
  assigned_by : "platform:admin-uuid"
  expires_at  : TIMESTAMPTZ (NULL = permanent)
  is_active   : BOOLEAN
  created_at  : TIMESTAMPTZ
```

### Design Decisions

```markdown
WHY IN-PROCESS (not microservice)?
  → Sub-millisecond enforcement (no network round-trip)
  → No additional infrastructure to maintain
  → Consistent transactional behavior with pgx pool
  → Simpler deployment and debugging

WHY CASBIN (not OPA / Cedar / custom)?
  → Casbin v2 is mature, widely deployed Go authorization library
  → Supports RBAC + deny-override + wildcard matching out of the box
  → The CONF model is declarative and readable
  → BatchEnforce enables pre-computation
  → Custom pgx adapter keeps persistence in-house (no Redis dependency)

WHY TWO TABLES (casbin_rule + role_assignments)?
  → casbin_rule is the source of truth for policy enforcement
  → role_assignments adds metadata Casbin doesn't store natively:
    expires_at, assigned_by, delegated_by, is_active, tenant FK
  → This lets the UI show "who assigned this role and when"
    without parsing Casbin's opaque v0-v5 columns

WHY DENY-OVERRIDE EFFECT?
  → ERP systems need hard blocks (sanctions, suspended accounts)
  → With deny-override: add one deny rule to block a user
    even if 10 other roles would allow the action
  → Security-first: explicit permission to act, not permission to block
```

---

Next: [Domain Model — 4 Actor Types](./04-domain-model.md)
