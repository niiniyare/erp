# AwoERP — Implementation Task Tracker

> **Chronological, dependency-ordered implementation plan.**
> Check each `[ ]` after verified, tested implementation.
> IDs are stable — never renumber. Prefix key: `PLAT` Platform Core · `FIN` Finance · `CRM` CRM ·
> `SAL` Sales · `PRO` Procurement · `INV` Inventory · `WHS` Warehouse · `HR` Human Resources ·
> `PAY` Payroll · `MFG` Manufacturing · `PRJ` Projects · `SVC` Service · `AST` Assets ·
> `TRV` Travel · `PROP` Property · `FOR` Forecourt · `AIR` Airline · `BI` Business Intelligence ·
> `SDUI` Server-Driven UI Platform

---

## 🏗️ PART A — PLATFORM CORE

---

### § 1 — Platform Foundation

---

- [ ] **PLAT-001** · Tenant Lifecycle Engine
  `platform → tenant-management → provisioning-state-machine`

  Implement the foundational tenant state machine: draft → provisioning → active →
  suspended → terminated → archived. Includes RLS schema isolation, `set_config`-based
  tenant context injection via `store.WithTenant()`, and per-transition audit entries.
  Done when all states are reachable via API with verified data isolation between tenants,
  each transition is audited, and the provisioning flow seeds a default org structure stub.

  - **Depends on:** []

---

- [ ] **PLAT-002** · Organizational Hierarchy
  `platform → tenant-management → org-hierarchy`

  Implement Companies, Departments, Divisions, Branches, Regions, Cost Centers, and
  Business Units as hierarchical tenant-scoped entities. Each node carries its own
  configuration scope and forms the boundary for budget ownership, permission delegation,
  and reporting segmentation. Done when the full org tree is traversable via API, parent-
  child relationships are enforced, and each node type can be created, updated, and
  deactivated independently.

  - **Depends on:** [PLAT-001]

---

- [ ] **PLAT-003** · User Identity Core
  `platform → iam → user-management`

  Implement user creation, profile data model, invitation flow, and status lifecycle
  (pending → active → suspended → locked). Users are tenant-scoped; user identity is
  separate from authentication credentials. Done when users can be created manually and
  via invitation, deactivated with data-ownership transfer, and their tenant membership
  is queryable with correct RLS boundaries.

  - **Depends on:** [PLAT-001]

---

- [ ] **PLAT-004** · Role & Permission Engine
  `platform → iam → rbac-core`

  Implement the RBAC foundation: role definitions, permission primitives
  (`<module>.<resource>.<action>` format), role-to-permission assignment, role-to-user
  assignment, and effective permission resolution. Includes system-provided role templates
  (Finance Manager, HR Clerk, Admin, Read-Only). Done when a user's effective permissions
  can be computed at login, stored in session, and enforced on every route via the
  `Authorize()` middleware without per-request DB lookup.

  - **Depends on:** [PLAT-003]

---

- [ ] **PLAT-005** · Authentication & Session Management
  `platform → iam → authentication`

  Implement username/password authentication, password hashing (Argon2id), JWT session
  issuance, session storage in Redis, session revocation, and concurrent session limits.
  Includes password complexity policy, breach detection hook, and account lockout after
  failed attempts. Done when a user can log in, receive a valid JWT, and be correctly
  rejected after policy violations or session expiry.

  - **Depends on:** [PLAT-003, PLAT-004]

---

- [ ] **PLAT-006** · Multi-Factor Authentication
  `platform → iam → mfa`

  Implement TOTP-based MFA (RFC 6238) with QR code provisioning, backup codes, MFA
  enforcement policies per tenant, and MFA bypass for trusted devices. Done when a tenant
  can mandate MFA, users can enrol and authenticate via TOTP, and backup codes are
  single-use with exhaustion alerting.

  - **Depends on:** [PLAT-005]

---

- [ ] **PLAT-007** · Single Sign-On (SSO)
  `platform → iam → sso`

  Implement SAML 2.0 and OIDC-based SSO. Includes IdP metadata exchange, attribute
  mapping (email → user profile), JIT provisioning of new users on first SSO login,
  and tenant-level SSO policy (SSO-required, SSO-optional). Done when a tenant can
  configure an external IdP, users can authenticate via SSO, and JIT-provisioned users
  receive correct default roles.

  - **Depends on:** [PLAT-005, PLAT-004]

---

- [ ] **PLAT-008** · Audit Log Engine
  `platform → audit → immutable-event-log`

  Implement the immutable audit event store: entity, action, actor, tenant, before/after
  state, timestamp, and IP/session metadata. Supports user activity, system activity, and
  security events. Done when every CRUD operation platform-wide emits an audit event,
  events are queryable by tenant with RLS isolation, and the log is append-only with no
  update or delete path.

  - **Depends on:** [PLAT-001, PLAT-003]

---

- [ ] **PLAT-009** · Feature Flag Management
  `platform → feature-flags → flag-lifecycle`

  Implement boolean and variant feature flags with tenant, user, role, and percentage-
  based targeting. Includes flag lifecycle (draft → active → retired), flag evaluation
  engine with Redis caching, and kill-switch capability. Done when a flag can be created,
  targeted to a specific tenant or user segment, evaluated in under 5ms via cache, and
  toggled without service restart.

  - **Depends on:** [PLAT-001, PLAT-003]

---

- [ ] **PLAT-010** · Configuration Management
  `platform → configuration → layered-config`

  Implement the three-layer configuration hierarchy: platform defaults → tenant config →
  branch/department config → user preferences. Includes runtime config updates without
  restart, config validation rules, and override conflict resolution. Done when a config
  value can be set at any layer, lower layers correctly override higher ones, and invalid
  config values are rejected with a structured error before persistence.

  - **Depends on:** [PLAT-002, PLAT-009]

---

- [ ] **PLAT-011** · Notification Framework
  `platform → notifications → channel-routing`

  Implement the notification abstraction layer with in-app, email, and SMS channels.
  Includes template engine (variable substitution, per-locale variants), channel routing
  rules, deduplication, and user opt-in preferences. Done when a notification event can
  be emitted from any domain service, routed to the correct channel(s) per user preference,
  and delivered with correct locale and tenant branding.

  - **Depends on:** [PLAT-003, PLAT-010]

---

- [ ] **PLAT-012** · Localization & Branding
  `platform → configuration → localization-and-branding`

  Implement per-tenant locale settings: language, date/number/currency formats, timezone,
  and calendar type. Add tenant branding: logo upload, primary colour palette, email
  header/footer, and white-label domain configuration. Done when two tenants can operate
  with completely different locales and all system-generated documents (emails, reports,
  PDFs) reflect the correct tenant branding.

  - **Depends on:** [PLAT-010]

---

- [ ] **PLAT-013** · Document & File Management
  `platform → documents → file-storage`

  Implement file upload (chunked, virus-scanned), folder organisation, document versioning,
  metadata tagging, and access control (view/download/edit/delete/share permissions). Includes
  tenant storage quotas, expiring share links, and retention policy hooks. Done when any
  entity in the system can have documents attached, versioned, and accessed only by
  authorised users within the same tenant.

  - **Depends on:** [PLAT-001, PLAT-004, PLAT-008]

---

- [ ] **PLAT-014** · Workflow Engine
  `platform → workflow → temporal-state-machine`

  Implement the Temporal-backed workflow engine: workflow definition format (states,
  transitions, conditions, actions), workflow execution start/signal/query, human task
  assignment and completion, and workflow audit trail. Done when a business domain can
  define a multi-step approval workflow in YAML/JSON, execute it via Temporal, and have
  every state transition recorded in the audit log.

  - **Depends on:** [PLAT-003, PLAT-004, PLAT-008]

---

- [ ] **PLAT-015** · Approval Management
  `platform → approvals → approval-policy-engine`

  Implement approval policy definitions: single, sequential, parallel, and majority-vote
  chain types. Includes threshold-based routing (e.g. amount > threshold → CFO approval),
  dynamic approver resolution at runtime, temporary delegation, and escalation on timeout.
  Done when any domain entity can be submitted for approval, routes correctly through the
  configured chain, and approval decisions are immutably recorded.

  - **Depends on:** [PLAT-014, PLAT-011]

---

- [ ] **PLAT-016** · Reporting Platform
  `platform → reporting → report-engine`

  Implement the report engine: report type registry, parameter-driven report execution,
  PDF/Excel/CSV export, scheduling (cron-based), and distribution (email/in-app). Includes
  tenant-scoped report templates and custom report builder hooks for business modules. Done
  when a parameterised report can be defined, executed on demand, exported in all three
  formats, and scheduled for automated delivery.

  - **Depends on:** [PLAT-001, PLAT-004, PLAT-012]

---

- [ ] **PLAT-017** · Integration Platform
  `platform → integration → api-gateway-and-webhooks`

  Implement outbound webhook delivery (event subscription, payload signing, retry with
  exponential backoff, delivery log), inbound API key management, and the integration
  connector registry. Done when a tenant can subscribe to platform events via webhooks,
  receive signed payloads, and replay failed deliveries from the delivery log.

  - **Depends on:** [PLAT-001, PLAT-005, PLAT-008]

---

- [ ] **PLAT-018** · Search Platform
  `platform → search → full-text-and-faceted-search`

  Implement full-text search with tenant-scoped indexes, faceted filtering, saved searches,
  and search analytics. Includes per-entity search indexing hooks so any module can register
  its entities as searchable. Done when a user can search across all their tenant's data,
  filter by facet, save a search query, and results respect all RLS boundaries.

  - **Depends on:** [PLAT-001, PLAT-004]

