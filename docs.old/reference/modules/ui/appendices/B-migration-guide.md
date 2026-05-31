# Migration Guide: Legacy → Modern Page Builder

> Last verified: 2026-05-18 | Code pointers: `internal/web/ui/types.go`, `internal/web/registry/registry.go`, `internal/web/amis/`

---

## Overview

The codebase has two approaches for building page schemas. Both work. New pages must use the modern approach. Existing legacy pages migrate during normal feature work — not as a dedicated sweep.

| | Legacy | Modern |
|--|--------|--------|
| Context type | `amis.Ctx` | `ui.UISessionContext` |
| Return type | `ui.Schema` (= `map[string]any`) | `any` (returns `ast.Node`) |
| Function type | `ui.PageFn` | `ui.ASTPageFn` |
| Registration key | `Fn:` | `ASTFn:` |
| Old registration API | `registry.Register(path, fn)` *(removed)* | `registry.RegisterPage(PageRegistration{})` |
| Permission resolution | Manual `Can()` closure in `amis.Ctx` | Pre-resolved by `AuthzStage` on `UISessionContext` |
| Structural validation | Runtime `ValidateStage` checks | Compile-time in `ast.CompileTree()` |

`CompileStage` always prefers `ASTFn`. If both `Fn` and `ASTFn` are set, `ASTFn` wins.

---

## What Changed and Why

The original design used `amis.Ctx` + raw `map[string]any`. This worked but:

1. **`amis.Ctx.Can` required a closure injection** — permission resolution happened at schema build time, not before. This made it impossible to fingerprint permissions for cache keys.
2. **No structural guarantees** — a missing `syncLocation: true` on a CRUD or a missing transparent background on a chart only surfaced at `ValidateStage` (runtime).
3. **Old registration API (`registry.Register`)** — did not carry `Module`, `Title`, `Description` metadata, which broke the registry debug endpoint and startup validation.

The modern approach resolves all three: `AuthzStage` resolves permissions before the page function runs, `ast.Node` types encode structural rules, and `PageRegistration` carries full metadata.

---

## Step-by-Step Migration

### Before (legacy)

```go
// internal/web/pages/finance/invoices/schema.go
package invoices

import (
    "awo.so/internal/web/amis"
    "awo.so/internal/web/registry"
    "awo.so/internal/web/ui"
)

func init() {
    registry.RegisterPage(registry.PageRegistration{
        Route:  "/finance/invoices",
        Module: "finance",
        Title:  "Invoices",
        Fn:     Schema, // ← PageFn, legacy path
    })
}

func Schema(sess ui.UISessionContext) ui.Schema {
    return amis.M{
        "type":  "page",
        "title": "Invoices",
        "data": amis.M{
            "can_create": sess.Can("create", "invoice"),
            "can_delete": sess.Can("delete", "invoice"),
        },
        "body": amis.M{
            "type":         "crud",
            "api":          "get:/api/v1/finance/invoices",
            "syncLocation": true,
            "headerToolbar": amis.A{
                amis.M{
                    "type":      "button",
                    "label":     "New Invoice",
                    "level":     "primary",
                    "visibleOn": "${can_create}",
                },
            },
            "columns": amis.A{
                amis.M{"name": "number", "label": "Invoice #"},
                amis.M{"name": "status", "label": "Status"},
                amis.M{
                    "type":  "operation",
                    "label": "Actions",
                    "buttons": amis.A{
                        amis.M{
                            "type":        "button",
                            "label":       "Delete",
                            "level":       "danger",
                            "visibleOn":   "${can_delete}",
                            "actionType":  "ajax",
                            "api":         "delete:/api/v1/finance/invoices/${id}",
                            "confirmText": "Delete this invoice?",
                        },
                    },
                },
            },
        },
    }
}
```

### After (modern — DSL screens)

```go
// internal/web/pages/finance/invoices/schema.go
package invoices

import (
    "awo.so/internal/web/dsl/screens"
    "awo.so/internal/web/registry"
    "awo.so/internal/web/ui"
)

func init() {
    registry.RegisterPage(registry.PageRegistration{
        Route:       "/finance/invoices",
        Module:      "finance",
        Title:       "Invoices",
        Description: "Invoice list with status filter and date range.",
        ASTFn:       Schema, // ← ASTPageFn, preferred
    })
}

func Schema(sess ui.UISessionContext) any { // returns ast.Node
    return screens.CRUDScreen(screens.CRUDScreenConfig{
        Title:     "Invoices",
        API:       "get:/api/v1/finance/invoices",
        CanCreate: sess.Can("create", "invoice"),
        CanDelete: sess.Can("delete", "invoice"),
        Columns: []screens.Column{
            {Name: "number", Label: "Invoice #", Sortable: true},
            {Name: "status", Label: "Status"},
        },
        DeleteAPI: "delete:/api/v1/finance/invoices/${id}",
    })
}
```

