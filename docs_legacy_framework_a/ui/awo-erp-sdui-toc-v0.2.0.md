> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# AwoERP Server-Driven UI Platform
## Official Architecture Documentation — Table of Contents & Documentation Plan

> **Status:** Proposed TOC — Pending Review & Approval
> **Version:** 0.2.0-draft
> **Authors:** Architecture Team
> **Changelog:** v0.2.0 — Added Chapters 19-B (Backend-Driven Navigation), 52 (API Integration Patterns), 53 (Platform / Tenant / Portal View Differentiation), 54 (amis Web Renderer — index.html & Bootstrap); promoted recommended additions B, C, D, E, F, G, H, I, J to numbered chapters.
> **Audience:** Platform Engineers · Backend Engineers · Mobile Engineers · Web Engineers · UI Framework Developers · Product Owners · Solution Architects · ERP Implementers · Third-Party Integrators

---

## Document Map

This TOC organises the entire AwoERP SDUI platform documentation into **11 volumes**, **58 chapters**, and approximately **250+ sections**. It is structured so that:

- Any engineer can locate their domain quickly.
- The documentation can be built incrementally across 4 phases.
- Each chapter maps to a discrete, ownable unit of work.

---

## Suggested Repository Structure

```
./docs/ui/
├── README.md
├── GLOSSARY.md
├── CHANGELOG.md
├── mkdocs.yml
│
├── vol-01-vision/
│   ├── 01-introduction.md
│   ├── 02-design-philosophy.md
│   ├── 03-architectural-principles.md
│   └── 04-system-overview.md
│
├── vol-02-dsl-and-ast/
│   ├── 05-sdui-fundamentals.md
│   ├── 06-ui-dsl-architecture.md
│   ├── 07-ast-design.md
│   ├── 07b-ui-composition-patterns.md          ← NEW (was D)
│   └── 08-json-compilation-pipeline.md
│
├── vol-03-component-system/
│   ├── 09-component-system.md
│   ├── 10-layout-system.md
│   ├── 10b-theming-and-design-tokens.md        ← NEW (was H)
│   ├── 11-forms-framework.md
│   ├── 12-tables-and-data-grids.md
│   ├── 12b-print-and-export-rendering.md       ← NEW (was G)
│   ├── 13-dashboard-framework.md
│   ├── 14-charts-and-analytics.md
│   ├── 15-workflow-and-approval-components.md
│   └── 16-erp-specific-components.md
│
├── vol-04-rendering/
│   ├── 17-mobile-rendering-architecture.md
│   ├── 18-web-rendering-architecture.md
│   ├── 19-navigation-framework.md
│   ├── 19b-backend-driven-navigation.md        ← NEW ★
│   ├── 20-action-system.md
│   ├── 20b-real-time-ui-updates.md             ← NEW (was F)
│   └── 21-event-system.md
│
├── vol-05-runtime/
│   ├── 22-state-management.md
│   ├── 23-validation-framework.md
│   ├── 24-data-sources.md
│   └── 24b-error-handling-strategy.md          ← NEW (was B)
│
├── vol-06-platform/
│   ├── 25-api-contracts.md
│   ├── 25b-api-integration-patterns.md         ← NEW ★
│   ├── 26-security-model.md
│   ├── 27-authorization-integration.md
│   ├── 28-multi-tenant-architecture.md
│   ├── 28b-platform-tenant-portal-views.md     ← NEW ★
│   ├── 29-feature-flag-architecture.md
│   ├── 30-customization-framework.md
│   ├── 31-extension-and-plugin-architecture.md
│   └── 31b-audit-and-compliance.md             ← NEW (was E)
│
├── vol-07-reliability/
│   ├── 32-performance-optimization.md
│   ├── 33-offline-support.md
│   ├── 34-caching-strategies.md
│   └── 35-synchronization.md
│
├── vol-08-dx/
│   ├── 36-internationalization.md
│   ├── 37-accessibility.md
│   ├── 38-observability.md
│   ├── 39-telemetry.md
│   ├── 40-logging.md
│   └── 40b-developer-tooling.md                ← NEW (was C)
│
├── vol-09-engineering/
│   ├── 41-testing-strategy.md
│   ├── 42-cicd-considerations.md
│   ├── 43-versioning-strategy.md
│   ├── 44-migration-strategy.md
│   └── 45-governance-model.md
│
├── vol-10-reference/
│   ├── 46-best-practices.md
│   ├── 47-anti-patterns.md
│   ├── 48-reference-architecture.md
│   ├── 49-end-to-end-examples.md
│   ├── 50-sdk-development.md
│   ├── 50b-onboarding-and-migration-guide.md   ← NEW (was I)
│   ├── 50c-playground-sandbox-environment.md   ← NEW (was J)
│   └── 51-future-roadmap.md
│
├── vol-11-amis-web-renderer/                   ← NEW VOLUME ★
│   ├── 52-amis-web-renderer-bootstrap.md
│   └── 52-appendix-amis-index-html-reference.md
│
├── schemas/
│   ├── component.schema.json
│   ├── layout.schema.json
│   ├── action.schema.json
│   ├── event.schema.json
│   ├── navigation.schema.json                  ← NEW
│   ├── portal.schema.json                      ← NEW
│   └── page.schema.json
│
├── examples/
│   ├── purchase-order-form/
│   ├── invoice-dashboard/
│   ├── approval-workflow/
│   ├── multi-tenant-demo/
│   ├── backend-driven-nav-demo/                ← NEW
│   └── portal-view-demo/                       ← NEW
│
└── assets/
    ├── diagrams/
    └── images/
```

---

## Estimated Documentation Phases

| Phase | Volumes | Chapters | Focus | Priority |
|-------|---------|----------|-------|----------|
| **Phase 1** — Foundation | I, II | 01–08, 07b | Vision, DSL, AST, composition patterns, compilation pipeline | Critical / Ship first |
| **Phase 2** — Components & Rendering | III, IV | 09–21, 10b, 12b, 19b, 20b | All components, theming, print/export, web & mobile rendering, backend-driven nav, real-time updates | Critical / Ship second |
| **Phase 3** — Platform & Runtime | V, VI, VII | 22–35, 24b, 25b, 28b, 31b | State, APIs, API integration, platform/tenant/portal views, security, tenancy, compliance, performance | High |
| **Phase 4** — Excellence & Reference | VIII, IX, X, XI | 36–54, 40b, 50b, 50c | DX, dev tooling, testing, governance, SDK, onboarding, playground, amis bootstrap | Medium |

