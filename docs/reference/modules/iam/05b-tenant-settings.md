[<-- Back to Index](README.md)

## Tenant Settings — Behavioural Configuration

### Settings vs. Flags

Settings are values that control how a feature behaves when it is on. They are never binary. A flag says "is approval workflow available?". A setting says "what is the approval threshold?".

| Dimension | Feature Flag | Tenant Setting |
|---|---|---|
| Type | `bool` | `text`, `int`, `decimal`, `enum`, `bool` |
| UI control | Toggle switch | Input / select / number |
| Question answered | Does this exist? | How does this behave? |
| Changed by | Platform operator or tenant admin | Tenant admin (within platform-defined limits) |
| Effect when changed | Feature appears or disappears | Form layout or business logic adjusts |
| Invalidates session | Yes | Not always (depends on setting) |

They share the same key namespace but live in separate tables.

---

### Schema

```sql
-- Setting catalogue — shared, no tenant_id, seeded by developers
CREATE TABLE setting_definitions (
  id            uuid  PRIMARY KEY DEFAULT gen_random_uuid(),
  module_id     uuid  REFERENCES modules(id),
  resource_id   uuid  REFERENCES resources(id),
  action_id     uuid  REFERENCES actions(id),   -- NULL for module/resource-level settings
  setting_key   text  UNIQUE NOT NULL,           -- 'finance.transactions.approval_threshold'
  label         text  NOT NULL,
  description   text,
  value_type    text  NOT NULL,  -- 'bool' | 'int' | 'decimal' | 'text' | 'enum'
  default_value text,
  enum_options  jsonb,           -- [{"value":"soft","label":"Soft"},{"value":"hard","label":"Hard"}]
  min_value     text,            -- for numeric validation
  max_value     text,
  is_system     bool NOT NULL DEFAULT false  -- true = only platform can change
);

-- Per-tenant setting values
CREATE TABLE tenant_settings (
  id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   uuid        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  setting_id  uuid        NOT NULL REFERENCES setting_definitions(id),
  setting_key text        NOT NULL,  -- denormalised for fast lookup
  value       text        NOT NULL,
  set_by      uuid        REFERENCES users(id),
  set_at      timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, setting_id)
);

-- Per-user preferences (display preferences, not business logic)
CREATE TABLE user_preferences (
  id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  pref_key    text        NOT NULL,  -- 'finance.entry_mode', 'finance.show_account_codes'
  value       text        NOT NULL,
  set_at      timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, pref_key)
);
```

---

### Setting Definitions Are Developer-Seeded

Unlike flag definitions which are auto-seeded by triggers, setting definitions are inserted by developers because settings require human decisions about types, defaults, and validation ranges:

```sql
-- Seeded in migration files alongside the module migration
INSERT INTO setting_definitions
    (module_id, setting_key, label, value_type, default_value, min_value)
SELECT m.id, 'finance.transactions.approval_threshold',
       'Approval threshold — transactions above this amount require approval',
       'decimal', '100000', '0'
FROM modules m WHERE m.slug = 'finance';

INSERT INTO setting_definitions
    (module_id, setting_key, label, value_type, default_value, enum_options)
SELECT m.id, 'finance.budget_control_mode',
       'Budget control mode',
       'enum', 'soft',
       '[{"value":"none","label":"None"},{"value":"soft","label":"Warn only"},{"value":"hard","label":"Block"}]'
FROM modules m WHERE m.slug = 'finance';

INSERT INTO setting_definitions
    (module_id, setting_key, label, value_type, default_value)
SELECT m.id, 'finance.decimal_places',
       'Number of decimal places for amounts',
       'int', '2'
FROM modules m WHERE m.slug = 'finance';
```

---

### Resolution

```go
// internal/platform/repo/setting_repo_impl.go

func (r *settingRepoImpl) ResolveForTenant(ctx context.Context,
    tenantID uuid.UUID) (map[string]string, error) {

    rows, _ := r.q.ResolveAllSettingsForTenant(ctx, db.ResolveAllSettingsForTenantParams{
        TenantID: tenantID,
    })
    settings := make(map[string]string, len(rows))
    for _, row := range rows {
        settings[row.SettingKey] = row.EffectiveValue // COALESCE(tenant_value, default_value)
    }
    return settings, nil
}
```

Stored in `sessions.configuration` at login. Accessed via typed helpers on `ResolvedSession`:

```go
func (s *ResolvedSession) SettingBool(key string, def bool) bool {
    if v, ok := s.Configuration.Settings[key]; ok {
        if b, err := strconv.ParseBool(v); err == nil { return b }
    }
    return def
}
func (s *ResolvedSession) SettingDecimal(key string, def decimal.Decimal) decimal.Decimal {
    if v, ok := s.Configuration.Settings[key]; ok {
        if d, err := decimal.NewFromString(v); err == nil { return d }
    }
    return def
}
func (s *ResolvedSession) SettingInt(key string, def int) int {
    if v, ok := s.Configuration.Settings[key]; ok {
        if i, err := strconv.Atoi(v); err == nil { return i }
    }
    return def
}
func (s *ResolvedSession) SettingString(key string, def string) string {
    if v, ok := s.Configuration.Settings[key]; ok { return v }
    return def
}
```

---

### SettingService Interface

```go
type SettingService interface {
    // Catalogue (read-only at runtime)
    ListDefinitions(ctx, params domain.ListSettingDefsParams) ([]*domain.SettingDefinition, error)

    // Tenant configuration
    ResolveForTenant(ctx, tenantID uuid.UUID)                 (map[string]string, error)
    ListForModule(ctx, params domain.ListSettingsParams)      ([]*domain.SettingWithValue, error)
    UpdateModuleSettings(ctx, params domain.UpdateSettingsParams) error
    GetSetting(ctx, params domain.GetSettingParams)           (string, error)
    ResetToDefault(ctx, params domain.ResetSettingParams)     error

    // User preferences
    GetUserPreferences(ctx, userID uuid.UUID)                 (map[string]string, error)
    SetUserPreference(ctx, params domain.SetPrefParams)       error
}
```

---

Next: [Authentication (AuthN)](./06b-authentication.md)