---

- [ ] **PLAT-019** · Platform Operations & Observability
  `platform → operations → health-and-runbooks`

  Implement health check endpoints, Prometheus metrics endpoints, structured log shipping,
  OpenTelemetry trace export, background job monitoring, and cache/storage management
  dashboards. Includes backup and recovery runbooks and disaster recovery playbooks. Done
  when the full observability stack emits traces, metrics, and logs for every request and
  background job, and on-call runbooks are validated against a live environment.

  - **Depends on:** [PLAT-001, PLAT-008]

---

- [ ] **PLAT-020** · Developer SDK & API Reference
  `platform → developer-resources → sdk-and-api-docs`

  Publish the Go SDK for third-party integrators, OpenAPI 3.1 spec auto-generated from
  Fiber routes, Postman collection, webhook event catalog, and extension point registry.
  Includes changelog, migration guides per breaking version, and a sandbox tenant for
  API exploration. Done when a third-party developer can authenticate, call every
  documented endpoint, and receive correct responses using only the published SDK and docs.

  - **Depends on:** [PLAT-001 → PLAT-019]

---

---

## 💰 PART B — BUSINESS MODULES

---

### Module 1 — Finance & Accounting

---

- [ ] **FIN-001** · Chart of Accounts
  `finance → general-ledger → chart-of-accounts`

  Implement the account hierarchy (Asset, Liability, Equity, Revenue, Expense, Statistical),
  account numbering conventions, parent/child rollup structure, account segments, and posting
  restrictions. Includes account creation governance (naming policy, mandatory cost centre,
  approval requirement), deactivation with outstanding-balance checks, and account merging
  with balance reclassification. Done when a full multi-level CoA can be seeded per tenant,
  accounts are queryable by type/group, and governance rules block invalid states.

  - **Depends on:** [PLAT-001, PLAT-002, PLAT-004, PLAT-008]

---

- [ ] **FIN-002** · Financial Period Management
  `finance → general-ledger → period-management`

  Implement financial period definitions (monthly, quarterly, annual), period open/close
  state machine, period-lock enforcement at the DB level to prevent backdated postings,
  and year-end carry-forward process. Done when no journal entry can be posted to a locked
  period, period transitions are audited, and the year-end process rolls retained earnings
  correctly.

  - **Depends on:** [FIN-001]

---

- [ ] **FIN-003** · Journal Entry Engine
  `finance → general-ledger → journal-entries`

  Implement the double-entry journal engine: journal types (standard, adjusting, reversing,
  recurring, statistical), balanced entry validation (sum debits = sum credits), posting
  pipeline with period check and budget check, and auto-reversal scheduling. Includes
  multi-currency journal lines with FX rate capture. Done when an unbalanced journal is
  rejected at the DB constraint level, all journal types post correctly, and reversals
  are linked back to their source entry.

  - **Depends on:** [FIN-001, FIN-002, PLAT-015]

---

- [ ] **FIN-004** · Multi-Currency & FX Management
  `finance → general-ledger → multi-currency`

  Implement base currency and functional currency per company, daily FX rate ingestion
  (manual and API-sourced), realised and unrealised FX gain/loss calculation, and
  currency revaluation at period-end. Done when transactions in a foreign currency
  post the correct FX gain/loss to the designated GL account and period-end revaluation
  produces a balanced journal entry.

  - **Depends on:** [FIN-003]

---

- [ ] **FIN-005** · Accounts Receivable
  `finance → accounts-receivable → invoice-and-collections`

  Implement customer ledger structure, sales invoice creation (line items, tax calculation,
  PDF generation), credit notes, payment recording with FIFO allocation, customer aging
  (configurable buckets), dunning letter workflow, and bad debt write-off with approval.
  Done when the full AR cycle from invoice to zero-balance is traceable, aging buckets
  are accurate at any point in time, and write-offs require an authorised approval.

  - **Depends on:** [FIN-003, FIN-004, PLAT-013, PLAT-015]

---

- [ ] **FIN-006** · Accounts Payable
  `finance → accounts-payable → bills-and-payments`

  Implement supplier ledger, supplier bill creation, 3-way match hook (PO → GRN →
  Invoice to be wired once Procurement exists), payment voucher creation, payment run
  (batch approval, bank file generation), and supplier aging. Done when the full AP
  cycle from bill entry to payment is auditable, aging is accurate, and payment runs
  produce a correctly formatted bank payment file.

  - **Depends on:** [FIN-003, FIN-004, PLAT-015]

---

- [ ] **FIN-007** · Cash & Bank Management
  `finance → cash-management → bank-reconciliation`

  Implement bank account master records, deposit/withdrawal/transfer transactions,
  bank statement import (CSV/OFX/MT940), auto-matching engine (reference, amount,
  date fuzzy matching), manual match, and reconciliation close. Done when a bank
  statement can be imported, at least 80% of lines auto-matched by the engine, and
  the reconciliation produces a zero-variance report when books and bank agree.

  - **Depends on:** [FIN-003, FIN-004]

---

- [ ] **FIN-008** · Budget Management
  `finance → budgeting → budget-control`

  Implement annual budget creation (top-down and bottom-up), department budget submission
  and consolidation, budget-to-actual monitoring with real-time commitment tracking, budget
  transfer workflow, and supplementary budget approval. Includes budget check integration
  into journal posting (warn or hard-block on over-budget). Done when a budget cycle can
  be opened, departmental inputs consolidated, and journal entries are blocked or warned
  when they would exceed approved budget.

  - **Depends on:** [FIN-003, FIN-002, PLAT-015]

---

- [ ] **FIN-009** · Fixed Assets
  `finance → fixed-assets → asset-lifecycle`

  Implement asset category hierarchy with default depreciation methods, asset master record,
  acquisition journal (purchase/donation/CIP capitalisation), depreciation run (straight-
  line, declining balance, units-of-production), revaluation, inter-entity transfer, and
  disposal with gain/loss calculation. Done when a full asset lifecycle from acquisition
  to disposal produces correct GL journals at every step and the depreciation schedule
  matches the expected amortisation table.

  - **Depends on:** [FIN-003, FIN-002]

---

- [ ] **FIN-010** · Tax Management
  `finance → tax → tax-calculation-and-compliance`

  Implement tax type configuration (VAT, withholding, excise), rate schedules with effective
  dating, tax rules (taxable/exempt/zero-rated/reverse charge), automatic tax calculation
  on AR/AP transactions, and periodic tax return report generation. Done when every taxable
  transaction calculates the correct tax amount per jurisdiction rules, and the VAT return
  report reconciles to the underlying AR/AP transactions.

  - **Depends on:** [FIN-005, FIN-006]

---

- [ ] **FIN-011** · Financial Reporting
  `finance → reporting → core-financial-statements`

  Implement Trial Balance, General Ledger detail/summary, Balance Sheet (IFRS/GAAP),
  Income Statement (by function and by nature), Cash Flow Statement (direct and indirect
  methods), and comparative period reporting. Includes departmental segmentation and
  profitability drill-down to transaction level. Done when all five core statements
  balance mathematically, comparative columns are accurate, and drill-down from a
  summary line reaches the originating journal entries.

  - **Depends on:** [FIN-003, FIN-004, FIN-008, PLAT-016]

---

- [ ] **FIN-012** · Multi-Company & Consolidation
  `finance → general-ledger → multi-company-consolidation`

  Implement intercompany transaction recording, intercompany elimination entries,
  consolidation group configuration, and consolidated financial statements. Done when
  two companies within a tenant can record intercompany transactions, the consolidation
  process eliminates them, and the consolidated Balance Sheet and Income Statement are
  materially correct.

  - **Depends on:** [FIN-011]

---

---

### Module 2 — Customer Relationship Management (CRM)

---

- [ ] **CRM-001** · Customer & Contact Management
  `crm → customer-management → account-and-contact`

  Implement customer account master record (360° view, credit limit, payment terms),
  contact management (multiple contacts per account with role/department mapping),
  and relationship mapping (subsidiaries, referral networks). Done when a customer
  account shows all linked contacts, financial summary, and relationship hierarchy
  in a single API response.

  - **Depends on:** [PLAT-001, PLAT-002, PLAT-004, PLAT-008]

---

- [ ] **CRM-002** · Lead Management
  `crm → leads → capture-qualify-convert`

  Implement lead capture (manual, import, web form hook), source attribution, lead
  assignment rules (round-robin, territory, skill-based), qualification scoring model,
  MQL threshold configuration, and lead conversion to Opportunity + Account + Contact.
  Done when a lead moves from raw capture through qualification to conversion with
  full lineage preserved.

  - **Depends on:** [CRM-001]

---

- [ ] **CRM-003** · Opportunity & Pipeline Management
  `crm → opportunities → pipeline-and-forecasting`

  Implement sales pipeline with configurable stages, stage probability, opportunity
  creation from lead conversion or manual entry, weighted revenue forecast, and win/loss
  analysis with reason capture. Done when the pipeline report reflects accurate weighted
  values, stage transitions are audited, and win/loss data feeds into conversion rate
  analytics.

  - **Depends on:** [CRM-002]

---

