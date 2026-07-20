> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Developer Quick Start
portal: 1 — Getting Started
section: 01-getting-started
audience: [backend-engineer, frontend-engineer]
related:
  - "[Prerequisites](02-prerequisites.md)"
  - "[Project Structure](03-project-structure.md)"
---

# Developer Quick Start

Get AwoERP running locally in under 10 minutes.

## Prerequisites

- Go 1.22+
- Docker + Docker Compose
- `make`
- PostgreSQL client (`psql`) — optional, for debugging

## 1. Clone and configure

```bash
git clone <repo-url> awoerp
cd awoerp
cp .env.example .env
```

Edit `.env` — at minimum set:
```env
DATABASE_URL=postgres://awoerp:awoerp@localhost:5432/awoerp?sslmode=disable
TEST_DATABASE_URL=postgres://awoerp:awoerp@localhost:5432/awoerp_test?sslmode=disable
REDIS_URL=redis://localhost:6379
TEMPORAL_URL=localhost:7233
```

## 2. Start infrastructure

```bash
docker compose up -d postgres redis temporal
```

Wait for PostgreSQL to be ready:
```bash
docker compose logs -f postgres  # look for "database system is ready to accept connections"
```

## 3. Run migrations

```bash
make migrate-up
```

## 4. Generate code

```bash
make sqlc   # generate DB access code
make wire   # generate dependency injection
```

## 5. Run the server

Tell the user to run the server — do not run it in this session (Termux):

```bash
go run cmd/server/main.go
```

Server starts on `:8080`. Health probe on `:8081`.

```bash
curl http://localhost:8081/health/ready
# {"status":"ready"}
```

## 6. Run tests

```bash
# Unit tests (no DB required)
go test ./internal/core/.../domain/...
go test ./internal/core/.../service/...

# Integration tests (requires TEST_DATABASE_URL)
go test ./internal/core/.../repository/...

# All tests
make test
```

## Common Make Targets

| Target | Description |
|--------|-------------|
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Roll back last migration |
| `make sqlc` | Regenerate SQLC code from queries |
| `make wire` | Regenerate Wire DI code |
| `make test` | Run full test suite |
| `make test-coverage` | Run tests with coverage report |
| `make lint` | Run golangci-lint |
| `make fmt` | Format all Go files |

## Next Steps

- Read the [Architecture Principles](../03-platform-architecture/00-overview/02-architecture-principles.md)
- Follow the [Module Development Guide](../04-backend-engineering/00-module-development-guide/01-overview/01-what-this-guide-covers.md) to build a new module
- Review the [Conventions Cheatsheet](../04-backend-engineering/00-module-development-guide/01-overview/04-conventions-cheatsheet.md)
