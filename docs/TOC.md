# AwoERP — Complete Documentation Architecture Guide

> **Edition:** Comprehensive Reference · **Version:** 1.0.0  
> **Scope:** Platform & Core Modules · Business Modules · Server-Driven UI Platform  
> **Status:** Authoritative — Supersedes all draft TOC documents  
> **Audience:** Platform Engineers · Backend Engineers · Mobile Engineers · Web Engineers · UI Framework Developers · Product Owners · Solution Architects · ERP Implementers · Third-Party Integrators

---

## Preface

This document is the unified, comprehensive reference for the entire AwoERP documentation corpus. It synthesizes three foundational documentation trees:

1. **Platform & Core Modules** — the foundation every business module builds upon
2. **Business Modules** — domain-specific ERP functionality (Finance, CRM, HR, Forecourt, Airline, and more)
3. **Server-Driven UI (SDUI) Platform** — the architecture that drives UI across web, mobile, and future surfaces from a single backend source of truth

The structure below follows the SDUI Platform documentation model (10 Volumes, phased delivery) and extends it to cover Platform Core and all Business Modules with the same rigour. Every section is annotated with its audience, delivery phase, and rationale.

---

## Document Map

| Layer | Coverage | Volumes | Chapters |
|---|---|---|---|
| SDUI Platform | UI compilation, AST, components, rendering | Vol. I–X | 51 core + 10 recommended |
| Platform Core | Tenancy, IAM, Workflow, Audit, Integration | Part A | 20 sections |
| Business Modules | Finance, CRM, HR, Forecourt, Airline, etc. | Part B | 17 modules |

**Total estimated documentation:** 1,500–2,000 pages when fully written.

---

## Repository Structure (Recommended)

```
./docs/
│
├── README.md                          # Docs home, quick-start, contribution guide
├── GLOSSARY.md                        # Platform-wide terminology (SDUI, AST, Tenant, Surface, Slot…)
├── CHANGELOG.md                       # Documentation changelog
├── ARCHITECTURE_DECISION_LOG.md       # ADRs for key platform decisions
├── mkdocs.yml                         # Or docusaurus.config.js
│
├── platform/                          # Part A — Platform & Core
│   ├── 01-overview/
│   ├── 02-tenant-management/
│   ├── 03-iam/
│   ├── 04-feature-flags/
│   ├── 05-configuration/
│   ├── 06-workflow/
│   ├── 07-approvals/
│   ├── 08-notifications/
│   ├── 09-documents/
│   ├── 10-audit-compliance/
│   ├── 11-reporting/
│   ├── 12-analytics-dashboards/
│   ├── 13-integration/
│   ├── 14-data-management/
│   ├── 15-search/
│   ├── 16-administration/
│   ├── 17-security-governance/
│   ├── 18-operations/
│   ├── 19-developer-resources/
│   └── 20-troubleshooting/
│
├── modules/                           # Part B — Business Modules
│   ├── finance/
│   ├── crm/
│   ├── sales/
│   ├── procurement/
│   ├── inventory/
│   ├── warehouse/
│   ├── hr/
│   ├── payroll/
│   ├── manufacturing/
│   ├── projects/
│   ├── service/
│   ├── assets/
│   ├── travel/
│   ├── property/
│   ├── forecourt/
│   ├── airline/
│   └── bi/
│
├── ui/                                # Part C — SDUI Platform (10 Volumes)
│   ├── vol-01-vision/
│   ├── vol-02-dsl-and-ast/
│   ├── vol-03-component-system/
│   ├── vol-04-rendering/
│   ├── vol-05-runtime/
│   ├── vol-06-platform/
│   ├── vol-07-reliability/
│   ├── vol-08-dx/
│   ├── vol-09-engineering/
│   ├── vol-10-reference/
│   ├── schemas/
│   ├── examples/
│   └── assets/
│
└── shared/
    ├── error-codes/
    ├── api-reference/
    └── release-notes/
```

---

## Delivery Phases

| Phase | Parts | Focus | Priority |
|---|---|---|---|
| **Phase 1 — Foundation** | SDUI Vol. I–II + Platform Core Overview + IAM | Vision, DSL, AST, Tenant Architecture | Critical — ship first |
| **Phase 2 — Components & Business Logic** | SDUI Vol. III–IV + Finance + CRM + HR core | All UI components, web/mobile rendering, primary ERP modules | Critical — ship second |
| **Phase 3 — Platform Depth & Secondary Modules** | SDUI Vol. V–VII + remaining Business Modules | State, APIs, security, tenancy, performance, Forecourt, Airline | High |
| **Phase 4 — Excellence, Reference & BI** | SDUI Vol. VIII–X + BI Module + SDK | DX, testing, governance, SDK, end-to-end examples | Medium |

---

---

# PART A — Platform & Core Modules

*The foundation that all business modules build upon. Every AwoERP installation shares this layer regardless of which business modules are activated.*

---

## Section 1 — Platform Overview

> **Phase:** 1 · **Audience:** All

### 1.1 Introduction

- **1.1.1** What is AwoERP — mission, positioning, and competitive context
- **1.1.2** Platform Architecture Overview — layered diagram from infrastructure to UI surfaces
- **1.1.3** Core Concepts Glossary Reference — Tenant, Workspace, Branch, Module, Surface
- **1.1.4** Platform Capabilities Matrix — which capabilities are available per deployment tier
- **1.1.5** Supported Deployment Models — SaaS multi-tenant, private cloud, on-premise, hybrid
- **1.1.6** Multi-Tenant Architecture — data isolation, resource sharing, tenant lifecycle overview
- **1.1.7** Platform Terminology — authoritative definitions used throughout all documentation

### 1.2 Getting Started

- **1.2.1** First Login — walkthrough for new users, MFA setup, workspace selection
- **1.2.2** Understanding Workspaces — personal vs. shared workspaces, context switching
- **1.2.3** Navigation Overview — primary navigation, global search, breadcrumbs, quick actions
- **1.2.4** Search and Global Actions — keyboard shortcuts, command palette, global search syntax
- **1.2.5** User Preferences — language, timezone, date/number format, notification preferences
- **1.2.6** Personal Settings — profile, password, connected devices, active sessions

### 1.3 Platform Concepts

Deep explanations of each foundational entity, its lifecycle, relationships, and common configuration patterns:

- **1.3.1** Organizations — root entity, multi-company structure, legal entity mapping
- **1.3.2** Tenants — isolation boundary, subscription model, lifecycle states
- **1.3.3** Departments — hierarchical structure, cost center alignment, HR linkage
- **1.3.4** Branches — physical locations, regional groupings, branch-level configuration
- **1.3.5** Projects — cross-functional containers, budget and cost tracking
- **1.3.6** Users — identity, profile, role assignments, status management
- **1.3.7** Roles — RBAC model, role templates, custom role creation
- **1.3.8** Permissions — permission types, inheritance, effective permission resolution
- **1.3.9** Workflows — definition model, execution model, human vs. system tasks
- **1.3.10** Documents — unified document model, versioning, metadata
- **1.3.11** Notifications — notification taxonomy, channel routing, preference model
- **1.3.12** Reports — report types, scheduling, distribution, export formats

---

## Section 2 — Tenant Management

> **Phase:** 1 · **Audience:** Platform Administrators, Solution Architects

### 2.1 Tenant Overview

- **2.1.1** Understanding Tenants — definition, isolation guarantees, billing unit
- **2.1.2** Tenant Isolation Model — data layer, compute layer, network layer isolation
- **2.1.3** Tenant Lifecycle — draft → provisioning → active → suspended → terminated
- **2.1.4** Tenant Architecture — single-schema vs. multi-schema, shared services

### 2.2 Tenant Setup

- **2.2.1** Creating a Tenant — admin workflow, required inputs, provisioning process
- **2.2.2** Tenant Registration — registration form, domain verification, initial admin assignment
- **2.2.3** Tenant Configuration — module activation, storage quotas, API limits
- **2.2.4** Initial Setup Wizard — guided onboarding: org structure, users, chart of accounts
- **2.2.5** Activation Process — activation checklist, go-live validation, handover

### 2.3 Organizational Structure

Each entity type has: definition, creation workflow, hierarchy rules, permission boundaries, and reporting implications.

- **2.3.1** Companies — legal entity setup, company codes, intercompany relationships
- **2.3.2** Departments — department hierarchy, budget ownership, HR integration
- **2.3.3** Divisions — business division structure, P&L ownership
- **2.3.4** Branches — physical location registry, branch-level configuration, POS/forecourt linkage
- **2.3.5** Regions — geographic groupings, regional reporting, region-level policies
- **2.3.6** Projects — project types, lifecycle, budget linkage, member management
- **2.3.7** Cost Centers — cost center creation, allocation rules, reporting
- **2.3.8** Business Units — business unit structure, strategic groupings

### 2.4 Tenant Administration

- **2.4.1** Managing Tenant Settings — runtime configuration, feature activation, limits
- **2.4.2** Tenant Branding — logo upload, primary colors, email branding, white-label config
- **2.4.3** Localization — locale settings, supported languages, regional defaults
- **2.4.4** Time Zones — tenant default TZ, branch-level overrides, DST handling
- **2.4.5** Languages — UI language, document language, multilingual field support
- **2.4.6** Currency Settings — base currency, functional currencies, FX rate sources

### 2.5 Multi-Tenant Governance

- **2.5.1** Shared Resources — shared services, shared component libraries, shared lookup data
- **2.5.2** Resource Isolation — CPU/memory quotas, storage isolation, network isolation
- **2.5.3** Cross-Tenant Restrictions — data access rules, API isolation, audit boundaries
- **2.5.4** Tenant Policies — security policies, compliance policies, SLA policies