- [ ] **CRM-004** · Activity Management
  `crm → activities → tasks-calls-meetings-notes`

  Implement Tasks (assignment, due date, completion), Call logs (outcome, follow-up),
  Meeting records (attendees, agenda, minutes), and rich-text Notes — all linkable to
  any CRM entity (Lead, Opportunity, Customer, Contact). Done when the activity timeline
  on any CRM entity shows a chronological, complete interaction history.

  - **Depends on:** [CRM-001]

---

- [ ] **CRM-005** · Campaign Management
  `crm → campaigns → email-campaigns-and-roi`

  Implement campaign definition (type, budget, target audience segment), email campaign
  execution (list management, template selection, send scheduling), and campaign tracking
  (open rate, click rate, lead conversion attribution, ROI calculation). Done when a
  campaign can be executed end-to-end and the ROI report shows revenue attributable to
  the campaign's converted leads.

  - **Depends on:** [CRM-002, PLAT-011]

---

- [ ] **CRM-006** · CRM Reporting
  `crm → reporting → pipeline-and-conversion-reports`

  Implement Lead Volume and Conversion Funnel, Opportunity Pipeline (weighted value,
  stage distribution, rep performance), Won/Lost analysis, and Campaign ROI reports.
  Done when each report is filterable by date range, owner, and segment, and numbers
  reconcile to underlying CRM records.

  - **Depends on:** [CRM-003, CRM-005, PLAT-016]

---

---

### Module 3 — Sales Management

---

- [ ] **SAL-001** · Quotation Management
  `sales → quotations → create-price-approve`

  Implement quotation creation from opportunity or from scratch, pricing rule application
  (volume, customer-specific, promotional discounts), multi-currency support, validity
  period, and approval workflow for non-standard pricing. Done when a quotation can be
  created, priced with all applicable rules, submitted for approval when discounts exceed
  threshold, and converted to a Sales Order.

  - **Depends on:** [CRM-003, PLAT-015]

---

- [ ] **SAL-002** · Sales Order Management
  `sales → orders → order-lifecycle`

  Implement sales order creation (from quotation or manual), order amendment workflow,
  partial fulfilment tracking, back-order management, and order cancellation with
  reversal of commitments. Done when an order moves from confirmed through partial
  shipments to fully fulfilled, each shipment updates the outstanding quantity, and
  cancellations release reserved inventory.

  - **Depends on:** [SAL-001]

---

- [ ] **SAL-003** · Sales Contracts
  `sales → contracts → terms-and-renewal`

  Implement contract creation with commercial terms, contract-based pricing enforcement
  (overriding standard price lists when a valid contract exists), renewal workflow with
  advance notification, and contract expiration alerting. Done when an active contract
  automatically enforces its pricing on new sales orders and the renewal workflow fires
  a configurable number of days before expiry.

  - **Depends on:** [SAL-002, PLAT-014]

---

- [ ] **SAL-004** · Price List Management
  `sales → pricing → price-lists-and-promotions`

  Implement multiple price lists per tenant (standard, customer-specific, channel-specific),
  time-bounded promotional pricing, multi-currency price lists, and price list assignment
  rules per customer/segment. Done when the pricing engine selects the correct price list
  for a given customer and date, and promotional prices automatically expire.

  - **Depends on:** [SAL-001]

---

- [ ] **SAL-005** · Sales Analytics
  `sales → reporting → rep-and-revenue-analytics`

  Implement rep performance dashboards (quota vs. attainment), revenue by product/
  customer/region, conversion rate from quotation to order, and order fulfilment
  rate. Done when all metrics reconcile to underlying order records and are filterable
  by rep, region, product, and period.

  - **Depends on:** [SAL-002, SAL-004, PLAT-016]

---

---

### Module 4 — Procurement Management

---

- [ ] **PRO-001** · Supplier Management
  `procurement → suppliers → registration-and-qualification`

  Implement supplier master record, onboarding workflow (document collection, qualification
  criteria checklist, approval), supplier classification (strategic/approved/preferred),
  and performance score model. Done when a supplier moves from unregistered through
  qualification to approved status, the approval is audited, and performance scores
  are calculable from transaction history.

  - **Depends on:** [PLAT-001, PLAT-004, PLAT-008, PLAT-015]

---

- [ ] **PRO-002** · Purchase Requisitions
  `procurement → requisitions → request-approve-release`

  Implement internal purchase request creation with requestor spending limits, item/
  service specification, multi-level approval routing (by amount threshold and department),
  and budget check before approval. Done when a requisition above a requestor's limit is
  auto-routed to the correct approver, budget check blocks approval when over-budget, and
  approved requisitions are releasable to RFQ or direct PO.

  - **Depends on:** [PRO-001, FIN-008, PLAT-015]

---

- [ ] **PRO-003** · RFQ Process
  `procurement → rfq → quote-comparison-and-award`

  Implement RFQ creation from requisition, supplier invitation (email with secure
  submission link), quote submission capture, automated comparison matrix (price,
  delivery, terms), and award decision workflow. Done when multiple suppliers can submit
  quotes against a single RFQ, the comparison matrix ranks them, and the award decision
  is audited with justification captured.

  - **Depends on:** [PRO-002, PRO-001, PLAT-011]

---

- [ ] **PRO-004** · Purchase Orders
  `procurement → purchase-orders → create-amend-receive`

  Implement PO creation from awarded RFQ or direct from requisition, PO amendment
  workflow (version history, supplier acknowledgement), PO cancellation, and goods
  receipt integration trigger. Done when a PO is issued to a supplier, amendments
  create a new version, the supplier acknowledgement is recorded, and receipt against
  the PO is possible once acknowledged.

  - **Depends on:** [PRO-003]

---

- [ ] **PRO-005** · Goods Receipt & Quality Inspection
  `procurement → goods-receipt → receive-inspect-confirm`

  Implement goods receipt note (GRN) creation against a PO, partial receipt handling,
  over-receipt policy enforcement, and quality inspection trigger. Done when a GRN records
  actual received quantities, partial receipts keep the PO open for the remaining balance,
  and the quality inspection workflow is optionally triggered per item category.

  - **Depends on:** [PRO-004, PLAT-014]

---

- [ ] **PRO-006** · 3-Way Invoice Matching
  `procurement → accounts-payable → three-way-match`

  Implement PO → GRN → Supplier Invoice matching with configurable tolerance rules
  (price variance %, quantity variance %), auto-approval for clean matches, and exception
  queue for mismatches. Done when a supplier invoice that matches PO and GRN within
  tolerance is auto-approved and posted to AP without manual intervention.

  - **Depends on:** [PRO-005, FIN-006]

---

- [ ] **PRO-007** · Procurement Analytics
  `procurement → reporting → spend-and-supplier-analytics`

  Implement spend analytics by category, supplier, and department; savings tracking
  against benchmark price; supplier performance KPIs (on-time delivery, quality pass
  rate, invoice accuracy); and procurement compliance metrics (PO coverage, contract
  utilisation). Done when the spend cube reconciles to underlying PO and invoice data
  and supplier scorecards are updated automatically.

  - **Depends on:** [PRO-006, PLAT-016]

---

---

### Module 5 — Inventory Management

---

- [ ] **INV-001** · Product Catalogue
  `inventory → products → catalogue-and-variants`

  Implement product master records with variants (size, colour, grade), unit of measure
  configuration and conversion rules, product categorisation hierarchy, and barcode/SKU
  management. Done when a product with three variants can be created, each variant has
  a unique SKU, UOM conversions are accurate, and the catalogue is searchable by category
  and barcode.

  - **Depends on:** [PLAT-001, PLAT-004, PLAT-008]

---

- [ ] **INV-002** · Stock Movements
  `inventory → stock → receipts-issues-transfers`

  Implement stock receipt (from GRN), stock issue (to production/sales), inter-location
  transfer, and adjustment journal with mandatory reason code. All movements update the
  perpetual stock ledger in real time. Done when every movement type creates a ledger
  entry, on-hand balance is accurate after each movement, and adjustments require an
  approved reason code.

  - **Depends on:** [INV-001, PRO-005]

---

- [ ] **INV-003** · Inventory Valuation
  `inventory → valuation → costing-methods`

  Implement FIFO, weighted average cost, and standard cost valuation methods per product
  category. Includes landed cost allocation across receipts and revaluation workflow when
  standard cost changes. Done when closing stock value reconciles to the sum of all
  receipts minus issues at the configured cost method, and a landed cost revaluation
  produces a balanced GL journal.

  - **Depends on:** [INV-002, FIN-003]

---

- [ ] **INV-004** · Batch & Serial Number Tracking
  `inventory → traceability → batch-and-serial`

  Implement batch creation with attributes (expiry date, manufacturer, lot number),
  serial number assignment, batch/serial selection on stock movements, and end-to-end
  traceability report (where used, where received from). Done when a batch can be traced
  from supplier receipt through every internal movement to customer delivery in a single
  traceability query.

  - **Depends on:** [INV-002]

---

- [ ] **INV-005** · Inventory Planning & Replenishment
  `inventory → planning → reorder-and-safety-stock`

  Implement reorder point calculation (demand-driven), safety stock formula, replenishment
  suggestion engine (generates draft purchase requisitions), and demand-driven parameters
  per location. Done when the planning engine generates accurate replenishment suggestions
  that, when approved, create requisitions preventing stockout based on historical demand.

  - **Depends on:** [INV-002, PRO-002]

---

