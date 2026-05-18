# IAM MODULE v1.0 DELIVERY TASK PLAN

> **Document type**: Engineering execution plan — build contract for production stability.
> **Architecture baseline**: RBAC-only, Casbin-driven, Session-as-context.
> **Source of authority**: `docs/reference/modules/iam/` (full reference suite) + `testing.md`
> **Last validated**: 2026-05-11 (third-pass architectural audit + documentation reconciliation)
> **Last updated**: 2026-05-18 (eighth-pass — RUNTIME-P5 transactional authz compensation complete)

---

## 0. Implementation Status Summary (as of 2026-05-11)

Completed BLOCK items:

| BLOCK | Item | Status |
|---|---|---|
| BLOCK-1 | Single-path Casbin enforcement — `session.Can()` / `CanDo()` never existed; all auth via Casbin | DONE |
| BLOCK-2 | Logout evicts Redis: `Invalidate(hash)` calls `cache.Delete()` synchronously | DONE |
| BLOCK-3 | `SessionInvalidator` wired into authz service; `RevokeRole` calls `InvalidateByUser` | DONE |
| BLOCK-4 | `casbin.SyncedEnforcer` with `StartAutoLoadPolicy(30s)` | DONE |
| BLOCK-5 | `internal/core/access/` gated with `//go:build ignore` | DONE |
| BLOCK-6 | JIT bootstrap moved out of session service; bootstrap only in `RegisterNewUser` | DONE |
| SES-1 | `Permissions` map and `Can()`/`CanDo()` removed from `ResolvedSession` | DONE |
| SES-2 | `buildPermissions()` removed from session service | DONE |
| SES-3 | Redis session cache contains context only (no permissions) | DONE |
| SES-5 | MFA pending token atomic consumption via Redis GETDEL | DONE |

**Completed since last audit (2026-05-11):**

| Item | Status |
|---|---|
| AUTHZ-4: platform domain write guard + AssignedBy required | DONE |
| AUTHZ-5: policy count limit per domain (DoS prevention) | DONE |
| AUTHZ-7: structured security event audit log (ROLE_ASSIGNED/REVOKED, POLICY_ADDED/REMOVED) | DONE |
| SES-4: `make sqlc` re-run after permissions column removal | DONE |
| DB-1: migration `001008` to drop `permissions` column from `user_sessions` | DONE |
| T-ROLES: all tests uncommented and passing (DB-backed, requires DATABASE_URL) | DONE |
| T-ISO (AZ-ISO-001..020): all passing | DONE |
| T-UNIT (AZ-TYP-*, session model, AssignOpts, errors): all passing | DONE |
| T-SEC: `TestSystemRole_ImmutableFromTenantActor`, `TestPolicyCountLimit_Enforced` | DONE |
| Casbin model wildcard subject fix (`p.sub == "*"` in matcher) | DONE |
| AUTHZ-2: `Authorize` middleware single Casbin path; bug fix (503→500 on Enforce error) | DONE |
| T-MIDDLEWARE: AZ-MID-030 + post-refactor suite (authorize_test.go) | DONE |
| AUTHZ-3: audit confirms no bypass paths; 4 regression guard tests added | DONE |
| IAM contract layer (`iam/contract/`): SessionContext, AuthService, InjectSessionContext, adapter, tests | DONE |
| System-wide boundary audit: `contract-adoption-audit.md` — 3 CRITICAL violations found in finance + shared | DONE |

**Remaining open items** — see Section 2 onward:

| Item | Status |
|---|---|
| AUTHZ-2: verify production `AuthorizeCasbin` middleware calls `Enforce()` | DONE |
| AUTHZ-3: audit no handler bypasses auth via direct DB role query | DONE |
| CONTRACT-FIN-1: remove iam.Service from finance/service.go + account.go | DONE |
| CONTRACT-FIN-2: replace featureflag.Service with SessionContext.FeatureEnabled in account.go | DONE |
| CONTRACT-FIN-3: delete all TODO(authz) planned bypasses (9 found across account.go) | DONE |
| CONTRACT-SHARED-1: remove CapabilityContext.Permissions/Features/Modules from shared/context.go | DONE |
| GUARD: iam/guard/ package — 8 static boundary tests enforcing contract rules | DONE |
| AUTHZ-6: subject prefix validation in authn middleware | DONE |
| RUNTIME-P1: session-as-authority audit + AUTHZ-RUNTIME-1..4 guard tests | DONE |
| RUNTIME-P2: Casbin distributed sync (watcher / LISTEN-NOTIFY) | DONE |
| RUNTIME-P3: session invalidation bridge (RevokeRole → evict, Tenant suspend → evict all) | DONE |
| RUNTIME-P4: middleware enforcement integrity audit | DONE |
| RUNTIME-P5: transactional authz mutations (Casbin + DB atomic) | DONE |
| RUNTIME-P6: observability + startup readiness gate | OPEN |
| RUNTIME-P7: documentation sync (operational guarantees section) | OPEN |
| T-INT: integration tests (DB + Redis required) | OPEN |
| T-ADAPTER (AZ-ADP-001..050): DB required | OPEN |
| T-SERVICE (AZ-SVC-001..061): DB required | OPEN |
| T-TEMP (AZ-TEMP-001..010): DB required | OPEN |
| T-MIDDLEWARE (AZ-MID-001..040 + post-refactor tests): Fiber mock | DONE |
| ROLE-1..5: role lifecycle, consistency, audit trail | OPEN |
| CACHE-2/3: invalidation matrix wiring, API key revocation doc | OPEN |
| DB-2/3/4: constraint/index validation, migration rollback safety | OPEN |
| SEC-4: API key lifecycle — document 5-min TTL window | OPEN |
| Documentation updates `10b`, `11`, `12b` | DONE (2026-05-11) |

---

## 1. Executive Delivery Summary

### Current State

IAM module implements **single-path Casbin enforcement**:

- Session carries identity + context only: `UserID`, `TenantID`, `UserType`, `EntityScope`, `Configuration`. No `Permissions` map. No `Can()` / `CanDo()`.
- Every authorization decision routes through `authzService.Enforce()`.
- `casbin.SyncedEnforcer` with `StartAutoLoadPolicy(30s)` provides goroutine safety and multi-instance convergence.
- `RevokeRole()` calls `SessionInvalidator.InvalidateByUser()` — session cache evicted on role revocation.
- `Logout()` synchronously deletes Redis session cache entry.
- MFA pending token consumed atomically via Redis `GETDEL`.
- `internal/core/access/` (ABAC) gated with `//go:build ignore`.

### Remaining Security Gaps (must fix before production)

| Issue | Priority |
|---|---|
| AUTHZ-4: platform domain write guard not enforced at service layer | HIGH |
| AUTHZ-5: no policy count limit per domain (DoS exposure) | HIGH |
| AUTHZ-7: no persistent security event audit log | HIGH |
| Subject prefix validation in authn middleware | HIGH |

---

## 2. Critical Architectural Fixes

> **[DONE]** = verified implemented. **[OPEN]** = requires work before production.

---

### BLOCK-1 — Migrate from Two-Path to Single Casbin Enforcement [DONE]

**Context**: Two-path system was previous design. Session layer rewritten to carry context only. `session.Can()` and `CanDo()` never present in current codebase — all auth through `authzService.Enforce()`. Documentation updated 2026-05-11.

**Performance implication**: Every protected request calls `Casbin.Enforce()` (in-memory, no DB on hot path). Expected latency: ~0.1ms per call vs previous O(1) map lookup. Acceptable for ERP workloads. Feature flag and setting reads remain O(1) from `session.Configuration`.

**Files**:
- `internal/core/iam/domain/session.go` — remove `Permissions`, `Can()`, `CanDo()`, `RiskScore`
- `internal/core/iam/service/session.go` — remove `buildPermissions()` and all callers
- `internal/core/iam/repository/session.go` — remove permission JSONB marshal/unmarshal
- `internal/api/middleware/` — rewrite `RequirePermission` to call `authzService.Enforce()`
- `db/queries/sessions.sql` — remove `permissions` from INSERT / SELECT

**Expected final state**:
```go
// REMOVED from domain/session.go:
//   Permissions map[string]bool
//   RiskScore float64
//   func (s *ResolvedSession) Can(permission string) bool
//   func (s *ResolvedSession) CanDo(resource, action string) bool

// REWRITTEN middleware (was session.CanDo, now Casbin):
func RequirePermission(authzSvc authz.Service, resource, action string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        sess := ContextSession(c)
        principal := sess.ToPrincipal()
        ok, err := authzSvc.Enforce(c.Context(), authz.Request{
            Subject: principal.Subject,
            Domain:  principal.Domain,
            Object:  resource,
            Action:  action,
        })
        if err != nil { return c.Status(500).JSON(...) }
        if !ok       { return c.Status(403).JSON(...) }
        return c.Next()
    }
}
```

