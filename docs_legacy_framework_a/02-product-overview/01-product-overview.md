> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Product Overview
portal: 2 — Product Overview
section: 02-product-overview
audience: [all]
related:
  - "[Module Map](02-module-overview.md)"
  - "[System Overview](../03-platform-architecture/00-overview/01-system-overview.md)"
---

# Product Overview

AwoERP is a cloud-native, multi-tenant Enterprise Resource Planning platform designed for mid-market organisations. It provides an integrated suite of business modules — contracts, finance, HR, procurement, inventory, and more — on a shared, tenant-isolated platform.

## Core Value Propositions

| Value | Description |
|-------|-------------|
| Multi-tenant SaaS | Single deployment serves many organisations. Each tenant's data is isolated at the database layer via PostgreSQL RLS. |
| Modular | Enable only the modules your organisation needs. Modules are independently activatable per tenant. |
| API-first | Every capability is available via REST API. The web UI and mobile apps are built on the same APIs as third-party integrations. |
| Auditable | Every data change is recorded with who changed what and when. Audit trail is immutable and exportable for compliance. |
| Extensible | Dynamic attributes, custom entity types, and configurable workflows without code changes. |

## Module Suite

### Platform Modules

| Module | Description |
|--------|-------------|
| IAM | Identity, authentication, role-based access control |
| Tenant Management | Tenant provisioning, settings, feature flags |
| Entity Hierarchy | Org structure — company, division, department, cost centre |
| Audit Trail | Immutable audit log, compliance export |
| Notifications | In-app and email notifications with per-user preferences |
| Configuration | System settings, configuration templates |

### Business Modules

| Module | Description |
|--------|-------------|
| Contracts | Contract lifecycle management — draft, review, approval, activation |
| Finance | Chart of accounts, journal entries, financial statements |
| HR | Employee records, org assignments, offboarding workflows |
| Procurement | Purchase requisitions, purchase orders, supplier management |
| Inventory | Stock management, warehouse tracking, adjustments |
| Projects | Project tracking, task management, budget control |
| Reporting | Cross-module reports, dashboards, scheduled exports |

## User Roles

| Role Level | Description |
|-----------|-------------|
| Platform Admin | Manages tenants, global config. No access to tenant data. |
| Tenant Admin | Full access within their tenant. Manages users, roles, settings. |
| Module Admin | Full access to a specific module within the tenant. |
| Module User | Access based on assigned role within the module. |
| Read-only | View access only within assigned scope. |

## Access Model

Users are scoped within a tenant to an entity in the organisational hierarchy. A finance manager scoped to "Division A" sees only Division A's financial data. Platform-level roles can have cross-entity visibility.
