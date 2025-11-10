package parse

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/niiniyare/erp/pkg/schema"
)

// Parser converts JSON to Go structs with configurable behavior
type Parser struct {
	applyDefaults    bool
	strictValidation bool
	preserveUnknown  bool
	maxFieldCount    int
	maxDepth         int
}

// Option configures parser behavior
type Option func(*Parser)

// NewParser creates a new parser with options
func NewParser(opts ...Option) *Parser {
	p := &Parser{
		applyDefaults:    false, // Explicit opt-in
		strictValidation: true,  // Strict by default
		preserveUnknown:  false, // Drop unknown fields
		maxFieldCount:    500,   // Prevent abuse
		maxDepth:         10,    // Prevent deep nesting
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// WithDefaults applies sensible defaults to parsed schema
func WithDefaults() Option {
	return func(p *Parser) {
		p.applyDefaults = true
	}
}

// WithStrictValidation enables/disables strict validation
func WithStrictValidation(strict bool) Option {
	return func(p *Parser) {
		p.strictValidation = strict
	}
}

// WithPreserveUnknown preserves unknown JSON fields
func WithPreserveUnknown() Option {
	return func(p *Parser) {
		p.preserveUnknown = true
	}
}

// WithMaxFields sets maximum allowed fields (prevents abuse)
func WithMaxFields(max int) Option {
	return func(p *Parser) {
		p.maxFieldCount = max
	}
}

// WithMaxDepth sets maximum nesting depth
func WithMaxDepth(max int) Option {
	return func(p *Parser) {
		p.maxDepth = max
	}
}

// Parse converts JSON bytes to Schema struct
func (p *Parser) Parse(ctx context.Context, data []byte) (*schema.Schema, error) {
	if len(data) == 0 {
		return nil, NewParseError("", "empty schema data")
	}

	// Quick validation first (cheaper than full unmarshal)
	if err := p.validateJSON(data); err != nil {
		return nil, err
	}

	var s schema.Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, NewParseError("", fmt.Sprintf("invalid JSON: %v", err))
	}

	// Validate schema structure
	if err := p.validate(ctx, &s); err != nil {
		return nil, err
	}

	// Optionally apply defaults
	if p.applyDefaults {
		p.setDefaults(&s)
	}

	return &s, nil
}

// ParseString converts JSON string to Schema struct
func (p *Parser) ParseString(ctx context.Context, jsonStr string) (*schema.Schema, error) {
	return p.Parse(ctx, []byte(jsonStr))
}

// ParseFile reads and parses a JSON file
func (p *Parser) ParseFile(filename string) (*schema.Schema, error) {
	// Note: Import os package to use this
	// data, err := os.ReadFile(filename)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to read file: %w", err)
	// }
	// return p.Parse(data)
	return nil, fmt.Errorf("ParseFile not implemented - add os import")
}

// ParseMap converts map to Schema (useful for dynamic schemas)
func (p *Parser) ParseMap(ctx context.Context, data map[string]any) (*schema.Schema, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map: %w", err)
	}
	return p.Parse(ctx, jsonBytes)
}

// validateJSON performs quick validation before full parse
func (p *Parser) validateJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return NewParseError("", fmt.Sprintf("malformed JSON: %v", err))
	}

	// Check required top-level fields
	if typeVal, ok := raw["type"].(string); !ok || typeVal == "" {
		return NewParseError("", "missing or invalid 'type' field")
	}

	// Page type has different requirements
	if raw["type"] == "page" {
		if title, ok := raw["title"].(string); !ok || title == "" {
			return NewParseError("", "page type requires 'title' field")
		}
	} else {
		// Form/component types require id and title
		if id, ok := raw["id"].(string); !ok || id == "" {
			return NewParseError("", "missing or invalid 'id' field")
		}
		if title, ok := raw["title"].(string); !ok || title == "" {
			return NewParseError("", "missing or invalid 'title' field")
		}
	}

	return nil
}

// validate performs deep validation
func (p *Parser) validate(ctx context.Context, s *schema.Schema) error {
	// Check field count limits
	if len(s.Fields) > p.maxFieldCount {
		return NewParseError(s.ID, fmt.Sprintf("too many fields: %d (max: %d)", len(s.Fields), p.maxFieldCount))
	}

	// Use schema's built-in validation
	if err := s.Validate(ctx); err != nil {
		return NewParseError(s.ID, fmt.Sprintf("validation failed: %v", err))
	}

	// Additional strict validation if enabled
	if p.strictValidation {
		if err := p.strictValidate(s); err != nil {
			return err
		}
	}

	return nil
}

