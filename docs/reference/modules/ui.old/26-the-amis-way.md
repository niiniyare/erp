[<-- Back to Index](README.md)

## The AMIS Way

This is the most important document in the UI module. Read it before writing any schema.

The current `web/index.html` shell is powerful and works well today. But as the project grows, it will accumulate custom code solving problems AMIS already solved. This file explains what the AMIS-native approach looks like, why it is better, and exactly which parts of the current implementation drift from it.

---

### The Mental Shift

Stop thinking in components. Start thinking in **data + intentions**.

```markdown
WRONG mental model:
  "I need a sidebar, a topbar, a table, some buttons, and a filter panel."

CORRECT mental model:
  "The user on this screen is a finance clerk.
   They need to find purchase orders by date and supplier.
   They need to submit drafts and see confirmation.
   What data does that require? What actions?
   Write that — AMIS renders the rest."
```

The schema is not a UI description. It is a **declaration of intent**. AMIS translates intent into rendered UI. When you try to describe pixels and layouts instead of intent, you end up fighting AMIS.

---

### Where the Current Shell Drifts from AMIS

**Problem 1: The custom shell duplicates what AMIS `app` already does.**

The 1,400-line `index.html` implements:
- Sidebar with collapse animation
- Mobile drawer with backdrop
- Accordion menu groups
- Active item highlighting
- Breadcrumb generation
- Hash-based routing
- Schema loading from files

AMIS's `app` component does all of this natively — driven by a JSON object from Go.

**Problem 2: Menu is hardcoded in JavaScript.**

```javascript
var menuConfig = [
  { type: 'group', id: 'finance', label: 'Finance', ... }
];
```

This means: every time a feature flag changes what modules are visible, a developer has to update `index.html`. The Go backend already knows what flags are active. The menu should come from Go.

**Problem 3: Static JSON schema files have no access control.**

`web/schemas/pages/invoices.json` is a static file. Any authenticated user who knows the URL can fetch the schema for a module they shouldn't have access to. The schema itself does not contain sensitive data, but it is a correctness issue — the UI for a disabled module should not be accessible.

---

### What the AMIS-Native Architecture Looks Like

Replace the entire custom shell with this:

```html
<!-- web/index.html — the entire shell -->
<!DOCTYPE html>
<html lang="en" data-theme="system">
<head>
  <meta charset="UTF-8">
  <link rel="stylesheet" href="/sdk/sdk.css">
  <link rel="stylesheet" href="/sdk/helper.css">
  <link rel="stylesheet" href="/public/awo-theme.css">  <!-- CSS variable overrides -->
</head>
<body>
  <div id="root"></div>
  <script src="/sdk/sdk.js"></script>
  <script src="/sdk/charts.js"></script>
  <script>
    (function() {
      var amis = amisRequire('amis/embed');

      // Theme system (still needed — AMIS has no dark mode)
      applyStoredTheme();

      amis.embed(
        '#root',
        { type: 'service', schemaApi: '/schema/app' },
        {},
        {
          locale: 'en-US',
          theme:  'cxd',
          fetcher: buildFetcher(),
        }
      );

      function applyStoredTheme() {
        var theme = localStorage.getItem('awo-theme') || 'system';
        var isDark = theme === 'dark' ||
          (theme === 'system' && matchMedia('(prefers-color-scheme: dark)').matches);
        document.documentElement.classList.toggle('dark', isDark);
      }
    })();
  </script>
</body>
</html>
```

The Go `/schema/app` handler returns:

```json
{
  "type":      "app",
  "brandName": "Awo ERP",
  "logo":      "/public/logo.svg",
  "header": [
    { "type": "tpl",          "tpl": "${tenant_name}" },
    { "type": "theme-toggle" }
  ],
  "pages": [
    {
      "label": "Dashboard",
      "icon":  "fa fa-chart-line",
      "url":   "/dashboard",
      "schema": { "type": "service", "schemaApi": "/schema/dashboard" }
    },
    {
      "label": "Finance",
      "icon":  "fa fa-calculator",
      "children": [
        { "label": "Invoices", "url": "/finance/invoices",
          "schema": { "type": "service", "schemaApi": "/schema/finance/invoices" } }
      ]
    }
  ]
}
```

What Go controls by returning different `pages` arrays:
- Feature-flag-filtered nav (payroll only if `payroll.module` is on)
- Permission-filtered nav (settings only for admins)
- Tenant-specific nav ordering

The entire sidebar, routing, breadcrumbs, and mobile responsiveness: handled by AMIS `app`.

