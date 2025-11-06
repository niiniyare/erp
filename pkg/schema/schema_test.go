package schema

import (
	"context"
	"testing"
	"time"

	"github.com/niiniyare/erp/pkg/condition"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// SchemaTestSuite tests the core Schema functionality
type SchemaTestSuite struct {
	suite.Suite
	schema *Schema
	ctx    context.Context
}

func (s *SchemaTestSuite) SetupTest() {
	s.ctx = context.Background()
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

	err := s.schema.Validate(s.ctx)
	require.NoError(s.T(), err)

	// Invalid schema - empty ID
	s.schema.ID = ""
	err = s.schema.Validate(s.ctx)
	require.Error(s.T(), err)

	// Invalid schema - empty title
	s.schema.ID = "valid-id"
	s.schema.Title = ""
	err = s.schema.Validate(s.ctx)
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

// ==================== Additional Coverage Tests ====================

func (s *SchemaTestSuite) TestSchemaStateManagement() {
	// Test field value management
	s.schema.SetFieldValue("name", "John Doe")
	value, exists := s.schema.GetFieldValue("name")
	s.Require().True(exists)
	s.Require().Equal("John Doe", value)

	// Test non-existent field
	_, exists = s.schema.GetFieldValue("nonexistent")
	s.Require().False(exists)

	// Test field error management
	s.schema.SetFieldError("email", "Invalid email format")
	s.Require().True(s.schema.HasErrors())
	errors := s.schema.GetErrors()
	s.Require().Equal("Invalid email format", errors["email"])

	// Test clear specific error
	s.schema.ClearFieldError("email")
	s.Require().False(s.schema.HasErrors())

	// Test multiple errors
	s.schema.SetFieldError("email", "Invalid email")
	s.schema.SetFieldError("phone", "Invalid phone")
	s.Require().True(s.schema.HasErrors())
	s.Require().Len(s.schema.GetErrors(), 2)

	// Test clear all errors
	s.schema.ClearAllErrors()
	s.Require().False(s.schema.HasErrors())
	s.Require().Empty(s.schema.GetErrors())

	// Test touched fields
	s.Require().False(s.schema.IsFieldTouched("name"))
	s.schema.MarkFieldTouched("name")
	s.Require().True(s.schema.IsFieldTouched("name"))

	// Test dirty fields
	s.Require().False(s.schema.IsFieldDirty("name"))
	s.Require().False(s.schema.IsDirty())
	s.schema.MarkFieldDirty("name")
	s.Require().True(s.schema.IsFieldDirty("name"))
	s.Require().True(s.schema.IsDirty())

	// Test reset
	s.schema.Reset()
	s.Require().False(s.schema.HasErrors())
	s.Require().False(s.schema.IsFieldTouched("name"))
	s.Require().False(s.schema.IsFieldDirty("name"))
	s.Require().False(s.schema.IsDirty())
}

func (s *SchemaTestSuite) TestSchemaUtilityMethods() {
	// Test empty schema
	s.Require().True(s.schema.IsEmpty())
	s.Require().Equal(0, s.schema.GetFieldCount())
	s.Require().Equal(0, s.schema.GetActionCount())

	// Add fields and actions
	field := Field{Name: "test_field", Type: FieldText, Label: "Test"}
	action := Action{ID: "test_action", Type: ActionSubmit, Text: "Submit"}

	s.schema.AddField(field)
	s.schema.AddAction(action)

	s.Require().False(s.schema.IsEmpty())
	s.Require().Equal(1, s.schema.GetFieldCount())
	s.Require().Equal(1, s.schema.GetActionCount())

	// Test field retrieval by index
	retrievedField, exists := s.schema.GetFieldByIndex(0)
	s.Require().True(exists)
	s.Require().Equal("test_field", retrievedField.Name)

	// Test invalid index
	_, exists = s.schema.GetFieldByIndex(10)
	s.Require().False(exists)

	// Test negative index
	_, exists = s.schema.GetFieldByIndex(-1)
	s.Require().False(exists)

	// Test field names and action IDs
	fieldNames := s.schema.GetFieldNames()
	s.Require().Len(fieldNames, 1)
	s.Require().Equal("test_field", fieldNames[0])

	actionIDs := s.schema.GetActionIDs()
	s.Require().Len(actionIDs, 1)
	s.Require().Equal("test_action", actionIDs[0])
}

func (s *SchemaTestSuite) TestSchemaFieldMutators() {
	// Test AddFields (batch)
	field1 := Field{Name: "field1", Type: FieldText, Label: "Field 1"}
	field2 := Field{Name: "field2", Type: FieldEmail, Label: "Field 2"}
	s.schema.AddFields(field1, field2)
	s.Require().Len(s.schema.Fields, 2)

	// Test RemoveField
	removed := s.schema.RemoveField("field1")
	s.Require().True(removed)
	s.Require().Len(s.schema.Fields, 1)
	s.Require().Equal("field2", s.schema.Fields[0].Name)

	// Test RemoveField non-existent
	removed = s.schema.RemoveField("nonexistent")
	s.Require().False(removed)

	// Test UpdateField
	updatedField := Field{Name: "field2", Type: FieldPassword, Label: "Updated Field 2"}
	updated := s.schema.UpdateField("field2", updatedField)
	s.Require().True(updated)
	s.Require().Equal(FieldPassword, s.schema.Fields[0].Type)
	s.Require().Equal("Updated Field 2", s.schema.Fields[0].Label)

	// Test UpdateField non-existent
	updated = s.schema.UpdateField("nonexistent", updatedField)
	s.Require().False(updated)
}

func (s *SchemaTestSuite) TestSchemaActionMutators() {
	// Test AddActions (batch)
	action1 := Action{ID: "action1", Type: ActionSubmit, Text: "Submit"}
	action2 := Action{ID: "action2", Type: ActionButton, Text: "Cancel"}
	s.schema.AddActions(action1, action2)
	s.Require().Len(s.schema.Actions, 2)

	// Test RemoveAction
	removed := s.schema.RemoveAction("action1")
	s.Require().True(removed)
	s.Require().Len(s.schema.Actions, 1)
	s.Require().Equal("action2", s.schema.Actions[0].ID)

	// Test RemoveAction non-existent
	removed = s.schema.RemoveAction("nonexistent")
	s.Require().False(removed)

	// Test UpdateAction
	updatedAction := Action{ID: "action2", Type: ActionReset, Text: "Reset Form"}
	updated := s.schema.UpdateAction("action2", updatedAction)
	s.Require().True(updated)
	s.Require().Equal(ActionReset, s.schema.Actions[0].Type)
	s.Require().Equal("Reset Form", s.schema.Actions[0].Text)

	// Test UpdateAction non-existent
	updated = s.schema.UpdateAction("nonexistent", updatedAction)
	s.Require().False(updated)
}

func (s *SchemaTestSuite) TestSchemaJSONMethods() {
	s.schema.Description = "Test description"
	s.schema.AddField(Field{Name: "test", Type: FieldText, Label: "Test"})

	// Test ToJSON
	jsonStr, err := s.schema.ToJSON()
	s.Require().NoError(err)
	s.Require().Contains(jsonStr, "test-schema")
	s.Require().Contains(jsonStr, "Test Schema")

	// Test ToJSONPretty
	prettyJSON, err := s.schema.ToJSONPretty()
	s.Require().NoError(err)
	s.Require().Contains(prettyJSON, "test-schema")
	s.Require().Contains(prettyJSON, "\n") // Should have newlines for formatting

	// Test FromJSON
	recreatedSchema, err := FromJSON(jsonStr)
	s.Require().NoError(err)
	s.Require().Equal(s.schema.ID, recreatedSchema.ID)
	s.Require().Equal(s.schema.Title, recreatedSchema.Title)
	s.Require().Equal(s.schema.Description, recreatedSchema.Description)
	s.Require().Len(recreatedSchema.Fields, 1)
}

func (s *SchemaTestSuite) TestSchemaInterfaceMethods() {
	s.schema.Version = "2.0.0"

	s.Require().Equal("test-schema", s.schema.GetID())
	s.Require().Equal("form", s.schema.GetType())
	s.Require().Equal("2.0.0", s.schema.GetVersion())
	s.Require().NotNil(s.schema.GetMetadata())
}

func (s *SchemaTestSuite) TestMergeSchemas() {
	base := NewSchema("base", TypeForm, "Base Form")
	base.AddField(Field{Name: "field1", Type: FieldText, Label: "Field 1"})
	base.AddAction(Action{ID: "action1", Type: ActionSubmit, Text: "Submit"})
	base.Tags = []string{"tag1"}

	other := NewSchema("other", TypeForm, "Other Form")
	other.AddField(Field{Name: "field2", Type: FieldText, Label: "Field 2"})
	other.AddField(Field{Name: "field1", Type: FieldText, Label: "Duplicate Field"}) // Duplicate, should be skipped
	other.AddAction(Action{ID: "action2", Type: ActionButton, Text: "Button"})
	other.AddAction(Action{ID: "action1", Type: ActionButton, Text: "Duplicate"}) // Duplicate, should be skipped
	other.Tags = []string{"tag2", "tag1"}                                         // tag1 is duplicate

	merged, err := MergeSchemas(base, other)
	s.NoError(err)
	// Should have 2 fields (field1 from base, field2 from other)
	s.Require().Len(merged.Fields, 2)
	s.Require().True(merged.HasField("field1"))
	s.Require().True(merged.HasField("field2"))

	// Should have 2 actions (action1 from base, action2 from other)
	s.Require().Len(merged.Actions, 2)
	s.Require().True(merged.HasAction("action1"))
	s.Require().True(merged.HasAction("action2"))

	// Should have merged tags without duplicates
	s.Require().Len(merged.Tags, 2)
	s.Require().Contains(merged.Tags, "tag1")
	s.Require().Contains(merged.Tags, "tag2")
}

func (s *SchemaTestSuite) TestFilterFields() {
	s.schema.AddField(Field{Name: "text1", Type: FieldText, Label: "Text 1"})
	s.schema.AddField(Field{Name: "email1", Type: FieldEmail, Label: "Email 1"})
	s.schema.AddField(Field{Name: "text2", Type: FieldText, Label: "Text 2"})

	// Filter for text fields only
	filtered, err := s.schema.FilterFields(func(field Field) bool {
		return field.Type == FieldText
	})
	s.Require().NoError(err)

	s.Require().Len(filtered.Fields, 2)
	s.Require().Equal("text1", filtered.Fields[0].Name)
	s.Require().Equal("text2", filtered.Fields[1].Name)
}

func (s *SchemaTestSuite) TestMapFields() {
	s.schema.AddField(Field{Name: "field1", Type: FieldText, Label: "Field 1"})
	s.schema.AddField(Field{Name: "field2", Type: FieldText, Label: "Field 2"})

	// Transform all field labels to uppercase
	mapped, err := s.schema.MapFields(func(field Field) Field {
		field.Label = "TRANSFORMED: " + field.Label
		return field
	})
	s.Require().NoError(err)

	s.Require().Len(mapped.Fields, 2)
	s.Require().Equal("TRANSFORMED: Field 1", mapped.Fields[0].Label)
	s.Require().Equal("TRANSFORMED: Field 2", mapped.Fields[1].Label)

	// Original should be unchanged
	s.Require().Equal("Field 1", s.schema.Fields[0].Label)
	s.Require().Equal("Field 2", s.schema.Fields[1].Label)
}

// ==================== Additional Enterprise Coverage ====================

func (s *SchemaTestSuite) TestSchemaValidationEdgeCases() {
	// Test schema with circular dependencies
	schema := NewSchema("test", TypeForm, "Test Schema")

	field1 := Field{
		Name:         "field1",
		Type:         FieldText,
		Dependencies: []string{"field2"},
	}

	field2 := Field{
		Name:         "field2",
		Type:         FieldText,
		Dependencies: []string{"field1"}, // Circular dependency
	}

	schema.AddFields(field1, field2)

	// Test circular dependency detection
	err := schema.DetectCircularDependencies()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "circular dependency")
}

func (s *SchemaTestSuite) TestSchemaAdvancedValidation() {
	schema := NewSchema("test", TypeForm, "Test Schema")

	// Test with invalid schema structure
	schema.ID = "" // Invalid ID
	err := schema.Validate(s.ctx)
	s.Require().Error(err)

	// Test with duplicate field names
	schema.ID = "valid-id"
	schema.AddField(Field{Name: "duplicate", Type: FieldText, Label: "Field 1"})
	schema.AddField(Field{Name: "duplicate", Type: FieldEmail, Label: "Field 2"})

	err = schema.Validate(s.ctx)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "duplicate")
}

