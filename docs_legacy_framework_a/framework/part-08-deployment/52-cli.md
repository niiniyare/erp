> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "Chapter 52: CLI Reference — the awo binary"
part: "Part VIII — Deployment and Operations"
chapter: 52
section: "52-cli"
related:
  - "[Chapter 11: Database Migrations](../part-02-entity-system/11-migrations.md)"
  - "[Chapter 38: Tenant Lifecycle](../part-07-multitenancy/38-tenant-lifecycle.md)"
  - "[Chapter 45: Environment Architecture](../part-08-deployment/45-environment.md)"
---

# Chapter 52: CLI Reference — the `awo` binary

`awo` is the command-line interface for Awo Framework projects. It is a standalone binary that ships with the framework — no separate download or version management tool required. Once you depend on `awo.so/framework` in your project, `go install awo.so/framework/cmd/awo@latest` gives you the binary.

It is conceptually similar to Frappe's `bench` tool, but simpler: one binary, one Go project, project-scoped commands. There is no concept of a "sites" directory or multi-app installation. An Awo project is a single Go binary with a single configuration.

---

## 52.1 Overview

### 52.1.1 Installation

```bash
go install awo.so/framework/cmd/awo@latest
```

Or, if you are working from a local checkout of the framework:

```bash
cd /path/to/awo-framework
go install ./cmd/awo
```

The binary is placed in `$GOPATH/bin/awo` (ensure this is on your `$PATH`).

### 52.1.2 Configuration

`awo` reads configuration from the `AWO_*` environment variables defined in Chapter 12. For local development, create a `.env` file in the project root and load it before running commands:

```bash
source .env && awo serve
# or
dotenv awo serve
```

Override the config source with `--config`:

```bash
awo --config /etc/awo/production.env migrate up
```

### 52.1.3 Global Flags

All subcommands accept these flags:

| Flag | Description |
|------|-------------|
| `--config <path>` | Path to a `.env` file (loaded before environment variables) |
| `--env <name>` | Environment name: `development`, `staging`, `production` (default: `development`) |
| `--log-level <level>` | `debug`, `info`, `warn`, `error` (default: `info`) |
| `--no-color` | Disable ANSI colour output |

---

## 52.2 Project Scaffolding

### 52.2.1 `awo new <project-name>`

Scaffold a new Awo project:

```bash
$ awo new acme-erp
Creating project: acme-erp

  Creating directory structure...
  Generating go.mod...
  Generating main.go...
  Generating docker-compose.yml...
  Generating .env.example...

Done! Next steps:

  cd acme-erp
  cp .env.example .env
  # Edit .env with your database credentials
  docker compose up -d
  awo migrate up
  awo tenant create --name="Acme Ltd" --slug=acme --email=admin@acme.co.ke
  awo serve
```

Scaffolded directory layout:

```
acme-erp/
  cmd/
    server/
      main.go         # Entry point: bootstraps framework, starts Fiber + worker
  internal/
    hooks/            # Custom hook implementations
    workflows/        # Application-specific Temporal workflows
  migrations/         # golang-migrate .up.sql / .down.sql files
  fixtures/           # Seed data JSON/SQL files
  templates/
    print/            # HTML print templates
    email/            # Notification email templates
  docker-compose.yml
  .env.example
  go.mod
  go.sum
  Makefile
```

`cmd/server/main.go` in the scaffold:

```go
package main

import (
    "awo.so/framework/bootstrap"
    "acme-erp/internal/hooks"
    "acme-erp/internal/workflows"
)

func main() {
    bootstrap.Mount(
        bootstrap.WithHooks(hooks.Register),
        bootstrap.WithWorkflows(workflows.Register),
    )
}
```

`bootstrap.Mount` wires all framework components (database, Redis, Temporal, EntityRegistry, Fiber) and starts the server. Application code passes hook and workflow registrars.

---

## 52.3 `awo serve` — Run the Application

### 52.3.1 `awo serve`

Start the API server and Temporal worker in one process. This is the recommended mode for development and for small deployments where separate processes are unnecessary:

