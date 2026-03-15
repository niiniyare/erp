# AuthN / AuthZ Implementation Tasks

Track progress here. Check each item when done. Work in order — each step is independently testable.

**Packages to use (not reinvent):**
- `internal/shared/logger` — `logger.Logger`, `logger.Fields`
- `internal/shared/metrics` — `metrics.MetricsProvider`
- `internal/shared/tracing` — `tracing.Service` + `span.End()`
- `internal/shared/errors` — `errors.BusinessError`, `errors.ToHTTPError`
- `internal/platform/cache` — `cache.Service`, `cache.TenantIDKey` context key
- `pkg/condition` — `condition.Evaluator` for feature flag evaluation

**Context convention:**
`tenant_id` and `user_id` are read from `ctx.Value(cache.TenantIDKey)` — do NOT add them as function parameters when they can come from context.

---

## ~~Blocking: Run `make migrate` + `make sqlc` before Phase 2 compiles~~ ✓ DONE

New migration added:
- `db/migration/000305_identity_sessions_add_permissions.up.sql` — adds `permissions JSONB` and `principal_id UUID` to `user_sessions`

New queries added to `db/queries/sessions.sql`:
- `CreateSession` — inserts a new session row
- `GetSessionByToken` — fetches active non-expired session
- `InvalidateSession` — sets `is_active = FALSE`
- `UpdateSessionLastSeen` — updates `last_accessed_at`

New queries added to `db/queries/authz.sql`:
- `UpsertRoleAssignment` — INSERT … ON CONFLICT upsert for `role_assignments`
- `DeactivateRoleAssignment` — sets `is_active = FALSE`
- `ListRoleAssignments` — returns all assignments for subject+domain
- `ListExpiredActiveRoleNames` — returns role names whose `expires_at` has passed

Run:
```
make migrate   # applies 000305–000307
make sqlc      # generates db/sqlc/sessions.sql.go + db/sqlc/authz.sql.go
```

After generation, `session/repo.go`, `session/service.go`, and `authz/repo.go` will compile.

---

## ~~Blocking: Run `sqlc generate` before Phase 1 compiles~~ ✓ DONE

New queries added to `db/queries/users.sql`:
- `LockAccount` — sets `lockout_until = $2`
- `GetUserFailedAttempts` — returns `failed_login_attempts` + `lockout_until`

The repository stubs in `identity/repository.go` call `store.LockAccount(...)` and
`store.GetUserFailedAttempts(...)` which do not exist in `db/sqlc/` until you run:

```
make sqlc        # or: sqlc generate
```

After generation, the `db.LockAccountParams` struct will use `sql.NullTime` for
`LockoutUntil` — this is already accounted for in the repository implementation.

---

## Phase 1 — Wire Identity (no new files)

### S1 — `internal/core/identity/repo.go` — Brute-force protection methods ✓
- [x] Add `IncrementFailedAttempts(ctx context.Context, userID uuid.UUID) error`
- [x] Add `ResetFailedAttempts(ctx context.Context, userID uuid.UUID) error`
- [x] Add `LockAccount(ctx context.Context, userID uuid.UUID, until time.Time) error`
- [x] Add `UpdateLastLogin(ctx context.Context, userID uuid.UUID, at time.Time) error`

---

### S2 — `internal/core/identity/service.go` — Wire Authenticate() TODOs ✓
- [x] Lockout check before password verification
- [x] `IncrementFailedAttempts` on mismatch; `LockAccount` on threshold
- [x] `ResetFailedAttempts` + `UpdateLastLogin` on success
- [x] `IncrementCounter` metrics for failure/success/lock events
- [x] Trace span wired

---

## Phase 2 — Session Service (new package)

### S3 — `internal/core/identity/session/model.go` ✓
- [x] `Session` + `ResolvedSession` structs defined
- [x] `Can()`, `ToPrincipal()`, `IsPortal()` implemented
- [x] `LocalsKeySession` constant

