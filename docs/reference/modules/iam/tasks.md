# IAM Module — Implementation Task List

> **How to use this file**
> Work top-to-bottom. Each phase depends on the one before it.
> Check off `[x]` when a task is done. Do not skip phases.
> Every task has an explanation written so a new developer can understand what it does and why before touching code.

---

## Reading before you start

| File | Why you should read it |
|---|---|
| `docs/reference/modules/iam/00-iam-overview.md` | The three questions every request answers (AuthN / AuthZ / Config) |
| `docs/reference/modules/iam/00b-code-architecture.md` | Repository layout and conventions |
| `docs/reference/modules/iam/06b-authentication.md` | Full login flow step-by-step |
| `docs/reference/modules/iam/10b-session-precomputation.md` | Why everything is computed at login |
| `internal/core/iam/domain/authz.go` | Actor types, Casbin model, policy primitives |
| `internal/core/iam/domain/session.go` | ResolvedSession — the object every handler reads |

**Shared packages — use these, never reinvent them:**
- `internal/shared/logger` — structured logging (`logger.Logger`, `logger.Fields`)
- `internal/shared/metrics` — metrics counters and histograms (`metrics.MetricsProvider`)
- `internal/shared/tracing` — distributed tracing (`tracing.Service`, `span.End()`)
- `internal/shared/errors` — HTTP-safe errors (`errors.BusinessError`, `errors.ToHTTPError`)
- `internal/platform/cache` — Redis cache wrapper (`cache.Service`, `cache.TenantIDKey`)
- `pkg/condition` — rule evaluation for feature flags (`condition.Evaluator`)

**Context convention:**
`tenant_id` and `user_id` travel in `ctx.Value(cache.TenantIDKey)` — do NOT add them as function parameters when the context already carries them.

---

## ✅ Phase 0 — Foundation (Already Done)

These were completed in the earlier `identity` + `authz` packages and migrated into `internal/core/iam/`.

### ✅ DB Migrations

> **What this is:** Database schema changes are tracked as numbered migration files. Before any Go code that touches a new table can compile, the migration must be applied and `sqlc` must regenerate the Go query functions.

- [x] `000305_identity_sessions_add_permissions.up.sql` — adds `permissions JSONB` and `principal_id UUID` to `user_sessions`
- [x] `000306_security_active_tenant_guard.up.sql` — blocks PENDING tenants from setting tenant context
- [x] `000307_security_policy_evaluations_rls.up.sql` — fixes RLS on `policy_evaluations` table

### ✅ SQLC Query Generation

> **What this is:** `sqlc` reads raw SQL files in `db/queries/` and generates type-safe Go functions. Every time you add a new SQL query, you run `make sqlc` to get the Go wrapper.

- [x] `db/queries/sessions.sql` — `CreateSession`, `GetSessionByToken`, `InvalidateSession`, `UpdateSessionLastSeen`
- [x] `db/queries/authz.sql` — `UpsertRoleAssignment`, `DeactivateRoleAssignment`, `ListRoleAssignments`, `ListExpiredActiveRoleNames`
- [x] `db/queries/users.sql` — `LockAccount`, `GetUserFailedAttempts`

---

## ✅ Phase 1 — Domain Layer

> **What this is:** The domain layer holds pure business logic. No database, no HTTP, no frameworks — just data types and rules. It is the vocabulary shared by every other layer.

### ✅ D1 — `internal/core/iam/domain/authz.go`

> Defines the authorization primitives. Every concept in Casbin (who, where, what, how) maps to a Go type here.

- [x] `ActorType` typed string enum (`ActorPlatform`, `ActorTenant`, `ActorPortal`, `ActorAPI`)
- [x] `ActorTypeFromUserType()` — single canonical mapping from DB user_type string to ActorType
- [x] Subject helpers: `PlatformSubject()`, `TenantSubject()`, `PortalSubject()`, `APISubject()`
- [x] Domain helpers: `TenantDomain()`, `PortalDomain()`, `APIDomain()`
- [x] `DomainPlatform` constant (`"_platform_"`)
- [x] `Principal` value object (subject + domain pair used by Casbin middleware)
- [x] `Request` struct (sub, dom, obj, act — mirrors Casbin tuple)
- [x] `Policy` struct with `Effect` field (`"allow"` or `"deny"`)
- [x] `RoleAssignment` entity with `ExpiresAt`, `IsExpired()`, `IsEffective()`
- [x] `AssignOpt` functional options (`WithExpiry`, `WithAssignedBy`, `WithDelegatedBy`)
- [x] `CasbinModel` const — the CONF model string with deny-override effect
  - **Note:** uses `keyMatch` (not `keyMatch2`) for actions — this is intentional. `keyMatch2` would break wildcard `*` actions.

### ✅ D2 — `internal/core/iam/domain/session.go`

> The `ResolvedSession` is the most important type in the whole IAM module. Every handler reads it. It carries pre-computed permissions, flags, settings, and entity scope — so handlers never need to hit the database for auth.

- [x] `SessionConfig` — process-level defaults (TTL, cookie name)
- [x] `DefaultSessionConfig()` — 8h TTL default
- [x] `EntityScopeType` (`all` / `subtree` / `entity`)
- [x] `EntityScope` struct with `PathPrefix` for ltree queries
- [x] `Configuration` struct — `Flags map[string]bool`, `Settings map[string]string`, `Prefs map[string]string`
- [x] `Session` aggregate root — maps to `user_sessions` DB row
- [x] `ResolvedSession` — lightweight request-scoped view used by handlers
- [x] `Can(permission string) bool` — O(1) map lookup, single string key (e.g. `"finance.transactions.read"`)
- [x] `CanDo(resource, action string) bool` — convenience wrapper over `Can`
- [x] `FeatureEnabled(flag string) bool` — O(1) flag lookup
- [x] `SettingString/Bool/Int/Decimal` — typed setting accessors
- [x] `ToPrincipal()` — converts session into `Principal` for Casbin
- [x] `IsPlatform()`, `IsPortal()` predicates
- [x] `LocalsKeySession`, `LocalsKeyPrincipal` — Fiber context keys

