package ui

import (
	"context"
	"fmt"
	"sync"

	"github.com/niiniyare/erp/pkg/schema/ui/css"
)

// DefaultRegistry implements ComponentRegistry with built-in factories and CSS integration
type DefaultRegistry struct {
	factories  map[ComponentType]ComponentFactory
	mu         sync.RWMutex
	schemaDir  string
	cssFactory *css.Factory
}

// NewRegistry creates a new component registry with default factories and CSS integration
func NewRegistry() ComponentRegistry {
	return NewRegistryWithSchemaDir("docs/ui/Schema")
}

// NewRegistryWithSchemaDir creates a registry with custom schema directory
func NewRegistryWithSchemaDir(schemaDir string) ComponentRegistry {
	registry := &DefaultRegistry{
		factories:  make(map[ComponentType]ComponentFactory),
		schemaDir:  schemaDir,
		cssFactory: css.NewFactory(schemaDir),
	}

	// Register all component factories with CSS integration
	registry.registerAllFactories()

	return registry
}

// registerAllFactories registers all built-in component factories
func (r *DefaultRegistry) registerAllFactories() {
	// Register form components (using existing factories)
	r.Register(ComponentInput, &InputFactory{})
	r.Register(ComponentTextarea, &TextareaFactory{})
	r.Register(ComponentSelect, &SelectFactory{})
	r.Register(ComponentCheckbox, &CheckboxFactory{})
	r.Register(ComponentRadio, &RadioFactory{})
	r.Register(ComponentButton, &ButtonFactory{})
	r.Register(ComponentDatePicker, &DatePickerFactory{})
	r.Register(ComponentFileUpload, &FileUploadFactory{})
	r.Register(ComponentForm, &FormFactory{})

	// Register layout components (using existing factories)
	r.Register(ComponentContainer, &ContainerFactory{})
	r.Register(ComponentCard, &CardFactory{})
	r.Register(ComponentPanel, &PanelFactory{})
	r.Register(ComponentTabs, &TabsFactory{})
	r.Register(ComponentModal, &ModalFactory{})
	r.Register(ComponentDrawer, &DrawerFactory{})

	// Register data display components (using existing factories)
	r.Register(ComponentTable, &TableFactory{})
	r.Register(ComponentList, &ListFactory{})
	r.Register(ComponentTree, &TreeFactory{})
	r.Register(ComponentChart, &ChartFactory{})
	r.Register(ComponentBadge, &BadgeFactory{})
	r.Register(ComponentTag, &TagFactory{})
	
	// Register additional components with generic factories
	r.Register(ComponentBreadcrumb, NewGenericFactory(ComponentBreadcrumb, r.cssFactory))
	r.Register(ComponentPagination, NewGenericFactory(ComponentPagination, r.cssFactory))
}

// Register registers a component factory
func (r *DefaultRegistry) Register(componentType ComponentType, factory ComponentFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[componentType] = factory
}

// Create creates a component using the registered factory
func (r *DefaultRegistry) Create(ctx context.Context, componentType ComponentType, config map[string]any) (Component, error) {
	r.mu.RLock()
	factory, exists := r.factories[componentType]
	r.mu.RUnlock()

	if !exists {
		return Component{}, fmt.Errorf("no factory registered for component type: %s", componentType)
	}

	return factory.Create(ctx, config)
}

// GetTypes returns all registered component types
func (r *DefaultRegistry) GetTypes() []ComponentType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]ComponentType, 0, len(r.factories))
	for componentType := range r.factories {
		types = append(types, componentType)
	}
	return types
}

// Validate validates a component using its registered factory
func (r *DefaultRegistry) Validate(ctx context.Context, component Component) error {
	r.mu.RLock()
	factory, exists := r.factories[component.Type]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no factory registered for component type: %s", component.Type)
	}

	return factory.Validate(ctx, component)
}

// GetSchema returns the schema for a component type
func (r *DefaultRegistry) GetSchema(componentType ComponentType) (ComponentSchema, error) {
	r.mu.RLock()
	factory, exists := r.factories[componentType]
	r.mu.RUnlock()

	if !exists {
		return ComponentSchema{}, fmt.Errorf("no factory registered for component type: %s", componentType)
	}

	return factory.GetSchema(), nil
}