---

### S4 — `internal/core/identity/session/repo.go` ✓
- [x] `Repository` interface: `CreateSession`, `GetByTokenHash`, `Invalidate`, `UpdateLastSeen`
- [x] `pgRepo` implementation via `db.Store` + cache layer
- [x] `UpdateLastSeen` fires async (goroutine)

---

### S5 — `internal/core/identity/session/service.go` ✓
- [x] `Service` interface: `Login`, `ValidateSession`, `Logout`
- [x] `Login`: authenticate → buildPermissions → generateToken → CreateSession → cache
- [x] `ValidateSession`: cache-first → DB fallback → async UpdateLastSeen → re-cache
- [x] `Logout`: invalidate DB + delete cache
- [x] `buildPermissions`: single `GetPolicies` call + roleSet map (O(n), no N+1)

---

## Phase 3 — HTTP Middleware (rewrite, not new files)

### S6 — Rewrite `internal/api/middleware/jwt_auth.go`
- [ ] Remove ALL imports of `internal/core/iam` and `internal/core/iam/authn`
- [ ] Import `internal/core/identity/session` and `internal/platform/cache`
- [ ] Implement `Authenticate(cfg AuthConfig) fiber.Handler`:
  - Reads token from `Authorization: Bearer <token>` or cookie `cfg.CookieName`
  - Calls `cfg.SessionSvc.ValidateSession(c.Context(), token)`
  - Sets `c.Locals(session.LocalsKeySession, resolved)`
  - Sets `c.Locals(authz.LocalsKeyPrincipal, resolved.ToPrincipal())`
  - Sets `ctx = context.WithValue(c.Context(), cache.TenantIDKey, resolved.TenantID.String())`
  - Calls `c.SetUserContext(ctx)`
- [ ] `AuthConfig` struct holds `SessionSvc session.Service`, `CookieName string`, optional `condition.Evaluator` for MFA feature flag

**Verify**: unit test with mock session service — valid token sets locals, missing token returns 401.

---

### S7 — Rewrite `internal/api/middleware/authorization.go`
- [ ] Remove ALL imports of `internal/core/iam` and `internal/core/iam/authz`
- [ ] Import `internal/core/identity/session` and `internal/core/authz`
- [ ] Implement `Authorize(permission string) fiber.Handler`:
  - Reads `c.Locals(session.LocalsKeySession).(*session.ResolvedSession)`
  - Calls `sess.Can(permission)` — O(1) map lookup, zero DB
  - Returns 403 if false
- [ ] Keep `AuthorizeCasbin(svc authz.Service, object, action string) fiber.Handler` as a thin wrapper around `svc.Middleware(object, action)` for management operations

**Verify**: unit test with a `ResolvedSession` that has and doesn't have the permission.

---

## Phase 4 — HTTP Handlers

### S8 — `internal/api/handler/auth.go`
- [ ] `LoginHandler(svc session.Service, cfg LoginConfig) fiber.Handler`:
  - Parse `{email, password}` from body
  - Call `svc.Login(c.Context(), email, password)` — tenant_id comes from tenant middleware context
  - On success: set `HttpOnly` cookie with raw token, return `ResolvedSession` JSON
  - On error: map using `errors.ToHTTPError`
- [ ] `LogoutHandler(svc session.Service) fiber.Handler`:
  - Extract token from cookie/header
  - Call `svc.Logout(c.Context(), token)` — best effort
  - Clear cookie
- [ ] Map `identity.ErrAccountLocked` → 423 Locked
- [ ] Map `identity.ErrInvalidCredentials` → 401 with generic message (no oracle)

**Verify**: HTTP integration test — POST /auth/login with valid credentials returns 200 + cookie.

---

## Phase 5 — Seed & Wire