### ✅ D3 — `internal/core/iam/domain/identity.go`

> User is the credential bundle. Person is the human. Employee is the HR record. They are separate because not every user is a person (API accounts) and not every person has a user account.

- [x] `AccountStatus` enum (`ACTIVE`, `INACTIVE`, `LOCKED`, `SUSPENDED`)
- [x] `EmploymentStatus` enum (`ACTIVE`, `INACTIVE`, `TERMINATED`, `ON_LEAVE`, `SUSPENDED`)
- [x] `User` aggregate — credentials, account state, lockout, MFA flag
- [x] `User.IsLocked()` — checks lockout timer without DB
- [x] `User.CanAuthenticate()` — combined liveness check
- [x] `Person` entity — PII (name, national ID, KRA PIN, date of birth)
- [x] `Employee` entity — HR record linked to a Person
- [x] `UserWithDetails` composite — User + optional Person + optional Employee
- [x] `UserRole` — audit record of a role assignment
- [x] `CreateUserRequest` with `Validate()` domain check
- [x] `UpdateUserRequest` — all-pointer PATCH semantics
- [x] `ListUsersRequest` — filters + pagination
- [x] `AuthenticateRequest` — email or username + password
- [x] `ChangePasswordRequest`
- [x] `CreatePersonRequest`, `CreateEmployeeRequest`

### ✅ D4 — `internal/core/iam/domain/errors.go`

> Domain errors stay inside the domain package — they never import HTTP frameworks. They carry an HTTP status code so the handler layer can map them without knowing the original error.

- [x] `Error` struct (`Code`, `Message`, `HTTPStatus`)
- [x] `ErrForbidden` (403), `ErrUnauthorized` (401), `ErrInvalidRequest` (400), `ErrPolicyConflict` (409)
- [x] `ErrInvalidIdentity(msg)` factory

---

## ✅ Phase 2 — Repository Layer

> **What this is:** Repositories are the only layer that talks to the database. They translate between domain types and SQL. The service layer calls repository methods — it never writes SQL directly.

### ✅ R1 — `internal/core/iam/repository/authz.go`

> This file contains two things: the `AuthzRepository` (for our metadata tables like `role_assignments`) and the `pgxAdapter` (the bridge that lets Casbin talk to PostgreSQL).

- [x] `AuthzRepository` interface: `UpsertRoleAssignment`, `DeactivateRoleAssignment`, `ListRoleAssignments`, `ListExpiredActiveRoleNames`
- [x] `authzRepo` implementation using `db.Store` (never raw pool)
- [x] `NewAuthzRepository()` constructor
- [x] `pgxAdapter` implementing Casbin `persist.BatchAdapter`
- [x] `LoadPolicy()` — reads all Casbin rules from `casbin_rule` into Casbin model
- [x] `SavePolicy()` — replaces all rules (used for bulk operations)
- [x] `AddPolicy()`, `AddPolicies()` — insert new rules
- [x] `RemovePolicy()`, `RemovePolicies()` — delete rules
- [x] `RemoveFilteredPolicy()` — delete rules matching a filter
- [x] `NewPgxAdapter()` constructor

### ✅ R2 — UserRepository and SessionRepository

> These repos handle the `users` and `user_sessions` tables respectively.

- [x] `UserRepository` interface + implementation with cache layer (5-min TTL)
- [x] `SessionRepository` interface + implementation with cache layer
- [x] `NewUserRepository()` constructor
- [x] `NewSessionRepository()` constructor

---

## ✅ Phase 3 — Service Layer (AuthzService)

> **What this is:** The service layer contains business logic. It orchestrates repositories, enforces business rules, and is the public API that handlers call.

### ✅ S1 — `internal/core/iam/service/authz.go`

> The AuthzService is the Go interface to Casbin. It manages roles, policies, and enforcement. Think of Casbin as the engine and AuthzService as the steering wheel.

- [x] `AuthzService` interface defined
- [x] `authzService` struct with Casbin enforcer + repository + cache + logger
- [x] `NewAuthzService(cfg Config)` — boots Casbin with PostgreSQL adapter
- [x] `NewInMemoryAuthzService()` — in-memory enforcer for unit tests
- [x] `Enforce(ctx, req Request) (bool, error)` — single authorization check
- [x] `EnforceBatch(ctx, reqs []Request) ([]bool, error)` — batch check
- [x] `AssignRole(ctx, tenantID, subject, role, domain string, opts ...AssignOpt) error`
- [x] `RevokeRole(ctx, subject, role, domain string) error`
- [x] `GetRoles(ctx, subject, domain string) ([]string, error)`
- [x] `GetImplicitRoles(ctx, subject, domain string) ([]string, error)` — includes inherited roles
- [x] `HasRole(ctx, subject, role, domain string) (bool, error)`
- [x] `GetAssignments(ctx, subject, domain string) ([]*RoleAssignment, error)`
- [x] `AddPolicy(ctx, p Policy) error`
- [x] `RemovePolicy(ctx, p Policy) error`
- [x] `GetPolicies(ctx, subject, domain string) ([]Policy, error)`
- [x] `InvalidateCache(ctx context.Context) error`
- [x] `revokeExpiredRoles()` — internal: lazy cleanup of expired assignments

### ✅ S2 — UserService and SessionService

> UserService owns user creation and credential management. SessionService owns login, session validation, and logout.

