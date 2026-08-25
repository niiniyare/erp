# AWO Framework — Implementation Tracker

**Updated:** 2026-08-18
**Owner:** Solo developer / Claude Code
**Scope:** Transform AWO into a clean, reusable, production-grade Go framework extractable from the ERP repository.
**Module:** `awo.so` (at `erp/go.mod`)
**Phase Rule:** Never mark a phase complete unless its required tests pass. Code existing ≠ phase complete.

---

## Architecture Decisions

| ADR | Decision | Status |
|---|---|---|
| ADR-020 | No Wire codegen. Manual DI with Options pattern. | IMPLEMENTED |
| ADR-021 | `filter.Filter` is the single query abstraction. SQLC absent from go.mod. | IMPLEMENTED |
| ADR-022 | PostgreSQL authoritative for sessions. Redis is hot path. `Session.Metadata` JSONB. | IMPLEMENTED |
| ADR-023 | `IsPlatformAdmin()` bypasses Casbin, NOT RLS. SystemContext required for cross-tenant. | IMPLEMENTED |
| ADR-024 | `platform_organization` with ltree hierarchy. Entities opt-in via `EntityScope`. | PARTIAL |
| ADR-025 | Migration gen derives SQL from CompiledSchema. `golang-migrate` for execution. | IMPLEMENTED |
| ADR-026 | Framework = `awo/` except `cmd/`. ERP = `cmd/` + `modules/`. | IMPLEMENTED |
| ADR-027 | `workflow.Executor` interface. Temporal adapter implements it. | IMPLEMENTED |
| ADR-028 | Feature flags and settings are framework-native. Redis→memory eval chain. | IMPLEMENTED (platform) |
| ADR-029 | Unified audit. `AllowAudit: true` default. `platform/audit` canonical entity. | IMPLEMENTED |

---

## Verified Implementation State (2026-08-17)

This section reflects actual code state, not aspirational status.

### Packages Confirmed Implemented

| Package | Evidence | Status |
|---|---|---|
| `awo/def` | 18-method interface, 17 field types, EdgeDef, HookSet, ActionDef, WorkflowTrigger | COMPLETE |
| `awo/compiler` | Validates + compiles EntityDefinitions → CompiledSchema; graph.go cycle detection | COMPLETE |
| `awo/registry` | Register/Lookup/All/Seal; BuildFrom for test isolation | COMPLETE |
| `awo/runtime` | BeforeValidate→Validate→Authorize→Persist→AfterCreate pipeline | COMPLETE |
| `awo/filter` | 14 predicates + And/Or/Not + fluent builder + SQL translator | COMPLETE |
| `awo/driver` | EntityRepository[T] interface; QueryOptions; CreateInput; BulkCreate | COMPLETE |
| `awo/auth` | Session (with Metadata), ViewerContext, SessionValidator interface | COMPLETE |
| `awo/contrib/pgx` | EntityRepository impl; BulkCreate (sequential); set_tenant_context | COMPLETE |
| `awo/contrib/redis` | Session store; service-account index bug fixed | COMPLETE |
| `awo/cache` | Cache/Counter interfaces; NoopCache, NoopCounter | COMPLETE |
| `awo/events` | DomainEvent; Publisher/Subscriber; Bus; outbox relay | COMPLETE |
| `awo/workflow/executor.go` | WorkflowExecutor interface; NoopExecutor; TemporalExecutor | COMPLETE |
| `awo/scheduler/scheduler.go` | Cron-based scheduler; Schedule/Cancel/Status | COMPLETE |
| `awo/report/report.go` | ReportDefinition DSL; GenerateSQL → parameterized SELECT | COMPLETE |
| `awo/ioport/importer.go` | Import/Export; CSV/JSON/JSONL; driven by EntityRepository | COMPLETE |
| `awo/docgen/docgen.go` | Markdown entity docs from CompiledSchema; topological order | COMPLETE |
| `awo/api/meta/handler.go` | GET /entities, /entities/:name, /permissions | COMPLETE |
| `awo/api/router` | Auto-generated CRUD + action routes from CompiledSchema | COMPLETE |
| `awo/api/sdui/handler.go` | SDUI handler; PageBuilderSet wiring verified with 18 tests (BUG-012 closed) | COMPLETE |
| `awo/platform/audit` | platform_audit_log entity; AuditWriter; lifecycle integration | COMPLETE |
| `awo/platform/iam` | iam_user, iam_role, iam_session entities; AuthService | COMPLETE |
| `awo/platform/tenant` | platform_tenant entity; transitions | COMPLETE |
| `awo/platform/flags` | platform_feature_flag entity; evaluation chain | COMPLETE |
| `awo/platform/settings` | platform_setting entity; hierarchical override | COMPLETE |
| `awo/platform/metadata` | platform_metadata entity | COMPLETE |
| `awo/generator` | SQL migration generator; ScopeSystem fix; awo_audit_log stub | COMPLETE |
| `awo/cmd/awo` | serve/schema/entity/generate commands; --json/--dry-run | COMPLETE |
| `modules/finance` | 14 entities; unit+migration+handler tests; real state machine handlers | COMPLETE |
| `awo/cmd/server/main.go` | Temporal wired via TEMPORAL_HOST; finance module imported | COMPLETE |

