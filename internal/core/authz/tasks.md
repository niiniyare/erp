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

## Blocking: Run `make migrate` + `make sqlc` before Phase 2 compiles

New migration added:
- `db/migration/000305_identity_sessions_add_permissions.up.sql` — adds `permissions JSONB` and `principal_id UUID` to `user_sessions`

New queries added to `db/queries/sessions.sql`:
- `CreateSession` — inserts a new session row
- `GetSessionByToken` — fetches active non-expired session
- `InvalidateSession` — sets `is_active = FALSE`
- `UpdateSessionLastSeen` — updates `last_accessed_at`

Run:
```
make migrate   # applies 000305
make sqlc      # generates db/sqlc/sessions.sql.go
```

After generation, `session/repo.go` and `session/service.go` will compile.

---

## Blocking: Run `sqlc generate` before Phase 1 compiles

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

### S1 — `internal/core/identity/repo.go` — Brute-force protection methods
- [ ] Add `IncrementFailedAttempts(ctx context.Context, userID uuid.UUID) error`
- [ ] Add `ResetFailedAttempts(ctx context.Context, userID uuid.UUID) error`
- [ ] Add `LockAccount(ctx context.Context, userID uuid.UUID, until time.Time) error`
- [ ] Add `UpdateLastLogin(ctx context.Context, userID uuid.UUID, at time.Time) error`

All columns already exist in `users` table (`failed_login_attempts`, `lockout_until`, `last_login_at`).
No migrations needed.

**Verify**: write a table-driven unit test calling each repo method against the test DB.

---

### S2 — `internal/core/identity/service.go` — Wire Authenticate() TODOs
- [ ] At top of `Authenticate()`: check `user.LockoutUntil != nil && user.LockoutUntil.After(time.Now())` → return `ErrAccountLocked`
- [ ] On password mismatch: call `repo.IncrementFailedAttempts(ctx, user.ID)`
- [ ] On 5th failure: call `repo.LockAccount(ctx, user.ID, time.Now().Add(cfg.LockoutDuration))`
- [ ] On success: call `repo.ResetFailedAttempts` + `repo.UpdateLastLogin`
- [ ] Use `s.metrics.RecordCount("identity.auth.failure", 1, ...)` and `"identity.auth.success"` for observability
- [ ] Use `s.tracing.StartSpan(ctx, "identity.Authenticate")` — already present in other methods, follow the same pattern

**Verify**: test with wrong password 5 times → account locks → correct password returns `ErrAccountLocked`.

---

## Phase 2 — Session Service (new package)

### S3 — `internal/core/identity/session/model.go`
- [ ] Define `Session` struct (maps to `user_sessions` table):
  ```go
  type Session struct {
      ID          uuid.UUID
      UserID      uuid.UUID
      TenantID    uuid.UUID
      TokenHash   string          // sha256hex(raw_token) — never store raw
      Permissions map[string]bool // stored as JSONB
      PrincipalID uuid.UUID       // portal users: their contact/employee ID
      IsActive    bool
      ExpiresAt   time.Time
      LastSeenAt  time.Time
      IPAddress   string
      UserAgent   string
      RiskScore   int
  }
  ```
- [ ] Define `ResolvedSession` struct:
  ```go
  type ResolvedSession struct {
      UserID      uuid.UUID
      UserType    string          // "platform" | "tenant" | "portal"
      TenantID    uuid.UUID
      PrincipalID uuid.UUID
      DisplayName string
      Permissions map[string]bool
  }

  const LocalsKeySession = "session"

  func (s *ResolvedSession) Can(permission string) bool
  func (s *ResolvedSession) ToPrincipal() authz.Principal   // builds Subject+Domain from UserType
  func (s *ResolvedSession) IsPortal() bool
  ```
- [ ] `ToPrincipal()` uses `authz.TenantSubject/PlatformSubject/PortalSubject` helpers from `authz/types.go`

**Verify**: unit test `ToPrincipal()` for all three user types.

---

