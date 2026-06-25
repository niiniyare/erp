---
title: "Platform Modules — Overview & Layout Convention"
part: "Part VI — Platform Entities"
chapter: 41
section: "platform-module-overview"
related:
  - "[Chapter 2: The EntityDefinition](../part-01-foundations/02-entity-definition.md)"
  - "[Chapter 34: Platform Entities Reference](./platform-entities.md)"
  - "[Chapter 11: Database Migrations](../part-02-entity-system/11-migrations.md)"
---

# Chapter 41 — Platform Modules: Overview & Layout

> **Who should read this?** Any developer building a new platform module, extending an existing one, or trying to understand how the Awo Framework's own infrastructure is structured. If you are building a business module (Finance, HR, CRM), this chapter is also your authoritative guide — platform modules and business modules use *identical* patterns.

---

## 41.1 What Is a Platform Module?

The Awo Framework ships with a set of **platform modules** — infrastructure that every Awo deployment needs unconditionally:

| Module | What It Does |
|---|---|
| **Tenant** | Manages the top-level isolation boundary. Every business on the platform is a tenant. |
| **IAM** | Identity & Access Management. Users, roles, permissions, sessions, and authentication. |
| **Feature Flags** | On/off switches for functionality, controllable per-tenant without code changes. |
| **Settings** | Hierarchical configuration (system → tenant → branch). Invoice prefixes, tax rates, locale. |
| **Audit Log** | Permanent, tamper-evident record of every data change. Required for legal compliance. |
| **Metadata** | Custom fields — lets tenant admins add new data fields to any entity at runtime. |
| **Module Registry** | Tracks which business modules are installed and which each tenant has activated. |

These are called "platform" modules because they are built into the framework itself (`internal/platform/`). A deployment cannot remove them.

**Business modules** — Finance, HR, Inventory, CRM, Forecourt — are separate Go module repositories that import `awo.so/framework`. The platform modules are their foundation.

---

## 41.2 The Critical Insight: No Special Framework Path

> **There is no "framework mode" for platform modules.**

Platform modules are built using *exactly* the same tools as Finance or HR:

- `definition.EntityDefinition` — declares entities
- `definition.Register()` — registers them at startup
- `definition.PolicyFunc` — controls access
- `definition.HookDef` — lifecycle callbacks
- `pgstore.EntityStore` — persistence
- golang-migrate `.up.sql` / `.down.sql` — schema management

A Finance module developer reading the IAM source code will recognise every pattern. A new team member who understands how to build a Finance entity immediately understands how the User entity works.

This is deliberate. Giving platform modules a "special path" would create two sets of patterns to learn. Instead, the platform modules are the most thoroughly documented examples of how to use the framework correctly.

---

## 41.3 Module Directory Layout

Every module — platform or business — follows this directory structure:

```
internal/platform/<module>/          ← platform modules live here
    ├── <module>.go                  ← package entry point; init() calls definition.Register()
    ├── definition.go                ← EntityDefinition variable declarations
    ├── policy.go                    ← PolicyFunc implementations
    ├── hooks.go                     ← HookDef function implementations
    ├── service.go                   ← thin service layer on top of EntityStore
    ├── handler.go                   ← custom HTTP handlers (if needed beyond CRUD)
    └── migrations/
        ├── 20240101000001_create_users.up.sql
        ├── 20240101000001_create_users.down.sql
        └── ...

awo.so/module/<name>/                ← business modules live in separate repos
    ├── <name>.go
    ├── definition.go
    ├── policy.go
    ├── hooks.go
    ├── service.go
    └── migrations/
        └── ...
```

### Why This Layout?

**One file per concern**: `definition.go` is the source of truth for "what is this entity?". `policy.go` is where access rules live. `hooks.go` is where lifecycle logic lives. When debugging an access issue, you open `policy.go`. When debugging unexpected data changes, you open `hooks.go`. You never hunt through a 2000-line file.

**Migrations next to the code that owns them**: the IAM module owns its users/sessions/roles tables. Its migrations are in `internal/platform/iam/migrations/`, not in a central `db/migrations/` folder. This means IAM can be versioned, tested, and deployed independently. The framework knows where to find migrations by scanning registered module paths.

---

## 41.4 The Registration Pattern

Every module registers its EntityDefinitions in an `init()` function. Go guarantees `init()` runs exactly once, before `main()`, in dependency order.

```go
// internal/platform/iam/iam.go
package iam

import "awo.so/framework/definition"

func init() {
    definition.Register(&UserDef)
    definition.Register(&RoleDef)
    definition.Register(&UserRoleDef)
    definition.Register(&SessionDef)
}
```