- [x] `UserService` interface + implementation (brute-force protection, lockout)
- [x] `IncrementFailedAttempts`, `ResetFailedAttempts`, `LockAccount`, `UpdateLastLogin` in repo
- [x] Escalating lockout logic (≥5 failures → lockout with backoff)
- [x] `SessionService` interface + implementation
- [x] `Login()` → authenticate → buildPermissions → generateToken → CreateSession → cache
- [x] `ValidateSession()` → cache-first → DB fallback → async UpdateLastSeen
- [x] `Logout()` → invalidate DB + delete cache key
- [x] `buildPermissions()` — single query, builds `map[string]bool` from user's roles

---

## ✅ Phase 4 — IAM Facade

> **What this is:** External code (handlers, other modules) imports only `internal/core/iam`. They never import the sub-packages directly. The facade re-exports everything they need through a single import path.

### ✅ F1 — `internal/core/iam/iam.go`

- [x] Re-exports all domain types (User, Session, ResolvedSession, Principal, Policy, etc.)
- [x] Re-exports all constants (ActorPlatform, DomainPlatform, EntityScopeAll, LocalsKeySession, etc.)
- [x] Re-exports all error sentinels
- [x] Re-exports Subject/Domain helper functions
- [x] Re-exports service interfaces (UserService, AuthzService, SessionService)
- [x] Re-exports repository interfaces
- [x] Constructor functions for wire injection (NewUserService, New, NewSessionService, etc.)

### ✅ F2 — `internal/core/iam/seed.go`

> When a new tenant is created, they need a starting set of roles and permissions. This file creates the `role:tenant.admin` with policies covering finance, people, and settings — and assigns it to the first admin user.

- [x] `SeedDefaultRoles(ctx, svc, tenantID)` — adds finance/people/settings policies for `role:tenant.admin`
- [x] `AssignAdminRole(ctx, svc, tenantID, userID)` — grants `role:tenant.admin` to a user
- [x] Idempotent: skips `ErrPolicyConflict` so safe to call multiple times

---

## ✅ Phase 5 — HTTP Layer

### ✅ H1 — Authentication Middleware

> **What this is:** Every route that requires login runs through `Authenticate` first. This middleware reads the session token from cookie or `Authorization` header, validates it, and puts the `ResolvedSession` into Fiber context so handlers can read it.

- [x] `Authenticate(cfg AuthConfig) fiber.Handler`
  - Reads token from `Authorization: Bearer <token>` or cookie
  - Calls `SessionService.ValidateSession()`
  - Sets `c.Locals(LocalsKeySession, resolved)`
  - Sets `c.Locals(LocalsKeyPrincipal, resolved.ToPrincipal())`
  - Injects `tenant_id` into context via `cache.TenantIDKey`
  - Returns 401 on missing/invalid token

### ✅ H2 — Authorization Middleware

> After authentication confirms who you are, authorization checks what you can do. This is an O(1) map lookup — it never touches the database.

- [x] `Authorize(permission string) fiber.Handler`
  - Reads `ResolvedSession` from Fiber locals
  - Calls `sess.Can(permission)` — zero DB, pure map lookup
  - Returns 403 if false
- [x] `AuthorizeCasbin(svc AuthzService, object, action string) fiber.Handler`
  - Thin wrapper around `svc.Middleware()` for management operations where live Casbin check is needed

### ✅ H3 — Auth Handlers

> The HTTP handlers that handle login and logout requests.

- [x] `LoginHandler` — POST /auth/login
  - Parses `{identifier, password}` from body (identifier = email or username)
  - Calls `SessionService.Login()`
  - Sets `HttpOnly` + `Secure` + `SameSite=Lax` cookie with raw token
  - Returns `ResolvedSession` JSON on success
  - Maps `ErrAccountLocked` → 423, `ErrInvalidCredentials` → 401 (generic message, no oracle)
- [x] `LogoutHandler` — POST /auth/logout
  - Extracts token from cookie or header
  - Calls `SessionService.Logout()` (best effort)
  - Clears cookie

---

## ✅ Phase 6 — Security Hardening

### ✅ SEC1 — Token stored as SHA-256 hash

> **Why:** If the `user_sessions` table is leaked, raw tokens cannot be replayed. The server returns the plaintext token to the client once, stores only the hash. On each request, it hashes the incoming token and looks up the hash.

- [x] `repo.CreateSession` stores `sha256hex(raw_token)` in DB
- [x] `repo.GetByTokenHash` takes `sha256hex(incoming_token)` for lookup

### ✅ SEC2 — PENDING tenant guard

> **Why:** A tenant that is still being set up (status = PENDING) should not be able to serve requests. This migration adds a DB-level check so even if the Go code forgets, Postgres rejects it.

- [x] Migration `000306` — `set_tenant_context()` function now requires `status = 'ACTIVE'`

### ✅ SEC3 — Fix `policy_evaluations` RLS

> **Why:** Using `current_setting()` directly in RLS policies is unsafe — if the setting isn't set, Postgres errors or returns wrong results. Wrapping it in a function with a safe default fixes this.

- [x] Migration `000307` — replaced `current_setting('app.current_tenant_id')::UUID` with `current_tenant_id()` wrapper function

### ✅ SEC4 — `AssignableTo` guard in `AssignRole()`

> **Why:** Without this guard, a portal user could in theory call `AssignRole` and assign themselves a tenant-admin role. The guard checks that the caller's actor type is allowed to assign the requested role.

- [x] `builtinRoles` registry in authz service defines which actor types can assign each role
- [x] `AssignRole()` checks caller's actor type against `role.AssignableTo` before writing to DB

---

## 🔲 Phase 7 — Session Construction: Wire Flags + Settings

