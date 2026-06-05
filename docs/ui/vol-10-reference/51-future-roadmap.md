---
title: "51 – Future Roadmap"
volume: "vol-10-reference"
chapter: 51
section: "Reference"
status: "reference — open items from tasks.md and review.md"
---

# Chapter 51 – Future Roadmap

This chapter tracks all open development items for the AwoERP UI platform.
Items are sourced from `tasks.md` open items and `review.md` follow-up notes.
Each item carries a priority tier and a rationale.

## Table of Contents
- [P0 — Blocking Current Work](#p0--blocking-current-work)
- [P1 — High Value, Next Quarter](#p1--high-value-next-quarter)
- [P2 — Important Infrastructure](#p2--important-infrastructure)
- [P3 — Developer Experience](#p3--developer-experience)
- [Planned: Platform Expansion](#planned-platform-expansion)

---

## P0 — Blocking Current Work

### Param Routing in Registry (OPEN item 2)

**Problem**: The registry matches routes by exact string.  Routes like
`/finance/invoices/:id` (for edit/detail screens) cannot be registered because
`:id` is a path parameter, not a literal string.

**Impact**: All document detail and edit screens are blocked.  Currently,
detail views are implemented as dialogs launched from list screens — a
workaround that limits deep linking and browser history support.

**Required work**:
- Add pattern-matching route resolution to `RegistryStage` (e.g. using
  `gorilla/mux`-style pattern matching or a simple `:param` segment scanner).
- Extract matched parameters into `PipelineContext` so `ASTPageFn` can access
  `id` via `sess.PathParam("id")`.
- Add `Route: "/finance/invoices/:id"` registration support to `ValidateRegistry`.

**Unblocks**: invoice detail screen, supplier detail, payment detail, and all
other document-level routes.

---

## P1 — High Value, Next Quarter

### Non-Finance Domain DSL Blocks

**Status**: Finance domain blocks are substantially complete.  Other domains
have no DSL blocks yet — screens for those domains are either missing or use
the legacy `PageFn` + `amis.*` builder path.

**Planned block sets:**

| Domain | Planned blocks |
|---|---|
| Inventory | `StockLevelBlock`, `WarehouseMapBlock`, `LowStockAlertBlock`, `GoodsReceiptBlock` |
| HR | `EmployeeCardBlock`, `LeaveBalanceBlock`, `PayslipBlock`, `OrgChartBlock` |
| Procurement | `SupplierListBlock`, `PurchaseOrderBlock`, `ApprovalQueueBlock`, `GRNBlock` |
| Sales | `CustomerCardBlock`, `QuotationBlock`, `SalesOrderBlock`, `DeliveryStatusBlock` |

Each domain block set should follow the same patterns as the finance blocks:
typed config structs, permission-gated rendering, no `map[string]any`.

### Remaining Screen Migrations

Screens using the deprecated `Register` API or legacy `PageFn`:

| Screen | Action needed |
|---|---|
| `/iam/users` | Migrate to `RegisterPage` + `ASTPageFn` |
| `/iam/roles` | Migrate to `RegisterPage` + `ASTPageFn` |
| `/tenants` | Migrate from deprecated `Register` + `ASTPageFn` |
| `/settings` | Migrate from deprecated `Register` + `ASTPageFn` |

See §44.5 for full migration status table.

---

## P2 — Important Infrastructure

### CI Architecture Guards (Task 13)

**Status**: PLANNED — not yet implemented.

A shell script `scripts/check-arch.sh` enforcing the 8 architectural rules
currently maintained by code review convention only:

1. No `map[string]any` in `internal/web/`
2. No direct IAM imports from `dsl/`
3. No permission string checks in `visibleOn` predicates
4. No schema generated outside the pipeline
5. Every stage has non-empty `DependsOn`
6. No `sync.Map` or third-party cache in `internal/web/`
7. No block-level concerns defined in multiple `screens/` files
8. Every `screens/` file is under 60 lines

See §42.2 and §42.3 for full specification.

### CRUDNode.Children() Fix

**Problem**: `CRUDNode.Children()` does not include `RowActions` in its returned
slice.  Validation errors inside row actions are silently missed by `CompileTree`.

**Required work**: Update `CRUDNode.Children()` to return `RowActions` alongside
`Columns` and `Filter`.  Write regression tests to confirm that a row action
missing its `Dialog` is caught by `CompileTree`.

See §47 anti-pattern #11.

### Formal Schema Versioning

**Status**: PLANNED — see §43.5.

Required before mobile SDK development can begin.  The server must emit schema
version information in HTTP response headers so clients can detect breaking
changes across app store release cycles.

---

## P3 — Developer Experience

### Go-Driven Navigation (NavFn to Web Shell)

**Status**: PLANNED.

Currently, the sidebar navigation in `web/pages/index.html` is a static HTML
list.  Module teams cannot add navigation entries from Go code — they must edit
the HTML manually.

The planned `NavFn` architecture:
- Module registration includes a `NavEntry` (label, icon, route, permission).
- A `NavStage` in the pipeline builds the navigation tree from registered entries,
  filtered by the current user's permissions.
- The web shell fetches the nav tree from a `/ui/nav` endpoint and renders it
  dynamically.
- Result: adding a new module screen automatically adds it to the nav without
  touching `index.html`.

### Deprecation Warnings for Legacy PageFn

Add a log warning when `CompileStage` executes via the legacy `PageFn` path,
so teams are notified to complete migration.  Currently there is no automated
signal for remaining legacy screens.

---

## Planned: Platform Expansion

The following items are on the long-term roadmap.  No implementation timeline
exists for these items.

### Flutter Client

A native Flutter application backed by the AwoERP schema pipeline.  The schema
JSON produced by the Go pipeline would drive a Flutter renderer, giving mobile
clients the same server-controlled UI definition as the web shell.

Prerequisite: formal schema versioning (§43.5) and the Flutter SDK described in
§50.3.

### React Native Client

A React Native application following the same schema-driven approach.  Requires
the `@awoerp/ui-schema` npm package described in §50.3.

### Tenant Customization UI

A UI surface allowing tenant administrators to customise:
- Which modules are visible in the navigation
- Default locale, timezone, and currency settings
- Feature flag overrides at the tenant level
- Custom dashboard widget layout

This requires a new `TenantCustomizationStage` in the pipeline that reads
tenant-specific overrides and merges them into the compiled schema.

### Plugin Architecture

A mechanism for third-party or customer-developed modules to register their own
DSL screens and blocks without modifying the core repository.

Requires:
- A plugin registration protocol (separate binary or shared library)
- Sandboxed execution of plugin `ASTPageFn` functions
- Plugin permission namespace isolation

### gRPC Streaming for Live Schema Updates

Replace the poll-on-navigation schema fetch with a gRPC server-streaming
connection.  When a deploy invalidates a schema, the server pushes an
invalidation event to connected browsers, which refetch the schema for the
current route without a user-visible page reload.

Requires:
- A gRPC streaming endpoint alongside the existing HTTP schema endpoint
- Client-side EventSource or gRPC-Web subscription in the web shell
- Invalidation event emission from `CacheStage` on generation mismatch

### Accessibility Audit and WCAG 2.1 Compliance

Formal accessibility audit and remediation of identified gaps.  See §37.4.

---

## Roadmap Summary

| Priority | Item | Status |
|---|---|---|
| P0 | Param routing (`:id` routes) | OPEN — blocking edit screens |
| P1 | Inventory / HR / Procurement DSL blocks | PLANNED |
| P1 | Legacy screen migration | IN PROGRESS |
| P2 | CI architecture guard script (Task 13) | PLANNED |
| P2 | CRUDNode.Children() fix | PLANNED |
| P2 | Formal schema versioning | PLANNED |
| P3 | Go-driven navigation (NavFn) | PLANNED |
| P3 | Legacy PageFn deprecation warnings | PLANNED |
| Future | Flutter client | PLANNED |
| Future | React Native client | PLANNED |
| Future | Tenant customisation UI | PLANNED |
| Future | Plugin architecture | PLANNED |
| Future | gRPC streaming schema updates | PLANNED |
| Future | Accessibility WCAG 2.1 audit | PLANNED |
