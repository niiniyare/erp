package ui

import (
	"context"
	"fmt"
	"sync"
)

// DefaultRegistry implements ComponentRegistry with built-in factories
type DefaultRegistry struct {
	factories map[ComponentType]ComponentFactory
	mu        sync.RWMutex
}

// NewRegistry creates a new component registry with default factories
func NewRegistry() *DefaultRegistry {
	registry := &DefaultRegistry{
		factories: make(map[ComponentType]ComponentFactory),
	}

	// Register form components
	registry.Register(ComponentInput, &InputFactory{})
	registry.Register(ComponentTextarea, &TextareaFactory{})
	registry.Register(ComponentSelect, &SelectFactory{})
	registry.Register(ComponentCheckbox, &CheckboxFactory{})
	registry.Register(ComponentRadio, &RadioFactory{})
	registry.Register(ComponentButton, &ButtonFactory{})
	registry.Register(ComponentDatePicker, &DatePickerFactory{})
	registry.Register(ComponentFileUpload, &FileUploadFactory{})
	registry.Register(ComponentForm, &FormFactory{})

	// Register layout components
	registry.Register(ComponentContainer, &ContainerFactory{})
	registry.Register(ComponentCard, &CardFactory{})
	registry.Register(ComponentPanel, &PanelFactory{})
	registry.Register(ComponentTabs, &TabsFactory{})
	registry.Register(ComponentModal, &ModalFactory{})
	registry.Register(ComponentDrawer, &DrawerFactory{})

	// Register data display components
	registry.Register(ComponentTable, &TableFactory{})
	registry.Register(ComponentList, &ListFactory{})
	registry.Register(ComponentTree, &TreeFactory{})
	registry.Register(ComponentChart, &ChartFactory{})
	registry.Register(ComponentBadge, &BadgeFactory{})
	registry.Register(ComponentTag, &TagFactory{})

	return registry
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

// ComponentBuilder helper functions
func CreateInput(id, inputType string) *ComponentBuilder {
	return NewComponent(ComponentInput, id).WithConfig(InputConfig{
		InputType: InputType(inputType),
	})
}

func CreateSelect(id string, options []Option) *ComponentBuilder {
	return NewComponent(ComponentSelect, id).WithConfig(SelectConfig{
		Options: options,
	})
}

func CreateButton(id, text string) *ComponentBuilder {
	return NewComponent(ComponentButton, id).WithConfig(ButtonConfig{
		Text: text,
	})
}

func CreateCard(id string) *ComponentBuilder {
	return NewComponent(ComponentCard, id)
}

func CreateTable(id string, columns []TableColumn) *ComponentBuilder {
	return NewComponent(ComponentTable, id).WithConfig(TableConfig{
		Columns: columns,
	})
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