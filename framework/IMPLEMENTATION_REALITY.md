# Framework Implementation Reality

Cross-reference of actual code against documentation. Use this before trusting any doc claim.

---

## Quick Reference

| Area | Status | Notes |
|------|--------|-------|
| EntityDefinition core struct | ⚠️ | Exists but missing `LabelPlural`, `WorkflowTriggers`, `Actions` |
| System vs Custom entity types | ❌ | Documented split not implemented — one struct only |
| OrgScope (`Global/Tenant/Unit`) | ✅ | Implemented, entirely undocumented |
| Privacy policy chain | ⚠️ | Enforcer exists; `ErrAllow/ErrDeny/ErrSkip` symbols **missing** |
| CRUD API handlers | ⚠️ | Works but two hook-timing bugs |
| Tenant isolation (RLS) | ✅ | `set_tenant_context` called correctly |
| BulkUpdate SQL safety | ❌ | SQL injection via unsanitized column names |
| Workflow triggers | ⚠️ | Fires **inside** TX — opposite of documented |
| SDUI schema generation | ⚠️ | No cache, no auth, no PageBuilderSet override |
| Hooks package | ❌ | **Missing entirely** — compilation blocker |
| `definition/policy.go` | ❌ | **Missing** — compilation blocker |
| `definition/viewer.go` | ❌ | **Missing** — compilation blocker |
| `sdui/amis/` sub-package | ❌ | **Missing** — compilation blocker |
| NamingSeriesDef date tokens | ❌ | `{YYYY}` etc. not implemented |
| Casbin RBAC | ❌ | Referenced in docs, zero code found |
| Logger | ⚠️ | Code uses `zerolog`; docs say `slog` |
| Error response envelope | ⚠️ | Code: `{status,msg}`; docs: `{error:{code,message,fields}}` |
| Validation: KRAPin/NHIF | ⚠️ | Implemented, undocumented |
| Validation: Currency precision | ❌ | Converts to `float64` — loses `decimal.Decimal` precision |
| `HookAfterCommit` (outside TX) | ❌ | Not implemented |
| Redis / SDUI caching | ❌ | Not implemented |
| Tenant status validation | ❌ | Not implemented anywhere |

---

## 1. Critical Compilation Blockers

These missing packages/files prevent `go build` from succeeding.

### 1.1 `framework/hooks/` — entire package missing

**Referenced in**: `framework/api/handler.go:7,49,50`
```go
import "awo.so/framework/hooks"
h.hooks = hooks.DefaultRunner   // line 49
h.hooks    *hooks.Runner        // line 34
```
**Needed symbols**: `hooks.Runner` (struct), `hooks.DefaultRunner` (var), `runner.Run(ctx, def, mut, timing) error`

**Fix**: Create `framework/hooks/runner.go` with `Runner` struct that iterates `def.Hooks`, filters by `Timing`, and calls `HookDef.Fn`.

---

### 1.2 `framework/definition/policy.go` — missing

**Referenced in**: `framework/privacy/enforcer.go:40,42,45`
```go
errors.Is(err, definition.ErrAllow)
errors.Is(err, definition.ErrDeny)
errors.Is(err, definition.ErrSkip)
```
Also needed: `PolicyDef` struct and `PolicyFunc` type referenced in `definition.go`.

**Fix**: Create `framework/definition/policy.go`:
```go
var (
    ErrAllow = errors.New("allow")
    ErrDeny  = errors.New("deny")
    ErrSkip  = errors.New("skip")
)
type PolicyFunc func(ctx context.Context, viewer ViewerContext, op Op, record Record) error
type PolicyDef struct {
    Ops []Op
    Fn  PolicyFunc
}
func (p PolicyDef) EffectiveOps() OpSet { ... }
```

---

### 1.3 `framework/definition/viewer.go` — missing

**Referenced in**: `framework/api/handler.go`, `framework/privacy/enforcer.go`, `framework/bootstrap/bootstrap.go`

**Needed interface**:
```go
type ViewerContext interface {
    ActorID() string
    TenantID() string
    OrgUnitID() uuid.UUID
    IsSystem() bool
    HasRole(role string) bool
}
```

**Fix**: Create `framework/definition/viewer.go` with the above interface.

---

### 1.4 `framework/sdui/amis/` — sub-package missing

