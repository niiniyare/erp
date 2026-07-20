> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: System Overview
portal: 3 — Platform Architecture
section: 00-overview
audience: [architect, backend-engineer, tech-lead]
related:
  - "[Architecture Principles](02-architecture-principles.md)"
  - "[Tenancy Model](../01-multi-tenancy/01-tenancy-model.md)"
---

# System Overview

AwoERP is a multi-tenant ERP platform built as a modular monolith. All modules run in a single deployable binary, share a PostgreSQL database, and communicate via an in-process event bus. Modules are decoupled by interface boundaries — not by network.

## High-Level Diagram

```
┌─────────────────────────────────────────────────────────────┐
│  Clients                                                     │
│  ┌──────────┐  ┌────────────┐  ┌───────────────────────┐  │
│  │ Web (AMIS)│  │ Mobile App │  │ Third-party / API Key │  │
│  └────┬─────┘  └─────┬──────┘  └──────────┬────────────┘  │
└───────┼──────────────┼─────────────────────┼────────────────┘
        │              │                     │
        ▼              ▼                     ▼
┌───────────────────────────────────────────────────────────┐
│  Fiber HTTP Server (:8080)                                 │
│  ┌───────────────────────────────────────────────────┐    │
│  │  Middleware Chain                                  │    │
│  │  RequestID → Logger → Recover → RateLimit → Auth  │    │
│  └───────────────────────────────────────────────────┘    │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────────┐   │
│  │ Contracts│ │ Finance  │ │   HR     │ │    IAM     │   │
│  │ Handler  │ │ Handler  │ │ Handler  │ │  Handler   │   │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └─────┬──────┘   │
│       │            │            │              │           │
│  ┌────▼─────┐ ┌────▼─────┐ ┌───▼──────┐ ┌────▼──────┐   │
│  │ Contracts│ │ Finance  │ │   HR     │ │    IAM    │   │
│  │ Service  │ │ Service  │ │ Service  │ │  Service  │   │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └─────┬──────┘   │
│       │            │            │              │           │
│  ┌────▼────────────▼────────────▼──────────────▼──────┐   │
│  │              Repository Layer (SQLC)                │   │
│  └────────────────────────┬────────────────────────────┘   │
└───────────────────────────┼────────────────────────────────┘
                            │
              ┌─────────────┼─────────────┐
              ▼             ▼             ▼
        PostgreSQL 15+   Redis        Temporal
        (primary store)  (sessions,   (workflows)
                          cache)
```

## Technology Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.22+ |
| HTTP Framework | Fiber v2 |
| Database | PostgreSQL 15+ |
| ORM / Queries | SQLC (compile-time generated) |
| DB Driver | pgx/v5 |
| Migrations | golang-migrate |
| Session Store | Redis |
| Workflow Engine | Temporal |
| DI | Google Wire (compile-time) |
| Authorization | Casbin v2 |
| Logging | Zerolog |
| Tracing | OpenTelemetry |
| Metrics | Prometheus |
| Monetary | shopspring/decimal |
| UUID | google/uuid |

## Deployment Topology

```
┌─────────────────────────────────────────────┐
│  Kubernetes Cluster                          │
│                                             │
│  ┌──────────────────┐  ┌─────────────────┐  │
│  │  awoerp Pod x3   │  │  worker Pod x2  │  │
│  │  (HTTP server)   │  │  (Temporal)     │  │
│  └──────────────────┘  └─────────────────┘  │
│                                             │
│  ┌──────────────────┐  ┌─────────────────┐  │
│  │  PostgreSQL      │  │  Redis Cluster  │  │
│  │  (managed RDS)   │  │  (managed)      │  │
│  └──────────────────┘  └─────────────────┘  │
│                                             │
│  ┌──────────────────┐                       │
│  │  Temporal Server │                       │
│  │  (managed)       │                       │
│  └──────────────────┘                       │
└─────────────────────────────────────────────┘
```

## Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Modular monolith vs. microservices | Monolith | Simpler ops, shared DB transactions, no network latency between modules |
| ORM vs. raw SQL | SQLC (raw SQL + generated types) | Full SQL control, type safety, no N+1 surprises |
| Event bus transport | In-process (pluggable) | No infrastructure dependency for development; swap to NATS/Kafka for production |
| Authorization | Casbin RBAC | Flexible policy model, tenant-scoped, auditable |
| Multi-tenancy | Row-level security (PostgreSQL) | Database-enforced isolation, application layer can't bypass |