**Documentation to update**: `10b-session-precomputation.md`, `12b-http-middleware.md` (Section 15).

**Risk if not fixed**: Two concurrent enforcement authorities. Role revocations don't take effect in fast path for up to 8h. Wildcard `"*"` sentinel outside Casbin control.

---

### BLOCK-2 — Fix Logout: Evict Redis on Session Invalidation [DONE]

**Context**: Was identified as bug. Fixed: `repo.Invalidate(hash)` calls `cache.Delete(ctx, sessionCacheKey(hash))` synchronously. `InvalidateByUser` evicts all user session Redis keys via `user_sessions:{userID}` index. `ValidateToken()` hits cache first — logged-out user with valid Redis key continued to pass authentication.

**File**: `internal/core/iam/repository/session.go`

Tasks:
- [ ] `Invalidate(ctx, tokenHash)`: call `cache.Delete(ctx, sessionCacheKey(tokenHash))` BEFORE returning, not as fallback
- [ ] `InvalidateByUser(ctx, userID)`: enumerate and delete all Redis keys for `userID` OR use Redis key prefix pattern for bulk eviction
- [ ] `InvalidateByTenant(ctx, tenantID)`: same — bulk evict tenant session keys
- [ ] Confirm: `Logout()` in session service calls `repo.Invalidate()` which now deletes Redis

**Risk if not fixed**: Logout is cosmetic. Forcibly terminated users retain active sessions.

---

### BLOCK-3 — Wire SessionInvalidator into Authz Service [DONE]

**Context**: Was missing. Now implemented: `authzService` holds `SessionInvalidator` interface (implemented by `SessionRepository`). `RevokeRole()` calls `sessionInv.InvalidateByUser(userID)` after removing Casbin g-rule. `AuthzConfig.SessionInvalidator` is optional (nil-safe for tests). Invalidation matrix documented in `10b-session-precomputation.md` — only flag/setting path was wired; authz path was not.

**Files**:
- `internal/core/iam/service/authz.go` — add `SessionInvalidator` dependency, call on mutations
- `internal/core/iam/service/session.go` — expose `InvalidateByUser` / `InvalidateByTenant` as `SessionInvalidator` implementation

**Interface** (add to `service/authz.go`):
```go
// SessionInvalidator is called when authorization state changes for a subject.
// Implementations must be nil-safe (called during bootstrap when no sessions exist).
type SessionInvalidator interface {
    InvalidateByUser(ctx context.Context, userID uuid.UUID) error
    InvalidateByTenant(ctx context.Context, tenantID uuid.UUID) error
}
```

**Wire into**:
- `RevokeRole()`: parse `userID` from subject string → `invalidator.InvalidateByUser()`
- `RemovePolicy()`: parse `tenantID` from domain → `invalidator.InvalidateByTenant()`
- `AssignRole()`: `InvalidateByUser()` to force re-authentication with new permissions

**Risk if not fixed**: Role revocation doesn't revoke access. Security contract broken.

---

### BLOCK-4 — Replace Enforcer with `casbin.SyncedEnforcer` [DONE]

**Context**: Now implemented: `NewAuthzService` uses `casbin.NewSyncedEnforcer(m, adapter)` and calls `e.StartAutoLoadPolicy(30 * time.Second)`. Thread safety and 30s multi-instance convergence active. Previous non-synced enforcer caused data race under concurrent `Enforce()` + `RevokeRole()` calls.

**File**: `internal/core/iam/service/authz.go`

Tasks:
- [ ] Replace `casbin.NewEnforcer(m, adapter)` with `casbin.NewSyncedEnforcer(m, adapter)`
- [ ] Call `enforcer.StartAutoLoadPolicy(30 * time.Second)` after init
- [ ] Stop auto-loader on shutdown: `enforcer.StopAutoLoadPolicy()` in cleanup hook
- [ ] `InvalidateCache()` becomes no-op wrapper over `enforcer.LoadPolicy()`
- [ ] Add metric: `iam.authz.policy_reload` counter (label: `method=auto_reload`)

**Note**: `SyncedEnforcer.StartAutoLoadPolicy(30s)` handles goroutine safety AND multi-instance sync. No separate watcher needed for v1.0.

**Risk if not fixed**: Data race under concurrent authz mutations. Undefined behavior in production.

---

### BLOCK-5 — Gate `internal/core/access/` Module [DONE]

**Context**: Implemented. Every `.go` file in `internal/core/access/` carries `//go:build ignore`. Package doesn't compile into binary. DB migrations 000404–000413 marked as v2.0 reserved.

Tasks:
- [x] Add `//go:build ignore` to all `.go` files in `internal/core/access/`
- [x] Confirm `go build ./...` succeeds with no `internal/core/access` imports in active code
- [x] Add `README.md` in `internal/core/access/`: reserved for v2.0 ABAC, not active in v1.0
- [x] Mark DB migrations `000404`–`000413` with comments: reserved for v2.0 (do not drop)
- [x] Audit `db/queries/policies.sql` — remove or gate queries used only by access module

**Risk if not fixed**: ABAC surface accidentally wired. Dead migration tables create confusion.

---

### BLOCK-6 — Move JIT Bootstrap Out of Session Service [DONE]

**Context**: Was privilege escalation path. Fixed: `BootstrapTenantAdmin()` called only from `UserService.RegisterNewUser()`, not from session/login path. Session service has no reference to authz service; no JIT bootstrap during login. Any roleless tenant user at login silently got `tenant_admin` — role assignment must not happen inside session layer.

**Files**:
- `internal/core/iam/service/session.go` — remove JIT bootstrap call
- `internal/core/iam/seed.go` — `SeedDefaultRoles` + `AssignAdminRole` already exist; wire at tenant provision time (not login)
- Tenant provisioning flow — explicit call to `authzSvc.BootstrapTenantAdmin()` after tenant creation

**Expected change**:
```go
// REMOVE from buildPermissions() / login path:
if len(roles) == 0 {
    _ = s.authz.BootstrapTenantAdmin(ctx, userID, tenantID) // REMOVE THIS
}

// ENSURE in tenant.Service.Provision():
authzSvc.BootstrapTenantAdmin(ctx, firstAdminUserID, tenantID)
```

**Risk if not fixed**: First login from any tenant without roles = silent privilege escalation.

---

## 3. Session Layer Refactor Tasks

### SES-1 — Strip Permission Snapshot from Session Domain Model

**File**: `internal/core/iam/domain/session.go`

Tasks:
- [x] Remove `Permissions map[string]bool` from `Session` struct
- [x] Remove `Permissions map[string]bool` from `ResolvedSession` struct
- [x] Remove `Can(permission string) bool` method
- [x] Remove `CanDo(resource, action string) bool` method
- [x] Remove `RiskScore float64` field (dead — no computation exists in any service)
- [x] Remove wildcard `"*"` sentinel logic from `Can()` (deletes with the method)
- [x] Keep `ToPrincipal()` — required for Casbin `Enforce()` calls
- [x] Keep `EntityScope` — legitimate query context, not enforcement
- [x] Keep `Configuration` (`Flags`, `Settings`, `Prefs`) — O(1) context reads remain valid
- [x] Keep `FeatureEnabled()`, `SettingString()`, `SettingBool()`, `SettingInt()`, `SettingDecimal()`
- [x] Keep `IsPlatform()`, `IsPortal()` — identity checks, not authorization

**Verify**: `grep -rn "\.Can\b\|\.CanDo\b" --include="*.go" internal/` returns zero hits.

---

### SES-2 — Remove buildPermissions from Session Service

**File**: `internal/core/iam/service/session.go`

Tasks:
- [x] Delete `buildPermissions()` function entirely
- [x] Remove all `authz.GetImplicitRoles()` calls from session service
- [x] Remove all `authz.GetPolicies()` calls from session service
- [x] Remove local deny-override logic (belongs to Casbin model, not session service)
- [x] Remove JIT bootstrap call (BLOCK-6)
- [x] Simplify `buildAndPersistSession()` — no permission computation, no role queries
- [x] Session payload after refactor: `UserID`, `TenantID`, `UserType`, `EntityScope`, `Configuration` (flags/settings/prefs), token fields, `IPAddress`, `UserAgent`, `ExpiresAt`
- [x] Five parallel queries in `buildSession` reduce to: `ResolveEntityScope`, `FlagService.ResolveForTenant`, `SettingService.ResolveForTenant`, `UserPreferences` — four total (permissions query removed)

