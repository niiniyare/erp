# API Layer Refactoring & Audit — Task Tracker

> **How to use:** Mark `[ ]` → `[x]` when done. Add date: `[x] (2026-05-20)`.
> New sessions: start at first unchecked task, re-read **Cross-Cutting Foundations** first.
>
> **Severity legend:** SEV-1 = system-breaking, SEV-2 = tenant isolation breach,
> SEV-3 = context integrity, SEV-4 = design contradiction, SEV-5 = infrastructure

---

## Cross-Cutting Foundations

### 1 · Error Handling — `awo.so/internal/shared/errors`

All handler errors must go through `BusinessError`. Never return raw `fmt.Errorf` or
`errors.New` to the client.

```go
// Validation error (bad input)
return sharedErrors.NewBusinessError("INVALID_FOO", "foo is not valid").
    WithHTTPStatus(fiber.StatusBadRequest).
    WithCategory(sharedErrors.CategoryValidation).
    WithDetail("field", "foo").
    WithSuggestion("Valid values: a, b, c")

// HTTP mapping — always use errors.As, never type switch
httpErr := sharedErrors.ToHTTPError(err)
c.Status(httpErr.Status).JSON(httpErr)
```

### 2 · Four-Step Handler Pattern

```go
func (h *XHandler) DoThing(c *fiber.Ctx) error {
    // 1. Observability
    ctx, span := h.tracer.StartSpan(c.UserContext(), "module.DoThing")
    defer span.End()
    c.SetUserContext(ctx)

    // 2. Parse & validate
    // ...

    // 3. Delegate
    result, err := h.services.X.DoThing(ctx, ...)
    if err != nil {
        span.RecordError(err)
        return h.fail(c, err)
    }

    // 4. Response
    return h.ok200(c, result)
}
```

### 3 · Tenant & User Context (UPDATED — see audit findings below)

**DO NOT** rely on `SetTenantContextFromCtx` from middleware to enforce RLS.
All service-layer queries that must be tenant-scoped **must** be wrapped in
`store.WithTenantFromCtx(ctx, fn)` or `store.BeginTxWithTenantFromCtx(ctx)`.
The middleware sets the tenant ID in the Go context (`shared.WithTenantID`) — that
is the source of truth for tenant identity in the request. The DB RLS enforcement
only works when queries run inside a tenant transaction.

Auth middleware injects `tenant_id` (as `uuid.UUID`) and `user_id` (as `string`)
into Fiber Locals. Use `extractUserID(c)` — never assert Locals directly to `uuid.UUID`.

---

## Phase 1: Finance Handler Refactor

- [x] **TASK 1 — Split finance handler into 11 files** (2026-05-19)
  - Status: _Complete_

- [x] **TASK 2 — Fix ProviderSet (wrong import)** (2026-05-19)
  - Was: `option.WithoutAuthentication` (Google API option)
  - Fixed: `wire.NewSet(NewFinanceHandler)` with `github.com/google/wire`
  - Status: _Complete_

- [x] **TASK 3 — Fix `itoa` reimplementation** (2026-05-19)
  - Was: buggy loop, silent `""` for negatives
  - Fixed: `func itoa(i int) string { return strconv.Itoa(i) }`
  - Status: _Complete_

- [x] **TASK 4 — Add `newValidationError` to helpers** (2026-05-19)
  - Was: called in `transactions.go`, `tax.go`, `reconciliation.go` but never defined
  - Fixed: added to `helpers.go`
  - Status: _Complete_

---

## Phase 2: SEV-1 — Authentication & Tenant Isolation (System-Breaking)

> These four findings were discovered in the full runtime audit (2026-05-20).
> Nothing in Phase 3+ is safe to deploy without these resolved.

