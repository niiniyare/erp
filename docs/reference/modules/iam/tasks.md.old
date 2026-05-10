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

##  Phase 7 — Session Construction: Wire Flags + Settings

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

###  O2 — SAML 2.0 handler (later, enterprise requirement)

- [ ] `GET /auth/saml/:tenant/metadata` — returns SP metadata XML
- [ ] `POST /auth/saml/:tenant/callback` — receives SAML assertion, validates, creates session
- [ ] Per-tenant SAML configuration (IdP metadata URL, entity ID, certificate)

---

## ✅ Phase 12 — API Key Authentication

> **Context:** Machine-to-machine integrations (mobile apps, accounting sync tools, webhooks) should not use browser sessions. They use API keys — a static credential that grants a specific, limited set of permissions.
>
> **Important:** API key sessions are NOT stored in `user_sessions` to avoid bloating the table with high-frequency API calls. Instead they are built fresh on each request (with Redis caching).
>
> The code has `ActorAPI` and `APIDomain` already defined in domain — this phase wires them up.

### ✅ A1 — DB migration: `api_keys` table

- [x] `db/migration/000310_iam_api_keys.up.sql` — table with `id`, `tenant_id`, `name`, `key_hash`, `scopes TEXT[]`, `created_by`, `expires_at`, `revoked_at`, `last_used_at`, `created_at`
- [x] Indexes on `key_hash` (UNIQUE) and `tenant_id`
- [x] RLS: `application_role` full tenant isolation; `admin_role` bypass (required for cross-tenant hash lookup)

### ✅ A2 — SQLC queries: `db/queries/api_keys.sql`

- [x] `CreateAPIKey` — inserts with `current_tenant_id()`; RETURNING *
- [x] `GetAPIKeyByHash` — cross-tenant hash lookup (no tenant filter); excludes `key_hash` from SELECT; filters `revoked_at IS NULL AND (expires_at IS NULL OR expires_at > NOW())`
- [x] `RevokeAPIKey` — sets `revoked_at = NOW()` for `current_tenant_id()`
- [x] `ListAPIKeys` — all keys for `current_tenant_id()`, newest first; excludes `key_hash`

### ✅ A3 — API key service: `internal/core/iam/service/apikey.go`

- [x] `CreateAPIKey` — generates `eak_{32-byte-hex}` bearer token, stores SHA-256 hash, returns plaintext once
- [x] `ValidateAPIKey` — cache-aside (Redis 5-min TTL); DB fallback via hash lookup; builds minimal `ResolvedSession` from scopes
- [x] `RevokeAPIKey` — sets revoked_at; cache entry expires naturally (TTL)
- [x] `ListAPIKeys` — delegates to repo; no tenant param (uses ctx)

### ✅ A4 — Wire into authentication middleware

- [x] `AuthConfig.APIKeyService` field added to `session_middleware.go`
- [x] `Authenticate` detects `eak_` prefix on Bearer token → routes to `ValidateAPIKey`
- [x] `authenticateMiddleware()` in routes injects `APIKeyService` from `Dependencies`
- [x] Management routes: `POST/GET /api/v1/auth/api-keys`, `DELETE /api/v1/auth/api-keys/:id`

---

## ✅ Phase 13 — MRA Registry & BootService

> **Context:** The docs describe a `Module / Resource / Action` (MRA) registry — three database tables that define every feature in the system. The `BootService` uses these tables + feature flags + permissions to generate the navigation sidebar that the AMIS frontend renders.
>
> Right now, navigation and feature registration happen manually. The MRA system makes it data-driven — adding a new module means inserting a row, and the UI picks it up automatically.

### ✅ B1 — DB migration: `modules`, `resources`, `actions` tables

> Already existed in migrations 000015-000017 with a richer schema than originally planned.

- [x] `modules`: `id`, `slug` (unique), `name`, `display_name`, `icon`, `nav_order`, `is_active`, `scope`, `category`, `module_type`
- [x] `resources`: `id`, `module_id`, `slug`, `name`, `display_name`, `nav_url`, `nav_order`, `resource_type`, unique `(module_id, slug)`
- [x] `actions`: `id`, `resource_id`, `slug`, `name`, `action_type`, `http_method`, `scope`, `risk_level`, `action_category`, unique `(resource_id, slug)`

### ✅ B2 — Seed MRA rows for existing modules

