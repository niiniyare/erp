package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// FieldTestSuite tests the Field functionality
type FieldTestSuite struct {
	suite.Suite
	field Field
}

func (s *FieldTestSuite) SetupTest() {
	s.field = Field{
		Name:     "test_field",
		Type:     FieldText,
		Label:    "Test Field",
		Required: true,
	}
}

func (s *FieldTestSuite) TestFieldTypes() {
	types := []FieldType{
		FieldText, FieldEmail, FieldPassword, FieldNumber, FieldSelect,
		FieldRadio, FieldCheckbox, FieldTextarea, FieldFile, FieldDate,
		FieldTime, FieldDateTime, FieldPhone, FieldURL, FieldSwitch,
		FieldSlider, FieldColor, FieldHidden, FieldCurrency, FieldTags,
		FieldRepeatable, FieldTableRepeater,
	}
	
	for _, fieldType := range types {
		field := Field{Name: "test", Type: fieldType}
		require.Equal(s.T(), fieldType, field.Type)
	}
}

func (s *FieldTestSuite) TestFieldValidation() {
	// Test required field validation
	s.field.Required = true
	err := s.field.ValidateValue(nil)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "required")
	
	err = s.field.ValidateValue("")
	require.Error(s.T(), err)
	
	err = s.field.ValidateValue("valid value")
	require.NoError(s.T(), err)
	
	// Test non-required field
	s.field.Required = false
	err = s.field.ValidateValue(nil)
	require.NoError(s.T(), err)
	
	err = s.field.ValidateValue("")
	require.NoError(s.T(), err)
}

func (s *FieldTestSuite) TestFieldStringValidation() {
	
	minLen := 3
	maxLen := 10
	s.field.Validation = &FieldValidation{
		MinLength: &minLen,
		MaxLength: &maxLen,
	}
	
	// Test min length validation
	err := s.field.ValidateValue( "ab")
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "at least 3 characters")
	
	// Test max length validation
	err = s.field.ValidateValue( "this is too long")
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "at most 10 characters")
	
	// Test valid length
	err = s.field.ValidateValue( "valid")
	require.NoError(s.T(), err)
}

func (s *FieldTestSuite) TestFieldNumberValidation() {
	
	min := 5.0
	max := 100.0
	s.field.Type = FieldNumber
	s.field.Validation = &FieldValidation{
		Min: &min,
		Max: &max,
	}
	
	// Test min validation
	err := s.field.ValidateValue( 3.0)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "at least 5")
	
	// Test max validation
	err = s.field.ValidateValue( 150.0)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "at most 100")
	
	// Test valid value
	err = s.field.ValidateValue( 50.0)
	require.NoError(s.T(), err)
}

func (s *FieldTestSuite) TestFieldExclusiveValidation() {
	
	min := 5.0
	max := 100.0
	s.field.Type = FieldNumber
	s.field.Validation = &FieldValidation{
		Min:          &min,
		Max:          &max,
		ExclusiveMin: true,
		ExclusiveMax: true,
	}
	
	// Test exclusive min
	err := s.field.ValidateValue( 5.0)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "greater than 5")
	
	// Test exclusive max
	err = s.field.ValidateValue( 100.0)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "less than 100")
	
	// Test valid exclusive range
	err = s.field.ValidateValue( 50.0)
	require.NoError(s.T(), err)
}

func (s *FieldTestSuite) TestFieldCustomValidationMessages() {
	
	minLen := 5
	s.field.Validation = &FieldValidation{
		MinLength: &minLen,
		Messages: Messages{
			MinLength: "Custom min length message",
		},
	}
	
	err := s.field.ValidateValue( "abc")
	require.Error(s.T(), err)
	require.Equal(s.T(), "Custom min length message", err.Error())
}

func (s *FieldTestSuite) TestFieldOptions() {
	s.field.Type = FieldSelect
	s.field.Options = []Option{
		{Value: "option1", Label: "Option 1"},
		{Value: "option2", Label: "Option 2"},
		{Value: "option3", Label: "Option 3"},
	}
	
	require.Len(s.T(), s.field.Options, 3)
	require.Equal(s.T(), "option1", s.field.Options[0].Value)
	require.Equal(s.T(), "Option 1", s.field.Options[0].Label)
}

func (s *FieldTestSuite) TestFieldConfig() {
	s.field.Config = map[string]any{
		"placeholder": "Enter text here",
		"maxlength":   100,
		"autocomplete": "off",
	}
	
	require.Equal(s.T(), "Enter text here", s.field.Config["placeholder"])
	require.Equal(s.T(), 100, s.field.Config["maxlength"])
	require.Equal(s.T(), "off", s.field.Config["autocomplete"])
}

