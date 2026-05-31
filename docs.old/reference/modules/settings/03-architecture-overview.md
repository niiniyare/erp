# Chapter 3: Architecture Overview

[← Why Configuration Management](./02-why-configuration-management.md) | [Next: Configuration Hierarchy →](./04-configuration-hierarchy.md)

---

## Clean Architecture

The Settings Module follows Clean Architecture / Domain-Driven Design principles with strict layer separation:

```
┌─────────────────────────────────────────────┐
│              HTTP Handlers                  │  ← transport/http
│         (request parsing, routing)          │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│              Service Layer                  │  ← core/settings/service
│   (business logic, validation, events)      │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│            Repository Layer                 │  ← core/settings/repository
│    (data access, SQLC queries, caching)     │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│             Database Layer                  │  ← db/sqlc, db/queries
│   (PostgreSQL, RLS, JSONB, triggers)        │
└─────────────────────────────────────────────┘
```

---

## Domain Model

### Core Aggregates

**ConfigDefinition** — The schema for a configuration key. Defines what is configurable, what type it is, its default value, and any validation rules.

```go
type ConfigDefinition struct {
    ModuleName          string
    ConfigKey           string
    ConfigType          string       // "string", "integer", "boolean", "decimal", "json"
    DefaultValue        []byte       // JSONB system default
    ValidationRules     []byte       // JSONB validation schema
    Description         string
    RequiredFeatureFlag pgtype.Text
}
```

**Configuration** — A resolved configuration value with its source and metadata.

```go
type Configuration struct {
    Module    ModuleName
    Key       ConfigKey
    Value     ConfigValue
    Source    ConfigSource   // "system" | "tenant" | "entity"
    IsInherited bool
}
```

**ConfigTemplate** — A named, versioned bundle of configuration values that can be applied to a tenant or entity.

```go
type ConfigTemplate struct {
    ID                  uuid.UUID
    Scope               string       // "SYSTEM" | "TENANT"
    TenantID            *uuid.UUID   // nil for SYSTEM templates
    Name                string
    Category            string
    Configurations      []byte       // JSONB map of module.key → value
    ConflictResolution  string       // "merge" | "replace" | "preserve"
    Version             string
}
```

---

## Infrastructure Architecture

```
┌─────────────────────────────────────────────────────────┐
│              External Configuration Sources             │
│    Environment Variables, CI/CD, Admin UI               │
└─────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────┐
│                   Settings API Gateway                  │
│    Request Routing, Authentication, Rate Limiting       │
└─────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────┐
│                Configuration Event Bus                  │
│    Redis Streams, Cache Invalidation, Notifications     │
└─────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────┐
│                  Settings Module Core                   │
│    Resolution, Templates, Validation, Audit             │
└─────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────┐
│               Consuming ERP Modules                     │
│    Finance, HR, Inventory, Sales, CRM, Projects         │
└─────────────────────────────────────────────────────────┘
```

---

## Database Tables

| Table | Purpose |
|---|---|
| `config_definitions` | Schema registry — defines all configurable keys |
| `tenant_configurations` | Tenant-level JSONB settings blob |
| `entities` | Entity-level `settings` JSONB column (shared with entity module) |
| `configuration_templates` | Versioned template bundles (SYSTEM + TENANT scope) |
| `template_applications` | Audit log of template applications |
| `configuration_audit` | Full change history with old/new values |

---

## SQLC Type Safety

All database queries are defined in `db/queries/settings.sql` and compiled by `sqlc generate` into strongly-typed Go functions in `db/sqlc/settings.sql.go`. No string-based query building — compile errors catch query mismatches before deployment.

---

[← Why Configuration Management](./02-why-configuration-management.md) | [Next: Configuration Hierarchy →](./04-configuration-hierarchy.md)
