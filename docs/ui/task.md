# SDUI Implementation Tasks

> **Status tracking for AwoERP UI platform implementation.**
> Each task has: goal, expected behavior (acceptance criteria), files, notes.
> Work tasks in order. After each task: review → approve → mark done.

---

## Task Status Key

| Symbol | Meaning |
|--------|---------|
| ⬜ | Not started |
| 🔄 | In progress |
| ✅ | Done — approved |
| ❌ | Blocked |

---

## T-0 — Demo Shell: View in Browser (No Server) ✅

**Goal:** Open `web/pages/demo.html` in browser without running the Go server. View and interact with the full app shell + sample AMIS schemas.

**Expected behavior:**
- Browser opens `http://localhost:8080/pages/demo.html` (via `python3 -m http.server 8080` from `web/`)
- Sidebar renders: Dashboard, Finance group (Overview, Accounts, Invoices, etc.), Administration, Settings
- Clicking nav items loads a sample AMIS schema in the content area
- Theme toggle (System/Light/Dark) works — dark mode applies AMIS token overrides
- Sidebar collapse/expand works
- Mobile responsive (hamburger menu on narrow viewport)
- Routes with no embedded schema show a friendly "Demo mode — no schema for this route" info card
- No Go server, no auth, no real API calls

**How to view:**
```bash
# From project root:
cd web
python3 -m http.server 8080
# Then open: http://localhost:8080/pages/demo.html
# To stop: Ctrl+C
```

**Files:**
- `web/pages/demo.html` — self-contained demo shell with embedded schemas
- `web/schemas/pages/dashboard.json` — main dashboard schema (static)
- `web/schemas/pages/finance/dashboard.json` — finance overview schema
- `web/schemas/pages/finance/invoices/new.json` — invoice form schema

**Notes:**
- `demo.html` embeds schemas as `DEMO_SCHEMAS` JS object — no backend needed
- Real `index.html` unchanged — still calls `/schema/<route>` (Go pipeline)
- AMIS SDK loaded from `../sdk/` — requires HTTP server (not `file://`)

---

## T-1 — Fix unauthenticatedEnvelope ✅

**Goal:** Replace raw `map[string]any` in `unauthenticatedEnvelope` (handler/schema.go) with typed `ast.PageNode`.

**Expected behavior:**
- Unauthenticated request to `/schema/*` returns valid AMIS page schema
- Page has title "Session Expired", body with alert + login button (`ActionNode{ActionType: "url", URL: "/ui/login"}`)
- ValidateStage passes on this schema (no AST violations)
- Response envelope: `{"status": 401, "data": <amis-page-schema>}`

**Files:**
- `internal/web/handler/schema.go` — replace `unauthenticatedEnvelope` map literal

**Notes:**
- Low-risk, ~20 lines. Purely cosmetic for now but sets precedent: no raw maps in handler layer.

---

## T-2 — CI Architecture Guards ⬜

**Goal:** Shell script that enforces platform-wide architectural rules. Runs in CI and as a pre-commit hook.

**Expected behavior:**
- `make check-arch` (or `scripts/check-arch.sh`) exits 0 on clean repo
- Fails with specific errors on violations:
  - Any `map[string]any` literal inside `internal/web/dsl/` → `ARCH001: zero-map rule violated`
  - Any screen file in `internal/web/dsl/screens/` with >60 lines → `ARCH002: screen too large`
  - Any `ASTPageFn` that doesn't type-assert to `ast.Node` at call site → `ARCH003` (grep-based)
  - Any `PageFn` registered for a new route (new routes must use `ASTPageFn`) → `ARCH004`

**Files:**
- `scripts/check-arch.sh` — main guard script
- `Makefile` — add `check-arch` target
- `.github/workflows/ci.yml` (or equivalent) — add `check-arch` step

**Notes:**
- Grep-based rules are fast and simple. No AST parsing needed.
- ARCH002: `wc -l` each screen file.
- ARCH001: `grep -r "map\[string\]any{" internal/web/dsl/`

---

## T-3 — Missing AST Nodes ✅

**Goal:** Implement 4 AST node types needed for navigation, multi-step forms, and cross-entity selectors.

**Expected behavior (per node):**

| Node | AMIS type | Validate rule | Compile output |
|------|-----------|--------------|----------------|
| `NavNode` | `"nav"` | Must have ≥1 link | `{type:"nav", links:[...]}` |
| `BreadcrumbNode` | `"breadcrumb"` | Must have ≥1 item | `{type:"breadcrumb", items:[...]}` |
| `WizardNode` | `"wizard"` | Must have ≥2 steps | `{type:"wizard", steps:[...], api:{...}}` |
| `PickerNode` | `"picker"` | Must have `source` | `{type:"picker", source:"...", ...}` |

