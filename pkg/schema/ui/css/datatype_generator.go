package css

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

// DataTypeGenerator generates Go types from CSS DataType JSON schemas
type DataTypeGenerator struct {
	schemaDir      string
	outputDir      string
	packageName    string
	dataTypes      map[string]*DataTypeSchema
	generatedTypes map[string]*GeneratedDataType
}

// DataTypeSchema represents a parsed CSS DataType JSON schema
type DataTypeSchema struct {
	ID        string             `json:"$id"`
	Schema    string             `json:"$schema"`
	Type      any                `json:"type"`
	Enum      []any              `json:"enum,omitempty"`
	AnyOf     []SchemaConstraint `json:"anyOf,omitempty"`
	OneOf     []SchemaConstraint `json:"oneOf,omitempty"`
	Const     any                `json:"const,omitempty"`
	Pattern   string             `json:"pattern,omitempty"`
	Format    string             `json:"format,omitempty"`
	Minimum   *float64           `json:"minimum,omitempty"`
	Maximum   *float64           `json:"maximum,omitempty"`
	RawSchema map[string]any     `json:"-"` // Store original for complex schemas
}

// SchemaConstraint represents a constraint in anyOf/oneOf
type SchemaConstraint struct {
	Type    any    `json:"type,omitempty"`
	Const   any    `json:"const,omitempty"`
	Ref     string `json:"$ref,omitempty"`
	Pattern string `json:"pattern,omitempty"`
	Format  string `json:"format,omitempty"`
	Enum    []any  `json:"enum,omitempty"`
}

// GeneratedDataType represents a generated Go type for a CSS DataType
type GeneratedDataType struct {
	Name           string
	GoTypeName     string
	BaseType       string
	IsEnum         bool
	EnumValues     []string
	IsUnion        bool
	UnionTypes     []string
	IsString       bool
	IsNumber       bool
	IsBoth         bool
	Description    string
	Constants      []EnumConstant
	ValidationFunc string
}

// EnumConstant represents an enum constant
type EnumConstant struct {
	Name  string
	Value string
}

// NewDataTypeGenerator creates a new DataType generator
func NewDataTypeGenerator(schemaDir string) *DataTypeGenerator {
	return &DataTypeGenerator{
		schemaDir:      schemaDir,
		outputDir:      filepath.Join("pkg", "schema", "ui", "css", "datatypes"),
		packageName:    "datatypes",
		dataTypes:      make(map[string]*DataTypeSchema),
		generatedTypes: make(map[string]*GeneratedDataType),
	}
}

// LoadAllDataTypes loads all CSS DataType schemas
func (g *DataTypeGenerator) LoadAllDataTypes() error {
	definitionsDir := filepath.Join(g.schemaDir, "definitions")
	pattern := filepath.Join(definitionsDir, "DataType.*.json")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob DataType files: %w", err)
	}

	fmt.Printf("Loading %d CSS DataType schemas...\n", len(matches))

	for _, file := range matches {
		if err := g.loadDataTypeSchema(file); err != nil {
			fmt.Printf("Warning: Failed to load %s: %v\n", filepath.Base(file), err)
			continue
		}
	}

	fmt.Printf("Successfully loaded %d CSS DataType schemas\n", len(g.dataTypes))
	return nil
}

// loadDataTypeSchema loads a single DataType schema
func (g *DataTypeGenerator) loadDataTypeSchema(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse into map first to preserve all data
	var rawSchema map[string]any
	if err := json.Unmarshal(data, &rawSchema); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Parse into structured schema
	var schema DataTypeSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("failed to parse structured JSON: %w", err)
	}

	schema.RawSchema = rawSchema

	// Extract name from filename or $id
	basename := filepath.Base(filePath)
	dataTypeName := strings.TrimPrefix(basename, "DataType.")
	dataTypeName = strings.TrimSuffix(dataTypeName, ".json")

	g.dataTypes[dataTypeName] = &schema
	return nil
}