---

---

# Full Hierarchical Table of Contents

---

## Volume I — Vision & Philosophy

*(Unchanged from v0.1.0 — see below for completeness)*

### Chapter 01 — Introduction to AwoERP UI Platform
### Chapter 02 — Design Philosophy
### Chapter 03 — Architectural Principles
### Chapter 04 — System Overview

*(Full section detail retained from v0.1.0)*

---

## Volume II — DSL & AST

### Chapter 05 — Server-Driven UI Fundamentals
### Chapter 06 — UI DSL Architecture
### Chapter 07 — AST Design

---

### Chapter 07-B — UI Composition Patterns *(NEW — was Recommended Addition D)*

> **Audience:** Platform Engineers, Backend Engineers · **Phase:** 1

- 07B.1 Why Composition Matters in SDUI
- 07B.2 Template System
  - 07B.2.1 Named Templates and Template Registry
  - 07B.2.2 Template Parameters and Defaults
  - 07B.2.3 Template Versioning
- 07B.3 Slot-Based Composition
  - 07B.3.1 Named Slots
  - 07B.3.2 Default Slot Content (Fallbacks)
  - 07B.3.3 Slot Override Rules (Tenant vs. User)
  - 07B.3.4 Recursive Slot Nesting
- 07B.4 Inheritance and Extension
  - 07B.4.1 Base Definitions and Derived Definitions
  - 07B.4.2 Override Semantics (Replace vs. Merge vs. Extend)
  - 07B.4.3 Sealed vs. Open Templates
- 07B.5 Mixin System
  - 07B.5.1 Declaring Mixins
  - 07B.5.2 Applying Mixins to Nodes
  - 07B.5.3 Mixin Conflict Resolution
- 07B.6 Fragment Reuse (Partial ASTs)
- 07B.7 Composition Anti-Patterns
- 07B.8 Composition and Multi-Tenancy

---

### Chapter 08 — JSON Compilation Pipeline

---

## Volume III — Component System

### Chapter 09 — Component System
### Chapter 10 — Layout System

---

### Chapter 10-B — Theming and Design Tokens *(NEW — was Recommended Addition H)*

> **Audience:** Web/Mobile Engineers, UI Framework Developers, ERP Implementers · **Phase:** 2

- 10B.1 Theming Philosophy
- 10B.2 Design Token System
  - 10B.2.1 Token Categories (Color, Typography, Spacing, Elevation, Radius, Motion)
  - 10B.2.2 Semantic Tokens vs. Primitive Tokens
  - 10B.2.3 Token Inheritance Hierarchy
  - 10B.2.4 Token Reference in Component Props
- 10B.3 Platform Default Theme
  - 10B.3.1 Light Mode Token Set
  - 10B.3.2 Dark Mode Token Set
  - 10B.3.3 High-Contrast Mode Token Set
- 10B.4 Tenant Branding Layer
  - 10B.4.1 Tenant Logo and Favicon
  - 10B.4.2 Tenant Primary / Accent Color Overrides
  - 10B.4.3 Tenant Typography Overrides
  - 10B.4.4 Tenant Theme Storage (PostgreSQL + Redis)
  - 10B.4.5 Tenant Theme API Endpoints
- 10B.5 Portal-Level Theme Overrides
  - 10B.5.1 Supplier Portal Theme
  - 10B.5.2 Customer Portal Theme
  - 10B.5.3 Employee Self-Service Portal Theme
- 10B.6 amis `cxd` Theme Mapping
  - 10B.6.1 CSS Variable Override Points
  - 10B.6.2 Permitted Customisation Surface
  - 10B.6.3 Constraints and Anti-Patterns
- 10B.7 Dark Mode Runtime Switching
- 10B.8 Component-Level Style Override Mechanism
- 10B.9 Theme Export and Import (Tenant Migration)
- 10B.10 Design Token CI Validation

---

### Chapter 11 — Forms Framework
### Chapter 12 — Tables and Data Grids

---

### Chapter 12-B — Print and Export Rendering *(NEW — was Recommended Addition G)*

> **Audience:** Backend Engineers, Web/Mobile Engineers, ERP Implementers · **Phase:** 2

- 12B.1 Print and Export Philosophy
- 12B.2 Printable Surface Type
  - 12B.2.1 Declaring a Surface as Printable
  - 12B.2.2 Print vs. Screen Layout Variants
  - 12B.2.3 Header / Footer Bands in Print Mode
- 12B.3 PDF Generation from UI Definitions
  - 12B.3.1 Server-Side PDF Pipeline (Go → chromedp / wkhtmltopdf)
  - 12B.3.2 Client-Side PDF (Browser Print API / jsPDF)
  - 12B.3.3 PDF Output Schema
  - 12B.3.4 Page Break Hints in AST
- 12B.4 ERP Print Templates
  - 12B.4.1 Invoice / Tax Invoice
  - 12B.4.2 Purchase Order
  - 12B.4.3 Delivery Note / GRN
  - 12B.4.4 Payslip
  - 12B.4.5 Statement of Account
  - 12B.4.6 Picking Slip
- 12B.5 Export Actions
  - 12B.5.1 Export to CSV
  - 12B.5.2 Export to Excel (XLSX)
  - 12B.5.3 Export to PDF
  - 12B.5.4 Bulk Export Queuing (Temporal Workflow)
- 12B.6 Print Permission Controls
- 12B.7 Watermarking and Draft Overlays
- 12B.8 Tenant Letterhead Injection
- 12B.9 KRA eTIMS QR Code Embedding in Print Output

---

### Chapter 13 — Dashboard Framework
### Chapter 14 — Charts and Analytics
### Chapter 15 — Workflow and Approval Components
### Chapter 16 — ERP-Specific Components

---

## Volume IV — Rendering Architecture

