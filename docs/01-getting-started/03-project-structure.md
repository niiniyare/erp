---
title: Project Structure
portal: 1 — Getting Started
section: 01-getting-started
audience: [backend-engineer, frontend-engineer]
related:
  - "[Quick Start](01-quick-start.md)"
  - "[Module Anatomy](../04-backend-engineering/00-module-development-guide/01-overview/02-module-anatomy.md)"
---

# Project Structure

```
awoerp/
├── cmd/
│   └── server/
│       ├── main.go          # Entry point
│       ├── config.go        # Config struct + env parsing
│       ├── wire.go          # Wire injector (//go:build wireinject)
│       └── wire_gen.go      # Wire-generated DI code (never edit)
│
├── db/
│   ├── migration/           # SQL migration files (011001_*.up.sql)
│   ├── queries/             # SQLC query files (*.sql)
│   └── sqlc/                # SQLC-generated Go code (never edit)
│
├── internal/
│   ├── core/                # Business modules
│   │   ├── iam/             # Authentication + authorization (001)
│   │   ├── tenant/          # Tenant management (002)
│   │   ├── audit/           # Audit trail (004)
│   │   ├── notifications/   # Notification delivery (005)
│   │   ├── contracts/       # Contracts module (011)
│   │   ├── finance/         # Finance module (020)
│   │   └── hr/              # HR module (030)
│   │
│   ├── platform/            # Infrastructure / shared platform code
│   │   ├── db/              # pgxpool provider, Store interface
│   │   ├── eventbus/        # Event bus interface + implementations
│   │   ├── middleware/      # Fiber middleware (auth, authz, logging)
│   │   ├── metrics/         # Prometheus helpers
│   │   ├── redis/           # Redis client provider
│   │   └── temporal/        # Temporal client provider
│   │
│   ├── server/              # Fiber app setup, route registration
│   │   ├── app.go
│   │   ├── setup.go
│   │   └── health.go
│   │
│   ├── shared/              # Shared utilities (no business logic)
│   │   ├── errors/          # HTTP error mapping helpers
│   │   └── pagination/      # Pagination helpers
│   │
│   └── testutil/            # Shared test helpers
│       ├── db.go            # NewTestStore
│       ├── fixtures.go      # CreateTestTenant, etc.
│       ├── session.go       # TestSession, TestPrincipal
│       └── assert.go        # Domain-specific assertions
│
├── web/                     # Frontend (AMIS-based web UI)
│   ├── pages/               # HTML entry points
│   ├── schemas/             # AMIS JSON schemas
│   └── sdk/                 # AMIS SDK assets
│
├── docs/                    # This documentation
├── Makefile
├── docker-compose.yml
├── .env.example
├── go.mod
└── go.sum
```

## Module Internal Layout

Each module under `internal/core/{module}/` follows this layout:

```
{module}/
├── domain/          # Entities, value objects, errors, events
├── repository/      # Interface + SQLC adapter
├── service/         # Business logic, use cases
├── handler/         # Fiber HTTP handlers + DTOs
├── workflows/       # Temporal workflow definitions
├── activities/      # Temporal activity implementations
├── worker/          # Temporal worker registration
├── pipeline/        # Import/export pipelines
└── {module}.go      # ProviderSet, WorkerSet for Wire
```

See [Module Anatomy](../04-backend-engineering/00-module-development-guide/01-overview/02-module-anatomy.md) for detailed explanation of each layer.