> **Context:** Right now, when a user logs in, the `ResolvedSession.Configuration.Flags` and `.Settings` maps are empty. The docs (and code comments) say these should be populated at login time by querying the feature flag and settings modules. This phase wires those up.
>
> **Why it matters:** Every handler checks `session.FeatureEnabled("finance.transactions")` or `session.SettingDecimal("finance.approval_threshold", 0)`. If the maps are empty, everything returns the default/false — features appear disabled even when they're on.

### ✅ T1 — `SessionService.Login()` — flag resolution

> Already implemented via `SessionRepository.LoadLoginConfig()` which calls `store.ResolveAllFlagsForTenant()` and populates `Configuration.Flags` in a single DB query at login.

- [x] `LoadLoginConfig()` in session repo calls `ResolveAllFlagsForTenant` → `cfg.Flags`
- [x] Non-fatal: partial failure logs warning, continues with what's available
- [x] SQLC query `ResolveAllFlagsForTenant` exists in `db/sqlc/querier.go`

### ✅ T2 — `SessionService.Login()` — settings resolution

> Already implemented via `SessionRepository.LoadLoginConfig()` which calls `store.ResolveAllSettingsForTenant()` and `store.GetUserPreferences()`.

- [x] `LoadLoginConfig()` calls `ResolveAllSettingsForTenant` → `cfg.Settings`
- [x] `LoadLoginConfig()` calls `GetUserPreferences` → `cfg.Prefs`
- [x] SQLC queries exist in `db/sqlc/querier.go`

### ✅ T3 — `SessionService.Login()` — resolve session TTL from settings

> Now that settings are loaded at login into `Configuration.Settings`, the session TTL reads `"iam.session_ttl_hours"` from that map. Falls back to the 8h process default when the setting is absent or invalid.

- [x] `resolveTTL(cfg domain.Configuration) time.Duration` helper added to session service
- [x] Reads `cfg.Settings["iam.session_ttl_hours"]`, parses as int, converts to `time.Duration`
- [x] Falls back to `s.cfg.SessionTTL` (8h) on missing/invalid value
- [x] `Login()` uses `ttl` from `resolveTTL()` for both `ExpiresAt` and `CacheResolved()`
- [x] TODO/FIXME comments removed from `domain/session.go`

### ✅ T4 — Session invalidation when flags change

> When a platform admin toggles a feature flag for a tenant (e.g. enables the Finance module), existing sessions still have the old `false` value. Those sessions must be invalidated so users re-login and get fresh flags.

- [x] In `featureflag.UpdateFeatureFlag()`, after the update is committed and audited, calls `s.store.InvalidateSessionsByTenant(ctx, t.ID)` — best-effort (error is silently dropped, flag update is not rolled back)
- [x] `InvalidateByTenant(ctx, tenantID)` already existed in `SessionRepository` interface
- [x] SQL `UPDATE user_sessions SET is_active = FALSE WHERE tenant_id = $1 AND is_active = TRUE` already existed in `db/queries/sessions.sql`

### T5 — Verify end-to-end: flags visible in session

- [ ] Write an integration test: create tenant → enable `finance` flag → login → assert `session.FeatureEnabled("finance") == true`
- [ ] Write a test: disable flag → login again → assert `session.FeatureEnabled("finance") == false`

---

## ✅ Phase 8 — Entity Module

> **Context:** The docs describe an entity hierarchy — a tree of org nodes (Company → Region → Branch → Department). Every user belongs to an entity. At login, the user's entity position determines their `EntityScope` (do they see all data, their subtree, or just their own entity?).
>
> Right now `EntityScope` is defined in the domain but there is no actual entity table management — no CRUD service, no repository, no migrations for entities.
>
> **Why it matters:** Without a real entity tree, all users default to `EntityScopeAll` which means everyone sees all data within their tenant — there's no org-level data partitioning.

### ✅ E1 — DB migration: `entities` table

> Already existed — richer schema than described: uuid PK, entity_path, entity_level, hierarchy_paths closure table, accrual settings, soft delete, etc.

- [x] `entities` table exists with all required columns including `entity_path` and `entity_level`
- [x] `hierarchy_paths` closure table exists for efficient subtree queries
- [x] `entity_path` computed in app layer on create (service sets `/parent_path/uuid/`)

### ✅ E2 — SQLC queries: `db/queries/entities.sql`

> Already existed — full set of queries generated.

- [x] `CreateEntity` — inserts entity; app layer sets `entity_path` and `entity_level`
- [x] `GetEntity` — fetch by UUID within current tenant (RLS)
- [x] `ListEntities` — list all non-deleted for current tenant (RLS)
- [x] `GetEntityPath` — full ancestor path via `hierarchy_paths`
- [x] `GetEntitySubtree` / `GetEntityDescendants` — subtree queries
- [x] `CreateHierarchyPath` — inserts closure table rows (app layer calls for self + parent→child)

### ✅ E3 — `internal/core/entity/` bounded context

> Placed as its own module (not inside IAM) per project conventions — each module is independent.

- [x] `internal/core/entity/domain/entity.go` — `EntityNode`, `EntityType` enum (company/subsidiary/department/region/branch), `CreateEntityRequest`
- [x] `internal/core/entity/repository/repository.go` — `Repository` interface + `postgresRepo` implementation using `store.CreateEntity`, `store.GetEntity`, `store.ListEntities`, `store.CreateHierarchyPath`
- [x] `internal/core/entity/service/service.go` — `Service` interface + `entityService`; `CreateRoot` opens a dedicated `store.WithTenant` transaction so RLS is set correctly during provisioning
- [x] `internal/core/entity/entity.go` — facade re-exporting types and `NewService`/`NewRepository` constructors

### ✅ E4 — Wire `EntityScope` into `SessionService.Login()`

> Already implemented via `SessionRepository.ResolveEntityScope()` which calls `store.ResolveEntityScope()`. Platform/nil-entity users get `EntityScopeAll`. Branch nodes get `EntityScopeSubtree` with `PathPrefix`. Leaf nodes get `EntityScopeEntity`.