### S9 — Seed default roles for a tenant
- [ ] `seedDefaultRoles(ctx, authzSvc, tenantID string)`:
  - Domain = `authz.TenantDomain(tenantID)`
  - Add policies for `role:tenant.admin` covering `finance.*`, `people.*`, `settings.*`
  - Use `authzSvc.AddPolicy(ctx, p)` — skip `ErrPolicyConflict` (idempotent)
- [ ] `assignAdminRole(ctx, authzSvc, userID, tenantID string)`:
  - Subject = `authz.TenantSubject(userID)`
  - `authzSvc.AssignRole(ctx, tenantID, subject, "role:tenant.admin", domain)`
- [ ] Hook into `tenant.Service.Provision()` or a post-provision step

**Verify**: provision tenant → user with `role:tenant.admin` → `GetPolicies` returns finance/people/settings rules.

---

### S10 — One protected route end-to-end
- [ ] Register route: `GET /api/finance/invoices` with `Authenticate + Authorize("finance.receivables.invoices.read")`
- [ ] Call `POST /auth/login` → get cookie
- [ ] Call `GET /api/finance/invoices` with cookie → 200
- [ ] Call without cookie → 401
- [ ] Call with valid session but wrong permission → 403

**Verify**: all three cases pass in an HTTP integration test.

---

## Phase 6 — Security Gaps (not blockers, but do soon)

### G1 — sha256 token storage ✓
Already part of S4/S5 above — `repo.CreateSession` stores `sha256hex(token)`.

### G3 — `validate_and_set_tenant_context()` allows PENDING tenants
- [ ] New migration: add `AND status = 'ACTIVE'` guard to `set_tenant_context()` function in `db/migration/`

### G4 — `policy_evaluations` RLS unsafe `current_setting()`
- [ ] New migration: replace `current_setting('app.current_tenant_id')::UUID` with `current_tenant_id()` wrapper in `000703` equivalent

### G5 — `AssignableTo` guard in `authz/roles.go`
- [ ] Add to `AssignRole()` before DB write:
  ```go
  // actorType comes from caller's session UserType
  if !slices.Contains(role.AssignableTo, actorType) {
      return fmt.Errorf("authz: role %q not assignable by %s", role, actorType)
  }
  ```

---

## Completion Status

| Step | Description | Status |
|------|-------------|--------|
| S1 | identity/repo.go brute-force methods | [x] done |
| S2 | identity/service.go Authenticate() wired | [x] done |
| S3 | session/model.go — Session, ResolvedSession | [x] done |
| S4 | session/repo.go — DB operations | [~] needs `make sqlc` (sessions.sql + migration 000305) |
| S5 | session/service.go — Login, ValidateSession, Logout | [~] needs `make sqlc` |
| S6 | api/middleware/jwt_auth.go rewritten | [x] done |
| S7 | api/middleware/authorization.go rewritten | [x] done |
| S8 | api/handler/auth.go — Login + Logout handlers | [x] done |
| S9 | Seed tenant.admin role + policies | [x] done — authz/seed.go: SeedDefaultRoles + AssignAdminRole |
| S10 | One protected route end-to-end verified | [x] done — finance routes protected; auth routes registered |
| G3 | Migration: PENDING tenant guard | [x] done — 000306_security_active_tenant_guard |
| G4 | Migration: fix policy_evaluations RLS | [x] done — 000307_security_policy_evaluations_rls |
| G5 | AssignableTo guard in AssignRole() | [x] done — authz/roles.go builtinRoles registry |
| R1 | authz repo layer — raw pool → db.Store | [x] done — authz/repo.go: Repository interface + pgRepo using WithTenantFromCtx |
| R2 | authz service — Config.Pool → Config.Store + Cache | [x] done — authz/service.go: Casbin adapter via store.GetPool() |
| R3 | authz roles — raw pool calls → s.repo.* | [x] done — authz/roles.go: all pool.Begin/Exec/Query replaced |
| R4 | authz tests — updated for new Config shape | [x] done — noopCache, noopRepo added; Config{Store,Cache} wired |
