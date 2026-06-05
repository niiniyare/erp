---
title: "Error Handling Strategy"
volume: "V — Runtime & State"
chapter: "24-B"
phase: 3
status: draft
audience: [All Engineers]
---

# Chapter 24-B — Error Handling Strategy

> **Volume:** V — Runtime & State
> **Audience:** All Engineers
> **Prerequisites:** Chapter 24 — Data Sources, Chapter 25 — API Contracts
> **Phase:** 3

---

## 24B.1 Error Handling Philosophy (Server-First, Consistent Surface)

AwoERP operates at a critical intersection: it handles financial transactions, payroll, fuel sales, and KRA tax submissions. When something goes wrong, the error message is not just a developer aid — it is an operational signal that a cashier, finance manager, or pump attendant needs to act on. Cryptic errors ("Error 500"), silent failures, and inconsistently formatted error responses erode trust in an ERP system faster than missing features.

The platform enforces a **server-first error philosophy**: all meaningful errors originate from the server, are typed, localised, and carry machine-readable codes. The UI never invents its own error messages for server-side failures. This ensures that:

1. Error messages are consistent across web and mobile (both receive the same server payload)
2. Error messages can be localised server-side for Swahili (`sw-KE`) or other locales without client code changes
3. Errors are traceable via `trace_id` back to the server log for incident investigation
4. Business-specific error codes (e.g. `etims_cu_not_reachable`, `mpesa_timeout`) are handled by the UI with specific recovery UX rather than generic "something went wrong" messages

The Go backend's error pipeline flows: **domain sentinel error → `BusinessError` struct → HTTP error response envelope**. The UI's error pipeline flows in reverse: **HTTP response → `responseAdaptor` → typed error → display pattern selection → user-facing UX**.

---

## 24B.2 Error Classification

### 24B.2.1 Validation Errors

Validation errors arise from user input that fails field-level or cross-field validation rules. They are always associated with specific form fields and must be displayed inline next to the offending field. HTTP status: `400 Bad Request`.

Example: submitting an invoice where the VAT rate is invalid per KRA rules, or entering a fuel quantity that exceeds the tank capacity.

### 24B.2.2 Business Rule Errors

Business rule errors arise from operations that are syntactically valid but violate a domain invariant. Examples: posting a journal entry where debits ≠ credits, approving a PO that exceeds the approval limit for the approver's role, or cancelling an invoice that has already been submitted to eTIMS. HTTP status: `422 Unprocessable Entity`.

### 24B.2.3 Not Found / Resource Errors

The requested resource does not exist in the current tenant's scope. HTTP status: `404 Not Found`. The UI should display a "not found" surface rather than a generic error toast.

### 24B.2.4 Authorization Errors

The current session does not have permission to perform the requested operation. HTTP status: `403 Forbidden`. The UI must handle this gracefully — showing a "permission denied" message without exposing which permission is missing (information disclosure risk).

### 24B.2.5 Network / Transport Errors

The HTTP request failed to reach the server (no connection, DNS failure, timeout). These errors are generated client-side, not by the server. HTTP status: none (request never completed). The UI displays a connectivity error with retry guidance.

### 24B.2.6 Rendering Engine Errors

The amis renderer encountered an invalid schema node or an unexpected runtime state. These errors are caught by the amis error boundary and reported to the platform error tracking service with the surface ID and schema context.

### 24B.2.7 Unknown / Unhandled Errors

Any error that does not match a known pattern. HTTP status: `500 Internal Server Error`. The UI displays a generic error boundary with the `trace_id` for support purposes.

---

## 24B.3 Error Response Envelope (from AWO ERP `BusinessError`)

All error responses from the AwoERP API use the `BusinessError` envelope:

```go
// internal/shared/errors/business_error.go

type BusinessError struct {
    Code        string            `json:"code"`         // Machine-readable error code
    Message     string            `json:"message"`      // Localised user-facing message
    FieldErrors map[string]string `json:"field_errors,omitempty"` // Per-field validation errors
    TraceID     string            `json:"trace_id"`     // OpenTelemetry trace ID
    Status      int               `json:"-"`            // HTTP status (not in body)
    Err         error             `json:"-"`            // Wrapped underlying error
}

func (e *BusinessError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *BusinessError) Unwrap() error { return e.Err }

// HTTP response body
type ErrorResponse struct {
    Error BusinessError `json:"error"`
}
```

The HTTP handler maps domain errors to HTTP responses via `shared/errors/http.go`:

```go
// internal/shared/errors/http.go

func ToHTTPError(c *fiber.Ctx, err error) error {
    var be *BusinessError
    if errors.As(err, &be) {
        // Inject trace ID from OpenTelemetry span
        span := trace.SpanFromContext(c.UserContext())
        be.TraceID = span.SpanContext().TraceID().String()

        return c.Status(be.Status).JSON(ErrorResponse{Error: *be})
    }
    // Unknown error — log and return generic 500
    log.Error().Err(err).Msg("unhandled error")
    return c.Status(500).JSON(ErrorResponse{Error: BusinessError{
        Code:    "internal_error",
        Message: "An unexpected error occurred. Please try again or contact support.",
        TraceID: trace.SpanFromContext(c.UserContext()).SpanContext().TraceID().String(),
    }})
}
```

> **✅ Convention:** Always use `errors.As` (not type switch) to unwrap `*BusinessError`. Domain errors are often wrapped: `fmt.Errorf("create invoice: %w", businessErr)`. A type switch would fail to find the `*BusinessError` through wrapping layers. `errors.As` walks the chain.

### 24B.3.1 `code` — Machine-Readable Error Code

Error codes are namespaced by domain: `finance.invoice.etims_unsigned`, `hr.leave.overlap`, `procurement.po.approval_limit_exceeded`, `mpesa.stk_push_timeout`. Codes use snake_case and must be documented in the domain's error catalogue.

### 24B.3.2 `message` — Localised User-Facing Message

Messages are resolved from the server-side i18n bundle keyed by `code` and the request `Accept-Language` header. For Kenya deployments: `en-KE` English and `sw-KE` Swahili are supported.

### 24B.3.3 `field_errors` — Per-Field Validation Map

```json
{
  "error": {
    "code": "validation_failed",
    "message": "Please correct the highlighted fields",
    "field_errors": {
      "invoice_date": "Invoice date cannot be in the future",
      "buyer_kra_pin": "Invalid KRA PIN format. Expected: A000000000A",
      "line_items[2].quantity": "Quantity must be greater than 0"
    },
    "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736"
  }
}
```

### 24B.3.4 `trace_id` — OpenTelemetry Trace Reference

The `trace_id` maps to an OpenTelemetry trace in Jaeger/Tempo. Support staff can use this ID to find the exact server-side log lines, DB queries, and external API calls associated with a failing request. The UI displays it prominently in error dialogs:

```
If this error persists, contact support with reference: 4bf92f3577b34da6a3ce929d0e0e4736
```

---

## 24B.4 Client-Side Error Display Patterns

### 24B.4.1 Inline Field Errors

Inline errors are displayed beneath the relevant form field. The amis `responseAdaptor` transforms the `field_errors` map into amis's `msg` format:

```javascript
// web/src/api/response-adaptor.js
function responseAdaptor(payload, response) {
    if (payload.error) {
        var err = payload.error;
        var result = { status: 0, msg: err.message };

        // Map field_errors to amis field-level errors
        if (err.field_errors) {
            result.errors = err.field_errors; // amis reads this for inline field errors
        }

        // Append trace_id to message for support
        if (err.trace_id) {
            result.msg += '\n\nReference: ' + err.trace_id;
        }

        return result;
    }
    // Success — unwrap the AWO data envelope
    return { status: 0, data: payload.data, msg: '' };
}
```

amis forms automatically display `errors` map values beneath their corresponding fields when the form submission API returns this format.

### 24B.4.2 Toast / Snackbar Notifications

For non-validation errors (business rule violations, network errors, success confirmations), the platform uses amis's `notify` env override to display a toast:

```javascript
// Configured in amis.embed() env
notify: function(type, msg) {
    // type: 'success' | 'error' | 'info' | 'warning'
    AwoToast.show({ type: type, message: msg, duration: type === 'error' ? 8000 : 3000 });
}
```

Error toasts persist for 8 seconds (double the default) to give users time to read the message and note the trace ID.

### 24B.4.3 Full-Page Error States

When a surface fails to load (404, 403, 500), the content area renders a full-page error surface:

```json
{
  "type": "container",
  "className": "awo-error-page",
  "body": [
    { "type": "icon", "icon": "fa fa-exclamation-circle", "className": "error-icon" },
    { "type": "tpl", "tpl": "<h2>{{errorTitle}}</h2>" },
    { "type": "tpl", "tpl": "<p>{{errorMessage}}</p>" },
    { "type": "tpl", "tpl": "<small>Reference: {{traceId}}</small>", "visibleOn": "!!traceId" },
    { "type": "action", "label": "Go Back", "actionType": "link", "link": "javascript:history.back()" }
  ]
}
```

### 24B.4.4 Component-Level Error Boundaries

Individual components can declare fallback content for when their data source fails:

