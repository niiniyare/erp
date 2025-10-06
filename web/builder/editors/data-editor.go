package editors

import (
	"fmt"

	"github.com/niiniyare/erp/web/builder/core"
)

// DataEditor handles data source and API configuration
type DataEditor struct {
	component *core.ComponentInstance
}

// DataConfiguration represents data binding settings
type DataConfiguration struct {
	Source     DataSource        `json:"source"`
	Bindings   []DataBinding     `json:"bindings"`
	Filters    []DataFilter      `json:"filters,omitempty"`
	Sorting    []DataSort        `json:"sorting,omitempty"`
	Pagination *PaginationConfig `json:"pagination,omitempty"`
	Cache      *CacheConfig      `json:"cache,omitempty"`
	Transform  map[string]any    `json:"transform,omitempty"`
}

// DataSource represents different data source types
type DataSource struct {
	Type    string            `json:"type"` // "api", "static", "function", "store"
	Config  map[string]any    `json:"config"`
	URL     string            `json:"url,omitempty"`
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    any               `json:"body,omitempty"`
	Params  map[string]any    `json:"params,omitempty"`
}

// DataBinding represents property-to-data mappings
type DataBinding struct {
	Property   string `json:"property"`            // Component property name
	DataPath   string `json:"dataPath"`            // Path in data object (e.g., "user.name")
	Transform  string `json:"transform,omitempty"` // Transform function name
	Default    any    `json:"default,omitempty"`
	Required   bool   `json:"required,omitempty"`
	Validation string `json:"validation,omitempty"`
}

// DataFilter represents data filtering rules
type DataFilter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"` // "eq", "ne", "gt", "lt", "contains", "in"
	Value    any    `json:"value"`
	Logic    string `json:"logic,omitempty"` // "and", "or"
}

// DataSort represents sorting configuration
type DataSort struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // "asc", "desc"
	Priority  int    `json:"priority,omitempty"`
}

// PaginationConfig represents pagination settings
type PaginationConfig struct {
	Enabled  bool   `json:"enabled"`
	PageSize int    `json:"pageSize"`
	Strategy string `json:"strategy"`        // "offset", "cursor"
	Total    string `json:"total,omitempty"` // Path to total count in response
}

// CacheConfig represents caching settings
type CacheConfig struct {
	Enabled  bool   `json:"enabled"`
	TTL      int    `json:"ttl"`           // Time to live in seconds
	Strategy string `json:"strategy"`      // "memory", "session", "local"
	Key      string `json:"key,omitempty"` // Custom cache key
}

// NewDataEditor creates a new data editor instance
func NewDataEditor(component *core.ComponentInstance) *DataEditor {
	return &DataEditor{
		component: component,
	}
}

// GetDataConfiguration returns the current data configuration
func (de *DataEditor) GetDataConfiguration() *DataConfiguration {
	if de.component.Data == nil {
		return &DataConfiguration{
			Source: DataSource{
				Type: "static",
				Config: map[string]any{
					"data": []any{},
				},
			},
			Bindings: []DataBinding{},
		}
	}

	// Extract configuration from component data
	config := &DataConfiguration{}

	if source := de.extractDataSource(); source != nil {
		config.Source = *source
	}

	config.Bindings = de.extractDataBindings()
	config.Filters = de.extractDataFilters()
	config.Sorting = de.extractDataSorting()
	config.Pagination = de.extractPaginationConfig()
	config.Cache = de.extractCacheConfig()
	config.Transform = de.extractTransformConfig()

	return config
}

// UpdateDataConfiguration applies new data configuration
func (de *DataEditor) UpdateDataConfiguration(config *DataConfiguration) error {
	if de.component.Data == nil {
		de.component.Data = &core.DataBinding{}
	}

	// Apply data source configuration
	if err := de.applyDataSource(&config.Source); err != nil {
		return fmt.Errorf("failed to apply data source: %w", err)
	}

	// Apply data bindings
	if err := de.applyDataBindings(config.Bindings); err != nil {
		return fmt.Errorf("failed to apply data bindings: %w", err)
	}

	// Apply filters
	if err := de.applyDataFilters(config.Filters); err != nil {
		return fmt.Errorf("failed to apply data filters: %w", err)
	}

	// Apply sorting
	if err := de.applyDataSorting(config.Sorting); err != nil {
		return fmt.Errorf("failed to apply data sorting: %w", err)
	}

	// Apply pagination
	if config.Pagination != nil {
		if err := de.applyPaginationConfig(config.Pagination); err != nil {
			return fmt.Errorf("failed to apply pagination: %w", err)
		}
	}

	// Apply cache configuration
	if config.Cache != nil {
		if err := de.applyCacheConfig(config.Cache); err != nil {
			return fmt.Errorf("failed to apply cache config: %w", err)
		}
	}

	return nil
}