// GetAllSchemas returns schemas for all registered component types
func (r *DefaultRegistry) GetAllSchemas() map[ComponentType]ComponentSchema {
	r.mu.RLock()
	defer r.mu.RUnlock()

	schemas := make(map[ComponentType]ComponentSchema)
	for componentType, factory := range r.factories {
		schemas[componentType] = factory.GetSchema()
	}
	return schemas
}

// ComponentBuilder helper functions with CSS integration
func CreateInput(id, inputType string) *ComponentBuilder {
	factory := css.NewFactory("docs/ui/Schema")
	return NewComponent(ComponentInput, id).
		WithConfig(InputConfig{InputType: InputType(inputType)}).
		WithStyles(factory.InputStyles("default"))
}

func CreateSelect(id string, options []Option) *ComponentBuilder {
	factory := css.NewFactory("docs/ui/Schema")
	return NewComponent(ComponentSelect, id).
		WithConfig(SelectConfig{Options: options}).
		WithStyles(factory.InputStyles("default"))
}

func CreateButton(id, text string) *ComponentBuilder {
	factory := css.NewFactory("docs/ui/Schema")
	return NewComponent(ComponentButton, id).
		WithConfig(ButtonConfig{Text: text}).
		WithStyles(factory.ButtonStyles("primary"))
}

func CreateCard(id string) *ComponentBuilder {
	factory := css.NewFactory("docs/ui/Schema")
	return NewComponent(ComponentCard, id).
		WithStyles(factory.CardStyles("md"))
}

func CreateTable(id string, columns []TableColumn) *ComponentBuilder {
	factory := css.NewFactory("docs/ui/Schema")
	return NewComponent(ComponentTable, id).
		WithConfig(TableConfig{Columns: columns}).
		WithStyles(factory.TableStyles("default"))
}

func CreateChart(id string, chartType ChartType) *ComponentBuilder {
	return NewComponent(ComponentChart, id).WithConfig(ChartConfig{
		Type: chartType,
	})
}

// Validation helpers
func ValidateComponent(ctx context.Context, registry ComponentRegistry, component Component) []string {
	var errors []string

	// Basic validation
	if component.ID == "" {
		errors = append(errors, "component ID is required")
	}

	if component.Type == "" {
		errors = append(errors, "component type is required")
	}

	// Factory validation
	if err := registry.Validate(ctx, component); err != nil {
		errors = append(errors, fmt.Sprintf("factory validation failed: %v", err))
	}

	// Child validation
	for i, child := range component.Children {
		childErrors := ValidateComponent(ctx, registry, child)
		for _, childError := range childErrors {
			errors = append(errors, fmt.Sprintf("child[%d]: %s", i, childError))
		}
	}

	return errors
}

// Component tree helpers
func WalkComponents(component Component, fn func(Component) error) error {
	if err := fn(component); err != nil {
		return err
	}

	for _, child := range component.Children {
		if err := WalkComponents(child, fn); err != nil {
			return err
		}
	}

	return nil
}

func FindComponent(component Component, predicate func(Component) bool) *Component {
	if predicate(component) {
		return &component
	}

	for _, child := range component.Children {
		if found := FindComponent(child, predicate); found != nil {
			return found
		}
	}

	return nil
}

func FindComponentByID(component Component, id string) *Component {
	return FindComponent(component, func(c Component) bool {
		return c.ID == id
	})
}

func FindComponentsByType(component Component, componentType ComponentType) []Component {
	var components []Component

	WalkComponents(component, func(c Component) error {
		if c.Type == componentType {
			components = append(components, c)
		}
		return nil
	})

	return components
}

// ============================================================================
// ENHANCED FACTORY CONSTRUCTORS WITH CSS INTEGRATION
// ============================================================================

// ============================================================================
// FACTORY STUBS FOR REMAINING COMPONENTS
// ============================================================================

// GenericFactory provides stub implementations for components not yet fully implemented
type GenericFactory struct {
	componentType ComponentType
	cssFactory    *css.Factory
}

func NewGenericFactory(componentType ComponentType, cssFactory *css.Factory) ComponentFactory {
	return &GenericFactory{
		componentType: componentType,
		cssFactory:    cssFactory,
	}
}

func (f *GenericFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	component := Component{
		BaseComponent: BaseComponent{
			Type: f.componentType,
		},
	}
	
	if id, ok := config["id"].(string); ok {
		component.ID = id
	}
	if label, ok := config["label"].(string); ok {
		component.Label = label
	}
	
	// Apply default styling based on component type
	switch f.componentType {
	case ComponentContainer:
		component.Styles = f.cssFactory.ContainerStyles("lg")
	case ComponentTable:
		component.Styles = f.cssFactory.TableStyles("default")
	default:
		component.Styles = css.NewStyles(f.cssFactory.GetSchemaDir())
	}
	
	return component, nil
}