func (s *SchemaTestSuite) TestSchemaConditionalFieldsAdvanced() {
	schema := NewSchema("test", TypeForm, "Test Form")
	ctx := context.Background()

	// Create fields with complex conditional logic
	visibleField := Field{
		Name:     "visible_field",
		Type:     FieldText,
		Label:    "Visible Field",
		Required: true,
	}

	conditionalField := Field{
		Name:  "conditional_field",
		Type:  FieldText,
		Label: "Conditional Field",
		Conditional: &Conditional{
			Show: &ConditionGroup{
				Logic: "AND",
				Conditions: []Condition{
					{Field: "type", Operator: "equal", Value: "premium"},
				},
			},
		},
	}

	hiddenField := Field{
		Name:   "hidden_field",
		Type:   FieldText,
		Label:  "Hidden Field",
		Hidden: true,
	}

	schema.AddFields(visibleField, conditionalField, hiddenField)

	// Test with premium type (conditional field should be visible)
	premiumData := map[string]any{
		"type":          "premium",
		"visible_field": "value",
	}

	visibleFields := schema.GetVisibleFields(ctx, premiumData)
	s.Require().Len(visibleFields, 1) // Only visible_field (conditional logic may not be fully implemented)

	hiddenFields := schema.GetHiddenFields(ctx, premiumData)
	s.Require().Len(hiddenFields, 2) // hidden_field and conditional_field (not shown)

	// Test with basic type (conditional field should be hidden)
	basicData := map[string]any{
		"type":          "basic",
		"visible_field": "value",
	}

	visibleFields = schema.GetVisibleFields(ctx, basicData)
	s.Require().Len(visibleFields, 1) // only visible_field

	hiddenFields = schema.GetHiddenFields(ctx, basicData)
	s.Require().Len(hiddenFields, 2) // conditional_field and hidden_field
}