---

### SES-3 — Fix Redis Session Cache Format

**File**: `internal/core/iam/repository/session.go`

Tasks:
- [x] Remove `Permissions` JSONB serialization from `CacheResolved()`
- [x] Remove `Permissions` deserialization from `ValidateToken()` cache hit path
- [x] Verify cached `ResolvedSession` contains no permission data
- [x] Fix `Invalidate()` to DELETE Redis key (BLOCK-2)
- [x] Cache TTL must match `session.ExpiresAt - time.Now()`, not fixed `SessionTTL` value

---

### SES-4 — Remove Permissions Column from Session DB/SQLC

**Files**: `db/queries/sessions.sql`, `db/sqlc/sessions.sql.go`

Tasks:
- [x] Remove `permissions` JSONB from `CreateSession` INSERT
- [x] Remove `permissions` from `GetSessionByToken` / `ValidateToken` SELECT
- [ ] Run `make sqlc` after query changes
- [x] Comment out `permissions` column in `db/migration/000305_identity_sessions_add_permissions.up.sql` (note: column is in 000305 not 000304; `principal_id` kept in same migration)
- [x] Keep `principal_id` column — valid for audit trail

---

### SES-5 — Fix MFA Pending Token Atomicity

**File**: `internal/core/iam/repository/session.go`

Tasks:
- [x] Replace `GetPendingMFA()` + separate `DeletePendingMFA()` with Redis `GETDEL` command
- [x] `GETDEL` atomically fetches and deletes — prevents concurrent `CompleteMFALogin` producing duplicate sessions
- [x] Added `GetAndDelete` to `cache.Service` interface + `redisClient` impl (uses `client.GetDel`)
- [x] `DeletePendingMFA` retained for explicit cancellation paths; `CompleteMFALogin` no longer calls it
- [x] Add test: two concurrent `CompleteMFALogin` calls for same pending token → only one session

---

## 4. Authorization Layer (Casbin Core)

### AUTHZ-1 — Enforce Single Enforcement Entrypoint

After BLOCK-1 removes `Can()` / `CanDo()`, callers will fail to compile. Use as forcing function — don't silence compile errors, fix each call site.

Tasks:
- [x] `grep -rn "\.Can\b\|\.CanDo\b\|RequirePermission" --include="*.go" internal/ cmd/` — fix all hits
- [x] No call sites found — `Can`/`CanDo` never existed; all authz via `authzService.Enforce()`
- [x] Subject and Domain always from `session.ToPrincipal()` (set in `setSessionLocals`)
- [x] `Authorize` middleware calls `Enforce()` directly (see BLOCK-1)

---

### AUTHZ-2 — Standardize Middleware: Single Casbin Path [DONE]

**Files**: `internal/api/middleware/session_middleware.go`, `internal/core/iam/`

Tasks:
- [x] `Authorize(cfg, permission)` calls `cfg.AuthzService.Enforce()` — single Casbin path confirmed
- [x] `RequireFlag(flagKey)` — reads `session.FeatureEnabled()` (context, not enforcement)
- [x] No `authzSvc.Middleware()` method — HTTP middleware lives in `api/middleware`, not on interface
- [x] "Two-path" table removed from `12b-http-middleware.md` — now documents single-path only
- [x] Middleware flow: `Authenticate` → set `Principal` in Locals → `Authorize(cfg, perm)` → `Enforce`
- [x] `Authorize` returns: 401 (no session/principal), 500 (Enforce error), 403 (denied), next (allowed)
- [x] Bug fixed: `Authorize` was returning 403 on `Enforce` error — now correctly returns 500
- [x] Tests: AZ-MID-030 (enforce error → 500), post-refactor suite in `authorize_test.go` — all pass

**Implementation notes (2026-05-11)**:
- Fixed: `internal/api/middleware/session_middleware.go` — split merged `err != nil || !allowed` condition
- Added: `internal/api/middleware/authorize_test.go` — 4 post-refactor tests (recording stub, RequireFlag, EnforceError→500, NoSessionCan compile proof)
- Added: `internal/core/iam/middleware_impl_test.go` — `errAuthzService` stub for AZ-MID-030
- Added: `TestMiddleware_EnforceError_Returns500` to `MiddlewareUnitSuite` in `unit_test.go`
- Updated: `docs/reference/modules/iam/12b-http-middleware.md` — single-path enforcement only

---

### AUTHZ-3 — Remove All Enforcement Bypass Paths [DONE]

Tasks:
- [x] No handler may check authorization via direct DB role query — verified, none found
- [x] No handler may inspect `session.UserType` to bypass authorization — verified, none found
- [x] No service may make authorization decisions based on role strings — verified, all decisions via `Enforce()`
- [x] `UserService` must not make authorization decisions — verified; identity-only interface confirmed via reflection

**Implementation notes (2026-05-17)**:
- Audit: comprehensive scan of `internal/api/handlers/`, `internal/api/middleware/`, `internal/core/` — no bypass patterns found
- `UserService` interface has no `Enforce`/`AddPolicy`/`GetPolicies` methods — confirmed via reflection test
- `AssignUserRole`/`RevokeUserRole` on `UserService` are HR-layer identity ops, not Casbin enforcement — separate concern (ROLE-3)
- Added: `internal/core/iam/bypass_audit_test.go` — 4 regression guard tests (TestAUTHZ3_*) that catch future bypass patterns

---

### AUTHZ-4 — System Role and Platform Domain Immutability Guard

**File**: `internal/core/iam/service/authz.go`

Tasks:
- [x] Guard in `AddPolicy()`: if `domain == DomainPlatform` and `p.Subject` prefix is not `"platform:"` or `"role:"` → return `ErrForbidden`
- [x] Guard in `AssignRole()`: if role has `"role:platform-"` prefix and `AssignedBy` prefix is not `"platform:"` → return `ErrForbidden`
- [x] Guard in `RemovePolicy()`: same domain check as `AddPolicy`
- [x] `AssignedBy` validated non-empty in `AssignRole()` — ErrInvalidRequest if absent
- [x] `builtinRoles` registry / `AssignableTo` check: not present in current codebase (not yet built)
- [x] Add test: AZ-SEC-050 — TestAddPolicy_PlatformDomainGuard, TestAssignRole_PlatformRoleGuard, etc.

---

### AUTHZ-5 — Policy Count Limit Guard (DoS Prevention)

**Context**: `17-security-considerations.md` T6 documents this as required control.

**File**: `internal/core/iam/service/authz.go`

Tasks:
- [x] In `AddPolicy()`, before Casbin insert, check `s.enforcer.GetFilteredPolicy(1, domain)` count
- [x] Returns `ErrPolicyLimitExceeded` (HTTP 429) when `count >= maxPolicies`
- [x] Add `ErrPolicyLimitExceeded` to `domain/errors.go`
- [x] `MaxPoliciesPerDomain` in `AuthzConfig`; `DefaultMaxPoliciesPerDomain = 10_000` constant
- [x] Metric `iam.authz.policy_count` emitted per domain on every `AddPolicy` call

---

### AUTHZ-6 — Tenant Domain Enforcement Audit

Tasks:
- [ ] Verify every `Enforce()` call passes domain derived from authenticated session (not request body)
- [ ] Verify `SetDBPool` middleware correctly propagates `TenantID` and `UserType` from session (not from request) — confirmed in `12b-http-middleware.md`
- [ ] Verify subject prefix validation in authn middleware (Rule 1 from `20-business-rules-and-validation.md`): subject without `"platform:"`, `"tenant:"`, `"portal:"`, or `"api:"` prefix must be rejected
- [ ] Add test: AZ-ISO-001 through AZ-ISO-020 pass

---

### AUTHZ-7 — Security Event Audit Logging

**Context**: `17-security-considerations.md` T7 requires audit log for every security-relevant IAM event.

**File**: `internal/core/iam/service/authz.go`, `service/identity.go`

Tasks:
- [x] `AssignRole()`: emits `SECURITY_EVENT: ROLE_ASSIGNED` (InfoContext / WarnContext for platform)
- [x] `RevokeRole()`: emits `SECURITY_EVENT: ROLE_REVOKED`
- [x] `AddPolicy()`: emits `SECURITY_EVENT: POLICY_ADDED`
- [x] `RemovePolicy()`: emits `SECURITY_EVENT: POLICY_REMOVED`
- [x] OTel span attribute `audit.event` set on each mutating call
- [x] Platform-domain events logged at WarnContext (monitoring alert hook); non-platform at InfoContext
- [x] `AssignedBy` validated non-empty in `AssignRole()` (AUTHZ-4)
- [ ] Persistent `audit_events` table + durable delivery — deferred (see Section 16 OPEN tasks)

