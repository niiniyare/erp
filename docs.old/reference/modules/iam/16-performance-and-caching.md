[<-- Back to Index](README.md)

## Performance & Caching

> **[IMPLEMENTED]** — describes v1.0 actual runtime behavior.
>
> **IMPORTANT**: A prior version of this document described a two-layer permission architecture
> with `session.Can()` as the hot path. That architecture is NOT active. All authorization goes
> through `authzService.Enforce()` (in-memory Casbin). The `Permissions` map and `Can()` /
> `CanDo()` methods were removed from `ResolvedSession`. Benchmark numbers have been updated
> to reflect the single-path model.

---

### Full Request Latency Profile

```
Session validation (Redis hit):    < 1ms    GET "session:{hash}" → unmarshal
Session validation (DB fallback):  ~2ms     TouchAndGetSession atomic query
revokeExpiredRoles():              < 0.1ms  partial-index scan, typically 0 rows
Casbin Enforce() (in-memory):      ~0.1ms   no DB, no Redis
FeatureEnabled() / Setting read:   < 0.1ms  map lookup in session.Configuration
Entity scope check:                0ms      struct field access

Total auth overhead per request: ~1–2ms (Redis hit path)
                                 ~3ms    (DB fallback path)
```

---

### Login Session Build (One-Time Cost)

At login, four queries run (sequentially, not fully parallelised in current implementation):

| Operation | Cost |
|---|---|
| ResolveEntityScope | ~1ms |
| ResolveAllFlagsForTenant | ~1ms |
| ResolveAllSettingsForTenant | ~1ms |
| GetUserPreferences | ~0.5ms |
| INSERT user_sessions | ~1ms |
| Redis CacheResolved SET | < 0.5ms |
| **Total at login** | **~5ms** |

No further DB hits for flags, settings, or preferences for the session lifetime (default 8h, configurable via `iam.session_ttl_hours` tenant setting).

**Note**: A `ComputePermissions` query is NOT performed at login. Permission checks happen live via Casbin on each request.

---

### Redis Cache Keys [IMPLEMENTED]

```
"session:{sha256hex(rawToken)}"
  Value: JSON-serialised ResolvedSession (UserID, TenantID, UserType, EntityScope, Configuration)
  TTL:   session.expires_at − time.Now()
  Note:  Contains NO permissions map — context only

"user_sessions:{userID}"
  Value: JSON array of token hashes belonging to this user
  TTL:   longest session TTL for this user
  Used:  InvalidateByUser() eviction — bulk-deletes all session cache entries for the user

"mfa:login:pending:{rawPendingToken}"
  Value: userID string
  TTL:   5 minutes
  Note:  Consumed atomically via GETDEL — prevents replay

"mfa:setup:pending:{userID}"
  Value: AES-256-GCM encrypted TOTP secret
  TTL:   set during InitiateMFA
  Used:  Consumed by ConfirmMFA; cleared after successful confirmation

"apikey:{sha256hex(rawKey)}"
  Value: JSON-serialised ResolvedSession for the API key
  TTL:   5 minutes
  Note:  Revocation takes effect within 5 minutes (no immediate eviction by key ID)
```

**What is NOT cached:**
- Role lists (no `roles:{userID}:{domain}` key)
- Permission maps (removed)
- Computed permissions per resource

---

### Casbin In-Memory Model

Casbin holds the complete policy set in memory, loaded from PostgreSQL at startup via `pgxAdapter.LoadPolicy()`. In steady state, `Enforce()` never hits the database.

```
Startup:
  NewSyncedEnforcer(model, pgxAdapter)
  → LoadPolicy(): SELECT ptype, v0..v5 FROM casbin_rule
  → builds in-memory p-rule and g-rule graphs
  StartAutoLoadPolicy(30s): spawns background goroutine to reload every 30s

Every Enforce() call:
  1. revokeExpiredRoles(subject, domain) — one partial-index DB query (usually 0 rows)
  2. enforcer.Enforce(sub, dom, obj, act) — pure in-memory evaluation
     → no DB, no Redis, no network
  3. Result returned

Total: 0.1–0.5ms typical
```

---

### Lazy Expiry Query Performance

The expiry check runs on every `Enforce()` call:

```sql
SELECT role_name FROM role_assignments
WHERE subject = $1 AND domain = $2
  AND is_active = TRUE
  AND expires_at IS NOT NULL
  AND expires_at < NOW()
```

Index used: `idx_role_assignments_expires` — partial index on `expires_at WHERE expires_at IS NOT NULL`.

- Most role assignments are permanent (`expires_at IS NULL`) → not indexed → near-zero I/O.
- For subjects with 0 expired roles (the common case): scan returns immediately.
- Typical measured cost: < 0.1ms.

---

### Multi-Instance Policy Propagation [IMPLEMENTED]

Each instance holds its own in-memory Casbin model. Propagation:

```
Instance A: AddPolicy(new rule)
  → pgxAdapter writes to casbin_rule in PostgreSQL  (immediate)
  → in-memory model on Instance A updated           (immediate)

Instance B, C, ...:
  → NOT updated yet
  → StartAutoLoadPolicy(30s) background goroutine fires
  → LoadPolicy() reloads all rules from DB
  → Maximum lag: 30 seconds

InvalidateCache() on an instance:
  → Forces immediate LoadPolicy() on THAT instance only
  → Does not broadcast to other instances
```

For critical revocations (terminated employee, suspended account), call `InvalidateCache()` on all instances or restart them. The 30-second window is acceptable for normal operations.

**[PLANNED - NOT IN v1.0]**: Redis pub/sub for sub-second propagation.

---

### Auto-Save Write Flow

`EnableAutoSave(true)` means every policy/role mutation writes to both DB and in-memory model atomically:

```
svc.AddPolicy(ctx, policy)
  → enforcer.AddPolicy(sub, dom, obj, act, eft)
     → pgxAdapter.AddPolicy() → INSERT INTO casbin_rule ON CONFLICT DO NOTHING
     → in-memory model updated
  → Returns to caller

DB and in-memory are always consistent after each write on the calling instance.
```

Exception: direct SQL writes to `casbin_rule` (migrations, manual DB operations) are not detected by the in-memory model. Call `InvalidateCache()` after any manual DB modification.

---

### Performance Anti-Patterns

```
DO NOT:
  svc.InvalidateCache(ctx)          // inside a request handler — reloads entire DB every request
  svc.Enforce(...)                   // inside a loop over N items — use EnforceBatch instead
  authzSvc.GetPolicies(domain)       // inside a request handler — expensive, admin-only operation

DO:
  EnforceBatch([]domain.Request{...})  // check N permissions in one call
  sess.FeatureEnabled("flag.key")      // O(1) map lookup — flags are context, not authorization
  sess.SettingString("key", default)   // O(1) map lookup — settings are context
```

---

### Benchmark Targets

| Benchmark | Target | Notes |
|---|---|---|
| `Enforce()` p99 | < 1ms | In-memory; includes revokeExpiredRoles |
| `EnforceBatch(10)` p99 | < 2ms | Batch in-memory |
| `ValidateSession` (Redis hit) | < 1ms | GET + unmarshal |
| `ValidateSession` (DB fallback) | < 5ms | TouchAndGetSession + cache write |
| `Login` end-to-end | < 10ms | 4 queries + DB insert + Redis SET |
| `LoadPolicy` (10k rules) | < 200ms | Admin-only; not on hot path |
| `LoadPolicy` (100k rules) | < 1s | Startup only |

---

Next: [Security Considerations](./17-security-considerations.md)
