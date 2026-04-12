[<-- Back to Index](README.md)

## Go Schema Builder

### Why a Typed Builder Package

Raw `map[string]any` chains are unreadable and have no compile-time checks. JSON string literals in Go source are worse. The `awo/web/schema` package provides typed Go structs that marshal to valid AMIS JSON.

```markdown
WITHOUT builder (unreadable):
  map[string]any{"type":"crud","api":"get:/api/v1/orders","columns":[]any{map[string]any{"name":"id"...}}}

WITH builder (clear):
  schema.CRUD{API: "get:/api/v1/orders", Columns: []schema.Column{{Name:"id", Label:"ID"}}}
```

### Package Structure

```
awo/web/schema/
├── types.go         — All AMIS JSON types as Go structs
├── builder.go       — Fluent helper constructors
├── page.go          — Page / toolbar helpers
├── crud.go          — CRUD builder helpers
├── form.go          — Form builder helpers
├── app.go           — App shell builder (for future Go-driven nav)
├── accounting/
│   ├── journal.go
│   └── accounts.go
├── purchasing/
│   └── orders.go
└── settings/
    └── rules.go
```

### Core Types

```go
// awo/web/schema/types.go

type Schema map[string]any

type Page struct {
    Type     string `json:"type"`
    Title    string `json:"title,omitempty"`
    SubTitle string `json:"subTitle,omitempty"`
    Data     any    `json:"data,omitempty"`
    Toolbar  []any  `json:"toolbar,omitempty"`
    Body     any    `json:"body"`
}

type CRUDSchema struct {
    Type          string   `json:"type"`
    ID            string   `json:"id,omitempty"`
    API           string   `json:"api"`
    SyncLocation  bool     `json:"syncLocation"`
    Filter        *Form    `json:"filter,omitempty"`
    HeaderToolbar []any    `json:"headerToolbar,omitempty"`
    Columns       []Column `json:"columns"`
    BulkActions   []any    `json:"bulkActions,omitempty"`
    FooterToolbar []any    `json:"footerToolbar,omitempty"`
}

type Column struct {
    Name      string `json:"name"`
    Label     string `json:"label"`
    Type      string `json:"type,omitempty"`
    Sortable  bool   `json:"sortable,omitempty"`
    Prefix    string `json:"prefix,omitempty"`
    ColorMap  any    `json:"colorMap,omitempty"`
    Buttons   []any  `json:"buttons,omitempty"`
    VisibleOn string `json:"visibleOn,omitempty"`
    QuickEdit any    `json:"quickEdit,omitempty"`
}

type Form struct {
    Type     string `json:"type"`
    Title    string `json:"title,omitempty"`
    API      string `json:"api,omitempty"`
    InitAPI  string `json:"initApi,omitempty"`
    Redirect string `json:"redirect,omitempty"`
    Body     []any  `json:"body"`
    Actions  []any  `json:"actions,omitempty"`
    Mode     string `json:"mode,omitempty"`
}

type Button struct {
    Type        string `json:"type"`
    Label       string `json:"label"`
    Level       string `json:"level,omitempty"`
    ActionType  string `json:"actionType,omitempty"`
    Link        string `json:"link,omitempty"`
    API         string `json:"api,omitempty"`
    ConfirmText string `json:"confirmText,omitempty"`
    VisibleOn   string `json:"visibleOn,omitempty"`
    DisabledOn  string `json:"disabledOn,omitempty"`
    OnEvent     any    `json:"onEvent,omitempty"`
    Dialog      any    `json:"dialog,omitempty"`
    Drawer      any    `json:"drawer,omitempty"`
}
```

### Module Schema Example

```go
// awo/web/schema/purchasing/orders.go

func OrdersListPage(fl map[string]bool, perms Permissions) schema.Page {
    return schema.Page{
        Type:     "page",
        Title:    "Purchase Orders",
        SubTitle: "Manage supplier purchase orders",
        Data: map[string]any{
            "can_create":  perms.CanCreate,
            "can_approve": perms.CanApprove,
        },
        Toolbar: []any{
            schema.Button{
                Type: "button", Label: "New Order",
                Level: "primary", ActionType: "link",
                Link: "/purchasing/orders/new",
                VisibleOn: "${can_create}",
            },
        },
        Body: schema.CRUDSchema{
            Type:         "crud",
            ID:           "po-list",
            API:          "get:/api/v1/purchase-orders",
            SyncLocation: true,
            Columns:      orderColumns(fl),
        },
    }
}

func orderColumns(fl map[string]bool) []schema.Column {
    cols := []schema.Column{
        {Name: "reference",     Label: "Reference",  Sortable: true},
        {Name: "supplier_name", Label: "Supplier",   Sortable: true},
        {Name: "total",         Label: "Total",      Type: "number",
         Prefix: "KES ", Sortable: true},
        {Name: "status",        Label: "Status",     Type: "tag",
         ColorMap: map[string]string{
            "draft":     "default",
            "submitted": "processing",
            "confirmed": "success",
            "cancelled": "error",
         }},
    }

    // Feature-flag conditional column
    if fl["manufacturing.mrp_enabled"] {
        cols = append(cols, schema.Column{
            Name: "mrp_order_id", Label: "MRP Order",
            VisibleOn: "${mrp_order_id !== null}",
        })
    }

    return cols
}
```

### Schema Handler

```go
// awo/web/handlers/schema/purchasing.go

func (h *SchemaHandlers) PurchasingOrders(c *fiber.Ctx) error {
    fl   := middleware.ContextFlags(c)
    user := middleware.ContextUser(c)

    canCreate  := h.access.Can(c.Context(), user.ID, "create",  "purchase_orders")
    canApprove := h.access.Can(c.Context(), user.ID, "approve", "purchase_orders")

    page := purchasingschema.OrdersListPage(fl, purchasingschema.Permissions{
        CanCreate:  canCreate,
        CanApprove: canApprove,
    })
    return c.JSON(page)
}
```

### Mixing Typed Structs and Maps

For rarely used one-off schema constructs, `map[string]any` is fine inline. Do not create a typed struct for every AMIS component — only the ones used frequently across modules deserve a type.

```go
// A one-off inline map is fine for a simple inline alert
body := map[string]any{
    "type":  "alert",
    "level": "info",
    "body":  "This module is not enabled for your organisation.",
}
```

---
