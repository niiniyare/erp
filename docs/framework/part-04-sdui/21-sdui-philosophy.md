---
title: "Chapter 21: Server-Driven UI Philosophy"
part: "Part IV — The SDUI Layer"
chapter: 21
section: "21-sdui-philosophy"
related:
  - "[Chapter 22: Foundation Components](22-foundation-components.md)"
  - "[Chapter 16: RBAC](../part-03-api/16-rbac.md)"
  - "[Chapter 20: Feature Flags](../part-03-api/20-feature-flags.md)"
---

# Chapter 21: Server-Driven UI Philosophy

Server-Driven UI (SDUI) inverts the traditional SPA architecture: the server generates a JSON description of the UI, and the client renders it. Awo uses amis — an open-source low-code framework from Baidu — as its SDUI renderer. This chapter explains why, how the pipeline works end-to-end, and how to write page builder functions.

---

## 21.1. Why SDUI for an ERP Framework

### 21.1.1. The Frontend Team Bottleneck Problem

Traditional ERP development bottleneck: every time a new entity is added, a new form is designed, a PR is raised, frontend and backend review, QA tests, deploy. For a framework with dozens of entity types per module, this cycle is unacceptably slow.

SDUI breaks the bottleneck: the backend developer writes a page builder function alongside the entity definition. The UI is generated on the server and rendered by the client without a separate frontend build or deploy. Adding a new field to the entity adds it to the form automatically.

### 21.1.2. How amis JSON Becomes a Rendered UI

amis is a React-based renderer that takes a JSON schema and renders a complete UI. The JSON schema is not a simplified subset — it can express full CRUD lists, multi-tab forms, charts, dashboards, kanban boards, and wizard flows. No JavaScript is written by Awo application developers for standard ERP views.

```
Server returns JSON:
{
  "type": "page",
  "body": {
    "type": "crud",
    "api": "GET /api/v1/invoices",
    "columns": [...]
  }
}

→ amis renders a full, paginated, filterable, sortable data table
→ No React components written
→ No build step
→ No npm install
```

### 21.1.3. The SDUI Contract

**Server guarantees:**
- The JSON schema is valid amis schema (validated before caching)
- All API endpoints referenced in the schema exist and return the documented envelope
- Permission-gated UI elements are absent from the schema (not just disabled)
- Schema version is compatible with the pinned amis version

**Client can rely on:**
- `type`, `api`, `columns`, `body` keys are stable within a major version
- Error responses from referenced APIs follow the standard error envelope
- The schema will not reference removed API endpoints

### 21.1.4. When SDUI Is Appropriate and When It Is Not

**SDUI is ideal for:**
- CRUD forms and lists (90% of ERP UI)
- Dashboards with standard chart types
- Multi-step approval flows
- Report output pages

**SDUI is not ideal for:**
- Highly interactive real-time UI (chat, live tracking, streaming updates)
- Pixel-perfect custom branded experiences for customer-facing portals
- Complex drag-and-drop interfaces with custom business logic
- Mobile-native applications

For these cases, build a custom frontend that consumes the Awo REST API directly.

### 21.1.5. amis Version Pinning

The amis SDK is pinned to a specific version in `web/sdk/`. Never auto-update amis. A new amis version may change component behaviour, rename JSON keys, or remove deprecated features — all of which break existing page builder output.

Update amis version only after:
1. Testing all existing page builder outputs against the new version
2. Updating page builders that use removed or changed features
3. Running the full SDUI snapshot test suite
4. Staging deployment with QA sign-off

---

## 21.2. The Page Builder Pipeline

### 21.2.1. Step 1: Request Arrives

```
GET /api/v1/pages/invoice-list
Cookie: awo_session=...
Host: acme.awo.app
```

### 21.2.2. Step 2: Context Resolution

The full middleware pipeline runs: tenant resolution, session validation, permission evaluation, feature flag loading. The page builder receives a fully resolved context — no additional DB calls needed for tenant/user/flag data.

### 21.2.3. Step 3: Page Builder Function Invoked

```go
handler.BuildPage("invoice-list", pageBuilders["invoice-list"])

func BuildInvoiceListPage(ctx PageBuilderContext) (*amis.Schema, error) {
    // ctx contains: Tenant, Actor, Permissions, Flags
    // Pure function: same ctx = same output (for caching to work)
    return buildInvoiceList(ctx), nil
}
```

