# AwoERP Server-Driven UI Platform
## Official Architecture Documentation — Table of Contents & Documentation Plan

> **Status:** Proposed TOC — Pending Review & Approval
> **Version:** 0.1.0-draft
> **Authors:** Architecture Team
> **Audience:** Platform Engineers · Backend Engineers · Mobile Engineers · Web Engineers · UI Framework Developers · Product Owners · Solution Architects · ERP Implementers · Third-Party Integrators

---

## Document Map

This TOC organizes the entire AwoERP SDUI platform documentation into **10 volumes**, **50 chapters**, and approximately **200+ sections**. It is structured so that:

- Any engineer can locate their domain quickly.
- The documentation can be built incrementally across 4 phases.
- Each chapter maps to a discrete, ownable unit of work.

---

## Suggested Repository Structure

```
./docs/ui/
├── README.md                          # Docs home, quick start, contribution guide
├── GLOSSARY.md                        # Platform-wide terminology
├── CHANGELOG.md                       # Documentation changelog
├── mkdocs.yml                         # Or docusaurus.config.js / docs site config
│
├── vol-01-vision/                     # Volume I: Vision & Philosophy
│   ├── 01-introduction.md
│   ├── 02-design-philosophy.md
│   ├── 03-architectural-principles.md
│   └── 04-system-overview.md
│
├── vol-02-dsl-and-ast/                # Volume II: DSL & AST
│   ├── 05-sdui-fundamentals.md
│   ├── 06-ui-dsl-architecture.md
│   ├── 07-ast-design.md
│   └── 08-json-compilation-pipeline.md
│
├── vol-03-component-system/           # Volume III: Component System
│   ├── 09-component-system.md
│   ├── 10-layout-system.md
│   ├── 11-forms-framework.md
│   ├── 12-tables-and-data-grids.md
│   ├── 13-dashboard-framework.md
│   ├── 14-charts-and-analytics.md
│   ├── 15-workflow-and-approval-components.md
│   └── 16-erp-specific-components.md
│
├── vol-04-rendering/                  # Volume IV: Rendering Architecture
│   ├── 17-mobile-rendering-architecture.md
│   ├── 18-web-rendering-architecture.md
│   ├── 19-navigation-framework.md
│   ├── 20-action-system.md
│   └── 21-event-system.md
│
├── vol-05-runtime/                    # Volume V: Runtime & State
│   ├── 22-state-management.md
│   ├── 23-validation-framework.md
│   └── 24-data-sources.md
│
├── vol-06-platform/                   # Volume VI: Platform & API
│   ├── 25-api-contracts.md
│   ├── 26-security-model.md
│   ├── 27-authorization-integration.md
│   ├── 28-multi-tenant-architecture.md
│   ├── 29-feature-flag-architecture.md
│   ├── 30-customization-framework.md
│   └── 31-extension-and-plugin-architecture.md
│
├── vol-07-reliability/                # Volume VII: Reliability & Performance
│   ├── 32-performance-optimization.md
│   ├── 33-offline-support.md
│   ├── 34-caching-strategies.md
│   └── 35-synchronization.md
│
├── vol-08-dx/                         # Volume VIII: Developer Experience
│   ├── 36-internationalization.md
│   ├── 37-accessibility.md
│   ├── 38-observability.md
│   ├── 39-telemetry.md
│   └── 40-logging.md
│
├── vol-09-engineering/                # Volume IX: Engineering Excellence
│   ├── 41-testing-strategy.md
│   ├── 42-cicd-considerations.md
│   ├── 43-versioning-strategy.md
│   ├── 44-migration-strategy.md
│   └── 45-governance-model.md
│
├── vol-10-reference/                  # Volume X: Reference & Guidance
│   ├── 46-best-practices.md
│   ├── 47-anti-patterns.md
│   ├── 48-reference-architecture.md
│   ├── 49-end-to-end-examples.md
│   ├── 50-sdk-development.md
│   └── 51-future-roadmap.md
│
├── schemas/                           # Canonical JSON schemas
│   ├── component.schema.json
│   ├── layout.schema.json
│   ├── action.schema.json
│   ├── event.schema.json
│   ├── page.schema.json
│   └── ...
│
├── examples/                          # Runnable reference examples
│   ├── purchase-order-form/
│   ├── invoice-dashboard/
│   ├── approval-workflow/
│   └── multi-tenant-demo/
│
└── assets/
    ├── diagrams/                      # Architecture diagrams (source + exported)
    └── images/
```

---

## Estimated Documentation Phases

| Phase | Volumes | Chapters | Focus | Priority |
|-------|---------|----------|-------|----------|
| **Phase 1** — Foundation | I, II | 01–08 | Vision, DSL, AST, compilation pipeline | Critical / Ship first |
| **Phase 2** — Components & Rendering | III, IV | 09–21 | All components, web & mobile rendering | Critical / Ship second |
| **Phase 3** — Platform & Runtime | V, VI, VII | 22–35 | State, APIs, security, tenancy, performance | High |
| **Phase 4** — Excellence & Reference | VIII, IX, X | 36–51 | DX, testing, governance, SDK, examples | Medium |

---

---

# Full Hierarchical Table of Contents

---

## Volume I — Vision & Philosophy

*Establishes the "why" and "what" of the platform. Every engineer and stakeholder should read this volume before touching any other chapter.*

---

### Chapter 01 — Introduction to AwoERP UI Platform

> **Audience:** All · **Phase:** 1 · **Status:** Draft first

- 1.1 What Is AwoERP?
- 1.2 The Problem This Platform Solves
  - 1.2.1 Traditional ERP UI Fragmentation
  - 1.2.2 The Multi-Platform Maintenance Burden
  - 1.2.3 Tenant Customization Without Code Forks
- 1.3 Platform Mission Statement
- 1.4 Key Value Propositions
  - 1.4.1 One Backend, Many Surfaces
  - 1.4.2 Business Logic Drives UI, Not the Inverse
  - 1.4.3 ERP-Native Component Vocabulary
  - 1.4.4 Permission-Aware UI at the Source
- 1.5 Who Should Read This Documentation
  - 1.5.1 Platform Engineers
  - 1.5.2 Backend (Go/Goa) Engineers
  - 1.5.3 Mobile Engineers
  - 1.5.4 Web Engineers
  - 1.5.5 UI Framework Developers
  - 1.5.6 Product Owners & Solution Architects
  - 1.5.7 ERP Implementers
  - 1.5.8 Third-Party Integrators
- 1.6 How to Use This Documentation
- 1.7 Relationship to Other AwoERP Documentation
- 1.8 Glossary Reference

---

### Chapter 02 — Design Philosophy

> **Audience:** All · **Phase:** 1

- 2.1 The Server-Driven UI Manifesto
- 2.2 Inspirations and Prior Art
  - 2.2.1 AMIS (Baidu Low-Code Framework)
  - 2.2.2 Retool's Declarative Component Model
  - 2.2.3 Appsmith's Event-Action System
  - 2.2.4 PowerApps Formula-Driven Bindings
  - 2.2.5 What AwoERP Does Differently
- 2.3 Core Design Tenets
  - 2.3.1 Tenet 1 — UI as Data, Not Code
  - 2.3.2 Tenet 2 — The Backend Is the Source of Truth
  - 2.3.3 Tenet 3 — Clients Are Rendering Engines, Not Decision Makers
  - 2.3.4 Tenet 4 — Components Speak Business, Not HTML
  - 2.3.5 Tenet 5 — Permissions Are First-Class Citizens
  - 2.3.6 Tenet 6 — Extensibility Without Forking
  - 2.3.7 Tenet 7 — Observability Built In
- 2.4 Design Trade-offs Accepted
  - 2.4.1 Accepting Backend Coupling for Consistency
  - 2.4.2 Accepting JSON Payload Overhead for Flexibility
  - 2.4.3 Accepting Rendering Engine Complexity for Platform Control
- 2.5 Non-Goals of This Platform
- 2.6 Philosophy vs. Implementation — Where Rules Can Flex

---

### Chapter 03 — Architectural Principles

> **Audience:** Platform Engineers, Architects · **Phase:** 1

- 3.1 Separation of Concerns: Business vs. Presentation
- 3.2 The AST-First Principle
- 3.3 Schema-Driven Everything
- 3.4 Progressive Enhancement
- 3.5 Backward Compatibility as a Contract
- 3.6 Fail-Safe Degradation
- 3.7 Least-Privilege UI Rendering
- 3.8 Composability Over Configuration
- 3.9 Immutability of Compiled Definitions
- 3.10 Audit-Ability of UI Decisions
- 3.11 Layered Architecture Overview
  - 3.11.1 Layer 1 — Business Services
  - 3.11.2 Layer 2 — UI Compilation Layer
  - 3.11.3 Layer 3 — Transport (JSON over REST/gRPC)
  - 3.11.4 Layer 4 — Client Rendering Engine
  - 3.11.5 Layer 5 — Client Interaction & State