- [x] `db/migration/000018_platform_registry.up.sql` — seeds all four core modules:
  - Finance: 5 resources + CRUD + domain-specific actions (approve, post, void, export)
  - People: 2 resources + CRUD
  - Settings: 3 resources + read/update
  - IAM: 5 resources + CRUD + revoke (sessions)

### ✅ B3 — `BootHandler` — build app shell schema

- [x] `GET /schema/boot` — `internal/api/handlers/schema/boot.go`
  - Feature flag gate: `sess.Configuration.Flags["{slug}.enabled"]` — absent key = allowed
  - Permission gate: `sess.Can("{module}.{resource}.read")`
  - Returns AMIS `app` type JSON with `pages` array
- [x] SQLC queries: `ListActiveSystemModules`, `ListActiveResourcesByModule` in `db/queries/boot.sql`
- [x] Route wired under `GET /api/v1/schema/boot` (authenticated)

### ✅ B4 — Permission key derivation from MRA

- [x] `permissions` table in `000018_platform_registry.up.sql`:
  - `full_key` is `GENERATED ALWAYS AS (module_slug || '.' || resource_slug || '.' || action_slug) STORED`
  - Trigger `fn_auto_create_permission` — fires `AFTER INSERT ON actions`, auto-inserts permission row
  - Unique constraint on `full_key` — canonical permission catalogue stays in sync automatically

---

## ✅ Phase 14 — Audit Trail

> **Context:** Financial systems need an audit trail — a log of who did what, when, on which record. This is a compliance requirement. The IAM module records sensitive operations via a universal DB trigger system (auto) + `CreateAuditEvent` for application-level events.

### ✅ AU1 — DB migration: `audit_log` table

- [x] `000450_audit_log.up.sql` — rich schema: `id`, `tenant_id`, `user_id`, `event_type`, `event_category`, `severity`, `risk_score`, `context JSONB`, `compliance_flags JSONB`, `ip_address`, `user_agent`, `session_id`, plus FK columns for resource/action/role/permission
- [x] RLS: tenant isolation for SELECT (application_role); full bypass for admin_role
- [x] `000451_audit_funcs.up.sql` — universal trigger system: `audit_trigger_function()`, `enable_audit_on_table()`, `enable_audit_on_schema()`, `get_audit_statistics()`

### ✅ AU2 — `AuditService` at `internal/core/audit/`

- [x] `Service` + `Repository` interfaces (`interface.go`) — full CRUD, analytics, forensics, bulk ops
- [x] Domain types in `model.go` — `AuditEvent`, `CreateAuditEventRequest`, analytics structs
- [x] `validation.go` — field-level validation helpers
- [x] `service.go` + `repository.go` — `CreateAuditEvent` fully wired; analytics methods stubbed
- [x] SQLC queries already generated (`db/sqlc/audit.sql.go`)
- [x] `AuditService audit.Service` added to `handlers.Dependencies`
- [x] `GET /api/v1/audit-logs` handler at `internal/api/handlers/audit/handler.go` — gated on `iam.sessions.read`
- [x] Route registered in `registerAuditAPI`
- [x] **FIXED**: `validation.go` `ValidEventCategories` updated to uppercase DB values (`ACCESS`, `ADMIN`, `DATA`, `AUTH`, `SYSTEM`, `COMPLIANCE`); `ValidSeverities` updated to `LOW`, `INFO`, `WARN`, `HIGH`, `CRITICAL`; `ValidDecisions` updated to `ALLOW`, `DENY`, `WARN`
- [ ] **TODO**: Implement analytics repository stubs (`GetAuditStatsByCategory`, `GetUserRiskProfile`, etc.)

---

## ✅ Phase 15 — Docs Corrections

> **Context:** During implementation, some decisions were made that diverged from the docs. The docs need to be updated to reflect the actual code.

### ✅ DC1 — Fix Casbin model docs

> The docs in `05-casbin-policy-engine.md` show `keyMatch2(r.act, p.act)`. This is wrong — the actual code uses `keyMatch(r.act, p.act)`. `keyMatch2` would break wildcard `*` actions.

- [x] Updated `docs/reference/modules/iam/05-casbin-policy-engine.md` matcher section to show `keyMatch(r.act, p.act)`
- [x] Added note explaining why (`keyMatch2` uses `:param` syntax, not glob `*`)
- [x] Split pattern reference table into obj (keyMatch2) and act (keyMatch) sections