### 21.2.4. Step 4: amis JSON Assembled

The page builder function assembles the amis JSON using the Awo amis builder library:

```go
func buildInvoiceList(ctx PageBuilderContext) *amis.Schema {
    crud := amis.NewCRUD().
        API("GET /api/v1/invoices").
        PrimaryField("id").
        DefaultParams(map[string]interface{}{
            "sort": "-created_at",
        })

    crud.AddColumn(amis.Column("invoice_number").Label("Invoice #").Sortable(true))
    crud.AddColumn(amis.Column("customer_id").Label("Customer").
        Type("link").
        Href("/customers/${customer_id}").
        LabelField("customer_name"))
    crud.AddColumn(amis.Column("total_amount").Label("Amount (KES)").
        Type("number").
        Precision(2).
        ThousandSeparator(true))
    crud.AddColumn(amis.Column("status").Label("Status").
        Type("tag").
        Map(invoiceStatusColors()))

    if ctx.Permissions.CanCreate("Invoice") {
        crud.AddToolbar(amis.NewButton("New Invoice").
            ActionType("link").Href("/invoices/new"))
    }

    return amis.NewPage().Body(crud).Schema()
}
```

### 21.2.5. Step 5: Cached in Redis

The assembled JSON is cached:
```
Key: page:{tenant_id}:{actor_role_hash}:{page_name}
TTL: 5 minutes (configurable per page)
```

Cache is per-tenant and per-role-set. Two users with different roles get different cached page schemas.

### 21.2.6. Step 6: JSON Returned to Client

```json
{
  "data": {
    "type": "page",
    "body": { "type": "crud", "api": "...", "columns": [...] }
  },
  "meta": { "request_id": "..." }
}
```

The amis client at `web/pages/index.html` receives the schema and renders it immediately.

### 21.2.7. Cache Invalidation

Page caches are invalidated when:
- A tenant's permissions change (role assignment changed)
- A feature flag changes for the tenant
- A tenant's custom field definitions change (form layout changes)
- Explicit invalidation via `awo cache invalidate --page=invoice-list --tenant=uuid`

---

## 21.3. Page Builder Functions

### 21.3.1. Function Signature

```go
type PageBuilderFunc func(ctx PageBuilderContext) (*amis.Schema, error)

type PageBuilderContext struct {
    Context     context.Context
    Tenant      *tenant.Context
    Actor       *actor.Actor
    Permissions *rbac.Permissions
    Flags       featureflags.FlagSet
    Request     *PageRequest  // path params, query params from the page request
}
```

### 21.3.2. Naming Convention

```
Build{Entity}{View}Page

BuildInvoiceListPage
BuildInvoiceDetailPage
BuildInvoiceCreatePage
BuildCustomerDashboardPage
BuildPayrollRunWizardPage
```

### 21.3.3. Page Builders Contain Presentation Logic Only

Page builders must not contain business logic. They should:
- Decide which fields to show (based on permissions and flags)
- Choose layout (tabs vs sections)
- Configure column formatting
- Set up filter presets
- Wire action buttons

Page builders must NOT:
- Calculate totals
- Make business rule decisions
- Query the database directly
- Call external services

If a page builder needs data that affects layout (e.g. "how many custom fields exist for this entity"), that data comes from the `PageBuilderContext` — pre-fetched by the framework.

### 21.3.4. Registering a Page Builder

```go
func (Invoice) Pages() map[string]sdui.PageBuilderFunc {
    return map[string]sdui.PageBuilderFunc{
        "invoice-list":   BuildInvoiceListPage,
        "invoice-detail": BuildInvoiceDetailPage,
        "invoice-create": BuildInvoiceCreatePage,
    }
}
```

The framework routes `/api/v1/pages/invoice-list` to `BuildInvoiceListPage` automatically.

### 21.3.5. Testing Page Builders — Snapshot Tests

```go
func TestBuildInvoiceListPage(t *testing.T) {
    ctx := pagetest.NewContext().
        WithRole("finance_manager").
        WithFlag("advanced_reporting", true)

    schema, err := BuildInvoiceListPage(ctx)
    require.NoError(t, err)

    // Snapshot test: compare JSON output to golden file
    pagetest.AssertMatchesSnapshot(t, "invoice-list-finance-manager", schema)
}
```

Snapshot tests catch unintended changes to page builder output. They fail when a page builder change alters the schema — requiring explicit snapshot update with `go test -update-snapshots`.