- All implement `ast.Node` interface
- `WizardNode` and `NavNode` implement `ast.ContainerNode` (have children)
- Compile-time assertions: `var _ ast.Node = NavNode{}`
- Error codes: `AST_NAV_EMPTY`, `AST_BREADCRUMB_EMPTY`, `AST_WIZARD_TOO_FEW_STEPS`, `AST_PICKER_NO_SOURCE`

**Files:**
- `internal/web/ast/layout.go` — add `NavNode`, `BreadcrumbNode`
- `internal/web/ast/form.go` — add `WizardNode`, `PickerNode`

**Notes:**
- `PickerNode` enables cross-entity selectors (e.g., "Select Vendor" from supplier list)
- `WizardNode` needed for multi-step procurement flows (T-6)
- Do T-3 before T-4 through T-9

---

## T-4 — ProductServiceLineBlock ✅

**Goal:** Implement the shared `ProductServiceLineBlock` — used by all document-type screens (invoices, purchase orders, bills, credit notes, delivery notes, goods receipts, journal entries).

**Expected behavior:**
- Block emits an editable line-items table with columns: Item, Description, Qty, Unit Price, Tax %, Amount
- Add/remove rows (AMIS `"combo"` or `"input-table"` component)
- Amount column: auto-computed from `Qty × UnitPrice × (1 + TaxRate)`
- Subtotal / Tax Total / Grand Total summary row below table
- `ProductServiceLineConfig` struct: `AllowDelete bool`, `CurrencyCode string`, `TaxRates []TaxRateOption`
- Block reusable across all 8 document types — no duplication

**Files:**
- `internal/web/dsl/blocks/product_service_line.go` — new block

**Notes:**
- This is the most important block in the system (review.md: "most critical block")
- Implement this before any document-type screen (T-5 through T-8)
- Test manually: does the line-items table render, can rows be added/removed, does total compute?

---

## T-5 — Sell Module DSL ✅

**Goal:** DSL screens for Sales module (quotations, sales orders, delivery notes, sales invoices).

**Expected behavior:**
- `QuotationScreen` — draft quotation with customer picker, validity date, `ProductServiceLineBlock`
- `SalesOrderScreen` — converts from quotation; status badge (DRAFT→CONFIRMED→DELIVERED→INVOICED)
- `DeliveryNoteScreen` — goods delivery confirmation; links to sales order
- `SalesInvoiceScreen` — reuses `InvoiceScreen` pattern (already exists); sales-side variant

**Files:**
- `internal/web/dsl/screens/sell_screens.go` — all sell screens (≤60 lines each)
- `internal/web/dsl/screens/register.go` — add sell routes

**Routes to register:**
```
/sell/quotations/new      → QuotationScreen
/sell/orders/new          → SalesOrderScreen
/sell/delivery-notes/new  → DeliveryNoteScreen
```

**Notes:**
- All screens <60 lines (composition only — logic in blocks)
- Customer picker uses `PickerNode` from T-3

---

## T-6 — Buy Module DSL ✅

**Goal:** DSL screens for Procurement module (requisitions, purchase orders, goods receipts, purchase invoices).

**Expected behavior:**
- `RequisitionScreen` — internal purchase request; requestor, department, items list
- `PurchaseOrderScreen` — approved PO sent to supplier; `ProductServiceLineBlock`, supplier picker
- `GoodsReceiptScreen` — receive goods against PO; quantity received per line
- Wizard pattern for multi-step approval flow (requires `WizardNode` from T-3)

**Files:**
- `internal/web/dsl/screens/buy_screens.go`
- `internal/web/dsl/screens/register.go` — add buy routes

**Routes:**
```
/buy/requisitions/new     → RequisitionScreen
/buy/purchase-orders/new  → PurchaseOrderScreen
/buy/goods-receipts/new   → GoodsReceiptScreen
```

---

## T-7 — Inventory Module DSL ✅

**Goal:** DSL screens for Inventory module (stock ledger, warehouse management, item catalog).

**Expected behavior:**
- `StockLedgerScreen` — paginated CRUD of stock movements; filters: warehouse, item, date range
- `WarehouseScreen` — warehouse list + detail with current stock levels
- `ItemCatalogScreen` — item master data; unit of measure, reorder point, supplier links

**Files:**
- `internal/web/dsl/screens/inventory_screens.go`
- `internal/web/dsl/screens/register.go` — add inventory routes

**Routes:**
```
/inventory/stock-ledger   → StockLedgerScreen
/inventory/warehouses     → WarehouseScreen
/inventory/items          → ItemCatalogScreen
```