- 3.12 Cross-Cutting Concerns
  - 3.12.1 Security
  - 3.12.2 Observability
  - 3.12.3 Tenancy
  - 3.12.4 Localization

---

### Chapter 04 — System Overview

> **Audience:** All · **Phase:** 1

- 4.1 High-Level System Diagram
- 4.2 Platform Components Inventory
  - 4.2.1 UI Compilation Service
  - 4.2.2 Component Registry
  - 4.2.3 Schema Validation Service
  - 4.2.4 Feature Flag Resolver
  - 4.2.5 Authorization Resolver
  - 4.2.6 Localization Service
  - 4.2.7 Web Rendering Engine
  - 4.2.8 Mobile Rendering Engine
- 4.3 Request Lifecycle: From Business Event to Rendered UI
  - 4.3.1 Step 1 — Client Requests a UI Surface
  - 4.3.2 Step 2 — Backend Resolves Context (Tenant, User, Flags)
  - 4.3.3 Step 3 — Business Service Emits UI AST
  - 4.3.4 Step 4 — Compiler Validates and Serializes
  - 4.3.5 Step 5 — JSON Delivered to Client
  - 4.3.6 Step 6 — Client Renders Definitively
  - 4.3.7 Step 7 — Client Submits Actions / Events
- 4.4 Technology Stack Mapping
  - 4.4.1 Go + Goa for Service Contracts
  - 4.4.2 PostgreSQL for Persistent UI Definitions
  - 4.4.3 Redis for UI Definition Caching
  - 4.4.4 Temporal for Long-Running Workflow UI
  - 4.4.5 OpenFGA/Casbin for Authorization-Aware UI
  - 4.4.6 Feature Flags for Tenant Variations
- 4.5 Deployment Topology
- 4.6 Scalability Characteristics
- 4.7 Known Limitations

---

## Volume II — DSL & AST

*Defines the language the platform speaks. The formal contract between backend and all rendering clients.*

---

### Chapter 05 — Server-Driven UI Fundamentals

> **Audience:** All Engineers · **Phase:** 1

- 5.1 What Is Server-Driven UI (SDUI)?
- 5.2 SDUI Spectrum: Thin Templates to Full AST
- 5.3 Why AwoERP Chose Full AST Over Thin Templates
- 5.4 The Role of JSON in SDUI
- 5.5 SDUI vs. Traditional Frontend Rendering
  - 5.5.1 Comparison: React/Vue Component Trees vs. SDUI JSON
  - 5.5.2 Comparison: Native Mobile vs. SDUI JSON
- 5.6 Client Contract Guarantees
  - 5.6.1 What the Client Must Render
  - 5.6.2 What the Client May Enhance
  - 5.6.3 What the Client Must Not Override
- 5.7 Schema Versioning and Client Compatibility
- 5.8 SDUI Security Boundaries
- 5.9 SDUI Debugging Mental Model

---

### Chapter 06 — UI DSL Architecture

> **Audience:** Platform Engineers, UI Framework Developers · **Phase:** 1

- 6.1 What Is the AwoERP UI DSL?
- 6.2 DSL Design Goals
- 6.3 DSL Grammar Overview
  - 6.3.1 Nodes
  - 6.3.2 Edges (Parent-Child Relationships)
  - 6.3.3 Slots
  - 6.3.4 Bindings
  - 6.3.5 Directives
- 6.4 Type System for the DSL
  - 6.4.1 Primitive Types
  - 6.4.2 Composite Types
  - 6.4.3 Reference Types (Component Refs, Data Refs)
  - 6.4.4 Conditional Types
  - 6.4.5 Expression Types
- 6.5 Binding Expressions
  - 6.5.1 Static Bindings
  - 6.5.2 Data Bindings (DataSource References)
  - 6.5.3 Conditional Bindings
  - 6.5.4 Computed Bindings
  - 6.5.5 Permission Bindings
- 6.6 Directives
  - 6.6.1 `v-if` / `v-show` Equivalents
  - 6.6.2 `v-for` / Repeat Directives
  - 6.6.3 Permission Directives
  - 6.6.4 Feature Flag Directives
  - 6.6.5 Locale Directives
- 6.7 DSL Extensibility — Defining Custom Node Types
- 6.8 DSL Validation Rules
- 6.9 DSL Versioning

---

### Chapter 07 — AST Design

> **Audience:** Platform Engineers, Backend Engineers · **Phase:** 1

- 7.1 What Is the UI AST?
- 7.2 AST Node Taxonomy
  - 7.2.1 Container Nodes
  - 7.2.2 Leaf Nodes (Atoms)
  - 7.2.3 Control Flow Nodes
  - 7.2.4 Data Nodes
  - 7.2.5 Action Nodes
  - 7.2.6 Layout Nodes
  - 7.2.7 Composite Nodes (Templates)
- 7.3 Node Schema — Universal Fields
  - 7.3.1 `id` — Unique Node Identity
  - 7.3.2 `type` — Node Type Discriminator
  - 7.3.3 `props` — Node Properties
  - 7.3.4 `children` — Child Node Array
  - 7.3.5 `slots` — Named Child Slots
  - 7.3.6 `events` — Event Handler Bindings
  - 7.3.7 `actions` — Action Bindings
  - 7.3.8 `permissions` — Authorization Metadata
  - 7.3.9 `flags` — Feature Flag Conditions
  - 7.3.10 `metadata` — Debug / Tracing Fields
- 7.4 AST Construction in Go
  - 7.4.1 Builder Pattern
  - 7.4.2 Typed Node Factories
  - 7.4.3 Slot Composition
  - 7.4.4 Recursive AST Construction
- 7.5 AST Validation
  - 7.5.1 Structural Validation
  - 7.5.2 Type Validation
  - 7.5.3 Reference Integrity Validation
  - 7.5.4 Permission Consistency Validation
  - 7.5.5 Cyclic Reference Detection
- 7.6 AST Transformation Passes
  - 7.6.1 Permission Pruning Pass
  - 7.6.2 Feature Flag Resolution Pass
  - 7.6.3 Localization Injection Pass
  - 7.6.4 Default Value Injection Pass
  - 7.6.5 Tenant Override Application Pass
- 7.7 AST Diffing and Patching
- 7.8 AST Serialization Formats
- 7.9 AST Canonical Form

---

### Chapter 08 — JSON Compilation Pipeline

> **Audience:** Platform Engineers, Backend Engineers · **Phase:** 1

- 8.1 Pipeline Overview and Stages
- 8.2 Stage 1 — Context Resolution
  - 8.2.1 Tenant Resolution
  - 8.2.2 User Identity & Role Resolution
  - 8.2.3 Feature Flag Resolution
  - 8.2.4 Locale Resolution
  - 8.2.5 Device / Client Context Resolution
- 8.3 Stage 2 — AST Construction
  - 8.3.1 Business Service Emits Raw AST
  - 8.3.2 Slot Composition
  - 8.3.3 Template Instantiation
- 8.4 Stage 3 — AST Transformation
  - 8.4.1 Permission Pruning
  - 8.4.2 Feature Flag Evaluation
  - 8.4.3 Tenant Overrides Application
  - 8.4.4 Localization String Injection
  - 8.4.5 Default Values and Fallbacks
- 8.5 Stage 4 — Validation
  - 8.5.1 Schema Validation
  - 8.5.2 Business Rule Validation
  - 8.5.3 Security Assertion Validation
- 8.6 Stage 5 — Serialization
  - 8.6.1 JSON Marshaling
  - 8.6.2 Compression Options
  - 8.6.3 Payload Size Budgets
  - 8.6.4 Field Stripping / Projection
- 8.7 Stage 6 — Transport
  - 8.7.1 REST Response Wrapping
  - 8.7.2 gRPC Streaming of UI Definitions
  - 8.7.3 Cache Header Strategy
- 8.8 Pipeline Error Handling
- 8.9 Pipeline Observability
  - 8.9.1 Span Tracing per Stage
  - 8.9.2 Stage Latency Metrics
  - 8.9.3 Payload Size Metrics
- 8.10 Pipeline Extension Points

---

## Volume III — Component System

*The vocabulary of the platform. All renderable primitives, composites, and ERP-domain-specific components.*

---

### Chapter 09 — Component System

> **Audience:** Platform Engineers, Web/Mobile Engineers, UI Framework Developers · **Phase:** 2

- 9.1 Component Philosophy
- 9.2 Component Identity and Registry
  - 9.2.1 Component Type Identifiers (Namespaced)
  - 9.2.2 The Component Registry Service
  - 9.2.3 Registry Lookup and Resolution
  - 9.2.4 Custom Component Registration
- 9.3 Component Category Taxonomy
  - 9.3.1 Primitive / Atomic Components
  - 9.3.2 Compound Components
  - 9.3.3 Layout Components
  - 9.3.4 Data Components
  - 9.3.5 Form Components
  - 9.3.6 Navigation Components
  - 9.3.7 Feedback & Status Components
  - 9.3.8 ERP Domain Components
  - 9.3.9 Chart / Analytics Components
  - 9.3.10 Workflow Components
