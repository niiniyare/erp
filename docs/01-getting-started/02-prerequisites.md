---
title: Prerequisites
portal: 1 — Getting Started
section: 01-getting-started
audience: [backend-engineer, frontend-engineer]
related:
  - "[Quick Start](01-quick-start.md)"
  - "[Project Structure](03-project-structure.md)"
---

# Prerequisites

## Required Tools

| Tool | Minimum Version | Install |
|------|----------------|---------|
| Go | 1.22 | `https://go.dev/dl/` |
| Docker | 24+ | `https://docs.docker.com/get-docker/` |
| Docker Compose | v2 (bundled with Docker Desktop) | — |
| make | System default | `apt install make` / Xcode CLT |
| git | 2.40+ | System default |

## Optional Tools

| Tool | Purpose |
|------|---------|
| `psql` | Direct PostgreSQL access for debugging |
| `redis-cli` | Redis inspection |
| `sqlc` CLI | If regenerating SQLC outside `make sqlc` |
| `wire` CLI | If running Wire outside `make wire` |
| `golangci-lint` | Local lint runs |

## Install Go Tools

```bash
# SQLC code generator
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Wire dependency injection
go install github.com/google/wire/cmd/wire@latest

# golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Environment Variables

| Variable | Required | Description |
|----------|---------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `TEST_DATABASE_URL` | For integration tests | Separate test database |
| `REDIS_URL` | Yes | Redis connection string |
| `TEMPORAL_URL` | Yes | Temporal server address |
| `HTTP_ADDR` | No | HTTP listen address (default `:8080`) |
| `LOG_LEVEL` | No | Log level: `debug`, `info`, `warn`, `error` (default `info`) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | OpenTelemetry collector endpoint |
| `OTEL_SERVICE_NAME` | No | Service name in traces (default `awoerp`) |

## Docker Compose Services

```yaml
# docker-compose.yml services used in development
services:
  postgres:   # PostgreSQL 15 on port 5432
  redis:      # Redis 7 on port 6379
  temporal:   # Temporal server on port 7233
  temporal-ui: # Temporal web UI on port 8088
```

Start only what you need:
```bash
docker compose up -d postgres redis          # minimal
docker compose up -d postgres redis temporal # with workflow engine
```
