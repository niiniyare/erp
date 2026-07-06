---
title: "Feature Flags Platform Module"
id: mod-013
status: accepted
category: SPEC
stability: STABLE
audience: [module-authors, framework-authors, operators]
since: "1.0"
normative-level: normative
related:
  - "[Platform Modules](platform-modules.md)"
  - "[Feature Flags Configuration](../12-configuration/feature-flags-config.md)"
  - "[Settings Patterns](settings-patterns.md)"
  - "[Awo Glossary](../GLOSSARY.md)"
---

# Feature Flags Platform Module

**MOD-013 | Status: Accepted | Stability: Stable**

The Feature Flags module provides per-tenant, per-user boolean capability switches that take effect without deployment.

---

## 1. Purpose and Scope

Feature flags control:
- Activating a new module for a specific tenant (without enabling for all)
- Gradual rollout of a risky code path (tenant by tenant)
- UI elements that are too early for general release
- Emergency circuit-breaker (disable a broken feature without rollback)

Feature flags are **boolean only**. For numeric/string parameters, use the [Settings module](settings-patterns.md).

---

## 2. Flag Entity

```go
var FeatureFlagDefinition = definition.SystemDefinition{
    Name:   "platform_feature_flag",
    Module: "flags",
    Fields: []definition.FieldDef{
        {Name: "key",           Type: definition.FieldData, Required: true, Immutable: true, Unique: true},
            // e.g. "inventory.lot_tracking", "finance.etims_integration"
        {Name: "label",         Type: definition.FieldData, Required: true},
        {Name: "description",   Type: definition.FieldSmallText},
        {Name: "default_value", Type: definition.FieldBool, Default: false},
        {Name: "lifecycle",     Type: definition.FieldSelect,
            Options: []string{"Experimental", "Beta", "GA", "Deprecated"}, Default: "Experimental"},
        {Name: "module",        Type: definition.FieldData},  // owning module name
        {Name: "since_version", Type: definition.FieldData},  // e.g. "1.2"
    },
    Permissions: definition.PermissionSet{
        Create: []string{"role:platform-admin"},
        Read:   []string{"role:platform-admin", "role:tenant.admin"},
        Write:  []string{"role:platform-admin"},
        Delete: []string{"role:platform-admin"},
    },
}
```

### Flag Override Entity (per-tenant/user)

```go
var FlagOverrideDefinition = definition.SystemDefinition{
    Name:   "platform_flag_override",
    Module: "flags",
    Fields: []definition.FieldDef{
        {Name: "flag_key",    Type: definition.FieldData, Required: true, Immutable: true},
        {Name: "scope",       Type: definition.FieldSelect,
            Options: []string{"tenant", "user"}, Required: true, Immutable: true},
        {Name: "scope_id",    Type: definition.FieldData, Required: true, Immutable: true},
            // tenant UUID for scope=tenant, user UUID for scope=user
        {Name: "value",       Type: definition.FieldBool, Required: true},
        {Name: "set_by",      Type: definition.FieldLink, LinkTarget: "iam_user"},
        {Name: "reason",      Type: definition.FieldSmallText},
        {Name: "expires_at",  Type: definition.FieldDateTime},  // optional expiry for rollouts
    },
    Permissions: definition.PermissionSet{
        Create: []string{"role:platform-admin", "role:tenant.admin"},
        Read:   []string{"role:platform-admin", "role:tenant.admin"},
        Write:  []string{"role:platform-admin", "role:tenant.admin"},
        Delete: []string{"role:platform-admin", "role:tenant.admin"},
    },
}
```

---

## 3. Evaluation Logic

Evaluation order (highest priority wins):

```
User override → Tenant override → System default (from flag definition)
```

```go
// FlagsService evaluates a flag for the current actor
func (s *FlagsService) Evaluate(ctx context.Context, key string) (bool, error) {
    actor := session.ActorFromContext(ctx)
    tenantID := session.TenantIDFromContext(ctx)

    // Check Redis cache first
    cacheKey := flagCacheKey(key, tenantID.String(), actor.UserID.String())
    if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
        return cached == "1", nil
    }

    var result bool

    // User override
    userOverride, err := s.getOverride(ctx, key, "user", actor.UserID.String())
    if err == nil {
        result = userOverride
    } else {
        // Tenant override
        tenantOverride, err := s.getOverride(ctx, key, "tenant", tenantID.String())
        if err == nil {
            result = tenantOverride
        } else {
            // System default
            flag, err := s.getFlagDef(ctx, key)
            if err != nil {
                return false, fmt.Errorf("FlagsService.Evaluate: get flag %q: %w", key, err)
            }
            result = flag.DefaultValue
        }
    }

    // Cache for 5 minutes
    val := "0"
    if result {
        val = "1"
    }
    s.Redis.Set(ctx, cacheKey, val, 5*time.Minute)

    return result, nil
}

func flagCacheKey(key, tenantID, userID string) string {
    raw := fmt.Sprintf("%s|%s|%s", key, tenantID, userID)
    hash := sha256.Sum256([]byte(raw))
    return fmt.Sprintf("eval:%x", hash[:16])
}
```

