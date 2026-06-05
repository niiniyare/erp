# UI Module Reference — Deprecation Notice

> **Status: Deprecated**
> **Replaced by:** `docs/ui/` (restructured documentation tree)
> **Date:** 2026-06-05

---

## What Was Here

This directory was planned as the reference location for UI module documentation. It was not populated with content before the documentation was reorganized into the `docs/ui/` volume structure.

---

## Where to Find Current Documentation

The authoritative UI platform documentation is at:

```
docs/ui/
├── README.md                          — index and pipeline quick reference
├── claude-review.md                   — audit of original docs with error list
├── vol-01-vision/
│   ├── 01-introduction.md
│   ├── 02-design-philosophy.md
│   ├── 03-architectural-principles.md
│   └── 04-system-overview.md
├── vol-02-dsl-and-ast/
│   ├── 05-sdui-fundamentals.md
│   ├── 06-ui-dsl-architecture.md
│   ├── 07-ast-design.md
│   └── 08-compilation-pipeline.md
└── vol-03-component-system/
    ├── 09-component-system.md
    ├── 10-layout-system.md
    ├── 11-forms-framework.md
    └── 12-tables-and-data-grids.md
```

---

## Key Code Locations

For engineers who want the source of truth:

| Package | Path | Description |
|---------|------|-------------|
| `ui` | `internal/web/ui/` | Core types: `UISessionContext`, `PageFn`, `ASTPageFn`, pipeline constants |
| `ast` | `internal/web/ast/` | Typed node structs, `CompileTree` |
| `registry` | `internal/web/registry/` | `RegisterPage`, `Match` |
| `handler` | `internal/web/handler/` | `SchemaHandler` — `/schema/*` HTTP handler |
| `amis` | `internal/web/amis/` | Legacy fluent builders (deprecated for new pages) |
| `dsl/blocks` | `internal/web/dsl/blocks/` | DSL blocks: `DataTableBlock`, `StatusBadgeBlock`, etc. |
| `dsl/screens` | `internal/web/dsl/screens/` | DSL screens: `FinanceDashboardScreen`, `InvoiceScreen`, etc. |
| `dsl/builders` | `internal/web/dsl/builders/` | Module-domain builder helpers |
| Web SDK | `web/sdk/` | AMIS browser SDK (sdk.js, sdk.css) |
| Web pages | `web/pages/` | Browser entry point (index.html) |
| Web schemas | `web/schemas/` | Static JSON schemas (non-pipeline pages) |

---

## Summary of What the UI Platform Is

AwoERP uses **Server-Driven UI (SDUI)** with **AMIS** as the only rendering client today. The backend compiles AMIS JSON schemas at request time, incorporating Casbin permissions and feature flags. Schemas are cached in Redis. The browser's AMIS SDK renders whatever schema it receives — no authorization logic in the browser.

Pipeline: 9 stages, priorities 10–90. Cache runs at priority 30, after authorization at priority 20. This ordering is required because the cache key includes the permission fingerprint.

> **Mobile clients (Flutter, React Native), gRPC streaming, schema versioning, and no-code customization are planned but not yet implemented.**
