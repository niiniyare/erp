# Response and Error Handling

> Last verified: 2026-05-18 | Code pointer: `internal/web/handler/schema.go`, `internal/web/stages/response.go`

---

## 📖 Why This Matters

AMIS is opinionated about response formats. Every API call the browser makes expects `{ "status": 0, "data": ... }` for success and `{ "status": 1, "msg": "..." }` for errors. A wrong status code or envelope shape produces silent failures — empty tables, blank forms, no error visible to the user. Getting the envelope right is table stakes.

---

## Schema Endpoint Response

Schema endpoints (`GET /schema/*`) always return AMIS envelope. SchemaHandler wraps the compiled schema:

```go
// internal/web/handler/schema.go:133
return c.JSON(fiber.Map{
    "status": 0,
    "data":   out.Schema,
})
```

The browser receives:
```json
{
  "status": 0,
  "data": {
    "type": "page",
    "title": "Invoices",
    "body": { "..." }
  }
}
```

AMIS reads `data` and renders it. The `status: 0` signals success.

**Critical:** Page functions (`PageFn` / `ASTPageFn`) return **raw schema** — not wrapped. The handler wraps it. Returning `{ "status": 0, "data": { ... } }` from a page function produces nested envelopes.

---

## Error Responses from the Schema Endpoint

Source: `internal/web/handler/schema.go:140–174`

| Error Type | HTTP Status | Response |
|-----------|------------|----------|
| Missing/invalid session | 401 | AMIS-renderable session-expired page (see below) |
| Route not found in registry | 404 | `{ "status": 404, "msg": "schema not found: /route" }` |
| Schema validation failure | 500 | `{ "status": 500, "msg": "An internal error occurred..." }` |
| Permission resolution failure (Casbin down) | 503 | `{ "status": 503, "msg": "service temporarily unavailable" }` |
| Unclassified pipeline error | 500 | `{ "status": 500, "msg": "An internal error occurred..." }` |

### 401 — Session Expired

When the session is missing or invalid, SchemaHandler returns an AMIS-renderable schema page instead of a plain JSON error. This lets the browser render a "Session expired" alert with a Login button rather than a blank screen:

```json
{
  "status": 401,
  "msg":    "Session expired. Please log in.",
  "data": {
    "type": "page",
    "body": {
      "type":  "alert",
      "body":  "Your session has expired.",
      "level": "warning"
    },
    "toolbar": [
      {
        "type":       "button",
        "label":      "Log In",
        "level":      "primary",
        "actionType": "url",
        "url":        "/login"
      }
    ]
  }
}
```

AMIS renders this as a page with a warning and a login button. No blank screen.

---

## Data API Response Formats

Data APIs (`/api/v1/*`) are separate from schema endpoints. They use `AmisResponse` directly.

### Success — list

```json
{
  "status": 0,
  "data": {
    "items": [ { "id": "...", "..." } ],
    "count": 142
  }
}
```

AMIS `crud` reads `items` for rows, `count` for pagination total.

### Success — single record

```json
{
  "status": 0,
  "data": {
    "id":     "uuid",
    "number": "INV-001",
    "..."
  }
}
```

Used by `form` with `initApi`.

### Success — mutation (create/update/action)

```json
{
  "status": 0,
  "msg":    "Invoice created",
  "data":   { "id": "new-uuid" }
}
```

AMIS merges `data` into the current data scope. Use `"redirect": "/invoices/${id}"` on the form to navigate to the new record.

### Validation failure

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

AMIS reads `errors` and renders field-level messages. The field `name` in the schema must exactly match the key in `errors`.

### Generic error

```json
{
  "status": 1,
  "msg": "Something went wrong"
}
```

AMIS shows `msg` as an error toast. Use `status: 1` for any non-success condition AMIS should surface to the user.

---

## Go Helper Types

Source: `internal/web/response/` (or inline in handlers)

```go
type AmisResponse struct {
    Status int    `json:"status"`
    Msg    string `json:"msg,omitempty"`
    Data   any    `json:"data,omitempty"`
}

func OK(data any) AmisResponse             { return AmisResponse{Status: 0, Data: data} }
func OKMsg(msg string, data any) AmisResponse { return AmisResponse{Status: 0, Msg: msg, Data: data} }
func Err(msg string) AmisResponse          { return AmisResponse{Status: 1, Msg: msg} }

type AmisValidationResponse struct {
    Status int               `json:"status"`
    Msg    string            `json:"msg"`
    Errors map[string]string `json:"errors,omitempty"`
}

func ValidationErr(msg string, fields map[string]string) AmisValidationResponse {
    return AmisValidationResponse{Status: 422, Msg: msg, Errors: fields}
}
```

