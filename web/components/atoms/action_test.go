package atoms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// ActionTestSuite provides comprehensive tests for Action component and utilities
type ActionTestSuite struct {
	suite.Suite
}

// SetupTest runs before each test method
func (suite *ActionTestSuite) SetupTest() {
	ResetIDCounter()
}

// TestActionTestSuite runs the action test suite
func TestActionTestSuite(t *testing.T) {
	suite.Run(t, new(ActionTestSuite))
}

// ============================================================================
// ACTION FACTORY TESTS
// ============================================================================

func (suite *ActionTestSuite) TestNewAction() {
	action := NewAction(ActionTypeClick, "handleClick()")
	
	suite.NotEmpty(action.ID) // ID is auto-generated
	suite.Equal(ActionTypeClick, action.ActionType)
	suite.Equal("handleClick()", action.Command)
	suite.Equal(VariantDefault, action.Variant)
	suite.Equal(SizeMD, action.Size)
}

func (suite *ActionTestSuite) TestNewClickAction() {
	action := NewClickAction("handleClick()")
	
	suite.Equal(ActionTypeClick, action.ActionType)
	suite.Equal("handleClick()", action.Command)
	suite.Equal("button", action.Element)
	suite.Equal("button", action.Type)
}

func (suite *ActionTestSuite) TestNewSubmitAction() {
	action := NewSubmitAction("submitForm()")
	
	suite.Equal(ActionTypeSubmit, action.ActionType)
	suite.Equal("submitForm()", action.Command)
	suite.Equal("button", action.Element)
	suite.Equal("submit", action.Type)
}

func (suite *ActionTestSuite) TestNewToggleAction() {
	action := NewToggleAction("menuOpen")
	
	suite.Equal(ActionTypeToggle, action.ActionType)
	suite.Equal("menuOpen = !menuOpen", action.Command)
	suite.Equal("button", action.Element)
	suite.Equal("button", action.Type)
}

func (suite *ActionTestSuite) TestNewConfirmAction() {
	action := NewConfirmAction("Are you sure?", "deleteItem()")
	
	suite.Equal(ActionTypeConfirm, action.ActionType)
	suite.Contains(action.Command, "Are you sure?")
	suite.Contains(action.Command, "deleteItem()")
	suite.Equal(ExecutionModeConfirm, action.ExecutionMode)
	suite.Equal("Are you sure?", action.ConfirmMessage)
}

func (suite *ActionTestSuite) TestNewNavigateAction() {
	action := NewNavigateAction("/dashboard")
	
	suite.Equal(ActionTypeNavigate, action.ActionType)
	suite.Contains(action.Command, "/dashboard")
	suite.Equal("/dashboard", action.URL)
}

func (suite *ActionTestSuite) TestNewApiAction() {
	action := NewApiAction("/api/data", "GET")
	
	suite.Equal(ActionTypeApi, action.ActionType)
	suite.Contains(action.Command, "/api/data")
	suite.Contains(action.Command, "GET")
	suite.Equal("/api/data", action.ApiEndpoint)
	suite.Equal("GET", action.ApiMethod)
}

// ============================================================================
// ACTION UTILITY TESTS
// ============================================================================

func (suite *ActionTestSuite) TestGetActionClasses() {
	tests := []struct {
		name     string
		action   ActionProps
		contains []string
	}{
		{
			"default action",
			ActionProps{Variant: VariantDefault, Size: SizeMD},
			[]string{"cursor-pointer", "select-none"},
		},
		{
			"disabled action",
			ActionProps{Disabled: true},
			[]string{"opacity-50", "cursor-not-allowed"},
		},
		{
			"loading action",
			ActionProps{Loading: true},
			[]string{"opacity-75", "cursor-wait"},
		},
	}
	
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			classes := getActionClasses(tt.action)
			for _, expected := range tt.contains {
				suite.Contains(classes, expected)
			}
		})
	}
}

func (suite *ActionTestSuite) TestGetEventAttribute() {
	tests := []struct {
		actionType string
		expected   string
	}{
		{ActionTypeClick, "x-on:click"},
		{ActionTypeSubmit, "x-on:click"},
		{ActionTypeToggle, "x-on:click"},
		{ActionTypeConfirm, "x-on:click"},
		{ActionTypeNavigate, "x-on:click"},
		{ActionTypeApi, "x-on:click"},
	}
	
	for _, tt := range tests {
		result := getEventAttribute(tt.actionType)
		suite.Equal(tt.expected, result)
	}
}