```json
{
  "type": "service",
  "api": "GET /api/v1/finance/dashboard/summary",
  "errorBody": {
    "type": "alert",
    "level": "warning",
    "body": "Dashboard data unavailable. ${error.message}",
    "actions": [{ "type": "action", "label": "Retry", "actionType": "reload" }]
  },
  "body": { /* ... normal dashboard content ... */ }
}
```

---

## 24B.5 Retry Policies

### 24B.5.1 Automatic Retry for Idempotent Actions

`GET` and `PUT` requests (idempotent) are automatically retried up to 3 times on network errors (`status: 0` in amis). `POST`, `PATCH`, `DELETE` are never automatically retried — they may have already executed server-side.

### 24B.5.2 User-Initiated Retry

Every error toast and error boundary includes a "Retry" action that re-executes the failed operation. The retry action reuses the original request parameters.

### 24B.5.3 Exponential Back-Off

The `env.fetcher` custom fetch wrapper implements exponential back-off for network errors:

```javascript
async function fetchWithRetry(url, options, maxRetries = 3, baseDelay = 500) {
    for (var attempt = 0; attempt <= maxRetries; attempt++) {
        try {
            var response = await fetch(url, options);
            if (response.ok || response.status >= 400) return response; // 4xx: don't retry
            throw new Error('Server error: ' + response.status);
        } catch (err) {
            if (attempt === maxRetries) throw err;
            // Only retry GET/HEAD
            if (options.method && options.method !== 'GET' && options.method !== 'HEAD') throw err;
            var delay = baseDelay * Math.pow(2, attempt) + Math.random() * 100;
            await new Promise(r => setTimeout(r, delay));
        }
    }
}
```

---

## 24B.6 Error Fallback UIs in AST

Surface definitions can declare module-level fallback UIs that activate when the surface's primary data source is unavailable. This is distinct from component-level `errorBody` — it applies to the entire surface:

```json
{
  "surface_id": "finance.invoices.list",
  "error_fallback": {
    "type": "container",
    "body": [
      { "type": "alert", "level": "info", "title": "Finance module temporarily unavailable" },
      { "type": "tpl", "tpl": "Please try again in a few minutes. If the issue persists, contact support." }
    ]
  }
}
```

The compilation pipeline validates that `error_fallback` schemas do not reference any API calls (to prevent recursive failures) and contain only static display nodes.

---

## 24B.7 Error Propagation in Action Chains

amis action chains execute sequentially; a failing action must not silently proceed to the next action. AwoERP's action chain error handling:

```json
{
  "type": "action",
  "label": "Submit Invoice",
  "actionType": "ajax",
  "api": "POST /api/v1/finance/invoices",
  "onSuccess": [
    {
      "actionType": "ajax",
      "api": "POST /api/v1/finance/invoices/${id}/etims-submit",
      "onSuccess": [
        { "actionType": "toast", "args": { "msg": "Invoice submitted to KRA eTIMS", "msgType": "success" } }
      ],
      "onFailed": [
        { "actionType": "toast", "args": { "msg": "Invoice saved but eTIMS submission failed. Reference: ${__error.trace_id}", "msgType": "warning" } }
      ]
    }
  ],
  "onFailed": [
    { "actionType": "toast", "args": { "msg": "${__error.message}", "msgType": "danger" } }
  ]
}
```

The `__error` amis variable is populated with the `BusinessError` payload when an action fails, making `trace_id`, `code`, and `message` available to subsequent actions in the `onFailed` chain.

Kenya-specific M-Pesa error codes deserve dedicated `onFailed` actions:

```json
{
  "onFailed": [
    {
      "actionType": "custom",
      "script": "if (event.data.__error.code === 'mpesa.stk_push_timeout') { amis.toast('warning', 'M-Pesa request timed out. Ask customer to check their phone and try again.'); } else { amis.toast('error', event.data.__error.message); }"
    }
  ]
}
```

---

## 24B.8 Error Logging and Alerting

All errors returned from the AwoERP API are logged via Zerolog with structured fields:

```go
log.Error().
    Err(err).
    Str("trace_id", traceID).
    Str("tenant_id", tenantID.String()).
    Str("user_id", userID.String()).
    Str("error_code", be.Code).
    Int("http_status", be.Status).
    Str("path", c.Path()).
    Msg("request failed")
```

Prometheus alert rules fire for:
- `awo_http_5xx_rate > 0.01` for 5 consecutive minutes → PagerDuty alert
- `awo_http_422_rate` spike (> 10× baseline) → Slack warning (possible data integrity issue)
- `awo_etims_error_rate > 0` → Immediate Slack alert (KRA compliance risk)

> See Chapter 38 — Observability for the full alerting ruleset and runbook references.
