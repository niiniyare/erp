package css

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// PropertyRegistry manages all CSS property definitions and validation
type PropertyRegistry struct {
	properties      map[string]*PropertyDefinition
	dataTypes       map[string]*DataTypeSpec
	schemaDir       string
	schemaValidator *SchemaBasedValidator
	categoryManager *CategoryManager
	mu              sync.RWMutex
}

// PropertyDefinition represents a CSS property with its validation rules
type PropertyDefinition struct {
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Values      []string       `json:"values,omitempty"`
	DataTypes   []string       `json:"dataTypes,omitempty"`
	Inherited   bool           `json:"inherited"`
	Animatable  bool           `json:"animatable"`
	Description string         `json:"description"`
	Schema      map[string]any `json:"schema"`
	GoType      any            `json:"-"` // Will hold reference to generated Go type
}

// DataTypeSpec represents a CSS data type with validation rules
type DataTypeSpec struct {
	Name        string         `json:"name"`
	Pattern     string         `json:"pattern,omitempty"`
	Values      []string       `json:"values,omitempty"`
	Format      string         `json:"format,omitempty"`
	Description string         `json:"description"`
	Schema      map[string]any `json:"schema"`
}

// NewPropertyRegistry creates a new CSS property registry
func NewPropertyRegistry(schemaDir string) *PropertyRegistry {
	registry := &PropertyRegistry{
		properties:      make(map[string]*PropertyDefinition),
		dataTypes:       make(map[string]*DataTypeSpec),
		schemaDir:       schemaDir,
		schemaValidator: NewSchemaBasedValidator(),
		categoryManager: NewCategoryManager(),
	}

	// Load all CSS property and data type definitions
	if err := registry.LoadDefinitions(); err != nil {
		// Log error but continue with empty registry for graceful degradation
		fmt.Printf("Warning: Failed to load CSS definitions: %v\n", err)
	}

	return registry
}

// LoadDefinitions loads all CSS property and data type definitions from schemas
func (pr *PropertyRegistry) LoadDefinitions() error {
	if pr.schemaDir == "" {
		return fmt.Errorf("schema directory not set")
	}

	definitionsDir := filepath.Join(pr.schemaDir, "definitions")

	// Load CSS Property schemas
	if err := pr.loadPropertySchemas(definitionsDir); err != nil {
		return fmt.Errorf("failed to load property schemas: %w", err)
	}

	// Load CSS DataType schemas
	if err := pr.loadDataTypeSchemas(definitionsDir); err != nil {
		return fmt.Errorf("failed to load data type schemas: %w", err)
	}

	return nil
}

// loadPropertySchemas loads all Property.*.json files
func (pr *PropertyRegistry) loadPropertySchemas(definitionsDir string) error {
	pattern := filepath.Join(definitionsDir, "Property.*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob property files: %w", err)
	}

	for _, file := range matches {
		if err := pr.loadPropertySchema(file); err != nil {
			// Log warning but continue loading other files
			fmt.Printf("Warning: Failed to load property schema %s: %v\n", file, err)
			continue
		}
	}

	return nil
}

// loadDataTypeSchemas loads all DataType.*.json files
func (pr *PropertyRegistry) loadDataTypeSchemas(definitionsDir string) error {
	pattern := filepath.Join(definitionsDir, "DataType.*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob data type files: %w", err)
	}

	for _, file := range matches {
		if err := pr.loadDataTypeSchema(file); err != nil {
			// Log warning but continue loading other files
			fmt.Printf("Warning: Failed to load data type schema %s: %v\n", file, err)
			continue
		}
	}

	return nil
}

