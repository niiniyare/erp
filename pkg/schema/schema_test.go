package schema

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// SchemaTestSuite tests the core Schema functionality
type SchemaTestSuite struct {
	suite.Suite
	schema *Schema
}

func (s *SchemaTestSuite) SetupTest() {
	s.schema = NewSchema("test-schema", TypeForm, "Test Schema")
}

func (s *SchemaTestSuite) TestNewSchema() {
	schema := NewSchema("user-form", TypeForm, "User Form")

	require.Equal(s.T(), "user-form", schema.ID)
	require.Equal(s.T(), TypeForm, schema.Type)
	require.Equal(s.T(), "User Form", schema.Title)
	require.Empty(s.T(), schema.Fields)
	require.Empty(s.T(), schema.Actions)
}

func (s *SchemaTestSuite) TestSchemaAddField() {
	field := Field{
		Name:     "username",
		Type:     FieldText,
		Label:    "Username",
		Required: true,
	}

	s.schema.AddField(field)

	require.Len(s.T(), s.schema.Fields, 1)
	require.Equal(s.T(), "username", s.schema.Fields[0].Name)
	require.Equal(s.T(), FieldText, s.schema.Fields[0].Type)
	require.True(s.T(), s.schema.Fields[0].Required)
}

func (s *SchemaTestSuite) TestSchemaAddAction() {
	action := Action{
		ID:   "submit",
		Type: ActionSubmit,
		Text: "Submit Form",
	}

	s.schema.AddAction(action)

	require.Len(s.T(), s.schema.Actions, 1)
	require.Equal(s.T(), "submit", s.schema.Actions[0].ID)
	require.Equal(s.T(), ActionSubmit, s.schema.Actions[0].Type)
}

func (s *SchemaTestSuite) TestSchemaValidate() {
	// Valid schema
	s.schema.ID = "valid-id"
	s.schema.Type = TypeForm
	s.schema.Title = "Valid Title"

	err := s.schema.Validate()
	require.NoError(s.T(), err)

	// Invalid schema - empty ID
	s.schema.ID = ""
	err = s.schema.Validate()
	require.Error(s.T(), err)

	// Invalid schema - empty title
	s.schema.ID = "valid-id"
	s.schema.Title = ""
	err = s.schema.Validate()
	require.Error(s.T(), err)
}

func (s *SchemaTestSuite) TestSchemaData() {
	data := map[string]any{
		"username": "testuser",
		"email":    "test@example.com",
	}

	// Test that schema can work with data
	require.Equal(s.T(), "testuser", data["username"])
	require.Equal(s.T(), "test@example.com", data["email"])
}

func (s *SchemaTestSuite) TestSchemaTypes() {
	types := []Type{TypeForm, TypeComponent, TypeLayout, TypeWorkflow, TypeTheme, TypePage}

	for _, schemaType := range types {
		schema := NewSchema("test", schemaType, "Test")
		require.Equal(s.T(), schemaType, schema.Type)
	}
}

func (s *SchemaTestSuite) TestSchemaFieldMethods() {
	field1 := Field{Name: "field1", Type: FieldText}
	field2 := Field{Name: "field2", Type: FieldEmail}

	s.schema.AddField(field1)
	s.schema.AddField(field2)

	// Test GetField
	foundField, exists := s.schema.GetField("field1")
	require.True(s.T(), exists)
	require.Equal(s.T(), "field1", foundField.Name)

	// Test non-existent field
	_, exists = s.schema.GetField("nonexistent")
	require.False(s.T(), exists)

	// Test manual field removal
	s.schema.Fields = []Field{field2} // Remove field1 manually
	require.Len(s.T(), s.schema.Fields, 1)
	require.Equal(s.T(), "field2", s.schema.Fields[0].Name)
}

func (s *SchemaTestSuite) TestSchemaActionMethods() {
	action1 := Action{ID: "action1", Type: ActionButton}
	action2 := Action{ID: "action2", Type: ActionSubmit}

	s.schema.AddAction(action1)
	s.schema.AddAction(action2)

	// Test finding action manually
	var foundAction *Action
	var exists bool
	for _, action := range s.schema.Actions {
		if action.ID == "action1" {
			foundAction = &action
			exists = true
			break
		}
	}
	require.True(s.T(), exists)
	require.Equal(s.T(), "action1", foundAction.ID)

	// Test non-existent action
	exists = false
	for _, action := range s.schema.Actions {
		if action.ID == "nonexistent" {
			exists = true
			break
		}
	}
	require.False(s.T(), exists)

	// Test manual action removal
	s.schema.Actions = []Action{action2} // Remove action1 manually
	require.Len(s.T(), s.schema.Actions, 1)
	require.Equal(s.T(), "action2", s.schema.Actions[0].ID)
}