### Chapter 17 — Mobile Rendering Architecture
### Chapter 18 — Web Rendering Architecture
### Chapter 19 — Navigation Framework

---

### Chapter 19-B — Backend-Driven Navigation *(NEW ★)*

> **Audience:** Platform Engineers, Backend (Go/Fiber) Engineers, Mobile/Web Engineers · **Phase:** 2

This chapter covers how the **Go backend is the authoritative source of the full navigation tree**. No navigation node is hard-coded in any client. Clients are rendering engines that receive, cache, and hydrate the navigation payload.

- 19B.1 Philosophy — Navigation as a First-Class API Resource
  - 19B.1.1 Why the Backend Owns Navigation
  - 19B.1.2 Navigation as a Permission Surface
  - 19B.1.3 Navigation as a Tenant Configuration Surface
  - 19B.1.4 Contrast with Client-Configured Routers
- 19B.2 Navigation Payload Schema
  - 19B.2.1 `NavigationTree` Root Object
  - 19B.2.2 `NavigationGroup` (Module-Level Grouping)
  - 19B.2.3 `NavigationItem` Node
    - Label, Icon, Badge, Tooltip
    - `route_key` — Logical Route Identifier (not a URL)
    - `surface_id` — UI Surface to Render on Activation
    - `permission_required` — Evaluated Server-Side Before Inclusion
    - `feature_flag` — Gate for Early/Beta Items
    - `external_url` — For Cross-System Links
    - `open_in` — `same_tab | new_tab | modal | drawer`
    - `children` — Recursive Sub-Navigation
    - `metadata` — Custom Extension Map
  - 19B.2.4 `NavigationBadge` Schema (Count, Dot, Status)
  - 19B.2.5 Versioning Field on `NavigationTree`
- 19B.3 Backend Construction of Navigation in Go
  - 19B.3.1 `NavigationService` Interface and Implementation
  - 19B.3.2 Module Registration Pattern (`RegisterNavItems()`)
  - 19B.3.3 Permission Pruning During Construction
    - Casbin / CEL Evaluation per Navigation Item
    - RLS-Aware Node Removal (Tenant Scope)
  - 19B.3.4 Feature Flag Evaluation During Construction
  - 19B.3.5 Tenant Navigation Overrides (Custom Items, Re-Ordering)
  - 19B.3.6 Portal Context Switching (Staff vs. Customer vs. Supplier)
  - 19B.3.7 Wire Provider Registration for `NavigationService`
- 19B.4 Navigation API Endpoints
  - 19B.4.1 `GET /api/v1/ui/navigation` — Full Tree for Authenticated User
  - 19B.4.2 Query Parameters: `portal`, `locale`, `surface_hint`
  - 19B.4.3 Response Envelope
  - 19B.4.4 ETag / Cache-Control Strategy
  - 19B.4.5 Navigation Diff Endpoint (`GET /api/v1/ui/navigation/diff?since={version}`)
- 19B.5 Navigation Caching Strategy
  - 19B.5.1 Cache Key: `tenant_id + user_role_hash + portal + locale + flag_hash`
  - 19B.5.2 Redis TTL Policy
  - 19B.5.3 Invalidation Triggers (Permission Change, Flag Change, Tenant Config Change)
  - 19B.5.4 Warm-Up on Tenant Provisioning
- 19B.6 Client Consumption of Navigation Payload
  - 19B.6.1 Fetch-on-Login Pattern
  - 19B.6.2 Client-Side Navigation Store
  - 19B.6.3 Route Key → Client Route Mapping Table
  - 19B.6.4 Handling Unknown `route_key` Values (Forward Compatibility)
  - 19B.6.5 Badge Count Refresh (Polling vs. SSE Push)
  - 19B.6.6 Navigation Re-Fetch on Permission Change
- 19B.7 Navigation for amis Web Renderer
  - 19B.7.1 amis `nav` Component Binding to Navigation Payload
  - 19B.7.2 `app` Component Shell Navigation Wiring
  - 19B.7.3 Active State Derivation from Current Surface
  - 19B.7.4 Collapsed / Expanded State Persistence
- 19B.8 Navigation for Flutter Mobile Renderer
  - 19B.8.1 Bottom Tab Bar Construction from `NavigationTree`
  - 19B.8.2 Drawer Navigation Construction
  - 19B.8.3 Deep Link Routing from `route_key`
- 19B.9 Portal-Specific Navigation Trees
  - 19B.9.1 Staff Portal Navigation (Full ERP Modules)
  - 19B.9.2 Supplier Portal Navigation (Scoped to Procurement)
  - 19B.9.3 Customer Portal Navigation (Scoped to Sales / Statements)
  - 19B.9.4 Employee Self-Service Navigation
  - 19B.9.5 Super-Admin Navigation (Cross-Tenant)
- 19B.10 Navigation and Multi-Tenancy
  - 19B.10.1 Per-Tenant Navigation Customisation Storage
  - 19B.10.2 Tenant Admin Navigation Builder UI
  - 19B.10.3 Module Enablement and Navigation Tree Reduction
- 19B.11 Navigation Audit and Observability
  - 19B.11.1 Navigation Fetch Latency Metrics
  - 19B.11.2 Pruned Item Audit Log
  - 19B.11.3 Navigation Click Telemetry

---

### Chapter 20 — Action System

---

### Chapter 20-B — Real-Time UI Updates *(NEW — was Recommended Addition F)*

> **Audience:** Platform Engineers, Backend Engineers, Web/Mobile Engineers · **Phase:** 2

- 20B.1 Real-Time UI Update Philosophy
- 20B.2 Use Cases
  - 20B.2.1 Live Dashboard Widget Refresh
  - 20B.2.2 Approval Queue Count Badge Updates
  - 20B.2.3 Push-Driven Form Field Updates (e.g. FX Rate Feed)
  - 20B.2.4 Workflow Status Transitions
  - 20B.2.5 Inventory Level Alerts
- 20B.3 Transport Mechanisms
  - 20B.3.1 Server-Sent Events (SSE) — Primary for Web
  - 20B.3.2 WebSocket — For Bidirectional Channels
  - 20B.3.3 Long Polling — Fallback
  - 20B.3.4 Temporal Workflow Signals to UI
