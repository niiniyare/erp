> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "Awo vs Frappe Framework — Feature Comparison and Gap Analysis"
section: "custom"
related:
  - "[Chapter 1: Introduction](../part-01-foundations/01-introduction.md)"
  - "[Chapter 2: The EntityDefinition](../part-01-foundations/02-entity-definition.md)"
---

# Awo vs Frappe Framework — Feature Comparison and Gap Analysis

This document maps Frappe Framework concepts to their Awo equivalents and identifies features present in Frappe that Awo does not yet have. It is intended for teams evaluating Awo, teams migrating from Frappe, and contributors deciding what to build next.

---

## Concept Mapping

### Core Entity Model

| Frappe | Awo | Notes |
|--------|-----|-------|
| DocType | `EntityDefinition` | Central primitive in both; Frappe's is stored as data (MariaDB), Awo's is a Go struct registered at compile time for system entities |
| Child DocType | `FieldTypeTable` edge + child EntityDefinition | Inline one-to-many in forms |
| Single DocType | *(planned: singleton entity type)* | One record per tenant, no list view. Awo has `TenantConfig` as a concrete example but no general pattern |
| Virtual DocType | *(not implemented)* | Computed/no-DB entity — see gap §1 |
| DocField | `FieldDef` | Field declaration; Awo has a richer type system |
| DocField.fieldtype "Link" | `FieldTypeLink` | FK to another entity |
| DocField.fieldtype "Dynamic Link" | `FieldTypeDynamicLink` | Polymorphic FK |
| DocField.fieldtype "Attach" | `FieldTypeAttach` / `FieldTypeAttachImage` | File attachment |
| DocField.fieldtype "Geolocation" | *(not implemented)* | See gap §16 |
| DocField.fieldtype "Signature" | *(not implemented)* | See gap §17 |

### Lifecycle Hooks

| Frappe | Awo | Notes |
|--------|-----|-------|
| `validate()` | `BeforeHook(OpCreate\|OpUpdate, ...)` | Pre-persist validation |
| `before_save()` | `BeforeHook(OpCreate\|OpUpdate, ...)` | Business rule enforcement |
| `after_save()` | `AfterHook(OpCreate\|OpUpdate, ...)` | Post-persist side effects |
| `before_submit()` | `BeforeHook(OpSubmit, ...)` *(planned)* | Pre-submit guard |
| `on_submit()` | Temporal workflow trigger on `OpSubmit` | Awo makes submission a durable workflow, not a synchronous hook |
| `on_cancel()` | Compensation workflow | Saga compensations are more robust than synchronous cancel hooks |
| `before_delete()` | `BeforeHook(OpDelete, ...)` | Deletion guard |
| `after_delete()` | `AfterHook(OpDelete, ...)` | Post-deletion cleanup |

### Permissions

| Frappe | Awo | Notes |
|--------|-----|-------|
| DocPerm role matrix | `PolicyDef` on `EntityDefinition` | Frappe: per-role CRUD matrix stored in DB. Awo: Go policy functions, compile-time checked |
| User Permission (record-level restriction per user) | Privacy policies (`AllowWithinOrgScope`, custom `PolicyFunc`) | Awo org-scope model is more structured than Frappe's user permission strings |
| Role Profile | Role aggregation | Frappe bundles roles into profiles; Awo uses `HasRole()` in policy functions |
| Permission Manager UI | *(not implemented)* | See gap §12 |

### Background Processing

| Frappe | Awo | Notes |
|--------|-----|-------|
| Background jobs (RQ/Redis) | Temporal activities | Awo's Temporal activities are durable, retried automatically, and have full execution history |
| Frappe Workflow (state machine) | Temporal workflows with signals | Awo's approach is more powerful: arbitrary Go code, waits for human input, survives process restarts |
| Scheduler (site:schedule_jobs) | Temporal Schedules | Both support cron-like scheduling; Temporal Schedules have better observability and catch-up |

