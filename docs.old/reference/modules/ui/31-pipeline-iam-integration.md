[<-- Back to Index](README.md)

> **📌 PARTIALLY CURRENT.** This document corrects earlier design specs (§28–§30) with actual
> implementation findings. Sections 1–3 are accurate. Later sections may describe patterns not
> yet implemented. The referenced docs (§28, §29, §30) are superseded.
>
> **For current authoritative docs:**
> - [Pipeline Deep Dive](../02-architecture/02-pipeline-deep-dive.md) — 9-stage pipeline
> - [IAM Integration](../02-architecture/03-iam-integration.md) — `contract.SessionContext` usage
> - [Glossary](../appendices/A-glossary.md) — type disambiguation

## UI Compiler + Pipeline + IAM Contract Integration

Reference: [§28 Go-First UI Compiler](28-go-ui-compiler.md) · [§29 IAM/Session Compiler](29-iam-session-compiler.md) · [§30 Implementation Pipeline](30-implementation-pipeline.md)

---

## 1. Architecture Rewrite — Ground Truth First

**Correction from initial analysis:** `internal/core/iam/contract/` exists and is fully implemented. The initial §31 draft stated it did not exist — that was wrong. This section replaces the incorrect analysis.

### What the contract package actually is

`internal/core/iam/contract/` is the stable IAM consumer interface for all non-IAM modules. From `doc.go`:

> All non-IAM modules (billing, inventory, HR, etc.) interact with IAM exclusively through this package.

It provides:

| Export | Purpose |
|---|---|
| `SessionContext` | Read-only identity + runtime metadata carrier. Wraps `*iam.ResolvedSession`, exposes only consumer-relevant fields. |
| `FromContext(ctx)` | Retrieve `SessionContext` from Go context. The standard pattern for service-layer code. |
| `WithContext(ctx, sc)` | Store `SessionContext` in Go context. Called by middleware, not by modules. |
| `InjectSessionContext()` | Fiber middleware: reads `*iam.ResolvedSession` from Fiber Locals, wraps in `SessionContext`, injects into Go context. |
| `AuthService` | Narrow interface: Login / Logout / ValidateSession only. |

### What `SessionContext` deliberately does NOT expose

From `session.go` comment:
```go
// What it does NOT expose:
//   - permission maps or Can()/CanDo() (removed — all authz via Casbin)
//   - ToPrincipal() — IAM-internal; needed only for Enforce() calls
//   - raw UserType string — use IsPlatform()/IsPortal() for context checks
```

**This is the critical design constraint for the UI pipeline.** Modules get identity and feature flags. They do NOT get a permission map. They do NOT call Casbin. The contract doc states:

> MUST NOT call authzService.Enforce() directly
> MAY assume the request is already authorized by IAM middleware

### The design tension

Casbin IAM middleware authorizes at route level: *"can this user access `/schema/finance/invoices`?"*

UI page functions need fine-grained render decisions: *"should the Delete button exist in the schema?"*

These are different concerns. Route-level authorization is handled by Casbin before the handler runs. UI-level conditional rendering is a schema composition concern — it produces different JSON, not different access decisions. The resolution:

- Route guard: Casbin middleware (before handler)
- UI conditional rendering: `sc.FeatureEnabled()` + `sc.IsPlatform()` + `sc.IsPortal()` + a `UIAuthzService` interface that wraps Casbin for bulk UI permission resolution

`UIAuthzService` is provided to UI stages as a constructor dependency. It is the ONLY place Casbin is called for UI purposes. Modules do not receive the `UIAuthzService` — only the pipeline stage does. This keeps the contract boundary intact.

### Facts established

**Fact 1: `contract.SessionContext` is the correct IAM type for all non-IAM code.**
Not `*iam.ResolvedSession`. The pipeline's `OperationContext.Session *iam.ResolvedSession` is a direct IAM domain import — acceptable for the pipeline infrastructure package, but the UI handler layer must use `contract.FromContext()`.

**Fact 2: Authorization is NOT in `SessionContext`. Feature flags ARE.**
`sc.FeatureEnabled("flag")` → `session.Configuration.Flags` (pre-computed at login). `sc.Can(...)` → does not exist. Bulk permission resolution for UI rendering goes through `UIAuthzService`.

**Fact 3: The correct session propagation chain is:**
```
Fiber Locals[iam.ResolvedSession]
  → contract.InjectSessionContext() middleware
    → contract.WithContext(ctx, sc)
      → contract.FromContext(ctx) in handler/service
```
`SchemaHandler` uses `contract.FromContext(c.UserContext())`, not `c.Locals(iam.LocalsKeySession)`.

**Fact 4: The pipeline infrastructure already supports the execution model.**
`OperationContext` is pooled, has `FeatureEnabled()` delegation to `Session`, priority-ordered stages, compensation, branch jumps. Nothing new is needed in `internal/pipeline/`.

---

### What Changes

The UI compiler as designed in §28/§29 executes PageFn directly in an HTTP handler:

```
HTTP request → AuthMiddleware → IAMMiddleware → SchemaHandler → compiler.Compile() → PageFn(sess) → JSON
```

This creates the problems identified in the brief: ad-hoc execution, no lifecycle, untracked observability, unclear cache ownership, duplicated orchestration.

**The unified model replaces `compiler.Compile()` with a pipeline execution:**

```
HTTP request → AuthMiddleware → SchemaHandler → pipeline.Run(opCtx) → stages → JSON
```

IAM resolution, cache lookup, schema compilation, validation, cache storage, and observability become discrete, ordered, testable pipeline stages. The `SchemaHandler` becomes a thin adapter: it builds an `OperationContext`, runs the pipeline, reads `opCtx.Data["ui.schema"]`, and returns the JSON envelope.

