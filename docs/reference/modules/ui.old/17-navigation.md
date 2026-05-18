[<-- Back to Index](README.md)

## Navigation

### Global Navigation (Sidebar)

The sidebar is the constant backbone. Always visible on desktop. Always reachable on mobile (hamburger → slide-in drawer with backdrop).

**Structure:**
```markdown
Sidebar
  ├── Brand (logo + name + collapse toggle)
  ├── Search box (cosmetic — full search TBD)
  └── Menu
        ├── Dashboard (top-level item)
        ├── Finance (group)
        │     ├── Invoices
        │     ├── Payments
        │     └── Reports
        ├── HR (group)
        │     ├── Employees
        │     └── Payroll Runs
        └── Settings (group)
```

**Maximum 7 top-level items.** Navigation research shows 7 is the practical maximum before users stop scanning and start searching. Group into categories beyond that.

**Accordion behaviour:** Only one group expanded at a time. Clicking a group item collapses all others automatically.

**Active state:** Always explicit. The user must never wonder which section they are in.

### Adding Items to the Sidebar

Edit the `menuConfig` array in `web/index.html`:

```javascript
var menuConfig = [
  {
    type: 'item',
    id:   'dashboard',
    label: 'Dashboard',
    icon:  'fa-chart-line',
    hash:  '#dashboard'
  },
  {
    type: 'group',
    id:   'finance',
    label: 'Finance',
    icon:  'fa-calculator',
    expanded: false,
    items: [
      { id: 'invoices',   label: 'Invoices',         hash: '#invoices'   },
      { id: 'bills',      label: 'Bills',             hash: '#bills'      },
      { id: 'payments',   label: 'Payments',          hash: '#payments'   },
      { id: 'fin-reports',label: 'Reports',           hash: '#fin-reports'}
    ]
  }
];
```

The `id` must match the JSON schema filename: `id: 'invoices'` → loads `web/schemas/pages/invoices.json`.

### Routing

All navigation is hash-based (`#dashboard`, `#invoices`). The shell JS:

1. Detects `hashchange`
2. Unmounts previous AMIS instance
3. Loads new schema from `web/schemas/pages/{route}.json`
4. Mounts AMIS in `#content`
5. Updates active sidebar item + breadcrumb

No server round-trip for page switches.

### Breadcrumb

The breadcrumb in the topbar automatically updates based on the current route:

```markdown
Single-level item:    Dashboard
Grouped item:         Finance  ›  Invoices
```

On mobile, only the immediate parent is shown (`← Finance`). The breadcrumb HTML is built by `updateBreadcrumb()` in `index.html`.

### "What Needs My Attention" Entry Point

The dashboard (default landing page) must surface pending items personalised to the current user. This is not a generic dashboard — it is a task list:

```markdown
YOUR ACTION ITEMS TODAY                            3 pending

  [ ! ] Leave Request                  Approve or Decline
        Amina Odhiambo · 23–31 Dec · Annual Leave
        Submitted 2 days ago

  [ ! ] Timesheet Approval             Approve or Decline
        Forecourt Team · November 2024
        7 employees pending

  [ ✓ ] Payroll Run Ready for Review   Review Now
        November 2024 · 47 employees · KES 4.2M gross
```

Implement this as a `crud` component querying `/api/v1/tasks/pending` — returns the user's scoped pending items across all modules.

### Module Navigation Order

```markdown
RECOMMENDED ORDERING (most frequently used first):
1. Dashboard
2. Finance / Accounting
3. HR / People
4. Purchasing
5. Inventory
6. Sales
7. Settings

Items hidden if module feature flag is off.
```

### Mobile Navigation

On screen widths ≤767px:
- Sidebar is offscreen (transforms left: -100%)
- Hamburger button in topbar slides it in
- Clicking backdrop closes it
- Navigating to a new page closes it automatically

Collapsed state on desktop does not apply on mobile — the sidebar always shows full-width when opened.

---
