---
title: Dark Theme Guide
portal: 5 — Frontend Engineering
section: 05-frontend-engineering
audience: [frontend-engineer]
related:
  - "[Frontend Overview](01-frontend-overview.md)"
  - "[AMIS Schema Guide](02-amis-schema-guide.md)"
  - "[Component Patterns](04-component-patterns.md)"
---

# Dark Theme Guide

## The AMIS Dark Theme Problem

AMIS SDK's `theme("dark")` uses `classPrefix: "dark-"` but ships **zero CSS rules** for the `dark-` prefix. Calling `theme("dark")` breaks all styling — buttons, forms, tables all lose their appearance.

**Do not use `theme("dark")`.**

## Correct Approach: CSS Custom Property Override

AMIS components use CSS custom property tokens throughout. Override the tokens on `html.dark` — all components inherit automatically.

### Token Scale Convention

AMIS neutral fill/text/line tokens use 1–11 scale:
- `1` = darkest
- `11` = lightest

In light mode, `fill-1` is near-white (light background). In dark mode, invert so `fill-1` is near-black.

### Implementation

In `web/styles/dark-theme.css`:

```css
/* ── Dark mode token overrides ─────────────────────────────────── */
html.dark {
  /* Backgrounds — inverted scale */
  --colors-neutral-fill-1: #1a1a1a;
  --colors-neutral-fill-2: #242424;
  --colors-neutral-fill-3: #2e2e2e;
  --colors-neutral-fill-4: #383838;
  --colors-neutral-fill-5: #424242;
  --colors-neutral-fill-6: #4c4c4c;
  --colors-neutral-fill-7: #616161;
  --colors-neutral-fill-8: #757575;
  --colors-neutral-fill-9: #9e9e9e;
  --colors-neutral-fill-10: #bdbdbd;
  --colors-neutral-fill-11: #e0e0e0;

  /* Text */
  --colors-neutral-text-1: #ffffff;
  --colors-neutral-text-2: #e0e0e0;
  --colors-neutral-text-3: #bdbdbd;
  --colors-neutral-text-4: #9e9e9e;
  --colors-neutral-text-5: #757575;

  /* Lines / borders */
  --colors-neutral-line-1: #333333;
  --colors-neutral-line-2: #3d3d3d;
  --colors-neutral-line-3: #474747;
  --colors-neutral-line-8: #616161;

  /* Page + panel backgrounds */
  --background: #121212;
  --body-bg: #121212;
  --Page-main-bg: #1a1a1a;
  --Panel-bg-color: #1e1e1e;
  --Table-bg: #1e1e1e;
  --Table-onHover-bg: #2a2a2a;

  /* Form inputs */
  --Form-input-bg: #2a2a2a;
  --Form-input-borderColor: #424242;
  --Form-input-color: #e0e0e0;
  --Form-input-placeholderColor: #616161;

  /* Sidebar / nav */
  --Nav-bg: #161616;
  --Nav-item-color: #bdbdbd;
  --Nav-item-onHover-bg: #2a2a2a;
  --Nav-item-active-bg: #1a3a5c;
  --Nav-item-active-color: #4da6ff;
}
```

### Toggle Logic

Toggle dark mode by adding/removing the `dark` class on `<html>`:

```javascript
// web/pages/index.html (inline script)
const stored = localStorage.getItem('theme');
const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
if (stored === 'dark' || (!stored && prefersDark)) {
    document.documentElement.classList.add('dark');
}

function toggleTheme() {
    const isDark = document.documentElement.classList.toggle('dark');
    localStorage.setItem('theme', isDark ? 'dark' : 'light');
}
```

### AMIS Renderer Initialization

Pass `theme: "cxd"` (the default theme) regardless of dark/light mode:

```javascript
amis.embed('#app', schema, initialData, {
    theme: 'cxd',        // always cxd — dark handled via CSS tokens
    locale: 'en-US',
});
```

## What NOT To Do

```css
/* ❌ Never use !important on .cxd-* classes — whack-a-mole nightmare */
.cxd-Panel {
    background: #1e1e1e !important;
}

/* ❌ Never override individual component backgrounds directly */
.cxd-Table-content {
    background-color: #2a2a2a !important;
}
```

Override tokens at root → components inherit → no specificity wars.

## Testing Dark Mode

1. Open browser devtools → Elements → add `dark` class to `<html>` manually
2. Check: tables, forms, panels, modals, dropdowns, tooltips
3. Check text contrast — all text must meet WCAG AA (4.5:1 for normal, 3:1 for large)
4. Check focus rings are visible on dark backgrounds

## Common Issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| White flash on page load | Theme applied after render | Apply `dark` class in `<head>` synchronously (not deferred) |
| Some modals stay white | Modal renders outside `html.dark` scope | Confirm modal container is inside `<body>` → inherits `html.dark` |
| Input placeholder invisible | Missing `--Form-input-placeholderColor` override | Add to token overrides |
| Charts not themed | ECharts uses its own theme | Pass dark palette to `chartRef.setOption({ backgroundColor, ... })` |
