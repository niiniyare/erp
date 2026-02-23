[<-- Back to Index](README.md)

## Troubleshooting Guide

### Issue 1: "Access Denied" When User Should Have Access

```markdown
SYMPTOM:
  Enforce() returns false.
  User believes they should have permission.

DIAGNOSIS STEPS:

Step 1: Verify the exact request being evaluated
  Print the Request struct before calling Enforce:
    log.Info("authz request", Fields{
        "subject": r.Subject,
        "domain":  r.Domain,
        "object":  r.Object,
        "action":  r.Action,
    })

  Common mistakes:
  → Subject has wrong prefix: "usr_001" instead of "tenant:usr_001"
  → Domain is wrong: using slug instead of UUID
  → Object path has wrong format: "invoices/123" instead of "invoice/123"

Step 2: Check what roles the user has
  roles, _ := svc.GetRoles(ctx, subject, domain)
  fmt.Println("Active roles:", roles)
  → If empty: user has no roles → assign the correct role
  → If has role: check if role has the required policy

Step 3: Check the policies for the domain
  policies, _ := svc.GetPolicies(ctx, domain)
  for _, p := range policies {
      fmt.Printf("%-30s %-15s %-20s [%s]\n", p.Subject, p.Action, p.Object, p.Effect)
  }
  → Look for a matching policy for the role
  → Look for a DENY rule that might be blocking

Step 4: Check if role is expired
  SELECT * FROM role_assignments
  WHERE subject = $1 AND domain = $2
  AND is_active = TRUE
  AND expires_at IS NOT NULL;
  → If expires_at < NOW(): role is expired → re-assign if intentional

Step 5: Check for deny rules
  SELECT * FROM casbin_rule
  WHERE ptype = 'p'
  AND v1 = $domain
  AND v4 = 'deny'
  AND v0 IN ('*', subject, role_of_user);
  → Any deny rule matching the request will override all allow rules
```

---

### Issue 2: Role Assigned But Enforce Still Returns False

```markdown
SYMPTOM:
  AssignRole() succeeded, HasRole() returns true,
  but Enforce() still returns false.

CAUSE: No policy assigned to the role.

DIAGNOSIS:
  HasRole() checks if the g-rule exists (user has the role).
  Enforce() requires BOTH:
  1. g-rule: user has the role ✓
  2. p-rule: role has permission on object ← MISSING

FIX:
  // Check if policy exists for this role:
  policies, _ := svc.GetPolicies(ctx, domain)
  hasPolicyForRole := false
  for _, p := range policies {
      if p.Subject == roleName && matchesObjectAction(p.Object, p.Action, requestedObj, requestedAct) {
          hasPolicyForRole = true
      }
  }

  // If missing, add it:
  svc.AddPolicy(ctx, authz.Policy{
      Subject: roleName,  // "role:finance-manager"
      Domain:  domain,
      Object:  "invoice/*",
      Action:  "*",
      Effect:  "allow",
  })
```

---

### Issue 3: In-Memory Model Out of Sync with Database

```markdown
SYMPTOM:
  casbin_rule has a policy in the DB (via direct SQL or migration).
  But Enforce() doesn't reflect the new rule.

CAUSE:
  The Casbin in-memory model is loaded at startup.
  Direct DB writes bypass the in-memory update.

FIX:
  // Force reload:
  err := svc.InvalidateCache(ctx)

  // Or via platform API (if no code access):
  POST /platform/authz/reload
  → Calls InvalidateCache() on the instance

PREVENTION:
  Always use svc.AddPolicy() / svc.AssignRole() instead of raw SQL.
  If bulk import is needed: import then call InvalidateCache().
  In multi-instance: InvalidateCache() must be called on ALL instances.
```

---

### Issue 4: Policy Works in One Tenant, Not Another

```markdown
SYMPTOM:
  role:finance-manager works in tenant-A but not tenant-B.

CAUSE: Policies are domain-scoped. They must be defined per-tenant.

DIAGNOSIS:
  // Check tenant A:
  policiesA, _ := svc.GetPolicies(ctx, authz.TenantDomain(tenantAID))
  // Check tenant B:
  policiesB, _ := svc.GetPolicies(ctx, authz.TenantDomain(tenantBID))
  // Compare — if tenant B has fewer policies, bootstrap is incomplete

FIX:
  Run the standard role policy bootstrap for tenant B.
  Or apply a policy template to tenant B.

PREVENTION:
  All new tenants should go through the same provisioning workflow
  that bootstraps default policies. Skipped provisioning steps → missing policies.
```

---