---

## 4. Declaring Flags in Modules

```go
// internal/core/inventory/flags.go

var (
    FlagLotTracking = definition.FeatureFlag{
        Key:         "inventory.lot_tracking",
        Label:       "Lot / Batch Tracking",
        Description: "Enable serial number and batch tracking on stock moves.",
        Default:     false,
        Lifecycle:   definition.FlagLifecycleBeta,
        Module:      "inventory",
    }

    FlagReorderAlerts = definition.FeatureFlag{
        Key:         "inventory.reorder_alerts",
        Label:       "Reorder Point Alerts",
        Description: "Notify purchasing when stock falls below reorder point.",
        Default:     true,
        Lifecycle:   definition.FlagLifecycleGA,
        Module:      "inventory",
    }
)

func init() {
    definition.RegisterFlag(&FlagLotTracking)
    definition.RegisterFlag(&FlagReorderAlerts)
}
```

---

## 5. Using Flags in Application Code

### In HTTP Handlers

```go
func (h *InventoryHandler) handleCreateStockMove(c *fiber.Ctx) error {
    ctx := c.UserContext()

    lotEnabled, err := h.Flags.Evaluate(ctx, "inventory.lot_tracking")
    if err != nil {
        // Flag evaluation error → fail open (use default)
        slog.Warn("flag evaluation failed, using default", "flag", "inventory.lot_tracking", "err", err)
        lotEnabled = false
    }

    if lotEnabled {
        // Include lot/serial number fields in response schema
    }
    // ...
}
```

### In SDUI Page Builders

```go
func BuildStockMoveForm(ctx context.Context, actor session.Actor) ([]byte, error) {
    psc := sdui.NewPageSchemaContext(ctx, actor)

    lotEnabled, _ := psc.Flag("inventory.lot_tracking")

    fields := []amis.FormField{
        amis.SelectField("product", "Product"),
        amis.NumberField("quantity", "Quantity"),
    }

    if lotEnabled {
        fields = append(fields, amis.DataField("lot_number", "Lot / Serial Number"))
    }

    // ...
}
```

### In Hooks

```go
func (h *LotTrackingGuard) BeforeCreate(ctx context.Context, record *definition.EntityRecord) error {
    enabled, err := h.Flags.Evaluate(ctx, "inventory.lot_tracking")
    if err != nil || !enabled {
        return nil  // Lot tracking off — skip validation
    }

    if record.Fields["lot_number"] == nil || record.Fields["lot_number"] == "" {
        return &definition.ValidationError{
            Fields: map[string]string{
                "lot_number": "Lot/serial number required when lot tracking is enabled",
            },
        }
    }
    return nil
}
```

---

## 6. Cache Invalidation

Flag changes take effect within 5 minutes (cache TTL). For immediate effect:

```
POST /api/v1/platform/flags/invalidate-cache
Body: {"tenant_id": "...", "flag_key": "inventory.lot_tracking"}
```

This deletes the matching Redis keys for all affected scope combinations.

---

## 7. Flag Lifecycle Management

| Lifecycle | Meaning | Action |
|---|---|---|
| `Experimental` | Internal/early testing only | Not visible in tenant admin UI |
| `Beta` | Opt-in for tenants | Visible in tenant admin UI, off by default |
| `GA` | Generally available | On by default; tenant can disable |
| `Deprecated` | Scheduled for removal | Displayed with removal notice in admin UI |

When removing a deprecated flag:
1. Search all callsites: `grep -r "flag_key"` in codebase
2. Remove conditional logic — execute the guarded code path unconditionally
3. Delete flag definition and `RegisterFlag` call
4. Write migration to delete `platform_feature_flag` and `platform_flag_override` records
5. Deploy

---

## Related Documents

- [Platform Modules](platform-modules.md) — flags module in context
- [Feature Flags Configuration](../12-configuration/feature-flags-config.md) — declaring flags
- [Settings Patterns](settings-patterns.md) — for non-boolean parameters
- [Metrics Reference](../13-observability/metrics-reference.md) — `cache_hit_total` for flag cache monitoring
