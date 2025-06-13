package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// DocType represents a document type configuration
type DocType struct {
	Name        string `json:"name"`
	IsSingle    bool   `json:"issingle"`
	Fields      []Field `json:"fields"`
	Permissions []Permission `json:"permissions"`
	Engine      string `json:"engine"`
	RowFormat   string `json:"row_format"`
}

// Field represents a field in a document type
type Field struct {
	Fieldname string  `json:"fieldname"`
	Fieldtype string  `json:"fieldtype"`
	Default   *string `json:"default"`
	Required  bool    `json:"reqd"`
	Options   *string `json:"options,omitempty"`
}

// Permission represents access permissions for a role
type Permission struct {
	Role string `json:"role"`
	Read bool   `json:"read"`
}

// UnmarshalJSON implements custom JSON unmarshaling for DocType
func (d *DocType) UnmarshalJSON(data []byte) error {
	type Alias DocType
	aux := &struct {
		IsSingle int `json:"issingle"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}
	
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	d.IsSingle = aux.IsSingle == 1
	return nil
}

// UnmarshalJSON implements custom JSON unmarshaling for Field
func (f *Field) UnmarshalJSON(data []byte) error {
	type Alias Field
	aux := &struct {
		Required int `json:"reqd"`
		*Alias
	}{
		Alias: (*Alias)(f),
	}
	
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	f.Required = aux.Required == 1
	return nil
}

// UnmarshalJSON implements custom JSON unmarshaling for Permission
func (p *Permission) UnmarshalJSON(data []byte) error {
	type Alias Permission
	aux := &struct {
		Read int `json:"read"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	p.Read = aux.Read == 1
	return nil
}

// PostgresSchemaGenerator generates PostgreSQL schema from DocType
type PostgresSchemaGenerator struct{}

// NewPostgresSchemaGenerator creates a new schema generator
func NewPostgresSchemaGenerator() *PostgresSchemaGenerator {
	return &PostgresSchemaGenerator{}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <doctype1.json> [<doctype2.json> ...]")
	}

	generator := NewPostgresSchemaGenerator()
	
	for _, filePath := range os.Args[1:] {
		if err := processFile(filePath, generator); err != nil {
			log.Printf("Error processing file %s: %v", filePath, err)
		}
	}
}

func processFile(filePath string, generator *PostgresSchemaGenerator) error {
	doc, err := parseDocType(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse doctype: %w", err)
	}

	schema := generator.Generate(doc)
	fmt.Print(schema)
	return nil
}

func parseDocType(path string) (*DocType, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	var doc DocType
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	return &doc, nil
}

// Generate creates a PostgreSQL schema from a DocType
func (g *PostgresSchemaGenerator) Generate(doc *DocType) string {
	var b strings.Builder
	tableName := "tab" + doc.Name

	g.writeHeader(&b, tableName, doc)
	g.writeTableDefinition(&b, tableName, doc)
	g.writeTableComment(&b, tableName, doc)
	g.writePermissionsComment(&b, doc)

	return b.String()
}

func (g *PostgresSchemaGenerator) writeHeader(b *strings.Builder, tableName string, doc *DocType) {
	fmt.Fprintf(b, "-- Table: %s\n", tableName)
	fmt.Fprintf(b, "-- Description: doctype: %s\n", doc.Name)
	fmt.Fprintf(b, "-- Single Document: %t\n", doc.IsSingle)
	fmt.Fprintf(b, "-- Engine: %s, Row Format: %s\n\n", doc.Engine, doc.RowFormat)
}

func (g *PostgresSchemaGenerator) writeTableDefinition(b *strings.Builder, tableName string, doc *DocType) {
	fmt.Fprintf(b, "CREATE TABLE %s (\n", quoteIdentifier(tableName))
	
	// Write standard fields
	g.writeStandardFields(b)
	
	// Write custom fields
	g.writeCustomFields(b, doc.Fields)
	
	// Remove trailing comma and close table definition
	schema := b.String()
	schema = strings.TrimSuffix(schema, ",\n") + "\n);\n\n"
	
	// Reset builder and write the corrected schema
	b.Reset()
	b.WriteString(schema)
}

func (g *PostgresSchemaGenerator) writeStandardFields(b *strings.Builder) {
	standardFields := []string{
		"name VARCHAR(140) PRIMARY KEY",
		"creation TIMESTAMP WITH TIME ZONE",
		"modified TIMESTAMP WITH TIME ZONE",
		"modified_by VARCHAR(140)",
		"owner VARCHAR(140)",
		"docstatus SMALLINT DEFAULT 0 NOT NULL",
		"parent VARCHAR(140)",
		"parentfield VARCHAR(140)",
		"parenttype VARCHAR(140)",
		"idx INTEGER DEFAULT 0 NOT NULL",
	}

	for _, field := range standardFields {
		fmt.Fprintf(b, "    %s,\n", field)
	}
}

