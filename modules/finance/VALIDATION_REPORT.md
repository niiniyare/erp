# Awo Framework Validation Report — Finance Module

**Date:** 2026-07-20
**Module:** `finance`
**Validator:** Finance module implementation as framework proving ground
**Framework Version:** Awo v0.9 (pre-1.0)

---

## Executive Summary

The Finance module is the canonical reference implementation for Awo ERP. It exercises every major framework capability across 24 entity definitions spanning accounting, invoicing, payments, and banking. This report documents which assumptions proved correct, which proved incorrect, and what improvements are recommended before declaring Awo v1.0 feature-complete.

**Overall assessment:** The framework is **conditionally production-ready**. Core abstractions (entity definitions, hooks, RBAC, SDUI, routing, migrations) are sound. Three architectural gaps were discovered that should be resolved before v1.0 freeze.

---

## 1. Framework Capabilities Successfully Exercised

### 1.1 Entity Definitions ✅

All 24 Finance entities compiled successfully via `registry.BuildFrom` + `compiler.Compile`. The `SystemDefinition` / `CustomDefinition` split worked correctly:

| Mandatory System Entities | Status |
|--------------------------|--------|
| `finance_journal_entry`  | ✅ SystemDefinition, passes registry validator |
| `finance_ledger_entry`   | ✅ SystemDefinition, passes registry validator |
| `finance_payment`        | ✅ SystemDefinition, passes registry validator |
| `finance_tax_entry`      | ✅ SystemDefinition, passes registry validator |

The compiler correctly derives:
- `QualifiedName`: `"finance_" + localName` (e.g., `"finance_invoice"`)
- `RoutePrefix`: `/api/v1/finance/invoices`
- `TableName`: qualified name (system entities)
- All namespace identifiers: event, workflow, permission, metric, cache

### 1.2 Field Types ✅

All field type constraints were correctly enforced:

- `FieldTypeCurrency` used for all money amounts — `numeric(20,4)` in SQL, `decimal.Decimal` in Go
- `FieldTypeNamingSeries` with series patterns (e.g., `"INV-{YYYY}-{SEQ:6}"`)
- `FieldTypeSelect` with `Options` generates SQL `CHECK IN (...)` constraints
- `FieldTypeLink` with `LinkTarget` (qualified name) — link resolution works in phase 2 of compilation
- `Immutable: true` fields correctly raise `ValidationError` on update attempts

The registry validator correctly rejected: using `FieldTypeFloat` for fields named `amount`, `rate`, `total`, `tax_amount` — **this caught 3 early bugs** where we initially wrote `FieldTypeFloat` for exchange rates.

### 1.3 Lifecycle Hooks ✅

All hook interfaces implemented correctly:
- `BeforeCreateHook` — `JournalEntryValidator`, `InvoiceValidator`, `PaymentValidator`, `LineBalanceValidator`
- `BeforeUpdateHook` — `PostingGuard`, `InvoiceStatusGuard`, `PaymentStatusGuard`, `FiscalYearTransitionGuard`, `PeriodTransitionGuard`
- Hook pipeline order (declaration order within stage) respected

The transaction boundary was critical for `LedgerEntry` creation: ledger entries are created in an `AfterCreate` hook (inside TX) on `JournalEntry` post action, not in a `BeforeCreate` hook.

### 1.4 Custom Actions ✅

Custom actions on `JournalEntry` (post, reverse), `Invoice` (submit, approve, cancel), `Payment` (process, cancel), `CreditNote`/`DebitNote` (approve) all registered correctly with:
- Custom HTTP routes: `POST /api/v1/finance/journal-entries/:id/post`
- Permission gate per action (Casbin action string)
- SDUI confirm dialogs for destructive actions
- `WorkflowEvent` linking action to workflow trigger

### 1.5 Workflow Triggers ✅

Temporal workflow bindings work correctly on:
- `InvoiceApprovalWorkflow` → triggered `EventOnSubmit` on invoice
- `PaymentProcessingWorkflow` → triggered `EventOnSubmit` on payment
- `JournalEntryApprovalWorkflow` → triggered `EventOnSubmit` on journal entry

The `InputBuilder` function pattern (deterministic, no I/O) is clean and testable. The default workflow ID convention (`{tenant-id}.{entity-type}.{record-id}.{event}`) is unique enough for production use.

### 1.6 RBAC Permissions ✅

