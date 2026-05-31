[<-- Back to Index](README.md)

## API Contract

### The Two Envelopes

The Go backend and AMIS speak different formats. The `fetcher` bridges them. **Understanding both is essential.**

```markdown
GO BACKEND ENVELOPE:           AMIS ENVELOPE:
{                              {
  "success": true,               "status": 0,
  "data": [...],                 "msg": "",
  "meta": {                      "data": { ... }
    "pagination": {            }
      "total_records": 142,
      "offset": 0,
      "limit": 20
    }
  }
}
```

### Fetcher Translation

The custom fetcher in `index.html` handles the translation automatically:

```markdown
REQUEST TRANSLATION:
  AMIS sends:  ?page=2&perPage=20
  Fetcher sends: ?offset=20&limit=20

RESPONSE TRANSLATION (list):
  Backend returns: { success: true, data: [...], meta: { pagination: { total_records: 142 } } }
  Fetcher returns: { status: 0, data: { items: [...], count: 142 } }

RESPONSE TRANSLATION (error):
  Backend returns: { success: false, message: "Not found" }
  Fetcher returns: { status: 1, msg: "Not found" }

RESPONSE TRANSLATION (204 No Content):
  Fetcher returns: { status: 0, msg: '' }
```

### If You Standardise to AMIS Envelope in Go

> **⚠️ DATA API HANDLERS ONLY.** The `AmisResponse` pattern below applies to `/api/v1/` data endpoints only.
> **Do NOT return `AmisResponse` from page/schema builder functions** (`PageFn` / `ASTPageFn`).
> `SchemaHandler` already wraps the schema output in `{status:0, data:...}` — returning `AmisResponse`
> from a page function produces nested envelopes: `{status:0, data:{status:0, data:{...}}}`.
> See [Page Registration Pattern](../03-implementation/02-page-registration-pattern.md#return-value) for correct page function return.

When Go handlers return AMIS format directly, the fetcher passes it through unchanged:

```go
// awo/web/response/amis.go

type AmisResponse struct {
    Status int    `json:"status"`
    Msg    string `json:"msg,omitempty"`
    Data   any    `json:"data"`
}

func OK(data any) AmisResponse {
    return AmisResponse{Status: 0, Data: data}
}

func OKMsg(msg string, data any) AmisResponse {
    return AmisResponse{Status: 0, Msg: msg, Data: data}
}

func Err(msg string) AmisResponse {
    return AmisResponse{Status: 1, Msg: msg}
}
```

The fetcher detects which format it received (`json.status !== undefined && json.success === undefined`) and passes AMIS envelopes through without transformation.

### List Response (for `crud` component)

```json
{
  "status": 0,
  "data": {
    "items": [ { "id": "uuid", "reference": "PO-001", "total": 15000.00 } ],
    "count": 142
  }
}
```

AMIS `crud` automatically sends these query params:
- `page` — current page (1-based), fetcher converts to `offset`
- `perPage` — page size, fetcher converts to `limit`
- `orderBy` — column name
- `orderDir` — `asc` | `desc`
- Any filter field names defined in the `filter` section

### Single Record Response (for `form` with `initApi`)

```json
{
  "status": 0,
  "data": {
    "id":        "uuid",
    "reference": "PO-001",
    "supplier":  "Supplier Name",
    "total":     15000.00,
    "status":    "confirmed"
  }
}
```

### Mutation Response (create / update / action)

```json
{
  "status": 0,
  "msg":    "Purchase order created",
  "data":   { "id": "new-uuid" }
}
```

AMIS merges `data` into the current data domain. Use this to navigate to the new record: set `"redirect": "/purchasing/orders/${id}"` on the form.

### Validation Error Response

```json
{
  "status": 422,
  "msg":    "Validation failed",
  "errors": {
    "supplier_id": "Supplier is required",
    "line_items":  "At least one line item is required"
  }
}
```

AMIS reads `errors` and renders field-level messages. The field `name` in your schema must exactly match the key in `errors`.

```go
type AmisValidationResponse struct {
    Status int               `json:"status"`
    Msg    string            `json:"msg"`
    Errors map[string]string `json:"errors,omitempty"`
}

func ValidationErr(msg string, fieldErrors map[string]string) AmisValidationResponse {
    return AmisValidationResponse{Status: 422, Msg: msg, Errors: fieldErrors}
}
```

### Schema Response (from `/schema/*` endpoints)

Schema endpoints return raw AMIS JSON — **NOT wrapped in the envelope**:

```json
{
  "type": "page",
  "title": "Purchase Orders",
  "body": { "...": "..." }
}
```

Do not wrap schema responses in `{ status, data }`. AMIS renders the returned JSON directly.

### Go Handler Pattern

```go
func (h *PurchaseOrderHandler) List(c *fiber.Ctx) error {
    tenantID := middleware.TenantID(c)

    filter := procurement.ListFilter{
        Offset:   c.QueryInt("offset", 0),
        Limit:    c.QueryInt("limit", 20),
        OrderBy:  c.Query("orderBy", "created_at"),
        OrderDir: c.Query("orderDir", "desc"),
        Status:   c.Query("status"),
    }

    orders, total, err := h.svc.List(c.Context(), tenantID, filter)
    if err != nil {
        return c.JSON(response.Err(err.Error()))
    }

    return c.JSON(response.OK(map[string]any{
        "items": orders,
        "count": total,
    }))
}
```

---
