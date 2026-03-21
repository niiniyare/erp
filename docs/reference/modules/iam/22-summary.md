[<-- Back to Index](README.md)

## Summary

### Module Recap

```markdown
WHAT WE COVERED:

Chapter  Topic                              Key Takeaway
─────────────────────────────────────────────────────────────────────────────
01       Executive Summary                  Security backbone — every request goes through here
02       Why Authorization in ERP           Centralized > scattered if/role checks
03       Architecture Overview              In-process, 8 Go files, 2 DB tables
04       Domain Model                       4 actor types × domain namespacing = full isolation
05       Casbin Policy Engine               RBAC + deny-override + keyMatch2 wildcard objects
06       Database Architecture              casbin_rule (enforcement) + role_assignments (audit)
07       Role Management                    AssignRole / RevokeRole / GetRoles — domain-scoped
08       Policy Management                  AddPolicy / RemovePolicy / GetPolicies
09       Temporal Roles & Expiry            expires_at + lazy revoke on Enforce() hot path
10       Domain Isolation                   r.dom == p.dom is the core isolation guarantee
11       Middleware & HTTP                  svc.Middleware("obj","act") — one line per route
12       System Module Integration          Tenant lifecycle, Settings flags, Entities, Audit
13       Business Module Integration        Finance and Sales resource path conventions
14       How Other Packages Use authz       API gateway, service layer, admin API, DI, testing
15       Workflow Integration               Temporal activities re-check authz per activity
16       Performance & Caching             < 1ms enforcement, in-memory model, indexed expiry
17       Security Considerations            8 threat vectors and their mitigations
18       Common Business Scenarios         7 real-world scenarios with full code examples
19       Troubleshooting Guide             8 common issues with diagnosis + fix steps
20       Business Rules & Validation       8 rules, validation table, idempotency guarantees
21       API Reference                     Complete Go interface and error type reference
```

### Architecture Summary

```markdown
AUTHZ MODULE ARCHITECTURE:

Go package: internal/core/authz/
├── Service (authz.go)
│   └── Single interface consumed by every module
│
├── Constructor (service.go)
│   ├── Loads Casbin model from casbinModel constant
│   ├── Creates pgxAdapter (persist.BatchAdapter)
│   ├── Calls casbin.NewEnforcer(model, adapter)
│   ├── EnableAutoSave(true) → DB + in-memory always in sync
│   └── Enforce / EnforceBatch / InvalidateCache
│
├── Role Engine (roles.go)
│   ├── AssignRole — pgx tx + Casbin g-rule
│   ├── RevokeRole — DB update + Casbin delete
│   ├── GetRoles / HasRole — in-memory (no DB)
│   ├── GetAssignments — DB query (metadata)
│   └── revokeExpiredRoles — lazy expiry on every Enforce
│
├── Policy Engine (policies.go)
│   ├── AddPolicy — validates + Casbin insert
│   ├── RemovePolicy — Casbin delete
│   └── GetPolicies — GetFilteredPolicy(1, domain)
│
├── HTTP (middleware.go)
│   └── Middleware(obj, act) → Fiber handler factory
│
└── Storage (adapter.go)
    └── pgxAdapter — persist.BatchAdapter backed by PostgreSQL

Database:
├── casbin_rule     — All p-rules and g-rules (Casbin-owned)
│   ├── UNIQUE index on (ptype,v0..v5) — no duplicates
│   ├── RLS: application_role has full access (domain in v1 provides isolation)
│   └── admin_role: unrestricted (platform engineering only)
│
└── role_assignments — Metadata for UI/audit/expiry
    ├── FK to tenants(id) CASCADE DELETE
    ├── Partial index on expires_at WHERE NOT NULL
    ├── RLS: application_role scoped by current_tenant_id()
    └── admin_role: unrestricted
```

### Key Design Decisions

```markdown
DESIGN CHOICES:

1. Casbin over custom engine
   → Replaces ~2,000 LOC with a battle-tested library
   → Formally verified policy semantics
   → Standard RBAC + deny-override + wildcard patterns

2. In-process (not microservice)
   → Sub-millisecond enforcement (no network hop)
   → Same binary → same transaction context
   → Simpler deployment and debugging

3. Domain namespacing for multi-tenancy
   → Each tenant has its own domain → policies never cross tenants
   → 4 actor types → distinct trust levels within tenants
   → "_platform_" reserved → operators completely isolated from tenants

4. Lazy expiry over scheduled jobs
   → Zero infrastructure for temporal roles
   → Exact-time revocation on first request after expiry
   → Partial DB index makes expiry check nearly free

5. Two-table design
   → casbin_rule: Casbin's source of truth (enforcement)
   → role_assignments: rich metadata (UI, audit, expiry, delegation)
   → Both maintained atomically by AssignRole/RevokeRole

6. Deny-override effect
   → ERP systems need hard blocks (sanctions, termination, closed periods)
   → One deny rule overrides all allows
   → Security-first design: explicit grant, not explicit block

7. Zero imports from abac/access/iam
   → Self-contained → independently deployable
   → No circular dependencies
   → Can be upgraded without affecting other modules
```

### Quick Reference

```markdown
ESSENTIAL CODE:

Constructor:
  svc, err := authz.New(authz.Config{Pool: pool, Logger: log})

Enforce:
  ok, err := svc.Enforce(ctx, authz.Request{
      Subject: authz.TenantSubject(userID),
      Domain:  authz.TenantDomain(tenantID),
      Object:  "invoice/inv_123",
      Action:  "read",
  })

Route protection:
  app.Get("/invoices/:id", svc.Middleware("invoice", "read"), handler)

Assign role:
  svc.AssignRole(ctx, tenantID, "tenant:usr_001", "role:finance-manager", dom)

Time-limited role:
  svc.AssignRole(ctx, tenantID, "tenant:usr_auditor", "role:auditor", dom,
      authz.WithExpiry(time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)),
      authz.WithAssignedBy("tenant:usr_cfo"),
  )

Blanket deny (termination):
  svc.RevokeRole(ctx, "tenant:usr_terminated", role, dom)
  svc.AddPolicy(ctx, authz.Policy{Subject:"tenant:usr_terminated",
      Domain:dom, Object:"*", Action:"*", Effect:"deny"})

Get what a user can do:
  roles, _ := svc.GetRoles(ctx, sub, dom)
  policies, _ := svc.GetPolicies(ctx, dom)

Reload after bulk change:
  svc.InvalidateCache(ctx)

Principal in HTTP context:
  p := c.Locals(authz.LocalsKeyPrincipal).(authz.Principal)
  // p.Subject = "tenant:usr_001"
  // p.Domain  = "a1b2c3d4-uuid"
```

### Performance Summary

```markdown
PERFORMANCE AT A GLANCE:

Enforce() (hot path):
  p99: < 1ms
  Components: 1 indexed DB query (expiry) + in-memory Casbin eval

EnforceBatch(N requests):
  p99: < 1ms per item (parallel eval in memory)

Policy load at startup:
  10,000 rules: < 200ms
  100,000 rules: < 1s

Policy change (AddPolicy/AssignRole):
  p99: < 10ms (DB write + in-memory update)

Memory per 1,000 tenants (50 policies each):
  ~25 MB — negligible
```

---

[Back to Index](README.md)
