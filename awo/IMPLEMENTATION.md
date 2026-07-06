# Awo Framework — Implementation Status

Module: `awo.so` (part of root module, packages under `awo.so/awo/...`)

## Phase Status

| Phase | Status | Packages |
|---|---|---|
| 1 — Repository structure | ✅ Done | Directory tree established |
| 2 — def kernel | ✅ Done | `awo.so/awo/def` |
| 3 — filter DSL | ✅ Done | `awo.so/awo/filter` |
| 3 — registry | ✅ Done | `awo.so/awo/registry` |
| 4 — compiler | ✅ Done | `awo.so/awo/compiler` |
| 5 — runtime pipeline | ✅ Done | `awo.so/awo/runtime` |
| 5 — runtime/tenant | ✅ Done | `awo.so/awo/runtime/tenant` |
| 5 — runtime/naming | ✅ Done | `awo.so/awo/runtime/naming` |
| 6 — driver interfaces | ✅ Done | `awo.so/awo/driver` |
| 6 — pgx driver | ✅ Done | `awo.so/awo/contrib/pgx` |
| 6 — pgx sqlbuild | ✅ Done | `awo.so/awo/contrib/pgx/sqlbuild` |
| 6 — redis cache | ✅ Done | `awo.so/awo/contrib/redis` |
| 7 — tx abstraction | ✅ Done | `awo.so/awo/tx` |
| 7 — events/outbox | ✅ Done | `awo.so/awo/events` |
| 7 — cache abstraction | ✅ Done | `awo.so/awo/cache` |
| 7 — lock abstraction | ✅ Done | `awo.so/awo/lock` |
| 7 — platform/tenant | ✅ Done | `awo.so/awo/platform/tenant` |
| 7 — platform/iam | ✅ Done | `awo.so/awo/platform/iam` |
| 7 — platform/audit | ✅ Done | `awo.so/awo/platform/audit` |
| 7 — platform/flags | ✅ Done | `awo.so/awo/platform/flags` |
| 7 — platform/settings | ✅ Done | `awo.so/awo/platform/settings` |
| 7 — platform/metadata | ✅ Done | `awo.so/awo/platform/metadata` |
| 7 — platform/registry | ✅ Done | `awo.so/awo/platform/registry` |
| 7 — platform/notifications | ✅ Done | `awo.so/awo/platform/notifications` |
| 7 — internal/dberr | ✅ Done | `awo.so/awo/internal/dberr` |
| 8 — bootstrap | ✅ Done | `awo.so/awo/bootstrap` |
| 8 — api/response | ✅ Done | `awo.so/awo/api/response` |
| 8 — api/middleware | ✅ Done | `awo.so/awo/api/middleware` |
| 8 — api/handler | ✅ Done | `awo.so/awo/api/handler` |
| 8 — api/service | ✅ Done | `awo.so/awo/api/service` |
| 8 — api/router | ✅ Done | `awo.so/awo/api/router` |
| 8 — sdui | ✅ Done | `awo.so/awo/sdui` |
| 8 — observability/health | ✅ Done | `awo.so/awo/observability/health` |
| 8 — observability/metrics | ✅ Done | `awo.so/awo/observability/metrics` |
| 8 — cmd/migrate | ✅ Done | `awo.so/awo/cmd/migrate` |
| 8 — cmd/server | ✅ Done | `awo.so/awo/cmd/server` |
| 9 — examples | ✅ Done | `awo.so/awo/examples/finance` |
| 9 — tests | 🔄 In Progress | Per-package *_test.go files |

## Pending / Next

| Item | Notes |
|---|---|
| Temporal worker setup | `awo/cmd/server` wires Temporal client + activity registrations |
| Edge preloading | Join queries for `WithPreload` option |
| Pagination cursors | Keyset pagination for high-volume entities |

## Completed (this session)

| Item | Resolution |
|---|---|
| Filter query DSL parser | `api/filterparse` implemented; wired into `handler/crud.go` List |
| OpenAPI endpoint | `GET /api/openapi.json` registered in `cmd/server/main.go` |
| Outbox relay wiring | `outbox.New(pool).Start(ctx)` goroutine in `cmd/server/main.go` |
| Tenant/Auth/RateLimit middleware | Applied to `/api/v1/entities` group in `api/router/router.go` |
| `cache.NoopCounter` | Added to `cache` package; used as fallback when Redis unavailable |
| CORS middleware | `api/middleware/cors.go`; `*.BaseDomain` + explicit allowlist; env: `BASE_DOMAIN`, `CORS_ALLOWED_ORIGINS` |
| Casbin RBAC integration | `api/authz` enforcer wired into `router.RegisterOptions.Authz`; per-entity per-action; nil = RBAC disabled |
| IAM HTTP handlers | `platform/iam/handler.go`; POST /api/v1/auth/login, /logout, GET /me |
| Migration CLI | `cmd/migrate` wired to `golang-migrate/v4`; up/down/version/force; run `go mod tidy` to resolve |
| Platform notifications | `platform/notifications`: definition.go, hooks.go, driver.go, service.go, module.go, migrations/ |
| Platform metadata service | `platform/metadata/service.go`: FieldsForEntity, AddField, DeactivateField |
| Platform registry service | `platform/registry/service.go`: RegisterModule, Activate, Disable, ListActive |
| SQL migrations | All 8 platform modules have .up.sql/.down.sql pairs under each module's migrations/ directory |

## Package Dependency Graph

```
def          ← (stdlib, uuid, decimal only)           [FROZEN]
  ↑
filter       ← def.Filter interface
  ↑
registry     ← def (reads All(), seals)
  ↑
compiler     ← registry, def
  ↑
driver       ← def, filter, compiler
  ↑
tx           ← (stdlib only)
  ↑
events       ← def, uuid
  ↑
cache        ← (stdlib only)
lock         ← (stdlib only)
  ↑
runtime      ← compiler, def
  runtime/tenant  ← (stdlib, uuid only)
  runtime/naming  ← cache
  ↑
internal/dberr ← runtime (for error types), pgconn
  ↑
contrib/pgx  ← pgx/v5, driver, runtime, compiler, dberr, tx, tenant
contrib/pgx/sqlbuild ← filter
contrib/redis ← go-redis, cache, lock
  ↑
platform/*   ← def, driver, runtime, cache
  ↑
bootstrap    ← compiler, registry, pgxpool, go-redis
  ↑
api/*        ← compiler, driver, runtime, platform/iam
sdui         ← compiler, def, cache
observability/* ← prometheus, pgxpool, go-redis, def
  ↑
cmd/*        ← bootstrap, api/*, observability/*
```

## Key Architectural Decisions

1. **def has zero framework deps** — only stdlib + uuid + decimal.
2. **filter.Filter implements def.Filter interface** — no circular import.
3. **Registry sealed once** — `def.Seal()` called from `registry.Build()`.
4. **Pipeline does not open transactions** — delegates to driver.
5. **TenantContext in context.Context only** — panics on absent context (intentional).
6. **No lazy loading** — edges always explicitly requested via QueryOption.
7. **No ORM types** — all persistence through EntityRepository[T] interface.
8. **Token hashing** — session tokens hashed with SHA-256; only hash stored in DB.
9. **Cache fail-open for flags/schemas** — Redis miss falls through to DB.
10. **Cache fail-closed for sessions** — Redis failure returns 503 (correct security).
11. **Workflow starts outside TX** — outbox table provides guaranteed retry.
12. **Custom fields use cf_ prefix** — enforced by FieldNameValidator hook.
13. **Entity names never renamed** — embedded in Temporal IDs, Redis keys, migrations.
