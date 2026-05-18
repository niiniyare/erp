# IAM Integration

> Last verified: 2026-05-18 | Code pointer: `internal/core/iam/contract/`, `internal/web/authz/`, `internal/web/stages/authz.go`

---

## 📖 Why This Matters

Every page in this ERP is personal. Whether a "Delete" button appears, whether a module is visible, whether a table shows restricted columns — all of it depends on who is asking. IAM integration is what connects "who asked?" to "what do they see?". Understanding this layer is essential before writing any page function that involves conditional rendering.

---

## Two Distinct Types — Do Not Confuse Them

The most common source of bugs in the UI layer. There are **two context types** in use:

| | `contract.SessionContext` | `ui.UISessionContext` |
|--|--|--|
| Package | `internal/core/iam/contract` | `internal/web/ui` |
| Who creates it | `contract.InjectSessionContext` middleware | `AuthzStage.Execute()` only |
| What it carries | Identity, feature flags, preferences, tenant settings | Identity + pre-resolved permission map + flag snapshot |
| Can() method | **No** — deliberately removed | **Yes** — reads pre-resolved map |
| Crosses module boundary | Yes — used by all non-IAM modules | No — UI pipeline only |
| When available | After `InjectSessionContext` middleware runs | After `AuthzStage` runs (Priority 20) |

**Rule: Page functions (`PageFn`, `ASTPageFn`) receive `UISessionContext`. They never see `contract.SessionContext`.**

---

## Data Flow Diagram

```
HTTP Request
     │
     ▼
Authenticate middleware
     │ validates session token
     ▼
contract.InjectSessionContext middleware
     │ wraps *iam.ResolvedSession → contract.SessionContext
     │ stores in Go context via contract.WithContext()
     ▼
SchemaHandler.Handle()
     │ reads contract.SessionContext via contract.FromContext(ctx)
     │ starts pipeline.Run()
     ▼
┌─────────────────────────────────────────────────────────┐
│ Pipeline                                                 │
│                                                          │
│  SessionStage (P=10)                                     │
│    contract.FromContext(opCtx.Ctx) → validates non-zero  │
│    ─────────────────────────────────────────────────     │
│  AuthzStage (P=20)                                       │
│    contract.FromContext(opCtx.Ctx) → sc                  │
│    UIAuthzService.BulkEnforce(ctx, sc) → map[string]bool │
│    authz.PermissionFingerprint(perms) → permFP           │
│    authz.FlagFingerprint(sc, UIFlagList) → flagFP        │
│    ui.NewUISessionContext(sc, perms) → UISessionContext   │
│    ─────────────────────────────────────────────────     │
│  ... (CacheLookup, Registry, Compile ...)                │
│    PageFn(UISessionContext) → Schema                     │
└─────────────────────────────────────────────────────────┘
```

---

## ⚙️ `contract.SessionContext` — What It Exposes

Source: `internal/core/iam/contract/session.go`

```go
// Identity
sc.UserID()      uuid.UUID   // authenticated user
sc.TenantID()    uuid.UUID   // RLS boundary
sc.DisplayName() string

// User type (safe for UI context decisions — NOT a permission check)
sc.IsPlatform()  bool  // platform admin actor
sc.IsPortal()    bool  // external portal user

// Feature flags — snapshotted AT LOGIN, not re-evaluated per request
sc.FeatureEnabled("bulk_import") bool

// Tenant settings — snapshotted at login
sc.Setting("tenant.currency", "USD")    string
sc.SettingBool("feature.approvals", false) bool
sc.SettingInt("limits.users", 10)        int

// User preferences — snapshotted at login
sc.Preference("ui.theme", "light")    string

// Entity visibility scope (for service-layer WHERE clauses)
sc.EntityScope() iam.EntityScope

// Go context helpers
contract.WithContext(ctx, sc) context.Context
contract.FromContext(ctx) (SessionContext, bool)
```

**What it deliberately does NOT expose:**
- `Can()` / `CanDo()` — removed. All authz via Casbin.
- `ToPrincipal()` — IAM internal only.
- Raw permission maps.

---

## ⚙️ `UIAuthzService` — The Only Casbin Boundary

Source: `internal/web/authz/service.go`

```go
type UIAuthzService interface {
    BulkEnforce(ctx context.Context, sc contract.SessionContext) (map[string]bool, error)
}
```

`BulkEnforce` evaluates **all** permissions in `AllUIPermissions` for the session in one Casbin batch call. The result is a `map[string]bool` where keys are `"resource.action"` strings:

```go
// Example result of BulkEnforce for a finance clerk role:
map[string]bool{
    "invoice.read":    true,
    "invoice.create":  true,
    "invoice.approve": false,
    "invoice.delete":  false,
    "account.read":    true,
    "account.create":  false,
    // ... all AllUIPermissions evaluated
}
```

**`authz` package is the ONLY place in `internal/web/` that may import Casbin.** All other UI code receives the pre-resolved map.

---

## ⚙️ `AllUIPermissions` — The Registry

Source: `internal/web/authz/service.go:34`

Every permission a page function might check **must be listed here**:

```go
var AllUIPermissions = []UIPermission{
    "invoice.read",
    "invoice.create",
    "invoice.update",
    "invoice.delete",
    "invoice.approve",
    "invoice.void",
    "account.read",
    // ... ~35 permissions total
}
```

**Critical rule:** If a permission is not in `AllUIPermissions`, `BulkEnforce` never evaluates it, and `UISessionContext.Can()` **always returns false** for it — even if the user has the policy. This is the most common cause of "buttons always hidden" bugs.