---

## T-8 — HR Module DSL ✅

**Goal:** DSL screens for HR module (employees, leave management, payroll summary).

**Expected behavior:**
- `EmployeeScreen` — employee detail: personal info, department, position, contract dates
- `LeaveRequestScreen` — leave request form: type, date range, approval chain timeline
- `PayrollSummaryScreen` — dashboard: headcount, total payroll, pending approvals

**Files:**
- `internal/web/dsl/screens/hr_screens.go`
- `internal/web/dsl/screens/register.go` — add HR routes

**Routes:**
```
/hr/employees/new         → EmployeeScreen
/hr/leave-requests/new    → LeaveRequestScreen
/hr/payroll               → PayrollSummaryScreen
```

---

## T-9 — AuthZ Module DSL ✅

**Goal:** DSL screens for IAM/Authorization module (roles, policies, user management).

**Expected behavior:**
- `RoleScreen` — role detail: name, description, assigned permissions (checklist), members
- `PolicyScreen` — Casbin policy list: subject, domain, object, action; add/remove rows
- `UserScreen` — user detail: profile, active roles, session history (TimelineNode)

**Files:**
- `internal/web/dsl/screens/authz_screens.go`
- `internal/web/dsl/screens/register.go` — add authz routes

**Routes:**
```
/iam/roles/new            → RoleScreen
/iam/policies             → PolicyScreen
/iam/users/:id            → UserScreen (param route)
```

**Notes:**
- `UserScreen` is the first param route outside finance — validates param routing works for authz

---

## T-10 — Docs: Promote Partial Chapters ⬜

**Goal:** Complete the 6 chapters marked "✅ Partial" in `docs/ui/README.md`. Each needs its stubbed sections filled with actual implementation content.

**Expected behavior (per chapter):**

| Chapter | What's missing |
|---------|---------------|
| Ch 19 Navigation | NavNode, BreadcrumbNode patterns (blocked on T-3) |
| Ch 21 Event System | Event bus integration with AMIS actions |
| Ch 22 State Management | Cross-tab state, optimistic updates |
| Ch 30 Customization | Per-tenant schema overrides (if implemented) |
| Ch 35 Synchronization | Stale schema detection, forced refresh |
| Ch 36 Internationalization | i18n key injection via UISessionContext.Locale |

**Files:**
- `docs/ui/vol-04-rendering/19-navigation-framework.md`
- `docs/ui/vol-04-rendering/21-event-system.md`
- `docs/ui/vol-05-runtime/22-state-management.md`
- `docs/ui/vol-06-platform/30-customization-framework.md`
- `docs/ui/vol-07-reliability/35-synchronization.md`
- `docs/ui/vol-08-dx/36-internationalization.md`

**Notes:**
- Do Ch 19 after T-3 (NavNode). Others can go in parallel.
- Update `docs/ui/README.md` status column from `✅ Partial` to `✅ Full` when done.

---

## T-11 — Planned Features (Future) ⬜

**Goal:** Implement features currently documented as "🔲 Planned". Each is a separate initiative.

| Feature | Chapter | Prerequisite |
|---------|---------|-------------|
| Flutter mobile client | Ch 17 | Stable REST contract + T-5 through T-9 |
| Plugin / extension architecture | Ch 31 | T-2 guards pass, T-9 done |
| Offline support (service worker) | Ch 33 | Infrastructure decision |
| Accessibility audit | Ch 37 | Design system decision |
| SDK packaging (npm) | Ch 50 | Stable AMIS API surface |
| Cache invalidation webhooks | Ch 08 §8.13 | Business events wired |

**Notes:**
- No code changes in this task — it's a planning checkpoint.
- When a feature is resourced, break it out into its own task set.
- Update chapter status in `docs/ui/README.md` when each goes from Planned → implemented.

---

## Progress Summary

| Task | Description | Status |
|------|-------------|--------|
| T-0 | Demo shell — view in browser | ✅ |
| T-1 | Fix unauthenticatedEnvelope | ✅ |
| T-2 | CI architecture guards | ⬜ |
| T-3 | Missing AST nodes (Nav, Breadcrumb, Wizard, Picker) | ✅ |
| T-4 | ProductServiceLineBlock | ✅ |
| T-5 | Sell module DSL | ✅ |
| T-6 | Buy module DSL | ✅ |
| T-7 | Inventory module DSL | ✅ |
| T-8 | HR module DSL | ✅ |
| T-9 | AuthZ module DSL | ✅ |
| T-10 | Docs: promote partials to full | ⬜ |
| T-11 | Planned features (future) | ⬜ |