---

## 2. Pipeline Flow Diagram

```
HTTP GET /schema/finance/invoices
│
├── iam/middleware.Authenticate(cfg)
│   ├── Extract token (cookie → Authorization: Bearer → reject)
│   ├── Validate against session store
│   ├── Hydrate *iam.ResolvedSession
│   └── Store in Fiber Locals[iam.LocalsKeySession]
│
├── contract.InjectSessionContext()   ← bridges Fiber Locals → Go context
│   ├── Read *iam.ResolvedSession from Locals
│   ├── Wrap: contract.newSessionContext(resolved) → SessionContext
│   └── Inject: contract.WithContext(c.UserContext(), sc)
│
└── SchemaHandler.Handle(c *fiber.Ctx)
    ├── sc, ok := contract.FromContext(c.UserContext())  ← NO Locals access
    ├── AcquireOperationContext()  ← pooled, zero GC
    ├── Populate opCtx:
    │   ├── Ctx          = c.UserContext()   ← carries contract.SessionContext
    │   ├── Session      = sc.resolved       ← pipeline infra needs *iam.ResolvedSession
    │   ├── TenantID     = sc.TenantID()
    │   ├── UserID       = sc.UserID()
    │   ├── OperationKey = "ui.schema.compile"
    │   └── Input        = UISchemaInput{Route: "/finance/invoices", Session: sc}
    │
    └── pipeline.Run(opCtx)
        │
        ├── STAGE 1: ui.iam_resolve          [Priority: 100, Required: true]
        │   ├── Read: opCtx.Session.ToPrincipal()
        │   ├── Call: authzService.BulkEnforce(principal, AllUIPermissions)
        │   ├── Write: opCtx.Data["ui.permissions"] = map[string]bool{...}
        │   └── Write: opCtx.Data["ui.perm_fingerprint"] = sha256(sorted true keys)
        │   [Feature flags come from opCtx.FeatureEnabled() — already resolved at login]
        │   [Flag fingerprint: sha256 of true flags from opCtx.Session.Configuration.Flags]
        │
        ├── STAGE 2: ui.cache_lookup         [Priority: 150, Required: false]
        │   ├── Read: opCtx.TenantID + input.Route + perm_fingerprint + flag_fingerprint
        │   ├── Build: cacheKey = "{tenantID}|{route}|{permFP}|{flagFP}"
        │   ├── Hit: write opCtx.Data["ui.schema"] = cached, SetFlag("ui.cache_hit", true)
        │   ├── Hit: opCtx.NextStageID = "ui.response" (skip compile/validate/store)
        │   └── Miss: continue normally
        │
        ├── STAGE 3: ui.registry_resolve     [Priority: 200, Required: true]
        │   ├── Read: input.Route
        │   ├── Call: registry.Get(input.Route)
        │   ├── 404: return error → pipeline aborts, SchemaHandler returns 404
        │   └── Write: opCtx.Data["ui.page_fn"] = PageFn
        │
        ├── STAGE 4: ui.compile              [Priority: 300, Required: true]
        │   ├── Read: opCtx.Data["ui.page_fn"], opCtx.Data["ui.permissions"]
        │   ├── Build UISessionContext from opCtx (see §4 below)
        │   ├── Call: pageFn(uiSessCtx) → map[string]any
        │   └── Write: opCtx.Data["ui.schema"] = schema
        │
        ├── STAGE 5: ui.normalize            [Priority: 400, Required: true]
        │   ├── Read: opCtx.Data["ui.schema"]
        │   ├── Run: SchemaValidator.Validate(schema)
        │   ├── Error: pipeline aborts → SchemaHandler returns 500 (compile-time bug)
        │   └── Pass: schema written back (may be modified by normalization)
        │
        ├── STAGE 6: ui.cache_store          [Priority: 450, Required: false]
        │   ├── Read: opCtx.Data["ui.schema"], cache key from stage 2
        │   ├── Write: cache.Set(key, schema, ttl)
        │   └── Failure: logged, does not fail the pipeline (cache is best-effort)
        │
        └── STAGE 7: ui.response             [Priority: 500, Required: true]
            ├── Read: opCtx.Data["ui.schema"]
            └── Write: opCtx.Data["ui.response"] = {status:0, data:schema}
            [SchemaHandler reads this and calls c.JSON()]

```

**Branch jump detail:** When `ui.cache_lookup` hits, it sets `result.NextStageID = "ui.response"`. The pipeline's existing branch jump logic (builder.go:165) skips stages 3–6 and jumps directly to stage 7. No special code needed — the mechanism exists.

**TxHooks:** None registered. UI compilation is read-only. `fireTxHooks` exits immediately when `len(opCtx.TxHooks) == 0`.

**Compensation:** Stages 3–6 are pure (no side effects worth reversing). Stage 6 (`ui.cache_store`) is non-required — a cache write failure does not abort. No `Compensatable` implementations needed.

---

## 3. Go Implementation Design

### 3.1 Operation Key and Input Type

```go
// internal/web/pipeline/types.go
package uipipeline

const OperationKey = "ui.schema.compile"

// UISchemaInput is the typed input for the UI compilation pipeline.
// Set as opCtx.Input before calling pipeline.Run().
type UISchemaInput struct {
    Route string // e.g. "/finance/invoices"
}

// UISchemaOutput is read from opCtx.Data["ui.response"] after pipeline.Run().
type UISchemaOutput struct {
    Schema map[string]any // the compiled AMIS schema
    CacheHit bool
}
```

---

### 3.2 UISessionContext — The IAM Boundary Object for PageFn

