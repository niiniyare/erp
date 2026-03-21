[<-- Back to Index](README.md)

## Security Considerations

### Threat Model

The authz module must resist the following threat categories:

```markdown
THREAT MODEL:

T1: Privilege escalation — tenant user gains platform rights
T2: Cross-tenant access — Tenant A reads Tenant B's data
T3: Stale authorization — ex-employee retains access after termination
T4: Policy injection — attacker adds allow policies via API
T5: Domain confusion — portal actor triggers tenant-domain policies
T6: Denial-of-service — policy flooding exhausts memory/DB
T7: Insider threat — platform admin abuses cross-tenant access
T8: Race condition — role revoked during in-flight request
```

### T1: Privilege Escalation Prevention

```markdown
HOW IT'S PREVENTED:

1. Domain separation in the Casbin matcher:
   r.dom == p.dom is a hard check
   "tenant:usr_hacker" cannot trigger "_platform_" policies
   even if they somehow forge a request

2. Subject prefix isolation:
   JWT claims encode the actor type ("platform", "tenant", etc.)
   JWT is signed — attacker cannot change actor type
   Even if they could: "tenant:forged" has no platform-domain rules

3. No privilege inheritance across domains:
   Casbin g-rules are domain-scoped: g(user, role, domain)
   A role in tenant domain ≠ same role in platform domain
   role:tenant-admin in {tenantID} ≠ role:tenant-admin in _platform_

AUDIT CHECK:
  Regularly query: platform-domain g-rules to ensure only known
  platform service accounts have platform roles.

  SELECT v0, v1 FROM casbin_rule
  WHERE ptype='g' AND v2='_platform_'
  → Should only show known platform:* subjects
```

### T2: Cross-Tenant Isolation

```markdown
HOW IT'S PREVENTED:

1. Casbin matcher: r.dom == p.dom
   Tenant A's domain = "a1b2c3d4-uuid"
   Tenant B's domain = "c5d6e7f8-uuid"
   A p-rule in Tenant A's domain CANNOT match a request in Tenant B's domain.

2. role_assignments RLS:
   application_role can only see role_assignments for current_tenant_id()
   Tenant B's role assignments are invisible to Tenant A's connection

3. No shared policies:
   There is no mechanism to create a policy that applies to multiple tenants.
   Cross-tenant access requires explicit platform-domain policies.

TESTING CROSS-TENANT ISOLATION:
  // Integration test that must pass before every release:
  tenantA := provision("tenant-a")
  tenantB := provision("tenant-b")

  // Add policy only for Tenant A
  svcA.AddPolicy(ctx, Policy{Subject: "role:viewer", Domain: tenantA.Domain, Object: "invoice/*", Action: "read", Effect: "allow"})
  svcA.AssignRole(ctx, tenantA.ID, "tenant:usr_a", "role:viewer", tenantA.Domain)

  // Enforce as Tenant A user in Tenant B domain
  ok, _ := svc.Enforce(ctx, Request{Subject: "tenant:usr_a", Domain: tenantB.Domain, Object: "invoice/123", Action: "read"})
  assert.False(t, ok, "Tenant A policy must NOT match in Tenant B domain")
```

### T3: Stale Authorization (Terminated Employees)

```markdown
HOW IT'S MITIGATED:

IMMEDIATE REVOCATION:
  1. HR/Admin calls RevokeRole for all roles
  2. In-memory model updated immediately
  3. Next request: DENY
  Total window: 0-1 seconds (in-process, no network)

DEFENSE IN DEPTH:
  After revoking roles, add explicit deny:
  AddPolicy(Policy{Subject: "tenant:usr_terminated", Domain: dom,
                   Object: "*", Action: "*", Effect: "deny"})
  → Even if a role was missed, deny rule blocks everything

TOKEN EXPIRY:
  JWTs expire (typically 15 minutes to 24 hours).
  Even if authz is somehow not called, the JWT becomes invalid.
  Two-layer protection: JWT expiry + authz revocation.

TEMPORAL ROLES:
  Contractors and time-limited users have expires_at set.
  Automatic revocation on first request after expiry.
  No manual cleanup needed.
```