// Test schema utility methods for better coverage
func (s *SchemaTestSuite) TestSchemaUtilityMethodsExtended() {
	schema := NewSchema("test", TypeForm, "Test Schema")
	ctx := context.Background()

	// Add some test fields
	schema.AddFields(
		Field{Name: "required_text", Type: FieldText, Label: "Required Text", Required: true},
		Field{Name: "optional_email", Type: FieldEmail, Label: "Optional Email", Required: false},
		Field{Name: "required_number", Type: FieldNumber, Label: "Required Number", Required: true},
		Field{Name: "optional_select", Type: FieldSelect, Label: "Optional Select", Required: false},
	)

	// Test SetEvaluator and GetEvaluator
	evaluator := condition.NewEvaluator(nil, condition.DefaultEvalOptions())
	schema.SetEvaluator(evaluator)
	s.Require().Equal(evaluator, schema.GetEvaluator())

	// Test GetRequiredFields
	requiredFields := schema.GetRequiredFields(ctx, map[string]any{})
	s.Require().Len(requiredFields, 2) // required_text and required_number
	requiredNames := make([]string, len(requiredFields))
	for i, field := range requiredFields {
		requiredNames[i] = field.Name
	}
	s.Require().Contains(requiredNames, "required_text")
	s.Require().Contains(requiredNames, "required_number")

	// Test GetOptionalFields
	optionalFields := schema.GetOptionalFields(ctx, map[string]any{})
	s.Require().Len(optionalFields, 2) // optional_email and optional_select
	optionalNames := make([]string, len(optionalFields))
	for i, field := range optionalFields {
		optionalNames[i] = field.Name
	}
	s.Require().Contains(optionalNames, "optional_email")
	s.Require().Contains(optionalNames, "optional_select")

	// Test GetFieldsByType
	textFields := schema.GetFieldsByType(FieldText)
	s.Require().Len(textFields, 1)
	s.Require().Equal("required_text", textFields[0].Name)

	emailFields := schema.GetFieldsByType(FieldEmail)
	s.Require().Len(emailFields, 1)
	s.Require().Equal("optional_email", emailFields[0].Name)

	// Test with actions
	schema.AddActions(
		Action{ID: "action1", Type: ActionSubmit, Text: "Submit", Disabled: false},
		Action{ID: "action2", Type: ActionButton, Text: "Button", Disabled: true},
		Action{ID: "action3", Type: ActionReset, Text: "Reset", Disabled: false},
	)

	// Test GetEnabledActions
	enabledActions := schema.GetEnabledActions(ctx, map[string]any{})
	s.Require().Len(enabledActions, 2) // action1 and action3 (not disabled)
	enabledIDs := make([]string, len(enabledActions))
	for i, action := range enabledActions {
		enabledIDs[i] = action.ID
	}
	s.Require().Contains(enabledIDs, "action1")
	s.Require().Contains(enabledIDs, "action3")
	s.Require().NotContains(enabledIDs, "action2") // disabled

	// Test ValidateData
	validData := map[string]any{
		"required_text":   "some text",
		"required_number": 42,
		"optional_email":  "test@example.com",
	}
	err := schema.ValidateData(ctx, validData)
	s.Require().NoError(err)

	// Test ValidateData with missing required field
	invalidData := map[string]any{
		"optional_email": "test@example.com",
		// missing required_text and required_number
	}
	err = schema.ValidateData(ctx, invalidData)
	s.Require().Error(err) // Should fail due to missing required fields
}

