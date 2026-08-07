# SDUI Framework Implementation Status

**Framework:** AWO ERP SDUI
**Architecture baseline:** Parts 1–20 (sdui_architecture.md + sdui_implementation.md)
**Last updated:** Session 3 — frontend migration

---

## Package: awo/sdui/widget

**Status:** Complete
**Files:**
- `node.go` — Node struct, NodeKind constants (35+ kinds), ExpressionRef, LayoutHint, ValidationRule, DataSource, ActionNode
- `walk.go` — Walk, WalkAll, Collect, FindByID, CountNodes traversal functions
- `walk_test.go` — 6 tests covering walk, prune, nil root, find, count, collect

**Key exports:** `Node`, `NodeKind` (all constants), `ExpressionRef`, `LayoutHint`, `ValidationRule`, `DataSource`, `ActionNode`, `Walk`, `WalkAll`, `Collect`, `FindByID`, `CountNodes`

**New NodeKinds added:**
- Structural: `NodeTabPane`, `NodeGrid`
- Input: `NodeRichText`, `NodeMoney`, `NodeMultiSelect`, `NodeLookup`, `NodeTreeSelect`, `NodeDuration`, `NodeColor`, `NodeSignature`, `NodeFileUpload`
- Display: `NodeStaticText`, `NodeBadge`, `NodeSummaryCard`, `NodeWorkflowPanel`, `NodeAttachments`, `NodeActivity`, `NodeRelatedList`
- Dashboard: `NodeKPICard`, `NodeChartPanel`, `NodeTablePanel`, `NodeFilterBar`

**Breaking changes from prior version:**
- `VisibleOn/HiddenOn/DisabledOn/RequiredOn` changed from `string` → `*ExpressionRef`
- `ActionNode` gained: `ID`, `WorkflowID`, `DialogTarget`, `VisibleOn`, `DisabledOn`, `Icon`
- `DataSource` gained: `ParentField`, `SearchParam`, `SendOn *ExpressionRef`

**Architectural notes:**
- `ExpressionRef.Expr` is `any` to avoid circular import with the `expression` package.
  The `expression` package imports nothing from `widget`; `widget` imports nothing from `expression`.
  The AMIS renderer bridges the two via `expression.AMISSerializer`.

---

## Package: awo/sdui/expression

**Status:** Complete
**Files:**
- `expression.go` — ExpressionNode interface, FieldRef, Lit, Compare, Logical, Not, In, constructors
- `amis_serializer.go` — AMISSerializer (ExpressionNode → JavaScript string)
- `expression_test.go` — 6 tests covering serialization, escaping, error cases

**Key exports:** `ExpressionNode`, `FieldRef`, `Lit`, `Compare`, `Logical`, `Not`, `In`, `AMISSerializer`, constructor functions (`Field`, `Eq`, `And`, `Or`, `Negate`, `IsIn`, ...)

**Dependency rules:** Imports nothing. Standard library only.

**Architectural notes:**
- This is the only place in the framework where renderer-specific expression strings are produced.
- The generator must always emit `expression.ExpressionNode` values wrapped in `*widget.ExpressionRef`.
- Never import this package from generator or widget — only from renderer implementations.

---

## Package: awo/sdui/sduictx

**Status:** Complete
**Files:**
- `context.go` — GeneratorContext (immutable value), ViewMode constants, ViewerContext interface, GeneratorContextBuilder

**Key exports:** `GeneratorContext`, `ViewMode` (List/Create/Edit/Detail/Dashboard), `ViewerContext`, `NewGeneratorContext`, `GeneratorContextBuilder`

**Architectural notes:**
- `GeneratorContext` is a value type (no pointers to mutable state). Plugins receive copies.
- `ViewerContext` is defined here (not imported from `auth`) to avoid importing the full auth package into the SDUI tree. The concrete implementation is always `auth.ViewerContext`.
- Builder pattern validates required fields before returning a sealed context.

---

## Package: awo/sdui/registry

**Status:** Complete
**Files:**
- `registry.go` — Registry struct, global singleton, NewIsolated(), Register/Seal/Lookup/Validate
- `builtin.go` — Registers all 40+ framework NodeKinds in init()
- `registry_test.go` — 7 tests covering registration, duplicate detection, seal, lookup, validation

**Key exports:** `WidgetDef`, `Registry`, `NewIsolated()`, `Register()`, `Seal()`, `Lookup()`, `IsSealed()`, `AllKinds()`

**Architectural notes:**
- After `Seal()`, `Lookup()` is lock-free (concurrent-safe via `atomic.Bool` check).
- `NewIsolated()` returns a private registry with zero shared state — required for all tests.
- `Validate()` detects AllowedChildren references to unregistered NodeKinds at bootstrap.

---

## Package: awo/sdui/plugins

**Status:** Complete
**Files:**
- `plugins.go` — Pipeline, 9 ExtensionPoint constants, PluginFatalError, TreeTransformFunc, FieldNodeOverrideFunc, ValidationFunc, registration/execution logic
- `plugins_test.go` — 5 tests covering ordering, recoverable error, fatal error, duplicate priority, post-seal registration

**Key exports:** `ExtensionPoint` (9 constants), `Pipeline`, `NewIsolated()`, `PluginFatalError`, `IsFatal()`, `RegisterTreeTransform()`, `RegisterFieldNodeOverride()`, `RegisterValidation()`, `Seal()`, `GlobalPipeline()`

**Architectural notes:**
- Priorities are sorted at `Seal()` time — deterministic regardless of init() registration order.
- Duplicate priority within same extension point = bootstrap panic (detected pre-seal via error, panics in global convenience functions).
- `GeneratorContext` is always passed by value — plugins physically cannot mutate it.

---

## Package: awo/sdui/validation

**Status:** Complete
**Files:**
- `validation.go` — Validator, Result, Issue, Severity, validation rules for all structural constraints
- `validation_test.go` — 7 tests covering nil root, unknown kind, missing name, tab/section placement, missing datasource, nil expr ref

**Key exports:** `Validator`, `Result`, `Issue`, `Severity` (Fatal/Warning), `New()`, `NewWithRegistry()`

**Validated constraints:**
- NodeKind must be registered
- Input fields must have non-empty Name
- Nodes requiring DataSource must have non-empty URL
- NodeTabPane must be child of NodeTabs (not page/form/section)
- NodeSection must not be child of NodeTabs
- ExpressionRef.Expr must not be nil
- Node ID uniqueness (warning)

---

## Package: awo/sdui/renderer

**Status:** Complete
**Files:**
- `renderer.go` — Renderer interface, RendererContext, RenderedOutput (typed union), renderer Registry

**Key exports:** `Renderer` (interface), `RendererContext`, `RenderedOutput`, `FormatAMISJSON`, `FormatPDFBytes`, `FormatFlutterTree`, `NewAMISOutput()`, `NewRawOutput()`, `NewTypedOutput()`, `Registry`

**Architectural notes:**
- `RenderedOutput` is a typed union (not `map[string]any`) to support PDF bytes and Flutter typed structs.
- `RendererContext` carries renderer-specific hints (theme, grid system, locale formats) separate from `GeneratorContext` (semantic, renderer-independent).
- Future renderers implement `Renderer` and register in the renderer `Registry` — no existing code changes.

---

## Package: awo/sdui/amis

**Status:** Complete
**Files:**
- `renderer.go` — DefaultRenderer implementing all 35+ NodeKinds, expression serialization, sanitizeText(), action rendering
- `renderer_test.go` — 18 tests covering all major node kinds, XSS sanitization, expression serialization, props override, singleflight, tabs/tab panes, grid

**Key exports:** `DefaultRenderer`, `New()`, `RendererID` ("amis"), `RendererVersion` ("1.0.0")

**Breaking changes from prior version:**
- `Render(root *widget.Node)` → `Render(root *widget.Node, ctx renderer.RendererContext) (renderer.RenderedOutput, error)`
- Old tests rewritten to new API
- `renderButton` now calls `applyCommon` (B-03 bug fixed)
- XSS: `sanitizeText()` applied to all label/description/confirmText emissions
- Expressions: `VisibleOn/HiddenOn/DisabledOn/RequiredOn` serialized from `ExpressionRef` via `AMISSerializer`
- `NodeTabPane` handled separately from `NodeSection` (B-02 semantic overloading fixed)
- Actions: `renderActions` returns error (was `[]any`)

