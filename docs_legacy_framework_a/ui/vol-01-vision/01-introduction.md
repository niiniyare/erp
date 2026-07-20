> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Chapter 01 — Introduction to AwoERP UI Platform

> **Volume:** I — Vision & Philosophy
> **Audience:** All Engineers, Product Owners, Solution Architects, ERP Implementers
> **Prerequisites:** None

---

## Table of Contents

- [1.1 What Is AwoERP?](#11-what-is-awoerp)
- [1.2 The Problem This Platform Solves](#12-the-problem-this-platform-solves)
- [1.3 Platform Mission Statement](#13-platform-mission-statement)
- [1.4 Key Value Propositions](#14-key-value-propositions)
- [1.5 Current State and Roadmap](#15-current-state-and-roadmap)
- [1.6 Who Should Read This Documentation](#16-who-should-read-this-documentation)
- [1.7 Glossary Reference](#17-glossary-reference)

---

## 1.1 What Is AwoERP?

AwoERP is a **multi-tenant Enterprise Resource Planning (ERP) platform** designed for organizations that operate across finance, procurement, inventory, human resources, and sales within a unified, cloud-native system.

At its technical core, AwoERP uses a **Server-Driven UI (SDUI)** architecture. The backend is the single, authoritative source that defines what users see, how they interact with the application, and what they are permitted to do. The user interface is composed on the server as a structured JSON document — specifically an **AMIS schema** — and the browser renders it using the AMIS SDK.

The backend does not serve pre-built HTML or React component trees. It compiles a page description at request time, incorporating the requesting user's authorization context, tenant configuration, and active feature flags, and delivers a JSON payload that the browser's AMIS renderer materializes into a live interface.

### 1.1.1 What AMIS Is

[AMIS](https://aisuda.bce.baidu.com/amis/en-US/docs) is an open-source, JSON-driven UI framework from Baidu. It renders complex UI — data tables, forms, charts, dashboards, dialog flows — from JSON schemas without custom frontend code. AwoERP generates these schemas dynamically on the backend using Go, applies authorization and tenant context, caches the result in Redis, and delivers it to the browser which renders it using the AMIS web SDK.

This is not a template-and-hydrate model. The schema is computed, not retrieved from static storage. Different users logged into the same page see different schemas if their permissions, feature flags, or tenant configuration differ.

---

## 1.2 The Problem This Platform Solves

### 1.2.1 Authorization Logic Duplication

In a conventional ERP, authorization is enforced at the API layer. The frontend independently decides which buttons to show, which fields to display, which menu items to render — re-implementing rules that already live in the backend. Every re-implementation is a divergence risk: field added to the API but still showing in the frontend for unauthorized users, or hidden from authorized users because the frontend check was not updated.

AwoERP moves this decision to the compilation pipeline. By the time a schema reaches the browser, it has already been filtered for the requesting user's permissions. The browser renders exactly what it receives. There are no hidden elements to reveal, no access decisions left to client-side JavaScript.

### 1.2.2 Multi-Platform Maintenance

Building separate UI codebases for web, iOS, and Android triples the effort for every feature and every permission change. Cross-platform frameworks reduce the platform count but do not eliminate the fundamental problem: business logic about what users should see still lives in the backend, and the frontend re-implements it.

AwoERP's response is to move the UI description itself to the backend. The frontend becomes a rendering engine: it receives a schema and materializes it. Adding a new screen or changing a field layout requires only a backend change; no frontend sprint, no app store release.

### 1.2.3 Tenant Customization

A multi-tenant ERP serves customers with genuinely different requirements. The platform must serve this diversity without code forks. Because AwoERP composes schemas dynamically at request time — resolving tenant context, user permissions, and feature flags — different tenants receive different experiences from the same underlying business logic, without code duplication.

---

## 1.3 Platform Mission Statement

> **AwoERP's UI platform ensures that every user, in every tenant, always experiences a UI that is consistent with the business rules, permissions, and configurations that govern their context — with no frontend code required to enforce or maintain that consistency.**

Three practical implications:

**For engineers:** The backend is where UI intent is expressed. Frontend work is building a capable rendering engine, not re-encoding authorization decisions.

**For product teams:** Adding a screen or modifying a workflow is a backend concern. It does not require a frontend sprint or a mobile release.

**For enterprise customers:** Customization is isolated to the tenant, version-controlled, and does not prevent them from receiving platform upgrades.

---

## 1.4 Key Value Propositions

### 1.4.1 Permission-Aware UI at the Source

The UI compilation pipeline integrates with Casbin before the schema leaves the backend. Fields the user cannot see are absent from the payload. Actions the user cannot execute do not appear. The frontend never receives a schema containing elements it should hide.

### 1.4.2 Cached and Efficient

The pipeline caches compiled schemas in Redis, keyed by route + tenant + permission fingerprint + feature flag fingerprint. A cache hit reduces compilation latency from ~20–50ms to ~2–5ms. Users with the same permission set within a tenant share cache entries.

### 1.4.3 Business Logic Drives UI

Go backend engineers define page schemas in the same codebase as the business logic that validates form submissions. The form and its validator are written by the same engineer, versioned together. There is no separate UI configuration layer maintained by a different team.

### 1.4.4 ERP-Native Component Vocabulary

The DSL provides blocks designed for ERP: `DataTableBlock`, `StatusBadgeBlock`, `DetailCardBlock`, `LineItemsBlock`, `ApprovalBlock`, `QuickActionsBlock`. Backend engineers compose pages from ERP vocabulary, not from generic HTML primitives.

---

## 1.5 Current State and Roadmap

### 1.5.1 What Is Implemented Today

- AMIS web rendering via `web/sdk/` (sdk.js, sdk.css)
- 9-stage compilation pipeline: Session → Authz → Cache → Registry → Compile → Normalize → Validate → CacheStore → Response
- Casbin-based permission resolution via `UIAuthzService.BulkEnforce`
- Redis schema caching with permission + feature flag fingerprints
- Page registry (`internal/web/registry/`) with exact and parameterized route matching
- AST package (`internal/web/ast/`) with typed node types
- AMIS builder package (`internal/web/amis/`) for legacy page functions
- DSL blocks: `DataTableBlock`, `StatusBadgeBlock`, `DetailCardBlock`, `LineItemsBlock`, `ApprovalBlock`, `QuickActionsBlock`
- DSL screens: `FinanceDashboardScreen`, `InvoiceScreen`, `JournalEntryScreen`, `TrialBalanceScreen`, `RegisterScreen`
- Registered routes: `/finance/dashboard`, `/finance/invoices/new`, `/finance/bills/new`, `/finance/journal-entries/new`, `/finance/reports/trial-balance`

### 1.5.2 What Is Planned (Not Yet Implemented)

> **Status: Planned — Not Yet Implemented**

The following capabilities are planned on the roadmap but have no production code today:

- Flutter mobile client
- React Native mobile client
- CLI / terminal client
- gRPC streaming for real-time UI updates
- Client capability detection (iOS vs Android vs web)
- Schema versioning and ETags
- Webhook-driven cache invalidation from business events
- Tenant form customization overlays (no-code)
- Plugin and extension architecture
- Brotli/gzip compression in the pipeline
- Parameterized routes for edit screens (`/finance/invoices/:id` — registry supports the pattern but no screens are registered yet)

---

## 1.6 Who Should Read This Documentation

| Role | Recommended Reading |
|------|-------------------|
| Platform engineers | All volumes |
| Backend (Go) engineers | Vol I–II (vision + DSL), Vol III (component system) |
| Web engineers | Vol I §1.1 (AMIS overview), Ch 03 (pipeline), Vol III (components) |
| Product owners | Vol I only |
| ERP implementers | Vol I + Vol III (component vocabulary) |

---

## 1.7 Glossary Reference

| Term | Working Definition |
|------|-------------------|
| **SDUI** | Server-Driven UI. The backend defines what the client renders. |
| **AMIS schema** | A JSON document (map[string]any) describing an AMIS page, form, table, or chart. The wire format consumed by the AMIS browser SDK. |
| **AST** | Abstract Syntax Tree. The typed, in-memory Go object graph that is compiled to an AMIS schema by `ast.CompileTree()`. |
| **DSL** | Domain-Specific Language. The Go vocabulary of blocks, screens, and builders used to compose pages. |
| **Route** | The URL path segment after `/schema`, e.g., `/finance/dashboard`. The registry maps routes to `PageFn` or `ASTPageFn`. |
| **PageFn** | `func(sess UISessionContext) Schema` — the legacy page builder signature. Returns a raw `map[string]any`. |
| **ASTPageFn** | `func(sess UISessionContext) any` — the preferred page builder. Returns an `ast.Node`; `CompileTree` validates and serializes it. |
| **UISessionContext** | The read-only, pre-resolved view of session + permissions + feature flags passed into every page function. |
| **Pipeline** | The 9-stage Go pipeline that transforms a route request into a compiled AMIS schema. |
| **Tenant** | A distinct organizational customer with isolated data, configuration, permissions, and UI. |
| **Feature Flag** | A runtime boolean that controls which variant of a UI definition a given user or tenant receives. |
| **Casbin** | The authorization engine used by `UIAuthzService.BulkEnforce` to resolve permissions before page compilation. |

---

*End of Chapter 01*

**Next:** [Chapter 02 — Design Philosophy](./02-design-philosophy.md)
