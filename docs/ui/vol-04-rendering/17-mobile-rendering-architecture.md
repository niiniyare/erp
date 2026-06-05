# Chapter 17 — Mobile Rendering Architecture

| | |
|---|---|
| **Volume** | 04 — Rendering |
| **Chapter** | 17 |
| **Status** | Planned — Not Yet Implemented |
| **Depends on** | Ch 18 (Web Rendering), Ch 10 (PageNode), Ch 12 (ActionNode) |

---

> **Status: Planned — Not Yet Implemented**
>
> No mobile client exists today. The web SDK (Ch 18) is the sole rendering engine.
> This chapter documents the intended contract so that a future Flutter or React
> Native client can be built against the same backend without changes to the
> schema pipeline.

---

## Table of Contents

1. [17.1 Overview](#171-overview)
2. [17.2 Schema Contract](#172-schema-contract)
3. [17.3 Node Types a Mobile Client Must Handle](#173-node-types-a-mobile-client-must-handle)
4. [17.4 Authentication Flow](#174-authentication-flow)
5. [17.5 Navigation Model](#175-navigation-model)
6. [17.6 Reference Implementation](#176-reference-implementation)

---

## 17.1 Overview

The AwoERP UI pipeline produces renderer-agnostic JSON schemas. The web browser
renders them with the AMIS SDK (see Ch 18). A mobile client would consume the
**same `/schema/*` endpoint** and the **same JSON schema format** — no separate
mobile API is planned.

The mobile client is responsible for:

- Translating AMIS node types (`"page"`, `"form"`, `"crud"`, `"chart"`) into
  native widgets.
- Honouring the AMIS data-chain contract (parent scope visible to children).
- Enforcing the same security rules (do not expose `visibleOn`/`disabledOn`
  permission strings to the user).

Two runtime targets are under consideration:

| Target | Framework | Status |
|---|---|---|
| iOS / Android | Flutter | Planned |
| iOS / Android | React Native | Planned |

---

## 17.2 Schema Contract

A mobile client fetches schemas over the same HTTP endpoint the web shell uses:

```
GET /schema/<route>
Authorization: Bearer <jwt>
Accept: application/json
```

The server responds with an AMIS envelope:

```json
{
  "status": 0,
  "data": {
    "type": "page",
    "title": "Finance Dashboard",
    "initApi": { "method": "GET", "url": "/api/v1/finance/summary" },
    "body": [ ... ]
  }
}
```

A `status` value of `0` means success. Non-zero means error. The mobile client
must unwrap `data` and pass it to its rendering engine.

### 401 envelope

When the session is expired or absent the server returns an AMIS page schema
containing an alert block and a Login button — **not** a bare HTTP 401. The
mobile client must detect this pattern and redirect to the native login screen:

```json
{
  "status": 0,
  "data": {
    "type": "page",
    "body": [
      {
        "type": "alert",
        "level": "warning",
        "body": "Your session has expired. Please log in again."
      },
      {
        "type": "button",
        "label": "Log In",
        "actionType": "url",
        "url": "/ui/login"
      }
    ]
  }
}
```

Detection heuristic: if the page body contains a single alert + a button with
`actionType: "url"` pointing to `/ui/login`, treat the response as a 401 and
navigate to the native login screen.

---

## 17.3 Node Types a Mobile Client Must Handle

The following AMIS node types appear in production schemas. A mobile client must
map each to a native widget.

| AMIS `type` | Native equivalent (Flutter suggestion) |
|---|---|
| `page` | `Scaffold` with `AppBar` and scrollable body |
| `panel` | `Card` with title and body |
| `form` | `Form` widget with `TextFormField`, `DropdownButtonFormField`, etc. |
| `input-text` | `TextFormField` |
| `input-number` | `TextFormField` with `TextInputType.number` |
| `select` | `DropdownButtonFormField` or bottom-sheet picker |
| `crud` | Paginated `ListView` or `DataTable` |
| `chart` | `fl_chart` or equivalent charting library |
| `grid` | `GridView.count` |
| `button` | `ElevatedButton` / `TextButton` (level → style) |
| `tpl` | Inline template string rendered as `Text` |
| `alert` | `Banner` or `SnackBar` |
| `service` | Invisible data-fetching widget that populates child scope |
| `tabs` | `TabBar` + `TabBarView` |
| `divider` | `Divider` |

Unknown `type` values must be rendered as an unsupported-node placeholder, not
silently dropped.

### ActionNode types

Every button or toolbar action is an `ActionNode`. A mobile client must handle:

| `actionType` | Behaviour |
|---|---|
| `ajax` | Call `api`, show loading indicator, handle `status != 0` as error |
| `url` | Navigate to `target` hash/path |
| `dialog` | Present `dialog` schema as a modal sheet |
| `drawer` | Present `drawer` schema as a bottom/side sheet |
| `reload` | Trigger a data reload on the current page's `initApi` |
| `link` | Same as `url` — open in the same navigation stack |

---

## 17.4 Authentication Flow

Authentication uses the same JWT the web client uses. The mobile client obtains
a token from the `/api/v1/auth/login` endpoint and stores it securely
(Keychain on iOS, EncryptedSharedPreferences on Android).

Every request to `/schema/*` and every AMIS `api` call must include the token:

```
Authorization: Bearer <jwt>
```

The mobile client must refresh the token silently using the refresh-token
endpoint before it expires. If refresh fails, navigate to the native login
screen.

---

## 17.5 Navigation Model

The web shell uses hash-based routing (`#finance/dashboard`). A mobile client
should implement a native navigation stack:

1. On launch, fetch the navigation structure from `GET /api/v1/ui/nav` (planned
   endpoint — not yet available; the web shell uses a hard-coded JS `menuConfig`
   array today).
2. Each menu item maps to a route string (`finance/dashboard`).
3. When the user taps a menu item, fetch `GET /schema/finance/dashboard`.
4. Render the returned schema using the native widget mapping above.

Param routing (e.g., `finance/invoices/:id`) is an open item — the web shell
also lacks this capability (see Ch 19, section 19.5).

---

## 17.6 Reference Implementation

The web rendering architecture (Ch 18) is the reference implementation. Any
behaviour documented there that is not marked PLANNED applies equally to a
mobile client unless explicitly noted otherwise.

Key chapters relevant to mobile:

- **Ch 18** — Web rendering: AMIS SDK, data chain, 401 handling
- **Ch 19** — Navigation framework: route resolution, param routing gaps
- **Ch 20** — Action system: all ActionNode types and their contracts
- **Ch 23** — Validation framework: security rules that the server enforces
  (mobile client does not need to replicate these, but must respect their
  output — e.g., `can_approve` is pre-resolved by the server before the schema
  reaches the client)
- **Ch 24** — Data sources: `initApi`, `APISpec`, AMIS expression variables
