> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "46 – Best Practices"
volume: "vol-10-reference"
chapter: 46
section: "Reference"
status: "implemented"
---

# Chapter 46 – Best Practices

Ten proven practices from the codebase, each backed by a concrete rationale.

## Table of Contents
1. [Value Receivers on Node Interface Methods](#1-value-receivers-on-node-interface-methods)
2. [CompileTree Collects All Errors Before Emitting](#2-compiletree-collects-all-errors-before-emitting)
3. [Compile-Time Interface Assertions](#3-compile-time-interface-assertions)
4. [UISessionContext Is a Value Type](#4-uisessioncontext-is-a-value-type)
5. [Pre-Resolved Permissions — No Casbin Inside Block Functions](#5-pre-resolved-permissions--no-casbin-inside-block-functions)
6. [Permission Fingerprinting — Share Cache Across Users](#6-permission-fingerprinting--share-cache-across-users)
7. [ASTPageFn Preferred Over PageFn](#7-astpagefn-preferred-over-pagefn)
8. [init() Pattern for Page Registration](#8-init-pattern-for-page-registration)
9. [60-Line Screen Size Limit](#9-60-line-screen-size-limit)
10. [Blocks Own Permission Checks — Screens Are Pure Composition](#10-blocks-own-permission-checks--screens-are-pure-composition)

---

## 1. Value Receivers on Node Interface Methods

All methods on `ast.*` node structs that implement the `Node` interface use
**value receivers**, not pointer receivers.

```go
// Correct — value receiver enforces immutability
func (n PageNode) Children() []Node   { return n.Body }
func (n PageNode) NodeType() string   { return "page" }
func (n PageNode) Validate() []error  { return nil }

// Wrong — pointer receiver allows mutation
func (n *PageNode) Children() []Node  { return n.Body } // do not do this
```

**Rationale**: nodes are constructed once and read many times (by `CompileTree`,
`ValidateStage`, serialisation).  Value receivers prevent accidental mutation of
shared nodes when the same block is composed into multiple screens.  If
modification of a node is needed, copy-and-modify rather than mutating in place.

---

## 2. CompileTree Collects All Errors Before Emitting

`ast.CompileTree` walks the entire node tree and **collects every error** before
returning.  It never short-circuits on the first error.

```go
func CompileTree(root Node) []error {
    var errs []error
    walkNodes(root, func(n Node) {
        if ve := n.Validate(); len(ve) > 0 {
            errs = append(errs, ve...)
        }
    })
    return errs
}
```

**Rationale**: short-circuiting on the first error forces iterative fix-one,
recompile, fix-one workflows.  Collecting all errors lets a developer fix every
problem in a single pass.  This is especially important during initial screen
development when a new block may have multiple issues.

In tests, always assert `len(errs) == 0` rather than `errs == nil` — the
difference matters when `CompileTree` returns an empty (non-nil) slice.

---

## 3. Compile-Time Interface Assertions

Every `ast.*` node struct declares a compile-time assertion that it implements
`Node`:

```go
// ast/nodes.go
var (
    _ Node = PageNode{}
    _ Node = CRUDNode{}
    _ Node = PanelNode{}
    _ Node = TabsNode{}
    _ Node = FormNode{}
    _ Node = StatNode{}
    _ Node = ChartNode{}
)
```

**Rationale**: these blank-identifier assignments fail to compile if a struct
misses a method in the `Node` interface.  They surface missing implementations
immediately — before any test runs — rather than at the point of use in a block
function.

Add a new assertion line whenever a new node type is added to the `ast` package.

---

## 4. UISessionContext Is a Value Type

`UISessionContext` is defined as a **struct**, not a pointer to a struct, and it
is passed by value everywhere.

```go
type UISessionContext struct {
    TenantID    string
    UserID      string
    Permissions []string
    Flags       map[string]bool
    Locale      string
    Timezone    string
    Currency    string
    // ...
}

// Passed by value — each callee gets its own copy
func KPIRowBlock(sess ui.UISessionContext) ast.Node { ... }
```

**Rationale**: passing by value means each block function receives its own copy
of the session.  A block cannot mutate the session seen by sibling blocks.  The
value type also makes it safe to cache a compiled schema keyed on the session
fingerprint — the schema is derived from an immutable snapshot of the session.

The `[]string` and `map[string]bool` fields inside `UISessionContext` are
treated as read-only by convention — block functions must not append to
`Permissions` or write to `Flags`.

---

## 5. Pre-Resolved Permissions — No Casbin Inside Block Functions

Permissions are resolved **before** the pipeline starts.  `AuthStage` calls
Casbin and populates `UISessionContext.Permissions` as a `[]string` once per
request.  Block functions call `sess.HasPermission("finance.invoices.create")`
which is a simple slice membership check — O(n) with n typically < 100.

```go
// AuthStage — called once per request
func (s *AuthStage) Execute(ctx context.Context, pCtx *PipelineContext) error {
    perms, err := s.iam.ResolvePermissions(ctx, userID, tenantID)
    if err != nil {
        return err
    }
    pCtx.Session.Permissions = perms // pre-resolved
    return nil
}

// Block function — no Casbin call
func QuickActionsBlock(sess ui.UISessionContext, ...) ast.Node {
    if !sess.HasPermission("finance.invoices.create") { // O(n) slice check
        return ast.NullNode{}
    }
    // ...
}
```

**Rationale**: Casbin evaluation is non-trivial — it involves policy table
lookups and rule matching.  Calling it inside a block function would make block
functions stateful (they'd need a context and a client), would add latency to
every block call, and would make block functions untestable without a real Casbin
setup.

---

## 6. Permission Fingerprinting — Share Cache Across Users

Multiple users with identical permission sets share the same cached schema.
`CacheStage` hashes `UISessionContext.Permissions` into a `permFingerprint`
string used as part of the cache key.

```go
func permFingerprint(perms []string) string {
    sorted := make([]string, len(perms))
    copy(sorted, perms)
    sort.Strings(sorted)
    h := sha256.Sum256([]byte(strings.Join(sorted, ",")))
    return hex.EncodeToString(h[:8]) // 8-byte prefix sufficient
}
```

**Rationale**: without fingerprinting, every user would require their own cache
entry even when their permissions are identical.  In a 500-user tenant, there
may be only 5 distinct permission profiles — so fingerprinting yields a ~100x
cache entry reduction.  The sort-before-hash step ensures that permission order
does not affect the fingerprint.

---

## 7. ASTPageFn Preferred Over PageFn

New screens must be registered as `ASTPageFn` (returns `ast.Node`), not `PageFn`
(returns `any`).

```go
// Preferred
registry.RegisterPage(registry.PageRegistration{
    ASTFn: MyScreen,   // ASTPageFn — typed, validated
})

// Avoid for new screens
registry.RegisterPage(registry.PageRegistration{
    Fn: MyScreenFn,    // PageFn — untyped, less validation
})
```

**Rationale**: `ASTPageFn` unlocks `CompileTree` validation, which catches
structural errors before the schema reaches `ValidateStage`.  Typed nodes
produce better error messages and are testable with simple type assertions.  See
§44 for migration guidance for existing `PageFn` screens.

---

## 8. init() Pattern for Page Registration

Each screen file registers its page(s) in an `init()` function.  This means the
registration happens automatically when the package is imported — zero wiring
code in `main`.

```go
// screens/finance_dashboard.go
func init() {
    registry.RegisterPage(registry.PageRegistration{
        Route:  "/finance/dashboard",
        Module: "finance",
        Title:  "Finance Dashboard",
        ASTFn:  FinanceDashboardScreen,
    })
}
```

The blank import in the application entry point causes `init()` to run:

```go
// cmd/server/main.go
import (
    _ "github.com/your-org/erp/internal/web/dsl/screens" // triggers all init()
)
```

**Rationale**: explicit wiring (calling `Register` from `main`) creates a file
that must be updated every time a screen is added — a maintenance burden.  The
`init()` pattern makes new screens self-registering; adding a screen file to the
package is sufficient.

---

## 9. 60-Line Screen Size Limit

A screen file may not exceed 60 lines.  When a screen grows beyond this, extract
a block.

```
screens/finance_dashboard.go — 38 lines  ✓
screens/finance_invoice_list.go — 44 lines  ✓
screens/iam_users.go — 67 lines  ✗ — needs block extraction
```

**Rationale**: see §45.2 for full reasoning.  In brief: screen files that exceed
60 lines have absorbed logic that belongs in `blocks/`.  The limit is a
governance signal enforced by CI guard #8 (planned) and code review.

---

## 10. Blocks Own Permission Checks — Screens Are Pure Composition

A block function that requires a permission checks it internally.  The screen
that composes the block does not duplicate or pre-check that permission.

```go
// Block — owns the check
func ApproveInvoiceBlock(sess ui.UISessionContext) ast.Node {
    if !sess.HasPermission("finance.invoices.approve") {
        return ast.NullNode{} // invisible to the screen
    }
    return ast.ActionNode{
        Label:      "Approve",
        ActionType: "dialog",
        Dialog:     approveDialog(),
    }
}

// Screen — composes blindly, trusts block to gate itself
func InvoiceDetailScreen(sess ui.UISessionContext) ast.Node {
    return ast.PageNode{
        Body: []ast.Node{
            blocks.InvoiceHeaderBlock(sess),
            blocks.InvoiceLinesBlock(sess),
            blocks.ApproveInvoiceBlock(sess), // screen doesn't check the permission
        },
    }
}
```

**Rationale**: screens are layout.  Business rules — including who can see what
— are block concerns.  Putting permission checks in screens means the same check
must be duplicated everywhere the block is used.  When the permission key
changes, every screen must be updated.  Block-owned checks ensure the rule lives
in exactly one place.
