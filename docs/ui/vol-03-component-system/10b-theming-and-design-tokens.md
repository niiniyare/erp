---
title: "Theming and Design Tokens"
volume: "III — Component System"
chapter: "10-B"
phase: 2
status: draft
audience: [Web Engineers, Mobile Engineers, UI Framework Developers, ERP Implementers]
---

# Chapter 10-B — Theming and Design Tokens

> **Volume:** III — Component System
> **Audience:** Web Engineers, Mobile Engineers, UI Framework Developers, ERP Implementers
> **Prerequisites:** Chapter 09 — Component System, Chapter 10 — Layout System
> **Phase:** 2

---

## 10B.1 Theming Philosophy

AwoERP operates as a multi-tenant platform where each tenant is a distinct business — a Nairobi fuel retailer, a Mombasa logistics company, a Kisumu SACCO — each with its own brand identity, colour palette, typography, and portal variants. The theming system exists to honour this identity diversity without compromising the integrity of the underlying component system or producing unmaintainable per-tenant CSS hacks.

The design philosophy follows a strict **token-first** approach: all visual properties that can vary are expressed as named tokens, never as hard-coded values. A component never says `color: #1a73e8`; it says `color: var(--colors-brand-primary)`. Tenants override the token, not the component. This separation means that upgrading a component library version will never accidentally overwrite a tenant's brand colour.

The theming system has three layers that apply in sequence. The **platform default theme** establishes a coherent baseline for all components. The **tenant branding layer** overrides a curated set of brand tokens (primary colour, font family, logo). The **portal-level overrides** adjust specific tokens for the supplier, customer, or employee portal shells — for example, the customer portal might use a lighter, consumer-friendly palette while the staff ERP uses a denser, information-rich layout.

All theme tokens are ultimately expressed as **CSS custom properties** on the `html` element, consumed by the amis `cxd` theme's component CSS. The Flutter mobile renderer maintains a parallel Dart token system (`AwoTheme` class) that maps to the same logical tokens but applied to Flutter's `ThemeData`.

> **⚠ Warning:** The amis `cxd` theme has **no built-in dark mode support**. Calling `amis.embed({ theme: 'dark' })` uses `classPrefix: "dark-"` but zero CSS rules exist under that prefix — all components lose their styling. The correct dark mode approach is to override `cxd`'s CSS custom properties at the `:root` level. See Section 10B.7 for the complete dark mode implementation pattern.

---

## 10B.2 Design Token System

### 10B.2.1 Token Categories

AwoERP uses six token categories. Every token belongs to exactly one category.

**Color tokens** are the most extensive category. They span brand colours, semantic status colours (success, warning, error, info), and neutral scale tokens for text, backgrounds, and borders. The amis `cxd` theme internally uses a neutral scale numbered 1–11 where 1 is the darkest and 11 is the lightest. This scale must be respected when constructing dark mode overrides.

**Typography tokens** cover font family, font size scale, font weight, line height, and letter spacing. AwoERP defaults to the system font stack on web (`-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif`) with tenants optionally specifying a Google Font via the tenant branding API. Font sizes follow a type scale from `--font-size-xs` (11px) to `--font-size-4xl` (36px).

**Spacing tokens** are derived from an 4px base grid. `--spacing-1` = 4px, `--spacing-2` = 8px, up to `--spacing-16` = 64px. Components use spacing tokens for padding, margin, and gap values. Custom spacing values are prohibited.

**Elevation tokens** define box-shadow styles for the three elevation levels: `--elevation-low` (subtle card shadow), `--elevation-mid` (modal/panel shadow), `--elevation-high` (dropdown/popover shadow).

**Radius tokens** define border-radius values: `--radius-sm` (2px), `--radius-md` (6px), `--radius-lg` (12px), `--radius-full` (9999px for pill shapes).

**Motion tokens** define transition durations and easing curves: `--motion-fast` (100ms ease-out), `--motion-mid` (200ms ease-in-out), `--motion-slow` (350ms ease-in-out). Tenants may set `--motion-fast: 0ms` to disable animations entirely (accessibility requirement for vestibular motion disorders).

### 10B.2.2 Semantic Tokens vs. Primitive Tokens

The token hierarchy has two levels:

**Primitive tokens** are raw values. They do not carry semantic meaning:
```css
--color-blue-500: #1565C0;
--color-green-600: #2E7D32;
--color-red-500: #C62828;
--color-neutral-100: #F5F5F5;
```