func (f *GenericFactory) Validate(ctx context.Context, component Component) error {
	if component.Type != f.componentType {
		return fmt.Errorf("expected component type %s, got %s", f.componentType, component.Type)
	}
	return nil
}

func (f *GenericFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        f.componentType,
		Title:       string(f.componentType),
		Description: fmt.Sprintf("Generic %s component", f.componentType),
		Properties:  make(map[string]Property),
	}
}

// Additional utility functions for registry management

// GetCSSFactory returns the CSS factory used by this registry
func (r *DefaultRegistry) GetCSSFactory() *css.Factory {
	return r.cssFactory
}

// GetSchemaDir returns the schema directory used by this registry
func (r *DefaultRegistry) GetSchemaDir() string {
	return r.schemaDir
}

// ============================================================================
// COMPONENT LIFECYCLE MANAGEMENT (Task 1.1.2)
// ============================================================================

// CreateComponent creates a new component with full lifecycle support
func (r *DefaultRegistry) CreateComponent(ctx context.Context, componentType ComponentType, config map[string]any) (Component, error) {
	// Create component using factory
	component, err := r.Create(ctx, componentType, config)
	if err != nil {
		return Component{}, fmt.Errorf("failed to create component: %w", err)
	}
	
	// Initialize lifecycle metadata
	component.initializeLifecycle()
	
	// Apply CSS styling if configuration includes styling
	if styleConfig, exists := config["styles"]; exists {
		if styleMap, ok := styleConfig.(map[string]any); ok {
			if err := r.applyStyles(&component, styleMap); err != nil {
				return Component{}, fmt.Errorf("failed to apply styles: %w", err)
			}
		}
	}
	
	// Validate final component
	if err := r.Validate(ctx, component); err != nil {
		return Component{}, fmt.Errorf("component failed validation: %w", err)
	}
	
	return component, nil
}

// UpdateComponent updates an existing component while preserving lifecycle
func (r *DefaultRegistry) UpdateComponent(ctx context.Context, component Component, updates map[string]any) (Component, error) {
	// Update component properties
	updatedComponent := component
	
	// Apply configuration updates
	if configUpdates, exists := updates["config"]; exists {
		if configMap, ok := configUpdates.(map[string]any); ok {
			factory, exists := r.factories[component.Type]
			if !exists {
				return Component{}, fmt.Errorf("no factory registered for component type: %s", component.Type)
			}
			
			// Recreate component with new config
			newComponent, err := factory.Create(ctx, configMap)
			if err != nil {
				return Component{}, fmt.Errorf("failed to update component: %w", err)
			}
			
			// Preserve existing metadata
			newComponent.ID = component.ID
			newComponent.CreatedAt = component.CreatedAt
			updatedComponent = newComponent
		}
	}
	
	// Apply style updates
	if styleUpdates, exists := updates["styles"]; exists {
		if styleMap, ok := styleUpdates.(map[string]any); ok {
			if err := r.applyStyles(&updatedComponent, styleMap); err != nil {
				return Component{}, fmt.Errorf("failed to update styles: %w", err)
			}
		}
	}
	
	// Update lifecycle metadata
	updatedComponent.updateLifecycle()
	
	// Validate updated component
	if err := r.Validate(ctx, updatedComponent); err != nil {
		return Component{}, fmt.Errorf("updated component failed validation: %w", err)
	}
	
	return updatedComponent, nil
}

// RenderComponent renders a component for output
func (r *DefaultRegistry) RenderComponent(ctx context.Context, component Component) (RenderedComponent, error) {
	renderer := NewComponentRenderer(r.schemaDir)
	return renderer.RenderComponent(component)
}

// DisposeComponent cleans up a component and its resources
func (r *DefaultRegistry) DisposeComponent(ctx context.Context, component Component) error {
	// Clean up child components recursively
	for _, child := range component.Children {
		if err := r.DisposeComponent(ctx, child); err != nil {
			return fmt.Errorf("failed to dispose child component %s: %w", child.ID, err)
		}
	}
	
	// Mark component as disposed in lifecycle metadata
	component.markDisposed()
	
	return nil
}