### T4: Policy Injection Prevention

```markdown
HOW IT'S PREVENTED:

1. AddPolicy is behind authz itself:
   app.Post("/policies", svc.Middleware("policy", "create"), addPolicyHandler)
   → Only actors with "policy create" permission can add policies
   → This permission is given only to tenant-admin role

2. Input validation in AddPolicy():
   Effect must be "allow" or "deny" (no arbitrary strings)
   Subject/Domain/Object/Action must be non-empty
   No SQL injection possible — parameterized queries via pgx

3. Domain scoping by authn middleware:
   Tenant admin can only create policies in their own domain
   The authn middleware extracts domain from JWT — not from request body

4. Platform policies require platform-domain JWT:
   Tenant admin cannot create "_platform_" policies
   (Their JWT domain is their tenant UUID, not "_platform_")

WHAT ABOUT DIRECT SQL?
   Only admin_role has unrestricted access to casbin_rule
   application_role also has full access (required for Casbin adapter)
   → DB access must be protected at the infrastructure level
   → VPN / IAM controls on DB access
   → After any manual SQL change: InvalidateCache() required
```

### T5: Domain Confusion

```markdown
HOW IT'S PREVENTED:

Domain is extracted from the JWT by the authn middleware:
  Actor type "portal" → domain always = "{tenantID}:portal"
  Actor type "api"    → domain always = "{tenantID}:api"
  Actor type "tenant" → domain always = "{tenantID}"
  Actor type "platform" → domain always = "_platform_"

The Principal stored in c.Locals is set by the authn middleware,
not by any user-provided input.

Even if an attacker sends a modified X-Domain header:
  → The authn middleware ignores it
  → Domain comes from JWT claims only
  → JWT is server-signed

CONCLUSION: Domain confusion is not possible without JWT forgery.
```

### T6: Denial of Service (Policy Flooding)

```markdown
MITIGATIONS:

Rate limiting on policy management endpoints:
  app.Post("/policies", rateLimiter(10, time.Minute), svc.Middleware("policy","create"), ...)
  → Tenant admin cannot add more than 10 policies/minute

Policy count limits (recommended governance):
  Before AddPolicy, check current policy count for domain:
    policies, _ := svc.GetPolicies(ctx, domain)
    if len(policies) > 10_000 {
        return ErrPolicyLimitExceeded
    }

Memory monitoring:
  Alert if authz_policy_count{domain=...} > threshold
  Alert if process memory grows unexpectedly after policy operations

Database protection:
  casbin_rule UNIQUE index prevents exact duplicate insertions
  (ON CONFLICT DO NOTHING)
  Does NOT prevent 10,000 similar-but-different rules — count limit needed
```

### T7: Insider Threat (Platform Admin Audit)

```markdown
CONTROLS:

Every platform-domain operation is auditable:
  Query audit log for:
    EventType: "ACCESS_DENIED" AND Subject: "platform:*"
    EventType: "ROLE_ASSIGNED" AND Subject: "platform:*"
    EventType: "POLICY_ADDED" AND Domain: "_platform_"

role_assignments records every platform assignment:
  SELECT * FROM role_assignments
  WHERE subject LIKE 'platform:%'
  AND created_at > NOW() - INTERVAL '30 days';

Separation of duties for platform admins:
  role:platform-admin  → can manage ALL tenants
  role:platform-support → can READ tenant data only
  role:platform-billing → can manage plans only
  Avoid giving role:platform-admin to all staff

Real-time alerts:
  Alert: any platform:* role assignment to unknown subject
  Alert: any new p-rule added to _platform_ domain
  Alert: platform:* access to sensitive resources (payroll/*, salary/*)
```

### T8: Race Conditions

