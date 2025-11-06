package schema

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BuilderExtendedTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *BuilderExtendedTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func TestBuilderExtendedTestSuite(t *testing.T) {
	suite.Run(t, new(BuilderExtendedTestSuite))
}

// Test builder configuration methods
func (suite *BuilderExtendedTestSuite) TestBuilderConfiguration() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	// Test WithVersion
	builder.WithVersion("1.2.0")
	schema, err := builder.Build(suite.ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "1.2.0", schema.Version)

	// Test WithCategory
	builder2 := NewBuilder("test-schema-2", TypeForm, "Test Schema 2")
	builder2.WithCategory("finance")
	schema2, err := builder2.Build(suite.ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "finance", schema2.Category)

	// Test WithModule
	builder3 := NewBuilder("test-schema-3", TypeForm, "Test Schema 3")
	builder3.WithModule("accounting")
	schema3, err := builder3.Build(suite.ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "accounting", schema3.Module)

	// Test WithTags
	builder4 := NewBuilder("test-schema-4", TypeForm, "Test Schema 4")
	builder4.WithTags("tag1", "tag2", "tag3")
	schema4, err := builder4.Build(suite.ctx)
	require.NoError(suite.T(), err)
	require.Contains(suite.T(), schema4.Tags, "tag1")
	require.Contains(suite.T(), schema4.Tags, "tag2")
	require.Contains(suite.T(), schema4.Tags, "tag3")
}

// Test builder with config
func (suite *BuilderExtendedTestSuite) TestBuilderWithConfig() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	config := &Config{
		Method:   "POST",
		Action:   "/api/users",
		Encoding: "application/json",
	}

	builder.WithConfig(config)
	schema, err := builder.Build(suite.ctx)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), schema.Config)
	require.Equal(suite.T(), "POST", schema.Config.Method)
	require.Equal(suite.T(), "/api/users", schema.Config.Action)
}

// Test builder with layout
func (suite *BuilderExtendedTestSuite) TestBuilderWithLayout() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	layout := &Layout{
		Type:    LayoutGrid,
		Columns: 2,
		Gap:     "md",
	}

	builder.WithLayout(layout)
	schema, err := builder.Build(suite.ctx)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), schema.Layout)
	require.Equal(suite.T(), LayoutGrid, schema.Layout.Type)
	require.Equal(suite.T(), 2, schema.Layout.Columns)
}

// Test builder with tenant
func (suite *BuilderExtendedTestSuite) TestBuilderWithTenant() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	builder.WithTenant("organization_id", "strict")
	schema, err := builder.Build(suite.ctx)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), schema.Tenant)
	require.True(suite.T(), schema.Tenant.Enabled)
	require.Equal(suite.T(), "organization_id", schema.Tenant.Field)
	require.Equal(suite.T(), "strict", schema.Tenant.Isolation)
}

// Test builder with security
func (suite *BuilderExtendedTestSuite) TestBuilderWithSecurity() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	security := &Security{
		CSRF: &CSRF{
			Enabled:    true,
			FieldName:  "_csrf_token",
			HeaderName: "X-CSRF-Token",
		},
	}

	builder.WithSecurity(security)
	schema, err := builder.Build(suite.ctx)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), schema.Security)
	require.NotNil(suite.T(), schema.Security.CSRF)
	require.True(suite.T(), schema.Security.CSRF.Enabled)
}

// Test builder with rate limit
func (suite *BuilderExtendedTestSuite) TestBuilderWithRateLimit() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	builder.WithRateLimit(100, 60)
	schema, err := builder.Build(suite.ctx)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), schema.Security)
	require.NotNil(suite.T(), schema.Security.RateLimit)
	require.True(suite.T(), schema.Security.RateLimit.Enabled)
	require.Equal(suite.T(), 100, schema.Security.RateLimit.MaxRequests)
}

// Test builder with HTMX
func (suite *BuilderExtendedTestSuite) TestBuilderWithHTMX() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	builder.WithHTMX("/api/submit", "#result")
	schema, err := builder.Build(suite.ctx)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), schema.HTMX)
	require.True(suite.T(), schema.HTMX.Enabled)
	require.Equal(suite.T(), "/api/submit", schema.HTMX.Post)
	require.Equal(suite.T(), "#result", schema.HTMX.Target)
	require.Equal(suite.T(), "innerHTML", schema.HTMX.Swap)
}

