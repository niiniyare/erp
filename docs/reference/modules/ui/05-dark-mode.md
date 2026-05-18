[<-- Back to Index](README.md)

> Last verified: 2026-05-18 | Code pointer: `web/pages/index.html` (lines 41–974)

## Dark Mode

### How It Works

Dark mode activates when JavaScript adds `dark` class to `<html>`. Three CSS layers switch simultaneously:

```
LAYER 1 — Shell variables (sidebar, topbar, footer)
  html.dark { --sidebar-bg: #1e1e2d; --text-primary: #e4e4e7; ... }

LAYER 2 — AMIS neutral scale inversion (components inside #content)
  html.dark { --colors-neutral-fill-11: #1e1e2d; ... }
  Full 11-step scale inverted for fill, text, and line.

LAYER 3 — Portal reinforcement (popups, modals, dropdowns appended to <body>)
  html.dark .cxd-PopOver { background: ... }
  html.dark .cxd-Modal-content { background: ... }
  etc. — explicit selectors because portals live outside #content
```

**All three layers are required.** Missing Layer 3 is the most common incomplete dark mode implementation — dropdowns, date pickers, and modals appear light-themed while the rest of the page is dark.

---

### Why AMIS Has No Built-in Dark Mode

AMIS ships a `theme("dark")` option that uses `classPrefix: "dark-"`. This generates CSS classes like `.dark-cxd-Button`. **There are zero CSS rules for this prefix in `sdk.css`.** Using it destroys all component styling.

**The correct approach:** Override AMIS's CSS custom property scales at `html.dark`.

---

### Layer 1 — Shell CSS Variables

Source: `web/pages/index.html:41–62`

```css
html.dark {
    --sidebar-bg: #1e1e2d;
    --sidebar-border: #2e2e3e;
    --sidebar-active: #1a2744;
    --sidebar-active-border: #4dabf7;
    --sidebar-hover: #262637;
    --text-primary: #e4e4e7;
    --text-secondary: #9a9ab0;
    --text-muted: #6b6b80;
    --bg-main: #151521;
    --bg-surface: #1e1e2d;
    --bg-input: #262637;
    --search-border: #2e2e3e;
    --scrollbar-thumb: #3a3a4e;
    --tooltip-bg: #e4e4e7;
    --tooltip-color: #1e1e2d;
    --backdrop-bg: rgba(0, 0, 0, 0.6);
    --shadow-sidebar: 4px 0 24px rgba(0, 0, 0, 0.4);
    --card-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
    --card-shadow-hover: 0 4px 20px rgba(0, 0, 0, 0.4);
    color-scheme: dark;
}
```

---

### Layer 2 — AMIS Neutral Scale Inversion

AMIS uses three scales numbered 1 (darkest) to 11 (lightest). In dark mode: **invert** — 11 becomes dark, 1 becomes near-white.

Source: `web/pages/index.html:741–894`

#### Neutral Fill (backgrounds)

```css
html.dark {
    --colors-neutral-fill-11: #1e1e2d;  /* was #ffffff  → surface */
    --colors-neutral-fill-10: #262637;  /* was #f7f8fa  → input/hover */
    --colors-neutral-fill-9:  #2e2e3e;  /* was #f2f3f5 */
    --colors-neutral-fill-8:  #363648;  /* was #e8e9eb */
    --colors-neutral-fill-7:  #3e3e52;  /* was #d4d6d9 */
    --colors-neutral-fill-6:  #52526a;  /* was #b8babf */
    --colors-neutral-fill-5:  #6b6b80;  /* was #84878c */
    --colors-neutral-fill-4:  #9a9ab0;  /* was #5c5f66 */
    --colors-neutral-fill-3:  #b8b8cc;  /* was #303540 */
    --colors-neutral-fill-2:  #d4d4e0;  /* was #151b26 */
    --colors-neutral-fill-1:  #e4e4e7;  /* was #070c14 */
    --colors-neutral-fill-12: #1a2744;  /* selection/highlight bg */
}
```

#### Neutral Text (foreground)