// loadPropertySchema loads a single property schema file
func (pr *PropertyRegistry) loadPropertySchema(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract property name from filename
	basename := filepath.Base(filePath)
	propertyName := strings.TrimPrefix(basename, "Property.")
	propertyName = strings.TrimSuffix(propertyName, ".json")
	propertyName = strings.ToLower(propertyName)

	// Convert schema to property definition
	definition := &PropertyDefinition{
		Name:   propertyName,
		Schema: schema,
		GoType: pr.getGeneratedGoType(propertyName),
	}

	// Extract additional information from schema
	pr.enrichPropertyDefinition(definition, schema)

	pr.mu.Lock()
	pr.properties[propertyName] = definition
	pr.mu.Unlock()

	return nil
}

// loadDataTypeSchema loads a single data type schema file
func (pr *PropertyRegistry) loadDataTypeSchema(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract data type name from filename
	basename := filepath.Base(filePath)
	dataTypeName := strings.TrimPrefix(basename, "DataType.")
	dataTypeName = strings.TrimSuffix(dataTypeName, ".json")
	dataTypeName = strings.ToLower(dataTypeName)

	definition := &DataTypeSpec{
		Name:        dataTypeName,
		Schema:      schema,
		Values:      []string{},
		Description: "",
		Pattern:     "",
		Format:      "",
	}

	// Extract additional information from schema
	pr.enrichDataTypeDefinition(definition, schema)

	pr.mu.Lock()
	pr.dataTypes[dataTypeName] = definition
	pr.mu.Unlock()

	return nil
}

// getGeneratedGoType returns the generated Go type for a property
func (pr *PropertyRegistry) getGeneratedGoType(propertyName string) any {
	// For now, return nil since we removed the generated types
	// TODO: Re-implement when we have clean generated types
	return nil
}

// enrichPropertyDefinition extracts information from JSON schema to enrich property definition
func (pr *PropertyRegistry) enrichPropertyDefinition(def *PropertyDefinition, schema map[string]any) {
	// Extract type information
	if typeVal, ok := schema["type"]; ok {
		if typeStr, ok := typeVal.(string); ok {
			def.Type = typeStr
		}
	}

	// Extract enum values if present
	if enumVal, ok := schema["enum"]; ok {
		if enumSlice, ok := enumVal.([]any); ok {
			for _, val := range enumSlice {
				if str, ok := val.(string); ok {
					def.Values = append(def.Values, str)
				}
			}
		}
	}

	// Extract anyOf values (common in CSS schemas)
	if anyOfVal, ok := schema["anyOf"]; ok {
		if anyOfSlice, ok := anyOfVal.([]any); ok {
			pr.extractAnyOfValues(def, anyOfSlice)
		}
	}

	// Extract description
	if descVal, ok := schema["description"]; ok {
		if descStr, ok := descVal.(string); ok {
			def.Description = descStr
		}
	}
}

// enrichDataTypeDefinition extracts information from JSON schema to enrich data type definition
func (pr *PropertyRegistry) enrichDataTypeDefinition(def *DataTypeSpec, schema map[string]any) {
	// Extract pattern if present
	if patternVal, ok := schema["pattern"]; ok {
		if patternStr, ok := patternVal.(string); ok {
			def.Pattern = patternStr
		}
	}

	// Extract enum values if present
	if enumVal, ok := schema["enum"]; ok {
		if enumSlice, ok := enumVal.([]any); ok {
			for _, val := range enumSlice {
				if str, ok := val.(string); ok {
					def.Values = append(def.Values, str)
				}
			}
		}
	}

	// Extract description
	if descVal, ok := schema["description"]; ok {
		if descStr, ok := descVal.(string); ok {
			def.Description = descStr
		}
	}
}

// extractAnyOfValues extracts values from anyOf schema constructs
func (pr *PropertyRegistry) extractAnyOfValues(def *PropertyDefinition, anyOfSlice []any) {
	for _, item := range anyOfSlice {
		if itemMap, ok := item.(map[string]any); ok {
			// Check for const values
			if constVal, ok := itemMap["const"]; ok {
				if constStr, ok := constVal.(string); ok {
					def.Values = append(def.Values, constStr)
				}
			}

			// Check for $ref (data type references)
			if refVal, ok := itemMap["$ref"]; ok {
				if refStr, ok := refVal.(string); ok {
					if strings.HasPrefix(refStr, "#/definitions/DataType.") {
						dataTypeName := strings.TrimPrefix(refStr, "#/definitions/DataType.")
						dataTypeName = strings.ToLower(dataTypeName)
						def.DataTypes = append(def.DataTypes, dataTypeName)
					}
				}
			}
		}
	}
}