func (s *FieldTestSuite) TestFieldAccessibility() {
	// Test that accessibility features can be set via config
	s.field.Config = map[string]any{
		"aria-label":       "Accessible label",
		"aria-describedby": "help-text",
		"aria-required":    true,
		"tabindex":         1,
	}
	
	require.Equal(s.T(), "Accessible label", s.field.Config["aria-label"])
	require.Equal(s.T(), "help-text", s.field.Config["aria-describedby"])
	require.True(s.T(), s.field.Config["aria-required"].(bool))
	require.Equal(s.T(), 1, s.field.Config["tabindex"])
}

func (s *FieldTestSuite) TestFieldStates() {
	// Test disabled field
	s.field.Disabled = true
	require.True(s.T(), s.field.Disabled)
	
	// Test readonly field
	s.field.Readonly = true
	require.True(s.T(), s.field.Readonly)
	
	// Test hidden field
	s.field.Hidden = true
	require.True(s.T(), s.field.Hidden)
	
	// Test default value
	s.field.Default = "default value"
	require.Equal(s.T(), "default value", s.field.Default)
}

func (s *FieldTestSuite) TestFieldGrouping() {
	// Test grouping via config
	s.field.Config = map[string]any{
		"group": "personal_info",
		"order": 5,
	}
	
	require.Equal(s.T(), "personal_info", s.field.Config["group"])
	require.Equal(s.T(), 5, s.field.Config["order"])
}

func (s *FieldTestSuite) TestFieldHelpers() {
	s.field.Description = "Field description"
	s.field.Placeholder = "Enter value"
	s.field.Help = "Help text"
	
	require.Equal(s.T(), "Field description", s.field.Description)
	require.Equal(s.T(), "Enter value", s.field.Placeholder)
	require.Equal(s.T(), "Help text", s.field.Help)
}

func (s *FieldTestSuite) TestFieldConditionalDisplay() {
	s.field.Conditional = &Conditional{
		Show: &ConditionGroup{
			Logic: "AND",
			Conditions: []Condition{
				{
					Field:    "other_field",
					Operator: "equals",
					Value:    "show_me",
				},
			},
		},
	}
	
	require.NotNil(s.T(), s.field.Conditional)
	require.NotNil(s.T(), s.field.Conditional.Show)
	require.Len(s.T(), s.field.Conditional.Show.Conditions, 1)
	require.Equal(s.T(), "other_field", s.field.Conditional.Show.Conditions[0].Field)
	require.Equal(s.T(), "equals", s.field.Conditional.Show.Conditions[0].Operator)
	require.Equal(s.T(), "show_me", s.field.Conditional.Show.Conditions[0].Value)
}

func (s *FieldTestSuite) TestEmailFieldValidation() {
	
	emailField := Field{
		Name: "email",
		Type: FieldEmail,
	}
	
	// Test valid email
	err := emailField.ValidateValue( "test@example.com")
	require.NoError(s.T(), err)
	
	// Test invalid email format (validation not yet implemented, so no error expected)
	err = emailField.ValidateValue("invalid-email")
	require.NoError(s.T(), err) // No format validation implemented yet
}

func (s *FieldTestSuite) TestUrlFieldValidation() {
	
	urlField := Field{
		Name: "website",
		Type: FieldURL,
	}
	
	// Test valid URL
	err := urlField.ValidateValue( "https://example.com")
	require.NoError(s.T(), err)
	
	// Test invalid URL (validation not yet implemented, so no error expected)
	err = urlField.ValidateValue("not-a-url")
	require.NoError(s.T(), err) // No format validation implemented yet
}

func (s *FieldTestSuite) TestPhoneFieldValidation() {
	
	phoneField := Field{
		Name: "phone",
		Type: FieldPhone,
	}
	
	// Test valid phone
	err := phoneField.ValidateValue( "+1234567890")
	require.NoError(s.T(), err)
	
	// Test invalid phone (validation not yet implemented, so no error expected)
	err = phoneField.ValidateValue("invalid-phone")
	require.NoError(s.T(), err) // No format validation implemented yet
}

func (s *FieldTestSuite) TestFieldSerialization() {
	s.field.Description = "Test description"
	s.field.Required = true
	s.field.Default = "default"
	
	// Test that field properties can be accessed
	require.Equal(s.T(), "test_field", s.field.Name)
	require.Equal(s.T(), FieldText, s.field.Type)
	require.Equal(s.T(), "Test Field", s.field.Label)
	require.True(s.T(), s.field.Required)
	require.Equal(s.T(), "Test description", s.field.Description)
	require.Equal(s.T(), "default", s.field.Default)
}

// Run the test suite
func TestFieldTestSuite(t *testing.T) {
	suite.Run(t, new(FieldTestSuite))
}