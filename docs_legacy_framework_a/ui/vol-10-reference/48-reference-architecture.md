> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "48 – Reference Architecture"
volume: "vol-10-reference"
chapter: 48
section: "Reference"
status: "implemented"
---

# Chapter 48 – Reference Architecture

## Table of Contents
- [48.1 System Diagram](#481-system-diagram)
- [48.2 Data Flow: Request Through Nine Stages](#482-data-flow-request-through-nine-stages)
- [48.3 Key Package Dependencies](#483-key-package-dependencies)
- [48.4 Decision Tree: PageFn vs ASTPageFn vs DSL Screen](#484-decision-tree-pagefn-vs-astpagefn-vs-dsl-screen)

---

## 48.1 System Diagram

```
  Browser (AMIS SDK)
       │
       │  GET /ui/schema?route=/finance/dashboard
       ▼
  ┌──────────────────────────────────────────────────┐
  │  SchemaHandler  (internal/api/handlers/)         │
  │  - Extract route, session token                  │
  │  - Invoke Pipeline                               │
  │  - Serialise schema to JSON                      │
  └───────────────────┬──────────────────────────────┘
                      │
                      ▼
  ┌──────────────────────────────────────────────────┐
  │  UIPipeline  (internal/web/pipeline/)            │
  │                                                  │
  │  Stage 1: AuthStage                              │
  │    └─ Validate session → populate UISessionCtx  │
  │                                                  │
  │  Stage 2: RegistryStage                          │
  │    └─ Lookup PageRegistration by route           │
  │                                                  │
  │  Stage 3: CacheStage (read)                      │
  │    └─ Check cache by (route+permFP+versions+flagFP)│
  │       Hit? → skip to SerialiseStage              │
  │                                                  │
  │  Stage 4: CompileStage                           │
  │    └─ Call ASTPageFn (or PageFn fallback)        │
  │    └─ CompileTree → collect all errors           │
  │                                                  │
  │  Stage 5: ValidateStage                          │
  │    └─ Structural rules (chart style, dialog, …) │
  │                                                  │
  │  Stage 6: CacheStage (write)                     │
  │    └─ Store compiled schema                      │
  │                                                  │
  │  Stage 7: SerialiseStage                         │
  │    └─ Marshal schema to JSON bytes               │
  │                                                  │
  │  (Each stage wrapped in InstrumentedStage)       │
  └──────────────────────────────────────────────────┘
                      │
            JSON AMIS schema
                      │
       ┌──────────────▼─────────────────┐
       │                                │
  OTel Traces              Prometheus Metrics
  (Jaeger / Tempo)         (6 UI metrics)
```

```
  ┌─────────────────────────────────────────────────┐
  │  DSL Layer  (internal/web/dsl/)                  │
  │                                                   │
  │  registry/        — PageRegistration store        │
  │  screens/         — Route → composed schema       │
  │    finance_dashboard.go                           │
  │    finance_invoice_list.go                        │
  │    iam_users.go                                   │
  │    ...                                            │
  │  blocks/          — Shared composable components  │
  │    finance_quick_actions.go                       │
  │    invoice_header.go                              │
  │    shared_stat_row.go                             │
  │    ...                                            │
  │  ast/             — Typed node structs            │
  │    PageNode, CRUDNode, PanelNode,                 │
  │    TabsNode, FormNode, StatNode, ChartNode, …     │
  └─────────────────────────────────────────────────┘
```

```
  ┌──────────────────────────────────────┐
  │  Web Shell  (web/)                    │
  │                                       │
  │  web/pages/index.html — sidebar + embed│
  │  web/sdk/             — AMIS SDK       │
  │  web/styles/css/      — theme tokens   │
  └──────────────────────────────────────┘
```

---

## 48.2 Data Flow: Request Through Nine Stages

A schema request travels through the following transformations:

```
HTTP Request
  │
  │  route="/finance/dashboard", Authorization: Bearer <token>
  ▼
AuthStage
  │  Validates token via IAM service
  │  Resolves permissions → []string
  │  Resolves feature flags → map[string]bool
  │  Builds UISessionContext{TenantID, UserID, Permissions, Flags, Locale, Currency}
  ▼
RegistryStage
  │  Looks up PageRegistration by route
  │  Returns: PageRegistration{ASTFn: FinanceDashboardScreen}
  │  Error if not found → HTTP 404
  ▼
CacheStage (read)
  │  Computes cacheKey = route + permFingerprint + versions + flagFingerprint
  │  Cache hit?  ──Yes──▶  SerialiseStage (skip compile+validate)
  │  Cache miss? ──No───▶  CompileStage
  ▼
CompileStage
  │  Calls FinanceDashboardScreen(UISessionContext)
  │  Returns ast.PageNode{...}
  │  Calls ast.CompileTree(node) → []error
  │  Error if CompileTree returns errors → HTTP 500
  ▼
ValidateStage
  │  Runs structural validation rules against the compiled node tree
  │  Checks: chart has Style, dialog actions have Dialog, CRUDNode columns non-empty, …
  │  Error if any rule fails → HTTP 500 with error code
  ▼
CacheStage (write)
  │  Stores compiled schema under cacheKey
  │  Increments ui_invalidation_events_total if generation mismatched
  ▼
SerialiseStage
  │  Marshals ast.Node tree to JSON bytes
  ▼
HTTP Response
  │  200 OK, Content-Type: application/json
  │  Body: { "type": "page", "title": "Finance Dashboard", ... }
  ▼
Browser AMIS SDK
  │  Receives schema JSON, renders component tree
  │  Executes InitAPI call to fetch data
  │  Renders data-bound UI
```

---

## 48.3 Key Package Dependencies

```
internal/api/handlers/
  └── depends on: internal/web/pipeline/

internal/web/pipeline/
  └── depends on: internal/web/dsl/registry/
  └── depends on: internal/web/dsl/ast/
  └── depends on: internal/web/pipeline/stages/
  └── depends on: internal/web/ui/  (UISessionContext)

internal/web/dsl/screens/
  └── depends on: internal/web/dsl/ast/
  └── depends on: internal/web/dsl/blocks/
  └── depends on: internal/web/dsl/registry/
  └── depends on: internal/web/ui/  (UISessionContext)

internal/web/dsl/blocks/
  └── depends on: internal/web/dsl/ast/
  └── depends on: internal/web/ui/  (UISessionContext)
  └── MUST NOT depend on: internal/core/iam/ (use UISessionContext instead)
  └── MUST NOT depend on: internal/web/pipeline/ (no stage imports in blocks)

internal/web/dsl/ast/
  └── depends on: nothing in internal/ (pure value types)

internal/web/ui/
  └── depends on: nothing in internal/ (pure value types)
```

The dependency graph is strictly layered.  No dependency goes upward (e.g.
`ast` must not import `pipeline`; `blocks` must not import `screens`).

---

## 48.4 Decision Tree: PageFn vs ASTPageFn vs DSL Screen

Use this tree when deciding how to implement a new UI surface:

```
Is this a new page accessible via a URL route?
├─ No → It is a block. Add to internal/web/dsl/blocks/
│         Use ASTPageFn-style: func ...Block(sess UISessionContext) ast.Node
│
└─ Yes → It is a screen. Add to internal/web/dsl/screens/
          │
          Is this a migration of an existing PageFn screen?
          ├─ Yes → Use CompileStage dual dispatch (§44.4)
          │         Migrate gradually: change Fn to ASTFn when ready
          │
          └─ No (new screen) → Always use ASTPageFn
                               Register with RegisterPage(PageRegistration{ASTFn: ...})
                               │
                               Does the screen compose only existing blocks?
                               ├─ Yes → Screen file should be < 60 lines
                               │
                               └─ No → Extract the new pattern as a block first,
                                       then compose it in the screen
```

**When to use PageFn (legacy):**
- Only when migrating an existing screen that uses `amis.*` builders
- Never for new screens
- Remove as soon as the screen is fully migrated to `ASTPageFn`

**When to use ASTPageFn (current):**
- All new screens
- All screens after migration from PageFn
- The function body should delegate to blocks; it should not contain business logic

**When to use a block (not a screen):**
- Any reusable UI component
- Any UI element with its own permission gate
- Any UI element with its own data fetch (`InitAPI`)
- Any UI element that appears on more than one screen