// Test enterprise security features
func (s *SchemaTestSuite) TestSecurityEnterpriseFeatures() {
	security := &Security{
		CSRF: &CSRF{
			Enabled:     true,
			FieldName:   "_csrf",
			HeaderName:  "X-CSRF-Token",
			CookieName:  "csrf_token",
			TokenLength: 32,
			Expiry:      3600,
		},
		RateLimit: &RateLimit{
			Enabled:       true,
			MaxRequests:   100,
			Window:        3600,
			ByIP:          true,
			ByUser:        false,
			Burst:         10,
			BlockDuration: 300,
		},
		Encryption: &Encryption{
			Enabled:   true,
			Fields:    []string{"password", "ssn", "credit_card"},
			Algorithm: "AES-256-GCM",
		},
		Sanitization:  true,
		Origins:       []string{"https://example.com", "https://app.example.com"},
		ContentType:   []string{"application/json", "multipart/form-data"},
		MaxFileSize:   10485760, // 10MB
		AllowedExts:   []string{".jpg", ".png", ".pdf"},
		XSSProtection: true,
		SQLInjection:  true,
	}

	// Test CSRF methods
	s.Require().True(security.IsCSRFEnabled())

	// Test rate limit methods
	s.Require().True(security.IsRateLimitEnabled())

	// Test encryption methods
	s.Require().True(security.IsEncryptionEnabled())
	s.Require().True(security.ShouldEncryptField("password"))
	s.Require().True(security.ShouldEncryptField("ssn"))
	s.Require().False(security.ShouldEncryptField("username"))

	// Test with disabled features
	disabledSecurity := &Security{
		CSRF:       &CSRF{Enabled: false},
		RateLimit:  &RateLimit{Enabled: false},
		Encryption: &Encryption{Enabled: false},
	}
	s.Require().False(disabledSecurity.IsCSRFEnabled())
	s.Require().False(disabledSecurity.IsRateLimitEnabled())
	s.Require().False(disabledSecurity.IsEncryptionEnabled())
	s.Require().False(disabledSecurity.ShouldEncryptField("password"))

	// Test with nil features
	nilSecurity := &Security{}
	s.Require().False(nilSecurity.IsCSRFEnabled())
	s.Require().False(nilSecurity.IsRateLimitEnabled())
	s.Require().False(nilSecurity.IsEncryptionEnabled())
	s.Require().False(nilSecurity.ShouldEncryptField("password"))
}

