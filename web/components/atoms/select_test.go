package atoms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// SelectTestSuite provides comprehensive tests for Select component
type SelectTestSuite struct {
	suite.Suite
}

// SetupTest runs before each test method
func (suite *SelectTestSuite) SetupTest() {
	ResetIDCounter()
}

// TestSelectTestSuite runs the select test suite
func TestSelectTestSuite(t *testing.T) {
	suite.Run(t, new(SelectTestSuite))
}

// ============================================================================
// SELECT PROPS TESTS
// ============================================================================

func (suite *SelectTestSuite) TestSelectProps_DefaultValues() {
	props := SelectProps{}
	
	// Test default values - check zero values
	suite.Equal(Size(""), props.Size) // Zero value for Size
	suite.False(props.InteractionProps.Disabled) // Disabled is in InteractionProps
	suite.False(props.Multiple)
	suite.False(props.InteractionProps.Required) // Required is in InteractionProps
}

func (suite *SelectTestSuite) TestSelectProps_BasicProperties() {
	props := SelectProps{
		BaseProps: BaseProps{
			Name: "country",
			ID:   "country-select",
		},
		LabelProps: LabelProps{
			Label: "Select Country",
		},
		PlaceholderProps: PlaceholderProps{
			Placeholder: "Choose a country",
		},
		InteractionProps: InteractionProps{
			Required: true,
		},
	}
	
	suite.Equal("country", props.BaseProps.Name)
	suite.Equal("country-select", props.BaseProps.ID)
	suite.Equal("Select Country", props.LabelProps.Label)
	suite.Equal("Choose a country", props.PlaceholderProps.Placeholder)
	suite.True(props.InteractionProps.Required)
}

func (suite *SelectTestSuite) TestSelectProps_WithOptions() {
	options := []SelectOption{
		{Value: "", Label: "Select an option", Selected: true, Disabled: false},
		{Value: "us", Label: "United States", Selected: false, Disabled: false},
		{Value: "ca", Label: "Canada", Selected: false, Disabled: false},
		{Value: "mx", Label: "Mexico", Selected: false, Disabled: true},
	}
	
	props := SelectProps{
		Options: options,
	}
	
	suite.Len(props.Options, 4)
	suite.Equal("", props.Options[0].Value)
	suite.True(props.Options[0].Selected)
	suite.True(props.Options[3].Disabled)
}

// ============================================================================
// SELECT OPTION TESTS
// ============================================================================

func (suite *SelectTestSuite) TestSelectOption_Properties() {
	option := SelectOption{
		Value:    "option1",
		Label:    "Option 1",
		Selected: true,
		Disabled: false,
		Group:    "Group A",
	}
	
	suite.Equal("option1", option.Value)
	suite.Equal("Option 1", option.Label)
	suite.True(option.Selected)
	suite.False(option.Disabled)
	suite.Equal("Group A", option.Group)
}

func (suite *SelectTestSuite) TestSelectOption_States() {
	// Test selected option
	selectedOption := SelectOption{
		Value:    "selected",
		Label:    "Selected Option",
		Selected: true,
	}
	suite.True(selectedOption.Selected)
	
	// Test disabled option
	disabledOption := SelectOption{
		Value:    "disabled",
		Label:    "Disabled Option",
		Disabled: true,
	}
	suite.True(disabledOption.Disabled)
}

// ============================================================================
// SELECT SIZE TESTS
// ============================================================================

func (suite *SelectTestSuite) TestSelectSize_AllSizes() {
	sizes := []Size{
		SizeXS, // Use general Size constants
		SizeSM,
		SizeMD,
		SizeLG,
		SizeXL,
	}
	
	for _, size := range sizes {
		props := SelectProps{Size: size}
		suite.Equal(size, props.Size)
	}
}

