---
title: Development Sequence
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Module Anatomy](02-module-anatomy.md)"
  - "[Conventions Cheatsheet](04-conventions-cheatsheet.md)"
  - "[Worked Example](../23-worked-example/01-worked-example-overview.md)"
---

# Development Sequence

Ordered checklist for building an AwoERP module from scratch. Every step maps to a guide section. Follow in order — later steps assume earlier ones are done.

## Pre-Work

- [ ] **Assign module group number** (`NNN`). Check `db/migration/` for the highest prefix in use. Increment by 1. For the contracts example: `011`.
- [ ] **Register module key** in the product backlog. Key is lowercase, plural (`contracts`).
- [ ] **List all actions** the module needs: `create`, `read`, `update`, `delete`, and any domain-specific verbs (`submit`, `approve`, `activate`, `terminate`).
- [ ] **Draw the state machine** on paper. Every state and every valid transition. This diagram is the source of truth for `CanTransitionTo()`.

---

## Step 1 — DDD Domain Design → §02

- [ ] Create `internal/core/<module>/domain/` directory.
- [ ] Define primary entity struct with all required fields (`ID`, `TenantID`, `EntityID`, `Status`, `Version`, `CreatedBy`, `UpdatedBy`, `DeletedAt`, `CreatedAt`, `UpdatedAt`).
- [ ] Define `<Noun>Status` type (`type ContractStatus string`) and all constants.
- [ ] Implement `CanTransitionTo(next <Noun>Status) bool` on the entity.
- [ ] Define value objects with validation constructors (monetary amounts, reference numbers, etc.).
- [ ] Write `domain/errors.go` — one sentinel error per distinct failure caller handles differently.
- [ ] Write `domain/events.go` — one event struct per state change; all include `TenantID`; names in past tense.

**Gate**: Domain compiles with zero external imports except `github.com/google/uuid`, `time`, `errors`.

---

## Step 2 — Database Schema + RLS → §03

- [ ] Create `db/migration/NNN_001_create_<module>.sql` — primary table.
  - Every column has `NOT NULL` or a documented reason it is nullable.
  - `tenant_id uuid NOT NULL REFERENCES tenants(id)`.
  - `entity_id uuid NOT NULL REFERENCES entities(id)`.
  - `version integer NOT NULL DEFAULT 1`.
  - `deleted_at timestamptz` (nullable — soft delete sentinel).
  - Monetary columns use `numeric(20,6)`.
  - PK is `uuid DEFAULT gen_random_uuid()`.
  - `status` column uses `VARCHAR(50)` with a `CHECK` constraint listing valid states.
- [ ] Add RLS block: `ALTER TABLE ... ENABLE ROW LEVEL SECURITY; CREATE POLICY ...`.
- [ ] Add indexes: tenant_id, status, any FK, any query-filter column.
- [ ] Create `db/migration/NNN_002_create_<module>_lines.sql` for child tables (if needed).
- [ ] Create `db/migration/NNN_003_create_<module>_views.sql` for useful read views (if needed).

**Gate**: Run `make migrate-up` locally; `make migrate-down` rolls back cleanly; `make migrate-up` again succeeds.

---

## Step 3 — SQLC Annotated Queries → §04

- [ ] Create `db/queries/<module>.sql`.
- [ ] Write `-- name: Create<Noun> :one` query (INSERT ... RETURNING *).
- [ ] Write `-- name: Get<Noun>ByID :one` query (includes `AND deleted_at IS NULL` guard).
- [ ] Write `-- name: List<Nouns> :many` query with dynamic filter pattern.
- [ ] Write `-- name: Count<Nouns> :one` for pagination.
- [ ] Write `-- name: Update<Noun> :one` (optimistic lock: `WHERE version = @version AND deleted_at IS NULL`, increments version).
- [ ] Write `-- name: UpdateStatus<Noun> :one` (separate query — only touches status + version + updated_by + updated_at).
- [ ] Write `-- name: SoftDelete<Noun> :one` (sets `deleted_at = now()`).
- [ ] Write `-- name: <Noun>ExistsWithNumber :one` (unique-key existence check).
- [ ] Run `make sqlc` — confirm generated code in `db/sqlc/`.

