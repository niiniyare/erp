package engine

import (
	"context"
	"fmt"
	"sync"

	"github.com/a-h/templ"
	"github.com/niiniyare/erp/web/components/atoms"
	"github.com/niiniyare/erp/web/components/molecules"
	"github.com/niiniyare/erp/web/components/organisms/sidebar"
	"github.com/niiniyare/erp/web/components/organisms/table"
	"github.com/niiniyare/erp/web/layouts"
	"github.com/niiniyare/erp/web/schemas"
)

// ComponentRegistry manages component resolution and registration
type ComponentRegistry struct {
	mu        sync.RWMutex
	factories map[string]schemas.ComponentFactory
	fallbacks map[string]string   // componentType -> fallback componentType
	versions  map[string][]string // componentType -> available versions
	aliases   map[string]string   // alias -> componentType
}

// NewComponentRegistry creates a new component registry with default components
func NewComponentRegistry() *ComponentRegistry {
	registry := &ComponentRegistry{
		factories: make(map[string]schemas.ComponentFactory),
		fallbacks: make(map[string]string),
		versions:  make(map[string][]string),
		aliases:   make(map[string]string),
	}

	// Register all available components
	registry.registerDefaultComponents()

	return registry
}

// Register adds a component factory to the registry
func (r *ComponentRegistry) Register(componentType string, factory schemas.ComponentFactory) error {
	if componentType == "" {
		return fmt.Errorf("component type cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("component factory cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.factories[componentType] = factory
	return nil
}

// RegisterWithVersion adds a versioned component factory
func (r *ComponentRegistry) RegisterWithVersion(componentType, version string, factory schemas.ComponentFactory) error {
	versionedType := componentType + "@" + version
	err := r.Register(versionedType, factory)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.versions[componentType]; !exists {
		r.versions[componentType] = []string{}
	}
	r.versions[componentType] = append(r.versions[componentType], version)

	return nil
}

// RegisterAlias creates an alias for a component type
func (r *ComponentRegistry) RegisterAlias(alias, componentType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.aliases[alias] = componentType
	return nil
}

// RegisterFallback sets a fallback component for when the primary component fails
func (r *ComponentRegistry) RegisterFallback(componentType, fallbackType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.fallbacks[componentType] = fallbackType
	return nil
}

// Resolve creates a component instance from a schema definition
func (r *ComponentRegistry) Resolve(componentType string) (templ.Component, error) {
	// Resolve aliases first
	actualType := r.resolveAlias(componentType)

	// Try to get the factory
	factory, exists := r.getFactory(actualType)
	if !exists {
		// Try fallback
		if fallback, hasFallback := r.fallbacks[actualType]; hasFallback {
			factory, exists = r.getFactory(fallback)
			if !exists {
				return nil, schemas.NewSchemaError(schemas.ErrCodeComponentNotFound,
					fmt.Sprintf("component '%s' and its fallback '%s' not found", actualType, fallback))
			}
		} else {
			return nil, schemas.NewSchemaError(schemas.ErrCodeComponentNotFound,
				fmt.Sprintf("component '%s' not found in registry", actualType))
		}
	}

	// Create a default component definition for basic resolution
	def := &schemas.ComponentDefinition{
		Type:  actualType,
		Props: make(map[string]interface{}),
	}

	return factory(def, context.Background())
}

// ResolveWithDefinition creates a component instance from a complete schema definition
func (r *ComponentRegistry) ResolveWithDefinition(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
	if def == nil {
		return nil, schemas.NewSchemaError(schemas.ErrCodeInvalidProps, "component definition cannot be nil")
	}

	actualType := r.resolveAlias(def.Type)

	factory, exists := r.getFactory(actualType)
	if !exists {
		// Try fallback
		if fallback, hasFallback := r.fallbacks[actualType]; hasFallback {
			factory, exists = r.getFactory(fallback)
			if !exists {
				return nil, schemas.NewSchemaError(schemas.ErrCodeComponentNotFound,
					fmt.Sprintf("component '%s' and its fallback '%s' not found", actualType, fallback)).
					WithComponent(def.Type)
			}
		} else {
			return nil, schemas.NewSchemaError(schemas.ErrCodeComponentNotFound,
				fmt.Sprintf("component '%s' not found in registry", actualType)).
				WithComponent(def.Type)
		}
	}

	return factory(def, ctx)
}

// List returns all registered component types
func (r *ComponentRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.factories))
	for componentType := range r.factories {
		types = append(types, componentType)
	}
	return types
}

