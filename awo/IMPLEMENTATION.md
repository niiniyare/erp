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

### Not yet implemented:
- `awo/sdui/layout` — LayoutEngine (column span computation, SectionLayout)
- `awo/sdui/observability` — metrics, tracing, logging

### Dependency: go.mod
- `golang.org/x/sync` (for singleflight in cache package)
- `github.com/google/uuid` (already present)

### Known limitations:
- Generator not yet implemented — SDUI pipeline cannot yet run end-to-end.
- `NodeDuration` renders as masked `input-text` in AMIS (no native widget); production requires a custom AMIS component.
- `NodeSignature` renders as `input-file` in AMIS (no native signature widget); production requires a custom AMIS component.
- `NodeMoney` currency selector requires a sibling `NodeSelect` emitted by the generator — not yet handled in generator.

### Pre-implementation checklist status (from Part 20):
- [x] ExpressionRef as portable DSL (not renderer strings)
- [x] RenderedOutput as typed union
- [x] GeneratorContext immutable value type
- [x] Registry seals at bootstrap; NewIsolated() for tests
- [x] Plugin priorities deterministic; duplicate = bootstrap error
- [x] NodeTabPane separate from NodeSection
- [x] Cache keys include all 9 dimensions
- [x] Validation before render (hard gate)
- [ ] Generator implementation (next priority)
- [ ] Layout engine implementation