PageFn receives `UISessionContext`. It is built from `contract.SessionContext` (identity + feature flags) plus the resolved permission map from the `ui.iam_resolve` stage.

**Why not pass `contract.SessionContext` directly to PageFn?**
`contract.SessionContext` has no `Can()` method — by design. Feature flags are there; permissions are not. PageFn needs both. `UISessionContext` combines the two without violating the contract: it is constructed by an internal pipeline stage (not by module code), and the permission map is populated exclusively by `ui.iam_resolve` via `UIAuthzService`.

```go
// internal/web/pipeline/session_ctx.go
package uipipeline

import (
    "awo.so/internal/core/iam/contract"
    "github.com/google/uuid"
)

// UISessionContext is the read-only view of session+IAM state passed to PageFn.
//
// Invariants:
//   - permissions map populated by ui.iam_resolve stage ONLY
//   - feature flags come from contract.SessionContext.FeatureEnabled() (pre-computed at login)
//   - PageFn MUST NOT receive contract.SessionContext or *iam.ResolvedSession directly
//   - PageFn MUST NOT call any IAM service, Casbin, or feature flag service
//   - UISessionContext is constructed once per cache miss, schema then cached
type UISessionContext struct {
    UserID      uuid.UUID
    TenantID    uuid.UUID
    DisplayName string
    IsPlatform  bool
    IsPortal    bool

    // Locale, Timezone, Currency surfaced as named fields for ergonomic PageFn access.
    // Source: contract.SessionContext.Preference() / Setting()
    Locale   string
    Timezone string
    Currency string

    permissions map[string]bool // private — populated by ui.iam_resolve only
}

// Can returns true iff the session has the resolved permission "resource.action".
func (u UISessionContext) Can(action, resource string) bool {
    return u.permissions[resource+"."+action]
}

// CanAny returns true iff the session has the named permission for any of the resources.
func (u UISessionContext) CanAny(action string, resources ...string) bool {
    for _, r := range resources {
        if u.Can(action, r) {
            return true
        }
    }
    return false
}

// CanAll returns true iff the session has the named permission for every resource.
func (u UISessionContext) CanAll(action string, resources ...string) bool {
    for _, r := range resources {
        if !u.Can(action, r) {
            return false
        }
    }
    return true
}

// Flag delegates to the original contract.SessionContext's FeatureEnabled.
// Note: featureFlags is stored as a plain map to avoid holding a reference to
// the contract.SessionContext after UISessionContext construction.
func (u UISessionContext) Flag(name string) bool {
    return u.featureFlags[name]
}

// Pref returns a user UI preference string.
func (u UISessionContext) Pref(key, def string) string {
    if v, ok := u.prefs[key]; ok {
        return v
    }
    return def
}

// unexported backing stores — populated at construction, never mutated
type UISessionContext struct {
    // ... (fields above)
    featureFlags map[string]bool // copy of contract.SessionContext flags at construction
    prefs        map[string]string
}

// fromOperationContext constructs UISessionContext after ui.iam_resolve has run.
// The ONLY constructor. Called exclusively by the ui.compile stage.
func fromOperationContext(opCtx *pipeline.OperationContext) (UISessionContext, bool) {
    perms, ok := opCtx.Data["ui.permissions"].(map[string]bool)
    if !ok {
        return UISessionContext{}, false
    }

    // Retrieve the contract.SessionContext from the Go context — canonical source.
    sc, ok := contract.FromContext(opCtx.Ctx)
    if !ok || sc.IsZero() {
        return UISessionContext{}, false
    }

    // Copy feature flags into a plain map. We do not hold a reference to sc
    // because UISessionContext outlives the request scope when cached.
    // NOTE: flags are copied as a snapshot — correct because UISessionContext
    // is only built on a cache miss (flags fingerprint already verified).
    flagsCopy := copyFlagsFromContract(sc)
    prefsCopy := copyPrefsFromContract(sc)

    return UISessionContext{
        UserID:      sc.UserID(),
        TenantID:    sc.TenantID(),
        DisplayName: sc.DisplayName(),
        IsPlatform:  sc.IsPlatform(),
        IsPortal:    sc.IsPortal(),
        Locale:      sc.Preference("ui.locale", "en"),
        Timezone:    sc.Preference("ui.timezone", "UTC"),
        Currency:    sc.Setting("tenant.currency", "USD"),
        permissions: perms,
        featureFlags: flagsCopy,
        prefs:        prefsCopy,
    }, true
}
```

**Why copy flags into a plain map rather than store `contract.SessionContext`?**
`UISessionContext` may be placed in a schema cache that outlives the request. `contract.SessionContext` holds a pointer to `*iam.ResolvedSession`. Caching that pointer would hold the session object in memory indefinitely and prevent GC. A plain `map[string]bool` copy has value semantics — safe to cache, safe to compare.

---

### 3.3 PageFn Type (updated)

```go
// internal/web/types.go
package web

import uipipeline "awo.so/internal/web/pipeline"

// PageFn is the signature all page builder functions must satisfy.
// Pure function: same UISessionContext → same Schema.
// PageFn MUST NOT access IAM services, databases, or external APIs.
type PageFn func(sess uipipeline.UISessionContext) Schema
```

---

### 3.4 Stage Implementations

#### Stage 1 — IAM Resolution