```markdown
RACE CONDITION: Role revoked while request is in-flight

Timeline:
  T=0ms: Request arrives, authn passes, authz Enforce() called
  T=1ms: Enforce() reads in-memory model → role exists → ALLOW
  T=2ms: Admin revokes role (RevokeRole called on another goroutine)
  T=3ms: Route handler executes → user performs action
  → User completed the action despite revocation

ANALYSIS:
  This is a known property of in-process authorization.
  The window is 1-10ms (time between Enforce and handler completion).
  This is ACCEPTABLE for an ERP system.

  For financial operations requiring zero-race tolerance:
    → Re-check inside the DB transaction (optimistic locking)
    → SELECT ... WHERE user_has_active_role(subject, role, domain)
      (custom DB function, not in current scope)

  For standard operations:
    → 1-10ms window is negligible
    → Next request will correctly return DENY
    → JWT expires within hours — outer bound on any stale access

CASBIN THREAD SAFETY:
  casbin.Enforcer is NOT goroutine-safe by default.
  The authz module should use casbin.SyncedEnforcer for production:

  // In service.go (future improvement):
  e, err := casbin.NewSyncedEnforcer(m, adapter)
  e.StartAutoLoadPolicy(30 * time.Second)  // periodic reload

  OR use a sync.RWMutex wrapper around Enforce() calls.
  Current implementation: single instance, acceptable for Phase 1.
```

### IAM-Level Threat Mitigations

| Threat | Key Mitigations |
|---|---|
| Password brute-force | bcrypt cost 12 (~4 guesses/sec); lockout after 5 fails (15m→30m→1h→2h backoff); 10/min/IP rate limit; HIBP top-10k check; generic errors (no user enumeration) |
| Session token theft | HttpOnly+Secure+SameSite=Lax; 24h absolute TTL; SHA-256 hash stored, not plaintext; instant deletion on logout/suspend |
| Privilege escalation | Guard 1 (namespace) + Guard 2 (tenant scope) + Guard 3 (delegation) — all atomic, all audited |
| Cross-tenant data access | Casbin `r.dom==p.dom` + DB RLS + service-layer tenantID — three independent layers |
| MFA code replay | Used codes cached 90s; one code valid once per 30-second window |
| Flag/setting manipulation | Require explicit `settings.*` permissions; system flags require `platform.*`; all changes audit-logged + session invalidated |
| Entity scope bypass | `entity_scope` from authenticated session only — never from request params; DB WHERE uses session's path prefix |
| Casbin rule injection | Validates subject prefix, domain ownership; cannot write `_platform_` from tenant JWT; all additions audit-logged |
| Account takeover via reset | 32-byte token (256-bit entropy); 1h expiry; stored hashed; one-time use; all sessions invalidated on success |

### Security Checklist

```markdown
DEPLOYMENT SECURITY CHECKLIST:

Database:
  ✓ casbin_rule RLS enabled
  ✓ role_assignments RLS enabled
  ✓ Only application_role and admin_role have access
  ✓ DB credentials rotated regularly
  ✓ VPN / network policy restricts direct DB access

Application:
  ✓ All routes protected with authn + authz middleware
  ✓ Default deny: routes with no authz middleware are explicitly reviewed
  ✓ Policy management endpoints rate-limited
  ✓ Policy count limits per domain enforced
  ✓ InvalidateCache called after bulk DB changes

Monitoring:
  ✓ Alert on unusual deny spike (> X% deny rate for a domain)
  ✓ Alert on new platform-domain policy additions
  ✓ Alert on platform actor accessing sensitive resources
  ✓ Log every ROLE_ASSIGNED, ROLE_REVOKED, POLICY_ADDED, POLICY_REMOVED

Incident Response:
  ✓ Terminate employee: RevokeRole + AddPolicy(deny) within 5 minutes
  ✓ API key compromised: RevokeRole api:cli_{key} + AddPolicy(deny)
  ✓ Tenant suspended: Block at authn layer (don't modify authz policies)
  ✓ Security breach: InvalidateCache + policy audit on all instances
```

---

Next: [Common Business Scenarios](./18-common-business-scenarios.md)