---

## Section 3 — Identity & Access Management (IAM)

> **Phase:** 1 · **Audience:** Security Engineers, Administrators, All Engineers

### 3.1 IAM Overview

- **3.1.1** Security Model — RBAC + ABAC hybrid model, OpenFGA integration
- **3.1.2** Access Control Concepts — subjects, resources, actions, conditions
- **3.1.3** Authentication vs. Authorization — clear boundary, enforcement points

### 3.2 User Management

- **3.2.1** Creating Users — manual creation, invitation flow, bulk import
- **3.2.2** User Profiles — identity attributes, contact info, system preferences
- **3.2.3** User Status Management — active, suspended, locked, pending states
- **3.2.4** User Lifecycle — onboarding, role changes, offboarding, data retention
- **3.2.5** User Import — CSV import format, validation rules, error handling
- **3.2.6** User Deactivation — deactivation vs. deletion, data ownership transfer

### 3.3 Roles

- **3.3.1** Understanding Roles — role types, role scope (global, tenant, module, branch)
- **3.3.2** Creating Roles — role builder, permission assignment, role hierarchy
- **3.3.3** Managing Roles — editing, cloning, comparing, deprecating roles
- **3.3.4** Role Templates — system-provided templates (Finance Manager, HR Clerk, etc.)
- **3.3.5** Custom Roles — building custom roles from permission primitives

### 3.4 Permissions

- **3.4.1** Permission Types — read, write, delete, execute, approve, admin
- **3.4.2** Resource Permissions — entity-level, module-level, field-level
- **3.4.3** Action Permissions — workflow actions, approval actions, export actions
- **3.4.4** Permission Inheritance — role hierarchy, group membership, explicit overrides
- **3.4.5** Effective Permissions — how effective permissions are computed, debugging tools

### 3.5 Groups

- **3.5.1** User Groups — purpose, creation, membership rules, dynamic groups
- **3.5.2** Team Management — team structure, team ownership, team-based routing
- **3.5.3** Group-Based Permissions — group-to-role mapping, group permission policies

### 3.6 Authentication

- **3.6.1** Login Methods — username/password, SSO, magic link, biometric (mobile)
- **3.6.2** Password Policies — complexity rules, rotation, breach detection
- **3.6.3** Multi-Factor Authentication — TOTP, SMS OTP, push auth, hardware key (FIDO2)
- **3.6.4** Password Recovery — self-service reset, admin-assisted reset, account unlock
- **3.6.5** Session Security — session lifetime, concurrent session limits, suspicious activity detection

### 3.7 Single Sign-On

- **3.7.1** SSO Concepts — IdP vs. SP roles, trust establishment, attribute mapping
- **3.7.2** SAML Integration — SAML 2.0 configuration, metadata exchange, assertion mapping
- **3.7.3** OpenID Connect — OIDC flow, client registration, claim mapping
- **3.7.4** OAuth Authentication — OAuth 2.0 flows, token management, scope mapping

### 3.8 Access Reviews

- **3.8.1** Reviewing User Access — access review campaigns, reviewer assignment
- **3.8.2** Certification Campaigns — periodic certification, automated scheduling
- **3.8.3** Audit Reviews — access audit reports, orphaned access detection

### 3.9 Security Policies

- **3.9.1** Password Policies — tenant-level, department-level, user-level policies
- **3.9.2** Login Restrictions — IP whitelist/blacklist, time-based restrictions
- **3.9.3** IP Restrictions — CIDR-based access control, VPN enforcement
- **3.9.4** Device Policies — trusted device management, device fingerprinting

---

## Section 4 — Feature Flag Management

> **Phase:** 1 · **Audience:** Product Owners, Platform Engineers, Backend Engineers

### 4.1 Overview

- **4.1.1** Feature Flag Concepts — boolean flags, variant flags, kill switches
- **4.1.2** Benefits — safe deployments, A/B testing, tenant-specific activation, gradual rollouts
- **4.1.3** Rollout Strategies — immediate, gradual, canary, tenant-targeted, user-targeted

### 4.2 Feature Lifecycle

- **4.2.1** Draft Features — creation, configuration, targeting rules, internal testing
- **4.2.2** Testing Features — QA targeting, test tenant activation, validation gates
- **4.2.3** Production Features — production rollout, monitoring, adoption metrics
- **4.2.4** Retirement Process — deprecation notice, cleanup tasks, flag removal

### 4.3 Feature Configuration

- **4.3.1** Creating Features — name, description, type, default state, environments
- **4.3.2** Feature Metadata — owner, team, linked issues, documentation links
- **4.3.3** Dependencies — feature dependency graphs, prerequisite flags

### 4.4 Targeting

- **4.4.1** Tenant Targeting — explicit tenant list, tenant attribute rules
- **4.4.2** User Targeting — user ID list, user attribute rules
- **4.4.3** Role Targeting — enable by role, disable for role
- **4.4.4** Segment Targeting — custom segments, dynamic segment rules

### 4.5 Rollouts

- **4.5.1** Percentage Rollouts — random percentage, deterministic hashing
- **4.5.2** Gradual Releases — time-phased percentage increase
- **4.5.3** Canary Releases — canary tenant/user group selection, promotion workflow
- **4.5.4** Controlled Releases — explicit approval gate before broader rollout

### 4.6 Monitoring

- **4.6.1** Usage Metrics — evaluation count, variant distribution, latency impact
- **4.6.2** Adoption Tracking — cohort adoption, feature engagement analytics
- **4.6.3** Rollout Analysis — flag performance, rollback triggers, experiment results

---

## Section 5 — Configuration Management

> **Phase:** 1 · **Audience:** Platform Administrators, Solution Architects

### 5.1 Overview

- **5.1.1** Configuration Architecture — layered hierarchy, precedence rules, override model
- **5.1.2** Configuration Hierarchy — platform defaults → tenant config → branch config → user preferences

### 5.2 System Settings

- **5.2.1** Global Settings — platform-wide defaults affecting all tenants
- **5.2.2** Default Values — system default values and how they cascade
- **5.2.3** Configuration Scopes — global, tenant, branch, department, user scopes

### 5.3 Tenant Settings

- **5.3.1** Tenant Configuration — tenant-level overrides, feature activation, module settings
- **5.3.2** Branch Configuration — branch-specific operational settings
- **5.3.3** Department Configuration — department-level policy and default settings

### 5.4 Dynamic Configuration

- **5.4.1** Runtime Updates — configuration changes without restart, propagation latency
- **5.4.2** Configuration Validation — validation rules, invalid state detection
- **5.4.3** Configuration Overrides — override hierarchy, conflict resolution

### 5.5 Localization

- **5.5.1** Languages — supported locales, locale package management, custom translations
- **5.5.2** Date Formats — format strings, regional defaults, calendar types
- **5.5.3** Number Formats — decimal separators, thousand separators, significant digits
- **5.5.4** Currency Formatting — symbol placement, decimal precision, currency display rules

### 5.6 Branding

- **5.6.1** Logos — logo upload, format requirements, placement contexts
- **5.6.2** Themes — color palette, typography, component style overrides
- **5.6.3** Email Branding — email header/footer, from-name, reply-to configuration
- **5.6.4** White Label Settings — domain configuration, app name customization, favicon

---

## Section 6 — Workflow Management

> **Phase:** 2 · **Audience:** Backend Engineers, Business Analysts, Solution Architects

### 6.1 Workflow Overview

- **6.1.1** Workflow Concepts — workflow vs. process vs. task, stateful vs. stateless
- **6.1.2** Workflow Lifecycle — draft → active → completed → archived
- **6.1.3** Workflow Components — states, transitions, conditions, actions, tasks

### 6.2 Workflow Design

- **6.2.1** Workflow Definitions — YAML/JSON definition format, versioning
- **6.2.2** States — state types (initial, intermediate, terminal), state metadata
- **6.2.3** Transitions — trigger types, guard conditions, transition actions
- **6.2.4** Conditions — condition expressions, context variables, rule evaluation
- **6.2.5** Actions — system actions, notification actions, integration actions

### 6.3 Workflow Execution

- **6.3.1** Starting Workflows — programmatic start, event-triggered start, manual start
- **6.3.2** Workflow Tracking — current state, history, pending tasks, participants
- **6.3.3** Workflow History — complete audit trail, state change log, time-in-state metrics

### 6.4 Workflow Tasks

- **6.4.1** User Tasks — assignment rules, task UI, completion conditions
- **6.4.2** System Tasks — automated task execution, retry policy, failure handling
- **6.4.3** Scheduled Tasks — time-triggered tasks, cron expressions, calendar-aware scheduling
- **6.4.4** Escalation Tasks — escalation rules, escalation chains, notification on escalation

### 6.5 Workflow Monitoring

- **6.5.1** Active Workflows — real-time active workflow dashboard, filter by state/owner
- **6.5.2** Failed Workflows — error classification, retry options, manual intervention
- **6.5.3** Performance Metrics — throughput, average cycle time, bottleneck analysis

### 6.6 Workflow Governance

- **6.6.1** Workflow Versioning — semantic versioning, version coexistence, migration
- **6.6.2** Change Management — RFC process for workflow changes, impact analysis
- **6.6.3** Workflow Auditing — immutable audit trail, compliance reporting

---

## Section 7 — Approval Management

> **Phase:** 2 · **Audience:** Business Analysts, Backend Engineers, ERP Implementers

### 7.1 Overview