// Exists checks if a component type is registered
func (r *ComponentRegistry) Exists(componentType string) bool {
	actualType := r.resolveAlias(componentType)
	_, exists := r.getFactory(actualType)
	return exists
}

// GetVersions returns available versions for a component type
func (r *ComponentRegistry) GetVersions(componentType string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.versions[componentType]
}

// Helper methods

func (r *ComponentRegistry) resolveAlias(componentType string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if alias, exists := r.aliases[componentType]; exists {
		return alias
	}
	return componentType
}

func (r *ComponentRegistry) getFactory(componentType string) (schemas.ComponentFactory, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, exists := r.factories[componentType]
	return factory, exists
}

// Component factory functions for our existing components

func buttonFactory(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
	props := atoms.ButtonProps{
		Text:     getStringProp(def.Props, "text", "Button"),
		Type:     getStringProp(def.Props, "type", "button"),
		Variant:  atoms.ButtonVariant(getStringProp(def.Props, "variant", "primary")),
		Size:     atoms.ButtonSize(getStringProp(def.Props, "size", "md")),
		Class:    getStringProp(def.Props, "class", ""),
		Loading:  getBoolProp(def.Props, "loading", false),
		Disabled: getBoolProp(def.Props, "disabled", false),
		HxGet:    getStringProp(def.Props, "hxGet", ""),
		HxPost:   getStringProp(def.Props, "hxPost", ""),
		HxTarget: getStringProp(def.Props, "hxTarget", ""),
		HxSwap:   getStringProp(def.Props, "hxSwap", ""),
		OnClick:  getStringProp(def.Props, "onClick", ""),
	}
	return atoms.Button(props), nil
}

func inputFactory(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
	props := atoms.InputProps{
		Type:        atoms.InputType(getStringProp(def.Props, "type", "text")),
		Name:        getStringProp(def.Props, "name", ""),
		ID:          getStringProp(def.Props, "id", ""),
		Value:       getStringProp(def.Props, "value", ""),
		Placeholder: getStringProp(def.Props, "placeholder", ""),
		Required:    getBoolProp(def.Props, "required", false),
		Disabled:    getBoolProp(def.Props, "disabled", false),
		ReadOnly:    getBoolProp(def.Props, "readonly", false),
		Class:       getStringProp(def.Props, "class", ""),
		Size:        atoms.InputSize(getStringProp(def.Props, "size", "md")),
		State:       atoms.InputState(getStringProp(def.Props, "state", "default")),
	}
	return atoms.Input(props), nil
}

func fieldFactory(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
	props := molecules.FieldProps{
		Type:        molecules.FieldType(getStringProp(def.Props, "type", "input")),
		Label:       getStringProp(def.Props, "label", ""),
		Name:        getStringProp(def.Props, "name", ""),
		ID:          getStringProp(def.Props, "id", ""),
		Value:       getStringProp(def.Props, "value", ""),
		Placeholder: getStringProp(def.Props, "placeholder", ""),
		Required:    getBoolProp(def.Props, "required", false),
		Disabled:    getBoolProp(def.Props, "disabled", false),
		ErrorText:   getStringProp(def.Props, "errorText", ""),
		HelpText:    getStringProp(def.Props, "helpText", ""),
		Class:       getStringProp(def.Props, "class", ""),
		InputType:   atoms.InputType(getStringProp(def.Props, "inputType", "text")),
		InputSize:   atoms.InputSize(getStringProp(def.Props, "inputSize", "md")),
		InputState:  atoms.InputState(getStringProp(def.Props, "inputState", "default")),
	}
	return molecules.Field(props), nil
}