```bash
$ awo serve
2024/06/15 09:00:00 INFO awo v1.4.2 starting
2024/06/15 09:00:00 INFO connecting to database...
2024/06/15 09:00:00 INFO database ready
2024/06/15 09:00:00 INFO connecting to Redis...
2024/06/15 09:00:00 INFO Redis ready
2024/06/15 09:00:00 INFO registering 47 entity definitions
2024/06/15 09:00:00 INFO Temporal worker started (queue: main)
2024/06/15 09:00:00 INFO Fiber listening on :8080
```

Press `Ctrl+C` to stop. The framework handles `SIGINT` and `SIGTERM` with a 30-second graceful drain: it stops accepting new requests, waits for in-flight requests to complete, closes the Temporal worker cleanly, and exits.

### 52.3.2 `awo serve --port=<N>`

Override the HTTP port (default: `8080`):

```bash
awo serve --port=3000
```

### 52.3.3 `awo serve --worker`

Explicitly request API + worker mode (same as default, but useful in scripts where you want to be explicit):

```bash
awo serve --worker
```

### 52.3.4 `awo worker`

Run only the Temporal worker, without starting the Fiber HTTP server. Use this in production when you deploy API and worker as separate containers:

```bash
# In the API container
awo serve --port=8080

# In the worker container (separate deployment)
awo worker
```

The worker process reads the same configuration as the API process. It connects to Temporal, registers all workflows and activities, and begins polling for tasks. It does not open an HTTP port.

---

## 52.4 `awo migrate` — Database Migrations

Wraps `golang-migrate` with Awo's configuration loading. All migration files live in `migrations/` by default (configurable via `AWO_MIGRATIONS_DIR`).

### 52.4.1 `awo migrate up`

Apply all pending migrations:

```bash
$ awo migrate up
Applying 20240101090000_initial_schema ... OK (1.234s)
Applying 20240615120000_add_customer_credit_limit ... OK (0.045s)
All migrations applied. Version: 20240615120000
```

Safe to run repeatedly — a second run with no pending migrations is a no-op.

Run this as the first step of your deploy pipeline, before restarting the API processes.

### 52.4.2 `awo migrate down [N]`

Roll back N migrations (default: 1):

```bash
$ awo migrate down
Rolling back 20240615120000_add_customer_credit_limit ... OK
Current version: 20240101090000

$ awo migrate down 3
# Rolls back 3 migrations
```

### 52.4.3 `awo migrate version`

Print the current applied version and dirty state:

```bash
$ awo migrate version
Version: 20240615120000 (dirty: false)
```

If `dirty: true`, a migration failed partway through. Resolve the issue and use `migrate force` to clear the dirty state.

### 52.4.4 `awo migrate create <name>`

Create a new migration file pair with a timestamp prefix:

```bash
$ awo migrate create add_inventory_batches
Created: migrations/20240720083000_add_inventory_batches.up.sql
Created: migrations/20240720083000_add_inventory_batches.down.sql

Edit these files, then run 'awo migrate up' to apply.
```

The files are created with placeholder comments:

```sql
-- migrations/20240720083000_add_inventory_batches.up.sql
-- Write your forward migration SQL here.
-- Remember: every table needs tenant_id, RLS enable, and a policy.
```

### 52.4.5 `awo migrate force <V>`

Mark version V as applied without running any SQL. Emergency use only, when a partially applied migration has been manually completed:

```bash
$ awo migrate force 20240720083000
Version forced to 20240720083000 (dirty: false)
WARNING: Use only if you have manually applied the migration SQL.
```

See Chapter 11.9.3 for the full procedure.

---

## 52.5 `awo tenant` — Tenant Management

These commands manage tenant lifecycle. They call the same Temporal workflows that the API would call — `awo tenant create` triggers `TenantProvisioningWorkflow`, not raw SQL.

### 52.5.1 `awo tenant create`

Provision a new tenant:

```bash
$ awo tenant create \
    --name="Acme Petroleum Ltd" \
    --slug=acme-petroleum \
    --email=admin@acmepetroleum.co.ke \
    --plan=growth

Provisioning tenant: acme-petroleum
Starting TenantProvisioningWorkflow...
  ✓ Seed reference data (chart of accounts, leave types, PAYE bands)
  ✓ Create admin user: admin@acmepetroleum.co.ke
  ✓ Assign default roles
  ✓ Configure feature flags (plan: growth)
  ✓ Activate tenant
  ✓ Send welcome email

Tenant created successfully!
  ID:    018f3a2b-7c4d-7000-8000-000000000001
  Slug:  acme-petroleum
  Plan:  growth
  Login: https://acme-petroleum.awo.app/login
```

