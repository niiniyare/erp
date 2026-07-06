---
title: "Operations — Section Overview"
id: ops-000-readme
status: accepted
category: GUIDE
stability: STABLE
audience: [operators, framework-authors]
since: "1.0"
normative-level: informative
related:
  - "[Migrations](migrations.md)"
  - "[Deployment](deployment.md)"
  - "[Configuration](../12-configuration/configuration.md)"
  - "[Observability](../13-observability/observability.md)"
  - "[Awo Glossary](../GLOSSARY.md)"
---

# Operations

**Section 14 | Operations**

Operations covers the processes operators follow to keep Awo deployments healthy: database migrations, deployment procedures, and rollback strategies.

---

## Contents

| Document | ID | Purpose | Stability |
|---|---|---|---|
| [Migrations](migrations.md) | OPS-001 | Migration workflow, zero-downtime patterns, rollback | STABLE |
| [Deployment](deployment.md) | OPS-002 | Container build, Kubernetes manifest, rolling deploy, Temporal worker lifecycle | STABLE |
| [Troubleshooting](troubleshooting.md) | OPS-003 | Symptom-first guide: 503, 401, migration hangs, slow queries, RLS issues | STABLE |
| [Performance Tuning](performance-tuning.md) | OPS-004 | PostgreSQL, Redis, PgBouncer, Fiber, Temporal worker tuning | STABLE |
| [Upgrade Guide](upgrade-guide.md) | OPS-005 | Zero-downtime upgrade procedure, migration compatibility, Temporal versioning | STABLE |
| [Backup and Restore](backup-restore.md) | OPS-006 | WAL archiving, point-in-time restore, Redis backup, disaster recovery | STABLE |
| [Local Development Setup](local-development.md) | OPS-007 | Prerequisites, docker compose, migrations, hot reload, common issues | STABLE |

---

## Prerequisites

- [Configuration](../12-configuration/configuration.md) — environment variables for deployment
- [Observability](../13-observability/observability.md) — health and readiness checks used during rollout
- [Architecture Laws](../02-architecture/laws.md) — LAW-012 (migrations append-only), LAW-019 (schema identical across instances)
- [Glossary](../GLOSSARY.md) — Migration, Zero-Downtime Deploy, Rolling Deploy
