# Chapter 1: Executive Summary

[← README](./README.md) | [Next: Why Configuration Management →](./02-why-configuration-management.md)

---

## What Is the Settings Module?

The Settings Module is the central configuration management system for the AWO ERP platform. It provides a hierarchical, multi-tenant configuration engine that governs how every other module behaves — from invoice number prefixes to overtime calculation rules to inventory valuation methods.

Rather than hardcoding business rules or scattering configuration across individual modules, the Settings Module gives administrators, developers, and business operators a single, auditable, and template-driven system to manage configuration at three levels:

```
System (Platform defaults)
  └── Tenant (Organization overrides)
        └── Entity (Branch / Location overrides)
```

The system resolves the most specific applicable value at runtime — entity beats tenant, tenant beats system — with full audit traceability.

---

## Key Capabilities

| Capability | Description |
|---|---|
| **Three-Level Inheritance** | System → Tenant → Entity resolution with automatic fallback |
| **Template-Based Deployment** | Industry and functional templates applied with one operation |
| **Bulk Operations** | Update hundreds of entities in a single batched request |
| **Full Audit Trail** | Every configuration change logged with old/new value, source, and user |
| **Type-Safe SQLC** | All queries generated from SQL with compile-time type checking |
| **RLS Enforcement** | Row-level security ensures tenants never see each other's data |
| **Event-Driven Cache** | Redis Streams propagate changes; caches invalidate automatically |

---

## Who Uses This Module?

**Configuration Administrators** manage tenant-level defaults, apply industry templates during onboarding, and perform bulk updates across branches.

**Module Developers** consume configuration via the settings service or HTTP API to drive business logic — approval limits, document prefixes, calculation methods — without hardcoding values.

**Business Operations** review configuration audit history, analyze which entities have customized which settings, and ensure compliance with organizational policies.

---

## Quick Facts

- **Base URL**: `/api/v1/settings`
- **Authentication**: JWT Bearer + `X-Tenant-ID` header
- **Storage**: PostgreSQL with JSONB for flexible config values
- **Cache**: Redis with source-tiered TTLs (system: 30 min, tenant: 10 min, entity: 5 min)
- **Audit**: `configuration_audit` table with `module_name` + `config_key_name` indexing
- **Templates**: Scope-aware (`SYSTEM` platform templates, `TENANT` custom templates)

---

[← README](./README.md) | [Next: Why Configuration Management →](./02-why-configuration-management.md)