Flags:
- `--name` (required): Display name
- `--slug` (required): URL-safe identifier, must be unique
- `--email` (required): Admin user email address
- `--plan` (optional): `starter` | `growth` | `enterprise` (default: `starter`)

### 52.5.2 `awo tenant list`

List all tenants with status and plan:

```bash
$ awo tenant list
SLUG                 NAME                      PLAN        STATUS     CREATED
acme-petroleum       Acme Petroleum Ltd        growth      ACTIVE     2024-06-15
sunrise-energy       Sunrise Energy KE         starter     ACTIVE     2024-07-01
lakeside-transport   Lakeside Transport Ltd    enterprise  SUSPENDED  2024-03-20
```

Add `--status=SUSPENDED` to filter by status.

### 52.5.3 `awo tenant suspend`

Suspend a tenant. Triggers `TenantSuspensionWorkflow`:

```bash
$ awo tenant suspend --slug=lakeside-transport --reason="Payment overdue: 30 days"
Suspending tenant: lakeside-transport
  ✓ Invalidate all sessions
  ✓ Set status: SUSPENDED
  ✓ Pause Temporal schedules
  ✓ Notify admin users

Tenant suspended. All API requests will now return 402.
```

### 52.5.4 `awo tenant activate`

Reactivate a suspended tenant. Triggers `TenantReactivationWorkflow`:

```bash
$ awo tenant activate --slug=lakeside-transport
Reactivating tenant: lakeside-transport
  ✓ Set status: ACTIVE
  ✓ Resume Temporal schedules
  ✓ Notify admin users

Tenant reactivated.
```

### 52.5.5 `awo tenant export`

Export all of a tenant's data to JSON files:

```bash
$ awo tenant export --slug=acme-petroleum --out=/tmp/acme-export
Exporting tenant: acme-petroleum
  Exporting table: customers (1,247 rows) ... OK
  Exporting table: journal_entries (8,923 rows) ... OK
  Exporting table: journal_entry_lines (35,691 rows) ... OK
  ...
  Exporting table: audit_log (192,847 rows) ... OK

Export complete: /tmp/acme-export/
  Total: 47 tables, 438,291 rows
  Size:  234 MB (uncompressed)
```

Files are written as newline-delimited JSON (one object per line) for streaming import.

---

## 52.6 `awo entity` — Entity Inspection

Read-only introspection of registered EntityDefinitions. Useful for debugging and for understanding what is registered in a running application.

### 52.6.1 `awo entity list`

```bash
$ awo entity list
NAME                   TYPE      OWNER                  FIELDS  EDGES
tenant                 system    awo.so/framework          15      2
org_node               system    awo.so/framework           9      2
user                   system    awo.so/framework          12      3
role                   system    awo.so/framework           6      1
permission             system    awo.so/framework           4      0
session                system    awo.so/framework          10      1
audit_log              system    awo.so/framework          14      0
feature_flag           system    awo.so/framework           7      0
notification           system    awo.so/framework          12      1
tenant_config          system    awo.so/framework           8      0
custom_field_def       system    awo.so/framework          13      0
report_definition      system    awo.so/framework          11      0
journal_entry          system    awo.so/finance             8      2
...
```

### 52.6.2 `awo entity inspect`

Print the full EntityDefinition for a given entity:

```bash
$ awo entity inspect --name=journal_entry
EntityDefinition: journal_entry
  Owner:    awo.so/finance
  Type:     system
  Table:    journal_entries
  Auditable: true
  PrintTemplate: journal_voucher

  Fields:
    id            UUID        required immutable
    tenant_id     UUID        required immutable
    series        Data        required immutable  max:50
    entry_type    Select      required  options:[DEBIT_NOTE CREDIT_NOTE JOURNAL OPENING]
    posting_date  Date        required
    ref_number    Data        optional  max:100
    memo          LongText    optional
    status        Select      required  options:[DRAFT SUBMITTED CANCELLED]
    total_debit   Currency    required
    total_credit  Currency    required
    custom_fields JSON
    created_at    DateTime
    updated_at    DateTime

  Edges:
    lines  → journal_entry_lines  (one-to-many, cascade delete)
    period → fiscal_periods        (link)

  Hooks:
    before_validate: ValidateDebitCreditBalance
    after_save:      InvalidatePDFCache
    on_submit:       PostToGeneralLedger
    on_cancel:       ReverseGLPostings

  Permission policy: TenantIsolation, DepartmentScope(accounting)
```

---

## 52.7 `awo seed` — Data Seeding

### 52.7.1 `awo seed --module=<module>`

Seed a module's default reference data. Used when onboarding a tenant manually (outside the provisioning workflow) or when resetting a development database:

```bash
$ awo seed --module=finance --tenant=acme-petroleum
Seeding module: finance → tenant: acme-petroleum
  ✓ Chart of accounts (42 accounts)
  ✓ Fiscal year: 2024-2025 (Jul–Jun)
  ✓ Tax groups: VAT 16%, NHIF, NSSF
  ✓ Payment terms: Net 30, Net 60, Immediate

Seed complete.
```

Available modules: `core` (roles, system config), `finance`, `crm`, `inventory`, `hr`, `forecourt`.

### 52.7.2 `awo seed --file=<path>`

Load a fixture file (JSON array of entity records):

```bash
$ awo seed --file=fixtures/test-customers.json --tenant=acme-petroleum
Loading fixtures from: fixtures/test-customers.json
  Inserting 25 customer records... OK

Seed complete.
```

Fixture files use `INSERT ... ON CONFLICT DO NOTHING` semantics — running a seed file twice is safe.

---

## 52.8 `awo console` — Interactive REPL

`awo console` opens a Go REPL with the full framework context loaded — database connection, Redis client, EntityRegistry, and a configurable tenant context. This is the Awo equivalent of `bench console` in Frappe.

```bash
$ awo console --tenant=acme-petroleum
Awo Console v1.4.2
Tenant: acme-petroleum (018f3a2b-7c4d-7000-8000-000000000001)

Available variables:
  db       *sqlx.DB          — raw database connection (tenant RLS set)
  redis    *rueidis.Client   — Redis client
  repo     EntityRepository  — entity repository for current tenant
  ctx      context.Context   — context with tenant + user (system user)

> customers, _, err := repo.Query(ctx, filter.New("customer").Limit(5))
> for _, c := range customers { fmt.Println(c.Field("name")) }
Acme Oils Ltd
Nairobi Fuel Distributors
...

> // Run a raw SQL query
> rows, err := db.QueryContext(ctx, "SELECT COUNT(*) FROM journal_entries")
```

The console uses `yaegi` (a Go interpreter) to evaluate expressions. Not all Go syntax is supported — complex type assertions and generics may not work. For complex ad-hoc scripts, write a small Go program in `cmd/scripts/` instead.

Use cases:
- Debugging data issues in production without writing a one-off endpoint
- Running ad-hoc queries with full RLS context
- Testing workflow inputs before running a workflow
- Inspecting entity records

---

## 52.9 `awo jobs` — Workflow Management

These commands interact with Temporal to inspect and manage running workflows for a tenant.

### 52.9.1 `awo jobs list`

List running or recent workflows for a tenant:

```bash
$ awo jobs list --tenant=acme-petroleum
WORKFLOW ID                                  TYPE                         STATUS    STARTED
acme-petroleum.provision                     TenantProvisioningWorkflow   COMPLETED 2024-06-15 09:00
acme-petroleum.shift.sh-2024-001.close       ShiftCloseWorkflow           RUNNING   2024-07-20 14:30
acme-petroleum.invoice.inv-2024-892.submit   InvoiceSubmitWorkflow        RUNNING   2024-07-20 15:01
acme-petroleum.payroll.run-2024-06           PayrollRunWorkflow           FAILED    2024-06-30 02:00
```

Filter by status: `awo jobs list --tenant=acme-petroleum --status=FAILED`

### 52.9.2 `awo jobs cancel`