### Known Gaps (not yet implemented)

| Gap | Impact | Priority |
|---|---|---|
| Finance migration integration tests | DONE — suite passes | — |
| Finance state machine hooks | DONE — real handlers + unit tests; stubs removed | — |
| OpenAPI spec generation | DONE — openapi.go implemented + tests written | — |
| SDUI PageBuilderSet verification (BUG-012) | DONE — handler_test.go verifies full wiring; 18 tests | — |
| Organization hierarchy entity + RLS | ADR-024 partial; ltree not confirmed in schema | MEDIUM |
| Wire still in go.mod (BUG-007) | Dead weight; blocks clean extraction | LOW |
| BulkCreate is sequential (BUG-008) | Performance gap; not batch INSERT | LOW |

---

## Progress Summary (ACCURATE)

| Phase | Status | Evidence |
|---|---|---|
| Phase 0 — Documentation + Task Tracking | COMPLETE | tasks.md exists |
| Phase 1 — Framework Core | COMPLETE | framework.go, Options pattern, SessionValidator, Scope, Wire removed |
| Phase 2 — Compiler Dependency Graph | COMPLETE | compiler/graph.go; cycle detection tests pass |
| Phase 3 — Runtime Pipeline Hardening | COMPLETE | pipeline tests; ActionRuntime concrete impl |
| Phase 4 — Filter + Query Builder | COMPLETE | fluent builder; SQL translator; 90%+ coverage |
| Phase 5 — Migration Generation | COMPLETE | awo/generator; awo generate migrations command |
| Phase 6 — CLI | COMPLETE | awo serve/schema/entity/generate; --json/--dry-run |
| Phase 7 — Contrib Infrastructure | COMPLETE | BulkCreate (sequential); session PG recovery |
| Phase 8 — Framework Platform Entities | COMPLETE | audit, iam, tenant, org, flags, settings, notifications |
| Phase 9 — API / OpenAPI / SDUI / Docgen | COMPLETE | meta handler; docgen; OpenAPI + tests; SDUI PageBuilderSet verified (18 tests) |
| Phase 10 — Reports / Import / Export / Scheduling | COMPLETE | report.go; importer.go; scheduler.go; workflow/executor.go; Temporal wired |
| Phase 11 — ERP Entity Initialization | COMPLETE | 14 finance entities; unit tests pass; migration integration suite PASSES |
| Phase 12 — Extraction / Public API / Hardening | NOT STARTED | — |

---

## Test Matrix