// ValidateProperty validates a CSS property value against its definition using schema-based validation
func (pr *PropertyRegistry) ValidateProperty(propertyName, value string) error {
	pr.mu.RLock()
	definition, exists := pr.properties[strings.ToLower(propertyName)]
	pr.mu.RUnlock()

	if !exists {
		// Graceful degradation - allow unknown properties
		return nil
	}

	return pr.validateValueWithSchema(value, definition)
}

// validateValue validates a value against a property definition
func (pr *PropertyRegistry) validateValue(value string, definition *PropertyDefinition) error {
	// Check direct enum values first
	for _, validValue := range definition.Values {
		if value == validValue {
			return nil
		}
	}

	// Check data types
	for _, dataTypeName := range definition.DataTypes {
		if err := pr.validateAgainstDataType(value, dataTypeName); err == nil {
			return nil
		}
	}

	// If no validation rules found, allow the value (graceful degradation)
	if len(definition.Values) == 0 && len(definition.DataTypes) == 0 {
		return nil
	}

	return fmt.Errorf("invalid value '%s' for property '%s'", value, definition.Name)
}

// validateValueWithSchema validates a value using enhanced schema-based validation
func (pr *PropertyRegistry) validateValueWithSchema(value string, definition *PropertyDefinition) error {
	// Check direct enum values first
	for _, validValue := range definition.Values {
		if value == validValue {
			return nil
		}
	}

	// Use schema-based validation for data types
	for _, dataTypeName := range definition.DataTypes {
		if pr.schemaValidator != nil {
			if err := pr.schemaValidator.ValidateDataType(dataTypeName, value); err == nil {
				return nil // Valid according to schema-based validator
			}
		}
	}

	// Fallback to original validation method
	return pr.validateValue(value, definition)
}

// validateAgainstDataType validates a value against a data type definition
func (pr *PropertyRegistry) validateAgainstDataType(value, dataTypeName string) error {
	pr.mu.RLock()
	definition, exists := pr.dataTypes[strings.ToLower(dataTypeName)]
	pr.mu.RUnlock()

	if !exists {
		// Graceful degradation - allow unknown data types
		return nil
	}

	// Check enum values if present
	for _, validValue := range definition.Values {
		if value == validValue {
			return nil
		}
	}

	// Pattern validation would go here (regex)
	// For now, we'll be permissive

	return fmt.Errorf("value '%s' does not match data type '%s'", value, dataTypeName)
}

// GetPropertyDefinition returns the definition for a CSS property
func (pr *PropertyRegistry) GetPropertyDefinition(propertyName string) (*PropertyDefinition, bool) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	definition, exists := pr.properties[strings.ToLower(propertyName)]
	return definition, exists
}

// GetDataTypeDefinition returns the definition for a CSS data type
func (pr *PropertyRegistry) GetDataTypeDefinition(dataTypeName string) (*DataTypeSpec, bool) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	definition, exists := pr.dataTypes[strings.ToLower(dataTypeName)]
	return definition, exists
}

// GetAllPropertyNames returns all registered CSS property names
func (pr *PropertyRegistry) GetAllPropertyNames() []string {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	names := make([]string, 0, len(pr.properties))
	for name := range pr.properties {
		names = append(names, name)
	}

	return names
}

// GetAllDataTypeNames returns all registered CSS data type names
func (pr *PropertyRegistry) GetAllDataTypeNames() []string {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	names := make([]string, 0, len(pr.dataTypes))
	for name := range pr.dataTypes {
		names = append(names, name)
	}

	return names
}

