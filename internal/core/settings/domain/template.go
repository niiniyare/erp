package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Template is an aggregate root for configuration templates
type Template struct {
	// Identity
	ID       TemplateID
	Name     string
	Category TemplateCategory // industry, functional, regional

	// Content
	Configurations []TemplateConfiguration
	Dependencies   []TemplateDependency

	// Metadata
	Description string
	Version     string
	IsActive    bool

	// Application Rules
	ApplicableToTenantTypes []TenantType
	RequiredFeatureFlags    []FeatureFlag
	ConflictResolution      ConflictStrategy

	// Audit
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
}

// TemplateID uniquely identifies a configuration template
type TemplateID string

func (id TemplateID) String() string {
	return string(id)
}

// NewTemplateID creates a new template ID
func NewTemplateID() TemplateID {
	return TemplateID(uuid.New().String())
}

// TemplateCategory represents the category of a template
type TemplateCategory string

const (
	TemplateCategoryIndustry   TemplateCategory = "INDUSTRY"
	TemplateCategoryFunctional TemplateCategory = "FUNCTIONAL"
	TemplateCategoryRegional   TemplateCategory = "REGIONAL"
)

// Valid returns true if the template category is valid
func (tc TemplateCategory) Valid() bool {
	switch tc {
	case TemplateCategoryIndustry, TemplateCategoryFunctional, TemplateCategoryRegional:
		return true
	default:
		return false
	}
}

func (tc TemplateCategory) String() string {
	return string(tc)
}

// ConflictStrategy defines how to handle configuration conflicts during template application
type ConflictStrategy string

const (
	ConflictStrategyMerge    ConflictStrategy = "MERGE"
	ConflictStrategyReplace  ConflictStrategy = "REPLACE"
	ConflictStrategyPreserve ConflictStrategy = "PRESERVE"
)

// Valid returns true if the conflict strategy is valid
func (cs ConflictStrategy) Valid() bool {
	switch cs {
	case ConflictStrategyMerge, ConflictStrategyReplace, ConflictStrategyPreserve:
		return true
	default:
		return false
	}
}

func (cs ConflictStrategy) String() string {
	return string(cs)
}

// TenantType represents the type of tenant the template applies to
type TenantType string

// TemplateConfiguration represents a single configuration within a template
type TemplateConfiguration struct {
	Module         ModuleName
	ConfigKey      ConfigKey
	Value          ConfigValue
	OverridePolicy OverridePolicy // preserve, replace, merge
	Priority       int
}

// OverridePolicy defines how a template configuration should override existing values
type OverridePolicy string

const (
	OverridePolicyPreserve OverridePolicy = "PRESERVE" // Keep existing value if present
	OverridePolicyReplace  OverridePolicy = "REPLACE"  // Replace existing value
	OverridePolicyMerge    OverridePolicy = "MERGE"    // Merge with existing value (for JSON)
)

// Valid returns true if the override policy is valid
func (op OverridePolicy) Valid() bool {
	switch op {
	case OverridePolicyPreserve, OverridePolicyReplace, OverridePolicyMerge:
		return true
	default:
		return false
	}
}

func (op OverridePolicy) String() string {
	return string(op)
}

// TemplateDependency represents a dependency on another template
type TemplateDependency struct {
	TemplateID      TemplateID
	RequiredVersion string
}

// ConfigurationTarget represents the target for template application
type ConfigurationTarget struct {
	ID         string
	Type       string // tenant, entity
	EntityID   *uuid.UUID
	Properties map[string]any // Additional properties for validation
}

// ApplicationResult represents the result of template application
type ApplicationResult struct {
	TemplateID TemplateID
	TargetID   string
	Applied    []ConfigurationChange
	Conflicts  []ConfigurationConflict
	Summary    ApplicationSummary
}

// ConfigurationChange represents a configuration change made during template application
type ConfigurationChange struct {
	Module    ModuleName
	ConfigKey ConfigKey
	OldValue  *ConfigValue
	NewValue  ConfigValue
	Action    ChangeAction
}

// ChangeAction represents the type of change made
type ChangeAction string

const (
	ChangeActionCreate ChangeAction = "CREATE"
	ChangeActionUpdate ChangeAction = "UPDATE"
	ChangeActionMerge  ChangeAction = "MERGE"
)

// ConfigurationConflict represents a conflict during template application
type ConfigurationConflict struct {
	Module        ModuleName
	ConfigKey     ConfigKey
	TemplateValue ConfigValue
	ExistingValue ConfigValue
	Resolution    ConflictResolution
	Reason        string
}

// ConflictResolution represents how a conflict was resolved
type ConflictResolution string

