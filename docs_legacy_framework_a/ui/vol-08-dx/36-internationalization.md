> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "36 – Internationalization"
volume: "vol-08-dx"
chapter: 36
section: "Developer Experience"
status: "partial — locale data wired; translation pipeline planned"
---

# Chapter 36 – Internationalization

## Table of Contents
- [36.1 Locale Data in UISessionContext](#361-locale-data-in-uisessioncontext)
- [36.2 Reading User Preferences](#362-reading-user-preferences)
- [36.3 Tenant Currency in Schemas](#363-tenant-currency-in-schemas)
- [36.4 AMIS SDK Locale Config](#364-amis-sdk-locale-config)
- [36.5 Planned: Translation Pipeline](#365-planned-translation-pipeline)

---

## 36.1 Locale Data in UISessionContext

`UISessionContext` carries three locale-relevant fields resolved before the pipeline
runs. Block functions and screen functions receive these values pre-filled — no
locale resolution logic belongs inside DSL code.

```go
type UISessionContext struct {
    // ... auth, permissions, flags ...

    Locale   string // BCP-47 tag e.g. "en", "ar", "fr-MA"
    Timezone string // IANA zone e.g. "Africa/Nairobi", "Asia/Riyadh"
    Currency string // ISO 4217 e.g. "KES", "SAR", "USD"
}
```

These three fields are the sole source of truth for rendering decisions within
the DSL.  A block that formats a monetary amount reads `sess.Currency`; it does
not query a settings service.

---

## 36.2 Reading User Preferences

`sess.Pref` is the lightweight accessor for arbitrary user or tenant preference
keys.  The second argument is the fallback value used when the key is absent.

```go
func BalanceSummaryBlock(sess ui.UISessionContext) ast.Node {
    locale := sess.Pref("ui.locale", "en")     // user-level override
    tz     := sess.Pref("ui.timezone", sess.Timezone) // falls back to tenant TZ

    return ast.PanelNode{
        Title: "Balance Summary",
        Body: []ast.Node{
            ast.StatNode{
                Label: "As Of",
                Value: fmt.Sprintf("${balance_date | date:%s}", tz),
            },
        },
    }
}
```

**Rules:**
- `sess.Pref` is read-only inside DSL functions — preferences are resolved before
  the pipeline starts, not written during schema generation.
- Always supply a non-empty fallback; a missing preference key must never cause a
  compile error.

---

## 36.3 Tenant Currency in Schemas

`sess.Currency` drives the currency symbol and formatting in rendered schemas.
Pass it into AMIS `InputNumber` or stat labels so every money field reflects the
tenant's base currency without hard-coding.

```go
func InvoiceTotalsBlock(sess ui.UISessionContext) ast.Node {
    currencyLabel := fmt.Sprintf("Amount (%s)", sess.Currency) // e.g. "Amount (KES)"

    return ast.PanelNode{
        Title: "Totals",
        Body: []ast.Node{
            ast.PropertyNode{
                Items: []ast.PropertyItem{
                    {Label: currencyLabel, Content: "${total_amount}"},
                    {Label: "Tax",         Content: "${tax_amount}"},
                    {Label: "Grand Total", Content: "${grand_total}"},
                },
            },
        },
    }
}
```

For data grids, use a column `label` that includes the currency suffix rather
than hard-coding "USD" or similar:

```go
ast.TableColumnNode{
    Name:  "amount",
    Label: fmt.Sprintf("Amount (%s)", sess.Currency),
    Type:  "number",
}
```

---

## 36.4 AMIS SDK Locale Config

The AMIS web SDK ships with locale packs for number formatting, date formatting,
and calendar widgets.  The active locale is set once when the SDK is initialised
in `web/pages/index.html`:

```html
<script>
  amis.embed('#root', schema, {}, {
    locale: '{{ .Locale }}',   // injected by Go template
    // or read from window.__ERP_SESSION__.locale
  });
</script>
```

The server-rendered template injects `Locale` from the session so the AMIS SDK
automatically:
- formats numbers with the correct decimal / grouping separators
- displays dates in the locale-preferred order (DMY vs MDY)
- uses locale-appropriate calendar in date pickers

**No DSL changes are required for basic number and date display** — AMIS handles
these once the locale is set at the SDK level.

---

## 36.5 Planned: Translation Pipeline

> **PLANNED** — not yet implemented.

The following capabilities are on the roadmap but have no implementation today:

| Capability | Description |
|---|---|
| Translation key injection | Block functions emit translation keys (e.g. `"label.invoice.amount"`) instead of raw English strings; a pipeline pass resolves keys against a locale bundle. |
| i18n pipeline stage | A dedicated `I18nStage` after `CompileStage` that walks the compiled schema tree and replaces key tokens with resolved strings. |
| Runtime locale switching | Allow tenants or users to switch locale without a full page reload; invalidate only locale-sensitive cache entries. |
| RTL layout support | Automatic `dir="rtl"` injection for Arabic/Hebrew locales; AMIS SDK RTL stylesheet loading. |
| Plural forms | Translation bundle support for plural rules (e.g. "1 item" vs "3 items") across supported locales. |

Until these are implemented, English string literals in DSL blocks are
acceptable.  When the translation pipeline lands, a codemod will scan for bare
string literals and offer to replace them with key tokens.