func dataTableFactory(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
	props := table.DataTableProps{
		Data:        getInterfaceProp(def.Props, "data", []map[string]interface{}{}).([]map[string]interface{}),
		Columns:     getInterfaceProp(def.Props, "columns", []table.DataTableColumn{}).([]table.DataTableColumn),
		Config:      getInterfaceProp(def.Props, "config", table.DataTableConfig{}).(table.DataTableConfig),
		Actions:     getInterfaceProp(def.Props, "actions", []table.DataTableAction{}).([]table.DataTableAction),
		BulkActions: getInterfaceProp(def.Props, "bulkActions", []table.DataTableAction{}).([]table.DataTableAction),
		ID:          getStringProp(def.Props, "id", ""),
		Class:       getStringProp(def.Props, "class", ""),
		Striped:     getBoolProp(def.Props, "striped", true),
		Bordered:    getBoolProp(def.Props, "bordered", false),
		Hover:       getBoolProp(def.Props, "hover", true),
		Compact:     getBoolProp(def.Props, "compact", false),
	}
	return table.DataTable(props), nil
}

func appLayoutFactory(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
	// For layouts, we need to handle content differently
	// This is a placeholder - actual layout rendering will be handled by the renderer
	props := layouts.AppLayoutProps{
		Base: layouts.BaseLayoutProps{
			Title:       getStringProp(def.Props, "title", "ERP System"),
			Description: getStringProp(def.Props, "description", ""),
			ThemeMode:   getStringProp(def.Props, "themeMode", "system"),
		},
		SidebarOpen:      getBoolProp(def.Props, "sidebarOpen", true),
		SidebarCollapsed: getBoolProp(def.Props, "sidebarCollapsed", false),
		ShowBreadcrumbs:  getBoolProp(def.Props, "showBreadcrumbs", true),
	}

	// Handle optional user field safely
	if userVal := getInterfaceProp(def.Props, "user", nil); userVal != nil {
		if user, ok := userVal.(*layouts.AppUser); ok {
			props.User = user
		}
	}

	// Handle optional sidebar props safely
	if sidebarVal := getInterfaceProp(def.Props, "sidebarProps", nil); sidebarVal != nil {
		if sidebarProps, ok := sidebarVal.(sidebar.SidebarProps); ok {
			props.SidebarProps = sidebarProps
		}
	}

	// Return a placeholder component - actual content injection handled by renderer
	return layouts.AppLayout(props, templ.NopComponent), nil
}