// GenerateAllTypes generates Go types for all loaded DataType schemas
func (g *DataTypeGenerator) GenerateAllTypes() error {
	fmt.Printf("Generating Go types for %d DataType schemas...\n", len(g.dataTypes))

	for name, schema := range g.dataTypes {
		generatedType, err := g.generateTypeFromSchema(name, schema)
		if err != nil {
			fmt.Printf("Warning: Failed to generate type for %s: %v\n", name, err)
			continue
		}

		if generatedType != nil {
			g.generatedTypes[name] = generatedType
		}
	}

	fmt.Printf("Successfully generated %d Go types\n", len(g.generatedTypes))
	return nil
}

// generateTypeFromSchema generates a Go type from a DataType schema
func (g *DataTypeGenerator) generateTypeFromSchema(name string, schema *DataTypeSchema) (*GeneratedDataType, error) {
	genType := &GeneratedDataType{
		Name:        name,
		GoTypeName:  g.sanitizeTypeName(name),
		Description: fmt.Sprintf("CSS DataType: %s", name),
	}

	// Handle direct enum types
	if len(schema.Enum) > 0 {
		return g.generateEnumType(genType, schema)
	}

	// Handle anyOf/oneOf types
	if len(schema.AnyOf) > 0 {
		return g.generateUnionType(genType, schema, schema.AnyOf)
	}

	if len(schema.OneOf) > 0 {
		return g.generateUnionType(genType, schema, schema.OneOf)
	}

	// Handle simple types
	if schema.Type != nil {
		return g.generateSimpleType(genType, schema)
	}

	// Default to string type
	genType.BaseType = "string"
	genType.IsString = true
	return genType, nil
}

// generateEnumType generates an enum-style type
func (g *DataTypeGenerator) generateEnumType(genType *GeneratedDataType, schema *DataTypeSchema) (*GeneratedDataType, error) {
	genType.IsEnum = true
	genType.BaseType = "string"

	for _, value := range schema.Enum {
		if strVal, ok := value.(string); ok {
			genType.EnumValues = append(genType.EnumValues, strVal)

			// Generate constant name
			constName := g.generateConstantName(genType.GoTypeName, strVal)
			genType.Constants = append(genType.Constants, EnumConstant{
				Name:  constName,
				Value: strVal,
			})
		}
	}

	if len(genType.EnumValues) == 0 {
		return nil, fmt.Errorf("no valid enum values found")
	}

	sort.Strings(genType.EnumValues)
	genType.ValidationFunc = g.generateEnumValidation(genType)

	return genType, nil
}

// generateUnionType generates a union-style type (anyOf/oneOf)
func (g *DataTypeGenerator) generateUnionType(genType *GeneratedDataType, schema *DataTypeSchema, constraints []SchemaConstraint) (*GeneratedDataType, error) {
	genType.IsUnion = true

	var hasString, hasNumber bool
	var enumValues []string

	for _, constraint := range constraints {
		// Handle const values
		if constraint.Const != nil {
			if strVal, ok := constraint.Const.(string); ok {
				enumValues = append(enumValues, strVal)
				hasString = true
			}
		}

		// Handle type constraints
		if constraint.Type != nil {
			switch typeVal := constraint.Type.(type) {
			case string:
				switch typeVal {
				case "string":
					hasString = true
				case "number":
					hasNumber = true
				}
			case []any:
				for _, t := range typeVal {
					if tStr, ok := t.(string); ok {
						switch tStr {
						case "string":
							hasString = true
						case "number":
							hasNumber = true
						}
					}
				}
			}
		}
	}

	// Determine base type
	if hasString && hasNumber {
		genType.BaseType = "string" // Use string as base, document that it can also accept numbers
		genType.IsBoth = true
		genType.UnionTypes = []string{"string", "number"}
	} else if hasNumber {
		genType.BaseType = "float64"
		genType.IsNumber = true
	} else {
		genType.BaseType = "string"
		genType.IsString = true
	}

	// Add enum values if any
	if len(enumValues) > 0 {
		genType.EnumValues = enumValues
		sort.Strings(genType.EnumValues)

		// Generate constants for enum values
		for _, value := range enumValues {
			constName := g.generateConstantName(genType.GoTypeName, value)
			genType.Constants = append(genType.Constants, EnumConstant{
				Name:  constName,
				Value: value,
			})
		}
	}

	genType.ValidationFunc = g.generateUnionValidation(genType)

	return genType, nil
}

