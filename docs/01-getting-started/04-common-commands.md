---
title: Common Commands
portal: 1 — Getting Started
section: 01-getting-started
audience: [backend-engineer, frontend-engineer]
related:
  - "[Quick Start](01-quick-start.md)"
  - "[Project Structure](03-project-structure.md)"
  - "[Prerequisites](02-prerequisites.md)"
---

# Common Commands

## Make Targets

Run `make help` to see all available targets.

### Development

| Command | What it does |
|---------|-------------|
| `make run` | Start API server (loads `.env`) |
| `make run-worker` | Start Temporal workers only |
| `make dev` | Run with live reload (uses `air`) |

### Code Generation

| Command | What it does |
|---------|-------------|
| `make generate` | Run all generators (wire + sqlc) |
| `make wire` | Run Wire dependency injection codegen |
| `make sqlc` | Run SQLC query codegen |

Run `make generate` after:
- Adding a new Wire provider
- Changing a `wire.go` ProviderSet
- Adding a new SQL query to `db/queries/`
- Modifying an existing SQL query

### Database

| Command | What it does |
|---------|-------------|
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Roll back last migration |
| `make migrate-status` | Show applied vs pending migrations |
| `make migrate-create name=add_foo_column` | Create a new migration file pair |
| `make db-reset` | Drop and recreate DB, run all migrations |

### Testing

| Command | What it does |
|---------|-------------|
| `make test` | Run all tests with race detector |
| `make test-unit` | Unit tests only (no DB) |
| `make test-integration` | Integration tests (requires DB) |
| `make test-cover` | Test with coverage report |

### Infrastructure

| Command | What it does |
|---------|-------------|
| `make infra-up` | Start Docker Compose services |
| `make infra-down` | Stop Docker Compose services |
| `make infra-reset` | Stop + remove volumes + restart |

## SQLC Generation

After adding or modifying `db/queries/*.sql`:

```bash
make sqlc
```

This regenerates `internal/shared/db/sqlc/*.go`. Never edit the generated files directly.

If SQLC errors on a new query:
1. Check column names match the schema exactly
2. Verify the table exists in migrations
3. Run `make migrate-up` first if you added a new migration

## Wire Generation

After modifying any `wire.go` file:

```bash
make wire
```

This regenerates `cmd/server/wire_gen.go`. If Wire errors:
- `no provider for X` — add the provider to the relevant ProviderSet
- `cycle detected` — restructure dependencies, extract a shared type

## Running Tests

```bash
# All tests
make test

# Single package
go test ./internal/core/contracts/... -v

# Single test function
go test ./internal/core/contracts/... -run TestContractService_Create -v

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Tests require `DATABASE_URL` and `REDIS_URL` environment variables. Copy `.env.test.example` to `.env.test` and fill in.

## Database Inspection

```bash
# Connect to local DB
psql "$DATABASE_URL"

# Show all tables
\dt

# Show table schema
\d contracts

# Check RLS policies
SELECT * FROM pg_policies WHERE tablename = 'contracts';
```

## Temporal

```bash
# Access Temporal UI
open http://localhost:8088

# List running workflows
tctl wf list --status Running

# Describe a workflow
tctl wf describe --workflow-id contract-approval-{uuid}

# Signal a workflow manually
tctl wf signal --workflow-id {id} --name {signal-name} --input '{}'
```

## awoctl

Project CLI for operational tasks:

```bash
# Session management
awoctl sessions revoke-all --user-id {uuid}
awoctl sessions revoke-tenant --tenant-id {uuid}

# Event management
awoctl events dlq list
awoctl events replay --dlq {topic} --from {date}

# Tenant management
awoctl tenants provision --name "ACME" --admin-email admin@acme.com

# Migrate (same as make migrate-up)
awoctl migrate up
awoctl migrate status
```

Build awoctl: `go build -o bin/awoctl ./cmd/awoctl`