### UI

| Frappe | Awo | Notes |
|--------|-----|-------|
| Jinja forms (JavaScript + Python templates) | amis JSON schema (SDUI) | Awo generates amis JSON from EntityDefinition; no JavaScript required for standard forms |
| Frappe UI (Vue components) | amis built-in components | Frappe's UI is bespoke JS; amis is a declarative component library |
| List view | `AmisList` / `AmisCRUD` | Auto-generated from EntityDefinition fields |
| Form view | `AmisForm` | Auto-generated |
| Report Builder | `AmisReportPage` *(partial)* | See gap §8 |
| Dashboard | `AmisDashboard` | Auto-generated charts and stat cards; no no-code builder |
| Kanban | `AmisKanban` | Available |
| Calendar view | *(not implemented)* | See gap §9 |
| Gantt chart | *(not implemented)* | See gap §10 |
| Tree view | *(not implemented)* | See gap §11 |
| Workspace builder | *(not implemented)* | See gap §14 |
| Print Format designer | HTML/template + chromedp | Awo uses Go templates, not a visual designer. See Chapter 25 |

### Database and Migrations

| Frappe | Awo | Notes |
|--------|-----|-------|
| MariaDB, schema-per-site | PostgreSQL, shared schema + RLS | Awo RLS is operationally simpler than schema-per-site |
| Auto-migrate (`bench migrate`) | `golang-migrate` with reviewed `.up.sql`/`.down.sql` | Frappe applies schema changes automatically; Awo requires human-reviewed migration files |
| Frappe patches (Python data migration scripts) | *(not implemented as a general mechanism)* | Awo uses seed data for provisioning; there is no general "patch" runner for data migrations. See gap §13 |

---

## Features in Frappe Not Yet in Awo

### §1 — Virtual DocType (computed entities with no database table)

Frappe allows DocTypes with `is_virtual = 1`. These have no corresponding database table. All CRUD operations are delegated to controller methods that compute the data on the fly — from external APIs, from computations over other tables, or from in-memory state.

Awo has no equivalent. System entities must map to a database table. The closest workaround is a reporting endpoint that returns computed data as a flat list, but it does not participate in the EntityDefinition lifecycle (no hooks, no SDUI auto-generation).

**Use cases it blocks:** Real-time dashboard widgets that aggregate data across entities without a materialized table; integration entities that proxy data from an external ERP or API without storing it locally; computed views that are too expensive to materialize but need to appear as first-class entities in the UI.

---

### §2 — Singleton EntityDefinition (one record per tenant, not a list)

Frappe's "Single" DocTypes have exactly one record per site. They are used for settings: `System Settings`, `HR Settings`, `Accounts Settings`. The list view is replaced by a form-only view that saves directly to the single record.

Awo has `TenantConfig` as a concrete platform-level singleton but no general mechanism for module developers to declare singleton entities. A module that needs per-tenant settings must either use `TenantConfig` categories (if the framework exposes the right extension point) or create a regular entity and enforce the single-record constraint in a hook.

**Use cases it blocks:** Module-specific settings pages (e.g., "Finance Settings", "HR Settings"); tenant-level feature configuration that does not belong in the platform `TenantConfig`.

---

### §3 — Document-level collaboration (comments, @mentions, assignments, sharing)

Frappe attaches a communication timeline to every document: comments with `@user` mentions, email threads, document assignments to users, and per-document sharing with external users. This is implemented through Frappe's `Communication` and `ToDo` DocTypes plus a timeline component on every form.

Awo has an `AuditLog` for write tracking but nothing for:
- **Comments**: users cannot leave notes on a specific document record.
- **@mentions**: no notification trigger tied to document comments.
- **Assignments**: no mechanism to assign a document to a user and track completion.
- **Document sharing**: no per-document ACL for sharing with specific users beyond role-based access.