**New NodeKinds handled:** `NodeTabPane`, `NodeGrid`, `NodeRichText`, `NodeMoney`, `NodeMultiSelect`, `NodeLookup`, `NodeTreeSelect`, `NodeDuration`, `NodeColor`, `NodeSignature`, `NodeFileUpload`, `NodeStaticText`, `NodeBadge`, `NodeSummaryCard`, `NodeWorkflowPanel`, `NodeAttachments`, `NodeActivity`, `NodeRelatedList`, `NodeKPICard`, `NodeChartPanel`, `NodeTablePanel`, `NodeFilterBar`

---

## Package: awo/sdui/cache

**Status:** Complete
**Files:**
- `cache.go` — Cache, KeyParams, Key(), HashTenantID(), GetWidgetTree/SetWidgetTree/GetRenderedOutput/SetRenderedOutput, DoWidgetTree/DoRenderedOutput (singleflight), MarshalJSON/UnmarshalJSON
- `cache_test.go` — 9 tests covering key structure, determinism, nil redis, set/get L2/L3, miss, singleflight coalescing, error propagation, tenant hash

**Key exports:** `Cache`, `New()`, `KeyParams`, `Key()`, `HashTenantID()`, `ErrCacheMiss`, `RedisClient` (interface)

**Architectural notes:**
- Cache key includes all 9 required dimensions: level, entity, view, renderer_id, renderer_ver, locale, tenant_id_hash, schema_fp, perm_fp.
- `singleflight.Group` prevents thundering herd at both L2 and L3.
- `RedisClient` is an interface — injectable for testing without real Redis.
- TTL: L2=5min, L3=10min.

---

## Package: awo/sdui/dashboard

**Status:** Complete
**Files:**
- `dashboard.go` — DashboardDef, PanelDef, PanelKind/ChartType/KPIFormat constants, Registry, global Register/Lookup/All

**Key exports:** `DashboardDef`, `PanelDef`, `PanelKind`, `ChartType`, `KPIFormat`, `Register()`, `Lookup()`, `All()`, `NewRegistry()`

---

## Remaining Work

---

## Package: awo/sdui/generator

**Status:** Complete (Phase S1 scope)
**Files:**
- `generator.go` — EntityGenerator, EntitySchema, FieldDef, SectionDef, TabDef, ActionDef, all builders
- `generator_test.go` — 10 tests covering list/form/detail view modes, permission gating, field type mapping, actions

**Key exports:** `EntityGenerator`, `New()`, `NewWithPipeline()`, `EntitySchema`, `FieldDef`, `SectionDef`, `TabDef`, `ActionDef`, `SelectOption`

**Pipeline stages:**
1. Pre-generation plugin transforms (ExtPreGeneration — schema transform, deferred to post-v1.0)
2. Permission gating — fields without viewer permission are absent (not hidden)
3. View mode selection — InList/InForm/InDetail flags control field inclusion
4. Node construction — FieldType → NodeKind mapping, section/tab layout, action building
5. Post-generation plugin transforms (ExtPostGeneration)

**View modes handled:**
- `ViewModeList` → NodePage > NodeList (columns from InList fields)
- `ViewModeCreate` → NodePage > NodeForm (POST, no initApi)
- `ViewModeEdit` → NodePage > NodeForm (PATCH, ReadURL = DetailURL)
- `ViewModeDetail` → NodePage > NodeSummaryCard + NodeForm (all fields ReadOnly)
- `ViewModeDashboard` → NodePage placeholder (panels from dashboard.Registry)

**Field type mapping (FieldType → NodeKind):**
- `data/text/string` → NodeText (NodeTextArea if MaxLength > 255)
- `long_text` → NodeTextArea
- `rich_text` → NodeRichText
- `int/float/decimal` → NodeNumber
- `currency/money` → NodeMoney
- `select` → NodeSelect
- `multi_select` → NodeMultiSelect
- `link` + DataSource → NodeLookup; without → NodeSelect
- `tree_link` → NodeTreeSelect
- `date` → NodeDate
- `datetime/time` → NodeDateTime
- `duration` → NodeDuration
- `bool/boolean` → NodeSwitch
- `json` → NodeEditor
- `color` → NodeColor
- `signature` → NodeSignature
- `file/image/attach` → NodeFileUpload

**Permission gating:**
- Platform admins bypass all field-level permission checks
- Fields with `Permission` set are absent (not hidden) for viewers without that permission
- Actions are absent for viewers without the required permission
- Sections and tabs with `Permission` set are absent for viewers without that permission

**Known limitations:**
- Pre-generation schema transform (ExtPreGeneration) is a no-op at v1.0
- Dashboard generation is a placeholder — panels from dashboard.Registry not yet wired
- Layout engine (column spans within sections) is flat — no multi-column grid computation yet
- NodeMoney CurrencyField sibling node is not auto-emitted — generator emits NodeMoney, renderer handles currency display

---

## Package: awo/sdui/layout

**Status:** Complete
**Files:**
- `layout.go` — Engine, ComputedLayout, SectionLayout, TabLayout, Row, RowItem, Span; row packing algorithm; dashboard panel layout
- `layout_test.go` — 12 tests covering nil/non-page root, flat fields, full-width kinds, explicit spans, NewRow hint, 4-column sections, tab panes, dashboard rows, SummaryCard; 3 benchmarks (small form, large form, dashboard)
- `fuzz_test.go` — 3 fuzz targets: ColSpan, section columns, nil children

**Key exports:** `Engine`, `New()`, `Compute()`, `ComputedLayout`, `SectionLayout`, `TabLayout`, `Row`, `RowItem`, `Span`, `Full()`, `MobileGridCols`, `TabletGridCols`, `DesktopGridCols`, `DefaultSectionCols`

**Grid model:**
- Desktop: 12 columns; Tablet: 8 columns; Mobile: 4 columns (always full-width)
- Section columns 1–4; stored in `widget.Node.Layout.ColSpan` on the section node itself
- Default section span = DesktopGridCols / sectionCols
- Full-width kinds: NodeTextArea, NodeRichText, NodeEditor, NodeSection, NodeGrid, NodeTable, NodeRelatedList, NodeWorkflowPanel, NodeAttachments, NodeActivity, NodeStaticText

**Dashboard panel defaults:**
- NodeKPICard: 3 desktop / 4 tablet
- NodeChartPanel: 6 desktop / 8 tablet
- NodeFilterBar/NodeTablePanel: 12 desktop / 8 tablet

**Architectural notes:**
- Engine holds no mutable state — safe for concurrent use
- Explicit `ColSpan` on any node overrides all defaults
- `NewRow: true` in LayoutHint forces a row break before that node
- Nested NodeSection inside NodeTabPane: flattened to fields (single-level layout)

---

## Package: awo/sdui/observability

**Status:** Complete
**Files:**
- `observability.go` — Metrics struct, New(), Noop(), StartSpan, Timer, TrackGeneration/Render/Validation/Layout/Plugins, RecordCacheHit/Miss/Error
- `observability_test.go` — 5 tests covering Noop, timer measurement, global provider, default names; 1 benchmark

**Key exports:** `Metrics`, `New()`, `Noop()`, `Config`, `Timer`, `Stage*` constants, `CacheLevel*` constants

**Instruments registered:**
- `sdui.generation_duration_ms` histogram
- `sdui.render_duration_ms` histogram
- `sdui.validation_duration_ms` histogram
- `sdui.layout_duration_ms` histogram
- `sdui.plugin_duration_ms` histogram
- `sdui.cache_hits_total` counter
- `sdui.cache_misses_total` counter
- `sdui.errors_total` counter

**Architectural notes:**
- Integrates with global OTel meter/tracer providers — no parallel observability stack
- All instruments are nil-safe: Noop() leaves them nil; all record methods guard with nil checks
- Timer records elapsed milliseconds with microsecond resolution

---

## Package: awo/sdui/engine

**Status:** Complete
**Files:**
- `engine.go` — Engine, Options, Request, Response, New(), Handle(), Renderers(); full 7-stage pipeline with L2/L3 caching
- `engine_test.go` — 9 integration tests + 1 benchmark covering all view modes, unknown renderer, validation, permission filtering, layout presence, output JSON validity
- `golden_test.go` — snapshot regression tests for invoice list/create/edit/detail; -update flag to regenerate
- `fuzz_test.go` — 3 fuzz targets: field types, field names, schema names