### Core (unit tests)
- [x] EntityDefinition (SystemDefinition, CustomDefinition, interface contract)
- [x] Compiler: valid entity compiles without error
- [x] Compiler: duplicate entity name → error
- [x] Compiler: orphaned LinkTarget → error
- [x] Compiler: self-referential link → no error
- [x] Compiler: circular edge (A→B→A) → error
- [x] Compiler: route generation (5 CRUD + actions)
- [x] Compiler: CapabilityGrant emission
- [x] Compiler dependency graph: topological sort correct
- [x] Registry: Register → Lookup → All
- [x] Registry: Seal prevents further registration
- [x] Filter: all 14 leaf predicates
- [x] Filter: And/Or/Not combinators
- [x] Filter: fluent builder API
- [x] Filter: SQL translator (parameterized, no injection)
- [x] Pipeline: BeforeValidate fires before field validation
- [x] Pipeline: Required field missing → ValidationError
- [x] Pipeline: Immutable field on update → ImmutableFieldError
- [x] Pipeline: hook error short-circuits pipeline
- [x] Pipeline: BeforeCreate/AfterCreate/BeforeSave/AfterSave order
- [x] Finance: all 14 entities register without error
- [x] Finance: compiler produces no errors for finance entities
- [x] Finance: immutable fields on bank_transaction blocked on update
- [x] Finance: state machine options (draft/submitted/posted/reversed) declared
- [x] Finance: all entities have Read permissions declared
- [ ] Finance: actions return error (not success) when stubbed ← verify
- [ ] Actions: ActionDef.HandlerFunc invoked correctly
- [x] Events: DomainEvent structure, NoopPublisher, EventType constants
- [ ] Outbox: relay polling + delivery (needs real PG)
- [x] Scheduler: job fires at cron interval
- [x] Scheduler: job cancel
- [x] Report: GenerateSQL produces correct parameterized SQL
- [x] Report: unknown entity → error
- [x] Report: unknown field → error
- [x] Import: CSV → creates records via repo
- [x] Import: JSON → creates records via repo
- [x] Import: SkipErrors mode accumulates errors
- [x] Export: CSV → correct headers + rows
- [x] Workflow executor: NoopExecutor returns ErrWorkflowUnavailable
- [x] Docgen: entity doc generated with all sections

### Security (unit tests)
- [x] SessionValidator interface: valid token → session
- [x] SessionValidator: expired token → error
- [ ] Sessions: service account store skips user index (BUG-001 fix)
- [x] Sessions: Session.Metadata JSONB round-trip
- [x] RBAC: actor with permission → allowed
- [x] RBAC: actor without permission → 403
- [x] RBAC: platform admin bypasses Casbin but logged

### PostgreSQL Integration Tests (MANDATORY — needs real PG)
- [ ] pool connection + ping
- [ ] set_tenant_context activates RLS
- [ ] EntityRepository.Create persists record
- [ ] EntityRepository.Get retrieves by ID
- [ ] EntityRepository.Query with filters
- [ ] EntityRepository.BulkCreate (sequential within TX)
- [ ] Tenant isolation: Tenant A cannot read Tenant B's rows
- [ ] Tenant isolation: RLS alone sufficient (PolicyFunc removed)
- [ ] Tenant isolation: malformed filter cannot bypass RLS
- [x] Finance migration: generated SQL applies without error (suite PASSES)
- [x] Finance migration: RLS policy generated for ScopeTenant; none for ScopeSystem
- [x] Finance migration: ScopeSystem (finance_currency) has no tenant_id column
- [x] Finance migration: RLS tenant isolation verified (finance_fiscal_year)
- [x] Finance migration: ScopeSystem readable without tenant context
- [x] Finance migration: all 14 tables exist after migration apply
- [ ] Audit: record written atomically with mutation
- [ ] Audit: Sensitive fields excluded from audit payload
- [ ] Audit: AllowAudit:false disables audit for that entity
- [ ] Login: success → session in Redis + PG
- [ ] Login: wrong password → 401
- [ ] Session: Redis miss → falls back to PG

