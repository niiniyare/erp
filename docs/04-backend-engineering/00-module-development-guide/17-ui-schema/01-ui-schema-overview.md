---
title: UI Schema Design (Server-Driven)
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Handler Layer](../07-handler-layer/01-handler-overview.md)"
  - "[amis Web Schemas](../05-frontend-engineering/02-amis-schemas/01-amis-schema-overview.md)"
  - "[Core Module Integrations](../07-core-integrations/01-core-integrations-overview.md)"
---

# UI Schema Design (Server-Driven)

AwoERP uses server-driven UI: the Go backend serves AMIS JSON schemas as HTTP responses rather than embedding them as static files. This lets schemas vary by tenant configuration, user roles, and feature flags at runtime.

## Why Server-Driven

| Static schema file | Server-driven schema |
|--------------------|---------------------|
| Same for all tenants | Tenant-specific fields, labels, options |
| Feature flags require rebuild | Flags reflected immediately |
| Role-based field hiding requires JS | Backend strips fields the user cannot see |
| Hard to test | Unit-testable Go code |

## Schema Endpoint Pattern

Every module exposes schema endpoints under its HTTP group:

```
GET /api/v1/{module}/schemas/{page}
```

| Parameter | Description |
|-----------|-------------|
| `module` | Module name (`contracts`, `finance`) |
| `page` | Page identifier (`list`, `detail`, `create`, `edit`) |

Examples:
```
GET /api/v1/contracts/schemas/list
GET /api/v1/contracts/schemas/detail
GET /api/v1/contracts/schemas/create
GET /api/v1/contracts/schemas/edit
```

The response is a raw AMIS JSON schema (no envelope wrapping):

```json
{
  "type": "page",
  "title": "Contracts",
  "body": { ... }
}
```

## Handler Registration

Schema endpoints are registered alongside data endpoints in `routes.go`:

```go
func (r *ContractRouteRegistrar) Register(api fiber.Router) {
    contracts := api.Group("/contracts", auth())

    // Data endpoints
    contracts.Get("",     authorize("contracts.contract.read"), r.handler.List)
    contracts.Post("",    authorize("contracts.contract.create"), r.handler.Create)
    contracts.Get("/:id", authorize("contracts.contract.read"), r.handler.GetByID)

    // Schema endpoints — same auth, no body, no DB call
    schemas := contracts.Group("/schemas", authenticate())
    schemas.Get("/list",   r.schemaHandler.List)
    schemas.Get("/detail", r.schemaHandler.Detail)
    schemas.Get("/create", r.schemaHandler.Create)
    schemas.Get("/edit",   r.schemaHandler.Edit)
}
```

## Schema Handler

The schema handler builds and returns AMIS JSON:

```go
// internal/core/contracts/handler/schema_handler.go
package handler

import (
    "github.com/gofiber/fiber/v2"
    "awo.so/internal/shared/middleware"
    "awo.so/internal/shared/amis"
)

type ContractSchemaHandler struct {
    authzSvc iam.AuthzService
}

func NewContractSchemaHandler(authz iam.AuthzService) *ContractSchemaHandler {
    return &ContractSchemaHandler{authzSvc: authz}
}

func (h *ContractSchemaHandler) List(c *fiber.Ctx) error {
    sess := middleware.SessionFrom(c)
    schema := buildListSchema(sess)
    return c.JSON(schema)
}

func (h *ContractSchemaHandler) Detail(c *fiber.Ctx) error {
    sess := middleware.SessionFrom(c)
    schema := buildDetailSchema(sess)
    return c.JSON(schema)
}
```

## Building Schemas with the amis Package

Use the `amis` builder package to construct schemas in Go rather than inline `map[string]any{}`:

```go
// internal/core/contracts/handler/schema_list.go
package handler

import (
    "awo.so/internal/shared/amis"
    "awo.so/internal/shared/iam"
)

func buildListSchema(sess iam.ResolvedSession) amis.Schema {
    canCreate := sess.HasPermission("contracts.contract.create")
    canDelete := sess.HasPermission("contracts.contract.delete")

    crud := amis.CRUD("/api/v1/contracts").
        Title("Contracts").
        Columns(
            amis.Column("contract_number", "Contract No.").Width(160),
            amis.Column("title", "Title"),
            amis.Column("status", "Status").Type("mapping").Map(contractStatusMap),
            amis.Column("total_value", "Value").Type("number").Prefix("$"),
            amis.Column("end_date", "Expires").Type("date"),
        ).
        Filter(
            amis.Select("status", "Status").Options(contractStatusOptions).Clearable(true),
            amis.TextInput("title", "Search title").Size("sm"),
        ).
        Toolbar(
            amis.If(canCreate, amis.CreateButton("Create Contract").Level("primary")),
        ).
        RowActions(
            amis.LinkButton("View", "/contracts/${id}"),
            amis.If(canDelete, amis.DeleteButton("Delete").ConfirmText("Delete this contract?")),
        )

    return amis.Page("Contracts").Body(crud).Build()
}
```

