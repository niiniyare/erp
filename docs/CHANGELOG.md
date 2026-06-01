---
title: Documentation Changelog
---

# Documentation Changelog

Changes to AwoERP documentation by portal and date.

## 2026-05

### Portal 4 — Backend Engineering

**Module Development Guide (MDG) — full coverage:**

- `01-overview/` — What the guide covers, conventions, quick checklist
- `02-domain-layer/` — Domain struct, state machine design
- `03-database-layer/` — Database overview, migration cookbook
- `04-sqlc-layer/` — SQLC overview, advanced patterns
- `05-repository-layer/` — Repository overview, patterns
- `06-service-layer/` — Service overview, patterns, testing
- `07-handler-layer/` — Handler overview, handler testing
- `09-wire-registration/` — Wire registration overview
- `10-notifications/` — Notifications overview, channels
- `11-temporal-workflows/` — Temporal overview, workflow patterns reference
- `12-pipeline-integration/` — Pipeline overview, patterns reference
- `13-event-driven/` — Event-driven overview, patterns reference
- `14-api-design/` — API design overview, design checklist
- `15-audit-logging/` — Audit overview, audit implementation
- `16-middleware-chain/` — Middleware overview, CORS middleware, reference
- `17-error-handling/` — Error handling overview, error catalog
- `18-testing/` — Testing overview, integration test setup
- `19-observability/` — Observability guide, structured logging
- `20-configuration/` — Configuration overview
- `21-server-startup/` — Startup overview, graceful shutdown
- `22-deployment-checklist/` — Deployment checklist
- `23-worked-example/` — Complete 10-part worked example (Contracts module)

### Portal 3 — Platform Architecture

- `00-overview/` — System overview, architecture principles
- `01-multi-tenancy/` — Tenancy model, RLS enforcement, entity scope, feature flags
- `02-iam/` — IAM overview, session architecture, authorization model, session data
- `03-data-architecture/` — Data overview, schema conventions, migration strategy, query patterns
- `05-observability/` — Observability overview, logging, tracing, metrics, alerting
- `06-dependency-injection/` — Wire overview, provider patterns
- `07-temporal-architecture/` — Temporal overview, workflow patterns
- `08-redis-architecture/` — Redis overview
- `09-module-group-allocation/` — Module groups registry

### Portal 1 — Getting Started

- `04-common-commands.md` — Make targets, SQLC/Wire regeneration, awoctl
- `05-troubleshooting.md` — DB, Wire, Go build, Redis, Temporal issues

### Portal 2 — Product Overview

- `03-user-roles.md` — System roles, permission composition, EntityScope
- `04-data-model-overview.md` — Core tables, entity diagram

### Portal 5 — Frontend Engineering

- `03-dark-theme-guide.md` — CSS custom property override, dark mode
- `04-component-patterns.md` — Status badge, optimistic lock, monetary display, date range
- `05-schema-conventions.md` — Schema organization, auth header, pagination
- `06-mobile-considerations.md` — Responsive grid, touch targets, performance

### Portal 6 — DevOps

- `02-environment-variables.md` — All env vars, secrets management
- `03-kubernetes-deployment.md` — Deployment YAML, HPA, PDB, migration Job
- `04-cicd-pipeline.md` — CI workflow, CD workflow, Dockerfile
- `05-local-development.md` — Docker Compose, first-time setup, live reload

### Portal 7 — Security

- `02-authentication-flows.md` — Login/logout sequence, token format, attack mitigations
- `03-authorization-guide.md` — Permission format, Casbin policy, fail-closed
- `04-data-security.md` — RLS bypass risk, PII handling, SQL injection prevention
- `05-incident-response.md` — Severity classification, immediate actions, escalation

### Portal 8 — API Reference

- `04-finance-api.md` — Chart of Accounts, transactions, reports
- `05-iam-api.md` — Users, roles, permissions, sessions
- `06-tenant-api.md` — Tenant CRUD, lifecycle, configuration, feature flags

### Portal 9 — Operations

- `02-service-down.md` — Outage verification, pod status, restart steps
- `03-database-issues.md` — Pool, slow queries, lock contention, replication
- `04-high-error-rate.md` — Quantify errors, identify type, rollback decision
- `05-migration-failure.md` — Dirty state recovery, lock timeout, manual recovery
- `06-temporal-worker.md` — Stuck workflow diagnosis, manual signal, bulk failed
- `07-event-outbox.md` — Backlog check, relay restart, replay DLQ

### Root

- `GLOSSARY.md` — A-Z technical terms and acronyms
- `CHANGELOG.md` — This file
