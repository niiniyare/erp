[<-- Back to Index](README.md)

## Performance & Caching

### Full Request Latency Profile

```
Session validation:     ~1ms      indexed UPDATE on token_hash
Permission check:       <0.001ms  map lookup — nanoseconds
Flag check:             <0.001ms  map lookup
Setting read:           <0.001ms  map lookup + parse
Entity scope check:     0ms       struct field access
Casbin Enforce():       ~0.1ms    in-memory model (management ops only)

Total auth+config overhead per request: ~1–2ms
```

### Login Session Build (One-Time Cost)

| Operation | Cost (all run concurrently) |
|---|---|
| ComputePermissions | ~2ms |
| ResolveEntityScope | ~1ms |
| FlagService.Resolve | ~1ms |
| SettingService.Resolve | ~1ms |
| UserPreferences | ~0.5ms |
| **Total at login** | **~5–6ms** |

No further DB hits for auth, flags, or settings for the 24-hour session lifetime.

### Redis Cache Keys

```
user:<uuid>                      5-minute TTL
session:<token_hash>             TTL = session.expires_at - now()
roles:<user_id>:<domain>         5-minute TTL; invalidated on role change
permissions:<user_id>            5-minute TTL; invalidated on sync
flags:tenant:<tenant_id>         10-minute TTL; invalidated on flag change
settings:tenant:<tenant_id>      10-minute TTL; invalidated on setting change
prefs:user:<user_id>             30-minute TTL; invalidated on preference update
entity:<uuid>                    10-minute TTL (changes rarely)
nav:tenant:<tenant_id>           5-minute TTL; invalidated on flag change
```

### Session Method Benchmarks

```go
func BenchmarkSessionCan(b *testing.B) {
    session := buildTestSession(b, 200)
    b.ResetTimer()
    for i := 0; i < b.N; i++ { session.Can("finance.transactions", "post") }
    // Expected: < 100ns
}

func BenchmarkSessionFeatureEnabled(b *testing.B) {
    session := buildTestSession(b, 200)
    b.ResetTimer()
    for i := 0; i < b.N; i++ { session.FeatureEnabled("finance.transactions.approval_workflow") }
    // Expected: < 200ns (two map lookups)
}

func BenchmarkSessionBuildAtLogin(b *testing.B) {
    for i := 0; i < b.N; i++ { iamSvc.buildSession(ctx, testUser) }
    // Expected: < 10ms (five concurrent queries)
}
```

### Performance Is a First-Class Concern

Authorization is on the hot path of every request. If enforcement is slow, every user experiences that slowness. The authz module uses a two-layer strategy to stay sub-millisecond.

```markdown
LATENCY BUDGET:

Total request budget (typical ERP endpoint): 100-200ms
  ├── Network / TLS:            10-20ms
  ├── Authenticate middleware:   1-3ms   (session token → DB lookup → Redis hit)
  ├── authz fast path:        < 0.1ms   (O(1) map lookup in ResolvedSession)
  ├── authz Casbin path:      < 1ms     (management ops only, in-memory)
  ├── DB query (with index):   10-30ms
  ├── Business logic:          10-50ms
  └── Response serialization:   5-10ms

authz must NOT be the slow part.
Target: p50 = 0.2ms, p95 = 0.5ms, p99 = 1ms
```

### Two-Layer Permission Architecture

```markdown
LAYER 1 — Pre-computed map (request hot path):
  Login time:
    Casbin.GetPolicies(domain) + Casbin.GetRoles(subject, domain)
    → build map[string]bool {"finance.invoices.read": true, ...}
    → store in user_sessions.permissions JSONB column

  Every request:
    ValidateSession() reads session row from Redis (cache hit) or DB
    → ResolvedSession.Permissions map already populated
    → Authorize middleware: sess.Can("finance.invoices.read")  ← O(1)
    → Zero DB hits, zero Casbin calls for permission checks

LAYER 2 — Casbin engine (source of truth / management):
  Used for:
    → Computing the permission map at login (once per session)
    → Management API: "what roles does user X have?"
    → Admin-driven InvalidateCache() after role changes
    → Direct svc.Enforce() in non-HTTP contexts (Temporal workflows, jobs)

SESSION CACHE (internal/platform/cache):
  ValidateSession() uses cache.Service to avoid DB hit on every request:
    cache key: "session:{sha256(token)}"  TTL = session.ExpiresAt - now
    Hit:  return cached ResolvedSession (no DB)
    Miss: query user_sessions → cache result → return

  Role/permission change → call svc.Logout(token) to invalidate session
  or call cache.Delete("session:{hash}") directly
```

