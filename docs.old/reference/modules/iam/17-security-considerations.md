[<-- Back to Index](README.md)

## Security Considerations

> **[IMPLEMENTED]** — describes the v1.0 security model as implemented.
> Items marked **[OPEN]** are known gaps that must be addressed before full production deployment.
> Items marked **[PLANNED]** are future-version features.

---

### Threat Model

| ID | Threat |
|---|---|
| T1 | Privilege escalation — tenant user gains platform rights |
| T2 | Cross-tenant access — Tenant A reads Tenant B's data |
| T3 | Stale authorization — ex-employee retains access after termination |
| T4 | Policy injection — attacker adds allow policies via API |
| T5 | Domain confusion — portal actor triggers tenant-domain policies |
| T6 | Denial-of-service — policy flooding exhausts memory/DB |
| T7 | Insider threat — platform admin abuses cross-tenant access |
| T8 | Race condition — role revoked during in-flight request |
| T9 | MFA code replay |
| T10 | Stale cached session — logged-out user retains Redis entry |

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

### T8: Race Conditions [IMPLEMENTED]

**Goroutine safety**: `casbin.SyncedEnforcer` is used in production. Concurrent `Enforce()` and `RevokeRole()` / `DeleteRoleForUserInDomain()` calls are goroutine-safe.

**In-flight request race**: The window between `Enforce()` returning `true` and the route handler completing is 1–10ms. A role revocation arriving during that window allows the handler to complete with its already-granted access. This is a known property of in-process authorization and is acceptable for ERP workloads.

For financial operations that require zero-race tolerance, re-check authorization inside the DB transaction (optimistic locking). This is not in scope for v1.0.

**Multi-instance convergence**: `StartAutoLoadPolicy(30s)` ensures all instances converge within 30 seconds of a policy change. The 30-second window is accepted for v1.0.

### T9: MFA Code Replay Prevention [IMPLEMENTED]

TOTP window tolerance is ±1 period (90 second total window). After each successful TOTP verify, `CheckAndMarkMFAReplay()` stores the window index in Redis for 90 seconds. A second use of the same code within the window is rejected as a replay.

MFA pending tokens are consumed atomically via Redis `GETDEL`. A second concurrent `CompleteMFALogin` call with the same pending token fails — only one session is ever created per MFA flow.

---

### T10: Stale Cached Session [IMPLEMENTED]

- `Logout()` calls `repo.Invalidate(hash)`, which synchronously deletes `"session:{hash}"` from Redis before returning. Logout is not cosmetic.
- `RevokeRole()` calls `SessionInvalidator.InvalidateByUser()`, which evicts all Redis session cache entries for the user.
- Cache TTL matches `session.expires_at − time.Now()`. A cached session cannot outlive its DB record.
- **Known gap**: Revoked API keys remain cached up to 5 minutes (the `"apikey:{hash}"` TTL). For emergency revocation, manually evict the Redis key or wait for TTL expiry.

---

### IAM-Level Threat Mitigations

| Threat | Key Mitigations | Status |
|---|---|---|
| Password brute-force | bcrypt; 5-attempt lockout (15min); generic error messages | Implemented |
| Session token theft | HttpOnly+Secure cookie; SHA-256 hash stored; synchronous Redis eviction on logout | Implemented |
| Privilege escalation | Casbin domain isolation (r.dom==p.dom); subject prefix from session, not request | Implemented; service-layer guard (AUTHZ-4) open |
| Cross-tenant data access | Casbin domain match + DB RLS + tenant_id in ctx — three independent layers | Implemented |
| MFA code replay | TOTP window tracking in Redis; GETDEL for pending token | Implemented |
| Policy injection | Effect validation; parameterised queries | Partial — domain write guard (AUTHZ-4) open |
| Policy flooding | Rate-limit on management endpoints | Policy count limit (AUTHZ-5) open |
| Stale session after logout | Synchronous Redis DELETE in Invalidate() | Implemented |
| Stale session after role revoke | InvalidateByUser() called from RevokeRole() | Implemented |
| Entity scope bypass | EntityScope from authenticated session only; never from request params | Implemented |
| API key stale-after-revoke | 5-min TTL window; document as known limitation | Known gap — acceptable for v1.0 |
| Password reset takeover | 32-byte token; 1h expiry; stored hashed; single-use | Implemented |
| Goroutine race in Casbin | SyncedEnforcer | Implemented |
| Multi-instance policy drift | StartAutoLoadPolicy(30s) | Implemented — 30s max drift |
| Security event audit log | OTel span attributes | Persistent audit log (AUTHZ-7) open |

### Security Checklist

```
DEPLOYMENT SECURITY CHECKLIST (v1.0):

Database:
  [x] casbin_rule — RLS present but set to full-access for application_role (required by Casbin)
  [x] role_assignments — RLS enforced per tenant_id
  [x] user_sessions — RLS enforced per tenant_id
  [x] users — RLS enforced per tenant_id
  [ ] Verify DB credentials are in secrets manager (not environment variables)
  [ ] Verify VPN / network policy restricts direct DB access

Application:
  [x] SyncedEnforcer — goroutine-safe concurrent enforcement
  [x] StartAutoLoadPolicy(30s) — multi-instance convergence
  [x] Logout() evicts Redis session cache synchronously
  [x] RevokeRole() calls InvalidateByUser() — session cache evicted
  [x] MFA pending tokens consumed atomically (GETDEL)
  [x] access/ module gated with //go:build ignore
  [ ] AUTHZ-4: platform domain write guard at service layer
  [ ] AUTHZ-5: policy count limit per domain (DoS prevention)
  [ ] AUTHZ-7: persistent security event audit log

Monitoring:
  [ ] Alert on deny spike (> threshold% deny rate for a domain)
  [ ] Alert on new _platform_ domain policy additions
  [ ] Alert on platform:* actor accessing sensitive resources
  [ ] Alert on ROLE_ASSIGNED with platform subject

Incident Response:
  Terminate employee:
    1. authzSvc.RevokeRole() for all roles (in-memory takes effect immediately)
    2. authzSvc.AddPolicy(deny, *, *) as defence-in-depth
    3. InvalidateByUser() called automatically by RevokeRole()

  API key compromised:
    1. APIKeyService.RevokeAPIKey(keyID)
    2. Wait ≤ 5 minutes for cache expiry
    3. Emergency: manually DELETE Redis key "apikey:{sha256(rawKey)}"

  Tenant suspended:
    Block at Authenticate middleware (check tenant status)
    Do NOT modify Casbin policies — tenant suspension is transient

  Policy store compromised:
    InvalidateCache() on all instances → audit casbin_rule → revoke affected rules
```