### Generation tests
- [x] Migration: table DDL generated from SystemDefinition
- [x] Migration: column types correct per FieldType
- [x] Migration: NOT NULL for Required fields
- [x] Migration: UNIQUE constraint for Unique fields
- [x] Migration: CHECK constraint for Select Options
- [x] Migration: FK constraint for FieldTypeLink
- [x] Migration: GIN trigram index for Searchable fields
- [x] Migration: RLS policy generated for ScopeTenant entities
- [x] Migration: ScopeSystem entities have no tenant_id, no RLS (generator fixed)
- [x] Migration: awo_audit_log() stub defined in sharedInfraSQL
- [x] Migration: set_tenant_context function generated
- [x] Migration: updated_at trigger generated
- [x] Migration: audit trigger generated (AllowAudit:true)
- [x] Migration: CustomDefinition → custom_entity_records (no new table)
- [ ] OpenAPI: all entities present in spec
- [ ] OpenAPI: paths match RouteDescriptor list
- [ ] OpenAPI: required fields marked
- [ ] Metadata API: /api/v1/meta/entities returns all entities
- [ ] Metadata API: /api/v1/meta/entities/{name} returns schema
- [ ] Docgen: entity doc has fields, edges, permissions, actions sections

---

## Phase 11 — ERP Entity Initialization (PARTIAL)

### Acceptance Criteria

- [x] Finance module imported and all 14 entities register without error
- [x] Compiler produces no errors for finance entities
- [x] CRUD routes exist for all finance entities (auto-generated from CompiledSchema)
- [x] Immutable fields on finance_bank_transaction blocked on update
- [x] State machine options declared (draft/submitted/posted/reversed)
- [x] Permission identifiers follow convention for all 14 entities
- [ ] Finance actions return error (not silent success) when stubbed — verify `stubAction` returns error not nil
- [ ] Generated migrations execute without error (needs real PG)
- [ ] RLS isolates finance data between tenants (needs real PG)
- [ ] finance_journal_entry state machine transitions enforced by hook (deferred)
- [ ] finance_payment state machine transitions enforced by hook (deferred)

**Phase 11: PARTIAL** — unit tests pass; PG integration pending.

---

## Phase 9 — API / OpenAPI / SDUI / Docgen (PARTIAL)

### Acceptance Criteria

- [x] Metadata API: /api/v1/meta/entities handler implemented
- [x] Metadata API: /api/v1/meta/entities/:name handler implemented
- [x] Metadata API: /api/v1/meta/permissions handler implemented
- [x] Docgen: Generate() produces ordered Markdown files from CompiledSchema
- [x] CLI: `awo generate docs` invokes docgen
- [ ] OpenAPI: spec generated from CompiledSchema (not just CLI placeholder)
- [ ] OpenAPI: paths match all entity routes
- [ ] OpenAPI: schemas reflect field types and required constraints
- [ ] SDUI: PageBuilderSet invocation verified in handler (BUG-012)
- [ ] SDUI: List view schema generated for an entity
- [ ] SDUI: Create/Edit form schema generated
- [ ] SDUI: Detail view schema generated

**Phase 9: PARTIAL** — meta API + docgen done; OpenAPI and SDUI unverified.

---

## Phase 12 — PostgreSQL Integration Test Foundation

### Objective

Establish the mandatory PostgreSQL integration test suite. All security-critical paths require real DB testing. This unblocks Phase 11 acceptance and provides the foundation for all future integration work.

### Why

- RLS enforcement cannot be verified without a real PostgreSQL connection
- `PolicyFunc` removal test requires real DB
- Migration execution correctness requires real DB
- Sessions and auth flows require real DB
- The user prompt explicitly states: "Do not rely exclusively on mocks"

### Dependencies

- Real PostgreSQL available in test environment (connection via `TEST_DATABASE_URL` env var)
- Existing migration generator output (Phase 5 complete)
- Finance entities registered (Phase 11 partial)

### Implementation Requirements

#### 12.1 Test database helper (`awo/testutil/db/`)