**Referenced in**: `framework/sdui/handler.go` (inferred from SDUI schema generation)

**Needed**: `amis.CRUDPage(def, opts)`, `amis.FormPage(def, opts)`, `amis.NavTree(defs)`, `amis.PageOpts` struct.

**Fix**: Create `framework/sdui/amis/` package with amis JSON schema builders.

---

## 2. Architecture Divergences

### 2.1 System vs Custom entity type split

**Documented** (`part-02`): Two distinct entity types — System entities backed by typed SQL structs, Custom entities stored in JSONB `entity_records` table.

**Actual** (`framework/definition/definition.go:24`): Single `EntityDefinition` struct. No `Type` field. No `SystemDefinition`/`CustomDefinition`. No JSONB storage path.

**Impact**: All custom-entity dynamic schema features (add fields at runtime, JSONB storage) are unimplemented.

**Recommended fix**: Add `EntityType` enum (`System/Custom`) to `EntityDefinition`. Route pgstore to different SQL paths based on type.

---

### 2.2 OrgScope levels undocumented

**Documented**: No mention of `ScopeLevel` in any doc.

**Actual** (`framework/definition/definition.go`, `framework/api/handler.go:145`):
```go
type ScopeLevel int
const (
    ScopeLevelGlobal ScopeLevel = iota
    ScopeLevelTenant
    ScopeLevelUnit
)
```
Controls whether `tenant_id` and `org_unit_id` columns exist and are filtered.

**Impact**: Developers don't know this exists. Global entities (no tenant isolation) are silently possible.

**Recommended fix**: Document in part-02 and part-07. Add to CLAUDE.md.

---

### 2.3 HookDef structure mismatch

**Documented** (`part-02`): `HookSet` struct with named fields (`BeforeValidate`, `BeforeSave`, `AfterSave`, `AfterCommit`, `BeforeDelete`, `OnSubmit`, `OnCancel`).

**Actual** (`framework/definition/hook.go`): Flat `[]HookDef` slice with `Timing` discriminator field.

**Impact**: Code that constructs `HookSet{BeforeSave: fn}` won't compile. Must use `HookDef{Timing: HookBeforeSave, Fn: fn}`.

---

### 2.4 NamingSeriesDef date tokens not implemented

**Documented** (`part-02`): `{YYYY}`, `{MM}`, `{DD}` tokens in prefix strings, auto-reset counter per period.

**Actual** (`framework/definition/naming.go`):
```go
type NamingSeriesDef struct {
    Field   string
    Prefix  string  // literal only
    Padding int
}
```
No token parsing, no date substitution, no period-based counter reset.

**Impact**: `NamingSeriesDef{Prefix: "INV-{YYYY}-{MM}-"}` silently produces `INV-{YYYY}-{MM}-00001`.

---

### 2.5 EdgeDef not wired to persistence

**Documented** (`part-02`): Edges drive JOIN queries and eager loading.

**Actual**: `EdgeDef` defined in `framework/definition/edge.go` but pgstore has no JOIN/eager-load logic. `EagerLoad: true` is a no-op.

---

### 2.6 Casbin RBAC absent

**Documented** (`part-03`, `part-07`): Casbin enforcer for role-permission checks, `HasPermission(subject, entity, action)` calls.

**Actual**: Zero Casbin imports anywhere in `framework/`. Privacy policies are pure Go functions. No role-string permission table.

**Impact**: `HasRole("admin")` works; structured permission checks (`HasPermission("admin","invoice","approve")`) do not exist.

---

## 3. API Contract Differences

### 3.1 HTTP method: PUT not PATCH

**Documented** (`part-03`): `PATCH /:id` for partial updates.

**Actual** (`framework/api/handler.go:84`): `PUT /:id`.

**Impact**: Clients sending `PATCH` get 405. Existing integrations built on docs will fail.

---

### 3.2 List response: no envelope

**Documented** (`part-03`): `{ "data": [...], "meta": { "total", "page", "per_page" } }`

**Actual** (`framework/api/handler.go:164-169`):
```go
fiber.Map{
    "data":   rows,
    "total":  page.Total,
    "limit":  page.Limit,
    "offset": page.Offset,
}
```
No `meta` wrapper. Pagination uses `limit`/`offset` (not `page`/`per_page`).

