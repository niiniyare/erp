# API Contracts

> Last verified: 2026-05-18 | Code pointer: `internal/web/handler/schema.go`, `internal/web/stages/`

Two API boundaries:
1. **Schema API** — Go → Browser (what the shell fetches to render a page)
2. **Backend data API** — AMIS → Go (what AMIS calls for CRUD, forms, select options)

---

## 1. Schema API (Go → Browser)

Served by `SchemaHandler`. Requires authenticated session.

### Endpoint

```
GET /schema/*
```

Path after `/schema` becomes the route key, e.g.:
- `GET /schema/finance/invoices` → route `/finance/invoices`
- `GET /schema/dashboard` → route `/dashboard`

### Authentication

Requires `contract.InjectSessionContext()` middleware on the route. Returns 401 if session is missing or zero.

### Response: Success

```json
{
  "status": 0,
  "data": { ... AMIS schema ... }
}
```

`data` is the raw compiled AMIS schema (usually `{"type": "page", ...}`).

### Response: Session Expired (401)

Returns an **AMIS-renderable envelope** so the browser can show a "Log In" button instead of a blank screen:

```json
{
  "status": 401,
  "msg": "Session expired. Please log in.",
  "data": {
    "type": "page",
    "body": { "type": "alert", "body": "Your session has expired.", "level": "warning" },
    "toolbar": [{ "type": "button", "label": "Log In", "level": "primary", "actionType": "url", "url": "/login" }]
  }
}
```

### Response: Not Found (404)

```json
{
  "status": 404,
  "msg": "schema not found: /finance/invoices"
}
```

Cause: route not registered, or page package not imported (blank import missing).

### Response: Internal Error (500)

```json
{
  "status": 500,
  "msg": "An internal error occurred. Please try again."
}
```

Cause: schema validation failure (`ErrSchemaInvalid`) or unclassified pipeline error.

### Response: Service Unavailable (503)

```json
{
  "status": 503,
  "msg": "service temporarily unavailable"
}
```

Cause: IAM permission resolution failed (`ErrPermissionResolution`). Usually Casbin or Redis unavailable.

### Error Mapping Table

| Pipeline error | HTTP status | Response body |
|---|---|---|
| `ErrPageNotFound` | 404 | `{"status":404,"msg":"schema not found: ..."}` |
| `ErrUnauthenticated` | 401 | AMIS envelope with "Log In" button |
| `ErrSchemaInvalid` | 500 | Generic 500 envelope |
| `ErrPermissionResolution` | 503 | `{"status":503,"msg":"service temporarily unavailable"}` |
| Unclassified | 500 | Generic 500 envelope |

### OTel Span Attributes

| Attribute | Value |
|---|---|
| `ui.route` | `/finance/invoices` |
| `ui.tenant_id` | UUID |
| `ui.user_id` | UUID |
| `ui.cache_hit` | `true` / `false` |

### Metrics

Counter: `ui_schema_requests_total` with labels `route`, `tenant_id`, `cache_hit`.

---

## 2. Backend Data API (AMIS → Go)

These are the patterns AMIS uses to call backend endpoints. The fetcher in `index.html` translates AMIS pagination params and normalizes the response envelope.

### General Contract

All backend endpoints return:

```json
{
  "success": true,
  "data": ...,
  "message": "...",
  "meta": { ... }
}
```

The fetcher maps this to AMIS format before AMIS processes it.

### List Response (paginated)

```
GET /api/v1/{module}/{resource}?offset=0&limit=20&...filters
```

```json
{
  "success": true,
  "data": [
    { "id": "uuid", "field": "value", ... },
    ...
  ],
  "meta": {
    "pagination": {
      "total_records": 142,
      "offset": 0,
      "limit": 20
    }
  }
}
```

Fetcher maps to AMIS as:
```json
{ "status": 0, "data": { "items": [...], "count": 142 } }
```