---

## 5. Role System Implementation Tasks

### ROLE-1 — Tenant Custom Role Creation

`builtinRoles` registry in `seed.go` defines system roles. Tenant custom roles implicitly created by adding policies with arbitrary role names. Explicit creation flow needed for validation and lifecycle management.

**Files**: `internal/core/iam/service/authz.go` or new `service/roles.go`

Tasks:
- [ ] `CreateTenantRole(ctx, tenantID, roleName, description string) error`
- [ ] Validate: `roleName` must have `"role:"` prefix
- [ ] Validate: `roleName` must NOT have `"role:platform-"` prefix
- [ ] Validate: `roleName` must be unique within domain (check existing policies/g-rules)
- [ ] Creating a role does NOT add policies — policies added separately via `AddPolicy()`
- [ ] System roles (`builtinRoles`) cannot be overridden or re-created

---

### ROLE-2 — Role Inheritance via Casbin g Rules Only

Tasks:
- [ ] Confirm no inheritance logic exists outside Casbin `g` rules — code audit
- [ ] Document (in code): child role inherits parent's policies via Casbin model automatically
- [ ] To make `role:finance-viewer` inherit `role:read-only`: add g rule `(role:finance-viewer, role:read-only, domain)` — Casbin g rule, not app concept
- [ ] Add test: `GetImplicitRoles` returns transitive role chain (hierarchy test)
- [ ] Add test: revoking parent role via Casbin removes implicit permissions from child subjects

---

### ROLE-3 — Role Assignment Consistency (DB ↔ Casbin)

**Context**: Rule 8 in `20-business-rules-and-validation.md` defines reconciliation queries. Both SQL queries documented there verbatim.

Tasks:
- [ ] `AssignRole()`: writes to BOTH `role_assignments` (audit) AND `casbin_rule` (enforcement) — must be atomic (transaction or two-phase write with compensating delete)
- [ ] `RevokeRole()`: deactivates `role_assignments` AND removes g-rule atomically
- [ ] Add startup reconciliation: run Rule 8 queries on service init, log WARNING per orphaned row
- [ ] Add `reconcileRoleConsistency(ctx)` as exported maintenance method
- [ ] Add metric: `iam.authz.reconciliation.orphaned_assignments` gauge (emitted at startup)

---

### ROLE-4 — Temporal Role Lazy Revoke Safety

**Context**: Documented in `09-temporal-roles-and-expiry.md`. Lazy revoke correct approach for v1.0. Phase 2 background sweep is future work.

Tasks:
- [ ] Confirm `revokeExpiredRoles()` failure is non-fatal — logs warning, `Enforce()` continues
- [ ] Confirm expired role removed from Casbin in-memory model BEFORE `Enforce()` result computed
- [ ] Expiry evaluated at PostgreSQL `NOW()`, not Go application time — confirm query uses `NOW()`
- [ ] `expires_at` stored in UTC — validate `WithExpiry()` converts to UTC before persist
- [ ] Add validation: `AssignRole()` with `ExpiresAt` in the past returns `ErrInvalidRequest`
- [ ] Tests: AZ-TEMP-001 through AZ-TEMP-010

---

### ROLE-5 — Role Audit Trail Completeness

**Context**: `role_assignments` is authoritative audit source (Rule 8, `20-business-rules-and-validation.md`).

Tasks:
- [ ] `AssignedBy` field must always be set — validate non-empty before DB write
- [ ] Add `RevokedBy` capture: when `RevokeRole()` deactivates a row, record who revoked it (either `updated_by` column or audit log from AUTHZ-7)
- [ ] `GetAssignments()` returns both active and inactive rows — full audit trail preserved
- [ ] Confirm `GetAssignments()` ordered by `created_at DESC`
- [ ] Tests: AZ-ROLE-040, AZ-ROLE-041, AZ-ROLE-042

---

## 6. Multi-Tenant Isolation Hardening

### ISO-1 — RLS Consistency Audit

Tables that MUST have RLS:
- `user_sessions` — tenant-scoped
- `role_assignments` — tenant-scoped (audit trail)
- `users`, `persons`, `employees` — tenant-scoped

Tables that must NOT have tenant RLS:
- `casbin_rule` — Casbin loads full ruleset; domain isolation is application-level

Tasks:
- [ ] Verify RLS policy on all tenant-scoped tables: `SELECT relrowsecurity FROM pg_class WHERE relname=...`
- [ ] Verify `casbin_rule` has NO RLS tenant filter — integration test AZ-INT-031
- [ ] Verify `000306_security_active_tenant_guard` migration applied — PENDING tenant blocks session
- [ ] Verify `000307_security_policy_evaluations_rls` migration applied
- [ ] Test AZ-INT-030: `application_role` sees only its tenant's `role_assignments`
- [ ] Test: unauthenticated DB session (`set_tenant_context` not called) returns zero rows from all tenant-scoped tables

---

### ISO-2 — Subject Prefix Validation in Authn Middleware

**Context**: Rule 1 in `20-business-rules-and-validation.md`: "The authn middleware MUST enforce subject format."

Tasks:
- [x] Authn middleware: validate subject prefix in `Authenticate` after `setSessionLocals` — rejects any subject without valid prefix
- [x] Valid prefixes: `"platform:"`, `"tenant:"`, `"portal:"`, `"api:"` — all must have non-empty suffix (`isValidSubject` helper)
- [x] Returns `401 Unauthorized` with generic "authentication required" (no detail leak)
- [x] Applied to both API key and session token paths in `session_middleware.go`
- [ ] Test: AZ-SEC-050 — request crafting `Domain="_platform_"` without platform JWT is rejected

---

### ISO-3 — Cross-Tenant Leakage Tests

All must run in CI on every PR touching `iam/`, `authz/`, or `middleware/`:
- [x] AZ-ISO-001: policy in dom-1 does not match request for dom-2
- [x] AZ-ISO-002: role assignment in dom-1 does not grant access in dom-2
- [x] AZ-ISO-003: platform domain policy does not match tenant domain request
- [x] AZ-ISO-004: tenant domain role does not apply in portal domain
- [x] AZ-ISO-005: tenant domain role does not apply in API domain
- [x] AZ-ISO-010: two tenants with identical role names — no cross-tenant access
- [x] AZ-ISO-020: wildcard subject deny in dom-1 stays domain-scoped

---

## 7. Cache & Session Consistency Tasks

### CACHE-1 — Redis Session Contains Only Context

After SES-1 and SES-3 complete:
- [ ] Verify cached `ResolvedSession` contains: `UserID`, `TenantID`, `UserType`, `EntityScope`, `Configuration` (flags/settings/prefs), `ExpiresAt`, `IsActive` — nothing else
- [ ] Verify no `Permissions` field in serialized cache entry (`grep` or JSON marshal test)
- [ ] Cache TTL = `session.ExpiresAt - time.Now()` — sessions cannot outlive DB expiry

---

### CACHE-2 — Invalidation Matrix Wiring

**Context**: Invalidation matrix correctly documented in `10b-session-precomputation.md`. Flag/setting path implemented. Authz path (role/policy change) NOT wired.

| Trigger | Scope | Method | Status |
|---|---|---|---|
| Logout | Single session | `InvalidateBySession()` + Redis DELETE | ❌ BLOCK-2 |
| User suspended/terminated | All user sessions | `InvalidateByUser()` | ❌ BLOCK-3 |
| Role revoked | All user sessions | `InvalidateByUser()` | ❌ BLOCK-3 |
| Policy removed from domain | All tenant sessions | `InvalidateByTenant()` | ❌ BLOCK-3 |
| Module/resource flag toggled | All tenant sessions | `InvalidateByTenant()` | ✅ done |
| Session TTL (8h default) | Expired rows | Natural expiry | ✅ done |

Tasks:
- [ ] Wire BLOCK-2 (Logout Redis delete)
- [ ] Wire BLOCK-3 (role/policy change → session invalidation)
- [ ] Test: revoke role → validate session → next `Enforce()` returns 403 (not re-login)
- [ ] Test: no over-invalidation — unrelated tenant policy changes don't invalidate other tenant sessions

---

### CACHE-3 — API Key Revocation Timing

