---
title: "Developer Tooling"
volume: "VIII — Developer Experience"
chapter: "40-B"
phase: 4
status: draft
audience: [All Engineers]
---

# Chapter 40-B — Developer Tooling

> **Volume:** VIII — Developer Experience
> **Audience:** All Engineers
> **Prerequisites:** Chapter 07-B — UI Composition Patterns, Chapter 19-B — Backend-Driven Navigation
> **Phase:** 4

---

## 40B.1 Developer Tooling Philosophy

AwoERP's SDUI platform introduces a new class of developer workflow: writing Go code that produces JSON schemas that drive UI rendering. Without dedicated tooling, this workflow degrades into a slow edit-deploy-view cycle where a developer makes a change to a surface definition in Go, rebuilds the binary, restarts the server, and navigates to the surface in a browser to see the result. For a platform with 200+ surface definitions across 10+ modules, this cycle is untenable.

The tooling philosophy is: **validate early, preview locally, diff confidently**. Every surface definition should be validatable without a running server. Every navigation tree should be previewable for any actor context without real user data. Every schema change should produce a readable diff that reviewers can understand without knowing the full schema structure.

All developer tooling is designed to run in the Go environment without external dependencies — no Node.js, no Docker for basic validation tasks. The `awo` CLI binary is built alongside the main server binary and is available in any environment with the Go toolchain.

> **ℹ Note:** All Go toolchain commands (`go build`, `go run`, `go test`) must be run by the developer, not by Claude Code or automated tooling in this project. See the project's CLAUDE.md for the rationale (Termux sandbox limitation). The `awo` CLI itself is a Go binary that the developer builds and runs.

---

## 40B.2 Local Development Environment Setup

### 40B.2.1 Prerequisites

```
Go 1.22+          — Backend and CLI
Node.js 20+       — amis web renderer development (not required for backend-only work)
Flutter 3.19+     — Mobile renderer development
Docker 24+        — PostgreSQL, Redis, Temporal server
make              — Task runner (GNU make)
```

Verify setup:
```bash
go version         # go1.22+
node --version     # v20+
flutter --version  # 3.19+
docker compose version
```

### 40B.2.2 `make dev-ui` — Spin Up Full SDUI Stack Locally

The `make dev-ui` target starts all services required for UI development:

```makefile
# Makefile (relevant targets)

.PHONY: dev-ui dev-db dev-server dev-web

# Full UI development stack
dev-ui: dev-db dev-server dev-web

# Start PostgreSQL and Redis
dev-db:
	docker compose up -d postgres redis temporal temporal-ui
	@echo "Waiting for PostgreSQL..."
	@until docker compose exec postgres pg_isready -U awo; do sleep 1; done
	@echo "Running migrations..."
	migrate -path db/migration -database "$$DATABASE_URL" up

# Start Go server in watch mode (using air)
dev-server:
	air -c .air.toml &
	@echo "Server starting at http://localhost:3000"

# Start web dev server (Vite proxy)
dev-web:
	cd web && npm run dev &
	@echo "Web UI starting at http://localhost:5173"

# Build the awo CLI tool
build-cli:
	go build -o bin/awo ./cmd/awo/
```

The `.air.toml` configuration reloads the Go server on any `.go` file change. The `web/vite.config.ts` proxies all `/api/` requests to `localhost:3000`.

### 40B.2.3 Hot Reload for UI Surface Definitions

Surface definitions stored in the database can be updated via the admin API without restarting the server. For surface definitions that are code-defined (Go structs), `air` watches for changes and rebuilds automatically.

For rapid JSON schema iteration, developers can use the `PUT /api/v1/ui/surfaces/{surface_id}` endpoint to upload a surface definition JSON directly and see it reflected immediately:

```bash
# Upload a modified surface definition for live preview
curl -X PUT http://localhost:3000/api/v1/ui/surfaces/finance.invoices.list \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $DEV_TOKEN" \
  -H "X-Tenant-ID: $DEV_TENANT_ID" \
  -d @web/schemas/pages/finance/invoices-list.json
```

---

## 40B.3 CLI Tools

The `awo` CLI provides commands for the full UI development lifecycle. Build it with `make build-cli`, then add `./bin` to `$PATH`.

### 40B.3.1 `awo ui validate` — Validate a JSON Surface Definition

```bash
# Validate a single surface file
awo ui validate web/schemas/pages/finance/invoices-list.json

# Validate all surfaces in a directory
awo ui validate web/schemas/pages/

# Validate with compliance checks
awo ui validate --compliance web/schemas/pages/finance/

# Validate API references against registered endpoints
awo ui validate --check-api-refs web/schemas/pages/

# Example output:
# ✅ finance.invoices.list — valid
# ❌ finance.accounts.new
#    Line 45: unknown action type "ajax2" (did you mean "ajax"?)
#    Line 67: API endpoint "POST /api/v1/finance/account" not registered (nearest: /api/v1/finance/accounts)
#    Compliance: pii_fields not declared; surface contains field "national_id"
```

