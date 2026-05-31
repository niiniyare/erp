# UX Principles and Personas

> Last verified: 2026-05-18 | Code pointer: `internal/web/ui/types.go` (UISessionContext), `docs/reference/modules/ui/13-ux-personas.md`

---

## Why This Matters

Every screen in Awo ERP serves a specific person. A payroll clerk sitting at a desktop all day has completely different needs from a warehouse worker who opens the app once a week on a phone. Building for the wrong person — or building for nobody in particular — produces an interface that is tolerable for everyone and excellent for no one.

The five personas documented here are the intended design targets. They are not marketing segments. They are the concrete humans whose workflow constraints should shape every layout and interaction decision made in this codebase.

This document covers:
1. Who the five intended personas are (from the design specification)
2. What the code actually knows about users today (`UISessionContext`)
3. The gap between those two things and what it means for development

---

## Intended Design: The Five Personas

These personas come from `docs/reference/modules/ui/13-ux-personas.md`. They are the design reference. Pages should be built with a primary persona in mind.

### Persona 1 — Power User (HR / Finance Staff)

```
Device:    Desktop browser, 8+ hours a day
Frequency: Every working day
Tasks:     Process payroll, manage records, generate reports

Mental model: "I know exactly what I need. Get out of my way."

Frustrations:
  - Five clicks to do something done every day
  - Confirmation dialogs for actions that need no confirmation
  - Reports that load slowly with no progress feedback

Delights:
  - Keyboard shortcuts for frequent actions
  - Bulk operations: "approve all", "export selected"
  - Dense, configurable tables with sort and filter
  - Saved views that persist between sessions

Design target: High information density. Keyboard-first. Compact layout default.
```

### Persona 2 — Line Manager

```
Device:    Mixed — desktop at desk, phone when on the floor
Frequency: 2-3 times per week
Tasks:     Approve requests, check team status, quick reports

Mental model: "Show me what needs my attention. Don't make me hunt."

Frustrations:
  - Not knowing pending approvals exist
  - Having to navigate deep to find team data
  - Seeing data that belongs to other departments

Delights:
  - Task list that surfaces exactly what needs action today
  - One-tap approval with enough context to decide without navigating
  - Team data auto-scoped — no manual filtering

Design target: Action-oriented. Push-them-in notifications. Context-aware defaults.
```

### Persona 3 — Employee (Self-Service)

```
Device:    Mobile-first
Frequency: Once a week or only on payday
Tasks:     Submit leave, check balance, view payslip, clock in/out

Mental model: "I need one thing. Make it obvious. I shouldn't need training."

Frustrations:
  - A desktop interface squeezed onto a phone screen
  - Not finding payslip without navigating through menus
  - Slow loading on mobile data

Delights:
  - Home screen showing leave balance + next shift + pending items
  - Leave request in three taps
  - Payslip download in one tap
  - Push notification when leave is approved

Design target: Mobile-first layout. 3 taps or fewer to any action. Zero training required.
```

### Persona 4 — Platform / Tenant Administrator

```
Device:    Desktop
Frequency: Weekly (setup), occasional (ongoing changes)
Tasks:     Jurisdiction settings, user management, feature flags, branding

Mental model: "I'm setting up infrastructure. Show me options clearly.
               Warn me before I do something consequential."

Frustrations:
  - Settings scattered across multiple unrelated locations
  - No indication that a change affects all users
  - Configuration changes that fail silently

Delights:
  - Structured settings hierarchy navigable by logic
  - Clear scope indicators: "This affects all Kenya tenants"
  - Live preview of branding changes before saving
  - Complete audit log of every configuration change

Design target: Structured, explicit, conservative. Prominent warnings on wide-scope actions.
```

### Persona 5 — Forecourt / Retail Operator

```
Device:    Tablet or kiosk, often outdoors or in a noisy environment
Frequency: Every shift (daily)
Tasks:     Fuel dispensing, shift reconciliation, till management

Mental model: "I'm on shift. I need to do this one thing fast."

Frustrations:
  - Small tap targets — precision tapping while standing
  - Error messages in technical language
  - Having to type anything that could be a scan
  - Session timeouts mid-transaction

Delights:
  - Large buttons, high contrast, unambiguous labels
  - Barcode scanning instead of manual entry
  - Shift summary loads instantly on login
  - Visual confirmation impossible to misread

Design target: Kiosk mode. Min 48x48px tap targets (prefer 60x80px).
               Minimal text input. AAA contrast (7:1 ratio).
```

### Persona to Component Mapping

| Persona | Primary Surface | Key AMIS Components | Density |
|---------|----------------|---------------------|---------|
| Power User | List | `crud` with all columns, bulk actions | Compact |
| Line Manager | Dashboard → Approval Inbox | `stat`, `crud` (approval list) | Comfortable |
| Employee | Self-service portal | Simple `form`, `descriptions` | Comfortable |
| Platform Admin | Settings | Tabbed `form`, `crud` (users/flags) | Comfortable |
| Forecourt Operator | Kiosk | Large-button forms, `wizard` | Kiosk |