### ✅ DC2 — Fix `session.Can()` docs

> The docs show `session.Can("finance.transactions", "approve")` with 2 args. The actual method is `Can(permission string)` with 1 arg (the full dot-notation key). The 2-arg form is `CanDo(resource, action string)`.

- [x] Updated `10b-session-precomputation.md` — corrected `Can` signature, changed 2-arg calls to `CanDo`, fixed TTL references (24h→8h)
- [x] Updated `12b-http-middleware.md` — `RequirePermission` now calls `CanDo(resource, action)`
- [x] Updated `15b-cross-module-integration.md` — integration table and Finance example now use `CanDo`

### ✅ DC3 — Fix UserType values in docs

> The DB stores ALL-CAPS enums (`"SYSADMIN"`, `"INTERNAL"`, `"PORTAL"`, `"CUSTOMER"`, `"API"`). The translation happens in `ActorTypeFromUserType()`.

- [x] Updated `06b-authentication.md` users table comment to show real DB values with `ActorTypeFromUserType()` reference

### ✅ DC4 — Fix session TTL in docs

> The code defaults to 8h (configurable via tenant setting `iam.session_ttl_hours`).

- [x] Updated `06b-authentication.md` login flow: `expires_at = NOW() + 8h (default; configurable via tenant setting iam.session_ttl_hours)`
- [x] Updated `10b-session-precomputation.md` session lifetime reference and invalidation table

### ✅ DC5 — Document the actual module structure

> The docs described `internal/platform/` as the facade. The actual code is `internal/core/iam/`.

- [x] Rewrote `00b-code-architecture.md` Repository Layout section to show actual `internal/core/` structure
- [x] Added note that `internal/platform/` facade does not exist; import from `internal/core/iam`

---

## ✅ Phase 16 — End-to-End Verification

> These are integration tests that prove the entire pipeline works together. Each test exercises multiple layers.
> All tests use in-memory / mock dependencies — no database required.

### ✅ V1 — Login returns populated session

- [x] POST `/api/v1/auth/login` with mock SessionService returning a populated ResolvedSession
- [x] Assert 200 + `HttpOnly` cookie set with raw token value
- [x] Assert `ResolvedSession.Permissions` is non-empty in the response body
- [x] Test file: `internal/api/handlers/pipeline_test.go` — `TestV1_Login_Returns200_WithCookieAndPermissions`

### ✅ V2 — Protected route end-to-end

- [x] Call with no cookie → assert 401 (Authenticate blocks)
- [x] Call with valid session but no `finance.accounts.read` permission → assert 403 (Authorize blocks)
- [x] Call with valid session + correct permission → assert 200 (passes through)
- [x] Test file: `internal/api/handlers/pipeline_test.go` — `TestV2_*`

### ✅ V3 — Feature flag gates route

- [x] Added `RequireFlag(flagKey string)` middleware to `internal/api/middleware/session_middleware.go`
- [x] Session with `finance` flag = false → 403 "feature not enabled"
- [x] Session with `finance` flag = true → 200 passes through
- [x] Test file: `internal/api/handlers/pipeline_test.go` — `TestV3_*`

### ✅ V4 — Role expiry: lazy revoke

