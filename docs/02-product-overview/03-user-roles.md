---
title: User Roles Reference
portal: 2 — Product Overview
section: 02-product-overview
audience: [all]
related:
  - "[Product Overview](01-product-overview.md)"
  - "[Module Overview](02-module-overview.md)"
  - "[IAM API](../08-api-reference/06-iam-api.md)"
---

# User Roles Reference

## Role Structure

Roles follow the pattern `{module}.{verb}`. Each tenant starts with a set of system-defined roles. Custom roles can be created by tenant admins.

## System Roles

### Platform Roles

| Role | Permissions | Intended for |
|------|------------|-------------|
| `platform.admin` | All platform operations including tenant management | Platform operators only |

### IAM Roles

| Role | Can do |
|------|--------|
| `iam.admin` | Create/update/delete users, assign roles, manage role definitions |
| `iam.viewer` | View users and role assignments (read-only) |

### Contracts Roles

| Role | Can do |
|------|--------|
| `contracts.viewer` | Read contracts and related data |
| `contracts.editor` | Create contracts, edit drafts, submit for review |
| `contracts.approver` | Read + approve submitted contracts |
| `contracts.manager` | Full contracts access including termination |

### Finance Roles

| Role | Can do |
|------|--------|
| `finance.viewer` | Read accounts, transactions, reports |
| `finance.accountant` | Create and post journal entries |
| `finance.manager` | Full finance access including void transactions |
| `finance.auditor` | Read-only with access to full audit trail |

## Permission Composition

A user can have multiple roles. Permissions are the union of all role permissions:

```
User: Jane Smith
Roles: [contracts.editor, finance.viewer]

Effective permissions:
  contracts.contract.create   (from contracts.editor)
  contracts.contract.update   (from contracts.editor)
  contracts.contract.submit   (from contracts.editor)
  contracts.contract.read     (from contracts.editor)
  finance.account.read        (from finance.viewer)
  finance.transaction.read    (from finance.viewer)
  finance.report.read         (from finance.viewer)
```

## Entity Scope

In addition to roles, users have an **entity scope** that restricts which organizational units' data they can see:

| Scope | Data access |
|-------|-------------|
| `all` | All data in tenant |
| `subtree:{entity_id}` | Entity + all child entities |
| `entity:{entity_id}` | Specific entity only |

Example: a regional manager with `contracts.editor` + `subtree:region-asia` can only see and edit contracts belonging to the Asia region and its sub-entities.

## Typical User Configurations

### Contracts Clerk

```
Roles: [contracts.editor]
Entity scope: entity:{department_id}
```

Creates and submits contracts for their department.

### Contracts Approver

```
Roles: [contracts.approver]
Entity scope: subtree:{division_id}
```

Approves contracts submitted by their division.

### Finance Accountant

```
Roles: [finance.accountant]
Entity scope: all
```

Posts transactions across all entities.

### Department Head

```
Roles: [contracts.manager, finance.viewer]
Entity scope: subtree:{department_id}
```

Full contracts control + finance visibility for their department.

### Internal Auditor

```
Roles: [contracts.viewer, finance.auditor]
Entity scope: all
```

Read-only access across all modules with audit trail access.

### Tenant Administrator

```
Roles: [iam.admin, contracts.manager, finance.manager]
Entity scope: all
```

Can manage users and has full business module access.

## Creating Custom Roles

Tenant admins can create custom roles via the IAM module or API:

```json
{
  "name": "procurement.reviewer",
  "display_name": "Procurement Reviewer",
  "permissions": [
    "contracts.contract.read",
    "contracts.contract.approve",
    "finance.account.read"
  ]
}
```

Custom role names must be unique within the tenant. Use `{module_prefix}.{verb}` naming convention.
