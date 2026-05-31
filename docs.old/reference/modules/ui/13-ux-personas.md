[<-- Back to Index](README.md)

> **⚠️ GAP: DESIGN INTENT vs CURRENT IMPLEMENTATION.**
> This document describes the intended 5-persona design. The current codebase implements
> **2 modes** only (`IsPlatform()` / `IsPortal()` booleans on `UISessionContext`).
>
> Features listed as persona "delights" (keyboard shortcuts, saved views, dense layouts) are
> **not yet implemented**. See [Planned Features](../05-roadmap/02-planned.md).
>
> For current implementation reality, read:
> [UX Principles and Personas](../01-theory/03-ux-principles-and-personas.md) — includes gap analysis
> table and developer guidance for working with the 2-mode system.
>
> This file is kept as the **canonical persona design spec** for when full persona detection is built.

## UX Personas

Every screen, layout, and default in Awo ERP is filtered through five personas. Before building any page, ask: which persona is the primary user of this surface?

### Persona 1 — Power User (HR / Finance Staff)

```markdown
Device:    Desktop browser, 8+ hours a day
Frequency: Every working day
Tasks:     Process payroll, manage records, generate reports

Mental model: "I know exactly what I need. Get out of my way."

Frustrations:
├── Five clicks to do something done every day
├── Confirmation dialogs for actions that need no confirmation
└── Reports that load slowly with no progress feedback

Delights:
├── Keyboard shortcuts for frequent actions
├── Bulk operations: "approve all", "export selected"
├── Dense, configurable tables with sort and filter
└── Saved views that persist between sessions

Design target: High information density. Keyboard-first. Compact layout default.
```

### Persona 2 — Line Manager

```markdown
Device:    Mixed — desktop at desk, phone when on the floor
Frequency: 2-3 times per week
Tasks:     Approve requests, check team status, quick reports

Mental model: "Show me what needs my attention. Don't make me hunt."

Frustrations:
├── Not knowing pending approvals exist
├── Having to navigate deep to find team data
└── Seeing data that belongs to other departments

Delights:
├── Task list that surfaces exactly what needs action today
├── One-tap approval with enough context to decide without navigating
└── Team data auto-scoped — no manual filtering

Design target: Action-oriented. Push-them-in notifications. Context-aware defaults.
```

### Persona 3 — Employee (Self-Service)

```markdown
Device:    Mobile-first
Frequency: Once a week or only on payday
Tasks:     Submit leave, check balance, view payslip, clock in/out

Mental model: "I need one thing. Make it obvious. I shouldn't need training."

Frustrations:
├── A desktop interface squeezed onto a phone screen
├── Not finding payslip without navigating through menus
└── Slow loading on mobile data

Delights:
├── Home screen showing leave balance + next shift + pending items
├── Leave request in three taps
├── Payslip download in one tap
└── Push notification when leave is approved

Design target: Mobile-first layout. ≤3 taps to any action. Zero training required.
```

### Persona 4 — Platform / Tenant Administrator

```markdown
Device:    Desktop
Frequency: Weekly (setup), occasional (ongoing changes)
Tasks:     Jurisdiction settings, user management, feature flags, branding

Mental model: "I'm setting up infrastructure. Show me options clearly.
               Warn me before I do something consequential."

Frustrations:
├── Settings scattered across multiple unrelated locations
├── No indication that a change affects all users
└── Configuration changes that fail silently

Delights:
├── Structured settings hierarchy navigable by logic
├── Clear scope indicators: "This affects all Kenya tenants"
├── Live preview of branding changes before saving
└── Complete audit log of every configuration change

Design target: Structured, explicit, conservative. Prominent warnings on wide-scope actions.
```

### Persona 5 — Forecourt / Retail Operator

```markdown
Device:    Tablet or kiosk, often outdoors or in a noisy environment
Frequency: Every shift (daily)
Tasks:     Fuel dispensing, shift reconciliation, till management

Mental model: "I'm on shift. I need to do this one thing fast."

Frustrations:
├── Small tap targets — precision tapping while standing
├── Error messages in technical language
├── Having to type anything that could be a scan
└── Session timeouts mid-transaction

Delights:
├── Large buttons, high contrast, unambiguous labels
├── Barcode scanning instead of manual entry
├── Shift summary loads instantly on login
└── Visual confirmation impossible to misread

Design target: Kiosk mode. Min 48×48px tap targets (prefer 60×80px).
               Minimal text input. AAA contrast (7:1 ratio).
```

### Persona → Component Mapping

| Persona | Primary Surface | Key AMIS Components | Density |
|---|---|---|---|
| Power User | List | `crud` with all columns, bulk actions | Compact |
| Line Manager | Dashboard → Approval Inbox | `stat`, `crud` (approval list) | Comfortable |
| Employee | Self-service portal | Simple `form`, `descriptions` | Comfortable |
| Platform Admin | Settings | Tabbed `form`, `crud` (users/flags) | Comfortable |
| Forecourt Operator | Kiosk | Large-button forms, `wizard` | Kiosk |

---
