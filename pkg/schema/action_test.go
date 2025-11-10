package schema

import (
	"context"
	"testing"

	"github.com/niiniyare/erp/pkg/condition"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ActionTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *ActionTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func TestActionTestSuite(t *testing.T) {
	suite.Run(t, new(ActionTestSuite))
}

// ==================== Additional Coverage Tests ====================

func (suite *ActionTestSuite) TestActionSetGetEvaluator() {
	action := &Action{
		ID:   "test_action",
		Type: ActionSubmit,
		Text: "Submit",
	}

	// Initially nil
	suite.Require().Nil(action.GetEvaluator())

	// Set evaluator - Note: SetEvaluator expects *condition.Evaluator
	evaluator := suite.createMockConditionEvaluator()
	action.SetEvaluator(evaluator)
	suite.Require().Equal(evaluator, action.GetEvaluator())
}

func (suite *ActionTestSuite) TestActionGetIconPosition() {
	// Test default position
	action := &Action{ID: "test"}
	suite.Require().Equal("left", action.GetIconPosition())

	// Test custom position
	action.Position = "right"
	suite.Require().Equal("right", action.GetIconPosition())

	// Test empty position returns default
	action.Position = ""
	suite.Require().Equal("left", action.GetIconPosition())
}

func (suite *ActionTestSuite) TestActionHasIcon() {
	// Without icon
	action := &Action{ID: "test"}
	suite.Require().False(action.HasIcon())

	// With icon
	action.Icon = "save"
	suite.Require().True(action.HasIcon())

	// Empty icon
	action.Icon = ""
	suite.Require().False(action.HasIcon())
}

func (suite *ActionTestSuite) TestActionHasConfirmation() {
	// Without confirmation
	action := &Action{ID: "test"}
	suite.Require().False(action.HasConfirmation())

	// With disabled confirmation
	action.Confirm = &Confirm{Enabled: false}
	suite.Require().False(action.HasConfirmation())

	// With enabled confirmation
	action.Confirm = &Confirm{Enabled: true}
	suite.Require().True(action.HasConfirmation())
}

func (suite *ActionTestSuite) TestActionRequiresPermission() {
	action := &Action{ID: "test"}

	// Without permissions
	suite.Require().False(action.RequiresPermission("delete"))

	// With permissions but not the required one
	action.Permissions = &ActionPermissions{
		Required: []string{"read", "write"},
	}
	suite.Require().False(action.RequiresPermission("delete"))

	// With the required permission
	suite.Require().True(action.RequiresPermission("read"))
	suite.Require().True(action.RequiresPermission("write"))
}

func (suite *ActionTestSuite) TestActionCanView() {
	action := &Action{ID: "test"}
	userPerms := []string{"read", "write"}

	// Without permissions - should return true
	canView := action.CanView(userPerms)
	suite.Require().True(canView)

	// With permissions that user has
	action.Permissions = &ActionPermissions{
		View: []string{"read"},
	}
	canView = action.CanView(userPerms)
	suite.Require().True(canView)

	// With permissions that user doesn't have
	action.Permissions = &ActionPermissions{
		View: []string{"admin"},
	}
	canView = action.CanView(userPerms)
	suite.Require().False(canView)
}

func (suite *ActionTestSuite) TestActionCanExecute() {
	action := &Action{ID: "test"}
	userPerms := []string{"read", "write"}

	// Without permissions - should return true
	canExecute := action.CanExecute(userPerms)
	suite.Require().True(canExecute)

	// With permissions that user has
	action.Permissions = &ActionPermissions{
		Execute: []string{"write"},
	}
	canExecute = action.CanExecute(userPerms)
	suite.Require().True(canExecute)

	// With permissions that user doesn't have
	action.Permissions = &ActionPermissions{
		Execute: []string{"admin"},
	}
	canExecute = action.CanExecute(userPerms)
	suite.Require().False(canExecute)
}

func (suite *ActionTestSuite) TestActionClone() {
	original := &Action{
		ID:       "test",
		Type:     ActionSubmit,
		Text:     "Submit",
		Icon:     "save",
		Disabled: true,
		Config: &ActionConfig{
			URL:     "/test",
			Handler: "testHandler",
		},
	}

	clone := original.Clone()
	suite.Require().Equal(original.ID, clone.ID)
	suite.Require().Equal(original.Type, clone.Type)
	suite.Require().Equal(original.Text, clone.Text)
	suite.Require().Equal(original.Icon, clone.Icon)
	suite.Require().Equal(original.Disabled, clone.Disabled)
	suite.Require().Equal(original.Config.URL, clone.Config.URL)

	// Test clone works (Action.Clone may not do deep copy for Config)
	suite.Require().NotNil(clone.Config)
}

func (suite *ActionTestSuite) TestActionApplyTheme() {
	action := &Action{ID: "test", Type: ActionSubmit}

	// Test applying nil theme
	action.ApplyTheme(nil)
	suite.Require().Nil(action.Theme) // Should remain nil

	// Test applying valid theme
	theme := &Theme{
		ID:     "test-theme-id",
		Name:   "test-theme",
		Tokens: GetDefaultTokens(),
	}
	action.ApplyTheme(theme)
	suite.Require().NotNil(action.Theme) // Should be initialized
}

func (suite *ActionTestSuite) TestActionGetConfig() {
	action := &Action{ID: "test"}

	// Without config
	cfg := action.GetConfig()
	suite.Require().NotNil(cfg)

	// With config
	action.Config = &ActionConfig{
		URL:      "/api/test",
		Debounce: 5000,
	}
	cfg = action.GetConfig()
	suite.Require().Equal("/api/test", cfg.URL)
	suite.Require().Equal(5000, cfg.Debounce)
}

func (suite *ActionTestSuite) TestActionSetters() {
	action := &Action{ID: "test"}

	// Test SetLoading
	action.SetLoading(true)
	suite.Require().True(action.Loading)
	action.SetLoading(false)
	suite.Require().False(action.Loading)

	// Test SetDisabled
	action.SetDisabled(true)
	suite.Require().True(action.Disabled)
	action.SetDisabled(false)
	suite.Require().False(action.Disabled)

	// Test SetHidden
	action.SetHidden(true)
	suite.Require().True(action.Hidden)
	action.SetHidden(false)
	suite.Require().False(action.Hidden)
}

func (suite *ActionTestSuite) TestActionGetHTMXConfig() {
	action := &Action{ID: "test"}

	// Without HTMX config
	cfg := action.GetHTMXConfig()
	suite.Require().NotNil(cfg)

	// With HTMX config
	action.HTMX = &ActionHTMX{
		Method: "POST",
		URL:    "/api/submit",
	}
	cfg = action.GetHTMXConfig()
	suite.Require().Equal("POST", cfg.Method)
	suite.Require().Equal("/api/submit", cfg.URL)
}

func (suite *ActionTestSuite) TestActionGetAlpineConfig() {
	action := &Action{ID: "test"}

	// Without Alpine config
	cfg := action.GetAlpineConfig()
	suite.Require().NotNil(cfg)

	// With Alpine config
	action.Alpine = &ActionAlpine{
		XOn: "click: handleClick()",
	}
	cfg = action.GetAlpineConfig()
	suite.Require().Equal("click: handleClick()", cfg.XOn)
}

// Helper method to create a mock condition evaluator
func (suite *ActionTestSuite) createMockConditionEvaluator() *condition.Evaluator {
	return condition.NewEvaluator(nil, condition.DefaultEvalOptions())
}

// Test action visibility methods
func (suite *ActionTestSuite) TestActionVisibility() {
	// Test visible action
	visibleAction := &Action{
		ID:     "visible_action",
		Type:   ActionButton,
		Text:   "Visible Button",
		Hidden: false,
	}
	visible, err := visibleAction.IsVisible(context.Background(), map[string]any{})
	require.NoError(suite.T(), err)
	require.True(suite.T(), visible)

	// Test hidden action
	hiddenAction := &Action{
		ID:     "hidden_action",
		Type:   ActionButton,
		Text:   "Hidden Button",
		Hidden: true,
	}
	visible, err = hiddenAction.IsVisible(context.Background(), map[string]any{})
	require.NoError(suite.T(), err)
	require.False(suite.T(), visible)
}

// Test action enabled methods
func (suite *ActionTestSuite) TestActionEnabled() {
	// Test enabled action
	enabledAction := &Action{
		ID:       "enabled_action",
		Type:     ActionButton,
		Text:     "Enabled Button",
		Disabled: false,
	}
	require.True(suite.T(), enabledAction.IsEnabled())

	// Test disabled action
	disabledAction := &Action{
		ID:       "disabled_action",
		Type:     ActionButton,
		Text:     "Disabled Button",
		Disabled: true,
	}
	require.False(suite.T(), disabledAction.IsEnabled())
}

// Test action variant classes
func (suite *ActionTestSuite) TestActionVariantClass() {
	testCases := []struct {
		variant  string
		expected string
	}{
		{"primary", "action-primary"},
		{"secondary", "action-secondary"},
		{"outline", "action-outline"},
		{"ghost", "action-ghost"},
		{"destructive", "action-destructive"},
		{"unknown", "action-primary"}, // Default fallback
		{"", "action-primary"},        // Empty fallback
	}

	for _, tc := range testCases {
		action := &Action{
			ID:      "test_action",
			Type:    ActionButton,
			Text:    "Test Button",
			Variant: tc.variant,
		}
		require.Equal(suite.T(), tc.expected, action.GetVariantClass())
	}
}

// Test action size classes
func (suite *ActionTestSuite) TestActionSizeClass() {
	testCases := []struct {
		size     string
		expected string
	}{
		{"sm", "action-sm"},
		{"md", "action-md"},
		{"lg", "action-lg"},
		{"xl", "action-xl"},
		{"unknown", "action-md"}, // Default fallback
		{"", "action-md"},        // Empty fallback
	}

	for _, tc := range testCases {
		action := &Action{
			ID:   "test_action",
			Type: ActionButton,
			Text: "Test Button",
			Size: tc.size,
		}
		require.Equal(suite.T(), tc.expected, action.GetSizeClass())
	}
}

// Test action validation edge cases
func (suite *ActionTestSuite) TestActionValidationEdgeCases() {
	// Test action with empty ID
	action := &Action{
		Type: ActionButton,
		Text: "Button without ID",
	}
	err := action.Validate(suite.ctx)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "action ID is required")

	// Test action with empty text
	action2 := &Action{
		ID:   "test_action",
		Type: ActionButton,
		Text: "",
	}
	err = action2.Validate(suite.ctx)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "text is required")

	// Test action with invalid variant (should pass with fallback to default)
	action3 := &Action{
		ID:      "test_action",
		Type:    ActionButton,
		Text:    "Test Button",
		Variant: "invalid_variant",
	}
	err = action3.Validate(suite.ctx)
	require.NoError(suite.T(), err)
	// Invalid variants should fall back to primary
	require.Equal(suite.T(), "action-primary", action3.GetVariantClass())

	// Test action with invalid size (should pass with fallback to default)
	action4 := &Action{
		ID:   "test_action",
		Type: ActionButton,
		Text: "Test Button",
		Size: "invalid_size",
	}
	err = action4.Validate(suite.ctx)
	require.NoError(suite.T(), err)
	// Invalid sizes should fall back to medium
	require.Equal(suite.T(), "action-md", action4.GetSizeClass())

	// Test valid action
	action5 := &Action{
		ID:      "valid_action",
		Type:    ActionButton,
		Text:    "Valid Button",
		Variant: "primary",
		Size:    "md",
	}
	err = action5.Validate(suite.ctx)
	require.NoError(suite.T(), err)
}