**Use cases it blocks:** Approval workflows that require the approver to leave a rejection reason inline on the document; sales pipeline notes; purchase order queries between the procurement team and suppliers; any workflow requiring human communication anchored to a specific record.

---

### §4 — Email integration (email accounts, communication threading, email from document)

Frappe has a full email inbox: configure IMAP/SMTP accounts, route incoming emails to document threads, send emails from documents (invoices, purchase orders, support tickets), and store all email communication in the `Email Queue` and `Communication` DocTypes.

Awo's `Notification` entity covers in-app notifications, email dispatch, and SMS. It does not have:
- Incoming email handling (no IMAP integration, no email-to-document threading).
- Email composition from a document form ("Email this invoice to the customer").
- Email template management with Jinja-equivalent rendering.
- Email queue management and delivery retry UI.

**Use cases it blocks:** Sending invoices directly from the system to customers; incoming supplier invoice acknowledgements tracked against purchase orders; support ticket email threading.

---

### §5 — Outgoing webhooks

Frappe has a `Webhook` DocType that lets administrators configure webhooks: select a DocType, select events (on_submit, on_update, etc.), specify a URL and headers, and Frappe will POST a JSON payload to that URL whenever the event fires.

Awo has no webhook system. Module developers can use `AfterHook` to call an HTTP endpoint, but there is no:
- Admin UI for configuring webhooks.
- Webhook delivery log with retry history.
- Payload template configuration.
- Secret-based HMAC signature for webhook verification.

**Use cases it blocks:** Integrating with external systems (payment gateways, logistics providers, CRM tools) that require real-time push notifications on document events.

---

### §6 — OAuth / Social Login and detailed 2FA

Frappe supports Google, GitHub, and other OAuth providers for login. It also supports TOTP-based 2FA with backup codes.

Awo's auth chapter mentions MFA in the session flow but does not implement or document:
- OAuth 2.0 / OIDC provider integration.
- Google Workspace SSO for enterprise tenants.
- TOTP 2FA implementation (library choice, QR code enrollment, backup codes).
- Passkeys / WebAuthn.

**Use cases it blocks:** Enterprise tenants requiring SSO; reducing password-related support load; compliance requirements for MFA.

---

### §7 — Bulk data import (CSV / Excel)

Frappe has a Data Import tool that accepts CSV or Excel files, maps columns to DocType fields, validates in bulk, previews errors before import, and submits records in batches with a progress indicator.

Awo has `BulkCreate` and `BulkUpdate` on the `EntityStore` interface, but no:
- File upload endpoint for CSV/Excel import.
- Column-to-field mapping UI.
- Bulk validation with row-level error reporting.
- Import progress tracking and partial success handling.

**Use cases it blocks:** Migrating data from a legacy system; importing a customer list from a spreadsheet; bulk stock adjustments from a physical count sheet.

---

### §8 — No-code Report Builder and SQL Query Reports

Frappe has two reporting mechanisms beyond simple list views:

1. **Report Builder**: point-and-click report designer that selects DocType, chooses columns, adds filters, sorts, and groups — no code required.
2. **Script Report (Query Report)**: a Python script that executes arbitrary SQL and returns rows and columns, displayed in a full-featured table with export.

Awo's `AmisReportPage` generates a parameterised report UI from an API endpoint, but:
- There is no no-code report builder. Every report requires a Go endpoint.
- There is no SQL-query-based report type. Complex cross-entity queries must be Go functions.
- There is no report scheduling / email delivery system.

---

### §9 — Calendar View

Frappe's Calendar view displays documents on a calendar based on date/datetime fields, with day/week/month views and drag-to-reschedule.

amis does not have a built-in calendar component that integrates with the EntityDefinition lifecycle. A custom amis renderer could implement this, but there is no built-in path.

---

### §10 — Gantt Chart View

Frappe has a Gantt chart view for project-management DocTypes with start/end dates, task dependencies, and drag-to-resize.

Not available in amis built-ins or Awo framework.

---

### §11 — Tree View

