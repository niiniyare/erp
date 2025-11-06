package parse

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/niiniyare/erp/pkg/schema"
	"github.com/stretchr/testify/suite"
)

type ParserTestSuite struct {
	suite.Suite
	ctx context.Context
}

func TestParserSuite(t *testing.T) {
	suite.Run(t, new(ParserTestSuite))
}

func (s *ParserTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *ParserTestSuite) TestParser_Parse_Basic() {
	jsonData := []byte(`{
		"type": "form",
		"id": "test-form",
		"title": "Test Form",
		"fields": [
			{
				"name": "email",
				"type": "email",
				"label": "Email",
				"required": true
			}
		]
	}`)

	parser := NewParser()
	schema, err := parser.Parse(s.ctx, jsonData)

	s.Require().NoError(err)
	s.Require().Equal("test-form", schema.ID)
	s.Require().Equal("form", string(schema.Type))
	s.Require().Equal("Test Form", schema.Title)
	s.Require().Len(schema.Fields, 1)
}

func (s *ParserTestSuite) TestParser_Parse_WithDefaults() {
	jsonData := []byte(`{
		"type": "form",
		"id": "test-form",
		"title": "Test Form",
		"fields": [
			{
				"name": "notes",
				"type": "textarea",
				"label": "Notes"
			}
		]
	}`)

	// Without defaults
	parser := NewParser()
	schema, err := parser.Parse(s.ctx, jsonData)
	s.Require().NoError(err)

	// Textarea should not have rows set
	s.Require().Nil(schema.Fields[0].Config)

	// With defaults
	parserWithDefaults := NewParser(WithDefaults())
	schema2, err := parserWithDefaults.Parse(s.ctx, jsonData)
	s.Require().NoError(err)

	// Textarea should have default rows
	s.Require().NotNil(schema2.Fields[0].Config)
	s.Require().Equal(4, schema2.Fields[0].Config["rows"])
}

func (s *ParserTestSuite) TestParser_Parse_EmptyData() {
	parser := NewParser()
	_, err := parser.Parse(s.ctx, []byte{})

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "empty schema data")
}

func (s *ParserTestSuite) TestParser_Parse_InvalidJSON() {
	parser := NewParser()
	_, err := parser.Parse(s.ctx, []byte(`{invalid json`))

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "malformed JSON")
}

func (s *ParserTestSuite) TestParser_Parse_MissingRequiredFields() {
	tests := []struct {
		name     string
		jsonData string
		errMsg   string
	}{
		{
			name:     "missing type",
			jsonData: `{"id": "test", "title": "Test"}`,
			errMsg:   "missing or invalid 'type' field",
		},
		{
			name:     "missing id for form",
			jsonData: `{"type": "form", "title": "Test"}`,
			errMsg:   "missing or invalid 'id' field",
		},
		{
			name:     "missing title for form",
			jsonData: `{"type": "form", "id": "test"}`,
			errMsg:   "missing or invalid 'title' field",
		},
	}

	parser := NewParser()
	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, err := parser.Parse(s.ctx, []byte(tt.jsonData))
			s.Require().Error(err)
			s.Require().Contains(err.Error(), tt.errMsg)
		})
	}
}

func (s *ParserTestSuite) TestParser_Parse_PageType() {
	jsonData := []byte(`{
		"id": "transaction-page",
		"type": "page",
		"title": "Transaction List"
	}`)

	parser := NewParser()
	schema, err := parser.Parse(s.ctx, jsonData)

	s.Require().NoError(err)
	s.Require().Equal("page", string(schema.Type))
	s.Require().Equal("Transaction List", schema.Title)
	s.Require().Equal("transaction-page", schema.ID)
}

func (s *ParserTestSuite) TestParser_Parse_MaxFieldsLimit() {
	// Create schema with too many fields
	fields := make([]map[string]any, 501)
	for i := 0; i < 501; i++ {
		fields[i] = map[string]any{
			"name":  "field" + string(rune(i)),
			"type":  "text",
			"label": "Field",
		}
	}

	schemaMap := map[string]any{
		"type":   "form",
		"id":     "test",
		"title":  "Test",
		"fields": fields,
	}

	jsonData, _ := json.Marshal(schemaMap)

	parser := NewParser(WithMaxFields(500))
	_, err := parser.Parse(s.ctx, jsonData)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "too many fields")
}

func (s *ParserTestSuite) TestParser_Parse_DuplicateFieldNames() {
	jsonData := []byte(`{
		"type": "form",
		"id": "test-form",
		"title": "Test Form",
		"fields": [
			{"name": "email", "type": "email", "label": "Email"},
			{"name": "email", "type": "text", "label": "Email 2"}
		]
	}`)

	parser := NewParser() // Strict validation by default
	_, err := parser.Parse(s.ctx, jsonData)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "field name already exists")
}