Validation checks:
- JSON syntax validity
- Schema compliance (all required fields present, no unknown fields)
- API endpoint references exist in the registered route table
- Permission strings follow `<module>.<resource>.<action>` format and exist in the permission registry
- Template references resolve to registered templates
- Composition depth ≤ 8
- Compliance metadata present (`pii_fields` if PII columns detected)

### 40B.3.2 `awo ui compile` — Compile an AST for a Given Context (Dry Run)

Compiles a surface definition for a specified actor context and outputs the resulting AST JSON. This is the primary tool for debugging composition, permission pruning, and slot resolution:

```bash
# Compile a surface for a specific actor context (dry run — no server needed)
awo ui compile finance.invoices.list \
  --tenant-id 3f2e1d4c-5b6a-7c8d-9e0f-a1b2c3d4e5f6 \
  --actor-type tenant_staff \
  --permissions finance.invoices.read,finance.invoices.write \
  --locale en-KE \
  --output compiled-invoice-list.json

# Compile with a mock context from a YAML file
awo ui compile finance.invoices.list --context dev/mock-contexts/finance-manager.yaml

# Show which nodes were pruned and why
awo ui compile finance.invoices.list --context dev/mock-contexts/pump-attendant.yaml --show-pruned
```

**Mock context YAML format** (`dev/mock-contexts/finance-manager.yaml`):
```yaml
actor_type: tenant_staff
tenant_id: 3f2e1d4c-5b6a-7c8d-9e0f-a1b2c3d4e5f6
user_id: a1b2c3d4-e5f6-7890-abcd-ef1234567890
permissions:
  - finance.accounts.read
  - finance.accounts.write
  - finance.invoices.read
  - finance.invoices.write
  - finance.invoices.post
  - finance.reports.read
feature_flags:
  etims_v2_enabled: true
  mpesa_express: false
enabled_modules:
  - finance
  - hr
locale: en-KE
```

### 40B.3.3 `awo ui diff` — Diff Two Surface Definition Versions

```bash
# Diff current surface against a previous version
awo ui diff finance.invoices.list --base HEAD~1 --head HEAD

# Diff compiled ASTs for two actor contexts (shows what Pump Attendant vs Finance Manager sees differently)
awo ui diff finance.invoices.list \
  --context-a dev/mock-contexts/pump-attendant.yaml \
  --context-b dev/mock-contexts/finance-manager.yaml

# Example output:
# Surface: finance.invoices.list
# Context diff (pump-attendant vs finance-manager):
# + [finance-manager] action: "Post to Ledger" (requires finance.invoices.post)
# + [finance-manager] column: "Cost Centre" (requires finance.cost_centres.read)
# - [pump-attendant] action: "Export CSV" (requires finance.invoices.export) — PRUNED
```

### 40B.3.4 `awo ui render` — Render to HTML Snapshot for Review

```bash
# Render a surface to static HTML for visual review (no amis JS — static HTML only)
awo ui render finance.invoices.list \
  --context dev/mock-contexts/finance-manager.yaml \
  --output /tmp/finance-invoices-list.html

# Open in browser
open /tmp/finance-invoices-list.html
```

This produces a static HTML snapshot of the surface layout useful for design reviews and documentation screenshots, without requiring a running server or browser with amis loaded.

### 40B.3.5 `awo nav tree` — Print Navigation Tree for a Given Actor Context

```bash
# Print the navigation tree for a given context
awo nav tree --context dev/mock-contexts/pump-attendant.yaml

# Example output:
# Navigation Tree — Shell Maanzoni / staff / pump-attendant
# └── Fuel Operations
#     ├── Today's Pump Sessions [fuel.sales.today]  ← VISIBLE
#     ✗   Fuel Grades (EPRA)     — PRUNED (missing: fuel.grades.read)
#     ✗   Tank Inventory          — PRUNED (missing: fuel.inventory.read)
#     ✗   Fuel Deliveries         — PRUNED (missing: fuel.deliveries.read)
# ✗ Finance                      — PRUNED (group: no visible items)
# ✗ Shop & Convenience           — PRUNED (group: no visible items)
#
# Total: 1 visible / 15 total (14 pruned)
# Cache key: nav:3f2e...:hash-a3f9:staff:en-KE:flag-b2c8

# Show tree as JSON (for piping to other tools)
awo nav tree --context dev/mock-contexts/finance-manager.yaml --format json > nav-tree.json
```

---

## 40B.4 Local Dev Emulator

The local dev emulator provides a lightweight server that simulates the full AwoERP request pipeline with mock data, eliminating the need for a real database connection during pure UI development work.

### 40B.4.1 Mock Tenant Context Injection

The emulator reads mock context files from `dev/mock-contexts/` and serves them at well-known endpoints:

```bash
# Start the emulator
awo ui emulator --context-dir dev/mock-contexts/ --port 3001

# The emulator serves:
# GET /api/v1/ui/navigation      — returns nav tree from mock context
# GET /api/v1/ui/surfaces/{id}   — returns compiled surface from local definition
# GET /api/v1/finance/accounts   — returns sample data from dev/fixtures/
```