- 20B.4 Update Payload Schema
  - 20B.4.1 Targeted Field Update (`field_patch`)
  - 20B.4.2 Full Component Refresh (`component_replace`)
  - 20B.4.3 Data Source Invalidation Signal (`datasource_invalidate`)
  - 20B.4.4 Badge Count Update (`badge_update`)
  - 20B.4.5 Navigation Tree Diff (`nav_diff`)
- 20B.5 Subscription Model
  - 20B.5.1 Surface-Scoped Subscriptions
  - 20B.5.2 Resource-Scoped Subscriptions (e.g. specific PO ID)
  - 20B.5.3 Tenant-Scoped Broadcast
- 20B.6 Authorization for Real-Time Updates
- 20B.7 Backpressure and Rate Limiting
- 20B.8 Reconnection and Missed Update Recovery
- 20B.9 Observability for Real-Time Channels

---

### Chapter 21 — Event System

---

## Volume V — Runtime & State

### Chapter 22 — State Management
### Chapter 23 — Validation Framework
### Chapter 24 — Data Sources

---

### Chapter 24-B — Error Handling Strategy *(NEW — was Recommended Addition B)*

> **Audience:** All Engineers · **Phase:** 3

- 24B.1 Error Handling Philosophy (Server-First, Consistent Surface)
- 24B.2 Error Classification
  - 24B.2.1 Validation Errors
  - 24B.2.2 Business Rule Errors
  - 24B.2.3 Not Found / Resource Errors
  - 24B.2.4 Authorization Errors
  - 24B.2.5 Network / Transport Errors
  - 24B.2.6 Rendering Engine Errors
  - 24B.2.7 Unknown / Unhandled Errors
- 24B.3 Error Response Envelope (from AWO ERP `BusinessError`)
  - 24B.3.1 `code` — Machine-Readable Error Code
  - 24B.3.2 `message` — Localised User-Facing Message
  - 24B.3.3 `field_errors` — Per-Field Validation Map
  - 24B.3.4 `trace_id` — OpenTelemetry Trace Reference
- 24B.4 Client-Side Error Display Patterns
  - 24B.4.1 Inline Field Errors
  - 24B.4.2 Toast / Snackbar Notifications
  - 24B.4.3 Full-Page Error States
  - 24B.4.4 Component-Level Error Boundaries
- 24B.5 Retry Policies
  - 24B.5.1 Automatic Retry for Idempotent Actions
  - 24B.5.2 User-Initiated Retry
  - 24B.5.3 Exponential Back-Off
- 24B.6 Error Fallback UIs in AST
- 24B.7 Error Propagation in Action Chains
- 24B.8 Error Logging and Alerting

---

## Volume VI — Platform & API

### Chapter 25 — API Contracts

---

### Chapter 25-B — API Integration Patterns *(NEW ★)*

> **Audience:** Backend Engineers, Frontend Engineers, Mobile Engineers, Third-Party Integrators · **Phase:** 3

This chapter addresses how AwoERP SDUI surfaces integrate with the AWO ERP REST API layer (Fiber v2, `internal/api/handlers/`), including authentication flow, tenant context propagation, and how UI data sources map to real API endpoints.

- 25B.1 Integration Philosophy
  - 25B.1.1 UI as a Thin Client on Top of the ERP API
  - 25B.1.2 Never Bypassing the API Layer from the UI
  - 25B.1.3 Contract-First API Design with Goa DSL
- 25B.2 Authentication Integration
  - 25B.2.1 JWT Bearer Token Propagation (`Authorization: Bearer …`)
  - 25B.2.2 Token Storage in Client (Secure Storage — Never LocalStorage)
  - 25B.2.3 Token Refresh Flow from UI Actions
  - 25B.2.4 Session Expiry Handling — Redirect vs. Silent Refresh
  - 25B.2.5 M-Pesa STK Push Integration Pattern (Kenya-Specific)
- 25B.3 Tenant Context Propagation
  - 25B.3.1 `X-Tenant-ID` Header — Mandatory on All Requests
  - 25B.3.2 How UI Data Sources Inject `X-Tenant-ID`
  - 25B.3.3 Subdomain-to-Tenant Resolution at the API Gateway
  - 25B.3.4 RLS Enforcement — What the UI Can Rely On
  - 25B.3.5 Tenant Context in amis API Requests (`api` component config)
- 25B.4 API Request Patterns from UI Data Sources
  - 25B.4.1 List Endpoints — Pagination, Filtering, Sorting Parameters
  - 25B.4.2 Detail Endpoints — Route Parameter Binding
  - 25B.4.3 Mutation Endpoints — POST/PATCH/DELETE from Form Submit Actions
  - 25B.4.4 Aggregation / Report Endpoints
  - 25B.4.5 File Upload Endpoints (multipart/form-data)
- 25B.5 API Module Integration Map
  - 25B.5.1 IAM Module (`/api/v1/iam/…`) — Users, Sessions, Roles
  - 25B.5.2 Tenant Module (`/api/v1/tenants/…`)
  - 25B.5.3 Finance Module (`/api/v1/finance/…`) — Accounts, Transactions, Reports
  - 25B.5.4 Audit Module (`/api/v1/audit/…`)
  - 25B.5.5 Navigation Module (`/api/v1/ui/navigation`)
  - 25B.5.6 UI Compilation Service (`/api/v1/ui/surfaces/…`)
  - 25B.5.7 Temporal Workflow Signal Endpoint (`/api/v1/workflows/…`)
- 25B.6 API Permission Guards in UI
  - 25B.6.1 Mapping `finance.accounts.read` to API Endpoint Visibility
  - 25B.6.2 Pre-Flight Permission Check vs. Optimistic Render + 403 Handling
  - 25B.6.3 Row-Level Permission Reflection in UI (OpenFGA Tuple Model)
- 25B.7 API Error Mapping to UI Error States
  - 25B.7.1 HTTP 400 → Field-Level Validation Errors
  - 25B.7.2 HTTP 401 → Re-Authentication Flow
  - 25B.7.3 HTTP 403 → Permission Denied UX
  - 25B.7.4 HTTP 404 → Not Found Surface
  - 25B.7.5 HTTP 409 → Conflict Resolution UX
  - 25B.7.6 HTTP 5xx → System Error Boundary