- **7.1.1** Approval Concepts — approval vs. workflow, decision types, approval authority
- **7.1.2** Approval Types — single approver, parallel, sequential, majority vote, any-of

### 7.2 Approval Policies

- **7.2.1** Policy Definitions — policy scope, triggers, approver resolution rules
- **7.2.2** Conditional Approvals — threshold-based routing (e.g. amount > $10,000 requires CFO)
- **7.2.3** Hierarchical Approvals — org chart traversal, level-based routing

### 7.3 Approval Chains

- **7.3.1** Sequential Approvals — ordered chain, each step requires previous completion
- **7.3.2** Parallel Approvals — all approvers notified simultaneously, configurable completion rule
- **7.3.3** Dynamic Approvals — approver list computed at runtime from context

### 7.4 Delegation

- **7.4.1** Temporary Delegation — date-bounded authority transfer
- **7.4.2** Permanent Delegation — standing delegation with scope restrictions
- **7.4.3** Escalation Rules — auto-escalation after timeout, escalation notification

### 7.5 Approval Tracking

- **7.5.1** Pending Approvals — approver inbox, priority sorting, bulk approval
- **7.5.2** Approval History — complete decision log, reason capture, reversal audit
- **7.5.3** Approval Analytics — SLA compliance, approver response time, bottleneck reports

---

## Section 8 — Notification Management

> **Phase:** 2 · **Audience:** Backend Engineers, Solution Architects

### 8.1 Overview

- **8.1.1** Notification Framework — channel abstraction, routing rules, deduplication
- **8.1.2** Notification Channels — in-app, email, SMS, push (mobile + browser), webhook

### 8.2 In-App Notifications

- **8.2.1** User Notifications — notification bell, notification list, read/unread state
- **8.2.2** System Notifications — platform-level alerts, maintenance notices
- **8.2.3** Actionable Notifications — approve/reject/acknowledge directly from notification

### 8.3 Email Notifications

- **8.3.1** Templates — template editor, variable substitution, HTML/plain-text versions
- **8.3.2** Delivery Rules — trigger conditions, recipient resolution, batching
- **8.3.3** Tracking — open rates, click tracking, bounce handling, deliverability monitoring

### 8.4 SMS Notifications

- **8.4.1** SMS Providers — Twilio, Africa's Talking, custom gateway integration
- **8.4.2** Delivery Tracking — delivery receipts, failure handling, retry logic

### 8.5 Push Notifications

- **8.5.1** Mobile Push — FCM/APNs integration, device token management, rich push
- **8.5.2** Browser Push — Web Push API, permission management, service worker

### 8.6 Notification Preferences

- **8.6.1** User Preferences — per-notification-type channel preference
- **8.6.2** Opt-In Management — mandatory vs. optional notifications, GDPR consent
- **8.6.3** Quiet Hours — do-not-disturb windows, timezone-aware scheduling

### 8.7 Templates

- **8.7.1** Template Management — template library, versioning, environment promotion
- **8.7.2** Variables — standard variables, context variables, custom expressions
- **8.7.3** Localization — per-locale templates, fallback chain, translation workflow

---

## Section 9 — Document Management

> **Phase:** 2 · **Audience:** All Engineers, Business Users

### 9.1 Overview

- **9.1.1** Document Concepts — document vs. file, document entity linkage, document types
- **9.1.2** Document Lifecycle — draft → submitted → approved → archived

### 9.2 File Storage

- **9.2.1** Uploading Files — upload API, chunked upload, virus scanning, format validation
- **9.2.2** Organizing Files — folder structure, document sets, tagging
- **9.2.3** Storage Limits — tenant quotas, per-entity limits, overage handling

### 9.3 Document Versioning

- **9.3.1** Version History — version list, author, timestamp, change notes
- **9.3.2** Version Comparison — diff viewer for text documents, side-by-side comparison
- **9.3.3** Rollback — restoring previous versions, audit trail of rollback actions

### 9.4 Metadata

- **9.4.1** Tags — free-form tags, taxonomy tags, auto-tagging rules
- **9.4.2** Categories — category hierarchy, category-based permissions
- **9.4.3** Labels — workflow labels, priority labels, custom label sets

### 9.5 Access Control

- **9.5.1** Sharing Documents — share with user, role, group, external link
- **9.5.2** Permissions — view, download, edit, delete, share permissions
- **9.5.3** Secure Access — expiring links, password-protected sharing, audit of access

### 9.6 Retention Policies

- **9.6.1** Archiving — archival rules, archive storage tier, retrieval latency
- **9.6.2** Retention Rules — legal hold, minimum retention, maximum retention
- **9.6.3** Disposal — secure deletion, disposal audit trail, compliance confirmation

---

## Section 10 — Audit & Compliance

> **Phase:** 2 · **Audience:** Security Engineers, Compliance Officers, All Engineers

### 10.1 Audit Overview

- **10.1.1** Audit Principles — immutability, completeness, non-repudiation
- **10.1.2** Compliance Requirements — SOC 2, ISO 27001, GDPR, local regulatory requirements

### 10.2 Audit Logs

- **10.2.1** User Activity Logs — login/logout, data access, data modification, exports
- **10.2.2** System Activity Logs — automated processes, integration events, background jobs
- **10.2.3** Security Logs — authentication failures, permission denials, suspicious activity

### 10.3 Change Tracking

- **10.3.1** Entity History — full before/after state for every change
- **10.3.2** Field-Level Tracking — granular tracking of individual field changes
- **10.3.3** Configuration Changes — system configuration change log, who changed what

### 10.4 Compliance Monitoring

- **10.4.1** Compliance Rules — rule definitions, evaluation schedules, violation detection
- **10.4.2** Compliance Reports — automated compliance reports, regulatory export formats

### 10.5 Investigations

- **10.5.1** Incident Reviews — incident creation, evidence collection, timeline reconstruction
- **10.5.2** Forensic Analysis — advanced log search, cross-entity correlation, export for legal

---

## Sections 11–20 (Summary)

The following sections follow the same deep structure as Sections 1–10. Full chapter expansions are available in their respective documentation packages.

| Section | Title | Phase | Key Sub-Sections |
|---|---|---|---|
| 11 | Reporting Platform | 2 | Report types, Custom builder, Scheduling, Exports (PDF/Excel/CSV) |
| 12 | Analytics & Dashboards | 2 | KPI/Chart/Table widgets, Dashboard builder, Trend/comparative/drill-down analysis |
| 13 | Integration Platform | 3 | API Management, Webhooks, Connectors (ERP, CRM, Accounting), Data Sync |
| 14 | Data Management | 3 | Import templates, Validation rules, Data quality, Data governance & stewardship |
| 15 | Search Platform | 3 | Full-text search, Advanced search, Saved searches, Faceted filters, Search analytics |
| 16 | Platform Administration | 3 | System health monitoring, Maintenance windows, Background jobs, Cache/storage management |
| 17 | Security & Governance | 3 | Encryption, Data privacy/masking, Threat detection, Risk assessments |
| 18 | Platform Operations | 3 | Backup/recovery, Disaster recovery, Business continuity, Performance/availability monitoring |
| 19 | Developer & Integration Resources | 4 | API overview, Auth, SDKs, Webhooks, Extensions, API reference, Release notes |
| 20 | Troubleshooting & Support | 4 | Common issues, Diagnostics, Error codes, Performance/security issues, Escalation, FAQ |

---

---

# PART B — Business Modules

*Each module is a self-contained functional domain that builds on the Platform Core. Modules can be licensed and activated independently.*

---

## Module 1 — Finance & Accounting

> **Phase:** 2 · **Audience:** Finance Engineers, ERP Implementers, Finance Business Users

### 1.1 Finance Overview

- **1.1.1** Introduction to Finance — scope, supported accounting standards (IFRS, GAAP, IPSAS)
- **1.1.2** Financial Architecture — multi-company, multi-currency, multi-period design
- **1.1.3** Financial Periods — period definition, opening/closing, period lock
- **1.1.4** Accounting Principles — double-entry enforcement, accrual vs. cash basis
- **1.1.5** Multi-Company Accounting — intercompany transactions, consolidation, elimination
- **1.1.6** Multi-Currency Accounting — base currency, functional currency, translation rates, FX gains/losses

### 1.2 Chart of Accounts

**Account Structure**
- Account hierarchy model, account segments, account numbering conventions
- Account types: Asset, Liability, Equity, Revenue, Expense, Statistical
- Parent/child relationships, rollup accounts, account groups

**Managing Accounts**
- Creating accounts: required fields, validation rules, effective dating
- Editing accounts: change impact analysis, history preservation
- Merging accounts: balance transfer, journal reclassification
- Deactivating accounts: outstanding balance checks, downstream impact

**Account Governance**
- Account creation policies: approval requirements, naming conventions
- Account review cycles: annual reviews, dormant account identification
- Account controls: posting restrictions, mandatory cost center, mandatory project

### 1.3 Journal Entries

**Journal Basics**
- Journal types: standard, adjusting, reversing, recurring, statistical
- Double-entry enforcement: debit/credit validation, balanced entry rules
- Journal sources: manual, system-generated, integration-sourced

**Transaction Lifecycle**
- Creating journal entries: line item entry, attachment, narration
- Draft entries: auto-save, draft review, draft approval
- Posting entries: posting validation, period check, budget check
- Reversing entries: auto-reversal setup, manual reversal, reversal linking
- Recurring entries: template setup, frequency, run schedule, exception handling

**Controls**
- Approval workflows: threshold-based routing, mandatory approval by type
- Audit trails: complete entry history, posting log, access log
- Validation rules: account combination rules, mandatory fields, intercompany rules