func (g *PostgresSchemaGenerator) writeCustomFields(b *strings.Builder, fields []Field) {
	for _, field := range fields {
		if shouldSkipField(field.Fieldtype) {
			continue
		}
		
		pgType := mapToPostgresType(field)
		defaultClause := getDefaultValue(field, pgType)
		
		var requiredClause string
		if field.Required {
			requiredClause = " NOT NULL"
		}
		
		column := fmt.Sprintf("%s %s%s%s", 
			quoteIdentifier(field.Fieldname), 
			pgType, 
			defaultClause, 
			requiredClause)
		
		fmt.Fprintf(b, "    %s,\n", column)
	}
}

func (g *PostgresSchemaGenerator) writeTableComment(b *strings.Builder, tableName string, doc *DocType) {
	fmt.Fprintf(b, "COMMENT ON TABLE %s IS 'doctype: %s';\n\n", 
		quoteIdentifier(tableName), doc.Name)
}

func (g *PostgresSchemaGenerator) writePermissionsComment(b *strings.Builder, doc *DocType) {
	b.WriteString("-- Permissions:\n")
	for _, perm := range doc.Permissions {
		if perm.Read {
			fmt.Fprintf(b, "--   Role %s has read access\n", perm.Role)
		}
	}
}

// fieldTypeMap maps field types to their skip status
var skipFieldTypes = map[string]bool{
	"Section Break": true,
	"Column Break":  true,
	"HTML":          true,
	"Button":        true,
	"Image":         true,
	"Fold":          true,
	"Tab Break":     true,
	"Heading":       true,
	"Table":         true,
}

func shouldSkipField(fieldtype string) bool {
	return skipFieldTypes[fieldtype]
}

func mapToPostgresType(field Field) string {
	switch field.Fieldtype {
	case "Link", "Select":
		return "VARCHAR(255)"
	case "Data":
		return getDataFieldType(field)
	case "Check":
		return "BOOLEAN"
	case "Int":
		return "INTEGER"
	case "Float", "Currency", "Percent":
		return "NUMERIC(18,6)"
	case "Date":
		return "DATE"
	case "Datetime":
		return "TIMESTAMP WITH TIME ZONE"
	case "Time":
		return "TIME"
	case "Text", "Small Text", "Long Text", "Markdown", "Code", "JSON", "HTML":
		return "TEXT"
	case "Attach", "Attach Image":
		return "TEXT"
	case "Color":
		return "VARCHAR(7)"
	case "Barcode", "Geolocation":
		return "TEXT"
	case "Duration":
		return "INTERVAL"
	default:
		return "TEXT"
	}
}

func getDataFieldType(field Field) string {
	if field.Options != nil && strings.Contains(*field.Options, "length:") {
		if parts := strings.Split(*field.Options, ":"); len(parts) > 1 {
			length := strings.TrimSpace(parts[1])
			return fmt.Sprintf("VARCHAR(%s)", length)
		}
	}
	return "VARCHAR(255)"
}

func getDefaultValue(field Field, pgType string) string {
	if field.Default == nil || *field.Default == "null" {
		return ""
	}
	
	val := *field.Default
	
	switch {
	case strings.HasPrefix(pgType, "VARCHAR"), pgType == "TEXT":
		return fmt.Sprintf(" DEFAULT '%s'", escapeString(val))
	case pgType == "BOOLEAN":
		return getBooleanDefault(val)
	case pgType == "NUMERIC(18,6)", pgType == "INTEGER":
		return fmt.Sprintf(" DEFAULT %s", val)
	case pgType == "DATE":
		return getDateDefault(val)
	case pgType == "TIMESTAMP WITH TIME ZONE":
		return getTimestampDefault(val)
	default:
		return fmt.Sprintf(" DEFAULT '%s'", escapeString(val))
	}
}

func getBooleanDefault(val string) string {
	if val == "1" || strings.EqualFold(val, "true") {
		return " DEFAULT TRUE"
	}
	return " DEFAULT FALSE"
}

func getDateDefault(val string) string {
	if strings.EqualFold(val, "Today") {
		return " DEFAULT CURRENT_DATE"
	}
	return fmt.Sprintf(" DEFAULT '%s'", val)
}

func getTimestampDefault(val string) string {
	if strings.EqualFold(val, "Now") {
		return " DEFAULT CURRENT_TIMESTAMP"
	}
	return fmt.Sprintf(" DEFAULT '%s'", val)
}

func escapeString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func quoteIdentifier(name string) string {
	if containsSpace(name) || isReservedKeyword(name) {
		return fmt.Sprintf(`"%s"`, name)
	}
	return name
}

func containsSpace(s string) bool {
	return strings.ContainsAny(s, " \t\n")
}

// reservedKeywords contains PostgreSQL reserved keywords
var reservedKeywords = map[string]bool{
	"user":       true,
	"group":      true,
	"table":      true,
	"select":     true,
	"insert":     true,
	"update":     true,
	"delete":     true,
	"where":      true,
	"order":      true,
	"primary":    true,
	"key":        true,
	"check":      true,
	"references": true,
}

func isReservedKeyword(name string) bool {
	return reservedKeywords[strings.ToLower(name)]
}