const (
	ConflictResolutionKeptExisting ConflictResolution = "KEPT_EXISTING"
	ConflictResolutionUsedTemplate ConflictResolution = "USED_TEMPLATE"
	ConflictResolutionMerged       ConflictResolution = "MERGED"
	ConflictResolutionSkipped      ConflictResolution = "SKIPPED"
)

// ApplicationSummary provides a summary of the template application
type ApplicationSummary struct {
	TotalConfigs int
	Applied      int
	Skipped      int
	Conflicts    int
	Errors       int
	DurationMS   int64
}

// NewTemplate creates a new configuration template
func NewTemplate(
	name string,
	category TemplateCategory,
	version string,
	createdBy uuid.UUID,
) (*Template, error) {
	if name == "" {
		return nil, ErrTemplateInvalidVersion // Should be a template name error
	}

	if !category.Valid() {
		return nil, ErrTemplateInvalidCategory
	}

	if version == "" {
		return nil, ErrTemplateInvalidVersion
	}

	template := &Template{
		ID:                 NewTemplateID(),
		Name:               name,
		Category:           category,
		Version:            version,
		IsActive:           true,
		Configurations:     make([]TemplateConfiguration, 0),
		Dependencies:       make([]TemplateDependency, 0),
		ConflictResolution: ConflictStrategyMerge,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		CreatedBy:          createdBy,
	}

	return template, nil
}

