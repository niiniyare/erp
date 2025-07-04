# Module Configuration

## ⚙️ Overview

The ERP system uses a sophisticated module configuration framework that allows fine-grained control over feature availability, business rules, and system behavior. This document outlines how to configure modules, manage feature toggles, customize business processes, and adapt the system for specific industry requirements.

## 🏗️ Configuration Architecture

### Configuration Hierarchy

```yaml
configuration_hierarchy:
  system_level:
    description: "Global system configurations"
    scope: "all_tenants"
    examples:
      - feature_availability
      - system_limits
      - security_policies
      - integration_endpoints
  
  tenant_level:
    description: "Tenant-specific configurations"
    scope: "single_tenant"
    examples:
      - enabled_modules
      - business_rules
      - workflow_definitions
      - custom_fields
  
  organization_level:
    description: "Organization-specific settings"
    scope: "single_organization"
    examples:
      - department_settings
      - approval_hierarchies
      - location_preferences
      - operational_parameters
  
  user_level:
    description: "Individual user preferences"
    scope: "single_user"
    examples:
      - ui_preferences
      - notification_settings
      - dashboard_layouts
      - default_values
```

### Configuration Storage

```sql
-- System-wide configuration
CREATE TABLE system_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(255) UNIQUE NOT NULL,
    config_value JSONB NOT NULL,
    config_type VARCHAR(50) NOT NULL, -- feature_toggle, system_setting, integration_config
    description TEXT,
    is_encrypted BOOLEAN DEFAULT false,
    requires_restart BOOLEAN DEFAULT false,
    validation_schema JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    updated_by UUID REFERENCES users(id)
);

-- Tenant-specific configuration
CREATE TABLE tenant_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    config_key VARCHAR(255) NOT NULL,
    config_value JSONB NOT NULL,
    config_type VARCHAR(50) NOT NULL,
    module_name VARCHAR(100),
    description TEXT,
    overrides_system_config BOOLEAN DEFAULT false,
    effective_date DATE DEFAULT CURRENT_DATE,
    expiry_date DATE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    updated_by UUID REFERENCES users(id),
    
    UNIQUE(tenant_id, config_key)
);

-- Organization-level configuration
CREATE TABLE organization_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    config_key VARCHAR(255) NOT NULL,
    config_value JSONB NOT NULL,
    config_scope VARCHAR(50) DEFAULT 'organization', -- organization, department, team
    inherits_from_parent BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(organization_id, config_key)
);

-- User preferences
CREATE TABLE user_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    preference_category VARCHAR(50) NOT NULL, -- ui, notifications, defaults, etc.
    preferences JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(user_id, preference_category)
);
```

## 🎛️ Feature Toggle System

### Feature Toggle Framework

