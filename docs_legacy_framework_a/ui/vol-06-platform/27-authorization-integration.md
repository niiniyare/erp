> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
chapter: 27
title: "Authorization Integration"
volume: "vol-06-platform"
section: "Platform"
description: "BulkEnforce, sess.Can/CanAny/CanAll, permission key format, permission-aware block patterns, DataTableBlock permissions, and ActionNode disable expressions."
status: implemented
---

# Chapter 27 — Authorization Integration

## Table of Contents

- [27.1 UIAuthzService.BulkEnforce](#271-uiauthzservicebulkenforce)
- [27.2 sess.Can — Reading Pre-Computed Permissions](#272-sesscan)
- [27.3 sess.CanAny and sess.CanAll](#273-sesscanyand-sesscanalll)
- [27.4 Permission Key Format](#274-permission-key-format)
- [27.5 Permission-Aware Block Pattern](#275-permission-aware-block-pattern)
- [27.6 DataTableBlock.CreatePermission](#276-datatableblock-createpermission)
- [27.7 ActionNode.DisabledOn](#277-actionnode-disabledon)

---

## 27.1 UIAuthzService.BulkEnforce

`UIAuthzService` is an interface with a single method. It is the only point where the UI platform calls the Casbin/IAM backend.

```go
type UIAuthzService interface {
    BulkEnforce(
        ctx       context.Context,
        userID    string,
        tenantID  string,
        permissions []string,
    ) (map[string]bool, error)
}
```

`BulkEnforce` is called once per request by `AuthzStage`. It receives the full list of all UI-known permission strings and returns a map of which ones the current user holds.

```go
// allUIPermissions is the static list defined in types.go
perms, err := svc.BulkEnforce(ctx, sess.UserID(), sess.TenantID(), allUIPermissions)
// Result:
// {
//   "finance.invoices.create": true,
//   "finance.invoices.read":   true,
//   "finance.invoices.delete": false,
//   "hr.employees.read":       true,
//   "hr.payroll.run":          false,
//   ...
// }
```

This single call pays the full cost of IAM evaluation once. All subsequent permission checks in the pipeline are free map lookups.

If `BulkEnforce` returns an error, the request fails with 503 — not 403. The distinction matters: 503 means "we don't know what permissions the user has," while 403 means "we know, and the answer is no."

---

## 27.2 sess.Can

`sess.Can(action, resource)` reads the pre-computed permission map. It never calls any external service.

```go
func (s *UISessionContext) Can(action, resource string) bool {
    key := resource + "." + action   // "finance.invoices" + "." + "create"
    return s.perms[key]              // pure map lookup
}
```

Usage in a block function:

```go
if sess.Can("create", "finance.invoices") {
    nodes = append(nodes, NewInvoiceButton())
}

if sess.Can("delete", "finance.invoices") {
    nodes = append(nodes, DeleteButton())
}

if sess.Can("export", "finance.reports") {
    nodes = append(nodes, ExportButton())
}
```

If the permission key is not present in the map (i.e., it was not in `allUIPermissions` when the registry was built), `Can` returns `false` — not an error. New permissions must be added to the `allUIPermissions` list in `types.go` to take effect.

---

## 27.3 sess.CanAny and sess.CanAll

Two multi-resource helpers cover common patterns.

### CanAny — OR Check

Returns `true` if the user has the given action on **at least one** of the listed resources.

```go
func (s *UISessionContext) CanAny(action string, resources ...string) bool
```

Use case: show a "Reports" section if the user can read any financial report.

```go
if sess.CanAny("read", "finance.invoices", "finance.reports", "finance.receipts") {
    sections = append(sections, ReportsSectionBlock(sess))
}
```

### CanAll — AND Check

Returns `true` only if the user has the given action on **all** listed resources.

```go
func (s *UISessionContext) CanAll(action string, resources ...string) bool
```

Use case: show a "Reconcile" button only if the user can both read invoices and manage accounts.

```go
if sess.CanAll("manage", "finance.invoices", "finance.accounts") {
    nodes = append(nodes, ReconcileButton())
}
```

Both methods use the same pre-computed map — no extra cost for checking multiple resources.

---

## 27.4 Permission Key Format

Permission strings follow a three-part dotted format:

```
<module>.<resource>.<action>
```

| Part       | Example values                    |
|------------|-----------------------------------|
| `module`   | `finance`, `hr`, `inventory`, `settings` |
| `resource` | `invoices`, `employees`, `payroll`, `items` |
| `action`   | `create`, `read`, `update`, `delete`, `export`, `manage`, `run` |

### Mapping to sess.Can

`sess.Can(action, resource)` assembles the key as `resource + "." + action`:

```go
sess.Can("create", "finance.invoices")
// internally looks up: "finance.invoices.create"

sess.Can("read", "hr.employees")
// internally looks up: "hr.employees.read"
```

The `module` is part of the `resource` argument. The action is the first argument.

### Full Examples

```go
sess.Can("create", "finance.invoices")    // "finance.invoices.create"
sess.Can("delete", "finance.invoices")    // "finance.invoices.delete"
sess.Can("read",   "hr.employees")        // "hr.employees.read"
sess.Can("run",    "hr.payroll")          // "hr.payroll.run"
sess.Can("export", "finance.reports")     // "finance.reports.export"
sess.Can("manage", "settings.roles")      // "settings.roles.manage"
```

---

## 27.5 Permission-Aware Block Pattern

### Correct: Blocks Own Their Checks

```go
// Block function — the right place for permission checks
func InvoiceToolbarBlock(sess *UISessionContext) []Node {
    toolbar := []Node{}

    if sess.Can("create", "finance.invoices") {
        toolbar = append(toolbar, ActionNode{
            Label:  "New Invoice",
            Action: "dialog",
            Target: "new-invoice-form",
        })
    }

    if sess.Can("export", "finance.reports") {
        toolbar = append(toolbar, ActionNode{
            Label:  "Export CSV",
            Action: "ajax",
            API:    "/api/finance/invoices/export",
        })
    }

    return toolbar
}

// Screen — assembles blocks without permission awareness
func InvoiceListPage(ctx context.Context, sess *UISessionContext) Schema {
    return PageSchema{
        Title: "Invoices",
        Body: []Node{
            InvoiceFilterBlock(sess),
            InvoiceTableBlock(sess),
            InvoiceToolbarBlock(sess),  // screen doesn't know what's inside
        },
    }
}
```

### Wrong: Screen Layer Doing Permission Checks

```go
// WRONG — screen is coupling itself to block internals
func InvoiceListPage(ctx context.Context, sess *UISessionContext) Schema {
    body := []Node{
        InvoiceFilterBlock(sess),
        InvoiceTableBlock(sess),
    }

    // This belongs inside InvoiceToolbarBlock
    if sess.Can("create", "finance.invoices") {
        body = append(body, InvoiceToolbarBlock(sess))
    }

    return PageSchema{Title: "Invoices", Body: body}
}
```

The wrong pattern has a subtle bug: if `InvoiceToolbarBlock` also contains the export button (which requires a different permission), the entire toolbar is hidden when the user lacks create permission — even if they have export permission.

---

## 27.6 DataTableBlock.CreatePermission

`DataTableBlock` has an explicit `CreatePermission` field. This field is set to a permission string — not derived from the URL or table name.

```go
DataTableBlock{
    Resource:         "/api/finance/invoices",
    CreatePermission: "finance.invoices.create",  // explicit string
    EditPermission:   "finance.invoices.update",
    DeletePermission: "finance.invoices.delete",
}
```

The block uses these strings to call `sess.Can()` internally when deciding whether to render toolbar buttons and row-level action columns. The string must match exactly what was registered in `allUIPermissions`.

Do not derive permission strings from the API URL. `/api/finance/invoices` → `finance.invoices.create` is not an automatic mapping — it is an explicit configuration that must be maintained.

---

## 27.7 ActionNode.DisabledOn

`ActionNode.DisabledOn` accepts an AMIS expression string that is evaluated **in the browser**. It controls whether a button is greyed out based on runtime data (e.g., record status), not permissions.

```go
// Correct: DisabledOn uses data fields, not permission strings
ActionNode{
    Label:      "Approve",
    DisabledOn: "${status !== 'pending'}",  // uses record data
}

// Correct: combining data check with a permission-based include/exclude
// The permission check happened at compile time (sess.Can) — the button
// is only in the schema if the user has the permission.
// DisabledOn refines when the already-permitted button is clickable.
if sess.Can("approve", "finance.invoices") {
    nodes = append(nodes, ActionNode{
        Label:      "Approve",
        DisabledOn: "${status !== 'pending'}",
    })
}
```

### What DisabledOn Must Not Contain

`DisabledOn` must not reference permission strings or permission-related variables. ValidateStage will reject any compiled schema where `visibleOn` or `disabledOn` expressions contain permission key patterns.

```go
// INVALID — fails ValidateStage
ActionNode{
    Label:      "Delete",
    DisabledOn: "${!permissions['finance.invoices.delete']}",  // REJECTED
}

// CORRECT — permission check at compile time, DisabledOn for data logic only
if sess.Can("delete", "finance.invoices") {
    nodes = append(nodes, ActionNode{
        Label:      "Delete",
        DisabledOn: "${selectedIds.length === 0}",  // uses UI state, not permissions
    })
}
```

The rule: `DisabledOn` is for UX-level state (no rows selected, wrong status, required field empty). Permissions are always resolved at compile time on the server.