**Semantic tokens** reference primitive tokens and carry intent:
```css
--colors-brand-primary:        var(--color-blue-500);
--colors-status-success:       var(--color-green-600);
--colors-status-error:         var(--color-red-500);
--colors-neutral-fill-11:      var(--color-neutral-100);  /* lightest background */
--colors-neutral-text-2:       #212121;                    /* primary text */
```

Tenants and components **only ever reference semantic tokens**. Primitive tokens exist solely to support the semantic layer. This means a tenant that wants "a slightly warmer primary blue" overrides `--colors-brand-primary` to a specific hex value — they do not need to create a new primitive token.

### 10B.2.3 Token Inheritance Hierarchy

```
Primitive tokens (platform hardcoded, never overridden)
  └── Platform semantic tokens (platform defaults, base theming)
        └── Tenant brand tokens (overridden per tenant, loaded from DB + Redis)
              └── Portal semantic tokens (overridden per portal type)
                    └── Component-level tokens (rare, scope-limited overrides)
```

At runtime, this hierarchy is materialised as a series of CSS custom property declarations on `html`, `html[data-portal="supplier"]`, `html[data-portal="customer"]`, etc.

### 10B.2.4 Token Reference in Component Props

Surface definitions can reference tokens using the `token()` syntax in prop values:

```json
{
  "type": "container",
  "style": {
    "background": "token(--colors-neutral-fill-11)",
    "border": "1px solid token(--colors-neutral-line-8)",
    "padding": "token(--spacing-4)"
  }
}
```

During compilation, `token(...)` references are validated against the registered token schema to catch typos at build time rather than displaying a broken UI at runtime.

---

## 10B.3 Platform Default Theme

### 10B.3.1 Light Mode Token Set

The platform light mode theme defines the baseline for all AwoERP deployments. Key tokens:

```css
/* awo-overrides.css — platform light mode defaults */
:root {
  /* Brand */
  --colors-brand-primary:          #1565C0;
  --colors-brand-primary-light:    #1976D2;
  --colors-brand-primary-dark:     #0D47A1;
  --colors-brand-accent:           #00897B;

  /* Neutral scale (amis cxd: 1=darkest, 11=lightest) */
  --colors-neutral-fill-1:         #212121;
  --colors-neutral-fill-2:         #424242;
  --colors-neutral-fill-3:         #616161;
  --colors-neutral-fill-8:         #E0E0E0;
  --colors-neutral-fill-9:         #EEEEEE;
  --colors-neutral-fill-10:        #F5F5F5;
  --colors-neutral-fill-11:        #FFFFFF;

  /* Text */
  --colors-neutral-text-2:         #212121;  /* primary text */
  --colors-neutral-text-3:         #616161;  /* secondary text */

  /* Lines/borders */
  --colors-neutral-line-8:         #E0E0E0;

  /* Status */
  --colors-status-success:         #2E7D32;
  --colors-status-warning:         #E65100;
  --colors-status-error:           #C62828;
  --colors-status-info:            #1565C0;

  /* Component surfaces */
  --background:                    #FFFFFF;
  --body-bg:                       #F5F5F5;
  --Page-main-bg:                  #F5F5F5;
  --Panel-bg-color:                #FFFFFF;
  --Table-bg:                      #FFFFFF;
  --Table-thead-bg:                #F5F5F5;
  --Table-tr-hover-bg:             #E3F2FD;

  /* Typography */
  --font-family-base:              -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  --font-size-base:                14px;
  --font-size-sm:                  12px;
  --font-size-lg:                  16px;
  --font-weight-normal:            400;
  --font-weight-medium:            500;
  --font-weight-bold:              600;

  /* Spacing (4px grid) */
  --spacing-1: 4px;   --spacing-2: 8px;   --spacing-3: 12px;
  --spacing-4: 16px;  --spacing-6: 24px;  --spacing-8: 32px;

  /* Radius */
  --radius-sm: 2px;  --radius-md: 6px;  --radius-lg: 12px;

  /* Elevation */
  --elevation-low:  0 1px 3px rgba(0,0,0,0.12), 0 1px 2px rgba(0,0,0,0.08);
  --elevation-mid:  0 4px 6px rgba(0,0,0,0.10), 0 2px 4px rgba(0,0,0,0.08);
  --elevation-high: 0 10px 15px rgba(0,0,0,0.10), 0 4px 6px rgba(0,0,0,0.06);

  /* Motion */
  --motion-fast: 100ms ease-out;
  --motion-mid:  200ms ease-in-out;
  --motion-slow: 350ms ease-in-out;
}
```