- [ ] **TASK 20 — Implement JWT authentication middleware (SEV-1-A)**
  - **Problem:** `ConfigureAPIRoutes` in `route_security.go` has auth/JWT/tenant middleware
    as commented-out stubs — only log statements:
    ```go
    // router.Use(m.createAuthMiddleware())   ← NEVER RUNS
    m.logger.Info("Auth middleware would be applied here")
    ```
    Every API endpoint is open to anonymous requests. No user identity is established.
  - **Fix:**
    1. Implement `createAuthMiddleware()` on `RouteSecurityManager` — parse and verify
       JWT from `Authorization: Bearer <token>` header
    2. On success, inject into Fiber Locals: `c.Locals("user_id", claims.UserID.String())`
       and `c.Locals("tenant_id", claims.TenantID)` as strings (matches `extractUserID`)
    3. Uncomment and wire `router.Use(m.createAuthMiddleware())` in `ConfigureAPIRoutes`
    4. Same for `TenantMiddleware` — it is implemented in `tenant.go` but **never registered**
       via `ConfigureAPIRoutes`
  - Files: `internal/api/middleware/route_security.go`, `internal/api/middleware/tenant.go`
  - **Blocks:** T5, T21, T22, T23 — every security task depends on auth existing
  - Status: _Not started_

- [ ] **TASK 21 — Fix `user_id` never injected — audit trail poisoned (SEV-1-B)**
  - **Problem:** `extractUserID(c)` reads `c.Locals("user_id").(string)`. No middleware
    sets `user_id` in Locals today. Result: every `byUserID` argument in budget submit/
    approve/reject, period status change, and transaction approve receives `uuid.Nil`.
    Audit trail is permanently corrupted — no record of who performed any approval action.
  - **Fix:**
    1. Auth middleware (T20) must set `c.Locals("user_id", claims.UserID.String())`
    2. Service layer must reject `uuid.Nil` as a `byUserID` argument with a descriptive error
       rather than silently recording a zero UUID
  - **Depends on:** T20
  - Status: _Not started_

- [ ] **TASK 22 — Fix RLS connection-pool timing bug (SEV-1-C)**
  - **Problem:** `TenantMiddleware` calls `store.SetTenantContextFromCtx(ctx)` which
    executes `SELECT set_tenant_context($1)` against `s.connPool` — a pool, not a
    transaction. pgxpool acquires one connection, runs the query, then **releases it**.
    `set_config('app.current_tenant_id', ..., TRUE)` is transaction-local; outside a
    transaction it is session-local for that one connection. The next query in the
    same request gets a **different** pool connection where the tenant context is not set.
    RLS policy sees `current_tenant_id()` = NULL and may return all rows.
    ```
    Middleware: conn-A → set_tenant_context(T1) → release conn-A
    Handler query: conn-B → current_tenant_id() = NULL → RLS bypass
    ```
  - **Fix:**
    1. Remove `config.Store.SetTenantContextFromCtx(ctx)` from `TenantMiddleware` — this
       call is not safe and gives false confidence
    2. All service-layer operations must wrap queries in `store.WithTenantFromCtx(ctx, fn)`
       or `store.BeginTxWithTenantFromCtx(ctx)` — these open a transaction on a single
       connection and call `setTenantContext` within that transaction
    3. Audit all `internal/core/*/service/*.go` files to verify every DB call that touches
       tenant data is inside a `WithTenantFromCtx` block, not a raw `store.QueryXxx(ctx, ...)`
    4. `SetTenantContextFromCtx` on the `Store` interface should be marked
       `// Deprecated: unsafe outside transactions. Use WithTenantFromCtx.`
  - Files: `internal/api/middleware/tenant.go:306`, `db/sqlc/store.go`,
    all service files under `internal/core/`
  - Status: _Not started_

