package repository

import (
	"time"

	"github.com/google/uuid"
	"awo/internal/core/settings/domain"
)

// ConfigDefinition represents a configuration definition in the database
type ConfigDefinition struct {
	ID                  uuid.UUID              `json:"id"`
	ModuleName          domain.ModuleName      `json:"module_name"`
	ConfigKey           domain.ConfigKey       `json:"config_key"`
	DataType            domain.DataType        `json:"data_type"`
	DefaultValue        *domain.ConfigValue    `json:"default_value,omitempty"`
	ValidationRules     domain.ValidationRules `json:"validation_rules"`
	Description         string                 `json:"description"`
	RequiredPermission  *domain.Permission     `json:"required_permission,omitempty"`
	RequiredFeatureFlag *domain.FeatureFlag    `json:"required_feature_flag,omitempty"`
	IsOverridable       bool                   `json:"is_overridable"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

// TemplateFilters represents filters for template queries
type TemplateFilters struct {
	Category              *domain.TemplateCategory `json:"category,omitempty"`
	ApplicableTenantTypes []domain.TenantType      `json:"applicable_tenant_types,omitempty"`
	IsActive              *bool                    `json:"is_active,omitempty"`
	Search                string                   `json:"search,omitempty"`
	Limit                 int                      `json:"limit"`
	Offset                int                      `json:"offset"`
}

// TemplateTarget represents the target for template application
type TemplateTarget struct {
	Type     string     `json:"type"` // tenant, entity
	TenantID uuid.UUID  `json:"tenant_id"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"`
}

// ApplyOptions represents options for template application
type ApplyOptions struct {
	PreserveExisting     bool                    `json:"preserve_existing"`
	ConflictResolution   domain.ConflictStrategy `json:"conflict_resolution"`
	ApplyDependencies    bool                    `json:"apply_dependencies"`
	DryRun               bool                    `json:"dry_run"`
	SelectiveApplication *SelectiveApplication   `json:"selective_application,omitempty"`
}

// SelectiveApplication represents selective template application options
type SelectiveApplication struct {
	Modules     []domain.ModuleName `json:"modules,omitempty"`
	ExcludeKeys []domain.ConfigKey  `json:"exclude_keys,omitempty"`
	IncludeKeys []domain.ConfigKey  `json:"include_keys,omitempty"`
}

// TemplateApplication represents a template application record
type TemplateApplication struct {
	ID                 uuid.UUID                  `json:"id"`
	TemplateID         domain.TemplateID          `json:"template_id"`
	TenantID           uuid.UUID                  `json:"tenant_id"`
	EntityID           *uuid.UUID                 `json:"entity_id,omitempty"`
	TargetType         string                     `json:"target_type"`
	AppliedConfigs     int                        `json:"applied_configs"`
	SkippedConfigs     int                        `json:"skipped_configs"`
	ConflictCount      int                        `json:"conflict_count"`
	ApplicationSummary *domain.ApplicationSummary `json:"application_summary,omitempty"`
	AppliedAt          time.Time                  `json:"applied_at"`
	AppliedBy          uuid.UUID                  `json:"applied_by"`
	CorrelationID      *string                    `json:"correlation_id,omitempty"`
}

// AuditRecord represents a configuration audit record
type AuditRecord struct {
	ID            uuid.UUID           `json:"id"`
	TenantID      uuid.UUID           `json:"tenant_id"`
	EntityID      *uuid.UUID          `json:"entity_id,omitempty"`
	ConfigKey     string              `json:"config_key"`
	OldValue      *domain.ConfigValue `json:"old_value,omitempty"`
	NewValue      domain.ConfigValue  `json:"new_value"`
	Source        domain.ConfigSource `json:"source"`
	Operation     string              `json:"operation"`
	UserID        uuid.UUID           `json:"user_id"`
	AppliedAt     time.Time           `json:"applied_at"`
	SessionID     *string             `json:"session_id,omitempty"`
	CorrelationID *string             `json:"correlation_id,omitempty"`
}

// BulkConfigurationUpdate represents a bulk configuration update
type BulkConfigurationUpdate struct {
	TargetType    string             `json:"target_type"` // tenant, entity
	TargetID      uuid.UUID          `json:"target_id"`
	EntityID      *uuid.UUID         `json:"entity_id,omitempty"`
	Module        domain.ModuleName  `json:"module"`
	ConfigKey     domain.ConfigKey   `json:"config_key"`
	Value         domain.ConfigValue `json:"value"`
	UserID        uuid.UUID          `json:"user_id"`
	SessionID     *string            `json:"session_id,omitempty"`
	CorrelationID string             `json:"correlation_id"`
}

