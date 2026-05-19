# Authorization Patterns

> Last verified: 2026-05-18 | Code pointers: `internal/web/stages/authz.go`, `internal/web/authz/service.go`, `internal/web/ui/types.go`

---

## How Authorization Works

`AuthzStage` (P=20) resolves all permissions before any page function runs. Page functions receive a `UISessionContext` with pre-resolved permissions — they never call IAM directly.

```
Request
  └─ SessionStage (P=10)   — validates session
  └─ AuthzStage (P=20)     — BulkEnforce → UISessionContext.permissions map
  └─ CacheLookupStage (P=30)
  └─ RegistryStage (P=40)
  └─ CompileStage (P=50)   — calls PageFn/ASTPageFn with UISessionContext
       └─ sess.Can("read", "invoice")  → reads pre-resolved map, no IAM call
```

**Key invariant:** `sess.Can()` is a pure map lookup. Zero latency, zero I/O. All IAM work happened at P=20.

---

## `sess.Can(action, resource)`

```go
// internal/web/ui/types.go
func (u UISessionContext) Can(action, resource string) bool {
    return u.permissions[resource+"."+action]
}
```

Permission key format: **`resource.action`** (resource first, dot-separated).

```go
sess.Can("read",   "invoice")   // checks "invoice.read"
sess.Can("create", "invoice")   // checks "invoice.create"
sess.Can("approve", "invoice")  // checks "invoice.approve"
```

**Common mistake:** reversed order `Can("invoice", "read")` → checks `"read.invoice"` → always false.

---

## AllUIPermissions Registry

Every permission used in any `sess.Can()` call **must** be listed in `AllUIPermissions`:

```go
// internal/web/authz/service.go
var AllUIPermissions = []string{
    "invoice.read",
    "invoice.create",
    "invoice.update",
    "invoice.delete",
    "invoice.approve",
    "invoice.export",
    // ... ~35 total
}
```

`BulkEnforce` only evaluates permissions in this list. If a permission is not listed here, `Can()` always returns `false` — silently, no error.

**When adding a new permission:**
1. Add `"resource.action"` to `AllUIPermissions`
2. Add the corresponding Casbin policy
3. Then use `sess.Can("action", "resource")` in page functions

---

## Pattern 1: AMIS Expression Gate (preferred for element-level)

Expose permissions as boolean variables in page `Data`. AMIS evaluates them client-side per render.

```go
func Schema(sess ui.UISessionContext) any {
    return ast.PageNode{
        Data: ui.M{
            "can_create": sess.Can("create", "invoice"),
            "can_edit":   sess.Can("update", "invoice"),
            "can_delete": sess.Can("delete", "invoice"),
            "can_approve": sess.Can("approve", "invoice"),
        },
        Body: ast.CRUDNode{
            // ...
            Toolbar: []ast.Node{
                ast.ActionNode{
                    Label:     "New Invoice",
                    Level:     "primary",
                    VisibleOn: "${can_create}",   // ← hidden if false
                },
            },
            Columns: []ast.TableColumn{
                // ...
                {
                    Type:  "operation",
                    Label: "Actions",
                    Buttons: []ast.ActionNode{
                        {Label: "Edit",   VisibleOn: "${can_edit}",   ActionType: "link", Target: "#invoices/${id}/edit"},
                        {Label: "Delete", VisibleOn: "${can_delete}", Level: "danger", ActionType: "ajax",
                            API: &ast.APISpec{Method: "delete", URL: "/api/v1/finance/invoices/${id}"},
                            ConfirmText: "Delete this invoice?"},
                    },
                },
            },
        },
    }
}
```

**When to use:** Buttons, columns, form fields — anything that appears/disappears per permission within a stable layout.

**Cache note:** Schema is cached per `permFP` (permission fingerprint). Two users with different permissions get different cached schemas. Pattern 1 still benefits from caching — same fingerprint = same schema.

---

## Pattern 2: Structural Go Gate (for entire sections)

Conditionally include or exclude entire nodes in Go before schema is compiled.

```go
func Schema(sess ui.UISessionContext) any {
    body := []ast.Node{invoiceTable(sess)}

    // Approval inbox: only for users who can approve
    if sess.Can("approve", "invoice") {
        body = append(body, approvalInboxSection(sess))
    }

    // Platform-only cross-tenant summary
    if sess.IsPlatform {
        body = append(body, crossTenantSummarySection(sess))
    }

    return ast.PageNode{
        Title: "Invoices",
        Body:  ast.FlexNode{Direction: "column", Items: body},
    }
}
```

**When to use:**
- Entire sections only certain roles ever see (avoid inflating schemas for most users)
- Complex conditional layouts where AMIS expressions would be unreadable
- `IsPlatform` gating (platform admin vs portal user)

**When NOT to use:**
- Simple button visibility — use Pattern 1
- Cases where the layout is the same but a field is disabled — use `disabledOn`

---

