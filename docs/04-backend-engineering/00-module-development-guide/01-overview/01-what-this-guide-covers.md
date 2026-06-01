---
title: What This Guide Covers
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Guide Conventions](02-guide-conventions.md)"
  - "[MDG Quick Checklist](03-mdg-checklist.md)"
  - "[Domain Layer](../02-domain-layer/01-domain-overview.md)"
---

# What This Guide Covers

This guide takes you from a blank directory to a production-running ERP module. Following it end-to-end produces a complete, tested, wire-registered module that handles authentication, authorisation, multi-tenancy, instrumentation, approval workflows, and full-stack UI — without omitting any layer.

## The Running Example

Every section uses a single concrete module: **Procurement Contracts** (`contracts`).

| Property | Value |
|----------|-------|
| Module key | `contracts` |
| Primary resource | `contract` |
| Child resource | `contract_line` |
| Actions | `create`, `read`, `update`, `delete`, `submit`, `approve`, `activate`, `terminate` |
| State machine | `draft → submitted → under_review → approved → active → suspended → terminated` |

Every code example, migration file, and schema in this guide is real, runnable code for this module.

## Layers Covered — In Order

| # | Layer | Guide Section |
|---|-------|---------------|
| 1 | DDD Domain Design | §02 DDD Domain Design |
| 2 | Database Schema + RLS | §03 Database Design |
| 3 | SQLC Annotated Queries | §04 SQLC Queries |
| 4 | Repository Interface + SQLC Adapter | §05 Repository Layer |
| 5 | Service Layer (business rules, state machine) | §06 Service Layer |
| 6 | Core Module Integrations (tenant, IAM, settings, feature flags, metadata) | §07 Core Module Integrations |
| 7 | Instrumentation (logging, tracing, metrics) | §08 Instrumentation |
| 8 | Audit Trail | §09 Audit Trail |
| 9 | Notifications | §10 Notifications |
| 10 | Temporal Workflows | §11 Temporal Workflows |
| 11 | Pipeline Architecture Integration | §12 Pipeline Integration |
| 12 | Event-Driven Integration | §13 Event-Driven Integration |
| 13 | REST API Design | §14 API Design |
| 14 | Fiber HTTP Handlers | §15 Fiber Handlers |
| 15 | Middleware Chain | §16 Middleware Chain |
| 16 | UI Schema Design (server-driven) | §17 UI Schema Design |
| 17 | amis Web UI Schemas | §18 amis Web Schemas |
| 18 | Flutter Mobile Schemas | §19 Flutter Mobile Schemas |
| 19 | Wire Dependency Injection Registration | §20 Wire Registration |
| 20 | Server Startup Sequence | §21 Server Startup |
| 21 | Testing Strategy | §22 Testing Guide |
| 22 | Complete Worked Example (all files) | §23 Worked Example |

## Complete File Tree After Following This Guide

```
internal/core/contracts/
├── domain/
│   ├── contract.go          # Entity, value objects, enums, state machine
│   ├── errors.go            # Domain sentinel errors
│   └── events.go            # Domain event types
├── repository/
│   ├── contract.go          # ContractRepository interface
│   └── contract_sqlc.go     # SQLC adapter (implements ContractRepository)
├── service/
│   ├── contract.go          # ContractService interface + implementation
│   ├── contract_workflow.go # Temporal workflow + activity definitions
│   └── contract_pipeline.go # Pipeline stage + hook definitions
└── contracts.go             # Module facade — Wire provider sets, public re-exports

internal/api/handlers/contracts/
├── handler.go               # Handler struct + constructor
├── routes.go                # Route registration function
├── request.go               # Request DTOs + validation
├── response.go              # Response DTOs + mapping
└── errors.go                # HTTP error mapping for this module

db/migration/
├── 011001_create_contracts.sql
├── 011002_create_contract_lines.sql
└── 011003_create_contracts_views.sql

db/queries/
└── contracts.sql            # SQLC annotated queries
```

## Technology Stack

| Concern | Technology |
|---------|-----------|
| Language | Go 1.22+ |
| HTTP framework | Fiber v2 |
| Database | PostgreSQL 15+ via pgx/v5 |
| Query generation | SQLC |
| Migrations | golang-migrate |
| Dependency injection | Google Wire (compile-time) |
| Long-running workflows | Temporal SDK v1.41.1 |
| Authorization | Casbin v2 |
| Structured logging | Zerolog |
| Distributed tracing | OpenTelemetry |
| Metrics | Prometheus |
| Web frontend | amis (Baidu) |
| Mobile frontend | Flutter <!-- PLANNED --> |

## Key Constraints

1. **Every table has `tenant_id`** — no exceptions. RLS is the hard security boundary.
2. **Every DB call uses `store.WithTenant(ctx, tenantID, ...)`** — never raw queries against the pool.
3. **Session extraction only via `c.Locals(domain.LocalsKeySession)`** — never trust URL/query params for identity.
4. **Route-level authorisation via `middlewarePkg.Authorize(cfg, "permission")`** — not duplicated inside handlers.
5. **Monetary values are `numeric(20,6)`** — never float in the DB schema.
6. **UUIDs are primary keys** — never serial integers.
7. **State transitions validated by `CanTransitionTo()`** on the domain entity.
8. **Wire runs at compile time** — after any provider change, run `make wire`.
9. **Instrumentation never fails the request** — wrap in deferred recovery if needed.
10. **`make wire` required after provider changes** — never edit `wire_gen.go` by hand.
