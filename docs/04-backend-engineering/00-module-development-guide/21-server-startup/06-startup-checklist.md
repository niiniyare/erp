---
title: Startup Checklist
portal: 4 — Backend Engineering
section: 00-module-development-guide/21-server-startup
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-startup-overview.md
    title: Startup Overview
---

# Startup Checklist

Use this checklist when adding a new module to the server.

## Wire

- [ ] Module has `ProviderSet` defined in `{module}/{module}.go`
- [ ] `ProviderSet` includes all providers: repo, service, handler, route registrar
- [ ] `ProviderSet` imported and added to `InitializeApp` wire injector
- [ ] `make wire` runs without errors
- [ ] `wire_gen.go` updated (committed)

## Routes

- [ ] `ContractRouteRegistrar` (or equivalent) added to `[]RouteRegistrar` in Wire
- [ ] All routes registered with correct middleware chain
- [ ] Auth middleware on all non-public routes
- [ ] Authorize middleware with correct permission string on each route
- [ ] No route registered without authentication

## Database

- [ ] Migration files present in `db/migration/`
- [ ] Migrations follow naming convention `{sequence}_{description}.{up|down}.sql`
- [ ] RLS enabled on all new tables
- [ ] `make migrate-up` succeeds on clean database
- [ ] `make migrate-down` succeeds (rollback works)

## Temporal Workers

- [ ] Worker registered in `WorkerSet` via Wire
- [ ] Worker started in startup goroutine
- [ ] Worker stopped in graceful shutdown
- [ ] All workflows and activities registered in `worker.go`

## Event Subscriptions

- [ ] Subscriber struct implements `EventSubscriber` interface
- [ ] Subscriber added to `[]EventSubscriber` slice in Wire
- [ ] `Register()` called at startup
- [ ] Handlers are idempotent

## Health

- [ ] `checkReadiness` verifies all critical deps (DB at minimum)
- [ ] `/health/live` returns 200 immediately
- [ ] `/health/ready` returns 503 during startup, 200 after

## Environment Variables

- [ ] All required env vars documented
- [ ] Required vars cause `Fatal` on startup if missing (not silent default)
- [ ] No secrets hardcoded — all via env vars
- [ ] `.env.example` updated with new variables

## Observability

- [ ] Tracer initialized before `InitializeApp`
- [ ] Module metrics registered in Prometheus at init time
- [ ] Log level configurable via `LOG_LEVEL` env var
