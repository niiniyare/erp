package ui

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ComponentType represents the type of UI component
type ComponentType string

const (
	// Form Components
	ComponentForm       ComponentType = "form"
	ComponentInput      ComponentType = "input"
	ComponentTextarea   ComponentType = "textarea"
	ComponentSelect     ComponentType = "select"
	ComponentCheckbox   ComponentType = "checkbox"
	ComponentRadio      ComponentType = "radio"
	ComponentButton     ComponentType = "button"
	ComponentDatePicker ComponentType = "date-picker"
	ComponentTimePicker ComponentType = "time-picker"
	ComponentFileUpload ComponentType = "file-upload"

	// Layout Components
	ComponentContainer ComponentType = "container"
	ComponentCard      ComponentType = "card"
	ComponentPanel     ComponentType = "panel"
	ComponentTabs      ComponentType = "tabs"
	ComponentModal     ComponentType = "modal"
	ComponentDrawer    ComponentType = "drawer"

	// Data Display
	ComponentTable ComponentType = "table"
	ComponentList  ComponentType = "list"
	ComponentTree  ComponentType = "tree"
	ComponentChart ComponentType = "chart"
	ComponentBadge ComponentType = "badge"
	ComponentTag   ComponentType = "tag"

	// Navigation
	ComponentNav        ComponentType = "nav"
	ComponentBreadcrumb ComponentType = "breadcrumb"
	ComponentPagination ComponentType = "pagination"
)

// Size represents component size variants
type Size string

const (
	SizeXS Size = "xs"
	SizeSM Size = "sm"
	SizeMD Size = "md"
	SizeLG Size = "lg"
	SizeXL Size = "xl"
)

// Variant represents component style variants
type Variant string

const (
	VariantPrimary   Variant = "primary"
	VariantSecondary Variant = "secondary"
	VariantSuccess   Variant = "success"
	VariantDanger    Variant = "danger"
	VariantWarning   Variant = "warning"
	VariantInfo      Variant = "info"
	VariantLight     Variant = "light"
	VariantDark      Variant = "dark"
)

// Position represents positioning options
type Position string

const (
	PositionTop    Position = "top"
	PositionBottom Position = "bottom"
	PositionLeft   Position = "left"
	PositionRight  Position = "right"
	PositionCenter Position = "center"
)

