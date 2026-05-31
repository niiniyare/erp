# AWO ERP — System Metadata Service
## Comprehensive Domain Knowledge Guide

> **Audience:** Business administrators, frontend developers, backend developers, and platform architects.
> **Scope:** The System Metadata Service and its sub-domains — Configuration, Feature Flags, Operational Metadata, Module Registry, Entity & Schema Metadata, and Custom Fields — their relationship to each other, how the rest of the system depends on them, and clear guidance for each audience.
> **Note:** Code examples and file paths appear in developer-facing sections only. Non-technical sections use plain language throughout.

---

## Table of Contents

- [Part I — Understanding the Domain](#part-i--understanding-the-domain)
  - [1. Executive Summary](#1-executive-summary)
  - [2. The Problem System Metadata Solves](#2-the-problem-system-metadata-solves)
  - [3. System Metadata in the ERP Landscape](#3-system-metadata-in-the-erp-landscape)
- [Part II — Domain Model](#part-ii--domain-model)
  - [4. Core Concepts and Vocabulary](#4-core-concepts-and-vocabulary)
  - [5. The Sub-Domains and How They Relate](#5-the-sub-domains-and-how-they-relate)
  - [6. Configuration Module — Deep Dive](#6-configuration-module--deep-dive)
  - [7. Feature Flag Module — Deep Dive](#7-feature-flag-module--deep-dive)
  - [8. Operational Metadata — Deep Dive](#8-operational-metadata--deep-dive)
  - [9. Module Registry — Deep Dive](#9-module-registry--deep-dive)
  - [10. Entity & Schema Metadata — Deep Dive](#10-entity--schema-metadata--deep-dive)
  - [11. Custom Fields Metadata — Deep Dive](#11-custom-fields-metadata--deep-dive)
  - [12. Metadata Versioning](#12-metadata-versioning)
- [Part III — How the System Uses Metadata](#part-iii--how-the-system-uses-metadata)
  - [13. The UI Rendering Pipeline and Metadata Gating](#13-the-ui-rendering-pipeline-and-metadata-gating)
  - [14. System Metadata as a Cross-Cutting Concern](#14-system-metadata-as-a-cross-cutting-concern)
- [Part IV — Guidance by Audience](#part-iv--guidance-by-audience)
  - [15. Guidance for Business Administrators](#15-guidance-for-business-administrators)
  - [16. Guidance for Frontend / AMIS UI Developers](#16-guidance-for-frontend--amis-ui-developers)
  - [17. Guidance for Backend Developers](#17-guidance-for-backend-developers)
  - [18. Guidance for Architects and Platform Engineers](#18-guidance-for-architects-and-platform-engineers)
- [Part V — Reference](#part-v--reference)
  - [19. Operational Runbook](#19-operational-runbook)
  - [20. Design Decisions and Rationale](#20-design-decisions-and-rationale)
  - [21. Full Glossary](#21-full-glossary)
  - [22. Related Documents](#22-related-documents)

---

# Part I — Understanding the Domain

---

## 1. Executive Summary

### 1.1 What Is the System Metadata Service — In Plain Language

Every software system has two kinds of knowledge baked into it: knowledge that never changes (like the rules of double-entry accounting) and knowledge that must be adjustable without rewriting the software (like which currency a business uses, or whether a new module is ready for a specific customer).

The **System Metadata Service** is AWO ERP's authoritative home for all adjustable knowledge. It is the single place in the system where the answers to questions like these live:

- *"What currency should invoices display for this tenant?"*
- *"Is the LPG supply module switched on for this business?"*
- *"Has this tenant completed the onboarding wizard?"*
- *"What payment methods are accepted at this station?"*

Without a service like this, those answers are scattered across dozens of database tables, environment variables, hardcoded defaults, and developer assumptions — making the system rigid, hard to maintain, and impossible to customise per tenant without code changes.

### 1.2 The Six Things It Manages

The System Metadata Service is one service with six logical modules:

| Module | Plain-language purpose |
|---|---|
| **Configuration** | Describes *how* the system should behave for a given tenant or globally |
| **Feature Flags** | Controls *whether* a capability exists and is visible right now |
| **Operational Metadata** | Tracks system-level state that doesn't belong in any business table |
| **Module Registry** | Declares which modules exist, their capabilities, dependencies, and lifecycle status |
| **Entity & Schema Metadata** | Describes the structure of every business entity — fields, types, validation rules, relationships |
| **Custom Fields** | Stores tenant-defined extensions to built-in entity schemas |

These six modules share the same storage, tenancy isolation, audit trail, and caching infrastructure. They are not separate services — they are specialised concerns inside one coherent domain.

### 1.3 Why It Matters

The System Metadata Service sits at the foundation of nearly every user-facing interaction in AWO ERP. Before any screen is rendered, before any API call proceeds, the system consults metadata to answer: *Is this tenant allowed to see this? Is this feature on? What are the rules for this business?*

Getting this service right means the entire platform becomes configurable, safe to evolve, and honest about what is and is not available to each tenant.

### 1.4 Who This Document Is For

This document is written in layers. You do not need to read all of it:

- **Business owners and administrators** — read Sections 1, 2, 4.1 (Glossary), 6.1–6.2, 7.1–7.2, 9.1, 11.1, and Section 15.
- **Frontend / AMIS UI developers** — read all of Part I, Part II, Section 13, and Section 16.
- **Backend developers** — read everything except Section 15.
- **Architects and platform engineers** — read the entire document.

---

## 2. The Problem System Metadata Solves

### 2.1 The Hardcoded System Problem

Imagine AWO ERP has two customers: a fuel station in Manzoni and a restaurant chain in Nairobi. The fuel station deals in Kenyan Shillings, runs a loyalty programme, and has the LPG module active. The restaurant chain uses the same currency but has no LPG, uses different tax categories, and is piloting a new reporting module that is not yet stable enough for the fuel station.

In a naive system, making these two customers behave differently requires either:

1. **Separate deployments** — running two copies of the entire application, diverging over time and doubling maintenance cost, or
2. **Hardcoded conditionals** — burying `if customerName == "restaurant"` branches deep in business logic, making the codebase brittle and untestable, or
3. **Separate configuration files** — storing per-customer settings in deployment environment variables, which requires a developer and a redeployment to change anything.

None of these scale. The System Metadata Service solves this by providing a **runtime-configurable, per-tenant, auditable store** for all the things that differentiate tenants from each other — managed through the application itself, not through code or deployments.

### 2.2 Real-World Scenarios at AWO

Here are concrete examples from the businesses AWO ERP already serves:

**Shell Maanzoni fuel station:**
- Configuration: `default_currency = KES`, `accepted_payment_methods = [cash, mpesa, shell_card]`, `fuel_tax_rate = 0.16`
- Feature flag: `lpg_module_enabled = true`, `car_wash_module_enabled = true`
- Operational metadata: `onboarding_completed = true`, `last_eod_report_generated = 2024-03-15`

**A new tenant being onboarded (hotel):**
- Configuration: `default_currency = KES`, `accepted_payment_methods = [cash, mpesa]`, `restaurant_covers = 40`
- Feature flag: `lpg_module_enabled = false`, `hotel_module_enabled = true`, `new_reporting_dashboard = false` (not yet stable)
- Operational metadata: `onboarding_completed = false`, `wizard_step = 3`

Neither of these scenarios requires a code change or redeployment. They require a value in the System Metadata Service.

### 2.3 The Spectrum from Static to Dynamic

Not all adjustable knowledge is the same. There is a spectrum:

```
Static config          Dynamic config         Feature flags       Runtime metadata
(environment vars)     (tenant settings)      (on/off switches)   (transient state)
     │                       │                      │                    │
  deploy-time             admin UI              engineer/admin         app writes
  never changes          changes rarely          changes sometimes      changes often
```

The System Metadata Service owns everything in the middle two columns — **dynamic configuration** and **feature flags** — and also manages **runtime metadata** in the rightmost column. Static environment-level configuration (database URLs, API secrets) is handled separately by Viper/environment variables and is outside this service's scope.

### 2.4 Why These Concerns Belong Together

Configuration, feature flags, operational metadata, the module registry, entity schemas, and custom fields are kept in one service because:

- They share the same **tenant scoping and RLS isolation** requirements
- They share the same **read-heavy, write-rare** access pattern (with the exception of operational metadata)
- They share the same **audit trail** requirement (changes must be tracked)
- They feed into the same **UI gating chain** (read once per request, not per-module)
- They participate in the same **versioning model** for schema-bearing sub-domains
- Splitting them creates multiple caches to keep in sync, multiple places to debug, and unnecessary inter-service latency

---

## 3. System Metadata in the ERP Landscape

### 3.1 How SAP Handles It

SAP is the world's largest ERP and its approach to system metadata is comprehensive but heavyweight. SAP uses a concept called the **CCMS (Computing Center Management System)** for system-level parameters, alongside **client-level settings** and **transport layers** for configuration management.

In SAP, "clients" are roughly equivalent to AWO's tenants. Client-specific configuration lives in tables like `T000` (client master) and is managed through transaction `SCC4`. Feature-equivalent control is achieved through **business functions** (`SFW5` transaction) which activate entire capability packages, and through **enhancement switches**.

What SAP does well: rigorous separation of client data, transport-controlled change management, and extensive customising tables (the "IMG" — Implementation Guide). What SAP does poorly for AWO's context: it is designed for large enterprise IT departments, requires trained basis administrators to change most settings, and has no concept of lightweight feature flags in the modern sense.

**AWO's takeaway from SAP:** The concept of client-level isolation (our RLS-based tenant isolation) and the idea that "customising" (configuration) is a first-class concern deserving its own structured store.

### 3.2 How Odoo Handles It

Odoo (formerly OpenERP) manages system configuration through two mechanisms: `ir.config_parameter` (a global key-value store in the database) and the **Settings UI** (`res.config.settings`), which maps UI form fields to model fields across multiple modules.

Odoo's approach is pragmatic and developer-friendly. Any module can register settings by adding fields to `res.config.settings`, and those settings are stored either in `ir.config_parameter` (for simple values) or directly on company or user records. Feature activation is handled through **Odoo Apps** — enabling a module installs it system-wide, with per-company activation handled by configuration fields.

What Odoo does well: the settings system is easy to extend, closely tied to the UI, and accessible to non-technical administrators. What it does less well: there is no formal feature flag concept, `ir.config_parameter` is a global store with no tenant isolation by default, and there is no rollout or targeting mechanism.

**AWO's takeaway from Odoo:** The pattern of connecting configuration to a structured UI (our AMIS UI page generation), and the idea that modules should be able to declare their own configuration needs in a self-describing way.

### 3.3 How ERPNext/Frappe Handles It

Frappe, the framework underlying ERPNext, has a **System Settings doctype** for global configuration and **individual site configuration** in `site_config.json` for infrastructure-level settings. Feature control is managed through a combination of module enable/disable toggles, workspace visibility, and role-based menu access.

Frappe's multi-tenancy model ("sites") is a full database-per-tenant approach, so tenant isolation of settings is implicit — each site has its own `System Settings` record. There is no concept of per-tenant feature flags within a shared database; features are controlled at the site level.

What Frappe does well: the DocType system makes all configuration introspectable and self-documenting. The separation between `site_config.json` (developer-managed) and `System Settings` (admin-managed) mirrors the static/dynamic spectrum described in Section 2.3. What it does less well: no formal feature flag lifecycle, no rollout percentage or tenant-targeting mechanism, and the site-per-tenant model makes shared-database multi-tenancy impossible.

**AWO's takeaway from ERPNext:** The introspectable, self-describing configuration model — every key should know its type, allowed values, and description, not just its value.

### 3.4 How Microsoft Dynamics 365 Handles It

Dynamics 365 manages configuration through **Organisation Settings**, **System Settings**, and **Environment Variables** (introduced in the Power Platform layer). Feature control is managed through **Feature Management** workspace, which provides a formal lifecycle for features: discover → enable → disable — with the ability to enable per-environment and, in some cases, per-organisation.

Dynamics' Feature Management is the closest analogue to modern feature flag systems in the enterprise ERP space. Features have descriptions, affected areas, and can be toggled without deployment. Environment Variables provide a typed, solution-aware configuration store.

What Dynamics does well: the Feature Management workspace is well-designed, feature lifecycle is formal, and Environment Variables are typed and can hold connection references (not just scalar values). What it does less well: the system is complex to navigate for non-developers, and the Power Platform layer introduces significant abstraction overhead.

**AWO's takeaway from Dynamics 365:** The formal feature lifecycle concept (draft → enabled → disabled → retired) and the idea that features should be self-describing with affected area documentation.

### 3.5 What AWO ERP Adopts, Adapts, and Does Differently

| Concept | Source inspiration | AWO implementation |
|---|---|---|
| Tenant-level isolation of settings | SAP client model | PostgreSQL RLS per `tenant_id` |
| Typed, self-describing configuration keys | ERPNext DocType / Dynamics Environment Variables | Schema-validated keys with type, description, allowed values |
| Configuration connected to UI | Odoo `res.config.settings` | AMIS UI pages generated by Go functions reading config values |
| Formal feature lifecycle | Dynamics Feature Management | `draft → enabled → disabled → retired` states with audit trail |
| Feature targeting | (absent in traditional ERPs) | Per-tenant flag evaluation with future rollout percentage support |
| Combined config + flags in one service | (split in most ERPs) | Single service, shared infrastructure, one gating chain |
| Admin accessibility | Odoo Settings UI | AMIS-rendered admin pages gated by `system_metadata.*.write` permissions |

### 3.6 Trade-Off Comparison Matrix

| Dimension | SAP | Odoo | ERPNext | Dynamics 365 | AWO ERP |
|---|---|---|---|---|---|
| Tenant isolation | Client-per-schema | Global (per company field) | Site-per-database | Org-level | RLS in shared DB |
| Feature flag lifecycle | None (business functions) | None | None | Formal (3 states) | Formal (4 states) |
| Rollout targeting | None | None | None | Per-environment | Per-tenant (planned %) |
| Admin accessibility | Low (basis admin needed) | High | High | Medium | High (AMIS UI) |
| Configuration type safety | High (customising tables) | Medium (Python fields) | High (DocType fields) | High (typed env vars) | High (schema-validated) |
| Audit trail on changes | Yes (transport log) | Partial | Yes (changelog) | Yes | Yes (audit service) |
| Performance (reads) | DB + buffer pool | DB + ORM cache | DB + cache | DB + platform cache | Redis cache-first |

---

# Part II — Domain Model

---

## 4. Core Concepts and Vocabulary

### 4.1 Glossary

This glossary is written for all audiences. Technical elaborations follow in parentheses for developer readers.

**Tenant** — A business or organisation that uses AWO ERP independently of others. Each tenant's data — including their metadata — is completely isolated from every other tenant's data, even though they share the same application and database. *(Enforced via PostgreSQL Row-Level Security; `tenant_id` column on all metadata rows.)*

**System Metadata** — The parent domain. The umbrella term for all runtime-adjustable knowledge that the system holds about itself and its tenants. The System Metadata Service manages this domain.

**Configuration** — A named setting with a typed value that describes how the system should behave. Example: `invoice_currency = KES`. Configuration values are almost always set by an administrator and change infrequently. *(Key-value store with type validation; tenant-scoped or global.)*

**Feature Flag** — A named switch that controls whether a capability is available at all. Example: `lpg_module_enabled = true`. Feature flags are typically managed by platform engineers and can be turned on or off without any code change or redeployment. *(Boolean result of an evaluation that may consider tenant context, rollout percentage, or explicit targeting.)*

**Operational Metadata** — System-managed state stored outside business tables. Example: `onboarding_wizard_step = 3`. Unlike configuration, operational metadata is often written by the application itself as part of normal operation, not manually set by administrators. *(Ephemeral or semi-persistent key-value entries scoped to tenant or user.)*

**Scope** — The level at which a metadata entry applies. There are three scopes: **global** (applies to all tenants), **tenant** (applies to one specific tenant), and **user** (applies to one specific user within a tenant). Tenant-scoped entries override global entries for that tenant.

**Key** — The identifier for a metadata entry. Keys follow a namespaced dot-notation: `<module>.<category>.<name>`. Example: `finance.tax.default_rate`, `station.fuel.accepted_payment_methods`.

**Value** — The data stored against a key. Values are typed: they can be strings, numbers, booleans, enumerations (a value from a predefined list), or JSON objects.

**Flag Evaluation** — The process of determining whether a feature flag is on or off for a given request. This is not a simple database read; it involves checking the flag's state, any targeting rules, and the requesting tenant's context to produce a final boolean answer.

**Gating** — The act of checking metadata before allowing a user or process to proceed. The UI is "gated" by metadata: before a page renders, the system checks whether the relevant flags are on and the relevant configuration is valid.

**Audit Trail** — A tamper-evident record of every change made to a metadata value: who changed it, when, what it was before, and what it became. Every write to the System Metadata Service is recorded in the Audit Service.

**Cache** — A fast in-memory copy of metadata values (stored in Redis) that allows the system to answer metadata queries in under 1 millisecond without hitting the database on every request. The cache is invalidated (cleared) whenever a value changes.

**Kill Switch** — A feature flag that, when turned off, immediately and completely disables a capability across all instances of the running application. Kill switches are used to safely disable unstable features in production without a redeployment.

**Module Registry** — The sub-domain of System Metadata that catalogues every module available in AWO ERP — its capabilities, dependencies, configuration keys, and lifecycle status. The Module Registry answers "what exists?" while Feature Flags answer "is it on for this tenant?"

**Entity Metadata** — The sub-domain of System Metadata that formally describes the structure of every business entity: its fields, data types, validation rules, relationships, and computed properties. Entity metadata is the single source of truth for the shape of AWO ERP's data model.

**Field Descriptor** — A metadata record describing a single field on an entity: its name, type, nullability, validation rules, and display properties. Field descriptors are the building blocks of entity metadata.

**Custom Fields** — Tenant-defined extensions to built-in entity schemas. A tenant can add new fields to an entity (e.g., "Pump Attendant Shift Badge" on a fuel sale) through the admin UI, without any code changes. Custom field definitions are stored in the Custom Fields sub-domain.

**Schema Version** — An immutable snapshot of a schema-bearing sub-domain (Entity Metadata, Custom Fields, or Module Registry) at a point in time. Schema versions allow the system to answer "what did this schema look like at version N?" and to produce structured diffs between versions.

**Module Dependency** — A declared requirement that one module needs another to be active before it can function. Enforced by the System Metadata Service at flag-enable time to prevent partial module activation.

### 4.2 The Metadata Lifecycle

All metadata entries — whether configuration keys or feature flags — move through a defined lifecycle:

```
                    ┌─────────┐
                    │  DRAFT  │  ← Key exists in the system but is not yet in effect.
                    └────┬────┘    Used during development or staged rollout planning.
                         │
                         ▼
                    ┌─────────┐
                    │ ACTIVE  │  ← The key is live and being read by the system.
                    └────┬────┘    This is the normal operational state.
                         │
              ┌──────────┴──────────┐
              ▼                     ▼
        ┌──────────┐          ┌──────────────┐
        │ DISABLED │          │  DEPRECATED  │
        └────┬─────┘          └──────┬───────┘
             │                       │
             │  (can re-enable)      │  (scheduled for removal)
             └──────────┬────────────┘
                        ▼
                  ┌──────────┐
                  │ ARCHIVED │  ← Permanently inactive. Retained for audit history
                  └──────────┘    but never read by live system queries.
```

---

## 5. The Sub-Domains and How They Relate

### 5.1 Visual Relationship Map

```
┌──────────────────────────────────────────────────────────────────────────┐
│                       SYSTEM METADATA SERVICE                             │
│                                                                            │
│  ┌───────────────────────┐   ┌───────────────────────────────────────┐   │
│  │    CONFIGURATION      │   │           FEATURE FLAGS               │   │
│  │  "How should this     │   │  "Should this capability exist?"      │   │
│  │   work?"              │   │  - Boolean evaluation                 │   │
│  │  - Typed key-values   │   │  - Formal lifecycle                   │   │
│  │  - Admin-managed      │   │  - Engineer-managed                   │   │
│  └───────────────────────┘   └───────────────────────────────────────┘   │
│                                                                            │
│  ┌───────────────────────┐   ┌───────────────────────────────────────┐   │
│  │  OPERATIONAL METADATA │   │         MODULE REGISTRY               │   │
│  │  "What state is the   │   │  "What modules exist and what can     │   │
│  │   system in now?"     │   │   they do?"                           │   │
│  │  - App-written        │   │  - Capabilities declaration           │   │
│  │  - Transient          │   │  - Dependencies & lifecycle           │   │
│  └───────────────────────┘   └───────────────────────────────────────┘   │
│                                                                            │
│  ┌───────────────────────┐   ┌───────────────────────────────────────┐   │
│  │  ENTITY & SCHEMA      │   │         CUSTOM FIELDS                 │   │
│  │  METADATA             │   │  "What has this tenant added to       │   │
│  │  "What is the shape   │   │   built-in entities?"                 │   │
│  │   of each entity?"    │   │  - Tenant-defined fields              │   │
│  │  - Field descriptors  │   │  - Type-validated extensions          │   │
│  │  - Validation rules   │   │  - Per-entity quotas                  │   │
│  └───────────────────────┘   └───────────────────────────────────────┘   │
│                                                                            │
│     Shared:  RLS tenancy │ Redis cache │ Audit trail │ Versioning         │
└──────────────────────────────────────────────────────────────────────────┘
```

### 5.2 What They Share

The six modules are kept in one service because they share the same critical infrastructure:

- **Tenant isolation** — every row carries a `tenant_id` and is protected by PostgreSQL Row-Level Security. A query made in the context of Tenant A can never return rows belonging to Tenant B.
- **Redis cache** — metadata is read on nearly every request. Without caching, every page load would require multiple database roundtrips. The cache is shared, keyed by `tenant_id + key`, and invalidated on every write.
- **Audit trail** — every mutation (create, update, disable) generates an audit event via the Audit Service. This is a non-negotiable requirement for enterprise software.
- **Access control** — writes to metadata are governed by the same permission system as every other part of AWO ERP (`system_metadata.<resource>.<action>` permission strings).
- **Versioning** — schema-bearing sub-domains (Entity Metadata, Custom Fields, Module Registry) participate in the versioning model described in Section 12.

### 5.3 How They Differ — The Critical Distinction

The single most important distinction to understand:

> **A feature flag can gate access to a configuration key.** If the `lpg_module_enabled` flag is off, the LPG configuration keys are never read — they are irrelevant.
>
> **A configuration value must never gate a feature flag.** That would be circular and would make the flag system unpredictable.

The direction of dependency is always: **Feature Flag → Configuration**. Not the other way around.

### 5.4 When to Use Each

| Situation | Use |
|---|---|
| Setting the default tax rate for a tenant | Configuration |
| Deciding whether the tax module is available at all | Feature Flag |
| Recording which step of onboarding a tenant is on | Operational Metadata |
| Defining which payment methods are accepted | Configuration |
| Switching off an unstable payment integration in production | Feature Flag (kill switch) |
| Storing the timestamp of the last successful sync | Operational Metadata |
| Setting the display timezone for a tenant's reports | Configuration |
| Rolling out a new reports UI to 10% of tenants | Feature Flag (with rollout %) |
| Declaring that the LPG module requires the Finance module | Module Registry |
| Describing the fields on an Invoice entity | Entity & Schema Metadata |
| Adding a "Pump Attendant Shift ID" field to fuel sales | Custom Fields |

---

## 6. Configuration Module — Deep Dive

### 6.1 What Configuration Is and Is Not

Configuration describes **how** a feature or module should behave. It never controls whether that feature exists — that is Feature Flags' job.

Configuration is appropriate when:
- The value meaningfully changes the *outcome* of a business process (e.g., tax rate, currency)
- Different tenants legitimately need different values
- The value is relatively stable and changes through a deliberate administrative action
- The value needs to be readable by multiple modules simultaneously

Configuration is **not** appropriate when:
- You want to turn a capability entirely on or off (use a Feature Flag)
- The value changes very frequently during normal operation (consider Operational Metadata)
- The value belongs to a specific business entity like a product or invoice (put it in the business table)

### 6.2 Configuration Categories

Configuration keys are namespaced to prevent collisions and communicate ownership:

**Infrastructure configuration** — keys under `system.*`
These are platform-level settings that affect how the application runs for a tenant. Examples: `system.timezone`, `system.locale`, `system.date_format`.

**Business rules configuration** — keys under `<module>.*`
These drive business logic within a specific module. Examples: `finance.tax.default_rate`, `finance.accounts.base_currency`, `station.fuel.low_stock_threshold_litres`.

**UI behaviour configuration** — keys under `ui.*`
These affect how the frontend presents information. Examples: `ui.dashboard.default_view`, `ui.reports.rows_per_page`.

**Integration configuration** — keys under `integrations.*`
These configure connections to external services. Examples: `integrations.mpesa.shortcode`, `integrations.pesapal.merchant_id`.

### 6.3 Tenant-Scoped vs Global Configuration

Every configuration key has a **default scope** defined when it is registered:

- **Global keys** have a single value that applies to all tenants unless a tenant explicitly overrides it. Example: `system.supported_currencies` — the list of currencies the platform supports is global.
- **Tenant-scoped keys** have per-tenant values. Example: `finance.accounts.base_currency` — each tenant has its own currency setting.

**Override hierarchy:**

```
Global default value
        │
        ▼  (overridden by)
Tenant-specific value
        │
        ▼  (for certain keys, overridden by)
User-specific value
```

When the system reads a configuration key, it always resolves the most specific value available for the current context. A tenant-specific value always wins over the global default. A user-specific value (where permitted) wins over the tenant value.

### 6.4 Configuration Type System

Every configuration key is registered with a type. The system validates values against this type on write — it is not possible to store a string where a number is expected. Supported types:

| Type | Description | Example value |
|---|---|---|
| `string` | Free-form text | `"Africa/Nairobi"` |
| `integer` | Whole number | `40` |
| `decimal` | Decimal number | `0.16` |
| `boolean` | True or false | `true` |
| `enum` | One value from a predefined list | `"KES"` from `["KES","USD","EUR"]` |
| `json` | Structured object or array | `["cash","mpesa","shell_card"]` |
| `duration` | Time period | `"24h"`, `"7d"` |

The type system prevents a class of silent bugs where a module reads a configuration value and receives an unexpected type.

### 6.5 Configuration Inheritance and Defaults

When a tenant is first provisioned, the system seeds their configuration with defaults. These defaults are defined per configuration key and represent sensible starting values for a Kenyan business context (KES currency, Africa/Nairobi timezone, etc.).

Administrators can then adjust values through the admin UI. The original default is never lost — the system stores the current value and the original default separately, making it always possible to "reset to default."

### 6.6 Who Can Change What

Configuration writes are governed by the permission system using the format `system_metadata.configuration.<action>`:

| Permission | What it allows |
|---|---|
| `system_metadata.configuration.read` | Read any configuration key for this tenant |
| `system_metadata.configuration.write` | Update tenant-scoped configuration values |
| `system_metadata.configuration.admin` | Create new keys, change global defaults, manage key schema |

Tenant administrators hold `read` and `write`. Platform engineers and system administrators hold `admin`. Regular users hold `read` for the keys relevant to their module access.

### 6.7 Audit Trail for Configuration Changes

Every change to a configuration value generates an audit event recording:
- The key that changed
- The previous value
- The new value
- Who made the change (actor ID)
- When the change was made
- The tenant context

This means it is always possible to answer the question: *"Why is this system behaving differently than it did last month?"* — by reviewing the configuration change history.

### 6.8 Practical Examples — AWO Operations

Below are real configuration scenarios from AWO-operated businesses:

**Shell Maanzoni fuel station:**
```
station.fuel.accepted_payment_methods = ["cash", "mpesa", "shell_card", "credit_account"]
station.fuel.low_stock_threshold_litres = 2000
finance.tax.fuel_vat_rate = 0.16
integrations.mpesa.paybill_number = "XXXXX"
system.timezone = "Africa/Nairobi"
system.date_format = "DD/MM/YYYY"
```

**Hotel/restaurant business:**
```
hotel.restaurant.covers = 40
hotel.restaurant.service_charge_rate = 0.10
finance.accounts.base_currency = "KES"
ui.dashboard.default_view = "hotel_occupancy"
```

---

## 7. Feature Flag Module — Deep Dive

### 7.1 What a Feature Flag Is — Plain Language

A feature flag is an on/off switch for a capability inside the software. When the flag is on, users can see and use the feature. When it is off, the feature is completely invisible — it does not appear in menus, its API endpoints return permission errors, and its data is not loaded.

Feature flags allow the engineering team to:
- **Deploy code without activating it** — new features can be shipped to production in a dormant state and activated when ready
- **Activate features per tenant** — some tenants can access a capability before it is rolled out to everyone
- **Disable features instantly** — if a feature causes problems in production, it can be switched off in seconds without a code deployment
- **Test in production safely** — a new feature can be exposed to internal users or a small percentage of tenants before full rollout

### 7.2 Why Feature Flags Live Inside System Metadata (Not a Separate Service)

In large organisations, feature flag management is sometimes a separate service (LaunchDarkly, Unleash, etc.). AWO ERP keeps feature flags inside the System Metadata Service for the following reasons:

- They require the same **tenant isolation** as configuration — a flag enabled for Tenant A must not affect Tenant B
- They are read as part of the same **gating chain** as configuration, before any UI is rendered
- They share the same **audit trail** requirements
- Separating them would require the UI rendering pipeline to make two service calls instead of one
- The flag evaluation complexity AWO needs (per-tenant targeting, kill switches) does not require a dedicated service's overhead

If AWO ERP ever needs probabilistic A/B testing across millions of users, that decision can be revisited. At the current scale and use case, keeping flags inside System Metadata is the right trade-off.

### 7.3 Flag Types

AWO ERP recognises three types of feature flags:

**Release flags** — used during development to separate deployment from activation.
These flags let engineers merge code to main and deploy without activating a feature. When the feature is tested and approved, the flag is enabled. Example: `new_reporting_dashboard_enabled`.

**Operational flags (kill switches)** — used to protect the system in production.
These flags wrap capabilities that could become unstable. If the integrated payment gateway starts behaving erratically, the flag `pesapal_integration_enabled` can be turned off immediately, falling back to manual payment recording. Example: `mpesa_stk_push_enabled`.

**Tenant entitlement flags** — used to control which tenants have access to which modules.
These flags represent commercial or readiness decisions about which tenants are entitled to which capabilities. Example: `lpg_module_enabled`, `hotel_module_enabled`. They are off by default for all tenants and explicitly enabled per tenant.

### 7.4 Flag Lifecycle

```
   DRAFT ──────► ENABLED ──────► DISABLED ──────► RETIRED
     │               │               │
     │               │   (can toggle │
     │               │    freely)    │
     │               └──────────────┘
     │
     └── (can be deleted before going live,
          but once ENABLED, it enters the
          audit-tracked lifecycle)
```

**DRAFT:** The flag has been registered in the system but is not yet evaluated. Code can reference it, but it evaluates as `false` everywhere. This state is used during development.

**ENABLED:** The flag is live and evaluating based on its targeting rules. For a simple tenant entitlement flag, this means "on for this tenant." For a rollout flag, this means "on for X% of eligible tenants."

**DISABLED:** The flag is turned off. It evaluates as `false` everywhere regardless of targeting rules. This is the state used for kill-switch activation.

**RETIRED:** The flag is permanently decommissioned. The code that checked it should be removed. The flag record is retained for audit history but never evaluated.

### 7.5 Flag Evaluation Logic

When the system asks "Is this flag on for this request?", the evaluator runs the following logic in order:

```
1. Is the flag in ENABLED state?           No → return false
         │
         ▼ Yes
2. Does the flag have tenant targeting?
   - Is the requesting tenant in the       No → continue
     explicit "enabled tenants" list?     Yes → return true
         │
         ▼
3. Does the flag have a rollout %?         No → return true (enabled for all)
   - Hash tenant ID → is it within         No → return false
     the rollout percentage?              Yes → return true
```

The result is always a boolean. The calling code never needs to understand targeting rules — it only receives `true` or `false`.

### 7.6 The Relationship Between Feature Flags and Configuration

These two modules are closely related but serve different purposes. The most common source of confusion is: "Should this be a flag or a config value?"

The rule is: **flags control existence; configuration controls behaviour.**

A practical test:

> *"If I turned this off, would the feature disappear entirely?"*
> → Yes → Feature Flag
>
> *"If I changed this, would the feature still exist but work differently?"*
> → Yes → Configuration

A real example: the LPG supply module has both:
- A feature flag `lpg_module_enabled` — when off, no LPG screens, no LPG API endpoints, no LPG data loaded
- Configuration keys like `lpg.pricing.base_margin_percent` and `lpg.delivery.minimum_order_kg` — these only matter when the flag is on

### 7.7 Kill-Switch Behaviour

A kill switch is a feature flag used defensively. The name comes from the idea of an emergency stop button.

When a kill switch is activated (flag moved to DISABLED):
- The AMIS UI stops rendering the affected pages/components — they disappear from menus
- API endpoints covered by the flag return a `503 Service Unavailable` or a structured "feature disabled" response
- Any in-flight Temporal workflows that check the flag mid-execution respect the new state on their next activity step
- The change takes effect within one cache TTL cycle (seconds, not minutes) across all running instances

Kill switches require no code deployment, no database migration, and no service restart.

### 7.8 Practical Examples

| Flag key | Type | Default | Use case |
|---|---|---|---|
| `lpg_module_enabled` | Tenant entitlement | `false` | Enable LPG supply management for qualifying tenants |
| `hotel_module_enabled` | Tenant entitlement | `false` | Enable hotel/hospitality module |
| `new_eod_report_enabled` | Release | `false` | Ship new end-of-day report UI without activating it |
| `mpesa_stk_push_enabled` | Kill switch | `true` | Emergency disable of M-Pesa STK push integration |
| `car_wash_module_enabled` | Tenant entitlement | `false` | Enable car wash service tracking |
| `advanced_analytics_enabled` | Release + entitlement | `false` | Gradual rollout of analytics dashboard |

---

## 8. Operational Metadata — Deep Dive

### 8.1 What Belongs Here

Operational metadata covers state that the application needs to remember about itself or a tenant's journey through the system, but that does not belong in any business domain table.

The test for operational metadata: *"Is this a fact about the business, or is it a fact about the system's relationship with the business?"*

- `invoice.status = PAID` — fact about the business → belongs in the invoice table
- `onboarding.wizard_step = 3` — fact about the system's current interaction with the tenant → operational metadata
- `last_eod_report_run_at = 2024-03-15T23:00:00Z` — fact about system process execution → operational metadata
- `data_migration_v2_completed = true` — fact about a system-side migration → operational metadata

### 8.2 Who Writes Operational Metadata

Unlike configuration (written by administrators) and feature flags (written by engineers), operational metadata is typically **written by the application itself** as part of normal operation. Examples:

- The onboarding wizard writes `onboarding.step` after each completed step
- The end-of-day report process writes `reports.last_eod_run_at` after successful completion
- The data migration worker writes `migration.v2.completed = true` after finishing

Operational metadata can also be written by administrators in exceptional cases (e.g., manually resetting an onboarding wizard for a tenant).

### 8.3 Scope and Ownership

Operational metadata follows the same scope hierarchy as configuration (global, tenant, user). In practice:

- Most operational metadata is **tenant-scoped** (onboarding state, last process runs)
- Some is **user-scoped** (wizard step for a specific user's setup flow, UI dismissal flags like "user has seen this announcement")

---

## 9. Module Registry — Deep Dive

### 9.1 What the Module Registry Is

The Module Registry is the System Metadata Service's authoritative catalogue of every module that exists within AWO ERP. Where Feature Flags answer *"is this module on for this tenant?"*, the Module Registry answers *"what modules exist at all, what do they do, what do they need, and what is their current state?"*

Think of it as AWO ERP's internal app store inventory — a structured declaration of every capability the platform can provide, independent of whether any given tenant has it enabled.

Without a registry, the platform has no formal record of its own capabilities. Modules are discovered implicitly through code rather than declared explicitly through metadata. This makes it impossible to answer questions like: *"Can Tenant A enable the Hotel module? Does it depend on anything they haven't activated yet? What features does it expose?"* — without reading source code.

### 9.2 What the Registry Stores Per Module

Every module registered in AWO ERP has a formal entry containing:

**Identity**
- `module_id` — unique identifier, kebab-case: `lpg-supply`, `hotel`, `station-fuel`
- `display_name` — human-readable name shown in the admin UI
- `description` — plain-language summary of what the module does
- `version` — the currently deployed version of this module's code

**Capability Declaration**
A list of named capabilities the module exposes. Each capability maps to a set of feature flags, configuration keys, and entity schemas it introduces. Example for the `station-fuel` module:

```
capabilities:
  - fuel_sales_tracking
  - tank_dip_management
  - pump_attendant_management
  - eod_report_generation
  - mpesa_payment_collection
```

This declaration is what allows the AMIS navigation generator to know which menu items belong to which module, and what permission strings protect them.

**Dependencies**
A list of other modules this module requires to function. The system validates these at flag-enable time — you cannot enable a module whose dependencies are not already active for that tenant.

```
dependencies:
  - finance          # required: all fuel sales produce finance transactions
  - iam              # required: always (implicit for all modules)
```

**Configuration Schema Reference**
A list of the configuration keys this module introduces. This links the module to its entries in the Configuration sub-domain and is used during tenant provisioning to seed the correct default values.

**Status**
The module's current operational status — one of: `stable`, `beta`, `deprecated`, or `disabled`. This is distinct from a tenant's feature flag state. A module marked `deprecated` still works but signals that tenants should migrate away from it.

### 9.3 Module Status Lifecycle

```
  DEVELOPMENT ──► BETA ──► STABLE ──► DEPRECATED ──► DISABLED
       │                       │
       │  (internal use only)  │  (all tenants eligible)
       │                       │
       └── never visible        └── visible in module catalogue
           in tenant admin           with stability badge
```

**DEVELOPMENT** — the module exists in code but is not registered in the live registry. No tenant can see or enable it.

**BETA** — the module is registered and can be enabled for selected tenants (using feature flags with explicit tenant targeting). Shown in the tenant admin with a "Beta" badge.

**STABLE** — the module is fully supported and eligible for all tenants. Feature flag defaults may still be `false` (off until explicitly requested), but the module is production-ready.

**DEPRECATED** — the module still functions but is being phased out. Existing tenants who have it enabled continue to use it. New tenants cannot enable it. The registry entry carries a `deprecated_reason` and `sunset_date`.

**DISABLED** — the module is no longer operational. Its feature flags are force-disabled for all tenants. Retained in the registry for audit history.

### 9.4 Relationship to Feature Flags

The Module Registry and Feature Flags work together but are not the same thing:

| | Module Registry | Feature Flag |
|---|---|---|
| **Question answered** | Does this module exist? What does it do? | Is this module on for this tenant? |
| **Scope** | System-wide (global) | Per-tenant |
| **Changes with** | New releases, deprecations | Tenant onboarding, commercial decisions |
| **Managed by** | Platform engineers (code deployments) | Platform engineers (admin operations) |

A module must be in `STABLE` or `BETA` status in the registry **and** have its corresponding feature flag enabled for a specific tenant before that tenant can access it. Both conditions must be true.

### 9.5 Module Dependency Validation

When a platform engineer attempts to enable a module for a tenant (by enabling its feature flag), the system performs a dependency check:

```
Enable request: hotel_module_enabled = true for Tenant X
       │
       ▼
Registry lookup: hotel module dependencies = [finance, iam]
       │
       ▼
Check: is finance_module_enabled = true for Tenant X?
       │
       ├── Yes → proceed
       └── No  → reject with: "Cannot enable 'hotel' module.
                               Required module 'finance' is not active for this tenant.
                               Enable 'finance' first."
```

This prevents partial module enablement that would leave a tenant in a broken state.

### 9.6 How Other Modules Use the Registry

Other modules query the registry to:
- **AMIS navigation builders** — read capability declarations to know which menu sections to include
- **Tenant provisioning service** — reads all `STABLE` modules' configuration schemas to seed correct defaults
- **Admin UI** — displays the module catalogue with status badges and dependency graphs
- **Health checks** — compares deployed code versions against registry versions to detect stale deployments

---

## 10. Entity & Schema Metadata — Deep Dive

### 10.1 What Entity Metadata Is

Entity metadata describes the *shape* of AWO ERP's business objects — their fields, types, validation rules, relationships, and computed properties. It is the system's formal, introspectable knowledge of its own data model.

Without entity metadata, the system's data model only exists in Go structs and database schemas. Adding a field means a code change, a migration, a redeployment. The UI must be updated separately. Documentation falls out of date. There is no single source of truth.

With entity metadata, the data model is declared once in the System Metadata Service and used everywhere: for dynamic form generation in AMIS, for validation, for generating Go struct types, for API documentation, and for guiding tenants when they extend entities with custom fields (Section 11).

### 10.2 What Counts as an Entity

An entity is any named business object that has a persistent representation in the database and is surfaced to users through the UI. Examples in AWO ERP:

| Entity ID | Business concept |
|---|---|
| `fuel_sale` | A single fuel dispensing transaction at the station |
| `tank_dip` | A physical measurement of fuel volume in a tank |
| `lpg_delivery` | An LPG supply delivery to a customer |
| `invoice` | A customer-facing billing document |
| `employee` | A person employed by a tenant |
| `tenant` | An organisation using AWO ERP |
| `user` | A person with login access |

The entity registry does not include every database table — only tables that represent first-class business objects meaningful to users and tenant administrators.

### 10.3 Field Descriptors

Each field on an entity has a descriptor stored in entity metadata containing:

| Property | Description | Example |
|---|---|---|
| `field_id` | Unique identifier within the entity | `total_amount` |
| `display_name` | Human-readable label | `"Total Amount"` |
| `description` | Plain-language explanation | `"The full billed amount including tax"` |
| `data_type` | One of the standard types (see 10.4) | `decimal` |
| `nullable` | Whether null/empty is permitted | `false` |
| `read_only` | Whether users can edit this field | `false` |
| `system_managed` | Whether only the system writes this | `true` (for computed fields) |
| `sensitive` | Whether to mask in audit logs | `false` |
| `searchable` | Whether this field is indexed for search | `true` |
| `sortable` | Whether list views can sort by this field | `true` |
| `default_value` | Value to use when not supplied | `0.00` |

### 10.4 Field Types Reference

Entity metadata uses a richer type system than Configuration, because entity fields map directly to database columns and UI form controls:

| Type | Database column | UI control | Notes |
|---|---|---|---|
| `string` | `varchar` / `text` | Text input | Has `max_length` constraint |
| `integer` | `bigint` | Number input | |
| `decimal` | `numeric(p,s)` | Currency/number input | Has `precision` and `scale` |
| `boolean` | `boolean` | Toggle / checkbox | |
| `date` | `date` | Date picker | |
| `datetime` | `timestamptz` | Datetime picker | Always stored in UTC |
| `enum` | `varchar` + check | Select / radio | Has `allowed_values` list |
| `uuid` | `uuid` | Hidden / reference | For FK relations |
| `json` | `jsonb` | JSON editor / structured form | |
| `text_long` | `text` | Textarea | No length limit |
| `money` | `numeric(19,4)` | Currency input | Paired with a currency field |
| `phone` | `varchar(20)` | Phone input | E.164 validation |

### 10.5 Validation Rules

Field descriptors carry validation rules that are enforced at both the service layer (on write) and the UI layer (form validation in AMIS). Rules are declarative — the UI and API share the same rules without duplicating logic.

Supported validation rule types:

**Range constraints** — for numeric and date fields: `min`, `max`.
**Length constraints** — for string fields: `min_length`, `max_length`.
**Pattern constraints** — for string fields: a regular expression the value must match. Example: a KRA PIN field with pattern `[A-Z]\d{9}[A-Z]`.
**Required-if rules** — conditional requirement: *"This field is required when field X equals value Y."* Example: `credit_account_number` is required when `payment_method == "credit_account"`.
**Custom validators** — references to named validation functions registered by the owning module. These handle domain-specific rules too complex for declarative expressions.

### 10.6 Relationships and Foreign Keys

Entity metadata also declares relationships between entities. These drive join behaviour in queries, referential integrity rules, and the rendering of related-entity pickers in AMIS forms.

Relationship types:

| Type | Meaning | Example |
|---|---|---|
| `belongs_to` | This entity has a FK to another | A `fuel_sale` belongs to an `employee` (the attendant) |
| `has_many` | Another entity has a FK to this one | A `tank` has many `tank_dip` records |
| `has_one` | Another entity has a unique FK to this one | A `tenant` has one `tenant_config` |
| `many_to_many` | Join table relationship | An `invoice` has many `products` |

Each relationship declaration includes: the related entity ID, the FK field names on each side, the cascade delete behaviour, and whether the relationship is tenant-scoped (it always is, in AWO ERP).

### 10.7 How Entity Metadata Drives AMIS Forms

When a Go page handler generates an AMIS form for creating or editing a `fuel_sale`, it does not hardcode the form fields. Instead:

1. It reads the `fuel_sale` entity descriptor from entity metadata
2. It iterates the field descriptors and produces AMIS form component definitions for each field, respecting type, validation rules, and read-only flags
3. It applies any tenant custom fields (Section 11) on top of the standard fields
4. It returns the complete AMIS JSON form schema

This means: adding a new field to the `fuel_sale` entity, or adding a validation rule to an existing field, changes the form automatically — without touching the AMIS JSON directly.

### 10.8 Localization in Entity Metadata

`display_name` and `description` on field descriptors are stored as translation keys, not literal strings. The full translation string lives in a localization table keyed by `(locale, translation_key)`. For AWO's current operating context (Kenya, English and Swahili), the locale registry holds:

- `en-KE` — English (Kenya) — primary locale, used as fallback
- `sw-KE` — Swahili (Kenya) — secondary locale, available for field labels in tenant-facing UI

When the AMIS page generator reads a field descriptor, it resolves the display name for the requesting user's locale setting. If a translation is missing for the user's locale, it falls back to `en-KE`.

---

## 11. Custom Fields Metadata — Deep Dive

### 15.1 What Custom Fields Are — Plain Language

Every tenant's business has unique data they need to track that the standard AWO ERP data model does not cover out of the box. A fuel station may need to track a *"Pump Attendant Shift Badge Number"* on every sale. A hotel may need a *"Room Number"* on every customer check-in. A restaurant may need a *"Table Number"* on orders.

Custom Fields allow tenant administrators to add new fields to built-in entities without any code changes. The field definition is stored in the System Metadata Service (under the Custom Fields sub-domain), and AWO ERP automatically incorporates it into forms, list views, reports, and data exports for that tenant.

This is a standard capability in mature ERPs. Salesforce calls them "Custom Objects and Fields." ERPNext calls them "Custom Fields via Customize Form." Dynamics 365 calls them "Custom Columns." AWO ERP implements the same concept with tenant isolation.

### 15.2 What a Custom Field Definition Contains

A tenant's custom field definition stores:

| Property | Description | Example |
|---|---|---|
| `field_id` | Auto-generated unique ID | `cf_pump_attendant_shift_badge` |
| `entity_id` | Which entity this extends | `fuel_sale` |
| `tenant_id` | Which tenant owns this field | (the owning tenant's UUID) |
| `display_name` | Label shown in UI | `"Shift Badge Number"` |
| `data_type` | Type from the standard type list | `string` |
| `nullable` | Whether the field is optional | `true` |
| `default_value` | Value when not supplied | `null` |
| `validation_rules` | Constraints on the value | `{max_length: 20}` |
| `display_order` | Position in forms and list views | `5` |
| `is_searchable` | Whether included in search indices | `false` |
| `status` | Active, disabled, or archived | `active` |

### 15.3 Custom Field Scoping

Custom fields are always **tenant-scoped**. A field created by Tenant A is completely invisible to Tenant B — it does not appear in their forms, their API responses, or their exports. The RLS policy on the custom fields table ensures this at the database level.

Custom fields extend the entity schema *for that tenant only*. From Tenant B's perspective, the `fuel_sale` entity has exactly the fields described in the base entity metadata (Section 10) — nothing more.

### 15.4 Custom Field Types Supported

Custom fields support a subset of the full entity field type list — specifically the types that are safe to add to an existing table without requiring schema migration in the base schema:

| Supported | Not supported (add to base schema instead) |
|---|---|
| `string`, `text_long` | `uuid` (FK relationships) |
| `integer`, `decimal`, `money` | `json` (performance/indexing concerns) |
| `boolean` | `many_to_many` (requires join table) |
| `date`, `datetime` | |
| `enum` (with `allowed_values`) | |
| `phone` | |

The restriction on `uuid`/FK types is intentional: custom fields extend the *descriptive* data about an entity, not its *relational* structure. Relationships between entities are always declared in base entity metadata, not through custom fields.

### 15.5 Custom Field Validation

Custom fields participate in the same validation system as base entity fields (Section 10.5). Validation rules on custom fields are declared at field-definition time and enforced on every write to the owning entity, whether through the UI or the API.

A tenant administrator creating a custom field for "Pump Attendant Badge Number" can specify:
- `max_length: 10`
- `pattern: ^\d+$` (digits only)
- `required_if: payment_method == "cash"` (required only on cash sales)

If validation fails on write, the error message references the custom field's `display_name`, not its internal `field_id`, so the user receives a meaningful error: *"Shift Badge Number: must be digits only."*

### 15.6 Limits and Quotas

Custom fields are stored in a flexible JSON column appended to the base entity's storage. To prevent abuse and maintain query performance, limits apply:

| Limit | Default value | Rationale |
|---|---|---|
| Maximum custom fields per entity | 20 | Prevents unbounded schema expansion |
| Maximum `allowed_values` for enum fields | 50 | Keeps enum validation fast |
| Maximum `max_length` for string fields | 1000 | Prevents abuse of text storage |
| Maximum display name length | 80 characters | UI display constraint |

These limits are stored as configuration values (`system_metadata.custom_fields.max_per_entity`, etc.) and can be adjusted per tenant by platform engineers with `admin` permission.

### 15.7 Custom Fields and the AMIS UI

When a Go handler generates an AMIS form for an entity, the process is:

1. Load base entity field descriptors from Entity Metadata (Section 10)
2. Load the tenant's custom field definitions for the same entity
3. Merge the two lists, ordering by `display_order`
4. Generate AMIS form components for both base and custom fields
5. Return the unified form schema

From the user's perspective, custom fields appear inline alongside standard fields. There is no visual distinction between a built-in field and a tenant-defined one — both render identically in the AMIS form.

### 15.8 Disabling vs Deleting a Custom Field

**Disabling** a custom field (`status = disabled`) removes it from all forms and list views immediately. Existing data stored in that field is preserved. The field can be re-enabled, restoring visibility of the data.

**Deleting** a custom field is a destructive operation. It removes the field definition, removes the data stored in that field from all existing records, and cannot be undone. The system requires explicit confirmation and generates an audit event with the total number of records affected.

The recommendation is to disable rather than delete, unless data removal is intentionally required (e.g., GDPR data minimisation request).

---

## 12. Metadata Versioning

### 16.1 Why Versioning Matters

The System Metadata Service's schema-bearing sub-domains — Entity Metadata, Custom Fields, and the Module Registry — change over time. A new field is added to an entity. A custom field's validation rule is tightened. A module declares a new capability. Each of these changes is a *schema evolution event*.

Without versioning, it is impossible to answer questions like: *"What did the fuel_sale entity look like six months ago when this batch of records was imported?"* or *"Which schema version was active when this validation error occurred?"* or *"What changed between the version running in staging and the version running in production?"*

Versioning provides a complete, auditable history of schema evolution.

### 16.2 What Gets Versioned

Not all metadata participates in versioning. The distinction:

| Sub-domain | Versioned? | Reason |
|---|---|---|
| Configuration | No | Key-value changes tracked via audit trail instead |
| Feature Flags | No | State changes tracked via audit trail instead |
| Operational Metadata | No | Transient state, not a schema concern |
| Entity & Schema Metadata | **Yes** | Structural schema changes need point-in-time snapshots |
| Custom Fields | **Yes** | Tenant schema changes need point-in-time snapshots |
| Module Registry | **Yes** | Module capability declarations need version tracking |

### 16.3 The Versioning Model

AWO ERP uses an **immutable snapshot** model for schema versioning. Each time a schema-bearing sub-domain is modified, the system:

1. Captures the complete current state of the affected schema as a JSON snapshot
2. Assigns it a monotonically increasing version number and a timestamp
3. Stores the snapshot in an append-only versioning table
4. Updates the "active version" pointer to the new snapshot

The previous snapshot is never modified or deleted. It remains permanently readable. This allows any part of the system to request: *"Give me the entity schema as it existed at version N"* — and receive an exact, reproducible answer.

### 16.4 Version Identifiers

Version identifiers use the format: `{sub-domain}.{entity-or-module}.v{N}`

Examples:
- `schema.fuel_sale.v12` — version 12 of the `fuel_sale` entity schema
- `schema.fuel_sale.custom.tenant-abc123.v3` — version 3 of Tenant abc123's custom fields for `fuel_sale`
- `modules.lpg-supply.v4` — version 4 of the `lpg-supply` module's capability declaration

### 16.5 Active vs Historical Versions

At any given time, only one version of each schema is **active** — meaning it is used for validation, form generation, and query processing. All other versions are **historical** and are available for audit, diff, and rollback purposes.

The active version is the latest one unless a rollback has been performed. Rollbacks move the active version pointer backward to a previous snapshot — they do not delete the intervening versions.

### 16.6 Schema Diffs

Given two version identifiers, the System Metadata Service can produce a structured diff describing exactly what changed between them:

```
schema.fuel_sale.v11 → schema.fuel_sale.v12

ADDED:
  + field: pump_nozzle_id  (type: string, nullable: true)

MODIFIED:
  ~ field: attendant_id
      nullable: true → false   (field is now required)

REMOVED:
  (none)
```

This diff is used by: the admin UI (to show change history), the migration tooling (to generate safe database migrations), and the audit system (to record what changed and who approved it).

### 16.7 Versioning and Migrations

Schema versioning in the System Metadata Service is distinct from database migrations (`db/migration/*.sql`). They are complementary:

- **Database migrations** change the physical PostgreSQL schema — adding columns, modifying constraints
- **Metadata versioning** changes the *declared* schema — what the application believes the shape of the data to be

In practice, most entity metadata changes *are* accompanied by a database migration. The metadata version snapshot is created as part of the same change process, ensuring the declared schema always matches the actual schema.

Custom field changes — being stored in a flexible JSONB column — do **not** require database migrations. Their metadata version snapshots are standalone.

---

# Part III — How the System Uses Metadata

---

## 13. The UI Rendering Pipeline and Metadata Gating

### 13.1 AMIS UI in AWO ERP — Overview

AWO ERP uses **AMIS** (a JSON-driven UI framework) for its frontend. Rather than a traditional JavaScript frontend application that fetches data and renders HTML, AWO's approach is different:

1. The **Go backend generates JSON** — page handler functions in Go produce AMIS-format JSON schema objects that describe the entire UI: layout, forms, tables, buttons, data sources, and actions.
2. The **AMIS renderer displays it** — the browser's AMIS renderer receives this JSON and draws the UI. The renderer itself is a general-purpose engine; all the intelligence about what to show is in the JSON.

This means that **showing or hiding a UI element is a backend decision**, not a frontend decision. The backend either includes a component in the JSON response or it does not. There is no client-side feature flag evaluation, no client-side configuration reading, and no client-side permission checking. The UI is what the server says it is.

This architecture has a significant implication for the System Metadata Service: metadata is evaluated **once, on the server, before the JSON is emitted**. The client never needs to know about flags or configuration.

### 13.2 The Gating Chain

Before any AMIS UI JSON is returned to the browser, it passes through a sequential gating chain. Each gate must pass before the next is evaluated:

```
Browser requests a page
         │
         ▼
┌─────────────────────────┐
│   1. IAM SESSION GATE   │  "Is this request authenticated? Does a valid session exist?"
│                         │  → No session → 401 Unauthorised
│  Session carries:       │  → Session expired → 401 + redirect to login
│  - user_id              │
│  - tenant_id            │
│  - permissions[]        │
└───────────┬─────────────┘
            │ Session valid
            ▼
┌─────────────────────────┐
│  2. CONFIGURATION GATE  │  "Is the system correctly configured to serve this page?"
│                         │  → Missing required config → 503 with admin guidance
│  Checks:                │  → Config value invalid → log + use safe default
│  - Required keys exist  │
│  - Values are valid     │
└───────────┬─────────────┘
            │ Config valid
            ▼
┌─────────────────────────┐
│  3. FEATURE FLAG GATE   │  "Is this capability switched on for this tenant?"
│                         │  → Flag off → 404 (feature does not exist) or
│  Evaluates:             │    omit component from JSON silently
│  - Flag state           │
│  - Tenant targeting     │
└───────────┬─────────────┘
            │ Flag on
            ▼
┌─────────────────────────┐
│  4. AUTHORISATION GATE  │  "Does this user have permission to perform this action?"
│                         │  → Permission missing → 403 Forbidden
│  Checks session's       │  → Some permissions missing → partial page (read-only)
│  permissions[] against  │
│  required permission    │
└───────────┬─────────────┘
            │ Authorised
            ▼
    Go function generates
    AMIS JSON response
         │
         ▼
    Browser renders UI
```

### 13.3 What Each Gate Means in Practice

**IAM Session Gate** — This gate is managed entirely by the auth middleware. The session object carries pre-loaded permissions (no per-request DB lookup), the tenant ID, and the user ID. The System Metadata Service is not involved at this gate — it reads the session to *use* it, not to create it.

**Configuration Gate** — The Go page-generation function reads configuration keys required for the page. If a required key is missing or invalid (e.g., `finance.accounts.base_currency` is not set for a tenant that is trying to use the finance module), the function cannot safely generate the page and returns an error. This gate is the System Metadata Service's first involvement.

**Feature Flag Gate** — The Go page-generation function evaluates the flag for the module being accessed. If `lpg_module_enabled` is `false` for the requesting tenant, the LPG page handler returns `404` or simply does not include the LPG navigation component in a composite page's JSON. This is the System Metadata Service's second involvement.

**Authorisation Gate** — This gate checks the session's pre-loaded `permissions[]` array against the permission string required by the route (`system_metadata.configuration.read`, `finance.accounts.read`, etc.). The System Metadata Service is not involved here — the Casbin/CEL authz system handles this.

### 13.4 How the Go Page Function Reads Metadata

In practical terms, a Go function generating an AMIS page for the finance module would:

```go
// 1. Extract tenant from context (set by TenantMiddleware)
tenantID, _ := shared.GetTenantID(ctx)

// 2. Read required configuration
currency, err := deps.MetadataService.GetConfig(ctx, tenantID, "finance.accounts.base_currency")
if err != nil {
    return nil, ErrMissingRequiredConfig // → 503
}

// 3. Evaluate feature flag
if !deps.MetadataService.IsFlagEnabled(ctx, tenantID, "advanced_analytics_enabled") {
    // Omit analytics component from the JSON schema
}

// 4. Generate AMIS JSON using the config values
return generateFinancePageSchema(currency, showAnalytics), nil
```

The AMIS JSON that comes out of this function is already personalised, permission-gated, and feature-adjusted for the specific tenant and user. The browser renders exactly what it receives — it makes no further decisions.

### 13.5 What a User Sees at Each Gate Failure

| Gate | Failure condition | User experience |
|---|---|---|
| IAM Session | Not logged in | Redirect to login page |
| IAM Session | Session expired | "Your session has expired. Please log in again." |
| Configuration | Required config missing | "This feature is not yet configured. Please contact your administrator." |
| Feature Flag | Flag off for tenant | Page/menu item simply does not appear (silent) or "This module is not available for your account." |
| Authorisation | Permission missing | "You do not have permission to access this." (403 page) |
| Authorisation | Partial permissions | Page renders in read-only mode; action buttons absent |

### 13.6 Developer Guidance: Gating a New AMIS Page

When adding a new AMIS page or component that should be gated behind a flag or configuration:

1. **Register the flag or config key** in the System Metadata Service schema before writing any handler code.
2. **Evaluate the flag/config in the Go page handler**, not in a middleware (middleware gates routes; business gating happens in the handler).
3. **Return a clean 404**, not 403, when a flag is off — the feature does not exist for this tenant; there is no authorisation question.
4. **Omit the navigation entry** from composite pages when the flag is off — do not render a greyed-out link.
5. **Never expose flag or config keys** in the AMIS JSON response — the client should not need to know why something is absent.

---

## 14. System Metadata as a Cross-Cutting Concern

### 14.1 The Dependency Map

The System Metadata Service is one of the few services in AWO ERP that is consumed by almost every other module. The dependency direction is always inward — other modules depend on System Metadata; System Metadata does not depend on any other business module.

```
                    ┌──────────────────────────────────────────┐
                    │         SYSTEM METADATA SERVICE           │
                    │  Config · Flags · Op Meta · Module Reg   │
                    │  Entity Schema · Custom Fields · Versions │
                    └──────────────────┬───────────────────────┘
                                       │  read
         ┌─────────────────────────────┼──────────────────────────────┐
         │                             │                              │
         ▼                             ▼                              ▼
 ┌───────────────┐         ┌──────────────────────┐        ┌──────────────────┐
 │  IAM / Auth   │         │    Finance Module     │        │  AMIS UI Pages   │
 │  session boot │         │  tax rates, currency  │        │  page generation │
 │  SSO config   │         │  account defaults     │        │  form schemas    │
 └───────────────┘         └──────────────────────┘        └──────────────────┘
         │                             │                              │
         ▼                             ▼                              ▼
 ┌───────────────┐         ┌──────────────────────┐        ┌──────────────────┐
 │    Tenant     │         │   Station Module      │        │ Temporal Workers │
 │ Provisioning  │         │  payment modes,       │        │  flag checks     │
 │ seed defaults │         │  thresholds           │        │  mid-workflow    │
 └───────────────┘         └──────────────────────┘        └──────────────────┘
         │                             │
         ▼                             ▼
 ┌───────────────┐         ┌──────────────────────┐
 │ Audit Service │         │   Entity Module       │
 │ records all   │         │   reads field         │
 │ meta writes   │         │   descriptors for     │
 └───────────────┘         │   validation & forms  │
                           └──────────────────────┘
```

### 14.2 IAM / Auth Module

The IAM module consults System Metadata at two points:

**Session bootstrap** — When a user logs in, the session service reads tenant-level configuration to populate the session with context-relevant defaults (timezone for log timestamps, locale for number formatting, etc.).

**SSO configuration** — If a tenant has SSO enabled, the SSO provider details are stored as integration configuration keys (`integrations.sso.provider`, `integrations.sso.metadata_url`). The SSO service reads these before initiating the OAuth flow.

The IAM module does **not** read feature flags — session validity is independent of feature state. A user may have a valid session while a flag is off; the flag controls what they can do, not whether they're authenticated.

### 14.3 Finance Module

The Finance module is the heaviest consumer of configuration within AWO ERP:

- Default currency (`finance.accounts.base_currency`) is read when creating invoices, transactions, and reports
- Tax rates (`finance.tax.*`) drive automatic tax calculation on sales and purchases
- Account code defaults (`finance.accounts.default_*`) pre-populate new account setups
- Fiscal year configuration (`finance.periods.fiscal_year_start`) drives period-close logic

The Finance module also has feature flags: `advanced_analytics_enabled` gates the new reporting dashboard, and `multi_currency_enabled` gates the experimental multi-currency ledger.

### 14.4 Tenant Provisioning

When a new tenant is created, the Tenant Provisioning service triggers a **metadata seeding step** via the System Metadata Service. This step:

1. Copies global default values for all standard configuration keys into a tenant-scoped record for the new tenant
2. Sets all tenant entitlement feature flags to `false` (disabled) for the new tenant
3. Creates initial operational metadata entries (`onboarding.wizard_step = 1`, `onboarding.completed = false`)

This seeding is idempotent — running it twice on the same tenant is safe; it will not overwrite values that have been manually set since initial provisioning.

### 14.5 Audit Service

The Audit Service has a one-way dependency from System Metadata: every write to the metadata service generates an audit event. The Audit Service does not read from System Metadata.

The relationship is: System Metadata uses the Audit Service, not the other way around. This means the Audit Service must initialise before the System Metadata Service in the dependency graph.

### 14.6 Temporal Workflows

Temporal workflows are long-running processes that may execute over minutes, hours, or days. During execution, they should respect the current state of feature flags, not the state at the time they were started.

The pattern: workflows do not receive flag values as input parameters. They call a **metadata activity** at the point in their execution where they need to check a flag. This ensures that a kill switch activated mid-workflow takes effect at the next activity boundary.

```
Workflow starts
    │
    ▼
Activity 1: Process data
    │
    ▼
Activity 2: Check metadata flag ◄── reads CURRENT flag state from System Metadata
    │
    ├── Flag ON  → Activity 3: Send notification
    │
    └── Flag OFF → Activity 3: Skip notification (graceful degradation)
```

### 14.7 Rules of Engagement — What Other Modules May and Must Not Do

**Other modules MAY:**
- Read configuration values for their module namespace
- Read feature flag states relevant to their functionality
- Write operational metadata keys they own (prefixed with their module name)
- Cache metadata values locally for the duration of a single request

**Other modules MUST NOT:**
- Write to configuration or feature flag tables directly — all writes go through the System Metadata Service
- Hold metadata values in long-lived in-process caches beyond a single request (use the shared Redis cache via the service interface)
- Read metadata from the database directly bypassing the service — this would bypass cache, audit, and RLS
- Evaluate flag logic themselves — always call `MetadataService.IsFlagEnabled()`, never implement evaluation in the calling module

---

# Part IV — Guidance by Audience

---

## 15. Guidance for Business Administrators

### 15.1 How to Think About Configuration vs Feature Flags

As a business administrator, you will encounter two types of settings in AWO ERP:

**Settings you can change yourself** through the administration panel — these are configuration values. Changing them adjusts *how* the system works. Examples: your business timezone, accepted payment methods, tax rates, report display preferences. These changes take effect immediately (within a few seconds as the system updates its cache) and are fully reversible.

**Capabilities that require platform engineer involvement** to switch on — these are feature flags. Examples: activating the LPG module, enabling a new reporting dashboard, switching on an integration. You can request these through your account manager or support channel; you cannot change them yourself through the admin UI because they affect system-level behaviour and require technical verification.

### 15.2 What You Can Change vs What Requires a Developer

| Change | Who can make it | Where |
|---|---|---|
| Business timezone | Tenant admin | System Settings in the admin panel |
| Accepted payment methods | Tenant admin | Station Configuration in the admin panel |
| Default tax rate | Tenant admin | Finance Settings |
| Report rows per page | Any user | User Preferences |
| Enabling a new module (e.g., LPG) | Platform engineer only | Not accessible via admin panel |
| Switching off a broken integration | Platform engineer only | Emergency contact required |
| Resetting the onboarding wizard | Tenant admin | Administration → Onboarding |

### 15.3 Understanding the Impact of a Configuration Change

Before changing a configuration value, consider:

- **Tax rate changes** take effect on the *next* transaction, not retroactively. Historical transactions are not affected.
- **Currency changes** affect display formatting only. Existing monetary values in the database are stored as numbers; changing the currency label does not convert them.
- **Payment method changes** take effect immediately. If you remove "Shell Card" from accepted methods, the system will refuse Shell Card payments from that moment.
- **All changes are audited.** Every configuration change records your name, the timestamp, and what the previous value was. Changes can always be reversed.

### 15.4 What to Do When Something Disappears from the UI

If a page, menu item, or button disappears, there are three possible causes:

1. **A feature flag was disabled** — the capability has been switched off, possibly as a precaution. Contact your account manager or platform engineer.
2. **Your permissions changed** — your role may have been adjusted. Check with your administrator.
3. **A required configuration value is missing** — a configuration key needed by that page was not set. The admin panel's System Health section will show configuration warnings.

Do not assume a UI element disappearing means a bug. It usually means one of the three gates described in Section 13.2 did not pass.

### 15.5 Managing Custom Fields

As a tenant administrator, you can extend AWO ERP's built-in screens with your own fields through **Administration → Custom Fields**.

**Creating a custom field:**
1. Select the entity you want to extend (e.g., "Fuel Sale", "Employee")
2. Enter a field name, select a type (text, number, date, yes/no, dropdown), and optionally mark it as required
3. Save — the field appears immediately in all forms and list views for that entity across your tenant

**Things to know before creating custom fields:**
- Custom fields are for your tenant only. Other AWO ERP tenants cannot see them and are not affected.
- Disabling a field hides it from forms but preserves all data stored in it. You can re-enable it at any time.
- Deleting a field permanently removes it and all data stored in it. This cannot be undone.
- There is a limit of 20 custom fields per entity. Contact your account manager if you need more.

**What you cannot do with custom fields:** You cannot create fields that link to other entities (e.g., "link this sale to a supplier record") or fields that perform calculations. Those require a developer to update the base entity schema.

### 15.6 How to Request a Feature Flag Be Enabled

To request that a new module or feature be enabled for your tenant:

1. Contact your AWO ERP account manager or support channel
2. Specify: which feature you need, the name of your tenant/organisation, and the desired activation date
3. The platform engineer will verify the feature is stable for production, enable the flag for your tenant, and confirm activation
4. You will see the new capability appear in your admin panel and navigation within minutes of the flag being enabled

---

## 16. Guidance for Frontend / AMIS UI Developers

### 16.1 The Core Principle

In AWO ERP's AMIS architecture, **all gating decisions are made in Go, on the server, before JSON is emitted.** You never write client-side flag evaluation. You never read configuration values from JavaScript. The AMIS renderer receives already-resolved JSON.

Your responsibility as an AMIS UI developer is to: **write Go page-generation functions that correctly read metadata and generate the appropriate JSON schema.**

### 16.2 Reading Configuration in Go Page Handlers

When your page or component needs a configuration value:

```go
func (h *StationHandler) GenerateDashboardPage(c *fiber.Ctx) error {
    ctx := c.UserContext()
    tenantID, _ := shared.GetTenantID(ctx)

    // Read a required config value
    paymentMethods, err := h.deps.MetadataService.GetConfigStringSlice(
        ctx, tenantID, "station.fuel.accepted_payment_methods",
    )
    if err != nil {
        // Config is required — cannot render this page without it
        return fiber.NewError(fiber.StatusServiceUnavailable,
            "Station payment configuration is missing. Please contact your administrator.")
    }

    // Read an optional config value with a default
    rowsPerPage := h.deps.MetadataService.GetConfigIntOrDefault(
        ctx, tenantID, "ui.dashboard.rows_per_page", 25,
    )

    // Build the AMIS JSON using these values
    schema := buildDashboardSchema(paymentMethods, rowsPerPage)
    return c.JSON(schema)
}
```

### 16.3 Checking Feature Flag State Before Emitting a Component

```go
func (h *StationHandler) GenerateNavigationMenu(c *fiber.Ctx) error {
    ctx := c.UserContext()
    tenantID, _ := shared.GetTenantID(ctx)

    menu := &AMISNavMenu{}
    menu.AddItem(navItem("dashboard", "/dashboard"))
    menu.AddItem(navItem("fuel_sales", "/fuel/sales"))

    // Only include LPG menu item if the flag is enabled for this tenant
    if h.deps.MetadataService.IsFlagEnabled(ctx, tenantID, "lpg_module_enabled") {
        menu.AddItem(navItem("lpg_supply", "/lpg/supply"))
    }

    // Only include analytics if flag is on
    if h.deps.MetadataService.IsFlagEnabled(ctx, tenantID, "advanced_analytics_enabled") {
        menu.AddItem(navItem("analytics", "/analytics"))
    }

    return c.JSON(menu)
}
```

### 16.4 Handling the Disabled/Hidden State Gracefully

When a flag is off, **do not render a greyed-out or locked version of the component** unless you have a specific design reason. The default behaviour should be total absence — the item simply does not appear.

This is correct:
```go
if h.deps.MetadataService.IsFlagEnabled(ctx, tenantID, "hotel_module_enabled") {
    schema.AddTab(hotelTab())  // present only when enabled
}
```

This is incorrect (unless you have a "coming soon" upgrade prompt):
```go
hotelTab := buildHotelTab()
if !h.deps.MetadataService.IsFlagEnabled(ctx, tenantID, "hotel_module_enabled") {
    hotelTab.Disabled = true  // greyed-out locked tab — avoid this pattern
}
schema.AddTab(hotelTab)
```

### 16.5 Do's and Don'ts

| Do | Don't |
|---|---|
| Evaluate flags in the Go handler, before building the schema | Put flag or config keys in the AMIS JSON response |
| Return 404 for flag-off page routes | Return 403 for flag-off features (they don't exist; it's not an auth question) |
| Return 503 for missing required configuration | Silently produce a broken page with missing data |
| Use `GetConfigOrDefault` for optional config with sensible defaults | Assume a key exists without error handling |
| Read metadata once per request at the handler level | Make multiple metadata reads for the same key within a single request |

---

## 17. Guidance for Backend Developers

### 17.1 When to Add a New Configuration Key vs a New Feature Flag vs Neither

Use this decision tree:

```
Is this a new piece of adjustable knowledge?
│
├── Is it about WHETHER something exists?
│   └── Yes → Feature Flag
│
├── Is it about HOW something should work?
│   └── Yes → Configuration Key
│
├── Is it runtime state the app writes to itself?
│   └── Yes → Operational Metadata
│
└── Is it static infrastructure data (DB URL, API secret)?
    └── Yes → Environment variable / Viper config
              (NOT in System Metadata Service)
```

**Do not add a metadata key for:**
- Values that never change (embed them as constants)
- Values that belong to a specific business entity (put them on the entity's table)
- Values that are user preferences surfaced through a user settings UI (those are user profile, not system metadata)

### 17.2 Consuming System Metadata from Another Module

The System Metadata Service exposes an interface that all other modules use. Never import the metadata package's implementation directly — always depend on the interface. This keeps your module testable and the metadata service replaceable.

```go
// In your module's service constructor
type MyService struct {
    metadata system_metadata.Service  // interface, not concrete type
    // ... other deps
}

func NewMyService(metadata system_metadata.Service, ...) (*MyService, error) {
    return &MyService{metadata: metadata}, nil
}

// In your service method
func (s *MyService) DoSomething(ctx context.Context, tenantID uuid.UUID) error {
    // Check flag
    if !s.metadata.IsFlagEnabled(ctx, tenantID, "my_module.feature_x_enabled") {
        return ErrFeatureNotAvailable
    }

    // Read config
    threshold, err := s.metadata.GetConfigInt(ctx, tenantID, "my_module.threshold")
    if err != nil {
        return fmt.Errorf("missing required config: %w", err)
    }

    // ... business logic using threshold
}
```

### 17.3 Caching Rules

The System Metadata Service caches values in Redis. The rules you must follow as a consumer:

- **Do not cache metadata values yourself.** The service's Redis cache already handles this. Double-caching creates consistency problems.
- **Do not hold metadata values across requests.** Read them at the start of each request. The in-process cost is negligible (Redis roundtrip) and is much safer than stale in-process state.
- **Cache invalidation is automatic.** When a metadata value changes, the System Metadata Service invalidates the relevant Redis keys. You do not need to do anything.
- **Cache TTL is short for feature flags** (seconds), longer for configuration (minutes). This ensures kill switches take effect quickly.

### 17.4 Writing Migrations for New Metadata Keys

New metadata keys are registered via migration. Never insert metadata keys directly in application startup code — that creates race conditions and makes the key's origin untraceable.

Migration naming convention: `NNN_system_metadata_add_<module>_<key_name>.sql`

```sql
-- db/migration/001025_system_metadata_add_lpg_base_margin.up.sql

INSERT INTO system_metadata_schema (
    key,
    scope,           -- 'tenant' | 'global' | 'user'
    type,            -- 'string' | 'integer' | 'decimal' | 'boolean' | 'enum' | 'json'
    default_value,
    description,
    module,
    is_required,
    allowed_values,  -- JSON array for enum types, null otherwise
    status           -- 'active'
) VALUES (
    'lpg.pricing.base_margin_percent',
    'tenant',
    'decimal',
    '0.05',
    'Base margin percentage applied to LPG supply pricing for this tenant.',
    'lpg',
    false,
    null,
    'active'
);
```

### 17.5 Registering a New Module in the Module Registry

Every new AWO ERP module must be registered in the Module Registry before it can be enabled for any tenant. Registration happens via a migration, not application startup code.

```sql
-- db/migration/001030_module_registry_add_car_wash.up.sql

INSERT INTO module_registry (
    module_id,
    display_name,
    description,
    status,           -- 'development' | 'beta' | 'stable' | 'deprecated' | 'disabled'
    capabilities,     -- JSON array of capability names
    dependencies,     -- JSON array of module_ids this module requires
    config_schema_keys -- JSON array of config key prefixes this module introduces
) VALUES (
    'car-wash',
    'Car Wash',
    'Tracks car wash service bookings, attendant assignments, and revenue.',
    'beta',
    '["car_wash_bookings", "car_wash_revenue_tracking", "car_wash_attendant_assignment"]',
    '["station-fuel", "finance"]',
    '["car_wash.*"]'
);
```

After registration, the corresponding feature flag (`car_wash_module_enabled`) must also exist in the feature flag table. The flag and the registry entry are separate records — the registry describes what the module is; the flag controls whether a tenant has it.

### 17.6 Adding Fields to Entity Metadata

When you add a column to a business entity's database table via a migration, you must also add the corresponding field descriptor to entity metadata. These two operations should be in the same PR and ideally part of the same migration script.

```sql
-- Part 1: database column (in the entity's table migration)
ALTER TABLE fuel_sales ADD COLUMN pump_nozzle_id VARCHAR(20) NULL;

-- Part 2: entity metadata field descriptor (in the same or accompanying migration)
INSERT INTO entity_field_descriptors (
    entity_id,
    field_id,
    display_name,
    description,
    data_type,
    nullable,
    read_only,
    system_managed,
    searchable,
    sortable,
    display_order
) VALUES (
    'fuel_sale',
    'pump_nozzle_id',
    'Pump Nozzle ID',
    'The identifier of the specific pump nozzle used for this dispensing.',
    'string',
    true,
    false,
    false,
    true,
    false,
    12  -- position after existing fields
);

-- This triggers a new schema version snapshot automatically via a DB trigger
```

A DB trigger on the `entity_field_descriptors` table creates a new schema version snapshot whenever a row is inserted, updated, or deleted. You do not need to create the snapshot manually.

### 17.7 Consuming Entity Metadata from a Handler

When writing a Go handler that generates an AMIS form for a business entity, use the `EntityMetadataService` to load field descriptors rather than hardcoding the form schema:

```go
func (h *FuelSaleHandler) GenerateCreateForm(c *fiber.Ctx) error {
    ctx := c.UserContext()
    tenantID, _ := shared.GetTenantID(ctx)

    // Load base entity fields
    fields, err := h.deps.EntityMetadataService.GetFieldDescriptors(ctx, "fuel_sale")
    if err != nil {
        return fiber.ErrInternalServerError
    }

    // Load tenant custom fields for this entity
    customFields, err := h.deps.MetadataService.GetCustomFields(ctx, tenantID, "fuel_sale")
    if err != nil {
        return fiber.ErrInternalServerError
    }

    // Merge and sort by display_order
    allFields := mergeFields(fields, customFields)

    // Generate AMIS form schema from descriptors
    form := buildAMISFormFromDescriptors(allFields)
    return c.JSON(form)
}
```

This pattern ensures form schemas stay in sync with the entity metadata automatically. A new field added to the descriptor appears in all forms that use this pattern without any handler changes.

Use the mock generated by `mockgen` for the `system_metadata.Service` interface. Never use a real database in unit tests of code that depends on metadata.

```go
// In your test file
func TestMyService_DoSomething(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockMeta := mock_system_metadata.NewMockService(ctrl)

    // Test: flag is off
    mockMeta.EXPECT().
        IsFlagEnabled(gomock.Any(), tenantID, "my_module.feature_x_enabled").
        Return(false)

    svc := NewMyService(mockMeta)
    err := svc.DoSomething(ctx, tenantID)
    assert.ErrorIs(t, err, ErrFeatureNotAvailable)

    // Test: flag is on, config is present
    mockMeta.EXPECT().
        IsFlagEnabled(gomock.Any(), tenantID, "my_module.feature_x_enabled").
        Return(true)
    mockMeta.EXPECT().
        GetConfigInt(gomock.Any(), tenantID, "my_module.threshold").
        Return(100, nil)

    err = svc.DoSomething(ctx, tenantID)
    assert.NoError(t, err)
}
```

---

## 18. Guidance for Architects and Platform Engineers

### 18.1 Extending the Metadata Type System

The current type system (string, integer, decimal, boolean, enum, json, duration) covers the known requirements. If a new type is needed:

1. Add the type constant to the domain type registry
2. Add validation logic to the write path in the service layer
3. Add serialisation/deserialisation logic for cache storage
4. Add a migration updating the `system_metadata_schema` type enum
5. Document the new type with examples in this guide

Do not add types speculatively. Every new type adds validation surface area. The `json` type already covers complex structured values.

### 18.2 Multi-Tenant Isolation Guarantees

The System Metadata Service's isolation guarantee: **no read or write operation can affect or observe another tenant's metadata.**

This is enforced at two levels:

- **Database level:** PostgreSQL Row-Level Security policies on all metadata tables enforce `tenant_id = current_setting('app.current_tenant_id')::uuid`. Queries that run outside a `store.WithTenant()` call cannot read any rows.
- **Service level:** Every service method takes `tenantID uuid.UUID` as an explicit parameter. The service sets the tenant context before any database operation. There is no global state.

The cache key includes the tenant ID: `metadata:{tenant_id}:{key}`. A cache read for Tenant A can never return a value for Tenant B.

### 18.3 Performance Considerations

Target latency for metadata reads: **< 1ms for cached reads, < 5ms for cache misses.**

Architectural decisions that achieve this:
- Redis cache with short TTL for flags (10s), longer for config (5 min)
- Cache is populated on first miss, not on write (lazy population)
- Keys are small (string key + scalar value) — no large object serialisation
- Bulk read API (`GetConfigBatch`) allows a page handler to fetch multiple keys in a single Redis roundtrip

Risk: a Redis outage degrades performance but must not cause system failure. The service must fall back to direct PostgreSQL reads when Redis is unavailable, accepting the latency cost. This fall-through must be implemented with a circuit breaker to prevent database overload during a Redis outage.

### 18.4 Flag Consistency Across Instances

AWO ERP may run multiple instances of the application server simultaneously. Flag state changes must propagate to all instances within the cache TTL. This means:

- A flag disabled at T=0 will be respected by all instances by T=10s (the flag cache TTL)
- For kill switches with sub-second requirements, consider implementing a Redis pub/sub invalidation channel that pushes a cache-bust event to all instances immediately on write

The current implementation uses TTL-based expiry. If sub-second kill-switch propagation becomes a requirement, the pub/sub channel is the prescribed upgrade path.

### 18.5 Operational Concerns — Cache Warm-Up

On cold start (new deployment, Redis flush), the metadata cache is empty. The first request for each tenant will be a cache miss, hitting PostgreSQL. Under normal traffic, this self-heals within seconds as values are cached on first read.

For deployments with high traffic at startup: implement a **warm-up routine** that reads the most commonly accessed configuration keys for all active tenants immediately after the application starts and before the load balancer routes traffic to the new instance.

### 18.6 Future Evolution

**Dynamic schema metadata** — Currently, module metadata keys are registered via SQL migrations. A future enhancement would allow modules to declare their metadata schema in Go structs with struct tags, auto-generating migrations. This would improve developer experience and reduce migration boilerplate.

**Client-push configuration** — Currently, configuration changes take effect within one cache TTL cycle. For real-time administrative UI, a WebSocket or SSE channel could push configuration change events to connected admin sessions so the UI reflects changes immediately without a page reload.

**Rollout percentages** — Feature flag rollout percentage targeting (e.g., enable for 10% of tenants) is designed into the evaluation model but not yet implemented. The evaluation logic in Section 7.5 shows the intended path.

---

# Part V — Reference

---

## 19. Operational Runbook

### 19.1 Metadata Not Reflecting After an Update

**Symptom:** An administrator changes a configuration value or enables a feature flag, but the system continues to behave as if the old value is still in effect.

**Cause:** The System Metadata Service uses Redis as a read-through cache. After a write, the old value persists in the cache until the TTL expires or an explicit invalidation event fires. Under normal operation, invalidation is automatic and near-instant. If it is not happening, the write may not have completed successfully, or the cache invalidation pathway may be broken.

**Diagnosis steps:**

1. Confirm the write actually succeeded — check the audit trail for an event matching the expected change. If no audit event exists, the write did not complete.
2. If the audit event exists, check the Redis cache directly for the key `metadata:{tenant_id}:{key}`. If the old value is still present, the cache invalidation did not fire.
3. Check the application logs for errors from the cache invalidation pathway. A Redis connection error at write time would cause the DB write to succeed but the cache to remain stale.

**Resolution:** If the cache contains a stale value, flush the specific key with `DEL metadata:{tenant_id}:{key}`. The next read will re-populate from the database. If cache invalidation is systematically broken, restart the application to clear in-process state and investigate the Redis connection health.

**Prevention:** Monitor the `metadata_cache_invalidation_errors_total` Prometheus counter. Alert if it rises above zero.

---

### 19.2 Custom Field Validation Rejections

**Symptom:** Users are receiving unexpected validation errors on custom fields, or records that previously saved successfully are now being rejected.

**Likely causes:**

- A tenant administrator tightened a validation rule on a custom field (e.g., reduced `max_length`, added a `pattern` constraint) after data was already stored under the old, less restrictive rule. Existing records are grandfathered but new writes must satisfy the new rule.
- A custom field was set from `nullable: true` to `nullable: false`, making it required. Any record that omits this field will now fail validation.

**Diagnosis steps:**

1. Retrieve the custom field definition: `GET /internal/metadata/custom-fields/{field_id}` and review the current `validation_rules`.
2. Check the field's version history to see what changed and when.
3. Compare the failing record's data against the current validation rules.

**Resolution:** Either adjust the failing record to satisfy the new rules, or (if the rule change was unintended) revert the validation rule to its previous state via the admin UI. Both actions generate audit events.

---

### 19.3 Schema Version Conflicts

**Symptom:** A data import, migration script, or external integration is failing with a schema mismatch error — referencing field names or types that no longer match the current entity schema.

**Cause:** The external process was built against an older version of the entity schema. A subsequent schema change removed or renamed a field the process depends on.

**Diagnosis steps:**

1. Identify the schema version the external process was built against (check its documentation or build timestamp).
2. Retrieve the diff between that version and the current active version: `GET /internal/metadata/diff?from=schema.fuel_sale.v{old}&to=schema.fuel_sale.v{current}`.
3. Identify the breaking changes (removed or renamed fields) in the diff.

**Resolution:** Update the external process to use current field names/types, or — if the removed field was a mistake — create a new schema version that restores the field. Never modify a historical snapshot to "fix" a version; create a new version instead.

---

### 19.4 Feature Flag Enabled but Feature Still Not Appearing

**Symptom:** A platform engineer has enabled a feature flag for a tenant, but the expected UI component or API endpoint still returns 404 or does not appear in navigation.

**This is almost always a multi-cause problem.** Work through the gating chain in order:

1. **Is the session valid?** Confirm the user's session has a fresh `tenant_id` claim. Old sessions pre-dating the flag change do not carry updated flag context. Ask the user to log out and back in.
2. **Is the configuration gate passing?** The feature may require a configuration key to be present. Check whether required configuration for this module has been seeded (tenant provisioning may have missed it). Look for a `503` response, which indicates a missing required config.
3. **Is the flag actually enabled?** Verify via the admin panel or internal API, not just by asking the engineer who enabled it. Confirm the `tenant_id` targeted matches the tenant reporting the issue.
4. **Does the module have all its dependencies active?** Check the Module Registry — if the module depends on `finance` and `finance_module_enabled` is false for this tenant, the module cannot activate.
5. **Is the user authorised?** Confirm the user's role includes the permission string for the new feature (`system_metadata.configuration.read`, or the module-specific permission).

---

### 19.5 Metadata Staleness Detection

The System Metadata Service exposes a health endpoint that reports staleness indicators:

- **Cache hit ratio** — if the ratio drops significantly, the cache may be cold or Redis may be unavailable. The system is falling back to direct DB reads, increasing latency.
- **Active version freshness** — compares the `updated_at` timestamp of schema active versions against the deployed code's expected schema versions. A mismatch indicates a deployment completed without its associated schema migration.
- **Module registry vs deployed modules** — detects modules present in the registry as `STABLE` that are not present in the running application's module list, or vice versa.

These indicators are surfaced on the platform admin's System Health dashboard and via Prometheus metrics. Alert thresholds should be configured as part of the production monitoring runbook.

---

### 19.6 Resetting Tenant Metadata to Defaults

Occasionally a tenant's configuration drifts into an unusable state and needs to be reset. This is a destructive operation that must be explicitly authorised.

**Scope options:**
- **Full reset** — removes all tenant-scoped configuration overrides, restoring global defaults. Does not affect custom fields or operational metadata.
- **Module reset** — removes configuration overrides for a specific module namespace only.
- **Single key reset** — restores one configuration key to its global default.

**Single key reset via admin UI:** Administration → System Configuration → select key → "Reset to Default."

**Full or module reset:** Requires `system_metadata.configuration.admin` permission. Generates a bulk audit event listing all keys reverted and their previous values. Cannot be undone — the previous values exist in the audit trail but must be manually re-applied if the reset was a mistake.

---

## 20. Design Decisions and Rationale

### 20.1 Why Configuration and Feature Flags Are Sub-Domains of One Service

The alternative was three separate services: a Configuration Service, a Feature Flag Service, and an Operational Metadata Service. This was rejected because:

- All three require identical tenant isolation infrastructure (RLS, context propagation)
- All three feed into the same gating chain; splitting them adds latency for every UI render
- All three have the same read/write ratio and caching strategy
- The team is small; operating three services means three deployment units, three schema migration tracks, three sets of monitoring alerts

The cost of the consolidation is a slightly larger service boundary. The benefit is dramatically simpler operations and consistent behaviour across all metadata concerns.

### 20.2 Why Metadata Is Not Stored in the Tenant Table

A simpler approach would add configuration columns to the `tenants` table. This was rejected because:

- It would require a schema migration every time a new configuration key is added
- It would make the tenant table enormous and poorly normalised
- It would not support the key schema (type, description, allowed values) needed for self-describing configuration
- It would not support global (cross-tenant) defaults

The System Metadata Service uses its own tables, allowing the metadata schema to evolve independently of the tenant schema.

### 20.3 Why Flag Evaluation Is Server-Side Only

Some feature flag systems provide client SDKs that evaluate flags in the browser or mobile app. AWO ERP does not do this because:

- The AMIS architecture makes server-side evaluation natural — the server generates the UI anyway
- Client-side flag evaluation requires sending flag state (and potentially targeting rules) to the client, which reveals product roadmap information and is a security surface
- Keeping all gating on the server means the client can never be manipulated to bypass a gate
- There is only one place to audit and debug gating decisions

### 20.4 Why Entity Metadata Lives in the Metadata Service (Not in Code Only)

An alternative approach is to define entity schemas purely in Go structs and PostgreSQL migrations, with no runtime metadata store. This is simpler initially but creates several problems at scale:

- **No introspection** — the UI cannot dynamically generate forms from the schema; it must be hardcoded or maintained separately
- **No self-description** — there is no way to ask the running system "what fields does a fuel_sale have?" and receive an authoritative answer
- **No tenant extension** — custom fields require either separate columns for each tenant (schema explosion) or a rigid EAV (Entity-Attribute-Value) pattern bolted on separately
- **No diff tooling** — comparing what the schema was last month against what it is now requires reading git history, not querying the system

Storing entity metadata in the System Metadata Service makes the data model a first-class citizen of the platform, introspectable, versioned, and extensible — at the cost of keeping metadata and migrations in sync. That sync cost is paid once per schema change, not on every read.

### 20.5 Why Custom Fields Use JSONB Storage (Not EAV or Separate Columns)

Three approaches exist for storing custom field data:

**Entity-Attribute-Value (EAV)** — a `(entity_id, field_id, value)` table where every custom field value is a separate row. Used by older ERPs. Problems: catastrophic query performance for entities with many custom fields (N+1 joins), poor type safety (all values stored as text), complex SQL for any query that filters by custom field value.

**Separate columns per tenant** — adding a `tenant_X_custom_1` style column to the entity table. Problems: schema explosion as tenants scale, requires migrations per tenant, impossible to implement per-tenant field labels.

**JSONB column** — a single `custom_fields JSONB` column on the entity table, containing all tenant-defined field values as a JSON object keyed by `field_id`. AWO ERP's chosen approach. Benefits: no schema migrations per tenant, GIN-indexed for query performance, type validation enforced at the service layer using the Custom Fields metadata definitions. Trade-off: JSONB queries are slightly more complex than column-based queries, and JSONB cannot enforce column-level constraints at the DB level.

---

## 21. Full Glossary

| Term | Definition |
|---|---|
| AMIS | A JSON-schema-driven UI framework. AWO ERP uses it by generating AMIS-format JSON in Go handlers, which the browser's AMIS renderer displays. |
| Audit Trail | A tamper-evident log of every write to the System Metadata Service, including actor, timestamp, previous value, and new value. |
| Cache TTL | Time To Live — how long a cached metadata value remains valid before being re-fetched from the database. |
| Configuration | The sub-domain of System Metadata that stores typed, named settings describing how the system behaves. |
| Feature Flag | The sub-domain of System Metadata that controls whether capabilities are active. Always evaluates to a boolean. |
| Flag Evaluation | The process of computing whether a feature flag is on for a given tenant request, considering flag state and targeting rules. |
| Gating Chain | The sequential set of checks (IAM → Configuration → Feature Flag → Authorisation) that must all pass before a UI page is rendered. |
| Global scope | A metadata key that applies to all tenants unless overridden at the tenant level. |
| Kill Switch | A feature flag used defensively to instantly disable a capability in production without redeployment. |
| Operational Metadata | The sub-domain of System Metadata that stores transient system state (onboarding progress, process timestamps). |
| RLS | Row-Level Security — a PostgreSQL feature that enforces tenant isolation at the database query level. |
| Scope | The level (global, tenant, user) at which a metadata entry applies. |
| Seeding | The process of creating default metadata entries for a newly provisioned tenant. |
| System Metadata Service | The AWO ERP service that manages all six metadata sub-domains: Configuration, Feature Flags, Operational Metadata, Module Registry, Entity & Schema Metadata, and Custom Fields. |
| Tenant | An isolated business using AWO ERP; equivalent to a "client" in SAP or a "site" in ERPNext. |
| Tenant entitlement flag | A feature flag that controls module access for a specific tenant, based on commercial or readiness decisions. |
| Module Registry | The sub-domain cataloguing every module in AWO ERP — its capabilities, dependencies, configuration keys, and lifecycle status. |
| Entity Metadata | The sub-domain formally describing the structure of every business entity: fields, types, validation rules, relationships, and computed properties. |
| Field Descriptor | A metadata record describing one field on an entity — its name, type, nullability, validation rules, and display properties. |
| Custom Fields | Tenant-defined extensions to built-in entity schemas, stored in a JSONB column and managed through the Custom Fields sub-domain. |
| Schema Version | An immutable snapshot of a schema-bearing sub-domain at a point in time, used for audit, diff, and rollback. |
| Module Dependency | A declared requirement that one module needs another to be active before it can function. Enforced at flag-enable time. |

---

## 22. Related Documents

| Document | Description |
|---|---|
| **AWO ERP IAM & RBAC Guide** | Covers the session model, permission strings, Casbin/CEL policy evaluation, and the user-role-permission hierarchy. Required reading for understanding Gate 1 (IAM Session) and Gate 4 (Authorisation) in the gating chain. |
| **Tenant Provisioning Guide** | Covers the tenant lifecycle (PENDING → ACTIVE → SUSPENDED → ARCHIVED), the provisioning process, and the metadata seeding step that occurs on new tenant creation. |
| **AMIS UI Page Generation Guide** | Covers the Go → JSON → AMIS renderer pipeline in detail, including schema conventions, component library, and data source binding patterns. |
| **RLS & Multi-Tenancy Reference** | Covers the PostgreSQL RLS implementation, `store.WithTenant()` pattern, and the `TenantMiddleware` that sets the RLS context on every request. |
| **Audit Service Guide** | Covers the audit event schema, storage, and query interface. Relevant for understanding what the audit trail for metadata changes looks like. |
| **AWO ERP Architecture Skill** | The developer-facing codebase reference (`SKILL.md`) encoding file paths, struct names, and patterns used across the entire project. |

---

*Document version: 1.0 — May 2026*
*Maintained by: AWO ERP Platform Team*
*Next review: when any of the three metadata sub-domains receives significant new functionality*