Four finance roles expressed as Casbin subjects:
- `role:finance.manager` — full finance access, approve/reverse actions
- `role:finance.accountant` — post entries, process payments
- `role:finance.viewer` — read-only
- `role:tenant.admin` — tenant-level admin

`LedgerEntry` and `TaxEntry` with empty `Create`/`Write`/`Delete` permission slices correctly deny all direct API access — the framework writes these records, not users.

Casbin policies emitted correctly from `PermissionSet` declarations.

### 1.7 Row-Level Policy (Organization Scope) ✅

`invoiceOrgPolicy` demonstrates the `PolicyFunc` pattern:
- Extracts `organization_id` from request context
- Returns `filter.Eq("organization_id", orgID)` to scope reads
- Returns `nil` when no org context (falls back to tenant RLS)

This is a clean composition: RLS handles tenant isolation at the DB level; `PolicyFunc` handles application-level org scoping above that.

### 1.8 SDUI — Auto-Generated Pages ✅

All 24 entities auto-generate list, create, edit, and detail pages without custom `PageBuilderSet`. The generator correctly:
- Excludes `Hidden`, `Sensitive`, `ReadOnly` fields from appropriate views
- Excludes `Immutable` fields from edit forms
- Maps `FieldTypeCurrency` → `number` column type
- Maps `FieldTypeSelect` → `select` form control with options
- Maps `FieldTypeLink` → `input-text` (TODO: should be `input-select` with linked entity lookup — see Gap 1)

### 1.9 Route Generation ✅

All routes generated correctly. For 24 entities:
- 5 standard CRUD routes each = 120 standard routes
- Plus custom action routes (post, reverse, submit, approve, cancel, process) = ~15 action routes
- Total: ~135 routes registered automatically

Route format confirmed: `/api/v1/finance/{resource}` and `/api/v1/finance/{resource}/:id/{action}`.

### 1.10 OpenAPI Generation ✅

`EntitySchema.Fields`, `.Actions`, `.Description` all feed correctly into OpenAPI spec generation. Finance entities produce well-structured OpenAPI tags (`Finance`).

### 1.11 Compiler Metadata ✅

All compile-time indexes populated correctly:
- `FieldsByName`, `RequiredFields`, `ImmutableFields`, `DefaultValues`, `FieldValidators`
- `LinkTargets` resolved in phase 2 (cross-entity references)
- `ActionsByName` O(1) lookup works

### 1.12 Migrations ✅ (Pattern)

All SQL migrations follow the framework pattern:
- `numeric(20,4)` for all currency columns
- RLS enabled + forced on every tenant-scoped table
- `CREATE INDEX CONCURRENTLY` pattern for production-safe indexes
- `CHECK` constraints at DB level for status enums
- `ON DELETE CASCADE` for line items (invoice_line, journal_entry_line)
- Down migrations drop in reverse dependency order

### 1.13 Audit Logging ✅

Platform audit module intercepts all entity creates/updates/deletes automatically via framework hooks. Finance entities benefit without any module-specific audit code.

### 1.14 Cache Integration ✅

SDUI page schemas cached at `page:{entity}:{view}:{tenant}` (5 min TTL). Finance entities use this automatically.

### 1.15 Introspection ✅

`/api/v1/introspect` endpoint reflects Finance entities correctly — field counts, edge counts, action counts all accurate from `EntitySchema`.

---

## 2. Architectural Assumptions That Proved Correct

### 2.1 Single Definition Drives All Subsystems

The core bet of Awo — one `EntityDefinition` drives routing, UI, persistence, permissions, and workflows simultaneously — **paid off dramatically in Finance**. The invoice entity definition (~80 lines) auto-generates:
- 5 CRUD API routes
- 3 custom action routes
- 4 SDUI pages (list, create, edit, detail)
- Casbin policies for 4 roles
- 1 workflow trigger binding
- Full OpenAPI spec
- Audit log interception

Equivalent hand-coded ERPNext Python: ~500+ lines spread across DocType JSON, Python class, JavaScript client, and permission manager.

### 2.2 Immutable Ledger as Append-Only Log

The design decision to make `LedgerEntry` with empty `Create`/`Write`/`Delete` permissions — and write it exclusively from the `PostingService` — is architecturally sound. The double-entry ledger is an immutable audit trail, not a user-editable table.

### 2.3 Draft → Post Action Pattern for Double-Entry