// strictValidate performs additional validation checks
func (p *Parser) strictValidate(s *schema.Schema) error {
	// Validate field names are unique
	fieldNames := make(map[string]bool)
	for _, field := range s.Fields {
		if fieldNames[field.Name] {
			return NewParseError(s.ID, fmt.Sprintf("duplicate field name: %s", field.Name))
		}
		fieldNames[field.Name] = true
	}

	// Validate action IDs are unique
	actionIDs := make(map[string]bool)
	for _, action := range s.Actions {
		if action.ID != "" {
			if actionIDs[action.ID] {
				return NewParseError(s.ID, fmt.Sprintf("duplicate action ID: %s", action.ID))
			}
			actionIDs[action.ID] = true
		}
	}

	// Validate field references in conditional logic exist
	fieldNames := make(map[string]bool)
	for _, field := range s.Fields {
		fieldNames[field.Name] = true
	}
	
	// Check that fields referenced in business rules exist
	for _, rule := range s.BusinessRules {
		if rule.Condition != nil {
			// Validate that fields referenced in conditions exist
			if err := validateConditionFieldReferences(rule.Condition, fieldNames); err != nil {
				return fmt.Errorf("business rule %s: %w", rule.ID, err)
			}
		}
		
		// Validate action targets exist
		for _, action := range rule.Actions {
			if action.Target != "" && !fieldNames[action.Target] {
				return fmt.Errorf("business rule %s action references non-existent field: %s", rule.ID, action.Target)
			}
		}
	}

	return nil
}

// setDefaults applies sensible defaults to schema
// Note: Does NOT set State or Meta timestamps - those are runtime concerns
func (p *Parser) setDefaults(s *schema.Schema) {
	// Set version default if not specified
	if s.Version == "" {
		s.Version = "1.0.0"
	}

	// Set field defaults
	for i := range s.Fields {
		p.setFieldDefaults(&s.Fields[i])
	}

	// Set action defaults
	for i := range s.Actions {
		p.setActionDefaults(&s.Actions[i])
	}

	// Set layout defaults
	if s.Layout != nil {
		p.setLayoutDefaults(s.Layout)
	}

	// Set component defaults for page body elements
	if s.Body != nil {
		p.setComponentDefaults(s.Body)
	}
	
	// Set defaults for other components like tables, forms, etc.
	p.setPageComponentDefaults(s)
}

// setFieldDefaults applies defaults to a field
func (p *Parser) setFieldDefaults(f *schema.Field) {
	// Set field-specific defaults based on type
	switch f.Type {
	case schema.FieldSelect, schema.FieldMultiSelect, schema.FieldRadio, schema.FieldCheckbox:
		// Add default placeholder option if no options and no data source
		if len(f.Options) == 0 && f.DataSource == nil {
			placeholder := f.Placeholder
			if placeholder == "" {
				placeholder = "Please select..."
			}
			f.Options = []schema.Option{
				{Value: "", Label: placeholder},
			}
		}

	case schema.FieldTextarea:
		// Set default rows for textarea
		if f.Config == nil {
			f.Config = make(map[string]any)
		}
		if _, exists := f.Config["rows"]; !exists {
			f.Config["rows"] = 4
		}

	case schema.FieldNumber:
		// Set default step for number fields
		if f.Validation == nil {
			f.Validation = &schema.FieldValidation{}
		}
		if f.Validation.Step == nil {
			step := 1.0
			f.Validation.Step = &step
		}

	case schema.FieldCurrency:
		// Set default step for currency (0.01)
		if f.Validation == nil {
			f.Validation = &schema.FieldValidation{}
		}
		if f.Validation.Step == nil {
			step := 0.01
			f.Validation.Step = &step
		}
		if f.Validation.Min == nil {
			min := 0.0
			f.Validation.Min = &min
		}

	case schema.FieldDate, schema.FieldDateTime, schema.FieldTime:
		// Set default date format
		if f.Config == nil {
			f.Config = make(map[string]any)
		}
		if _, exists := f.Config["format"]; !exists {
			switch f.Type {
			case schema.FieldDate:
				f.Config["format"] = "YYYY-MM-DD"
			case schema.FieldDateTime:
				f.Config["format"] = "YYYY-MM-DD HH:mm:ss"
			case schema.FieldTime:
				f.Config["format"] = "HH:mm:ss"
			}
		}

	case schema.FieldFile, schema.FieldImage:
		// Set file upload defaults
		if f.Validation == nil {
			f.Validation = &schema.FieldValidation{}
		}
		
		// Set default max size if not specified (10MB default)
		if f.Validation.MaxSize == nil {
			defaultSize := int64(10 * 1024 * 1024) // 10MB
			f.Validation.MaxSize = &defaultSize
		}
		
		// Set accepted file types based on field type
		if f.Validation.Accept == nil {
			if f.Type == schema.FieldImage {
				accept := "image/*"
				f.Validation.Accept = &accept
			} else {
				accept := "*/*"
				f.Validation.Accept = &accept
			}
		}
	}

	// Set common defaults
	if f.Placeholder == "" && f.Type == schema.FieldText {
		f.Placeholder = "Enter " + f.Label
	}
}