- 9.4 Component Contract
  - 9.4.1 Props Schema (Required vs. Optional)
  - 9.4.2 Slot Definitions
  - 9.4.3 Emitted Events
  - 9.4.4 Supported Actions
  - 9.4.5 Permission Surface
  - 9.4.6 Rendering Constraints
- 9.5 Component Versioning
  - 9.5.1 Component Version Lifecycle
  - 9.5.2 Breaking vs. Non-Breaking Component Changes
  - 9.5.3 Deprecation Strategy
- 9.6 Primitive Component Library
  - 9.6.1 Text, Label, Heading
  - 9.6.2 Button, IconButton, FAB
  - 9.6.3 Icon
  - 9.6.4 Image
  - 9.6.5 Divider, Spacer
  - 9.6.6 Badge, Tag, Chip
  - 9.6.7 Avatar
  - 9.6.8 Spinner / Loader
  - 9.6.9 Progress Bar
  - 9.6.10 Tooltip, Popover
- 9.7 Compound Component Library
  - 9.7.1 Card, List Item
  - 9.7.2 Alert, Banner, Notification
  - 9.7.3 Modal / Dialog
  - 9.7.4 Drawer / Side Panel
  - 9.7.5 Stepper / Wizard
  - 9.7.6 Accordion / Collapsible
  - 9.7.7 Tab Container
  - 9.7.8 Breadcrumb
- 9.8 Theming and Style Props
  - 9.8.1 Design Token References
  - 9.8.2 Tenant-Level Theming
  - 9.8.3 Component Style Overrides
- 9.9 Conditional Rendering on Components
- 9.10 Component Rendering Fallbacks

---

### Chapter 10 — Layout System

> **Audience:** Web/Mobile Engineers, UI Framework Developers · **Phase:** 2

- 10.1 Layout Philosophy
- 10.2 Layout Model Overview
  - 10.2.1 Flow Layout
  - 10.2.2 Grid Layout
  - 10.2.3 Flex Layout
  - 10.2.4 Stack Layout (Vertical / Horizontal)
  - 10.2.5 Absolute/Overlay Layout
  - 10.2.6 Responsive Layout
- 10.3 Grid System
  - 10.3.1 Column Definitions
  - 10.3.2 Row Definitions
  - 10.3.3 Span and Offset
  - 10.3.4 Gap and Gutter
  - 10.3.5 Nested Grids
- 10.4 Responsive Breakpoints
  - 10.4.1 Breakpoint Definitions (xs, sm, md, lg, xl)
  - 10.4.2 Per-Breakpoint Visibility
  - 10.4.3 Per-Breakpoint Column Spans
  - 10.4.4 Breakpoint Overrides from Backend
- 10.5 Layout Containers
  - 10.5.1 Page Container
  - 10.5.2 Section Container
  - 10.5.3 Panel Container
  - 10.5.4 Split Pane
  - 10.5.5 Scrollable Container
  - 10.5.6 Sticky Header / Footer
- 10.6 Spacing System
  - 10.6.1 Margin and Padding Tokens
  - 10.6.2 Density Modes (Compact, Normal, Comfortable)
- 10.7 Layout for Mobile
  - 10.7.1 Single Column Default
  - 10.7.2 Safe Area Insets
  - 10.7.3 Bottom Sheet Layouts
- 10.8 Layout for Web
  - 10.8.1 Sidebar + Content Shell
  - 10.8.2 Master-Detail Layout
  - 10.8.3 Embedded / Portlet Layouts
- 10.9 Layout Serialization in AST

---

### Chapter 11 — Forms Framework

> **Audience:** Backend Engineers, Web/Mobile Engineers · **Phase:** 2

- 11.1 Forms Philosophy in SDUI
- 11.2 Form Node Types
  - 11.2.1 Form Container
  - 11.2.2 Form Section / Group
  - 11.2.3 Form Field
  - 11.2.4 Field Label, Help Text, Error Text
- 11.3 Field Input Types
  - 11.3.1 Text Input (Single Line, Multi-Line)
  - 11.3.2 Number Input (Integer, Decimal, Currency)
  - 11.3.3 Date / Time / DateTime Pickers
  - 11.3.4 Date Range Picker
  - 11.3.5 Select / Dropdown (Static, Dynamic)
  - 11.3.6 Multi-Select
  - 11.3.7 Checkbox, Checkbox Group
  - 11.3.8 Radio Group
  - 11.3.9 Toggle / Switch
  - 11.3.10 File Upload
  - 11.3.11 Image Upload
  - 11.3.12 Signature Input
  - 11.3.13 Rich Text Editor
  - 11.3.14 Code Editor (for formula/expression fields)
  - 11.3.15 Rating / Star Input
  - 11.3.16 Slider
  - 11.3.17 Color Picker
  - 11.3.18 Address / Location Input
  - 11.3.19 Phone Number Input
  - 11.3.20 Currency Input (Multi-Currency)
  - 11.3.21 Lookup / Search (Entity Picker)
  - 11.3.22 Dependent / Cascading Selects
- 11.4 Form Layout
  - 11.4.1 Field Grid Placement
  - 11.4.2 Multi-Column Forms
  - 11.4.3 Collapsible Sections
  - 11.4.4 Tab-Based Form Sections
  - 11.4.5 Wizard / Multi-Step Forms
- 11.5 Dynamic Forms
  - 11.5.1 Conditionally Visible Fields
  - 11.5.2 Conditionally Required Fields
  - 11.5.3 Dynamically Populated Options
  - 11.5.4 Repeating Field Groups (Line Items)
  - 11.5.5 Computed / Read-Only Fields
- 11.6 Form State Management
  - 11.6.1 Initial Values
  - 11.6.2 Dirty State
  - 11.6.3 Touched State
  - 11.6.4 Error State
  - 11.6.5 Submitting State
- 11.7 Form Validation (See Chapter 23)
- 11.8 Form Actions
  - 11.8.1 Submit Action
  - 11.8.2 Save Draft Action
  - 11.8.3 Reset Action
  - 11.8.4 Cancel Action
  - 11.8.5 Custom Form Actions
- 11.9 Form Permissions
  - 11.9.1 Field-Level Read/Write Permissions
  - 11.9.2 Section-Level Visibility
  - 11.9.3 Submit Permission
- 11.10 ERP-Specific Form Patterns
  - 11.10.1 Header + Line Items Pattern
  - 11.10.2 Approval Routing Fields
  - 11.10.3 Document Attachment Fields
  - 11.10.4 Audit Trail Display in Forms

---

### Chapter 12 — Tables and Data Grids

> **Audience:** Backend Engineers, Web/Mobile Engineers · **Phase:** 2

- 12.1 Table vs. Data Grid Philosophy
- 12.2 Table Node Schema
  - 12.2.1 Column Definitions
  - 12.2.2 Row Data Source Binding
  - 12.2.3 Key Field Configuration
- 12.3 Column Types
  - 12.3.1 Text Column
  - 12.3.2 Number / Currency Column
  - 12.3.3 Date / DateTime Column
  - 12.3.4 Boolean / Status Column
  - 12.3.5 Link Column
  - 12.3.6 Action Column (Buttons / Icons)
  - 12.3.7 Badge / Tag Column
  - 12.3.8 Image / Avatar Column
  - 12.3.9 Progress Column
  - 12.3.10 Custom Render Column
- 12.4 Table Features
  - 12.4.1 Sorting (Client vs. Server)
  - 12.4.2 Filtering (Column-Level, Global)
  - 12.4.3 Pagination (Offset, Cursor, Infinite Scroll)
  - 12.4.4 Row Selection (Single, Multi)
  - 12.4.5 Row Expansion / Details Panel
  - 12.4.6 Frozen Columns
  - 12.4.7 Column Reordering (Persisted per User)
  - 12.4.8 Column Resizing
  - 12.4.9 Column Visibility Toggle
  - 12.4.10 Row Grouping
  - 12.4.11 Aggregation Rows (Sum, Average, Count)
  - 12.4.12 Inline Editing
  - 12.4.13 Row-Level Actions
  - 12.4.14 Bulk Actions on Selected Rows
  - 12.4.15 Export (CSV, Excel, PDF)
- 12.5 Table Permissions
  - 12.5.1 Column-Level Visibility
  - 12.5.2 Row-Level Visibility (FGA-Driven)
  - 12.5.3 Action Column Permissions
- 12.6 ERP-Specific Table Patterns
  - 12.6.1 Line Item Grid
  - 12.6.2 Ledger / Journal Table
  - 12.6.3 Audit Log Table
  - 12.6.4 Hierarchical / Tree Table
- 12.7 Table Performance Considerations
- 12.8 Mobile Table Adaptations

---

### Chapter 13 — Dashboard Framework

> **Audience:** Backend Engineers, Web/Mobile Engineers · **Phase:** 2

- 13.1 Dashboard Design Goals
- 13.2 Dashboard Node Schema
  - 13.2.1 Dashboard Container
  - 13.2.2 Widget Container
  - 13.2.3 Widget Manifest