---

### 3.3 Error response format

**Documented** (`part-03`):
```json
{ "error": { "code": "ENTITY_NOT_FOUND", "message": "...", "fields": {...} } }
```

**Actual** (`cmd/framework-server/main.go:186-187`):
```go
fiber.Map{"status": code, "msg": msg}
```
No `error` wrapper. No `code` string. No `fields` map.

---

### 3.4 Route path uses TableName not entity name

**Documented** (`part-03`): Routes at `/{prefix}/{entity-name}`.

**Actual** (`framework/bootstrap/bootstrap.go:92`):
```go
api.Register(apiGroup, "/"+def.TableName(), h)
```
Routes at `/{prefix}/{table_name}`. For entity `Invoice` with table `invoices` these match, but for entities where name ≠ table name they diverge.

---

### 3.5 HookBeforeValidate runs inside TX on update

**File**: `framework/api/handler.go`

**Create path** (line 204): `HookBeforeValidate` runs OUTSIDE transaction — correct for normalizing input before validation.

**Update path** (line 269): `HookBeforeValidate` runs INSIDE transaction — inconsistent, blocks connection longer, prevents external API calls in hook.

**Fix**: Move `HookBeforeValidate` call to before `h.store.WithTx(...)` in the update handler.

---

### 3.6 HookAfterSave called after delete

**File**: `framework/api/handler.go:330`
```go
return h.hooks.Run(c.Context(), h.def, mut, definition.HookAfterSave)
```
Called from `h.delete()`. Should be `HookAfterDelete`. Any hook registered on `HookAfterSave` will incorrectly fire after deletion.

**Fix**: Define `HookAfterDelete` timing constant. Use it here.

---

## 4. Multi-Tenancy Gaps

### 4.1 Tenant status never validated

**Documented** (`part-07`): Requests to SUSPENDED or ARCHIVED tenants return 403.

**Actual**: Zero tenant status checks anywhere. A suspended tenant's data is fully accessible if the caller knows the tenant ID.

**Fix**: Add middleware or `extractContext` check that queries `tenants.status` and rejects non-ACTIVE tenants.

---

### 4.2 `X-Awo-Tenant` vs documented header name

**Documented** (`part-07`): Header `X-Tenant-ID`.

**Actual** (`cmd/framework-server/main.go:63`, `framework/bootstrap/bootstrap.go:101`): Header `X-Awo-Tenant`.

**Fix**: Align docs to match code (`X-Awo-Tenant`) — code is more specific and less likely to clash.

---

### 4.3 RLS `set_tenant_context` function assumed to exist

**Actual** (`framework/persistence/pgstore/store.go:474,497`):
```go
"SELECT set_tenant_context($1)"
```
If this PostgreSQL function is not created in migrations, every query silently skips RLS. No error surfaces — the function call result is ignored (`pool.Exec` return not checked for function-missing errors).

**Fix**: Assert function existence at startup, or check migration output.

---

## 5. SDUI Gaps

### 5.1 No cache

**Documented** (`part-04`): Redis-backed schema cache, TTL configurable.

**Actual** (`framework/sdui/handler.go`): Schema generated fresh on every request. No Redis dependency, no cache invalidation.

**Impact**: Under load, SDUI endpoints regenerate JSON on every hit.

---

### 5.2 No permission check on SDUI routes

**Documented** (`part-04`): SDUI schema filtered by viewer permissions — hidden fields, disabled actions.

**Actual**: `GET /sdui/:entity` returns full schema with no `ViewerContext`. Sensitive field names visible to unauthenticated callers.

---

### 5.3 PageBuilderSet override not implemented

**Documented** (`part-04`): `EntityDefinition.PageBuilderSet` lets modules provide custom amis schema.

**Actual**: Field missing from `EntityDefinition` struct. SDUI always generates default schema.

---

### 5.4 No form schema for create vs edit

**Documented** (`part-04`): Separate form schemas for create/edit modes, field visibility rules.

**Actual** (`framework/sdui/handler.go`): Single `GET /sdui/:entity/form` endpoint, one schema for both modes.

---

## 6. Workflow Gaps

### 6.1 Trigger fires inside transaction

**Documented** (`part-05`): Temporal triggers fire AFTER commit so DB failure doesn't launch orphaned workflows.

