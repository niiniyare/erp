package parse

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/niiniyare/erp/pkg/schema"
)

// Parser converts JSON to Go structs with validation and defaults
type Parser struct{}

// NewParser creates a new parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse converts JSON bytes to Schema struct
func (p *Parser) Parse(data []byte) (*schema.Schema, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty schema data")
	}

	var s schema.Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Set defaults
	p.setDefaults(&s)

	// Validate structure
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("schema validation failed: %w", err)
	}

	return &s, nil
}

// ParseString converts JSON string to Schema struct
func (p *Parser) ParseString(jsonStr string) (*schema.Schema, error) {
	return p.Parse([]byte(jsonStr))
}

// setDefaults applies default values to schema
func (p *Parser) setDefaults(s *schema.Schema) {
	now := time.Now()

	// Set metadata defaults
	if s.Meta == nil {
		s.Meta = &schema.Meta{
			CreatedAt: now,
			UpdatedAt: now,
		}
	} else {
		if s.Meta.CreatedAt.IsZero() {
			s.Meta.CreatedAt = now
		}
		s.Meta.UpdatedAt = now
	}

	// Set state defaults
	if s.State == nil {
		s.State = &schema.State{
			Values:      make(map[string]any),
			Errors:      make(map[string]string),
			Touched:     make(map[string]bool),
			Dirty:       make(map[string]bool),
			Valid:       true,
			LastUpdated: now,
		}
	}

	// Set field defaults
	for i := range s.Fields {
		p.setFieldDefaults(&s.Fields[i])
	}

	// Set action defaults
	for i := range s.Actions {
		p.setActionDefaults(&s.Actions[i])
	}

	// Set version default
	if s.Version == "" {
		s.Version = "1.0.0"
	}
}

// setFieldDefaults applies defaults to a field
func (p *Parser) setFieldDefaults(f *schema.Field) {
	// Set default value if not set
	if f.Default == nil {
		f.Default = f.GetDefaultValue()
	}

	// Set field-specific defaults based on type
	switch f.Type {
	case schema.FieldSelect, schema.FieldMultiSelect, schema.FieldRadio:
		// Ensure options exist for selection fields
		if len(f.Options) == 0 && f.DataSource == nil {
			f.Options = []schema.Option{
				{Value: "", Label: "Please select..."},
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
	case schema.FieldNumber, schema.FieldCurrency:
		// Set default step for number fields
		if f.Validation == nil {
			f.Validation = &schema.FieldValidation{}
		}
		if f.Validation.Step == nil {
			step := 1.0
			if f.Type == schema.FieldCurrency {
				step = 0.01
			}
			f.Validation.Step = &step
		}
	}
}

// setActionDefaults applies defaults to an action
func (p *Parser) setActionDefaults(a *schema.Action) {
	// Set default variant
	if a.Variant == "" {
		switch a.Type {
		case schema.ActionSubmit:
			a.Variant = "primary"
		case schema.ActionReset:
			a.Variant = "secondary"
		default:
			a.Variant = "outline"
		}
	}

	// Set default size
	if a.Size == "" {
		a.Size = "md"
	}
}

// Serialize converts a Schema struct to JSON bytes
func (p *Parser) Serialize(s *schema.Schema) ([]byte, error) {
	// Update timestamp before serializing
	if s.Meta != nil {
		s.Meta.UpdatedAt = time.Now()
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
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

// ValidateJSON validates JSON structure without full parsing
func (p *Parser) ValidateJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Check required top-level fields
	required := []string{"id", "type", "title"}
	for _, field := range required {
		if _, exists := raw[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// Check field types
	if id, ok := raw["id"].(string); !ok || id == "" {
		return fmt.Errorf("id must be a non-empty string")
	}

	if schemaType, ok := raw["type"].(string); !ok || schemaType == "" {
		return fmt.Errorf("type must be a non-empty string")
	}

	if title, ok := raw["title"].(string); !ok || title == "" {
		return fmt.Errorf("title must be a non-empty string")
	}

	return nil
}