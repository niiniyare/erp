# ERP Settings Module - Integration Guide

**Version**: 1.0  
**Date**: September 2025  
**Status**: Technical Integration Manual  

---

## Table of Contents

1. [Integration Architecture](#integration-architecture)
2. [ERP Module Integration](#erp-module-integration)
3. [Configuration Resolution Integration](#configuration-resolution-integration)
4. [Template System Integration](#template-system-integration)
5. [Event-Driven Integration](#event-driven-integration)
6. [Real-time Configuration Updates](#real-time-configuration-updates)
7. [Cache Integration Patterns](#cache-integration-patterns)
8. [Feature Flag Integration](#feature-flag-integration)
9. [Integration Testing](#integration-testing)
10. [Troubleshooting Integration Issues](#troubleshooting-integration-issues)

---

## Integration Architecture

### **Multi-Layer Integration Strategy**

```
┌─────────────────────────────────────────────────────────────────┐
│                    External Configuration Sources              │
│  Environment Variables, Config Management Tools, CI/CD        │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Settings API Gateway                        │
│  Request Routing, Authentication, Rate Limiting, Caching      │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Configuration Event Bus                   │
│   Redis Streams, Change Notifications, Cache Invalidation     │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Settings Module Core                         │
│   Configuration Resolution, Template Management, Validation    │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Consuming ERP Modules                         │
│  Finance, HR, Inventory, Sales, CRM, Project Management       │
└─────────────────────────────────────────────────────────────────┘
```

### **Integration Patterns**
```yaml
Pattern Types:
  Configuration Resolution: Real-time configuration lookup with caching
  Event-Driven Updates: Asynchronous configuration change notifications
  Template Application: Bulk configuration deployment and management
  Feature-Gate Integration: Configuration availability based on feature flags
  
Communication Protocols:
  Internal: gRPC for low-latency config resolution, Redis for events
  External: REST APIs for configuration management, WebHooks for notifications
  
Data Flow:
  Read-Heavy: Optimized for frequent configuration lookups
  Write-Light: Batch operations for configuration changes
  Cache-First: Multi-layer caching for performance optimization
```

---

## ERP Module Integration

### **Finance Module Integration**

#### **Configuration-Driven Financial Settings**
```go
// Finance module configuration integration
package integration

import (
    "context"
    "fmt"
    
    "awo.so/internal/core/settings/service"
    "awo.so/internal/core/finance/domain"
)

type FinanceSettingsIntegration struct {
    settingsService service.SettingsService
    cache          cache.Service
    eventBus       EventBus
}

// Get finance configuration with inheritance
func (f *FinanceSettingsIntegration) GetFinanceConfig(ctx context.Context, tenantID tenant.ID, entityID *entity.ID) (*FinanceConfiguration, error) {
    // Define required configurations
    configKeys := []service.ConfigurationKey{
        {Module: "finance", Key: "default_currency"},
        {Module: "finance", Key: "fiscal_year_start"},
        {Module: "finance", Key: "invoice_prefix"},
        {Module: "finance", Key: "auto_approval_limit"},
        {Module: "finance", Key: "payment_terms_default"},
        {Module: "finance", Key: "tax_calculation_method"},
        {Module: "finance", Key: "multi_currency_enabled"},
    }
    
    // Resolve configurations with inheritance
    resolved, err := f.settingsService.ResolveConfigurations(ctx, service.ResolutionRequest{
        TenantID:       tenantID,
        EntityID:       entityID,
        Configurations: configKeys,
        IncludeMetadata: true,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to resolve finance configurations: %w", err)
    }
    
    // Map to strongly-typed finance configuration
    config := &FinanceConfiguration{
        TenantID: tenantID,
        EntityID: entityID,
    }
    
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

// Handle configuration changes with real-time updates
func (f *FinanceSettingsIntegration) HandleConfigurationChange(ctx context.Context, event ConfigurationChangedEvent) error {
    if event.Module != "finance" {
        return nil // Not a finance configuration
    }
    
    // Invalidate cached configurations
    cacheKeys := []string{
        fmt.Sprintf("finance:config:%s:%s", event.TenantID, event.EntityID),
        fmt.Sprintf("finance:config:%s:*", event.TenantID),
    }
    
    for _, key := range cacheKeys {
        f.cache.DeletePattern(ctx, key)
    }
    
    // Publish finance-specific event for interested services
    financeEvent := FinanceConfigurationChangedEvent{
        TenantID:    event.TenantID,
        EntityID:    event.EntityID,
        ConfigKey:   event.ConfigKey,
        OldValue:    event.OldValue,
        NewValue:    event.NewValue,
        ChangedAt:   event.ChangedAt,
        ChangedBy:   event.ChangedBy,
    }
    
    return f.eventBus.Publish(ctx, "finance.configuration.changed", financeEvent)
}

// Apply finance template with business rule validation
func (f *FinanceSettingsIntegration) ApplyFinanceTemplate(ctx context.Context, templateType FinanceTemplateType, targetID string) error {
    templateMap := map[FinanceTemplateType]string{
        FinanceTemplateManufacturing: "finance-manufacturing-standard",
        FinanceTemplateServices:      "finance-services-standard", 
        FinanceTemplateRetail:        "finance-retail-standard",
        FinanceTemplateNonProfit:     "finance-nonprofit-standard",
    }
    
    templateID, exists := templateMap[templateType]
    if !exists {
        return fmt.Errorf("unknown finance template type: %v", templateType)
    }
    
    // Apply template with finance-specific conflict resolution
    result, err := f.settingsService.ApplyTemplate(ctx, service.TemplateApplicationRequest{
        TemplateID: templateID,
        TargetType: "entity",
        TargetID:   targetID,
        Options: service.ApplicationOptions{
            PreserveExisting:    true,
            ConflictResolution:  "merge",
            ValidateBusinessRules: true,
        },
    })
    if err != nil {
        return err
    }
    
    // Handle any conflicts with business rule validation
    for _, conflict := range result.Conflicts {
        if err := f.validateFinanceConfigConflict(conflict); err != nil {
            return fmt.Errorf("finance configuration conflict: %w", err)
        }
    }
    
    return nil
}
```

#### **Document Sequence Configuration Integration**
```go
//  document sequence with configuration support
func (f *FinanceSettingsIntegration) ConfigureDocumentSequences(ctx context.Context, entityID entity.ID) error {
    // Get entity-specific document sequence configurations
    seqConfig, err := f.GetDocumentSequenceConfig(ctx, entityID)
    if err != nil {
        return err
    }
    
    // Apply configuration to entity state
    sequences := []DocumentSequenceConfig{
        {
            DocumentType: "INVOICE",
            Prefix:       seqConfig.InvoicePrefix,
            PadLength:    seqConfig.PadLength,
            ResetFreq:    seqConfig.ResetFrequency,
            Format:       seqConfig.FormatTemplate,
        },
        {
            DocumentType: "QUOTE",
            Prefix:       seqConfig.QuotePrefix,
            PadLength:    seqConfig.PadLength,
            ResetFreq:    seqConfig.ResetFrequency,
            Format:       seqConfig.FormatTemplate,
        },
        {
            DocumentType: "PURCHASE_ORDER",
            Prefix:       seqConfig.POPrefix,
            PadLength:    seqConfig.PadLength,
            ResetFreq:    seqConfig.ResetFrequency,
            Format:       seqConfig.FormatTemplate,
        },
    }
    
    return f.applySequenceConfigurations(ctx, entityID, sequences)
}

func (f *FinanceSettingsIntegration) GetDocumentSequenceConfig(ctx context.Context, entityID entity.ID) (*DocumentSequenceConfiguration, error) {
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

### **HR Module Integration**

#### **Payroll Configuration Integration**
```go
// HR module payroll configuration
type HRSettingsIntegration struct {
    settingsService service.SettingsService
    payrollService  PayrollService
    eventSubscriber EventSubscriber
}

func (h *HRSettingsIntegration) GetPayrollConfiguration(ctx context.Context, tenantID tenant.ID, entityID *entity.ID) (*PayrollConfiguration, error) {
    configKeys := []service.ConfigurationKey{
        {Module: "hr", Key: "default_pay_frequency"},
        {Module: "hr", Key: "overtime_threshold"},
        {Module: "hr", Key: "overtime_multiplier"},
        {Module: "hr", Key: "pay_period_start_day"},
        {Module: "hr", Key: "tax_jurisdiction"},
        {Module: "hr", Key: "benefits_waiting_period"},
        {Module: "hr", Key: "vacation_accrual_method"},
        {Module: "hr", Key: "sick_leave_annual_hours"},
    }
    
    resolved, err := h.settingsService.ResolveConfigurations(ctx, service.ResolutionRequest{
        TenantID:       tenantID,
        EntityID:       entityID,
        Configurations: configKeys,
    })
    if err != nil {
        return nil, err
    }
    
    config := &PayrollConfiguration{
        TenantID: tenantID,
        EntityID: entityID,
    }
    
    // Map resolved configurations to HR domain model
    for _, cfg := range resolved.Configurations {
        switch cfg.ConfigKey {
        case "default_pay_frequency":
            config.PayFrequency = PayFrequency(cfg.Value.(string))
        case "overtime_threshold":
            config.OvertimeThreshold = int(cfg.Value.(float64))
        case "overtime_multiplier":
            config.OvertimeMultiplier = decimal.NewFromFloat(cfg.Value.(float64))
        case "pay_period_start_day":
            config.PayPeriodStartDay = cfg.Value.(string)
        case "tax_jurisdiction":
            config.TaxJurisdiction = cfg.Value.(string)
        case "benefits_waiting_period":
            config.BenefitsWaitingPeriod = int(cfg.Value.(float64))
        case "vacation_accrual_method":
            config.VacationAccrualMethod = cfg.Value.(string)
        case "sick_leave_annual_hours":
            config.SickLeaveAnnualHours = int(cfg.Value.(float64))
        }
    }
    
    return config, nil
}

// Listen for HR configuration changes and update payroll calculations
func (h *HRSettingsIntegration) StartConfigurationListener(ctx context.Context) error {
    return h.eventSubscriber.Subscribe(ctx, "settings.configuration.changed", func(event ConfigurationChangedEvent) error {
        if event.Module != "hr" {
            return nil
        }
        
        // Check if this affects payroll calculations
        payrollAffectingConfigs := []string{
            "default_pay_frequency",
            "overtime_threshold", 
            "overtime_multiplier",
            "tax_jurisdiction",
        }
        
        for _, affectingConfig := range payrollAffectingConfigs {
            if event.ConfigKey == affectingConfig {
                return h.handlePayrollConfigurationChange(ctx, event)
            }
        }
        
        return nil
    })
}

func (h *HRSettingsIntegration) handlePayrollConfigurationChange(ctx context.Context, event ConfigurationChangedEvent) error {
    // Get affected employees for the entity/tenant
    employees, err := h.payrollService.GetActiveEmployees(ctx, event.TenantID, event.EntityID)
    if err != nil {
        return err
    }
    
    // Schedule payroll recalculation for affected employees
    for _, employee := range employees {
        recalcRequest := PayrollRecalculationRequest{
            EmployeeID:    employee.ID,
            TenantID:      event.TenantID,
            EntityID:      event.EntityID,
            ConfigChange:  event.ConfigKey,
            EffectiveDate: time.Now(),
            Reason:        "Configuration change",
        }
        
        err := h.payrollService.ScheduleRecalculation(ctx, recalcRequest)
        if err != nil {
            // Log error but continue processing other employees
            log.Printf("Failed to schedule payroll recalculation for employee %s: %v", employee.ID, err)
        }
    }
    
    return nil
}
```

### **Inventory Module Integration**

#### **Valuation and Warehouse Configuration**
```go
// Inventory module configuration integration
type InventorySettingsIntegration struct {
    settingsService    service.SettingsService
    inventoryService   InventoryService
    warehouseService   WarehouseService
    valuationService   ValuationService
}

func (i *InventorySettingsIntegration) GetInventoryConfiguration(ctx context.Context, tenantID tenant.ID, warehouseID *entity.ID) (*InventoryConfiguration, error) {
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
        EntityID:       warehouseID, // Warehouse as entity context
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

// Handle inventory configuration changes with business impact analysis
func (i *InventorySettingsIntegration) HandleInventoryConfigChange(ctx context.Context, event ConfigurationChangedEvent) error {
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

func (i *InventorySettingsIntegration) handleValuationMethodChange(ctx context.Context, event ConfigurationChangedEvent) error {
    oldMethod := ValuationMethod(event.OldValue.(string))
    newMethod := ValuationMethod(event.NewValue.(string))
    
    if oldMethod == newMethod {
        return nil
    }
    
    // Schedule valuation recalculation for all inventory items
    recalcRequest := ValuationRecalculationRequest{
        TenantID:         event.TenantID,
        EntityID:         event.EntityID,
        FromMethod:       oldMethod,
        ToMethod:         newMethod,
        EffectiveDate:    time.Now(),
        RecalculateHistory: false, // Only forward-looking
    }
    
    return i.valuationService.ScheduleRecalculation(ctx, recalcRequest)
}

func (i *InventorySettingsIntegration) handleTrackingRequirementChange(ctx context.Context, event ConfigurationChangedEvent, trackingType string) error {
    newRequirement := event.NewValue.(bool)
    
    if newRequirement {
        // Tracking requirement enabled - validate existing inventory
        return i.validateInventoryForTracking(ctx, event.TenantID, event.EntityID, trackingType)
    } else {
        // Tracking requirement disabled - clean up tracking data if needed
        return i.cleanupTrackingData(ctx, event.TenantID, event.EntityID, trackingType)
    }
}
```

---

## Configuration Resolution Integration

### **High-Performance Configuration Client**
```go
// High-performance configuration client with caching
package client

import (
    "context"
    "sync"
    "time"
    
    "github.com/patrickmn/go-cache"
    "go.uber.org/zap"
)

type ConfigurationClient struct {
    settingsService service.SettingsService
    cache          *cache.Cache
    logger         *zap.Logger
    
    // Configuration subscriptions
    subscriptions map[string][]ConfigurationSubscriber
    mu           sync.RWMutex
}

type ConfigurationSubscriber interface {
    OnConfigurationChanged(ctx context.Context, change ConfigurationChange) error
}

type ConfigurationChange struct {
    TenantID   tenant.ID
    EntityID   *entity.ID
    Module     string
    ConfigKey  string
    OldValue   interface{}
    NewValue   interface{}
    ChangedAt  time.Time
}

func NewConfigurationClient(settingsService service.SettingsService, logger *zap.Logger) *ConfigurationClient {
    return &ConfigurationClient{
        settingsService: settingsService,
        cache:          cache.New(5*time.Minute, 10*time.Minute),
        logger:         logger,
        subscriptions:  make(map[string][]ConfigurationSubscriber),
    }
}

// Get configuration with intelligent caching
func (c *ConfigurationClient) GetConfiguration(ctx context.Context, tenantID tenant.ID, entityID *entity.ID, module, key string) (interface{}, error) {
    cacheKey := c.buildCacheKey(tenantID, entityID, module, key)
    
    // Check cache first
    if cached, found := c.cache.Get(cacheKey); found {
        c.logger.Debug("Configuration cache hit", 
            zap.String("cache_key", cacheKey))
        return cached, nil
    }
    
    // Resolve from settings service
    resolved, err := c.settingsService.ResolveConfiguration(ctx, service.ResolutionRequest{
        TenantID: tenantID,
        EntityID: entityID,
        Module:   module,
        Key:      key,
    })
    if err != nil {
        return nil, err
    }
    
    // Cache the result with appropriate TTL
    ttl := c.calculateCacheTTL(resolved.Source)
    c.cache.Set(cacheKey, resolved.Value, ttl)
    
    c.logger.Debug("Configuration resolved and cached",
        zap.String("cache_key", cacheKey),
        zap.String("source", resolved.Source),
        zap.Duration("ttl", ttl))
    
    return resolved.Value, nil
}

// Batch configuration resolution for efficiency
func (c *ConfigurationClient) GetConfigurations(ctx context.Context, req BatchConfigurationRequest) (map[string]interface{}, error) {
    results := make(map[string]interface{})
    uncachedKeys := make([]service.ConfigurationKey, 0)
    
    // Check cache for each configuration
    for _, configKey := range req.ConfigurationKeys {
        cacheKey := c.buildCacheKey(req.TenantID, req.EntityID, configKey.Module, configKey.Key)
        
        if cached, found := c.cache.Get(cacheKey); found {
            fullKey := fmt.Sprintf("%s.%s", configKey.Module, configKey.Key)
            results[fullKey] = cached
        } else {
            uncachedKeys = append(uncachedKeys, configKey)
        }
    }
    
    // Resolve uncached configurations in batch
    if len(uncachedKeys) > 0 {
        resolved, err := c.settingsService.ResolveConfigurations(ctx, service.ResolutionRequest{
            TenantID:       req.TenantID,
            EntityID:       req.EntityID,
            Configurations: uncachedKeys,
        })
        if err != nil {
            return nil, err
        }
        
        // Add resolved configurations to cache and results
        for _, cfg := range resolved.Configurations {
            cacheKey := c.buildCacheKey(req.TenantID, req.EntityID, cfg.Module, cfg.ConfigKey)
            fullKey := fmt.Sprintf("%s.%s", cfg.Module, cfg.ConfigKey)
            
            ttl := c.calculateCacheTTL(cfg.Source)
            c.cache.Set(cacheKey, cfg.Value, ttl)
            results[fullKey] = cfg.Value
        }
    }
    
    return results, nil
}

// Subscribe to configuration changes
func (c *ConfigurationClient) Subscribe(pattern string, subscriber ConfigurationSubscriber) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.subscriptions[pattern] = append(c.subscriptions[pattern], subscriber)
}

// Handle configuration change notifications
func (c *ConfigurationClient) HandleConfigurationChange(ctx context.Context, event ConfigurationChangedEvent) error {
    change := ConfigurationChange{
        TenantID:  event.TenantID,
        EntityID:  event.EntityID,
        Module:    event.Module,
        ConfigKey: event.ConfigKey,
        OldValue:  event.OldValue,
        NewValue:  event.NewValue,
        ChangedAt: event.ChangedAt,
    }
    
    // Invalidate cache
    cacheKey := c.buildCacheKey(event.TenantID, event.EntityID, event.Module, event.ConfigKey)
    c.cache.Delete(cacheKey)
    
    // Invalidate related cache patterns
    c.invalidateRelatedCache(event.TenantID, event.EntityID, event.Module)
    
    // Notify subscribers
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    for pattern, subscribers := range c.subscriptions {
        if c.matchesPattern(pattern, event.Module, event.ConfigKey) {
            for _, subscriber := range subscribers {
                go func(sub ConfigurationSubscriber) {
                    if err := sub.OnConfigurationChanged(ctx, change); err != nil {
                        c.logger.Error("Configuration change notification failed",
                            zap.String("pattern", pattern),
                            zap.Error(err))
                    }
                }(subscriber)
            }
        }
    }
    
    return nil
}

func (c *ConfigurationClient) buildCacheKey(tenantID tenant.ID, entityID *entity.ID, module, key string) string {
    if entityID != nil {
        return fmt.Sprintf("config:%s:%s:%s:%s", tenantID, *entityID, module, key)
    }
    return fmt.Sprintf("config:%s:tenant:%s:%s", tenantID, module, key)
}

func (c *ConfigurationClient) calculateCacheTTL(source string) time.Duration {
    // System configurations rarely change - cache longer
    if source == "system" {
        return 30 * time.Minute
    }
    // Tenant configurations change occasionally  
    if source == "tenant" {
        return 10 * time.Minute
    }
    // Entity configurations change more frequently
    return 5 * time.Minute
}
```

### **Typed Configuration Wrapper**
```go
// Strongly-typed configuration wrapper
package config

type TypedConfiguration struct {
    client *client.ConfigurationClient
}

// Finance configuration accessor
func (t *TypedConfiguration) Finance(ctx context.Context, tenantID tenant.ID, entityID *entity.ID) *FinanceConfig {
    return &FinanceConfig{
        client:   t.client,
        tenantID: tenantID,
        entityID: entityID,
    }
}

type FinanceConfig struct {
    client   *client.ConfigurationClient
    tenantID tenant.ID
    entityID *entity.ID
}

func (f *FinanceConfig) DefaultCurrency(ctx context.Context) (string, error) {
    value, err := f.client.GetConfiguration(ctx, f.tenantID, f.entityID, "finance", "default_currency")
    if err != nil {
        return "", err
    }
    return value.(string), nil
}

func (f *FinanceConfig) InvoicePrefix(ctx context.Context) (string, error) {
    value, err := f.client.GetConfiguration(ctx, f.tenantID, f.entityID, "finance", "invoice_prefix")
    if err != nil {
        return "", err
    }
    return value.(string), nil
}

func (f *FinanceConfig) AutoApprovalLimit(ctx context.Context) (decimal.Decimal, error) {
    value, err := f.client.GetConfiguration(ctx, f.tenantID, f.entityID, "finance", "auto_approval_limit")
    if err != nil {
        return decimal.Zero, err
    }
    return decimal.NewFromFloat(value.(float64)), nil
}

func (f *FinanceConfig) MultiCurrencyEnabled(ctx context.Context) (bool, error) {
    value, err := f.client.GetConfiguration(ctx, f.tenantID, f.entityID, "finance", "multi_currency_enabled")
    if err != nil {
        return false, err
    }
    return value.(bool), nil
}

// Batch configuration loading for efficiency
func (f *FinanceConfig) LoadAll(ctx context.Context) (*AllFinanceConfig, error) {
    req := client.BatchConfigurationRequest{
        TenantID: f.tenantID,
        EntityID: f.entityID,
        ConfigurationKeys: []service.ConfigurationKey{
            {Module: "finance", Key: "default_currency"},
            {Module: "finance", Key: "invoice_prefix"},
            {Module: "finance", Key: "auto_approval_limit"},
            {Module: "finance", Key: "multi_currency_enabled"},
            {Module: "finance", Key: "fiscal_year_start"},
            {Module: "finance", Key: "payment_terms_default"},
        },
    }
    
    results, err := f.client.GetConfigurations(ctx, req)
    if err != nil {
        return nil, err
    }
    
    return &AllFinanceConfig{
        DefaultCurrency:       results["finance.default_currency"].(string),
        InvoicePrefix:        results["finance.invoice_prefix"].(string),
        AutoApprovalLimit:    decimal.NewFromFloat(results["finance.auto_approval_limit"].(float64)),
        MultiCurrencyEnabled: results["finance.multi_currency_enabled"].(bool),
        FiscalYearStart:      int(results["finance.fiscal_year_start"].(float64)),
        PaymentTermsDefault:  results["finance.payment_terms_default"].(string),
    }, nil
}
```

---

## Template System Integration

### **Template Application Workflow**
```go
// Template application with business validation
package template

import (
    "context"
    "fmt"
    "time"
)

type TemplateApplicator struct {
    settingsService    service.SettingsService
    validationService  ValidationService
    workflowService    WorkflowService
    auditService      AuditService
}

// Apply template with  validation and rollback support
func (t *TemplateApplicator) ApplyTemplateWithValidation(ctx context.Context, req TemplateApplicationRequest) (*TemplateApplicationResult, error) {
    // Start audit trail
    auditID := t.auditService.StartTemplateApplication(ctx, req)
    
    // Phase 1: Pre-application validation
    validationResult, err := t.validateTemplateApplication(ctx, req)
    if err != nil {
        t.auditService.RecordError(ctx, auditID, "validation_failed", err)
        return nil, err
    }
    
    if !validationResult.IsValid {
        return &TemplateApplicationResult{
            Status:           "validation_failed",
            ValidationErrors: validationResult.Errors,
        }, nil
    }
    
    // Phase 2: Create rollback point
    rollbackPoint, err := t.createRollbackPoint(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to create rollback point: %w", err)
    }
    
    // Phase 3: Apply template
    application, err := t.settingsService.ApplyTemplate(ctx, service.TemplateApplicationRequest{
        TemplateID: req.TemplateID,
        TargetType: req.TargetType,
        TargetID:   req.TargetID,
        Options: service.ApplicationOptions{
            DryRun:             false,
            PreserveExisting:   req.PreserveExisting,
            ConflictResolution: req.ConflictResolution,
        },
    })
    if err != nil {
        // Rollback on failure
        rollbackErr := t.rollback(ctx, rollbackPoint)
        if rollbackErr != nil {
            t.auditService.RecordError(ctx, auditID, "rollback_failed", rollbackErr)
        }
        return nil, err
    }
    
    // Phase 4: Post-application validation
    postValidation, err := t.validatePostApplication(ctx, req, application)
    if err != nil {
        // Rollback on post-validation failure
        rollbackErr := t.rollback(ctx, rollbackPoint)
        if rollbackErr != nil {
            t.auditService.RecordError(ctx, auditID, "rollback_failed", rollbackErr)
        }
        return nil, err
    }
    
    // Phase 5: Trigger dependent workflows
    if err := t.triggerDependentWorkflows(ctx, req, application); err != nil {
        // Log but don't fail the application
        t.auditService.RecordWarning(ctx, auditID, "dependent_workflows_failed", err)
    }
    
    // Complete audit trail
    t.auditService.CompleteTemplateApplication(ctx, auditID, application)
    
    return &TemplateApplicationResult{
        Status:              "completed",
        ApplicationResult:   application,
        ValidationResult:    postValidation,
        AuditID:            auditID,
        RollbackSupported:  true,
        RollbackPointID:    rollbackPoint.ID,
    }, nil
}

// Create industry-specific template application workflows
func (t *TemplateApplicator) ApplyIndustryTemplate(ctx context.Context, industry IndustryType, tenantID tenant.ID) error {
    templates := t.getIndustryTemplates(industry)
    
    // Apply templates in dependency order
    for _, templateConfig := range templates {
        req := TemplateApplicationRequest{
            TemplateID:         templateConfig.ID,
            TargetType:         "tenant",
            TargetID:           string(tenantID),
            PreserveExisting:   true,
            ConflictResolution: "merge",
            Industry:           industry,
        }
        
        result, err := t.ApplyTemplateWithValidation(ctx, req)
        if err != nil {
            return fmt.Errorf("failed to apply template %s: %w", templateConfig.Name, err)
        }
        
        if result.Status != "completed" {
            return fmt.Errorf("template %s application failed: %v", templateConfig.Name, result.ValidationErrors)
        }
        
        // Wait for dependent configurations to propagate
        time.Sleep(100 * time.Millisecond)
    }
    
    return nil
}

func (t *TemplateApplicator) getIndustryTemplates(industry IndustryType) []TemplateConfiguration {
    templateMap := map[IndustryType][]TemplateConfiguration{
        IndustryManufacturing: {
            {ID: "base-erp-foundation", Name: "Base ERP", Priority: 1},
            {ID: "manufacturing-core", Name: "Manufacturing Core", Priority: 2},
            {ID: "inventory-manufacturing", Name: "Manufacturing Inventory", Priority: 3},
            {ID: "finance-manufacturing", Name: "Manufacturing Finance", Priority: 4},
        },
        IndustryServices: {
            {ID: "base-erp-foundation", Name: "Base ERP", Priority: 1},
            {ID: "services-core", Name: "Services Core", Priority: 2},
            {ID: "project-management", Name: "Project Management", Priority: 3},
            {ID: "finance-services", Name: "Services Finance", Priority: 4},
        },
        IndustryRetail: {
            {ID: "base-erp-foundation", Name: "Base ERP", Priority: 1},
            {ID: "retail-core", Name: "Retail Core", Priority: 2},
            {ID: "pos-integration", Name: "POS Integration", Priority: 3},
            {ID: "inventory-retail", Name: "Retail Inventory", Priority: 4},
            {ID: "finance-retail", Name: "Retail Finance", Priority: 5},
        },
    }
    
    return templateMap[industry]
}
```

### **Custom Template Creation**
```go
// Custom template builder for tenant-specific configurations
type CustomTemplateBuilder struct {
    settingsService service.SettingsService
    templateService TemplateService
}

func (c *CustomTemplateBuilder) CreateCustomTemplate(ctx context.Context, req CustomTemplateRequest) (*CustomTemplate, error) {
    // Extract current configurations from source tenant/entity
    currentConfigs, err := c.extractConfigurations(ctx, req.SourceTenantID, req.SourceEntityID)
    if err != nil {
        return nil, err
    }
    
    // Filter configurations based on template scope
    filteredConfigs := c.filterConfigurations(currentConfigs, req.IncludeModules, req.ExcludeKeys)
    
    // Create template definition
    template := &domain.Template{
        Name:         req.Name,
        Category:     "custom",
        Description:  req.Description,
        Version:      "1.0",
        IsActive:     true,
        CreatedBy:    req.CreatedBy,
        Configurations: c.buildTemplateConfigurations(filteredConfigs),
        ConflictResolution: req.ConflictResolution,
    }
    
    // Validate template
    if err := c.validateTemplate(ctx, template); err != nil {
        return nil, err
    }
    
    // Save template
    savedTemplate, err := c.templateService.CreateTemplate(ctx, template)
    if err != nil {
        return nil, err
    }
    
    return &CustomTemplate{
        Template:         savedTemplate,
        ConfigurationCount: len(filteredConfigs),
        SourceTenantID:   req.SourceTenantID,
        SourceEntityID:   req.SourceEntityID,
    }, nil
}

// Template versioning and evolution
func (c *CustomTemplateBuilder) CreateTemplateVersion(ctx context.Context, baseTemplateID string, changes []TemplateChange) (*Template, error) {
    baseTemplate, err := c.templateService.GetTemplate(ctx, baseTemplateID)
    if err != nil {
        return nil, err
    }
    
    // Create new version
    newVersion := c.calculateNextVersion(baseTemplate.Version)
    
    newTemplate := &domain.Template{
        Name:                baseTemplate.Name,
        Category:           baseTemplate.Category,
        Description:        baseTemplate.Description,
        Version:            newVersion,
        IsActive:           true,
        Configurations:     c.applyTemplateChanges(baseTemplate.Configurations, changes),
        ApplicableTenantTypes: baseTemplate.ApplicableTenantTypes,
        RequiredFeatureFlags:  baseTemplate.RequiredFeatureFlags,
        ConflictResolution:    baseTemplate.ConflictResolution,
    }
    
    // Validate new template version
    if err := c.validateTemplateEvolution(ctx, baseTemplate, newTemplate); err != nil {
        return nil, err
    }
    
    return c.templateService.CreateTemplate(ctx, newTemplate)
}
```

---

## Event-Driven Integration

### **Configuration Change Event Bus**
```go
// Event-driven configuration propagation
package events

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/go-redis/redis/v8"
)

type ConfigurationEventBus struct {
    redisClient *redis.Client
    subscribers map[string][]ConfigurationEventHandler
}

type ConfigurationEvent interface {
    EventType() string
    TenantID() string
    Timestamp() time.Time
    GetPayload() interface{}
}

type ConfigurationEventHandler interface {
    Handle(ctx context.Context, event ConfigurationEvent) error
    GetSubscriptionPattern() string
}

// Configuration change event
type ConfigurationChangedEvent struct {
    ID         string          `json:"id"`
    Type       string          `json:"type"`
    Time       time.Time       `json:"timestamp"`
    TenantID   string          `json:"tenant_id"`
    EntityID   *string         `json:"entity_id,omitempty"`
    Module     string          `json:"module"`
    ConfigKey  string          `json:"config_key"`
    OldValue   interface{}     `json:"old_value"`
    NewValue   interface{}     `json:"new_value"`
    Source     string          `json:"source"`
    ChangedBy  string          `json:"changed_by"`
    Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

func (e ConfigurationChangedEvent) EventType() string { return e.Type }
func (e ConfigurationChangedEvent) TenantID() string { return e.TenantID }
func (e ConfigurationChangedEvent) Timestamp() time.Time { return e.Time }
func (e ConfigurationChangedEvent) GetPayload() interface{} { return e }

// Template application event
type TemplateAppliedEvent struct {
    ID              string    `json:"id"`
    Type            string    `json:"type"`
    Time            time.Time `json:"timestamp"`
    TenantID        string    `json:"tenant_id"`
    TemplateID      string    `json:"template_id"`
    TemplateName    string    `json:"template_name"`
    TargetType      string    `json:"target_type"`
    TargetID        string    `json:"target_id"`
    AppliedConfigs  int       `json:"applied_configs"`
    ConflictCount   int       `json:"conflict_count"`
    AppliedBy       string    `json:"applied_by"`
}

func (e TemplateAppliedEvent) EventType() string { return e.Type }
func (e TemplateAppliedEvent) TenantID() string { return e.TenantID }
func (e TemplateAppliedEvent) Timestamp() time.Time { return e.Time }
func (e TemplateAppliedEvent) GetPayload() interface{} { return e }

// Publish configuration change event
func (c *ConfigurationEventBus) PublishConfigurationChange(ctx context.Context, change ConfigurationChangedEvent) error {
    streamKey := "settings:configuration:changed"
    
    data, err := json.Marshal(change)
    if err != nil {
        return err
    }
    
    return c.redisClient.XAdd(ctx, &redis.XAddArgs{
        Stream: streamKey,
        Values: map[string]interface{}{
            "tenant_id":   change.TenantID,
            "event_type":  change.EventType(),
            "module":      change.Module,
            "config_key":  change.ConfigKey,
            "data":        string(data),
            "timestamp":   change.Timestamp().Unix(),
        },
    }).Err()
}

// Subscribe to configuration changes with pattern matching
func (c *ConfigurationEventBus) Subscribe(ctx context.Context, pattern string, handler ConfigurationEventHandler) error {
    consumerGroup := fmt.Sprintf("settings-consumer-%s", handler.GetSubscriptionPattern())
    streamKey := "settings:configuration:changed"
    
    // Create consumer group if not exists
    c.redisClient.XGroupCreateMkStream(ctx, streamKey, consumerGroup, "0")
    
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            default:
                streams, err := c.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
                    Group:    consumerGroup,
                    Consumer: "settings-handler",
                    Streams:  []string{streamKey, ">"},
                    Count:    10,
                    Block:    time.Second,
                }).Result()
                
                if err != nil {
                    continue
                }
                
                for _, stream := range streams {
                    for _, message := range stream.Messages {
                        err := c.processMessage(ctx, handler, message)
                        if err != nil {
                            // Handle error, possibly move to dead letter queue
                            continue
                        }
                        
                        // Acknowledge message
                        c.redisClient.XAck(ctx, streamKey, consumerGroup, message.ID)
                    }
                }
            }
        }
    }()
    
    return nil
}

func (c *ConfigurationEventBus) processMessage(ctx context.Context, handler ConfigurationEventHandler, message redis.XMessage) error {
    eventData, ok := message.Values["data"].(string)
    if !ok {
        return fmt.Errorf("invalid event data format")
    }
    
    var event ConfigurationChangedEvent
    if err := json.Unmarshal([]byte(eventData), &event); err != nil {
        return err
    }
    
    // Check if handler pattern matches
    if c.matchesPattern(handler.GetSubscriptionPattern(), event.Module, event.ConfigKey) {
        return handler.Handle(ctx, event)
    }
    
    return nil
}
```

### **Real-time Configuration Synchronization**
```go
// Real-time configuration sync across services
type ConfigurationSynchronizer struct {
    eventBus        *ConfigurationEventBus
    cacheManager    CacheManager
    notificationSvc NotificationService
}

// Synchronize configuration changes across all services
func (c *ConfigurationSynchronizer) SynchronizeConfigurationChange(ctx context.Context, change ConfigurationChangedEvent) error {
    // 1. Invalidate distributed caches
    err := c.invalidateDistributedCache(ctx, change)
    if err != nil {
        log.Printf("Failed to invalidate cache for config change: %v", err)
    }
    
    // 2. Notify affected services
    affectedServices := c.identifyAffectedServices(change.Module, change.ConfigKey)
    for _, service := range affectedServices {
        notification := ServiceNotification{
            ServiceName:     service,
            NotificationType: "configuration_changed",
            Payload:         change,
            DeliveryMethod:  "webhook",
        }
        
        go c.notificationSvc.NotifyService(ctx, notification)
    }
    
    // 3. Update service discovery metadata if needed
    if c.isServiceDiscoveryConfig(change.Module, change.ConfigKey) {
        err := c.updateServiceDiscovery(ctx, change)
        if err != nil {
            log.Printf("Failed to update service discovery: %v", err)
        }
    }
    
    // 4. Trigger dependent configuration recalculations
    dependentConfigs := c.identifyDependentConfigurations(change)
    for _, depConfig := range dependentConfigs {
        go c.recalculateDependentConfiguration(ctx, depConfig, change)
    }
    
    return nil
}

func (c *ConfigurationSynchronizer) invalidateDistributedCache(ctx context.Context, change ConfigurationChangedEvent) error {
    patterns := []string{
        fmt.Sprintf("config:*:%s:%s:%s", change.TenantID, change.Module, change.ConfigKey),
        fmt.Sprintf("config:*:%s:*:%s:%s", change.TenantID, change.Module, change.ConfigKey),
        fmt.Sprintf("batch:config:%s:*", change.TenantID),
    }
    
    for _, pattern := range patterns {
        if err := c.cacheManager.InvalidatePattern(ctx, pattern); err != nil {
            return err
        }
    }
    
    return nil
}

func (c *ConfigurationSynchronizer) identifyAffectedServices(module, configKey string) []string {
    // Service dependency mapping
    serviceDependencies := map[string]map[string][]string{
        "finance": {
            "default_currency":       {"finance-service", "reporting-service", "analytics-service"},
            "invoice_prefix":         {"finance-service", "document-service"},
            "auto_approval_limit":    {"finance-service", "workflow-service"},
            "multi_currency_enabled": {"finance-service", "exchange-service", "reporting-service"},
        },
        "hr": {
            "default_pay_frequency":  {"payroll-service", "hr-service"},
            "overtime_threshold":     {"payroll-service", "time-tracking-service"},
            "benefits_waiting_period": {"hr-service", "benefits-service"},
        },
        "inventory": {
            "default_valuation_method": {"inventory-service", "accounting-service", "reporting-service"},
            "auto_reorder_enabled":     {"inventory-service", "purchasing-service"},
            "lot_tracking_required":    {"inventory-service", "warehouse-service", "quality-service"},
        },
    }
    
    if moduleConfigs, exists := serviceDependencies[module]; exists {
        if services, exists := moduleConfigs[configKey]; exists {
            return services
        }
    }
    
    return []string{}
}
```

---

This  integration guide demonstrates enterprise-grade configuration management integration with proper event-driven architecture, caching strategies, and real-time synchronization. The remaining sections would continue with similar detail for cache integration patterns, feature flag integration, testing strategies, and troubleshooting procedures.

**Document Status**: Core Integration Patterns Complete
**Next Sections**: Cache Integration, Testing, Troubleshooting