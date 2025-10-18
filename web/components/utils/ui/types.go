package ui

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/pkg/schema/ui/css"
)

// ============================================================================
// CORE COMPONENT TYPES
// ============================================================================

// ComponentType represents the type of UI component
// These are the primary component types supported by the ERP system
type ComponentType string

const (
	// Form Components - Interactive input elements
	ComponentForm       ComponentType = "form"        // Form container with validation
	ComponentInput      ComponentType = "input"       // Single-line text input
	ComponentTextarea   ComponentType = "textarea"    // Multi-line text input
	ComponentSelect     ComponentType = "select"      // Dropdown selection
	ComponentCheckbox   ComponentType = "checkbox"    // Boolean checkbox input
	ComponentRadio      ComponentType = "radio"       // Radio button group
	ComponentButton     ComponentType = "button"      // Action button
	ComponentDatePicker ComponentType = "date-picker" // Date/time selection
	ComponentTimePicker ComponentType = "time-picker" // Time-only selection
	ComponentFileUpload ComponentType = "file-upload" // File upload component

	// Layout Components - Structural containers
	ComponentContainer ComponentType = "container" // General purpose container
	ComponentCard      ComponentType = "card"      // Content card with optional header/footer
	ComponentPanel     ComponentType = "panel"     // Collapsible panel
	ComponentTabs      ComponentType = "tabs"      // Tabbed interface
	ComponentModal     ComponentType = "modal"     // Overlay dialog
	ComponentDrawer    ComponentType = "drawer"    // Side panel

	// Data Display Components - Information presentation
	ComponentTable ComponentType = "table" // Data table with sorting/filtering
	ComponentList  ComponentType = "list"  // Simple list display
	ComponentTree  ComponentType = "tree"  // Hierarchical tree view
	ComponentChart ComponentType = "chart" // Data visualization charts
	ComponentBadge ComponentType = "badge" // Status indicator
	ComponentTag   ComponentType = "tag"   // Label/category tag

	// Navigation Components - User navigation
	ComponentNav        ComponentType = "nav"        // Navigation menu
	ComponentBreadcrumb ComponentType = "breadcrumb" // Breadcrumb trail
	ComponentPagination ComponentType = "pagination" // Page navigation
)

// ============================================================================
// COMMON ENUMS & CONSTANTS
// ============================================================================

// Size represents standard component sizing options
// Follows common UI library conventions (xs, sm, md, lg, xl)
type Size string

const (
	SizeXS Size = "xs" // Extra small (minimal padding, compact)
	SizeSM Size = "sm" // Small (reduced padding)
	SizeMD Size = "md" // Medium (default size)
	SizeLG Size = "lg" // Large (increased padding)
	SizeXL Size = "xl" // Extra large (maximum padding)
)

// Variant represents component style/color variants
// Based on semantic color system for consistent UI
type Variant string

const (
	VariantPrimary   Variant = "primary"   // Primary brand color (blue)
	VariantSecondary Variant = "secondary" // Secondary color (gray)
	VariantSuccess   Variant = "success"   // Success state (green)
	VariantDanger    Variant = "danger"    // Error/destructive actions (red)
	VariantWarning   Variant = "warning"   // Warning state (orange/yellow)
	VariantInfo      Variant = "info"      // Informational (light blue)
	VariantLight     Variant = "light"     // Light background
	VariantDark      Variant = "dark"      // Dark background
)

// Position represents positioning/alignment options
// Used for tooltips, dropdowns, modals, etc.
type Position string

const (
	PositionTop    Position = "top"    // Above the element
	PositionBottom Position = "bottom" // Below the element
	PositionLeft   Position = "left"   // Left of the element
	PositionRight  Position = "right"  // Right of the element
	PositionCenter Position = "center" // Centered
)

// Alignment represents text/content alignment
// Standard CSS text-align values
type Alignment string

const (
	AlignLeft   Alignment = "left"   // Left-aligned content
	AlignCenter Alignment = "center" // Center-aligned content
	AlignRight  Alignment = "right"  // Right-aligned content
)

// ============================================================================
// SHARED DATA TYPES
// ============================================================================

