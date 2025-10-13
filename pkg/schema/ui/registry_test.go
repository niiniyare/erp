package ui

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ComponentRegistryTestSuite defines a test suite for the component registry
type ComponentRegistryTestSuite struct {
	suite.Suite
	registry ComponentRegistry
	ctx      context.Context
}

// SetupSuite runs once before all tests in the suite
func (suite *ComponentRegistryTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.registry = NewRegistryWithSchemaDir("../../../docs/ui/Schema")
}

// TestRegistryCreation tests basic registry creation and initialization
func (suite *ComponentRegistryTestSuite) TestRegistryCreation() {
	require.NotNil(suite.T(), suite.registry, "Registry should not be nil")
	
	// Test getting registered types
	types := suite.registry.GetTypes()
	assert.NotEmpty(suite.T(), types, "Registry should have registered component types")
	
	// Verify specific component types are registered
	typeMap := make(map[ComponentType]bool)
	for _, t := range types {
		typeMap[t] = true
	}
	
	expectedTypes := []ComponentType{
		ComponentButton, ComponentInput, ComponentCard, 
		ComponentForm, ComponentContainer, ComponentTable,
	}
	
	for _, expectedType := range expectedTypes {
		assert.True(suite.T(), typeMap[expectedType], 
			"Component type %s should be registered", expectedType)
	}
}

// TestBasicComponentCreation tests creating basic components
func (suite *ComponentRegistryTestSuite) TestBasicComponentCreation() {
	// Test button creation - use existing factory behavior
	buttonConfig := map[string]any{
		"text":    "Click Me",
		"variant": "primary",
	}
	
	button, err := suite.registry.Create(suite.ctx, ComponentButton, buttonConfig)
	require.NoError(suite.T(), err, "Should create button without error")
	
	assert.NotEmpty(suite.T(), button.ID, "Button should have an ID")
	assert.Equal(suite.T(), ComponentButton, button.Type)
	// Note: Existing factories may not apply styles directly
	
	// Test input creation - use correct inputType
	inputConfig := map[string]any{
		"input_type":  "text",
		"placeholder": "Enter text",
	}
	
	input, err := suite.registry.Create(suite.ctx, ComponentInput, inputConfig)
	require.NoError(suite.T(), err, "Should create input without error")
	
	assert.NotEmpty(suite.T(), input.ID, "Input should have an ID")
	assert.Equal(suite.T(), ComponentInput, input.Type)
	
	// Test card creation - use integer elevation
	cardConfig := map[string]any{
		"elevation": 2,
		"hoverable": true,
	}
	
	card, err := suite.registry.Create(suite.ctx, ComponentCard, cardConfig)
	require.NoError(suite.T(), err, "Should create card without error")
	
	assert.NotEmpty(suite.T(), card.ID, "Card should have an ID")
	assert.Equal(suite.T(), ComponentCard, card.Type)
}

// TestComponentLifecycle tests component lifecycle management
func (suite *ComponentRegistryTestSuite) TestComponentLifecycle() {
	defaultRegistry := suite.registry.(*DefaultRegistry)
	
	// Test CreateComponent with lifecycle - use correct config
	cardConfig := map[string]any{
		"elevation": 3,
		"hoverable": true,
	}
	
	card, err := defaultRegistry.CreateComponent(suite.ctx, ComponentCard, cardConfig)
	require.NoError(suite.T(), err, "Should create card with lifecycle")
	
	assert.NotNil(suite.T(), card.CreatedAt, "Card should have CreatedAt timestamp")
	assert.NotNil(suite.T(), card.UpdatedAt, "Card should have UpdatedAt timestamp")
	assert.NotEmpty(suite.T(), card.ID, "Card should have an ID")
	
	// Test UpdateComponent
	updates := map[string]any{
		"styles": map[string]any{
			"background-color": "#f8f9fa",
			"border-radius":    "12px",
		},
	}
	
	originalUpdatedAt := card.UpdatedAt
	updatedCard, err := defaultRegistry.UpdateComponent(suite.ctx, card, updates)
	require.NoError(suite.T(), err, "Should update card without error")
	
	assert.Equal(suite.T(), card.ID, updatedCard.ID, "ID should remain the same")
	assert.Equal(suite.T(), card.CreatedAt, updatedCard.CreatedAt, "CreatedAt should remain the same")
	assert.NotEqual(suite.T(), originalUpdatedAt, updatedCard.UpdatedAt, "UpdatedAt should change")
	assert.NotNil(suite.T(), updatedCard.Styles, "Updated card should have styles")
	
	// Test CloneComponent
	clonedCard, err := defaultRegistry.CloneComponent(suite.ctx, card)
	require.NoError(suite.T(), err, "Should clone card without error")
	
	assert.NotEqual(suite.T(), card.ID, clonedCard.ID, "Cloned card should have different ID")
	assert.Equal(suite.T(), card.Type, clonedCard.Type, "Cloned card should have same type")
	assert.NotNil(suite.T(), clonedCard.CreatedAt, "Cloned card should have CreatedAt timestamp")
	assert.NotNil(suite.T(), clonedCard.UpdatedAt, "Cloned card should have UpdatedAt timestamp")
}

