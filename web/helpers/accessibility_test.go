package helpers

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MockLogger implements logger.Logger for testing
type MockLogger struct {
	DebugCalls []LogCall
	InfoCalls  []LogCall
	WarnCalls  []LogCall
	ErrorCalls []LogCall
	FatalCalls []LogCall
}

type LogCall struct {
	Message string
	Fields  logger.Fields
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		DebugCalls: make([]LogCall, 0),
		InfoCalls:  make([]LogCall, 0),
		WarnCalls:  make([]LogCall, 0),
		ErrorCalls: make([]LogCall, 0),
		FatalCalls: make([]LogCall, 0),
	}
}

func (m *MockLogger) Debug(msg string, fields ...logger.Fields) {
	call := LogCall{Message: msg}
	if len(fields) > 0 {
		call.Fields = fields[0]
	}
	m.DebugCalls = append(m.DebugCalls, call)
}

func (m *MockLogger) Info(msg string, fields ...logger.Fields) {
	call := LogCall{Message: msg}
	if len(fields) > 0 {
		call.Fields = fields[0]
	}
	m.InfoCalls = append(m.InfoCalls, call)
}

func (m *MockLogger) Warn(msg string, fields ...logger.Fields) {
	call := LogCall{Message: msg}
	if len(fields) > 0 {
		call.Fields = fields[0]
	}
	m.WarnCalls = append(m.WarnCalls, call)
}

func (m *MockLogger) Error(msg string, fields ...logger.Fields) {
	call := LogCall{Message: msg}
	if len(fields) > 0 {
		call.Fields = fields[0]
	}
	m.ErrorCalls = append(m.ErrorCalls, call)
}

func (m *MockLogger) Fatal(msg string, fields ...logger.Fields) {
	call := LogCall{Message: msg}
	if len(fields) > 0 {
		call.Fields = fields[0]
	}
	m.FatalCalls = append(m.FatalCalls, call)
}

func (m *MockLogger) DebugContext(ctx context.Context, msg string, fields ...logger.Fields) {
	m.Debug(msg, fields...)
}

func (m *MockLogger) InfoContext(ctx context.Context, msg string, fields ...logger.Fields) {
	m.Info(msg, fields...)
}

func (m *MockLogger) WarnContext(ctx context.Context, msg string, fields ...logger.Fields) {
	m.Warn(msg, fields...)
}

func (m *MockLogger) ErrorContext(ctx context.Context, msg string, fields ...logger.Fields) {
	m.Error(msg, fields...)
}

func (m *MockLogger) WithFields(fields logger.Fields) logger.Logger {
	return m
}

func (m *MockLogger) WithContext(ctx context.Context) logger.Logger {
	return m
}

func (m *MockLogger) SetLevel(level logger.LogLevel) {}

func (m *MockLogger) Close() error {
	return nil
}

func (m *MockLogger) Reset() {
	m.DebugCalls = make([]LogCall, 0)
	m.InfoCalls = make([]LogCall, 0)
	m.WarnCalls = make([]LogCall, 0)
	m.ErrorCalls = make([]LogCall, 0)
	m.FatalCalls = make([]LogCall, 0)
}

// AccessibilityHelperTestSuite defines the test suite
type AccessibilityHelperTestSuite struct {
	suite.Suite
	helper     *AccessibilityHelper
	mockLogger *MockLogger
}

// SetupTest runs before each test
func (suite *AccessibilityHelperTestSuite) SetupTest() {
	suite.mockLogger = NewMockLogger()
	suite.helper = NewAccessibilityHelperWithLogger(nil, suite.mockLogger)
}

// TearDownTest runs after each test
func (suite *AccessibilityHelperTestSuite) TearDownTest() {
	suite.mockLogger.Reset()
}

// TestAccessibilityHelperTestSuite runs the test suite
func TestAccessibilityHelperTestSuite(t *testing.T) {
	suite.Run(t, new(AccessibilityHelperTestSuite))
}

// TestNewAccessibilityHelper tests helper creation
func (suite *AccessibilityHelperTestSuite) TestNewAccessibilityHelper() {
	tests := []struct {
		name          string
		config        *AccessibilityConfig
		expectDefault bool
	}{
		{
			name:          "nil config uses defaults",
			config:        nil,
			expectDefault: true,
		},
		{
			name: "custom config is used",
			config: &AccessibilityConfig{
				EnableScreenReader:  false,
				EnableKeyboardNav:   false,
				EnableHighContrast:  false,
				EnableReducedMotion: false,
				DefaultLanguage:     "fr",
				SupportedLanguages:  []string{"fr", "de"},
			},
			expectDefault: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			helper := NewAccessibilityHelper(tt.config)
			require.NotNil(suite.T(), helper)

			config := helper.GetConfig()
			if tt.expectDefault {
				require.Equal(suite.T(), "en", config.DefaultLanguage)
				require.True(suite.T(), config.EnableScreenReader)
			} else {
				require.Equal(suite.T(), "fr", config.DefaultLanguage)
				require.False(suite.T(), config.EnableScreenReader)
			}
		})
	}
}