- 25B.8 amis-Specific API Integration
  - 25B.8.1 amis `api` Adapter Configuration
  - 25B.8.2 Global `env.adaptor` for Header Injection
  - 25B.8.3 `requestAdaptor` / `responseAdaptor` for AWO Envelope Unwrapping
  - 25B.8.4 amis `messages` Locale Binding to AWO Error Codes
- 25B.9 Flutter-Specific API Integration
  - 25B.9.1 Dio HTTP Client Configuration
  - 25B.9.2 Interceptor Chain (Auth, Tenant, Logging)
  - 25B.9.3 Offline Queue Integration with API Client
- 25B.10 Rate Limiting and Retry from UI Layer
- 25B.11 API Contract Testing from the UI Perspective (Chapter 41 cross-ref)

---

### Chapter 26 — Security Model
### Chapter 27 — Authorization Integration
### Chapter 28 — Multi-Tenant Architecture

---

### Chapter 28-B — Platform / Tenant / Portal View Differentiation *(NEW ★)*

> **Audience:** Platform Engineers, Solution Architects, Backend Engineers, Product Owners · **Phase:** 3

This chapter defines how AwoERP serves **fundamentally different UI experiences** to different actor classes — the platform super-admin layer, individual tenant staff, and external portal users — from a single compiled SDUI backend, using the same permission-pruning and tenant-context pipeline.

- 28B.1 The Three-Layer View Model
  - 28B.1.1 Layer 1 — Platform Layer (Anthropic/AWO Operator View)
  - 28B.1.2 Layer 2 — Tenant Layer (Business / Organisation View)
  - 28B.1.3 Layer 3 — Portal Layer (External Actor Views)
  - 28B.1.4 Isolation Guarantees Between Layers
  - 28B.1.5 How RLS and Casbin Enforce Layer Boundaries
- 28B.2 Platform (Super-Admin) View
  - 28B.2.1 Scope: Cross-Tenant Visibility
  - 28B.2.2 Platform Admin Navigation Tree
  - 28B.2.3 Tenant Management Surface
  - 28B.2.4 Global Feature Flag Management
  - 28B.2.5 Platform Audit Log Surface
  - 28B.2.6 Billing and Subscription Management
  - 28B.2.7 Security: Platform Admin Identity Verification (No Tenant RLS)
  - 28B.2.8 Platform Admin amis Surface Schema Example
- 28B.3 Tenant (Staff) View
  - 28B.3.1 Scope: Single-Tenant, Full ERP Module Access (By Role)
  - 28B.3.2 Tenant Admin Sub-View vs. Regular Staff View
  - 28B.3.3 Module Enablement and Navigation Reduction
  - 28B.3.4 Role-Based Surface Differentiation Within a Tenant
    - Finance Manager vs. Finance Clerk vs. Auditor
    - HR Admin vs. Employee (Self-Service)
    - Operations Manager vs. Pump Attendant (Kenya: Fuel Station)
  - 28B.3.5 Tenant Branding Application in Staff View
  - 28B.3.6 Tenant Staff View amis Surface Schema Example
- 28B.4 Portal Views (External Actor Access)
  - 28B.4.1 Portal Architecture Overview
    - Portals Are Tenant-Scoped Sub-Applications
    - Portal Users Are Not Tenant Staff
    - Separate Login / Identity Flow for Portal Users
  - 28B.4.2 Supplier Portal
    - Scope: PO Acknowledgement, Invoice Submission, GRN Confirmation, Statement View
    - Navigation: Minimal — Inbox, POs, Invoices, My Account
    - Permission Model: `portal.supplier.*` CEL-scoped permissions
    - Data Isolation: Supplier can only see their own documents
    - amis Surface Schema Example
  - 28B.4.3 Customer Portal
    - Scope: Sales Order Status, Invoice Payment, Statement Download, Support Tickets
    - Navigation: Orders, Invoices, Payments, Statements, Support
    - Permission Model: `portal.customer.*` CEL-scoped permissions
    - Data Isolation: Customer sees only their account data
    - M-Pesa Payment Integration Surface
    - amis Surface Schema Example
  - 28B.4.4 Employee Self-Service Portal
    - Scope: Leave Applications, Payslip View, Expense Claims, Personal Details
    - Navigation: My Leaves, Payslips, Expenses, Profile
    - Permission Model: `portal.employee.*` CEL-scoped permissions
    - Integration with HR Module
    - amis Surface Schema Example
  - 28B.4.5 (Future) Brand Owner / Franchise Reporting Portal
    - Cross-Tenant Consent-Gated Reporting (Consent Model from v0.1 design)
    - See `DisclosureConsent` pattern
- 28B.5 View Context Resolution in the Compilation Pipeline
  - 28B.5.1 `ViewContext` Object
    - `actor_type`: `platform_admin | tenant_staff | portal_supplier | portal_customer | portal_employee`
    - `tenant_id`
    - `portal_id` (if portal actor)
    - `permissions` (pre-loaded at login per actor type)
    - `feature_flags`
  - 28B.5.2 How `ViewContext` Flows Through Pipeline Stages
  - 28B.5.3 Navigation Tree Selection by `actor_type`
  - 28B.5.4 Surface Access Guard by `actor_type`
- 28B.6 Portal Authentication and Identity
  - 28B.6.1 Separate Auth Endpoint for Portal Users (`/portal/auth/…`)
  - 28B.6.2 Portal JWT Claims (`portal_id`, `actor_type`, `tenant_id`)
  - 28B.6.3 Portal Session Lifetime Policy
  - 28B.6.4 Portal SSO Options
- 28B.7 amis Multi-View Shell Configuration
  - 28B.7.1 Shell Selection by `actor_type` at `index.html` Bootstrap Time
  - 28B.7.2 Portal Shell vs. Staff Shell amis `app` Config
  - 28B.7.3 Theme Token Swap on Shell Init