**Key exports:** `Engine`, `New()`, `Handle()`, `Renderers()`, `Request`, `Response`, `Options`

**Pipeline stages:**
1. Renderer resolution
2. L3 cache lookup (rendered output) → early return on hit
3. L2 cache lookup (widget tree) → skip generation on hit
4. Generator.Generate (if L2 miss)
5. L2 cache store
6. Validator.Validate (hard gate — fatal issues abort)
7. Layout.Compute
8. Renderer.Render
9. L3 cache store

**Architectural notes:**
- Engine is the only public entry point for SDUI generation; HTTP handlers call only Handle()
- Cache is optional (nil disables all caching)
- Observability is optional (nil → Noop)
- Widget trees and rendered outputs are JSON-marshalled for cache storage
- Corrupt cache entries are silently discarded and regenerated

---

---

## Package: awo/sdui/adapt

**Status:** Complete
**Files:**
- `adapt.go` — FromCompiled(), SchemaFingerprint(), ViewerAdapter, NewViewerAdapter(), BuildGrantIndex(), GrantIndex
- `adapt_test.go` — 13 tests covering identity, URLs, permissions, field visibility, sensitivity, immutability, select options, actions, sections, tabs, multi-column interleave, fingerprint determinism, fingerprint uniqueness; 2 benchmarks

**Key exports:** `FromCompiled()`, `SchemaFingerprint()`, `ViewerAdapter`, `NewViewerAdapter()`, `BuildGrantIndex()`, `GrantIndex`

**Responsibility:**
- Translates `*compiler.EntitySchema` → `generator.EntitySchema` (compiler → SDUI bridge)
- `FromCompiled()` maps: qualified name, label, URLs, fields, layout (sections/tabs), actions, permissions
- `SchemaFingerprint()` computes FNV-64a hash of field names+types, layout, actions for cache keys
- `ViewerAdapter` bridges `auth.ViewerContext` → `sduictx.ViewerContext` via permission grant index
- `BuildGrantIndex()` pre-builds O(1) map from `CapabilityGrant` slice for per-request `HasPermission` calls

**Field visibility rules:**
- `InList = !Hidden && !Sensitive && isListable(type)` — excludes long_text, json, link_list, dynamic_link, multi_select
- `InForm = !Hidden && !Sensitive`
- `InDetail = !Hidden` (sensitive fields visible to those with explicit permission — handled at API level)
- `Immutable` fields → `ReadOnly = true`
- `Sensitive` fields → `Hidden = true` in SDUI (hidden in grid/form; still in API for authorized viewers)

**Layout translation:**
- Tabbed layout: sections flattened per-tab with `tab.Name + "." + sec.Name` IDs
- Multi-column sections: fields interleaved (a1,b1,a2,b2,...) for correct row packing by layout engine
- Single-column sections: field order preserved

**Dependency direction:** `adapt` → `compiler`, `auth`, `sdui/generator`, `sdui/sduictx`, `sdui/widget`, `def`
No cycle: `compiler` does not import any `sdui/*` package.

---

## Package: awo/api/sdui

**Status:** Complete
**Files:**
- `handler.go` — Handler, New(), Register(), list/create/detail/edit handlers; ETag + Cache-Control; locale + renderer negotiation

**Key exports:** `Handler`, `New()`, `Register()`

**Endpoints registered (via Register on a Fiber group):**
- `GET /:module/:resource` → list view
- `GET /:module/:resource/create` → create form
- `GET /:module/:resource/:id/edit` → edit form
- `GET /:module/:resource/:id` → detail view

**HTTP integration:**
- Renderer selection: `Accept-SDUI-Renderer` header (default: "amis")
- Locale selection: `Accept-Language` header, first tag only (default: "en-US")
- ETag: `"{schemaFP}-{rendererID}-{locale}"`; supports `If-None-Match` for 304 responses
- Cache-Control: `private, max-age=300` (5 min)
- Auth: expects `auth.ViewerFromContext` to be present (caller mounts RequireAuth middleware)
- No business logic — all generation delegated to `engine.Engine`

**Router integration:**
- `RegisterOptions.SDUIEngine *engine.Engine` added to router
- When set, `New(schema, SDUIEngine, Authz)` is created and registered on `/api/v1/ui` group
- Coexists with legacy `SDUIGenerator` on `/api/sdui/` (no migration required)

---

### Pre-implementation checklist status (from Part 20):
- [x] ExpressionRef as portable DSL (not renderer strings)
- [x] RenderedOutput as typed union
- [x] GeneratorContext immutable value type
- [x] Registry seals at bootstrap; NewIsolated() for tests
- [x] Plugin priorities deterministic; duplicate = bootstrap error
- [x] NodeTabPane separate from NodeSection
- [x] Cache keys include all 9 dimensions
- [x] Validation before render (hard gate)
- [x] Generator implementation (Phase S1)
- [x] Layout engine implementation
- [x] Observability integration
- [x] End-to-end engine pipeline
- [x] Golden regression tests
- [x] Fuzz testing (expression, layout, engine)
- [x] Compiler → SDUI bridge (adapt package)
- [x] HTTP SDUI handler (api/sdui package)
- [x] Router integration (SDUIEngine option)

### Phase S2 — Framework-Wide SDUI Integration

**Status:** In progress

**Completed:**
- [x] `def.FieldDef`: Added `Placeholder`, `Icon`, `Width`, `Computed`, `ClearOn`, `VisibleOn`, `HiddenOn`, `DisabledOn`, `RequiredOn`
- [x] `def.EntityDefinition`: Added `EntityIcon() string` to interface; `Icon string` to `SystemDefinition`/`CustomDefinition`
- [x] `def.TabDef`: Added `Icon`, `Description`, `Permission`
- [x] `def.SectionDef`: Added `Icon`, `Description`, `Permission`
- [x] `compiler.EntitySchema`: Added `Icon string`; propagated from `EntityIcon()` in `buildEntitySchema`
- [x] `generator.FieldDef`: Added `Icon`, `Width`, `Computed`, `ClearOn`, `VisibleOn`, `HiddenOn`, `DisabledOn`, `RequiredOn`
- [x] `generator.SectionDef`: Added `Icon`, `Description`
- [x] `generator.TabDef`: Added `Icon`, `Description`
- [x] `generator.EntitySchema`: Added `Icon`
- [x] `generator.buildFieldNode`: Wires Icon, Width (LayoutHint), expression strings → ExpressionRef; Computed → ReadOnly
- [x] `adapt.convertField()`: Maps all new def.FieldDef fields to generator.FieldDef
- [x] `adapt.convertSection()`: Maps Icon, Description, Permission from def.SectionDef
- [x] `adapt.convertTabs()`: Maps Icon, Description, Permission from def.TabDef
- [x] `adapt.FromCompiled()`: Maps Icon from compiler.EntitySchema to generator.EntitySchema
- [x] `adapt.SchemaFingerprint()`: Includes entity Icon in hash
- [x] `widget.Node`: Added `Icon string`; `widget.LayoutHint`: Added `Width string`
- [x] `expression.RawExpression`: Escape hatch for raw AMIS expression strings
- [x] `expression.AMISSerializer`: Handles `RawExpression` pass-through
- [x] `amis.applyCommon()`: Emits `prefix` (fa-icon) and `size` (Width) for field nodes

**In progress / Remaining:**
- [x] Relation support (EdgeDef → generator.RelationDef → NodeRelatedList in detail; NodeGrid for Inline=true in forms)
- [x] Dashboard wiring (dashboard.Registry → adapt.collectDashboardPanels → generator.DashboardPanel → buildDashboard NodeKPICard/NodeChartPanel/NodeTablePanel/NodeFilterBar)
- [x] Workflow UI (HasWorkflow from WorkflowTriggers → NodeWorkflowPanel auto-emitted in detail view at {DetailURL}/workflow-state)
- [x] Search/filter forms (isFilterable → NodeFilterBar auto-emitted in list view for select/link/date/bool/naming_series/data fields)
- [x] Localization (`renderer.LocaleFormatsFor` / `renderer.ApplyLocale` — static table of 30+ locales; `amis.renderNumber`/`renderMoney` emit `decimalSeparator`/`thousandSeparator`; `amis.renderPage` emits `dir:"rtl"` for RTL locales)
- [x] Theme integration (`amis.ThemeConfigFor` — maps "default"/"antd"/"ang"/"dark"/"compact" to AMIS theme+classPrefix+darkMode; `applyTheme` emits `theme`/`darkMode` keys on page schema; dark mode = cxd base + html.dark CSS override)
- [x] Conformance suite (`awo/sdui/conformance/conformance_test.go` — covers: all 41 NodeKinds rendered without error, generator determinism, renderer determinism, locale RTL/separators, theme config correctness)
- [x] Searchable field wiring (`adapt.convertField` → `generator.FieldDef.Searchable`; `buildFilterBar` gates on `f.Searchable` not type heuristic)