// Option represents a selectable option in dropdowns, radio groups, etc.
// Reused across Select, Radio, Checkbox components
type Option struct {
	Value    string `json:"value" validate:"required"` // The actual value submitted
	Label    string `json:"label" validate:"required"` // Display text for the option
	Disabled bool   `json:"disabled,omitempty"`        // Whether option is disabled
	Group    string `json:"group,omitempty"`           // Option group (for grouped selects)
	Icon     string `json:"icon,omitempty"`            // Optional icon identifier
	Badge    string `json:"badge,omitempty"`           // Optional badge/count display
}

// Validator represents validation rules for form components
// Centralized validation configuration reused across all form fields
type Validator struct {
	// Basic validation
	Required bool `json:"required,omitempty"` // Field is required

	// Length validation
	MinLength *int `json:"min_length,omitempty"` // Minimum character length
	MaxLength *int `json:"max_length,omitempty"` // Maximum character length

	// Numeric validation
	Min *float64 `json:"min,omitempty"` // Minimum numeric value
	Max *float64 `json:"max,omitempty"` // Maximum numeric value

	// Pattern validation
	Pattern string `json:"pattern,omitempty"` // Regular expression pattern

	// Custom validation
	CustomRules []string `json:"custom_rules,omitempty"` // Custom validation rule names

	// Error messaging
	Message string `json:"message,omitempty"` // Custom validation error message
}

// Action represents an actionable button or link
// Reused in tables, cards, forms, and other interactive components
type Action struct {
	Key      string         `json:"key" validate:"required"`   // Unique identifier
	Label    string         `json:"label" validate:"required"` // Display text
	Icon     string         `json:"icon,omitempty"`            // Icon identifier
	Type     ActionType     `json:"type,omitempty"`            // Action type
	Variant  Variant        `json:"variant,omitempty"`         // Visual style
	OnClick  string         `json:"on_click,omitempty"`        // Click handler
	Disabled string         `json:"disabled,omitempty"`        // Disable condition
	Confirm  *ConfirmDialog `json:"confirm,omitempty"`         // Confirmation dialog
}

// ActionType represents different action behaviors
type ActionType string

const (
	ActionButton   ActionType = "button"   // Regular button action
	ActionLink     ActionType = "link"     // Navigation link
	ActionDropdown ActionType = "dropdown" // Dropdown menu
)

// ConfirmDialog represents a confirmation dialog for destructive actions
type ConfirmDialog struct {
	Title       string `json:"title,omitempty"`       // Dialog title
	Description string `json:"description,omitempty"` // Warning message
	OkText      string `json:"ok_text,omitempty"`     // Confirm button text (default: "OK")
	CancelText  string `json:"cancel_text,omitempty"` // Cancel button text (default: "Cancel")
}

// ============================================================================
// COMPONENT BASE TYPES
// ============================================================================

// BaseComponent contains common properties shared by all UI components
// This provides consistent structure and multi-tenant support
type BaseComponent struct {
	// Core Identity
	ID   string        `json:"id" validate:"required"`   // Unique component identifier
	Type ComponentType `json:"type" validate:"required"` // Component type
	Name string        `json:"name,omitempty"`           // Form field name (for inputs)

	// Display Properties
	Label       string `json:"label,omitempty"`       // Display label/title
	Description string `json:"description,omitempty"` // Help text or description
	Placeholder string `json:"placeholder,omitempty"` // Input placeholder text

	// Visual Styling
	Class   string  `json:"class,omitempty"`   // CSS classes
	Style   string  `json:"style,omitempty"`   // Inline CSS styles
	Size    Size    `json:"size,omitempty"`    // Component size
	Variant Variant `json:"variant,omitempty"` // Visual variant/theme
	Width   string  `json:"width,omitempty"`   // Component width
	Height  string  `json:"height,omitempty"`  // Component height

	// Behavioral States
	Disabled bool `json:"disabled,omitempty"` // Disable user interaction
	Hidden   bool `json:"hidden,omitempty"`   // Hide component
	Required bool `json:"required,omitempty"` // Required field (forms)
	ReadOnly bool `json:"readonly,omitempty"` // Read-only mode

	// Accessibility Support
	AriaLabel       string `json:"aria_label,omitempty"`        // ARIA label for screen readers
	AriaDescribedBy string `json:"aria_described_by,omitempty"` // ARIA described-by reference
	TabIndex        int    `json:"tab_index,omitempty"`         // Tab order index

	// Event Handlers
	OnClick  string `json:"on_click,omitempty"`  // Click event handler
	OnChange string `json:"on_change,omitempty"` // Change event handler
	OnFocus  string `json:"on_focus,omitempty"`  // Focus event handler
	OnBlur   string `json:"on_blur,omitempty"`   // Blur event handler

	// Multi-Tenant Support
	TenantID uuid.UUID `json:"tenant_id,omitempty"` // Tenant context

	// Metadata
	CreatedAt *time.Time `json:"created_at,omitempty"` // Creation timestamp
	UpdatedAt *time.Time `json:"updated_at,omitempty"` // Last update timestamp
}

