# Chapter 18: Inventory Module Integration

[← HR Integration](./17-hr-integration.md) | [Next: Event-Driven Integration →](./19-event-driven-integration.md)

---

## Inventory Configuration Keys

Inventory relies on configuration for valuation methods, tracking requirements, and reorder automation. Warehouses (entities) frequently override tenant defaults because different facilities have different operational constraints.

---

## Loading Warehouse Configuration

```go
type InventorySettingsIntegration struct {
    settingsService  service.SettingsService
    inventoryService InventoryService
    valuationService ValuationService
}

func (i *InventorySettingsIntegration) GetInventoryConfiguration(
    ctx context.Context, tenantID tenant.ID, warehouseID *entity.ID,
) (*InventoryConfiguration, error) {

    configKeys := []service.ConfigurationKey{
        {Module: "inventory", Key: "default_valuation_method"},
        {Module: "inventory", Key: "allow_negative_stock"},
        {Module: "inventory", Key: "cycle_count_frequency"},
        {Module: "inventory", Key: "auto_reorder_enabled"},
        {Module: "inventory", Key: "lead_time_buffer_days"},
        {Module: "inventory", Key: "lot_tracking_required"},
        {Module: "inventory", Key: "serial_tracking_required"},
        {Module: "inventory", Key: "location_tracking_required"},
    }

    resolved, err := i.settingsService.ResolveConfigurations(ctx, service.ResolutionRequest{
        TenantID:       tenantID,
        EntityID:       warehouseID,
        Configurations: configKeys,
    })
    if err != nil {
        return nil, err
    }

    config := &InventoryConfiguration{}

    for _, cfg := range resolved.Configurations {
        switch cfg.ConfigKey {
        case "default_valuation_method":
            config.ValuationMethod = ValuationMethod(cfg.Value.(string))
        case "allow_negative_stock":
            config.AllowNegativeStock = cfg.Value.(bool)
        case "cycle_count_frequency":
            config.CycleCountFrequency = cfg.Value.(string)
        case "auto_reorder_enabled":
            config.AutoReorderEnabled = cfg.Value.(bool)
        case "lead_time_buffer_days":
            config.LeadTimeBufferDays = int(cfg.Value.(float64))
        case "lot_tracking_required":
            config.LotTrackingRequired = cfg.Value.(bool)
        case "serial_tracking_required":
            config.SerialTrackingRequired = cfg.Value.(bool)
        case "location_tracking_required":
            config.LocationTrackingRequired = cfg.Value.(bool)
        }
    }

    return config, nil
}
```

---

## Reorder Point Check

```go
func (s *InventoryService) CheckReorderPoint(
    ctx context.Context, item *InventoryItem,
) (bool, error) {

    reorderConfig, err := s.configService.GetEffectiveConfiguration(
        ctx, &item.EntityID, "inventory", "reorder_point_days",
    )
    if err != nil {
        s.logger.WarnContext(ctx, "Failed to get reorder configuration, using default",
            logger.Fields{"item_id": item.ID, "error": err.Error()})
        return s.checkReorderWithDays(item, 30), nil
    }

    reorderDays, err := reorderConfig.Value.AsInt()
    if err != nil {
        reorderDays = 30 // fallback
    }

    return s.checkReorderWithDays(item, reorderDays), nil
}

func (s *InventoryService) checkReorderWithDays(item *InventoryItem, days int) bool {
    if item.AverageDailyUsage <= 0 {
        return false
    }
    daysOfStock := float64(item.CurrentStock) / item.AverageDailyUsage
    return daysOfStock <= float64(days)
}
```

---

## Reacting to Inventory Configuration Changes

Certain configuration changes have significant operational impact and need immediate action:

```go
func (i *InventorySettingsIntegration) HandleInventoryConfigChange(
    ctx context.Context, event ConfigurationChangedEvent,
) error {
    if event.Module != "inventory" {
        return nil
    }

    switch event.ConfigKey {
    case "default_valuation_method":
        return i.handleValuationMethodChange(ctx, event)
    case "lot_tracking_required":
        return i.handleTrackingRequirementChange(ctx, event, "lot")
    case "serial_tracking_required":
        return i.handleTrackingRequirementChange(ctx, event, "serial")
    case "auto_reorder_enabled":
        return i.handleAutoReorderChange(ctx, event)
    }

    return nil
}

func (i *InventorySettingsIntegration) handleValuationMethodChange(
    ctx context.Context, event ConfigurationChangedEvent,
) error {
    oldMethod := ValuationMethod(event.OldValue.(string))
    newMethod := ValuationMethod(event.NewValue.(string))

    if oldMethod == newMethod {
        return nil
    }

    // Schedule forward-looking valuation recalculation
    return i.valuationService.ScheduleRecalculation(ctx, ValuationRecalculationRequest{
        TenantID:           event.TenantID,
        EntityID:           event.EntityID,
        FromMethod:         oldMethod,
        ToMethod:           newMethod,
        EffectiveDate:      time.Now(),
        RecalculateHistory: false,
    })
}

func (i *InventorySettingsIntegration) handleTrackingRequirementChange(
    ctx context.Context, event ConfigurationChangedEvent, trackingType string,
) error {
    if event.NewValue.(bool) {
        // Requirement enabled: validate existing inventory for compliance
        return i.validateInventoryForTracking(ctx, event.TenantID, event.EntityID, trackingType)
    }
    // Requirement disabled: clean up tracking metadata if needed
    return i.cleanupTrackingData(ctx, event.TenantID, event.EntityID, trackingType)
}
```

---

## Common Inventory Configuration Scenarios

| Scenario | Config Key | Level |
|---|---|---|
| Transit warehouse allows negative stock | `inventory.allow_negative_stock = true` | Entity (transit WH) |
| Switch org to AVCO valuation | `inventory.default_valuation_method = "AVCO"` | Tenant |
| Enable lot tracking for pharma warehouse | `inventory.lot_tracking_required = true` | Entity |
| Disable auto-reorder during fiscal year close | `inventory.auto_reorder_enabled = false` | Tenant |
| Extend lead time buffer for remote location | `inventory.lead_time_buffer_days = 14` | Entity |

---

[← HR Integration](./17-hr-integration.md) | [Next: Event-Driven Integration →](./19-event-driven-integration.md)
