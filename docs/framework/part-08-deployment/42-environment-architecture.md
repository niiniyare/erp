---
title: "Chapter 42: Environment Architecture"
part: "Part VIII — Deployment and Operations"
chapter: 42
section: "42-environment-architecture"
related:
  - "[Chapter 43: Docker Containerisation](43-docker-containerisation.md)"
  - "[Chapter 44: CI/CD Pipeline](44-cicd-pipeline.md)"
---

# Chapter 42: Environment Architecture

Awo runs across three environments: local development, staging, and production. Each has different infrastructure requirements, security posture, and operational constraints. This chapter defines what each environment looks like, what differs between them, and why.

---

## 42.1. Environment Overview

| Aspect | Local | Staging | Production |
|---|---|---|---|
| Purpose | Development, debugging | Integration testing, UAT | Live customer traffic |
| Tenants | 1-3 test tenants | Mirrors of production tenants | All customer tenants |
| Infrastructure | Docker Compose | Kubernetes (small) | Kubernetes (HA) |
| Database | Single PostgreSQL | Single PostgreSQL | HA PostgreSQL (primary + replicas) |
| Redis | Single instance | Single instance | Redis Cluster |
| Temporal | Single node | Single node | Temporal Cloud or HA cluster |
| TLS | Self-signed or none | Let's Encrypt | Let's Encrypt via cert-manager |
| Secrets | `.env` files | Kubernetes Secrets | Vault or Kubernetes Secrets |
| Observability | Logs to stdout | Logs + traces + metrics | Full stack (Loki + Tempo + Prometheus) |
| Backups | None | Daily | Continuous WAL archiving |

---

## 42.2. Local Development Environment

### 42.2.1. Docker Compose Stack

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: awo
      POSTGRES_PASSWORD: awo_dev
      POSTGRES_DB: awo
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  temporal:
    image: temporalio/auto-setup:1.24
    environment:
      - DB=postgresql
      - DB_PORT=5432
      - POSTGRES_USER=awo
      - POSTGRES_PWD=awo_dev
      - POSTGRES_SEEDS=postgres
    ports:
      - "7233:7233"

  temporal-ui:
    image: temporalio/ui:2.26
    environment:
      - TEMPORAL_ADDRESS=temporal:7233
    ports:
      - "8080:8080"

volumes:
  postgres_data:
```

### 42.2.2. Environment Variables for Local

```bash
# .env.local — not committed to git
AWO_ENV=development
AWO_PORT=3000

# Database
AWO_DB_URL=postgres://awo:awo_dev@localhost:5432/awo?sslmode=disable

# Redis
AWO_REDIS_URL=redis://localhost:6379

# Temporal
AWO_TEMPORAL_HOST=localhost:7233
AWO_TEMPORAL_NAMESPACE=default

# Security (dev only — do not use in production)
AWO_SESSION_SECRET=dev-session-secret-not-for-production
AWO_JWT_SECRET=dev-jwt-secret-not-for-production

# Feature flags
AWO_FEATURE_FLAG_OVERRIDE=true  # bypass flag cache for instant flag changes

# CORS (allow the Vite dev server)
AWO_CORS_ORIGINS=http://localhost:5173,http://localhost:3000
```

### 42.2.3. Hot Reload Setup

The server binary is built and run with `air` for hot reload during development:

```toml
# .air.toml
[build]
  cmd = "go build -o ./tmp/awo ./cmd/server"
  bin = "./tmp/awo"
  include_ext = ["go"]
  exclude_dir = ["tmp", "vendor", "web"]

[log]
  main_only = true
```

The `web/` directory is excluded — frontend assets are served by Vite's dev server with its own hot reload.

---

## 42.3. Staging Environment

### 42.3.1. Purpose and Isolation

Staging is an integration environment that mirrors production as closely as possible:
- Same container images as production (built from the same CI pipeline)
- Same database schema (migrations run on staging first)
- Anonymised copy of production data (PII scrubbed)
- External integrations use sandbox/test endpoints (M-PESA sandbox, SMS sandbox)

### 42.3.2. Staging Configuration Differences

```bash
# Staging overrides
AWO_ENV=staging
AWO_DB_URL=postgres://awo:${STAGING_DB_PASSWORD}@staging-postgres:5432/awo
AWO_MPESA_ENVIRONMENT=sandbox
AWO_SMS_SANDBOX=true
AWO_LOG_LEVEL=debug      # more verbose for debugging
AWO_FEATURE_FLAG_OVERRIDE=false
```

### 42.3.3. Data Anonymisation for Staging

A weekly job copies production data to staging with PII scrubbed:

```bash
#!/bin/bash
# scripts/refresh_staging.sh