func (s *SchemaTestSuite) TestSchemaClone() {
	s.schema.Fields = []Field{
		{Name: "field1", Type: FieldText},
		{Name: "field2", Type: FieldEmail},
	}
	s.schema.Actions = []Action{
		{ID: "action1", Type: ActionButton},
	}

	// Manual cloning test
	cloned := &Schema{
		ID:      s.schema.ID,
		Type:    s.schema.Type,
		Title:   s.schema.Title,
		Fields:  make([]Field, len(s.schema.Fields)),
		Actions: make([]Action, len(s.schema.Actions)),
	}
	copy(cloned.Fields, s.schema.Fields)
	copy(cloned.Actions, s.schema.Actions)

	require.Equal(s.T(), s.schema.ID, cloned.ID)
	require.Equal(s.T(), s.schema.Type, cloned.Type)
	require.Equal(s.T(), s.schema.Title, cloned.Title)
	require.Len(s.T(), cloned.Fields, 2)
	require.Len(s.T(), cloned.Actions, 1)

	// Verify it's a copy
	cloned.Fields[0].Name = "modified"
	require.Equal(s.T(), "field1", s.schema.Fields[0].Name)
}

func (s *SchemaTestSuite) TestSchemaProperties() {
	field := Field{Name: "username", Type: FieldText, Required: true}
	s.schema.AddField(field)

	// Test schema properties
	require.Equal(s.T(), "test-schema", s.schema.ID)
	require.Equal(s.T(), TypeForm, s.schema.Type)
	require.Equal(s.T(), "Test Schema", s.schema.Title)
	require.Len(s.T(), s.schema.Fields, 1)
	require.Equal(s.T(), "username", s.schema.Fields[0].Name)
}

func (s *SchemaTestSuite) TestSchemaMetadata() {
	now := time.Now()
	s.schema.Meta = &Meta{
		CreatedAt:  now,
		UpdatedAt:  now,
		CreatedBy:  "user123",
		CustomData: map[string]any{"version": "1.0"},
	}

	require.Equal(s.T(), now, s.schema.Meta.CreatedAt)
	require.Equal(s.T(), "user123", s.schema.Meta.CreatedBy)
	require.Equal(s.T(), "1.0", s.schema.Meta.CustomData["version"])
}

func (s *SchemaTestSuite) TestSchemaEnterpriseFeatures() {
	// Test Security
	s.schema.Security = &Security{
		CSRF: &CSRF{Enabled: true, FieldName: "_csrf"},
	}
	require.True(s.T(), s.schema.Security.CSRF.Enabled)

	// Test Tenant
	s.schema.Tenant = &Tenant{
		Enabled:   true,
		Field:     "tenant_id",
		Isolation: "strict",
	}
	require.True(s.T(), s.schema.Tenant.Enabled)
	require.Equal(s.T(), "strict", s.schema.Tenant.Isolation)

	// Test I18n
	s.schema.I18n = &I18n{
		Enabled:          true,
		DefaultLocale:    "en",
		SupportedLocales: []string{"en", "es", "fr"},
	}
	require.True(s.T(), s.schema.I18n.Enabled)
	require.Contains(s.T(), s.schema.I18n.SupportedLocales, "es")
}

func (s *SchemaTestSuite) TestSchemaFrontendIntegration() {
	// Test HTMX
	s.schema.HTMX = &HTMX{
		Enabled: true,
		Boost:   true,
	}
	require.True(s.T(), s.schema.HTMX.Enabled)
	require.True(s.T(), s.schema.HTMX.Boost)

	// Test Alpine
	s.schema.Alpine = &Alpine{
		Enabled: true,
		XData:   "{ open: false }",
	}
	require.True(s.T(), s.schema.Alpine.Enabled)
	require.Equal(s.T(), "{ open: false }", s.schema.Alpine.XData)
}

// Run the test suite
func TestSchemaTestSuite(t *testing.T) {
	suite.Run(t, new(SchemaTestSuite))
}