### The In-Memory Cache

Casbin holds the complete policy set in memory, loaded from PostgreSQL at startup. The `Enforce()` call never hits the database in steady state — it's a pure in-memory computation.

```markdown
ENFORCEMENT DATA FLOW:

Startup:
  casbin.NewEnforcer(model, pgxAdapter)
  → adapter.LoadPolicy() runs once
  → SELECT ptype, v0..v5 FROM casbin_rule → loads all rows into memory
  → In-memory model built: p-rules and g-rule graphs

Every Enforce() call:
  1. revokeExpiredRoles() — ONE indexed DB query (usually returns 0 rows)
  2. enforcer.Enforce() — pure in-memory evaluation
     → no DB round-trip
     → no Redis lookup
     → no network call
  3. Result returned

Total time: typically 0.1 - 0.5ms
```

### Policy Cache Size Analysis

```markdown
CACHE SIZING:

Example: 1,000 tenants, 50 roles per tenant, 20 policy rules per role

p-rules: 1,000 × 50 × 20 = 1,000,000 policy rows
g-rules: 1,000 × 1,000 users × 2 roles avg = 2,000,000 role rows
Total: ~3,000,000 rows

Memory per row: ~200 bytes (6 strings, avg 30 chars each)
Total memory: 3M × 200 = 600 MB

OPTIMIZATION FOR SCALE:
  Most tenants share the same role structure.
  Instead of 1,000 × 50 role policies, use shared role templates:

  SYSTEM-LEVEL ROLE POLICIES (stored once in platform domain):
    p | role:finance-manager | _platform_ | invoice/* | * | allow
    ↑ This pattern does NOT work directly — Casbin requires domain match

  BETTER APPROACH — policy templates:
    Platform admin defines a template JSON
    On tenant provision: template is expanded into tenant-specific casbin_rule rows
    1,000 tenants × 50 policies = 50,000 rows (not 1M)
    Users are assigned to roles → g-rules handle the rest

PRACTICAL NUMBERS FOR AWO ERP (initial deployment):
  50 tenants × 50 policies × 50 users = 125,000 rows
  Memory: 125,000 × 200 bytes = 25 MB — negligible
```

### Auto-Save Behavior

`EnableAutoSave(true)` means every `AddPolicy`, `RemovePolicy`, `AddRoleForUserInDomain`, and `DeleteRoleForUserInDomain` call writes to the database and updates the in-memory model atomically.

```markdown
AUTO-SAVE WRITE FLOW:

svc.AddPolicy(ctx, policy)
  →  enforcer.AddPolicy(...)      [Casbin]
     → adapter.AddPolicy(...)     [pgx INSERT ON CONFLICT DO NOTHING]
     → in-memory model updated    [immediate]
  → Return to caller

Result: DB and in-memory are always consistent after each call.
        No explicit flush needed.
        Cache is never stale relative to the DB.

EXCEPTION: Direct SQL writes to casbin_rule (migrations, bulk imports)
           → enforcer does NOT know about these
           → Must call InvalidateCache() after
```

### The Lazy Expiry Query Performance

The expiry check runs on every `Enforce()` call. Here's why it's fast:

```markdown
QUERY:
  SELECT role_name FROM role_assignments
  WHERE subject=$1 AND domain=$2
    AND is_active = TRUE
    AND expires_at IS NOT NULL
    AND expires_at < NOW()

INDEX USED: idx_role_assignments_expires
  CREATE INDEX idx_role_assignments_expires
  ON role_assignments(expires_at)
  WHERE expires_at IS NOT NULL;

This is a PARTIAL INDEX — only indexes rows with non-null expires_at.
In a typical deployment:
  → 95% of role assignments are permanent (expires_at = NULL)
  → Only ~5% have expiry
  → Partial index is tiny (5% of all rows)
  → Filtered by subject + domain (idx_role_assignments_subject)
  → For most subjects: 0 rows returned → near-zero I/O

COMBINED QUERY PLAN:
  Index Scan on idx_role_assignments_subject (subject=$1, domain=$2)
  → Filter: expires_at IS NOT NULL AND expires_at < NOW()
  → Rows: 0 (typical) → scan stops immediately

MEASUREMENT:
  PostgreSQL EXPLAIN ANALYZE shows < 0.1ms for this query (0 rows case)
  Even with 10 expired roles: < 0.5ms (rarely happens)
```