### 1.4 Accounts Receivable

**Customer Accounts**
- Customer ledger structure, customer credit limit management
- Customer statements: period, aging, running balance formats

**Invoicing**
- Sales invoices: creation, line items, tax calculation, PDF generation
- Credit notes: linked to invoice, partial/full credit, re-application
- Debit notes: debit note creation, customer notification

**Collections**
- Payment recording: payment types, payment reference, bank allocation
- Payment allocation: FIFO, specific invoice, partial payment handling
- Customer aging: configurable aging buckets, aging by branch/department
- Collection management: dunning letters, collection stages, promises to pay
- Write-offs: write-off approval, bad debt provision, tax implications

### 1.5 Accounts Payable

**Supplier Accounts**
- Supplier ledger structure, supplier statements, outstanding balance tracking

**Payables Processing**
- Supplier bills: bill creation, purchase order matching, tax verification
- Credit notes: supplier credit notes, allocation to bills
- Payment vouchers: voucher creation, authorization, bank linkage

**Payment Management**
- Payment runs: payment run parameters, payment selection, batch approval
- Supplier aging: aging buckets, overdue alerts, payment priority
- Outstanding balances: balance reconciliation, disputed invoices

### 1.6 Cash & Bank Management

**Bank Accounts**
- Bank account setup: bank details, account types, currency
- Bank relationships: primary bank, backup bank, inter-bank transfers

**Transactions**
- Deposits: cash deposit, cheque deposit, EFT receipt
- Withdrawals: cash withdrawal, cheque, EFT payment
- Transfers: inter-bank, inter-branch, inter-currency transfers

**Reconciliation**
- Bank reconciliation workflow: import statement, auto-match, manual match
- Statement matching: fuzzy matching rules, match confidence scoring
- Variance resolution: unmatched items, outstanding items, adjustments

### 1.7 Budget Management

**Budget Planning**
- Annual budget creation: top-down, bottom-up, iterative planning
- Department budgets: budget ownership, submission, consolidation
- Project budgets: project cost budgets, capital budgets

**Budget Control**
- Budget allocations: periodic distribution, quarterly/monthly spread
- Budget monitoring: real-time budget vs. actual, commitment tracking
- Budget adjustments: budget transfers, supplementary budgets, revision requests

**Analysis**
- Budget vs. actual reporting: variance analysis, drill-down to transactions
- Variance analysis: favorable/unfavorable classification, root cause linking

### 1.8 Fixed Assets

**Asset Setup**
- Asset categories: category hierarchy, default depreciation method
- Asset registration: asset master data, acquisition details, useful life

**Asset Lifecycle**
- Acquisition: purchase, donation, construction-in-progress capitalization
- Depreciation: straight-line, declining balance, units of production methods
- Revaluation: revaluation models, revaluation reserve accounting
- Transfer: inter-company, inter-branch, inter-department transfers
- Disposal: sale, scrapping, partial disposal, gain/loss calculation

**Asset Reporting**
- Asset register: full register with filters by category, location, status
- Depreciation reports: depreciation schedule, accumulated depreciation

### 1.9 Tax Management

**Tax Configuration**
- Tax types: VAT, withholding tax, excise duty, customs duty
- Tax rates: rate history, effective dating, jurisdiction mapping
- Tax rules: taxable/exempt rules, reverse charge, zero-rating

**Tax Processing**
- Tax calculation: automatic calculation on transactions
- Tax collection: output tax tracking, input tax claims
- Tax reporting: periodic summaries, adjustments, tax returns

**Compliance**
- VAT reports: standard-rated, zero-rated, exempt, input tax summary
- Withholding tax reports: payee listing, certificate generation, payment schedule

### 1.10 Financial Reporting

**Core Statements**
- Trial balance: pre-adjusted, post-adjusted, comparative
- General ledger: full detail, summary, account activity report
- Balance sheet: IFRS format, GAAP format, comparative periods
- Income statement: by function, by nature, departmental segmentation
- Cash flow statement: direct method, indirect method

**Analysis Reports**
- Profitability reports: product, customer, department, project profitability
- Financial ratios: liquidity, solvency, profitability, efficiency ratios
- Comparative reports: period-over-period, budget vs. actual, branch comparison

---

## Module 2 — Customer Relationship Management (CRM)

> **Phase:** 2 · **Audience:** Sales Engineers, CRM Implementers, Sales Business Users

### 2.1 CRM Overview — Philosophy, data model, integration with Sales and Finance modules

### 2.2 Lead Management

- Lead capture: web forms, email parsing, manual entry, import
- Lead sources: source tracking, source attribution, UTM parameter capture
- Lead assignment: round-robin, territory-based, skill-based routing
- Lead qualification: qualification criteria, scoring model, MQL threshold
- Lead conversion: conversion to opportunity, account, and contact creation

### 2.3 Opportunity Management

- Sales pipeline: stage definitions, stage probability, weighted forecast
- Opportunity stages: custom stage configuration, stage entry/exit criteria
- Forecasting: forecast categories, rollup logic, forecast accuracy reporting
- Win/loss analysis: loss reason capture, competitive analysis, win rate by segment

### 2.4 Customer Management

- Customer profiles: 360° view, relationship history, financial summary
- Contact management: multiple contacts per account, role/department mapping
- Relationship mapping: subsidiary relationships, referral networks

### 2.5 Activities

- Tasks: task creation, assignment, due date, completion tracking
- Meetings: calendar integration, attendee management, agenda, minutes
- Calls: call logging, call outcome, follow-up creation
- Notes: rich-text notes, note linking to any entity, searchable

### 2.6 Campaign Management

- Marketing campaigns: campaign types, budget, target audience
- Email campaigns: list management, template selection, send scheduling
- Campaign tracking: open rates, click rates, conversion tracking, ROI

### 2.7 CRM Reports

- Lead reports: lead volume, conversion funnel, lead age, source performance
- Opportunity reports: pipeline report, won/lost report, stage duration
- Pipeline reports: weighted pipeline value, stage distribution, rep performance

---

## Module 3 — Sales Management

> **Phase:** 2 · **Audience:** Sales Engineers, ERP Implementers

### 3.1–3.6 Coverage

- **Quotations** — Creation, pricing rule application, discounts (volume, customer, promotional), approval workflows for non-standard pricing
- **Sales Orders** — Order creation from quote, manual order, partial fulfillment, back-order management, order amendment workflow
- **Contracts** — Contract creation with terms, renewal management, expiration alerting, contract-based pricing enforcement
- **Pricing** — Price list management, customer-specific pricing, time-bounded promotions, multi-currency pricing
- **Sales Analytics** — Rep performance dashboards, revenue by product/customer/region, conversion rate metrics, quota attainment

---

## Module 4 — Procurement Management

> **Phase:** 2 · **Audience:** Procurement Engineers, Finance Engineers

### 4.1–4.8 Coverage

- **Supplier Management** — Supplier registration and onboarding, qualification criteria, performance scoring, classification (strategic, approved, preferred)
- **Purchase Requisitions** — Internal request creation, requestor limits, multi-level approval, budget check before approval
- **RFQ Process** — RFQ creation from requisition, supplier invitation, quote submission portal, automated comparison matrix
- **Purchase Orders** — PO creation from approved RFQ, PO amendment, PO cancellation, supplier acknowledgement
- **Goods Receipt** — Receipt against PO, partial receipt, over-receipt policy, quality inspection trigger
- **Supplier Billing** — 3-way matching (PO → GRN → Invoice), tolerance rules, exception handling, auto-approval for clean matches
- **Procurement Analytics** — Spend by category/supplier/department, savings tracking, supplier performance KPIs, compliance metrics

---

## Module 5 — Inventory Management

> **Phase:** 2 · **Audience:** Inventory Engineers, Warehouse Operations

### 5.1–5.8 Coverage

- **Product Management** — Product catalog with variants (size, color, grade), unit of measure configuration, product categorization, barcode/SKU management
- **Stock Management** — Receipts (from GRN), issues (to production/sales), inter-location transfers, adjustment journals with reason codes
- **Inventory Valuation** — FIFO, weighted average cost, standard cost methods; revaluation workflow; landed cost allocation
- **Inventory Planning** — Reorder point calculation, safety stock formulas, replenishment suggestion engine, demand-driven parameters
- **Batch & Serial Tracking** — Batch creation, batch attributes (expiry, manufacturer), serial number assignment, traceability report
- **Physical Counts** — Cycle count scheduling, blind count process, variance approval, perpetual vs. periodic inventory
- **Inventory Reports** — Stock ledger by product/location/batch, inventory valuation report, aging by lot, ABC analysis

---

## Module 6 — Warehouse Management

> **Phase:** 3 · **Audience:** Warehouse Operations Engineers

### 6.1–6.6 Coverage

- **Warehouse Structure** — Warehouse hierarchy (warehouse → zone → aisle → bay → bin), capacity configuration, bin type (bulk, rack, cold)
- **Inbound Operations** — Receiving dock management, putaway strategy (fixed, dynamic, zone-directed), quality inspection workflow
- **Outbound Operations** — Pick list generation (wave, zone, batch picking), packing workflow, dispatch confirmation, shipping documents
- **Transfers** — Internal bin-to-bin transfers, inter-warehouse transfers, transfer order workflow
- **Warehouse Analytics** — Utilization by zone, throughput analysis, pick accuracy rate, dock turnaround time

---

## Module 7 — Human Resource Management

> **Phase:** 2 · **Audience:** HR Engineers, HRIS Administrators

### 7.1–7.7 Coverage