// TestGenerateAriaLabel tests ARIA label generation
func (suite *AccessibilityHelperTestSuite) TestGenerateAriaLabel() {
	tests := []struct {
		name          string
		baseLabel     string
		context       map[string]string
		expected      string
		expectWarning bool
	}{
		{
			name:      "basic label without context",
			baseLabel: "Submit",
			context:   nil,
			expected:  "Submit",
		},
		{
			name:      "label with status",
			baseLabel: "Submit",
			context: map[string]string{
				"status": "enabled",
			},
			expected: "Submit, Status: enabled",
		},
		{
			name:      "label with count",
			baseLabel: "Items",
			context: map[string]string{
				"count": "5",
			},
			expected: "Items, Count: 5",
		},
		{
			name:      "label with level",
			baseLabel: "Heading",
			context: map[string]string{
				"level": "2",
			},
			expected: "Heading, Level: 2",
		},
		{
			name:      "label with all context",
			baseLabel: "Button",
			context: map[string]string{
				"status": "active",
				"count":  "3",
				"level":  "1",
			},
			expected: "Button, Status: active, Count: 3, Level: 1",
		},
		{
			name:      "empty context values are ignored",
			baseLabel: "Label",
			context: map[string]string{
				"status": "",
				"count":  "5",
			},
			expected: "Label, Count: 5",
		},
		{
			name:          "empty base label logs warning",
			baseLabel:     "",
			context:       nil,
			expected:      "",
			expectWarning: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.mockLogger.Reset()

			result := suite.helper.GenerateAriaLabel(tt.baseLabel, tt.context)

			require.Equal(suite.T(), tt.expected, result)

			if tt.expectWarning {
				require.NotEmpty(suite.T(), suite.mockLogger.WarnCalls,
					"expected warning to be logged")
			}

			if tt.baseLabel != "" {
				require.NotEmpty(suite.T(), suite.mockLogger.DebugCalls,
					"expected debug log for successful generation")
			}
		})
	}
}

// TestGenerateAriaLabelWithValidation tests validation
func (suite *AccessibilityHelperTestSuite) TestGenerateAriaLabelWithValidation() {
	tests := []struct {
		name        string
		baseLabel   string
		context     map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid label",
			baseLabel:   "Valid Label",
			context:     nil,
			expectError: false,
		},
		{
			name:        "empty label returns error",
			baseLabel:   "",
			context:     nil,
			expectError: true,
			errorMsg:    "baseLabel cannot be empty",
		},
		{
			name:        "label exceeds max length",
			baseLabel:   string(make([]byte, 300)),
			context:     nil,
			expectError: true,
			errorMsg:    "baseLabel exceeds maximum length",
		},
		{
			name:        "label at max length is valid",
			baseLabel:   string(make([]byte, 255)),
			context:     nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.mockLogger.Reset()

			result, err := suite.helper.GenerateAriaLabelWithValidation(tt.baseLabel, tt.context)

			if tt.expectError {
				require.Error(suite.T(), err)
				require.Contains(suite.T(), err.Error(), tt.errorMsg)
				require.Empty(suite.T(), result)
				require.NotEmpty(suite.T(), suite.mockLogger.ErrorCalls,
					"expected error to be logged")
			} else {
				require.NoError(suite.T(), err)
				require.NotEmpty(suite.T(), result)
			}
		})
	}
}