- [x] `ResolveEntityScope(ctx, entityID)` called in `Login()` via session repo
- [x] `uuid.Nil` entity → `EntityScopeAll` (platform/system users)
- [x] `entity_level == 1` → `EntityScopeAll` (tenant root admin)
- [x] `has_children == true` → `EntityScopeSubtree` with `PathPrefix`
- [x] Leaf node → `EntityScopeEntity`
- [x] SQLC query `ResolveEntityScope` exists in `db/sqlc/querier.go`
- [x] Non-fatal: falls back to `EntityScopeEntity` on DB error

### ✅ E5 — Auto-create root entity when tenant is provisioned

> Every tenant needs at least one entity (their company root) before any users can be created.

- [x] Added `OnProvision func(ctx, tenantID, tenantName)` callback field to `tenant.Dependencies`
- [x] `tenant.ProvisionTenant()` calls `OnProvision` after successful provisioning (best-effort, non-fatal)
- [x] Callers wire it: `deps.OnProvision = func(ctx, id, name) { entitySvc.CreateRoot(ctx, id, name) }`
- [x] Root entity ID is accessible via `entitySvc.List(ctx)` in the tenant context after provisioning

---

## ✅ Phase 9 — MFA (Multi-Factor Authentication)

> **Context:** MFA is a second verification step after password. After entering their password, users with MFA enabled must also enter a 6-digit TOTP code from an authenticator app (Google Authenticator, Authy, etc.). The code changes every 30 seconds based on a shared secret.
>
> **Why TOTP, not SMS:** SMS OTP is vulnerable to SIM-swapping. A malicious actor can port someone's phone number and receive their OTP. TOTP secrets never leave the user's device.
>
> **Who requires MFA:** Users with `finance.*` or `platform.*` permissions. Configurable for others via `iam.mfa.required` feature flag.

### ✅ M1 — DB migration: MFA fields

- [x] `mfa_secret text NULL` — AES-256-GCM encrypted TOTP secret — already present in `users` table
- [x] `mfa_enabled bool NOT NULL DEFAULT false` — already present in `users` table
- [x] Replay prevention: Redis key `mfa:replay:{userID}:{window}` with 90s TTL (documented, no migration needed)

### ✅ M2 — SQLC queries for MFA

> Run `make sqlc` after these query additions to generate Go code.

- [x] `GetUserMFASecret(userID)` — returns `mfa_enabled`, `mfa_secret` from `users`
- [x] `EnableMFA(userID, encryptedSecret)` — sets `mfa_secret = $2`, `mfa_enabled = TRUE`
- [x] `DisableMFA(userID)` — sets `mfa_secret = NULL`, `mfa_enabled = FALSE`
- [x] Queries added to `db/queries/users.sql`

### ✅ M3 — `UserService` — MFA methods

> Pure-stdlib TOTP (RFC 6238/4226 HMAC-SHA1) and AES-256-GCM encryption in `service/mfa_totp.go`.
> Secret lifecycle: generate → encrypt → cache (pending) → confirm → DB. Always encrypted at rest.

- [x] `InitiateMFA(ctx, userID) (*domain.MFASetup, error)` — 20-byte random secret, base32-encoded, AES-256-GCM encrypted, cached 10 min under `mfa:setup:{userID}`; returns `{Secret, QRURI}`
- [x] `ConfirmMFA(ctx, userID, code string) error` — get pending from cache → decrypt → verifyTOTP ±1 window → save to DB → clear cache
- [x] `ValidateMFACode(ctx, userID, code string) (bool, error)` — get from DB → decrypt → verifyTOTP ±1 window → replay check via `CheckAndMarkMFAReplay(userID, window)` with 90s TTL
- [x] `DisableMFA(ctx, userID, password string) error` — re-verify password with bcrypt, then clear DB secret
- [x] `domain.MFASetup` struct added to `domain/identity.go`
- [x] `UserConfig.MFAEncryptionKey []byte` and `UserConfig.MFAIssuer string` added

### ✅ M4 — Wire MFA into `SessionService.Login()`

- [x] `Login()` checks `user.MfaEnabled` after password passes
- [x] If enabled: generate 32-byte pending token, store in Redis as `mfa:login:pending:{token}` with 5-min TTL, return `(nil, pendingToken, ErrMFARequired)`
- [x] `CompleteMFALogin(ctx, pendingToken, mfaCode string) (*ResolvedSession, string, error)` added to `SessionService`
- [x] Pending token is single-use: deleted immediately after lookup regardless of TOTP outcome
- [x] On TOTP success: full session construction via `buildAndPersistSession()` helper
- [x] `ErrMFARequired` (202 Accepted) and `ErrMFAInvalid` (401) added to `internal/shared/errors/business.go`
- [x] MFA pending login methods (`StorePendingMFA`, `GetPendingMFA`, `DeletePendingMFA`) added to `SessionRepository`

### ✅ M5 — MFA handler endpoints

- [x] `POST /auth/mfa/initiate` — authenticated; calls `UserService.InitiateMFA`; returns `{qr_uri, secret}`
- [x] `POST /auth/mfa/confirm` — authenticated; calls `UserService.ConfirmMFA`; activates MFA
- [x] `POST /auth/mfa/complete` — public; exchanges `{pending_token, code}` for full session + cookie
- [x] `DELETE /auth/mfa` — authenticated; calls `UserService.DisableMFA`; requires password in body
- [x] `LoginHandler` updated to detect `ErrMFARequired` and return `{mfa_required: true, pending_token}`
- [x] `/api/v1/auth/mfa/complete` added to public whitelist (no tenant context required)
- [x] `iam.MFASetup` re-exported from facade in `internal/core/iam/iam.go`

---