// BulkOperationResult represents the result of a bulk operation
type BulkOperationResult struct {
	OperationID       string               `json:"operation_id"`
	TotalTargets      int                  `json:"total_targets"`
	ProcessedTargets  int                  `json:"processed_targets"`
	SuccessfulUpdates int                  `json:"successful_updates"`
	FailedUpdates     int                  `json:"failed_updates"`
	Errors            []BulkOperationError `json:"errors,omitempty"`
	StartedAt         time.Time            `json:"started_at"`
	CompletedAt       *time.Time           `json:"completed_at,omitempty"`
	DurationMS        int64                `json:"duration_ms"`
}

// BulkOperationError represents an error during bulk operations
type BulkOperationError struct {
	TargetID  string `json:"target_id"`
	ConfigKey string `json:"config_key"`
	Error     string `json:"error"`
}

// SearchCriteria represents search criteria for configurations
type SearchCriteria struct {
	EntityID       *uuid.UUID            `json:"entity_id,omitempty"`
	Modules        []domain.ModuleName   `json:"modules,omitempty"`
	ConfigKeys     []string              `json:"config_keys,omitempty"`
	ValueContains  string                `json:"value_contains,omitempty"`
	Sources        []domain.ConfigSource `json:"sources,omitempty"`
	DataTypes      []domain.DataType     `json:"data_types,omitempty"`
	ModifiedAfter  *time.Time            `json:"modified_after,omitempty"`
	ModifiedBefore *time.Time            `json:"modified_before,omitempty"`
	SortBy         string                `json:"sort_by,omitempty"`    // module, updated_at, config_key
	SortOrder      string                `json:"sort_order,omitempty"` // asc, desc
	Limit          int                   `json:"limit"`
	Offset         int                   `json:"offset"`
}

// SearchResult represents the result of a configuration search
type SearchResult struct {
	Configurations []*ConfigurationSearchResult `json:"configurations"`
	TotalCount     int                          `json:"total_count"`
	QueryTimeMS    int64                        `json:"query_time_ms"`
	Filters        SearchCriteria               `json:"filters"`
}

