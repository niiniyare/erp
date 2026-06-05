---
title: "API Integration Patterns"
volume: "VI — Platform & API"
chapter: "25-B"
phase: 3
status: draft
audience: [Backend Engineers, Frontend Engineers, Mobile Engineers, Third-Party Integrators]
---

# Chapter 25-B — API Integration Patterns

> **Volume:** VI — Platform & API
> **Audience:** Backend Engineers, Frontend Engineers, Mobile Engineers, Third-Party Integrators
> **Prerequisites:** Chapter 25 — API Contracts, Chapter 26 — Security Model, Chapter 24-B — Error Handling Strategy
> **Phase:** 3

---

## 25B.1 Integration Philosophy

### 25B.1.1 UI as a Thin Client on Top of the ERP API

The AwoERP SDUI platform treats every UI surface as a thin declarative shell over the ERP REST API. No business logic lives in the UI — all computation, validation, and state transitions happen in the Go backend. The UI's responsibilities are: present data, capture input, and invoke API endpoints. This constraint enables two critical properties: the mobile Flutter renderer and the web amis renderer can be replaced or upgraded independently without any backend changes, and the same API that drives the UI can be used by third-party integrators without a separate integration layer.

### 25B.1.2 Never Bypassing the API Layer from the UI

UI surfaces never read from or write to PostgreSQL, Redis, or any other infrastructure component directly. All data access flows through `internal/api/handlers/` endpoints. This is not a performance recommendation — it is a hard architectural rule enforced by the fact that the browser and Flutter app have no database credentials.

In the Go backend, all handlers run under `store.WithTenant(ctx, tenantID, ...)` which establishes the RLS context. A UI surface bypassing this layer would lose RLS isolation — a critical multi-tenant security guarantee.

### 25B.1.3 Contract-First API Design

New API endpoints consumed by UI surfaces must have their request/response contracts documented in the API contracts system before implementation begins. The UI surface definition references the endpoint, and the endpoint implements the contract. This ensures that UI and backend development can proceed in parallel using mock responses.

---

## 25B.2 Authentication Integration

### 25B.2.1 JWT Bearer Token Propagation

Every API request from the UI includes the JWT bearer token in the `Authorization` header:

```
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

The token is validated by `internal/api/middleware/auth.go`'s `authenticateMiddleware`. The token payload includes `user_id`, `tenant_id`, `portal_type`, and `session_id`. It does not include permissions — permissions are loaded from the session store on the first request and cached in the request context.

### 25B.2.2 Token Storage (Secure Storage — Never LocalStorage)

Web clients store the JWT in an `HttpOnly`, `Secure`, `SameSite=Strict` cookie managed by the Go backend's session endpoint. `localStorage` is explicitly forbidden for token storage — it is accessible to any JavaScript on the page, making it vulnerable to XSS.

```go
// internal/api/handlers/auth/auth.go
func (h *AuthHandler) Login(c *fiber.Ctx) error {
    // ... validate credentials ...
    c.Cookie(&fiber.Cookie{
        Name:     "awo_session",
        Value:    token,
        HTTPOnly: true,
        Secure:   true,
        SameSite: "Strict",
        MaxAge:   int(sessionDuration.Seconds()),
        Path:     "/",
    })
    // Also return token in response body for Flutter (which cannot use cookies)
    return c.JSON(fiber.Map{"token": token, "expires_at": expiresAt})
}
```

Flutter stores the token in `flutter_secure_storage` (uses Android Keystore / iOS Secure Enclave):

```dart
// lib/auth/token_store.dart
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class TokenStore {
    static const _storage = FlutterSecureStorage();
    static const _key = 'awo_jwt';

    static Future<void> save(String token) => _storage.write(key: _key, value: token);
    static Future<String?> get() => _storage.read(key: _key);
    static Future<void> delete() => _storage.delete(key: _key);
}
```

### 25B.2.3 Token Refresh Flow from UI Actions

When the server returns HTTP 401 with code `token_expired`, the amis `env.fetcher` intercepts the response, triggers a token refresh via `POST /api/v1/auth/refresh`, and retries the original request:

```javascript
// Full env.fetcher with token refresh — see Appendix of this chapter
async function fetcher(fetcherConfig) {
    var response = await doFetch(fetcherConfig);
    if (response.status === 401) {
        var refreshed = await refreshToken();
        if (refreshed) {
            return doFetch(fetcherConfig); // retry once
        } else {
            redirectToLogin();
            return { status: 401, data: {} };
        }
    }
    return response;
}
```

### 25B.2.4 Session Expiry Handling

When the refresh token itself is expired (long session timeout), the SSE connection sends a `session_expired` event and the UI navigates to the login page, preserving the current URL as `redirect_after_login` for seamless re-authentication.

### 25B.2.5 M-Pesa STK Push Integration Pattern (Kenya-Specific)

Customer portal payment surfaces integrate with M-Pesa STK Push via a two-step UI flow:

**Step 1**: User enters their M-Pesa phone number and confirms amount.

**Step 2**: The UI submits `POST /api/v1/payments/mpesa/stk-push` and enters a polling loop:

```json
{
  "type": "form",
  "api": "POST /api/v1/payments/mpesa/stk-push",
  "onSuccess": {
    "actionType": "custom",
    "script": "startMpesaPolling(event.data.checkout_request_id)"
  },
  "body": [
    { "type": "input-text", "name": "phone", "label": "M-Pesa Phone Number", "placeholder": "e.g. 0712 345 678" },
    { "type": "static", "name": "amount", "label": "Amount (KES)", "value": "${invoice.total_due}" }
  ]
}
```

The `startMpesaPolling` function polls `GET /api/v1/payments/mpesa/status/{checkout_request_id}` every 3 seconds for up to 60 seconds:

```javascript
function startMpesaPolling(checkoutRequestId) {
    var attempts = 0;
    var interval = setInterval(async function() {
        attempts++;
        if (attempts > 20) { // 60 seconds
            clearInterval(interval);
            amis.toast('warning', 'M-Pesa request timed out. Check your phone and try again.');
            return;
        }
        var result = await api.get('/api/v1/payments/mpesa/status/' + checkoutRequestId);
        if (result.data.status === 'COMPLETED') {
            clearInterval(interval);
            amis.reload('invoice-detail'); // refresh the invoice view
            amis.toast('success', 'Payment received via M-Pesa. Thank you!');
        } else if (result.data.status === 'FAILED') {
            clearInterval(interval);
            amis.toast('error', result.data.message || 'M-Pesa payment failed.');
        }
    }, 3000);
}
```

---

## 25B.3 Tenant Context Propagation

### 25B.3.1 `X-Tenant-ID` Header — Mandatory on All Requests

Every API request from the UI must include the `X-Tenant-ID` header. The tenant ID is obtained from `window.__AWO_BOOTSTRAP__.tenantId` at page load and injected by the `env.fetcher` wrapper. Requests without this header are rejected by `internal/api/middleware/tenant.go` with HTTP 400.

```
X-Tenant-ID: 3f2e1d4c-5b6a-7c8d-9e0f-a1b2c3d4e5f6
```

### 25B.3.2 How UI Data Sources Inject `X-Tenant-ID`

The custom `env.fetcher` is the single injection point — all amis `api`, `service`, and `crud` components use the same fetcher, so the header is added to every request automatically:

```javascript
// Registered in amis.embed({ env: { fetcher: awoFetcher } })
async function awoFetcher(fetcherConfig) {
    var config = Object.assign({}, fetcherConfig);
    config.headers = Object.assign({}, config.headers || {}, {
        'X-Tenant-ID':    window.__AWO_BOOTSTRAP__.tenantId,
        'Authorization':  'Bearer ' + await getToken(),
        'Accept':         'application/json',
        'X-Portal-Type':  window.__AWO_BOOTSTRAP__.portalType
    });
    return window.fetch(config.url, config);
}
```

### 25B.3.3 Subdomain-to-Tenant Resolution at API Gateway

For tenants with custom domains (e.g. `erp.shellmaanzoni.co.ke`), the API gateway resolves the tenant from the request hostname and injects `X-Tenant-ID` before forwarding to the Go backend. The UI still includes the header for redundancy and local dev consistency.

### 25B.3.4 RLS Enforcement

The Go backend's `TenantMiddleware` extracts `X-Tenant-ID`, validates it against the session's `tenant_id` claim (preventing header spoofing), and stores it in the request context. All database operations then execute under RLS using `store.WithTenant(ctx, tenantID, ...)`. The UI can rely on the fact that the server will never return data from another tenant, regardless of what is in the URL parameters.

### 25B.3.5 Tenant Context in amis API Requests (`api` component config)

When defining API calls in amis schemas, never hardcode the tenant ID in the URL:

```json
// ❌ Wrong: hardcoded tenant ID
{ "type": "crud", "api": "GET /api/v1/tenants/3f2e1d4c/finance/accounts" }