func (s *ParserTestSuite) TestParser_ParseString() {
	jsonStr := `{
		"type": "form",
		"id": "test-form",
		"title": "Test Form"
	}`

	parser := NewParser()
	schema, err := parser.ParseString(s.ctx, jsonStr)

	s.Require().NoError(err)
	s.Require().Equal("test-form", schema.ID)
}

func (s *ParserTestSuite) TestParser_ParseMap() {
	data := map[string]any{
		"type":  "form",
		"id":    "test-form",
		"title": "Test Form",
		"fields": []map[string]any{
			{
				"name":  "email",
				"type":  "email",
				"label": "Email",
			},
		},
	}

	parser := NewParser()
	schema, err := parser.ParseMap(s.ctx, data)

	s.Require().NoError(err)
	s.Require().Equal("test-form", schema.ID)
	s.Require().Len(schema.Fields, 1)
}

func (s *ParserTestSuite) TestParser_Serialize() {
	schema := &schema.Schema{
		Type:  "form",
		ID:    "test-form",
		Title: "Test Form",
		Fields: []schema.Field{
			{
				Name:     "email",
				Type:     schema.FieldEmail,
				Label:    "Email",
				Required: true,
			},
		},
	}

	parser := NewParser()
	data, err := parser.Serialize(schema)

	s.Require().NoError(err)
	s.Require().Contains(string(data), `"id": "test-form"`)
	s.Require().Contains(string(data), `"email"`)

	// Should be able to parse back
	schema2, err := parser.Parse(s.ctx, data)
	s.Require().NoError(err)
	s.Require().Equal(schema.ID, schema2.ID)
}

func (s *ParserTestSuite) TestParser_SerializeCompact() {
	schema := &schema.Schema{
		Type:  "form",
		ID:    "test-form",
		Title: "Test Form",
	}

	parser := NewParser()
	data, err := parser.SerializeCompact(schema)

	s.Require().NoError(err)
	// Compact should have no indentation
	s.Require().NotContains(string(data), "\n  ")
	s.Require().Contains(string(data), `"id":"test-form"`)
}

func (s *ParserTestSuite) TestParser_Clone() {
	original := &schema.Schema{
		Type:  "form",
		ID:    "test-form",
		Title: "Test Form",
		Fields: []schema.Field{
			{
				Name:  "email",
				Type:  schema.FieldEmail,
				Label: "Email",
			},
		},
	}

	parser := NewParser()
	cloned, err := parser.Clone(s.ctx, original)

	s.Require().NoError(err)
	s.Require().Equal(original.ID, cloned.ID)
	s.Require().Equal(original.Title, cloned.Title)
	s.Require().Len(cloned.Fields, 1)

	// Modify clone should not affect original
	cloned.Title = "Modified Title"
	s.Require().NotEqual(original.Title, cloned.Title)
}

func (s *ParserTestSuite) TestParser_Merge() {
	base := &schema.Schema{
		Type:  "form",
		ID:    "base-form",
		Title: "Base Form",
		Fields: []schema.Field{
			{Name: "name", Type: schema.FieldText, Label: "Name"},
		},
	}

	overlay := &schema.Schema{
		Type:  "form",
		ID:    "overlay-form",
		Title: "Overlay Form",
		Fields: []schema.Field{
			{Name: "email", Type: schema.FieldEmail, Label: "Email"},
		},
	}

	parser := NewParser()
	merged, err := parser.Merge(s.ctx, base, overlay)

	s.Require().NoError(err)
	s.Require().Equal("Overlay Form", merged.Title) // Overlay wins
	s.Require().Len(merged.Fields, 2)               // Both fields present
	s.Require().Equal("name", merged.Fields[0].Name)
	s.Require().Equal("email", merged.Fields[1].Name)
}

func (s *ParserTestSuite) TestParser_ValidateQuick() {
	validJSON := []byte(`{
		"type": "form",
		"id": "test",
		"title": "Test"
	}`)

	invalidJSON := []byte(`{
		"type": "form"
	}`)

	parser := NewParser()

	err := parser.ValidateQuick(validJSON)
	s.Require().NoError(err)

	err = parser.ValidateQuick(invalidJSON)
	s.Require().Error(err)
}

func (s *ParserTestSuite) TestParser_GetSchemaInfo() {
	jsonData := []byte(`{
		"type": "form",
		"id": "test-form",
		"title": "Test Form",
		"version": "1.2.0",
		"fields": [
			{"name": "field1", "type": "text", "label": "Field 1"},
			{"name": "field2", "type": "email", "label": "Field 2"}
		]
	}`)

	parser := NewParser()
	info, err := parser.GetSchemaInfo(jsonData)

	s.Require().NoError(err)
	s.Require().Equal("test-form", info.ID)
	s.Require().Equal("form", info.Type)
	s.Require().Equal("Test Form", info.Title)
	s.Require().Equal("1.2.0", info.Version)
	s.Require().Equal(2, info.FieldCount)
}

