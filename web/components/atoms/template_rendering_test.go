package atoms

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// TemplateRenderingTestSuite tests the actual template rendering functions
type TemplateRenderingTestSuite struct {
	suite.Suite
}

// SetupTest runs before each test method
func (suite *TemplateRenderingTestSuite) SetupTest() {
	ResetIDCounter()
}

// TestTemplateRenderingTestSuite runs the template rendering test suite
func TestTemplateRenderingTestSuite(t *testing.T) {
	suite.Run(t, new(TemplateRenderingTestSuite))
}

// ============================================================================
// BUTTON TEMPLATE RENDERING TESTS
// ============================================================================

func (suite *TemplateRenderingTestSuite) TestButtonTemplate() {
	tests := []struct {
		name     string
		props    ButtonProps
		contains []string
	}{
		{
			name: "basic button",
			props: ButtonProps{
				BaseProps: BaseProps{ID: "test-button"},
				Text:      "Click me",
				Type:      ButtonTypeButton,
			},
			contains: []string{"button", "Click me", "test-button"},
		},
		{
			name: "submit button",
			props: ButtonProps{
				BaseProps: BaseProps{ID: "submit-btn"},
				Text:      "Submit",
				Type:      ButtonTypeSubmit,
			},
			contains: []string{"type=\"submit\"", "Submit", "submit-btn"},
		},
		{
			name: "disabled button",
			props: ButtonProps{
				BaseProps:        BaseProps{ID: "disabled-btn"},
				Text:             "Disabled",
				InteractionProps: InteractionProps{Disabled: true},
			},
			contains: []string{"disabled", "Disabled"},
		},
		{
			name: "button with icon",
			props: ButtonProps{
				BaseProps: BaseProps{ID: "icon-btn"},
				Text:      "With Icon",
				Icon:      IconProps{Name: "plus"},
			},
			contains: []string{"With Icon", "icon-btn"},
		},
		{
			name: "loading button",
			props: ButtonProps{
				BaseProps: BaseProps{ID: "loading-btn"},
				Text:      "Loading",
				Loading:   true,
			},
			contains: []string{"Loading"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var output strings.Builder
			err := Button(tt.props).Render(context.Background(), &output)
			suite.NoError(err)

			result := output.String()
			suite.NotEmpty(result)

			for _, expected := range tt.contains {
				suite.Contains(result, expected, "Expected %s to be in rendered output", expected)
			}
		})
	}
}

// ============================================================================
// INPUT TEMPLATE RENDERING TESTS
// ============================================================================

func (suite *TemplateRenderingTestSuite) TestInputTemplate() {
	tests := []struct {
		name     string
		props    InputProps
		contains []string
	}{
		{
			name: "text input",
			props: InputProps{
				BaseProps: BaseProps{ID: "text-input"},
				Type:      InputTypeText,
				Value:     "test value",
			},
			contains: []string{"input", "type=\"text\"", "text-input", "test value"},
		},
		{
			name: "email input",
			props: InputProps{
				BaseProps:        BaseProps{ID: "email-input"},
				Type:             InputTypeEmail,
				PlaceholderProps: PlaceholderProps{Placeholder: "Enter email"},
			},
			contains: []string{"type=\"email\"", "email-input", "Enter email"},
		},
		{
			name: "password input",
			props: InputProps{
				BaseProps: BaseProps{ID: "pwd-input"},
				Type:      InputTypePassword,
			},
			contains: []string{"type=\"password\"", "pwd-input"},
		},
		{
			name: "required input",
			props: InputProps{
				BaseProps:          BaseProps{ID: "required-input"},
				Type:               InputTypeText,
				AccessibilityProps: AccessibilityProps{AriaRequired: true},
			},
			contains: []string{"required", "required-input"},
		},
		{
			name: "disabled input",
			props: InputProps{
				BaseProps:        BaseProps{ID: "disabled-input"},
				Type:             InputTypeText,
				InteractionProps: InteractionProps{Disabled: true},
			},
			contains: []string{"disabled", "disabled-input"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var output strings.Builder
			err := Input(tt.props).Render(context.Background(), &output)
			suite.NoError(err)

			result := output.String()
			suite.NotEmpty(result)

			for _, expected := range tt.contains {
				suite.Contains(result, expected, "Expected %s to be in rendered output", expected)
			}
		})
	}
}