// ConfigurationSearchResult represents a single configuration in search results
type ConfigurationSearchResult struct {
	Module    domain.ModuleName   `json:"module"`
	ConfigKey domain.ConfigKey    `json:"config_key"`
	FullKey   string              `json:"full_key"`
	Value     domain.ConfigValue  `json:"value"`
	Source    domain.ConfigSource `json:"source"`
	TenantID  uuid.UUID           `json:"tenant_id"`
	EntityID  *uuid.UUID          `json:"entity_id,omitempty"`
	DataType  domain.DataType     `json:"data_type"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// ConfigurationUsageAnalysis represents usage analysis results
type ConfigurationUsageAnalysis struct {
	AnalysisPeriod        string                `json:"analysis_period"`
	ModuleSummaries       []ModuleUsageSummary  `json:"module_summaries"`
	ConfigurationPatterns ConfigurationPatterns `json:"configuration_patterns"`
	TemplateAdoption      TemplateAdoptionStats `json:"template_adoption"`
}

// ModuleUsageSummary represents usage summary for a module
type ModuleUsageSummary struct {
	Module          domain.ModuleName   `json:"module"`
	TotalConfigs    int                 `json:"total_configs"`
	EntityOverrides int                 `json:"entity_overrides"`
	MostCustomized  []CustomizationStat `json:"most_customized"`
}

// CustomizationStat represents customization statistics
type CustomizationStat struct {
	ConfigKey          domain.ConfigKey `json:"config_key"`
	OverridePercentage float64          `json:"override_percentage"`
	EntityCount        int              `json:"entity_count"`
}

// ConfigurationPatterns represents configuration usage patterns
type ConfigurationPatterns struct {
	MostOverridden  []string `json:"most_overridden"`
	NeverOverridden []string `json:"never_overridden"`
	RecentlyChanged []string `json:"recently_changed"`
}

// TemplateAdoptionStats represents template adoption statistics
type TemplateAdoptionStats struct {
	TotalApplications      int     `json:"total_applications"`
	MostAppliedTemplate    string  `json:"most_applied_template"`
	RecentApplications     int     `json:"recent_applications"`
	AverageConfigsPerApply float64 `json:"average_configs_per_apply"`
}

// ConfigurationExport represents configuration data for export
type ConfigurationExport struct {
	ExportID       string                 `json:"export_id"`
	TenantID       uuid.UUID              `json:"tenant_id"`
	EntityID       *uuid.UUID             `json:"entity_id,omitempty"`
	ExportType     string                 `json:"export_type"` // full, selective, template
	Configurations []*ExportConfiguration `json:"configurations"`
	Metadata       ExportMetadata         `json:"metadata"`
	ExportedAt     time.Time              `json:"exported_at"`
	ExportedBy     uuid.UUID              `json:"exported_by"`
}

// ExportConfiguration represents a configuration in an export
type ExportConfiguration struct {
	Module      domain.ModuleName   `json:"module"`
	ConfigKey   domain.ConfigKey    `json:"config_key"`
	Value       domain.ConfigValue  `json:"value"`
	Source      domain.ConfigSource `json:"source"`
	DataType    domain.DataType     `json:"data_type"`
	IsInherited bool                `json:"is_inherited"`
	CanOverride bool                `json:"can_override"`
}

// ExportMetadata represents metadata for configuration exports
type ExportMetadata struct {
	TotalConfigurations int      `json:"total_configurations"`
	ModulesIncluded     []string `json:"modules_included"`
	SourcesIncluded     []string `json:"sources_included"`
	ExportFormat        string   `json:"export_format"` // json, yaml, csv
	CompressionUsed     bool     `json:"compression_used"`
	EncryptionUsed      bool     `json:"encryption_used"`
	SchemaVersion       string   `json:"schema_version"`
}

// ConfigurationImport represents configuration data for import
type ConfigurationImport struct {
	ImportID       string                 `json:"import_id"`
	TenantID       uuid.UUID              `json:"tenant_id"`
	EntityID       *uuid.UUID             `json:"entity_id,omitempty"`
	ImportType     string                 `json:"import_type"` // full, selective, merge
	Configurations []*ImportConfiguration `json:"configurations"`
	Options        ImportOptions          `json:"options"`
	ImportedAt     time.Time              `json:"imported_at"`
	ImportedBy     uuid.UUID              `json:"imported_by"`
}

// ImportConfiguration represents a configuration to be imported
type ImportConfiguration struct {
	Module         domain.ModuleName  `json:"module"`
	ConfigKey      domain.ConfigKey   `json:"config_key"`
	Value          domain.ConfigValue `json:"value"`
	OverridePolicy string             `json:"override_policy"` // skip, replace, merge
}

// ImportOptions represents options for configuration import
type ImportOptions struct {
	ConflictResolution   domain.ConflictStrategy `json:"conflict_resolution"`
	ValidateBeforeImport bool                    `json:"validate_before_import"`
	CreateBackup         bool                    `json:"create_backup"`
	DryRun               bool                    `json:"dry_run"`
	IgnoreErrors         bool                    `json:"ignore_errors"`
}

// ImportResult represents the result of a configuration import
type ImportResult struct {
	ImportID        string           `json:"import_id"`
	TotalConfigs    int              `json:"total_configs"`
	ImportedConfigs int              `json:"imported_configs"`
	SkippedConfigs  int              `json:"skipped_configs"`
	ErrorConfigs    int              `json:"error_configs"`
	Conflicts       []ImportConflict `json:"conflicts,omitempty"`
	Errors          []ImportError    `json:"errors,omitempty"`
	BackupID        *string          `json:"backup_id,omitempty"`
	DurationMS      int64            `json:"duration_ms"`
}

// ImportConflict represents a conflict during import
type ImportConflict struct {
	ConfigKey     string             `json:"config_key"`
	ImportValue   domain.ConfigValue `json:"import_value"`
	ExistingValue domain.ConfigValue `json:"existing_value"`
	Resolution    string             `json:"resolution"`
	Reason        string             `json:"reason"`
}

// ImportError represents an error during import
type ImportError struct {
	ConfigKey  string `json:"config_key"`
	Error      string `json:"error"`
	LineNumber *int   `json:"line_number,omitempty"`
}
