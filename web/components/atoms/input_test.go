package atoms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// InputTestSuite provides comprehensive tests for Input component
type InputTestSuite struct {
	suite.Suite
}

// SetupTest runs before each test method
func (suite *InputTestSuite) SetupTest() {
	ResetIDCounter()
}

// TestInputTestSuite runs the input test suite
func TestInputTestSuite(t *testing.T) {
	suite.Run(t, new(InputTestSuite))
}

// ============================================================================
// INPUT PROPS TESTS
// ============================================================================

func (suite *InputTestSuite) TestInputProps_DefaultValues() {
	props := InputProps{}
	
	// Test default values - check zero values
	suite.Equal(InputType(""), props.Type) // Zero value for InputType
	suite.Equal(Size(""), props.Size)      // Zero value for Size
	suite.False(props.InteractionProps.Disabled)  // Disabled is in InteractionProps
	suite.False(props.InteractionProps.Required)  // Required is in InteractionProps
}

func (suite *InputTestSuite) TestInputProps_BasicProperties() {
	props := InputProps{
		BaseProps: BaseProps{
			Name: "email",
		},
		Type: InputTypeEmail,
		PlaceholderProps: PlaceholderProps{
			Placeholder: "Enter your email",
		},
		Value: "user@example.com",
		InteractionProps: InteractionProps{
			Required: true,
		},
	}
	
	suite.Equal("email", props.BaseProps.Name)
	suite.Equal(InputTypeEmail, props.Type)
	suite.Equal("Enter your email", props.PlaceholderProps.Placeholder)
	suite.Equal("user@example.com", props.Value)
	suite.True(props.InteractionProps.Required)
}

func (suite *InputTestSuite) TestInputProps_ValidationState() {
	props := InputProps{
		ValidationProps: ValidationProps{
			State:     StateError,
			ErrorText: "Email is required",
		},
	}
	
	suite.Equal(StateError, props.ValidationProps.State)
	suite.Equal("Email is required", props.ValidationProps.ErrorText)
	suite.True(props.ValidationProps.State.IsError())
}

// ============================================================================
// INPUT TYPE TESTS
// ============================================================================

func (suite *InputTestSuite) TestInputType_AllTypes() {
	types := []InputType{
		InputTypeText,
		InputTypeEmail,
		InputTypePassword,
		InputTypeNumber,
		InputTypeSearch,
		InputTypeDate,
		InputTypeURL,
		InputTypeTel,
	}
	
	for _, inputType := range types {
		props := InputProps{Type: inputType}
		suite.Equal(inputType, props.Type)
	}
}

func (suite *InputTestSuite) TestInputType_StringConversion() {
	tests := []struct {
		inputType InputType
		expected  string
	}{
		{InputTypeText, "text"},
		{InputTypeEmail, "email"},
		{InputTypePassword, "password"},
		{InputTypeNumber, "number"},
		{InputTypeSearch, "search"},
		{InputTypeDate, "date"},
		{InputTypeURL, "url"},
		{InputTypeTel, "tel"},
	}
	
	for _, tt := range tests {
		suite.Equal(tt.expected, tt.inputType.String())
	}
}

func (suite *InputTestSuite) TestInputType_AriaRole() {
	tests := []struct {
		inputType    InputType
		expectedRole string
	}{
		{InputTypeSearch, "searchbox"},
		{InputTypeText, ""},
		{InputTypeEmail, ""},
		{InputTypePassword, ""},
	}
	
	for _, tt := range tests {
		suite.Equal(tt.expectedRole, tt.inputType.GetAriaRole())
	}
}

// ============================================================================
// INPUT SIZE TESTS
// ============================================================================

func (suite *InputTestSuite) TestInputSize_AllSizes() {
	sizes := []Size{
		SizeXS, // Use the general Size constants
		SizeSM,
		SizeMD,
		SizeLG,
		SizeXL,
	}
	
	for _, size := range sizes {
		props := InputProps{Size: size}
		suite.Equal(size, props.Size)
	}
}