```css
html.dark {
    --colors-neutral-text-11: #1e1e2d;  /* was #ffffff */
    --colors-neutral-text-10: #262637;  /* was #f7f8fa */
    --colors-neutral-text-9:  #2e2e3e;  /* was #f2f3f5 */
    --colors-neutral-text-8:  #363648;  /* was #e8e9eb */
    --colors-neutral-text-7:  #3e3e52;  /* was #d4d6d9 */
    --colors-neutral-text-6:  #6b6b80;  /* was #b8babf → muted */
    --colors-neutral-text-5:  #9a9ab0;  /* was #84878c */
    --colors-neutral-text-4:  #b8b8cc;  /* was #5c5f66 */
    --colors-neutral-text-3:  #d0d0db;  /* was #303540 */
    --colors-neutral-text-2:  #e4e4e7;  /* was #151b26 → primary text */
    --colors-neutral-text-1:  #f0f0f3;  /* was #070c14 */
}
```

#### Neutral Line (borders)

```css
html.dark {
    --colors-neutral-line-11: #1e1e2d;  /* was #ffffff */
    --colors-neutral-line-10: #262637;  /* was #f7f8fa */
    --colors-neutral-line-9:  #2e2e3e;  /* was #f2f3f5 */
    --colors-neutral-line-8:  #363648;  /* was #e8e9eb → default border */
    --colors-neutral-line-7:  #3e3e52;  /* was #d4d6d9 */
    --colors-neutral-line-6:  #52526a;  /* was #b8babf */
    --colors-neutral-line-5:  #6b6b80;  /* was #84878c */
    --colors-neutral-line-4:  #9a9ab0;  /* was #5c5f66 */
    --colors-neutral-line-3:  #b8b8cc;  /* was #303540 */
    --colors-neutral-line-2:  #d4d4e0;  /* was #151b26 */
    --colors-neutral-line-1:  #e4e4e7;  /* was #070c14 */
}
```

#### Core AMIS Aliases

These are the component-specific tokens AMIS reads directly. Setting these overrides component backgrounds regardless of the neutral scale:

```css
html.dark {
    --background: #151521;
    --body-bg: #151521;
    --body-color: #e4e4e7;
    --text-color: #e4e4e7;
    --white: #1e1e2d;         /* AMIS uses --white as default bg — invert it */
    --light: #262637;
    --Page-main-bg: #151521;
    --Panel-bg-color: #1e1e2d;
    --Panel-heading-bg-color: #1e1e2d;
    --Panel-footer-bg-color: #1e1e2d;
    --Table-bg: #1e1e2d;
    --Table-thead-bg: #262637;
    --Table-strip-bg: #1a1a28;
    --Table-onHover-bg: #262637;
    --Card-bg: #1e1e2d;
    --Modal-bg: #1e1e2d;
    --Drawer-bg: #1e1e2d;
    --Drawer-header-bg: #1e1e2d;
    --Tabs-content-bg: #1e1e2d;
    --Form-input-bg: #262637;
    --Form-input-borderColor: #2e2e3e;
    --Form-input-onFocus-borderColor: #4dabf7;
    --Select-input-bg: #262637;
    --Select-menu-bg: #1e1e2d;
    --DatePicker-bg: #1e1e2d;
    --DatePicker-cell-bg: #262637;
    --PopOver-bg: #1e1e2d;
    --PopOver-borderColor: #2e2e3e;
    --Tooltip-bg: #2e2e3e;
    --Tooltip-color: #e4e4e7;
    --DropDown-menu-bg: #1e1e2d;
    --boxShadow: 0 2px 12px rgba(0, 0, 0, 0.4);
    --borderColor: #2e2e3e;
}
```

---

### Layer 3 — Portal Reinforcement

**This is the most commonly missed layer.** AMIS portals (dropdowns, date pickers, modals, drawers, tooltips) are appended directly to `<body>`, outside `#content`. CSS custom properties cascade from `html.dark` to `body`, so the variable values are correct — but AMIS components sometimes bypass variables and apply inline styles or class-based backgrounds.

Explicit selectors are required as a safety net:

Source: `web/pages/index.html:902–960`