- 28B.8 Flutter Multi-View Shell Configuration
  - 28B.8.1 Portal-Aware Navigation Tree Consumer
  - 28B.8.2 Deep Link Namespace Separation (`/portal/…` vs. `/app/…`)
- 28B.9 Cross-Portal Concerns
  - 28B.9.1 Shared Component Library (Same Components, Different Themes)
  - 28B.9.2 Localisation Per Portal
  - 28B.9.3 Accessibility Requirements Per Portal
  - 28B.9.4 Analytics Separation Between Portals

---

### Chapter 29 — Feature Flag Architecture
### Chapter 30 — Customization Framework
### Chapter 31 — Extension and Plugin Architecture

---

### Chapter 31-B — Audit and Compliance *(NEW — was Recommended Addition E)*

> **Audience:** Platform Engineers, Security Engineers, ERP Implementers · **Phase:** 3

- 31B.1 Compliance Scope for AwoERP UI Platform
- 31B.2 GDPR Considerations
  - 31B.2.1 PII in UI Payloads — Identification and Minimisation
  - 31B.2.2 Right to Erasure — Impact on UI Definitions and Cached Payloads
  - 31B.2.3 Data Retention for UI Audit Logs
- 31B.3 SOC 2 Controls
  - 31B.3.1 Access Control Evidence via Navigation Pruning Logs
  - 31B.3.2 Audit Trail Completeness
  - 31B.3.3 Change Management for UI Definitions
- 31B.4 Kenya Data Protection Act (DPA 2019) Considerations
  - 31B.4.1 Data Localisation Requirements
  - 31B.4.2 Consent Management in UI
- 31B.5 KRA eTIMS Integration Compliance
  - 31B.5.1 Invoice UI Fields Required by eTIMS
  - 31B.5.2 QR Code Display in Invoice Surface
- 31B.6 UI Audit Log Schema
  - 31B.6.1 UI Surface Access Events
  - 31B.6.2 Action Execution Events
  - 31B.6.3 Navigation Pruning Decision Logs
- 31B.7 Compliance Review Checklist for New UI Surfaces

---

## Volume VII — Reliability & Performance

### Chapter 32 — Performance Optimization
### Chapter 33 — Offline Support
### Chapter 34 — Caching Strategies
### Chapter 35 — Synchronization

---

## Volume VIII — Developer Experience

### Chapter 36 — Internationalization (i18n)
### Chapter 37 — Accessibility (a11y)
### Chapter 38 — Observability
### Chapter 39 — Telemetry
### Chapter 40 — Logging

---

### Chapter 40-B — Developer Tooling *(NEW — was Recommended Addition C)*

> **Audience:** All Engineers · **Phase:** 4

- 40B.1 Developer Tooling Philosophy
- 40B.2 Local Development Environment Setup
  - 40B.2.1 Prerequisites (Go, Node, Flutter, Docker)
  - 40B.2.2 `make dev-ui` — Spin Up Full SDUI Stack Locally
  - 40B.2.3 Hot Reload for UI Surface Definitions
- 40B.3 CLI Tools
  - 40B.3.1 `awo ui validate` — Validate a JSON Surface Definition
  - 40B.3.2 `awo ui compile` — Compile an AST for a Given Context (Dry Run)
  - 40B.3.3 `awo ui diff` — Diff Two Surface Definition Versions
  - 40B.3.4 `awo ui render` — Render to HTML Snapshot for Review
  - 40B.3.5 `awo nav tree` — Print Navigation Tree for a Given Actor Context
- 40B.4 Local Dev Emulator
  - 40B.4.1 Mock Tenant Context Injection
  - 40B.4.2 Mock Permission Sets
  - 40B.4.3 Feature Flag Override for Local Dev
- 40B.5 JSON Schema Browser
  - 40B.5.1 Schema Browser Web UI
  - 40B.5.2 Component Catalogue with Live Preview
- 40B.6 amis Sandbox Integration (See Chapter 54)
- 40B.7 VS Code Extension
  - 40B.7.1 Schema Autocomplete for Surface JSON
  - 40B.7.2 Inline Component Documentation
  - 40B.7.3 Permission String Autocomplete

---

## Volume IX — Engineering Excellence

### Chapter 41 — Testing Strategy
### Chapter 42 — CI/CD Considerations
### Chapter 43 — Versioning Strategy
### Chapter 44 — Migration Strategy
### Chapter 45 — Governance Model

---

## Volume X — Reference & Guidance

### Chapter 46 — Best Practices
### Chapter 47 — Anti-Patterns
### Chapter 48 — Reference Architecture
### Chapter 49 — End-to-End Examples
### Chapter 50 — SDK Development

---

### Chapter 50-B — Onboarding and Migration Guide for Implementers *(NEW — was Recommended Addition I)*

> **Audience:** ERP Implementers, Solution Architects · **Phase:** 4

- 50B.1 Onboarding Overview
- 50B.2 Pre-Onboarding Checklist
- 50B.3 Tenant Provisioning Steps
  - 50B.3.1 Tenant Creation via Platform Admin Surface
  - 50B.3.2 Module Enablement
  - 50B.3.3 Navigation Configuration
  - 50B.3.4 Theme and Branding Upload
  - 50B.3.5 Portal Activation (Supplier / Customer / Employee)
- 50B.4 Data Migration for UI Customisations
- 50B.5 User Onboarding Flows
- 50B.6 Migration from Legacy ERP UI
  - 50B.6.1 Field Mapping from Legacy to AwoERP Components
  - 50B.6.2 Print Template Migration
- 50B.7 Go-Live Checklist

---

### Chapter 50-C — Playground and Sandbox Environment *(NEW — was Recommended Addition J)*

> **Audience:** All Engineers, ERP Implementers · **Phase:** 4

- 50C.1 Playground Philosophy
- 50C.2 Playground Architecture
  - 50C.2.1 Sandboxed Tenant Context
  - 50C.2.2 Live amis JSON Editor with Preview Pane
  - 50C.2.3 Permission Simulator
  - 50C.2.4 Feature Flag Toggle Panel
- 50C.3 Shareable Playground Links
- 50C.4 Playground Limitations and Security Boundaries

---

### Chapter 51 — Future Roadmap

---

