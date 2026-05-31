# Awoerp Custom Plugin System
## Comprehensive Technical & Product Documentation — v1.0

> **Audience:** Engineering teams, product managers, system architects, third-party developers, and business stakeholders.
> **Status:** Pre-release Reference Document
> **Last Updated:** 2026

---

## Table of Contents

1. [Introduction](#1-introduction)
2. [The Case for a Plugin System](#2-the-case-for-a-plugin-system)
3. [Plugin System Architecture Overview](#3-plugin-system-architecture-overview)
4. [Plugin Types & Extension Points](#4-plugin-types--extension-points)
5. [The Plugin Contract & SDK](#5-the-plugin-contract--sdk)
6. [Technology Stack Deep Dive & Decision Rationale](#6-technology-stack-deep-dive--decision-rationale)
7. [Plugin Registry & Discovery](#7-plugin-registry--discovery)
8. [Security Model](#8-security-model)
9. [Plugin Development Guide](#9-plugin-development-guide)
10. [Core System Integration Guide](#10-core-system-integration-guide)
11. [Deployment & Operations](#11-deployment--operations)
12. [Comparison to Existing Plugin Systems](#12-comparison-to-existing-plugin-systems)
13. [v1.0 Scope, Deferred Features & Roadmap](#13-v10-scope-deferred-features--roadmap)
14. [Appendices](#14-appendices)

---

---

# 1. Introduction

## 1.1 What Is the Awoerp Plugin System?

Awoerp is a modern, open-extensible Enterprise Resource Planning (ERP) platform built on a Go-native backend stack. At its core, Awoerp manages the operational backbone of a business — finance, inventory, HR, procurement, and more. But no single ERP, however well-designed, can anticipate every industry vertical, every regulatory requirement, or every workflow variation that its users will eventually need.

The **Awoerp Plugin System** is the architectural answer to that reality. It is a formally defined, versioned, and secured framework that allows **external developers, system integrators, and businesses themselves** to extend Awoerp's features without modifying — or even accessing — the core source code. Think of the core system as a city's road infrastructure: it provides the roads, traffic rules, and utilities. Plugins are the buildings, shops, and services that rise on that infrastructure, each following the city's building codes but free to serve entirely different purposes.

More precisely, the plugin system provides:

- A **stable Go interface contract** that plugins must implement, ensuring compatibility across Awoerp versions.
- **Extension points** across every layer of the application: HTTP routes, database schemas, background workflows, UI screens, and event hooks.
- A **registry and discovery mechanism** for finding, installing, and versioning plugins.
- A **security and sandboxing model** that ensures third-party code cannot compromise tenant data or system integrity.
- An **SDK and toolchain** that makes plugin authorship as friction-free as possible.

This document covers all of that — from the 30,000-foot strategic rationale to line-level Go interface definitions.

---

## 1.2 Who This Document Is For

This document is intentionally written for **two audiences simultaneously**:

| Audience | What They Need From This Document |
|---|---|
| **Non-technical stakeholders** (product managers, business owners, investors) | Understand *why* the plugin system exists, what business value it creates, and what risks arise if it is not built or deferred. |
| **Technical stakeholders** (engineers, architects, DevOps, plugin authors) | Understand *how* the plugin system works, including interface definitions, technology choices, security guarantees, and operational concerns. |

Sections marked with a 🏢 icon are primarily strategic and business-focused. Sections marked with ⚙️ are primarily technical. Sections with no icon serve both audiences.

When deeply technical content appears within a primarily strategic section (or vice versa), it is clearly marked with a callout box so readers can skip or dive deeper according to their needs.

---

## 1.3 Document Conventions & Reading Guide

The following conventions are used throughout this document:

```
Code blocks contain Go source code, YAML, JSON, or shell commands.
```

> **💡 Insight:** Callout boxes highlight important concepts that deserve extra attention.

> **⚠️ Warning:** Warning callouts indicate failure modes, security concerns, or anti-patterns.

> **🏢 Business Note:** Strategic or non-technical context for decision-makers.

> **⚙️ Technical Note:** Implementation detail that engineers should not skip.

**Bold text** introduces key terms on first use. A full glossary is provided in Appendix D.

---

## 1.4 Vision: Why Extensibility Is a First-Class Citizen

> **🏢 Business Note**

The history of enterprise software is littered with systems that were powerful but closed. Systems that required expensive professional services engagements for every small customization. Systems that locked businesses into a single vendor's roadmap. Businesses either accepted that constraint, hired armies of internal developers to fork the codebase, or eventually migrated to something more flexible — at enormous cost.

The most successful software platforms of the last decade — Shopify, Salesforce, Atlassian, Stripe — all share a common architectural philosophy: **the platform provides the foundation; the ecosystem provides the breadth**. Shopify's real moat is not its checkout flow — it is the ten thousand apps in its marketplace that make it adaptable to any kind of merchant. Salesforce's value is not its CRM — it is AppExchange and the billions of dollars of complementary software built on its APIs.

Awoerp is built with this philosophy from day one. The plugin system is not a feature added to an existing product — it is an **architectural constraint** that shapes how every part of the core system is designed. Every time a core developer asks "should I build this inside the core or as a plugin?", the plugin system wins unless there is a compelling reason otherwise. This keeps the core lean, focused, and fast, while the ecosystem carries the long tail of needs.

The goal is explicit: **Awoerp should be the last ERP a business ever needs** — not because it ships everything, but because anything it doesn't ship can be built on top of it.

---

---

# 2. The Case for a Plugin System

## 2.1 The Problem with Monolithic ERP Systems

> **🏢 Business Note**

Traditional ERP systems — whether legacy giants like SAP and Oracle, or newer mid-market players — tend to evolve in one of two directions. Either they grow into massive monoliths where every customer-requested feature is added to the core product (resulting in a system so bloated and complex that even basic deployments require months of configuration), or they produce a rigid, opinionated system where businesses must change their processes to match the software rather than the other way around.

Both outcomes are symptoms of the same root problem: **the system has no principled, low-friction mechanism for extending itself**. Every new requirement either becomes a core feature (adding complexity for everyone) or a custom fork (creating a maintenance nightmare for the customer and excluding them from future upgrades).

The real-world consequences of this are well-documented:

- SAP implementations routinely take 12–24 months and cost millions of dollars, largely because of customization complexity.
- ERPNext, while open source and extensible via its Frappe framework, requires Python knowledge and direct server access to build and deploy apps — a barrier that excludes non-engineering businesses.
- Odoo's module system is powerful but tightly coupled to Odoo's own ORM and view engine, meaning plugins cannot easily be written in other languages or deployed as independent services.

Awoerp's plugin system is designed to avoid all three of these failure modes: it keeps the core simple, gives businesses a standard and secure way to extend the system themselves, and allows plugins to be written by anyone with Go knowledge — not just Awoerp core contributors.

---

## 2.2 What "Public Extensibility" Means in Practice

The word "public" in the plugin system's scope is deliberate and important. It means three distinct things:

**1. Publicly Accessible Extension Points**
The plugin system exposes well-documented, stable interfaces that are part of Awoerp's public API contract. These interfaces do not change without versioning and a documented deprecation period. Any developer in the world — not just Awoerp employees — can build against these interfaces with confidence that their plugin will continue to work across Awoerp releases.

**2. Publicly Distributable Plugins**
Plugins built on the Awoerp Plugin SDK can be packaged, versioned, and distributed either through the official Awoerp Plugin Registry or through private/enterprise registries. A third-party logistics company could build an Awoerp plugin for their specific carrier integrations, sell it on the registry, and have any Awoerp customer install it in minutes — just like installing a mobile app.

**3. No Source Code Access Required**
Plugin authors do not need access to Awoerp's core source code, internal databases, or infrastructure secrets. All interaction between a plugin and the core system happens through the defined plugin interfaces and the Core Services SDK. This is a security and commercial boundary: it means Awoerp can maintain a commercial open-core model where the core is open source, but certain premium features or the registry itself may be commercial.

---

## 2.3 Value Delivered to Stakeholders

Different stakeholders experience the value of the plugin system differently. The table below maps each stakeholder group to the concrete value they receive:

| Stakeholder | Value Delivered |
|---|---|
| **End-user businesses** | Ability to tailor Awoerp to their exact industry, workflow, and regulatory context without forking or waiting for the core roadmap. |
| **System integrators / consultants** | A repeatable, standardized way to build and deliver customizations — replacing one-off fork-and-patch work with productized plugins they can resell. |
| **Independent software vendors (ISVs)** | A new distribution channel: build once, sell to all Awoerp users. Lower go-to-market cost than building a standalone ERP module from scratch. |
| **Awoerp core team** | The core roadmap remains focused on platform stability and core ERP primitives. Community innovation handles the long tail. |
| **Enterprise IT departments** | Ability to write internal plugins for company-specific integrations (LDAP, bespoke reporting, legacy system bridges) without maintaining a full fork. |

---

## 2.4 Risk of NOT Shipping a Plugin System in v1.0

> **🏢 Business Note — This section is critical for product and leadership teams.**

This is perhaps the most important section in this document for decision-makers. The temptation when building v1.0 of any product is to defer extensibility infrastructure: "We'll add it later when we have more users." This reasoning has a seductive logic — why build infrastructure for plugins before you have plugin authors? The following analysis explains why that logic is deeply flawed for a product like Awoerp.

### 2.4.1 Technical Debt Accumulation

Every feature built into the Awoerp core **before** a plugin system exists is built without the discipline of the plugin contract. Core developers take shortcuts that would be forbidden under the plugin model: direct database access across module boundaries, tight coupling between UI components and backend services, hardcoded assumption about what business logic exists.

When the plugin system is added later, these shortcuts must be untangled. The cost is not linear — it is exponential. Refactoring a system that has 50,000 lines of tightly-coupled code to support plugin isolation takes significantly more effort than building with isolation in mind from line one.

The analogy: it is cheaper to install plumbing during construction than to retrofit it into a finished building.

### 2.4.2 Market Positioning Consequences

The ERP market is highly competitive. Early adopters of Awoerp will be businesses that are technology-forward enough to choose a new platform over an established incumbent. These are exactly the businesses most likely to evaluate extensibility as a selection criterion. If Awoerp ships v1.0 without a plugin system:

- System integrators will not build practices around Awoerp because there is no repeatable customization model.
- ISVs will not invest in building Awoerp integrations because there is no stable platform to build on.
- Enterprise buyers will not sign contracts if their IT team cannot extend the system for their specific needs.

Competitors who ship with extensibility — even a basic one — will win these deals. **Extensibility is a prerequisite for ecosystem formation, and ecosystem formation is a prerequisite for platform defensibility.**

### 2.4.3 Migration Pain for Adopters

Early adopters who deploy Awoerp v1.0 without a plugin system will inevitably begin customizing it anyway — by forking, by patching, by writing scripts around the edges. These informal extension patterns create a "shadow plugin system" that is undocumented, unsecured, and unsupported.

When the official plugin system ships in v1.5 or v2.0, these businesses face a painful migration: their existing customizations must be rewritten to conform to the new plugin model. Some will defer the migration indefinitely, creating a permanently fragmented install base. Others will be stuck on old versions. Both outcomes harm Awoerp's reputation and increase support burden.

**The cost of deferring the plugin system is paid by early adopters, not by the Awoerp team — and those adopters will remember.**

> **⚙️ Technical Note:** From an engineering standpoint, the cost of retrofitting a plugin-aware architecture into an existing Go codebase is primarily in: (a) introducing interface boundaries where direct function calls exist today, (b) establishing database schema namespacing where shared migrations exist today, and (c) introducing process or goroutine isolation where in-process function calls exist today. Each of these is significantly cheaper to do at greenfield than at retrofit stage.

---

---

# 3. Plugin System Architecture Overview

## 3.1 High-Level Architecture Diagram

The following diagram describes the relationship between the Awoerp core, the plugin runtime, individual plugins, and the supporting infrastructure services.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          AWOERP HOST PROCESS                            │
│                                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────────┐  │
│  │  Fiber HTTP  │  │  Temporal    │  │  PostgreSQL  │  │   Redis    │  │
│  │  Router      │  │  Worker      │  │  (via PGX)   │  │  Broker    │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └─────┬──────┘  │
│         │                 │                  │                │         │
│  ┌──────▼─────────────────▼──────────────────▼────────────────▼──────┐  │
│  │                    PLUGIN RUNTIME (Core Layer)                     │  │
│  │                                                                    │  │
│  │   Plugin Registry │ Lifecycle Manager │ SDK Bridge │ Event Bus    │  │
│  └───────────────────────────────┬────────────────────────────────────┘  │
│                                  │                                      │
│              ┌───────────────────┼───────────────────┐                  │
│              │                   │                   │                  │
│   ┌──────────▼──────┐ ┌──────────▼──────┐ ┌─────────▼───────┐         │
│   │   Plugin A      │ │   Plugin B      │ │   Plugin C      │         │
│   │  (Logistics)    │ │  (Payroll Ext)  │ │  (Custom UI)    │         │
│   │                 │ │                 │ │                 │         │
│   │ Routes | Schema │ │ Routes | Schema │ │  AMIS Schemas   │         │
│   │ Workflows | UI  │ │ Workflows | UI  │ │  Event Hooks    │         │
│   └─────────────────┘ └─────────────────┘ └─────────────────┘         │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
         │                         │                        │
         ▼                         ▼                        ▼
  ┌─────────────┐          ┌──────────────┐        ┌──────────────┐
  │  Plugin     │          │  Plugin      │        │  AMIS-UI     │
  │  Registry   │          │  Marketplace │        │  Frontend    │
  │  (Registry  │          │  (Optional   │        │  Shell       │
  │   Service)  │          │   Commerce)  │        │              │
  └─────────────┘          └──────────────┘        └──────────────┘
```

---

## 3.2 Core Concepts & Mental Model

Before diving into implementation details, it is important to establish a shared vocabulary. Different systems use "plugin," "module," "extension," and "add-on" interchangeably, but in the Awoerp context these terms have precise meanings.

### 3.2.1 Plugin vs. Module vs. Extension — Clarifying Terminology

| Term | Awoerp Definition |
|---|---|
| **Core Module** | A first-party, built-in functional unit of Awoerp (e.g., the Finance module, the HR module). Core modules are part of the Awoerp codebase and are maintained by the Awoerp team. They are not plugins. |
| **Plugin** | A third-party or first-party extension unit that is distributed separately from the Awoerp core, installed via the plugin registry, and interacts with the core exclusively through the defined plugin interface. Plugins are the primary subject of this document. |
| **Extension Point** | A specific, named location in the core system where a plugin is permitted to inject behavior. Examples: "before invoice is finalized," "new tab in the customer detail screen," "new API route under /api/v1/plugins/." |
| **Hook** | A specific extension point that is event-driven: the plugin registers a callback and the core calls it when a specific event occurs. |
| **SDK** | The Go package (and toolchain) that Awoerp provides to plugin authors, abstracting away the raw plugin interface into a more ergonomic development experience. |

### 3.2.2 The Host-Guest Contract

The plugin system operates on a **host-guest model**. Awoerp is the **host** — it provides the runtime environment, the infrastructure services, and the execution context. Each plugin is a **guest** — it operates within the host's environment, subject to the host's rules, and communicates with the host only through the defined interface.

This is a deliberately asymmetric relationship:

- The **host** can terminate a guest at any time (health check failure, policy violation, admin action).
- The **guest** cannot access any host resource not explicitly granted via the permission declaration.
- The **host** guarantees the stability of its public interfaces (the plugin contract) across minor versions.
- The **guest** is responsible for declaring all its dependencies, side effects, and required permissions in its manifest.

This contract is not merely a software pattern — it is also a legal and commercial boundary. Plugin authors are responsible for their plugins' behavior. Awoerp's security model enforces that guests cannot exceed their declared permissions, but the host is not liable for what a guest does within its granted permissions.

---

## 3.3 Plugin Lifecycle

Every plugin goes through a defined lifecycle, managed by the Plugin Runtime. Understanding this lifecycle is essential for both plugin authors and core maintainers.

### 3.3.1 Discovery

Discovery is the process by which the Plugin Runtime becomes aware that a plugin exists. There are three discovery mechanisms:

**Static Discovery:** Plugins listed in the `awoerp.plugins.yaml` configuration file are loaded at application startup. This is the recommended mechanism for production deployments where the plugin set is known and controlled.

**Registry Pull:** The Plugin Runtime queries the Plugin Registry for plugins that should be installed based on the current environment's plugin subscription list. This enables centrally managed plugin deployments in multi-tenant SaaS scenarios.

**Hot Discovery (v1.1+):** A filesystem or registry watcher detects new plugin artifacts and triggers loading without a restart. This is deferred from v1.0 due to complexity of ensuring safe hot-loading in a Go process.

### 3.3.2 Registration

Once a plugin artifact is discovered, the Plugin Runtime reads and validates its **manifest** — a YAML file embedded in the plugin artifact that declares the plugin's identity, version, required permissions, extension points, and dependencies.

```yaml
# Example plugin manifest
apiVersion: awoerp.io/v1
kind: Plugin
metadata:
  id: com.example.logistics-bridge
  name: Logistics Bridge
  version: 1.2.0
  author: Example Logistics Ltd
  license: MIT
  minCoreVersion: "1.0.0"
  maxCoreVersion: "2.x"

permissions:
  database:
    - namespace: logistics_bridge    # own namespace only
    - read: [invoices, shipments]    # read access to core tables
  api:
    - mount: /api/v1/plugins/logistics
  workflows:
    - register: true
  events:
    - subscribe: [invoice.finalized, shipment.created]
    - publish: [logistics.tracking.updated]
  ui:
    - extend: customer_detail_tabs
    - extend: invoice_actions_menu

dependencies:
  - id: com.awoerp.core-finance
    version: ">=1.0.0"
```

If manifest validation passes, the plugin is registered in the runtime's internal registry. If it fails (missing required fields, invalid permissions, incompatible version), the plugin is rejected and an error is logged — but the host process continues normally.

### 3.3.3 Initialization

After registration, the Plugin Runtime calls the plugin's `Init` method, passing it a `PluginContext` struct that provides access to all the core services the plugin declared it needs. This is the plugin's opportunity to:

- Run database migrations for its own namespace.
- Register its HTTP routes with the Fiber router.
- Register its Temporal workflows and activities.
- Subscribe to Redis event channels.
- Register its AMIS-UI schema extensions.

If `Init` returns an error, the plugin is marked as failed and its extension points are not activated. The system logs the failure and continues operating; other plugins and core functionality are unaffected.

### 3.3.4 Execution

Once initialized, the plugin's extension points are live. HTTP requests to the plugin's mounted routes are handled by the plugin's route handlers. Events published on subscribed channels invoke the plugin's event handlers. Temporal workflow triggers call the plugin's workflow definitions.

During execution, all plugin activity is instrumented by the core runtime: request latency, error rates, database query counts, and event processing times are all tracked per plugin and exposed through Awoerp's observability stack.

### 3.3.5 Teardown / Unloading

When a plugin is disabled, updated, or the host process is shutting down, the Plugin Runtime calls the plugin's `Shutdown` method with a context that carries a deadline. The plugin must:

- Complete or cancel in-flight requests within the deadline.
- Flush any buffered writes.
- Deregister its routes, workflows, and event subscriptions.
- Release any resources it holds.

If the plugin does not return from `Shutdown` within the deadline, the runtime forcibly terminates its goroutines and logs a warning. This is a non-graceful shutdown and may result in partial writes or stuck workflows — both of which are the plugin author's responsibility to handle via idempotency and compensating transactions.

---

## 3.4 Plugin Isolation Boundaries

> **⚙️ Technical Note**

A central challenge of any plugin system is determining how isolated plugins are from each other and from the host. Higher isolation means better safety but higher overhead and more friction for plugin authors. The Awoerp v1.0 plugin system makes a deliberate choice: **process-level isolation via gRPC subprocess**, with in-process execution available as an opt-in for trusted (first-party) plugins.

The two supported isolation models are:

**In-Process Plugins (Trusted/First-Party)**
The plugin runs as a Go package within the same OS process as the Awoerp host. This is the lowest-overhead model and is appropriate for:
- Plugins developed and audited by the Awoerp team.
- Internal enterprise plugins deployed by a single organization that trusts its own code.

In-process plugins interact with the host through the same interface as subprocess plugins, but function calls are direct Go method invocations. The constraint is that a bug in an in-process plugin can theoretically affect the host process (panic, memory corruption). Careful runtime recovery (via `recover()`) mitigates panics, but this model is not appropriate for untrusted third-party code.

**Subprocess Plugins (Third-Party/Public)**
The plugin runs as a separate OS process, communicating with the host via a local Unix domain socket using gRPC. The host and plugin each hold one end of the gRPC connection. All calls across the boundary are serialized as Protocol Buffer messages.

This provides true process isolation: a crashing plugin process does not affect the host. The Plugin Runtime monitors subprocess health and can restart a crashed plugin automatically, subject to a backoff and circuit-breaker policy.

The overhead of gRPC serialization on a Unix socket is approximately 0.1–0.5ms per call on modern hardware — acceptable for plugin interactions, which should never be in the hot path of latency-sensitive operations.

> **⚠️ Warning:** WebAssembly (WASM) sandboxing was evaluated as a third isolation model and is documented in Appendix C (ADR-004). It was deferred from v1.0 due to the immature state of Go-to-WASM compilation for server-side use and significant performance overhead for database-access patterns. It remains a candidate for v2.0.

---

## 3.5 Versioning & Compatibility Guarantees

The plugin system's value is directly tied to the stability of its interfaces. Plugin authors need confidence that a plugin built against Awoerp 1.0 will continue to work against Awoerp 1.3 without modification.

Awoerp adopts **Semantic Versioning (SemVer)** for both the core system and the plugin SDK:

- **Major version bumps** (1.x → 2.0) may include breaking changes to the plugin interface. Plugins must declare a `maxCoreVersion` in their manifest. When a major version bump occurs, old plugins continue to work until their `maxCoreVersion` is exceeded, at which point they are disabled with a clear error message.
- **Minor version bumps** (1.0 → 1.1) may add new extension points or new fields to existing interfaces, but may not remove or modify existing ones. All additions are backward-compatible.
- **Patch version bumps** (1.0.0 → 1.0.1) contain no interface changes whatsoever.

The plugin SDK is versioned independently of the core but follows the same compatibility policy. The SDK's Go module path includes a major version suffix (`awoerp.io/sdk/v1`, `awoerp.io/sdk/v2`) to allow multiple major versions to coexist in the Go module graph if needed.

---

---

# 4. Plugin Types & Extension Points

Awoerp's plugin system defines **seven categories of extension points**, each corresponding to a different layer of the application stack. A single plugin can use any combination of these extension types.

## 4.1 UI Plugins (Frontend Extensions via AMIS-UI)

**What it is:** UI plugins extend the Awoerp frontend by injecting new screens, form fields, tabs, menu items, dashboard widgets, or entirely new pages — all without modifying the frontend codebase.

**Why it's needed:** The ERP frontend is where users spend the majority of their time. If plugins can only add backend behavior but not surface it in the UI, they are severely limited in usefulness. An inventory management plugin that adds a new "Warehouse Zones" concept needs to expose that concept in the UI — create/edit/list screens, navigation links, and inline widgets in existing screens (like showing zone information on a product detail page).

**How it works:** Awoerp's frontend is built on **AMIS-UI**, a JSON/schema-driven frontend framework (documented in detail in Section 6.7). UI plugins register **AMIS schema fragments** with the Plugin Runtime during initialization. The runtime merges these fragments into the frontend shell's schema at runtime. The frontend shell requests its full schema from the backend on load, which includes all enabled plugins' schema contributions.

This means UI extensions require **zero JavaScript or React knowledge** from plugin authors. A plugin author writes Go code that returns a structured AMIS schema — a JSON object describing a form, a table, a button, or an entire page — and Awoerp's frontend automatically renders it.

**Extension sub-types:**
- New top-level navigation pages
- New tabs in existing detail views (customer, invoice, product, etc.)
- New columns in existing list views
- New actions in contextual action menus
- New dashboard widgets
- New form fields in existing create/edit forms

**What is lost without this in v1.0:** Plugins become "headless" — they can run backend logic but cannot expose any UI. This forces plugin authors to build and host separate frontend applications and link to them from Awoerp, creating a fragmented user experience. Users must navigate between multiple applications rather than having a unified interface. This single limitation would disqualify most business workflow plugins from being viable.

---

## 4.2 Business Logic Plugins

**What it is:** Business logic plugins implement custom processing rules, calculations, validations, and transformations that run as part of the core system's execution flow — either synchronously (in-line with a request) or asynchronously (as a background process).

**Why it's needed:** Business rules are the most diverse and rapidly-changing part of any ERP system. Tax calculation rules differ by country and change with legislation. Approval workflows differ by industry and company size. Discount and pricing rules differ by customer segment. The core system cannot and should not try to enumerate all of these variations. Business logic plugins allow each of these concerns to be encapsulated and independently maintained.

**How it works:** Business logic plugins register as handlers for specific **core lifecycle hooks**. These hooks are called at defined points in the core system's processing pipeline:

```go
// Example: A plugin that adds custom validation before an invoice is finalized
type InvoicePreFinalizeHook struct{}

func (h *InvoicePreFinalizeHook) Handle(ctx context.Context, event *sdk.InvoiceEvent) error {
    invoice := event.Invoice
    // Custom validation: require a purchase order reference for invoices > $10,000
    if invoice.TotalAmount > 10000 && invoice.POReference == "" {
        return sdk.NewValidationError(
            "PO_REQUIRED",
            "Invoices over $10,000 require a purchase order reference",
        )
    }
    return nil
}
```

The core system calls all registered hooks for a given event in priority order. If any hook returns an error, the pipeline is halted and the error is returned to the caller. Hooks can also modify event data (enrichment hooks) without halting the pipeline.

**What is lost without this in v1.0:** Without business logic hooks, every custom rule must either be built into the core (adding permanent complexity) or implemented as a workaround outside the system (webhook receiver, cron job, manual process). The former degrades the core; the latter creates unreliable, hard-to-maintain shadow processes.

---

## 4.3 Workflow Plugins (Temporal-Backed)

**What it is:** Workflow plugins define long-running, multi-step, durable business processes that may span hours, days, or weeks and involve human approval steps, external system calls, and conditional branching.

**Why it's needed:** Modern ERP processes are rarely simple request/response operations. An employee onboarding process might involve: creating an HR record, triggering IT to provision accounts, sending welcome emails, scheduling orientation, waiting for signed documents, then activating payroll. This sequence can take days and must survive server restarts, network failures, and human delays. Standard HTTP request handlers cannot model this — they are stateless and time-bounded.

**How it works:** Temporal (documented in Section 6.3) is Awoerp's workflow orchestration engine. Plugins can define **Temporal Workflow and Activity functions** in their own Go code. The Plugin Runtime registers these workflow and activity definitions with the Temporal worker at plugin initialization time.

```go
// A plugin-defined Temporal workflow
func (w *OnboardingPlugin) EmployeeOnboardingWorkflow(
    ctx workflow.Context,
    input *OnboardingInput,
) error {
    // Step 1: Create HR record (activity)
    var hrRecord HRRecord
    if err := workflow.ExecuteActivity(ctx, w.CreateHRRecord, input).Get(ctx, &hrRecord); err != nil {
        return err
    }

    // Step 2: Wait for IT provisioning signal (could take hours/days)
    var itConfirmation ITConfirmation
    workflow.GetSignalChannel(ctx, "it-provisioning-complete").Receive(ctx, &itConfirmation)

    // Step 3: Activate payroll
    return workflow.ExecuteActivity(ctx, w.ActivatePayroll, hrRecord.ID).Get(ctx, nil)
}
```

Because Temporal persists workflow state to its own database, a workflow can survive any number of server restarts and continues exactly where it left off. This durability is provided by Temporal itself — the plugin author writes normal Go code and the infrastructure handles fault tolerance.

**What is lost without this in v1.0:** Long-running process automation must be implemented as fragile chains of webhooks and cron jobs. Without Temporal's durability guarantees, failures at any step leave processes in undefined states with no reliable compensation mechanism. This is one of the most common sources of data corruption in ERP systems built without proper workflow infrastructure.

---

## 4.4 Data Layer Plugins (Schema & Query Extensions via SQLC + PGX)

**What it is:** Data layer plugins extend the Awoerp database schema by adding tables, indexes, views, and stored procedures within a plugin-specific namespace. They interact with the database using type-safe Go code generated from SQL queries via SQLC.

**Why it's needed:** Most meaningful plugins need to persist data. A logistics plugin needs to store shipment records. A custom reporting plugin needs to store report definitions. A third-party CRM integration needs to store sync state. Without the ability to extend the database schema, plugins are forced to encode their data as unstructured blobs in generic key-value tables — an approach that is unqueryable, unmaintainable, and unsafe.

**How it works:** Each plugin is allocated a **PostgreSQL schema namespace** matching its plugin ID (e.g., `com_example_logistics_bridge`). All plugin tables are created within this namespace. The core system's tables live in the `public` schema. Plugins can read from the `public` schema (subject to their declared permissions) but can only write to their own namespace.

Plugin migrations are managed by a migration runner embedded in the Plugin SDK. On `Init`, the plugin calls `sdk.RunMigrations()` which applies any pending migrations from the plugin's embedded `migrations/` directory:

```go
func (p *LogisticsPlugin) Init(ctx context.Context, pc *sdk.PluginContext) error {
    // Run plugin-specific database migrations
    if err := pc.DB.RunMigrations(ctx, p.migrations); err != nil {
        return fmt.Errorf("migration failed: %w", err)
    }
    // ...
}
```

The plugin's query layer is generated by SQLC from `.sql` files in the plugin's source tree. This provides compile-time type safety for all database interactions:

```sql
-- name: GetShipmentsByInvoice :many
SELECT s.* FROM com_example_logistics_bridge.shipments s
WHERE s.invoice_id = $1
ORDER BY s.created_at DESC;
```

SQLC generates a corresponding Go function `GetShipmentsByInvoice(ctx, invoiceID)` that returns a typed `[]Shipment` slice. No runtime reflection, no stringly-typed query building.

**What is lost without this in v1.0:** Without a principled data extension model, plugins either cannot persist data at all (severely limiting their utility) or resort to storing data outside the Awoerp database entirely (creating data silos, synchronization problems, and backup gaps). The schema namespace model is also the primary mechanism for tenant isolation — without it, multi-tenant safety cannot be guaranteed.

---

## 4.5 API Plugins (Route & Middleware Extensions via Fiber)

**What it is:** API plugins add new HTTP endpoints to the Awoerp API server, allowing plugins to expose their own REST (or other) APIs that client applications, external systems, and webhooks can call.

**Why it's needed:** Plugins need to receive external input, not just react to internal events. A payment gateway integration plugin needs to receive webhook callbacks from the payment provider. A mobile app extension needs new API endpoints specific to its data model. An external reporting tool needs to query plugin-specific data. Without the ability to add routes, these interactions require building and hosting a separate API server — defeating the purpose of having a plugin system.

**How it works:** The Plugin SDK provides a route registration mechanism that mounts plugin routes under the plugin's declared API prefix:

```go
func (p *PaymentPlugin) RegisterRoutes(router *sdk.Router) {
    // All routes are mounted under /api/v1/plugins/com.example.payments/
    router.POST("/webhook/stripe", p.handleStripeWebhook)
    router.GET("/transactions", p.listTransactions)
    router.GET("/transactions/:id", p.getTransaction)
    router.POST("/refunds", p.initiateRefund)
}
```

The SDK router automatically applies authentication middleware, request logging, and rate limiting to all plugin routes. Plugins can additionally declare custom middleware:

```go
router.Use(p.requireAPIKey) // Plugin-specific auth for external webhook receivers
```

Plugin routes are versioned alongside the plugin itself. When a plugin is updated with a new API version, both old and new route versions can be active simultaneously during a transition period.

**What is lost without this in v1.0:** Plugins are limited to receiving input only through the event system (reactive) and cannot expose new APIs (proactive). This makes it impossible to build integrations that require callback URLs, impossible to build mobile-app-facing plugins with custom data shapes, and impossible to build plugins that other systems query directly.

---

## 4.6 Event/Hook Plugins (Pub-Sub with Redis)

**What it is:** Event plugins subscribe to named events published by the core system and by other plugins, reacting to them asynchronously. They can also publish new events for other plugins or external systems to consume.

**Why it's needed:** Not all plugin reactions need to be synchronous or in-band with the triggering request. Sending a notification when a new customer is created, syncing data to an external CRM when a contact is updated, updating a reporting aggregate when an invoice is closed — all of these are better modeled as asynchronous reactions to events. Forcing them to be synchronous hooks would add latency to every core operation; the event bus decouples them.

**How it works:** The Plugin Runtime wraps Redis Pub/Sub (and Redis Streams for durable delivery) into a typed event bus abstraction:

```go
func (p *NotificationPlugin) Init(ctx context.Context, pc *sdk.PluginContext) error {
    // Subscribe to core events
    pc.Events.Subscribe("invoice.finalized", p.onInvoiceFinalized)
    pc.Events.Subscribe("customer.created", p.onCustomerCreated)

    // Subscribe to another plugin's events
    pc.Events.Subscribe("com.example.logistics.shipment.dispatched", p.onShipmentDispatched)
    return nil
}

func (p *NotificationPlugin) onInvoiceFinalized(ctx context.Context, e *sdk.Event) error {
    invoice := &Invoice{}
    if err := e.UnmarshalPayload(invoice); err != nil {
        return err
    }
    // Send email notification, push notification, etc.
    return p.notifier.SendInvoiceNotification(ctx, invoice)
}
```

Events are delivered with **at-least-once semantics** via Redis Streams. Each plugin's event handler group has a consumer group in Redis, ensuring that each event is processed by each subscribing plugin exactly once (with retry on failure up to a configurable limit).

**What is lost without this in v1.0:** Without an event bus, the only way for plugins to react to core system events is through synchronous hooks (adding latency to every operation) or polling (inefficient and with poor real-time behavior). Asynchronous decoupled reactions are one of the most common patterns in ERP integration, and their absence would force integrations to be brittle polling loops or latency-adding synchronous hooks.

---

## 4.7 Cross-cutting: Auth, Audit & Tenant-Aware Plugins

**What it is:** These are not a separate extension type but rather a set of capabilities that all plugin types share: awareness of the authenticated user's identity and tenant context, and automatic audit logging of all plugin actions.

**Why it's needed:** Awoerp is designed for multi-tenant deployment (multiple independent businesses running on the same infrastructure). Every plugin action must be scoped to the correct tenant. A logistics plugin for Tenant A must never be able to read or modify data belonging to Tenant B. This is not something plugin authors should have to implement themselves — it must be enforced by the infrastructure.

**How it works:** The Plugin Runtime injects a **TenantContext** into every plugin call. Database queries executed through the SDK's DB handle automatically prepend a `SET app.current_tenant_id = $tenantID` command using PostgreSQL's application-level parameters, and Row-Level Security (RLS) policies on plugin schema tables enforce that all queries are scoped to the current tenant.

Audit logging is similarly automatic: every plugin SDK call that results in a data modification is recorded in the `audit_log` table with the plugin ID, tenant ID, user ID, timestamp, operation type, and a diff of the changed data.

**What is lost without this in v1.0:** Multi-tenant safety cannot be guaranteed. A plugin bug or malicious plugin could access cross-tenant data. Audit log gaps make compliance reporting (SOC 2, ISO 27001, GDPR) impossible for actions taken by plugins. Both of these are dealbreakers for enterprise sales.

---

---

# 5. The Plugin Contract & SDK

## 5.1 The Plugin Interface (Go Interface Definition)

The plugin interface is the most critical piece of the entire plugin system. It is the formal contract between the Awoerp host and every plugin. Its design must balance expressiveness (giving plugins enough hooks to be useful) with stability (changing the interface is a breaking change for all existing plugins).

The core interface is intentionally minimal — additional capabilities are accessed through the `PluginContext` rather than being methods on the interface itself. This means new capabilities can be added without changing the interface:

```go
// Package sdk — awoerp.io/sdk/v1
package sdk

import (
    "context"
)

// Plugin is the interface every Awoerp plugin must implement.
// All methods must be safe to call concurrently.
type Plugin interface {
    // Metadata returns the plugin's static identity information.
    // Must not block and must return consistent values across calls.
    Metadata() PluginMetadata

    // Init is called once when the plugin is loaded. The plugin should
    // use the provided PluginContext to register routes, subscribe to events,
    // run migrations, and acquire any resources it needs.
    // If Init returns an error, the plugin will not be activated.
    Init(ctx context.Context, pc *PluginContext) error

    // HealthCheck is called periodically by the runtime to assess plugin health.
    // Return nil for healthy, a descriptive error for degraded/unhealthy.
    HealthCheck(ctx context.Context) error

    // Shutdown is called when the plugin is being unloaded or the host is
    // shutting down. The plugin must release all resources before ctx is
    // cancelled. Failure to return before ctx cancellation results in
    // forced termination.
    Shutdown(ctx context.Context) error
}

// PluginMetadata contains static information about the plugin.
type PluginMetadata struct {
    ID          string // Reverse-domain identifier: com.example.plugin-name
    Name        string // Human-readable name
    Version     string // SemVer string: "1.2.3"
    Description string
    Author      string
    License     string
    Homepage    string
}

// PluginContext is provided to a plugin during Init and provides
// access to all core services the plugin declared it needs.
// The available services are determined by the plugin's manifest;
// accessing an undeclared service panics with a clear error message
// to aid plugin authors during development.
type PluginContext struct {
    // TenantID is the ID of the tenant this plugin instance serves.
    // In a multi-tenant deployment, one plugin instance is created per tenant.
    TenantID string

    // DB provides access to the database with automatic tenant scoping
    // and migration management.
    DB *DBHandle

    // Router provides HTTP route registration under the plugin's
    // declared API prefix.
    Router *Router

    // Events provides event subscription and publishing.
    Events *EventBus

    // Workflows provides Temporal workflow and activity registration.
    Workflows *WorkflowRegistry

    // UI provides AMIS schema fragment registration.
    UI *UIExtensionRegistry

    // Cache provides access to the Redis cache with automatic
    // key namespacing for the plugin.
    Cache *CacheHandle

    // Logger is a structured logger pre-configured with the plugin's
    // identity and tenant context.
    Logger *Logger

    // Config provides access to the plugin's declared configuration
    // values, sourced from environment variables or the config service.
    Config *ConfigHandle
}
```

---

## 5.2 Plugin Manifest Schema (JSON/YAML)

The manifest is validated at registration time against a versioned JSON Schema. The full schema reference is in Appendix B. The following annotated example covers the most important fields:

```yaml
apiVersion: awoerp.io/v1          # Schema version for the manifest format itself
kind: Plugin                       # Always "Plugin" for plugin manifests

metadata:
  id: com.example.payroll-ext      # Globally unique reverse-domain ID
  name: Payroll Extensions         # Human-readable display name
  version: 2.1.0                   # Plugin's own SemVer version
  description: >
    Adds support for commission-based payroll structures,
    multi-currency payroll, and custom payroll approval workflows.
  author: Example Corp <plugins@example.com>
  license: Apache-2.0
  homepage: https://example.com/awoerp-plugins/payroll-ext
  repository: https://github.com/example/awoerp-payroll-ext
  tags: [payroll, hr, finance, commissions]
  icon: https://example.com/icons/payroll-ext.png

compatibility:
  minCoreVersion: "1.0.0"          # Minimum Awoerp core version required
  maxCoreVersion: "1.x"            # Will refuse to load on 2.0+
  sdkVersion: "^1.0.0"             # Minimum SDK version

permissions:
  # Database permissions
  database:
    ownNamespace: true              # Always granted; plugin's own schema
    readTables:                     # Core tables the plugin may read
      - public.employees
      - public.payroll_runs
      - public.salary_structures
    # Writing to core tables is never permitted; plugins use events/hooks

  # API permissions
  api:
    mount: /api/v1/plugins/payroll-ext   # The plugin's route prefix

  # Workflow permissions
  workflows:
    register: true                  # Plugin may register workflow definitions

  # Event permissions
  events:
    subscribe:
      - payroll.run.started
      - payroll.run.completed
      - hr.employee.updated
    publish:
      - payroll-ext.commission.calculated
      - payroll-ext.approval.required

  # UI permissions
  ui:
    extend:
      - employee_detail_tabs        # Add a tab to the employee detail view
      - payroll_run_actions         # Add actions to the payroll run action menu
      - hr_dashboard_widgets        # Add dashboard widgets

dependencies:
  plugins:
    - id: com.awoerp.core-hr
      version: ">=1.0.0"
    - id: com.awoerp.core-finance
      version: ">=1.0.0"

config:
  # Declared configuration keys — values sourced from environment/config service
  - key: commission_calculation_mode
    description: How commissions are calculated (tiered, flat, custom)
    type: string
    default: tiered
    required: false
  - key: multi_currency_enabled
    type: boolean
    default: false
    required: false
  - key: approval_webhook_url
    type: string
    required: false
    sensitive: false
```

---

## 5.3 Metadata, Permissions & Capability Declaration

**Why explicit capability declaration matters:**

The manifest's permissions section is not advisory — it is enforced. When the Plugin Runtime initializes a `PluginContext`, it checks the manifest's permissions and only populates the fields that were declared. Attempting to access `pc.DB.ReadTable("public.invoices")` from a plugin that did not declare `readTables: [public.invoices]` in its manifest results in an immediate, clearly described panic during development and a logged security violation in production.

This explicit-declaration model serves three purposes:

1. **Security:** No plugin can silently acquire access to resources it didn't declare. Security audits of a plugin are bounded by its manifest.
2. **Transparency:** Users and administrators installing a plugin can review exactly what it will access before installing it, analogous to mobile app permission prompts.
3. **Dependency management:** The Plugin Runtime uses capability declarations to construct the correct `PluginContext` for each plugin, avoiding unnecessary initialization overhead for unused services.

---

## 5.4 Plugin Configuration & Environment Injection

Plugins often need configuration values that vary by deployment: API keys for external services, feature flags, thresholds. The Plugin SDK provides a typed configuration system:

```go
// In plugin Init:
commissionMode := pc.Config.GetString("commission_calculation_mode") // "tiered"
multiCurrency := pc.Config.GetBool("multi_currency_enabled")         // false

// Config values are sourced in priority order:
// 1. Environment variables: PLUGIN_COM_EXAMPLE_PAYROLL_EXT_COMMISSION_CALCULATION_MODE
// 2. Awoerp configuration service (for multi-tenant deployments)
// 3. Default values from the manifest
```

For sensitive values (API keys, secrets), the SDK integrates with Awoerp's secrets management layer. Sensitive config values are never logged, never exposed in the admin UI in plaintext, and are stored encrypted at rest.

---

## 5.5 Error Handling Conventions

The SDK defines a typed error hierarchy that allows the Plugin Runtime and core system to handle plugin errors appropriately:

```go
// Validation errors are returned to the end user with the provided message
sdk.NewValidationError(code, message string) *PluginError

// Permission errors indicate the plugin attempted to access something
// it wasn't granted — these are logged as security events
sdk.NewPermissionError(resource string) *PluginError

// Retryable errors indicate a transient failure — the runtime may retry
sdk.NewRetryableError(cause error) *PluginError

// Fatal errors indicate the plugin cannot continue operating and should
// be marked as failed
sdk.NewFatalError(cause error) *PluginError
```

Plugin authors should use these typed errors rather than raw `errors.New()` or `fmt.Errorf()`. The runtime inspects error types to determine the appropriate response: retrying transient failures, logging security violations, returning validation messages to users.

---

## 5.6 SDK Packaging & Distribution

The Awoerp Plugin SDK is distributed as a standard Go module:

```bash
go get awoerp.io/sdk/v1
```

The SDK module includes:
- All plugin interface types and the `PluginContext` definition.
- Helper packages for common patterns: pagination, cursor-based queries, event payload marshaling, AMIS schema builders.
- The `awoerp-plugin` CLI tool for scaffolding, building, signing, and publishing plugins.
- A local development server that emulates the Plugin Runtime environment for testing without a full Awoerp installation.

A minimal plugin skeleton is scaffolded with:

```bash
awoerp-plugin init com.example.my-plugin \
    --name "My Plugin" \
    --author "Your Name <you@example.com>"
```

This generates a `go.mod`, a `main.go` with a skeleton implementation of the `Plugin` interface, a `manifest.yaml`, a `migrations/` directory, and a `plugin_test.go` with examples of how to test against the emulated runtime.

---

---

# 6. Technology Stack Deep Dive & Decision Rationale

## 6.1 Overview — Why These Six Components?

The Awoerp plugin system is not built in a vacuum — it is built on top of Awoerp's existing Go-native technology stack. Each component in this stack was chosen for specific properties that matter for a plugin system: performance, type safety, operational simplicity, and Go-nativeness. This section examines each component in depth, compares it to alternatives, and explains how it specifically serves the plugin system's needs.

The six components are:

| Component | Role in Plugin System |
|---|---|
| **Fiber** | HTTP server and plugin route mounting |
| **Temporal** | Durable workflow orchestration for plugin-defined processes |
| **PostgreSQL + PGX** | Primary data store with plugin schema namespacing |
| **SQLC** | Type-safe query generation for plugin data access code |
| **Redis** | Asynchronous event bus and distributed plugin coordination |
| **AMIS-UI** | Schema-driven frontend extension by plugins |

---

## 6.2 HTTP Framework — Fiber

### 6.2.1 What It Does in the Plugin System

Fiber is the HTTP framework powering Awoerp's API layer. Within the plugin system, Fiber provides the mechanism by which plugins mount their own HTTP route handlers alongside the core API, apply custom middleware, and handle incoming webhook callbacks or API requests.

The plugin system uses Fiber's **sub-application mounting** feature: each plugin gets its own `fiber.App` instance (or `fiber.Router` group), which is mounted under the plugin's declared prefix at initialization time and unmounted cleanly at shutdown.

### 6.2.2 Comparison: Fiber vs. Echo vs. Chi vs. Gin vs. net/http

| Framework | Performance (req/sec) | Plugin Isolation | Middleware Model | Maturity | Go Idioms |
|---|---|---|---|---|---|
| **Fiber** | ~500k (fasthttp) | Sub-app mounting | Chain-based | High | Moderate (fasthttp diverges from std) |
| **Echo** | ~450k | Router group | Chain-based | Very High | High (std-compatible) |
| **Chi** | ~420k | Sub-router | Chain-based | High | Very High (net/http native) |
| **Gin** | ~430k | Engine group | Chain-based | Very High | High |
| **net/http** | ~380k | ServeMux | Handler wrap | stdlib | Perfect |

**Why Fiber for Awoerp:**

Fiber is built on **fasthttp** rather than the standard `net/http`. This delivers the highest raw throughput — important for an ERP system that may process hundreds of concurrent API requests. Fiber's sub-application mounting is a clean match for the plugin model: each plugin is its own sub-app with its own middleware stack, and the host app mounts it. This provides clean isolation without any shared state in the router layer.

The main tradeoff with Fiber is that fasthttp's `*fasthttp.RequestCtx` is not interchangeable with the standard `*http.Request`, meaning some standard library middleware cannot be used directly. The Plugin SDK wraps this complexity: plugin route handlers work with a standard-looking SDK request/response API, and the SDK internally handles the fasthttp layer.

**Considered alternative — Echo:** Echo was a close second. It has better standard library compatibility and a similarly clean sub-router model. If Awoerp's throughput requirements were less aggressive, Echo would be the recommendation. For the specific use case of a high-traffic multi-tenant ERP API, Fiber's performance edge is worth the fasthttp complexity at the framework boundary.

**Considered alternative — Chi:** Chi's pure `net/http` compatibility is appealing from a Go idiom standpoint. However, Chi lacks Fiber's performance characteristics and does not have a clean sub-application isolation concept — it uses sub-routers, which are mount points but not isolated middleware stacks. This makes clean plugin teardown harder.

### 6.2.3 Plugin Router Mounting & Middleware Chaining

```go
// Core system initializes the Fiber app
coreApp := fiber.New()

// Core API routes
coreApp.Get("/api/v1/invoices", invoiceHandler.List)
// ... other core routes ...

// Plugin Runtime mounts each plugin's router
pluginRouter := coreApp.Group("/api/v1/plugins/" + plugin.Metadata().ID)
pluginRouter.Use(
    middleware.Authenticate,       // Core auth applies to all plugin routes
    middleware.TenantScope,        // Tenant isolation applies to all plugin routes
    middleware.PluginRateLimit,    // Per-plugin rate limiting
    middleware.PluginAuditLog,     // Audit logging for all plugin API calls
)

// Plugin registers its own routes on the provided router group
plugin.RegisterRoutes(sdk.NewRouter(pluginRouter))
```

### 6.2.4 What's Lost Without This in v1.0

Without Fiber's plugin route mounting, there is no standard way for a plugin to add HTTP endpoints. Plugins cannot receive webhooks, cannot expose custom APIs, and cannot integrate with external services that require callback URLs. The entire category of "integration plugins" — payment gateways, logistics carriers, communication platforms, third-party SaaS tools — becomes impossible to build cleanly.

---

## 6.3 Workflow Orchestration — Temporal

### 6.3.1 What It Does in the Plugin System

Temporal provides the infrastructure for **durable, long-running workflow execution**. In the plugin system, Temporal enables plugin authors to define multi-step business processes — employee onboarding, multi-level approval chains, automated reconciliation sequences — that are guaranteed to complete or compensate even in the face of infrastructure failures.

Without Temporal, these processes would need to be implemented as fragile state machines stored in the database and driven by cron jobs — a pattern that every experienced engineer recognizes as a source of bugs, race conditions, and invisible failures.

### 6.3.2 Comparison: Temporal vs. Cadence vs. Inngest vs. Machinery vs. Asynq

| System | Durability | Go SDK | Multi-Tenant | Complexity | Self-Hosted | Cloud Managed |
|---|---|---|---|---|---|---|
| **Temporal** | Excellent (event sourced) | First-class | Via namespace | High | Yes | Temporal Cloud |
| **Cadence** | Excellent (Temporal predecessor) | First-class | Via domain | High | Yes | No |
| **Inngest** | Good (step functions) | Via HTTP | Limited | Low | Partial | Yes |
| **Machinery** | Limited (Redis/AMQP tasks) | Native | No | Low | Yes | No |
| **Asynq** | Good (Redis streams) | Native | No | Low | Yes | No |
| **River** | Good (PostgreSQL-backed) | Native | Limited | Moderate | Yes | No |

**Why Temporal for Awoerp:**

Temporal's core value proposition — **workflows as code with automatic durability** — is exactly what complex ERP processes need. The key technical differentiator is Temporal's event-sourced execution model: a workflow's entire execution history is persisted, and if a worker crashes mid-workflow, a new worker picks up the workflow from its last event and continues. This is not something that can be replicated with a simple job queue.

For the plugin system specifically, Temporal's **namespace** isolation model is important: different plugins can register workflow definitions in the same Temporal cluster without collision, because workflow type names are scoped by the Temporal task queue, which is plugin-specific.

**Considered alternative — Cadence:** Cadence is Temporal's open-source predecessor (created by Uber). Temporal was founded by Cadence's creators and is essentially Cadence v2 with a more active community and better multi-language support. For a new system, Temporal is the correct choice.

**Considered alternative — Asynq:** Asynq (PostgreSQL-backed job queue via River or Redis-backed via Asynq) covers simple delayed jobs and queue-based task processing very well and with far less operational complexity than Temporal. For simple background tasks (sending an email, syncing a record), Asynq or a Redis-streams queue is appropriate. The plugin system uses Redis Streams for these simple cases. Temporal is reserved for true long-running, stateful workflows. Having both is not redundant — they serve genuinely different patterns.

**Considered alternative — Inngest:** Inngest is an emerging player with a developer-friendly step-function model and a managed cloud offering. It lacks the operational maturity and the Go-native SDK quality of Temporal for a production ERP context. Worth re-evaluating at v2.0.

### 6.3.3 Plugin-Defined Workflows & Activities

```go
// Plugin registers its workflow and activity definitions during Init
func (p *ProcurementPlugin) Init(ctx context.Context, pc *sdk.PluginContext) error {
    // Register with Temporal worker — task queue is automatically scoped
    // to this plugin to avoid collisions with other plugins
    pc.Workflows.Register(p.PurchaseOrderApprovalWorkflow)
    pc.Workflows.RegisterActivity(p.NotifyApprover)
    pc.Workflows.RegisterActivity(p.ProcessApproval)
    pc.Workflows.RegisterActivity(p.UpdatePOStatus)
    return nil
}

// Plugin can also trigger workflows from its HTTP handlers or event handlers
func (p *ProcurementPlugin) onPOSubmitted(ctx context.Context, e *sdk.Event) error {
    po := &PurchaseOrder{}
    e.UnmarshalPayload(po)

    // Start a durable workflow — this survives any server restart
    _, err := pc.Workflows.Start(ctx, "PurchaseOrderApprovalWorkflow", po)
    return err
}
```

### 6.3.4 What's Lost Without This in v1.0

Without Temporal, long-running process plugins cannot be built reliably. Plugins that need multi-step approval chains, scheduled follow-ups, or human-in-the-loop processes must either implement their own durable state machines (complex, error-prone) or use unreliable cron-job patterns (brittle, invisible failures). The most valuable ERP automation use cases — procurement approval, employee onboarding, period-end close sequences — all require this capability.

---

## 6.4 Relational Database — PostgreSQL + PGX

### 6.4.1 What It Does in the Plugin System

PostgreSQL is Awoerp's primary data store. PGX is the Go driver used to interact with it. In the plugin system context, PostgreSQL provides the plugin schema namespacing mechanism, Row-Level Security enforcement, and the transactional guarantees that make plugin data access safe in a multi-tenant environment.

### 6.4.2 Comparison: PGX vs. database/sql vs. GORM vs. XORM vs. Bun

| Driver/ORM | Type Safety | Performance | PostgreSQL Features | Code Generation | Multi-Tenant Support |
|---|---|---|---|---|---|
| **PGX v5** | High (named types) | Excellent | Full (LISTEN/NOTIFY, pgvector, etc.) | No (use with SQLC) | Via RLS + app params |
| **database/sql** | Moderate | Good | Limited | No | Manual |
| **GORM** | Low (interface{}) | Moderate | Limited | Partial (gen) | Via scopes (fragile) |
| **XORM** | Moderate | Moderate | Moderate | Partial | Manual |
| **Bun** | High | Good | Good | Partial | Via scopes |
| **Ent** | Very High | Good | Good | Yes | Via hooks |

**Why PGX for Awoerp:**

PGX v5 is the most performant and feature-complete PostgreSQL driver for Go. Key advantages for the plugin system:

- **Native PostgreSQL protocol support:** PGX uses the binary wire protocol for data transfer, which is significantly faster than the text protocol used by `database/sql`-based drivers for large result sets.
- **Connection pooling (pgxpool):** PGX's connection pool handles multi-tenant concurrent plugin queries efficiently, with pool size configuration per plugin.
- **PostgreSQL-specific features:** The plugin system relies on several PostgreSQL features that are not available through generic drivers: `SET LOCAL app.tenant_id` for RLS, `LISTEN/NOTIFY` for change events, named prepared statements, and `pg_notify` triggers.
- **Row-Level Security (RLS) integration:** PGX allows setting session-level and transaction-level application parameters, which is the mechanism by which tenant ID is communicated to PostgreSQL RLS policies transparently.

**Considered alternative — GORM:** GORM's Active Record-style ORM is appealing for rapid development but has significant problems for a plugin system. Its use of empty interface (`interface{}`) for query parameters bypasses Go's type system, making it easy to introduce runtime errors that SQLC would catch at compile time. GORM's "magic" (auto-migrations, hooks, callbacks) conflicts with the plugin system's need for explicit, auditable schema changes. For the core ERP schema where fine-grained control matters, GORM's convenience is outweighed by its opacity.

**Considered alternative — Ent:** Ent (from Meta/Facebook) is a compelling graph-based ORM with excellent code generation. It was evaluated seriously. The primary reason it was not chosen is that Ent's schema definition is Go code, which would require plugin authors to write Ent schema definitions rather than SQL — adding a new conceptual layer on top of what is fundamentally a SQL database. SQLC keeps plugin authors close to SQL, which is a universal skill for backend engineers.

### 6.4.3 Plugin Schema Namespacing & Migration Strategy

```sql
-- Core schema lives in the 'public' schema
-- Plugin schemas are named after the plugin ID (dots replaced with underscores)
CREATE SCHEMA IF NOT EXISTS com_example_logistics_bridge;

-- Row-Level Security on plugin tables
CREATE TABLE com_example_logistics_bridge.shipments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,   -- Every plugin table has a tenant_id column
    invoice_id  UUID NOT NULL REFERENCES public.invoices(id),
    carrier     TEXT NOT NULL,
    tracking_no TEXT,
    status      TEXT NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- RLS policy: all queries are automatically scoped to current tenant
ALTER TABLE com_example_logistics_bridge.shipments ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON com_example_logistics_bridge.shipments
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

### 6.4.4 What's Lost Without This in v1.0

Without PGX's PostgreSQL-specific features, multi-tenant security cannot be enforced at the database layer. Moving tenant isolation into application code is possible but creates a single critical failure path: one missed `WHERE tenant_id = ?` in any plugin query exposes cross-tenant data. RLS at the database layer provides defense-in-depth: even a buggy or malicious plugin cannot read another tenant's data because PostgreSQL enforces the policy for every query, regardless of what the application code does.

---

## 6.5 Type-Safe Query Generation — SQLC

### 6.5.1 What It Does in the Plugin System

SQLC is a Go code generator that takes `.sql` files (containing named SQL queries) and generates fully type-safe Go functions and structs for each query. In the plugin system, SQLC ensures that every database interaction a plugin makes is type-checked at compile time — not at runtime.

### 6.5.2 Comparison: SQLC vs. GORM vs. Ent vs. SQLBoiler vs. Raw SQL

| Approach | Type Safety | SQL Control | Code Gen | Learning Curve | Plugin Authoring DX |
|---|---|---|---|---|---|
| **SQLC** | Compile-time | Full (raw SQL) | Yes | Low | Excellent |
| **GORM** | Runtime only | Partial | Partial | Low | Good but risky |
| **Ent** | Compile-time | Partial (DSL) | Yes | High | Good |
| **SQLBoiler** | Compile-time | Full | Yes (from DB schema) | Moderate | Good |
| **Raw SQL + PGX** | Runtime only | Full | No | Low | Error-prone |
| **sqlx** | Runtime only | Full | No | Low | Moderate |

**Why SQLC for Awoerp:**

The plugin system specifically benefits from SQLC because:

1. **Plugin authors write SQL, not a Go ORM DSL.** SQL knowledge is nearly universal among backend engineers. Every developer who might author an Awoerp plugin can write SQL. Not every developer knows Ent's schema DSL or GORM's association magic.

2. **Compile-time type checking catches errors before deployment.** A plugin that queries `SELECT first_name FROM public.employees` but tries to assign the result to an `invoice.Amount` field will fail at compile time, not at runtime for a real user.

3. **The generated code is readable and auditable.** SQLC generates plain Go functions and structs. Security auditors reviewing a plugin's database access can read the generated code directly — it looks like hand-written code.

4. **SQLC validates SQL against the actual database schema at code generation time.** Referencing a column that doesn't exist, using the wrong type, or making a JOIN to a table the plugin doesn't have access to all produce errors during the `sqlc generate` step, before any code is committed.

```sql
-- queries/shipments.sql in a logistics plugin

-- name: GetShipmentsByInvoice :many
SELECT id, tenant_id, invoice_id, carrier, tracking_no, status, created_at
FROM com_example_logistics_bridge.shipments
WHERE invoice_id = @invoice_id
  AND tenant_id = current_setting('app.current_tenant_id')::uuid
ORDER BY created_at DESC;

-- name: CreateShipment :one
INSERT INTO com_example_logistics_bridge.shipments
    (tenant_id, invoice_id, carrier, tracking_no, status)
VALUES
    (current_setting('app.current_tenant_id')::uuid, @invoice_id, @carrier, @tracking_no, 'pending')
RETURNING *;
```

SQLC generates:

```go
// Auto-generated — do not edit
func (q *Queries) GetShipmentsByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]Shipment, error)
func (q *Queries) CreateShipment(ctx context.Context, arg CreateShipmentParams) (Shipment, error)
```

### 6.5.3 Plugin Query Contracts & Code Generation Pipeline

Each plugin runs `sqlc generate` as part of its build process. The SQLC configuration file references both the plugin's own schema files and (read-only) the core schema files for any core tables the plugin is permitted to query:

```yaml
# sqlc.yaml in a plugin
version: "2"
sql:
  - schema:
      - "../../core-schema/public/*.sql"  # Core schema (read reference only)
      - "migrations/*.sql"                # Plugin's own schema
    queries: "queries/*.sql"
    engine: "postgresql"
    gen:
      go:
        package: "db"
        out: "internal/db"
        emit_json_tags: true
        emit_interface: true
```

### 6.5.4 What's Lost Without This in v1.0

Without SQLC, plugin authors write raw SQL in string literals or use an ORM that sacrifices type safety. Both approaches increase the likelihood of runtime errors, SQL injection vulnerabilities (if string interpolation is used naively), and incorrect data access patterns. For a system where data integrity and security are paramount (ERP handles financial and HR data), compile-time query verification is not a luxury — it is a baseline safety requirement.

---

## 6.6 In-Memory Store & Message Broker — Redis

### 6.6.1 What It Does in the Plugin System

Redis serves three distinct roles in the Awoerp plugin system:

1. **Event Bus:** Redis Streams provide the asynchronous event pub/sub mechanism that allows plugins to react to core and other plugin events without blocking the main request path.
2. **Distributed Cache:** Plugins can cache frequently accessed, rarely-changing data (configuration, reference data, computed aggregates) in Redis with automatic key namespacing to prevent collisions between plugins.
3. **Distributed Locks and Coordination:** The Plugin Runtime uses Redis-based distributed locks (via Redlock algorithm) to ensure that plugin initialization, migration execution, and certain exclusive operations are not run concurrently across multiple Awoerp instances.

### 6.6.2 Comparison: Redis vs. NATS vs. RabbitMQ vs. Kafka vs. Valkey

| System | Throughput | Durability | Simplicity | Stream Support | Go SDK | Operational Cost |
|---|---|---|---|---|---|---|
| **Redis** | Very High | Good (AOF/RDB) | Very High | Yes (Streams) | Excellent | Low |
| **NATS** | Extremely High | Good (JetStream) | High | Yes (JetStream) | Excellent | Low |
| **RabbitMQ** | High | Excellent | Moderate | Partial | Good | Moderate |
| **Kafka** | Extremely High | Excellent | Low | Yes (core) | Good | High |
| **Valkey** | Very High | Good | Very High | Yes | Via Redis clients | Low |
| **PostgreSQL LISTEN/NOTIFY** | Moderate | Excellent | Very High | No | Via PGX | None (already there) |

**Why Redis for Awoerp:**

Redis was chosen for its combination of **operational simplicity and functional breadth**. A single Redis deployment handles the event bus, the cache layer, and distributed locking — three capabilities that would otherwise require three separate infrastructure components.

For the specific throughput requirements of an ERP event bus (not a high-frequency trading system — more like hundreds to low thousands of events per second), Redis Streams are more than sufficient. The streams model provides consumer groups with acknowledgment semantics, enabling reliable at-least-once delivery with automatic redelivery of unacknowledged messages.

**Considered alternative — NATS JetStream:** NATS is arguably the better purely technical choice for a messaging system: higher throughput, lower latency, and JetStream provides excellent stream durability. However, it is an additional infrastructure component that the Awoerp team would need to operate. Redis is already in the stack for caching; adding NATS adds operational overhead without proportionate benefit for ERP-scale event volumes. NATS is the recommended upgrade path if Awoerp's event throughput outgrows Redis Streams.

**Considered alternative — Kafka:** Kafka is the gold standard for high-throughput, durable event streaming but comes with significant operational complexity (ZooKeeper or KRaft, broker management, partition tuning). For ERP-scale workloads, Kafka is over-engineered. It is the right answer if Awoerp ever enters the "event sourcing as the system of record" architectural pattern, but that is a v3.0 conversation.

**Considered alternative — PostgreSQL LISTEN/NOTIFY:** PGX supports PostgreSQL's native LISTEN/NOTIFY mechanism, which would require no additional infrastructure. It was seriously considered for simplicity. The limitation is that NOTIFY payloads are limited to 8KB and LISTEN/NOTIFY has no persistence — if a subscriber is offline when a notification is sent, it is lost. For the plugin event bus where at-least-once delivery is required, this is a dealbreaker. LISTEN/NOTIFY is used internally for certain real-time core system events where loss is acceptable, but not for the plugin event bus.

**Note on Valkey:** Valkey is the open-source fork of Redis created after the Redis licensing change in 2024. It is API-compatible with Redis and all Redis client libraries work with Valkey. Awoerp officially supports both Redis and Valkey as the backend for the event bus and cache, configurable at deployment time.

### 6.6.3 Plugin Event Bus, Caching Layer & Distributed Locks

```go
// Event publishing from a plugin
func (p *FinancePlugin) onInvoiceFinalized(invoice *Invoice) error {
    return pc.Events.Publish(ctx, "invoice.finalized", &InvoiceFinalizedEvent{
        InvoiceID:  invoice.ID,
        TotalAmount: invoice.Total,
        Currency:   invoice.Currency,
        FinalizedAt: time.Now(),
    })
}

// Caching in a plugin — keys are automatically namespaced to prevent collisions
func (p *ReportingPlugin) getExchangeRates(ctx context.Context, currency string) (*ExchangeRates, error) {
    cacheKey := fmt.Sprintf("exchange_rates:%s", currency)

    // Try cache first
    var rates ExchangeRates
    if err := pc.Cache.Get(ctx, cacheKey, &rates); err == nil {
        return &rates, nil  // Cache hit
    }

    // Fetch from external source and cache for 1 hour
    rates, err = p.fetchFromExternalAPI(ctx, currency)
    if err != nil {
        return nil, err
    }
    pc.Cache.Set(ctx, cacheKey, rates, time.Hour)
    return &rates, nil
}

// Distributed lock during plugin migration (prevents concurrent runs in multi-instance deployment)
func (p *Plugin) Init(ctx context.Context, pc *sdk.PluginContext) error {
    lock, err := pc.Cache.AcquireLock(ctx, "migration_lock", 30*time.Second)
    if err != nil {
        return fmt.Errorf("could not acquire migration lock: %w", err)
    }
    defer lock.Release(ctx)

    return pc.DB.RunMigrations(ctx, p.migrations)
}
```

### 6.6.4 What's Lost Without This in v1.0

Without Redis, the plugin event bus cannot exist in its decoupled asynchronous form. Plugins lose the ability to cache data, meaning every plugin request that needs reference data (exchange rates, tax tables, configuration) must hit the database, adding latency and load. Without distributed locks, running multiple Awoerp instances creates race conditions in plugin initialization and migration execution. All three losses compound to make reliable multi-instance deployment of plugin-enabled Awoerp impossible.

---

## 6.7 Frontend Rendering Engine — AMIS-UI

### 6.7.1 What It Does in the Plugin System

AMIS-UI is a JSON schema-driven React frontend framework originally developed by Baidu and open-sourced. It allows entire UI screens — forms, tables, dashboards, wizards — to be defined as structured JSON/YAML objects rather than as React component trees. In the plugin system, AMIS enables plugin authors to extend the Awoerp frontend **without writing any JavaScript, React, or CSS**.

### 6.7.2 Comparison: AMIS vs. React (manual) vs. Retool vs. AppSmith vs. Refine

| Approach | No-Code UI Extension | Go-Friendly | Self-Hosted | Theming | Real-Time | Maturity |
|---|---|---|---|---|---|---|
| **AMIS-UI** | Yes (JSON schema) | Yes (schema from Go) | Yes | High | Yes | High (Baidu) |
| **React (manual)** | No (code required) | No (JS/TS only) | Yes | Full | Yes | Very High |
| **Retool** | Partial (low-code) | Via API | Limited | Moderate | Yes | High |
| **AppSmith** | Partial (low-code) | Via API | Yes | Moderate | Partial | High |
| **Refine** | No (code required) | No | Yes | High | Yes | High |
| **Budibase** | Yes (low-code) | Via API | Yes | Moderate | Partial | Moderate |

**Why AMIS-UI for Awoerp:**

AMIS's central value for the plugin system is that **plugin authors write Go code that produces JSON schemas, and AMIS renders those schemas as real, interactive UI**. This means:

1. **Plugin authors stay in Go.** A backend Go developer can build a complete plugin — database schema, business logic, API routes, and UI screens — without switching to JavaScript or React.

2. **Schema-driven UI is merge-friendly.** When the Plugin Runtime needs to combine the core UI schema with contributions from multiple plugins (adding tabs, menu items, widgets), it merges JSON objects — a well-defined, deterministic operation. Merging React component trees is not.

3. **AMIS schemas are declarative and auditable.** Reviewing what a plugin will add to the UI is as simple as reading its AMIS schema JSON. There is no React code execution at audit time.

```go
// Plugin contributing a UI tab to the employee detail view
func (p *PayrollPlugin) RegisterUI(registry *sdk.UIExtensionRegistry) {
    // Define an AMIS schema fragment for a new "Commission" tab
    commissionTab := amis.Tab{
        Title: "Commission Settings",
        Body: amis.Form{
            API: amis.API{
                Method: "GET",
                URL:    "/api/v1/plugins/com.example.payroll-ext/commission-settings/${id}",
            },
            Body: []amis.FormItem{
                amis.Select{
                    Name:    "commission_type",
                    Label:   "Commission Type",
                    Options: []amis.Option{
                        {Label: "Tiered", Value: "tiered"},
                        {Label: "Flat Rate", Value: "flat"},
                    },
                },
                amis.Number{
                    Name:  "base_rate",
                    Label: "Base Commission Rate (%)",
                    Min:   0, Max: 100,
                },
            },
        },
    }

    // Register the tab extension
    registry.ExtendTab("employee_detail_tabs", commissionTab)
}
```

The AMIS framework renders this Go-defined schema as a complete, styled, interactive form — with API data loading, validation, and submission handling — automatically.

**Considered alternative — React (manual):** Allowing plugins to contribute arbitrary React components would offer the most flexibility but would require every plugin author to be a JavaScript/React developer, maintain a separate frontend build pipeline, and solve the hard problem of loading third-party JavaScript bundles at runtime securely (import maps, Module Federation). This is the approach taken by systems like Backstage and Grafana plugins — it is powerful but expensive for plugin authors and complex for the host to manage safely.

**Considered alternative — Retool/AppSmith (embedded):** Embedding a commercial low-code tool as the plugin UI surface was evaluated. The concerns were: vendor lock-in for all plugin authors, cost at scale (per-user pricing for some tiers), and the disconnect between the AMIS-native Awoerp frontend and an embedded external tool. These were disqualifying for a v1.0 decision.

### 6.7.3 Schema-Driven UI Extension by Plugins

The Plugin Runtime aggregates all plugin AMIS schema contributions at request time:

1. The frontend shell loads, requesting the full UI schema from `/api/ui/schema`.
2. The Schema Aggregator service queries the Plugin Runtime for all enabled plugins' schema contributions.
3. The Plugin Runtime calls each plugin's `UIContributions()` method.
4. Contributions are merged into the base core schema using a deterministic merge algorithm (deep merge with conflict resolution by plugin priority).
5. The aggregated schema is returned to the frontend, which AMIS renders.

All of this happens server-side in Go. The browser never knows whether a UI element comes from the core or a plugin.

### 6.7.4 What's Lost Without This in v1.0

Without AMIS-UI integration in the plugin system, plugins are invisible to end users. A plugin can add backend logic, API routes, and database tables, but users interact with Awoerp through its UI — if the UI doesn't reflect the plugin's capabilities, those capabilities might as well not exist. The inability to extend the UI from plugins is the single most impactful gap in usability, and it would prevent adoption of all user-facing plugin types.

---

---

# 7. Plugin Registry & Discovery

## 7.1 Local vs. Remote Plugin Registry

The **Plugin Registry** is the catalog of available and installed plugins. Awoerp supports two registry types:

**Local Registry:** A configuration file (`awoerp.plugins.yaml`) that lists the plugins to be loaded, their versions, and their source artifacts. This is the simplest model and is appropriate for single-tenant on-premise deployments where the plugin set is manually controlled.

**Remote Registry:** A hosted service (the Awoerp Plugin Registry) that provides an API for discovering, searching, installing, and updating plugins. The remote registry stores plugin artifacts (compiled binaries or WASM modules), manifests, signatures, and metadata. The Plugin Runtime polls or subscribes to the remote registry for updates.

In a multi-tenant SaaS deployment, each tenant has their own plugin subscription list in the remote registry. The Plugin Runtime loads tenant-appropriate plugins on a per-tenant basis.

## 7.2 Plugin Signing & Trust Chain

Every plugin artifact in the remote registry is **signed** by its author using a private key. The corresponding public key is registered with the registry at publish time. When the Plugin Runtime downloads a plugin artifact, it verifies the signature before loading it.

The trust chain is:

```
Awoerp Registry CA → Author Certificate → Plugin Artifact Signature
```

For enterprise deployments with private registries, the trust chain can be extended with a company-internal CA, allowing enterprises to sign and distribute internal plugins through their own registry without involving the public Awoerp Registry.

Unsigned plugins are rejected by default. This policy can be overridden for development environments via the `AWOERP_PLUGIN_TRUST_UNSIGNED=true` environment variable.

## 7.3 Registry API Design

The Plugin Registry exposes a standard REST API:

```
GET  /registry/v1/plugins                     # Search/list plugins
GET  /registry/v1/plugins/{id}                # Get plugin metadata
GET  /registry/v1/plugins/{id}/versions       # List available versions
GET  /registry/v1/plugins/{id}/{version}      # Get specific version metadata
GET  /registry/v1/plugins/{id}/{version}/artifact  # Download artifact
POST /registry/v1/plugins                     # Publish a new plugin (authenticated)
POST /registry/v1/plugins/{id}/versions       # Publish a new version
DELETE /registry/v1/plugins/{id}/{version}    # Yank a version (author/admin only)

GET  /registry/v1/tenants/{tenantID}/plugins  # List tenant's installed plugins
POST /registry/v1/tenants/{tenantID}/plugins  # Install a plugin for a tenant
DELETE /registry/v1/tenants/{tenantID}/plugins/{id} # Uninstall
```

## 7.4 Comparison: NPM-style, OCI/Helm-style, Backstage Plugin Registry, VSCode Marketplace

| Registry Model | Content Addressing | Signing | Access Control | Go Native | Notes |
|---|---|---|---|---|---|
| **NPM-style** | By name@version | Optional | Public/private | No | Familiar to devs; not ideal for binary artifacts |
| **OCI/Helm-style** | By digest (SHA256) | Yes (Notary/Sigstore) | Registry-level | Yes | Strong content addressing; overkill for most plugins |
| **Backstage Plugin Registry** | npm-based | npm signing | Public | No (React-only) | Good model for documentation; JS-only |
| **VSCode Marketplace** | Proprietary | Microsoft CA | Publisher-gated | No | Excellent UX reference; closed ecosystem |
| **Awoerp Registry** | By ID + SemVer | Author + CA | Tenant-scoped | Yes | Hybrid: NPM-style UX, OCI-style artifact integrity |

The Awoerp Registry borrows the **developer experience** from NPM (simple version references, familiar semver, tag-based searching) and the **artifact integrity** model from OCI (content-addressed storage, signed artifacts, immutable digests). The result is a registry that plugin authors find intuitive and administrators find trustworthy.

---

---

# 8. Security Model

## 8.1 Threat Model for a Public Plugin System

Opening the system to public plugin authors introduces several threat categories that do not exist in a closed system:

| Threat | Description | Mitigation |
|---|---|---|
| **Malicious plugin** | A plugin intentionally designed to exfiltrate tenant data, mine cryptocurrency, or corrupt the host | Process isolation, capability model, code signing, registry vetting |
| **Buggy plugin** | A plugin with unintentional bugs that corrupt data, crash the host, or create performance issues | Process isolation, circuit breakers, resource limits, health checks |
| **Cross-tenant data leak** | A plugin accessing data belonging to a different tenant | PostgreSQL RLS, automatic tenant scoping in SDK, no direct DB access without SDK |
| **Privilege escalation** | A plugin attempting to access resources beyond its declared permissions | Manifest-enforced capability model, runtime permission checks |
| **Supply chain attack** | A legitimate plugin being compromised after publication | Immutable artifact digests, author signing, version pinning |
| **Dependency confusion** | A malicious package with the same name as an internal plugin published to the public registry | Namespace reservation (reverse-domain IDs), registry access controls |

## 8.2 Sandboxing Options in Go

The Plugin Runtime's isolation model was selected from four evaluated options:

| Isolation Model | Overhead | Safety | Complexity | Go Support | v1.0 Decision |
|---|---|---|---|---|---|
| **gRPC subprocess** | ~0.2ms/call | High (process boundary) | Moderate | Excellent | ✅ Default for third-party |
| **Go plugin (.so)** | ~0.01ms/call | Low (same address space) | Low | Good but limited | ✅ In-process for trusted |
| **WebAssembly (WASM)** | ~1-5ms/call | Very High (VM boundary) | High | Emerging | ❌ Deferred to v2.0 |
| **gVisor/container** | ~0.5ms/call | Very High (kernel boundary) | Very High | Works | ❌ Too heavy for plugin-level |

The **gRPC subprocess** model (default) provides a genuine OS process boundary: the plugin runs in a separate process, and communication happens over a Unix domain socket. The plugin cannot access the host's memory, cannot call host functions directly, and cannot survive the host's process termination or vice versa.

The **Go plugin (.so)** model uses Go's built-in `plugin` package to load compiled `.so` files into the host process. This is the fastest option but provides no memory isolation. It is reserved for first-party plugins (Awoerp-developed modules) that have been fully audited.

## 8.3 Permission & Capability Model

The permission model is enforced at three layers:

1. **Manifest validation (registration time):** The Plugin Runtime validates that the manifest's declared permissions are within the allowed scope for the plugin's trust level (community, verified, official).

2. **PluginContext construction (initialization time):** The SDK constructs a `PluginContext` that only exposes handles for declared capabilities. Attempting to use an undeclared capability panics with a clear developer-facing error.

3. **Runtime enforcement (execution time):** All SDK calls that interact with core services include a runtime permission check against the plugin's registered capabilities. This provides defense-in-depth: even if a plugin somehow obtained a handle to a service it didn't declare, the service itself would reject the call.

## 8.4 Plugin Code Signing & Verification

```
┌─────────────────────────────────────────────────────────────┐
│                    SIGNING PIPELINE                          │
│                                                              │
│  Plugin Source → Build → Artifact (.tar.gz)                  │
│                              │                               │
│                              ▼                               │
│              awoerp-plugin sign --key author.key             │
│                              │                               │
│                              ▼                               │
│                   Signed Artifact + Signature                │
│                              │                               │
│                              ▼                               │
│              awoerp-plugin publish → Registry                │
│                                                              │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   VERIFICATION PIPELINE                      │
│                                                              │
│  Plugin Runtime → Download artifact from Registry            │
│                              │                               │
│                              ▼                               │
│           Verify signature against Registry CA trust chain   │
│                              │                               │
│                              ▼                               │
│       Verify artifact digest matches registry-stored digest  │
│                              │                               │
│                              ▼                               │
│              Load plugin / Reject with error log             │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## 8.5 Tenant Isolation in Multi-Tenant Deployments

Tenant isolation is enforced at every layer:

- **HTTP layer:** All plugin route handlers receive a request context containing the authenticated tenant ID. The SDK's router middleware rejects requests without a valid tenant context.
- **Database layer:** PostgreSQL RLS policies ensure every query is automatically filtered to the current tenant. The SDK's DB handle sets `app.current_tenant_id` as a session variable on every connection checkout.
- **Event layer:** Redis Stream keys are namespaced by tenant ID. A plugin subscribed to `invoice.finalized` for Tenant A cannot receive events published by Tenant B.
- **Cache layer:** All cache keys are automatically prefixed with the tenant ID and plugin ID. Cache poisoning across tenants is architecturally impossible.
- **Workflow layer:** Temporal workflows are tagged with the tenant ID. Workflow queries and signals are scoped to the correct tenant context.

## 8.6 Audit Logging for Plugin Actions

Every plugin action that modifies data — database writes, event publications, workflow starts — is automatically recorded in the `audit_log` table:

```sql
CREATE TABLE public.audit_log (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL,
    plugin_id    TEXT,          -- NULL for core actions
    user_id      UUID,
    action       TEXT NOT NULL, -- e.g., "create", "update", "delete", "workflow.start"
    resource_type TEXT NOT NULL,
    resource_id  TEXT,
    diff         JSONB,         -- Before/after for updates
    metadata     JSONB,         -- Plugin-specific context
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

This log is append-only (enforced by PostgreSQL row-level security: no UPDATE/DELETE permissions) and is the source of truth for compliance reporting.

---

---

# 9. Plugin Development Guide

## 9.1 Prerequisites & Tooling Setup

To build an Awoerp plugin, you need:

```bash
# Go 1.22 or later
go version  # go version go1.22.0 linux/amd64

# Awoerp Plugin CLI
go install awoerp.io/cli/awoerp-plugin@latest

# SQLC (for database query generation)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Temporal CLI (for local workflow testing)
brew install temporal  # or see temporal.io/docs/cli

# A running Awoerp development environment
# (Docker Compose file provided in the Awoerp dev-env repo)
docker compose -f awoerp-dev-env/docker-compose.yml up -d
```

## 9.2 Scaffolding Your First Plugin

```bash
# Scaffold a new plugin
awoerp-plugin init com.example.my-first-plugin \
    --name "My First Plugin" \
    --author "Jane Developer <jane@example.com>" \
    --license MIT

# This generates:
# my-first-plugin/
# ├── go.mod
# ├── go.sum
# ├── main.go           # Plugin entry point and interface implementation
# ├── plugin.go         # Core plugin logic
# ├── manifest.yaml     # Plugin manifest
# ├── sqlc.yaml         # SQLC configuration
# ├── migrations/
# │   └── 001_initial.sql
# ├── queries/
# │   └── example.sql
# ├── internal/
# │   └── db/          # Generated by SQLC — do not edit
# └── plugin_test.go
```

## 9.3 Registering Extension Points

A complete minimal plugin implementation:

```go
// plugin.go
package main

import (
    "context"
    sdk "awoerp.io/sdk/v1"
)

type MyFirstPlugin struct {
    db     *sdk.DBHandle
    events *sdk.EventBus
}

// Metadata implements sdk.Plugin
func (p *MyFirstPlugin) Metadata() sdk.PluginMetadata {
    return sdk.PluginMetadata{
        ID:          "com.example.my-first-plugin",
        Name:        "My First Plugin",
        Version:     "1.0.0",
        Description: "A demonstration plugin",
        Author:      "Jane Developer",
        License:     "MIT",
    }
}

// Init implements sdk.Plugin
func (p *MyFirstPlugin) Init(ctx context.Context, pc *sdk.PluginContext) error {
    p.db = pc.DB
    p.events = pc.Events

    // Run database migrations
    if err := pc.DB.RunMigrations(ctx, migrationFiles); err != nil {
        return fmt.Errorf("migrations failed: %w", err)
    }

    // Register HTTP routes
    pc.Router.GET("/hello", p.handleHello)

    // Subscribe to core events
    pc.Events.Subscribe("invoice.finalized", p.onInvoiceFinalized)

    // Register UI extension
    pc.UI.ExtendTab("customer_detail_tabs", p.buildCustomerTab())

    return nil
}

// HealthCheck implements sdk.Plugin
func (p *MyFirstPlugin) HealthCheck(ctx context.Context) error {
    return p.db.Ping(ctx)
}

// Shutdown implements sdk.Plugin
func (p *MyFirstPlugin) Shutdown(ctx context.Context) error {
    // Clean up any open connections, flush buffers, etc.
    return nil
}
```

## 9.4 Interacting with Core Services

```go
// Reading from a core table (read access must be declared in manifest)
func (p *MyFirstPlugin) getCustomer(ctx context.Context, customerID uuid.UUID) (*CoreCustomer, error) {
    // SDK provides type-safe access to permitted core tables
    return p.db.Core.GetCustomer(ctx, customerID)
}

// Writing to plugin's own tables (SQLC-generated)
func (p *MyFirstPlugin) saveRecord(ctx context.Context, data CreateRecordParams) error {
    _, err := p.queries.CreateRecord(ctx, data)
    return err
}

// Publishing an event
func (p *MyFirstPlugin) notifyExternalSystem(ctx context.Context, payload interface{}) error {
    return p.events.Publish(ctx, "my-first-plugin.record.created", payload)
}
```

## 9.5 Writing Plugin Tests

The SDK provides a test harness that simulates the Plugin Runtime without requiring a running Awoerp instance:

```go
// plugin_test.go
func TestInvoiceFinalizedHandler(t *testing.T) {
    // Create a test harness with an in-memory SQLite (or test PostgreSQL) database
    harness := sdktest.NewHarness(t, &MyFirstPlugin{})
    harness.SetTenantID("test-tenant-001")

    // Initialize the plugin
    require.NoError(t, harness.Init(context.Background()))

    // Publish a test event
    harness.PublishEvent("invoice.finalized", &sdk.InvoiceEvent{
        Invoice: &Invoice{ID: uuid.New(), Total: 5000},
    })

    // Assert the expected side effect occurred
    harness.AssertEventPublished(t, "my-first-plugin.record.created")
    harness.AssertDBRowExists(t, "com_example_my_first_plugin.records", "invoice_id = $1", invoiceID)
}
```

## 9.6 Submitting to the Registry

```bash
# Build and sign the plugin artifact
awoerp-plugin build
awoerp-plugin sign --key ~/.awoerp/author.key

# Publish to the registry (requires a registered author account)
awoerp-plugin publish --registry https://registry.awoerp.io

# Or publish to a private enterprise registry
awoerp-plugin publish --registry https://plugins.yourcompany.com
```

---

---

# 10. Core System Integration Guide

## 10.1 Exposing New Extension Points

When a core Awoerp developer adds a new feature to the core system, they should ask: **"Should this feature's behavior be extensible by plugins?"** If yes, they need to expose an extension point.

Adding a new hook to the core:

```go
// In the core invoice service
type InvoiceService struct {
    // ...
    hooks *sdk.HookRegistry
}

func (s *InvoiceService) FinalizeInvoice(ctx context.Context, invoice *Invoice) error {
    // Pre-finalize hooks — synchronous, can abort the operation
    if err := s.hooks.Call(ctx, "invoice.pre_finalize", &InvoiceHookPayload{Invoice: invoice}); err != nil {
        return err  // Hook returned an error — operation aborted
    }

    // Core finalization logic
    if err := s.db.UpdateInvoiceStatus(ctx, invoice.ID, "finalized"); err != nil {
        return err
    }

    // Post-finalize event — asynchronous, fire and forget
    s.events.Publish(ctx, "invoice.finalized", &InvoiceFinalizedEvent{InvoiceID: invoice.ID})

    return nil
}
```

Every new hook and event must be:
- Documented in the extension point registry (`docs/extension-points.md`).
- Added to the compatibility changelog.
- Covered by a test that ensures the hook is called with the correct payload.

## 10.2 Managing Breaking Changes in the Plugin API

When a breaking change to the plugin API is unavoidable:

1. **Announce** the change in the release notes with a minimum 2-minor-version deprecation period.
2. **Provide a compatibility shim** in the SDK that translates old call signatures to new ones.
3. **Bump the major SDK version** (`v1` → `v2`) and update the module path.
4. **Run the migration guide** through the developer relations team before release.

Breaking changes to the core plugin `Plugin` interface are the most severe and should require a full RFC (Request for Comments) process with community input.

## 10.3 Plugin Load Order & Dependency Resolution

The Plugin Runtime resolves plugin dependencies using a topological sort on the plugin dependency graph:

```
Plugin A depends on Plugin B → B initializes before A
Plugin C depends on Plugin B → B initializes before C
Plugin A and Plugin C are independent → order between them is deterministic (alphabetical by ID)
```

Circular dependencies are detected during registration and cause both plugins to be rejected with a clear error message.

## 10.4 Performance Budgeting for Plugin Overhead

Each synchronous hook call adds latency to the core operation that triggers it. The Plugin Runtime enforces a **per-hook timeout** (default: 100ms for synchronous hooks) and a **per-request aggregate plugin overhead budget** (default: 500ms). If a plugin exceeds these thresholds:

- First offense: warning logged, plugin health score decremented.
- Repeated offenses: plugin is temporarily suspended and administrators are alerted.
- Persistent offenses: plugin is disabled and marked as incompatible.

Plugin authors can view their plugin's latency contribution in the Awoerp admin panel under Plugins → Performance.

---

---

# 11. Deployment & Operations

## 11.1 Plugin Deployment Modes

| Mode | Description | When to Use |
|---|---|---|
| **Bundled** | Plugin binary embedded in the Awoerp Docker image | Controlled internal plugins, air-gapped environments |
| **Sidecar** | Plugin runs as a separate container in the same Kubernetes Pod, communicating via Unix socket mounted on a shared volume | Third-party plugins in containerized deployments |
| **Remote** | Plugin runs as a standalone service reachable over the network via gRPC | Large enterprise plugins with their own scaling needs |

The **sidecar model** is recommended for most production plugin deployments. It provides process isolation without the network overhead of fully remote plugins, and Kubernetes manages the plugin process lifecycle alongside the Awoerp host.

```yaml
# Kubernetes Pod spec for sidecar plugin deployment
spec:
  containers:
    - name: awoerp
      image: awoerp/core:1.0.0
      volumeMounts:
        - name: plugin-socket
          mountPath: /var/run/plugins

    - name: logistics-plugin
      image: example/awoerp-logistics-plugin:2.1.0
      env:
        - name: AWOERP_PLUGIN_SOCKET
          value: /var/run/plugins/logistics.sock
      volumeMounts:
        - name: plugin-socket
          mountPath: /var/run/plugins

  volumes:
    - name: plugin-socket
      emptyDir: {}
```

## 11.2 Health Checks & Circuit Breakers

The Plugin Runtime performs health checks on each plugin every 30 seconds (configurable). If a plugin fails three consecutive health checks:

1. The plugin is marked as `degraded` and administrators are alerted.
2. The circuit breaker opens: new requests to the plugin's routes return `503 Service Unavailable`.
3. The runtime attempts to restart the plugin process (for subprocess plugins).
4. After a successful restart and three consecutive health check passes, the plugin returns to `healthy` status and the circuit breaker closes.

## 11.3 Observability: Metrics, Tracing & Logs per Plugin

Every plugin is automatically instrumented with:

**Metrics (Prometheus):**
- `awoerp_plugin_request_duration_seconds{plugin_id, route}` — HTTP handler latency
- `awoerp_plugin_event_processing_duration_seconds{plugin_id, event_type}` — Event handler latency
- `awoerp_plugin_db_query_duration_seconds{plugin_id, query_name}` — Database query latency
- `awoerp_plugin_errors_total{plugin_id, error_type}` — Error counts

**Tracing (OpenTelemetry):**
All plugin calls are automatically wrapped in OpenTelemetry spans. When a plugin is called as part of a request trace, the span is a child of the core request span. This means end-to-end request traces in Jaeger or Tempo show exactly which plugins contributed to the total request latency.

**Logs (Structured JSON):**
The Plugin SDK's `Logger` produces structured JSON log lines with automatic fields: `plugin_id`, `tenant_id`, `trace_id`. All plugin logs are routed through the core's log aggregation pipeline (Loki, CloudWatch, etc.) and can be queried with plugin-specific filters.

## 11.4 Hot Reload vs. Rolling Restart Strategy

**v1.0: Rolling Restart**
Plugin updates in v1.0 require a rolling restart of the Awoerp service. The deployment process is:
1. Update the plugin version in `awoerp.plugins.yaml`.
2. Trigger a rolling restart of the Awoerp deployment.
3. Each new instance starts with the updated plugin; old instances drain gracefully.

**v1.1: Hot Reload (Planned)**
Hot reload allows a plugin to be updated without restarting the host process:
1. New plugin artifact is pushed to the registry.
2. Plugin Runtime detects the new version.
3. Runtime initializes the new plugin version in parallel with the old.
4. Once new version is healthy, traffic is cut over and old version is shutdown.
5. No host process restart; no downtime.

Hot reload is deferred from v1.0 due to the complexity of ensuring safe in-process state transition for in-process plugins and the need to drain in-flight requests before switching plugin versions.

---

---

# 12. Comparison to Existing Plugin Systems

## 12.1 Open Source: Odoo Modules, ERPNext Apps, Frappe Framework

### Odoo Modules

Odoo's module system is one of the most mature plugin systems in the open-source ERP world. Odoo modules can extend any part of the system: models (database tables), views (UI), business logic, and reports. Every piece of Odoo's functionality — including core ERP features — is delivered as a module.

**Strengths:** Extremely powerful; modules can override any core behavior. Large ecosystem (10,000+ modules on Odoo Apps marketplace). Deeply integrated theming.

**Weaknesses:** Requires Python expertise. Modules run in the same process with no isolation — a buggy module can crash the entire Odoo instance. The ORM coupling makes modules fragile across major version upgrades. UI extension requires QWeb (Odoo's XML-based template language) knowledge in addition to Python. No Go support.

**Awoerp advantage:** Process isolation, Go-native development, manifest-declared permissions (Odoo modules can access any model they want), and AMIS schema-driven UI extension that doesn't require separate template language knowledge.

### ERPNext / Frappe Framework

Frappe is the Python web framework underlying ERPNext. Frappe Apps are ERPNext's plugin equivalent. Like Odoo, they can extend models, views, and business logic. Frappe's "hooks" system is conceptually similar to Awoerp's hook system.

**Strengths:** Clean separation between framework (Frappe) and application (ERPNext). Good developer documentation. Bench tool simplifies multi-site deployment.

**Weaknesses:** Requires full server access to install apps (not suitable for SaaS deployment without custom infrastructure). Python/Jinja2/Vue knowledge required. No process isolation. Schema extensions use Frappe's DocType system rather than raw SQL, which limits database design flexibility.

**Awoerp advantage:** SaaS-friendly deployment model (tenant-scoped plugin installation without server access), process isolation, SQL-native data layer via SQLC, and Go performance characteristics.

---

## 12.2 Proprietary: SAP BTP Extensions, Oracle APEX Plugins, Salesforce AppExchange

### SAP BTP (Business Technology Platform)

SAP BTP is SAP's cloud extensibility platform. It allows developers to build extensions to SAP S/4HANA and other SAP products using side-by-side extensibility (building separate applications that connect to SAP via APIs) rather than in-core extensibility (modifying SAP's code).

**Strengths:** Enterprise-grade security and compliance. Large partner ecosystem. Strong governance model for extensions.

**Weaknesses:** Extremely complex and expensive to develop for. BTP development requires significant SAP-specific knowledge (CAP framework, SAP Fiori, SAP Cloud Connector). Extensions are fundamentally separate applications — there is no schema-level or in-process extensibility. The UX integration between SAP and an extension is necessarily limited.

**Awoerp advantage:** Plugin authors build with standard Go tooling. The plugin system supports genuine in-system extensibility (UI integration, database schema extension, in-process hooks) rather than just side-by-side external applications.

### Salesforce AppExchange

AppExchange is the most commercially successful ERP/CRM app marketplace in existence. Salesforce's metadata-driven architecture (Apex code, Visualforce/LWC components, custom objects) makes extensions deeply integrated with the host system.

**Strengths:** Enormous marketplace. Good integration depth. Strong multi-tenant isolation model (each Org is isolated). Mature developer tooling (Salesforce CLI, scratch orgs).

**Weaknesses:** Proprietary everything — Apex language, SOQL query language, Visualforce/LWC component model. No open standards. Extensions can only run within the Salesforce platform. Significant licensing costs.

**Awoerp advantage:** Standard Go tooling, open standards (gRPC, PostgreSQL, Redis, standard HTTP), self-hostable, no vendor lock-in for extension authors.

---

## 12.3 Developer Platforms: Shopify Apps, Stripe Apps, Atlassian Forge

### Shopify Apps

The Shopify App ecosystem is arguably the best model for what Awoerp's plugin system aspires to be. Shopify provides:
- A stable, versioned Admin API and Storefront API.
- A CLI for scaffolding and deploying apps.
- An app marketplace with install/uninstall flows.
- Webhooks for all significant platform events.
- App extensions for UI injection (App Bridge).

**Strengths:** Developer experience is excellent. The App Bridge model for UI injection is conceptually similar to Awoerp's AMIS schema extension. Strong versioning guarantees.

**Weaknesses:** Shopify apps are separate hosted applications that integrate via API — they cannot extend Shopify's database schema or run in-process. UI extensions are limited to predefined injection points. JavaScript/React required for UI extensions.

**Awoerp advantage:** Schema-level extensibility (plugins can add database tables), in-process hooks (Shopify has no equivalent for synchronous pre-action hooks), Go-native UI extension via AMIS (no JavaScript required).

### Atlassian Forge

Forge is Atlassian's cloud-first development platform for Jira, Confluence, and other Atlassian products. Forge apps run as serverless functions on Atlassian's infrastructure, invoked by the Atlassian host when an extension point is triggered.

**Strengths:** Zero-ops for extension authors (Atlassian manages the runtime). Strong sandboxing (each function invocation is isolated). UI extensions via a React-based component kit.

**Weaknesses:** Vendor-managed runtime means no self-hosting. Cold start latency for serverless functions. JavaScript/React required. Limited to Atlassian's predefined extension points.

**Awoerp advantage:** Self-hostable, no cold-start latency (plugins are persistent processes), Go-native, broader extension point coverage (schema, workflows, etc.).

---

## 12.4 Awoerp Positioning Matrix

The following matrix positions Awoerp's plugin system against the comparison systems across the dimensions that matter most:

| Capability | Awoerp | Odoo | ERPNext | SAP BTP | Salesforce | Shopify |
|---|---|---|---|---|---|---|
| **Process isolation** | ✅ gRPC subprocess | ❌ In-process | ❌ In-process | ✅ Separate app | ✅ Apex sandbox | ✅ Separate app |
| **Schema extensibility** | ✅ Namespaced schema | ✅ ORM-based | ✅ DocType | ❌ API only | ✅ Custom objects | ❌ API only |
| **UI extension (no JS)** | ✅ AMIS schema | ❌ QWeb required | ❌ Jinja required | ❌ Fiori required | ❌ LWC required | ❌ React required |
| **Go-native SDK** | ✅ | ❌ Python | ❌ Python | ❌ Java/JS | ❌ Apex | ❌ Node/Ruby |
| **Durable workflows** | ✅ Temporal | ❌ Manual | ❌ Manual | ✅ SAP Workflow | ✅ Flow | ❌ |
| **Self-hostable** | ✅ | ✅ | ✅ | ❌ Cloud only | ❌ Cloud only | ❌ Cloud only |
| **Tenant-scoped install** | ✅ | ❌ Whole-system | ❌ Whole-system | N/A | ✅ Org-scoped | ✅ Store-scoped |
| **Open standards** | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Marketplace** | ✅ (planned) | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Manifest-declared permissions** | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ |

Awoerp's plugin system is unique in combining **schema extensibility, process isolation, durable workflow support, schema-driven UI extension without JavaScript, and a Go-native SDK** in a single, self-hostable platform. No comparable system offers all five of these simultaneously.

---

---

# 13. v1.0 Scope, Deferred Features & Roadmap

## 13.1 What Must Ship in v1.0 (Non-Negotiable)

The following capabilities are required in v1.0. Shipping without any of these would create adoption blockers that are unacceptable for the reasons documented in their respective sections:

| Capability | Reason It Cannot Be Deferred |
|---|---|
| **Core Plugin interface** | Every other capability depends on it. Changing it later is a breaking change for all plugins. |
| **Plugin manifest schema + validation** | Required for security; permissions cannot be enforced without declared intent. |
| **gRPC subprocess isolation** | Required for third-party plugin safety. In-process only = cannot allow public plugins. |
| **Database schema namespacing + RLS** | Required for multi-tenant safety. Not shipping = cannot guarantee data isolation. |
| **Fiber route mounting** | Without HTTP route extension, integration plugins cannot be built. |
| **Redis event bus** | Required for asynchronous plugin reactions; without it, all hooks must be synchronous (performance impact). |
| **Temporal workflow registration** | Without this, long-running process plugins cannot be built reliably. |
| **AMIS UI extension** | Without this, plugins are invisible to users. Adoption will be near-zero for user-facing use cases. |
| **Plugin SDK v1.0** | Without an SDK, plugin authorship is too complex for widespread adoption. |
| **Plugin signing + verification** | Required for public plugin security. Unsigned plugins cannot be trusted. |
| **Plugin Registry (basic)** | Required for plugin distribution. Without it, plugins must be manually installed on every server. |
| **Tenant-scoped plugin installation** | Required for SaaS deployment model. Without it, plugins affect all tenants or none. |

## 13.2 What Can Be Deferred and at What Cost

| Capability | Deferral Version | Cost of Deferring |
|---|---|---|
| **Hot reload** | v1.1 | Updates require rolling restart. Acceptable for v1.0; annoying but not blocking. |
| **WASM sandboxing** | v2.0 | gRPC subprocess provides adequate isolation for v1.0. |
| **Plugin marketplace UI** | v1.1 | Registry API exists; marketplace browse UI can be post-launch. |
| **Private enterprise registry** | v1.1 | Enterprise deals may stall without this; plan to ship early in v1.x. |
| **Plugin performance analytics in admin UI** | v1.1 | Prometheus metrics exist; dedicated UI is UX improvement, not functional blocker. |
| **Automatic plugin dependency resolution** | v1.1 | Manual dependency ordering in config is acceptable for v1.0. |
| **Plugin-to-plugin direct communication** | v1.2 | Event bus covers most P2P use cases for v1.0. |
| **Multi-language plugin SDK (Python, TS)** | v2.0 | Go SDK covers v1.0; broader language support expands ecosystem in v2.0. |

## 13.3 Roadmap: v1.1 → v2.0 Plugin System Evolution

```
v1.0  ─── Core plugin system, SDK, registry, all extension types
  │
  ├── v1.1 ── Hot reload, marketplace UI, enterprise registry,
  │           automatic dependency resolution, performance analytics
  │
  ├── v1.2 ── Plugin-to-plugin direct messaging,
  │           Plugin capability delegation (plugin A grants permission to plugin B),
  │           GraphQL extension points,
  │           Plugin-specific custom RBAC roles
  │
  ├── v1.3 ── Plugin monetization infrastructure (billing, usage metering),
  │           Plugin review and certification program,
  │           A/B testing framework for plugin UI extensions
  │
  └── v2.0 ── WASM sandboxing for enhanced isolation,
              Multi-language SDKs (Python, TypeScript),
              Event sourcing integration (plugins as event stream processors),
              AI-assisted plugin scaffolding,
              Visual plugin builder (no-code plugin creation for simple use cases)
```

---

---

# 14. Appendices

## Appendix A: Full Plugin Interface Go Code Reference

```go
// Package sdk — awoerp.io/sdk/v1
// This is the complete, canonical plugin interface definition for Awoerp v1.0.
// This file is the authoritative reference; the commentary in the main
// documentation derives from this source.

package sdk

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
)

// ─────────────────────────────────────────────────────
// Core Plugin Interface
// ─────────────────────────────────────────────────────

type Plugin interface {
    Metadata() PluginMetadata
    Init(ctx context.Context, pc *PluginContext) error
    HealthCheck(ctx context.Context) error
    Shutdown(ctx context.Context) error
}

type PluginMetadata struct {
    ID          string
    Name        string
    Version     string
    Description string
    Author      string
    License     string
    Homepage    string
    Tags        []string
}

// ─────────────────────────────────────────────────────
// Plugin Context
// ─────────────────────────────────────────────────────

type PluginContext struct {
    TenantID  string
    DB        *DBHandle
    Router    *Router
    Events    *EventBus
    Workflows *WorkflowRegistry
    UI        *UIExtensionRegistry
    Cache     *CacheHandle
    Logger    *Logger
    Config    *ConfigHandle
}

// ─────────────────────────────────────────────────────
// Database Handle
// ─────────────────────────────────────────────────────

type DBHandle struct {
    pool   *pgxpool.Pool
    tenant string
    plugin string
}

func (db *DBHandle) RunMigrations(ctx context.Context, migrations embed.FS) error
func (db *DBHandle) Ping(ctx context.Context) error

// Core provides read-only access to declared core tables
type CoreReadHandle struct{}
func (db *DBHandle) Core() *CoreReadHandle

// ─────────────────────────────────────────────────────
// Router
// ─────────────────────────────────────────────────────

type Router struct{ /* Fiber sub-app wrapper */ }

func (r *Router) GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc)
func (r *Router) POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc)
func (r *Router) PUT(path string, handler HandlerFunc, middleware ...MiddlewareFunc)
func (r *Router) PATCH(path string, handler HandlerFunc, middleware ...MiddlewareFunc)
func (r *Router) DELETE(path string, handler HandlerFunc, middleware ...MiddlewareFunc)
func (r *Router) Use(middleware ...MiddlewareFunc)

type HandlerFunc func(ctx context.Context, req *Request, res *Response) error
type MiddlewareFunc func(ctx context.Context, req *Request, res *Response, next func() error) error

// ─────────────────────────────────────────────────────
// Event Bus
// ─────────────────────────────────────────────────────

type EventBus struct{ /* Redis Streams wrapper */ }

type EventHandler func(ctx context.Context, event *Event) error

func (eb *EventBus) Subscribe(eventType string, handler EventHandler) error
func (eb *EventBus) Publish(ctx context.Context, eventType string, payload interface{}) error

type Event struct {
    ID        string
    Type      string
    TenantID  string
    PublishedAt time.Time
    payload   []byte
}

func (e *Event) UnmarshalPayload(dest interface{}) error

// ─────────────────────────────────────────────────────
// Workflow Registry
// ─────────────────────────────────────────────────────

type WorkflowRegistry struct {
    temporalClient client.Client
    worker         worker.Worker
    pluginID       string
}

func (wr *WorkflowRegistry) Register(workflowFunc interface{})
func (wr *WorkflowRegistry) RegisterActivity(activityFunc interface{})
func (wr *WorkflowRegistry) Start(ctx context.Context, workflowType string, input interface{}) (client.WorkflowRun, error)
func (wr *WorkflowRegistry) Signal(ctx context.Context, workflowID, signalName string, payload interface{}) error

// ─────────────────────────────────────────────────────
// UI Extension Registry
// ─────────────────────────────────────────────────────

type UIExtensionRegistry struct{}

func (ui *UIExtensionRegistry) ExtendTab(location string, tab AMISTab) error
func (ui *UIExtensionRegistry) ExtendMenu(location string, item AMISNavItem) error
func (ui *UIExtensionRegistry) AddPage(path string, page AMISPage) error
func (ui *UIExtensionRegistry) ExtendForm(location string, fields []AMISFormItem) error
func (ui *UIExtensionRegistry) AddDashboardWidget(widget AMISWidget) error

// ─────────────────────────────────────────────────────
// Cache Handle
// ─────────────────────────────────────────────────────

type CacheHandle struct{ /* Redis wrapper with key namespacing */ }

func (c *CacheHandle) Get(ctx context.Context, key string, dest interface{}) error
func (c *CacheHandle) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
func (c *CacheHandle) Delete(ctx context.Context, key string) error
func (c *CacheHandle) AcquireLock(ctx context.Context, key string, ttl time.Duration) (*Lock, error)

// ─────────────────────────────────────────────────────
// Plugin Errors
// ─────────────────────────────────────────────────────

type PluginErrorType string

const (
    ErrValidation  PluginErrorType = "validation"
    ErrPermission  PluginErrorType = "permission"
    ErrRetryable   PluginErrorType = "retryable"
    ErrFatal       PluginErrorType = "fatal"
)

type PluginError struct {
    Type    PluginErrorType
    Code    string
    Message string
    Cause   error
}

func (e *PluginError) Error() string { return fmt.Sprintf("[%s] %s: %s", e.Type, e.Code, e.Message) }

func NewValidationError(code, message string) *PluginError
func NewPermissionError(resource string) *PluginError
func NewRetryableError(cause error) *PluginError
func NewFatalError(cause error) *PluginError
```

---

## Appendix B: Plugin Manifest Schema Reference

The full JSON Schema for the plugin manifest is available at:
`https://registry.awoerp.io/schemas/plugin-manifest/v1.json`

Key constraints:
- `metadata.id` must match the regex `^[a-z0-9]+(\.[a-z0-9-]+)+$` (reverse-domain format).
- `metadata.version` must be a valid SemVer string.
- `compatibility.minCoreVersion` and `maxCoreVersion` must be valid SemVer range specifiers.
- `permissions.database.readTables` entries must match `^(public|[a-z_]+)\.[a-z_]+$`.
- `permissions.api.mount` must start with `/api/v1/plugins/` followed by the plugin ID.
- `permissions.events.subscribe` and `publish` entries must be non-empty strings.

---

## Appendix C: Architecture Decision Records (ADRs)

### ADR-001: gRPC over Unix Socket for Plugin Subprocess Communication
**Decision:** Use gRPC over Unix domain sockets (not TCP) for host-plugin subprocess communication.
**Rationale:** Unix sockets have lower latency than TCP loopback (no TCP stack overhead). Security: Unix socket file permissions restrict access to the same host. No port allocation required. **Rejected:** Named pipes (not cross-platform), shared memory (too complex for RPC).

### ADR-002: PostgreSQL Schema Namespacing over Separate Databases
**Decision:** Each plugin gets a schema within the shared PostgreSQL database, not a separate database.
**Rationale:** Schema-level isolation is sufficient for the security model (enforced by RLS). Separate databases would require separate connection pools, separate backup policies, and cross-database JOINs (impossible in PostgreSQL). Schema-level isolation allows plugins to JOIN against core tables within the same transaction.

### ADR-003: AMIS-UI over React Custom Components for Plugin UI Extension
**Decision:** Plugin UI extensions are defined as AMIS JSON schemas served from the Go backend, not as React components bundled with the plugin.
**Rationale:** Eliminates JavaScript as a required skill for plugin authors. Avoids the security risks of loading arbitrary third-party JavaScript in the browser. Schema-driven extensions are merge-safe. **Rejected:** React Module Federation (complex, requires plugin authors to maintain a frontend build pipeline), iframes (poor UX, no style integration).

### ADR-004: WASM Sandboxing Deferred to v2.0
**Decision:** WebAssembly sandbox isolation is not included in v1.0.
**Rationale:** Go's WASM compilation target is still maturing for server-side use. WASM execution overhead for database-heavy plugins (the most common pattern) would be unacceptable without significant optimization. The gRPC subprocess model provides adequate security for v1.0. WASM remains the target for maximum isolation in future versions.

### ADR-005: Redis Streams over PostgreSQL LISTEN/NOTIFY for Event Bus
**Decision:** Use Redis Streams for the plugin event bus rather than PostgreSQL LISTEN/NOTIFY.
**Rationale:** LISTEN/NOTIFY has no message persistence (lost if subscriber is offline), 8KB payload limit, and no consumer group semantics. Redis Streams provide durability, at-least-once delivery, consumer groups, and backpressure. **Rejected:** Kafka (over-engineered for ERP-scale), NATS (additional infrastructure component).

---

## Appendix D: Glossary

| Term | Definition |
|---|---|
| **Activity** | In Temporal, an Activity is a single step in a workflow — an ordinary Go function that performs a unit of work (e.g., sending an email, calling an API). Unlike workflows, activities do not have durable state. |
| **AMIS** | An open-source JSON schema-driven frontend framework developed by Baidu, used as Awoerp's frontend rendering engine. |
| **Circuit Breaker** | A fault-tolerance pattern where a system stops calling a failing component after a threshold of failures, giving it time to recover. |
| **Consumer Group** | A Redis Streams concept where multiple consumers share the processing of a stream, each receiving a subset of messages. Used in Awoerp to ensure each plugin processes each event exactly once. |
| **Extension Point** | A specific, named location in the Awoerp core system where a plugin can inject behavior. |
| **gRPC** | Google Remote Procedure Call — a high-performance RPC framework using Protocol Buffers for serialization. Used for host-plugin subprocess communication. |
| **Hook** | An event-driven extension point; a plugin registers a callback and the core calls it when a specific event occurs. |
| **Manifest** | A YAML file embedded in a plugin artifact that declares the plugin's identity, version, permissions, and dependencies. |
| **Namespace (PostgreSQL)** | A PostgreSQL schema — a named container for database objects (tables, indexes, functions). Used to isolate each plugin's database objects. |
| **Plugin** | A separately-distributed extension unit that interacts with Awoerp exclusively through the defined plugin interface. |
| **Plugin Runtime** | The component of the Awoerp host process responsible for loading, initializing, monitoring, and shutting down plugins. |
| **PGX** | The most feature-complete and performant Go driver for PostgreSQL, used as Awoerp's database client. |
| **RLS (Row-Level Security)** | A PostgreSQL feature that automatically filters rows based on a policy, used to enforce tenant isolation. |
| **SDK** | Software Development Kit — the Go module and CLI tools that plugin authors use to build Awoerp plugins. |
| **SemVer** | Semantic Versioning — a versioning scheme (MAJOR.MINOR.PATCH) with defined compatibility rules. |
| **SQLC** | A Go code generator that produces type-safe Go functions from SQL query files. |
| **Temporal** | An open-source workflow orchestration platform that provides durable, fault-tolerant execution of multi-step processes. |
| **Tenant** | A single isolated organizational customer in a multi-tenant Awoerp deployment. |
| **Workflow** | In Temporal, a Workflow is a durable, fault-tolerant process definition — a Go function that coordinates Activities. Workflows survive infrastructure failures. |

---

## Appendix E: Further Reading & References

**Temporal Workflow Engine**
- Temporal Documentation: https://docs.temporal.io
- "Designing Durable Workflows" — Temporal Engineering Blog
- Cadence/Temporal: Uber Engineering post on the original design

**PostgreSQL Row-Level Security**
- PostgreSQL RLS Documentation: https://www.postgresql.org/docs/current/ddl-rowsecurity.html
- "Multi-Tenancy with PostgreSQL Row-Level Security" — Citus Data

**SQLC**
- SQLC Documentation: https://docs.sqlc.dev
- "Type-Safe SQL in Go" — SQLC blog

**Fiber**
- Fiber Documentation: https://docs.gofiber.io
- fasthttp benchmarks: https://github.com/valyala/fasthttp#http-server-performance-comparison

**AMIS-UI**
- AMIS Documentation: https://aisuda.bce.baidu.com/amis/en-US/docs/index
- AMIS GitHub: https://github.com/baidu/amis

**Plugin System Design References**
- "Designing Extensible Systems" — Michael Feathers, Working Effectively with Legacy Code
- VS Code Extension API Design: https://code.visualstudio.com/api
- Shopify App Bridge documentation
- Backstage Plugin development guide

**Security**
- OWASP Top 10 for API Security
- "The Confused Deputy Problem" — original paper on capability-based security
- Sigstore project for supply chain security: https://www.sigstore.dev

---

*End of Document*

---
> **Document Version:** 1.0-draft
> **Maintained by:** Awoerp Platform Team
> **Review Cycle:** Updated with each minor version release of the plugin system
> **Feedback:** Open an issue at https://github.com/awoerp/awoerp/issues with label `docs/plugin-system`