- 13.3 Dashboard Layouts
  - 13.3.1 Fixed Grid Dashboard
  - 13.3.2 Drag-and-Drop Grid (User-Configurable)
  - 13.3.3 Responsive Dashboard
  - 13.3.4 Mobile Dashboard Layout
- 13.4 Widget Types
  - 13.4.1 KPI / Metric Widget
  - 13.4.2 Trend Card
  - 13.4.3 Chart Widget (References Chapter 14)
  - 13.4.4 Table Widget (References Chapter 12)
  - 13.4.5 Activity Feed Widget
  - 13.4.6 Notification Widget
  - 13.4.7 Quick Action Widget
  - 13.4.8 Approval Pending Widget
  - 13.4.9 Embedded Report Widget
  - 13.4.10 Map / Geo Widget
  - 13.4.11 Calendar / Schedule Widget
  - 13.4.12 Custom Iframe Widget (Gated)
- 13.5 Dashboard Data Sources
- 13.6 Dashboard Personalization
  - 13.6.1 Role-Based Default Layouts
  - 13.6.2 User Customization Persistence
  - 13.6.3 Tenant Dashboard Templates
- 13.7 Dashboard Permissions
- 13.8 Dashboard Refresh and Polling
- 13.9 Dashboard Export

---

### Chapter 14 — Charts and Analytics

> **Audience:** Backend Engineers, Web/Mobile Engineers · **Phase:** 2

- 14.1 Chart System Overview
- 14.2 Chart Node Schema
  - 14.2.1 Chart Container
  - 14.2.2 Data Series Bindings
  - 14.2.3 Axis Configuration
  - 14.2.4 Legend Configuration
  - 14.2.5 Tooltip Configuration
- 14.3 Chart Types
  - 14.3.1 Bar Chart (Vertical, Horizontal, Stacked, Grouped)
  - 14.3.2 Line Chart (Single, Multi-Series, Area)
  - 14.3.3 Pie / Donut Chart
  - 14.3.4 Scatter Plot
  - 14.3.5 Bubble Chart
  - 14.3.6 Funnel Chart
  - 14.3.7 Gauge / Speedometer
  - 14.3.8 Heatmap
  - 14.3.9 Waterfall Chart
  - 14.3.10 Gantt / Timeline Chart
  - 14.3.11 Treemap
  - 14.3.12 Sankey Diagram
  - 14.3.13 Combo Charts (Bar + Line)
- 14.4 Chart Data Binding
  - 14.4.1 Static Data
  - 14.4.2 DataSource Reference
  - 14.4.3 Aggregation Expressions
  - 14.4.4 Real-Time Data Subscriptions
- 14.5 Chart Interactivity
  - 14.5.1 Click Events → Actions
  - 14.5.2 Drill-Down Navigation
  - 14.5.3 Cross-Filter Between Charts
- 14.6 Chart Theming and Branding
- 14.7 Chart Export (PNG, SVG, PDF)
- 14.8 Chart Performance for Large Datasets
- 14.9 Chart Accessibility

---

### Chapter 15 — Workflow and Approval Components

> **Audience:** Backend Engineers, Mobile/Web Engineers, ERP Implementers · **Phase:** 2

- 15.1 Workflow UI Philosophy (Temporal Integration)
- 15.2 Workflow Component Taxonomy
  - 15.2.1 Workflow Status Indicator
  - 15.2.2 Workflow Progress Stepper
  - 15.2.3 Approval Action Bar
  - 15.2.4 Approval History Timeline
  - 15.2.5 Delegation Panel
  - 15.2.6 Escalation Indicator
  - 15.2.7 Comments / Notes Thread
  - 15.2.8 Attachment Viewer
  - 15.2.9 Document Comparison Panel
- 15.3 Approval Actions
  - 15.3.1 Approve
  - 15.3.2 Reject (with Reason)
  - 15.3.3 Request Revision
  - 15.3.4 Delegate / Re-route
  - 15.3.5 Escalate
  - 15.3.6 Withdraw (Submitter Action)
- 15.4 Workflow State Machine Representation in UI
- 15.5 Temporal Signal Binding
- 15.6 Push Notification Integration
- 15.7 Pending Approval Lists and Queues
- 15.8 SLA Timers and Deadline Indicators
- 15.9 Audit Trail Components

---

### Chapter 16 — ERP-Specific Components

> **Audience:** Backend Engineers, ERP Implementers · **Phase:** 2

- 16.1 ERP Component Design Principles
- 16.2 Financial Components
  - 16.2.1 Currency Display (Multi-Currency, FX Rate)
  - 16.2.2 Amount Breakdown Panel
  - 16.2.3 Tax Summary Panel
  - 16.2.4 Ledger Entry Row
  - 16.2.5 Journal Voucher Viewer
  - 16.2.6 Trial Balance Table
  - 16.2.7 Profit & Loss Summary Card
  - 16.2.8 Budget vs. Actual Bar
- 16.3 Procurement Components
  - 16.3.1 Purchase Order Header
  - 16.3.2 Line Item Grid (PO)
  - 16.3.3 Vendor Card
  - 16.3.4 RFQ / Quotation Comparison Table
  - 16.3.5 GRN (Goods Received Note) Viewer
- 16.4 Inventory Components
  - 16.4.1 Stock Level Indicator
  - 16.4.2 Bin / Location Selector
  - 16.4.3 Batch / Serial Number Panel
  - 16.4.4 Stock Movement Timeline
- 16.5 HR Components
  - 16.5.1 Employee Card
  - 16.5.2 Org Chart
  - 16.5.3 Leave Balance Panel
  - 16.5.4 Payslip Viewer
  - 16.5.5 Attendance Heatmap
- 16.6 Sales Components
  - 16.6.1 Customer Card
  - 16.6.2 Sales Order Header
  - 16.6.3 Pipeline Stage Indicator
  - 16.6.4 Invoice Viewer
  - 16.6.5 Credit Limit Warning
- 16.7 Document Components
  - 16.7.1 PDF Viewer (Embedded)
  - 16.7.2 Document Status Badge
  - 16.7.3 E-Signature Capture
  - 16.7.4 QR Code Display
  - 16.7.5 Barcode Scanner Input

---

## Volume IV — Rendering Architecture

*How the JSON payload is brought to life on each platform.*

---

### Chapter 17 — Mobile Rendering Architecture

> **Audience:** Mobile Engineers · **Phase:** 2

- 17.1 Mobile Rendering Philosophy
- 17.2 Supported Mobile Platforms
  - 17.2.1 iOS (Swift / SwiftUI)
  - 17.2.2 Android (Kotlin / Jetpack Compose)
  - 17.2.3 Cross-Platform (React Native / Flutter) Considerations
- 17.3 Mobile Rendering Engine Architecture
  - 17.3.1 JSON Parser and AST Deserializer
  - 17.3.2 Component Resolver
  - 17.3.3 Rendering Tree Builder
  - 17.3.4 View Reconciler
  - 17.3.5 Action Dispatcher
  - 17.3.6 Event Bus
- 17.4 Component-to-Native Mapping
  - 17.4.1 Mapping Table: UI DSL Type → Native View
  - 17.4.2 Unsupported Component Fallback Strategy
  - 17.4.3 Custom Native Component Bridges
- 17.5 Layout Rendering on Mobile
  - 17.5.1 Flex Layout to Native Stack/HStack/VStack
  - 17.5.2 Grid on Mobile
  - 17.5.3 Safe Area Handling
- 17.6 Form Rendering on Mobile
  - 17.6.1 Keyboard Avoidance
  - 17.6.2 Native Input Pickers (Date, Location)
  - 17.6.3 File and Camera Inputs
- 17.7 Navigation Rendering on Mobile (See Chapter 19)
- 17.8 Mobile State Management
- 17.9 Mobile Performance
  - 17.9.1 Lazy Loading of Nested Components
  - 17.9.2 Image Caching
  - 17.9.3 List Virtualization
- 17.10 Mobile Offline Support (See Chapter 33)
- 17.11 Mobile Accessibility (See Chapter 37)
- 17.12 Mobile Telemetry (See Chapter 39)
- 17.13 Testing Mobile Rendering (See Chapter 41)

---

### Chapter 18 — Web Rendering Architecture

> **Audience:** Web Engineers · **Phase:** 2

- 18.1 Web Rendering Philosophy
- 18.2 Web Technology Choices
  - 18.2.1 Framework Selection Rationale (React, Vue, etc.)
  - 18.2.2 Server-Side Rendering (SSR) Considerations
  - 18.2.3 Static vs. Dynamic Routes
- 18.3 Web Rendering Engine Architecture
  - 18.3.1 JSON Fetcher and Schema Validator
  - 18.3.2 Component Registry (Web)
  - 18.3.3 Virtual DOM Tree Builder from AST
  - 18.3.4 Reactive State Binding
  - 18.3.5 Action Dispatcher (Web)
  - 18.3.6 Event System (Web)