```typescript
interface FeatureToggle {
  key: string;
  name: string;
  description: string;
  enabled: boolean;
  strategy: FeatureStrategy;
  conditions?: FeatureCondition[];
  rollout_percentage?: number;
  target_audience?: TargetAudience;
  dependencies?: string[]; // Other feature keys this depends on
  conflicts?: string[]; // Features that conflict with this one
}

interface FeatureStrategy {
  type: 'simple' | 'gradual_rollout' | 'targeted' | 'a_b_test' | 'kill_switch';
  parameters?: Record<string, any>;
}

interface FeatureCondition {
  type: 'tenant_id' | 'user_role' | 'subscription_plan' | 'date_range' | 'custom';
  operator: 'equals' | 'in' | 'not_in' | 'greater_than' | 'less_than';
  value: any;
}

interface TargetAudience {
  tenant_ids?: string[];
  subscription_plans?: string[];
  user_roles?: string[];
  geographic_regions?: string[];
  custom_attributes?: Record<string, any>;
}

// Feature Toggle Service
class FeatureToggleService {
  private cache: Map<string, FeatureToggle> = new Map();
  private refreshInterval: number = 60000; // 1 minute
  
  async isFeatureEnabled(
    featureKey: string,
    context: FeatureContext
  ): Promise<boolean> {
    const feature = await this.getFeature(featureKey);
    if (!feature) {
      return false;
    }
    
    // Check dependencies
    if (feature.dependencies) {
      for (const dependency of feature.dependencies) {
        if (!(await this.isFeatureEnabled(dependency, context))) {
          return false;
        }
      }
    }
    
    // Check conflicts
    if (feature.conflicts) {
      for (const conflict of feature.conflicts) {
        if (await this.isFeatureEnabled(conflict, context)) {
          return false;
        }
      }
    }
    
    return this.evaluateFeatureStrategy(feature, context);
  }
  
  private async evaluateFeatureStrategy(
    feature: FeatureToggle,
    context: FeatureContext
  ): Promise<boolean> {
    if (!feature.enabled) {
      return false;
    }
    
    switch (feature.strategy.type) {
      case 'simple':
        return this.evaluateConditions(feature.conditions || [], context);
      
      case 'gradual_rollout':
        const rolloutPercentage = feature.rollout_percentage || 0;
        const hash = this.generateConsistentHash(feature.key, context.user_id);
        return (hash % 100) < rolloutPercentage;
      
      case 'targeted':
        return this.evaluateTargetAudience(feature.target_audience, context);
      
      case 'a_b_test':
        return this.evaluateABTest(feature, context);
      
      case 'kill_switch':
        return false; // Emergency disable
      
      default:
        return false;
    }
  }
  
  private evaluateConditions(
    conditions: FeatureCondition[],
    context: FeatureContext
  ): boolean {
    return conditions.every(condition => {
      const contextValue = this.getContextValue(condition.type, context);
      return this.evaluateCondition(condition, contextValue);
    });
  }
}

// Usage in application
const featureToggleService = new FeatureToggleService();

// In API endpoint
app.get('/api/v1/advanced-analytics', async (req, res) => {
  const context = {
    tenant_id: req.tenant.id,
    user_id: req.user.id,
    subscription_plan: req.tenant.subscription_plan,
    user_roles: req.user.roles
  };
  
  const hasAdvancedAnalytics = await featureToggleService.isFeatureEnabled(
    'advanced_analytics',
    context
  );
  
  if (!hasAdvancedAnalytics) {
    return res.status(403).json({
      error: 'Feature not available',
      feature: 'advanced_analytics'
    });
  }
  
  // Proceed with advanced analytics logic
});
```

### Core Module Feature Toggles

```yaml
core_features:
  financial_management:
    multi_currency_support:
      key: "finance.multi_currency"
      default: false
      plans: ["premium", "enterprise"]
      description: "Enable multi-currency transactions and reporting"
    
    advanced_reporting:
      key: "finance.advanced_reporting"
      default: false
      plans: ["premium", "enterprise"]
      description: "Advanced financial reporting and analytics"
    
    automated_bank_reconciliation:
      key: "finance.auto_bank_reconciliation"
      default: false
      plans: ["enterprise"]
      description: "Automated bank statement reconciliation"
      dependencies: ["finance.bank_integration"]
  
  inventory_management:
    serial_tracking:
      key: "inventory.serial_tracking"
      default: false
      plans: ["premium", "enterprise"]
      description: "Track individual serial numbers"
    
    batch_tracking:
      key: "inventory.batch_tracking"
      default: false
      plans: ["premium", "enterprise"]
      description: "Track inventory batches and lots"
    
    multi_warehouse:
      key: "inventory.multi_warehouse"
      default: false
      plans: ["premium", "enterprise"]
      description: "Support for multiple warehouses"
    
    cycle_counting:
      key: "inventory.cycle_counting"
      default: false
      plans: ["premium", "enterprise"]
      description: "Automated cycle counting processes"
      dependencies: ["inventory.multi_warehouse"]
  
  human_resources:
    payroll_processing:
      key: "hr.payroll_processing"
      default: false
      plans: ["premium", "enterprise"]
      description: "Full payroll processing capabilities"
    
    performance_management:
      key: "hr.performance_management"
      default: false
      plans: ["premium", "enterprise"]
      description: "Performance reviews and goal tracking"
    
    recruitment_module:
      key: "hr.recruitment"
      default: false
      plans: ["enterprise"]
      description: "Recruitment and applicant tracking"
```

## 🏭 Industry-Specific Configurations

### Airline Industry Configuration

