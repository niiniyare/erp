 API Runtime Security Audit Report

  Date: 2026-05-19 | Auditor: Architecture Compliance Review

  ---
  EXECUTIVE SUMMARY

  The system has four SEV-1 production blockers that must be resolved before any tenant
  data exists in the system. The tenant isolation chain has a critical gap:
  authentication middleware is stub-only, RLS context has a connection-pool timing bug,
  and a type assertion will cause runtime panics. The underlying DB transaction
  architecture is sound, but the middleware layer does not correctly exercise it.

  ---
  SEV-1 — SYSTEM-BREAKING (Fix before any production data)

  SEV-1-A · Authentication Middleware Is Stub Code

  File: internal/api/middleware/route_security.go:211-233

  if m.config.API.RequireAuth {
      // router.Use(m.createAuthMiddleware())   ← COMMENTED OUT
      if m.logger != nil {
          m.logger.Info("Auth middleware would be applied here")  ← ONLY LOGS
      }
  }
  if m.config.API.ValidateJWT {
      // router.Use(m.createJWTMiddleware())    ← COMMENTED OUT
      if m.logger != nil {
          m.logger.Info("JWT middleware would be applied here")
      }
  }
  if m.config.API.RequireTenant {
      // This would use your existing tenant middleware ← COMMENTED OUT
      if m.logger != nil {
          m.logger.Info("Tenant middleware would be applied here")
      }
  }

  ConfigureAPIRoutes — the function that wires all security to the API route group —
  contains no actual authentication, no JWT validation, and no tenant middleware. Flags
  are read from config, then only logged. Every API endpoint is open to anonymous
  requests.

  TenantMiddleware exists as a functional implementation in tenant.go but it is never
  registered via ConfigureAPIRoutes.

  Impact: Complete authentication bypass. All endpoints accessible without credentials.

  ---
  SEV-1-B · user_id Is Never Injected — Always uuid.Nil

  Files: helpers.go:101, tenant.go, all of middleware/

  extractUserID(c) reads c.Locals("user_id").(string). No middleware in the codebase
  sets user_id into Fiber Locals. Because auth middleware is stub code (SEV-1-A), the
  JWT is never parsed and user identity is never injected.

  Every workflow operation that passes byUserID to a service — budget
  submit/approve/reject, period status change, transaction approve — silently receives
  uuid.Nil. Audit trails are poisoned. Any approval-required action can be performed
  without a recorded actor.

  ---
  SEV-1-C · RLS Context Set on Wrong Connection (Pool Timing Bug)

  Files: db/sqlc/store.go:324-331, middleware/tenant.go:306

  TenantMiddleware calls:
  config.Store.SetTenantContextFromCtx(ctx)

  SetTenantContextFromCtx does:
  func (s *SQLStore) SetTenantContextFromCtx(ctx context.Context) error {
      tenantID, err := s.getTenantIDFromContext(ctx)
      return s.setTenantContext(ctx, s.connPool, tenantID)  // ← s.connPool, not a tx
  }

  func (s *SQLStore) setTenantContext(ctx context.Context, exec DBTX, tenantID
  uuid.UUID) error {
      _, err := exec.Exec(ctx, "SELECT set_tenant_context($1)", tenantID)
      // exec.Exec with a pool: acquires one connection, runs query, RELEASES it

  When exec is s.connPool, pgxpool.Exec acquires a connection, runs SELECT
  set_tenant_context($1), and releases the connection back to the pool. The
  set_config('app.current_tenant_id', ..., TRUE) written by that stored procedure uses
  is_local = TRUE, meaning it is transaction-local. Outside a transaction, PostgreSQL
  treats it as session-level for that specific connection.

  Once that connection is released, the next query in the same request can acquire a
  different connection from the pool where app.current_tenant_id has not been set. The
  RLS policy USING (tenant_id = current_tenant_id()::uuid) sees NULL, potentially
  returning all rows or no rows depending on the fallback.

  The correct API is WithTenantFromCtx or BeginTxWithTenantFromCtx, which opens a
  transaction on a specific connection, sets the tenant context within that transaction,
   and ensures all queries in fn use that same connection. The store comment on line 234
   confirms this:

  ▎ "Note: set_tenant_context() uses set_config(..., TRUE) — transaction-local scope —
  so context is already cleared on every COMMIT/ROLLBACK."

  This means SetTenantContextFromCtx as called from middleware does not reliably enforce
   RLS on subsequent handler queries unless every service call wraps queries in
  WithTenantFromCtx.

  Impact: Cross-tenant data leakage on any query not wrapped in a tenant transaction.

  ---
  SEV-1-D · Observability Middleware Type Assertion Panic

  File: middleware/observability.go:215

  if tenantID := c.Locals("tenant_id"); tenantID != nil {
      span.SetAttributes(attribute.String("tenant.id", tenantID.(string)))  // ← PANIC
  }

  TenantMiddleware stores tenantID as uuid.UUID:
  c.Locals(TenantIDKey, tenantID)  // tenant.go:285 — type is uuid.UUID

  The type assertion tenantID.(string) will panic at runtime on every request that has
  tenant context set, because the actual type is uuid.UUID, not string. This means the
  observability middleware crashes every authenticated request once TenantMiddleware is
  actually wired in.

  Same pattern at observability.go:218 for user_id — will panic once auth middleware
  sets it as uuid.UUID.

  ---
  SEV-2 — TENANT ISOLATION BREACH

  SEV-2-A · Tenant ID Accepted from Query Parameter

  File: middleware/tenant.go:346-355

  // Priority 2: Check query parameter (useful for webhooks, callbacks)
  if tenantID := c.Query("tenant_id"); tenantID != "" {

  Any client can append ?tenant_id=<OTHER_TENANT_UUID> to any URL and claim to be that
  tenant. The middleware validates the tenant exists and is active, but does not verify
  that the requesting user belongs to it. Combined with the absent auth (SEV-1-A), this
  is a complete tenant isolation bypass via URL manipulation.

  Even with auth in place, the comment says "NOTE: Consider removing this if query-based
   tenant ID is a security concern" — it should be removed unconditionally. Tenant
  identity must come from a verified source (JWT claim or subdomain), never from a
  client-supplied query parameter.

  SEV-2-B · No Tenant-User Binding Verification

  File: middleware/tenant.go:195-213

  extractTenantID accepts the claimed tenant ID from X-Tenant-ID header. The middleware
  validates that this tenant exists and is active, but does not verify that the
  authenticated user's JWT claims include membership in this tenant. A user from tenant
  A can set X-Tenant-ID: <tenant-B-uuid> in any request.

  The correct model: extract tenant_id from the verified JWT token, not from a client
  header. Headers are advisory; JWT claims are signed.

  SEV-2-C · Suspended Tenant Cache Window

  File: middleware/tenant.go:44-116

  TenantCache is an in-process sync.RWMutex map with a 5-minute TTL. When a tenant is
  suspended in the database, the middleware continues serving requests for up to 5
  minutes for any server instance that has the tenant cached as ACTIVE.

  In a multi-instance deployment, each instance has an independent cache. Suspension
  takes effect only when the TTL expires per instance.

  ---
  SEV-3 — CONTEXT PROPAGATION INTEGRITY

  SEV-3-A · Dual Context Stores — String vs Typed Key Collision Risk

  Files: middleware/tenant.go:23-27, internal/shared/context.go:19-23

  Two different key namespaces exist:

  ┌────────────────────────────────────┬───────────────────────────────┬───────────┐
  │               Store                │              Key              │  Value    │
  │                                    │                               │   type    │
  ├────────────────────────────────────┼───────────────────────────────┼───────────┤
  │ c.Locals(TenantIDKey, ...) (Fiber) │ "tenant_id" (untyped string   │ uuid.UUID │
  │                                    │ const)                        │           │
  ├────────────────────────────────────┼───────────────────────────────┼───────────┤
  │ shared.WithTenantID(ctx, ...) (Go  │ contextKey("tenant_id")       │ uuid.UUID │
  │ context)                           │ (typed)                       │           │
  └────────────────────────────────────┴───────────────────────────────┴───────────┘

  The middleware correctly populates both. The store reads from the typed Go context via
   shared.GetTenantID(ctx). Handlers read from Fiber Locals via c.Locals("tenant_id").
  These are separate stores — they don't interfere.

  Risk: code that tries ctx.Value("tenant_id") with an untyped string will miss the
  value (typed key prevents that). The legacy function GetTenantIDFromContext fallback
  (ctx.Value(TenantIDKey).(string)) uses the middleware's untyped string const — this
  will always miss because shared.WithTenantID uses a typed key. The fallback is dead
  code. Mark it as broken, not deprecated.

  SEV-3-B · GetUserIDPtr Returns Pointer to Zero Value

  File: internal/shared/context.go:79-82

  func GetUserIDPtr(ctx context.Context) *uuid.UUID {
      userID, _ := ctx.Value(UserIDKey).(uuid.UUID)
      return &userID  // ← always returns a non-nil pointer, even when userID ==
  uuid.Nil
  }

  When user ID is absent from context, this returns &uuid.Nil — a non-nil pointer to the
   zero UUID. Callers checking if ptr == nil will incorrectly conclude a user is
  present. Any service that uses GetUserIDPtr for optional user scoping will silently
  associate operations with uuid.Nil instead of treating them as "no user".

  Compare with GetEntityIDPtr which correctly returns nil when absent.

  ---
  SEV-4 — DESIGN CONTRADICTIONS

  SEV-4-A · SetTenantContextFromCtx Vs. WithTenantFromCtx — Naming Implies Parity,
  Behavior Differs

  The Store interface marks WithTenantFromCtx as the "preferred" API and
  SetTenantContextFromCtx as... also preferred (line 21 comment: "context-based
  (preferred)"). These are not equivalent:

  ┌──────────────────────────┬─────────────────┬────────────────┬───────────────────┐
  │          Method          │    Creates      │    RLS safe    │    Correct for    │
  │                          │   transaction   │                │                   │
  ├──────────────────────────┼─────────────────┼────────────────┼───────────────────┤
  │ SetTenantContextFromCtx  │ No              │ ❌ Pool timing │ Incorrect pattern │
  ├──────────────────────────┼─────────────────┼────────────────┼───────────────────┤
  │ WithTenantFromCtx        │ Yes             │ ✅ Same        │ Service-layer     │
  │                          │                 │ connection     │ queries           │
  ├──────────────────────────┼─────────────────┼────────────────┼───────────────────┤
  │ BeginTxWithTenantFromCtx │ Yes (manual)    │ ✅ Same        │ Multi-step        │
  │                          │                 │ connection     │ service ops       │
  └──────────────────────────┴─────────────────┴────────────────┴───────────────────┘

  SetTenantContextFromCtx exists in the interface but has no safe use case for query
  execution. Its only valid use is pre-setting the session variable before a long-lived
  connection — not in a pooled middleware context.

  The middleware must be changed: instead of SetTenantContextFromCtx, the tenant context
   should be stored in the Go context (which TenantMiddleware already does via
  shared.WithTenantID), and service code must wrap all queries in WithTenantFromCtx.

  SEV-4-B · Finance Handler extractUserID vs. Zero-Auth System

  extractUserID(c) reads c.Locals("user_id").(string). The finance handlers pass this to
   SubmitBudget, ApproveBudget, RejectBudget, ChangePeriodStatus. If auth middleware is
  absent (SEV-1-A), user_id is never set, the assertion fails silently (returns "",
  parsed to uuid.Nil), and every workflow action is attributed to uuid.Nil. The service
  layer must refuse uuid.Nil as a byUserID argument.

  SEV-4-C · Permission Comments vs. Zero Enforcement

  Finance handler permission annotations (// Permission: finance.budgets.create) are
  documentation with no enforcement behind them. FinanceHandler has no authz field.
  Every permission check is a comment. This was already recorded in tasks.md (T5) but
  the contradiction with SEV-1-A is that implementing T5 requires SEV-1-A to be resolved
   first — you cannot check authorization if the user has never been authenticated.

  ---
  SEV-5 — INFRASTRUCTURE

  SEV-5-A · In-Process Rate Limiting (Not Distributed)

  File: route_security.go:208
  router.Use(CreateRateLimitMiddleware(&rateLimitConfig, nil, m.logger))
  // nil = no cache backend → in-process only

  In a multi-instance deployment, each instance maintains its own rate limit counter. A
  client can distribute requests across instances to bypass the rate limit entirely.

  SEV-5-B · clearTenantContext in Close() Targets Random Pool Connection

  File: store.go:247
  if err := s.clearTenantContext(ctx); err != nil { ... }
  func (s *SQLStore) clearTenantContext(ctx context.Context) error {
      _, err := s.connPool.Exec(ctx, "SELECT clear_tenant_context()")

  clearTenantContext on pool close calls clear_tenant_context() on one arbitrary
  connection from the pool. In a graceful shutdown with in-flight requests, it may clear
   a different connection's tenant context, not the ones actually in use. This is low
  severity (shutdown path only) but indicates the pool-vs-connection confusion is
  systemic.

  SEV-5-C · generateRequestID Uses time.Now().UnixNano() — Not UUID

  File: observability.go:418
  func generateRequestID() string {
      return strconv.FormatInt(time.Now().UnixNano(), 36)
  }

  Nanosecond timestamps are not globally unique under concurrent requests on the same
  machine. Two requests landing at the same nanosecond get the same request ID. Use
  uuid.New().String() or the existing requestid.New() middleware (already applied by
  ConfigureAPIRoutes).

  ---
  ARCHITECTURAL VERDICT: TENANT ISOLATION CHAIN STATUS

  Client Request
        │
        ▼
    RequestID ✓            (applied in ConfigureAPIRoutes)
    Recover ✓              (applied)
    Observability ✓*        (*but panics on type assertion — SEV-1-D)
    Helmet / CORS ✓
    Rate Limit ✓*           (*in-process only — SEV-5-A)
        │
        ▼
    JWT Auth ✗             (commented out — SEV-1-A)
    TenantMiddleware ✗     (commented out in ConfigureAPIRoutes — SEV-1-A)
        │                   (implemented in tenant.go but never wired)
        ▼
    Finance Handler
      user_id = uuid.Nil   (never set — SEV-1-B)
      tenant context set in Go context ✓
      RLS via SetTenantContextFromCtx ✗  (pool timing bug — SEV-1-C)
      Authorization = comments only ✗   (SEV-4-C / T5)

  The DB transaction layer (WithTenantFromCtx, BeginTxWithTenantFromCtx) is
  architecturally correct — RLS will be enforced if service code always wraps queries in
   those methods. The bug is in the middleware calling SetTenantContextFromCtx instead,
  and in the auth layer being entirely absent.

  ---
  Remediation Priority

  ┌─────────┬───────────────┬───────────────────────────────────────┬───────────────┐
  │    #    │    Finding    │                Action                 │    Blocks     │
  ├─────────┼───────────────┼───────────────────────────────────────┼───────────────┤
  │         │ Auth          │ Implement JWT auth, wire into         │               │
  │ SEV-1-A │ middleware is │ ConfigureAPIRoutes                    │ Everything    │
  │         │  stubs        │                                       │               │
  ├─────────┼───────────────┼───────────────────────────────────────┼───────────────┤
  │ SEV-1-B │ user_id never │ Auth middleware must inject user_id   │ Audit trail   │
  │         │  set          │ into Locals                           │               │
  ├─────────┼───────────────┼───────────────────────────────────────┼───────────────┤
  │         │               │ Remove SetTenantContextFromCtx from   │               │
  │ SEV-1-C │ RLS pool      │ middleware; wire TenantMiddleware     │ Data          │
  │         │ timing bug    │ correctly; require service code to    │ isolation     │
  │         │               │ use WithTenantFromCtx                 │               │
  ├─────────┼───────────────┼───────────────────────────────────────┼───────────────┤
  │         │ Type          │ tenantID.(string) → fmt.Sprintf("%v", │ Every authed  │
  │ SEV-1-D │ assertion     │  tenantID) or store as string         │ request       │
  │         │ panic         │                                       │               │
  ├─────────┼───────────────┼───────────────────────────────────────┼───────────────┤
  │ SEV-2-A │ Tenant from   │ Remove c.Query("tenant_id") path in   │ Tenant        │
  │         │ query param   │ extractTenantID                       │ spoofing      │
  ├─────────┼───────────────┼───────────────────────────────────────┼───────────────┤
  │         │ No            │ Extract tenant from JWT claims, not   │ Cross-tenant  │
  │ SEV-2-B │ tenant-user   │ from X-Tenant-ID                      │ access        │
  │         │ binding       │                                       │               │
  ├─────────┼───────────────┼───────────────────────────────────────┼───────────────┤
  │ SEV-3-B │ GetUserIDPtr  │ Return nil when uuid.Nil, matching    │ Silent data   │
  │         │ nil bug       │ GetEntityIDPtr pattern                │ corruption    │
  ├─────────┼───────────────┼───────────────────────────────────────┼───────────────┤
  │ SEV-2-C │ Suspended     │ Add distributed cache invalidation    │ Suspension    │
  │         │ tenant cache  │ hook on tenant status change          │ enforcement   │
  └─────────┴───────────────┴───────────────────────────────────────┴───────────────┘

  ---
  Update to tasks.md

  Add to Phase 2 (Security) in internal/api/tasks.md:

  - T5 (existing) depends on SEV-1-A being resolved first
  - New tasks needed:
    - Implement JWT auth middleware and wire it into ConfigureAPIRoutes
    - Remove tenant_id query param from extractTenantID
    - Fix observability type assertions (SEV-1-D)
    - Fix GetUserIDPtr nil return (SEV-3-B)
    - Audit all service code to confirm WithTenantFromCtx wrapping
