# Troubleshooting

> Last verified: 2026-05-18 | Sources: audit findings, `internal/web/handler/schema.go`, `internal/web/stages/`

---

## Quick Diagnosis

| Symptom | Most Likely Cause | Jump To |
|---------|------------------|---------|
| 404 on `/schema/` route | Page not registered | [#page-not-found](#page-not-found) |
| Blank page, no error | Schema returns empty / compile panic | [#blank-page](#blank-page) |
| All buttons hidden for all users | Permission key missing from `AllUIPermissions` | [#buttons-always-hidden](#buttons-always-hidden) |
| Double `{status:0}` envelope | Page function returns `AmisResponse` | [#double-envelope](#double-envelope) |
| Dark mode: dropdowns still light | Missing portal CSS selectors | [#dark-mode-portals](#dark-mode-portals) |
| Cache stale after code change | Schema version not incremented | [#stale-cache](#stale-cache) |
| Schema served despite role change | Cache key collision (old entry TTL not expired) | [#stale-cache](#stale-cache) |
| Permission denied page for valid user | Wrong permission key format | [#permission-key-format](#permission-key-format) |
| Table always empty | Fetcher envelope mismatch | [#empty-table](#empty-table) |
| Compile error: `UIContext` undefined | Old type name — use `UISessionContext` | [#old-type-names](#old-type-names) |
| Panic at startup | Duplicate route or invalid PageRegistration | [#startup-panic](#startup-panic) |
| 503 on schema request | Casbin enforcer down | [#casbin-down](#casbin-down) |

---

## Page Not Found {#page-not-found}

**Symptom:** `GET /schema/finance/invoices` returns HTTP 404.

**Cause:** Route not registered in the page registry.

**Diagnose:**
```go
// RegistryStage logs: "page not registered for route /finance/invoices"
// SchemaHandler logs: "page not found" warn
```

**Fix:** Add `registry.RegisterPage` call in `init()`:
```go
func init() {
    registry.RegisterPage(registry.PageRegistration{
        Route:  "/finance/invoices",
        Module: "finance",
        Title:  "Invoices",
        ASTFn:  Schema,
    })
}
```

Import the package blank in your wire/routes setup:
```go
import _ "awo.so/internal/web/pages/finance/invoices"
```

**Common mistakes:**
- Missing `init()` entirely
- Package imported but `init()` uses `registry.Register()` (old API) not `registry.RegisterPage()`
- Route has trailing slash: `/finance/invoices/` → registry expects `/finance/invoices`
- Package not blank-imported → `init()` never runs

See [Page Registration Pattern](../03-implementation/02-page-registration-pattern.md).

---

## Blank Page {#blank-page}

**Symptom:** Page renders nothing. No error in browser. HTTP 200 returned.

**Causes + fixes:**

**A: Page function returns empty schema**
```go
// WRONG
func Schema(sess ui.UISessionContext) ui.Schema {
    return ui.Schema{} // empty map → ResponseStage returns error
}

// CORRECT
func Schema(sess ui.UISessionContext) ui.Schema {
    return ui.Schema{"type": "page", "body": ...}
}
```

**B: Page function panics**
CompileStage catches panics and returns `ErrSchemaInvalid`. Check server logs for:
```
schema validation failed: panic in page function: ...
```

**C: Schema fails validation**
`ValidateStage` returns `ErrSchemaInvalid` → HTTP 500 → AMIS shows error.
Check logs for: `"schema validation failed"` with the specific rule that failed.

**D: Cache hit serving old empty schema**
Delete the cache entry: `cacheSvc.DeletePattern(ctx, "ui:schema:{tenantID}:{route}:*")`

---

## Buttons Always Hidden {#buttons-always-hidden}

**Symptom:** `visibleOn: "${can_create}"` never shows the button for any user.

**Root cause:** Permission not in `AllUIPermissions` → `BulkEnforce` never evaluates it → `Can()` always returns false.

**Diagnose:**
```go
// internal/web/authz/service.go
var AllUIPermissions = []UIPermission{
    // Is "invoice.create" in this list?
}
```

**Fix:**
1. Add the permission to `AllUIPermissions`:
   ```go
   "invoice.create", // ADD THIS
   ```
2. Verify the permission key format: `resource.action` (e.g. `"invoice.create"`, not `"create:invoice"`)
3. Verify the page data sets the variable:
   ```go
   "data": ui.M{"can_create": sess.Can("create", "invoice")},
   ```
4. Verify AMIS expression matches variable name: `"visibleOn": "${can_create}"`

**Secondary cause:** Casbin policy does not grant the permission to the role. Verify policy file grants `invoice.create` to the expected role.

---

## Double Envelope {#double-envelope}

**Symptom:** Browser receives `{"status":0,"data":{"status":0,"data":{"type":"page",...}}}`.

**Cause:** Page function returns `AmisResponse` or a map with `status/data` keys.

**Fix:** Page functions return raw schema — SchemaHandler wraps it:
```go
// WRONG
func Schema(sess ui.UISessionContext) ui.Schema {
    return ui.Schema{
        "status": 0,
        "data": ui.Schema{"type": "page", ...},
    }
}

// CORRECT
func Schema(sess ui.UISessionContext) ui.Schema {
    return ui.Schema{
        "type": "page",
        ...
    }
}
```

See [API Contracts](02-api-contracts.md#schema-endpoint-response).

---

## Dark Mode: Portals Still Light {#dark-mode-portals}

**Symptom:** Most of the page is dark but dropdowns, date pickers, or modals appear light.

**Cause:** AMIS portals append to `<body>` outside `#content`. CSS variables cascade but some components need explicit selectors.

**Fix:** Ensure `web/pages/index.html` has Layer 3 portal selectors:
```css
html.dark .cxd-PopOver        { background: var(--PopOver-bg); }
html.dark .cxd-Modal-content  { background: var(--Modal-bg); }
html.dark .cxd-Drawer-content { background: var(--Drawer-bg); }
html.dark .cxd-Select-menu    { background: var(--Select-menu-bg); }
html.dark .cxd-DropDown-menu  { background: var(--DropDown-menu-bg); }
html.dark .cxd-DatePicker-popover { background: var(--DatePicker-bg); }
html.dark .cxd-Tooltip-body   { background: var(--Tooltip-bg); color: var(--Tooltip-color); }
```

See [Dark Mode → Layer 3](../05-dark-mode.md#layer-3--portal-reinforcement).

---

## Stale Cache {#stale-cache}

**Symptom:** Code change deployed but old schema still served. Or: role changed but old schema still served.

**Normal TTL expiry:** Schema cache TTL is 5 minutes. Wait 5 minutes or force-invalidate.

**Force invalidation:**
```go
// Invalidate all schemas for tenant
cacheSvc.DeletePattern(ctx, "ui:schema:{tenantID}:*")

// Invalidate specific route for all users
cacheSvc.DeletePattern(ctx, "ui:schema:{tenantID}:/finance/invoices:*")
```

**For deployments:** Bump `SchemaGeneration` in `CacheVersions`. This changes the cache key suffix → all existing entries become unreachable:
```go
uicache.CacheVersions{
    SchemaGeneration: 2, // was 1 — bump on each deployment
}
```

**Role-change automatic invalidation:** Permission changes change `permFP` → different cache key → automatic miss on next request. No action needed for permission changes.

**Stale flag behavior:** Feature flags are session-scoped (snapshot at login). User must re-authenticate to pick up flag changes.

---

## Permission Key Format {#permission-key-format}

**Symptom:** `sess.Can("create", "invoice")` always returns false even though user has the policy.

**Correct format:** `resource.action` (resource first, dot, action)

```go
// UISessionContext.Can internals:
func (u UISessionContext) Can(action, resource string) bool {
    return u.permissions[resource+"."+action]
    // Can("create", "invoice") → checks "invoice.create"
}
```

**Check: Is the permission in AllUIPermissions?**
```go
// internal/web/authz/service.go
var AllUIPermissions = []UIPermission{
    "invoice.create", // must be here
}
```

**Check: Does the Casbin policy grant it?**
The Casbin policy must have an entry granting `invoice.create` to the user's role.

**Check: Is the variable name set in page data?**
```go
"data": ui.M{
    "can_create": sess.Can("create", "invoice"), // ← must be set
}
```

**Check: Does the AMIS expression reference the right variable?**
```json
"visibleOn": "${can_create}"   ✓
"visibleOn": "${canCreate}"    ✗ (camelCase doesn't match snake_case key)
```

---

## Empty Table {#empty-table}

**Symptom:** CRUD table renders but shows no rows. No network error.

**Cause A: Wrong response format from Go handler**
AMIS `crud` expects:
```json
{ "status": 0, "data": { "items": [...], "count": 142 } }
```
Not `{ "success": true, "data": [...] }`. The fetcher translates the Go format automatically, but only if `json.data` is an array with `json.meta.pagination`.

**Cause B: Go handler returning array at wrong path**
```go
// WRONG
return c.JSON(response.OK(orders))
// Produces: { "status": 0, "data": [...] }  — data is array, not { items, count }

// CORRECT
return c.JSON(response.OK(map[string]any{
    "items": orders,
    "count": total,
}))
```

**Cause C: API string wrong in schema**
```go
// WRONG — missing method prefix
"api": "/api/v1/finance/invoices"

// CORRECT
"api": "get:/api/v1/finance/invoices"
```

**Cause D: CRUD missing `syncLocation: true`**
NormalizeStage sets this but if the schema bypasses normalize (e.g. returned from cache with old entry), check manually.

---

## Old Type Names {#old-type-names}

**Symptom:** Compile error `undefined: ui.UIContext` or `undefined: amis.Ctx` in new code.

`UIContext` does not exist. Use `UISessionContext`:
```go
// WRONG (old docs)
func Schema(ctx amis.Ctx) amis.Schema { ... }

// CORRECT (current)
func Schema(sess ui.UISessionContext) ui.Schema { ... }
// or for ASTPageFn:
func Schema(sess ui.UISessionContext) any { ... }
```

See [Glossary → UIContext](../appendices/A-glossary.md#uicontext-does-not-exist).

---

## Startup Panic {#startup-panic}

**Symptom:** Server panics at startup with message about duplicate route or invalid registration.

`registry.ValidateRegistry()` is called at startup and panics on:

| Panic message | Fix |
|--------------|-----|
| `duplicate route: /finance/invoices` | Two `RegisterPage` calls with same route — remove one |
| `route must start with /` | Change `"finance/invoices"` → `"/finance/invoices"` |
| `route must not end with /` | Remove trailing slash |
| `module must not be empty` | Add `Module: "finance"` to registration |
| `title must not be empty` | Add `Title: "Invoices"` to registration |
| `must set Fn or ASTFn` | Add either `Fn:` or `ASTFn:` to registration |

---

## Casbin Enforcer Down {#casbin-down}

**Symptom:** Schema endpoints return HTTP 503 `"service temporarily unavailable"`.

**Cause:** `UIAuthzService.BulkEnforce()` returned an error → `AuthzStage` wraps as `ErrPermissionResolution` → SchemaHandler returns 503.

**Diagnose:** Check server logs for:
```
IAM permission resolution failed: ...
```

**Fix:** Restore Casbin enforcer connectivity. The UI pipeline has no fallback for authz failures — serving a schema without resolved permissions would be a security defect.

---

## Feature Flag Not Taking Effect

**Symptom:** Flag was toggled in admin UI but page still shows/hides wrong elements.

**Cause:** Feature flags are **snapshotted at login**. The user's current session has the old flag value.

**Fix:** User must log out and log back in. After re-authentication, `NewUISessionContext` snapshots the new flag value.

**For testing:** Invalidate the session and re-authenticate. There is no per-request flag re-evaluation.

---

## `API_BASE` Hardcoded (Production Issue)

**Symptom:** All API calls fail with CORS errors in non-localhost deployment.

**Location:** `web/pages/index.html:1040`
```javascript
var API_BASE = 'http://localhost:8080/'; // ← REMOVE THIS
```

**Fix:** Delete `API_BASE` entirely. Use relative URLs in the fetcher. AMIS fetcher defaults to same-origin relative paths.

This is a tracked critical issue from the 2026-05-17 architecture audit.

---

## Sidebar Collapses on Table Row Click

**Symptom:** Clicking any AMIS component (table row, form field, button) collapses the sidebar.

**Location:** `web/pages/index.html:1220`
```javascript
document.getElementById('main').addEventListener('click', function() {
    collapseSidebar(); // fires on EVERY click in main area
});
```

**Fix:** Remove this event listener. It was intended for mobile overlay dismissal but fires universally.