func (suite *SelectTestSuite) TestSelectSize_StringConversion() {
	tests := []struct {
		size     Size
		expected string
	}{
		{SizeXS, "xs"}, // Use general Size constants
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
// SELECT VALIDATION TESTS
// ============================================================================

func (suite *SelectTestSuite) TestSelectProps_ValidationStates() {
	states := []ValidationState{
		StateDefault,
		StateSuccess,
		StateError,
		StateWarning,
	}
	
	for _, state := range states {
		props := SelectProps{
			ValidationProps: ValidationProps{State: state},
		}
		suite.Equal(state, props.ValidationProps.State)
	}
}

func (suite *SelectTestSuite) TestSelectProps_ValidationMessages() {
	props := SelectProps{
		ValidationProps: ValidationProps{
			State:     StateError,
			ErrorText: "Please select a valid option",
			HelpText:  "Choose from the available options",
		},
	}
	
	msg, state := props.ValidationProps.GetFeedbackMessage()
	suite.Equal("Please select a valid option", msg)
	suite.Equal(StateError, state)
	suite.True(state.IsError())
}

// ============================================================================
// SELECT COMPOSITION TESTS
// ============================================================================

func (suite *SelectTestSuite) TestSelectProps_BasePropsComposition() {
	props := SelectProps{
		BaseProps: BaseProps{
			ID:    "priority-select",
			Name:  "priority",
			Class: "w-full",
		},
	}
	
	suite.Equal("priority-select", props.BaseProps.ID)
	suite.Equal("priority", props.BaseProps.Name)
	suite.Equal("w-full", props.BaseProps.Class)
}

func (suite *SelectTestSuite) TestSelectProps_InteractionPropsComposition() {
	props := SelectProps{
		InteractionProps: InteractionProps{
			Disabled: true,
			ReadOnly: false,
		},
	}
	
	suite.True(props.InteractionProps.Disabled)
	suite.False(props.InteractionProps.ReadOnly)
	suite.False(props.InteractionProps.IsInteractive())
}

func (suite *SelectTestSuite) TestSelectProps_AccessibilityPropsComposition() {
	props := SelectProps{
		AccessibilityProps: AccessibilityProps{
			AriaLabel:       "Priority level",
			AriaRequired:    true,
			AriaInvalid:     false,
			AriaDescribedBy: "priority-help",
		},
	}
	
	suite.Equal("Priority level", props.AccessibilityProps.AriaLabel)
	suite.True(props.AccessibilityProps.AriaRequired)
	suite.False(props.AccessibilityProps.AriaInvalid)
	suite.Equal("priority-help", props.AccessibilityProps.AriaDescribedBy)
}

func (suite *SelectTestSuite) TestSelectProps_AlpineEventHandlers() {
	props := SelectProps{
		AlpinEventHandlers: AlpinEventHandlers{
			OnChange: "handleSelectChange($event.target.value)",
			OnFocus:  "showSelectHelp()",
			OnBlur:   "hideSelectHelp()",
		},
	}
	
	suite.Equal("handleSelectChange($event.target.value)", props.AlpinEventHandlers.OnChange)
	suite.Equal("showSelectHelp()", props.AlpinEventHandlers.OnFocus)
	suite.Equal("hideSelectHelp()", props.AlpinEventHandlers.OnBlur)
}

// ============================================================================
// SELECT MULTIPLE TESTS
// ============================================================================

func (suite *SelectTestSuite) TestSelectProps_MultipleSelection() {
	props := SelectProps{
		BaseProps: BaseProps{
			Name: "skills[]", // Array notation for multiple values
		},
		LabelProps: LabelProps{
			Label: "Select Skills",
		},
		Multiple: true,
		Size:     SizeLG, // Use general Size constants
	}
	
	suite.True(props.Multiple)
	suite.Equal("skills[]", props.BaseProps.Name)
	suite.Equal("Select Skills", props.LabelProps.Label)
}

func (suite *SelectTestSuite) TestSelectProps_MultipleWithOptions() {
	options := []SelectOption{
		{Value: "js", Label: "JavaScript", Selected: true},
		{Value: "py", Label: "Python", Selected: false},
		{Value: "go", Label: "Go", Selected: true},
		{Value: "rust", Label: "Rust", Selected: false},
	}
	
	props := SelectProps{
		Options:  options,
		Multiple: true,
	}
	
	suite.True(props.Multiple)
	
	// Count selected options
	selectedCount := 0
	for _, option := range props.Options {
		if option.Selected {
			selectedCount++
		}
	}
	suite.Equal(2, selectedCount)
}

// ============================================================================
// SELECT GROUPED OPTIONS TESTS
// ============================================================================

func (suite *SelectTestSuite) TestSelectProps_GroupedOptions() {
	options := []SelectOption{
		{Value: "frontend", Label: "Frontend", Group: "Development"},
		{Value: "backend", Label: "Backend", Group: "Development"},
		{Value: "designer", Label: "UI/UX Designer", Group: "Design"},
		{Value: "writer", Label: "Content Writer", Group: "Content"},
		{Value: "manager", Label: "Project Manager", Group: "Management"},
	}
	
	props := SelectProps{
		LabelProps: LabelProps{
			Label: "Job Role",
		},
		Options: options,
	}
	
	// Check that groups are assigned
	groups := make(map[string]int)
	for _, option := range props.Options {
		groups[option.Group]++
	}
	
	suite.Equal(2, groups["Development"])
	suite.Equal(1, groups["Design"])
	suite.Equal(1, groups["Content"])
	suite.Equal(1, groups["Management"])
}

// ============================================================================
// SELECT EDGE CASES
// ============================================================================

func (suite *SelectTestSuite) TestSelectProps_EmptyOptions() {
	props := SelectProps{
		LabelProps: LabelProps{
			Label: "No Options Available",
		},
		Options:     []SelectOption{},
		PlaceholderProps: PlaceholderProps{
			Placeholder: "No options to select",
		},
	}
	
	suite.Empty(props.Options)
	suite.Equal("No options to select", props.PlaceholderProps.Placeholder)
}

func (suite *SelectTestSuite) TestSelectProps_NoPlaceholder() {
	options := []SelectOption{
		{Value: "option1", Label: "Option 1"},
		{Value: "option2", Label: "Option 2"},
	}
	
	props := SelectProps{
		Options: options,
	}
	
	suite.Empty(props.Placeholder)
	suite.Len(props.Options, 2)
}

func (suite *SelectTestSuite) TestSelectProps_LongOptionLabels() {
	options := []SelectOption{
		{
			Value: "long-option",
			Label: "This is a very long option label that might wrap to multiple lines in the dropdown",
		},
		{
			Value: "short",
			Label: "Short",
		},
	}
	
	props := SelectProps{Options: options}
	
	suite.True(len(props.Options[0].Label) > 50)
	suite.True(len(props.Options[1].Label) < 10)
}

// ============================================================================
// SELECT FORM INTEGRATION TESTS
// ============================================================================

func (suite *SelectTestSuite) TestSelectProps_FormIntegration() {
	props := SelectProps{
		BaseProps: BaseProps{
			Name: "category",
		},
		LabelProps: LabelProps{
			Label: "Product Category",
		},
		InteractionProps: InteractionProps{
			Required: true,
		},
		Options: []SelectOption{
			{Value: "", Label: "Select a category", Selected: true, Disabled: true},
			{Value: "electronics", Label: "Electronics"},
			{Value: "clothing", Label: "Clothing"},
			{Value: "books", Label: "Books"},
		},
	}
	
	suite.True(props.Required)
	suite.Equal("category", props.BaseProps.Name)
	
	// First option should be placeholder (disabled)
	suite.True(props.Options[0].Disabled)
	suite.Empty(props.Options[0].Value)
}

func (suite *SelectTestSuite) TestSelectProps_SearchableSelect() {
	props := SelectProps{
		LabelProps: LabelProps{
			Label: "Search Countries",
		},
		PlaceholderProps: PlaceholderProps{
			Placeholder: "Type to search...",
		},
		Options: []SelectOption{
			{Value: "us", Label: "United States"},
			{Value: "ca", Label: "Canada"},
			{Value: "uk", Label: "United Kingdom"},
			{Value: "fr", Label: "France"},
			{Value: "de", Label: "Germany"},
		},
		AlpinEventHandlers: AlpinEventHandlers{
			OnInput: "filterOptions($event.target.value)",
		},
	}
	
	suite.Equal("Type to search...", props.Placeholder)
	suite.Equal("filterOptions($event.target.value)", props.AlpinEventHandlers.OnInput)
	suite.Len(props.Options, 5)
}

// ============================================================================
// SELECT STYLING TESTS
// ============================================================================

func (suite *SelectTestSuite) TestSelectProps_CustomStyling() {
	props := SelectProps{
		BaseProps: BaseProps{
			Class: "custom-select border-2 border-blue-500 rounded-lg",
		},
		Size: SizeLG, // Use general Size constants
	}
	
	suite.Equal("custom-select border-2 border-blue-500 rounded-lg", props.BaseProps.Class)
	suite.Equal(SizeLG, props.Size)
}

func (suite *SelectTestSuite) TestSelectProps_ConditionalStyling() {
	// Test different states that might affect styling
	states := []struct {
		name  string
		props SelectProps
	}{
		{
			"disabled",
			SelectProps{InteractionProps: InteractionProps{Disabled: true}},
		},
		{
			"error",
			SelectProps{ValidationProps: ValidationProps{State: StateError}},
		},
		{
			"success",
			SelectProps{ValidationProps: ValidationProps{State: StateSuccess}},
		},
	}
	
	for _, state := range states {
		suite.Run(state.name, func() {
			// Each state should maintain its properties
			if state.props.InteractionProps.Disabled {
				suite.True(state.props.InteractionProps.Disabled)
			}
			if state.props.ValidationProps.State != "" {
				suite.NotEqual(StateDefault, state.props.ValidationProps.State)
			}
		})
	}
}