---

## Current Implementation: What the Code Actually Knows

`UISessionContext` is the only identity and session object that reaches a `PageFn` or `ASTPageFn`. It is constructed by `AuthzStage` from the IAM contract session and a pre-resolved permissions map.

Here is what it contains today (`internal/web/ui/types.go`):

```go
type UISessionContext struct {
    // Identity — sourced from contract.SessionContext
    UserID      string
    TenantID    string
    DisplayName string
    IsPlatform  bool  // true for platform-level admin users
    IsPortal    bool  // true for portal/self-service users

    // Locale settings
    Locale   string // e.g. "en-GB"
    Timezone string // e.g. "Africa/Nairobi"
    Currency string // e.g. "KES"

    // private: populated by AuthzStage only
    permissions  map[string]bool
    featureFlags map[string]bool
    prefs        map[string]string
}
```

The two boolean flags `IsPlatform` and `IsPortal` are sourced directly from `contract.SessionContext`:

```go
IsPlatform: sc.IsPlatform(),
IsPortal:   sc.IsPortal(),
```

These are the only persona signals available to page builders today. Their meaning:

- `IsPlatform == true`: The user is a platform-level administrator (Persona 4 territory). They can see cross-tenant configuration surfaces.
- `IsPortal == true`: The user is accessing the self-service portal (Persona 3 territory). Portal sessions typically have a reduced permission set.
- Both `false`: The user is a standard tenant user — could be a Power User, Line Manager, or Forecourt Operator. The code cannot distinguish between these three today.

---

## Gap Analysis

| Persona | Intended Behavior | Current Code Signal | Gap |
|---------|------------------|---------------------|-----|
| Power User (HR/Finance) | Compact density, keyboard shortcuts, saved views, bulk ops | No dedicated signal. Neither `IsPlatform` nor `IsPortal`. Permissions may hint at role. | No persona detection. Keyboard shortcuts not implemented. Saved views not implemented. |
| Line Manager | Action-oriented dashboard, scoped team data, push notifications | No dedicated signal. May be inferred from having approval permissions. | No persona detection. Notification system not implemented. |
| Employee (Self-Service) | `IsPortal == true` narrows the surface. Mobile-first layout. | `IsPortal` exists and is passed through. Portal layout is not yet separate from the standard shell. | Signal exists but portal-specific layout not implemented. |
| Platform Admin | `IsPlatform == true` enables cross-tenant views. | `IsPlatform` exists and is passed through. Platform pages exist (tenant management). | Signal exists. Platform pages are partially implemented. |
| Forecourt Operator | Kiosk layout, large tap targets, AAA contrast | No dedicated signal. Indistinguishable from Power User in current code. | No kiosk mode. No layout variant. No persona detection. |

**Summary:** The code has two signals (`IsPlatform`, `IsPortal`) that map loosely to two of the five personas. The remaining three personas are invisible to the current `UISessionContext`. Page builders must use permissions and feature flags as indirect hints until persona detection is implemented.

---

## What This Means for Developers Right Now

When building a page today, you have these tools to approximate persona-appropriate behavior:

```go
func MyPageSchema(sess ui.UISessionContext) ui.Schema {
    // Platform admin path — show cross-tenant controls
    if sess.IsPlatform {
        return platformAdminSchema(sess)
    }

    // Portal/self-service path — simplified surface
    if sess.IsPortal {
        return selfServiceSchema(sess)
    }

    // Standard tenant user — cannot distinguish persona today
    // Use permissions to infer role-appropriate layout
    canApprove := sess.Can("approve", "leave_request")
    canRunPayroll := sess.Can("run", "payroll")

    return standardSchema(sess, canApprove, canRunPayroll)
}
```

For density and layout preferences, use `sess.Pref()`:

```go
// Users can set ui.density; default is "comfortable"
density := sess.Pref("ui.density", "comfortable")
// Values: "compact", "comfortable", "kiosk"
```

Note: `ui.density` is a preference key the system recognizes, but there is no UI for users to set it yet. It is a forward-compatible hook.

---

## Roadmap: What Full Persona Detection Would Look Like

A complete implementation would add a `Persona` field or method to `UISessionContext`:

```go
type Persona int

const (
    PersonaUnknown         Persona = 0
    PersonaPowerUser       Persona = 1
    PersonaLineManager     Persona = 2
    PersonaEmployee        Persona = 3
    PersonaPlatformAdmin   Persona = 4
    PersonaForecourtOp     Persona = 5
)

// Persona() returns the detected persona for this session.
// Resolved at AuthzStage based on role assignments in IAM.
func (u UISessionContext) Persona() Persona { ... }
```

Detection logic would run in `AuthzStage` alongside `BulkEnforce()`, using the set of assigned IAM roles to determine which persona best matches the user's role profile.

This feature is not yet implemented. See [Planned Features](../05-roadmap/02-planned.md#full-persona-system) for scope and dependencies.