```go
// internal/web/pipeline/stages/iam_resolve.go
package stages

import (
    "awo.so/internal/core/iam/contract"
    "awo.so/internal/pipeline"
    "awo.so/internal/web/policy"
)

// UIAuthzService wraps Casbin for bulk UI permission resolution.
// Defined in internal/web/policy — the ONLY place in the UI layer
// that is allowed to call Casbin. Provided to this stage at wire time.
//
// It accepts contract.SessionContext (not *iam.ResolvedSession) so that
// the stage never imports internal IAM domain types directly.
type UIAuthzService interface {
    // BulkEnforce resolves all permissions in AllUIPermissions for the given
    // session's principal. Returns a map[permission]bool.
    BulkEnforce(ctx context.Context, sc contract.SessionContext, perms []policy.UIPermission) (map[string]bool, error)
}

type IAMResolveStage struct {
    pipeline.BaseStage
    authz UIAuthzService
}

func NewIAMResolveStage(authz UIAuthzService) *IAMResolveStage {
    return &IAMResolveStage{
        BaseStage: pipeline.BaseStage{
            StageName:       "ui.iam_resolve",
            StageOperations: []string{"ui.schema.compile"},
            StagePriority:   100,
            StageRequired:   true,
        },
        authz: authz,
    }
}

func (s *IAMResolveStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
    // Retrieve SessionContext from Go context — contract.InjectSessionContext()
    // middleware put it here before the handler ran.
    sc, ok := contract.FromContext(opCtx.Ctx)
    if !ok || sc.IsZero() {
        return pipeline.StageResult{}, errors.New("ui.iam_resolve: no session in context")
    }

    // Bulk-evaluate all UI permissions. UIAuthzService internally calls
    // sc.resolved.ToPrincipal() — that call is inside internal/web/policy,
    // which is allowed to import internal IAM types.
    perms, err := s.authz.BulkEnforce(opCtx.Ctx, sc, policy.AllUIPermissions)
    if err != nil {
        return pipeline.StageResult{}, fmt.Errorf("ui.iam_resolve: %w", err)
    }

    permFP := fingerprintMap(perms)

    // Feature flag fingerprint from contract.SessionContext.FeatureEnabled().
    // No external call — flags pre-computed at login.
    flagFP := fingerprintFlagsFromContract(sc, policy.AllUIFlags)

    return pipeline.StageResult{
        Status: "completed",
        Outputs: map[string]any{
            "ui.permissions":      perms,
            "ui.perm_fingerprint": permFP,
            "ui.flag_fingerprint": flagFP,
        },
    }, nil
}
```

**Critical boundary enforcement:**
- `IAMResolveStage` calls `contract.FromContext()` — never `c.Locals()` or `opCtx.Session.ToPrincipal()` directly
- `UIAuthzService` interface takes `contract.SessionContext` — the `BulkEnforce` implementation (in `internal/web/policy/casbin_engine.go`) is allowed to call `ToPrincipal()` because it is a policy package, not a module
- All downstream stages read `opCtx.Data["ui.permissions"]` — a plain `map[string]bool` with no IAM type dependency

#### Stage 2 — Cache Lookup

```go
type CacheLookupStage struct {
    pipeline.BaseStage
    cache platform.SchemaCache
}

func (s *CacheLookupStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
    input, ok := opCtx.Input.(UISchemaInput)
    if !ok {
        return pipeline.StageResult{Status: "skipped", Message: "no input"}, nil
    }

    permFP, _ := opCtx.Data["ui.perm_fingerprint"].(string)
    flagFP, _ := opCtx.Data["ui.flag_fingerprint"].(string)

    key := buildCacheKey(opCtx.TenantID.String(), input.Route, permFP, flagFP)

    schema, hit := s.cache.Get(key)
    if !hit {
        return pipeline.StageResult{
            Status:  "completed",
            Outputs: map[string]any{"ui.cache_key": key},
        }, nil
    }

    return pipeline.StageResult{
        Status:      "completed",
        NextStageID: "ui.response", // skip compile, validate, cache_store
        Outputs: map[string]any{
            "ui.schema":    schema,
            "ui.cache_key": key,
            "ui.cache_hit": true,
        },
    }, nil
}
```

**Cache key invariant:** Key is built AFTER IAM resolution. Building before IAM resolution would allow cache serving of a schema compiled for different permissions — a security defect. Stage ordering (100 before 150) enforces this.

#### Stage 3 — Registry Resolution

```go
func (s *RegistryResolveStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
    input := opCtx.Input.(UISchemaInput)
    fn := registry.Get(input.Route)
    if fn == nil {
        return pipeline.StageResult{}, &UINotFoundError{Route: input.Route}
    }
    return pipeline.StageResult{
        Status:  "completed",
        Outputs: map[string]any{"ui.page_fn": fn},
    }, nil
}
```

**`UINotFoundError` is a sentinel.** SchemaHandler checks `errors.As(err, &UINotFoundError{})` to return 404. All other errors return 500.

#### Stage 4 — Compilation

```go
func (s *CompileStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
    fn, ok := opCtx.Data["ui.page_fn"].(web.PageFn)
    if !ok {
        return pipeline.StageResult{}, errors.New("ui.compile: page_fn not found in context")
    }

    uiSess, ok := fromOperationContext(opCtx)
    if !ok {
        return pipeline.StageResult{}, errors.New("ui.compile: could not build UISessionContext")
    }

    // PageFn is a pure function. Panics indicate programmer error (not user error).
    schema := fn(uiSess)

    return pipeline.StageResult{
        Status:  "completed",
        Outputs: map[string]any{"ui.schema": schema},
    }, nil
}
```

**PageFn receives `UISessionContext`, not `*ResolvedSession`.** This enforces the IAM contract boundary at the type level. PageFn cannot access `ToPrincipal()`, `EntityScope`, `UserType`, or any other IAM internals — those fields do not exist on `UISessionContext`.

#### Stage 5 — Normalization/Validation