`definition.Register` panics immediately if the definition is invalid — duplicate name, unknown field type, or empty edge target. This is intentional: a misconfigured EntityDefinition is a programming error, not a runtime error. It should fail loud and fast at startup, not silently at request time.

The framework calls `definition.All()` at startup (inside `bootstrap.Mount`) to get the complete list of every registered EntityDefinition and:
- Mounts REST CRUD routes for each entity
- Generates SDUI navigation and form schemas
- Validates that migration files exist for each entity's table

---

## 41.5 Startup Wire-Up: Import Order Matters

Platform modules must be imported before business modules, because business modules reference platform entity types (User, OrgNode) in their `EdgeDef` definitions.

```go
// cmd/server/main.go
package main

import (
    "github.com/gofiber/fiber/v2"
    "awo.so/framework/bootstrap"

    // ── Platform modules ── (registered first)
    _ "awo.so/internal/platform/tenant"
    _ "awo.so/internal/platform/iam"
    _ "awo.so/internal/platform/featureflag"
    _ "awo.so/internal/platform/settings"
    _ "awo.so/internal/platform/audit"
    _ "awo.so/internal/platform/metadata"
    _ "awo.so/internal/platform/registry"

    // ── Business modules ── (registered after platform)
    _ "awo.so/module/finance"
    _ "awo.so/module/hr"
    _ "awo.so/module/inventory"
    _ "awo.so/module/crm"
    _ "awo.so/module/forecourt"
)

func main() {
    app := fiber.New()

    bootstrap.Mount(app, bootstrap.Options{
        DSN:           mustEnv("DATABASE_URL"),
        RedisAddr:     mustEnv("REDIS_ADDR"),
        TemporalAddr:  mustEnv("TEMPORAL_ADDR"),
    })

    log.Fatal(app.Listen(":8080"))
}
```

`bootstrap.Mount` reads `definition.All()` and builds the complete server in a single pass. Adding a new module is as simple as adding one blank import line and recompiling.

---

## 41.6 Platform Module vs Business Module: The Comparison

| Concern | Platform Module | Business Module |
|---|---|---|
| **Go package** | `awo.so/internal/platform/<name>/` | Separate repo: `awo.so/module/<name>/` |
| **Imported in** | `cmd/server/main.go` (always) | `cmd/server/main.go` (deployment choice) |
| **EntityDefinition** | `definition.EntityDefinition{}` | `definition.EntityDefinition{}` — identical struct |
| **Registration** | `definition.Register(&Def)` in `init()` | `definition.Register(&Def)` in `init()` — identical |
| **OrgScope** | Global (catalogues) or Tenant | Usually Tenant or Unit |
| **Who writes policies** | Platform team | Module team |
| **Hooks** | Cache invalidation, Casbin reload | Business rule validation, workflow triggers |
| **Migrations** | `internal/platform/<name>/migrations/` | `migrations/` in module repo |
| **Service layer** | Thin wrapper + domain operations (login, evaluation) | Thin wrapper + domain operations (submit invoice, approve leave) |
| **HTTP API** | Auto-generated CRUD + custom endpoints | Auto-generated CRUD + custom endpoints |
| **SDUI** | Auto-generated AMIS pages | Auto-generated AMIS pages |
| **Removable?** | No — built into every deployment | Yes — optional per deployment |

The framework cannot distinguish, at runtime, whether an EntityDefinition came from `internal/platform/iam` or `awo.so/module/finance`. Both are Go values in the same in-memory registry. The policies and scoping rules create the distinction — not any special framework mechanism.

---

## 41.7 Quick-Reference: Module Chapters

Each platform module is documented in its own chapter:

| Module | Chapter | File |
|---|---|---|
| Tenant & Organisation | [Chapter 42](./platform-tenant-module.md) | `platform-tenant-module.md` |
| IAM — Users, Roles, Sessions | [Chapter 43](./platform-iam-module.md) | `platform-iam-module.md` |
| Feature Flags | [Chapter 44](./platform-feature-flags-module.md) | `platform-feature-flags-module.md` |
| Settings & Configuration | [Chapter 45](./platform-settings-module.md) | `platform-settings-module.md` |
| Audit Log | [Chapter 46](./platform-audit-module.md) | `platform-audit-module.md` |
| Metadata & Custom Fields | [Chapter 47](./platform-metadata-module.md) | `platform-metadata-module.md` |
| Plugin & Module Registry | [Chapter 48](./platform-module-registry.md) | `platform-module-registry.md` |

Each chapter is self-contained. Read the chapter for the module you are working with. If you are new to the framework, read [Chapter 42 (Tenant)](./platform-tenant-module.md) first — it establishes the mental model that all other modules build on.