**`AmisResponse` is for `/api/v1/` data endpoints ONLY.** Never return it from `PageFn` or `ASTPageFn`.

---

## ResponseStage (Pipeline Terminal)

Source: `internal/web/stages/response.go`

Priority 90. Always runs — whether schema came from cache or fresh compilation.

```go
func (s *ResponseStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
    schema, ok := opCtx.Data[ui.DataKeySchema].(ui.Schema)
    if !ok || len(schema) == 0 {
        return pipeline.StageResult{}, fmt.Errorf(
            "ui.response: DataKeySchema missing — pipeline stage ordering defect",
        )
    }

    cacheHit, _ := opCtx.Data[ui.DataKeyCacheHit].(bool)

    out := ui.UISchemaOutput{
        Schema:   schema,
        CacheHit: cacheHit,
        Route:    input.Route,
    }
    // Writes DataKeyResponse — SchemaHandler reads it after Run() returns.
}
```

If `DataKeySchema` is missing when `ResponseStage` executes, it returns an error — this surfaces stage ordering bugs immediately rather than silently returning an empty schema.

---

## OTel Tracing + Metrics

Source: `internal/web/handler/schema.go:73–130`

`SchemaHandler` instruments every request:

**Trace span** (`ui.schema_handler`):
```go
requestSpan.SetAttributes(
    attribute.String("ui.route",     route),
    attribute.String("ui.tenant_id", sc.TenantID().String()),
    attribute.String("ui.user_id",   sc.UserID().String()),
    attribute.Bool  ("ui.cache_hit", out.CacheHit),
)
```

**Counter** (`ui_schema_requests_total`):
```go
h.metrics.IncrementCounter("ui_schema_requests_total", metrics.Fields{
    "route":     route,
    "tenant_id": sc.TenantID().String(),
    "cache_hit": out.CacheHit,
})
```

Use `ui_schema_requests_total` + `cache_hit` dimension to monitor cache hit rate per route and tenant.

---

## Error Sentinel Reference

Source: `internal/web/ui/` (errors defined alongside pipeline types)

| Sentinel | Produced By | Meaning |
|----------|------------|---------|
| `ui.ErrUnauthenticated` | `SessionStage`, `SchemaHandler` | No valid session in context |
| `ui.ErrPermissionResolution` | `AuthzStage` | Casbin BulkEnforce failed |
| `ui.ErrPageNotFound` | `RegistryStage` | Route not in registry |
| `ui.ErrSchemaInvalid` | `ValidateStage` | Schema fails structural/security rules |

All are checked with `errors.Is()` in `SchemaHandler.handlePipelineError()`. Always wrap errors with `fmt.Errorf("%w: ...", sentinel)` to preserve the chain.

---

## Request Flow Diagram

```
GET /schema/finance/invoices
        │
        ▼
Authenticate middleware (401 if no session)
        │
contract.InjectSessionContext middleware
        │
SchemaHandler.Handle():
  contract.FromContext → session valid?
  ├── No  → return unauthenticatedEnvelope() HTTP 401
  └── Yes → start OTel span, build opCtx, Run(opCtx)
                │
                ▼
           pipeline.Run()
                │
           ┌────┴──────────────────────────────────────────────┐
           │ Error?                                             │
           │  ErrUnauthenticated  → HTTP 401                   │
           │  ErrPageNotFound     → HTTP 404                   │
           │  ErrSchemaInvalid    → HTTP 500                   │
           │  ErrPermissionRes.   → HTTP 503                   │
           │  other               → HTTP 500                   │
           └───────────────────────┬───────────────────────────┘
                                   │ success
                                   ▼
                      opCtx.Data[DataKeyResponse].(UISchemaOutput)
                                   │
                      emit metrics: ui_schema_requests_total
                      set span: cache_hit
                                   │
                                   ▼
                      c.JSON({"status": 0, "data": schema})
```