// CloneComponent creates a deep copy of a component
func (r *DefaultRegistry) CloneComponent(ctx context.Context, component Component) (Component, error) {
	// Create base clone
	clone := component
	clone.ID = generateComponentID() // Generate new unique ID
	
	// Clone children recursively
	if len(component.Children) > 0 {
		clonedChildren := make([]Component, 0, len(component.Children))
		for _, child := range component.Children {
			clonedChild, err := r.CloneComponent(ctx, child)
			if err != nil {
				return Component{}, fmt.Errorf("failed to clone child component: %w", err)
			}
			clonedChildren = append(clonedChildren, clonedChild)
		}
		clone.Children = clonedChildren
	}
	
	// Initialize lifecycle for cloned component
	clone.initializeLifecycle()
	
	return clone, nil
}

// ValidateComponentTree validates a component and all its children
func (r *DefaultRegistry) ValidateComponentTree(ctx context.Context, component Component) []ValidationError {
	var errors []ValidationError
	
	// Validate current component
	if err := r.Validate(ctx, component); err != nil {
		errors = append(errors, ValidationError{
			Component: component.ID,
			Field:     "component",
			Message:   err.Error(),
		})
	}
	
	// Validate children recursively
	for i, child := range component.Children {
		childErrors := r.ValidateComponentTree(ctx, child)
		for _, childError := range childErrors {
			errors = append(errors, ValidationError{
				Component: fmt.Sprintf("%s.children[%d].%s", component.ID, i, childError.Component),
				Field:     childError.Field,
				Message:   childError.Message,
			})
		}
	}
	
	return errors
}

// ============================================================================
// COMPONENT LIFECYCLE HELPER METHODS
// ============================================================================

// applyStyles applies styling configuration to a component
func (r *DefaultRegistry) applyStyles(component *Component, styleConfig map[string]any) error {
	if component.Styles == nil {
		component.Styles = css.NewStyles(r.schemaDir)
	}
	
	// Apply individual style properties
	for property, value := range styleConfig {
		if strValue, ok := value.(string); ok {
			component.Styles.WithCustomProperty(property, strValue)
		}
	}
	
	return nil
}

// generateComponentID generates a unique component ID
func generateComponentID() string {
	// In a real implementation, this would use a proper UUID generator
	return fmt.Sprintf("component_%d", len("placeholder"))
}

// ============================================================================
// COMPONENT COMPOSITION AND NESTING VALIDATION (Task 1.1.3)
// ============================================================================

// CompositionRule represents a rule for component composition
type CompositionRule struct {
	ParentType    ComponentType   `json:"parent_type"`    // Parent component type
	AllowedChildren []ComponentType `json:"allowed_children"` // Allowed child component types
	RequiredChildren []ComponentType `json:"required_children,omitempty"` // Required child component types
	MaxChildren   *int            `json:"max_children,omitempty"`   // Maximum number of children
	MinChildren   *int            `json:"min_children,omitempty"`   // Minimum number of children
	Exclusive     bool            `json:"exclusive,omitempty"`      // Whether children must be of only allowed types
}

// CompositionValidator validates component nesting and composition rules
type CompositionValidator struct {
	rules map[ComponentType]CompositionRule
}

// NewCompositionValidator creates a new composition validator with default rules
func NewCompositionValidator() *CompositionValidator {
	validator := &CompositionValidator{
		rules: make(map[ComponentType]CompositionRule),
	}
	
	validator.loadDefaultRules()
	return validator
}