- **Employee Management** — Employee master data, employment history, job assignments (position, grade, department), contract management
- **Recruitment** — Job requisition with approval, job posting, applicant tracking, interview scheduling, offer management, hiring
- **Attendance** — Attendance tracking (biometric, manual, mobile), shift management, timesheet submission and approval
- **Leave Management** — Leave type configuration (annual, sick, maternity, etc.), accrual rules, leave request and approval workflow, carry-forward
- **Performance Management** — Goal setting and alignment, mid-year and annual appraisal cycles, 360° feedback, calibration
- **Employee Self-Service** — Personal information updates, leave requests, payslip access, training requests, expense claims

---

## Module 8 — Payroll Management

> **Phase:** 3 · **Audience:** Payroll Engineers, Finance Engineers

### 8.1–8.5 Coverage

- **Payroll Setup** — Salary structures (basic, allowances, deductions), earning components (taxable/non-taxable), statutory deduction configuration
- **Payroll Processing** — Monthly/bi-monthly run initiation, earnings/deductions import, variance alerts, approval workflow before payment
- **Statutory Compliance** — PAYE computation (country-specific tax tables), NSSF/NHIF/pension, filing format generation for tax authorities
- **Payroll Reporting** — Payroll register, bank payment file, P9/P60 equivalent, cost allocation by department/project

---

## Module 9 — Manufacturing

> **Phase:** 3 · **Audience:** Manufacturing Engineers, Operations

### 9.1–9.7 Coverage

- **Bill of Materials** — Multi-level BOM, BOM versioning and revision control, phantom assemblies, co-products and by-products
- **Production Planning** — MPS/MRP generation, capacity planning (work center, machine, labor), planned orders
- **Work Orders** — Work order creation from planned orders, operation sequencing, routing, labor and machine time capture
- **Material Management** — Material reservation, pick list from work order, backflush vs. manual consumption
- **Quality Control** — Inspection plans per operation, in-process quality checks, SPC charts, non-conformance management
- **Production Reporting** — Production output vs. planned, OEE metrics, scrap analysis, cost per unit

---

## Module 10 — Project Management

> **Phase:** 3 · **Audience:** Project Managers, Finance Engineers

### 10.1–10.6 Coverage

- **Project Planning** — Project creation with WBS, milestone definition, deliverable tracking, Gantt chart view
- **Task Management** — Task creation, assignment, dependency mapping (FS, SS, FF, SF), critical path calculation
- **Time Tracking** — Timesheet entry against project/task, timesheet approval, utilization reporting
- **Project Finance** — Project budget (cost/revenue), committed cost tracking, EVM (Earned Value Management), project P&L
- **Project Reporting** — Progress dashboard, resource utilization, budget burn, milestone status

---

## Module 11 — Service Management

> **Phase:** 3 · **Audience:** Service Operations Engineers

### 11.1–11.5 Coverage

- **Service Requests** — Multi-channel intake (email, portal, phone), auto-classification, SLA assignment, routing
- **Work Orders** — Field technician dispatch, mobile work order completion, parts consumption, labor capture
- **SLA Management** — SLA policy definitions (response time, resolution time), real-time SLA monitoring, breach alerting
- **Service Reporting** — First call resolution rate, mean time to resolution, SLA breach rate, technician productivity

---

## Module 12 — Asset Management

> **Phase:** 3 · **Audience:** Asset Managers, Operations

### 12.1–12.5 Coverage

- **Asset Registry** — Asset master record, location tracking, responsible person, asset photo documentation
- **Maintenance** — Preventive maintenance schedules (time-based, meter-based), corrective maintenance work orders, maintenance cost tracking
- **Asset Lifecycle** — Acquisition to activation, utilization monitoring, depreciation integration with Finance, retirement and disposal
- **Asset Reporting** — Asset register export, maintenance history, MTBF analysis, total cost of ownership

---

## Module 13 — Travel Management

> **Phase:** 3 · **Audience:** Finance Engineers, HR Engineers

### 13.1–13.5 Coverage

- **Travel Requests** — Multi-leg trip requests, purpose, estimated cost, travel policy validation, approval workflow
- **Travel Booking** — Flight booking integration, hotel booking, ground transport, preferred vendor enforcement
- **Expense Claims** — Post-trip expense submission, receipt upload, per diem calculation, policy violation flagging
- **Travel Reporting** — Travel spend by department/employee/destination, policy compliance rate, vendor spend analysis

---

## Module 14 — Property Management

> **Phase:** 3 · **Audience:** Property Managers, Finance Engineers

### 14.1–14.6 Coverage

- **Property Registry** — Building/unit/facility master records, floor plans, amenities, condition ratings
- **Lease Management** — Lease agreement capture, rent escalation schedules, lease renewal workflow, IFRS 16 lease liability calculation
- **Tenant Management** — Tenant master records, occupancy tracking, vacancy management, tenant communication log
- **Billing** — Automated rent billing, utility billing (submetering), service charge allocation, billing dispute management
- **Maintenance** — Tenant maintenance requests, work order management, contractor management, maintenance cost recovery

---

## Module 15 — Forecourt Management (Fuel Station)

> **Phase:** 2 · **Audience:** Forecourt Engineers, ERP Implementers, Fuel Retail Operators

This is a specialized, high-value module with deep operational requirements.

### 15.1 Forecourt Overview

- Module scope: wet stock management, pump control, shift management, attendant tracking
- Integration points: Finance (revenue), Inventory (fuel stock), HR (attendant payroll)
- Regulatory context: weights and measures, environmental reporting, tax on fuel

### 15.2 Station Setup

- **Stations** — Station master record: name, address, operating license, tax registration
- **Pumps** — Pump master: make/model, installation date, assigned nozzles, calibration due date
- **Nozzles** — Nozzle configuration: fuel grade, side (A/B), opening meter reading
- **Tanks** — Tank master: capacity, compartments, fuel grade, tank shape (for dip chart)
- **Fuel Products** — Product definition: grade, density, tax classification, selling price

### 15.3 Wet Stock Management

- **Tank Dips** — Manual dip entry, dip-to-volume conversion using tank strapping table, dip variance
- **Deliveries** — Delivery note entry, actual vs. invoiced quantity, delivery gain/loss recording
- **Stock Reconciliation** — Theoretical stock calculation, actual stock vs. theoretical, variance by tank
- **Variance Analysis** — Variance classification (evaporation, meter error, shrinkage, theft), variance tolerance policy

### 15.4 Pump Management

- **Meter Readings** — Opening/closing meter reading capture, totalizer reset handling
- **Pump Testing** — Test sale recording, test quantity, accuracy verification
- **Calibration** — Calibration scheduling, calibration result recording, regulatory certificate tracking

### 15.5 Shift Management

- **Shift Opening** — Opening readings (tank dips, meter readings), cash float, opening checklist
- **Shift Handover** — Mid-shift handover between attendants, partial readings, handover sign-off
- **Shift Closing** — Closing readings, sales reconciliation (cash + card + credit), short/over calculation, shift report generation

### 15.6 Fuel Deliveries

- **Delivery Scheduling** — Delivery order raising, reorder level trigger, supplier notification
- **Delivery Verification** — Temperature measurement, density check, ullage confirmation before discharge
- **Loss/Gain Analysis** — Delivery loss/gain report, supplier variance claim management

### 15.7 Attendant Management

- **Attendant Assignment** — Shift assignment, pump assignment, cash float assignment
- **Performance Tracking** — Sales volume per attendant, short/over history, speed of service metrics

### 15.8 Forecourt Reports

- **Sales Reports** — Daily sales by product/pump/attendant, payment method split, price variance
- **Wet Stock Reports** — Tank dip history, delivery history, closing stock position
- **Variance Reports** — Daily/weekly/monthly wet stock variance, variance trend analysis, comparative bench

---

## Module 16 — Airline Management

> **Phase:** 3 · **Audience:** Airline Systems Engineers, Revenue Management Teams

### 16.1 Reservation Management

- **Passenger Profiles** — PNR-linked profiles, FFP linkage, document management (passport, visa)
- **Bookings** — One-way/return/multi-city booking, seat map selection, ancillary upsell
- **Ticketing** — E-ticket generation, ticket number series, reissue/refund/exchange lifecycle

### 16.2 Flight Inventory

- **Flights** — Flight schedule management, flight leg vs. segment, codeshare configuration
- **Seat Inventory** — Cabin configuration, seat map, overbooking levels per cabin
- **Fare Classes** — Fare basis codes, booking class (RBD), fare rules and conditions

### 16.3 Revenue Management

- **Demand Forecasting** — Historical demand analysis, seasonality patterns, event-driven spikes
- **Dynamic Pricing** — Bid price calculation, price ladder management, competitive fare monitoring
- **Yield Management** — Seat allocation optimization, EMSRb model, availability controls

### 16.4 Check-In Operations

- **Passenger Check-In** — Web check-in, airport check-in, mobile boarding pass
- **Boarding Passes** — Barcode/QR formats, boarding zone assignment, upgrade processing

### 16.5 Airline Reporting

- **Revenue Reports** — Revenue per available seat kilometer (RASK), yield analysis, revenue by route/cabin/fare class
- **Load Factor Analysis** — Passenger load factor, weight load factor, breakeven load factor

---

## Module 17 — Business Intelligence (BI)

> **Phase:** 4 · **Audience:** Business Analysts, Executive Users, Data Engineers

### 17.1 Analytics Overview

- **17.1.1** BI philosophy: augmenting ERP transactional data with analytical insight
- **17.1.2** Data model: pre-built star schemas for each business module
- **17.1.3** Refresh cadence: near-real-time, daily, weekly refresh options

### 17.2 Dashboards

