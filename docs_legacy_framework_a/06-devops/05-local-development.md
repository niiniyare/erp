> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Local Development Setup
portal: 6 — DevOps
section: 06-devops
audience: [backend-engineer, frontend-engineer]
related:
  - "[DevOps Overview](01-devops-overview.md)"
  - "[Quick Start](../01-getting-started/01-quick-start.md)"
  - "[Environment Variables](02-environment-variables.md)"
---

# Local Development Setup

## Docker Compose Services

`docker-compose.yml` starts all infrastructure dependencies:

```yaml
version: "3.9"
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: awoerp
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "postgres"]
      interval: 5s
      retries: 10

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      retries: 10

  temporal:
    image: temporalio/auto-setup:1.24
    environment:
      DB: sqlite
    ports:
      - "7233:7233"   # gRPC frontend
    depends_on:
      - postgres

  temporal-ui:
    image: temporalio/ui:2.26
    environment:
      TEMPORAL_ADDRESS: temporal:7233
      TEMPORAL_CORS_ORIGINS: "http://localhost:3000"
    ports:
      - "8088:8080"
    depends_on:
      - temporal

volumes:
  postgres_data:
```

Start with: `make infra-up`
Stop with: `make infra-down`
Reset (delete all data): `make infra-reset`

## Environment File

Copy `.env.example` to `.env` and fill in:

```bash
cp .env.example .env
```

`.env.example` contents:

```bash
DATABASE_URL=postgres://postgres:postgres@localhost:5432/awoerp?sslmode=disable
REDIS_URL=redis://localhost:6379/0
TEMPORAL_HOST=localhost:7233
TEMPORAL_NAMESPACE=default
SESSION_SECRET=dev-secret-not-for-production-at-least-64-chars-long-padding-here
ENV=development
LOG_LEVEL=debug
LOG_FORMAT=text
HTTP_PORT=8080
HEALTH_PORT=8081
METRICS_PORT=9090
```

`.env` is gitignored — never commit it.

## First-Time Setup

```bash
# 1. Start infrastructure
make infra-up

# 2. Wait for postgres to be ready (health check)
# The make target handles this automatically

# 3. Run migrations
make migrate-up

# 4. Generate code
make generate

# 5. Start the server
make run
```

## Live Reload

Install `air` for hot-reload during development:

```bash
go install github.com/air-verse/air@latest
```

Then:

```bash
make dev   # uses air for live reload
```

Air config (`.air.toml`) watches `*.go` files and rebuilds on change.

## Test Database

Tests use a separate DB to avoid clobbering dev data. Configure in `.env.test`:

```bash
DATABASE_URL=postgres://postgres:postgres@localhost:5432/awoerp_test?sslmode=disable
REDIS_URL=redis://localhost:6379/1   # DB 1 to separate from dev
```

Create and migrate the test DB:

```bash
createdb -h localhost -U postgres awoerp_test
DATABASE_URL=postgres://postgres:postgres@localhost:5432/awoerp_test?sslmode=disable make migrate-up
```

Or use `make db-test-setup` if defined.

## Multiple Tenants for Local Testing

The seed script creates two tenants for local development:

```bash
make db-seed
```

This creates:
- Tenant: `Acme Corp` (ID: `aaaaaaaa-0000-0000-0000-000000000001`)
- Admin user: `admin@acme.com` / `Password123!`
- Tenant: `Beta Inc` (ID: `bbbbbbbb-0000-0000-0000-000000000002`)
- Admin user: `admin@beta.com` / `Password123!`

Use these for testing cross-tenant isolation.

## Frontend Development

```bash
# Serve frontend locally (simple static server)
cd web && python3 -m http.server 3000

# Or use any static file server
npx serve web -p 3000
```

Point browser to `http://localhost:3000`. The AMIS frontend will call `http://localhost:8080/api/v1/`.

## Ports Reference

| Service | Port | URL |
|---------|------|-----|
| API server | 8080 | `http://localhost:8080/api/v1` |
| Health checks | 8081 | `http://localhost:8081/health/live` |
| Prometheus metrics | 9090 | `http://localhost:9090/metrics` |
| PostgreSQL | 5432 | `postgres://localhost:5432/awoerp` |
| Redis | 6379 | `redis://localhost:6379` |
| Temporal gRPC | 7233 | — |
| Temporal UI | 8088 | `http://localhost:8088` |
| Frontend | 3000 | `http://localhost:3000` |
