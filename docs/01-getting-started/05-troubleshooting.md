---
title: Troubleshooting
portal: 1 — Getting Started
section: 01-getting-started
audience: [backend-engineer, frontend-engineer]
related:
  - "[Quick Start](01-quick-start.md)"
  - "[Prerequisites](02-prerequisites.md)"
  - "[Common Commands](04-common-commands.md)"
---

# Troubleshooting

Common issues during local development setup and their fixes.

## Database

### `FATAL: role "postgres" does not exist`

The PostgreSQL container isn't running, or `DATABASE_URL` points to the wrong host.

```bash
# Check containers
docker ps | grep postgres

# Start infra if not running
make infra-up

# Wait a few seconds, then retry
```

### `connection refused` on port 5432

PostgreSQL isn't ready yet (still starting up):

```bash
# Wait for it to accept connections
until pg_isready -h localhost -p 5432; do sleep 1; done
```

### `pq: RLS policy blocked row`

You're running a query without setting `app.tenant_id`. Make sure the call goes through `store.WithTenant()`. Never call SQLC query functions directly outside a `WithTenant` block in application code.

In tests: make sure the test helper sets up the tenant and calls `WithTenant` correctly.

### Migration says `dirty database version N`

A previous migration run failed mid-way:

```bash
# Force the version to clean state
migrate -database "$DATABASE_URL" -path db/migration force N

# Then re-run
make migrate-up
```

### `no migrations found` after `make migrate-create`

Check the file was created in the right directory:

```bash
ls db/migration/ | tail -5
```

Also confirm the filename uses the correct format: `{seq}_{description}.up.sql`.

## Wire / Code Generation

### `wire: no provider for X`

You added a new dependency to a constructor but didn't add its provider to the relevant `ProviderSet`.

Steps:
1. Find which package provides type X
2. Add its `ProviderSet` to the Wire `Build()` call in `cmd/server/wire.go`
3. Run `make wire`

### `wire_gen.go` compile error after pull

Someone changed a provider. Regenerate:

```bash
make wire
```

### SQLC `unknown column` error

The query references a column that doesn't exist yet. Either:
- The migration hasn't been applied: `make migrate-up`
- There's a typo — check the column name against `\d {table}` in psql

## Go Build

### Import cycle

You imported the handler from the service, or domain from the repository. The dependency direction is:

```
handler → service → repository → domain
```

Domain imports nothing from the module. If you need a type from another layer in domain, extract it to a shared package.

### `cannot use X as type Y`

Usually a `*pgtype.Date` where a `time.Time` is expected, or `decimal.Decimal` where `float64` is expected. Check the mapper functions — SQLC types must be converted to domain types in `mappers.go`.

## Redis

### `NOAUTH Authentication required`

Redis requires a password but `REDIS_URL` doesn't include one. Check `.env`:

```
REDIS_URL=redis://:{password}@localhost:6379/0
```

### Sessions always 401

1. Check Redis is running: `docker ps | grep redis`
2. Check the session TTL hasn't expired
3. Confirm `SESSION_SECRET` in `.env` matches the value used when the token was issued

## Temporal

### Workflows not executing

1. Temporal server must be running: `docker ps | grep temporal`
2. Workers must be registered — the app must be running
3. Check the task queue name matches: `tctl tq describe --task-queue awoerp.contracts`

### `namespace not found`

Set `TEMPORAL_NAMESPACE=default` for local development. The `default` namespace is created by the auto-setup image automatically.

## Tests

### Tests pass locally, fail in CI

1. Race condition — run locally with `-race` flag: `go test -race ./...`
2. Missing environment variable in CI — check `.github/workflows/ci.yml` env section
3. Migration not run in CI — the CI workflow must run `make migrate-up` before tests

### Integration tests skip

Integration tests require `DATABASE_URL` to be set. If running `go test ./...` without the env var, integration tests are skipped automatically (build tag guard).

Run explicitly:
```bash
DATABASE_URL=postgres://... go test -tags integration ./...
```

## General

### `make: command not found`

Install `make`:
```bash
# macOS
xcode-select --install

# Ubuntu/Debian
sudo apt-get install make
```

### Permission denied on `/tmp` (Termux)

Termux blocks `/tmp` creation. Do **not** run `go build`, `go run`, or `go test` directly. Use `make` targets which are configured for the Termux environment, or have the user run them.