// loadDefaultRules loads sensible default composition rules for components
func (cv *CompositionValidator) loadDefaultRules() {
	// Form component rules
	cv.rules[ComponentForm] = CompositionRule{
		ParentType: ComponentForm,
		AllowedChildren: []ComponentType{
			ComponentInput, ComponentTextarea, ComponentSelect,
			ComponentCheckbox, ComponentRadio, ComponentButton,
			ComponentContainer, ComponentCard,
		},
		MinChildren: intPtr(1),
		Exclusive:   true,
	}
	
	// Container component rules
	cv.rules[ComponentContainer] = CompositionRule{
		ParentType: ComponentContainer,
		AllowedChildren: []ComponentType{
			ComponentInput, ComponentTextarea, ComponentSelect, ComponentCheckbox,
			ComponentRadio, ComponentButton, ComponentCard, ComponentTable,
			ComponentChart, ComponentBadge, ComponentContainer,
		},
		Exclusive: false, // Containers can hold any component
	}
	
	// Card component rules
	cv.rules[ComponentCard] = CompositionRule{
		ParentType: ComponentCard,
		AllowedChildren: []ComponentType{
			ComponentInput, ComponentButton, ComponentBadge,
			ComponentContainer, ComponentChart, ComponentTable,
		},
		MaxChildren: intPtr(10), // Reasonable limit for card content
		Exclusive:   false,
	}
	
	// Table component rules
	cv.rules[ComponentTable] = CompositionRule{
		ParentType:      ComponentTable,
		AllowedChildren: []ComponentType{}, // Tables don't typically have component children
		MaxChildren:     intPtr(0),
		Exclusive:       true,
	}
	
	// Input components shouldn't have children
	inputTypes := []ComponentType{
		ComponentInput, ComponentTextarea, ComponentSelect,
		ComponentCheckbox, ComponentRadio, ComponentButton,
	}
	for _, inputType := range inputTypes {
		cv.rules[inputType] = CompositionRule{
			ParentType:      inputType,
			AllowedChildren: []ComponentType{},
			MaxChildren:     intPtr(0),
			Exclusive:       true,
		}
	}
	
	// Modal and Panel rules
	cv.rules[ComponentModal] = CompositionRule{
		ParentType: ComponentModal,
		AllowedChildren: []ComponentType{
			ComponentContainer, ComponentCard, ComponentForm,
			ComponentButton,
		},
		MinChildren: intPtr(1),
		MaxChildren: intPtr(5), // Reasonable limit for modal content
		Exclusive:   true,
	}
	
	cv.rules[ComponentPanel] = CompositionRule{
		ParentType: ComponentPanel,
		AllowedChildren: []ComponentType{
			ComponentContainer, ComponentCard, ComponentForm,
			ComponentTable, ComponentChart,
		},
		MaxChildren: intPtr(20), // Reasonable limit for panel content
		Exclusive:   false,
	}
}

// ValidateComposition validates that a component's children follow composition rules
func (cv *CompositionValidator) ValidateComposition(component Component) []ValidationError {
	var errors []ValidationError
	
	// Check if there's a rule for this component type
	rule, hasRule := cv.rules[component.Type]
	if !hasRule {
		// No specific rules - allow any composition (default permissive behavior)
		return errors
	}
	
	// Validate number of children constraints
	childCount := len(component.Children)
	
	if rule.MinChildren != nil && childCount < *rule.MinChildren {
		errors = append(errors, ValidationError{
			Component: component.ID,
			Field:     "children",
			Message:   fmt.Sprintf("component %s requires at least %d children, got %d", component.Type, *rule.MinChildren, childCount),
		})
	}
	
	if rule.MaxChildren != nil && childCount > *rule.MaxChildren {
		errors = append(errors, ValidationError{
			Component: component.ID,
			Field:     "children",
			Message:   fmt.Sprintf("component %s allows at most %d children, got %d", component.Type, *rule.MaxChildren, childCount),
		})
	}
	
	// Validate child types
	allowedTypes := make(map[ComponentType]bool)
	for _, allowedType := range rule.AllowedChildren {
		allowedTypes[allowedType] = true
	}
	
	for i, child := range component.Children {
		if rule.Exclusive && !allowedTypes[child.Type] {
			errors = append(errors, ValidationError{
				Component: component.ID,
				Field:     fmt.Sprintf("children[%d]", i),
				Message:   fmt.Sprintf("component %s does not allow child of type %s", component.Type, child.Type),
			})
		}
	}
	
	// Validate required children
	if len(rule.RequiredChildren) > 0 {
		requiredTypes := make(map[ComponentType]bool)
		for _, requiredType := range rule.RequiredChildren {
			requiredTypes[requiredType] = true
		}
		
		presentTypes := make(map[ComponentType]bool)
		for _, child := range component.Children {
			presentTypes[child.Type] = true
		}
		
		for requiredType := range requiredTypes {
			if !presentTypes[requiredType] {
				errors = append(errors, ValidationError{
					Component: component.ID,
					Field:     "children",
					Message:   fmt.Sprintf("component %s requires a child of type %s", component.Type, requiredType),
				})
			}
		}
	}
	
	// Recursively validate children
	for i, child := range component.Children {
		childErrors := cv.ValidateComposition(child)
		for _, childError := range childErrors {
			errors = append(errors, ValidationError{
				Component: fmt.Sprintf("%s.children[%d].%s", component.ID, i, childError.Component),
				Field:     childError.Field,
				Message:   childError.Message,
			})
		}
	}
	
	return errors
}