**Context**: `review.md` finding: "revoked API key remains valid for up to 5 minutes by design (cache TTL)." Known gap.

Tasks:
- [ ] Document the 5-minute revocation window explicitly in `17-security-considerations.md`
- [ ] For v1.0: accept this window — document as known limitation
- [ ] For critical revocations (compromised key): `InvalidateCache()` call + direct Redis eviction of API key's session entry
- [ ] `SEC-4` test: API key revocation eventually takes effect within TTL window

---

## 8. Removal / Deprecation Tasks

### DEL-1 — Delete from `domain/session.go`

- [ ] `Session.Permissions map[string]bool`
- [ ] `Session.RiskScore float64`
- [ ] `ResolvedSession.Can(permission string) bool`
- [ ] `ResolvedSession.CanDo(resource, action string) bool`
- [ ] Wildcard `"*"` sentinel logic

Keep: `ToPrincipal()`, `EntityScope`, `Configuration`, `FeatureEnabled()`, all Setting helpers, `IsPlatform()`, `IsPortal()`.

---

### DEL-2 — Delete from `service/session.go`

- [ ] `buildPermissions()` function
- [ ] All `authz.GetImplicitRoles()` calls from session service
- [ ] All `authz.GetPolicies()` calls from session service
- [ ] Local deny-override logic
- [ ] JIT bootstrap call (BLOCK-6)

---

### DEL-3 — Gate `internal/core/access/`

- [ ] `//go:build ignore` on all `.go` files in `internal/core/access/`
- [ ] `go build ./...` clean with no access imports
- [ ] Add `README.md` in `internal/core/access/`: reserved for v2.0 ABAC

---

### DEL-4 — Dead DB Queries

- [ ] Remove `permissions` from session queries (SES-4)
- [ ] Remove or mark queries used only by `access/` module: `db/queries/policies.sql` (any query referencing `000404`–`000413` tables)
- [ ] Mark migrations `000404`–`000413` as v2.0-reserved (comment header, do not drop)
- [ ] Run `make sqlc` after all query changes

---

## 9. Database & Migration Tasks

### DB-1 — Sessions Permissions Column Cleanup

- [ ] Migration `000XXX_sessions_drop_permissions.up.sql`:
  ```sql
  ALTER TABLE user_sessions DROP COLUMN IF EXISTS permissions;
  ```
- [ ] Down migration: re-add as `permissions JSONB NULL`
- [ ] Coordinate with SES-4 to ensure SQLC generates before migration applies

---

### DB-2 — casbin_rule Constraints Validation

- [ ] Unique constraint on `(ptype, v0, v1, v2, v3, v4, v5)` exists
- [ ] Index on `v1` (domain column) for `RemoveFilteredPolicy` performance
- [ ] No RLS on `casbin_rule` — confirmed by `pg_class.relrowsecurity = false`
- [ ] `application_role` has `SELECT, INSERT, UPDATE, DELETE` on `casbin_rule`

---

### DB-3 — role_assignments Constraints Validation

- [ ] Partial index on `expires_at WHERE expires_at IS NOT NULL` exists — `idx_role_assignments_expires`
- [ ] Index on `(subject, domain)` for `ListRoleAssignments` performance
- [ ] `is_active` column with default `TRUE`
- [ ] Constraint: `CHECK (expires_at IS NULL OR expires_at > created_at)` — prevents past expiry at assign time
- [ ] RLS policy: tenant sees only its domain's assignments

---

### DB-4 — Migration Rollback Safety

For all new migrations:
- [ ] Every `.up.sql` has matching `.down.sql`
- [ ] Down migrations tested: up → down → state returns to baseline
- [ ] New columns added as `NULLABLE` or with defaults — never `NOT NULL` without default on existing tables
- [ ] All migrations idempotent: `IF NOT EXISTS` / `IF EXISTS` guards throughout

---

## 10. Testing Strategy

Tests organized by layer. Every test maps to specific risk or audit finding. Tests marked `[uncomment]` exist in test files but are disabled — enable and fix, do not rewrite.

### T-UNIT — Unit Tests (no DB)

**Types & Helpers** (`types_test.go`, `errors_test.go`):
- [x] AZ-TYP-001 `TestSubjectBuilders` — PlatformSubject, TenantSubject, PortalSubject, APISubject
- [x] AZ-TYP-010 `TestDomainBuilders` — TenantDomain, PortalDomain, APIDomain, DomainPlatform
- [x] AZ-TYP-020 `TestAssignOpts` — WithExpiry, WithAssignedBy, WithDelegatedBy, multiple opts
- [x] AZ-TYP-030 `TestErrorString` — Error.Error() format
- [x] AZ-TYP-031 `TestSentinelErrors` — HTTP status codes (incl. ErrPolicyLimitExceeded=429)

**Session model post-refactor** (new, risk: BLOCK-1):
- [x] `TestResolvedSession_NoPermissions` — compile-time proof no Permissions field
- [x] `TestResolvedSession_ToPrincipal` — all actor types produce correct Subject+Domain
- [x] `TestResolvedSession_FeatureEnabled` — reads from Configuration.Flags; nil-safe
- [x] `TestResolvedSession_Configuration` — SettingString, SettingBool, SettingInt, SettingDecimal

---

### T-ADAPTER — Adapter Layer Tests (DB required)

From `testing.md` AZ-ADP-001 to AZ-ADP-050:

- [ ] AZ-ADP-001 `TestLoadPolicy_Empty`
- [ ] AZ-ADP-002 `TestLoadPolicy_PRules`
- [ ] AZ-ADP-003 `TestLoadPolicy_GRules`
- [ ] AZ-ADP-010 `TestSavePolicy_FullReplace`
- [ ] AZ-ADP-020 `TestAddPolicy_SingleInsert`
- [ ] AZ-ADP-021 `TestAddPolicy_Idempotent`
- [ ] AZ-ADP-022 `TestAddPolicies_BatchInsert`
- [ ] AZ-ADP-030 `TestRemovePolicy_ExactMatch`
- [ ] AZ-ADP-031 `TestRemovePolicy_NotFound`
- [ ] AZ-ADP-040 `TestRemoveFilteredPolicy_ByDomain`
- [ ] AZ-ADP-050 `TestRuleToValues`

**File**: `internal/core/iam/adapter_test.go`

---

### T-SERVICE — Service Layer Tests (DB required)

From `testing.md` AZ-SVC-001 to AZ-SVC-061:

**Constructor**:
- [ ] AZ-SVC-001 `TestNew_NilPool`
- [ ] AZ-SVC-002 `TestNew_NilLogger`
- [ ] AZ-SVC-003 `TestNew_Success` — confirm `SyncedEnforcer` initialized (BLOCK-4)

**Enforce**:
- [ ] AZ-SVC-010–013 `TestEnforce_InvalidRequest` — empty Subject, Domain, Object, Action
- [ ] AZ-SVC-020 `TestEnforce_Allow`
- [ ] AZ-SVC-021 `TestEnforce_WildcardObject`
- [ ] AZ-SVC-022 `TestEnforce_WildcardAction`
- [ ] AZ-SVC-030 `TestEnforce_DenyOverride`
- [ ] AZ-SVC-031 `TestEnforce_DefaultDeny`
- [ ] AZ-SVC-040 `TestEnforce_TriggersExpiry`

**EnforceBatch**:
- [ ] AZ-SVC-050 `TestEnforceBatch_Empty`
- [ ] AZ-SVC-051 `TestEnforceBatch_InvalidRequest`
- [ ] AZ-SVC-052 `TestEnforceBatch_MixedResults`

**InvalidateCache**:
- [ ] AZ-SVC-060 `TestInvalidateCache` — confirm `SyncedEnforcer.LoadPolicy()` called
- [ ] AZ-SVC-061 `TestInvalidateCache_Blocking`

**File**: `internal/core/iam/service_test.go`

---

### T-ROLES — Role Management Tests

> **Note**: `review.md` states "every single role-management DB integration test is commented out." Tests exist — uncomment, fix compilation errors, verify against current schema.

From `testing.md` AZ-ROLE-001 to AZ-ROLE-042 (`roles_test.go`):