Frappe renders hierarchical DocTypes (Account tree, Employee group tree, Item group tree) as collapsible tree views with indent levels.

amis has a `tree-select` control for picking from a hierarchy, but there is no tree view for displaying entity hierarchies as a navigable list. Materialized-path hierarchies (chart of accounts, department trees) have no dedicated SDUI rendering in Awo.

---

### §12 — Permission Manager UI

Frappe has a Permission Manager page in the System Settings that lets administrators add, remove, and configure DocPerm rows for each role and DocType through a web interface.

Awo's permission system is defined in Go code (`PolicyDef` on `EntityDefinition`). There is no admin UI for inspecting or modifying permissions at runtime. This is a deliberate tradeoff (Go compile-time safety over runtime configurability) but it means:
- Administrators cannot adjust permissions without a code deploy.
- There is no way to inspect effective permissions for a given user + entity combination from the UI.

---

### §13 — Patch Runner (data migration scripts)

Frappe has a patch system: Python scripts placed in a `patches/` directory, each run exactly once (tracked in the `Patch Log` DocType). Patches handle data migrations that cannot be expressed as schema DDL — recomputing derived fields, fixing corrupted data, backfilling denormalized columns, migrating from one data model to another.

Awo has golang-migrate for schema changes and Temporal workflows for provisioning seed data, but no equivalent for one-time data migration scripts that must run once against existing tenant data.

**Use cases it blocks:** Renaming a field's values after a Select field's options change; backfilling a new computed column from existing data; fixing data quality issues introduced by a previous bug.

---

### §14 — Workspace and Homepage Builder

Frappe's Desk has a configurable homepage (Workspace) with shortcut icons, recent documents, charts, and to-do lists. Administrators can build custom workspaces for different roles.

Awo's SDUI serves navigation from `GET /sdui/nav` (derived from EntityDefinition Module groupings) and individual entity CRUD pages. There is no:
- Configurable homepage/workspace per role.
- Shortcut management.
- Recent documents widget.
- Role-specific landing page.

---

### §15 — Module Onboarding Flows

Frappe has an onboarding module: when a user first enables a module (e.g., Accounts), they see a guided checklist of setup steps ("Create your Company", "Set up Chart of Accounts", "Open the first fiscal year"). Each step links to the relevant setup form.

Awo has no onboarding framework. New tenant setup is handled by the provisioning workflow which seeds data, but there is no guided UI flow for the tenant administrator to complete setup interactively.

---

### §16 — Geolocation Field Type

Frappe has a `Geolocation` field type that stores GeoJSON and renders as a Leaflet map with point/polygon editing in the form.

Awo's `FieldTypeJSON` can store GeoJSON but there is no corresponding amis control for map display or geometry editing. Geolocation-dependent entities (site locations, delivery addresses with coordinates, meter reading locations) cannot display maps without a custom amis renderer.

---

### §17 — Signature Field Type

Frappe has a `Signature` field type that renders a canvas signature pad in the form and stores the signature as a base64 PNG.

Not available in Awo. Required for: delivery confirmation signatures, timesheet signatures, customer approval signatures.

---

### §18 — Auto-Repeat (Recurring Document Generation)

Frappe has an Auto Repeat feature: attach a schedule to any submittable document and Frappe will automatically create a new draft copy on the specified schedule (daily, weekly, monthly, on specific dates). Commonly used for recurring invoices, standing purchase orders, and scheduled journal entries.

Awo can implement this with a Temporal Schedule that starts a workflow to copy a document, but there is no general framework mechanism. Each use case requires a custom workflow. There is no admin UI for configuring auto-repeat on arbitrary entities.

---

### §19 — Document Versioning (full history with diff)

Frappe stores full document snapshots in a `Version` DocType on every update if `track_changes = 1` is set. Users can browse version history and see a field-by-field diff between versions.

