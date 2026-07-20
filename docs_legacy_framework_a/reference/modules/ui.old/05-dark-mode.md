> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

[<-- Back to Index](README.md)

## Dark Mode

### How It Works

Dark mode is implemented in two layers:

```markdown
LAYER 1 — Shell CSS variables
  html.dark { --sidebar-bg: #1e1e2d; --text-primary: #e4e4e7; ... }
  These style the sidebar, topbar, footer.

LAYER 2 — AMIS CSS variable overrides
  html.dark { --colors-neutral-fill-11: #1e1e2d; ... }
  These style everything AMIS renders inside #content.
```

Both layers activate when JS adds the `dark` class to `<html>`.

### Why AMIS Has No Built-in Dark Mode

AMIS ships a `theme("dark")` option that uses `classPrefix: "dark-"`. This generates CSS classes like `.dark-cxd-Button`. But there are **zero CSS rules for this prefix** in `sdk.css`. Using it breaks all component styling.

**The correct approach:** Override AMIS's CSS custom property scales at the root.

### AMIS CSS Token Scales

AMIS uses three scales, numbered 1 (darkest) to 11 (lightest):

```markdown
--colors-neutral-fill-N   → background colours
--colors-neutral-text-N   → foreground/text colours
--colors-neutral-line-N   → border colours

In light mode: 11 = white (#fff), 1 = near-black (#070c14)
In dark mode:  INVERT — 11 becomes dark, 1 becomes near-white
```

### The Inversion

```css
/* Light mode (default, no override needed) */
--colors-neutral-fill-11: #ffffff;   /* Surface bg */
--colors-neutral-text-2:  #151b26;   /* Primary text */
--colors-neutral-line-8:  #e8e9eb;   /* Default border */

/* Dark mode — invert the scale */
html.dark {
  --colors-neutral-fill-11: #1e1e2d;  /* was white → now dark surface */
  --colors-neutral-fill-10: #262637;  /* was near-white → now slightly lighter dark */
  --colors-neutral-text-2:  #e4e4e7;  /* was near-black text → now near-white */
  --colors-neutral-line-8:  #363648;  /* was light border → now dark border */
  /* ... full scale in index.html */
}
```

### Component-Specific Overrides

AMIS portals (dropdowns, modals, drawers, tooltips) attach to `<body>` outside `#content`. They need explicit dark overrides:

```css
html.dark {
  /* Component backgrounds */
  --Panel-bg-color:      #1e1e2d;
  --Table-bg:            #1e1e2d;
  --Table-thead-bg:      #262637;
  --Modal-bg:            #1e1e2d;
  --Drawer-bg:           #1e1e2d;
  --Form-input-bg:       #262637;
  --Select-menu-bg:      #1e1e2d;
  --DropDown-menu-bg:    #1e1e2d;
  --DatePicker-bg:       #1e1e2d;

  /* Portal reinforcement (appended to body) */
  --PopOver-bg:          #1e1e2d;
  --Tooltip-bg:          #2e2e3e;
}

/* Explicit portal selectors (CSS vars cascade but we add safety) */
html.dark .cxd-PopOver          { background: var(--PopOver-bg); }
html.dark .cxd-Modal-content    { background: var(--Modal-bg); }
html.dark .cxd-Drawer-content   { background: var(--Drawer-bg); }
html.dark .cxd-Select-menu      { background: var(--Select-menu-bg); }
```

### Do NOT Use These Approaches

```markdown
❌ theme: 'dark' in embed() — no CSS behind it, breaks components
❌ Overriding .cxd-* class backgrounds with !important everywhere
   — endless whack-a-mole, each AMIS update breaks it again
❌ Separate dark stylesheet — hard to maintain alongside sdk.css updates
```

### Testing Dark Mode

Every new page/schema must be tested in dark mode explicitly. Issues to look for:

```markdown
COMMON DARK MODE BUGS:
├── White background showing through on panel load (use --Panel-bg-color)
├── Dropdown/select menus appearing light (portal issue — add .cxd-Select-menu rule)
├── Date picker cells not inverting (add --DatePicker-cell-bg)
├── Chart backgrounds staying white (ECharts needs explicit backgroundColor: 'transparent')
└── Custom HTML in tpl components with hardcoded colors
```

### Tenant Brand Colour in Dark Mode

The brand colour (`#2196f3` by default) must have sufficient contrast in both modes. When tenant branding is implemented:

```css
:root {
  --colors-brand-5: #2196f3;    /* light mode primary */
  --colors-brand-6: #1976d2;    /* light mode hover */
}

html.dark {
  --colors-brand-5: #4dabf7;    /* lighter shade for dark bg contrast */
  --colors-brand-6: #339af0;
}
```

---