// Component represents a complete UI component with configuration and children
// This is the main component structure used throughout the system
type Component struct {
	BaseComponent                 // Embedded base properties
	Config        json.RawMessage `json:"config,omitempty"`    // Component-specific configuration
	Children      []Component     `json:"children,omitempty"`  // Child components
	Validator     *Validator      `json:"validator,omitempty"` // Validation rules
	Styles        *css.Styles     `json:"styles,omitempty"`    // CSS styling properties

	// Security enhancements for military-grade multi-tenant architecture
	metadata        map[string]string `json:"-"` // Internal metadata, never serialized
	encrypted       bool              `json:"-"` // Encryption status
	encryptedConfig []byte            `json:"-"` // Encrypted configuration data
}

// ============================================================================
// COMPONENT BUILDER PATTERN
// ============================================================================

// ComponentBuilder provides a fluent interface for constructing components
// Implements the Builder pattern for easy component creation
type ComponentBuilder struct {
	component Component
}

// NewComponent creates a new component builder with the specified type and ID
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

// Fluent Builder Methods for Common Properties

// WithLabel sets the component display label
func (b *ComponentBuilder) WithLabel(label string) *ComponentBuilder {
	b.component.Label = label
	return b
}

// WithName sets the form field name (for form controls)
func (b *ComponentBuilder) WithName(name string) *ComponentBuilder {
	b.component.Name = name
	return b
}

// WithDescription sets the help text or description
func (b *ComponentBuilder) WithDescription(desc string) *ComponentBuilder {
	b.component.Description = desc
	return b
}

// WithPlaceholder sets the input placeholder text
func (b *ComponentBuilder) WithPlaceholder(placeholder string) *ComponentBuilder {
	b.component.Placeholder = placeholder
	return b
}

// WithSize sets the component size
func (b *ComponentBuilder) WithSize(size Size) *ComponentBuilder {
	b.component.Size = size
	return b
}

// WithVariant sets the visual style variant
func (b *ComponentBuilder) WithVariant(variant Variant) *ComponentBuilder {
	b.component.Variant = variant
	return b
}

// WithClass sets custom CSS classes
func (b *ComponentBuilder) WithClass(class string) *ComponentBuilder {
	b.component.Class = class
	return b
}

// Disabled marks the component as disabled
func (b *ComponentBuilder) Disabled() *ComponentBuilder {
	b.component.Disabled = true
	return b
}

// Hidden marks the component as hidden
func (b *ComponentBuilder) Hidden() *ComponentBuilder {
	b.component.Hidden = true
	return b
}

// Required marks the component as required (for form fields)
func (b *ComponentBuilder) Required() *ComponentBuilder {
	b.component.Required = true
	return b
}

// ReadOnly marks the component as read-only
func (b *ComponentBuilder) ReadOnly() *ComponentBuilder {
	b.component.ReadOnly = true
	return b
}

// WithTenantID sets the tenant context for multi-tenant deployments
func (b *ComponentBuilder) WithTenantID(tenantID uuid.UUID) *ComponentBuilder {
	b.component.TenantID = tenantID
	return b
}