## ✅ Phase 10 — Password Management

> **Context:** Password handling requires careful implementation. Passwords must be hashed with bcrypt (slow by design — makes brute force expensive). Reset tokens must be single-use, time-limited, and stored as hashes.

### ✅ P1 — Password reset flow

- [x] `POST /auth/forgot-password` — always returns 200 (prevents user enumeration)
  - Calls `UserService.ForgotPassword(ctx, email)` which returns `("", uuid.Nil, nil)` when email not found
  - Generates 32-byte random token; stores SHA-256 hash in `password_reset_tokens` with 1-hour TTL
  - **NOTE(notification):** Email delivery is a TODO — wire a notification service to send the raw token
- [x] `POST /auth/reset-password` — validates token + sets new password
  - Checks: token exists, not expired, not used, password strength, not reused from last 5
  - Marks token as used only after password update succeeds (no double-use on transient failure)
  - Handler file: `internal/api/handlers/auth/password.go`
- [x] Both endpoints added to public whitelist (no tenant context required)

### ✅ P2 — DB migration: `password_reset_tokens` table

- [x] Migration `000308_iam_password_reset.up.sql`
- [x] `password_reset_tokens`: `id`, `tenant_id`, `user_id`, `token_hash UNIQUE`, `expires_at`, `used_at NULL`, `created_at`
- [x] Indexes on `token_hash` and `(tenant_id, user_id)`
- [x] `password_history jsonb DEFAULT '[]'` column added to `users` table

### ✅ P3 — Password strength validation

- [x] `validatePasswordStrength(password)` — 12+ chars, uppercase + lowercase + digit + special
- [x] `isPasswordReused(plain, history)` — bcrypt compare against last ≤5 hashes
- [x] `prependHistory(newHash, existing)` — prepends and caps at 5 entries
- [x] Applied in `ChangePassword()` (strength + history check before update)
- [x] Applied in `ResetPassword()` (strength + history check before update)
- [x] History persisted via `UpdatePasswordAndHistory` SQLC query (JSON array on `users.password_history`)
- [x] New error codes: `ErrPasswordTooWeak` (400), `ErrPasswordReused` (400), `ErrPasswordResetToken*` (404/410)
- [x] `PasswordResetToken` domain type re-exported from IAM facade
- [x] **SQLC**: Run `make sqlc` after migration to generate `CreatePasswordResetToken`, `GetPasswordResetToken`, `MarkPasswordResetTokenUsed`, `GetUserPasswordHistory`, `UpdatePasswordAndHistory`
- [x] **NOTE**: After `make sqlc`, verify generated param field names in `UpdatePasswordAndHistory` params (SQLC names positional params `Column2`, `Column3` etc. — adjust repo calls if needed)

---

## ✅ Phase 11 — OAuth / OIDC / SAML (SSO)

> **Context:** Some companies use single sign-on (SSO) — instead of managing passwords in our system, they authenticate with Google Workspace, Azure AD, or their own identity provider. OIDC (OpenID Connect) is the modern protocol; SAML 2.0 is the older enterprise standard.
>
> **JIT Provisioning:** When an SSO user logs in for the first time, we automatically create a User record — this is "just-in-time provisioning", controlled by the `iam.sso.auto_provision` feature flag.

### ✅ O1 — OAuth/OIDC handler

- [x] `GET /auth/oauth/:provider` — redirects to provider's authorize URL
  - Reads `tenant_id` from query param, looks up `sso_providers` config
  - Generates 24-byte CSRF state, stores `sso:state:{state}` in Redis (10-min TTL)
  - Builds provider-specific auth URL (Google, Microsoft)
- [x] `GET /auth/oauth/:provider/callback`:
  - Validates CSRF state (single-use — deleted immediately after lookup)
  - Exchanges code for access token via POST to provider token endpoint
  - Fetches user info from provider userinfo endpoint
  - If user exists: calls `SessionService.LoginWithSSO` → full session
  - If user doesn't exist and `provider.AutoProvision=true`: JIT-provisions user then proceeds
  - If user doesn't exist and `AutoProvision=false`: returns 403
- [x] `SessionService.LoginWithSSO(ctx, user)` added — skips MFA (IdP is the second factor)
- [x] Support providers: Google, Microsoft (Azure AD)
- [x] `sso_providers` table: `id`, `tenant_id`, `provider`, `client_id`, `client_secret_enc` (AES-256-GCM), `scopes`, `redirect_uri`, `extra_params` JSONB, `auto_provision`, `default_entity_id`, `is_active`
- [x] Migration `000309_iam_sso_providers.up.sql` — RLS with tenant isolation + admin bypass
- [x] SQLC queries: `GetSSOProvider`, `UpsertSSOProvider`, `DeactivateSSOProvider`, `ListSSOProviders` (run `make sqlc`)
- [x] `SSOService` interface + `ssoService` implementation (`internal/core/iam/service/sso.go`)
- [x] `SSORepository` interface + `ssoRepo` implementation (`internal/core/iam/repository/sso.go`)
- [x] `GET /api/v1/auth/oauth/*` added to public whitelist
- [x] `SSOService`, `SSORepository`, `SSOConfig`, `OAuthProvider*` constants re-exported from IAM facade
- [x] `SSOService` optional field added to `handlers.Dependencies`
- [x] **NOTE**: `default_entity_id` must be configured when `auto_provision=true`; JIT provisioning fails if nil

### 🔲 O2 — SAML 2.0 handler (later, enterprise requirement)

- [ ] `GET /auth/saml/:tenant/metadata` — returns SP metadata XML
- [ ] `POST /auth/saml/:tenant/callback` — receives SAML assertion, validates, creates session
- [ ] Per-tenant SAML configuration (IdP metadata URL, entity ID, certificate)

---

## 🔲 Phase 12 — API Key Authentication

