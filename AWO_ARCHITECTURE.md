# AWO ERP — Architecture Reference

> Auto-generated from codebase analysis on 2026-03-29.
> Module path: `awo.so`

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Directory Structure](#2-directory-structure)
3. [Technology Stack](#3-technology-stack)
4. [Multi-Tenancy Architecture](#4-multi-tenancy-architecture)
5. [Dependency Injection (Wire)](#5-dependency-injection-wire)
6. [Temporal Workflows](#6-temporal-workflows)
7. [Database Layer](#7-database-layer)
8. [API Layer](#8-api-layer)
9. [ABAC/RBAC Authorization](#9-abacrbac-authorization)
10. [Configuration](#10-configuration)
11. [Naming Conventions](#11-naming-conventions)
12. [Testing](#12-testing)
13. [Development Workflow](#13-development-workflow)
14. [Other Modules](#14-other-modules)
15. [Open Questions / Gaps](#15-open-questions--gaps)

---

## 1. Project Overview

**AWO** is a multi-tenant Enterprise Resource Planning (ERP) system exposing a REST API built with Go + Fiber.

**What it solves:**
- Identity & Access Management (IAM) — users, sessions, MFA, SSO, API keys, RBAC/ABAC
- Tenant lifecycle management — create, activate, suspend, archive tenants with full RLS isolation
- Finance module — chart of accounts, journal transactions, trial-balance reporting
- Organizational entity hierarchy — ltree-based subtree queries for scoped data access
- Audit logging — immutable event trail with permission-gated read access
- Observable infrastructure — OpenTelemetry tracing, Prometheus metrics, structured logging

**Application type:** Fiber HTTP server with an embedded AMIS-based schema-driven web UI (`/ui/*`).

**Module path:** `awo.so` (declared in `go.mod`)

---

## 2. Directory Structure

```
/project/erp/
├── cmd/                        # CLI entry point(s) — cobra commands
├── db/
│   ├── migration/              # SQL migration files (golang-migrate format)
│   ├── queries/                # SQLC source .sql query files
│   └── sqlc/                   # SQLC-generated Go code (db.Store, models, tx helpers)
├── docs/                       # MkDocs documentation source
│   └── reference/modules/      # Module-level task docs (e.g. iam/tasks.md)
├── internal/
│   ├── api/
│   │   ├── handlers/           # HTTP handlers, grouped by domain
│   │   │   ├── auth/           # auth.go (login/logout/password-reset), mfa.go, oauth.go, apikey.go
│   │   │   ├── audit/          # handler.go — ListAuditEventsHandler
│   │   │   ├── entity/         # handler.go — entity CRUD
│   │   │   ├── finance/        # handler.go — accounts, transactions, reports
│   │   │   ├── health/         # handler.go — liveness check
│   │   │   ├── schema/         # handler.go — AMIS boot schema
│   │   │   ├── tenant/         # handler.go — tenant CRUD + lifecycle
│   │   │   ├── ui/             # handler.go — server-side rendered UI pages
│   │   │   ├── user/           # handler.go — user CRUD + password change
│   │   │   └── routes.go       # Central route registration + Dependencies struct
│   │   └── middleware/
│   │       ├── auth.go         # authenticateMiddleware, Authorize()
│   │       ├── observability.go# request tracing + structured logging
│   │       ├── security.go     # CORS, CSRF, security headers (RouteSecurityManager)
│   │       ├── tenant.go       # TenantMiddleware — extracts + validates tenant, injects RLS
│   │       └── whitelist.go    # EndpointWhitelist — public route bypass list
│   ├── core/                   # Business logic (bounded contexts)
│   │   ├── audit/              # audit.Service + audit repository
│   │   ├── entity/             # Organizational entity service (ltree)
│   │   ├── finance/
│   │   │   └── service/        # account.go, transaction.go, reporting.go
│   │   ├── iam/                # Identity & Access Management
│   │   │   ├── domain/         # Value objects: user, session, authorization, SSO, MFA, errors
│   │   │   ├── repository/     # Ports: identity.go, session.go, auth.go (interfaces + sqlc impls)
│   │   │   ├── service/        # user.go, session.go, authz.go, sso.go, apikey.go
│   │   │   └── iam.go          # Facade — re-exports domain types + service interfaces
│   │   └── tenant/
│   │       ├── domain/         # tenant.go, status.go, errors.go
│   │       ├── repository/     # repository.go (interface + sqlc implementation)
│   │       ├── service.go      # Business logic (lifecycle transitions, validation)
│   │       └── tenant.go       # Facade
│   ├── platform/
│   │   ├── cache/              # Redis-backed cache.Service + context keys
│   │   ├── config/             # config.go — Viper-based config loading
│   │   └── temporal/           # Temporal client + worker bootstrap
│   └── shared/
│       ├── context.go          # Context key constants: TenantIDKey, UserIDKey, SessionKey, …
│       ├── encryption/         # Symmetric encryption service
│       ├── errors/
│       │   ├── http.go         # ToHTTPError() using errors.As unwrapping
│       │   └── types.go        # BusinessError, RepositoryError, HTTPError structs
│       ├── logger/             # Logger interface + Zerolog / Zap / Slog adapters
│       ├── metrics/            # MetricsProvider interface + Prometheus implementation
│       └── tracing/            # tracing.Service interface + OpenTelemetry implementation
├── scripts/
│   ├── generate-wire.sh        # Wire codegen helper (called by `make wire`)
│   └── seed.sh                 # Bootstrap: tenant + admin user via REST API
├── test/
│   ├── integration/            # Integration tests (real DB via docker-compose.test.yml)
│   └── e2e/                    # End-to-end tests
├── web/                        # Frontend (AMIS-based)
│   ├── pages/index.html        # Sidebar layout shell
│   ├── schemas/pages/          # AMIS JSON page schemas
│   ├── sdk/                    # AMIS sdk.js, sdk.css, charts bundle
│   ├── static/                 # Static assets
│   └── utils/                  # Frontend utilities
├── AWO_ARCHITECTURE.md         # This document
├── CLAUDE.md                   # AI assistant project instructions
├── Makefile                    # Build / dev workflow
├── docker-compose.test.yml     # Test infrastructure (Postgres, Redis)
├── go.mod
├── go.sum
└── sqlc.yaml                   # SQLC configuration
```

---

## 3. Technology Stack

| Dependency | Version | Role |
|---|---|---|
| `github.com/gofiber/fiber/v2` | v2.52.12 | HTTP server framework |
| `github.com/jackc/pgx/v5` | v5.8.0 | PostgreSQL driver + connection pool |
| `github.com/jackc/pgconn` | v1.14.3 | Low-level PG connection (used by pgx) |
| `sqlc` (via Makefile) | — | Type-safe SQL codegen from `.sql` files |
| `github.com/go-redis/redis/v8` | v8.11.5 | Redis client (session cache, feature-flag cache) |
| `github.com/alicebob/miniredis/v2` | v2.37.0 | In-memory Redis mock for tests |
| `github.com/patrickmn/go-cache` | v2.1.0 | In-process memory cache |
| `github.com/casbin/casbin/v2` | v2.135.0 | RBAC/ABAC policy engine |
| `github.com/google/cel-go` | v0.27.0 | Common Expression Language (dynamic policies) |
| `github.com/expr-lang/expr` | v1.17.8 | Dynamic expression evaluation |
| `go.temporal.io/sdk` | v1.41.1 | Temporal workflow SDK |
| `go.temporal.io/api` | v1.62.5 | Temporal protobuf API types |
| `go.opentelemetry.io/otel` | v1.42.0 | OpenTelemetry tracing core |
| `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc` | — | OTLP gRPC trace exporter |
| `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp` | — | OTLP HTTP trace exporter |
| `github.com/exaring/otelpgx` | v0.10.0 | PgX ↔ OpenTelemetry bridge |
| `github.com/prometheus/client_golang` | v1.23.2 | Prometheus metrics exporter |
| `github.com/rs/zerolog` | v1.34.0 | Structured JSON logging |
| `go.uber.org/zap` | v1.27.1 | High-performance logging (alternative adapter) |
| `github.com/go-playground/validator/v10` | v10.30.1 | Struct-tag validation |
| `github.com/google/uuid` | v1.6.0 | UUID generation |
| `github.com/shopspring/decimal` | v1.4.0 | Precision decimal arithmetic (finance) |
| `github.com/gosimple/slug` | v1.15.0 | URL-safe slug generation |
| `github.com/google/wire` | v0.7.0 | Compile-time dependency injection codegen |
| `github.com/spf13/viper` | v1.21.0 | Config from YAML + env vars |
| `github.com/spf13/cobra` | v1.10.2 | CLI subcommand framework |
| `github.com/gorilla/websocket` | v1.5.3 | WebSocket protocol |
| `golang.org/x/crypto` | v0.49.0 | Password hashing (bcrypt), TOTP |
| `github.com/a-h/templ` | v0.3.1001 | Go HTML templating for UI pages |
| `github.com/stretchr/testify` | v1.11.1 | Test assertions + mocking |
| `go.uber.org/mock` | v0.6.0 | Mockgen-based interface mocking |
| `github.com/dlclark/regexp2` | v1.11.5 | Advanced regex (PCRE-compatible) |
| `gopkg.in/yaml.v3` | v3.0.1 | YAML parsing |

---

## 4. Multi-Tenancy Architecture

### Isolation model

PostgreSQL Row-Level Security (RLS) is the enforcement layer. Every table that stores tenant data has an RLS policy keyed to a PostgreSQL session variable.

**Session variable:** `app.current_tenant_id`

**RLS policy pattern (example):**
```sql
CREATE POLICY tenant_isolation ON users
  USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

### Request lifecycle

```
HTTP Request
│
├─ TenantMiddleware  (internal/api/middleware/tenant.go)
│   ├─ 1. Extract tenant identity (priority order):
│   │     a. Header: X-Tenant-ID
│   │     b. Query param: ?tenant_id=
│   │     c. Subdomain: tenant1.domain.com → "tenant1"
│   │        bo.tenant1.domain.com → "tenant1"   (back-office prefix stripped)
│   ├─ 2. Validate tenant via TenantService (cache-first, TTL ~5 min)
│   ├─ 3. Assert tenant.Status == ACTIVE
│   ├─ 4. Store in Fiber locals:    c.Locals(shared.TenantIDKey, tenantID)
│   ├─ 5. Store in Go context:      shared.WithTenantID(ctx, tenantID)
│   └─ 6. Set DB session variable:  store.SetTenantContextFromCtx(ctx)
│         └─ executes: SET LOCAL app.current_tenant_id = '<uuid>'
│
├─ authenticateMiddleware  (internal/api/middleware/auth.go)
│   ├─ Reads session cookie / Bearer token
│   ├─ Validates session via SessionService / APIKeyService
│   └─ Stores Session in context
│
├─ Authorize("permission")  (inline fiber.Handler)
│   └─ Checks session.Permissions contains required permission
│
└─ Handler
    └─ Repository methods receive ctx — RLS enforced automatically by Postgres
```

### Context helpers (`internal/shared/context.go`)

```go
// Writing
shared.WithTenantID(ctx, tenantID uuid.UUID) context.Context
shared.WithUserID(ctx, userID uuid.UUID) context.Context
shared.WithSession(ctx, sess *Session) context.Context

// Reading
shared.GetTenantID(ctx) (uuid.UUID, bool)
shared.GetUserID(ctx) (uuid.UUID, bool)
shared.GetSession(ctx) (*Session, bool)
```

### Tenant status lifecycle

```
PENDING ──► ACTIVE ──► SUSPENDED ──► ACTIVE
   │            │
   └────────────┴──► ARCHIVED
```

Transitions enforced in `internal/core/tenant/service.go`.

### Public endpoint bypass

`internal/api/middleware/whitelist.go` — `EndpointWhitelist` config holds exact paths and patterns that skip tenant resolution (e.g., `POST /api/v1/tenants` for tenant creation, `GET /health/`).

---

## 5. Dependency Injection (Wire)

**Framework:** `github.com/google/wire` v0.7.0 (compile-time codegen)

**Generation:** `./scripts/generate-wire.sh` called by `make wire`.
The generated `wire_gen.go` is not committed; regenerate with `make wire`.

### Top-level `Dependencies` struct

Defined in `internal/api/handlers/routes.go`:

```go
type Dependencies struct {
    Logger           logger.Logger
    Metrics          metrics.MetricsProvider
    Tracer           tracing.Service
    TenantService    coreTenant.Service
    UserService      iam.UserService
    FinanceServices  *financeService.Services
    TenantMiddleware fiber.Handler
    SecurityManager  *middlewarePkg.RouteSecurityManager
    SessionService   iam.SessionService
    AuthConfig       *middlewarePkg.AuthConfig
    SSOService       iam.SSOService
    APIKeyService    iam.APIKeyService
    Store            db.Store
    AuditService     audit.Service
}
```

`Dependencies.Validate()` asserts Logger, Metrics, Tracer are non-nil at startup.

### Provider set groups (inferred)

| Set name | Contents |
|---|---|
| ObservabilitySet | Logger, MetricsProvider, tracing.Service |
| DatabaseSet | `db.Store` (SQLC-generated pool wrapper) |
| CacheSet | `cache.Service` (Redis) |
| IAMRepositorySet | UserRepository, SessionRepository, AuthzRepository, SSORepository, APIKeyRepository |
| IAMServiceSet | UserService, SessionService, SSOService, APIKeyService |
| TenantSet | TenantRepository, TenantService |
| AuditSet | AuditRepository, AuditService |
| FinanceSet | AccountService, TransactionService, ReportingService wrapped in `*Services` |
| MiddlewareSet | TenantMiddleware, SecurityManager, AuthConfig |

---

## 6. Temporal Workflows

**SDK:** `go.temporal.io/sdk` v1.41.1

**Config structure** (in `internal/platform/config/config.go`):
```go
type TemporalConfig struct { /* fields bound from TEMPORAL_* env vars */ }
```
Defaults set by `SetTemporalDefaults(v *viper.Viper)`.
Env vars bound by `BindTemporalEnvVars(v *viper.Viper)`.

**Bootstrap:** `internal/platform/temporal/` — Temporal client initialization and worker registration.

**Status:** The SDK and config infrastructure are wired in, but no concrete workflow/activity definitions were found in the current codebase snapshot. The integration is scaffolded for future async operations (e.g., onboarding emails, password-reset delivery, bulk data jobs).

---

## 7. Database Layer

### SQLC

**Config file:** `sqlc.yaml` (repo root)

```yaml
version: "2"
sql:
  - schema: db/migration
    queries: db/queries
    engine: postgresql
    gen:
      go:
        out: db/sqlc
        emit_json_tags: true
        emit_interface: true          # generates db.Querier / db.Store
        emit_pointers_for_null_types: true
        # type overrides: uuid.UUID, time.Time, decimal.Decimal
```

Generated output lives in `db/sqlc/` — do not edit manually.
Regenerate with: `make sqlc`

### Migrations

| Detail | Value |
|---|---|
| Location | `db/migration/` |
| Format | golang-migrate (`.up.sql` / `.down.sql` per version) |
| Naming | `<6-digit-seq>_<description>.sql` (e.g., `001001_init_schema.sql`) |
| Notable | `001002` — adds UPPERCASE constraint on `company_size` |

### Repository pattern

```
internal/core/<domain>/repository/
├── <noun>.go        — Port interface (e.g. UserRepository, SessionRepository)
└── <noun>_sqlc.go   — SQLC adapter implementing the interface
```

**Interface location:** `internal/core/<domain>/repository/<noun>.go`
**Implementation location:** same package, `<noun>_sqlc.go` (or named after the adapter)

### Transaction / tenant context helper

All repository methods receive `context.Context` from which `shared.GetTenantID(ctx)` extracts the tenant UUID.

Pattern used in `internal/core/audit/repository.go`:
```go
tenantID, ok := shared.GetTenantID(ctx)
if !ok {
    return nil, fmt.Errorf("tenantID not set in context")
}
err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
    // all queries in here run within a transaction with SET LOCAL app.current_tenant_id
    return s.SomeQuery(ctx, params)
})
```

`db.Store.WithTenant()` wraps a callback in a transaction and sets the PG session variable before executing.

---

## 8. API Layer

**Framework:** Fiber v2 (`github.com/gofiber/fiber/v2`)
**Route file:** `internal/api/handlers/routes.go` — `RegisterRoutes(app *fiber.App, deps Dependencies)`

### Middleware chain (global)

1. **Observability** (`middleware/observability.go`) — injects trace span, logs request/response, records Prometheus metrics
2. **Security** (`middleware/security.go` — `RouteSecurityManager`) — CORS, CSRF, security headers; configured per route group
3. **Tenant** (`middleware/tenant.go`) — applied to all `/api/v1/` groups except public auth routes
4. **Authenticate** (`middleware/auth.go`) — session cookie or `Authorization: Bearer <api-key>` validation
5. **Authorize** (inline `fiber.Handler`) — permission string check against `session.Permissions`

### Route table

#### Auth (public + session-gated)

| Method | Path | Handler | Auth |
|---|---|---|---|
| `POST` | `/api/v1/auth/login` | `LoginHandler` | — |
| `POST` | `/api/v1/auth/logout` | `LogoutHandler` | session |
| `POST` | `/api/v1/auth/forgot-password` | `ForgotPasswordHandler` | — |
| `POST` | `/api/v1/auth/reset-password` | `ResetPasswordHandler` | — |
| `POST` | `/api/v1/auth/mfa/complete` | `MFACompleteHandler` | — (pending token) |
| `POST` | `/api/v1/auth/mfa/initiate` | `MFAInitiateHandler` | session |
| `POST` | `/api/v1/auth/mfa/confirm` | `MFAConfirmHandler` | session |
| `DELETE` | `/api/v1/auth/mfa/` | `MFADisableHandler` | session |
| `GET` | `/api/v1/auth/oauth/:provider` | `OAuthBeginHandler` | — |
| `GET` | `/api/v1/auth/oauth/:provider/callback` | `OAuthCallbackHandler` | — |
| `POST` | `/api/v1/auth/api-keys/` | `CreateAPIKeyHandler` | session |
| `GET` | `/api/v1/auth/api-keys/` | `ListAPIKeysHandler` | session |
| `DELETE` | `/api/v1/auth/api-keys/:id` | `RevokeAPIKeyHandler` | session |

#### Tenants

| Method | Path | Handler | Auth |
|---|---|---|---|
| `GET` | `/api/v1/tenants` | `ListTenantsHandler` | — |
| `POST` | `/api/v1/tenants` | `CreateTenantHandler` | — |
| `GET` | `/api/v1/tenants/:id` | `GetTenantHandler` | — |
| `PUT` | `/api/v1/tenants/:id` | `UpdateTenantHandler` | — |
| `PATCH` | `/api/v1/tenants/:id` | `PatchTenantHandler` | — |
| `DELETE` | `/api/v1/tenants/:id` | `DeleteTenantHandler` | — |
| `POST` | `/api/v1/tenants/:id/activate` | `ActivateTenantHandler` | — |
| `POST` | `/api/v1/tenants/:id/suspend` | `SuspendTenantHandler` | — |
| `POST` | `/api/v1/tenants/:id/archive` | `ArchiveTenantHandler` | — |
| `POST` | `/api/v1/tenants/onboard` | `OnboardTenantHandler` | — |

#### Users

| Method | Path | Handler | Auth |
|---|---|---|---|
| `GET` | `/api/v1/users` | `ListUsersHandler` | session |
| `POST` | `/api/v1/users` | `CreateUserHandler` | — |
| `GET` | `/api/v1/users/:id` | `GetUserHandler` | session |
| `PUT` | `/api/v1/users/:id` | `UpdateUserHandler` | session |
| `DELETE` | `/api/v1/users/:id` | `DeleteUserHandler` | session |
| `POST` | `/api/v1/users/:id/change-password` | `ChangePasswordHandler` | session |

#### Finance

| Method | Path | Handler | Permission |
|---|---|---|---|
| `POST` | `/api/v1/finance/accounts/` | `CreateAccountHandler` | `finance.accounts.create` |
| `GET` | `/api/v1/finance/accounts/` | `ListAccountsHandler` | `finance.accounts.read` |
| `GET` | `/api/v1/finance/accounts/:id` | `GetAccountHandler` | `finance.accounts.read` |
| `PUT` | `/api/v1/finance/accounts/:id` | `UpdateAccountHandler` | `finance.accounts.update` |
| `DELETE` | `/api/v1/finance/accounts/:id` | `DeleteAccountHandler` | `finance.accounts.delete` |
| `GET` | `/api/v1/finance/accounts/:id/balance` | `GetAccountBalanceHandler` | `finance.accounts.read` |
| `POST` | `/api/v1/finance/transactions/` | `CreateTransactionHandler` | `finance.transactions.create` |
| `GET` | `/api/v1/finance/transactions/` | `ListTransactionsHandler` | `finance.transactions.read` |
| `GET` | `/api/v1/finance/transactions/:id` | `GetTransactionHandler` | `finance.transactions.read` |
| `GET` | `/api/v1/finance/reports/trial-balance` | `TrialBalanceHandler` | `finance.reports.read` |

#### Other

| Method | Path | Handler | Auth |
|---|---|---|---|
| `POST` | `/api/v1/entities/` | `CreateEntityHandler` | session |
| `GET` | `/api/v1/entities/` | `ListEntitiesHandler` | session |
| `GET` | `/api/v1/schema/boot` | `BootSchemaHandler` | session |
| `GET` | `/api/v1/audit-logs` | `ListAuditEventsHandler` | `iam.sessions.read` |
| `GET` | `/health/` | `HealthHandler` | — |
| `GET` | `/metrics` | Prometheus handler | — |
| `GET` | `/ui/login` | `LoginPageHandler` | — |
| `GET` | `/ui/` | redirect → `/ui/demo` | — |
| `GET` | `/ui/demo` | `DemoPageHandler` | — |
| `GET` | `/ui/demo/components` | component showcase | — |

### Authentication flow

**Login (`POST /api/v1/auth/login`):**
1. Validate credentials → hash compare
2. If `user.MFAEnabled` → return `202 { mfa_required: true, pending_token: "..." }`
3. Else → create session row, set `Set-Cookie: <session_id>`

**MFA flow:**
1. `POST /api/v1/auth/mfa/complete` — exchange `pending_token` + TOTP code for full session + cookie
2. `POST /api/v1/auth/mfa/initiate` (authenticated) — returns `{ qr_uri, secret }`
3. `POST /api/v1/auth/mfa/confirm` (authenticated) — activates MFA after verifying first valid code
4. `DELETE /api/v1/auth/mfa/` — disable MFA (requires password in body)

**SSO flow:**
1. `GET /api/v1/auth/oauth/:provider` → redirect to IdP
2. `GET /api/v1/auth/oauth/:provider/callback` → exchange code, create/link user, issue session
3. MFA intentionally skipped for SSO (IdP acts as second factor)

---

## 9. ABAC/RBAC Authorization

**Engine:** `github.com/casbin/casbin/v2` v2.135.0

**Model:** Casbin CONF-string defined in `internal/core/iam/domain/authorization.go` as `domain.CasbinModel`.

### Actor types

```go
// internal/core/iam/domain/authorization.go
type ActorType string

const (
    ActorPlatform ActorType = "platform"  // system-wide admin
    ActorTenant   ActorType = "tenant"    // tenant-level users
    ActorPortal   ActorType = "portal"    // external/customer users
    ActorAPI      ActorType = "api"       // API clients (API key auth)
)
```

**Default domain:** `domain.DomainPlatform`

### Entity-scoped visibility

Sessions embed an `EntityScope` that drives repository-level WHERE clauses:

```go
type EntityScopeType string

const (
    EntityScopeAll      EntityScopeType = "all"      // unrestricted (admins)
    EntityScopeSubtree  EntityScopeType = "subtree"  // home entity + descendants (managers)
    EntityScopeEntity   EntityScopeType = "entity"   // own entity only (staff)
)
```

Repository filtering (conceptual):
```go
switch sess.EntityScope.Type {
case domain.EntityScopeAll:
    // RLS already handles tenant; no extra WHERE needed
case domain.EntityScopeSubtree:
    query = query.Where("entity_path <@ $1", sess.EntityScope.PathPrefix)
case domain.EntityScopeEntity:
    query = query.Where("entity_id = $1", sess.EntityScope.EntityID)
}
```

### Permission evaluation

Route-level:
```go
// inline in routes.go
app.Get("/api/v1/finance/accounts/", Authorize("finance.accounts.read"), handler)
```

`Authorize(perm string) fiber.Handler` (in `middleware/auth.go`):
```go
func Authorize(permission string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        sess := shared.GetSession(c.UserContext())
        if !slices.Contains(sess.Permissions, permission) {
            return fiber.ErrForbidden
        }
        return c.Next()
    }
}
```

### Session-embedded authorization snapshot

At login the session row is hydrated with:
- `Permissions []string` — flat list of allowed permission strings
- `EntityScope EntityScope` — visibility scope
- `Configuration.Flags map[string]bool` — resolved feature flags
- `Configuration.Settings map[string]any` — tenant settings (e.g. `iam.session_ttl_hours`)
- `Configuration.Prefs map[string]any` — user preferences (e.g. `ui.theme`)

This avoids per-request policy lookups after login.

---

## 10. Configuration

**File:** `internal/platform/config/config.go`

**Entry points:**
- `config.Load() *Config` — load and validate
- `config.LoadWithViper() (*Config, *viper.Viper)` — also return raw Viper instance

**Priority (highest → lowest):**
1. Environment variables (via `v.BindEnv()`)
2. `.env` file (loaded by `loadDotEnvFile()`)
3. `config.yaml` searched in: `.`, `./config`, `/etc/myapp`
4. Viper defaults

### Config struct

```go
type Config struct {
    App       AppConfig
    Server    ServerConfig
    Database  DatabaseConfig
    Migration MigrationConfig
    Redis     RedisConfig
    Temporal  TemporalConfig
    Auth      AuthConfig
    Features  FeatureConfig
    Logger    LoggerConfig
    Tracing   TracingConfig
    UI        UIConfig
}
```

### Key environment variables

| Env var | Config field | Default |
|---|---|---|
| `DB_HOST` | `Database.Host` | `localhost` |
| `DB_PORT` | `Database.Port` | `5432` |
| `DB_USER` | `Database.User` | — |
| `DB_PASSWORD` | `Database.Password` | — |
| `DB_NAME` | `Database.Name` | — |
| `DB_SSL_MODE` | `Database.SSLMode` | `disable` |
| `REDIS_HOST` | `Redis.Host` | `localhost` |
| `REDIS_PORT` | `Redis.Port` | `6379` |
| `REDIS_PASSWORD` | `Redis.Password` | — |
| `AUTH_SESSION_TTL` | `Auth.SessionTTL` | — |
| `AUTH_COOKIE_NAME` | `Auth.CookieName` | — |
| `LOG_LEVEL` | `Logger.Level` | `info` |
| `LOG_FORMAT` | `Logger.Format` | `json` |
| `TRACING_ENABLED` | `Tracing.Enabled` | `false` |
| `TRACING_EXPORTER` | `Tracing.Exporter` | `stdout` |
| `OTEL_ENDPOINT` | `Tracing.Endpoint` | — |
| `TEMPORAL_*` | `Temporal.*` | — |

**Validation:** `Config.Validate()` panics on missing required fields (DB host, port ranges, etc.).

---

## 11. Naming Conventions

### Files

| Pattern | Usage |
|---|---|
| `domain/*.go` | Value objects, domain errors, constants |
| `repository/*.go` | Port interfaces |
| `repository/*_sqlc.go` | SQLC adapter implementations |
| `service/*.go` or `service.go` | Business logic |
| `handler.go` / `auth.go` / `mfa.go` | HTTP handlers |
| `middleware/*.go` | Named by concern (`tenant.go`, `auth.go`, `whitelist.go`) |
| `*_test.go` | Tests alongside source |
| `*_mock.go` | Generated mocks (mockgen) |

### Interfaces

Suffixed by role: `UserService`, `UserRepository`, `SessionService`, `AuditRepository`, `cache.Service`, `logger.Logger`, `tracing.Service`, `metrics.MetricsProvider`.
No `I` prefix convention.

### Constructors

`New<Type>(deps...) (*Type, error)` for primary constructors.
Functional options: `With<Param>(val) func(*Type) error`.

### Errors

```go
// Domain sentinels (domain/errors.go)
var ErrTenantNotFound = errors.New("tenant not found")
var ErrTenantAlreadyExists = errors.New("tenant already exists")

// Business errors with HTTP status
type BusinessError struct {
    Code     string
    Message  string
    Category Category
    Severity Severity
    Details  map[string]any
}

// Code constants
const CodeTenantExists = "TENANT_EXISTS"

// Category constants
const CategoryTenant    = "tenant"
const CategorySecurity  = "security"
const CategoryValidation = "validation"
```

**Error propagation:**
- Repo layer: `parseTenantDBError(err, op)` → domain sentinel or `*BusinessError`
- Handler layer: `mapTenantError(err)` → `*BusinessError` with HTTP status
- HTTP layer: `shared/errors/http.go:ToHTTPError(err)` using `errors.As` unwrapping

### Constants/Enums

`Status<Name>` for lifecycle states, `Actor<Name>` for actor types, `EntityScope<Name>` for scope types, `OAuthProvider<Name>` for providers.

---

## 12. Testing

### Unit tests

Collocated with source (`*_test.go` in same package).
Framework: `github.com/stretchr/testify` (`assert`, `require`).
Mocks generated by `go.uber.org/mock/mockgen` via `//go:generate` directives.

### Integration tests

Location: `test/integration/`
Infrastructure: `docker-compose.test.yml` (real Postgres + Redis).
Timeout: 8 minutes.

### E2E tests

Location: `test/e2e/`
Timeout: 15 minutes.

### Redis mocking

`github.com/alicebob/miniredis/v2` used for unit tests that touch the cache layer.

### Makefile test targets

```makefile
make test            # unit tests, timeout 5m
make test-integration # integration tests, timeout 8m
make test-e2e        # e2e tests, timeout 15m
```

### Shell bootstrap / smoke test

`scripts/seed.sh` — creates a platform tenant + admin user via HTTP. Acts as a smoke test for the full API stack.

---

## 13. Development Workflow

### Makefile targets

| Target | Action |
|---|---|
| `make sqlc` | Run `sqlc generate` from `sqlc.yaml` → regenerate `db/sqlc/` |
| `make mock` | Run `go generate ./...` for all `//go:generate mockgen ...` directives |
| `make wire` | Run `./scripts/generate-wire.sh` → regenerate `wire_gen.go` |
| `make generate` | Run all three: sqlc + mock + wire |
| `make test` | Unit tests |
| `make test-integration` | Integration tests (needs running DB/Redis) |
| `make test-e2e` | End-to-end tests |
| `make check-tools` | Verify required CLI tools are installed |
| `make sqlc-lint` | Lint SQL via `./db/queries/lint.sh` |
| `make docs` | Build MkDocs site (serves on port 8081) |

### Required tools

`sqlc`, `mockgen`, `migrate`, `psql`, `golangci-lint`, `mkdocs`, `sleek`, `awoctl`

### Running the server

> **Termux note:** Do NOT start the server from this assistant — start it manually in a separate Termux session.

```bash
go build -o awo . && ./awo serve
# or
go run . serve
```

### Running migrations

```bash
migrate -path db/migration -database "postgresql://user:pass@localhost/dbname?sslmode=disable" up
```

### Seeding a fresh instance

```bash
BASE_URL=http://localhost:8080 \
TENANT_NAME="Acme Corp" \
ADMIN_EMAIL=admin@acme.com \
ADMIN_PASSWORD=secret \
bash scripts/seed.sh
```

### Regenerating all codegen

```bash
make generate   # sqlc → mock → wire
```

---

## 14. Other Modules

### Audit logging

- **Service:** `audit.Service` (`internal/core/audit/`)
- **Repository:** `audit/repository.go` — `CreateAuditEvent`, `ListAuditEvents`
- **Endpoint:** `GET /api/v1/audit-logs?limit=50&offset=0` (max limit 1000)
- **Permission gate:** `iam.sessions.read`
- **Pattern:** Every state-changing handler calls `auditService.Record(ctx, event)` after successful mutation

### Observability

**Tracing (OpenTelemetry):**
- Service interface: `tracing.Service` (`internal/shared/tracing/`)
- Methods: `StartSpan`, `SpanFromContext`, `InjectHTTPHeaders`, `RecordError`, `GetTraceID`, `Shutdown`
- Exporters: OTLP gRPC, OTLP HTTP, stdout (chosen by `TRACING_EXPORTER` config)
- PgX integration: `github.com/exaring/otelpgx` — DB queries emit spans automatically

**Metrics (Prometheus):**
- Interface: `metrics.MetricsProvider` (`internal/shared/metrics/`)
- Methods: `Counter`, `Gauge`, `Histogram`, `Timer`, `Handler()` (returns `/metrics` handler)
- Usage example: `metrics.IncrementCounter("iam.session.login.failure", labels...)`

**Logging (Zerolog):**
- Interface: `logger.Logger` (`internal/shared/logger/`)
- Adapters: Zerolog (default), Zap, Slog
- Scoped logger: `log.WithFields(logger.Fields{"component": "auth", "tenant": tenantID})`

### Feature flags

Resolved at login time via `ResolveAllFlagsForTenant(userID, tenantID)` and stored in `session.Configuration.Flags map[string]bool`. Handlers read `sess.Configuration.Flags["feature.name"]` — no per-request DB lookup.

### Tenant settings

Stored in `session.Configuration.Settings map[string]any` at login. Example: `iam.session_ttl_hours` overrides the process-level session TTL for a specific tenant.

### SSO / OAuth

- Service: `iam.SSOService`
- Providers: Google (confirmed via constants), extensible
- MFA intentionally skipped for SSO (IdP is the second factor)
- Flow: state-based PKCE, callback exchanges code for user profile, creates or links account, issues session

### API keys

- Service: `iam.APIKeyService`
- CRUD endpoints under `/api/v1/auth/api-keys/`
- `Authenticate` middleware accepts `Authorization: Bearer <key>` in addition to session cookie

### Entity hierarchy

- Module: `internal/core/entity/`
- Storage: PostgreSQL `ltree` extension for ancestor/subtree queries
- Used for: scoping data access within an organizational structure (companies, divisions, departments)

### Schema-driven UI

- AMIS SDK served from `/sdk/`
- JSON page schemas in `web/schemas/pages/`
- Boot endpoint `GET /api/v1/schema/boot` returns navigation filtered by session permissions and feature flags
- Dark mode: override AMIS CSS custom properties at `html.dark` (do NOT use AMIS `theme("dark")`)

---

## 15. Open Questions / Gaps

| # | Observation |
|---|---|
| 1 | **Temporal workflows not implemented.** SDK and config wired in, but no concrete `workflow.go` or `activity.go` files found. Password reset emails and onboarding jobs likely need this. |
| 2 | **`internal/core/iam/domain/user.go` appears nearly empty.** The user domain model may be defined elsewhere or is under construction. |
| 3 | **Casbin model string not inspectable.** `domain.CasbinModel` referenced but the actual CONF content was not visible in analysis. |
| 4 | **Email/notification service absent.** Password reset and MFA flows generate tokens but no mailer implementation was found. |
| 5 | **Rate limiting marked as TODO.** Comments in `routes.go` suggest rate limiting middleware is planned but not implemented. |
| 6 | **TLD-aware subdomain parsing incomplete.** `tenant.go` notes a TODO for multi-part TLDs (`.co.uk`, `.com.au`). |
| 7 | **`wire.go` location unclear.** The Wire source file(s) were not directly located during analysis. May need `make wire` to bootstrap after fresh clone. |
| 8 | **CORS config is dev-only.** Origins `localhost:3000` and `localhost:8080` hardcoded; production CORS config not visible. |
| 9 | **`db/queries/lint.sh` not inspected.** SQL linting script referenced in Makefile but contents unknown. |
| 10 | **Circuit breaker for TenantService commented out.** Noted in `middleware/tenant.go`; tenant validation on every request has no fallback under service degradation. |
| 11 | **`FeatureConfig` struct fields unknown.** Config struct includes `Features FeatureConfig` but fields and supported feature-flag keys are not documented. |
| 12 | **Finance module transactions.** It is unclear whether double-entry ledger constraints (debit = credit) are enforced at the DB level, service level, or both. |