AMIS CRUD configuration must use `itemsMapping: "items"` and `countMapping: "count"` (or the default keys if matching).

**Pagination translation:** AMIS sends `page` + `perPage`; fetcher converts to `offset` + `limit` before the request reaches the backend.

### Single Record Response

```
GET /api/v1/{module}/{resource}/{id}
```

```json
{
  "success": true,
  "data": { "id": "uuid", "field": "value", ... }
}
```

Fetcher maps to AMIS as:
```json
{ "status": 0, "data": { "id": "uuid", "field": "value", ... } }
```

Used by form `initApi` to populate field values.

### Create / Update Response

```
POST /api/v1/{module}/{resource}
PUT  /api/v1/{module}/{resource}/{id}
```

```json
{
  "success": true,
  "data": { "id": "uuid" },
  "message": "Invoice created"
}
```

Fetcher maps to AMIS as:
```json
{ "status": 0, "data": { "id": "uuid" }, "msg": "Invoice created" }
```

AMIS uses `data.id` for the `redirect` template token: `"redirect": "#invoices/${id}"`.

### Delete Response

```
DELETE /api/v1/{module}/{resource}/{id}
```

Backend should return **204 No Content** for delete success. Fetcher returns `{ "status": 0, "msg": "" }` for 204 — AMIS treats this as success.

Alternatively, a 200 with body:
```json
{ "success": true, "data": null, "message": "Deleted" }
```

### Error Response

```json
{
  "success": false,
  "message": "Invoice not found",
  "data": null
}
```

Fetcher maps to AMIS as:
```json
{ "status": 1, "msg": "Invoice not found" }
```

AMIS displays `msg` in a toast notification.

### Select Options Response

```
GET /api/v1/{module}/{resource}/options
```

```json
{
  "success": true,
  "data": [
    { "label": "Supplier A", "value": "uuid-1" },
    { "label": "Supplier B", "value": "uuid-2" }
  ]
}
```

Fetcher maps to AMIS as:
```json
{ "status": 0, "data": { "options": [...] } }
```

`SelectAPIField` / `ast.SelectNode` uses `source` pointing to this endpoint. AMIS expects `label` and `value` keys (configurable via `labelField` / `valueField`).

**Do not** use a paginated list endpoint as a select source — list responses have `items` + `count`, not `options`.

### File Upload Response

```
POST /api/v1/files/upload
```

```json
{
  "success": true,
  "data": { "value": "https://storage.example.com/files/abc123.pdf" }
}
```

AMIS `input-file` / `input-image` reads `data.value` as the stored file URL.

---

## 3. AMIS Pagination → Backend Translation

The fetcher automatically converts AMIS pagination params before the request leaves the browser:

| AMIS sends | Backend receives |
|---|---|
| `?page=1&perPage=20` | `?offset=0&limit=20` |
| `?page=2&perPage=20` | `?offset=20&limit=20` |
| `?page=3&perPage=50` | `?offset=100&limit=50` |

Formula: `offset = (page - 1) * limit`

This translation is in `fetcher` in `web/pages/index.html`. Backend endpoints must accept `offset` + `limit`, not `page` + `perPage`.

---

## 4. Common Contract Mistakes

| Mistake | Symptom | Fix |
|---|---|---|
| List endpoint returns `options` key | Select field shows list data | Use separate `/options` endpoint for selects |
| Delete returns 200 `{success:true}` with no body | AMIS reports "undefined" success | Return 204, or include `"data": null` |
| Create response omits `data.id` | `redirect` template token `${id}` is empty | Always return `"data": { "id": "..." }` on create |
| Backend returns `page`/`perPage` in response | Pagination loop | Backend ignores pagination in response; use `meta.pagination` |
| File upload returns URL at root level | File upload silently fails | URL must be at `data.value`, not top-level |
| Paginated endpoint used as select source | Select shows empty or broken options | Create a dedicated `/options` endpoint |