- [ ] **INV-006** · Physical Inventory Count
  `inventory → cycle-counts → blind-count-and-variance`

  Implement cycle count scheduling, blind count process (counts recorded without showing
  system qty), variance calculation, variance approval workflow, and perpetual inventory
  adjustment posting. Done when a count sheet is generated blind, variances are calculated
  correctly on count entry, and approved variances post a balanced adjustment journal.

  - **Depends on:** [INV-003, PLAT-015]

---

- [ ] **INV-007** · Inventory Reporting
  `inventory → reporting → stock-ledger-and-valuation-reports`

  Implement stock ledger by product/location/batch, inventory valuation report, lot ageing
  analysis, and ABC classification analysis. Done when reports reconcile to perpetual ledger
  balances and ABC analysis correctly segments products by consumption value.

  - **Depends on:** [INV-003, INV-004, PLAT-016]

---

---

### Module 6 — Warehouse Management

---

- [ ] **WHS-001** · Warehouse Structure
  `warehouse → setup → hierarchy-and-bin-configuration`

  Implement the warehouse hierarchy: warehouse → zone → aisle → bay → bin. Each bin
  has type (bulk, rack, cold), capacity (weight and volume), and status (active, blocked,
  quarantine). Done when bin capacity enforcement prevents over-allocation and the full
  hierarchy is navigable via API.

  - **Depends on:** [INV-001, PLAT-002]

---

- [ ] **WHS-002** · Inbound Operations
  `warehouse → inbound → receive-and-putaway`

  Implement receiving dock management, putaway strategy engine (fixed location, dynamic
  nearest-available, zone-directed), and quality inspection workflow on inbound receipts.
  Done when an inbound shipment is received at a dock, the putaway engine directs stock
  to a bin within capacity, and QC inspection is triggered per configured item category.

  - **Depends on:** [WHS-001, INV-002]

---

- [ ] **WHS-003** · Outbound Operations
  `warehouse → outbound → pick-pack-dispatch`

  Implement pick list generation strategies (wave, zone, batch picking), packing workflow
  (carton assignment, weight capture), dispatch confirmation, and shipping document
  generation (packing list, delivery note). Done when a sales order triggers a pick
  list, picker completes it on mobile, packing is confirmed, and dispatch closes the
  outbound shipment.

  - **Depends on:** [WHS-002, SAL-002]

---

- [ ] **WHS-004** · Internal Transfers
  `warehouse → transfers → bin-to-bin-and-inter-warehouse`

  Implement bin-to-bin transfer (same warehouse), inter-warehouse transfer order workflow
  (request → approve → ship → receive), and transfer discrepancy handling. Done when
  stock moves between bins update location records instantly and inter-warehouse transfers
  maintain in-transit stock visibility.

  - **Depends on:** [WHS-001, INV-002]

---

- [ ] **WHS-005** · Warehouse Analytics
  `warehouse → reporting → utilisation-and-throughput`

  Implement utilisation by zone (% capacity used), throughput analysis (lines per hour),
  pick accuracy rate, and dock turnaround time. Done when metrics are calculable from
  warehouse operation records and update on a configurable refresh cadence.

  - **Depends on:** [WHS-003, WHS-004, PLAT-016]

---

---

### Module 7 — Human Resource Management

---

- [ ] **HR-001** · Employee Master Data
  `hr → employees → master-record-and-contracts`

  Implement employee master record (personal info, emergency contacts, identity documents),
  employment history, job assignment (position, grade, department, branch), and contract
  management (contract type, start/end date, probation period). Done when an employee's
  full profile including all job assignments and contracts is queryable and changes are
  fully audited.

  - **Depends on:** [PLAT-002, PLAT-003, PLAT-008]

---

- [ ] **HR-002** · Recruitment
  `hr → recruitment → requisition-to-hire`

  Implement job requisition with headcount approval, job posting, applicant tracking
  (pipeline stages: applied → screened → interviewed → offered → hired/rejected),
  interview scheduling with calendar integration, offer letter generation, and conversion
  to employee record on hire. Done when a complete recruitment cycle from approved
  requisition to hired employee is traceable in a single pipeline view.

  - **Depends on:** [HR-001, PLAT-015, PLAT-013]

---

- [ ] **HR-003** · Attendance & Shift Management
  `hr → attendance → shifts-timesheets-and-tracking`

  Implement shift configuration (shift patterns, rotation schedules), attendance recording
  (manual, biometric hook, mobile check-in), timesheet submission and supervisor approval,
  and overtime calculation. Done when daily attendance is recorded, variance from scheduled
  shift is flagged, and approved timesheets feed into payroll-ready hours.

  - **Depends on:** [HR-001, PLAT-015]

---

- [ ] **HR-004** · Leave Management
  `hr → leave → accrual-request-approval`

  Implement leave type configuration (annual, sick, maternity, paternity, etc.), accrual
  rules (pro-rated by start date, carry-forward limits), leave request and multi-level
  approval workflow, leave balance deduction on approval, and leave calendar view.
  Done when an employee's leave balance is correctly computed from accrual rules, a
  request routes to the right approver, and approval deducts from the correct balance.

  - **Depends on:** [HR-003, PLAT-015]

---

- [ ] **HR-005** · Performance Management
  `hr → performance → appraisal-cycles`

  Implement goal setting and alignment (individual → department → company cascade),
  mid-year and annual appraisal cycle management, 360° feedback collection, calibration
  session workflow, and final rating with outcome (salary review trigger, promotion
  recommendation). Done when a full appraisal cycle from goal-setting through calibration
  to final rating can be completed and results are locked for the period.

  - **Depends on:** [HR-001, PLAT-014]

---

- [ ] **HR-006** · Employee Self-Service
  `hr → self-service → personal-updates-and-requests`

  Implement employee-facing self-service: personal information update requests, leave
  requests, payslip access, training enrolment requests, and expense claim submission.
  All updates that affect payroll or contracts require manager approval. Done when
  an employee can perform all self-service actions from the mobile app and each
  action follows the correct approval path.

  - **Depends on:** [HR-004, PLAT-015, PLAT-011]

---

---

### Module 8 — Payroll Management

---

- [ ] **PAY-001** · Salary Structure & Components
  `payroll → setup → salary-structures-and-components`

  Implement salary structure definitions (basic, allowances, deductions), earning
  component configuration (taxable/non-taxable flag, pensionable flag), statutory
  deduction configuration (PAYE tax tables by jurisdiction, NSSF/NHIF/SHA rates,
  Housing Levy), and salary grade ranges. Done when a complete salary structure can
  be assigned to an employee and all components calculate correctly given a gross
  salary input.

  - **Depends on:** [HR-001, FIN-001]

---

- [ ] **PAY-002** · Payroll Processing Run
  `payroll → processing → monthly-run-and-approval`

  Implement payroll run initiation (period selection, employee scope), earnings and
  deductions computation, variance alerts (salary change > X% triggers review), payroll
  register generation, and approval workflow before payment release. Done when a payroll
  run produces a balanced payroll register where gross pay minus all deductions equals
  net pay for every employee, and the run cannot be released without approval.

  - **Depends on:** [PAY-001, HR-003, PLAT-015]

---

- [ ] **PAY-003** · Statutory Compliance & Filing
  `payroll → compliance → statutory-deductions-and-returns`

  Implement country-specific PAYE computation (Kenya: KRA tax bands, personal relief,
  insurance relief), NSSF/SHA/Housing Levy deduction with correct rate tiers, filing
  format generation (KRA P10, NSSF schedule, NHIF/SHA schedule), and filing submission
  audit trail. Done when a Kenyan payroll run produces filing-ready statutory schedules
  that match KRA's expected format and all computed deductions reconcile to the
  payroll register.

  - **Depends on:** [PAY-002]

---

- [ ] **PAY-004** · Payroll Reporting
  `payroll → reporting → register-and-cost-allocation`

  Implement payroll register (full detail per employee), bank payment file generation
  (EFT format), payslip PDF generation per employee, P9 certificate generation, and
  payroll cost allocation by department/project for GL posting. Done when payslips are
  mathematically correct, the bank file is importable by a Kenyan bank, and the GL
  allocation journal balances.

  - **Depends on:** [PAY-003, FIN-003, PLAT-016]

---

---

### Module 9 — Manufacturing

---

- [ ] **MFG-001** · Bill of Materials
  `manufacturing → bom → multi-level-bom`

  Implement multi-level BOM with component quantities and UOM, BOM versioning and
  revision control, phantom assembly handling, and co-product/by-product configuration.
  Done when a three-level BOM can be exploded to show all raw material requirements
  for a finished good quantity, and a BOM revision creates a new version without
  affecting in-progress production orders.

  - **Depends on:** [INV-001]

---

- [ ] **MFG-002** · Production Planning
  `manufacturing → planning → mps-and-mrp`

  Implement Master Production Schedule (MPS) from demand signals, Material Requirements
  Planning (MRP) explosion from MPS + BOM + stock, capacity planning against work centre
  calendars, and planned order generation. Done when MRP correctly calculates net
  requirements (gross requirement − on-hand − on-order) and generates planned orders
  for each component at the right time.

  - **Depends on:** [MFG-001, INV-005]

---

- [ ] **MFG-003** · Work Orders
  `manufacturing → work-orders → execution-and-tracking`

  Implement work order creation from planned orders, operation sequencing with routing,
  labour and machine time capture, material consumption recording (backflush and manual),
  and work order completion with variance capture. Done when a work order moves from
  released through operations to completed, consumed materials are deducted from inventory,
  and actual vs. standard cost variance is visible.

  - **Depends on:** [MFG-002, INV-002]

---