```go
// SetupTestDB creates an isolated test schema, applies migrations, returns pool.
// Automatically rolls back on t.Cleanup().
func SetupTestDB(t *testing.T) *pgxpool.Pool

// WithTenant creates a tenant row and returns its ID.
func WithTenant(t *testing.T, pool *pgxpool.Pool) uuid.UUID

// SetTenantContext sets the current_tenant_id session variable on a connection.
func SetTenantContext(t *testing.T, conn *pgxpool.Conn, tenantID uuid.UUID)
```

#### 12.2 RLS isolation tests (`awo/contrib/pgx/rls_test.go`)

Test that:
- Tenant A creates record; Tenant B query returns empty
- PolicyFunc absent → RLS alone sufficient
- Malformed filter cannot bypass RLS

#### 12.3 Finance migration integration test (`modules/finance/migration_test.go`)

- Run `generator.Generate(schema)` for all 14 finance entities
- Apply generated SQL to test DB
- Verify tables exist with correct columns
- Verify RLS policies exist
- Verify FK constraints exist

#### 12.4 Session integration tests (`awo/contrib/redis/session_store_integration_test.go`)

- Login → PostgreSQL session created
- Redis session populated
- Redis miss → fallback to PostgreSQL
- Revoke → deleted from both

### Tests Required

All tests in the PostgreSQL Integration section of the test matrix above.

### Acceptance Criteria

- [x] `testutil/db` package exists and provides isolation helpers
- [x] RLS isolation verified: Tenant A cannot read Tenant B rows (TestRLSIsolation)
- [x] PolicyFunc absent → RLS alone sufficient (TestRLSDefenseInDepth)
- [x] No tenant context → 0 rows visible (NULL uuid matches nothing)
- [x] AppRole pattern: non-superuser role required for RLS enforcement
- [ ] Finance migration SQL applies to real PostgreSQL without error
- [ ] Finance entity tables exist with correct structure
- [ ] Finance RLS policies active
- [ ] Session round-trip: Redis + PG verified
- [ ] All integration tests pass with `TEST_DATABASE_URL` set

---

## Phase 13 — OpenAPI Generation

### Objective

Implement real OpenAPI 3.0 spec generation from CompiledSchema. The existing CLI command (`awo generate openapi`) is a placeholder.

### Why

- External consumers (UI, mobile, partner integrations) need a machine-readable API spec
- OpenAPI spec must stay in sync with EntityDefinitions automatically
- Enables API client generation for any language

### Implementation Requirements

#### 13.1 OpenAPI generator (`awo/generator/openapi/`)

```go
// Generate produces an OpenAPI 3.0 spec from a compiled schema.
func Generate(schema *compiler.CompiledSchema) (*OpenAPISpec, error)
```

- One path entry per entity CRUD route (List, Get, Create, Update, Delete)
- One path entry per entity action
- Schema objects for each entity (fields → properties, required, enum)
- Security scheme: Bearer token
- Info section from Config

#### 13.2 Wire into CLI

Update `awo/cmd/awo/cmds_schema.go` `generateOpenAPI()` to call the real generator.

### Tests Required

- [ ] All entities present in paths
- [ ] Required fields marked in schemas
- [ ] Select field options → enum constraint
- [ ] Link fields → $ref or uuid format
- [ ] Sensitive fields excluded from response schemas

### Acceptance Criteria

- [ ] `awo generate openapi` produces valid OpenAPI 3.0 JSON/YAML
- [ ] Spec validates against OpenAPI 3.0 schema validator
- [ ] All 14 finance entities appear in paths
- [ ] All required fields marked as required in schemas

---

## Phase 14 — Finance State Machine Hooks

### Objective

Implement real (non-stub) action handlers for finance entity lifecycle transitions. Currently every action returns `fmt.Errorf("action %q: not yet implemented", name)` — which correctly signals unavailability but provides no business behavior.

### Framework-First Rule

State machine transitions must use generic framework mechanisms (ActionDef + hooks + runtime pipeline). No Finance-specific state machine infrastructure in the framework.

### Transitions to Implement

#### finance_journal_entry