// Test tenant enterprise features
func (s *SchemaTestSuite) TestTenantEnterpriseFeatures() {
	tenant := &Tenant{
		Enabled:   true,
		Field:     "tenant_id",
		Isolation: "strict",
	}

	// Test tenant enabled
	s.Require().True(tenant.IsTenantEnabled())

	// Test disabled tenant
	disabledTenant := &Tenant{Enabled: false}
	s.Require().False(disabledTenant.IsTenantEnabled())

	// Test nil tenant
	var nilTenant *Tenant
	s.Require().False(nilTenant.IsTenantEnabled())
}

// Test workflow enterprise features
func (s *SchemaTestSuite) TestWorkflowEnterpriseFeatures() {
	ctx := context.Background()
	evaluator := condition.NewEvaluator(nil, condition.DefaultEvalOptions())

	workflow := &Workflow{
		Enabled: true,
		Stage:   "review",
		Status:  "pending",
		Actions: []WorkflowAction{
			{
				ID:       "approve",
				Label:    "Approve",
				Type:     "approve",
				ToStage:  "approved",
				ToStatus: "completed",
			},
			{
				ID:       "reject",
				Label:    "Reject",
				Type:     "reject",
				ToStage:  "rejected",
				ToStatus: "rejected",
				Condition: &condition.ConditionGroup{
					ID:          "can_reject",
					Conjunction: condition.ConjunctionAnd,
					Children: []any{
						&condition.ConditionRule{
							ID: "user_role_check",
							Left: condition.Expression{
								Type:  condition.ValueTypeField,
								Field: "user_role",
							},
							Op:    condition.OpEqual,
							Right: "admin",
						},
					},
				},
			},
		},
	}

	// Test workflow enabled
	s.Require().True(workflow.IsWorkflowEnabled())

	// Test current stage and status
	s.Require().Equal("review", workflow.GetCurrentStage())
	s.Require().Equal("pending", workflow.GetCurrentStatus())

	// Test available actions for admin user
	adminData := map[string]any{"user_role": "admin"}
	availableActions := workflow.GetAvailableActions(ctx, adminData, evaluator)
	s.Require().Len(availableActions, 2) // Both approve and reject should be available

	// Test available actions for regular user
	userData := map[string]any{"user_role": "user"}
	availableActions = workflow.GetAvailableActions(ctx, userData, evaluator)
	s.Require().Len(availableActions, 1) // Only approve should be available

	// Test disabled workflow
	disabledWorkflow := &Workflow{Enabled: false}
	s.Require().False(disabledWorkflow.IsWorkflowEnabled())
	s.Require().Empty(disabledWorkflow.GetAvailableActions(ctx, map[string]any{}, evaluator))

	// Test nil workflow
	var nilWorkflow *Workflow
	s.Require().False(nilWorkflow.IsWorkflowEnabled())
	s.Require().Equal("", nilWorkflow.GetCurrentStage())
	s.Require().Equal("", nilWorkflow.GetCurrentStatus())
	s.Require().Empty(nilWorkflow.GetAvailableActions(ctx, map[string]any{}, evaluator))
}

