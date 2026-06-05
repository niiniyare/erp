---
title: Worked Example — Launch Checklist
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[MDG Quick Checklist](../01-overview/03-mdg-checklist.md)"
  - "[MDG Testing Overview](../18-testing/01-testing-overview.md)"
  - "[Worked Example Overview](01-worked-example-overview.md)"
---

# Worked Example — Launch Checklist

Use this checklist to verify the Contracts module is complete and ready for review.

## Domain

- [ ] `Contract` struct has all required fields (uuid, tenant_id, entity_id, version, deleted_at)
- [ ] `ContractStatus` is `type ContractStatus string` — not int/iota
- [ ] `CanTransitionTo` tested with complete state matrix
- [ ] All 7+ sentinel errors defined as `errors.New` — no aliases
- [ ] All domain events implement `Event` interface (Topic() + GetTenantID())
- [ ] Monetary values use `decimal.Decimal` — no float64

## Database

- [ ] Migration files follow `{group}{seq}_{name}.{up|down}.sql` naming
- [ ] All PKs use `uuid DEFAULT gen_random_uuid()` — no SERIAL
- [ ] `tenant_id` column on every table with NOT NULL + FK
- [ ] `version integer NOT NULL DEFAULT 1` on every mutable table
- [ ] `deleted_at timestamptz` (nullable) on every table — NOT NULL prohibited
- [ ] RLS enabled: `ALTER TABLE t ENABLE ROW LEVEL SECURITY`
- [ ] RLS policy created using `current_setting('app.tenant_id')::uuid`
- [ ] `missing_ok` never used in RLS policy
- [ ] Indexes on `(tenant_id, status)` and `(tenant_id, entity_id)` with `WHERE deleted_at IS NULL`
- [ ] `make migrate-up` succeeds on clean DB
- [ ] `make migrate-down` succeeds (rollback verified)

## SQLC Queries

- [ ] `make sqlc` runs without errors
- [ ] `CreateContract` hardcodes `status = 'draft'` — not a parameter
- [ ] `UpdateContract` uses `WHERE version = @version RETURNING *`
- [ ] `SoftDeleteContract` uses `WHERE deleted_at IS NULL` — not hard delete
- [ ] `ListContracts` uses `sqlc.narg()` for optional filters
- [ ] `CountContracts` mirrors `ListContracts` WHERE clause exactly

## Repository

- [ ] `ContractRepository` interface matches all service calls
- [ ] Every method wraps queries in `WithTenant`
- [ ] `mapContractDBError` handles SQLSTATE 23505 → `ErrContractAlreadyExists`
- [ ] `Update` disambiguates ErrNoRows as conflict vs. not found
- [ ] `List` returns empty slice (not nil) when no rows

## Service

- [ ] `NewContractService` injects all dependencies
- [ ] Every write method: authorize first, then persist
- [ ] Every transition method: authorize → fetch → `CanTransitionTo` → persist
- [ ] Audit recorded async with `context.Background()` + 5s timeout + recover()
- [ ] Events published async with `context.Background()` + 5s timeout + recover()
- [ ] Notifications published async with `context.Background()` + 5s timeout + recover()
- [ ] None of the async goroutines propagate errors to the caller

## Handler

- [ ] `mapError` uses `errors.Is` — not type switch
- [ ] `mapError` handles all 7 domain sentinel errors
- [ ] `Create` returns 201 + `Location` header
- [ ] `Delete` returns 204 No Content
- [ ] Invalid UUID param → 400
- [ ] Missing required field → 422 with field-level details
- [ ] Handler tests cover all status codes in §22 coverage checklist

## Wire

- [ ] `ProviderSet` defined in `contracts.go`
- [ ] `WorkerSet` defined in `contracts.go`
- [ ] Both imported in `InitializeApp`
- [ ] `make wire` succeeds
- [ ] `wire_gen.go` committed
- [ ] Service compiles with `go build ./...`

## Routes

- [ ] All routes have `Authenticate` + `TenantContext` middleware
- [ ] Each route has `Authorize` with correct permission string
- [ ] Status transition endpoints use `POST /:id/{action}` pattern
- [ ] Import/export endpoints have `StrictRateLimit` applied

## Tests

- [ ] Domain state machine tests: all valid + invalid transitions covered
- [ ] Repository integration tests: Create, GetByID, Update conflict, cross-tenant block
- [ ] Service unit tests: auth deny, invalid transition, conflict propagation
- [ ] Handler tests: 200, 201, 400, 401, 403, 404, 409, 422
- [ ] Coverage >= 80% for `./internal/core/contracts/...`

## Observability

- [ ] Service constructor creates a child logger with `"service": "contracts"`
- [ ] Every method starts an OTel span: `s.tracer.Start(ctx, "contract.service.{method}")`
- [ ] Metrics: request counter and latency histogram registered

## Operations

- [ ] Feature flag `contracts.enabled` checked before create/update operations
- [ ] `/health/ready` continues to pass after contracts module added
- [ ] No secrets or credentials in code or migration files
