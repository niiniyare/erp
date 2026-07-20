> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Chapter 03 — Architectural Principles

> **Volume:** I — Vision & Philosophy
> **Audience:** Platform Engineers, Solution Architects, Senior Backend Engineers
> **Prerequisites:** Chapter 02 — Design Philosophy

---

## Table of Contents

- [3.1 Separation of Concerns: Business vs. Presentation](#31-separation-of-concerns-business-vs-presentation)
- [3.2 The AST-First Principle](#32-the-ast-first-principle)
- [3.3 Authorization at the Source](#33-authorization-at-the-source)
- [3.4 Cache After Authz, Never Before](#34-cache-after-authz-never-before)
- [3.5 Pure Page Functions](#35-pure-page-functions)
- [3.6 Fail-Safe Pipeline](#36-fail-safe-pipeline)
- [3.7 Composability: Blocks, Screens, Builders](#37-composability-blocks-screens-builders)
- [3.8 Immutability of Compiled Schemas](#38-immutability-of-compiled-schemas)
- [3.9 AMIS Compliance as a Platform Invariant](#39-amis-compliance-as-a-platform-invariant)
- [3.10 Layered Architecture Summary](#310-layered-architecture-summary)

---

## 3.1 Separation of Concerns: Business vs. Presentation

**Business concerns** require domain knowledge:
- Does this user have write access to this GL code?
- Is this invoice above the approval threshold?
- Which workflow stages are available given the current document state?

**Presentation concerns** require platform knowledge:
- How wide should this column be on a narrow viewport?
- Which AMIS component renders a date range picker?
- How many decimal places does the currency format require?

Business concerns are answered in Go, in `UISessionContext.Can()`, in `UISessionContext.Flag()`, in the page function body. They are encoded into the schema as semantic data: which actions are present, which fields are read-only, which tabs are included. The AMIS renderer reads semantic data and presents it using its own conventions.

**The separation fails when:**
- A page function sets `"color": "#FF0000"` to communicate "this field has an error." That is a presentation decision. Correct expression: `"status": "error"` — AMIS decides what error state looks like.
- A page function conditionally omits a field based on CSS class availability on the client. That is a presentation concern that belongs in the browser.

---

## 3.2 The AST-First Principle

Every new page is expressed as a typed Go struct graph before it becomes JSON. `ast.Node` implementations are value types, immutable after construction:

```go
// internal/web/ast/node.go
type Node interface {
    NodeType() string      // AMIS "type" field, e.g. "page", "crud", "form"
    Validate() error       // structural invariants
    Compile() map[string]any
}

type ContainerNode interface {
    Node
    Children() []Node      // recurse for validation and compilation
}
```

`ast.CompileTree(root)` collects all validation errors from the entire tree before emitting any JSON. This means a broken child node does not produce a partially valid schema that reaches the browser — it produces an error that surfaces at the first request, not at deployment time.

Implemented node types: `PageNode`, `GridNode`, `FlexNode`, `TabsNode`, `SplitPaneNode`, `SectionNode`, `TableNode`, `CRUDNode`, `ChartNode`, `CardNode`, `StatNode`, `TimelineNode`, `TreeNode`, `MappingNode`, `PropertyNode`, `FormulaNode`, `FormNode`, `FilterBarNode`, `InputTextNode`, `InputNumberNode`, `InputDateNode`, `InputDateRangeNode`, `SelectNode`, `ActionNode`, `DialogNode`, `DrawerNode`.

---

## 3.3 Authorization at the Source

Casbin permissions are resolved once per request in AuthzStage (priority 20), via `UIAuthzService.BulkEnforce`. The result is a `map[string]bool` sealed inside `UISessionContext`. Page functions receive this context and call `sess.Can(action, resource)` — no Casbin round-trip inside a page function.

```go
// What this enables in a page function:
func InvoiceScreen(sess ui.UISessionContext) ast.Node {
    actions := []ast.Node{}
    if sess.Can("approve", "finance.invoices") {
        actions = append(actions, ast.ActionNode{
            Label:      "Approve",
            ActionType: "ajax",
            Level:      "success",
        })
    }
    return ast.PageNode{Body: actions}
}
```

The schema delivered to the browser never includes the `Approve` button for a user who lacks the permission. The browser has no authorization logic to bypass.

---

## 3.4 Cache After Authz, Never Before

The cache key is: `route + tenantID + permFingerprint + flagFingerprint`.

`permFingerprint` is a hash of the user's resolved `map[string]bool`. `flagFingerprint` is a hash of the user's active feature flags. Both are computed by AuthzStage.

**Why CacheStage runs at priority 30 (after AuthzStage at 20):** Without the fingerprints, the cache cannot be keyed on authorization state. Caching before authorization would serve one user's schema to another user — a security failure. The pipeline design makes this ordering structural, not optional.

```
Priority 10: SessionStage    — extract IAM session from context
Priority 20: AuthzStage      — BulkEnforce → UISessionContext → perm+flag fingerprints
Priority 30: CacheStage      — lookup: route + tenant + perm_fp + flag_fp
Priority 40: RegistryStage   — Match(route) → PageFn or ASTPageFn  [skipped on hit]
Priority 50: CompileStage    — execute PageFn(sess) or CompileTree(ASTPageFn(sess))  [skipped]
Priority 60: NormalizeStage  — canonicalize AMIS types and whitespace  [skipped]
Priority 70: ValidateStage   — structural + security rules  [skipped]
Priority 80: CacheStoreStage — write compiled schema to Redis  [skipped]
Priority 90: ResponseStage   — {"status": 0, "data": schema}  [always runs]
```

---

## 3.5 Pure Page Functions

Both `PageFn` and `ASTPageFn` must be pure:

- **Deterministic:** Same `UISessionContext` → same schema. Always.
- **No I/O:** No database calls, no HTTP calls, no file reads. Data is fetched by AMIS's `initApi` / `api` fields at browser runtime.
- **No goroutines:** Executes synchronously within the pipeline stage.
- **No IAM imports:** `UISessionContext` is the only identity source. No direct access to `contract.SessionContext` or internal IAM types.

These constraints are what make caching correct. The pipeline caches the output because the output is a deterministic function of the inputs, which are captured in the cache key.

Data loading is delegated to the AMIS runtime via `InitAPI`:

```go
ast.PageNode{
    Title: "Finance Dashboard",
    InitAPI: &ast.APISpec{
        Method: "get",
        URL:    "/api/v1/finance/dashboard/summary",
    },
    Body: []ast.Node{
        // Nodes reference ${revenue}, ${ar_balance} from InitAPI response
        ast.StatNode{Label: "Revenue", Value: "${revenue}"},
    },
}
```

The page function does not fetch `revenue`. It describes where to fetch it and what to do with the result.

---

## 3.6 Fail-Safe Pipeline

Every error in the pipeline produces an HTTP response that the AMIS browser SDK can render:

- **401 Unauthenticated:** Returns an AMIS page schema with an alert and a "Log In" button. The browser renders a readable session-expired page, not a blank screen.
- **404 Page Not Found:** Returns `{"status": 404, "msg": "schema not found: /route"}`.
- **500 Internal Error:** Returns `{"status": 500, "msg": "An internal error occurred."}`.
- **Validation failure (VALIDATE_* codes):** Returns 500 — schema validation failures are programmer errors, not user errors.

The 401 response returning an AMIS schema is intentional. The AMIS SDK always expects a JSON response with a `status` field. Returning a plain HTTP 401 would cause the SDK to display a raw error, not a usable UI.

---

## 3.7 Composability: Blocks, Screens, Builders

Pages are composed from three layers:

**DSL Blocks** (`internal/web/dsl/blocks/`) — Reusable, session-aware schema fragments. Each block is a function: `func(sess UISessionContext, cfg Config) ast.Node`. Implemented blocks: `DataTableBlock`, `StatusBadgeBlock`, `DetailCardBlock`, `LineItemsBlock`, `ApprovalBlock`, `QuickActionsBlock`.

**DSL Screens** (`internal/web/dsl/screens/`) — Complete page functions for specific routes. Each screen composes blocks. Implemented screens: `FinanceDashboardScreen`, `InvoiceScreen`, `JournalEntryScreen`, `TrialBalanceScreen`, `RegisterScreen`.

**DSL Builders** (`internal/web/dsl/builders/`) — Domain-specific page builder helpers grouped by module: `finance.go`, `inventory.go`, `hr.go`, `approval.go`, `tenant.go`. Builders produce AST fragments for common ERP patterns.

A new ERP module adds its screens to `dsl/screens/`, registers them in `registry.RegisterPage()` from `init()`, and the pipeline serves them automatically.

---

## 3.8 Immutability of Compiled Schemas

The compiled schema stored in Redis is treated as immutable for the life of the cache entry. The pipeline never partially updates a cached schema. On cache invalidation (new deployment, permission change producing a new fingerprint, explicit invalidation), the next request recompiles from scratch and writes the new schema atomically.

This ensures that no browser receives a partially updated schema — the transition is always from one complete, valid schema to another.

---

## 3.9 AMIS Compliance as a Platform Invariant

NormalizeStage (priority 60) enforces AMIS compliance rules:
- All `type` values are lowercased
- API URL whitespace is trimmed

ValidateStage (priority 70) enforces:
- **Structural rules:** CRUD components have `syncLocation` set; chart components have transparent background in both `style` and `config`.
- **Security rules:** No IAM expression strings in the schema (e.g., no `"${user.role}"` binding that leaks IAM data to the browser).

Both stages are skipped on cache hits. A cached schema is already compliant.

For AST-compiled schemas (where `DataKeyASTCompiled` is set true), NormalizeStage skips structural rules that are already guaranteed by the typed node implementations. The typed nodes emit correct AMIS structure by construction.

---

## 3.10 Layered Architecture Summary

```
┌─────────────────────────────────────────────┐
│              Browser (AMIS SDK)             │  Renders schema; no authorization logic
└─────────────────┬───────────────────────────┘
                  │  GET /schema/<route>
┌─────────────────▼───────────────────────────┐
│            SchemaHandler (Fiber)            │  /schema/* — requires Authenticate middleware
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│          9-Stage UI Pipeline                │
│  10 Session → 20 Authz → 30 Cache           │
│  40 Registry → 50 Compile → 60 Normalize    │
│  70 Validate → 80 CacheStore → 90 Response  │
└──────┬──────────┬──────────────────────────┘
       │          │
┌──────▼──┐  ┌────▼─────────────────────────┐
│  Redis  │  │  Page Registry               │
│  Cache  │  │  (route → PageFn/ASTPageFn)  │
└─────────┘  └────┬─────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│  ASTPageFn / PageFn                         │
│  ┌──────────────────────────────────────┐   │
│  │  DSL Screens  →  DSL Blocks  →  AST │   │
│  └──────────────────────────────────────┘   │
└──────────────────────────────────────────────┘
```

---

*End of Chapter 03*

**Previous:** [Chapter 02 — Design Philosophy](./02-design-philosophy.md)
**Next:** [Chapter 04 — System Overview](./04-system-overview.md)