```go
func (s *NormalizeStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
    schema, ok := opCtx.Data["ui.schema"].(map[string]any)
    if !ok {
        return pipeline.StageResult{}, errors.New("ui.normalize: no schema in context")
    }

    if err := s.validator.Validate(schema); err != nil {
        // Validation failure = programmer error in a PageFn.
        // Required: true → pipeline aborts. Never reaches the client.
        return pipeline.StageResult{}, fmt.Errorf("ui.normalize: %w", err)
    }

    return pipeline.StageResult{
        Status:  "completed",
        Outputs: map[string]any{"ui.schema": schema},
    }, nil
}
```

#### Stage 6 — Cache Store (non-required)

```go
func (s *CacheStoreStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
    schema, _ := opCtx.Data["ui.schema"].(map[string]any)
    key, _ := opCtx.Data["ui.cache_key"].(string)

    if schema == nil || key == "" {
        return pipeline.StageResult{Status: "skipped"}, nil
    }

    s.cache.Set(key, schema, s.ttl)

    return pipeline.StageResult{Status: "completed"}, nil
}
// StageRequired: false — cache failure never aborts pipeline
```

#### Stage 7 — Response Assembly

```go
func (s *ResponseStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
    schema, ok := opCtx.Data["ui.schema"].(map[string]any)
    if !ok {
        return pipeline.StageResult{}, errors.New("ui.response: no schema in context")
    }

    cacheHit, _ := opCtx.Data["ui.cache_hit"].(bool)

    return pipeline.StageResult{
        Status: "completed",
        Outputs: map[string]any{
            "ui.response": UISchemaOutput{
                Schema:   schema,
                CacheHit: cacheHit,
            },
        },
    }, nil
}
```

---

### 3.5 SchemaHandler — Thin Adapter

```go
// internal/web/handler/schema.go
//
// Route setup MUST include contract.InjectSessionContext() before this handler:
//   schema := app.Group("/schema",
//       iammiddleware.Authenticate(cfg),     // validates token, sets Fiber Locals
//       contract.InjectSessionContext(),       // injects contract.SessionContext into Go context
//   )
//   schema.Get("/*", schemaHandler.Handle)

type SchemaHandler struct {
    pipeline *pipeline.PipelineBuilder
}

func (h *SchemaHandler) Handle(c *fiber.Ctx) error {
    // Use contract.FromContext — NEVER c.Locals(iam.LocalsKeySession).
    // contract.InjectSessionContext() middleware has already bridged Fiber Locals
    // into the Go context. This is the canonical consumer pattern.
    sc, ok := contract.FromContext(c.UserContext())
    if !ok || sc.IsZero() {
        return c.Status(401).JSON(unauthenticatedAMISSchema())
    }

    opCtx := pipeline.AcquireOperationContext()
    defer pipeline.ReleaseOperationContext(opCtx)

    // opCtx.Ctx carries the contract.SessionContext — stages retrieve it via
    // contract.FromContext(opCtx.Ctx). This is the bridge between the contract
    // layer and the pipeline infrastructure.
    opCtx.Ctx = c.UserContext()

    // opCtx.Session is the pipeline infra field (*iam.ResolvedSession).
    // The pipeline package is allowed to import internal IAM domain types.
    // The UI pipeline STAGES must not use this field — they use opCtx.Ctx.
    // We still set it here so the pipeline's FeatureEnabled() delegation works.
    opCtx.Session      = sc.resolved   // accessible via unexported field accessor in adapter
    opCtx.TenantID     = sc.TenantID()
    opCtx.UserID       = sc.UserID()
    opCtx.OperationKey = uipipeline.OperationKey
    opCtx.Input        = uipipeline.UISchemaInput{Route: "/" + c.Params("*")}

    if err := h.pipeline.Run(opCtx); err != nil {
        var notFound *uipipeline.UINotFoundError
        if errors.As(err, &notFound) {
            return c.Status(404).JSON(fiber.Map{"status": 404, "msg": "Page not found"})
        }
        return c.Status(500).JSON(fiber.Map{"status": 500, "msg": "Schema compilation failed"})
    }

    out, ok := opCtx.Data["ui.response"].(uipipeline.UISchemaOutput)
    if !ok {
        return c.Status(500).JSON(fiber.Map{"status": 500, "msg": "Internal error"})
    }

    return c.JSON(fiber.Map{"status": 0, "data": out.Schema})
}
```

**`AcquireOperationContext()` + `defer ReleaseOperationContext()`** — existing pool pattern, zero extra allocation.

**`opCtx.Session = sc.resolved` note:** `contract.SessionContext.resolved` is unexported. The handler cannot set this directly. Two options:
1. Add `contract.SessionContext.ResolvedSession() *iam.ResolvedSession` — acceptable if the pipeline package is listed as a trusted consumer in `internal/core/iam/contract/export_test.go` (the file exists, likely for this purpose)
2. Skip setting `opCtx.Session` and have pipeline stages use `contract.FromContext(opCtx.Ctx)` exclusively — `FeatureEnabled()` delegation on `OperationContext` would need a contract-based fallback

Option 2 is cleaner and requires no export from the contract. The `OperationContext.FeatureEnabled()` method can be augmented to fall back to `contract.FromContext(opCtx.Ctx)` when `opCtx.Session == nil`.

---

### 3.6 Observability — Pipeline Stages vs Shared Infrastructure

The pipeline already records per-stage `StageLog` entries with `Duration`. This IS the observability backbone. Additional instrumentation hooks into the existing stage lifecycle:

```go
// internal/web/pipeline/stages/metrics.go
// ObservabilityStage wraps every other stage via a decorator pattern.
// OR: embed metrics emission directly in each stage's Execute() method.
// Recommendation: embed directly — less abstraction, same result.

// In each stage's Execute():
span, ctx := tracing.Start(opCtx.Ctx, "ui.stage."+s.Name())
defer span.End()
opCtx.Ctx = ctx  // propagate updated context

// After execution:
metrics.Record("ui_stage_duration_ms",
    float64(time.Since(started).Milliseconds()),
    "stage", s.Name(),
    "tenant", opCtx.TenantID.String(),
    "cache_hit", strconv.FormatBool(opCtx.Flag("ui.cache_hit")),
)
```

**Minimum required metrics** (emitted by stages, not the handler):

| Metric | Stage | Tags |
|---|---|---|
| `ui_pipeline_duration_ms` | SchemaHandler (wraps Run()) | route, tenant |
| `ui_iam_resolution_ms` | `ui.iam_resolve` | tenant |
| `ui_schema_compile_ms` | `ui.compile` | route, tenant |
| `ui_cache_hit_total` | `ui.cache_lookup` | route, tenant |
| `ui_cache_miss_total` | `ui.cache_lookup` | route, tenant |
| `ui_stage_failure_total` | Each required stage on error | stage, route |
| `ui_validation_error_total` | `ui.normalize` | route, rule |

**Structured log entries** from each stage are captured in `opCtx.Log` (existing `StageLog` array). SchemaHandler can flush these to the shared logger after `pipeline.Run()` returns:

```go
for _, entry := range opCtx.Log {
    logger.Info("ui.pipeline.stage",
        "stage", entry.StageName,
        "status", entry.Status,
        "duration_ms", entry.Duration.Milliseconds(),
        "error", entry.Error,
        "tenant", opCtx.TenantID,
        "route", input.Route,
    )
}
```

---

### 3.7 Cache Integration

Cache lives in `internal/platform/cache`. The pipeline accesses it through two stages:

```go
// internal/platform/cache — interface (existing or to be defined)
type SchemaCache interface {
    Get(key string) (map[string]any, bool)
    Set(key string, schema map[string]any, ttl time.Duration)
    Invalidate(tenantID string)
}

// Cache key construction — after IAM resolution (always)
func buildCacheKey(tenantID, route, permFingerprint, flagFingerprint string) string {
    return fmt.Sprintf("%s|%s|%s|%s", tenantID, route, permFingerprint, flagFingerprint)
}
```

**Cache invalidation trigger:** When a role is assigned/revoked, the IAM service calls `cache.Invalidate(tenantID)`. This flushes all schema cache entries for that tenant. Next request re-runs the pipeline from IAM resolution.

**Cache key correctness guarantees:**
- `tenantID` prefix → tenant isolation (never cross-tenant cache serving)
- `permFingerprint` → users with identical role sets share one entry
- `flagFingerprint` → flag change (e.g. enabling `advanced_reporting`) creates a new cache entry
- Route → different pages have different entries
- Key built AFTER stage 1 → fingerprint always reflects current IAM state

---

## 4. IAM Contract Rules — Enforced by Architecture

The `internal/core/iam/contract` rules (from `doc.go`) map to concrete code constraints:

### RULE 1 — No IAM import outside `contract` package boundary

```
ALLOWED in UI pipeline stages:
  contract.FromContext(opCtx.Ctx)           → SessionContext (identity + feature flags)
  opCtx.Data["ui.permissions"]              → map[string]bool (resolved by ui.iam_resolve)
  policy.UIAuthzService.BulkEnforce(sc,...) → called by ui.iam_resolve ONLY

FORBIDDEN in UI pipeline stages:
  import "awo.so/internal/core/iam"          → direct IAM domain import
  import "awo.so/internal/core/iam/service"  → explicitly forbidden by contract doc
  opCtx.Session.ToPrincipal()               → IAM-internal, must not be called by modules
  authzService.Enforce() directly            → explicitly forbidden by contract doc

ALLOWED in internal/web/policy (the UIAuthzService implementation):
  sc.resolved.ToPrincipal()                 → policy is a bridge package, not a module
  casbin.Enforcer.Enforce(...)              → the single allowed Casbin call site for UI
```

### RULE 2 — PageFn receives UISessionContext only

`UISessionContext` is a value type constructed from `contract.SessionContext` + resolved permissions. It exposes `Can()`, `Flag()`, `Pref()`, and identity fields. It cannot produce a `Principal`. It has no `EntityScope` method. No IAM type appears in its public API.

```go
// This code does not compile — UISessionContext has no ToPrincipal():
func BadPageFn(sess uipipeline.UISessionContext) web.Schema {
    sess.ToPrincipal()             // compile error — field does not exist
    sess.EntityScope()             // compile error — method does not exist
    sess.resolved.ToPrincipal()   // compile error — resolved is unexported
}
```

### RULE 3 — Feature flags from `contract.SessionContext` only

`sc.FeatureEnabled(key)` reads `session.Configuration.Flags` — pre-computed at login, no external call. In `UISessionContext`, `Flag(name)` reads a copy of those flags made at construction time. No feature flag service is called during UI compilation. This satisfies `doc.go`:

> MAY assume the request is already authorized by IAM middleware

Flag resolution happened at login. UI compilation reads the result. No flag service call during request handling.

---

## 5. Anti-Patterns — Strictly Forbidden

### AP-1: Direct PageFn execution outside pipeline

```go
// FORBIDDEN
schema := pageFn(sess) // no IAM resolution, no cache, no validation, no observability

// REQUIRED
pipeline.Run(opCtx) // always go through the pipeline
```

**Why it breaks:** Cache is never populated. IAM resolution is skipped. Validation is skipped. Observability gaps. Two users with different permissions could get the same schema if they somehow share an execution path.

---

### AP-2: Direct IAM domain access bypassing contract