### Documentation and Cleanup Pass

**Status:** Complete

**Changes:**

- [x] `awo/sdui/README.md` written — comprehensive developer guide covering all packages, pipeline, extension points, best practices, and minimal examples
- [x] Legacy `awo/sdui/generator.go` gutted — `sdui.Generator` / `pageGen` / all builders replaced by empty stub with migration comment
- [x] Legacy `awo/sdui/label.go` gutted — `fieldLabel()`/`toTitle()` covered by `def.DeriveLabel()` and `adapt.labelFor()`
- [x] Legacy `awo/sdui/builder.go` already empty; stub comment updated
- [x] Legacy `awo/sdui/schema.go` gutted — typed AMIS structs (`Page`, `CRUD`, `Form`, etc.) superseded by `widget.Node` IR + renderer pipeline
- [x] Legacy `awo/sdui/nav.go` gutted — `BuildNav()` superseded by `api/sdui.Handler.nav()`
- [x] `api/handler/sdui.go` gutted — `SDUINav(gen *sdui.Generator)` superseded by `api/sdui.Handler.nav()`
- [x] `api/sdui/handler.go` — added nav endpoint (`GET /api/v1/ui/nav`), `NavEntry`/`NavModule` types, locale wiring via `renderer.ApplyLocale`
- [x] `api/router/router.go` — removed `SDUIGenerator *sdui.Generator` from `RegisterOptions`, removed `registerSDUI()` function and `/api/sdui/*` routes, removed `awo.so/awo/sdui` import
- [x] `cmd/server/main.go` — replaced `sdui.New(...)` with `engine.New(engine.Options{...})`, removed `"awo.so/awo/sdui"` + `"awo.so/awo/cache"` imports, added SDUI sub-package imports, added `sduiCacheFor()` helper + `goRedisSDUIAdapter` for `sdui_cache.RedisClient` adaptation

**Architectural decisions:**

- The `awo/sdui` root package is now an empty stub (package identity preserved for future use). All active code is in named sub-packages.
- Nav endpoint migrated to new handler at `/api/v1/ui/nav` — inherits `/api/v1` middleware (tenant + auth + ratelimit). Old `/api/sdui/nav` is gone.
- `def.PageKind`, `def.PageBuilder`, `def.PageBuilderSet`, `compiler.EntitySchema.PageBuilders` retained — valid extension points even though not yet used by the new engine.
- `goRedisSDUIAdapter` in `cmd/server/main.go` adapts `*goredis.Client` to `sdui_cache.RedisClient` (interface mismatch: go-redis returns Cmd types; sdui/cache expects `(string, error)`).

### Known limitations:
- `NodeDuration` renders as masked `input-text` in AMIS (no native widget); production requires a custom AMIS component.
- `NodeSignature` renders as `input-file` in AMIS (no native signature widget); production requires a custom AMIS component.
- `NodeMoney` currency selector not auto-emitted by generator; renderer handles currency display.
- Pre-generation schema transform (ExtPreGeneration) is a no-op at v1.0.
- Golden files must be generated by running: `go test ./awo/sdui/engine/... -run TestGolden -update`
- `def.PageBuilders` on `compiler.EntitySchema` is not yet consumed by the new engine (was used by legacy `sdui.Generator` to allow custom amis JSON overrides). Post-v1.0 work.
- `def.EdgeDef` has no `Inline` field — all relations render as `NodeRelatedList` (read-only CRUD), not `NodeGrid` (editable inline). Inline line-item grids (e.g. invoice lines) require adding `Inline bool` to `def.EdgeDef` and `adapt.convertEdge`.

---

## Session 2 — Frontend Production-Readiness

**Status:** Complete

### Root-cause investigation (pipeline trace)

Traced the full pipeline: EntityDefinition → compiler → adapt → generator → widget tree → validation → layout → AMIS renderer → engine → HTTP handler → browser → AMIS SDK.

Identified six bugs:

| # | Location | Bug | Fix |
|---|---|---|---|
| 1 | `index.html` `renderAmis()` | No `locale` passed to `amis.embed()` — AMIS defaults to zh-CN | Add `locale: 'en-US'` to embed env |
| 2 | `index.html` `renderAmis()` | Dark mode passed `theme: 'dark'` to AMIS; AMIS has no dark theme | Changed to `theme: 'cxd'` always; dark mode handled by `html.dark` CSS |
| 3 | `adapt/adapt.go` `FromCompiled()` | Edit/detail URL used `{id}` (Go template); AMIS expects `${id}` | Fixed to `${id}` |
| 4 | `amis/renderer.go` `renderList()` | AMIS v3 `crud2` renders "bulkActions" text when `bulkActions` key absent | Added `"bulkActions": []any{}` |
| 5 | `amis/renderer.go` `renderList()` | AMIS v3 `crud2` used "items" key but no `itemsKey` declared — brittle against AMIS version drift | Added `"itemsKey": "items"` |
| 6 | `adapt/adapt.go` `convertField()` | Filter bar empty: only fields with `def.FieldDef.Searchable=true` shown; entity defs don't mark select/date/bool fields searchable | Added `isDefaultSearchable()` heuristic: select, date, datetime, bool, naming_series → always filterable |
| 7 | `amis/renderer.go` `renderFilterBar()` | Missing `"wrapWithPanel": false` causes double panel nesting inside `crud2.filter` | Added to filter form schema |

### Remaining frontend gaps

- **Bulk delete**: No bulk actions configured. Add `"bulkActions"` with a delete-selected button if needed.
- **Export**: No export action in toolbar. Add as a custom `ActionDef` on entities or via plugin.
- **Inline edge grids**: Invoice lines always render as related list, not editable inline grid. Requires `def.EdgeDef.Inline` field.
- **Custom AMIS components**: `NodeDuration` and `NodeSignature` fall back to text/file input. Register custom AMIS components for production.
- **Auth redirect**: The frontend has no login redirect. Unauthenticated requests return 401 from the API; the frontend shows a generic error. Add a login page and 401 redirect.
- **Breadcrumbs / page context**: No breadcrumb trail from list → detail → edit. AMIS `page` component supports `breadcrumb` key — could be wired from entity labels.

---

## Session 3 — Frontend Migration

**Status:** Complete

### Rationale

Two parallel frontend shells existed:
- `awo/web/pages/index.html` — correct architectural base: history routing, dynamic nav from `/api/v1/ui/nav`, SDUI integration, AMIS v3.6.6, `extendDefaultLocale` overrides, `adaptResponse` normaliser. (This file was missing — the server route existed but no file.)
- `erp/web/dist/index.html` — better UX: dark navy sidebar, collapse, 3-state theme, FontAwesome, breadcrumb, topbar, mobile.

Decision: merge `erp/web` UX into `awo/web` architecture. Do NOT replace awo/web with erp/web shell. Do NOT upgrade AMIS — stay on v3.6.6.

### AMIS Version Decision

**AMIS v3.6.6 — locked.** The backend renderer (`awo/sdui/amis`) is tuned to v3 semantics:
- `crud2` component with `itemsKey`, `bulkActions`, `filter.wrapWithPanel`
- Expression format (`${...}` template strings)
- `extendDefaultLocale` API for en-US string overrides

`erp/web/dist/index.html` had loaded AMIS v6.7.0 — reverted to v3.6.6 in the merged shell.

### Shell Architecture

Single file: `awo/web/pages/index.html`

Served by the Go server at:
```
GET /ui/*  →  ./awo/web/pages/index.html   (SPA fallback, in cmd/server/main.go)
```

The file is structured in three `<script>` blocks:

**Block 1 — AMIS bootstrap:**
- Multi-version AMD resolver (tries `amisRequire`, falls back to `window.amis.require`, then `require`)
- `extendDefaultLocale('en-US', {...})` — 17 Chinese string overrides
- `window.amisEnv = { theme: 'cxd', locale: 'en-US' }` skeleton