// GetPropertyCount returns the number of registered properties
func (pr *PropertyRegistry) GetPropertyCount() int {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	return len(pr.properties)
}

// GetDataTypeCount returns the number of registered data types
func (pr *PropertyRegistry) GetDataTypeCount() int {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	return len(pr.dataTypes)
}

// Enhanced schema-based validation methods

// ValidateDataType validates a value against a specific CSS DataType using schema-based validation
func (pr *PropertyRegistry) ValidateDataType(dataTypeName, value string) error {
	if pr.schemaValidator == nil {
		return fmt.Errorf("schema validator not initialized")
	}
	return pr.schemaValidator.ValidateDataType(dataTypeName, value)
}

// IsSchemaBasedValidationAvailable returns true if schema-based validation is available
func (pr *PropertyRegistry) IsSchemaBasedValidationAvailable() bool {
	return pr.schemaValidator != nil
}

// GetSupportedDataTypes returns all DataTypes supported by schema-based validation
func (pr *PropertyRegistry) GetSupportedDataTypes() []string {
	if pr.schemaValidator == nil {
		return nil
	}
	return pr.schemaValidator.GetSupportedDataTypes()
}

// GetDataTypeInfo returns information about a specific DataType
func (pr *PropertyRegistry) GetDataTypeInfo(dataTypeName string) map[string]any {
	if pr.schemaValidator == nil {
		return map[string]any{
			"name":      dataTypeName,
			"supported": false,
			"error":     "schema validator not initialized",
		}
	}
	return pr.schemaValidator.GetDataTypeInfo(dataTypeName)
}

// Enhanced categorization methods

// GetPropertyCategory returns the category for a CSS property
func (pr *PropertyRegistry) GetPropertyCategory(propertyName string) string {
	if pr.categoryManager == nil {
		return "Other"
	}
	return pr.categoryManager.GetPropertyCategory(propertyName)
}

// GetAllCategories returns all CSS property categories
func (pr *PropertyRegistry) GetAllCategories() []*PropertyCategory {
	if pr.categoryManager == nil {
		return nil
	}
	return pr.categoryManager.GetAllCategories()
}

// GetCategoryNames returns all category names
func (pr *PropertyRegistry) GetCategoryNames() []string {
	if pr.categoryManager == nil {
		return nil
	}
	return pr.categoryManager.GetCategoryNames()
}

// GetPropertiesByCategory returns all properties in a specific category
func (pr *PropertyRegistry) GetPropertiesByCategory(categoryName string) []string {
	if pr.categoryManager == nil {
		return nil
	}
	return pr.categoryManager.GetPropertiesByCategory(categoryName)
}

// GetCategorizedProperties returns all properties organized by category
func (pr *PropertyRegistry) GetCategorizedProperties() map[string][]string {
	if pr.categoryManager == nil {
		return nil
	}
	return pr.categoryManager.GetCategorizedProperties()
}

// SearchProperties searches for properties across all categories
func (pr *PropertyRegistry) SearchProperties(query string) map[string][]string {
	if pr.categoryManager == nil {
		return nil
	}
	return pr.categoryManager.SearchProperties(query)
}

// GetRelatedProperties returns properties related to a given property
func (pr *PropertyRegistry) GetRelatedProperties(propertyName string) []string {
	if pr.categoryManager == nil {
		return nil
	}
	return pr.categoryManager.GetRelatedProperties(propertyName)
}

// GetCategoryStats returns statistics about each category
func (pr *PropertyRegistry) GetCategoryStats() map[string]CategoryStats {
	if pr.categoryManager == nil {
		return nil
	}
	return pr.categoryManager.GetCategoryStats()
}

// GetCategory returns a specific category by name
func (pr *PropertyRegistry) GetCategory(categoryName string) (*PropertyCategory, bool) {
	if pr.categoryManager == nil {
		return nil, false
	}
	return pr.categoryManager.GetCategory(categoryName)
}
