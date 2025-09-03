package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Configuration is an aggregate root managing setting inheritance and validation
type Configuration struct {
	// Identity
	ID       ConfigurationID
	EntityID *uuid.UUID // Nil for tenant-level configs
	// Configuration Scope
	Module    ModuleName
	ConfigKey ConfigKey

	// Value and Metadata
	Value         ConfigValue
	Source        ConfigSource  // system, tenant, entity, template
	IsInherited   bool          // true if inherited from parent level
	OverridesFrom *ConfigSource // tracks what this overrides

	// Validation and Rules
	DataType        DataType
	ValidationRules ValidationRules

	// Permissions and Features
	RequiredPermissions []Permission
	RequiredFeatureFlag *FeatureFlag

	// Metadata
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   int64 // Optimistic locking
}

// ConfigurationID uniquely identifies a configuration
type ConfigurationID string

func (id ConfigurationID) String() string {
	return string(id)
}

// NewConfigurationID creates a new configuration ID
func NewConfigurationID() ConfigurationID {
	return ConfigurationID(uuid.New().String())
}

// ModuleName represents an official ERP module.
type ModuleName string

const (
	ModuleFinance       ModuleName = "FINANCE"
	ModuleHR            ModuleName = "HUMAN_RESOURCES" // Often aliasable to ModuleHCM
	ModuleHCM           ModuleName = "HUMAN_CAPITAL_MGMT"
	ModuleInventory     ModuleName = "INVENTORY"
	ModuleWarehouse     ModuleName = "WAREHOUSE_MGMT"
	ModuleSales         ModuleName = "SALES"
	ModulePurchase      ModuleName = "PURCHASING"
	ModuleCRM           ModuleName = "CUSTOMER_RELATION_MGMT"
	ModuleProject       ModuleName = "PROJECT_MGMT"
	ModuleSCM           ModuleName = "SUPPLY_CHAIN_MGMT"
	ModuleManufacturing ModuleName = "MANUFACTURING"
	ModuleEcommerce     ModuleName = "ECOMMERCE"
	ModuleBusinessIntel ModuleName = "BUSINESS_INTELLIGENCE"
	ModuleAssetMgmt     ModuleName = "ASSET_MGMT"
	ModulePointOfSale   ModuleName = "POINT_OF_SALE" // The POS module we discussed
)

// Valid returns true if the module name is valid
func (mn ModuleName) Valid() bool {
	switch mn {
	case ModuleFinance, ModuleHR, ModuleInventory, ModuleSales, ModulePurchase, ModuleCRM, ModuleProject:
		return true
	default:
		return false
	}
}

func (mn ModuleName) String() string {
	return string(mn)
}

// ConfigSource represents where a configuration value originated
type ConfigSource string

const (
	ConfigSourceSystem   ConfigSource = "SYSTEM"
	ConfigSourceTenant   ConfigSource = "TENANT"
	ConfigSourceEntity   ConfigSource = "ENTITY"
	ConfigSourceTemplate ConfigSource = "TEMPLATE"
)

// Valid returns true if the config source is valid
func (cs ConfigSource) Valid() bool {
	switch cs {
	case ConfigSourceSystem, ConfigSourceTenant, ConfigSourceEntity, ConfigSourceTemplate:
		return true
	default:
		return false
	}
}

func (cs ConfigSource) String() string {
	return string(cs)
}

// Priority returns the inheritance priority (higher number = higher priority)
func (cs ConfigSource) Priority() int {
	switch cs {
	case ConfigSourceSystem:
		return 0
	case ConfigSourceTenant:
		return 1
	case ConfigSourceEntity:
		return 2
	case ConfigSourceTemplate:
		return 1 // Same as tenant, but applied differently
	default:
		return -1
	}
}

// DataType represents the data type of a configuration value
type DataType string

const (
	DataTypeString  DataType = "STRING"
	DataTypeInteger DataType = "INTEGER"
	DataTypeBoolean DataType = "BOOLEAN"
	DataTypeDecimal DataType = "DECIMAL"
	DataTypeJSON    DataType = "JSON"
)

// Valid returns true if the data type is valid
func (dt DataType) Valid() bool {
	switch dt {
	case DataTypeString, DataTypeInteger, DataTypeBoolean, DataTypeDecimal, DataTypeJSON:
		return true
	default:
		return false
	}
}

func (dt DataType) String() string {
	return string(dt)
}

// Permission represents a required permission
type Permission string

// FeatureFlag represents a required feature flag
type FeatureFlag string