- [ ] **MFG-004** · Quality Control
  `manufacturing → quality → inspection-and-non-conformance`

  Implement inspection plans per operation, in-process quality checks with pass/fail
  recording, SPC control charts (Xbar-R), non-conformance report (NCR) creation, and
  disposition workflow (rework, scrap, use-as-is). Done when a failed inspection auto-
  creates an NCR, the disposition workflow routes to QA approval, and scrap disposition
  posts a stock adjustment.

  - **Depends on:** [MFG-003, PLAT-015]

---

- [ ] **MFG-005** · Production Reporting
  `manufacturing → reporting → oee-and-cost-per-unit`

  Implement production output vs. planned, OEE metrics (Availability × Performance ×
  Quality), scrap analysis, and cost-per-unit report. Done when OEE is calculable from
  work order actuals and cost-per-unit reconciles to the sum of material, labour, and
  overhead consumed.

  - **Depends on:** [MFG-004, PLAT-016]

---

---

### Module 10 — Project Management

---

- [ ] **PRJ-001** · Project Planning
  `projects → planning → wbs-and-milestones`

  Implement project creation with Work Breakdown Structure (WBS), milestone definitions,
  deliverable tracking, Gantt chart data model, and project budget (cost and revenue
  budget by WBS node). Done when a project with a three-level WBS can be created, its
  Gantt data is queryable for rendering, and budget is assignable at any WBS level.

  - **Depends on:** [PLAT-002, FIN-008]

---

- [ ] **PRJ-002** · Task Management
  `projects → tasks → assignment-and-dependencies`

  Implement task creation, assignment to team members, dependency mapping (FS, SS, FF,
  SF relationship types), and critical path calculation. Done when adding a task
  dependency correctly shifts dependent task dates, the critical path is identifiable,
  and task completion percentage rolls up to the parent WBS node.

  - **Depends on:** [PRJ-001]

---

- [ ] **PRJ-003** · Time Tracking
  `projects → timesheets → entry-and-approval`

  Implement timesheet entry against project and task, weekly timesheet submission,
  supervisor approval workflow, and resource utilisation reporting. Done when approved
  timesheets feed actual hours into project cost calculation and utilisation rate
  is calculable per resource.

  - **Depends on:** [PRJ-002, PLAT-015]

---

- [ ] **PRJ-004** · Project Finance
  `projects → finance → budget-burn-and-evm`

  Implement committed cost tracking (POs and contracts against project), Earned Value
  Management metrics (PV, EV, AC, SPI, CPI), and project P&L report. Done when EVM
  metrics update as work orders and timesheets are posted and the project P&L reconciles
  to the underlying GL transactions.

  - **Depends on:** [PRJ-003, FIN-003]

---

- [ ] **PRJ-005** · Project Reporting
  `projects → reporting → progress-and-budget-dashboards`

  Implement project progress dashboard (% complete by WBS, milestone status, risk
  register summary), resource utilisation heatmap, and budget burn chart. Done when
  all metrics update within one minute of an underlying transaction and the dashboard
  is accessible to project stakeholders with appropriate read-only access.

  - **Depends on:** [PRJ-004, PLAT-016]

---

---

### Module 11 — Service Management

---

- [ ] **SVC-001** · Service Request Management
  `service → requests → intake-classify-route`

  Implement multi-channel service request intake (email parsing, portal form, manual),
  auto-classification by category/priority, SLA policy assignment, and routing to
  assigned technician or team. Done when a request submitted via email is automatically
  classified, assigned the correct SLA, and routed to the right team without manual
  intervention.

  - **Depends on:** [PLAT-011, PLAT-015, PLAT-004]

---

- [ ] **SVC-002** · Field Work Orders
  `service → work-orders → dispatch-complete-capture`

  Implement field technician dispatch from service request, mobile work order completion
  (checklist, photo attachment, customer signature), parts consumption recording, and
  labour time capture. Done when a dispatched work order can be completed on mobile,
  consumed parts deduct from inventory, and the completion record is available in the
  service history.

  - **Depends on:** [SVC-001, INV-002]

---

- [ ] **SVC-003** · SLA Management
  `service → sla → policy-monitoring-breach-alerting`

  Implement SLA policy definitions (response time and resolution time by priority/
  category), real-time SLA clock (pausing on waiting-customer status), breach prediction
  alerts, and breach recording. Done when an SLA clock starts on request creation, pauses
  correctly, and a breach alert fires before the deadline with enough lead time to act.

  - **Depends on:** [SVC-001, PLAT-011]

---

- [ ] **SVC-004** · Service Reporting
  `service → reporting → fcr-and-sla-compliance`

  Implement First Call Resolution rate, Mean Time to Resolution, SLA breach rate by
  category/team, and technician productivity metrics. Done when all metrics reconcile
  to underlying work order records and are available filtered by period, team, and
  service category.

  - **Depends on:** [SVC-003, PLAT-016]

---

---

### Module 12 — Asset Management

---

- [ ] **AST-001** · Asset Registry
  `assets → registry → master-record-and-location`

  Implement asset master record (description, serial number, category, acquisition date,
  cost, responsible person), physical location tracking (building/floor/room), and asset
  photo documentation. Done when every asset has a unique ID, its location is current,
  and the registry is filterable by category, location, and responsible person.

  - **Depends on:** [PLAT-002, FIN-009]

---

- [ ] **AST-002** · Maintenance Management
  `assets → maintenance → preventive-and-corrective`

  Implement preventive maintenance schedules (time-based and meter-based triggers),
  corrective maintenance work order creation from asset failures, maintenance cost
  tracking, and contractor assignment. Done when a preventive schedule automatically
  raises a work order on its trigger date/meter reading and maintenance costs post
  to the asset's cost history.

  - **Depends on:** [AST-001, PLAT-014]

---

- [ ] **AST-003** · Asset Lifecycle & Disposal
  `assets → lifecycle → utilisation-to-retirement`

  Implement utilisation monitoring (runtime hours, odometer), depreciation integration
  with Finance module, inter-entity and inter-location transfer workflow, and retirement/
  disposal process with gain/loss calculation. Done when an asset's disposal produces
  a correct GL entry writing off net book value and recording any gain or loss.

  - **Depends on:** [AST-002, FIN-009]

---

- [ ] **AST-004** · Asset Reporting
  `assets → reporting → register-and-tco`

  Implement full asset register export, maintenance history report, MTBF analysis, and
  Total Cost of Ownership report. Done when TCO reconciles to all acquisition, maintenance,
  and disposal costs recorded against the asset.

  - **Depends on:** [AST-003, PLAT-016]

---

---

### Module 13 — Travel Management

---

- [ ] **TRV-001** · Travel Request & Approval
  `travel → requests → multi-leg-approval`

  Implement multi-leg trip request (purpose, destination, travel dates, estimated cost),
  travel policy validation (class of travel, hotel rate caps), and approval workflow. Done
  when a request that violates policy is flagged before submission, the approval routes
  correctly by cost threshold, and an approved request creates a travel itinerary record.

  - **Depends on:** [PLAT-015, HR-001]

---

- [ ] **TRV-002** · Expense Claims
  `travel → expenses → post-trip-claims-and-per-diem`

  Implement post-trip expense submission, receipt upload per line item, per diem
  calculation (destination-based daily rates), policy violation flagging, and expense
  approval workflow with finance posting. Done when an expense claim with receipts
  is submitted, per diem auto-calculates, policy violations are highlighted, and
  an approved claim posts to the correct GL accounts.

  - **Depends on:** [TRV-001, FIN-003, PLAT-013]

---

- [ ] **TRV-003** · Travel Reporting
  `travel → reporting → spend-and-compliance`

  Implement travel spend by department/employee/destination, policy compliance rate,
  and vendor spend analysis. Done when all spend figures reconcile to approved expense
  claims and policy compliance rate accurately reflects flagged-and-overridden violations.

  - **Depends on:** [TRV-002, PLAT-016]

---

---

### Module 14 — Property Management

---

- [ ] **PROP-001** · Property Registry
  `property → registry → buildings-units-and-facilities`

  Implement building/unit/facility master records, floor plan metadata, amenity listing,
  and condition rating tracking. Done when a multi-storey building with individual units
  is fully described in the registry and units can be queried by vacancy status,
  condition, and floor.

  - **Depends on:** [PLAT-002, PLAT-013]

---

- [ ] **PROP-002** · Lease Management
  `property → leases → agreement-escalation-renewal`

  Implement lease agreement capture (tenant, unit, term, rent amount, payment frequency),
  rent escalation schedule (fixed amount, CPI-linked, percentage), lease renewal workflow
  with advance notification, and IFRS 16 right-of-use asset and lease liability calculation.
  Done when lease liability amortisation schedule matches IFRS 16 calculations and renewal
  workflow fires the configured number of days before lease expiry.

  - **Depends on:** [PROP-001, FIN-009, PLAT-015]

---

- [ ] **PROP-003** · Tenant & Billing Management
  `property → tenants → occupancy-and-automated-billing`

  Implement tenant (lessee) master records, occupancy tracking, vacancy management,
  automated rent billing on schedule, utility billing with submetering, service charge
  allocation, and billing dispute workflow. Done when the billing engine generates
  invoices automatically on the rent due date, utility charges are allocated correctly
  from meter readings, and dispute resolution produces a credit note or revised invoice.

  - **Depends on:** [PROP-002, FIN-005]

---