```yaml
airline_configuration:
  module_enablement:
    core_modules:
      - financial_management
      - inventory_management
      - human_resources
      - project_management
    
    industry_modules:
      - airline_operations
      - flight_scheduling
      - passenger_management
      - crew_management
      - aircraft_maintenance
      - revenue_management
  
  feature_settings:
    reservation_system:
      enable_gds_integration: true
      support_code_share: true
      enable_dynamic_pricing: true
      loyalty_program_integration: true
    
    operational_features:
      real_time_flight_tracking: true
      automated_delay_notifications: true
      crew_scheduling_optimization: true
      fuel_management: true
      maintenance_tracking: true
    
    compliance_features:
      dot_reporting: true
      faa_compliance: true
      icao_standards: true
      passenger_data_protection: true
  
  business_rules:
    booking_rules:
      advance_booking_limit_days: 365
      minimum_booking_time_hours: 2
      overbooking_percentage: 5.0
      upgrade_bidding_enabled: true
    
    operational_rules:
      minimum_connection_time_minutes: 45
      crew_duty_time_limits: "14_hours"
      maintenance_interval_hours: 500
      fuel_reserve_percentage: 15.0
  
  integrations:
    gds_systems:
      - amadeus
      - sabre
      - travelport
    
    airport_systems:
      - departure_control_system
      - baggage_handling_system
      - gate_management_system
    
    regulatory_systems:
      - customs_systems
      - security_systems
      - immigration_systems
```

### Restaurant Industry Configuration

```yaml
restaurant_configuration:
  module_enablement:
    core_modules:
      - financial_management
      - inventory_management
      - human_resources
    
    industry_modules:
      - restaurant_operations
      - menu_management
      - kitchen_operations
      - table_management
      - customer_experience
      - food_cost_management
  
  operational_settings:
    pos_integration:
      enable_kitchen_display: true
      support_mobile_ordering: true
      table_side_payment: true
      split_bill_functionality: true
    
    inventory_settings:
      track_expiry_dates: true
      fifo_rotation_enforcement: true
      recipe_cost_calculation: true
      waste_tracking: true
    
    service_settings:
      reservation_system: true
      wait_list_management: true
      loyalty_program: true
      online_ordering: true
  
  compliance_features:
    food_safety:
      haccp_compliance: true
      temperature_monitoring: true
      allergen_tracking: true
      nutritional_information: true
    
    regulatory_compliance:
      health_department_reporting: true
      labor_law_compliance: true
      tax_reporting: true
      liquor_license_tracking: true
  
  business_rules:
    operational_rules:
      table_turnover_target_minutes: 90
      food_cost_percentage_target: 30.0
      labor_cost_percentage_target: 25.0
      waste_percentage_threshold: 2.0
    
    pricing_rules:
      menu_price_rounding: "nearest_0_25"
      seasonal_price_adjustments: true
      happy_hour_discounts: true
      group_booking_discounts: true
```

### Retail Industry Configuration

```yaml
retail_configuration:
  module_enablement:
    core_modules:
      - financial_management
      - inventory_management
      - human_resources
      - customer_management
    
    industry_modules:
      - retail_operations
      - omnichannel_management
      - customer_analytics
      - merchandising
      - e_commerce_integration
      - loyalty_management
  
  channel_settings:
    online_channels:
      e_commerce_platform: true
      mobile_app: true
      social_commerce: true
      marketplace_integration: true
    
    physical_channels:
      multiple_stores: true
      pop_up_stores: true
      kiosks: true
      click_and_collect: true
    
    inventory_synchronization:
      real_time_sync: true
      channel_specific_allocation: true
      automatic_rebalancing: true
      cross_channel_fulfillment: true
  
  customer_features:
    analytics_capabilities:
      customer_segmentation: true
      purchase_behavior_analysis: true
      lifetime_value_calculation: true
      churn_prediction: true
    
    personalization:
      product_recommendations: true
      personalized_pricing: true
      targeted_promotions: true
      dynamic_content: true
  
  business_rules:
    pricing_rules:
      dynamic_pricing: true
      competitor_price_monitoring: true
      markdown_optimization: true
      promotion_management: true
    
    inventory_rules:
      safety_stock_optimization: true
      automatic_replenishment: true
      seasonal_planning: true
      slow_moving_identification: true
```

## 🔧 Configuration Management System

### Configuration API

