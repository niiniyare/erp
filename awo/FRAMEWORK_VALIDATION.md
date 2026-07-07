# Awo Framework Validation

This document summarises what the validation test suite exercises, which components
are confirmed working, and what remains to be validated in future milestones.

---

## Execution Pipeline (validated end-to-end)

```
EntityDefinition
    │  def.SystemDefinition{...}
    ▼
registry.BuildFrom(defs)
    │  validates names, required fields, link targets, module tags
    ▼
compiler.Compile(reg)
    │  resolves links, emits routes, emits Casbin policies, computes defaults
    ▼
compiler.Fingerprint(schema)
    │  SHA-256 of deterministic JSON serialisation — stable across restarts
    ▼
fakestore / pgx driver
    │  Create → hook pipeline → persist → return EntityRecord
    ▼
filter.Filter evaluation
    │  Eq / Contains / And / Or / Not predicates applied in-memory
    ▼
Fiber HTTP routing (stub)
    │  all compiled routes registered — correct status codes and JSON envelopes
    ▼
introspect.Handler
    │  /api/introspect → SchemaInfo JSON (entities, routes, fingerprint)
    ▼
docgen.WriteMarkdown
       Markdown entity reference with fields, permissions, routes
```

---

## Test Suite Layout

| Package | File | What it validates |
|---------|------|-------------------|
| `awo/tests/integration` | `registry_compiler_test.go` | Registry build, module lookup, duplicate detection, compiler compile, diagnostics, fingerprint determinism, route emission |
| `awo/tests/integration` | `pipeline_test.go` | Full CRUD pipeline (Create/Get/Update/Delete/Query), filter predicates, Exists/Count, hook validation, BulkCreate/BulkUpdate, WithTx, TenantContext propagation, org scope isolation |
| `awo/tests/integration` | `db_test.go` *(build tag: `integration`)* | Bootstrap against real PostgreSQL + Redis, TenantContext carry, org scope interface contract, demo schema presence, Casbin policies generated |
| `awo/tests/e2e` | `api_test.go` | Fiber route registration (all compiled routes ≥5), introspect endpoint, list/create/get/update/delete HTTP verbs, fingerprint stability across server instances |
| `awo/tests/bench` | `bench_test.go` | RegistryBuild, CompilerCompile, CompilerFingerprint, FakeStoreCreate, FakeStoreQuery (1 000 records), FakeStoreGet, FilterEval (500 records, compound filter) |

### Run commands

```bash
# Unit + integration tests (no DB required)
go test ./awo/... -cover

# Benchmarks
go test -bench=. -benchmem ./awo/tests/bench/...

# DB integration tests (requires PostgreSQL + Redis)
AWO_TEST_DATABASE_URL=postgresql://user:pass@localhost:5432/awo_test \
AWO_TEST_REDIS_URL=redis://localhost:6379/1 \
go test -tags=integration -run TestDB ./awo/tests/integration/...
```

---

## Components Confirmed Working (no DB required)

| Component | Validated by |
|-----------|-------------|
| `registry.BuildFrom` | `TestRegistryBuild`, `BenchmarkRegistryBuild` |
| `registry.ByModule` | `TestRegistryModuleLookup` |
| `registry.BuildFrom` — duplicate dedup | `TestRegistryDuplicateDetection` |
| `registry.BuildFrom` — no mandatory check | `TestRegistryBuildFrom_NoMandatoryCheck` |
| `compiler.Compile` — entity/field/edge/action | `TestCompilerCompile` |
| `compiler.Validate` — diagnostics | `TestCompilerDiagnostics` |
| `compiler.Fingerprint` — SHA-256, deterministic | `TestCompilerFingerprint`, `TestSchemaFingerprint` |
| `compiler` — route emission (5 CRUD per entity) | `TestCompilerRoutes`, `TestAllEntitiesHaveRoutes` |
| `compiler` — Casbin policy emission | `TestDBCRUDPipeline` (schema + policies check) |
| `fakestore.Create` + hook pipeline | `TestCRUDCreate`, `BenchmarkFakeStoreCreate` |
| `fakestore.Get` | `TestCRUDGet`, `BenchmarkFakeStoreGet` |
| `fakestore.Update` | `TestCRUDUpdate` |
| `fakestore.Delete` | `TestCRUDDelete` |
| `fakestore.Query` with filter | `TestQueryFilter`, `BenchmarkFakeStoreQuery` |
| `fakestore.Exists` / `Count` | `TestExistsCount` |
| `fakestore.BulkCreate` | `TestBulkCreate` |
| `fakestore.BulkUpdate` | `TestBulkUpdate` |
| `fakestore.WithTx` | `TestWithTx` |
| `filter.Eq`, `Contains`, `And`, `Or`, `Not` | `TestQueryFilter`, `BenchmarkFilterEval` |
| Hook: `BeforeCreate` — validation | `TestHookValidation` |
| `runtime/tenant.WithContext` / `FromContext` | `TestTenantContextPropagation` |
| `platform/organization` — ViewerContext | `TestOrgScopeIsolation` |
| `platform/organization.VisibilitySelf` | `TestOrgScopeIsolation` |
| `platform/organization.VisibilityEntireTenant` | `TestOrgScopeIsolation` |
| `introspect.Inspect` / `introspect.Handler` | `TestIntrospectEndpoint` |
| Fiber route registration (all verbs) | `TestRouteRegistration`, `TestEntity*Route` |
| `docgen.WriteMarkdown` | `TestDocgenMarkdown` (pipeline_test.go) |
| `harness.New` — full in-memory environment | All pipeline tests |