> **Context:** Machine-to-machine integrations (mobile apps, accounting sync tools, webhooks) should not use browser sessions. They use API keys — a static credential that grants a specific, limited set of permissions.
>
> **Important:** API key sessions are NOT stored in `user_sessions` to avoid bloating the table with high-frequency API calls. Instead they are built fresh on each request (with Redis caching).
>
> The code has `ActorAPI` and `APIDomain` already defined in domain — this phase wires them up.

### A1 — DB migration: `api_keys` table

- [ ] Create `api_keys` table: `id`, `tenant_id`, `name`, `key_hash` (SHA-256), `scopes text[]`, `expires_at`, `created_by`, `revoked_at`, `created_at`
- [ ] Add index on `key_hash`

### A2 — SQLC queries for API keys

- [ ] `CreateAPIKey`, `GetAPIKeyByHash`, `RevokeAPIKey`, `ListAPIKeysByTenant`

### A3 — API key service

- [ ] `CreateAPIKey(ctx, params) (*APIKey, string, error)` — returns key + plaintext secret (shown once only)
- [ ] `ValidateAPIKey(ctx, rawKey string) (*ResolvedSession, error)`:
  - Hash incoming key, look up in DB
  - Verify not revoked, not expired
  - Build a minimal `ResolvedSession` from the key's `scopes` (scopes are the ceiling of permissions)
  - Cache the resolved session keyed by `sha256hex(rawKey)` with 5-minute TTL
- [ ] `RevokeAPIKey(ctx, keyID uuid.UUID) error`
- [ ] `ListAPIKeys(ctx, tenantID uuid.UUID) ([]*APIKey, error)`

### A4 — Wire into authentication middleware

- [ ] In `Authenticate` middleware: detect `Authorization: Bearer <key>` vs session cookie
- [ ] If Bearer token does not match any session, try `apiKeySvc.ValidateAPIKey()`
- [ ] Set `ActorAPI` domain in the resulting `Principal`

---

## 🔲 Phase 13 — MRA Registry & BootService

> **Context:** The docs describe a `Module / Resource / Action` (MRA) registry — three database tables that define every feature in the system. The `BootService` uses these tables + feature flags + permissions to generate the navigation sidebar that the AMIS frontend renders.
>
> Right now, navigation and feature registration happen manually. The MRA system makes it data-driven — adding a new module means inserting a row, and the UI picks it up automatically.

### B1 — DB migration: `modules`, `resources`, `actions` tables

- [ ] `modules` table: `id`, `slug` (unique), `label`, `icon`, `nav_order`, `is_active`
- [ ] `resources` table: `id`, `module_id`, `slug`, `label`, `nav_url`, `nav_order`, unique `(module_id, slug)`
- [ ] `actions` table: `id`, `resource_id`, `slug`, `label`, `http_method`, unique `(resource_id, slug)`
- [ ] Auto-seed trigger: on `INSERT INTO modules`, auto-create a feature flag definition for the module
- [ ] Auto-seed trigger: on `INSERT INTO resources`, auto-create a resource-level feature flag

### B2 — Seed MRA rows for existing modules

> Insert rows for every module/resource/action that already exists in the codebase. This is a one-time data migration.

- [ ] Finance module: `finance` → `transactions`, `accounts`, `receivables`, `payables`, `reports`
- [ ] People module: `people` → `employees`, `persons`
- [ ] Settings module: `settings` → `iam`, `general`, `modules`
- [ ] Each resource: seed standard actions (`read`, `create`, `update`, `delete`) + domain-specific actions (`approve`, `post`, `void`, `export` etc.)

### B3 — `BootService` — build app shell schema

> The `BootService` is called when the AMIS frontend loads (`GET /schema/boot`). It returns the navigation structure and page schemas the frontend needs to render. It reads MRA rows, checks which flags are on for the tenant, and which permissions the user has.

- [ ] `BootService.Build(ctx, session *ResolvedSession) (*BootSchema, error)`:
  - Query `modules` where `is_active = true` and module flag is enabled for tenant
  - For each module, query its `resources` where resource flag is enabled
  - Filter resources to those where user has at least `read` permission
  - Return structured nav + page schema pointers
- [ ] `GET /schema/boot` handler — returns `BootSchema` JSON for AMIS

### B4 — Permission key derivation from MRA

> Replace hardcoded permission strings like `"finance.receivables.invoices"` with keys derived from MRA slugs. This ensures the permission catalogue stays in sync with the MRA registry.

- [ ] Add `permissions` table: `(module_slug, resource_slug, action_slug, full_key, description)`
- [ ] Seed permission rows from the MRA `actions` table
- [ ] Update `SeedDefaultRoles()` to derive permission keys from the `permissions` table instead of hardcoded strings

---

## 🔲 Phase 14 — Audit Trail

> **Context:** Financial systems need an audit trail — a log of who did what, when, on which record. This is a compliance requirement. The IAM module should record all sensitive operations (login, role assignment, policy change, session invalidation).

### AU1 — DB migration: `audit_logs` table

- [ ] Create `audit_logs`: `id`, `tenant_id`, `user_id`, `actor_type`, `action`, `resource_type`, `resource_id`, `changes jsonb`, `ip_address`, `user_agent`, `created_at`
- [ ] Add index on `(tenant_id, created_at)` for tenant-scoped queries
- [ ] RLS policy: tenants can read their own audit logs; platform can read all

### AU2 — `AuditService`

- [ ] `AuditService` interface: `Log(ctx, event AuditEvent) error`
- [ ] `AuditEvent` struct: `Action`, `ResourceType`, `ResourceID`, `Changes map[string]any`
- [ ] Async implementation — write to audit log in a goroutine so it never blocks the request
- [ ] Wire into IAM: log all `Login`, `Logout`, `AssignRole`, `RevokeRole`, `AddPolicy`, `RemovePolicy`, `CreateUser`, `SuspendUser` operations