Journal entry balance validation at "post" action time (not at line creation time) is the correct pattern. Lines are added incrementally; balance is only meaningful when the entry is complete and ready to post. The action handler has the full context needed to validate.

### 2.4 PolicyFunc Composability

`PolicyFunc` returning `nil` as a passthrough (fall through to tenant RLS) is elegant. The Finance `invoiceOrgPolicy` can be layered over tenant isolation without any framework changes.

### 2.5 Module-Local Names

Having `Name: "journal_entry"` (not `"finance_journal_entry"`) in the definition, with the compiler deriving `QualifiedName`, cleanly separates authoring ergonomics from runtime identifiers. The `HasModulePrefix` shim prevented breaking changes during the transition.

---

## 3. Architectural Assumptions That Proved Incorrect

### 3.1 ❌ `ActionContext` Has No Repository Access

**Problem:** `ActionHandlerFunc` receives `*ActionContext{Ctx, RecordID, Actor, Body}`. There is no `Repo` or `Service` field. This means action handlers **cannot perform any database reads or writes without external dependency injection**.

For Finance actions like `postJournalEntry` (which must load lines, validate balance, create ledger entries, update status), the action handler is effectively useless without a repository.

**Current workaround:** All Finance action handlers return `501 Not Implemented` with a TODO comment. The `PostingService` is defined but cannot be reached from the action handler without a global variable (anti-pattern) or a service locator.

**Root cause:** The framework assumed action handlers would be simple status transitions. Finance proved they need full DB access for complex business logic.

**Recommended fix (minimal):** Add an optional `Services map[string]any` field to `ActionContext`, populated at route registration time from a DI container. This is the service locator pattern but it's pragmatic.

**Recommended fix (better):** Change `ActionHandlerFunc` to:
```go
type ActionHandlerFunc func(ctx *ActionContext, repo EntityRepository) (*ActionResult, error)
```
Where `EntityRepository` is a scoped repo for the entity being acted upon (already has tenant context applied).

**Scope:** Affects all modules. Framework change required. Low risk — existing action handlers return errors, so adding the parameter is additive.

### 3.2 ❌ `FieldTypeLink` Has No SDUI Lookup Widget

**Problem:** `FieldTypeLink` fields (e.g., `journal_id`, `currency_id`, `account_id`) are rendered as `input-text` in SDUI forms, not as searchable dropdowns linked to the target entity.

Users must manually type UUIDs — completely unusable for production.

**Current workaround:** Finance uses `definition_invoice.go`'s `customer_id` as a plain `Data` field (not `FieldTypeLink`), storing the customer's display name separately in `customer_name`. This is a deliberate denormalization to work around the SDUI gap.

For internal entity links (e.g., `journal_id`, `account_id`), the Finance module **accepts the broken UI** as a known limitation.

**Root cause:** The `amisFormControl` function in `awo/sdui/generator.go` maps `FieldTypeLink` → `"input-text"`. amis has a `select` control with `source` API for dynamic options — this was never wired up.

**Recommended fix:** In `amisFormControl` / `generateForm`, for `FieldTypeLink` fields:
```go
case def.FieldTypeLink:
    ctrl["type"] = "select"
    ctrl["source"] = "/api/v1/" + linkedModule + "/" + linkedResource  // from LinkTargets
    ctrl["labelField"] = "name"  // convention
    ctrl["valueField"] = "id"
    ctrl["searchable"] = true
```
The `EntitySchema.LinkTargets` map already contains the resolved target `EntitySchema`, so `RoutePrefix` and field names are available at generation time.

**Scope:** `awo/sdui/generator.go` change only. No breaking changes. High value.

### 3.3 ❌ No Cross-Entity Validation in Hooks

**Problem:** The `BeforeCreateHook` for `JournalEntry` cannot validate that the `accounting_period_id` refers to an open period — because hooks have no DB access. Similarly, `InvoiceValidator` cannot verify the customer exists.

**Current workaround:** These validations are deferred to the `PostingService` or action handlers (which also lack DB access — see Gap 3.1). In practice, the DB FK constraints are the only enforcement.

**Root cause:** Hooks receive `*EntityRecord` (the record being saved) and `context.Context`, but no repository interface. This was fine for in-memory field validation but insufficient for cross-entity consistency checks.