---

## Demo Module (`awo/examples/demo`)

`demo.CustomerDefinition` drives all framework layers simultaneously:

| Layer | Validated |
|-------|-----------|
| Registry — name, module, field list | ✓ |
| Compiler — RequiredFields, ImmutableFields, DefaultValues | ✓ |
| Compiler — routes emitted | ✓ |
| Compiler — Casbin policies emitted | ✓ |
| Hook — `CustomerValidator.BeforeCreate` rejects empty name | ✓ |
| Hook — `CustomerValidator.BeforeCreate` rejects empty code | ✓ |
| Hook — `CustomerValidator.BeforeCreate` rejects invalid email | ✓ |
| fakestore CRUD round-trip | ✓ |
| Introspect — appears in `/api/introspect` output | ✓ |
| Fiber — all 5 CRUD routes registered | ✓ |
| DB migration — `demo_customer` table + RLS + indexes | migration file present |

---

## Remaining Gaps (future milestones)

### v1.1 — pgx Driver

The production persistence layer (`awo/driver/pgx`) is not yet implemented.
Until then:

- DB CRUD validation is annotated as pending in `TestDBCRUDPipeline`
- E2E routes are backed by `fakestore`, not PostgreSQL
- `set_tenant_context()` is not wired into the pgx connection lifecycle
- RLS enforcement is asserted at context level only, not SQL level

### v1.2 — Casbin Enforcement

`schema.CasbinPolicies` are generated and validated in tests, but the Casbin
enforcer is not wired into the Fiber middleware pipeline. Permission gates are
structural stubs only.

### v1.3 — Temporal Worker

Workflow triggers declared in `WorkflowTriggers` are compiled into `EntityDefinition`
but `StartWorkflow` calls are not executed in any test. Temporal integration
requires a running Temporal server or `testsuite.WorkflowTestSuite`.

### v1.4 — SDUI Generation

`PageBuilderSet` compilation and the amis JSON schema output are not exercised
in this test suite. Requires `awo/sdui` package implementation.

### v1.5 — OpenAPI Generation

`GET /api/openapi.json` is not yet emitted. Requires `awo/openapi` package.

### Staging validation (pre-production checklist)

- [ ] Run `go test -tags=integration ./awo/tests/integration/...` against staging DB
- [ ] Verify `TestDBTenantRLS` SQL-level isolation once pgx driver lands
- [ ] Run `go test -race ./awo/...` — no data races under concurrent load
- [ ] Load test fakestore benchmarks against production-comparable data sizes
- [ ] Verify `compiler.Fingerprint` stability across Go version upgrades

---

## Architecture Exercised

```
┌─────────────────────────────────────────────────────┐
│  EntityDefinition (def.SystemDefinition)             │
│    Fields / Edges / Hooks / Permissions / Actions    │
└──────────────────────┬──────────────────────────────┘
                       │ registry.BuildFrom
┌──────────────────────▼──────────────────────────────┐
│  Registry (*registry.Registry)                       │
│    Lookup / ByModule / All                           │
└──────────────────────┬──────────────────────────────┘
                       │ compiler.Compile
┌──────────────────────▼──────────────────────────────┐
│  CompiledSchema                                      │
│    EntitySchema / RouteDescriptor / CasbinPolicy     │
│    Fingerprint (SHA-256)                             │
└────┬──────────┬──────────────┬───────────────────────┘
     │          │              │
     │          │              │ introspect.Inspect
     │          │              ▼
     │          │        SchemaInfo → /api/introspect
     │          │
     │          │ docgen.WriteMarkdown
     │          ▼
     │    Markdown API reference
     │
     │ wireStubRoutes → fiber.App
     ▼
Fiber HTTP Router (all CRUD routes registered)
     │
     ▼
fakestore.Store (in-memory driver.RecordRepository)
     │
     ├── Create → BeforeCreate hooks → persist
     ├── Get / Query / Exists / Count
     ├── Update → persist
     ├── Delete
     ├── BulkCreate / BulkUpdate
     └── WithTx (atomic commit/rollback)
```

All boxes above are exercised without PostgreSQL, Redis, Temporal, or Casbin.