// WithValidator sets validation rules for form components
func (b *ComponentBuilder) WithValidator(validator *Validator) *ComponentBuilder {
	b.component.Validator = validator
	return b
}

// WithStyles sets CSS styling properties for the component
func (b *ComponentBuilder) WithStyles(styles *css.Styles) *ComponentBuilder {
	b.component.Styles = styles
	return b
}

// WithChildren adds child components (for container components)
func (b *ComponentBuilder) WithChildren(children ...Component) *ComponentBuilder {
	b.component.Children = append(b.component.Children, children...)
	return b
}

// WithConfig sets component-specific configuration
func (b *ComponentBuilder) WithConfig(config any) *ComponentBuilder {
	if configBytes, err := json.Marshal(config); err == nil {
		b.component.Config = configBytes
	}
	return b
}

// WithMetadata sets internal security metadata
func (b *ComponentBuilder) WithMetadata(key, value string) *ComponentBuilder {
	if b.component.metadata == nil {
		b.component.metadata = make(map[string]string)
	}
	b.component.metadata[key] = value
	return b
}

// Build returns the constructed component with timestamps
func (b *ComponentBuilder) Build() Component {
	now := time.Now()
	if b.component.CreatedAt == nil {
		b.component.CreatedAt = &now
	}
	b.component.UpdatedAt = &now
	return b.component
}

// ============================================================================
// COMPONENT REGISTRY INTERFACES
// ============================================================================

// ComponentRegistry manages component types and their configurations
// Provides a centralized system for component creation and validation
type ComponentRegistry interface {
	// Register a component factory for a specific type
	Register(componentType ComponentType, factory ComponentFactory)

	// Create a component using the registered factory
	Create(ctx context.Context, componentType ComponentType, config map[string]any) (Component, error)

	// Get all registered component types
	GetTypes() []ComponentType

	// Validate a component using its factory
	Validate(ctx context.Context, component Component) error
}

// ComponentFactory creates and validates components of a specific type
// Each component type implements this interface for creation and validation
type ComponentFactory interface {
	// Create a new component instance with the given configuration
	Create(ctx context.Context, config map[string]any) (Component, error)

	// Validate a component instance
	Validate(ctx context.Context, component Component) error

	// Get the schema definition for this component type
	GetSchema() ComponentSchema
}

// ComponentSchema defines the structure and validation rules for a component type
// Used for documentation, validation, and UI generation
type ComponentSchema struct {
	Type        ComponentType       `json:"type"`        // Component type identifier
	Title       string              `json:"title"`       // Human-readable title
	Description string              `json:"description"` // Detailed description
	Properties  map[string]Property `json:"properties"`  // Configuration properties
	Required    []string            `json:"required"`    // Required property names
	Examples    []map[string]any    `json:"examples"`    // Usage examples
}

// Property defines a component configuration property
// Describes the type, constraints, and documentation for config properties
type Property struct {
	Type        string   `json:"type"`                  // Property data type (string, number, boolean, etc.)
	Format      string   `json:"format,omitempty"`      // Format specification (email, url, etc.)
	Description string   `json:"description,omitempty"` // Property description
	Default     any      `json:"default,omitempty"`     // Default value
	Enum        []string `json:"enum,omitempty"`        // Allowed values (for enums)
	MinLength   *int     `json:"min_length,omitempty"`  // Minimum string length
	MaxLength   *int     `json:"max_length,omitempty"`  // Maximum string length
	Min         *float64 `json:"min,omitempty"`         // Minimum numeric value
	Max         *float64 `json:"max,omitempty"`         // Maximum numeric value
	Pattern     string   `json:"pattern,omitempty"`     // Regular expression pattern
}

// ============================================================================
// UTILITY TYPES
// ============================================================================

// InputType represents different HTML input types
// Used by the Input component for type-specific behavior
type InputType string