- [x] AZ-ROLE-001 `TestAssignRole_PersistsToDB`
- [x] AZ-ROLE-002 `TestAssignRole_Idempotent`
- [x] AZ-ROLE-003 `TestAssignRole_WithExpiry_StoredCorrectly`
- [x] AZ-ROLE-004 `TestAssignRole_WithAssignedBy_StoredCorrectly`
- [x] AZ-ROLE-005 `TestAssignRole_WithDelegatedBy_StoredCorrectly`
- [x] AZ-ROLE-006 `TestAssignRole_ReactivatesPreviouslyRevoked`
- [x] AZ-ROLE-010 `TestRevokeRole_MarksInactiveAndRemovesGRule`
- [x] AZ-ROLE-011 `TestRevokeRole_NonExistent_NoError`
- [x] AZ-ROLE-020 `TestGetRoles_ReturnsSingleRole`
- [x] AZ-ROLE-021 `TestGetRoles_NoRoles_ReturnsEmptySlice_NotError`
- [x] AZ-ROLE-022 `TestGetRoles_ReturnsMultipleRoles`
- [x] AZ-ROLE-030 `TestHasRole_TrueAfterAssign`
- [x] AZ-ROLE-031 `TestHasRole_FalseAfterRevoke`
- [x] AZ-ROLE-032 `TestHasRole_IsDomainScoped`
- [x] AZ-ROLE-040 `TestGetAssignments_IncludesInactiveRows`
- [x] AZ-ROLE-041 `TestGetAssignments_ReturnsRow` (full metadata)
- [x] AZ-ROLE-042 `TestGetAssignments_EmptyForUnknownSubject`
- [x] AZ-ROLE-043 `TestEnforce_ExpiredRole_IsRevoked` (temporal)
- [x] AZ-ROLE-044 `TestEnforce_ActiveRole_NotExpired`
- [x] AZ-ROLE-045 `TestEnforce_PermanentRole_NeverExpires`
- [x] AZ-ROLE-046 `TestEnforce_MultipleExpiredRoles_AllRevoked`

---

### T-TEMP — Temporal Role Tests

From `testing.md` AZ-TEMP-001 to AZ-TEMP-010:

- [ ] AZ-TEMP-001 `TestRevokeExpiredRoles_LazyFire`
- [ ] AZ-TEMP-002 `TestRevokeExpiredRoles_NotExpired`
- [ ] AZ-TEMP-003 `TestRevokeExpiredRoles_Permanent`
- [ ] AZ-TEMP-004 `TestRevokeExpiredRoles_Multiple`
- [ ] AZ-TEMP-005 `TestRevokeExpiredRoles_IndexUsed` — EXPLAIN ANALYZE asserts partial index used
- [ ] AZ-TEMP-010 `TestRevokeExpiredRoles_NonFatalError` — DB failure → warn, Enforce continues

---

### T-MIDDLEWARE — Middleware Tests

From `testing.md` AZ-MID-001 to AZ-MID-040:

- [x] AZ-MID-001 `TestMiddleware_NoPrincipal` → 401
- [x] AZ-MID-002 `TestMiddleware_EmptyPrincipal` → 401
- [x] AZ-MID-010 `TestMiddleware_Forbidden` → 403
- [x] AZ-MID-011 `TestMiddleware_Allow` → calls Next()
- [x] AZ-MID-020 `TestMiddleware_ObjectExpansion` (with/without :id param)
- [x] AZ-MID-021 `TestMiddleware_NoIDParam` → plain object
- [ ] AZ-MID-030 `TestMiddleware_EnforceError` → 500 (not yet covered)
- [x] AZ-MID-040 `TestMiddleware_LocalsKey` (wrong key → 401)

**New — post-refactor middleware tests** (risk: BLOCK-1):
- [ ] `TestRequirePermission_CallsEnforce` — mock authz service, confirm `Enforce()` called
- [ ] `TestRequirePermission_NoSessionCan` — confirm `session.Can()` NOT called (method absent)
- [ ] `TestRequireFlag_UsesSessionConfiguration` — confirm flag check hits `session.FeatureEnabled()` (not Casbin)

**File**: `internal/core/iam/middleware_test.go`

---

### T-ISO — Domain Isolation Tests

From `testing.md` AZ-ISO-001 to AZ-ISO-020:

- [x] AZ-ISO-001 `TestDomainIsolation_PolicyDoesNotCrossDomain`
- [x] AZ-ISO-002 `TestDomainIsolation_RoleDoesNotCrossDomain`
- [x] AZ-ISO-003 `TestDomainIsolation_PlatformVsTenant`
- [x] AZ-ISO-004 `TestDomainIsolation_TenantVsPortal`
- [x] AZ-ISO-005 `TestDomainIsolation_TenantVsAPI`
- [x] AZ-ISO-010 `TestDomainIsolation_CrossTenant`
- [x] AZ-ISO-020 `TestDomainIsolation_WildcardSubjectDomainScoped`

**File**: `internal/core/iam/authz_enforce_test.go`

---

### T-INT — Integration Tests

From `testing.md` AZ-INT-001 to AZ-INT-031:

- [ ] AZ-INT-001 `TestFullRBACFlow`
- [ ] AZ-INT-002 `TestTerminationFlow`
- [ ] AZ-INT-003 `TestTemporalRoleFlow`
- [ ] AZ-INT-010 `TestMultiInstanceSync` — two `authz.New()` on same DB; `SyncedEnforcer` auto-reload closes gap within 30s
- [ ] AZ-INT-020 `TestReconciliation_DetectsOrphan` — runs Rule 8 queries from `20-business-rules-and-validation.md`
- [ ] AZ-INT-030 `TestRLS_RoleAssignments`
- [ ] AZ-INT-031 `TestRLS_CasbinRule_NoTenantFilter`

**New — invalidation integration tests** (risk: BLOCK-2, BLOCK-3):
- [ ] `TestLogout_EvictsRedis` — logout → `ValidateSession()` → expect 401 immediately
- [ ] `TestRevokeRole_InvalidatesSession` — revoke → `ValidateSession()` → next `Enforce()` → 403
- [ ] `TestRemovePolicy_InvalidatesTenantSessions` — domain policy removed → all tenant sessions expire
- [ ] `TestAssignRole_ImmediateEffect` — assign role → `Enforce()` returns true (in-memory update)
- [ ] `TestMFA_ConcurrentCompletion` — two concurrent `CompleteMFALogin` → only one session (SES-5)

---

### T-SEC — Security Tests

From `testing.md` AZ-SEC-001 to AZ-SEC-050:

- [ ] AZ-SEC-001 `TestNoEscalation_RoleAssignment`
- [ ] AZ-SEC-002 `TestSQLInjection_Subject`
- [ ] AZ-SEC-003 `TestSQLInjection_Object`
- [ ] AZ-SEC-010 `TestCrossTenant_Impossible`
- [ ] AZ-SEC-020 `TestEnforce_ReadOnly`
- [ ] AZ-SEC-030 `TestDenyCannotBeBypassed`
- [ ] AZ-SEC-040 `TestExpiredRole_OldTokenReplay`
- [ ] AZ-SEC-050 `TestPlatformDomain_NotInjectable`

**New — post-refactor security tests**:
- [x] `TestSystemRole_ImmutableFromTenantActor` — tenant actor cannot modify `_platform_` policies (AUTHZ-4)
- [x] `TestPolicyCountLimit_Enforced` — adding policy beyond limit returns ErrPolicyLimitExceeded (AUTHZ-5)
- [ ] `TestJITBootstrap_Removed` — login as roleless user does NOT auto-assign tenant_admin (BLOCK-6)
- [ ] `TestAssignRole_Guards` — guard1 (namespace), guard2 (cross-tenant), guard3 (delegation)

---

## 11. Security & Attack Surface Validation

### SEC-1 — Session Bypass Elimination
- [ ] BLOCK-1 complete: `Can()` / `CanDo()` deleted
- [ ] All middleware uses `Enforce()`: confirmed by `grep`
- [ ] Test: valid session without matching Casbin policy → 403

### SEC-2 — Privilege Escalation Prevention
- [ ] BLOCK-6 complete: JIT bootstrap removed
- [ ] AUTHZ-4 complete: system role / platform domain guards
- [ ] Delegation guard: `AssignRole` caller must hold the role they grant
- [ ] Tests: `TestAssignRole_Guards` (three guard scenarios)

### SEC-3 — Stale Session Exploitation Prevention
- [ ] BLOCK-2 complete: Logout evicts Redis
- [ ] BLOCK-3 complete: role revoke → InvalidateByUser
- [ ] Test: revoke → attempt resource within session lifetime → denied

### SEC-4 — API Key Lifecycle
- [ ] API key auth path calls `Enforce()` (not session.Can)
- [ ] API key domain = `APIDomain(tenantID)` — same isolation as session-based auth
- [ ] Revoked API key: document 5-min TTL window; for emergency: `InvalidateCache()` + Redis evict
- [ ] Test: revoked API key — access denied within 1 TTL window