- [ ] **TASK 23 — Fix observability middleware type assertion panic (SEV-1-D)**
  - **Problem:** `observability.go:215` does `tenantID.(string)` but `TenantMiddleware`
    stores `tenantID` as `uuid.UUID` at `tenant.go:285`. This panics at runtime on every
    request once `TenantMiddleware` is wired in (T20).
    Same issue at line 218 for `user_id` — same panic pattern.
    ```go
    // PANICS — tenantID is uuid.UUID, not string
    span.SetAttributes(attribute.String("tenant.id", tenantID.(string)))
    ```
  - **Fix:**
    ```go
    if tenantID := c.Locals("tenant_id"); tenantID != nil {
        span.SetAttributes(attribute.String("tenant.id", fmt.Sprintf("%v", tenantID)))
    }
    if userID := c.Locals("user_id"); userID != nil {
        span.SetAttributes(attribute.String("user.id", fmt.Sprintf("%v", userID)))
    }
    ```
    Or use a typed helper that handles both `string` and `uuid.UUID` gracefully.
  - File: `internal/api/middleware/observability.go:215,218`
  - **Must fix before T20** — wiring auth will trigger the panic immediately
  - Status: _Not started_

---

## Phase 3: SEV-2 — Tenant Isolation Breaches

- [ ] **TASK 24 — Remove tenant ID from query parameter (SEV-2-A)**
  - **Problem:** `extractTenantID` in `tenant.go:346` accepts `?tenant_id=<UUID>` as
    priority-2 tenant source. Any client can append this to any URL and claim a
    different tenant's identity. No user-tenant binding check exists.
    ```go
    // REMOVE THIS BLOCK ENTIRELY
    if tenantID := c.Query("tenant_id"); tenantID != "" {
        return tenantID, nil
    }
    ```
  - **Fix:** Remove the query parameter path. Tenant identity must come from either:
    a) JWT claims (`sub_tenant` or equivalent — preferred), or
    b) Verified subdomain extraction (priority 3 in current code)
    The `X-Tenant-ID` header (priority 1) is acceptable only if the JWT also contains
    the tenant and the header is validated against the claim.
  - File: `internal/api/middleware/tenant.go:346-355`
  - Status: _Not started_

- [ ] **TASK 25 — Bind tenant identity to JWT claims, not client headers (SEV-2-B)**
  - **Problem:** `X-Tenant-ID` header is accepted as the primary tenant source. An
    authenticated user from tenant A can set `X-Tenant-ID: <tenant-B-uuid>` and pass
    the tenant existence/active check. The middleware never verifies the user belongs
    to the claimed tenant.
  - **Fix:**
    1. JWT token must include `tenant_id` as a signed claim (set at login time)
    2. Auth middleware (T20) must extract `tenant_id` from JWT claims and inject it into
       context — **not** accept it from `X-Tenant-ID` header
    3. `extractTenantID` should be eliminated entirely; tenant comes from verified JWT only
    4. For multi-tenant admin routes (platform users managing multiple tenants), use a
       separate endpoint pattern with explicit platform-admin authorization
  - **Depends on:** T20
  - Status: _Not started_

- [ ] **TASK 26 — Add distributed cache invalidation for tenant suspension (SEV-2-C)**
  - **Problem:** `TenantCache` in `tenant.go` is an in-process `sync.RWMutex` map with
    5-minute TTL. Suspending a tenant takes up to 5 minutes per server instance to
    take effect. In a multi-instance deployment, each instance has its own cache.
  - **Fix:** On tenant status change (suspend/archive), publish an invalidation event
    (Redis pub/sub or existing `cache.Service.DeletePattern`) that all instances consume.
    Each instance calls `tenantCache.Invalidate(tenantID)` on receipt.
  - Files: `internal/api/middleware/tenant.go`, `internal/core/tenant/service.go`
  - Status: _Not started_

---

## Phase 4: SEV-3 — Context Propagation Integrity

- [ ] **TASK 27 — Fix `GetUserIDPtr` returns non-nil pointer to zero UUID (SEV-3-B)**
  - **Problem:** `internal/shared/context.go:79-82`:
    ```go
    func GetUserIDPtr(ctx context.Context) *uuid.UUID {
        userID, _ := ctx.Value(UserIDKey).(uuid.UUID)
        return &userID  // always non-nil, even when userID == uuid.Nil
    }
    ```
    Callers checking `if ptr == nil` will incorrectly think a user is present.
    Service code using `GetUserIDPtr` for optional user scoping will silently associate
    operations with `uuid.Nil`.
  - **Fix:** Match the `GetEntityIDPtr` pattern:
    ```go
    func GetUserIDPtr(ctx context.Context) *uuid.UUID {
        userID, ok := GetUserID(ctx)
        if !ok || userID == uuid.Nil {
            return nil
        }
        return &userID
    }
    ```
  - File: `internal/shared/context.go:79-82`
  - Status: _Not started_