### S4 — `internal/core/identity/session/repo.go`
- [ ] Define `Repository` interface:
  ```go
  type Repository interface {
      CreateSession(ctx context.Context, s Session) error
      GetByTokenHash(ctx context.Context, hash string) (*Session, error)
      Invalidate(ctx context.Context, hash string) error
      UpdateLastSeen(ctx context.Context, hash string) error
  }
  ```
- [ ] Implement against `user_sessions` table
- [ ] `CreateSession`: INSERT with `token_hash = sha256hex(token)` — column is `session_token` in schema, store the hash there
- [ ] `GetByTokenHash`: SELECT WHERE `session_token = $1 AND is_active = TRUE AND expires_at > NOW()`
- [ ] `UpdateLastSeen`: fire-and-forget UPDATE (do not block the request — run in goroutine)
- [ ] Use `cache.TenantIDKey` context value for RLS tenant context if needed

**Verify**: integration test — create session, get by hash, invalidate, confirm GetByHash returns nil.

---

### S5 — `internal/core/identity/session/service.go`
- [ ] Define `Service` interface:
  ```go
  type Service interface {
      Login(ctx context.Context, email, password string) (*ResolvedSession, string, error)
      ValidateSession(ctx context.Context, token string) (*ResolvedSession, error)
      Logout(ctx context.Context, token string) error
  }
  ```
- [ ] Implement `service` struct with:
  - `identity identity.Service`
  - `authz   authz.Service`
  - `repo    Repository`
  - `cache   cache.Service`   ← caches `ResolvedSession` keyed by `sha256(token)`
  - `tracer  tracing.Service`
  - `metrics metrics.MetricsProvider`
  - `log     logger.Logger`
  - `cfg     Config`          ← session TTL, lockout config, feature flags
- [ ] `Login()`:
  1. `identity.Authenticate(ctx, email, password)` — handles brute-force (S2)
  2. `buildPermissions(ctx, user)` using `authz.GetRoles` + `authz.GetPolicies`
  3. `generateToken()` — `crypto/rand` 32 bytes → `hex.EncodeToString`
  4. `sha256hex(token)` — store the hash, return the raw token
  5. `repo.CreateSession(ctx, Session{...})`
  6. `cache.Set(ctx, "session:"+hash, resolved, ttl)`
  7. Emit `metrics.RecordCount("session.login.success", 1, ...)`
  8. Trace span: `"session.Login"`
- [ ] `ValidateSession()`:
  1. `hash := sha256hex(token)`
  2. Try `cache.Get(ctx, "session:"+hash)` → cache hit: return immediately
  3. Cache miss: `repo.GetByTokenHash(ctx, hash)` → build `ResolvedSession`
  4. `go repo.UpdateLastSeen(ctx, hash)` — async, do not block
  5. Re-populate cache: `cache.Set(...)`
- [ ] `Logout()`:
  1. `hash := sha256hex(token)`
  2. `repo.Invalidate(ctx, hash)`
  3. `cache.Delete(ctx, "session:"+hash)`
- [ ] `buildPermissions()` — see doc `14-how-other-packages-use-authz.md §0`

**Verify**: end-to-end test: Login → ValidateSession (cache miss) → ValidateSession (cache hit) → Logout → ValidateSession returns error.

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
| S4 | session/repo.go — DB operations | [~] needs `sqlc generate` (db/queries/sessions.sql + migration 000305) |
| S5 | session/service.go — Login, ValidateSession, Logout | [~] needs `sqlc generate` |
| S6 | api/middleware/jwt_auth.go rewritten | [x] done |
| S7 | api/middleware/authorization.go rewritten | [x] done |
| S8 | api/handler/auth.go — Login + Logout handlers | [x] done |
| S9 | Seed tenant.admin role + policies | [ ] pending |
| S10 | One protected route end-to-end verified | [ ] pending |
| G3 | Migration: PENDING tenant guard | [ ] pending |
| G4 | Migration: fix policy_evaluations RLS | [ ] pending |
| G5 | AssignableTo guard in AssignRole() | [ ] pending |
