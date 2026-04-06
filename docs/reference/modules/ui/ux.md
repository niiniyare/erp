# UX & UI Architecture — Awo ERP

**Version:** 1.0.0  
**Status:** Design Specification  
**Audience:** Engineering · Product · Design  
**Scope:** Stack-agnostic — principles and patterns applicable to any frontend technology

---

## Table of Contents

1. [The Core Problem](#1-the-core-problem)
2. [User Personas & Mental Models](#2-user-personas--mental-models)
3. [Information Architecture](#3-information-architecture)
4. [Organising Data on the Screen](#4-organising-data-on-the-screen)
5. [Navigation Architecture](#5-navigation-architecture)
6. [Forms — Getting Data In](#6-forms--getting-data-in)
7. [Tables & Lists — Getting Data Out](#7-tables--lists--getting-data-out)
8. [Personalisation & User Preferences](#8-personalisation--user-preferences)
9. [Theming, Dark Mode & Tenant Branding](#9-theming-dark-mode--tenant-branding)
10. [Notifications & Alerts](#10-notifications--alerts)
11. [Accessibility](#11-accessibility)
12. [Adoption — Making People Want to Use It](#12-adoption--making-people-want-to-use-it)
13. [Design Tokens & Visual Language](#13-design-tokens--visual-language)
14. [Implementation Checklist](#14-implementation-checklist)

---

## 1. The Core Problem

Enterprise software fails its users in two predictable ways.

The first is **information overload**. The screen shows everything the database knows, laid out the way the schema is structured — not the way the user thinks. A payroll clerk opens a screen and sees fifteen tabs, forty fields, and a table with twenty columns. They learn to ignore most of it, which means they miss the things that matter.

The second is **friction**. The user knows what they want to do — approve three leave requests — but the app makes them navigate four screens, fill in a confirmation dialog, and click a button that is greyed out for reasons never explained. They give up and use WhatsApp instead. The data never gets in. The reports are wrong.

Both failures share the same root cause: **the UI is designed around the data model, not around what the user is trying to accomplish.**

Awo ERP solves this by starting with user intent — what is this person trying to do right now, in this session — and working backwards to the data and the layout. Every design decision in this document flows from that principle.

---

## 2. User Personas & Mental Models

Every screen, every layout decision, every default in Awo ERP is filtered through five distinct personas. Each has a different device, a different level of system familiarity, a different tolerance for complexity, and a different definition of a successful session.

### Persona 1 — The Power User (HR / Finance Staff)

**Context:** Desktop browser, 8 hours a day. Processes payroll, manages records, generates reports. Deep domain knowledge. Wants to move fast and not be patronised by the UI.

**Mental model:** *"I know exactly what I need. Get out of my way and let me do ten things efficiently."*

**Frustrations:**
- Five clicks to do something done every day
- Confirmation dialogs for actions that need no confirmation
- Data hidden behind horizontal scrolling
- Reports that load slowly with no progress feedback

**Delights:**
- Keyboard shortcuts for the most frequent actions
- Bulk operations — "approve all", "export selected", "run payroll for this department"
- Dense, configurable tables they can sort and filter
- Saved views and filters that persist between sessions

**Design target:** High information density. Minimal decorative chrome. Fast navigation. Keyboard-first interactions.

---

### Persona 2 — The Line Manager

**Context:** Mixed device — desktop at their desk, phone when walking the floor. Uses the system 2–3 times per week, mainly for approvals, team visibility, and quick reports. Comfortable with smartphones; less so with complex desktop software.

**Mental model:** *"Just show me what needs my attention. Don't make me go looking for it."*

**Frustrations:**
- Not knowing there are pending approvals waiting
- Having to navigate deep into the system to find their team's data
- Forms asking for information they don't have
- Seeing data that belongs to other departments

**Delights:**
- A task list that surfaces exactly what needs action today
- One-tap approval with enough context to decide without leaving the screen
- Automatic scoping to their own team without having to filter manually

**Design target:** Action-oriented. Context-aware. Pull-them-in notifications rather than expecting them to come looking.

---

### Persona 3 — The Employee (Self-Service)

**Context:** Mobile-first. Checks leave balance, submits leave requests, clocks in/out, views payslips, requests shift swaps. May use the app once a week or only on payday. Very low tolerance for complexity.

**Mental model:** *"I need one thing. Make it obvious. I shouldn't need to learn this app."*

**Frustrations:**
- A desktop interface squeezed onto a phone screen
- Not finding their payslip without navigating through menus
- Fields on a form that clearly don't apply to them
- Slow loading on a mobile data connection

**Delights:**
- A home screen showing their leave balance, next shift, and any pending items
- Leave request in three taps
- Payslip download in one tap
- Push notification when their leave is approved

**Design target:** Mobile-first layout. Minimal navigation depth. Immediate feedback. Zero training required.

---

### Persona 4 — The Platform / Tenant Administrator

**Context:** Desktop. Configures the system — jurisdiction settings, tenant branding, user management, feature flags, billing. Technically literate but not necessarily an ERP domain expert.

**Mental model:** *"I'm setting up infrastructure. Show me the options clearly. Warn me before I do something consequential."*

**Frustrations:**
- Settings scattered across multiple unrelated locations
- No indication that a change will affect all tenants or all users
- Configuration changes that fail silently
- Having to contact engineering to change a flag

**Delights:**
- A structured settings hierarchy navigable by logic
- Clear scope indicators ("this setting affects all Kenya tenants")
- Live preview of branding changes before saving
- A complete audit log of every configuration change

**Design target:** Structured, explicit, conservative. Prominent warnings on wide-scope or destructive actions. Confirmation with visible consequences shown before committing.

---

### Persona 5 — The Forecourt / Retail Operator

**Context:** Tablet or kiosk, often in a noisy or outdoor environment, standing, potentially wearing gloves. Processes fuel dispensing, manages shift reconciliation, handles the till. Uses the app under time pressure. May have limited technical confidence.

**Mental model:** *"I'm on shift. I need to do this one thing fast and get back to work."*

**Frustrations:**
- Small tap targets that require precision tapping
- Error messages in technical language
- Having to type anything that could be a tap or a scan
- Session timeouts mid-transaction

**Delights:**
- Large buttons, high contrast, unambiguous labels
- Barcode or camera scanning instead of manual entry
- Shift summary that loads instantly on login
- Visual confirmation that is impossible to misread

**Design target:** Kiosk mode. Large touch targets (minimum 48×48px). Minimal text input. Fast, decisive, unambiguous feedback.

---

## 3. Information Architecture

Information architecture answers: **what lives where, and how does a user find it without thinking about it?**

### The Four-Layer Hierarchy

The entire system is structured in four levels. Users only ever see the levels they have access to. A regular employee sees no indication that a "platform" level exists.

```
Platform
  └── Tenant
        └── Module
              └── Resource
```

This maps directly to URL structure and breadcrumb navigation:

```
/platform/jurisdictions
/settings/payroll/statutory
/hr/employees
/hr/employees/{id}/payslips
/finance/payroll-runs/{id}
```

### The Three Surfaces Every Module Exposes

Rather than building a different layout for every screen, every module exposes exactly three types of surface. Each serves a different purpose and a different persona.

**Surface 1 — The Dashboard**

*Intent: situational awareness*

What does this person need to know right now? Not everything — just what is actionable, pending, or unusual. Dashboards are scannable in under ten seconds. Line managers and employees spend most of their time here.

Rules for dashboard design:
- Maximum 6 widgets on any dashboard
- Every widget either shows a number, a list of pending actions, or a trend
- Every actionable item has a direct action button in the widget — no "click here to go to the list and then find it"
- Widgets are filtered to the user's own scope automatically — a manager sees their team only
- No raw tables on a dashboard — tables belong in the list surface

**Surface 2 — The List**

*Intent: find, filter, and act on a collection of records*

Power users live here. This is the HR clerk reviewing all pending leave requests, the finance manager browsing payroll runs, the platform admin auditing jurisdiction changes. Lists must be fast, filterable, and configurable.

Rules for list design:
- Default to the most useful sort order for that resource (newest first for runs, name for employees)
- Show the most important 5–7 columns by default; let users add more
- Filters are prominent and persistent between sessions
- Bulk actions are available when multiple rows are selected
- Every row has a clear primary action (the most common next step for that record type)

**Surface 3 — The Detail**

*Intent: inspect and act on a single record*

Full information about one entity, all associated sub-records, and all applicable actions. The detail view is where forms live (editing a record) and where workflows are advanced (approving a run, terminating an employee).

Rules for detail design:
- The most important information is above the fold, always
- Related records are shown as embedded lists or summary cards, not links to other screens
- Actions are grouped by consequence — safe actions (view, export) separated from destructive actions (delete, reverse, terminate)
- The status of the record is always visible — never buried

---

## 4. Organising Data on the Screen

This is the most direct answer to "how do I organise data on the screen" — a set of concrete layout rules that apply across every surface in Awo ERP.

### The F-Pattern and Z-Pattern

Eye-tracking research shows that users scan screens in two predictable patterns depending on content type.

**F-Pattern** — applies to data-heavy screens (lists, tables, forms). Users read across the top, then scan down the left edge, then occasionally scan right for interesting rows. This means:
- Put the most important column first (leftmost)
- Put row-level actions on the far right (users will look there after finding the row)
- Put page-level controls (filters, search, bulk actions) at the top

**Z-Pattern** — applies to summary and dashboard screens. Users look top-left, scan right, diagonal down-left, then right again. This means:
- Put the most important KPI or status top-left
- Put the primary action button top-right
- Put secondary information bottom-left, supporting context bottom-right

### The Information Hierarchy on Every Screen

Every screen should answer four questions, in this order:

```
1. WHERE AM I?        → Page title, breadcrumb, module indicator
2. WHAT IS THE STATE? → Status badge, count, summary numbers
3. WHAT CAN I DO?     → Primary action button, available actions
4. WHAT IS THE DETAIL? → The data itself — table, form, cards
```

If a screen cannot answer questions 1–3 within the first visible section (above the fold), it needs to be redesigned.

### Screen Zones

Divide every screen into four named zones. Each zone has a fixed purpose. Never put content in the wrong zone.

```
┌─────────────────────────────────────────────────┐
│  GLOBAL NAV                                     │  Zone 1: Navigation
│  Module nav | User menu | Notifications         │  Never changes; always accessible
├────────────┬────────────────────────────────────┤
│  CONTEXT   │  WORKSPACE                         │
│  PANEL     │                                    │  Zone 2: Context (optional)
│            │  Page title + status               │  Filters, record list, sub-nav
│  Filters   │  Primary actions                   │
│  Sub-nav   │  ─────────────────────             │  Zone 3: Workspace
│  Summary   │  Main content                      │  The actual page content
│  stats     │  (table / form / detail)           │
│            │                                    │
│            │                                    │
├────────────┴────────────────────────────────────┤
│  STATUS BAR (optional)                          │  Zone 4: Status
│  Last saved · Sync status · Keyboard shortcuts  │  Global feedback and hints
└─────────────────────────────────────────────────┘
```

On mobile, Zone 2 (Context Panel) collapses into a drawer or bottom sheet. Zone 3 occupies the full screen. Zone 1 moves to a bottom tab bar.

### Progressive Disclosure

Never show everything at once. Reveal information as the user needs it.

**Level 1 — Summary:** Show the headline. What is this record? What is its status?  
**Level 2 — Key details:** Show the most important attributes. Still fits above the fold.  
**Level 3 — Full detail:** Show everything, including sub-records, history, and metadata.  
**Level 4 — Advanced / audit:** Raw data, event log, developer information. Behind a disclosure toggle.

In practice, a payroll run detail page looks like this:

```
Level 1: Run period, status badge, total gross, total net           ← always visible
Level 2: Employee count, PAYE total, statutory totals, pay date     ← always visible
Level 3: Per-employee payslip list (paginated table)                ← expanded by default
Level 4: Rate snapshot, audit events, GL journal instructions       ← collapsed, expand on demand
```

### Data Density Modes

Different personas need different density. Awo ERP supports three density modes that users can set in their preferences:

**Comfortable** (default for employees and line managers)
- Generous padding in table rows (36–40px row height)
- Larger font sizes (14–16px body)
- More white space between sections
- Fewer columns visible by default

**Compact** (default for HR / Finance power users)
- Tight padding (28–32px row height)
- Smaller body font (12–13px)
- Tighter section spacing
- More columns visible by default

**Kiosk** (default for forecourt / retail operators)
- Very large tap targets (60–80px row height)
- Large font (16–18px minimum)
- Maximum contrast
- Single-column layout optimised for one-handed or gloved use

These are not just CSS tweaks — they change the number of columns shown, the controls available, and sometimes the layout structure itself.

### Cards vs Tables vs Lists: When to Use Each

| Pattern | Use When | Example |
|---|---|---|
| **KPI Card** | Single number with context | "47 employees · 3 pending approvals" |
| **Summary Card** | One record with 4–6 key fields | Employee profile card, payslip summary |
| **Data Table** | Many records, multiple attributes, user needs to compare | Leave request list, payroll run history |
| **Feed / Timeline** | Ordered events, chronological importance | Employee activity log, payroll audit trail |
| **Kanban / Status Board** | Records moving through stages | Leave approvals, onboarding pipeline |
| **Form** | User is creating or editing a record | New employee, leave request |
| **Wizard / Stepper** | Multi-step process with dependencies | Run payroll, onboard employee |

Never use a data table when a feed fits better. Never use a card layout when a table would let the user compare across rows. The choice of pattern is a design decision, not a framework default.

---

## 5. Navigation Architecture

### Global Navigation

The global navigation is the constant backbone of the application. It is always visible on desktop and always reachable on mobile (bottom tab bar or hamburger).

**Sections:**
- **Module switcher** — HR, Finance, Inventory, etc. Filtered to modules the user has access to
- **Search** — global search across all records the user can access
- **Notifications bell** — count badge; opens notification panel
- **User menu** — profile, preferences, help, sign out

The number of top-level items in the module switcher must never exceed 7. If a tenant has more than 7 active modules, group them into categories. Navigation research consistently shows that 7 items is the practical maximum before users stop scanning and start searching.

### Module Navigation

Within each module, navigation is secondary — a sidebar on desktop, a tab bar on mobile. Module navigation items map to the three surfaces:

```
HR Module
  ├── Dashboard          ← always first
  ├── Employees
  ├── Payroll Runs
  ├── Leave
  ├── Shifts & Rosters
  ├── Attendance
  └── Reports
```

Active state is always explicit. The user must never wonder which section they are in.

### Breadcrumbs

Breadcrumbs are mandatory on detail pages. They give the user a mental model of where they are and a one-click escape back to the list.

```
HR > Employees > Amina Odhiambo > Payslips > November 2024
```

On mobile, show only the immediate parent, not the full path:

```
← Amina Odhiambo
```

### The "What Needs My Attention" Entry Point

Every persona except the platform admin has a personalised task list as their landing page. This is not a generic dashboard — it is a list of things specifically waiting for them:

```
YOUR ACTION ITEMS TODAY                         3 pending

  [ ! ] Leave Request                    Approve or Decline
        Amina Odhiambo · 23–31 Dec · Annual Leave
        Submitted 2 days ago

  [ ! ] Timesheet Approval              Approve or Decline
        Forecourt Team · November 2024
        7 employees pending

  [ ✓ ] Payroll Run Ready for Review    Review Now
        November 2024 · 47 employees · KES 4.2M gross
```

Every item links directly to the relevant action, pre-loaded with context. The user never needs to navigate to find these items — they are surfaced automatically based on role, team scope, and outstanding workflows.

---

## 6. Forms — Getting Data In

Forms are where most UX debt accumulates in enterprise software. The following rules apply to every form in Awo ERP.

### The Golden Rule of Forms

**Only show a field when the user needs to fill it in right now.** If a field is optional and rarely used, hide it behind an "Add more details" expansion. If a field is only relevant based on a prior answer, show it only after that answer is given.

A form with 8 visible fields that intelligently reveals 4 more when needed feels simpler than a form with 12 always-visible fields — even though the total field count is higher.

### Field Types and When to Use Them

| Field Type | Use When | Never Use When |
|---|---|---|
| **Text input** | Free text, names, descriptions | Selecting from a known set of options |
| **Number input** | Quantities, amounts, percentages | Codes or IDs that happen to be numeric |
| **Select / Dropdown** | 5–15 mutually exclusive options | Fewer than 5 options (use radio buttons) |
| **Radio buttons** | 2–5 mutually exclusive options, all visible | More than 5 options (use select) |
| **Checkboxes** | Multiple selections from a known set | Single selection (use radio/select) |
| **Toggle / Switch** | Binary on/off state | Anything with meaningful intermediate states |
| **Date picker** | Specific date selection | Relative dates ("last 30 days") |
| **Date range picker** | Period selection (payroll period, leave period) | Single-date selection |
| **Autocomplete / search** | Large collections (employees, GL accounts, countries) | Small sets where a dropdown is clearer |
| **File upload** | Document attachments, evidence | Anything that can be entered as structured data |
| **Rich text editor** | Long-form content (policies, descriptions) | Short text — use a plain text input |

### Form Layout Principles

**Single column is almost always right.** Two-column form layouts feel efficient on a wide screen but cause problems — users skip fields, the reading order is ambiguous, and they break completely on mobile. Use a single column for all forms. The exception is pairs of clearly related fields (start date / end date, minimum salary / maximum salary) which can sit on the same row.

**Group related fields visually.** Use section headers, not card borders, to group. Card borders create visual noise. A clear section heading ("Contract Details", "Bank Account", "Tax Information") is enough.

**Label position.** Labels sit above their input, not to the left. Left-aligned labels look clean on desktop but cause problems for long labels, translated strings, and any responsive layout. Top labels always work.

**Placeholder text is not a label.** Placeholder text disappears the moment the user starts typing. Never use placeholder text as the only label for a field. Always use a visible label. Use placeholder text only to show example values.

**Required vs optional.** Mark optional fields as "optional" rather than marking required fields with an asterisk. Most fields in a form should be required. Marking the rare optional fields is less visual noise and clearer.

### Validation

**Validate in real time, not only on submit.** As soon as the user leaves a field (on blur), validate it and show the result immediately. Do not wait until they click submit to reveal twenty errors.

**Error messages must explain and guide.** Never show "Invalid input". Always show what is wrong and how to fix it.

```
✗ The salary you entered (KES 8,000) is below the minimum wage for Kenya (KES 16,200).
  Enter KES 16,200 or more, or check the employee's jurisdiction setting.

✓ Basic Salary
  KES 120,000
  ✓ Within Grade G3 salary band (KES 80,000 – KES 180,000)
```

Positive inline validation (showing a green check when a field is correct) significantly reduces anxiety in complex forms like payroll configuration or statutory rate setup.

**Block on errors; warn on advisories.** Distinguish between errors that prevent saving (hard block) and warnings that advise but allow proceeding (soft advisory). Use different visual treatments — red for errors, amber for warnings.

### Multi-Step Forms (Wizards)

For complex, sequential processes — running payroll, onboarding a new employee, configuring a jurisdiction — use a wizard pattern with explicit steps.

Rules for wizards:
- Show all steps upfront so the user knows how long the process is
- Mark completed steps visually (checkmark, filled dot)
- Allow going back to any completed step without losing data
- Auto-save state between steps — if the user's session expires, they return to where they left off
- The final step is always a review screen: show everything the user entered, let them go back to any section, and confirm once
- Show a loading state on the submit action — payroll runs can take seconds; the user needs to know something is happening

```
Step 1     Step 2     Step 3     Step 4     Step 5
  ●─────────●─────────○─────────○─────────○
Period    Employees  Review    Approve   Post

Completed  Active     Upcoming
```

### Smart Defaults

Every field should have the best possible default value pre-filled. The user should only need to change fields that differ from the typical case.

- Payroll run period defaults to the current calendar month
- Currency defaults to the tenant's base currency
- Jurisdiction defaults to the tenant's primary jurisdiction
- "All active employees" defaults to checked for payroll runs
- Pay date defaults to the last working day of the period

Smart defaults are not guesses — they are informed by the most common usage pattern, configurable per tenant, and always overridable.

---

## 7. Tables & Lists — Getting Data Out

Tables are the dominant UI pattern in ERP software and the one most often implemented badly. The rules below make tables fast, readable, and useful.

### Column Design

**Every column earns its place.** Before adding a column, ask: does the user need to compare this value across rows? If no, it belongs in the detail view, not the table.

**Default column selection by persona:**

The same resource list shows different default columns depending on who is viewing it:

```
Employee list — HR Admin view (power user, desktop):
  Photo | Name | Employee# | Department | Grade | Status | Join Date | Actions

Employee list — Line Manager view:
  Photo | Name | Status | Shift Today | Leave Balance | Actions

Employee list — Employee view (their team only):
  Photo | Name | Department | Contact | Actions
```

This is not access control — it is relevance control. The manager can see all data if they add columns. But the default shows what they actually need.

**Column types and alignment:**

| Data Type | Alignment | Format |
|---|---|---|
| Text (names, descriptions) | Left | As-is |
| Numbers (counts, quantities) | Right | Locale-aware formatting |
| Currency amounts | Right | With currency code, 2 decimal places |
| Percentages | Right | e.g. 14.5% |
| Dates | Left or center | Relative for recent ("2 days ago"), absolute for older ("14 Nov 2024") |
| Status | Center | Badge / pill with colour |
| Boolean | Center | Icon (check/cross), not "true/false" |
| Actions | Right | Icon buttons or "..." menu |

**Never mix alignment in a column.** If currency amounts are right-aligned, every row in that column must be right-aligned, including the header and any totals row.

### Sorting & Filtering

**Default sort** must be the most useful for the user's task, not whatever the database returns. Payroll runs: newest first. Employees: alphabetical by last name. Leave requests: oldest pending first (the ones that need action soonest).

**Column sorting:** Click the column header to sort ascending; click again for descending. Show the sort direction with a clear arrow indicator. Allow sorting by only one column at a time unless the user explicitly requests multi-column sort.

**Filters are first-class citizens.** They are not tucked into a small dropdown or hidden in an "Advanced" section. Primary filters (status, date range, department) are always visible above the table. Secondary filters are one click away in a filter panel.

Filter design rules:
- Filters persist across page refreshes (saved in user preferences or URL params)
- Show the active filter count so users know filters are applied: "Filters (3)"
- A single "Clear all filters" action resets everything
- Filter options should show counts: "Active (41) · Pending (3) · Terminated (12)"

### Pagination vs Infinite Scroll

Use **pagination** for tables, not infinite scroll.

Infinite scroll sounds user-friendly but creates serious problems in enterprise contexts:
- The user cannot jump to "page 5 of 12" — they have to scroll there
- The URL doesn't represent a position, so sharing a link loses context
- Bulk selection becomes confusing when the list is growing
- Performance degrades as the DOM grows

Use pagination with explicit page size control (25 / 50 / 100 rows) and a clear indicator of position ("Showing 51–100 of 847"). Power users set this to 100 and stay on page 1. That is the right behaviour.

Exception: use infinite scroll in feed-style displays (activity logs, notifications, audit trails) where the user is scanning chronologically and position is less important.

### Bulk Actions

For list views where the user manages many records of the same type, bulk actions dramatically reduce friction:

```
☑ 3 of 47 selected             [Approve] [Export] [Send Payslip] [···]
```

Rules for bulk actions:
- The bulk action bar appears only when at least one row is selected
- "Select all on this page" and "Select all 847 results" are distinct actions — make the difference clear
- Destructive bulk actions (delete, terminate) require a confirmation that shows the count: "Terminate 3 employees. This cannot be undone."
- After a bulk action completes, show a result summary: "3 leave requests approved. 1 failed — Amina Odhiambo (insufficient balance)."

### Empty States

Every list view must have a designed empty state. An empty table with just column headers is confusing — it looks broken. The empty state should:
- Explain why the list is empty (no records exist, or filters exclude everything)
- Tell the user what they can do about it
- Provide a direct action if one exists

```
No payroll runs yet
─────────────────────────────────────
Start your first payroll run for November 2024.
Awo will guide you through the process step by step.

                [Start Payroll Run]
```

vs

```
No results match your filters
─────────────────────────────────────
Try removing some filters or expanding the date range.

                [Clear Filters]
```

These are different empty states for the same list. Show the right one based on whether filters are active.

---

## 8. Personalisation & User Preferences

Personalisation is what turns a generic enterprise system into something that feels built for a specific person. Awo ERP implements two types: **automatic personalisation** (the system learns from behaviour) and **explicit preferences** (the user configures deliberately).

### Automatic Personalisation

These happen without the user doing anything.

**Scope filtering:** Every list is automatically filtered to the user's organisational scope. A department manager sees their department's employees by default. They can expand the scope, but the default is always their own world.

**Recent items:** The top of the search dropdown shows recently accessed records before the user types anything. The user navigates to their most common records in one keystroke.

**Usage-based shortcuts:** If a user visits the "Run Payroll" wizard every first working day of the month, the system surfaces this as a quick action on the dashboard around that time.

**Adaptive column defaults:** If a user consistently adds the same column to a table within their first session, that column becomes a default for them on subsequent sessions.

### Explicit User Preferences

Stored per user, per tenant. Accessible from the user menu → Preferences.

#### Appearance

| Preference | Options | Default |
|---|---|---|
| Colour scheme | Light / Dark / System | System |
| Data density | Comfortable / Compact / Kiosk | Comfortable |
| Font size | Small / Default / Large | Default |
| Language | Available locales for the tenant | Tenant default |
| Date format | DD/MM/YYYY · MM/DD/YYYY · YYYY-MM-DD | Tenant default |
| Number format | 1,234.56 · 1.234,56 · 1 234,56 | Tenant locale |
| Currency display | Symbol (KES) · Code (KES) · Both | Tenant default |

#### Navigation

| Preference | Options |
|---|---|
| Default landing page | Task list / Module dashboard / Last visited |
| Sidebar state on desktop | Always expanded / Always collapsed / Remember last state |
| Module order | Drag-and-drop reordering of module switcher |
| Pinned items | Pin specific records or views to the sidebar |

#### Table Preferences (per table, per user)

| Preference | Description |
|---|---|
| Column selection | Which columns are visible |
| Column order | Left-to-right order of visible columns |
| Column widths | Pixel widths saved after manual resize |
| Default sort | Which column and direction |
| Page size | Rows per page |
| Saved filters | Named filter sets the user can recall in one click |

#### Notifications

| Preference | Options |
|---|---|
| In-app notifications | All / Mentions only / None |
| Email digest | Real-time / Daily / Weekly / Never |
| Push notifications (mobile) | All / Urgent only / None |
| Notification types | Individually toggle each notification category |

### Saved Views

A saved view is a named combination of filters, column selection, and sort order on a list. This is the single highest-leverage personalisation feature for power users.

```
My Saved Views — Employee List
  ★ Active Forecourt Staff          [Default]
  ○ Pending Onboarding
  ○ Terminated This Year
  ○ Grade G3 and above

                           [+ Save current view]
```

Saved views can be private (user only) or shared with a role or department. An HR manager can create a shared view "Employees Due for Probation Review" and share it with all line managers. This is configuration that travels with the user, not the tenant.

### User Preference Storage Architecture

User preferences are stored in a `user_preferences` JSONB column rather than individual columns. This keeps the schema stable as new preferences are added.

```json
{
  "appearance": {
    "colour_scheme": "dark",
    "density": "compact",
    "font_size": "default"
  },
  "locale": {
    "language": "en-KE",
    "date_format": "DD/MM/YYYY",
    "number_format": "1,234.56"
  },
  "navigation": {
    "default_landing": "task_list",
    "sidebar_state": "expanded",
    "module_order": ["hr", "finance", "inventory"]
  },
  "tables": {
    "hr.employees": {
      "columns": ["photo", "name", "department", "grade", "status", "join_date"],
      "sort": { "column": "last_name", "direction": "asc" },
      "page_size": 50
    }
  },
  "notifications": {
    "in_app": "all",
    "email_digest": "daily",
    "push": "urgent_only"
  }
}
```

Preferences are loaded once at session start and cached client-side. Changes are persisted asynchronously — the UI updates immediately without waiting for the server response.

---

## 9. Theming, Dark Mode & Tenant Branding

### Why This Matters

Tenant branding is not a vanity feature. When an operator logs into a system that shows their company's logo and colours, they trust it. When every tenant sees the same generic purple and grey, the system feels like shared infrastructure — not their system. Adoption is significantly higher when users feel ownership.

### The Three-Tier Token System

All visual values — colours, spacing, typography, border radius — are stored as design tokens in a three-tier hierarchy. No hardcoded colour values anywhere in the UI.

```
TIER 1 — PRIMITIVE TOKENS
Raw values. Never used directly in components.

  colour-blue-500: #3B82F6
  colour-grey-900: #111827
  spacing-4: 16px
  radius-md: 6px
  font-size-sm: 13px

TIER 2 — SEMANTIC TOKENS
Named by purpose. Reference primitives. Used in components.

  colour-primary:         → colour-blue-500
  colour-background:      → colour-white     (light) / colour-grey-950 (dark)
  colour-surface:         → colour-grey-50   (light) / colour-grey-900 (dark)
  colour-text-primary:    → colour-grey-900  (light) / colour-grey-50  (dark)
  colour-text-secondary:  → colour-grey-500  (light) / colour-grey-400 (dark)
  colour-border:          → colour-grey-200  (light) / colour-grey-700 (dark)
  colour-danger:          → colour-red-600
  colour-warning:         → colour-amber-500
  colour-success:         → colour-green-600

TIER 3 — COMPONENT TOKENS
Named by component. Reference semantic tokens.

  button-primary-bg:      → colour-primary
  button-primary-text:    → colour-white
  sidebar-bg:             → colour-surface
  table-row-hover:        → colour-grey-50 (light) / colour-grey-800 (dark)
  badge-active-bg:        → colour-green-100
  badge-active-text:      → colour-green-800
```

Only Tier 2 and Tier 3 tokens are ever referenced in component styles. When a tenant overrides their primary colour, they change `colour-primary`. Every button, link, badge, and focus ring that references `colour-primary` updates automatically.

### Tenant Branding Configuration

Tenant administrators configure branding in Settings → Branding. Changes take effect immediately for all users of that tenant.

**What tenants can configure:**

| Setting | Description | Validation |
|---|---|---|
| Primary colour | Main brand colour used for buttons, links, active states | Must pass AA contrast ratio against white and against `colour-surface` |
| Logo (light mode) | SVG or PNG, max 200×60px | Validated dimensions; virus-scanned |
| Logo (dark mode) | Separate logo for dark backgrounds (optional) | Same constraints; falls back to light logo with CSS inversion if absent |
| Favicon | 32×32 or 64×64 ICO/PNG | Used in browser tab |
| Application name | Displayed in the tab title and email subjects | Max 40 characters |
| Custom CSS | Power escape hatch for advanced tenants | Sandboxed; platform can disable per tenant |

**What tenants cannot configure:**
- Font family (consistency and performance)
- Layout structure (menu position, grid)
- Status colours (danger, warning, success must remain consistent for safety)
- Any colour that would fail WCAG AA contrast

### Dark Mode

Dark mode is a first-class feature, not an afterthought. It reduces eye strain for power users working long hours and is frequently requested in markets where users work in variable lighting conditions (forecourt operations, night shifts).

Implementation rules:
- Every colour in the UI uses semantic tokens — no hardcoded hex values in component styles
- Dark mode swaps Tier 2 semantic tokens only — component styles are untouched
- Images and logos that are dark on a transparent background need a light-mode variant; this is solved by the per-tenant dark logo upload
- Charts and data visualisations must have explicit dark mode styles — most charting libraries do not handle this automatically
- The mode switches instantly (no page reload) by toggling a `data-theme` attribute on the root element

### Per-User Colour Scheme vs Tenant Override

Individual users set their preferred colour scheme in Preferences. The tenant can optionally enforce a colour scheme for all users (e.g. a tenant that always wants to show dark mode). The resolution order:

```
Tenant forced scheme → User preference → System preference (OS dark/light)
```

---

## 10. Notifications & Alerts

Notifications are the primary mechanism for pulling users back into the system when their attention is needed. Done well, they reduce the "I didn't know I had to approve that" problem. Done badly, they train users to ignore them.

### The Four Notification Types

**Type 1 — Inline Alert (immediate, contextual)**  
Appears on the current screen in response to something just done or something wrong with data being viewed. Not a popup. Not a toast. A banner or inline message within the relevant section.

```
⚠ This payroll run includes 3 employees whose timesheets have not been approved.
  They will be excluded from the run unless timesheets are approved before processing.
  [View Unapproved Timesheets]
```

Use for: validation warnings, data quality issues visible on the current screen, information the user needs before taking an action.

**Type 2 — Toast / Snackbar (transient feedback)**  
A small message that appears briefly (3–5 seconds) after an action is completed. Confirms the action worked. Disappears automatically. Optionally includes an undo action.

```
✓ Leave request approved — Amina Odhiambo · 23–31 Dec
                                                   [Undo]  3s
```

Use for: success confirmation, undo opportunities. Never for errors (errors must persist until resolved).

**Type 3 — Notification Panel (async, pull)**  
The bell icon in the global navigation. Opens a panel listing recent notifications. Each item is actionable — clicking it takes the user directly to the relevant record.

```
NOTIFICATIONS                           Mark all read

  ● Payroll run ready for review                   2h ago
    November 2024 · 47 employees · KES 4.2M gross
    [Review Run →]

  ● Leave request pending your approval            1d ago
    Amina Odhiambo · Annual Leave · 23–31 Dec
    [Approve] [Decline]

  ✓ NSSF remittance schedule generated             2d ago
    November 2024 [Download →]
```

Rules:
- Unread count badge on the bell icon
- Notifications are scoped to the user — a line manager only sees notifications about their team
- Each notification has a direct action link — no navigation required
- Older than 30 days are archived, not deleted
- The user can configure which event types generate notifications (§8)

**Type 4 — Push Notification / Email (async, push)**  
Delivered outside the application. Used only for high-priority events that require action even when the user is not in the app.

Push notification triggers (configurable, opt-in by default):
- Leave request awaiting your approval (> 24 hours pending)
- Payroll run ready for your review
- Your leave request has been approved or declined
- Your payslip is available
- System maintenance window (platform admin)

Email digest triggers (configurable):
- Daily summary of pending approvals
- Weekly payroll summary for finance managers
- Monthly statutory filing reminders

Rules:
- Every push notification deep-links directly into the relevant action in the app
- Never send a notification that does not contain enough context to decide without opening the app ("You have a pending item" is useless; "Amina Odhiambo's leave request for 23–31 Dec needs your approval" is useful)
- Always honour user notification preferences — never override them for non-emergency notifications
- Provide a one-click unsubscribe from every email notification

### Alert Severity Levels

Use consistent severity styling across all notification types:

| Level | Colour | Icon | Use For |
|---|---|---|---|
| **Info** | Blue | ℹ | Non-critical information, tips, system notices |
| **Success** | Green | ✓ | Completed actions, confirmations |
| **Warning** | Amber | ⚠ | Advisory — action recommended but not required |
| **Error** | Red | ✗ | Action failed; problem that must be resolved |
| **Critical** | Red (bold) |  | System-level issue requiring immediate attention |

The severity level determines persistence:
- Info and Success: auto-dismiss after 5 seconds
- Warning: persists until dismissed manually
- Error: persists until the underlying issue is resolved
- Critical: persists globally until acknowledged by an administrator

---

## 11. Accessibility

Accessibility in Awo ERP is not a compliance checklist — it is a design constraint that makes the system usable for all five personas in all their contexts: a forecourt operator with gloves, a finance manager with a repetitive strain injury navigating by keyboard, a line manager in bright sunlight on a low-quality screen.

### WCAG 2.1 AA as the Baseline

Every screen in Awo ERP must pass WCAG 2.1 AA. This is not aspirational — it is a requirement. The following are the most frequently violated criteria in enterprise software and the specific rules for Awo.

**Colour contrast**

Text and interactive elements must meet minimum contrast ratios against their backgrounds:
- Normal text (< 18pt): 4.5:1 minimum
- Large text (≥ 18pt or ≥ 14pt bold): 3:1 minimum
- UI components and graphical elements: 3:1 minimum

This is why the branding system validates tenant primary colours against contrast requirements before saving. A tenant cannot accidentally choose a primary colour that makes buttons unreadable.

**Colour is never the only indicator**

Status badges use both colour and text. Error states use both a red border and an error icon. Chart lines use both colour and pattern. A user who cannot distinguish red from green must still be able to understand status from the UI.

**Keyboard navigation**

Every interactive element — every button, link, input, dropdown, table row action — must be reachable and operable by keyboard alone.

- Tab order follows the visual reading order (left to right, top to bottom)
- Focus rings are visible and high-contrast — never hidden with `outline: none`
- Modal dialogs trap focus — Tab cycles only within the open dialog
- Dropdowns and custom controls respond to Arrow keys and Enter
- Data tables support Arrow key navigation between cells
- Shortcuts are documented in the interface (visible in the status bar, accessible via `?`)

**Semantic structure**

Every page has exactly one `<h1>` heading (the page title). Subsequent headings follow the logical document order (`<h2>`, `<h3>`) without skipping levels. Screen readers use heading structure to navigate — a flat heading hierarchy breaks this entirely.

Form inputs have explicit labels, not just placeholder text. Tables have proper `<thead>` and `<th>` elements with scope attributes. Buttons have descriptive text or `aria-label` — not just an icon.

**Screen reader support**

Dynamic content changes — a toast appearing, a filter applying, a row being added to a table — must be announced to screen readers via appropriate ARIA live regions. Forms that submit asynchronously must announce success or failure.

Status badges, icons, and visual indicators that carry meaning must have appropriate `aria-label` or descriptive text.

### Accessibility for the Kiosk / Forecourt Persona

Standard WCAG addresses visual and motor accessibility for typical desktop and mobile contexts. The forecourt persona adds additional requirements:

- **Touch target size:** Minimum 48×48px for all interactive elements. Prefer 60×80px for primary actions in kiosk mode
- **Glove-friendly spacing:** At least 8px between adjacent tap targets to prevent accidental activation
- **Sunlight readability:** Kiosk mode enforces a contrast ratio of at least 7:1 (AAA) rather than 4.5:1, because direct sunlight substantially reduces effective display contrast
- **Timeout tolerance:** Kiosk sessions do not time out during active use. An inactivity timeout (configurable, default 5 minutes) returns to a locked start screen rather than logging out — the operator can re-authenticate without losing their place in a transaction
- **Error recovery:** If an operation fails on a kiosk device (network error, device scan failure), the error message is simple, large, and actionable: "Could not save. Try again." not "HTTP 503 Service Unavailable"

---

## 12. Adoption — Making People Want to Use It

The best-designed system fails if people don't use it. In African enterprise markets, two factors uniquely influence adoption: the system must be fast on variable connectivity, and it must feel immediately useful without a training course.

### First-Run Experience

The first time a user logs in, they should feel the system is already set up for them — not a blank slate they have to configure before they can do anything.

**For employees:** The home screen immediately shows their leave balance, their next scheduled shift, and any pending items. There is nothing to configure. They can submit a leave request in the first 30 seconds.

**For line managers:** Their team is pre-loaded. Pending approvals are already surfaced. The dashboard shows their department's attendance and leave overview without any setup.

**For HR staff:** A setup checklist guides them through the key configuration steps (jurisdiction, leave types, departments, grades) in a logical sequence. Progress is saved. They can stop and return. The checklist disappears once all steps are complete.

**Onboarding tips:** Contextual tips appear the first time a user encounters each key feature — not on login, and not as a tour that blocks the UI. A small "?" badge on a new feature dismisses permanently when clicked. Tips are stored as dismissed in user preferences, so they never return.

### Speed on Variable Connectivity

Slow software is abandoned software, especially in markets where mobile data speeds are variable and 3G is common.

**Offline-first for mobile:** The employee mobile experience caches all data the user is likely to need — their own record, their leave balances, their upcoming shifts — for offline access. Clock-in/out events are queued offline and synced when connectivity returns.

**Optimistic updates:** When a user submits an action (approve a leave request, save an employee detail), the UI updates immediately as if the server confirmed it. If the server responds with an error, the UI reverts and shows the error. This makes the app feel instantaneous.

**Skeleton loading:** When data is loading, show skeleton placeholders (greyed-out shape indicators) rather than a spinner or a blank page. Skeleton loading gives the user a sense that progress is happening and reduces perceived wait time.

**Lazy loading:** Load only what is on screen. A table with 100 rows loads 25 at a time. A dashboard with 6 widgets loads the most important 2 first. Charts and heavy visualisations load last.

**Progressive enhancement:** The core functionality (view data, submit forms, approve actions) works even on a slow connection. Rich features (real-time charts, live collaboration indicators) are enhancement layers that degrade gracefully.

### Reducing Friction at Every Step

**One-click actions from notifications:** A line manager who receives a notification about a pending leave request must be able to approve it directly from the notification, without opening the full record. If they have to click more than twice, a significant percentage will not complete the action.

**Confirmation dialogs only when truly needed:** Confirmation dialogs create friction. Reserve them for irreversible or high-consequence actions: terminating an employee, reversing a posted payroll run, deleting a department with active staff. Do not ask for confirmation on reversible actions like approving a leave request or editing a salary.

**Undo instead of confirmation:** For many actions, offering an undo is a better experience than asking for confirmation. "Leave request approved — [Undo] 5s" is faster, friendlier, and equally safe for reversible actions.

**Contextual help:** Every form field that might confuse a user has a `?` tooltip that explains what the field means and how to fill it in — written in plain language, not technical terms. Every error message includes a help link to the relevant documentation.

**Progress preservation:** A user who starts a payroll run wizard and closes the browser tab can return to exactly where they left off. A user who is filling in a new employee form and the session expires finds their data intact when they log back in. Never silently lose user input.

### Measuring Adoption

The system tracks a small set of engagement signals (with user awareness, covered in privacy settings) to identify where users are dropping off:

- **Time to first meaningful action:** How long does a new user take to complete their first real task?
- **Task completion rate:** What percentage of initiated actions are completed vs abandoned?
- **Return visit frequency:** Are users coming back regularly or only when forced to?
- **Support ticket topics:** Which features generate the most confusion?
- **Feature usage rate:** Which features are used by less than 10% of the users who have access?

Features with low usage and high support tickets are redesigned, not just documented better.

---

## 13. Design Tokens & Visual Language

### Typography Scale

One typeface. One scale. No exceptions.

| Token | Size | Weight | Line Height | Use |
|---|---|---|---|---|
| `text-xs` | 11px | 400 | 1.5 | Table metadata, secondary labels |
| `text-sm` | 13px | 400 | 1.5 | Table body, form labels, secondary text |
| `text-base` | 15px | 400 | 1.6 | Primary body copy |
| `text-lg` | 17px | 500 | 1.4 | Card titles, section headings |
| `text-xl` | 20px | 600 | 1.3 | Page subtitles |
| `text-2xl` | 24px | 700 | 1.2 | Page titles |
| `text-3xl` | 30px | 700 | 1.1 | Dashboard KPI numbers |
| `text-4xl` | 36px | 800 | 1.0 | Hero numbers |

### Spacing Scale

All spacing values are multiples of 4px. Using only values from this scale ensures visual consistency without thinking about it.

`4 · 8 · 12 · 16 · 20 · 24 · 32 · 40 · 48 · 64 · 80 · 96 · 128`

### Status Colours (consistent across all modules)

| Status | Background | Text | Border | Use |
|---|---|---|---|---|
| Active / Success | Green-100 | Green-800 | Green-200 | Active employees, approved requests, successful runs |
| Pending / Warning | Amber-100 | Amber-800 | Amber-200 | Awaiting approval, pending payment |
| Draft | Grey-100 | Grey-700 | Grey-200 | Unsaved or unsubmitted records |
| Processing | Blue-100 | Blue-800 | Blue-200 | Running workflows, loading states |
| Error / Failed | Red-100 | Red-800 | Red-200 | Failed runs, rejected requests, errors |
| Suspended | Orange-100 | Orange-800 | Orange-200 | Suspended employees, paused processes |
| Terminated / Closed | Grey-200 | Grey-600 | Grey-300 | Terminal states |
| Reversed | Purple-100 | Purple-800 | Purple-200 | Reversed payroll runs, cancelled transactions |

These colours must be consistent across every module. A user who learns that green means "active" in the HR module must never encounter a screen where green means something else.

### Motion & Animation

Animation serves function, not decoration.

| Motion Type | Duration | Easing | Use |
|---|---|---|---|
| Micro-interaction | 100–150ms | ease-out | Button press, checkbox toggle |
| UI state change | 150–200ms | ease-in-out | Tab switch, dropdown open |
| Content transition | 200–300ms | ease-in-out | Page-level transitions |
| Loading skeleton | Infinite loop | linear | Placeholder shimmer |

Never animate anything that does not help the user understand what is happening. Never use animation to fill time while data loads — use it to communicate state change.

Respect the `prefers-reduced-motion` media query. Users who have configured their OS to reduce motion must see no animations. This is both an accessibility requirement and a preference that must be honoured.

---

## 14. Implementation Checklist

Use this checklist when building any new screen or feature in Awo ERP.

### Information Architecture
- [ ] The screen clearly answers: where am I, what is the state, what can I do, what is the detail
- [ ] Content is in the correct zone (navigation, context panel, workspace, status bar)
- [ ] Progressive disclosure applied — not everything is shown at Level 1
- [ ] The correct surface type is used (dashboard, list, or detail — not a hybrid)

### Data Presentation
- [ ] Default columns are appropriate for the primary persona viewing this screen
- [ ] Column alignment follows the data type rules
- [ ] Empty state is designed for both "no data exists" and "filters exclude all results"
- [ ] Density mode (comfortable / compact / kiosk) is respected

### Forms
- [ ] Single-column layout used
- [ ] Labels are above inputs, not placeholder-only
- [ ] Required vs optional fields are marked correctly
- [ ] Validation runs on blur, not only on submit
- [ ] Error messages explain what is wrong and how to fix it
- [ ] Smart defaults are pre-filled
- [ ] Multi-step forms auto-save between steps

### Tables
- [ ] Default sort is appropriate for the user's task
- [ ] Filters are prominent and persistent
- [ ] Pagination used (not infinite scroll)
- [ ] Bulk actions available when relevant
- [ ] Column selection is user-configurable and saved in preferences

### Personalisation
- [ ] Table column preferences are saved per-user per-table
- [ ] Filters persist between sessions (URL params or user preferences)
- [ ] Scope filtering applies automatically (user sees their own data by default)

### Theming
- [ ] No hardcoded colour values — all colours use semantic tokens
- [ ] Dark mode tested explicitly on this screen
- [ ] Tenant primary colour overrides look correct on this screen

### Notifications
- [ ] The correct notification type is used for each feedback scenario (inline alert / toast / notification panel / push)
- [ ] Every notification is actionable and includes enough context to act without clicking through
- [ ] Notification preferences are respected

### Accessibility
- [ ] Colour contrast meets 4.5:1 for normal text, 3:1 for large text
- [ ] Status information is not conveyed by colour alone
- [ ] All interactive elements are keyboard-reachable in logical tab order
- [ ] Focus rings are visible
- [ ] Form inputs have explicit labels
- [ ] Dynamic content changes are announced to screen readers
- [ ] Touch targets are at least 48×48px (60×80px in kiosk mode)

### Adoption
- [ ] Primary action requires the fewest possible steps
- [ ] Confirmation dialog is used only for irreversible or high-consequence actions
- [ ] Undo is offered where confirmation was avoided
- [ ] Contextual help is available for complex fields and workflows
- [ ] User input is never silently lost on session expiry

---

*This document is the authoritative UX and UI architecture specification for Awo ERP. Every new screen, component, and interaction pattern should be evaluated against the principles in this document before implementation.*