- [ ] **PROP-004** · Property Maintenance
  `property → maintenance → tenant-requests-and-contractors`

  Implement tenant maintenance request submission (via self-service portal), work order
  creation and contractor assignment, cost tracking, and maintenance cost recovery billing
  to tenant where applicable. Done when a tenant's maintenance request becomes a work
  order, is completed by a contractor, and the recoverable cost is automatically added
  to the next billing cycle.

  - **Depends on:** [PROP-003, SVC-002]

---

---

### Module 15 — Forecourt Management (Fuel Station)

---

- [ ] **FOR-001** · Station & Equipment Setup
  `forecourt → setup → station-pump-nozzle-tank`

  Implement station master record (licence, EPRA registration, tax registration), pump
  master (make/model, calibration due date), nozzle configuration (fuel grade, side A/B,
  opening meter reading), and tank master (capacity, compartments, fuel grade, dip chart
  strapping table). Done when a complete station can be configured with all equipment and
  the dip-to-volume conversion produces accurate litre figures from raw dip measurements.

  - **Depends on:** [PLAT-001, PLAT-002, PLAT-004, PLAT-008]

---

- [ ] **FOR-002** · Fuel Product & Pricing
  `forecourt → products → grade-pricing-and-tax`

  Implement fuel product definitions (grade, density, EPRA tax classification), selling
  price management (effective dating, pump price change audit trail), and tax rate
  linkage. Done when a pump price change is recorded with timestamp and previous price,
  tax is correctly applied to each grade, and pricing history is queryable for any date.

  - **Depends on:** [FOR-001, FIN-010]

---

- [ ] **FOR-003** · Wet Stock Management
  `forecourt → wet-stock → dips-deliveries-reconciliation`

  Implement manual tank dip entry, dip-to-volume conversion using the strapping table,
  fuel delivery recording (actual vs. invoiced quantity, density and temperature
  measurement), theoretical stock calculation (opening + deliveries − sales), and
  daily wet stock reconciliation with variance classification (evaporation, meter
  error, shrinkage, theft). Done when the daily reconciliation report shows correct
  theoretical vs. actual variance per tank and variances outside tolerance auto-raise
  a review task.

  - **Depends on:** [FOR-002, INV-002]

---

- [ ] **FOR-004** · Pump & Meter Management
  `forecourt → pumps → meter-readings-and-calibration`

  Implement opening and closing meter reading capture per shift, totaliser reset
  handling (meter rollover), pump test sale recording, and calibration scheduling
  with EPRA certificate tracking. Done when meter readings correctly accumulate
  volume sold per nozzle, totaliser resets are handled without data loss, and an
  overdue calibration triggers a compliance alert.

  - **Depends on:** [FOR-003]

---

- [ ] **FOR-005** · Shift Management
  `forecourt → shifts → open-handover-close-reconcile`

  Implement shift opening (opening dips, meter readings, cash float, opening checklist),
  mid-shift handover between attendants (partial readings, sign-off), and shift closing
  (closing readings, sales reconciliation across cash/card/credit, short/over calculation,
  shift report generation). Done when a closed shift report shows zero variance when
  cash collected equals total sales by payment method, and any short/over is flagged
  and linked to the responsible attendant.

  - **Depends on:** [FOR-004, HR-001]

---

- [ ] **FOR-006** · Fuel Delivery Management
  `forecourt → deliveries → schedule-verify-loss-gain`

  Implement delivery order raising (reorder level trigger), supplier notification,
  pre-delivery ullage confirmation, post-delivery dip comparison, temperature and
  density verification, and delivery gain/loss report with supplier variance claim
  management. Done when the delivery gain/loss report is auto-generated after every
  delivery and a loss exceeding tolerance auto-drafts a supplier variance claim.

  - **Depends on:** [FOR-003, PRO-004]

---

- [ ] **FOR-007** · Attendant Performance Tracking
  `forecourt → attendants → assignment-and-performance`

  Implement shift-to-attendant assignment, pump assignment, cash float assignment per
  attendant, and performance metrics (sales volume per attendant, short/over history,
  speed-of-service where measurable). Done when each shift's sales and cash variance
  are attributed to the correct attendant and a performance history view is available
  per employee.

  - **Depends on:** [FOR-005, HR-001]

---

- [ ] **FOR-008** · Forecourt Reporting
  `forecourt → reporting → sales-wetstock-variance`

  Implement daily sales reports (by product/pump/attendant, payment method split, price
  variance), wet stock reports (tank dip history, delivery history, closing stock position),
  and variance trend analysis (daily/weekly/monthly wet stock variance, comparative bench).
  Done when all figures in the daily sales report reconcile to shift closing records and
  wet stock reports reconcile to the perpetual fuel inventory.

  - **Depends on:** [FOR-007, PLAT-016]

---

---

### Module 16 — Airline Management

---

- [ ] **AIR-001** · Flight Inventory & Schedule
  `airline → inventory → schedule-and-seat-inventory`

  Implement flight schedule management (flight number, route, operating days, aircraft
  type), seat inventory configuration (cabin classes, seat map, overbooking level per
  cabin), and fare class structure (RBD booking class, fare basis code, fare rules and
  conditions). Done when a flight's seat inventory correctly opens and closes booking
  classes as seats sell and the overbooking level is not exceeded.

  - **Depends on:** [PLAT-001, PLAT-004, PLAT-008]

---

- [ ] **AIR-002** · Reservation & Ticketing
  `airline → reservations → booking-ticketing-lifecycle`

  Implement one-way/return/multi-city booking, seat map selection, ancillary upsell,
  e-ticket generation with unique ticket number, and ticket lifecycle management
  (reissue, refund, exchange). Done when a booking creates a PNR, an e-ticket is
  issued, and exchange/refund correctly adjusts the fare and logs the fee.

  - **Depends on:** [AIR-001, FIN-005]

---

- [ ] **AIR-003** · Revenue Management
  `airline → revenue-management → dynamic-pricing-and-yield`

  Implement demand forecasting (historical demand, seasonality patterns), dynamic pricing
  via bid price calculation, availability control per booking class, and seat allocation
  optimisation (EMSRb model). Done when closing a booking class when the bid price is
  exceeded demonstrably increases average fare on a simulated flight.

  - **Depends on:** [AIR-001]

---

- [ ] **AIR-004** · Check-In & Boarding
  `airline → operations → check-in-and-boarding-pass`

  Implement web check-in, airport agent check-in, boarding pass generation (QR/barcode),
  boarding zone assignment, and upgrade processing. Done when a checked-in passenger
  has a valid scannable boarding pass and upgrades correctly move the passenger to the
  new cabin with fare difference calculation.

  - **Depends on:** [AIR-002]

---

- [ ] **AIR-005** · Airline Reporting
  `airline → reporting → rask-yield-and-load-factor`

  Implement RASK (Revenue per Available Seat Kilometre), yield analysis by route/cabin/
  fare class, passenger and weight load factor, and breakeven load factor calculation.
  Done when RASK reconciles to ticketed revenue divided by available seat kilometres
  for the period.

  - **Depends on:** [AIR-003, PLAT-016]

---

---

### Module 17 — Business Intelligence

---

- [ ] **BI-001** · Data Model & Star Schemas
  `bi → data-model → star-schemas-per-module`

  Implement pre-built star schemas for Finance (GL facts, budget facts), Sales (order
  facts, revenue facts), HR (headcount, attendance), and Procurement (spend facts).
  Includes refresh pipeline (near-real-time, daily, weekly cadences) and dimension tables
  (time, org, product, customer, supplier). Done when each star schema refreshes on
  schedule and query performance on 12-month rolling data is under 3 seconds.

  - **Depends on:** [FIN-011, SAL-005, HR-006, PRO-007]

---

- [ ] **BI-002** · KPI Management
  `bi → kpis → definitions-targets-and-monitoring`

  Implement KPI formula builder (referencing star schema measures), target setting by
  period and responsibility owner, KPI scorecard with RAG status (Red/Amber/Green),
  trend sparklines, and breach alert configuration. Done when a KPI misses its target
  and an alert fires to the responsible owner within the configured notification window.

  - **Depends on:** [BI-001, PLAT-011]

---

- [ ] **BI-003** · Executive & Operational Dashboards
  `bi → dashboards → board-level-and-module-operational`

  Implement board-level KPI pack (revenue, margins, working capital, headcount,
  compliance health) and module-specific operational dashboards with drill-down to
  transaction level. Includes role-based default dashboard assignment and user
  personalisation. Done when an executive sees the board pack on login and can drill
  from a revenue KPI to the underlying GL journal entries.

  - **Depends on:** [BI-002, PLAT-004]

---

- [ ] **BI-004** · Ad Hoc Analysis & Forecasting
  `bi → exploration → pivot-analysis-and-forecasting`

  Implement drag-and-drop pivot builder over star schemas, cross-module drill paths,
  revenue forecasting (linear, exponential smoothing, ARIMA models), and demand
  forecasting linked to inventory planning. Done when a non-technical user can build
  a pivot report without writing SQL and a statistical forecast is generated and
  exportable to the inventory planning module.

  - **Depends on:** [BI-003, INV-005]

---

---

## 🖥️ PART C — SERVER-DRIVEN UI (SDUI) PLATFORM

---

### Volume I — Vision & Philosophy

---

