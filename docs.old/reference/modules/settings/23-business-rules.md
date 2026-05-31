# Chapter 23: Business Rules & Validation

[← Troubleshooting](./22-troubleshooting.md) | [Next: Summary →](./24-summary.md)

---

## Configuration Key Rules

| Rule | Detail |
|---|---|
| Key must be registered | A `(module_name, config_key)` pair must exist in `config_definitions` before it can be set |
| Key format | `module_name`: lowercase, max 50 chars; `config_key`: lowercase with underscores, max 100 chars |
| Value must match type | A `string` key cannot receive an integer value; validated against `config_type` |
| Validation rules | Value must pass the JSON Schema in `validation_rules` if present |
| Feature flag gate | If `required_feature_flag` is set, the tenant must have that flag enabled |

---

## Hierarchy Rules

| Rule | Detail |
|---|---|
| Entity beats tenant | Entity-level value always wins over tenant-level for the same key |
| Tenant beats system | Tenant-level value always wins over system default |
| No entity-to-entity inheritance | Entity A does not inherit from its parent entity B's settings; only from tenant/system |
| Deletion falls back | Deleting an entity override restores the tenant value; deleting a tenant override restores the system default |

---

## Template Rules

| Rule | Detail |
|---|---|
| SYSTEM templates: `tenant_id IS NULL` | Platform-level templates must have no tenant |
| TENANT templates: `tenant_id IS NOT NULL` | Custom templates must belong to a tenant |
| Conflict resolution required | Every template must specify `merge`, `replace`, or `preserve` |
| Version is required | Every template must have a semantic version string |
| `applicable_tenant_types` | Optional; templates may restrict which tenant types they apply to |
| `required_feature_flags` | If set, all flags must be enabled on target tenant for application to succeed |

---

## Multi-Tenant Isolation Rules

All configuration operations are enforced by Row-Level Security:

| Policy | Detail |
|---|---|
| Tenant sees own data only | `current_tenant_id()` must match `tenant_id` on all tables |
| SYSTEM templates are read-only for tenants | Tenants can read SYSTEM templates but cannot modify them |
| Cross-tenant reads blocked | RLS on `tenant_configurations` prevents any cross-tenant read |
| Entity isolation | RLS on `entities` ensures entity configs are tenant-scoped |
| Audit isolation | `configuration_audit` is tenant-scoped; tenants see only their own history |

---

## Write Rules

| Rule | Detail |
|---|---|
| Optimistic locking | `settings_version` must be checked before concurrent writes to avoid lost updates |
| Atomic JSONB update | Use `jsonb_set` (not full blob replacement) to avoid overwriting concurrent changes |
| Audit required | Every configuration write must produce an audit record in `configuration_audit` |
| `current_tenant_id()` must be set | All writes must run within a tenant-scoped DB session |

---

## Validation at Service Layer

Before writing a configuration value, the service layer:

1. **Checks key exists** in `config_definitions`
2. **Validates type** — casts the value to the declared `config_type`
3. **Runs validation rules** — evaluates the JSON Schema in `validation_rules`
4. **Checks feature flags** — confirms `required_feature_flag` is enabled for the tenant
5. **Checks permissions** — verifies the calling user has the appropriate `settings:{scope}:write` permission

---

## Constraints on Configuration Values

### String Configurations
- Non-empty unless explicitly allowed by validation rules
- Max length enforced by `validation_rules.maxLength`
- Pattern matching enforced by `validation_rules.pattern`

### Numeric Configurations
- `minimum` / `maximum` enforced by `validation_rules`
- Decimal precision depends on application-layer handling

### Enum Configurations
- Value must be one of the `validation_rules.enum` values

### Boolean Configurations
- Must be strictly `true` or `false` (not `1`, `0`, `"true"`, `"false"`)

---

## Bulk Operation Rules

| Rule | Detail |
|---|---|
| Batch size limit | Max 500 entities per batch; use pagination for larger sets |
| Continue on error | When `continue_on_error: true`, errors are collected and reported but don't stop the operation |
| Correlation ID | All audit records from one bulk operation share the same `correlation_id` |
| `preserve_explicit_overrides` | When `true`, entities with an existing override for the key are skipped |

---

[← Troubleshooting](./22-troubleshooting.md) | [Next: Summary →](./24-summary.md)