func (suite *InputTestSuite) TestInputSize_StringConversion() {
	tests := []struct {
		size     Size
		expected string
	}{
		{SizeXS, "xs"}, // Use the general Size constants
		{SizeSM, "sm"},
		{SizeMD, "md"},
		{SizeLG, "lg"},
		{SizeXL, "xl"},
	}
	
	for _, tt := range tests {
		suite.Equal(tt.expected, tt.size.String())
	}
}

// ============================================================================
// INPUT VALIDATION TESTS
// ============================================================================

func (suite *InputTestSuite) TestInputProps_ValidationStates() {
	states := []ValidationState{
		StateDefault,
		StateSuccess,
		StateError,
		StateWarning,
		StateInfo,
	}
	
	for _, state := range states {
		props := InputProps{
			ValidationProps: ValidationProps{State: state},
		}
		suite.Equal(state, props.ValidationProps.State)
	}
}

func (suite *InputTestSuite) TestInputProps_ValidationMessages() {
	props := InputProps{
		ValidationProps: ValidationProps{
			State:       StateError,
			ErrorText:   "Field is required",
			SuccessText: "Looks good!",
			HelpText:    "Enter a valid value",
		},
	}
	
	msg, state := props.ValidationProps.GetFeedbackMessage()
	suite.Equal("Field is required", msg)
	suite.Equal(StateError, state)
	
	// Test success state
	props.ValidationProps.State = StateSuccess
	msg, state = props.ValidationProps.GetFeedbackMessage()
	suite.Equal("Looks good!", msg)
	suite.Equal(StateSuccess, state)
}

// ============================================================================
// INPUT COMPOSITION TESTS
// ============================================================================

func (suite *InputTestSuite) TestInputProps_BasePropsComposition() {
	props := InputProps{
		BaseProps: BaseProps{
			ID:    "email-input",
			Name:  "email", // Name is in BaseProps
			Class: "w-full",
		},
	}
	
	suite.Equal("email-input", props.BaseProps.ID)
	suite.Equal("email", props.BaseProps.Name)
	suite.Equal("w-full", props.BaseProps.Class)
}

func (suite *InputTestSuite) TestInputProps_InteractionPropsComposition() {
	props := InputProps{
		BaseProps: BaseProps{
			Name: "readonly-field", // Name is in BaseProps
		},
		InteractionProps: InteractionProps{
			Disabled: true,
			ReadOnly: false,
		},
	}
	
	suite.True(props.InteractionProps.Disabled)
	suite.False(props.InteractionProps.ReadOnly)
	suite.False(props.InteractionProps.IsInteractive())
}

func (suite *InputTestSuite) TestInputProps_AccessibilityPropsComposition() {
	props := InputProps{
		AccessibilityProps: AccessibilityProps{
			AriaLabel:       "Email address",
			AriaRequired:    true,
			AriaInvalid:     false,
			AriaDescribedBy: "email-help",
		},
		Type: InputTypeEmail,
	}
	
	suite.Equal("Email address", props.AccessibilityProps.AriaLabel)
	suite.True(props.AccessibilityProps.AriaRequired)
	suite.False(props.AccessibilityProps.AriaInvalid)
	suite.Equal("email-help", props.AccessibilityProps.AriaDescribedBy)
}

func (suite *InputTestSuite) TestInputProps_AlpineEventHandlers() {
	props := InputProps{
		AlpinEventHandlers: AlpinEventHandlers{
			OnChange: "validateEmail($event.target.value)",
			OnFocus:  "showHelp()",
			OnBlur:   "hideHelp()",
		},
		Type: InputTypeEmail,
	}
	
	suite.Equal("validateEmail($event.target.value)", props.AlpinEventHandlers.OnChange)
	suite.Equal("showHelp()", props.AlpinEventHandlers.OnFocus)
	suite.Equal("hideHelp()", props.AlpinEventHandlers.OnBlur)
}