Awo's `AuditLog` records who changed what field at what time (a diff), but does not store full document snapshots. You can reconstruct history from the audit log but you cannot restore a previous version from the UI.

---

### §20 — Letter Head Templates

Frappe has `Letter Head` DocType: upload a company logo and define a top/bottom banner HTML. Letter heads are applied to printed documents (invoices, purchase orders, quotations) to match the company's branding.

Awo's `PrintTemplate` (Chapter 25) uses Go HTML templates per document type. There is no general letter head system that non-developers can configure. Tenant branding on printed documents requires a developer to update the HTML template.

---

### §21 — i18n / Translation Management

Frappe has a built-in translation framework: mark strings with `_("string")` in Python and JS, export them to CSV, have translators fill in target language values, import back. The Translation DocType in the admin panel lets admins add/override specific translations for their site.

Awo mentions Swahili support for the Kenyan market in the error message localisation section (§18.4) but has no:
- General i18n string extraction framework.
- Translation management UI.
- Per-tenant translation overrides.
- amis component label translation.

---

### §22 — API Key Management

Frappe lets users generate API key/secret pairs from their user profile. API calls use `Authorization: token {api_key}:{api_secret}` instead of a session cookie. This allows programmatic access without maintaining a session.

Awo supports JWT for mobile/API clients (§15.1) but has no:
- API key generation UI for end users.
- API key management (list, revoke, set expiry).
- API key scoping (limit a key to specific entities or operations).

---

## Features Awo Has That Frappe Does Not

For balance, Awo provides capabilities that Frappe lacks or implements poorly:

| Feature | Awo | Frappe |
|---------|-----|--------|
| Type-safe entity declarations | Go compiler catches field reference errors | Python; type errors are runtime errors |
| Durable workflow engine | Temporal — survives process restarts, human-in-the-loop | RQ jobs — no durability across restarts; workflow state machine is simple |
| Row-Level Security | Database-enforced tenant isolation; cannot be bypassed by a forgotten WHERE clause | Application-level filtering; a bug in Python code can leak cross-tenant data |
| Organisational hierarchy in data model | `OrgScope` (Global/Tenant/Company/Division) is a first-class concept | Site = company; no built-in multi-company hierarchy |
| Reviewed migrations | `golang-migrate` with explicit `.up.sql`/`.down.sql` reviewed by humans | Auto-migrate applies schema changes without human review |
| Performance | Go binary; handles hundreds of concurrent requests per instance | Python (Gunicorn/Gevent); CPU-bound tasks are significantly slower |
| Deployment simplicity | Single static Go binary | Python environment, multiple daemons (gunicorn, redis-queue workers, scheduler), bench tool |
| Server-driven UI | amis JSON from the API; UI updates without frontend deploys | Jinja + Vue; frontend and backend must be updated together |
| Test infrastructure | Standard Go testing, `testsuite.WorkflowTestSuite` for workflows | Python unittest; Frappe test fixtures are complex to set up |

---

## Priority Gaps for the Roadmap

Based on the above, the highest-impact gaps for a Kenya ERP deployment are:

1. **§4 Email integration** — invoicing workflows require emailing documents to customers directly from the system.
2. **§7 Bulk data import** — every new tenant onboarding requires importing historical data from spreadsheets.
3. **§3 Document comments** — approval workflows need inline rejection reasons; teams need notes on documents.
4. **§5 Outgoing webhooks** — integrations with M-Pesa, KRA eTIMS callbacks, and third-party systems need push delivery.
5. **§6 2FA/TOTP** — required for compliance and enterprise tenants.
6. **§2 Singleton entity type** — Finance Settings, HR Settings per module need a clean pattern.
7. **§13 Patch runner** — data migration scripts are needed as the system evolves.
8. **§8 Query reports** — finance and operations teams need ad-hoc SQL reporting without Go deployments.
9. **§20 Letter head** — printed documents need tenant branding without developer involvement.
10. **§21 i18n** — Swahili support must be systematic, not ad-hoc.
