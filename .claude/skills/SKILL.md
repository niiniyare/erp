---
name: awo-erp
description: >
  AWO ERP architecture assistant. Use this skill for ANY question or task involving the AWO ERP
  codebase — including writing new handlers, services, repositories, Wire providers, Temporal
  workflows, RLS policies, migrations, SQLC queries, middleware, tests, or anything else touching
  the awo.so module. Trigger whenever the user mentions AWO, the ERP, multi-tenancy, RLS tenant
  isolation, Wire DI wiring, Temporal workflows, Fiber handlers, SQLC queries, audit logging,
  IAM/RBAC/ABAC, finance module, entity hierarchy, or asks about Go patterns in the context of
  this codebase. Do not wait for the user to say "use the skill" — activate whenever AWO ERP
  context is relevant.
---

# AWO ERP — Architecture Skill

You are working inside the **AWO ERP** codebase (`awo.so`). Before writing any code or giving
any architectural advice, internalise all of the following. Do not guess — use these exact
file paths, struct names, and patterns.

For deep reference on specific areas, read the appropriate file in `references/`:
- `references/rls-tenancy.md` — Multi-tenancy, RLS, context propagation
- `references/wire-di.md` — Wire provider sets, Dependencies struct, wiring new domains
- `references/temporal.md` — Temporal scaffolding, patterns for new workflows
- `references/conventions.md` — Naming, errors, repository pattern, testing

---

## Stack at a glance

| Concern | Library / pattern |
|---|---|
| HTTP | Fiber v2 |
| DB driver | pgx/v5 |
| Query gen | SQLC (`db/sqlc/`) |
| Migrations | golang-migrate (`db/migration/`) |
| DI | Google Wire (compile-time) |
| Workflows | Temporal SDK v1.41.1 |
| Auth/Authz | Casbin v2 + CEL |
| Cache | Redis (`cache.Service`) |
| Config | Viper (`internal/platform/config/`) |
| Logging | Zerolog (default adapter via `logger.Logger`) |
| Errors | custom unified Errors `internal/shared/errors`  |
| Tracing | OpenTelemetry (`tracing.Service`) |
| Metrics | Prometheus (`metrics.MetricsProvider`) |
| Mocking | `go.uber.org/mock/mockgen` |
| Tests | testify + miniredis |

---

## Critical rules — never violate these

1. **Never hand-edit `db/sqlc/`** — regenerate with `make sqlc`.
2. **Never hand-edit `wire_gen.go`** — regenerate with `make wire`.
3. **Every table query must run inside `store.WithTenantFromCtx()`** — raw queries without tenant
   context will silently bypass RLS.
4. **Repository interfaces live in `repository/<noun>.go`; SQLC adapters in `repository/<noun>_sqlc.go`**.
5. **No `I` prefix on interfaces** — use role suffix: `UserService`, `UserRepository`.
6. **Constructors are `New<Type>(deps...) (*Type, error)`**.
7. **Errors follow the three-layer pattern**: domain sentinel → `BusinessError` → HTTP mapped error.
8. **Do not start the server from the assistant** — user runs it manually in a separate Termux session.
9. **`wire.go` has build tag `//go:build wireinject`** — always include it.
10. **`make generate` = sqlc → mock → wire** — run after any codegen change.

---

## Directory map (key paths)

