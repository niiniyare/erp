# Chapter 15: Service Integration Patterns

[← API Reference](./14-api-reference.md) | [Next: Finance Integration →](./16-finance-integration.md)

---

## Overview

Other ERP modules (Finance, HR, Inventory, etc.) consume configuration from the Settings Module rather than hardcoding values. This chapter covers the three main integration patterns and best practices.

---

## Dependency Injection Setup

```go
func setupSettingsServices(container *ServiceContainer) {
    configRepo := repository.NewConfigurationRepository(
        container.Store,
        container.Cache,
        container.Tracing,
        container.Metrics,
    )

    container.ConfigurationService = service.NewConfigurationService(
        configRepo,
        container.AuditService,
        container.Logger.WithFields(logger.Fields{"service": "settings"}),
        container.Tracing,
        container.Metrics,
    )

    container.TemplateService = service.NewTemplateService(
        configRepo,
        container.AuditService,
        container.Logger.WithFields(logger.Fields{"service": "templates"}),
        container.Tracing,
        container.Metrics,
    )
}
```

---

## Pattern 1: Direct Configuration Lookup

Use when a service needs a single configuration value for a business decision:

```go
func (s *FinanceService) ProcessPayment(ctx context.Context, payment *Payment) error {
    approvalConfig, err := s.configService.GetEffectiveConfiguration(
        ctx, &payment.EntityID, "finance", "approval_limit",
    )
    if err != nil {
        if errors.Is(err, domain.ErrConfigurationNotFound) {
            return s.processWithDefaultLimit(ctx, payment) // graceful degradation
        }
        return err
    }

    approvalLimit, _ := approvalConfig.Value.AsDecimal()
    if payment.Amount.GreaterThan(approvalLimit) {
        return s.requireApproval(ctx, payment)
    }
    return s.processDirectly(ctx, payment)
}
```

---

## Pattern 2: Batch Configuration Loading

Preferred for services that need multiple config values — avoids N separate database round-trips:

```go
func (s *FinanceService) getEntityConfiguration(
    ctx context.Context, entityID uuid.UUID,
) (*FinanceConfiguration, error) {

    // Single call returns all finance configs for this entity
    configs, err := s.configService.ListEffectiveConfigurations(ctx, &entityID, "finance")
    if err != nil {
        return nil, fmt.Errorf("failed to get finance configurations: %w", err)
    }

    config := &FinanceConfiguration{
        ApprovalLimit: decimal.NewFromInt(1000), // safe defaults
        InvoicePrefix: "INV-",
    }

    for _, cfg := range configs {
        switch cfg.Key.String() {
        case "approval_limit":
            if val, err := cfg.Value.AsDecimal(); err == nil {
                config.ApprovalLimit = val
            }
        case "invoice_prefix":
            if val, err := cfg.Value.AsString(); err == nil {
                config.InvoicePrefix = val
            }
        }
    }

    return config, nil
}
```

---

## Pattern 3: Template-Based Initial Setup

For tenant onboarding — apply an industry template automatically:

```go
func (s *TenantOnboardingService) SetupNewTenant(
    ctx context.Context, tenant *Tenant, industryType string,
) error {
    templates, err := s.templateService.ListTemplates(ctx, repository.TemplateFilters{
        Category: domain.TemplateCategoryIndustry,
        IsActive: true,
    })

    for _, template := range templates {
        if strings.Contains(template.Name, industryType) {
            _, err := s.templateService.ApplyTemplate(
                ctx, template.ID,
                repository.TemplateTarget{Type: "tenant", EntityID: &tenant.ID},
                repository.ApplyOptions{DryRun: false},
                tenant.CreatedBy,
            )
            return err
        }
    }
    return nil
}
```

---

## Graceful Degradation

Services must continue operating when configuration is unavailable:

```go
func (s *ServiceExample) getConfigWithFallback(
    ctx context.Context, entityID uuid.UUID, key string, defaultValue interface{},
) interface{} {
    config, err := s.configService.GetEffectiveConfiguration(
        ctx, &entityID, s.moduleName, domain.ConfigKey(key),
    )
    if err != nil {
        s.logger.WarnContext(ctx, "Configuration unavailable, using default", logger.Fields{
            "key": key, "default": defaultValue, "error": err.Error(),
        })
        return defaultValue
    }

    switch v := defaultValue.(type) {
    case string:
        if val, err := config.Value.AsString(); err == nil { return val }
    case int:
        if val, err := config.Value.AsInt(); err == nil { return val }
    case bool:
        if val, err := config.Value.AsBool(); err == nil { return val }
    }
    return defaultValue
}
```

---

## Caching Pattern

Services should cache resolved configurations to avoid repeated database calls:

```go
func (s *ServiceWithConfigCache) getCachedConfig(
    ctx context.Context, entityID uuid.UUID, module, key string,
) (*domain.Configuration, error) {
    cacheKey := fmt.Sprintf("%s:%s:%s", entityID, module, key)

    s.cacheMutex.RLock()
    if cached, exists := s.cache[cacheKey]; exists {
        s.cacheMutex.RUnlock()
        return cached, nil
    }
    s.cacheMutex.RUnlock()

    config, err := s.configService.GetEffectiveConfiguration(
        ctx, &entityID, domain.ModuleName(module), domain.ConfigKey(key),
    )
    if err != nil {
        return nil, err
    }

    s.cacheMutex.Lock()
    s.cache[cacheKey] = config
    s.cacheMutex.Unlock()

    // Auto-expire after cache duration
    go func() {
        time.Sleep(s.cacheExpiry)
        s.cacheMutex.Lock()
        delete(s.cache, cacheKey)
        s.cacheMutex.Unlock()
    }()

    return config, nil
}
```

---

## Error Handling

| Error | Recommended Action |
|---|---|
| `ErrConfigurationNotFound` | Use hardcoded default, log a warning |
| `ErrInvalidConfigurationValue` | Log warning, use default — the config data has a type mismatch |
| Service unavailable | Use cached value if available, else use default |
| Permission denied | Return `403` to the caller |

---

[← API Reference](./14-api-reference.md) | [Next: Finance Integration →](./16-finance-integration.md)