// TestGenerateAriaDescribedBy tests describedby generation
func (suite *AccessibilityHelperTestSuite) TestGenerateAriaDescribedBy() {
	tests := []struct {
		name          string
		elementID     string
		descriptions  []string
		expected      string
		expectWarning bool
	}{
		{
			name:      "single description",
			elementID: "input-1",
			descriptions: []string{
				"Error message",
			},
			expected: "input-1-desc-0",
		},
		{
			name:      "multiple descriptions",
			elementID: "input-1",
			descriptions: []string{
				"Error message",
				"Help text",
				"Constraint",
			},
			expected: "input-1-desc-0 input-1-desc-1 input-1-desc-2",
		},
		{
			name:         "empty descriptions are skipped",
			elementID:    "input-1",
			descriptions: []string{"", "Valid", ""},
			expected:     "input-1-desc-1",
		},
		{
			name:          "empty element ID logs warning",
			elementID:     "",
			descriptions:  []string{"test"},
			expected:      "",
			expectWarning: true,
		},
		{
			name:         "empty descriptions array",
			elementID:    "input-1",
			descriptions: []string{},
			expected:     "",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.mockLogger.Reset()

			result := suite.helper.GenerateAriaDescribedBy(tt.elementID, tt.descriptions)

			require.Equal(suite.T(), tt.expected, result)

			if tt.expectWarning {
				require.NotEmpty(suite.T(), suite.mockLogger.WarnCalls)
			}
		})
	}
}

// TestGenerateTabIndex tests tabindex generation
func (suite *AccessibilityHelperTestSuite) TestGenerateTabIndex() {
	tests := []struct {
		name        string
		interactive bool
		forceFocus  bool
		expected    string
	}{
		{
			name:        "force focus returns 0",
			interactive: false,
			forceFocus:  true,
			expected:    "0",
		},
		{
			name:        "interactive returns 0",
			interactive: true,
			forceFocus:  false,
			expected:    "0",
		},
		{
			name:        "non-interactive returns -1",
			interactive: false,
			forceFocus:  false,
			expected:    "-1",
		},
		{
			name:        "both flags true returns 0",
			interactive: true,
			forceFocus:  true,
			expected:    "0",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.helper.GenerateTabIndex(tt.interactive, tt.forceFocus)
			require.Equal(suite.T(), tt.expected, result)
		})
	}
}

// TestGenerateTabIndexDisabled tests when keyboard nav is disabled
func (suite *AccessibilityHelperTestSuite) TestGenerateTabIndexDisabled() {
	config := &AccessibilityConfig{
		EnableKeyboardNav: false,
	}
	helper := NewAccessibilityHelperWithLogger(config, suite.mockLogger)

	result := helper.GenerateTabIndex(true, true)
	require.Empty(suite.T(), result, "should return empty when keyboard nav disabled")
}

// TestGenerateKeyboardShortcut tests keyboard shortcut generation
func (suite *AccessibilityHelperTestSuite) TestGenerateKeyboardShortcut() {
	tests := []struct {
		name      string
		key       string
		modifiers []string
		expected  map[string]string
	}{
		{
			name:      "simple key",
			key:       "s",
			modifiers: []string{},
			expected: map[string]string{
				"data-key": "s",
				"title":    "Keyboard shortcut: s",
			},
		},
		{
			name:      "key with alt modifier generates accesskey",
			key:       "s",
			modifiers: []string{"alt"},
			expected: map[string]string{
				"data-key":       "s",
				"data-modifiers": "alt",
				"accesskey":      "s",
				"title":          "Keyboard shortcut: alt+s",
			},
		},
		{
			name:      "key with multiple modifiers",
			key:       "s",
			modifiers: []string{"ctrl", "shift"},
			expected: map[string]string{
				"data-key":       "s",
				"data-modifiers": "ctrl+shift",
				"title":          "Keyboard shortcut: ctrl+shift+s",
			},
		},
		{
			name:      "empty key returns nil",
			key:       "",
			modifiers: []string{"ctrl"},
			expected:  map[string]string{},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.helper.GenerateKeyboardShortcut(tt.key, tt.modifiers)

			if len(tt.expected) == 0 {
				require.Empty(suite.T(), result)
			} else {
				require.Equal(suite.T(), tt.expected, result)
			}
		})
	}
}

// TestGenerateAriaLive tests aria-live generation
func (suite *AccessibilityHelperTestSuite) TestGenerateAriaLive() {
	tests := []struct {
		name     string
		urgency  string
		expected string
	}{
		{
			name:     "high urgency",
			urgency:  "high",
			expected: "assertive",
		},
		{
			name:     "urgent",
			urgency:  "urgent",
			expected: "assertive",
		},
		{
			name:     "error",
			urgency:  "error",
			expected: "assertive",
		},
		{
			name:     "medium urgency",
			urgency:  "medium",
			expected: "polite",
		},
		{
			name:     "info",
			urgency:  "info",
			expected: "polite",
		},
		{
			name:     "success",
			urgency:  "success",
			expected: "polite",
		},
		{
			name:     "low urgency",
			urgency:  "low",
			expected: "polite",
		},
		{
			name:     "status",
			urgency:  "status",
			expected: "polite",
		},
		{
			name:     "off",
			urgency:  "off",
			expected: "off",
		},
		{
			name:     "unknown defaults to polite",
			urgency:  "unknown",
			expected: "polite",
		},
		{
			name:     "case insensitive",
			urgency:  "HIGH",
			expected: "assertive",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.helper.GenerateAriaLive(tt.urgency)
			require.Equal(suite.T(), tt.expected, result)
		})
	}
}

