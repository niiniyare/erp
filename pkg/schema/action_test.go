package schema

import (
	"context"
	"testing"

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