// NewConfiguration creates a new configuration with validation
func NewConfiguration(
	entityID *uuid.UUID,
	module ModuleName,
	key ConfigKey,
	value ConfigValue,
	source ConfigSource,
) (*Configuration, error) {
	if !module.Valid() {
		return nil, ErrInvalidModule
	}

	if !key.IsValid() {
		return nil, ErrInvalidConfigKey
	}

	if !source.Valid() {
		return nil, ErrInvalidConfigSource
	}

	config := &Configuration{
		ID:        NewConfigurationID(),
		EntityID:  entityID,
		Module:    module,
		ConfigKey: key,
		Value:     value,
		Source:    source,
		DataType:  value.DataType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate performs domain validation on the configuration
func (c *Configuration) Validate() error {
	if c.Module == "" {
		return ErrModuleRequired
	}

	if !c.Module.Valid() {
		return ErrInvalidModule
	}

	if c.ConfigKey == "" {
		return ErrConfigKeyRequired
	}

	if !c.ConfigKey.IsValid() {
		return ErrInvalidConfigKey
	}

	if !c.Value.IsValidForType(c.DataType) {
		return ErrInvalidValueForType
	}

	if !c.Source.Valid() {
		return ErrInvalidConfigSource
	}

	// Inherited values must have a parent source
	if c.IsInherited && c.Source != ConfigSourceSystem {
		if c.OverridesFrom == nil {
			return ErrInheritedValueMustHaveParent
		}
	}

	// Entity-level configs must have an entity ID
	if c.Source == ConfigSourceEntity && c.EntityID == nil {
		return ErrEntityConfigRequiresEntityID
	}

	return nil
}

// CanBeOverridden returns true if this configuration can be overridden
func (c *Configuration) CanBeOverridden() bool {
	// System configs cannot be overridden if they're marked as non-overridable
	// This would be determined by the config definition
	return true // TODO: Check config definition
}

// Override updates the configuration with a new value from a different source
func (c *Configuration) Override(newValue ConfigValue, source ConfigSource, userID uuid.UUID) error {
	if !c.CanBeOverridden() {
		return ErrConfigurationNotOverridable
	}

	if !source.Valid() {
		return ErrInvalidConfigSource
	}

	// Cannot override with a lower priority source
	if source.Priority() <= c.Source.Priority() {
		return ErrCannotOverrideWithLowerPriority
	}

	oldSource := c.Source
	c.Value = newValue
	c.Source = source
	c.IsInherited = false
	c.OverridesFrom = &oldSource
	c.UpdatedAt = time.Now()
	c.Version++

	return c.Validate()
}

// ResetToInherited resets the configuration to its inherited value
func (c *Configuration) ResetToInherited(parentValue ConfigValue, parentSource ConfigSource, userID uuid.UUID) error {
	if c.Source == ConfigSourceSystem {
		return ErrCannotResetSystemConfiguration
	}

	if !parentSource.Valid() {
		return ErrInvalidConfigSource
	}

	c.Value = parentValue
	c.Source = parentSource
	c.IsInherited = true
	c.OverridesFrom = nil
	c.UpdatedAt = time.Now()
	c.Version++

	return nil
}

// UpdateValue updates the configuration value while maintaining the same source
func (c *Configuration) UpdateValue(newValue ConfigValue, userID uuid.UUID) error {
	if !newValue.IsValidForType(c.DataType) {
		return ErrInvalidValueForType
	}

	c.Value = newValue
	c.UpdatedAt = time.Now()
	c.Version++

	return nil
}

// GetInheritanceLevel returns the inheritance level of this configuration
func (c *Configuration) GetInheritanceLevel() int {
	return c.Source.Priority()
}

// IsSystemDefault returns true if this is a system default configuration
func (c *Configuration) IsSystemDefault() bool {
	return c.Source == ConfigSourceSystem
}

// IsTenantLevel returns true if this is a tenant-level configuration
func (c *Configuration) IsTenantLevel() bool {
	return c.Source == ConfigSourceTenant
}

// IsEntityLevel returns true if this is an entity-level configuration
func (c *Configuration) IsEntityLevel() bool {
	return c.Source == ConfigSourceEntity
}

// GetFullKey returns the full configuration key (module.key)
func (c *Configuration) GetFullKey() string {
	return fmt.Sprintf("%s.%s", c.Module, c.ConfigKey)
}

// Clone creates a copy of the configuration
func (c *Configuration) Clone() *Configuration {
	clone := &Configuration{
		ID:                  c.ID,
		Module:              c.Module,
		ConfigKey:           c.ConfigKey,
		Value:               c.Value,
		Source:              c.Source,
		IsInherited:         c.IsInherited,
		DataType:            c.DataType,
		ValidationRules:     c.ValidationRules,
		RequiredPermissions: make([]Permission, len(c.RequiredPermissions)),
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
		Version:             c.Version,
	}

	if c.EntityID != nil {
		entityID := *c.EntityID
		clone.EntityID = &entityID
	}

	if c.OverridesFrom != nil {
		source := *c.OverridesFrom
		clone.OverridesFrom = &source
	}

	if c.RequiredFeatureFlag != nil {
		flag := *c.RequiredFeatureFlag
		clone.RequiredFeatureFlag = &flag
	}

	copy(clone.RequiredPermissions, c.RequiredPermissions)

	return clone
}

// Equals compares two configurations for equality
func (c *Configuration) Equals(other *Configuration) bool {
	if other == nil {
		return false
	}

	return c.ID == other.ID &&
		c.Module == other.Module &&
		c.ConfigKey == other.ConfigKey &&
		c.Value.Equals(other.Value) &&
		c.Source == other.Source &&
		c.Version == other.Version
}

// String returns a string representation of the configuration
func (c *Configuration) String() string {
	entityStr := "nil"
	if c.EntityID != nil {
		entityStr = c.EntityID.String()
	}

	return fmt.Sprintf("Configuration{ID: %s, Entity: %s, Key: %s.%s, Value: %v, Source: %s}",
		c.ID, entityStr, c.Module, c.ConfigKey, c.Value.Raw, c.Source)
}