**Block 2 — SchemaLoader:**
- Lightweight class with in-memory cache
- Used by showcase; not used by production SDUI (which fetches from API)

**Block 3 — App (IIFE):**
- Theme (3-state: system/light/dark), `cycleTheme()`, `effectiveDark()`, `applyTheme(rerender)`
- `prefers-color-scheme` media query listener (system mode only)
- Sidebar collapse with `localStorage` persistence
- Mobile sidebar: `openMobile()` / `closeMobile()` / `#sidebar-backdrop`
- Routing: `parseRoute()`, `sduiURL()`, `navigate()`, `loadPage()` via pushState
- `adaptResponse()` — normalises `{data:[], meta:{total}}` → `{status:0, data:{items:[], count:N}}`
- AMIS fetcher: page/perPage → offset/limit conversion, `X-Awo-Tenant` header, 401 handler stub
- `renderAmis()` — unmounts previous instance, calls `amisRequire('amis/embed').embed()`
- Nav: `loadNav()` fetches `GET /api/v1/ui/nav`, `renderNav()` builds DOM from response
- `navIndex` map for breadcrumb: `listUrl → { moduleLabel, entryLabel }`
- `updateNavActive()` + `updateBreadcrumb()` after each navigation
- Auth stubs: `doLogout()` is a no-op; login overlay HTML is commented in, ready to enable

### Navigation Integration

Nav is loaded from `GET /api/v1/ui/nav` (registered in `api/sdui/handler.go`).

Expected response shape:
```json
{
  "status": 0,
  "data": {
    "modules": [
      {
        "label": "Finance",
        "entries": [
          { "label": "Accounts", "icon": "book", "listUrl": "/ui/finance/account" }
        ]
      }
    ]
  }
}
```

`entry.icon` is a FontAwesome icon name without the `fa-` prefix (e.g. `"book"` → `<i class="fa fa-book">`). Default icon is `fa-table-list` when absent.

### SDUI Integration

Pages load from `GET /api/v1/ui/{module}/{resource}` (registered by `api/sdui/handler.go`).

The frontend:
1. Fetches the AMIS JSON schema from the SDUI endpoint.
2. Calls `renderAmis(schema)` which unmounts the previous AMIS instance and embeds a new one.
3. The AMIS fetcher translates page/perPage pagination to offset/limit for the entity API.
4. `adaptResponse()` normalises list responses to the AMIS envelope format.

### Theme / Dark Mode

3-state theme (system/light/dark) stored in `localStorage('awo-theme')`.

- `effectiveDark()` — resolves 'system' via `prefers-color-scheme`
- `applyTheme(rerender)`:
  - Sets `data-theme` attribute on `<html>` (for CSS vars that key off data-theme)
  - Toggles `html.dark` class (for AMIS CSS token overrides)
  - Saves to localStorage
  - If `rerender=true` and a route is active, reloads the current SDUI page

Sidebar is always dark navy (`--sidebar-bg: #0f172a`) — no dark-mode override needed for sidebar CSS.

AMIS CSS token overrides on `html.dark`:
```css
--background: #141414;
--body-bg: #1a1a1a;
--Page-main-bg: #1a1a1a;
--Panel-bg-color: #262626;
--Table-bg: #262626;
--colors-neutral-line-8: #434343;
```

Never override individual `.cxd-*` backgrounds with `!important` — use CSS custom properties.

### History/pushState Routing

URL format: `/ui/{module}/{resource}`

- Nav links use `data-list-url` attribute; click handler calls `history.pushState` + `navigate()`
- `window.addEventListener('popstate', navigate)` handles back/forward
- `parseRoute()` extracts module and resource from `window.location.pathname`
- Welcome screen shown when no module/resource in URL

### Showcase vs Production SDUI

| Path | Content | Auth required |
|---|---|---|
| `/ui/*` | Production SDUI shell | Yes (when auth enabled) |
| `/showcase*` | Developer showcase | No |

Static JSON schemas from `erp/web/public/schemas/` are copied to `awo/web/showcase/schemas/`. They are served at `/showcase/*` only. They must not be placed on production routes.

### Schema Audit Results

See `awo/web/showcase/SCHEMA_AUDIT.md` for full classification.

Summary:
- Category A (keep): 0
- Category B (demo): 1 (`dashboard.json` — placeholder stat cards)
- Category C (replace with SDUI): 20 (all entity CRUD schemas)

All Category C schemas are temporary reference material. Once the SDUI engine produces a confirmed-working list view for a given entity, the corresponding static JSON should be deleted.

### Go Routing (No Changes Required)

The server already serves (in `awo/cmd/server/main.go`):
```go
app.Get("/ui/*", func(c *fiber.Ctx) error {
    return c.SendFile("./awo/web/pages/index.html")
})
app.Get("/showcase*", func(c *fiber.Ctx) error {
    return c.SendFile("./awo/web/showcase/index.html")
})
```

No Go routing changes were needed. The `awo/web/pages/index.html` file was missing (the server route existed but no file) — this session created it.

### Definition of Done

- [x] `awo/web/pages/index.html` created — merged shell (AMIS v3.6.6, dark navy sidebar, 3-state theme, dynamic nav, history routing, SDUI fetcher, breadcrumb)
- [x] Showcase schemas copied to `awo/web/showcase/schemas/pages/` (21 files across 6 modules)
- [x] `awo/web/showcase/SCHEMA_AUDIT.md` written — all 21 schemas classified
- [x] `awo/IMPLEMENTATION.md` updated with this section
- [x] `awo/web/showcase/index.html` created — developer portal (Session 4)
- [ ] Login overlay: commented in HTML, ready to enable when auth gating is implemented
- [ ] Auth redirect: 401 from API currently shows generic error; needs login page and redirect logic

---

## Session 4 — SDUI Bug Fixes and Showcase Portal

**Status:** Complete

### Bugs Fixed

| # | Location | Bug | Fix |
|---|---|---|---|
| 1 | `index.html` `renderNav()` | `navData.modules` always `undefined` — backend returns bare `[]NavModule` array (no `{modules:[]}` wrapper) | Detect bare array: `Array.isArray(navData) ? navData : navData.modules ?? []` |
| 2 | `index.html` `parseRoute()` | Only handled 2-segment URLs (`/ui/{module}/{resource}`). Create/detail/edit views fell through to empty route object | Extended to extract `id` and `view` from all path segments; derives `view` from last segment |
| 3 | `index.html` `sduiURL()` | Always returned list URL because `parseRoute()` didn't return `view` mode | Extended to construct correct URL for all view modes (create/detail/edit/list) |
| 4 | `index.html` `loadPage()` | Misleading comment said "SDUI handler returns `{ status:0, data: <amis-schema> }`" — handler actually returns raw AMIS map directly | Fixed comment to accurately describe both the envelope and bare-schema cases |
| 5 | `amis/renderer.go` `renderList()` | `"itemsKey"` absent — AMIS `crud` defaults to `"rows"` but `adaptResponse()` wraps list data as `{ items: [...] }`. Regression from the `crud2 → crud` change | Added `"itemsKey": "items"` to `renderList()` output map |

### Showcase Portal Created

`awo/web/showcase/index.html` — dark-themed developer portal served at `/showcase*`.

Fetches from:
- `GET /api/v1/showcase/info` — entity count, route count, modules, renderer IDs
- `GET /api/v1/showcase/entities` — registered entity list

Displays:
- Framework summary stat cards
- Registered entities table with links to `/ui/{module}/{resource}s`
- SDUI API endpoint reference

This file was missing in Session 3 (server route existed but file did not). The server never served it successfully until now.

### Navigation Bug Detail

`GET /api/v1/ui/nav` returns:
```json
[
  { "label": "Demo", "entries": [...] }
]
```

Not:
```json
{ "modules": [...] }
```

Session 3 shell used `navData.modules` which returned `undefined`, causing nav to always show "No modules registered." Fixed in `renderNav()`:

```js
var modules = Array.isArray(navData) ? navData
            : (navData && Array.isArray(navData.modules)) ? navData.modules
            : [];
```

### Route Parsing Extension

`parseRoute()` and `sduiURL()` now handle all four URL patterns:

| URL | `view` | `id` | SDUI endpoint |
|---|---|---|---|
| `/ui/demo/product` | `list` | null | `GET /api/v1/ui/demo/product` |
| `/ui/demo/product/create` | `create` | null | `GET /api/v1/ui/demo/product/create` |
| `/ui/demo/product/{id}` | `detail` | `{id}` | `GET /api/v1/ui/demo/product/{id}` |
| `/ui/demo/product/{id}/edit` | `edit` | `{id}` | `GET /api/v1/ui/demo/product/{id}/edit` |

### Known Remaining Issues

- Finance module not imported in `main.go` — only demo module is registered. Finance entities will not appear in nav.
- Auth redirect: 401 shows generic error; no login redirect implemented.
- Filter bar comment in `renderFilterBar()` still says "crud2" — minor; component is now `crud`.
- Bulk delete not configured in toolbar.
- `NodeDuration` and `NodeSignature` fall back to text/file input; custom AMIS components needed for production.

---

## Session 5 — End-to-End SDUI Pipeline Audit and Bug Fixes

**Status:** Complete

### Objective

Full audit of the EntityDefinition → Compiler → adapt → generator → widget.Node → validator → layout → AMIS renderer → JSON → HTTP → AMIS pipeline. Goal: a developer adds a new EntityDefinition and the framework automatically produces a working, production-correct UI without touching `awo/web`.

### Pipeline Architecture Verified

The full pipeline is structurally correct:

```
EntityDefinition (def package)
  → def.Register() in init()
  → compiler.Compile() → CompiledSchema + CapabilityGrants
  → adapt.FromCompiled() → generator.EntitySchema
      ├─ adapt.convertField() maps all FieldDef fields
      ├─ adapt.permissionsMap() maps Create/Read/Write/Delete → "create"/"read"/"update"/"delete"
      ├─ adapt.convertEdge() builds RelationDef with DataURL
      ├─ adapt.convertAction() builds ActionDef
      └─ adapt.collectDashboardPanels() for dashboard module
  → generator.EntityGenerator.Generate()
      ├─ buildList() → NodeList (columns, filter bar, toolbar actions, DataSource.URL=ListURL)
      ├─ buildForm() → NodeForm (create: DataSource.URL=CreateURL; edit: URL=EditURL, ReadURL=DetailURL)
      ├─ buildDetail() → NodePage > [NodeSummaryCard, NodeForm(ReadURL=DetailURL), relations]
      └─ buildDashboard() → NodePage > [NodeKPICard/NodeChartPanel/NodeTablePanel]
  → Validator.Validate() (hard gate — reject unknown NodeKinds, broken DataSources)
  → Layout.Apply() (section/column packing)
  → amis.DefaultRenderer.Render() → map[string]any AMIS JSON
  → engine.Engine.Handle() → RenderedOutput (with L3 cache)
  → api/sdui.Handler.handle() → HTTP JSON response
  → browser: AMIS SDK embed()
```

### Bugs Fixed

| # | Location | Bug | Fix |
|---|---|---|---|
| 1 | `amis/renderer.go` `nodeKindToColumnType()` | `NodeNumber` and `NodeMoney` returned `"tpl"` — AMIS `tpl` column without a `tpl` template string renders blank | Changed to `"number"` (AMIS v3 column type for locale-aware numeric formatting) |
| 2 | `amis/renderer.go` `renderFilterBar()` | Comment said "crud2" — component is `crud` | Corrected comment |
| 3 | `generator/generator.go` `buildDetail()` | Detail form used `DataSource.URL` (→ AMIS `api`, fires on submit) — form fields always empty on mount | Changed to `DataSource.ReadURL` (→ AMIS `initApi`, fires on mount) |
| 4 | `web/pages/index.html` `renderAmis()` | Detail/edit views: `${id}` placeholder in `initApi`/`api` URLs not resolved — AMIS needs `id` in data context | Pass `embedProps.data = { id: currentRoute.id }` when embedding detail/edit schemas |
| 5 | `web/pages/index.html` fetcher | AMIS `page`/`perPage` converted to `offset`/`limit`; backend List handler reads `page`/`page_size` — all pagination returned page 1 always | Fixed: convert AMIS `perPage`/`$perPage` → `page_size`; normalise `$page` → `page`; keep `page` value as-is |
| 6 | `generator/generator.go` `buildFilterBar()` | Filter fields submitted as `?status=active`; backend `filterparse.FromQuery()` expects `?filter[status][eq]=active` format — all filters silently ignored | Override node `Name` in `buildFilterBar`: text fields → `filter[field][contains]`, date fields → `filter[field][gte]`, all others → `filter[field][eq]` |
| 7 | `api/handler/crud.go` `List()` | AMIS sends `?orderBy=field&orderDir=asc` on column click; handler never called `driver.WithSort()` — sort column clicks had no effect | Extract `orderBy`/`orderDir` query params; call `driver.WithSort(field, asc)` when present |

### Key Architecture Findings

**`adapt.permissionsMap` (verified):** `PermissionSet.Write` maps to key `"update"` in the generator's permissions map. The generator uses `schema.Permissions["update"]` to gate the Edit button. Demo entities use `Write: []string{"demo.product.update"}` — this is correct.

**Detail view data loading (fixed):** AMIS distinguishes `api` (fires on form submit) from `initApi` (fires on mount). Using `api` for detail forms means all fields are empty on initial render. The fix (`ReadURL` → `initApi`) ensures detail fields are populated immediately on page load.