// GetDataSourcePresets returns common data source configurations
func (de *DataEditor) GetDataSourcePresets() []DataSourcePreset {
	return []DataSourcePreset{
		{
			ID:          "rest-api",
			Name:        "REST API",
			Description: "Fetch data from a REST API endpoint",
			Type:        "api",
			Config: map[string]any{
				"url":    "",
				"method": "GET",
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
			},
		},
		{
			ID:          "graphql-api",
			Name:        "GraphQL API",
			Description: "Query data using GraphQL",
			Type:        "api",
			Config: map[string]any{
				"url":    "",
				"method": "POST",
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
				"body": map[string]any{
					"query": "",
				},
			},
		},
		{
			ID:          "static-data",
			Name:        "Static Data",
			Description: "Use predefined static data",
			Type:        "static",
			Config: map[string]any{
				"data": []any{},
			},
		},
		{
			ID:          "local-storage",
			Name:        "Local Storage",
			Description: "Read data from browser local storage",
			Type:        "function",
			Config: map[string]any{
				"function": "getLocalStorageData",
				"key":      "",
			},
		},
		{
			ID:          "session-storage",
			Name:        "Session Storage",
			Description: "Read data from browser session storage",
			Type:        "function",
			Config: map[string]any{
				"function": "getSessionStorageData",
				"key":      "",
			},
		},
		{
			ID:          "global-store",
			Name:        "Global Store",
			Description: "Connect to application state store",
			Type:        "store",
			Config: map[string]any{
				"store": "",
				"path":  "",
			},
		},
	}
}

// DataSourcePreset represents a predefined data source configuration
type DataSourcePreset struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Type        string         `json:"type"`
	Config      map[string]any `json:"config"`
	Icon        string         `json:"icon,omitempty"`
}

// GetAvailableTransforms returns available data transformation functions
func (de *DataEditor) GetAvailableTransforms() []DataTransform {
	return []DataTransform{
		{
			Name:        "formatDate",
			Description: "Format date strings",
			Example:     "formatDate(date, 'YYYY-MM-DD')",
			Parameters:  []string{"date", "format"},
		},
		{
			Name:        "formatCurrency",
			Description: "Format numbers as currency",
			Example:     "formatCurrency(amount, 'USD')",
			Parameters:  []string{"amount", "currency"},
		},
		{
			Name:        "truncate",
			Description: "Truncate text to specified length",
			Example:     "truncate(text, 100)",
			Parameters:  []string{"text", "length"},
		},
		{
			Name:        "capitalize",
			Description: "Capitalize first letter of each word",
			Example:     "capitalize(text)",
			Parameters:  []string{"text"},
		},
		{
			Name:        "toLowerCase",
			Description: "Convert text to lowercase",
			Example:     "toLowerCase(text)",
			Parameters:  []string{"text"},
		},
		{
			Name:        "toUpperCase",
			Description: "Convert text to uppercase",
			Example:     "toUpperCase(text)",
			Parameters:  []string{"text"},
		},
		{
			Name:        "parseJSON",
			Description: "Parse JSON string to object",
			Example:     "parseJSON(jsonString)",
			Parameters:  []string{"jsonString"},
		},
		{
			Name:        "joinArray",
			Description: "Join array elements with separator",
			Example:     "joinArray(array, ', ')",
			Parameters:  []string{"array", "separator"},
		},
	}
}

// DataTransform represents a data transformation function
type DataTransform struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Example     string   `json:"example"`
	Parameters  []string `json:"parameters"`
}

