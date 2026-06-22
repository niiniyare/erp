---
title: "Chapter 17: REST API Conventions"
part: "Part III — The API Layer"
chapter: 17
section: "17-rest-api-conventions"
related:
  - "[Chapter 18: Error Handling](18-error-handling.md)"
  - "[Chapter 19: API Versioning](19-api-versioning.md)"
---

# Chapter 17: REST API Conventions

Awo auto-generates CRUD API routes from EntityDefinitions and uses a consistent response envelope throughout. Callers (especially the amis SDUI client) rely on this consistency for auto-wiring without custom glue code.

---

## 17.1. URL Structure

### 17.1.1. Auto-Generated CRUD Routes

Every registered EntityDefinition automatically gets:

```
GET    /api/v1/{entity-name}          → list (paginated)
POST   /api/v1/{entity-name}          → create
GET    /api/v1/{entity-name}/{id}     → get single
PUT    /api/v1/{entity-name}/{id}     → replace (full update)
PATCH  /api/v1/{entity-name}/{id}     → partial update
DELETE /api/v1/{entity-name}/{id}     → delete (or soft-delete)
```

Entity names in URLs are kebab-case: `invoice-items`, `purchase-orders`, `cost-centres`.

### 17.1.2. Entity Action Routes

Actions beyond CRUD follow the `/{id}/{action}` pattern:

```
POST /api/v1/invoices/{id}/submit
POST /api/v1/invoices/{id}/cancel
POST /api/v1/invoices/{id}/amend
POST /api/v1/purchase-orders/{id}/approve
POST /api/v1/purchase-orders/{id}/dispatch
```

Actions always use POST regardless of idempotency. GET for actions is wrong — GET must be safe (no side effects).

### 17.1.3. Tenant Context From Middleware, Not URL

The tenant is never in the URL path. `/api/v1/acme/invoices` is the wrong pattern. Tenant comes from the subdomain or `X-Tenant-ID` header. This:
- Keeps URLs tenant-agnostic (same URL pattern for all tenants)
- Prevents a user from accessing another tenant's data by changing the URL
- Simplifies API client code (no tenant URL segment to manage)

### 17.1.4. Custom Routes — Module-Specific Endpoints

```
/api/v1/finance/gl-postings
/api/v1/finance/period-close
/api/v1/crm/pipeline-summary
/api/v1/payroll/run/{id}/payslips
```

Module routes use a module prefix before the resource name. They are registered by module developers alongside standard entity routes.

---

## 17.2. Standard Response Envelope

### 17.2.1. Success Shape — Single Resource

```json
{
  "data": {
    "id": "550e8400-...",
    "invoice_number": "INV-2025-00042",
    "customer_id": "a1b2c3...",
    "total_amount": "45000.0000",
    "status": "submitted",
    "created_at": "2025-07-04T09:00:00Z",
    "_links": {
      "self": "/api/v1/invoices/550e8400-...",
      "customer": "/api/v1/customers/a1b2c3...",
      "items": "/api/v1/invoices/550e8400-.../items"
    }
  },
  "meta": {
    "request_id": "req-uuid"
  }
}
```

### 17.2.2. Paginated List Shape

```json
{
  "data": [ { ... }, { ... } ],
  "meta": {
    "total": 1247,
    "limit": 20,
    "cursor": "eyJpZCI6Ii4uLiJ9",
    "has_next": true,
    "request_id": "req-uuid"
  }
}
```

### 17.2.3. Error Shape

```json
{
  "status": 422,
  "code": "VALIDATION_ERROR",
  "message": "The request contains invalid data",
  "errors": [
    { "field": "email", "message": "must be a valid email address" },
    { "field": "phone", "message": "must be in E.164 format (+254...)" }
  ],
  "request_id": "req-uuid"
}
```

### 17.2.4. Why a Consistent Envelope Matters for amis

amis data source configuration maps API responses to UI components. With a consistent `data` key, all CRUD components work identically:

```json
{
  "type": "crud",
  "api": "GET /api/v1/invoices",
  "source": "$.data",
  "headerToolbar": ["bulkActions", "pagination"],
  "footerToolbar": ["statistics", "pagination"],
  "columns": [...]
}
```

amis reads `$.data` for records and `$.meta.total` for pagination — no custom adapters needed if every endpoint follows the same envelope.

---

## 17.3. Filtering

### 17.3.1. Query String Filter Convention

```
GET /api/v1/invoices?filter[status][eq]=submitted&filter[total_amount][gte]=10000
```

The `filter[field][operator]=value` format maps directly to the Filter DSL:

| Query param | Filter DSL |
|---|---|
| `filter[status][eq]=submitted` | `filter.Eq("status", "submitted")` |
| `filter[amount][gte]=1000` | `filter.Gte("amount", 1000)` |
| `filter[name][contains]=acme` | `filter.Contains("name", "acme")` |
| `filter[deleted_at][null]=true` | `filter.IsNull("deleted_at")` |

### 17.3.2. Logical Operators in Query Strings

```
# OR: invoices that are overdue OR submitted
filter[_or][0][status][eq]=overdue&filter[_or][1][status][eq]=submitted

# AND is implicit when multiple filter keys exist
filter[status][eq]=submitted&filter[customer_id][eq]=uuid
```