- [ ] **SDUI-001** · Platform Vision & Design Tenets Documentation
  `sdui → vision → tenets-and-philosophy`

  Document and ratify the seven core design tenets (UI as Data, Backend as Source of
  Truth, Clients as Rendering Engines, Business Vocabulary, Permissions as First-Class,
  Extensibility Without Forking, Observability Built In). Includes non-goals, accepted
  trade-offs, and the audience guide with tailored reading paths per engineering role.
  Done when the vision document passes review by all engineering leads and is published
  as the authoritative reference.

  - **Depends on:** []

---

- [ ] **SDUI-002** · System Architecture Overview
  `sdui → vision → component-inventory-and-request-lifecycle`

  Document the full platform component inventory (Compilation Service, Component Registry,
  Schema Validator, Feature Flag Resolver, Authorization Resolver, Localization Service,
  Web Engine, Mobile Engine) and the 7-step request lifecycle from client request to
  rendered UI. Includes deployment topology, scalability characteristics, and known
  limitation register. Done when any new engineer can trace a UI request end-to-end
  using only this document.

  - **Depends on:** [SDUI-001, PLAT-001]

---

---

### Volume II — DSL & AST

---

- [ ] **SDUI-003** · AST Node Schema & Go Builders
  `sdui → dsl-and-ast → node-schema-and-builders`

  Define the canonical AST node schema: id, type, props, children, slots, events,
  actions, permissions, flags, and metadata. Implement Go typed node builders and
  the Component Registry with namespaced type IDs. Done when any screen definition
  can be expressed as a valid, serializable AST, round-tripped through the registry,
  and structurally validated without runtime panics.

  - **Depends on:** [SDUI-002, PLAT-004, PLAT-009]

---

- [ ] **SDUI-004** · Binding Expressions & Directives
  `sdui → dsl-and-ast → bindings-and-directives`

  Implement the DSL expression system: static, data, conditional, computed, and
  permission bindings. Implement directives: visibility, repeat, permission guard,
  feature flag guard, and locale directive. Done when a screen definition using all
  five binding types and all five directives compiles to a valid AST and each directive
  produces the correct structural transformation.

  - **Depends on:** [SDUI-003]

---

- [ ] **SDUI-005** · AST Transformation Pipeline
  `sdui → dsl-and-ast → transformation-passes`

  Implement the six transformation passes: permission pruning (removes nodes the actor
  cannot see), feature flag resolution (removes flag-gated nodes), localization injection,
  default value population, tenant override application, and canonical form normalization.
  Done when a raw AST for a Finance screen, when transformed for a user with limited
  permissions and a non-English locale, correctly removes restricted nodes and injects
  translated strings.

  - **Depends on:** [SDUI-004, PLAT-004, PLAT-009, PLAT-012]

---

- [ ] **SDUI-006** · JSON Compilation Pipeline
  `sdui → dsl-and-ast → compilation-pipeline`

  Implement the full 6-stage compilation pipeline: Context Resolution → AST Construction
  → AST Transformation → Validation → Serialization → Transport. Includes Redis caching
  of compiled ASTs keyed by tenant + user-role-set + screen + locale, cache invalidation
  on permission or flag change, payload compression, and per-stage OpenTelemetry spans.
  Done when p99 compilation latency for a cached screen is under 10ms and a cold compile
  is under 100ms.

  - **Depends on:** [SDUI-005]

---

---

### Volume III — Component System

---

- [ ] **SDUI-007** · Primitive Component Library
  `sdui → components → primitive-library`

  Implement the primitive component set: Text, Button, Icon, Image, Badge, Avatar,
  Spinner, Tooltip, Divider, Spacer, Link, and Tag. Each component has a fully specified
  props schema, emitted events, and renders correctly on both web (HTML) and mobile
  (native) surfaces. Done when all primitives pass the component contract test suite
  on both renderers with zero layout regressions.

  - **Depends on:** [SDUI-003]

---

- [ ] **SDUI-008** · Compound Component Library
  `sdui → components → compound-library`

  Implement compound components: Card, Alert, Modal, Drawer, Stepper, Accordion, Tabs,
  Breadcrumb, Empty State, and Error Boundary. Each component composes primitives, exposes
  named slots, and handles its own internal state (open/closed, active step, active tab).
  Done when all compound components render on both surfaces and slot composition produces
  the expected output for all documented slot combinations.

  - **Depends on:** [SDUI-007]

---

- [ ] **SDUI-009** · Layout System
  `sdui → components → layout-and-grid`

  Implement the layout model: flow, grid, flex, stack, and absolute. Includes the column
  grid system (12-column, configurable gap/gutter), responsive breakpoints (xs/sm/md/lg/xl),
  spacing token system (margin/padding, density modes), and layout container components
  (Page, Section, Panel, SplitPane, ScrollArea, StickyHeader). Done when a three-column
  desktop layout correctly collapses to single-column on mobile using only AST breakpoint
  directives.

  - **Depends on:** [SDUI-008]

---

- [ ] **SDUI-010** · Forms Framework
  `sdui → components → forms-framework`

  Implement the forms system: 22 field input types (text, number, select, multi-select,
  date, time, datetime, currency, entity-picker, file-upload, rich-text, toggle, radio,
  checkbox, colour, rating, signature, address, phone, email, URL, multi-currency),
  form layout (multi-column, collapsible sections, wizard tabs), dynamic forms
  (conditional visibility/required, computed fields, repeating groups), and ERP-specific
  patterns (header + line items, approval routing fields, document attachments). Done when
  a multi-section dynamic form with conditional logic renders correctly on both surfaces
  and all 22 field types round-trip their values through submit without data loss.

  - **Depends on:** [SDUI-009, PLAT-004]

---

- [ ] **SDUI-011** · Tables & Data Grids
  `sdui → components → tables-and-grids`

  Implement the table framework with 10 column types (text, number, currency, date,
  badge, link, action, boolean, progress, custom) and 15 features (sorting, filtering,
  pagination, row selection, row expansion, frozen columns, column resize, inline editing,
  bulk actions, export, row-level permissions, column-level permissions, tree/hierarchy,
  aggregation row, virtual scroll). Done when a 10,000-row table renders without visible
  lag using virtual scroll and row-level permissions from FGA correctly hide individual
  rows.

  - **Depends on:** [SDUI-009, PLAT-004]

---

- [ ] **SDUI-012** · Dashboard Framework
  `sdui → components → dashboard-widgets`

  Implement the dashboard framework with 12 widget types (KPI card, trend sparkline,
  bar/line/pie chart, data table, activity feed, pending approvals, map/geo, calendar,
  alert panel, gauge, funnel). Includes four layout types (fixed grid, responsive masonry,
  tab-based, split panel), role-based default layouts, user personalisation, and
  widget-level data source binding. Done when a dashboard with all 12 widget types
  renders on both surfaces and user layout preferences persist across sessions.

  - **Depends on:** [SDUI-011, PLAT-004]

---

- [ ] **SDUI-013** · Charts & Analytics Components
  `sdui → components → charts-and-analytics`

  Implement 13 chart types (bar, stacked bar, line, area, pie, donut, scatter, bubble,
  heatmap, Gantt, treemap, waterfall, Sankey), chart data binding (static, DataSource
  reference, aggregation expressions, real-time subscription), interactivity (click-to-
  action, drill-down navigation, cross-filter), and export (PNG/SVG/PDF). Done when
  cross-filter between two charts on the same dashboard correctly filters both data
  sources and drill-down navigates to the configured detail screen.

  - **Depends on:** [SDUI-012]

---

- [ ] **SDUI-014** · Workflow & Approval Components
  `sdui → components → workflow-and-approvals`

  Implement workflow state machine display, approval action buttons (Approve, Reject,
  Request Revision, Delegate, Escalate, Withdraw) with Temporal signal binding, SLA
  countdown timer, pending approval queue widget, and immutable approval audit trail
  component. Done when an approval action button fires the correct Temporal signal,
  the SLA timer updates in real time, and the audit trail shows all decisions in
  chronological order.

  - **Depends on:** [SDUI-012, PLAT-014, PLAT-015]

---

- [ ] **SDUI-015** · ERP Domain Components
  `sdui → components → erp-domain-library`

  Implement ERP-specific components: Financial (currency display, amount breakdown,
  tax summary, ledger row, journal voucher, trial balance table, P&L summary, budget
  vs. actual bar), Procurement (PO header, line item grid, vendor card, RFQ comparison
  table, GRN viewer), HR (employee card, org chart, leave balance panel, payslip viewer,
  attendance heatmap), Sales (customer card, pipeline stage indicator, invoice viewer,
  credit limit warning), and Document (PDF viewer, e-signature capture, QR code display,
  barcode scanner input). Done when each domain component renders correctly with realistic
  test data on both web and mobile surfaces.

  - **Depends on:** [SDUI-014, FIN-001, CRM-001, HR-001, SAL-001]

---

---

### Volume IV — Rendering Architecture

---

- [ ] **SDUI-016** · Mobile Rendering Engine
  `sdui → rendering → mobile-engine`

  Implement the mobile rendering engine: JSON Parser, Component Resolver, Rendering Tree
  Builder, View Reconciler, Action Dispatcher, and Event Bus. Includes component-to-native
  mapping table, fallback strategy for unknown components, Flutter/React Native layout
  bridge, form keyboard avoidance, native picker integration, and list virtualisation.
  Done when any valid AST produced by the compilation pipeline renders on the mobile
  surface without crashes and all interactive elements dispatch actions correctly.

  - **Depends on:** [SDUI-015]

---