---

### The Theme Toggle Problem (and Solution)

The `app` component renders a shell we do not control. The dark mode toggle button needs to be inside the header. AMIS allows custom components via `registerRenderer`:

```typescript
// web/utils/theme-toggle-renderer.ts
import { registerRenderer, RendererProps } from 'amis';

const ThemeToggle: React.FC = () => {
  const [theme, setTheme] = React.useState(
    () => localStorage.getItem('awo-theme') || 'system'
  );

  const cycle = () => {
    const next = { system: 'light', light: 'dark', dark: 'system' }[theme];
    setTheme(next);
    localStorage.setItem('awo-theme', next);
    const isDark = next === 'dark' ||
      (next === 'system' && matchMedia('(prefers-color-scheme: dark)').matches);
    document.documentElement.classList.toggle('dark', isDark);
  };

  const icons = { system: '⬤◯', light: '☀', dark: '☾' };
  return <button onClick={cycle} className="awo-theme-toggle">{icons[theme]}</button>;
};

registerRenderer({ test: /\btheme-toggle\b/, component: ThemeToggle });
```

This one custom renderer gives us the theme toggle inside AMIS's header. Everything else is native AMIS. This is the trade-off point: write one renderer, get the whole shell for free.

---

### When to Stay with the Custom Shell

The custom shell (`web/index.html`) is the right choice if:
- You need highly custom mobile behaviour AMIS's `app` does not support
- You need a kiosk mode with a completely different layout (not the standard sidebar)
- You are prototyping and do not need flag-driven nav yet

The custom shell is the wrong choice (or signals it is time to migrate) when:
- You have to manually edit `menuConfig` every time a feature is added
- You duplicate Go permission logic in the JS menu
- You add feature flag checks to JS that already exist in Go

---

### The Right Way to Use `service`

Most developers use `crud` for everything. `service` is often the right component and is underused.

```markdown
USE crud WHEN:
  The user needs to browse, filter, and act on a collection of records.
  The component IS the list.

USE service WHEN:
  You need to load data and inject it into a static schema.
  You need a sub-schema to be loaded dynamically from Go.
  You need to compose multiple child components that share one API call.
  You need per-tenant/per-flag dynamic UI without rebuilding the page.
```

Example — dashboard widget that loads once and renders two stats:

```json
{
  "type": "service",
  "api":  "get:/api/v1/payroll/dashboard-summary",
  "body": [
    { "type": "stat", "source": "${pending_runs}",   "label": "Pending Runs" },
    { "type": "stat", "source": "${total_employees}", "label": "Active Employees" }
  ]
}
```

If you used `crud` here, you would get pagination, filter, column headers — none of which belong on a dashboard KPI card.

---

### The Right Way to Use `data` on `page`

Injecting data at the page level makes it available to ALL child components via the data chain. Use this for:
- Permissions (`can_create`, `can_approve`)
- Tenant context (`tenant_id`, `currency`, `locale`)
- The current user (`user_id`, `user_role`)

```json
{
  "type":  "page",
  "title": "Purchase Orders",
  "data": {
    "can_create":  "${can_create}",
    "can_approve": "${can_approve}",
    "currency":    "KES"
  },
  "body": { "...": "..." }
}
```

Child components anywhere in the tree can read `${can_create}` without being passed it explicitly. This is the data chain working correctly. Use it. Do not pass data down through intermediate components manually.

---

### Schema as the Only Source of Truth for UI

The schema is the UI. Not the HTML. Not the CSS. Not the JS. The moment you write JavaScript to show or hide an AMIS component based on business logic, you have a bug waiting to happen: the same logic lives in two places.

```markdown
WRONG: JS hides the "Run Payroll" button when the user has no permission
       (duplicates Go permission check, races with schema load)

CORRECT: Go injects { can_run_payroll: false } into page data.
         The button has visibleOn: "${can_run_payroll}".
         The button simply does not render.
```

One source of truth. The schema is it.

---

### Deciding When to Write a Custom Renderer

Write a custom renderer when:
1. AMIS has no component that does what you need (map picker, signature pad, barcode scanner)
2. The component will be used on multiple pages (worth the investment)
3. The custom renderer integrates INTO the AMIS data chain (reads/writes `value`)

Do not write a custom renderer when:
1. You are trying to make AMIS look different from how it renders by default — use CSS variables
2. You want to add a button to a form header — use AMIS's toolbar
3. You are fighting AMIS's layout — step back and use the correct component instead

---