- 18.4 Component-to-HTML/DOM Mapping
  - 18.4.1 Mapping Table: UI DSL Type → Web Component
  - 18.4.2 Unknown Type Fallback Strategy
  - 18.4.3 Custom Web Component Registration
- 18.5 Layout Rendering on Web
  - 18.5.1 CSS Grid Implementation
  - 18.5.2 Responsive Breakpoint Handling
  - 18.5.3 Dark Mode Support
- 18.6 Form Rendering on Web
  - 18.6.1 Controlled Input Pattern
  - 18.6.2 Form Library Integration
  - 18.6.3 File Upload Handling
- 18.7 Navigation Rendering on Web (See Chapter 19)
- 18.8 Web State Management (See Chapter 22)
- 18.9 Web Performance
  - 18.9.1 Component Code Splitting
  - 18.9.2 Memoization of Static Subtrees
  - 18.9.3 Virtual Scrolling for Large Lists
  - 18.9.4 Progressive Hydration
- 18.10 Web Offline Support (See Chapter 33)
- 18.11 Web Accessibility (See Chapter 37)
- 18.12 Web Telemetry (See Chapter 39)
- 18.13 Testing Web Rendering (See Chapter 41)

---

### Chapter 19 — Navigation Framework

> **Audience:** Backend Engineers, Mobile/Web Engineers · **Phase:** 2

- 19.1 Navigation Philosophy
- 19.2 Navigation Node Types
  - 19.2.1 App Shell / Root Navigator
  - 19.2.2 Tab Bar Navigator
  - 19.2.3 Side Navigation (Sidebar/Drawer)
  - 19.2.4 Stack Navigator (Push/Pop)
  - 19.2.5 Modal Navigator
  - 19.2.6 Bottom Sheet Navigator
  - 19.2.7 Breadcrumb Trail
- 19.3 Navigation Item Schema
  - 19.3.1 Label, Icon, Badge
  - 19.3.2 Target Route
  - 19.3.3 Permission Guard
  - 19.3.4 Feature Flag Guard
  - 19.3.5 Active State Conditions
  - 19.3.6 Children (Sub-Navigation)
- 19.4 Route System
  - 19.4.1 Route Definitions
  - 19.4.2 Route Parameters
  - 19.4.3 Query Parameters
  - 19.4.4 Deep Linking
  - 19.4.5 Dynamic Route Generation from Backend
- 19.5 Navigation Actions
  - 19.5.1 Push
  - 19.5.2 Pop / Back
  - 19.5.3 Replace
  - 19.5.4 Reset Stack
  - 19.5.5 Open Modal
  - 19.5.6 Close Modal
  - 19.5.7 Open Drawer
  - 19.5.8 Navigate to External URL
- 19.6 Permission-Aware Navigation
  - 19.6.1 Menu Item Visibility
  - 19.6.2 Route Access Guards
  - 19.6.3 Redirect on Unauthorized
- 19.7 Navigation for Multi-Tenant
  - 19.7.1 Tenant-Specific Menu Configurations
  - 19.7.2 Module-Based Navigation Trees
- 19.8 Navigation State Persistence
- 19.9 Platform-Specific Navigation Patterns
  - 19.9.1 iOS Navigation Conventions
  - 19.9.2 Android Navigation Conventions
  - 19.9.3 Web Router Integration

---

### Chapter 20 — Action System

> **Audience:** Platform Engineers, Backend/Web/Mobile Engineers · **Phase:** 2

- 20.1 Action System Philosophy
- 20.2 Action Node Schema
  - 20.2.1 Action Type Discriminator
  - 20.2.2 Action Parameters
  - 20.2.3 Action Conditions (Pre-Conditions)
  - 20.2.4 Action Confirmation (Dialogs)
  - 20.2.5 Action Permissions
  - 20.2.6 Action Feedback (Loading, Success, Error)
- 20.3 Built-In Action Types
  - 20.3.1 HTTP Request Action (REST API Call)
  - 20.3.2 Navigate Action
  - 20.3.3 Open Modal Action
  - 20.3.4 Close Modal Action
  - 20.3.5 Open Drawer Action
  - 20.3.6 Submit Form Action
  - 20.3.7 Reset Form Action
  - 20.3.8 Refresh Data Source Action
  - 20.3.9 Set Variable Action
  - 20.3.10 Emit Event Action
  - 20.3.11 Show Notification Action
  - 20.3.12 Download File Action
  - 20.3.13 Copy to Clipboard Action
  - 20.3.14 Open External Link Action
  - 20.3.15 Workflow Signal Action (Temporal)
  - 20.3.16 Print Action
  - 20.3.17 Custom Action (Plugin Hook)
- 20.4 Action Chaining
  - 20.4.1 Sequential Chains
  - 20.4.2 Conditional Chains
  - 20.4.3 Parallel Chains
  - 20.4.4 Error Branches
- 20.5 Optimistic UI Actions
- 20.6 Action Debounce and Throttle
- 20.7 Action Audit Trail
- 20.8 Action Permissions Enforcement (Client + Server)

---

### Chapter 21 — Event System

> **Audience:** Platform Engineers, Backend/Web/Mobile Engineers · **Phase:** 2

- 21.1 Event System Overview
- 21.2 Event Types
  - 21.2.1 Component Lifecycle Events
  - 21.2.2 User Interaction Events
  - 21.2.3 Data Events (Load, Update, Error)
  - 21.2.4 Navigation Events
  - 21.2.5 Form Events
  - 21.2.6 Workflow Events (Temporal Signals)
  - 21.2.7 System Events (Auth, Session)
  - 21.2.8 Custom Domain Events
- 21.3 Event Node Schema
  - 21.3.1 Event Name
  - 21.3.2 Event Payload Schema
  - 21.3.3 Handler Action(s)
  - 21.3.4 Event Propagation Settings
- 21.4 Event Bus Architecture
  - 21.4.1 Client-Side Event Bus
  - 21.4.2 Server-Sent Events (SSE) for Real-Time
  - 21.4.3 WebSocket Event Channels
  - 21.4.4 gRPC Server Streaming
- 21.5 Event Filtering and Routing
- 21.6 Cross-Component Event Communication
- 21.7 Event Logging and Replay

---

## Volume V — Runtime & State

---

### Chapter 22 — State Management

> **Audience:** Web/Mobile Engineers · **Phase:** 3

- 22.1 State Management Philosophy in SDUI
- 22.2 State Taxonomy
  - 22.2.1 Page State (Ephemeral)
  - 22.2.2 Form State
  - 22.2.3 Data Cache State
  - 22.2.4 Navigation State
  - 22.2.5 Global Application State
  - 22.2.6 User Preference State (Persisted)
  - 22.2.7 Workflow State (Server-Owned)
- 22.3 State Variable System
  - 22.3.1 Declaring Variables in DSL
  - 22.3.2 Scoped Variables (Page, Component, Global)
  - 22.3.3 Variable Bindings in Props
  - 22.3.4 Variable Mutation via Actions
- 22.4 Derived / Computed State
- 22.5 State Synchronization with Server
- 22.6 State Persistence (Local Storage, Keychain)
- 22.7 State Debugging Tools

---

### Chapter 23 — Validation Framework

> **Audience:** Backend Engineers, Web/Mobile Engineers · **Phase:** 3

- 23.1 Validation Philosophy (Server-First, Client-Assisted)
- 23.2 Validation Rule Types
  - 23.2.1 Required
  - 23.2.2 MinLength / MaxLength
  - 23.2.3 Min / Max (Numeric)
  - 23.2.4 Pattern (Regex)
  - 23.2.5 Email, URL, Phone
  - 23.2.6 Date Range
  - 23.2.7 Custom Expression Validation
  - 23.2.8 Cross-Field Validation
  - 23.2.9 Async Remote Validation (API Call)
  - 23.2.10 Business Rule Validation (Server-Side Only)
- 23.3 Validation Rule Schema in AST
- 23.4 Validation Execution
  - 23.4.1 On-Change Validation
  - 23.4.2 On-Blur Validation
  - 23.4.3 On-Submit Validation
  - 23.4.4 Server-Side Validation Response Mapping
- 23.5 Validation Error Display
  - 23.5.1 Field-Level Errors
  - 23.5.2 Section-Level Errors
  - 23.5.3 Form-Level Errors (Summary)
- 23.6 Validation and Multi-Step Forms
- 23.7 Validation Internationalization

---

### Chapter 24 — Data Sources

> **Audience:** Backend Engineers, Web/Mobile Engineers · **Phase:** 3

- 24.1 Data Source Philosophy
- 24.2 Data Source Node Schema
  - 24.2.1 Data Source ID
  - 24.2.2 Source Type Discriminator
  - 24.2.3 Query Parameters
  - 24.2.4 Transformation / Mapping
  - 24.2.5 Pagination Config
  - 24.2.6 Refresh Config
  - 24.2.7 Cache Config
