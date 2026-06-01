---
title: MDG Quick Checklist
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[What This Guide Covers](01-what-this-guide-covers.md)"
  - "[Guide Conventions](02-guide-conventions.md)"
  - "[Launch Checklist](../23-worked-example/10-launch-checklist.md)"
---

# MDG Quick Checklist

Condensed checklist for building a new module. Each item links to the detailed section.

## Domain Layer
- [ ] Domain struct with `uuid.UUID` PKs, `decimal.Decimal` for money, string status types
- [ ] Status type with `CanTransitionTo()` method
- [ ] Transition map tested with full matrix
- [ ] `BusinessError` struct with Code, Message, Status, Err
- [ ] Sentinel errors: `ErrNotFound`, `ErrNotEditable`, `ErrVersionConflict`, `ErrForbidden`
- [ ] Domain events with `Topic()` and `GetTenantID()` methods
- [ ] Monetary event fields as `string` (not `decimal.Decimal`)

## Database Layer
- [ ] Migration file named `{group}{seq:04d}_{description}.up.sql`
- [ ] `.down.sql` reverse migration written and tested
- [ ] All columns: identity → business → relationships → control (in order)
- [ ] `RLS ENABLE` + `CREATE POLICY tenant_isolation`
- [ ] `updated_at` trigger created
- [ ] Indexes on `(tenant_id, status)` and `(tenant_id, created_at DESC)` with `WHERE deleted_at IS NULL`
- [ ] Status column uses `text NOT NULL CHECK (status IN (...))`
- [ ] Money columns use `numeric(20,6)`
- [ ] No `missing_ok` in `SET LOCAL app.tenant_id`

## SQLC Layer
- [ ] Query file in `db/queries/{module}.sql`
- [ ] All queries use `current_setting('app.tenant_id')::uuid` not `@tenant_id`
- [ ] `RETURNING *` on all mutations
- [ ] `COUNT(*) OVER()` for paginated list queries
- [ ] `make sqlc` run; generated files committed

## Repository Layer
- [ ] Interface defined in `interface.go`
- [ ] All methods call `store.WithTenant(ctx, tenantID, ...)`
- [ ] DB errors mapped in `parseModuleDBError()`: `ErrNoRows` → sentinel, 23505 → BusinessError{409}
- [ ] Domain mapper for every SQLC row type
- [ ] No `pgx` types escape the repository layer

## Service Layer
- [ ] Permission check (`authz.Can`) is first operation
- [ ] State machine validated before repo write
- [ ] Async side effects in goroutines: `context.Background()`, 5s timeout, `recover()`
- [ ] Audit logged after every state change
- [ ] Events published for external consumers
- [ ] Feature flag guard where applicable
- [ ] `sess.TenantID` for all repo calls (never URL/body)

## Handler Layer
- [ ] Request DTO with validation tags
- [ ] UUID path params validated with 400 on bad format
- [ ] Service errors mapped via `mapError()` using `errors.Is`/`errors.As`
- [ ] 201 with `Location` header for creates
- [ ] List response: `{"data": [...], "pagination": {...}}`
- [ ] `page_size` capped at 100

## Wire Registration
- [ ] `ProviderSet` in `internal/core/{module}/wire.go`
- [ ] `WorkerSet` for Temporal (if module has workflows)
- [ ] Interface bindings with `wire.Bind`
- [ ] Added to `cmd/server/wire.go` `Build()` call
- [ ] `RouteRegistrar.Register()` added to `NewRouteRegistry`
- [ ] `make wire` run; `wire_gen.go` committed

## API
- [ ] Routes follow `{module-noun}` pattern (no verbs in paths)
- [ ] Actions use `POST /:id/{action}`
- [ ] All routes protected: `Authenticate` + `Authorize(perm)`
- [ ] Permission strings follow `{module}.{resource}.{action}` convention
- [ ] API documented in Portal 8

## Testing
- [ ] Unit tests: all service methods with mock repo + authz
- [ ] Unit tests: state machine transition matrix
- [ ] Integration tests: RLS isolation (cross-tenant returns 404)
- [ ] Integration tests: create → list → get round trip
- [ ] Handler tests: 2xx success, 4xx error cases, validation
- [ ] All tests pass with `-race` flag

## Observability
- [ ] Module logger with `"module"` field
- [ ] Span created per operation: `module.Type.Method`
- [ ] Metrics registered: created_total, operation_duration
- [ ] Error logged with tenant_id, resource_id, error fields
- [ ] No sensitive data (password, token) in logs

## Deployment
- [ ] Migration applied to staging and verified
- [ ] `make wire && make sqlc` outputs committed
- [ ] Feature flags registered (if applicable)
- [ ] Default permissions added for system roles
- [ ] API reference updated