### Issue 5: Temporal Role Not Expiring

```markdown
SYMPTOM:
  Role should have expired (expires_at < NOW()) but user still has access.

CAUSE:
  Lazy expiry only triggers when Enforce() is called for that subject+domain.
  If the user hasn't made a request since expiry, the g-rule is still in Casbin's memory.

DIAGNOSIS:
  SELECT * FROM role_assignments
  WHERE subject = $1 AND domain = $2
  AND is_active = TRUE
  AND expires_at < NOW();
  → If rows exist: expiry hasn't been triggered yet (no request from this user)

FIX (manual trigger):
  svc.RevokeRole(ctx, subject, role, domain)
  → Marks is_active=FALSE and removes g-rule

OR: Make a test Enforce() call for the expired subject:
  svc.Enforce(ctx, authz.Request{Subject: subject, Domain: domain, Object:"*", Action:"*"})
  → revokeExpiredRoles runs → role revoked → next Enforce returns false

PREVENTION (Phase 2):
  Add background sweep job to catch users who never trigger lazy expiry.
```

---

### Issue 6: "authz.New: pool is required" on Startup

```markdown
SYMPTOM:
  Service constructor panics or returns error: "authz.New: pool is required"

CAUSE:
  Config.Pool is nil — DB pool not injected.

FIX:
  Ensure the pgxpool.Pool is created before authz.New() is called.
  Check dependency injection order.

  // Correct order:
  pool, err := pgxpool.New(ctx, dsn)
  if err != nil { log.Fatal(err) }

  authzSvc, err := authz.New(authz.Config{
      Pool:   pool,
      Logger: logger,
  })
```

---

### Issue 7: Cross-Domain Access Not Working (Intentional Cross-Grant)

```markdown
SYMPTOM:
  Deliberately granting a tenant user access to portal domain.
  But Enforce() in portal domain still returns false.

DIAGNOSIS:
  Check that AssignRole was called with the PORTAL domain, not the tenant domain:

  WRONG:
    svc.AssignRole(ctx, tenantID, "tenant:usr_admin", "role:portal-admin",
        authz.TenantDomain(tenantID))  // ← assigns role in TENANT domain

  CORRECT:
    svc.AssignRole(ctx, tenantID, "tenant:usr_admin", "role:portal-admin",
        authz.PortalDomain(tenantID))  // ← assigns role in PORTAL domain

  Then Enforce must be called with portal domain:
    svc.Enforce(ctx, authz.Request{
        Subject: "tenant:usr_admin",
        Domain:  authz.PortalDomain(tenantID),  // ← must match assignment domain
        Object:  "invoice/*",
        Action:  "read",
    })
```

---

### Issue 8: Slow Enforce in Production

```markdown
SYMPTOM:
  authz enforcement taking > 5ms in production.

POSSIBLE CAUSES:

Cause A: revokeExpiredRoles is slow
  Check: Are there many subjects with many expired roles?
  Check: Is idx_role_assignments_expires partial index present?
         \d role_assignments → look for idx_role_assignments_expires
  Fix:   CREATE INDEX if missing (see migration 000064)

Cause B: In-memory model is very large (> 500,000 rules)
  Check: SELECT COUNT(*) FROM casbin_rule;
  Fix:   Remove stale policies for archived tenants
         Implement policy count limits per domain

Cause C: InvalidateCache called too frequently
  Check: Log when InvalidateCache is called
  Fix:   Only call on explicit admin action or on startup

Cause D: SyncedEnforcer mutex contention (future improvement)
  Check: Goroutine profiling shows lock contention in casbin
  Fix:   Shard the enforcer by tenant domain (advanced optimization)

PROFILING:
  go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
  Look for authz.(*service).Enforce in the flamegraph
```

---

### Quick Diagnosis Checklist

```markdown
ENFORCE RETURNS FALSE — QUICK CHECKLIST:

□ Subject has correct prefix (platform:/tenant:/portal:/api:) ?
□ Domain is the tenant UUID (not slug, not name) ?
□ Object path matches policy pattern (keyMatch2 test) ?
□ Action matches exactly or policy uses wildcard "*" ?
□ User has a role assigned (GetRoles returns non-empty) ?
□ Role has a matching p-rule (GetPolicies shows the rule) ?
□ No DENY rule overriding the allow (check effect="deny" rows) ?
□ Role is not expired (check role_assignments.expires_at) ?
□ In-memory model is fresh (was InvalidateCache called after changes) ?
□ Correct domain used in both role assignment and Enforce call ?
```

---

Next: [Business Rules & Validation](./20-business-rules-and-validation.md)