```typescript
// Configuration Management API
interface ConfigurationAPI {
  // Get configuration value with fallback hierarchy
  getConfig<T>(key: string, context: ConfigContext): Promise<T>;
  
  // Set configuration value
  setConfig(key: string, value: any, scope: ConfigScope): Promise<void>;
  
  // Validate configuration value
  validateConfig(key: string, value: any): Promise<ValidationResult>;
  
  // Get all configurations for a scope
  getConfigurations(scope: ConfigScope, filters?: ConfigFilter[]): Promise<Configuration[]>;
  
  // Subscribe to configuration changes
  subscribeToChanges(keys: string[], callback: ConfigChangeCallback): void;
}

interface ConfigContext {
  tenant_id: string;
  organization_id?: string;
  user_id?: string;
  environment?: string;
}

interface ConfigScope {
  level: 'system' | 'tenant' | 'organization' | 'user';
  target_id?: string;
}

class ConfigurationService implements ConfigurationAPI {
  private cache: ConfigurationCache;
  private validator: ConfigurationValidator;
  private eventBus: EventBus;
  
  async getConfig<T>(key: string, context: ConfigContext): Promise<T> {
    // Check cache first
    const cacheKey = this.buildCacheKey(key, context);
    let value = await this.cache.get<T>(cacheKey);
    
    if (value !== undefined) {
      return value;
    }
    
    // Resolve configuration hierarchy
    value = await this.resolveConfigurationHierarchy<T>(key, context);
    
    // Cache the resolved value
    await this.cache.set(cacheKey, value, { ttl: 300 }); // 5 minutes
    
    return value;
  }
  
  private async resolveConfigurationHierarchy<T>(
    key: string,
    context: ConfigContext
  ): Promise<T> {
    const resolutionOrder = [
      // User-level (highest priority)
      { level: 'user', id: context.user_id },
      // Organization-level
      { level: 'organization', id: context.organization_id },
      // Tenant-level
      { level: 'tenant', id: context.tenant_id },
      // System-level (lowest priority/default)
      { level: 'system', id: null }
    ];
    
    for (const scope of resolutionOrder) {
      if (!scope.id && scope.level !== 'system') continue;
      
      const value = await this.getConfigFromStorage<T>(key, scope);
      if (value !== undefined) {
        return value;
      }
    }
    
    // No configuration found, return default value if available
    const defaultValue = await this.getDefaultValue<T>(key);
    if (defaultValue !== undefined) {
      return defaultValue;
    }
    
    throw new Error(`Configuration not found: ${key}`);
  }
  
  async setConfig(
    key: string,
    value: any,
    scope: ConfigScope
  ): Promise<void> {
    // Validate the configuration value
    const validation = await this.validator.validate(key, value);
    if (!validation.valid) {
      throw new Error(`Invalid configuration: ${validation.errors.join(', ')}`);
    }
    
    // Check permissions
    await this.checkPermissions(key, scope);
    
    // Get previous value for change tracking
    const previousValue = await this.getConfigFromStorage(key, scope);
    
    // Store the new value
    await this.storeConfig(key, value, scope);
    
    // Invalidate cache
    await this.invalidateCache(key, scope);
    
    // Emit change event
    this.eventBus.emit('config.changed', {
      key,
      previousValue,
      newValue: value,
      scope,
      timestamp: new Date()
    });
    
    // Handle configurations that require restart
    const configMeta = await this.getConfigurationMetadata(key);
    if (configMeta.requires_restart) {
      this.eventBus.emit('config.restart_required', { key, scope });
    }
  }
}

// Configuration validation
interface ConfigurationValidator {
  validate(key: string, value: any): Promise<ValidationResult>;
  registerValidator(key: string, validator: ValueValidator): void;
}

interface ValueValidator {
  validate(value: any): ValidationResult;
}

class TypedConfigurationValidator implements ConfigurationValidator {
  private validators: Map<string, ValueValidator> = new Map();
  private schemas: Map<string, JSONSchema> = new Map();
  
  async validate(key: string, value: any): Promise<ValidationResult> {
    // Custom validator
    const validator = this.validators.get(key);
    if (validator) {
      return validator.validate(value);
    }
    
    // JSON Schema validation
    const schema = this.schemas.get(key);
    if (schema) {
      return this.validateAgainstSchema(value, schema);
    }
    
    // Default validation (basic type checking)
    return this.defaultValidation(key, value);
  }
  
  registerValidator(key: string, validator: ValueValidator): void {
    this.validators.set(key, validator);
  }
  
  registerSchema(key: string, schema: JSONSchema): void {
    this.schemas.set(key, schema);
  }
}

// Example validator registration
const validator = new TypedConfigurationValidator();

// Register schema for feature toggle
validator.registerSchema('features.*', {
  type: 'object',
  properties: {
    enabled: { type: 'boolean' },
    rollout_percentage: { type: 'number', minimum: 0, maximum: 100 },
    target_audience: {
      type: 'object',
      properties: {
        tenant_ids: { type: 'array', items: { type: 'string' } },
        subscription_plans: { type: 'array', items: { type: 'string' } }
      }
    }
  },
  required: ['enabled']
});

// Register custom validator for business rules
validator.registerValidator('business_rules.approval_limits', {
  validate: (value: any): ValidationResult => {
    if (typeof value !== 'object') {
      return { valid: false, errors: ['Approval limits must be an object'] };
    }
    
    const limits = value as Record<string, number>;
    const errors: string[] = [];
    
    // Validate that amounts are positive
    for (const [role, limit] of Object.entries(limits)) {
      if (typeof limit !== 'number' || limit < 0) {
        errors.push(`Invalid limit for role ${role}: must be a positive number`);
      }
    }
    
    // Validate hierarchy (higher roles should have higher limits)
    const roleHierarchy = ['employee', 'supervisor', 'manager', 'director', 'ceo'];
    for (let i = 1; i < roleHierarchy.length; i++) {
      const currentRole = roleHierarchy[i];
      const previousRole = roleHierarchy[i - 1];
      
      if (limits[currentRole] && limits[previousRole] && 
          limits[currentRole] < limits[previousRole]) {
        errors.push(`${currentRole} limit cannot be less than ${previousRole} limit`);
      }
    }
    
    return {
      valid: errors.length === 0,
      errors
    };
  }
});
```