The Vite proxy config in `web/vite.config.ts` can point to the emulator instead of the real server:

```typescript
// web/vite.config.ts
export default defineConfig({
    server: {
        proxy: {
            '/api': {
                target: process.env.DEV_API_TARGET || 'http://localhost:3001', // emulator
                changeOrigin: true
            }
        }
    }
});
```

### 40B.4.2 Mock Permission Sets

Predefined mock contexts cover the key actor types for Shell Maanzoni:

```
dev/mock-contexts/
├── platform-admin.yaml
├── tenant-admin.yaml
├── finance-manager.yaml
├── finance-clerk.yaml
├── operations-manager.yaml
├── pump-attendant.yaml
├── hr-admin.yaml
├── portal-supplier.yaml
├── portal-customer.yaml
└── portal-employee.yaml
```

Developers can create custom contexts for edge-case testing by copying and modifying an existing file.

### 40B.4.3 Feature Flag Override for Local Dev

Feature flags can be overridden in the emulator via a flags file:

```yaml
# dev/flags-override.yaml
# Override feature flags for local development
etims_v2_enabled: true
mpesa_express: true
new_dashboard_widgets: true  # testing unreleased feature
```

```bash
awo ui emulator --flags-override dev/flags-override.yaml
```

---

## 40B.5 JSON Schema Browser

### 40B.5.1 Schema Browser Web UI

The schema browser is a development-mode web UI that documents every registered surface definition, template, mixin, and fragment with live preview:

```bash
# Start the schema browser (dev only — not available in production)
awo ui schema-browser --port 8080
open http://localhost:8080
```

The schema browser provides:
- Full-text search across all surface definitions
- Component-by-component breakdown of each surface
- Interactive context switcher (change actor type → see how the surface changes)
- Template registry browser (all registered templates with their parameters)
- Mixin registry browser
- Token registry (all CSS custom properties with current values)

### 40B.5.2 Component Catalogue with Live Preview

The component catalogue displays every amis component used in AwoERP surfaces with:
- Live preview (rendered by real amis SDK)
- Props reference
- Example JSON
- Usage restrictions (which actor types can use which components)
- Link to the relevant chapter in this documentation

---

## 40B.6 amis Sandbox Integration

The amis sandbox (Chapter 50-C) integrates with the developer tooling by providing a shareable live editor. The `awo ui sandbox` command uploads a surface definition to the sandbox and opens a browser with a preview:

```bash
# Open a surface in the amis sandbox for interactive editing
awo ui sandbox finance.invoices.list --context dev/mock-contexts/finance-manager.yaml

# Output: Sandbox URL: https://sandbox.awoerp.dev/s/abc123
# Opens browser automatically
```

> See Chapter 50-C — Playground and Sandbox Environment for full sandbox documentation.

---

## 40B.7 VS Code Extension

The AwoERP VS Code extension (`awo.awoerp-vscode`) provides IDE-level tooling for surface definition development.

### 40B.7.1 Schema Autocomplete for Surface JSON

The extension provides JSON Schema-driven autocomplete for all surface definition files matching the glob `web/schemas/pages/**/*.json` and `internal/ui/surfaces/**/*.go`. Component names, prop names, and valid prop values are autocompleted from the live schema registry.

```json
// .vscode/settings.json — configure the extension
{
  "awoerp.schemaRegistryUrl": "http://localhost:3000/api/v1/ui/schema-registry",
  "awoerp.mockContextDir": "dev/mock-contexts",
  "awoerp.surfaceDirs": ["web/schemas/pages", "internal/ui/surfaces"],
  "json.schemas": [
    {
      "fileMatch": ["web/schemas/pages/**/*.json"],
      "url": "http://localhost:3000/api/v1/ui/schema-registry/surface.schema.json"
    }
  ]
}
```

### 40B.7.2 Inline Component Documentation

Hovering over a component type name (e.g. `"type": "crud"`) in a JSON surface file shows a documentation tooltip with:
- Component description
- Required and optional props with types
- Link to the relevant chapter in this documentation
- Example JSON snippet

### 40B.7.3 Permission String Autocomplete

When typing a value for `permission_required`, `visibleOn`, or `Authorize()` call arguments, the extension provides autocomplete from the registered permission registry:

```json
// Typing "finance." shows:
"permission_required": "finance.|"
//                               ^ autocomplete here
// Suggestions:
// finance.accounts.read
// finance.accounts.write
// finance.invoices.read
// finance.invoices.write
// finance.invoices.post
// finance.invoices.export
// ...
```

The permission registry is fetched from `GET /api/v1/platform/permissions` (platform admin endpoint) and cached locally. The extension refreshes the cache when the Go server restarts (detected via a file watcher on `cmd/server/wire_gen.go`).

The extension also highlights invalid permission strings (ones not in the registry) with a warning underline, helping catch typos before they reach code review.