## Role-Based Schema Modification

Strip fields and actions the current user cannot use:

```go
func buildDetailSchema(sess iam.ResolvedSession) amis.Schema {
    canApprove  := sess.HasPermission("contracts.contract.approve")
    canTerminate := sess.HasPermission("contracts.contract.terminate")

    toolbar := amis.Toolbar()
    if canApprove {
        toolbar.Add(
            amis.DialogButton("Approve").
                Level("success").
                VisibleOn("${status === 'under_review'}").
                Dialog(buildApproveDialog()),
        )
    }
    if canTerminate {
        toolbar.Add(
            amis.DialogButton("Terminate").
                Level("danger").
                VisibleOn("${status === 'active'}").
                Dialog(buildTerminateDialog()),
        )
    }

    return amis.Page("Contract Detail").
        Body(buildDetailPanel(), toolbar).
        Build()
}
```

## Feature Flag Schema Variation

```go
func buildCreateSchema(sess iam.ResolvedSession) amis.Schema {
    form := amis.Form("POST", "/api/v1/contracts").
        Fields(
            amis.TextInput("contract_number", "Contract No.").Required(true),
            amis.TextInput("title", "Title").Required(true),
            amis.Select("contract_type", "Type").Options(contractTypeOptions).Required(true),
            amis.NumberInput("total_value", "Value").Required(true).Min(0),
        )

    // Bulk import tab only shown when feature flag is enabled
    if sess.FeatureFlags["contracts.bulk_import"] {
        form.AddField(
            amis.FileInput("import_file", "Bulk Import CSV").
                HideOn("${mode !== 'bulk'}"),
        )
    }

    return amis.Page("New Contract").Body(form).Build()
}
```

## Static Schema Fallback

For schemas that don't need dynamic content, embed as `json.RawMessage`:

```go
//go:embed schemas/contracts_list.json
var contractsListSchema []byte

func (h *ContractSchemaHandler) ListStatic(c *fiber.Ctx) error {
    c.Set("Content-Type", "application/json")
    return c.Send(contractsListSchema)
}
```

Use static schemas only when the schema is genuinely identical for all tenants and roles. When in doubt, use the builder.

## Schema Caching

Schema responses are deterministic given a session's permissions and feature flags. Cache at the handler level with a short TTL:

```go
func (h *ContractSchemaHandler) List(c *fiber.Ctx) error {
    sess := middleware.SessionFrom(c)

    // Cache keyed by role set + feature flags hash
    cacheKey := fmt.Sprintf("schema:contracts:list:%s", sess.CacheKey())
    if cached, ok := h.cache.Get(cacheKey); ok {
        c.Set("Content-Type", "application/json")
        return c.Send(cached.([]byte))
    }

    schema := buildListSchema(sess)
    b, _ := json.Marshal(schema)
    h.cache.Set(cacheKey, b, 5*time.Minute)
    return c.JSON(schema)
}
```

Cache TTL of 5 minutes is the default. On role or feature flag change, the cache naturally expires. Do not cache schemas that contain user-specific data (e.g., pre-filled values from DB).

## Wire Registration

```go
// internal/core/contracts/wire.go
var ContractsSet = wire.NewSet(
    // ...
    handler.NewContractSchemaHandler,
)
```

No additional dependencies beyond what the data handler already uses.

## Schema URL Convention

| Page | URL | Description |
|------|-----|-------------|
| List | `/api/v1/contracts/schemas/list` | CRUD list with filters |
| Detail | `/api/v1/contracts/schemas/detail` | Read-only detail view |
| Create | `/api/v1/contracts/schemas/create` | Create form |
| Edit | `/api/v1/contracts/schemas/edit` | Edit form |

The AMIS `app` or `nav` schema loads these URLs dynamically via the `service` component:

```json
{
  "type": "service",
  "api": "/api/v1/contracts/schemas/list",
  "body": "${body}"
}
```
