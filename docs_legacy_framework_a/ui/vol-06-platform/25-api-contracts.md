> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
chapter: 25
title: "API Contracts"
volume: "vol-06-platform"
section: "Platform"
description: "The schema endpoint contract, request requirements, response envelopes, and route naming conventions."
status: implemented
---

# Chapter 25 — API Contracts

## Table of Contents

- [25.1 The Schema Endpoint — GET /schema/*](#251-the-schema-endpoint)
- [25.2 Request Requirements](#252-request-requirements)
- [25.3 Response Envelopes](#253-response-envelopes)
- [25.4 The 401 AMIS Envelope](#254-the-401-amis-envelope)
- [25.5 Route Naming Conventions](#255-route-naming-conventions)
- [25.6 App Shell Endpoint](#256-app-shell-endpoint)

---

## 25.1 The Schema Endpoint

The schema API has a single entry point: a wildcard GET route that captures everything after `/schema/` as the route key.

```
GET /schema/*
```

The `*` captures the full remainder of the path — including slashes — and becomes the route key used to look up the registered `PageFn` in the schema registry. Examples:

| Request URL                          | Resolved route key            |
|--------------------------------------|-------------------------------|
| `GET /schema/finance/invoices`       | `finance/invoices`            |
| `GET /schema/hr/payroll/run`         | `hr/payroll/run`              |
| `GET /schema/settings/tenant`        | `settings/tenant`             |
| `GET /schema/`                       | `` (empty — nav tree request) |

The route key must exactly match the key used when registering the page with the schema registry. There is no fuzzy matching, prefix expansion, or inference.

---

## 25.2 Request Requirements

### Middleware Chain

Every request to `/schema/*` must pass through the full middleware chain in this order:

```
Authenticate → InjectSessionContext → Handle
```

**Authenticate** validates the JWT from the `Authorization: Bearer <token>` header. If validation fails, the pipeline short-circuits with a 401 AMIS envelope (see §25.4). The raw token claims are attached to the Fiber context.

**InjectSessionContext** reads the tenant ID and user ID from the validated claims, calls the IAM layer to hydrate user permissions and feature flags, and stores the resulting `contract.SessionContext` in the Fiber context. Downstream handlers retrieve it via `contract.FromContext(ctx)`.

**Handle** is the schema handler itself. It never reads Fiber locals directly — all identity information is accessed exclusively through `contract.FromContext`.

### Request Body

None. The schema endpoint is GET-only. The route path is the only input parameter. Query parameters are not used by the pipeline and are not forwarded to page functions.

### Headers

| Header          | Required | Notes                              |
|-----------------|----------|------------------------------------|
| `Authorization` | Yes      | `Bearer <jwt>` — RS256 or HS256    |
| `Accept`        | No       | Ignored; response is always JSON   |

---

## 25.3 Response Envelopes

All responses use a consistent JSON envelope regardless of success or failure. The `status` field mirrors HTTP status codes except for the success case, where it is `0`.

### 200 — Schema Found

```json
{
  "status": 0,
  "data": {
    "type": "page",
    "title": "Invoices",
    "body": [ ... ]
  }
}
```

The `data` field contains the fully compiled AMIS page schema. The schema is ready to render — permissions have been applied, disabled buttons removed, and invisible components pruned.

### 401 — Authentication Failure

The 401 response is an AMIS envelope, not a plain error object. See §25.4 for the rationale.

```json
{
  "status": 401,
  "msg": "Session expired. Please log in.",
  "data": {
    "type": "page",
    "body": [
      {
        "type": "alert",
        "level": "warning",
        "body": "Your session has expired. Redirecting to login..."
      }
    ]
  }
}
```

### 404 — Route Not Registered

```json
{
  "status": 404,
  "msg": "schema not found: /finance/nonexistent"
}
```

The `msg` always includes the route that was requested. There is no `data` field in 404 responses.

### 500 — Internal Error

```json
{
  "status": 500,
  "msg": "An internal error occurred. Please try again."
}
```

Internal error details are never surfaced to the client. All errors are logged server-side with a request trace ID.

### 503 — IAM Resolution Failure

Returned when the AuthzStage cannot resolve permissions from the IAM service (e.g., the IAM backend is unavailable or returns an error).

```json
{
  "status": 503,
  "msg": "Authorization service unavailable. Please try again."
}
```

The app shell handles 503 by displaying a retry banner without clearing the current page content.

---

## 25.4 The 401 AMIS Envelope

The 401 response carries an AMIS page schema in its `data` field. This is intentional.

The AMIS SDK renders whatever is returned by the schema API, including error responses. If the 401 returned a plain `{"error": "unauthorized"}` object, the SDK would attempt to render it as a schema and produce a blank or broken page.

By returning a valid AMIS page schema in the 401 envelope, the app shell can:

1. Display a user-friendly "session expired" message without custom JavaScript.
2. Automatically render a redirect button or countdown timer as an AMIS action.
3. Preserve the AMIS rendering lifecycle — no special-case handling required in the frontend.

The `status: 401` field tells the app shell's response interceptor to also trigger a session-expiry flow (clearing the auth token, stopping polling), while the `data` field gives the user something visible to read.

---

## 25.5 Route Naming Conventions

Route keys follow a dotted-module-then-slash hierarchy. Rules:

- **Module prefix first**: `finance/invoices`, not `invoices/finance`
- **Lowercase, hyphenated**: `purchase-orders`, not `PurchaseOrders`
- **Singular resource**: `finance/invoice` for a single-record view, `finance/invoices` for a list
- **No leading slash in the key**: the registry key is `finance/invoices`, not `/finance/invoices`
- **No trailing slash**: `hr/payroll` not `hr/payroll/`

The registry performs exact key lookup. A page registered as `finance/invoices` will not be found by a request to `GET /schema/finance/invoices/` (trailing slash) or `GET /schema/Finance/Invoices` (wrong case).

### Good Examples

```
finance/invoices            ← list page
finance/invoice             ← detail page
finance/invoice/new         ← create form
hr/employees                ← list
hr/employee                 ← detail
settings/tenant             ← tenant settings page
```

### Bad Examples

```
Finance/Invoices            ← wrong case
finance/invoices/           ← trailing slash
invoices                    ← missing module prefix
finance_invoices            ← underscore instead of slash
finance/invoicesList        ← camelCase
```

---

## 25.6 App Shell Endpoint

The app shell requests the navigation tree via a special compilation key defined as `ui.AppOperationKey`. This is not a user-facing page — it returns the sidebar navigation schema.

```
GET /schema/    (empty wildcard → resolves to AppOperationKey)
```

The navigation tree compilation follows the same pipeline as regular pages: AuthzStage resolves permissions, CacheStage checks for a cached nav tree keyed by tenant + permission fingerprint, and the resulting schema is a top-level AMIS `nav` component listing only the modules and routes the user has permission to see.

The nav tree is compiled once per unique permission fingerprint and cached. Users who share the same permission set — even across tenants — will have isolated nav trees because the cache key includes `tenant_id`.

The nav tree response uses the same `status: 0, data: <schema>` envelope as page schemas.
