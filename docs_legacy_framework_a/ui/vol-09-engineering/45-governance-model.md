> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "45 – Governance Model"
volume: "vol-09-engineering"
chapter: 45
section: "Engineering"
status: "implemented"
---

# Chapter 45 – Governance Model

## Table of Contents
- [45.1 Block Ownership vs Screen Ownership](#451-block-ownership-vs-screen-ownership)
- [45.2 The 60-Line Screen Limit](#452-the-60-line-screen-limit)
- [45.3 Zero map[string]any Rule in dsl/](#453-zero-mapstringany-rule-in-dsl)
- [45.4 Permission Rule: Blocks Own Checks, Screens Are Pure Composition](#454-permission-rule-blocks-own-checks-screens-are-pure-composition)
- [45.5 New Block Extraction Process](#455-new-block-extraction-process)

---

## 45.1 Block Ownership vs Screen Ownership

The DSL layer has two distinct ownership domains:

| Domain | Location | Owner | Responsibility |
|---|---|---|---|
| Blocks | `internal/web/dsl/blocks/` | Shared — any module team | Reusable composable UI components, own their permissions |
| Screens | `internal/web/dsl/screens/` | Individual module team | Route → composed schema, pure composition only |

**Blocks are shared infrastructure.**  A block once written in `blocks/` can be
used by any screen in any module.  Block changes are breaking changes that affect
all consumers — they require cross-team review.

**Screens are module-owned.**  Each module team owns the screen files for their
module routes.  A screen file should not be modified by a different module team
without the owning team's sign-off.

### Naming convention

Block files are named by their domain concern:
```
blocks/
  finance_quick_actions.go    // QuickActionsBlock for finance
  invoice_header.go           // InvoiceHeaderBlock
  shared_stat_row.go          // StatRowBlock — domain-agnostic
```

Screen files are named by their route:
```
screens/
  finance_dashboard.go        // /finance/dashboard
  finance_invoice_list.go     // /finance/invoices
  iam_users.go                // /iam/users
```

---

## 45.2 The 60-Line Screen Limit

A screen file over 60 lines is a governance signal, not a style preference.  It
indicates that logic which belongs in a block has escaped into the screen layer.

**What 60 lines accommodates:**

```go
package screens

import (
    "github.com/your-org/erp/internal/web/dsl/ast"
    "github.com/your-org/erp/internal/web/dsl/blocks"
    ui "github.com/your-org/erp/internal/web/ui"
    "github.com/your-org/erp/internal/web/dsl/registry"
)

func init() {
    registry.RegisterPage(registry.PageRegistration{
        Route:  "/finance/dashboard",
        Module: "finance",
        Title:  "Finance Dashboard",
        ASTFn:  FinanceDashboardScreen,
    })
}

func FinanceDashboardScreen(sess ui.UISessionContext) ast.Node {
    return ast.PageNode{
        Title: "Finance Dashboard",
        InitAPI: &ast.APISpec{
            Method: "get",
            URL:    "/api/v1/finance/dashboard/summary",
        },
        Body: []ast.Node{
            blocks.KPIRowBlock(sess),
            blocks.RevenueChartBlock(sess),
            blocks.QuickActionsBlock(sess, financeQuickActions),
            blocks.RecentInvoicesBlock(sess),
        },
    }
}

var financeQuickActions = []blocks.QuickAction{
    {Label: "New Invoice",    Permission: "finance.invoices.create"},
    {Label: "Record Payment", Permission: "finance.payments.create"},
}
```

That is a complete, production-quality screen at ~40 lines including imports and
the `var` block.  A screen that needs more than 60 lines is doing block work.

**Enforcement**: see §42.4 for the manual line-count check command.

---

## 45.3 Zero map[string]any Rule in dsl/

No file under `internal/web/dsl/` may contain `map[string]any`.  This rule is
absolute — it has no exceptions.

**Why:**
- `map[string]any` bypasses `ast.CompileTree` validation entirely.
- Type errors in map keys/values are silent at compile time and surface as
  rendering bugs in the browser.
- `ValidateStage` cannot inspect the contents of a `map[string]any` node — it
  can only check the outer structure.

**What to do instead:**

```go
// Wrong — zero tolerance
Body: []ast.Node{
    map[string]any{
        "type":  "stat",
        "label": "Revenue",
        "value": "${revenue}",
    },
}

// Correct
Body: []ast.Node{
    ast.StatNode{
        Label: "Revenue",
        Value: "${revenue}",
    },
}
```

If a needed AMIS component has no `ast.*` struct, the correct action is to
**add a new struct to the `ast` package** — not to use `map[string]any` as a
workaround.

This rule is enforced by CI guard #1 (§42.3) and is a blocking code review
comment.

---

## 45.4 Permission Rule: Blocks Own Checks, Screens Are Pure Composition

Permissions are checked inside block functions, not in screen functions.  A
screen function never inspects `sess.Permissions` directly.

**Correct — block owns the permission check:**

```go
// blocks/finance_quick_actions.go
func QuickActionsBlock(sess ui.UISessionContext, actions []QuickAction) ast.Node {
    var buttons []ast.ActionNode
    for _, a := range actions {
        if sess.HasPermission(a.Permission) {   // permission check lives here
            buttons = append(buttons, ast.ActionNode{Label: a.Label})
        }
    }
    return ast.ToolbarNode{Buttons: buttons}
}
```

**Correct — screen is pure composition:**

```go
// screens/finance_dashboard.go
func FinanceDashboardScreen(sess ui.UISessionContext) ast.Node {
    return ast.PageNode{
        Body: []ast.Node{
            blocks.QuickActionsBlock(sess, financeQuickActions), // no perm check here
        },
    }
}
```

**Wrong — screen checks permissions:**

```go
// screens/finance_dashboard.go — WRONG
func FinanceDashboardScreen(sess ui.UISessionContext) ast.Node {
    body := []ast.Node{blocks.KPIRowBlock(sess)}
    if sess.HasPermission("finance.invoices.create") {  // WRONG: belongs in block
        body = append(body, blocks.QuickActionsBlock(sess, financeQuickActions))
    }
    return ast.PageNode{Body: body}
}
```

**Rationale**: screens are pure composition.  When permission logic moves into
screens, it duplicates what the block would do anyway — and diverges when the
block is updated.  A block called with insufficient permissions should return
a reduced (or empty) node, not be withheld by the screen.

---

## 45.5 New Block Extraction Process

When should a piece of screen code be extracted into a block?

**Extract when any of these is true:**
1. The same UI pattern appears in more than one screen file.
2. The screen file would exceed 60 lines without extraction.
3. The UI pattern has its own permission gate.
4. The UI pattern makes its own `InitAPI` call.
5. The pattern is complex enough that it benefits from isolated testing.

**Extraction steps:**

1. Create `internal/web/dsl/blocks/<domain>_<name>.go`.
2. Define `func <Name>Block(sess ui.UISessionContext, ...) ast.Node`.
3. Move the permission check (if any) into the block function.
4. Write a unit test: `blocks/<domain>_<name>_test.go`.
5. Replace the inlined code in the original screen with a call to the new block.
6. Verify the screen file is back under 60 lines.
7. Check no other screen duplicates the same pattern — migrate those too.

**Block function signatures:**

Blocks that need no configuration beyond `sess`:
```go
func KPIRowBlock(sess ui.UISessionContext) ast.Node
```

Blocks that accept configuration:
```go
func QuickActionsBlock(sess ui.UISessionContext, actions []QuickAction) ast.Node
```

Avoid blocks that accept generic `any` configuration — use typed config structs.
