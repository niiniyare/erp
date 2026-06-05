# AwoERP UI Docs — Structured Audit

> Prepared against the actual codebase in `internal/web/`. All issues are specific and actionable.

---

## Section 1 — Factual Errors in Chapters 01–04

### 1.1 Client-Agnostic Framing

**Where:** Ch 01 §1.1, §1.4.1 ("One Backend, Many Surfaces"), Ch 04 §4.2.7–4.2.8
**Problem:** Chapters describe iOS (Swift/SwiftUI), Android (Kotlin/Jetpack Compose), and web as three equally live rendering targets. Code ships exactly one rendering client: the AMIS web SDK (`web/sdk/`). Flutter/React Native are planned but not implemented.
**Fix:** Lead with the AMIS reality. Add `> **Status: Planned — Not Yet Implemented**` callouts for mobile targets everywhere they appear. Remove SwiftUI/Kotlin code examples.

### 1.2 "OpenFGA" in the Authorization Description

**Where:** Ch 01 §1.8 (Glossary — OpenFGA entry), Ch 04 §4.4.5
**Problem:** `internal/web/ui/types.go` comment is explicit: `"permissions come from UIAuthzService.BulkEnforce() (Casbin, once per request)"`. The system uses Casbin. OpenFGA is mentioned nowhere in production code.
**Fix:** Replace all OpenFGA references with Casbin. Ch 04 §4.4.5 can note OpenFGA as a considered alternative.

### 1.3 gRPC Interface for UI Compilation Service

**Where:** Ch 04 §4.2.1 ("Interface: gRPC (`UICompilationService`) with REST facade")
**Problem:** There is no gRPC service for UI. The handler is `internal/web/handler/schema.go`, a plain Fiber HTTP handler mounted at `/schema/*`. No proto file, no gRPC stream.
**Fix:** Replace with accurate description: HTTP GET `/schema/<route>` via Fiber, with auth middleware chain.

### 1.4 CompilationContext Struct

**Where:** Ch 04 §4.3.2 (code block using `CompilationContext{Surface, Platform, Capabilities, ...}`)
**Problem:** This type does not exist. The actual type is `ui.UISessionContext` (`internal/web/ui/types.go`). It has no `Surface`, `Platform`, or `Capabilities` fields. It has `UserID`, `TenantID`, `DisplayName`, `IsPlatform`, `IsPortal`, `Locale`, `Timezone`, `Currency`, `Params`.
**Fix:** Replace the Go code block with actual `UISessionContext` construction from `NewUISessionContext()`.

### 1.5 Surface / SurfaceID Vocabulary

**Where:** Throughout Ch 01–04, the request URL example `GET /api/ui/surfaces/procurement.purchase-order.create`, glossary entry for "Surface"
**Problem:** The code uses the word "route", not "surface". The registry stores `Route string` (`internal/web/registry/registry.go:64`). The handler extracts `route := "/" + c.Params("*")`. There is no `SurfaceID` type anywhere in the codebase.
**Fix:** Use "route" everywhere. URL example becomes `GET /schema/finance/dashboard`.

---

## Section 2 — Pipeline Stage Errors (Chapter 08 predecessor text)

**Where:** Any documentation describing pipeline stages
**Problem:** Existing docs describe 6 stages. The real pipeline has 9 stages defined in `internal/web/ui/pipeline.go`.

Actual stages with priorities:

| Priority | Stage Name | Behaviour |
|----------|------------|-----------|
| 10 | SessionStage | Extract `contract.SessionContext` via `contract.FromContext` |
| 20 | AuthzStage | `UIAuthzService.BulkEnforce`; build `UISessionContext`; compute perm+flag fingerprints |
| 30 | CacheStage | Cache lookup — key = route + tenant + perm_fingerprint + flag_fingerprint. Short-circuits on hit |
| 40 | RegistryStage | `registry.Match(route)` → `PageFn` or `ASTPageFn`. Skipped on cache hit |
| 50 | CompileStage | Execute `PageFn(sess)` or `ast.CompileTree(ASTPageFn(sess))`. Skipped on cache hit |
| 60 | NormalizeStage | Canonicalize lowercase type, trim API whitespace. Never errors. Skipped on cache hit |
| 70 | ValidateStage | Structural + security rules; all errors are `BusinessError` with `VALIDATE_*` codes. Skipped on cache hit |
| 80 | CacheStoreStage | Write compiled schema to Redis. Skipped on cache hit |
| 90 | ResponseStage | Assemble `{"status": 0, "data": schema}`. Always runs |

**Critical ordering note:** CacheStage runs at priority 30, after AuthzStage (20). This is intentional — the cache key requires `perm_fingerprint` and `flag_fingerprint`, which AuthzStage computes. The old docs place cache lookup before authorization, which is architecturally wrong.

---

## Section 3 — Invented Architecture Not in Code

### 3.1 Slot Composition and Template Instantiation

**Where:** Ch 04 §4.1 diagram, glossary "Slot" entry, Ch 08 stage descriptions
**Problem:** `ast/node.go` defines `Node` and `ContainerNode`. There is no `Slot` type, no `SlotNode`, no `NewTemplateRef`, no slot composition mechanism anywhere in `internal/web/ast/`.
**Fix:** Remove slot references. The composition model is `Children() []Node` on `ContainerNode`. Document that.

### 3.2 Brotli/gzip Compression

