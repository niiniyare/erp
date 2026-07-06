# Awo Framework — ROADMAP

## v1.0 Readiness Criteria

A release is considered v1.0 when every item in this checklist is satisfied.

### Core Framework (all complete)

- [x] `awo/def` — EntityDefinition, FieldDef, EdgeDef, ActionDef, HookSet, PermissionSet, WorkflowTrigger
- [x] `awo/registry` — Build, BuildFrom, validateDefinition (mandatory-system-entity rules)
- [x] `awo/compiler` — Compile, Validate (diagnostics), Fingerprint, DepGraph (Kahn topo-sort)
- [x] `awo/filter` — full predicate DSL (Eq, Neq, Gt, Gte, Lt, Lte, In, NotIn, IsNull, IsNotNull, Contains, StartsWith, EndsWith, Between, And, Or, Not)
- [x] `awo/driver` — EntityRepository[T] generic interface, RecordRepository alias
- [x] `awo/migration` — Scan, BuildPlan, VerifyChecksums
- [x] `awo/module` — Manifest, Version (semver), ModuleRegistry (Kahn topo-sort + version validation)
- [x] `awo/workflow` — ExecuteActivity, Saga, SignalChannel, WorkflowTestEnv, Worker, BuildID/ParseID
- [x] `awo/sdui` — Builder (list/create/edit/detail pages), BuildNav, all 14+ field-type mappings
- [x] `awo/introspect` — Inspect, HTTP handler (schema JSON endpoint)
- [x] `awo/docgen` — WriteMarkdown (entity reference docs from CompiledSchema)
- [x] `awo/perf` — BenchmarkStore standard suite
- [x] `awo/runtime/tenant` — TenantContext propagation, WithContext, TryFromContext
- [x] `awo/testing/fakestore` — in-memory RecordRepository, full filter evaluation
- [x] `awo/testing/golden` — golden-file assertion helper
- [x] `awo/testing/harness` — Harness (Registry + Schema + Store + context helpers)
- [x] `awo/testing/conformance` — StoreSuite (11 sub-tests)

### Test Coverage (target: all public APIs exercised)

- [x] `compiler/depgraph_test.go`
- [x] `compiler/fingerprint_test.go`
- [x] `compiler/validate_test.go`
- [x] `migration/scanner_test.go`
- [x] `migration/planner_test.go`
- [x] `migration/checksum_test.go`
- [x] `module/manifest_test.go`
- [x] `module/registry_test.go`
- [x] `workflow/id_test.go`
- [ ] `filter/*_test.go` — Between, Contains, StartsWith, EndsWith filters not yet tested
- [ ] `sdui/builder_test.go` — field-type → control mapping
- [ ] `introspect/introspect_test.go`
- [ ] `docgen/docgen_test.go`
- [ ] `testing/conformance` — fakestore passes all 11 sub-tests (integration run needed)

### Before v1.0 Tag

1. **`go test ./awo/...` must pass** — run by developer; not auto-run by framework tooling.
2. **No `panic("not implemented")`** anywhere in framework packages — replace all with returning errors or documented TODOs.
3. **`awo/cmd/awo`** — `schema validate`, `schema fingerprint`, `docgen` commands fully wired to compiled binary via plugin or `go:generate` hook. Currently documented as requiring host binary.
4. **Temporal integration tested** — at minimum one end-to-end workflow test against `testsuite.WorkflowTestSuite`.
5. **`fakestore` passes `StoreSuite`** — TenantIsolation sub-test should be skipped (documented non-enforcement by design); all others must pass.
6. **`VerifyChecksums` called in migration runner** — the cmd/migrate entrypoint must verify checksums before executing any plan.
7. **`Fingerprint` used for Redis page-cache invalidation** — SDUI builder cache key must incorporate schema fingerprint.

---

## Platform Modules (Complete)

All 8 platform modules are implemented under `awo/platform/`. Each has: entity definitions, hooks, service layer, SQL migrations (.up.sql/.down.sql).