func (s *ParserTestSuite) TestParser_FieldDefaults() {
	tests := []struct {
		name      string
		fieldType schema.FieldType
		checkFunc func(*ParserTestSuite, *schema.Field)
	}{
		{
			name:      "textarea_rows",
			fieldType: schema.FieldTextarea,
			checkFunc: func(s *ParserTestSuite, f *schema.Field) {
				s.Require().Equal(4, f.Config["rows"])
			},
		},
		{
			name:      "number_step",
			fieldType: schema.FieldNumber,
			checkFunc: func(s *ParserTestSuite, f *schema.Field) {
				s.Require().NotNil(f.Validation)
				s.Require().NotNil(f.Validation.Step)
				s.Require().Equal(1.0, *f.Validation.Step)
			},
		},
		{
			name:      "currency_step",
			fieldType: schema.FieldCurrency,
			checkFunc: func(s *ParserTestSuite, f *schema.Field) {
				s.Require().NotNil(f.Validation)
				s.Require().NotNil(f.Validation.Step)
				s.Require().Equal(0.01, *f.Validation.Step)
			},
		},
		{
			name:      "date_format",
			fieldType: schema.FieldDate,
			checkFunc: func(s *ParserTestSuite, f *schema.Field) {
				s.Require().Equal("YYYY-MM-DD", f.Config["format"])
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			jsonData := []byte(`{
				"type": "form",
				"id": "test",
				"title": "Test",
				"fields": [
					{
						"name": "testfield",
						"type": "` + string(tt.fieldType) + `",
						"label": "Test Field"
					}
				]
			}`)

			parser := NewParser(WithDefaults())
			schema, err := parser.Parse(s.ctx, jsonData)

			s.Require().NoError(err)
			s.Require().Len(schema.Fields, 1)
			tt.checkFunc(s, &schema.Fields[0])
		})
	}
}

func (s *ParserTestSuite) TestParser_ActionDefaults() {
	jsonData := []byte(`{
		"type": "form",
		"id": "test",
		"title": "Test",
		"actions": [
			{"id": "submit", "type": "submit", "text": "Submit"},
			{"id": "reset", "type": "reset", "text": "Reset"},
			{"id": "custom", "type": "button", "text": "Custom"}
		]
	}`)

	parser := NewParser(WithDefaults())
	schema, err := parser.Parse(s.ctx, jsonData)

	s.Require().NoError(err)
	s.Require().Len(schema.Actions, 3)

	// Submit should be primary
	s.Require().Equal("primary", schema.Actions[0].Variant)
	s.Require().Equal("Submit", schema.Actions[0].Text)

	// Reset should be secondary
	s.Require().Equal("secondary", schema.Actions[1].Variant)
	s.Require().Equal("Reset", schema.Actions[1].Text)

	// Custom should be default
	s.Require().Equal("default", schema.Actions[2].Variant)
	s.Require().Equal("Custom", schema.Actions[2].Text)
}

func (s *ParserTestSuite) TestParser_ParseError() {
	err := NewParseError("test-schema", "something went wrong")

	s.Require().Contains(err.Error(), "test-schema")
	s.Require().Contains(err.Error(), "something went wrong")
	s.Require().True(IsParseError(err))
}

func (s *ParserTestSuite) TestParser_Options() {
	s.Run("WithDefaults", func() {
		parser := NewParser(WithDefaults())
		s.Require().True(parser.applyDefaults)
	})

	s.Run("WithStrictValidation", func() {
		parser := NewParser(WithStrictValidation(false))
		s.Require().False(parser.strictValidation)
	})

	s.Run("WithMaxFields", func() {
		parser := NewParser(WithMaxFields(100))
		s.Require().Equal(100, parser.maxFieldCount)
	})

	s.Run("Multiple options", func() {
		parser := NewParser(
			WithDefaults(),
			WithStrictValidation(false),
			WithMaxFields(200),
		)
		s.Require().True(parser.applyDefaults)
		s.Require().False(parser.strictValidation)
		s.Require().Equal(200, parser.maxFieldCount)
	})
}

// Benchmark tests - these remain as regular test functions
func BenchmarkParser_Parse(b *testing.B) {
	jsonData := []byte(`{
		"type": "form",
		"id": "benchmark-form",
		"title": "Benchmark Form",
		"fields": [
			{"name": "field1", "type": "text", "label": "Field 1"},
			{"name": "field2", "type": "email", "label": "Field 2"},
			{"name": "field3", "type": "number", "label": "Field 3"}
		]
	}`)

	parser := NewParser()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(context.Background(), jsonData)
	}
}

func BenchmarkParser_ParseWithDefaults(b *testing.B) {
	jsonData := []byte(`{
		"type": "form",
		"id": "benchmark-form",
		"title": "Benchmark Form",
		"fields": [
			{"name": "field1", "type": "text", "label": "Field 1"},
			{"name": "field2", "type": "email", "label": "Field 2"},
			{"name": "field3", "type": "number", "label": "Field 3"}
		]
	}`)

	parser := NewParser(WithDefaults())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(context.Background(), jsonData)
	}
}