// ✅ Correct: tenant ID in header (injected by env.fetcher)
{ "type": "crud", "api": "GET /api/v1/finance/accounts" }
```

---

## 25B.4 API Request Patterns from UI Data Sources

### 25B.4.1 List Endpoints — Pagination, Filtering, Sorting

amis `crud` components pass pagination, filtering, and sorting as query parameters:

```json
{
  "type": "crud",
  "api": {
    "method": "get",
    "url": "/api/v1/finance/accounts",
    "adaptor": "return { ...payload, data: payload.data?.items, count: payload.data?.total };"
  },
  "defaultParams": { "page": 1, "perPage": 20 },
  "filterDefaultParams": { "status": "active" }
}
```

The Go handler reads these parameters:

```go
type ListAccountsParams struct {
    Page    int    `query:"page"`
    PerPage int    `query:"perPage"`
    Status  string `query:"status"`
    Search  string `query:"q"`
    SortBy  string `query:"sortBy"`
    SortDir string `query:"sortDir"`
}
```

### 25B.4.2 Detail Endpoints — Route Parameter Binding

amis `service` components use `${id}` template interpolation to bind route parameters from the URL context:

```json
{
  "type": "service",
  "api": "GET /api/v1/finance/invoices/${id}",
  "body": [ /* ... detail fields ... */ ]
}
```

### 25B.4.3 Mutation Endpoints — POST/PATCH/DELETE from Form Submit

Forms submit to mutation endpoints using the `form` component's `api` prop:

```json
{
  "type": "form",
  "api": {
    "method": "patch",
    "url": "/api/v1/finance/invoices/${id}",
    "requestAdaptor": "api.data = { ...api.data, updated_via: 'ui' }; return api;"
  }
}
```

### 25B.4.5 File Upload Endpoints (multipart/form-data)

The amis `input-file` component handles multipart uploads. The upload endpoint path is declared in the component:

```json
{
  "type": "input-file",
  "name": "attachment",
  "label": "Supporting Document",
  "receiver": "/api/v1/documents/upload",
  "accept": ".pdf,.jpg,.png",
  "maxSize": 5242880
}
```

---

## 25B.5 API Module Integration Map

| Module | Base Path | Key Operations |
|--------|-----------|---------------|
| IAM | `/api/v1/iam/` | Users, sessions, roles, permissions |
| Tenant | `/api/v1/tenants/` | Tenant CRUD, modules, navigation config |
| Finance | `/api/v1/finance/` | Accounts, invoices, journal entries, reports |
| HR | `/api/v1/hr/` | Employees, leaves, payroll, payslips |
| Procurement | `/api/v1/procurement/` | POs, GRNs, suppliers |
| Audit | `/api/v1/audit/` | Audit log retrieval |
| Navigation | `/api/v1/ui/navigation` | Navigation tree |
| UI Compilation | `/api/v1/ui/surfaces/` | Surface definitions |
| Temporal Signals | `/api/v1/workflows/` | Signal running workflows |
| Payments | `/api/v1/payments/` | M-Pesa, bank transfers |

---

## 25B.6 API Permission Guards in UI

### 25B.6.1 Mapping Permissions to Endpoint Visibility

The `Authorize()` middleware in `internal/api/middleware/auth.go` enforces permissions server-side. The UI uses the pre-loaded `session.Permissions` to conditionally show/hide UI elements:

```json
{
  "type": "action",
  "label": "Post Journal Entry",
  "visibleOn": "ARRAYINCLUDES(session.permissions, 'finance.journal_entries.post')",
  "actionType": "ajax",
  "api": "POST /api/v1/finance/journal-entries/${id}/post"
}
```

### 25B.6.2 Pre-Flight Permission Check vs. Optimistic Render + 403 Handling

The recommended pattern is **optimistic render + graceful 403 handling**. The UI shows elements based on `session.permissions` (loaded at login) and handles 403 responses from the API gracefully. It does not make extra permission-check API calls. The server is always the final authority.

---

## 25B.7 API Error Mapping to UI Error States

| HTTP Status | Error Code Pattern | UI Pattern |
|-------------|-------------------|------------|
| 400 | `validation_failed` | Inline field errors via `responseAdaptor` |
| 401 | `token_expired` | Silent token refresh, then retry |
| 401 | `session_invalid` | Redirect to login |
| 403 | `permission_denied` | Toast: "You don't have permission to do this" |
| 404 | `not_found` | Full-page not-found surface |
| 409 | `conflict` | Toast with conflict details |
| 422 | `business_rule_*` | Toast with domain-specific message |
| 429 | `rate_limited` | Toast: "Too many requests. Please wait and try again." |
| 5xx | `internal_error` | Error boundary with trace_id |

---

## 25B.8 amis-Specific API Integration

### 25B.8.1 amis `env.fetcher` — Complete Implementation

```javascript
// web/src/api/fetcher.js
// This is the single fetch wrapper used by ALL amis API calls.

