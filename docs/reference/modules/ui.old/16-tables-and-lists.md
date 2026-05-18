[<-- Back to Index](README.md)

## Tables & Lists

Tables are the dominant UI pattern in ERP. These rules prevent the most common mistakes.

### Every Column Must Earn Its Place

Before adding a column: **does the user need to compare this value across rows?** If no, it belongs in the detail view, not the table.

**Default columns differ by persona:**

```markdown
Employee List — HR Admin (power user):
  Photo | Name | Employee# | Department | Grade | Status | Join Date | Actions

Employee List — Line Manager:
  Photo | Name | Status | Shift Today | Leave Balance | Actions

Employee List — Employee view (team only):
  Photo | Name | Department | Contact | Actions
```

This is relevance control, not access control. The manager can add columns — but the default shows what they actually need.

### Column Alignment Rules

| Data Type | Alignment | Format |
|---|---|---|
| Text (names, descriptions) | Left | As-is |
| Numbers (counts, quantities) | Right | Locale-aware |
| Currency amounts | Right | With currency code, 2 decimal places |
| Percentages | Right | e.g. 14.5% |
| Dates | Left or center | Relative for recent ("2 days ago"), absolute for older ("14 Nov 2024") |
| Status | Center | Badge/pill with colour |
| Boolean | Center | Icon (check/cross) — never "true/false" |
| Actions | Right | Icon buttons or "..." menu |

Never mix alignment in a column — header, rows, and totals must all match.

### Status Badges

Status columns must use both colour AND text. Never colour alone (accessibility):

```json
{
  "name":     "status",
  "label":    "Status",
  "type":     "tag",
  "colorMap": {
    "active":     "success",
    "pending":    "warning",
    "draft":      "default",
    "processing": "processing",
    "cancelled":  "error",
    "suspended":  "warning",
    "terminated": "default"
  }
}
```

### Default Sort

Default sort must be the most useful for the user's task, not whatever the database returns:

```markdown
Payroll runs:      newest first     (orderBy: created_at, orderDir: desc)
Employees:         alphabetical     (orderBy: last_name,  orderDir: asc)
Leave requests:    oldest pending   (oldest item needing action — sort by submitted_at asc)
Invoices:          newest first     (orderBy: invoice_date, orderDir: desc)
Audit log:         most recent      (orderBy: occurred_at, orderDir: desc)
```

### Filters

Filters are first-class. They are not hidden in a dropdown or an "Advanced" section:

```markdown
Filter design rules:
✅ Primary filters (status, date range) always visible above the table
✅ Filters persist across page refreshes (URL params or user preferences)
✅ Show active filter count: "Filters (3)"
✅ Single "Clear all filters" button
✅ Filter options show counts: "Active (41) · Pending (3) · Terminated (12)"
```

AMIS `crud` with `syncLocation: true` stores filter state in the URL hash automatically — browser back/forward works, and sharing the URL preserves the filters.

### Pagination vs Infinite Scroll

**Always use pagination for tables. Never infinite scroll.**

Infinite scroll problems in enterprise context:
- Cannot jump to "page 5 of 12"
- URL doesn't represent position — sharing a link loses context
- Bulk selection becomes confusing as the list grows
- DOM performance degrades

Use `"footerToolbar": ["statistics", "pagination"]` on all `crud` components.

Exception: infinite scroll is fine for feed-style displays (activity log, notifications) where position is less important.

### Bulk Actions

```markdown
Pattern:
  ☑ 3 of 47 selected    [Approve] [Export] [Send Payslip] [···]

Rules:
✅ Bulk action bar appears ONLY when ≥1 row is selected
✅ "Select all on this page" and "Select all 847 results" are distinct
✅ Destructive bulk actions require confirmation with count:
   "Terminate 3 employees. This cannot be undone."
✅ Show result summary after completion:
   "3 approved. 1 failed — Amina Odhiambo (insufficient balance)."
```

### Empty States

Every list must have a designed empty state — an empty table with just headers looks broken.

```markdown
TWO DIFFERENT EMPTY STATES for the same list:

STATE 1 — No records exist:
┌──────────────────────────────────────┐
│  No payroll runs yet                 │
│  ─────────────────────────────────   │
│  Start your first payroll run for    │
│  November 2024. Awo will guide you   │
│  through step by step.               │
│                                      │
│           [Start Payroll Run]        │
└──────────────────────────────────────┘

STATE 2 — Filters exclude everything:
┌──────────────────────────────────────┐
│  No results match your filters       │
│  ─────────────────────────────────   │
│  Try removing some filters or        │
│  expanding the date range.           │
│                                      │
│           [Clear Filters]            │
└──────────────────────────────────────┘
```

In AMIS, use the `crud` `placeholder` property. For the filter-active empty state, use `syncLocation: true` — the "clear filters" action simply navigates to the base URL.

---