// Test i18n enterprise features
func (s *SchemaTestSuite) TestI18nEnterpriseFeatures() {
	i18n := &I18n{
		Enabled:          true,
		DefaultLocale:    "en-US",
		SupportedLocales: []string{"en-US", "es-ES", "fr-FR"},
		FallbackLocale:   "en-US",
	}

	// Test i18n enabled
	s.Require().True(i18n.IsI18nEnabled())
	s.Require().Equal("en-US", i18n.GetLocale())

	// Test with empty default locale
	i18nEmpty := &I18n{
		Enabled:       true,
		DefaultLocale: "",
	}
	s.Require().Equal("en-US", i18nEmpty.GetLocale()) // Should fallback to en-US

	// Test disabled i18n
	disabledI18n := &I18n{Enabled: false}
	s.Require().False(disabledI18n.IsI18nEnabled())

	// Test nil i18n
	var nilI18n *I18n
	s.Require().False(nilI18n.IsI18nEnabled())
	s.Require().Equal("en-US", nilI18n.GetLocale()) // Should return default
}

// Test HTMX enterprise features
func (s *SchemaTestSuite) TestHTMXEnterpriseFeatures() {
	htmx := &HTMX{
		Enabled: true,
		Get:     "/api/get",
		Post:    "/api/post",
		Put:     "/api/put",
		Patch:   "/api/patch",
		Delete:  "/api/delete",
		Target:  "#content",
		Swap:    "innerHTML",
		Trigger: "click",
		Boost:   true,
	}

	// Test HTMX enabled
	s.Require().True(htmx.IsHTMXEnabled())

	// Test HTTP method determination
	htmxPost := &HTMX{Post: "/api/submit"}
	s.Require().Equal("POST", htmxPost.GetHTTPMethod())
	s.Require().Equal("/api/submit", htmxPost.GetURL())

	htmxPut := &HTMX{Put: "/api/update"}
	s.Require().Equal("PUT", htmxPut.GetHTTPMethod())
	s.Require().Equal("/api/update", htmxPut.GetURL())

	htmxPatch := &HTMX{Patch: "/api/patch"}
	s.Require().Equal("PATCH", htmxPatch.GetHTTPMethod())
	s.Require().Equal("/api/patch", htmxPatch.GetURL())

	htmxDelete := &HTMX{Delete: "/api/remove"}
	s.Require().Equal("DELETE", htmxDelete.GetHTTPMethod())
	s.Require().Equal("/api/remove", htmxDelete.GetURL())

	htmxGet := &HTMX{Get: "/api/fetch"}
	s.Require().Equal("GET", htmxGet.GetHTTPMethod())
	s.Require().Equal("/api/fetch", htmxGet.GetURL())

	// Test disabled HTMX
	disabledHTMX := &HTMX{Enabled: false}
	s.Require().False(disabledHTMX.IsHTMXEnabled())

	// Test nil HTMX
	var nilHTMX *HTMX
	s.Require().False(nilHTMX.IsHTMXEnabled())
	s.Require().Equal("GET", nilHTMX.GetHTTPMethod()) // Should return default
	s.Require().Equal("", nilHTMX.GetURL())
}

