# Entity Definition Specification

## Overview

An `EntityDefinition` is the single declaration that drives five subsystems
simultaneously: persistence, API generation, SDUI, authorization, and workflow.
One registration call wires everything.

## Registration

```go
func init() {
    def.Register(&MyEntityDefinition)
}
```

`Register` is package-level. `init()` is the only correct call site.
Never call `Register` from a request handler or after bootstrap completes.

## EntityDefinition Interface

Every entity must implement `def.EntityDefinition`. The two concrete types
provided by the framework are:

| Type | Backing store | Use when |
|---|---|---|
| `def.SystemDefinition` | PostgreSQL typed columns | Financial integrity, IAM, inventory |
| `def.CustomDefinition` | JSONB in `custom_entity_records` | Tenant-specific, evolving schema |

## Naming

```
{module}_{noun}
```

Examples: `iam_user`, `finance_invoice`, `platform_tenant`.

- Module prefix is always the Go package name of the owning module.
- Noun is singular snake_case.
- **Never rename** — embedded in migration filenames, Temporal workflow IDs,
  Redis cache keys, Casbin policy objects.

## Fields

Declared in `EntityDefinition.Fields []def.FieldDef`.

### Field Types

| Type | PostgreSQL | Go type |
|---|---|---|
| `FieldTypeData` | `varchar(n)` | `string` |
| `FieldTypeSmallText` | `varchar(1024)` | `string` |
| `FieldTypeLongText` | `text` | `string` |
| `FieldTypeInt` | `bigint` | `int64` |
| `FieldTypeFloat` | `double precision` | `float64` |
| `FieldTypeCurrency` | `numeric(20,4)` | `decimal.Decimal` |
| `FieldTypeBool` | `boolean` | `bool` |
| `FieldTypeDate` | `date` | `time.Time` |
| `FieldTypeDateTime` | `timestamptz` | `time.Time` |
| `FieldTypeTime` | `time` | `string` |
| `FieldTypeSelect` | `varchar` + CHECK | `string` |
| `FieldTypeMultiSelect` | `text[]` | `[]string` |
| `FieldTypeNamingSeries` | `varchar(100)` | `string` |
| `FieldTypeJSON` | `jsonb` | `map[string]any` |
| `FieldTypeLink` | FK column + index | `uuid.UUID` |
| `FieldTypeLinkList` | JSONB array | `[]uuid.UUID` |
| `FieldTypeDynamicLink` | two varchar columns | `string` |

### Field Constraints

```go
def.FieldDef{
    Name:       "email",
    Type:       def.FieldTypeData,
    Required:   true,   // non-null, validated before save
    Unique:     true,   // unique index
    Immutable:  true,   // rejected on update
    Sensitive:  true,   // excluded from logs and standard API responses
    Searchable: true,   // GIN trigram index; enables ?q= search
    MaxLen:     255,
    ReadOnly:   true,   // computed; API ignores writes
}
```

### NamingSeries Format

```
{PREFIX}-{YYYY}-{SEQ:N}
```

- `{YYYY}` — 4-digit year, resets counter each year when `YearReset: true`
- `{SEQ:N}` — zero-padded sequence, N digits wide
- Example: `INV-{YYYY}-{SEQ:6}` → `INV-2026-000001`
- `TenantOverridable: true` → tenant can customise prefix via Settings module

## Edges

```go
def.EdgeDef{
    Name:          "lines",
    Target:        "finance_invoice_line",  // qualified name
    Type:          def.EdgeOneToMany,
    ForeignKey:    "invoice_id",
    CascadeDelete: true,
    Label:         "Invoice Lines",
    OrderBy:       "sort_order ASC, created_at ASC",
}
```

Edge types: `EdgeOneToMany`, `EdgeManyToOne`, `EdgeManyToMany`.

**Never lazy-load edges.** Explicitly request via `QueryOption`s.

## Permissions

See `AUTHORIZATION_SPEC.md` for full details.

