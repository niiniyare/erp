[<-- Back to Index](README.md)

> **⚠️ MIXED STATUS.** Parts of this document describe implemented features; others are roadmap.
>
> | Feature | Status |
> |---------|--------|
> | User preferences (`Pref()` / `Setting()` access) | `[IMPLEMENTED]` — pre-loaded at login into `UISessionContext` |
> | Feature flag snapshot at login | `[IMPLEMENTED]` — `FeatureEnabled()` on `UISessionContext` |
> | Tenant currency/timezone/locale config | `[IMPLEMENTED]` — via `contract.SessionContext` |
> | Saved views / filter persistence | `[ROADMAP]` — not implemented |
> | Module navigation ordering | `[ROADMAP]` — not implemented |
> | Recent items / search history | `[ROADMAP]` — not implemented |
> | Dashboard widget layout | `[ROADMAP]` — not implemented |
>
> For implemented preference access, see [Glossary — UISessionContext](../appendices/A-glossary.md#uisessioncontext).

## Personalisation

### Two Types of Personalisation

```markdown
AUTOMATIC (system learns from behaviour — zero user effort)
├── Scope filtering: lists auto-scoped to user's organisational unit
├── Recent items: search dropdown shows recently accessed records first
└── Adaptive columns: if a user always adds a column, it becomes their default

EXPLICIT (user configures deliberately)
├── Colour scheme: Light / Dark / System
├── Data density: Comfortable / Compact / Kiosk
├── Language: available locales for the tenant
├── Date format: DD/MM/YYYY, MM/DD/YYYY, YYYY-MM-DD
├── Number format: 1,234.56 / 1.234,56 / 1 234,56
├── Currency display: Symbol (KES) / Code / Both
├── Default landing page: Task list / Dashboard / Last visited
├── Sidebar state: Expanded / Collapsed / Remember last
├── Table column selection (per table)
├── Table sort and page size (per table)
└── Saved filter views (per table)
```

### User Preference Storage

Stored as JSONB in a `user_preferences` column — schema stays stable as new preferences are added:

```json
{
  "appearance": {
    "colour_scheme": "dark",
    "density":       "compact",
    "font_size":     "default"
  },
  "locale": {
    "language":       "en-KE",
    "date_format":    "DD/MM/YYYY",
    "number_format":  "1,234.56"
  },
  "navigation": {
    "default_landing": "task_list",
    "sidebar_state":   "expanded",
    "module_order":    ["hr", "finance", "inventory"]
  },
  "tables": {
    "hr.employees": {
      "columns":   ["photo", "name", "department", "grade", "status", "join_date"],
      "sort":      { "column": "last_name", "direction": "asc" },
      "page_size": 50
    }
  },
  "notifications": {
    "in_app":       "all",
    "email_digest": "daily",
    "push":         "urgent_only"
  }
}
```

Preferences are loaded once at session start, cached client-side, and updated asynchronously. The UI updates immediately without waiting for the server response.

### Saved Views (Highest-Value Feature)

A saved view is a named combination of filters + column selection + sort order. This is the single highest-leverage personalisation feature for power users.

```markdown
MY SAVED VIEWS — Employee List
  ★ Active Forecourt Staff          [Default]
  ○ Pending Onboarding
  ○ Terminated This Year
  ○ Grade G3 and above

                           [+ Save current view]
```

Saved views can be:
- **Private** — visible only to the user who created it
- **Shared** — shared with a role or department (e.g. HR manager creates "Due for Probation Review" shared with all line managers)

Implementation: Save filter state to user preferences JSONB. On load, check for a saved view matching the current table and apply it.

### Theme Preference

The theme system in `index.html` already persists the user's choice in `localStorage`:

```javascript
function getStoredTheme() {
  try { return localStorage.getItem('awo-theme') || 'system'; }
  catch (e) { return 'system'; }
}
```

This is per-browser, not per-user-account. For true cross-device persistence, sync this to `user_preferences.appearance.colour_scheme` via the preferences API when the user changes the theme.

---