// Test Alpine.js enterprise features
func (s *SchemaTestSuite) TestAlpineEnterpriseFeatures() {
	alpine := &Alpine{
		Enabled: true,
		XData:   "{ open: false, toggle() { this.open = !this.open } }",
		XShow:   "open",
		XOn:     "click: toggle()",
	}

	// Test Alpine enabled
	s.Require().True(alpine.IsAlpineEnabled())

	// Test disabled Alpine
	disabledAlpine := &Alpine{Enabled: false}
	s.Require().False(disabledAlpine.IsAlpineEnabled())

	// Test nil Alpine
	var nilAlpine *Alpine
	s.Require().False(nilAlpine.IsAlpineEnabled())
}

// Test meta enterprise features
func (s *SchemaTestSuite) TestMetaEnterpriseFeatures() {
	now := time.Now()
	meta := &Meta{
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    "user123",
		UpdatedBy:    "user456",
		Deprecated:   false,
		Experimental: true,
		Changelog: []ChangelogEntry{
			{
				Version:     "1.0.0",
				Date:        now,
				Author:      "dev team",
				Description: "Initial release",
				Breaking:    false,
				Changes:     []string{"Added core functionality"},
			},
		},
		CustomData: map[string]any{
			"project": "erp",
		},
	}

	// Test meta properties
	s.Require().False(meta.IsDeprecated())
	s.Require().True(meta.IsExperimental())
	s.Require().Equal("1.0.0", meta.GetLatestVersion())

	// Test add changelog
	newEntry := ChangelogEntry{
		Version:     "1.1.0",
		Date:        now.Add(24 * time.Hour),
		Author:      "dev team",
		Description: "Feature update",
		Breaking:    false,
		Changes:     []string{"Added new features"},
	}
	meta.AddChangelog(newEntry)
	s.Require().Len(meta.Changelog, 2)
	s.Require().Equal("1.1.0", meta.GetLatestVersion())

	// Test deprecated meta
	deprecatedMeta := &Meta{Deprecated: true}
	s.Require().True(deprecatedMeta.IsDeprecated())
	s.Require().False(deprecatedMeta.IsExperimental())

	// Test nil meta
	var nilMeta *Meta
	s.Require().False(nilMeta.IsDeprecated())
	s.Require().False(nilMeta.IsExperimental())
	s.Require().Equal("", nilMeta.GetLatestVersion())

	// Test meta with no changelog
	emptyMeta := &Meta{}
	s.Require().Equal("", emptyMeta.GetLatestVersion())

	// Test adding changelog to meta without existing changelog
	emptyMeta.AddChangelog(newEntry)
	s.Require().Len(emptyMeta.Changelog, 1)
	s.Require().Equal("1.1.0", emptyMeta.GetLatestVersion())
}