**Actual** (`framework/workflow/trigger.go:75-78`):
```go
// HookAfterSave — still inside transaction
wfExec.trigger(ctx, def, mut)
```
If Temporal is unavailable → hook returns error → **transaction rolls back** (data lost). Documented behavior is the opposite: DB commit succeeds, workflow trigger is best-effort.

**Fix**: Use `HookAfterCommit` (currently unimplemented) or pass trigger to an outbox/goroutine launched after commit.

---

### 6.2 Workflow ID format mismatch

**Documented** (`part-05`): `"{tenant}.{entity}.{id}.{event}"`

**Actual** (`framework/workflow/trigger.go`):
```
"{op}/{record-id}/{WorkflowType}"
```
No tenant prefix. Workflow IDs not globally unique across tenants — collision risk in shared Temporal namespace.

---

### 6.3 `HookAfterCommit` not implemented

**Documented** (`part-05`): `AfterCommit` hook timing runs after DB commit, outside transaction, for side effects.

**Actual** (`framework/definition/hook.go`): Timings: `HookBeforeValidate`, `HookBeforeSave`, `HookAfterSave`, `HookBeforeDelete`, `HookOnSubmit`, `HookOnCancel`. No `HookAfterCommit`.

---

## 7. Security Issues

### 7.1 SQL injection in BulkUpdate

**File**: `framework/persistence/pgstore/store.go:304-325`
**Severity**: HIGH

```go
for col, val := range updates {
    fmt.Fprintf(&sb, "%s = $%d", col, idx)  // col unsanitized!
    args = append(args, val)
    idx++
}
```
Caller-supplied map keys flow directly into SQL without quoting or allowlist validation. Any caller (including API body deserialization paths) can inject arbitrary SQL.

**Fix**: Validate column names against `def.Fields` allowlist before building query:
```go
allowed := def.FieldSet()  // map[string]bool
if !allowed[col] {
    return fmt.Errorf("unknown field %q", col)
}
fmt.Fprintf(&sb, `"%s" = $%d`, col, idx)  // quote identifier
```

---

### 7.2 SDUI routes unauthenticated

**File**: `framework/sdui/handler.go`

No auth middleware on SDUI routes. Any caller gets full entity schema including field names marked `IsSensitive`. See §5.2.

---

## 8. Implementation Inconsistencies

### 8.1 Logger: zerolog vs slog

**Documented**: `slog` (standard library).
**Actual**: `zerolog` (`github.com/rs/zerolog`).

All framework logging uses `log.Info().Str(...).Msg(...)` zerolog API. Zero `slog` imports found.

---

### 8.2 Validation: Currency precision loss

**File**: `framework/validate/validate.go` (`validateNumericRange`)

For `FieldTypeCurrency`, min/max comparison converts value to `float64`:
```go
f, _ := strconv.ParseFloat(fmt.Sprint(val), 64)
```
`decimal.Decimal` values lose precision. A value of `1234567890.123456789` becomes `1234567890.1234567` after round-trip.

**Fix**: Type-assert to `decimal.Decimal` and use its `Cmp` method.

---

### 8.3 KRAPin / NHIF validators undocumented

**File**: `framework/validate/validate.go`

`validate.KRAPin()` and `validate.NHIF()` exist — Kenya-specific business validators. Not mentioned in any doc. Developers won't discover them without reading source.

---

### 8.4 `EntityRepository[T]` generic interface unused by handlers

**File**: `framework/persistence/store.go`

Generic `EntityRepository[T any]` defined with `Query(ctx, ListOptions) ([]T, int64, error)`. Actual handlers use `persistence.TenantStore` / `persistence.TenantTx` / `EntityStore` (untyped `map[string]any`). Generic interface is dead code or intended future API.

---

### 8.5 `anonymousViewer` returns no error — silently allows unauthenticated access

**File**: `framework/bootstrap/bootstrap.go:100-106`

Default `ViewerFn` returns an anonymous viewer with empty `TenantID()`. Down the call stack, `extractContext` will call `tenantResolver("")` or fail UUID parse — resulting in 401. But for Global-scoped entities, tenant check may be skipped, allowing anonymous reads.

---

## 9. Documentation Needs

