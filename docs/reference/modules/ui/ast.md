# Typed UI AST

The `internal/web/ast` package defines every AMIS schema node as a typed Go
struct. Schemas are never constructed as `map[string]any` in application code —
the AST is the only construction path.

---

## Node Interface

```go
type Node interface {
    NodeType() string           // AMIS "type" field
    Validate() error            // check invariants — called by CompileTree
    Compile() map[string]any    // emit AMIS-compatible map
}

type ContainerNode interface {
    Node
    Children() []Node           // for recursive validation
}
```

`CompileTree(root Node)` is the only entry point for converting a node tree to
an AMIS schema. It:
1. Walks the tree via `ContainerNode.Children()`
2. Calls `Validate()` on every node — collects all errors before emitting JSON
3. Calls `Compile()` on the root only when all nodes pass validation

---

## Node Catalogue

### Layout Nodes

| Type | AMIS type | Use |
|------|-----------|-----|
| `PageNode` | `page` | Root of every page schema |
| `GridNode` | `grid` | Multi-column layout |
| `FlexNode` | `flex` | Flexbox row/column |
| `TabsNode` | `tabs` | Tabbed content |
| `SectionNode` | `collapse` | Collapsible labelled group |
| `SplitPaneNode` | `grid` (2-col) | Master-detail split view |

### Data Nodes

| Type | AMIS type | Use |
|------|-----------|-----|
| `CRUDNode` | `crud` | Paginated, filterable listing |
| `TableNode` | `table` | Static or API-sourced table |
| `ChartNode` | `chart` | ECharts chart |
| `StatNode` | `tpl` | KPI metric card |
| `TimelineNode` | `timeline` | Chronological event list |
| `TreeNode` | `tree` | Collapsible tree view |
| `CardNode` | `panel` | Card with header + body |

### Form Nodes

| Type | AMIS type | Use |
|------|-----------|-----|
| `FormNode` | `form` | Full data-entry form |
| `FilterBarNode` | `form` | Filter header for listing pages |
| `InputTextNode` | `input-text` | Text field |
| `InputNumberNode` | `input-number` | Numeric field |
| `InputDateNode` | `input-date` | Date picker |
| `InputDateRangeNode` | `input-date-range` | Date range picker |
| `SelectNode` | `select` | Dropdown |
| `MultiSelectNode` | `select` (multiple) | Multi-select dropdown |
| `ComboNode` | `combo` | Repeatable field group (line items) |
| `CheckboxNode` | `checkbox` | Boolean checkbox |
| `ToggleNode` | `switch` | Toggle switch |

### Action / Dialog Nodes

| Type | AMIS type | Use |
|------|-----------|-----|
| `ActionNode` | `button` | Button or link action |
| `DialogNode` | `dialog` | Modal dialog |
| `DrawerNode` | `drawer` | Side-drawer panel |

---

## Structural Invariants

Some invariants are enforced at compile time (in `Compile()`) rather than at
validation time, because they must be unconditionally true for AMIS correctness:

| Node | Invariant | Why |
|------|-----------|-----|
| `CRUDNode` | `syncLocation: true` always emitted | AMIS requires this for URL sync |
| `ChartNode` | `style.background: "transparent"` always emitted | Dark-mode compatibility |

ValidateStage does **not** check these for AST-compiled schemas — they are
guaranteed by the type system.

---

## APISpec

`APISpec` describes an AMIS API call. All API-sourced nodes use this type.

```go
type APISpec struct {
    Method  string            // "get" | "post" | "put" | "patch" | "delete"
    URL     string            // e.g. "/api/v1/finance/invoices"
    Headers map[string]string // optional
    Data    map[string]any    // merged into request body (POST/PUT)
    SendOn  string            // AMIS expression — skips call when false
}
```

`APISpec.Compile()` emits the AMIS object form when headers/data/sendOn are
set, otherwise the compact `"method:url"` string.

---

## Writing a New Node

1. Define a struct in `internal/web/ast/` — pick the file for its family
   (layout, form, display).
2. Implement `NodeType()`, `Validate()`, `Compile()`.
3. If it contains child nodes, implement `Children() []Node` and embed
   `ContainerNode`.
4. Add `var _ Node = MyNode{}` compile-time assertion.
5. Use value receivers throughout — pointer receivers allow mutation through
   an interface, which breaks the immutability contract.