// ============================================================================
// ICON TEMPLATE RENDERING TESTS
// ============================================================================

func (suite *TemplateRenderingTestSuite) TestIconTemplate() {
	tests := []struct {
		name     string
		props    IconProps
		contains []string
	}{
		{
			name: "basic icon",
			props: IconProps{
				Name: "home",
				Size: IconSizeMD,
			},
			contains: []string{"svg", "w-5", "h-5"},
		},
		{
			name: "large icon",
			props: IconProps{
				Name: "search",
				Size: IconSizeLG,
			},
			contains: []string{"svg", "w-6", "h-6"},
		},
		{
			name: "icon with custom class",
			props: IconProps{
				Name:  "settings",
				Class: "custom-icon-class",
			},
			contains: []string{"svg", "custom-icon-class"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var output strings.Builder
			err := Icon(tt.props).Render(context.Background(), &output)
			suite.NoError(err)

			result := output.String()
			suite.NotEmpty(result)

			for _, expected := range tt.contains {
				suite.Contains(result, expected, "Expected %s to be in rendered output", expected)
			}
		})
	}
}

// ============================================================================
// CHECKBOX TEMPLATE RENDERING TESTS
// ============================================================================

func (suite *TemplateRenderingTestSuite) TestCheckboxTemplate() {
	tests := []struct {
		name     string
		props    CheckboxProps
		contains []string
	}{
		{
			name: "basic checkbox",
			props: CheckboxProps{
				BaseProps: BaseProps{ID: "basic-checkbox"},
				LabelProps: LabelProps{
					Label: "Accept terms",
				},
			},
			contains: []string{"checkbox", "basic-checkbox", "Accept terms"},
		},
		{
			name: "checked checkbox",
			props: CheckboxProps{
				BaseProps: BaseProps{ID: "checked-checkbox"},
				Checked:   true,
				LabelProps: LabelProps{
					Label: "Checked option",
				},
			},
			contains: []string{"checked", "checked-checkbox", "Checked option"},
		},
		{
			name: "disabled checkbox",
			props: CheckboxProps{
				BaseProps:        BaseProps{ID: "disabled-checkbox"},
				InteractionProps: InteractionProps{Disabled: true},
				LabelProps: LabelProps{
					Label: "Disabled option",
				},
			},
			contains: []string{"disabled", "disabled-checkbox", "Disabled option"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var output strings.Builder
			err := Checkbox(tt.props).Render(context.Background(), &output)
			suite.NoError(err)

			result := output.String()
			suite.NotEmpty(result)

			for _, expected := range tt.contains {
				suite.Contains(result, expected, "Expected %s to be in rendered output", expected)
			}
		})
	}
}

// ============================================================================
// SELECT TEMPLATE RENDERING TESTS
// ============================================================================

func (suite *TemplateRenderingTestSuite) TestSelectTemplate() {
	tests := []struct {
		name     string
		props    SelectProps
		contains []string
	}{
		{
			name: "basic select",
			props: SelectProps{
				BaseProps:        BaseProps{ID: "basic-select"},
				PlaceholderProps: PlaceholderProps{Placeholder: "Choose option"},
				Options: []SelectOption{
					{Value: "1", Label: "Option 1"},
					{Value: "2", Label: "Option 2"},
				},
			},
			contains: []string{"select", "basic-select", "Choose option", "Option 1", "Option 2"},
		},
		{
			name: "select with selected value",
			props: SelectProps{
				BaseProps: BaseProps{ID: "selected-select"},
				Value:     "2",
				Options: []SelectOption{
					{Value: "1", Label: "Option 1"},
					{Value: "2", Label: "Option 2", Selected: true},
				},
			},
			contains: []string{"selected-select", "Option 1", "Option 2", "selected"},
		},
		{
			name: "disabled select",
			props: SelectProps{
				BaseProps:        BaseProps{ID: "disabled-select"},
				InteractionProps: InteractionProps{Disabled: true},
				Options: []SelectOption{
					{Value: "1", Label: "Option 1"},
				},
			},
			contains: []string{"disabled", "disabled-select", "Option 1"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var output strings.Builder
			err := Select(tt.props).Render(context.Background(), &output)
			suite.NoError(err)

			result := output.String()
			suite.NotEmpty(result)

			for _, expected := range tt.contains {
				suite.Contains(result, expected, "Expected %s to be in rendered output", expected)
			}
		})
	}
}