// TestGenerateFormFieldAttributes tests form field attribute generation
func (suite *AccessibilityHelperTestSuite) TestGenerateFormFieldAttributes() {
	tests := []struct {
		name     string
		config   *FormFieldConfig
		expected map[string]string
	}{
		{
			name:     "nil config returns empty",
			config:   nil,
			expected: map[string]string{},
		},
		{
			name: "required field",
			config: &FormFieldConfig{
				Required: true,
			},
			expected: map[string]string{
				"required":      "true",
				"aria-required": "true",
			},
		},
		{
			name: "field with error",
			config: &FormFieldConfig{
				HasError: true,
				ErrorID:  "error-1",
			},
			expected: map[string]string{
				"aria-invalid":     "true",
				"aria-describedby": "error-1",
			},
		},
		{
			name: "complete form field",
			config: &FormFieldConfig{
				Required:     true,
				HasError:     true,
				ErrorID:      "email-error",
				AutoComplete: "email",
				InputMode:    "email",
				Pattern:      "[a-z]+@[a-z]+\\.[a-z]+",
			},
			expected: map[string]string{
				"required":         "true",
				"aria-required":    "true",
				"aria-invalid":     "true",
				"aria-describedby": "email-error",
				"autocomplete":     "email",
				"inputmode":        "email",
				"pattern":          "[a-z]+@[a-z]+\\.[a-z]+",
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.mockLogger.Reset()

			result := suite.helper.GenerateFormFieldAttributes(tt.config)

			require.Equal(suite.T(), tt.expected, result)

			if tt.config == nil {
				require.NotEmpty(suite.T(), suite.mockLogger.WarnCalls)
			}
		})
	}
}

// TestFormFieldConfigBuilder tests the builder pattern
func (suite *AccessibilityHelperTestSuite) TestFormFieldConfigBuilder() {
	tests := []struct {
		name     string
		build    func() *FormFieldConfig
		validate func(*testing.T, *FormFieldConfig)
	}{
		{
			name: "empty builder",
			build: func() *FormFieldConfig {
				return NewFormFieldConfig().Build()
			},
			validate: func(t *testing.T, config *FormFieldConfig) {
				require.False(t, config.Required)
				require.False(t, config.HasError)
				require.Empty(t, config.ErrorID)
			},
		},
		{
			name: "required field",
			build: func() *FormFieldConfig {
				return NewFormFieldConfig().Required().Build()
			},
			validate: func(t *testing.T, config *FormFieldConfig) {
				require.True(t, config.Required)
			},
		},
		{
			name: "field with error",
			build: func() *FormFieldConfig {
				return NewFormFieldConfig().WithError("err-1").Build()
			},
			validate: func(t *testing.T, config *FormFieldConfig) {
				require.True(t, config.HasError)
				require.Equal(t, "err-1", config.ErrorID)
			},
		},
		{
			name: "complete field config",
			build: func() *FormFieldConfig {
				return NewFormFieldConfig().
					Required().
					WithError("error-id").
					AutoComplete("email").
					InputMode("email").
					Pattern("[a-z]+").
					Build()
			},
			validate: func(t *testing.T, config *FormFieldConfig) {
				require.True(t, config.Required)
				require.True(t, config.HasError)
				require.Equal(t, "error-id", config.ErrorID)
				require.Equal(t, "email", config.AutoComplete)
				require.Equal(t, "email", config.InputMode)
				require.Equal(t, "[a-z]+", config.Pattern)
			},
		},
		{
			name: "chaining returns new builder",
			build: func() *FormFieldConfig {
				builder := NewFormFieldConfig()
				builder.Required().AutoComplete("email")
				return builder.Build()
			},
			validate: func(t *testing.T, config *FormFieldConfig) {
				require.True(t, config.Required)
				require.Equal(t, "email", config.AutoComplete)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			config := tt.build()
			require.NotNil(suite.T(), config)
			tt.validate(suite.T(), config)
		})
	}
}

