# Chapter 16: Finance Module Integration

[← Service Integration Patterns](./15-service-integration.md) | [Next: HR Integration →](./17-hr-integration.md)

---

## Finance Configuration Keys

The Finance module relies on Settings for document formatting, approval workflows, and accounting rules. See [Chapter 6](./06-configuration-definitions.md) for the full list of finance configuration definitions.

---

## Loading Finance Configuration

```go
type FinanceSettingsIntegration struct {
    settingsService service.SettingsService
}

func (f *FinanceSettingsIntegration) GetFinanceConfig(
    ctx context.Context, tenantID tenant.ID, entityID *entity.ID,
) (*FinanceConfiguration, error) {

    configKeys := []service.ConfigurationKey{
        {Module: "finance", Key: "default_currency"},
        {Module: "finance", Key: "fiscal_year_start"},
        {Module: "finance", Key: "invoice_prefix"},
        {Module: "finance", Key: "auto_approval_limit"},
        {Module: "finance", Key: "payment_terms_default"},
        {Module: "finance", Key: "tax_calculation_method"},
        {Module: "finance", Key: "multi_currency_enabled"},
    }

    resolved, err := f.settingsService.ResolveConfigurations(ctx, service.ResolutionRequest{
        TenantID:        tenantID,
        EntityID:        entityID,
        Configurations:  configKeys,
        IncludeMetadata: true,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to resolve finance configurations: %w", err)
    }

    config := &FinanceConfiguration{TenantID: tenantID, EntityID: entityID}

    for _, cfg := range resolved.Configurations {
        switch cfg.ConfigKey {
        case "default_currency":
            config.DefaultCurrency = domain.Currency(cfg.Value.(string))
        case "fiscal_year_start":
            config.FiscalYearStart = int(cfg.Value.(float64))
        case "invoice_prefix":
            config.InvoicePrefix = cfg.Value.(string)
        case "auto_approval_limit":
            config.AutoApprovalLimit = decimal.NewFromFloat(cfg.Value.(float64))
        case "payment_terms_default":
            config.PaymentTermsDefault = cfg.Value.(string)
        case "tax_calculation_method":
            config.TaxCalculationMethod = cfg.Value.(string)
        case "multi_currency_enabled":
            config.MultiCurrencyEnabled = cfg.Value.(bool)
        }
    }

    return config, nil
}
```

---

## Document Sequence Integration

Entity-specific document sequences (invoice numbers, quote numbers, PO numbers) are driven by configuration. The entity's `entitystate.config` JSONB stores the sequence format; the `entities.settings` JSONB stores the prefix and reset rules.

```go
func (f *FinanceSettingsIntegration) GetDocumentSequenceConfig(
    ctx context.Context, entityID entity.ID,
) (*DocumentSequenceConfiguration, error) {

    configKeys := []service.ConfigurationKey{
        {Module: "finance", Key: "invoice_prefix"},
        {Module: "finance", Key: "quote_prefix"},
        {Module: "finance", Key: "po_prefix"},
        {Module: "finance", Key: "document_pad_length"},
        {Module: "finance", Key: "document_reset_frequency"},
        {Module: "finance", Key: "document_format_template"},
    }

    resolved, err := f.settingsService.ResolveConfigurations(ctx, service.ResolutionRequest{
        TenantID:       f.getTenantIDFromEntity(ctx, entityID),
        EntityID:       &entityID,
        Configurations: configKeys,
    })
    if err != nil {
        return nil, err
    }

    config := &DocumentSequenceConfiguration{}
    for _, cfg := range resolved.Configurations {
        switch cfg.ConfigKey {
        case "invoice_prefix":
            config.InvoicePrefix = cfg.Value.(string)
        case "quote_prefix":
            config.QuotePrefix = cfg.Value.(string)
        case "po_prefix":
            config.POPrefix = cfg.Value.(string)
        case "document_pad_length":
            config.PadLength = int(cfg.Value.(float64))
        case "document_reset_frequency":
            config.ResetFrequency = cfg.Value.(string)
        case "document_format_template":
            config.FormatTemplate = cfg.Value.(string)
        }
    }

    return config, nil
}
```

---

## Handling Configuration Changes

When finance configuration changes, downstream caches need invalidation and dependent services need notification:

```go
func (f *FinanceSettingsIntegration) HandleConfigurationChange(
    ctx context.Context, event ConfigurationChangedEvent,
) error {
    if event.Module != "finance" {
        return nil
    }

    // Invalidate finance configuration caches
    cachePatterns := []string{
        fmt.Sprintf("finance:config:%s:%s", event.TenantID, event.EntityID),
        fmt.Sprintf("finance:config:%s:*", event.TenantID),
    }
    for _, pattern := range cachePatterns {
        f.cache.DeletePattern(ctx, pattern)
    }

    // Notify finance service subscribers
    return f.eventBus.Publish(ctx, "finance.configuration.changed",
        FinanceConfigurationChangedEvent{
            TenantID:  event.TenantID,
            EntityID:  event.EntityID,
            ConfigKey: event.ConfigKey,
            OldValue:  event.OldValue,
            NewValue:  event.NewValue,
        })
}
```

---

## Industry-Specific Finance Templates

Apply a finance template appropriate for the industry during onboarding:

```go
func (f *FinanceSettingsIntegration) ApplyFinanceTemplate(
    ctx context.Context, templateType FinanceTemplateType, targetID string,
) error {
    templateMap := map[FinanceTemplateType]string{
        FinanceTemplateManufacturing: "finance-manufacturing-standard",
        FinanceTemplateServices:      "finance-services-standard",
        FinanceTemplateRetail:        "finance-retail-standard",
        FinanceTemplateNonProfit:     "finance-nonprofit-standard",
    }

    templateID := templateMap[templateType]

    result, err := f.settingsService.ApplyTemplate(ctx, service.TemplateApplicationRequest{
        TemplateID: templateID,
        TargetType: "entity",
        TargetID:   targetID,
        Options: service.ApplicationOptions{
            PreserveExisting:     true,
            ConflictResolution:   "merge",
            ValidateBusinessRules: true,
        },
    })
    if err != nil {
        return err
    }

    for _, conflict := range result.Conflicts {
        if err := f.validateFinanceConfigConflict(conflict); err != nil {
            return fmt.Errorf("finance configuration conflict: %w", err)
        }
    }

    return nil
}
```

---

## Common Finance Configuration Scenarios

| Scenario | Config Key | Action |
|---|---|---|
| New branch needs its own invoice series | `finance.invoice_prefix` | Set entity override |
| Raise org-wide approval limit | `finance.auto_approval_limit` | Set tenant config |
| Enable multi-currency for one entity | `finance.multi_currency_enabled` | Set entity override |
| Change fiscal year (org-wide) | `finance.fiscal_year_start` | Set tenant config |
| Roll out manufacturing template | — | Apply `finance-manufacturing-standard` template |

---

[← Service Integration Patterns](./15-service-integration.md) | [Next: HR Integration →](./17-hr-integration.md)