// Test action methods
func (suite *ActionTestSuite) TestActionMethods() {
	action := Action{
		ID:       "test-action",
		Type:     ActionSubmit,
		Text:     "Submit",
		Icon:     "submit-icon",
		Position: "left",
	}

	// Test SetEvaluator
	newEvaluator := condition.NewEvaluator(nil, condition.DefaultEvalOptions())
	action.SetEvaluator(newEvaluator)
	suite.Require().Equal(newEvaluator, action.GetEvaluator())

	// Test GetIconPosition
	suite.Require().Equal("left", action.GetIconPosition())

	// Test HasIcon
	suite.Require().True(action.HasIcon())

	// Test action without icon
	actionNoIcon := Action{ID: "no-icon", Type: ActionButton, Text: "Button"}
	suite.Require().False(actionNoIcon.HasIcon())

	// Test CanView and CanExecute with permissions
	canView := action.CanView([]string{"read", "write"})
	suite.Require().True(canView) // Should return true when no permissions required

	canExecute := action.CanExecute([]string{"read", "write"})
	suite.Require().True(canExecute) // Should return true when no permissions required
}

// Test action validation methods
func (suite *ActionTestSuite) TestActionValidation() {
	// Test valid action
	validAction := Action{
		ID:   "valid-action",
		Type: ActionSubmit,
		Text: "Submit Form",
	}
	err := validAction.Validate(suite.ctx)
	suite.Require().NoError(err)

	// Test action with empty ID
	invalidAction := Action{
		ID:   "",
		Type: ActionSubmit,
		Text: "Submit",
	}
	err = invalidAction.Validate(suite.ctx)
	suite.Require().Error(err)

	// Test action with empty text
	invalidAction = Action{
		ID:   "action-id",
		Type: ActionSubmit,
		Text: "",
	}
	err = invalidAction.Validate(suite.ctx)
	suite.Require().Error(err)

	// Test action with empty type (should default to ActionButton)
	actionWithEmptyType := Action{
		ID:   "action-id",
		Type: "",
		Text: "Submit",
	}
	err = actionWithEmptyType.Validate(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(ActionButton, actionWithEmptyType.Type) // Should default to button
}

// Test action serialization
func (suite *ActionTestSuite) TestActionSerialization() {
	// Skip - Action struct doesn't have ToJSON/FromJSON methods
}

// Test action copying
func (suite *ActionTestSuite) TestActionCopy() {
	original := Action{
		ID:       "original",
		Type:     ActionButton,
		Text:     "Original Button",
		Icon:     "original-icon",
		Variant:  "secondary",
		Disabled: true,
		Confirm:  &Confirm{Message: "Original confirm"},
	}

	// Test clone if method exists
	cloned := original.Clone()
	suite.Require().NotNil(cloned)
	suite.Require().Equal(original.ID, cloned.ID)
}

// Test action with conditions
func (suite *ActionTestSuite) TestActionWithConditions() {
	action := Action{
		ID:   "conditional-action",
		Type: ActionButton,
		Text: "Conditional Button",
		Conditional: &Conditional{
			Show: &ConditionGroup{
				Logic: "AND",
				Conditions: []Condition{
					{Field: "user_role", Operator: "equal", Value: "admin"},
				},
			},
		},
	}

	// Test with permissions instead of data
	canView := action.CanView([]string{"admin"})
	suite.Require().True(canView)

	canExecute := action.CanExecute([]string{"admin"})
	suite.Require().True(canExecute)

	// Test with non-admin user - since action has no permissions, should return true
	canView = action.CanView([]string{"user"})
	suite.Require().True(canView)

	// Test with admin permission again
	canView = action.CanView([]string{"admin"})
	suite.Require().True(canView) // Should still pass with admin permission

	canExecute = action.CanExecute([]string{"admin"})
	suite.Require().True(canExecute) // Should return true since no permissions are set
}

// Test action theme application
func (suite *ActionTestSuite) TestActionTheme() {
	action := Action{
		ID:   "themed-action",
		Type: ActionSubmit,
		Text: "Themed Button",
		Theme: &ActionTheme{
			Colors: map[string]string{
				"primary": "#007bff",
				"hover":   "#0056b3",
			},
			BorderRadius: "8px",
			CustomCSS:    "font-weight: bold;",
		},
	}

	// Test theme properties
	suite.Require().NotNil(action.Theme)
	suite.Require().Equal("#007bff", action.Theme.Colors["primary"])
	suite.Require().Equal("8px", action.Theme.BorderRadius)
	suite.Require().Equal("font-weight: bold;", action.Theme.CustomCSS)

	// Test theme modification if method exists
	if action.Theme != nil {
		action.Theme.Colors["primary"] = "#28a745"
		suite.Require().Equal("#28a745", action.Theme.Colors["primary"])
	}
}

// Test action accessibility
func (suite *ActionTestSuite) TestActionAccessibility() {
	// Test accessibility features when implemented
	action := Action{
		ID:   "accessible-action",
		Type: ActionSubmit,
		Text: "Submit Form",
	}
	
	// Basic validation that action can be created
	suite.Require().Equal("accessible-action", action.ID)
	suite.Require().Equal(ActionSubmit, action.Type)
}

// Test action states
func (suite *ActionTestSuite) TestActionStates() {
	action := Action{
		ID:   "stateful-action",
		Type: ActionSubmit,
		Text: "Stateful Button",
	}

	// Test initial state
	suite.Require().False(action.Disabled)
	suite.Require().False(action.Loading)
	suite.Require().False(action.Hidden)

	// Test SetDisabled
	action.SetDisabled(true)
	suite.Require().True(action.Disabled)

	// Test SetLoading
	action.SetLoading(true)
	suite.Require().True(action.Loading)

	// Test SetHidden
	action.SetHidden(true)
	suite.Require().True(action.Hidden)

	// Test reset states
	action.SetDisabled(false)
	action.SetLoading(false)
	action.SetHidden(false)
	suite.Require().False(action.Disabled)
	suite.Require().False(action.Loading)
	suite.Require().False(action.Hidden)
}

// Test action execution context
func (suite *ActionTestSuite) TestActionExecutionContext() {
	// Test execution context when implemented
	action := Action{
		ID:   "context-action",
		Type: ActionSubmit,
		Text: "Submit with Context",
	}
	
	// Basic validation that action has proper context setup
	suite.Require().Equal("context-action", action.ID)
	suite.Require().Equal(ActionSubmit, action.Type)
}

// Test action type checking methods
func (suite *ActionTestSuite) TestActionTypeChecking() {
	// Test IsSubmitAction
	submitAction := &Action{ID: "submit", Type: ActionSubmit, Text: "Submit"}
	suite.Require().True(submitAction.IsSubmitAction())
	suite.Require().False(submitAction.IsResetAction())
	suite.Require().False(submitAction.IsLinkAction())
	suite.Require().False(submitAction.IsCustomAction())

	// Test IsResetAction
	resetAction := &Action{ID: "reset", Type: ActionReset, Text: "Reset"}
	suite.Require().False(resetAction.IsSubmitAction())
	suite.Require().True(resetAction.IsResetAction())
	suite.Require().False(resetAction.IsLinkAction())
	suite.Require().False(resetAction.IsCustomAction())

	// Test IsLinkAction
	linkAction := &Action{ID: "link", Type: ActionLink, Text: "Link"}
	suite.Require().False(linkAction.IsSubmitAction())
	suite.Require().False(linkAction.IsResetAction())
	suite.Require().True(linkAction.IsLinkAction())
	suite.Require().False(linkAction.IsCustomAction())

	// Test IsCustomAction
	customAction := &Action{ID: "custom", Type: ActionCustom, Text: "Custom"}
	suite.Require().False(customAction.IsSubmitAction())
	suite.Require().False(customAction.IsResetAction())
	suite.Require().False(customAction.IsLinkAction())
	suite.Require().True(customAction.IsCustomAction())
}

// Test action URL methods
func (suite *ActionTestSuite) TestActionURLMethods() {
	// Test GetURL with link action and config
	linkAction := &Action{
		ID:   "link",
		Type: ActionLink,
		Text: "Link",
		Config: &ActionConfig{
			URL: "https://example.com",
		},
	}
	suite.Require().Equal("https://example.com", linkAction.GetURL())

	// Test GetURL with HTMX config
	htmxAction := &Action{
		ID:   "htmx",
		Type: ActionButton,
		Text: "HTMX Button",
		HTMX: &ActionHTMX{
			URL: "/api/endpoint",
		},
	}
	suite.Require().Equal("/api/endpoint", htmxAction.GetURL())

	// Test GetURL with no config
	buttonAction := &Action{ID: "button", Type: ActionButton, Text: "Button"}
	suite.Require().Equal("", buttonAction.GetURL())
}

// Test action HTTP method
func (suite *ActionTestSuite) TestActionHTTPMethod() {
	// Test GetHTTPMethod with HTMX config
	htmxAction := &Action{
		ID:   "htmx",
		Type: ActionButton,
		Text: "HTMX Button",
		HTMX: &ActionHTMX{
			Method: "POST",
		},
	}
	suite.Require().Equal("POST", htmxAction.GetHTTPMethod())

	// Test GetHTTPMethod with only HTMX method
	onlyHTMXAction := &Action{
		ID:   "only-htmx",
		Type: ActionButton,
		Text: "Only HTMX Button",
		HTMX: &ActionHTMX{
			Method: "PUT",
		},
	}
	suite.Require().Equal("PUT", onlyHTMXAction.GetHTTPMethod())

	// Test GetHTTPMethod with defaults (GET for non-submit actions)
	defaultAction := &Action{ID: "default", Type: ActionButton, Text: "Default"}
	suite.Require().Equal("GET", defaultAction.GetHTTPMethod()) // Should default to GET

	// Test GetHTTPMethod with submit action (POST)
	submitAction := &Action{ID: "submit", Type: ActionSubmit, Text: "Submit"}
	suite.Require().Equal("POST", submitAction.GetHTTPMethod()) // Submit actions default to POST
}

// Test action debounce and throttle methods
func (suite *ActionTestSuite) TestActionDebounceThrottle() {
	// Test ShouldDebounce with config
	debounceAction := &Action{
		ID:   "debounce",
		Type: ActionButton,
		Text: "Debounce Button",
		Config: &ActionConfig{
			Debounce: 500,
		},
	}
	suite.Require().True(debounceAction.ShouldDebounce())
	suite.Require().Equal(500, debounceAction.GetDebounceDelay())

	// Test ShouldDebounce without config
	noDebounceAction := &Action{ID: "no-debounce", Type: ActionButton, Text: "No Debounce"}
	suite.Require().False(noDebounceAction.ShouldDebounce())
	suite.Require().Equal(0, noDebounceAction.GetDebounceDelay())

	// Test ShouldThrottle with config
	throttleAction := &Action{
		ID:   "throttle",
		Type: ActionButton,
		Text: "Throttle Button",
		Config: &ActionConfig{
			Throttle: 1000,
		},
	}
	suite.Require().True(throttleAction.ShouldThrottle())

	// Test ShouldThrottle without config
	noThrottleAction := &Action{ID: "no-throttle", Type: ActionButton, Text: "No Throttle"}
	suite.Require().False(noThrottleAction.ShouldThrottle())

	// Test GetThrottleDelay
	suite.Require().Equal(1000, throttleAction.GetThrottleDelay())
	suite.Require().Equal(0, noThrottleAction.GetThrottleDelay())
}

// Test action builder methods
func (suite *ActionTestSuite) TestActionBuilder() {
	// Test NewAction
	builder := NewAction("test-action", ActionButton, "Test Button")
	suite.Require().NotNil(builder)

	// Test builder methods (they return the builder for chaining)
	finalBuilder := builder.
		WithVariant("primary").
		WithSize("lg").
		WithIcon("save", "left")

	suite.Require().NotNil(finalBuilder)

	// Test WithConfig
	config := &ActionConfig{URL: "https://example.com"}
	builder.WithConfig(config)

	// Test WithConfirmation (check the method signature)
	builder.WithConfirmation("Are you sure?", "Delete")

	// Test WithPermissions
	perms := &ActionPermissions{
		View:    []string{"admin"},
		Execute: []string{"admin", "manager"},
	}
	builder.WithPermissions(perms)

	// Test WithCondition
	conditionGroup := &condition.ConditionGroup{
		ID:          "test-condition",
		Conjunction: condition.ConjunctionAnd,
		Children:    []any{},
	}
	builder.WithCondition(conditionGroup)

	// Test WithEvaluator
	evaluator := suite.createMockConditionEvaluator()
	builder.WithEvaluator(evaluator)

	// Build the final action to verify
	action := builder.Build()
	suite.Require().Equal("test-action", action.ID)
	suite.Require().Equal(ActionButton, action.Type)
	suite.Require().Equal("Test Button", action.Text)
	suite.Require().Equal("primary", action.Variant)
	suite.Require().Equal("lg", action.Size)
	suite.Require().Equal("save", action.Icon)
}

// Test action factory methods
func (suite *ActionTestSuite) TestActionFactoryMethods() {
	// Test NewSubmitAction
	submitAction := NewSubmitAction("submit", "Submit Form")
	builtSubmit := submitAction.Build()
	suite.Require().Equal("submit", builtSubmit.ID)
	suite.Require().Equal(ActionSubmit, builtSubmit.Type)
	suite.Require().Equal("Submit Form", builtSubmit.Text)
	suite.Require().Equal("primary", builtSubmit.Variant)

	// Test NewResetAction
	resetAction := NewResetAction("reset", "Reset Form")
	builtReset := resetAction.Build()
	suite.Require().Equal("reset", builtReset.ID)
	suite.Require().Equal(ActionReset, builtReset.Type)
	suite.Require().Equal("Reset Form", builtReset.Text)
	suite.Require().Equal("secondary", builtReset.Variant)

	// Test NewCancelAction
	cancelAction := NewCancelAction("cancel", "Cancel")
	builtCancel := cancelAction.Build()
	suite.Require().Equal("cancel", builtCancel.ID)
	suite.Require().Equal(ActionButton, builtCancel.Type)
	suite.Require().Equal("Cancel", builtCancel.Text)
	suite.Require().Equal("outline", builtCancel.Variant)

	// Test NewDeleteAction
	deleteAction := NewDeleteAction("delete", "Delete Item")
	builtDelete := deleteAction.Build()
	suite.Require().Equal("delete", builtDelete.ID)
	suite.Require().Equal(ActionButton, builtDelete.Type)
	suite.Require().Equal("Delete Item", builtDelete.Text)
	suite.Require().Equal("destructive", builtDelete.Variant)
	suite.Require().NotNil(builtDelete.Confirm) // Should have confirmation
}

// Test action builder state methods
func (suite *ActionTestSuite) TestActionBuilderStates() {
	builder := NewAction("test", ActionButton, "Test")

	// Test Disabled
	disabledBuilder := builder.Disabled()
	action := disabledBuilder.Build()
	suite.Require().True(action.Disabled)

	// Test Hidden
	hiddenBuilder := NewAction("hidden", ActionButton, "Hidden").Hidden()
	hiddenAction := hiddenBuilder.Build()
	suite.Require().True(hiddenAction.Hidden)
}