Cancel a running workflow. The workflow receives a cancellation signal and should run compensations:

```bash
$ awo jobs cancel acme-petroleum.shift.sh-2024-001.close
Cancel reason (optional): Emergency maintenance
Cancelling workflow: acme-petroleum.shift.sh-2024-001.close
Workflow cancelled.
```

### 52.9.3 `awo jobs retry`

Restart a failed workflow from the beginning with the same input:

```bash
$ awo jobs retry acme-petroleum.payroll.run-2024-06
Retrying workflow: acme-petroleum.payroll.run-2024-06
New run started. Monitor at: http://localhost:8088/namespaces/default/workflows/acme-petroleum.payroll.run-2024-06
```

Note: retry starts a **new run** of the workflow. If the workflow reached a partially complete state, review whether the activities that already succeeded are idempotent before retrying. Most Awo activities use `INSERT ... ON CONFLICT DO NOTHING` or conditional updates to be retry-safe.

---

## 52.10 `awo config` — Tenant Configuration

Inspect and set TenantConfig values from the command line.

### 52.10.1 `awo config get`

```bash
$ awo config get locale.timezone --tenant=acme-petroleum
Key:    locale.timezone
Value:  Africa/Nairobi
Source: system_default  (no tenant override)

$ awo config get integrations.etims.kra_pin --tenant=acme-petroleum
Key:    integrations.etims.kra_pin
Value:  P051234567A
Source: tenant_override
```

### 52.10.2 `awo config set`

```bash
$ awo config set integrations.etims.env production --tenant=acme-petroleum
Updated: integrations.etims.env = "production" for tenant: acme-petroleum

$ awo config set locale.language sw --tenant=acme-petroleum
Updated: locale.language = "sw" for tenant: acme-petroleum
```

Changes take effect on the next request — the in-memory cache is invalidated when the config row is updated.

To unset a tenant override (revert to system default):

```bash
$ awo config unset locale.language --tenant=acme-petroleum
Removed tenant override for locale.language. Now inherits system default: "en"
```

---

## 52.11 `awo version`

Print version information:

```bash
$ awo version
awo CLI:         v1.4.2
Framework:       v1.4.2
App:             v2.1.0 (git: a8c09e76)
Go runtime:      go1.22.4 linux/arm64
Build time:      2024-07-20T08:30:00Z
```

Application version and git SHA are injected at build time via `-ldflags`:

```bash
go build -ldflags "-X awo.so/framework/version.AppVersion=v2.1.0 -X awo.so/framework/version.GitSHA=$(git rev-parse HEAD)"
```

---

## 52.12 Makefile Integration

The starter project's `Makefile` wraps common `awo` commands for convenience:

```makefile
.PHONY: serve worker migrate seed test

serve:
	awo serve

worker:
	awo worker

migrate:
	awo migrate up

migrate-rollback:
	awo migrate down

seed-dev:
	awo tenant create --name="Dev Tenant" --slug=dev --email=dev@localhost --plan=enterprise || true
	awo seed --module=finance --tenant=dev
	awo seed --module=forecourt --tenant=dev

test:
	go test ./... -count=1

test-integration:
	AWO_ENV=test go test ./... -tags=integration -count=1
```

---

## 52.13 Differences from Frappe bench

For teams migrating from Frappe, the key differences:

| bench | awo | Notes |
|-------|-----|-------|
| `bench new-site` | `awo tenant create` | No schema creation — just seed data |
| `bench migrate` | `awo migrate up` | golang-migrate, not Frappe migration runner |
| `bench console` | `awo console` | Go REPL vs Python REPL |
| `bench start` | `awo serve` | API + worker in one command |
| `bench run-tests` | `go test ./...` | Standard Go testing, no bench wrapper |
| `bench install-app` | `go get` + register in `main.go` | Module packages are Go imports |
| `bench get-app` | `go get awo.so/module-name` | No separate app repository mechanism |
| `bench backup` | `awo tenant export` | Per-tenant JSON export |

The biggest conceptual shift is that Frappe bench manages multiple "apps" installed into a "site" directory. Awo is a single Go binary with modules imported as Go packages. There is no runtime app installation — adding a module requires a code change, recompile, and redeploy.