- 24.3 Data Source Types
  - 24.3.1 REST API Data Source
  - 24.3.2 gRPC Data Source
  - 24.3.3 Static / Inline Data Source
  - 24.3.4 Computed Data Source (Derived from Others)
  - 24.3.5 Real-Time (WebSocket / SSE) Data Source
  - 24.3.6 File / Blob Data Source
- 24.4 Data Transformation Layer
  - 24.4.1 Field Mapping
  - 24.4.2 Type Coercion
  - 24.4.3 Aggregation Expressions
  - 24.4.4 Filtering Expressions
  - 24.4.5 Sorting Expressions
- 24.5 Data Source Lifecycle
  - 24.5.1 Lazy vs. Eager Loading
  - 24.5.2 Dependent Data Sources
  - 24.5.3 Data Source Invalidation
- 24.6 Data Source Error Handling
- 24.7 Data Source Security
  - 24.7.1 Auth Token Forwarding
  - 24.7.2 Tenant Data Isolation
  - 24.7.3 Field-Level Data Masking

---

## Volume VI — Platform & API

---

### Chapter 25 — API Contracts

> **Audience:** Backend Engineers, Third-Party Integrators · **Phase:** 3

- 25.1 API Design Philosophy
- 25.2 REST API Contract
  - 25.2.1 UI Definition Endpoint (`GET /ui/{surface}`)
  - 25.2.2 UI Definition with Context (`POST /ui/{surface}/resolve`)
  - 25.2.3 Component Registry Endpoint
  - 25.2.4 Action Execution Endpoint
  - 25.2.5 Event Submission Endpoint
  - 25.2.6 User Preference Endpoints
  - 25.2.7 Tenant Configuration Endpoints
- 25.3 gRPC API Contract
  - 25.3.1 `UIService` Proto Definition
  - 25.3.2 Streaming UI Definitions
  - 25.3.3 Real-Time Event Channel
- 25.4 Request / Response Envelope Schema
  - 25.4.1 Pagination Envelope
  - 25.4.2 Error Envelope
  - 25.4.3 Metadata Headers
- 25.5 API Versioning Strategy (See Chapter 43)
- 25.6 API Authentication
- 25.7 API Rate Limiting
- 25.8 API SDK Generation from Goa

---

### Chapter 26 — Security Model

> **Audience:** Platform Engineers, Security Engineers · **Phase:** 3

- 26.1 Security Threat Model for SDUI
- 26.2 Authentication in UI Flows
  - 26.2.1 JWT / OAuth2 Token Validation
  - 26.2.2 Session Management
  - 26.2.3 Token Refresh in Long-Running Sessions
- 26.3 UI Definition Security
  - 26.3.1 Never Trust Client-Provided UI
  - 26.3.2 Server-Side Permission Pruning (Not Client-Side Hiding)
  - 26.3.3 Sensitive Field Masking in JSON Output
- 26.4 Action Security
  - 26.4.1 Action Signature Verification
  - 26.4.2 CSRF Protection for Action Endpoints
  - 26.4.3 Idempotency Keys
- 26.5 Transport Security
  - 26.5.1 TLS Everywhere
  - 26.5.2 Certificate Pinning (Mobile)
  - 26.5.3 HTTP Security Headers (Web)
- 26.6 Data Security
  - 26.6.1 PII Handling in UI Payloads
  - 26.6.2 Data Retention in Client State
  - 26.6.3 Secure Logging Practices
- 26.7 Supply Chain Security (Third-Party Components)
- 26.8 Security Audit Checklist

---

### Chapter 27 — Authorization Integration

> **Audience:** Platform Engineers, Backend Engineers · **Phase:** 3

- 27.1 Authorization Philosophy (OpenFGA / Casbin)
- 27.2 Authorization Check Points in the Pipeline
  - 27.2.1 Menu / Navigation Authorization
  - 27.2.2 Page-Level Authorization
  - 27.2.3 Section-Level Authorization
  - 27.2.4 Field-Level Authorization
  - 27.2.5 Row-Level Authorization (Fine-Grained)
  - 27.2.6 Action Authorization
  - 27.2.7 Data Source Authorization
- 27.3 OpenFGA Tuple Model for UI Permissions
- 27.4 Authorization Caching Strategy
- 27.5 Authorization in Multi-Tenant Context
- 27.6 Dynamic Permission Updates (Real-Time)
- 27.7 Authorization Audit Trail
- 27.8 Permission Denial UX Patterns

---

### Chapter 28 — Multi-Tenant Architecture

> **Audience:** Platform Engineers, Solution Architects · **Phase:** 3

- 28.1 Tenancy Model Overview
- 28.2 Tenant Resolution in UI Pipeline
  - 28.2.1 Tenant Identification (Subdomain, Header, Token)
  - 28.2.2 Tenant Context Object
- 28.3 Tenant-Specific UI Customizations
  - 28.3.1 Tenant Theme Overrides
  - 28.3.2 Tenant Logo / Branding
  - 28.3.3 Tenant-Specific Component Overrides
  - 28.3.4 Tenant-Specific Navigation
  - 28.3.5 Tenant-Specific Form Fields
- 28.4 Tenant Data Isolation in UI Payloads
- 28.5 Tenant Configuration Storage
  - 28.5.1 PostgreSQL Tenant Config Schema
  - 28.5.2 Redis Tenant Config Cache
- 28.6 Tenant Onboarding UI Flows
- 28.7 Super-Admin vs. Tenant-Admin UI Separation

---

### Chapter 29 — Feature Flag Architecture

> **Audience:** Platform Engineers, Backend Engineers, Product Owners · **Phase:** 3

- 29.1 Feature Flag Philosophy in SDUI
- 29.2 Flag Types
  - 29.2.1 Boolean Flags
  - 29.2.2 Multi-Variant Flags
  - 29.2.3 Percentage Rollout Flags
  - 29.2.4 Tenant-Targeted Flags
  - 29.2.5 User-Targeted Flags
  - 29.2.6 Time-Gated Flags
- 29.3 Flag Resolution in UI Compilation Pipeline
  - 29.3.1 Flag Evaluation Before AST Transformation
  - 29.3.2 Flag Context Object
  - 29.3.3 Flag DSL Directives in AST
- 29.4 Flag-Driven UI Patterns
  - 29.4.1 Showing/Hiding Components
  - 29.4.2 Switching Between Component Variants
  - 29.4.3 Enabling/Disabling Actions
  - 29.4.4 Swapping Data Sources
- 29.5 Feature Flag Service Integration
  - 29.5.1 SDK Integration in Go
  - 29.5.2 Flag Evaluation Caching
  - 29.5.3 Flag Change Propagation
- 29.6 Testing with Feature Flags
- 29.7 Flag Lifecycle and Cleanup

---

### Chapter 30 — Customization Framework

> **Audience:** ERP Implementers, Solution Architects · **Phase:** 3

- 30.1 Customization Layers
  - 30.1.1 Platform Layer (Core, Unmodifiable)
  - 30.1.2 Tenant Configuration Layer
  - 30.1.3 User Preference Layer
- 30.2 What Can Be Customized
  - 30.2.1 Layouts and Field Arrangements
  - 30.2.2 Field Labels and Help Text
  - 30.2.3 Required / Optional Fields
  - 30.2.4 Visible / Hidden Fields
  - 30.2.5 Default Values
  - 30.2.6 Validation Rules
  - 30.2.7 Navigation Structure
  - 30.2.8 Theme and Branding
  - 30.2.9 Custom Fields (EAV Pattern)
  - 30.2.10 Custom Actions
  - 30.2.11 Custom Reports / Dashboard Widgets
- 30.3 Customization Storage Model
- 30.4 Customization UI (Admin Panel)
- 30.5 Customization Merge Strategy
- 30.6 Customization Export and Import
- 30.7 Customization Version Control

---

### Chapter 31 — Extension and Plugin Architecture

> **Audience:** Third-Party Integrators, Platform Engineers · **Phase:** 3

- 31.1 Extension Model Overview
- 31.2 Plugin Types
  - 31.2.1 Custom Component Plugins
  - 31.2.2 Custom Action Plugins
  - 31.2.3 Custom Data Source Plugins
  - 31.2.4 Custom Validation Plugins
  - 31.2.5 Custom Theme Plugins
  - 31.2.6 Custom Navigation Plugins
- 31.3 Plugin Registration and Discovery
  - 31.3.1 Plugin Manifest Schema
  - 31.3.2 Plugin Registry Service
  - 31.3.3 Plugin Version Compatibility
- 31.4 Plugin Isolation and Sandboxing
- 31.5 Plugin Security Review Process
- 31.6 Plugin Distribution (Marketplace Model)
- 31.7 Plugin SDK (See Chapter 50)

---

## Volume VII — Reliability & Performance

---

### Chapter 32 — Performance Optimization

> **Audience:** Platform Engineers, Web/Mobile Engineers · **Phase:** 3