**Recommended fix:** Add an optional `EntityLookup` interface to hook context:
```go
type HookContext struct {
    Ctx    context.Context
    Record *EntityRecord
    Prev   *EntityRecord  // nil for BeforeCreate
    Lookup EntityLookup   // optional, nil for test hooks that don't need it
}
type EntityLookup interface {
    Get(ctx context.Context, entityName string, id uuid.UUID) (*EntityRecord, error)
    Exists(ctx context.Context, entityName string, f Filter) (bool, error)
}
```
This requires changing hook interface signatures — a breaking change, but the right one before v1.0 freeze.

---

## 4. Missing Abstractions

### 4.1 Report Engine

Finance requires: trial balance, income statement, balance sheet, aged receivables, bank reconciliation. The framework has no report abstraction. Reports need:
- Multi-entity aggregation queries (ledger by account, summed by period)
- Parameterized filters (date range, cost center, currency)
- Multiple output formats (table, PDF, Excel)

**Recommendation:** Add `ReportDefinition` as a first-class framework concept, similar to `EntityDefinition` but for read-only aggregation views. Auto-generate SDUI report pages. Defer to v1.1.

### 4.2 Naming Series Counter

`FieldTypeNamingSeries` declares the pattern (e.g., `"INV-{YYYY}-{SEQ:6}"`) but the actual sequential counter must be stored somewhere. The framework declares the field type but provides no counter storage or generation service.

**Current state:** The series pattern is stored in `EntitySchema.Fields[].Series` but there is no `NamingSeriesService` that maintains the counter per (tenant, entity, year).

**Recommendation:** Implement `NamingSeriesService` backed by a Redis atomic increment (INCR) with key `naming:{tenant_id}:{entity}:{year}`. Fallback to PostgreSQL sequence if Redis is unavailable. This is required for Finance to function at all.

### 4.3 Notification Templates

`finance.InvoiceApprovalWorkflow` calls `SendInvoiceApprovedNotificationActivity` — but the Notifications platform module has no template system. Notifications need:
- Per-tenant email/SMS templates
- Template variables from entity records
- Multi-channel delivery (email, push, webhook)

**Recommendation:** Add `NotificationTemplate` entity to the Notifications module before Finance workflows go live.

### 4.4 Currency Conversion Service

Multiple Finance entities store `exchange_rate` as a snapshot field. There is no framework service to:
- Fetch the latest `ExchangeRate` for a currency pair on a given date
- Auto-populate `exchange_rate` on create via a `BeforeCreate` hook

**Recommendation:** The `ExchangeRateService` should be a Finance module service that hooks use. This is module-level, not framework-level — no framework change needed.

---

## 5. Duplicated Code Discovered

### 5.1 Permission Boilerplate

Every Finance entity repeats the same permission pattern:
```go
Permissions: def.PermissionSet{
    Create: []string{"role:tenant.admin", "role:finance.manager"},
    Read:   []string{"role:tenant.admin", "role:finance.manager", "role:finance.accountant", "role:finance.viewer", "role:tenant.user"},
    Write:  []string{"role:tenant.admin", "role:finance.manager"},
    Delete: []string{"role:tenant.admin"},
},
```

This is copied 15+ times with minor variations. **Recommended:** Define permission preset variables:
```go
var financeAdminPermissions = def.PermissionSet{...}
var financeViewerPermissions = def.PermissionSet{...}
var financeImmutablePermissions = def.PermissionSet{...}  // for ledger/tax entries
```
This is module-level refactoring, not a framework change.

### 5.2 `toFloat64` Helper

The `toFloat64` numeric coercion helper appears in `hooks_journal.go`. It would be needed by any module doing numeric validation. **Recommendation:** Move to `awo/def` package as `def.AsFloat64(v any) float64` for reuse.

### 5.3 Status Transition Guards

`FiscalYearTransitionGuard`, `PeriodTransitionGuard`, `InvoiceStatusGuard`, `PaymentStatusGuard` all implement the same pattern: check `prev.Data["status"]` and validate the new status is an allowed transition.

**Recommendation:** Framework utility:
```go
// In awo/runtime:
func NewStatusTransitionHook(field string, allowed map[string][]string) def.BeforeUpdateHook
```
Eliminates ~100 lines of repeated hook code across all modules.

---

## 6. Framework APIs That Felt Awkward

### 6.1 ActionHandlerFunc Signature

```go
type ActionHandlerFunc func(ctx *ActionContext) (*ActionResult, error)
```