// ============================================================================
// BADGE TEMPLATE RENDERING TESTS
// ============================================================================

func (suite *TemplateRenderingTestSuite) TestBadgeTemplate() {
	tests := []struct {
		name     string
		props    BadgeProps
		contains []string
	}{
		{
			name: "basic badge",
			props: BadgeProps{
				BaseProps: BaseProps{ID: "basic-badge"},
				Text:      "New",
			},
			contains: []string{"basic-badge", "New"},
		},
		{
			name: "badge with icon",
			props: BadgeProps{
				BaseProps: BaseProps{ID: "icon-badge"},
				Text:      "Status",
				Icon:      "check",
			},
			contains: []string{"icon-badge", "Status", "check"},
		},
		{
			name: "colored badge",
			props: BadgeProps{
				BaseProps: BaseProps{ID: "colored-badge"},
				Text:      "Error",
				Color:     ColorDanger,
			},
			contains: []string{"colored-badge", "Error"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var output strings.Builder
			err := Badge(tt.props).Render(context.Background(), &output)
			suite.NoError(err)

			result := output.String()
			suite.NotEmpty(result)

			for _, expected := range tt.contains {
				suite.Contains(result, expected, "Expected %s to be in rendered output", expected)
			}
		})
	}
}

// ============================================================================
// SPINNER TEMPLATE RENDERING TESTS
// ============================================================================

func (suite *TemplateRenderingTestSuite) TestSpinnerTemplate() {
	tests := []struct {
		name     string
		props    SpinnerProps
		contains []string
	}{
		{
			name: "basic spinner",
			props: SpinnerProps{
				BaseProps: BaseProps{ID: "basic-spinner"},
			},
			contains: []string{"basic-spinner"},
		},
		{
			name: "spinner with label",
			props: SpinnerProps{
				BaseProps: BaseProps{ID: "labeled-spinner"},
				Label:     "Loading...",
			},
			contains: []string{"labeled-spinner", "Loading..."},
		},
		{
			name: "large spinner",
			props: SpinnerProps{
				BaseProps: BaseProps{ID: "large-spinner"},
				Size:      SpinnerSizeLG,
			},
			contains: []string{"large-spinner"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var output strings.Builder
			err := Spinner(tt.props).Render(context.Background(), &output)
			suite.NoError(err)

			result := output.String()
			suite.NotEmpty(result)

			for _, expected := range tt.contains {
				suite.Contains(result, expected, "Expected %s to be in rendered output", expected)
			}
		})
	}
}

// ============================================================================
// ACTION TEMPLATE RENDERING TESTS
// ============================================================================

func (suite *TemplateRenderingTestSuite) TestActionTemplate() {
	tests := []struct {
		name     string
		props    ActionProps
		contains []string
	}{
		{
			name: "button action",
			props: ActionProps{
				BaseProps: BaseProps{ID: "button-action"},
				Label:     "Click me",
				Element:   "button",
			},
			contains: []string{"button-action", "Click me"},
		},
		{
			name: "link action",
			props: ActionProps{
				BaseProps: BaseProps{ID: "link-action"},
				Label:     "Go to page",
				Element:   "a",
				URL:       "/page",
			},
			contains: []string{"link-action", "Go to page", "/page"},
		},
		{
			name: "action with icon",
			props: ActionProps{
				BaseProps: BaseProps{ID: "icon-action"},
				Label:     "Save",
				Element:   "button",
				Icon:      IconProps{Name: "save"},
			},
			contains: []string{"icon-action", "Save"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var output strings.Builder
			err := Action(tt.props).Render(context.Background(), &output)
			suite.NoError(err)

			result := output.String()
			suite.NotEmpty(result)

			for _, expected := range tt.contains {
				suite.Contains(result, expected, "Expected %s to be in rendered output", expected)
			}
		})
	}
}