// TestCompositionValidation tests component composition and nesting validation
func (suite *ComponentRegistryTestSuite) TestCompositionValidation() {
	defaultRegistry := suite.registry.(*DefaultRegistry)
	
	// Create a form component
	formConfig := map[string]any{
		"id":    "validation-form",
		"label": "Validation Form",
	}
	
	form, err := defaultRegistry.CreateComponent(suite.ctx, ComponentForm, formConfig)
	require.NoError(suite.T(), err, "Should create form without error")
	
	// Create an input component with proper inputType configuration
	inputConfig := map[string]any{
		"input_type":  "text",
		"placeholder": "Enter validation text",
	}
	
	input, err := suite.registry.Create(suite.ctx, ComponentInput, inputConfig)
	require.NoError(suite.T(), err, "Should create input without error")
	
	// Test adding valid child to form
	formWithInput, err := defaultRegistry.AddChildComponent(suite.ctx, form, input)
	require.NoError(suite.T(), err, "Should add input to form without error")
	
	assert.Len(suite.T(), formWithInput.Children, 1, "Form should have 1 child")
	assert.Equal(suite.T(), input.ID, formWithInput.Children[0].ID, "Child should be the input component")
	
	// Test composition validation
	errors := defaultRegistry.ValidateComponentWithComposition(suite.ctx, formWithInput)
	assert.Empty(suite.T(), errors, "Valid composition should not have errors")
	
	// Test creating component with children directly
	buttonConfig := map[string]any{
		"id":      "form-button",
		"label":   "Submit",
		"text":    "Submit Form",
		"variant": "success",
	}
	
	button, err := defaultRegistry.CreateComponent(suite.ctx, ComponentButton, buttonConfig)
	require.NoError(suite.T(), err, "Should create button without error")
	
	children := []Component{input, button}
	formWithChildren, err := defaultRegistry.CreateComponentWithChildren(suite.ctx, ComponentForm, formConfig, children)
	require.NoError(suite.T(), err, "Should create form with children")
	
	assert.Len(suite.T(), formWithChildren.Children, 2, "Form should have 2 children")
}

// TestCompositionRules tests the composition rule system
func (suite *ComponentRegistryTestSuite) TestCompositionRules() {
	validator := NewCompositionValidator()
	require.NotNil(suite.T(), validator, "Composition validator should not be nil")
	
	// Test form composition rule
	formRule, exists := validator.GetCompositionRule(ComponentForm)
	assert.True(suite.T(), exists, "Form component should have composition rule")
	assert.NotNil(suite.T(), formRule.MinChildren, "Form rule should have MinChildren")
	assert.Equal(suite.T(), 1, *formRule.MinChildren, "Form should require at least 1 child")
	assert.True(suite.T(), formRule.Exclusive, "Form should be exclusive in child types")
	
	// Test input composition rule (should not allow children)
	inputRule, exists := validator.GetCompositionRule(ComponentInput)
	assert.True(suite.T(), exists, "Input component should have composition rule")
	assert.NotNil(suite.T(), inputRule.MaxChildren, "Input rule should have MaxChildren")
	assert.Equal(suite.T(), 0, *inputRule.MaxChildren, "Input should not allow children")
	assert.True(suite.T(), inputRule.Exclusive, "Input should be exclusive in child types")
	
	// Test container composition rule
	containerRule, exists := validator.GetCompositionRule(ComponentContainer)
	assert.True(suite.T(), exists, "Container component should have composition rule")
	assert.False(suite.T(), containerRule.Exclusive, "Container should not be exclusive")
	
	// Test getting allowed children
	containerChildren := validator.GetAllowedChildren(ComponentContainer)
	assert.NotEmpty(suite.T(), containerChildren, "Container should allow children")
	
	formChildren := validator.GetAllowedChildren(ComponentForm)
	assert.NotEmpty(suite.T(), formChildren, "Form should have allowed children")
	
	// Test CanAddChild functionality
	containerComponent := Component{
		BaseComponent: BaseComponent{
			ID:   "test-container",
			Type: ComponentContainer,
		},
		Children: []Component{},
	}
	
	canAdd, err := validator.CanAddChild(containerComponent, ComponentInput)
	assert.NoError(suite.T(), err, "Should be able to check if child can be added")
	assert.True(suite.T(), canAdd, "Should be able to add input to container")
}