| Gap | Where to add |
|-----|-------------|
| `ScopeLevel` (Global/Tenant/Unit) semantics | `part-02`, `part-07` |
| Actual hook timing constants and slice-based registration | `part-02` |
| `X-Awo-Tenant` header name (not `X-Tenant-ID`) | `part-03`, `part-07` |
| `PUT` not `PATCH` for updates | `part-03` |
| Actual list response shape (`limit`/`offset` not `page`/`per_page`) | `part-03` |
| Actual error response shape (`{status, msg}`) | `part-03` |
| `zerolog` as actual logger | `part-01` |
| KRAPin / NHIF validators | `part-02` |
| `NamingSeriesDef` date tokens not implemented | `part-02` |
| Casbin RBAC not implemented | `part-03`, `part-07` |
| Workflow trigger fires inside TX (current behavior) | `part-05` |
| SDUI cache not implemented | `part-04` |
| Tenant status not validated on requests | `part-07` |

---

## 10. Go Package Violations

| Violation | Location | Rule |
|-----------|----------|------|
| `hooks` package imported but doesn't exist | `api/handler.go:7` | Build fails |
| `definition.ErrAllow/ErrDeny/ErrSkip` undefined | `privacy/enforcer.go:40-45` | Build fails |
| `definition.ViewerContext` undefined | multiple | Build fails |
| `sdui/amis` sub-package undefined | `sdui/handler.go` | Build fails |
| Unsanitized fmt.Sprintf into SQL | `pgstore/store.go:311` | `SA1006` / security |
| `float64` for decimal comparison | `validate/validate.go` | Precision loss |
| Hook timing constant `HookAfterDelete` missing | `api/handler.go:330` | Wrong semantic |
| Workflow ID not tenant-prefixed | `workflow/trigger.go` | Namespace collision |
| SDUI handler has no `ViewerContext` param | `sdui/handler.go` | Auth bypass |

---

## Priority Fix Order

Fix in this sequence — earlier items unblock later ones.

### P0 — Build Blockers (do first, nothing compiles without these)

1. **Create `framework/definition/viewer.go`** — `ViewerContext` interface
2. **Create `framework/definition/policy.go`** — `ErrAllow`, `ErrDeny`, `ErrSkip`, `PolicyDef`, `PolicyFunc`
3. **Create `framework/hooks/runner.go`** — `Runner` struct, `DefaultRunner` var, `Run()` method
4. **Create `framework/sdui/amis/`** — `CRUDPage`, `FormPage`, `NavTree`, `PageOpts`

### P1 — Security (fix before any production traffic)

5. **Fix `BulkUpdate` SQL injection** (`pgstore/store.go:311`) — allowlist column names against `def.Fields`
6. **Add auth to SDUI routes** (`sdui/handler.go`) — require `ViewerContext`, filter sensitive fields

### P2 — Data Integrity

7. **Fix `HookAfterSave` called after delete** (`api/handler.go:330`) — use `HookAfterDelete`
8. **Fix `HookBeforeValidate` inside TX on update** (`api/handler.go:269`) — move outside `WithTx`
9. **Fix workflow trigger inside TX** (`workflow/trigger.go:75`) — defer to post-commit or outbox
10. **Fix Currency `float64` precision** (`validate/validate.go`) — use `decimal.Decimal.Cmp`

### P3 — Correctness

11. **Add tenant status validation** — middleware or `extractContext` check against `tenants.status`
12. **Fix workflow ID format** — prefix with tenant ID for Temporal namespace isolation
13. **Implement `HookAfterCommit`** — add timing constant; needed by workflow triggers and outbox

### P4 — Feature Completeness

14. **Implement NamingSeriesDef date tokens** — `{YYYY}`, `{MM}`, `{DD}` with period-reset counter
15. **Implement SDUI Redis cache** — add cache layer to `sdui/handler.go`
16. **Implement eager-load edges** — wire `EdgeDef.EagerLoad` into pgstore JOIN queries
17. **Implement Custom entity JSONB path** — add `EntityType` field, route to `entity_records` table

### P5 — Doc / Contract Alignment

18. Update docs: header name, HTTP method, response shapes, logger, scope levels, hook structure
19. Add `HookAfterDelete` constant and use it
20. Decide: implement Casbin RBAC or remove from docs