// generateSimpleType generates a simple type
func (g *DataTypeGenerator) generateSimpleType(genType *GeneratedDataType, schema *DataTypeSchema) (*GeneratedDataType, error) {
	switch typeVal := schema.Type.(type) {
	case string:
		switch typeVal {
		case "string":
			genType.BaseType = "string"
			genType.IsString = true
		case "number":
			genType.BaseType = "float64"
			genType.IsNumber = true
		case "integer":
			genType.BaseType = "int"
			genType.IsNumber = true
		case "boolean":
			genType.BaseType = "bool"
		default:
			genType.BaseType = "string"
			genType.IsString = true
		}
	default:
		genType.BaseType = "string"
		genType.IsString = true
	}

	genType.ValidationFunc = g.generateSimpleValidation(genType)

	return genType, nil
}

// WriteAllTypes writes all generated types to files
func (g *DataTypeGenerator) WriteAllTypes() error {
	// Ensure output directory exists
	if err := os.MkdirAll(g.outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Printf("Writing %d generated types to %s...\n", len(g.generatedTypes), g.outputDir)

	for name, genType := range g.generatedTypes {
		if err := g.writeTypeToFile(name, genType); err != nil {
			fmt.Printf("Warning: Failed to write type %s: %v\n", name, err)
			continue
		}
	}

	// Write a registry file
	if err := g.writeRegistryFile(); err != nil {
		return fmt.Errorf("failed to write registry file: %w", err)
	}

	fmt.Printf("Successfully wrote all DataType files to %s\n", g.outputDir)
	return nil
}

// writeTypeToFile writes a single type to a file
func (g *DataTypeGenerator) writeTypeToFile(name string, genType *GeneratedDataType) error {
	filename := fmt.Sprintf("%s.go", strings.ToLower(genType.GoTypeName))
	filepath := filepath.Join(g.outputDir, filename)

	content, err := g.generateTypeFileContent(genType)
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	if err := os.WriteFile(filepath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// generateTypeFileContent generates the content for a type file
func (g *DataTypeGenerator) generateTypeFileContent(genType *GeneratedDataType) (string, error) {
	tmpl := template.Must(template.New("datatype").Parse(dataTypeTemplate))

	var content strings.Builder
	if err := tmpl.Execute(&content, genType); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	return content.String(), nil
}

// Helper functions

func (g *DataTypeGenerator) sanitizeTypeName(name string) string {
	// Remove special characters and convert to PascalCase
	sanitized := strings.ReplaceAll(name, "<(string|number)>", "StringOrNumber")
	sanitized = strings.ReplaceAll(sanitized, "<", "")
	sanitized = strings.ReplaceAll(sanitized, ">", "")
	sanitized = strings.ReplaceAll(sanitized, "(", "")
	sanitized = strings.ReplaceAll(sanitized, ")", "")
	sanitized = strings.ReplaceAll(sanitized, "|", "Or")
	sanitized = strings.ReplaceAll(sanitized, ".", "")
	sanitized = strings.ReplaceAll(sanitized, "-", "")
	sanitized = strings.ReplaceAll(sanitized, "_", "")

	return toPascalCase(sanitized)
}

func (g *DataTypeGenerator) generateConstantName(typeName, value string) string {
	// Clean value for constant name
	cleaned := strings.ReplaceAll(value, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "_", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = toPascalCase(cleaned)

	if cleaned == "" {
		cleaned = "Empty"
	}

	return typeName + cleaned
}

func (g *DataTypeGenerator) generateEnumValidation(genType *GeneratedDataType) string {
	values := make([]string, len(genType.EnumValues))
	for i, val := range genType.EnumValues {
		values[i] = fmt.Sprintf(`"%s"`, val)
	}

	return fmt.Sprintf(`func (d %s) IsValid() bool {
	validValues := []string{%s}
	for _, valid := range validValues {
		if string(d) == valid {
			return true
		}
	}
	return false
}`, genType.GoTypeName, strings.Join(values, ", "))
}

func (g *DataTypeGenerator) generateUnionValidation(genType *GeneratedDataType) string {
	if genType.IsBoth {
		return fmt.Sprintf(`// IsValid validates that the value is a non-empty string or valid number
func (d %s) IsValid() bool {
	// Since this is a union type (string|number), we validate as string
	// Numbers should be converted to string when using this type
	return string(d) != ""
}`, genType.GoTypeName)
	}

	return g.generateSimpleValidation(genType)
}

func (g *DataTypeGenerator) generateSimpleValidation(genType *GeneratedDataType) string {
	switch genType.BaseType {
	case "string":
		return fmt.Sprintf(`func (d %s) IsValid() bool {
	return string(d) != ""
}`, genType.GoTypeName)
	default:
		return fmt.Sprintf(`func (d %s) IsValid() bool {
	return true
}`, genType.GoTypeName)
	}
}

// writeRegistryFile writes a registry file that exports all types
func (g *DataTypeGenerator) writeRegistryFile() error {
	filename := "registry.go"
	filepath := filepath.Join(g.outputDir, filename)

	var types []string
	for _, genType := range g.generatedTypes {
		types = append(types, genType.GoTypeName)
	}
	sort.Strings(types)

	content := fmt.Sprintf(`// Code generated by CSS DataType generator. DO NOT EDIT.

package %s

// AllDataTypes returns all generated CSS DataType names
func AllDataTypes() []string {
	return []string{
%s	}
}

// DataTypeCount returns the number of generated CSS DataTypes
func DataTypeCount() int {
	return %d
}
`, g.packageName,
		func() string {
			var lines []string
			for _, t := range types {
				lines = append(lines, fmt.Sprintf("\t\t\"%s\",", t))
			}
			return strings.Join(lines, "\n") + "\n"
		}(),
		len(types))

	return os.WriteFile(filepath, []byte(content), 0o644)
}

func toPascalCase(s string) string {
	if s == "" {
		return ""
	}

	parts := strings.FieldsFunc(s, func(c rune) bool {
		return c == '-' || c == '_' || c == ' ' || c == '.'
	})

	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				result.WriteString(strings.ToLower(part[1:]))
			}
		}
	}

	return result.String()
}

const dataTypeTemplate = `// Code generated by CSS DataType generator. DO NOT EDIT.
// Source: DataType.{{.Name}}.json

package datatypes

{{if or .IsUnion .IsNumber (eq .BaseType "any")}}
import "fmt"
{{end}}

{{if .IsEnum}}
// {{.GoTypeName}} represents the CSS DataType {{.Name}} with predefined values
type {{.GoTypeName}} string

const (
{{range .Constants}}	{{.Name}} {{$.GoTypeName}} = "{{.Value}}"
{{end}})

{{.ValidationFunc}}

// String returns the string representation
func (d {{.GoTypeName}}) String() string {
	return string(d)
}

{{else if .IsUnion}}
// {{.GoTypeName}} represents the CSS DataType {{.Name}} (union type: {{.UnionTypes}})
{{if .IsBoth}}// This type accepts both strings and numbers, but uses string as the underlying type
{{end}}type {{.GoTypeName}} {{.BaseType}}

{{if .Constants}}
// Predefined constants
const (
{{range .Constants}}	{{.Name}} = "{{.Value}}"
{{end}})
{{end}}

{{.ValidationFunc}}

// String returns the string representation
func (d {{.GoTypeName}}) String() string {
	return {{if eq .BaseType "string"}}string(d){{else}}fmt.Sprintf("%v", d){{end}}
}

{{else}}
// {{.GoTypeName}} represents the CSS DataType {{.Name}}
type {{.GoTypeName}} {{.BaseType}}

{{.ValidationFunc}}

// String returns the string representation
func (d {{.GoTypeName}}) String() string {
	return {{if eq .BaseType "string"}}string(d){{else}}fmt.Sprintf("%v", d){{end}}
}

{{end}}
`
