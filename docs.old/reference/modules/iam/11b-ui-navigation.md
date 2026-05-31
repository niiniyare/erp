[<-- Back to Index](README.md)

## UI Structure — How Configuration Becomes Interface

### The Data-Driven Navigation

The app shell navigation is not hardcoded in Go. It is built by querying the MRA tables, filtered through flags and permissions. The `BootService` runs this at every `/schema/boot` call for an authenticated user.

```
MRA tables (modules + resources)
     ↓
Filter by tenant flags           → only enabled modules and resources
     ↓
Filter by session permissions    → only resources user can read
     ↓
Build nav sections               → what the user sees in the sidebar
```

```go
// internal/platform/boot.go

func (s *BootService) BuildAppShell(ctx context.Context,
    session *domain.ResolvedSession) (map[string]any, error) {

    // One DB query: modules → resources, filtered by tenant flags
    enabledModules, _ := s.moduleRepo.ListEnabledWithResources(ctx,
        domain.ListEnabledModulesParams{TenantID: session.TenantID})

    nav := s.buildNav(enabledModules, session.Permissions)

    return map[string]any{
        "type":      "app",
        "brandName": "Awo",
        "pages":     nav,
    }, nil
}

func (s *BootService) buildNav(modules []domain.ModuleWithResources,
    perms map[string]bool) []any {

    var sections []any
    for _, mod := range modules {
        var items []any
        for _, res := range mod.Resources {
            readKey := mod.Slug + "." + res.Slug + ".read"
            if !perms[readKey] {
                continue  // no read permission → not visible
            }
            items = append(items, map[string]any{
                "label": res.Label,
                "url":   res.NavURL,
                "schema": map[string]any{
                    "type":      "service",
                    "schemaApi": "/schema" + res.NavURL,
                    "fallback":  errorFallback(res.Label),
                },
            })
        }
        if len(items) > 0 {
            sections = append(sections, map[string]any{
                "label":    mod.Label,
                "icon":     mod.Icon,
                "children": items,
            })
        }
    }
    return sections
}
```

The SQL that powers this:

```sql
-- name: ListEnabledModulesWithResources :many
SELECT
    m.id, m.slug AS module_slug, m.label AS module_label,
    m.icon, m.nav_order AS module_nav_order,
    r.id AS resource_id, r.slug AS resource_slug,
    r.label AS resource_label, r.nav_url, r.nav_order AS resource_nav_order
FROM modules m
JOIN resources r ON r.module_id = m.id

-- Module must be enabled (flag=true or no tenant override and default=true)
JOIN feature_flag_definitions mfd ON mfd.module_id = m.id AND mfd.resource_id IS NULL
LEFT JOIN tenant_feature_flags mtf ON mtf.flag_id = mfd.id AND mtf.tenant_id = @tenant_id
WHERE COALESCE(mtf.enabled, mfd.default_value) = true
  AND m.is_active = true

-- Resource must also be enabled
LEFT JOIN feature_flag_definitions rfd ON rfd.resource_id = r.id
LEFT JOIN tenant_feature_flags rtf ON rtf.flag_id = rfd.id AND rtf.tenant_id = @tenant_id
WHERE (rfd.id IS NULL OR COALESCE(rtf.enabled, rfd.default_value) = true)

ORDER BY m.nav_order, r.nav_order;
```

One query. Returns only what exists and is enabled. Permission filter applied in Go over the returned set. Adding a new module requires only DB rows — no Go nav code changes.

---

### UI Form Sections Adapt to Flags and Settings

Forms are built in sections. Each section is conditionally rendered based on the session's pre-computed flags and settings — zero additional DB queries:

```
┌─────────────────────────────────────────────────────────────┐
│ SECTION 1: Transaction Header (always shown)                │
│ Date | Number | Type | Description                         │
├─────────────────────────────────────────────────────────────┤
│ SECTION 2: Currency                                         │
│ Shown when: session.FeatureEnabled("finance.multi_currency")│
├─────────────────────────────────────────────────────────────┤
│ SECTION 3: Dimensions                                       │
│ Cost Centre: shown always; required if setting = true       │
│ Project: shown if FeatureEnabled("finance.project_tracking")│
├─────────────────────────────────────────────────────────────┤
│ SECTION 4: Entry Lines (always shown)                       │
│ Account | Description | Debit | Credit                      │
├─────────────────────────────────────────────────────────────┤
│ SECTION 5: Approval Info (read-only after submission)       │
│ Shown when: FeatureEnabled("finance.transactions            │
│                             .approval_workflow")            │
│ AND total > settings.SettingDecimal("approval_threshold")   │
└─────────────────────────────────────────────────────────────┘
```

---

### Settings Screen — Data-Driven Form

Tenant admin settings screens are generated from `setting_definitions`. No hardcoded forms:

```go
func buildSettingsFields(defs []domain.SettingDefinitionWithValue) []any {
    var fields []any
    for _, def := range defs {
        field := map[string]any{
            "name":  def.SettingKey,
            "label": def.Label,
            "value": def.CurrentValue,
        }
        switch def.ValueType {
        case "bool":    field["type"] = "switch"
        case "decimal": field["type"] = "input-number"; field["precision"] = 2
        case "int":     field["type"] = "input-number"; field["precision"] = 0
        case "enum":    field["type"] = "select"; field["options"] = def.EnumOptions
        default:        field["type"] = "input-text"
        }
        if def.IsSystem {
            field["disabled"] = true
            field["hint"] = "Managed by Awo platform"
        }
        fields = append(fields, field)
    }
    return fields
}
```

Adding a new setting requires one SQL row in `setting_definitions`. The settings screen renders it automatically.

---

### Flags Screen — Module/Resource Toggles

```go
func ModuleFlagsSchema(deps *app.Deps) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := middleware.ContextSession(c)
        if !session.Can("settings.modules", "update") {
            return c.Status(403).JSON(response.Err("access denied"))
        }

        flags, _ := deps.Platform.Flags.ListForTenant(c.Context(),
            domain.ListFlagsParams{
                TenantID:      session.TenantID,
                ExcludeSystem: true,  // system flags not shown to tenant admins
            })

        return c.JSON(buildFlagsForm(flags))
    }
}
```

The flags form is a series of toggle switches, grouped by module. Each module has a master toggle; when a module is turned off, all its resource toggles are disabled in the UI (and the module flag check in the nav query handles the actual enforcement).

---

### BootService Interface

```go
type BootService interface {
    // Called by GET /schema/boot on every page load
    BuildAppShell(ctx, session *domain.ResolvedSession) (map[string]any, error)

    // Called by GET /schema/boot when no session exists — returns login form schema
    LoginSchema(message string) map[string]any
}
```

---

Next: [HTTP Middleware Chain](./12b-http-middleware.md)