**Gate**: SQLC generates without warnings. All param structs match query placeholders.

---

## Step 4 — Repository Interface + SQLC Adapter → §05

- [ ] Create `internal/core/<module>/repository/<noun>.go` — interface only, zero implementation.
  - Methods match query set: `Create`, `GetByID`, `List`, `Count`, `Update`, `UpdateStatus`, `Delete`, `ExistsWithNumber`.
  - All methods accept `ctx context.Context` as first arg.
  - All list/filter methods use a dedicated params struct, not variadic args.
- [ ] Create `internal/core/<module>/repository/<noun>_sqlc.go` — SQLC adapter.
  - Struct embeds `db.Store`.
  - Constructor: `func New<Noun>Repository(store db.Store) <Noun>Repository`.
  - Each method calls `store.WithTenant(ctx, tenantID, func(q *db.Queries) error {...})`.
  - Each method maps SQLC row → domain struct; maps DB errors → domain sentinels.

**Gate**: `repository` package compiles. No business logic in any repository method.

---

## Step 5 — Service Layer → §06

- [ ] Create `internal/core/<module>/service/<noun>.go`.
- [ ] Define `<Noun>Service` interface — one method per use-case.
- [ ] Implement service struct with constructor injecting: repo, authzSvc, auditSvc, notifSvc, logger, tracer, metrics.
- [ ] Every write method follows sequence:
  1. Extract `ResolvedSession` from `c.Locals(domain.LocalsKeySession)` (in handler, not service).
  2. Service receives pre-extracted principal; calls `authzSvc.Enforce(ctx, domain.Request{...})`.
  3. Validate business rules.
  4. Call `repo` method inside `store.WithTenant`.
  5. Emit domain event.
  6. Fire async notifications.
  7. Return domain entity.
- [ ] Every status-change method calls `entity.CanTransitionTo(newStatus)` before calling repo.
- [ ] Optimistic lock conflict from repo propagates as `ErrContractConflict` to handler.

**Gate**: Service compiles. No SQLC types leak through service interface — only domain types.

---

## Step 6 — Core Module Integrations → §07

- [ ] **Tenant**: All DB calls use `store.WithTenant(ctx, sess.TenantID, ...)`. Never read tenant_id from URL params.
- [ ] **IAM — route level**: `middlewarePkg.Authorize(cfg, "module.resource.action")` on each route.
- [ ] **IAM — service level**: `authzSvc.Enforce(ctx, domain.Request{Subject: sess.ToPrincipal().Subject, Domain: iam.TenantDomain(sess.TenantID), Object: "module/resource/*", Action: "action"})` for non-route checks.
- [ ] **Settings**: `sess.SettingString("module.setting_key", "default")` for tenant-specific behaviour.
- [ ] **Feature flags**: `sess.FeatureEnabled("module.flag_name")` gates new behaviour.
- [ ] **Metadata**: Attach `EntityID` on create; respect EntityScope from session for list queries.

**Gate**: All integration points use session snapshot, never re-fetch during request.

---

## Step 7 — Instrumentation → §08

- [ ] Add span per service method: `ctx, span := r.tracer.Start(ctx, "ContractService.Create")`.
- [ ] Add `span.SetAttributes(...)` for tenant_id, contract_id on success.
- [ ] Add `span.RecordError(err)` on failure path.
- [ ] Log at method entry (`DEBUG`) and on error (`ERROR`).
- [ ] Increment counter metric on create/update/delete.
- [ ] Wrap instrumentation in deferred recovery — never let span/metric code fail the request.

**Gate**: No instrumentation call can return an error that propagates to the caller.

---

## Step 8 — Audit Trail → §09

- [ ] Call `auditSvc.Record(ctx, audit.Event{...})` after every successful write.
- [ ] Include: `TenantID`, `EntityID`, `ActorID`, `Action`, `ResourceType`, `ResourceID`, `Before` (JSON), `After` (JSON), `OccurredAt`.
- [ ] Fire audit asynchronously (goroutine + timeout) — never block the response.
- [ ] Test that audit record is written in integration tests.