var tokenRefreshPromise = null;

async function getToken() {
    // Tokens are in HttpOnly cookies for web — we rely on credentials: 'include'
    // For Flutter, this function reads from flutter_secure_storage via a bridge
    return null; // Cookie-based auth: token is automatic via cookie
}

async function refreshToken() {
    if (tokenRefreshPromise) return tokenRefreshPromise;
    tokenRefreshPromise = fetch('/api/v1/auth/refresh', {
        method: 'POST',
        credentials: 'include'
    }).then(function(r) {
        tokenRefreshPromise = null;
        return r.ok;
    }).catch(function() {
        tokenRefreshPromise = null;
        return false;
    });
    return tokenRefreshPromise;
}

async function awoFetcher(config) {
    var init = {
        method:      config.method || 'GET',
        credentials: 'include',  // sends HttpOnly session cookie
        headers: Object.assign({
            'X-Tenant-ID':   window.__AWO_BOOTSTRAP__.tenantId,
            'X-Portal-Type': window.__AWO_BOOTSTRAP__.portalType,
            'Accept':        'application/json',
            'Accept-Language': window.__AWO_BOOTSTRAP__.locale || 'en-KE'
        }, config.headers || {}),
    };

    if (config.data && init.method !== 'GET' && init.method !== 'HEAD') {
        init.headers['Content-Type'] = 'application/json';
        init.body = JSON.stringify(config.data);
    }

    var response;
    try {
        response = await fetch(config.url, init);
    } catch (networkErr) {
        // Network failure — no response
        return { status: 0, data: {}, msg: 'Network error. Please check your connection.' };
    }

    if (response.status === 401) {
        var refreshed = await refreshToken();
        if (refreshed) {
            response = await fetch(config.url, init); // single retry
        } else {
            var loginUrl = '/login?redirect=' + encodeURIComponent(window.location.pathname);
            window.location.href = loginUrl;
            return { status: 401, data: {}, msg: 'Session expired. Redirecting to login...' };
        }
    }

    var payload;
    try {
        payload = await response.json();
    } catch (e) {
        return { status: response.status, data: {}, msg: 'Invalid server response' };
    }

    return { status: response.status, data: payload, headers: response.headers };
}
```

### 25B.8.2 `requestAdaptor` and `responseAdaptor`

```javascript
// Global amis adaptors registered in amis.embed() env

function requestAdaptor(api) {
    // Normalise all request data: remove undefined keys
    if (api.data) {
        Object.keys(api.data).forEach(function(k) {
            if (api.data[k] === undefined) delete api.data[k];
        });
    }
    return api;
}

function responseAdaptor(payload, response) {
    // Unwrap AWO envelope: { data: {...}, meta: {...}, error: {...} }
    if (!payload) return { status: 1, msg: 'Empty response' };

    if (payload.error) {
        var err = payload.error;
        return {
            status: 0,
            msg: err.message + (err.trace_id ? '\n\nRef: ' + err.trace_id : ''),
            errors: err.field_errors || undefined
        };
    }

    return {
        status: 0,
        data: payload.data,
        msg: payload.meta?.message || ''
    };
}
```

### 25B.8.3 Full Working amis `crud` for Finance Accounts

```json
{
  "type": "crud",
  "syncLocation": false,
  "api": {
    "method": "get",
    "url": "/api/v1/finance/accounts",
    "adaptor": "return { status: 0, data: { items: payload.data?.items || [], total: payload.data?.total || 0 } };"
  },
  "defaultParams": { "page": 1, "perPage": 25 },
  "headerToolbar": [
    { "type": "button", "label": "New Account", "actionType": "link", "link": "/app/finance/accounts/new",
      "visibleOn": "ARRAYINCLUDES(session.permissions, 'finance.accounts.write')" },
    { "type": "export-excel", "label": "Export", "api": "/api/v1/finance/accounts/export?format=xlsx" }
  ],
  "filter": {
    "body": [
      { "type": "input-text", "name": "q", "label": "Search", "placeholder": "Account code or name" },
      { "type": "select", "name": "type", "label": "Type",
        "options": [
          { "label": "All", "value": "" },
          { "label": "Asset", "value": "ASSET" },
          { "label": "Liability", "value": "LIABILITY" },
          { "label": "Income", "value": "INCOME" },
          { "label": "Expense", "value": "EXPENSE" }
        ]
      }
    ]
  },
  "columns": [
    { "name": "code", "label": "Code", "sortable": true },
    { "name": "name", "label": "Account Name", "sortable": true },
    { "name": "type", "label": "Type" },
    { "name": "balance", "label": "Balance (KES)", "type": "number", "precision": 2, "sortable": true },
    { "name": "status", "label": "Status", "type": "tag",
      "map": { "active": { "label": "Active", "color": "success" }, "inactive": { "label": "Inactive", "color": "default" } }
    },
    {
      "name": "actions",
      "label": "Actions",
      "type": "operation",
      "buttons": [
        { "label": "View", "actionType": "link", "link": "/app/finance/accounts/${id}" },
        { "label": "Edit", "actionType": "link", "link": "/app/finance/accounts/${id}/edit",
          "visibleOn": "ARRAYINCLUDES(session.permissions, 'finance.accounts.write')" }
      ]
    }
  ]
}
```

---

## 25B.9 Flutter-Specific API Integration

### 25B.9.1 Dio HTTP Client Configuration

```dart
// lib/api/dio_client.dart
import 'package:dio/dio.dart';

