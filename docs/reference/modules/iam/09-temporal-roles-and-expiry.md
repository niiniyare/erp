[<-- Back to Index](README.md)

## Temporal Roles & Expiry

### What Are Temporal Roles?

A temporal role is a standard role assignment with an `expires_at` timestamp. When the timestamp passes, the role is automatically revoked on the next authorization check for that subject. No cron job, no scheduler, no external system required.

```markdown
TEMPORAL ROLE EXAMPLES:

External auditor — 30-day read access to last year's ledger:
  subject:  "tenant:usr_auditor_deloitte"
  role:     "role:auditor"
  domain:   "{tenantID}"
  expires:  2026-03-31 23:59:59 UTC

Contractor — 90-day write access to project tasks:
  subject:  "tenant:usr_contractor_xyz"
  role:     "role:project-contributor"
  domain:   "{tenantID}"
  expires:  2026-05-15 00:00:00 UTC

Temp cover — sales manager access while permanent manager is on leave:
  subject:  "tenant:usr_acting_manager"
  role:     "role:sales-manager"
  domain:   "{tenantID}"
  expires:  2026-03-10 08:00:00 EAT

Portal access — customer given read access for one month after dispute:
  subject:  "portal:cust_disputed"
  role:     "role:portal-extended"
  domain:   "{tenantID}:portal"
  expires:  2026-04-01 00:00:00 UTC
```

### Assigning a Temporal Role

```go
expiry := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)

err := svc.AssignRole(
    ctx,
    tenantID,
    "tenant:usr_auditor_deloitte",
    "role:auditor",
    authz.TenantDomain(tenantID),
    authz.WithExpiry(expiry),
    authz.WithAssignedBy("platform:admin-1"),
)
```

This writes:
```markdown
role_assignments row:
  subject:     "tenant:usr_auditor_deloitte"
  role_name:   "role:auditor"
  domain:      "{tenantID}"
  expires_at:  2026-03-31 23:59:59+00
  is_active:   TRUE
  assigned_by: "platform:admin-1"

casbin_rule g-row:
  ptype: g
  v0:    "tenant:usr_auditor_deloitte"
  v1:    "role:auditor"
  v2:    "{tenantID}"
```

### The Lazy Revoke Strategy

Temporal roles are not revoked by a background timer. They are revoked **lazily** — on the first `Enforce()` call for that subject after the expiry time has passed.

```markdown
LAZY REVOKE FLOW (inside service.Enforce):

Request: Enforce(ctx, Request{Subject: "tenant:usr_auditor", Domain: dom, ...})

Step 1: revokeExpiredRoles(ctx, subject, domain)
  Query:
    SELECT role_name FROM role_assignments
    WHERE subject=$1 AND domain=$2
      AND is_active = TRUE
      AND expires_at IS NOT NULL
      AND expires_at < NOW()

  Most calls → 0 rows → negligible overhead
                         (idx_role_assignments_expires is a partial index)

  If rows found (expiry happened) → for each expired role:
    1. UPDATE role_assignments SET is_active=FALSE
    2. enforcer.DeleteRoleForUserInDomain(sub, role, dom)
    → Role gone from in-memory enforcer

Step 2: enforcer.Enforce(sub, dom, obj, act)
  Now evaluates WITHOUT the expired role
  → Correctly returns false (no remaining allow rules)
```

### Why Lazy Revoke?

```markdown
ALTERNATIVE: Background cron job that sweeps expired roles every N minutes.

Problems with cron approach:
  → Requires a scheduler (Temporal/cron) just for cleanup
  → Race conditions: user makes request 5 seconds after expiry
    but sweep hasn't run yet → user still has access
  → Adds infrastructure complexity for a simple feature
  → Multiple app instances need coordination to avoid double-sweeps

LAZY REVOKE advantages:
  → Zero infrastructure: no scheduler needed
  → Exact-time revocation: expired role is gone on the VERY NEXT request
  → Self-cleaning: only sweeps the subject who triggered the request
  → Zero cost when no roles are expired (partial index → 0 rows in O(1))
  → Consistent with database truth: role_assignments is always authoritative
```

### Expiry Precision

Expiry is evaluated at the PostgreSQL server time (`NOW()`), not the Go application time. This eliminates clock drift between app server and DB.

```markdown
expires_at: 2026-03-31 23:59:59+03:00 (EAT, East Africa Time)

PostgreSQL stores this as UTC: 2026-03-31 20:59:59+00:00

When NOW() > 2026-03-31 20:59:59+00 → query returns this row → role revoked

GO TIP: Always store expires_at in UTC:
  expiry := time.Date(2026, 3, 31, 20, 59, 59, 0, time.UTC)
  // or convert from local:
  loc, _ := time.LoadLocation("Africa/Nairobi")
  localExpiry := time.Date(2026, 3, 31, 23, 59, 59, 0, loc)
  expiry := localExpiry.UTC()
```

### Checking Upcoming Expirations (UI / Notifications)

The `role_assignments` table with its indexed `expires_at` column makes it easy to build expiry notifications:

```sql
-- Roles expiring in the next 7 days (for tenant admin dashboard)
SELECT subject, role_name, expires_at, assigned_by
FROM role_assignments
WHERE tenant_id = current_tenant_id()
  AND is_active = TRUE
  AND expires_at IS NOT NULL
  AND expires_at BETWEEN NOW() AND NOW() + INTERVAL '7 days'
ORDER BY expires_at ASC;

-- All expired (but not yet cleaned up) roles
SELECT subject, role_name, expires_at
FROM role_assignments
WHERE tenant_id = current_tenant_id()
  AND is_active = TRUE
  AND expires_at < NOW();
```

### Phase 2: Background Sweep (Future)

When tenant count grows large, lazy revoke alone is sufficient for correctness but leaves stale g-rules in `casbin_rule` for subjects who never make another request (e.g., an API client that was decommissioned). A future background job will handle this:

```markdown
PLANNED BACKGROUND SWEEP:

Runs every hour (or daily for low-traffic systems):
  SELECT subject, role_name, domain FROM role_assignments
  WHERE is_active = TRUE
    AND expires_at < NOW();

  For each row:
    UPDATE role_assignments SET is_active = FALSE
    DELETE FROM casbin_rule WHERE ptype='g' AND v0=subject AND v1=role AND v2=domain

  Trigger enforcer.LoadPolicy() on all instances (via Redis pub/sub or DB flag)

Benefits:
  → Keeps casbin_rule clean
  → Reduces policy load time on restart
  → Useful for compliance: "no stale rules in the DB"

Phase 1 (current): lazy revoke only — sufficient for initial deployment
Phase 2 (future):  add background sweep when needed
```

### Temporal Role Audit Trail

Every expiry event is fully traceable:

```markdown
AUDIT QUERY: History of auditor_deloitte access

SELECT
    subject,
    role_name,
    expires_at,
    is_active,
    assigned_by,
    created_at,
    CASE
        WHEN NOT is_active AND expires_at < NOW() THEN 'EXPIRED'
        WHEN NOT is_active THEN 'MANUALLY_REVOKED'
        WHEN is_active AND expires_at > NOW() THEN 'ACTIVE'
        ELSE 'UNKNOWN'
    END as status
FROM role_assignments
WHERE subject = 'tenant:usr_auditor_deloitte'
ORDER BY created_at DESC;
```

---

Next: [Domain Isolation](./10-domain-isolation.md)