# 1. Take production snapshot
pg_dump --schema=public $PROD_DB_URL > /tmp/platform.sql

# 2. For each tenant schema, dump + anonymise
for schema in $(psql $PROD_DB_URL -t -c "SELECT schema_name FROM information_schema.schemata WHERE schema_name LIKE 'tenant_%'"); do
    pg_dump --schema=$schema $PROD_DB_URL | \
        python3 scripts/anonymise_dump.py | \
        psql $STAGING_DB_URL
done
```

The `anonymise_dump.py` script runs SQL UPDATEs after the schema is restored:
- `UPDATE employees SET national_id = 'XXXXXXXXX', kra_pin = 'XXXXXXXXX'`
- `UPDATE customers SET email = 'anon_' || id || '@staging.local', phone = '0700000000'`
- `UPDATE bank_accounts SET account_number = '00000000000'`

---

## 42.4. Production Environment

### 42.4.1. Kubernetes Topology

```
                        Internet
                           │
                    ┌──────┴──────┐
                    │   Ingress   │ (nginx/Caddy)
                    │  TLS Term.  │
                    └──────┬──────┘
                           │
            ┌──────────────┼──────────────┐
            │              │              │
       ┌────▼────┐   ┌─────▼───┐   ┌─────▼────┐
       │  API    │   │  API    │   │  API     │
       │ Pod 1   │   │ Pod 2   │   │  Pod 3   │
       └────┬────┘   └─────┬───┘   └─────┬────┘
            │              │              │
            └──────────────┼──────────────┘
                           │
               ┌───────────┼────────────┐
               │           │            │
        ┌──────▼──┐  ┌─────▼──┐  ┌─────▼──┐
        │Temporal │  │  Redis │  │ Postgres│
        │Worker 1 │  │Cluster │  │Primary  │
        └─────────┘  └────────┘  └─────┬──┘
                                        │
                                   ┌────▼───┐
                                   │Postgres│
                                   │Replica │
                                   └────────┘
```

### 42.4.2. Resource Allocation

| Component | Replicas | CPU request | Memory request | Notes |
|---|---|---|---|---|
| API server | 3 | 500m | 512Mi | HPA: scale to 10 on CPU >70% |
| Temporal worker | 2 | 1000m | 1Gi | Scale manually based on workflow volume |
| PostgreSQL primary | 1 | 2000m | 4Gi | PVC: 100Gi SSD |
| PostgreSQL replica | 1 | 1000m | 2Gi | Read-only, for reporting queries |
| Redis | 3 | 500m | 1Gi | Redis Cluster mode |

### 42.4.3. Production Environment Variables (Secrets)

Production secrets are stored in Kubernetes Secrets, mounted as environment variables. Never stored in ConfigMaps or committed to git:

```yaml
# k8s/secrets/awo-secrets.yaml (applied via kubectl, not committed)
apiVersion: v1
kind: Secret
metadata:
  name: awo-secrets
  namespace: awo-prod
type: Opaque
stringData:
  AWO_DB_URL: "postgres://awo:${PROD_DB_PASSWORD}@postgres-primary:5432/awo"
  AWO_REDIS_URL: "redis://:${REDIS_PASSWORD}@redis-cluster:6379"
  AWO_SESSION_SECRET: "${SESSION_SECRET_32_BYTES_HEX}"
  AWO_MPESA_CONSUMER_KEY: "${MPESA_KEY}"
  AWO_MPESA_CONSUMER_SECRET: "${MPESA_SECRET}"
```

### 42.4.4. Horizontal Pod Autoscaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: awo-api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: awo-api
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

**Why minimum 3 replicas?** One replica can handle a rolling deployment with zero downtime (one pod is taken down, two remain). Three replicas also spread load across availability zones (if the cluster spans AZs) for resilience.
