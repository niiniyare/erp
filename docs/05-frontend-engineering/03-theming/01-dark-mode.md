---
title: Dark Mode Implementation
portal: 5 — Frontend Engineering
section: 05-frontend-engineering
audience: [frontend-engineer]
related:
  - "[Web UI Architecture](../01-web-ui-architecture/01-web-ui-overview.md)"
  - "[AMIS Schema Overview](../02-amis-schemas/01-amis-schema-overview.md)"
---

# Dark Mode Implementation

## How AMIS Theming Works

AMIS SDK uses CSS custom properties (variables) prefixed with `--colors-neutral-*`, `--background`, and component-specific vars like `--Panel-bg-color`. All component styles reference these variables.

AMIS has no built-in dark theme. Calling `theme("dark")` only sets `classPrefix: "dark-"` — there are no CSS rules for that prefix. Dark mode must be implemented by overriding the CSS variable tokens.

## Variable Override Strategy

Define overrides on `html.dark`. Every AMIS component inherits via CSS cascade:

```css
/* web/styles/theme.css */

/* ─── Light Mode (defaults — usually fine as-is) ─── */
:root {
  --brand-primary: #2563eb;
  --brand-primary-hover: #1d4ed8;
}

/* ─── Dark Mode ─── */
html.dark {
  /* Neutral fills: 1=darkest surface, 11=lightest (inverted from light) */
  --colors-neutral-fill-1:  #f8f9fa;   /* lightest surface in light → text in dark */
  --colors-neutral-fill-2:  #e9ecef;
  --colors-neutral-fill-3:  #dee2e6;
  --colors-neutral-fill-4:  #ced4da;
  --colors-neutral-fill-5:  #adb5bd;
  --colors-neutral-fill-6:  #6c757d;
  --colors-neutral-fill-7:  #495057;
  --colors-neutral-fill-8:  #343a40;
  --colors-neutral-fill-9:  #2a2f34;
  --colors-neutral-fill-10: #212529;
  --colors-neutral-fill-11: #1a1d21;  /* darkest surface */

  /* Text */
  --colors-neutral-text-2: #e8e8e8;
  --colors-neutral-text-3: #c0c0c0;
  --colors-neutral-text-4: #909090;

  /* Lines / borders */
  --colors-neutral-line-8: #3a3a3a;

  /* Page backgrounds */
  --background:       #121212;
  --body-bg:          #1e1e1e;
  --Page-main-bg:     #1e1e1e;
  --Page-aside-bg:    #161b22;

  /* Component surfaces */
  --Panel-bg-color:   #2a2a2a;
  --Table-bg:         #2a2a2a;
  --Table-tr-bg:      #2a2a2a;
  --Table-tr-hover-bg: #333333;

  /* Form inputs */
  --Form-input-bg:        #1e1e1e;
  --Form-input-border:    #3a3a3a;
  --Form-input-color:     #e8e8e8;
  --Form-input-focusBorder: #2563eb;

  /* Sidebar */
  --Nav-item-bg:        #161b22;
  --Nav-item-hover-bg:  #21262d;
  --Nav-item-active-bg: #1f3a5f;
  --Nav-item-color:     #c9d1d9;

  /* Buttons */
  --Button-bg:          #21262d;
  --Button-border:      #30363d;
  --Button-color:       #c9d1d9;
}
```

## Toggle Implementation

```javascript
// web/pages/index.html

function toggleDarkMode() {
  const isDark = document.documentElement.classList.toggle('dark');
  localStorage.setItem('theme', isDark ? 'dark' : 'light');
}

// Restore preference on load
(function () {
  const saved = localStorage.getItem('theme');
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
  if (saved === 'dark' || (!saved && prefersDark)) {
    document.documentElement.classList.add('dark');
  }
})();
```

Add a toggle button to the AMIS app header:

```json
{
  "type": "app",
  "headerNav": [
    {
      "type": "button",
      "icon": "fa fa-moon",
      "tooltip": "Toggle dark mode",
      "onClick": "toggleDarkMode()"
    }
  ]
}
```

## Rules

- Override CSS variables only — never target individual `.cxd-*` class backgrounds
- All overrides go in `web/styles/theme.css` under `html.dark`
- Test all screens in both light and dark after adding new components
- AMIS charts (ECharts) need separate dark theme config — see ECharts `theme` option

## ECharts Dark Theme

AMIS chart components pass through to ECharts. Set chart background via schema:

```json
{
  "type": "chart",
  "config": {
    "backgroundColor": "transparent",
    "textStyle": { "color": "#e8e8e8" },
    "series": [...]
  }
}
```

Or register a global ECharts dark theme in `index.html`:

```javascript
echarts.registerTheme('awo-dark', {
  backgroundColor: 'transparent',
  textStyle: { color: '#e8e8e8' },
  categoryAxis: { axisLine: { lineStyle: { color: '#3a3a3a' } } }
});

// In AMIS chart schema:
// "chartTheme": "awo-dark"
```