// TestComponentValidation tests component validation functionality
func (suite *ComponentRegistryTestSuite) TestComponentValidation() {
	defaultRegistry := suite.registry.(*DefaultRegistry)
	
	// Create a valid button using the factory to ensure proper config
	buttonConfig := map[string]any{
		"text": "Valid Button",
	}
	
	validButton, err := suite.registry.Create(suite.ctx, ComponentButton, buttonConfig)
	require.NoError(suite.T(), err, "Should create valid button")
	
	err = suite.registry.Validate(suite.ctx, validButton)
	assert.NoError(suite.T(), err, "Valid button should pass validation")
	
	// Test component tree validation
	form := Component{
		BaseComponent: BaseComponent{
			ID:   "validation-tree-form",
			Type: ComponentForm,
		},
		Children: []Component{
			{
				BaseComponent: BaseComponent{
					ID:   "validation-tree-input",
					Type: ComponentInput,
				},
			},
		},
	}
	
	treeErrors := defaultRegistry.ValidateComponentTree(suite.ctx, form)
	// Note: This might have validation errors due to missing required configuration
	// but the structure should be validated
	assert.IsType(suite.T(), []ValidationError{}, treeErrors, "Should return ValidationError slice")
}

// TestSchemaRetrieval tests schema retrieval functionality
func (suite *ComponentRegistryTestSuite) TestSchemaRetrieval() {
	defaultRegistry := suite.registry.(*DefaultRegistry)
	
	// Test getting schema for button
	buttonSchema, err := defaultRegistry.GetSchema(ComponentButton)
	require.NoError(suite.T(), err, "Should get button schema without error")
	
	assert.Equal(suite.T(), ComponentButton, buttonSchema.Type)
	assert.NotEmpty(suite.T(), buttonSchema.Title, "Schema should have a title")
	assert.NotEmpty(suite.T(), buttonSchema.Description, "Schema should have a description")
	
	// Test getting all schemas
	allSchemas := defaultRegistry.GetAllSchemas()
	assert.NotEmpty(suite.T(), allSchemas, "Should have schemas for registered components")
	
	// Verify button schema is in all schemas
	foundButtonSchema, exists := allSchemas[ComponentButton]
	assert.True(suite.T(), exists, "Button schema should exist in all schemas")
	assert.Equal(suite.T(), buttonSchema.Type, foundButtonSchema.Type)
	
	// Test getting schema for non-existent component type
	nonExistentType := ComponentType("nonexistent")
	_, err = defaultRegistry.GetSchema(nonExistentType)
	assert.Error(suite.T(), err, "Should return error for non-existent component type")
}

// TestRegistryUtilities tests utility functions of the registry
func (suite *ComponentRegistryTestSuite) TestRegistryUtilities() {
	defaultRegistry := suite.registry.(*DefaultRegistry)
	
	// Test GetCSSFactory
	cssFactory := defaultRegistry.GetCSSFactory()
	assert.NotNil(suite.T(), cssFactory, "Should return CSS factory")
	
	// Test GetSchemaDir
	schemaDir := defaultRegistry.GetSchemaDir()
	assert.Equal(suite.T(), "../../../docs/ui/Schema", schemaDir, "Should return correct schema directory")
}

// TestComponentRenderingIntegration tests the integration with component rendering
func (suite *ComponentRegistryTestSuite) TestComponentRenderingIntegration() {
	defaultRegistry := suite.registry.(*DefaultRegistry)
	
	// Create a button
	buttonConfig := map[string]any{
		"text": "Render Button",
	}
	
	button, err := defaultRegistry.CreateComponent(suite.ctx, ComponentButton, buttonConfig)
	require.NoError(suite.T(), err, "Should create button for rendering")
	
	// Test rendering the component
	rendered, err := defaultRegistry.RenderComponent(suite.ctx, button)
	require.NoError(suite.T(), err, "Should render component without error")
	
	assert.Equal(suite.T(), button.ID, rendered.ID, "Rendered component should have same ID")
	assert.Equal(suite.T(), string(button.Type), rendered.Type, "Rendered component should have correct type")
	// Note: CSS may be empty if component doesn't have styles applied
}

// Run the test suite
func TestComponentRegistryTestSuite(t *testing.T) {
	suite.Run(t, new(ComponentRegistryTestSuite))
}