- [ ] **TASK 28 — Mark legacy context fallback functions as broken (SEV-3-A)**
  - **Problem:** `GetTenantIDFromContext` fallback at `tenant.go:594` does
    `ctx.Value(TenantIDKey).(string)` using the untyped string constant. The value is
    stored with a typed `contextKey` by `shared.WithTenantID`. The fallback never finds
    the value — it is silently dead code, not a working fallback.
  - **Fix:** Replace the TODO/Deprecated comment with `// BROKEN: ctx.Value with untyped
    key never matches typed contextKey. Do not use.` and remove the fallback branch.
    Any caller must migrate to `GetTenantID(c)` (Fiber) or `shared.GetTenantID(ctx)`.
  - File: `internal/api/middleware/tenant.go:587-596`
  - Status: _Not started_

---

## Phase 5: SEV-4 — Design Contradictions

- [ ] **TASK 5 — Enforce authorization in finance handlers (SEV-4-C / C-02)**
  - **Problem:** All `Permission: finance.X.Y` comments are documentation only.
    No permission check is performed. `FinanceHandler` has no authz field.
  - **Fix:** Inject `authz.Service` into `FinanceHandler`; call `h.authz.Require(ctx,
    "finance.budgets.create")` at step 2 of each handler.
  - **Depends on:** T20 (auth must exist before authz is meaningful)
  - Files: `handler.go`, all finance handler files
  - Status: _Not started_

- [ ] **TASK 6 — Remove client-controlled `checks_passed` flag (SEV-4 / C-04)**
  - **Problem:** `changePeriodStatusRequest.ChecksPassed bool` lets client bypass
    server-side period-close checks.
  - **Fix:** Remove field; service must run checks internally.
  - File: `fiscal_periods.go`, `internal/core/finance/service/period.go`
  - Status: _Not started_

- [ ] **TASK 29 — Deprecate `SetTenantContextFromCtx` on Store interface (SEV-4-A)**
  - **Problem:** The Store interface labels `SetTenantContextFromCtx` as "preferred"
    alongside `WithTenantFromCtx`. They are NOT equivalent. The former is unsafe
    outside transactions (see T22). The naming implies safety it does not provide.
  - **Fix:**
    1. Add `// Deprecated: unsafe with connection pools. Use WithTenantFromCtx.`
    2. Add a `_ = SetTenantContextFromCtx` guard in a `_test.go` that forces
       any future caller to acknowledge the deprecation
  - File: `db/sqlc/store.go:22`
  - **Depends on:** T22 completion (all call sites migrated first)
  - Status: _Not started_

---

## Phase 6: Input Validation & Limits

- [ ] **TASK 7 — Add pagination to unbounded list endpoints (M-06)**
  - **Problem:** `ListFiscalYears`, `ListBudgets`, `ListCostCenters` — no limit.
  - **Fix:** Apply `h.pagination(c)`, pass to service filter struct.
  - Files: `fiscal_periods.go`, `budgets.go`, `cost_centers.go` + service/repo
  - Status: _Not started_

- [ ] **TASK 8 — Verify and cap exchange rate history limit (H-05)**
  - **Fix:** Read `exchange_rates.go`; cap at 500 if not already.
  - Status: _Not started_

---

## Phase 7: Per-Handler Audit

- [ ] **TASK 9 — Audit `accounts.go`**
- [ ] **TASK 10 — Audit `transactions.go`** — verify `PostTransaction(ctx, txID, nil)` nil user intent
- [ ] **TASK 11 — Audit `cost_centers.go`** — hierarchy bounds
- [ ] **TASK 12 — Audit `tax.go`**
- [ ] **TASK 13 — Audit `reconciliation.go`** — file upload size limit
- [ ] **TASK 14 — Audit `exchange_rates.go`**