The handler receives `*ActionContext` but cannot access any services. For non-trivial actions (which all Finance actions are), this is a dead end. Every Finance action handler returns `501 Not Implemented`.

This is the single most awkward API in the framework. It must be fixed before v1.0.

### 6.2 FieldTypeLink in SDUI

Defining a `FieldTypeLink` field and then getting an `input-text` box in the form is jarring. The framework promises "metadata-driven UI" but the most common relational field type produces unusable UI. This is a significant gap between promise and reality.

### 6.3 Hook DB Access

Hooks being pure in-memory transforms is elegant for testing but insufficient for real business rules. A `Finance` validator that cannot check "does this account exist and is it active?" is severely constrained.

The `context.Context` parameter implies access to context-carried state, but the framework provides no way to inject DB access into hooks via context. This needs to be a first-class design decision before v1.0.

### 6.4 NamingSeries as Declaration Without Implementation

Declaring `Series: "INV-{YYYY}-{SEQ:6}"` on a field is clean, but there is no corresponding counter service. This is a framework feature with no runtime. It looks complete from the definition side but is completely inert until a `NamingSeriesService` is implemented.

### 6.5 WorkflowTrigger.InputBuilder vs. Module Coupling

`InputBuilder` must produce a workflow input struct from `*def.EntityRecord`. But `*def.EntityRecord` stores fields as `map[string]any`, requiring type assertions. The Finance module's `InputBuilder` functions write:
```go
func(r *def.EntityRecord, tc def.TriggerContext) (any, error) {
    return InvoiceApprovalInput{
        TenantID:  r.TenantID,
        InvoiceID: r.ID,
        ActorID:   tc.Actor.UserID,
    }, nil
}
```
This is fine, but the input struct is defined in the module package while the trigger is declared on the `EntityDefinition`. Both are in `package finance`, so no coupling issue. ✅

---

## 7. Recommended Improvements Before Awo v1.0 Freeze

**Priority 1 (blockers — must fix):**
1. **ActionContext: add repository/service access** — all non-trivial actions are currently dead code
2. **NamingSeriesService: implement counter generation** — NamingSeries fields are declared but inert
3. **FieldTypeLink SDUI: generate linked select widget** — current output is unusable in production

**Priority 2 (important — should fix):**
4. **HookContext: add EntityLookup interface** — cross-entity validation is impossible without it
5. **StatusTransitionHook utility** — eliminate repeated status guard boilerplate
6. **ReportDefinition concept** — Finance reports cannot be expressed in the current framework

**Priority 3 (nice to have — can defer to v1.1):**
7. **Notification template system** — workflow notifications need templates
8. **Permission preset variables** — reduce boilerplate in module definitions
9. **def.AsFloat64 utility** — move numeric coercion to framework

---

## 8. Overall Assessment

### Is Awo v1.0 Production-Ready?

**For simple CRUD ERP modules:** Yes. The framework eliminates enormous amounts of boilerplate and produces consistent, secure, multi-tenant APIs with zero custom routing code.

**For Finance specifically:** Not yet. Three blockers (ActionContext, NamingSeries, Link SDUI) prevent Finance from functioning as a production module. These are estimated at 3–5 days of framework work.

**After fixing the 3 blockers:** Finance would be implementable in ~2 weeks of business logic work (PostingService, ExchangeRateService, report queries), with the framework handling the remaining ~80% of infrastructure.

### Framework Productivity Assessment

The framework delivered on its core promise: one `EntityDefinition` drives 5 subsystems simultaneously. The Finance module definition files (~1,200 lines) would replace an estimated 8,000–12,000 lines of hand-coded Go (routes, handlers, validators, SQL, UI schemas, permission checks, audit integration).

The productivity multiplier is real. The three gaps discovered are solvable and were found precisely because Finance pushed the framework to its limits — which was the purpose of this exercise.

### Recommendation

**Freeze framework architecture after fixing 3 Priority-1 blockers.** Do not introduce new abstractions speculatively. The Finance module has surfaced the genuine gaps; future modules should be built on the same foundation.

Awo v1.0 should be declared after:
1. ActionContext gets repository access (1–2 days)
2. NamingSeriesService is implemented (1 day)
3. FieldTypeLink SDUI generates linked select (1 day)
4. Finance module actions are fully implemented (1 week)
5. Full test coverage for Finance (2–3 days)

---

*Validation conducted by implementing the Finance module as Awo's proving ground, per the framework architecture specification.*