## Volume XI — amis Web Renderer *(NEW VOLUME ★)*

*Covers the integration of the Baidu amis low-code framework as AwoERP's primary web rendering engine — from the bootstrap HTML file through runtime configuration, API wiring, and multi-view shell setup.*

---

### Chapter 52 — amis Web Renderer — Bootstrap, Shell, and index.html

> **Audience:** Web Engineers, Platform Engineers, UI Framework Developers · **Phase:** 4

#### 52.1 amis in the AwoERP Architecture

- 52.1.1 Role of amis as the Web Rendering Engine
- 52.1.2 amis vs. Custom React Renderer — Why amis Was Chosen
- 52.1.3 amis Constraint Model (What AwoERP Does Not Expose)
- 52.1.4 amis Version Pinning and Upgrade Policy
- 52.1.5 amis `cxd` Theme as the Design Base

#### 52.2 What Goes in `index.html` — The Complete Reference

The `index.html` is the **single entry point** for the amis web renderer. Every decision made here affects all tenants, all portals, and all surfaces. This section is the authoritative reference for its contents.

- 52.2.1 Document Structure and `<meta>` Requirements
  - `charset`, `viewport`, `X-UA-Compatible`
  - `<title>` — Tenant-Aware Dynamic Title Strategy
  - `<meta name="theme-color">` — Portal-Driven Value
  - CSP `<meta>` — Permitted Sources for amis CDN Assets

- 52.2.2 amis Asset Loading
  - CDN vs. Self-Hosted amis Bundle Decision
  - Required CSS Files
    - `amis/sdk/sdk.css`
    - `amis/sdk/helper.css`
    - `amis/sdk/iconfont.css`
    - AwoERP Global Override CSS (`awo-overrides.css`)
  - Required JS Files
    - `amis/sdk/sdk.js` — Core amis SDK
  - Asset Integrity Hashes and Subresource Integrity (SRI)
  - Load Order Requirements (CSS before JS)

- 52.2.3 Tenant and Portal Bootstrap Data
  - Inline `<script>` Block — `window.__AWO_BOOTSTRAP__`
    - `tenantId` — Injected by SSR or gateway
    - `portalType` — `staff | supplier | customer | employee | platform_admin`
    - `locale` — Default locale for this session
    - `themeTokens` — Tenant-specific CSS variable overrides (if any)
    - `featureFlags` — Resolved flag map for this session
    - `apiBaseUrl` — Backend API base URL
    - `navigationEndpoint` — `/api/v1/ui/navigation`
    - `authConfig` — Token storage key, refresh endpoint, session timeout
  - Security Considerations: Never inject sensitive data into `__AWO_BOOTSTRAP__`
  - SSR Bootstrap Injection Pattern (Go template rendering `index.html`)

- 52.2.4 amis SDK Initialisation Block
  ```
  <script>
    // Pattern: amis.embed() call
    // - Target mount node (#app)
    // - Initial page schema (loading skeleton or home surface JSON)
    // - env configuration (requestAdaptor, responseAdaptor, theme, locale)
    // - fetcher override for AWO tenant header injection
    // - notify/alert/confirm overrides for AWO design system
    // - session expiry handler
  </script>
  ```
  - `amis.embed()` — Parameters Reference
  - `env` Object — Full Field Reference
    - `theme` — Always `'cxd'` for AwoERP
    - `locale` — From `__AWO_BOOTSTRAP__.locale`
    - `fetcher` — Custom fetch wrapper (adds `Authorization`, `X-Tenant-ID`)
    - `requestAdaptor` — Outbound request normalisation
    - `responseAdaptor` — AWO envelope unwrapping (`{ data, meta, error }`)
    - `notify` — Wired to AWO Toast component
    - `alert` — Wired to AWO Alert Dialog
    - `confirm` — Wired to AWO Confirm Dialog
    - `jumpTo` — SPA navigation handler
    - `isCancel` — Axios cancel token check (if applicable)
    - `copy` — Clipboard handler
    - `getModalContainer` — Portal DOM node for modals
  - Initial Page Schema Bootstrap Strategies
    - Strategy A: Fetch navigation + first surface on load
    - Strategy B: Inline loading skeleton schema, then hydrate
    - Strategy C: SSR-injected first surface (Go template)

- 52.2.5 Dynamic Theme Application
  - Applying `themeTokens` from `__AWO_BOOTSTRAP__` to `:root` CSS variables
  - `awo-overrides.css` — Permitted Override Surface (Two CSS Variable Rule)
  - Dark Mode Init (Check `prefers-color-scheme` vs. stored user preference)

- 52.2.6 Mount Node and Loading State
  - `<div id="app">` — The amis Mount Point
  - Pre-amis Loading Skeleton (Inline CSS Spinner)
  - `<noscript>` Fallback Message

- 52.2.7 Error Boundary at Bootstrap Time
  - Graceful Degradation if amis SDK Fails to Load
  - Network Error Detection before `amis.embed()`
  - Auth Failure at Init (Token Missing / Expired on Page Load)

- 52.2.8 Security Hardening in `index.html`
  - Content Security Policy Headers (HTTP layer vs. `<meta>` fallback)
  - No Inline Event Handlers (`onclick=...` forbidden)
  - `Referrer-Policy` and `X-Frame-Options` via HTTP headers
  - Clickjacking Protection

- 52.2.9 Multi-Portal `index.html` Strategy
  - Option A: Single `index.html` with Portal Detection at Runtime
  - Option B: Per-Portal `index.html` (staff/index.html, supplier/index.html, …)
  - Recommended Approach for AwoERP + Trade-offs
  - Nginx / Caddy Routing Configuration

- 52.2.10 `index.html` for Development vs. Production
  - Dev: Vite Proxy + Hot Reload Config
  - Production: Static File Serving via Go `embed.FS` or CDN
  - Cache-Busting Strategy for amis Assets

#### 52.3 amis `app` Component — The Application Shell

- 52.3.1 `app` Schema Overview
- 52.3.2 Navigation Population from Backend Navigation API
  - `pages` Array Construction from `NavigationTree` Response
  - Static Fallback Schema for Offline / Error States