```go
// FORBIDDEN in SchemaHandler:
sess := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession) // bypass contract
// REQUIRED:
sc, ok := contract.FromContext(c.UserContext()) // contract is the gate

// FORBIDDEN in any UI stage:
opCtx.Session.ToPrincipal() // direct IAM domain call
// REQUIRED:
sc, _ := contract.FromContext(opCtx.Ctx) // then pass sc to UIAuthzService

// FORBIDDEN in PageFn:
sess.authz.Enforce(...) // does not compile — UISessionContext has no authz field
```

**Why it breaks:** Bypassing `contract.FromContext` means the `contract.InjectSessionContext()` middleware can be removed from the route chain without breaking compilation — only runtime panics reveal the gap. The contract package is the enforcement point. Going around it makes the "all modules use contract" guarantee unverifiable at static analysis time.

---

### AP-3: Session mutation inside PageFn

```go
// FORBIDDEN:
func BadPageFn(sess UISessionContext) web.Schema {
    sess.permissions["invoice.delete"] = true  // fields are private — does not compile
    sess.featureFlags["advanced"] = true       // fields are private — does not compile
}
```

**Why it breaks:** The fingerprint computed in stage 1 would no longer match the state seen by the page function. Cache correctness collapses. This is prevented at the language level by unexported fields.

---

### AP-4: Cache lookup before IAM resolution

```go
// FORBIDDEN stage ordering:
// Priority 50: ui.cache_lookup    ← NO. Fingerprint not yet computed.
// Priority 100: ui.iam_resolve

// REQUIRED:
// Priority 100: ui.iam_resolve    ← fingerprint computed here
// Priority 150: ui.cache_lookup   ← fingerprint available
```

**Why it breaks:** Cache key lacks permission fingerprint. Two users with different roles could receive the same cached schema. This is a security defect, not just a correctness issue.

---

### AP-5: Schema generation outside pipeline stages

```go
// FORBIDDEN in SchemaHandler directly:
schema := pageRegistry.Get(route)(buildSessionCtx(sess))
return c.JSON(schema)
```

**Why it breaks:** No cache, no validation, no observability, no IAM resolution stage. Identical to AP-1. The handler must only build `OperationContext` and call `pipeline.Run()`.

---

### AP-6: Missing observability per stage

```go
// FORBIDDEN:
func (s *CompileStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
    schema := pageFn(sess)
    return pipeline.StageResult{Status: "completed", Outputs: ...}, nil
    // no span, no metrics, no duration
}
```

**Why it breaks:** Slow page functions are invisible in production. You cannot distinguish a slow IAM call from a slow PageFn. The pipeline's `StageLog.Duration` is populated automatically by the builder — but trace spans and counter metrics must be emitted by the stage itself.

---

## 6. Failure Mode Analysis

### F-1: Pipeline bypassed (handler calls PageFn directly)

**What breaks:**
- Cache never populated → every request recompiles → Casbin called N times/request → latency degrades under load
- SchemaValidator never runs → IAM-leaking expressions reach the browser → security defect
- No trace spans → zero observability
- Permission fingerprint not computed → permission changes not reflected in schema until session refresh

**Detection:** Integration test that asserts `SchemaHandler` calls `pipeline.Run`, not `compiler.Compile` or a registry directly. Enforce at code review.

---

### F-2: IAM contract bypassed (Casbin called in PageFn or downstream stage)

**What breaks:**
- Permissions resolved outside fingerprint → cache key does not include these permissions → two users with different permissions may share a cache entry
- If a PageFn calls Casbin directly: PageFn is no longer a pure function → same session may produce different schemas at different times → cache correctness undefined
- Adds undeclared Casbin latency per schema compilation

**Detection:** Lint rule forbidding import of `github.com/casbin/casbin` in `internal/web/` (except `internal/web/policy/`). Enforce in CI.

---

### F-3: Wrong cache key (missing tenantID, missing fingerprint)

**What breaks:**
- Missing `tenantID`: tenant A's schema served to tenant B — catastrophic data isolation failure
- Missing `permFingerprint`: read-only user receives admin user's cached schema — privilege escalation
- Missing `flagFingerprint`: user without `advanced_reporting` receives advanced metrics panel — feature leak

**Detection:** Unit test for `buildCacheKey()` asserting all four components present. Integration test asserting cache miss when tenant changes, when permissions change, when flags change.

---

### F-4: Stage ordering violated (cache before IAM)

Already covered under AP-4. Detection: test that asserts `ui.cache_lookup.Priority > ui.iam_resolve.Priority`.

---

### F-5: Validation stage removed or made non-required

**What breaks:**
- IAM-leaking `visibleOn` expressions reach the browser (e.g. `"visibleOn": "${role == 'admin'}"`)
- CRUD without `syncLocation` breaks URL-based pagination state
- Missing chart transparent background: visual inconsistency across dark/light themes

**Detection:** Do not set `StageRequired: false` on `ui.normalize`. Code review gate: `NormalizeStage` must always be `StageRequired: true`.

---

### F-6: Nil ResolvedSession reaches pipeline

**What breaks:**
- `IAMResolveStage` panics on `opCtx.Session.ToPrincipal()` → unhandled panic → 500 with no log
- Or returns empty permissions → page function renders as if all permissions are false → blank/degraded UI without clear error

**Detection:** `SchemaHandler` must validate `sess != nil && sess.IsValid()` BEFORE calling `pipeline.Run()`. This check is in place (see §3.5 handler code). Never remove it.

---

## 7. Final Verdict

### Is this production-grade?

**Yes, with conditions.** The architecture as specified is sound for a bank-grade ERP UI execution system when the following conditions hold:

1. `AllUIPermissions` is kept current with all `sess.Can()` calls in page functions
2. Cache is backed by Redis (not in-memory) for multi-instance deployments
3. Casbin enforcer has connection pooling and read replicas under load
4. `ui.iam_resolve` has a timeout and circuit breaker — a slow Casbin call must not block all UI rendering

---

### Where will it break first?

**`AllUIPermissions` maintenance** at scale.

This is the highest-probability failure. When a developer adds a new page with `sess.Can("approve", "purchase_order")` but forgets to add `purchase_order.approve` to `AllUIPermissions`, the IAM resolution stage will never evaluate that permission. `ui.permissions["purchase_order.approve"]` will always be `false`. The page will render as if the user cannot approve purchase orders — even for users who have that permission. This is a silent, permission-silent failure with no error.

Second failure point: **cache invalidation latency under high role-change throughput**. `cache.Invalidate(tenantID)` flushes ALL schema entries for a tenant on every role assignment. For a tenant with 500 users all changing roles simultaneously (e.g. a mass import), the cache drops to 0% hit rate until all users make their next request. This is acceptable for in-memory cache (local warmup). It is a thundering herd problem with Redis + many app instances.

---

### What MUST be enforced at code review level?

| Rule | How to enforce |
|---|---|
| No Casbin import in `internal/web/` outside `internal/web/policy/` | CI lint rule (`forbidigo` or custom AST check) |
| `NormalizeStage.StageRequired` must be `true` | Code review gate — never set false |
| `ui.cache_lookup.Priority` > `ui.iam_resolve.Priority` | Unit test asserting priority ordering |
| `SchemaHandler` must not call `registry.Get()` or `PageFn` directly | Integration test + code review |
| PageFn signature must be `func(UISessionContext) web.Schema` — no other parameter | Compile-time enforcement via `web.PageFn` type alias |
| `AllUIPermissions` must be updated when new `sess.Can()` calls are added | PR checklist item + schema validator cross-check |
| Cache key must include all four components | Unit test for `buildCacheKey()` |

---

### Biggest long-term architectural risk?

**The `AllUIPermissions` static list is an architectural time bomb.**

Currently it is a manually maintained slice in `internal/web/policy/engine.go`. As the system grows to 50+ pages, 200+ permission checks, and 10+ developers, the probability of `AllUIPermissions` diverging from actual page-function usage approaches 1.

The mitigation is self-registration (same pattern as page registry):

```go
// Each permission declares itself at init() time
func init() {
    policy.Register(InvoiceApprove) // "invoice.approve"
}
```

`AllUIPermissions` becomes a `registry.All()` call. A permission is added to the bulk-evaluation set automatically the moment a developer defines it and imports the package. The static list disappears.

This is the single structural change that upgrades the architecture from "production-grade with maintenance risk" to "self-maintaining at scale."

---

## Appendix — Stage Priority Reference (UI Pipeline)

| Priority | Stage | Required | Branch Jump |
|---|---|---|---|
| 100 | `ui.iam_resolve` | true | — |
| 150 | `ui.cache_lookup` | false | → `ui.response` on hit |
| 200 | `ui.registry_resolve` | true | — |
| 300 | `ui.compile` | true | — |
| 400 | `ui.normalize` | true | — |
| 450 | `ui.cache_store` | false | — |
| 500 | `ui.response` | true | — |

All stages declare `Operations() []string{"ui.schema.compile"}`. No wildcard — they do not fire for business operation pipelines.

---

## Appendix — Package Structure

```
internal/core/iam/contract/          ← EXISTS. Do not modify.
├── session.go    — SessionContext, FromContext, WithContext
├── auth.go       — AuthService interface
├── adapter.go    — SessionServiceAdapter
├── middleware.go — InjectSessionContext() Fiber middleware
└── doc.go        — contract rules (the law)

internal/web/
├── pipeline/
│   ├── types.go           — OperationKey, UISchemaInput, UISchemaOutput
│   ├── session_ctx.go     — UISessionContext, fromOperationContext()
│   │                        builds from contract.SessionContext + resolved perms
│   └── stages/
│       ├── iam_resolve.go  — calls contract.FromContext + UIAuthzService
│       ├── cache_lookup.go
│       ├── registry_resolve.go
│       ├── compile.go      — calls fromOperationContext(), executes PageFn
│       ├── normalize.go
│       ├── cache_store.go
│       └── response.go
├── policy/
│   ├── engine.go          — UIAuthzService interface, AllUIPermissions
│   └── casbin_engine.go   — CasbinEngine.BulkEnforce()
│                            ONLY place in internal/web that calls Casbin directly
│                            ONLY place allowed to call sc.resolved.ToPrincipal()
├── handler/
│   └── schema.go          — SchemaHandler: contract.FromContext → pipeline.Run
├── amis/                  — builder package (unchanged from §28)
├── registry/              — page registry (unchanged from §28)
└── pages/                 — page modules (unchanged from §28)

internal/pipeline/                   ← EXISTS. Unchanged.
├── builder.go    — PipelineBuilder.Run(), compensation, TxHooks
├── context.go    — OperationContext (pooled)
├── stage.go      — Stage interface, BaseStage, StageResult
└── stage_registry.go — StageRegistry

Import boundary (enforced by CI lint):
  internal/web/pipeline/stages/  → MAY import contract, pipeline, web/policy
  internal/web/pipeline/stages/  → MUST NOT import internal/core/iam (domain, service, repo)
  internal/web/pages/            → MUST NOT import contract or internal/core/iam
  internal/web/pages/            → receives UISessionContext only
  internal/web/policy/           → MAY import internal/core/iam (it is the bridge)
```

The existing `internal/pipeline/` package is unchanged. UI stages implement the existing `Stage` interface and register with the existing `StageRegistry`. No new pipeline infrastructure is needed.