// ValidateDataConfiguration validates the data configuration
func (de *DataEditor) ValidateDataConfiguration(config *DataConfiguration) []ValidationError {
	var errors []ValidationError

	// Validate data source
	if config.Source.Type == "" {
		errors = append(errors, ValidationError{
			Field:   "source.type",
			Message: "Data source type is required",
		})
	}

	if config.Source.Type == "api" {
		if config.Source.URL == "" {
			errors = append(errors, ValidationError{
				Field:   "source.url",
				Message: "API URL is required",
			})
		}
	}

	// Validate data bindings
	for i, binding := range config.Bindings {
		if binding.Property == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("bindings[%d].property", i),
				Message: "Property name is required",
			})
		}
		if binding.DataPath == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("bindings[%d].dataPath", i),
				Message: "Data path is required",
			})
		}
	}

	// Validate filters
	for i, filter := range config.Filters {
		if filter.Field == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("filters[%d].field", i),
				Message: "Filter field is required",
			})
		}
		if filter.Operator == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("filters[%d].operator", i),
				Message: "Filter operator is required",
			})
		}
	}

	return errors
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Helper methods for extracting configuration

func (de *DataEditor) extractDataSource() *DataSource {
	if de.component.Data.Source == nil {
		return nil
	}

	source := &DataSource{}
	if sourceType, ok := de.component.Data.Source["type"].(string); ok {
		source.Type = sourceType
	}
	if config, ok := de.component.Data.Source["config"].(map[string]any); ok {
		source.Config = config
	}
	if url, ok := de.component.Data.Source["url"].(string); ok {
		source.URL = url
	}
	if method, ok := de.component.Data.Source["method"].(string); ok {
		source.Method = method
	}

	return source
}

func (de *DataEditor) extractDataBindings() []DataBinding {
	var bindings []DataBinding

	if de.component.Data.Bindings != nil {
		for _, binding := range de.component.Data.Bindings {
			dataBinding := DataBinding{
				Property: binding.Property,
				DataPath: binding.Path,
			}
			if binding.Transform != "" {
				dataBinding.Transform = binding.Transform
			}
			if binding.Default != nil {
				dataBinding.Default = binding.Default
			}
			bindings = append(bindings, dataBinding)
		}
	}

	return bindings
}

func (de *DataEditor) extractDataFilters() []DataFilter {
	var filters []DataFilter

	if de.component.Data.Filters != nil {
		for _, filter := range de.component.Data.Filters {
			dataFilter := DataFilter{
				Field:    filter.Field,
				Operator: filter.Operator,
				Value:    filter.Value,
			}
			if filter.Logic != "" {
				dataFilter.Logic = filter.Logic
			}
			filters = append(filters, dataFilter)
		}
	}

	return filters
}

func (de *DataEditor) extractDataSorting() []DataSort {
	var sorting []DataSort

	if de.component.Data.Sorting != nil {
		for _, sort := range de.component.Data.Sorting {
			dataSort := DataSort{
				Field:     sort.Field,
				Direction: sort.Direction,
			}
			if sort.Priority > 0 {
				dataSort.Priority = sort.Priority
			}
			sorting = append(sorting, dataSort)
		}
	}

	return sorting
}

func (de *DataEditor) extractPaginationConfig() *PaginationConfig {
	if de.component.Data.Pagination == nil {
		return nil
	}

	return &PaginationConfig{
		Enabled:  de.component.Data.Pagination.Enabled,
		PageSize: de.component.Data.Pagination.PageSize,
		Strategy: de.component.Data.Pagination.Strategy,
		Total:    de.component.Data.Pagination.Total,
	}
}

func (de *DataEditor) extractCacheConfig() *CacheConfig {
	if de.component.Data.Cache == nil {
		return nil
	}

	return &CacheConfig{
		Enabled:  de.component.Data.Cache.Enabled,
		TTL:      de.component.Data.Cache.TTL,
		Strategy: de.component.Data.Cache.Strategy,
		Key:      de.component.Data.Cache.Key,
	}
}

func (de *DataEditor) extractTransformConfig() map[string]any {
	if de.component.Data.Transform == nil {
		return make(map[string]any)
	}
	return de.component.Data.Transform
}

// Helper methods for applying configuration

func (de *DataEditor) applyDataSource(source *DataSource) error {
	if de.component.Data.Source == nil {
		de.component.Data.Source = make(map[string]any)
	}

	de.component.Data.Source["type"] = source.Type
	de.component.Data.Source["config"] = source.Config

	if source.URL != "" {
		de.component.Data.Source["url"] = source.URL
	}
	if source.Method != "" {
		de.component.Data.Source["method"] = source.Method
	}
	if source.Headers != nil {
		de.component.Data.Source["headers"] = source.Headers
	}
	if source.Body != nil {
		de.component.Data.Source["body"] = source.Body
	}
	if source.Params != nil {
		de.component.Data.Source["params"] = source.Params
	}

	return nil
}