### SEC-5 — MFA Enforcement
- [ ] SES-5 complete: GETDEL atomicity
- [ ] TOTP replay prevention: used codes cached 90s per `17-security-considerations.md`
- [ ] Test: replay of used TOTP code within window → rejected
- [ ] Test: concurrent `CompleteMFALogin` → only one session

### SEC-6 — Tenant Boundary & Policy Injection Prevention
- [ ] ISO-1 complete: RLS on all tenant-scoped tables
- [ ] AUTHZ-5 complete: policy count limit
- [ ] `AddPolicy` protected by `authzSvc.Middleware("policy", "create")` — only tenant-admin can add
- [ ] Domain for policy add derived from JWT, not request body
- [ ] Rate limit on policy management endpoints (T6: 10 policies/min/tenant-admin)
- [ ] Platform admin audit query scheduled: `SELECT v0, v1 FROM casbin_rule WHERE ptype='g' AND v2='_platform_'`

---

## 12. Performance & Scalability Tasks

### PERF-1 — Enforce Hot-Path
- [ ] `SyncedEnforcer.Enforce()` is in-memory — no DB on hot path
- [ ] `revokeExpiredRoles()` uses partial index — verify `idx_role_assignments_expires`
- [ ] Target: p99 Enforce latency < 1ms (AZ-PERF-001)
- [ ] `BenchmarkEnforce` — 10,000 calls, measure p50 and p99

### PERF-2 — Session Validation Hot-Path
- [ ] Cache-first: `ValidateToken()` hits Redis first, DB only on miss
- [ ] `UpdateLastSeen()` fire-and-forget goroutine — never blocks auth path
- [ ] Metric: `iam.session.cache_hit` counter — target >95%

### PERF-3 — Login Performance After Refactor
- [ ] `buildSession` parallel queries reduced from 5 to 4 (no permissions query)
- [ ] Expected login latency: ~4ms (was ~5–6ms per `10b-session-precomputation.md`)
- [ ] `BenchmarkLogin` — confirm no regression; new baseline documented

### PERF-4 — Policy Load at Startup
- [ ] `SyncedEnforcer` loads policy at init and every 30s
- [ ] `TestLoadPolicy_10kRules` < 200ms (AZ-PERF-003)
- [ ] `TestLoadPolicy_100kRules` < 1s, < 100MB (AZ-PERF-004)

### PERF-5 — Memory Footprint
- [ ] 1,000 tenants × 50 policies = 50,000 rules — memory increase < 25MB (AZ-PERF-020)
- [ ] Update `16-performance-and-caching.md` with actual benchmark results

---

## 13. Deployment Readiness Checklist

### Readiness Gates (all must be true)

- [ ] BLOCK-1 through BLOCK-6 completed and tested
- [ ] `grep -rn "\.Can\b\|\.CanDo\b" --include="*.go" internal/` returns zero hits
- [ ] `grep -rn "internal/core/access" --include="*.go" internal/ cmd/` returns zero hits
- [ ] All T-SEC tests pass
- [ ] All T-ISO tests pass
- [ ] All T-INT tests pass (including invalidation tests)
- [ ] All T-ROLES tests uncommented and passing
- [ ] `BenchmarkEnforce` p99 < 1ms
- [ ] `SyncedEnforcer.StartAutoLoadPolicy(30s)` active — multi-instance gap ≤ 30s
- [ ] Security checklist from `17-security-considerations.md` fully checked

### Rollback Strategy

- [ ] All new migrations have down migrations tested (DB-4)
- [ ] `permissions` column added back as nullable in down migration — no data loss
- [ ] `SyncedEnforcer` can be disabled via config; fallback: standard enforcer with manual reload
- [ ] Session invalidation changes are additive (Redis deletes are safe to add)

### Rollout Stages

**Stage 1 — Single instance, internal**
- Deploy all BLOCK fixes; no real user traffic
- Validate: all tests pass against staging DB
- Validate: grep confirms zero `session.Can()` call sites

**Stage 2 — Single instance, limited traffic (5%)**
- Monitor `iam.authz.enforce` metric — no unexpected spike
- Monitor `iam.session.login.failure` — watch for regressions
- Rollback trigger: auth failure rate > 1% or previously-passing routes return 403

**Stage 3 — Multi-instance**
- `SyncedEnforcer.StartAutoLoadPolicy(30s)` active on all instances
- Test: revoke on instance A → within 30s → denied on instance B
- Monitor `iam.authz.policy_reload` counter — incrementing on schedule

**Stage 4 — Full traffic**
- All tenants on new IAM
- grep CI check confirms zero `session.Can()` remaining
- SLO: p99 auth latency < 50ms end-to-end

### Observability Requirements (required before Stage 2)

- [ ] `iam.auth.success` / `iam.auth.failure` counters (by reason)
- [ ] `iam.authz.enforce` counter (by result: allow, deny, error)
- [ ] `iam.authz.role.assigned` / `iam.authz.role.revoked` counters
- [ ] `iam.session.login.success` / `failure` counters
- [ ] `iam.session.cache_hit` counter
- [ ] `iam.authz.policy_reload` counter (with `method` label)
- [ ] `iam.authz.reconciliation.orphaned_assignments` gauge
- [ ] `iam.authz.policy_count` gauge per domain
- [ ] OTel spans: `Enforce`, `Login`, `ValidateSession`, `AssignRole`, `RevokeRole`

---

## 14. Final Delivery Milestones

### Phase 1 — Security Blockers

**Tasks**: BLOCK-1, BLOCK-2, BLOCK-3, BLOCK-4, BLOCK-5, BLOCK-6, SES-3

**Exit criteria**:
- `session.Can()` / `CanDo()` deleted — compile enforced
- `Logout()` deletes Redis key — `TestLogout_EvictsRedis` passes
- `RevokeRole()` calls `InvalidateByUser()` — `TestRevokeRole_InvalidatesSession` passes
- `SyncedEnforcer` in place — data race eliminated
- `access/` build-tagged — `go build ./...` clean
- JIT bootstrap removed — `TestJITBootstrap_Removed` passes

**Required tests passing**: T-SEC (SEC-1, SEC-2, SEC-3), all new invalidation tests
**Risk level**: HIGH — security and correctness blockers

---

### Phase 2 — Core RBAC Stabilization

**Tasks**: SES-1, SES-2, SES-4, SES-5, AUTHZ-1 through AUTHZ-7, ROLE-1 through ROLE-5, DEL-1 through DEL-4

**Exit criteria**:
- `grep "\.Can\b\|\.CanDo\b" internal/` returns zero
- `buildPermissions()` deleted; session build is 4-query parallel (no permissions)
- `RequirePermission` calls `Enforce()` — `TestRequirePermission_CallsEnforce` passes
- System role guards in place — `TestSystemRole_ImmutableFromTenantActor` passes
- Policy count limit enforced — `TestPolicyCountLimit_Enforced` passes
- Security event audit log emitting for ROLE_ASSIGNED, REVOKED, POLICY_ADDED, REMOVED
- All T-ROLES tests uncommented and passing
- All AZ-SVC-*, AZ-TEMP-*, AZ-MID-* tests pass

**Required tests passing**: T-UNIT, T-SERVICE, T-ROLES (uncommented), T-TEMP, T-MIDDLEWARE
**Risk level**: MEDIUM — bounded refactor

---

### Phase 3 — Multi-Instance Readiness

**Tasks**: BLOCK-4 verification, CACHE-1, CACHE-2, CACHE-3, ISO-1, ISO-2, ISO-3, DB-1 through DB-4

**Exit criteria**:
- `SyncedEnforcer.StartAutoLoadPolicy(30s)` verified across two instances — `TestMultiInstanceSync` passes
- Full invalidation matrix wired and tested
- RLS audit complete — all tables verified
- Cross-tenant isolation tests pass (AZ-ISO-001 to AZ-ISO-020)
- `permissions` column migration applied and rolled back safely

**Required tests passing**: T-ISO, T-INT (full suite including invalidation), `TestRLS_*`
**Risk level**: MEDIUM — distributed systems correctness

---

### Phase 4 — Production Hardening

**Tasks**: SEC-1 through SEC-6, PERF-1 through PERF-5, Section 15 (documentation updates)

**Exit criteria**:
- All AZ-SEC-* tests pass
- `BenchmarkEnforce` p99 < 1ms
- `TestLoadPolicy_10kRules` < 200ms
- All documentation files updated (Section 15 complete)
- Performance benchmarks recorded in `16-performance-and-caching.md`

