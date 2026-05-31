> **📌 AUDIT DOCUMENT.** This is a formal architecture audit, not a response format guide.
> For response format and error handling, see [Response and Error Handling](02-architecture/05-response-and-error-handling.md).
> For the issue tracker derived from this audit, see [Implemented Features](05-roadmap/01-implemented.md#known-production-issues-not-yet-fixed).

# AWO ERP — Architecture Audit Response

**Date:** 2026-05-17 | **Reviewer:** Principal Architect / AMIS Compliance Auditor

---

## Scores

| Dimension | Score | Verdict |
|---|---|---|
| AMIS Alignment | **6/10** | Shell contradicts own documentation |
| Architecture Coherence | **7/10** | Backend solid, frontend split-brained |
| ERP Readiness | **3/10** | Documentation ERP, Implementation CRUD |
| Scalability | **5/10** | Hard ceilings already visible |
| Developer Experience | **6/10** | Good docs, dangerous gap between docs and code |
| Long-Term Maintainability | **5/10** | Several time bombs already planted |

---

## CRITICAL SEVERITY — Will Break Production

### C1: `API_BASE` Hardcoded to `http://localhost:8080/`

**Location:** `web/pages/index.html:1040`

```javascript
var API_BASE = 'http://localhost:8080/';
```

This is a production deploy killer. Every API call is prefixed with this absolute URL. Deploy behind a reverse proxy, a different port, HTTPS, or any non-localhost environment and every single AMIS data fetch breaks with CORS errors. AMIS renders shells, loads schemas, shows empty tables. Silent failure.

**Future failure mode:** First cloud deploy. First demo on a real domain. First container deploy. Breaks all three.

**Fix:** Remove `API_BASE` entirely (AMIS fetcher defaults to relative paths), or read it from a `<meta>` tag injected by Go at serve time.

---

### C2: 401 → Login Redirect Missing From Actual Fetcher

`12-routes-and-auth.md` documents:
```javascript
if (res.status === 401) {
  window.location.href = '/login';
  return;
}
```

Actual fetcher (`index.html:1314–1360`): No 401 redirect. Session expiry produces AMIS error toasts on every component that tries to load. User sees a broken page covered in "Session expired" toasts with no path to login.

This is not a nice-to-have. An expired session in a running ERP produces a visually catastrophic experience.

---

### C3: Vite Proxy Covers `/api`, Not `/schema` — Dev Server Is Decorative

**Location:** `web/vite.config.js:8–12`

```javascript
proxy: {
  '/api': { target: 'http://localhost:8080', changeOrigin: true }
}
```

`navigate()` fetches `/schema/` + route. Vite does not proxy `/schema`. This is rescued only because `API_BASE = 'http://localhost:8080/'` makes all API calls bypass Vite entirely — but that means Vite's proxy provides **zero** functionality. All development traffic goes directly to `:8080`. Vite is a build tool being used only as a build tool, but configured as if it were a dev server.

**Consequence:** Future developers who run `npm run dev` get a broken app. The dev workflow is undocumented and counter-intuitive.

---

### C4: XSS Injection Vector in Breadcrumb Renderer

**Location:** `index.html:1272–1276`

```javascript
breadcrumb.innerHTML = '<span>' + groupName + '</span>'
  + '<i class="fa fa-chevron-right bc-separator"></i>'
  + '<span class="bc-current">' + pageName + '</span>';
```

Currently not exploitable — `groupName` comes from static `menuConfig`. The moment you migrate to Go-driven menu (`GET /api/v1/me/nav`), a backend returning `groupName: "<img src=x onerror=alert(1)>"` from a compromised or misconfigured tenant produces stored XSS in the shell. The docs explicitly plan this migration. The vulnerability is pre-planted.

**Fix required before Go-driven nav:** Use `textContent`, not `innerHTML`, for interpolated strings.

---

## HIGH SEVERITY — Architectural Failures

### H1: Shell Already Migrated But Docs Still Say "Future"

The shell is in a contradictory state.

`04-shell-implementation.md` describes the schema loader as:
```javascript
var schema = await loader.load('pages/' + route + '.json');
```

**Actual `navigate()` in `index.html:1387`:**
```javascript
var res = await fetch('/schema/' + route, {headers: {'Accept': 'application/json'}});
```

The implementation has **already migrated to Go-driven schemas** (`/schema/` endpoints). But:
- `Decision 2` says "switch when first schema needs variation" — already switched
- `Decision 1` says "migrate to AMIS `app` when nav needs Go-driving" — not yet done
- `menuConfig` still hardcodes navigation in JS
- `web/schemas/pages/` still exists as a static folder

Result: Two routing systems, partially migrated, both partially alive, documentation lagging implementation by at least one refactor. Future developers will read docs describing a different architecture than what runs.

---

### H2: Hardcoded `max-height: 500px` on Menu Groups

**Location:** `index.html:305`

```css
.menu-group.expanded .menu-group-items {
  max-height: 500px;
}
```

At ~40px per item, this clips at ~12 items. Stated target is 100+ modules. The navigation system **physically cannot display** a module list at ERP scale. A payroll group alone could have 8+ pages. Finance could have 15+.

CSS animation laziness (avoiding `height: auto` transition) producing a hard ceiling that hurts at scale.

---

### H3: Menu Has 4 Items vs Documented ERP Scope

**Actual `menuConfig`:**
- Dashboard
- Finance (2 items: Chart of Accounts, Transactions)
- Administration (2 items: Users, Tenants)
- Settings

**Documented ERP scope:** GL, AR, AP, Procurement, Inventory, Payroll, Approvals, Reporting, Audit, People management, Multi-currency ledger.

This is not a roadmap issue — it is an architectural honesty issue. The documentation describes a comprehensive multi-module ERP. The implementation is a 4-item sidebar. Aspirational docs in a reference folder teach wrong patterns and destroy trust in the documentation.

---

### H4: Two API Envelope Formats — The Bridge Has a Bug

The fetcher bridges `{success, data, meta}` ↔ `{status, data}`. But the bridge has a silent error:

```javascript
// index.html:1341-1349
if (Array.isArray(json.data) && json.meta && json.meta.pagination) {
  return {
    status: 0,
    data: {
      items: json.data,
      options: json.data,    // ← conflates two different use cases
      count: json.meta.pagination.total_records
    }
  };
}
```

`options: json.data` — every list response populates `options` with the full item list. This is the shape AMIS `select` expects for dropdown options. Someone wired list API responses to also work as select option sources. A `select` that fetches a paginated list endpoint gets only page-1 items as options. A `crud` table gets an `options` key it ignores. Neither component fails, but the contract is wrong for both.

---

### H5: Go Schema Builder — Documented But Unverified

`08-go-schema-builder.md` describes:
```
internal/web/
├── amis/builder.go
├── amis/crud.go
├── amis/form.go
├── registry/registry.go
└── handler/schema.go
```

No evidence this exists in the codebase. If it doesn't exist, `/schema/` endpoints return ad-hoc `map[string]any` from handlers — or schemas are still static JSON files and `navigate()` is broken.

This is the most dangerous unknown in the audit. If schema endpoints don't exist, the entire shell is broken. If they exist as raw maps, the maintenance problem the builder was designed to prevent is already here.

---

### H6: No CSRF Protection

Cookie-based session auth + no CSRF token in fetcher headers. The documented middleware stack:

```
ResolveTenant → Authenticate → InjectFlags → RequirePermission → AuditWrap
```

No CSRF middleware. A malicious page can trigger state-mutating requests (POST/PUT/DELETE) using the victim's session cookie if `SameSite` is not `Strict`.

---

### H7: External CDN Dependency With No SRI

```html
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@fortawesome/fontawesome-free@6.4.0/css/all.min.css">
```

No `integrity` attribute. In enterprise ERP deployment: (1) offline environments break, (2) CDN compromise injects malicious CSS, (3) corporate proxies block CDNs. Font Awesome must be bundled locally in `web/sdk/`.

---

### H8: `alert(msg)` Native Browser Dialog in Production

**Location:** `index.html:1365`
```javascript
alert: function (msg) { alert(msg); }
```

Native `alert()` blocks the JavaScript thread. In an ERP this fires on AMIS-level errors. The user gets a modal that freezes all JavaScript until dismissed. AMIS has a built-in notification system — use it.

---

## MEDIUM SEVERITY — Scalability Traps

### M1: Single AMIS Instance With No Page Cache

Every hash navigation: unmounts current instance → destroys all state → fetches schema → re-initializes → fetches data. For a GL drilldown (5 levels deep), navigating back means re-fetching everything. For a finance manager switching between P&L, trial balance, and journal entries multiple times per session — slow and disorienting.

AMIS `app` with `keepAlive` or tab-based shell preserves component state across navigation. The current model intentionally destroys it.

---

### M2: Breadcrumb Breaks for Schema-Driven Detail Pages

```javascript
function updateBreadcrumb(route) {
  menuConfig.forEach(function(entry) { /* linear search */ });
  if (!pageName) pageName = route.charAt(0).toUpperCase() + route.slice(1);
}
```

Breadcrumb resolves from `menuConfig`. For Go-driven schemas at dynamic paths (`/schema/purchasing/orders/:id`), the lookup fails. `finance/invoices/abc-123-uuid` renders as breadcrumb "Finance/invoices/abc-123-uuid". Every detail page in the system has a broken breadcrumb.

---

### M3: `collapseSidebar()` on Every Main Content Click

**Location:** `index.html:1220`

```javascript
document.getElementById('main').addEventListener('click', function() {
  collapseSidebar();
});
```

Click a table row → sidebar collapses. Click a form field → sidebar collapses. Click an AMIS dialog backdrop → sidebar collapses. Fires on every interaction with every AMIS component. Productivity-destroying UX bug disguised as a feature.

---

### M4: Sidebar Search Is Cosmetic

```html
<div class="search-box">
  <i class="fa fa-search"></i>
  <span>Search tools</span>
</div>
```

No click handler. No event listener. No functionality. In an ERP with 100+ modules, nav search is the primary navigation mechanism at scale. You have built the UI affordance, deferred the implementation. When the menu has 50 items, users click it repeatedly and nothing happens.

---

### M5: Single-Open Accordion Is Hostile to ERP Cross-Module Workflows

Every group open collapses all others. A finance manager going Finance → Procurement → AP must perform 3+ accordion operations per workflow navigation. At 100 modules across 10+ groups, this navigation model is actively hostile to ERP users.

---

## AMIS Compliance Audit

### A1: AMIS `app` Component — The Right Tool, Not Used

The custom shell is ~1,400 lines solving a problem AMIS already solved. The docs correctly identify this and state the trigger: "migrate when nav needs Go-driving." **That trigger condition is already met.** `navigate()` already calls `/schema/` Go endpoints. `middleware.Authenticate` and `middleware.InjectFlags` already exist. The precondition is satisfied and the migration has not happened.

Every day this migration is deferred, the custom shell accumulates more features that are reinventions of AMIS `app` capabilities.

---

### A2: Permission Injection — Correct in Docs, Fragile in Practice

The correct pattern (from `27-decisions-and-opinions.md`):
```go
// Go injects { can_create: true } into page data
// Schema: visibleOn: "${can_create}"
```

The wrong pattern that will emerge without enforcement:
```json
"visibleOn": "${user.role === 'ADMIN'}"
```

This **will happen** without explicit code review enforcement. It is the path of least resistance for any developer who skips the decision docs. Once it appears in one schema, it spreads. You end up with permission logic split between Go handlers and JSON schemas, diverging over time.

---

### A3: `service` Component Likely Underused

Developers default to `crud` for everything. Dashboard KPI cards, summary panels, and notification feeds built with `crud` get unwanted pagination controls, filter panels, and column headers. This is the most common AMIS misuse pattern at scale.

---

### A4: `condition-builder` Converter Path Inconsistent

`24-known-issues.md` shows:
```go
// awo/web/converters/conditions.go
```

Module imports in `08-go-schema-builder.md` use `awo.so/internal/web/...`. Path `awo/web/converters/` is inconsistent with `awo.so/internal/web/`. Inconsistent package paths are the first symptom of convention drift in a multi-developer codebase.

---

## Go Backend Review

### G1: Backend Architecture Is the Strong Part

Hexagonal architecture, SQLC type-safe queries, Wire DI, `errors.As` unwrapping, the `parseTenantDBError → mapTenantError → ToHTTPError` error ladder — these are correct decisions. The backend architecture will survive scale. No major criticism based on available evidence.

---

### G2: Tenant Resolution — No `X-Tenant` Header in Fetcher

`ResolveTenant` middleware reads `X-Tenant` header or subdomain. The fetcher never sends `X-Tenant`. On `localhost:8080`, subdomain resolution fails. Multi-tenant local development is broken unless `ResolveTenant` has a fallback (e.g., session-bound tenant). Needs explicit verification and documentation.

---

### G3: Schema Handler Coupling Will Explode

Current pattern:
```go
schemaRoutes.Get("/purchasing/orders",     s.SchemaHandlers.PurchasingOrders)
schemaRoutes.Get("/purchasing/orders/new", s.SchemaHandlers.PurchasingOrdersNew)
schemaRoutes.Get("/purchasing/orders/:id", s.SchemaHandlers.PurchasingOrderDetail)
```

3 methods per module × 100 modules = 300 methods on `SchemaHandlers`. This is either a God struct or a fragmented handler hierarchy with its own DI complexity. The `init()`-based self-registration in `08-go-schema-builder.md` exists to prevent exactly this. If that pattern isn't being used, the handler struct is already the wrong direction.

---

## ERP Scalability Assessment

### E1: This Is CRUD, Not ERP — Yet

Current state: 4 functional pages. All list views. The system is a well-architected CRUD application with ERP documentation.

Real ERP features require AMIS patterns not yet visible in the implementation:
- Journal entry posting with debit/credit validation (`wizard`, formula expressions)
- Period management state machine (`visibleOn` + backend guards)
- GL drill-down reporting (5-level dialog chain with `data` scoping)
- Approval chains (`schemaApi` per entity type)
- Payroll computation (tabbed `form` with `initApi` per section)

**Risk:** Easy pages get built first (lists, forms). AMIS limits are discovered when implementing approval chains and drill-down reporting. Pressure to add custom JavaScript mounts. AMIS philosophy erodes feature by feature.

---

### E2: Approval Inbox Pattern — Correct

`11-page-patterns.md` Pattern 4:
```json
"schemaApi": "get:/schema/approvals/${record_type}/${record_id}"
```

This is AMIS working correctly. Go returns different schemas per entity type. Declarative, extensible, scales to 100 entity types without frontend changes. Execute this pattern exactly as documented.

---

### E3: Drill-Down Architecture — Dialog vs URL Navigation Trade-Off Not Resolved

Dialog-based drill-down: browser Back doesn't work. URL-based drill-down: bookmarkable state. For an ERP where finance managers share report links with auditors, bookmarkability matters. The docs identify this trade-off but don't mandate a choice. This will be decided inconsistently by individual developers — some modules dialog-based, others URL-based, no cross-module consistency.

---

## Long-Term Maintainability Predictions

### What Breaks First (6–12 months)

**The shell.** Every new module requires editing `menuConfig` in `index.html`. After 20 modules, `index.html` has 80 menu items in a JS array. After 50 modules, it is the most-edited file in the frontend. Merge conflicts on `menuConfig` become weekly. Feature flags for module visibility must be duplicated in JS. Permission-based nav hiding gets implemented ad-hoc, inconsistently, by individual developers.

### What Becomes Painful at Scale (12–24 months)

**Schema authoring without the builder.** If `internal/web/amis/` builder package does not exist, schemas are raw `map[string]any`. At 100 schemas with 20+ columns each, a field rename cascades to every schema that references it — no compile-time check, pure runtime discovery.

**The dual envelope.** Every new Go developer writes handlers and forgets which format to use. Code review catches it sometimes. The fetcher silently handles both. Incorrect responses show as empty tables, not errors. Debugging time compounds.

### What Causes a Rewrite Conversation (2–3 years)

**The custom shell vs AMIS `app`.** The shell accumulates 3,000+ lines of custom JavaScript solving problems AMIS already solved. When AMIS releases breaking changes in component APIs, custom shell is unaffected but all page schemas must be migrated. AMIS `app` absorbs framework-level changes automatically. The compounding maintenance cost of the custom shell eventually exceeds the one-time migration cost.

---

## Remediation Roadmap

### Priority 1 — Fix Production Blockers (This Week)

1. **Remove `API_BASE` hardcode.** Delete the variable. Use relative URLs in the fetcher.
2. **Add 401 → `/login` redirect** in the fetcher.
3. **Move Font Awesome to `web/sdk/`.** Remove CDN dependency.
4. **Fix breadcrumb XSS.** Replace `innerHTML` string interpolation with `textContent` assignments.

### Priority 2 — Structural Fixes (This Sprint)

5. **Standardize on AMIS envelope** `{status, data}` in all Go handlers. Remove bridge translation from fetcher once complete.
6. **Add CSRF middleware** to Fiber API routes.
7. **Remove `collapseSidebar()` main content click handler.**
8. **Replace native `alert()` in `amisEnv`** with AMIS notification action.
9. **Add `/schema` to Vite proxy** — or document explicitly that Vite is a build tool only.
10. **Fix `options: json.data`** in the list response bridge — remove or scope to select-only endpoints.

### Priority 3 — Architecture Migration (Next Quarter)

11. **Migrate to AMIS `app` with Go-driven nav.** The trigger condition is met. The custom shell is already at the point where Go-driving is needed. Migrate before module #10, not after module #50.
12. **Confirm or build `internal/web/amis/` builder package.** If it does not exist, build it before writing more schemas. Pays back at schema #5.
13. **Implement sidebar search.** Non-negotiable at ERP scale.
14. **Remove `max-height: 500px`** CSS ceiling on menu groups.
15. **Mandate drill-down navigation pattern** (dialog vs URL). One pattern across all modules. Document it in `27-decisions-and-opinions.md`.

### Priority 4 — ERP Completeness (Ongoing)

16. **Implement approval inbox** using `schemaApi` pattern exactly as documented. Highest-leverage ERP feature.
17. **Implement journal entry wizard** with debit/credit balance validation using AMIS formula expressions.
18. **Clarify Temporal usage.** Document active workflows or remove the dependency.
19. **Verify `ResolveTenant` behavior on localhost.** Document how multi-tenant dev works.

---

## Final Verdict

**Documentation quality:** Excellent. The doc set is unusually thorough. AMIS philosophy is understood and articulated correctly in `26-the-amis-way.md` and `27-decisions-and-opinions.md`.

**Implementation quality:** Below average for production. Three production-breaking issues (hardcoded `API_BASE`, missing 401 redirect, CDN dependency) should not exist at this stage. The gap between documented ERP scope and actual implementation scope is large enough to be a credibility risk.

**The architecture will survive to ERP scale if:** The AMIS `app` migration happens before module #10, the builder package exists and is used consistently, and the dual envelope is eliminated before production.

**The architecture will fail at scale if:** The custom shell continues accumulating custom JavaScript, the dual envelope persists, and each developer decides independently whether to inject permissions from Go or check them in schema expressions.

The bones are good. The build quality needs a serious pass before any claim of production readiness.
