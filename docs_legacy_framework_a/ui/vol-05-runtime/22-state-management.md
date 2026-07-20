> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Chapter 22 — State Management

| | |
|---|---|
| **Volume** | 05 — Runtime |
| **Chapter** | 22 |
| **Status** | Implemented (AMIS data chain); client-side state beyond AMIS built-ins: Planned |
| **Source** | AMIS SDK, `web/pages/index.html` |

---

## Table of Contents

1. [22.1 AMIS Data Chain as State Container](#221-amis-data-chain-as-state-container)
2. [22.2 Server as Source of Truth](#222-server-as-source-of-truth)
3. [22.3 Form State](#223-form-state)
4. [22.4 Page Reload Semantics](#224-page-reload-semantics)
5. [22.5 Theme State](#225-theme-state)
6. [22.6 Planned: Extended Client-Side State](#226-planned-extended-client-side-state)

---

## 22.1 AMIS Data Chain as State Container

AwoERP's web UI has **no Redux, Zustand, MobX, or any global state store**. All
UI state is managed by the AMIS SDK's built-in **data chain** — a hierarchical
scope system where each component inherits its parent's data and can extend it.

This is an intentional constraint. ERP state is complex and mutable; keeping the
server as the single source of truth prevents client-side state from drifting out
of sync with business records.

### Data scopes in practice

```
Page scope (root)
├── can_approve: true
├── tenant_currency: "USD"
├── invoice: { id, number, amount, status }
│
├── Service scope (line items)
│   ├── items: [{ ... }, { ... }]
│   └── total_lines: 3
│
└── Form scope (edit form)
    ├── amount: 1500.00     ← being edited by user
    └── notes: ""           ← user is typing
```

Each scope level can read parent variables. A form field can reference
`${tenant_currency}` from the page scope to show the currency symbol.

### What state AMIS manages

| State type | Managed by |
|---|---|
| Page data (loaded from `initApi`) | AMIS page scope |
| Form field values (in-progress edit) | AMIS form scope |
| CRUD filter values | AMIS CRUD filter state |
| CRUD pagination (current page, sort) | AMIS CRUD internal state |
| Dialog / drawer open/closed | AMIS component lifecycle |
| Button disabled/visible | AMIS expression evaluation |
| Tab active index | AMIS tabs component |
| Loading indicators | AMIS component lifecycle |

All of this is handled automatically by AMIS — no explicit state management code
is written in Go or JavaScript.

---

## 22.2 Server as Source of Truth

The pipeline injects all business state into the schema before it is served.
Permissions, tenant settings, and record data are resolved server-side. The
browser receives a complete, pre-resolved schema and only needs to display it.

### Flow

```
User navigates to #finance/invoices/abc-123
  ↓
Shell: GET /schema/finance/invoices/abc-123
  ↓
Pipeline (server):
  1. Authenticate session
  2. Resolve can_approve, can_delete, can_edit
  3. Fetch invoice record from DB (optional — can use initApi instead)
  4. Compile schema with data injected
  ↓
Shell: receives { "data": { can_approve: true, invoice: {...} } }
  ↓
AMIS: mounts page, evaluates expressions, renders
```

All business rules are encoded in the schema the server sends. The browser
executes only display logic (show/hide, enable/disable, format numbers).

### Consequence: no optimistic updates

Because the server is the source of truth, AMIS performs no optimistic updates.
When a user approves an invoice:

1. AMIS sends `POST /api/v1/finance/invoices/{id}/approve`.
2. AMIS shows a loading indicator.
3. Server responds with `{ "status": 0 }`.
4. AMIS triggers a reload of the containing CRUD or form.
5. The fresh data (status: APPROVED) is fetched from the server.
6. AMIS re-renders with the new data.

The user sees a brief loading state, then the updated record. This is the
correct behaviour for financial data where the server must validate state
transitions before the client reflects them.

---

## 22.3 Form State

AMIS tracks form field values internally from the moment a form mounts. The user
can type into fields, change dropdowns, and pick dates — AMIS holds these values
in the form scope until:

- The user submits the form (AMIS posts to `form.api`), or
- The form is closed (dialog/drawer closed, or navigation away).

**Unsaved changes are lost on navigation.** The shell unmounts the current AMIS
instance when the user navigates to a different hash route:

```javascript
if (currentInstance && currentInstance.unmount) {
  currentInstance.unmount();
  currentInstance = null;
}
```

This destroys all AMIS state including in-progress form edits. There is no
"unsaved changes" warning today — see section 22.6 for the planned improvement.

### Form initialisation

Forms that edit existing records initialise their fields from the AMIS data
chain. If the page `initApi` fetches the record and puts it in page scope:

```json
{ "invoice": { "id": "abc", "amount": 1500, "notes": "Urgent" } }
```

Then a form field with `"name": "amount"` and `"value": "${invoice.amount}"`
will initialise to `1500`. AMIS handles this mapping automatically when the
form is defined inside a page or service that already has the data.

---

## 22.4 Page Reload Semantics

When a user navigates away and back to the same route, the shell:

1. Unmounts the old AMIS instance (all state destroyed).
2. Re-fetches `GET /schema/<route>` from the server.
3. Mounts a fresh AMIS instance.

This means **every navigation is a full page re-initialisation**. There is no
client-side route cache. The schema pipeline cache (Ch 16) ensures the server
responds quickly to repeated schema requests, but the browser always starts
fresh.

This is a deliberate design choice: it eliminates stale-state bugs that are
common in SPAs. The trade-off is a brief loading spinner on every navigation.

### Within-page reload

The AMIS `"reload"` action type (Ch 20, section 20.1) triggers a data reload
on a specific component without unmounting it. A CRUDNode reload re-fetches its
`api` and re-renders the table without touching the rest of the page. This is
the mechanism used after a successful create/update/delete action.

---

## 22.5 Theme State

The light/dark theme choice is the only client-side state persisted across
sessions. It is stored in `localStorage` under key `awo-theme`:

```javascript
try { localStorage.setItem('awo-theme', theme); } catch (e) { }
try { return localStorage.getItem('awo-theme') || 'system'; } catch (e) { return 'system'; }
```

The `try/catch` guards against storage-quota exceeded or private-browsing
restrictions. If storage is unavailable, the theme falls back to `'system'`
(follows OS preference) on every page load.

Theme state is **not** sent to the server. The server has no knowledge of the
user's display preference.

---

## 22.6 Planned: Extended Client-Side State

> **Status: Planned — Not Yet Implemented**

The following client-side state capabilities are planned:

### Unsaved-changes guard

Before unmounting a form that has been modified, show a browser confirmation
dialog: "You have unsaved changes. Leave anyway?"

Implementation: AMIS form exposes a `dirty` flag. The shell's `navigate()`
function can check `currentInstance.getState().formDirty` before unmounting.

### Cross-page shared context

Some workflows span multiple pages (e.g. a multi-step purchase order creation).
Each step is a separate hash route but they share context (vendor, line items).
Today this context is lost on navigation.

Planned approach: a small in-memory context object in the shell that survives
navigation but is cleared on page reload:

```javascript
var pageContext = {};  // key-value store, cleared on hard reload
```

Pages participating in a workflow would read and write to this object via the
`amisEnv` hooks.

### Client-side schema cache

Cache the most recently visited schemas in `sessionStorage` to speed up the
back-navigation experience. The cache would be invalidated on schema version
change (requires schema ETag support from the server — also planned, see Ch 23,
section 23.5).

Until these are implemented, client-side state is limited to the AMIS data chain
and the theme preference in `localStorage`.
