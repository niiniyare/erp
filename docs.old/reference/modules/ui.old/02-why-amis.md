[<-- Back to Index](README.md)

## Why AMIS

### The Alternative — Traditional Frontend

```markdown
TRADITIONAL REACT/VUE APPROACH:
├── Write 50+ component files per module
├── Wire up state management (Redux / Pinia)
├── Build custom filter components
├── Build custom pagination
├── Build custom table with sort
├── Build form validation
├── Write CSS for every component
├── Maintain component tests
└── Build pipeline: Webpack/Vite, TypeScript, ESLint...

FOR EACH MODULE:
4-8 weeks of frontend work per module × 10 modules = 40-80 weeks
```

```markdown
AMIS APPROACH:
├── Write JSON schema files (or Go structs that produce them)
├── One crud component = table + filter + sort + pagination + bulk actions
├── One form component = validation + conditional fields + API submit
├── One wizard component = multi-step flow with back/forward
└── No build pipeline required (CDN or pre-built sdk.js)

FOR EACH MODULE:
1-3 days of schema writing per module × 10 modules = 2-6 weeks
```

### Why AMIS Fits ERP Specifically

ERP software has predictable, repeating UI patterns. AMIS was built for exactly this:

| ERP Need | AMIS Component |
|---|---|
| List of records with filter/sort/pagination | `crud` |
| Create / edit record form | `form` |
| Multi-step process (payroll run, onboarding) | `wizard` |
| Quick edit without leaving the list | `dialog` / `drawer` |
| Dashboard with KPIs and charts | `grid` + `stat` + `chart` |
| Approval history / activity log | `timeline` |
| Visual condition editor (business rules) | `condition-builder` |
| Inline table editing | `quickEdit` on columns |

### What AMIS Is NOT Good For

```markdown
AVOID AMIS FOR:
❌ Highly custom, pixel-perfect marketing pages
❌ Real-time collaborative features (no native WebSocket)
❌ Complex drag-and-drop beyond basic list reordering
❌ Custom map pickers, signature pads, barcode scanners
   → These need custom renderers (see §24 Known Issues)

USE AMIS FOR:
✅ Everything that looks like an admin panel or ERP screen
✅ 95%+ of Awo ERP's surface area
```

### Why the Custom Shell (Not AMIS `app` Component)

AMIS has an `app` component that provides a sidebar. We chose NOT to use it because:

```markdown
AMIS app COMPONENT LIMITATIONS:
├── Sidebar styling is CSS-class-based — hard to override cleanly
├── Dark mode does not exist in AMIS (see §05)
├── Mobile responsiveness requires custom CSS fighting AMIS's own
├── Collapse animation is limited
└── Cannot control the nav JS behaviour independently

CUSTOM SHELL ADVANTAGES:
├── Full control over sidebar, collapse, mobile drawer
├── Dark mode as pure CSS variable swap (no AMIS involvement)
├── Theme system (system/light/dark) with localStorage
├── AMIS only renders in #content — zero conflict with shell CSS
└── Menu config in JS → easy to switch to Go-driven later
```

---
