---
title: Module Map
portal: 3 — Platform Architecture
section: 00-overview
audience: [architect, backend-engineer, tech-lead]
related:
  - "[System Overview](01-system-overview.md)"
  - "[Architecture Principles](02-architecture-principles.md)"
---

# Module Map

Each module owns a bounded context, a database group prefix, and a permission namespace.

## Module Registry

| Module Key | Group | Package | DB Prefix | Permission Prefix |
|-----------|-------|---------|-----------|------------------|
| `iam` | 001 | `internal/core/iam` | `001` | `iam.` |
| `tenant` | 002 | `internal/core/tenant` | `002` | `tenant.` |
| `metadata` | 003 | `internal/core/metadata` | `003` | `metadata.` |
| `audit` | 004 | `internal/core/audit` | `004` | `audit.` |
| `notifications` | 005 | `internal/core/notifications` | `005` | `notifications.` |
| `config` | 006 | `internal/core/config` | `006` | `config.` |
| `entities` | 010 | `internal/core/entities` | `010` | `entities.` |
| `contracts` | 011 | `internal/core/contracts` | `011` | `contracts.` |
| `finance` | 020 | `internal/core/finance` | `020` | `finance.` |
| `hr` | 030 | `internal/core/hr` | `030` | `hr.` |
| `inventory` | 040 | `internal/core/inventory` | `040` | `inventory.` |
| `procurement` | 050 | `internal/core/procurement` | `050` | `procurement.` |
| `projects` | 060 | `internal/core/projects` | `060` | `projects.` |
| `reporting` | 070 | `internal/core/reporting` | `070` | `reporting.` |

## Platform Modules (001–009)

These modules provide infrastructure to all business modules. They are never imported by name — only via their interfaces.

| Module | Responsibility |
|--------|---------------|
| `iam` | Authentication, session resolution, Casbin authorization |
| `tenant` | Tenant lifecycle, settings, feature flags |
| `metadata` | Entity hierarchy, dynamic attributes, entity types |
| `audit` | Audit event storage, query, compliance export |
| `notifications` | In-app and email notification delivery |
| `config` | Configuration templates, tenant-level overrides |

## Business Modules (010+)

Business modules implement ERP domain logic. Each follows the Module Development Guide pattern (§04 portal).

## Module Dependencies (allowed directions)

```
Business Modules (011+)
    │
    ▼  (via interfaces only)
Platform Modules (001–009)
    │
    ▼  (via interfaces only)
Shared Libraries (internal/shared/*)
    │
    ▼
Standard Library + Third-party packages
```

**Prohibited**: Business module A importing business module B's internal types. Cross-module communication uses events or platform service interfaces.

## Group Number Allocation

Groups are allocated in blocks to leave room for expansion:

| Range | Purpose |
|-------|---------|
| 001–009 | Platform/infrastructure modules |
| 010–019 | Entity and organization management |
| 020–029 | Finance |
| 030–039 | HR and Payroll |
| 040–049 | Inventory and Warehouse |
| 050–059 | Procurement |
| 060–069 | Projects and PMO |
| 070–079 | Reporting and Analytics |
| 080–089 | CRM |
| 090–099 | Reserved |
