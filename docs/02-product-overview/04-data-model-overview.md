---
title: Data Model Overview
portal: 2 — Product Overview
section: 02-product-overview
audience: [all]
related:
  - "[Product Overview](01-product-overview.md)"
  - "[Module Overview](02-module-overview.md)"
  - "[Schema Conventions](../03-platform-architecture/03-data-architecture/02-schema-conventions.md)"
---

# Data Model Overview

High-level view of AwoERP's core entities and how they relate. For detailed column definitions see the schema docs.

## Core Entities

```
Tenant
  └── Entity (org unit: company / division / department / cost center)
        └── Employees
        └── Contracts
        └── Finance Accounts
        └── Finance Transactions

Tenant
  └── Users
        └── Roles
              └── Permissions
```

## Tenant

Every piece of business data belongs to a tenant. The `tenant_id` column is on every business table, enforced by PostgreSQL RLS.

```
tenants
  id, name, subdomain, status, plan, company_size, country
  created_at, activated_at, suspended_at, archived_at
```

**Lifecycle**: `PENDING → ACTIVE ↔ SUSPENDED → ARCHIVED`

## Entity (Org Structure)

`entities` is the flexible organizational unit table. It supports any depth of hierarchy via closure table (`hierarchy_paths`).

```
entities
  id, tenant_id, parent_id, name, code, entity_type, is_active

entity_type values:
  company, division, department, cost_center, project, location
```

Hierarchy examples:
```
Acme Corp (company)
  ├── Asia Division (division)
  │     ├── Singapore Office (location)
  │     └── Engineering (department)
  │           └── Platform Team (cost_center)
  └── Europe Division (division)
```

## Contracts

```
contracts
  id, tenant_id, contract_number, title, status
  contract_type, total_value, currency
  start_date, end_date, vendor_id, entity_id
  version, created_at, updated_at, deleted_at

contract_lines
  id, contract_id, tenant_id
  description, quantity, unit_price, amount
  line_order

Status flow: draft → under_review → approved → active → terminated
             (or)  approved → draft (returned for revision)
```

Contracts reference an `entity_id` (which org unit owns the contract) and a `vendor_id` (which entity is the counterparty).

## Finance

```
finance_account_groups
  id, tenant_id, name, account_type, display_order

finance_accounts
  id, tenant_id, account_code, account_name
  account_type, parent_id, group_id
  balance, currency, is_active

finance_transactions
  id, tenant_id, transaction_number, description
  transaction_date, status, currency, total_amount
  reference, posted_by_id

finance_transaction_entries
  id, transaction_id, tenant_id
  account_id, debit_amount, credit_amount
```

Double-entry bookkeeping: every transaction has entries where `SUM(debits) = SUM(credits)`.

`finance_account_balances` stores pre-computed running balances updated on each post.

## IAM

```
users
  id, tenant_id, email, name, password_hash, is_active
  last_login_at, created_at

roles
  id, tenant_id, name, display_name, is_system_role

user_roles
  user_id, role_id, tenant_id, assigned_at, assigned_by_id

permissions
  id, key, module, resource, action, description

role_permissions
  role_id, permission_id

user_sessions
  (Redis-backed, not Postgres table)
```

## Audit

```
audit_log
  id, tenant_id, user_id
  action, resource_type, resource_id
  before_state (jsonb), after_state (jsonb)
  ip_address, user_agent, request_id
  created_at

(append-only, no updates or deletes)
```

## Event Outbox

```
event_outbox
  id, tenant_id, topic, payload (jsonb)
  created_at, delivered_at

(relay goroutine drains to Redis Streams)
```

## Key Relationships

```
tenant ──< entity (1:many)
entity ──< entity (self-referential tree)

tenant ──< user (1:many)
user ──< user_roles ──> roles (many-many)
role ──< role_permissions ──> permissions (many-many)

tenant ──< contract (1:many)
contract ──< contract_line (1:many)
entity ──< contract (vendor or owner, optional FK)

tenant ──< finance_account (1:many)
finance_account ──< finance_account (self-referential tree)
finance_transaction ──< finance_transaction_entry (1:many)
finance_transaction_entry ──> finance_account (many:1)
```