## 📊 Configuration Templates

### Industry-Specific Templates

```yaml
configuration_templates:
  airline_starter:
    name: "Airline Operations - Starter"
    description: "Basic airline operations configuration"
    target_industry: "airline"
    subscription_plan: "professional"
    
    modules:
      enabled:
        - financial_management
        - inventory_management
        - human_resources
        - airline_operations
        - flight_scheduling
        - passenger_management
      
      disabled:
        - revenue_management
        - crew_optimization
        - maintenance_planning
    
    features:
      airline_operations:
        basic_reservations: true
        flight_scheduling: true
        passenger_checkin: true
        baggage_tracking: false
        loyalty_program: false
        revenue_optimization: false
    
    business_rules:
      booking_window_days: 90
      overbooking_percentage: 0
      minimum_connection_minutes: 60
      advance_checkin_hours: 24
  
  restaurant_full:
    name: "Restaurant Management - Full Service"
    description: "Complete restaurant management configuration"
    target_industry: "restaurant"
    subscription_plan: "enterprise"
    
    modules:
      enabled:
        - financial_management
        - inventory_management
        - human_resources
        - restaurant_operations
        - menu_management
        - kitchen_operations
        - table_management
        - customer_experience
        - pos_integration
    
    features:
      restaurant_operations:
        reservations: true
        waitlist: true
        table_management: true
        kitchen_display: true
        mobile_ordering: true
        loyalty_program: true
        inventory_recipe_costing: true
        waste_tracking: true
    
    business_rules:
      table_turnover_minutes: 90
      reservation_hold_minutes: 15
      food_cost_target_percentage: 28
      labor_cost_target_percentage: 30
      inventory_reorder_days: 3
  
  retail_omnichannel:
    name: "Retail - Omnichannel"
    description: "Multi-channel retail operations"
    target_industry: "retail"
    subscription_plan: "enterprise"
    
    modules:
      enabled:
        - financial_management
        - inventory_management
        - human_resources
        - customer_management
        - retail_operations
        - omnichannel_management
        - e_commerce_integration
        - customer_analytics
        - merchandising
        - loyalty_management
    
    features:
      retail_operations:
        multi_store: true
        ecommerce_sync: true
        click_and_collect: true
        ship_from_store: true
        customer_analytics: true
        dynamic_pricing: true
        loyalty_program: true
        pos_integration: true
    
    business_rules:
      inventory_sync_frequency: "real_time"
      price_update_frequency: "daily"
      loyalty_points_ratio: 100 # points per dollar
      return_window_days: 30
      stock_allocation_method: "demand_based"
```

This comprehensive module configuration system ensures that the ERP can be precisely tailored to meet specific industry requirements while maintaining flexibility for customization and future expansion.