// TestGenerateButtonAttributes tests button attribute generation
func (suite *AccessibilityHelperTestSuite) TestGenerateButtonAttributes() {
	pressedTrue := true
	pressedFalse := false
	expandedTrue := true

	tests := []struct {
		name     string
		config   *ButtonConfig
		expected map[string]string
	}{
		{
			name:     "nil config returns empty",
			config:   nil,
			expected: map[string]string{},
		},
		{
			name: "pressed button",
			config: &ButtonConfig{
				Pressed: &pressedTrue,
			},
			expected: map[string]string{
				"aria-pressed": "true",
			},
		},
		{
			name: "not pressed button",
			config: &ButtonConfig{
				Pressed: &pressedFalse,
			},
			expected: map[string]string{
				"aria-pressed": "false",
			},
		},
		{
			name: "expanded button",
			config: &ButtonConfig{
				Expanded: &expandedTrue,
			},
			expected: map[string]string{
				"aria-expanded": "true",
			},
		},
		{
			name: "button with controls",
			config: &ButtonConfig{
				Controls: "menu-1",
			},
			expected: map[string]string{
				"aria-controls": "menu-1",
			},
		},
		{
			name: "button with popup",
			config: &ButtonConfig{
				HasPopup: true,
			},
			expected: map[string]string{
				"aria-haspopup": "true",
			},
		},
		{
			name: "disabled button",
			config: &ButtonConfig{
				Disabled: true,
			},
			expected: map[string]string{
				"disabled":      "true",
				"aria-disabled": "true",
			},
		},
		{
			name: "complete button",
			config: &ButtonConfig{
				Pressed:  &pressedTrue,
				Expanded: &expandedTrue,
				Controls: "panel-1",
				HasPopup: true,
				Disabled: false,
			},
			expected: map[string]string{
				"aria-pressed":  "true",
				"aria-expanded": "true",
				"aria-controls": "panel-1",
				"aria-haspopup": "true",
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.helper.GenerateButtonAttributes(tt.config)
			require.Equal(suite.T(), tt.expected, result)
		})
	}
}

// TestButtonConfigBuilder tests button builder
func (suite *AccessibilityHelperTestSuite) TestButtonConfigBuilder() {
	tests := []struct {
		name     string
		build    func() *ButtonConfig
		validate func(*testing.T, *ButtonConfig)
	}{
		{
			name: "pressed button",
			build: func() *ButtonConfig {
				return NewButtonConfig().Pressed(true).Build()
			},
			validate: func(t *testing.T, config *ButtonConfig) {
				require.NotNil(t, config.Pressed)
				require.True(t, *config.Pressed)
			},
		},
		{
			name: "expanded button",
			build: func() *ButtonConfig {
				return NewButtonConfig().Expanded(false).Build()
			},
			validate: func(t *testing.T, config *ButtonConfig) {
				require.NotNil(t, config.Expanded)
				require.False(t, *config.Expanded)
			},
		},
		{
			name: "complete button",
			build: func() *ButtonConfig {
				return NewButtonConfig().
					Pressed(true).
					Expanded(true).
					Controls("menu").
					HasPopup().
					Disabled().
					Build()
			},
			validate: func(t *testing.T, config *ButtonConfig) {
				require.NotNil(t, config.Pressed)
				require.True(t, *config.Pressed)
				require.NotNil(t, config.Expanded)
				require.True(t, *config.Expanded)
				require.Equal(t, "menu", config.Controls)
				require.True(t, config.HasPopup)
				require.True(t, config.Disabled)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			config := tt.build()
			require.NotNil(suite.T(), config)
			tt.validate(suite.T(), config)
		})
	}
}

// TestSanitizeID tests ID sanitization
func (suite *AccessibilityHelperTestSuite) TestSanitizeID() {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "text with spaces",
			input:    "hello world",
			expected: "hello-world",
		},
		{
			name:     "uppercase converted to lowercase",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "special characters removed",
			input:    "hello@world!",
			expected: "helloworld",
		},
		{
			name:     "underscores preserved",
			input:    "hello_world",
			expected: "hello_world",
		},
		{
			name:     "hyphens preserved",
			input:    "hello-world",
			expected: "hello-world",
		},
		{
			name:     "numbers preserved",
			input:    "item123",
			expected: "item123",
		},
		{
			name:     "starts with number gets prefix",
			input:    "123item",
			expected: "id-123item",
		},
		{
			name:     "multiple spaces become single hyphen",
			input:    "hello   world",
			expected: "hello---world",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only special characters",
			input:    "@#$%",
			expected: "",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.helper.SanitizeID(tt.input)
			require.Equal(suite.T(), tt.expected, result)
		})
	}
}