**Where:** Ch 04 §4.3.4 step 6 ("`Content-Encoding: gzip`"), §4.6 scalability
**Problem:** No compression middleware is registered in the UI handler or pipeline. The `handler/schema.go` returns plain JSON via `c.JSON()`.
**Fix:** Remove compression from the request lifecycle example. Mark as planned if desired.

### 3.3 ETag / Cache-Control Headers

**Where:** Ch 04 §4.3.5 (response headers include `ETag`, `Cache-Control`, `X-Compilation-Time-Ms`, `X-Cache`)
**Problem:** None of these headers are set in `handler/schema.go`. The handler only sets status code and JSON body.
**Fix:** Remove these headers from examples. The current response is exactly `{"status": 0, "data": <schema>}`.

### 3.4 X-Client-Capabilities Header

**Where:** Ch 04 §4.3.1 (request includes `X-Client-Capabilities: charts,inline-edit,...`)
**Problem:** No capability detection logic exists anywhere in `internal/web/`. `UISessionContext` has no `Capabilities` field.
**Fix:** Remove. Mark client capability detection as planned.

### 3.5 Tenant Override Passes (Localization Injection, Flag Variant Replacement)

**Where:** Ch 04 §4.3.4 steps 2–4
**Problem:** These are described as pipeline passes. The pipeline stages in `pipeline.go` are the 9 stages above. There is no `LocalizationInjectionStage`, no `FlagVariantReplacementStage`, no `TenantOverrideStage`. Tenant config is read by `PageFn` if needed; it is not a pipeline transformation.
**Fix:** Move these to the `PageFn` author's responsibility. The pipeline handles caching, authz, compile, normalize, validate. Everything else is inside the page function.

---

## Section 4 — Missing Documentation (Gaps)

### 4.1 AMIS Response Envelope

The AMIS envelope `{"status": 0, "data": schema}` is central to how the web client works, but is not explained in any existing chapter. The 401 response also returns an AMIS envelope (not a plain error) so the browser renders a login page rather than a blank screen. This pattern (`handler/schema.go:194–218`) deserves a dedicated section.

### 4.2 DSL Blocks and Screens

`internal/web/dsl/blocks/` and `internal/web/dsl/screens/` are fully implemented and are the primary API for page authors. They are not mentioned in any existing documentation.

### 4.3 Registry Registration Pattern

The `init()` → `registry.RegisterPage()` pattern (`internal/web/registry/registry.go:9–17`) is the entry point for every new page, but is not documented.

### 4.4 Legacy PageFn vs ASTPageFn Migration

Both `PageFn` and `ASTPageFn` are active. The codebase is in migration. Docs must explain when to use which and that `CompileStage` prefers `ASTPageFn`.

---

## Section 5 — Go Code Examples That Do Not Compile

1. **Ch 04 §4.3.2** — `CompilationContext{Platform: PlatformIOS, Capabilities: CapabilitySet{...}}` — types do not exist.
2. **Ch 04 §4.3.3** — `ast.NewSurface(...)`, `ast.NewSingleColumnLayout()`, `ast.NewForm(...)`, `ast.NewVendorSelector(...)`, `ast.NewLineItemSection(...)`, `ast.NewSubmitAction(...)`, `ast.NewNavigateAction(...)`, `ast.ComputedBinding(...)`, `ast.ExprNow()` — none of these constructors exist in `internal/web/ast/`. The AST uses struct literals, not constructors. E.g., `ast.PageNode{Title: "...", Body: []ast.Node{...}}`.
3. **Ch 04 §4.3.3** — `i18n.Key(...)` — no `i18n` package exists in the codebase.
4. **Ch 04 §4.3.3** — `ast.NewFormSection("header").WithTitle(...)` — fluent builder API does not exist for AST nodes.

All AST construction uses struct literals. See `internal/web/dsl/screens/finance_dashboard.go` for the correct pattern.

---

## Section 6 — JSON Wire Format Errors

The JSON wire format shown in Ch 04 §4.3.5 uses:
- `"schema_version"`, `"surface_id"`, `"compiled_at"`, `"root"` — none of these fields exist in the actual response.
- Dotted type names like `"layout.single_column"`, `"form.container"`, `"form.section"`, `"field.vendor_selector"`, `"action.http_request"`, `"action.navigate"` — AMIS uses plain type names: `"page"`, `"form"`, `"crud"`, `"grid"`, `"chart"`, `"action"`.
- `"$expr"` binding syntax — AMIS uses `${expression}` string interpolation, not `{"$expr": "..."}` objects.

The actual response envelope is exactly:

```json
{
  "status": 0,
  "data": {
    "type": "page",
    "title": "Finance Dashboard",
    "initApi": { "method": "get", "url": "/api/v1/finance/dashboard/summary" },
    "body": [...]
  }
}
```

---

## Section 7 — Recommended Rewrite Priority

| Priority | Action |
|----------|--------|
| P0 | Rewrite pipeline stage documentation with the actual 9-stage model |
| P0 | Fix all Go code examples to use actual types (`UISessionContext`, struct literals, no fluent builders) |
| P0 | Fix the JSON wire format to the actual AMIS envelope |
| P1 | Replace OpenFGA with Casbin throughout |
| P1 | Remove or clearly mark planned: mobile clients, gRPC, compression, ETags, capabilities, slot composition |
| P1 | Add documentation for DSL blocks and screens |
| P2 | Align vocabulary: "route" not "surface", "PageFn/ASTPageFn" not "CompilationContext" |
| P2 | Document the AMIS response envelope and 401 behavior |
| P3 | Add registry registration pattern documentation |