- [x] Assign `role:finance-manager` to a subject; add invoice/* allow policy
- [x] Verify Enforce returns true before expiry
- [x] Simulate repo returning the role as expired on next call
- [x] Assert Enforce returns false (role lazily removed from in-memory enforcer)
- [x] Assert `DeactivateRoleAssignment` was called with the expired role
- [x] Test file: `internal/core/iam/authz_enforce_test.go` — `TestV4_ExpiredRole_LazilyCleaned_DeniesAccess`

### ✅ V5 — Tenant isolation: cross-tenant policy leak

- [x] Assign role + policy in tenantA domain only
- [x] Assert Enforce(subject, tenantA, ...) → true
- [x] Assert Enforce(subject, tenantB, ...) → false (Casbin `r.dom == p.dom` prevents cross-domain match)
- [x] Test file: `internal/core/iam/authz_enforce_test.go` — `TestV5_CrossTenantPolicyLeak_Denied`

---

## Completion Summary

| Phase | Description | Status |
|---|---|---|
| 0 | DB migrations + SQLC generation | ✅ Done |
| 1 | Domain layer (authz, session, identity, errors) | ✅ Done |
| 2 | Repository layer (authz repo + Casbin adapter, user repo, session repo) | ✅ Done |
| 3 | Service layer (AuthzService, UserService, SessionService) | ✅ Done |
| 4 | IAM facade (iam.go) + seed.go | ✅ Done |

---

> **NOTE — Phases 17–20 below were added after a code audit revealed that several items marked ✅ in the summary above were not actually implemented in the codebase. Work through these phases in order before closing the IAM module.**

---

## ✅ Phase 17 — UserService: Fix Missing Implementation

> **Resolution:** Code audit found `internal/core/iam/service/identity.go` and `service/mfa_totp.go` already contained the full `UserService` implementation. The task was to remove `service/user.go` which contained duplicate declarations and was causing build errors. Fixed by replacing `service/user.go` with an empty package stub that redirects to `identity.go`.

### ✅ U1 — Create `internal/core/iam/service/user.go`

- [x] `UserService` interface defined with all methods — exists in `service/identity.go`
- [x] `userService` struct + `NewUserService` + `NewUserServiceWithConfig` — exists in `service/identity.go`
- [x] `UserConfig` struct with `MaxFailedAttempts`, `LockoutDuration`, `MFAEncryptionKey`, `MFAIssuer` — exists in `service/identity.go`

### ✅ U2 — Authenticate + brute-force protection

- [x] `Authenticate(ctx, email, password string) (*domain.User, error)` — implemented in `service/identity.go`

### ✅ U3 — User CRUD methods

- [x] `GetUser`, `GetUserByEmail`, `GetUserWithDetails`, `CreateUser`, `UpdateUser`, `ListUsers` — all in `service/identity.go`

### ✅ U4 — Password management

- [x] `ChangePassword`, `ForgotPassword`, `ResetPassword` — all in `service/identity.go`

### ✅ U5 — MFA secret lifecycle

- [x] `InitiateMFA`, `ConfirmMFA`, `ValidateMFACode`, `DisableMFA` — all in `service/identity.go` + `service/mfa_totp.go`

---

## ✅ Phase 18 — HTTP Middleware: Rewrite for Session-Based Auth

> **Resolution:** All three files were already correctly implemented. `session_middleware.go` had `Authenticate`, `Authorize`, `RequireFlag`. Fix applied: added missing `LocalsKeyPrincipal` + `cache.TenantIDKey` injection via new `setSessionLocals()` helper. `jwt_auth.go` emptied to remove duplicate `AuthConfig`/`Authenticate` declarations. `authorization.go` already had context helpers and `AuthorizeCasbin`.

### ✅ MW1 — Rewrite `internal/api/middleware/jwt_auth.go` → session-based Authenticate

> Delete the commented-out JWT code. Write a clean `Authenticate(cfg AuthConfig)` middleware that validates sessions. The name `jwt_auth.go` is a legacy filename — keep it to avoid renaming across the codebase.

- [ ] Delete all commented-out code in `jwt_auth.go`
- [ ] `AuthConfig` struct:
  ```go
  type AuthConfig struct {
      SessionSvc  iam.SessionService
      APIKeySvc   iam.APIKeyService  // optional — nil = no API key auth
      CookieName  string             // default "awo_session"
  }
  ```
- [ ] `Authenticate(cfg AuthConfig) fiber.Handler`:
  - Extract token: check `Authorization: Bearer <token>` header first, then cookie `cfg.CookieName`
  - If no token found: return 401 `{"error": "authentication required"}`
  - If token has prefix `eak_`: call `cfg.APIKeySvc.ValidateAPIKey(ctx, token)` — nil APIKeySvc returns 401
  - Otherwise: call `cfg.SessionSvc.ValidateSession(ctx, token)`
  - On `ErrUnauthorized` / `ErrSessionExpired`: return 401
  - On success: set `c.Locals(iam.LocalsKeySession, resolved)`
  - Set `c.Locals(iam.LocalsKeyPrincipal, resolved.ToPrincipal())`
  - Inject tenant into Go context: `ctx = context.WithValue(c.Context(), cache.TenantIDKey, resolved.TenantID.String())`; `c.SetUserContext(ctx)`

### MW2 — Add `Authorize(permission string)` to `authorization.go`

> The fast path. Reads pre-computed permissions from the session — zero DB, zero Casbin. Every hot-path route uses this. Keep `AuthorizeCasbin()` for management operations that need live Casbin enforcement.

- [ ] Add `Authorize(permission string) fiber.Handler`:
  ```go
  func Authorize(permission string) fiber.Handler {
      return func(c *fiber.Ctx) error {
          sess, ok := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
          if !ok || sess == nil {
              return c.Status(401).JSON(fiber.Map{"error": "authentication required"})
          }
          if !sess.Can(permission) {
              return c.Status(403).JSON(fiber.Map{"error": "permission denied: " + permission})
          }
          return c.Next()
      }
  }
  ```
- [ ] Add `RequireFlag(flagKey string) fiber.Handler`:
  ```go
  func RequireFlag(flagKey string) fiber.Handler {
      return func(c *fiber.Ctx) error {
          sess := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)
          if !sess.FeatureEnabled(flagKey) {
              return c.Status(403).JSON(fiber.Map{"error": "feature not enabled for your organisation"})
          }
          return c.Next()
      }
  }
  ```

### MW3 — Add context helpers to `authorization.go` (or a new `context_helpers.go`)

> Handlers must never read tenant_id or user_id from query params. They always read from the session. These helpers make that pattern one line.

- [ ] `ContextSession(c *fiber.Ctx) *iam.ResolvedSession` — `c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)`
- [ ] `ContextTenantID(c *fiber.Ctx) uuid.UUID` — `ContextSession(c).TenantID`
- [ ] `ContextUserID(c *fiber.Ctx) uuid.UUID` — `ContextSession(c).UserID`
- [ ] `ContextPrincipal(c *fiber.Ctx) *iam.Principal` — `c.Locals(iam.LocalsKeyPrincipal).(*iam.Principal)`

### ✅ MW1–MW4 — All middleware tasks

- [x] `jwt_auth.go` — emptied (legacy JWT removed, no duplicate declarations)
- [x] `session_middleware.go` — `Authenticate`, `Authorize`, `RequireFlag` all present; fixed to call `setSessionLocals()` which sets `LocalsKeySession`, `LocalsKeyPrincipal`, and injects `cache.TenantIDKey` into Go context
- [x] `authorization.go` — context helpers (`ContextSession`, `ContextTenantID`, `ContextUserID`, `ContextPrincipal`) and `AuthorizeCasbin` present
- [x] `routes.go` — `authenticateMiddleware()` calls `middleware.Authenticate(cfg)` with correct `SessionService` field name; `APIKeyService` injected via `Dependencies`

---

## ✅ Phase 19 — DB & SQLC Verification

### ✅ DB1 — Verify migration 000309: `sso_providers`

- [x] `db/migration/000309_iam_sso_providers.up.sql` exists
- [x] `GetSSOProvider`, `UpsertSSOProvider`, `DeactivateSSOProvider`, `ListSSOProviders` confirmed in `db/sqlc/querier.go`

### ✅ DB2 — Verify migration 000310: `api_keys`

- [x] `db/migration/000310_iam_api_keys.up.sql` exists
- [x] `CreateAPIKey`, `GetAPIKeyByHash`, `RevokeAPIKey`, `ListAPIKeys` confirmed in `db/sqlc/querier.go`

### ✅ DB3 — Full build check

- [x] `go build ./internal/core/iam/...` — zero errors (confirmed by user)
- [x] `go build ./internal/api/middleware/...` — zero errors (confirmed by user)

---

## Phase 20 — API Tests (Behavioral Verification)

> **Test script written:** `scripts/api/services/iam.sh` — run with:
> ```
> bash scripts/api/run.sh --service iam
> ```
> or interactively: `bash scripts/api/run.sh`
>
> **Prerequisites:** Server running + seed data applied (`bash scripts/seed.sh`).
>
> **Manual tests** (stateful/destructive — run separately):
> - AT3 — account lockout: hammer AT2 five times first, then verify 423 on correct login
> - AT6 — missing permission 403: create a user with no finance permissions
> - AT9 — MFA confirm: scan QR from AT8 and submit real TOTP code

---

### AT1 — Login: valid credentials, no MFA

**Expected behavior:** User provides correct email and password. Server builds a full session (permissions, flags, settings, entity scope), stores the token hash in DB, caches the resolved session in Redis, returns the session data and sets an `HttpOnly` cookie.

**Request:**
```
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@tenant.test",
  "password": "correct-password"
}
```

**Expected outcome:**
- Status: `200 OK`
- Header: `Set-Cookie: awo_session=<token>; HttpOnly; SameSite=Lax`
- Body contains: `user_id`, `tenant_id`, non-empty `permissions` map
- Body does NOT contain the raw token (only cookie gets it)
- [ ] Pass

---

### AT2 — Login: wrong password

**Expected behavior:** Brute-force protection increments `failed_login_attempts`. The error message is generic — never reveals whether email or password is wrong.

**Request:**
```
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@tenant.test",
  "password": "wrong-password"
}
```

**Expected outcome:**
- Status: `401 Unauthorized`
- Body: `{"error": "invalid email or password"}` — exact wording, no oracle
- Body does NOT say "user not found" or "password incorrect"
- [ ] Pass

---

### AT3 — Login: account locked after 5 failures

**Expected behavior:** After ≥5 failed attempts the account is locked for a backoff period. Login is refused even with the correct password until the lockout expires.

**Setup:** Call AT2 five times for the same account first.

**Request:**
```
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@tenant.test",
  "password": "correct-password"
}
```

**Expected outcome:**
- Status: `423 Locked`
- Body: `{"error": "account locked, try again later"}`
- [ ] Pass

---

### AT4 — Access protected route with valid session cookie

**Expected behavior:** Cookie is present, session validates, permission is satisfied → handler runs.

**Setup:** First call AT1 to obtain the cookie.

**Request:**
```
GET /api/v1/finance/invoices
Cookie: awo_session=<token-from-AT1>
```

**Expected outcome:**
- Status: `200 OK`
- Body: invoice list (may be empty array, not 401/403)
- [ ] Pass

---

### AT5 — Access protected route without session

**Expected behavior:** No cookie and no Authorization header → `Authenticate` middleware rejects the request immediately; the route handler never runs.

**Request:**
```
GET /api/v1/finance/invoices
```
(no Cookie, no Authorization header)

**Expected outcome:**
- Status: `401 Unauthorized`
- Body: `{"error": "authentication required"}`
- [ ] Pass

---

### AT6 — Access protected route with valid session but missing permission

**Expected behavior:** Session is valid but the user does not have the specific permission for this route → `Authorize` middleware rejects with 403.

**Setup:** Login as a user who has no `finance.invoices.read` permission.

**Request:**
```
GET /api/v1/finance/invoices
Cookie: awo_session=<token-of-user-without-permission>
```

**Expected outcome:**
- Status: `403 Forbidden`
- Body: `{"error": "permission denied: finance.invoices.read"}` (or similar)
- [ ] Pass

---

### AT7 — Logout

**Expected behavior:** Cookie is cleared, session row is invalidated in DB, Redis cache key is deleted. Subsequent requests with the same token return 401.

**Setup:** First call AT1 to obtain a token.

**Request:**
```
POST /api/v1/auth/logout
Cookie: awo_session=<token-from-AT1>
```

**Expected outcome:**
- Status: `200 OK`
- Header: `Set-Cookie: awo_session=; Max-Age=0` (cookie cleared)
- [ ] Pass

**Follow-up — token is dead:**
```
GET /api/v1/finance/invoices
Cookie: awo_session=<same-token>
```
- Status: `401 Unauthorized`
- [ ] Pass

---

### AT8 — MFA: initiate setup

**Expected behavior:** Authenticated user calls initiate; gets back a TOTP secret and QR URI to scan with an authenticator app. Secret is NOT yet activated — user must call confirm next.

**Setup:** Login as a user who does not yet have MFA enabled. Use the resulting cookie.

**Request:**
```
POST /api/v1/auth/mfa/initiate
Cookie: awo_session=<token>
```

**Expected outcome:**
- Status: `200 OK`
- Body: `{"secret": "<base32>", "qr_uri": "otpauth://totp/..."}`
- MFA is still NOT enabled on the user record yet
- [ ] Pass

---

### AT9 — MFA: confirm setup

**Expected behavior:** User scans the QR code, generates a TOTP code, sends it. Server activates MFA on the user record. From now on, login requires a second step.

**Setup:** Run AT8 first; generate a valid TOTP code from the returned secret.

**Request:**
```
POST /api/v1/auth/mfa/confirm
Cookie: awo_session=<token>
Content-Type: application/json

{
  "code": "123456"
}
```

**Expected outcome:**
- Status: `200 OK`
- `users.mfa_enabled = TRUE` in DB
- [ ] Pass

---

### AT10 — MFA: login flow (two steps)

**Expected behavior:** After MFA is enabled, login returns 202 with a `pending_token` instead of 200. User must exchange that token + TOTP code at `/auth/mfa/complete` to get a full session.

**Step 1 — credentials:**
```
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "mfa-user@tenant.test",
  "password": "correct-password"
}
```
**Expected outcome of Step 1:**
- Status: `202 Accepted`
- Body: `{"mfa_required": true, "pending_token": "<token>"}`
- No session cookie set yet
- [ ] Pass

**Step 2 — TOTP exchange (public endpoint — no cookie needed):**
```
POST /api/v1/auth/mfa/complete
Content-Type: application/json

{
  "pending_token": "<token-from-step-1>",
  "code": "123456"
}
```
**Expected outcome of Step 2:**
- Status: `200 OK`
- Header: `Set-Cookie: awo_session=<token>; HttpOnly; SameSite=Lax`
- Body: full `ResolvedSession` JSON
- [ ] Pass

---

### AT11 — MFA: wrong TOTP code

**Expected behavior:** Invalid code at `/mfa/complete` is rejected. Pending token is consumed (single-use) regardless.

**Request:**
```
POST /api/v1/auth/mfa/complete
Content-Type: application/json

{
  "pending_token": "<valid-pending-token>",
  "code": "000000"
}
```

**Expected outcome:**
- Status: `401 Unauthorized`
- Body: `{"error": "invalid MFA code"}`
- Pending token is now gone (re-using same pending_token returns the same 401 or 404)
- [ ] Pass

---

### AT12 — MFA: disable

**Expected behavior:** Authenticated user re-provides their password and disables MFA. Login no longer requires TOTP.

**Request:**
```
DELETE /api/v1/auth/mfa
Cookie: awo_session=<token>
Content-Type: application/json

{
  "password": "correct-password"
}
```

**Expected outcome:**
- Status: `200 OK`
- `users.mfa_enabled = FALSE`, `users.mfa_secret = NULL` in DB
- [ ] Pass

---

### AT13 — Password reset: forgot password

**Expected behavior:** Always returns 200 regardless of whether the email exists — prevents user enumeration. If email exists, a reset token is generated and would normally be emailed (currently logged or returned for testing).

**Request:**
```
POST /api/v1/auth/forgot-password
Content-Type: application/json

{
  "email": "admin@tenant.test"
}
```

**Expected outcome:**
- Status: `200 OK`
- Body: `{"message": "if that email exists, a reset link has been sent"}`
- Status is also `200` if email does not exist (identical response)
- [ ] Pass

---

### AT14 — Password reset: set new password

**Expected behavior:** Valid token sets a new password. Token is marked used. All existing sessions are invalidated.

**Setup:** Obtain a raw reset token from AT13 (check server logs or DB `password_reset_tokens.token_hash` lookup).

**Request:**
```
POST /api/v1/auth/reset-password
Content-Type: application/json

{
  "token": "<raw-token>",
  "password": "NewStr0ng!Password"
}
```

**Expected outcome:**
- Status: `200 OK`
- `password_reset_tokens.used_at` is now set in DB
- All `user_sessions` rows for that user are invalidated
- Old session cookie returns 401
- [ ] Pass

---

### AT15 — Password reset: token reuse rejected

**Expected behavior:** Reset tokens are single-use. Second call with the same token returns 410 Gone.

**Setup:** Run AT14 first (token is now used).

**Request:**
```
POST /api/v1/auth/reset-password
Content-Type: application/json

{
  "token": "<same-token>",
  "password": "AnotherStr0ng!Pass"
}
```

**Expected outcome:**
- Status: `410 Gone`
- Body: `{"error": "reset token already used"}`
- [ ] Pass

---

### AT16 — API key: create

**Expected behavior:** Authenticated user creates an API key. Raw key is returned once. Subsequent reads show the key name and scopes but never the raw value.

**Request:**
```
POST /api/v1/auth/api-keys
Cookie: awo_session=<admin-token>
Content-Type: application/json

{
  "name": "accounting-sync",
  "scopes": ["finance.invoices.read", "finance.transactions.read"]
}
```

**Expected outcome:**
- Status: `201 Created`
- Body: `{"id": "...", "name": "accounting-sync", "key": "eak_...", "scopes": [...]}`
- `key` field starts with `eak_` — returned only in this response
- DB `api_keys.key_hash` is SHA-256 of the raw key (never plaintext)
- [ ] Pass

---

### AT17 — API key: authenticate a request

**Expected behavior:** API key passed as Bearer token resolves to a minimal `ResolvedSession` scoped to the key's permissions. No session cookie involved.

**Setup:** Use the `eak_...` key from AT16.

**Request:**
```
GET /api/v1/finance/invoices
Authorization: Bearer eak_<key-from-AT16>
```

**Expected outcome:**
- Status: `200 OK`
- Request is served (key has `finance.invoices.read` scope)
- [ ] Pass

---

### AT18 — API key: scope ceiling enforced

**Expected behavior:** Even if the owning user has broad permissions, the API key can only exercise the scopes it was created with. A request for a route requiring `finance.transactions.create` should fail.

**Request:**
```
POST /api/v1/finance/transactions
Authorization: Bearer eak_<key-with-only-read-scopes>
Content-Type: application/json

{ ... }
```

**Expected outcome:**
- Status: `403 Forbidden`
- Body: `{"error": "permission denied: finance.transactions.create"}`
- [ ] Pass

---

### AT19 — API key: revoke

**Expected behavior:** Revoked key is immediately rejected. Cache eviction happens on revoke; subsequent requests with the key return 401.

**Request:**
```
DELETE /api/v1/auth/api-keys/<key-id>
Cookie: awo_session=<admin-token>
```

**Expected outcome:**
- Status: `200 OK`
- Subsequent call with `Authorization: Bearer eak_<revoked-key>` → `401 Unauthorized`
- [ ] Pass

---

### AT20 — Session: cross-tenant isolation

**Expected behavior:** A valid session for tenant A cannot access tenant B's data. RLS at DB level and tenant context injection in middleware enforce this.

**Setup:** Two separate tenants, each with their own admin user and session.

**Request (tenant A session trying tenant B path — if route accepts tenant in path):**
```
GET /api/v1/finance/invoices
Cookie: awo_session=<tenant-A-token>
X-Tenant-ID: <tenant-B-uuid>   (if your routes accept this header)
```

**Expected outcome:**
- Status: `403 Forbidden` or returns empty data scoped to tenant A only
- Never returns tenant B records
- [ ] Pass

---

### AT21 — Boot schema: session-gated navigation

**Expected behavior:** `GET /schema/boot` returns the AMIS navigation schema. Modules the user cannot see (flag disabled or no read permission) are absent from the response. This is what drives the frontend — no session data is fetched client-side for permissions.

**Request:**
```
GET /api/v1/schema/boot
Cookie: awo_session=<token>
```

**Expected outcome:**
- Status: `200 OK`
- Body: AMIS `app` schema JSON with `pages` array
- `pages` only includes modules where `sess.FeatureEnabled(slug)` AND `sess.Can(module+".read")`
- [ ] Pass

---

## Updated Completion Summary

| Phase | Description | Status |
|---|---|---|
| 17 | UserService implementation (`service/user.go`) | [ ] Todo |
| 18 | HTTP middleware rewrite (Authenticate, Authorize, RequireFlag, wire routes) | [ ] Todo |
| 19 | DB migrations + SQLC verification (SSO, API keys) + `go build` passes | [ ] Todo |
| 20 | API tests AT1–AT21 (behavioral verification) | [ ] Todo |
| 5 | HTTP middleware (Authenticate, Authorize) + Login/Logout handlers | ✅ Done |
| 6 | Security hardening (SHA-256 tokens, tenant guard, RLS fix, AssignableTo) | ✅ Done |
| 7 | Wire flags + settings into session at login | ✅ Done |
| 8 | Entity module (table, repo, scope resolution at login) | ✅ Done |
| 9 | MFA / TOTP | ✅ Done |
| 10 | Password management (reset flow, strength validation, history) | ✅ Done |
| 11 | OAuth / OIDC / SAML SSO | ✅ Done |
| 12 | API key authentication | ✅ Done |
| 13 | MRA registry + BootService | ✅ Done |
| 14 | Audit trail | ✅ Done (analytics stubs pending) |
| 15 | Docs corrections | ✅ Done |
| 16 | End-to-end integration tests (mock-based) | ✅ Done |