**Gate**: Every state-changing operation has a corresponding audit event.

---

## Step 9 — Notifications → §10

- [ ] Define notification events for significant state changes (e.g., `ContractApproved`, `ContractExpiringSoon`).
- [ ] Call `notifSvc.Send(ctx, notif.Message{...})` asynchronously after successful writes.
- [ ] Respect user notification preferences — check before sending.
- [ ] Never block response path waiting for notification delivery.

**Gate**: Notification failure does not fail the originating request.

---

## Step 10 — Temporal Workflows → §11

_(Skip if module has no multi-step or human-approval flows.)_

- [ ] Create `internal/core/<module>/service/<noun>_workflow.go`.
- [ ] Define workflow function: `func <Noun>ApprovalWorkflow(ctx workflow.Context, input <Noun>WorkflowInput) error`.
- [ ] Define activity struct with injected service dependency.
- [ ] Register workflow and activities with Temporal worker.
- [ ] Call `temporalClient.ExecuteWorkflow(...)` from service when triggering the workflow.
- [ ] Handle workflow timeouts and compensation paths.

**Gate**: Workflow compiles and registers without name conflicts.

---

## Step 11 — Pipeline Integration → §12

_(Skip if module has no hook/plugin extension points.)_

- [ ] Create `internal/core/<module>/service/<noun>_pipeline.go`.
- [ ] Define `<Noun>Stage` and `<Noun>Hook` types.
- [ ] Implement pipeline `Execute(ctx, OperationContext)` with compensation stack.
- [ ] Support `DryRun` mode — no writes, returns what would happen.
- [ ] Register hooks in module facade.

**Gate**: Pipeline dry-run test passes with no DB writes.

---

## Step 12 — Event-Driven Integration → §13

- [ ] Publish `domain.ContractCreated`, `domain.ContractActivated`, etc. to the event bus after writes.
- [ ] Subscribe to cross-module events this module cares about (if any).
- [ ] Ensure idempotent event handlers — duplicate delivery must not double-mutate state.

**Gate**: Event publish is fire-and-forget; failure is logged not propagated.

---

## Step 13 — REST API Design → §14

- [ ] Define resource paths following `/<version>/tenants/:tenantID/<module>/<nouns>`.
- [ ] Map HTTP verbs to use-cases: `POST`=create, `GET`=read/list, `PUT`=update, `DELETE`=delete, domain verbs as `POST /<id>/submit`.
- [ ] Write OpenAPI / inline documentation for all endpoints.
- [ ] Decide pagination strategy: cursor or offset.

**Gate**: No two routes have the same method + path combination.

---

## Step 14 — Fiber HTTP Handlers → §15

- [ ] Create `internal/api/handlers/<module>/handler.go` — struct + constructor.
- [ ] Create `internal/api/handlers/<module>/request.go` — request DTOs with `validate` tags.
- [ ] Create `internal/api/handlers/<module>/response.go` — response DTOs + `MapXxxToResponse()`.
- [ ] Create `internal/api/handlers/<module>/errors.go` — `mapError(err error) error`.
- [ ] Create `internal/api/handlers/<module>/routes.go` — `RegisterRoutes(apiGroup fiber.Router, deps *handlers.Dependencies)`.
- [ ] Every handler: parse → validate → extract session → call service → map response.
- [ ] Extract session: `sess, ok := c.Locals(domain.LocalsKeySession).(*iam.ResolvedSession)`.
- [ ] Return errors via `mapError(err)` — never raw errors to client.

**Gate**: No handler imports SQLC or `db` packages directly.

---

## Step 15 — Middleware Chain → §16

- [ ] Confirm every route group has: authenticate → authorize → handler.
- [ ] `Authenticate` middleware: validates JWT/session, sets `c.Locals(domain.LocalsKeySession, resolvedSession)`.
- [ ] `Authorize` middleware: `middlewarePkg.Authorize(*deps.AuthConfig, "module.resource.action")`.
- [ ] Rate limiting applied at API group level (not per-route).
- [ ] No business logic in middleware — only cross-cutting concerns.

