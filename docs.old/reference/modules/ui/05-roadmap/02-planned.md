# Planned Features

> Last verified: 2026-05-18

This document covers UI features that are **not yet implemented**. All items here are planned but have no committed timeline. Do not cite this document as a commitment — it is a design reference.

For what is already built, see [01-implemented.md](01-implemented.md).

---

## Notification System

### What It Will Look Like

Four notification types, each with distinct delivery and urgency semantics:

| Type | Trigger | Display | Dismissal |
|------|---------|---------|-----------|
| **Toast** | Immediate feedback for user-initiated actions (save, submit, error) | Overlay in corner, auto-dismisses after 4 seconds | Auto or tap |
| **Inbox** | Background events requiring attention (approval received, payroll run complete) | Bell icon badge + full inbox panel | Manual per-item |
| **Banner** | System-wide announcements (maintenance window, feature update) | Pinned bar below topbar | Manual once per session or per day |
| **Inline** | Form field errors, contextual warnings on data rows | Embedded in the page component | Clears when condition resolves |

Toast notifications would use AMIS's built-in `toast` action, which is already available in page schemas. The work is standardizing when and how Go triggers them.

Inbox and Banner notifications require a backend notification store and a WebSocket or polling endpoint. This is the bulk of the implementation work.

### Estimated Complexity

Medium-high. Toast: low complexity (AMIS native, needs conventions). Inbox: requires backend notification table, delivery service, and a read/unread state model. Banner: requires a broadcast mechanism and per-user dismissal state.

### Dependencies

- Notification domain model (not yet designed)
- WebSocket infrastructure or SSE endpoint for push delivery
- Decision on polling interval for fallback
- `UISessionContext` extended with notification preferences (do-not-disturb flag, per-type opt-out)

---

## Saved Views

### What It Will Look Like

On any page with a `crud` component, users can save the current filter state (selected columns, active filters, sort order, page size) as a named view. Saved views persist across sessions and appear in a dropdown above the table.

```
| View: [All Invoices v]  + Save current view  |
|----------------------------------------------|
| Status: Pending  |  Date: Last 30 days  |  ...  |
```

Each saved view is owned by a user and scoped to a route. Views can optionally be shared with the user's team (same tenant, same role group).

Power Users (Persona 1) and Line Managers (Persona 2) are the primary beneficiaries.

### Estimated Complexity

Medium. The AMIS `crud` component exposes its current filter/sort state via the data chain. Saving and restoring this state is a standard API + localStorage concern. The main complexity is the UI for managing (rename, delete, share) saved views without cluttering the primary interface.

### Dependencies

- Saved views backend store (user_id + route + JSON state blob)
- Decision on sharing model (personal only vs. team-shared vs. global per-tenant)
- `UISessionContext.Pref()` extended to serve saved view list at page compile time (so the view selector is server-rendered)

---

## Keyboard Shortcuts

### What It Will Look Like

Global and per-page keyboard shortcuts for Power Users (Persona 1):

| Shortcut | Action |
|----------|--------|
| `g i` | Go to Invoices |
| `g u` | Go to Users |
| `n` | New record (on list pages) |
| `/` | Focus search/filter |
| `?` | Open shortcut help overlay |
| `Escape` | Close modal / cancel form |
| `Ctrl+Enter` | Submit form |

A help overlay (`?` key) lists all active shortcuts for the current page.

### Estimated Complexity

Medium. AMIS does not have native keyboard shortcut management. The implementation requires a custom renderer or a thin JavaScript layer that maps key sequences to AMIS actions. The main design challenge is scoping shortcuts to the current page (a shortcut that means "approve" on an approval page should not fire on a settings page).

### Dependencies

- Decision on shortcut registration API: schema-driven (Go declares shortcuts per page) vs. JavaScript convention
- Custom renderer or JS module for key sequence detection
- Shortcut help overlay component
- `PageRegistration` extended with optional `Shortcuts []ShortcutDefinition` field (so Go controls the shortcut map per page)

---

## Per-User Module Navigation Ordering

### What It Will Look Like

Users can drag-and-drop module groups in the sidebar to reorder them. The ordering persists per user across sessions. Platform admins can set a default order per tenant. Individual user preferences override the tenant default.

```
Sidebar (user's saved order):
  ✦ Finance      (user moved this to top)
  ✦ HR
  ✦ Dashboard
  ✦ Inventory
```

### Estimated Complexity

Low to medium. The ordering state is a simple ordered list of module IDs stored as a user preference. The complexity is in the drag-and-drop UI (needs a custom renderer or a supported AMIS pattern) and in integrating the ordering into the nav tree builder (`NavFn`).

### Dependencies

- User preferences store that supports arbitrary JSON blobs per key per user
- `UISessionContext.Pref()` or a dedicated `NavPrefs()` method
- Nav tree compilation (`NavFn`) updated to respect ordering preference
- Drag-and-drop UI: either AMIS sortable list component or custom renderer

---

## Full Persona System

### What It Will Look Like

A `Persona()` method on `UISessionContext` that returns one of five values:

```go
type Persona int

const (
    PersonaUnknown         Persona = 0
    PersonaPowerUser       Persona = 1  // HR/Finance staff
    PersonaLineManager     Persona = 2
    PersonaEmployee        Persona = 3  // Self-service
    PersonaPlatformAdmin   Persona = 4
    PersonaForecourtOp     Persona = 5
)

func (u UISessionContext) Persona() Persona
```

`AuthzStage` would detect the persona from the user's IAM role set during `BulkEnforce()`. The detection logic maps from role names or role types to persona constants. Users with ambiguous role sets (e.g., both a finance role and a management role) fall back to `PersonaPowerUser` by default.

Page builders would use `sess.Persona()` to select layout variants:

```go
switch sess.Persona() {
case ui.PersonaEmployee:
    return mobileFirstSchema(sess)
case ui.PersonaForecourtOp:
    return kioskSchema(sess)
default:
    return standardSchema(sess)
}
```

### Estimated Complexity

Medium. The detection logic requires agreement on which IAM role names or role types map to which personas — this is a product decision, not just a technical one. The code change itself is modest: a field added to `UISessionContext`, populated by `AuthzStage` based on the resolved role set.

The larger work is building the persona-specific layout variants for the pages that need them (employee self-service portal, kiosk mode for forecourt operators).

### Dependencies

- IAM role taxonomy finalized: which roles correspond to which personas
- Decision on handling multi-persona users (primary persona selection or blended behavior)
- `contract.SessionContext.RoleType()` or equivalent method added to the IAM contract
- Kiosk layout: separate HTML shell or AMIS layout variant (design decision needed)
- `UISessionContext` struct change: requires coordination with all existing `PageFn`/`ASTPageFn` callers that use `IsPlatform`/`IsPortal` (those remain; `Persona()` is additive)