- 32.1 Performance Goals and SLOs
- 32.2 Payload Optimization
  - 32.2.1 JSON Compression (Gzip, Brotli)
  - 32.2.2 Partial Payloads (Field Projection)
  - 32.2.3 Incremental UI Updates (Diff Patching)
  - 32.2.4 Payload Size Budgets per Surface
- 32.3 Backend Compilation Performance
  - 32.3.1 AST Construction Benchmarks
  - 32.3.2 Transformation Pass Profiling
  - 32.3.3 Parallelization in Pipeline
- 32.4 Caching (See Chapter 34)
- 32.5 Client Rendering Performance
  - 32.5.1 Virtual/Windowed List Rendering
  - 32.5.2 Component Memoization
  - 32.5.3 Lazy Loading of Sub-Trees
  - 32.5.4 Image Optimization
- 32.6 Network Performance
  - 32.6.1 HTTP/2 Multiplexing
  - 32.6.2 Request Coalescing
  - 32.6.3 Prefetching and Preloading
- 32.7 Performance Testing and Benchmarking
- 32.8 Performance Regression Detection in CI/CD

---

### Chapter 33 — Offline Support

> **Audience:** Mobile Engineers, Web Engineers · **Phase:** 3

- 33.1 Offline Philosophy
- 33.2 Offline Modes
  - 33.2.1 Read-Only Offline (View Cached Data)
  - 33.2.2 Optimistic Write Offline (Queue Actions)
  - 33.2.3 Full Offline (Precached Business Process)
- 33.3 UI Definition Caching for Offline
  - 33.3.1 Which Surfaces to Cache
  - 33.3.2 Cache Invalidation Strategy
- 33.4 Data Caching for Offline
- 33.5 Offline Action Queue
  - 33.5.1 Action Serialization
  - 33.5.2 Conflict Resolution on Sync
  - 33.5.3 Failed Action Handling
- 33.6 Offline Indicators in UI
- 33.7 Sync on Reconnect (See Chapter 35)

---

### Chapter 34 — Caching Strategies

> **Audience:** Platform Engineers, Backend Engineers · **Phase:** 3

- 34.1 Caching Philosophy
- 34.2 Cache Taxonomy
  - 34.2.1 UI Definition Cache (Redis)
  - 34.2.2 Authorization Cache
  - 34.2.3 Feature Flag Cache
  - 34.2.4 Data Source Response Cache
  - 34.2.5 Client-Side UI Cache
- 34.3 Cache Key Design
  - 34.3.1 Surface + Tenant + Role + Flags as Cache Key
  - 34.3.2 Cache Key Hashing Strategy
- 34.4 Cache Invalidation
  - 34.4.1 TTL-Based Invalidation
  - 34.4.2 Event-Driven Invalidation
  - 34.4.3 Manual Invalidation (Admin)
- 34.5 Cache Warming
- 34.6 Cache Stampede Prevention
- 34.7 Cache Observability

---

### Chapter 35 — Synchronization

> **Audience:** Mobile Engineers, Web Engineers, Backend Engineers · **Phase:** 3

- 35.1 Sync Philosophy
- 35.2 UI Definition Sync
  - 35.2.1 Version-Based Sync (ETag / If-None-Match)
  - 35.2.2 Polling vs. Push-Based Sync
- 35.3 Data Sync
  - 35.3.1 Delta Sync (Changed Records Only)
  - 35.3.2 Conflict Detection
  - 35.3.3 Conflict Resolution Strategies (Last-Write-Wins, Merge, User-Prompted)
- 35.4 Real-Time Sync (WebSocket / SSE)
- 35.5 Background Sync on Mobile
- 35.6 Sync State Indicators

---

## Volume VIII — Developer Experience

---

### Chapter 36 — Internationalization (i18n)

> **Audience:** All Engineers · **Phase:** 4

- 36.1 i18n Philosophy
- 36.2 String Externalization
  - 36.2.1 Translation Keys in AST
  - 36.2.2 Locale Resolution Pipeline
  - 36.2.3 Fallback Locale Chain
- 36.3 Locale-Specific Formatting
  - 36.3.1 Number Formatting
  - 36.3.2 Currency Formatting
  - 36.3.3 Date / Time Formatting
  - 36.3.4 Pluralization
- 36.4 Right-to-Left (RTL) Layout Support
  - 36.4.1 Layout Mirroring
  - 36.4.2 RTL-Aware Components
- 36.5 Locale Switching at Runtime
- 36.6 Tenant-Specific Translations
- 36.7 Translation Management Workflow

---

### Chapter 37 — Accessibility (a11y)

> **Audience:** Web/Mobile Engineers, UI Framework Developers · **Phase:** 4

- 37.1 Accessibility Philosophy
- 37.2 Compliance Targets (WCAG 2.1 AA, Section 508)
- 37.3 Accessibility in the DSL
  - 37.3.1 `aria-*` Props in Component Schema
  - 37.3.2 Role Declarations
  - 37.3.3 Label Associations
  - 37.3.4 Live Regions
- 37.4 Keyboard Navigation
- 37.5 Screen Reader Support
- 37.6 Color Contrast and Visual Design
- 37.7 Focus Management in Navigation
- 37.8 Mobile Accessibility (iOS VoiceOver, Android TalkBack)
- 37.9 Accessibility Testing

---

### Chapter 38 — Observability

> **Audience:** Platform Engineers, DevOps · **Phase:** 4

- 38.1 Observability Philosophy (Pillars: Traces, Metrics, Logs)
- 38.2 Tracing
  - 38.2.1 Distributed Trace Propagation
  - 38.2.2 Per-Pipeline-Stage Spans
  - 38.2.3 Client-Side Trace Contribution
- 38.3 Metrics
  - 38.3.1 UI Definition Compilation Latency
  - 38.3.2 Payload Size Distribution
  - 38.3.3 Cache Hit Rate
  - 38.3.4 Client Render Time
  - 38.3.5 Action Success / Failure Rate
- 38.4 Alerting Thresholds
- 38.5 Observability Infrastructure (OpenTelemetry, Prometheus, Grafana)

---

### Chapter 39 — Telemetry

> **Audience:** Product Owners, Platform Engineers · **Phase:** 4

- 39.1 Telemetry Philosophy (Product Analytics + Operational)
- 39.2 Client Telemetry Events
  - 39.2.1 Page View Events
  - 39.2.2 Component Interaction Events
  - 39.2.3 Form Submission Events
  - 39.2.4 Error Events
  - 39.2.5 Performance Events
- 39.3 Telemetry Schema
- 39.4 Telemetry Privacy and Consent
- 39.5 Telemetry Pipeline
- 39.6 Telemetry Dashboards

---

### Chapter 40 — Logging

> **Audience:** Platform Engineers, DevOps · **Phase:** 4

- 40.1 Logging Standards
- 40.2 Structured Log Format
- 40.3 Log Taxonomy
  - 40.3.1 Compilation Logs
  - 40.3.2 Authorization Decision Logs
  - 40.3.3 Action Execution Logs
  - 40.3.4 Error Logs
  - 40.3.5 Audit Logs
- 40.4 Sensitive Data in Logs (PII Scrubbing)
- 40.5 Log Aggregation and Retention

---

## Volume IX — Engineering Excellence

---

### Chapter 41 — Testing Strategy

> **Audience:** All Engineers · **Phase:** 4

- 41.1 Testing Philosophy
- 41.2 Testing Pyramid for SDUI
  - 41.2.1 Unit Tests — AST Node Builders
  - 41.2.2 Unit Tests — Transformation Passes
  - 41.2.3 Unit Tests — Validation Rules
  - 41.2.4 Integration Tests — Compilation Pipeline
  - 41.2.5 Contract Tests — JSON Schema Compliance
  - 41.2.6 Snapshot Tests — UI Definition Output
  - 41.2.7 Component Tests — Web Rendering
  - 41.2.8 Component Tests — Mobile Rendering
  - 41.2.9 End-to-End Tests — Full Surface Rendering
  - 41.2.10 Visual Regression Tests
  - 41.2.11 Performance Tests
  - 41.2.12 Accessibility Tests
- 41.3 Test Data Management
- 41.4 Mocking the UI Compilation Service
- 41.5 Testing Multi-Tenant Scenarios
- 41.6 Testing Feature Flag Combinations
- 41.7 Testing Offline Scenarios

---

### Chapter 42 — CI/CD Considerations

> **Audience:** DevOps, Platform Engineers · **Phase:** 4

- 42.1 CI/CD Pipeline Overview
- 42.2 Schema Validation in CI
- 42.3 Breaking Change Detection in CI
- 42.4 Visual Regression Testing in CI
- 42.5 Performance Budget Enforcement in CI
- 42.6 Multi-Platform Build Strategy (Web + iOS + Android)
- 42.7 Feature Flag-Gated Deployments
- 42.8 Canary Deployments for UI Platform Changes

---

### Chapter 43 — Versioning Strategy

> **Audience:** Platform Engineers, Third-Party Integrators · **Phase:** 4