- **Executive Dashboards** — Board-level KPI pack: revenue, margins, working capital, headcount, compliance health
- **Operational Dashboards** — Module-specific operational views with drill-down to transaction level

### 17.3 KPI Management

- **KPI Definitions** — KPI formula builder, target setting by period, responsibility assignment
- **KPI Monitoring** — KPI scorecard, RAG status, trend sparklines, alert configuration

### 17.4 Data Exploration

- **Ad Hoc Analysis** — Drag-and-drop pivot builder, no-code data exploration
- **Drill Down Analysis** — Drill from summary to detail, cross-module drill paths

### 17.5 Forecasting

- **Revenue Forecasting** — Statistical models (linear, exponential smoothing, ARIMA), driver-based forecasting
- **Demand Forecasting** — Demand signal integration, inventory-linked demand forecasts

---

---

# PART C — Server-Driven UI (SDUI) Platform

*The cross-cutting UI architecture that serves all Platform Core and Business Module interfaces. One backend, many surfaces.*

---

## Volume I — Vision & Philosophy

> **Phase:** 1 · **Audience:** All

### Chapter 01 — Introduction to AwoERP UI Platform

- **1.1** What Is AwoERP? — mission, ERP context, multi-surface ambition
- **1.2** The Problem This Platform Solves
  - **1.2.1** Traditional ERP UI Fragmentation — separate web and mobile codebases for every module
  - **1.2.2** The Multi-Platform Maintenance Burden — n modules × m platforms = unsustainable cost
  - **1.2.3** Tenant Customization Without Code Forks — customization without diverging deployments
- **1.3** Platform Mission Statement — backend as the single source of UI truth
- **1.4** Key Value Propositions
  - **1.4.1** One Backend, Many Surfaces — web, mobile, terminal, print from one definition
  - **1.4.2** Business Logic Drives UI, Not the Inverse — UI is a rendering concern only
  - **1.4.3** ERP-Native Component Vocabulary — components speak "invoice line" not "table row"
  - **1.4.4** Permission-Aware UI at the Source — pruned before delivery, not hidden client-side
- **1.5** Audience Guide — tailored reading paths for each engineering role
- **1.6** How to Use This Documentation — chapter dependencies, learning paths
- **1.7** Relationship to Other AwoERP Documentation — links to Platform Core and Business Module docs
- **1.8** Glossary Reference — links to GLOSSARY.md

### Chapter 02 — Design Philosophy

- **2.1** The Server-Driven UI Manifesto — core beliefs, non-negotiables
- **2.2** Inspirations and Prior Art — AMIS, Retool, Appsmith, PowerApps; what AwoERP takes and what it rejects
- **2.3** Core Design Tenets
  - Tenet 1 — UI as Data, Not Code
  - Tenet 2 — The Backend Is the Source of Truth
  - Tenet 3 — Clients Are Rendering Engines, Not Decision Makers
  - Tenet 4 — Components Speak Business, Not HTML
  - Tenet 5 — Permissions Are First-Class Citizens
  - Tenet 6 — Extensibility Without Forking
  - Tenet 7 — Observability Built In
- **2.4** Design Trade-offs Accepted — coupling for consistency; JSON overhead for flexibility; rendering complexity for platform control
- **2.5** Non-Goals — what this platform deliberately does not solve
- **2.6** Philosophy vs. Implementation — where rules can flex with justification

### Chapter 03 — Architectural Principles

- **3.1–3.10** Foundational principles: Separation of Concerns, AST-First, Schema-Driven Everything, Progressive Enhancement, Backward Compatibility as Contract, Fail-Safe Degradation, Least-Privilege Rendering, Composability over Configuration, Immutability of Compiled Definitions, Audit-Ability of UI Decisions
- **3.11** Layered Architecture — 5-layer model from Business Services to Client Interaction
- **3.12** Cross-Cutting Concerns — Security, Observability, Tenancy, Localization

### Chapter 04 — System Overview

- **4.1** High-Level System Diagram
- **4.2** Platform Components Inventory — Compilation Service, Component Registry, Schema Validator, Feature Flag Resolver, Authorization Resolver, Localization Service, Web Engine, Mobile Engine
- **4.3** Request Lifecycle — 7-step journey from client request to rendered UI
- **4.4** Technology Stack — Go + Goa, PostgreSQL, Redis, Temporal, OpenFGA/Casbin, Feature Flags
- **4.5** Deployment Topology — single-region, multi-region, edge caching
- **4.6** Scalability Characteristics — horizontal scaling, cache hit ratios, compilation throughput
- **4.7** Known Limitations — current constraints and planned resolution timeline

---

## Volume II — DSL & AST

> **Phase:** 1 · **Audience:** Platform Engineers, Backend Engineers, UI Framework Developers

### Chapter 05 — Server-Driven UI Fundamentals

- **5.1–5.5** SDUI concepts, spectrum from thin templates to full AST, why AwoERP chose full AST, JSON's role, comparison with React/Vue/native
- **5.6** Client Contract Guarantees — what must render, what may be enhanced, what must not be overridden
- **5.7** Schema Versioning and Client Compatibility — version negotiation, compatibility matrix
- **5.8** SDUI Security Boundaries — trust model, client-side security non-guarantees
- **5.9** SDUI Debugging Mental Model — how to reason about failures at each layer

### Chapter 06 — UI DSL Architecture

- **6.1–6.4** DSL definition, design goals, grammar overview (nodes, edges, slots, bindings, directives), type system (primitive, composite, reference, conditional, expression)
- **6.5** Binding Expressions — static, data, conditional, computed, permission bindings
- **6.6** Directives — visibility, repeat, permission, feature flag, locale directives
- **6.7–6.9** DSL extensibility, validation rules, versioning

### Chapter 07 — AST Design

- **7.1–7.2** AST definition, node taxonomy (container, leaf, control flow, data, action, layout, composite)
- **7.3** Node Schema — universal fields: id, type, props, children, slots, events, actions, permissions, flags, metadata
- **7.4** AST Construction in Go — builder pattern, typed factories, slot composition, recursive construction
- **7.5** AST Validation — structural, type, reference integrity, permission consistency, cyclic reference detection
- **7.6** AST Transformation Passes — permission pruning, feature flag resolution, localization injection, default values, tenant overrides
- **7.7–7.9** AST diffing/patching, serialization formats, canonical form

### Chapter 08 — JSON Compilation Pipeline

- **8.1** Pipeline Overview — 6-stage pipeline diagram
- **8.2** Stage 1: Context Resolution — tenant, user/role, feature flags, locale, device context
- **8.3** Stage 2: AST Construction — business service emits raw AST, slot composition, template instantiation
- **8.4** Stage 3: AST Transformation — permission pruning, flag evaluation, tenant overrides, localization, defaults
- **8.5** Stage 4: Validation — schema, business rules, security assertions
- **8.6** Stage 5: Serialization — JSON marshaling, compression, payload size budgets, field projection
- **8.7** Stage 6: Transport — REST wrapping, gRPC streaming, cache header strategy
- **8.8–8.10** Pipeline error handling, observability (spans, latency, size metrics), extension points

---

## Volume III — Component System

> **Phase:** 2 · **Audience:** Platform Engineers, Web/Mobile Engineers, UI Framework Developers

### Chapter 09 — Component System

- **9.1–9.2** Philosophy, identity and registry (namespaced type IDs, registry service, lookup, custom registration)
- **9.3** Component Category Taxonomy — 10 categories from primitives to ERP domain components
- **9.4** Component Contract — props schema, slot definitions, emitted events, supported actions, permission surface, rendering constraints
- **9.5** Component Versioning — lifecycle, breaking vs. non-breaking changes, deprecation strategy
- **9.6** Primitive Component Library — Text, Button, Icon, Image, Badge, Avatar, Spinner, Tooltip, and more
- **9.7** Compound Component Library — Card, Alert, Modal, Drawer, Stepper, Accordion, Tabs, Breadcrumb
- **9.8–9.10** Theming and style props (design tokens, tenant theming, overrides), conditional rendering, fallbacks

### Chapter 10 — Layout System

- **10.1–10.2** Philosophy, layout model (flow, grid, flex, stack, absolute, responsive)
- **10.3** Grid System — column definitions, row definitions, span/offset, gap/gutter, nested grids
- **10.4** Responsive Breakpoints — xs/sm/md/lg/xl definitions, per-breakpoint visibility and spans
- **10.5** Layout Containers — Page, Section, Panel, Split Pane, Scrollable, Sticky Header/Footer
- **10.6** Spacing System — margin/padding tokens, density modes (Compact, Normal, Comfortable)
- **10.7** Mobile Layout — single column default, safe area insets, bottom sheet layouts
- **10.8** Web Layout — sidebar + content shell, master-detail, embedded/portlet layouts
- **10.9** Layout Serialization in AST

### Chapter 11 — Forms Framework

- **11.1–11.2** Philosophy, form node types (container, section, field, label, help, error)
- **11.3** Field Input Types — 22 field types from text to multi-currency to entity picker
- **11.4** Form Layout — grid placement, multi-column, collapsible sections, tab sections, wizards
- **11.5** Dynamic Forms — conditional visibility/required, dynamic options, repeating groups, computed fields
- **11.6** Form State — initial values, dirty, touched, error, submitting states
- **11.8** Form Actions — submit, draft, reset, cancel, custom
- **11.9** Form Permissions — field-level, section-level, submit permission
- **11.10** ERP-Specific Patterns — header+line-items, approval routing fields, document attachments, audit trail in forms

### Chapter 12 — Tables and Data Grids