**Gate**: Request without valid session returns `401`. Request with valid session but wrong role returns `403`.

---

## Step 16 — UI Schema Design → §17

- [ ] Define schema pages: list view, detail view, create form, edit form.
- [ ] Map every API endpoint to a schema action.
- [ ] Define column set for list: which columns, sortable, filterable.
- [ ] Define form field set: input types, validation, conditional visibility.
- [ ] Ensure `tenant_id` never appears as a user-editable field.

**Gate**: Schema renders all CRUD operations without hardcoded data.

---

## Step 17 — amis Web Schemas → §18

- [ ] Create `web/schemas/pages/<module>/list.json` — CRUD list with toolbar.
- [ ] Create `web/schemas/pages/<module>/detail.json` — read-only detail view.
- [ ] Create `web/schemas/pages/<module>/form.json` — create/edit form with validation.
- [ ] Wire API paths using amis `api` object with correct HTTP method and URL.
- [ ] Map domain status values to amis `mapping` badge colors.
- [ ] Test schema in amis playground before registering.

**Gate**: List loads data, form submits, detail shows all fields without console errors.

---

## Step 18 — Flutter Mobile Schemas → §19

_(Mark PLANNED if mobile not yet targeted.)_

- [ ] Define Flutter DSL schema for list screen.
- [ ] Define Flutter DSL schema for detail screen.
- [ ] Define Flutter DSL schema for create/edit form.
- [ ] Map validation rules to Flutter form field validators.

---

## Step 19 — Wire Dependency Injection → §20

- [ ] Create `internal/core/<module>/<module>.go` — module facade.
- [ ] Define `<Module>RepositorySet`, `<Module>ServiceSet`, `<Module>Set` wire.NewSet.
- [ ] Add `<Module>Set` to the application wire provider in `internal/app/wire.go` (or equivalent).
- [ ] Add handler constructor to handler wire set.
- [ ] Run `make wire` — fix any injection errors.

**Gate**: `make wire` succeeds with zero errors. `wire_gen.go` never edited by hand.

---

## Step 20 — Route Registration → §21

- [ ] Call `<module>Handler.RegisterRoutes(apiGroup, deps)` in `internal/api/handlers/routes.go`.
- [ ] Confirm route appears in application startup log.
- [ ] Run a smoke test: `curl` the list endpoint with a valid token.

**Gate**: All routes appear in Fiber route list. No 404 on known paths.

---

## Step 21 — Server Startup Sequence → §21

- [ ] Module's Wire provider included in `app.go` provider set.
- [ ] No `init()` functions in module code — all initialisation via Wire constructors.
- [ ] Module-specific config loaded from environment variables via config struct.

**Gate**: Server starts without panic. Module routes registered.

---

## Step 22 — Testing → §22

- [ ] Unit tests: domain state machine (`CanTransitionTo` matrix), value object constructors, error sentinel coverage.
- [ ] Repository integration tests: each repo method against a real test DB (no mocks).
- [ ] Service unit tests: mock repository; verify authz is called before every write; verify `CanTransitionTo` is checked before status updates.
- [ ] Handler integration tests: full fiber app with test DB; test 401/403/422/200 paths.
- [ ] Run `make test` — all pass.
- [ ] Run `make lint` — zero linter errors.

**Gate**: Coverage ≥ 80% on service package. All integration tests use real DB — no mocked DB.

---

## Step 23 — Worked Example Verification → §23

- [ ] Review complete worked example in §23 against your implementation.
- [ ] Verify migration prefix matches assigned `NNN`.
- [ ] Verify permission strings match `module.resource.action` format throughout.
- [ ] Verify no raw SQL outside `db/queries/`.
- [ ] Verify no float64 for monetary fields anywhere.
- [ ] Verify `wire_gen.go` reflects current provider sets.

**Gate**: Module is production-ready. Hand off to QA.