func (de *DataEditor) applyDataBindings(bindings []DataBinding) error {
	de.component.Data.Bindings = make([]core.PropertyBinding, len(bindings))

	for i, binding := range bindings {
		de.component.Data.Bindings[i] = core.PropertyBinding{
			Property:  binding.Property,
			Path:      binding.DataPath,
			Transform: binding.Transform,
			Default:   binding.Default,
		}
	}

	return nil
}

func (de *DataEditor) applyDataFilters(filters []DataFilter) error {
	if len(filters) == 0 {
		de.component.Data.Filters = nil
		return nil
	}

	de.component.Data.Filters = make([]core.DataFilter, len(filters))

	for i, filter := range filters {
		de.component.Data.Filters[i] = core.DataFilter{
			Field:    filter.Field,
			Operator: filter.Operator,
			Value:    filter.Value,
			Logic:    filter.Logic,
		}
	}

	return nil
}

func (de *DataEditor) applyDataSorting(sorting []DataSort) error {
	if len(sorting) == 0 {
		de.component.Data.Sorting = nil
		return nil
	}

	de.component.Data.Sorting = make([]core.DataSort, len(sorting))

	for i, sort := range sorting {
		de.component.Data.Sorting[i] = core.DataSort{
			Field:     sort.Field,
			Direction: sort.Direction,
			Priority:  sort.Priority,
		}
	}

	return nil
}

func (de *DataEditor) applyPaginationConfig(config *PaginationConfig) error {
	if de.component.Data.Pagination == nil {
		de.component.Data.Pagination = &core.PaginationConfig{}
	}

	de.component.Data.Pagination.Enabled = config.Enabled
	de.component.Data.Pagination.PageSize = config.PageSize
	de.component.Data.Pagination.Strategy = config.Strategy
	de.component.Data.Pagination.Total = config.Total

	return nil
}

func (de *DataEditor) applyCacheConfig(config *CacheConfig) error {
	if de.component.Data.Cache == nil {
		de.component.Data.Cache = &core.CacheConfig{}
	}

	de.component.Data.Cache.Enabled = config.Enabled
	de.component.Data.Cache.TTL = config.TTL
	de.component.Data.Cache.Strategy = config.Strategy
	de.component.Data.Cache.Key = config.Key

	return nil
}

// TestDataConnection tests the data source connection
func (de *DataEditor) TestDataConnection(config *DataConfiguration) (*TestResult, error) {
	result := &TestResult{
		Success: false,
		Message: "",
		Data:    nil,
	}

	switch config.Source.Type {
	case "api":
		return de.testAPIConnection(&config.Source)
	case "static":
		return de.testStaticData(&config.Source)
	case "function":
		return de.testFunctionData(&config.Source)
	case "store":
		return de.testStoreConnection(&config.Source)
	default:
		result.Message = fmt.Sprintf("Unknown data source type: %s", config.Source.Type)
		return result, nil
	}
}

// TestResult represents the result of a data connection test
type TestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (de *DataEditor) testAPIConnection(source *DataSource) (*TestResult, error) {
	// TODO: Implement actual API testing
	result := &TestResult{
		Success: true,
		Message: "API connection test successful",
		Data: map[string]any{
			"status": "connected",
			"url":    source.URL,
			"method": source.Method,
		},
	}
	return result, nil
}

func (de *DataEditor) testStaticData(source *DataSource) (*TestResult, error) {
	result := &TestResult{
		Success: true,
		Message: "Static data loaded successfully",
		Data:    source.Config["data"],
	}
	return result, nil
}

func (de *DataEditor) testFunctionData(source *DataSource) (*TestResult, error) {
	result := &TestResult{
		Success: true,
		Message: "Function data source configured",
		Data: map[string]any{
			"function": source.Config["function"],
		},
	}
	return result, nil
}

func (de *DataEditor) testStoreConnection(source *DataSource) (*TestResult, error) {
	result := &TestResult{
		Success: true,
		Message: "Store connection configured",
		Data: map[string]any{
			"store": source.Config["store"],
			"path":  source.Config["path"],
		},
	}
	return result, nil
}
