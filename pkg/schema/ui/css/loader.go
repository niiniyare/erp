package css

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// SchemaLoader manages loading and parsing CSS property schemas
type SchemaLoader struct {
	schemaDir string
	cache     map[string]*PropertySchema
}

// PropertySchema represents a CSS property schema definition
type PropertySchema struct {
	ID          string          `json:"$id"`
	Schema      string          `json:"$schema"`
	Type        string          `json:"type,omitempty"`
	AnyOf       []SchemaVariant `json:"anyOf,omitempty"`
	OneOf       []SchemaVariant `json:"oneOf,omitempty"`
	AllOf       []SchemaVariant `json:"allOf,omitempty"`
	Const       any             `json:"const,omitempty"`
	Enum        []any           `json:"enum,omitempty"`
	Properties  map[string]any  `json:"properties,omitempty"`
	Items       *PropertySchema `json:"items,omitempty"`
	Description string          `json:"description,omitempty"`
}

// SchemaVariant represents a variant in anyOf/oneOf/allOf
type SchemaVariant struct {
	Ref         string   `json:"$ref,omitempty"`
	Type        string   `json:"type,omitempty"`
	Const       any      `json:"const,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Minimum     *float64 `json:"minimum,omitempty"`
	Maximum     *float64 `json:"maximum,omitempty"`
	Pattern     string   `json:"pattern,omitempty"`
	Description string   `json:"description,omitempty"`
}

// NewSchemaLoader creates a new CSS schema loader
func NewSchemaLoader(schemaDir string) *SchemaLoader {
	return &SchemaLoader{
		schemaDir: schemaDir,
		cache:     make(map[string]*PropertySchema),
	}
}

// LoadPropertySchema loads a specific CSS property schema by name
func (sl *SchemaLoader) LoadPropertySchema(propertyName string) (*PropertySchema, error) {
	// Check cache first
	if schema, exists := sl.cache[propertyName]; exists {
		return schema, nil
	}

	// Try different file name patterns
	patterns := []string{
		fmt.Sprintf("Property.%s.json", propertyName),
		fmt.Sprintf("DataType.%s.json", propertyName),
		fmt.Sprintf("%s.json", propertyName),
	}

	var schema *PropertySchema
	var err error

	for _, pattern := range patterns {
		filePath := filepath.Join(sl.schemaDir, "definitions", pattern)
		schema, err = sl.loadSchemaFile(filePath)
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load schema for property '%s': %w", propertyName, err)
	}

	// Cache the loaded schema
	sl.cache[propertyName] = schema
	return schema, nil
}

// LoadAllSchemas loads all available CSS property schemas
func (sl *SchemaLoader) LoadAllSchemas() (map[string]*PropertySchema, error) {
	definitionsDir := filepath.Join(sl.schemaDir, "definitions")

	schemas := make(map[string]*PropertySchema)

	err := filepath.WalkDir(definitionsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			schema, loadErr := sl.loadSchemaFile(path)
			if loadErr != nil {
				// Log error but continue with other files
				fmt.Printf("Warning: failed to load schema %s: %v\n", path, loadErr)
				return nil
			}

			// Extract property name from filename
			filename := filepath.Base(path)
			propertyName := strings.TrimSuffix(filename, ".json")

			// Clean up property name (remove Property. or DataType. prefixes)
			if strings.HasPrefix(propertyName, "Property.") {
				propertyName = strings.TrimPrefix(propertyName, "Property.")
			} else if strings.HasPrefix(propertyName, "DataType.") {
				propertyName = strings.TrimPrefix(propertyName, "DataType.")
			}

			schemas[propertyName] = schema
			sl.cache[propertyName] = schema
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk schema directory: %w", err)
	}

	return schemas, nil
}

// loadSchemaFile loads a single schema file
func (sl *SchemaLoader) loadSchemaFile(filePath string) (*PropertySchema, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file %s: %w", filePath, err)
	}

	var schema PropertySchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema JSON in %s: %w", filePath, err)
	}

	return &schema, nil
}

// GetAvailableProperties returns a list of all available CSS properties
func (sl *SchemaLoader) GetAvailableProperties() ([]string, error) {
	schemas, err := sl.LoadAllSchemas()
	if err != nil {
		return nil, err
	}

	properties := make([]string, 0, len(schemas))
	for name := range schemas {
		properties = append(properties, name)
	}

	return properties, nil
}

// ValidateValue validates a value against a specific property schema
func (sl *SchemaLoader) ValidateValue(propertyName string, value any) error {
	schema, err := sl.LoadPropertySchema(propertyName)
	if err != nil {
		return err
	}

	return sl.validateValueAgainstSchema(value, schema)
}

// validateValueAgainstSchema validates a value against a schema
func (sl *SchemaLoader) validateValueAgainstSchema(value any, schema *PropertySchema) error {
	// Handle anyOf validation (value must match at least one variant)
	if len(schema.AnyOf) > 0 {
		for _, variant := range schema.AnyOf {
			if err := sl.validateValueAgainstVariant(value, variant); err == nil {
				return nil // Value matches this variant
			}
		}
		return fmt.Errorf("value %v does not match any allowed variant", value)
	}

	// Handle direct type validation
	if schema.Type != "" {
		return sl.validateTypeMatch(value, schema.Type)
	}

	// Handle const validation
	if schema.Const != nil {
		if value != schema.Const {
			return fmt.Errorf("value must be exactly %v, got %v", schema.Const, value)
		}
		return nil
	}

	// Handle enum validation
	if len(schema.Enum) > 0 {
		for _, enumValue := range schema.Enum {
			if value == enumValue {
				return nil
			}
		}
		return fmt.Errorf("value %v is not one of the allowed enum values: %v", value, schema.Enum)
	}

	return nil
}

// validateValueAgainstVariant validates a value against a schema variant
func (sl *SchemaLoader) validateValueAgainstVariant(value any, variant SchemaVariant) error {
	// Handle reference to another schema
	if variant.Ref != "" {
		// For now, we'll skip reference resolution
		// TODO: In a full implementation, we'd resolve the reference and validate against it
		return nil
	}

	// Handle type validation
	if variant.Type != "" {
		if err := sl.validateTypeMatch(value, variant.Type); err != nil {
			return err
		}
	}

	// Handle const validation
	if variant.Const != nil {
		if value != variant.Const {
			return fmt.Errorf("value must be exactly %v, got %v", variant.Const, value)
		}
	}

	// Handle enum validation
	if len(variant.Enum) > 0 {
		for _, enumValue := range variant.Enum {
			if fmt.Sprintf("%v", value) == enumValue {
				return nil
			}
		}
		return fmt.Errorf("value %v is not one of the allowed enum values: %v", value, variant.Enum)
	}

	return nil
}

// validateTypeMatch validates that a value matches the expected JSON schema type
func (sl *SchemaLoader) validateTypeMatch(value any, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "number":
		switch value.(type) {
		case int, int32, int64, float32, float64:
			return nil
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
	case "integer":
		switch value.(type) {
		case int, int32, int64:
			return nil
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "array":
		// Check if it's a slice or array
		switch value.(type) {
		case []any, []string, []int, []float64:
			return nil
		default:
			return fmt.Errorf("expected array, got %T", value)
		}
	case "object":
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	default:
		// Unknown type, allow it
		return nil
	}

	return nil
}

// GetSchemaInfo returns information about a loaded schema
func (sl *SchemaLoader) GetSchemaInfo(propertyName string) (SchemaInfo, error) {
	schema, err := sl.LoadPropertySchema(propertyName)
	if err != nil {
		return SchemaInfo{}, err
	}

	info := SchemaInfo{
		Name:        propertyName,
		Type:        schema.Type,
		Description: schema.Description,
		ID:          schema.ID,
	}

	// Extract possible values from schema variants
	if len(schema.AnyOf) > 0 {
		for _, variant := range schema.AnyOf {
			if variant.Const != nil {
				info.PossibleValues = append(info.PossibleValues, fmt.Sprintf("%v", variant.Const))
			}
			if len(variant.Enum) > 0 {
				for _, enumVal := range variant.Enum {
					info.PossibleValues = append(info.PossibleValues, enumVal)
				}
			}
			if variant.Type != "" {
				info.AcceptedTypes = append(info.AcceptedTypes, variant.Type)
			}
		}
	}

	return info, nil
}

// SchemaInfo contains information about a CSS property schema
type SchemaInfo struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	Description    string   `json:"description"`
	ID             string   `json:"id"`
	PossibleValues []string `json:"possible_values"`
	AcceptedTypes  []string `json:"accepted_types"`
}

