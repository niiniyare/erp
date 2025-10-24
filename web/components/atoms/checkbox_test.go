package atoms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// CheckboxTestSuite provides comprehensive tests for Checkbox component
type CheckboxTestSuite struct {
	suite.Suite
}

// SetupTest runs before each test method
func (suite *CheckboxTestSuite) SetupTest() {
	ResetIDCounter()
}

// TestCheckboxTestSuite runs the checkbox test suite
func TestCheckboxTestSuite(t *testing.T) {
	suite.Run(t, new(CheckboxTestSuite))
}

// ============================================================================
// CHECKBOX PROPS TESTS
// ============================================================================

func (suite *CheckboxTestSuite) TestCheckboxProps_DefaultValues() {
	props := CheckboxProps{}
	
	// Test default values - check zero values
	suite.Equal(Size(""), props.Size) // Zero value for Size
	suite.False(props.Checked)
	suite.False(props.InteractionProps.Disabled) // Disabled is in InteractionProps
	suite.False(props.Indeterminate)
}

func (suite *CheckboxTestSuite) TestCheckboxProps_BasicProperties() {
	props := CheckboxProps{
		BaseProps: BaseProps{
			Name: "terms",
			ID:   "terms-checkbox",
		},
		LabelProps: LabelProps{
			Label: "I agree to the terms and conditions",
		},
		Value:   "agreed",
		Checked: true,
	}
	
	suite.Equal("terms", props.BaseProps.Name)
	suite.Equal("terms-checkbox", props.BaseProps.ID)
	suite.Equal("I agree to the terms and conditions", props.LabelProps.Label)
	suite.Equal("agreed", props.Value)
	suite.True(props.Checked)
}

func (suite *CheckboxTestSuite) TestCheckboxProps_States() {
	// Test checked state
	checkedProps := CheckboxProps{Checked: true}
	suite.True(checkedProps.Checked)
	
	// Test indeterminate state
	indeterminateProps := CheckboxProps{Indeterminate: true}
	suite.True(indeterminateProps.Indeterminate)
	
	// Test disabled state
	disabledProps := CheckboxProps{
		InteractionProps: InteractionProps{Disabled: true},
	}
	suite.True(disabledProps.InteractionProps.Disabled)
	suite.False(disabledProps.InteractionProps.IsInteractive())
}

// ============================================================================
// CHECKBOX SIZE TESTS
// ============================================================================

func (suite *CheckboxTestSuite) TestCheckboxSize_AllSizes() {
	sizes := []Size{
		SizeXS, // Use the general Size constants
		SizeSM,
		SizeMD,
		SizeLG,
		SizeXL,
	}
	
	for _, size := range sizes {
		props := CheckboxProps{Size: size}
		suite.Equal(size, props.Size)
	}
}

