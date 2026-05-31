[<-- Back to Index](README.md)

## Design Tokens

### Typography Scale

One typeface. One scale. No exceptions. Deviation from this creates visual noise that undermines the professional feel of the product.

| Token | Size | Weight | Line Height | Use |
|---|---|---|---|---|
| `text-xs` | 11px | 400 | 1.5 | Table metadata, secondary labels |
| `text-sm` | 13px | 400 | 1.5 | Table body, form labels |
| `text-base` | 15px | 400 | 1.6 | Primary body copy |
| `text-lg` | 17px | 500 | 1.4 | Card titles, section headings |
| `text-xl` | 20px | 600 | 1.3 | Page subtitles |
| `text-2xl` | 24px | 700 | 1.2 | Page titles |
| `text-3xl` | 30px | 700 | 1.1 | Dashboard KPI numbers |
| `text-4xl` | 36px | 800 | 1.0 | Hero numbers |

AMIS uses its own internal font scale. You do not control it per-component. Control it globally via `font-size` on `body` and let AMIS's `em`-based scale inherit.

### Spacing Scale

All spacing values are multiples of 4px:

```
4 · 8 · 12 · 16 · 20 · 24 · 32 · 40 · 48 · 64 · 80 · 96 · 128
```

Never use values outside this scale. `margin: 5px` breaks visual rhythm. `margin: 4px` or `margin: 8px` — always.

### Status Colours — Consistent Across Every Module

A user who learns that green means "active" in the HR module must never encounter green meaning something else anywhere else. This is non-negotiable.

| Status | Background | Text | Use |
|---|---|---|---|
| Active / Success | Green-100 / Green-800 | Approved requests, active employees |
| Pending / Warning | Amber-100 / Amber-800 | Awaiting approval, pending payment |
| Draft | Grey-100 / Grey-700 | Unsaved or unsubmitted records |
| Processing | Blue-100 / Blue-800 | Running workflows, loading states |
| Error / Failed | Red-100 / Red-800 | Failed runs, rejected requests |
| Suspended | Orange-100 / Orange-800 | Suspended employees, paused processes |
| Terminated / Closed | Grey-200 / Grey-600 | Terminal states |
| Reversed | Purple-100 / Purple-800 | Reversed payroll runs, cancelled transactions |

In AMIS `tag` component, these map to: `success`, `warning`, `default`, `processing`, `error`.

### Motion & Animation

Animation serves function. Not decoration. Not personality. Not filling time while data loads.

| Motion Type | Duration | Easing | Use |
|---|---|---|---|
| Micro-interaction | 100–150ms | ease-out | Button press, checkbox toggle |
| UI state change | 150–200ms | ease-in-out | Tab switch, dropdown open |
| Content transition | 200–300ms | ease-in-out | Page-level transitions |
| Skeleton shimmer | Infinite loop | linear | Loading placeholder |

Never animate anything that does not help the user understand what is happening.

Always respect `prefers-reduced-motion`:

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

---