const (
	InputText     InputType = "text"     // Plain text input
	InputEmail    InputType = "email"    // Email input with validation
	InputPassword InputType = "password" // Password input (masked)
	InputNumber   InputType = "number"   // Numeric input
	InputTel      InputType = "tel"      // Telephone number input
	InputURL      InputType = "url"      // URL input with validation
	InputSearch   InputType = "search"   // Search input
	InputHidden   InputType = "hidden"   // Hidden form field
	InputDate     InputType = "date"     // Date picker
	InputDateTime InputType = "datetime" // Date and time picker
	InputTime     InputType = "time"     // Time picker
	InputMonth    InputType = "month"    // Month and year picker
	InputWeek     InputType = "week"     // Week picker
	InputColor    InputType = "color"    // Color picker
	InputRange    InputType = "range"    // Slider control
	InputFile     InputType = "file"     // File upload
	InputCheckbox InputType = "checkbox" // Checkbox
	InputRadio    InputType = "radio"    // Radio button
	InputSubmit   InputType = "submit"   // Form submit button
	InputReset    InputType = "reset"    // Form reset button
	InputButton   InputType = "button"   // Generic button
	InputImage    InputType = "image"    // Image submit button
	// Additional specialized input types
	InputDateTimeLocal InputType = "datetime-local" // Local date and time picker
	InputTextarea      InputType = "textarea"       // Multi-line text input
	InputSelect        InputType = "select"         // Dropdown select
	InputMultiSelect   InputType = "multiselect"    // Multiple selection
	InputToggle        InputType = "toggle"         // Switch toggle
	InputCurrency      InputType = "currency"       // Currency input
	InputPercent       InputType = "percent"        // Percentage input
	// Modern/experimental input types
	InputOtp     InputType = "otp"     // One-time password input
	InputTags    InputType = "tags"    // Tag input
	InputRating  InputType = "rating"  // Star rating input
	InputSlider  InputType = "slider"  // Enhanced slider
	InputCaptcha InputType = "captcha" // CAPTCHA input

)

// ButtonType represents HTML button types
// Used by Button component for form behavior
type ButtonType string

const (
	// Standard HTML button types
	ButtonSubmit ButtonType = "submit" // Submit form button
	ButtonButton ButtonType = "button" // Regular button with no default behavior
	ButtonReset  ButtonType = "reset"  // Reset form button

	// UI button types
	ButtonMenu      ButtonType = "menu"      // Opens a menu (HTML5)
	ButtonLink      ButtonType = "link"      // Button that behaves like a link
	ButtonIcon      ButtonType = "icon"      // Icon-only button
	ButtonLoading   ButtonType = "loading"   // Button in loading state
	ButtonDisabled  ButtonType = "disabled"  // Disabled button state
	ButtonOutline   ButtonType = "outline"   // Outline style button
	ButtonGhost     ButtonType = "ghost"     // Ghost/transparent button
	ButtonDanger    ButtonType = "danger"    // Destructive action button
	ButtonSuccess   ButtonType = "success"   // Success state button
	ButtonWarning   ButtonType = "warning"   // Warning state button
	ButtonInfo      ButtonType = "info"      // Informational button
	ButtonLight     ButtonType = "light"     // Light theme button
	ButtonDark      ButtonType = "dark"      // Dark theme button
	ButtonPrimary   ButtonType = "primary"   // Primary action button
	ButtonSecondary ButtonType = "secondary" // Secondary action button
	ButtonTertiary  ButtonType = "tertiary"  // Tertiary action button
	ButtonText      ButtonType = "text"      // Text-only button (minimal styling)
	ButtonFab       ButtonType = "fab"       // Floating Action Button
	ButtonPill      ButtonType = "pill"      // Pill-shaped button
	ButtonCircle    ButtonType = "circle"    // Circular button
	ButtonSquare    ButtonType = "square"    // Square button
	ButtonBlock     ButtonType = "block"     // Full-width block button
	ButtonSmall     ButtonType = "small"     // Small size button
	ButtonLarge     ButtonType = "large"     // Large size button
	ButtonXLarge    ButtonType = "xlarge"    // Extra large size button
)

// DataType represents different data display types
// Used by Table component for column formatting
type DataType string

