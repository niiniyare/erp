# SDUI Framework Implementation Status

**Framework:** AWO ERP SDUI
**Architecture baseline:** Parts 1–20 (sdui_architecture.md + sdui_implementation.md)
**Last updated:** Session 1

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