### Multi-Instance Deployment

When multiple app instances run (Kubernetes pods), each has its own in-memory Casbin enforcer loaded at startup. Policy changes made on one instance must propagate to others.

```markdown
CURRENT BEHAVIOR (Phase 1):

Instance A: AddPolicy(new rule)
  → Writes to PostgreSQL casbin_rule
  → Updates Instance A's in-memory model

Instance B: NOT updated yet
  → Still has old in-memory model
  → Will NOT see the new rule until:
     a) Instance B calls InvalidateCache()
     b) Instance B restarts

ACCEPTABLE FOR INITIAL DEPLOYMENT:
  → Most policy changes are low-frequency (admin operations)
  → A short window of inconsistency (seconds to minutes) is acceptable
  → Critical denials (sanctions, termination) should call InvalidateCache
    on all instances via a broadcast mechanism

PRODUCTION-READY (Phase 2) — Redis pub/sub:
  On any policy change:
    Publish to Redis channel: "authz:policy-changed:{tenantID}"

  Each instance subscribes:
    go func() {
        for range pubsub.Channel() {
            svc.InvalidateCache(context.Background())
        }
    }()

  Effect: Policy changes propagate to all instances within 50-100ms.

ALTERNATIVES:
  → Polling: each instance calls LoadPolicy() every 30 seconds
    (acceptable for low-security environments, adds DB load)
  → Shared enforcer via distributed lock (complex, not recommended)
  → Per-request LoadPolicy (catastrophic for performance — never do this)
```

### Performance Benchmarks

```markdown
BENCHMARK TARGETS:

Enforce() — cached hit (common case):
  Target: < 0.5ms p99
  Measurement: go test -bench=BenchmarkEnforce ./internal/core/authz/...

EnforceBatch(10 requests):
  Target: < 1ms p99

InvalidateCache() — full policy reload:
  10,000 rules: < 200ms
  100,000 rules: < 1s
  (This is NOT on the hot path — admin-only operation)

AssignRole():
  Target: < 10ms (includes DB write + in-memory update)
  DB write: INSERT + Casbin g-rule update

AddPolicy():
  Target: < 5ms (DB INSERT + in-memory update)
```

### Memory Management

```markdown
MEMORY CONSIDERATIONS:

Casbin's in-memory model uses Go maps and slices.
For 100,000 policy rules: ~20-50 MB resident memory

GC pressure:
  → Casbin model is stable (reads >> writes)
  → Low allocation rate on Enforce() calls
  → InvalidateCache() causes a GC pause (rebuild all maps)
  → Schedule InvalidateCache() during low-traffic periods

MONITORING:
  Track: authz_policy_count{domain="..."}
  Alert: If policy count grows unexpectedly (possible attack: policy flooding)
  Limit: Implement max policies per domain (platform-level governance)
         Suggested: max 10,000 p-rules per tenant domain
```

### Performance Anti-Patterns

```markdown
❌ PER-REQUEST POLICY RELOAD:
   // NEVER DO THIS:
   func handler(c *fiber.Ctx) error {
       svc.InvalidateCache(ctx) // reloads entire DB every request!
       ok, _ := svc.Enforce(...)
       ...
   }

❌ POLICY CHECK INSIDE LOOP:
   // AVOID:
   for _, invoice := range invoices {
       ok, _ := svc.Enforce(ctx, Request{Object: "invoice/"+invoice.ID, ...})
   }

   // PREFER:
   reqs := make([]authz.Request, len(invoices))
   for i, inv := range invoices { reqs[i] = Request{Object: "invoice/"+inv.ID, ...} }
   results, _ := svc.EnforceBatch(ctx, reqs)  // single call

❌ DUPLICATE ENFORCE CALLS:
   // AVOID checking the same permission multiple times per request:
   ok1, _ := svc.Enforce(ctx, Request{Object: "invoice/*", Action: "read"})
   // ... later in same request ...
   ok2, _ := svc.Enforce(ctx, Request{Object: "invoice/123", Action: "read"})
   // If middleware already checked, handler should trust the result

✅ PERMISSION PRE-COMPUTATION FOR FRONTEND:
   // Check all permissions for a page in one batch call at page load
   // Cache result in request context
   // UI uses cached result to show/hide buttons
```

---

Next: [Security Considerations](./17-security-considerations.md)