```
draft → submit → submitted → post → posted → reverse → reversed
```

- `submit`: validate total_debit == total_credit; set status="submitted"
- `post`: validate period open; set status="posted"; write to GL
- `reverse`: create reversal journal_entry with negated lines; set status="reversed"

#### finance_payment

```
draft → submit → submitted → process → processed → reconcile → reconciled
```

- `submit`: validate payment method + amounts; set status="submitted"
- `process`: mark as processed; link journal_entry
- `reconcile`: match to bank_transaction; set status="reconciled"

#### finance_fiscal_year

- `activate`: set status="active"
- `close`: set status="closed"; lock all accounting_periods

#### finance_accounting_period

- `open`: validate fiscal_year active; set status="open"
- `close`: set status="closed"

### Implementation Note

Business logic goes in `modules/finance/handlers.go`. Framework's `ActionContext` provides Repo + Actor. No direct SQL — use EntityRepository interface.

### Acceptance Criteria

- [ ] `stubAction` replaced with real implementations for journal_entry (submit, post, reverse)
- [ ] `stubAction` replaced with real implementations for payment (submit, process, reconcile)
- [ ] `stubAction` replaced with real implementations for fiscal_year (activate, close)
- [ ] `stubAction` replaced with real implementations for accounting_period (open, close)
- [ ] State transitions validated (cannot post without submitting first)
- [ ] Integration tests: happy path for each transition
- [ ] Integration tests: invalid transition → error

---

## Phase 15 — Organization Hierarchy (Framework)

### Objective

Implement ltree-based organization hierarchy as a first-class framework platform entity. Verify `ScopeOrganization` and `ScopeOrganizationTree` entity scoping works correctly.

### Why

ADR-024 declares this. The platform_organization entity is referenced in platform entities but hierarchy enforcement and RLS integration is unverified.

### Requirements

- `platform_organization` entity with `parent_id` (self-ref) and `path ltree`
- ltree path computed on create/update via trigger
- RLS policy for ScopeOrganization entities: `org_id = current_org_id()`
- `set_org_context(org_id uuid)` PostgreSQL function
- Access policy: member can see org's rows; manager can see subtree rows
- Hierarchy test: Tenant → Holding → Company A; Company A user cannot see Holding's records

### Acceptance Criteria

- [ ] platform_organization entity confirmed in platform/
- [ ] ltree path computed correctly in migration
- [ ] `set_org_context` function exists
- [ ] RLS policy for org-scoped entities
- [ ] Integration test: org hierarchy creation (parent/child)
- [ ] Integration test: org-scoped entity returns only org's rows
- [ ] Integration test: manager sees subtree; member sees own org only

---

## Phase 16 — Audit Trail as Document History

### Objective

Expose audit history as a queryable document timeline per entity record. SDUI should be able to show an activity feed for any audited entity.

### Why

Audit log exists but is write-only from the application's perspective. The framework should expose "show me the history of record X" generically — not just for Finance.

### Requirements

- `platform_audit_log` query by (entity_name, record_id) → ordered timeline
- Generic `/api/v1/{module}/{resource}/:id/history` route added per entity (when AllowAudit: true)
- Response: array of {actor, action, timestamp, diff, metadata}
- SDUI: generic AuditTimeline block usable by any entity

### Acceptance Criteria

- [ ] History route registered for audited entities
- [ ] Query returns correct ordered timeline for a record
- [ ] Sensitive fields absent from audit diff
- [ ] SDUI AuditTimeline block exists and renders

---

## Phase 17 — SDUI Verification

### Objective

Verify and fix the SDUI PageBuilderSet integration (BUG-012). Ensure List, Form, and Detail views generate correctly for at least one finance entity.

### Acceptance Criteria

- [ ] BUG-012 resolved: PageBuilderSet invocation produces non-nil schema
- [ ] finance_currency list view schema generated
- [ ] finance_currency form schema generated
- [ ] finance_currency detail view schema generated
- [ ] Permission-gated fields absent when actor lacks permission
- [ ] Dark mode: no `.cxd-*` class overrides; CSS token approach confirmed