| Module | Package | Key Entities | Service Methods |
|---|---|---|---|
| Tenant | `platform/tenant` | platform_tenant, platform_org_unit, platform_branch | Create, Activate, Suspend, Archive |
| IAM | `platform/iam` | iam_user (mandatory), iam_role, iam_session, iam_api_token, iam_user_role | Login, ValidateToken, Logout |
| Audit | `platform/audit` | iam_audit_log (append-only) | Write, ComputeDiff |
| Settings | `platform/settings` | platform_setting | Get, GetString, Set (scope hierarchy) |
| Feature Flags | `platform/flags` | platform_feature_flag, platform_flag_tenant_override | IsEnabled, Invalidate |
| Metadata | `platform/metadata` | platform_custom_field | FieldsForEntity, AddField, DeactivateField |
| Module Registry | `platform/registry` | platform_module, platform_tenant_module | RegisterModule, Activate, Disable, ListActive |
| Notifications | `platform/notifications` | platform_notification, platform_notification_template | Send, MarkRead, UnreadCount, MarkAllRead |

Framework discoveries during platform implementation:
- `def.Filter` is `interface{}` — PolicyFunc returning nil is sufficient for admin-scoped entities
- `driver.BulkUpdate` with `driver.Patch{Set: ...}` is the correct API for MarkAllRead
- Template rendering belongs in the service layer, not in hooks (hooks must not call external services)
- Driver pattern (RegisterDriver) decouples delivery from entity persistence cleanly
- Notification delivery must be best-effort in AfterCreate hooks — Temporal activity preferred for durability

---

## v1.1 — Production Hardening

- [ ] `awo/driver/postgres` — production pgx RecordRepository implementation
- [ ] Row-level security wiring: `set_tenant_context()` called in every repository operation
- [ ] `awo/iam` — Casbin enforcer wired from `CasbinPolicies` in CompiledSchema
- [ ] `awo/api` — Fiber route registration from `Routes` in CompiledSchema
- [ ] `awo/cache` — Redis page-schema cache with Fingerprint-keyed invalidation
- [ ] `awo/outbox` — Temporal start-after-commit retry queue
- [ ] `awo/naming` — NamingSeries atomic counter (Redis + PostgreSQL sequence fallback)
- [ ] Audit log integration in repository layer (AfterSave hook writing to iam_audit_log)
- [ ] Notifications: wire TemporalDriver for at-least-once delivery guarantee
- [ ] platform/tenant service: wire EntityRepository (currently interface only)

## v1.2 — Developer Experience

- [ ] `awo/cmd/awo generate` — scaffold new module from template
- [ ] `awo/cmd/awo lint` — static check for CLAUDE.md rule violations (money-as-float, lazy-load, etc.)
- [ ] `awo/docgen` HTML output mode (beyond Markdown)
- [ ] `awo/sdui` — amis schema test against pinned SDK version

## v2.0 — Multi-Region / Scale

- [ ] Read-replica routing for Query/Count/Exists operations
- [ ] Per-tenant schema partitioning (investigate; defer unless > 10k tenants)
- [ ] Workflow history archival policy (Temporal namespace config)
- [ ] gRPC transport for inter-service entity reads (replaces HTTP CRUD for internal callers)

---

## Known Limitations (v1.0)

| Limitation | Impact | Mitigation |
|---|---|---|
| `awo/cmd/awo` sub-commands require compiled host binary | CLI not self-contained at framework level | Document in README; use `go generate` hooks |
| `fakestore` does not enforce tenant isolation | Test suites may not catch cross-tenant leaks | Skip TenantIsolation sub-test; use real Postgres for integration tests |
| `sdui.Builder` generates amis v2.x schemas — pinned SDK | Upgrading amis requires full audit | Never auto-update `web/sdk/`; audit before any upgrade |
| Temporal `ExecuteActivity` uses generic wrapper | Requires Go 1.21+ type inference | Documented in package-level comment |
| `compiler.Validate` does not check circular edge references | Self-referencing edges pass validation | Detected at DB level via FK constraints; add graph check in v1.1 |