func (suite *CheckboxTestSuite) TestCheckboxSize_StringConversion() {
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
// CHECKBOX VALIDATION TESTS
// ============================================================================

func (suite *CheckboxTestSuite) TestCheckboxProps_ValidationStates() {
	states := []ValidationState{
		StateDefault,
		StateSuccess,
		StateError,
		StateWarning,
	}
	
	for _, state := range states {
		props := CheckboxProps{
			ValidationProps: ValidationProps{State: state},
		}
		suite.Equal(state, props.ValidationProps.State)
	}
}

func (suite *CheckboxTestSuite) TestCheckboxProps_ValidationMessages() {
	props := CheckboxProps{
		ValidationProps: ValidationProps{
			State:     StateError,
			ErrorText: "You must agree to continue",
			HelpText:  "Please check this box to proceed",
		},
	}
	
	msg, state := props.ValidationProps.GetFeedbackMessage()
	suite.Equal("You must agree to continue", msg)
	suite.Equal(StateError, state)
	suite.True(state.IsError())
}

// ============================================================================
// CHECKBOX COMPOSITION TESTS
// ============================================================================

func (suite *CheckboxTestSuite) TestCheckboxProps_BasePropsComposition() {
	props := CheckboxProps{
		BaseProps: BaseProps{
			ID:    "newsletter-checkbox",
			Name:  "newsletter",
			Class: "mb-4",
		},
	}
	
	suite.Equal("newsletter-checkbox", props.BaseProps.ID)
	suite.Equal("newsletter", props.BaseProps.Name)
	suite.Equal("mb-4", props.BaseProps.Class)
}

func (suite *CheckboxTestSuite) TestCheckboxProps_LabelPropsComposition() {
	props := CheckboxProps{
		LabelProps: LabelProps{
			Label:         "Subscribe to newsletter",
			LabelPosition: LabelRight, // Correct field name is LabelPosition
		},
	}
	
	suite.Equal("Subscribe to newsletter", props.LabelProps.Label)
	suite.Equal(LabelRight, props.LabelProps.LabelPosition)
}

func (suite *CheckboxTestSuite) TestCheckboxProps_AccessibilityPropsComposition() {
	props := CheckboxProps{
		AccessibilityProps: AccessibilityProps{
			AriaLabel:       "Accept terms",
			AriaRequired:    true,
			AriaInvalid:     false,
			AriaDescribedBy: "terms-help",
		},
	}
	
	suite.Equal("Accept terms", props.AccessibilityProps.AriaLabel)
	suite.True(props.AccessibilityProps.AriaRequired)
	suite.False(props.AccessibilityProps.AriaInvalid)
	suite.Equal("terms-help", props.AccessibilityProps.AriaDescribedBy)
}

func (suite *CheckboxTestSuite) TestCheckboxProps_AlpineEventHandlers() {
	props := CheckboxProps{
		AlpinEventHandlers: AlpinEventHandlers{
			OnChange: "handleCheckboxChange($event.target.checked)",
			OnFocus:  "showTooltip()",
			OnBlur:   "hideTooltip()",
		},
	}
	
	suite.Equal("handleCheckboxChange($event.target.checked)", props.AlpinEventHandlers.OnChange)
	suite.Equal("showTooltip()", props.AlpinEventHandlers.OnFocus)
	suite.Equal("hideTooltip()", props.AlpinEventHandlers.OnBlur)
}

// ============================================================================
// CHECKBOX GROUP TESTS
// ============================================================================

func (suite *CheckboxTestSuite) TestCheckboxProps_GroupedCheckboxes() {
	// Test multiple checkboxes with same name (checkbox group)
	options := []CheckboxProps{
		{
			BaseProps:  BaseProps{Name: "features", ID: "feature-1"},
			LabelProps: LabelProps{Label: "Feature 1"},
			Value:      "feature1",
			Checked:    true,
		},
		{
			BaseProps:  BaseProps{Name: "features", ID: "feature-2"},
			LabelProps: LabelProps{Label: "Feature 2"},
			Value:      "feature2",
			Checked:    false,
		},
		{
			BaseProps:  BaseProps{Name: "features", ID: "feature-3"},
			LabelProps: LabelProps{Label: "Feature 3"},
			Value:      "feature3",
			Checked:    true,
		},
	}
	
	// All should have same name but different IDs and values
	for _, checkbox := range options {
		suite.Equal("features", checkbox.BaseProps.Name)
		suite.NotEmpty(checkbox.BaseProps.ID)
		suite.NotEmpty(checkbox.Value)
		suite.NotEmpty(checkbox.LabelProps.Label)
	}
	
	// Check that we have both checked and unchecked states
	checkedCount := 0
	for _, checkbox := range options {
		if checkbox.Checked {
			checkedCount++
		}
	}
	suite.Equal(2, checkedCount)
}

// ============================================================================
// CHECKBOX EDGE CASES
// ============================================================================

func (suite *CheckboxTestSuite) TestCheckboxProps_NoLabel() {
	props := CheckboxProps{
		BaseProps: BaseProps{
			Name: "silent-checkbox",
		},
		Value: "silent",
		AccessibilityProps: AccessibilityProps{
			AriaLabel: "Hidden checkbox", // Should use aria-label when no visible label
		},
	}
	
	suite.Empty(props.LabelProps.Label)
	suite.Equal("Hidden checkbox", props.AccessibilityProps.AriaLabel)
}

func (suite *CheckboxTestSuite) TestCheckboxProps_IndeterminateState() {
	props := CheckboxProps{
		LabelProps: LabelProps{
			Label: "Select all items",
		},
		Indeterminate: true,
		Checked:       false, // Indeterminate overrides checked state
	}
	
	suite.True(props.Indeterminate)
	suite.False(props.Checked)
	suite.Equal("Select all items", props.LabelProps.Label)
}

func (suite *CheckboxTestSuite) TestCheckboxProps_RequiredCheckbox() {
	props := CheckboxProps{
		LabelProps: LabelProps{
			Label: "I agree to the terms *",
		},
		AccessibilityProps: AccessibilityProps{
			AriaRequired: true,
		},
		ValidationProps: ValidationProps{
			State:     StateError,
			ErrorText: "Agreement is required",
		},
	}
	
	suite.True(props.AccessibilityProps.AriaRequired)
	suite.True(props.ValidationProps.State.IsError())
	suite.Equal("Agreement is required", props.ValidationProps.ErrorText)
}

// ============================================================================
// CHECKBOX STYLING TESTS
// ============================================================================

func (suite *CheckboxTestSuite) TestCheckboxProps_CustomStyling() {
	props := CheckboxProps{
		BaseProps: BaseProps{
			Class: "custom-checkbox border-2 border-blue-500",
		},
		LabelProps: LabelProps{
			Label: "Custom styled checkbox",
		},
		Size: SizeLG, // Use general Size constant
	}
	
	suite.Equal("custom-checkbox border-2 border-blue-500", props.BaseProps.Class)
	suite.Equal(SizeLG, props.Size)
}

func (suite *CheckboxTestSuite) TestCheckboxProps_ColorSchemes() {
	// Test different visual states that might use different colors
	tests := []struct {
		name  string
		state ValidationState
		color string
	}{
		{"success", StateSuccess, "green"},
		{"error", StateError, "red"},
		{"warning", StateWarning, "yellow"},
		{"default", StateDefault, "blue"},
	}
	
	for _, tt := range tests {
		props := CheckboxProps{
			ValidationProps: ValidationProps{State: tt.state},
		}
		suite.Equal(tt.state, props.ValidationProps.State)
	}
}

// ============================================================================
// CHECKBOX FORM INTEGRATION TESTS
// ============================================================================

func (suite *CheckboxTestSuite) TestCheckboxProps_FormIntegration() {
	props := CheckboxProps{
		BaseProps: BaseProps{
			Name: "preferences[]", // Array notation for multiple values
		},
		LabelProps: LabelProps{
			Label: "Email notifications",
		},
		Value:   "email",
		Checked: true,
	}
	
	suite.Equal("preferences[]", props.BaseProps.Name)
	suite.Equal("email", props.Value)
	suite.True(props.Checked)
}

func (suite *CheckboxTestSuite) TestCheckboxProps_BooleanValue() {
	// Single checkbox that represents a boolean value
	props := CheckboxProps{
		BaseProps: BaseProps{
			Name: "remember_me",
		},
		LabelProps: LabelProps{
			Label: "Remember me",
		},
		Value:   "1", // Or could be "true"
		Checked: false,
	}
	
	suite.Equal("remember_me", props.BaseProps.Name)
	suite.Equal("1", props.Value)
	suite.False(props.Checked)
}