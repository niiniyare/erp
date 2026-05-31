[<-- Back to Index](README.md)

## Session Model — Context Only

> **[IMPLEMENTED]** — This document describes the v1.0 session model.
>
> **IMPORTANT**: This document has been significantly revised from a prior version that described
> a "pre-computed permissions" model. That model was replaced. The current session carries
> **context only** — no permission snapshot, no `Can()` method, no `CanDo()` method.
> All authorization decisions go through `authzService.Enforce()`.

---

### What Gets Built at Login

At login, the session service calls `buildAndPersistSession()`, which computes:

1. `generateToken()` — 32 bytes of random entropy → raw token (returned to client once) + sha256hex hash (stored).
2. `repo.ResolveEntityScope(ctx, user.EntityID)` — one DB query to resolve entity hierarchy.
3. `repo.LoadLoginConfig(ctx, userID, tenantID)` — resolves flags, tenant settings, and user preferences.
4. Computes TTL from `iam.session_ttl_hours` tenant setting, fallback to 8h.
5. INSERT `user_sessions` row.
6. Cache `ResolvedSession` in Redis at key `"session:{hash}"` with TTL.

**What is NOT computed at login:**
- No permission map (no `ComputePermissions()` call).
- No role list.
- No Casbin policy evaluation.

---

### The ResolvedSession Type [IMPLEMENTED]

```go
// internal/core/iam/domain/session.go

type ResolvedSession struct {
    UserID        uuid.UUID     // who the user is
    UserType      string        // persisted enum: "INTERNAL"|"SYSADMIN"|"CUSTOMER"|"PORTAL"|"API"
    TenantID      uuid.UUID     // RLS boundary — SET LOCAL awo.tenant_id = '<TenantID>'
    PrincipalID   *uuid.UUID    // non-nil for portal users; identifies represented party
    DisplayName   string        // for display only
    EntityScope   EntityScope   // application-layer entity visibility
    Configuration Configuration // feature flags, tenant settings, user preferences
    // NOTE: No Permissions map. No Can() method. No CanDo() method.
    // Authorization is delegated entirely to authzService.Enforce().
}

type EntityScope struct {
    Type       EntityScopeType // "all" | "subtree" | "entity"
    EntityID   string          // home entity UUID (for entity/subtree scopes)
    PathPrefix string          // ltree path prefix (for subtree queries)
}

type Configuration struct {
    Flags    map[string]bool   // "feature.name" → true/false
    Settings map[string]string // "setting.key" → "value"
    Prefs    map[string]string // "pref.key" → "value"
}
```

**Available helper methods on `*ResolvedSession`:**
- `ToPrincipal()` — builds the Casbin `(Subject, Domain)` pair for `Enforce()` calls.
- `FeatureEnabled(flag string) bool` — O(1) map lookup on `Configuration.Flags`.
- `SettingString(key, default)`, `SettingBool(key, default)`, `SettingInt(key, default)`, `SettingDecimal(key, default)`.
- `IsPlatform() bool`, `IsPortal() bool` — identity checks.

**NOT available:**
- `Can(permission string) bool` — removed; use `authzService.Enforce()`.
- `CanDo(resource, action string) bool` — removed; use `authzService.Enforce()`.

---

### Why Session-as-Context (Not Session-as-Authority)

**Single enforcement authority**: Having two enforcement paths (session.Can() + Casbin.Enforce()) creates a split: role revocations would take effect in Casbin immediately but persist in the session for up to 8h via the cached permissions map. This is a correctness defect.

**Performance**: Casbin `Enforce()` is in-memory (~0.1ms). The session permission map's O(1) advantage is real but not significant for ERP workloads where total request latency is 50–200ms.

**Simplicity**: One path means one place to audit, one place to debug, one place to add deny rules.

---

### Configuration — Pre-Computed at Login [IMPLEMENTED]

Feature flags, tenant settings, and user preferences are still pre-computed at login. These are **context reads** — they tell the code *how* to behave, not *whether* to act.

```go
// Reading flags (O(1), no DB, no Casbin)
if sess.FeatureEnabled("hr.payroll_v2.enabled") {
    // use new payroll logic
}

// Reading settings (O(1), no DB, no Casbin)
ttl := sess.SettingInt("iam.session_ttl_hours", 8)

// Reading preferences (O(1), no DB, no Casbin)
theme := sess.Configuration.Prefs["ui.theme"]
```

Flags and settings may be stale if they change after the session was created (up to the session TTL). For most ERP settings this is acceptable. For security-relevant flag changes, call `InvalidateByTenant()` to force re-login.

---

### Entity Scope [IMPLEMENTED]

`EntityScope` controls data visibility **within a tenant**, enforced by service methods that add WHERE clauses. It is not an authorization decision — it is a query filter.

```
EntityScopeAll     → no entity WHERE clause; user sees all entities in tenant
EntityScopeSubtree → WHERE entity_path <@ '{PathPrefix}'
EntityScopeEntity  → WHERE entity_id = '{EntityID}'
```

Resolved once at login via `ResolveEntityScope(ctx, entityID)`:
- `entityID == uuid.Nil` → `EntityScopeAll` (platform/system users with no entity).
- Root entity (level 1) → `EntityScopeAll`.
- Branch entity (has children) → `EntityScopeSubtree` with path prefix.
- Leaf entity (no children) → `EntityScopeEntity`.

---

### Session Storage [IMPLEMENTED]

```
Redis key: "session:{sha256hex(rawToken)}"
  Value: JSON-serialized ResolvedSession
  TTL: session.expires_at - now

Redis key: "user_sessions:{userID}"
  Value: JSON array of token hashes for this user
  TTL: longest session TTL for this user
  Used by: InvalidateByUser() to find and evict all user's cached sessions

Redis key: "mfa:login:pending:{pendingToken}"
  Value: userID string
  TTL: 5 minutes
  Used by: MFA step 1→2 handoff
  Note: consumed atomically via GETDEL (prevents replay)

DB table: user_sessions
  Authoritative record. Redis is cache-aside — DB is the source of truth.
  TouchAndGetSession atomically touches last_seen_at and reads the row.
```

---

### Session Invalidation [IMPLEMENTED]

| Trigger | Call | Redis behavior |
|---------|------|----------------|
| Logout | `repo.Invalidate(hash)` | DELETE "session:{hash}" immediately |
| Role revoked | `SessionInvalidator.InvalidateByUser(userID)` (wired in RevokeRole) | Evict all user's hashes from Redis |
| Explicit user suspension | `repo.InvalidateByUser(userID)` | Evict all user's hashes from Redis |
| Tenant-wide event | `repo.InvalidateByTenant(tenantID)` | DB only; Redis expires naturally |
| Session TTL expires | `expires_at < NOW()` in DB | Redis entry TTL also expires |

**Note on role assignment**: Assigning a new role does NOT require session invalidation in the Casbin model. Since `Enforce()` is always called live, the new role takes effect immediately on the next request.

---

### Login Performance [IMPLEMENTED]

```
At login (one-time cost):
  ResolveEntityScope:          ~1ms   (entity level/path DB query)
  LoadLoginConfig:             ~2ms   (flags + settings + prefs queries)
  INSERT user_sessions:        ~1ms
  CacheResolved (Redis SET):   <1ms
  Total:                       ~4-5ms

Per request (hot path):
  Redis cache hit (GET):       <1ms
  FeatureEnabled/SettingRead:  <0.1ms  (map lookup)
  authzService.Enforce():      ~0.1ms  (in-memory Casbin)
  Total auth overhead:         ~1-2ms
```

---

Next: [Middleware & HTTP Integration](./11-middleware-and-http.md)
