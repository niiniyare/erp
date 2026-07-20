> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Chapter 07 — AST Design

> **Volume:** II — DSL and AST
> **Audience:** Platform Engineers, Backend Engineers
> **Prerequisites:** Chapter 06 — UI DSL Architecture

---

## Table of Contents

- [7.1 Why an AST?](#71-why-an-ast)
- [7.2 The Node Interface](#72-the-node-interface)
- [7.3 ContainerNode and the Tree Structure](#73-containernode-and-the-tree-structure)
- [7.4 CompileTree: Validate First, Emit Second](#74-compiletree-validate-first-emit-second)
- [7.5 Node Types Reference](#75-node-types-reference)
- [7.6 Struct Literal Construction](#76-struct-literal-construction)
- [7.7 APISpec: The Data Binding Type](#77-apispec-the-data-binding-type)
- [7.8 ActionNode: Actions and Dialogs](#78-actionnode-actions-and-dialogs)
- [7.9 Validation Errors](#79-validation-errors)
- [7.10 What the AST Does Not Model](#710-what-the-ast-does-not-model)

---

## 7.1 Why an AST?

Before the typed AST, page functions returned `map[string]any` directly. Errors in the schema — wrong `type` string, missing required field, incorrect nesting — were invisible until the browser tried to render the result. Sometimes they were invisible even then, because AMIS silently ignores unknown fields.

The typed AST catches structural errors before JSON is emitted:
- A `CRUDNode` with no `Columns` fails `Validate()`.
- A `ChartNode` must have `Style.Background = "transparent"` and `Config.BackgroundColor = "transparent"` — validated structurally.
- A `TabsNode` must use `MountOnEnter`/`UnmountOnExit`, not `Mountable` — the struct type enforces the correct fields.

`CompileTree` collects all validation errors from the entire tree before emitting any JSON. If any node fails, the whole compilation fails with a list of all errors. The pipeline returns an error to the caller; no broken schema reaches the browser.

---

## 7.2 The Node Interface

```go
// internal/web/ast/node.go

// Node is the base interface for every typed AMIS schema node.
// All implementations use value receivers to enforce immutability.
type Node interface {
    // NodeType returns the AMIS "type" field value.
    // Examples: "page", "crud", "form", "chart", "grid"
    NodeType() string

    // Validate checks the node's invariants.
    // Returns nil when valid.
    // Returns *sharedErrors.BusinessError with code prefix "AST_" when invalid.
    // Called by CompileTree before Compile — never call Validate inside Compile.
    Validate() error

    // Compile emits the node as AMIS-compliant map[string]any.
    // Must only be called after Validate() returns nil.
    Compile() map[string]any
}

// ContainerNode extends Node for nodes that own child nodes.
// CompileTree recurses into children via Children() for validation.
type ContainerNode interface {
    Node
    Children() []Node
}
```

The `any` return type in `ASTPageFn` (`func(sess UISessionContext) any`) avoids a circular import between `ui` and `ast` packages. At runtime, `CompileStage` asserts the returned value to `ast.Node`.

Every concrete node type must include a compile-time assertion:

```go
var _ ast.Node = PageNode{}
var _ ast.ContainerNode = GridNode{}
```

---

## 7.3 ContainerNode and the Tree Structure

Nodes that contain children implement `ContainerNode`. `CompileTree` uses `Children()` to recurse through the tree for validation and compilation.

A page is a tree:

```
PageNode (ContainerNode)
├── GridNode (ContainerNode)
│   ├── ChartNode (Node)
│   └── StatNode (Node)
├── CRUDNode (ContainerNode)
│   ├── ActionNode (row action — Node)
│   └── FilterBarNode (Node)
└── FormNode (ContainerNode)
    ├── InputTextNode (Node)
    ├── SelectNode (Node)
    └── ActionNode (form action — Node)
```

**Important:** `CRUDNode.RowActions` must be included in `CRUDNode.Children()` for validation. RowActions are `ActionNode` values that can contain `Dialog` or `Drawer` fields — these need validation. Failing to include them in `Children()` would silently skip their validation.

---

## 7.4 CompileTree: Validate First, Emit Second

```go
// ast/compile.go

// CompileTree validates all nodes in the tree before emitting any JSON.
// If any node fails Validate(), returns a combined error listing all failures.
// On success, returns the compiled map[string]any for the root node.
func CompileTree(root Node) (map[string]any, error)
```

The algorithm:
1. Walk the tree depth-first, calling `Validate()` on every node.
2. Collect all errors — do not short-circuit on the first failure.
3. If any errors exist, return a `BusinessError` listing all of them.
4. If the tree is fully valid, walk again calling `Compile()` and build the JSON.

This two-pass design means authors get all validation errors at once, not one at a time. A page with three structural problems produces three errors in a single compile attempt.

```go
// In CompileStage, after calling ASTPageFn(sess):
root := ASTPageFn(sess) // returns ast.Node (typed as any)
node, ok := root.(ast.Node)
if !ok {
    return ErrSchemaInvalid
}
schema, err := ast.CompileTree(node)
if err != nil {
    // err contains all validation failures — pipeline returns ErrSchemaInvalid
    return err
}
// schema is map[string]any — stored in DataKeySchema
```

---

## 7.5 Node Types Reference

All node types implemented in `internal/web/ast/`:

### Layout Nodes

| Go Type | AMIS type | Description |
|---------|-----------|-------------|
| `PageNode` | `"page"` | Root page container. Has `InitAPI`, `Title`, `Body`, `Toolbar` |
| `GridNode` | `"grid"` | N-column grid with `GridColumn` children |
| `FlexNode` | `"flex"` | Flexbox row/column container |
| `TabsNode` | `"tabs"` | Tabbed container. Uses `MountOnEnter`, `UnmountOnExit` |
| `SplitPaneNode` | `"splitpane"` | Resizable split pane |
| `SectionNode` | `"fieldset"` | Collapsible section with title |

### Data Display Nodes

| Go Type | AMIS type | Description |
|---------|-----------|-------------|
| `TableNode` | `"table"` | Static data table (no pagination). Use `CRUDNode` for server data |
| `CRUDNode` | `"crud"` | Paginated, filterable, sortable data table with server API |
| `ChartNode` | `"chart"` | ECharts chart. Requires transparent background in `Style` and `Config` |
| `CardNode` | `"card"` | Single record card display |
| `StatNode` | `"statistic"` | KPI number display with optional trend |
| `TimelineNode` | `"timeline"` | Chronological event list |
| `TreeNode` | `"tree"` | Hierarchical tree display |
| `MappingNode` | `"mapping"` | Value-to-label/HTML mapper (used for status badges) |
| `PropertyNode` | `"property"` | Key-value property list (used for detail cards) |
| `FormulaNode` | `"formula"` | Computed value display |

### Form Nodes

| Go Type | AMIS type | Description |
|---------|-----------|-------------|
| `FormNode` | `"form"` | Form container |
| `FilterBarNode` | `"filter"` | CRUD filter form |
| `InputTextNode` | `"input-text"` | Text input |
| `InputNumberNode` | `"input-number"` | Numeric input |
| `InputDateNode` | `"input-date"` | Date picker |
| `InputDateRangeNode` | `"input-date-range"` | Date range picker |
| `SelectNode` | `"select"` | Single or multi select |

### Interaction Nodes

| Go Type | AMIS type | Description |
|---------|-----------|-------------|
| `ActionNode` | `"button"` | Button/link/ajax action |
| `DialogNode` | `"dialog"` | Modal dialog (referenced by `ActionNode.Dialog`) |
| `DrawerNode` | `"drawer"` | Side drawer (referenced by `ActionNode.Drawer`) |

---

## 7.6 Struct Literal Construction

AST nodes are plain Go structs. Construction uses struct literals, not constructors or fluent builders:

```go
// PageNode
ast.PageNode{
    Title: "Finance Dashboard",
    InitAPI: &ast.APISpec{
        Method: "get",
        URL:    "/api/v1/finance/dashboard/summary",
    },
    Body: []ast.Node{
        ast.StatNode{Label: "Revenue", Value: "${revenue}"},
    },
}

// CRUDNode
ast.CRUDNode{
    API:        ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices"},
    PrimaryKey: "id",
    PageSize:   25,
    SyncLocation: false,
    Columns: []ast.TableColumn{
        {Name: "invoice_number", Label: "Invoice #", Sortable: true},
        {Name: "amount", Label: "Amount", Type: "currency"},
    },
    Toolbar: []ast.Node{
        ast.ActionNode{Label: "New", ActionType: "link", Target: "/finance/invoices/new", Level: "primary"},
    },
    RowActions: []ast.ActionNode{
        {Label: "View", ActionType: "link", Target: "/finance/invoices/${id}"},
    },
}

// GridNode
ast.GridNode{
    Columns: []ast.GridColumn{
        {Body: []ast.Node{chartNode}, MD: 8},
        {Body: []ast.Node{statNode}, MD: 4},
    },
}

// TabsNode
ast.TabsNode{
    MountOnEnter:   true,
    UnmountOnExit:  false,
    Tabs: []ast.TabItem{
        {Title: "Overview", Body: []ast.Node{summaryNode}},
        {Title: "Line Items", Body: []ast.Node{lineItemsNode}},
    },
}
```

There are no `NewPageNode()`, `NewCRUDNode()` constructors. Use struct literals. Zero values of optional fields are safe — the `Compile()` method omits zero-value fields.

---

## 7.7 APISpec: The Data Binding Type

`APISpec` describes an HTTP call that the AMIS SDK makes at runtime. It is used in `InitAPI`, `API`, `ActionNode.API`, and filter/action targets.

```go
type APISpec struct {
    Method  string            // "get", "post", "put", "delete", "patch"
    URL     string            // can contain AMIS expressions: "/api/invoices/${id}"
    Data    map[string]string // additional request body fields
    Headers map[string]string // additional headers
}
```

`PageNode.InitAPI` uses `*APISpec` (pointer) — nil means no initial data load. `CRUDNode.API` uses `APISpec` (value) — it is always required.

```go
// InitAPI is optional
ast.PageNode{
    InitAPI: nil, // no initial API call — page is purely static
}

// InitAPI with expression in URL
ast.PageNode{
    InitAPI: &ast.APISpec{
        Method: "get",
        URL:    "/api/v1/finance/invoices/${id}", // id from Params or page scope
    },
}
```

---

## 7.8 ActionNode: Actions and Dialogs

`ActionNode` is the most versatile node — it covers links, AJAX calls, dialogs, drawers, and URL navigation.

```go
type ActionNode struct {
    Label       string
    Level       string    // "primary"|"success"|"warning"|"danger"|"default"
    Icon        string    // CSS class e.g. "fa fa-plus"
    ActionType  string    // "link"|"ajax"|"dialog"|"drawer"|"url"|"copy"
    Target      string    // for "link" — internal route e.g. "/finance/invoices/${id}"
    URL         string    // for "url" — external URL
    API         *APISpec  // for "ajax" — the HTTP call to make
    Dialog      *DialogNode
    Drawer      *DrawerNode
    ConfirmText string    // confirmation prompt before executing
    DisabledOn  string    // AMIS expression: "${status !== 'DRAFT'}"
    VisibleOn   string    // AMIS expression: "${can_approve}"
}
```

**Dialog actions:** When `ActionType` is `"dialog"`, the `Dialog` field must be set. The `Dialog` node is included in `ActionNode.Children()` for validation.

```go
ast.ActionNode{
    Label:      "Approve",
    ActionType: "dialog",
    Level:      "success",
    Dialog: &ast.DialogNode{
        Title: "Confirm Approval",
        Body: []ast.Node{
            ast.FormNode{
                API: &ast.APISpec{Method: "post", URL: "/api/v1/finance/invoices/${id}/approve"},
                Body: []ast.Node{
                    ast.InputTextNode{Name: "note", Label: "Approval Note", Required: false},
                },
            },
        },
    },
}
```

**AMIS expression in conditions:** The AMIS SDK evaluates `DisabledOn` and `VisibleOn` at runtime against the current data scope:

```go
ast.ActionNode{
    Label:      "Approve",
    ActionType: "ajax",
    DisabledOn: "${!can_approve}",      // correct: "!" inside the expression
    VisibleOn:  "${status === 'PENDING'}",
}
```

> **Warning:** `"!${can_approve}"` is wrong — it produces the string `"!true"` or `"!false"`, not a boolean. The `!` must be inside `${}`.

---

## 7.9 Validation Errors

`Validate()` returns `*sharedErrors.BusinessError` with an `"AST_"` prefix code. Examples:

| Condition | Error Code |
|-----------|-----------|
| `CRUDNode` with no columns | `AST_CRUD_NO_COLUMNS` |
| `CRUDNode` with `SyncLocation: true` | `AST_CRUD_SYNC_LOCATION` |
| `ChartNode` without transparent background | `AST_CHART_OPAQUE_BG` |
| `TabsNode` with no tabs | `AST_TABS_EMPTY` |
| `FormNode` with no body fields | `AST_FORM_EMPTY` |
| `ActionNode` type "dialog" with no Dialog | `AST_ACTION_DIALOG_MISSING` |

All errors collected by `CompileTree` are surfaced as `VALIDATE_*` codes by `ValidateStage`. They are programmer errors, not user errors — the pipeline returns HTTP 500.

---

## 7.10 What the AST Does Not Model

The AST deliberately excludes certain concerns:

**CSS class names.** Nodes emit semantic AMIS types. The AMIS SDK applies its own CSS. Page functions do not set `className`, `style`, or color tokens directly (except for chart transparent backgrounds, which are an AMIS structural requirement).

**I18n keys.** There is no `i18n.Key()` type. Labels are Go strings. Translation is the responsibility of the page author or a future i18n layer.

**Slot composition.** There is no `SlotNode` or template reference type. Composition is `Children() []Node`.

> **Status: Planned — Not Yet Implemented**
> Template instantiation (`ast.NewTemplateRef`) and slot composition are planned. Until implemented, use function call composition in Go.

**AMIS expression validation.** String fields that contain `${...}` expressions (e.g., `DisabledOn`, `VisibleOn`, `Value`) are not syntax-checked by the AST. Expression correctness is the author's responsibility.

---

*End of Chapter 07*

**Previous:** [Chapter 06 — UI DSL Architecture](./06-ui-dsl-architecture.md)
**Next:** [Chapter 08 — The Compilation Pipeline](./08-compilation-pipeline.md)
