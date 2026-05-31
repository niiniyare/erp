[<-- Back to Index](README.md)

## Layout Principles

### The Core Problem

Enterprise software fails in two ways:

```markdown
FAILURE 1 — Information Overload:
  The screen shows everything the database knows, laid out the way
  the schema is structured — not the way the user thinks.
  A payroll clerk sees 15 tabs, 40 fields, 20 columns.

FAILURE 2 — Friction:
  The user knows what they want to do (approve 3 leave requests)
  but the app makes them navigate 4 screens and fill in a dialog.
  They give up and use WhatsApp instead. Data never gets in.

ROOT CAUSE: UI designed around the data model, not user intent.

SOLUTION: Start with what the user is trying to do right now,
          then work backwards to data and layout.
```

### Three Surfaces Per Module

Every module exposes exactly three types of surface:

```markdown
SURFACE 1 — Dashboard (situational awareness)
Purpose:  What does this person need to know right now?
Rules:    Max 6 widgets. Every widget has a direct action.
          No raw tables. Auto-scoped to user's own data.
Persona:  Line manager, employee

SURFACE 2 — List (find, filter, act on a collection)
Purpose:  Browse records, filter, bulk-act
Rules:    Dense table. Filters prominent and persistent.
          Bulk actions available. Default sort meaningful.
Persona:  Power user (HR/Finance staff)

SURFACE 3 — Detail (inspect and act on one record)
Purpose:  Full information + all actions for one entity
Rules:    Most important info above the fold.
          Related records as embedded lists, not separate screens.
          Actions grouped by consequence (safe vs destructive).
Persona:  All personas
```

### Screen Zones

Every screen uses four fixed zones. Content in the wrong zone is a design error:

```markdown
┌─────────────────────────────────────────────────┐
│  GLOBAL NAV (sidebar)                           │ Zone 1: Navigation
│  Module list | User menu | Notifications        │ Never changes
├────────────┬────────────────────────────────────┤
│  CONTEXT   │  WORKSPACE                         │
│  PANEL     │                                    │ Zone 2: Context (optional)
│            │  Page title + status               │ Filters, sub-nav
│  Filters   │  Primary actions                   │
│  Sub-nav   │  ─────────────────────             │ Zone 3: Workspace
│  Stats     │  Main content                      │ The page content
│            │  (table / form / detail)           │
│            │                                    │
├────────────┴────────────────────────────────────┤
│  STATUS BAR (optional)                          │ Zone 4: Status
│  Last saved · Keyboard shortcut hints           │ Global feedback
└─────────────────────────────────────────────────┘
```

On mobile: Zone 2 collapses to a drawer. Zone 3 is full screen. Zone 1 moves to bottom tab bar.

### The Four Questions Every Screen Must Answer

```markdown
1. WHERE AM I?         → Page title, breadcrumb, module indicator
2. WHAT IS THE STATE?  → Status badge, summary numbers
3. WHAT CAN I DO?      → Primary action button, available actions
4. WHAT IS THE DETAIL? → The data — table, form, cards

If a screen cannot answer 1–3 above the fold: redesign it.
```

### Eye-Tracking Patterns

**F-Pattern** — applies to data-heavy screens (lists, tables, forms):
- Put the most important column FIRST (leftmost)
- Row-level actions on the FAR RIGHT
- Page-level controls (filter, search, bulk) at the TOP

**Z-Pattern** — applies to dashboards and summaries:
- Most important KPI or status: TOP-LEFT
- Primary action button: TOP-RIGHT
- Secondary info: BOTTOM-LEFT
- Supporting context: BOTTOM-RIGHT

### Progressive Disclosure

Never show everything at once. Reveal information as the user needs it:

```markdown
Level 1 — Summary:    Headline. What is this record? Status?       ← Always visible
Level 2 — Key Details: Most important attributes. Above the fold.  ← Always visible
Level 3 — Full Detail: Sub-records, history.                       ← Expanded by default
Level 4 — Advanced:    Raw data, event log, debug info.            ← Collapsed toggle

Example — Payroll Run Detail:
  Level 1: Period, status badge, total gross, total net
  Level 2: Employee count, PAYE total, pay date
  Level 3: Per-employee payslip list (paginated table)
  Level 4: Rate snapshot, audit events, GL journal — behind "Show Details" toggle
```

### Data Density Modes

Three density modes, configurable per user:

| Mode | Row Height | Font | Use For |
|---|---|---|---|
| Comfortable | 36–40px | 14–16px | Employees, line managers |
| Compact | 28–32px | 12–13px | HR/Finance power users |
| Kiosk | 60–80px | 16–18px | Forecourt operators |

These are not just CSS tweaks — they change column count, available controls, and sometimes layout structure.

### Cards vs Tables vs Lists: When to Use Each

| Pattern | Use When | Example |
|---|---|---|
| KPI Card | Single number with context | "47 employees · 3 pending" |
| Summary Card | One record with 4–6 fields | Employee profile card |
| Data Table | Many records, user needs to compare | Leave request list |
| Feed / Timeline | Ordered events, chronological | Audit log, activity feed |
| Kanban | Records moving through stages | Leave approvals pipeline |
| Form | Creating or editing a record | New employee, leave request |
| Wizard | Multi-step process with dependencies | Run payroll, onboard employee |

---