```css
/* PopOver — base for many components */
html.dark .cxd-PopOver,
html.dark .cxd-PopOverAble-popover {
    background: var(--PopOver-bg);
    border-color: var(--PopOver-borderColor);
    color: var(--text-color);
}

/* Dropdowns */
html.dark .cxd-DropDown-popover,
html.dark .cxd-DropDown-menu {
    background: var(--DropDown-menu-bg);
    border-color: var(--DropDown-menu-borderColor);
    color: var(--text-color);
}
html.dark .cxd-DropDown-menuItem:hover {
    background: var(--DropDown-menuItem-onHover-bg);
}

/* Date picker */
html.dark .cxd-DatePicker-popover {
    background: var(--DatePicker-bg);
    color: var(--text-color);
}

/* Modals */
html.dark .cxd-Modal-content {
    background: var(--Modal-bg);
    color: var(--text-color);
}

/* Drawers */
html.dark .cxd-Drawer-content {
    background: var(--Drawer-bg);
    color: var(--text-color);
}

/* Tooltips */
html.dark .cxd-Tooltip-body {
    background: var(--Tooltip-bg);
    color: var(--Tooltip-color);
}

/* Select menus */
html.dark .cxd-Select-menu {
    background: var(--Select-menu-bg);
    color: var(--text-color);
}
html.dark .cxd-Select-option:hover {
    background: var(--Select-option-onHover-bg);
}
html.dark .cxd-Select-option.is-active {
    background: var(--Select-option-onActive-bg);
}
```

---

### Theme Toggle Implementation

Source: `web/pages/index.html` (JS section)

Dark mode persists via `localStorage`:

```javascript
// Read saved preference
const savedTheme = localStorage.getItem('theme') || 'system';
// Apply on load
applyTheme(savedTheme);

function applyTheme(mode) {
    const isDark = mode === 'dark' ||
        (mode === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);
    document.documentElement.classList.toggle('dark', isDark);
}

// Toggle
function toggleTheme() {
    const current = localStorage.getItem('theme') || 'system';
    const next = current === 'dark' ? 'light' : 'dark';
    localStorage.setItem('theme', next);
    applyTheme(next);
}
```

`data-theme="system"` attribute on `<html>` is the initial state before JS runs (avoids FOUC).

---

### Safe Area Support (Notch Devices)

Source: `web/pages/index.html:82–86`

```css
@supports (padding: env(safe-area-inset-bottom)) {
    body {
        padding-bottom: env(safe-area-inset-bottom);
    }
}

/* Mobile: footer safe area */
#footer {
    padding-bottom: max(16px, calc(env(safe-area-inset-bottom, 0px) + 8px));
}
```

---

### Approaches That Don't Work

```
❌ theme: 'dark' in embed() call
   → AMIS generates .dark-cxd-* classes but sdk.css has zero rules for them
   → All component styling breaks

❌ Overriding individual .cxd-* backgrounds with !important everywhere
   → Whack-a-mole: every AMIS SDK update breaks different components
   → 500+ classes to maintain

❌ Separate dark stylesheet loaded after sdk.css
   → Hard to maintain as sdk.css evolves
   → Still requires portal selectors (Layer 3) anyway
```

---

### Testing Checklist

Every new page schema must be tested in dark mode. Open DevTools and run:
```javascript
document.documentElement.classList.add('dark');
```

Check these components explicitly:

| Component | What to check |
|-----------|--------------|
| Table/CRUD | Row hover, header background, stripe rows |
| Panel/Card | Panel heading, body background |
| Forms | Input backgrounds, focus border, placeholder color |
| Select dropdowns | Menu background, option hover |
| Date pickers | Calendar popup background, cell hover |
| Modals/Dialogs | Modal content background, overlay backdrop |
| Drawers | Drawer panel background |
| Tooltips | Tooltip background and text contrast |
| Charts | Background must be transparent (not white) |
| Custom `tpl` content | Hardcoded `color: #333` or `background: white` will break |

---

### Tenant Brand Colour in Dark Mode

The primary brand colour (`#2196f3` default) needs a lighter variant for dark backgrounds:

```css
:root {
    --colors-brand-5: #2196f3;   /* light mode primary */
    --colors-brand-6: #1976d2;   /* light mode hover */
}

html.dark {
    --colors-brand-5: #4dabf7;   /* lighter for dark bg contrast */
    --colors-brand-6: #339af0;
}
```

When tenant-configurable branding is implemented, generate both variants from the tenant's primary hex and override these variables.

---