- **12.1–12.3** Philosophy, table node schema, 10 column types
- **12.4** Table Features — 15 features including sorting, filtering, pagination, row selection, expansion, frozen columns, inline editing, bulk actions, export
- **12.5** Table Permissions — column, row-level (FGA-driven), action column
- **12.6** ERP-Specific Patterns — line item grid, ledger/journal table, audit log table, hierarchical tree table
- **12.7–12.8** Performance considerations, mobile adaptations

### Chapter 13 — Dashboard Framework

- **13.1–13.3** Design goals, node schema, 4 layout types
- **13.4** Widget Types — 12 types including KPI, trend, chart, table, activity feed, approval pending, map/geo, calendar
- **13.5–13.9** Data sources, personalization (role-based defaults, user customization, tenant templates), permissions, refresh/polling, export

### Chapter 14 — Charts and Analytics

- **14.1–14.3** Overview, node schema, 13 chart types from bar/line/pie to Gantt, treemap, and Sankey
- **14.4** Chart Data Binding — static, DataSource reference, aggregation expressions, real-time subscriptions
- **14.5** Interactivity — click-to-action, drill-down navigation, cross-filter between charts
- **14.6–14.9** Theming, export (PNG/SVG/PDF), performance for large datasets, accessibility

### Chapter 15 — Workflow and Approval Components

- **15.1–15.2** Temporal integration philosophy, 9 workflow component types
- **15.3** Approval Actions — Approve, Reject, Request Revision, Delegate, Escalate, Withdraw
- **15.4–15.9** State machine representation, Temporal signal binding, push notification integration, pending queues, SLA timers, audit trail components

### Chapter 16 — ERP-Specific Components

- **16.2** Financial — Currency display, amount breakdown, tax summary, ledger row, journal voucher, trial balance table, P&L summary, budget vs. actual bar
- **16.3** Procurement — PO header, line item grid, vendor card, RFQ comparison table, GRN viewer
- **16.4** Inventory — Stock level indicator, bin/location selector, batch/serial panel, stock movement timeline
- **16.5** HR — Employee card, org chart, leave balance panel, payslip viewer, attendance heatmap
- **16.6** Sales — Customer card, sales order header, pipeline stage indicator, invoice viewer, credit limit warning
- **16.7** Document — PDF viewer, document status badge, e-signature capture, QR code display, barcode scanner input

---

## Volume IV — Rendering Architecture

> **Phase:** 2 · **Audience:** Mobile Engineers, Web Engineers

### Chapter 17 — Mobile Rendering Architecture

- **17.1–17.2** Philosophy, supported platforms (iOS/Swift/SwiftUI, Android/Kotlin/Compose, React Native/Flutter notes)
- **17.3** Engine Architecture — JSON Parser, Component Resolver, Rendering Tree Builder, View Reconciler, Action Dispatcher, Event Bus
- **17.4** Component-to-Native Mapping — mapping table, fallback strategy, custom native bridges
- **17.5–17.6** Layout rendering (flex to native stack), form rendering (keyboard avoidance, native pickers)
- **17.8–17.13** State management, performance (lazy loading, image caching, list virtualization), offline, accessibility, telemetry, testing

### Chapter 18 — Web Rendering Architecture

- **18.1–18.2** Philosophy, framework selection rationale, SSR considerations
- **18.3** Engine Architecture — JSON Fetcher, Schema Validator, Web Component Registry, Virtual DOM Builder, Reactive State Binding, Action Dispatcher, Event System
- **18.4** Component-to-HTML Mapping — mapping table, fallback strategy, custom web component registration
- **18.5–18.6** Layout (CSS Grid, responsive breakpoints, dark mode), form rendering (controlled inputs, file upload)
- **18.9–18.13** Performance (code splitting, memoization, virtual scrolling, progressive hydration), offline, accessibility, telemetry, testing

### Chapter 19 — Navigation Framework

- **19.1–19.2** Philosophy, 7 navigation node types
- **19.3** Navigation Item Schema — label/icon/badge, target route, permission guard, feature flag guard, active state, sub-navigation
- **19.4** Route System — definitions, parameters, deep linking, dynamic route generation from backend
- **19.5** Navigation Actions — 8 action types from Push to Navigate to External URL
- **19.6** Permission-Aware Navigation — menu item visibility, route guards, redirect on unauthorized
- **19.7–19.9** Multi-tenant navigation, state persistence, platform-specific conventions (iOS, Android, Web Router)

### Chapter 20 — Action System

- **20.1–20.2** Philosophy, action node schema (type, parameters, conditions, confirmation, permissions, feedback)
- **20.3** Built-In Action Types — 17 types including HTTP request, navigate, modal, form submit, workflow signal, download, print, custom
- **20.4** Action Chaining — sequential, conditional, parallel, error branches
- **20.5–20.8** Optimistic UI, debounce/throttle, audit trail, permission enforcement (client + server)

### Chapter 21 — Event System

- **21.1–21.2** Overview, 8 event type categories from component lifecycle to custom domain events
- **21.3** Event Node Schema — name, payload schema, handler actions, propagation settings
- **21.4** Event Bus Architecture — client-side bus, SSE for real-time, WebSocket channels, gRPC streaming
- **21.5–21.7** Event filtering/routing, cross-component communication, event logging and replay

---

## Volume V — Runtime & State

> **Phase:** 3 · **Audience:** Web Engineers, Mobile Engineers, Backend Engineers

### Chapter 22 — State Management

- **22.1–22.2** Philosophy, 7-category state taxonomy (page ephemeral through server-owned workflow state)
- **22.3** State Variable System — declaration, scoping (page/component/global), bindings, mutation via actions
- **22.4–22.7** Derived/computed state, server synchronization, persistence, debugging tools

### Chapter 23 — Validation Framework

- **23.1–23.2** Server-first philosophy, 10 validation rule types from required through async remote validation
- **23.3–23.4** Validation rule schema, execution (onChange, onBlur, onSubmit, server-side response mapping)
- **23.5–23.7** Error display (field, section, form-level), multi-step form validation, i18n

### Chapter 24 — Data Sources

- **24.1–24.2** Philosophy, data source node schema (6 fields including pagination, refresh, cache config)
- **24.3** Data Source Types — REST API, gRPC, static/inline, computed, real-time (WebSocket/SSE), file/blob
- **24.4** Transformation Layer — field mapping, type coercion, aggregation, filtering, sorting expressions
- **24.5–24.7** Lifecycle (lazy/eager, dependency, invalidation), error handling, security (auth forwarding, tenant isolation, field masking)

---

## Volume VI — Platform & API

> **Phase:** 3 · **Audience:** Platform Engineers, Backend Engineers, Security Engineers, Third-Party Integrators

### Chapter 25 — API Contracts

- REST endpoints (UI definition GET/resolve, component registry, action execution, event submission, preferences, tenant config)
- gRPC (`UIService` proto, streaming definitions, real-time event channel)
- Request/response envelopes, API versioning, authentication, rate limiting, SDK generation from Goa

### Chapter 26 — Security Model

- SDUI threat model, authentication in UI flows (JWT/OAuth2, session management, token refresh)
- UI definition security (never trust client, server-side pruning, sensitive field masking)
- Action security (signature verification, CSRF, idempotency), transport security (TLS, certificate pinning, HTTP headers)
- Data security (PII handling, client-side retention, secure logging), supply chain security

### Chapter 27 — Authorization Integration

- OpenFGA/Casbin philosophy, 7 authorization check points in the pipeline
- OpenFGA tuple model for UI permissions, authorization caching, multi-tenant context
- Dynamic permission updates, authorization audit trail, permission denial UX patterns

### Chapter 28 — Multi-Tenant Architecture

- Tenancy model, tenant resolution (subdomain, header, token)
- Tenant-specific customizations: theme, branding, component overrides, navigation, form fields
- Data isolation in payloads, configuration storage (PostgreSQL schema, Redis cache)
- Tenant onboarding UI flows, super-admin vs. tenant-admin UI separation

### Chapter 29 — Feature Flag Architecture

- 6 flag types (boolean, multi-variant, percentage, tenant-targeted, user-targeted, time-gated)
- Resolution in UI compilation pipeline, flag context object, DSL directives
- Flag-driven patterns: show/hide, variant switching, action enabling, data source swapping
- Feature flag service integration (Go SDK, caching, change propagation), testing with flags, lifecycle/cleanup

### Chapter 30 — Customization Framework

- 3-layer model (platform → tenant configuration → user preference)
- 11 customizable dimensions: layouts, labels, required/optional/visible fields, defaults, validation, navigation, branding, custom fields, actions, reports
- Storage model, admin panel UI, merge strategy, export/import, version control

### Chapter 31 — Extension and Plugin Architecture

- 6 plugin types (component, action, data source, validation, theme, navigation)
- Plugin registration/discovery (manifest schema, registry service, version compatibility)
- Isolation and sandboxing, security review process, distribution/marketplace model

---

## Volume VII — Reliability & Performance

> **Phase:** 3 · **Audience:** Platform Engineers, Web/Mobile Engineers, Backend Engineers

### Chapter 32 — Performance Optimization

- SLOs: compilation < 50ms P99; payload delivery < 100ms P95; client render < 200ms P95
- Payload optimization: compression, partial payloads, incremental diff patching, payload budgets per surface
- Backend compilation performance: benchmarks, transformation profiling, parallelization
- Client rendering: virtual lists, memoization, lazy loading, image optimization
- Network: HTTP/2 multiplexing, request coalescing, prefetching; performance regression CI detection

### Chapter 33 — Offline Support