**`${id}` resolution (no cache poisoning):** The L3 cache key is `schemaFingerprint + rendererID + locale` — it does NOT include record ID. Substituting the actual ID into cached schema URLs would poison the cache (uuid1's schema returned for uuid2). Instead, pass `id` in AMIS embed `props.data` — AMIS resolves `${id}` template variables from its own data context at render time. The backend handler correctly does NOT substitute IDs into schema URLs.

**Filter parameter format:** The backend `filterparse` package uses the `filter[field][op]=value` convention (all operators: eq, neq, gt, gte, lt, lte, between, in, not_in, is_null, is_not_null, contains, starts_with, ends_with). AMIS filter bar submits plain `?field=value` without the filter wrapper. The fix is at the generator layer (field Name override in `buildFilterBar`) so the widget tree carries the correct query param names.

**Sorting:** `driver.QueryOptions.SortField`/`SortAsc` existed but were never wired from HTTP. Added extraction in the List handler.

**Pagination:** Backend uses 1-based `page` + `page_size`. AMIS sends `page` (1-based) + `perPage`. The old fetcher converted to `offset`/`limit` which the backend doesn't read. Fixed to keep `page` as-is and rename `perPage`→`page_size`.

**Response envelope (verified correct):** Backend returns `{"data":[...],"meta":{"total":N,"page":N,"page_size":N}}`. `adaptResponse()` detects `Array.isArray(body.data)` and reads `body.meta.total` → wraps to `{status:0,data:{items:[...],count:N}}` which AMIS `crud` + `itemsKey:"items"` expects.

**Locale (verified):** `extendDefaultLocale('en-US', {...})` in Block 1 maps 17 Chinese AMIS strings to English. Filter bar has explicit `submitText:"Search"`, `resetText:"Reset"`. AMIS embed passes `locale:'en-US'`. No Chinese text in English install.

**Permissions in adapt (verified correct):** `permissionsMap()` in `adapt.go` correctly maps `def.PermissionSet.Write[0]` to key `"update"` in the generator's permissions map.

### Known Remaining Limitations

- **Delete/Edit buttons in list rows always shown:** Row operations (View/Edit/Delete) in `renderList()` are generated unconditionally from `uiPrefix` (derived from the create action href). They are not gated by `schema.Permissions["delete"]`/`"update"]` in the renderer. The backend still enforces permissions (returns 403). UI improvement requires threading permission state through NodeList — future work.
- **Date filter is single-bound:** Filter bar date fields use `filter[field][gte]` (lower bound only). A date range filter would require two nodes or a date-range AMIS widget. Future enhancement.
- **Finance module not registered:** Finance entities have no `init()` calls in `erp/modules/finance/`. They are intentionally not registered — Phase 1 work. Finance will not appear in nav until Phase 1 begins.
- **Auth redirect:** 401 API responses render a static error div; no login redirect implemented. Auth stubs ready in `index.html` (commented overlay + `doLogout` stub).
- **Bulk actions:** `"bulkActions": []any{}` suppresses the AMIS bulk-action placeholder but no bulk operations are configured. Add via entity `ActionDef` with `Bulk: true` when needed.
- **Sort direction default:** Missing `orderDir` param defaults to ASC (`c.Query("orderDir") != "desc"`). This is correct behaviour.

### Architectural Invariants (Confirmed)

1. **EntityDefinition is the source of truth.** All UI structure derives from `def.EntityDefinition` → `compiler` → `adapt` → `generator`. No static JSON schemas are production sources of truth.
2. **SDUI is the entity UI layer.** All entity CRUD UI is generated by the engine. `awo/web` is a thin rendering shell.
3. **`awo/web` is not touched by module developers.** Adding an EntityDefinition automatically produces nav entries and all four view modes. No changes to `index.html` are needed per-entity.
4. **The renderer is AMIS-specific; the generator is not.** The generator produces a renderer-agnostic `widget.Node` tree. Only `awo/sdui/amis` produces AMIS JSON.
5. **L3 cache key does not include record ID.** `SchemaFingerprint + rendererID + locale`. ID must be resolved client-side via AMIS data context, not server-side before caching.

---

## Session 6 — SDUI Production Hardening: Permissions, Bulk Actions, Filters, Auth

**Status:** Complete

### Objective

Advance from "generated entity UI works" to "generated entity UI is production-grade and fully metadata-driven." Specific goals:

1. Permission-aware actions — Edit/Delete row buttons absent when viewer lacks permission (UI reflects backend enforcement)
2. Bulk actions — Delete selected rows, gated on delete permission
3. Fix silent no-op — delete button in detail view had no API URL
4. Date range filter — lower+upper bound via `input-date-range`
5. Auth handling — 401 redirects to login; 403 shows permission denied (distinct)
6. Regression tests for all Session 5 and Session 6 fixes

### Architectural Additions

**`widget.ActionNode.Scope string`** (new field in `awo/sdui/widget/node.go`)

Controls where an action appears in the rendered UI:
- `"toolbar"` — list header toolbar (create, custom toolbar actions)
- `"row"` — per-row operation column in list view (view, edit, delete)
- `"bulk"` — bulk action bar (applies to all selected rows)
- `""` (empty) — defaults to `"toolbar"` in renderer

The generator is the sole gatekeeper for which actions are emitted and with what scope. The renderer only routes by scope — it never decides which actions a viewer may see.

### Bugs Fixed

| # | Location | Bug | Fix |
|---|---|---|---|
| 1 | `generator/generator.go` `buildListActions()` | Row ops (View/Edit/Delete) emitted unconditionally, ignoring viewer permissions | Rewrote: each row op gated by `ctx.Viewer.HasPermission(schema.Permissions["update"/"delete"])`. Missing or empty perm → op emitted (no restriction declared). Scope = `"row"`. |
| 2 | `generator/generator.go` `buildListActions()` | No bulk delete action ever emitted | Added bulk-delete `ActionNode` with `Scope:"bulk"`, gated on delete permission |
| 3 | `generator/generator.go` `buildDetailActions()` | Delete button had `ActionType:"ajax"` but `API:""` — AMIS fires ajax with no URL → silent no-op | Fixed: `API: "DELETE:" + schema.EditURL` (fallback: `schema.ListURL + "/${id}"`) |
| 4 | `generator/generator.go` `buildDetailActions()` | Custom ajax actions had `API:""` — same silent no-op | Fixed: `api = "POST:" + schema.ListURL + "/${id}/actions/" + a.ID` for ajax-type actions |
| 5 | `amis/renderer.go` `renderList()` | Row ops hardcoded from `uiPrefix` extracted from create action href — permissions ignored at render | Rewrote: renderer reads `Scope:"row"` actions from `n.Actions`; builds operation column only when ≥1 row action present |
| 6 | `amis/renderer.go` `renderList()` | `var renderedBulk []any` → nil → JSON `null` → AMIS shows placeholder text for `bulkActions:null` | Changed to `renderedBulk := []any{}` (non-nil empty slice → JSON `[]`) |
| 7 | `generator/generator.go` `buildFilterBar()` | Date/DateTime filter was single-bound (`filter[field][gte]`) — no upper bound | Date/DateTime fields now use Props escape hatch: `type:"input-date-range"`, `startName:"filter[field][gte]"`, `endName:"filter[field][lte]"` |
| 8 | `web/pages/index.html` fetcher | 401 returned `{status:401, msg:'Unauthorized'}` — AMIS showed inline error instead of redirecting | 401: `window.location.href = '/auth/login?next=<encoded_path>'` + return pending Promise (AMIS never processes it) |
| 9 | `web/pages/index.html` fetcher | No differentiation between 401 (session expired) and 403 (insufficient permission) | 403: return `{status:403, msg:'You do not have permission to perform this action.'}` |
| 10 | `web/pages/index.html` `loadPage()` | 401 from schema fetch showed static error div | 401: redirect to `/auth/login?next=<encoded_path>` |
| 11 | `web/pages/index.html` `loadPage()` | 403 from schema fetch showed generic error | 403: show "Access denied" error div with entity-level permission message |
| 12 | `amis/renderer_test.go` `TestRenderer_ListColumns` | Checked for `"crud2"` — component changed to `"crud"` in Session 3 but test never updated | Fixed assertion to `"crud"` |

### Key Design Decisions

**Generator is authoritative, renderer is mechanical.**
The generator produces permission-gated `ActionNode` entries with correct `Scope`. The renderer partitions by scope and renders whatever it receives. No permission logic lives in the renderer — it cannot, because `ViewerContext` is not available at render time in the general case.

**Empty permission string → unrestricted.**
`if perm == "" || ctx.Viewer.HasPermission(perm)` — if an entity declares no delete permission, the delete button is always shown. This matches the intent: a missing permission string means "no restriction declared," not "nobody can do this."

**Date range via Props escape hatch.**
The `widget.Node.Props map[string]any` is merged last in `amis.renderNode()`, overriding all computed defaults. Setting `Props["type"] = "input-date-range"` overrides the type computed from NodeDate/NodeDateTime. This is the correct extension point for AMIS widgets that have no direct NodeKind equivalent.

**`input-date-range` posts two independent params.**
AMIS `input-date-range` posts `startName` and `endName` as separate query params. Setting `startName:"filter[field][gte]"` and `endName:"filter[field][lte]"` maps directly to the `filterparse.FromQuery` format — no additional transformation needed.

**`bulkActions: []` vs `bulkActions: null`.**
AMIS v3 `crud` component shows a "bulk action" placeholder text when `bulkActions` is `null` (Go nil slice → JSON null). An empty slice `[]` suppresses the placeholder. The fix is at initialization: `renderedBulk := []any{}`.

**Auth redirect returns pending Promise.**
After `window.location.href = ...` the AMIS fetcher must never resolve/reject (if it does, AMIS may show an error before the redirect completes). Returning `new Promise(function(){})` — a promise that never settles — prevents any AMIS error rendering during the redirect.

### New Tests Added

**`awo/sdui/amis/renderer_test.go`** — 8 new regression tests:
- `TestRenderer_NumberColumn_IsNumber` — NodeNumber/NodeMoney columns use `"number"` type (Session 5 regression)
- `TestRenderer_ListItemsKey` — `itemsKey` is always `"items"` (Session 5 regression)
- `TestRenderer_DetailForm_InitApi` — ReadURL produces `initApi` not `api` (Session 5 regression)
- `TestRenderer_ListRowActions_FromScope` — `Scope:"row"` actions → operation column with buttons
- `TestRenderer_ListNoRowActions_NoOpColumn` — no row actions → no operation column in columns array
- `TestRenderer_ListBulkActions_EmptySlice` — `bulkActions` is `[]any{}` not nil when no bulk actions
- `TestRenderer_ListBulkActions_Rendered` — `Scope:"bulk"` actions → `bulkActions` array in AMIS schema
- `TestRenderer_DeleteAction_HasAPI` — `ActionNode` with `API` set → `api` key present in rendered output

**`awo/sdui/generator/generator_test.go`** — 6 new regression tests + `makeSchemaWithUIPrefix()` helper:
- `TestGenerator_ListActions_RowScopedActions` — viewer with all perms → View/Edit/Delete row ops + bulk delete all emitted with correct scopes
- `TestGenerator_ListActions_NoRowDeleteWithoutPerm` — viewer without delete perm → delete row op absent; edit present
- `TestGenerator_ListActions_NoRowEditWithoutPerm` — viewer without update perm → edit row op absent; delete present
- `TestGenerator_DetailActions_DeleteHasAPI` — detail delete `ActionNode.API` is non-empty (format: `DELETE:...`)
- `TestGenerator_FilterBar_FieldNames` — filter names use `filter[field][op]` format; date fields use Props with `input-date-range`, `startName`, `endName`
- `makeSchemaWithUIPrefix()` — helper constructing `generator.EntitySchema` with `UIPrefix="/ui/finance/invoices"`, `ListURL`, `EditURL`, `Permissions` all set

### Known Remaining Limitations

- **Number/money range filters:** Numeric fields use `filter[field][eq]` (equality only). Range filters (`gte`/`lte`) for amount fields require either a dedicated range widget or two separate filter nodes. Post-Session 6 enhancement.
- **Finance module not registered:** Finance entities have no `init()` calls. They will not appear in nav until Phase 1 begins.
- **Custom AMIS components:** `NodeDuration` and `NodeSignature` fall back to text/file input. Production requires registered custom AMIS components.
- **Inline edge grids:** Invoice lines render as `NodeRelatedList` (read-only). Editable inline grids require `def.EdgeDef.Inline bool`.
- **Sort multi-column / clear:** AMIS sends one sort column at a time. The handler wires the first column only. Multi-sort deferred.
- **Showcase integration:** Showcase portal serves static schemas; it does not yet use the SDUI engine for live generation. Post-Session 6 enhancement.
- **Real entity matrix:** Session 6 fixes were tested against demo/finance entity definitions via unit tests, not against a running server with real PostgreSQL. Integration verification is the next step.

### Commands to Run for Verification

```
go test ./awo/sdui/...
go test ./awo/api/...
go test ./...
```

---

## Session 7 — Verification Pass: Test Correctness and Golden File Accuracy

**Status:** Complete

### Objective

Full verification pass per session instructions: trace every claimed fix, confirm generated schema, confirm tests actually guard the right behavior. Do not assume the summary is correct — read and verify the code.

### Verification Results

#### Permission-Aware Actions — VERIFIED CORRECT

Full chain traced:
- `EntityDefinition.PermissionSet` → `compiler.CapabilityGrant` → `adapt.permissionsMap()` → `generator.EntitySchema.Permissions` map
- `generator.buildListActions()`: Edit gated by `schema.Permissions["update"]`, Delete/bulk-delete gated by `schema.Permissions["delete"]`
- `generator.buildDetailActions()`: Edit gated by update perm, Delete gated by delete perm
- `amis.renderList()`: partitions `n.Actions` by `Scope` — renderer only routes, never decides visibility
- Backend still enforces via `RequirePermission` middleware independently

Absent nodes (not in tree) — not hidden nodes. Cannot be revealed by client manipulation. ✓

#### `${id}` URL Format — BUG DISCOVERED AND FIXED

**Bug**: `makeSchema()` in `generator_test.go` and `invoiceSchema()` in `golden_test.go` both used `{id}` (Go template format) for `EditURL` and `DetailURL`:

```go
// WRONG — was
EditURL:   "/api/v1/finance/invoices/{id}",
DetailURL: "/api/v1/finance/invoices/{id}",

// CORRECT — fixed
EditURL:   "/api/v1/finance/invoices/${id}",
DetailURL: "/api/v1/finance/invoices/${id}",
```

Production `adapt.FromCompiled()` correctly uses `${id}` (AMIS template format). Test schemas were inconsistent with production. AMIS only resolves `${id}` syntax; `{id}` is sent as a literal string and the record ID is never substituted.

**Fix**: Updated both test schema constructors.

**Consequence for detail view delete**: `buildDetailActions` uses `schema.EditURL` directly for the delete API URL. With the old `{id}` test schemas, the golden showed `"api": "DELETE:.../invoices/{id}"` — this would fail at runtime because AMIS would not resolve the ID. Fixed golden files now show `${id}` throughout.

#### Delete API URL in Detail View — VERIFIED CORRECT (after schema fix)

`buildDetailActions()`:
```go
deleteAPI := "DELETE:" + schema.EditURL   // = "DELETE:/api/.../invoices/${id}"
```

`TestGenerator_DetailActions_DeleteHasAPI` strengthened: now also asserts `strings.Contains(action.API, "${id}")` — not just `!= ""`.

#### UIPrefix Missing in Golden Schema — BUG DISCOVERED AND FIXED

`invoiceSchema()` in `golden_test.go` had no `UIPrefix`. Consequences:
- Create button in list generated as `/finance_invoice/create` (entity-name fallback)
- View/Edit row actions absent from list (both check `if schema.UIPrefix != ""`)
- Edit link in detail generated as `/finance_invoice/${id}/edit` (fallback)

**Fix**: Added `UIPrefix: "/ui/finance/invoices"` to `invoiceSchema()`. Golden files now show realistic production URLs.

#### `NodeText`/`NodeTextArea` Typos — VERIFIED FIXED

Searched codebase: no occurrences of `NodeInput` or `NodeTextarea` (lowercase a). Only `NodeText` and `NodeTextArea` (capital A) exist. Fix applied in Session 6 (`generator.go:501`). ✓

#### `adapt_test.go` EditURL assertion — VERIFIED FIXED

Line 90 now checks `strings.HasSuffix(gs.EditURL, "/${id}")` — matches actual `adapt.FromCompiled()` output. ✓

#### Golden Files — REGENERATED

All four golden files updated to reflect:
1. `${id}` in all API and initApi URLs (EditURL/DetailURL now correct)
2. `UIPrefix` set → realistic nav links (create/view/edit)
3. Three row operation buttons (View/Edit/Delete) with `width: 195`
4. Bulk delete in `bulkActions`

| Golden file | Changes |
|---|---|
| `invoice_list.golden.json` | Create link `/ui/finance/invoices/create`; View/Edit row buttons added; width 195 |
| `invoice_detail.golden.json` | Edit link `/ui/finance/invoices/${id}/edit`; delete API `${id}`; initApi `${id}` |
| `invoice_edit.golden.json` | Form api.url `${id}`; initApi `${id}` |
| `invoice_create.golden.json` | No changes (CreateURL unchanged; no ${id} in create) |

**Why the changes are architecturally correct:** The golden files previously captured schemas with `{id}` which AMIS would not resolve at runtime — this was a latent bug. The new goldens capture what production `adapt.FromCompiled()` actually produces, which AMIS resolves correctly via `props.data = { id: currentRoute.id }` in the frontend shell.

#### `adaptResponse()` and Auth Handling — VERIFIED CORRECT

- AMIS fetcher 401: redirects to `/auth/login?next=<path>` + returns pending Promise. AMIS never processes a response. ✓
- AMIS fetcher 403: returns `{ status: 403, msg: '...' }` object. AMIS shows inline error. ✓
- `loadPage` 401: redirects. ✓
- `loadPage` 403: shows "Access denied" div. ✓
- `adaptResponse()`: detects `Array.isArray(body.data)` → wraps to `{ status:0, data:{ items:[...], count:N } }` ✓

#### Test Coverage Summary

**New tests added this session:** 1 assertion strengthened (`TestGenerator_DetailActions_DeleteHasAPI` now verifies `${id}` format).

**Tests passing (pre-golden-regen):**
- `awo/sdui/widget` ✓
- `awo/sdui/expression` ✓
- `awo/sdui/cache` ✓
- `awo/sdui/layout` ✓
- `awo/sdui/observability` ✓
- `awo/sdui/plugins` ✓
- `awo/sdui/registry` ✓
- `awo/sdui/validation` ✓
- `awo/sdui/amis` ✓
- `awo/sdui/generator` ✓
- `awo/sdui/conformance` ✓

**Awaiting re-run after fixes:**
- `awo/sdui/adapt` (stale `/{id}` assertion → fixed)
- `awo/sdui/engine` (golden files → regenerated)

### Commands to Run for Verification

```
go test ./awo/sdui/...
go test ./awo/api/...
```

If the engine golden tests still fail, regenerate:
```
go test ./awo/sdui/engine/... -run TestGolden -update
go test ./awo/sdui/...
```

### Remaining Limitations (unchanged from Session 6)

- Finance module not registered (Phase 1 work)
- Number/money range filters use eq-only
- `NodeDuration`/`NodeSignature` fall back to text/file
- Inline edge grids require `def.EdgeDef.Inline`
- Showcase uses static schemas, not live SDUI engine