```go
// WRONG — "report.export" not in AllUIPermissions
if sess.Can("export", "report") { // always false
    ...
}

// FIX — add to AllUIPermissions first, then use
"report.export", // in AllUIPermissions
// now sess.Can("export", "report") works correctly
```

---

## ⚙️ `UISessionContext` — What Page Functions Receive

Source: `internal/web/ui/types.go`

```go
// Constructed ONLY by AuthzStage.Execute(). Never construct directly.
type UISessionContext struct {
    // Public identity fields
    UserID      string
    TenantID    string
    DisplayName string
    IsPlatform  bool   // from contract.SessionContext.IsPlatform()
    IsPortal    bool   // from contract.SessionContext.IsPortal()

    // Locale (from preferences)
    Locale   string // e.g. "en-GB"
    Timezone string // e.g. "Africa/Nairobi"
    Currency string // e.g. "KES"

    // Private — access via methods only
    // permissions  map[string]bool  ← from BulkEnforce
    // featureFlags map[string]bool  ← snapshotted from contract.SessionContext
    // prefs        map[string]string
}

// Permission check — key format is "resource.action"
func (u UISessionContext) Can(action, resource string) bool {
    return u.permissions[resource+"."+action]
    // Can("create", "invoice") → checks "invoice.create"
}

// Convenience helpers
func (u UISessionContext) CanAny(action string, resources ...string) bool
func (u UISessionContext) CanAll(action string, resources ...string) bool

// Feature flag — snapshotted at login, NOT re-evaluated per request
func (u UISessionContext) Flag(name string) bool

// User preference
func (u UISessionContext) Pref(key, fallback string) string
```

---

## ⚙️ Using `UISessionContext` in Page Functions

```go
func Schema(sess ui.UISessionContext) ui.Schema {
    return amis.M{
        "type":  "page",
        "title": "Invoices",
        "data": amis.M{
            // Expose as template variables for AMIS visibleOn/disabledOn
            "can_create":  sess.Can("create", "invoice"),
            "can_approve": sess.Can("approve", "invoice"),
            "can_delete":  sess.Can("delete", "invoice"),
            "is_platform": sess.IsPlatform,
        },
        "body": amis.M{
            "type":       "crud",
            "api":        "get:/api/v1/finance/invoices",
            "syncLocation": true,
            "headerToolbar": []any{
                amis.M{
                    "type":      "button",
                    "label":     "New Invoice",
                    "level":     "primary",
                    "visibleOn": "${can_create}",
                },
            },
        },
    }
}
```

### Using `IsPlatform` / `IsPortal` for layout branching

```go
func Schema(sess ui.UISessionContext) ui.Schema {
    title := "Invoices"
    if sess.IsPlatform {
        title = "All Tenant Invoices" // platform sees cross-tenant view
    }
    // ...
}
```

### Using feature flags

```go
func Schema(sess ui.UISessionContext) ui.Schema {
    body := []any{invoiceTable()}
    if sess.Flag("ai_assist") {
        body = append(body, aiSuggestPanel())
    }
    return amis.M{"type": "page", "body": body}
}
```

---

## Permission Key Format

**Format: `resource.action`** (lowercase, dot-separated)

```go
// These are equivalent:
sess.Can("read",   "invoice")  // checks "invoice.read"
sess.Can("create", "invoice")  // checks "invoice.create"
sess.Can("approve","invoice")  // checks "invoice.approve"

// Multi-resource helpers
sess.CanAny("read", "invoice", "account")  // true if either readable
sess.CanAll("read", "invoice", "account")  // true only if both readable
```

**Common mistake:** Old docs showed `:` as separator (`"create:invoice"`). The separator is `.`. Using the wrong format results in a map key miss → `Can()` returns `false` for all users.

---

## Feature Flag Semantics

Feature flags are **snapshotted at login** into `contract.SessionContext.Configuration`. They are copied into `UISessionContext` by `NewUISessionContext` at `AuthzStage` time.

**Implication:** If a flag is toggled for a tenant while a user is logged in, the user sees the **old** flag state until their session expires or they re-authenticate. There is no live flag re-evaluation per request.

Known UI flags (source: `internal/web/ui/types.go:176`):

| Flag | Controls |
|------|---------|
| `advanced_reporting` | Enhanced report builder |
| `bulk_import` | CSV bulk import buttons |
| `multi_currency` | Multi-currency fields |
| `approval_workflow` | Approval inbox pages |
| `ai_assist` | AI suggestion panels |
| `billing.autopay` | Autopay settings |
| `hr.payroll_v2` | New payroll UI |
| `inventory.lot_tracking` | Lot tracking columns |
| `finance.auto_reconcile` | Auto-reconcile controls |

---

## Debugging

### "Button always hidden for all users"

1. Confirm the permission is in `AllUIPermissions` (`internal/web/authz/service.go:34`)
2. Confirm the permission string format is `resource.action` (e.g. `"invoice.create"`, not `"create:invoice"`)
3. Confirm the page data sets the variable: `"can_create": sess.Can("create", "invoice")`
4. Confirm the AMIS expression references the correct variable: `"visibleOn": "${can_create}"`

### "Permission never evaluates correctly"

Check `AuthzStage` output in logs. It logs:
```
resolved 35 permissions for user <uuid> (permFP=a1b2c3d4 flagFP=e5f6g7h8)
```
If `AuthzStage` aborts, check `SessionStage` first — the session context must be present before `AuthzStage` runs.

### "Feature flag changes not taking effect"

Flags are session-scoped. Ask the user to log out and back in. Session TTL determines when new flag values take effect.
