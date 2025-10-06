package core

import (
	"fmt"
	"strings"
	"sync"
)

// ComponentRegistry manages visual components available for composition
type ComponentRegistry struct {
	components map[string]*ComponentDefinition
	categories map[string]*ComponentCategory
	mu         sync.RWMutex
}

// ComponentDefinition defines a visual component for the builder
type ComponentDefinition struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`

	// Visual metadata
	Icon      string   `json:"icon"`
	Preview   string   `json:"preview"`   // Preview image URL
	Thumbnail string   `json:"thumbnail"` // Small icon/thumbnail
	Tags      []string `json:"tags"`

	// Builder behavior
	Draggable bool `json:"draggable"`
	Resizable bool `json:"resizable"`
	Container bool `json:"container"` // Can contain other components

	// Default configuration
	DefaultProps map[string]any `json:"defaultProps"`
	DefaultSize  Size           `json:"defaultSize"`
	MinSize      Size           `json:"minSize"`
	MaxSize      Size           `json:"maxSize"`

	// Property definitions for visual editing
	PropSchema PropertySchema `json:"propSchema"`

	// Composition rules
	AllowedParents  []string `json:"allowedParents,omitempty"`
	AllowedChildren []string `json:"allowedChildren,omitempty"`

	// Template and snippet data
	Template ComponentTemplate `json:"template"`
}

// ComponentCategory organizes components into groups
type ComponentCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Order       int    `json:"order"`
	Collapsed   bool   `json:"collapsed"`
}

// PropertySchema defines editable properties for a component
type PropertySchema struct {
	Properties map[string]PropertyDefinition `json:"properties"`
	Required   []string                      `json:"required,omitempty"`
	Groups     []PropertyGroup               `json:"groups,omitempty"`
}

// PropertyDefinition defines a single property
type PropertyDefinition struct {
	Type        string `json:"type"` // "string", "number", "boolean", "select", "color", etc.
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Default     any    `json:"default,omitempty"`

	// Type-specific options
	Options []PropertyOption `json:"options,omitempty"` // For select/enum types
	Min     *float64         `json:"min,omitempty"`     // For number types
	Max     *float64         `json:"max,omitempty"`     // For number types
	Pattern string           `json:"pattern,omitempty"` // For string validation

	// UI hints
	Placeholder string `json:"placeholder,omitempty"`
	Help        string `json:"help,omitempty"`
	Advanced    bool   `json:"advanced,omitempty"` // Show in advanced section

	// Conditional display
	ShowWhen *PropertyCondition `json:"showWhen,omitempty"`
	HideWhen *PropertyCondition `json:"hideWhen,omitempty"`
}

// PropertyOption for select/enum properties
type PropertyOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Icon  string `json:"icon,omitempty"`
}

// PropertyGroup organizes properties into sections
type PropertyGroup struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Collapsed   bool     `json:"collapsed,omitempty"`
	Properties  []string `json:"properties"`
}

// PropertyCondition for conditional property display
type PropertyCondition struct {
	Property string `json:"property"`
	Operator string `json:"operator"` // "equals", "not_equals", "contains", etc.
	Value    any    `json:"value"`
}

// ComponentTemplate defines the component's template structure
type ComponentTemplate struct {
	HTML     string              `json:"html,omitempty"`
	CSS      string              `json:"css,omitempty"`
	JS       string              `json:"js,omitempty"`
	Props    map[string]any      `json:"props,omitempty"`
	Children []ComponentTemplate `json:"children,omitempty"`
}

// NewComponentRegistry creates a new component registry
func NewComponentRegistry() *ComponentRegistry {
	registry := &ComponentRegistry{
		components: make(map[string]*ComponentDefinition),
		categories: make(map[string]*ComponentCategory),
	}

	registry.initializeDefaultCategories()
	registry.registerDefaultComponents()

	return registry
}

// RegisterComponent adds a component to the registry
func (r *ComponentRegistry) RegisterComponent(component *ComponentDefinition) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if component.Type == "" {
		return fmt.Errorf("component type cannot be empty")
	}

	if component.Name == "" {
		component.Name = component.Type
	}

	// Set defaults
	if component.DefaultSize.Width == 0 {
		component.DefaultSize.Width = 200
	}
	if component.DefaultSize.Height == 0 {
		component.DefaultSize.Height = 100
	}

	if component.MinSize.Width == 0 {
		component.MinSize.Width = 50
	}
	if component.MinSize.Height == 0 {
		component.MinSize.Height = 30
	}

	r.components[component.Type] = component
	return nil
}

// GetComponent retrieves a component definition
func (r *ComponentRegistry) GetComponent(componentType string) *ComponentDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if component, exists := r.components[componentType]; exists {
		// Return a copy to prevent external modification
		componentCopy := *component
		return &componentCopy
	}

	return nil
}

// HasComponent checks if a component type is registered
func (r *ComponentRegistry) HasComponent(componentType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.components[componentType]
	return exists
}

// GetComponents returns all registered components
func (r *ComponentRegistry) GetComponents() []*ComponentDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	components := make([]*ComponentDefinition, 0, len(r.components))
	for _, component := range r.components {
		componentCopy := *component
		components = append(components, &componentCopy)
	}

	return components
}

// GetComponentsByCategory returns components in a specific category
func (r *ComponentRegistry) GetComponentsByCategory(category string) []*ComponentDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var components []*ComponentDefinition
	for _, component := range r.components {
		if component.Category == category {
			componentCopy := *component
			components = append(components, &componentCopy)
		}
	}

	return components
}

// SearchComponents searches for components by name, description, or tags
func (r *ComponentRegistry) SearchComponents(query string) []*ComponentDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = strings.ToLower(query)
	var results []*ComponentDefinition

	for _, component := range r.components {
		// Check name, description, and tags
		if r.componentMatches(component, query) {
			componentCopy := *component
			results = append(results, &componentCopy)
		}
	}

	return results
}

// IsValidComponent checks if a component type is registered
func (r *ComponentRegistry) IsValidComponent(componentType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.components[componentType]
	return exists
}

// GetDefaultProps returns default properties for a component
func (r *ComponentRegistry) GetDefaultProps(componentType string) map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if component, exists := r.components[componentType]; exists {
		// Return a copy
		defaultProps := make(map[string]any)
		for key, value := range component.DefaultProps {
			defaultProps[key] = value
		}
		return defaultProps
	}

	return make(map[string]any)
}

// GetCategories returns all component categories
func (r *ComponentRegistry) GetCategories() []*ComponentCategory {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categories := make([]*ComponentCategory, 0, len(r.categories))
	for _, category := range r.categories {
		categoryCopy := *category
		categories = append(categories, &categoryCopy)
	}

	return categories
}

// Helper methods

func (r *ComponentRegistry) componentMatches(component *ComponentDefinition, query string) bool {
	// Check name
	if strings.Contains(strings.ToLower(component.Name), query) {
		return true
	}

	// Check description
	if strings.Contains(strings.ToLower(component.Description), query) {
		return true
	}

	// Check type
	if strings.Contains(strings.ToLower(component.Type), query) {
		return true
	}

	// Check tags
	for _, tag := range component.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}

	return false
}

// initializeDefaultCategories sets up default component categories
func (r *ComponentRegistry) initializeDefaultCategories() {
	categories := []*ComponentCategory{
		{
			ID:          "atoms",
			Name:        "Atoms",
			Description: "Basic building blocks",
			Icon:        "atom",
			Order:       1,
		},
		{
			ID:          "molecules",
			Name:        "Molecules",
			Description: "Simple component combinations",
			Icon:        "molecule",
			Order:       2,
		},
		{
			ID:          "organisms",
			Name:        "Organisms",
			Description: "Complex component structures",
			Icon:        "organism",
			Order:       3,
		},
		{
			ID:          "layouts",
			Name:        "Layouts",
			Description: "Page structure components",
			Icon:        "layout",
			Order:       4,
		},
		{
			ID:          "forms",
			Name:        "Forms",
			Description: "Form and input components",
			Icon:        "form",
			Order:       5,
		},
		{
			ID:          "data",
			Name:        "Data Display",
			Description: "Tables, lists, and data visualization",
			Icon:        "table",
			Order:       6,
		},
		{
			ID:          "navigation",
			Name:        "Navigation",
			Description: "Menus, tabs, and navigation elements",
			Icon:        "navigation",
			Order:       7,
		},
		{
			ID:          "feedback",
			Name:        "Feedback",
			Description: "Alerts, notifications, and status indicators",
			Icon:        "bell",
			Order:       8,
		},
	}

	for _, category := range categories {
		r.categories[category.ID] = category
	}
}

// registerDefaultComponents registers built-in components
func (r *ComponentRegistry) registerDefaultComponents() {
	// Atoms
	r.RegisterComponent(&ComponentDefinition{
		Type:        "atoms.button",
		Name:        "Button",
		Description: "Basic button component",
		Category:    "atoms",
		Icon:        "square",
		Tags:        []string{"button", "action", "click"},
		Draggable:   true,
		Resizable:   true,
		DefaultProps: map[string]any{
			"label":   "Button",
			"variant": "primary",
			"size":    "medium",
		},
		DefaultSize: Size{Width: 120, Height: 40},
		MinSize:     Size{Width: 60, Height: 30},
		PropSchema: PropertySchema{
			Properties: map[string]PropertyDefinition{
				"label": {
					Type:        "string",
					Label:       "Label",
					Description: "Button text",
					Default:     "Button",
					Placeholder: "Enter button text",
				},
				"variant": {
					Type:    "select",
					Label:   "Variant",
					Default: "primary",
					Options: []PropertyOption{
						{Value: "primary", Label: "Primary"},
						{Value: "secondary", Label: "Secondary"},
						{Value: "danger", Label: "Danger"},
						{Value: "ghost", Label: "Ghost"},
					},
				},
				"size": {
					Type:    "select",
					Label:   "Size",
					Default: "medium",
					Options: []PropertyOption{
						{Value: "small", Label: "Small"},
						{Value: "medium", Label: "Medium"},
						{Value: "large", Label: "Large"},
					},
				},
				"disabled": {
					Type:    "boolean",
					Label:   "Disabled",
					Default: false,
				},
				"icon": {
					Type:        "string",
					Label:       "Icon",
					Description: "Icon name (optional)",
					Placeholder: "e.g., plus, edit, delete",
				},
			},
			Groups: []PropertyGroup{
				{
					ID:         "content",
					Label:      "Content",
					Properties: []string{"label", "icon"},
				},
				{
					ID:         "appearance",
					Label:      "Appearance",
					Properties: []string{"variant", "size"},
				},
				{
					ID:         "behavior",
					Label:      "Behavior",
					Properties: []string{"disabled"},
				},
			},
		},
	})

	r.RegisterComponent(&ComponentDefinition{
		Type:        "atoms.input",
		Name:        "Input",
		Description: "Text input field",
		Category:    "forms",
		Icon:        "type",
		Tags:        []string{"input", "text", "form"},
		Draggable:   true,
		Resizable:   true,
		DefaultProps: map[string]any{
			"type":        "text",
			"placeholder": "Enter text...",
			"label":       "Input Field",
		},
		DefaultSize: Size{Width: 250, Height: 40},
		MinSize:     Size{Width: 100, Height: 30},
		PropSchema: PropertySchema{
			Properties: map[string]PropertyDefinition{
				"label": {
					Type:        "string",
					Label:       "Label",
					Description: "Field label",
					Placeholder: "Enter field label",
				},
				"type": {
					Type:    "select",
					Label:   "Type",
					Default: "text",
					Options: []PropertyOption{
						{Value: "text", Label: "Text"},
						{Value: "email", Label: "Email"},
						{Value: "password", Label: "Password"},
						{Value: "number", Label: "Number"},
						{Value: "tel", Label: "Phone"},
						{Value: "url", Label: "URL"},
					},
				},
				"placeholder": {
					Type:        "string",
					Label:       "Placeholder",
					Description: "Placeholder text",
					Placeholder: "Enter placeholder...",
				},
				"required": {
					Type:    "boolean",
					Label:   "Required",
					Default: false,
				},
				"disabled": {
					Type:    "boolean",
					Label:   "Disabled",
					Default: false,
				},
			},
		},
	})

	// Molecules
	r.RegisterComponent(&ComponentDefinition{
		Type:        "molecules.card",
		Name:        "Card",
		Description: "Content container with border and shadow",
		Category:    "molecules",
		Icon:        "square",
		Tags:        []string{"card", "container", "content"},
		Draggable:   true,
		Resizable:   true,
		Container:   true,
		DefaultProps: map[string]any{
			"title":   "Card Title",
			"content": "Card content goes here...",
			"shadow":  "medium",
		},
		DefaultSize: Size{Width: 300, Height: 200},
		MinSize:     Size{Width: 200, Height: 100},
		PropSchema: PropertySchema{
			Properties: map[string]PropertyDefinition{
				"title": {
					Type:        "string",
					Label:       "Title",
					Description: "Card title",
					Placeholder: "Enter card title",
				},
				"content": {
					Type:        "textarea",
					Label:       "Content",
					Description: "Card content",
					Placeholder: "Enter card content...",
				},
				"shadow": {
					Type:    "select",
					Label:   "Shadow",
					Default: "medium",
					Options: []PropertyOption{
						{Value: "none", Label: "None"},
						{Value: "small", Label: "Small"},
						{Value: "medium", Label: "Medium"},
						{Value: "large", Label: "Large"},
					},
				},
			},
		},
	})

	// Organisms
	r.RegisterComponent(&ComponentDefinition{
		Type:        "organisms.table",
		Name:        "Data Table",
		Description: "Advanced data table with sorting and filtering",
		Category:    "data",
		Icon:        "table",
		Tags:        []string{"table", "data", "list", "grid"},
		Draggable:   true,
		Resizable:   true,
		DefaultProps: map[string]any{
			"title":      "Data Table",
			"pageSize":   20,
			"sortable":   true,
			"filterable": true,
			"searchable": true,
		},
		DefaultSize: Size{Width: 600, Height: 400},
		MinSize:     Size{Width: 300, Height: 200},
		PropSchema: PropertySchema{
			Properties: map[string]PropertyDefinition{
				"title": {
					Type:        "string",
					Label:       "Title",
					Description: "Table title",
					Placeholder: "Enter table title",
				},
				"pageSize": {
					Type:    "number",
					Label:   "Page Size",
					Default: 20,
					Min:     func() *float64 { v := 5.0; return &v }(),
					Max:     func() *float64 { v := 100.0; return &v }(),
				},
				"sortable": {
					Type:    "boolean",
					Label:   "Sortable",
					Default: true,
				},
				"filterable": {
					Type:    "boolean",
					Label:   "Filterable",
					Default: true,
				},
				"searchable": {
					Type:    "boolean",
					Label:   "Searchable",
					Default: true,
				},
			},
		},
	})

	r.RegisterComponent(&ComponentDefinition{
		Type:        "organisms.form",
		Name:        "Form",
		Description: "Complete form with validation and submission",
		Category:    "forms",
		Icon:        "form",
		Tags:        []string{"form", "input", "validation", "submit"},
		Draggable:   true,
		Resizable:   true,
		Container:   true,
		DefaultProps: map[string]any{
			"title":      "Form",
			"layout":     "vertical",
			"validation": true,
		},
		DefaultSize: Size{Width: 400, Height: 300},
		MinSize:     Size{Width: 250, Height: 200},
		PropSchema: PropertySchema{
			Properties: map[string]PropertyDefinition{
				"title": {
					Type:        "string",
					Label:       "Title",
					Description: "Form title",
					Placeholder: "Enter form title",
				},
				"layout": {
					Type:    "select",
					Label:   "Layout",
					Default: "vertical",
					Options: []PropertyOption{
						{Value: "vertical", Label: "Vertical"},
						{Value: "horizontal", Label: "Horizontal"},
						{Value: "inline", Label: "Inline"},
					},
				},
				"validation": {
					Type:    "boolean",
					Label:   "Enable Validation",
					Default: true,
				},
			},
		},
	})

	// Layout components
	r.RegisterComponent(&ComponentDefinition{
		Type:        "layouts.grid",
		Name:        "Grid Layout",
		Description: "CSS Grid layout container",
		Category:    "layouts",
		Icon:        "grid",
		Tags:        []string{"grid", "layout", "container"},
		Draggable:   true,
		Resizable:   true,
		Container:   true,
		DefaultProps: map[string]any{
			"columns": "2",
			"gap":     "1rem",
		},
		DefaultSize: Size{Width: 500, Height: 300},
		MinSize:     Size{Width: 200, Height: 100},
		PropSchema: PropertySchema{
			Properties: map[string]PropertyDefinition{
				"columns": {
					Type:        "string",
					Label:       "Columns",
					Description: "Grid template columns (CSS)",
					Default:     "2",
					Placeholder: "e.g., 1fr 2fr, repeat(3, 1fr), 200px 1fr",
				},
				"rows": {
					Type:        "string",
					Label:       "Rows",
					Description: "Grid template rows (CSS)",
					Placeholder: "e.g., auto 1fr, repeat(2, 100px)",
				},
				"gap": {
					Type:        "string",
					Label:       "Gap",
					Description: "Grid gap",
					Default:     "1rem",
					Placeholder: "e.g., 1rem, 10px 20px",
				},
			},
		},
	})
}
