[<-- Back to Index](README.md)

## Theming & Tenant Branding

### The Three-Tier Token System

All visual values are CSS custom properties in a three-tier hierarchy. No hardcoded colour values in component styles.

```markdown
TIER 1 — PRIMITIVE TOKENS (raw values, never used directly in components)
  --color-blue-500: #2196f3
  --color-grey-900: #1a1a2e
  --spacing-4:      16px
  --radius-md:      6px

TIER 2 — SEMANTIC TOKENS (named by purpose, reference primitives)
  --color-primary:         → --color-blue-500
  --color-background:      #f5f6f8  (light)  / #151521  (dark)
  --color-surface:         #ffffff  (light)  / #1e1e2d  (dark)
  --color-text-primary:    #1a1a2e  (light)  / #e4e4e7  (dark)
  --color-text-secondary:  #6b7280  (light)  / #9a9ab0  (dark)
  --color-border:          #e8eaed  (light)  / #2e2e3e  (dark)
  --color-danger:          #e74c3c  (both modes)
  --color-warning:         #f59e0b  (both modes)
  --color-success:         #16a34a  (both modes)

TIER 3 — COMPONENT TOKENS (named by component, reference semantic)
  --sidebar-bg:            → --color-surface
  --sidebar-active:        #e8f4fd  (light)  / #1a2744  (dark)
  --sidebar-active-border: → --color-primary
  --Table-bg:              → --color-surface
  --Panel-bg-color:        → --color-surface
```

Only Tier 2 and Tier 3 tokens appear in component styles. When a tenant overrides their primary colour, they change `--color-primary`. Every button, link, badge, and focus ring that references it updates automatically.

### Current Tokens in `index.html`

Shell tokens are in `:root` and `html.dark`. AMIS component tokens are in the large `html.dark` block. See [§05 Dark Mode](./05-dark-mode.md) for the full AMIS token list.

### Brand Colour

Currently hardcoded as `#2196f3` in two places:

```css
/* index.html */
.brand-icon { background: #2196f3; }   /* sidebar logo bg */
#topbar .brand-icon { background: #2196f3; }
```

And in CSS variables:
```css
:root { --sidebar-active-border: #2196f3; }
html.dark { --sidebar-active-border: #4dabf7; }
```

And in AMIS variables:
```css
:root { --colors-brand-5: /* not yet explicitly set — AMIS default applies */ }
```

**To implement tenant branding:**

1. Add a CSS variable for the brand colour:
```css
:root {
  --brand-primary:       #2196f3;
  --brand-primary-dark:  #4dabf7;   /* lighter shade for dark bg */
}
html.dark {
  --colors-brand-5: var(--brand-primary-dark);
}
```

2. Fetch tenant config from `/api/v1/me/tenant` on shell load, then:
```javascript
document.documentElement.style.setProperty('--brand-primary', tenantConfig.brand_color);
document.documentElement.style.setProperty('--brand-primary-dark', tenantConfig.brand_color_dark);
```

### Tenant Branding Configuration (Settings → Branding)

| Setting | Description | Constraint |
|---|---|---|
| Primary colour | Buttons, links, active states | Must pass AA contrast vs white |
| Logo (light mode) | SVG or PNG, max 200×60px | Validated dimensions |
| Logo (dark mode) | Separate logo for dark (optional) | Falls back to light logo with CSS inversion |
| Favicon | 32×32 or 64×64 ICO/PNG | Browser tab |
| Application name | Tab title, email subjects | Max 40 characters |
| Custom CSS | Power escape hatch | Sandboxed; platform can disable |

**What tenants cannot configure:**
- Font family (consistency and performance)
- Layout structure (menu position, grid)
- Status colours (danger/warning/success must remain consistent for safety)
- Any colour that fails WCAG AA contrast

### Per-User vs Tenant Colour Scheme

Resolution order:
```markdown
Tenant forced scheme → User preference → System preference (OS dark/light)
```

If a tenant enforces dark mode for all users (e.g. for kiosk deployments), the user preference is overridden. This is a future feature — current implementation uses only user preference + system preference.

---