// Validate performs domain validation on the template
func (t *Template) Validate() error {
	if t.Name == "" {
		return ErrTemplateInvalidVersion // Should be template name error
	}

	if !t.Category.Valid() {
		return ErrTemplateInvalidCategory
	}

	if t.Version == "" {
		return ErrTemplateInvalidVersion
	}

	if !t.ConflictResolution.Valid() {
		return ErrTemplateConflictResolution
	}

	// Validate all configurations
	for _, config := range t.Configurations {
		if err := config.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates a template configuration
func (tc *TemplateConfiguration) Validate() error {
	if !tc.Module.Valid() {
		return ErrInvalidModule
	}

	if !tc.ConfigKey.IsValid() {
		return ErrInvalidConfigKey
	}

	if err := tc.Value.Validate(); err != nil {
		return err
	}

	if !tc.OverridePolicy.Valid() {
		return ErrTemplateConfigurationInvalid
	}

	return nil
}

// AddConfiguration adds a configuration to the template
func (t *Template) AddConfiguration(config TemplateConfiguration) error {
	if err := config.Validate(); err != nil {
		return err
	}

	// Check for duplicate configurations
	for _, existing := range t.Configurations {
		if existing.Module == config.Module && existing.ConfigKey == config.ConfigKey {
			return ErrTemplateConfigurationInvalid // Duplicate configuration
		}
	}

	t.Configurations = append(t.Configurations, config)
	t.UpdatedAt = time.Now()
	return nil
}

// RemoveConfiguration removes a configuration from the template
func (t *Template) RemoveConfiguration(module ModuleName, configKey ConfigKey) error {
	for i, config := range t.Configurations {
		if config.Module == module && config.ConfigKey == configKey {
			// Remove configuration by slicing
			t.Configurations = append(t.Configurations[:i], t.Configurations[i+1:]...)
			t.UpdatedAt = time.Now()
			return nil
		}
	}

	return ErrConfigurationNotFound
}

// AddDependency adds a template dependency
func (t *Template) AddDependency(dependency TemplateDependency) {
	t.Dependencies = append(t.Dependencies, dependency)
	t.UpdatedAt = time.Now()
}

// IsApplicableTo checks if the template is applicable to the target
func (t *Template) IsApplicableTo(target ConfigurationTarget) bool {
	// Check tenant type compatibility
	if len(t.ApplicableToTenantTypes) > 0 {
		tenantType, exists := target.Properties["tenant_type"].(string)
		if !exists {
			return false
		}

		applicable := false
		for _, applicableType := range t.ApplicableToTenantTypes {
			if string(applicableType) == tenantType {
				applicable = true
				break
			}
		}

		if !applicable {
			return false
		}
	}

	return true
}

// Apply applies the template to a configuration target
func (t *Template) Apply(target ConfigurationTarget) (*ApplicationResult, error) {
	if !t.IsActive {
		return nil, ErrTemplateNotActive
	}

	// Validate target compatibility
	if !t.IsApplicableTo(target) {
		return nil, ErrTemplateNotApplicable
	}

	result := &ApplicationResult{
		TemplateID: t.ID,
		TargetID:   target.ID,
		Applied:    make([]ConfigurationChange, 0),
		Conflicts:  make([]ConfigurationConflict, 0),
		Summary: ApplicationSummary{
			TotalConfigs: len(t.Configurations),
		},
	}

	startTime := time.Now()

	// Apply configurations in priority order
	sortedConfigs := t.getSortedConfigurations()
	for _, config := range sortedConfigs {
		change, conflict, err := t.applyConfiguration(config, target)
		if err != nil {
			result.Summary.Errors++
			continue
		}

		if conflict != nil {
			result.Conflicts = append(result.Conflicts, *conflict)
			result.Summary.Conflicts++
		} else if change != nil {
			result.Applied = append(result.Applied, *change)
			result.Summary.Applied++
		} else {
			result.Summary.Skipped++
		}
	}

	result.Summary.DurationMS = time.Since(startTime).Milliseconds()

	return result, nil
}

// getSortedConfigurations returns configurations sorted by priority
func (t *Template) getSortedConfigurations() []TemplateConfiguration {
	// Create a copy to avoid modifying the original
	configs := make([]TemplateConfiguration, len(t.Configurations))
	copy(configs, t.Configurations)

	// Sort by priority (higher priority first)
	for i := 0; i < len(configs)-1; i++ {
		for j := i + 1; j < len(configs); j++ {
			if configs[i].Priority < configs[j].Priority {
				configs[i], configs[j] = configs[j], configs[i]
			}
		}
	}

	return configs
}

// applyConfiguration applies a single configuration from the template
func (t *Template) applyConfiguration(config TemplateConfiguration, target ConfigurationTarget) (*ConfigurationChange, *ConfigurationConflict, error) {
	// This is a simplified implementation
	// In a real implementation, this would interact with the configuration repository
	// to check for existing values and apply the configuration according to the override policy

	change := &ConfigurationChange{
		Module:    config.Module,
		ConfigKey: config.ConfigKey,
		NewValue:  config.Value,
		Action:    ChangeActionCreate, // Simplified - would be determined by existing state
	}

	return change, nil, nil
}

// Deactivate deactivates the template
func (t *Template) Deactivate() {
	t.IsActive = false
	t.UpdatedAt = time.Now()
}

// Activate activates the template
func (t *Template) Activate() error {
	if err := t.Validate(); err != nil {
		return err
	}

	t.IsActive = true
	t.UpdatedAt = time.Now()
	return nil
}

// Clone creates a copy of the template
func (t *Template) Clone() *Template {
	clone := &Template{
		ID:                      t.ID,
		Name:                    t.Name,
		Category:                t.Category,
		Description:             t.Description,
		Version:                 t.Version,
		IsActive:                t.IsActive,
		ConflictResolution:      t.ConflictResolution,
		CreatedAt:               t.CreatedAt,
		UpdatedAt:               t.UpdatedAt,
		CreatedBy:               t.CreatedBy,
		Configurations:          make([]TemplateConfiguration, len(t.Configurations)),
		Dependencies:            make([]TemplateDependency, len(t.Dependencies)),
		ApplicableToTenantTypes: make([]TenantType, len(t.ApplicableToTenantTypes)),
		RequiredFeatureFlags:    make([]FeatureFlag, len(t.RequiredFeatureFlags)),
	}

	copy(clone.Configurations, t.Configurations)
	copy(clone.Dependencies, t.Dependencies)
	copy(clone.ApplicableToTenantTypes, t.ApplicableToTenantTypes)
	copy(clone.RequiredFeatureFlags, t.RequiredFeatureFlags)

	return clone
}

// GetConfigurationCount returns the number of configurations in the template
func (t *Template) GetConfigurationCount() int {
	return len(t.Configurations)
}

// HasConfiguration checks if the template has a specific configuration
func (t *Template) HasConfiguration(module ModuleName, configKey ConfigKey) bool {
	for _, config := range t.Configurations {
		if config.Module == module && config.ConfigKey == configKey {
			return true
		}
	}
	return false
}

// GetConfiguration retrieves a specific configuration from the template
func (t *Template) GetConfiguration(module ModuleName, configKey ConfigKey) (*TemplateConfiguration, error) {
	for _, config := range t.Configurations {
		if config.Module == module && config.ConfigKey == configKey {
			return &config, nil
		}
	}
	return nil, ErrConfigurationNotFound
}

// String returns a string representation of the template
func (t *Template) String() string {
	return fmt.Sprintf("Template{ID: %s, Name: %s, Category: %s, Version: %s, Active: %t, Configs: %d}",
		t.ID, t.Name, t.Category, t.Version, t.IsActive, len(t.Configurations))
}