### 10B.3.2 Dark Mode Token Set

Dark mode is implemented by inverting the neutral scale and adjusting surface backgrounds. The neutral scale flip (1=lightest, 11=darkest in dark mode) is the core of the dark mode implementation:

```css
html.dark {
  /* Invert the neutral scale */
  --colors-neutral-fill-1:         #FAFAFA;
  --colors-neutral-fill-2:         #F5F5F5;
  --colors-neutral-fill-8:         #424242;
  --colors-neutral-fill-9:         #303030;
  --colors-neutral-fill-10:        #212121;
  --colors-neutral-fill-11:        #121212;

  /* Text */
  --colors-neutral-text-2:         #FAFAFA;
  --colors-neutral-text-3:         #BDBDBD;

  /* Lines/borders */
  --colors-neutral-line-8:         #424242;

  /* Component surfaces */
  --background:                    #1E1E1E;
  --body-bg:                       #121212;
  --Page-main-bg:                  #121212;
  --Panel-bg-color:                #1E1E1E;
  --Table-bg:                      #1E1E1E;
  --Table-thead-bg:                #2A2A2A;
  --Table-tr-hover-bg:             #1A2332;

  /* Status (lighter variants for dark backgrounds) */
  --colors-status-success:         #81C784;
  --colors-status-warning:         #FFB74D;
  --colors-status-error:           #EF9A9A;
  --colors-status-info:            #90CAF9;
}
```

> **✅ Convention:** Never override individual `.cxd-*` class backgrounds with `!important`. This is a whack-a-mole approach that breaks with every amis SDK update. Override the CSS custom properties — all `cxd` components consume them and will update automatically.

### 10B.3.3 High-Contrast Mode Token Set

High-contrast mode is provided for accessibility compliance (WCAG 2.1 AA requires 4.5:1 contrast for normal text, AAA requires 7:1). AwoERP provides a high-contrast light and high-contrast dark variant:

```css
html.high-contrast {
  --colors-neutral-text-2:         #000000;
  --colors-neutral-text-3:         #000000;
  --colors-neutral-line-8:         #000000;
  --colors-brand-primary:          #003087;
  --colors-status-error:           #8B0000;
  --colors-status-success:         #005A00;
}

html.high-contrast.dark {
  --colors-neutral-text-2:         #FFFFFF;
  --colors-neutral-text-3:         #FFFFFF;
  --colors-neutral-line-8:         #FFFFFF;
  --colors-brand-primary:          #80BFFF;
}
```

---

## 10B.4 Tenant Branding Layer

### 10B.4.1 Tenant Logo and Favicon

Every tenant can provide a logo and favicon via the tenant settings API. Logos are stored in object storage (S3-compatible) and referenced by URL. The logo URL is injected into `window.__AWO_BOOTSTRAP__.themeTokens.logoUrl` at page load time by the Go template renderer:

```go
// internal/api/handlers/web/index.go
type BootstrapData struct {
    TenantID     string            `json:"tenantId"`
    ThemeTokens  map[string]string `json:"themeTokens"`
    // ...
}

func buildBootstrapData(ctx context.Context, tenant *tenantDomain.Tenant) BootstrapData {
    tokens := map[string]string{
        "--colors-brand-primary": tenant.BrandColorPrimary,
        "--font-family-base":     tenant.FontFamily,
        "logoUrl":                tenant.LogoURL,
        "faviconUrl":             tenant.FaviconURL,
    }
    return BootstrapData{TenantID: tenant.ID.String(), ThemeTokens: tokens}
}
```

The favicon URL is set dynamically in `index.html` via JavaScript before `amis.embed()` is called:

```javascript
// In the init script block
const bootstrap = window.__AWO_BOOTSTRAP__;
if (bootstrap.themeTokens?.faviconUrl) {
    document.querySelector("link[rel='icon']").href = bootstrap.themeTokens.faviconUrl;
}
```

### 10B.4.2 Tenant Primary / Accent Color Overrides

Tenants can override the following brand tokens:

| Token | Description | Example (Shell Maanzoni) |
|-------|-------------|--------------------------|
| `--colors-brand-primary` | Main brand colour | `#DD1700` (Shell red) |
| `--colors-brand-primary-light` | Hover state | `#E53935` |
| `--colors-brand-primary-dark` | Active state | `#B71C1C` |
| `--colors-brand-accent` | Secondary accent | `#FFCC00` (Shell yellow) |

These four tokens are the only brand colour tokens tenants can override. Tenants cannot override status colours, neutral scales, or typography scale values — these are platform-controlled for consistency and accessibility.

### 10B.4.3 Tenant Typography Overrides

Tenants may override `--font-family-base` by providing a Google Fonts CSS import URL. The platform validates that the URL is a legitimate `fonts.googleapis.com` URL before storing it:

```go
func validateFontURL(url string) error {
    if !strings.HasPrefix(url, "https://fonts.googleapis.com/css") {
        return fmt.Errorf("font URL must be a Google Fonts CSS URL")
    }
    return nil
}
```

The font link element is injected into `index.html` by the Go SSR handler before serving the page, ensuring the font is loaded before amis renders.

### 10B.4.4 Tenant Theme Storage (PostgreSQL + Redis)

Tenant theme data is stored in the `tenant_theme` table (inside the tenant's RLS-scoped schema):

```sql
CREATE TABLE tenant_themes (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL REFERENCES tenants(id) UNIQUE,
    tokens       JSONB NOT NULL DEFAULT '{}',
    logo_url     TEXT,
    favicon_url  TEXT,
    font_url     TEXT,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Theme data is cached in Redis with the key pattern `theme:{tenant_id}` and a TTL of 1 hour. On tenant theme update, the cache key is invalidated immediately:

```go
func (s *TenantThemeService) UpdateTheme(ctx context.Context, tenantID uuid.UUID, update ThemeUpdate) error {
    // ... validate and persist to PostgreSQL
    err = store.WithTenant(ctx, tenantID, func(q *db.Queries) error {
        return q.UpsertTenantTheme(ctx, db.UpsertTenantThemeParams{
            TenantID: tenantID,
            Tokens:   update.TokensJSON(),
            LogoURL:  update.LogoURL,
        })
    })
    if err != nil {
        return err
    }
    // Invalidate cache
    return s.cache.Delete(ctx, fmt.Sprintf("theme:%s", tenantID))
}
```

Theme data is loaded at request time by the `TenantMiddleware`. On a cache miss, it reads from PostgreSQL and re-populates the cache. The theme is then available in the request context for injection into `__AWO_BOOTSTRAP__`.

### 10B.4.5 Tenant Theme API Endpoints

```
GET    /api/v1/tenants/{id}/theme          — Retrieve current theme (requires tenant.settings.read)
PUT    /api/v1/tenants/{id}/theme          — Replace full theme (requires tenant.settings.write)
PATCH  /api/v1/tenants/{id}/theme          — Update specific tokens (requires tenant.settings.write)
POST   /api/v1/tenants/{id}/theme/logo     — Upload logo (multipart, requires tenant.settings.write)
POST   /api/v1/tenants/{id}/theme/favicon  — Upload favicon (multipart, requires tenant.settings.write)
DELETE /api/v1/tenants/{id}/theme          — Reset to platform defaults (requires tenant.settings.admin)
```

---

## 10B.5 Portal-Level Theme Overrides

Portals (Supplier, Customer, Employee) each have a distinct visual identity layered on top of the tenant theme. Portal overrides are applied when the amis `app` shell detects the portal type from `window.__AWO_BOOTSTRAP__.portalType`.

### 10B.5.1 Supplier Portal Theme

The supplier portal uses a professional, neutral palette. The primary brand colour is retained from the tenant theme (suppliers interact with a tenant-branded portal), but the shell is simplified:

```css
html[data-portal="supplier"] {
  --Page-main-bg:   #F8F9FA;
  --Panel-bg-color: #FFFFFF;
  --spacing-page-h: var(--spacing-6);  /* wider page padding for readability */
  --font-size-base: 15px;              /* slightly larger for external users */
}
```

### 10B.5.2 Customer Portal Theme

The customer portal has a consumer-facing feel with more colour and a warmer tone:

```css
html[data-portal="customer"] {
  --Page-main-bg:               #F0F4F8;
  --Panel-bg-color:             #FFFFFF;
  --colors-brand-primary:       var(--tenant-brand-primary, #1565C0);
  --radius-md:                  10px;  /* rounder corners for consumer feel */
  --font-size-base:             15px;
}
```

The customer portal also integrates a dedicated payment accent colour (`--colors-payment-cta`) used for M-Pesa payment action buttons, set to the Safaricom green `#4CAF50` by default.

### 10B.5.3 Employee Self-Service Portal Theme

The employee portal is optimised for mobile-first usage (many Kenyan employees access it from mobile devices on variable connectivity):

```css
html[data-portal="employee"] {
  --font-size-base:  15px;
  --spacing-4:       18px;   /* larger tap targets */
  --radius-md:       8px;
  --Table-bg:        #FFFFFF;
}
```

---

## 10B.6 amis `cxd` Theme Mapping

### 10B.6.1 CSS Variable Override Points

The amis `cxd` theme exposes its visual properties through CSS custom properties. AwoERP targets a curated subset of these override points. The full list of properties overridden in `awo-overrides.css` is:

```css
/* Key amis cxd override points in awo-overrides.css */
:root {
  /* Colors */
  --colors-brand-5:           var(--colors-brand-primary);
  --colors-brand-6:           var(--colors-brand-primary-dark);
  --colors-neutral-fill-1:    /* ... */;
  /* ... through fill-11 */
  --colors-neutral-text-2:    /* ... */;
  --colors-neutral-text-3:    /* ... */;
  --colors-neutral-line-8:    /* ... */;

  /* Component-level */
  --background:               /* page background */;
  --body-bg:                  /* body background */;
  --Page-main-bg:             /* main content area */;
  --Panel-bg-color:           /* card/panel background */;
  --Table-bg:                 /* table background */;
  --Table-thead-bg:           /* table header background */;
  --Button-primary-bg:        var(--colors-brand-primary);
  --Button-primary-border:    var(--colors-brand-primary-dark);
  --Button-primary-color:     #FFFFFF;
}
```

### 10B.6.2 Permitted Customisation Surface

The **two-variable rule**: tenant theme overrides may only directly set CSS custom properties that appear in the official AwoERP token registry. Tenants may not inject arbitrary CSS. This is enforced at the theme API validation layer:

```go
var allowedTokens = map[string]bool{
    "--colors-brand-primary":       true,
    "--colors-brand-primary-light": true,
    "--colors-brand-primary-dark":  true,
    "--colors-brand-accent":        true,
    "--font-family-base":           true,
    // ... 12 more permitted tokens
}

func validateTokens(tokens map[string]string) error {
    for key := range tokens {
        if !allowedTokens[key] {
            return fmt.Errorf("token %q is not in the permitted override set", key)
        }
    }
    return nil
}
```

### 10B.6.3 Constraints and Anti-Patterns

> **⚠ Warning:** Do not add `!important` to any `.cxd-*` class overrides. The amis SDK updates its class names between versions; `!important` overrides on class names will silently break on SDK upgrades, producing invisible styling issues that are difficult to debug.

> **⚠ Warning:** Do not override amis component colours by targeting element structure (e.g. `.cxd-Panel > .cxd-Panel-body`). These DOM structures change between amis versions. Use CSS custom properties exclusively.

> **✅ Convention:** Every CSS rule in `awo-overrides.css` must have a comment linking it to a token registry entry. Undocumented overrides fail CI:
> ```css
> /* [TOKEN: --Panel-bg-color] Panel card background */
> --Panel-bg-color: var(--colors-neutral-fill-11);
> ```

---

## 10B.7 Dark Mode Runtime Switching

Dark mode can be activated in three ways:
1. **System preference**: `prefers-color-scheme: dark` is detected at page load
2. **User preference**: Stored in `localStorage` under `awo-color-mode` (only colour-mode preference is stored here — never sensitive data)
3. **Tenant policy**: Tenant admin can force light or dark mode for all users

The dark mode initialisation script runs before `amis.embed()` to prevent a flash of light mode:

```javascript
// Dark mode init — runs synchronously before amis loads
(function() {
    var stored = localStorage.getItem('awo-color-mode');
    var tenantForce = window.__AWO_BOOTSTRAP__.colorModePolicy; // 'light' | 'dark' | 'auto'
    var prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;

    var mode;
    if (tenantForce && tenantForce !== 'auto') {
        mode = tenantForce;
    } else if (stored) {
        mode = stored;
    } else if (prefersDark) {
        mode = 'dark';
    } else {
        mode = 'light';
    }

    if (mode === 'dark') {
        document.documentElement.classList.add('dark');
    }
    document.documentElement.setAttribute('data-color-mode', mode);
})();
```

The dark mode toggle in the user profile dropdown dispatches a custom DOM event that the init script handler responds to:

```javascript
function toggleDarkMode() {
    var current = document.documentElement.getAttribute('data-color-mode');
    var next = current === 'dark' ? 'light' : 'dark';
    document.documentElement.classList.toggle('dark', next === 'dark');
    document.documentElement.setAttribute('data-color-mode', next);
    localStorage.setItem('awo-color-mode', next);
    // Trigger amis re-render of theme-sensitive components
    window.dispatchEvent(new CustomEvent('awo:theme-change', { detail: { mode: next } }));
}
```

---

## 10B.8 Component-Level Style Override Mechanism

In rare cases, a specific surface needs to override a token value only within that surface's scope (not globally). Component-level overrides are expressed via the `themeOverride` prop on container nodes:

```json
{
  "type": "container",
  "themeOverride": {
    "--Panel-bg-color": "#FFF8E1",
    "--colors-status-warning": "#F57F17"
  },
  "children": [
    { "type": "alert", "level": "warning", "body": "This invoice has outstanding eTIMS submission." }
  ]
}
```

During compilation, the `themeOverride` map is converted to an inline `style` attribute on the container's DOM element, scoping the token values to its subtree. This is the only approved mechanism for in-surface token overrides.

> **✅ Convention:** Use component-level overrides sparingly. If you find yourself overriding the same tokens in 3+ surfaces, the token value should be changed in the tenant branding layer instead.

---

## 10B.9 Theme Export and Import (Tenant Migration)

When migrating a tenant from one AwoERP installation to another, the theme configuration is exported as a portable JSON bundle:

```json
{
  "schema_version": "1.0",
  "exported_at": "2026-06-05T10:00:00Z",
  "tenant_id": "3f2e1d4c-...",
  "tokens": {
    "--colors-brand-primary": "#DD1700",
    "--colors-brand-accent": "#FFCC00",
    "--font-family-base": "'Shell Font', sans-serif"
  },
  "logo_url": "https://cdn.awoerp.com/tenants/shell-maanzoni/logo.png",
  "favicon_url": "https://cdn.awoerp.com/tenants/shell-maanzoni/favicon.ico",
  "portal_overrides": {
    "customer": {
      "--colors-payment-cta": "#4CAF50"
    }
  }
}
```

This bundle is imported via `POST /api/v1/tenants/{id}/theme/import` and validated against the token registry before being applied. Asset URLs (logo, favicon) are re-uploaded to the target installation's object storage during import.

> See Chapter 50-B — Onboarding and Migration Guide for the full tenant migration workflow.

---

## 10B.10 Design Token CI Validation

Token integrity is enforced in CI by the `awo ui validate-tokens` command:

```yaml
# .github/workflows/ui-tokens.yml
name: Design Token Validation
on: [push, pull_request]
jobs:
  validate-tokens:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Validate token registry completeness
        run: awo ui validate-tokens --registry internal/ui/tokens/registry.go --overrides web/assets/awo-overrides.css
      - name: Check no undocumented overrides in awo-overrides.css
        run: awo ui validate-tokens --check-comments web/assets/awo-overrides.css
      - name: Verify all amis cxd variables are mapped
        run: awo ui validate-tokens --check-coverage --amis-version $(cat .amis-version)
      - name: Check dark mode has counterpart for all light tokens
        run: awo ui validate-tokens --check-dark-parity web/assets/awo-overrides.css
```

The validation rules enforced:
1. Every token referenced in `awo-overrides.css` must exist in the Go token registry
2. Every token in the registry must appear in `awo-overrides.css` (no orphan tokens)
3. Every token overridden in light mode must have a corresponding dark mode override
4. No `.cxd-*` class-targeted rules exist in `awo-overrides.css` (CSS property overrides only)
5. All override comments follow the `[TOKEN: ...]` format

> See Chapter 42 — CI/CD Considerations for how token validation integrates into the full CI pipeline.