- 3 offline modes: read-only, optimistic write (queued actions), full offline (precached process)
- UI definition caching for offline: surface selection, invalidation strategy
- Offline action queue: serialization, conflict resolution on sync, failed action handling
- Offline indicators in UI

### Chapter 34 — Caching Strategies

- 5-tier cache taxonomy: UI definition (Redis), authorization, feature flag, data source response, client-side
- Cache key design: surface + tenant + role + flags; hashing strategy
- Cache invalidation: TTL, event-driven, manual (admin); cache warming, stampede prevention, observability

### Chapter 35 — Synchronization

- UI definition sync: ETag/If-None-Match, polling vs. push
- Data sync: delta sync, conflict detection, conflict resolution (last-write-wins, merge, user-prompted)
- Real-time sync (WebSocket/SSE), background sync on mobile, sync state indicators

---

## Volume VIII — Developer Experience

> **Phase:** 4 · **Audience:** All Engineers, Product Owners

### Chapter 36 — Internationalization (i18n)
Translation keys in AST, locale resolution pipeline, fallback chain, number/currency/date formatting, pluralization, RTL layout support, locale switching at runtime, tenant-specific translations, translation management workflow

### Chapter 37 — Accessibility (a11y)
WCAG 2.1 AA + Section 508 targets, accessibility DSL props (aria, roles, labels, live regions), keyboard navigation, screen reader support, color contrast, focus management, mobile VoiceOver/TalkBack, accessibility testing

### Chapter 38 — Observability
Three-pillar philosophy (traces, metrics, logs); distributed trace propagation + per-stage spans + client contribution; key metrics (compilation latency, payload size, cache hit rate, render time, action success rate); OpenTelemetry, Prometheus, Grafana infrastructure

### Chapter 39 — Telemetry
Product analytics + operational telemetry; 5 client event types (page view, interaction, form submission, error, performance); telemetry schema, privacy/consent, pipeline, dashboards

### Chapter 40 — Logging
Structured log format, 5-category taxonomy (compilation, authorization, action, error, audit), PII scrubbing, aggregation and retention

---

## Volume IX — Engineering Excellence

> **Phase:** 4 · **Audience:** All Engineers, DevOps

### Chapter 41 — Testing Strategy
12-level testing pyramid from unit tests on AST node builders through visual regression; test data management, mocking compilation service, multi-tenant scenarios, feature flag combinations, offline scenarios

### Chapter 42 — CI/CD Considerations
Schema validation in CI, breaking change detection, visual regression in CI, performance budget enforcement, multi-platform build (Web + iOS + Android), feature flag-gated deployments, canary deployments

### Chapter 43 — Versioning Strategy
Semantic versioning for schema, compatibility matrix (forward vs. backward), API versioning (URL vs. header), component versioning, client version negotiation, multi-version support window, deprecation timeline policy

### Chapter 44 — Migration Strategy
4 migration patterns (additive, rename, restructure, breaking), client migration (forced upgrade vs. graceful degradation), data migration for UI customizations, tenant migration coordination, migration testing runbook

### Chapter 45 — Governance Model
Component proposal process, schema change RFC process, breaking change approval gates, platform stewardship, third-party component vetting, documentation standards, accessibility review gate, security review gate

---

## Volume X — Reference & Guidance

> **Phase:** 4 · **Audience:** All Engineers, Solution Architects

### Chapter 46 — Best Practices
DSL authoring, AST construction, component composition, performance, security, multi-tenant, offline, testing — 8 domains of opinionated guidance

### Chapter 47 — Anti-Patterns
10 named anti-patterns with explanation, impact, and remediation: business logic in rendering engines, client-only permission enforcement, monolithic page definitions, UI-to-DB schema coupling, giant payload syndrome, skipping validation, hardcoded tenant logic, action chains without error branches, ignoring accessibility in DSL, inconsistent component versioning

### Chapter 48 — Reference Architecture
4 deployment topologies: small tenant (single-region), enterprise multi-tenant, high-availability, disaster recovery, hybrid on-premise/cloud; Infrastructure as Code templates

### Chapter 49 — End-to-End Examples

Full worked examples, each covering backend Go code, compiled JSON output, web rendering result, and mobile rendering result:

1. **Purchase Order Form** — permission pruning, header+line items, approval routing
2. **Invoice Approval Workflow** — workflow status component, approval action bar, Temporal signal
3. **Multi-Tenant Dashboard** — Tenant A standard layout vs. Tenant B custom widgets via feature flag, diff output
4. **Dynamic ERP Report Table** — column permission pruning, aggregation row, export action
5. **Offline-Capable Stock Take Form** — offline mode indicators, action queue, sync on reconnect

### Chapter 50 — SDK Development
Go SDK (AST builder helpers, typed factories, validation utilities, pipeline hooks), TypeScript SDK (auto-generated schema types, component/action registration, custom data source adapter), iOS Swift SDK (AST decoder, native component bridge), Android Kotlin SDK (AST decoder, Composable bridge); SDK versioning, documentation, publishing (npm, CocoaPods, Maven)

### Chapter 51 — Future Roadmap
Platform maturity model; near-term (0–6 months): AST stabilization, Web/Mobile Rendering v1, ERP Component Library v1; mid-term (6–18 months): No-Code Customization UI, AI-Assisted Form Generation, Plugin Marketplace; long-term (18+ months): Voice/Conversational UI, AR/Industrial Terminal Surface, Third-Party ERP Connector Framework, AI-Driven Layout Optimization; planned breaking changes and migration windows

---

---

# Supplementary Topics (Recommended Additions)

The following chapters are recommended for inclusion across the documentation corpus. They are not yet placed in a specific volume but are high-priority gaps.

| ID | Proposed Chapter | Rationale | Suggested Volume/Part |
|---|---|---|---|
| A | **GLOSSARY** — SDUI, AST, Surface, Slot, Tenant, Branch, Cost Center, Compilation Pass, Binding, Directive, RBD, Wet Stock, Tank Dip, Nozzle, and 100+ terms | Critical for cross-team alignment; must be the first document any new engineer reads | Top-level `GLOSSARY.md` |
| B | **Error Handling Strategy** — standard error taxonomy, client error display patterns, retry policies, circuit breakers | Referenced by every action and data source chapter without a canonical definition | Vol. V, Chapter 24.5 expansion |
| C | **Developer Tooling** — CLI for local SDUI development, JSON definition validator, schema browser/explorer, hot-reload emulator | Dramatically accelerates adoption; without this, feedback loop is too slow | Vol. VIII, new Chapter 38.5 |
| D | **UI Composition Patterns** — template inheritance, slot overrides, mixin composition, partial definitions | Core reuse mechanism not fully covered by AST chapter alone | Vol. II, new Chapter 08.5 |
| E | **Audit & Compliance for UI** — GDPR considerations for UI telemetry, SOC 2 evidence from UI audit logs, ISO 27001 controls | Required for enterprise ERP customers; compliance teams need explicit UI-layer coverage | Vol. VI, new Chapter 26.5 |
| F | **Real-Time UI Updates** — live dashboard auto-refresh, push-driven form field updates (e.g. stock price in invoice), collaborative editing indicators | Distinct from the Event System chapter; needs its own treatment with WebSocket lifecycle | Vol. IV, new Chapter 21.5 |
| G | **Print and Export Rendering** — PDF generation from UI definitions, print layout DSL, picking slip templates, invoice print, report print | Core ERP requirement; invoices, delivery notes, payslips all require print-quality output | Vol. III, new Chapter 16.8 |
| H | **Theming and Design Tokens** — full token system (color, typography, spacing, radius, shadow, elevation), tenant branding pipeline, dark mode implementation, high-contrast mode | Referenced in 12+ chapters without a canonical source; this needs to be the single source | Vol. III, new Chapter 09.5 expansion |
| I | **Implementer Onboarding Guide** — step-by-step guide for new tenant deployments: org setup, module activation, user import, chart of accounts, opening balances | Essential for ERP go-lives; without it, implementation teams invent their own process | Part A supplement + Vol. X Chapter 48.5 |
| J | **Playground / Sandbox** — interactive browser-based tool for testing UI definitions without a full backend, syntax highlighting, live preview, JSON validation, component explorer | Accelerates developer onboarding from days to hours; the #1 request from implementation teams | Vol. VIII, new Chapter 38.6 |

---

## Summary Statistics

| Metric | Count |
|---|---|
| Documentation Parts | 3 (Platform Core, Business Modules, SDUI Platform) |
| SDUI Volumes | 10 |
| SDUI Core Chapters | 51 |
| Recommended Additional Chapters | 10 |
| Platform Core Sections | 20 |
| Business Modules | 17 |
| Finance Sub-Modules | 10 |
| Forecourt Sub-Sections | 8 |
| Total Top-Level Sections | 300+ |
| Estimated Pages When Written | 1,500–2,000 |
| Delivery Phases | 4 |
| Phase 1 (Foundation) | SDUI Ch. 01–08 + Platform Core Sec. 1–5 |
| Phase 2 (Components & Primary Modules) | SDUI Ch. 09–21 + Finance, CRM, HR, Forecourt |
| Phase 3 (Platform Depth & Secondary Modules) | SDUI Ch. 22–35 + remaining modules |
| Phase 4 (Excellence & Reference) | SDUI Ch. 36–51 + BI + SDK + Tooling |

---

*End of AwoERP Complete Documentation Architecture Guide — Version 1.0.0*

*This document supersedes: "AwoERP Platform & Core Modules Documentation TOC", "AwoERP Business Modules Documentation TOC", and "AwoERP Server-Driven UI Platform — Official Architecture Documentation v0.1.0-draft"*