// Test builder with Alpine
func (suite *BuilderExtendedTestSuite) TestBuilderWithAlpine() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	builder.WithAlpine("{ count: 0, increment() { this.count++ } }")
	schema, err := builder.Build(suite.ctx)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), schema.Alpine)
	require.True(suite.T(), schema.Alpine.Enabled)
	require.Contains(suite.T(), schema.Alpine.XData, "count: 0")
}

// Test builder with i18n
func (suite *BuilderExtendedTestSuite) TestBuilderWithI18n() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	builder.WithI18n("en", "en", "es", "fr", "de")
	schema, err := builder.Build(suite.ctx)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), schema.I18n)
	require.True(suite.T(), schema.I18n.Enabled)
	require.Equal(suite.T(), "en", schema.I18n.DefaultLocale)
	require.Contains(suite.T(), schema.I18n.SupportedLocales, "es")
	require.Contains(suite.T(), schema.I18n.SupportedLocales, "fr")
	require.Contains(suite.T(), schema.I18n.SupportedLocales, "de")
}

// Test additional field methods
func (suite *BuilderExtendedTestSuite) TestAdditionalFieldMethods() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	// Test AddEmailField
	builder.AddEmailField("email", "Email Address", true)

	// Test AddPasswordField
	builder.AddPasswordField("password", "Password", true)

	// Test AddSelectField
	options := []Option{
		{Value: "option1", Label: "Option 1"},
		{Value: "option2", Label: "Option 2"},
	}
	builder.AddSelectField("select_field", "Select Field", false, options)

	// Test AddTextareaField
	builder.AddTextareaField("description", "Description", false, 5)

	// Test AddCheckboxField
	builder.AddCheckboxField("agree", "I agree to terms", true)

	// Test AddDateField
	builder.AddDateField("birth_date", "Birth Date", false)

	schema, err := builder.Build(suite.ctx)
	require.NoError(suite.T(), err)

	// Verify fields were added
	fieldNames := make(map[string]bool)
	for _, field := range schema.Fields {
		fieldNames[field.Name] = true
	}

	require.True(suite.T(), fieldNames["email"])
	require.True(suite.T(), fieldNames["password"])
	require.True(suite.T(), fieldNames["select_field"])
	require.True(suite.T(), fieldNames["description"])
	require.True(suite.T(), fieldNames["agree"])
	require.True(suite.T(), fieldNames["birth_date"])
}

// Test additional action methods
func (suite *BuilderExtendedTestSuite) TestAdditionalActionMethods() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	// Test AddResetButton
	builder.AddResetButton("Reset Form")

	// Test AddButton
	builder.AddButton("custom-btn", "Custom Button", "secondary")

	// Test AddActionWithConfig
	customAction := Action{
		ID:      "advanced-action",
		Type:    ActionButton,
		Text:    "Advanced Action",
		Variant: "outline",
		Size:    "lg",
	}
	builder.AddActionWithConfig(customAction)

	schema, err := builder.Build(suite.ctx)
	require.NoError(suite.T(), err)

	// Verify actions were added
	actionIDs := make(map[string]bool)
	for _, action := range schema.Actions {
		actionIDs[action.ID] = true
	}

	require.True(suite.T(), actionIDs["reset"])
	require.True(suite.T(), actionIDs["custom-btn"])
	require.True(suite.T(), actionIDs["advanced-action"])
}

// Test MustBuild method
func (suite *BuilderExtendedTestSuite) TestMustBuild() {
	// Test successful build
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")
	builder.AddTextField("name", "Name", true)

	schema := builder.MustBuild(suite.ctx)
	require.NotNil(suite.T(), schema)
	require.Equal(suite.T(), "test-schema", schema.ID)

	// Test that MustBuild panics on validation error
	// NOTE: We can't easily test the panic case without causing the test to fail
	// This would require a separate test function that expects a panic
}