- [ ] **SDUI-017** · Web Rendering Engine
  `sdui → rendering → web-engine`

  Implement the web rendering engine: JSON Fetcher, Schema Validator, Web Component
  Registry, Virtual DOM Builder, Reactive State Binding, Action Dispatcher, and Event
  System. Includes CSS Grid layout, responsive breakpoints, dark mode token switching,
  file upload handling, code splitting per screen, virtual scrolling, and progressive
  hydration. Done when any valid AST renders on the web surface, dark mode switches
  without page reload, and Lighthouse performance score is ≥ 90 for a representative
  ERP screen.

  - **Depends on:** [SDUI-015]

---

- [ ] **SDUI-018** · Navigation Framework
  `sdui → rendering → navigation-and-routing`

  Implement 7 navigation node types (top nav, side nav, bottom tab bar, breadcrumb,
  back button, contextual menu, command palette), route system with parameters and deep
  linking, permission-aware menu pruning, and 8 navigation action types (Push, Replace,
  Pop, Modal Open, Modal Close, External URL, Deep Link, Back). Done when menu items
  invisible to a role are absent from the navigation AST before delivery and deep links
  resolve correctly on both web and mobile.

  - **Depends on:** [SDUI-017, SDUI-016, PLAT-004]

---

- [ ] **SDUI-019** · Action System
  `sdui → rendering → action-dispatch`

  Implement the full action type registry: HTTP API call, Temporal workflow trigger,
  Navigation action, Form submit, Download, Copy to clipboard, Custom event emit,
  Confirmation dialog gate, and Optimistic UI update. Includes action chaining (on-
  success / on-error / on-finally), action permission guard, and action audit logging.
  Done when a chained action sequence (confirm → API call → navigate on success → show
  error on failure) executes correctly and the audit log records the API call action.

  - **Depends on:** [SDUI-018, PLAT-008]

---

---

### Volume V — Runtime & State

---

- [ ] **SDUI-020** · Client State Management
  `sdui → runtime → client-state`

  Implement client-side state model: form state (dirty/touched/error/submitting),
  list state (selection, sort, filter, pagination cursor), modal state stack, and
  global notification queue. Includes state reset on navigation and cross-component
  state sharing via named state slots. Done when navigating away from a dirty form
  prompts the user before discarding state and state shared between two components
  via a named slot updates both simultaneously.

  - **Depends on:** [SDUI-019]

---

- [ ] **SDUI-021** · Offline Support & Sync
  `sdui → runtime → offline-and-sync`

  Implement offline detection, local operation queue (actions queued while offline),
  conflict resolution strategy on reconnect (last-write-wins with server validation),
  and offline-available screen configuration. Done when a user submits a form while
  offline, the action is queued, and on reconnect the action is replayed and the result
  reflected in the UI without manual refresh.

  - **Depends on:** [SDUI-020]

---

---

### Volume VI — Platform Integration

---

- [ ] **SDUI-022** · Multi-Tenant Screen Customisation
  `sdui → platform → tenant-screen-overrides`

  Implement the tenant override layer: field label overrides, field visibility overrides,
  additional tenant-specific fields, and layout overrides — all applied at AST
  transformation time without forking screen definitions. Done when a tenant-specific
  override hides a field that is visible in the base screen definition, and the override
  is applied without any change to the Go screen definition code.

  - **Depends on:** [SDUI-005, PLAT-010]

---

- [ ] **SDUI-023** · Feature-Flag-Gated UI
  `sdui → platform → feature-flag-ui-integration`

  Implement feature flag directives in the AST so entire screens, sections, or individual
  components can be gated behind a flag evaluated at compile time. Includes flag-change
  cache invalidation (new compiled AST on flag toggle) and graceful fallback when a
  flagged component is removed. Done when toggling a feature flag for a tenant causes
  the affected screen section to appear or disappear on next page load without a
  deployment.

  - **Depends on:** [SDUI-005, PLAT-009]

---

---

### Volume VII — Reliability & Security

---

- [ ] **SDUI-024** · Schema Versioning & Client Compatibility
  `sdui → reliability → schema-versioning`

  Implement AST schema versioning (semantic version in every compiled response),
  client version negotiation (client sends min-supported version, server responds with
  compatible schema or upgrade prompt), and backward compatibility guarantees (additive-
  only changes within a major version). Done when an older client receiving a newer
  schema version gracefully degrades unrecognised nodes to their fallback and does not
  crash.

  - **Depends on:** [SDUI-006]

---

- [ ] **SDUI-025** · SDUI Security Model
  `sdui → reliability → trust-model-and-permission-enforcement`

  Document and implement the SDUI trust model: permissions are enforced server-side
  during AST transformation (pruning), client-side hiding is treated as UX only and
  never a security control, all actions are re-authorised at the API layer on execution.
  Includes penetration test checklist for SDUI-specific vectors (AST injection, field
  enumeration via partial renders, action replay). Done when the security checklist
  passes and no permission can be bypassed by crafting a manual API call that the
  SDUI action system would have blocked.

  - **Depends on:** [SDUI-005, PLAT-004]

---

---

### Volume VIII — Developer Experience

---

- [ ] **SDUI-026** · Screen Development Workflow
  `sdui → dx → screen-authoring-and-testing`

  Implement the local screen development workflow: Go screen builder functions,
  hot-reload of compiled AST in development mode, AST inspector (visualise the compiled
  tree in browser devtools), and screen unit test harness (assert compiled AST shape
  without a running renderer). Done when a developer can write a new screen, see it
  compiled in under 2 seconds, and assert its structure in a unit test that runs in CI.

  - **Depends on:** [SDUI-006]

---

- [ ] **SDUI-027** · Component Extension API
  `sdui → dx → custom-component-registration`

  Implement the extension API allowing third-party components to be registered into
  the Component Registry with a namespaced type ID, props schema, and renderer
  implementation for each target surface. Done when a custom component registered
  via the extension API renders correctly on both web and mobile surfaces and is
  indistinguishable from a built-in component from the compilation pipeline's perspective.

  - **Depends on:** [SDUI-026]

---

---

### Volume IX — Engineering Standards

---

- [ ] **SDUI-028** · SDUI Testing Strategy
  `sdui → engineering → testing-pyramid`

  Implement and document the SDUI testing pyramid: AST unit tests (screen shape
  assertions), renderer integration tests (component renders given valid AST node),
  end-to-end tests (user journey from API request to pixel), and performance regression
  tests (compilation latency budget, payload size budget). Done when CI enforces all
  four test layers and a failing compilation latency test blocks merge.

  - **Depends on:** [SDUI-026, SDUI-017, SDUI-016]

---

- [ ] **SDUI-029** · SDUI Observability
  `sdui → engineering → observability`

  Implement per-stage compilation spans (OpenTelemetry), payload size metrics per screen
  per tenant, client-side render time telemetry (reported back via beacon API), and
  error rate dashboards (compilation failures, renderer crashes, action failures). Done
  when a compilation latency spike for a specific tenant is diagnosable within 5 minutes
  using only the observability dashboards.

  - **Depends on:** [SDUI-006, PLAT-019]

---

---

### Volume X — Reference

---

- [ ] **SDUI-030** · Component API Reference
  `sdui → reference → component-api-docs`

  Publish the complete component API reference: every component type, its full props
  schema with types and defaults, all named slots, all emitted events, all supported
  actions, permission surface, rendering constraints, and code examples for each surface.
  Done when every component documented in Volumes III–IV has a corresponding reference
  entry and examples are verified by the component test suite.

  - **Depends on:** [SDUI-015, SDUI-017, SDUI-016]

---

- [ ] **SDUI-031** · End-to-End Screen Examples
  `sdui → reference → worked-examples`

  Produce complete worked examples for: a Finance journal entry form (multi-section,
  approval action), a Procurement purchase order screen (header + line item grid, 3-way
  match status), a Forecourt shift closing screen (meter reading form, reconciliation
  table), and an HR leave request screen (balance widget, date picker, approval chain).
  Each example includes the Go screen definition, compiled AST JSON, and screenshots
  on both web and mobile. Done when all four examples pass the screen unit test harness
  and render without errors on both surfaces.

  - **Depends on:** [SDUI-030, FIN-003, PRO-006, FOR-005, HR-004]

---

---

## Summary Stats

| Domain | Tasks | First Task | Last Task |
|---|---|---|---|
| Platform Core | 20 | PLAT-001 | PLAT-020 |
| Finance | 12 | FIN-001 | FIN-012 |
| CRM | 6 | CRM-001 | CRM-006 |
| Sales | 5 | SAL-001 | SAL-005 |
| Procurement | 7 | PRO-001 | PRO-007 |
| Inventory | 7 | INV-001 | INV-007 |
| Warehouse | 5 | WHS-001 | WHS-005 |
| HR | 6 | HR-001 | HR-006 |
| Payroll | 4 | PAY-001 | PAY-004 |
| Manufacturing | 5 | MFG-001 | MFG-005 |
| Projects | 5 | PRJ-001 | PRJ-005 |
| Service | 4 | SVC-001 | SVC-004 |
| Assets | 4 | AST-001 | AST-004 |
| Travel | 3 | TRV-001 | TRV-003 |
| Property | 4 | PROP-001 | PROP-004 |
| Forecourt | 8 | FOR-001 | FOR-008 |
| Airline | 5 | AIR-001 | AIR-005 |
| Business Intelligence | 4 | BI-001 | BI-004 |
| SDUI Platform | 31 | SDUI-001 | SDUI-031 |
| **Total** | **164** | | |