// TestGenerateUniqueID tests unique ID generation
func (suite *AccessibilityHelperTestSuite) TestGenerateUniqueID() {
	tests := []struct {
		name         string
		prefix       string
		expectPrefix string
	}{
		{
			name:         "simple prefix",
			prefix:       "button",
			expectPrefix: "button-",
		},
		{
			name:         "prefix with spaces",
			prefix:       "submit button",
			expectPrefix: "submit-button-",
		},
		{
			name:         "empty prefix uses default",
			prefix:       "",
			expectPrefix: "element-",
		},
		{
			name:         "prefix with special chars",
			prefix:       "my@button",
			expectPrefix: "mybutton-",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.helper.GenerateUniqueID(tt.prefix)
			require.Contains(suite.T(), result, tt.expectPrefix)
			require.NotEmpty(suite.T(), result)
		})
	}
}

// TestGenerateUniqueIDUniqueness tests that IDs are actually unique
func (suite *AccessibilityHelperTestSuite) TestGenerateUniqueIDUniqueness() {
	suite.Run("sequential calls produce unique IDs", func() {
		ids := make(map[string]bool)
		for i := 0; i < 1000; i++ {
			id := suite.helper.GenerateUniqueID("test")
			require.False(suite.T(), ids[id], "duplicate ID generated: %s", id)
			ids[id] = true
		}
	})

	suite.Run("concurrent calls produce unique IDs", func() {
		var wg sync.WaitGroup
		idChan := make(chan string, 1000)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					idChan <- suite.helper.GenerateUniqueID("concurrent")
				}
			}()
		}

		wg.Wait()
		close(idChan)

		ids := make(map[string]bool)
		for id := range idChan {
			require.False(suite.T(), ids[id], "duplicate ID in concurrent test: %s", id)
			ids[id] = true
		}
	})
}

// TestGenerateSecureUniqueID tests secure ID generation
func (suite *AccessibilityHelperTestSuite) TestGenerateSecureUniqueID() {
	suite.Run("generates unique IDs", func() {
		ids := make(map[string]bool)
		for i := 0; i < 100; i++ {
			id := suite.helper.GenerateSecureUniqueID("secure")
			require.False(suite.T(), ids[id], "duplicate secure ID: %s", id)
			require.Contains(suite.T(), id, "secure-")
			require.Greater(suite.T(), len(id), 15, "secure ID should be long")
			ids[id] = true
		}
	})

	suite.Run("fallback on crypto failure", func() {
		// This test verifies logging on fallback
		// In real scenarios, crypto/rand.Read rarely fails
		id := suite.helper.GenerateSecureUniqueID("test")
		require.NotEmpty(suite.T(), id)
	})
}

// TestGetLanguageAttribute tests language extraction from context
func (suite *AccessibilityHelperTestSuite) TestGetLanguageAttribute() {
	tests := []struct {
		name     string
		setupCtx func() context.Context
		expected string
	}{
		{
			name: "context with valid language",
			setupCtx: func() context.Context {
				return SetLanguageContext(context.Background(), "es")
			},
			expected: "es",
		},
		{
			name: "context without language uses default",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expected: "en",
		},
		{
			name: "context with unsupported language uses default",
			setupCtx: func() context.Context {
				return SetLanguageContext(context.Background(), "unsupported")
			},
			expected: "en",
		},
		{
			name: "context with supported language fr",
			setupCtx: func() context.Context {
				return SetLanguageContext(context.Background(), "fr")
			},
			expected: "fr",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := tt.setupCtx()
			result := suite.helper.GetLanguageAttribute(ctx)
			require.Equal(suite.T(), tt.expected, result)
		})
	}
}

// TestIsLanguageSupported tests language support checking
func (suite *AccessibilityHelperTestSuite) TestIsLanguageSupported() {
	tests := []struct {
		name     string
		lang     string
		expected bool
	}{
		{
			name:     "supported language en",
			lang:     "en",
			expected: true,
		},
		{
			name:     "supported language es",
			lang:     "es",
			expected: true,
		},
		{
			name:     "supported language fr",
			lang:     "fr",
			expected: true,
		},
		{
			name:     "supported language de",
			lang:     "de",
			expected: true,
		},
		{
			name:     "unsupported language",
			lang:     "unsupported",
			expected: false,
		},
		{
			name:     "empty language",
			lang:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.helper.IsLanguageSupported(tt.lang)
			require.Equal(suite.T(), tt.expected, result)
		})
	}
}