// BaseComponent represents the common properties of all UI components
type BaseComponent struct {
	// Identity
	ID   string        `json:"id" validate:"required"`
	Type ComponentType `json:"type" validate:"required"`
	Name string        `json:"name,omitempty"`

	// Display
	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`

	// Styling
	Class   string  `json:"class,omitempty"`
	Style   string  `json:"style,omitempty"`
	Size    Size    `json:"size,omitempty"`
	Variant Variant `json:"variant,omitempty"`
	Width   string  `json:"width,omitempty"`
	Height  string  `json:"height,omitempty"`

	// Behavior
	Disabled bool `json:"disabled,omitempty"`
	Hidden   bool `json:"hidden,omitempty"`
	Required bool `json:"required,omitempty"`
	ReadOnly bool `json:"readonly,omitempty"`

	// Accessibility
	AriaLabel       string `json:"aria_label,omitempty"`
	AriaDescribedBy string `json:"aria_described_by,omitempty"`
	TabIndex        int    `json:"tab_index,omitempty"`

	// Events
	OnClick  string `json:"on_click,omitempty"`
	OnChange string `json:"on_change,omitempty"`
	OnFocus  string `json:"on_focus,omitempty"`
	OnBlur   string `json:"on_blur,omitempty"`

	// Multi-tenant context
	TenantID uuid.UUID `json:"tenant_id,omitempty"`

	// Metadata
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// Component represents a complete UI component with its configuration
type Component struct {
	BaseComponent
	Config    json.RawMessage `json:"config,omitempty"`
	Children  []Component     `json:"children,omitempty"`
	Validator *Validator      `json:"validator,omitempty"`
}

// Validator represents validation rules for a component
type Validator struct {
	Required    bool     `json:"required,omitempty"`
	MinLength   *int     `json:"min_length,omitempty"`
	MaxLength   *int     `json:"max_length,omitempty"`
	Min         *float64 `json:"min,omitempty"`
	Max         *float64 `json:"max,omitempty"`
	Pattern     string   `json:"pattern,omitempty"`
	CustomRules []string `json:"custom_rules,omitempty"`
	Message     string   `json:"message,omitempty"`
}

// ComponentBuilder provides a fluent interface for building components
type ComponentBuilder struct {
	component Component
}

// NewComponent creates a new component builder
func NewComponent(componentType ComponentType, id string) *ComponentBuilder {
	return &ComponentBuilder{
		component: Component{
			BaseComponent: BaseComponent{
				ID:   id,
				Type: componentType,
			},
		},
	}
}

// WithLabel sets the component label
func (b *ComponentBuilder) WithLabel(label string) *ComponentBuilder {
	b.component.Label = label
	return b
}

// WithName sets the component name
func (b *ComponentBuilder) WithName(name string) *ComponentBuilder {
	b.component.Name = name
	return b
}

// WithDescription sets the component description
func (b *ComponentBuilder) WithDescription(desc string) *ComponentBuilder {
	b.component.Description = desc
	return b
}

// WithSize sets the component size
func (b *ComponentBuilder) WithSize(size Size) *ComponentBuilder {
	b.component.Size = size
	return b
}

// WithVariant sets the component variant
func (b *ComponentBuilder) WithVariant(variant Variant) *ComponentBuilder {
	b.component.Variant = variant
	return b
}

// WithClass sets the component CSS class
func (b *ComponentBuilder) WithClass(class string) *ComponentBuilder {
	b.component.Class = class
	return b
}

// Disabled marks the component as disabled
func (b *ComponentBuilder) Disabled() *ComponentBuilder {
	b.component.Disabled = true
	return b
}

// Required marks the component as required
func (b *ComponentBuilder) Required() *ComponentBuilder {
	b.component.Required = true
	return b
}

// WithTenantID sets the tenant context
func (b *ComponentBuilder) WithTenantID(tenantID uuid.UUID) *ComponentBuilder {
	b.component.TenantID = tenantID
	return b
}

// WithValidator sets the component validator
func (b *ComponentBuilder) WithValidator(validator *Validator) *ComponentBuilder {
	b.component.Validator = validator
	return b
}

// WithChildren adds child components
func (b *ComponentBuilder) WithChildren(children ...Component) *ComponentBuilder {
	b.component.Children = append(b.component.Children, children...)
	return b
}

// WithConfig sets the component configuration
func (b *ComponentBuilder) WithConfig(config any) *ComponentBuilder {
	if configBytes, err := json.Marshal(config); err == nil {
		b.component.Config = configBytes
	}
	return b
}

// Build returns the constructed component
func (b *ComponentBuilder) Build() Component {
	now := time.Now()
	if b.component.CreatedAt == nil {
		b.component.CreatedAt = &now
	}
	b.component.UpdatedAt = &now
	return b.component
}

// ComponentRegistry manages component types and their configurations
type ComponentRegistry interface {
	Register(componentType ComponentType, factory ComponentFactory)
	Create(ctx context.Context, componentType ComponentType, config map[string]any) (Component, error)
	GetTypes() []ComponentType
	Validate(ctx context.Context, component Component) error
}

// ComponentFactory creates components of a specific type
type ComponentFactory interface {
	Create(ctx context.Context, config map[string]any) (Component, error)
	Validate(ctx context.Context, component Component) error
	GetSchema() ComponentSchema
}

// ComponentSchema defines the structure and validation rules for a component type
type ComponentSchema struct {
	Type        ComponentType            `json:"type"`
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	Properties  map[string]Property      `json:"properties"`
	Required    []string                 `json:"required"`
	Examples    []map[string]any `json:"examples"`
}

// Property defines a component property schema
type Property struct {
	Type        string      `json:"type"`
	Format      string      `json:"format,omitempty"`
	Description string      `json:"description,omitempty"`
	Default     any `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	MinLength   *int        `json:"min_length,omitempty"`
	MaxLength   *int        `json:"max_length,omitempty"`
	Min         *float64    `json:"min,omitempty"`
	Max         *float64    `json:"max,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
}