---

### Open Security Items (Known Gaps)

The following security items are acknowledged gaps in the v1.0 implementation. Each entry documents the risk, current mitigation, and planned remediation.

---

#### AUTHZ-4: Platform Domain Write Guard at Service Layer

**Status**: NOT IMPLEMENTED

**Risk**: The `AddPolicy` and `AssignRole` methods on `AuthzService` accept any domain string from the caller. A misconfigured Casbin policy or a bug in a handler could write a policy to the `_platform_` domain from a tenant-scoped request. This would not grant immediate access (tenant sessions produce tenant-domain principals), but it would pollute the platform domain's policy set and could be exploited if combined with other bugs.

**Current mitigation**: PostgreSQL RLS does not protect `casbin_rule` for `application_role` (Casbin requires unrestricted access to its table). The only boundary is the session's domain — tenant admins produce tenant-domain Principals, so any policy they create via the normal API path lands in their tenant domain. This relies on the handler correctly using `session.ToPrincipal().Domain`, not an arbitrary request body parameter.

**Planned remediation**: Add an explicit domain ownership check in `AddPolicy` and `AssignRole` at the service layer:
```go
if targetDomain == DomainPlatform && !caller.IsPlatform() {
    return ErrForbidden
}
```

---

#### AUTHZ-5: Policy Count Limit per Tenant

**Status**: NOT IMPLEMENTED

**Risk**: A tenant admin (or a compromised tenant admin account) could create an unbounded number of Casbin rules. Since the `SyncedEnforcer` loads all policies into memory, a very large policy set degrades enforcement performance for all tenants sharing the same process. It also increases DB load during the 30-second auto-reload cycle.

**Current mitigation**: Rate limiting on policy management API endpoints prevents rapid policy creation. The `casbin_rule` UNIQUE index prevents exact duplicate insertions. Operational monitoring of `casbin_rule` table row counts per domain.

**Planned remediation**: Add a configurable per-tenant policy rule limit check before `AddPolicy`:
```go
policies, _ := svc.GetPolicies(ctx, domain)
if len(policies) >= cfg.MaxPoliciesPerDomain {
    return ErrPolicyLimitExceeded
}
```
The limit should be configurable (e.g., via tenant settings) with a platform-enforced maximum.

---

#### DB-1: Unused `permissions` Column in `user_sessions` Table

**Status**: Column exists in schema; queries updated to not populate it; column not yet dropped

**Risk**: Low — the column is unused at runtime and carries no sensitive data in current sessions. The risk is developer confusion: new developers reading the schema may incorrectly assume the column is populated and attempt to read permissions from the session table, reintroducing the removed permissions-in-session anti-pattern.

**Current mitigation**: The `ResolvedSession` domain model has no `Permissions` field. The session repository does not read or write this column. Documentation explicitly states sessions carry no permission data (see `00-iam-overview.md`).

**Planned remediation**: Write and apply a migration to drop the column:
```sql
ALTER TABLE user_sessions DROP COLUMN IF EXISTS permissions;
```
Then re-run `sqlc generate` to remove the column from generated models. Track this as migration `000XXX_drop_session_permissions_column.up.sql`.

---

#### AUTHZ-7: Persistent Audit Log for Authorization Decisions

**Status**: OTel spans only — no persistent audit table for authorization decisions

**Risk**: Post-incident forensics rely on ephemeral trace data (OpenTelemetry spans). If the trace backend has a short retention policy or is unavailable during an incident, authorization decisions and policy changes cannot be reconstructed. This affects compliance posture (SOC 2, ISO 27001) and incident response capability.

**Current partial coverage**:
- `role_assignments` table provides a record of role assignment changes (who assigned what, when)
- `configuration_audit` table (migration `000601`) records configuration changes with full old/new value history
- OTel spans on `Enforce()`, `AddPolicy()`, `AssignRole()`, `RevokeRole()` provide real-time visibility

**Planned remediation**: Design and implement a dedicated `audit_events` table for security-sensitive IAM events:
```sql
CREATE TABLE audit_events (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id     UUID REFERENCES tenants(id),
  event_type    TEXT NOT NULL,    -- 'ROLE_ASSIGNED', 'ROLE_REVOKED', 'POLICY_ADDED', 'POLICY_REMOVED', 'ACCESS_DENIED', 'ACCESS_GRANTED'
  actor_subject TEXT NOT NULL,   -- who performed the action
  actor_domain  TEXT NOT NULL,
  target_subject TEXT,           -- who was affected (for role events)
  resource      TEXT,            -- object (for access events)
  action        TEXT,            -- verb (for access events)
  outcome       TEXT,            -- 'allowed' or 'denied'
  metadata      JSONB,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

This table should be immutable (no UPDATE or DELETE for application_role), append-only, with long retention for compliance evidence.

---

Next: [Common Business Scenarios](./18-common-business-scenarios.md)
