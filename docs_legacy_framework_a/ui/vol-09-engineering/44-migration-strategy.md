> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "44 – Migration Strategy"
volume: "vol-09-engineering"
chapter: 44
section: "Engineering"
status: "in progress — dual dispatch implemented; screens being migrated"
---

# Chapter 44 – Migration Strategy

## Table of Contents
- [44.1 PageFn → ASTPageFn Migration Path](#441-pagefn--astpagefn-migration-path)
- [44.2 Deprecated Registration API](#442-deprecated-registration-api)
- [44.3 amis.\* Builders → ast.\* Struct Literals](#443-amis-builders--ast-struct-literals)
- [44.4 CompileStage Dual Dispatch](#444-compilestage-dual-dispatch)
- [44.5 Migration Status by Screen](#445-migration-status-by-screen)

---

## 44.1 PageFn → ASTPageFn Migration Path

The original `PageFn` type returns `any` — an untyped AMIS-compatible map or
struct.  `ASTPageFn` returns `ast.Node` — a typed, validated node.

```go
// Legacy — returns untyped any
type PageFn func(sess UISessionContext) any

// Current — returns typed ast.Node
type ASTPageFn func(sess UISessionContext) ast.Node
```

**Migration steps for a single screen:**

1. Change the function signature from `func(UISessionContext) any` to
   `func(UISessionContext) ast.Node`.
2. Replace any `amis.Page(...)` builder call with an `ast.PageNode{...}` struct
   literal.
3. Replace inline `map[string]any{...}` nodes with typed `ast.*` struct literals.
4. Run `ast.CompileTree(node)` in the unit test to confirm no validation errors.
5. Update the `PageRegistration` to use `ASTFn` instead of `Fn`.

```go
// Before
registry.RegisterPage(registry.PageRegistration{
    Route:  "/finance/invoices",
    Module: "finance",
    Title:  "Invoices",
    Fn:     InvoiceListFn,  // PageFn
})

// After
registry.RegisterPage(registry.PageRegistration{
    Route:  "/finance/invoices",
    Module: "finance",
    Title:  "Invoices",
    ASTFn:  InvoiceListScreen,  // ASTPageFn
})
```

---

## 44.2 Deprecated Registration API

The original `Register(route string, fn PageFn)` function is deprecated.
Do not use it for new screens.  Use `RegisterPage(PageRegistration{...})`.

```go
// Deprecated — do not use for new screens
registry.Register("/finance/invoices", InvoiceListFn)

// Current
registry.RegisterPage(registry.PageRegistration{
    Route:  "/finance/invoices",
    Module: "finance",
    Title:  "Invoices",
    ASTFn:  InvoiceListScreen,
})
```

`Register` remains in the codebase for backward compatibility but will be
removed once all screens have been migrated.  `ValidateRegistry` emits a
deprecation warning (not an error) for routes registered via the old API.

---

## 44.3 amis.\* Builders → ast.\* Struct Literals

Early DSL code used `amis.*` builder functions that returned `map[string]any`.
These are being replaced with typed `ast.*` struct literals.

| Legacy (`amis.*`) | Current (`ast.*`) |
|---|---|
| `amis.Page(title, body...)` | `ast.PageNode{Title: title, Body: body}` |
| `amis.CRUD(api, columns...)` | `ast.CRUDNode{API: api, Columns: columns}` |
| `amis.Panel(title, body)` | `ast.PanelNode{Title: title, Body: body}` |
| `amis.Tabs(tabs...)` | `ast.TabsNode{Tabs: tabs}` |
| `amis.Button(label, action)` | `ast.ActionNode{Label: label, ActionType: action}` |
| `amis.Column(name, label)` | `ast.TableColumnNode{Name: name, Label: label}` |
| `amis.Form(api, controls...)` | `ast.FormNode{API: api, Controls: controls}` |
| `map[string]any{"type": "stat", ...}` | `ast.StatNode{...}` |

The key advantage of the migration is that `CompileTree` can validate `ast.*`
structs — it cannot validate `map[string]any` blobs.

---

## 44.4 CompileStage Dual Dispatch

`CompileStage` supports both `ASTPageFn` and the legacy `PageFn` via dual
dispatch.  `ASTPageFn` is the preferred path:

```go
func (s *CompileStage) Execute(ctx context.Context, pCtx *PipelineContext) error {
    reg := pCtx.Registration

    switch {
    case reg.ASTFn != nil:
        // Preferred: typed AST path
        node := reg.ASTFn(pCtx.Session)
        errs := ast.CompileTree(node)
        if len(errs) > 0 {
            return fmt.Errorf("%w: %v", ErrASTCompileFailed, errs)
        }
        pCtx.Schema = node

    case reg.Fn != nil:
        // Legacy: untyped PageFn path — no CompileTree validation
        defer func() {
            if r := recover(); r != nil {
                pCtx.PipelineError = fmt.Errorf("PageFn panic: %v", r)
            }
        }()
        pCtx.Schema = reg.Fn(pCtx.Session)

    default:
        return ErrNoPageFn
    }

    return nil
}
```

The legacy path skips `CompileTree` — `ValidateStage` still runs on the output,
but without typed node information, some validation rules cannot fire.  This is
why migrating all screens to `ASTPageFn` improves validation coverage.

---

## 44.5 Migration Status by Screen

Current inventory of DSL screens and their migration state:

| Screen route | Module | Registration | Fn type | Status |
|---|---|---|---|---|
| `/finance/dashboard` | finance | `RegisterPage` | `ASTPageFn` | Migrated |
| `/finance/invoices` | finance | `RegisterPage` | `ASTPageFn` | Migrated |
| `/finance/invoices/:id` | finance | — | — | Not yet (param routing OPEN) |
| `/finance/payments` | finance | `RegisterPage` | `ASTPageFn` | Migrated |
| `/finance/chart-of-accounts` | finance | `RegisterPage` | `ASTPageFn` | Migrated |
| `/iam/users` | iam | `RegisterPage` | `PageFn` | Pending migration |
| `/iam/roles` | iam | `RegisterPage` | `PageFn` | Pending migration |
| `/tenants` | tenants | `Register` (deprecated) | `PageFn` | Pending migration + API upgrade |
| `/settings` | settings | `Register` (deprecated) | `PageFn` | Pending migration + API upgrade |

Screens using the deprecated `Register` API or `PageFn` type are tracked as
migration debt.  New screens must always use `RegisterPage` + `ASTPageFn`.
