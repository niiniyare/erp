> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Mobile Considerations
portal: 5 — Frontend Engineering
section: 05-frontend-engineering
audience: [frontend-engineer]
related:
  - "[Frontend Overview](01-frontend-overview.md)"
  - "[Component Patterns](04-component-patterns.md)"
  - "[Schema Conventions](05-schema-conventions.md)"
---

# Mobile Considerations

## Current State

AwoERP's primary frontend is web-based using AMIS SDK. Mobile access is currently via responsive web only — no native mobile app exists. These guidelines apply to making the web frontend usable on mobile browsers and tablets.

## Responsive Layout

AMIS provides responsive grid components. Use `flex` or `grid` containers with breakpoints:

```json
{
  "type": "grid",
  "columns": [
    {
      "md": 6,
      "xs": 12,
      "body": [{ "type": "panel", "title": "Contract Details", "body": [...] }]
    },
    {
      "md": 6,
      "xs": 12,
      "body": [{ "type": "panel", "title": "Line Items", "body": [...] }]
    }
  ]
}
```

Column widths use Bootstrap-like breakpoints (`xs` = mobile, `md` = tablet+, `lg` = desktop).

## CRUD on Mobile

Full CRUD tables are often too wide for mobile. Use AMIS `cards` view for mobile-friendly list display:

```json
{
  "type": "crud",
  "defaultViewMode": "cards",
  "switchDefaultView": true,
  "cardBodyFields": ["contract_number", "title", "status", "total_value"],
  "columns": [...]
}
```

Provide both table view (desktop) and card view (mobile), toggleable via the `switchDefaultView` control.

## Touch-Friendly Interactions

- Buttons must be at least 44×44px touch target — use `size: "lg"` for primary action buttons
- Confirmation dialogs should have large tap targets for confirm/cancel
- Avoid hover-only interactions — all interactive elements must work on tap
- Dropdown menus should use full-width on mobile (`className: "w-full"`)

## Forms on Mobile

Keep forms minimal on mobile:

```json
{
  "type": "form",
  "mode": "normal",
  "wrapWithPanel": true,
  "body": [
    {
      "type": "input-text",
      "name": "contract_number",
      "label": "Contract Number",
      "inputClassName": "w-full"
    }
  ]
}
```

Use `mode: "normal"` (stacked layout) on mobile vs `mode: "horizontal"` on desktop. Detect via CSS and switch with `className`.

## Navigation on Mobile

The sidebar navigation (`web/pages/index.html`) should collapse on small screens:

```javascript
// Auto-collapse sidebar on mobile
function initMobileNav() {
    const sidebar = document.querySelector('.sidebar');
    const menuBtn = document.querySelector('.menu-toggle');

    menuBtn.addEventListener('click', () => {
        sidebar.classList.toggle('sidebar--collapsed');
    });

    // Auto-collapse on small screens
    if (window.innerWidth < 768) {
        sidebar.classList.add('sidebar--collapsed');
    }
}
```

## Date Inputs

Use AMIS `input-date` which renders the native mobile date picker on iOS/Android:

```json
{
  "type": "input-date",
  "name": "start_date",
  "label": "Start Date",
  "format": "YYYY-MM-DD",
  "inputFormat": "YYYY-MM-DD"
}
```

AMIS's date picker falls back to native `<input type="date">` on mobile, which shows the platform date wheel.

## File Upload on Mobile

CSV import supports camera/document upload on mobile via native file picker:

```json
{
  "type": "input-file",
  "name": "file",
  "label": "Upload CSV",
  "accept": ".csv,text/csv",
  "maxSize": 10485760,
  "btnLabel": "Choose File or Take Photo"
}
```

`accept` controls which file types are offered in the mobile picker.

## Performance on Mobile

- Keep initial schema payload < 50KB — mobile networks are slower
- Paginate aggressively — `page_size: 10` on mobile vs 20 on desktop
- Lazy-load sub-panels — use `service` component with `initFetch: false` and load on demand
- Avoid rendering large tables on initial load — filter first, then show results

## Testing on Mobile

Before shipping a new page:
1. Test in Chrome DevTools mobile simulation (iPhone SE + iPad)
2. Test actual tap interactions — buttons, forms, dropdowns
3. Test with slow network (3G simulation in DevTools)
4. Check that all text is readable without zooming (min 16px base font)

## Viewport Meta Tag

Ensure the HTML template has the viewport meta tag:

```html
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0">
```

Without this, mobile browsers zoom out to show the desktop layout.