```
cmd/                          # Cobra CLI entry points
db/
  migration/                  # *.up.sql / *.down.sql  (format: 001001_description.sql)
  queries/                    # SQLC source .sql files
  db/sqlc/                    # Generated — do not edit
internal/
  api/
    handlers/
      routes.go               # RegisterRoutes() + Dependencies struct
      auth/                   # auth.go, mfa.go, oauth.go, apikey.go
      audit/handler.go
      entity/handler.go
      finance/handler.go
      tenant/handler.go
      user/handler.go
    middleware/
      auth.go                 # authenticateMiddleware, Authorize()
      tenant.go               # TenantMiddleware — RLS injection
      observability.go
      security.go             # RouteSecurityManager (CORS/CSRF/headers)
      whitelist.go            # EndpointWhitelist — public route bypass
  core/
    audit/
    entity/
    finance/service/          # account.go, transaction.go, reporting.go
    iam/
      domain/                 # Value objects, ActorType, EntityScopeType, errors
      repository/             # Interfaces + SQLC adapters
      service/                # user.go, session.go, authz.go, sso.go, apikey.go
      iam.go                  # Facade — re-exports types + interfaces
    tenant/
      domain/                 # tenant.go, status.go, errors.go
      repository/             # repository.go interface + sqlc adapter
      service.go
      tenant.go               # Facade
  platform/
    cache/                    # cache.Service (Redis)
    config/config.go          # Config struct, Load(), LoadWithViper()
    temporal/                 # Client + worker bootstrap
  shared/
    context.go                # TenantIDKey, UserIDKey, SessionKey helpers
    errors/http.go            # ToHTTPError() via errors.As
    errors/types.go           # BusinessError, RepositoryError, HTTPError
    logger/                   # Logger interface + adapters
    metrics/                  # MetricsProvider interface
    tracing/                  # tracing.Service interface
scripts/
  generate-wire.sh
  seed.sh                     # REST smoke test / bootstrap
test/
  integration/                # Real Postgres + Redis (docker-compose.test.yml)
  e2e/
web/schemas/pages/            # AMIS JSON page schemas
```

---

## Quick patterns (use these verbatim)

### Reading tenant / user from context

```go
tenantID, ok := shared.GetTenantID(ctx)
if !ok {
    return nil, shared.ErrMissingTenantID // or equivalent BusinessError
}

userID, ok := shared.GetUserID(ctx)
sess, ok := shared.GetSession(ctx)
```

### Running a query with RLS

```go
err := r.store.WithTenantFromCtx(ctx,func(ctx context.Context, s db.Store) error {
    _, err := s.SomeQuery(ctx, db.SomeQueryParams{...})
    return err
})
```

### Authorize middleware on a route

```go
app.Get("/api/v1/finance/accounts/",
    Authorize("finance.accounts.read"),
    deps.FinanceHandler.ListAccountsHandler,
)
```

### Domain error → HTTP error mapping

```go
// domain/errors.go
var ErrFooNotFound = errors.New("foo not found")

// handler layer
func mapFooError(err error) *fiber.Error {
    switch {
    case errors.Is(err, domain.ErrFooNotFound):
        return fiber.NewError(fiber.StatusNotFound, err.Error())
    default:
        return fiber.ErrInternalServerError
    }
}
```

### Audit recording

```go
// after successful mutation in handler
_ = deps.AuditService.Record(c.UserContext(), audit.Event{
    Action:   "foo.created",
    ActorID:  userID,
    TenantID: tenantID,
    Resource: "foo",
    ResourceID: foo.ID.String(),
})
```

---

## Tenant status lifecycle

```
PENDING → ACTIVE → SUSPENDED → ACTIVE
   │           │
   └───────────┴──► ARCHIVED
```

Transitions enforced in `internal/core/tenant/service.go`. Do not implement ad-hoc.

---

## Permission string format

`<module>.<resource>.<action>` — e.g. `finance.accounts.read`, `iam.sessions.read`.
Permissions are pre-loaded into `session.Permissions []string` at login. No per-request DB lookup.

---

## Open gaps to be aware of

- Temporal: SDK wired, zero concrete workflows yet — patterns in `references/temporal.md`
- Rate limiting: TODO in `routes.go`
- Email/notification service: absent — password reset tokens are generated but not delivered
- CORS: dev-only origins hardcoded; production config needed
- Double-entry enforcement for finance: enforcement layer (DB vs service) not confirmed
- TLD-aware subdomain parsing: TODO for `.co.uk`, `.com.au`