**Required tests passing**: T-SEC, T-ADAPTER, T-INT (full suite)
**Risk level**: LOW-MEDIUM — cleanup and hardening

---

### Phase 5 — Final Validation + Sign-Off

**Exit criteria**:
- Every test in `testing.md` has `[x]` status (passing)
- All readiness gates in Section 13 checked
- Stage 1 and Stage 2 rollout complete without rollback
- `grep -rn "\.Can\b\|\.CanDo\b" --include="*.go" internal/` returns zero
- `grep -rn "internal/core/access" --include="*.go" internal/ cmd/` returns zero
- Audit score: ≤ 3/10 (down from 7.5/10)

**Required tests passing**: FULL SUITE — all AZ-* test IDs
**Risk level**: LOW

---

## 15. Documentation Update Tasks

These documents describe OLD two-path architecture. Must be updated to reflect single Casbin enforcement path before or alongside Phase 2.

### DOC-1 — Update `10b-session-precomputation.md`

**Current content**: Describes 5-query parallel `buildSession()` including permissions query. Shows `session.Can()` / `session.CanDo()` as enforcement mechanism. Documents pre-computed permission map as primary authorization fast-path.

**Required updates**:
- [ ] Remove permissions from 5-query parallel block (reduce to 4 queries)
- [ ] Remove `Permissions map[string]bool` from `ResolvedSession` code block
- [ ] Remove `Can()` / `CanDo()` methods from `ResolvedSession` code block
- [ ] Update "What Gets Built at Login" section — flags/settings/prefs/entity-scope only
- [ ] Update "In-Handler Usage" example — replace `session.CanDo()` with `authzSvc.Enforce()`
- [ ] Retain invalidation matrix (correct and fully valid)
- [ ] Add note: "Permission checks use Casbin.Enforce() — O(memory), no DB. Feature flags and settings remain O(1) from session.Configuration."

---

### DOC-2 — Update `12b-http-middleware.md`

**Current content**: Shows `RequirePermission` using `session.CanDo()` and documents "two-path authorization" table (fast path vs Casbin path).

**Required updates**:
- [ ] Rewrite `RequirePermission` implementation to call `authzSvc.Enforce()` (match BLOCK-1 code)
- [ ] Remove "two-path authorization" table — single Casbin path now
- [ ] Update router example — `RequirePermission` now takes `authzSvc` as parameter
- [ ] Keep `RequireFlag` unchanged — correctly reads `session.FeatureEnabled()` (context)
- [ ] Update "Both read from pre-computed session" note — permissions no longer pre-computed

---

### DOC-3 — Update `17-security-considerations.md`

- [ ] T8 section: mark `casbin.SyncedEnforcer` as DONE once BLOCK-4 complete
- [ ] Add API key revocation window (5 min) to threat mitigations table (CACHE-3)
- [ ] Add policy count limit as T6 mitigation (AUTHZ-5)

---

### DOC-4 — Update `16-performance-and-caching.md`

- [ ] Update login latency: "5-6ms (5 parallel queries)" → "4ms (4 parallel queries, no permissions)"
- [ ] Update `ResolvedSession` description — remove Permissions from "pre-computed" list
- [ ] Add section: "Casbin Enforce Latency" with benchmark results from PERF-1
- [ ] Update "Two-layer permission architecture" to reflect single Casbin path

---

### DOC-5 — Update `11-middleware-and-http.md`

- [ ] Reflect that `RequirePermission` calls `Enforce()`, not session map
- [ ] Update any code examples showing `session.Can()` or `session.CanDo()`

---

## 16. Documentation Expansion Tasks (2026-05-11)

Tracks operational/admin/user-facing architecture documentation added in second documentation pass (after first-round reconciliation complete).

### Completed Tasks

- [DONE] Tenant administration operational guide added (`23-tenant-administration.md`)
  - Bootstrap flow, `BootstrapTenantAdmin` behavior, `SeedDefaultRoles`
  - What tenant admins can and cannot do
  - Custom role creation, hierarchy, immutable system roles
  - User lifecycle: create → assign → entity scope → activate/deactivate
  - Separation of platform vs tenant authority

- [DONE] Platform administration guide added (`24-platform-administration.md`)
  - Platform vs tenant vs service actor types
  - Platform roles (super_admin, support, operator, billing, compliance)
  - Super admin bootstrap mechanism and emergency recovery
  - Platform operational boundaries per role
  - Casbin domain isolation for `_platform_`

- [DONE] Entity scope guide added (`25-user-entity-scope.md`)
  - Three scope types: ALL, SUBTREE, ENTITY_ONLY
  - Worked examples: retail, finance, HR, airline, forecourt
  - How enforcement works (application-layer WHERE clauses)
  - Assignment, changing scope, default scope behavior

- [DONE] API keys and service accounts guide added (`26-api-keys-and-service-accounts.md`)
  - Key lifecycle: create, prefix+hash format, scopes (PARTIAL), expiry, revocation
  - 5-minute cache TTL limitation documented
  - Service account model and M2M recommended role structure
  - Security guarantees and operational best practices

- [DONE] Resource/action ownership model documented (`27-resource-action-ownership.md`)
  - Module-owned resources — no central compile-time registry
  - Naming conventions: `module.submodule.resource` dot notation
  - Standard action verbs and wildcard rules
  - Module examples: finance, inventory, CRM, airline, forecourt
  - Feature flags vs tenant settings vs user preferences — clear distinctions

- [DONE] Open security items documented in `17-security-considerations.md`
  - AUTHZ-4: Platform domain write guard — NOT IMPLEMENTED
  - AUTHZ-5: Policy count limit — NOT IMPLEMENTED
  - DB-1: Unused `permissions` column in user_sessions
  - AUTHZ-7: Persistent audit log — OTel spans only

### Open Tasks from This Expansion

- [OPEN] **AUTHZ-4**: Implement platform domain write guard in `AuthzService.AddPolicy()` and `AssignRole()`
  - See `service/authz.go`
  - Guard: `if targetDomain == DomainPlatform && !caller.IsPlatform() { return ErrForbidden }`
  - Requires service-layer access to caller's actor type

- [OPEN] **AUTHZ-5**: Implement policy count limit per tenant domain
  - In `AddPolicy()`: check `len(GetPolicies(domain))` before inserting
  - Configurable limit (suggest default: 10,000 rules per domain)
  - Return `ErrPolicyLimitExceeded` when limit reached
  - Expose limit as tenant setting or platform config

- [DONE] **DB-1**: Migration `001008_sessions_drop_permissions` written
  - Up: `ALTER TABLE user_sessions DROP COLUMN IF EXISTS permissions;`
  - Down: re-adds as `permissions JSONB NULL`
  - After applying: run `make sqlc` to regenerate models (SES-4)

- [OPEN] **AUTHZ-7**: Design and implement persistent audit log for authorization decisions
  - Design `audit_events` table (see `17-security-considerations.md` for schema draft)
  - Instrument `AssignRole`, `RevokeRole`, `AddPolicy`, `RemovePolicy` to emit events
  - Table must be append-only (no UPDATE/DELETE for application_role)
  - Long retention (90+ days) for compliance evidence

---

## Appendix: File Reference Map

| Task | Primary Files |
|---|---|
| BLOCK-1 | `domain/session.go`, `service/session.go`, middleware layer |
| BLOCK-2 | `repository/session.go` |
| BLOCK-3 | `service/authz.go`, `service/session.go` |
| BLOCK-4 | `service/authz.go` |
| BLOCK-5 | `internal/core/access/**` |
| BLOCK-6 | `service/session.go`, `seed.go`, tenant provisioning |
| SES-1..4 | `domain/session.go`, `service/session.go`, `repository/session.go`, `db/queries/sessions.sql` |
| SES-5 | `repository/session.go` |
| AUTHZ-1..7 | `service/authz.go`, middleware files, `domain/errors.go` |
| ROLE-1..5 | `service/authz.go`, `seed.go` |
| ISO-1..3 | `db/migration/000056*`, `000107*`, `000064*`, middleware |
| CACHE-1..3 | `repository/session.go`, `service/authz.go` |
| DEL-1..4 | `domain/session.go`, `service/session.go`, `internal/core/access/**`, `db/queries/` |
| DB-1..4 | `db/migration/`, `db/queries/sessions.sql` |
| T-* | `internal/core/iam/*_test.go` |
| DOC-1..5 | `docs/reference/modules/iam/10b*`, `12b*`, `16*`, `17*`, `11*` |

---

*Single source of execution truth for IAM v1.0. Update checkboxes as work completes.*