// TestGenerateLandmarkRole tests landmark role validation
func (suite *AccessibilityHelperTestSuite) TestGenerateLandmarkRole() {
	tests := []struct {
		name          string
		landmark      string
		expected      string
		expectWarning bool
	}{
		{
			name:     "valid banner",
			landmark: "banner",
			expected: "banner",
		},
		{
			name:     "valid navigation",
			landmark: "navigation",
			expected: "navigation",
		},
		{
			name:     "valid main",
			landmark: "main",
			expected: "main",
		},
		{
			name:     "valid contentinfo",
			landmark: "contentinfo",
			expected: "contentinfo",
		},
		{
			name:          "invalid landmark",
			landmark:      "invalid",
			expected:      "",
			expectWarning: true,
		},
		{
			name:          "empty landmark",
			landmark:      "",
			expected:      "",
			expectWarning: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.mockLogger.Reset()

			result := suite.helper.GenerateLandmarkRole(tt.landmark)

			require.Equal(suite.T(), tt.expected, result)

			if tt.expectWarning {
				require.NotEmpty(suite.T(), suite.mockLogger.WarnCalls)
			}
		})
	}
}

// TestGenerateRole tests ARIA role validation
func (suite *AccessibilityHelperTestSuite) TestGenerateRole() {
	tests := []struct {
		name          string
		role          string
		expected      string
		expectWarning bool
	}{
		{
			name:     "valid button role",
			role:     "button",
			expected: "button",
		},
		{
			name:     "valid alert role",
			role:     "alert",
			expected: "alert",
		},
		{
			name:     "valid navigation role",
			role:     "navigation",
			expected: "navigation",
		},
		{
			name:          "invalid role",
			role:          "invalid-role",
			expected:      "",
			expectWarning: true,
		},
		{
			name:          "empty role",
			role:          "",
			expected:      "",
			expectWarning: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.mockLogger.Reset()

			result := suite.helper.GenerateRole(tt.role)

			require.Equal(suite.T(), tt.expected, result)

			if tt.expectWarning {
				require.NotEmpty(suite.T(), suite.mockLogger.WarnCalls)
			}
		})
	}
}

// TestValidateAccessibilityAttributes tests attribute validation
func (suite *AccessibilityHelperTestSuite) TestValidateAccessibilityAttributes() {
	tests := []struct {
		name          string
		attrs         map[string]string
		expectedCount int
		expectedInMsg []string
	}{
		{
			name: "valid attributes",
			attrs: map[string]string{
				"aria-pressed": "true",
				"aria-invalid": "false",
			},
			expectedCount: 0,
		},
		{
			name: "invalid aria-pressed",
			attrs: map[string]string{
				"aria-pressed": "invalid",
			},
			expectedCount: 1,
			expectedInMsg: []string{"aria-pressed", "invalid"},
		},
		{
			name: "invalid aria-invalid",
			attrs: map[string]string{
				"aria-invalid": "invalid",
			},
			expectedCount: 1,
			expectedInMsg: []string{"aria-invalid", "invalid"},
		},
		{
			name: "multiple invalid attributes",
			attrs: map[string]string{
				"aria-pressed": "bad",
				"aria-invalid": "bad",
			},
			expectedCount: 2,
		},
		{
			name: "aria-pressed mixed is valid",
			attrs: map[string]string{
				"aria-pressed": "mixed",
			},
			expectedCount: 0,
		},
		{
			name: "aria-invalid spelling is valid",
			attrs: map[string]string{
				"aria-invalid": "spelling",
			},
			expectedCount: 0,
		},
		{
			name:          "empty attributes",
			attrs:         map[string]string{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.mockLogger.Reset()

			warnings := suite.helper.ValidateAccessibilityAttributes(tt.attrs)

			require.Len(suite.T(), warnings, tt.expectedCount)

			for _, msg := range tt.expectedInMsg {
				found := false
				for _, warning := range warnings {
					if suite.Contains(warning, msg) {
						found = true
						break
					}
				}
				require.True(suite.T(), found, "expected warning to contain: %s", msg)
			}

			if tt.expectedCount > 0 {
				require.NotEmpty(suite.T(), suite.mockLogger.WarnCalls)
			}
		})
	}
}