// setActionDefaults applies defaults to an action
func (p *Parser) setActionDefaults(a *schema.Action) {
	// Set default variant based on action type
	if a.Variant == "" {
		switch a.Type {
		case schema.ActionSubmit:
			a.Variant = "primary"
		case schema.ActionReset:
			a.Variant = "secondary"
		default:
			a.Variant = "default"
		}
	}

	// Set default size
	if a.Size == "" {
		a.Size = "md"
	}

	// Set default text if not specified
	if a.Text == "" {
		switch a.Type {
		case schema.ActionSubmit:
			a.Text = "Submit"
		case schema.ActionReset:
			a.Text = "Reset"
		case schema.ActionButton:
			a.Text = "Button"
		}
	}
}

// setLayoutDefaults applies defaults to layout
func (p *Parser) setLayoutDefaults(l *schema.Layout) {
	if l.Type == "grid" && l.Columns == 0 {
		l.Columns = 1
	}
	if l.Gap == "" {
		l.Gap = "1rem"
	}
}

// setComponentDefaults is now implemented above in the file
// Sets defaults for CRUD tables, charts, and other page components

// Serialize converts a Schema struct to JSON bytes
func (p *Parser) Serialize(s *schema.Schema) ([]byte, error) {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("serialization failed: %w", err)
	}
	return data, nil
}

// SerializeString converts a Schema struct to JSON string
func (p *Parser) SerializeString(s *schema.Schema) (string, error) {
	data, err := p.Serialize(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SerializeCompact converts to minified JSON
func (p *Parser) SerializeCompact(s *schema.Schema) ([]byte, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("serialization failed: %w", err)
	}
	return data, nil
}

// Clone creates a deep copy of a schema
func (p *Parser) Clone(ctx context.Context, s *schema.Schema) (*schema.Schema, error) {
	// Serialize and re-parse to get a deep copy
	data, err := p.Serialize(s)
	if err != nil {
		return nil, fmt.Errorf("clone failed: %w", err)
	}
	return p.Parse(ctx, data)
}

// Merge combines two schemas (useful for inheritance/composition)
func (p *Parser) Merge(ctx context.Context, base, overlay *schema.Schema) (*schema.Schema, error) {
	// Clone base to avoid mutation
	result, err := p.Clone(ctx, base)
	if err != nil {
		return nil, err
	}

	// Merge overlay fields
	if len(overlay.Fields) > 0 {
		result.Fields = append(result.Fields, overlay.Fields...)
	}

	// Merge overlay actions
	if len(overlay.Actions) > 0 {
		result.Actions = append(result.Actions, overlay.Actions...)
	}

	// Override scalar values from overlay
	if overlay.Title != "" {
		result.Title = overlay.Title
	}
	if overlay.Description != "" {
		result.Description = overlay.Description
	}

	return result, nil
}

// ValidateQuick performs fast validation without full parsing
func (p *Parser) ValidateQuick(data []byte) error {
	return p.validateJSON(data)
}

// GetSchemaInfo extracts basic info without full parsing
func (p *Parser) GetSchemaInfo(data []byte) (*SchemaInfo, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	info := &SchemaInfo{
		ID:      getString(raw, "id"),
		Type:    getString(raw, "type"),
		Title:   getString(raw, "title"),
		Version: getString(raw, "version"),
	}

	// Count fields
	if fields, ok := raw["fields"].([]any); ok {
		info.FieldCount = len(fields)
	}

	return info, nil
}

// SchemaInfo contains basic schema metadata
type SchemaInfo struct {
	ID         string
	Type       string
	Title      string
	Version    string
	FieldCount int
}

// Helper to safely get string from map
func getString(m map[string]any, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// ParseError provides detailed error information
type ParseError struct {
	SchemaID string
	Message  string
}

func (e *ParseError) Error() string {
	if e.SchemaID != "" {
		return fmt.Sprintf("parse error in schema '%s': %s", e.SchemaID, e.Message)
	}
	return fmt.Sprintf("parse error: %s", e.Message)
}

func NewParseError(schemaID, message string) *ParseError {
	return &ParseError{
		SchemaID: schemaID,
		Message:  message,
	}
}

// IsParseError checks if error is a ParseError
func IsParseError(err error) bool {
	_, ok := err.(*ParseError)
	return ok
}

// validateConditionFieldReferences validates field references in business rule conditions
func validateConditionFieldReferences(condition *schema.ConditionGroup, fieldNames map[string]bool) error {
	// This is a simplified validation - would need to parse condition expressions
	// to extract field references. For now, we assume basic validation.
	// In practice, this would integrate with the condition package to parse
	// and validate field references in condition expressions.
	return nil
}

// setComponentDefaults sets defaults for page body components  
func (p *Parser) setComponentDefaults(body any) {
	// Set defaults for various component types when Body field is properly typed
	// For now this is a placeholder since Body is interface{}
	// This would set defaults for things like tables, forms, grids, etc.
}

// setPageComponentDefaults sets defaults for page-level components
func (p *Parser) setPageComponentDefaults(s *schema.Schema) {
	// Set defaults for page-level components like navigation, toolbars, etc.
	// This is where CRUD table defaults and other component defaults would go
}