```go
Permissions: def.PermissionSet{
    Create: []string{"iam.user.create"},
    Read:   []string{"iam.user.read"},
    Write:  []string{"iam.user.update"},
    Delete: []string{"iam.user.delete"},
    Actions: map[string][]string{
        "activate": {"iam.user.activate"},
        "suspend":  {"iam.user.suspend"},
    },
    Policy: ownerOrAdminPolicy,
}
```

**INVARIANT**: Every string in `PermissionSet` is a permission identifier.
Never a role name. Never `"role:*"`. Never `"user:*"`.

## Hooks

```go
Hooks: def.HookSet{
    BeforeCreate: []def.BeforeCreateHook{&UserValidator{}},
    AfterCreate:  []def.AfterCreateHook{&WelcomeEmailHook{}},
    BeforeUpdate: []def.BeforeUpdateHook{&StatusGuard{}},
    AfterUpdate:  []def.AfterUpdateHook{&AuditHook{}},
    BeforeDelete: []def.BeforeDeleteHook{&DeletionGuard{}},
}
```

Execution order: declaration order within each stage.

| Hook | Inside TX? | Abort via |
|---|---|---|
| `BeforeCreate/Update/Delete` | No | Return `ValidationError` or `BusinessError` |
| `AfterCreate/Update/Delete` | **Yes** | Return error → rollback |

## Actions

```go
Actions: []def.ActionDef{
    {
        Name:        "activate",
        Method:      def.ActionMethodPost,
        Label:       "Activate User",
        Description: "Moves user from pending to active status.",
        Permission:  "iam.user.activate",   // MUST be a permission identifier
        Icon:        "fa fa-check",
        HandlerFunc: activateUser,
    },
}
```

Auto-generates route: `POST /api/v1/{module}/{resource}/:id/{action_name}`

## Workflow Triggers

```go
WorkflowTriggers: []def.WorkflowTrigger{
    {
        On:         def.EventOnCreate,
        WorkflowFn: "UserOnboardingWorkflow",
        TaskQueue:  "iam.user.onboarding",
        InputBuilder: func(r *def.EntityRecord, tc def.TriggerContext) (any, error) {
            return UserOnboardingInput{TenantID: r.TenantID, UserID: r.ID}, nil
        },
    },
}
```

Workflow start is **outside** the PostgreSQL transaction. Failure → retry queue.

## Policy Function

```go
Policy: def.PolicyFunc(func(ctx context.Context) def.Filter {
    actor := session.ActorFromContext(ctx)
    if actor.HasRole("role:finance.manager") {
        return nil  // unrestricted within tenant
    }
    return filter.Eq("owner_id", actor.UserID)
})
```

Policy functions inject WHERE predicates into every Query call.
They are applied **after** RBAC gates, providing row-level filtering.

## System Entities (Mandatory)

These entities MUST be `def.SystemDefinition` (enforced by compiler):

- All entities in `iam` module
- `finance_ledger_entry`, `finance_journal_entry`, `finance_payment`
- `finance_tax_entry`
- `platform_tenant`

## Auto-Generated Infrastructure

One `Register` call produces:

1. **5 CRUD routes** — GET list, GET :id, POST, PATCH :id, DELETE :id
2. **N action routes** — one per `ActionDef`
3. **SDUI pages** — list, create form, edit form, detail view (cached 5 min in Redis)
4. **CapabilityGrants** — one per permission identifier per operation
5. **Migration scaffolding hints** — compiler emits table names and column list

## Required Columns (all system entity tables)

Every system entity table must have:

```sql
id          uuid            PRIMARY KEY DEFAULT gen_random_uuid(),
tenant_id   uuid            NOT NULL REFERENCES platform_tenants(id),
created_at  timestamptz     NOT NULL DEFAULT now(),
updated_at  timestamptz     NOT NULL DEFAULT now(),
created_by  uuid,           -- iam_user.id, nullable (background ops)
updated_by  uuid,           -- iam_user.id, nullable
custom_fields jsonb         NOT NULL DEFAULT '{}'
```

Plus entity-specific columns. Plus RLS (see `RLS_SPEC.md`).