---

## Phase 8: Other Handler Packages

- [ ] **TASK 15 — Audit `internal/api/handlers/iam/`**
- [ ] **TASK 16 — Audit `internal/api/handlers/tenant/`**
- [ ] **TASK 17 — Audit remaining handler packages**

---

## Phase 9: Infrastructure

- [ ] **TASK 30 — Fix `generateRequestID` uniqueness (SEV-5-C)**
  - **Problem:** `observability.go:418` uses `time.Now().UnixNano()` — not unique under
    concurrent requests on the same machine.
  - **Fix:** `return uuid.New().String()` (or rely on `requestid.New()` middleware already
    applied in `ConfigureAPIRoutes` — check if it's redundant first).
  - File: `internal/api/middleware/observability.go:418`
  - Status: _Not started_

- [ ] **TASK 31 — Wire distributed rate limiting (SEV-5-A)**
  - **Problem:** `CreateRateLimitMiddleware(&rateLimitConfig, nil, m.logger)` — nil cache
    means in-process only. Multi-instance deployments can bypass per-instance limits.
  - **Fix:** Pass `cache.Service` as the rate limit backend.
  - File: `internal/api/middleware/route_security.go:208`
  - Status: _Not started_

- [ ] **TASK 19 — Verify CSRF chain is complete**
  - Server sets non-HttpOnly `csrf_token` cookie ✓
  - AMIS fetcher sends `X-Csrf-Token` header ✓
  - Verify CSRF middleware is wired for all mutating routes
  - File: `route_security.go`, `internal/api/routes.go`
  - Status: _Not started_

---

## Parallel Execution Map

| Track | Tasks | Can Start When |
|-------|-------|----------------|
| SEV-1 runtime fixes | T23 | Now — independent of auth |
| SEV-1 auth implementation | T20 | Now |
| SEV-1 user_id / audit trail | T21 | T20 done |
| SEV-1 RLS pool fix | T22 | T20 done (need auth to test) |
| SEV-2 tenant spoofing | T24 | Now |
| SEV-2 JWT-bound tenant | T25 | T20 done |
| SEV-2 cache invalidation | T26 | Now |
| SEV-3 context fixes | T27, T28 | Now |
| SEV-4 authz enforcement | T5 | T20 done |
| SEV-4 ChecksPassed | T6 | Now |
| SEV-4 Store deprecation | T29 | T22 done |
| Pagination / limits | T7, T8 | Now |
| Per-handler audit | T9–T14 | T4 done |
| Other packages | T15–T17 | T9–T14 done |
| Infrastructure | T30, T31, T19 | Now |

---

## Current Blockers

- **T20 blocks T5, T21, T22, T25** — auth middleware must exist before auth-dependent work

---

## Audit Checklist (use per handler file)

```
- [ ] Pagination on all list endpoints (offset + limit, capped)
- [ ] Authorization check at top of each handler (not just in comments)
- [ ] All UUID params go through parseUUID()
- [ ] All monetary amounts use decimalFromString(), not float64
- [ ] Enum/code values validated with descriptive error before service call
- [ ] span.RecordError(err) on every error path
- [ ] No business logic in handler (pure delegation to service)
- [ ] Correct HTTP status codes (201 for create, 204 for delete, 200 for others)
- [ ] User ID extracted via extractUserID(c), not direct Locals assertion
- [ ] Service queries wrapped in WithTenantFromCtx (not raw store.QueryXxx)
```

---

## Quick Resume Guide

1. Find first unchecked `[ ]` task
2. Re-read **Cross-Cutting Foundations** before every session
3. **Start with T23** (panic fix) then **T20** (auth) — nothing else is safe to ship first
4. For audit tasks (T9–T17), use the Audit Checklist above
5. When done: `[ ]` → `[x]`, add date

**Current priority order:** T23 → T20 → T21+T22+T24 (parallel) → T25+T26+T27+T28 → T5 → T6 → T7+T8 → T9–T17