- 52.3.3 Header / Toolbar Configuration
  - Logo (Tenant-Specific)
  - Global Search
  - Notification Bell (Badge from SSE)
  - User Menu (Profile, Settings, Logout)
  - Tenant Switcher (Platform Admin Only)
- 52.3.4 Sidebar Navigation
  - Collapsed vs. Expanded State Persistence
  - Permission-Pruned Items (Already Removed Server-Side)
  - Active Item Derivation from Current Route
- 52.3.5 Content Area — Dynamic Surface Loading
  - Surface Fetch Pattern (`GET /api/v1/ui/surfaces/{surface_id}`)
  - Loading State Skeleton
  - Error Boundary Within Content Area
- 52.3.6 Footer Band (Optional)
- 52.3.7 Portal Shell Variants
  - Staff Shell Schema (Full Sidebar + Header)
  - Supplier / Customer Portal Shell (Minimal Header, No Sidebar)
  - Employee Self-Service Shell

#### 52.4 amis Renderer Runtime Lifecycle

- 52.4.1 Bootstrap → Navigation Fetch → First Surface Render
- 52.4.2 Surface Navigation (SPA-style Route Changes)
- 52.4.3 Action Execution Lifecycle in amis
- 52.4.4 Form Submit → API Call → Response Handling → Next Action
- 52.4.5 SSE Channel Subscription on Mount
- 52.4.6 Session Expiry Detection and Re-Auth Flow
- 52.4.7 Renderer Teardown and Cleanup

#### 52.5 amis Component Governance for AwoERP

- 52.5.1 Approved amis Component List (Cross-Ref with Chapter 09)
- 52.5.2 Banned amis Components and Why
- 52.5.3 Wrapper Conventions for amis Components
- 52.5.4 Custom amis Component Registration (`amis.registerRendererPlugin`)
- 52.5.5 60-Line Screen File Limit Rule (From UI DSL Architecture)
- 52.5.6 No `map[string]any` in amis Schema Go Code (CI Guard)

#### 52.6 amis Performance Considerations

- 52.6.1 amis SDK Bundle Size and Lazy Loading
- 52.6.2 Page Schema Payload Size Budget
- 52.6.3 Virtual Scrolling in amis CRUD Tables
- 52.6.4 Schema Caching (Redis → HTTP Cache Headers → Browser)

#### 52.7 amis Testing

- 52.7.1 Schema Snapshot Tests
- 52.7.2 Component Rendering Tests (Playwright)
- 52.7.3 amis Action Execution Tests

---

### Appendix 52-A — `index.html` Annotated Reference Template

> A complete, annotated `index.html` template for production AwoERP amis deployment, with inline comments explaining every decision.

```
<!-- Structure:
  1. <!DOCTYPE html> and <html lang="{locale}">
  2. <head>
       - charset, viewport, theme-color
       - CSP meta (dev) / HTTP header (prod)
       - amis CSS (sdk.css, helper.css, iconfont.css)
       - awo-overrides.css
       - Preconnect hints for API and CDN
  3. <body>
       - #app mount node with loading skeleton
       - <noscript> fallback
       - window.__AWO_BOOTSTRAP__ inline script
       - amis sdk.js
       - amis init script (amis.embed)
       - Dark mode init script
-->
```

*(Full annotated template with all decisions documented — generated as companion file `52-appendix-amis-index-html-reference.md`)*

---

## Summary Statistics

| Metric | v0.1.0 | v0.2.0 |
|--------|--------|--------|
| Volumes | 10 | **11** |
| Core Chapters | 51 | **58** |
| New Chapters (this revision) | — | **7 (19B, 25B, 28B, 52 + 3 promoted)** |
| Promoted Recommended Additions | — | **10 (all B–J promoted to numbered chapters)** |
| Top-Level Sections | ~200+ | **~280+** |
| Estimated Pages (when written) | 800–1,200 | **1,100–1,600** |
| Documentation Phases | 4 | 4 |
| Phase 1 Chapters | 8 | **9** |
| Phase 2 Chapters | 13 | **17** |
| Phase 3 Chapters | 14 | **18** |
| Phase 4 Chapters | 16 | **14** |

---

## New Chapter Summary (v0.2.0 Additions)

| Chapter | Title | Volume | Phase | Why Added |
|---------|-------|--------|-------|-----------|
| **07-B** | UI Composition Patterns | II | 1 | Template/slot/mixin reuse mechanism; was Rec. D |
| **10-B** | Theming and Design Tokens | III | 2 | Full token system, tenant + portal branding; was Rec. H |
| **12-B** | Print and Export Rendering | III | 2 | ERP invoices, picking slips, KRA eTIMS QR; was Rec. G |
| **19-B** | **Backend-Driven Navigation** | IV | 2 | Go service owns full nav tree; permission pruning; portal switching ★ |
| **20-B** | Real-Time UI Updates | IV | 2 | Live dashboards, badge counts, FX feeds; was Rec. F |
| **24-B** | Error Handling Strategy | V | 3 | Standard error envelope, retry, fallback UIs; was Rec. B |
| **25-B** | **API Integration Patterns** | VI | 3 | Fiber v2 endpoint map, M-Pesa, tenant header, amis `env.fetcher` ★ |
| **28-B** | **Platform / Tenant / Portal View Differentiation** | VI | 3 | Three-layer view model; supplier/customer/employee portals ★ |
| **31-B** | Audit and Compliance | VI | 3 | GDPR, SOC2, KRA eTIMS, Kenya DPA; was Rec. E |
| **40-B** | Developer Tooling | VIII | 4 | CLI, emulator, schema browser, VS Code ext; was Rec. C |
| **50-B** | Onboarding and Migration Guide | X | 4 | Tenant provisioning step-by-step; was Rec. I |
| **50-C** | Playground / Sandbox Environment | X | 4 | Live amis JSON editor + permission sim; was Rec. J |
| **52** | **amis Web Renderer — Bootstrap & index.html** | XI | 4 | Complete `index.html` reference, `amis.embed()`, shell variants ★ |

★ = Entirely new (not previously in Recommended Additions)

---

*End of Proposed Table of Contents v0.2.0 — Awaiting Review and Approval*