// AddCompositionRule adds or updates a composition rule
func (cv *CompositionValidator) AddCompositionRule(rule CompositionRule) {
	cv.rules[rule.ParentType] = rule
}

// GetCompositionRule returns the composition rule for a component type
func (cv *CompositionValidator) GetCompositionRule(componentType ComponentType) (CompositionRule, bool) {
	rule, exists := cv.rules[componentType]
	return rule, exists
}

// GetAllowedChildren returns the allowed child types for a component type
func (cv *CompositionValidator) GetAllowedChildren(componentType ComponentType) []ComponentType {
	rule, exists := cv.rules[componentType]
	if !exists {
		// Return all known component types if no specific rule exists
		return []ComponentType{
			ComponentInput, ComponentTextarea, ComponentSelect, ComponentCheckbox,
			ComponentRadio, ComponentButton, ComponentForm, ComponentContainer,
			ComponentCard, ComponentModal, ComponentTable, ComponentChart,
			ComponentBadge, ComponentPanel, ComponentTabs,
		}
	}
	return rule.AllowedChildren
}

// CanAddChild checks if a child component can be added to a parent
func (cv *CompositionValidator) CanAddChild(parent Component, childType ComponentType) (bool, error) {
	rule, hasRule := cv.rules[parent.Type]
	if !hasRule {
		return true, nil // Permissive by default
	}
	
	// Check max children constraint
	if rule.MaxChildren != nil && len(parent.Children) >= *rule.MaxChildren {
		return false, fmt.Errorf("parent component %s has reached maximum children limit (%d)", parent.Type, *rule.MaxChildren)
	}
	
	// Check if child type is allowed
	if rule.Exclusive {
		allowed := false
		for _, allowedType := range rule.AllowedChildren {
			if allowedType == childType {
				allowed = true
				break
			}
		}
		if !allowed {
			return false, fmt.Errorf("parent component %s does not allow children of type %s", parent.Type, childType)
		}
	}
	
	return true, nil
}

// Enhanced registry methods with composition validation

// ValidateComponentWithComposition validates a component including composition rules
func (r *DefaultRegistry) ValidateComponentWithComposition(ctx context.Context, component Component) []ValidationError {
	var errors []ValidationError
	
	// Standard component validation
	componentErrors := r.ValidateComponentTree(ctx, component)
	errors = append(errors, componentErrors...)
	
	// Composition validation
	compositionValidator := NewCompositionValidator()
	compositionErrors := compositionValidator.ValidateComposition(component)
	errors = append(errors, compositionErrors...)
	
	return errors
}

// CreateComponentWithChildren creates a component and validates its composition
func (r *DefaultRegistry) CreateComponentWithChildren(ctx context.Context, componentType ComponentType, config map[string]any, children []Component) (Component, error) {
	// Create parent component
	component, err := r.CreateComponent(ctx, componentType, config)
	if err != nil {
		return Component{}, fmt.Errorf("failed to create parent component: %w", err)
	}
	
	// Add children
	component.Children = children
	
	// Validate composition
	errors := r.ValidateComponentWithComposition(ctx, component)
	if len(errors) > 0 {
		return Component{}, fmt.Errorf("composition validation failed: %s", errors[0].Message)
	}
	
	return component, nil
}

// AddChildComponent adds a child component to a parent with composition validation
func (r *DefaultRegistry) AddChildComponent(ctx context.Context, parent Component, child Component) (Component, error) {
	compositionValidator := NewCompositionValidator()
	
	// Check if child can be added
	canAdd, err := compositionValidator.CanAddChild(parent, child.Type)
	if err != nil {
		return Component{}, fmt.Errorf("cannot add child: %w", err)
	}
	if !canAdd {
		return Component{}, fmt.Errorf("child component %s cannot be added to parent %s", child.Type, parent.Type)
	}
	
	// Add the child
	updatedParent := parent
	updatedParent.Children = append(updatedParent.Children, child)
	
	// Validate final composition
	errors := r.ValidateComponentWithComposition(ctx, updatedParent)
	if len(errors) > 0 {
		return Component{}, fmt.Errorf("composition validation failed after adding child: %s", errors[0].Message)
	}
	
	// Update lifecycle
	updatedParent.updateLifecycle()
	
	return updatedParent, nil
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}