- 43.1 Versioning Philosophy
- 43.2 Schema Versioning
  - 43.2.1 Semantic Versioning for Schema
  - 43.2.2 Schema Compatibility Matrix
  - 43.2.3 Forward vs. Backward Compatibility
- 43.3 API Versioning (URL vs. Header)
- 43.4 Component Versioning
- 43.5 Client Version Negotiation
- 43.6 Multi-Version Support Window
- 43.7 Deprecation Timeline Policy

---

### Chapter 44 — Migration Strategy

> **Audience:** Platform Engineers, ERP Implementers · **Phase:** 4

- 44.1 Migration Philosophy
- 44.2 Schema Migration Patterns
  - 44.2.1 Additive Migrations
  - 44.2.2 Rename Migrations
  - 44.2.3 Restructure Migrations
  - 44.2.4 Breaking Change Migrations
- 44.3 Client Migration
  - 44.3.1 Forced Upgrade vs. Graceful Degradation
  - 44.3.2 Client Version Detection
- 44.4 Data Migration for UI Customizations
- 44.5 Tenant Migration Coordination
- 44.6 Migration Testing Runbook

---

### Chapter 45 — Governance Model

> **Audience:** Platform Engineers, Product Owners · **Phase:** 4

- 45.1 Platform Governance Philosophy
- 45.2 Component Proposal Process
- 45.3 Schema Change RFC Process
- 45.4 Breaking Change Approval Gates
- 45.5 Platform Ownership and Stewardship
- 45.6 Third-Party Component Vetting
- 45.7 Documentation Standards
- 45.8 Accessibility Review Gate
- 45.9 Security Review Gate

---

## Volume X — Reference & Guidance

---

### Chapter 46 — Best Practices

> **Audience:** All Engineers · **Phase:** 4

- 46.1 DSL Authoring Best Practices
- 46.2 AST Construction Best Practices
- 46.3 Component Composition Best Practices
- 46.4 Performance Best Practices
- 46.5 Security Best Practices
- 46.6 Multi-Tenant Best Practices
- 46.7 Offline Best Practices
- 46.8 Testing Best Practices

---

### Chapter 47 — Anti-Patterns

> **Audience:** All Engineers · **Phase:** 4

- 47.1 Anti-Pattern: Business Logic in Rendering Engines
- 47.2 Anti-Pattern: Client-Only Permission Enforcement
- 47.3 Anti-Pattern: Monolithic Page Definitions
- 47.4 Anti-Pattern: Coupling UI Types to Database Schema
- 47.5 Anti-Pattern: Giant Payload Syndrome
- 47.6 Anti-Pattern: Skipping Validation in Pipeline
- 47.7 Anti-Pattern: Hardcoding Tenant Logic in Core
- 47.8 Anti-Pattern: Action Chains Without Error Branches
- 47.9 Anti-Pattern: Ignoring Accessibility in DSL
- 47.10 Anti-Pattern: Inconsistent Component Versioning

---

### Chapter 48 — Reference Architecture

> **Audience:** Solution Architects, ERP Implementers · **Phase:** 4

- 48.1 Reference Architecture Overview
- 48.2 Small Tenant Deployment (Single-Region)
- 48.3 Enterprise Multi-Tenant Deployment
- 48.4 High-Availability Configuration
- 48.5 Disaster Recovery Configuration
- 48.6 Hybrid On-Premise + Cloud Deployment
- 48.7 Infrastructure as Code Templates

---

### Chapter 49 — End-to-End Examples

> **Audience:** All Engineers · **Phase:** 4

- 49.1 Example 1 — Purchase Order Form (Full Lifecycle)
  - 49.1.1 Business Service Emitting AST in Go
  - 49.1.2 Permission Pruning in Action
  - 49.1.3 Compiled JSON Output
  - 49.1.4 Web Rendering Result
  - 49.1.5 Mobile Rendering Result
- 49.2 Example 2 — Invoice Approval Workflow
  - 49.2.1 Workflow Status Component
  - 49.2.2 Approval Action Bar
  - 49.2.3 Temporal Signal Integration
- 49.3 Example 3 — Multi-Tenant Dashboard
  - 49.3.1 Tenant A (Standard Layout)
  - 49.3.2 Tenant B (Custom Widgets via Feature Flag)
  - 49.3.3 Compiled Difference Between Tenants
- 49.4 Example 4 — Dynamic ERP Report Table
  - 49.4.1 Column Permission Pruning
  - 49.4.2 Aggregation Row
  - 49.4.3 Export Action
- 49.5 Example 5 — Offline-Capable Stock Take Form

---

### Chapter 50 — SDK Development

> **Audience:** Third-Party Integrators, Platform Engineers · **Phase:** 4

- 50.1 SDK Design Goals
- 50.2 Go SDK (Backend / UI Compiler)
  - 50.2.1 AST Builder Helpers
  - 50.2.2 Typed Component Factories
  - 50.2.3 Validation Utilities
  - 50.2.4 Pipeline Extension Hooks
- 50.3 TypeScript SDK (Web Rendering Engine)
  - 50.3.1 JSON Schema Types (Auto-Generated)
  - 50.3.2 Component Registration API
  - 50.3.3 Action Registration API
  - 50.3.4 Custom Data Source Adapter
- 50.4 iOS SDK (Swift)
  - 50.4.1 AST Decoder
  - 50.4.2 Native Component Bridge
  - 50.4.3 Custom Component Registration
- 50.5 Android SDK (Kotlin)
  - 50.5.1 AST Decoder
  - 50.5.2 Composable Component Bridge
  - 50.5.3 Custom Component Registration
- 50.6 SDK Versioning and Compatibility
- 50.7 SDK Documentation and Examples
- 50.8 SDK Publishing (npm, CocoaPods, Maven)

---

### Chapter 51 — Future Roadmap

> **Audience:** Product Owners, Platform Engineers · **Phase:** 4

- 51.1 Platform Maturity Model
- 51.2 Near-Term Roadmap (0–6 Months)
  - 51.2.1 Core AST Stabilization
  - 51.2.2 Web and Mobile Rendering v1
  - 51.2.3 ERP Component Library v1
- 51.3 Mid-Term Roadmap (6–18 Months)
  - 51.3.1 No-Code Customization UI
  - 51.3.2 AI-Assisted Form Generation
  - 51.3.3 Advanced Analytics Components
  - 51.3.4 Plugin Marketplace
- 51.4 Long-Term Roadmap (18+ Months)
  - 51.4.1 Voice / Conversational UI Surface
  - 51.4.2 AR / Industrial Terminal Surface
  - 51.4.3 Third-Party ERP Connector Framework
  - 51.4.4 AI-Driven Layout Optimization
- 51.5 Planned Breaking Changes and Migration Windows
- 51.6 Research Topics

---

## Identified Additional Topics (Recommended Additions)

The following topics were identified as important but not explicitly listed in the original requirements. They are **recommended for inclusion**:

| # | Proposed Chapter / Section | Rationale | Suggested Location |
|---|---------------------------|-----------|-------------------|
| A | **GLOSSARY** — Platform-wide terminology (SDUI, AST, Surface, Slot, etc.) | Critical for alignment across teams | Top-level file |
| B | **Error Handling Strategy** — Standard error types, client error display, retry policies | Needed for all action and data source chapters | Vol. V, after Ch. 24 |
| C | **Developer Tooling** — CLI tools, local dev emulator, JSON validator, schema browser | Dramatically accelerates adoption | Vol. VIII or X |
| D | **UI Composition Patterns** — Templates, inheritance, mixins, slot overrides | Core reuse mechanism not covered in AST chapter alone | Vol. II, after Ch. 7 |
| E | **Audit & Compliance** — GDPR, SOC2, ISO 27001 considerations for UI platform | Required for enterprise ERP customers | Vol. VI |
| F | **Real-Time UI Updates** — Live dashboard auto-refresh, push-driven form field updates | Distinct from event system; needs its own treatment | Vol. IV or V |
| G | **Print and Export Rendering** — PDF generation from UI definitions, print layouts | ERP requirement (invoices, reports, picking slips) | Vol. III |
| H | **Theming and Design Tokens** — Full token system, tenant branding, dark mode | Referenced in multiple chapters but needs a dedicated chapter | Vol. III |
| I | **Onboarding and Migration Guide for Implementers** — Step-by-step for new tenants | Essential for ERP deployment teams | Vol. X |
| J | **Playground / Sandbox Environment** — Interactive UI builder for testing definitions | Accelerates developer onboarding | Vol. VIII or X |

---

## Summary Statistics

| Metric | Count |
|--------|-------|
| Volumes | 10 |
| Core Chapters | 51 |
| Recommended Additional Topics | 10 |
| Top-Level Sections | ~200+ |
| Estimated Pages (when written) | 800–1,200 |
| Documentation Phases | 4 |
| Phase 1 Chapters | 8 |
| Phase 2 Chapters | 13 |
| Phase 3 Chapters | 14 |
| Phase 4 Chapters | 16 |

---

*End of Proposed Table of Contents — Awaiting Review and Approval*
