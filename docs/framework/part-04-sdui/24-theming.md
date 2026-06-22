---
title: "Chapter 24: Theming and Branding Per Tenant"
part: "Part IV — The SDUI Layer"
chapter: 24
section: "24-theming"
related:
  - "[Chapter 21: SDUI Philosophy](21-sdui-philosophy.md)"
  - "[Chapter 14: Multi-Tenancy Middleware](../part-03-api/14-multitenancy-middleware.md)"
---

# Chapter 24: Theming and Branding Per Tenant

Awo supports per-tenant visual branding: primary colour, logo, and dark mode preference. Theming is delivered via a CSS endpoint and CSS custom properties — no per-tenant JavaScript bundles, no build steps.

---

## 24.1. CSS Variable Injection

### 24.1.1. The `/api/v1/theme.css` Endpoint

Each tenant has a unique theme endpoint:

```
GET https://acme.awo.app/api/v1/theme.css
```

The response is a CSS file scoped to `:root`:

```css
:root {
  --primary-color: #1a73e8;
  --primary-color-hover: #1557b0;
  --border-radius: 6px;
  --font-family: 'Inter', sans-serif;
  --body-bg: #f8f9fa;
}
```

amis loads this stylesheet at startup and applies it to all components via CSS custom properties. This is the only mechanism for theming — never inline styles.

### 24.1.2. Overridable Variables

| Variable | Default | Effect |
|---|---|---|
| `--primary-color` | `#1a73e8` | Buttons, links, active states |
| `--primary-color-hover` | auto-darkened | Hover state |
| `--border-radius` | `4px` | Card/input corner radius |
| `--font-family` | `Inter, sans-serif` | All text |
| `--body-bg` | `#f5f7fa` | Page background |
| `--sidebar-bg` | `#001529` | Navigation sidebar |
| `--sidebar-text` | `#fff` | Sidebar text |

### 24.1.3. How amis Loads the Theme

```html
<!-- In web/pages/index.html -->
<link id="tenant-theme" rel="stylesheet" href="/api/v1/theme.css">
```

The browser loads this CSS before rendering any amis components. Changes to tenant theme require a page reload to take effect (CSS is cached with `Cache-Control: max-age=300`).

---

## 24.2. Logo and Colour Configuration

### 24.2.1. Logo Upload and Storage

Tenants upload their logo via `POST /api/v1/tenant/logo`. The server:
1. Validates: image type, max 1MB, min dimensions 100×40px
2. Generates a thumbnail version (200×80px max)
3. Stores in object storage at `logos/{tenant_id}/logo.{ext}`
4. Updates `tenant.logo_path`

The logo is served via `GET /api/v1/tenant/logo` with long cache headers (365 days + ETag for invalidation).

### 24.2.2. Colour Picker and Contrast Ratio Enforcement

The admin UI provides a colour picker for `primary_color`. On save, the server validates:

```go
func validateBrandColour(hex string) error {
    colour, err := parseHexColour(hex)
    if err != nil {
        return validate.FieldErrorf("primary_color", "must be a valid hex colour")
    }

    // Check contrast ratio against white background (WCAG AA requires 4.5:1 for normal text)
    ratio := contrastRatio(colour, white)
    if ratio < 3.0 {  // minimum for large UI elements
        return validate.FieldErrorf("primary_color",
            "colour contrast ratio %.1f:1 is too low for accessibility (minimum 3:1)", ratio)
    }
    return nil
}
```

Colours that fail the contrast check are rejected with a user-facing error and a suggestion of a darker/lighter alternative that would pass.

### 24.2.3. Dark Mode Support

Awo supports dark mode per-tenant. Three modes:
- `system` (default): follows the user's OS preference (`prefers-color-scheme`)
- `always_light`: always use light theme
- `always_dark`: always use dark theme

Dark mode is implemented by overriding CSS custom properties under `html.dark`:

```css
html.dark {
  --body-bg: #1a1a2e;
  --colors-neutral-fill-11: #0d0d1a;
  --colors-neutral-text-2: #e8e8f0;
  /* ... other overrides */
}
```

See the AMIS SDK Dark Theme memory entry for implementation details on why individual component overrides with `!important` are avoided.

---

## 24.3. Print Stylesheets

### 24.3.1. Print Layout for Invoices

The invoice detail page includes a print stylesheet that formats the page for A4 paper:

```css
@media print {
  .sidebar, .toolbar, .tabs-nav, .action-buttons { display: none; }
  .invoice-header {
    display: flex;
    justify-content: space-between;
    border-bottom: 2px solid #000;
    padding-bottom: 16px;
    margin-bottom: 16px;
  }
  .line-items-table { width: 100%; border-collapse: collapse; }
  .line-items-table th, .line-items-table td {
    border: 1px solid #ccc;
    padding: 6px 8px;
    font-size: 11pt;
  }
  .company-logo { max-width: 160px; max-height: 60px; }
}
```

Print stylesheets are served from the theme endpoint and can be customised per-tenant.

### 24.3.2. PDF Generation

PDFs are generated server-side from the print layout using a headless browser (Chromium via `rod`):

```go
func GenerateInvoicePDF(ctx context.Context, invoiceID uuid.UUID) ([]byte, error) {
    // Get the authenticated print URL
    url := fmt.Sprintf("https://%s.awo.app/invoices/%s/print",
        tenant.SlugFromContext(ctx), invoiceID)

    browser := rod.New().MustConnect()
    defer browser.MustClose()

    page := browser.MustPage(url)
    page.MustWaitLoad()

    return page.PDF(&proto.PagePrintToPDF{
        Format:            "A4",
        PrintBackground:   true,
        MarginTop:         0.5,
        MarginBottom:      0.5,
        MarginLeft:        0.5,
        MarginRight:       0.5,
    })
}
```

PDF generation is executed as a Temporal activity (async) for large documents. Small PDFs (<5 pages) can be generated synchronously and returned inline.