// Register all default components
func (r *ComponentRegistry) registerDefaultComponents() {
	// Atoms
	r.Register("atoms.button", buttonFactory)
	r.Register("atoms.input", inputFactory)
	r.Register("atoms.textarea", func(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
		props := atoms.TextareaProps{
			Name:        getStringProp(def.Props, "name", ""),
			ID:          getStringProp(def.Props, "id", ""),
			Value:       getStringProp(def.Props, "value", ""),
			Placeholder: getStringProp(def.Props, "placeholder", ""),
			Required:    getBoolProp(def.Props, "required", false),
			Disabled:    getBoolProp(def.Props, "disabled", false),
			ReadOnly:    getBoolProp(def.Props, "readonly", false),
			Rows:        getIntProp(def.Props, "rows", 3),
			Class:       getStringProp(def.Props, "class", ""),
		}
		return atoms.Textarea(props), nil
	})

	r.Register("atoms.select", func(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
		props := atoms.SelectProps{
			Name:     getStringProp(def.Props, "name", ""),
			ID:       getStringProp(def.Props, "id", ""),
			Value:    getStringProp(def.Props, "value", ""),
			Required: getBoolProp(def.Props, "required", false),
			Disabled: getBoolProp(def.Props, "disabled", false),
			Multiple: getBoolProp(def.Props, "multiple", false),
			Class:    getStringProp(def.Props, "class", ""),
			Options:  getInterfaceProp(def.Props, "options", []atoms.SelectOption{}).([]atoms.SelectOption),
		}
		return atoms.Select(props), nil
	})

	r.Register("atoms.checkbox", func(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
		props := atoms.CheckboxProps{
			Name:     getStringProp(def.Props, "name", ""),
			ID:       getStringProp(def.Props, "id", ""),
			Value:    getStringProp(def.Props, "value", ""),
			Label:    getStringProp(def.Props, "label", ""),
			Checked:  getBoolProp(def.Props, "checked", false),
			Required: getBoolProp(def.Props, "required", false),
			Disabled: getBoolProp(def.Props, "disabled", false),
			Class:    getStringProp(def.Props, "class", ""),
		}
		return atoms.Checkbox(props), nil
	})

	r.Register("atoms.radio", func(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
		props := atoms.RadioProps{
			Name:     getStringProp(def.Props, "name", ""),
			ID:       getStringProp(def.Props, "id", ""),
			Value:    getStringProp(def.Props, "value", ""),
			Label:    getStringProp(def.Props, "label", ""),
			Checked:  getBoolProp(def.Props, "checked", false),
			Required: getBoolProp(def.Props, "required", false),
			Disabled: getBoolProp(def.Props, "disabled", false),
			Class:    getStringProp(def.Props, "class", ""),
		}
		return atoms.Radio(props), nil
	})

	r.Register("atoms.toggle", func(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
		props := atoms.ToggleProps{
			Name:     getStringProp(def.Props, "name", ""),
			ID:       getStringProp(def.Props, "id", ""),
			Label:    getStringProp(def.Props, "label", ""),
			Checked:  getBoolProp(def.Props, "checked", false),
			Disabled: getBoolProp(def.Props, "disabled", false),
			Class:    getStringProp(def.Props, "class", ""),
		}
		return atoms.Toggle(props), nil
	})

	r.Register("atoms.icon", func(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
		props := atoms.IconProps{
			Name:  getStringProp(def.Props, "name", ""),
			Size:  atoms.IconSize(getStringProp(def.Props, "size", "md")),
			Class: getStringProp(def.Props, "class", ""),
		}
		return atoms.Icon(props), nil
	})

	r.Register("atoms.spinner", func(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
		props := atoms.SpinnerProps{
			Size:  atoms.SpinnerSize(getStringProp(def.Props, "size", "md")),
			Color: getStringProp(def.Props, "color", "blue"),
			Class: getStringProp(def.Props, "class", ""),
		}
		return atoms.Spinner(props), nil
	})

	// Molecules
	r.Register("molecules.field", fieldFactory)
	r.Register("molecules.field-group", func(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
		props := molecules.FieldGroupProps{
			Title:       getStringProp(def.Props, "title", ""),
			Description: getStringProp(def.Props, "description", ""),
			Fields:      getInterfaceProp(def.Props, "fields", []molecules.FieldProps{}).([]molecules.FieldProps),
			Layout:      molecules.FieldGroupLayout(getStringProp(def.Props, "layout", "vertical")),
			Collapsible: getBoolProp(def.Props, "collapsible", false),
			Collapsed:   getBoolProp(def.Props, "collapsed", false),
			Class:       getStringProp(def.Props, "class", ""),
		}
		return molecules.FieldGroup(props), nil
	})

	r.Register("molecules.search", func(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
		props := molecules.SearchProps{
			Name:            getStringProp(def.Props, "name", "search"),
			ID:              getStringProp(def.Props, "id", "search"),
			Placeholder:     getStringProp(def.Props, "placeholder", "Search..."),
			Value:           getStringProp(def.Props, "value", ""),
			ShowClearButton: getBoolProp(def.Props, "showClearButton", true),
			Class:           getStringProp(def.Props, "class", ""),
		}
		return molecules.Search(props), nil
	})

	r.Register("molecules.base-card", func(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
		props := molecules.BaseCardProps{
			Variant: molecules.CardVariant(getStringProp(def.Props, "variant", "default")),
			Size:    molecules.CardSize(getStringProp(def.Props, "size", "md")),
			Class:   getStringProp(def.Props, "class", ""),
		}

		// Handle optional header
		if title := getStringProp(def.Props, "title", ""); title != "" {
			props.Header = &molecules.CardHeaderProps{
				Title:    title,
				Subtitle: getStringProp(def.Props, "subtitle", ""),
				Actions:  getInterfaceProp(def.Props, "headerActions", []molecules.CardAction{}).([]molecules.CardAction),
			}
		}

		// Handle optional footer
		if footerActions := getInterfaceProp(def.Props, "footerActions", nil); footerActions != nil {
			if actions, ok := footerActions.([]molecules.CardAction); ok && len(actions) > 0 {
				props.Footer = &molecules.CardFooterProps{
					Actions: actions,
				}
			}
		}

		return molecules.BaseCard(props), nil
	})

	r.Register("molecules.alert", func(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
		props := molecules.AlertProps{
			Type:        molecules.AlertType(getStringProp(def.Props, "type", "info")),
			Size:        molecules.AlertSize(getStringProp(def.Props, "size", "md")),
			Title:       getStringProp(def.Props, "title", ""),
			Message:     getStringProp(def.Props, "message", ""),
			Dismissible: getBoolProp(def.Props, "dismissible", false),
			AutoDismiss: getBoolProp(def.Props, "autoDismiss", false),
			Duration:    getIntProp(def.Props, "duration", 0),
			Actions:     getInterfaceProp(def.Props, "actions", []molecules.AlertAction{}).([]molecules.AlertAction),
		}
		return molecules.Alert(props), nil
	})

	r.Register("molecules.dropdown", func(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
		props := molecules.DropdownProps{
			ID:                getStringProp(def.Props, "id", "dropdown"),
			TriggerText:       getStringProp(def.Props, "triggerText", "Click me"),
			TriggerIcon:       getStringProp(def.Props, "triggerIcon", ""),
			TriggerVariant:    atoms.ButtonVariant(getStringProp(def.Props, "triggerVariant", "primary")),
			TriggerSize:       atoms.ButtonSize(getStringProp(def.Props, "triggerSize", "md")),
			Items:             getInterfaceProp(def.Props, "items", []molecules.DropdownItem{}).([]molecules.DropdownItem),
			Placement:         molecules.DropdownPlacement(getStringProp(def.Props, "placement", "bottom-left")),
			SearchEnabled:     getBoolProp(def.Props, "searchEnabled", false),
			SearchPlaceholder: getStringProp(def.Props, "searchPlaceholder", "Search..."),
			MaxHeight:         getStringProp(def.Props, "maxHeight", ""),
			Class:             getStringProp(def.Props, "class", ""),
		}
		return molecules.Dropdown(props), nil
	})

	// Organisms
	r.Register("organisms.table", dataTableFactory)

	// Layouts
	r.Register("layouts.app", appLayoutFactory)
	r.Register("layouts.auth", func(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
		props := layouts.AuthLayoutProps{
			Base: layouts.BaseLayoutProps{
				Title:       getStringProp(def.Props, "title", "Authentication"),
				Description: getStringProp(def.Props, "description", ""),
				ThemeMode:   getStringProp(def.Props, "themeMode", "system"),
			},
			FormTitle: getStringProp(def.Props, "formTitle", ""),
			AppName:   getStringProp(def.Props, "appName", "ERP System"),
			AppLogo:   getStringProp(def.Props, "appLogo", ""),
		}
		return layouts.AuthLayout(props, templ.NopComponent), nil
	})

	r.Register("layouts.minimal", func(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
		props := layouts.MinimalLayoutProps{
			Base: layouts.BaseLayoutProps{
				Title:       getStringProp(def.Props, "title", "ERP System"),
				Description: getStringProp(def.Props, "description", ""),
				ThemeMode:   getStringProp(def.Props, "themeMode", "system"),
			},
		}
		return layouts.MinimalLayout(props, templ.NopComponent), nil
	})

	// Register common aliases
	r.RegisterAlias("button", "atoms.button")
	r.RegisterAlias("input", "atoms.input")
	r.RegisterAlias("field", "molecules.field")
	r.RegisterAlias("table", "organisms.table")
	r.RegisterAlias("data-table", "organisms.table")

	// Register fallbacks for common failures
	r.RegisterFallback("organisms.advanced-table", "organisms.table")
	r.RegisterFallback("molecules.enhanced-field", "molecules.field")
}

// Prop helper functions
func getStringProp(props map[string]interface{}, key, defaultValue string) string {
	if val, exists := props[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getBoolProp(props map[string]interface{}, key string, defaultValue bool) bool {
	if val, exists := props[key]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func getIntProp(props map[string]interface{}, key string, defaultValue int) int {
	if val, exists := props[key]; exists {
		if i, ok := val.(int); ok {
			return i
		}
		if f, ok := val.(float64); ok {
			return int(f)
		}
	}
	return defaultValue
}

func getInterfaceProp(props map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if val, exists := props[key]; exists {
		return val
	}
	return defaultValue
}