const (
	// Basic Data Types
	DataTypeText     DataType = "text"     // Plain text display
	DataTypeNumber   DataType = "number"   // Numeric formatting
	DataTypeDate     DataType = "date"     // Date formatting
	DataTypeDateTime DataType = "datetime" // Date/time formatting
	DataTypeBoolean  DataType = "boolean"  // Boolean checkbox/switch
	DataTypeCurrency DataType = "currency" // Currency formatting
	DataTypePercent  DataType = "percent"  // Percentage formatting

	// Media & Content Types
	DataTypeImage  DataType = "image"  // Image display
	DataTypeAvatar DataType = "avatar" // Avatar/circular image
	DataTypeIcon   DataType = "icon"   // Icon display
	DataTypeFile   DataType = "file"   // File with download link
	DataTypeVideo  DataType = "video"  // Video embed/thumbnail
	DataTypeAudio  DataType = "audio"  // Audio player

	// Interactive Types
	DataTypeLink     DataType = "link"     // Clickable link
	DataTypeEmail    DataType = "email"    // Email link
	DataTypePhone    DataType = "phone"    // Phone link
	DataTypeButton   DataType = "button"   // Action button
	DataTypeActions  DataType = "actions"  // Multiple action buttons
	DataTypeCheckbox DataType = "checkbox" // Checkbox input
	DataTypeRadio    DataType = "radio"    // Radio button
	DataTypeSwitch   DataType = "switch"   // Toggle switch

	// Status & Indicator Types
	DataTypeBadge    DataType = "badge"    // Status badge
	DataTypeTag      DataType = "tag"      // Tag/chip
	DataTypeStatus   DataType = "status"   // Status indicator
	DataTypeProgress DataType = "progress" // Progress bar
	DataTypeRating   DataType = "rating"   // Star rating

	// Specialized Formatting
	DataTypeCode     DataType = "code"     // Code snippet
	DataTypeJSON     DataType = "json"     // JSON formatting
	DataTypeHTML     DataType = "html"     // HTML content
	DataTypeMarkdown DataType = "markdown" // Markdown content

	// Time & Duration
	DataTypeTime     DataType = "time"     // Time only formatting
	DataTypeDuration DataType = "duration" // Duration formatting
	DataTypeRelative DataType = "relative" // Relative time (e.g., "2 hours ago")

	// Custom Components
	DataTypeCustom DataType = "custom" // Custom render component
	DataTypeSlot   DataType = "slot"   // Slot for custom content

	// Layout & Structure
	DataTypeEmpty    DataType = "empty"    // Empty state
	DataTypeSkeleton DataType = "skeleton" // Loading skeleton
	DataTypeExpand   DataType = "expand"   // Expand/collapse row
	DataTypeSelect   DataType = "select"   // Row selection checkbox

	// Additional specialized data types for specific use cases
	DataTypeColor    DataType = "color"    // Color swatch
	DataTypeGradient DataType = "gradient" // Gradient display
	DataTypeQRCode   DataType = "qrcode"   // QR code display
	DataTypeBarcode  DataType = "barcode"  // Barcode display
	DataTypeMap      DataType = "map"      // Mini map display
	DataTypeChart    DataType = "chart"    // Mini chart/sparkline
	DataTypeRange    DataType = "range"    // Range slider display
	DataTypePassword DataType = "password" // Password (masked)
	DataTypeCopyable DataType = "copyable" // Copy-to-clipboard text
	DataTypeEditable DataType = "editable" // Inline editable field
	DataTypeSortable DataType = "sortable" // Sortable column
	DataTypeFilter   DataType = "filter"   // Filterable column
)

// ============================================================================
// COMPONENT LIFECYCLE METHODS
// ============================================================================

// initializeLifecycle initializes component lifecycle metadata
func (c *Component) initializeLifecycle() {
	now := time.Now()
	if c.CreatedAt == nil {
		c.CreatedAt = &now
	}
	c.UpdatedAt = &now
}

// updateLifecycle updates component lifecycle metadata
func (c *Component) updateLifecycle() {
	now := time.Now()
	c.UpdatedAt = &now
}

// markDisposed marks a component as disposed (for cleanup tracking)
func (c *Component) markDisposed() {
	// In a real implementation, this might set a disposed flag
	// For now, we just update the timestamp
	now := time.Now()
	c.UpdatedAt = &now
}