---

## 🔲 Phase 15 — Docs Corrections

> **Context:** During implementation, some decisions were made that diverged from the docs. The docs need to be updated to reflect the actual code.

### DC1 — Fix Casbin model docs

> The docs in `05-casbin-policy-engine.md` show `keyMatch2(r.act, p.act)`. This is wrong — the actual code uses `keyMatch(r.act, p.act)`. `keyMatch2` would break wildcard `*` actions.

- [ ] Update `docs/reference/modules/iam/05-casbin-policy-engine.md` matcher section to show `keyMatch(r.act, p.act)`
- [ ] Add a note explaining why (`keyMatch2` uses `:param` syntax, not glob `*`)

### DC2 — Fix `session.Can()` docs

> The docs show `session.Can("finance.transactions", "approve")` with 2 args. The actual method is `Can(permission string)` with 1 arg (the full dot-notation key). The 2-arg form is `CanDo(resource, action string)`.

- [ ] Update all doc examples that show `session.Can(resource, action)` to `session.CanDo(resource, action)` or `session.Can("resource.action")`
- [ ] Files to update: `10b-session-precomputation.md`, `12b-http-middleware.md`, `15b-cross-module-integration.md`

### DC3 — Fix UserType values in docs

> The docs say user types are `'platform' | 'tenant' | 'portal' | 'third_party'`. The DB stores ALL-CAPS enums (`"SYSADMIN"`, `"INTERNAL"`, `"PORTAL"`, `"CUSTOMER"`, `"API"`). The translation happens in `ActorTypeFromUserType()`.

- [ ] Update `06b-authentication.md` users table to show real DB values
- [ ] Document `ActorTypeFromUserType()` as the canonical mapping point
- [ ] Update `04-domain-model.md` user type section

### DC4 — Fix session TTL in docs

> The docs say 24h session lifetime. The code defaults to 8h (with a TODO to make it configurable per tenant via settings).

- [ ] Update `06b-authentication.md` login flow to say "8h default, configurable via tenant setting `iam.session_ttl_hours`"

### DC5 — Document the actual module structure

> The docs describe `internal/platform/` as the facade. The actual code is `internal/core/iam/`. The Platform facade described in the docs does not exist.

- [ ] Update `00b-code-architecture.md` repository layout to reflect actual `internal/core/` structure
- [ ] Document that FeatureFlag and Settings are separate bounded contexts, not parts of IAM
- [ ] Add a section describing how business modules should import these bounded contexts

---

## 🔲 Phase 16 — End-to-End Verification

> These are integration tests that prove the entire pipeline works together. Each test exercises multiple layers.

### V1 — Login returns populated session

- [ ] POST `/auth/login` with valid credentials
- [ ] Assert response has 200 + `HttpOnly` cookie
- [ ] Assert `ResolvedSession.Permissions` is not empty (has at least one key)
- [ ] Assert `ResolvedSession.Configuration.Flags` is populated (Phase 7 must be done)
- [ ] Assert `ResolvedSession.EntityScope.Type` is set (Phase 8 must be done)

### V2 — Protected route end-to-end

- [ ] Register `GET /api/finance/invoices` with `Authenticate + Authorize("finance.receivables.invoices.read")`
- [ ] Call with no cookie → assert 401
- [ ] Call with valid session but user lacks permission → assert 403
- [ ] Call with valid session and correct permission → assert 200

### V3 — Feature flag gates route

- [ ] Disable `finance` feature flag for a tenant
- [ ] Login as user in that tenant
- [ ] Call `GET /api/finance/invoices` (which has `RequireFlag("finance")` middleware)
- [ ] Assert 403 with "feature not enabled" message

### V4 — Role expiry: lazy revoke

- [ ] Assign `role:finance-manager` to a user with `ExpiresAt = now() - 1 minute`
- [ ] Call `Enforce()` for a finance permission
- [ ] Assert: returns false AND the expired assignment is deactivated in DB

### V5 — Tenant isolation: cross-tenant policy leak

- [ ] Create two tenants A and B, each with their own `role:finance-manager`
- [ ] Tenant A user has finance permissions; Tenant B user does not
- [ ] Assert `Enforce()` for Tenant B user returns false for Tenant B's domain
- [ ] Assert Tenant B's session cannot access Tenant A's data via any route

---

## Completion Summary

| Phase | Description | Status |
|---|---|---|
| 0 | DB migrations + SQLC generation | ✅ Done |
| 1 | Domain layer (authz, session, identity, errors) | ✅ Done |
| 2 | Repository layer (authz repo + Casbin adapter, user repo, session repo) | ✅ Done |
| 3 | Service layer (AuthzService, UserService, SessionService) | ✅ Done |
| 4 | IAM facade (iam.go) + seed.go | ✅ Done |
| 5 | HTTP middleware (Authenticate, Authorize) + Login/Logout handlers | ✅ Done |
| 6 | Security hardening (SHA-256 tokens, tenant guard, RLS fix, AssignableTo) | ✅ Done |
| 7 | Wire flags + settings into session at login | 🔲 Not started |
| 8 | Entity module (table, repo, scope resolution at login) | 🔲 Not started |
| 9 | MFA / TOTP | 🔲 Not started |
| 10 | Password management (reset flow, strength validation, history) | 🔲 Not started |
| 11 | OAuth / OIDC / SAML SSO | 🔲 Not started |
| 12 | API key authentication | 🔲 Not started |
| 13 | MRA registry + BootService | 🔲 Not started |
| 14 | Audit trail | 🔲 Not started |
| 15 | Docs corrections | 🔲 Not started |
| 16 | End-to-end integration tests | 🔲 Not started |
