> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "47 – Anti-Patterns"
volume: "vol-10-reference"
chapter: 47
section: "Reference"
status: "reference — all confirmed bugs from code review"
---

# Chapter 47 – Anti-Patterns

Every entry in this chapter is a real bug found during code review.  Each
anti-pattern gets: what it is, why it is wrong, and the correct alternative.

## Table of Contents
1. [Using "status" as an AMIS Column Type](#1-using-status-as-an-amis-column-type)
2. [Single-Row Table for Detail Views](#2-single-row-table-for-detail-views)
3. [Negation Expression Syntax Error](#3-negation-expression-syntax-error)
4. [resourceFromURL for Permission Checks](#4-resourcefromurl-for-permission-checks)
5. [DetailCardBlock with ResponseData Restriction](#5-detailcardblock-with-responsedata-restriction)
6. [map[string]any in dsl/](#6-mapstringany-in-dsl)
7. [Custom CSS Class Names on StatNode](#7-custom-css-class-names-on-statnode)
8. [Chart Node Missing Style](#8-chart-node-missing-style)
9. [TabsNode Mountable bool Field](#9-tabsnode-mountable-bool-field)
10. [ActionNode Without Dialog Field](#10-actionnode-without-dialog-field)
11. [CRUDNode RowActions Excluded from Children](#11-crudnode-rowactions-excluded-from-children)
12. [Permission Check Before Calling a Block](#12-permission-check-before-calling-a-block)

---

## 1. Using "status" as an AMIS Column Type

**What:** Setting `Type: "status"` on a `TableColumnNode` to display a status badge.

**Why it's wrong:** `"status"` is not a valid AMIS column type.  AMIS will fall
back to rendering the raw string value — no badge, no colour coding.

**What to do instead:** Use `Type: "mapping"` with an explicit `Map` that maps
status strings to badge labels and colours.

```go
// Wrong
ast.TableColumnNode{Name: "status", Label: "Status", Type: "status"}

// Correct
ast.TableColumnNode{
    Name:  "status",
    Label: "Status",
    Type:  "mapping",
    Map: map[string]string{
        "PENDING":   "label label-warning",
        "ACTIVE":    "label label-success",
        "SUSPENDED": "label label-danger",
        "ARCHIVED":  "label label-default",
    },
}
```

---

## 2. Single-Row Table for Detail Views

**What:** Wrapping a single record's fields in a `CRUDNode` or table to display
a detail view — producing a table with exactly one row.

**Why it's wrong:** A one-row table is semantically wrong, wastes space with a
table header, and produces confusing pagination controls.  AMIS renders it as a
degenerate grid.

**What to do instead:** Use `ast.PropertyNode` for detail views.  It renders
field/value pairs in a readable layout without table chrome.

```go
// Wrong — one-row table
ast.CRUDNode{
    API:     &ast.APISpec{URL: "/api/v1/invoices/${id}"},
    Columns: []ast.TableColumnNode{{Name: "number"}, {Name: "amount"}},
}

// Correct — property node
ast.PropertyNode{
    Source: "${body}",
    Items: []ast.PropertyItem{
        {Label: "Invoice Number", Content: "${number}"},
        {Label: "Amount",         Content: "${amount}"},
        {Label: "Status",         Content: "${status}"},
    },
}
```

---

## 3. Negation Expression Syntax Error

**What:** Using `"!${can_approve}"` in an AMIS `visibleOn` or `disabledOn`
expression to negate a boolean variable.

**Why it's wrong:** AMIS expression engine evaluates `"!${can_approve}"` by
substituting `can_approve`, yielding e.g. `"!false"`.  The `!` operator in this
position operates on the **string** `"false"`, which is truthy in JavaScript —
the expression always evaluates to `false`, meaning the element is always
hidden.

**What to do instead:** Use `"${!can_approve}"` — place the negation operator
_inside_ the expression interpolation, where the AMIS engine evaluates it as a
JavaScript boolean operation.

```go
// Wrong — negates a string, always falsy result
ast.ActionNode{
    Label:     "Approve",
    DisabledOn: "!${can_approve}",  // bug: "!false" == false always
}

// Correct — negation inside expression
ast.ActionNode{
    Label:     "Approve",
    DisabledOn: "${!can_approve}",  // correct: !false == true
}
```

---

## 4. resourceFromURL for Permission Checks

**What:** Using a helper `resourceFromURL(route)` to extract a resource
identifier for permission checks.

**Why it's wrong:** `resourceFromURL` prepends a dot separator, yielding keys
like `".finance.invoices"` with a leading dot.  These keys never match any
registered permission (which have the form `"finance.invoices.read"`), so the
permission check always returns false — every user is denied.

**What to do instead:** Use explicit permission string constants, not derived
keys.

```go
// Wrong — produces ".finance.invoices" — always denied
if !sess.HasPermission(resourceFromURL(route)) {
    return ast.NullNode{}
}

// Correct — explicit permission key
if !sess.HasPermission("finance.invoices.read") {
    return ast.NullNode{}
}
```

If you need to check permissions for a parameterised resource, derive the key
explicitly without `resourceFromURL`.

---

## 5. DetailCardBlock with ResponseData Restriction

**What:** Wrapping a detail panel in `DetailCardBlock` with a `ResponseData`
field that restricts the data scope.

**Why it's wrong:** Setting `ResponseData: "${body}"` on a `PanelNode` (or
equivalent wrapper) restricts the AMIS data scope to `body`.  Child nodes that
reference `${totals}` or `${tax_lines}` — data that exists at the page scope
but not inside `body` — silently receive `undefined` and render as empty.

**What to do instead:** Avoid `ResponseData` restrictions on panels that have
children needing broader scope.  Use AMIS `data` merging patterns or restructure
the `InitAPI` response to flatten required data to a single scope.

```go
// Wrong — restricts scope, children can't access ${totals} or ${tax_lines}
ast.PanelNode{
    Source:       "${body}",
    ResponseData: "${body}",  // kills ${totals}, ${tax_lines} for all children
    Body:         []ast.Node{
        blocks.InvoiceLinesBlock(sess),  // needs ${tax_lines} — silently empty
        blocks.TotalsBlock(sess),        // needs ${totals} — silently empty
    },
}

// Correct — no ResponseData restriction; all page-scope data available
ast.PanelNode{
    Body: []ast.Node{
        blocks.InvoiceLinesBlock(sess),
        blocks.TotalsBlock(sess),
    },
}
```

---

## 6. map[string]any in dsl/

**What:** Using `map[string]any{"type": "stat", "label": "Revenue", ...}` to
construct schema nodes in `dsl/` code.

**Why it's wrong:** `map[string]any` bypasses all typed validation.  Typos in
field names are silent.  `CompileTree` cannot inspect or validate the map
contents.  `ValidateStage` sees an opaque blob.  Bugs surface as rendering
failures in the browser.

**What to do instead:** Use `ast.*` struct literals.  If no struct exists for
the AMIS component you need, add one to the `ast` package.

```go
// Wrong — zero tolerance
map[string]any{
    "type":  "stat",
    "label": "Revenue",
    "value": "${revenue}",
}

// Correct
ast.StatNode{
    Label: "Revenue",
    Value: "${revenue}",
}
```

This rule is enforced by CI guard #1 (§42.3) and is a blocking code review
comment with no exceptions.

---

## 7. Custom CSS Class Names on StatNode

**What:** Adding custom CSS class strings (e.g. `ClassName: "my-stat-card"`)
to `StatNode` and defining those classes in a custom stylesheet.

**Why it's wrong:** Custom class names sit outside the AMIS SDK CSS custom
property system.  They conflict with AMIS SDK upgrades (which may change
internal class names), are not scoped to the dark/light theme token system, and
produce whack-a-mole override fights when AMIS updates.

**What to do instead:** Override AMIS CSS custom properties at the theme level
in `web/styles/css/`.  Never add custom classes to nodes generated by the DSL.

```go
// Wrong — custom class that fights with AMIS internals
ast.StatNode{
    Label:     "Revenue",
    Value:     "${revenue}",
    ClassName: "my-stat-kpi-card",  // do not do this
}

// Correct — no class, themed via CSS custom properties
ast.StatNode{
    Label: "Revenue",
    Value: "${revenue}",
}
```

Theme customisation belongs in `web/styles/css/` using AMIS CSS custom property
tokens (e.g. `--colors-neutral-fill-11`, `--Panel-bg-color`).

---

## 8. Chart Node Missing Style

**What:** Declaring an `ast.ChartNode` without setting the `Style` field.

**Why it's wrong:** `ValidateStage` runs a validation rule that requires every
`ChartNode` to have a non-empty `Style`.  A chart without `Style` causes
`ValidateStage` to return an error, which the pipeline maps to HTTP 500.  Every
request for a page containing that chart returns a 500 until the bug is fixed.

**What to do instead:** Always set `Style` on `ChartNode`.

```go
// Wrong — ValidateStage will reject this with HTTP 500
ast.ChartNode{
    ChartType: "bar",
    Data:      "${revenue_by_month}",
    // Style is missing — server error on every render
}

// Correct
ast.ChartNode{
    ChartType: "bar",
    Data:      "${revenue_by_month}",
    Style:     map[string]any{"height": "300px"},
}
```

The validation error code emitted is `VALIDATE_CHART_MISSING_STYLE`.

---

## 9. TabsNode Mountable bool Field

**What:** Setting `Mountable: true` on a `TabsNode` to control tab mounting
behaviour.

**Why it's wrong:** `TabsNode` has no `Mountable` field.  The correct fields for
controlling tab mount behaviour are `MountOnEnter` (mount tab content when the
tab is first activated) and `UnmountOnExit` (destroy tab content when switching
away).  Setting a non-existent `Mountable` field is either a compile error or
silently ignored.

**What to do instead:**

```go
// Wrong — Mountable does not exist on TabsNode
ast.TabsNode{
    Mountable: true,  // compile error or silently ignored
    Tabs: []ast.TabItem{...},
}

// Correct
ast.TabsNode{
    MountOnEnter:  true,  // mount content when tab becomes active
    UnmountOnExit: false, // keep content alive when switching away
    Tabs: []ast.TabItem{...},
}
```

---

## 10. ActionNode Without Dialog Field

**What:** Setting `ActionType: "dialog"` on an `ActionNode` without setting the
`Dialog` field.

**Why it's wrong:** When `ActionType` is `"dialog"`, the AMIS SDK expects the
`dialog` property to be a valid dialog schema.  If `Dialog` is nil/empty, the
AMIS runtime panics when the user clicks the button — a client-side JavaScript
error that crashes the page.

**What to do instead:** Always pair `ActionType: "dialog"` with a non-nil
`Dialog`.

```go
// Wrong — runtime panic on click
ast.ActionNode{
    Label:      "Approve",
    ActionType: "dialog",
    // Dialog is nil — browser crash on click
}

// Correct
ast.ActionNode{
    Label:      "Approve",
    ActionType: "dialog",
    Dialog: &ast.DialogNode{
        Title: "Confirm Approval",
        Body:  []ast.Node{ast.TplNode{Tpl: "Approve this invoice?"}},
        Actions: []ast.ActionNode{
            {Label: "Confirm", ActionType: "ajax", API: approveAPI},
            {Label: "Cancel",  ActionType: "close"},
        },
    },
}
```

`ValidateStage` enforces this: a `dialog`-type `ActionNode` with no `Dialog`
emits `VALIDATE_ACTION_MISSING_DIALOG`.

---

## 11. CRUDNode RowActions Excluded from Children

**What:** Assuming that `CRUDNode.RowActions` is traversed by `CompileTree` as
part of the node's `Children()` method.

**Why it's wrong:** The `CRUDNode.Children()` implementation does not include
`RowActions` in its returned slice.  This means validation errors inside
`RowActions` (e.g. a `dialog`-type action without a `Dialog`) are **silently
skipped** by `CompileTree` — they produce no compile error.  The bug only
surfaces at runtime.

**What to do instead:** Validate `RowActions` manually in unit tests by calling
each action's `Validate()` method, or add them explicitly to the `Children()`
implementation (preferred fix to the `ast` package).

```go
// Bug: CompileTree won't catch errors in RowActions
node := ast.CRUDNode{
    API: &ast.APISpec{URL: "/api/v1/invoices"},
    RowActions: []ast.ActionNode{
        {Label: "Edit", ActionType: "dialog"}, // missing Dialog — not caught by CompileTree
    },
}
errs := ast.CompileTree(node) // returns empty — RowActions not traversed

// Manual validation workaround
for _, action := range node.RowActions {
    errs = append(errs, action.Validate()...)
}
```

Until `CRUDNode.Children()` is fixed to include `RowActions`, every CRUDNode
with row actions should be validated in a dedicated unit test.

---

## 12. Permission Check Before Calling a Block

**What:** Checking a permission in a screen function before deciding whether to
call a block, instead of letting the block handle its own permission gate.

**Why it's wrong:** This duplicates the permission check that the block already
performs internally.  When the permission key changes or the block's gating logic
changes, the screen must be updated too — and the duplication will eventually
drift out of sync.  It also violates the governance rule that screens are pure
composition (§45.4).

**What to do instead:** Call the block unconditionally.  Let the block return
`ast.NullNode{}` (or a reduced schema) when the permission is absent.

```go
// Wrong — screen duplicates block's permission logic
func InvoiceDetailScreen(sess ui.UISessionContext) ast.Node {
    body := []ast.Node{blocks.InvoiceHeaderBlock(sess)}

    // Duplicates what ApproveInvoiceBlock already checks internally
    if sess.HasPermission("finance.invoices.approve") {
        body = append(body, blocks.ApproveInvoiceBlock(sess))
    }

    return ast.PageNode{Body: body}
}

// Correct — screen composes blindly; block gates itself
func InvoiceDetailScreen(sess ui.UISessionContext) ast.Node {
    return ast.PageNode{
        Body: []ast.Node{
            blocks.InvoiceHeaderBlock(sess),
            blocks.ApproveInvoiceBlock(sess), // returns NullNode if no permission
        },
    }
}
```
