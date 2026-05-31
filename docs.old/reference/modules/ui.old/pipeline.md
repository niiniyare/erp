# UI Pipeline

The UI pipeline is a stage DAG driven by `pipeline.PipelineBuilder`. Each stage
is independently testable, declares its dependencies via `DependsOn()`, and
is wrapped by `InstrumentedStage` for tracing, metrics, and logging.

---

## Stage DAG

```
ui.session (10)
    │
    └──► ui.authz (20)
              │
              └──► ui.cache_lookup (30)
                        │
                        ├─── [CACHE HIT] ──────────────────► ui.response (90)
                        │
                        └──► ui.registry (40)
                                  │
                                  └──► ui.compile (50)
                                            │
                                            └──► ui.normalize (60)
                                                      │
                                                      └──► ui.validate (70)
                                                                │
                                                                └──► ui.cache_store (80)
                                                                          │
                                                                          └──► ui.response (90)
```

Numbers in parentheses are `Priority` values. Higher priority = runs later.

---

## Stage Reference

### `ui.session` (Priority 10, Required)

Reads `contract.SessionContext` from the Go context. Aborts with
`ErrUnauthenticated` if absent. The contract.InjectSessionContext() Fiber
middleware must be in the route chain before the handler.

**Sets:** nothing — downstream stages call `contract.FromContext()` directly.

---

### `ui.authz` (Priority 20, Required)

Calls `UIAuthzService.BulkEnforce()` to resolve all UI permissions via Casbin
in one batch. Computes stable SHA256 fingerprints for the permission set and
feature flag set.

**Sets:**
- `DataKeyUISession` — the fully populated `UISessionContext`
- `DataKeyPermFingerprint` — permission set fingerprint (hex string)
- `DataKeyFlagFingerprint` — feature flag fingerprint (hex string)

---

### `ui.cache_lookup` (Priority 30, Not Required)

Builds the generation-aware 8-component cache key and checks Redis + in-process
LRU. On hit, injects the cached schema and jumps directly to `ui.response`,
skipping registry → compile → normalize → validate → store.

A cache failure is non-fatal — the pipeline continues to the cold path.

**Sets on miss:**
- `DataKeyCacheKey` — computed key for CacheStoreStage
- `DataKeyCacheHit` = `false`

**Sets on hit:**
- `DataKeyCacheKey`, `DataKeyCacheHit` = `true`, `DataKeySchema`
- `StageResult.NextStageID` = `"ui.response"` (pipeline jump)

---

### `ui.registry` (Priority 40, Required)

Looks up the route in the page registry. Returns `ErrPageNotFound` (→ HTTP 404)
when no registration exists. Sets both `DataKeyPageFn` and `DataKeyASTPageFn`
when present in the registration.

**Sets:**
- `DataKeyASTPageFn` (when `PageRegistration.ASTFn != nil`)
- `DataKeyPageFn` (when `PageRegistration.Fn != nil`)

---

### `ui.compile` (Priority 50, Required)

Dispatches to `ASTPageFn` (preferred) or `PageFn` (legacy). Recovers from
panics in both paths.

**ASTPageFn path:** `fn(sess) → ast.Node → ast.CompileTree(node) → Schema`.
Sets `DataKeyASTCompiled = true`.

**PageFn path:** `fn(sess) → Schema` directly. NormalizeStage applies full
rule set.

**Sets:** `DataKeySchema`, optionally `DataKeyASTCompiled`

---

### `ui.normalize` (Priority 60, Required)

**Never returns an error.** Pure canonicalization:
- Lowercase component `type` field values
- Trim leading/trailing whitespace from API URLs

Skips nodes already produced by ASTPageFn (they are already canonical).

---

### `ui.validate` (Priority 70, Required)

Applies structural and security rules. Returns `*sharedErrors.BusinessError`
with `VALIDATE_*` code + `.WithCause(ui.ErrSchemaInvalid)` on violation.

**Structural rules** (skip when `DataKeyASTCompiled=true`):
- `VALIDATE_CRUD_SYNC_LOCATION` — CRUDNode must have `syncLocation: true`
- `VALIDATE_CHART_TRANSPARENT_BG` — ChartNode must have transparent background
- `VALIDATE_API_METHOD_PREFIX` — API strings must start with a valid HTTP method

**Security rules** (always run, including AST-compiled schemas):
- `VALIDATE_IAM_IN_EXPRESSION` — no `iam`, `role`, or `permission` keywords
  in `visibleOn` / `disabledOn` expressions

---

### `ui.cache_store` (Priority 80, Not Required)

Writes the compiled schema to cache if `DataKeyCacheHit` is false. Non-fatal
write failures are surfaced in `StageResult.Message` but do not abort.

---

### `ui.response` (Priority 90, Required)

Reads `DataKeySchema` and serialises it to the HTTP response as JSON.

---

## Adding a Stage

1. Create `internal/web/stages/my_stage.go`.
2. Embed `pipeline.BaseStage` and set all fields including `StageDependsOn`.
3. Implement `Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error)`.
4. Register in `NewUIPipeline` in `internal/web/wire.go`.
5. Add the stage name to `DependsOn` of any stage that reads its output.

**Every non-root stage must declare `StageDependsOn`.** The CI guard
(`scripts/check-arch.sh` Guard 5) enforces this.

---

## Observability

Every stage is wrapped by `InstrumentedStage` at wire time (when tracer/mp/log
are non-nil). Per-stage observability:

| Signal | What | When |
|--------|------|------|
| OTel span | `ui.stage.<name>` with route/tenant/cache_hit attributes | Every Execute |
| Histogram | `ui_stage_execution_duration_ms` | Every Execute |
| Counter | `ui_schema_validation_failures_total` | ValidateStage error |
| Counter | `ui_registry_resolution_failures_total` | RegistryStage error |
| WARN log | `slow ui stage: <name>` | Duration > 50 ms |
| ERROR log | `ui stage failed: <name>` | Non-nil error |