### 17.3.3. JSONB Path Filters for Custom Fields

```
GET /api/v1/invoices?filter[custom_fields.kra_customs_code][eq]=HS-8471
```

The dot-notation path is parsed and converted to a JSONB path predicate.

### 17.3.4. Full-Text Search

```
GET /api/v1/customers?q=acme+nairobi
```

The `q` parameter triggers a `pg_trgm` trigram search across indexed text fields. The entity must declare searchable fields:

```go
func (Customer) SearchFields() []string {
    return []string{"name", "email", "phone", "kra_pin"}
}
```

---

## 17.4. Sorting and Pagination

### 17.4.1. Sort Parameter

```
GET /api/v1/invoices?sort=total_amount,-created_at
```

Comma-separated fields. Prefix `-` for descending. Multiple fields applied left-to-right.

### 17.4.2. Cursor Pagination

```
# First page
GET /api/v1/invoices?limit=20

# Next page (cursor from previous meta.cursor)
GET /api/v1/invoices?limit=20&cursor=eyJpZCI6Ii4uLiJ9
```

Cursors are base64-encoded JSON containing the sort key values of the last seen record. They are opaque to callers — never parse or construct them manually.

### 17.4.3. Offset Pagination — Reports Only

```
GET /api/v1/invoices/export?offset=100&limit=500
```

Only available on export/report endpoints. Not available on standard list endpoints to prevent the "missing records" problem under concurrent inserts.

### 17.4.4. Default and Maximum Page Size

| Context | Default limit | Maximum limit |
|---|---|---|
| Standard list | 20 | 100 |
| Report export | 500 | 5000 |
| Bulk operation | 100 | 1000 |

Requests exceeding the maximum limit receive 400 with `LIMIT_TOO_LARGE`.

---

## 17.5. Bulk Operations

### 17.5.1. `POST /api/v1/{entity}/bulk-create`

```json
{
  "data": [
    { "name": "Product A", "sku": "SKU-001", "price": 1500.00 },
    { "name": "Product B", "sku": "SKU-002", "price": 2000.00 }
  ]
}
```

All records validated first. Any validation failure aborts the entire batch (no partial creates by default). Use `"partial": true` to allow partial success with per-record error reporting.

### 17.5.2. `POST /api/v1/{entity}/bulk-update`

```json
{
  "filter": { "status": { "eq": "draft" } },
  "patch": { "status": "submitted" }
}
```

**Warning**: bulk update bypasses per-record hooks. See §8.3.5.

### 17.5.3. Partial Success Handling

```json
{
  "data": {
    "created": 8,
    "failed": 2,
    "failures": [
      { "index": 3, "errors": [{ "field": "sku", "message": "SKU-005 already exists" }] },
      { "index": 7, "errors": [{ "field": "price", "message": "must be greater than 0" }] }
    ]
  }
}
```

### 17.5.4. Bulk Operation Size Limits

Maximum 1000 records per bulk-create/bulk-update. Larger imports must use the `BulkImport` Temporal workflow which handles batching, progress reporting, and partial failure recovery.

---

## 17.6. Idempotency

### 17.6.1. `Idempotency-Key` Header

```
POST /api/v1/invoices
Idempotency-Key: 7f3b2a9e-1234-5678-abcd-ef0123456789
Content-Type: application/json

{ ... invoice data ... }
```

The server stores the response (status code + body) keyed by `{tenant_id}:{idempotency_key}` in Redis for 24 hours. If the same key is used in a subsequent request (e.g. network retry), the server returns the cached response without re-processing.

### 17.6.2. How the Server Deduplicates

```go
func idempotency(redis rueidis.Client) fiber.Handler {
    return func(c *fiber.Ctx) error {
        key := c.Get("Idempotency-Key")
        if key == "" || c.Method() == "GET" {
            return c.Next()
        }

        cacheKey := fmt.Sprintf("idempotency:%s:%s",
            c.Locals("tenant_id"), key)

        cached, err := redis.Get(c.Context(), cacheKey)
        if err == nil {
            // Replay cached response
            var resp CachedResponse
            json.Unmarshal([]byte(cached), &resp)
            c.Set("X-Idempotency-Replayed", "true")
            return c.Status(resp.Status).JSON(resp.Body)
        }

        // Process the request
        err = c.Next()

        // Cache the response
        respBody := c.Response().Body()
        cached, _ = json.Marshal(CachedResponse{
            Status: c.Response().StatusCode(),
            Body:   respBody,
        })
        redis.Set(c.Context(), cacheKey, cached, 24*time.Hour)
        return err
    }
}
```

### 17.6.3. Idempotency Key TTL

24 hours. After 24 hours, the same key on the same endpoint triggers a new execution. For operations where idempotency must be guaranteed beyond 24 hours (e.g. payment submission), use Temporal workflow deduplication instead (workflow IDs provide permanent deduplication within the Temporal namespace).

### 17.6.4. Operations That Should Use Idempotency Keys

Any non-GET mutation that might be retried on network failure:
- Creating invoices (a network timeout might cause the client to retry and create a duplicate)
- Initiating payments
- Triggering approval workflows
- Sending notifications

GET, DELETE, and list operations do not need idempotency keys (GET is safe, DELETE is naturally idempotent, list results may legitimately change).