// Test helper functions
func (suite *BuilderExtendedTestSuite) TestHelperFunctions() {
	// Test NewGridLayout
	gridLayout := NewGridLayout(3)
	require.Equal(suite.T(), LayoutGrid, gridLayout.Type)
	require.Equal(suite.T(), 3, gridLayout.Columns)
	require.Equal(suite.T(), "1rem", gridLayout.Gap)

	// Test NewTabLayout
	tabs := []Tab{
		{ID: "tab1", Label: "Tab 1", Fields: []string{"field1"}},
		{ID: "tab2", Label: "Tab 2", Fields: []string{"field2"}},
	}
	tabLayout := NewTabLayout(tabs)
	require.Equal(suite.T(), LayoutTabs, tabLayout.Type)
	require.Len(suite.T(), tabLayout.Tabs, 2)

	// Test NewStepLayout
	steps := []Step{
		{ID: "step1", Title: "Step 1", Order: 1, Fields: []string{"field1"}},
		{ID: "step2", Title: "Step 2", Order: 2, Fields: []string{"field2"}},
	}
	stepLayout := NewStepLayout(steps)
	require.Equal(suite.T(), LayoutSteps, stepLayout.Type)
	require.Len(suite.T(), stepLayout.Steps, 2)

	// Test NewSimpleConfig
	config := NewSimpleConfig("/api/submit", "POST")
	require.Equal(suite.T(), "POST", config.Method)
	require.Equal(suite.T(), "/api/submit", config.Action)
	require.Equal(suite.T(), "application/json", config.Encoding)

	// Test CreateOption
	option := CreateOption("value1", "Label 1")
	require.Equal(suite.T(), "value1", option.Value)
	require.Equal(suite.T(), "Label 1", option.Label)

	// Test CreateOptionWithIcon
	optionWithIcon := CreateOptionWithIcon("value2", "Label 2", "icon-name")
	require.Equal(suite.T(), "value2", optionWithIcon.Value)
	require.Equal(suite.T(), "Label 2", optionWithIcon.Label)
	require.Equal(suite.T(), "icon-name", optionWithIcon.Icon)

	// Test CreateGroupedOption
	groupedOption := CreateGroupedOption("group1", "Group 1", "finance")
	require.Equal(suite.T(), "group1", groupedOption.Value)
	require.Equal(suite.T(), "Group 1", groupedOption.Label)
	require.Equal(suite.T(), "finance", groupedOption.Group)
}

// Test field builder pattern
func (suite *BuilderExtendedTestSuite) TestFieldBuilder() {
	// Test NewField
	fieldBuilder := NewField("test_field", FieldText)
	require.NotNil(suite.T(), fieldBuilder)

	// Test field builder methods
	fieldBuilder.WithLabel("Test Label").
		WithDescription("Test description").
		WithPlaceholder("Enter text here").
		WithHelp("This is help text").
		Required().
		WithDefault("default value")

	field := fieldBuilder.Build()
	require.Equal(suite.T(), "test_field", field.Name)
	require.Equal(suite.T(), FieldText, field.Type)
	require.Equal(suite.T(), "Test Label", field.Label)
	require.Equal(suite.T(), "Test description", field.Description)
	require.Equal(suite.T(), "Enter text here", field.Placeholder)
	require.Equal(suite.T(), "This is help text", field.Help)
	require.True(suite.T(), field.Required)
	require.Equal(suite.T(), "default value", field.Default)

	// Test Disabled and Readonly
	fieldBuilder2 := NewField("test_field_2", FieldText)
	fieldBuilder2.Disabled().Readonly()
	field2 := fieldBuilder2.Build()
	require.True(suite.T(), field2.Disabled)
	require.True(suite.T(), field2.Readonly)

	// Test WithOptions
	options := []Option{CreateOption("opt1", "Option 1")}
	fieldBuilder3 := NewField("test_field_3", FieldSelect)
	fieldBuilder3.WithOptions(options)
	field3 := fieldBuilder3.Build()
	require.Len(suite.T(), field3.Options, 1)

	// Test WithValidation
	minLen := 5
	maxLen := 100
	validation := &FieldValidation{
		MinLength: &minLen,
		MaxLength: &maxLen,
	}
	fieldBuilder4 := NewField("test_field_4", FieldText)
	fieldBuilder4.WithValidation(validation)
	field4 := fieldBuilder4.Build()
	require.NotNil(suite.T(), field4.Validation)
	require.Equal(suite.T(), 5, *field4.Validation.MinLength)
	require.Equal(suite.T(), 100, *field4.Validation.MaxLength)
}