## Pattern 3: `DisabledOn` (read-only fields)

Field is visible but disabled. User can see the value, not change it.

```go
ast.InputTextNode{
    Name:       "ref_number",
    Label:      "Reference #",
    DisabledOn: "${!can_edit}",   // disabled when user cannot edit
}
```

Or in Go:
```go
readOnly := !sess.Can("update", "invoice")

ast.InputTextNode{
    Name:      "ref_number",
    Label:     "Reference #",
    DisabledOn: boolExpr(readOnly),  // "true" or ""
}
```

`boolExpr(true)` returns `"true"` (always disabled). `boolExpr(false)` returns `""` (never disabled).

**When to use:** Forms where view role sees all fields but edit role can change them. Prevents two separate view/edit page schemas.

---

## Pattern 4: Mode Gate (`IsPlatform` / `IsPortal`)

```go
func Schema(sess ui.UISessionContext) any {
    return ast.PageNode{
        Data: ui.M{
            "is_platform": sess.IsPlatform,
        },
        Body: ast.TabsNode{
            Tabs: []ast.TabItem{
                {Title: "Overview",     Body: overviewTab(sess)},
                {Title: "All Tenants",  Body: tenantListTab(sess), VisibleOn: "${is_platform}"},
                {Title: "Audit Log",    Body: auditTab(sess),      VisibleOn: "${is_platform}"},
            },
        },
    }
}
```

`sess.IsPlatform` is set by `AuthzStage` from `contract.SessionContext.IsPlatform()`. No separate permission needed.

---

## Combining Patterns

Typical production page uses all three:

```go
func Schema(sess ui.UISessionContext) any {
    readOnly := !sess.Can("update", "invoice")

    var body []ast.Node
    body = append(body, documentForm(sess, readOnly))    // Pattern 3 inside

    if sess.Can("approve", "invoice") {                  // Pattern 2
        body = append(body, approvalPanel(sess))
    }

    return ast.PageNode{
        Data: ui.M{
            "can_edit":   !readOnly,                     // Pattern 1
            "can_delete": sess.Can("delete", "invoice"),
            "can_void":   sess.Can("void", "invoice"),
        },
        Body: ast.GridNode{
            Columns: []ast.GridColumn{
                {MD: 8, Body: body},
                {MD: 4, Body: []ast.Node{
                    blocks.TotalsSummaryBlock(sess),
                    blocks.ApprovalWorkflowBlock(sess), // internally gates on feature flag
                }},
            },
        },
    }
}
```

---

## Feature Flag Gating

Feature flags are on `contract.SessionContext`, not `UISessionContext`. Access via `sess.FeatureEnabled`:

```go
// UISessionContext exposes flags snapshotted at login
if sess.FeatureEnabled("approval_workflow") {
    body = append(body, approvalPanel(sess))
}
```

**Flags are snapshotted at login.** User must re-authenticate to see flag changes. Not re-evaluated per request.

`ApprovalWorkflowBlock` checks `approval_workflow` internally — safe to always include; renders collapsed placeholder when flag is off.

---

## Debugging: Permission Always False

**Symptom:** Button always hidden, even for admin user.

**Checklist:**

1. **Check `AllUIPermissions`** — is `"resource.action"` in the list?
   ```go
   // internal/web/authz/service.go
   var AllUIPermissions = []string{ ... }
   ```
   If missing, add it. `BulkEnforce` skips unlisted permissions.

2. **Check key format** — `Can(action, resource)` → key is `resource.action`:
   ```go
   sess.Can("read", "invoice")  // key: "invoice.read"   ✓
   sess.Can("invoice", "read")  // key: "read.invoice"   ✗ — always false
   ```

3. **Check Casbin policy** — user's role must have the permission assigned.

4. **Check permission fingerprint** — if cache hit, the cached schema was built with the old permission state. Bump `SchemaGeneration` or wait for TTL.

5. **Verify `Data:` includes the variable** — if using Pattern 1, the variable must be in `Data:` before the `VisibleOn` expression can read it:
   ```go
   Data: ui.M{
       "can_create": sess.Can("create", "invoice"),  // must be here
   }
   // Then in schema:
   VisibleOn: "${can_create}"
   ```

---

## Quick Reference

| Scenario | Pattern | Code |
|---|---|---|
| Button show/hide | Pattern 1 | `VisibleOn: "${can_create}"` + `Data: M{"can_create": sess.Can(...)}` |
| Field read-only | Pattern 3 | `DisabledOn: "${!can_edit}"` |
| Entire section excluded | Pattern 2 | `if sess.Can(...) { body = append(...) }` |
| Platform-only tab | Pattern 4 + 1 | `VisibleOn: "${is_platform}"` + `Data: M{"is_platform": sess.IsPlatform}` |
| Feature-gated section | Flag check in Go | `if sess.FeatureEnabled("flag") { ... }` |
| Column hidden for some roles | Pattern 1 | `VisibleOn` on column node |