### After (modern — raw ast.Node, when DSL screen is not flexible enough)

```go
func Schema(sess ui.UISessionContext) any {
    return ast.PageNode{
        Title: "Invoices",
        Data: ui.M{
            "can_create": sess.Can("create", "invoice"),
        },
        Body: ast.CRUDNode{
            // syncLocation: true is automatic — no need to set it
            API: ast.APIGET("/api/v1/finance/invoices"),
            HeaderToolbar: []any{
                ast.ButtonNode{
                    Label:     "New Invoice",
                    Level:     "primary",
                    VisibleOn: "${can_create}",
                },
            },
            Columns: []ast.ColumnNode{
                {Name: "number", Label: "Invoice #"},
                {Name: "status", Label: "Status"},
            },
        },
    }
}
```

---

## Migration Checklist

For each legacy page:

- [ ] Change function signature from `func Schema(sess ui.UISessionContext) ui.Schema` → `func Schema(sess ui.UISessionContext) any`
- [ ] Change registration from `Fn: Schema` → `ASTFn: Schema`
- [ ] Add `Description:` field to `PageRegistration` (optional but recommended)
- [ ] Replace `amis.M{"type": "crud", ...}` with `ast.CRUDNode{...}` or `screens.CRUDScreen(...)`
  - Remove manual `"syncLocation": true` — `CRUDNode` sets it automatically
- [ ] Replace `amis.M{"type": "chart", ...}` with `ast.ChartNode{...}`
  - Remove manual transparent background style — `ChartNode` sets it automatically
- [ ] Ensure all API strings use method prefix: `"get:/api/..."` — enforced at compile time in AST nodes
- [ ] Remove old `Fn:` field from `PageRegistration` once `ASTFn:` is working
- [ ] Test: hit `/schema/<route>` and verify no 500 (CompileStage panic means type mismatch)

---

## Dual-Registration (safe migration window)

During migration, set both fields. `CompileStage` prefers `ASTFn`:

```go
registry.RegisterPage(registry.PageRegistration{
    Route:  "/finance/invoices",
    Module: "finance",
    Title:  "Invoices",
    Fn:     LegacySchema,  // active until ASTFn is ready
    ASTFn:  NewSchema,     // CompileStage always picks this when non-nil
})
```

This lets you deploy `ASTFn` behind `Fn` and verify it in staging before removing `Fn`. No downtime, no feature flag needed.

---

## Common Errors During Migration

### `ASTPageFn return value does not implement ast.Node`

**Cause:** Function returns a type that doesn't satisfy `ast.Node`.

**Fix:** Use types from `internal/web/ast/` (e.g. `ast.PageNode`, `ast.CRUDNode`). Check the package for the current node catalog. The DSL screens in `internal/web/dsl/screens/` return `ast.Node` and are the safest option.

### 500 after migration, was 200 before

**Cause:** `CompileStage` panic — the AST node has a required field not set or a type assertion failed.

**Diagnose:** Check server logs for `panic in ASTPageFn`. The log line includes the panic message and the route.

**Fix:** Match the error to the failing node type. Missing `API` field on `CRUDNode` is the most common cause.

### Cache serving the old schema after migration

**Cause:** Old `PageFn` schema cached under the same key. Since the cache key includes route + permFP + flagFP (not the schema content itself), the new `ASTFn` result won't be fetched until TTL expires or the key is evicted.

**Fix (development):** Restart server to clear in-process LRU. Or call `DeletePattern` for the route.

**Fix (production):** Bump `SchemaGeneration` in `CacheVersions` at deploy time. This changes all cache keys → full miss on first request after deploy.

---

## Permissions: Nothing Changes

`sess.Can(action, resource)` works identically in both legacy and modern pages. The permission fingerprint is computed by `AuthzStage` before the page function runs, regardless of whether `Fn` or `ASTFn` is used.

The only difference: in the old `amis.Ctx` approach, `Can` was a closure that called the authz service inline. In the modern approach, `Can` reads from pre-resolved permissions stored in `UISessionContext`. The result is the same; the modern approach is faster (one Casbin call per request, not one per `Can()` invocation).

---

## What NOT to Migrate

Do not migrate pages that are under active development for an upcoming release. Do not create a migration branch that touches >5 pages. Migrate pages in the course of regular feature work — when you open a page file to add a column or fix a bug, migrate it at the same time.

Migrations that span many files in one PR make review harder and increase the chance of a silent regression (wrong column order, missing permission gate). Small, incremental migrations are safer.