func (suite *ActionTestSuite) TestGetEventHandler() {
	tests := []struct {
		name     string
		action   ActionProps
		expected string
	}{
		{
			"simple command",
			ActionProps{Command: "handleClick()"},
			"handleClick()",
		},
		{
			"deferred execution",
			ActionProps{
				Command:       "doAction()",
				ExecutionMode: ExecutionModeDeferred,
				Delay:         1000,
			},
			"setTimeout(",
		},
		{
			"throttled execution",
			ActionProps{
				Command:       "doAction()",
				ExecutionMode: ExecutionModeThrottle,
				ThrottleMs:    500,
			},
			"$throttle(",
		},
	}
	
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			handler := getEventHandler(tt.action)
			suite.Contains(handler, tt.expected)
		})
	}
}

func (suite *ActionTestSuite) TestGetActionIconSize() {
	tests := []struct {
		size     Size
		expected Size
	}{
		{SizeXS, SizeXS},
		{SizeSM, SizeSM},
		{SizeMD, SizeSM}, // Default returns SM
		{SizeLG, SizeMD},
		{SizeXL, SizeLG},
	}
	
	for _, tt := range tests {
		result := getActionIconSize(tt.size)
		suite.Equal(tt.expected, result)
	}
}

// ============================================================================
// ACTION BUILDER TESTS
// ============================================================================

func (suite *ActionTestSuite) TestActionPropsModification() {
	action := NewAction(ActionTypeClick, "handleClick()")
	
	// Test modifying properties directly
	action.Label = "Test Label"
	action.Icon = IconProps{Name: "home"}
	action.Variant = VariantPrimary
	action.Size = SizeLG
	
	suite.Equal("Test Label", action.Label)
	suite.Equal("home", action.Icon.Name)
	suite.Equal(VariantPrimary, action.Variant)
	suite.Equal(SizeLG, action.Size)
}

// ============================================================================
// ACTION PROPS TESTS
// ============================================================================

func (suite *ActionTestSuite) TestActionProps_DefaultValues() {
	props := ActionProps{}
	
	// Empty ActionProps should have zero values
	suite.Equal("", props.ActionType)
	suite.Equal(Variant(""), props.Variant)
	suite.Equal(Size(""), props.Size)
	suite.False(props.Disabled)
	suite.False(props.Loading)
}

func (suite *ActionTestSuite) TestActionProps_ComplexAction() {
	props := ActionProps{
		BaseProps: BaseProps{
			ID:    "complex-action",
			Class: "custom-class",
		},
		ActionType:     ActionTypeApi,
		Label:          "Save Data",
		Command:        "saveData()",
		ApiEndpoint:    "/api/save",
		ApiMethod:      "POST",
		Target:         "#result",
		Icon:           IconProps{Name: "save"},
		Variant:        VariantPrimary,
		Size:           SizeLG,
		Loading:        true,
		ConfirmMessage: "Save changes?",
	}
	
	suite.Equal("complex-action", props.BaseProps.ID)
	suite.Equal(ActionTypeApi, props.ActionType)
	suite.Equal("Save Data", props.Label)
	suite.Equal("saveData()", props.Command)
	suite.Equal("/api/save", props.ApiEndpoint)
	suite.Equal("POST", props.ApiMethod)
	suite.Equal("#result", props.Target)
	suite.Equal("save", props.Icon.Name)
	suite.Equal(VariantPrimary, props.Variant)
	suite.Equal(SizeLG, props.Size)
	suite.True(props.Loading)
	suite.Equal("Save changes?", props.ConfirmMessage)
}

// ============================================================================
// ACTION TYPE TESTS
// ============================================================================

func (suite *ActionTestSuite) TestActionTypes() {
	types := []string{
		ActionTypeClick,
		ActionTypeSubmit,
		ActionTypeToggle,
		ActionTypeConfirm,
		ActionTypeNavigate,
		ActionTypeApi,
		ActionTypeCustom,
	}
	
	for _, actionType := range types {
		suite.NotEmpty(actionType)
	}
}

// ============================================================================
// ACTION EDGE CASES
// ============================================================================

func (suite *ActionTestSuite) TestEmptyActionProps() {
	props := ActionProps{
		Label:   "",
		Command: "",
		Icon:    IconProps{},
	}
	
	suite.Empty(props.Label)
	suite.Empty(props.Command)
	suite.Empty(props.Icon.Name)
}

func (suite *ActionTestSuite) TestActionWithLongLabel() {
	longLabel := "This is a very long action label that might wrap to multiple lines"
	action := NewAction(ActionTypeClick, "handleClick()")
	action.Label = longLabel
	
	suite.Equal(longLabel, action.Label)
	suite.True(len(action.Label) > 50)
}

func (suite *ActionTestSuite) TestActionWithSpecialCharacters() {
	specialCommand := "confirm('Are you sure?') && deleteItem($event)"
	action := NewAction(ActionTypeConfirm, specialCommand)
	
	suite.Equal(specialCommand, action.Command)
	suite.Contains(action.Command, "'")
	suite.Contains(action.Command, "&")
}