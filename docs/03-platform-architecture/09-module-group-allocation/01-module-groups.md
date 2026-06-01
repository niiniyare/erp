---
title: Module Group Allocation
portal: 3 — Platform Architecture
section: 03-platform-architecture
audience: [architect, backend-engineer]
related:
  - "[Module Map](../00-overview/03-module-map.md)"
  - "[Migration Strategy](../03-data-architecture/03-migration-strategy.md)"
  - "[Wire Overview](../06-dependency-injection/01-wire-overview.md)"
---

# Module Group Allocation

## Purpose

Module groups partition:
1. **Migration sequence numbers** — each module has a reserved range preventing conflicts
2. **Wire ProviderSet files** — each module registers in its own `wire.go`
3. **Task queue names** — `awoerp.{module}` for Temporal workers

## Group Registry

| Group | Range | Module | Package path |
|-------|-------|--------|-------------|
| 01 | `01xxxx` | Platform / Core | `internal/core/platform` |
| 02 | `02xxxx` | IAM | `internal/core/iam` |
| 03 | `03xxxx` | Tenant | `internal/core/tenant` |
| 04 | `04xxxx` | Entities | `internal/core/entities` |
| 05 | `05xxxx` | (reserved) | — |
| 06 | `06xxxx` | Employees | `internal/core/employees` |
| 07 | `07xxxx` | (reserved) | — |
| 08 | `08xxxx` | (reserved) | — |
| 09 | `09xxxx` | (reserved) | — |
| 10 | `10xxxx` | (reserved) | — |
| 11 | `11xxxx` | Contracts | `internal/core/contracts` |
| 12 | `12xxxx` | Finance | `internal/core/finance` |
| 13 | `13xxxx` | Procurement | `internal/core/procurement` |
| 14 | `14xxxx` | Projects | `internal/core/projects` |
| 15 | `15xxxx` | HR | `internal/core/hr` |
| 16 | `16xxxx` | CRM | `internal/core/crm` |
| 17-99 | `17-99xxxx` | (reserved for future) | — |

## Migration Numbering

Within each group, sequences are 4-digit zero-padded:

```
011001_create_contracts.up.sql           ← first migration in contracts group
011002_create_contract_lines.up.sql
011003_create_contract_views.up.sql
...
011999_...                               ← maximum 999 migrations per group
```

If a group needs more than 999 migrations — extremely unlikely — request a group extension.

## Claiming a New Group

When adding a new business module:

1. Choose the next unreserved group number
2. Update this document
3. Create the first migration: `{group}0001_create_{module}_schema.up.sql`
4. Create `internal/core/{module}/wire.go` with empty `ProviderSet`
5. Register the ProviderSet in `cmd/server/wire.go`

## Temporal Task Queues

Task queue name convention:

```
awoerp.{module_name}

Examples:
  awoerp.contracts
  awoerp.finance
  awoerp.iam
  awoerp.entities
```

Register workers in the module's `WorkerSet`:

```go
// internal/core/contracts/wire.go
var WorkerSet = wire.NewSet(
    NewContractActivities,
    NewContractsWorker,
)
```

## Cross-Module Dependencies

Allowed dependency directions:

```
contracts → entities (read entity name/code)
contracts → finance  (trigger invoice on activation)
finance   → entities
hr        → employees
hr        → entities
```

**Prohibited:**
- Circular dependencies (contracts → finance → contracts)
- Lower-group modules importing higher-group modules (IAM importing contracts)
- Domain packages importing from other module domains

Cross-module access goes through **interfaces**, not direct struct imports:

```go
// contracts domain doesn't import finance domain
// Instead, contracts defines what it needs:
type VendorLookup interface {
    GetVendorName(ctx context.Context, entityID uuid.UUID) (string, error)
}
```

The finance module (or entity module) provides an implementation. Wire binds them.