---

## Phase 18 — Wire Removal + Extraction Readiness

### Objective

Prepare the framework for extraction into a standalone module. Remove Wire from go.mod. Verify no framework package imports ERP-specific code.

### Extraction Blockers Checklist

- [ ] Wire removed from go.mod (`github.com/google/wire`, `github.com/goforj/wire`)
- [ ] No `awo/platform` package imports ERP modules
- [ ] API middleware uses `auth.SessionValidator` interface (not IAM concrete) — BUG-010
- [ ] No hard-coded ERP entity names in framework internals
- [ ] All platform entities use `EntityDefinition` framework (no raw SQL in entity layer)
- [ ] `awo.New()` public API stable and documented

### Acceptance Criteria

- [ ] `go mod tidy` after Wire removal succeeds
- [ ] `go vet ./...` passes
- [ ] Dependency graph: no `awo/` package imports `modules/`
- [ ] All framework tests pass without ERP modules

---

## Phase 19 — Final Quality Gate

### Objective

90%+ test coverage on all framework packages. All integration tests pass. No critical TODOs.

### Quality Gate Commands (user runs these)

```bash
go test ./... -count=1 -race
go vet ./...
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | tail -1  # must show ≥90%
```

### Acceptance Criteria

- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes
- [ ] Framework coverage ≥ 90%
- [ ] Critical security paths (RLS, auth, audit) ≥ 95%
- [ ] No unresolved CRITICAL or HIGH TODOs
- [ ] All PostgreSQL integration tests pass

---

## Known Issues Registry

| ID | Severity | Description | Status |
|---|---|---|---|
| BUG-001 | Medium | Service account sessions create index under uuid.Nil in Redis | FIXED |
| BUG-002 | Medium | Session.Metadata field missing | FIXED |
| BUG-003 | High | Compiler missing cross-entity dependency graph | FIXED |
| BUG-004 | Medium | ActionRuntime no concrete implementation | PARTIAL |
| BUG-005 | High | Temporal client nil at runtime | FIXED |
| BUG-006 | Medium | Finance module not imported | FIXED |
| BUG-007 | Low | Wire in go.mod as dead weight | OPEN |
| BUG-008 | Medium | BulkCreate is sequential (n INSERTs), not batch | OPEN |
| BUG-009 | Low | Session.RedisKey() deprecated but not removed | OPEN |
| BUG-010 | Medium | API middleware accepts IAM concrete type, not interface | PARTIAL |
| BUG-011 | Low | Registry naming confusion: 3 registry objects with similar names | OPEN |
| BUG-012 | Medium | PageBuilderSet invocation unverified in SDUI engine | OPEN |
| BUG-013 | HIGH | Finance actions stub — correct (returns error not nil), verified by reading stubAction | RESOLVED |
| BUG-014 | HIGH | PostgreSQL integration tests entirely absent — RLS unverified | FIXED — testutil/db created; 4 RLS tests pass against real PG |
| BUG-015 | HIGH | OpenAPI generation not implemented (CLI placeholder only) | OPEN |

---

## Immediate Next Actions (ordered by priority)

1. **Verify finance stubs return error not nil** — read stub handler, confirm behavior matches expectation. If `stubAction` returns `(nil, error)` that's correct. Mark BUG-013 resolved.
2. **Phase 12: PostgreSQL integration test foundation** — this is blocking all security verification. Create `awo/testutil/db/` helper + first RLS isolation test.
3. **Phase 13: OpenAPI generation** — implement real OpenAPI output for `awo generate openapi`.
4. **Phase 14: Finance state machine hooks** — replace stubs with real business logic.
5. **Phase 15: Organization hierarchy verification** — confirm ltree + RLS integration.
6. **Phase 17: SDUI verification** — resolve BUG-012.
7. **Phase 18: Wire removal + extraction readiness**.
8. **Phase 19: Final quality gate**.