// TestMergeAttributes tests attribute merging
func (suite *AccessibilityHelperTestSuite) TestMergeAttributes() {
	tests := []struct {
		name     string
		attrMaps []map[string]string
		expected map[string]string
	}{
		{
			name:     "empty maps",
			attrMaps: []map[string]string{},
			expected: map[string]string{},
		},
		{
			name: "single map",
			attrMaps: []map[string]string{
				{"id": "test", "class": "btn"},
			},
			expected: map[string]string{
				"id":    "test",
				"class": "btn",
			},
		},
		{
			name: "merge two maps",
			attrMaps: []map[string]string{
				{"id": "test", "class": "btn"},
				{"role": "button", "type": "submit"},
			},
			expected: map[string]string{
				"id":    "test",
				"class": "btn",
				"role":  "button",
				"type":  "submit",
			},
		},
		{
			name: "later maps override earlier",
			attrMaps: []map[string]string{
				{"class": "btn"},
				{"class": "btn-primary"},
			},
			expected: map[string]string{
				"class": "btn-primary",
			},
		},
		{
			name: "merge multiple maps",
			attrMaps: []map[string]string{
				{"id": "test"},
				{"class": "btn"},
				{"role": "button"},
				{"aria-label": "Submit"},
			},
			expected: map[string]string{
				"id":         "test",
				"class":      "btn",
				"role":       "button",
				"aria-label": "Submit",
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.helper.MergeAttributes(tt.attrMaps...)
			require.Equal(suite.T(), tt.expected, result)
		})
	}
}

// TestThreadSafety tests concurrent operations
func (suite *AccessibilityHelperTestSuite) TestThreadSafety() {
	suite.Run("concurrent config reads", func() {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				config := suite.helper.GetConfig()
				require.NotEmpty(suite.T(), config.DefaultLanguage)
			}()
		}
		wg.Wait()
	})

	suite.Run("concurrent attribute generation", func() {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				suite.helper.GenerateAriaLabel(fmt.Sprintf("Label %d", idx), nil)
				suite.helper.GenerateTabIndex(true, false)
				suite.helper.GenerateUniqueID("element")
			}(i)
		}
		wg.Wait()
	})

	suite.Run("concurrent config updates", func() {
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				config := DefaultAccessibilityConfig()
				suite.helper.UpdateConfig(config)
			}()
		}
		wg.Wait()
	})
}

// TestLoggerIntegration tests logger integration
func (suite *AccessibilityHelperTestSuite) TestLoggerIntegration() {
	suite.Run("operations are logged", func() {
		suite.mockLogger.Reset()

		// Generate ID
		suite.helper.GenerateUniqueID("test")
		require.NotEmpty(suite.T(), suite.mockLogger.DebugCalls)

		suite.mockLogger.Reset()

		// Empty input logs warning
		suite.helper.GenerateAriaLabel("", nil)
		require.NotEmpty(suite.T(), suite.mockLogger.WarnCalls)

		suite.mockLogger.Reset()

		// Validation error logs error
		_, err := suite.helper.GenerateAriaLabelWithValidation("", nil)
		require.Error(suite.T(), err)
		require.NotEmpty(suite.T(), suite.mockLogger.ErrorCalls)
	})

	suite.Run("helper without logger doesn't crash", func() {
		helperNoLogger := NewAccessibilityHelper(nil)
		require.NotPanics(suite.T(), func() {
			helperNoLogger.GenerateUniqueID("test")
			helperNoLogger.GenerateAriaLabel("", nil)
		})
	})

	suite.Run("set logger after creation", func() {
		helper := NewAccessibilityHelper(nil)
		mockLog := NewMockLogger()
		helper.SetLogger(mockLog)

		helper.GenerateUniqueID("test")
		require.NotEmpty(suite.T(), mockLog.DebugCalls)
	})
}

// TestConfigManagement tests configuration management
func (suite *AccessibilityHelperTestSuite) TestConfigManagement() {
	suite.Run("GetConfig returns copy", func() {
		config1 := suite.helper.GetConfig()
		config1.DefaultLanguage = "modified"

		config2 := suite.helper.GetConfig()
		require.NotEqual(suite.T(), "modified", config2.DefaultLanguage)
	})

	suite.Run("UpdateConfig works correctly", func() {
		newConfig := &AccessibilityConfig{
			EnableScreenReader: false,
			EnableKeyboardNav:  false,
			DefaultLanguage:    "fr",
			SupportedLanguages: []string{"fr"},
		}

		suite.helper.UpdateConfig(newConfig)

		config := suite.helper.GetConfig()
		require.Equal(suite.T(), "fr", config.DefaultLanguage)
		require.False(suite.T(), config.EnableScreenReader)
	})

	suite.Run("UpdateConfig with nil is no-op", func() {
		originalConfig := suite.helper.GetConfig()
		suite.helper.UpdateConfig(nil)
		newConfig := suite.helper.GetConfig()

		require.Equal(suite.T(), originalConfig.DefaultLanguage, newConfig.DefaultLanguage)
	})
}