// ============================================================================
// INPUT HTML ATTRIBUTES TESTS
// ============================================================================

func (suite *InputTestSuite) TestInputProps_HTMLAttributes() {
	props := InputProps{
		BaseProps: BaseProps{
			Name: "password",
		},
		Type: InputTypePassword,
		PlaceholderProps: PlaceholderProps{
			Placeholder: "Enter password",
		},
		MinLength:    8,
		MaxLength:    50,
		Pattern:      "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)[a-zA-Z\\d]{8,}$",
		AutoComplete: "new-password", // Note: AutoComplete not Autocomplete
	}
	
	suite.Equal("password", props.BaseProps.Name)
	suite.Equal(InputTypePassword, props.Type)
	suite.Equal("Enter password", props.PlaceholderProps.Placeholder)
	suite.Equal(8, props.MinLength)
	suite.Equal(50, props.MaxLength)
	suite.NotEmpty(props.Pattern)
	suite.Equal("new-password", props.AutoComplete)
}

func (suite *InputTestSuite) TestInputProps_NumericAttributes() {
	props := InputProps{
		Type: InputTypeNumber,
		Min:  "0",
		Max:  "100",
		Step: "0.1",
	}
	
	suite.Equal(InputTypeNumber, props.Type)
	suite.Equal("0", props.Min)
	suite.Equal("100", props.Max)
	suite.Equal("0.1", props.Step)
}

// ============================================================================
// INPUT EDGE CASES
// ============================================================================

func (suite *InputTestSuite) TestInputProps_EmptyValues() {
	props := InputProps{
		BaseProps: BaseProps{
			Name: "",
		},
		Value: "",
		PlaceholderProps: PlaceholderProps{
			Placeholder: "",
		},
	}
	
	suite.Empty(props.BaseProps.Name)
	suite.Empty(props.Value)
	suite.Empty(props.PlaceholderProps.Placeholder)
}

func (suite *InputTestSuite) TestInputProps_SearchInput() {
	props := InputProps{
		Type: InputTypeSearch,
		BaseProps: BaseProps{
			Name: "search",
		},
		PlaceholderProps: PlaceholderProps{
			Placeholder: "Search...",
		},
		AccessibilityProps: AccessibilityProps{
			AriaLabel: "Search content",
		},
	}
	
	suite.Equal(InputTypeSearch, props.Type)
	suite.Equal("searchbox", props.Type.GetAriaRole())
	suite.Equal("Search content", props.AccessibilityProps.AriaLabel)
}

func (suite *InputTestSuite) TestInputProps_DateInput() {
	props := InputProps{
		Type: InputTypeDate,
		BaseProps: BaseProps{
			Name: "birthdate",
		},
		Min: "1900-01-01",
		Max: "2024-12-31",
	}
	
	suite.Equal(InputTypeDate, props.Type)
	suite.Equal("1900-01-01", props.Min)
	suite.Equal("2024-12-31", props.Max)
}

// ============================================================================
// BACKWARD COMPATIBILITY TESTS
// ============================================================================

func (suite *InputTestSuite) TestInputProps_DeprecatedConstants() {
	// Test that deprecated constants still work
	props := InputProps{
		Type: InputText, // This is an alias for InputTypeText
		Size: SizeLG,    // Use the standard size constant
	}
	
	suite.Equal(InputTypeText, props.Type)
	suite.Equal(SizeLG, props.Size)
}

func (suite *InputTestSuite) TestInputProps_BackwardCompatibleTypes() {
	// Test all backward compatible type constants
	tests := []struct {
		deprecated InputType
		current    InputType
	}{
		{InputText, InputTypeText},
		{InputPassword, InputTypePassword},
		{InputEmail, InputTypeEmail},
		{InputNumber, InputTypeNumber},
		{InputTel, InputTypeTel},
		{InputURL, InputTypeURL},
	}
	
	for _, tt := range tests {
		suite.Equal(tt.current, tt.deprecated)
	}
}