Dio buildDioClient(AppConfig config) {
  var dio = Dio(BaseOptions(
    baseUrl:         config.apiBaseUrl,
    connectTimeout:  const Duration(seconds: 10),
    receiveTimeout:  const Duration(seconds: 30),
    headers: {
      'Accept':       'application/json',
      'Content-Type': 'application/json',
    },
  ));

  dio.interceptors.addAll([
    AuthInterceptor(tokenStore: TokenStore()),
    TenantInterceptor(tenantId: config.tenantId, portalType: config.portalType),
    LoggingInterceptor(),
    ErrorInterceptor(),
  ]);

  return dio;
}
```

### 25B.9.2 Interceptor Chain (Auth + Tenant + Logging)

```dart
// lib/api/interceptors/auth_interceptor.dart
class AuthInterceptor extends Interceptor {
  final TokenStore tokenStore;

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) async {
    var token = await tokenStore.get();
    if (token != null) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    handler.next(options);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    if (err.response?.statusCode == 401) {
      var refreshed = await refreshToken();
      if (refreshed) {
        // Retry original request
        var retryResponse = await err.requestOptions.send();
        return handler.resolve(retryResponse);
      }
      // Navigate to login
      navigatorKey.currentState?.pushReplacementNamed('/login');
    }
    handler.next(err);
  }
}

// lib/api/interceptors/tenant_interceptor.dart
class TenantInterceptor extends Interceptor {
  final String tenantId;
  final String portalType;

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    options.headers['X-Tenant-ID']   = tenantId;
    options.headers['X-Portal-Type'] = portalType;
    handler.next(options);
  }
}
```

### 25B.9.3 Offline Queue Integration

Flutter uses an offline queue (`Hive`-backed) that stores failed mutations and replays them on reconnection:

```dart
class OfflineQueue {
  static Future<void> enqueue(QueuedRequest request) async {
    var box = await Hive.openBox<QueuedRequest>('offline_queue');
    await box.add(request);
  }

  static Future<void> flush(Dio dio) async {
    var box = await Hive.openBox<QueuedRequest>('offline_queue');
    for (var req in box.values.toList()) {
      try {
        await dio.request(req.path, data: req.body, options: Options(method: req.method));
        await req.delete();
      } catch (_) { break; } // Stop on first failure; retry next connectivity event
    }
  }
}
```

---

## 25B.10 Rate Limiting and Retry from UI Layer

The Go backend applies rate limiting at the API gateway level (`100 requests/minute/user` for standard users, `1000 requests/minute` for platform admin). When the UI receives HTTP 429, it implements exponential back-off as described in Chapter 24-B and displays a user-facing warning:

```json
{
  "type": "alert",
  "level": "warning",
  "visibleOn": "lastError.code === 'rate_limited'",
  "body": "You're making requests too quickly. Please wait a moment before trying again."
}
```

---

## 25B.11 API Contract Testing from the UI Perspective

UI integration tests assert that the API contracts match what the UI surface definitions expect. The test suite at `internal/api/handlers/*_test.go` uses `httptest` to validate request/response shapes. The amis schema files in `web/schemas/pages/` are validated against the API contract using the `awo ui validate` CLI:

```bash
# Validate that all API references in UI schemas match registered endpoints
awo ui validate --check-api-refs web/schemas/pages/

# Check that response shapes match responseAdaptor expectations
awo ui validate --check-response-shapes web/schemas/pages/
```

> See Chapter 41 — Testing Strategy for the full contract testing approach and Chapter 42 — CI/CD Considerations for how these tests run in the pipeline.
